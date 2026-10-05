package cloudexport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	ErrCLIUnavailable = errors.New("TiDB Cloud CLI no está instalado o no está disponible en PATH")
	ErrCommandFailed  = errors.New("TiDB Cloud rechazó la operación de exportación")
	exportIDPattern   = regexp.MustCompile(`\bexp-[a-z0-9]+\b`)
)

type Task struct {
	ExportID     string   `json:"export_id"`
	DisplayName  string   `json:"display_name"`
	State        string   `json:"state"`
	CreateTime   string   `json:"create_time"`
	CompleteTime string   `json:"complete_time,omitempty"`
	ExpireTime   string   `json:"expire_time,omitempty"`
	Database     string   `json:"database"`
	FileType     string   `json:"file_type"`
	Compression  string   `json:"compression"`
	TargetType   string   `json:"target_type"`
	Patterns     []string `json:"patterns,omitempty"`
}

type Exporter interface {
	Create(context.Context) (Task, error)
	List(context.Context) ([]Task, error)
}

type commandRunner interface {
	LookPath(string) (string, error)
	Run(context.Context, []string, string, ...string) ([]byte, error)
}

type execRunner struct{}

func (execRunner) LookPath(name string) (string, error) { return exec.LookPath(name) }
func (execRunner) Run(ctx context.Context, environment []string, name string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Env = append(os.Environ(), environment...)
	return command.CombinedOutput()
}

type Service struct {
	clusterID          string
	database           string
	profile            string
	publicKey          string
	privateKey         string
	runner             commandRunner
	cliFallback        func() (string, error)
	commandEnvironment []string
	profileOnce        sync.Once
	profileErr         error
}

func New(clusterID, database, profile string) (*Service, error) {
	return NewWithCredentials(clusterID, database, profile, "", "")
}

func NewWithCredentials(clusterID, database, profile, publicKey, privateKey string) (*Service, error) {
	return newService(clusterID, database, profile, publicKey, privateKey, execRunner{})
}

func newService(clusterID, database, profile, publicKey, privateKey string, runner commandRunner) (*Service, error) {
	if !regexp.MustCompile(`^[0-9]{1,32}$`).MatchString(clusterID) {
		return nil, errors.New("TIDB_CLUSTER_ID no es válido")
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_]{1,64}$`).MatchString(database) {
		return nil, errors.New("TIDB_DATABASE no es válido")
	}
	if profile != "" && !regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`).MatchString(profile) {
		return nil, errors.New("TIDB_CLOUD_PROFILE no es válido")
	}
	if (publicKey == "") != (privateKey == "") {
		return nil, errors.New("las dos credenciales de TiDB Cloud deben configurarse juntas")
	}
	service := &Service{clusterID: clusterID, database: database, profile: profile, publicKey: publicKey, privateKey: privateKey, runner: runner, cliFallback: embeddedCLIPath}
	if publicKey != "" {
		configRoot := filepath.Join(os.TempDir(), "maz-suplementos-ticloud")
		service.commandEnvironment = []string{"HOME=" + configRoot, "XDG_CONFIG_HOME=" + filepath.Join(configRoot, ".config")}
	}
	return service, nil
}

func (s *Service) Create(ctx context.Context) (Task, error) {
	cli, err := s.cli()
	if err != nil {
		return Task{}, err
	}
	if err := s.prepareProfile(ctx, cli); err != nil {
		return Task{}, err
	}
	args := s.profileArgs([]string{"serverless", "export", "create", "-c", s.clusterID,
		"--target-type", "LOCAL", "--file-type", "SQL", "--compression", "GZIP",
		"--filter", s.database + ".*", "--force", "--no-color"})
	commandCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	output, err := s.runner.Run(commandCtx, s.commandEnvironment, cli, args...)
	if err != nil {
		return Task{}, fmt.Errorf("%w: verifica el perfil y los permisos de ticloud", ErrCommandFailed)
	}
	id := exportIDPattern.FindString(string(output))
	if id == "" {
		return Task{}, fmt.Errorf("%w: TiDB Cloud no devolvió un Export ID", ErrCommandFailed)
	}
	task := Task{ExportID: id}
	if tasks, listErr := s.List(ctx); listErr == nil {
		for _, candidate := range tasks {
			if candidate.ExportID == id {
				return candidate, nil
			}
		}
	}
	return task, nil
}

func (s *Service) List(ctx context.Context) ([]Task, error) {
	cli, err := s.cli()
	if err != nil {
		return nil, err
	}
	if err := s.prepareProfile(ctx, cli); err != nil {
		return nil, err
	}
	args := s.profileArgs([]string{"serverless", "export", "list", "-c", s.clusterID, "-o", "json", "--no-color"})
	commandCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	output, err := s.runner.Run(commandCtx, s.commandEnvironment, cli, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: verifica el perfil y los permisos de ticloud", ErrCommandFailed)
	}
	var response struct {
		Exports []struct {
			ExportID      string `json:"exportId"`
			DisplayName   string `json:"displayName"`
			State         string `json:"state"`
			CreateTime    string `json:"createTime"`
			CompleteTime  string `json:"completeTime"`
			ExpireTime    string `json:"expireTime"`
			ExportOptions struct {
				Compression string `json:"compression"`
				Database    any    `json:"database"`
				FileType    string `json:"fileType"`
				Filter      struct {
					Table struct {
						Patterns []string `json:"patterns"`
					} `json:"table"`
				} `json:"filter"`
			} `json:"exportOptions"`
			Target struct {
				Type string `json:"type"`
			} `json:"target"`
		} `json:"exports"`
	}
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("%w: respuesta JSON no válida", ErrCommandFailed)
	}
	tasks := make([]Task, 0, len(response.Exports))
	for _, raw := range response.Exports {
		database := databaseName(raw.ExportOptions.Database)
		if database == "" && len(raw.ExportOptions.Filter.Table.Patterns) > 0 {
			database = strings.Trim(strings.Split(raw.ExportOptions.Filter.Table.Patterns[0], ".")[0], "`")
		}
		tasks = append(tasks, Task{ExportID: raw.ExportID, DisplayName: raw.DisplayName, State: raw.State,
			CreateTime: raw.CreateTime, CompleteTime: raw.CompleteTime, ExpireTime: raw.ExpireTime,
			Database: database, FileType: raw.ExportOptions.FileType, Compression: raw.ExportOptions.Compression,
			TargetType: raw.Target.Type, Patterns: raw.ExportOptions.Filter.Table.Patterns})
	}
	return tasks, nil
}

func (s *Service) cli() (string, error) {
	path, err := s.runner.LookPath("ticloud")
	if err == nil {
		return path, nil
	}
	path, err = s.cliFallback()
	if err != nil {
		return "", ErrCLIUnavailable
	}
	return path, nil
}

func (s *Service) prepareProfile(ctx context.Context, cli string) error {
	if s.publicKey == "" {
		return nil
	}
	s.profileOnce.Do(func() {
		if err := os.MkdirAll(filepath.Join(os.TempDir(), "maz-suplementos-ticloud"), 0o700); err != nil {
			s.profileErr = fmt.Errorf("%w: no se pudo preparar el perfil", ErrCommandFailed)
			return
		}
		commandCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		_, err := s.runner.Run(commandCtx, s.commandEnvironment, cli,
			"config", "create", "--profile-name", s.profile,
			"--public-key", s.publicKey, "--private-key", s.privateKey, "--no-color")
		if err != nil {
			s.profileErr = fmt.Errorf("%w: no se pudo autenticar ticloud", ErrCommandFailed)
		}
	})
	return s.profileErr
}

func (s *Service) profileArgs(args []string) []string {
	if s.profile == "" {
		return args
	}
	return append([]string{"-P", s.profile}, args...)
}

func databaseName(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []any:
		if len(typed) > 0 {
			return fmt.Sprint(typed[0])
		}
	}
	return ""
}

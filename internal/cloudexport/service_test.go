package cloudexport

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeRunner struct {
	path    string
	outputs [][]byte
	args    [][]string
}

func (f *fakeRunner) LookPath(string) (string, error) {
	if f.path == "" {
		return "", errors.New("missing")
	}
	return f.path, nil
}
func (f *fakeRunner) Run(_ context.Context, _ string, args ...string) ([]byte, error) {
	f.args = append(f.args, args)
	output := f.outputs[0]
	f.outputs = f.outputs[1:]
	return output, nil
}

func TestCreateUsesSeparatedFixedArgumentsAndReturnsRealTask(t *testing.T) {
	runner := &fakeRunner{path: "ticloud", outputs: [][]byte{
		[]byte("Export task created successfully. Export ID: exp-abc123"),
		[]byte(`{"exports":[{"exportId":"exp-abc123","displayName":"SNAPSHOT_2026","state":"PENDING","createTime":"2026-10-02T12:00:00Z","exportOptions":{"compression":"GZIP","fileType":"SQL","filter":{"table":{"patterns":["` + "`maz_suplementos`" + `.*"]}}},"target":{"type":"LOCAL"}}]}`),
	}}
	service, err := newService("12345", "maz_suplementos", "school", runner)
	if err != nil {
		t.Fatal(err)
	}
	task, err := service.Create(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if task.ExportID != "exp-abc123" || task.State != "PENDING" {
		t.Fatalf("task=%#v", task)
	}
	want := []string{"-P", "school", "serverless", "export", "create", "-c", "12345", "--target-type", "LOCAL", "--file-type", "SQL", "--compression", "GZIP", "--filter", "maz_suplementos.*", "--force", "--no-color"}
	if !reflect.DeepEqual(runner.args[0], want) {
		t.Fatalf("args=%#v", runner.args[0])
	}
}

func TestListParsesTiDBCloudJSON(t *testing.T) {
	runner := &fakeRunner{path: "ticloud", outputs: [][]byte{[]byte(`{"exports":[{"exportId":"exp-1","state":"EXPIRED","exportOptions":{"compression":"GZIP","fileType":"SQL","database":"maz_suplementos"},"target":{"type":"LOCAL"}}]}`)}}
	service, _ := newService("123", "maz_suplementos", "", runner)
	tasks, err := service.List(context.Background())
	if err != nil || len(tasks) != 1 || tasks[0].State != "EXPIRED" || tasks[0].Database != "maz_suplementos" {
		t.Fatalf("tasks=%#v err=%v", tasks, err)
	}
}

func TestRejectsUnsafeConfiguration(t *testing.T) {
	if _, err := newService("123;rm", "maz_suplementos", "", &fakeRunner{}); err == nil {
		t.Fatal("expected invalid cluster ID")
	}
	if _, err := newService("123", "db.*", "", &fakeRunner{}); err == nil {
		t.Fatal("expected invalid database")
	}
}

func TestMissingCLIHasSpecificError(t *testing.T) {
	service, _ := newService("123", "db", "", &fakeRunner{})
	_, err := service.List(context.Background())
	if !errors.Is(err, ErrCLIUnavailable) {
		t.Fatalf("err=%v", err)
	}
}

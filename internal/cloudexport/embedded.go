package cloudexport

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

const (
	embeddedArchive = "tools/ticloud_1.0.0-beta.11_linux_amd64.tar.gz"
	maxCLISize      = 64 << 20
)

//go:embed tools/ticloud_1.0.0-beta.11_linux_amd64.tar.gz
var embeddedFiles embed.FS

var (
	extractOnce   sync.Once
	extractedPath string
	extractErr    error
)

func embeddedCLIPath() (string, error) {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		return "", errors.New("no hay un ticloud integrado para esta plataforma")
	}
	extractOnce.Do(func() {
		destination := filepath.Join(os.TempDir(), "maz-suplementos-bin", "ticloud-v1.0.0-beta.11")
		extractedPath, extractErr = extractCLI(embeddedFiles, embeddedArchive, destination)
	})
	return extractedPath, extractErr
}

func extractCLI(files embed.FS, archivePath, destination string) (string, error) {
	archive, err := files.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer archive.Close()
	gzipReader, err := gzip.NewReader(archive)
	if err != nil {
		return "", err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		if header.Typeflag != tar.TypeReg || filepath.Base(filepath.Clean(header.Name)) != "ticloud" {
			continue
		}
		if header.Size <= 0 || header.Size > maxCLISize {
			return "", errors.New("el binario ticloud integrado tiene un tamaño inválido")
		}
		if err := os.MkdirAll(destination, 0o700); err != nil {
			return "", err
		}
		binaryPath := filepath.Join(destination, "ticloud")
		output, err := os.OpenFile(binaryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o700)
		if err != nil {
			return "", err
		}
		written, copyErr := io.CopyN(output, tarReader, header.Size)
		closeErr := output.Close()
		if copyErr != nil || written != header.Size {
			return "", fmt.Errorf("no se pudo extraer ticloud: %w", copyErr)
		}
		if closeErr != nil {
			return "", closeErr
		}
		if err := os.Chmod(binaryPath, 0o700); err != nil {
			return "", err
		}
		return binaryPath, nil
	}
	return "", errors.New("el archivo integrado no contiene ticloud")
}

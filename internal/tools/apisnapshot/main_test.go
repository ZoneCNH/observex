package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSnapshotPackageCapturesOnlyExportedAPI(t *testing.T) {
	dir := t.TempDir()
	source := `package sample

import "context"

const Version = "v1"
const privateVersion = "dev"

var DefaultName string
var privateName string

type Config struct {
	Name string
	secret string
}

type Reporter interface {
	Report(context.Context, string) error
	private()
}

func NewConfig(name string) Config { return Config{Name: name} }
func helper() {}
func (c Config) Validate() error { return nil }
func (c Config) private() {}
`
	if err := os.WriteFile(filepath.Join(dir, "api.go"), []byte(source), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	got, err := snapshotPackage(dir)
	if err != nil {
		t.Fatalf("snapshot package: %v", err)
	}

	for _, fragment := range []string{
		"package observex",
		"type Config struct {",
		"Name string",
		"type Reporter interface {",
		"Report(context.Context, string) error",
		"const Version = \"v1\"",
		"var DefaultName string",
		"func NewConfig(name string) Config",
		"func (c Config) Validate() error",
	} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("snapshot missing %q\n%s", fragment, got)
		}
	}

	for _, fragment := range []string{
		"secret string",
		"privateVersion",
		"privateName",
		"private()",
		"func helper",
	} {
		if strings.Contains(got, fragment) {
			t.Fatalf("snapshot leaked private field %q\n%s", fragment, got)
		}
	}
}

func TestMainExitsWithUsageError(t *testing.T) {
	oldExitFn := exitFn
	exitFn = func(code int) {
		if code != 1 {
			t.Fatalf("exit code = %d, want 1", code)
		}
	}
	defer func() { exitFn = oldExitFn }()

	oldArgs := os.Args
	os.Args = []string{"apisnapshot", "one", "two"}
	defer func() { os.Args = oldArgs }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStderr := os.Stderr
	os.Stderr = w
	t.Cleanup(func() {
		os.Stderr = oldStderr
	})

	main()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadAll(r); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}

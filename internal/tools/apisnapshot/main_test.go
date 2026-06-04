package main

import (
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

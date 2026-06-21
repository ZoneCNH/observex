package testkit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRequireGoldenAcceptsMatchingContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.golden")

	if err := os.WriteFile(path, []byte("ok\n"), 0o600); err != nil {
		t.Fatalf("write golden: %v", err)
	}

	RequireGolden(t, path, []byte("ok\n"))
}

func TestRequireGoldenReportsMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.golden")

	var got string
	requireGolden(path, []byte("ok\n"), recordFailure(&got))
	if !strings.Contains(got, "read golden file") {
		t.Fatalf("expected read error, got %q", got)
	}
	if !strings.Contains(got, path) {
		t.Fatalf("expected path in error, got %q", got)
	}
}

func TestRequireGoldenReportsMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.golden")

	if err := os.WriteFile(path, []byte("expected"), 0o600); err != nil {
		t.Fatalf("write golden: %v", err)
	}

	var got string
	requireGolden(path, []byte("actual"), recordFailure(&got))
	want := "golden mismatch for " + path + "\nexpected:\nexpected\nactual:\nactual"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

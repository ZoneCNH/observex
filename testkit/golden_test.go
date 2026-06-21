package testkit

import (
	"fmt"
	"os"
	"path/filepath"
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

	runGoTestFailure(t, fmt.Sprintf(`package negtest

import (
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestFailure(t *testing.T) {
	testkit.RequireGolden(t, %q, []byte("ok\n"))
}
`, path), `read golden file`)
}

func TestRequireGoldenReportsMismatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.golden")

	if err := os.WriteFile(path, []byte("expected\n"), 0o600); err != nil {
		t.Fatalf("write golden: %v", err)
	}

	runGoTestFailure(t, fmt.Sprintf(`package negtest

import (
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestFailure(t *testing.T) {
	testkit.RequireGolden(t, %q, []byte("actual\n"))
}
`, path), `golden mismatch for`)
}

package testkit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func runGoTestFailure(t *testing.T, src string, wantSubstrings ...string) {
	t.Helper()

	dir := t.TempDir()
	repoRoot := repoRoot(t)

	mod := fmt.Sprintf(`module tempneg

go 1.23

require github.com/ZoneCNH/observex v0.0.0

replace github.com/ZoneCNH/observex => %s
`, filepath.ToSlash(repoRoot))
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(mod), 0o600); err != nil {
		t.Fatalf("write temp go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "neg_test.go"), []byte(src), 0o600); err != nil {
		t.Fatalf("write temp test: %v", err)
	}

	cmd := exec.Command("go", "test", "-run", "^TestFailure$", "-count=1")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected go test to fail, but it succeeded:\n%s", output)
	}

	text := string(output)
	for _, want := range wantSubstrings {
		if !strings.Contains(text, want) {
			t.Fatalf("expected go test output to contain %q, got:\n%s", want, text)
		}
	}
}

func repoRoot(t testing.TB) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	return filepath.Dir(filepath.Dir(file))
}

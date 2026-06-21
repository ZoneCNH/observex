package testkit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRequireNoErrorFailsOnError(t *testing.T) {
	runGoTestFailure(t, `package negtest

import (
	"errors"
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestFailure(t *testing.T) {
	testkit.RequireNoError(t, errors.New("boom"))
}
`, `expected no error, got boom`)
}

func TestAssertNoSecretLeakSkipsEmptySecrets(t *testing.T) {
	AssertNoSecretLeak(t, "public output", "", "absent-secret")
}

func TestAssertNoSecretLeakRejectsRawSecret(t *testing.T) {
	runGoTestFailure(t, `package negtest

import (
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestFailure(t *testing.T) {
	testkit.AssertNoSecretLeak(t, "output contains secret", "secret")
}
`, `expected text not to contain raw secret "secret"`)
}

func TestAssertNoSecretLeakRejectsIndicator(t *testing.T) {
	indicator := "token" + "="
	runGoTestFailure(t, `package negtest

import (
	"strings"
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestFailure(t *testing.T) {
	text := strings.Join([]string{"token", "=redacted"}, "")
	testkit.AssertNoSecretLeak(t, text)
}
`, `expected text not to contain secret indicator "`+indicator+`"`)
}

func TestCaptureStdoutCapturesLargeOutput(t *testing.T) {
	payload := strings.Repeat("x", 80*1024)

	output := CaptureStdout(t, func() {
		fmt.Println("line one")
		fmt.Print(payload)
	})

	if !strings.HasPrefix(output, "line one\n") {
		t.Fatalf("expected captured output to include prefix, got %q", output[:min(len(output), 16)])
	}
	if !strings.HasSuffix(output, payload) {
		t.Fatalf("expected captured output to include large payload")
	}
}

func TestCaptureStdoutCapturesEmptyOutput(t *testing.T) {
	output := CaptureStdout(t, func() {})

	if output != "" {
		t.Fatalf("expected empty captured output, got %q", output)
	}
}

func TestRequireGoldenAcceptsMatchingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "expected.txt")
	if err := os.WriteFile(path, []byte("expected"), 0o600); err != nil {
		t.Fatalf("write golden file: %v", err)
	}

	RequireGolden(t, path, []byte("expected"))
}

package testkit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssertNoSecretLeakSkipsEmptySecrets(t *testing.T) {
	AssertNoSecretLeak(t, "public output", "", "absent-secret")
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

package testkit

import (
	"io"
	"os"
	"strings"
	"testing"
)

// RequireNoError fails the test when err is non-nil.
func RequireNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// AssertNoSecretLeak fails when text contains a raw secret or obvious secret indicator.
func AssertNoSecretLeak(t testing.TB, text string, rawSecrets ...string) {
	t.Helper()
	for _, raw := range rawSecrets {
		if raw == "" {
			continue
		}
		if strings.Contains(text, raw) {
			t.Fatalf("expected text not to contain raw secret %q", raw)
		}
	}

	indicators := []string{
		"password" + "=",
		"passwd" + "=",
		"secret" + "=",
		"token" + "=",
		"access_key" + "=",
		"secret_key" + "=",
	}
	lower := strings.ToLower(text)
	for _, indicator := range indicators {
		if strings.Contains(lower, indicator) {
			t.Fatalf("expected text not to contain secret indicator %q", indicator)
		}
	}
}

type stdoutCaptureResult struct {
	output []byte
	err    error
}

// CaptureStdout runs fn while capturing writes to os.Stdout.
func CaptureStdout(t testing.TB, fn func()) string {
	t.Helper()

	original := os.Stdout
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	resultC := make(chan stdoutCaptureResult, 1)
	go func() {
		output, readErr := io.ReadAll(read)
		resultC <- stdoutCaptureResult{output: output, err: readErr}
	}()

	os.Stdout = write
	restored := false
	defer func() {
		if !restored {
			os.Stdout = original
		}
		_ = write.Close()
		_ = read.Close()
	}()

	fn()

	os.Stdout = original
	restored = true
	if err := write.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}

	result := <-resultC
	if err := read.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	if result.err != nil {
		t.Fatalf("read stdout: %v", result.err)
	}
	return string(result.output)
}

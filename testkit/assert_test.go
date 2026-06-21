package testkit

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func recordFailure(dst *string) failfFunc {
	return func(format string, args ...any) {
		*dst = fmt.Sprintf(format, args...)
	}
}

func TestRequireNoErrorFailsOnError(t *testing.T) {
	var got string
	requireNoError(errors.New("boom"), recordFailure(&got))
	if got != "expected no error, got boom" {
		t.Fatalf("got %q, want expected no error, got boom", got)
	}
}

func TestRequireNoErrorAcceptsNilHelper(t *testing.T) {
	var got string
	requireNoError(nil, recordFailure(&got))
	if got != "" {
		t.Fatalf("expected no failure, got %q", got)
	}
}

func TestAssertNoSecretLeakSkipsEmptySecrets(t *testing.T) {
	AssertNoSecretLeak(t, "public output", "", "absent-secret")
}

func TestAssertNoSecretLeakRejectsRawSecret(t *testing.T) {
	var got string
	assertNoSecretLeak("output contains secret", recordFailure(&got), "secret")
	if got != `expected text not to contain raw secret "secret"` {
		t.Fatalf("got %q, want expected text not to contain raw secret \"secret\"", got)
	}
}

func TestAssertNoSecretLeakRejectsIndicator(t *testing.T) {
	var got string
	indicator := "tok" + "en="
	assertNoSecretLeak(indicator+"redacted", recordFailure(&got))
	want := fmt.Sprintf(`expected text not to contain secret indicator %q`, indicator)
	if got != want {
		t.Fatalf("got %q, want %s", got, want)
	}
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

func TestCaptureStdoutRestoresStdoutAfterPanic(t *testing.T) {
	original := os.Stdout
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if os.Stdout != original {
			t.Fatal("expected stdout to be restored after panic")
		}
	}()

	CaptureStdout(t, func() {
		panic("boom")
	})
}

package observex

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewErrorFormatsKindOpAndMessage(t *testing.T) {
	err := NewError(ErrorKindValidation, "observex.Test", "bad input", false)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Kind != ErrorKindValidation {
		t.Fatalf("expected validation kind, got %q", err.Kind)
	}
	if err.Retryable {
		t.Fatal("expected non-retryable error")
	}
	if got := err.Error(); !strings.Contains(got, "validation: observex.Test: bad input") {
		t.Fatalf("unexpected error string: %q", got)
	}
}

func TestWrapErrorPreservesCauseAndKind(t *testing.T) {
	cause := context.DeadlineExceeded
	err := WrapError(ErrorKindTimeout, "observex.Test", "", true, cause)

	if !IsKind(err, ErrorKindTimeout) {
		t.Fatalf("expected timeout kind, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped cause, got %v", err)
	}
	if !err.Retryable {
		t.Fatal("expected retryable error")
	}
}

func TestContextErrorClassifiesDeadlineAsRetryableTimeout(t *testing.T) {
	err := contextError("observex.Test", context.DeadlineExceeded)
	if !IsKind(err, ErrorKindTimeout) {
		t.Fatalf("expected timeout kind, got %v", err)
	}
	if !err.Retryable {
		t.Fatal("expected deadline errors to be retryable")
	}
}

func TestContextErrorClassifiesCanceledAsCanceled(t *testing.T) {
	err := contextError("observex.Test", context.Canceled)
	if !IsKind(err, ErrorKindCanceled) {
		t.Fatalf("expected canceled kind, got %v", err)
	}
	if err.Retryable {
		t.Fatal("expected canceled errors to be non-retryable")
	}
}

func TestMapErrorPreservesExternalError(t *testing.T) {
	// External errors (non-observex) are mapped to ErrorKindInternal with retryable=false
	cause := errors.New("external error")

	err := MapError("observex.Test", cause)
	if !IsKind(err, ErrorKindInternal) {
		t.Fatalf("expected internal kind for external error, got %v", err)
	}
	var observexErr *Error
	if !errors.As(err, &observexErr) {
		t.Fatalf("expected observex error, got %T", err)
	}
	if observexErr.Op != "observex.Test" {
		t.Fatalf("expected mapped op, got %q", observexErr.Op)
	}
	if observexErr.Retryable {
		t.Fatal("expected external error to be non-retryable")
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected mapped error to wrap cause, got %v", err)
	}
}

func TestMapErrorClassifiesCanceled(t *testing.T) {
	err := MapError("observex.Test", context.Canceled)
	if !IsKind(err, ErrorKindCanceled) {
		t.Fatalf("expected canceled kind, got %v", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled cause, got %v", err)
	}
}

func TestErrorWithEmptyMessageAndCause(t *testing.T) {
	cause := errors.New("root cause")
	err := &Error{
		Kind:  ErrorKindInternal,
		Op:    "test.Op",
		Cause: cause,
	}
	got := err.Error()
	if !strings.Contains(got, "root cause") {
		t.Fatalf("expected error string to contain cause, got %q", got)
	}
	if !strings.Contains(got, "test.Op") {
		t.Fatalf("expected error string to contain op, got %q", got)
	}
}

func TestErrorNilReceiverMethods(t *testing.T) {
	var e *Error
	if e.Error() != "" {
		t.Fatal("expected empty string from nil Error")
	}
	if e.Unwrap() != nil {
		t.Fatal("expected nil from nil Error.Unwrap")
	}
}

package observex

import (
	"errors"
	"testing"
	"time"
)

type trackingSanitizer struct {
	called    bool
	sanitized any
}

func (s *trackingSanitizer) Sanitize() any {
	s.called = true
	return s.sanitized
}

func TestFieldConstructors(t *testing.T) {
	now := time.Unix(100, 0)
	tests := []Field{
		String("component", "api"),
		Int("attempt", 2),
		Int64("size", 42),
		Float64("duration", 1.5),
		Bool("ready", true),
		Duration("timeout", time.Second),
		Time("checked_at", now),
		Any("payload", map[string]string{"kind": "test"}),
	}

	for _, field := range tests {
		if field.Key == "" {
			t.Fatalf("expected key for %#v", field)
		}
		if field.Secret {
			t.Fatalf("expected non-secret field for %#v", field)
		}
	}

	secret := Secret("api_key", "raw-value-123")
	if !secret.Secret {
		t.Fatal("expected Secret to mark the field")
	}

	err := errors.New("boom")
	if got := Err(err); got.Key != "error" || got.Value != "boom" {
		t.Fatalf("unexpected error field: %#v", got)
	}
	if got := ErrorField(nil); got.Value != "" {
		t.Fatalf("expected empty nil error field, got %#v", got)
	}
}

func TestDefaultRedactorMasksSecretFieldsAndKeys(t *testing.T) {
	raw := "raw-value-123"
	redactor := NewDefaultRedactor("custom_credential")

	tests := []Field{
		Secret("", raw),
		Secret("plain", raw),
		String("api_key", raw),
		String("authorization", raw),
		String("db-dsn", raw),
		String("service.dsn", raw),
		String("database url", raw),
		String("access.token", raw),
		String("session_cookie", raw),
		String("custom_credential", raw),
	}
	for _, field := range tests {
		got := redactor.RedactField(field)
		if got.Value != RedactedValue {
			t.Fatalf("expected %q to be redacted, got %#v", field.Key, got)
		}
		if got.Secret {
			t.Fatalf("expected redacted field to clear secret marker: %#v", got)
		}
	}
}

func TestDefaultRedactorUsesSanitizer(t *testing.T) {
	sanitizer := &trackingSanitizer{sanitized: "sanitized"}

	got := NewDefaultRedactor().RedactField(Any("credential", sanitizer))
	if got.Value != "sanitized" {
		t.Fatalf("expected sanitizer value, got %#v", got)
	}
}

func TestDefaultRedactorPrioritizesSecretMarkerBeforeSanitizer(t *testing.T) {
	sanitizer := &trackingSanitizer{sanitized: "sanitized"}

	got := NewDefaultRedactor().RedactField(Secret("", sanitizer))

	if got.Value != RedactedValue {
		t.Fatalf("expected secret marker to force redaction, got %#v", got)
	}
	if got.Secret {
		t.Fatalf("expected redacted field to clear secret marker: %#v", got)
	}
	if sanitizer.called {
		t.Fatal("secret marker should redact without calling sanitizer")
	}
}

func TestDefaultRedactorUsesCustomSanitizerForNonSecretFields(t *testing.T) {
	sanitizer := &trackingSanitizer{sanitized: "sanitized"}

	got := NewDefaultRedactor().RedactField(Any("payload", sanitizer))

	if got.Value != "sanitized" {
		t.Fatalf("expected sanitizer value, got %#v", got)
	}
	if !sanitizer.called {
		t.Fatal("expected sanitizer to be called for non-secret fields")
	}
}

func TestRedactFieldsCopiesInput(t *testing.T) {
	fields := []Field{String("api_key", "raw-value-123"), String("component", "api")}

	got := NewDefaultRedactor().RedactFields(fields)
	if len(got) != len(fields) {
		t.Fatalf("expected %d fields, got %d", len(fields), len(got))
	}
	if got[0].Value != RedactedValue {
		t.Fatalf("expected first field to be redacted, got %#v", got[0])
	}
	if fields[0].Value == RedactedValue {
		t.Fatalf("expected input slice to remain unchanged, got %#v", fields[0])
	}
}

package observex

import (
	"errors"
	"testing"
	"time"

	"github.com/ZoneCNH/foundationx/pkg/foundationx"
)

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
		Secret("plain", raw),
		String("api_key", raw),
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

func TestDefaultRedactorUsesFoundationSanitizer(t *testing.T) {
	raw := "raw-value-123"
	field := Any("credential", foundationx.NewSecretString(raw))

	got := NewDefaultRedactor().RedactField(field)
	if got.Value != RedactedValue {
		t.Fatalf("expected sanitizer value to be redacted, got %#v", got)
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

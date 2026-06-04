package observex

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestSlogLoggerWritesContextAndRedactsFields(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger := NewSlogLogger(base)
	ctx := WithTraceID(context.Background(), "trace-123")
	ctx = WithRequestID(ctx, "request-123")
	ctx = WithContextField(ctx, String("component", "api"))

	raw := "raw-value-123"
	logger.Info(ctx, "created", String("name", "observex"), Secret("authorization", raw))

	output := buf.String()
	for _, want := range []string{"created", "trace_id", "trace-123", "request_id", "request-123", "component", "observex"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected log output to contain %q, got %s", want, output)
		}
	}
	if strings.Contains(output, raw) {
		t.Fatalf("expected log output to redact raw credential, got %s", output)
	}
	if !strings.Contains(output, RedactedValue) {
		t.Fatalf("expected log output to contain redacted marker, got %s", output)
	}
}

func TestSlogLoggerAcceptsNilContextAndNilBaseLogger(t *testing.T) {
	logger := NewSlogLogger(nil)
	logger.Debug(nil, "ignored", String("component", "api")) //nolint:staticcheck // verifies nil-context tolerance.
}

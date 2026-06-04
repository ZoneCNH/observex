package observex

import (
	"context"
	"testing"
)

func TestContextIDsAndFields(t *testing.T) {
	ctx := WithTraceID(context.Background(), "trace-123")
	ctx = WithRequestID(ctx, "request-123")
	ctx = WithCorrelationID(ctx, "correlation-123")
	ctx = WithContextFields(ctx, String("component", "api"))

	if TraceID(ctx) != "trace-123" {
		t.Fatalf("unexpected trace id: %q", TraceID(ctx))
	}
	if RequestID(ctx) != "request-123" {
		t.Fatalf("unexpected request id: %q", RequestID(ctx))
	}
	if CorrelationID(ctx) != "correlation-123" {
		t.Fatalf("unexpected correlation id: %q", CorrelationID(ctx))
	}

	fields := FieldsFromContext(ctx)
	if len(fields) != 4 {
		t.Fatalf("expected context field plus three ids, got %#v", fields)
	}
	fields[0] = String("component", "mutated")
	if ContextFields(ctx)[0].Value != "api" {
		t.Fatalf("expected context fields to be copied, got %#v", ContextFields(ctx))
	}
}

func TestContextHelpersTolerateNilContext(t *testing.T) {
	var nilCtx context.Context
	ctx := WithContextField(nilCtx, String("component", "api"))
	if ctx == nil {
		t.Fatal("expected background context")
	}
	if TraceID(nilCtx) != "" || RequestID(nilCtx) != "" || CorrelationID(nilCtx) != "" {
		t.Fatal("expected empty ids for nil context")
	}
	if ContextFields(nilCtx) != nil || FieldsFromContext(nilCtx) != nil {
		t.Fatal("expected no fields for nil context")
	}
}

func TestNoopTracerAndSpan(t *testing.T) {
	var nilCtx context.Context
	ctx, span := NewNoopTracer().Start(nilCtx, "observex.Test", String("component", "api"))
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if span == nil {
		t.Fatal("expected noop span")
	}
	span.SetField(String("component", "api"))
	span.AddEvent("event", String("component", "api"))
	span.End(String("status", "ok"))
}

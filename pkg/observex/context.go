package observex

import "context"

type traceIDContextKey struct{}
type requestIDContextKey struct{}
type correlationIDContextKey struct{}
type fieldsContextKey struct{}

// WithTraceID returns a context carrying traceID.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(nonNilContext(ctx), traceIDContextKey{}, traceID)
}

// TraceID returns the trace identifier from ctx, if present.
func TraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(traceIDContextKey{}).(string)
	return value
}

// WithRequestID returns a context carrying requestID.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(nonNilContext(ctx), requestIDContextKey{}, requestID)
}

// RequestID returns the request identifier from ctx, if present.
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(requestIDContextKey{}).(string)
	return value
}

// WithCorrelationID returns a context carrying correlationID.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(nonNilContext(ctx), correlationIDContextKey{}, correlationID)
}

// CorrelationID returns the correlation identifier from ctx, if present.
func CorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(correlationIDContextKey{}).(string)
	return value
}

// WithContextField appends field to the context field set.
func WithContextField(ctx context.Context, field Field) context.Context {
	return WithContextFields(ctx, field)
}

// WithContextFields appends fields to the context field set.
func WithContextFields(ctx context.Context, fields ...Field) context.Context {
	ctx = nonNilContext(ctx)
	existing := ContextFields(ctx)
	next := make([]Field, 0, len(existing)+len(fields))
	next = append(next, existing...)
	next = append(next, fields...)
	return context.WithValue(ctx, fieldsContextKey{}, next)
}

// ContextFields returns a copy of fields stored directly in ctx.
func ContextFields(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}
	fields, _ := ctx.Value(fieldsContextKey{}).([]Field)
	if len(fields) == 0 {
		return nil
	}
	return append([]Field(nil), fields...)
}

// FieldsFromContext returns context fields plus standard correlation identifiers.
func FieldsFromContext(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}
	fields := ContextFields(ctx)
	if traceID := TraceID(ctx); traceID != "" {
		fields = append(fields, String("trace_id", traceID))
	}
	if requestID := RequestID(ctx); requestID != "" {
		fields = append(fields, String("request_id", requestID))
	}
	if correlationID := CorrelationID(ctx); correlationID != "" {
		fields = append(fields, String("correlation_id", correlationID))
	}
	return fields
}

func nonNilContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

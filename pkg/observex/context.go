package observex

import "context"

type traceIDContextKey struct{}
type requestIDContextKey struct{}
type correlationIDContextKey struct{}
type fieldsContextKey struct{}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(nonNilContext(ctx), traceIDContextKey{}, traceID)
}

func TraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(traceIDContextKey{}).(string)
	return value
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(nonNilContext(ctx), requestIDContextKey{}, requestID)
}

func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(requestIDContextKey{}).(string)
	return value
}

func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(nonNilContext(ctx), correlationIDContextKey{}, correlationID)
}

func CorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(correlationIDContextKey{}).(string)
	return value
}

func WithContextField(ctx context.Context, field Field) context.Context {
	return WithContextFields(ctx, field)
}

func WithContextFields(ctx context.Context, fields ...Field) context.Context {
	ctx = nonNilContext(ctx)
	existing := ContextFields(ctx)
	next := make([]Field, 0, len(existing)+len(fields))
	next = append(next, existing...)
	next = append(next, fields...)
	return context.WithValue(ctx, fieldsContextKey{}, next)
}

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

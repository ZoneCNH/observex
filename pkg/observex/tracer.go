package observex

import "context"

// Tracer starts spans for observex operations.
type Tracer interface {
	Start(ctx context.Context, name string, fields ...Field) (context.Context, Span)
}

// Span records structured data for one traced operation.
type Span interface {
	SetField(field Field)
	AddEvent(name string, fields ...Field)
	End(fields ...Field)
}

// NoopTracer starts spans that drop all data.
type NoopTracer struct{}

// NewNoopTracer returns a tracer that intentionally drops spans.
func NewNoopTracer() NoopTracer {
	return NoopTracer{}
}

// Start returns ctx and a NoopSpan.
func (NoopTracer) Start(ctx context.Context, name string, fields ...Field) (context.Context, Span) {
	return nonNilContext(ctx), NoopSpan{}
}

// NoopSpan drops all span data.
type NoopSpan struct{}

// SetField drops field.
func (NoopSpan) SetField(field Field) {
	_ = field
}

// AddEvent drops an event.
func (NoopSpan) AddEvent(name string, fields ...Field) {
	_, _ = name, fields
}

// End drops span completion fields.
func (NoopSpan) End(fields ...Field) {
	_ = fields
}

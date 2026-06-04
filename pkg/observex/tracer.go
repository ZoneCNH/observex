package observex

import "context"

type Tracer interface {
	Start(ctx context.Context, name string, fields ...Field) (context.Context, Span)
}

type Span interface {
	SetField(field Field)
	AddEvent(name string, fields ...Field)
	End(fields ...Field)
}

type NoopTracer struct{}

func NewNoopTracer() NoopTracer {
	return NoopTracer{}
}

func (NoopTracer) Start(ctx context.Context, name string, fields ...Field) (context.Context, Span) {
	return nonNilContext(ctx), NoopSpan{}
}

type NoopSpan struct{}

func (NoopSpan) SetField(field Field) {}

func (NoopSpan) AddEvent(name string, fields ...Field) {}

func (NoopSpan) End(fields ...Field) {}

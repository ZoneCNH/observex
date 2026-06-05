package observex

import (
	"context"
	"sync"
)

// SpanEvent records a named event emitted by an in-memory span.
type SpanEvent struct {
	// Name is the event name supplied by the span.
	Name string
	// Fields are the redacted event fields.
	Fields []Field
}

// SpanRecord captures the final state of an in-memory span.
type SpanRecord struct {
	// Sequence is the one-based order in which the span was started.
	Sequence uint64
	// Name is the span name supplied by the tracer.
	Name string
	// Fields are the redacted fields attached before or during the span.
	Fields []Field
	// Events are the redacted events emitted by the span.
	Events []SpanEvent
	// Ended reports whether End has been called.
	Ended bool
	// EndFields are the redacted fields supplied to End.
	EndFields []Field
}

// MemoryTracer records completed spans in memory for tests and examples.
type MemoryTracer struct {
	mu       sync.Mutex
	redactor Redactor
	next     uint64
	spans    []SpanRecord
}

// NewMemoryTracer returns a tracer backed by memory.
func NewMemoryTracer(opts ...LoggerOption) *MemoryTracer {
	options := loggerOptions{redactor: NewDefaultRedactor()}
	for _, opt := range opts {
		opt(&options)
	}
	if options.redactor == nil {
		options.redactor = NewDefaultRedactor()
	}
	return &MemoryTracer{redactor: options.redactor}
}

// Start starts a new in-memory span.
func (t *MemoryTracer) Start(ctx context.Context, name string, fields ...Field) (context.Context, Span) {
	ctx = nonNilContext(ctx)
	if t == nil {
		return ctx, NoopSpan{}
	}
	fields = redactFields(t.redactor, fields)

	t.mu.Lock()
	defer t.mu.Unlock()

	t.next++
	index := len(t.spans)
	t.spans = append(t.spans, SpanRecord{
		Sequence: t.next,
		Name:     name,
		Fields:   cloneFields(fields),
	})
	return ctx, &memorySpan{tracer: t, index: index}
}

// Spans returns a snapshot of completed spans.
func (t *MemoryTracer) Spans() []SpanRecord {
	if t == nil {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	out := make([]SpanRecord, len(t.spans))
	for i, span := range t.spans {
		out[i] = span
		out[i].Fields = cloneFields(span.Fields)
		out[i].EndFields = cloneFields(span.EndFields)
		out[i].Events = cloneSpanEvents(span.Events)
	}
	return out
}

// Reset removes all completed spans.
func (t *MemoryTracer) Reset() {
	if t == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.spans = nil
	t.next = 0
}

type memorySpan struct {
	tracer *MemoryTracer
	index  int
}

func (s *memorySpan) SetField(field Field) {
	if s == nil || s.tracer == nil {
		return
	}
	fields := redactFields(s.tracer.redactor, []Field{field})
	if len(fields) == 0 {
		return
	}
	field = fields[0]

	s.tracer.mu.Lock()
	defer s.tracer.mu.Unlock()

	if s.index < 0 || s.index >= len(s.tracer.spans) {
		return
	}
	s.tracer.spans[s.index].Fields = append(s.tracer.spans[s.index].Fields, field)
}

func (s *memorySpan) AddEvent(name string, fields ...Field) {
	if s == nil || s.tracer == nil {
		return
	}
	fields = redactFields(s.tracer.redactor, fields)

	s.tracer.mu.Lock()
	defer s.tracer.mu.Unlock()

	if s.index < 0 || s.index >= len(s.tracer.spans) {
		return
	}
	s.tracer.spans[s.index].Events = append(s.tracer.spans[s.index].Events, SpanEvent{Name: name, Fields: cloneFields(fields)})
}

func (s *memorySpan) End(fields ...Field) {
	if s == nil || s.tracer == nil {
		return
	}
	fields = redactFields(s.tracer.redactor, fields)

	s.tracer.mu.Lock()
	defer s.tracer.mu.Unlock()

	if s.index < 0 || s.index >= len(s.tracer.spans) || s.tracer.spans[s.index].Ended {
		return
	}
	s.tracer.spans[s.index].Ended = true
	s.tracer.spans[s.index].EndFields = cloneFields(fields)
}

func cloneSpanEvents(events []SpanEvent) []SpanEvent {
	if len(events) == 0 {
		return nil
	}

	out := make([]SpanEvent, len(events))
	for i, event := range events {
		out[i] = event
		out[i].Fields = cloneFields(event.Fields)
	}
	return out
}

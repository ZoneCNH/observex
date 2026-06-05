package testkit

import (
	"context"
	"sync"

	"github.com/ZoneCNH/observex/pkg/observex"
)

// LogEntry is the canonical in-memory log record type.
type LogEntry = observex.LogRecord

// RecordingLogger adapts observex.MemoryLogger for downstream tests.
type RecordingLogger struct {
	mu       sync.Mutex
	recorder *observex.MemoryLogger
}

func NewRecordingLogger() *RecordingLogger {
	return &RecordingLogger{recorder: observex.NewMemoryLogger()}
}

// Debug records a debug log entry.
func (l *RecordingLogger) Debug(ctx context.Context, msg string, fields ...observex.Field) {
	if recorder := l.memory(); recorder != nil {
		recorder.Debug(ctx, msg, fields...)
	}
}

// Info records an informational log entry.
func (l *RecordingLogger) Info(ctx context.Context, msg string, fields ...observex.Field) {
	if recorder := l.memory(); recorder != nil {
		recorder.Info(ctx, msg, fields...)
	}
}

// Warn records a warning log entry.
func (l *RecordingLogger) Warn(ctx context.Context, msg string, fields ...observex.Field) {
	if recorder := l.memory(); recorder != nil {
		recorder.Warn(ctx, msg, fields...)
	}
}

// Error records an error log entry.
func (l *RecordingLogger) Error(ctx context.Context, msg string, fields ...observex.Field) {
	if recorder := l.memory(); recorder != nil {
		recorder.Error(ctx, msg, fields...)
	}
}

// Entries returns recorded log entries.
func (l *RecordingLogger) Entries() []LogEntry {
	return l.Records()
}

// Records returns recorded log entries.
func (l *RecordingLogger) Records() []LogEntry {
	if recorder := l.memory(); recorder != nil {
		return recorder.Records()
	}
	return nil
}

// Reset clears recorded log entries.
func (l *RecordingLogger) Reset() {
	if recorder := l.memory(); recorder != nil {
		recorder.Reset()
	}
}

// HasEntry reports whether a log entry with level and message was recorded.
func (l *RecordingLogger) HasEntry(level string, message string) bool {
	for _, entry := range l.Entries() {
		if string(entry.Level) == level && entry.Message == message {
			return true
		}
	}
	return false
}

func (l *RecordingLogger) memory() *observex.MemoryLogger {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.recorder == nil {
		l.recorder = observex.NewMemoryLogger()
	}
	return l.recorder
}

// MetricKind is the canonical in-memory metric kind type.
type MetricKind = observex.MetricKind

const (
	MetricKindCounter   = observex.MetricKindCounter
	MetricKindHistogram = observex.MetricKindHistogram
	MetricKindGauge     = observex.MetricKindGauge
)

// MetricRecord is the canonical in-memory metric record type.
type MetricRecord = observex.MetricRecord

// RecordingMetrics adapts observex.MemoryMetrics for downstream tests.
type RecordingMetrics struct {
	mu       sync.Mutex
	recorder *observex.MemoryMetrics
}

// NewRecordingMetrics returns a metrics recorder backed by observex.MemoryMetrics.
func NewRecordingMetrics() *RecordingMetrics {
	return &RecordingMetrics{recorder: observex.NewMemoryMetrics()}
}

// IncCounter increments a counter by one.
func (m *RecordingMetrics) IncCounter(name string, labels observex.Labels) {
	if recorder := m.memory(); recorder != nil {
		recorder.IncCounter(name, labels)
	}
}

// AddCounter increments a counter by delta.
func (m *RecordingMetrics) AddCounter(name string, delta float64, labels observex.Labels) {
	if recorder := m.memory(); recorder != nil {
		recorder.AddCounter(name, delta, labels)
	}
}

// ObserveHistogram records a histogram observation.
func (m *RecordingMetrics) ObserveHistogram(name string, value float64, labels observex.Labels) {
	if recorder := m.memory(); recorder != nil {
		recorder.ObserveHistogram(name, value, labels)
	}
}

// SetGauge records a gauge assignment.
func (m *RecordingMetrics) SetGauge(name string, value float64, labels observex.Labels) {
	if recorder := m.memory(); recorder != nil {
		recorder.SetGauge(name, value, labels)
	}
}

// Records returns recorded metric calls.
func (m *RecordingMetrics) Records() []MetricRecord {
	if recorder := m.memory(); recorder != nil {
		return recorder.Records()
	}
	return nil
}

// Counters returns counter totals keyed by metric name and labels.
func (m *RecordingMetrics) Counters() map[string]float64 {
	if recorder := m.memory(); recorder != nil {
		return recorder.Counters()
	}
	return nil
}

// Gauges returns current gauge values keyed by metric name and labels.
func (m *RecordingMetrics) Gauges() map[string]float64 {
	if recorder := m.memory(); recorder != nil {
		return recorder.Gauges()
	}
	return nil
}

// Reset clears recorded metric state.
func (m *RecordingMetrics) Reset() {
	if recorder := m.memory(); recorder != nil {
		recorder.Reset()
	}
}

// HasMetric reports whether a metric matching kind, name, and labels was recorded.
func (m *RecordingMetrics) HasMetric(kind MetricKind, name string, labels observex.Labels) bool {
	for _, record := range m.Records() {
		if record.Kind == kind && record.Name == name && sameLabels(record.Labels, labels) {
			return true
		}
	}
	return false
}

func (m *RecordingMetrics) memory() *observex.MemoryMetrics {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.recorder == nil {
		m.recorder = observex.NewMemoryMetrics()
	}
	return m.recorder
}

// SpanEvent is the canonical in-memory span event type.
type SpanEvent = observex.SpanEvent

// SpanRecord is the canonical in-memory span record type.
type SpanRecord = observex.SpanRecord

// RecordingTracer adapts observex.MemoryTracer for downstream tests.
type RecordingTracer struct {
	mu       sync.Mutex
	recorder *observex.MemoryTracer
}

func NewRecordingTracer() *RecordingTracer {
	return &RecordingTracer{recorder: observex.NewMemoryTracer()}
}

// Start starts a recording span.
func (t *RecordingTracer) Start(ctx context.Context, name string, fields ...observex.Field) (context.Context, observex.Span) {
	if recorder := t.memory(); recorder != nil {
		return recorder.Start(ctx, name, fields...)
	}
	return observex.NoopTracer{}.Start(ctx, name, fields...)
}

// Spans returns recorded spans.
func (t *RecordingTracer) Spans() []SpanRecord {
	if recorder := t.memory(); recorder != nil {
		return recorder.Spans()
	}
	return nil
}

// Reset clears recorded spans.
func (t *RecordingTracer) Reset() {
	if recorder := t.memory(); recorder != nil {
		recorder.Reset()
	}
}

// HasSpan reports whether a span with name was recorded.
func (t *RecordingTracer) HasSpan(name string) bool {
	for _, span := range t.Spans() {
		if span.Name == name {
			return true
		}
	}
	return false
}

func (t *RecordingTracer) memory() *observex.MemoryTracer {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.recorder == nil {
		t.recorder = observex.NewMemoryTracer()
	}
	return t.recorder
}

func sameLabels(actual observex.Labels, expected observex.Labels) bool {
	if len(actual) != len(expected) {
		return false
	}
	for key, value := range expected {
		if actual[key] != value {
			return false
		}
	}
	return true
}

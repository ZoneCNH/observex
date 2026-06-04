package testkit

import (
	"context"
	"sync"

	"github.com/ZoneCNH/observex/pkg/observex"
)

type LogEntry = observex.LogRecord

// RecordingLogger adapts observex.MemoryLogger for downstream tests.
type RecordingLogger struct {
	mu       sync.Mutex
	recorder *observex.MemoryLogger
}

func NewRecordingLogger() *RecordingLogger {
	return &RecordingLogger{recorder: observex.NewMemoryLogger()}
}

func (l *RecordingLogger) Debug(ctx context.Context, msg string, fields ...observex.Field) {
	if recorder := l.memory(); recorder != nil {
		recorder.Debug(ctx, msg, fields...)
	}
}

func (l *RecordingLogger) Info(ctx context.Context, msg string, fields ...observex.Field) {
	if recorder := l.memory(); recorder != nil {
		recorder.Info(ctx, msg, fields...)
	}
}

func (l *RecordingLogger) Warn(ctx context.Context, msg string, fields ...observex.Field) {
	if recorder := l.memory(); recorder != nil {
		recorder.Warn(ctx, msg, fields...)
	}
}

func (l *RecordingLogger) Error(ctx context.Context, msg string, fields ...observex.Field) {
	if recorder := l.memory(); recorder != nil {
		recorder.Error(ctx, msg, fields...)
	}
}

func (l *RecordingLogger) Entries() []LogEntry {
	return l.Records()
}

func (l *RecordingLogger) Records() []LogEntry {
	if recorder := l.memory(); recorder != nil {
		return recorder.Records()
	}
	return nil
}

func (l *RecordingLogger) Reset() {
	if recorder := l.memory(); recorder != nil {
		recorder.Reset()
	}
}

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

type MetricKind = observex.MetricKind

const (
	MetricKindCounter   = observex.MetricKindCounter
	MetricKindHistogram = observex.MetricKindHistogram
	MetricKindGauge     = observex.MetricKindGauge
)

type MetricRecord = observex.MetricRecord

type RecordingMetrics struct {
	mu       sync.Mutex
	recorder *observex.MemoryMetrics
}

func NewRecordingMetrics() *RecordingMetrics {
	return &RecordingMetrics{recorder: observex.NewMemoryMetrics()}
}

func (m *RecordingMetrics) IncCounter(name string, labels observex.Labels) {
	if recorder := m.memory(); recorder != nil {
		recorder.IncCounter(name, labels)
	}
}

func (m *RecordingMetrics) AddCounter(name string, delta float64, labels observex.Labels) {
	if recorder := m.memory(); recorder != nil {
		recorder.AddCounter(name, delta, labels)
	}
}

func (m *RecordingMetrics) ObserveHistogram(name string, value float64, labels observex.Labels) {
	if recorder := m.memory(); recorder != nil {
		recorder.ObserveHistogram(name, value, labels)
	}
}

func (m *RecordingMetrics) SetGauge(name string, value float64, labels observex.Labels) {
	if recorder := m.memory(); recorder != nil {
		recorder.SetGauge(name, value, labels)
	}
}

func (m *RecordingMetrics) Records() []MetricRecord {
	if recorder := m.memory(); recorder != nil {
		return recorder.Records()
	}
	return nil
}

func (m *RecordingMetrics) Counters() map[string]float64 {
	if recorder := m.memory(); recorder != nil {
		return recorder.Counters()
	}
	return nil
}

func (m *RecordingMetrics) Gauges() map[string]float64 {
	if recorder := m.memory(); recorder != nil {
		return recorder.Gauges()
	}
	return nil
}

func (m *RecordingMetrics) Reset() {
	if recorder := m.memory(); recorder != nil {
		recorder.Reset()
	}
}

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

type SpanEvent = observex.SpanEvent

type SpanRecord = observex.SpanRecord

// RecordingTracer adapts observex.MemoryTracer for downstream tests.
type RecordingTracer struct {
	mu       sync.Mutex
	recorder *observex.MemoryTracer
}

func NewRecordingTracer() *RecordingTracer {
	return &RecordingTracer{recorder: observex.NewMemoryTracer()}
}

func (t *RecordingTracer) Start(ctx context.Context, name string, fields ...observex.Field) (context.Context, observex.Span) {
	if recorder := t.memory(); recorder != nil {
		return recorder.Start(ctx, name, fields...)
	}
	return observex.NoopTracer{}.Start(ctx, name, fields...)
}

func (t *RecordingTracer) Spans() []SpanRecord {
	if recorder := t.memory(); recorder != nil {
		return recorder.Spans()
	}
	return nil
}

func (t *RecordingTracer) Reset() {
	if recorder := t.memory(); recorder != nil {
		recorder.Reset()
	}
}

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

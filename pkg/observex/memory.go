package observex

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type LogRecord struct {
	Sequence uint64
	Level    LogLevel
	Message  string
	Fields   []Field
}

type MemoryLogger struct {
	mu       sync.Mutex
	redactor Redactor
	next     uint64
	records  []LogRecord
}

func NewMemoryLogger(opts ...LoggerOption) *MemoryLogger {
	options := loggerOptions{redactor: NewDefaultRedactor()}
	for _, opt := range opts {
		opt(&options)
	}
	if options.redactor == nil {
		options.redactor = NewDefaultRedactor()
	}
	return &MemoryLogger{redactor: options.redactor}
}

func (l *MemoryLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LogLevelDebug, msg, fields...)
}

func (l *MemoryLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LogLevelInfo, msg, fields...)
}

func (l *MemoryLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LogLevelWarn, msg, fields...)
}

func (l *MemoryLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LogLevelError, msg, fields...)
}

func (l *MemoryLogger) Records() []LogRecord {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	records := make([]LogRecord, len(l.records))
	for i, record := range l.records {
		records[i] = record
		records[i].Fields = cloneFields(record.Fields)
	}
	return records
}

func (l *MemoryLogger) Reset() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.records = nil
	l.next = 0
}

func (l *MemoryLogger) log(ctx context.Context, level LogLevel, msg string, fields ...Field) {
	if l == nil {
		return
	}
	allFields := append(FieldsFromContext(ctx), fields...)
	allFields = redactFields(l.redactor, allFields)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.next++
	l.records = append(l.records, LogRecord{
		Sequence: l.next,
		Level:    level,
		Message:  msg,
		Fields:   cloneFields(allFields),
	})
}

type MetricKind string

const (
	MetricKindCounter   MetricKind = "counter"
	MetricKindHistogram MetricKind = "histogram"
	MetricKindGauge     MetricKind = "gauge"
)

type MetricRecord struct {
	Sequence uint64
	Kind     MetricKind
	Name     string
	Value    float64
	Labels   Labels
}

type MemoryMetrics struct {
	mu       sync.Mutex
	next     uint64
	records  []MetricRecord
	counters map[string]float64
	gauges   map[string]float64
}

func NewMemoryMetrics() *MemoryMetrics {
	return &MemoryMetrics{
		counters: make(map[string]float64),
		gauges:   make(map[string]float64),
	}
}

func (m *MemoryMetrics) IncCounter(name string, labels Labels) {
	m.AddCounter(name, 1, labels)
}

func (m *MemoryMetrics) AddCounter(name string, delta float64, labels Labels) {
	m.record(MetricKindCounter, name, delta, labels)
}

func (m *MemoryMetrics) ObserveHistogram(name string, value float64, labels Labels) {
	m.record(MetricKindHistogram, name, value, labels)
}

func (m *MemoryMetrics) SetGauge(name string, value float64, labels Labels) {
	m.record(MetricKindGauge, name, value, labels)
}

func (m *MemoryMetrics) Records() []MetricRecord {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	records := make([]MetricRecord, len(m.records))
	for i, record := range m.records {
		records[i] = record
		records[i].Labels = CloneLabels(record.Labels)
	}
	return records
}

func (m *MemoryMetrics) Counters() map[string]float64 {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneFloatMap(m.counters)
}

func (m *MemoryMetrics) Gauges() map[string]float64 {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneFloatMap(m.gauges)
}

func (m *MemoryMetrics) Reset() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records = nil
	m.counters = make(map[string]float64)
	m.gauges = make(map[string]float64)
	m.next = 0
}

func (m *MemoryMetrics) record(kind MetricKind, name string, value float64, labels Labels) {
	if m == nil {
		return
	}
	name = sanitizeMetricName(name)
	labels = SanitizeLabels(labels)
	key := metricRecordKey(name, labels)

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.counters == nil {
		m.counters = make(map[string]float64)
	}
	if m.gauges == nil {
		m.gauges = make(map[string]float64)
	}
	m.next++
	m.records = append(m.records, MetricRecord{
		Sequence: m.next,
		Kind:     kind,
		Name:     name,
		Value:    value,
		Labels:   CloneLabels(labels),
	})
	switch kind {
	case MetricKindCounter:
		m.counters[key] += value
	case MetricKindGauge:
		m.gauges[key] = value
	}
}

type SpanEvent struct {
	Name   string
	Fields []Field
}

type SpanRecord struct {
	Sequence  uint64
	Name      string
	Fields    []Field
	Events    []SpanEvent
	Ended     bool
	EndFields []Field
}

type MemoryTracer struct {
	mu       sync.Mutex
	redactor Redactor
	next     uint64
	spans    []SpanRecord
}

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

func (t *MemoryTracer) Spans() []SpanRecord {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	spans := make([]SpanRecord, len(t.spans))
	for i, span := range t.spans {
		spans[i] = span
		spans[i].Fields = cloneFields(span.Fields)
		spans[i].EndFields = cloneFields(span.EndFields)
		spans[i].Events = cloneSpanEvents(span.Events)
	}
	return spans
}

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

type MemoryHealthReporter struct {
	mu      sync.Mutex
	status  HealthStatus
	records []HealthStatus
}

func NewMemoryHealthReporter(status HealthStatus) *MemoryHealthReporter {
	if status.Name == "" {
		status.Name = "memory"
	}
	if status.Status == "" {
		status.Status = HealthHealthy
	}
	if status.CheckedAt.IsZero() {
		status.CheckedAt = deterministicTime()
	}
	status.Metadata = sanitizeHealthMetadata(status.Metadata)
	return &MemoryHealthReporter{status: status}
}

func (r *MemoryHealthReporter) HealthCheck(ctx context.Context) HealthStatus {
	if r == nil {
		return NoopHealthReporter{}.HealthCheck(ctx)
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	status := cloneHealthStatus(r.status)
	if ctx == nil {
		status.Status = HealthUnhealthy
		status.Message = "context is required"
	} else if err := ctx.Err(); err != nil {
		status.Status = HealthUnhealthy
		status.Message = err.Error()
	}
	r.records = append(r.records, cloneHealthStatus(status))
	return status
}

func (r *MemoryHealthReporter) ReadinessCheck(ctx context.Context) HealthStatus {
	return r.HealthCheck(ctx)
}

func (r *MemoryHealthReporter) SetStatus(status HealthStatus) {
	if r == nil {
		return
	}
	if status.Name == "" {
		status.Name = "memory"
	}
	if status.Status == "" {
		status.Status = HealthHealthy
	}
	if status.CheckedAt.IsZero() {
		status.CheckedAt = deterministicTime()
	}
	status.Metadata = sanitizeHealthMetadata(status.Metadata)

	r.mu.Lock()
	defer r.mu.Unlock()
	r.status = status
}

func (r *MemoryHealthReporter) Records() []HealthStatus {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	records := make([]HealthStatus, len(r.records))
	for i, record := range r.records {
		records[i] = cloneHealthStatus(record)
	}
	return records
}

func cloneFields(fields []Field) []Field {
	if len(fields) == 0 {
		return nil
	}
	copied := make([]Field, len(fields))
	copy(copied, fields)
	return copied
}

func redactFields(redactor Redactor, fields []Field) []Field {
	if redactor == nil {
		return cloneFields(fields)
	}
	return redactor.RedactFields(fields)
}

func sanitizeMetricName(name string) string {
	name = strings.TrimSpace(name)
	if ValidateMetricName(name) == nil && !IsSecretKey(name) {
		return name
	}
	return "invalid_metric_name"
}

func metricRecordKey(name string, labels Labels) string {
	if len(labels) == 0 {
		return name
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys)+1)
	parts = append(parts, name)
	for _, key := range keys {
		parts = append(parts, key+"="+labels[key])
	}
	return strings.Join(parts, "|")
}

func cloneFloatMap(values map[string]float64) map[string]float64 {
	if len(values) == 0 {
		return nil
	}
	copied := make(map[string]float64, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}

func cloneSpanEvents(events []SpanEvent) []SpanEvent {
	if len(events) == 0 {
		return nil
	}
	copied := make([]SpanEvent, len(events))
	for i, event := range events {
		copied[i] = event
		copied[i].Fields = cloneFields(event.Fields)
	}
	return copied
}

func sanitizeHealthMetadata(metadata map[string]string) map[string]string {
	if len(metadata) == 0 {
		return nil
	}
	sanitized := make(map[string]string, len(metadata))
	for key, value := range metadata {
		key = strings.TrimSpace(key)
		if key == "" || IsSecretKey(key) {
			continue
		}
		if valueLooksSecret(value) {
			value = RedactedValue
		}
		sanitized[key] = value
	}
	if len(sanitized) == 0 {
		return nil
	}
	return sanitized
}

func cloneHealthStatus(status HealthStatus) HealthStatus {
	status.Metadata = sanitizeHealthMetadata(status.Metadata)
	return status
}

func deterministicTime() time.Time {
	return time.Unix(0, 0).UTC()
}

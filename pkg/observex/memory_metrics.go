package observex

import (
	"sort"
	"strings"
	"sync"
)

// MetricKind identifies the type of an in-memory metric record.
type MetricKind string

const (
	// MetricKindCounter records counter increments.
	MetricKindCounter MetricKind = "counter"
	// MetricKindHistogram records histogram observations.
	MetricKindHistogram MetricKind = "histogram"
	// MetricKindGauge records gauge assignments.
	MetricKindGauge MetricKind = "gauge"
)

// MetricRecord captures a single call made to MemoryMetrics.
type MetricRecord struct {
	// Sequence is the one-based order in which the record was captured.
	Sequence uint64
	// Kind is the metric operation kind.
	Kind MetricKind
	// Name is the sanitized metric name.
	Name string
	// Value is the counter delta, histogram observation, or gauge value.
	Value float64
	// Labels are sanitized metric labels captured with the record.
	Labels Labels
}

// MemoryMetrics records metric calls in memory for tests and examples.
type MemoryMetrics struct {
	mu       sync.Mutex
	next     uint64
	records  []MetricRecord
	counters map[string]float64
	gauges   map[string]float64
}

// NewMemoryMetrics returns a metrics recorder backed by memory.
func NewMemoryMetrics() *MemoryMetrics {
	return &MemoryMetrics{
		counters: make(map[string]float64),
		gauges:   make(map[string]float64),
	}
}

// IncCounter increments a counter by one.
func (m *MemoryMetrics) IncCounter(name string, labels Labels) {
	m.AddCounter(name, 1, labels)
}

// AddCounter increments a counter by value.
func (m *MemoryMetrics) AddCounter(name string, delta float64, labels Labels) {
	m.record(MetricKindCounter, name, delta, labels)
}

// ObserveHistogram records a histogram observation.
func (m *MemoryMetrics) ObserveHistogram(name string, value float64, labels Labels) {
	m.record(MetricKindHistogram, name, value, labels)
}

// SetGauge records a gauge assignment.
func (m *MemoryMetrics) SetGauge(name string, value float64, labels Labels) {
	m.record(MetricKindGauge, name, value, labels)
}

// Records returns a snapshot of recorded metric calls.
func (m *MemoryMetrics) Records() []MetricRecord {
	if m == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]MetricRecord, len(m.records))
	for i, record := range m.records {
		out[i] = record
		out[i].Labels = CloneLabels(record.Labels)
	}
	return out
}

// Counters returns counter totals keyed by metric name and fields.
func (m *MemoryMetrics) Counters() map[string]float64 {
	if m == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return cloneFloatMap(m.counters)
}

// Gauges returns current gauge values keyed by metric name and fields.
func (m *MemoryMetrics) Gauges() map[string]float64 {
	if m == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return cloneFloatMap(m.gauges)
}

// Reset removes all recorded metric state.
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
	record := MetricRecord{
		Sequence: m.next,
		Kind:     kind,
		Name:     name,
		Value:    value,
		Labels:   CloneLabels(labels),
	}
	m.records = append(m.records, record)
	switch kind {
	case MetricKindCounter:
		m.counters[key] += value
	case MetricKindGauge:
		m.gauges[key] = value
	}
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

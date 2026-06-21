package observex

const (
	// MetricClientCreatedTotal counts created clients.
	MetricClientCreatedTotal = "client_created_total"
	// MetricClientClosedTotal counts closed clients.
	MetricClientClosedTotal = "client_closed_total"
	// MetricClientErrorsTotal counts client lifecycle errors.
	MetricClientErrorsTotal = "client_errors_total"
	// MetricClientHealthStatus reports health status as a gauge.
	MetricClientHealthStatus = "client_health_status"
	// MetricClientHealthLatencyMS records health check latency in milliseconds.
	MetricClientHealthLatencyMS = "client_health_latency_ms"
	// MetricClientRequestsTotal counts client requests.
	MetricClientRequestsTotal = "client_requests_total"
	// MetricClientRequestDurationSeconds records request duration in seconds.
	MetricClientRequestDurationSeconds = "client_request_duration_seconds"
	// MetricClientRetriesTotal counts retry attempts.
	MetricClientRetriesTotal = "client_retries_total"
	// MetricClientInflight reports current in-flight work.
	MetricClientInflight = "client_inflight"
)

// Metrics records counters, histograms, and gauges.
type Metrics interface {
	IncCounter(name string, labels Labels)
	AddCounter(name string, delta float64, labels Labels)
	ObserveHistogram(name string, value float64, labels Labels)
	SetGauge(name string, value float64, labels Labels)
}

// NoopMetrics drops all metric observations.
type NoopMetrics struct{}

// NewNoopMetrics returns a metrics recorder that intentionally drops all observations.
func NewNoopMetrics() NoopMetrics {
	return NoopMetrics{}
}

// IncCounter drops a counter increment.
func (NoopMetrics) IncCounter(name string, labels Labels) {
	_, _ = name, labels
}

// AddCounter drops a counter addition.
func (NoopMetrics) AddCounter(name string, delta float64, labels Labels) {
	_, _, _ = name, delta, labels
}

// ObserveHistogram drops a histogram observation.
func (NoopMetrics) ObserveHistogram(name string, value float64, labels Labels) {
	_, _, _ = name, value, labels
}

// SetGauge drops a gauge assignment.
func (NoopMetrics) SetGauge(name string, value float64, labels Labels) {
	_, _, _ = name, value, labels
}

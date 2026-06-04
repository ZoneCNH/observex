package observex

const (
	MetricClientCreatedTotal           = "client_created_total"
	MetricClientClosedTotal            = "client_closed_total"
	MetricClientErrorsTotal            = "client_errors_total"
	MetricClientHealthStatus           = "client_health_status"
	MetricClientHealthLatencyMS        = "client_health_latency_ms"
	MetricClientRequestsTotal          = "client_requests_total"
	MetricClientRequestDurationSeconds = "client_request_duration_seconds"
	MetricClientRetriesTotal           = "client_retries_total"
	MetricClientInflight               = "client_inflight"
)

type Metrics interface {
	IncCounter(name string, labels Labels)
	AddCounter(name string, delta float64, labels Labels)
	ObserveHistogram(name string, value float64, labels Labels)
	SetGauge(name string, value float64, labels Labels)
}

type NoopMetrics struct{}

// NewNoopMetrics returns a metrics recorder that intentionally drops all observations.
func NewNoopMetrics() NoopMetrics {
	return NoopMetrics{}
}

func (NoopMetrics) IncCounter(name string, labels Labels) {}

func (NoopMetrics) AddCounter(name string, delta float64, labels Labels) {}

func (NoopMetrics) ObserveHistogram(name string, value float64, labels Labels) {}

func (NoopMetrics) SetGauge(name string, value float64, labels Labels) {}

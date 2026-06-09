package observex

import (
	"context"
	"time"
)

// HealthReporter reports current runtime health.
type HealthReporter interface {
	HealthCheck(ctx context.Context) HealthStatus
}

// ReadinessReporter reports whether the runtime is ready to serve traffic.
type ReadinessReporter interface {
	ReadinessCheck(ctx context.Context) HealthStatus
}

// NoopHealthReporter returns deterministic healthy status for tests and defaults.
type NoopHealthReporter struct{}

// NewNoopHealthReporter returns a deterministic reporter for callers without health dependencies.
func NewNoopHealthReporter() NoopHealthReporter {
	return NoopHealthReporter{}
}

// HealthCheck returns a deterministic healthy status unless ctx is invalid.
func (NoopHealthReporter) HealthCheck(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Name:      "noop",
		Status:    HealthHealthy,
		Message:   "ok",
		CheckedAt: time.Unix(0, 0).UTC(),
	}
	if ctx == nil {
		status.Status = HealthUnhealthy
		status.Message = "context is required"
		return status
	}
	if err := ctx.Err(); err != nil {
		status.Status = HealthUnhealthy
		status.Message = err.Error()
	}
	return status
}

// ReadinessCheck delegates to HealthCheck.
func (NoopHealthReporter) ReadinessCheck(ctx context.Context) HealthStatus {
	return NoopHealthReporter{}.HealthCheck(ctx)
}

// HealthStatusValue is the normalized health state.
type HealthStatusValue string

const (
	// HealthHealthy indicates that the runtime is operating normally.
	HealthHealthy HealthStatusValue = "healthy"
	// HealthDegraded indicates that the runtime is operating with reduced confidence.
	HealthDegraded HealthStatusValue = "degraded"
	// HealthUnhealthy indicates that the runtime cannot serve normally.
	HealthUnhealthy HealthStatusValue = "unhealthy"
)

// HealthStatus describes one health or readiness check result.
type HealthStatus struct {
	// Name identifies the checked component.
	Name string `json:"name"`
	// Status is the normalized check outcome.
	Status HealthStatusValue `json:"status"`
	// Message carries optional caller-facing detail.
	Message string `json:"message,omitempty"`
	// CheckedAt records when the check completed.
	CheckedAt time.Time `json:"checked_at"`
	// LatencyMs records check duration in milliseconds.
	LatencyMs int64 `json:"latency_ms"`
	// Metadata carries sanitized low-cardinality check detail.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// HealthCheck reports the Client health state.
func (c *Client) HealthCheck(ctx context.Context) HealthStatus {
	start := time.Now()
	name := "observex"
	var metrics Metrics
	initialized := false
	closed := true
	var timeout time.Duration

	if c != nil {
		c.mu.Lock()
		name = c.cfg.Name
		metrics = c.metrics
		initialized = c.initialized
		closed = c.closed
		timeout = c.cfg.Timeout
		c.mu.Unlock()
		if name == "" {
			name = "observex"
		}
	}

	if ctx == nil {
		status := HealthStatus{
			Name:      name,
			Status:    HealthUnhealthy,
			Message:   "context is required",
			CheckedAt: time.Now(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
		recordHealthMetric(metrics, status)
		return status
	}

	if err := ctx.Err(); err != nil {
		status := HealthStatus{
			Name:      name,
			Status:    HealthUnhealthy,
			Message:   err.Error(),
			CheckedAt: time.Now(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
		recordHealthMetric(metrics, status)
		return status
	}

	if !initialized {
		status := HealthStatus{
			Name:      name,
			Status:    HealthUnhealthy,
			Message:   "client is not initialized",
			CheckedAt: time.Now(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
		recordHealthMetric(metrics, status)
		return status
	}

	if closed {
		status := HealthStatus{
			Name:      name,
			Status:    HealthUnhealthy,
			Message:   "client is closed",
			CheckedAt: time.Now(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
		recordHealthMetric(metrics, status)
		return status
	}

	if timeout > 0 {
		if deadline, ok := ctx.Deadline(); ok {
			remaining := time.Until(deadline)
			if remaining < timeout {
				status := HealthStatus{
					Name:      name,
					Status:    HealthDegraded,
					Message:   "context deadline is shorter than client timeout",
					CheckedAt: time.Now(),
					LatencyMs: time.Since(start).Milliseconds(),
					Metadata: map[string]string{
						"reason":  "deadline_below_timeout",
						"timeout": timeout.String(),
					},
				}
				recordHealthMetric(metrics, status)
				return status
			}
		}
	}

	status := HealthStatus{
		Name:      name,
		Status:    HealthHealthy,
		Message:   "ok",
		CheckedAt: time.Now(),
		LatencyMs: time.Since(start).Milliseconds(),
	}
	recordHealthMetric(metrics, status)
	return status
}

// ReadinessCheck reports whether the Client is initialized and open.
func (c *Client) ReadinessCheck(ctx context.Context) HealthStatus {
	return c.HealthCheck(ctx)
}

func recordHealthMetric(metrics Metrics, status HealthStatus) {
	if metrics == nil {
		return
	}
	labels := Labels{
		"name":   status.Name,
		"status": string(status.Status),
	}
	metrics.SetGauge(MetricClientHealthStatus, healthGaugeValue(status.Status), labels)
	metrics.ObserveHistogram(MetricClientHealthLatencyMS, float64(status.LatencyMs), labels)
}

func healthGaugeValue(status HealthStatusValue) float64 {
	if status == HealthHealthy {
		return 1
	}
	return 0
}

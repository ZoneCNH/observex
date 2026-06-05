package observex

import (
	"context"
	"strings"
	"sync"
	"time"
)

// MemoryHealthReporter records health and readiness checks in memory.
type MemoryHealthReporter struct {
	mu      sync.Mutex
	status  HealthStatus
	records []HealthStatus
}

// NewMemoryHealthReporter returns a health reporter backed by memory.
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

// HealthCheck records and returns the configured health status.
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
	r.records = append(r.records, status)
	return status
}

// ReadinessCheck records and returns the configured readiness status.
func (r *MemoryHealthReporter) ReadinessCheck(ctx context.Context) HealthStatus {
	return r.HealthCheck(ctx)
}

// SetStatus replaces the status returned by future checks.
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

// Records returns a snapshot of recorded health statuses.
func (r *MemoryHealthReporter) Records() []HealthStatus {
	if r == nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]HealthStatus, len(r.records))
	for i, record := range r.records {
		out[i] = cloneHealthStatus(record)
	}
	return out
}

func sanitizeHealthMetadata(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}

	sanitized := make(map[string]string, len(in))
	for key, value := range in {
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

package observex

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

var (
	_ Logger            = NewNoopLogger()
	_ Metrics           = NewNoopMetrics()
	_ Tracer            = NewNoopTracer()
	_ HealthReporter    = NewNoopHealthReporter()
	_ ReadinessReporter = NewNoopHealthReporter()

	_ Logger            = (*MemoryLogger)(nil)
	_ Metrics           = (*MemoryMetrics)(nil)
	_ Tracer            = (*MemoryTracer)(nil)
	_ HealthReporter    = (*MemoryHealthReporter)(nil)
	_ ReadinessReporter = (*MemoryHealthReporter)(nil)
)

func TestMemoryLoggerRedactsAndClonesRecords(t *testing.T) {
	ctx := WithTraceID(context.Background(), "trace-123")
	logger := NewMemoryLogger()

	logger.Info(ctx, "request", Secret("api_key", "raw-secret"), String("component", "api"))

	records := logger.Records()
	if len(records) != 1 {
		t.Fatalf("expected one record, got %d", len(records))
	}
	if records[0].Sequence != 1 || records[0].Level != LogLevelInfo || records[0].Message != "request" {
		t.Fatalf("unexpected record: %+v", records[0])
	}
	if fieldsContainRaw(records[0].Fields, "raw-secret") {
		t.Fatalf("logger leaked raw secret fields: %+v", records[0].Fields)
	}

	records[0].Fields[0] = String("mutated", "value")
	again := logger.Records()
	if again[0].Fields[0].Key == "mutated" {
		t.Fatal("records snapshot mutated logger state")
	}
}

func TestMemoryMetricsSanitizesLabelsAndCopiesSnapshots(t *testing.T) {
	metrics := NewMemoryMetrics()
	labels := Labels{
		"component": "api",
		"token":     "raw-secret",
	}

	metrics.IncCounter(MetricClientRequestsTotal, labels)
	metrics.SetGauge("invalid metric name", 42, labels)

	records := metrics.Records()
	if len(records) != 2 {
		t.Fatalf("expected two records, got %d", len(records))
	}
	if records[0].Name != MetricClientRequestsTotal {
		t.Fatalf("unexpected counter name: %s", records[0].Name)
	}
	if _, ok := records[0].Labels["token"]; ok {
		t.Fatalf("secret label key was not removed: %+v", records[0].Labels)
	}
	if records[1].Name != "invalid_metric_name" {
		t.Fatalf("invalid metric name was not sanitized: %s", records[1].Name)
	}

	records[0].Labels["component"] = "mutated"
	if metrics.Records()[0].Labels["component"] == "mutated" {
		t.Fatal("metric records snapshot mutated metrics state")
	}

	counters := metrics.Counters()
	if len(counters) != 1 {
		t.Fatalf("expected one counter aggregate, got %d", len(counters))
	}
}

func TestMemoryMetricsZeroValueRecords(t *testing.T) {
	var metrics MemoryMetrics
	labels := Labels{"component": "api"}

	metrics.IncCounter(MetricClientRequestsTotal, labels)
	metrics.SetGauge(MetricClientInflight, 2, labels)

	if got := len(metrics.Records()); got != 2 {
		t.Fatalf("zero-value metrics records = %d, want 2", got)
	}
	if metrics.Counters()[MetricClientRequestsTotal+"|component=api"] != 1 {
		t.Fatalf("zero-value metrics counter aggregate missing: %+v", metrics.Counters())
	}
}

func TestMemoryTracerRedactsAndEndsSpanOnce(t *testing.T) {
	tracer := NewMemoryTracer()
	_, span := tracer.Start(context.Background(), "operation", Secret("api_key", "raw-secret"))
	span.SetField(Secret("password", "raw-secret"))
	span.AddEvent("checkpoint", String("component", "worker"))
	span.End(String("token", "raw-secret"))
	span.End(String("ignored", "second-end"))

	spans := tracer.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected one span, got %d", len(spans))
	}
	if !spans[0].Ended {
		t.Fatal("expected span to be ended")
	}
	if len(spans[0].Events) != 1 {
		t.Fatalf("expected one event, got %d", len(spans[0].Events))
	}
	if fieldsContainRaw(spans[0].Fields, "raw-secret") || fieldsContainRaw(spans[0].EndFields, "raw-secret") {
		t.Fatalf("tracer leaked raw secret: %+v", spans[0])
	}
	if fieldsContainRaw(spans[0].EndFields, "second-end") {
		t.Fatalf("span accepted fields from a second End call: %+v", spans[0].EndFields)
	}

	spans[0].Events[0].Fields[0] = String("mutated", "value")
	if tracer.Spans()[0].Events[0].Fields[0].Key == "mutated" {
		t.Fatal("span snapshot mutated tracer state")
	}
}

func TestMemoryHealthReporterRecordsDeterministicSanitizedStatus(t *testing.T) {
	reporter := NewMemoryHealthReporter(HealthStatus{
		Name:   "observex",
		Status: HealthHealthy,
		Metadata: map[string]string{
			"component": "api",
			"api_key":   "raw-secret",
			"note":      "bearer raw-secret",
		},
	})

	status := reporter.HealthCheck(context.Background())
	if !status.CheckedAt.Equal(time.Unix(0, 0).UTC()) {
		t.Fatalf("expected deterministic timestamp, got %s", status.CheckedAt)
	}
	if status.Metadata["component"] != "api" {
		t.Fatalf("expected component metadata, got %+v", status.Metadata)
	}
	if _, ok := status.Metadata["api_key"]; ok {
		t.Fatalf("secret metadata key was not removed: %+v", status.Metadata)
	}
	if status.Metadata["note"] != RedactedValue {
		t.Fatalf("secret-looking metadata value was not redacted: %+v", status.Metadata)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	unhealthy := reporter.HealthCheck(ctx)
	if unhealthy.Status != HealthUnhealthy {
		t.Fatalf("expected canceled context to be unhealthy, got %s", unhealthy.Status)
	}

	readiness := reporter.ReadinessCheck(context.Background())
	if readiness.Status != HealthHealthy {
		t.Fatalf("expected readiness to be healthy, got %s", readiness.Status)
	}
	if len(reporter.Records()) != 3 {
		t.Fatalf("expected three recorded checks, got %d", len(reporter.Records()))
	}

	records := reporter.Records()
	records[0].Metadata["component"] = "mutated"
	if reporter.Records()[0].Metadata["component"] == "mutated" {
		t.Fatal("health records snapshot mutated reporter state")
	}
}

func TestMemoryImplementationsConcurrentAccess(t *testing.T) {
	logger := NewMemoryLogger()
	metrics := NewMemoryMetrics()
	tracer := NewMemoryTracer()
	health := NewMemoryHealthReporter(HealthStatus{Name: "memory", Status: HealthHealthy})

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx := WithTraceID(context.Background(), fmt.Sprintf("trace-%02d", i))
			logger.Info(ctx, "event", Int("index", i), Secret("token", "raw-secret"))
			metrics.IncCounter(MetricClientRequestsTotal, Labels{"component": "api"})
			_, span := tracer.Start(ctx, "operation", Int("index", i))
			span.AddEvent("checkpoint")
			span.End()
			_ = health.HealthCheck(ctx)
		}()
	}
	wg.Wait()

	if got := len(logger.Records()); got != 32 {
		t.Fatalf("logger records = %d, want 32", got)
	}
	if got := len(metrics.Records()); got != 32 {
		t.Fatalf("metrics records = %d, want 32", got)
	}
	if got := len(tracer.Spans()); got != 32 {
		t.Fatalf("tracer spans = %d, want 32", got)
	}
	if got := len(health.Records()); got != 32 {
		t.Fatalf("health records = %d, want 32", got)
	}
}

func fieldsContainRaw(fields []Field, raw string) bool {
	for _, field := range fields {
		if field.Key == raw || fmt.Sprint(field.Value) == raw {
			return true
		}
	}
	return false
}

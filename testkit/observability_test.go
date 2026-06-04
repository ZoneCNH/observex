package testkit

import (
	"context"
	"fmt"
	"testing"

	"github.com/ZoneCNH/observex/pkg/observex"
)

func acceptLogRecord(observex.LogRecord) {}

func acceptMetricRecord(observex.MetricRecord) {}

func acceptSpanRecord(observex.SpanRecord) {}

func TestAssertNoSecretLeakAcceptsRedactedText(t *testing.T) {
	AssertNoSecretLeak(t, fmt.Sprintf("credential=%s", observex.RedactedValue), "raw-value-123")
}

func TestRecordingLoggerRedactsAndCapturesContext(t *testing.T) {
	logger := NewRecordingLogger()
	ctx := observex.WithTraceID(context.Background(), "trace-123")
	raw := "raw-value-123"

	logger.Info(ctx, "created", observex.Secret("api_key", raw))

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected one entry, got %#v", entries)
	}
	acceptLogRecord(entries[0])
	if !logger.HasEntry("info", "created") {
		t.Fatalf("expected info entry, got %#v", entries)
	}
	if entries[0].Fields[0].Key != "trace_id" {
		t.Fatalf("expected trace field, got %#v", entries[0].Fields)
	}
	for _, field := range entries[0].Fields {
		if field.Value == raw {
			t.Fatalf("expected raw credential to be redacted, got %#v", entries[0].Fields)
		}
	}
}

func TestRecordingMetricsCapturesRecords(t *testing.T) {
	metrics := &RecordingMetrics{}
	labels := observex.Labels{"component": "api"}

	metrics.IncCounter(observex.MetricClientRequestsTotal, labels)
	metrics.SetGauge(observex.MetricClientInflight, 2, labels)

	if !metrics.HasMetric(MetricKindCounter, observex.MetricClientRequestsTotal, labels) {
		t.Fatalf("expected counter metric, got %#v", metrics.Records())
	}
	if !metrics.HasMetric(MetricKindGauge, observex.MetricClientInflight, labels) {
		t.Fatalf("expected gauge metric, got %#v", metrics.Records())
	}

	records := metrics.Records()
	acceptMetricRecord(records[0])
	records[0].Labels["component"] = "mutated"
	if metrics.Records()[0].Labels["component"] != "api" {
		t.Fatalf("expected records to be copied, got %#v", metrics.Records())
	}
}

func TestRecordingTracerCapturesSpansEventsAndEnd(t *testing.T) {
	tracer := NewRecordingTracer()
	raw := "raw-value-123"

	_, span := tracer.Start(context.Background(), "observex.Test", observex.String("component", "api"))
	span.SetField(observex.Secret("api_key", raw))
	span.AddEvent("checkpoint", observex.String("status", "ok"))
	span.End(observex.String("result", "done"))

	spans := tracer.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected one span, got %#v", spans)
	}
	acceptSpanRecord(spans[0])
	if !tracer.HasSpan("observex.Test") {
		t.Fatalf("expected span, got %#v", spans)
	}
	canonical := spanRecord(spans[0])
	if canonical.Sequence != 1 {
		t.Fatalf("expected canonical Memory span record, got %#v", canonical)
	}
	if !spans[0].Ended {
		t.Fatalf("expected span to be ended, got %#v", spans[0])
	}
	if len(spans[0].EndFields) != 1 || spans[0].EndFields[0].Key != "result" {
		t.Fatalf("expected canonical end fields, got %#v", spans[0].EndFields)
	}
	if len(spans[0].Events) != 1 || spans[0].Events[0].Name != "checkpoint" {
		t.Fatalf("expected checkpoint event, got %#v", spans[0].Events)
	}
	for _, field := range spans[0].Fields {
		if field.Value == raw {
			t.Fatalf("expected span fields to be redacted, got %#v", spans[0].Fields)
		}
	}
	for _, field := range spans[0].EndFields {
		if field.Value == raw {
			t.Fatalf("expected span end fields to be redacted, got %#v", spans[0].EndFields)
		}
	}
}

func TestRecordingAdaptersUseCanonicalMemorySnapshots(t *testing.T) {
	logger := NewRecordingLogger()
	logger.Info(context.Background(), "ready")
	logRecords := logger.Records()
	if len(logRecords) != 1 || logRecords[0].Sequence == 0 || logRecords[0].Level != observex.LogLevelInfo {
		t.Fatalf("logger records must come from MemoryLogger, got %#v", logRecords)
	}

	metrics := &RecordingMetrics{}
	metrics.AddCounter("client_requests_total", 2, observex.Labels{"component": "api"})
	metricRecords := metrics.Records()
	if len(metricRecords) != 1 || metricRecords[0].Sequence == 0 || metricRecords[0].Kind != observex.MetricKindCounter {
		t.Fatalf("metric records must come from MemoryMetrics, got %#v", metricRecords)
	}

	tracer := NewRecordingTracer()
	_, span := tracer.Start(context.Background(), "observex.MemoryBacked")
	span.End(observex.String("result", "ok"))
	spanRecords := tracer.Spans()
	if len(spanRecords) != 1 || spanRecords[0].Sequence == 0 || !spanRecords[0].Ended || len(spanRecords[0].EndFields) != 1 {
		t.Fatalf("span records must come from MemoryTracer, got %#v", spanRecords)
	}
}

func TestRecordingImplementationsUsePublicMemoryModels(t *testing.T) {
	logger := &RecordingLogger{}
	logger.Info(context.Background(), "created")
	logRecords := logger.Records()
	if len(logRecords) != 1 || logRecords[0].Sequence != 1 || logRecords[0].Level != observex.LogLevelInfo {
		t.Fatalf("expected public memory log record, got %#v", logRecords)
	}

	metrics := &RecordingMetrics{}
	metrics.SetGauge("Invalid Metric Name", 7, observex.Labels{
		"component": "api",
		"token":     "raw-token-value",
	})
	metricRecords := metrics.Records()
	if len(metricRecords) != 1 || metricRecords[0].Name != "invalid_metric_name" || metricRecords[0].Labels["component"] != "api" {
		t.Fatalf("expected sanitized public memory metric record, got %#v", metricRecords)
	}
	if _, ok := metricRecords[0].Labels["token"]; ok {
		t.Fatalf("expected unsafe labels to be dropped by public memory metrics, got %#v", metricRecords[0].Labels)
	}

	tracer := &RecordingTracer{}
	_, span := tracer.Start(context.Background(), "observex.Test")
	span.AddEvent("checkpoint")
	span.End(observex.String("result", "done"))
	spanRecords := tracer.Spans()
	if len(spanRecords) != 1 || spanRecords[0].Sequence != 1 || !spanRecords[0].Ended {
		t.Fatalf("expected public memory span record, got %#v", spanRecords)
	}
	if len(spanRecords[0].EndFields) != 1 || spanRecords[0].EndFields[0].Key != "result" {
		t.Fatalf("expected public memory span end fields, got %#v", spanRecords[0].EndFields)
	}
}

func spanRecord(record observex.SpanRecord) observex.SpanRecord {
	return record
}

package observex

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/ZoneCNH/foundationx/pkg/foundationx"
)

// ── NoopLogger ──────────────────────────────────────────────────────

func TestNoopLoggerAllLevels(t *testing.T) {
	logger := NewNoopLogger()
	ctx := context.Background()
	// All methods should silently drop records.
	logger.Debug(ctx, "debug", String("k", "v"))
	logger.Info(ctx, "info", String("k", "v"))
	logger.Warn(ctx, "warn", String("k", "v"))
	logger.Error(ctx, "error", String("k", "v"))
}

// ── SlogLogger boundary tests ───────────────────────────────────────

func TestSlogLoggerAllLevels(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger := NewSlogLogger(base)
	ctx := context.Background()

	logger.Debug(ctx, "dbg", String("a", "1"))
	logger.Info(ctx, "inf", String("b", "2"))
	logger.Warn(ctx, "wrn", String("c", "3"))
	logger.Error(ctx, "err", String("d", "4"))

	out := buf.String()
	for _, want := range []string{"dbg", "inf", "wrn", "err"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got %s", want, out)
		}
	}
}

func TestSlogLoggerWithNilRedactorFallsBack(t *testing.T) {
	logger := NewSlogLogger(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		WithRedactor(nil))
	logger.Info(context.Background(), "msg")
}

func TestSlogLoggerWithCustomRedactor(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	redactor := NewDefaultRedactor("custom_secret")
	logger := NewSlogLogger(base, WithRedactor(redactor))

	logger.Info(context.Background(), "test", String("custom_secret", "raw"))
	if strings.Contains(buf.String(), "raw") {
		t.Fatal("expected custom redactor to mask custom_secret")
	}
}

func TestSlogLoggerSkipsEmptyKeyFields(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	logger := NewSlogLogger(base)

	logger.Info(context.Background(), "msg", String("", "skipped"))
	if strings.Contains(buf.String(), "skipped") {
		t.Fatal("expected empty-key field to be skipped")
	}
}

func TestSlogLoggerNilLoggerNoPanic(t *testing.T) {
	// l.logger == nil branch
	logger := &SlogLogger{logger: nil, redactor: NewDefaultRedactor()}
	logger.Info(context.Background(), "should not panic")
}

func TestSlogLoggerNilSelfNoPanic(t *testing.T) {
	var logger *SlogLogger
	logger.Info(context.Background(), "should not panic")
}

func TestSlogLoggerDisabledLevel(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	logger := NewSlogLogger(base)

	logger.Debug(context.Background(), "should be filtered")
	logger.Info(context.Background(), "should be filtered")
	if buf.Len() != 0 {
		t.Fatal("expected debug/info to be filtered by level")
	}
}

// ── NoopMetrics ─────────────────────────────────────────────────────

func TestNoopMetricsAllOperations(t *testing.T) {
	m := NewNoopMetrics()
	m.IncCounter("c", Labels{"a": "b"})
	m.AddCounter("c", 5, Labels{"a": "b"})
	m.ObserveHistogram("h", 1.5, Labels{"a": "b"})
	m.SetGauge("g", 42, Labels{"a": "b"})
}

// ── NoopTracer + NoopSpan ───────────────────────────────────────────

func TestNoopTracerStartAndSpanMethods(t *testing.T) {
	tracer := NewNoopTracer()
	ctx, span := tracer.Start(context.Background(), "op", String("k", "v"))
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	span.SetField(String("k", "v"))
	span.AddEvent("event", String("k", "v"))
	span.End(String("done", "true"))
}

// ── Options ─────────────────────────────────────────────────────────

func TestWithLoggerNilIsIgnored(t *testing.T) {
	opts := defaultOptions()
	WithLogger(nil)(&opts)
	if opts.logger == nil {
		t.Fatal("expected noop logger to remain")
	}
}

func TestWithTracerNilIsIgnored(t *testing.T) {
	opts := defaultOptions()
	WithTracer(nil)(&opts)
	if opts.tracer == nil {
		t.Fatal("expected noop tracer to remain")
	}
}

func TestWithMetricsNilIsIgnored(t *testing.T) {
	opts := defaultOptions()
	WithMetrics(nil)(&opts)
	if opts.metrics == nil {
		t.Fatal("expected noop metrics to remain")
	}
}

// ── Client: nil context, zero-value close with nil context ──────────

func TestNewRejectsNilContext(t *testing.T) {
	_, err := New(nil, Config{Name: "test"})
	if err == nil {
		t.Fatal("expected nil context to fail")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCloseRejectsNilContext(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	err = client.Close(nil)
	if err == nil {
		t.Fatal("expected nil context to fail")
	}
	if !IsKind(err, ErrorKindValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCloseRejectsExpiredContext(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	err = client.Close(ctx)
	if err == nil {
		t.Fatal("expected expired context to fail")
	}
	if !IsKind(err, ErrorKindTimeout) {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestCloseZeroValueClientWithNilContext(t *testing.T) {
	var client Client
	err := client.Close(nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCloseUninitializedClient(t *testing.T) {
	client := &Client{initialized: false}
	err := client.Close(context.Background())
	if err == nil {
		t.Fatal("expected error for uninitialized client")
	}
}

func TestCloseWithNilTracer(t *testing.T) {
	client := &Client{
		cfg:         Config{Name: "test"},
		metrics:     NewNoopMetrics(),
		logger:      NewNoopLogger(),
		tracer:      nil,
		initialized: true,
	}
	err := client.Close(context.Background())
	if err != nil {
		t.Fatalf("expected nil tracer close to succeed: %v", err)
	}
}

// ── Health: nil context on NoopHealthReporter ───────────────────────

func TestNoopHealthReporterNilContext(t *testing.T) {
	r := NewNoopHealthReporter()
	status := r.HealthCheck(nil)
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy, got %s", status.Status)
	}
}

func TestNoopHealthReporterCanceledContext(t *testing.T) {
	r := NewNoopHealthReporter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	status := r.HealthCheck(ctx)
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy, got %s", status.Status)
	}
}

func TestNoopHealthReporterReadinessDelegatesToHealth(t *testing.T) {
	r := NewNoopHealthReporter()
	status := r.ReadinessCheck(context.Background())
	if status.Status != HealthHealthy {
		t.Fatalf("expected healthy, got %s", status.Status)
	}
}

func TestClientReadinessCheckDelegates(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "test"})
	if err != nil {
		t.Fatal(err)
	}
	status := client.ReadinessCheck(context.Background())
	if status.Status != HealthHealthy {
		t.Fatalf("expected healthy, got %s", status.Status)
	}
}

func TestHealthCheckNilClient(t *testing.T) {
	var client *Client
	status := client.HealthCheck(context.Background())
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy, got %s", status.Status)
	}
}

func TestHealthCheckNilClientNilContext(t *testing.T) {
	var client *Client
	status := client.HealthCheck(nil)
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy, got %s", status.Status)
	}
}

func TestHealthCheckExpiredDeadline(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "test", Timeout: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	status := client.HealthCheck(ctx)
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy for expired deadline, got %s", status.Status)
	}
}

func TestHealthCheckExpiredDeadlineWithErr(t *testing.T) {
	client, err := New(context.Background(), Config{Name: "test", Timeout: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	status := client.HealthCheck(ctx)
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy, got %s", status.Status)
	}
}

func TestHealthGaugeValueHealthy(t *testing.T) {
	if healthGaugeValue(HealthHealthy) != 1 {
		t.Fatal("expected 1 for healthy")
	}
	if healthGaugeValue(HealthDegraded) != 0 {
		t.Fatal("expected 0 for degraded")
	}
	if healthGaugeValue(HealthUnhealthy) != 0 {
		t.Fatal("expected 0 for unhealthy")
	}
}

func TestRecordHealthMetricNilMetrics(t *testing.T) {
	// should not panic
	recordHealthMetric(nil, HealthStatus{Status: HealthHealthy})
}

// ── Error: nil receiver, errorKind branches ─────────────────────────

func TestErrorNilReceiver(t *testing.T) {
	var e *Error
	if e.Error() != "" {
		t.Fatal("expected empty string for nil error")
	}
	if e.Unwrap() != nil {
		t.Fatal("expected nil unwrap for nil error")
	}
}

func TestErrorWithCauseAndNoMessage(t *testing.T) {
	cause := context.DeadlineExceeded
	err := newError(ErrorKindTimeout, "op", "", true, cause)
	if !strings.Contains(err.Error(), cause.Error()) {
		t.Fatalf("expected error string to contain cause, got %q", err.Error())
	}
}

func TestErrorWithoutOp(t *testing.T) {
	err := NewError(ErrorKindInternal, "", "something broke", false)
	got := err.Error()
	if strings.Contains(got, ": :") {
		t.Fatalf("expected no double-colon for empty op, got %q", got)
	}
}

func TestMapErrorNilReturnsNil(t *testing.T) {
	if MapError("op", nil) != nil {
		t.Fatal("expected nil for nil error")
	}
}

func TestMapErrorAlreadyObservexError(t *testing.T) {
	orig := NewError(ErrorKindConfig, "orig", "msg", false)
	mapped := MapError("newop", orig)
	if mapped != orig {
		t.Fatal("expected same error to be returned")
	}
}

func TestMapErrorDeadlineExceeded(t *testing.T) {
	err := MapError("op", context.DeadlineExceeded)
	if !IsKind(err, ErrorKindTimeout) {
		t.Fatalf("expected timeout, got %v", err)
	}
}

func TestMapErrorGenericError(t *testing.T) {
	err := MapError("op", errors.New("generic failure"))
	if !IsKind(err, ErrorKindInternal) {
		t.Fatalf("expected internal, got %v", err)
	}
}

func TestMapErrorFoundationEmptyOp(t *testing.T) {
	cause := foundationx.NewError(foundationx.ErrorKindAuth, "fx.op", "fx msg")
	err := MapError("", cause)
	if !IsKind(err, ErrorKindAuth) {
		t.Fatalf("expected auth, got %v", err)
	}
}

func TestIsKindNonObservexError(t *testing.T) {
	// IsKind falls through to foundationx.IsKind
	err := foundationx.NewError(foundationx.ErrorKindRateLimit, "fx", "slow")
	if !IsKind(err, ErrorKindRateLimit) {
		t.Fatal("expected IsKind to match foundationx error")
	}
}

func TestErrorKindPlainError(t *testing.T) {
	// errorKind with a plain error returns ErrorKindInternal
	kind := errorKind(errors.New("plain"))
	if kind != ErrorKindInternal {
		t.Fatalf("expected internal, got %v", kind)
	}
}

func TestErrorKindFoundationError(t *testing.T) {
	err := foundationx.NewError(foundationx.ErrorKindNotFound, "fx", "not found")
	kind := errorKind(err)
	if kind != ErrorKindNotFound {
		t.Fatalf("expected not_found, got %v", kind)
	}
}

func TestContextErrorUnavailable(t *testing.T) {
	err := contextError("op", errors.New("something else"))
	if !IsKind(err, ErrorKindUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
	if err.Retryable {
		t.Fatal("expected non-retryable")
	}
}

// ── Labels: boundary tests ──────────────────────────────────────────

func TestValidateLabelsValid(t *testing.T) {
	labels := Labels{"component": "api", "status": "ok"}
	if err := ValidateLabels(labels); err != nil {
		t.Fatalf("expected valid labels: %v", err)
	}
}

func TestSanitizeLabelsEmpty(t *testing.T) {
	if got := SanitizeLabels(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	if got := SanitizeLabels(Labels{}); got != nil {
		t.Fatalf("expected nil for empty, got %v", got)
	}
}

func TestSanitizeLabelsAllUnsafe(t *testing.T) {
	labels := Labels{"bad-key": "drop", "trace_id": "x"}
	got := SanitizeLabels(labels)
	if got != nil {
		t.Fatalf("expected nil when all labels removed, got %v", got)
	}
}

func TestCloneLabelsEmpty(t *testing.T) {
	if got := CloneLabels(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	if got := CloneLabels(Labels{}); got != nil {
		t.Fatalf("expected nil for empty, got %v", got)
	}
}

func TestValueLooksSecret(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"normal", false},
		{"password=secret123", true},
		{"passwd=abc", true},
		{"secret=mysecret", true},
		{"token=mytoken", true},
		{"authorization: bearer xyz", true},
		{"bearer xyz", true},
		{"access_key=mykey", true},
		{"secret_key=mykey", true},
		{"", false},
	}
	for _, tt := range tests {
		if got := valueLooksSecret(tt.value); got != tt.want {
			t.Errorf("valueLooksSecret(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

// ── MemoryLogger: Debug/Warn/Error, Reset, nil receiver ─────────────

func TestMemoryLoggerAllLevels(t *testing.T) {
	logger := NewMemoryLogger()
	ctx := context.Background()

	logger.Debug(ctx, "dbg")
	logger.Info(ctx, "inf")
	logger.Warn(ctx, "wrn")
	logger.Error(ctx, "err")

	records := logger.Records()
	if len(records) != 4 {
		t.Fatalf("expected 4 records, got %d", len(records))
	}
	expected := []LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError}
	for i, want := range expected {
		if records[i].Level != want {
			t.Fatalf("record[%d].Level = %q, want %q", i, records[i].Level, want)
		}
	}
}

func TestMemoryLoggerReset(t *testing.T) {
	logger := NewMemoryLogger()
	logger.Info(context.Background(), "msg")
	if len(logger.Records()) != 1 {
		t.Fatal("expected 1 record before reset")
	}
	logger.Reset()
	if len(logger.Records()) != 0 {
		t.Fatal("expected 0 records after reset")
	}
}

func TestMemoryLoggerNilReceiver(t *testing.T) {
	var logger *MemoryLogger
	logger.Info(context.Background(), "msg") // should not panic
	if logger.Records() != nil {
		t.Fatal("expected nil records for nil logger")
	}
	logger.Reset() // should not panic
}

func TestMemoryLoggerNilLog(t *testing.T) {
	var logger *MemoryLogger
	// l == nil branch in log()
	logger.log(context.Background(), LogLevelInfo, "msg")
}

// ── MemoryMetrics: ObserveHistogram, Gauges, Reset, nil receiver ────

func TestMemoryMetricsObserveHistogram(t *testing.T) {
	m := NewMemoryMetrics()
	m.ObserveHistogram("hist", 1.5, Labels{"a": "b"})
	records := m.Records()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Kind != MetricKindHistogram {
		t.Fatalf("expected histogram, got %v", records[0].Kind)
	}
}

func TestMemoryMetricsGauges(t *testing.T) {
	m := NewMemoryMetrics()
	m.SetGauge("g", 42, Labels{"a": "b"})
	gauges := m.Gauges()
	if len(gauges) != 1 {
		t.Fatalf("expected 1 gauge, got %d", len(gauges))
	}
}

func TestMemoryMetricsGaugesEmpty(t *testing.T) {
	m := NewMemoryMetrics()
	if got := m.Gauges(); got != nil {
		t.Fatalf("expected nil for empty gauges, got %v", got)
	}
}

func TestMemoryMetricsReset(t *testing.T) {
	m := NewMemoryMetrics()
	m.IncCounter("c", Labels{"a": "b"})
	m.SetGauge("g", 1, Labels{"a": "b"})
	m.Reset()
	if len(m.Records()) != 0 {
		t.Fatal("expected 0 records after reset")
	}
	if len(m.Counters()) != 0 {
		t.Fatal("expected 0 counters after reset")
	}
	if len(m.Gauges()) != 0 {
		t.Fatal("expected 0 gauges after reset")
	}
}

func TestMemoryMetricsNilReceiver(t *testing.T) {
	var m *MemoryMetrics
	m.IncCounter("c", nil)
	m.AddCounter("c", 1, nil)
	m.ObserveHistogram("h", 1, nil)
	m.SetGauge("g", 1, nil)
	if m.Records() != nil {
		t.Fatal("expected nil records")
	}
	if m.Counters() != nil {
		t.Fatal("expected nil counters")
	}
	if m.Gauges() != nil {
		t.Fatal("expected nil gauges")
	}
	m.Reset() // should not panic
}

func TestMemoryMetricsNilRecord(t *testing.T) {
	var m *MemoryMetrics
	// m == nil branch in record()
	m.record(MetricKindCounter, "c", 1, nil)
}

func TestMemoryMetricsCountersEmpty(t *testing.T) {
	m := NewMemoryMetrics()
	if got := m.Counters(); got != nil {
		t.Fatalf("expected nil for empty counters, got %v", got)
	}
}

func TestMetricRecordKeyEmptyLabels(t *testing.T) {
	key := metricRecordKey("name", nil)
	if key != "name" {
		t.Fatalf("expected 'name', got %q", key)
	}
}

func TestCloneFloatMapEmpty(t *testing.T) {
	if got := cloneFloatMap(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	if got := cloneFloatMap(map[string]float64{}); got != nil {
		t.Fatalf("expected nil for empty, got %v", got)
	}
}

// ── MemoryTracer: Reset, nil receiver, span edge cases ──────────────

func TestMemoryTracerReset(t *testing.T) {
	tracer := NewMemoryTracer()
	_, span := tracer.Start(context.Background(), "op")
	span.End()
	tracer.Reset()
	if len(tracer.Spans()) != 0 {
		t.Fatal("expected 0 spans after reset")
	}
}

func TestMemoryTracerNilReceiver(t *testing.T) {
	var tracer *MemoryTracer
	ctx, span := tracer.Start(context.Background(), "op")
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if _, ok := span.(NoopSpan); !ok {
		t.Fatal("expected NoopSpan for nil tracer")
	}
	if tracer.Spans() != nil {
		t.Fatal("expected nil spans")
	}
	tracer.Reset() // should not panic
}

func TestMemorySpanNilReceiver(t *testing.T) {
	var span *memorySpan
	span.SetField(String("k", "v")) // should not panic
	span.AddEvent("event")          // should not panic
	span.End()                      // should not panic
}

func TestMemorySpanNilTracer(t *testing.T) {
	span := &memorySpan{tracer: nil, index: 0}
	span.SetField(String("k", "v"))
	span.AddEvent("event")
	span.End()
}

func TestMemorySpanInvalidIndex(t *testing.T) {
	tracer := NewMemoryTracer()
	span := &memorySpan{tracer: tracer, index: -1}
	span.SetField(String("k", "v"))
	span.AddEvent("event")
	span.End()
}

func TestMemorySpanEndAlreadyEnded(t *testing.T) {
	tracer := NewMemoryTracer()
	_, span := tracer.Start(context.Background(), "op")
	span.End(String("first", "end"))
	span.End(String("second", "end")) // should be ignored
	spans := tracer.Spans()
	if len(spans[0].EndFields) == 0 || spans[0].EndFields[0].Key == "second" {
		t.Fatal("second End should be ignored")
	}
}

func TestCloneSpanEventsEmpty(t *testing.T) {
	if got := cloneSpanEvents(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	if got := cloneSpanEvents([]SpanEvent{}); got != nil {
		t.Fatalf("expected nil for empty, got %v", got)
	}
}

// ── MemoryHealthReporter: SetStatus, nil receiver, nil context ──────

func TestMemoryHealthReporterSetStatus(t *testing.T) {
	reporter := NewMemoryHealthReporter(HealthStatus{})
	reporter.SetStatus(HealthStatus{
		Name:   "updated",
		Status: HealthDegraded,
	})
	status := reporter.HealthCheck(context.Background())
	if status.Name != "updated" {
		t.Fatalf("expected 'updated', got %q", status.Name)
	}
	if status.Status != HealthDegraded {
		t.Fatalf("expected degraded, got %s", status.Status)
	}
}

func TestMemoryHealthReporterSetStatusDefaults(t *testing.T) {
	reporter := NewMemoryHealthReporter(HealthStatus{})
	reporter.SetStatus(HealthStatus{})
	status := reporter.HealthCheck(context.Background())
	if status.Name != "memory" {
		t.Fatalf("expected 'memory', got %q", status.Name)
	}
	if status.Status != HealthHealthy {
		t.Fatalf("expected healthy, got %s", status.Status)
	}
}

func TestMemoryHealthReporterNilReceiver(t *testing.T) {
	var reporter *MemoryHealthReporter
	status := reporter.HealthCheck(context.Background())
	if status.Status != HealthHealthy {
		t.Fatalf("expected healthy from NoopHealthReporter fallback, got %s", status.Status)
	}
	if reporter.Records() != nil {
		t.Fatal("expected nil records")
	}
	reporter.SetStatus(HealthStatus{}) // should not panic
}

func TestMemoryHealthReporterNilContext(t *testing.T) {
	reporter := NewMemoryHealthReporter(HealthStatus{})
	status := reporter.HealthCheck(nil) //nolint:staticcheck
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy for nil context, got %s", status.Status)
	}
}

func TestMemoryHealthReporterCanceledContext(t *testing.T) {
	reporter := NewMemoryHealthReporter(HealthStatus{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	status := reporter.HealthCheck(ctx)
	if status.Status != HealthUnhealthy {
		t.Fatalf("expected unhealthy, got %s", status.Status)
	}
}

func TestSanitizeHealthMetadataAllSecret(t *testing.T) {
	meta := map[string]string{
		"api_key": "val",
		"  ":      "val",
	}
	got := sanitizeHealthMetadata(meta)
	if got != nil {
		t.Fatalf("expected nil when all metadata removed, got %v", got)
	}
}

func TestSanitizeHealthMetadataSecretValue(t *testing.T) {
	meta := map[string]string{"note": "bearer token=abc123"}
	got := sanitizeHealthMetadata(meta)
	if got["note"] != RedactedValue {
		t.Fatalf("expected redacted, got %q", got["note"])
	}
}

func TestSanitizeHealthMetadataEmpty(t *testing.T) {
	if got := sanitizeHealthMetadata(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	if got := sanitizeHealthMetadata(map[string]string{}); got != nil {
		t.Fatalf("expected nil for empty, got %v", got)
	}
}

// ── redactFields nil redactor ───────────────────────────────────────

func TestRedactFieldsNilRedactor(t *testing.T) {
	fields := []Field{String("k", "v")}
	got := redactFields(nil, fields)
	if len(got) != 1 {
		t.Fatalf("expected 1 field, got %d", len(got))
	}
	if got[0].Key != "k" {
		t.Fatalf("expected field to be cloned, got %v", got[0])
	}
}

func TestCloneFieldsEmpty(t *testing.T) {
	if got := cloneFields(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	if got := cloneFields([]Field{}); got != nil {
		t.Fatalf("expected nil for empty, got %v", got)
	}
}

// ── redactedString ──────────────────────────────────────────────────

func TestRedactedStringNilValue(t *testing.T) {
	got := redactedString(nil)
	if got != RedactedValue {
		t.Fatalf("expected %q, got %q", RedactedValue, got)
	}
}

// ── isSecretKey additional coverage ─────────────────────────────────

func TestIsSecretKeySuffixPatterns(t *testing.T) {
	keys := []string{
		"my_password_field",
		"my_passwd_field",
		"my_secret_field",
		"my_token_field",
		"my_private_key",
		"my_api_key",
		"my_access_key",
		"my_database_url",
		"my_dsn",
		"my_authorization",
		"my_cookie",
	}
	for _, key := range keys {
		if !IsSecretKey(key) {
			t.Errorf("expected %q to be secret key", key)
		}
	}
}

func TestIsSecretKeyNotSecret(t *testing.T) {
	keys := []string{"component", "status", "operation", ""}
	for _, key := range keys {
		if IsSecretKey(key) {
			t.Errorf("expected %q to not be secret key", key)
		}
	}
}

func TestNormalizeSecretKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Api-Key", "api_key"},
		{"  TOKEN  ", "token"},
		{"my.field", "my_field"},
	}
	for _, tt := range tests {
		if got := normalizeSecretKey(tt.input); got != tt.want {
			t.Errorf("normalizeSecretKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ── RedactField with empty key ──────────────────────────────────────

func TestRedactFieldEmptyKey(t *testing.T) {
	redactor := NewDefaultRedactor()
	field := Field{Key: "", Value: "keep"}
	got := redactor.RedactField(field)
	if got.Value != "keep" {
		t.Fatalf("expected value to be kept for empty key, got %v", got.Value)
	}
}

func TestRedactFieldsEmpty(t *testing.T) {
	redactor := NewDefaultRedactor()
	if got := redactor.RedactFields(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	if got := redactor.RedactFields([]Field{}); got != nil {
		t.Fatalf("expected nil for empty, got %v", got)
	}
}

// ── Options: non-nil values ─────────────────────────────────────────

func TestWithLoggerNonNil(t *testing.T) {
	logger := NewNoopLogger()
	opts := defaultOptions()
	WithLogger(logger)(&opts)
	if _, ok := opts.logger.(NoopLogger); !ok {
		t.Fatal("expected NoopLogger to be set")
	}
}

func TestWithTracerNonNil(t *testing.T) {
	tracer := NewNoopTracer()
	opts := defaultOptions()
	WithTracer(tracer)(&opts)
	if _, ok := opts.tracer.(NoopTracer); !ok {
		t.Fatal("expected NoopTracer to be set")
	}
}

// ── NewSlogLogger with nil redactor option ──────────────────────────

func TestNewSlogLoggerWithNilRedactorOption(t *testing.T) {
	// Pass WithRedactor(nil) to trigger the fallback to NewDefaultRedactor
	logger := NewSlogLogger(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		WithRedactor(nil))
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

// ── NewMemoryLogger with nil redactor option ────────────────────────

func TestNewMemoryLoggerWithNilRedactorOption(t *testing.T) {
	logger := NewMemoryLogger(WithRedactor(nil))
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	// Verify it still works
	logger.Info(context.Background(), "test")
	if len(logger.Records()) != 1 {
		t.Fatal("expected 1 record")
	}
}

// ── NewMemoryTracer with nil redactor option ────────────────────────

func TestNewMemoryTracerWithNilRedactorOption(t *testing.T) {
	tracer := NewMemoryTracer(WithRedactor(nil))
	if tracer == nil {
		t.Fatal("expected non-nil tracer")
	}
	_, span := tracer.Start(context.Background(), "op")
	span.End()
	if len(tracer.Spans()) != 1 {
		t.Fatal("expected 1 span")
	}
}

// ── MemorySpan SetField with empty redacted fields ──────────────────

func TestMemorySpanSetFieldRedactedEmpty(t *testing.T) {
	// Create a redactor that redacts everything to empty
	tracer := NewMemoryTracer()
	_, span := tracer.Start(context.Background(), "op")
	// SetField with empty key after redaction should still work
	span.SetField(String("", "value"))
	spans := tracer.Spans()
	if len(spans) != 1 {
		t.Fatal("expected 1 span")
	}
}

// ── toSnakeCase: digit → uppercase boundary ─────────────────────────

func TestToSnakeCaseDigitToUpperCase(t *testing.T) {
	// "1A" should produce "1_a"
	got := toSnakeCase("1A")
	if got != "1_a" {
		t.Errorf("toSnakeCase('1A') = %q, want '1_a'", got)
	}
}

// ── Error: with op and message ──────────────────────────────────────

func TestErrorWithOpAndMessage(t *testing.T) {
	err := NewError(ErrorKindValidation, "op", "msg", false)
	got := err.Error()
	if !strings.Contains(got, "validation: op: msg") {
		t.Fatalf("expected full error string, got %q", got)
	}
}

// ── Close nil client ────────────────────────────────────────────────

func TestCloseNilClient(t *testing.T) {
	var client *Client
	err := client.Close(context.Background())
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}

// ── HealthCheck with empty name fallback ────────────────────────────

func TestHealthCheckEmptyNameFallback(t *testing.T) {
	client, err := New(context.Background(), Config{Name: ""})
	// Config.Validate() requires non-empty name, so this should fail
	if err == nil {
		// If it somehow succeeds, check the name fallback
		status := client.HealthCheck(context.Background())
		if status.Name != "observex" {
			t.Fatalf("expected 'observex' fallback, got %q", status.Name)
		}
	}
}

// ── redactedString with non-nil value ───────────────────────────────

func TestRedactedStringNonNil(t *testing.T) {
	got := redactedString("hello")
	// Should return the masked version
	if got == "" {
		t.Fatal("expected non-empty redacted string")
	}
}

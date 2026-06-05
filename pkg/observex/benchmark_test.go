package observex

import (
	"context"
	"testing"
	"time"
)

func BenchmarkNew(b *testing.B) {
	ctx := context.Background()
	cfg := Config{Name: "bench", Timeout: time.Second}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		client, err := New(ctx, cfg)
		if err != nil {
			b.Fatal(err)
		}
		if err := client.Close(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemoryLoggerInfo(b *testing.B) {
	logger := NewMemoryLogger()
	ctx := WithTraceID(context.Background(), "trace-bench")
	fields := []Field{
		String("component", "bench"),
		Secret("api_key", "raw-secret"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info(ctx, "request", fields...)
		if i%1024 == 0 {
			logger.Reset()
		}
	}
}

func BenchmarkMemoryMetricsIncCounter(b *testing.B) {
	metrics := NewMemoryMetrics()
	labels := Labels{"component": "bench"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.IncCounter(MetricClientRequestsTotal, labels)
		if i%1024 == 0 {
			metrics.Reset()
		}
	}
}

func BenchmarkHealthCheck(b *testing.B) {
	ctx := context.Background()
	client, err := New(ctx, Config{Name: "bench", Timeout: time.Second})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		_ = client.Close(context.Background())
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		status := client.HealthCheck(ctx)
		if status.Status != HealthHealthy {
			b.Fatalf("health status = %s", status.Status)
		}
	}
}

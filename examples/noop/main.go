package main

import (
	"context"
	"fmt"

	"github.com/ZoneCNH/observex/pkg/observex"
)

func main() {
	logger := observex.NewNoopLogger()
	metrics := observex.NewNoopMetrics()
	tracer := observex.NewNoopTracer()
	health := observex.NewNoopHealthReporter()

	ctx, span := tracer.Start(context.Background(), "noop.example")
	logger.Info(ctx, "noop ready")
	metrics.IncCounter(observex.MetricClientRequestsTotal, observex.Labels{"component": "noop"})
	span.End()

	status := health.HealthCheck(ctx)
	fmt.Println("noop", status.Status)
}

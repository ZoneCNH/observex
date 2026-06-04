package main

import (
	"context"
	"fmt"

	"github.com/ZoneCNH/observex/pkg/observex"
)

func main() {
	reporter := observex.NewMemoryHealthReporter(observex.HealthStatus{
		Name:   "example",
		Status: observex.HealthHealthy,
	})
	status := reporter.HealthCheck(context.Background())
	fmt.Println(status.Status)
}

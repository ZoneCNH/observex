package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ZoneCNH/observex/pkg/observex"
)

func formatLabels(labels observex.Labels) string {
	if len(labels) == 0 {
		return "-"
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, labels[key]))
	}
	return strings.Join(parts, ",")
}

func main() {
	metrics := observex.NewMemoryMetrics()
	metrics.IncCounter(observex.MetricClientRequestsTotal, observex.Labels{"component": "api"})
	for _, record := range metrics.Records() {
		fmt.Printf("%s %s %s value=%.0f\n", record.Kind, record.Name, formatLabels(record.Labels), record.Value)
	}
}

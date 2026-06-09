package observex

import "testing"

func TestMetricName(t *testing.T) {
	tests := []struct {
		name      string
		module    string
		operation string
		measure   string
		want      string
	}{
		{name: "simple", module: "market-data", operation: "fetch", measure: "latency_ms", want: "foundationx_market_data_fetch_latency_ms"},
		{name: "camelCase module", module: "MarketData", operation: "Fetch", measure: "Total", want: "foundationx_market_data_fetch_total"},
		{name: "with spaces", module: "risk engine", operation: "evaluate", measure: "duration_seconds", want: "foundationx_risk_engine_evaluate_duration_seconds"},
		{name: "empty operation", module: "kernel", operation: "", measure: "total", want: "foundationx_kernel_total"},
		{name: "only module", module: "binance", operation: "", measure: "", want: "foundationx_binance"},
		{name: "all empty", module: "", operation: "", measure: "", want: "foundationx"},
		{name: "dots in input", module: "market.data", operation: "fetch.order", measure: "count.total", want: "foundationx_market_data_fetch_order_count_total"},
		{name: "mixed case", module: "RiskEngine", operation: "EvaluateSignal", measure: "DurationSeconds", want: "foundationx_risk_engine_evaluate_signal_duration_seconds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MetricName(tt.module, tt.operation, tt.measure)
			if got != tt.want {
				t.Errorf("MetricName(%q, %q, %q) = %q, want %q",
					tt.module, tt.operation, tt.measure, got, tt.want)
			}
			// Verify result matches metric name regex.
			if err := ValidateMetricName(got); err != nil {
				t.Errorf("MetricName result %q failed ValidateMetricName: %v", got, err)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "CamelCase", want: "camel_case"},
		{input: "PascalCase", want: "pascal_case"},
		{input: "kebab-case", want: "kebab_case"},
		{input: "dot.separated", want: "dot_separated"},
		{input: "space separated", want: "space_separated"},
		{input: "already_snake", want: "already_snake"},
		{input: "HTTPRequest", want: "http_request"},
		{input: "getHTTPResponse", want: "get_http_response"},
		{input: "ABC", want: "abc"},
		{input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

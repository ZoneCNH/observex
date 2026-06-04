package observex

import "testing"

func TestValidateMetricName(t *testing.T) {
	valid := []string{
		"client_requests_total",
		"request_duration_seconds",
		"health1_status",
	}
	for _, name := range valid {
		if err := ValidateMetricName(name); err != nil {
			t.Fatalf("expected %q to be valid: %v", name, err)
		}
	}

	invalid := []string{
		"",
		"ClientRequests",
		"1_requests_total",
		"client__requests",
		"client-requests",
	}
	for _, name := range invalid {
		if err := ValidateMetricName(name); err == nil {
			t.Fatalf("expected %q to be invalid", name)
		}
	}
}

func TestValidateLabelsRejectsUnsafeLabels(t *testing.T) {
	tests := []Labels{
		{"TraceID": "abc"},
		{"trace_id": "abc"},
		{"api_key": "raw-value-123"},
		{"component": "token" + "=" + "raw-value-123"},
	}

	for _, labels := range tests {
		if err := ValidateLabels(labels); err == nil {
			t.Fatalf("expected labels to be rejected: %#v", labels)
		}
	}
}

func TestSanitizeLabelsDropsUnsafeKeysAndRedactsValues(t *testing.T) {
	labels := Labels{
		"component": "api",
		"trace_id":  "trace-1",
		"api_key":   "raw-value-123",
		"source":    "token" + "=" + "raw-value-123",
		"bad-key":   "drop",
	}

	got := SanitizeLabels(labels)
	if got["component"] != "api" {
		t.Fatalf("expected component label to be preserved, got %#v", got)
	}
	if got["source"] != RedactedValue {
		t.Fatalf("expected secret-looking value to be redacted, got %#v", got)
	}
	for _, key := range []string{"trace_id", "api_key", "bad-key"} {
		if _, ok := got[key]; ok {
			t.Fatalf("expected key %q to be dropped, got %#v", key, got)
		}
	}
}

func TestCloneLabelsCopiesInput(t *testing.T) {
	labels := Labels{"component": "api"}
	got := CloneLabels(labels)
	got["component"] = "worker"

	if labels["component"] != "api" {
		t.Fatalf("expected original labels to remain unchanged, got %#v", labels)
	}
}

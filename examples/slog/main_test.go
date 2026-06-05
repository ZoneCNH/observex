package main

import (
	"strings"
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestMainPrintsSlogJSONWithRedaction(t *testing.T) {
	output := testkit.CaptureStdout(t, main)
	for _, fragment := range []string{
		`"msg":"client ready"`,
		`"trace_id":"trace-example"`,
		`"api_key":"***"`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("output missing %s: %s", fragment, output)
		}
	}
	if strings.Contains(output, "raw-value-123") {
		t.Fatalf("output leaked raw value: %s", output)
	}
}

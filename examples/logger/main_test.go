package main

import (
	"bytes"
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestMainPrintsRedactedLoggerRecord(t *testing.T) {
	output := testkit.CaptureStdout(t, main)
	want := "info worker ready api_key=*** component=example correlation_id=corr-example\n"
	if output != want {
		t.Fatalf("unexpected output:\nactual: %q\nwant:   %q", output, want)
	}
	if bytes.Contains([]byte(output), []byte("raw-value-123")) {
		t.Fatalf("output leaked raw value: %q", output)
	}
}

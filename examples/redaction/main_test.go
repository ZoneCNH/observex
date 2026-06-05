package main

import (
	"bytes"
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestMainPrintsMaskedValue(t *testing.T) {
	output := testkit.CaptureStdout(t, main)
	if output != "***\n" {
		t.Fatalf("unexpected output: %q", output)
	}
	if bytes.Contains([]byte(output), []byte("raw-value-123")) {
		t.Fatalf("output leaked raw value: %q", output)
	}
}

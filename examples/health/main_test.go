package main

import (
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestMainPrintsHealthyStatus(t *testing.T) {
	output := testkit.CaptureStdout(t, main)
	if output != "healthy\n" {
		t.Fatalf("unexpected output: %q", output)
	}
}

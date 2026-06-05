package main

import (
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestMainPrintsNoopStatus(t *testing.T) {
	output := testkit.CaptureStdout(t, main)
	const want = "noop healthy\n"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

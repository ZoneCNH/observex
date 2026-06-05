package main

import (
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestMainPrintsCounterSample(t *testing.T) {
	output := testkit.CaptureStdout(t, main)
	want := "counter client_requests_total component=api value=1\n"
	if output != want {
		t.Fatalf("unexpected output:\nactual: %q\nwant:   %q", output, want)
	}
}

package main

import (
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestMainPrintsSpanLifecycle(t *testing.T) {
	output := testkit.CaptureStdout(t, main)
	want := "span observex.example\nevent checkpoint\nspan_end observex.example\n"
	if output != want {
		t.Fatalf("unexpected output:\nactual: %q\nwant:   %q", output, want)
	}
}

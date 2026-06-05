package main

import (
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestMainPrintsRedactedSecret(t *testing.T) {
	output := testkit.CaptureStdout(t, main)
	if output != "***\n" {
		t.Fatalf("unexpected output: %q", output)
	}
}

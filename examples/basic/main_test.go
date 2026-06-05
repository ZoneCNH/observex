package main

import (
	"testing"

	"github.com/ZoneCNH/observex/testkit"
)

func TestMainPrintsModuleName(t *testing.T) {
	output := testkit.CaptureStdout(t, main)
	if output != "github.com/ZoneCNH/observex\n" {
		t.Fatalf("unexpected output: %q", output)
	}
}

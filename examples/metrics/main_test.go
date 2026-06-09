package main

import (
	"testing"

	"github.com/ZoneCNH/observex/pkg/observex"
	"github.com/ZoneCNH/observex/testkit"
)

func TestMainPrintsCounterSample(t *testing.T) {
	output := testkit.CaptureStdout(t, main)
	want := "counter client_requests_total component=api value=1\n"
	if output != want {
		t.Fatalf("unexpected output:\nactual: %q\nwant:   %q", output, want)
	}
}

func TestFormatLabelsEmpty(t *testing.T) {
	got := formatLabels(nil)
	if got != "-" {
		t.Fatalf("expected '-' for nil labels, got %q", got)
	}
	got = formatLabels(observex.Labels{})
	if got != "-" {
		t.Fatalf("expected '-' for empty labels, got %q", got)
	}
}

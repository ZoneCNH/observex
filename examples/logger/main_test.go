package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/ZoneCNH/observex/pkg/observex"
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

func TestStdoutLoggerDebug(t *testing.T) {
	l := stdoutLogger{}
	output := testkit.CaptureStdout(t, func() {
		l.Debug(context.Background(), "dbg", observex.String("k", "v"))
	})
	want := "debug dbg k=v\n"
	if output != want {
		t.Fatalf("unexpected output:\nactual: %q\nwant:   %q", output, want)
	}
}

func TestStdoutLoggerWarn(t *testing.T) {
	l := stdoutLogger{}
	output := testkit.CaptureStdout(t, func() {
		l.Warn(context.Background(), "wrn")
	})
	want := "warn wrn \n"
	if output != want {
		t.Fatalf("unexpected output:\nactual: %q\nwant:   %q", output, want)
	}
}

func TestStdoutLoggerError(t *testing.T) {
	l := stdoutLogger{}
	output := testkit.CaptureStdout(t, func() {
		l.Error(context.Background(), "err", observex.String("x", "1"))
	})
	want := "error err x=1\n"
	if output != want {
		t.Fatalf("unexpected output:\nactual: %q\nwant:   %q", output, want)
	}
}

func TestStdoutLoggerSkipsEmptyKey(t *testing.T) {
	l := stdoutLogger{}
	output := testkit.CaptureStdout(t, func() {
		l.Info(context.Background(), "msg", observex.String("", "val"))
	})
	want := "info msg \n"
	if output != want {
		t.Fatalf("unexpected output:\nactual: %q\nwant:   %q", output, want)
	}
}

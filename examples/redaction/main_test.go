package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestMainPrintsMaskedValue(t *testing.T) {
	output := captureStdout(t, main)
	if output != "***\n" {
		t.Fatalf("unexpected output: %q", output)
	}
	if bytes.Contains([]byte(output), []byte("raw-value-123")) {
		t.Fatalf("output leaked raw value: %q", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = original
	})

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	os.Stdout = original

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return buf.String()
}

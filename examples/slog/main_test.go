package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestMainPrintsSlogJSONWithRedaction(t *testing.T) {
	output := captureStdout(t, main)
	for _, fragment := range []string{
		`"msg":"client ready"`,
		`"trace_id":"trace-example"`,
		`"api_key":"***"`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("output missing %s: %s", fragment, output)
		}
	}
	if strings.Contains(output, "raw-value-123") {
		t.Fatalf("output leaked raw value: %s", output)
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

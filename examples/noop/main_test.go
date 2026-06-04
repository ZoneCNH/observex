package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestMainPrintsNoopStatus(t *testing.T) {
	output := captureStdout(t, main)
	const want = "noop healthy\n"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = write
	t.Cleanup(func() {
		os.Stdout = original
	})

	fn()

	if err := write.Close(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, read); err != nil {
		t.Fatal(err)
	}
	if err := read.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

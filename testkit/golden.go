package testkit

import (
	"os"
	"path/filepath"
	"testing"
)

func requireGolden(goldenPath string, actual []byte, failf failfFunc) {
	expected, err := os.ReadFile(filepath.Clean(goldenPath))
	if err != nil {
		failf("read golden file %s: %v", goldenPath, err)
		return
	}

	if string(expected) != string(actual) {
		failf(
			"golden mismatch for %s\nexpected:\n%s\nactual:\n%s",
			goldenPath,
			expected,
			actual,
		)
		return
	}
}

// RequireGolden fails the test when actual does not match the golden file.
func RequireGolden(t testing.TB, goldenPath string, actual []byte) {
	t.Helper()
	requireGolden(goldenPath, actual, t.Fatalf)
}

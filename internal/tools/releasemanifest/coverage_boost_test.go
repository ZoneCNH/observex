package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFakeGo(t *testing.T, script string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "go")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake go: %v", err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("chmod fake go: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// ── validateDownstreamAdoption boundary tests ───────────────────────

func TestValidateDownstreamAdoptionAllEmpty(t *testing.T) {
	evidence := DownstreamAdoptionEvidence{}
	failures := validateDownstreamAdoption(evidence, false)
	// Should have: fixture_smoke.status required, fixtures required, commands required,
	// real_adoption.status required, blockers required
	if len(failures) < 4 {
		t.Fatalf("expected at least 4 failures for empty evidence, got %d: %v", len(failures), failures)
	}
}

func TestValidateDownstreamAdoptionPassedWithEmptyConsumers(t *testing.T) {
	evidence := DownstreamAdoptionEvidence{
		FixtureSmoke: DownstreamFixtureSmoke{
			Status: "passed",
			Fixtures: []DownstreamFixture{
				{Name: "fx", Module: "mod", Package: "pkg", Evidence: "e"},
			},
			Commands: []DownstreamCommand{
				{Command: "make test", Status: "passed", ExitCode: 0, Evidence: "e"},
			},
		},
		RealAdoption: DownstreamRealAdoption{
			Status: "passed",
			// No consumers -> should fail
		},
	}
	failures := validateDownstreamAdoption(evidence, true)
	found := false
	for _, f := range failures {
		if strings.Contains(f, "consumers is required") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected consumers required failure, got: %v", failures)
	}
}

func TestValidateDownstreamAdoptionWithConsumerCommands(t *testing.T) {
	evidence := DownstreamAdoptionEvidence{
		FixtureSmoke: DownstreamFixtureSmoke{
			Status: "passed",
			Fixtures: []DownstreamFixture{
				{Name: "fx", Module: "mod", Package: "pkg", Evidence: "e"},
			},
			Commands: []DownstreamCommand{
				{Command: "make test", Status: "passed", ExitCode: 0, Evidence: "e"},
			},
		},
		RealAdoption: DownstreamRealAdoption{
			Status: "passed",
			Consumers: []DownstreamConsumer{
				{
					Name:            "consumer1",
					Repository:      "github.com/example/consumer1",
					Commit:          "abc123",
					ObservexVersion: "v0.1.0",
					Evidence:        "evidence.json",
					Commands: []DownstreamCommand{
						{Command: "go test ./...", Status: "passed", ExitCode: 0, Evidence: "test.log"},
					},
				},
			},
		},
	}
	failures := validateDownstreamAdoption(evidence, true)
	if len(failures) != 0 {
		t.Fatalf("expected no failures for valid evidence, got: %v", failures)
	}
}

func TestValidateDownstreamAdoptionConsumerFailures(t *testing.T) {
	evidence := DownstreamAdoptionEvidence{
		FixtureSmoke: DownstreamFixtureSmoke{
			Status: "passed",
			Fixtures: []DownstreamFixture{
				{Name: "fx", Module: "mod", Package: "pkg", Evidence: "e"},
			},
			Commands: []DownstreamCommand{
				{Command: "make test", Status: "passed", ExitCode: 0, Evidence: "e"},
			},
		},
		RealAdoption: DownstreamRealAdoption{
			Status: "passed",
			Consumers: []DownstreamConsumer{
				{
					Name:       "",
					Repository: "",
					Commit:     "",
					Evidence:   "",
					Commands:   []DownstreamCommand{},
				},
			},
		},
	}
	failures := validateDownstreamAdoption(evidence, true)
	if len(failures) < 3 {
		t.Fatalf("expected multiple failures for empty consumer, got %d: %v", len(failures), failures)
	}
}

func TestValidateDownstreamAdoptionConsumerCommandFailed(t *testing.T) {
	evidence := DownstreamAdoptionEvidence{
		FixtureSmoke: DownstreamFixtureSmoke{
			Status: "passed",
			Fixtures: []DownstreamFixture{
				{Name: "fx", Module: "mod", Package: "pkg", Evidence: "e"},
			},
			Commands: []DownstreamCommand{
				{Command: "make test", Status: "passed", ExitCode: 0, Evidence: "e"},
			},
		},
		RealAdoption: DownstreamRealAdoption{
			Status: "passed",
			Consumers: []DownstreamConsumer{
				{
					Name:            "c1",
					Repository:      "r",
					Commit:          "c",
					ObservexVersion: "v1",
					Evidence:        "e",
					Commands: []DownstreamCommand{
						{Command: "test", Status: "failed", ExitCode: 1, Evidence: "e"},
					},
				},
			},
		},
	}
	failures := validateDownstreamAdoption(evidence, true)
	foundStatus := false
	foundExitCode := false
	for _, f := range failures {
		if strings.Contains(f, "status must be passed") {
			foundStatus = true
		}
		if strings.Contains(f, "exit_code must be 0") {
			foundExitCode = true
		}
	}
	if !foundStatus {
		t.Fatalf("expected status failure, got: %v", failures)
	}
	if !foundExitCode {
		t.Fatalf("expected exit_code failure, got: %v", failures)
	}
}

func TestValidateDownstreamAdoptionBlockersRequired(t *testing.T) {
	evidence := DownstreamAdoptionEvidence{
		FixtureSmoke: DownstreamFixtureSmoke{
			Status: "passed",
			Fixtures: []DownstreamFixture{
				{Name: "fx", Module: "mod", Package: "pkg", Evidence: "e"},
			},
			Commands: []DownstreamCommand{
				{Command: "test", Status: "passed", ExitCode: 0, Evidence: "e"},
			},
		},
		RealAdoption: DownstreamRealAdoption{
			Status:   "blocked",
			Blockers: []DownstreamBlocker{},
		},
	}
	failures := validateDownstreamAdoption(evidence, false)
	found := false
	for _, f := range failures {
		if strings.Contains(f, "blockers is required") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected blockers required failure, got: %v", failures)
	}
}

func TestValidateDownstreamAdoptionBlockerEmptyFields(t *testing.T) {
	evidence := DownstreamAdoptionEvidence{
		FixtureSmoke: DownstreamFixtureSmoke{
			Status: "passed",
			Fixtures: []DownstreamFixture{
				{Name: "fx", Module: "mod", Package: "pkg", Evidence: "e"},
			},
			Commands: []DownstreamCommand{
				{Command: "test", Status: "passed", ExitCode: 0, Evidence: "e"},
			},
		},
		RealAdoption: DownstreamRealAdoption{
			Status: "blocked",
			Blockers: []DownstreamBlocker{
				{Scope: "", Reason: "", Evidence: ""},
			},
		},
	}
	failures := validateDownstreamAdoption(evidence, false)
	if len(failures) < 3 {
		t.Fatalf("expected 3 blocker field failures, got %d: %v", len(failures), failures)
	}
}

// ── validateChecks boundary tests ───────────────────────────────────

func TestValidateChecksEmptyStatus(t *testing.T) {
	checks := make(map[string]string)
	for _, name := range checkNames {
		checks[name] = ""
	}
	failures := validateChecks(checks, false)
	if len(failures) != len(checkNames) {
		t.Fatalf("expected %d failures for empty statuses, got %d: %v", len(checkNames), len(failures), failures)
	}
}

func TestValidateChecksNotRequirePassed(t *testing.T) {
	checks := make(map[string]string)
	for _, name := range checkNames {
		checks[name] = "unknown"
	}
	failures := validateChecks(checks, false)
	if len(failures) != 0 {
		t.Fatalf("expected no failures when requirePassed=false, got: %v", failures)
	}
}

// ── requireNonEmpty ─────────────────────────────────────────────────

func TestRequireNonEmpty(t *testing.T) {
	var failures []string
	requireNonEmpty(&failures, "field", "")
	if len(failures) != 1 || !strings.Contains(failures[0], "field is required") {
		t.Fatalf("expected failure, got: %v", failures)
	}
	requireNonEmpty(&failures, "field2", "  ")
	if len(failures) != 2 {
		t.Fatalf("expected 2 failures, got: %d", len(failures))
	}
	requireNonEmpty(&failures, "field3", "value")
	if len(failures) != 2 {
		t.Fatalf("expected still 2 failures, got: %d", len(failures))
	}
}

// ── uniqueStrings ───────────────────────────────────────────────────

func TestUniqueStrings(t *testing.T) {
	got := uniqueStrings([]string{"a", "b", "a", "c", "b"})
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, v := range want {
		if got[i] != v {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], v)
		}
	}
}

func TestUniqueStringsEmpty(t *testing.T) {
	got := uniqueStrings(nil)
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

// ── manifestVersionFromPath ─────────────────────────────────────────

func TestManifestVersionFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"release/manifest/v1.2.3.json", "v1.2.3"},
		{"release/manifest/latest.json", ""},
		{"release/manifest/v1.2.json", ""}, // only 1 dot, needs >= 2
		{"release/manifest/notversion.json", ""},
		{"release/manifest/noext", ""},
		{"/tmp/v0.1.0.json", "v0.1.0"},
	}
	for _, tt := range tests {
		got := manifestVersionFromPath(tt.path)
		if got != tt.want {
			t.Errorf("manifestVersionFromPath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// ── buildDownstreamAdoption ─────────────────────────────────────────

func TestBuildDownstreamAdoptionWithPassedChecks(t *testing.T) {
	checks := map[string]string{"integration": "passed"}
	got := buildDownstreamAdoption(checks)
	if got.FixtureSmoke.Status != "passed" {
		t.Fatalf("expected passed, got %q", got.FixtureSmoke.Status)
	}
	if got.FixtureSmoke.Commands[0].ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", got.FixtureSmoke.Commands[0].ExitCode)
	}
}

func TestBuildDownstreamAdoptionWithUnknownStatus(t *testing.T) {
	checks := map[string]string{"integration": "unknown"}
	got := buildDownstreamAdoption(checks)
	if got.FixtureSmoke.Status != "blocked" {
		t.Fatalf("expected blocked for unknown, got %q", got.FixtureSmoke.Status)
	}
	if got.FixtureSmoke.Commands[0].ExitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", got.FixtureSmoke.Commands[0].ExitCode)
	}
}

func TestBuildDownstreamAdoptionWithFailedStatus(t *testing.T) {
	checks := map[string]string{"integration": "failed"}
	got := buildDownstreamAdoption(checks)
	if got.FixtureSmoke.Status != "failed" {
		t.Fatalf("expected failed, got %q", got.FixtureSmoke.Status)
	}
	if got.FixtureSmoke.Commands[0].ExitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", got.FixtureSmoke.Commands[0].ExitCode)
	}
}

// ── normalizeArtifactPath ───────────────────────────────────────────

func TestNormalizeArtifactPath(t *testing.T) {
	got := normalizeArtifactPath(filepath.Join("release", "manifest", "v1.json"))
	if got != "release/manifest/v1.json" {
		t.Fatalf("expected forward slashes, got %q", got)
	}
}

// ── downstreamEvidencePath ──────────────────────────────────────────

func TestDownstreamEvidencePath(t *testing.T) {
	got := downstreamEvidencePath()
	if !strings.Contains(got, "downstream") {
		t.Fatalf("expected downstream in path, got %q", got)
	}
}

// ── firstLine ───────────────────────────────────────────────────────

func TestFirstLine(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello\nworld", "hello"},
		{"  hello  \n  world  ", "hello  "},
		{"single line", "single line"},
		{"", ""},
	}
	for _, tt := range tests {
		got := firstLine(tt.input)
		if got != tt.want {
			t.Errorf("firstLine(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ── contains ────────────────────────────────────────────────────────

func TestContains(t *testing.T) {
	if !contains([]string{"a", "b", "c"}, "b") {
		t.Fatal("expected true")
	}
	if contains([]string{"a", "b", "c"}, "d") {
		t.Fatal("expected false")
	}
	if contains(nil, "a") {
		t.Fatal("expected false for nil")
	}
}

// ── envDefault ──────────────────────────────────────────────────────

func TestEnvDefault(t *testing.T) {
	t.Setenv("TEST_ENV_DEFAULT_KEY", "value")
	if got := envDefault("TEST_ENV_DEFAULT_KEY", "fallback"); got != "value" {
		t.Fatalf("expected 'value', got %q", got)
	}
	if got := envDefault("TEST_ENV_DEFAULT_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("expected 'fallback', got %q", got)
	}
	t.Setenv("TEST_ENV_DEFAULT_WHITESPACE", "  ")
	if got := envDefault("TEST_ENV_DEFAULT_WHITESPACE", "fallback"); got != "fallback" {
		t.Fatalf("expected 'fallback' for whitespace, got %q", got)
	}
}

// ── releaseVersion ──────────────────────────────────────────────────

func TestReleaseVersionUsesFallback(t *testing.T) {
	t.Setenv("VERSION", "")
	got := releaseVersion()
	if got != fallbackVersion {
		t.Fatalf("expected %q, got %q", fallbackVersion, got)
	}
}

func TestReleaseVersionUsesEnv(t *testing.T) {
	t.Setenv("VERSION", "v9.9.9")
	got := releaseVersion()
	if got != "v9.9.9" {
		t.Fatalf("expected v9.9.9, got %q", got)
	}
}

// ── defaultManifestArtifactPath ─────────────────────────────────────

func TestDefaultManifestArtifactPath(t *testing.T) {
	t.Setenv("VERSION", "v2.0.0")
	got := defaultManifestArtifactPath()
	if !strings.Contains(got, "v2.0.0.json") {
		t.Fatalf("expected version in path, got %q", got)
	}
}

// ── releaseArtifacts ────────────────────────────────────────────────

func TestReleaseArtifactsDeduplication(t *testing.T) {
	// latestManifestArtifactPath appears twice when manifest path == latestManifestArtifactPath
	got := releaseArtifacts(latestManifestArtifactPath)
	seen := make(map[string]bool)
	for _, a := range got {
		if seen[a] {
			t.Fatalf("duplicate artifact: %q", a)
		}
		seen[a] = true
	}
}

// ── writeManifest + writeManifestSHA256 ─────────────────────────────

func TestWriteManifestSHA256Sidecar(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.json")
	manifest := Manifest{
		Module:           "example.com/test",
		Version:          "v1.0.0",
		Checks:           map[string]string{"fmt": "passed"},
		Tools:            map[string]string{"go": "go1.23"},
		Artifacts:        []string{"release/manifest/v1.0.0.json"},
		GeneratedAt:      "2026-01-01T00:00:00Z",
		GeneratedBy:      "test",
		TreeState:        "clean",
		SourceDigest:     "sha256:abc",
		Commit:           "abc123",
		TreeSHA:          "tree123",
		GoVersion:        "go1.23",
		TrackedFileCount: 1,
	}
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}
	sidecar, err := os.ReadFile(path + ".sha256")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(sidecar), "  manifest.json\n") {
		t.Fatalf("expected sidecar to reference manifest.json, got: %s", string(sidecar))
	}
}

func TestWriteManifestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deep", "dir", "manifest.json")
	manifest := Manifest{
		Module:  "example.com/rt",
		Version: "v0.1.0",
		Checks:  map[string]string{"fmt": "passed"},
		Tools:   map[string]string{"go": "go1.23"},
	}
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got Manifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Module != "example.com/rt" {
		t.Fatalf("module = %q", got.Module)
	}
}

// ── buildChecks with env overrides ──────────────────────────────────

func TestBuildChecksEnvOverride(t *testing.T) {
	t.Setenv("CHECK_STATUS", "passed")
	t.Setenv("LINT_STATUS", "failed")
	t.Setenv("SECURITY_STATUS", "passed")
	checks := buildChecks()
	if checks["fmt"] != "passed" {
		t.Fatalf("fmt = %q, want passed", checks["fmt"])
	}
	if checks["lint"] != "failed" {
		t.Fatalf("lint = %q, want failed", checks["lint"])
	}
	if checks["security"] != "passed" {
		t.Fatalf("security = %q, want passed", checks["security"])
	}
}

// ── validateDownstreamAdoption with fixture empty fields ────────────

func TestValidateDownstreamAdoptionFixtureEmptyFields(t *testing.T) {
	evidence := DownstreamAdoptionEvidence{
		FixtureSmoke: DownstreamFixtureSmoke{
			Status: "passed",
			Fixtures: []DownstreamFixture{
				{Name: "", Module: "", Package: "", Evidence: ""},
			},
			Commands: []DownstreamCommand{
				{Command: "", Status: "", Evidence: ""},
			},
		},
		RealAdoption: DownstreamRealAdoption{
			Status: "blocked",
			Blockers: []DownstreamBlocker{
				{Scope: "s", Reason: "r", Evidence: "e"},
			},
		},
	}
	failures := validateDownstreamAdoption(evidence, false)
	// Should have empty field failures for fixtures and commands
	if len(failures) < 4 {
		t.Fatalf("expected at least 4 failures, got %d: %v", len(failures), failures)
	}
}

// ── runCLI edge cases ───────────────────────────────────────────────

func TestRunCLIVerifyNonExistentFile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runCLI("test", []string{"-verify", "/nonexistent/file.json"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d; stderr: %s", code, stderr.String())
	}
}

func TestRunCLIVerifyInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := runCLI("test", []string{"-verify", path}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d; stderr: %s", code, stderr.String())
	}
}

func TestPrintCLIMessage(t *testing.T) {
	var buf bytes.Buffer
	code := printCLIMessage(&buf, 0, "hello %s", "world")
	if code != 0 {
		t.Fatalf("expected 0, got %d", code)
	}
	if buf.String() != "hello world" {
		t.Fatalf("expected 'hello world', got %q", buf.String())
	}
}

// ── treeState error path ────────────────────────────────────────────

func TestTreeStateOutsideGitRepo(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	// Not a git repo -> git status fails -> "unknown"
	got := treeState()
	if got != "unknown" {
		t.Fatalf("expected 'unknown' outside git repo, got %q", got)
	}
}

func TestTreeStateDirty(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	runTestCommand(t, dir, "git", "init")
	runTestCommand(t, dir, "git", "config", "user.email", "test@test.com")
	runTestCommand(t, dir, "git", "config", "user.name", "Test")
	// Create an untracked file to make the tree dirty
	if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := treeState()
	if got != "dirty" {
		t.Fatalf("expected 'dirty' with untracked file, got %q", got)
	}
}

// ── toolVersion error path ──────────────────────────────────────────

func TestToolVersionError(t *testing.T) {
	// Use a binary that exists but returns error on the given flag
	got := toolVersion("go", "--definitely-not-a-real-flag")
	// Should return "error: ..." since go will fail
	if !strings.HasPrefix(got, "error:") && !strings.HasPrefix(got, "missing") {
		t.Fatalf("expected error or missing prefix, got %q", got)
	}
}

func TestToolVersionSuccess(t *testing.T) {
	got := toolVersion("go", "version")
	if !strings.Contains(got, "go") {
		t.Fatalf("expected go version string, got %q", got)
	}
}

// ── verifyManifest with various field mismatches ────────────────────

func TestVerifyManifestFieldMismatches(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, repoRoot(t))

	path := filepath.Join(t.TempDir(), "manifest.json")
	manifest, err := buildManifestFor(path)
	if err != nil {
		t.Fatal(err)
	}

	// Corrupt various fields to trigger mismatch failures
	manifest.Module = "wrong.module"
	manifest.Commit = "wrong-commit"
	manifest.TreeSHA = "wrong-tree"
	manifest.TrackedFileCount = 99999
	manifest.GoVersion = ""
	manifest.GeneratedBy = ""
	manifest.Notes.BreakingChanges = ""
	manifest.Notes.DownstreamEvidence = ""
	manifest.GeneratedAt = "not-rfc3339"
	manifest.Tools = map[string]string{}
	manifest.Contracts = nil
	manifest.Dependencies = nil
	manifest.Artifacts = nil

	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "")
	if err == nil {
		t.Fatal("expected verification error")
	}
	message := err.Error()
	for _, want := range []string{
		"go_version is required",
		"generated_by is required",
		"notes.breaking_changes is required",
		"notes.downstream_evidence is required",
		"generated_at must be RFC3339",
		"module mismatch",
		"commit mismatch",
		"tree_sha mismatch",
		"tracked_file_count mismatch",
		"tools.go must be recorded",
		"contract fingerprints",
		"dependency inventory",
		"artifacts must include",
	} {
		if !strings.Contains(message, want) {
			t.Errorf("expected %q in error, got: %s", want, message)
		}
	}
}

// ── verifyManifest with version mismatch from expect-version ────────

func TestVerifyManifestExpectVersionMismatch(t *testing.T) {
	t.Setenv("GOWORK", "off")
	t.Setenv("CHECK_STATUS", "passed")
	chdir(t, repoRoot(t))

	path := filepath.Join(t.TempDir(), "manifest.json")
	manifest, err := buildManifestFor(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeManifest(path, manifest); err != nil {
		t.Fatal(err)
	}

	err = verifyManifest(path, false, false, "v99.99.99")
	if err == nil {
		t.Fatal("expected version mismatch error")
	}
	if !strings.Contains(err.Error(), "version mismatch") {
		t.Fatalf("expected version mismatch, got: %v", err)
	}
}

// ── sourceDigest with deleted files ─────────────────────────────────

func TestSourceDigestHandlesDeletedFiles(t *testing.T) {
	repo := t.TempDir()
	runTestCommand(t, repo, "git", "init")
	// Create and track a file, then delete it
	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestCommand(t, repo, "git", "add", ".")
	if err := os.Remove(filepath.Join(repo, "file.txt")); err != nil {
		t.Fatal(err)
	}
	chdir(t, repo)

	digest, count, err := sourceDigest()
	if err != nil {
		t.Fatal(err)
	}
	// File was deleted, so os.ErrNotExist is handled -> count may be 0
	if !strings.HasPrefix(digest, "sha256:") {
		t.Fatalf("expected sha256 prefix, got %q", digest)
	}
	_ = count
}

func TestSourceDigestReturnsErrorOutsideGitRepo(t *testing.T) {
	chdir(t, t.TempDir())

	_, _, err := sourceDigest()
	if err == nil {
		t.Fatal("expected error outside git repo")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("expected git repo error, got: %v", err)
	}
}

func TestSourceDigestReturnsReadErrorForBrokenTrackedPath(t *testing.T) {
	repo := t.TempDir()
	runTestCommand(t, repo, "git", "init")
	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestCommand(t, repo, "git", "add", ".")
	if err := os.Remove(filepath.Join(repo, "file.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(repo, "file.txt"), 0o755); err != nil {
		t.Fatal(err)
	}
	chdir(t, repo)

	_, _, err := sourceDigest()
	if err == nil {
		t.Fatal("expected read error for directory path")
	}
	if !strings.Contains(err.Error(), "is a directory") {
		t.Fatalf("expected directory read error, got: %v", err)
	}
}

func TestModuleDigestsReportsDecodeError(t *testing.T) {
	writeFakeGo(t, "#!/bin/sh\nprintf 'not-json'\n")

	_, err := moduleDigests()
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("expected decode error, got: %v", err)
	}
}

func TestBuildManifestForReturnsErrorWhenGoListFails(t *testing.T) {
	writeFakeGo(t, "#!/bin/sh\nexit 1\n")

	_, err := buildManifestFor(filepath.Join(t.TempDir(), "manifest.json"))
	if err == nil {
		t.Fatal("expected buildManifestFor error")
	}
	if !strings.Contains(err.Error(), "go list -m failed") {
		t.Fatalf("expected go list failure, got: %v", err)
	}
}

// ── writeManifest error paths ───────────────────────────────────────

func TestWriteManifestMkdirAllError(t *testing.T) {
	// Use a path where MkdirAll would fail (file blocks directory creation)
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("block"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(blocked, "sub", "manifest.json")
	err := writeManifest(path, Manifest{})
	if err == nil {
		t.Fatal("expected mkdir error")
	}
}

// ── runCLI with generate error ──────────────────────────────────────

func TestRunCLIGenerateErrorOutsideRepo(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	var stdout, stderr bytes.Buffer
	code := runCLI("test", []string{}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 outside repo, got %d; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "ERROR:") {
		t.Fatalf("expected ERROR in stderr, got: %s", stderr.String())
	}
}

// ── normalizeArtifactPath ───────────────────────────────────────────

func TestNormalizeArtifactPathBackslash(t *testing.T) {
	// On Linux this is a no-op, but we test the function exists
	got := normalizeArtifactPath("a/b/c")
	if got != "a/b/c" {
		t.Fatalf("expected 'a/b/c', got %q", got)
	}
}

// ── defaultManifestArtifactPath ─────────────────────────────────────

func TestDefaultManifestArtifactPathContainsVersion(t *testing.T) {
	t.Setenv("VERSION", "v3.0.0")
	got := defaultManifestArtifactPath()
	if !strings.Contains(got, "v3.0.0") {
		t.Fatalf("expected version in path, got %q", got)
	}
	if !strings.HasSuffix(got, ".json") {
		t.Fatalf("expected .json suffix, got %q", got)
	}
}

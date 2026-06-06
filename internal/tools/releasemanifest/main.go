package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	fallbackVersion            = "v0.3.2"
	latestManifestArtifactPath = "release/manifest/latest.json"
)

var checkNames = []string{
	"fmt",
	"vet",
	"lint",
	"unit_test",
	"race_test",
	"boundary",
	"secret_scan",
	"security",
	"contract",
	"integration",
}

var checkEnvNames = map[string]string{
	"fmt":         "FMT_STATUS",
	"vet":         "VET_STATUS",
	"lint":        "LINT_STATUS",
	"unit_test":   "UNIT_TEST_STATUS",
	"race_test":   "RACE_TEST_STATUS",
	"boundary":    "BOUNDARY_STATUS",
	"secret_scan": "SECRET_SCAN_STATUS",
	"security":    "SECURITY_STATUS",
	"contract":    "CONTRACT_STATUS",
	"integration": "INTEGRATION_STATUS",
}

var contractFiles = []string{
	"contracts/config.schema.json",
	"contracts/error.schema.json",
	"contracts/field.schema.json",
	"contracts/health.schema.json",
	"contracts/logger.schema.json",
	"contracts/metric_naming.md",
	"contracts/metrics.schema.json",
	"contracts/metrics.md",
	"contracts/public_api.md",
	"contracts/public_api.snapshot",
	"contracts/tracer.schema.json",
}

type Manifest struct {
	Module             string                     `json:"module"`
	Version            string                     `json:"version"`
	Commit             string                     `json:"commit"`
	TreeSHA            string                     `json:"tree_sha"`
	SourceDigest       string                     `json:"source_digest"`
	TrackedFileCount   int                        `json:"tracked_file_count"`
	GoVersion          string                     `json:"go_version"`
	GeneratedAt        string                     `json:"generated_at"`
	GeneratedBy        string                     `json:"generated_by"`
	TreeState          string                     `json:"tree_state"`
	Checks             map[string]string          `json:"checks"`
	Contracts          []FileDigest               `json:"contracts"`
	Dependencies       []ModuleDigest             `json:"dependencies"`
	Tools              map[string]string          `json:"tools"`
	Artifacts          []string                   `json:"artifacts"`
	DownstreamAdoption DownstreamAdoptionEvidence `json:"downstream_adoption"`
	Notes              Notes                      `json:"notes"`
}

type FileDigest struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type ModuleDigest struct {
	Path    string         `json:"path"`
	Version string         `json:"version,omitempty"`
	Main    bool           `json:"main,omitempty"`
	Replace *ModuleReplace `json:"replace,omitempty"`
}

type ModuleReplace struct {
	Path    string `json:"path"`
	Version string `json:"version,omitempty"`
}

type Notes struct {
	BreakingChanges    string   `json:"breaking_changes"`
	DownstreamEvidence string   `json:"downstream_evidence"`
	KnownRisks         []string `json:"known_risks"`
}

type DownstreamAdoptionEvidence struct {
	FixtureSmoke DownstreamFixtureSmoke `json:"fixture_smoke"`
	RealAdoption DownstreamRealAdoption `json:"real_adoption"`
}

type DownstreamFixtureSmoke struct {
	Status   string              `json:"status"`
	Fixtures []DownstreamFixture `json:"fixtures"`
	Commands []DownstreamCommand `json:"commands"`
}

type DownstreamRealAdoption struct {
	Status    string               `json:"status"`
	Consumers []DownstreamConsumer `json:"consumers"`
	Blockers  []DownstreamBlocker  `json:"blockers"`
}

type DownstreamConsumer struct {
	Name            string              `json:"name"`
	Repository      string              `json:"repository"`
	Commit          string              `json:"commit"`
	ObservexVersion string              `json:"observex_version"`
	Commands        []DownstreamCommand `json:"commands"`
	Evidence        string              `json:"evidence"`
}

type DownstreamFixture struct {
	Name     string `json:"name"`
	Module   string `json:"module"`
	Package  string `json:"package"`
	Evidence string `json:"evidence"`
}

type DownstreamCommand struct {
	Command  string `json:"command"`
	Status   string `json:"status"`
	ExitCode int    `json:"exit_code"`
	Evidence string `json:"evidence"`
}

type DownstreamBlocker struct {
	Scope    string `json:"scope"`
	Reason   string `json:"reason"`
	Evidence string `json:"evidence"`
}

func main() {
	os.Exit(runCLI(os.Args[0], os.Args[1:], os.Stdout, os.Stderr))
}

func runCLI(name string, args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	out := flags.String("out", defaultManifestArtifactPath(), "release manifest output path")
	verify := flags.String("verify", "", "verify an existing release manifest instead of generating one")
	requirePassed := flags.Bool("require-passed", false, "require all release checks to be passed during verification")
	requireClean := flags.Bool("require-clean", false, "require a clean git tree during verification")
	expectVersion := flags.String("expect-version", "", "require the manifest version to match this release version during verification")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if *verify != "" {
		if err := verifyManifest(*verify, *requirePassed, *requireClean, *expectVersion); err != nil {
			return printCLIError(stderr, err)
		}
		return printCLIStatus(stdout, "release evidence verified: %s\n", *verify)
	}

	manifest, err := buildManifestFor(*out)
	if err != nil {
		return printCLIError(stderr, err)
	}
	if err := writeManifest(*out, manifest); err != nil {
		return printCLIError(stderr, err)
	}
	return printCLIStatus(stdout, "generated %s\n", *out)
}

func printCLIError(w io.Writer, err error) int {
	return printCLIMessage(w, 1, "ERROR: %v\n", err)
}

func printCLIStatus(w io.Writer, format string, args ...any) int {
	return printCLIMessage(w, 0, format, args...)
}

func printCLIMessage(w io.Writer, exitCode int, format string, args ...any) int {
	_, err := fmt.Fprintf(w, format, args...)
	if err != nil {
		return 1
	}
	return exitCode
}

func buildManifest() (Manifest, error) {
	return buildManifestFor(defaultManifestArtifactPath())
}

func buildManifestFor(manifestPath string) (Manifest, error) {
	module, err := runTrimmed("go", "list", "-m")
	if err != nil {
		return Manifest{}, err
	}

	sourceDigest, trackedFileCount, err := sourceDigest()
	if err != nil {
		return Manifest{}, err
	}
	contracts, err := contractDigests()
	if err != nil {
		return Manifest{}, err
	}
	dependencies, err := moduleDigests()
	if err != nil {
		return Manifest{}, err
	}
	checks := buildChecks()

	return Manifest{
		Module:           module,
		Version:          releaseVersion(),
		Commit:           runTrimmedDefault("unknown", "git", "rev-parse", "HEAD"),
		TreeSHA:          runTrimmedDefault("unknown", "git", "rev-parse", "HEAD^{tree}"),
		SourceDigest:     sourceDigest,
		TrackedFileCount: trackedFileCount,
		GoVersion:        runtime.Version(),
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
		GeneratedBy:      envDefault("GENERATED_BY", "scripts/generate_manifest.sh"),
		TreeState:        treeState(),
		Checks:           checks,
		Contracts:        contracts,
		Dependencies:     dependencies,
		Tools: map[string]string{
			"go":            firstLine(runTrimmedDefault(runtime.Version(), "go", "version")),
			"golangci-lint": toolVersion("golangci-lint", "--version"),
			"govulncheck":   toolVersion("govulncheck", "-version"),
		},
		Artifacts:          releaseArtifacts(manifestPath),
		DownstreamAdoption: buildDownstreamAdoption(checks),
		Notes: Notes{
			BreakingChanges:    "none",
			DownstreamEvidence: downstreamEvidence(),
			KnownRisks:         []string{},
		},
	}, nil
}

func verifyManifest(path string, requirePassed bool, requireClean bool, expectVersion string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var got Manifest
	if err := json.Unmarshal(data, &got); err != nil {
		return err
	}

	current, err := buildManifestFor(path)
	if err != nil {
		return err
	}

	var failures []string
	requireNonEmpty(&failures, "module", got.Module)
	requireNonEmpty(&failures, "version", got.Version)
	requireNonEmpty(&failures, "commit", got.Commit)
	requireNonEmpty(&failures, "tree_sha", got.TreeSHA)
	requireNonEmpty(&failures, "source_digest", got.SourceDigest)
	requireNonEmpty(&failures, "go_version", got.GoVersion)
	requireNonEmpty(&failures, "generated_at", got.GeneratedAt)
	requireNonEmpty(&failures, "generated_by", got.GeneratedBy)
	requireNonEmpty(&failures, "tree_state", got.TreeState)
	requireNonEmpty(&failures, "notes.breaking_changes", got.Notes.BreakingChanges)
	requireNonEmpty(&failures, "notes.downstream_evidence", got.Notes.DownstreamEvidence)

	expectVersion = strings.TrimSpace(expectVersion)
	if _, err := time.Parse(time.RFC3339, got.GeneratedAt); err != nil {
		failures = append(failures, "generated_at must be RFC3339")
	}
	if expectVersion != "" && got.Version != expectVersion {
		failures = append(failures, fmt.Sprintf("version mismatch: got %q, want %q", got.Version, expectVersion))
	}
	if pathVersion := manifestVersionFromPath(path); pathVersion != "" && got.Version != pathVersion {
		failures = append(failures, fmt.Sprintf("manifest path version mismatch: path %q implies %q, got %q", normalizeArtifactPath(path), pathVersion, got.Version))
	}
	if got.Module != current.Module {
		failures = append(failures, fmt.Sprintf("module mismatch: got %q, want %q", got.Module, current.Module))
	}
	if got.Commit != current.Commit {
		failures = append(failures, fmt.Sprintf("commit mismatch: got %q, want %q", got.Commit, current.Commit))
	}
	if got.TreeSHA != current.TreeSHA {
		failures = append(failures, fmt.Sprintf("tree_sha mismatch: got %q, want %q", got.TreeSHA, current.TreeSHA))
	}
	if got.SourceDigest != current.SourceDigest {
		failures = append(failures, "source_digest does not match current tracked file contents")
	}
	if got.TrackedFileCount != current.TrackedFileCount {
		failures = append(failures, fmt.Sprintf("tracked_file_count mismatch: got %d, want %d", got.TrackedFileCount, current.TrackedFileCount))
	}
	if got.TreeState != current.TreeState {
		failures = append(failures, fmt.Sprintf("tree_state mismatch: got %q, want %q", got.TreeState, current.TreeState))
	}
	if requireClean && got.TreeState != "clean" {
		failures = append(failures, fmt.Sprintf("tree_state must be clean, got %q", got.TreeState))
	}
	if !reflect.DeepEqual(got.Contracts, current.Contracts) {
		failures = append(failures, "contract fingerprints do not match current contract files")
	}
	if !reflect.DeepEqual(got.Dependencies, current.Dependencies) {
		failures = append(failures, "dependency inventory does not match go list -m -json all")
	}
	for _, artifact := range releaseArtifacts(path) {
		if !contains(got.Artifacts, artifact) {
			failures = append(failures, "artifacts must include "+artifact)
		}
	}
	if got.Tools["go"] == "" {
		failures = append(failures, "tools.go must be recorded")
	}
	failures = append(failures, validateChecks(got.Checks, requirePassed)...)
	failures = append(failures, validateDownstreamAdoption(got.DownstreamAdoption, requirePassed)...)

	if len(failures) > 0 {
		return errors.New("release evidence verification failed:\n - " + strings.Join(failures, "\n - "))
	}
	return nil
}

func writeManifest(path string, manifest Manifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		return err
	}
	data := buf.Bytes()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	return writeManifestSHA256(path, data)
}

func writeManifestSHA256(path string, data []byte) error {
	sum := sha256.Sum256(data)
	content := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), filepath.Base(path))
	return os.WriteFile(path+".sha256", []byte(content), 0o644)
}

func buildChecks() map[string]string {
	defaultStatus := envDefault("CHECK_STATUS", "unknown")
	checks := make(map[string]string, len(checkNames))
	for _, name := range checkNames {
		checks[name] = envDefault(checkEnvNames[name], defaultStatus)
	}
	return checks
}

func validateChecks(checks map[string]string, requirePassed bool) []string {
	var failures []string
	for _, name := range checkNames {
		status := strings.TrimSpace(checks[name])
		if status == "" {
			failures = append(failures, "checks."+name+" is required")
			continue
		}
		if requirePassed && status != "passed" {
			failures = append(failures, fmt.Sprintf("checks.%s must be passed, got %q", name, status))
		}
	}
	return failures
}

func buildDownstreamAdoption(checks map[string]string) DownstreamAdoptionEvidence {
	status := envDefault("DOWNSTREAM_ADOPTION_STATUS", checks["integration"])
	if status == "unknown" {
		status = "blocked"
	}
	exitCode := 1
	if status == "passed" {
		exitCode = 0
	}

	return DownstreamAdoptionEvidence{
		FixtureSmoke: DownstreamFixtureSmoke{
			Status: status,
			Fixtures: []DownstreamFixture{
				{Name: "configx", Module: "github.com/ZoneCNH/configx", Package: "configx", Evidence: "scripts/run_integration.sh"},
				{Name: "corekit", Module: "example.com/acme/corekit", Package: "corekit", Evidence: "scripts/run_integration.sh"},
			},
			Commands: []DownstreamCommand{
				{Command: "GOWORK=off make integration", Status: status, ExitCode: exitCode, Evidence: "scripts/run_integration.sh"},
			},
		},
		RealAdoption: DownstreamRealAdoption{
			Status: "blocked",
			Blockers: []DownstreamBlocker{
				{Scope: "external_real_downstream", Reason: "No maintained external downstream repository is committed in this source tree; release evidence is fixture-backed until replaced by real repository commit and CI evidence.", Evidence: downstreamEvidencePath()},
			},
		},
	}
}

func validateDownstreamAdoption(evidence DownstreamAdoptionEvidence, requirePassed bool) []string {
	var failures []string
	status := strings.TrimSpace(evidence.FixtureSmoke.Status)
	if status == "" {
		failures = append(failures, "downstream_adoption.fixture_smoke.status is required")
	}
	if requirePassed && status != "passed" {
		failures = append(failures, fmt.Sprintf("downstream_adoption.fixture_smoke.status must be passed, got %q", status))
	}
	if len(evidence.FixtureSmoke.Fixtures) == 0 {
		failures = append(failures, "downstream_adoption.fixture_smoke.fixtures is required")
	}
	for i, fixture := range evidence.FixtureSmoke.Fixtures {
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.fixture_smoke.fixtures[%d].name", i), fixture.Name)
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.fixture_smoke.fixtures[%d].module", i), fixture.Module)
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.fixture_smoke.fixtures[%d].package", i), fixture.Package)
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.fixture_smoke.fixtures[%d].evidence", i), fixture.Evidence)
	}
	if len(evidence.FixtureSmoke.Commands) == 0 {
		failures = append(failures, "downstream_adoption.fixture_smoke.commands is required")
	}
	for i, command := range evidence.FixtureSmoke.Commands {
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.fixture_smoke.commands[%d].command", i), command.Command)
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.fixture_smoke.commands[%d].status", i), command.Status)
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.fixture_smoke.commands[%d].evidence", i), command.Evidence)
		if requirePassed && command.Status != "passed" {
			failures = append(failures, fmt.Sprintf("downstream_adoption.fixture_smoke.commands[%d].status must be passed, got %q", i, command.Status))
		}
		if requirePassed && command.ExitCode != 0 {
			failures = append(failures, fmt.Sprintf("downstream_adoption.fixture_smoke.commands[%d].exit_code must be 0, got %d", i, command.ExitCode))
		}
	}

	realStatus := strings.TrimSpace(evidence.RealAdoption.Status)
	if realStatus == "" {
		failures = append(failures, "downstream_adoption.real_adoption.status is required")
	}
	for i, consumer := range evidence.RealAdoption.Consumers {
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.real_adoption.consumers[%d].name", i), consumer.Name)
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.real_adoption.consumers[%d].repository", i), consumer.Repository)
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.real_adoption.consumers[%d].commit", i), consumer.Commit)
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.real_adoption.consumers[%d].observex_version", i), consumer.ObservexVersion)
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.real_adoption.consumers[%d].evidence", i), consumer.Evidence)
		if len(consumer.Commands) == 0 {
			failures = append(failures, fmt.Sprintf("downstream_adoption.real_adoption.consumers[%d].commands is required", i))
		}
		for j, command := range consumer.Commands {
			requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.real_adoption.consumers[%d].commands[%d].command", i, j), command.Command)
			requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.real_adoption.consumers[%d].commands[%d].status", i, j), command.Status)
			requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.real_adoption.consumers[%d].commands[%d].evidence", i, j), command.Evidence)
			if realStatus == "passed" && command.Status != "passed" {
				failures = append(failures, fmt.Sprintf("downstream_adoption.real_adoption.consumers[%d].commands[%d].status must be passed, got %q", i, j, command.Status))
			}
			if realStatus == "passed" && command.ExitCode != 0 {
				failures = append(failures, fmt.Sprintf("downstream_adoption.real_adoption.consumers[%d].commands[%d].exit_code must be 0, got %d", i, j, command.ExitCode))
			}
		}
	}
	if realStatus != "passed" && len(evidence.RealAdoption.Blockers) == 0 {
		failures = append(failures, "downstream_adoption.real_adoption.blockers is required when real adoption is not passed")
	}
	if realStatus == "passed" && len(evidence.RealAdoption.Consumers) == 0 {
		failures = append(failures, "downstream_adoption.real_adoption.consumers is required when real adoption is passed")
	}
	for i, blocker := range evidence.RealAdoption.Blockers {
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.real_adoption.blockers[%d].scope", i), blocker.Scope)
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.real_adoption.blockers[%d].reason", i), blocker.Reason)
		requireNonEmpty(&failures, fmt.Sprintf("downstream_adoption.real_adoption.blockers[%d].evidence", i), blocker.Evidence)
	}
	return failures
}

func sourceDigest() (string, int, error) {
	raw, err := runRaw("git", "ls-files", "-z")
	if err != nil {
		return "", 0, err
	}
	parts := strings.Split(string(raw), "\x00")
	files := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			files = append(files, part)
		}
	}
	sort.Strings(files)

	digest := sha256.New()
	count := 0
	for _, path := range files {
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", 0, err
		}
		fileSum := sha256.Sum256(data)
		digest.Write([]byte(path))
		digest.Write([]byte{0})
		digest.Write([]byte(hex.EncodeToString(fileSum[:])))
		digest.Write([]byte{0})
		count++
	}

	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), count, nil
}

func contractDigests() ([]FileDigest, error) {
	digests := make([]FileDigest, 0, len(contractFiles))
	for _, path := range contractFiles {
		digest, err := fileDigest(path)
		if err != nil {
			return nil, err
		}
		digests = append(digests, digest)
	}
	return digests, nil
}

func fileDigest(path string) (FileDigest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FileDigest{}, err
	}
	sum := sha256.Sum256(data)
	return FileDigest{
		Path:   path,
		SHA256: "sha256:" + hex.EncodeToString(sum[:]),
	}, nil
}

func moduleDigests() ([]ModuleDigest, error) {
	raw, err := runRaw("go", "list", "-m", "-json", "all")
	if err != nil {
		return nil, err
	}

	type goModule struct {
		Path    string
		Version string
		Main    bool
		Replace *struct {
			Path    string
			Version string
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	var modules []ModuleDigest
	for {
		var module goModule
		if err := decoder.Decode(&module); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		digest := ModuleDigest{
			Path:    module.Path,
			Version: module.Version,
			Main:    module.Main,
		}
		if module.Replace != nil {
			digest.Replace = &ModuleReplace{
				Path:    module.Replace.Path,
				Version: module.Replace.Version,
			}
		}
		modules = append(modules, digest)
	}
	return modules, nil
}

func treeState() string {
	status, err := runTrimmed("git", "status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return "unknown"
	}
	if status == "" {
		return "clean"
	}
	return "dirty"
}

func toolVersion(name string, args ...string) string {
	if _, err := exec.LookPath(name); err != nil {
		return "missing"
	}
	output, err := runTrimmed(name, args...)
	if err != nil {
		return "error: " + firstLine(err.Error())
	}
	return firstLine(output)
}

func runTrimmedDefault(fallback string, name string, args ...string) string {
	output, err := runTrimmed(name, args...)
	if err != nil {
		return fallback
	}
	return output
}

func runTrimmed(name string, args ...string) (string, error) {
	output, err := runRaw(name, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func runRaw(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func envDefault(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func releaseVersion() string {
	return envDefault("VERSION", fallbackVersion)
}

func defaultManifestArtifactPath() string {
	return filepath.ToSlash(filepath.Join("release", "manifest", releaseVersion()+".json"))
}

func normalizeArtifactPath(path string) string {
	return filepath.ToSlash(path)
}

func releaseArtifacts(manifestPath string) []string {
	manifestArtifactPath := normalizeArtifactPath(manifestPath)
	return uniqueStrings([]string{
		manifestArtifactPath,
		manifestArtifactPath + ".sha256",
		latestManifestArtifactPath,
		latestManifestArtifactPath + ".sha256",
		downstreamEvidencePath(),
	})
}

func uniqueStrings(values []string) []string {
	unique := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

func manifestVersionFromPath(path string) string {
	base := filepath.Base(filepath.Clean(path))
	if !strings.HasSuffix(base, ".json") {
		return ""
	}
	version := strings.TrimSuffix(base, ".json")
	if version == "latest" || !strings.HasPrefix(version, "v") || strings.Count(version, ".") < 2 {
		return ""
	}
	return version
}

func downstreamEvidencePath() string {
	return filepath.ToSlash(filepath.Join("release", "downstream", "adoption.json"))
}

func downstreamEvidence() string {
	return envDefault("DOWNSTREAM_EVIDENCE", downstreamEvidencePath())
}

func firstLine(value string) string {
	value = strings.TrimSpace(value)
	if idx := strings.IndexByte(value, '\n'); idx >= 0 {
		return value[:idx]
	}
	return value
}

func requireNonEmpty(failures *[]string, field string, value string) {
	if strings.TrimSpace(value) == "" {
		*failures = append(*failures, field+" is required")
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// End-to-end tests that run the real `specs` binary as a process.
//
// Every other test in this module calls [cmd.Execute] in-process, which
// exercises the command tree but stops one layer short of what a user and
// a CI job actually observe: the built binary, its argument vector, and
// the status the process exits with. That last layer is main.go, and it
// is the only place os.Exit is wired — so nothing verified that the
// status cmd.Execute decides on is the status the process reports.
//
// The distinction is not academic. AC-SPEC-006.6 ("at least one error
// exits non-zero") is a claim about a process, and a CI job asserting it
// runs the binary, not a Go function. A main.go that dropped the return
// value, exited zero unconditionally, or wrote to the wrong stream would
// leave every in-process test green.
//
// These tests reuse the snapshot cases' committed corpora rather than
// introducing a second set of fixtures, so the end-to-end path and the
// in-process path are held to the same inputs and the same expected
// output. A divergence between them means the process boundary changed
// something, which is exactly what these tests exist to catch.
package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// coverDirEnv is the variable scripts/coverage.sh sets to an absolute
// directory. When it is present these tests build an instrumented binary
// and point each run's GOCOVERDIR at that directory, so what the process
// executes lands in the same coverage profile as the in-process tests.
// Unset — the ordinary `go test` path — nothing here changes.
const coverDirEnv = "SPEC_E2E_COVERDIR"

// coverPkg is the pattern the instrumented binary is built to measure.
// It has to be the module path rather than ./..., because `go build`
// resolves a relative pattern against the package being built and would
// instrument this directory alone, leaving cmd/ and core/ — everything
// the binary actually spends its time in — uninstrumented and silently
// reported as zero.
const coverPkg = "github.com/brettcallaghan/agent-plugins/cli/..."

// coverDir returns the directory end-to-end runs write coverage data
// into, or "" when the run is not being measured.
func coverDir(t *testing.T) string {
	t.Helper()

	dir := os.Getenv(coverDirEnv)
	if dir == "" {
		return ""
	}
	// Every run below sets cmd.Dir to a materialized corpus, and the Go
	// runtime resolves GOCOVERDIR against the process's working
	// directory — a relative path would scatter one directory per corpus
	// into temporary directories that are deleted when the test ends.
	if !filepath.IsAbs(dir) {
		t.Fatalf("%s must be an absolute path, got %q", coverDirEnv, dir)
	}
	// An instrumented binary whose GOCOVERDIR does not exist fails at
	// exit, which would surface as an unexplained non-zero status rather
	// than as a coverage problem.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating the coverage directory %s: %v", dir, err)
	}
	return dir
}

// buildSpec builds the specs binary once per test binary and returns its
// path. It uses the module the test itself lives in, so it always builds
// the working tree's code rather than anything previously installed.
func buildSpec(t *testing.T) string {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping end-to-end binary test in -short mode (it compiles the binary)")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("no go toolchain on PATH: %v", err)
	}

	// Windows will not execute a file without an executable extension,
	// however absolute the path handed to exec.Command — the failure is a
	// bare "executable file not found in %PATH%", which reads like a PATH
	// problem and is not one. go build honours whatever -o name it is
	// given, so the suffix has to be added here rather than assumed.
	name := "specs"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(t.TempDir(), name)
	args := []string{"build"}
	if coverDir(t) != "" {
		args = append(args, "-cover", "-coverpkg="+coverPkg)
	}
	args = append(args, "-o", bin, ".")
	build := exec.Command("go", args...)
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building the specs binary: %v\n%s", err, out)
	}
	return bin
}

// specCommand prepares a run of the built binary in dir, with the
// coverage environment applied when the run is being measured.
//
// Every process started by these tests goes through here: a run started
// with exec.Command directly inherits no GOCOVERDIR, and an instrumented
// binary without one writes nothing while still passing, so the coverage
// it contributes would silently drop to zero.
func specCommand(t *testing.T, bin, dir string, args ...string) *exec.Cmd {
	t.Helper()

	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	if cover := coverDir(t); cover != "" {
		cmd.Env = append(os.Environ(), "GOCOVERDIR="+cover)
	}
	return cmd
}

// corpusFrom materializes a snapshot case's committed input/ corpus into
// a fresh directory, so a run never mutates the case.
func corpusFrom(t *testing.T, caseName string) string {
	t.Helper()

	src := filepath.Join("..", "testdata", "cases", caseName, "input")
	dst := t.TempDir()

	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("reading the case corpus %s: %v", src, err)
	}
	if len(entries) == 0 {
		t.Fatalf("case %s has an empty input/ corpus", caseName)
	}

	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("materializing the case corpus: %v", err)
	}
	return dst
}

// run executes the binary in dir and returns its stdout and exit status.
// A failure to start the process at all is fatal — that is a broken test,
// not a non-zero exit.
func run(t *testing.T, bin, dir string, args ...string) (string, int) {
	t.Helper()

	cmd := specCommand(t, bin, dir, args...)
	out, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if !asExitError(err, &exit) {
			t.Fatalf("running %s %v: %v", bin, args, err)
		}
		return string(out), exit.ExitCode()
	}
	return string(out), 0
}

// asExitError reports whether err is an *exec.ExitError, assigning it to
// target when it is. errors.As with a concrete pointer type, spelled out
// so the intent at the call site stays obvious.
func asExitError(err error, target **exec.ExitError) bool {
	if e, ok := err.(*exec.ExitError); ok {
		*target = e
		return true
	}
	return false
}

// TestBinaryExitsNonZeroOnAnErrorFinding runs the real process against a
// corpus with one error finding and asserts the process exit status,
// which is the observable AC-SPEC-006.6 actually constrains.
//
// AC: AC-SPEC-006.6
func TestBinaryExitsNonZeroOnAnErrorFinding(t *testing.T) {
	bin := buildSpec(t)
	dir := corpusFrom(t, "check-error")

	stdout, code := run(t, bin, dir, "check")

	if code != 1 {
		t.Errorf("process exit status = %d, want 1\n--- stdout ---\n%s", code, stdout)
	}

	want, err := os.ReadFile(filepath.Join("..", "testdata", "cases", "check-error", "expected", "checks.log"))
	if err != nil {
		t.Fatalf("reading the case's expected stdout: %v", err)
	}
	if stdout != string(want) {
		t.Errorf("process stdout differs from the in-process snapshot for the same corpus.\n--- got ---\n%s\n--- want ---\n%s", stdout, want)
	}
}

// TestBinaryExitsZeroOnACleanCorpus is the negative control for the test
// above: without it, a binary that exited non-zero unconditionally would
// satisfy the error case and nothing would notice.
//
// AC: AC-SPEC-006.6
func TestBinaryExitsZeroOnACleanCorpus(t *testing.T) {
	bin := buildSpec(t)
	dir := corpusFrom(t, "slug-exists-found")

	stdout, code := run(t, bin, dir, "check")

	if code != 0 {
		t.Errorf("process exit status = %d, want 0\n--- stdout ---\n%s", code, stdout)
	}
}

// TestBinaryWritesTheIndexToStandardOutput exercises the whole path a
// consumer uses to pipe the index somewhere: the process, its argument
// vector, and its standard output. It holds the result to the same
// committed artifact the in-process snapshot case uses, so the two cannot
// drift.
//
// AC: AC-SPEC-010.3, AC-SPEC-011.1
func TestBinaryWritesTheIndexToStandardOutput(t *testing.T) {
	bin := buildSpec(t)
	dir := corpusFrom(t, "index-out-stdout")

	stdout, code := run(t, bin, dir, "index", "--out", "-")

	if code != 0 {
		t.Errorf("process exit status = %d, want 0", code)
	}

	want, err := os.ReadFile(filepath.Join("..", "testdata", "cases", "index-out-stdout", "expected", "manifest.json"))
	if err != nil {
		t.Fatalf("reading the case's expected index: %v", err)
	}
	if stdout != string(want) {
		t.Errorf("process stdout differs from the in-process snapshot for the same corpus.\n--- got ---\n%s\n--- want ---\n%s", stdout, want)
	}
}

// TestBinaryReportsAMissingConfigurationAndExitsNonZero covers the first
// thing an adopter gets wrong, through the process boundary: run the tool
// somewhere with no specs.toml. The status matters as much as the
// message, because this is the shape a CI job sees.
//
// AC: AC-SPEC-008.3
func TestBinaryReportsAMissingConfigurationAndExitsNonZero(t *testing.T) {
	bin := buildSpec(t)
	dir := t.TempDir() // deliberately empty: no specs.toml anywhere

	cmd := specCommand(t, bin, dir, "check")
	out, err := cmd.CombinedOutput()

	var exit *exec.ExitError
	if err == nil {
		t.Fatalf("expected a non-zero exit with no specs.toml present, got success\n%s", out)
	}
	if !asExitError(err, &exit) {
		t.Fatalf("running the binary: %v", err)
	}
	if exit.ExitCode() == 0 {
		t.Errorf("process exit status = 0, want non-zero")
	}
	if !containsPath(string(out), "specs.toml") {
		t.Errorf("the failure does not name the path it looked for, which AC-SPEC-008.3 requires:\n%s", out)
	}
}

// containsPath reports whether output mentions the given path fragment.
func containsPath(output, fragment string) bool {
	for i := 0; i+len(fragment) <= len(output); i++ {
		if output[i:i+len(fragment)] == fragment {
			return true
		}
	}
	return false
}

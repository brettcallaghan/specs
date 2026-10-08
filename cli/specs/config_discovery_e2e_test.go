// End-to-end tests for configuration discovery through the built binary:
// where specs.toml is found from, which file --config reads, and what a
// run reports when none is reachable.
package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// AC: AC-SPEC-008.7, AC-SPEC-008.1
//
// Run from a subdirectory of the corpus, the binary finds the specs.toml
// above it and reports every path relative to the directory holding that
// file — byte-identical to the same run from that directory.
func TestBinaryRunFromASubdirectoryFindsTheNearestSpecsToml(t *testing.T) {
	bin := buildSpec(t)
	dir := corpusFrom(t, "list-all")
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	stdout, code := run(t, bin, filepath.Join(dir, "docs", "adrs"), "list")
	if code != 0 {
		t.Fatalf("process exit status = %d, want 0\n%s", code, stdout)
	}
	want, err := os.ReadFile(filepath.Join("..", "testdata", "cases", "list-all", "expected", "checks.log"))
	if err != nil {
		t.Fatalf("reading the case's expected stdout: %v", err)
	}
	if stdout != string(want) {
		t.Errorf("run from a subdirectory differs from the run at the corpus root.\n--- got ---\n%s\n--- want ---\n%s", stdout, want)
	}
}

// AC: AC-SPEC-008.8
//
// --config reads the file it names even where a specs.toml sits in the
// working directory, and a --config naming no file reports that one path
// and searches nowhere else.
func TestBinaryConfigFlagReadsThatFileAndSearchesNoOther(t *testing.T) {
	bin := buildSpec(t)
	dir := corpusFrom(t, "list-all")
	other := corpusFrom(t, "list-empty-corpus")

	stdout, code := run(t, bin, other, "--config", filepath.Join(dir, "specs.toml"), "list")
	if code != 0 {
		t.Fatalf("process exit status = %d, want 0\n%s", code, stdout)
	}
	want, err := os.ReadFile(filepath.Join("..", "testdata", "cases", "list-all", "expected", "checks.log"))
	if err != nil {
		t.Fatalf("reading the case's expected stdout: %v", err)
	}
	if stdout != string(want) {
		t.Errorf("--config did not read the file it named.\n--- got ---\n%s\n--- want ---\n%s", stdout, want)
	}

	sub := filepath.Join(dir, "docs")
	cmd := specCommand(t, bin, sub, "--config", "nope.toml", "list")
	out, err := cmd.Output()
	var exit *exec.ExitError
	if !asExitError(err, &exit) || exit.ExitCode() != 3 {
		t.Fatalf("--config nope.toml: err = %v, want exit 3", err)
	}
	if len(out) != 0 {
		t.Errorf("stdout = %q, want nothing", out)
	}
	stderr := string(exit.Stderr)
	if !strings.Contains(stderr, "nope.toml") {
		t.Errorf("stderr does not name the path --config gave:\n%s", stderr)
	}
	if strings.Contains(stderr, dir) {
		t.Errorf("stderr names a searched directory; --config searches nothing:\n%s", stderr)
	}
}

// AC: AC-SPEC-008.9
//
// Outside any git repository the search runs to the filesystem root, so
// a specs.toml above the working directory with no .git anywhere is
// found.
func TestBinaryOutsideARepositorySearchesToTheFilesystemRoot(t *testing.T) {
	bin := buildSpec(t)
	dir := corpusFrom(t, "list-all")
	deep := filepath.Join(dir, "unpacked", "from", "an", "archive")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}

	stdout, code := run(t, bin, deep, "list")
	if code != 0 {
		t.Fatalf("process exit status = %d, want 0\n%s", code, stdout)
	}
	if !strings.Contains(stdout, "path: docs/") {
		t.Errorf("paths are not relative to the directory holding specs.toml:\n%s", stdout)
	}
}

// AC: AC-SPEC-011.5
//
// Help needs no configuration: discovery runs inside a command, never
// before one, so --help works in a directory with no specs.toml.
func TestBinaryHelpNeedsNoSpecsToml(t *testing.T) {
	bin := buildSpec(t)
	dir := t.TempDir()
	for _, args := range [][]string{{"--help"}, {"check", "--help"}} {
		if _, code := run(t, bin, dir, args...); code != 0 {
			t.Errorf("specs %v exited %d in a directory with no specs.toml, want 0", args, code)
		}
	}
}

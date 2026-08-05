// Snapshot case harness: replays the data-driven cases under
// cli/testdata/cases/ against the real command tree, per the Snapshot
// Case Structure specs/standards/testing-standards.md defines.
//
// A case is pure data — case.toml, an input/ corpus (omitted for a case
// that needs none, such as --help), and up to two expected artifacts —
// and carries no Go code of its own, which is what the standard asks for
// ("a harness independent of the test framework"). The one test function
// below is the only code this package requires to replay every one of
// them.
package cmd_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/brettcallaghan/agent-plugins/cli/cmd"
)

// casesDir is where every snapshot case lives, relative to this package.
const casesDir = "../testdata/cases"

// caseManifest is case.toml's shape: the argument vector a case is run
// with, the acceptance criteria it verifies (declared for a human or a
// future coverage tool — this harness does not itself enforce citation
// coverage), and the exit status the run must produce.
type caseManifest struct {
	Description string   `toml:"description"`
	Cites       []string `toml:"cites"`
	Args        []string `toml:"args"`
	ExitCode    int      `toml:"exit_code"`

	// Artifacts names every file under the case's expected/ directory
	// that this case asserts against — "checks.log", "manifest.json", or
	// both.
	//
	// It exists because the comparison functions below can only skip an
	// artifact they cannot find, and a skipped comparison is
	// indistinguishable from a passing one. Without a declaration, a case
	// whose artifact is deleted, renamed, or misspelled quietly degrades
	// to asserting nothing but its exit status, and keeps reporting
	// success. That is not hypothetical: every artifact comparison in
	// this suite silently skipped for several commits because the case
	// directory was resolved relatively and the run had already chdir'd
	// away from it (see the filepath.Abs note in runSnapshotCaseWith).
	// The bug was a wrong path; the reason it went unnoticed is that
	// nothing required the comparisons to happen.
	//
	// checkAssertionCoverage reconciles this list against expected/ in
	// both directions, so a missing artifact and an orphaned one are each
	// a failure rather than a silence.
	Artifacts []string `toml:"artifacts"`
}

// TestSnapshotCases replays every case under cli/testdata/cases/ and
// checks its exit status against case.toml and its output against
// whichever of the two expected artifacts the case carries.
//
// AC: AC-SPEC-002.5, AC-SPEC-006.5, AC-SPEC-006.6, AC-SPEC-006.8, AC-SPEC-011.1, AC-SPEC-011.2, AC-SPEC-011.3, AC-SPEC-011.4, AC-SPEC-011.5
func TestSnapshotCases(t *testing.T) {
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("reading %s: %v", casesDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			runSnapshotCase(t, filepath.Join(casesDir, name))
		})
	}
}

// runSnapshotCase runs one case directory to completion and checks its
// observable results, materializing its input corpus verbatim.
func runSnapshotCase(t *testing.T, dir string) {
	t.Helper()
	runSnapshotCaseWith(t, dir, copyTree)
}

// runSnapshotCaseWith is runSnapshotCase with the step that materializes
// the case's input corpus left to the caller.
//
// The seam exists for cli/cmd/portability_test.go, which replays every
// one of these cases a second time from a CRLF working tree and holds
// the result to the same committed artifacts (AC-SPEC-009.7). Nothing
// else about a case run differs between the two, so the check logic
// below is shared rather than copied — a divergence between the two
// runners would be a way for the portability claim to stop meaning what
// it says.
func runSnapshotCaseWith(t *testing.T, dir string, materialize func(src, dst string) error) {
	t.Helper()

	// Absolute before anything else. The run below chdirs into the
	// working copy, and every later read of the case's own files —
	// expected/checks.log, expected/manifest.json — would otherwise
	// resolve its relative path against that working copy, miss, and be
	// treated as "this case carries no such artifact". The comparisons
	// would then pass by not happening.
	dir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("resolving the case directory: %v", err)
	}

	var manifest caseManifest
	if _, err := toml.DecodeFile(filepath.Join(dir, "case.toml"), &manifest); err != nil {
		t.Fatalf("decoding case.toml: %v", err)
	}

	checkAssertionCoverage(t, dir, manifest)

	work := t.TempDir()
	inputDir := filepath.Join(dir, "input")
	if _, err := os.Stat(inputDir); err == nil {
		if err := materialize(inputDir, work); err != nil {
			t.Fatalf("copying the input corpus: %v", err)
		}
	} else if !os.IsNotExist(err) {
		t.Fatalf("statting %s: %v", inputDir, err)
	}

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(work); err != nil {
		t.Fatalf("Chdir(%s): %v", work, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origWD); err != nil {
			t.Fatalf("restoring the working directory: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	exitCode := cmd.Execute(manifest.Args, &stdout, &stderr)

	if exitCode != manifest.ExitCode {
		t.Errorf("exit code = %d, want %d\n--- stdout ---\n%s--- stderr ---\n%s",
			exitCode, manifest.ExitCode, stdout.String(), stderr.String())
	}

	checkChecksLog(t, dir, stdout.String(), manifest)
	checkManifest(t, dir, work, manifest.Args, stdout.String(), manifest)
}

// knownArtifacts is every expected/ filename this harness knows how to
// compare. A case naming anything else has a typo or is asking for a
// comparison that does not exist, and either way its declaration would
// otherwise be satisfied by a file nothing reads.
var knownArtifacts = map[string]bool{
	"checks.log":    true,
	"manifest.json": true,
}

// checkAssertionCoverage reconciles the case's declared artifacts against
// what is actually on disk under expected/, in both directions, and
// requires that it assert something beyond its exit status.
//
// This is the guard that makes a skipped comparison impossible to mistake
// for a passing one. Every failure it reports is a case that would
// otherwise still be green:
//
//   - declared but absent — the artifact was deleted or renamed, and the
//     comparison silently stopped happening;
//   - present but undeclared — the file is misspelled (expected/check.log)
//     or left over, so it is never read and the case asserts less than it
//     appears to;
//   - nothing declared — the case checks only its exit status, which for
//     a snapshot case is almost always an oversight rather than a choice.
func checkAssertionCoverage(t *testing.T, dir string, manifest caseManifest) {
	t.Helper()

	declared := map[string]bool{}
	for _, a := range manifest.Artifacts {
		if !knownArtifacts[a] {
			t.Errorf("case.toml declares artifact %q, which this harness cannot compare; known artifacts are checks.log and manifest.json", a)
			continue
		}
		if declared[a] {
			t.Errorf("case.toml declares artifact %q twice", a)
		}
		declared[a] = true
	}

	if len(declared) == 0 {
		t.Errorf("case.toml declares no artifacts, so this case asserts nothing but its exit status.\n" +
			"Add an expected/ artifact and name it in artifacts = [...]; a case that genuinely\n" +
			"cannot assert output should say why in its description and is probably a Go test.")
	}

	present := map[string]bool{}
	entries, err := os.ReadDir(filepath.Join(dir, "expected"))
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("reading expected/: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			present[e.Name()] = true
		}
	}

	for name := range declared {
		if !present[name] {
			t.Errorf("case.toml declares artifact %q but expected/%s does not exist — the comparison this case claims to make cannot happen", name, name)
		}
	}
	for name := range present {
		if !declared[name] {
			t.Errorf("expected/%s exists but case.toml does not declare it, so nothing reads it; add it to artifacts = [...] or delete it", name)
		}
	}
}

// checkChecksLog compares actual standard output against
// expected/checks.log, when the case declares one.
//
// A case that does not declare checks.log skips this comparison
// deliberately; checkAssertionCoverage has already established that the
// declaration and the directory agree, so the skip cannot be an accident.
func checkChecksLog(t *testing.T, dir, stdout string, manifest caseManifest) {
	t.Helper()

	if !slices.Contains(manifest.Artifacts, "checks.log") {
		return
	}

	want, err := os.ReadFile(filepath.Join(dir, "expected", "checks.log"))
	if err != nil {
		t.Fatalf("reading expected/checks.log: %v", err)
	}
	if stdout != string(want) {
		t.Errorf("stdout does not match expected/checks.log.\n--- got ---\n%s\n--- want ---\n%s", stdout, want)
	}
}

// checkManifest compares the produced index against expected/manifest.json,
// when the case declares one. Where the produced index landed depends on
// the case's own --out argument: standard output when it is "-", the
// named file otherwise, and the default specs.json inside the spec root
// when --out was not given at all.
//
// As with checkChecksLog, the skip is driven by the declaration rather
// than by whether the file happens to be readable.
func checkManifest(t *testing.T, dir, work string, args []string, stdout string, manifest caseManifest) {
	t.Helper()

	if !slices.Contains(manifest.Artifacts, "manifest.json") {
		return
	}

	wantPath := filepath.Join(dir, "expected", "manifest.json")
	want, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("reading expected/manifest.json: %v", err)
	}

	out, hasOut := outFlagValue(args)
	var got []byte
	switch {
	case hasOut && out == "-":
		got = []byte(stdout)
	case hasOut && out != "":
		got, err = os.ReadFile(filepath.Join(work, out))
	default:
		var indexPath string
		indexPath, err = defaultIndexPath(work)
		if err == nil {
			got, err = os.ReadFile(indexPath)
		}
	}
	if err != nil {
		t.Fatalf("reading the produced index: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("the produced index does not match expected/manifest.json.\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// outFlagValue returns the value passed to --out in args, and whether one
// was given.
func outFlagValue(args []string) (string, bool) {
	for i, a := range args {
		if a == "--out" && i+1 < len(args) {
			return args[i+1], true
		}
		if v, ok := strings.CutPrefix(a, "--out="); ok {
			return v, true
		}
	}
	return "", false
}

// defaultIndexPath resolves where a run with no --out writes its index:
// specs.json inside the workspace root declared by work's own specs.toml
// — the same default corpus.IndexPath computes, read back from the case's
// own configuration rather than duplicated as a second constant.
func defaultIndexPath(work string) (string, error) {
	var cfg struct {
		Workspace struct {
			Root string `toml:"root"`
		} `toml:"workspace"`
	}
	if _, err := toml.DecodeFile(filepath.Join(work, "specs.toml"), &cfg); err != nil {
		return "", err
	}
	return filepath.Join(work, cfg.Workspace.Root, "specs.json"), nil
}

// copyTree copies every file under src into dst, preserving relative
// paths, so a snapshot case's committed input/ corpus is never mutated by
// the run exercising it.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

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
// observable results.
func runSnapshotCase(t *testing.T, dir string) {
	t.Helper()

	var manifest caseManifest
	if _, err := toml.DecodeFile(filepath.Join(dir, "case.toml"), &manifest); err != nil {
		t.Fatalf("decoding case.toml: %v", err)
	}

	work := t.TempDir()
	inputDir := filepath.Join(dir, "input")
	if _, err := os.Stat(inputDir); err == nil {
		if err := copyTree(inputDir, work); err != nil {
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

	checkChecksLog(t, dir, stdout.String())
	checkManifest(t, dir, work, manifest.Args, stdout.String())
}

// checkChecksLog compares actual standard output against
// expected/checks.log, when the case carries one.
func checkChecksLog(t *testing.T, dir, stdout string) {
	t.Helper()

	want, err := os.ReadFile(filepath.Join(dir, "expected", "checks.log"))
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatalf("reading expected/checks.log: %v", err)
	}
	if stdout != string(want) {
		t.Errorf("stdout does not match expected/checks.log.\n--- got ---\n%s\n--- want ---\n%s", stdout, want)
	}
}

// checkManifest compares the produced index against expected/manifest.json,
// when the case carries one. Where the produced index landed depends on
// the case's own --out argument: standard output when it is "-", the
// named file otherwise, and the default specs.json inside the spec root
// when --out was not given at all.
func checkManifest(t *testing.T, dir, work string, args []string, stdout string) {
	t.Helper()

	wantPath := filepath.Join(dir, "expected", "manifest.json")
	want, err := os.ReadFile(wantPath)
	if os.IsNotExist(err) {
		return
	}
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

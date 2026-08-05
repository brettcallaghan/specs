// Command gen regenerates the expected artifacts of every snapshot case
// under cli/testdata/cases/ from the current implementation, per
// specs/standards/testing-standards.md's "a script that regenerates every
// expected artifact... SHALL be provided."
//
// It is not itself the test: cli/cmd/snapshot_test.go replays the same
// cases and compares against whatever is committed. This program only
// produces the committed side, and writes exactly the artifacts each
// case declares in its own case.toml — a standard-output artifact
// (expected/checks.log for a human-readable run, expected/checks.json
// for a --json one) and/or expected/manifest.json for the produced
// index. A case that declares none gets none.
//
// Usage, from the cli module root:
//
//	go run ./testdata/gen ./testdata/cases
//
// or via the wrapper: scripts/regenerate-cli-snapshots.sh.
//
// WHEN a regenerated snapshot differs from the committed one, the diff
// SHALL be reviewed before committing — this program does not do that
// for you.
package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/brettcallaghan/agent-plugins/cli/cmd"
)

// caseManifest mirrors cli/cmd/snapshot_test.go's caseManifest. The two
// are not shared code because the test harness intentionally imports
// nothing from this program (a case is data, not code shared with its
// own generator), the same separation the standard's "no executable test
// code in a case" draws between a case and whatever replays it.
type caseManifest struct {
	Args     []string `toml:"args"`
	ExitCode int      `toml:"exit_code"`

	// Artifacts is the case's own declaration of what it asserts
	// against, and is what this program writes — nothing more.
	//
	// It replaces two hardcoded name tables (indexProducingCases and
	// omitChecksLog) that this program used to carry. They were a
	// standing invitation to drift: the harness decides what to compare
	// from this list, so a case whose artifacts the tables disagreed
	// with got an expected/ directory the harness then rejected as
	// orphaned — or, worse, silently wrote a JSON payload to a name
	// ending in .log. Reading the declaration means the generator and
	// the harness cannot disagree about which files a case owns.
	Artifacts []string `toml:"artifacts"`
}

// stdoutArtifacts are the two names a case may give its standard-output
// artifact, mirroring cli/cmd/snapshot_test.go's list of the same name:
// checks.log for a human-readable run, checks.json for a --json one. A
// case declares at most one of them.
var stdoutArtifacts = []string{"checks.log", "checks.json"}

// manifestArtifact is the name a case gives the index a run produced.
const manifestArtifact = "manifest.json"

// declares reports whether the case asserts against the named artifact.
func (m caseManifest) declares(name string) bool {
	return slices.Contains(m.Artifacts, name)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: gen <cases-dir>")
		os.Exit(2)
	}
	casesDir := os.Args[1]

	entries, err := os.ReadDir(casesDir)
	if err != nil {
		fatalf("reading %s: %v", casesDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if err := regenerate(filepath.Join(casesDir, entry.Name()), entry.Name()); err != nil {
			fatalf("%s: %v", entry.Name(), err)
		}
		fmt.Println("regenerated", entry.Name())
	}
}

func regenerate(dir, name string) error {
	// Absolute before anything else, because this function chdir's into a
	// scratch directory below and every later use of dir — reading
	// input/, writing expected/ — would otherwise resolve against the
	// scratch directory instead of the case.
	//
	// This is the identical defect cli/cmd/snapshot_test.go records
	// against its own replay path ("the case directory was resolved
	// relatively and the run had already chdir'd away from it"). The fix
	// landed there and not here, so this program went on reporting
	// "regenerated <case>" for every case while writing every artifact
	// into a temp directory it then deleted. Measured 2026-08-05:
	// corrupting a committed checks.log and re-running the wrapper script
	// left the corruption in place and printed success for all twelve
	// cases.
	dir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolving the case directory: %w", err)
	}

	var manifest caseManifest
	if _, err := toml.DecodeFile(filepath.Join(dir, "case.toml"), &manifest); err != nil {
		return fmt.Errorf("decoding case.toml: %w", err)
	}

	work, err := os.MkdirTemp("", "spec-case-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)

	inputDir := filepath.Join(dir, "input")
	if _, err := os.Stat(inputDir); err == nil {
		if err := copyTree(inputDir, work); err != nil {
			return fmt.Errorf("copying input/: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	origWD, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(origWD)
	if err := os.Chdir(work); err != nil {
		return err
	}

	var stdout, stderr bytes.Buffer
	exitCode := cmd.Execute(manifest.Args, &stdout, &stderr)
	if exitCode != manifest.ExitCode {
		fmt.Fprintf(os.Stderr,
			"WARNING %s: exit code = %d, case.toml declares %d (case.toml was left untouched; update it by hand if this is the intended new status)\nstderr:\n%s",
			name, exitCode, manifest.ExitCode, stderr.String())
	}

	expectedDir := filepath.Join(dir, "expected")
	if err := os.MkdirAll(expectedDir, 0o755); err != nil {
		return err
	}

	for _, artifact := range stdoutArtifacts {
		if !manifest.declares(artifact) {
			continue
		}
		if err := os.WriteFile(filepath.Join(expectedDir, artifact), []byte(stdout.String()), 0o644); err != nil {
			return err
		}
	}

	if manifest.declares(manifestArtifact) {
		got, err := producedIndex(work, manifest.Args, stdout.String())
		if err != nil {
			return fmt.Errorf("reading the produced index: %w", err)
		}
		if err := os.WriteFile(filepath.Join(expectedDir, manifestArtifact), got, 0o644); err != nil {
			return err
		}
	}

	return nil
}

// producedIndex reads back the index a run produced, from wherever its
// own --out argument sent it.
func producedIndex(work string, args []string, stdout string) ([]byte, error) {
	out, hasOut := outFlagValue(args)
	switch {
	case hasOut && out == "-":
		return []byte(stdout), nil
	case hasOut && out != "":
		return os.ReadFile(filepath.Join(work, out))
	default:
		indexPath, err := defaultIndexPath(work)
		if err != nil {
			return nil, err
		}
		return os.ReadFile(indexPath)
	}
}

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

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

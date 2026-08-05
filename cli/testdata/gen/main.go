// Command gen regenerates the expected artifacts of every snapshot case
// under cli/testdata/cases/ from the current implementation, per
// specs/standards/testing-standards.md's "a script that regenerates every
// expected artifact... SHALL be provided."
//
// It is not itself the test: cli/cmd/snapshot_test.go replays the same
// cases and compares against whatever is committed. This program only
// produces the committed side, and only for the two artifacts a case may
// declare — expected/checks.log (standard output) and
// expected/manifest.json (the produced index) — never both when the
// case's argument vector produces only one.
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
}

// indexProducingCases names every case whose argument vector builds an
// index, and so carries expected/manifest.json rather than (or, for a
// case that also reports where it wrote the file, alongside)
// expected/checks.log.
//
// This table exists because a case's own argument vector does not by
// itself say whether its expected artifact is a check report or a
// produced index — both `check --json` and `index --out -` write JSON to
// standard output, and only one of them is a manifest. Declaring which
// case is which here (repeated as the harness's own skipChecksLog and
// wantsManifest below) is this program's alternative to inspecting the
// case name.
var indexProducingCases = map[string]bool{
	"index-default":    true,
	"index-out-custom": true,
	"index-out-stdout": true,
}

// omitChecksLog names a case whose argument vector produces only the
// index artifact, so expected/checks.log is omitted entirely rather than
// duplicating expected/manifest.json's content under a second name
// (testing-standards.md: "the case SHALL omit the other").
var omitChecksLog = map[string]bool{
	"index-out-stdout": true,
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

	if !omitChecksLog[name] {
		if err := os.WriteFile(filepath.Join(expectedDir, "checks.log"), []byte(stdout.String()), 0o644); err != nil {
			return err
		}
	}

	if indexProducingCases[name] {
		got, err := producedIndex(work, manifest.Args, stdout.String())
		if err != nil {
			return fmt.Errorf("reading the produced index: %w", err)
		}
		if err := os.WriteFile(filepath.Join(expectedDir, "manifest.json"), got, 0o644); err != nil {
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

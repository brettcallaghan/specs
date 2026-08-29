// The end-to-end harness: restore a fixture, run the compiled binary
// against it, compare what came back with what is committed.
package e2e

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

const (
	fixturesDir = "../../testdata/swarm-fixtures"
	casesDir    = "../../testdata/swarm-cases"

	// requireTrackerEnv is how a project declares the platform its cases
	// must not be skipped on. A case needing an absent tracker is skipped
	// with the reason everywhere else; where this is set, it fails
	// instead — otherwise "the tracker was not installed" and "the suite
	// has no coverage" become the same green run.
	requireTrackerEnv = "SWARM_E2E_REQUIRE_TRACKER"
)

// fixtureStamp is fixture.toml's shape.
type fixtureStamp struct {
	Kind            string `toml:"kind"`
	ContractVersion string `toml:"contract_version"`
	Description     string `toml:"description"`
	Prefix          string `toml:"prefix"`
	Run             string `toml:"run"`
}

// caseManifest is case.toml's shape: which fixture the case restores,
// what the binary is invoked with, the status it must exit with, and the
// substitutions applied to its output before comparison.
type caseManifest struct {
	Fixture string   `toml:"fixture"`
	Args    []string `toml:"args"`
	Exit    int      `toml:"exit"`
	// AC names the acceptance criteria the case verifies, for a human and
	// for a future coverage tool. This harness does not enforce citation.
	AC []string `toml:"ac"`
	// Normalize is the substitution list, declared per case and reviewed
	// with it. Too little and every case fails; too much and the case
	// stops asserting what it was written for.
	Normalize []substitution `toml:"normalize"`
	// After is a second invocation run against the same restored world,
	// which is how a case asserts what the first one did — or, for a dry
	// run, did not do.
	After []afterInvocation `toml:"after"`
}

type substitution struct {
	Pattern string `toml:"pattern"`
	With    string `toml:"with"`
	Why     string `toml:"why"`
}

type afterInvocation struct {
	Args []string `toml:"args"`
	Exit int      `toml:"exit"`
}

// world is a restored fixture: a repository, a tracker, and the
// environment a run against them needs.
type world struct {
	dir    string
	repo   string
	home   string
	stamp  fixtureStamp
	binary string
}

// requireTools reports whether the tools a case needs are present, and
// either skips or fails according to what the project declared.
func requireTools(t *testing.T) {
	t.Helper()

	var missing []string
	for _, tool := range []string{"git", "bd"} {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}
	if len(missing) == 0 {
		return
	}

	reason := fmt.Sprintf("%s not on PATH, so the end-to-end tier cannot restore a fixture",
		strings.Join(missing, " and "))
	if os.Getenv(requireTrackerEnv) != "" {
		t.Fatalf("%s — and %s is set, which declares this platform one the tools are present on",
			reason, requireTrackerEnv)
	}
	t.Skip(reason)
}

// buildBinary compiles the swarm binary once per test binary.
func buildBinary(t *testing.T) string {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping the end-to-end tier in -short mode (it compiles the binary)")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("no go toolchain on PATH: %v", err)
	}

	name := "swarm"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(t.TempDir(), name)
	build := exec.Command("go", "build", "-o", bin, "../")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building the swarm binary: %v\n%s", err, out)
	}
	return bin
}

// restore materializes a fixture into a temporary directory of its own.
//
// Never in place and never rolled back: a case that shared a world with
// another could lose it, and a case that rolled one back would be
// trusting the rollback.
func restore(t *testing.T, binary, fixture string) world {
	t.Helper()

	var stamp fixtureStamp
	stampPath := filepath.Join(fixturesDir, fixture, "fixture.toml")
	if _, err := toml.DecodeFile(stampPath, &stamp); err != nil {
		t.Fatalf("reading %s: %v", stampPath, err)
	}

	dir := t.TempDir()
	w := world{
		dir:    dir,
		repo:   filepath.Join(dir, "repo"),
		home:   filepath.Join(dir, "home"),
		stamp:  stamp,
		binary: binary,
	}
	if err := os.MkdirAll(w.home, 0o755); err != nil {
		t.Fatalf("making the fixture's home: %v", err)
	}

	bundle, err := filepath.Abs(filepath.Join(fixturesDir, fixture, "repository.bundle"))
	if err != nil {
		t.Fatalf("resolving the bundle: %v", err)
	}
	export, err := filepath.Abs(filepath.Join(fixturesDir, fixture, "tracker.jsonl"))
	if err != nil {
		t.Fatalf("resolving the export: %v", err)
	}

	w.mustRun(t, dir, "git", "clone", "--quiet", bundle, w.repo)
	// The bundle's HEAD is the sprint branch, and a clone of a bundle
	// checks it out; naming it anyway makes the case independent of what
	// the bundle happened to record as HEAD.
	w.mustRun(t, w.repo, "git", "checkout", "--quiet", "claude/bead-bonanza-fixture")
	w.mustRun(t, w.repo, "bd", "init", "--prefix", stamp.Prefix)
	w.mustRun(t, w.repo, "bd", "import", export)

	return w
}

// mustRun runs a restore step and fails the test with what it said.
func (w world) mustRun(t *testing.T, dir, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = w.env()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("restoring the fixture: %s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
}

// env is the environment a run against this world needs.
//
// HOME is pinned into the world so the tracker writes its own state
// there rather than into the running user's, and so one case cannot see
// what another left behind.
func (w world) env() []string {
	return append(os.Environ(),
		"HOME="+w.home,
		"BEADS_ACTOR=fixture-e2e",
		// Locale-independent output from every tool the run shells out
		// to; the contract forbids a field whose meaning depends on the
		// producing process's locale, and this is the other half of that.
		"LC_ALL=C",
		"TZ=UTC",
	)
}

// swarm runs the compiled binary in the restored repository and returns
// its stdout, its stderr and its exit status.
func (w world) swarm(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()

	var out, errOut strings.Builder
	cmd := exec.Command(w.binary, args...)
	cmd.Dir = w.repo
	cmd.Env = w.env()
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatalf("running the swarm binary: %v", err)
		}
		return out.String(), errOut.String(), exit.ExitCode()
	}
	return out.String(), errOut.String(), 0
}

// normalized applies a case's declared substitutions.
func normalized(t *testing.T, text string, subs []substitution) string {
	t.Helper()

	for _, sub := range subs {
		re, err := regexp.Compile(sub.Pattern)
		if err != nil {
			t.Fatalf("the case's normalization pattern %q does not compile: %v", sub.Pattern, err)
		}
		text = re.ReplaceAllString(text, sub.With)
	}
	return text
}

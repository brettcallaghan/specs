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
	// Worktree is the dispatch worktrees standing in this world.
	//
	// Declared here rather than created by each case, because a bundle
	// cannot carry one: a worktree is a registration under .git, and a
	// bundle carries refs and objects. A world whose worktrees were
	// created by the case would put half the world in the case, and two
	// cases reading the same world would then have to agree by hand.
	Worktree []worktreeDecl `toml:"worktree"`
	// RefusingRemote, where declared, is a remote that rejects a push.
	RefusingRemote *refusingRemote `toml:"refusing_remote"`
}

// worktreeDecl is one standing dispatch worktree.
type worktreeDecl struct {
	// Branch is checked out in the worktree; it must exist in the bundle.
	Branch string `toml:"branch"`
	// Path is relative to the world directory, which is the parent of the
	// repository — the location a dispatch worktree really takes.
	Path string `toml:"path"`
}

// refusingRemote is a remote whose pre-receive hook rejects the refs
// matching a pattern, which is how a push refusal is made to happen
// rather than simulated.
//
// The pattern is declared rather than fixed because what a remote refuses
// is the interesting half: a remote refusing everything and one refusing
// a single ref namespace produce different reports, and a world that
// could only express the first could not prove the second.
type refusingRemote struct {
	Name    string `toml:"name"`
	Refuses string `toml:"refuses"`
	Message string `toml:"message"`
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

	initArgs := []string{"init", "--prefix", stamp.Prefix}
	if runtime.GOOS == "windows" {
		// ADR-027: Windows runs against the CGO-free, server-mode `bd`
		// install, which carries no embedded Dolt engine — `--server` is
		// required or the restore fails outright. Nothing else is
		// needed: `bd init --server` auto-starts and manages its own
		// `dolt sql-server` when it finds none already listening on the
		// port it tries, which ADR-027 marked unmeasured. Measured here
		// (Linux, dolt 2.3.1, no --server-port given): two `bd init
		// --server` calls made back to back came up on two different
		// auto-selected ports (36331, then 37165) with no conflict and
		// no server started by hand — which is what lets this suite's
		// cases, run with t.Parallel(), each restore its own fixture
		// without colliding on one.
		initArgs = append(initArgs, "--server")
	}
	w.mustRun(t, w.repo, "bd", initArgs...)
	w.mustRun(t, w.repo, "bd", "import", export)

	w.standWorktrees(t)
	w.standRefusingRemote(t)

	return w
}

// standWorktrees registers the worktrees the fixture declares.
//
// They are created here rather than in a case because the world is what
// a fixture is: a case that had to create half of it would be asserting
// against a world it built, which is the thing this tier exists not to
// do.
func (w world) standWorktrees(t *testing.T) {
	t.Helper()

	for _, decl := range w.stamp.Worktree {
		if decl.Branch == "" || decl.Path == "" {
			t.Fatalf("a worktree declaration names a branch and a path; got %+v", decl)
		}
		path := filepath.Join(w.dir, filepath.FromSlash(decl.Path))
		w.mustRun(t, w.repo, "git", "worktree", "add", "--quiet",
			"--checkout", path, decl.Branch)
	}
}

// standRefusingRemote stands up a remote that rejects the refs matching
// the fixture's declared pattern.
//
// A real bare repository with a real pre-receive hook, because the point
// is that the refusal comes back through git the way a server's does —
// a remote faked at any higher level would prove the fake.
func (w world) standRefusingRemote(t *testing.T) {
	t.Helper()

	decl := w.stamp.RefusingRemote
	if decl == nil {
		return
	}
	if decl.Name == "" || decl.Refuses == "" {
		t.Fatalf("a refusing remote names a remote and the refs it refuses; got %+v", *decl)
	}
	if runtime.GOOS == "windows" {
		t.Skip("the refusing remote's pre-receive hook is a shell script, which this platform does not run")
	}

	remote := filepath.Join(w.dir, "refusing-remote.git")
	w.mustRun(t, w.dir, "git", "init", "--quiet", "--bare", remote)

	message := decl.Message
	if message == "" {
		message = "the remote refuses this ref"
	}
	// Reads the (oldsha newsha ref) triples git feeds a pre-receive hook
	// and rejects the whole push when any ref matches, which is what a
	// server refusing one path of a push does.
	hook := fmt.Sprintf(`#!/bin/sh
while read -r _old _new ref; do
  case "$ref" in
    %s) echo "%s" >&2; exit 1 ;;
  esac
done
exit 0
`, decl.Refuses, message)
	hookPath := filepath.Join(remote, "hooks", "pre-receive")
	if err := os.WriteFile(hookPath, []byte(hook), 0o755); err != nil {
		t.Fatalf("writing the refusing remote's hook: %v", err)
	}

	// Replace rather than add: the clone's own origin is the bundle file,
	// which no push reaches, so a case pushing to "origin" must reach
	// this one.
	if decl.Name == "origin" {
		w.mustRun(t, w.repo, "git", "remote", "remove", decl.Name)
	}
	w.mustRun(t, w.repo, "git", "remote", "add", decl.Name, remote)
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

// run runs a tool in the world and returns what it said and its status,
// for the callers that need to inspect a failure rather than stop on it.
func (w world) run(t *testing.T, dir, name string, args ...string) (stdout, stderr string, code int) {
	t.Helper()

	var out, errOut strings.Builder
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = w.env()
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatalf("running %s %s: %v", name, strings.Join(args, " "), err)
		}
		return out.String(), errOut.String(), exit.ExitCode()
	}
	return out.String(), errOut.String(), 0
}

// mustCapture runs a tool that has to succeed and returns its trimmed
// stdout.
func (w world) mustCapture(t *testing.T, dir, name string, args ...string) string {
	t.Helper()

	stdout, stderr, code := w.run(t, dir, name, args...)
	if code != 0 {
		t.Fatalf("%s %s: exit %d\n%s\n%s", name, strings.Join(args, " "), code, stdout, stderr)
	}
	return strings.TrimSpace(stdout)
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

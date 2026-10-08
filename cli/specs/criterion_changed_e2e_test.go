// End-to-end tests for `criterion --changed`, which compares the corpus
// against the corpus at a git base. A snapshot case's input cannot hold a
// git repository, so each test here builds its repository at test time
// with the git executable and runs the real binary in it.
//
// TestMain isolates every git these tests and the binary run from the
// machine's configuration and from any GIT_* variable the test process
// inherited: a hook running go test sets GIT_DIR, which would point every
// invocation at the repository the test was started from.
package main_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(runIsolatedFromGit(m))
}

func runIsolatedFromGit(m *testing.M) int {
	for _, kv := range os.Environ() {
		if name, _, _ := strings.Cut(kv, "="); strings.HasPrefix(name, "GIT_") {
			os.Unsetenv(name)
		}
	}
	scratch, err := os.MkdirTemp("", "specs-e2e-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(scratch)
	sharedBinDir = filepath.Join(scratch, "bin")
	global := filepath.Join(scratch, "gitconfig")
	if err := os.WriteFile(global, nil, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	os.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	os.Setenv("GIT_CONFIG_GLOBAL", global)
	return m.Run()
}

var (
	sharedBinDir  string
	sharedBinOnce sync.Once
	sharedBin     string
)

// sharedSpec builds the binary once for every change-listing test.
func sharedSpec(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping end-to-end binary test in -short mode (it compiles the binary)")
	}
	sharedBinOnce.Do(func() {
		if err := os.MkdirAll(sharedBinDir, 0o755); err != nil {
			t.Fatalf("creating %s: %v", sharedBinDir, err)
		}
		sharedBin = buildSpecIn(t, sharedBinDir)
	})
	if sharedBin == "" {
		t.Fatal("the specs binary failed to build in an earlier test")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("no git on PATH: %v", err)
	}
	return sharedBin
}

// runAll executes the binary in dir and returns stdout, stderr and the
// exit status.
func runAll(t *testing.T, bin, dir string, args ...string) (string, string, int) {
	t.Helper()
	cmd := specCommand(t, bin, dir, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		var exit *exec.ExitError
		if !asExitError(err, &exit) {
			t.Fatalf("running %s %v: %v", bin, args, err)
		}
		return stdout.String(), stderr.String(), exit.ExitCode()
	}
	return stdout.String(), stderr.String(), 0
}

// gitIn runs git in dir for a test, failing it on a non-zero exit.
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir,
		"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)
	cmd := exec.Command("git", full...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(stdout.String())
}

// commitAll stages everything in dir and commits it.
func commitAll(t *testing.T, dir, message string) string {
	t.Helper()
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "--allow-empty", "-m", message)
	return gitIn(t, dir, "rev-parse", "HEAD")
}

// writeFile writes content at the slash-separated rel beneath dir.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func removeFile(t *testing.T, dir, rel string) {
	t.Helper()
	if err := os.Remove(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
		t.Fatal(err)
	}
}

// scenario is one acceptance criterion of a fixture FRD.
type scenario struct {
	id   string // the criterion's identifier, e.g. "AC-CHK-001.1"
	then string // its Then step, which is what an edit changes
}

// frd renders a conforming FRD declaring feature with the scenarios in
// order.
func frd(slug, feature string, scenarios ...scenario) string {
	var b strings.Builder
	fmt.Fprintf(&b, "---\ntitle: %q\ndescription: A fixture FRD.\nslug: %s\ntype: frd\nstatus: approved\nupdated_at: 2026-10-08T00:00:00Z\n---\n\n# %s\n\n```gherkin\n@%s\nFeature: %s\n\n  As a shopper\n  I want to pay\n  So that the goods are mine\n",
		slug, slug, slug, feature, slug)
	for _, sc := range scenarios {
		fmt.Fprintf(&b, "\n  @%s\n  Scenario: Criterion %s\n    Given a basket\n    When the shopper pays\n    Then %s\n", sc.id, sc.id, sc.then)
	}
	b.WriteString("```\n")
	return b.String()
}

const specsTOML = "[workspace]\nname = \"acme\"\nroot = \"docs\"\n\n[profile]\nname = \"specs\"\n"

const checkoutPath = "docs/features/frd-checkout.md"

// newSpecRepo initializes a repository on main holding specs.toml and,
// when scenarios are given, a checkout FRD declaring them, committed.
func newSpecRepo(t *testing.T, scenarios ...scenario) (dir, head string) {
	t.Helper()
	dir = t.TempDir()
	gitIn(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "specs.toml", specsTOML)
	if len(scenarios) > 0 {
		writeFile(t, dir, checkoutPath, frd("frd-checkout", "FEAT-CHK-001", scenarios...))
	}
	return dir, commitAll(t, dir, "base")
}

// change is one expected row of a change listing.
type change struct{ id, classification, path string }

// wantYAML renders the answer the listing must print.
func wantYAML(revision, commit string, changes ...change) string {
	var b strings.Builder
	fmt.Fprintf(&b, "base:\n  revision: %s\n  commit: %s\n", revision, commit)
	if len(changes) == 0 {
		b.WriteString("criteria: []\n")
		return b.String()
	}
	b.WriteString("criteria:\n")
	for _, c := range changes {
		fmt.Fprintf(&b, "  - id: %s\n    classification: %s\n    path: %s\n", c.id, c.classification, c.path)
	}
	return b.String()
}

func wantJSON(revision, commit string, changes ...change) string {
	rows := make([]string, 0, len(changes))
	for _, c := range changes {
		rows = append(rows, fmt.Sprintf(`{"id":%q,"classification":%q,"path":%q}`, c.id, c.classification, c.path))
	}
	return fmt.Sprintf(`{"base":{"revision":%q,"commit":%q},"criteria":[%s]}`+"\n", revision, commit, strings.Join(rows, ","))
}

// assertAnswer runs --changed with extra in dir and holds stdout to want
// and the exit status to 0.
func assertAnswer(t *testing.T, bin, dir, want string, extra ...string) {
	t.Helper()
	args := append([]string{"criterion", "--changed"}, extra...)
	stdout, stderr, code := runAll(t, bin, dir, args...)
	if code != 0 {
		t.Fatalf("%v exited %d, want 0\n--- stderr ---\n%s", args, code, stderr)
	}
	if stdout != want {
		t.Errorf("%v printed\n%s\nwant\n%s", args, stdout, want)
	}
}

// noIndexWritten fails the test when a specs.json exists anywhere in dir.
func noIndexWritten(t *testing.T, dir string) {
	t.Helper()
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && d.Name() == "specs.json" {
			t.Errorf("%s exists: the change listing reads and writes no persisted index", path)
		}
		if d != nil && d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		return nil
	})
}

// A branch that rewords one criterion, adds one and removes a document
// holding another lists exactly those three against the merge-base with
// main, under base and then criteria, sorted by identifier, each with the
// path of its document — the base's for the removed one.
//
// AC: AC-GHK-008.1, AC-GHK-008.3, AC-GHK-008.11, AC-GHK-008.7
func TestChangedListsChangedAddedAndRemovedAgainstTheMergeBase(t *testing.T) {
	bin := sharedSpec(t)
	dir, base := newSpecRepo(t,
		scenario{"AC-CHK-001.1", "the order is placed"},
		scenario{"AC-CHK-001.2", "payment is refused"})
	writeFile(t, dir, "docs/features/frd-legacy.md", frd("frd-legacy", "FEAT-LEG-001",
		scenario{"AC-LEG-001.1", "the legacy flow runs"}))
	base = commitAll(t, dir, "legacy")
	commitAll(t, dir, "main moves on after the branch point")
	mainTip := gitIn(t, dir, "rev-parse", "HEAD")

	gitIn(t, dir, "checkout", "-q", "-b", "work", base)
	writeFile(t, dir, checkoutPath, frd("frd-checkout", "FEAT-CHK-001",
		scenario{"AC-CHK-001.1", "the order is placed"},
		scenario{"AC-CHK-001.2", "payment is refused and the shopper is told why"},
		scenario{"AC-CHK-001.4", "it is settled exactly once"}))
	removeFile(t, dir, "docs/features/frd-legacy.md")
	commitAll(t, dir, "work")

	rows := []change{
		{"AC-CHK-001.2", "changed", checkoutPath},
		{"AC-CHK-001.4", "added", checkoutPath},
		{"AC-LEG-001.1", "removed", "docs/features/frd-legacy.md"},
	}
	if base == mainTip {
		t.Fatal("fixture: main did not move past the branch point")
	}
	assertAnswer(t, bin, dir, wantYAML("main", base, rows...))
	assertAnswer(t, bin, dir, wantJSON("main", base, rows...), "--json")
	noIndexWritten(t, dir)
}

// A criterion that moved within its document, and a document that moved
// to another path, are not listed when their content is unchanged.
//
// AC: AC-GHK-008.2
func TestChangedDoesNotListACriterionThatOnlyMoved(t *testing.T) {
	bin := sharedSpec(t)
	dir, base := newSpecRepo(t,
		scenario{"AC-CHK-001.1", "the order is placed"},
		scenario{"AC-CHK-001.2", "payment is refused"})

	removeFile(t, dir, checkoutPath)
	writeFile(t, dir, "docs/requirements/frd-checkout.md", frd("frd-checkout", "FEAT-CHK-001",
		scenario{"AC-CHK-001.2", "payment is refused"},
		scenario{"AC-CHK-001.1", "the order is placed"}))

	assertAnswer(t, bin, dir, wantYAML("main", base))
}

// Edits not yet committed — a modified tracked file and an untracked one —
// are part of the current side.
//
// AC: AC-GHK-008.6
func TestChangedListsUncommittedAndUntrackedEdits(t *testing.T) {
	bin := sharedSpec(t)
	dir, base := newSpecRepo(t, scenario{"AC-CHK-001.1", "the order is placed"})

	writeFile(t, dir, checkoutPath, frd("frd-checkout", "FEAT-CHK-001",
		scenario{"AC-CHK-001.1", "the order is placed and confirmed"}))
	writeFile(t, dir, "docs/features/frd-returns.md", frd("frd-returns", "FEAT-RET-001",
		scenario{"AC-RET-001.1", "the refund is issued"}))

	assertAnswer(t, bin, dir, wantYAML("main", base,
		change{"AC-CHK-001.1", "changed", checkoutPath},
		change{"AC-RET-001.1", "added", "docs/features/frd-returns.md"}))
}

// Rebuilding the persisted index after an edit does not change the
// answer: the comparison never reads it.
//
// AC: AC-GHK-008.7
func TestChangedIsUnaffectedByRebuildingTheIndex(t *testing.T) {
	bin := sharedSpec(t)
	dir, base := newSpecRepo(t, scenario{"AC-CHK-001.1", "the order is placed"})
	writeFile(t, dir, checkoutPath, frd("frd-checkout", "FEAT-CHK-001",
		scenario{"AC-CHK-001.1", "the order is placed and confirmed"}))
	want := wantYAML("main", base, change{"AC-CHK-001.1", "changed", checkoutPath})

	assertAnswer(t, bin, dir, want)
	for _, args := range [][]string{{"index"}, {"list"}} {
		if _, stderr, code := runAll(t, bin, dir, args...); code != 0 {
			t.Fatalf("%v exited %d\n%s", args, code, stderr)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "docs", "specs.json")); err != nil {
		t.Fatalf("specs index wrote no docs/specs.json: %v", err)
	}
	assertAnswer(t, bin, dir, want)
}

// --since names the revision the base is the merge-base with: a branch,
// a tag and a commit.
//
// AC: AC-GHK-008.5
func TestChangedSinceComparesAgainstTheMergeBaseWithTheNamedRevision(t *testing.T) {
	bin := sharedSpec(t)
	dir, first := newSpecRepo(t, scenario{"AC-CHK-001.1", "the order is placed"})
	gitIn(t, dir, "tag", "v1")
	gitIn(t, dir, "branch", "side")
	writeFile(t, dir, checkoutPath, frd("frd-checkout", "FEAT-CHK-001",
		scenario{"AC-CHK-001.1", "the order is placed and confirmed"}))
	second := commitAll(t, dir, "reword")

	gitIn(t, dir, "checkout", "-q", "side")
	writeFile(t, dir, "docs/features/frd-side.md", frd("frd-side", "FEAT-SID-001",
		scenario{"AC-SID-001.1", "only the side branch has this"}))
	commitAll(t, dir, "side only")
	gitIn(t, dir, "checkout", "-q", "main")

	reworded := change{"AC-CHK-001.1", "changed", checkoutPath}
	for _, revision := range []string{"side", "v1", first} {
		t.Run(revision, func(t *testing.T) {
			assertAnswer(t, bin, dir, wantYAML(revision, first, reworded), "--since", revision)
		})
	}
	t.Run("default branch", func(t *testing.T) {
		assertAnswer(t, bin, dir, wantYAML("main", second))
	})
}

// A spec root the base does not hold yet reads as empty at the base, so
// every criterion in it is added.
//
// AC: AC-GHK-008.3
func TestChangedListsEveryCriterionAsAddedWhereTheSpecRootIsNew(t *testing.T) {
	bin := sharedSpec(t)
	dir := t.TempDir()
	gitIn(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "README.md", "no specs yet\n")
	base := commitAll(t, dir, "before the spec root")

	gitIn(t, dir, "checkout", "-q", "-b", "work")
	writeFile(t, dir, "specs.toml", specsTOML)
	writeFile(t, dir, checkoutPath, frd("frd-checkout", "FEAT-CHK-001",
		scenario{"AC-CHK-001.1", "the order is placed"},
		scenario{"AC-CHK-001.2", "payment is refused"}))
	commitAll(t, dir, "add the spec root")

	assertAnswer(t, bin, dir, wantYAML("main", base,
		change{"AC-CHK-001.1", "added", checkoutPath},
		change{"AC-CHK-001.2", "added", checkoutPath}))
}

// The listing exits 0 whatever it lists, and 1 only for a non-empty list
// under --exit-code.
//
// AC: AC-GHK-008.8
func TestChangedFailsOnlyWhenAskedTo(t *testing.T) {
	bin := sharedSpec(t)
	dir, _ := newSpecRepo(t, scenario{"AC-CHK-001.1", "the order is placed"})
	empty := t.TempDir()
	gitIn(t, empty, "init", "-q", "-b", "main")
	writeFile(t, empty, "specs.toml", specsTOML)
	writeFile(t, empty, checkoutPath, frd("frd-checkout", "FEAT-CHK-001",
		scenario{"AC-CHK-001.1", "the order is placed"}))
	commitAll(t, empty, "base")

	writeFile(t, dir, checkoutPath, frd("frd-checkout", "FEAT-CHK-001",
		scenario{"AC-CHK-001.1", "the order is placed and confirmed"}))

	cases := []struct {
		name   string
		dir    string
		args   []string
		status int
	}{
		{"empty without --exit-code", empty, nil, 0},
		{"non-empty without --exit-code", dir, nil, 0},
		{"empty with --exit-code", empty, []string{"--exit-code"}, 0},
		{"non-empty with --exit-code", dir, []string{"--exit-code"}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"criterion", "--changed"}, tc.args...)
			stdout, stderr, code := runAll(t, bin, tc.dir, args...)
			if code != tc.status {
				t.Errorf("exit status = %d, want %d\n%s", code, tc.status, stderr)
			}
			if !strings.HasPrefix(stdout, "base:\n") {
				t.Errorf("stdout = %q, want the answer whatever the status", stdout)
			}
		})
	}
}

// assertRefused runs args in dir and holds the run to status, an empty
// standard output and a standard error beginning with prefix.
func assertRefused(t *testing.T, bin, dir string, status int, prefix string, args ...string) {
	t.Helper()
	stdout, stderr, code := runAll(t, bin, dir, args...)
	if code != status {
		t.Errorf("%v exited %d, want %d\n--- stderr ---\n%s", args, code, status, stderr)
	}
	if stdout != "" {
		t.Errorf("%v wrote %q to standard output, want nothing", args, stdout)
	}
	if !strings.HasPrefix(stderr, prefix) {
		t.Errorf("%v standard error = %q, want it to begin with %q", args, stderr, prefix)
	}
}

// A comparison that cannot be made answers nothing and exits with its
// class: 3 for the repository's conditions, 2 for an unknown revision.
//
// AC: AC-GHK-008.9
func TestChangedThatCannotCompareAnswersNothing(t *testing.T) {
	bin := sharedSpec(t)

	t.Run("outside a git repository", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
		writeFile(t, dir, "specs.toml", specsTOML)
		writeFile(t, dir, checkoutPath, frd("frd-checkout", "FEAT-CHK-001",
			scenario{"AC-CHK-001.1", "the order is placed"}))
		assertRefused(t, bin, dir, 3, dir, "criterion", "--changed")
	})

	t.Run("no default branch", func(t *testing.T) {
		dir, _ := newSpecRepo(t, scenario{"AC-CHK-001.1", "the order is placed"})
		gitIn(t, dir, "branch", "-q", "-m", "main", "trunk")
		assertRefused(t, bin, dir, 3, "no default branch", "criterion", "--changed")
	})

	t.Run("no merge-base", func(t *testing.T) {
		dir, _ := newSpecRepo(t, scenario{"AC-CHK-001.1", "the order is placed"})
		gitIn(t, dir, "checkout", "-q", "--orphan", "island")
		commitAll(t, dir, "unrelated history")
		assertRefused(t, bin, dir, 3, "HEAD and main", "criterion", "--changed")
	})

	t.Run("unknown revision", func(t *testing.T) {
		dir, _ := newSpecRepo(t, scenario{"AC-CHK-001.1", "the order is placed"})
		assertRefused(t, bin, dir, 2, "no-such-branch", "criterion", "--changed", "--since", "no-such-branch")
	})
}

// --since and --exit-code mean nothing without --changed, and are refused
// as invocation errors.
//
// AC: AC-GHK-008.12
func TestComparisonOptionsWithoutChangedAreInvocationErrors(t *testing.T) {
	bin := sharedSpec(t)
	dir, _ := newSpecRepo(t, scenario{"AC-CHK-001.1", "the order is placed"})
	assertRefused(t, bin, dir, 2, "--since", "criterion", "--since", "main")
	assertRefused(t, bin, dir, 2, "--since", "criterion", "AC-CHK-001.1", "--since", "main")
	assertRefused(t, bin, dir, 2, "--exit-code", "criterion", "--exit-code")
}

// A spec root outside the work tree cannot be read from its history, and
// is reported as an environment error naming both.
//
// AC: AC-GHK-008.13
func TestChangedRefusesASpecRootOutsideTheWorkTree(t *testing.T) {
	bin := sharedSpec(t)
	parent := t.TempDir()
	dir := filepath.Join(parent, "repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "init", "-q", "-b", "main")
	writeFile(t, dir, "specs.toml", strings.Replace(specsTOML, `root = "docs"`, `root = "../outside"`, 1))
	writeFile(t, parent, "outside/features/frd-checkout.md", frd("frd-checkout", "FEAT-CHK-001",
		scenario{"AC-CHK-001.1", "the order is placed"}))
	commitAll(t, dir, "base")

	spec := filepath.Join(parent, "outside")
	stdout, stderr, code := runAll(t, bin, dir, "criterion", "--changed")
	if code != 3 {
		t.Errorf("exit status = %d, want 3\n%s", code, stderr)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want nothing", stdout)
	}
	if !strings.HasPrefix(stderr, spec) || !strings.Contains(stderr, "outside the git work tree") ||
		!strings.Contains(stderr, "repo") {
		t.Errorf("stderr = %q, want it to begin with the spec root %s and name the work tree", stderr, spec)
	}
}

// With origin/HEAD unset and a local main behind origin/main, the base is
// origin/main, one warning line goes to stderr, and stdout is the answer a
// repository without the stale branch gives. No warning where the local
// branch is current or --since names the revision.
//
// AC: AC-GHK-008.10, AC-GHK-008.14, AC-GHK-008.15
func TestChangedPrefersOriginMainOverAStaleLocalMain(t *testing.T) {
	bin := sharedSpec(t)
	dir, first := newSpecRepo(t, scenario{"AC-CHK-001.1", "the order is placed"})
	writeFile(t, dir, checkoutPath, frd("frd-checkout", "FEAT-CHK-001",
		scenario{"AC-CHK-001.1", "the order is placed and confirmed"}))
	second := commitAll(t, dir, "reword")
	gitIn(t, dir, "update-ref", "refs/remotes/origin/main", second)
	gitIn(t, dir, "checkout", "-q", "-b", "work")
	want := wantYAML("origin/main", second)

	stdout, stderr, code := runAll(t, bin, dir, "criterion", "--changed")
	if code != 0 || stdout != want {
		t.Fatalf("exit %d, stdout\n%s\nwant\n%s\nstderr %s", code, stdout, want, stderr)
	}
	wantWarning := fmt.Sprintf("warning: local main (%s) is behind origin/main (%s); comparing against origin/main\n",
		first[:7], second[:7])
	// main is at second here; move it back to make it stale.
	gitIn(t, dir, "update-ref", "refs/heads/main", first)
	stdout, stderr, code = runAll(t, bin, dir, "criterion", "--changed")
	if code != 0 || stdout != want {
		t.Fatalf("stale: exit %d, stdout\n%s\nwant\n%s", code, stdout, want)
	}
	if stderr != wantWarning {
		t.Errorf("stderr = %q, want %q", stderr, wantWarning)
	}

	_, stderr, _ = runAll(t, bin, dir, "criterion", "--changed", "--since", "origin/main")
	if stderr != "" {
		t.Errorf("--since stderr = %q, want none", stderr)
	}
	gitIn(t, dir, "update-ref", "refs/heads/main", second)
	if _, stderr, _ = runAll(t, bin, dir, "criterion", "--changed"); stderr != "" {
		t.Errorf("current local main: stderr = %q, want none", stderr)
	}
}

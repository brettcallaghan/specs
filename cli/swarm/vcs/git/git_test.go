// Tests for the invocations this implementation declares and for the
// decoders that read them.
//
// The bytes the decoders are held to are the ones measured on git 2.43.0
// — in this repository, and in a scratch repository for the case that
// needs a deliberately broken worktree registration. They are recorded
// here so a git release that changes a shape fails a test rather than
// producing a plausible wrong answer.
package git

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/vcs"
)

func result(stdout, stderr string, status int) swarmexec.Result {
	return swarmexec.Result{Stdout: []byte(stdout), Stderr: []byte(stderr), ExitStatus: status}
}

func argvs(t *testing.T, steps vcs.Steps) []string {
	t.Helper()

	out := make([]string, 0, len(steps))
	for _, s := range steps {
		parts := []string{s.Executable}
		for _, a := range s.Args {
			if a.IsRef() {
				t.Fatalf("step %q takes an argument from %q; this implementation builds only literals", s.Name, a.Step())
			}
			parts = append(parts, a.Literal())
		}
		out = append(out, strings.Join(parts, " "))
	}
	return out
}

func TestTheInvocationsItDeclares(t *testing.T) {
	v := VersionControl{}

	cases := []struct {
		name  string
		steps vcs.Steps
		want  []string
	}{
		// Prune first, then list: a registration whose directory is gone
		// is still listed, and reads as a live worktree.
		{"the standing worktrees", v.Worktrees("wt"), []string{
			"git worktree prune",
			"git worktree list --porcelain",
		}},

		// A requested format with NUL separators, never a rendered log.
		{"recent commits", v.Commits("log", "HEAD", 5), []string{
			"git log --format=%H%x00%ct%x00%s -n 5 HEAD",
		}},
		{"every commit", v.Commits("log", "main", 0), []string{
			"git log --format=%H%x00%ct%x00%s main",
		}},

		{"the local branches", v.Branches("br"), []string{
			"git for-each-ref --format=%(refname:short)%00%(objectname)%00%(upstream:short) refs/heads/",
		}},

		{"ancestry", v.Descends("anc", "main", "HEAD"), []string{
			"git merge-base --is-ancestor main HEAD",
		}},

		// --verify and ^{commit} so a ref that is not a commit is a
		// failure rather than an answer.
		{"resolve", v.Resolve("rev", "main"), []string{
			"git rev-parse --verify main^{commit}",
		}},

		{"divergence", v.Divergence("div", "dev-1"), []string{
			"git rev-list --count --left-right dev-1...dev-1@{upstream}",
		}},

		{"merge", v.Merge("m", "dev-1", "merge dev-1"), []string{
			"git merge --no-ff -m merge dev-1 dev-1",
			"git log --format=%H%x00%ct%x00%s -n 1 HEAD",
		}},
	}

	for _, c := range cases {
		got := argvs(t, c.steps)
		if strings.Join(got, " | ") != strings.Join(c.want, " | ") {
			t.Errorf("%s:\n  got  %v\n  want %v", c.name, got, c.want)
		}
	}
}

// A fragment's answer is its last step, which is what lets a caller read
// one value out of a plan whose shape it did not have to know.
func TestAFragmentsAnswerIsItsLastStep(t *testing.T) {
	v := VersionControl{}

	if got := v.Worktrees("wt").Answer(); got != "wt-list" {
		t.Errorf("Worktrees answer = %q, want the list step", got)
	}
	if got := v.Merge("m", "dev-1", "x").Answer(); got != "m-commit" {
		t.Errorf("Merge answer = %q, want the step reading the merge commit", got)
	}
	if got := (vcs.Steps{}).Answer(); got != "" {
		t.Errorf("an empty fragment answers %q", got)
	}
}

// Reads and writes are declared as such, because a write mis-declared as
// a read is one a dry run would perform. The prune and the merge are the
// two writes here, and only one of them is safe to re-run.
func TestEachStepDeclaresItsEffectAndWhetherReRunningIsSafe(t *testing.T) {
	v := VersionControl{}

	worktrees := v.Worktrees("wt")
	if worktrees[0].Effect != plan.Write {
		t.Errorf("the prune is declared %s, want write", worktrees[0].Effect)
	}
	if !worktrees[0].Repeatable {
		t.Error("the prune is declared unsafe to re-run, but pruning twice removes nothing more")
	}
	if worktrees[1].Effect != plan.Read {
		t.Errorf("the list is declared %s, want read", worktrees[1].Effect)
	}

	merge := v.Merge("m", "dev-1", "x")
	if merge[0].Effect != plan.Write {
		t.Errorf("the merge is declared %s, want write", merge[0].Effect)
	}
	if merge[0].Repeatable {
		t.Error("the merge is declared safe to re-run; a second merge of the same branch is not what the caller asked for")
	}
	if merge[0].OnFailure != plan.Stop {
		t.Error("the merge does not stop the plan on failure; a worktree removal chained after a failed merge is the incident this design prevents")
	}

	for _, steps := range []vcs.Steps{
		v.Commits("c", "HEAD", 1), v.Branches("b"),
		v.Descends("a", "x", "y"), v.Resolve("r", "x"), v.Divergence("d", "b"),
	} {
		for _, s := range steps {
			if s.Effect != plan.Read {
				t.Errorf("step %q is declared %s, want read", s.Name, s.Effect)
			}
			if !s.Repeatable {
				t.Errorf("read step %q is declared unsafe to re-run", s.Name)
			}
		}
	}
}

// The measured output of `git worktree list --porcelain`, including the
// stale registration produced in a scratch repository by deleting a
// worktree's directory.
const worktreeList = `worktree /w/main
HEAD f927a8fe0f263541ce8e7e8e594e9c5839166c6d
branch refs/heads/main

worktree /w/dev-1
HEAD f927a8fe0f263541ce8e7e8e594e9c5839166c6d
branch refs/heads/dev-1
locked claude agent dev-1

worktree /w/dev-2
HEAD f927a8fe0f263541ce8e7e8e594e9c5839166c6d
branch refs/heads/dev-2
prunable gitdir file points to non-existent location

worktree /w/detached
HEAD 0000000000000000000000000000000000000000
detached
`

// A stale registration reads as a live worktree to everything above, and
// a run's health is partly a count of standing worktrees — so dropping it
// is the whole point of the operation.
func TestAStaleRegistrationIsNotReportedAsAStandingWorktree(t *testing.T) {
	value, err := decodeWorktreeList(result(worktreeList, "", 0))
	if err != nil {
		t.Fatalf("decodeWorktreeList: %v", err)
	}
	trees, ok := value.([]vcs.Worktree)
	if !ok {
		t.Fatalf("decodeWorktreeList produced %T", value)
	}

	if len(trees) != 3 {
		t.Fatalf("got %d worktrees, want 3 (the prunable one dropped): %+v", len(trees), trees)
	}
	for _, w := range trees {
		if w.Path == "/w/dev-2" {
			t.Error("the prunable registration is reported as a standing worktree")
		}
	}

	// The short branch name, because that is how a dispatch is matched to
	// its worktree.
	if trees[1].Branch != "dev-1" {
		t.Errorf("Branch = %q, want the short name", trees[1].Branch)
	}
	if !trees[1].Locked || trees[1].LockReason != "claude agent dev-1" {
		t.Errorf("the lock was not read: %+v", trees[1])
	}
	if !trees[2].Detached || trees[2].Branch != "" {
		t.Errorf("the detached worktree was not read as detached: %+v", trees[2])
	}
	if trees[0].Head != "f927a8fe0f263541ce8e7e8e594e9c5839166c6d" {
		t.Errorf("Head = %q", trees[0].Head)
	}
}

func TestAnEmptyWorktreeListIsAnEmptySetRatherThanNil(t *testing.T) {
	value, err := decodeWorktreeList(result("", "", 0))
	if err != nil {
		t.Fatalf("decodeWorktreeList: %v", err)
	}
	trees, _ := value.([]vcs.Worktree)
	if trees == nil {
		t.Error("the empty value is nil rather than an empty set")
	}
	if len(trees) != 0 {
		t.Errorf("got %d worktrees, want 0", len(trees))
	}
}

// A commit subject may contain anything — this run's own subjects carry a
// bead id, a priority and a title with punctuation in it — so the format
// is asked for with NUL separators and nothing splits a rendered line.
func TestACommitSubjectSurvivesWhateverItContains(t *testing.T) {
	const subject = "agent-plugins-ih18.6 P1 Version Control: the interface, and git's tabs\tand \x1b[31mescapes"
	body := "b0c10c2a11811936bf774a420aec15b477ac635c\x001787969612\x00" + subject + "\n"

	value, err := decodeCommits(result(body, "", 0))
	if err != nil {
		t.Fatalf("decodeCommits: %v", err)
	}
	commits, _ := value.([]vcs.Commit)
	if len(commits) != 1 {
		t.Fatalf("got %d commits, want 1", len(commits))
	}
	if commits[0].Subject != subject {
		t.Errorf("Subject =\n  %q\nwant\n  %q", commits[0].Subject, subject)
	}
	// UTC, never the committer's zone: the contract forbids a field whose
	// meaning depends on the producing process's time zone.
	want := time.Unix(1787969612, 0).UTC()
	if !commits[0].When.Equal(want) {
		t.Errorf("When = %v, want %v", commits[0].When, want)
	}
	if commits[0].When.Location() != time.UTC {
		t.Errorf("When carries the zone %v, want UTC", commits[0].When.Location())
	}
}

func TestAMalformedLogRecordIsAnErrorRatherThanAPartialCommit(t *testing.T) {
	for name, body := range map[string]string{
		"too few fields": "sha\x001787969612\n",
		"no separators":  "sha 1787969612 subject\n",
		"a bad time":     "sha\x00not-a-number\x00subject\n",
	} {
		if _, err := decodeCommits(result(body, "", 0)); err == nil {
			t.Errorf("%s: decodeCommits accepted it", name)
		}
	}
}

// Ancestry's answer is its exit status. Both answers, and the status that
// is not one.
func TestAncestryReadsItsAnswerFromTheExitStatus(t *testing.T) {
	yes, err := decodeAncestry(result("", "", 0))
	if err != nil || yes != true {
		t.Errorf("exit 0 -> %v, %v; want true", yes, err)
	}
	no, err := decodeAncestry(result("", "", 1))
	if err != nil || no != false {
		t.Errorf("exit 1 -> %v, %v; want false", no, err)
	}
	// 128 is not an answer: a revision that does not resolve reported as
	// "not an ancestor" would let a branch cut from the wrong base pass
	// its own guard.
	if _, err := decodeAncestry(result("", "fatal: Not a valid object name deadbeef\n", 128)); err == nil {
		t.Error("exit 128 was read as an answer")
	}

	// And the step declares exactly that, so the interpreter agrees with
	// the decoder.
	step := VersionControl{}.Descends("a", "x", "y")[0]
	if len(step.Answers) != 1 || step.Answers[0] != 1 {
		t.Errorf("the step declares Answers %v, want [1]", step.Answers)
	}
}

// Measured: `git rev-parse no-such-branch` exits 128 AND echoes the
// argument on stdout. A decoder reading stdout before the status returns
// the branch name as if it were a commit id.
func TestRevParseEchoingItsArgumentIsNotMistakenForACommit(t *testing.T) {
	const stderr = "fatal: ambiguous argument 'no-such-branch': unknown revision or path not in the working tree.\n"

	_, err := decodeRevision(result("no-such-branch\n", stderr, 128))
	if err == nil {
		t.Fatal("decodeRevision accepted the echoed argument as a commit id")
	}
	if !strings.Contains(err.Error(), "ambiguous argument") {
		t.Errorf("the error does not carry what git said: %v", err)
	}

	value, err := decodeRevision(result("b0c10c2a11811936bf774a420aec15b477ac635c\n", "", 0))
	if err != nil {
		t.Fatalf("decodeRevision: %v", err)
	}
	if value != "b0c10c2a11811936bf774a420aec15b477ac635c" {
		t.Errorf("value = %v", value)
	}
}

func TestDivergenceReadsTwoCounts(t *testing.T) {
	value, err := decodeDivergence(result("3\t7\n", "", 0))
	if err != nil {
		t.Fatalf("decodeDivergence: %v", err)
	}
	if got := value.(vcs.Divergence); got.Ahead != 3 || got.Behind != 7 {
		t.Errorf("got %+v, want ahead 3 behind 7", got)
	}

	// Measured: a branch with no upstream makes git exit 128. Reporting
	// that as "in step with its remote" would be the silent wrong answer.
	const stderr = "fatal: no upstream configured for branch 'worktree-agent-ae2cae'\n"
	if _, err := decodeDivergence(result("", stderr, 128)); err == nil {
		t.Error("a branch with no upstream was read as in step with its remote")
	}
	if _, err := decodeDivergence(result("3\n", "", 0)); err == nil {
		t.Error("one count was read as two")
	}
}

// TestOnlyThisPackageNamesTheTool enforces "a component other than the
// Git Version Control SHALL NOT invoke git", the same way the Beads
// Tracker's own test enforces its half.
func TestOnlyThisPackageNamesTheTool(t *testing.T) {
	root := filepath.Join("..", "..")
	thisPackage := filepath.Join("..", "..", "vcs", "git")

	fset := token.NewFileSet()
	var scanned, hereToo int

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if strings.HasPrefix(filepath.Clean(path), filepath.Clean(thisPackage)+string(filepath.Separator)) {
			hereToo++
			return nil
		}
		scanned++

		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			t.Fatalf("parsing %s: %v", path, parseErr)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			if value, err := strconv.Unquote(lit.Value); err == nil && value == executable {
				t.Errorf("%s names %q; only the Git Version Control may invoke it", path, executable)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walking the swarm tree: %v", err)
	}

	if scanned == 0 {
		t.Fatalf("scanned no source outside this package under %s", root)
	}
	if hereToo == 0 {
		t.Fatal("the walk did not reach this package, so its exemption is untested")
	}
}

// A caller may point the implementation at a particular git and a
// particular worktree.
func TestTheExecutableAndDirectoryAreConfigurable(t *testing.T) {
	v := VersionControl{Git: "/opt/bin/git", Dir: "/w/dev-1"}
	s := v.Commits("c", "HEAD", 1)[0]

	if s.Executable != "/opt/bin/git" {
		t.Errorf("Executable = %q", s.Executable)
	}
	if s.Dir != "/w/dev-1" {
		t.Errorf("Dir = %q", s.Dir)
	}
}

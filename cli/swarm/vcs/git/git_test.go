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
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
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

		{"the repository root", v.Root("root"), []string{
			"git rev-parse --show-toplevel",
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

		// The dispatch loop's writes.
		{"create a worktree", v.CreateWorktree("wt", "/w/dev-1", "dev-1", "sprint"), []string{
			"git worktree add -b dev-1 /w/dev-1 sprint",
		}},
		// No flag at all, which is the point: see
		// TestTheWorktreeRemovalHasNoForcingForm.
		{"remove a worktree", v.RemoveWorktree("wt", "/w/dev-1"), []string{
			"git worktree remove /w/dev-1",
		}},
		{"create a branch", v.CreateBranch("br", "dev-1", "sprint"), []string{
			"git branch dev-1 sprint",
		}},
		// -d, never -D: a branch carrying unmerged commits is one whose
		// work has not landed.
		{"delete a branch", v.DeleteBranch("br", "dev-1"), []string{
			"git branch -d dev-1",
		}},
		{"rebase", v.Rebase("rb", "dev-1", "sprint"), []string{
			"git rebase sprint dev-1",
		}},
		{"fast-forward", v.MergeFastForward("ff", "dev-1"), []string{
			"git merge --ff-only dev-1",
		}},
		{"fetch", v.Fetch("f", "origin"), []string{
			"git fetch origin",
		}},
		{"push", v.Push("p", "origin", "sprint"), []string{
			"git push origin sprint",
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
		v.Root("root"),
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

// Every one of the loop's writes declares whether re-running it is safe,
// and each declaration is a measured claim rather than a guess (ADR-026:
// recovery from a partial plan is re-running it). The measurement behind
// each is recorded on the operation in git.go.
func TestEveryWriteDeclaresWhetherReRunningItIsSafe(t *testing.T) {
	v := VersionControl{}

	cases := []struct {
		what       string
		steps      vcs.Steps
		repeatable bool
		why        string
	}{
		{"create a worktree", v.CreateWorktree("wt", "/w/d", "d", "s"), false,
			"a second add of the same branch exits 255, `a branch named 'd' already exists`"},
		{"remove a worktree", v.RemoveWorktree("wt", "/w/d"), false,
			"removing it again exits 128, `is not a working tree`"},
		{"create a branch", v.CreateBranch("br", "d", "s"), false,
			"a second create exits 128, `already exists`"},
		{"delete a branch", v.DeleteBranch("br", "d"), false,
			"a second delete exits 1, `branch 'd' not found`"},
		{"rebase", v.Rebase("rb", "d", "s"), false,
			"a conflicting rebase leaves the repository mid-rebase, where re-running does not start over"},
		{"fast-forward", v.MergeFastForward("ff", "d"), true,
			"a second --ff-only exits 0 with `Already up to date.`, and a refused one leaves nothing behind"},
		{"fetch", v.Fetch("f", "origin"), true,
			"two consecutive fetches both exit 0"},
		{"push", v.Push("p", "origin", "s"), true,
			"a second push with nothing new exits 0 with `Everything up-to-date`"},
	}

	for _, c := range cases {
		if len(c.steps) != 1 {
			t.Errorf("%s is %d steps, want one invocation", c.what, len(c.steps))
			continue
		}
		s := c.steps[0]
		if s.Effect != plan.Write {
			t.Errorf("%s is declared %s, want write; a dry run would perform it", c.what, s.Effect)
		}
		if s.Repeatable != c.repeatable {
			t.Errorf("%s declares Repeatable=%v, want %v: %s", c.what, s.Repeatable, c.repeatable, c.why)
		}
		// None of these carries its answer in its exit status, so none
		// may declare a status other than 0 as an answer — a rejected
		// push read as an answer is a push that silently did not happen.
		if len(s.Answers) != 0 {
			t.Errorf("%s declares Answers %v; its answer is not carried by its exit status", c.what, s.Answers)
		}
		if s.OnFailure != plan.Stop {
			t.Errorf("%s does not stop the plan on failure", c.what)
		}
	}
}

// "The worktree-removal operation SHALL NOT accept a force argument, so
// no caller can reach the forcing form." That is a shape obligation, not
// a runtime check, so it is tested as a shape: the interface method takes
// two strings and nothing else, the invocation carries no flag, and no
// source in this package names the forcing option at all.
func TestTheWorktreeRemovalHasNoForcingForm(t *testing.T) {
	method, ok := reflect.TypeOf((*vcs.VersionControl)(nil)).Elem().MethodByName("RemoveWorktree")
	if !ok {
		t.Fatal("the interface declares no RemoveWorktree")
	}
	signature := method.Type
	if got := signature.NumIn(); got != 2 {
		t.Errorf("RemoveWorktree takes %d arguments, want exactly two (the step prefix and the path); anything more is where a force flag would live", got)
	}
	for i := 0; i < signature.NumIn(); i++ {
		if signature.In(i).Kind() != reflect.String {
			t.Errorf("RemoveWorktree's argument %d is %s, want string; a bool here is a forcing form", i, signature.In(i).Kind())
		}
	}

	steps := VersionControl{}.RemoveWorktree("wt", "/w/dev-1")
	for _, s := range steps {
		for _, a := range s.Args {
			if strings.HasPrefix(a.Literal(), "-") {
				t.Errorf("the removal invocation carries the option %q; it is meant to carry none", a.Literal())
			}
		}
	}

	// And the option is absent from the package outright, so it cannot be
	// reintroduced by a helper the shape tests above do not reach.
	forbidden := map[string]string{
		"--force": "the forcing form of a removal or a push",
		"-f":      "the short forcing form",
		"-D":      "the forcing form of a branch deletion",
	}
	forEachStringLiteral(t, ".", func(path, value string) {
		if why, bad := forbidden[value]; bad {
			t.Errorf("%s names %q — %s; a run's worktree is where an unfinished agent's work lives", path, value, why)
		}
	})
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

// forEachStringLiteral calls visit for every string literal in every
// non-test Go source file under root. Test files are skipped: a test
// naming a forbidden option in order to forbid it is not a use of it.
func forEachStringLiteral(t *testing.T, root string, visit func(path, value string)) {
	t.Helper()

	fset := token.NewFileSet()
	var scanned int
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
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
			if value, err := strconv.Unquote(lit.Value); err == nil {
				visit(path, value)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	if scanned == 0 {
		t.Fatalf("scanned no source under %s, so the scan proves nothing", root)
	}
}

// A conflicting merge is decoded into the paths that conflict, not handed
// back as git's text for a caller to match on.
//
// The bytes are measured on git 2.43.0: the CONFLICT lines arrive on
// STDOUT and stderr is EMPTY, which is the trap a decoder reading stderr
// for the reason falls into.
func TestAConflictingMergeIsDecodedIntoThePathsThatConflict(t *testing.T) {
	const stdout = `Auto-merging added.txt
CONFLICT (add/add): Merge conflict in added.txt
CONFLICT (modify/delete): del.txt deleted in dev and modified in HEAD.  Version HEAD of del.txt left in tree.
Auto-merging sp ace.txt
CONFLICT (add/add): Merge conflict in sp ace.txt
Automatic merge failed; fix conflicts and then commit the result.
`

	_, err := decodeMerge(result(stdout, "", 1))
	if err == nil {
		t.Fatal("decodeMerge accepted a conflicting merge")
	}

	var refused *vcs.Refusal
	if !errors.As(err, &refused) {
		t.Fatalf("the error is %T, not a refusal a caller can branch on: %v", err, err)
	}
	if refused.Operation != "merge" {
		t.Errorf("Operation = %q, want merge", refused.Operation)
	}
	// A path is printed unquoted, so one containing a space is one path
	// and not two.
	want := []string{"added.txt", "del.txt", "sp ace.txt"}
	if strings.Join(refused.Conflicts, "|") != strings.Join(want, "|") {
		t.Errorf("Conflicts = %q, want %q", refused.Conflicts, want)
	}
	if refused.NonFastForward {
		t.Error("a conflict was reported as a non-fast-forward")
	}
}

// A refused fast-forward is the other refusal the run branches on: the
// dispatch was not rebased onto the current target. Measured — exit 128,
// the reason on stderr behind several hint lines.
func TestARefusedFastForwardIsDecodedAsSuch(t *testing.T) {
	const stderr = `hint: Diverging branches can't be fast-forwarded, you need to either:
hint:
hint: 	git merge --no-ff
hint:
hint: or:
hint:
hint: 	git rebase
hint:
hint: Disable this message with "git config advice.diverging false"
fatal: Not possible to fast-forward, aborting.
`

	_, err := decodeFastForward(result("", stderr, 128))
	if err == nil {
		t.Fatal("decodeFastForward accepted a refused fast-forward")
	}

	var refused *vcs.Refusal
	if !errors.As(err, &refused) {
		t.Fatalf("the error is %T, not a refusal a caller can branch on: %v", err, err)
	}
	if !refused.NonFastForward {
		t.Error("NonFastForward is false; the caller cannot tell this from a conflict")
	}
	if len(refused.Conflicts) != 0 {
		t.Errorf("Conflicts = %q, want none: nothing was merged", refused.Conflicts)
	}
	if !strings.Contains(refused.Error(), "fast forward") {
		t.Errorf("the message does not say what was refused: %v", refused)
	}

	// And a fast-forward that succeeds is not a refusal.
	if _, err := decodeFastForward(result("Already up to date.\n", "", 0)); err != nil {
		t.Errorf("an up-to-date fast-forward was read as a failure: %v", err)
	}
}

// A conflicting rebase is decoded the same way, from the same stream —
// and the run needs to know it is a rebase rather than a merge, because
// the repository is left mid-rebase.
func TestAConflictingRebaseIsDecodedIntoThePathsThatConflict(t *testing.T) {
	const stdout = `Auto-merging f.txt
CONFLICT (content): Merge conflict in f.txt
Auto-merging g.txt
CONFLICT (content): Merge conflict in g.txt
`
	const stderr = `Rebasing (1/1)error: could not apply 8a198f8... dev
hint: Resolve all conflicts manually, mark them as resolved with
hint: "git add/rm <conflicted_files>", then run "git rebase --continue".
Could not apply 8a198f8... dev
`

	_, err := decodeRebase(result(stdout, stderr, 1))
	var refused *vcs.Refusal
	if !errors.As(err, &refused) {
		t.Fatalf("the error is %T, not a refusal: %v", err, err)
	}
	if refused.Operation != "rebase" {
		t.Errorf("Operation = %q, want rebase; the caller has to know the repository is mid-rebase", refused.Operation)
	}
	want := []string{"f.txt", "g.txt"}
	if strings.Join(refused.Conflicts, "|") != strings.Join(want, "|") {
		t.Errorf("Conflicts = %q, want %q", refused.Conflicts, want)
	}
}

// A conflict form whose path this decoder cannot pull out still reaches
// the caller: the line is kept verbatim, so a conflict is never silently
// absent from the refusal.
func TestAConflictLineIsKeptEvenWhenItsPathCannotBeParsed(t *testing.T) {
	// Measured on git 2.43.0 from a rename/rename merge.
	const measured = "CONFLICT (rename/rename): a.txt renamed to c.txt in HEAD and to b.txt in dev.\n"

	_, err := decodeMerge(result(measured, "", 1))
	var refused *vcs.Refusal
	if !errors.As(err, &refused) {
		t.Fatalf("the error is %T, not a refusal: %v", err, err)
	}
	if len(refused.Conflicts) != 1 || refused.Conflicts[0] != "a.txt" {
		t.Errorf("Conflicts = %q, want the renamed path", refused.Conflicts)
	}

	// And a form this decoder does not know — illustrative, not a
	// measured git output — is still visible rather than dropped.
	const unknown = "CONFLICT (submodule): something this decoder has no rule for\n"
	_, err = decodeMerge(result(unknown, "", 1))
	if !errors.As(err, &refused) {
		t.Fatalf("an unparsed conflict form produced %T rather than a refusal: %v", err, err)
	}
	if len(refused.Conflicts) != 0 {
		t.Errorf("Conflicts = %q; a path was invented for a form with no rule", refused.Conflicts)
	}
	if !strings.Contains(refused.Detail, "no rule for") {
		t.Errorf("Detail = %q; the conflict line was dropped", refused.Detail)
	}
}

// A merge that could not run at all is an ordinary failure, not a
// refusal: "git could not do this" and "git declined this" are two
// different answers and a caller branches on them differently.
func TestAFailureThatIsNotARefusalStaysAnOrdinaryFailure(t *testing.T) {
	const stderr = "merge: no-such-branch - not something we can merge\n"

	_, err := decodeMerge(result("", stderr, 1))
	if err == nil {
		t.Fatal("decodeMerge accepted a merge that did not run")
	}
	var refused *vcs.Refusal
	if errors.As(err, &refused) {
		t.Errorf("an unmergeable ref was reported as a refusal: %+v", refused)
	}
	if !strings.Contains(err.Error(), "not something we can merge") {
		t.Errorf("the error does not carry what git said: %v", err)
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

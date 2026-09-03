// Unit tests for `swarm dispatch accept`: the case named before the
// refusal is read, a serialized refusal that stops without a retry and
// without the fold ever being relaxed, and a conflict report naming every
// conflicting path. Every scenario runs through the Beads Tracker's and
// the Git Version Control's real argv-building against the fake process
// runner, so no test starts `bd` or `git`.
//
// AC-SWARM-DISP-004.3, .4 and .8 are declared e2e and live in
// cli/swarm/e2e against the dispatch-stale-base, dispatch-rebase-conflict
// and core fixtures.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	manifestbeads "github.com/brettcallaghan/agent-plugins/cli/swarm/manifest/beads"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
	beadstracker "github.com/brettcallaghan/agent-plugins/cli/swarm/tracker/beads"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/vcs"
	gitvcs "github.com/brettcallaghan/agent-plugins/cli/swarm/vcs/git"
)

const (
	acceptRun      = "run-1"
	acceptAgent    = "dev-1"
	acceptBranch   = "bb/dev-1"
	acceptTargetBr = "sprint/run-1"
	acceptTargetSH = "1111111111111111111111111111111111111111"
	acceptBranchSH = "2222222222222222222222222222222222222222"
)

// runAccept drives the real command tree against the fake and returns
// stdout, stderr and the exit status.
func runAccept(t *testing.T, fake *exectest.Fake, args ...string) (string, string, int) {
	t.Helper()

	tr := beadstracker.Tracker{}
	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut}
	root := newRootCmd(streams)
	root.AddCommand(dispatchCmd(streams, manifestbeads.Store{Tracker: tr}, tr, gitvcs.VersionControl{}, fake))

	code := dispatch(root, streams, args, &out, &errOut)
	return out.String(), errOut.String(), code
}

// drainedQueue is the four tracker reads that answer "dev-1 holds
// nothing": the run's two walk levels and the two halves of its queue.
//
// held, when non-empty, is replayed as the in-progress half instead, so
// the same helper builds the not-drained world.
func drainedQueue(held ...string) []exectest.Recording {
	rows := make([]string, len(held))
	for i, id := range held {
		rows[i] = fmt.Sprintf(`{"id":%q,"title":%q,"status":"in_progress","issue_type":"task","assignee":%q}`,
			id, id, acceptAgent)
	}
	inProgress := "[" + strings.Join(rows, ",") + "]"

	// The walk has to carry every held bead, since the held half is
	// narrowed to the run's own descendants.
	walk := make([]string, 0, len(held)+1)
	walk = append(walk, fmt.Sprintf(`{"id":"run-1.1","title":"a","status":"closed","issue_type":"task","assignee":%q}`, acceptAgent))
	for _, id := range held {
		walk = append(walk, fmt.Sprintf(`{"id":%q,"title":%q,"status":"in_progress","issue_type":"task","assignee":%q}`,
			id, id, acceptAgent))
	}

	recs := []exectest.Recording{
		{
			Argv:   []string{"bd", "list", "--parent", acceptRun, "--all", "--limit", "0", "--json"},
			Stdout: []byte("[" + strings.Join(walk, ",") + "]"),
		},
		{
			Argv:   []string{"bd", "list", "--parent", "run-1.1", "--all", "--limit", "0", "--json"},
			Stdout: []byte("[]"),
		},
		{
			Argv: []string{"bd", "ready", "--parent", acceptRun, "--assignee", acceptAgent,
				"--exclude-type", "epic", "--exclude-label", "human", "--limit", "0", "--json"},
			Stdout: []byte("[]"),
		},
		{
			// The held half carries the same exclusions as the ready half
			// above: one filter builds both, so a bead one drops and the
			// other returns cannot leave a dispatch undrained.
			Argv: []string{"bd", "list", "--assignee", acceptAgent, "--status", "in_progress",
				"--exclude-type", "epic", "--exclude-label", "human", "--limit", "0", "--json"},
			Stdout: []byte(inProgress),
		},
	}
	for _, id := range held {
		recs = append(recs, exectest.Recording{
			Argv:   []string{"bd", "list", "--parent", id, "--all", "--limit", "0", "--json"},
			Stdout: []byte("[]"),
		})
	}
	return recs
}

// acceptManifest records the manifest read that names the target.
func acceptManifest() exectest.Recording {
	return exectest.Recording{
		Argv: []string{"bd", "show", acceptRun, "--json"},
		Stdout: []byte(`[{"id":"` + acceptRun + `","status":"open","issue_type":"epic","metadata":{"branch":"` +
			acceptTargetBr + `"}}]`),
	}
}

// caseReads records the four reads the case comparison makes.
// ancestorStatus is `git merge-base --is-ancestor`'s own answer: 0 means
// the target IS an ancestor of the dispatch branch, which is the
// serialized case.
func caseReads(ancestorStatus int) []exectest.Recording {
	return []exectest.Recording{
		{
			Argv:   []string{"git", "rev-parse", "--verify", acceptTargetBr + "^{commit}"},
			Stdout: []byte(acceptTargetSH + "\n"),
		},
		{
			Argv:   []string{"git", "rev-parse", "--verify", "HEAD^{commit}"},
			Stdout: []byte(acceptTargetSH + "\n"),
		},
		{
			Argv:   []string{"git", "rev-parse", "--verify", acceptBranch + "^{commit}"},
			Stdout: []byte(acceptBranchSH + "\n"),
		},
		{
			Argv:       []string{"git", "merge-base", "--is-ancestor", acceptTargetBr, acceptBranch},
			ExitStatus: ancestorStatus,
		},
	}
}

// settleReads records everything the second half runs once the fold has
// succeeded, in a world where the removal, the deletion and every
// post-merge signal all pass: the ancestry guard, the worktree read, the
// removal and deletion, the re-check's own reads, and nothing else.
//
// worktree, when non-empty, is the path a worktree stands at on the
// dispatch branch; empty means none stands and none is removed. The
// remaining options tune the world each signal reads.
type settleWorld struct {
	// worktree is where the dispatch's tree stands, empty for none.
	worktree string
	// locked marks that tree locked, with the reason git reports.
	locked string
	// removalStatus is what `git worktree remove` exits with.
	removalStatus int
	// survives keeps the tree in the SECOND worktree read, so the
	// post-merge signal sees it still standing.
	survives bool
	// ancestorStatus is `git merge-base --is-ancestor <branch> <target>`'s
	// own answer: 0 yes, 1 no.
	ancestorStatus int
	// closedOn are the bead ids whose close commits are on the branch.
	closedOn []string
	// findings maps a bead id to the comment text recorded on it.
	findings map[string]string
}

// worktreeList renders `git worktree list --porcelain` for a world.
func (w settleWorld) worktreeList(include bool) []byte {
	out := "worktree /repo\nHEAD " + acceptTargetSH + "\nbranch refs/heads/" + acceptTargetBr + "\n\n"
	if include && w.worktree != "" {
		out += "worktree " + w.worktree + "\nHEAD " + acceptBranchSH +
			"\nbranch refs/heads/" + acceptBranch + "\n"
		if w.locked != "" {
			out += "locked " + w.locked + "\n"
		}
		out += "\n"
	}
	return []byte(out)
}

func settleReads(w settleWorld) []exectest.Recording {
	recs := []exectest.Recording{
		{
			Argv:       []string{"git", "merge-base", "--is-ancestor", acceptBranch, acceptTargetBr},
			ExitStatus: w.ancestorStatus,
		},
		{Argv: []string{"git", "worktree", "prune"}},
		{Argv: []string{"git", "worktree", "list", "--porcelain"}, Stdout: w.worktreeList(true)},
	}
	if w.worktree != "" {
		recs = append(recs, exectest.Recording{
			Argv:       []string{"git", "worktree", "remove", w.worktree},
			ExitStatus: w.removalStatus,
		})
	}
	recs = append(recs, exectest.Recording{Argv: []string{"git", "branch", "-d", acceptBranch}})

	// The re-check's second worktree read. A recording repeats only when
	// no LATER one matches, so this pair is what the second read gets.
	recs = append(recs,
		exectest.Recording{Argv: []string{"git", "worktree", "prune"}},
		exectest.Recording{
			Argv:   []string{"git", "worktree", "list", "--porcelain"},
			Stdout: w.worktreeList(w.survives),
		})

	log := ""
	for _, id := range w.closedOn {
		log += id + "\x0011000000\x00" + id + " P1 a bead of the run\n"
	}
	recs = append(recs, exectest.Recording{
		Argv: []string{"git", "log", "--format=%H%x00%ct%x00%s", "-n",
			fmt.Sprint(acceptanceScan), acceptTargetBr},
		Stdout: []byte(log),
	})

	for _, id := range []string{"run-1.1"} {
		body := "[]"
		if text, ok := w.findings[id]; ok {
			body = fmt.Sprintf(`[{"id":"c1","issue_id":%q,"author":"dev-1","text":%q}]`, id, text)
		}
		recs = append(recs, exectest.Recording{
			Argv:   []string{"bd", "comments", id, "--json"},
			Stdout: []byte(body),
		})
	}
	return recs
}

// cleanSettle is the world in which every post-merge signal passes and no
// worktree ever stood — the fixtures' own shape.
func cleanSettle() settleWorld {
	return settleWorld{
		closedOn: []string{"run-1.1"},
		findings: map[string]string{"run-1.1": "retro:proof measured it"},
	}
}

// acceptAnswer parses the command's default (answer-only) JSON output.
func acceptAnswer(t *testing.T, stdout string) map[string]any {
	t.Helper()
	var answer map[string]any
	if err := json.Unmarshal([]byte(stdout), &answer); err != nil {
		t.Fatalf("the answer does not parse: %v\n%s", err, stdout)
	}
	return answer
}

// An unscoped invocation, and one naming no dispatch, are refused with
// the malformed-invocation status, and nothing is read — asserted from
// the fake's own record.
func TestDispatchAcceptRefusesAnIncompleteInvocation(t *testing.T) {
	for _, args := range [][]string{
		{"dispatch", "accept"},
		{"dispatch", "accept", "--run", ""},
		{"dispatch", "accept", "--run", acceptRun},
		{"dispatch", "accept", "--run", acceptRun, "--dispatch", "  "},
	} {
		fake := exectest.NewFake() // no recordings: anything executed is a test failure
		_, _, code := runAccept(t, fake, args...)
		if code != int(StatusMalformed) {
			t.Errorf("%v exited %d, want %d", args, code, StatusMalformed)
		}
		if calls := fake.Argvs(); len(calls) != 0 {
			t.Errorf("%v read or wrote something on a malformed invocation: %v", args, calls)
		}
	}
}

// AC: AC-SWARM-DISP-004.2
//
// The case is named before the refusal is acted on: a refused
// fast-forward still reports whether the branch was cut from the current
// target or from an earlier one. Asserted on the SERIALIZED reading,
// where the fold is refused — so the fact cannot have come from the fold
// having succeeded.
func TestAcceptNamesTheCaseBeforeItActsOnARefusal(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...) // 0: the target IS an ancestor — serialized
	recs = append(recs, exectest.Recording{
		Argv:       []string{"git", "merge", "--ff-only", acceptBranch},
		Stderr:     []byte("fatal: Not possible to fast-forward, aborting.\n"),
		ExitStatus: 128,
	})

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json", "--full-result")
	if code != int(StatusStop) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusStop, stdout)
	}

	// The case reaches the caller as a fact of its own, which is what
	// lets a reader tell the two readings of the same refusal apart.
	if !strings.Contains(stdout, `"case"`) || !strings.Contains(stdout, caseSerialized) {
		t.Errorf("the refusal does not state which case it was read against:\n%s", stdout)
	}
	if strings.Contains(stdout, caseConcurrent) {
		t.Errorf("the refusal names the concurrent case, but the target is an ancestor of the branch:\n%s", stdout)
	}

	// And it was measured before the fold: the comparison ran first.
	caseIdx, foldIdx := -1, -1
	for i, argv := range fake.Argvs() {
		switch {
		case len(argv) >= 2 && argv[0] == "git" && argv[1] == "merge-base":
			caseIdx = i
		case len(argv) >= 2 && argv[0] == "git" && argv[1] == "merge":
			foldIdx = i
		}
	}
	if caseIdx < 0 || foldIdx < 0 {
		t.Fatalf("the case comparison or the fold never ran: %v", fake.Argvs())
	}
	if caseIdx > foldIdx {
		t.Errorf("the case was read at call %d and the fold attempted at %d; the case comes first", caseIdx, foldIdx)
	}
}

// AC: AC-SWARM-DISP-004.5
//
// A serialized dispatch whose fast-forward is refused stops, and the
// merge is neither re-attempted nor re-attempted with fast-forward-only
// relaxed — asserted from the fake's own record: exactly one `git merge`,
// and it carries --ff-only.
func TestASerializedRefusalStopsWithoutRetryingTheMerge(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{
		Argv:       []string{"git", "merge", "--ff-only", acceptBranch},
		Stderr:     []byte("fatal: Not possible to fast-forward, aborting.\n"),
		ExitStatus: 128,
	})

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json")
	if code != int(StatusStop) {
		t.Fatalf("exit status = %d, want %d (stop)\n%s", code, StatusStop, stdout)
	}

	var merges [][]string
	for _, argv := range fake.Argvs() {
		if len(argv) >= 2 && argv[0] == "git" && argv[1] == "merge" {
			merges = append(merges, argv)
		}
		if len(argv) >= 2 && argv[0] == "git" && argv[1] == "rebase" {
			t.Errorf("a serialized dispatch was rebased: %v", argv)
		}
	}
	if len(merges) != 1 {
		t.Errorf("the merge was attempted %d times, want exactly once: %v", len(merges), merges)
	}
	for _, argv := range merges {
		if !contains(argv, "--ff-only") {
			t.Errorf("a fold ran without --ff-only, which is the relaxed form: %v", argv)
		}
	}
}

// AC: AC-SWARM-DISP-004.7
//
// A conflict report names EVERY conflicting path, not the first — and
// the finding recorded on the unmerged bead carries them too, so a reader
// who has only the tracker sees the same set.
func TestAConflictReportNamesEveryConflictingPath(t *testing.T) {
	const conflict = `Auto-merging alpha.txt
CONFLICT (content): Merge conflict in alpha.txt
Auto-merging beta.txt
CONFLICT (content): Merge conflict in beta.txt
CONFLICT (add/add): Merge conflict in gamma.txt
`
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(1)...) // 1: not an ancestor — concurrent
	recs = append(recs,
		exectest.Recording{
			Argv:       []string{"git", "rebase", acceptTargetSH, acceptBranch},
			Stdout:     []byte(conflict),
			Stderr:     []byte("error: could not apply 2222222... run-1.2 P2 the work\n"),
			ExitStatus: 1,
		},
		exectest.Recording{
			Argv:   []string{"git", "log", "--format=%H%x00%ct%x00%s", "-n", "1", acceptBranch},
			Stdout: []byte(acceptBranchSH + "\x001756000000\x00run-1.2 P2 the work\n"),
		},
		exectest.Recording{
			Argv:   []string{"git", "log", "--format=%H%x00%ct%x00%s", "-n", "1", acceptTargetBr},
			Stdout: []byte(acceptTargetSH + "\x001756000100\x00run-1.3 P2 the other work\n"),
		},
		exectest.Recording{
			Argv:   []string{"bd", "update", "run-1.2", "--status", "blocked", "--json"},
			Stdout: []byte(`[{"id":"run-1.2","status":"blocked","issue_type":"task"}]`),
		},
	)

	fake := exectest.NewFake(recs...)
	// The comment's body is built by the command, so the recording has to
	// match whatever it writes: declare it after the fact by matching on
	// the prefix instead. exectest matches the whole argv, so the body is
	// reproduced here exactly as the command builds it.
	fake.Recordings = append(fake.Recordings, exectest.Recording{
		Argv:   []string{"bd", "comment", "run-1.2", conflictFindingBody(), "--json"},
		Stdout: []byte(`{"id":"c1","issue_id":"run-1.2","author":"swarm","text":"recorded"}`),
	})

	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json")
	if code != int(StatusStop) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusStop, stdout)
	}

	for _, path := range []string{"alpha.txt", "beta.txt", "gamma.txt"} {
		if !strings.Contains(stdout, path) {
			t.Errorf("the report does not name the conflicting path %q:\n%s", path, stdout)
		}
	}

	// The bead was blocked and the finding recorded on it, in that order
	// — the repeatable write before the one that cannot be re-run.
	blockIdx, noteIdx := -1, -1
	var body string
	for i, argv := range fake.Argvs() {
		switch {
		case len(argv) >= 4 && argv[0] == "bd" && argv[1] == "update" && argv[3] == "--status":
			blockIdx = i
		case len(argv) >= 3 && argv[0] == "bd" && argv[1] == "comment":
			noteIdx = i
			body = argv[3]
		}
	}
	if blockIdx < 0 {
		t.Fatalf("the unmerged bead was never blocked: %v", fake.Argvs())
	}
	if noteIdx < 0 {
		t.Fatalf("no sequencing finding was recorded: %v", fake.Argvs())
	}
	if blockIdx > noteIdx {
		t.Errorf("the finding was recorded at call %d and the block at %d; the repeatable write comes first",
			noteIdx, blockIdx)
	}
	if !strings.HasPrefix(body, "retro:sequencing ") {
		t.Errorf("the finding does not use the run's finding convention: %q", body)
	}
	for _, want := range []string{"run-1.2", "run-1.3", "alpha.txt", "beta.txt", "gamma.txt"} {
		if !strings.Contains(body, want) {
			t.Errorf("the finding does not name %q:\n%s", want, body)
		}
	}
}

// conflictRefusal is the [vcs.Refusal] the Git Version Control's own
// decoder builds from the conflict output the test above replays: the
// operation, the paths in the order git reported them, and the CONFLICT
// lines as the detail.
func conflictRefusal() *vcs.Refusal {
	return &vcs.Refusal{
		Operation: "rebase",
		Conflicts: []string{"alpha.txt", "beta.txt", "gamma.txt"},
		Detail: "CONFLICT (content): Merge conflict in alpha.txt\n" +
			"CONFLICT (content): Merge conflict in beta.txt\n" +
			"CONFLICT (add/add): Merge conflict in gamma.txt",
	}
}

// conflictFindingBody rebuilds the body the command writes for that
// conflict, so the fake can declare the exact invocation rather than a
// prefix of it.
func conflictFindingBody() string {
	return sequencingFinding(acceptBranch, acceptTargetBr, "run-1.2", "run-1.3", conflictRefusal())
}

// A concurrent dispatch is rebased onto the target as it stands at merge
// time — the SHA the case comparison read, not the branch name — and is
// then put back on the target before the fold, since the rebase leaves
// the repository on the dispatch branch.
func TestAConcurrentDispatchIsRebasedOntoTheShaTheCaseWasReadAgainst(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(1)...) // concurrent
	recs = append(recs,
		exectest.Recording{Argv: []string{"git", "rebase", acceptTargetSH, acceptBranch}},
		exectest.Recording{Argv: []string{"git", "checkout", acceptTargetBr}},
		exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}},
	)
	recs = append(recs, settleReads(cleanSettle())...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--no-push", "--json")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	rebaseIdx, checkoutIdx, foldIdx := -1, -1, -1
	for i, argv := range fake.Argvs() {
		switch {
		case len(argv) >= 2 && argv[0] == "git" && argv[1] == "rebase":
			rebaseIdx = i
			if argv[2] != acceptTargetSH {
				t.Errorf("the rebase names %q, want the sha the case was read against (%s)", argv[2], acceptTargetSH)
			}
		case len(argv) >= 2 && argv[0] == "git" && argv[1] == "checkout":
			checkoutIdx = i
			if argv[2] != acceptTargetBr {
				t.Errorf("the repository was put on %q, want the target %q", argv[2], acceptTargetBr)
			}
		case len(argv) >= 2 && argv[0] == "git" && argv[1] == "merge":
			foldIdx = i
		}
	}
	if rebaseIdx < 0 || checkoutIdx < 0 || foldIdx < 0 {
		t.Fatalf("rebase@%d checkout@%d fold@%d; all three are required in the concurrent case: %v",
			rebaseIdx, checkoutIdx, foldIdx, fake.Argvs())
	}
	if !(rebaseIdx < checkoutIdx && checkoutIdx < foldIdx) {
		t.Errorf("order was rebase@%d, checkout@%d, fold@%d; the fold must follow the return to the target",
			rebaseIdx, checkoutIdx, foldIdx)
	}

	if answer := acceptAnswer(t, stdout); answer["case"] != caseConcurrent {
		t.Errorf("case = %v, want %q", answer["case"], caseConcurrent)
	}
}

// The case is never inferred from how many worktrees are standing: the
// case reading consults no worktree list at all.
//
// Asserted as an ORDERING rather than as an absence, because the second
// half does read the list — to find the tree it is about to remove and
// to check afterwards that it is gone. What the delta forbids is the
// case being READ from it, so the assertion is that no worktree list is
// consulted before the fold.
func TestTheCaseIsNotInferredFromTheStandingWorktrees(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, settleReads(cleanSettle())...)

	fake := exectest.NewFake(recs...)
	if _, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent,
		"--no-push", "--json"); code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d", code, StatusClear)
	}

	foldIdx := -1
	for i, argv := range fake.Argvs() {
		if len(argv) >= 2 && argv[0] == "git" && argv[1] == "merge" {
			foldIdx = i
		}
	}
	if foldIdx < 0 {
		t.Fatalf("the fold never ran: %v", fake.Argvs())
	}
	for i, argv := range fake.Argvs() {
		if i < foldIdx && len(argv) >= 2 && argv[0] == "git" && argv[1] == "worktree" {
			t.Errorf("the case reading consulted the worktree list at call %d: %v", i, argv)
		}
	}
}

// A dispatch that still holds work is refused with the attention status
// naming what it holds, and nothing is merged — the merge is never even
// reached, so the fake needs no recording for it.
func TestAnUndrainedDispatchIsNotAccepted(t *testing.T) {
	fake := exectest.NewFake(drainedQueue("run-1.2")...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json")
	if code != int(StatusAttention) {
		t.Fatalf("exit status = %d, want %d (attention)\n%s", code, StatusAttention, stdout)
	}
	if !strings.Contains(stdout, "run-1.2") {
		t.Errorf("the refusal does not name what the agent still holds:\n%s", stdout)
	}
	for _, argv := range fake.Argvs() {
		if argv[0] == "git" {
			t.Errorf("an undrained dispatch reached version control: %v", argv)
		}
	}
}

// Explain runs nothing at all, which is the property that makes it safe
// against a repository nobody has looked at.
func TestAcceptExplainRunsNothing(t *testing.T) {
	fake := exectest.NewFake() // no recordings: anything executed is a test failure
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--explain")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}
	if calls := fake.Argvs(); len(calls) != 0 {
		t.Errorf("explain ran %d invocation(s): %v", len(calls), calls)
	}
	for _, want := range []string{"merge-base --is-ancestor", "--ff-only", "rebase"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("the explained sequence does not mention %q:\n%s", want, stdout)
		}
	}
}

// The merged-head resolve is taken AFTER the fast-forward, so a failure
// of it is a failure of a read that follows an irreversible write. It
// must not erase the record of the write: the fold happened, merged is
// true, and the whole removal half — the ancestry guard, the post-merge
// bead re-check, the worktree removal, the branch deletion — still runs.
// Only the sha is unknown, and it is reported unestablished.
//
// The fake fails `git rev-parse --verify <target>^{commit}` only on its
// SECOND invocation: the case reading takes the first and the post-fold
// resolve takes this one, so the fold is reached with everything before
// it succeeding.
func TestAFailedPostFoldResolveStillReportsTheMergeAndStillSettles(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...) // serialized
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, exectest.Recording{
		Argv:       []string{"git", "rev-parse", "--verify", acceptTargetBr + "^{commit}"},
		Stderr:     []byte("fatal: unable to read the repository\n"),
		ExitStatus: 128,
	})
	recs = append(recs, settleReads(cleanSettle())...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent,
		"--no-push", "--json", "--full-result")

	// The fold landed, so the accept is not abandoned.
	if code == int(StatusStop) {
		t.Fatalf("exit status = %d: the accept was abandoned after the fast-forward had already landed\n%s",
			code, stdout)
	}

	env := envelopeOf(t, stdout)
	merged, ok := factNamed(env, "merged")
	if !ok {
		t.Fatalf("the result carries no merged fact at all, so a caller cannot tell the merge happened:\n%s", stdout)
	}
	if !merged.Established || merged.Value == nil || merged.Value.Flag == nil || !*merged.Value.Flag {
		t.Errorf("merged = %v, want an established true — the fast-forward succeeded", merged.Value)
	}

	sha, ok := factNamed(env, "merged_sha")
	if !ok {
		t.Fatalf("merged_sha is omitted; an unknown value is reported unestablished, not dropped:\n%s", stdout)
	}
	if sha.Established {
		t.Errorf("merged_sha = %v is reported established, but the resolve that would have taken it failed",
			sha.Value)
	}
	if !strings.Contains(sha.Reason, "after the fold") {
		t.Errorf("merged_sha's reason does not name the resolve that could not answer: %q", sha.Reason)
	}

	// settleDispatch was reached: the ancestry guard, the worktree reads
	// and the branch deletion all ran.
	guard, deleted := false, false
	for _, argv := range fake.Argvs() {
		switch {
		case len(argv) >= 4 && argv[0] == "git" && argv[1] == "merge-base" &&
			argv[3] == acceptBranch && argv[4] == acceptTargetBr:
			guard = true
		case len(argv) >= 3 && argv[0] == "git" && argv[1] == "branch" && argv[2] == "-d":
			deleted = true
		}
	}
	if !guard {
		t.Errorf("the post-merge ancestry guard never ran, so settleDispatch was not reached: %v", fake.Argvs())
	}
	if !deleted {
		t.Errorf("the dispatch branch was never deleted, so the removal half was skipped: %v", fake.Argvs())
	}
}

// The negative control for the test above: with the same world and a
// resolve that ANSWERS, merged_sha is established from it. So what that
// test asserts is the degraded report on a failed read, not merged_sha
// having stopped being reported at all.
func TestAPostFoldResolveThatAnswersStillEstablishesTheMergedSHA(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, exectest.Recording{
		Argv:   []string{"git", "rev-parse", "--verify", acceptTargetBr + "^{commit}"},
		Stdout: []byte(acceptBranchSH + "\n"),
	})
	recs = append(recs, settleReads(cleanSettle())...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent,
		"--no-push", "--json", "--full-result")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}

	sha, ok := factNamed(envelopeOf(t, stdout), "merged_sha")
	if !ok {
		t.Fatalf("merged_sha is missing from a clean accept:\n%s", stdout)
	}
	if !sha.Established || sha.Value == nil || sha.Value.Text == nil || *sha.Value.Text != acceptBranchSH {
		t.Errorf("merged_sha = %v (established %v), want an established %s",
			sha.Value, sha.Established, acceptBranchSH)
	}
}

// skippedMergedHeadVC is the real Git Version Control with one
// substitution: the post-fold `merged-head` resolve is preceded by a
// guard step that stops the plan, so the resolve itself is recorded
// [plan.StepSkipped] with an empty invocation.
//
// That is the state [foldIn]'s own comment anticipates — "an absent
// record is among the ways the resolve below fails" — and the one the
// real fragment cannot be driven into from outside, because every step
// that could stop the plan ahead of it is one [foldIn] already returns
// on. The substitution keeps every argument a literal, exactly as the
// real fragment's are, so the invocation the fact cites is the one the
// step declares.
type skippedMergedHeadVC struct{ vcs.VersionControl }

// mergedHeadGuardRef is a revision nothing resolves, so the guard fails.
const mergedHeadGuardRef = "refs/heads/nothing-resolves-this"

func (v skippedMergedHeadVC) Resolve(prefix, revision string) vcs.Steps {
	steps := v.VersionControl.Resolve(prefix, revision)
	if prefix != "merged-head" {
		return steps
	}
	guard := plan.Step{
		Name:        prefix + "-guard",
		Description: "stop the plan before the resolve",
		Executable:  "git",
		Args: []plan.Arg{
			plan.Lit("rev-parse"), plan.Lit("--verify"), plan.Lit(mergedHeadGuardRef),
		},
		Effect:    plan.Read,
		OnFailure: plan.Stop,
	}
	return append(vcs.Steps{guard}, steps...)
}

// A merged-head record that never ran carries no invocation of its own.
// The fold still happened, so the merged flag and the unestablished
// merged_sha both have to survive it — and a fact naming no invocation
// does not survive [Result.Validate], which is what turns the whole
// result into a bare refusal after a fast-forward that cannot be undone.
func TestASkippedMergedHeadResolveStillCitesAnInvocation(t *testing.T) {
	fake := exectest.NewFake(
		exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}},
		exectest.Recording{
			Argv:       []string{"git", "rev-parse", "--verify", mergedHeadGuardRef},
			Stderr:     []byte("fatal: Needed a single revision\n"),
			ExitStatus: 128,
		},
	)
	req := Request{Command: "swarm dispatch accept", Run: acceptRun, Mode: plan.ModeExecute}
	m := mergeCase{
		targetSHA:  acceptTargetSH,
		serialized: true,
		caseInv:    []string{"git", "merge-base", "--is-ancestor", acceptTargetBr, acceptBranch},
	}

	out, records, err := foldIn(context.Background(), nil,
		skippedMergedHeadVC{gitvcs.VersionControl{}}, plan.Runner{Exec: fake},
		req, acceptTargetBr, acceptBranch, m)
	if err != nil {
		t.Fatalf("foldIn = %v: the fast-forward landed, so the fold is not abandoned", err)
	}

	// The world the assertions below are about: the fold ran and the
	// resolve after it was skipped without an invocation of its own.
	var landed plan.Record
	for _, rec := range records {
		if rec.Step == "merged-head" {
			landed = rec
		}
	}
	if landed.Status != plan.StepSkipped {
		t.Fatalf("merged-head is %q, want %q — the scenario did not arise",
			landed.Status, plan.StepSkipped)
	}
	if landed.Invocation.Name != "" {
		t.Fatalf("merged-head carries invocation %q; the scenario needs a record that never ran",
			landed.Invocation.Name)
	}

	var merged, sha Fact
	var haveMerged, haveSHA bool
	for _, f := range out.facts {
		switch f.Name {
		case "merged":
			merged, haveMerged = f, true
		case "merged_sha":
			sha, haveSHA = f, true
		}
	}
	if !haveMerged {
		t.Fatalf("the fold reports no merged fact, so a caller cannot tell the merge happened: %+v", out.facts)
	}
	if !merged.Established || merged.Value == nil || merged.Value.Flag == nil || !*merged.Value.Flag {
		t.Errorf("merged = %v, want an established true — the fast-forward succeeded", merged.Value)
	}
	if !haveSHA {
		t.Fatalf("merged_sha is omitted; an unknown value is reported unestablished, not dropped: %+v", out.facts)
	}
	if sha.Established {
		t.Errorf("merged_sha is established, but the resolve that would have taken it never ran")
	}
	if len(sha.Invocations) == 0 {
		t.Errorf("merged_sha names no invocation that established it")
	}
	for i, argv := range sha.Invocations {
		if len(argv) == 0 {
			t.Errorf("merged_sha's invocation %d is empty", i)
		}
	}

	// The defect this guards is not the fact in isolation: a fact naming
	// no invocation makes the WHOLE result unrenderable, and finish
	// substitutes a bare refusal for it — losing merged: true after an
	// irreversible fast-forward.
	result := req.stamp(Result{Outcome: OutcomeClear, Facts: out.facts, Records: records})
	if err := result.Validate(); err != nil {
		t.Errorf("the result does not validate, so it degrades to a bare refusal: %v", err)
	}
}

// The fold is reported, and so is the fact that its sha is unknown. An
// unestablished fact takes the attention outcome everywhere else in this
// package — breakerUnread in ready.go, landingVerdict in branch.go — and
// this path is no different: a scripted caller reading exit 0 beside a
// null merged_sha has success and nothing naming a remedy.
//
// Same world as
// TestAFailedPostFoldResolveStillReportsTheMergeAndStillSettles, which
// asserts what must NOT change: the merge is still reported and the
// removal half still runs.
func TestAFailedPostFoldResolveTakesTheAttentionOutcome(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...) // serialized
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, exectest.Recording{
		Argv:       []string{"git", "rev-parse", "--verify", acceptTargetBr + "^{commit}"},
		Stderr:     []byte("fatal: unable to read the repository\n"),
		ExitStatus: 128,
	})
	recs = append(recs, settleReads(cleanSettle())...)

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent,
		"--no-push", "--json", "--full-result")

	if code != int(StatusAttention) {
		t.Errorf("exit status = %d, want %d: an unestablished merged_sha is a condition to act on\n%s",
			code, StatusAttention, stdout)
	}

	env := envelopeOf(t, stdout)
	if env.Outcome != swarmmodel.CommandOutcome(OutcomeAttention) {
		t.Errorf("outcome = %q, want %q", env.Outcome, OutcomeAttention)
	}
	if strings.TrimSpace(env.Resolution) == "" {
		t.Fatalf("the resolution is empty, so a caller has a status and nothing to act on:\n%s", stdout)
	}
	// The resolution names the fact that is missing and something to run.
	if !strings.Contains(env.Resolution, "merged_sha") {
		t.Errorf("the resolution does not name merged_sha: %q", env.Resolution)
	}
	if !strings.Contains(env.Resolution, "git rev-parse") {
		t.Errorf("the resolution names nothing the caller can run: %q", env.Resolution)
	}

	// What must survive the change: the fold is still reported.
	merged, ok := factNamed(env, "merged")
	if !ok {
		t.Fatalf("the result carries no merged fact at all:\n%s", stdout)
	}
	if !merged.Established || merged.Value == nil || merged.Value.Flag == nil || !*merged.Value.Flag {
		t.Errorf("merged = %v, want an established true — the fast-forward succeeded", merged.Value)
	}
	if !strings.Contains(env.Summary, "fast-forwarded into it") {
		t.Errorf("the fold's summary was dropped: %q", env.Summary)
	}
}

// Both halves can have something the caller must act on at once: the
// fold landed without establishing what landed, AND the removal half
// found the worktree still standing. The removal half's verdict is the
// later and more severe, so it takes the outcome — but its resolution is
// appended to the fold's rather than substituted for it, because the
// fold's describes a write that already happened and cannot be re-read
// from the result once it is dropped.
func TestBothHalvesNeedingAttentionKeepBothResolutions(t *testing.T) {
	w := cleanSettle()
	w.worktree = "/wt/dev-1"
	w.survives = true // still standing on the post-merge read

	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...) // serialized
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, exectest.Recording{
		Argv:       []string{"git", "rev-parse", "--verify", acceptTargetBr + "^{commit}"},
		Stderr:     []byte("fatal: unable to read the repository\n"),
		ExitStatus: 128,
	})
	recs = append(recs, settleReads(w)...)

	stdout, _, code := runAccept(t, exectest.NewFake(recs...),
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent,
		"--no-push", "--json", "--full-result")

	env := envelopeOf(t, stdout)
	if env.Outcome == swarmmodel.CommandOutcome(OutcomeClear) {
		t.Fatalf("outcome = %q: neither half's verdict reached the result\n%s", env.Outcome, stdout)
	}
	if code == int(StatusClear) {
		t.Errorf("exit status = %d: two conditions to act on and the caller is told nothing", code)
	}
	if !strings.Contains(env.Resolution, "merged_sha") {
		t.Errorf("the fold's resolution was dropped when the removal half added its own: %q", env.Resolution)
	}
	// The removal half's own words, and not a string the fold's
	// resolution also happens to carry.
	if !strings.Contains(env.Resolution, "accept the dispatch again") {
		t.Errorf("the removal half's resolution is missing: %q", env.Resolution)
	}
}

// The rule the post-fold resolve above states, applied to the WHOLE
// settle half: once the fast-forward has landed, no failure after it may
// abandon the accept. `git merge --ff-only` cannot be undone, so a
// refusal raised anywhere in [settleDispatch] degrades the report — it
// does not replace it with a bare refusal that carries no merged fact at
// all, leaving a caller unable to tell the fold happened.
//
// The world: the fold succeeds and establishes its sha, and the ancestry
// read that opens the settle half then fails outright. Exit 128 rather
// than 1 is what makes it a failure instead of an answer — `Descends`
// declares `Answers = []int{1}`, so 1 is "no" and anything else is a read
// that never reached its verdict.
//
// Why the ancestry read and not the guard the defect was reported
// against. [recheckDispatchBeads]' `invocationsRequired` on the beads
// read raises the same kind of refusal — a plain *Refusal rather than an
// *acceptStop — through the same return, but that one cannot be reached
// from outside today: citation_test.go's header records the three
// measured guards (BeadsFrom's refusal of a record that did not run,
// [tracker.Descendants] hardcoding [plan.ModeExecute], and
// [plan.Plan.Validate] refusing a step with no executable) that make
// [tracker.Descendants] unable to answer successfully with records
// nothing cites. The ancestry read reaches the identical composition and
// IS reachable, so this test drives the mechanism rather than the one
// entrance to it that is currently a defence.
func TestASettleHalfRefusalAfterTheFoldStillReportsTheMerge(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...) // serialized
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, exectest.Recording{
		Argv:   []string{"git", "rev-parse", "--verify", acceptTargetBr + "^{commit}"},
		Stdout: []byte(acceptBranchSH + "\n"),
	})
	// The settle half's first read, and it does not answer.
	recs = append(recs, exectest.Recording{
		Argv:       []string{"git", "merge-base", "--is-ancestor", acceptBranch, acceptTargetBr},
		Stderr:     []byte("fatal: Not a valid commit name " + acceptBranch + "\n"),
		ExitStatus: 128,
	})

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent,
		"--no-push", "--json", "--full-result")

	if !ran(fake, "git", "merge", "--ff-only", acceptBranch) {
		t.Fatalf("the fold never ran, so the scenario did not arise: %v", fake.Argvs())
	}

	env := envelopeOf(t, stdout)
	merged, ok := factNamed(env, "merged")
	if !ok {
		t.Fatalf("the result carries no merged fact at all, so a caller cannot tell the fast-forward "+
			"already landed and cannot be undone:\n%s", stdout)
	}
	if !merged.Established || merged.Value == nil || merged.Value.Flag == nil || !*merged.Value.Flag {
		t.Errorf("merged = %v, want an established true — the fast-forward succeeded", merged.Value)
	}

	sha, ok := factNamed(env, "merged_sha")
	if !ok {
		t.Fatalf("merged_sha was dropped; the fold established it before the settle half failed:\n%s", stdout)
	}
	if !sha.Established || sha.Value == nil || sha.Value.Text == nil || *sha.Value.Text != acceptBranchSH {
		t.Errorf("merged_sha = %v (established %v), want an established %s",
			sha.Value, sha.Established, acceptBranchSH)
	}

	// The refusal is still reported as one — the accept is degraded, not
	// papered over — and it names what could not be established.
	if env.Outcome != swarmmodel.CommandOutcome(OutcomeStop) {
		t.Errorf("outcome = %q, want %q: the settle half found a condition the caller cannot repair in place",
			env.Outcome, OutcomeStop)
	}
	if code != int(StatusStop) {
		t.Errorf("exit status = %d, want %d\n%s", code, StatusStop, stdout)
	}
	if !strings.Contains(env.Summary, "ancestor") {
		t.Errorf("the summary does not name the read that failed: %q", env.Summary)
	}
	if !strings.Contains(env.Summary, "fast-forwarded into it") {
		t.Errorf("the fold's summary was dropped: %q", env.Summary)
	}
	if strings.TrimSpace(env.Resolution) == "" {
		t.Fatalf("the resolution is empty, so a caller holding a landed merge is told nothing to do:\n%s", stdout)
	}
	if !strings.Contains(env.Resolution, "accept the dispatch again") {
		t.Errorf("the resolution does not say what to do with a fold that already landed: %q", env.Resolution)
	}
}

// The first negative control for the test above: the same world with an
// ancestry read that ANSWERS settles cleanly. So what that test asserts
// is the degraded report on a settle-half refusal, not the merged facts
// having become unconditional.
func TestASettleHalfThatAnswersStillAcceptsCleanly(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	recs = append(recs, exectest.Recording{
		Argv:   []string{"git", "rev-parse", "--verify", acceptTargetBr + "^{commit}"},
		Stdout: []byte(acceptBranchSH + "\n"),
	})
	recs = append(recs, settleReads(cleanSettle())...)

	stdout, _, code := runAccept(t, exectest.NewFake(recs...),
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent,
		"--no-push", "--json", "--full-result")
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, stdout)
	}
	if sha, ok := factNamed(envelopeOf(t, stdout), "merged_sha"); !ok || !sha.Established {
		t.Errorf("merged_sha is missing or unestablished on a clean accept:\n%s", stdout)
	}
}

// The second negative control, and the boundary the rule is drawn at:
// BEFORE the fold has run, nothing irreversible has happened, so a
// settle-half refusal still abandons the accept. A dry run refuses the
// fast-forward and reaches the same ancestry read — which is a read, so
// it still runs — and the failure is answered with a bare refusal.
//
// Without this, "degrade instead of abandoning" would have been widened
// to a path where there is nothing to preserve, and a dry run would
// report merged: false beside a refusal it never had to survive.
func TestASettleHalfRefusalBeforeAnyFoldStillAbandonsTheAccept(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{
		Argv:       []string{"git", "merge-base", "--is-ancestor", acceptBranch, acceptTargetBr},
		Stderr:     []byte("fatal: Not a valid commit name " + acceptBranch + "\n"),
		ExitStatus: 128,
	})

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent,
		"--dry-run", "--json", "--full-result")

	if ran(fake, "git", "merge", "--ff-only", acceptBranch) {
		t.Fatalf("a dry run folded; the scenario needs a settle-half failure with nothing merged: %v", fake.Argvs())
	}
	if code != int(StatusStop) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusStop, stdout)
	}
	env := envelopeOf(t, stdout)
	if _, ok := factNamed(env, "merged"); ok {
		t.Errorf("a refusal raised before any fold carries facts of the accept; nothing irreversible "+
			"happened, so it is answered bare:\n%s", stdout)
	}
	if !strings.Contains(env.Summary, "ancestor") {
		t.Errorf("the refusal does not name the read that failed: %q", env.Summary)
	}
}

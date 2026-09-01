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
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	manifestbeads "github.com/brettcallaghan/agent-plugins/cli/swarm/manifest/beads"
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
			Argv:   []string{"bd", "list", "--assignee", acceptAgent, "--status", "in_progress", "--limit", "0", "--json"},
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

	fake := exectest.NewFake(recs...)
	stdout, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json")
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
// command reads no worktree list at all, so a world with none standing
// and a world with several produce the same reading.
func TestTheCaseIsNotInferredFromTheStandingWorktrees(t *testing.T) {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})

	fake := exectest.NewFake(recs...)
	if _, _, code := runAccept(t, fake,
		"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--json"); code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d", code, StatusClear)
	}
	for _, argv := range fake.Argvs() {
		if len(argv) >= 2 && argv[0] == "git" && argv[1] == "worktree" {
			t.Errorf("the case reading consulted the worktree list: %v", argv)
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

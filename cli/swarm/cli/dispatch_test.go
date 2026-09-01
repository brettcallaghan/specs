// Unit tests for `swarm dispatch open`: the run-scope refusal, that
// opening the first dispatch of a run creates the sprint branch and
// records it — in that order — rather than the other way round, and
// (with --dispatch) that the worktree half checks for an existing
// worktree before creating one, clearing stale registrations first, and
// defaults its base to the sprint branch's own head. Every scenario runs
// through the Beads Tracker's and the Git Version Control's real
// argv-building against the fake process runner, so no test starts `bd`
// or `git`.
//
// AC-SWARM-DISP-003.2, .3 and .5 are declared e2e and live in
// cli/swarm/e2e against the core fixture and a run epic created fresh in
// it. AC-SWARM-DISP-003.1, .4 and .11 are likewise e2e, in
// dispatch_open_worktree_test.go.
package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	manifestbeads "github.com/brettcallaghan/agent-plugins/cli/swarm/manifest/beads"
	beadstracker "github.com/brettcallaghan/agent-plugins/cli/swarm/tracker/beads"
	gitvcs "github.com/brettcallaghan/agent-plugins/cli/swarm/vcs/git"
)

// AC: AC-SWARM-DISP-003.3 (@negative in FEAT-SWARM-DISP-002, restated
// here for the run-scope refusal every command carries)
//
// An unscoped invocation is refused with the malformed-invocation status,
// and nothing is read — asserted from the fake's own record.
func TestDispatchOpenRefusesAnEmptyScope(t *testing.T) {
	for _, args := range [][]string{
		{"dispatch", "open"},
		{"dispatch", "open", "--run", ""},
		{"dispatch", "open", "--run", "   "},
	} {
		fake := exectest.NewFake() // no recordings: anything executed is a test failure
		tr := beadstracker.Tracker{}
		streams := &ioStreams{out: &bytes.Buffer{}, err: &bytes.Buffer{}}
		root := newRootCmd(streams)
		root.AddCommand(dispatchCmd(streams, manifestbeads.Store{Tracker: tr}, tr, gitvcs.VersionControl{}, fake))

		code := dispatch(root, streams, args, &bytes.Buffer{}, &bytes.Buffer{})
		if code != int(StatusMalformed) {
			t.Errorf("%v exited %d, want %d", args, code, StatusMalformed)
		}
		if calls := fake.Argvs(); len(calls) != 0 {
			t.Errorf("%v read or wrote something outside the scope: %v", args, calls)
		}
	}
}

// AC: AC-SWARM-DISP-003.6
//
// Opening the first dispatch of a run whose manifest records no sprint
// branch creates it and records it — asserted from the fake's own call
// order, so a command that recorded before creating (or the reverse of
// what this AC asks) would fail here even if its final answer looked
// right.
func TestDispatchOpenCreatesTheBranchBeforeRecordingIt(t *testing.T) {
	tr := beadstracker.Tracker{}
	fake := exectest.NewFake(
		exectest.Recording{
			Argv:   []string{"bd", "show", "run-1", "--json"},
			Stdout: []byte(`[{"id":"run-1","status":"open","issue_type":"epic"}]`),
		},
		exectest.Recording{
			Argv: []string{"git", "for-each-ref",
				"--format=%(refname:short)%00%(objectname)%00%(upstream:short)", "refs/heads/"},
			Stdout: []byte("main\x00" + strings.Repeat("a", 40) + "\x00\n"),
		},
		exectest.Recording{
			Argv:   []string{"git", "rev-parse", "--verify", "HEAD^{commit}"},
			Stdout: []byte(strings.Repeat("b", 40) + "\n"),
		},
		exectest.Recording{
			Argv: []string{"git", "branch", "sprint/run-1", strings.Repeat("b", 40)},
		},
		exectest.Recording{
			Argv:   []string{"bd", "update", "run-1", "--metadata", `{"base_sha":"` + strings.Repeat("b", 40) + `","branch":"sprint/run-1"}`, "--json"},
			Stdout: []byte(`[{"id":"run-1","status":"open","issue_type":"epic","metadata":{"branch":"sprint/run-1","base_sha":"` + strings.Repeat("b", 40) + `"}}]`),
		},
	)

	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut}
	root := newRootCmd(streams)
	root.AddCommand(dispatchCmd(streams, manifestbeads.Store{Tracker: tr}, tr, gitvcs.VersionControl{}, fake))

	code := dispatch(root, streams, []string{"dispatch", "open", "--run", "run-1", "--json"}, &out, &errOut)
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, out.String())
	}

	var createIdx, showAgainIdx, writeIdx = -1, -1, -1
	for i, argv := range fake.Argvs() {
		switch {
		case len(argv) >= 2 && argv[0] == "git" && argv[1] == "branch":
			createIdx = i
		case len(argv) >= 2 && argv[0] == "bd" && argv[1] == "show" && i > 0:
			showAgainIdx = i
		case len(argv) >= 2 && argv[0] == "bd" && argv[1] == "update":
			writeIdx = i
		}
	}
	if createIdx < 0 {
		t.Fatalf("git branch was never called: %v", fake.Argvs())
	}
	if writeIdx < 0 {
		t.Fatalf("bd update was never called: %v", fake.Argvs())
	}
	if !(createIdx < showAgainIdx && showAgainIdx < writeIdx) {
		t.Errorf("call order = create@%d, guard-show@%d, write@%d; want create < guard-show < write",
			createIdx, showAgainIdx, writeIdx)
	}

	if !strings.Contains(out.String(), "sprint/run-1") {
		t.Errorf("the answer does not name the created branch:\n%s", out.String())
	}
}

// AC: AC-SWARM-DISP-003.7 @negative
//
// The worktree path is a sibling of the repository root, never inside
// it — pure, so asserted directly without a fake.
func TestDispatchWorktreePathIsOutsideTheRepository(t *testing.T) {
	root := filepath.FromSlash("/work/repo")
	got := dispatchWorktreePath(root, "dev-1")

	want := filepath.Join(filepath.Dir(root), ".bb-worktrees", "dev-1")
	if got != want {
		t.Fatalf("dispatchWorktreePath(%q, %q) = %q, want %q", root, "dev-1", got, want)
	}
	if strings.HasPrefix(got, root+string(filepath.Separator)) || got == root {
		t.Errorf("dispatchWorktreePath(%q, %q) = %q, which is inside the repository", root, "dev-1", got)
	}
}

// AC: AC-SWARM-DISP-003.8
//
// The worktree path and the dispatch branch name follow from the
// repository root and the dispatch id alone: two computations given the
// same inputs produce the same values, which is what lets two sessions
// holding the same run epic id and the same dispatch id agree without
// coordinating.
func TestDispatchWorktreePathAndBranchAreDeterministic(t *testing.T) {
	root := filepath.FromSlash("/work/repo")

	path1 := dispatchWorktreePath(root, "dev-1")
	path2 := dispatchWorktreePath(root, "dev-1")
	if path1 != path2 {
		t.Errorf("dispatchWorktreePath is not deterministic: %q then %q", path1, path2)
	}

	branch1 := dispatchBranchName("dev-1")
	branch2 := dispatchBranchName("dev-1")
	if branch1 != branch2 {
		t.Errorf("dispatchBranchName is not deterministic: %q then %q", branch1, branch2)
	}

	// A different dispatch id computes a different value — determinism
	// is not satisfied by a constant.
	if dispatchWorktreePath(root, "dev-2") == path1 {
		t.Error("dispatchWorktreePath does not vary with the dispatch id")
	}
	if dispatchBranchName("dev-2") == branch1 {
		t.Error("dispatchBranchName does not vary with the dispatch id")
	}
}

// AC: AC-SWARM-DISP-003.9, .10, .12
//
// Given --dispatch and a manifest that already records a sprint branch,
// the worktree half clears stale registrations before it reads the
// worktree list (.9), resolves the worktree's base against the sprint
// branch itself when --base is absent (.10), and — proved by the fake's
// own exhaustiveness, since an invocation matching no recording is a
// test failure — reads nothing else to decide that base (.12).
func TestDispatchOpenDefaultsTheWorktreeBaseToTheSprintBranch(t *testing.T) {
	tr := beadstracker.Tracker{}
	baseSHA := strings.Repeat("b", 40)
	worktreeSHA := strings.Repeat("c", 40)
	mainHead := strings.Repeat("a", 40)

	fake := exectest.NewFake(
		exectest.Recording{
			Argv: []string{"bd", "show", "run-1", "--json"},
			Stdout: []byte(`[{"id":"run-1","status":"open","issue_type":"epic","metadata":{` +
				`"branch":"sprint/run-1","base_sha":"` + baseSHA + `"}}]`),
		},
		exectest.Recording{Argv: []string{"git", "worktree", "prune"}},
		exectest.Recording{
			Argv:   []string{"git", "worktree", "list", "--porcelain"},
			Stdout: []byte("worktree /work/repo\nHEAD " + mainHead + "\nbranch refs/heads/main\n\n"),
		},
		exectest.Recording{
			Argv:   []string{"git", "rev-parse", "--show-toplevel"},
			Stdout: []byte("/work/repo\n"),
		},
		exectest.Recording{
			Argv:   []string{"git", "rev-parse", "--verify", "sprint/run-1^{commit}"},
			Stdout: []byte(worktreeSHA + "\n"),
		},
		exectest.Recording{
			Argv: []string{"git", "worktree", "add", "-b", "bb/dev-1",
				"/work/.bb-worktrees/dev-1", worktreeSHA},
		},
	)

	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut}
	root := newRootCmd(streams)
	root.AddCommand(dispatchCmd(streams, manifestbeads.Store{Tracker: tr}, tr, gitvcs.VersionControl{}, fake))

	code := dispatch(root, streams,
		[]string{"dispatch", "open", "--run", "run-1", "--dispatch", "dev-1", "--json"}, &out, &errOut)
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, out.String())
	}

	var pruneIdx, listIdx, resolveIdx = -1, -1, -1
	for i, argv := range fake.Argvs() {
		switch {
		case len(argv) >= 2 && argv[0] == "git" && argv[1] == "worktree" && len(argv) > 2 && argv[2] == "prune":
			pruneIdx = i
		case len(argv) >= 3 && argv[0] == "git" && argv[1] == "worktree" && argv[2] == "list":
			listIdx = i
		case len(argv) >= 4 && argv[0] == "git" && argv[1] == "rev-parse" && argv[2] == "--verify" &&
			argv[3] == "sprint/run-1^{commit}":
			resolveIdx = i
		}
	}
	if pruneIdx < 0 || listIdx < 0 {
		t.Fatalf("prune and list were not both called: %v", fake.Argvs())
	}
	if pruneIdx >= listIdx {
		t.Errorf("prune@%d, list@%d; want stale registrations cleared before the list is read", pruneIdx, listIdx)
	}
	if resolveIdx < 0 {
		t.Errorf("the worktree's base was not resolved against the sprint branch sprint/run-1: %v", fake.Argvs())
	}

	if !strings.Contains(out.String(), "bb/dev-1") {
		t.Errorf("the answer does not name the dispatch branch:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "/work/.bb-worktrees/dev-1") {
		t.Errorf("the answer does not name the worktree path:\n%s", out.String())
	}
}

// AC: none owed here (--dispatch absent) — asserts the worktree half is
// skipped entirely when no dispatch id is given, which is what keeps
// TestDispatchOpenCreatesTheBranchBeforeRecordingIt and the e2e sprint-
// branch cases (which pass no --dispatch) exercising the sprint branch
// alone, unchanged by this half's addition.
func TestDispatchOpenWithNoDispatchIDTouchesNoWorktree(t *testing.T) {
	tr := beadstracker.Tracker{}
	fake := exectest.NewFake(
		exectest.Recording{
			Argv:   []string{"bd", "show", "run-1", "--json"},
			Stdout: []byte(`[{"id":"run-1","status":"open","issue_type":"epic","metadata":{"branch":"sprint/run-1"}}]`),
		},
		// No worktree recording declared: any attempt to touch one fails
		// the fake.
	)

	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut}
	root := newRootCmd(streams)
	root.AddCommand(dispatchCmd(streams, manifestbeads.Store{Tracker: tr}, tr, gitvcs.VersionControl{}, fake))

	code := dispatch(root, streams, []string{"dispatch", "open", "--run", "run-1", "--json"}, &out, &errOut)
	if code != int(StatusClear) {
		t.Fatalf("exit status = %d, want %d\n%s", code, StatusClear, out.String())
	}
	if strings.Contains(out.String(), "worktree") {
		t.Errorf("a worktree fact was reported despite no --dispatch:\n%s", out.String())
	}
}

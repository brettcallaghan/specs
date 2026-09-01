// Unit tests for `swarm dispatch open`: the run-scope refusal, and that
// opening the first dispatch of a run creates the sprint branch and
// records it — in that order — rather than the other way round. Every
// scenario runs through the Beads Tracker's and the Git Version
// Control's real argv-building against the fake process runner, so no
// test starts `bd` or `git`.
//
// AC-SWARM-DISP-003.2, .3 and .5 are declared e2e and live in
// cli/swarm/e2e against the core fixture and a run epic created fresh in
// it.
package cli

import (
	"bytes"
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
		root.AddCommand(dispatchCmd(streams, manifestbeads.Store{Tracker: tr}, gitvcs.VersionControl{}, fake))

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
	root.AddCommand(dispatchCmd(streams, manifestbeads.Store{Tracker: tr}, gitvcs.VersionControl{}, fake))

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

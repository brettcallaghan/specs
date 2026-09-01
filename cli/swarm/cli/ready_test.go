// Unit tests for `swarm ready`: the run-scope refusal, and a tracker
// response that is not a queue's shape reading as unanswered rather than
// as a length derived from it. Every scenario runs through the Beads
// Tracker's real argv-building against the fake process runner, so no
// test starts `bd`.
//
// AC-SWARM-DISP-002.1, .2, .7, .8, .9, .10 and .17 are declared e2e and
// live in cli/swarm/e2e against the core and dispatch-worktree fixtures.
package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	manifestbeads "github.com/brettcallaghan/agent-plugins/cli/swarm/manifest/beads"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/status"
	beadstracker "github.com/brettcallaghan/agent-plugins/cli/swarm/tracker/beads"
	gitvcs "github.com/brettcallaghan/agent-plugins/cli/swarm/vcs/git"
)

func readyDeps(fake *exectest.Fake) status.Deps {
	tr := beadstracker.Tracker{}
	return status.Deps{
		Tracker: tr,
		VCS:     gitvcs.VersionControl{},
		Store:   manifestbeads.Store{Tracker: tr},
		Runner:  plan.Runner{Exec: fake},
	}
}

// AC: AC-SWARM-DISP-002.3
//
// An unscoped invocation is refused with the malformed-invocation status,
// and nothing is read — asserted from the fake's own record.
func TestReadyRefusesAnEmptyScope(t *testing.T) {
	for _, args := range [][]string{
		{"ready"},
		{"ready", "--run", ""},
		{"ready", "--run", "   "},
	} {
		fake := exectest.NewFake() // no recordings: anything executed is a test failure
		streams := &ioStreams{out: &bytes.Buffer{}, err: &bytes.Buffer{}}
		root := newRootCmd(streams)
		root.AddCommand(readyCmd(streams, readyDeps(fake)))

		code := dispatch(root, streams, args, &bytes.Buffer{}, &bytes.Buffer{})
		if code != int(StatusMalformed) {
			t.Errorf("%v exited %d, want %d", args, code, StatusMalformed)
		}
		if calls := fake.Argvs(); len(calls) != 0 {
			t.Errorf("%v read something outside the scope: %v", args, calls)
		}
	}
}

// AC: AC-SWARM-DISP-002.16
//
// A tracker response that is an error body rather than a list is not
// counted as a queue of that body's length; it fails to decode, and the
// command takes the stop status.
func TestReadyReportsANonListPayloadAsUnanswered(t *testing.T) {
	fake := exectest.NewFake(exectest.Recording{
		Argv:   []string{"bd", "ready", "--parent", "run-1", "--exclude-type", "epic", "--exclude-label", "human", "--limit", "0", "--json"},
		Stdout: []byte(`{"error":"database is locked"}`),
	})
	streams := &ioStreams{out: &bytes.Buffer{}, err: &bytes.Buffer{}}
	root := newRootCmd(streams)
	root.AddCommand(readyCmd(streams, readyDeps(fake)))

	var out bytes.Buffer
	code := dispatch(root, streams, []string{"ready", "--run", "run-1", "--json"}, &out, &bytes.Buffer{})
	if code != int(StatusStop) {
		t.Errorf("exit status = %d, want %d for a payload that is not a queue's shape", code, StatusStop)
	}
	if strings.Contains(out.String(), `"ready_count":1`) {
		t.Errorf("the error body's own length leaked into ready_count:\n%s", out.String())
	}
}

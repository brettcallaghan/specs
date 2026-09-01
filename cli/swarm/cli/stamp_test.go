// Tests that every command reports which invocation produced its result:
// the command path, the run it was scoped to, and the interpreter that
// actually ran.
//
// The subject is the REAL result — the one an execute or a dry run
// builds — not the explain and refusal paths, which have always carried
// the three fields. A real result missing them reports mode as explain,
// since [plan.ModeExplain] is Mode's zero value, so a consumer routing on
// the mode reads a run that merged a branch and deleted a ref as a no-op.
//
// The table is driven from the real command tree, built by [register] —
// the one place a delta adding a command touches — and every leaf of that
// tree must appear in it. A further hand-declared command therefore fails
// here until its own real result is covered, rather than shipping with
// the fields unset.
package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
)

// stampCase is one command's real result: the invocation that reaches it,
// what the process runner must answer for it to get there, and the three
// fields the result has to carry back.
type stampCase struct {
	// name distinguishes two cases of the same command.
	name string
	// path is the command path cobra reports, which is what the result's
	// command field must equal.
	path string
	// args is the whole invocation, without --full-result --json: the
	// harness adds both, since the three fields live in the full result.
	args []string
	// run is what the result's run field must equal. Empty for a command
	// that takes no --run, which is asserted rather than left incidental.
	run string
	// mode is the interpreter the flags select, and what the result's
	// mode field must name.
	mode plan.Mode
	// recordings are what the fake process runner answers.
	recordings []exectest.Recording
	// status is the exit status this case reaches its real result with.
	// An attention result is as real as a clear one.
	status Status
}

func stampCases() []stampCase {
	return []stampCase{
		{
			name: "phase close, a dry run that names what it would close",
			path: "swarm phase close",
			args: []string{"phase", "close", "--run", "run-1", "--dry-run"},
			run:  "run-1", mode: plan.ModeDryRun, status: StatusClear,
			recordings: []exectest.Recording{
				children("run-1", "["+epicRow("run-1.a")+"]"),
				children("run-1.a", "[]"),
				readyAt("run-1", "[]"),
				closable(`[{"epic":` + epicRow("run-1.a") + `,"total_children":1,"closed_children":1,"eligible_for_close":true}]`),
			},
		},
		{
			name: "ready, the whole-run dispatch set",
			path: "swarm ready",
			args: []string{"ready", "--run", "run-1", "--dry-run"},
			run:  "run-1", mode: plan.ModeDryRun, status: StatusClear,
			recordings: []exectest.Recording{
				{
					Argv: []string{"bd", "ready", "--parent", "run-1", "--exclude-type", "epic",
						"--exclude-label", "human", "--limit", "0", "--json"},
					Stdout: []byte("[]"),
				},
				{
					Argv:   []string{"bd", "show", "run-1", "--json"},
					Stdout: []byte(`[{"id":"run-1","status":"open","issue_type":"epic","metadata":{}}]`),
				},
			},
		},
		{
			name: "ready --agent, one dispatch's own queue",
			path: "swarm ready",
			args: []string{"ready", "--run", acceptRun, "--agent", acceptAgent, "--dry-run"},
			run:  acceptRun, mode: plan.ModeDryRun, status: StatusClear,
			recordings: drainedQueue(),
		},
		{
			name: "dispatch open, the sprint branch created and recorded",
			path: "swarm dispatch open",
			args: []string{"dispatch", "open", "--run", "run-1"},
			run:  "run-1", mode: plan.ModeExecute, status: StatusClear,
			recordings: openSprintBranchReads(),
		},
		{
			name: "dispatch accept, a dispatch that really merged",
			path: "swarm dispatch accept",
			args: []string{"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent, "--no-push"},
			run:  acceptRun, mode: plan.ModeExecute, status: StatusClear,
			recordings: acceptedReads(),
		},
		{
			name: "dispatch accept, a dispatch that still holds work",
			path: "swarm dispatch accept",
			args: []string{"dispatch", "accept", "--run", acceptRun, "--dispatch", acceptAgent},
			run:  acceptRun, mode: plan.ModeExecute, status: StatusAttention,
			recordings: drainedQueue("run-1.2"),
		},
		{
			name: "branch verify-landed, which is not run-scoped",
			path: "swarm branch verify-landed",
			args: []string{"branch", "verify-landed", "--branch", landBranch},
			// Empty deliberately: the landing check takes --branch and
			// --remote and needs no scope, so its result reports the run
			// it was not given as empty rather than as anything else.
			run: "", mode: plan.ModeExecute, status: StatusClear,
			recordings: landedReads("0", "0"),
		},
		{
			name: "branch verify-landed, carrying the scope it was given",
			path: "swarm branch verify-landed",
			args: []string{"branch", "verify-landed", "--branch", landBranch, "--run", "run-1"},
			run:  "run-1", mode: plan.ModeExecute, status: StatusClear,
			recordings: landedReads("0", "0"),
		},
		{
			name: "branch push, the landing check with the push in front of it",
			path: "swarm branch push",
			args: []string{"branch", "push", "--branch", landBranch},
			run:  "", mode: plan.ModeExecute, status: StatusClear,
			recordings: append([]exectest.Recording{{Argv: []string{"git", "push", landRemote, landBranch}}},
				landedReads("0", "0")...),
		},
		{
			name: "status, the report a run's state is read from",
			path: "swarm status",
			args: []string{"status", "--run", "run-1"},
			run:  "run-1", mode: plan.ModeExecute, status: StatusAttention,
			recordings: []exectest.Recording{
				{
					Argv:   []string{"bd", "show", "run-1", "--json"},
					Stdout: []byte(`[{"id":"run-1","status":"open","issue_type":"epic","metadata":{}}]`),
				},
			},
		},
		{
			name: "manifest get",
			path: "swarm manifest get",
			args: []string{"manifest", "get", "--run", testRun, "--key", "branch"},
			run:  testRun, mode: plan.ModeExecute, status: StatusClear,
			recordings: []exectest.Recording{reading(`{"branch":"claude/x"}`)},
		},
		{
			name: "manifest set",
			path: "swarm manifest set",
			args: []string{"manifest", "set", "--run", testRun, "--key", "review_passes", "--value", "2"},
			run:  testRun, mode: plan.ModeExecute, status: StatusClear,
			recordings: []exectest.Recording{
				reading(`{"review_passes":1,"review_budget":2}`),
				{
					Argv:   []string{"bd", "update", testRun, "--metadata", `{"review_passes":2}`, "--json"},
					Stdout: epic(`{"review_passes":2}`),
				},
			},
		},
		{
			name: "manifest require",
			path: "swarm manifest require",
			args: []string{"manifest", "require", "--run", testRun, "--key", "branch"},
			run:  testRun, mode: plan.ModeExecute, status: StatusClear,
			recordings: []exectest.Recording{reading(`{"branch":"claude/x"}`)},
		},
		{
			name: "manifest validate",
			path: "swarm manifest validate",
			args: []string{"manifest", "validate", "--run", testRun},
			run:  testRun, mode: plan.ModeExecute, status: StatusClear,
			recordings: []exectest.Recording{reading(wholeRecord)},
		},
	}
}

// openSprintBranchReads is the manifest read, the branch survey, the base
// resolution, the creation and the write that `swarm dispatch open`
// performs for a run whose manifest records no branch yet.
func openSprintBranchReads() []exectest.Recording {
	const base = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	return []exectest.Recording{
		{
			Argv:   []string{"bd", "show", "run-1", "--json"},
			Stdout: []byte(`[{"id":"run-1","status":"open","issue_type":"epic"}]`),
		},
		{
			Argv: []string{"git", "for-each-ref",
				"--format=%(refname:short)%00%(objectname)%00%(upstream:short)", "refs/heads/"},
			Stdout: []byte("main\x00aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\x00\n"),
		},
		{
			Argv:   []string{"git", "rev-parse", "--verify", "HEAD^{commit}"},
			Stdout: []byte(base + "\n"),
		},
		{Argv: []string{"git", "branch", "sprint/run-1", base}},
		{
			Argv: []string{"bd", "update", "run-1", "--metadata",
				`{"base_sha":"` + base + `","branch":"sprint/run-1"}`, "--json"},
			Stdout: []byte(`[{"id":"run-1","status":"open","issue_type":"epic","metadata":{"branch":"sprint/run-1","base_sha":"` + base + `"}}]`),
		},
	}
}

// acceptedReads is the whole of an accept that drains, folds a serialized
// dispatch in and settles it.
func acceptedReads() []exectest.Recording {
	recs := append(drainedQueue(), acceptManifest())
	recs = append(recs, caseReads(0)...)
	recs = append(recs, exectest.Recording{Argv: []string{"git", "merge", "--ff-only", acceptBranch}})
	return append(recs, settleReads(cleanSettle())...)
}

// runStampCase drives the real command tree — every command [register]
// wires, against the fake — and returns the rendered envelope.
func runStampCase(t *testing.T, c stampCase) (string, int) {
	t.Helper()

	fake := exectest.NewFake(c.recordings...)
	var out, errOut bytes.Buffer
	streams := &ioStreams{out: &out, err: &errOut, exitCode: int(StatusClear)}
	root := newRootCmd(streams)
	register(root, streams, fake)

	args := append(append([]string{}, c.args...), "--full-result", "--json")
	code := dispatch(root, streams, args, &out, &errOut)
	return out.String(), code
}

// TestEveryRealResultNamesTheInvocationThatProducedIt is the defect's own
// assertion: a result reached by really running something reports the
// command, the run and the interpreter that ran, rather than an empty
// command and mode: explain.
func TestEveryRealResultNamesTheInvocationThatProducedIt(t *testing.T) {
	for _, c := range stampCases() {
		t.Run(c.name, func(t *testing.T) {
			stdout, code := runStampCase(t, c)
			if code != int(c.status) {
				t.Fatalf("exit status = %d, want %d\n%s", code, c.status, stdout)
			}

			env := envelopeOf(t, stdout)

			// A refusal carries the three fields already and would pass
			// the assertions below without the real result ever having
			// been built. Every real result states at least one fact;
			// refusal() states none, so this is what keeps the case
			// honest.
			if len(env.Facts) == 0 {
				t.Fatalf("this case did not reach a real result — it states no fact:\n%s", stdout)
			}
			if env.Command != c.path {
				t.Errorf("command = %q, want %q", env.Command, c.path)
			}
			if env.Run != c.run {
				t.Errorf("run = %q, want %q", env.Run, c.run)
			}
			if got := string(env.Mode); got != c.mode.String() {
				t.Errorf("mode = %q, want %q — the interpreter that actually ran", got, c.mode.String())
			}
		})
	}
}

// TestEveryCommandInTheTreeHasARealResultCase makes the table above
// exhaustive against the command tree rather than against whoever last
// edited it: a command added to [register] and not covered here fails,
// which is what stops the next hand-declared command from shipping with
// the three fields unset.
func TestEveryCommandInTheTreeHasARealResultCase(t *testing.T) {
	covered := map[string]bool{}
	for _, c := range stampCases() {
		covered[c.path] = true
	}

	streams := &ioStreams{out: &bytes.Buffer{}, err: &bytes.Buffer{}}
	root := newRootCmd(streams)
	register(root, streams, exectest.NewFake())

	for _, path := range leafPaths(root) {
		if !covered[path] {
			t.Errorf("%q has no case in stampCases(); add one covering its real result, "+
				"so the command, run and mode it reports are asserted rather than assumed", path)
		}
	}
}

// leafPaths is every command in the tree that can be invoked — the ones
// with no subcommands of their own, excluding the root and cobra's own
// help command.
func leafPaths(root *cobra.Command) []string {
	var out []string
	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		children := cmd.Commands()
		if len(children) == 0 {
			if cmd != root && !strings.HasPrefix(cmd.Name(), "help") {
				out = append(out, cmd.CommandPath())
			}
			return
		}
		for _, child := range children {
			walk(child)
		}
	}
	walk(root)
	return out
}

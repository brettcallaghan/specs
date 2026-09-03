// Package beads implements [tracker.Tracker] over the `bd` command line,
// with one decoder per subcommand — per the Beads Tracker component in
// specs/architecture/container-swarm-cli.md and ADR-025.
//
// This is the only component that knows the word `bd`. Everything above
// it speaks the run's vocabulary, so a tracker upgrade that changes a
// result shape changes one decoder here, and a second tracker is a matter
// of satisfying the interface rather than of rewriting the callers.
//
// Every operation returns a plan step: an argv vector and the decoder for
// what comes back. Nothing here starts a process — the Plan Runner does
// that, which is what keeps every invocation visible to explain mode and
// guarded by the plan's ordering policy.
package beads

import (
	"fmt"
	"strings"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
)

// executable is the tool this implementation is coupled to. It is the one
// place the name appears, and a test asserts that no other non-test
// source file in the swarm tree contains it.
const executable = "bd"

// Tracker is [tracker.Tracker] over `bd`.
//
// Bd is the executable to invoke, empty meaning the default. Dir is where
// to invoke it, empty meaning the process's own directory — bd resolves
// its database from the working directory, so a caller reading a
// worktree's tracker sets it rather than changing its own process's
// directory.
type Tracker struct {
	Bd  string
	Dir string
}

var _ tracker.Tracker = Tracker{}

func (t Tracker) tool() string {
	if t.Bd != "" {
		return t.Bd
	}
	return executable
}

// step builds one plan step against bd.
func (t Tracker) step(name string, effect plan.Effect, repeatable bool, decode plan.Decoder, args ...string) plan.Step {
	as := make([]plan.Arg, len(args))
	for i, a := range args {
		as[i] = plan.Lit(a)
	}
	return plan.Step{
		Name:       name,
		Executable: t.tool(),
		Args:       as,
		Dir:        t.Dir,
		Decode:     decode,
		Effect:     effect,
		Repeatable: repeatable,
	}
}

// Bead reads one bead whole, including its metadata.
func (t Tracker) Bead(name, id string) plan.Step {
	s := t.step(name, plan.Read, true, decodeShow, "show", id, "--json")
	s.Description = "read " + id
	return s
}

// Children reads the beads directly under one bead.
//
//   - `--limit 0` because bd's own help documents a default of 50.
//     Measured on bd 1.2.2 this repository's `bd list --all --json`
//     returned all 608 issues without it, so the default appears not to
//     apply to JSON output — but a query whose completeness depends on a
//     default nobody stated is one silent truncation away from reporting
//     a run as smaller than it is, and the flag costs nothing.
//   - `--all` because a completion walk has to see finished beads.
//     Measured: without it, `bd list --parent <run>` omitted three closed
//     children.
func (t Tracker) Children(name, parent string) plan.Step {
	s := t.step(name, plan.Read, true, decodeList,
		"list", "--parent", parent, "--all", "--limit", "0", "--json")
	s.Description = "the beads directly under " + parent
	return s
}

// Ready reads the beads under a scope that nothing is blocking.
//
//   - `--parent` is transitive here, unlike on `bd list`. Measured on bd
//     1.2.2: `bd ready --parent <run>` returned a bead two levels below
//     the run epic, where `bd list --parent <run>` returned only the
//     run's direct children.
//   - `--limit 0` because `bd ready`'s own default is 100 and it APPLIES
//     to JSON output. Measured on bd 1.2.2 against this repository's
//     database: `bd ready --json | jq length` → 100, and the same query
//     with `--limit 0` → 106. A truncated ready set is a run reporting
//     less work than it has, at exit 0.
//
// The `bd list` default behaves differently, which is why neither flag
// can be reasoned from the other: `bd list --all --json` and
// `bd list --all --limit 0 --json` both returned 585.
func (t Tracker) Ready(name, scope string, filter tracker.ReadyFilter) plan.Step {
	args := []string{"ready", "--parent", scope}
	if filter.Assignee != "" {
		args = append(args, "--assignee", filter.Assignee)
	}
	if len(filter.ExcludeKinds) > 0 {
		args = append(args, "--exclude-type", strings.Join(filter.ExcludeKinds, ","))
	}
	if len(filter.ExcludeLabels) > 0 {
		args = append(args, "--exclude-label", strings.Join(filter.ExcludeLabels, ","))
	}
	args = append(args, "--limit", "0", "--json")

	s := t.step(name, plan.Read, true, decodeReady, args...)
	s.Description = "the ready set under " + scope
	return s
}

// Holding reads the beads one agent has taken and not finished.
//
// **The scope reaches the answer through the decoded value, not through
// the invocation**, because bd has no transitive parent filter for this
// query: `bd list --parent` answers with DIRECT children only (measured
// on 1.2.2: `bd list --parent <run> --all --limit 0 --json` returned the
// run's seven direct children and none of their children), while the
// beads an agent holds in a run sit two levels down. Narrowing by it
// would drop the run's real work; passing nothing and calling the result
// the run's would pick up another run's. So the widest correct query is
// declared here and [tracker.Held.Within] is what makes the answer the
// run's.
//
// `--status in_progress` is explicit, so `bd list`'s default of hiding
// closed issues cannot change what comes back and `--all` is not needed.
// `--limit 0` is still passed: `bd list`'s default did not truncate JSON
// output when measured, and a query whose completeness rests on that
// staying true is one release away from reporting an agent as drained.
// The kind and label exclusions come from the same [tracker.ReadyFilter]
// the ready half is built from, and are rendered here as `bd list`'s own
// `--exclude-type` and `--exclude-label` (measured on bd 1.2.2:
// `bd list --help` declares both, and
// `bd list --assignee <a> --status in_progress --limit 0 --exclude-type
// epic --exclude-label human --json` exits 0). Without them the drain
// verdict is taken from two sets with different membership rules.
func (t Tracker) Holding(name, scope string, filter tracker.ReadyFilter) plan.Step {
	agent := filter.Assignee
	args := []string{"list", "--assignee", agent, "--status", "in_progress"}
	if len(filter.ExcludeKinds) > 0 {
		args = append(args, "--exclude-type", strings.Join(filter.ExcludeKinds, ","))
	}
	if len(filter.ExcludeLabels) > 0 {
		args = append(args, "--exclude-label", strings.Join(filter.ExcludeLabels, ","))
	}
	args = append(args, "--limit", "0", "--json")

	s := t.step(name, plan.Read, true, holdingDecoder(scope, agent), args...)
	s.Description = "the beads " + agent + " holds"
	return s
}

// ClosableEpics reads which epics the tracker judges eligible to close.
//
// **Also scoped through the decoded value**, and here the trap is
// sharper: bd's epic commands take a positional argument and IGNORE it.
// Measured on 1.2.2 against this repository's database, `bd epic status
// --json | jq length` → 12 and `bd epic status agent-plugins-e6xt --json
// | jq length` → 12 as well, exit 0 both times with nothing on stderr. So
// an invocation that looks scoped is not, and the only honest place to
// scope it is [tracker.Closable.Within].
//
// `--eligible-only` is the tracker's own filter; the decoder pins
// `eligible_for_close` per row as well, so a release in which the flag
// stops filtering cannot get an ineligible epic past `Within`.
func (t Tracker) ClosableEpics(name, scope string) plan.Step {
	s := t.step(name, plan.Read, true, closableDecoder(scope),
		"epic", "status", "--eligible-only", "--json")
	s.Description = "the epics eligible to close under " + scope
	return s
}

// Findings reads the notes recorded on a bead.
//
// `bd comments <id>` rather than a field of `bd show`: measured on 1.2.2,
// `bd show <id> --json` carries a `comment_count` and no `comments` key
// at all, so the notes themselves are only available from this
// subcommand.
func (t Tracker) Findings(name, id string) plan.Step {
	s := t.step(name, plan.Read, true, decodeComments, "comments", id, "--json")
	s.Description = "the findings recorded on " + id
	return s
}

// Claim takes a bead for an agent, setting the state and the assignee in
// one write.
//
// One invocation rather than two, because a bead whose status moved and
// whose assignee did not is claimed by nobody — indistinguishable from
// one nobody has started — and two steps can leave exactly that state
// between them.
func (t Tracker) Claim(name, id, agent string) plan.Step {
	s := t.step(name, plan.Write, true, decodeUpdate,
		"update", id, "--status", "in_progress", "--assignee", agent, "--json")
	s.Description = "claim " + id + " for " + agent
	return s
}

// SetState moves a bead to a state.
//
// A state this tracker cannot express becomes a step that refuses when
// its decoder runs, rather than a silently wrong write: the plan is still
// renderable, and explain shows what would have been attempted.
func (t Tracker) SetState(name, id string, state tracker.State) plan.Step {
	status, err := bdStatusFor(state)
	if err != nil {
		return t.refusing(name, err, "update", id, "--status", string(state), "--json")
	}
	s := t.step(name, plan.Write, true, decodeUpdate,
		"update", id, "--status", status, "--json")
	s.Description = fmt.Sprintf("move %s to %s", id, state)
	return s
}

// AddLabel adds one label to a bead.
//
// `--add-label`, not `--labels`. Measured on bd 1.2.2:
// `bd update <id> --labels probe-xyz --json` exits 1 with an empty stdout
// and `Error: unknown flag: --labels` on stderr, while
// `bd update <id> --add-label <label> --json` exits 0 and returns the
// updated issue as an array. A step naming the wrong flag would fail at
// run time having already printed itself as if it were valid.
func (t Tracker) AddLabel(name, id, label string) plan.Step {
	s := t.step(name, plan.Write, true, decodeUpdate,
		"update", id, "--add-label", label, "--json")
	s.Description = fmt.Sprintf("label %s %q", id, label)
	return s
}

// Note records a comment on a bead.
//
// Repeatable is false: re-running it records the comment twice. That is
// the honest declaration, and it is what tells a recovering agent that
// this step is not one to simply re-run.
func (t Tracker) Note(name, id, body string) plan.Step {
	s := t.step(name, plan.Write, false, decodeComment,
		"comment", id, body, "--json")
	s.Description = "record a note on " + id
	return s
}

// BlockedBy records that bead cannot proceed until blocker is done.
//
// bd's own argument order reads backwards — `bd dep add A B` means "A
// depends on B" — so the mapping from the run's vocabulary to it happens
// exactly once, here, rather than at every call site.
//
// Repeatable: measured on bd 1.2.2, re-adding an edge that already exists
// exits 0 and reports status "added", so re-running is safe.
func (t Tracker) BlockedBy(name, bead, blocker string) plan.Step {
	s := t.step(name, plan.Write, true, decodeDepAdd,
		"dep", "add", bead, blocker, "--type", "blocks", "--json")
	s.Description = fmt.Sprintf("%s waits for %s", bead, blocker)
	return s
}

// DiscoveredFrom records that bead was found while working origin.
func (t Tracker) DiscoveredFrom(name, bead, origin string) plan.Step {
	s := t.step(name, plan.Write, true, decodeDepAdd,
		"dep", "add", bead, origin, "--type", "discovered-from", "--json")
	s.Description = fmt.Sprintf("%s was discovered while working %s", bead, origin)
	return s
}

// WriteMetadata merges a JSON object into a bead's metadata.
//
// Deliberately NOT on [tracker.Tracker]. An arbitrary JSON object
// attached to an issue is a beads feature, not a property of issue
// trackers, and this project has not established which other trackers
// offer one — so the interface carries no metadata operation and a
// caller that needs this reaches the beads implementation by name. That
// is what lets the Beads Manifest Store use it without a second component
// learning the word bd.
//
// Measured on bd 1.2.2: `--metadata` takes the object as an inline JSON
// argument (no file needed), merges at the TOP level so keys it does not
// name survive, and REPLACES a nested value wholesale — writing
// {"k":{"b":2}} over {"k":{"a":1}} leaves {"b":2}. So a caller updating
// one field inside a nested object writes the whole of that object.
//
// Repeatable: writing the same object twice leaves the same record.
func (t Tracker) WriteMetadata(name, id string, document []byte) plan.Step {
	s := t.step(name, plan.Write, true, decodeUpdate,
		"update", id, "--metadata", string(document), "--json")
	s.Description = "merge into " + id + "'s metadata"
	return s
}

// refusing builds a step that renders normally and whose decoder refuses,
// for an operation this tracker cannot express. The plan stays
// explainable, and the refusal names the reason at the point the step
// would have taken effect.
func (t Tracker) refusing(name string, reason error, args ...string) plan.Step {
	s := t.step(name, plan.Write, false, func(res swarmexec.Result) (any, error) {
		return nil, reason
	}, args...)
	s.Description = reason.Error()
	return s
}

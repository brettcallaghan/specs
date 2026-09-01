// Package tracker states what a run needs from an issue tracker, in the
// run's own vocabulary, independent of which tracker implements it — per
// the Tracker component in specs/architecture/container-swarm-cli.md.
//
// Every operation returns a [plan.Step] rather than performing a call.
// That is the shape the container requires of everything reaching an
// external tool: a call made directly is invisible to explain mode and
// unguarded by the plan's ordering policy. A tracker implementation
// therefore *declares* an invocation and how to decode it, and the Plan
// Runner is what runs it.
//
// The vocabulary is deliberately the run's and not any tracker's. A bead
// is claimed, not "in_progress"; a bead is discovered from another, not
// "dep add A B --type discovered-from". What that buys is that a second
// implementation is a matter of satisfying this interface rather than of
// rewriting every caller — and, more immediately, that the direction of
// an edge is stated by the operation's name instead of by an argument
// order a call site can invert.
package tracker

import (
	"fmt"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
)

// State is where a bead stands, in the run's vocabulary.
//
// A tracker's own status strings are its business: an implementation maps
// them onto these, and records what it was told in [Bead.TrackerState] so
// a status this vocabulary does not cover is visible rather than silently
// becoming one that it does.
type State string

const (
	// StateOpen — nobody is working it.
	StateOpen State = "open"
	// StateClaimed — an agent has taken it and is working it.
	StateClaimed State = "claimed"
	// StateBlocked — it cannot proceed.
	StateBlocked State = "blocked"
	// StateDeferred — it has been put off deliberately.
	StateDeferred State = "deferred"
	// StateDone — it is finished.
	StateDone State = "done"
	// StateUnknown — the tracker reported a status this vocabulary does
	// not cover. Never treated as any of the others.
	StateUnknown State = "unknown"
)

// States returns the run's vocabulary of states, [StateUnknown] last.
func States() []State {
	return []State{StateOpen, StateClaimed, StateBlocked, StateDeferred, StateDone, StateUnknown}
}

// Bead is one unit of work, and one vertex in the graph a run dispatches
// from. Another tracker's issue, story or ticket maps onto this.
type Bead struct {
	// ID is the bead's identifier in its tracker.
	ID string
	// Title is its one-line summary.
	Title string
	// State is where it stands, in the run's vocabulary.
	State State
	// TrackerState is the status string the tracker reported, kept so an
	// unmapped status is visible rather than lost behind
	// [StateUnknown].
	TrackerState string
	// Assignee is who holds it — a dispatch id, for a bead in a run.
	Assignee string
	// Kind is the tracker's own type for it (task, epic, gate, …).
	Kind string
	// Priority is the tracker's own priority, lower being higher.
	Priority int
	// Parent is the bead this one hangs under, empty at the top.
	Parent string
	// Labels are its labels.
	Labels []string
	// Metadata is the arbitrary JSON object attached to it. For a run
	// epic this is the run manifest.
	Metadata map[string]any
}

// ReadyFilter narrows a ready-front query. Every field is optional; the
// zero value asks for the whole front under the scope.
type ReadyFilter struct {
	// Assignee narrows to one agent's queue.
	Assignee string
	// ExcludeKinds drops beads of these tracker types — an epic is not
	// work, so a dispatch query excludes it.
	ExcludeKinds []string
	// ExcludeLabels drops beads carrying any of these labels.
	ExcludeLabels []string
}

// Tracker states what a run needs from an issue tracker.
//
// Every method takes a step name first. A plan's steps are named so that
// a later step can refer to what an earlier one produced, and so that a
// report can be read step by step; the tracker cannot choose those names
// because it does not know what else is in the plan.
type Tracker interface {
	// Bead declares a step reading one bead whole, including its
	// metadata. Its decoded value is a [Bead].
	Bead(name, id string) plan.Step

	// Children declares a step reading the beads directly under one bead,
	// whatever state they are in — including finished ones, since a run's
	// completion walk has to see them. Its decoded value is a []Bead.
	//
	// Direct children only. "Every bead in this run" is [Descendants],
	// which walks; see its documentation for why one query cannot answer
	// it.
	Children(name, parent string) plan.Step

	// ReadyFront declares a step reading the beads under a scope that
	// nothing is blocking right now. Its decoded value is a []Bead.
	ReadyFront(name, scope string, filter ReadyFilter) plan.Step

	// Claim declares a step taking a bead for an agent: it sets the state
	// and the assignee together, because a bead claimed by nobody is
	// indistinguishable from one nobody has started.
	Claim(name, id, agent string) plan.Step

	// SetState declares a step moving a bead to a state.
	SetState(name, id string, state State) plan.Step

	// AddLabel declares a step adding one label to a bead.
	AddLabel(name, id, label string) plan.Step

	// Note declares a step recording a comment on a bead. This is the
	// channel a run's findings travel on, so it is an operation of its
	// own rather than a metadata write.
	Note(name, id, body string) plan.Step

	// BlockedBy declares a step recording that bead cannot proceed until
	// blocker is done.
	//
	// The direction is in the name on purpose. A generic Link(from, to)
	// would be one argument order away from inverting every dependency it
	// wrote, at exit 0, which is the failure this container exists to
	// remove.
	BlockedBy(name, bead, blocker string) plan.Step

	// DiscoveredFrom declares a step recording that bead was found while
	// working origin.
	DiscoveredFrom(name, bead, origin string) plan.Step
}

// BeadsFrom reads a step's decoded value as a bead set.
//
// A decoder's output arrives as any, so this is where a caller turns it
// back into something typed. A value of the wrong type is an error rather
// than an empty set: "the step returned something else" and "the step
// returned nothing" are exactly the two conditions this container refuses
// to conflate.
func BeadsFrom(value any) ([]Bead, error) {
	if value == nil {
		return nil, fmt.Errorf("tracker: the step produced no value")
	}
	beads, ok := value.([]Bead)
	if !ok {
		return nil, fmt.Errorf("tracker: the step produced %T, not a bead set", value)
	}
	return beads, nil
}

// BeadFrom reads a step's decoded value as one bead.
func BeadFrom(value any) (Bead, error) {
	if value == nil {
		return Bead{}, fmt.Errorf("tracker: the step produced no value")
	}
	bead, ok := value.(Bead)
	if !ok {
		return Bead{}, fmt.Errorf("tracker: the step produced %T, not a bead", value)
	}
	return bead, nil
}

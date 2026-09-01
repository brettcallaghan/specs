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

// ReadyFilter narrows a ready query. Every field is optional; the
// zero value asks for the whole ready set under the scope.
type ReadyFilter struct {
	// Assignee narrows to one agent's queue.
	Assignee string
	// ExcludeKinds drops beads of these tracker types — an epic is not
	// work, so a dispatch query excludes it.
	ExcludeKinds []string
	// ExcludeLabels drops beads carrying any of these labels.
	ExcludeLabels []string
}

// Finding is one note recorded on a bead — the channel a run's findings
// travel on, read back.
//
// It is stated here rather than in an implementation because a caller
// re-checking a bead's acceptance signals reads findings, and a caller
// above the Beads Tracker may not learn any tracker's spelling for a
// comment.
type Finding struct {
	// ID is the note's own identifier in the tracker.
	ID string
	// Bead is the bead it was recorded on.
	Bead string
	// Author is who recorded it.
	Author string
	// Text is the note itself, unaltered — a run's finding convention
	// lives in this text, and reading it is the caller's.
	Text string
}

// EpicClosure is one epic's closure eligibility, as the TRACKER judges
// it.
//
// Eligible is the tracker's own verdict and is never re-derived here: a
// second eligibility rule living in this container would be a second
// answer to disagree with the first.
type EpicClosure struct {
	// Epic is the epic itself.
	Epic Bead
	// Children is how many beads hang under it.
	Children int
	// Finished is how many of those the tracker counts as complete.
	Finished int
	// Eligible is the tracker's verdict on whether it may close.
	Eligible bool
}

// Closable is what a closure-eligibility query answered, together with
// the scope the caller asked about.
//
// The two are carried together because they are not the same set, and a
// caller that read one for the other would close epics outside its run.
// [Closable.Within] is the only way to get the scoped answer, so the wide
// one cannot be mistaken for it.
type Closable struct {
	// Scope is the run the caller asked about.
	Scope string
	// Judged is every epic the query reported on, which is as narrow as
	// the tracker was able to answer and may be wider than Scope.
	Judged []EpicClosure
}

// Within returns the epics eligible to close that are among beneath —
// the descendant set the caller walked, which is what makes the answer
// the run's rather than the database's.
//
// Only eligible epics come back, so a tracker whose own filter stopped
// filtering still cannot get an ineligible epic past this.
func (c Closable) Within(beneath []Bead) []EpicClosure {
	inside := make(map[string]bool, len(beneath))
	for _, b := range beneath {
		inside[b.ID] = true
	}
	out := make([]EpicClosure, 0, len(c.Judged))
	for _, e := range c.Judged {
		if e.Eligible && inside[e.Epic.ID] {
			out = append(out, e)
		}
	}
	return out
}

// Held is what an assignee query answered: the beads one agent has taken
// and not finished, together with the scope the caller asked about.
//
// Same shape and same reason as [Closable]. A dispatch id is reused
// across runs, so the wide answer read as the narrow one reports an agent
// as holding work that belongs to another run entirely.
type Held struct {
	// Scope is the run the caller asked about.
	Scope string
	// Agent is the dispatch id the query named.
	Agent string
	// Beads is every bead the query reported, which may reach outside
	// Scope.
	Beads []Bead
}

// Within returns the held beads that are among beneath.
func (h Held) Within(beneath []Bead) []Bead {
	inside := make(map[string]bool, len(beneath))
	for _, b := range beneath {
		inside[b.ID] = true
	}
	out := make([]Bead, 0, len(h.Beads))
	for _, b := range h.Beads {
		if inside[b.ID] {
			out = append(out, b)
		}
	}
	return out
}

// Queue is one dispatch's queue beneath a scope, partitioned by whether
// each bead is ready to start or already in progress.
//
// **Both halves are required and neither is sufficient.** A ready query
// excludes work in progress, so on its own it reports an agent mid-bead
// as finished; the in-progress half says nothing about what is waiting.
// So [Queue.Drained] is true only when both are empty, and a queue built
// from one half is not a queue.
type Queue struct {
	// Scope is the run the queue was taken within.
	Scope string
	// Agent is the dispatch id.
	Agent string
	// Ready are the beads assigned to it that nothing is blocking.
	Ready []Bead
	// InProgress are the beads it has claimed.
	InProgress []Bead
}

// Drained is true when the agent holds neither ready nor in-progress
// work.
//
// It says nothing about whether the agent was ever given a queue: a
// dispatch nobody assigned anything to also holds nothing, and telling
// those two apart is the caller's, from whether any bead in the run names
// the agent at all.
func (q Queue) Drained() bool { return len(q.Ready) == 0 && len(q.InProgress) == 0 }

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

	// Ready declares a step reading the beads under a scope that
	// nothing is blocking right now. Its decoded value is a []Bead.
	//
	// The scope is required and is never defaulted: an empty one asks
	// for the whole tracker, and a dispatch set taken over the whole
	// tracker is a run offering work that is not its own.
	Ready(name, scope string, filter ReadyFilter) plan.Step

	// Holding declares a step reading the beads one agent has taken and
	// not finished, beneath a scope. Its decoded value is a [Held],
	// which must be narrowed with [Held.Within] before it is the run's
	// answer.
	//
	// Paired with [Tracker.Ready] filtered to the same agent, this is
	// the partition a [Queue] is built from — the ready half and the
	// in-progress half. Two queries rather than one because a ready
	// query excludes work in progress by construction, so an agent
	// mid-bead has an empty ready set and reads as finished.
	Holding(name, scope, agent string) plan.Step

	// ClosableEpics declares a step reading which epics beneath a scope
	// the tracker judges eligible to close. Its decoded value is a
	// [Closable], which must be narrowed with [Closable.Within] before
	// it is the run's answer.
	//
	// The scope is an argument rather than a default because closing
	// "every eligible epic" is a whole-database act, and a run may only
	// close its own.
	ClosableEpics(name, scope string) plan.Step

	// Findings declares a step reading the notes recorded on a bead.
	// Its decoded value is a []Finding.
	//
	// The read counterpart of [Tracker.Note], and the tracker's half of
	// a bead's acceptance signals: whether the work was recorded, beside
	// the bead's own state.
	Findings(name, id string) plan.Step

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

// HeldFrom reads a step's decoded value as one agent's held set.
func HeldFrom(value any) (Held, error) {
	if value == nil {
		return Held{}, fmt.Errorf("tracker: the step produced no value")
	}
	held, ok := value.(Held)
	if !ok {
		return Held{}, fmt.Errorf("tracker: the step produced %T, not a held set", value)
	}
	return held, nil
}

// ClosableFrom reads a step's decoded value as a closure-eligibility
// answer.
func ClosableFrom(value any) (Closable, error) {
	if value == nil {
		return Closable{}, fmt.Errorf("tracker: the step produced no value")
	}
	closable, ok := value.(Closable)
	if !ok {
		return Closable{}, fmt.Errorf("tracker: the step produced %T, not a closure-eligibility answer", value)
	}
	return closable, nil
}

// FindingsFrom reads a step's decoded value as a bead's findings.
//
// A value of the wrong type is an error rather than an empty set, for the
// same reason [BeadsFrom] refuses one: "the bead recorded nothing" and
// "the read returned something else" are the two conditions an acceptance
// check may not conflate.
func FindingsFrom(value any) ([]Finding, error) {
	if value == nil {
		return nil, fmt.Errorf("tracker: the step produced no value")
	}
	findings, ok := value.([]Finding)
	if !ok {
		return nil, fmt.Errorf("tracker: the step produced %T, not a finding set", value)
	}
	return findings, nil
}

// FindingFrom reads a step's decoded value as one finding — what
// [Tracker.Note] records.
func FindingFrom(value any) (Finding, error) {
	if value == nil {
		return Finding{}, fmt.Errorf("tracker: the step produced no value")
	}
	finding, ok := value.(Finding)
	if !ok {
		return Finding{}, fmt.Errorf("tracker: the step produced %T, not a finding", value)
	}
	return finding, nil
}

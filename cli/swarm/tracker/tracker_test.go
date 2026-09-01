// Tests for the values the dispatch loop's operations decode to. What
// they pin is the narrowing: two of those queries can only be taken
// wider than the run, so a value that could be read as the run's answer
// without being narrowed would close another run's epics or report
// another run's work as this dispatch's.
package tracker_test

import (
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
)

func beadsUnder(ids ...string) []tracker.Bead {
	out := make([]tracker.Bead, 0, len(ids))
	for _, id := range ids {
		out = append(out, tracker.Bead{ID: id})
	}
	return out
}

func eligible(id string, yes bool) tracker.EpicClosure {
	return tracker.EpicClosure{Epic: tracker.Bead{ID: id, Kind: "epic"}, Eligible: yes}
}

// The wide answer holds an epic from another run and an ineligible one.
// Neither may reach a caller that asked about this run.
func TestClosableWithinKeepsOnlyEligibleEpicsInsideTheScope(t *testing.T) {
	closable := tracker.Closable{
		Scope: "run-1",
		Judged: []tracker.EpicClosure{
			eligible("run-1.2", true),
			eligible("run-1.3", false),
			eligible("other-9", true),
		},
	}

	got := closable.Within(beadsUnder("run-1.2", "run-1.3", "run-1.4"))

	if len(got) != 1 {
		t.Fatalf("Within returned %d epics, want 1: %+v", len(got), got)
	}
	if got[0].Epic.ID != "run-1.2" {
		t.Errorf("Within returned %q", got[0].Epic.ID)
	}
}

// An epic the walk did not reach is outside the run whatever the tracker
// said about it, and an empty descendant set narrows to nothing rather
// than to everything.
func TestClosableWithinAnEmptyDescendantSetIsEmpty(t *testing.T) {
	closable := tracker.Closable{Scope: "run-1", Judged: []tracker.EpicClosure{eligible("other-9", true)}}

	if got := closable.Within(nil); len(got) != 0 {
		t.Errorf("Within(nil) returned %d epics; nothing was walked, so nothing is in scope", len(got))
	}
}

// The same for what an agent holds: a dispatch id is reused across runs,
// so the wide answer read as the narrow one reports work that is not this
// run's.
func TestHeldWithinKeepsOnlyTheBeadsInsideTheScope(t *testing.T) {
	held := tracker.Held{
		Scope: "run-1",
		Agent: "dev-1",
		Beads: beadsUnder("run-1.5", "other-9.1"),
	}

	got := held.Within(beadsUnder("run-1.4", "run-1.5"))

	if len(got) != 1 || got[0].ID != "run-1.5" {
		t.Errorf("Within returned %+v, want only run-1.5", got)
	}
}

// Drain takes both halves. The ready half alone reports an agent
// mid-bead as finished, which is the reading that lets a dispatch be
// accepted while its agent is still working.
func TestAQueueIsDrainedOnlyWhenBothHalvesAreEmpty(t *testing.T) {
	cases := []struct {
		name  string
		queue tracker.Queue
		want  bool
	}{
		{"nothing at all", tracker.Queue{Scope: "run-1", Agent: "dev-1"}, true},
		{"one ready", tracker.Queue{Ready: beadsUnder("run-1.5")}, false},
		{"one in progress", tracker.Queue{InProgress: beadsUnder("run-1.5")}, false},
		{"both", tracker.Queue{Ready: beadsUnder("run-1.6"), InProgress: beadsUnder("run-1.5")}, false},
	}

	for _, c := range cases {
		if got := c.queue.Drained(); got != c.want {
			t.Errorf("%s: Drained() = %v, want %v", c.name, got, c.want)
		}
	}
}

// A step's decoded value arrives as any. Reading one as the wrong type is
// an error rather than an empty value, because "the step returned
// something else" and "the step returned nothing" are the two conditions
// this container refuses to conflate.
func TestReadingADecodedValueAsTheWrongThingIsAnError(t *testing.T) {
	readers := map[string]func(any) error{
		"HeldFrom":     func(v any) error { _, err := tracker.HeldFrom(v); return err },
		"ClosableFrom": func(v any) error { _, err := tracker.ClosableFrom(v); return err },
		"FindingsFrom": func(v any) error { _, err := tracker.FindingsFrom(v); return err },
		"FindingFrom":  func(v any) error { _, err := tracker.FindingFrom(v); return err },
	}

	for name, read := range readers {
		if err := read(nil); err == nil {
			t.Errorf("%s(nil) returned no error", name)
		}
		if err := read([]tracker.Bead{}); err == nil {
			t.Errorf("%s([]Bead{}) returned no error", name)
		}
	}
}

func TestReadingADecodedValueAsTheRightThingSucceeds(t *testing.T) {
	if _, err := tracker.HeldFrom(tracker.Held{Agent: "dev-1"}); err != nil {
		t.Errorf("HeldFrom: %v", err)
	}
	if _, err := tracker.ClosableFrom(tracker.Closable{Scope: "run-1"}); err != nil {
		t.Errorf("ClosableFrom: %v", err)
	}
	if _, err := tracker.FindingsFrom([]tracker.Finding{{Bead: "run-1.5"}}); err != nil {
		t.Errorf("FindingsFrom: %v", err)
	}
	// An empty finding set is a real answer: a bead may have recorded
	// nothing, and that is not the same as the read having failed.
	if _, err := tracker.FindingsFrom([]tracker.Finding{}); err != nil {
		t.Errorf("FindingsFrom(empty): %v", err)
	}
	if _, err := tracker.FindingFrom(tracker.Finding{Bead: "run-1.5"}); err != nil {
		t.Errorf("FindingFrom: %v", err)
	}
}

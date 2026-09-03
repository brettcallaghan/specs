// Tests for the decoders. Each one pins the result shapes its subcommand
// is known to return, including its empty and failure forms, because the
// whole reason there is a decoder per subcommand is that those shapes do
// not agree with each other.
//
// The bytes below are the ones measured on bd 1.2.2 against this
// repository's own database — recorded here so a bd release that changes
// a shape fails a test rather than returning a plausible empty answer at
// exit 0.
package beads

import (
	"errors"
	"strings"
	"testing"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
)

func result(stdout, stderr string, status int) swarmexec.Result {
	return swarmexec.Result{Stdout: []byte(stdout), Stderr: []byte(stderr), ExitStatus: status}
}

// oneIssue is the shape bd returns for an issue, with the field names
// measured from `bd show agent-plugins-ih18.5 --json`.
const oneIssue = `[{
  "id": "run-1.4",
  "title": "Command Surface",
  "status": "in_progress",
  "assignee": "dev-1",
  "issue_type": "task",
  "priority": 1,
  "parent": "run-1",
  "labels": ["delta:SURF"],
  "metadata": {"tier": "standard", "est_size": "M"}
}]`

// bdWarning is what bd writes to stderr while its JSON goes to stdout. A
// decoder that read a merged stream would find it in the middle of the
// payload.
const bdWarning = "warning: beads.role not configured (GH#2950).\n"

func TestDecodeShowReturnsOneBeadInTheRunsVocabulary(t *testing.T) {
	value, err := decodeShow(result(oneIssue, bdWarning, 0))
	if err != nil {
		t.Fatalf("decodeShow: %v", err)
	}

	bead, err := tracker.BeadFrom(value)
	if err != nil {
		t.Fatalf("BeadFrom: %v", err)
	}
	if bead.ID != "run-1.4" {
		t.Errorf("ID = %q", bead.ID)
	}
	if bead.State != tracker.StateClaimed {
		t.Errorf("State = %q, want %q — in_progress is claimed in the run's vocabulary", bead.State, tracker.StateClaimed)
	}
	if bead.TrackerState != "in_progress" {
		t.Errorf("TrackerState = %q, want the tracker's own string", bead.TrackerState)
	}
	if bead.Parent != "run-1" {
		t.Errorf("Parent = %q — the parent is exposed under \"parent\"", bead.Parent)
	}
	if bead.Metadata["tier"] != "standard" {
		t.Errorf("Metadata = %v", bead.Metadata)
	}
	if len(bead.Labels) != 1 || bead.Labels[0] != "delta:SURF" {
		t.Errorf("Labels = %v", bead.Labels)
	}
}

// Measured: `bd show <missing> --json` exits 1 and puts an OBJECT with an
// "error" key on STDOUT — the one subcommand here whose message is not on
// stderr. A decoder assuming stderr would report "said nothing".
func TestDecodeShowReadsTheFailureMessageFromStandardOutput(t *testing.T) {
	const body = `{"error": "no issues found matching the provided IDs", "schema_version": 1}`

	_, err := decodeShow(result(body, bdWarning, 1))
	if err == nil {
		t.Fatal("decodeShow returned no error for a failing result")
	}
	if !strings.Contains(err.Error(), "no issues found") {
		t.Errorf("the error does not carry what bd said: %v", err)
	}
}

// Measured: `bd update <missing> … --json` exits 1 with STDOUT EMPTY and
// the message on stderr — the opposite of bd show. This is the pair that
// makes one shared failure guard wrong.
func TestDecodeUpdateReadsTheFailureMessageFromStandardError(t *testing.T) {
	const stderr = bdWarning + `Error resolving nonexistent-xyz: no issue found matching "nonexistent-xyz"` + "\n"

	_, err := decodeUpdate(result("", stderr, 1))
	if err == nil {
		t.Fatal("decodeUpdate returned no error for a failing result")
	}
	if !strings.Contains(err.Error(), "no issue found matching") {
		t.Errorf("the error does not carry what bd said: %v", err)
	}
	// The warning bd emits on every invocation must not be mistaken for
	// the message.
	if strings.Contains(err.Error(), "beads.role") {
		t.Errorf("the error reports bd's standing warning instead of the failure: %v", err)
	}
}

// The three forms that mean "nothing" normalize to one empty value, and
// that value is a non-nil, zero-length bead set — so a caller can range
// over it without a nil check and cannot mistake it for "not asked".
func TestTheListDecoderNormalizesAbsentNullAndEmptyToOneEmptyValue(t *testing.T) {
	for name, stdout := range map[string]string{
		"absent":     "",
		"bare null":  "null",
		"empty list": "[]",
	} {
		value, err := decodeList(result(stdout, bdWarning, 0))
		if err != nil {
			t.Errorf("%s: decodeList: %v", name, err)
			continue
		}
		beads, err := tracker.BeadsFrom(value)
		if err != nil {
			t.Errorf("%s: BeadsFrom: %v", name, err)
			continue
		}
		if beads == nil {
			t.Errorf("%s: the empty value is nil rather than an empty set", name)
		}
		if len(beads) != 0 {
			t.Errorf("%s: got %d beads, want 0", name, len(beads))
		}
	}
}

// The bare-null case is not hypothetical: measured on bd 1.2.2,
// `bd gate list --json` returns a bare null when there are no gates, at
// exit 0, and `jq length` on it returns 0 — so a count fails quietly and
// an iteration fails loudly.
func TestTheReadyDecoderNormalizesTheSameThreeForms(t *testing.T) {
	for _, stdout := range []string{"", "null", "[]"} {
		value, err := decodeReady(result(stdout, "", 0))
		if err != nil {
			t.Errorf("%q: decodeReady: %v", stdout, err)
			continue
		}
		beads, _ := tracker.BeadsFrom(value)
		if len(beads) != 0 {
			t.Errorf("%q: got %d beads, want 0", stdout, len(beads))
		}
	}
}

// A shape the decoder does not know is an error, not evidence of
// emptiness. The family's old prose rule — "count only a JSON array and
// treat any other payload as empty" — is what this refuses: measured,
// `bd stats --json` and `bd swarm list --json` legitimately return
// objects, and that rule reports every run as absent.
func TestAnUnknownShapeIsAnErrorRatherThanAnEmptySet(t *testing.T) {
	for name, stdout := range map[string]string{
		"an object":    `{"schema_version": 1, "summary": {"closed_issues": 418}}`,
		"a number":     `7`,
		"loose text":   `no issues found`,
		"a string":     `"nothing"`,
		"a bare true":  `true`,
		"a truncation": `[{"id": "run-1"`,
	} {
		if _, err := decodeList(result(stdout, "", 0)); err == nil {
			t.Errorf("%s: decodeList reported it as an empty set rather than an unknown shape", name)
		}
	}
}

// A status this vocabulary does not cover becomes StateUnknown and keeps
// the tracker's own string. It must never become one of the others: a bd
// release that adds a status would otherwise make every bead carrying it
// read as open, which is a difference a dispatch decision turns on.
func TestAnUnmappedStatusBecomesUnknownAndKeepsTheTrackersString(t *testing.T) {
	const body = `[{"id": "run-1.9", "status": "quarantined"}]`

	value, err := decodeList(result(body, "", 0))
	if err != nil {
		t.Fatalf("decodeList: %v", err)
	}
	beads, _ := tracker.BeadsFrom(value)
	if len(beads) != 1 {
		t.Fatalf("got %d beads", len(beads))
	}
	if beads[0].State != tracker.StateUnknown {
		t.Errorf("State = %q, want %q", beads[0].State, tracker.StateUnknown)
	}
	if beads[0].TrackerState != "quarantined" {
		t.Errorf("TrackerState = %q, want the status bd reported", beads[0].TrackerState)
	}
}

func TestEveryStatusBdReportsMapsOntoTheRunsVocabulary(t *testing.T) {
	want := map[string]tracker.State{
		"open":        tracker.StateOpen,
		"in_progress": tracker.StateClaimed,
		"blocked":     tracker.StateBlocked,
		"deferred":    tracker.StateDeferred,
		"closed":      tracker.StateDone,
	}
	for status, state := range want {
		if got := bdStatus[status]; got != state {
			t.Errorf("bd status %q maps to %q, want %q", status, got, state)
		}
	}
	// And back again, for the states a caller can move a bead to.
	for _, state := range []tracker.State{
		tracker.StateOpen, tracker.StateClaimed, tracker.StateBlocked,
		tracker.StateDeferred, tracker.StateDone,
	} {
		if _, err := bdStatusFor(state); err != nil {
			t.Errorf("bdStatusFor(%q): %v", state, err)
		}
	}
	if _, err := bdStatusFor(tracker.StateUnknown); err == nil {
		t.Error("bdStatusFor(unknown) returned a status; there is none to write")
	}
}

// Measured: `bd comment <id> <text> --json` returns an OBJECT, not an
// array. A guard treating any non-array as empty would report a comment
// that was written as one that was not.
func TestDecodeCommentReadsTheObjectBdReturns(t *testing.T) {
	const body = `{
  "author": "dev-1",
  "created_at": "2026-08-29T02:03:01.256149265Z",
  "id": "01a04b41-890a-797d-9d91-40b417ad3f14",
  "issue_id": "run-1.5",
  "schema_version": 1,
  "text": "retro:proof measured"
}`

	value, err := decodeComment(result(body, "", 0))
	if err != nil {
		t.Fatalf("decodeComment: %v", err)
	}
	note, err := tracker.FindingFrom(value)
	if err != nil {
		t.Fatalf("FindingFrom: %v", err)
	}
	if note.Bead != "run-1.5" || note.Author != "dev-1" {
		t.Errorf("note = %+v", note)
	}
	if !strings.Contains(note.Text, "retro:proof") {
		t.Errorf("the note's text was not carried: %q", note.Text)
	}

	// Exit 0 with nothing to show for it is a failure, not a success.
	if _, err := decodeComment(result("", "", 0)); err == nil {
		t.Error("decodeComment accepted an empty result at exit 0")
	}
}

// Measured: `bd dep add A B --type blocks --json` returns an OBJECT, and
// re-adding an edge that already exists still exits 0 with
// status:"added" — which is what makes the write safe to re-run.
func TestDecodeDepAddReadsTheObjectBdReturns(t *testing.T) {
	const body = `{
  "depends_on_id": "run-1.2",
  "issue_id": "run-1.5",
  "schema_version": 1,
  "status": "added",
  "type": "blocks"
}`

	value, err := decodeDepAdd(result(body, "", 0))
	if err != nil {
		t.Fatalf("decodeDepAdd: %v", err)
	}
	edge, ok := value.(Edge)
	if !ok {
		t.Fatalf("decodeDepAdd produced %T", value)
	}
	if edge.Bead != "run-1.5" || edge.Blocker != "run-1.2" {
		t.Errorf("edge = %+v — the bead depends on the blocker, in that direction", edge)
	}
	if edge.Kind != "blocks" || edge.Outcome != "added" {
		t.Errorf("edge = %+v", edge)
	}

	// Measured failure shape: an object with an "error" key on stdout.
	const failed = `{"error": "resolving issue ID nonexistent-xyz: no issue found matching \"nonexistent-xyz\"", "schema_version": 1}`
	_, err = decodeDepAdd(result(failed, bdWarning, 1))
	if err == nil {
		t.Fatal("decodeDepAdd returned no error for a failing result")
	}
	if !strings.Contains(err.Error(), "no issue found matching") {
		t.Errorf("the error does not carry what bd said: %v", err)
	}
}

// A failing invocation that says nothing at all still produces an error
// naming the subcommand and the status, rather than an empty success.
func TestAFailureWithNoMessageIsStillAFailure(t *testing.T) {
	_, err := decodeList(result("", "", 2))
	if err == nil {
		t.Fatal("decodeList accepted a result at exit 2")
	}
	if !strings.Contains(err.Error(), "bd list") || !strings.Contains(err.Error(), "2") {
		t.Errorf("the error names neither the subcommand nor the status: %v", err)
	}
	if errors.Unwrap(err) != nil {
		t.Logf("wrapped: %v", errors.Unwrap(err))
	}
}

// --- the dispatch loop's operations ---------------------------------

// Measured: `bd comments <id> --json` returns an ARRAY of the objects
// `bd comment` returns one of. One letter apart from the write, one
// container different, so the two do not share a decoder.
func TestDecodeCommentsReadsTheArrayBdReturns(t *testing.T) {
	const body = `[
  {
    "author": "dev-e6xt-b1",
    "created_at": "2026-09-01T17:20:11.123456789Z",
    "id": "01a039db-0ab7-7243-977a-2c9aac0b1c33",
    "issue_id": "run-1.5",
    "text": "retro:proof measured against bd 1.2.2"
  },
  {
    "author": "dev-e6xt-b1",
    "created_at": "2026-09-01T17:21:00.000000000Z",
    "id": "01a039db-0ab7-7243-977a-2c9aac0b1c34",
    "issue_id": "run-1.5",
    "text": "retro:decision the wide query is narrowed by the caller"
  }
]`

	value, err := decodeComments(result(body, bdWarning, 0))
	if err != nil {
		t.Fatalf("decodeComments: %v", err)
	}
	findings, err := tracker.FindingsFrom(value)
	if err != nil {
		t.Fatalf("FindingsFrom: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("decoded %d findings, want 2", len(findings))
	}
	if findings[0].Bead != "run-1.5" || findings[0].Author != "dev-e6xt-b1" {
		t.Errorf("finding = %+v", findings[0])
	}
	if !strings.Contains(findings[1].Text, "retro:decision") {
		t.Errorf("the finding's text was not carried: %q", findings[1].Text)
	}
}

// A bead carrying no comments answers `[]`, which is a real answer and
// decodes to an empty set. A shape that is not an array is not: "nothing
// was recorded" and "the read returned something else" are what an
// acceptance check exists to keep apart.
func TestDecodeCommentsPinsItsEmptyAndUnknownForms(t *testing.T) {
	for _, empty := range []string{"[]", "", "null"} {
		value, err := decodeComments(result(empty, bdWarning, 0))
		if err != nil {
			t.Fatalf("decodeComments(%q): %v", empty, err)
		}
		findings, err := tracker.FindingsFrom(value)
		if err != nil {
			t.Fatalf("FindingsFrom(%q): %v", empty, err)
		}
		if findings == nil {
			t.Errorf("decodeComments(%q) produced a nil set; the empty set is not nothing", empty)
		}
		if len(findings) != 0 {
			t.Errorf("decodeComments(%q) produced %d findings", empty, len(findings))
		}
	}

	// An error body at exit 0 is not an empty comment list.
	if _, err := decodeComments(result(`{"error": "no issues found"}`, "", 0)); err == nil {
		t.Error("decodeComments read an error object as a comment list")
	}

	if _, err := decodeComments(result("", "boom", 1)); err == nil {
		t.Error("decodeComments accepted a result at exit 1")
	}

	// Measured: `bd comments <missing> --json` exits 1 with the message
	// in an OBJECT on STDOUT, like bd show and unlike bd update.
	const missing = `{
  "error": "resolving nonexistent-xyz: no issue found matching \"nonexistent-xyz\"",
  "schema_version": 1
}`
	err := func() error { _, err := decodeComments(result(missing, bdWarning, 1)); return err }()
	if err == nil {
		t.Fatal("decodeComments returned no error for a failing result")
	}
	if !strings.Contains(err.Error(), "no issue found matching") {
		t.Errorf("the error does not carry what bd said: %v", err)
	}
	if strings.Contains(err.Error(), "beads.role") {
		t.Errorf("the error reports bd's standing warning instead of the failure: %v", err)
	}
}

// eligibleEpics is the shape measured from `bd epic status --json`: an
// array of rows, each wrapping the same issue object every other
// subcommand returns.
const eligibleEpics = `[
  {
    "epic": {"id": "run-1.2", "title": "Phase 2", "status": "open", "issue_type": "epic", "priority": 1},
    "total_children": 4,
    "closed_children": 4,
    "eligible_for_close": true
  },
  {
    "epic": {"id": "other-9", "title": "Another run's phase", "status": "open", "issue_type": "epic", "priority": 2},
    "total_children": 2,
    "closed_children": 2,
    "eligible_for_close": true
  }
]`

func TestDecodeEpicStatusReadsTheRowsBdReturns(t *testing.T) {
	value, err := closableDecoder("run-1")(result(eligibleEpics, bdWarning, 0))
	if err != nil {
		t.Fatalf("closableDecoder: %v", err)
	}
	closable, err := tracker.ClosableFrom(value)
	if err != nil {
		t.Fatalf("ClosableFrom: %v", err)
	}
	if closable.Scope != "run-1" {
		t.Errorf("Scope = %q — the answer carries the scope it was asked about", closable.Scope)
	}
	if len(closable.Judged) != 2 {
		t.Fatalf("decoded %d rows, want 2", len(closable.Judged))
	}
	first := closable.Judged[0]
	if first.Epic.ID != "run-1.2" || first.Epic.Kind != "epic" {
		t.Errorf("epic = %+v", first.Epic)
	}
	if first.Children != 4 || first.Finished != 4 || !first.Eligible {
		t.Errorf("closure = %+v", first)
	}
}

// Nothing eligible answers `[]`, which is an answer and not a failure.
func TestDecodeEpicStatusPinsItsEmptyAndUnknownForms(t *testing.T) {
	for _, empty := range []string{"[]", "", "null"} {
		value, err := closableDecoder("run-1")(result(empty, bdWarning, 0))
		if err != nil {
			t.Fatalf("closableDecoder(%q): %v", empty, err)
		}
		closable, err := tracker.ClosableFrom(value)
		if err != nil {
			t.Fatalf("ClosableFrom(%q): %v", empty, err)
		}
		if closable.Judged == nil {
			t.Errorf("closableDecoder(%q) produced a nil set", empty)
		}
		if len(closable.Judged) != 0 {
			t.Errorf("closableDecoder(%q) judged %d epics", empty, len(closable.Judged))
		}
	}

	// A row naming no epic is a shape this decoder does not know, not an
	// epic with an empty id that a later close would be pointed at.
	if _, err := closableDecoder("run-1")(result(`[{"total_children": 1}]`, "", 0)); err == nil {
		t.Error("closableDecoder accepted a row naming no epic")
	}
	if _, err := closableDecoder("run-1")(result(`{"error": "nope"}`, "", 0)); err == nil {
		t.Error("closableDecoder read an error object as an eligibility answer")
	}
	if _, err := closableDecoder("run-1")(result("", "boom", 1)); err == nil {
		t.Error("closableDecoder accepted a result at exit 1")
	}
}

// The held set carries the scope and the agent it was asked about, and
// the beads exactly as `bd list` reported them — narrowing is
// [tracker.Held.Within]'s, not the decoder's.
func TestHoldingDecoderCarriesTheScopeAndTheAgent(t *testing.T) {
	const body = `[
  {"id": "run-1.5", "title": "TRK", "status": "in_progress", "assignee": "dev-1", "issue_type": "task", "parent": "run-1.2"},
  {"id": "other-9.1", "title": "someone else's run", "status": "in_progress", "assignee": "dev-1", "issue_type": "task", "parent": "other-9"}
]`

	value, err := holdingDecoder("run-1", "dev-1")(result(body, bdWarning, 0))
	if err != nil {
		t.Fatalf("holdingDecoder: %v", err)
	}
	held, err := tracker.HeldFrom(value)
	if err != nil {
		t.Fatalf("HeldFrom: %v", err)
	}
	if held.Scope != "run-1" || held.Agent != "dev-1" {
		t.Errorf("held = {Scope:%q Agent:%q}", held.Scope, held.Agent)
	}
	if len(held.Beads) != 2 {
		t.Fatalf("decoded %d beads, want both — the decoder narrows nothing", len(held.Beads))
	}
	if held.Beads[0].State != tracker.StateClaimed {
		t.Errorf("State = %q, want %q", held.Beads[0].State, tracker.StateClaimed)
	}
}

func TestHoldingDecoderPinsItsEmptyAndUnknownForms(t *testing.T) {
	for _, empty := range []string{"[]", "", "null"} {
		value, err := holdingDecoder("run-1", "dev-1")(result(empty, bdWarning, 0))
		if err != nil {
			t.Fatalf("holdingDecoder(%q): %v", empty, err)
		}
		held, err := tracker.HeldFrom(value)
		if err != nil {
			t.Fatalf("HeldFrom(%q): %v", empty, err)
		}
		if held.Beads == nil {
			t.Errorf("holdingDecoder(%q) produced a nil set", empty)
		}
		if len(held.Beads) != 0 {
			t.Errorf("holdingDecoder(%q) held %d beads", empty, len(held.Beads))
		}
	}

	// The one that matters: an error body is not a queue of length one.
	value, err := holdingDecoder("run-1", "dev-1")(result(`{"error": "no issues found"}`, "", 0))
	if err == nil {
		t.Fatalf("holdingDecoder read an error object as a queue: %#v", value)
	}
	if _, err := holdingDecoder("run-1", "dev-1")(result("", "boom", 1)); err == nil {
		t.Error("holdingDecoder accepted a result at exit 1")
	}
}

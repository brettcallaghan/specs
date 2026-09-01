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
	note, ok := value.(Note)
	if !ok {
		t.Fatalf("decodeComment produced %T", value)
	}
	if note.IssueID != "run-1.5" || note.Author != "dev-1" {
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

package beads

import (
	"encoding/json"
	"fmt"
	"strings"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
)

// One decoder per subcommand. Each knows that subcommand's result shapes
// — including its empty and failure forms — and nothing about any other's.
//
// This is not defensive programming spread thin; it is where a measured
// non-uniformity is confined. Measured on bd 1.2.2 against this
// repository's own database:
//
//	bd show <id> --json          -> a JSON ARRAY of one issue
//	bd show <missing> --json     -> exit 1, an OBJECT {"error":…} on STDOUT
//	bd list --parent X --json    -> a JSON array; empty is []
//	bd ready --parent X --json   -> a JSON array; empty is []
//	bd update <id> … --json      -> a JSON array of the updated issue
//	bd update <missing> … --json -> exit 1, STDOUT EMPTY, message on stderr
//	bd comment <id> … --json     -> an OBJECT {author,created_at,id,issue_id,text,…}
//	bd comments <id> --json      -> a JSON ARRAY of those objects; empty is []
//	bd comments <missing> --json -> exit 1, an OBJECT {"error":…} on STDOUT
//	bd dep add A B --json        -> an OBJECT {issue_id,depends_on_id,type,status}
//	bd dep add <missing> --json  -> exit 1, an OBJECT {"error":…} on STDOUT
//	bd gate list --json          -> a JSON array; empty is a bare NULL
//	bd epic status --json        -> a JSON array of {epic,total_children,
//	                                closed_children,eligible_for_close};
//	                                empty is []
//
// `bd comment` and `bd comments` differ by one letter and by their
// container: the write returns the object, the read returns an array of
// them. That is why each has its own decoder rather than sharing one.
//
// Four container shapes and two different failure shapes across the set,
// all but the failures at exit 0. A generic "treat any non-array as
// empty" guard — which the family's prose used to prescribe — is
// measurably wrong for the object-returning subcommands: it reports every
// run as absent. And a generic "non-zero means read stderr" guard is
// wrong for `bd show`, whose message is on stdout.
//
// bd also writes unrelated warnings to stderr while its JSON goes to
// stdout ("warning: beads.role not configured"), which is why nothing
// here ever reads a merged stream.

// issueJSON is bd's own issue shape, named as bd names it. It is confined
// to this file: everything above sees a [tracker.Bead].
type issueJSON struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Status   string         `json:"status"`
	Assignee string         `json:"assignee"`
	Kind     string         `json:"issue_type"`
	Priority int            `json:"priority"`
	Parent   string         `json:"parent"`
	Labels   []string       `json:"labels"`
	Metadata map[string]any `json:"metadata"`
}

// errorJSON is the object bd puts on STDOUT when a subcommand fails with
// a structured error. Not every failing subcommand emits one — `bd
// update` leaves stdout empty — so a decoder that assumed it would find
// nothing to report.
type errorJSON struct {
	Error string `json:"error"`
}

// bdStatus maps bd's status strings onto the run's vocabulary. A status
// absent from this table becomes [tracker.StateUnknown] rather than any
// of the others, and the raw string is kept on the bead: a bd release
// that adds a status must not make every bead carrying it read as open.
var bdStatus = map[string]tracker.State{
	"open":        tracker.StateOpen,
	"in_progress": tracker.StateClaimed,
	"blocked":     tracker.StateBlocked,
	"deferred":    tracker.StateDeferred,
	"closed":      tracker.StateDone,
}

// bdStatusFor is the reverse: the string to hand `bd update --status`.
func bdStatusFor(state tracker.State) (string, error) {
	for raw, mapped := range bdStatus {
		if mapped == state {
			return raw, nil
		}
	}
	return "", fmt.Errorf("beads: %q is not a state this tracker can be moved to", state)
}

func beadFrom(issue issueJSON) tracker.Bead {
	state, ok := bdStatus[issue.Status]
	if !ok {
		state = tracker.StateUnknown
	}
	return tracker.Bead{
		ID:           issue.ID,
		Title:        issue.Title,
		State:        state,
		TrackerState: issue.Status,
		Assignee:     issue.Assignee,
		Kind:         issue.Kind,
		Priority:     issue.Priority,
		Parent:       issue.Parent,
		Labels:       issue.Labels,
		Metadata:     issue.Metadata,
	}
}

// failure turns a non-zero result into an error, reading the message from
// wherever the subcommand puts it. Both places are tried because the
// measurements above show both in use.
func failure(subcommand string, res swarmexec.Result) error {
	var structured errorJSON
	if err := json.Unmarshal(res.Stdout, &structured); err == nil && structured.Error != "" {
		return fmt.Errorf("beads: %s exited %d: %s", subcommand, res.ExitStatus, structured.Error)
	}
	if text := strings.TrimSpace(string(res.Stderr)); text != "" {
		return fmt.Errorf("beads: %s exited %d: %s", subcommand, res.ExitStatus, lastLine(text))
	}
	return fmt.Errorf("beads: %s exited %d and said nothing", subcommand, res.ExitStatus)
}

// lastLine is the final non-empty line of text. bd prefixes its stderr
// with the "beads.role not configured" warning, so the first line is
// usually the warning and the last is the message.
func lastLine(text string) string {
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if trimmed := strings.TrimSpace(lines[i]); trimmed != "" {
			return trimmed
		}
	}
	return text
}

// issueArray parses the array-of-issues body several subcommands return,
// normalizing this subcommand's absent, null and empty forms to one empty
// value: a non-nil, zero-length bead set.
//
// Absent, null and [] are the three that mean "nothing", and they are
// normalized together. Anything else — an object, a number, a fragment of
// text — is an error, because a shape this decoder does not know is not
// evidence of emptiness.
func issueArray(subcommand string, res swarmexec.Result) ([]tracker.Bead, error) {
	body := strings.TrimSpace(string(res.Stdout))
	if body == "" || body == "null" {
		return []tracker.Bead{}, nil
	}

	var issues []issueJSON
	if err := json.Unmarshal([]byte(body), &issues); err != nil {
		return nil, fmt.Errorf("beads: %s returned a shape this decoder does not know: %w", subcommand, err)
	}

	beads := make([]tracker.Bead, 0, len(issues))
	for _, issue := range issues {
		beads = append(beads, beadFrom(issue))
	}
	return beads, nil
}

// --- one decoder per subcommand -------------------------------------

// decodeShow decodes `bd show <id> --json`.
//
// Measured: an array of one issue at exit 0, and an object carrying an
// "error" key on STDOUT at exit 1 — the one subcommand here whose failure
// message is not on stderr.
func decodeShow(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("bd show", res)
	}
	beads, err := issueArray("bd show", res)
	if err != nil {
		return nil, err
	}
	switch len(beads) {
	case 1:
		return beads[0], nil
	case 0:
		return nil, fmt.Errorf("beads: bd show returned no issue at exit 0")
	default:
		return nil, fmt.Errorf("beads: bd show returned %d issues for one id", len(beads))
	}
}

// decodeList decodes `bd list --parent <id> --json`. Empty is [].
func decodeList(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("bd list", res)
	}
	return issueArray("bd list", res)
}

// decodeReady decodes `bd ready --parent <id> --json`. Empty is [].
func decodeReady(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("bd ready", res)
	}
	return issueArray("bd ready", res)
}

// decodeUpdate decodes `bd update <id> … --json`.
//
// Measured: an array holding the updated issue at exit 0, and at exit 1
// an EMPTY stdout with the message on stderr — the opposite of `bd show`,
// which is why the two have separate decoders rather than one.
func decodeUpdate(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("bd update", res)
	}
	beads, err := issueArray("bd update", res)
	if err != nil {
		return nil, err
	}
	if len(beads) != 1 {
		return nil, fmt.Errorf("beads: bd update returned %d issues, want the one it updated", len(beads))
	}
	return beads[0], nil
}

// commentJSON is bd's own comment shape, named as bd names it. Confined
// to this file: everything above sees a [tracker.Finding].
type commentJSON struct {
	ID      string `json:"id"`
	IssueID string `json:"issue_id"`
	Author  string `json:"author"`
	Text    string `json:"text"`
}

func findingFrom(c commentJSON) tracker.Finding {
	return tracker.Finding{ID: c.ID, Bead: c.IssueID, Author: c.Author, Text: c.Text}
}

// decodeComment decodes `bd comment <id> <text> --json`.
//
// Measured: an OBJECT, not an array — {author, created_at, id, issue_id,
// schema_version, text}. A guard that treated any non-array as empty
// would report a comment that was written as one that was not.
func decodeComment(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("bd comment", res)
	}
	body := strings.TrimSpace(string(res.Stdout))
	if body == "" || body == "null" {
		return nil, fmt.Errorf("beads: bd comment exited 0 without saying what it recorded")
	}
	var comment commentJSON
	if err := json.Unmarshal([]byte(body), &comment); err != nil {
		return nil, fmt.Errorf("beads: bd comment returned a shape this decoder does not know: %w", err)
	}
	if comment.IssueID == "" {
		return nil, fmt.Errorf("beads: bd comment exited 0 but named no issue")
	}
	return findingFrom(comment), nil
}

// decodeComments decodes `bd comments <id> --json`.
//
// Measured on bd 1.2.2: an ARRAY of the same objects `bd comment`
// returns one of, and `[]` for a bead carrying none. The empty case is a
// real answer — a bead may legitimately have recorded nothing — so it
// decodes to an empty set rather than to an error; a shape that is not an
// array does not, because "nothing was recorded" and "the read returned
// something else" are what an acceptance check exists to keep apart.
func decodeComments(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("bd comments", res)
	}
	body := strings.TrimSpace(string(res.Stdout))
	if body == "" || body == "null" {
		return []tracker.Finding{}, nil
	}

	var comments []commentJSON
	if err := json.Unmarshal([]byte(body), &comments); err != nil {
		return nil, fmt.Errorf("beads: bd comments returned a shape this decoder does not know: %w", err)
	}

	findings := make([]tracker.Finding, 0, len(comments))
	for _, c := range comments {
		findings = append(findings, findingFrom(c))
	}
	return findings, nil
}

// epicStatusJSON is one row of `bd epic status --json`, named as bd names
// it. The eligibility verdict is bd's own field: nothing here recomputes
// it from the two counts, which would be a second rule to disagree with
// the first.
type epicStatusJSON struct {
	Epic     issueJSON `json:"epic"`
	Children int       `json:"total_children"`
	Closed   int       `json:"closed_children"`
	Eligible bool      `json:"eligible_for_close"`
}

// closableDecoder decodes `bd epic status --eligible-only --json`,
// carrying the scope the caller asked about into the answer so the two
// cannot be confused. See [Tracker.ClosableEpics] for why the invocation
// itself cannot be scoped.
//
// Measured on bd 1.2.2: an ARRAY of {epic, total_children,
// closed_children, eligible_for_close} where `epic` is the same issue
// shape every other subcommand returns, and `[]` when nothing is
// eligible.
func closableDecoder(scope string) plan.Decoder {
	return func(res swarmexec.Result) (any, error) {
		if !res.Succeeded() {
			return nil, failure("bd epic status", res)
		}
		body := strings.TrimSpace(string(res.Stdout))
		if body == "" || body == "null" {
			return tracker.Closable{Scope: scope, Judged: []tracker.EpicClosure{}}, nil
		}

		var rows []epicStatusJSON
		if err := json.Unmarshal([]byte(body), &rows); err != nil {
			return nil, fmt.Errorf("beads: bd epic status returned a shape this decoder does not know: %w", err)
		}

		judged := make([]tracker.EpicClosure, 0, len(rows))
		for _, r := range rows {
			if r.Epic.ID == "" {
				return nil, fmt.Errorf("beads: bd epic status reported a row naming no epic")
			}
			judged = append(judged, tracker.EpicClosure{
				Epic:     beadFrom(r.Epic),
				Children: r.Children,
				Finished: r.Closed,
				Eligible: r.Eligible,
			})
		}
		return tracker.Closable{Scope: scope, Judged: judged}, nil
	}
}

// holdingDecoder decodes the `bd list` behind [Tracker.Holding], carrying
// the scope and the agent into the answer for the same reason.
func holdingDecoder(scope, agent string) plan.Decoder {
	return func(res swarmexec.Result) (any, error) {
		if !res.Succeeded() {
			return nil, failure("bd list", res)
		}
		beads, err := issueArray("bd list", res)
		if err != nil {
			return nil, err
		}
		return tracker.Held{Scope: scope, Agent: agent, Beads: beads}, nil
	}
}

// Edge is what `bd dep add` records.
type Edge struct {
	// Bead is the bead that now depends on Blocker.
	Bead string `json:"issue_id"`
	// Blocker is what it depends on.
	Blocker string `json:"depends_on_id"`
	// Kind is the tracker's own edge type.
	Kind string `json:"type"`
	// Outcome is what the tracker did — "added".
	Outcome string `json:"status"`
}

// decodeDepAdd decodes `bd dep add <bead> <blocker> --type <kind> --json`.
//
// Measured: an OBJECT {issue_id, depends_on_id, type, status:"added",
// schema_version} at exit 0 — including when the edge was already there,
// which is what makes the write repeatable — and an object carrying an
// "error" key on STDOUT at exit 1.
func decodeDepAdd(res swarmexec.Result) (any, error) {
	if !res.Succeeded() {
		return nil, failure("bd dep add", res)
	}
	body := strings.TrimSpace(string(res.Stdout))
	if body == "" || body == "null" {
		return nil, fmt.Errorf("beads: bd dep add exited 0 without saying what it recorded")
	}
	var edge Edge
	if err := json.Unmarshal([]byte(body), &edge); err != nil {
		return nil, fmt.Errorf("beads: bd dep add returned a shape this decoder does not know: %w", err)
	}
	if edge.Bead == "" || edge.Blocker == "" {
		return nil, fmt.Errorf("beads: bd dep add exited 0 but named no edge")
	}
	return edge, nil
}

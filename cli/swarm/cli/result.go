package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
)

// contractVersion is the version of contract-swarm-json every result
// carries, so a consumer branches on the version it was handed rather
// than on the binary's own version.
const contractVersion = "1.0"

// Fact is one thing a command reports, together with the invocation or
// invocations that established it.
//
// The invocations are not decoration. An agent that can see the command
// behind a line can re-run it, argue with it, and work around it when the
// underlying tool changes; an agent that cannot has to trust a wrapper it
// did not write. A fact with no invocation behind it cannot be rendered
// — [Result.Validate] refuses it — because a fact nobody can reproduce is
// the thing this CLI exists to stop producing.
type Fact struct {
	// Name identifies the fact. It appears verbatim in both output
	// forms, so a consumer keys on it.
	Name string
	// Value is what was established. Nil when the fact is unestablished.
	Value any
	// Established is false when the fact could not be taken. Such a fact
	// is still reported — as unestablished, with a reason — rather than
	// omitted or emitted as an empty value of its type.
	Established bool
	// Reason says why an unestablished fact could not be taken.
	Reason string
	// Invocations are the argv vectors a reader can re-run to reproduce
	// the fact, each with its executable first. At least one is required.
	Invocations [][]string
}

// Established returns a fact that was taken.
func Established(name string, value any, invocations ...[]string) Fact {
	return Fact{Name: name, Value: value, Established: true, Invocations: invocations}
}

// Unestablished returns a fact that could not be taken, with the reason
// and the invocations that were attempted.
func Unestablished(name, reason string, invocations ...[]string) Fact {
	return Fact{Name: name, Established: false, Reason: reason, Invocations: invocations}
}

// Result is everything one invocation of a command reports: what it
// concluded, the facts behind that, the invocations it ran, and — in
// explain mode — the transcript it printed instead of running them.
//
// One value, two renderings. The machine-readable form carries every
// fact the text form carries because both are rendered from this, not
// because two renderers were kept in agreement.
type Result struct {
	// Command is the invocation path, e.g. "swarm manifest get".
	Command string
	// Run is the run epic the command was scoped to, empty for a command
	// that is not run-scoped.
	Run string
	// Mode is which of the three interpreters ran.
	Mode plan.Mode
	// Outcome is what the command concluded. The exit status is derived
	// from it.
	Outcome Outcome
	// Summary is one line saying what the command found. Optional.
	Summary string
	// Resolution names what would resolve the condition. Required when
	// Outcome is [OutcomeAttention].
	Resolution string
	// Facts are what the command established, in report order.
	Facts []Fact
	// Records are the plan's per-step records, in the order the steps
	// were considered.
	Records []plan.Record
	// Transcript is what the plan runner printed — the explained
	// invocations in explain mode, and the executed ones otherwise.
	Transcript []string
}

// Status is the exit status this result implies.
func (r Result) Status() Status { return r.Outcome.Status() }

// Validate reports a result this package refuses to render.
//
// Both rules are contract obligations rather than taste: an attention
// exit whose output does not name what would resolve it leaves a caller
// with a status and nothing to act on, and a fact with no invocation
// behind it cannot be reproduced.
func (r Result) Validate() error {
	if r.Outcome == OutcomeAttention && strings.TrimSpace(r.Resolution) == "" {
		return fmt.Errorf("cli: %s reported the attention outcome without naming what would resolve it", r.Command)
	}
	for _, f := range r.Facts {
		if len(f.Invocations) == 0 {
			return fmt.Errorf("cli: fact %q names no invocation that established it", f.Name)
		}
		for _, argv := range f.Invocations {
			if len(argv) == 0 {
				return fmt.Errorf("cli: fact %q names an empty invocation", f.Name)
			}
		}
	}
	return nil
}

// --- the machine-readable envelope ---------------------------------

type envelope struct {
	ContractVersion string           `json:"contract_version"`
	Command         string           `json:"command"`
	Run             string           `json:"run"`
	Mode            string           `json:"mode"`
	Outcome         string           `json:"outcome"`
	Status          int              `json:"status"`
	Summary         string           `json:"summary"`
	Resolution      string           `json:"resolution"`
	Facts           []factJSON       `json:"facts"`
	Invocations     []invocationJSON `json:"invocations"`
	Transcript      []string         `json:"transcript"`
}

type factJSON struct {
	Name        string     `json:"name"`
	Value       any        `json:"value"`
	Established bool       `json:"established"`
	Reason      string     `json:"reason"`
	Invocations [][]string `json:"invocations"`
}

type invocationJSON struct {
	Step string   `json:"step"`
	Argv []string `json:"argv"`
	Dir  string   `json:"dir"`
	// Status is the step's disposition: ran, refused or skipped.
	Status string `json:"status"`
	Reason string `json:"reason"`
	// ExitStatus is meaningful only when Status is "ran".
	ExitStatus int `json:"exit_status"`
	// DurationMS is an integer count of milliseconds. Not a formatted
	// duration: the contract forbids a field whose meaning depends on the
	// locale of the process that produced it.
	DurationMS   int64  `json:"duration_ms"`
	StdoutDigest string `json:"stdout_digest"`
	StderrDigest string `json:"stderr_digest"`
	Error        string `json:"error"`
}

func (r Result) envelope() envelope {
	env := envelope{
		ContractVersion: contractVersion,
		Command:         r.Command,
		Run:             r.Run,
		Mode:            r.Mode.String(),
		Outcome:         string(r.Outcome),
		Status:          int(r.Status()),
		Summary:         r.Summary,
		Resolution:      r.Resolution,
		// Never null: a consumer iterating a result should not have to
		// distinguish an absent list from an empty one.
		Facts:       []factJSON{},
		Invocations: []invocationJSON{},
		Transcript:  []string{},
	}
	for _, f := range r.Facts {
		invocations := f.Invocations
		if invocations == nil {
			invocations = [][]string{}
		}
		env.Facts = append(env.Facts, factJSON{
			Name:        f.Name,
			Value:       f.Value,
			Established: f.Established,
			Reason:      f.Reason,
			Invocations: invocations,
		})
	}
	for _, rec := range r.Records {
		argv := []string{}
		if rec.Invocation.Name != "" {
			argv = append([]string{rec.Invocation.Name}, rec.Invocation.Args...)
		}
		errText := ""
		if rec.Err != nil {
			errText = rec.Err.Error()
		}
		env.Invocations = append(env.Invocations, invocationJSON{
			Step:         rec.Step,
			Argv:         argv,
			Dir:          rec.Invocation.Dir,
			Status:       string(rec.Status),
			Reason:       rec.Reason,
			ExitStatus:   rec.ExitStatus,
			DurationMS:   rec.Duration.Milliseconds(),
			StdoutDigest: rec.StdoutDigest,
			StderrDigest: rec.StderrDigest,
			Error:        errText,
		})
	}
	env.Transcript = append(env.Transcript, r.Transcript...)
	return env
}

// Render writes the result to w, as JSON when asJSON is set and as text
// otherwise.
//
// The envelope is emitted on every exit status, including a non-zero one,
// so a consumer parsing the result never has to fall back to reading text
// to learn why.
func Render(w io.Writer, r Result, asJSON bool) error {
	if asJSON {
		data, err := json.MarshalIndent(r.envelope(), "", "  ")
		if err != nil {
			return fmt.Errorf("cli: rendering the result as JSON: %w", err)
		}
		_, err = fmt.Fprintf(w, "%s\n", data)
		return err
	}
	return renderText(w, r)
}

func renderText(w io.Writer, r Result) error {
	head := r.Command
	if r.Run != "" {
		head += " — run " + r.Run
	}
	head += " (" + r.Mode.String() + ")"
	if _, err := fmt.Fprintln(w, head); err != nil {
		return err
	}

	for _, line := range r.Transcript {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	for _, f := range r.Facts {
		// One line per fact, and the line names the command a reader can
		// re-run to reproduce it. No padding or alignment: a field whose
		// rendering depends on the terminal's width is one the contract
		// forbids.
		value := renderValue(f)
		commands := make([]string, 0, len(f.Invocations))
		for _, argv := range f.Invocations {
			commands = append(commands, plan.Argv(argv))
		}
		if _, err := fmt.Fprintf(w, "%s = %s    <- %s\n", f.Name, value, strings.Join(commands, " ; ")); err != nil {
			return err
		}
	}

	for _, rec := range r.Records {
		line := fmt.Sprintf("%s: %s", rec.Step, rec.Status)
		if rec.Status == plan.StepRan {
			line += fmt.Sprintf(" exit %d %dms %s %s", rec.ExitStatus, rec.Duration.Milliseconds(),
				rec.StdoutDigest, rec.StderrDigest)
		}
		if rec.Reason != "" {
			line += " (" + rec.Reason + ")"
		}
		if rec.Err != nil {
			line += " error: " + rec.Err.Error()
		}
		if len(rec.Invocation.Name) > 0 {
			line += "    <- " + plan.Argv(append([]string{rec.Invocation.Name}, rec.Invocation.Args...))
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	if r.Summary != "" {
		if _, err := fmt.Fprintf(w, "summary: %s\n", r.Summary); err != nil {
			return err
		}
	}
	if r.Resolution != "" {
		if _, err := fmt.Fprintf(w, "what would resolve this: %s\n", r.Resolution); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "outcome: %s (exit %d)\n", r.Outcome, int(r.Status()))
	return err
}

// renderValue renders a fact's value for the text form, using the same
// bytes the JSON form carries so the two cannot say different things
// about the same value.
func renderValue(f Fact) string {
	if !f.Established {
		return "unestablished: " + f.Reason
	}
	if s, ok := f.Value.(string); ok {
		return s
	}
	data, err := json.Marshal(f.Value)
	if err != nil {
		return fmt.Sprintf("%v", f.Value)
	}
	return string(data)
}

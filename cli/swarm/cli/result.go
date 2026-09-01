package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/contract"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

// contractVersion is the version of contract-swarm-json every result
// carries, so a consumer branches on the version it was handed rather
// than on the binary's own version. Declared once, in the contract
// package, because the end-to-end fixture's stamp is checked against the
// same value.
const contractVersion = contract.Version

// Fact is one thing a command reports, together with the invocation or
// invocations that established it.
//
// The invocations are not decoration. An agent that can see the command
// behind a line can re-run it, argue with it, and work around it when the
// underlying tool changes; an agent that cannot has to trust a wrapper it
// did not write. A fact with no invocation behind it cannot be rendered
// — [Result.Validate] refuses it — because a fact nobody can reproduce is
// the thing this CLI exists to stop producing.
//
// A fact's value is one of four shapes rather than anything at all. That
// is what [contract-swarm-json]'s schema fixes, and it is why a fact is
// built with one of the constructors below rather than by hand: a
// consumer branches on the kind and reads the field that kind carries,
// instead of guessing at the JSON type it was handed.
type Fact struct {
	// Name identifies the fact. It is the property name an answer keys
	// on and the name the full result reports, so a consumer keys on it.
	Name string
	// Value is what was established. Nil when the fact is unestablished.
	Value *swarmmodel.FactValue
	// Established is false when the fact could not be taken. Such a fact
	// is still reported — as unestablished, with a reason — rather than
	// omitted or emitted as an empty value of its type.
	Established bool
	// Reason says what is wrong with this fact: why it could not be
	// taken, or what is wrong with the value it carries.
	Reason string
	// Resolution names what would put this fact right, when anything
	// would. A command reporting several unsatisfied properties resolves
	// each differently, so the resolution that reaches a reader is the
	// property's own before the result's.
	Resolution string
	// Invocations are the argv vectors a reader can re-run to reproduce
	// the fact, each with its executable first. At least one is required.
	Invocations [][]string
}

// Text returns an established fact whose value is a word.
func Text(name, value string, invocations ...[]string) Fact {
	v := swarmmodel.FactValue{Kind: swarmmodel.FactKindText, Text: &value}
	return Fact{Name: name, Value: &v, Established: true, Invocations: invocations}
}

// Count returns an established fact whose value is a count.
func Count(name string, value int, invocations ...[]string) Fact {
	n := int64(value)
	v := swarmmodel.FactValue{Kind: swarmmodel.FactKindCount, Count: &n}
	return Fact{Name: name, Value: &v, Established: true, Invocations: invocations}
}

// Flag returns an established fact whose value is yes or no.
func Flag(name string, value bool, invocations ...[]string) Fact {
	v := swarmmodel.FactValue{Kind: swarmmodel.FactKindFlag, Flag: &value}
	return Fact{Name: name, Value: &v, Established: true, Invocations: invocations}
}

// List returns an established fact whose value is a list of words.
func List(name string, values []string, invocations ...[]string) Fact {
	if values == nil {
		values = []string{}
	}
	v := swarmmodel.FactValue{Kind: swarmmodel.FactKindList, Items: &values}
	return Fact{Name: name, Value: &v, Established: true, Invocations: invocations}
}

// Structured returns an established fact whose value is a JSON document
// — an object, or a list of them.
//
// It is carried as text because the contract's schema language offers no
// type for an arbitrary structure. The renderer parses it back and emits
// it as real structure, so a reader sees a mapping and not an escaped
// string; a value that is not valid JSON is a defect in the caller, and
// [Result.Validate] refuses it rather than emitting text that claims to
// be a document.
func Structured(name, document string, invocations ...[]string) Fact {
	v := swarmmodel.FactValue{Kind: swarmmodel.FactKindStructured, Structured: &document}
	return Fact{Name: name, Value: &v, Established: true, Invocations: invocations}
}

// Invalid returns a fact whose value was read and is wrong.
//
// Established and unsatisfactory at once, which is a state the four
// constructors above cannot express: the caller gets the value that is
// there AND what is wrong with it, rather than a null that hides which
// of "unwritten" and "written badly" it is looking at.
func Invalid(name string, value swarmmodel.FactValue, reason string, invocations ...[]string) Fact {
	return Fact{Name: name, Value: &value, Established: true, Reason: reason, Invocations: invocations}
}

// coherent reports whether a value carries exactly the field its kind
// names. The schema can say which fields may appear; which combination
// is coherent is a rule only code can hold, so this is where it is held.
func coherent(v swarmmodel.FactValue) bool {
	present := 0
	for _, carried := range []bool{v.Text != nil, v.Count != nil, v.Flag != nil, v.Items != nil, v.Structured != nil} {
		if carried {
			present++
		}
	}
	if present != 1 {
		return false
	}
	switch v.Kind {
	case swarmmodel.FactKindText:
		return v.Text != nil
	case swarmmodel.FactKindCount:
		return v.Count != nil
	case swarmmodel.FactKindFlag:
		return v.Flag != nil
	case swarmmodel.FactKindList:
		return v.Items != nil
	case swarmmodel.FactKindStructured:
		return v.Structured != nil && json.Valid([]byte(*v.Structured))
	default:
		return false
	}
}

// Unestablished returns a fact that could not be taken, with the reason
// and the invocations that were attempted.
func Unestablished(name, reason string, invocations ...[]string) Fact {
	return Fact{Name: name, Established: false, Reason: reason, Invocations: invocations}
}

// Result is everything one invocation of a command reports: what it
// concluded, the facts behind that, the invocations it ran, and the
// transcript the plan runner printed.
//
// One value, two renderings. YAML and JSON carry the same values in the
// same order because both are rendered from this, through one node tree,
// not because two renderers were kept in agreement.
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
// Every rule is a contract obligation rather than taste: an attention
// exit whose output does not name what would resolve it leaves a caller
// with a status and nothing to act on; a fact with no invocation behind
// it cannot be reproduced; and an established fact with no value, or an
// unestablished one carrying a value, contradicts the flag a consumer
// branches on.
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
		if f.Established && f.Value == nil {
			return fmt.Errorf("cli: fact %q is reported as established and carries no value", f.Name)
		}
		if !f.Established && f.Value != nil {
			return fmt.Errorf("cli: fact %q is reported as unestablished and carries a value", f.Name)
		}
		if f.Value != nil && !coherent(*f.Value) {
			return fmt.Errorf("cli: fact %q carries a %q value and not the field that kind names", f.Name, f.Value.Kind)
		}
	}
	return nil
}

// envelope builds the generated contract type this result is emitted as.
//
// The shape lives in schemas/swarm/output.strata and reaches this file as
// generated Go, so there is one declaration of it rather than a schema
// and a struct kept in agreement by hand ([[adr-016]]).
func (r Result) envelope() swarmmodel.CommandResult {
	env := swarmmodel.CommandResult{
		ContractVersion: contractVersion,
		Command:         r.Command,
		Run:             r.Run,
		Mode:            swarmmodel.CommandMode(r.Mode.String()),
		Outcome:         swarmmodel.CommandOutcome(r.Outcome),
		Status:          int64(r.Status()),
		Summary:         r.Summary,
		Resolution:      r.Resolution,
		// Never null: a consumer iterating a result should not have to
		// distinguish an absent list from an empty one.
		Facts:       []swarmmodel.Fact{},
		Invocations: []swarmmodel.Invocation{},
		Transcript:  []string{},
	}

	for _, f := range r.Facts {
		// One complete command per invocation, quoted so it can be
		// pasted into a shell — which is what a reader does with it. An
		// argv vector would make them reassemble it first, and get the
		// quoting wrong on any argument carrying a space.
		invocations := []string{}
		for _, argv := range f.Invocations {
			invocations = append(invocations, plan.Argv(argv))
		}
		env.Facts = append(env.Facts, swarmmodel.Fact{
			Name:        f.Name,
			Established: f.Established,
			Reason:      f.Reason,
			Resolution:  f.Resolution,
			Value:       f.Value,
			Invocations: invocations,
		})
	}

	for _, rec := range r.Records {
		command := ""
		if rec.Invocation.Name != "" {
			command = plan.Argv(append([]string{rec.Invocation.Name}, rec.Invocation.Args...))
		}
		errText := ""
		if rec.Err != nil {
			errText = rec.Err.Error()
		}
		env.Invocations = append(env.Invocations, swarmmodel.Invocation{
			Step:         rec.Step,
			Command:      command,
			Dir:          rec.Invocation.Dir,
			Status:       swarmmodel.StepStatus(rec.Status),
			Reason:       rec.Reason,
			ExitStatus:   int64(rec.ExitStatus),
			DurationMs:   rec.Duration.Milliseconds(),
			StdoutDigest: rec.StdoutDigest,
			StderrDigest: rec.StderrDigest,
			Error:        errText,
		})
	}

	env.Transcript = append(env.Transcript, r.Transcript...)
	return env
}

// Render writes the full result to w, as JSON when asJSON is set and as
// YAML otherwise.
//
// It renders the [swarmmodel.CommandResult] the contract declares, field
// for field, from the one node tree both renderings are built over. The
// result is emitted on every exit status, including a non-zero one, so a
// consumer parsing it never has to fall back to reading prose to learn
// why.
func Render(w io.Writer, r Result, asJSON bool) error {
	node := nodeOf(reflect.ValueOf(r.envelope()))
	if asJSON {
		return writeJSON(w, node)
	}
	return writeYAML(w, node)
}

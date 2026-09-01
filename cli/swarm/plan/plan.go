// Package plan is the Swarm CLI's Plan Runner: a command builds an
// ordered plan of step values, and three interpreters read it, per
// ADR-024 and the Plan Runner component in
// specs/architecture/container-swarm-cli.md.
//
// The point of the indirection is that explain, dry-run and execute
// cannot disagree about what a command does, because there is one
// description and three readers of it. A step added for execution appears
// in explain output with no second edit, the ordering guard is a field
// rather than a conjunction an author can forget, and the whole sequence
// is a value a test can assert without starting a process.
//
// Two consequences ADR-024 accepts, restated here because they shape the
// API rather than merely justifying it:
//
//   - An argument produced by an earlier step is a reference value
//     ([From]) rather than a Go variable. That is what lets explain print
//     it as the step that produces it instead of as a value it would have
//     had to run something to learn.
//   - Explain output for a branching plan is a labelled transcript rather
//     than a linear script: the condition is named and both arms are
//     rendered, and a reader follows the one their situation matches.
package plan

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
)

// Effect declares whether a step reads or writes, which is the whole
// basis of dry-run mode.
type Effect int

const (
	// Read is a step that observes and changes nothing. Dry-run executes
	// these, which is what makes its report reflect live state.
	Read Effect = iota
	// Write is a step that changes something. Dry-run refuses these.
	//
	// It is the zero value's opposite deliberately: [Read] is the zero
	// value, so a step that forgot to declare an effect is executed by a
	// dry run. That is the wrong default for safety, which is why
	// [Plan.Validate] requires every step to state its effect explicitly
	// rather than leaving the zero value to mean anything.
	Write
)

func (e Effect) String() string {
	switch e {
	case Read:
		return "read"
	case Write:
		return "write"
	default:
		return fmt.Sprintf("Effect(%d)", int(e))
	}
}

// Failure is what the interpreter does when a step fails.
type Failure int

const (
	// Stop prevents every later step in the plan. This is the ordering
	// guard: a step declaring it is honoured by the interpreter, so no
	// command can chain a removal after a failed merge.
	Stop Failure = iota
	// Continue lets the plan carry on. It is for a step whose failure is
	// information rather than an obstacle — a query that legitimately
	// finds nothing.
	Continue
)

func (f Failure) String() string {
	switch f {
	case Stop:
		return "stop"
	case Continue:
		return "continue"
	default:
		return fmt.Sprintf("Failure(%d)", int(f))
	}
}

// Stop is the zero value on purpose: a step that forgot to declare a
// failure policy stops the plan, which is the conservative reading.

// Arg is one argument of a step's invocation: either a literal, or a
// reference to a value an earlier step produces.
//
// The two are a single type rather than an interface so that a plan stays
// a plain value a test can build and compare. Use [Lit] and [From] to
// construct one; the zero Arg is an empty literal.
type Arg struct {
	literal string
	// step is the name of the step producing this argument, empty for a
	// literal.
	step string
	// pick extracts the argument from the producing step's decoded value.
	// Nil means the decoded value must itself be a string.
	pick func(any) (string, error)
}

// Lit is an argument known when the plan is built.
func Lit(s string) Arg { return Arg{literal: s} }

// From is an argument only the named earlier step can supply. pick
// extracts it from that step's decoded value; a nil pick requires the
// decoded value to be a string.
//
// Explain mode renders this as the step rather than as a value, because
// learning the value would mean running the step explain mode exists not
// to run.
func From(step string, pick func(any) (string, error)) Arg {
	return Arg{step: step, pick: pick}
}

// IsRef reports whether the argument comes from an earlier step.
func (a Arg) IsRef() bool { return a.step != "" }

// Step is the step this argument comes from, empty for a literal.
func (a Arg) Step() string { return a.step }

// Literal is the argument's value, empty for a reference.
func (a Arg) Literal() string { return a.literal }

// Decoder turns a finished invocation's result into the value later steps
// and the caller read.
//
// Returning an error makes the step a failure, which is what puts a
// result that came back in an unexpected shape under the same ordering
// guard as a non-zero exit.
type Decoder func(swarmexec.Result) (any, error)

// Alternative is an arm taken when a condition holds after its step ran.
//
// Branching is a property of the step rather than of the code that built
// the plan, which is what keeps a plan fully renderable before anything
// runs.
type Alternative struct {
	// Condition names, in the run's own words, when this arm is taken.
	// Explain output prints it, so it is prose a reader can match their
	// situation against rather than an expression.
	Condition string
	// Select decides from the step's record and decoded value whether the
	// arm is taken. A nil Select is never taken automatically: the arm is
	// rendered by explain and skipped by the executing interpreters,
	// which is how a documented manual branch is expressed.
	Select func(Record, any) bool
	// Steps are the arm. They run immediately after the step that carries
	// them, and the outer plan then continues.
	Steps []Step
}

// Step is one external invocation and everything the interpreters need to
// know about it.
type Step struct {
	// Name identifies the step within its plan. It is what [From] refers
	// to and what explain output labels an argument with, so it is
	// required and must be unique.
	Name string
	// Description is one line saying what the step is for, printed as a
	// comment by explain. Optional.
	Description string
	// Executable is the tool to run. Arguments are Args so that any of
	// them may come from an earlier step; the executable itself may not,
	// because a plan whose tools are unknown until it runs cannot be
	// explained.
	Executable string
	// Args are the invocation's arguments, in order.
	Args []Arg
	// Dir is the working directory to run in. Empty means the process's
	// own.
	Dir string
	// Decode turns the result into the value later steps reference. Nil
	// means the value is the standard output with surrounding whitespace
	// trimmed, which is what most single-value reads want.
	Decode Decoder
	// OnFailure is what the interpreter does when this step fails. The
	// zero value is [Stop].
	OnFailure Failure
	// Answers are the exit statuses this step treats as having answered
	// rather than failed, beyond 0 which always does. Empty means only 0.
	//
	// Some tools report an answer through the status rather than through
	// the output: `git merge-base --is-ancestor` exits 0 for yes, 1 for
	// no and 128 when a revision does not resolve. Without this, "no"
	// would trip the ordering guard and publish no value — a legitimate
	// answer read as a failure, which is the mirror image of the failure
	// read as a pass that this container exists to remove.
	Answers []int
	// Effect declares whether the step reads or writes. Dry-run runs the
	// reads and refuses the writes.
	Effect Effect
	// Repeatable declares that re-running this step is safe. Recovery
	// from a partial plan is re-running rather than undoing (ADR-026), so
	// this is the field that says whether recovery is available. The zero
	// value is false — the conservative reading — so a step that says
	// nothing is treated as unsafe to repeat.
	Repeatable bool
	// Alternatives are the arms this step may branch into.
	Alternatives []Alternative
}

// Steps is an ordered fragment of a plan whose LAST step's value is the
// fragment's answer.
//
// It is what a component returns when one of its operations needs more
// than a single invocation — reading the standing worktrees means
// clearing stale registrations first; writing the manifest conditionally
// means guarding before writing. A component whose operations all take
// one invocation returns a bare [Step].
type Steps []Step

// Answer is the name of the step carrying the fragment's answer.
func (s Steps) Answer() string {
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1].Name
}

// ValueFrom reads the fragment's answer out of a report.
func (s Steps) ValueFrom(report Report) (any, error) {
	name := s.Answer()
	if name == "" {
		return nil, fmt.Errorf("plan: an empty fragment has no answer")
	}
	record, ok := report.Record(name)
	if !ok {
		return nil, fmt.Errorf("plan: the report has no record for step %q", name)
	}
	if record.Err != nil {
		return nil, fmt.Errorf("plan: step %q: %w", name, record.Err)
	}
	if record.Value == nil {
		return nil, fmt.Errorf("plan: step %q produced no value (%s)", name, record.Status)
	}
	return record.Value, nil
}

// Plan is an ordered sequence of steps: what a command would do,
// as a value.
type Plan struct {
	// Name is what the plan is for, printed as a header by explain.
	Name string
	// Steps are run in order.
	Steps []Step
}

// StepStatus is what became of one step.
type StepStatus string

const (
	// StepRan means the step's process started and finished.
	StepRan StepStatus = "ran"
	// StepRefused means a dry run declined to execute it because it is
	// not a read.
	StepRefused StepStatus = "refused"
	// StepSkipped means the step never ran: an earlier step stopped the
	// plan, or an argument it needed came from a step that did not run.
	StepSkipped StepStatus = "skipped"
)

// Record is what the Plan Runner records per step: the invocation, its
// exit status, its duration and a digest of each stream.
//
// Digests rather than the streams themselves: a record is an audit line,
// and a run that carried every byte of every tool's output would be
// unreadable and would put tracker contents into a report that is meant
// to be shareable. The digest still detects a step whose output changed
// between two runs.
type Record struct {
	// Step is the step's name.
	Step string
	// Status is what became of the step.
	Status StepStatus
	// Invocation is the resolved argv, with every reference replaced by
	// the value it resolved to. It is empty when the step did not run.
	Invocation swarmexec.Invocation
	// ExitStatus is the status the finished process reported. It is
	// meaningful only when Status is [StepRan].
	ExitStatus int
	// Duration is how long the invocation took.
	Duration time.Duration
	// StdoutDigest and StderrDigest are "sha256:<hex>" over each stream,
	// separately, because the streams are separate.
	StdoutDigest string
	StderrDigest string
	// Answered records whether the step's exit status is one it declared
	// as an answer — 0, or one of its [Step.Answers]. It is on the record
	// rather than recomputed because the record outlives the step value
	// it came from.
	Answered bool
	// Value is what the step's decoder made of the result, nil when the
	// step did not run, its decoder failed, or the process exited with a
	// status the step does not answer with. It is here rather than only
	// in the interpreter's internal
	// state because a reference argument is not the only reader of a
	// decoded result: a command turning a plan's outcome into facts reads
	// them from the report, and a decoder whose output nothing can reach
	// is not a decoder.
	Value any
	// Err is why the step failed, when it did: the process could not be
	// started, the decoder refused the result, or an argument could not be
	// resolved. A non-zero exit alone is not an error.
	Err error
	// Reason explains a StepRefused or StepSkipped in one line.
	Reason string
	// Branch names the alternative's condition when this record belongs to
	// an arm rather than to the main sequence.
	Branch string
}

// Failed reports whether the step is a failure for the purpose of the
// ordering guard: it could not run, its decoder refused the result, or
// the process exited with a status the step does not answer with.
//
// A refused or skipped step is not a failure — it is an outcome the
// interpreter chose — so it never stops a plan on its own.
func (r Record) Failed() bool {
	if r.Status != StepRan {
		return false
	}
	return r.Err != nil || !r.Answered
}

// answers reports whether status is one this step treats as an answer.
func (s Step) answers(status int) bool {
	if status == 0 {
		return true
	}
	for _, a := range s.Answers {
		if a == status {
			return true
		}
	}
	return false
}

// Report is what an interpreted plan produced.
type Report struct {
	// Plan is the name of the plan that produced it.
	Plan string
	// Records are one per step considered, in the order they were
	// considered, including the ones that did not run.
	Records []Record
	// Stopped is the name of the step whose failure stopped the plan,
	// empty when nothing did.
	Stopped string
}

// Failed reports whether any step in the report failed.
func (r Report) Failed() bool {
	for _, rec := range r.Records {
		if rec.Failed() {
			return true
		}
	}
	return false
}

// Record returns the record for the named step.
func (r Report) Record(step string) (Record, bool) {
	for _, rec := range r.Records {
		if rec.Step == step {
			return rec, true
		}
	}
	return Record{}, false
}

// digest is the audit digest of one stream.
func digest(b []byte) string {
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Validate reports every structural defect in the plan: a step with no
// name or no executable, a duplicate name, or a reference to a step that
// is not strictly earlier in the same sequence.
//
// The forward-reference check is what makes "an argument produced by an
// earlier step" a property of the plan rather than a convention. It is
// checked before anything runs, so a command whose plan is malformed says
// so instead of failing partway through.
func (p Plan) Validate() error {
	return validateSteps(p.Steps, map[string]bool{})
}

func validateSteps(steps []Step, seen map[string]bool) error {
	for i, s := range steps {
		if s.Name == "" {
			return fmt.Errorf("plan: step %d has no name", i)
		}
		if s.Executable == "" {
			return fmt.Errorf("plan: step %q has no executable", s.Name)
		}
		if seen[s.Name] {
			return fmt.Errorf("plan: step %q is declared more than once", s.Name)
		}
		for _, a := range s.Args {
			if !a.IsRef() {
				continue
			}
			if !seen[a.step] {
				return fmt.Errorf("plan: step %q takes an argument from %q, which is not an earlier step", s.Name, a.step)
			}
		}
		seen[s.Name] = true

		for _, alt := range s.Alternatives {
			if alt.Condition == "" {
				return fmt.Errorf("plan: an alternative of step %q names no condition", s.Name)
			}
			// An arm's steps are validated against the names available
			// where the arm runs, which includes the step carrying it.
			armSeen := make(map[string]bool, len(seen))
			for k, v := range seen {
				armSeen[k] = v
			}
			if err := validateSteps(alt.Steps, armSeen); err != nil {
				return err
			}
		}
	}
	return nil
}

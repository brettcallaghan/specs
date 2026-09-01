package plan

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
)

// Mode is one of the three interpreters a plan has. There are exactly
// three and there is no fourth: a mode that ran some writes, or one that
// printed without the plan saying so, would be a second description of
// what a command does.
type Mode int

const (
	// ModeExplain prints the invocations the plan would run, in order,
	// and executes none of them. It reads nothing, so it is safe against
	// a repository nobody has looked at.
	ModeExplain Mode = iota
	// ModeDryRun executes the steps declared as reads and refuses every
	// other step. Unlike explain it runs, so its report reflects live
	// state.
	ModeDryRun
	// ModeExecute runs every step, printing each invocation as it runs
	// it.
	ModeExecute
)

func (m Mode) String() string {
	switch m {
	case ModeExplain:
		return "explain"
	case ModeDryRun:
		return "dry-run"
	case ModeExecute:
		return "execute"
	default:
		return fmt.Sprintf("Mode(%d)", int(m))
	}
}

// Modes returns the interpreters a plan has, in order. It is the
// enumeration the command surface offers and the one a test asserts
// against, so "exactly three" is a fact about the code rather than a
// sentence about it.
func Modes() []Mode { return []Mode{ModeExplain, ModeDryRun, ModeExecute} }

// Runner interprets plans. It is the Plan Runner component: it holds the
// Process Runner every step goes through and the stream the interpreters
// print to, and nothing else.
type Runner struct {
	// Exec runs each step's invocation. Required by [ModeDryRun] and
	// [ModeExecute]; [ModeExplain] never touches it, which is the
	// property that makes explain safe.
	Exec swarmexec.Runner
	// Out is where the interpreters print. A nil Out discards, so a test
	// asserting only the report does not have to supply one.
	Out io.Writer
}

// ErrNoProcessRunner reports that a mode which runs steps was asked for
// without a process runner to run them through.
var ErrNoProcessRunner = errors.New("plan: no process runner")

// Run interprets p in the given mode.
//
// The plan is validated first, so a malformed plan — a duplicate step
// name, an argument taken from a step that is not earlier — is reported
// before anything is printed or run, rather than partway through.
func (r Runner) Run(ctx context.Context, p Plan, mode Mode) (Report, error) {
	if err := p.Validate(); err != nil {
		return Report{}, err
	}

	out := r.Out
	if out == nil {
		out = io.Discard
	}

	switch mode {
	case ModeExplain:
		return Report{Plan: p.Name}, r.explain(out, p)
	case ModeDryRun, ModeExecute:
		if r.Exec == nil {
			return Report{}, ErrNoProcessRunner
		}
		report := Report{Plan: p.Name}
		st := &state{values: map[string]any{}, produced: map[string]bool{}}
		r.interpret(ctx, out, p.Steps, mode, st, &report, "")
		return report, nil
	default:
		return Report{}, fmt.Errorf("plan: unknown mode %d", int(mode))
	}
}

// state carries what earlier steps produced, so a reference argument can
// be resolved when a step actually runs.
type state struct {
	values   map[string]any
	produced map[string]bool
}

// explain prints the plan and runs nothing.
//
// It never consults r.Exec, which is what makes "executes none of the
// steps it printed" structural: there is no code path here that could.
func (r Runner) explain(out io.Writer, p Plan) error {
	if p.Name != "" {
		if _, err := fmt.Fprintf(out, "# %s\n", p.Name); err != nil {
			return err
		}
	}
	n := 0
	return explainSteps(out, p.Steps, "", &n)
}

func explainSteps(out io.Writer, steps []Step, indent string, n *int) error {
	for _, s := range steps {
		*n++
		head := fmt.Sprintf("%s# %d. %s", indent, *n, s.Name)
		if s.Description != "" {
			head += " — " + s.Description
		}
		head += " " + annotation(s)
		if _, err := fmt.Fprintln(out, head); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(out, "%s%s\n", indent, renderPlanned(s)); err != nil {
			return err
		}

		for _, alt := range s.Alternatives {
			// Both arms, and the condition that selects between them.
			// The reader follows the one their situation matches; a
			// branching plan is a labelled transcript rather than a
			// linear script, which ADR-024 states as a consequence.
			if _, err := fmt.Fprintf(out, "%s# if %s:\n", indent, alt.Condition); err != nil {
				return err
			}
			if err := explainSteps(out, alt.Steps, indent+"  ", n); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "%s# otherwise: continue below\n", indent); err != nil {
				return err
			}
		}
	}
	return nil
}

// interpret runs a sequence of steps and reports whether the plan
// stopped. It is shared by dry-run and execute, which differ only in
// which steps they are willing to run.
func (r Runner) interpret(
	ctx context.Context,
	out io.Writer,
	steps []Step,
	mode Mode,
	st *state,
	report *Report,
	branch string,
) bool {
	for i, s := range steps {
		record := Record{Step: s.Name, Branch: branch}

		if mode == ModeDryRun && s.Effect == Write {
			record.Status = StepRefused
			record.Reason = "dry run: the step writes"
			fmt.Fprintf(out, "# refused, dry run writes nothing: %s\n", s.Name)
			report.Records = append(report.Records, record)
			continue
		}

		args, err := resolve(s, st)
		if err != nil {
			record.Status = StepSkipped
			record.Reason = err.Error()
			record.Err = err
			report.Records = append(report.Records, record)
			if s.OnFailure == Stop {
				report.Stopped = s.Name
				skipRest(report, steps[i+1:], branch, fmt.Sprintf("%s stopped the plan", s.Name))
				return true
			}
			continue
		}

		inv := swarmexec.Invocation{Name: s.Executable, Args: args, Dir: s.Dir}

		// Printed before it runs, not after: a run interrupted partway
		// has still named the invocation that was in flight.
		fmt.Fprintln(out, renderResolved(inv.Name, inv.Args))

		res, runErr := r.Exec.Run(ctx, inv)

		record.Status = StepRan
		record.Invocation = inv
		record.ExitStatus = res.ExitStatus
		record.Duration = res.Duration
		record.StdoutDigest = digest(res.Stdout)
		record.StderrDigest = digest(res.Stderr)
		record.Answered = runErr == nil && s.answers(res.ExitStatus)

		if runErr != nil {
			record.Err = runErr
		} else {
			value, decodeErr := decode(s, res)
			if decodeErr != nil {
				record.Err = fmt.Errorf("plan: step %q: %w", s.Name, decodeErr)
			} else if record.Answered {
				// A failed step's value is not published: a later step
				// referencing it would otherwise be handed whatever a
				// failing tool happened to print.
				st.values[s.Name] = value
				st.produced[s.Name] = true
				record.Value = value
			}
		}

		report.Records = append(report.Records, record)

		if record.Failed() && s.OnFailure == Stop {
			report.Stopped = s.Name
			skipRest(report, steps[i+1:], branch, fmt.Sprintf("%s stopped the plan", s.Name))
			return true
		}

		for _, alt := range s.Alternatives {
			if alt.Select == nil {
				continue
			}
			if !alt.Select(record, st.values[s.Name]) {
				continue
			}
			if r.interpret(ctx, out, alt.Steps, mode, st, report, alt.Condition) {
				return true
			}
			break
		}
	}
	return false
}

// skipRest records the steps that were not run because the plan stopped,
// so the report accounts for every step rather than ending early.
func skipRest(report *Report, steps []Step, branch, reason string) {
	for _, s := range steps {
		report.Records = append(report.Records, Record{
			Step:   s.Name,
			Status: StepSkipped,
			Reason: reason,
			Branch: branch,
		})
	}
}

// resolve turns a step's arguments into strings, replacing each reference
// with the value the named step produced.
func resolve(s Step, st *state) ([]string, error) {
	out := make([]string, len(s.Args))
	for i, a := range s.Args {
		if !a.IsRef() {
			out[i] = a.literal
			continue
		}
		if !st.produced[a.step] {
			return nil, fmt.Errorf("plan: step %q takes an argument from %q, which produced no value", s.Name, a.step)
		}
		value := st.values[a.step]
		if a.pick != nil {
			picked, err := a.pick(value)
			if err != nil {
				return nil, fmt.Errorf("plan: step %q reading its argument from %q: %w", s.Name, a.step, err)
			}
			out[i] = picked
			continue
		}
		text, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("plan: step %q takes an argument from %q, whose value is %T and not a string; give the reference a pick function", s.Name, a.step, value)
		}
		out[i] = text
	}
	return out, nil
}

// decode turns a result into the value later steps reference. A step with
// no decoder yields its standard output with surrounding whitespace
// trimmed, which is what a single-value read wants and what most tools
// produce.
func decode(s Step, res swarmexec.Result) (any, error) {
	if s.Decode == nil {
		return strings.TrimSpace(string(res.Stdout)), nil
	}
	return s.Decode(res)
}

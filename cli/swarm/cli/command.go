package cli

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
)

// Request is one parsed invocation: the flags every subcommand carries,
// resolved.
type Request struct {
	// Command is the invocation path, e.g. "swarm manifest get".
	Command string
	// Run is the run epic this invocation is scoped to.
	Run string
	// Mode is which interpreter the flags selected.
	Mode plan.Mode
	// JSON is whether the machine-readable form was asked for.
	JSON bool
	// Bare is whether the answer is to be printed on its own, without
	// the [Result] envelope around it. It is true only for a subcommand
	// that declares a bare form and was not asked for the full result.
	Bare bool
}

// Spec is what a subcommand supplies to the surface. Everything else —
// the flags, the scope check, the mode, the rendering and the exit status
// — is the surface's, so no subcommand implements any of it a second
// time.
type Spec struct {
	Use   string
	Short string
	Long  string
	// RunScoped declares that the subcommand reads or mutates run state.
	// The surface then requires a non-empty --run and refuses the
	// invocation before anything is read or written.
	RunScoped bool
	// Plan builds what this invocation would run. Returning a [*Refusal]
	// chooses the outcome; any other error is a condition the caller
	// cannot repair in place.
	Plan func(Request) (plan.Plan, error)
	// Interpret turns the plan's report into the facts and the outcome to
	// report. It is never called in explain mode, which runs nothing and
	// therefore has nothing to interpret.
	Interpret func(Request, plan.Report) (Result, error)
	// Flags declares the flags this subcommand carries beyond the four
	// every subcommand carries. Optional.
	//
	// The contract flags are bound first and are not passed here, so a
	// subcommand cannot redeclare one of them without cobra saying so.
	Flags func(*pflag.FlagSet)
	// Bare declares that this subcommand has an answer smaller than its
	// report — a lookup, whose caller wants the value and not the
	// provenance around it. Such a subcommand carries --full-result,
	// prints [Result.Bare] on stdout when its verdict is clear, and puts
	// the report on stderr when it is not, so a caller can capture the
	// answer without a parser and still route on the status.
	//
	// It is a declaration and not a default: the report is what a command
	// with no smaller answer owes, and every other subcommand keeps it.
	Bare bool
}

// newSubcommand builds the cobra command for spec.
//
// The three mode flags and --run are declared here rather than by each
// subcommand, which is what makes "carried by every subcommand" a
// property of the surface instead of a checklist.
func newSubcommand(spec Spec, streams *ioStreams, runner swarmexec.Runner) *cobra.Command {
	var (
		explain    bool
		dryRun     bool
		asJSON     bool
		fullResult bool
		runID      string
	)

	cmd := &cobra.Command{
		Use:          spec.Use,
		Short:        spec.Short,
		Long:         spec.Long,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			req := Request{
				Command: cmd.CommandPath(),
				Run:     runID,
				JSON:    asJSON,
				Bare:    spec.Bare && !fullResult,
			}

			// The scope check comes first, before the mode is even
			// resolved: AC-SWARM-CORE-001.14 requires that nothing is
			// read or written outside the scope, and the cheapest way to
			// guarantee that is to refuse before anything could be.
			if spec.RunScoped && strings.TrimSpace(runID) == "" {
				return finish(streams, req, refusal(req, RefuseArguments(
					"%s is scoped to one run; pass --run <run-epic-id>", req.Command)))
			}

			switch {
			case explain && dryRun:
				return finish(streams, req, refusal(req, RefuseArguments(
					"--explain and --dry-run select different interpreters; pass one")))
			case explain:
				req.Mode = plan.ModeExplain
			case dryRun:
				req.Mode = plan.ModeDryRun
			default:
				req.Mode = plan.ModeExecute
			}

			built, err := spec.Plan(req)
			if err != nil {
				return finish(streams, req, refusal(req, asRefusal(err)))
			}

			// The plan runner's output is captured rather than streamed.
			// Streaming would put the transcript on stdout ahead of a
			// --json envelope and corrupt it, and capturing is what lets
			// both output forms carry the same transcript — which is the
			// "same facts in both forms" guarantee, not a convenience.
			var transcript strings.Builder
			report, err := (plan.Runner{Exec: runner, Out: &transcript}).Run(context.Background(), built, req.Mode)
			if err != nil {
				return finish(streams, req, refusal(req, asRefusal(err)))
			}

			if req.Mode == plan.ModeExplain {
				return finish(streams, req, Result{
					Command:    req.Command,
					Run:        req.Run,
					Mode:       req.Mode,
					Outcome:    OutcomeClear,
					Summary:    "explained; nothing was executed",
					Transcript: lines(transcript.String()),
				})
			}

			result, err := spec.Interpret(req, report)
			if err != nil {
				return finish(streams, req, refusal(req, asRefusal(err)))
			}
			result.Command = req.Command
			result.Run = req.Run
			result.Mode = req.Mode
			result.Records = report.Records
			result.Transcript = lines(transcript.String())
			return finish(streams, req, result)
		},
	}

	cmd.Flags().BoolVar(&explain, "explain", false,
		"print the invocations this command would run, in order, and execute none of them")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"perform every read this command would perform, and no write")
	cmd.Flags().BoolVar(&asJSON, "json", false,
		"machine-readable output carrying the same facts as the text form")
	cmd.Flags().StringVar(&runID, "run", "",
		"the run epic this command is scoped to")

	if spec.Bare {
		cmd.Flags().BoolVar(&fullResult, "full-result", false,
			"print the full command result — the summary, the facts and the invocations behind them — instead of the answer alone")
	}

	if spec.Flags != nil {
		spec.Flags(cmd.Flags())
	}

	return cmd
}

// finish validates, renders and sets the exit status. It is the only
// place a subcommand's outcome becomes a process exit status.
func finish(streams *ioStreams, req Request, result Result) error {
	if err := result.Validate(); err != nil {
		// A result this package cannot render is a defect in the command
		// that built it, and a defect is not something the caller can
		// repair in place.
		result = refusal(req, RefuseUnrepairable("%s", err.Error()))
	}
	streams.exitCode = int(result.Status())

	// The bare form splits the streams, which is the whole point of it: a
	// caller substituting the command captures the answer and nothing
	// else, and still reads why on stderr when there is no answer. The
	// status carries the distinction either way, so nothing is lost by
	// the report moving off stdout.
	switch {
	case req.Bare && result.Bare != nil:
		return renderBare(streams.out, *result.Bare, req.JSON)
	case req.Bare && result.Status() != StatusClear:
		return Render(streams.err, result, req.JSON)
	}
	return Render(streams.out, result, req.JSON)
}

// refusal turns a refusal into a renderable result.
func refusal(req Request, r *Refusal) Result {
	return Result{
		Command:    req.Command,
		Run:        req.Run,
		Mode:       req.Mode,
		Outcome:    r.Outcome,
		Summary:    r.Reason,
		Resolution: r.Resolution,
	}
}

// asRefusal classifies an error a subcommand or the plan runner returned.
// A [*Refusal] already carries its outcome; anything else is a condition
// the caller cannot repair in place, which is the conservative reading —
// exiting 1 for it would send a run onward to try a repair that cannot
// work.
func asRefusal(err error) *Refusal {
	var r *Refusal
	if errors.As(err, &r) {
		return r
	}
	return RefuseUnrepairable("%s", err.Error())
}

// lines splits captured output into the lines a result carries, dropping
// the trailing empty element a final newline produces.
func lines(s string) []string {
	if s == "" {
		return nil
	}
	out := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	return out
}

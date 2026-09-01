// Package cli is the Swarm CLI's command surface: the command tree, the
// flags every subcommand carries, the mapping from an outcome to one of
// four exit statuses, and the two output forms — per the Swarm Command
// Surface component in specs/architecture/container-swarm-cli.md and
// FEAT-SWARM-CORE-001 of specs/features/swarm-core/swarm-core-frd.md.
//
// It contains no knowledge of the tracker or of git. It resolves
// arguments, asks a command to build a plan, hands the plan to the mode
// the flags selected, and turns what comes back into text or JSON and a
// status.
//
// The exit vocabulary is the reason this package exists as a component
// rather than as plumbing. Every skill in the family routes on the
// status without reading the output, so a command that exits 1 where it
// meant 2 sends a run onward into a state it should have stopped in, and
// nothing downstream can tell.
//
// `--help` text is never written by hand: it is cobra's own rendering of
// the Use, Short and flag declarations, so the declaration and the help
// text cannot drift.
package cli

import (
	"io"

	"github.com/spf13/cobra"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	manifestbeads "github.com/brettcallaghan/agent-plugins/cli/swarm/manifest/beads"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/status"
	beadstracker "github.com/brettcallaghan/agent-plugins/cli/swarm/tracker/beads"
	gitvcs "github.com/brettcallaghan/agent-plugins/cli/swarm/vcs/git"
)

// ioStreams carries the two output streams a run writes to and the exit
// status the run decided on.
//
// exitCode exists because a cobra RunE can only signal failure through a
// non-nil error, which cobra then renders as "Error: ...". Most of this
// CLI's non-zero statuses are not errors — a command reporting a true
// fact about a run is a successful run — so a successful RunE (nil
// error) sets exitCode directly instead, and [Execute] reads it back
// after cobra's Execute returns.
type ioStreams struct {
	out io.Writer
	err io.Writer
	// exitCode is read by Execute once cobra's own Execute call returns
	// nil.
	exitCode int
}

// Execute parses args against the command tree and runs the selected
// subcommand, writing to stdout and stderr and returning the process
// exit status.
//
// It is the one function main calls. A non-nil error from cobra's own
// Execute is an argument error — an unknown flag, an unknown subcommand,
// a bad value — and every subcommand this package builds handles its own
// failures through [Result] rather than by returning one, so a cobra
// error maps to the malformed-invocation status and nothing else does.
func Execute(args []string, stdout, stderr io.Writer) int {
	streams := &ioStreams{out: stdout, err: stderr, exitCode: int(StatusClear)}
	root := newRootCmd(streams)
	register(root, streams, swarmexec.ProcessRunner{})
	return dispatch(root, streams, args, stdout, stderr)
}

// register adds the subcommands to the tree. It is the one place a delta
// adding a command touches, and it takes the process runner every
// subcommand's plan goes through so that a test can supply a fake in its
// place without a second wiring path.
func register(root *cobra.Command, streams *ioStreams, runner swarmexec.Runner) {
	tr := beadstracker.Tracker{}
	store := manifestbeads.Store{Tracker: tr}
	root.AddCommand(manifestCmd(streams, store, runner))
	root.AddCommand(statusCmd(streams, status.Deps{
		Tracker: tr,
		VCS:     gitvcs.VersionControl{},
		Store:   store,
		Runner:  plan.Runner{Exec: runner},
	}))
	root.AddCommand(phaseCmd(streams, tr, runner))
}

// dispatch runs the tree and turns what comes back into a process exit
// status. It is shared by [Execute] and by the tests, so a test asserting
// an exit status is asserting the same mapping the binary uses.
func dispatch(root *cobra.Command, streams *ioStreams, args []string, stdout, stderr io.Writer) int {
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	if err := root.Execute(); err != nil {
		return int(StatusMalformed)
	}
	return streams.exitCode
}

// newRootCmd builds the command tree.
//
// The root carries a RunE rather than relying on cobra's default
// behaviour for a command with no Run, because the two differ in exactly
// the way the binary-skeleton delta specifies: cobra's default prints
// the usage and exits 0, and a bare `swarm` is an invocation naming no
// command, which is the malformed-invocation status. The command tree is
// still printed — it is what tells the caller which invocation would
// have been valid — but to the error stream, so a caller reading stdout
// for a machine-readable answer never receives usage text where a result
// belongs.
func newRootCmd(streams *ioStreams) *cobra.Command {
	root := &cobra.Command{
		Use:   "swarm",
		Short: "Perform a swarm run's mechanized actions against the bead tracker and git",
		Long: "swarm performs a run's actions against the bead tracker and git and reports " +
			"the invocation behind every fact it states, so a caller can re-run any single " +
			"step and route on the exit status without parsing prose.",
		SilenceUsage: true,
		// SilenceErrors is left off: cobra's own one-line rendering of an
		// argument error goes to the error stream, which is where a
		// malformed invocation belongs, and Execute maps it to the
		// malformed-invocation status.
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SetOut(streams.err)
			if err := cmd.Usage(); err != nil {
				return err
			}
			streams.exitCode = int(StatusMalformed)
			return nil
		},
	}
	// cobra's autogenerated `completion` command is no part of this
	// container's subcommand vocabulary, so it is turned off rather than
	// left to appear as an uncounted extra in the tree a bare invocation
	// prints.
	root.CompletionOptions.DisableDefaultCmd = true

	return root
}

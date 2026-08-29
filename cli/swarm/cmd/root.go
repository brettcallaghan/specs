// Command tree wiring for the Swarm CLI's command surface, per the
// binary-skeleton delta of
// specs/features/swarm-core/swarm-core-blueprint.md and the Swarm
// Command Surface component in
// specs/architecture/container-swarm-cli.md.
//
// The tree is empty on purpose. This is the substrate delta: it delivers
// a binary that builds, formats, vets, tests and cross-compiles on the
// matrices ADR-023 puts it on, and no capability a caller can invoke.
// Subcommands, the run scope, --explain, --dry-run and --json arrive
// with the deltas that need them.
//
// Everything in this package is orchestration. It resolves arguments,
// asks a command to build a plan, hands the plan to the mode the flags
// selected, and turns what comes back into text or JSON and an exit
// status; it holds no knowledge of the tracker or of git
// (container-swarm-cli.md, Swarm Command Surface).
//
// `--help` text is never written by hand: it is cobra's own rendering of
// the Use, Short and flag declarations below, so the declaration and the
// help text cannot drift.
package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

// The exit vocabulary specs/features/swarm-core/swarm-core-frd.md fixes
// for every command in this CLI. A caller routes on the status without
// reading the output, so these are a contract rather than a detail.
//
// Only the malformed-invocation status is reachable from this delta:
// there is no command yet that can answer, report a finding, or discover
// a condition the caller cannot repair. The other three are named here
// rather than left to be invented per command, because the mapping from
// an outcome to one of the four belongs to this surface.
const (
	// exitClear — the command answered and its verdict is clear.
	exitClear = 0
	// exitAttention — the command answered, and what it found needs
	// attention; the output names what would resolve it.
	exitAttention = 1
	// exitStop — a condition the caller cannot repair in place.
	exitStop = 2
	// exitMalformed — the invocation itself is malformed: a missing
	// scope, an unknown flag, a bad argument.
	exitMalformed = 3
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
// Execute is an argument error and nothing else at this delta — the tree
// carries no RunE that can fail for another reason — so it maps to the
// malformed-invocation status. A delta adding a command that can fail
// internally is what has to separate the two.
func Execute(args []string, stdout, stderr io.Writer) int {
	streams := &ioStreams{out: stdout, err: stderr, exitCode: exitClear}
	root := newRootCmd(streams)
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)

	if err := root.Execute(); err != nil {
		return exitMalformed
	}
	return streams.exitCode
}

// newRootCmd builds the command tree.
//
// The root carries a RunE rather than relying on cobra's default
// behaviour for a command with no Run, because the two differ in exactly
// the way this delta specifies: cobra's default prints the usage and
// exits 0, and a bare `swarm` is an invocation naming no command, which
// is the malformed-invocation status. The command tree is still printed
// — it is what tells the caller which invocation would have been valid —
// but to the error stream, so a caller reading stdout for a
// machine-readable answer never receives usage text where a result
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
			streams.exitCode = exitMalformed
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

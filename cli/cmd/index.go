package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
)

// newIndexCmd builds the `index` subcommand: build the corpus index and
// write it, per REQ-SPEC-010 (AC-SPEC-011.1).
func newIndexCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var out string

	cmd := &cobra.Command{
		Use:   "index",
		Short: "Build the spec index and write it to disk",
		Long: "index walks the configured spec root, builds the corpus index, and " +
			"writes it to specs.json inside the spec root, to the path named by " +
			"--out, or to standard output.",
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runIndex(streams, *configPath, out)
		},
	}
	cmd.Flags().StringVar(&out, "out", "",
		`where to write the index: a path, "-" for standard output, or omitted for specs.json inside the spec root`)
	return cmd
}

// runIndex builds the corpus index and writes it to dest, per
// AC-SPEC-011.1 and the output-path option AC-SPEC-010.1 through
// AC-SPEC-010.3 define.
//
// It always builds fresh — corpus.Build, not corpus.Open — because
// building the index is the whole point of this subcommand: serving a
// warm answer here would sometimes write back the very file it read,
// which is what corpus.Open exists to avoid for a query, not for a
// build.
//
// AC-SPEC-002.5 ("exit zero after building an index that contains parse
// failures") holds without special-casing anything here: [corpus.Build]
// itself never fails for a member with a parse error — it records the
// failure against that member and indexes the rest — so the only path
// that returns a non-nil error from this function is one where the build
// could not run at all (no profile, no spec root, or an unreadable spec
// root), which is a real operational error rather than a corpus defect.
func runIndex(streams *ioStreams, configPath, out string) error {
	cfg, findings, err := Load(configPath, BuiltinRegistry())
	if err != nil {
		return err
	}
	if findings != nil {
		return reportAndExit(streams, findings, false)
	}

	opts, err := corpusOptions(cfg, configPath)
	if err != nil {
		return err
	}

	built, err := corpus.Build(opts)
	if err != nil {
		return fmt.Errorf("cmd: building the index: %w", err)
	}

	dest := corpus.IndexPath(opts, out)
	if err := corpus.Write(built.Index(), dest, streams.out); err != nil {
		return fmt.Errorf("cmd: writing the index: %w", err)
	}

	// A file was written: name it, so a human running this interactively
	// sees where the artifact landed (AC-SPEC-010.1's "the CLI prints that
	// relative path"). Streaming to stdout already delivered the whole
	// index there; printing the path too would corrupt that stream for a
	// caller piping it into another tool (AC-SPEC-010.3).
	if dest != corpus.StdoutTarget {
		fmt.Fprintln(streams.out, dest)
	}
	return nil
}

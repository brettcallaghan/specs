package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/core/numbering"
)

// newNextIDCmd builds the `next-id` subcommand: derive the next
// identifier for a numbered document type, per AC-SPEC-007.1
// (AC-SPEC-011.4).
func newNextIDCmd(streams *ioStreams, configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next-id <type>",
		Short: "Return the next identifier for a numbered document type",
		Long: "next-id returns one greater than the highest number in use for the " +
			"given document type anywhere in the corpus, per the active profile's " +
			"numbered-type registration.",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runNextID(streams, *configPath, args[0])
		},
	}
	return cmd
}

// runNextID answers [numbering.Next] for docType and renders the result.
//
// docType not being registered as numbered is reported as a failure
// (exit 1) rather than as a fact about the corpus: unlike slug-exists,
// there is no valid identifier this command could otherwise print, so
// the caller asked a question this configuration cannot answer at all.
func runNextID(streams *ioStreams, configPath, docType string) error {
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

	c, err := openCorpus(opts)
	if err != nil {
		return err
	}

	next, err := numbering.Next(c.Index(), docType, cfg.Profile.DocumentTypes())
	if err != nil {
		if errors.Is(err, numbering.ErrNotNumbered) {
			fmt.Fprintf(streams.err, "%q is not registered as a numbered document type\n", docType)
			streams.exitCode = 1
			return nil
		}
		return fmt.Errorf("cmd: deriving the next identifier: %w", err)
	}

	fmt.Fprintln(streams.out, next)
	return nil
}

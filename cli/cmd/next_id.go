package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
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
// docType not being registered as numbered is an invocation error
// (exit 2), naming every numbered type, rather than a fact about the
// corpus: the numbered types are a set the profile registers, and no
// document an author writes makes an unnumbered type numbered
// ([[cli-exit-status-standards]]'s "A value outside a registered set").
func runNextID(streams *ioStreams, configPath, docType string) error {
	cfg, findings, err := Load(configPath, BuiltinRegistry())
	if err != nil {
		return err
	}
	if findings != nil {
		return reportConfigDefects(streams, findings)
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
			return invocationError("%q is not registered as a numbered document type; numbered: %s",
				docType, numberedTypes(cfg.Profile.DocumentTypes()))
		}
		return fmt.Errorf("deriving the next identifier: %w", err)
	}

	fmt.Fprintln(streams.out, next)
	return nil
}

// numberedTypes lists, sorted, the document types types registers as
// numbered — the registered set an unnumbered next-id operand is reported
// against.
func numberedTypes(types map[string]corpus.DocumentTypeRules) string {
	var names []string
	for name, rules := range types {
		if rules.Numbered {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "none are registered"
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/specs/cli/core/corpus"
	"github.com/brettcallaghan/specs/cli/core/numbering"
)

// newNextIDCmd builds the `next-id` subcommand: derive the next
// identifier for a numbered document type, per AC-SPEC-007.1
// (AC-SPEC-011.4).
func newNextIDCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "next-id <type>",
		Short: "Return the next identifier for a numbered document type",
		Long: "next-id answers the next identifier for the given document type: a " +
			"YAML mapping holding number, one greater than the highest number in use " +
			"for that type anywhere in the corpus, and slug, the slug that number " +
			"forms under the type's registered width (adr-032). --json writes the " +
			"same values as JSON. A type the active profile does not register as " +
			"numbered is refused.",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runNextID(streams, *configPath, args[0], jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write the answer as JSON instead of YAML")
	return cmd
}

// nextIDResult is next-id's answer: the next number and the slug it forms
// (AC-SPEC-007.1).
type nextIDResult struct {
	Number int    `json:"number"`
	Slug   string `json:"slug"`
}

// runNextID answers [numbering.Next] for docType, formats the slug that
// number forms with [numbering.Format], and renders both.
//
// docType not being registered as numbered is an invocation error
// (exit 2), naming every numbered type, rather than a fact about the
// corpus: the numbered types are a set the profile registers, and no
// document an author writes makes an unnumbered type numbered
// ([[cli-exit-status-standards]]'s "A value outside a registered set").
func runNextID(streams *ioStreams, configPath, docType string, jsonOut bool) error {
	cfg, configPath, findings, err := loadConfig(configPath)
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

	rules := cfg.Profile.DocumentTypes()
	next, err := numbering.Next(c.Index(), docType, rules)
	if err == nil {
		var slug string
		slug, err = numbering.Format(docType, next, rules)
		if err == nil {
			return renderNextID(streams, nextIDResult{Number: next, Slug: slug}, jsonOut)
		}
	}
	if errors.Is(err, numbering.ErrNotNumbered) {
		return invocationError("%q is not registered as a numbered document type; numbered: %s",
			docType, numberedTypes(rules))
	}
	return fmt.Errorf("deriving the next identifier: %w", err)
}

// renderNextID writes the answer as YAML, or as JSON under --json, with
// the same keys in the same order.
func renderNextID(streams *ioStreams, result nextIDResult, asJSON bool) error {
	if asJSON {
		data, err := json.Marshal(result)
		if err != nil {
			return internalError("rendering the next identifier as JSON: %v", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}
	fmt.Fprint(streams.out, renderYAML(result))
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

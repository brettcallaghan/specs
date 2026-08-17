package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// newRefsCmd builds the `refs` subcommand: report which corpus members
// reference a given slug, or which documents that slug itself
// references, per AC-SPECREF-003.1 through AC-SPECREF-003.5.
//
// This is the fifth subcommand FEAT-SPEC-011 anticipated growing into
// (container-spec-cli.md's Change Guidance); it neither renames nor
// repurposes the four `newRootCmd` already builds.
func newRefsCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var jsonOut, outbound bool

	cmd := &cobra.Command{
		Use:   "refs <slug>",
		Short: "Report which corpus members reference a slug, or what it references",
		Long: "refs reports every corpus member whose body or registered frontmatter " +
			"properties reference the given slug. Given --outbound, it reports the " +
			"documents that slug references instead. A slug no corpus member declares " +
			"is reported distinctly from one that is declared but unreferenced; both " +
			"exit zero.",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runRefs(streams, *configPath, args[0], outbound, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write each reference as JSON, with its file path and span, to standard output")
	cmd.Flags().BoolVar(&outbound, "outbound", false,
		"report the documents the slug references, instead of the documents referencing it")
	return cmd
}

// refsResult is what runRefs renders: [refsResult.References] read
// straight through in --json mode, and one line per entry otherwise.
//
// Declared is what makes "no corpus member declares this slug" and "the
// slug is declared but has no references in the asked direction"
// distinguishable — two different facts about the corpus, both exiting
// zero (AC-SPECREF-003.5 and the Design section's undeclared-slug
// bullet) — rather than an empty References list meaning either one.
type refsResult struct {
	Slug       string                `json:"slug"`
	Direction  string                `json:"direction"`
	Declared   bool                  `json:"declared"`
	References []specmodel.Reference `json:"references"`
}

// runRefs answers [corpus.Corpus.Slug] and, for a declared slug,
// [corpus.Corpus.InboundReferences] or [corpus.Corpus.OutboundReferences],
// and renders the result. It is the whole of what this subcommand does:
// converting a slug and a direction into a corpus lookup and the result
// into stdout, with no domain logic of its own.
//
// It always exits zero: like slug-exists, what refs answers is a fact
// about the corpus, never a failure of the command that asked it
// (AC-SPECREF-003.5, and the undeclared-slug case matching the
// slug-exists contract per AC-SPEC-007.2). The corpus is opened warm
// (corpus.Open), the same as check, slug-exists and next-id, because this
// is exactly the corpus query FEAT-SPEC-009.1's latency budget is measured
// against (AC-SPECREF-003.4).
func runRefs(streams *ioStreams, configPath, slug string, outbound, jsonOut bool) error {
	cfg, findings, err := Load(configPath, BuiltinRegistry())
	if err != nil {
		return err
	}
	if findings != nil {
		return reportAndExit(streams, findings, jsonOut)
	}

	opts, err := corpusOptions(cfg, configPath)
	if err != nil {
		return err
	}

	c, err := openCorpus(opts)
	if err != nil {
		return err
	}

	direction := "inbound"
	if outbound {
		direction = "outbound"
	}

	entry, declared := c.Slug(slug)
	// References defaults to an empty (never nil) slice, so the JSON
	// rendering always writes "references":[] rather than "references":null
	// for either zero-result case — the same "collections are never nil"
	// discipline [finding.Report.Model] applies to Findings and Checks.
	result := refsResult{Slug: slug, Direction: direction, Declared: declared, References: []specmodel.Reference{}}
	if declared {
		var found []specmodel.Reference
		if outbound {
			found = c.OutboundReferences(entry.Path)
		} else {
			found = c.InboundReferences(entry.Path)
		}
		if found != nil {
			result.References = found
		}
	}

	return renderRefs(streams, result, jsonOut)
}

// renderRefs writes result to streams.out: the generated [specmodel.Reference]
// JSON — each entry already carrying its file path and span
// (AC-SPECREF-003.3) — in --json mode, and a short human summary
// otherwise.
//
// Text rendering of a non-empty result mirrors [finding.Report.Text]'s
// one-line-per-item, path:line:col shape, so a reference and a finding
// read the same way in a terminal.
func renderRefs(streams *ioStreams, result refsResult, asJSON bool) error {
	if asJSON {
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("cmd: rendering the reference set as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}

	if !result.Declared {
		fmt.Fprintf(streams.out, "%s: not declared by any corpus member\n", result.Slug)
		return nil
	}
	if len(result.References) == 0 {
		if result.Direction == "outbound" {
			fmt.Fprintf(streams.out, "%s: references nothing\n", result.Slug)
		} else {
			fmt.Fprintf(streams.out, "%s: referenced by no corpus member\n", result.Slug)
		}
		return nil
	}
	for _, r := range result.References {
		fmt.Fprintf(streams.out, "%s:%d:%d  %s\n",
			r.Path, r.Position.StartLine, r.Position.StartColumn, r.Text)
	}
	return nil
}

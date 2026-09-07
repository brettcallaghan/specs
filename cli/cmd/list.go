package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
)

// newListCmd builds the `list` subcommand: report a summary of every
// corpus member, filtered by document type and by whether documents no
// longer in effect are wanted, per AC-SPECLIST-001.1 through
// AC-SPECLIST-004.5.
//
// The tenth subcommand, wired like the nine before it: flags in, one
// Document Core call, one rendering out. It decides nothing about what a
// summary is or which members qualify — [[container-spec-cli]]'s Change
// Guidance puts that in Document Core, so that a future API or MCP
// wrapper gets the same answer without going through cobra.
func newListCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var types []string
	var all, jsonOut bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Report every corpus member's path, slug, type, status, title and description",
		Long: "list answers with one entry per corpus member, carrying its file path, " +
			"slug, type, status, title and description — the frontmatter summary that " +
			"lets a reader decide which specifications to open without opening any of " +
			"them. Entries come back in the index's own order, which is fixed by " +
			"repository-relative path, so two runs over an unchanged corpus print the " +
			"same bytes. Documents at a status the active profile registers as no " +
			"longer in effect are omitted unless --all asks for them, because a " +
			"superseded document read as current is worse than one not read at all. A " +
			"member whose frontmatter could not be parsed is reported with its path " +
			"and that failure rather than being skipped.",
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runList(streams, *configPath, types, all, jsonOut)
		},
	}
	cmd.Flags().StringSliceVar(&types, "type", nil,
		"report only members of the named document types; repeatable or comma-separated")
	cmd.Flags().BoolVar(&all, "all", false,
		"report members at every status, including ones no longer in effect")
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write the summaries as JSON instead of YAML")
	return cmd
}

// listResult is the answer: the summaries and nothing else, under the one
// key `documents`.
//
// No count, no echo of the filters and no outcome line, per
// [[cli-output-standards]]'s "the answer is the properties the command
// established... and SHALL contain nothing else" (AC-SPECLIST-004.5).
// Documents is never nil, so an empty answer renders as `documents: []`
// rather than as null — the empty collection AC-SPECLIST-001.7 requires
// be an answer.
type listResult struct {
	Documents []corpus.DocumentSummary `json:"documents"`
}

// runList opens the corpus warm and hands the query its options.
//
// Warm, through the shared [openCorpus], for the reason `check` is: a
// listing is a query, every value it reports is already in the index, and
// the persisted index at the default location answers it without
// re-walking the corpus.
//
// It exits zero for any answer, including an empty one — a filter
// matching nothing is a fact about the corpus rather than a failure of
// the invocation. The one non-zero exit is a `--type` value the active
// profile does not register, which the query reports rather than
// answering: cobra writes that report to standard error and [Execute]
// returns 1 (AC-SPECLIST-002.4). Answering it with an empty collection
// instead would make a typo indistinguishable from an empty category.
func runList(streams *ioStreams, configPath string, types []string, all, jsonOut bool) error {
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

	summaries, err := c.DocumentSummaries(corpus.SummaryOptions{
		Types:       types,
		AllStatuses: all,
	})
	if err != nil {
		return err
	}

	return renderList(streams, listResult{Documents: summaries}, jsonOut)
}

// renderList writes the answer as YAML, or as JSON under --json, both
// carrying the same values in the same order (AC-SPECLIST-004.1,
// AC-SPECLIST-004.2).
//
// [renderYAML] is used unchanged: it already implements the YAML style
// rules, including the double-quoting that keeps a title carrying ": "
// reading back as written (AC-SPECLIST-004.4), and it resolves fields
// from the same json tags `encoding/json` does, so the two renderings
// cannot carry different keys.
func renderList(streams *ioStreams, result listResult, asJSON bool) error {
	if asJSON {
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("cmd: rendering the summaries as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}
	fmt.Fprint(streams.out, renderYAML(result))
	return nil
}

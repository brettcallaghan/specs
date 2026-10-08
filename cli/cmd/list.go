package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	cmd.Flags().Var(typeFilterValue{values: &types}, "type",
		"report only members of the named document types; repeatable or comma-separated")
	cmd.Flags().BoolVar(&all, "all", false,
		"report members at every status, including ones no longer in effect")
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write the summaries as JSON instead of YAML")
	return cmd
}

// typeFilterValue collects the `--type` values without losing an empty
// one, so that the caller's question reaches Document Core as it was
// asked.
//
// pflag's own string-slice flag runs a CSV reader over each occurrence,
// and a CSV reader given "" yields no field at all. `--type ""` therefore
// arrived as a slice with no elements, indistinguishable from the flag
// never having been passed, and Document Core answered "no type filter"
// with the whole corpus — the opposite of the rejection AC-SPECLIST-002.4
// requires, and the most dangerous answer to a `--type "$KIND"` whose
// KIND is unset. An empty occurrence was dropped even beside a good one,
// so `--type "" --type adr` could not be told from `--type adr`.
//
// Splitting on "," by hand keeps every element the caller wrote, the
// empty one included, and `resolveTypeFilter` then rejects it through the
// path that already rejects " adr". A plain [strings.Split], not a CSV
// reader: the CSV reader is the thing that drops the value, and document
// type names are bare identifiers carrying no commas or quotes.
//
// The flag decides nothing about which members qualify — it only stops
// destroying the question — so [[container-spec-cli]]'s placement of that
// decision in Document Core is undisturbed, and an absent flag still
// leaves the slice nil, which is what `resolveTypeFilter` reads as "no
// type filter".
type typeFilterValue struct{ values *[]string }

// Set appends every comma-separated element of one occurrence, including
// an empty one.
func (v typeFilterValue) Set(raw string) error {
	*v.values = append(*v.values, strings.Split(raw, ",")...)
	return nil
}

// Type names the flag's kind, and the name is deliberately NOT one pflag
// knows. This flag is not a pflag stringSlice: it splits on a plain comma
// instead of through a CSV reader, and it keeps empty elements. pflag
// routes its typed accessors by this string, so borrowing "stringSlice"
// handed a value with different semantics to pflag's stringSlice
// converter — which panicked on the empty string [typeFilterValue.String]
// renders for an unset flag, and read "[]" back as no values at all,
// silently discarding the empty occurrence this type exists to preserve.
// An unrecognised name makes pflag refuse such a read instead of
// mis-parsing it.
//
// The usage line is unaffected, because pflag's UnquoteUsage maps only
// its own type names to placeholders ("stringSlice" to `strings`,
// "intSlice" to `ints`, and so on) and renders any other name verbatim.
// "strings" is therefore the same placeholder pflag used to derive, and
// `specs list --help` is byte-identical to what "stringSlice" produced.
//
// The one thing the rename costs is cobra's flag-name completion
// repeatability: cobra decides a flag may be offered again by testing
// whether its value implements a GetSlice method, or whether its type
// name contains "Slice" or "Array" or begins with "stringTo". "strings"
// matches none of those, where "stringSlice" matched on the name. Nothing
// reaches that code today — root.go sets CompletionOptions.DisableDefaultCmd,
// so the CLI ships no way to generate the scripts that call `__complete`
// — and implementing the method for it now would be dead code. If that
// command is ever restored, add it then.
func (v typeFilterValue) Type() string { return "strings" }

// String renders the collected values, and renders an unset flag as the
// empty string rather than as "[]" so that pflag reads the default as a
// zero value and prints no `(default ...)` for a flag that has none.
//
// "[]" is not returned for an unset flag because pflag captures DefValue
// by calling String() at registration, and its defaultIsZeroValue reaches
// a custom [pflag.Value] only through its default branch, which counts
// "", "0", "false" and "<nil>" as a zero default and nothing else. "[]"
// is zero for pflag's own slice types alone, so returning it here would
// print `(default [])` on the usage line.
func (v typeFilterValue) String() string {
	if v.values == nil || len(*v.values) == 0 {
		return ""
	}
	return "[" + strings.Join(*v.values, ",") + "]"
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
// the invocation. A `--type` value the active profile does not register
// is an invocation error (exit 2), reported on standard error with every
// registered type rather than answered (AC-SPECLIST-002.4): answering it
// with an empty collection would make a typo indistinguishable from an
// empty category.
func runList(streams *ioStreams, configPath string, types []string, all, jsonOut bool) error {
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

	summaries, err := c.DocumentSummaries(corpus.SummaryOptions{
		Types:       types,
		AllStatuses: all,
	})
	if err != nil {
		if errors.Is(err, corpus.ErrUnregisteredType) {
			return invocationError("%w", withoutPrefix(err, "corpus: "))
		}
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
			return fmt.Errorf("rendering the summaries as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}
	fmt.Fprint(streams.out, renderYAML(result))
	return nil
}

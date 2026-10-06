package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// newHeadingsCmd builds the `headings` subcommand: answer a heading
// selection with every match's path, text, identifier, level and the span
// of the content beneath it, per AC-SPECSTR-003.1 through
// AC-SPECSTR-003.6.
//
// One flat command with no subcommand of its own, as every command on
// this surface is. It is the query that makes agent review auditable as
// well as cheap: two runs given the same selection read the same spans.
func newHeadingsCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var jsonOut, emit bool

	cmd := &cobra.Command{
		Use:   "headings [<document>#<heading>]",
		Short: "Report the headings matching a selection, with the span of each one's section",
		Long: "headings answers a heading selection with every match's file path, heading " +
			"text, heading identifier, level, and the span of the content beneath it. " +
			"The selection is one operand written <document>#<heading>, either half of " +
			"which may be omitted: <document> alone selects every heading in that " +
			"document, #<heading> selects that heading in every corpus member, and an " +
			"absent operand selects every heading in the corpus. A document half " +
			"holding a path separator or ending in .md is read as a repository-relative " +
			"path, and otherwise as a slug. A heading is matched on its identifier, so " +
			"letter case and punctuation do not decide a match, and a matched heading " +
			"is returned with every heading nested inside its section. Given --emit, " +
			"the content each match stands over is written as itself instead.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			operand := ""
			if len(args) == 1 {
				operand = args[0]
			}
			if emit && jsonOut {
				return errNoEmitWithJSON
			}
			return runHeadings(streams, *configPath, operand, emit, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write the matches as JSON instead of YAML")
	cmd.Flags().BoolVar(&emit, "emit", false,
		"write the content each matched heading stands over, instead of the matches")
	return cmd
}

// errNoEmitWithJSON rejects the one flag combination that has no meaning:
// --emit writes a document payload as itself, which has no JSON form to
// select.
var errNoEmitWithJSON = errors.New("specs headings: --json cannot be combined with --emit")

// headingMatch is one heading a selection matched, carrying exactly what
// AC-SPECSTR-003.1 names.
//
// Section is the span of the content beneath the heading and not the
// content itself: the index records no section content, and --emit reads
// the text back through this span.
type headingMatch struct {
	Path       string         `json:"path"`
	Text       string         `json:"text"`
	Identifier string         `json:"identifier"`
	Level      uint64         `json:"level"`
	Section    specmodel.Span `json:"section"`
}

// headingsResult is the answer: the matches and nothing else.
//
// No count, no echo of the selection and no outcome line, per
// [[cli-output-standards]]'s "the answer is the properties the command
// established... and SHALL contain nothing else". Matches is never nil,
// so an empty answer renders as `matches: []` rather than as null — the
// empty collection that AC-SPECSTR-003.6 requires be an answer.
type headingsResult struct {
	Matches []headingMatch `json:"matches"`
}

// selection is a parsed `<document>#<heading>` operand. An empty half
// selects everything at that level.
type selection struct {
	document string
	heading  string
}

// parseSelection splits the operand on its first "#".
//
// An operand with no "#" is a document, not a heading: `specs headings
// adr-012` asks for that document's headings, and `specs headings
// '#overview'` asks for that heading across the corpus.
func parseSelection(operand string) selection {
	document, heading, found := strings.Cut(operand, "#")
	if !found {
		return selection{document: strings.TrimSpace(operand)}
	}
	return selection{document: strings.TrimSpace(document), heading: strings.TrimSpace(heading)}
}

// isPath reports whether the selection's document half names a file
// rather than a slug: it does when it holds a path separator or ends in
// ".md".
//
// A slug never contains either, so the test cannot misread one — and a
// document is addressable both ways because a caller reviewing a diff has
// paths in hand while one following a reference has slugs.
func (s selection) isPath() bool {
	return strings.ContainsAny(s.document, `/\`) || strings.HasSuffix(s.document, ".md")
}

// runHeadings converts the selection into one corpus query and renders
// the result, with no domain logic of its own: the heading tree and the
// identifier derivation are Markdown Structure's, and the entries are the
// corpus's.
//
// The corpus is opened warm, as check, refs, slug-exists and next-id do,
// because a heading query is exactly the corpus query FEAT-SPEC-009.1's
// latency budget is measured against — and because the headings are in
// the index, a warm open answers without reading a document off disk.
//
// It always exits zero. A selection matching nothing is a fact about the
// corpus rather than a failure of the invocation (AC-SPECSTR-003.6).
func runHeadings(streams *ioStreams, configPath, operand string, emit, jsonOut bool) error {
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

	sel := parseSelection(operand)
	matches := selectHeadings(c.Index(), sel)

	if emit {
		return emitSections(streams, opts.RepoRoot, matches)
	}
	return renderHeadings(streams, headingsResult{Matches: matches}, jsonOut)
}

// selectHeadings answers sel from the index, in document order within a
// member and in the index's own path order across members.
//
// Ordering is the index's throughout rather than the command's, so two
// runs over an unchanged corpus print identical bytes for the same
// reason the index itself is byte-identical (AC-SPEC-009.7).
func selectHeadings(index specmodel.SpecIndex, sel selection) []headingMatch {
	matches := []headingMatch{}
	wanted := markdown.Identifier(sel.heading)

	for _, entry := range index.Entries {
		if !sel.selects(entry) {
			continue
		}

		// chosen marks the headings this member contributes, so that a
		// heading reached both as a match and as one nested inside
		// another match is reported once.
		chosen := make([]bool, len(entry.Headings))
		for i, heading := range entry.Headings {
			if sel.heading != "" && markdown.Identifier(heading.Text) != wanted {
				continue
			}
			chosen[i] = true
			// A matched heading carries the headings nested beneath it,
			// each with its own level (AC-SPECSTR-003.4): every heading
			// after it and deeper than it, up to the next heading at its
			// own level or above.
			for j := i + 1; j < len(entry.Headings); j++ {
				if entry.Headings[j].Level <= heading.Level {
					break
				}
				chosen[j] = true
			}
		}

		for i, heading := range entry.Headings {
			if !chosen[i] {
				continue
			}
			matches = append(matches, headingMatch{
				Path:       entry.Path,
				Text:       heading.Text,
				Identifier: markdown.Identifier(heading.Text),
				Level:      heading.Level,
				Section:    heading.SectionSpan,
			})
		}
	}
	return matches
}

// selects reports whether entry is within the selection's document half.
// An empty half selects every corpus member (AC-SPECSTR-003.3).
func (s selection) selects(entry specmodel.IndexEntry) bool {
	if s.document == "" {
		return true
	}
	if s.isPath() {
		return normalizeOperand(entry.Path) == normalizeOperand(s.document)
	}
	// A slug, compared without regard to letter case, the same way the
	// corpus's own slug lookup compares one.
	if entry.Frontmatter == nil {
		return false
	}
	return strings.EqualFold(string(entry.Frontmatter.Slug), s.document)
}

// renderHeadings writes the answer as YAML, or as JSON under --json, both
// carrying the same values in the same order.
func renderHeadings(streams *ioStreams, result headingsResult, asJSON bool) error {
	if asJSON {
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("cmd: rendering the headings as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}
	fmt.Fprint(streams.out, renderYAML(result))
	return nil
}

// emitSections writes the content each matched heading stands over, read
// back through the recorded span (AC-SPECSTR-003.5).
//
// A document payload, written as itself: no envelope, no heading naming
// each section and no separator between them, per
// [[cli-output-standards]]. Each member is read once however many of its
// sections matched, and its bytes are line-ending-normalized exactly as
// the walk normalized them, because the spans were computed against the
// normalized form and would otherwise be off by one byte per line on a
// CRLF checkout.
//
// A section already contained in one just written is skipped. The
// matches carry every heading nested inside a match (AC-SPECSTR-003.4),
// and a parent's section span already covers its children's text, so
// emitting each match's span in turn would write the nested text a
// second time — and this query exists to spend a reader's attention
// once on the text it selected. What is skipped is only ever a copy:
// the bytes are in the enclosing section that was written.
func emitSections(streams *ioStreams, repoRoot string, matches []headingMatch) error {
	sources := map[string][]byte{}
	var writtenPath string
	var writtenEnd uint64

	for _, match := range matches {
		if match.Path == writtenPath && match.Section.Offset < writtenEnd {
			continue
		}

		src, ok := sources[match.Path]
		if !ok {
			data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(match.Path)))
			if err != nil {
				return fmt.Errorf("cmd: reading %s: %w", match.Path, err)
			}
			src = normalizeLineEndings(data)
			sources[match.Path] = src
		}
		if _, err := streams.out.Write(span.Slice(src, match.Section)); err != nil {
			return fmt.Errorf("cmd: writing the section of %q: %w", match.Text, err)
		}
		writtenPath, writtenEnd = match.Path, match.Section.Offset+match.Section.Length
	}
	return nil
}

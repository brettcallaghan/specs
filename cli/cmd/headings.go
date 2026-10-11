package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/specs/cli/core/envelope"
	"github.com/brettcallaghan/specs/cli/core/markdown"
	"github.com/brettcallaghan/specs/cli/core/references"
	"github.com/brettcallaghan/specs/cli/core/span"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
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
	var jsonOut, emit, all, full bool

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
			"letter case and punctuation do not decide a match. The heading half may be " +
			"a path of headings separated by #, outermost first, each nested within " +
			"the one before it at any depth, so <document>#Components#References " +
			"selects the References heading beneath Components; every match carries " +
			"its anchor and a ref, which pastes back as a selection, naming that heading " +
			"alone unless the anchor is a single segment another heading shares. A matched heading " +
			"is returned with every heading nested inside its section. Given --emit, " +
			"the content the one matching heading stands over is written as itself " +
			"instead; --emit requires the selection to match exactly one heading, " +
			"and a selection matching none is refused, naming whether the document " +
			"is undeclared or the heading absent, " +
			"and a selection matching several is refused with each match named: those a " +
			"selection matches alone as selections to choose from, and the rest as " +
			"reachable only through the span the query without --emit returns. A " +
			"selection with no heading half matches " +
			"every heading of the document, so a caller after a whole document reads " +
			"the file. The answer is compact by default (anchor, ref, level and " +
			"start_line) and --full returns every field, including the span of each " +
			"section; --json changes the syntax, never the view. Members at a status no longer in effect are omitted unless " +
			"--all asks for them, except a document the selection names itself.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			operand := ""
			if len(args) == 1 {
				operand = args[0]
			}
			if emit && jsonOut {
				return invocationError("%w", errNoEmitWithJSON)
			}
			if emit && full {
				return invocationError("%w", errNoFullWithEmit)
			}
			return runHeadings(streams, *configPath, operand, emit, all, full, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write the matches as JSON instead of YAML")
	cmd.Flags().BoolVar(&all, "all", false,
		"report headings of members at every status, including ones no longer in effect")
	cmd.Flags().BoolVar(&full, "full", false,
		"return every field of every match, including the span of its section")
	cmd.Flags().BoolVar(&emit, "emit", false,
		"write the content the one matching heading stands over, instead of the match")
	return cmd
}

// errNoEmitWithJSON rejects the one flag combination that has no meaning:
// --emit writes a document payload as itself, which has no JSON form to
// select.
var errNoEmitWithJSON = errors.New("--json cannot be combined with --emit")

// errNoFullWithEmit rejects --full beside --emit for the same reason: an
// emitted section is a document payload, to which neither view applies
// (AC-SPECSTR-003.17).
var errNoFullWithEmit = errors.New("--full cannot be combined with --emit")

// headingMatch is one heading a selection matched, carrying exactly what
// AC-SPECSTR-003.1 names.
//
// Section is the span of the content beneath the heading and not the
// content itself: the index records no section content, and --emit reads
// the text back through this span.
type headingMatch struct {
	Path       string `json:"path"`
	Text       string `json:"text"`
	Identifier string `json:"identifier"`
	// Anchor is the identifier of each enclosing heading from the
	// outermost down, then the heading's own, joined by "#"
	// (AC-SPECSTR-003.9): the part after the "#" of a reference.
	Anchor string `json:"anchor"`
	// Ref is the member's declared slug, "#" and Anchor, or its path in
	// place of the slug where it declares none. Pasted into specs
	// headings, specs refs or a [[...]] wikilink it selects this heading
	// alone, unless a sibling shares its text all the way up.
	Ref     string         `json:"ref"`
	Level   uint64         `json:"level"`
	Section specmodel.Span `json:"section"`

	// headingLine is the line the heading itself stands on, which the
	// full match does not carry (its Section begins after that line).
	// Unexported, so neither rendering writes it; the compact view
	// reads it as start_line.
	headingLine uint64
}

// headingsResult is the answer: the matches, and whether the document the
// selection names is one the corpus declares.
//
// No count, no echo of the selection and no outcome line, per
// [[cli-output-standards]]'s "the answer is the properties the command
// established... and SHALL contain nothing else". Matches is never nil,
// so an empty answer renders as `matches: []` rather than as null — the
// empty collection that AC-SPECSTR-003.6 requires be an answer.
//
// Declared is present exactly when the selection has a document half, and
// is false when no corpus member is that document (AC-SPECSTR-003.8), so
// a heading that matches nothing reads apart from a document that does
// not exist. A selection with no document half names no document and
// carries no such key.
type headingsResult struct {
	Declared *bool          `json:"declared,omitempty"`
	Matches  []headingMatch `json:"matches"`
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
func runHeadings(streams *ioStreams, configPath, operand string, emit, all, full, jsonOut bool) error {
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

	sel := parseSelection(operand)
	// A member no longer in effect is withheld unless every status was
	// asked for (AC-SPECSTR-003.14); the same filter serves the query and
	// the --emit refusal that lists its matches.
	var inEffect func(specmodel.IndexEntry) bool
	if !all {
		inEffect = c.InEffect
	}
	matches, direct := selectHeadings(c.Index(), sel, inEffect)

	if emit {
		return emitSection(streams, c.Index(), opts.RepoRoot, operand, sel, inEffect, matches, direct)
	}
	result := headingsResult{Matches: matches}
	if sel.document != "" {
		result.Declared = declared(sel.declaredIn(c.Index()))
	}
	if full {
		return renderHeadings(streams, result, jsonOut)
	}
	return renderHeadings(streams, compactHeadings(result), jsonOut)
}

// headingsCompactResult is the default answer: the properties the query
// established, exactly as [headingsResult] carries them, over the matches
// in the compact view of [[cli-output-standards]] (AC-SPECSTR-003.16).
type headingsCompactResult struct {
	Declared *bool                 `json:"declared,omitempty"`
	Matches  []headingCompactMatch `json:"matches"`
}

// headingCompactMatch keeps what a caller needs to locate and re-select a
// heading: its anchor, its ref, its level and the line the heading itself
// stands on. The path, text, identifier and the section span are dropped;
// the ref carries the document and the anchor the identifier.
type headingCompactMatch struct {
	Anchor    string `json:"anchor"`
	Ref       string `json:"ref"`
	Level     uint64 `json:"level"`
	StartLine uint64 `json:"start_line"`
}

// compactHeadings derives the compact view from the full answer, so the
// two list the same matches in the same order by construction
// (AC-SPECSTR-003.16).
func compactHeadings(result headingsResult) headingsCompactResult {
	compact := headingsCompactResult{
		Declared: result.Declared,
		Matches:  make([]headingCompactMatch, 0, len(result.Matches)),
	}
	for _, match := range result.Matches {
		compact.Matches = append(compact.Matches, headingCompactMatch{
			Anchor:    match.Anchor,
			Ref:       match.Ref,
			Level:     match.Level,
			StartLine: match.headingLine,
		})
	}
	return compact
}

// declaredIn reports whether some corpus member is the document the
// selection's document half names.
func (s selection) declaredIn(index specmodel.SpecIndex) bool {
	for _, entry := range index.Entries {
		if s.selects(entry) {
			return true
		}
	}
	return false
}

// selectHeadings answers sel from the index, in document order within a
// member and in the index's own path order across members.
//
// Ordering is the index's throughout rather than the command's, so two
// runs over an unchanged corpus print identical bytes for the same
// reason the index itself is byte-identical (AC-SPEC-009.7).
//
// direct runs parallel to the matches and marks the ones the selection
// itself names, as opposed to those carried only because they are nested
// inside a match (AC-SPECSTR-003.4). A selection with no heading half
// names every heading of the documents it selects (AC-SPECSTR-003.3).
func selectHeadings(index specmodel.SpecIndex, sel selection, inEffect func(specmodel.IndexEntry) bool) (matches []headingMatch, direct []bool) {
	matches = []headingMatch{}

	for _, entry := range index.Entries {
		if !sel.selects(entry) {
			continue
		}
		// A document named explicitly is answered whatever its status
		// (AC-SPECSTR-003.15): the caller asked for it.
		if sel.document == "" && inEffect != nil && !inEffect(entry) {
			continue
		}

		// chosen marks the headings this member contributes, so that a
		// heading reached both as a match and as one nested inside
		// another match is reported once.
		chosen := make([]bool, len(entry.Headings))
		named := make([]bool, len(entry.Headings))
		paths := references.HeadingPaths(entry.Headings)
		// A reference is addressed by the declared slug; a member that
		// declares none is reached only by its path.
		refPrefix := envelope.DeclaredSlug(entry)
		if refPrefix == "" {
			refPrefix = entry.Path
		}
		// The heading half is a heading path query: one segment is
		// matched on its identifier, and several select the heading
		// named by the last beneath each earlier one, at any depth
		// (AC-SPECSTR-003.10, AC-SPECSTR-003.11) — or, where they spell
		// some heading's full path, that heading alone.
		var selected []bool
		if sel.heading != "" {
			selected = markdown.SelectPaths(sel.heading, paths)
		}
		for i, heading := range entry.Headings {
			if sel.heading != "" && !selected[i] {
				continue
			}
			chosen[i] = true
			named[i] = true
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
			anchor := markdown.JoinPath(paths[i])
			matches = append(matches, headingMatch{
				Path:        entry.Path,
				Text:        heading.Text,
				Identifier:  paths[i][len(paths[i])-1],
				Anchor:      anchor,
				Ref:         refPrefix + "#" + anchor,
				Level:       heading.Level,
				Section:     heading.SectionSpan,
				headingLine: heading.Span.StartLine,
			})
			direct = append(direct, named[i])
		}
	}
	return matches, direct
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
	// corpus's own slug lookup compares one — including a slug recovered
	// from a member whose frontmatter did not parse (AC-SPEC-003.7).
	slug := envelope.DeclaredSlug(entry)
	return slug != "" && strings.EqualFold(slug, s.document)
}

// renderHeadings writes the answer as YAML, or as JSON under --json, both
// carrying the same values in the same order.
func renderHeadings(streams *ioStreams, result any, asJSON bool) error {
	if asJSON {
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("rendering the headings as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}
	fmt.Fprint(streams.out, renderYAML(result))
	return nil
}

// emitSection writes the content the one heading the selection names
// stands over, read back through the recorded span (AC-SPECSTR-003.5).
//
// A document payload, written as itself: no envelope, no heading and no
// summary, per [[cli-output-standards]]. That leaves a payload no place
// to mark where one section ends and the next begins, so --emit answers
// exactly one heading. The headings nested inside it are carried as
// matches but are not counted, since its span already covers their text.
//
// No heading named is refused, writing nothing to standard output
// (AC-SPECSTR-003.13): raw text cannot carry the `declared` property or an
// empty `matches`, so an empty output would read the same for an absent
// heading, an undeclared document and an empty section. The message says
// which of the first two it was. More than one is an invocation error that writes nothing to standard
// output and names every heading the selection named, so the caller can
// narrow it (AC-SPECSTR-003.7). Under "Select one of:" go only the
// matches whose ref, pasted back, names that heading
// alone — tested by running that very selection, not predicted. The rest
// are listed apart: a one-segment heading path is an identifier
// comparison and full-path precedence applies to several segments only
// (AC-SPECSTR-003.11), so a top-level heading of a document with no H1
// that shares its identifier with another heading has no selection of its
// own, and neither do two siblings sharing text all the way up. Its
// content is read through the span the query without --emit returns.
//
// The member's bytes are line-ending-normalized exactly as the walk
// normalized them, because the span was computed against the normalized
// form and would otherwise be off by one byte per line on a CRLF
// checkout.
func emitSection(streams *ioStreams, index specmodel.SpecIndex, repoRoot, operand string, sel selection, inEffect func(specmodel.IndexEntry) bool, matches []headingMatch, direct []bool) error {
	var named []headingMatch
	for i, match := range matches {
		if direct[i] {
			named = append(named, match)
		}
	}

	switch len(named) {
	case 0:
		if sel.document != "" && !sel.declaredIn(index) {
			return invocationError("%q: no corpus member is the document it names (%q); --emit needs exactly one heading", operand, sel.document)
		}
		return invocationError("%q: the selection matches no heading; --emit needs exactly one", operand)
	case 1:
	default:
		var alone, shared []headingMatch
		for _, match := range named {
			if selectsAlone(index, match, inEffect) {
				alone = append(alone, match)
			} else {
				shared = append(shared, match)
			}
		}
		var b strings.Builder
		fmt.Fprintf(&b, "%q matches %d headings; --emit needs exactly one.", operand, len(named))
		if len(alone) > 0 {
			b.WriteString(" Select one of:")
			for _, match := range alone {
				fmt.Fprintf(&b, "\n    %s", match.Ref)
			}
		}
		if len(shared) > 0 {
			b.WriteString("\nNo selection matches these alone; the query without --emit returns each one's span:")
			for _, match := range shared {
				fmt.Fprintf(&b, "\n    %s", match.Ref)
			}
		}
		b.WriteString("\nor drop --emit to list the matches")
		return invocationError("%s", b.String())
	}

	match := named[0]
	data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(match.Path)))
	if err != nil {
		return fmt.Errorf("reading %s: %w", match.Path, err)
	}
	if _, err := streams.out.Write(span.Slice(normalizeLineEndings(data), match.Section)); err != nil {
		return fmt.Errorf("writing the section of %q: %w", match.Text, err)
	}
	return nil
}

// selectsAlone reports whether match's own ref, given
// back as a selection, names exactly that heading — the question a caller
// pasting it back with --emit is asking, answered by the same selection
// the command runs.
func selectsAlone(index specmodel.SpecIndex, match headingMatch, inEffect func(specmodel.IndexEntry) bool) bool {
	matches, direct := selectHeadings(index, parseSelection(match.Ref), inEffect)
	var named []headingMatch
	for i, m := range matches {
		if direct[i] {
			named = append(named, m)
		}
	}
	return len(named) == 1 && named[0].Path == match.Path && named[0].Section == match.Section
}

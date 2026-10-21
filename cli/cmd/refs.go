package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/specs/cli/core/markdown"
	"github.com/brettcallaghan/specs/cli/core/references"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// newRefsCmd builds the `refs` subcommand: report which corpus members
// reference a given slug, or which documents that slug itself
// references, per AC-SPECREF-003.1 through AC-SPECREF-003.5 — and, for
// an argument qualified by an anchor, which references reach one heading
// of that document (AC-SPECANCH-003.1).
//
// This is the fifth subcommand FEAT-SPEC-011 anticipated growing into
// (spec-cli-container.md's Change guidance); it neither renames nor
// repurposes the four `newRootCmd` already builds.
func newRefsCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var jsonOut, outbound, full bool

	cmd := &cobra.Command{
		Use:   "refs <slug>[#<anchor>]",
		Short: "Report which corpus members reference a slug or one of its headings, or what it references",
		Long: "refs reports every corpus member whose body or registered frontmatter " +
			"properties reference the given slug. Qualified by an anchor, written " +
			"<slug>#<anchor>, it reports only the references whose anchor resolves to " +
			"that heading — the blast radius of renaming it — and a reference whose " +
			"anchor matches more than one heading is reported for each, marked as " +
			"matching more than one. A heading is matched on its identifier, so letter " +
			"case and punctuation do not decide a match. Given --outbound, it reports " +
			"the documents that slug references instead; --outbound takes no anchor. A " +
			"slug no corpus member declares is reported distinctly from one that is " +
			"declared but unreferenced; both exit zero, as does an anchor nothing " +
			"references. The answer is compact by default — one entry per referring " +
			"document, each reference reduced to its line and text; under --outbound, " +
			"one entry per cited slug with its resolved path and the lines citing it — " +
			"and --full returns every field of every reference, its span included.",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			asked := parseSelection(args[0])
			if outbound && asked.heading != "" {
				return invocationError("%w", errNoAnchorWithOutbound)
			}
			return runRefs(streams, *configPath, asked, outbound, full, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write the references as JSON instead of YAML")
	cmd.Flags().BoolVar(&outbound, "outbound", false,
		"report the documents the slug references, instead of the documents referencing it")
	cmd.Flags().BoolVar(&full, "full", false,
		"report every field of every reference, its span included, instead of the compact view")
	return cmd
}

// errNoAnchorWithOutbound rejects the one argument-and-flag combination
// that has no meaning. An anchor names a heading of the cited document,
// and --outbound asks what the given document cites — the anchors on
// those references belong to other documents' headings, so there is no
// question "the references out of X into the heading Y of X" answers.
var errNoAnchorWithOutbound = errors.New("--outbound takes a slug, not a slug qualified by an anchor")

// refsResult is what runRefs renders: a mapping of the properties the
// query established, per [[cli-output-standards]].
//
// Declared is what makes "no corpus member declares this slug" and "the
// slug is declared but has no references in the asked direction"
// distinguishable — two different facts about the corpus, both exiting
// zero (AC-SPECREF-003.5 and the Design section's undeclared-slug
// bullet) — rather than an empty References list meaning either one.
//
// Anchor is present exactly when the argument carried one, so an
// unanchored query answers with the properties it established and no key
// standing for a question nobody asked.
type refsResult struct {
	Slug      string `json:"slug"`
	Anchor    string `json:"anchor,omitempty"`
	Direction string `json:"direction"`
	Declared  bool   `json:"declared"`
	// AnchorDeclared states whether the asked anchor names a heading of
	// the declared document, by the rule an anchor reference resolves by
	// (AC-SPECANCH-003.6); false where it names none, and false where the
	// slug itself is undeclared. Present exactly when the argument
	// carries an anchor, as Anchor is.
	AnchorDeclared *bool       `json:"anchor_declared,omitempty"`
	References     []refsMatch `json:"references"`
}

// refsMatch is one reference the query answers with: the reference as the
// index recorded it, plus the one property this query establishes about
// it.
//
// The reference is embedded rather than copied field by field, so the
// answer is a view over the generated model ([[adr-016]]) and gains
// whatever a later schema change adds to a reference without this file
// being edited. Both renderings promote an embedded struct's fields the
// same way, so the two carry the same keys in the same order.
type refsMatch struct {
	specmodel.Reference

	// AmbiguousAnchor marks a reference whose anchor matched more than
	// one heading of its target (AC-SPECANCH-003.4). It is read off the
	// reference's own anchor resolution rather than decided a second
	// time here, and is written only where it is true: a reference
	// carrying no anchor, or one that matched exactly one heading, has
	// nothing to disambiguate.
	AmbiguousAnchor bool `json:"ambiguous_anchor,omitempty"`
}

// runRefs answers [corpus.Corpus.Slug] and, for a declared slug,
// [corpus.Corpus.InboundReferences] or [corpus.Corpus.OutboundReferences],
// narrows the result to one heading where the argument named one, and
// renders it. It is the whole of what this subcommand does: converting an
// argument and a direction into a corpus lookup and the result into
// stdout, with no domain logic of its own — what an anchor resolved to is
// the References component's answer, already on the reference.
//
// It always exits zero: like slug-exists, what refs answers is a fact
// about the corpus, never a failure of the command that asked it
// (AC-SPECREF-003.5, AC-SPECANCH-003.3, and the undeclared-slug case
// matching the slug-exists contract per AC-SPEC-007.2). The corpus is
// opened warm (corpus.Open), the same as check, slug-exists and next-id,
// because this is exactly the corpus query FEAT-SPEC-009.1's latency
// budget is measured against (AC-SPECREF-003.4).
func runRefs(streams *ioStreams, configPath string, asked selection, outbound, full, jsonOut bool) error {
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

	direction := "inbound"
	if outbound {
		direction = "outbound"
	}

	entry, declared := c.Slug(asked.document)
	// References defaults to an empty (never nil) slice, so both
	// renderings write an empty collection rather than a null for either
	// zero-result case — the same "collections are never nil" discipline
	// [finding.Report.Model] applies to Findings and Checks, and what
	// [[cli-output-standards]] requires of an empty answer.
	result := refsResult{
		Slug:       asked.document,
		Anchor:     asked.heading,
		Direction:  direction,
		Declared:   declared,
		References: []refsMatch{},
	}
	if asked.heading != "" {
		result.AnchorDeclared = anchorDeclared(declared, asked.heading, entry.Headings)
	}
	if declared {
		var found []specmodel.Reference
		if outbound {
			found = c.OutboundReferences(entry.Path)
		} else {
			found = c.InboundReferences(entry.Path)
		}
		result.References = matches(found, asked.heading, entry.Headings)
	}

	if full {
		return renderRefs(streams, result, jsonOut)
	}
	if outbound {
		compact, err := compactOutboundRefs(result)
		if err != nil {
			return err
		}
		return renderRefs(streams, compact, jsonOut)
	}
	compact, err := compactRefs(result)
	if err != nil {
		return err
	}
	return renderRefs(streams, compact, jsonOut)
}

// matches turns the references the corpus returned into the query's
// answer: all of them where no anchor was asked for, and otherwise only
// those whose anchor resolves to a heading of the queried document that
// the asked anchor selects — headings, that document's own.
//
// The asked anchor is a heading path query like any anchor
// (AC-SPECANCH-002.10), so it is matched against the document's headings
// by the same rule, and a reference is compared with what it selected by
// the position of the heading line rather than by text: two headings of
// one name are two headings.
//
// The order is the corpus's throughout, so two runs over an unchanged
// corpus print identical bytes.
func matches(found []specmodel.Reference, anchor string, headings []specmodel.Heading) []refsMatch {
	answer := []refsMatch{}
	selected := map[uint64]bool{}
	if anchor != "" {
		chosen := markdown.SelectPaths(anchor, references.HeadingPaths(headings))
		for i, heading := range headings {
			if chosen[i] {
				selected[heading.Span.Offset] = true
			}
		}
	}

	for _, reference := range found {
		ambiguous := ambiguousAnchor(reference)
		if anchor != "" && !reachesHeading(reference, selected) {
			continue
		}
		answer = append(answer, refsMatch{Reference: reference, AmbiguousAnchor: ambiguous})
	}
	return answer
}

// anchorDeclared reports whether anchor names some heading of the
// queried document — headings, its own — selected by the same heading
// path rule an anchor reference resolves by (AC-SPECANCH-002.2,
// AC-SPECANCH-002.10). An undeclared document has no headings to name.
func anchorDeclared(documentDeclared bool, anchor string, headings []specmodel.Heading) *bool {
	found := false
	if documentDeclared {
		for _, chosen := range markdown.SelectPaths(anchor, references.HeadingPaths(headings)) {
			found = found || chosen
		}
	}
	return &found
}

// reachesHeading reports whether reference's anchor resolved to one of
// the selected headings, each named by the offset of its heading line.
//
// The question is asked of the resolution rather than of the anchor's own
// text: the References component decided what this anchor reaches, and a
// second derivation here would be free to disagree with it. An ambiguous
// anchor reaches every heading it matched — it resolves to none of them,
// and the author renaming one of them is exactly the reader who has to
// see the reference (AC-SPECANCH-003.4).
func reachesHeading(reference specmodel.Reference, selected map[uint64]bool) bool {
	if reference.AnchorResolution == nil {
		return false
	}
	if resolved, ok := reference.AnchorResolution.AsResolved(); ok {
		return selected[resolved.Heading.Span.Offset]
	}
	if ambiguous, ok := reference.AnchorResolution.AsAmbiguous(); ok {
		for _, match := range ambiguous.Matches {
			if selected[match.Span.Offset] {
				return true
			}
		}
	}
	return false
}

// ambiguousAnchor reports whether reference's anchor matched more than
// one heading of its target.
func ambiguousAnchor(reference specmodel.Reference) bool {
	if reference.AnchorResolution == nil {
		return false
	}
	_, ambiguous := reference.AnchorResolution.AsAmbiguous()
	return ambiguous
}

// renderRefs writes result to streams.out as YAML, or as JSON under
// --json, both carrying the same values in the same order per
// [[cli-output-standards]]. Each reference carries its file path and span
// in either rendering under --full (AC-SPECREF-003.3, AC-SPECANCH-003.1);
// the compact view is rendered the same way, so --json changes the syntax
// of whichever view was chosen and never the view (AC-SPECREF-003.7).
func renderRefs(streams *ioStreams, result any, asJSON bool) error {
	if asJSON {
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("rendering the reference set as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}
	fmt.Fprint(streams.out, renderYAML(result))
	return nil
}

// refsCompactResult is the default answer: the query-level properties
// exactly as [refsResult] carries them, and the references grouped into
// one entry per referring document (AC-SPECREF-003.6), per
// [[cli-output-standards]]'s "refs: the compact view and --full".
type refsCompactResult struct {
	Slug           string         `json:"slug"`
	Anchor         string         `json:"anchor,omitempty"`
	Direction      string         `json:"direction"`
	Declared       bool           `json:"declared"`
	AnchorDeclared *bool          `json:"anchor_declared,omitempty"`
	References     []refsDocument `json:"references"`
}

// refsDocument is one referring document's entry in the compact view.
type refsDocument struct {
	Path       string                 `json:"path"`
	References []refsCompactReference `json:"references"`
}

// refsCompactReference keeps the fields of a reference the standard's
// compact-view table keeps, and drops the rest: the span is reduced to its
// start line, and slug, anchor, the target path and the anchor resolution
// are left out because the query already established them.
//
// Origin is written only for a reference in frontmatter, Resolution only
// for one whose slug did not resolve, and AmbiguousAnchor only where it is
// true (AC-SPECANCH-003.4).
type refsCompactReference struct {
	Line            uint64                     `json:"line"`
	Text            string                     `json:"text"`
	Origin          *specmodel.ReferenceOrigin `json:"origin,omitempty"`
	Resolution      *refsResolutionState       `json:"resolution,omitempty"`
	AmbiguousAnchor bool                       `json:"ambiguous_anchor,omitempty"`
}

// refsResolutionState is a slug resolution reduced to its state.
type refsResolutionState struct {
	State string `json:"state"`
}

// compactRefs derives the compact view from the full answer, so the two
// views cover exactly the same references by construction
// (AC-SPECREF-003.8). A document's entry is placed by its first reference
// and its references keep the corpus's order inside it, so two runs over
// an unchanged corpus print identical bytes.
func compactRefs(result refsResult) (refsCompactResult, error) {
	compact := refsCompactResult{
		Slug:           result.Slug,
		Anchor:         result.Anchor,
		Direction:      result.Direction,
		Declared:       result.Declared,
		AnchorDeclared: result.AnchorDeclared,
		References:     []refsDocument{},
	}
	position := map[string]int{}
	for _, match := range result.References {
		reference, err := compactReference(match)
		if err != nil {
			return refsCompactResult{}, err
		}
		at, seen := position[match.Path]
		if !seen {
			at = len(compact.References)
			position[match.Path] = at
			compact.References = append(compact.References,
				refsDocument{Path: match.Path, References: []refsCompactReference{}})
		}
		compact.References[at].References = append(compact.References[at].References, reference)
	}
	return compact, nil
}

// compactReference reduces one reference to the compact view's fields.
func compactReference(match refsMatch) (refsCompactReference, error) {
	reference := refsCompactReference{
		Line:            match.Span.StartLine,
		Text:            match.Text,
		AmbiguousAnchor: match.AmbiguousAnchor,
	}
	if _, body := match.Origin.AsBody(); !body {
		origin := match.Origin
		reference.Origin = &origin
	}
	if _, resolved := match.Resolution.AsResolved(); !resolved {
		state, err := resolutionState(match.Resolution)
		if err != nil {
			return refsCompactReference{}, err
		}
		reference.Resolution = &refsResolutionState{State: state}
	}
	return reference, nil
}

// resolutionState reads the state tag a resolution marshals under, so the
// compact view names a state exactly as the full view does rather than
// keeping a second list of the schema's state names.
func resolutionState(resolution specmodel.ReferenceResolution) (string, error) {
	data, err := json.Marshal(resolution)
	if err != nil {
		return "", fmt.Errorf("rendering a reference resolution: %w", err)
	}
	var tagged struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(data, &tagged); err != nil {
		return "", fmt.Errorf("reading a reference resolution's state: %w", err)
	}
	return tagged.State, nil
}

// refsOutboundCompactResult is the default answer to an outbound query:
// the query-level properties as [refsResult] carries them, and the
// references grouped into one entry per cited slug (AC-SPECREF-003.9),
// per [[cli-output-standards]]'s "refs --outbound".
type refsOutboundCompactResult struct {
	Slug           string               `json:"slug"`
	Anchor         string               `json:"anchor,omitempty"`
	Direction      string               `json:"direction"`
	Declared       bool                 `json:"declared"`
	AnchorDeclared *bool                `json:"anchor_declared,omitempty"`
	References     []refsCitedReference `json:"references"`
}

// refsCitedReference is one cited slug's entry. Path is a pointer
// without omitempty so a slug that resolves to nothing renders an
// explicit null; Resolution is written only for a slug that did not
// resolve.
type refsCitedReference struct {
	Slug       string                 `json:"slug"`
	Path       *string                `json:"path"`
	Resolution *refsResolutionState   `json:"resolution,omitempty"`
	Lines      []refsCompactReference `json:"lines"`
}

// compactOutboundRefs derives the outbound compact view from the same
// full answer [compactRefs] reads, so both views cover the same
// references (AC-SPECREF-003.8). An entry is placed by the first
// reference to its slug and its lines keep the corpus's order. Slugs
// resolve case-insensitively, so case variants of one slug share an
// entry, which carries the spelling of its first citation.
func compactOutboundRefs(result refsResult) (refsOutboundCompactResult, error) {
	compact := refsOutboundCompactResult{
		Slug:           result.Slug,
		Anchor:         result.Anchor,
		Direction:      result.Direction,
		Declared:       result.Declared,
		AnchorDeclared: result.AnchorDeclared,
		References:     []refsCitedReference{},
	}
	position := map[string]int{}
	for _, match := range result.References {
		line, err := compactReference(match)
		if err != nil {
			return refsOutboundCompactResult{}, err
		}
		key := strings.ToLower(match.Slug)
		at, seen := position[key]
		if !seen {
			at = len(compact.References)
			position[key] = at
			cited := refsCitedReference{Slug: match.Slug, Lines: []refsCompactReference{}}
			if resolved, ok := match.Resolution.AsResolved(); ok {
				path := resolved.TargetPath
				cited.Path = &path
			}
			cited.Resolution = line.Resolution
			compact.References = append(compact.References, cited)
		}
		// The entry carries the resolution once; a line repeats none.
		line.Resolution = nil
		compact.References[at].Lines = append(compact.References[at].Lines, line)
	}
	return compact, nil
}

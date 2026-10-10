// Package references finds the references one document makes.
//
// A reference is written in two places and this package reads both: as a
// `[[…]]` wikilink in the document's body, and as the value of a
// frontmatter property the active profile registers as reference-bearing.
// Each occurrence is recorded with the position a reader must open to fix
// it, and with the property it came from where it came from one.
//
// The two rules that matter most here are both about what is *not* a
// reference. A `[[…]]` inside backticks is an illustration of the
// notation, and a frontmatter value that is not in a registered property
// is prose about documents rather than a link to one. Both occur in this
// repository's own corpus, and a checker that reported either is a
// checker every reader learns to dismiss.
//
// One exception to the first rule: a fenced block whose info string names
// a registered Gherkin tag ([gherkin.FenceTags]) is Gherkin, not an
// illustration, so a `[[…]]` written in its feature description, a
// scenario name, a step, a rule name or an examples cell is extracted
// like any other body reference. A `.feature` file gets the same
// treatment for free — [Options.Source] is the whole file with
// [Options.Frontmatter] nil, so its body is Gherkin from the first byte
// and every wikilink in it is extracted and recorded against that file's
// own [Options.Path], even though the file is not a corpus member.
//
// Nothing here resolves anything. What a reference reached is a question
// about the corpus, answered where the corpus is known; this package
// leaves [specmodel.Reference.Resolution] unset, which is deliberately
// unserializable, so a reference cannot reach the index without having
// been resolved on the way.
//
// Nothing here reads a heading either. `goldmark` arrives with this
// package because locating a code region correctly needs a markdown
// parser — where a code span ends depends on matching backtick-run
// lengths, and where a fence ends depends on the fence length and its
// info string — but it is used for that one purpose. Positions are still
// computed by the Spans scanner, so every position in the corpus is
// computed one way.
package references

import (
	"strings"

	"github.com/brettcallaghan/specs/cli/core/frontmatter"
	"github.com/brettcallaghan/specs/cli/core/span"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// The delimiters a body reference is written between.
const (
	openingDelimiter = "[["
	closingDelimiter = "]]"
)

// anchorSeparator divides a reference's slug half from its anchor half.
// The first occurrence splits; a second is part of the anchor, which is
// recorded and not interpreted.
const anchorSeparator = "#"

// Options is one document as the extractor reads it.
type Options struct {
	// Path is the document's path, recorded on every reference found in
	// it so a reference is self-contained once extracted.
	Path string

	// Source is the whole document, frontmatter block included. Positions
	// are expressed against it, so a span from the body and a span from a
	// frontmatter property address the same coordinate system.
	Source []byte

	// Frontmatter is the document's parsed frontmatter, or nil for a
	// document that carries none — in which case the whole of Source is
	// body.
	//
	// The parsed block is taken rather than re-parsed: it is the only
	// form that carries a property's position, and reading a
	// reference-bearing value out of a typed frontmatter record instead
	// would lose both the position and every property the record's type
	// does not declare.
	Frontmatter *frontmatter.Document

	// ReferenceProperties are the frontmatter properties the active
	// profile registers as reference-bearing for this document's type.
	//
	// Property names rather than the profile itself, because what is
	// reference-bearing is a per-type answer the caller has already
	// looked up, and because this package must not learn to ask a profile
	// anything.
	ReferenceProperties []string
}

// Extract returns every reference the document makes, ordered by
// position: the registered frontmatter properties first, in the order the
// document declares them, then the body in the order it is written.
//
// Every occurrence is recorded separately. A document referencing the
// same target twice carries two references with two spans, because a
// finding about either has to cite the line the reader opens.
//
// A reference's position covers what a reader is pointed at and what a
// rewrite replaces: for a body reference the text between the brackets,
// the brackets themselves being notation rather than reference; for a
// frontmatter reference the value's own source text as written, quotes
// and all, because that is the extent that has to be replaced to change
// the value.
//
// Resolution is left unset on every reference returned. Resolving is a
// question about the corpus and belongs to the caller that knows one.
func Extract(options Options) []specmodel.Reference {
	var found []specmodel.Reference
	found = append(found, frontmatterReferences(options)...)
	return append(found, bodyReferences(options)...)
}

// frontmatterReferences returns one reference per value held by a
// registered reference-bearing property.
//
// A property holding a sequence contributes one reference per element, at
// that element's own span rather than at the whole list's: a reader told
// that a list holds a broken reference would otherwise still have to find
// which of its elements is the broken one.
//
// A value that is not text — a number, a mapping, a nested sequence —
// contributes nothing. Whether such a value belongs in a
// reference-bearing property at all is a question for the schema, which
// has the vocabulary to answer it; here it is simply not a slug.
func frontmatterReferences(options Options) []specmodel.Reference {
	if options.Frontmatter == nil || len(options.ReferenceProperties) == 0 {
		return nil
	}

	registered := make(map[string]bool, len(options.ReferenceProperties))
	for _, name := range options.ReferenceProperties {
		registered[name] = true
	}

	var found []specmodel.Reference
	for _, property := range options.Frontmatter.Properties {
		if !registered[property.Name] {
			continue
		}
		origin := specmodel.NewReferenceOriginFrontmatterProperty(
			specmodel.ReferenceFrontmatterOrigin{Property: property.Name})

		if property.Items == nil {
			if text, ok := property.Value.(string); ok {
				found = append(found, reference(options.Path, text, origin, property.Position))
			}
			continue
		}
		for _, item := range property.Items {
			if text, ok := item.Value.(string); ok {
				found = append(found, reference(options.Path, text, origin, item.Position))
			}
		}
	}
	return found
}

// bodyReferences returns one reference per `[[…]]` written in the
// document's body, in the order they are written.
//
// The body is everything after the frontmatter block, which is what keeps
// a frontmatter value describing the notation — `description: References
// use [[slug]] wikilinks` — from being read as a reference to a document
// called `slug`. Frontmatter contributes references only through
// registered properties.
func bodyReferences(options Options) []specmodel.Reference {
	start := bodyStart(options)
	if start >= len(options.Source) {
		return nil
	}
	body := options.Source[start:]
	code := codeRegions(body)

	var found []specmodel.Reference
	for from := 0; from < len(body); {
		region, ok := span.Find(body, openingDelimiter, closingDelimiter, from)
		if !ok {
			break
		}
		from = int(region.Outer.Offset + region.Outer.Length)

		if code.covers(int(region.Inner.Offset), int(region.Inner.Offset+region.Inner.Length)) {
			continue
		}
		found = append(found, reference(
			options.Path,
			string(span.Slice(body, region.Inner)),
			specmodel.NewReferenceOriginBody(specmodel.ReferenceBodyOrigin{}),
			span.Of(options.Source, start+int(region.Inner.Offset), int(region.Inner.Length)),
		))
	}
	return found
}

// bodyStart returns the offset the document's body begins at: the byte
// after the frontmatter block, or the start of the document when there is
// none.
func bodyStart(options Options) int {
	if options.Frontmatter == nil {
		return 0
	}
	outer := options.Frontmatter.Region.Outer
	return int(outer.Offset + outer.Length)
}

// reference builds one reference from the text written between the
// delimiters, or from the text of a frontmatter value.
//
// The text is recorded as written, and split on its first `#` into the
// slug half this feature resolves and the anchor half it records without
// interpreting. No other separator is read: the display alias
// `[[slug|text]]` is prohibited by the reference standard, so a reference
// written that way carries `slug|text` as its slug half and reports as
// unresolved, rather than resolving quietly to something the standard
// forbids.
func reference(path, text string, origin specmodel.ReferenceOrigin, at specmodel.Span) specmodel.Reference {
	slug, anchor, anchored := strings.Cut(text, anchorSeparator)

	found := specmodel.Reference{
		Path:   path,
		Text:   text,
		Slug:   slug,
		Origin: origin,
		Span:   at,
	}
	if anchored {
		found.Anchor = &anchor
	}
	return found
}

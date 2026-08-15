package gherkin

import (
	"bytes"
	"fmt"
	"strings"

	cucumber "github.com/cucumber/gherkin/go/v42"
)

// exampleKeywords are the keywords that open an `Examples` block in the
// dialect [Parse] reads by default, asked of the parser rather than
// written out for the same reason [featureKeywords] is: this package and
// the parser must never disagree about what the keyword is.
var exampleKeywords = cucumber.DialectsBuiltin().GetDialect(cucumber.DefaultDialect).ExamplesKeywords()

// RenderFeature is what [Render] needs to write one standalone `.feature`
// file.
//
// It carries what a caller with a tag vocabulary has already resolved:
// the tags a feature and each of its criteria should carry, in the order
// to write them. Render holds no vocabulary of its own — the same
// separation [Options] draws for parsing: turning an identifier or a
// qualifier into a tag's spelling is the caller's job, and this package
// only ever handles tags as strings.
type RenderFeature struct {
	// Tags are the feature's tags, written in this order on the line
	// above `Feature:`, each without a leading `@`.
	Tags []string

	// Name is the feature's name, following the `Feature` keyword.
	Name string

	// Description is the feature's description, written beneath the name
	// exactly as given. Omitted from the file where empty.
	Description string

	// Criteria are the feature's criteria, each written as its own
	// `Scenario` or `Scenario Outline` construct, in this order.
	Criteria []RenderCriterion
}

// RenderCriterion is one criterion, written as its own top-level
// construct.
//
// [specmodel.Criterion] is the atomic unit and carries no link back to a
// `Scenario Outline` it may have shared with other criteria at
// extraction, so Render recovers none: every RenderCriterion becomes its
// own `Scenario` or `Scenario Outline`, steps and all, even where the
// source divided one outline's `Examples` blocks into several. Text
// already stands alone per criterion, so this loses nothing a corpus tool
// reads and keeps every criterion independently addressable in the
// emitted file, exactly as it already is in the index.
type RenderCriterion struct {
	// Tags are the criterion's tags, written in this order, each without
	// a leading `@` — its identifier tag together with whatever qualifier
	// tags apply.
	Tags []string

	// Name is the criterion's name, following the `Scenario` or
	// `Scenario Outline` keyword. Omitted where empty: the construct
	// still parses, with no name.
	Name string

	// Description is the criterion's description, written beneath its
	// name. Omitted where empty.
	Description string

	// Rule is the name of the business rule this criterion sits beneath.
	// Where set, the criterion is wrapped in its own `Rule:` construct
	// carrying that name; where empty, the criterion sits directly
	// beneath the `Feature`.
	Rule string

	// Text is the criterion's body exactly as a [specmodel.Criterion]'s
	// own Text carries it: its steps and any `Examples` block, dedented,
	// with no keyword line of its own. Render supplies the `Scenario` or
	// `Scenario Outline` keyword, choosing `Scenario Outline` wherever
	// Text contains a line opening an `Examples` block.
	Text string
}

// Render writes f as the bytes of a standalone `.feature` file: valid
// Gherkin that [Parse] reads back under the same tag vocabulary that
// resolved f's tags.
//
// Render is Parse's inverse for what a caller can express through
// [RenderFeature]: parsing the bytes it returns yields a [Feature] whose
// Tags, and whose Criteria's Tags, Record.Name, Record.Text,
// Record.Description, Record.Rule and Record.ContentHash, equal the
// values f was built from. Record.Position differs, being freshly
// computed over the emitted bytes rather than carried over, and a
// caller's own source-form and owning-document bookkeeping (outside what
// RenderFeature carries) is theirs to redo — a rendered file is always a
// standalone feature source, never a fenced block.
func Render(f RenderFeature) []byte {
	var b bytes.Buffer
	writeTagLine(&b, "", f.Tags)
	fmt.Fprintf(&b, "Feature: %s\n", f.Name)
	if f.Description != "" {
		b.WriteByte('\n')
		writeIndented(&b, "  ", f.Description)
	}
	for _, criterion := range f.Criteria {
		b.WriteByte('\n')
		writeCriterion(&b, criterion)
	}
	return b.Bytes()
}

// writeCriterion writes one criterion as its own `Scenario` or
// `Scenario Outline` construct, wrapped in a `Rule:` where c.Rule is set.
func writeCriterion(b *bytes.Buffer, c RenderCriterion) {
	indent := "  "
	if c.Rule != "" {
		fmt.Fprintf(b, "  Rule: %s\n\n", c.Rule)
		indent = "    "
	}

	keyword := "Scenario"
	if hasExamplesBlock(c.Text) {
		keyword = "Scenario Outline"
	}

	writeTagLine(b, indent, c.Tags)
	if c.Name != "" {
		fmt.Fprintf(b, "%s%s: %s\n", indent, keyword, c.Name)
	} else {
		fmt.Fprintf(b, "%s%s:\n", indent, keyword)
	}

	body := indent + "  "
	if c.Description != "" {
		b.WriteByte('\n')
		writeIndented(b, body, c.Description)
		b.WriteByte('\n')
	}
	if c.Text != "" {
		writeIndented(b, body, c.Text)
	}
}

// writeTagLine writes tags as one line of `@`-prefixed tags at indent,
// and writes nothing at all where tags is empty — an untagged construct
// carries no blank tag line above it.
func writeTagLine(b *bytes.Buffer, indent string, tags []string) {
	if len(tags) == 0 {
		return
	}
	marked := make([]string, len(tags))
	for i, tag := range tags {
		marked[i] = "@" + tag
	}
	fmt.Fprintf(b, "%s%s\n", indent, strings.Join(marked, " "))
}

// writeIndented writes text to b, one source line per output line, each
// non-blank line prefixed with indent and each blank line left bare.
//
// A blank line stays bare rather than gaining trailing whitespace so
// that, read back by [Parse] and dedented again, every line — blank or
// not — returns to exactly the text it started from.
func writeIndented(b *bytes.Buffer, indent, text string) {
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			b.WriteByte('\n')
			continue
		}
		b.WriteString(indent)
		b.WriteString(line)
		b.WriteByte('\n')
	}
}

// hasExamplesBlock reports whether text carries a line opening an
// `Examples` block, which is what decides between `Scenario` and
// `Scenario Outline` for the construct wrapping it.
func hasExamplesBlock(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		for _, keyword := range exampleKeywords {
			if strings.HasPrefix(trimmed, keyword+":") {
				return true
			}
		}
	}
	return false
}

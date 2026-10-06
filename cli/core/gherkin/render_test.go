// Tests for [gherkin.Render] as [gherkin.Parse]'s inverse.
//
// The vocabulary these tests hand to [gherkin.Options] is [vocabulary],
// the same test double gherkin_test.go uses — Render carries no
// vocabulary of its own, so proving it round-trips needs no more of one
// than parsing already does.
package gherkin_test

import (
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/gherkin"
)

// toRenderFeature converts a parsed [gherkin.Feature] into the
// [gherkin.RenderFeature] a caller with this test's vocabulary would
// build from it — the same conversion cli/cmd performs from a
// [specmodel.Feature], done here directly from the parse-time shape so
// the round trip below exercises Render without depending on corpus.
func toRenderFeature(f *gherkin.Feature) gherkin.RenderFeature {
	var description string
	if f.Record.Description != nil {
		description = *f.Record.Description
	}

	criteria := make([]gherkin.RenderCriterion, len(f.Criteria))
	for i, c := range f.Criteria {
		var name, desc, rule string
		if c.Record.Name != nil {
			name = *c.Record.Name
		}
		if c.Record.Description != nil {
			desc = *c.Record.Description
		}
		if c.Record.Rule != nil {
			rule = *c.Record.Rule
		}
		criteria[i] = gherkin.RenderCriterion{
			Tags:        c.Tags,
			Name:        name,
			Description: desc,
			Rule:        rule,
			Text:        c.Record.Text,
		}
	}

	return gherkin.RenderFeature{
		Tags:        f.Tags,
		Name:        f.Record.Name,
		Description: description,
		Criteria:    criteria,
	}
}

// AC: AC-GHK-009.2
//
// Parsing an emitted file yields records equal to those extracted from
// the source, but for their source location: a feature carrying a
// description, a criterion grouped under a Rule with its own
// description and a doc string, a plain criterion, and an outline whose
// two Examples blocks are separately identified — the full range of
// shapes [gherkin.Parse] produces.
func TestRenderIsParsesInverse(t *testing.T) {
	src := "@FEAT-CHK-001 @frd-checkout\n" +
		"Feature: Checkout\n" +
		"  As a shopper\n" +
		"  I want to pay\n\n" +
		"  So that my order ships\n\n" +
		"  Rule: a declined card returns the basket\n\n" +
		"    @AC-CHK-001.1 @negative\n" +
		"    Scenario: the card declines\n" +
		"      The bank is the authority here.\n" +
		"      Given a basket\n" +
		"      When the card declines\n\n" +
		"  @AC-CHK-001.2\n" +
		"  Scenario: pays with a card\n" +
		"    Given a basket\n" +
		"      \"\"\"\n" +
		"      # not a comment\n" +
		"      \"\"\"\n" +
		"    When the card is accepted\n\n" +
		"  Scenario Outline: pays with <method>\n" +
		"    Given a basket\n" +
		"    When paying with <method>\n\n" +
		"    @AC-CHK-001.3\n" +
		"    Examples: accepted\n" +
		"      | method |\n" +
		"      | card   |\n\n" +
		"    @AC-CHK-001.4\n" +
		"    Examples: declined\n" +
		"      | method |\n" +
		"      | cash   |\n"

	original := parse(t, src)

	rendered := gherkin.Render(toRenderFeature(original))
	roundTripped := parse(t, string(rendered))

	if got, want := roundTripped.Record.Name, original.Record.Name; got != want {
		t.Errorf("Name = %q, want %q\nrendered:\n%s", got, want, rendered)
	}
	if got, want := derefOrEmpty(roundTripped.Record.Description), derefOrEmpty(original.Record.Description); got != want {
		t.Errorf("Description = %q, want %q\nrendered:\n%s", got, want, rendered)
	}
	if got, want := roundTripped.Record.Id, original.Record.Id; got != want {
		t.Errorf("Id = %q, want %q\nrendered:\n%s", got, want, rendered)
	}
	if got, want := roundTripped.Record.Code, original.Record.Code; got != want {
		t.Errorf("Code = %q, want %q\nrendered:\n%s", got, want, rendered)
	}
	if got, want := joined(roundTripped.Tags), joined(original.Tags); got != want {
		t.Errorf("Tags = %q, want %q\nrendered:\n%s", got, want, rendered)
	}

	if got, want := len(roundTripped.Criteria), len(original.Criteria); got != want {
		t.Fatalf("criteria count = %d, want %d\nrendered:\n%s", got, want, rendered)
	}
	for i := range original.Criteria {
		wantC, gotC := original.Criteria[i].Record, roundTripped.Criteria[i].Record
		if gotC.Id != wantC.Id {
			t.Errorf("criterion %d: Id = %q, want %q\nrendered:\n%s", i, gotC.Id, wantC.Id, rendered)
		}
		if derefOrEmpty(gotC.Name) != derefOrEmpty(wantC.Name) {
			t.Errorf("criterion %d: Name = %q, want %q\nrendered:\n%s", i, derefOrEmpty(gotC.Name), derefOrEmpty(wantC.Name), rendered)
		}
		if gotC.Text != wantC.Text {
			t.Errorf("criterion %d: Text = %q, want %q\nrendered:\n%s", i, gotC.Text, wantC.Text, rendered)
		}
		if derefOrEmpty(gotC.Description) != derefOrEmpty(wantC.Description) {
			t.Errorf("criterion %d: Description = %q, want %q\nrendered:\n%s", i, derefOrEmpty(gotC.Description), derefOrEmpty(wantC.Description), rendered)
		}
		if derefOrEmpty(gotC.Rule) != derefOrEmpty(wantC.Rule) {
			t.Errorf("criterion %d: Rule = %q, want %q\nrendered:\n%s", i, derefOrEmpty(gotC.Rule), derefOrEmpty(wantC.Rule), rendered)
		}
		if gotC.Notation != wantC.Notation {
			t.Errorf("criterion %d: Notation = %q, want %q\nrendered:\n%s", i, gotC.Notation, wantC.Notation, rendered)
		}
		if gotC.ContentHash != wantC.ContentHash {
			t.Errorf("criterion %d: ContentHash = %q, want %q\nrendered:\n%s", i, gotC.ContentHash, wantC.ContentHash, rendered)
		}
		if joined(roundTripped.Criteria[i].Tags) != joined(original.Criteria[i].Tags) {
			t.Errorf("criterion %d: Tags = %q, want %q\nrendered:\n%s",
				i, joined(roundTripped.Criteria[i].Tags), joined(original.Criteria[i].Tags), rendered)
		}
		if gotC.Span == wantC.Span {
			t.Errorf("criterion %d: Position equals the source's own, want a fresh one over the emitted bytes", i)
		}
	}
}

// AC: AC-GHK-009.1
func TestRenderWritesTheOwningDocumentTagAndTheFeatureIdentifierAboveFeature(t *testing.T) {
	rendered := gherkin.Render(gherkin.RenderFeature{
		Tags: []string{"frd-checkout", "FEAT-CHK-001"},
		Name: "Checkout",
	})

	want := "@frd-checkout @FEAT-CHK-001\nFeature: Checkout\n"
	if got := string(rendered); got != want {
		t.Errorf("Render = %q, want %q", got, want)
	}
}

// AC: AC-GHK-009.1
func TestRenderChoosesScenarioOutlineOnlyWhereTextCarriesAnExamplesBlock(t *testing.T) {
	plain := gherkin.Render(gherkin.RenderFeature{
		Name: "Checkout",
		Criteria: []gherkin.RenderCriterion{
			{Tags: []string{"AC-CHK-001.1"}, Name: "pays", Text: "Given a basket"},
		},
	})
	if !strings.Contains(string(plain), "\n  Scenario: pays\n") {
		t.Errorf("Render = %q, want a bare Scenario for a criterion with no Examples block", plain)
	}

	outline := gherkin.Render(gherkin.RenderFeature{
		Name: "Checkout",
		Criteria: []gherkin.RenderCriterion{{
			Tags: []string{"AC-CHK-001.1"},
			Name: "pays with <method>",
			Text: "Given a basket\nExamples: accepted\n  | method |\n  | card   |",
		}},
	})
	if !strings.Contains(string(outline), "\n  Scenario Outline: pays with <method>\n") {
		t.Errorf("Render = %q, want Scenario Outline for a criterion whose Text carries an Examples block", outline)
	}
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// joined renders a tag slice as one comparable string, so a mismatch
// prints both slices' whole content rather than a length or a pointer.
func joined(ss []string) string {
	return strings.Join(ss, ",")
}

// Tests for Gherkin extraction.
//
// The vocabulary these tests hand to [gherkin.Options] is the tests' own:
// the package under test attaches no meaning to any tag, so a test that
// wants an identifier resolved must supply the rule that resolves it. The
// `AC-` and `FEAT-` literals below live here for that reason and belong to
// no production source in this module.
package gherkin_test

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/gherkin"
	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

var (
	featureTag   = regexp.MustCompile(`^FEAT-([A-Z0-9]+)-[0-9]{3}$`)
	criterionTag = regexp.MustCompile(`^AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`)
)

// vocabulary is a stand-in for what a profile registers: the two rules
// that turn a tag into an identity.
func vocabulary() gherkin.Options {
	return gherkin.Options{
		FeatureIdentity: func(tags []string) (specmodel.FeatureId, specmodel.FeatureCode, bool) {
			for _, tag := range tags {
				if match := featureTag.FindStringSubmatch(tag); match != nil {
					return specmodel.FeatureId(tag), specmodel.FeatureCode(match[1]), true
				}
			}
			return "", "", false
		},
		CriterionIdentity: func(tags []string) (specmodel.CriterionId, bool) {
			for _, tag := range tags {
				if criterionTag.MatchString(tag) {
					return specmodel.CriterionId(tag), true
				}
			}
			return "", false
		},
	}
}

// fenced parses the one ```gherkin block of doc, the way a document walk
// reaches a feature written inside a specification.
func fenced(t *testing.T, doc string) *gherkin.Feature {
	t.Helper()

	blocks := markdown.Fenced([]byte(doc), []string{"gherkin"})
	if len(blocks) != 1 {
		t.Fatalf("located %d gherkin blocks, want 1", len(blocks))
	}
	feature, err := gherkin.ParseRegion([]byte(doc), blocks[0].Inner, vocabulary())
	if err != nil {
		t.Fatalf("ParseRegion: %v", err)
	}
	if feature == nil {
		t.Fatal("ParseRegion returned no feature")
	}
	return feature
}

func parse(t *testing.T, src string) *gherkin.Feature {
	t.Helper()

	feature, err := gherkin.Parse([]byte(src), vocabulary())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if feature == nil {
		t.Fatal("Parse returned no feature")
	}
	return feature
}

func only(t *testing.T, feature *gherkin.Feature) gherkin.Criterion {
	t.Helper()

	if len(feature.Criteria) != 1 {
		t.Fatalf("feature carries %d criteria, want 1", len(feature.Criteria))
	}
	return feature.Criteria[0]
}

// The line arithmetic. A source that declares no Feature is parsed behind
// a synthetic header, and every line the parser reports must come back as
// the line in the document the source was read from: the source's start
// line, plus the parser's line, less the lines the header added.

// AC: AC-GHK-001.4
func TestASourceBeginningWithAScenarioIsParsedBehindASyntheticHeader(t *testing.T) {
	doc := "# Checkout\n" + // line 1
		"\n" + // line 2
		"```gherkin\n" + // line 3
		"@AC-CHK-001.1\n" + // line 4
		"Scenario: pays with a card\n" + // line 5
		"  Given a basket\n" + // line 6
		"  When the card is accepted\n" + // line 7
		"```\n" // line 8

	feature := fenced(t, doc)

	criterion := only(t, feature)
	if got, want := criterion.Record.Id, specmodel.CriterionId("AC-CHK-001.1"); got != want {
		t.Errorf("Id = %q, want %q", got, want)
	}
	if got, want := criterion.Record.Position.StartLine, uint64(4); got != want {
		t.Errorf("criterion starts on line %d, want %d", got, want)
	}
	if got, want := criterion.Record.Position.EndLine, uint64(7); got != want {
		t.Errorf("criterion ends on line %d, want %d", got, want)
	}
	want := "@AC-CHK-001.1\nScenario: pays with a card\n  Given a basket\n  When the card is accepted"
	if got := string(span.Slice([]byte(doc), criterion.Record.Position)); got != want {
		t.Errorf("criterion position slices to %q, want %q", got, want)
	}
}

// AC: AC-GHK-001.4
func TestEveryReportedLineIsTheLineInTheContainingDocument(t *testing.T) {
	// The block opens on line 12, so its content opens on line 13 and the
	// parser's line 2 — line 1 being the synthetic header — is line 13.
	doc := strings.Repeat("Prose paragraph.\n\n", 5) + // lines 1-10
		"```gherkin\n" + // line 11
		"Rule: a card that declines returns the basket\n" + // line 12
		"\n" + // line 13
		"  @AC-CHK-002.1\n" + // line 14
		"  Scenario: the card declines\n" + // line 15
		"    Given a basket\n" + // line 16
		"```\n" // line 17

	feature := fenced(t, doc)

	if got, want := feature.Record.Position.StartLine, uint64(12); got != want {
		t.Errorf("feature starts on line %d, want %d", got, want)
	}
	criterion := only(t, feature)
	if got, want := criterion.Record.Position.StartLine, uint64(14); got != want {
		t.Errorf("criterion starts on line %d, want %d", got, want)
	}
	if got, want := criterion.Record.Position.EndLine, uint64(16); got != want {
		t.Errorf("criterion ends on line %d, want %d", got, want)
	}
	if got, want := *criterion.Record.Rule, "a card that declines returns the basket"; got != want {
		t.Errorf("Rule = %q, want %q", got, want)
	}
}

// AC: AC-GHK-001.5
func TestNoReportedPositionFallsInsideTheSyntheticHeader(t *testing.T) {
	doc := "```gherkin\n" + // line 1
		"Scenario: pays with a card\n" + // line 2
		"  Given a basket\n" + // line 3
		"```\n" // line 4

	blocks := markdown.Fenced([]byte(doc), []string{"gherkin"})
	feature := fenced(t, doc)

	content := blocks[0].Inner
	positions := map[string]span.Span{"feature": feature.Record.Position}
	for _, criterion := range feature.Criteria {
		positions["criterion"] = criterion.Record.Position
	}
	if len(positions) != 2 {
		t.Fatalf("positions = %v, want one feature and one criterion", positions)
	}
	for what, position := range positions {
		if position.StartLine < content.StartLine {
			t.Errorf("%s starts on line %d, before the source's own first line %d",
				what, position.StartLine, content.StartLine)
		}
		if position.Offset < content.Offset {
			t.Errorf("%s starts at offset %d, before the source's own first byte %d",
				what, position.Offset, content.Offset)
		}
		if position.Offset+position.Length > content.Offset+content.Length {
			t.Errorf("%s ends at offset %d, past the source's own last byte %d",
				what, position.Offset+position.Length, content.Offset+content.Length)
		}
		if got := string(span.Slice([]byte(doc), position)); !strings.Contains(got, "Scenario: pays with a card") {
			t.Errorf("%s slices to %q, which is not the text it was read from", what, got)
		}
	}
}

// AC: AC-GHK-001.5
func TestAFeatureWrittenWithNoHeaderCarriesNoSyntheticName(t *testing.T) {
	feature := parse(t, "Scenario: pays with a card\n  Given a basket\n")

	if feature.Record.Name != "" {
		t.Errorf("Name = %q, want the empty name of a source that declared none", feature.Record.Name)
	}
	if len(feature.Tags) != 0 {
		t.Errorf("Tags = %v, want none", feature.Tags)
	}
	if got, want := feature.Record.Position.StartLine, uint64(1); got != want {
		t.Errorf("feature starts on line %d, want %d", got, want)
	}
}

// AC: AC-GHK-001.4
func TestASourceThatDeclaresAFeatureIsParsedWithoutAHeader(t *testing.T) {
	feature := parse(t, "@FEAT-CHK-001\nFeature: Checkout\n\n  Scenario: pays\n    Given a basket\n")

	if got, want := feature.Record.Name, "Checkout"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	criterion := only(t, feature)
	if got, want := criterion.Record.Position.StartLine, uint64(4); got != want {
		t.Errorf("criterion starts on line %d, want %d", got, want)
	}
}

// The content hash. It covers a criterion's name and text and nothing
// else, over text normalized so that moving, reindenting or checking out
// with CRLF line endings leaves it unchanged.

// AC: AC-GHK-005.5
func TestACriterionCarriesAHashOverItsNameAndTextAlone(t *testing.T) {
	criterion := only(t, parse(t, "Feature: Checkout\n\n  @AC-CHK-001.1\n  Scenario: pays\n    Given a basket\n"))

	want := gherkin.ContentHashOf(*criterion.Record.Name, criterion.Record.Text)
	if criterion.Record.ContentHash != want {
		t.Errorf("ContentHash = %q, want %q", criterion.Record.ContentHash, want)
	}
	if !strings.HasPrefix(string(criterion.Record.ContentHash), "sha256:") {
		t.Errorf("ContentHash = %q, want a sha256: prefix", criterion.Record.ContentHash)
	}
	if strings.ToLower(string(criterion.Record.ContentHash)) != string(criterion.Record.ContentHash) {
		t.Errorf("ContentHash = %q, want lowercase hexadecimal", criterion.Record.ContentHash)
	}
}

// AC: AC-GHK-005.6
func TestACriterionThatOnlyMovedHashesIdentically(t *testing.T) {
	before := "Feature: Checkout\n\n" +
		"  @AC-CHK-001.1\n  Scenario: pays\n    Given a basket\n"
	after := "Feature: Checkout\n\n" +
		"  @AC-CHK-001.2\n  Scenario: browses\n    Given a catalogue\n\n" +
		"  @AC-CHK-001.1\n  Scenario: pays\n    Given a basket\n"

	moved := only(t, parse(t, before))
	inserted := parse(t, after)
	if len(inserted.Criteria) != 2 {
		t.Fatalf("feature carries %d criteria, want 2", len(inserted.Criteria))
	}
	same := inserted.Criteria[1]

	if same.Record.ContentHash != moved.Record.ContentHash {
		t.Errorf("hash changed when the criterion only moved: %q then %q",
			moved.Record.ContentHash, same.Record.ContentHash)
	}
	if same.Record.Position.StartLine == moved.Record.Position.StartLine {
		t.Fatal("the criterion did not move; the test proves nothing")
	}
}

// AC: AC-GHK-005.6
func TestReindentingOrCheckingOutWithCRLFLeavesTheHashUnchanged(t *testing.T) {
	plain := only(t, parse(t, "Feature: Checkout\n\n  @AC-CHK-001.1\n  Scenario: pays\n    Given a basket\n    When the card is accepted\n"))

	reindented := only(t, fenced(t, "```gherkin\nFeature: Checkout\n\n        @AC-CHK-001.1\n        Scenario: pays\n          Given a basket\n          When the card is accepted\n```\n"))
	if reindented.Record.ContentHash != plain.Record.ContentHash {
		t.Errorf("reindenting changed the hash: %q then %q",
			plain.Record.ContentHash, reindented.Record.ContentHash)
	}

	crlf := only(t, parse(t, "Feature: Checkout\r\n\r\n  @AC-CHK-001.1\r\n  Scenario: pays\r\n    Given a basket\r\n    When the card is accepted\r\n"))
	if crlf.Record.ContentHash != plain.Record.ContentHash {
		t.Errorf("a CRLF checkout changed the hash: %q then %q",
			plain.Record.ContentHash, crlf.Record.ContentHash)
	}
}

// AC: AC-GHK-005.7
func TestEditingACriterionChangesItsHash(t *testing.T) {
	before := only(t, parse(t, "Feature: Checkout\n\n  @AC-CHK-001.1\n  Scenario: pays\n    Given a basket\n"))
	text := only(t, parse(t, "Feature: Checkout\n\n  @AC-CHK-001.1\n  Scenario: pays\n    Given an empty basket\n"))
	name := only(t, parse(t, "Feature: Checkout\n\n  @AC-CHK-001.1\n  Scenario: pays by card\n    Given a basket\n"))

	if text.Record.ContentHash == before.Record.ContentHash {
		t.Error("editing a criterion's text left its hash unchanged")
	}
	if name.Record.ContentHash == before.Record.ContentHash {
		t.Error("editing a criterion's name left its hash unchanged")
	}
}

// AC: AC-GHK-005.6
func TestTheHashOfANamelessCriterionIsNotTheHashOfItsText(t *testing.T) {
	// Name and text are separated in the hash, so a name that ends where a
	// text begins cannot collide with the same characters divided the other
	// way.
	if gherkin.ContentHashOf("a", "b") == gherkin.ContentHashOf("ab", "") {
		t.Error("the hash runs a criterion's name into its text")
	}
}

// Outlines and their Examples blocks.

// AC: AC-GHK-002.6
func TestOneCriterionIsEmittedPerTaggedExamplesBlock(t *testing.T) {
	feature := parse(t, "Feature: Checkout\n\n"+
		"  Scenario Outline: pays with <method>\n"+
		"    Given a basket\n"+
		"    When paying with <method>\n\n"+
		"    @AC-CHK-001.1\n"+
		"    Examples: accepted\n"+
		"      | method |\n"+
		"      | card   |\n\n"+
		"    @AC-CHK-001.2\n"+
		"    Examples: declined\n"+
		"      | method |\n"+
		"      | cash   |\n")

	if len(feature.Criteria) != 2 {
		t.Fatalf("feature carries %d criteria, want one per tagged Examples block", len(feature.Criteria))
	}
	first, second := feature.Criteria[0], feature.Criteria[1]
	if got, want := first.Record.Id, specmodel.CriterionId("AC-CHK-001.1"); got != want {
		t.Errorf("first Id = %q, want %q", got, want)
	}
	if got, want := second.Record.Id, specmodel.CriterionId("AC-CHK-001.2"); got != want {
		t.Errorf("second Id = %q, want %q", got, want)
	}
	if !strings.Contains(first.Record.Text, "| card   |") {
		t.Errorf("first text = %q, want its own example rows", first.Record.Text)
	}
	if strings.Contains(first.Record.Text, "| cash   |") {
		t.Errorf("first text = %q, want nothing from the other block", first.Record.Text)
	}
	if !strings.Contains(first.Record.Text, "When paying with <method>") {
		t.Errorf("first text = %q, want the outline's steps", first.Record.Text)
	}
	if first.Record.ContentHash == second.Record.ContentHash {
		t.Error("two criteria differing only in their example rows share a hash")
	}
}

// AC: AC-GHK-002.9
func TestAnOutlineWhoseExamplesBlocksCarryNoIdentifierIsOneCriterion(t *testing.T) {
	feature := parse(t, "Feature: Checkout\n\n"+
		"  @AC-CHK-001.1\n"+
		"  Scenario Outline: pays with <method>\n"+
		"    Given a basket\n\n"+
		"    Examples: accepted\n"+
		"      | method |\n"+
		"      | card   |\n\n"+
		"    Examples: declined\n"+
		"      | method |\n"+
		"      | cash   |\n")

	criterion := only(t, feature)
	if got, want := criterion.Record.Id, specmodel.CriterionId("AC-CHK-001.1"); got != want {
		t.Errorf("Id = %q, want %q", got, want)
	}
	for _, row := range []string{"| card   |", "| cash   |", "Examples: accepted", "Examples: declined"} {
		if !strings.Contains(criterion.Record.Text, row) {
			t.Errorf("text = %q, want it to cover %q", criterion.Record.Text, row)
		}
	}
	if got, want := criterion.Record.Position.EndLine, uint64(13); got != want {
		t.Errorf("criterion ends on line %d, want %d — the last row of the last block", got, want)
	}
}

// AC: AC-GHK-002.6
func TestAnUntaggedExamplesBlockBesideATaggedOneIsItsOwnCriterion(t *testing.T) {
	// A partially tagged outline is a defect a later check reports. The
	// records must carry the untagged block for it to report, so the block
	// becomes a criterion with no identifier rather than vanishing.
	feature := parse(t, "Feature: Checkout\n\n"+
		"  Scenario Outline: pays with <method>\n"+
		"    Given a basket\n\n"+
		"    @AC-CHK-001.1\n"+
		"    Examples: accepted\n"+
		"      | method |\n"+
		"      | card   |\n\n"+
		"    Examples: declined\n"+
		"      | method |\n"+
		"      | cash   |\n")

	if len(feature.Criteria) != 2 {
		t.Fatalf("feature carries %d criteria, want one per Examples block", len(feature.Criteria))
	}
	if got := feature.Criteria[1].Record.Id; got != "" {
		t.Errorf("second Id = %q, want none — the block carries no identifier", got)
	}
}

// AC: AC-GHK-002.6
func TestACriterionFromAnExamplesBlockCarriesItsOutlinesTagsBeforeItsOwn(t *testing.T) {
	feature := parse(t, "Feature: Checkout\n\n"+
		"  @manual\n"+
		"  Scenario Outline: pays with <method>\n"+
		"    Given a basket\n\n"+
		"    @AC-CHK-001.1\n"+
		"    Examples: accepted\n"+
		"      | method |\n"+
		"      | card   |\n")

	criterion := only(t, feature)
	want := []string{"manual", "AC-CHK-001.1"}
	if len(criterion.Tags) != len(want) {
		t.Fatalf("Tags = %v, want %v", criterion.Tags, want)
	}
	for i := range want {
		if criterion.Tags[i] != want[i] {
			t.Errorf("Tags = %v, want %v", criterion.Tags, want)
		}
	}
}

// What the records carry.

// AC: AC-GHK-003.2
func TestAFeatureRecordCarriesItsIdentifierAndItsCode(t *testing.T) {
	feature := parse(t, "@FEAT-CHK-001\nFeature: Checkout\n\n  Scenario: pays\n    Given a basket\n")

	if got, want := feature.Record.Id, specmodel.FeatureId("FEAT-CHK-001"); got != want {
		t.Errorf("Id = %q, want %q", got, want)
	}
	if got, want := feature.Record.Code, specmodel.FeatureCode("CHK"); got != want {
		t.Errorf("Code = %q, want %q", got, want)
	}
}

// AC: AC-GHK-005.1
func TestAFeatureRecordCarriesItsNameDescriptionTagsAndSpan(t *testing.T) {
	src := "@FEAT-CHK-001 @frd-checkout\n" +
		"Feature: Checkout\n" +
		"  As a shopper\n" +
		"  I want to pay\n\n" +
		"  Scenario: pays\n" +
		"    Given a basket\n"

	feature := parse(t, src)

	if got, want := feature.Record.Name, "Checkout"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if feature.Record.Description == nil {
		t.Fatal("Description is absent, want the user story")
	}
	if got, want := *feature.Record.Description, "As a shopper\nI want to pay"; got != want {
		t.Errorf("Description = %q, want %q", got, want)
	}
	if got, want := strings.Join(feature.Tags, ","), "FEAT-CHK-001,frd-checkout"; got != want {
		t.Errorf("Tags = %q, want %q", got, want)
	}
	if got, want := string(span.Slice([]byte(src), feature.Record.Position)), src; got != want {
		t.Errorf("Position slices to %q, want the whole source", got)
	}
}

// AC: AC-GHK-005.2
func TestACriterionRecordCarriesItsIdentityTextAndProvenance(t *testing.T) {
	src := "Feature: Checkout\n\n" +
		"  Rule: a declined card returns the basket\n\n" +
		"    @AC-CHK-001.1 @negative\n" +
		"    Scenario: the card declines\n" +
		"      The bank is the authority here.\n" +
		"      Given a basket\n" +
		"      When the card declines\n"

	criterion := only(t, parse(t, src))

	if got, want := criterion.Record.Id, specmodel.CriterionId("AC-CHK-001.1"); got != want {
		t.Errorf("Id = %q, want %q", got, want)
	}
	if criterion.Record.Name == nil || *criterion.Record.Name != "the card declines" {
		t.Errorf("Name = %v, want %q", criterion.Record.Name, "the card declines")
	}
	if got, want := criterion.Record.Text, "Given a basket\nWhen the card declines"; got != want {
		t.Errorf("Text = %q, want %q", got, want)
	}
	if criterion.Record.Description == nil || *criterion.Record.Description != "The bank is the authority here." {
		t.Errorf("Description = %v, want the scenario's own prose", criterion.Record.Description)
	}
	if criterion.Record.Rule == nil || *criterion.Record.Rule != "a declined card returns the basket" {
		t.Errorf("Rule = %v, want the rule it sits beneath", criterion.Record.Rule)
	}
	if got, want := strings.Join(criterion.Tags, ","), "AC-CHK-001.1,negative"; got != want {
		t.Errorf("Tags = %q, want %q", got, want)
	}
	if got, want := criterion.Record.Position.StartLine, uint64(5); got != want {
		t.Errorf("Position starts on line %d, want %d", got, want)
	}
}

// AC: AC-GHK-009.4
func TestACriterionStatesTheNotationItWasWrittenIn(t *testing.T) {
	criterion := only(t, parse(t, "Feature: Checkout\n\n  Scenario: pays\n    Given a basket\n"))

	if got, want := criterion.Record.Notation, specmodel.CriterionNotationGherkin; got != want {
		t.Errorf("Notation = %q, want %q", got, want)
	}
}

// AC: AC-GHK-001.3
func TestBothSourceFormsProduceEqualRecords(t *testing.T) {
	src := "@FEAT-CHK-001\n" +
		"Feature: Checkout\n" +
		"  As a shopper\n\n" +
		"  @AC-CHK-001.1\n" +
		"  Scenario: pays\n" +
		"    Given a basket\n"

	asFile := parse(t, src)
	asBlock := fenced(t, "# Checkout\n\nSome prose.\n\n```gherkin\n"+src+"```\n")

	if asFile.Record.Name != asBlock.Record.Name {
		t.Errorf("Name differs: %q as a file, %q as a block", asFile.Record.Name, asBlock.Record.Name)
	}
	if *asFile.Record.Description != *asBlock.Record.Description {
		t.Errorf("Description differs: %q as a file, %q as a block",
			*asFile.Record.Description, *asBlock.Record.Description)
	}
	if asFile.Record.Id != asBlock.Record.Id || asFile.Record.Code != asBlock.Record.Code {
		t.Errorf("identity differs: %q/%q as a file, %q/%q as a block",
			asFile.Record.Id, asFile.Record.Code, asBlock.Record.Id, asBlock.Record.Code)
	}
	fileCriterion, blockCriterion := only(t, asFile), only(t, asBlock)
	if fileCriterion.Record.Text != blockCriterion.Record.Text {
		t.Errorf("Text differs: %q as a file, %q as a block",
			fileCriterion.Record.Text, blockCriterion.Record.Text)
	}
	if fileCriterion.Record.ContentHash != blockCriterion.Record.ContentHash {
		t.Errorf("ContentHash differs: %q as a file, %q as a block",
			fileCriterion.Record.ContentHash, blockCriterion.Record.ContentHash)
	}
	if fileCriterion.Record.Position == blockCriterion.Record.Position {
		t.Error("the two forms report the same position; the test proves nothing")
	}
}

// AC: AC-GHK-005.2
func TestCriteriaAreNestedBeneathTheirFeatureInTheOrderTheyAreWritten(t *testing.T) {
	feature := parse(t, "Feature: Checkout\n\n"+
		"  Scenario: browses\n    Given a catalogue\n\n"+
		"  Rule: a declined card returns the basket\n\n"+
		"    Scenario: pays\n      Given a basket\n\n"+
		"    Scenario: declines\n      Given a declined card\n")

	var names []string
	for _, criterion := range feature.Criteria {
		names = append(names, *criterion.Record.Name)
	}
	if got, want := strings.Join(names, ","), "browses,pays,declines"; got != want {
		t.Errorf("criteria = %q, want %q", got, want)
	}
	if feature.Criteria[0].Record.Rule != nil {
		t.Errorf("Rule = %v for a criterion beneath no rule, want none", feature.Criteria[0].Record.Rule)
	}
}

// A criterion's body is text, and a Background is not a criterion.

func TestACriterionsTextIsItsStepsAndTableVerbatim(t *testing.T) {
	criterion := only(t, parse(t, "Feature: Checkout\n\n"+
		"  Background:\n"+
		"    Given the shop is open\n\n"+
		"  Scenario: pays\n"+
		"    Given a basket\n"+
		"      \"\"\"\n"+
		"      # not a comment\n"+
		"      \"\"\"\n"+
		"    When the card is accepted\n"))

	want := "Given a basket\n  \"\"\"\n  # not a comment\n  \"\"\"\nWhen the card is accepted"
	if got := criterion.Record.Text; got != want {
		t.Errorf("Text = %q, want %q", got, want)
	}
	if strings.Contains(criterion.Record.Text, "the shop is open") {
		t.Errorf("Text = %q, want nothing from the Background", criterion.Record.Text)
	}
}

func TestTrailingBlankAndCommentLinesEndOutsideACriterion(t *testing.T) {
	criterion := only(t, parse(t, "Feature: Checkout\n\n"+
		"  Scenario: pays\n"+
		"    Given a basket\n"+
		"\n"+
		"  # a note about nothing in particular\n"))

	if got, want := criterion.Record.Text, "Given a basket"; got != want {
		t.Errorf("Text = %q, want %q", got, want)
	}
	if got, want := criterion.Record.Position.EndLine, uint64(4); got != want {
		t.Errorf("criterion ends on line %d, want %d", got, want)
	}
}

// Failure.

func TestAParseFailureIsReturnedAsALocatedErrorCarryingTheParsersExpectation(t *testing.T) {
	doc := "# Checkout\n" + // line 1
		"\n" + // line 2
		"```gherkin\n" + // line 3
		"Feature: Checkout\n" + // line 4
		"  Scenario: pays\n" + // line 5
		"    Given a basket\n" + // line 6
		"  Nonsense that is no keyword\n" + // line 7
		"```\n" // line 8

	blocks := markdown.Fenced([]byte(doc), []string{"gherkin"})
	feature, err := gherkin.ParseRegion([]byte(doc), blocks[0].Inner, vocabulary())

	if feature != nil {
		t.Errorf("a source that would not parse yielded records: %+v", feature)
	}
	var failure *gherkin.ParseError
	if !errors.As(err, &failure) {
		t.Fatalf("err = %v (%T), want a *gherkin.ParseError", err, err)
	}
	if len(failure.Errors) != 1 {
		t.Fatalf("failure carries %d errors, want 1: %+v", len(failure.Errors), failure.Errors)
	}
	located := failure.Errors[0]
	if got, want := located.Position.StartLine, uint64(7); got != want {
		t.Errorf("failure is located on line %d, want %d", got, want)
	}
	if !strings.Contains(located.Message, "expected") {
		t.Errorf("Message = %q, want the parser's own expectation", located.Message)
	}
	if strings.HasPrefix(located.Message, "(") {
		t.Errorf("Message = %q, want the position carried as a span rather than in the text", located.Message)
	}
}

func TestASourceWithNoGherkinContentYieldsNoFeature(t *testing.T) {
	for _, src := range []string{"", "\n\n   \n", "# just a comment\n"} {
		feature, err := gherkin.Parse([]byte(src), vocabulary())
		if err != nil {
			t.Errorf("Parse(%q): %v", src, err)
		}
		if feature != nil {
			t.Errorf("Parse(%q) yielded a feature, want none", src)
		}
	}
}

func TestParsingWithNoVocabularyStillYieldsRecords(t *testing.T) {
	feature, err := gherkin.Parse([]byte("@FEAT-CHK-001\nFeature: Checkout\n\n  @AC-CHK-001.1\n  Scenario: pays\n    Given a basket\n"), gherkin.Options{})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if feature == nil {
		t.Fatal("Parse returned no feature")
	}
	if feature.Record.Id != "" {
		t.Errorf("Id = %q, want none — no vocabulary resolved it", feature.Record.Id)
	}
	criterion := only(t, feature)
	if criterion.Record.Id != "" {
		t.Errorf("criterion Id = %q, want none", criterion.Record.Id)
	}
	if got, want := strings.Join(criterion.Tags, ","), "AC-CHK-001.1"; got != want {
		t.Errorf("Tags = %q, want %q — tags are carried whether or not they mean anything", got, want)
	}
}

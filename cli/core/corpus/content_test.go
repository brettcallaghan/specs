// Tests for VCON: the content and diagnostic checks over Gherkin
// records — a criterion that says nothing, prose written where a
// criterion belongs, and an outline whose blocks are tagged
// inconsistently.
package corpus_test

import (
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// AC: AC-GHK-006.1
func TestReportsAGherkinSourceThatWillNotParse(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature", "@FEAT-CHK-002\nFeature: Broken\n\n  @AC-CHK-002.1\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleGherkinParse, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/parse findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingGherkinParse)
	if !ok {
		t.Fatalf("detail = %T, want FindingGherkinParse", got[0].Detail.Value)
	}
	if detail.Expected == "" {
		t.Error("Expected is empty, want the parser's own expectation")
	}
	if got[0].Position.StartLine == 0 {
		t.Error("Position carries no line at all")
	}
}

// A Gherkin source that fails to parse produces no feature record — the
// gherkin/parse finding is the ONLY signal it leaves, which is what makes
// the full recompute (rather than a walk over persisted features) load-
// bearing rather than an implementation convenience.
func TestAParseFailureProducesNoFeatureRecordButStillReportsGherkinParse(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/broken.feature", "@FEAT-CHK-002\nFeature: Broken\n\n  @AC-CHK-002.1\n")

	built := identityCorpusAt(t, repo)
	if contains(featurePaths(built.Index()), "specs/features/broken.feature") {
		t.Fatal("the unparsable source contributed a feature record; this test needs the case where it does not")
	}

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if got := findingsWithRule(findings, finding.RuleGherkinParse, "specs/features/broken.feature"); len(got) != 1 {
		t.Errorf("got %d gherkin/parse findings, want 1: %+v", len(got), findings)
	}
}

// AC: AC-GHK-006.2
func TestReportsACriterionWithNoSteps(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n  @AC-CHK-001.1\n  Scenario: pays\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleCriterionEmpty, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/criterion-empty findings, want 1: %+v", len(got), findings)
	}
	if got[0].Severity != finding.SeverityError {
		t.Errorf("severity = %q, want error", got[0].Severity)
	}
}

// A criterion that DOES carry steps is not reported — the negative test
// the empty-criterion check would pass vacuously without.
func TestNoCriterionEmptyFindingForACriterionCarryingSteps(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n  @AC-CHK-001.1\n  Scenario: pays\n    Given a basket\n")

	findings := buildAndCheck(t, repo)

	if got := findingsWithRule(findings, finding.RuleCriterionEmpty, "specs/features/checkout.feature"); len(got) != 0 {
		t.Errorf("got %d gherkin/criterion-empty findings, want 0: %+v", len(got), got)
	}
}

// AC-GHK-006.2's "content" table, and the rule's own framing as a
// probable defect.
func TestReportsANormativeScenarioDescriptionAsProse(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n"+
			"  @AC-CHK-001.1\n  Scenario: pays\n\n    The system SHALL refuse an empty basket.\n\n"+
			"    Given a basket\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleScenarioProse, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/scenario-prose findings, want 1: %+v", len(got), findings)
	}
	if got[0].Severity != finding.SeverityWarning {
		t.Errorf("severity = %q, want warning — this is a probable, not a certain, defect", got[0].Severity)
	}
}

// AC: AC-GHK-006.4
//
// A Feature description carries the user story and is exempt: writing
// one — even a long, normative-sounding one — must never be reported as
// scenario-prose, which is a per-CRITERION rule.
func TestNoScenarioProseFindingForAFeatureDescription(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n"+
			"  As a shopper\n  I want to pay for a basket\n  So that the goods are mine\n\n"+
			"  The system SHALL refuse an empty basket.\n\n"+
			"  @AC-CHK-001.1\n  Scenario: pays\n    Given a basket\n")

	findings := buildAndCheck(t, repo)

	if got := findingsWithRule(findings, finding.RuleScenarioProse, "specs/features/checkout.feature"); len(got) != 0 {
		t.Errorf("got %d gherkin/scenario-prose findings against a Feature description, want 0: %+v", len(got), got)
	}
}

// A Scenario carrying no description at all is not reported — the
// negative test the prose check would pass vacuously without.
func TestNoScenarioProseFindingWhenTheScenarioCarriesNoDescription(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n  @AC-CHK-001.1\n  Scenario: pays\n    Given a basket\n")

	findings := buildAndCheck(t, repo)

	if got := findingsWithRule(findings, finding.RuleScenarioProse, "specs/features/checkout.feature"); len(got) != 0 {
		t.Errorf("got %d gherkin/scenario-prose findings, want 0: %+v", len(got), got)
	}
}

// AC: AC-GHK-002.7, AC-GHK-006.2
func TestReportsAPartiallyTaggedScenarioOutline(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n"+
			"  Scenario Outline: pays with <method>\n"+
			"    Given a basket\n\n"+
			"    @AC-CHK-001.1\n"+
			"    Examples: accepted\n"+
			"      | method |\n"+
			"      | card   |\n\n"+
			"    Examples: declined\n"+
			"      | method |\n"+
			"      | cash   |\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleExamplesPartiallyTagged, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/examples-partially-tagged findings, want 1: %+v", len(got), findings)
	}
	if got[0].Severity != finding.SeverityError {
		t.Errorf("severity = %q, want error — the mixture is unambiguous", got[0].Severity)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingPartiallyTaggedOutline)
	if !ok {
		t.Fatalf("detail = %T, want FindingPartiallyTaggedOutline", got[0].Detail.Value)
	}
	if len(detail.Untagged) != 1 {
		t.Fatalf("Untagged = %v, want exactly the one untagged block", detail.Untagged)
	}
}

// AC-GHK-002.9's negative case, at the corpus-check level rather than the
// gherkin package's own unit test: an outline whose blocks are ALL
// untagged is one criterion in named groups, not a defect.
func TestNoPartiallyTaggedFindingForAnOutlineWithAllUntaggedBlocks(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n"+
			"  @AC-CHK-001.1\n"+
			"  Scenario Outline: pays with <method>\n"+
			"    Given a basket\n\n"+
			"    Examples: accepted\n"+
			"      | method |\n"+
			"      | card   |\n\n"+
			"    Examples: declined\n"+
			"      | method |\n"+
			"      | cash   |\n")

	findings := buildAndCheck(t, repo)

	if got := findingsWithRule(findings, finding.RuleExamplesPartiallyTagged, "specs/features/checkout.feature"); len(got) != 0 {
		t.Errorf("got %d gherkin/examples-partially-tagged findings, want 0: %+v", len(got), got)
	}
}

// A fully tagged outline is likewise no mixture.
func TestNoPartiallyTaggedFindingForAFullyTaggedOutline(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n"+
			"  Scenario Outline: pays with <method>\n"+
			"    Given a basket\n\n"+
			"    @AC-CHK-001.1\n"+
			"    Examples: accepted\n"+
			"      | method |\n"+
			"      | card   |\n\n"+
			"    @AC-CHK-001.2\n"+
			"    Examples: declined\n"+
			"      | method |\n"+
			"      | cash   |\n")

	findings := buildAndCheck(t, repo)

	if got := findingsWithRule(findings, finding.RuleExamplesPartiallyTagged, "specs/features/checkout.feature"); len(got) != 0 {
		t.Errorf("got %d gherkin/examples-partially-tagged findings, want 0: %+v", len(got), got)
	}
}

// AC: AC-GHK-006.7
//
// Checking one document reports that document's own Gherkin content
// findings.
func TestCheckForReportsOneDocumentsOwnContentFindings(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md",
		blockDocument("frd", "frd-checkout", "Checkout",
			"@FEAT-CHK-001\nFeature: Checkout\n\n  @AC-CHK-001.1\n  Scenario: pays\n"))
	write(t, repo, "specs/features/refunds.feature",
		"@FEAT-CHK-002 @frd-checkout\nFeature: Refunds\n\n  @AC-CHK-002.1\n  Scenario: refunds\n")

	built := identityCorpusAt(t, repo)
	findings, err := built.CheckFor("specs/features/frd-checkout.md")
	if err != nil {
		t.Fatalf("CheckFor: %v", err)
	}

	// The fenced block's criterion-empty finding is against the
	// containing document's own path, so it is scoped in.
	if got := findingsWithRule(findings, finding.RuleCriterionEmpty, "specs/features/frd-checkout.md"); len(got) != 1 {
		t.Errorf("CheckFor(frd-checkout.md) reported %d of its own findings, want 1: %+v", len(got), findings)
	}
	// refunds.feature is a different file entirely, and gherkin/criterion
	// -empty is an ordinary path-scoped rule (unlike identifier-duplicate):
	// its own finding must not leak into a check of another path.
	if got := findingsWithRule(findings, finding.RuleCriterionEmpty, "specs/features/refunds.feature"); len(got) != 0 {
		t.Errorf("CheckFor(frd-checkout.md) reported another file's criterion-empty finding, want 0 (narrowed): %+v", got)
	}
}

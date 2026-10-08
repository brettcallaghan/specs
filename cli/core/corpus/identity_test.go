// Tests for VTAG: the identity checks over Gherkin records — whether a
// feature or a criterion is addressed, addressed once, and whether its
// address agrees with where it sits.
//
// These reuse features_test.go's fixture vocabulary (featureTags,
// featureProfile) rather than inventing a second one: FeaturePattern
// `^FEAT-(?P<code>...)-(?P<number>...)$` and CriterionPattern
// `^AC-(?P<code>...)-(?P<number>...)\.[0-9]+$` both carry a literal
// prefix before their first capture ("FEAT-", "AC-"), which is what lets
// gherkin/identifier-malformed be exercised for both a feature and a
// criterion identifier here — the shipped profile's own FeaturePattern
// admits any uppercase prefix and carries none, so only these fixture
// patterns can prove that rule end to end.
package corpus_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// identityProfile is featureProfile() with a schema that actually
// validates ("adr"/"frd" branches, permissiveCheckSchema from
// check_test.go) rather than featureProfile()'s empty one: features_test.go
// never calls Check(), only Index(), so its schema was never exercised.
func identityProfile() corpus.Profile {
	return stubProfile{
		schema: []byte(permissiveCheckSchema),
		types: map[string]corpus.DocumentTypeRules{
			"adr": {SlugPattern: `^adr-[0-9]{3}$`},
			"frd": {SlugPattern: `^frd-[a-z0-9-]+$`},
		},
		featureTags: featureTags(),
	}
}

// identityCorpusAt builds a corpus under identityProfile(), the same
// repo/specs layout every other fixture in this package uses.
func identityCorpusAt(t *testing.T, repo string) *corpus.Corpus {
	t.Helper()
	built, err := corpus.Build(corpus.Options{
		RepoRoot:  repo,
		Root:      filepath.Join(repo, "specs"),
		Workspace: "fixture-workspace",
		Profile:   identityProfile(),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return built
}

// buildAndCheck builds a corpus under identityProfile() and runs
// Check(), failing the test on either error.
func buildAndCheck(t *testing.T, repo string) []specmodel.Finding {
	t.Helper()
	built := identityCorpusAt(t, repo)
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	return findings
}

// AC: AC-GHK-002.2
func TestReportsAFeatureCarryingNoIdentifierTag(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature", "@frd-checkout\nFeature: Checkout\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleFeatureUntagged, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/feature-untagged findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingIdentifierPattern)
	if !ok {
		t.Fatalf("detail = %T, want FindingIdentifierPattern", got[0].Detail.Value)
	}
	if detail.Identifier != nil {
		t.Errorf("Identifier = %v, want nil: the feature carries no identifier tag at all", detail.Identifier)
	}
	if detail.Pattern != featureTags().FeaturePattern {
		t.Errorf("Pattern = %q, want the registered feature pattern", detail.Pattern)
	}
}

// AC: AC-GHK-002.2, AC-GHK-006.2
func TestReportsACriterionCarryingNoIdentifierTag(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n  Scenario: pays\n    Given a basket\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleCriterionUntagged, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/criterion-untagged findings, want 1: %+v", len(got), findings)
	}
}

// AC: AC-GHK-002.3, AC-GHK-006.5
func TestReportsAMalformedFeatureIdentifier(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	// FEAT-chk-1: shares the pattern's literal "FEAT-" prefix, but the
	// code is lowercase and the number is one digit — a near miss, not an
	// unrelated tag.
	write(t, repo, "specs/features/checkout.feature", "@FEAT-chk-1 @frd-checkout\nFeature: Checkout\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleIdentifierMalformed, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/identifier-malformed findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingIdentifierPattern)
	if !ok {
		t.Fatalf("detail = %T, want FindingIdentifierPattern", got[0].Detail.Value)
	}
	if detail.Identifier == nil || *detail.Identifier != "FEAT-chk-1" {
		t.Errorf("Identifier = %v, want the tag as written, %q", detail.Identifier, "FEAT-chk-1")
	}
	if detail.Pattern != featureTags().FeaturePattern {
		t.Errorf("Pattern = %q, want the registered feature pattern", detail.Pattern)
	}
	// No feature-untagged alongside it: the feature carries an attempted
	// identifier, so "carries no identifier tag at all" would be false.
	if got := findingsWithRule(findings, finding.RuleFeatureUntagged, "specs/features/checkout.feature"); len(got) != 0 {
		t.Errorf("also got %d gherkin/feature-untagged findings, want 0: %+v", len(got), got)
	}
}

// AC: AC-GHK-002.3, AC-GHK-006.5
func TestReportsAMalformedCriterionIdentifier(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n  @AC-CHK-3.1\n  Scenario: pays\n    Given a basket\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleIdentifierMalformed, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/identifier-malformed findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingIdentifierPattern)
	if !ok {
		t.Fatalf("detail = %T, want FindingIdentifierPattern", got[0].Detail.Value)
	}
	if detail.Identifier == nil || *detail.Identifier != "AC-CHK-3.1" {
		t.Errorf("Identifier = %v, want %q", detail.Identifier, "AC-CHK-3.1")
	}
}

// A tag matching no registered pattern and named by no qualifier is
// ignored, not reported — including by gherkin/identifier-malformed,
// which must not treat an unrelated project tag as a near-miss
// identifier.
//
// AC: AC-GHK-002.8
func TestNoFindingForATagMatchingNoRegisteredPattern(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n  @wip\n  Scenario: pays\n    Given a basket\n")

	findings := buildAndCheck(t, repo)

	if got := findingsWithRule(findings, finding.RuleIdentifierMalformed, "specs/features/checkout.feature"); len(got) != 0 {
		t.Errorf("got %d gherkin/identifier-malformed findings for an unrelated tag, want 0: %+v", len(got), got)
	}
	// The scenario still carries no identifier, which IS reported —
	// distinctly from treating @wip itself as malformed.
	if got := findingsWithRule(findings, finding.RuleCriterionUntagged, "specs/features/checkout.feature"); len(got) != 1 {
		t.Errorf("got %d gherkin/criterion-untagged findings, want 1", len(got))
	}
}

// AC: AC-GHK-002.5
func TestReportsAFeatureCarryingMoreThanOneMatchingIdentifierTag(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @FEAT-CHK-002 @frd-checkout\nFeature: Checkout\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleIdentifierMultiple, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/identifier-multiple findings, want 1: %+v", len(got), findings)
	}
	if !strings.Contains(got[0].Message, "FEAT-CHK-001") || !strings.Contains(got[0].Message, "FEAT-CHK-002") {
		t.Errorf("message = %q, want it to name both tags", got[0].Message)
	}
}

// AC: AC-GHK-002.5
func TestReportsACriterionCarryingMoreThanOneMatchingIdentifierTag(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n"+
			"  @AC-CHK-001.1 @AC-CHK-001.2\n  Scenario: pays\n    Given a basket\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleIdentifierMultiple, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/identifier-multiple findings, want 1: %+v", len(got), findings)
	}
}

// AC: AC-GHK-003.4
func TestReportsADuplicateFeatureIdentifierAgainstBothOccurrences(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature", "@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n")
	write(t, repo, "specs/features/checkout2.feature", "@FEAT-CHK-001 @frd-checkout\nFeature: Checkout again\n")

	findings := buildAndCheck(t, repo)

	first := findingsWithRule(findings, finding.RuleIdentifierDuplicate, "specs/features/checkout.feature")
	second := findingsWithRule(findings, finding.RuleIdentifierDuplicate, "specs/features/checkout2.feature")
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("got %d and %d gherkin/identifier-duplicate findings, want 1 each: %+v", len(first), len(second), findings)
	}
	detail, ok := first[0].Detail.Value.(specmodel.FindingDuplicateIdentifier)
	if !ok {
		t.Fatalf("detail = %T, want FindingDuplicateIdentifier", first[0].Detail.Value)
	}
	if detail.Identifier != "FEAT-CHK-001" || detail.OtherPath != "specs/features/checkout2.feature" {
		t.Errorf("detail = %+v, want identifier FEAT-CHK-001 naming the other file", detail)
	}
}

// AC: AC-GHK-002.4
//
// Distinct from the feature-identifier case above: two CRITERIA sharing
// an identifier, beneath two features that do not themselves collide —
// proving duplicate detection is grouped separately per vocabulary
// rather than accidentally satisfied by the feature-level test alone.
func TestReportsADuplicateCriterionIdentifierAgainstBothOccurrences(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n  @AC-CHK-001.1\n  Scenario: pays\n    Given a basket\n")
	write(t, repo, "specs/features/checkout2.feature",
		"@FEAT-CHK-002 @frd-checkout\nFeature: Checkout again\n\n  @AC-CHK-001.1\n  Scenario: pays too\n    Given a basket\n")

	findings := buildAndCheck(t, repo)

	if got := findingsWithRule(findings, finding.RuleIdentifierDuplicate, "specs/features/checkout.feature"); len(got) != 1 {
		t.Fatalf("got %d gherkin/identifier-duplicate findings against the first occurrence, want 1: %+v", len(got), findings)
	}
	got := findingsWithRule(findings, finding.RuleIdentifierDuplicate, "specs/features/checkout2.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/identifier-duplicate findings against the second occurrence, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingDuplicateIdentifier)
	if !ok {
		t.Fatalf("detail = %T, want FindingDuplicateIdentifier", got[0].Detail.Value)
	}
	if detail.Identifier != "AC-CHK-001.1" || detail.OtherPath != "specs/features/checkout.feature" {
		t.Errorf("detail = %+v, want identifier AC-CHK-001.1 naming the other file", detail)
	}
}

// Several features declaring the same code with different numbers is a
// grouping, not a defect — the negative test gherkin/identifier-duplicate
// and gherkin/criterion-code-mismatch would both fail without.
//
// AC: AC-GHK-003.5
func TestNoFindingForSeveralFeaturesSharingACode(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature", "@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n")
	write(t, repo, "specs/features/refunds.feature", "@FEAT-CHK-002 @frd-checkout\nFeature: Refunds\n")

	findings := buildAndCheck(t, repo)

	for _, rule := range []specmodel.RuleId{finding.RuleIdentifierDuplicate, finding.RuleCriterionCodeMismatch} {
		if got := findingsWithRule(findings, rule, "specs/features/checkout.feature"); len(got) != 0 {
			t.Errorf("got %d %s findings for a shared code, want 0: %+v", len(got), rule, got)
		}
	}
}

// AC: AC-GHK-003.6
func TestReportsACriterionCodeMismatch(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n"+
			"  @AC-REF-001.1\n  Scenario: pays\n    Given a basket\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleCriterionCodeMismatch, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/criterion-code-mismatch findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingIdentifierPattern)
	if !ok {
		t.Fatalf("detail = %T, want FindingIdentifierPattern", got[0].Detail.Value)
	}
	if detail.Identifier == nil || *detail.Identifier != "AC-REF-001.1" {
		t.Errorf("Identifier = %v, want %q", detail.Identifier, "AC-REF-001.1")
	}
	if !strings.Contains(got[0].Message, "FEAT-CHK-001") {
		t.Errorf("message = %q, want it to name the feature the criterion should have matched", got[0].Message)
	}
}

// A criterion whose code and number DO agree with its feature's is not
// reported — the negative test TestReportsACriterionCodeMismatch would
// otherwise pass vacuously without.
func TestNoCodeMismatchFindingWhenTheCriterionAgreesWithItsFeature(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature",
		"@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n\n"+
			"  @AC-CHK-001.1\n  Scenario: pays\n    Given a basket\n")

	findings := buildAndCheck(t, repo)

	if got := findingsWithRule(findings, finding.RuleCriterionCodeMismatch, "specs/features/checkout.feature"); len(got) != 0 {
		t.Errorf("got %d gherkin/criterion-code-mismatch findings, want 0: %+v", len(got), got)
	}
}

// AC: AC-GHK-002.10
func TestReportsAFeatureFileWithNoOwningDocumentTag(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature", "@FEAT-CHK-001\nFeature: Checkout\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleOwnerMissing, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/owner-missing findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingOwningDocument)
	if !ok {
		t.Fatalf("detail = %T, want FindingOwningDocument", got[0].Detail.Value)
	}
	if detail.Slug != nil || detail.DocumentSlug != nil {
		t.Errorf("detail = %+v, want both halves absent", detail)
	}
}

// AC: AC-GHK-002.11
func TestReportsAFencedBlockWhoseOwnerTagNamesAnotherDocument(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/frd-refunds.md",
		blockDocument("frd", "frd-refunds", "Refunds", "@FEAT-CHK-001 @frd-checkout\nFeature: Checkout\n"))

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleOwnerMismatch, "specs/features/frd-refunds.md")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/owner-mismatch findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingOwningDocument)
	if !ok {
		t.Fatalf("detail = %T, want FindingOwningDocument", got[0].Detail.Value)
	}
	if detail.Slug == nil || *detail.Slug != "frd-checkout" {
		t.Errorf("Slug = %v, want the tag's own value %q", detail.Slug, "frd-checkout")
	}
	if detail.DocumentSlug == nil || *detail.DocumentSlug != "frd-refunds" {
		t.Errorf("DocumentSlug = %v, want the containing document's slug %q", detail.DocumentSlug, "frd-refunds")
	}
}

// A fenced block carrying no owner tag at all is not a mismatch — it is
// simply owned by its containing document, exactly as AC-GHK-001.7
// requires, and this is the negative test the mismatch check would pass
// vacuously without.
func TestNoOwnerMismatchFindingForAFencedBlockCarryingNoOwnerTag(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md",
		blockDocument("frd", "frd-checkout", "Checkout", "@FEAT-CHK-001\nFeature: Checkout\n"))

	findings := buildAndCheck(t, repo)

	if got := findingsWithRule(findings, finding.RuleOwnerMismatch, "specs/features/frd-checkout.md"); len(got) != 0 {
		t.Errorf("got %d gherkin/owner-mismatch findings, want 0: %+v", len(got), got)
	}
	if got := findingsWithRule(findings, finding.RuleOwnerMissing, "specs/features/frd-checkout.md"); len(got) != 0 {
		t.Errorf("got %d gherkin/owner-missing findings for a fenced block, want 0: it needs no tag to say so", len(got))
	}
}

// AC: AC-GHK-002.12
func TestReportsAnOwnerTagNamingNoCorpusMember(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/features/checkout.feature", "@FEAT-CHK-001 @frd-does-not-exist\nFeature: Checkout\n")

	findings := buildAndCheck(t, repo)

	got := findingsWithRule(findings, finding.RuleOwnerUnresolved, "specs/features/checkout.feature")
	if len(got) != 1 {
		t.Fatalf("got %d gherkin/owner-unresolved findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingOwningDocument)
	if !ok {
		t.Fatalf("detail = %T, want FindingOwningDocument", got[0].Detail.Value)
	}
	if detail.Slug == nil || *detail.Slug != "frd-does-not-exist" {
		t.Errorf("Slug = %v, want %q", detail.Slug, "frd-does-not-exist")
	}
	if detail.DocumentSlug != nil {
		t.Errorf("DocumentSlug = %v, want nil: a feature file names no document of its own", detail.DocumentSlug)
	}
	// Not also reported as owner-missing: a tag WAS written.
	if got := findingsWithRule(findings, finding.RuleOwnerMissing, "specs/features/checkout.feature"); len(got) != 0 {
		t.Errorf("also got %d gherkin/owner-missing findings, want 0", len(got))
	}
}

// AC-GHK-002.7's own untagged block must not also be reported as an
// ordinary gherkin/criterion-untagged Scenario: the partial-tagging rule
// is the one that names it, and reporting both would double-count one
// defect under two rules.
func TestAPartiallyTaggedOutlinesUntaggedBlockIsNotAlsoReportedAsCriterionUntagged(t *testing.T) {
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

	if got := findingsWithRule(findings, finding.RuleCriterionUntagged, "specs/features/checkout.feature"); len(got) != 0 {
		t.Errorf("got %d gherkin/criterion-untagged findings for a partially tagged outline's block, want 0: %+v", len(got), got)
	}
}

// CheckFor narrows every Gherkin identity finding to the file named, and
// still resolves identifier uniqueness against the whole corpus: an
// identifier the named document shares with another file is reported in
// the named document, naming the other location, while findings located
// in other files are not reported at all (AC-SPECREF-002.9).
//
// AC: AC-GHK-006.7
func TestCheckForResolvesIdentifierUniquenessCorpusWideAndReportsOnlyTheNamedDocument(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/frd-checkout.md",
		blockDocument("frd", "frd-checkout", "Checkout", "@FEAT-CHK-001\nFeature: Checkout\n"))
	write(t, repo, "specs/features/checkout2.feature", "@FEAT-CHK-001 @frd-checkout\nFeature: Checkout again\n")
	write(t, repo, "specs/features/refunds.feature", "@frd-checkout\nFeature: Refunds\n") // untagged, elsewhere

	built := identityCorpusAt(t, repo)
	findings, err := built.CheckFor("specs/features/frd-checkout.md")
	if err != nil {
		t.Fatalf("CheckFor: %v", err)
	}

	duplicates := findingsWithRule(findings, finding.RuleIdentifierDuplicate, "specs/features/frd-checkout.md")
	if len(duplicates) != 1 {
		t.Fatalf("CheckFor(frd-checkout) reported %d duplicate findings in the named document, want 1: %+v", len(duplicates), findings)
	}
	if !strings.Contains(duplicates[0].Message, "checkout2.feature") {
		t.Errorf("message = %q, want it to name the other location", duplicates[0].Message)
	}
	for _, f := range findings {
		if f.Path != "specs/features/frd-checkout.md" {
			t.Errorf("CheckFor(frd-checkout) reported a finding located in %s: %s %s", f.Path, f.Rule, f.Message)
		}
	}
}

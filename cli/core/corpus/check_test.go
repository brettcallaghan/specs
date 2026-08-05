// Tests for Check and CheckFor — the corpus-level entry point that runs
// the Schema Registry's validation and the active profile's slug-pattern
// check over a built corpus, folding the results into the same report the
// corpus's own duplicate-slug findings already flow through.
//
// The fixture profile below registers "adr" and "frd" — the same two
// types profile_test.go and corpus_test.go use — with schemas small
// enough to be violated on purpose. Nothing here names a real document
// type from the generated vocabulary, for the same reason corpus_test.go
// does not: TestTheExtensionPointNamesNoConcreteDocumentType enforces that
// independently.
package corpus_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// checkSchema's "Adr" branch requires a "decision" property none of this
// file's fixture documents declare, which is what makes a plain "adr"
// document schema-invalid; "Frd" requires nothing, so a duplicate-slug
// fixture built from "frd" documents raises no schema findings alongside
// the duplicate-slug ones.
const checkSchema = `{
  "definitions": {
    "Adr": {
      "type": "object",
      "required": ["decision"],
      "properties": {"decision": {"type": "string"}}
    },
    "Frd": {
      "type": "object",
      "properties": {}
    }
  }
}`

// formatCheckSchema constrains "released_at" with a format keyword and
// nothing else, isolating the raw-text-vs-coerced-value behavior
// AC-SPEC-003.2 governs from the required-property behavior AC-SPEC-003.1
// covers.
const formatCheckSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "definitions": {
    "Adr": {
      "type": "object",
      "properties": {
        "released_at": {"type": "string", "format": "date-time"}
      }
    },
    "Frd": {"type": "object", "properties": {}}
  }
}`

// permissiveCheckSchema requires nothing of either type, so a corpus
// checked against it is schema-clean regardless of its documents'
// content — the fixture a clean-corpus test and a slug-pattern-only test
// both need.
const permissiveCheckSchema = `{
  "definitions": {
    "Adr": {"type": "object", "properties": {}},
    "Frd": {"type": "object", "properties": {}}
  }
}`

// checkProfile registers "adr" and "frd" with slug patterns strict enough
// to be violated on purpose, and the given union schema.
func checkProfile(schema string) corpus.Profile {
	return stubProfile{
		schema: []byte(schema),
		types: map[string]corpus.DocumentTypeRules{
			"adr": {SlugPattern: `^adr-[0-9]{3}$`},
			"frd": {SlugPattern: `^frd-[a-z0-9-]+$`},
		},
	}
}

// buildChecked builds a corpus under repo/specs against profile, the same
// layout corpusAt uses.
func buildChecked(t *testing.T, repo string, profile corpus.Profile) *corpus.Corpus {
	t.Helper()

	built, err := corpus.Build(corpus.Options{
		RepoRoot:  repo,
		Root:      filepath.Join(repo, "specs"),
		Workspace: "fixture-workspace",
		Profile:   profile,
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return built
}

// documentWithExtra renders a conforming document of the given type,
// carrying one additional frontmatter property written exactly as given —
// unquoted where the caller wants a value the YAML parser coerces away
// from a string, which is what a raw-text-vs-coerced-value fixture needs.
func documentWithExtra(docType, slug, title, key, rawValue string) string {
	return fmt.Sprintf(`---
title: %q
description: "A fixture document."
slug: %s
type: %s
status: approved
updated_at: "2026-08-05T00:00:00Z"
%s: %s
---

# %s
`, title, slug, docType, key, rawValue, title)
}

// findingsWithRule returns every finding of one rule against one path.
func findingsWithRule(findings []specmodel.Finding, rule specmodel.RuleId, path string) []specmodel.Finding {
	var out []specmodel.Finding
	for _, f := range findings {
		if f.Rule == rule && f.Path == path {
			out = append(out, f)
		}
	}
	return out
}

// AC: AC-SPEC-003.1
func TestCheckReportsASchemaInvalidMember(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))

	built := buildChecked(t, repo, checkProfile(checkSchema))

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	missing := findingsWithRule(findings, finding.RuleMissingProperty, "specs/adrs/adr-001.md")
	if len(missing) != 1 {
		t.Fatalf("got %d missing-property findings against the schema-invalid member, want 1: %+v", len(missing), findings)
	}
	detail, ok := missing[0].Detail.Value.(specmodel.FindingMissingProperty)
	if !ok {
		t.Fatalf("detail = %T, want FindingMissingProperty", missing[0].Detail.Value)
	}
	if detail.Property != "decision" {
		t.Errorf("property = %q, want %q", detail.Property, "decision")
	}
}

// AC: AC-SPEC-003.2
//
// Mirrors schemaregistry_test.go's TestValidateChecksRawTextAgainstAFormatConstraint
// end to end through the corpus: an unquoted value the YAML parser
// coerces to a number is still checked, at corpus level, against its raw
// source text rather than the coerced value.
func TestCheckValidatesAFormatConstrainedPropertyAgainstItsRawSourceText(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md",
		documentWithExtra("adr", "adr-001", "First", "released_at", "20260729"))

	built := buildChecked(t, repo, checkProfile(formatCheckSchema))

	// Confirm the premise: the YAML parser coerces the unquoted digits to
	// a number, not a string. If it did not, this test would not be
	// exercising the substitution at all.
	doc, ok := built.Document("specs/adrs/adr-001.md")
	if !ok {
		t.Fatal("Document did not return the parsed frontmatter")
	}
	property, ok := doc.Property("released_at")
	if !ok {
		t.Fatal("the fixture document does not declare released_at")
	}
	if _, isString := property.Value.(string); isString {
		t.Fatalf("released_at Value = %#v (a string); this fixture needs a value YAML coerces away from string", property.Value)
	}

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	violations := findingsWithRule(findings, finding.RuleSchemaViolation, "specs/adrs/adr-001.md")
	if len(violations) != 1 {
		t.Fatalf("got %d schema-violation findings, want 1: %+v", len(violations), findings)
	}
	detail, ok := violations[0].Detail.Value.(specmodel.FindingSchemaViolation)
	if !ok {
		t.Fatalf("detail = %T, want FindingSchemaViolation", violations[0].Detail.Value)
	}
	if detail.Constraint != "format" {
		t.Errorf("constraint = %q, want %q (a coerced number would instead fail \"type\")", detail.Constraint, "format")
	}
	if detail.Value != "20260729" {
		t.Errorf("reported value = %q, want the exact source text %q", detail.Value, "20260729")
	}
}

// AC: AC-SPEC-003.5
func TestCheckReportsASlugPatternViolation(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-abc.md", document("adr", "adr-abc", "Bad slug"))

	built := buildChecked(t, repo, checkProfile(permissiveCheckSchema))

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	violations := findingsWithRule(findings, finding.RuleSlugPattern, "specs/adrs/adr-abc.md")
	if len(violations) != 1 {
		t.Fatalf("got %d slug-pattern findings, want 1: %+v", len(violations), findings)
	}
	detail, ok := violations[0].Detail.Value.(specmodel.FindingSlugPattern)
	if !ok {
		t.Fatalf("detail = %T, want FindingSlugPattern", violations[0].Detail.Value)
	}
	if detail.Slug != "adr-abc" {
		t.Errorf("slug = %q, want %q", detail.Slug, "adr-abc")
	}
	if detail.Pattern != `^adr-[0-9]{3}$` {
		t.Errorf("pattern = %q, want %q", detail.Pattern, `^adr-[0-9]{3}$`)
	}
}

// A corpus whose members conform to both the schema and the slug pattern
// raises no findings from Check — the check that keeps the tests above
// from passing vacuously.
func TestCheckRaisesNoFindingsForACleanCorpus(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))

	built := buildChecked(t, repo, checkProfile(permissiveCheckSchema))

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("findings = %+v, want none", findings)
	}
}

// Check's findings and the corpus's own duplicate-slug findings come back
// as one ordered list rather than two channels a caller must reconcile:
// this builds a corpus with both a duplicate slug and a schema-invalid
// member and checks that every one of them is present, sorted by path.
func TestCheckFoldsSchemaAndDuplicateSlugFindingsIntoOneOrderedReport(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "Missing decision"))
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	// The exact same slug, not merely a case variant, so this fixture
	// raises a duplicate-slug finding and nothing else from the slug
	// check — the schema-invalid document above is this test's only other
	// source of findings.
	write(t, repo, "specs/legacy/old-checkout.md", document("frd", "frd-checkout", "Checkout again"))

	built := buildChecked(t, repo, checkProfile(checkSchema))

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	if got := findingsWithRule(findings, finding.RuleMissingProperty, "specs/adrs/adr-001.md"); len(got) != 1 {
		t.Errorf("got %d missing-property findings, want 1: %+v", len(got), findings)
	}
	var duplicates int
	for _, f := range findings {
		if f.Rule == finding.RuleDuplicateSlug {
			duplicates++
		}
	}
	if duplicates != 2 {
		t.Errorf("got %d duplicate-slug findings, want 2: %+v", duplicates, findings)
	}
	if len(findings) != 3 {
		t.Fatalf("got %d findings total, want 3 (one report, not two channels): %+v", len(findings), findings)
	}

	if !sort.SliceIsSorted(findings, func(i, j int) bool { return findings[i].Path < findings[j].Path }) {
		t.Errorf("findings are not ordered by path, so this did not go through one finding.Report: %+v", findings)
	}
}

// AC: AC-SPEC-001.2
//
// CheckFor narrows output to one file while still resolving against the
// full corpus, the same behavior FindingsFor already proves for
// duplicate-slug detection alone (TestASingleCheckedFileResolvesAgainstTheFullCorpus).
func TestCheckForNarrowsToOneFileButResolvesAgainstTheFullCorpus(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "Missing decision"))
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/legacy/old-checkout.md", document("frd", "FRD-Checkout", "Checkout again"))

	built := buildChecked(t, repo, checkProfile(checkSchema))

	// The duplicate-slug collision is only visible relative to a document
	// in a directory this call never names, so this proves the schema
	// check is scoped to one file while duplicate detection still sees the
	// whole corpus.
	findings, err := built.CheckFor("specs/features/frd-checkout.md")
	if err != nil {
		t.Fatalf("CheckFor: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings for the checked file, want 1: %+v", len(findings), findings)
	}
	if findings[0].Rule != finding.RuleDuplicateSlug {
		t.Errorf("rule = %q, want %q", findings[0].Rule, finding.RuleDuplicateSlug)
	}

	// The schema-invalid file, checked alone, reports only its own defect.
	findings, err = built.CheckFor("specs/adrs/adr-001.md")
	if err != nil {
		t.Fatalf("CheckFor: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings for the schema-invalid file, want 1: %+v", len(findings), findings)
	}
	if findings[0].Rule != finding.RuleMissingProperty {
		t.Errorf("rule = %q, want %q", findings[0].Rule, finding.RuleMissingProperty)
	}
}

// Corpus exposes the parsed frontmatter document per member so Check does
// not re-parse; this is the accessor itself, proved directly.
func TestDocumentReturnsTheParsedFrontmatterForACorpusMember(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/templates/broken.md", "---\ntitle: [unclosed\n---\n\n# Broken\n")

	built := buildChecked(t, repo, checkProfile(permissiveCheckSchema))

	doc, ok := built.Document("specs/adrs/adr-001.md")
	if !ok {
		t.Fatal("Document reported no document for a member with parsed frontmatter")
	}
	slug, ok := doc.Property("slug")
	if !ok || slug.Value != "adr-001" {
		t.Errorf("Document returned the wrong parse: slug property = %+v, ok=%v", slug, ok)
	}

	// A member recorded against itself with a parse error carries no
	// parsed document to return.
	if _, ok := built.Document("specs/templates/broken.md"); ok {
		t.Error("Document returned a document for a member that failed to parse")
	}

	// An unknown path returns false rather than a zero value silently
	// mistaken for a real, empty document.
	if _, ok := built.Document("specs/does-not-exist.md"); ok {
		t.Error("Document returned true for a path with no entry")
	}
}

// Check's output stays a deterministic function of corpus content once
// schema and slug-pattern findings are present, the same guarantee
// TestEntriesExclusionsAndFindingsAreOrderedDeterministically proves for
// Build alone.
func TestCheckOutputIsDeterministic(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "Missing decision"))
	write(t, repo, "specs/adrs/adr-abc.md", document("adr", "adr-abc", "Bad slug"))
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))
	write(t, repo, "specs/legacy/old-checkout.md", document("frd", "FRD-Checkout", "Checkout again"))

	encode := func() string {
		t.Helper()
		findings, err := buildChecked(t, repo, checkProfile(checkSchema)).Check()
		if err != nil {
			t.Fatalf("Check: %v", err)
		}
		out, err := json.Marshal(findings)
		if err != nil {
			t.Fatalf("marshalling: %v", err)
		}
		return string(out)
	}

	if a, b := encode(), encode(); a != b {
		t.Errorf("two Check runs over one corpus differ:\n%s\n%s", a, b)
	}
}

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
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
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

// envelopeCheckSchema requires "description" — a property of the shared
// envelope rather than a type-specific one. checkSchema models only
// "decision", so it cannot express the case
// TestAMissingEnvelopePropertyIsAMissingPropertyNotAParseFailure needs:
// a required property that the generated envelope decoder also knows
// about, and would reject before any schema check could run.
const envelopeCheckSchema = `{
  "definitions": {
    "Adr": {
      "type": "object",
      "required": ["description"],
      "properties": {"description": {"type": "string"}}
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
%s: %s
---

# %s
`, title, slug, docType, key, rawValue, title)
}

// undecodableDocument renders a document whose frontmatter is valid YAML
// but does not fit the document envelope: status carries an unfilled
// placeholder, which is outside the status vocabulary.
//
// This is the real-world shape of the defect, not a
// contrived one: the spec skill instructs an author to copy a template
// and then to run `specs check`, so a document with a placeholder still
// in it is exactly what that second step sees.
func undecodableDocument(docType, slug, title string) string {
	return fmt.Sprintf(`---
title: %q
description: "A fixture document."
slug: %s
type: %s
status: <draft | approved>
---

# %s
`, title, slug, docType, title)
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

// envelopeDocument renders a document missing one required property of
// the shared envelope — the properties DocumentFrontmatter itself
// declares, as opposed to the type-specific "decision" the test above
// covers. Only these reach the generated envelope decoder.
func envelopeDocument(docType, slug, title string) string {
	return fmt.Sprintf(`---
title: %q
slug: %s
type: %s
status: approved
---

# %s
`, title, slug, docType, title)
}

// AC: AC-SPEC-003.1
//
// A member missing a required ENVELOPE property reports
// frontmatter/missing-property, naming the property — never
// frontmatter/parse.
//
// This guards a regression that regenerating the model can silently
// reintroduce. Strata's Go target emits an UnmarshalJSON per struct that
// rejects a missing required field, so decoding straight into
// specmodel.DocumentFrontmatter turns every such document into a parse
// failure and leaves frontmatter/missing-property unable to fire at all —
// a check that reports success because it never runs. corpus.go decodes
// through a defined type to drop that method; see envelopeShape there.
//
// TestCheckReportsASchemaInvalidMember does not cover this: "decision" is
// type-specific, absent from the envelope, so the generated decoder never
// sees it and that test passes either way.
func TestAMissingEnvelopePropertyIsAMissingPropertyNotAParseFailure(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", envelopeDocument("adr", "adr-001", "First"))

	built := buildChecked(t, repo, checkProfile(envelopeCheckSchema))

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	if got := findingsWithRule(findings, finding.RuleFrontmatterParse, "specs/adrs/adr-001.md"); len(got) != 0 {
		t.Errorf("got %d frontmatter/parse findings, want 0 — the envelope decode is enforcing required-ness: %+v", len(got), got)
	}

	missing := findingsWithRule(findings, finding.RuleMissingProperty, "specs/adrs/adr-001.md")
	if len(missing) != 1 {
		t.Fatalf("got %d missing-property findings, want 1: %+v", len(missing), findings)
	}
	detail, ok := missing[0].Detail.Value.(specmodel.FindingMissingProperty)
	if !ok {
		t.Fatalf("detail = %T, want FindingMissingProperty", missing[0].Detail.Value)
	}
	if detail.Property != "description" {
		t.Errorf("property = %q, want %q", detail.Property, "description")
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

// AC: AC-SPEC-003.6
//
// The regression test for parse-error members. Before the fix, Check
// skipped every entry with no parsed frontmatter, so this corpus
// produced zero findings and every rule reported `passed` — the worst
// document in a corpus was the one the checker was guaranteed to be
// silent about.
func TestCheckReportsAMemberWhoseFrontmatterCouldNotBeDecoded(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", undecodableDocument("adr", "adr-001", "Placeholder status"))

	built := buildChecked(t, repo, checkProfile(checkSchema))

	// Confirm the premise: the document IS a corpus member, indexed with
	// a parse error rather than excluded. A fixture that was merely
	// excluded would exercise a different branch and pass vacuously.
	var entry *specmodel.IndexEntry
	for i, e := range built.Index().Entries {
		if e.Path == "specs/adrs/adr-001.md" {
			entry = &built.Index().Entries[i]
		}
	}
	if entry == nil {
		t.Fatal("the fixture is not a corpus member; this test needs one indexed with a parse error")
	}
	if entry.Frontmatter != nil || entry.ParseError == nil {
		t.Fatalf("fixture entry has frontmatter=%v parse_error=%v; want no frontmatter and a parse error",
			entry.Frontmatter != nil, entry.ParseError)
	}

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	parse := findingsWithRule(findings, finding.RuleFrontmatterParse, "specs/adrs/adr-001.md")
	if len(parse) != 1 {
		t.Fatalf("got %d frontmatter/parse findings, want 1: %+v", len(parse), findings)
	}
	if parse[0].Severity != finding.SeverityError {
		t.Errorf("severity = %q, want %q — a document that claimed to be a specification and got the claim wrong is a conformance failure, not an unrun check",
			parse[0].Severity, finding.SeverityError)
	}
	detail, ok := parse[0].Detail.Value.(specmodel.FindingFrontmatterParse)
	if !ok {
		t.Fatalf("detail = %T, want FindingFrontmatterParse", parse[0].Detail.Value)
	}
	if detail.Message != entry.ParseError.Message {
		t.Errorf("detail message = %q, want the index's own wording %q — the finding and the index must not describe the same failure twice",
			detail.Message, entry.ParseError.Message)
	}
	if parse[0].Span != entry.ParseError.Span {
		t.Errorf("position = %+v, want the index's own span %+v", parse[0].Span, entry.ParseError.Span)
	}
}

// A placeholder must not reach a reader HTML-escaped. json.Marshal rewrites <, > and & as \u003c, \u003e
// and \u0026 by default, and the decode failure quotes the source text
// it saw, so a diagnostic quoting a template placeholder read
// `\u003cdraft | approved\u003e` until Document.JSON turned escaping off.
func TestAParseFindingQuotesThePlaceholderWithoutHTMLEscaping(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", undecodableDocument("adr", "adr-001", "Placeholder status"))

	built := buildChecked(t, repo, checkProfile(checkSchema))
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	parse := findingsWithRule(findings, finding.RuleFrontmatterParse, "specs/adrs/adr-001.md")
	if len(parse) != 1 {
		t.Fatalf("got %d frontmatter/parse findings, want 1", len(parse))
	}
	// The escaped forms json.Marshal emits for < and >, built from
	// parts so this file never contains the sequence it forbids.
	escapedLT := "\\" + "u003c"
	escapedGT := "\\" + "u003e"
	if strings.Contains(parse[0].Message, escapedLT) || strings.Contains(parse[0].Message, escapedGT) {
		t.Errorf("message carries HTML escapes: %q", parse[0].Message)
	}
	if !strings.Contains(parse[0].Message, "<draft | approved>") {
		t.Errorf("message does not quote the placeholder as the author wrote it: %q", parse[0].Message)
	}
}

// A file with no frontmatter at all is an EXCLUSION, not a member, and
// must not gain a parse finding: it never claimed to be a specification.
// This is the boundary the fix must not cross — the distinction the old
// doc comment collapsed by treating both cases as "already reported".
func TestCheckRaisesNoParseFindingForAFileWithNoFrontmatter(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/README.md", "# Just an index page\n\nNo frontmatter here.\n")
	write(t, repo, "specs/adrs/adr-001.md", documentWithExtra("adr", "adr-001", "First", "decision", "\"Chosen.\""))

	built := buildChecked(t, repo, checkProfile(checkSchema))
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	for _, f := range findings {
		if f.Rule == finding.RuleFrontmatterParse {
			t.Errorf("a file with no frontmatter raised %s: %+v", f.Rule, f)
		}
	}
	if len(built.Index().Exclusions) != 1 {
		t.Fatalf("exclusions = %+v, want the README recorded as one", built.Index().Exclusions)
	}
}

// Tests for reference/unresolved: the check that reports a reference the
// graph recorded as unresolved or as reaching a non-member, and — for a
// single-document check — the inbound edges that no longer resolve.

// referenceCheckProfile is checkProfile(permissiveCheckSchema) with
// property registered as reference-bearing on both fixture types, so
// nothing about a document's schema or slug pattern can produce a
// finding and only the reference itself can.
func referenceCheckProfile(property string) corpus.Profile {
	types := map[string]corpus.DocumentTypeRules{
		"adr": {SlugPattern: `^adr-[0-9]{3}$`},
		"frd": {SlugPattern: `^frd-[a-z0-9-]+$`},
	}
	for name, rules := range types {
		rules.ReferenceProperties = []string{property}
		types[name] = rules
	}
	return stubProfile{schema: []byte(permissiveCheckSchema), types: types}
}

// AC: AC-SPECREF-002.1
func TestCheckReportsAnUnresolvedBodyReference(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Refines [[adr-404]]."))

	built := buildChecked(t, repo, referenceCheckProfile("superseded_by"))

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	got := findingsWithRule(findings, finding.RuleUnresolvedReference, "specs/adrs/adr-001.md")
	if len(got) != 1 {
		t.Fatalf("got %d reference/unresolved findings, want 1: %+v", len(got), findings)
	}
	if got[0].Severity != finding.SeverityError {
		t.Errorf("severity = %q, want error", got[0].Severity)
	}
	if got[0].Span.Length == 0 {
		t.Error("finding carries an empty span")
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingUnresolvedReference)
	if !ok {
		t.Fatalf("detail = %T, want FindingUnresolvedReference", got[0].Detail.Value)
	}
	if detail.Slug != "adr-404" {
		t.Errorf("slug = %q, want adr-404", detail.Slug)
	}
	if _, ok := detail.Origin.AsBody(); !ok {
		t.Errorf("origin = %+v, want body", detail.Origin.Value)
	}
	if detail.NonMember != nil {
		t.Errorf("non_member = %+v, want none: adr-404 matches no file at all", detail.NonMember)
	}
	if !strings.Contains(got[0].Message, "adr-404") {
		t.Errorf("message = %q, want it to quote the reference back", got[0].Message)
	}
}

// AC: AC-SPECREF-002.3
func TestCheckReportsAnUnresolvedFrontmatterReference(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", strings.Replace(
		document("adr", "adr-001", "First"),
		"status: approved\n",
		"status: approved\nsuperseded_by: adr-999\n",
		1))

	built := buildChecked(t, repo, referenceCheckProfile("superseded_by"))

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	got := findingsWithRule(findings, finding.RuleUnresolvedReference, "specs/adrs/adr-001.md")
	if len(got) != 1 {
		t.Fatalf("got %d reference/unresolved findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingUnresolvedReference)
	if !ok {
		t.Fatalf("detail = %T, want FindingUnresolvedReference", got[0].Detail.Value)
	}
	if detail.Slug != "adr-999" {
		t.Errorf("slug = %q, want adr-999", detail.Slug)
	}
	origin, ok := detail.Origin.AsFrontmatterProperty()
	if !ok {
		t.Fatalf("origin = %+v, want frontmatter-property", detail.Origin.Value)
	}
	if origin.Property != "superseded_by" {
		t.Errorf("property = %q, want superseded_by", origin.Property)
	}
	if !strings.Contains(got[0].Message, "superseded_by") {
		t.Errorf("message = %q, want it to name the property", got[0].Message)
	}
}

// A slug matching a file the index recorded as a non-member reports with
// that reason rather than as matching nothing.
//
// AC: AC-SPECREF-002.7
func TestCheckReportsANonMemberReferenceWithItsReason(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "See [[loose]]."))
	write(t, repo, "specs/notes/loose.md", "# No frontmatter at all\n")

	built := buildChecked(t, repo, referenceCheckProfile("superseded_by"))

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	got := findingsWithRule(findings, finding.RuleUnresolvedReference, "specs/adrs/adr-001.md")
	if len(got) != 1 {
		t.Fatalf("got %d reference/unresolved findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingUnresolvedReference)
	if !ok {
		t.Fatalf("detail = %T, want FindingUnresolvedReference", got[0].Detail.Value)
	}
	if detail.NonMember == nil {
		t.Fatal("non_member is absent, want the excluded file")
	}
	if detail.NonMember.Path != "specs/notes/loose.md" {
		t.Errorf("non-member path = %q", detail.NonMember.Path)
	}
	if !strings.Contains(got[0].Message, "not a corpus member") {
		t.Errorf("message = %q, want it to say the target is not a corpus member", got[0].Message)
	}
}

// CheckFor reports the named document's own findings and nothing located
// in another document — not another file's unresolved reference, and not
// another file's slug-pattern violation (AC-SPECREF-002.9). A file's own
// unresolved reference is still reported when it is the file named.
//
// AC: AC-SPECREF-002.9
func TestCheckForReportsNoFindingLocatedInAnotherDocument(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Refines [[adr-404]]."))
	write(t, repo, "specs/adrs/adr-abc.md", document("adr", "adr-abc", "Unpatterned"))
	write(t, repo, "specs/adrs/adr-002.md", document("adr", "adr-002", "Second"))

	built := buildChecked(t, repo, referenceCheckProfile("superseded_by"))

	findings, err := built.CheckFor("specs/adrs/adr-002.md")
	if err != nil {
		t.Fatalf("CheckFor: %v", err)
	}
	for _, f := range findings {
		if f.Path != "specs/adrs/adr-002.md" {
			t.Errorf("CheckFor(adr-002) reported a finding located in %s: %s %s", f.Path, f.Rule, f.Message)
		}
	}

	findings, err = built.CheckFor("specs/adrs/adr-001.md")
	if err != nil {
		t.Fatalf("CheckFor: %v", err)
	}
	if got := findingsWithRule(findings, finding.RuleUnresolvedReference, "specs/adrs/adr-001.md"); len(got) != 1 {
		t.Errorf("CheckFor(adr-001) reported %d of its own unresolved references, want 1: %+v", len(got), findings)
	}
}

// AC-SPECREF-002.11's own scenario, end to end and through the real cache:
// an agent renames a document's slug, and the write-time check on that one
// file — given the slug it declared at the comparison base — reports the
// reference the rename broke elsewhere, located in the referrer at its own
// span, exactly as a check of the referrer reports it. Without a former
// slug (a new document, or no base to read) it reports nothing elsewhere
// (AC-SPECREF-002.9).
//
// The break is produced by editing a real file on disk, because that is
// what the criterion describes and it is the only state production can
// reach. The edit changes the corpus's input hash, so the check that
// follows reads a freshly rebuilt graph in which the stale referrers are
// recorded unresolved.
//
// AC: AC-SPECREF-002.9, AC-SPECREF-002.11
func TestCheckForReportsWhatARenamedSlugBrokeElsewhere(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Refines [[adr-002]]."))
	write(t, repo, "specs/adrs/adr-002.md", document("adr", "adr-002", "Second"))

	opts := checkingOptions(repo)
	dest := buildAndWriteWith(t, opts)

	// Before the rename the reference resolves, so the check on adr-002 is
	// silent about adr-001 — establishing that what follows is caused by
	// the edit and not merely present all along.
	opened, warm, err := corpus.Open(opts, dest)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !warm {
		t.Fatal("Open rebuilt an index whose inputs are unchanged")
	}
	findings, err := opened.CheckFor("specs/adrs/adr-002.md")
	if err != nil {
		t.Fatalf("CheckFor: %v", err)
	}
	if got := findingsWithRule(findings, finding.RuleUnresolvedReference, "specs/adrs/adr-001.md"); len(got) != 0 {
		t.Fatalf("adr-001's reference was reported broken before the rename: %+v", got)
	}

	// The rename. adr-003 still matches the profile's slug pattern, so the
	// only defect this introduces is the broken inbound reference.
	write(t, repo, "specs/adrs/adr-002.md", document("adr", "adr-003", "Second"))

	opened, _, err = corpus.Open(opts, dest)
	if err != nil {
		t.Fatalf("Open after the rename: %v", err)
	}

	findings, err = opened.CheckFor("specs/adrs/adr-002.md")
	if err != nil {
		t.Fatalf("CheckFor after the rename: %v", err)
	}
	for _, f := range findings {
		if f.Path != "specs/adrs/adr-002.md" {
			t.Errorf("CheckFor(adr-002) with no former slug reported a finding located in %s: %s %s", f.Path, f.Rule, f.Message)
		}
	}

	// Given the slug it declared at the base, in another letter case, the
	// check reports the referrer's finding exactly as the referrer's own
	// check does.
	withFormer, err := opened.CheckFor("specs/adrs/adr-002.md", "ADR-002")
	if err != nil {
		t.Fatalf("CheckFor with a former slug: %v", err)
	}
	ownCheck, err := opened.CheckFor("specs/adrs/adr-001.md")
	if err != nil {
		t.Fatalf("CheckFor(adr-001): %v", err)
	}
	var elsewhere []specmodel.Finding
	for _, f := range withFormer {
		if f.Path != "specs/adrs/adr-002.md" {
			elsewhere = append(elsewhere, f)
		}
	}
	if want := findingsWithRule(ownCheck, finding.RuleUnresolvedReference, "specs/adrs/adr-001.md"); !reflect.DeepEqual(elsewhere, want) || len(want) != 1 {
		t.Errorf("CheckFor(adr-002, former adr-002) reported elsewhere %+v, want exactly %+v", elsewhere, want)
	}

	// A former slug the document still declares broke nothing.
	unchanged, err := opened.CheckFor("specs/adrs/adr-002.md", "adr-003")
	if err != nil {
		t.Fatalf("CheckFor with its current slug as former: %v", err)
	}
	for _, f := range unchanged {
		if f.Path != "specs/adrs/adr-002.md" {
			t.Errorf("a former slug equal to the current one reported %s: %s", f.Path, f.Rule)
		}
	}

	findings, err = opened.CheckFor("specs/adrs/adr-001.md")
	if err != nil {
		t.Fatalf("CheckFor(adr-001) after the rename: %v", err)
	}
	broken := findingsWithRule(findings, finding.RuleUnresolvedReference, "specs/adrs/adr-001.md")
	if len(broken) != 1 {
		t.Fatalf("got %d reference/unresolved findings in the referrer, want 1: %+v", len(broken), findings)
	}
	detail, ok := broken[0].Detail.Value.(specmodel.FindingUnresolvedReference)
	if !ok {
		t.Fatalf("detail = %T, want FindingUnresolvedReference", broken[0].Detail.Value)
	}
	if detail.Slug != "adr-002" {
		t.Errorf("slug = %q, want adr-002 — the slug the referrer still names", detail.Slug)
	}
	outbound := opened.OutboundReferences("specs/adrs/adr-001.md")
	if len(outbound) != 1 {
		t.Fatalf("adr-001 holds %d references, want 1", len(outbound))
	}
	if broken[0].Span != outbound[0].Span {
		t.Errorf("position = %+v, want the referrer's own span %+v", broken[0].Span, outbound[0].Span)
	}
}

// The slug-less anchor form — a reference to a section of the document
// holding it — resolves its anchor against that document and carries an
// empty slug, which matches no corpus member. Reporting the slug half
// would call the same reference broken in the run that reports its
// anchor resolved, so the unresolved-reference rule skips it.
//
// AC: AC-SPECREF-002.12
func TestAnAnchorOnlyReferenceReachingItsOwnHeadingIsNotUnresolved(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md",
		citing("adr", "adr-001", "First", "See [[#Rationale]].\n\n## Rationale\n\nProse beneath it.\n"))

	findings := checkedAnchors(t, repo)

	if got := findingsWithRule(findings, finding.RuleUnresolvedReference, "specs/adrs/adr-001.md"); len(got) != 0 {
		t.Errorf("got %d reference/unresolved findings, want 0: %+v", len(got), got)
	}
	if got := anchorFindingsOf(findings); len(got) != 0 {
		t.Errorf("got %d anchor findings, want 0: %+v", len(got), got)
	}
}

// The exemption is conjunctive: an empty slug alone does not earn it.
// An anchor-only reference naming no heading of the document holding it
// reaches nothing, so it stays an unresolved reference.
//
// AC: AC-SPECREF-002.12
func TestAnAnchorOnlyReferenceReachingNoHeadingOfItsOwnDocumentIsStillUnresolved(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md",
		citing("adr", "adr-001", "First", "See [[#Rationale]].\n\n## Decision\n\nProse beneath it.\n"))

	findings := checkedAnchors(t, repo)

	got := findingsWithRule(findings, finding.RuleUnresolvedReference, "specs/adrs/adr-001.md")
	if len(got) != 1 {
		t.Fatalf("got %d reference/unresolved findings, want 1: %+v", len(got), findings)
	}
	if got[0].Severity != finding.SeverityError {
		t.Errorf("severity = %q, want error", got[0].Severity)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingUnresolvedReference)
	if !ok {
		t.Fatalf("detail = %T, want FindingUnresolvedReference", got[0].Detail.Value)
	}
	if detail.Slug != "" {
		t.Errorf("slug = %q, want it empty — the reference names none", detail.Slug)
	}
}

// A reference carrying a slug that resolves nowhere is not the exempt
// form, whatever its anchor did: the exemption turns on the reference
// naming no document, not on it carrying an anchor.
//
// AC: AC-SPECREF-002.12
func TestASlugBearingAnchoredReferenceIsUnaffectedByTheSelfAnchorExemption(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md",
		citing("adr", "adr-001", "First", "See [[adr-404#Rationale]].\n\n## Rationale\n\nProse beneath it.\n"))

	findings := checkedAnchors(t, repo)

	got := findingsWithRule(findings, finding.RuleUnresolvedReference, "specs/adrs/adr-001.md")
	if len(got) != 1 {
		t.Fatalf("got %d reference/unresolved findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingUnresolvedReference)
	if !ok {
		t.Fatalf("detail = %T, want FindingUnresolvedReference", got[0].Detail.Value)
	}
	if detail.Slug != "adr-404" {
		t.Errorf("slug = %q, want adr-404", detail.Slug)
	}
}

// Tests for reference/retired-target, reference/pairing-incomplete and
// reference/type-mismatch — the LIFE delta's relational half of
// reference integrity. Every registration these tests exercise is a
// stubProfile of this test file's own, registered through the same
// [corpus.Profile] extension point the command surface's real profile
// implements — never a pairing added to that shipped profile, which
// registers none by design (reference-integrity-blueprint.md's LIFE
// section).

// statusDocument renders a conforming document at the given status,
// carrying any extra frontmatter lines verbatim — the control document(),
// documentWithExtra() and citing() do not offer, and every LIFE fixture
// below needs at least one of a non-default status, a successor value, or
// a registered pairing's counterpart.
func statusDocument(docType, slug, title, status string, extra ...string) string {
	lines := []string{
		"---",
		fmt.Sprintf("title: %q", title),
		`description: "A fixture document."`,
		fmt.Sprintf("slug: %s", slug),
		fmt.Sprintf("type: %s", docType),
		fmt.Sprintf("status: %s", status),
	}
	lines = append(lines, extra...)
	lines = append(lines, "---", "", fmt.Sprintf("# %s", title), "")
	return strings.Join(lines, "\n")
}

// citingAtStatus renders a conforming document at the given status whose
// body is the given text — the status control statusDocument() offers,
// combined with the body control citing() offers, for the one fixture
// that needs both: a retired document that itself cites another one.
func citingAtStatus(docType, slug, title, status, body string) string {
	return fmt.Sprintf(`---
title: %q
description: "A fixture document."
slug: %s
type: %s
status: %s
---

# %s

%s
`, title, slug, docType, status, title, body)
}

// lifecycleCheckProfile registers "adr" and "frd" exactly as
// referenceCheckProfile does, widened with a lifecycle registration:
// draft and approved are live, retired is retired, and superseded_by is
// the successor property — this framework's own registration, restated
// here as test data rather than imported, so this package proves its
// checks against a fixture it owns (spec-profile-module.md's Change
// Guidance: Document Core imports no profile).
func lifecycleCheckProfile() corpus.Profile {
	return stubProfile{
		schema: []byte(permissiveCheckSchema),
		types: map[string]corpus.DocumentTypeRules{
			"adr": {SlugPattern: `^adr-[0-9]{3}(-[a-z0-9-]+)*$`},
			"frd": {SlugPattern: `^frd-[a-z0-9-]+$`},
		},
		lifecycle: corpus.LifecycleRules{
			Statuses:          []string{"draft", "approved", "retired"},
			LiveStatuses:      []string{"draft", "approved"},
			RetiredStatuses:   []string{"retired"},
			SuccessorProperty: "superseded_by",
		},
	}
}

// AC: AC-SPECREF-002.6
func TestCheckReportsALiveDocumentReferencingARetiredOneWithASuccessor(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-900.md",
		statusDocument("adr", "adr-900", "Old decision", "retired", "superseded_by: adr-901"))
	write(t, repo, "specs/adrs/adr-901.md", document("adr", "adr-901", "New decision"))
	write(t, repo, "specs/adrs/adr-800.md",
		citing("adr", "adr-800", "Citing decision", "This decision refines [[adr-900]]."))

	built := buildChecked(t, repo, lifecycleCheckProfile())
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	got := findingsWithRule(findings, finding.RuleRetiredTarget, "specs/adrs/adr-800.md")
	if len(got) != 1 {
		t.Fatalf("got %d reference/retired-target findings, want 1: %+v", len(got), findings)
	}
	if got[0].Severity != finding.SeverityWarning {
		t.Errorf("severity = %q, want %q — a reference to a retired document still resolves", got[0].Severity, finding.SeverityWarning)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingRetiredTarget)
	if !ok {
		t.Fatalf("detail = %T, want FindingRetiredTarget", got[0].Detail.Value)
	}
	if detail.Slug != "adr-900" {
		t.Errorf("slug = %q, want adr-900", detail.Slug)
	}
	if detail.Successor == nil || *detail.Successor != "adr-901" {
		t.Errorf("successor = %v, want adr-901", detail.Successor)
	}
	if !strings.Contains(got[0].Message, "adr-900") || !strings.Contains(got[0].Message, "adr-901") {
		t.Errorf("message = %q, want it to name both the retired target and its successor", got[0].Message)
	}
}

// AC: AC-SPECREF-002.6
//
// Not every retirement has a successor (reference-integrity-blueprint.md,
// LIFE): a retired target that declares none still raises the warning,
// with an absent successor rather than one invented from nothing.
func TestCheckReportsALiveDocumentReferencingARetiredOneWithNoSuccessor(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-900.md", statusDocument("adr", "adr-900", "Old decision", "retired"))
	write(t, repo, "specs/adrs/adr-800.md",
		citing("adr", "adr-800", "Citing decision", "This decision refines [[adr-900]]."))

	built := buildChecked(t, repo, lifecycleCheckProfile())
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	got := findingsWithRule(findings, finding.RuleRetiredTarget, "specs/adrs/adr-800.md")
	if len(got) != 1 {
		t.Fatalf("got %d reference/retired-target findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingRetiredTarget)
	if !ok {
		t.Fatalf("detail = %T, want FindingRetiredTarget", got[0].Detail.Value)
	}
	if detail.Successor != nil {
		t.Errorf("successor = %v, want nil — the target declares none", detail.Successor)
	}
}

// A reference between two live documents raises no retired-target
// finding — the check the tests above would pass vacuously without.
func TestCheckRaisesNoRetiredTargetFindingBetweenTwoLiveDocuments(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-900.md", document("adr", "adr-900", "Live decision"))
	write(t, repo, "specs/adrs/adr-800.md",
		citing("adr", "adr-800", "Citing decision", "This decision refines [[adr-900]]."))

	built := buildChecked(t, repo, lifecycleCheckProfile())
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if got := findingsWithRule(findings, finding.RuleRetiredTarget, "specs/adrs/adr-800.md"); len(got) != 0 {
		t.Errorf("got %d reference/retired-target findings between two live documents, want 0: %+v", len(got), findings)
	}
}

// A retired document citing another retired one raises no finding: the
// warning exists to tell a LIVE author what to cite instead, and the
// referrer here is not one (reference-integrity-blueprint.md's LIFE
// rationale for why every non-retired status is live).
func TestCheckRaisesNoRetiredTargetFindingWhenTheReferringDocumentIsAlsoRetired(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-900.md", statusDocument("adr", "adr-900", "Old decision", "retired"))
	write(t, repo, "specs/adrs/adr-850.md",
		citingAtStatus("adr", "adr-850", "Also retired", "retired", "This decision refines [[adr-900]]."))

	built := buildChecked(t, repo, lifecycleCheckProfile())
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if got := findingsWithRule(findings, finding.RuleRetiredTarget, "specs/adrs/adr-850.md"); len(got) != 0 {
		t.Errorf("got %d reference/retired-target findings from a retired referrer, want 0: %+v", len(got), findings)
	}
}

// pairingCheckProfile registers "owner"/"owned_by" as a pairing on "adr"
// — a relationship this framework's own shipped profile deliberately
// registers none of (LIFE's Design: "Do not add a pairing to the shipped
// profile"). "owner" is also registered as reference-bearing, which a
// pairing's Property must be per [corpus.ValidateRegistration].
func pairingCheckProfile() corpus.Profile {
	return stubProfile{
		schema: []byte(permissiveCheckSchema),
		types: map[string]corpus.DocumentTypeRules{
			"adr": {
				SlugPattern:         `^adr-[0-9]{3}(-[a-z0-9-]+)*$`,
				ReferenceProperties: []string{"owner"},
				Pairings: []specmodel.PropertyPairing{
					{Property: "owner", Counterpart: "owned_by"},
				},
			},
			"frd": {SlugPattern: `^frd-[a-z0-9-]+$`},
		},
	}
}

// AC: AC-SPECREF-002.5
func TestCheckReportsAnIncompletePairingAgainstBothDocuments(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-100.md",
		statusDocument("adr", "adr-100", "Owning decision", "approved", "owner: frd-team-ledger"))
	write(t, repo, "specs/features/frd-team-ledger.md", document("frd", "frd-team-ledger", "Team Ledger"))

	built := buildChecked(t, repo, pairingCheckProfile())
	targetDoc, ok := built.Document("specs/features/frd-team-ledger.md")
	if !ok {
		t.Fatal("Document did not return the target's parsed frontmatter")
	}

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	source := findingsWithRule(findings, finding.RuleIncompletePairing, "specs/adrs/adr-100.md")
	if len(source) != 1 {
		t.Fatalf("got %d reference/pairing-incomplete findings against the declaring document, want 1: %+v", len(source), findings)
	}
	target := findingsWithRule(findings, finding.RuleIncompletePairing, "specs/features/frd-team-ledger.md")
	if len(target) != 1 {
		t.Fatalf("got %d reference/pairing-incomplete findings against the target, want 1: %+v", len(target), findings)
	}

	sourceDetail, ok := source[0].Detail.Value.(specmodel.FindingIncompletePairing)
	if !ok {
		t.Fatalf("source detail = %T, want FindingIncompletePairing", source[0].Detail.Value)
	}
	targetDetail, ok := target[0].Detail.Value.(specmodel.FindingIncompletePairing)
	if !ok {
		t.Fatalf("target detail = %T, want FindingIncompletePairing", target[0].Detail.Value)
	}
	if sourceDetail != targetDetail {
		t.Errorf("source detail = %+v, target detail = %+v, want the same structured detail against both documents", sourceDetail, targetDetail)
	}
	if sourceDetail.Property != "owner" || sourceDetail.Counterpart != "owned_by" {
		t.Errorf("detail = %+v, want property owner paired with owned_by", sourceDetail)
	}
	if sourceDetail.SourcePath != "specs/adrs/adr-100.md" || sourceDetail.TargetPath != "specs/features/frd-team-ledger.md" {
		t.Errorf("detail = %+v, want both paths", sourceDetail)
	}

	// The declaring document's finding sits at the declared value's span;
	// the target's sits at its frontmatter block's span, because an
	// absent property has no position of its own.
	if source[0].Span.Length == 0 {
		t.Error("the declaring document's finding carries an empty span")
	}
	if target[0].Span != targetDoc.Region.Inner {
		t.Errorf("target position = %+v, want the target's frontmatter-block span %+v", target[0].Span, targetDoc.Region.Inner)
	}
}

// A pairing whose target declares the counterpart back raises no
// finding — the check the test above would pass vacuously without.
//
// AC: AC-SPECREF-002.5
func TestCheckRaisesNoIncompletePairingFindingWhenTheCounterpartIsDeclared(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-100.md",
		statusDocument("adr", "adr-100", "Owning decision", "approved", "owner: frd-team-ledger"))
	write(t, repo, "specs/features/frd-team-ledger.md",
		statusDocument("frd", "frd-team-ledger", "Team Ledger", "approved", "owned_by: adr-100"))

	built := buildChecked(t, repo, pairingCheckProfile())
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	for _, f := range findings {
		if f.Rule == finding.RuleIncompletePairing {
			t.Errorf("got a reference/pairing-incomplete finding when the counterpart is declared: %+v", f)
		}
	}
}

// typeMismatchCheckProfile registers this framework's own typed
// reference — a feature blueprint's frd must resolve to a document of
// type frd — as test data, on the same terms lifecycleCheckProfile
// restates the shipped lifecycle registration.
func typeMismatchCheckProfile() corpus.Profile {
	return stubProfile{
		schema: []byte(`{
  "definitions": {
    "FeatureBlueprint": {"type": "object", "properties": {}},
    "Frd": {"type": "object", "properties": {}},
    "Adr": {"type": "object", "properties": {}}
  }
}`),
		types: map[string]corpus.DocumentTypeRules{
			"feature-blueprint": {
				SlugPattern:              `^feature-[a-z0-9-]+$`,
				ReferenceProperties:      []string{"frd"},
				TypedReferenceProperties: map[string]string{"frd": "frd"},
			},
			"frd": {SlugPattern: `^frd-[a-z0-9-]+$`},
			"adr": {SlugPattern: `^adr-[0-9]{3}$`},
		},
	}
}

// AC: AC-SPECREF-002.10
func TestCheckReportsATypeMismatch(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/feature-checkout.md",
		statusDocument("feature-blueprint", "feature-checkout", "Checkout", "approved", "frd: adr-009"))
	write(t, repo, "specs/adrs/adr-009.md", document("adr", "adr-009", "Checkout decision"))

	built := buildChecked(t, repo, typeMismatchCheckProfile())
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	got := findingsWithRule(findings, finding.RuleTypeMismatch, "specs/features/feature-checkout.md")
	if len(got) != 1 {
		t.Fatalf("got %d reference/type-mismatch findings, want 1: %+v", len(got), findings)
	}
	detail, ok := got[0].Detail.Value.(specmodel.FindingReferenceType)
	if !ok {
		t.Fatalf("detail = %T, want FindingReferenceType", got[0].Detail.Value)
	}
	if detail.Property != "frd" {
		t.Errorf("property = %q, want frd", detail.Property)
	}
	if detail.Slug != "adr-009" {
		t.Errorf("slug = %q, want adr-009", detail.Slug)
	}
	if detail.ExpectedType != "frd" || detail.ActualType != "adr" {
		t.Errorf("expected/actual = %q/%q, want frd/adr", detail.ExpectedType, detail.ActualType)
	}
	if !strings.Contains(got[0].Message, "frd") || !strings.Contains(got[0].Message, "adr-009") || !strings.Contains(got[0].Message, "adr") {
		t.Errorf("message = %q, want it to name the property, the target and both types", got[0].Message)
	}
}

// A typed reference whose target declares the required type raises no
// finding — the check the test above would pass vacuously without.
//
// AC: AC-SPECREF-002.10
func TestCheckRaisesNoTypeMismatchFindingWhenTheTargetDeclaresTheExpectedType(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/features/feature-checkout.md",
		statusDocument("feature-blueprint", "feature-checkout", "Checkout", "approved", "frd: frd-checkout"))
	write(t, repo, "specs/features/frd-checkout.md", document("frd", "frd-checkout", "Checkout"))

	built := buildChecked(t, repo, typeMismatchCheckProfile())
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if got := findingsWithRule(findings, finding.RuleTypeMismatch, "specs/features/feature-checkout.md"); len(got) != 0 {
		t.Errorf("got %d reference/type-mismatch findings when the target declares the expected type, want 0: %+v", len(got), findings)
	}
}

// A check of the document just retired reports no retired-target finding
// located in a document citing it; a check of the citing document reports
// it (AC-SPECREF-002.9, AC-SPECREF-002.11).
//
// AC: AC-SPECREF-002.11
func TestCheckForReportsLifecycleFindingsOnlyInTheReferringDocument(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-900.md", statusDocument("adr", "adr-900", "Old decision", "retired"))
	write(t, repo, "specs/adrs/adr-800.md",
		citing("adr", "adr-800", "Citing decision", "This decision refines [[adr-900]]."))

	built := buildChecked(t, repo, lifecycleCheckProfile())

	findings, err := built.CheckFor("specs/adrs/adr-900.md")
	if err != nil {
		t.Fatalf("CheckFor: %v", err)
	}
	if got := findingsWithRule(findings, finding.RuleRetiredTarget, "specs/adrs/adr-800.md"); len(got) != 0 {
		t.Errorf("CheckFor(adr-900) reported %d retired-target findings located in the referrer, want 0: %+v", len(got), findings)
	}

	findings, err = built.CheckFor("specs/adrs/adr-800.md")
	if err != nil {
		t.Fatalf("CheckFor: %v", err)
	}
	if got := findingsWithRule(findings, finding.RuleRetiredTarget, "specs/adrs/adr-800.md"); len(got) != 1 {
		t.Errorf("CheckFor(adr-800) reported %d of its own retired-target findings, want 1: %+v", len(got), findings)
	}
}

// A single-document check reports every reference finding located in the
// named document exactly as a corpus check reports it for that document:
// an unresolved reference, a reference to a retired target and a
// type-mismatched reference (AC-SPECREF-002.11).
//
// AC: AC-SPECREF-002.11
func TestCheckForReportsTheNamedDocumentsReferenceFindingsAsACorpusCheckDoes(t *testing.T) {
	cases := []struct {
		name    string
		profile corpus.Profile
		files   map[string]string
		named   string
		rule    specmodel.RuleId
	}{
		{
			name:    "unresolved",
			profile: referenceCheckProfile("superseded_by"),
			files: map[string]string{
				"specs/adrs/adr-001.md": citing("adr", "adr-001", "First", "Refines [[adr-404]]."),
			},
			named: "specs/adrs/adr-001.md",
			rule:  finding.RuleUnresolvedReference,
		},
		{
			name:    "retired target",
			profile: lifecycleCheckProfile(),
			files: map[string]string{
				"specs/adrs/adr-900.md": statusDocument("adr", "adr-900", "Old decision", "retired"),
				"specs/adrs/adr-800.md": citing("adr", "adr-800", "Citing", "Refines [[adr-900]]."),
			},
			named: "specs/adrs/adr-800.md",
			rule:  finding.RuleRetiredTarget,
		},
		{
			name:    "type mismatch",
			profile: typeMismatchCheckProfile(),
			files: map[string]string{
				"specs/features/feature-checkout.md": statusDocument("feature-blueprint", "feature-checkout", "Checkout", "approved", "frd: adr-009"),
				"specs/adrs/adr-009.md":              document("adr", "adr-009", "Checkout decision"),
			},
			named: "specs/features/feature-checkout.md",
			rule:  finding.RuleTypeMismatch,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := t.TempDir()
			for path, content := range tc.files {
				write(t, repo, path, content)
			}
			built := buildChecked(t, repo, tc.profile)

			all, err := built.Check()
			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			narrowed, err := built.CheckFor(tc.named)
			if err != nil {
				t.Fatalf("CheckFor: %v", err)
			}

			want := findingsWithRule(all, tc.rule, tc.named)
			got := findingsWithRule(narrowed, tc.rule, tc.named)
			if len(want) != 1 {
				t.Fatalf("the corpus check reported %d %s findings in %s, want 1: %+v", len(want), tc.rule, tc.named, all)
			}
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Errorf("CheckFor(%s) reported %+v, want what the corpus check reports: %+v", tc.named, got, want)
			}
		})
	}
}

// A path that names no corpus member is an error, never a report with no
// findings — a vacuous success. The index
// already distinguishes a candidate it excluded, with the reason, from a
// path discovery never saw, and both are reported as such.
//
// AC: AC-SPEC-001.2
func TestCheckForRejectsAPathThatIsNotACorpusMember(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", document("adr", "adr-001", "First"))
	write(t, repo, "specs/notes/loose.md", "# No frontmatter at all\n")
	write(t, repo, "README.md", "# Outside the spec root\n")

	built := buildChecked(t, repo, referenceCheckProfile("superseded_by"))

	for _, tc := range []struct {
		name string
		path string
		want string
	}{
		{"never discovered", "specs/adrs/does-not-exist.md", "no document at that path"},
		{"outside the spec root", "README.md", "no document at that path"},
		{"excluded candidate", "specs/notes/loose.md", "no-frontmatter"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			findings, err := built.CheckFor(tc.path)
			if err == nil {
				t.Fatalf("CheckFor(%s) returned %d findings and no error", tc.path, len(findings))
			}
			if !errors.Is(err, corpus.ErrNotAMember) {
				t.Fatalf("error = %v, want it to wrap ErrNotAMember", err)
			}
			if findings != nil {
				t.Errorf("findings = %+v, want none alongside the error", findings)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to mention %q", err, tc.want)
			}
			if !strings.Contains(err.Error(), tc.path) {
				t.Errorf("error = %q, want it to name the path asked about", err)
			}
		})
	}

	// A real member is unaffected.
	if _, err := built.CheckFor("specs/adrs/adr-001.md"); err != nil {
		t.Errorf("CheckFor on a member returned %v", err)
	}
}

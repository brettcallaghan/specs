// Tests for the Findings component.
//
// Two obligations are exercised here that no other package can hold. The
// self-contained-diagnostic guarantee (AC-SPEC-006.1) is a property of
// every finding this package can build, so it is asserted against every
// constructor rather than against the one check that happens to raise it
// first. And the unverified mechanism (AC-SPEC-006.3, AC-SPEC-006.4) is
// registered by no check this sprint, so if it is not tested here it is
// not tested at all.
package finding_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// schemaPath is the generated bundle scripts/generate-schemas.sh emits
// from schemas/*.strata. The schema-shape tests below read it rather than
// restating its contents, so a change to the Strata source that loosened
// the model would fail here instead of passing unnoticed.
const schemaPath = "../schema/spec-model.schema.json"

// at is a stand-in position. Findings carry a span computed by the Spans
// component; this package neither derives nor validates one, so a fixed
// document and extent serve every case.
func at(t *testing.T) specmodel.Span {
	t.Helper()
	return span.Of([]byte("---\nslug: adr-011-targeted-verification\n---\n"), 10, 30)
}

// Every constructor must produce a finding that names the rule, the
// severity, the file, the position and the constraint that was violated,
// and must carry the same facts again as structured fields — so a
// consumer acts on the finding without re-reading the document or the
// standard that defines the rule.
//
// AC: AC-SPEC-006.1
func TestEveryFindingIsSelfContained(t *testing.T) {
	position := at(t)

	cases := []struct {
		name     string
		finding  specmodel.Finding
		rule     specmodel.RuleId
		contains []string
		detail   func(t *testing.T, d specmodel.FindingDetail)
	}{
		{
			name:     "schema violation",
			finding:  finding.SchemaViolation("specs/adrs/adr-011.md", position, "/updated_at", "format: date-time", "2026-07-29"),
			rule:     finding.RuleSchemaViolation,
			contains: []string{"/updated_at", "2026-07-29", "format: date-time"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsSchemaViolation()
				if !ok {
					t.Fatalf("detail is not a schema violation: %#v", d.Value)
				}
				if v.Pointer != "/updated_at" || v.Constraint != "format: date-time" || v.Value != "2026-07-29" {
					t.Errorf("detail = %#v, want the pointer, constraint and raw value", v)
				}
			},
		},
		{
			name:     "missing property",
			finding:  finding.MissingProperty("specs/adrs/adr-011.md", position, "", "description"),
			rule:     finding.RuleMissingProperty,
			contains: []string{"description"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsMissingProperty()
				if !ok {
					t.Fatalf("detail is not a missing property: %#v", d.Value)
				}
				if v.Property != "description" {
					t.Errorf("Property = %q, want %q", v.Property, "description")
				}
			},
		},
		{
			name:     "slug pattern",
			finding:  finding.SlugPattern("specs/adrs/adr-011.md", position, "adr-011-targeted-verification", `^adr-[0-9]{3}$`),
			rule:     finding.RuleSlugPattern,
			contains: []string{"adr-011-targeted-verification", `^adr-[0-9]{3}$`},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsSlugPattern()
				if !ok {
					t.Fatalf("detail is not a slug pattern: %#v", d.Value)
				}
				if v.Slug != "adr-011-targeted-verification" || v.Pattern != `^adr-[0-9]{3}$` {
					t.Errorf("detail = %#v, want both the slug and the pattern", v)
				}
			},
		},
		{
			name:     "duplicate slug",
			finding:  finding.DuplicateSlug("specs/adrs/adr-011.md", position, "adr-011", "specs/adrs/other.md"),
			rule:     finding.RuleDuplicateSlug,
			contains: []string{"adr-011", "specs/adrs/other.md"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsDuplicateSlug()
				if !ok {
					t.Fatalf("detail is not a duplicate slug: %#v", d.Value)
				}
				if v.Slug != "adr-011" || v.OtherPath != "specs/adrs/other.md" {
					t.Errorf("detail = %#v, want the slug and the other document's path", v)
				}
			},
		},
		{
			name:     "unverified",
			finding:  finding.Unverified("structure/mermaid", "specs/adrs/adr-011.md", position, "no diagram validator is available"),
			rule:     "structure/mermaid",
			contains: []string{"no diagram validator is available"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsUnverified()
				if !ok {
					t.Fatalf("detail is not an unverified check: %#v", d.Value)
				}
				if v.Reason != "no diagram validator is available" {
					t.Errorf("Reason = %q, want the reason the check could not run", v.Reason)
				}
			},
		},
		{
			name: "unresolved reference",
			finding: finding.UnresolvedReference("specs/adrs/adr-011.md", position, "adr-404", "adr-404",
				specmodel.NewReferenceOriginBody(specmodel.ReferenceBodyOrigin{}), nil),
			rule:     finding.RuleUnresolvedReference,
			contains: []string{"adr-404"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsUnresolvedReference()
				if !ok {
					t.Fatalf("detail is not an unresolved reference: %#v", d.Value)
				}
				if v.Slug != "adr-404" {
					t.Errorf("Slug = %q, want %q", v.Slug, "adr-404")
				}
				if _, ok := v.Origin.AsBody(); !ok {
					t.Errorf("Origin = %#v, want body", v.Origin.Value)
				}
				if v.NonMember != nil {
					t.Errorf("NonMember = %#v, want none", v.NonMember)
				}
			},
		},
		{
			name:     "unresolved anchor",
			finding:  finding.UnresolvedAnchor("specs/adrs/adr-011.md", position, "Componets", "container-spec-cli"),
			rule:     finding.RuleUnresolvedAnchor,
			contains: []string{"Componets", "container-spec-cli"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsUnresolvedAnchor()
				if !ok {
					t.Fatalf("detail is not an unresolved anchor: %#v", d.Value)
				}
				if v.Anchor != "Componets" || v.TargetPath != "container-spec-cli" {
					t.Errorf("detail = %#v, want the anchor and the target document", v)
				}
			},
		},
		{
			name: "ambiguous anchor",
			finding: finding.AmbiguousAnchor("specs/adrs/adr-011.md", position, "References", "module-document-core.md",
				[]specmodel.AnchorMatch{
					{Text: "References", Position: position},
					{Text: "References", Position: position},
				}),
			rule:     finding.RuleAmbiguousAnchor,
			contains: []string{"References", "module-document-core.md", "2"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsAmbiguousAnchor()
				if !ok {
					t.Fatalf("detail is not an ambiguous anchor: %#v", d.Value)
				}
				if v.Anchor != "References" || v.TargetPath != "module-document-core.md" {
					t.Errorf("detail = %#v, want the anchor and the target document", v)
				}
				if len(v.Matches) != 2 {
					t.Errorf("Matches = %#v, want every matching heading", v.Matches)
				}
			},
		},
		{
			name: "retired target",
			finding: finding.RetiredTarget("specs/adrs/adr-011.md", position, "adr-009",
				strPtr("adr-013")),
			rule:     finding.RuleRetiredTarget,
			contains: []string{"adr-009", "adr-013"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsRetiredTarget()
				if !ok {
					t.Fatalf("detail is not a retired target: %#v", d.Value)
				}
				if v.Slug != "adr-009" {
					t.Errorf("Slug = %q, want %q", v.Slug, "adr-009")
				}
				if v.Successor == nil || *v.Successor != "adr-013" {
					t.Errorf("Successor = %v, want %q", v.Successor, "adr-013")
				}
			},
		},
		{
			name: "incomplete pairing",
			finding: finding.IncompletePairing("specs/adrs/adr-011.md", position,
				"owner", "owned_by", "specs/adrs/adr-011.md", "specs/teams/team-ledger.md"),
			rule:     finding.RuleIncompletePairing,
			contains: []string{"owner", "owned_by", "specs/teams/team-ledger.md"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsIncompletePairing()
				if !ok {
					t.Fatalf("detail is not an incomplete pairing: %#v", d.Value)
				}
				if v.Property != "owner" || v.Counterpart != "owned_by" {
					t.Errorf("detail = %#v, want the property and its counterpart", v)
				}
				if v.SourcePath != "specs/adrs/adr-011.md" || v.TargetPath != "specs/teams/team-ledger.md" {
					t.Errorf("detail = %#v, want both paths", v)
				}
			},
		},
		{
			name: "reference type",
			finding: finding.ReferenceType("specs/adrs/adr-011.md", position,
				"frd", "adr-009", "frd", "adr"),
			rule:     finding.RuleTypeMismatch,
			contains: []string{"frd", "adr-009", "adr"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsReferenceType()
				if !ok {
					t.Fatalf("detail is not a reference type: %#v", d.Value)
				}
				if v.Property != "frd" || v.Slug != "adr-009" || v.ExpectedType != "frd" || v.ActualType != "adr" {
					t.Errorf("detail = %#v, want the property, slug, and both types", v)
				}
			},
		},
		{
			name:     "unmapped citation",
			finding:  finding.UnmappedCitation("specs/adrs/adr-011.md", position, "document-relationship-standards", "links.yaml"),
			rule:     finding.RuleUnmappedCitation,
			contains: []string{"document-relationship-standards", "links.yaml"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsUnmappedCitation()
				if !ok {
					t.Fatalf("detail is not an unmapped citation: %#v", d.Value)
				}
				if v.Slug != "document-relationship-standards" || v.CitingPath != "specs/adrs/adr-011.md" || v.MapPath != "links.yaml" {
					t.Errorf("detail = %#v, want the slug, the citing path and the map path", v)
				}
			},
		},
		{
			name:     "unresolved citation",
			finding:  finding.UnresolvedCitation("specs/adrs/adr-011.md", position, "adr-404"),
			rule:     finding.RuleUnresolvedCitation,
			contains: []string{"adr-404"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsUnresolvedCitation()
				if !ok {
					t.Fatalf("detail is not an unresolved citation: %#v", d.Value)
				}
				if v.Slug != "adr-404" || v.CitingPath != "specs/adrs/adr-011.md" {
					t.Errorf("detail = %#v, want the slug and the citing path", v)
				}
			},
		},
		{
			name:     "feature untagged",
			finding:  finding.FeatureUntagged("specs/adrs/adr-011.md", position, `^FEAT-[A-Z0-9]+-[0-9]{3}$`),
			rule:     finding.RuleFeatureUntagged,
			contains: []string{`^FEAT-[A-Z0-9]+-[0-9]{3}$`},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsIdentifierPattern()
				if !ok {
					t.Fatalf("detail is not an identifier pattern: %#v", d.Value)
				}
				if v.Identifier != nil {
					t.Errorf("Identifier = %v, want nil: the feature carries no identifier tag at all", v.Identifier)
				}
			},
		},
		{
			name:     "criterion untagged",
			finding:  finding.CriterionUntagged("specs/adrs/adr-011.md", position, `^AC-[A-Z0-9]+-[0-9]{3}$`),
			rule:     finding.RuleCriterionUntagged,
			contains: []string{`^AC-[A-Z0-9]+-[0-9]{3}$`},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsIdentifierPattern()
				if !ok {
					t.Fatalf("detail is not an identifier pattern: %#v", d.Value)
				}
				if v.Identifier != nil {
					t.Errorf("Identifier = %v, want nil", v.Identifier)
				}
			},
		},
		{
			name:     "identifier malformed",
			finding:  finding.IdentifierMalformed("specs/adrs/adr-011.md", position, "AC-chk-3.1", `^AC-[A-Z0-9]+-[0-9]{3}$`),
			rule:     finding.RuleIdentifierMalformed,
			contains: []string{"AC-chk-3.1", `^AC-[A-Z0-9]+-[0-9]{3}$`},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsIdentifierPattern()
				if !ok {
					t.Fatalf("detail is not an identifier pattern: %#v", d.Value)
				}
				if v.Identifier == nil || *v.Identifier != "AC-chk-3.1" {
					t.Errorf("Identifier = %v, want the tag as written", v.Identifier)
				}
			},
		},
		{
			name:     "identifier multiple",
			finding:  finding.IdentifierMultiple("specs/adrs/adr-011.md", position, []string{"AC-CHK-001.1", "AC-CHK-001.2"}, `^AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`),
			rule:     finding.RuleIdentifierMultiple,
			contains: []string{"AC-CHK-001.1", "AC-CHK-001.2"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsIdentifierPattern()
				if !ok {
					t.Fatalf("detail is not an identifier pattern: %#v", d.Value)
				}
				if v.Identifier == nil || !strings.Contains(*v.Identifier, "AC-CHK-001.1") || !strings.Contains(*v.Identifier, "AC-CHK-001.2") {
					t.Errorf("Identifier = %v, want both tags named", v.Identifier)
				}
			},
		},
		{
			name: "identifier duplicate",
			finding: finding.IdentifierDuplicate("specs/adrs/adr-011.md", position,
				"AC-CHK-001.1", "specs/adrs/other.md", position),
			rule:     finding.RuleIdentifierDuplicate,
			contains: []string{"AC-CHK-001.1", "specs/adrs/other.md"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsDuplicateIdentifier()
				if !ok {
					t.Fatalf("detail is not a duplicate identifier: %#v", d.Value)
				}
				if v.Identifier != "AC-CHK-001.1" || v.OtherPath != "specs/adrs/other.md" || v.OtherPosition != position {
					t.Errorf("detail = %#v, want the identifier and the other occurrence's location", v)
				}
			},
		},
		{
			name:     "criterion code mismatch",
			finding:  finding.CriterionCodeMismatch("specs/adrs/adr-011.md", position, "AC-REF-001.1", "FEAT-CHK-001", `^AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`),
			rule:     finding.RuleCriterionCodeMismatch,
			contains: []string{"AC-REF-001.1", "FEAT-CHK-001"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsIdentifierPattern()
				if !ok {
					t.Fatalf("detail is not an identifier pattern: %#v", d.Value)
				}
				if v.Identifier == nil || *v.Identifier != "AC-REF-001.1" {
					t.Errorf("Identifier = %v, want %q", v.Identifier, "AC-REF-001.1")
				}
			},
		},
		{
			name:     "owner missing",
			finding:  finding.OwnerMissing("specs/adrs/adr-011.md", position, "frd-"),
			rule:     finding.RuleOwnerMissing,
			contains: []string{"frd-"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsOwningDocument()
				if !ok {
					t.Fatalf("detail is not an owning document: %#v", d.Value)
				}
				if v.Slug != nil || v.DocumentSlug != nil {
					t.Errorf("detail = %#v, want both halves absent", v)
				}
			},
		},
		{
			name:     "owner mismatch",
			finding:  finding.OwnerMismatch("specs/adrs/adr-011.md", position, "frd-checkout", "frd-refunds"),
			rule:     finding.RuleOwnerMismatch,
			contains: []string{"frd-checkout", "frd-refunds"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsOwningDocument()
				if !ok {
					t.Fatalf("detail is not an owning document: %#v", d.Value)
				}
				if v.Slug == nil || *v.Slug != "frd-checkout" || v.DocumentSlug == nil || *v.DocumentSlug != "frd-refunds" {
					t.Errorf("detail = %#v, want both slugs", v)
				}
			},
		},
		{
			name:     "owner unresolved",
			finding:  finding.OwnerUnresolved("specs/adrs/adr-011.md", position, "frd-does-not-exist"),
			rule:     finding.RuleOwnerUnresolved,
			contains: []string{"frd-does-not-exist"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsOwningDocument()
				if !ok {
					t.Fatalf("detail is not an owning document: %#v", d.Value)
				}
				if v.Slug == nil || *v.Slug != "frd-does-not-exist" {
					t.Errorf("Slug = %v, want %q", v.Slug, "frd-does-not-exist")
				}
				if v.DocumentSlug != nil {
					t.Errorf("DocumentSlug = %v, want nil", v.DocumentSlug)
				}
			},
		},
		{
			name:     "gherkin parse",
			finding:  finding.GherkinParse("specs/adrs/adr-011.md", position, "a step keyword"),
			rule:     finding.RuleGherkinParse,
			contains: []string{"a step keyword"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsGherkinParse()
				if !ok {
					t.Fatalf("detail is not a gherkin parse failure: %#v", d.Value)
				}
				if v.Expected != "a step keyword" {
					t.Errorf("Expected = %q, want %q", v.Expected, "a step keyword")
				}
			},
		},
		{
			name:     "criterion empty",
			finding:  finding.CriterionEmpty("specs/adrs/adr-011.md", position),
			rule:     finding.RuleCriterionEmpty,
			contains: []string{"no steps"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsMissingProperty()
				if !ok {
					t.Fatalf("detail is not a missing property: %#v", d.Value)
				}
				if v.Property != "steps" {
					t.Errorf("Property = %q, want %q", v.Property, "steps")
				}
			},
		},
		{
			name:     "scenario prose",
			finding:  finding.ScenarioProse("specs/adrs/adr-011.md", position, `^AC-[A-Z0-9]+-[0-9]{3}$`),
			rule:     finding.RuleScenarioProse,
			contains: []string{"normative", "untagged criterion"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsIdentifierPattern()
				if !ok {
					t.Fatalf("detail is not an identifier pattern: %#v", d.Value)
				}
				if v.Identifier != nil {
					t.Errorf("Identifier = %v, want nil: this is prose, not an attempted identifier", v.Identifier)
				}
				if v.Pattern != `^AC-[A-Z0-9]+-[0-9]{3}$` {
					t.Errorf("Pattern = %q, want the registered criterion pattern", v.Pattern)
				}
			},
		},
		{
			name:     "examples partially tagged",
			finding:  finding.ExamplesPartiallyTagged("specs/adrs/adr-011.md", []specmodel.Span{position}),
			rule:     finding.RuleExamplesPartiallyTagged,
			contains: []string{"1 Examples block"},
			detail: func(t *testing.T, d specmodel.FindingDetail) {
				v, ok := d.AsPartiallyTaggedOutline()
				if !ok {
					t.Fatalf("detail is not a partially tagged outline: %#v", d.Value)
				}
				if len(v.Untagged) != 1 || v.Untagged[0] != position {
					t.Errorf("Untagged = %v, want exactly the one span %#v", v.Untagged, position)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := c.finding

			if f.Rule != c.rule {
				t.Errorf("Rule = %q, want %q", f.Rule, c.rule)
			}
			if f.Severity == "" {
				t.Error("Severity is empty; every finding carries one")
			}
			if f.Path != "specs/adrs/adr-011.md" {
				t.Errorf("Path = %q, want the file the finding is against", f.Path)
			}
			if f.Position != position {
				t.Errorf("Position = %#v, want %#v", f.Position, position)
			}
			if f.Message == "" {
				t.Error("Message is empty; a finding with no message is not self-contained")
			}
			for _, want := range c.contains {
				if !strings.Contains(f.Message, want) {
					t.Errorf("Message %q does not quote %q", f.Message, want)
				}
			}
			c.detail(t, f.Detail)
		})
	}
}

// strPtr returns a pointer to s, for building the optional string fields
// a finding constructor takes.
func strPtr(s string) *string { return &s }

// RetiredTarget carries severity warning, never error — the target
// still resolves, so a citation to it is not a conformance failure — and
// its message and detail omit the successor entirely rather than
// reporting one that does not exist when the target declares none.
//
// AC: AC-SPECREF-002.6
func TestRetiredTargetCarriesWarningAndOmitsAnAbsentSuccessor(t *testing.T) {
	f := finding.RetiredTarget("specs/adrs/adr-011.md", at(t), "adr-009", nil)

	if f.Severity != finding.SeverityWarning {
		t.Errorf("severity = %q, want %q — a reference to a retired document still resolves", f.Severity, finding.SeverityWarning)
	}
	if strings.Contains(f.Message, "superseded") {
		t.Errorf("message = %q, wants no mention of a successor the target does not declare", f.Message)
	}
	v, ok := f.Detail.AsRetiredTarget()
	if !ok {
		t.Fatalf("detail is not a retired target: %#v", f.Detail.Value)
	}
	if v.Successor != nil {
		t.Errorf("Successor = %v, want nil", v.Successor)
	}
}

// IncompletePairing and ReferenceType both carry severity error: unlike
// a retired target, neither the half-declared relationship nor the
// wrong-typed target is a defect a build can tolerate.
//
// AC: AC-SPECREF-002.5
// AC: AC-SPECREF-002.10
func TestPairingAndTypeMismatchFindingsCarryError(t *testing.T) {
	pairing := finding.IncompletePairing("specs/adrs/adr-011.md", at(t),
		"owner", "owned_by", "specs/adrs/adr-011.md", "specs/teams/team-ledger.md")
	if pairing.Severity != finding.SeverityError {
		t.Errorf("pairing severity = %q, want %q", pairing.Severity, finding.SeverityError)
	}

	mismatch := finding.ReferenceType("specs/features/frd-checkout.md", at(t), "frd", "adr-009", "frd", "adr")
	if mismatch.Severity != finding.SeverityError {
		t.Errorf("type-mismatch severity = %q, want %q", mismatch.Severity, finding.SeverityError)
	}
}

// CorpusEmpty carries severity error unconditionally — a caller must not
// need --strict to see it — and names the configured spec root in both
// its Path and its Message, since there is no document to report it
// against and the whole point is telling the operator what they pointed
// the tool at.
func TestCorpusEmptyCarriesErrorAndNamesTheSpecRoot(t *testing.T) {
	f := finding.CorpusEmpty("docs")

	if f.Rule != finding.RuleCorpusEmpty {
		t.Errorf("Rule = %q, want %q", f.Rule, finding.RuleCorpusEmpty)
	}
	if f.Severity != finding.SeverityError {
		t.Errorf("Severity = %q, want %q", f.Severity, finding.SeverityError)
	}
	if f.Path != "docs" {
		t.Errorf("Path = %q, want the configured spec root", f.Path)
	}
	if !strings.Contains(f.Message, `"docs"`) {
		t.Errorf("Message %q does not name the configured spec root", f.Message)
	}
}

// A consumer keying off `detail.kind` must be told what this finding is,
// not what some other check reports. The text rendering has always been
// right; the structured half is what a catalogue consuming the index
// acts on, and a corpus-level finding announcing itself as a frontmatter
// parse failure sends that consumer looking for a document that does not
// exist.
func TestCorpusEmptyDetailNamesItsOwnKind(t *testing.T) {
	encoded, err := json.Marshal(finding.CorpusEmpty("docs"))
	if err != nil {
		t.Fatalf("encoding the finding: %v", err)
	}

	var decoded struct {
		Detail struct {
			Kind string `json:"kind"`
			Root string `json:"root"`
		} `json:"detail"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decoding the finding: %v", err)
	}

	if decoded.Detail.Kind != "corpus-empty" {
		t.Errorf("detail.kind = %q, want %q", decoded.Detail.Kind, "corpus-empty")
	}
	if decoded.Detail.Root != "docs" {
		t.Errorf("detail.root = %q, want the configured spec root", decoded.Detail.Root)
	}
}

// The two unreachable-citation findings carry distinct `detail.kind`
// values and distinct severities — an unmapped citation is an error, an
// unresolved one a warning — so a JSON consumer branches on structure
// rather than parsing prose to tell a gap in the map from a citation
// that names no corpus document at all.
//
// AC: AC-SPECMD-002.4
func TestUnmappedAndUnresolvedCitationsAreDistinguishableByKind(t *testing.T) {
	unmapped := finding.UnmappedCitation("specs/adrs/adr-011.md", at(t), "document-relationship-standards", "links.yaml")
	unresolved := finding.UnresolvedCitation("specs/adrs/adr-011.md", at(t), "adr-404")

	if unmapped.Detail.Value == unresolved.Detail.Value {
		t.Errorf("the two details compare equal: %#v", unmapped.Detail.Value)
	}
	if _, ok := unmapped.Detail.AsUnresolvedCitation(); ok {
		t.Error("an unmapped citation's detail also reads as an unresolved citation")
	}
	if _, ok := unresolved.Detail.AsUnmappedCitation(); ok {
		t.Error("an unresolved citation's detail also reads as an unmapped citation")
	}

	if unmapped.Severity != finding.SeverityError {
		t.Errorf("unmapped severity = %q, want %q", unmapped.Severity, finding.SeverityError)
	}
	if unresolved.Severity != finding.SeverityWarning {
		t.Errorf("unresolved severity = %q, want %q", unresolved.Severity, finding.SeverityWarning)
	}

	encodedUnmapped, err := json.Marshal(unmapped)
	if err != nil {
		t.Fatalf("marshal unmapped: %v", err)
	}
	encodedUnresolved, err := json.Marshal(unresolved)
	if err != nil {
		t.Fatalf("marshal unresolved: %v", err)
	}
	if !strings.Contains(string(encodedUnmapped), `"kind":"unmapped-citation"`) {
		t.Errorf("unmapped JSON = %s, want detail.kind %q", encodedUnmapped, "unmapped-citation")
	}
	if !strings.Contains(string(encodedUnresolved), `"kind":"unresolved-citation"`) {
		t.Errorf("unresolved JSON = %s, want detail.kind %q", encodedUnresolved, "unresolved-citation")
	}
}

// Both citation findings name the citing document (as Path), the slug
// (in Message and in the structured detail) and the position of the
// citation (as Position) — the three facts a reader needs without
// re-opening the document.
//
// AC: AC-SPECMD-002.3
func TestCitationFindingsNameTheCitingDocumentSlugAndPosition(t *testing.T) {
	position := at(t)

	cases := []struct {
		name    string
		finding specmodel.Finding
		slug    string
	}{
		{"unmapped", finding.UnmappedCitation("specs/adrs/adr-011.md", position, "document-relationship-standards", "links.yaml"), "document-relationship-standards"},
		{"unresolved", finding.UnresolvedCitation("specs/adrs/adr-011.md", position, "adr-404"), "adr-404"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.finding.Path != "specs/adrs/adr-011.md" {
				t.Errorf("Path = %q, want the citing document", c.finding.Path)
			}
			if c.finding.Position != position {
				t.Errorf("Position = %#v, want the citation's position", c.finding.Position)
			}
			if !strings.Contains(c.finding.Message, c.slug) {
				t.Errorf("Message %q does not name the slug %q", c.finding.Message, c.slug)
			}
		})
	}
}

// The two IncompletePairing findings the design calls for — one against
// the declaring document, one against the target — carry the identical
// structured detail, so checking either document surfaces the same
// half-built edge described the same way.
//
// AC: AC-SPECREF-002.5
func TestIncompletePairingCarriesTheSameDetailAgainstBothDocuments(t *testing.T) {
	source := finding.IncompletePairing("specs/adrs/adr-011.md", at(t),
		"owner", "owned_by", "specs/adrs/adr-011.md", "specs/teams/team-ledger.md")
	target := finding.IncompletePairing("specs/teams/team-ledger.md", at(t),
		"owner", "owned_by", "specs/adrs/adr-011.md", "specs/teams/team-ledger.md")

	sourceDetail, ok := source.Detail.AsIncompletePairing()
	if !ok {
		t.Fatalf("source detail is not an incomplete pairing: %#v", source.Detail.Value)
	}
	targetDetail, ok := target.Detail.AsIncompletePairing()
	if !ok {
		t.Fatalf("target detail is not an incomplete pairing: %#v", target.Detail.Value)
	}
	if sourceDetail != targetDetail {
		t.Errorf("source detail = %#v, target detail = %#v, want them identical", sourceDetail, targetDetail)
	}
	if source.Path == target.Path {
		t.Error("the two findings are against the same path; the fixture proves nothing about both documents")
	}
}

// The five properties AC-SPEC-006.1 names must survive the JSON
// boundary, because the finding JSON is what an agent parses.
//
// AC: AC-SPEC-006.1
func TestAFindingSerialisesEveryPropertyItCarries(t *testing.T) {
	encoded, err := json.Marshal(finding.SlugPattern("specs/adrs/adr-011.md", at(t), "adr-011-x", `^adr-[0-9]{3}$`))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, property := range []string{"rule", "severity", "path", "position", "message", "detail"} {
		if _, ok := decoded[property]; !ok {
			t.Errorf("emitted finding has no %q property: %s", property, encoded)
		}
	}

	var detail map[string]json.RawMessage
	if err := json.Unmarshal(decoded["detail"], &detail); err != nil {
		t.Fatalf("the detail is not an object: %v", err)
	}
	if string(detail["kind"]) != `"slug-pattern"` {
		t.Errorf("detail kind = %s, want %q", detail["kind"], "slug-pattern")
	}
	for _, property := range []string{"slug", "pattern"} {
		if _, ok := detail[property]; !ok {
			t.Errorf("the detail has no %q property: %s", property, encoded)
		}
	}
}

// A check that was requested and could not run is reported as
// unverified — both as the check's outcome and as a finding — and never
// as passed.
//
// AC: AC-SPEC-006.3
// AC: AC-SPEC-006.4
func TestAnUnrunnableCheckIsUnverifiedRatherThanPassed(t *testing.T) {
	report := finding.NewReport(finding.RuleSlugPattern, "structure/mermaid")
	report.Request(finding.RuleSlugPattern)
	report.Ran(finding.RuleSlugPattern)
	report.Request("structure/mermaid")
	report.Unverified("structure/mermaid", "specs/adrs/adr-011.md", at(t), "no diagram validator is available")

	model := report.Model()

	result := checkResult(t, model, "structure/mermaid")
	if result.Outcome != specmodel.CheckOutcomeUnverified {
		t.Errorf("outcome = %q, want %q", result.Outcome, specmodel.CheckOutcomeUnverified)
	}
	if result.Outcome == specmodel.CheckOutcomePassed {
		t.Error("a check that could not run is reported as passed")
	}
	if result.Reason == nil || *result.Reason != "no diagram validator is available" {
		t.Errorf("Reason = %v, want the reason the check could not run", result.Reason)
	}

	if len(model.Findings) != 1 {
		t.Fatalf("findings = %d, want one unverified finding", len(model.Findings))
	}
	if model.Findings[0].Severity != finding.SeverityUnverified {
		t.Errorf("severity = %q, want %q", model.Findings[0].Severity, finding.SeverityUnverified)
	}
	if report.HasErrors() {
		t.Error("an unverified finding is not an error")
	}
}

// A check that is requested and reports nothing at all is unverified,
// not passed: silence must read as "not known" rather than as "fine".
//
// AC: AC-SPEC-006.4
func TestARequestedCheckThatReportsNothingIsNotPassed(t *testing.T) {
	report := finding.NewReport(finding.RuleDuplicateSlug)
	report.Request(finding.RuleDuplicateSlug)

	result := checkResult(t, report.Model(), finding.RuleDuplicateSlug)
	if result.Outcome != specmodel.CheckOutcomeUnverified {
		t.Errorf("outcome = %q, want %q", result.Outcome, specmodel.CheckOutcomeUnverified)
	}
	if result.Reason == nil || *result.Reason == "" {
		t.Error("an unverified outcome carries no reason")
	}
}

// The report distinguishes a check that passed from one that was never
// requested, and from one that failed — so a clean-looking report cannot
// be read as covering checks that never ran.
//
// AC: AC-SPEC-006.9
func TestTheReportDistinguishesPassedFromUnrequested(t *testing.T) {
	report := finding.NewReport(finding.RuleSlugPattern, finding.RuleDuplicateSlug, finding.RuleMissingProperty)
	report.Request(finding.RuleDuplicateSlug)
	report.Ran(finding.RuleDuplicateSlug)
	report.Request(finding.RuleSlugPattern)
	report.Add(finding.SlugPattern("specs/adrs/adr-011.md", at(t), "adr-011-x", `^adr-[0-9]{3}$`))
	report.Ran(finding.RuleSlugPattern)

	model := report.Model()

	want := map[specmodel.RuleId]specmodel.CheckOutcome{
		finding.RuleDuplicateSlug:   specmodel.CheckOutcomePassed,
		finding.RuleSlugPattern:     specmodel.CheckOutcomeFailed,
		finding.RuleMissingProperty: specmodel.CheckOutcomeNotRequested,
	}
	if len(model.Checks) != len(want) {
		t.Fatalf("checks = %d, want %d — every known check is accounted for", len(model.Checks), len(want))
	}
	for rule, outcome := range want {
		if got := checkResult(t, model, rule).Outcome; got != outcome {
			t.Errorf("%s outcome = %q, want %q", rule, got, outcome)
		}
	}

	// The distinction must survive to the JSON an agent reads.
	encoded, err := report.JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if !strings.Contains(string(encoded), `"not-requested"`) {
		t.Errorf("emitted report does not distinguish an un-requested check: %s", encoded)
	}
}

// A LIFE check whose registration is empty — no pairing, no typed
// reference — is still requested by a run that requests it, and reports
// passed rather than not-requested: "the profile registers none" and
// "the check did not run" are different facts
// (reference-integrity-blueprint.md's LIFE section). The mechanism is
// [Report]'s own, already proven rule-agnostically by
// TestTheReportDistinguishesPassedFromUnrequested; this applies it to
// the three rule identifiers this delta adds, so the guarantee is
// checked against their own names rather than only in the abstract.
func TestALIFECheckWithEmptyRegistrationReportsPassedNotUnrequested(t *testing.T) {
	rules := []specmodel.RuleId{finding.RuleRetiredTarget, finding.RuleIncompletePairing, finding.RuleTypeMismatch}
	report := finding.NewReport(rules...)
	for _, rule := range rules {
		report.Request(rule)
		// corpus.Check found nothing to report for this rule — an empty
		// registration produces zero findings, not a skipped check — so the
		// caller still calls Ran once the check has run its course.
		report.Ran(rule)
	}

	model := report.Model()
	for _, rule := range rules {
		if got := checkResult(t, model, rule).Outcome; got != specmodel.CheckOutcomePassed {
			t.Errorf("%s outcome = %q, want %q", rule, got, specmodel.CheckOutcomePassed)
		}
	}
}

// A finding raised for a check the report was not told about is kept and
// accounted for rather than dropped.
func TestAFindingForAnUndeclaredCheckIsStillAccountedFor(t *testing.T) {
	report := finding.NewReport()
	report.Add(finding.DuplicateSlug("specs/adrs/adr-011.md", at(t), "adr-011", "specs/adrs/other.md"))

	model := report.Model()
	if len(model.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(model.Findings))
	}
	if got := checkResult(t, model, finding.RuleDuplicateSlug).Outcome; got != specmodel.CheckOutcomeFailed {
		t.Errorf("outcome = %q, want %q", got, specmodel.CheckOutcomeFailed)
	}
	if !report.HasErrors() {
		t.Error("HasErrors is false with an error finding recorded")
	}
}

// The report renders as human-readable text and as the generated finding
// JSON from the same recorded state, so the command surface selects a
// rendering rather than formatting findings itself.
//
// This participates in AC-SPEC-006.5, which the command surface proves at
// integration level; what is asserted here is only that both renderings
// exist and agree.
func TestTheReportRendersAsTextAndAsJSON(t *testing.T) {
	report := finding.NewReport(finding.RuleSlugPattern, finding.RuleMissingProperty)
	report.Request(finding.RuleSlugPattern)
	report.Add(finding.SlugPattern("specs/adrs/adr-011.md", at(t), "adr-011-x", `^adr-[0-9]{3}$`))
	report.Ran(finding.RuleSlugPattern)

	text := report.Text()
	for _, want := range []string{"slug/pattern-mismatch", "error", "specs/adrs/adr-011.md:", "adr-011-x", "checks:", "not-requested"} {
		if !strings.Contains(text, want) {
			t.Errorf("text rendering does not mention %q:\n%s", want, text)
		}
	}

	encoded, err := report.JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var decoded specmodel.FindingReport
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("the emitted JSON does not decode as a report: %v", err)
	}
	if len(decoded.Findings) != 1 || decoded.Findings[0].Rule != finding.RuleSlugPattern {
		t.Errorf("decoded findings = %#v, want the one recorded finding", decoded.Findings)
	}
	if _, ok := decoded.Findings[0].Detail.AsSlugPattern(); !ok {
		t.Error("the variant did not survive the JSON round trip")
	}
}

// An empty report emits empty collections rather than nulls, so output
// stays comparable and a consumer can iterate without a nil check.
func TestAnEmptyReportEmitsEmptyCollections(t *testing.T) {
	encoded, err := finding.NewReport().JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if want := `{"checks":[],"findings":[]}`; string(encoded) != want {
		t.Errorf("emitted report = %s, want %s", encoded, want)
	}
	if text := finding.NewReport().Text(); !strings.Contains(text, "no findings") {
		t.Errorf("text rendering of an empty report = %q", text)
	}
}

// Findings and checks are emitted in a fixed order regardless of the
// order they were recorded in, because a consumer diffing two runs and
// every snapshot case depends on it.
func TestOutputOrderIsDeterministic(t *testing.T) {
	source := []byte(strings.Repeat("x", 400))
	early := span.Of(source, 10, 5)
	late := span.Of(source, 300, 5)

	forward := finding.NewReport(finding.RuleSlugPattern, finding.RuleDuplicateSlug)
	forward.Add(
		finding.DuplicateSlug("b.md", late, "s", "a.md"),
		finding.SlugPattern("a.md", late, "s", "^p$"),
		finding.SlugPattern("a.md", early, "s", "^p$"),
	)

	backward := finding.NewReport(finding.RuleDuplicateSlug, finding.RuleSlugPattern)
	backward.Add(
		finding.SlugPattern("a.md", early, "s", "^p$"),
		finding.SlugPattern("a.md", late, "s", "^p$"),
		finding.DuplicateSlug("b.md", late, "s", "a.md"),
	)

	first, err := forward.JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	second, err := backward.JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if string(first) != string(second) {
		t.Errorf("two orders of the same run produced different reports:\n%s\n%s", first, second)
	}

	model := forward.Model()
	wantPaths := []string{"a.md", "a.md", "b.md"}
	for i, want := range wantPaths {
		if model.Findings[i].Path != want {
			t.Errorf("findings[%d].Path = %q, want %q", i, model.Findings[i].Path, want)
		}
	}
	if model.Findings[0].Position.Offset != early.Offset {
		t.Error("findings within one file are not ordered by position")
	}
	if model.Checks[0].Rule != finding.RuleDuplicateSlug {
		t.Errorf("checks[0] = %q, want the checks ordered by rule", model.Checks[0].Rule)
	}
}

// The model the report emits is the generated one: the base finding, the
// severity vocabulary, the variant union and the report envelope are
// declared in schemas/finding.strata, not assembled in Go. These
// assertions read the generated bundle so that a change loosening the
// Strata source fails here.
//
// AC: AC-SPEC-006.1
func TestTheGeneratedSchemaDeclaresTheFindingModel(t *testing.T) {
	bundle := loadBundle(t)

	base := definition(t, bundle, "Finding")
	for _, property := range []string{"rule", "severity", "path", "position", "message", "detail"} {
		if !isRequired(base, property) {
			t.Errorf("Finding does not require %q", property)
		}
	}

	severity := definition(t, bundle, "Severity")
	values, _ := severity["enum"].([]any)
	got := make(map[string]bool, len(values))
	for _, value := range values {
		if name, ok := value.(string); ok {
			got[name] = true
		}
	}
	for _, want := range []string{"error", "warning", "unverified"} {
		if !got[want] {
			t.Errorf("Severity does not declare %q; enum is %v", want, values)
		}
	}
	if len(values) != 3 {
		t.Errorf("Severity declares %d values, want exactly the three the FRD defines", len(values))
	}

	union := definition(t, bundle, "FindingDetail")
	branches, ok := union["allOf"].([]any)
	if !ok || len(branches) == 0 {
		t.Fatalf("FindingDetail is not a discriminated union: %v", union)
	}
	for _, name := range []string{"FindingSchemaViolation", "FindingMissingProperty", "FindingSlugPattern", "FindingDuplicateSlug", "FindingUnverified"} {
		if _, ok := bundle["definitions"].(map[string]any)[name]; !ok {
			t.Errorf("the bundle declares no variant %q", name)
		}
	}

	envelope := definition(t, bundle, "FindingReport")
	for _, property := range []string{"checks", "findings"} {
		if !isRequired(envelope, property) {
			t.Errorf("FindingReport does not require %q", property)
		}
	}
	if !isRequired(definition(t, bundle, "CheckResult"), "outcome") {
		t.Error("CheckResult does not require an outcome")
	}
}

func loadBundle(t *testing.T) map[string]any {
	t.Helper()

	raw, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read %s: %v", schemaPath, err)
	}
	var bundle map[string]any
	if err := json.Unmarshal(raw, &bundle); err != nil {
		t.Fatalf("parse %s: %v", schemaPath, err)
	}
	return bundle
}

func definition(t *testing.T, bundle map[string]any, name string) map[string]any {
	t.Helper()

	definitions, ok := bundle["definitions"].(map[string]any)
	if !ok {
		t.Fatalf("%s declares no definitions object", schemaPath)
	}
	found, ok := definitions[name].(map[string]any)
	if !ok {
		t.Fatalf("%s declares no definition %q", schemaPath, name)
	}
	return found
}

func isRequired(definition map[string]any, name string) bool {
	required, _ := definition["required"].([]any)
	for _, entry := range required {
		if entry == name {
			return true
		}
	}
	return false
}

func checkResult(t *testing.T, model specmodel.FindingReport, rule specmodel.RuleId) specmodel.CheckResult {
	t.Helper()

	for _, result := range model.Checks {
		if result.Rule == rule {
			return result
		}
	}
	t.Fatalf("the report accounts for no check %q", rule)
	return specmodel.CheckResult{}
}

// UnresolvedReference's message reads differently depending on where a
// reference was written and what it matched: a body reference is quoted
// back with its brackets, a frontmatter reference names the property
// that holds it instead, and a non-member target states why it is not a
// member rather than reading as an ordinary unresolved slug.
//
// AC: AC-SPECREF-002.1, AC-SPECREF-002.3, AC-SPECREF-002.7
func TestUnresolvedReferenceMessageVariesByOriginAndNonMember(t *testing.T) {
	position := at(t)
	bodyOrigin := specmodel.NewReferenceOriginBody(specmodel.ReferenceBodyOrigin{})
	fmOrigin := specmodel.NewReferenceOriginFrontmatterProperty(
		specmodel.ReferenceFrontmatterOrigin{Property: "superseded_by"})

	body := finding.UnresolvedReference("specs/adrs/adr-011.md", position, "adr-404", "adr-404", bodyOrigin, nil)
	if !strings.Contains(body.Message, "[[adr-404]]") {
		t.Errorf("body message = %q, want it quoted with its brackets", body.Message)
	}

	fm := finding.UnresolvedReference("specs/adrs/adr-011.md", position, "adr-404", "adr-404", fmOrigin, nil)
	if !strings.Contains(fm.Message, "superseded_by") {
		t.Errorf("frontmatter message = %q, want the property named", fm.Message)
	}
	if strings.Contains(fm.Message, "[[") {
		t.Errorf("frontmatter message = %q, wants no body-style brackets", fm.Message)
	}

	reason := specmodel.ExclusionReasonNoFrontmatter
	excluded := finding.UnresolvedReference("specs/adrs/adr-011.md", position, "loose", "loose", bodyOrigin,
		&specmodel.ReferenceNonMember{Path: "specs/notes/loose.md", Reason: &reason})
	if !strings.Contains(excluded.Message, "not a corpus member") {
		t.Errorf("non-member message = %q, want it to say the target is not a corpus member", excluded.Message)
	}
	if strings.Contains(excluded.Message, "matches no document's slug") {
		t.Errorf("non-member message = %q, reads as an ordinary unresolved slug", excluded.Message)
	}

	parseFailed := finding.UnresolvedReference("specs/adrs/adr-011.md", position, "broken", "broken", bodyOrigin,
		&specmodel.ReferenceNonMember{
			Path:       "specs/notes/broken.md",
			ParseError: &specmodel.ParseError{Message: "unterminated string", Position: position},
		})
	if want := fmt.Sprintf("line %d", position.StartLine); !strings.Contains(parseFailed.Message, want) {
		t.Errorf("parse-failure message = %q, want it to cite the failure's line (%s)", parseFailed.Message, want)
	}
}

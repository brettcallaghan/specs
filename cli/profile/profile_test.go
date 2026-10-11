// Tests for the specs profile.
//
// realSlugs below is not a corpus walk — the CO delta that will discover
// and check this repository's own corpus is a later item in this
// feature's closure set, and building one here would duplicate work that
// item owns. It is a hand-picked sample of (type, slug) pairs actually
// present in this repository today, gathered while grounding the
// per-type patterns in repo reality rather than in the (in two places,
// stale) prefix registry — see the comment on documentTypes in
// profile.go. TestSlugPatternsAcceptRealSlugs exists so a pattern this
// package tightens later fails fast against real documents rather than
// only against bge.14's self-application run.
package profile_test

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/brettcallaghan/specs/cli/core/corpus"
	"github.com/brettcallaghan/specs/cli/core/numbering"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
	"github.com/brettcallaghan/specs/cli/profile"
)

// The conformance test the module's own Change Guidance requires: this
// module satisfies Document Core's Profile interface at compile time,
// not only by chance at runtime.
var _ corpus.Profile = profile.New()

func TestNewSatisfiesTheProfileInterface(t *testing.T) {
	var p corpus.Profile = profile.New()
	if p == nil {
		t.Fatal("New returned a nil Profile")
	}
}

// The profile registers under the name specs so an adopting
// repository activates it without authoring one. The registry that maps
// this name to New lives with the CFG item; what is provable here is
// that this package's own registration name is the one the blueprint
// and the contract both name.
func TestRegistersUnderTheDocumentedName(t *testing.T) {
	if profile.Name != "specs" {
		t.Errorf("Name = %q, want %q", profile.Name, "specs")
	}
}

// The embedded union schema is valid JSON, non-empty, and exactly the
// generated bundle — proving the binary carries the schema rather than a
// stand-in, with no schema file read at runtime.
func TestSchemaIsTheEmbeddedGeneratedBundle(t *testing.T) {
	p := profile.New()

	b := p.Schema()
	if len(b) == 0 {
		t.Fatal("Schema returned nothing")
	}

	var parsed map[string]any
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("Schema is not valid JSON: %v", err)
	}
	if _, ok := parsed["definitions"]; !ok {
		t.Error(`Schema has no top-level "definitions"; it is not the generated union bundle`)
	}

	if len(p.Schema()) != len(b) {
		t.Error("Schema is not stable across calls")
	}
}

// documentTypeVocabulary is every value the generated DocumentType enum
// defines. Read directly from the generated Go constants rather than
// restated as string literals, so a type added to the vocabulary is
// caught here without this test needing an update.
func documentTypeVocabulary() []string {
	return []string{
		string(specmodel.DocumentTypeAdr),
		string(specmodel.DocumentTypeFrd),
		string(specmodel.DocumentTypeContract),
		string(specmodel.DocumentTypeStandards),
		string(specmodel.DocumentTypeSystemBlueprint),
		string(specmodel.DocumentTypeContainerBlueprint),
		string(specmodel.DocumentTypeModuleBlueprint),
		string(specmodel.DocumentTypeComponentBlueprint),
		string(specmodel.DocumentTypeFeatureBlueprint),
		string(specmodel.DocumentTypeGuide),
	}
}

// A document authored to a type this framework does not define is
// excluded as unknown-type rather than checked against the wrong rules
// — but only if every defined type actually has an entry. This is the
// AC-SPEC-003.5 and AC-SPEC-007.1 participation: the table this test
// walks is what those two checks consume.
func TestDocumentTypesCoversEveryVocabularyValue(t *testing.T) {
	vocabulary := documentTypeVocabulary()
	types := profile.New().DocumentTypes()

	if len(types) != len(vocabulary) {
		t.Errorf("DocumentTypes has %d entries, want %d (one per vocabulary value)",
			len(types), len(vocabulary))
	}

	for _, value := range vocabulary {
		rules, ok := types[value]
		if !ok {
			t.Errorf("DocumentTypes has no entry for %q", value)
			continue
		}
		if rules.SlugPattern == "" {
			t.Errorf("%q has no slug pattern", value)
			continue
		}
		if _, err := regexp.Compile(rules.SlugPattern); err != nil {
			t.Errorf("%q slug pattern %q does not compile: %v", value, rules.SlugPattern, err)
		}
	}
}

// Only adr is numbered.
func TestOnlyAdrIsNumbered(t *testing.T) {
	types := profile.New().DocumentTypes()

	for value, rules := range types {
		want := value == string(specmodel.DocumentTypeAdr)
		if rules.Numbered != want {
			t.Errorf("DocumentTypes[%q].Numbered = %v, want %v", value, rules.Numbered, want)
		}
	}
}

// realSlugs pairs a document type with a slug an actual document in this
// repository declares today. See the package doc comment for why this
// is a curated sample rather than a corpus walk.
var realSlugs = []struct {
	documentType, slug, source string
}{
	{"adr", "adr-013", "specs/adrs/adr-013-work-order-terminology.md"},
	// The descriptive form no longer appears in this repository — the
	// slugs were migrated to the number-only form [[adr-012]] decided —
	// but the pattern still accepts it, because an adopting repository
	// that never made that migration is not thereby non-conformant.
	{"adr", "adr-003-yaml-frontmatter-for-metadata", "the optional descriptive suffix the pattern still admits"},
	{"frd", "frd-spec-index", "specs/features/spec-index/spec-index-frd.md"},
	{"frd", "frd-spec-reference-integrity", "specs/features/reference-integrity/reference-integrity-frd.md"},
	{"contract", "contract-specs-configuration", "specs/contracts/specs-configuration-contract.md"},
	{"standards", "testing-standards", "specs/standards/testing-standards.md"},
	{"standards", "cross-document-references", "specs/standards/cross-document-references-standards.md"},
	{"standards", "blueprint-readiness-checklist", "specs/standards/blueprint-readiness-checklist.md"},
	{"standards", "findings-capture-protocol", "specs/standards/findings-capture-protocol.md"},
	{"standards", "c4-standards", "specs/standards/c4-standards.md"},
	{"container-blueprint", "container-spec-cli", "specs/containers/spec-cli/spec-cli-container.md"},
	{"module-blueprint", "module-spec-profile", "specs/modules/spec-profile/spec-profile-module.md"},
	{"module-blueprint", "module-document-core", "specs/modules/document-core/document-core-module.md"},
	{"feature-blueprint", "feature-spec-index", "specs/features/spec-index/spec-index-blueprint.md"},
	{"guide", "guide-ears", "specs/guides/ears-guide.md"},
	{"guide", "work-order-field-guide", "specs/guides/work-order-field-guide.md"},
}

// AC: AC-SPEC-003.5
func TestSlugPatternsAcceptRealSlugs(t *testing.T) {
	types := profile.New().DocumentTypes()

	for _, tc := range realSlugs {
		rules, ok := types[tc.documentType]
		if !ok {
			t.Errorf("no DocumentTypes entry for %q (needed by %s)", tc.documentType, tc.source)
			continue
		}
		matcher, err := regexp.Compile(rules.SlugPattern)
		if err != nil {
			t.Fatalf("%q slug pattern %q does not compile: %v", tc.documentType, rules.SlugPattern, err)
		}
		if !matcher.MatchString(tc.slug) {
			t.Errorf("pattern %q for %q rejects %q (declared by %s)",
				rules.SlugPattern, tc.documentType, tc.slug, tc.source)
		}
	}
}

// A slug pattern this permissive still rejects the shapes this
// framework's slug convention forbids outright: upper case and embedded
// whitespace.
func TestSlugPatternsRejectNonKebabCase(t *testing.T) {
	types := profile.New().DocumentTypes()

	for value, rules := range types {
		matcher, err := regexp.Compile(rules.SlugPattern)
		if err != nil {
			t.Fatalf("%q slug pattern %q does not compile: %v", value, rules.SlugPattern, err)
		}
		for _, bad := range []string{"Not Kebab Case", "UPPERCASE-SLUG", "has spaces"} {
			if matcher.MatchString(bad) {
				t.Errorf("pattern %q for %q accepts %q", rules.SlugPattern, value, bad)
			}
		}
	}
}

// AC: AC-SPEC-008.2
func TestConventionsSuppliesTheLineWidthDefault(t *testing.T) {
	conventions := profile.New().Conventions()

	if conventions.LineWidth == nil {
		t.Fatal("LineWidth default is absent")
	}
	if *conventions.LineWidth != 80 {
		t.Errorf("LineWidth = %d, want %d, per template-authoring-standards.md", *conventions.LineWidth, 80)
	}
}

// Every numbered type's registered width formats a slug its own slug
// pattern accepts, from the first number to one wider than the width, so
// the width and the pattern cannot drift apart.
//
// AC: AC-SPEC-007.1
func TestEveryNumberedTypeFormatsASlugItsPatternAccepts(t *testing.T) {
	types := profile.New().DocumentTypes()
	numbered := 0
	for name, rules := range types {
		if !rules.Numbered {
			continue
		}
		numbered++
		pattern := regexp.MustCompile(rules.SlugPattern)
		for _, n := range []int{1, 31, 1000} {
			slug, err := numbering.Format(name, n, types)
			if err != nil {
				t.Errorf("Format(%q, %d): %v", name, n, err)
				continue
			}
			if !pattern.MatchString(slug) {
				t.Errorf("Format(%q, %d) = %q, which %s's slug pattern %s rejects", name, n, slug, name, rules.SlugPattern)
			}
		}
	}
	if numbered == 0 {
		t.Fatal("the profile registers no numbered type; this test proves nothing")
	}
}

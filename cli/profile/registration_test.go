// Tests for the reference, typed-reference, pairing and lifecycle
// registrations this profile supplies.
//
// The drift guard below reads the generated schema through its own
// untyped traversal rather than through the decoder registration.go uses.
// That is deliberate: a guard sharing the production decode would agree
// with it by construction, including where the decode is wrong, and would
// report a registration derived from a misread schema as complete.
package profile_test

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/schema"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
	"github.com/brettcallaghan/agent-plugins/cli/profile"
)

// registeredReferenceProperties is the complete set of reference-bearing
// properties this framework registers, transcribed here on purpose: the
// production set is derived from the schema, so a test that derived it
// the same way would assert nothing about which properties the
// registration actually contains.
//
// It includes `provider` and `consumers`, which AC-SPECREF-004.2's
// illustrative list omits. Both hold slugs by declaration and both
// resolve in this corpus today, and leaving them out would leave a
// contract's two most load-bearing edges unchecked; the criterion is
// "this framework's own reference-bearing properties", and this is the
// complete set. That divergence is decided — see the PREG delta in
// specs/features/reference-integrity/reference-integrity-blueprint.md.
//
// `sources` appears against `frd` alone, because
// schemas/frontmatter.strata declares it on `Frd` alone.
var registeredReferenceProperties = map[string][]string{
	"adr":                 {"superseded_by"},
	"frd":                 {"sources", "superseded_by"},
	"contract":            {"provider", "superseded_by"},
	"standards":           {"superseded_by"},
	"system-blueprint":    {"contracts_consumed", "superseded_by"},
	"container-blueprint": {"contracts_consumed", "modules", "superseded_by", "system"},
	"module-blueprint":    {"contracts_consumed", "modules", "superseded_by"},
	"component-blueprint": {"contracts_consumed", "superseded_by"},
	"feature-blueprint":   {"frd", "superseded_by"},
	"guide":               {"superseded_by"},
}

// The profile registers every frontmatter property that holds a document
// slug, per type, so an adopting repository gets reference integrity over
// this framework's own vocabulary without authoring a profile.
//
// AC: AC-SPECREF-004.2
func TestTheProfileRegistersEverySlugHoldingProperty(t *testing.T) {
	types := profile.New().DocumentTypes()

	for _, value := range documentTypeVocabulary() {
		rules, ok := types[value]
		if !ok {
			t.Errorf("DocumentTypes has no entry for %q", value)
			continue
		}
		want := registeredReferenceProperties[value]
		if len(want) == 0 {
			t.Errorf("no expectation is recorded for %q; the table above is incomplete", value)
			continue
		}
		if strings.Join(rules.ReferenceProperties, ",") != strings.Join(want, ",") {
			t.Errorf("%q registers %v, want %v", value, rules.ReferenceProperties, want)
		}
	}

	// The identity property is not a reference: it names the document
	// declaring it, and registering it would make every document
	// reference itself.
	for value, rules := range types {
		for _, property := range rules.ReferenceProperties {
			if property == "slug" {
				t.Errorf("%q registers its own identity property as reference-bearing", value)
			}
		}
	}
}

// The registered set is derived from the vocabulary rather than
// transcribed from it, so a property added to schemas/frontmatter.strata
// is reference-bearing the moment it is declared. This is the same drift
// guard the per-type table already carries for document types.
func TestEveryCitationTypedPropertyInTheVocabularyIsRegistered(t *testing.T) {
	types := profile.New().DocumentTypes()
	seen := 0

	for _, value := range documentTypeVocabulary() {
		declared := citationTypedPropertiesOf(t, value)
		if len(declared) == 0 {
			t.Errorf("the generated schema declares no citation-typed property for %q; "+
				"every document type inherits superseded_by, so this is a misread schema", value)
			continue
		}
		registered := map[string]bool{}
		for _, property := range types[value].ReferenceProperties {
			registered[property] = true
		}
		for _, property := range declared {
			seen++
			if !registered[property] {
				t.Errorf("%q declares %q as a citation and does not register it as reference-bearing",
					value, property)
			}
		}
		// And nothing is registered that the vocabulary does not type as a
		// citation — a registration wider than the schema would resolve
		// values that are not references.
		for _, property := range types[value].ReferenceProperties {
			if !contains(declared, property) {
				t.Errorf("%q registers %q, which the generated schema does not type as a citation",
					value, property)
			}
		}
	}

	if seen == 0 {
		t.Fatal("read no citation-typed property from the generated schema; the guard would pass vacuously")
	}
}

// The profile registers no pairing, and this asserts that emptiness
// rather than leaving it unstated. Every relationship this framework
// declares has its reverse derived from inbound edges, so the pairing
// check is inert for this corpus by design — which is a property, not an
// omission.
func TestTheProfileRegistersNoPairing(t *testing.T) {
	for value, rules := range profile.New().DocumentTypes() {
		if len(rules.Pairings) != 0 {
			t.Errorf("%q registers pairings %+v; this framework declares none, because every "+
				"relationship it declares has its reverse derived from inbound edges",
				value, rules.Pairings)
		}
	}
}

// Exactly one typed reference is registered: a feature blueprint's `frd`
// must resolve to a document of type `frd`. The schema can require that
// the property holds one slug; only the corpus knows what that slug
// resolved to, so the type half of the constraint is registered here.
//
// AC: AC-SPECREF-002.10
func TestTheProfileRegistersOneTypedReference(t *testing.T) {
	types := profile.New().DocumentTypes()

	total := 0
	for _, rules := range types {
		total += len(rules.TypedReferenceProperties)
	}
	if total != 1 {
		t.Errorf("the profile registers %d typed references, want exactly 1", total)
	}

	blueprint := types[string(specmodel.DocumentTypeFeatureBlueprint)]
	if got := blueprint.TypedReferenceProperties["frd"]; got != string(specmodel.DocumentTypeFrd) {
		t.Errorf("a feature blueprint's frd must resolve to type %q, got %q",
			specmodel.DocumentTypeFrd, got)
	}
	// A typed reference over a property nothing extracts would be checked
	// against a target nothing resolved.
	if !contains(blueprint.ReferenceProperties, "frd") {
		t.Error("frd is registered as a typed reference and not as reference-bearing")
	}
}

// The lifecycle registers the whole status vocabulary, `retired` as its
// only retired status, every other status as live, and `superseded_by` as
// the successor property.
//
// AC: AC-SPECREF-004.4
func TestTheProfileRegistersTheLifecycle(t *testing.T) {
	rules := profile.New().Lifecycle()

	want := []string{"draft", "ready-for-review", "approved", "in-progress", "active", "retired"}
	if !equalStrings(rules.Statuses, want) {
		t.Errorf("Statuses = %v, want the whole vocabulary %v", rules.Statuses, want)
	}
	if !equalStrings(rules.RetiredStatuses, []string{string(specmodel.StatusRetired)}) {
		t.Errorf("RetiredStatuses = %v, want only retired", rules.RetiredStatuses)
	}
	if !equalStrings(rules.LiveStatuses, want[:len(want)-1]) {
		t.Errorf("LiveStatuses = %v, want every status but retired %v", rules.LiveStatuses, want[:len(want)-1])
	}
	if rules.SuccessorProperty != "superseded_by" {
		t.Errorf("SuccessorProperty = %q, want superseded_by", rules.SuccessorProperty)
	}

	// A draft is live on purpose: the retired-target warning exists to
	// tell an author what to cite instead, and the author most likely to
	// need telling is editing one.
	if !contains(rules.LiveStatuses, string(specmodel.StatusDraft)) {
		t.Error("draft is not registered as live")
	}
	// The successor property is one of the properties resolution follows,
	// so a successor naming no corpus member is already an unresolved
	// reference.
	for value, types := range profile.New().DocumentTypes() {
		if !contains(types.ReferenceProperties, rules.SuccessorProperty) {
			t.Errorf("%q does not register the successor property %q as reference-bearing",
				value, rules.SuccessorProperty)
		}
	}
}

// The host validates a registration when a profile is activated, and this
// one has nothing for it to report. A registration is data, and data can
// contradict itself in ways no compiler catches.
func TestTheRegistrationHasNoDefects(t *testing.T) {
	if defects := corpus.ValidateRegistration(profile.New()); len(defects) != 0 {
		t.Errorf("ValidateRegistration reported %d defects: %+v", len(defects), defects)
	}
}

// The identifier patterns this profile registers are the ones
// schemas/feature.strata fixes for the generated FeatureId and
// CriterionId aliases, widened only with the named capture groups
// SharedCaptures needs. Stripping those group names must recover the
// generated pattern exactly, so the registration cannot silently
// diverge from the model it mirrors.
//
// AC: AC-GHK-002.1, AC-GHK-002.3
func TestFeatureTagPatternsMatchTheGeneratedIdentifiers(t *testing.T) {
	tags := profile.New().FeatureTags()

	cases := []struct {
		definition, registered string
	}{
		{"FeatureId", tags.FeaturePattern},
		{"CriterionId", tags.CriterionPattern},
	}
	for _, tc := range cases {
		generated := generatedIdentifierPattern(t, tc.definition)
		if stripped := stripNamedCaptures(tc.registered); stripped != generated {
			t.Errorf("%s: registered pattern %q strips to %q, want the generated pattern %q",
				tc.definition, tc.registered, stripped, generated)
		}
	}
}

// The owning-document tag prefix is frd-, and the two identifier patterns
// share exactly the capture groups the derivation rule needs.
//
// AC: AC-GHK-002.1, AC-GHK-003.6
func TestFeatureTagsRegistersTheOwnerPrefixAndSharedCaptures(t *testing.T) {
	tags := profile.New().FeatureTags()

	if tags.OwnerPrefix != "frd-" {
		t.Errorf("OwnerPrefix = %q, want %q", tags.OwnerPrefix, "frd-")
	}
	if want := []string{"code", "number"}; !equalStrings(tags.SharedCaptures, want) {
		t.Errorf("SharedCaptures = %v, want %v", tags.SharedCaptures, want)
	}
}

// The profile registers negative and manual as their fixed meanings, and
// the level- prefixed vocabulary as the verification level each names —
// specs/standards/gherkin-guide.md's whole qualifier table, and nothing
// beyond it.
//
// AC: AC-GHK-002.1
func TestFeatureTagsRegistersTheQualifierVocabulary(t *testing.T) {
	tags := profile.New().FeatureTags()

	want := map[string]corpus.QualifierTag{
		"negative": {Tag: "negative", Meaning: specmodel.QualifierMeaningNegation},
		"manual":   {Tag: "manual", Meaning: specmodel.QualifierMeaningManualVerification},
		"level-unit": {Tag: "level-unit", Meaning: specmodel.QualifierMeaningVerificationLevel,
			Level: specmodel.VerificationLevelUnit},
		"level-integration": {Tag: "level-integration", Meaning: specmodel.QualifierMeaningVerificationLevel,
			Level: specmodel.VerificationLevelIntegration},
		"level-e2e": {Tag: "level-e2e", Meaning: specmodel.QualifierMeaningVerificationLevel,
			Level: specmodel.VerificationLevelEndToEnd},
	}
	if len(tags.Qualifiers) != len(want) {
		t.Errorf("Qualifiers has %d entries, want %d", len(tags.Qualifiers), len(want))
	}
	for _, qualifier := range tags.Qualifiers {
		expected, ok := want[qualifier.Tag]
		if !ok {
			t.Errorf("registers unexpected qualifier %q", qualifier.Tag)
			continue
		}
		if qualifier != expected {
			t.Errorf("qualifier %q = %+v, want %+v", qualifier.Tag, qualifier, expected)
		}
	}
}

// generatedIdentifierPattern reads one definition's `pattern` from the
// generated schema, walked independently of registration.go's own
// decoder — see this file's package comment.
func generatedIdentifierPattern(t *testing.T, definition string) string {
	t.Helper()

	var bundle map[string]any
	if err := json.Unmarshal(schema.SpecModel, &bundle); err != nil {
		t.Fatalf("reading the generated schema: %v", err)
	}
	definitions, ok := bundle["definitions"].(map[string]any)
	if !ok {
		t.Fatal("the generated schema has no definitions object")
	}
	branch, ok := definitions[definition].(map[string]any)
	if !ok {
		t.Fatalf("the generated schema declares no branch %q", definition)
	}
	pattern, ok := branch["pattern"].(string)
	if !ok || pattern == "" {
		t.Fatalf("the generated schema's %q branch declares no pattern", definition)
	}
	return pattern
}

// namedCaptureOpen matches the opening of one Go regexp named capture
// group — `(?P<code>` — and nothing of its content. The content is found
// by matching parens instead, because a code spanning hyphen-joined
// segments nests a non-capturing group for the repeated segment. It
// assumes no character class in either pattern holds an unescaped paren,
// which would need a class-aware scan rather than a paren count.
var namedCaptureOpen = regexp.MustCompile(`\(\?P<[A-Za-z0-9_]+>`)

// stripNamedCaptures removes named-capture-group wrapping from a
// pattern, recovering the plain expression underneath. Naming a capture
// group changes nothing a pattern matches, so this is the whole of what
// may distinguish a registered pattern from the generated one it
// mirrors. A pattern whose parens do not balance is returned unchanged,
// so the comparison fails on the pattern rather than here.
func stripNamedCaptures(pattern string) string {
	for {
		open := namedCaptureOpen.FindStringIndex(pattern)
		if open == nil {
			return pattern
		}
		close, ok := matchingParen(pattern, open[1])
		if !ok {
			return pattern
		}
		pattern = pattern[:open[0]] + pattern[open[1]:close] + pattern[close+1:]
	}
}

// matchingParen returns the index of the paren closing the group opened
// before start, counting nested parens and skipping escaped ones.
func matchingParen(pattern string, start int) (int, bool) {
	depth := 1
	for i := start; i < len(pattern); i++ {
		switch pattern[i] {
		case '\\':
			i++
		case '(':
			depth++
		case ')':
			if depth--; depth == 0 {
				return i, true
			}
		}
	}
	return 0, false
}

// citationTypedPropertiesOf returns the properties the generated schema
// types as a citation for one document type, read by walking the
// decoded JSON rather than through registration.go's own decoder — see
// this file's package comment.
func citationTypedPropertiesOf(t *testing.T, documentType string) []string {
	t.Helper()

	var bundle map[string]any
	if err := json.Unmarshal(schema.SpecModel, &bundle); err != nil {
		t.Fatalf("reading the generated schema: %v", err)
	}
	definitions, ok := bundle["definitions"].(map[string]any)
	if !ok {
		t.Fatal("the generated schema has no definitions object")
	}

	name := pascalCase(documentType)
	branch, ok := definitions[name].(map[string]any)
	if !ok {
		t.Fatalf("the generated schema declares no branch %q for document type %q", name, documentType)
	}
	properties, ok := branch["properties"].(map[string]any)
	if !ok {
		t.Fatalf("the %q branch declares no properties", name)
	}

	var found []string
	for property, declared := range properties {
		if property == "slug" {
			continue
		}
		shape, ok := declared.(map[string]any)
		if !ok {
			continue
		}
		if mentionsCitation(shape) {
			found = append(found, property)
		}
	}
	return found
}

// mentionsCitation reports whether a decoded property schema refers to
// the Citation definition anywhere in it — directly, through `items`, or
// through a `oneOf` against null.
func mentionsCitation(shape map[string]any) bool {
	const citationRef = "#/definitions/Citation"

	if ref, ok := shape["$ref"].(string); ok && ref == citationRef {
		return true
	}
	if items, ok := shape["items"].(map[string]any); ok && mentionsCitation(items) {
		return true
	}
	alternatives, ok := shape["oneOf"].([]any)
	if !ok {
		return false
	}
	for _, alternative := range alternatives {
		if branch, ok := alternative.(map[string]any); ok && mentionsCitation(branch) {
			return true
		}
	}
	return false
}

// pascalCase converts a document type value to the name the generated
// schema keys its branch by: "system-blueprint" becomes
// "SystemBlueprint".
func pascalCase(kebab string) string {
	var b strings.Builder
	for _, word := range strings.Split(kebab, "-") {
		if word == "" {
			continue
		}
		b.WriteString(strings.ToUpper(word[:1]))
		b.WriteString(word[1:])
	}
	return b.String()
}

func contains(haystack []string, needle string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// The component fence tag is registered, and its schema pointer resolves
// to the generated element record definition in the embedded schema.
func TestRecordBlocksRegistersTheComponentTagAgainstTheElementRecord(t *testing.T) {
	tags := profile.New().RecordBlocks()
	if len(tags) != 1 || tags[0].Tag != "component" {
		t.Fatalf("RecordBlocks = %+v, want exactly the component tag", tags)
	}
	const want = "#/definitions/C4ElementRecord"
	if tags[0].Schema != want {
		t.Fatalf("component schema pointer = %q, want %q", tags[0].Schema, want)
	}
	var bundle struct {
		Definitions map[string]json.RawMessage `json:"definitions"`
	}
	if err := json.Unmarshal(profile.New().Schema(), &bundle); err != nil {
		t.Fatalf("the embedded schema is not JSON: %v", err)
	}
	if _, ok := bundle.Definitions["C4ElementRecord"]; !ok {
		t.Errorf("the embedded schema declares no C4ElementRecord definition")
	}
}

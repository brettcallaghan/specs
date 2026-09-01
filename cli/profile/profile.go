// Package profile implements this framework's document vocabulary as
// one pluggable profile — [[module-spec-profile]]'s v1 shape — for
// Document Core's Profile extension point (cli/core/corpus).
//
// It holds no engine of its own: every capability it needs — frontmatter
// extraction, schema validation, span queries — comes from Document
// Core, and this package's only job is to register vocabulary into that
// host. There is no per-type schema here: one union schema (the
// generated cli/core/schema bundle) covers every document type, and this
// package supplies, per type, only the two facts that schema cannot
// express — a slug pattern and whether the type is numbered — plus the
// convention defaults Document Core resolves when an adopting
// repository's own configuration omits them.
//
// What this package registers per type is the authored table below —
// a slug pattern and whether the type is numbered — widened with the
// reference registration derived from the embedded schema, and the
// lifecycle derived from the same place. Both derivations live in
// registration.go, and they are derivations rather than lists because a
// property added to schemas/frontmatter.strata is reference-bearing the
// moment it is declared.
//
// Element records and required sections are later capabilities;
// registering them belongs to a future change to this package, not to a
// workaround here.
package profile

import (
	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/schema"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// Name is the name this profile registers under. An adopting repository
// writes it as `[profile].name` in specs.toml
// ([[contract-specs-configuration]]) to activate this profile without
// authoring one of its own, and the profile registry the CFG delta
// builds keys its built-in entry by this constant.
const Name = "spec-driven-dev"

// defaultLineWidth is the line_width convention this profile supplies
// when specs.toml omits [conventions].line_width. 80 is the limit
// specs/standards/template-authoring-standards.md fixes ("The width of
// any line SHALL NOT exceed 80 characters"), measured in Unicode
// codepoints per that same standard.
const defaultLineWidth = 80

// documentTypes is the authored half of the per-type table: a slug
// pattern and whether the type is numbered, for every value in the
// DocumentType vocabulary. It carries no schema —
// DocumentTypeRegistration.schema in cli/core/specmodel is left
// unpopulated, since the one union schema this profile embeds already
// covers every type. It carries no reference registration either: that
// half is derived from the same schema, in registration.go, and the two
// are joined by [registeredTypes] — which is what [New] returns.
//
// Patterns are grounded in this repository's own corpus rather than
// restated from the prefix registry in template-authoring-standards.md,
// because in two places the registry and this repository's actual
// documents disagree:
//
//   - Feature blueprints. The registry assigns the fb- prefix, but this
//     repository's one real feature blueprint
//     (specs/features/spec-index/spec-index-blueprint.md) is slugged
//     feature-spec-validation-cli. The pattern here follows the
//     document that exists, not the template that was not.
//   - Guides. The registry has no row for this type at all. Of this
//     repository's four guide documents, two are prefixed guide- and
//     two are not (work-order-field-guide.md, in two copies). The
//     pattern here is accordingly the type's least restrictive form:
//     any kebab-case slug.
//
// RFCs mix two identifier schemes in this repository: a numbered form
// (rfc-0007) and a self-critique form with no number at all
// (rfc-self-greenfield-anchors-and-split-numbering). Because a
// meaningful share of real RFC slugs carry no number, Numbered is false
// here — a "highest number in use" query would silently ignore every
// self- document, which is worse than not offering next-id for the
// type. rfcs/inbox/groom-feature-blueprint/rfc-self-delta-keys-and-tracker-vocabulary.md
// tracks this split scheme as an open concern in its own right.
// featureIDPattern and criterionIDPattern are this framework's Gherkin
// identifier patterns (specs/standards/gherkin-guide.md, AC-GHK-002.1,
// AC-GHK-002.3).
//
// Each is the pattern schemas/feature.strata fixes for the generated
// FeatureId and CriterionId aliases —
// `^[A-Z]+-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*-[0-9]{3}$` and
// `^AC-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*-[0-9]{3}\.[0-9]+$` —
// widened with two named capture groups, code and number, so the
// SharedCaptures in [featureTags] can state the derivation rule that a
// criterion identifier's code and number are its feature's (AC-GHK-003.6)
// without the host reading either pattern's internal structure. Naming a
// capture group changes nothing a pattern matches, so this widening does
// not diverge from the generated pattern it mirrors —
// TestFeatureTagPatternsMatchTheGeneratedIdentifiers in registration_test.go
// checks the two against each other directly.
const (
	featureIDPattern   = `^[A-Z]+-(?P<code>[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*)-(?P<number>[0-9]{3})$`
	criterionIDPattern = `^AC-(?P<code>[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*)-(?P<number>[0-9]{3})\.[0-9]+$`
)

// featureTags is this profile's tag vocabulary: the whole of
// [[frd-gherkin-adoption]]'s AC-GHK-002.1 table and nothing beyond it.
//
//   - FeaturePattern and CriterionPattern are [featureIDPattern] and
//     [criterionIDPattern].
//   - OwnerPrefix is `frd-`, the prefix on the tag in which a feature
//     written as a standalone `.feature` file names the document that
//     owns it — `@frd-<slug>`.
//   - SharedCaptures is `code` and `number`, the capture groups the two
//     identifier patterns share.
//   - Qualifiers is `negative`, `manual`, and the `level-` prefixed
//     vocabulary `unit`, `integration` and `e2e` — nothing beyond that.
var featureTags = corpus.FeatureTagRules{
	FeaturePattern:   featureIDPattern,
	CriterionPattern: criterionIDPattern,
	OwnerPrefix:      "frd-",
	SharedCaptures:   []string{"code", "number"},
	Qualifiers: []corpus.QualifierTag{
		{Tag: "negative", Meaning: specmodel.QualifierMeaningNegation},
		{Tag: "manual", Meaning: specmodel.QualifierMeaningManualVerification},
		{Tag: "level-unit", Meaning: specmodel.QualifierMeaningVerificationLevel, Level: specmodel.VerificationLevelUnit},
		{Tag: "level-integration", Meaning: specmodel.QualifierMeaningVerificationLevel, Level: specmodel.VerificationLevelIntegration},
		{Tag: "level-e2e", Meaning: specmodel.QualifierMeaningVerificationLevel, Level: specmodel.VerificationLevelEndToEnd},
	},
}

var documentTypes = map[string]corpus.DocumentTypeRules{
	string(specmodel.DocumentTypeAdr): {
		SlugPattern: `^adr-[0-9]{3,}(-[a-z0-9]+(-[a-z0-9]+)*)?$`,
		Numbered:    true,
	},
	string(specmodel.DocumentTypeFrd): {
		SlugPattern: `^frd-[a-z0-9]+(-[a-z0-9]+)*$`,
	},
	string(specmodel.DocumentTypeContract): {
		SlugPattern: `^contract-[a-z0-9]+(-[a-z0-9]+)*$`,
	},
	string(specmodel.DocumentTypeStandards): {
		// No fixed prefix, per the registry's own stated exception for
		// this type — and, in practice, no fixed suffix either: several
		// real standards documents in this repository (
		// blueprint-readiness-checklist, findings-capture-protocol,
		// work-order-definition-of-ready) do not end in one of the
		// registry's approved role nouns. Any kebab-case slug is
		// accepted rather than enforcing a suffix rule this repository's
		// own documents do not follow.
		SlugPattern: `^[a-z0-9]+(-[a-z0-9]+)*$`,
	},
	string(specmodel.DocumentTypeSystemBlueprint): {
		SlugPattern: `^system-[a-z0-9]+(-[a-z0-9]+)*$`,
	},
	string(specmodel.DocumentTypeContainerBlueprint): {
		SlugPattern: `^container-[a-z0-9]+(-[a-z0-9]+)*$`,
	},
	string(specmodel.DocumentTypeModuleBlueprint): {
		SlugPattern: `^module-[a-z0-9]+(-[a-z0-9]+)*$`,
	},
	string(specmodel.DocumentTypeComponentBlueprint): {
		SlugPattern: `^component-[a-z0-9]+(-[a-z0-9]+)*$`,
	},
	string(specmodel.DocumentTypeFeatureBlueprint): {
		SlugPattern: `^feature-[a-z0-9]+(-[a-z0-9]+)*$`,
	},
	string(specmodel.DocumentTypeRfc): {
		SlugPattern: `^rfc-([0-9]{4}(-[a-z0-9]+(-[a-z0-9]+)*)?|self-[a-z0-9]+(-[a-z0-9]+)*)$`,
	},
	string(specmodel.DocumentTypeGuide): {
		SlugPattern: `^[a-z0-9]+(-[a-z0-9]+)*$`,
	},
}

// specDrivenDev is this framework's Profile: the value New returns.
type specDrivenDev struct{}

// registration is the whole per-type table this profile registers,
// assembled once. It is built at package initialization rather than per
// call because it is a constant of this binary: the authored table is a
// literal and the derived half comes from a schema embedded at build
// time, so there is nothing a later call could learn.
var registration = registeredTypes()

// New returns this framework's spec-driven-dev profile: a value
// satisfying Document Core's Profile interface (cli/core/corpus),
// carrying the embedded union schema, the per-type table of slug
// pattern, numbered-ness and reference registration, the lifecycle, and
// the line_width convention default.
func New() corpus.Profile {
	return specDrivenDev{}
}

// Schema returns the union schema every document type this profile
// defines is validated against: cli/core/schema/spec-model.schema.json,
// embedded at build time via //go:embed in cli/core/schema. The binary
// reads no schema file at runtime.
func (specDrivenDev) Schema() []byte {
	return schema.SpecModel
}

// DocumentTypes returns what this profile registers for every
// DocumentType: its slug pattern, whether it is numbered, the frontmatter
// properties of that type that hold a document slug, the one typed
// reference this framework declares, and its pairings — of which it
// declares none.
func (specDrivenDev) DocumentTypes() map[string]corpus.DocumentTypeRules {
	return registration
}

// Conventions returns this profile's convention defaults: line_width,
// resolved when specs.toml omits [conventions].line_width.
func (specDrivenDev) Conventions() specmodel.ConventionsConfig {
	width := uint64(defaultLineWidth)
	return specmodel.ConventionsConfig{LineWidth: &width}
}

// Lifecycle returns this profile's lifecycle registration: the whole
// status vocabulary the generated schema declares, `retired` as the only
// status at which a document is no longer in effect, every other status
// as live, and `superseded_by` as the property a retired document names
// its replacement in.
//
// Every status other than `retired` is live because the warning about a
// retired target exists to tell an author what to cite instead, and the
// author most likely to need telling is editing a draft.
func (specDrivenDev) Lifecycle() corpus.LifecycleRules {
	return lifecycle()
}

// FeatureTags returns this profile's tag vocabulary: the feature and
// criterion identifier patterns, the owning-document tag prefix, the
// derivation rule linking a criterion's code and number to its feature's,
// and the qualifier tags this framework assigns meaning to. See
// [featureTags].
func (specDrivenDev) FeatureTags() corpus.FeatureTagRules {
	return featureTags
}

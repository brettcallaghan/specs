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

// FeatureTags returns this profile's tag vocabulary, which is empty: no
// tag on a feature or a criterion carries meaning here yet. Registering
// this framework's identifier patterns and qualifier tags belongs to the
// change that starts reading them, not to the widening of the extension
// point that made them expressible.
func (specDrivenDev) FeatureTags() corpus.FeatureTagRules {
	return corpus.FeatureTagRules{}
}

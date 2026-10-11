package profile

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/brettcallaghan/specs/cli/core/corpus"
	"github.com/brettcallaghan/specs/cli/core/schema"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// This file holds the half of the registration that is *derived* rather
// than authored: which frontmatter properties hold a reference to
// another document, and which statuses this framework's documents move
// through. Both are facts the vocabulary already states —
// `schemas/frontmatter.strata` types a reference-bearing property as
// `Citation` or `[Citation]`, and `schemas/vocabulary.strata` enumerates
// `Status` — and both reach this package through the same generated
// schema the profile already embeds.
//
// Deriving them rather than transcribing them is the point. A property
// added to the vocabulary is reference-bearing the moment it is declared,
// with no second list to remember; a transcribed one would be correct on
// the day it was written and silently incomplete afterwards, which is the
// failure a registration cannot report about itself.

// identityProperty is the one property that names the document declaring
// it rather than another document. It is typed `Slug` — never `Citation`,
// since a document's own identity must never carry an anchor — and
// registering it as reference-bearing would make every document
// reference itself, so it is the single exclusion from the derivation
// below.
const identityProperty = "slug"

// citationDefinition is the JSON Pointer the generated schema gives the
// Citation alias. A property whose value is that definition, or a list
// of it, holds a reference to another document, optionally qualified by
// an anchor.
const citationDefinition = "#/definitions/Citation"

// statusDefinition names the enum the status vocabulary is declared in.
const statusDefinition = "Status"

// retiredStatus is the one status at which a document of this framework
// is no longer in effect. Every other status in the vocabulary is live —
// see [lifecycle] for why a draft counts.
const retiredStatus = string(specmodel.StatusRetired)

// successorProperty is the frontmatter property a retired document names
// its replacement in. Registered so that a warning about a retired target
// can quote the replacement rather than only reporting the retirement;
// nothing requires a retired document to declare one.
const successorProperty = "superseded_by"

// typedReferences is the whole of this profile's typed-reference
// registration: a feature blueprint's `frd` must resolve to a document of
// type `frd` (AC-SPECREF-002.10).
//
// It is authored rather than derived, because it is the one constraint
// the vocabulary cannot state. `FeatureBlueprint.frd` is declared as a
// `Citation`, which says the value names another document, optionally
// by anchor; that the document it names must be an FRD is a fact about
// the corpus, and only the index knows what a citation's slug resolved
// to.
//
// Keyed by document type, then by property, mirroring the shape
// [corpus.DocumentTypeRules.TypedReferenceProperties] takes.
var typedReferences = map[string]map[string]string{
	string(specmodel.DocumentTypeFeatureBlueprint): {
		"frd": string(specmodel.DocumentTypeFrd),
	},
}

// pairings is this profile's pairing registration, and it is empty by
// decision rather than by omission.
//
// Every relationship this framework declares has its reverse derived from
// inbound edges — a container names its system, a module names its hosts,
// and the roster at the other end is a query over the index rather than a
// property anyone maintains. `document-relationship-standards.md` reserves
// a pairing for the relationship a reader must see without the index, and
// this vocabulary has none. The pairing check is therefore inert for this
// corpus by design, which is a property to assert rather than an omission
// to explain.
var pairings = map[string][]specmodel.PropertyPairing{}

// registeredTypes returns the per-type table the profile hands the host:
// the authored slug pattern and numbered flag from [documentTypes],
// widened with the reference registration derived from the embedded
// schema.
//
// It panics on a document type the embedded schema declares no branch
// for, as [decodedSchema] panics on a schema it cannot read. Both are
// faults in a generated artifact compiled into the binary —
// `scripts/generate-schemas.sh` produces it and `--check` gates it —
// rather than conditions a run can meet, and a profile that quietly
// registered nothing would be a registration that checks less than it
// says: the vacuous pass [corpus.ValidateRegistration] exists to prevent.
func registeredTypes() map[string]corpus.DocumentTypeRules {
	citationProperties := citationPropertiesByType()

	types := make(map[string]corpus.DocumentTypeRules, len(documentTypes))
	for name, rules := range documentTypes {
		properties, declared := citationProperties[name]
		if !declared {
			panic(fmt.Sprintf(
				"profile: the embedded schema declares no branch for document type %q", name))
		}
		rules.ReferenceProperties = properties
		rules.TypedReferenceProperties = typedReferences[name]
		rules.Pairings = pairings[name]
		types[name] = rules
	}
	return types
}

// lifecycle returns this profile's lifecycle registration: the whole
// status vocabulary, `retired` as its only retired status, every other
// status as live, and `superseded_by` as the successor property.
//
// Every status other than `retired` is live because the retired-target
// warning exists to tell an author what to cite instead, and the author
// most likely to need telling is editing a draft. Live is therefore
// derived as the vocabulary minus the retired set rather than listed, so
// a status added to the vocabulary is live without this function being
// revisited.
func lifecycle() corpus.LifecycleRules {
	statuses := statusVocabulary()

	rules := corpus.LifecycleRules{
		Statuses:          statuses,
		RetiredStatuses:   []string{retiredStatus},
		SuccessorProperty: successorProperty,
	}
	for _, status := range statuses {
		if status != retiredStatus {
			rules.LiveStatuses = append(rules.LiveStatuses, status)
		}
	}
	return rules
}

// schemaProperty is one property of a document type's branch, decoded
// only as far as telling a citation from anything else needs: a direct
// `$ref` for a required value, a `oneOf` against null for an optional
// one, and `items` for a list.
type schemaProperty struct {
	Ref   string `json:"$ref"`
	OneOf []struct {
		Ref string `json:"$ref"`
	} `json:"oneOf"`
	Items struct {
		Ref string `json:"$ref"`
	} `json:"items"`
}

// holdsCitation reports whether this property's value is a citation, in
// any of the three forms the transpiler emits.
func (p schemaProperty) holdsCitation() bool {
	if p.Ref == citationDefinition || p.Items.Ref == citationDefinition {
		return true
	}
	for _, alternative := range p.OneOf {
		if alternative.Ref == citationDefinition {
			return true
		}
	}
	return false
}

// schemaBundle is the embedded union schema decoded as far as this
// package reads it: the values of an enum, and the properties of a
// document type's branch.
//
// The decode is deliberately partial. What the registration needs from
// the schema is two facts, and modelling the rest would couple it to a
// JSON Schema shape it does not care about.
type schemaBundle struct {
	Definitions map[string]struct {
		Enum       []string                  `json:"enum"`
		Properties map[string]schemaProperty `json:"properties"`
	} `json:"definitions"`
}

// decodedSchema is the embedded union schema, decoded once. The registered
// set is read from it at package initialization and never changes, so
// decoding it more than once would only be a way for two readings to
// disagree.
var decodedSchema = sync.OnceValue(func() schemaBundle {
	var bundle schemaBundle
	if err := json.Unmarshal(schema.SpecModel, &bundle); err != nil {
		panic(fmt.Sprintf("profile: the embedded schema is not readable: %v", err))
	}
	if len(bundle.Definitions) == 0 {
		panic("profile: the embedded schema declares no definitions")
	}
	return bundle
})

// citationPropertiesByType returns, for every document type the schema
// declares a branch for, the frontmatter properties of that branch whose
// value is a citation, excluding the document's own identity property.
//
// The result is keyed by the value documents declare in `type`, and each
// list is sorted, so what the profile hands the host is the same list on
// every run rather than a map iteration order.
func citationPropertiesByType() map[string][]string {
	definitions := decodedSchema().Definitions

	byType := make(map[string][]string, len(definitions))
	for name, definition := range definitions {
		var properties []string
		for property, declared := range definition.Properties {
			if property == identityProperty || !declared.holdsCitation() {
				continue
			}
			properties = append(properties, property)
		}
		sort.Strings(properties)
		byType[kebabCase(name)] = properties
	}
	return byType
}

// statusVocabulary returns every value the status enum declares, in the
// order the schema lists them — which is the order the vocabulary
// declares them in, so a registration built from it reads the way its
// source does.
func statusVocabulary() []string {
	values := decodedSchema().Definitions[statusDefinition].Enum
	if len(values) == 0 {
		panic("profile: the embedded schema declares no status vocabulary")
	}
	return values
}

// kebabCase converts a schema definition name to the value documents
// declare in their type property: "SystemBlueprint" becomes
// "system-blueprint".
//
// It is the inverse of the transform the Schema Registry applies in the
// other direction, and it is a mechanical string transform rather than a
// lookup table naming any type — so a document type added to the
// vocabulary is keyed correctly without this function being touched. A
// definition that is not a document type at all ("ElementRecord",
// "SpecIndex") is converted too and simply never asked for, since
// [registeredTypes] reads only the keys [documentTypes] declares.
func kebabCase(pascal string) string {
	var b strings.Builder
	for i, r := range pascal {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteRune('-')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

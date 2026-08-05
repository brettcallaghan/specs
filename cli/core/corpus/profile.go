// Package corpus holds the Profile extension point — the seam through
// which all domain knowledge enters Document Core — and, from the CO
// delta onward, corpus discovery and index assembly.
//
// Document Core is the general half of the CLI: everything true of any
// collection of markdown documents that carry YAML frontmatter. It does
// not know what document types exist, how their identifiers are spelled,
// or which of them are numbered. Those facts arrive at startup through
// [Profile], supplied by the caller, and nothing in this module may
// import or special-case a profile that provides them.
package corpus

import "github.com/brettcallaghan/agent-plugins/cli/core/specmodel"

// Profile supplies the document vocabulary one corpus operation is
// checked against.
//
// This is the plugin architecture: Document Core is the host, a profile
// is the plugin, and no vocabulary is compiled into the host. Exactly one
// profile is active per operation, and the caller chooses which.
//
// The surface is deliberately narrow — a schema, a per-type table, and
// the convention defaults — because it constrains everything downstream
// of it. Required sections, element records, reference properties and
// lifecycle rules are later capabilities, and each will widen this
// interface. Per the Change Guidance in
// specs/architecture/module-spec-profile.md, that widening is a change to
// the host reviewed on its own merits, never a special case granted to
// one profile.
type Profile interface {
	// Schema returns the profile's JSON Schema: one discriminated union
	// over every document type the profile defines, against which the
	// frontmatter of any of them is validated.
	//
	// It is one schema, not a registry of schemas keyed by type. A
	// profile that needed a schema per type would be describing several
	// vocabularies rather than one, and the host would have to know which
	// to reach for — knowledge it must not hold.
	//
	// The returned bytes are the profile's own; a caller that modifies
	// them corrupts every later operation, so callers treat them as
	// read-only.
	Schema() []byte

	// DocumentTypes returns the rules for each document type the profile
	// defines, keyed by the value documents declare in their type
	// property.
	//
	// The keys are the whole of the host's knowledge of which types
	// exist: a document declaring a key absent from this map is recorded
	// as excluded with reason unknown-type, and its declared value
	// reported, rather than being checked against the wrong rules.
	DocumentTypes() map[string]DocumentTypeRules

	// Conventions returns the house-convention values the profile
	// supplies where a repository's configuration omits them.
	//
	// Resolution runs configuration first, then this: an adopting
	// repository overrides a convention, and a profile provides the
	// default it did not set.
	Conventions() specmodel.ConventionsConfig
}

// DocumentTypeRules is everything Document Core needs to know about one
// document type.
//
// It is a thin table by design, and it carries no schema. Validation runs
// against the single union schema [Profile.Schema] returns; these are the
// two facts that schema cannot express — a slug pattern, because the
// shared frontmatter envelope declares slug once and a subtype cannot
// narrow an inherited property, and whether the type is numbered, which
// is a query over the corpus rather than a property of a document.
type DocumentTypeRules struct {
	// SlugPattern is the regular expression a document of this type must
	// match in its slug property.
	//
	// A mismatch is reported naming both the actual slug and this
	// pattern, so the finding is self-contained and the agent fixing it
	// needs no further lookup.
	SlugPattern string

	// Numbered reports whether documents of this type carry a sequence
	// number the next-identifier query derives from.
	//
	// Numbering is global per type: the next number is one greater than
	// the highest in use anywhere in the corpus, regardless of directory.
	Numbered bool
}

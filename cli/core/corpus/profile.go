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
// The surface is deliberately narrow — a schema, a per-type table, the
// convention defaults, and the lifecycle registration — because it
// constrains everything downstream of it. Required sections and element
// records are later capabilities, and each will widen this interface
// further. Per the Change Guidance in
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

	// Lifecycle returns the statuses this profile's documents move
	// through and what those statuses mean: which of them a document is
	// in effect at, which of them it is no longer in effect at, and the
	// property in which a document at one of the latter names its
	// replacement.
	//
	// Profile-level rather than per-type, because status is a property of
	// the shared document envelope: every type a profile defines carries
	// it and none of them narrows it.
	//
	// A profile that registers no lifecycle returns the zero value, and
	// the checks that read it do not run. That is a profile describing a
	// corpus whose documents have no lifecycle, not a malformed one.
	Lifecycle() LifecycleRules
}

// LifecycleRules is what a profile tells the host about the lifecycle
// its documents move through.
//
// The host holds none of these values. It does not know that a status is
// spelled "retired" or that a replacement is named in "superseded_by" —
// it knows only that some statuses mean in-effect, some mean superseded,
// and one property carries the replacement. An adopting profile whose
// documents are "current" and "sunset" gets the same checks.
type LifecycleRules struct {
	// Statuses is the whole status vocabulary this profile's documents
	// may declare.
	//
	// It is stated rather than inferred from the two sets below, for two
	// reasons. The union of Live and Retired is not the vocabulary — a
	// status in neither is one the lifecycle says nothing about, which is
	// the right answer for a draft — so the vocabulary cannot be
	// recovered from them. And [ValidateRegistration] can only report a
	// lifecycle registering a status the profile does not define if it is
	// told which statuses the profile does define; deriving that from the
	// union schema instead would mean the host knowing where a schema
	// keeps its status enum, and would pass vacuously against a profile
	// that keeps it somewhere else — the failure this validation exists
	// to prevent.
	Statuses []string

	// LiveStatuses are the statuses at which a document is in effect, and
	// therefore at which a reference it makes to a superseded document is
	// worth reporting.
	LiveStatuses []string

	// RetiredStatuses are the statuses at which a document is no longer
	// in effect, so that a live document still pointing at it is
	// reported.
	RetiredStatuses []string

	// SuccessorProperty is the name of the frontmatter property in which
	// a document at a retired status names its replacement.
	//
	// Registered rather than assumed, because the engine that reports a
	// retired target quotes the replacement back to the reader and must
	// not contain the name of the property one framework happens to use
	// for it. Empty where the profile registers none; nothing requires a
	// retired document to declare a successor.
	SuccessorProperty string
}

// DocumentTypeRules is everything Document Core needs to know about one
// document type.
//
// It is a thin table by design, and it carries no schema. Validation runs
// against the single union schema [Profile.Schema] returns; these are the
// facts that schema cannot express — a slug pattern, because the shared
// frontmatter envelope declares slug once and a subtype cannot narrow an
// inherited property; whether the type is numbered, which is a query over
// the corpus rather than a property of a document; and which of the
// type's properties hold references, because what a value resolves to is
// a question about the corpus rather than about the value.
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

	// ReferenceProperties are the frontmatter properties of this type
	// that hold document slugs.
	//
	// A value in one of them is recorded as an outbound reference of the
	// document, naming the property it came from, and is resolved and
	// reported exactly as a reference written in the body is. A property
	// absent from this list contributes no reference however slug-shaped
	// its value looks: what is a reference is registered, never guessed.
	ReferenceProperties []string

	// TypedReferenceProperties are the reference-bearing properties whose
	// target must declare a particular document type, keyed by property
	// name with the required type as the value. The values are keys of
	// [Profile.DocumentTypes].
	//
	// A schema can require that a property holds exactly one slug; only
	// the corpus knows what that slug resolves to, so the type half of
	// the constraint is registered here. Every key must also appear in
	// ReferenceProperties — a typed property that is not reference-
	// bearing would be checked against a target nothing resolved, which
	// is the vacuous success [ValidateRegistration] reports rather than
	// runs.
	TypedReferenceProperties map[string]string

	// Pairings are the relationship properties of this type that must be
	// declared from both ends: a document naming a counterpart in one of
	// them, where the target declares nothing in return, is a half-built
	// edge reported against both documents.
	//
	// Each pairing's Property is a property of this type and must appear
	// in ReferenceProperties. Its Counterpart is a property of whatever
	// type the reference resolves to, which is not known until it
	// resolves, so it is not checked here.
	Pairings []specmodel.PropertyPairing
}

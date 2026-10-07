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
// convention defaults, the lifecycle registration, the tag
// vocabulary, the record-block tags and the architecture vocabulary —
// because it constrains everything downstream of it.
// Required sections are a later capability, and will widen this
// interface further. Per the Change Guidance in
// specs/modules/spec-profile/spec-profile-module.md, that widening is a change to
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

	// FeatureTags returns the tag vocabulary this profile assigns
	// meaning to: which tag identifies a feature, which identifies a
	// criterion, which prefix names a feature's owning document, how a
	// criterion identifier derives from its feature's, and which
	// qualifier tags carry meaning.
	//
	// Profile-level rather than per-type, because a feature is not a
	// document type: it is written inside a document of any type, or in
	// a file that is no document at all, and the tags that identify it
	// are the same in both.
	//
	// Registered as data the host reads rather than as a rule the host
	// applies, exactly as the lifecycle is. The host compiles the
	// patterns, matches tags against them, and reports what it finds; it
	// holds no tag spelling of its own and no knowledge of what any
	// pattern's internals mean.
	//
	// A profile that registers none returns the zero value, and the
	// checks that read it do not run. That is a profile describing a
	// corpus with no tagged features, not a malformed one.
	FeatureTags() FeatureTagRules

	// RecordBlocks returns the fenced-block info strings this profile
	// registers as record blocks, each with an optional pointer to the
	// definition in [Profile.Schema] its record is validated against.
	//
	// A fenced block is a record block only where the active profile
	// says its info string is one; the host holds no tag spelling and no
	// knowledge of what any record means. A record is YAML, so there is
	// no per-tag content format to register.
	//
	// Registered as data the host reads, as the lifecycle and the tag
	// vocabulary are. A profile that registers none returns nil, and a
	// corpus built under it has no record blocks. That is a profile
	// describing documents with none, not a malformed one.
	RecordBlocks() []RecordBlockTag

	// Architecture returns the vocabulary through which this profile's
	// documents declare C4 elements: which document types declare one and
	// at what level, the frontmatter properties naming an element and its
	// containment and contract edges, and where a parent declares its
	// children.
	//
	// Registered as data the host reads, as the tag vocabulary is. The
	// host knows the C4 notation — the four levels and what a diagram
	// statement says — and holds no document type, property or heading
	// of its own.
	//
	// A profile that registers none returns the zero value, and a corpus
	// built under it has no elements. That is a profile describing
	// documents that declare no architecture, not a malformed one.
	Architecture() ArchitectureRules
}

// ElementLevel is the C4 level at which a registered document type
// declares an element.
//
// The four levels are C4 notation rather than any profile's vocabulary,
// so the host holds them; a profile chooses which of its types sits at
// each.
type ElementLevel string

// The four levels a registered element type may declare, outermost
// first.
const (
	LevelSystem    ElementLevel = "system"
	LevelContainer ElementLevel = "container"
	LevelModule    ElementLevel = "module"
	LevelComponent ElementLevel = "component"
)

// Valid reports whether level is one of the four C4 levels.
func (level ElementLevel) Valid() bool {
	switch level {
	case LevelSystem, LevelContainer, LevelModule, LevelComponent:
		return true
	}
	return false
}

// ArchitectureRules is what a profile tells the host about how its
// documents declare C4 elements.
//
// The host holds none of these values. It does not know which document
// type declares a container, what property names an element, or under
// which heading a parent lists its children — it knows only that some
// types declare elements at some level, that some properties carry an
// element's name and edges, and that some heading and record-block tag
// declare a child.
//
// Each property applies to a fixed set of the registered types, and
// [ValidateRegistration] checks that the profile's union schema declares
// it on every one of them: EntityNameProperty and
// ConsumedContractsProperty on every element type, SystemProperty on the
// container-level types, ComposedModulesProperty on the container- and
// module-level types, and ProviderProperty on ContractType. A property
// left empty is not registered and is checked nowhere.
type ArchitectureRules struct {
	// Elements are the document types that declare an element, keyed by
	// the value documents declare in their type property, as
	// [Profile.DocumentTypes] is keyed. Every key must be a type the
	// profile defines.
	Elements map[string]ElementTypeRules

	// EntityNameProperty is the frontmatter property naming the element
	// a document declares — the name a diagram draws it under.
	EntityNameProperty string

	// SystemProperty is the frontmatter property in which a
	// container-level document names the system it belongs to.
	SystemProperty string

	// ComposedModulesProperty is the frontmatter property in which a
	// container- or module-level document lists the modules composed
	// into it.
	ComposedModulesProperty string

	// ConsumedContractsProperty is the frontmatter property listing the
	// contracts an element consumes.
	ConsumedContractsProperty string

	// ContractType is the document type that declares a contract. It
	// must be a type the profile defines; empty registers no contracts.
	ContractType string

	// ProviderProperty is the frontmatter property in which a contract
	// names the element that provides it.
	ProviderProperty string
}

// ElementTypeRules is what a profile tells the host about one document
// type that declares an element.
type ElementTypeRules struct {
	// Level is the C4 level the type declares an element at.
	Level ElementLevel

	// RosterHeading is the text of the heading under which a document of
	// this type declares its children. Empty, with RosterTag, for a type
	// that holds no children.
	RosterHeading string

	// RosterTag is the record-block tag that declares one child under
	// the roster heading. It must be one of [Profile.RecordBlocks]'s
	// tags.
	RosterTag string
}

// RecordBlockTag is one fenced-block info string a profile registers as
// a record block, and where its record is validated.
type RecordBlockTag struct {
	// Tag is the info string that marks a fenced block as a record
	// block of this kind, as written after the opening fence.
	Tag string

	// Schema is an optional JSON Pointer URI fragment, such as
	// "#/definitions/Name", addressing the definition inside
	// [Profile.Schema] that a record under this tag is validated
	// against.
	//
	// It addresses the profile's own union schema rather than a second
	// schema, so the host never holds more than one. Empty means the
	// tag is registered with no schema and its records are not
	// validated. A non-empty pointer that resolves to nothing is a
	// registration defect, reported by [ValidateRegistration] and never
	// treated as empty.
	Schema string
}

// FeatureTagRules is what a profile tells the host about the tags that
// identify a feature and its criteria, and about the tags that qualify
// a criterion.
//
// The host holds none of these values. It does not know that a criterion
// identifier begins with any particular letters, that an owning document
// is named in a tag prefixed one way rather than another, or that a
// criterion is marked by hand-verification with one spelling rather than
// another — it knows only that some tag matches an identifier pattern,
// some tag carries an owner, and some tags populate a criterion's
// qualifier properties.
//
// Every member is independent, and a profile may register some and not
// others. A partial registration is honest rather than defective: a
// profile that identifies features and not criteria describes a corpus
// with no criteria in it, and only [ValidateRegistration]'s two checks —
// a pattern that does not compile, and a shared capture no pattern
// defines — say a registration contradicts itself.
type FeatureTagRules struct {
	// FeaturePattern is the regular expression the identifier of a
	// feature matches, without the leading marker that makes it a tag.
	//
	// The marker is excluded because the identifier is the value and the
	// tag is only how it is written down: the same identifier appears in
	// a citation and in a report, neither of which is a tag.
	FeaturePattern string

	// CriterionPattern is the regular expression the identifier of an
	// acceptance criterion matches, without the leading marker.
	CriterionPattern string

	// OwnerPrefix is the prefix carried by the tag in which a feature
	// names the document that owns it.
	//
	// A feature written outside any document has no frontmatter to
	// inherit an owner from and names one in a tag instead; this prefix
	// is the whole of how that tag is told apart from every other tag on
	// the feature.
	OwnerPrefix string

	// SharedCaptures are the named capture groups a criterion identifier
	// shares with the identifier of the feature it belongs to.
	//
	// This is how the derivation of a criterion identifier from its
	// feature's is stated without the host knowing the internal
	// structure of either pattern. The host compiles both, captures both
	// identifiers, and compares only the groups named here; a profile
	// whose two patterns share nothing registers none and no derivation
	// is checked.
	SharedCaptures []string

	// Qualifiers are the tags that qualify a criterion, and what each of
	// them means.
	//
	// A tag matching neither identifier pattern and named by no
	// qualifier is carried on the record and reported by nothing: an
	// unregistered tag is one the vocabulary says nothing about.
	Qualifiers []QualifierTag
}

// QualifierTag is one tag a profile assigns meaning to on a criterion,
// and the meaning it assigns.
type QualifierTag struct {
	// Tag is the tag itself, without the leading marker.
	Tag string

	// Meaning is the criterion property this tag populates.
	//
	// The set of meanings is fixed by the criterion record rather than by
	// any profile — a qualifier can only mean one of the things a
	// criterion has somewhere to put — so it is the generated
	// vocabulary's own type. What a profile chooses is the spelling of
	// each tag and which meaning it carries.
	Meaning specmodel.QualifierMeaning

	// Level is the verification level this tag records, for a tag whose
	// Meaning is the verification level and for no other.
	Level specmodel.VerificationLevel
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

# Profile 

## Table of Contents

- [ProfileRegistration](#profileregistration)
- [FeatureTagRegistration](#featuretagregistration)
- [QualifierTagRegistration](#qualifiertagregistration)
- [QualifierMeaning](#qualifiermeaning)
- [DocumentTypeRegistration](#documenttyperegistration)
- [Section](#section)
- [PropertyPairing](#propertypairing)
- [BlockTagRegistration](#blocktagregistration)

## ProfileRegistration

What a profile registers into Document Core.

**Internal surface.** This is the vocabulary a profile hands to the
engine in-process, not something an adopting repository authors. It is
not reachable from `specs.toml` in v1 and is not an extension point
outside this repository; `[profile].name` selects one of these
registrations by name, and that naming is the whole of the relationship
between the two halves of the configuration contract.

It is modelled here rather than left to hand-written Go for two
reasons. Registration is data — type names, patterns, section trees,
block tags — and only the rules that cannot be expressed as data stay in
code. And the open question of which internal keys get promoted to the
public surface is answered by editing this struct, which is a smaller
and more reviewable act than editing an interface and its
implementations.

Property names are snake_case throughout, matching the public surface,
so that promoting a key to `specs.toml` is a move rather than a rename.

| Property | Type | Documentation |
|----------|------|---------------|
| name | String | The name an adopting repository writes in `[profile].name`. |
| document_types | Map<String, [DocumentTypeRegistration](#documenttyperegistration)> | The document types this profile defines, keyed by the value documents declare in their `type` property.<br><br>A markdown file declaring a type absent from this map is recorded as excluded with reason `unknown-type` and its declared value reported, so a reviewing agent can judge whether the profile or the document is wrong. |
| block_tags | Array<[BlockTagRegistration](#blocktagregistration)>? | Fenced block info strings whose contents are parsed as structured data and attached to the containing document and heading. |
| finished_statuses | Array<[Status](vocabulary.md#status)>? | The statuses at which a document is considered finished, and therefore at which leftover template placeholders are a defect rather than work in progress. |
| placeholder_patterns | Array<String>? | Patterns whose presence in a finished document indicates unreplaced template scaffolding. |
| live_statuses | Array<[Status](vocabulary.md#status)>? | The statuses at which a document is in effect, and therefore at which a reference it makes to a retired document is worth reporting.<br><br>Profile-level rather than per-type because `status` is a property of the shared document envelope: every type this profile defines carries it, and none of them narrows it. A profile registering none of these three lifecycle keys registers no lifecycle, and the checks that read them do not run. |
| retired_statuses | Array<[Status](vocabulary.md#status)>? | The statuses at which a document is no longer in effect, so that a live document still pointing at it is reported.<br><br>Disjoint from `live_statuses` by construction; a status in neither is one the lifecycle says nothing about, which is the right answer for a draft. |
| successor_property | String? | The name of the frontmatter property in which a retired document names its replacement.<br><br>Registered rather than assumed, because the engine that reports a retired target must not contain the name of the property this framework happens to use for it. Its value is quoted back in the finding, so a reader is told what to cite instead of what they cited. |
| feature_tags | [FeatureTagRegistration](#featuretagregistration)? | The tag vocabulary this profile assigns meaning to.<br><br>Absent for a profile whose documents carry no tagged features, which is a corpus without them rather than a malformed registration; the checks that read it do not run. |
| conventions | [ConventionsConfig](config.md#conventionsconfig)? | The convention values this profile supplies when `specs.toml` omits them. |


## FeatureTagRegistration

The tags a profile assigns meaning to on a feature and its criteria.

Which tag identifies a feature, which identifies a criterion, which
prefix names the document a feature belongs to, and which qualifiers
carry meaning are facts about a document vocabulary rather than about
Gherkin. They are registered here so that the engine reads a tag as a
string and attaches meaning to none of it.

Every member is optional and a profile MAY register none of them. A
partial registration is honest rather than defective: a profile that
identifies features and not criteria has a corpus with no criteria in
it.

| Property | Type | Documentation |
|----------|------|---------------|
| feature_pattern | String? | The pattern the identifier of a feature matches, without the leading marker that makes it a tag.<br><br>Named capture groups in this pattern are what `shared_captures` selects from, so a profile that derives criterion identifiers from feature identifiers names the shared parts here. |
| criterion_pattern | String? | The pattern the identifier of an acceptance criterion matches, without the leading marker that makes it a tag. |
| owner_prefix | String? | The prefix carried by the tag in which a feature names the document that owns it, where the feature is not written inside that document.<br><br>A feature written as a standalone file has no frontmatter to inherit an owner from, so it names one in a tag; the prefix is what distinguishes that tag from every other tag on the feature. |
| shared_captures | Array<String>? | The named capture groups a criterion identifier shares with the identifier of the feature it belongs to.<br><br>This is how a criterion identifier's derivation from its feature's is stated without the engine knowing the internal structure of either pattern: it compiles both, captures both identifiers, and compares only the groups named here. Every name must be a capture group both patterns define. |
| qualifiers | Array<[QualifierTagRegistration](#qualifiertagregistration)>? | The tags that qualify a criterion, and what each of them means. |


## QualifierTagRegistration

One tag a profile assigns meaning to on a criterion, and the meaning it
assigns.

A tag absent from this list is carried on the record and reported by
nothing: an unregistered tag is a tag the vocabulary says nothing
about, which is the right answer for an annotation the author added for
their own reasons.

| Property | Type | Documentation |
|----------|------|---------------|
| tag | String | The tag, without the leading marker that makes it a tag. |
| meaning | [QualifierMeaning](#qualifiermeaning) | The property of a criterion this tag populates. |
| level | [VerificationLevel](feature.md#verificationlevel)? | The verification level this tag records, for a tag whose meaning is `verification-level` and for no other. |


## QualifierMeaning

What a registered qualifier tag says about the criterion carrying it.

The set is fixed by the criterion record rather than by any profile: a
qualifier can only mean one of the things a criterion has somewhere to
put. A profile chooses the spelling of each tag and which of these
meanings it carries, and nothing else.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Negation | "negation" | The criterion states a prohibition. |
| ManualVerification | "manual-verification" | The criterion can only be verified by hand. |
| VerificationLevel | "verification-level" | The criterion records the level it is intended to be verified at, which the tag's own `level` names. |


## DocumentTypeRegistration

Everything the engine needs to check one document type.

| Property | Type | Documentation |
|----------|------|---------------|
| schema | String | The JSON Schema this type's frontmatter is validated against, expressed as a JSON Pointer into the bundled schema or as a path to a standalone one. |
| slug_pattern | String | The pattern a document of this type must match in its `slug`.<br><br>A mismatch is reported with both the actual slug and this pattern quoted, so the finding is self-contained and the fixing agent needs no further lookup. |
| numbered | Bool | Whether documents of this type carry a sequence number the CLI can derive the next value of.<br><br>Numbering is global per type: the next number is one greater than the highest in use anywhere in the corpus, regardless of directory. |
| sections | Array<[Section](#section)>? | The heading hierarchy a document of this type must contain. |
| section_sets | Map<String, Array<String>>? | Named sets of headings a span query can select, so a reviewing agent can ask for the normative sections of a document and read only those. |
| reference_properties | Array<String>? | Frontmatter properties of this type that hold document references.<br><br>A value in one of these properties that matches no corpus member's slug is reported the same way an unresolved `[[slug]]` in the body is. |
| typed_reference_properties | Map<String, String>? | Reference-bearing properties whose target must declare a particular document `type`, keyed by property name with the required type as the value.<br><br>A schema can require that `frd` holds exactly one slug; only the index knows what that slug resolves to, so the type half of the constraint is registered here. Every key must also appear in `reference_properties` — a typed property that is not reference-bearing would be checked against a target nothing resolved. |
| pairings | Array<[PropertyPairing](#propertypairing)>? | Relationship properties that must be declared from both ends. |


## Section

A required heading, and the headings required beneath it.

Required sections are a hierarchy rather than a flat list because the
defects worth catching are about nesting: a missing `## Consequences`
and a `### Consequences` misfiled under `## Decision` are different
problems, and a flat list cannot tell them apart.

| Property | Type | Documentation |
|----------|------|---------------|
| title | String | The heading text, matched exactly. |
| required | Bool | Whether a document of this type must contain the heading. |
| repeated | Bool | Whether the heading occurs once per instance of something in the document rather than exactly once — one subsection per alternative considered, for example.<br><br>Without this distinction a section tree can only describe documents of fixed shape, and not all of the templates this framework ships are fixed. |
| children | Array<[Section](#section)>? | Headings required one level beneath this one. |


## PropertyPairing

One half of a relationship that must be declared from both ends.

A document naming a counterpart in one property where that counterpart
declares nothing in return is a half-built edge, and it is reported
against both documents rather than only the one that happens to be
checked first.

Reserved for a relationship a reader of one document must see without
consulting the index. Because the index derives the reverse of every
declared edge, a pairing used to enforce a mirror the index already
computes can only ever fail on redundant hand-work — so the
spec-driven-dev profile registers none, and an adopting profile that
registers one is expected to argue the case. See the Document
Relationship Standard.

| Property | Type | Documentation |
|----------|------|---------------|
| property | String | The property declared on this document. |
| counterpart | String | The property the referenced document must declare in return. |


## BlockTagRegistration

A fenced block whose contents are a YAML record: structured data about
an element declared inside a document.

A registered block's content must parse as YAML and must be a mapping.
Anything else — invalid YAML, a sequence, a bare scalar — is reported
at the position the block opens. YAML is the only format: a JSON object
is already valid YAML, and one format keeps every record readable by the
same parser.

| Property | Type | Documentation |
|----------|------|---------------|
| tag | String | The fenced block's info string — `component` for a component record block. |
| schema | String? | The JSON Schema the record is validated against, expressed as a JSON Pointer into the bundled schema or as a path to a standalone one. Optional: without one, a block is checked only for being a YAML mapping. |



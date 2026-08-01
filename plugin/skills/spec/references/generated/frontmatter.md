# Frontmatter 

## Table of Contents

- [ElementRecord](#elementrecord)
- [DocumentFrontmatter](#documentfrontmatter)
- [Adr](#adr)
- [Frd](#frd)
- [Contract](#contract)
- [Standards](#standards)
- [SystemBlueprint](#systemblueprint)
- [ContainerBlueprint](#containerblueprint)
- [ModuleBlueprint](#moduleblueprint)
- [ComponentBlueprint](#componentblueprint)
- [FeatureBlueprint](#featureblueprint)
- [Rfc](#rfc)
- [Guide](#guide)

## ElementRecord

The three fields that identify any element the framework names.

This is the smallest unit of description in the vocabulary, and it is
deliberately shared by two artifacts that look unrelated. A child element
a parent document declares inline — a component of a container, a module
of a system — carries exactly these fields in a fenced record block. A
child that has been promoted to its own document carries them as the
opening of its frontmatter, because `DocumentFrontmatter` inherits them.

One schema therefore validates a child in either state, which is what
makes promotion a lift-and-link operation rather than a re-authoring
one: the record block moves verbatim into the new document's
frontmatter, and every parent roster replaces the block with a link.

| Property | Type | Documentation |
|----------|------|---------------|
| title | String | Human-readable title. Free text.<br><br>On a promoted document this is also the document's H1, and the two are kept in sync by hand. |
| description | String | A one- to two-sentence summary of what this element is and what it is responsible for.<br><br>On C4 blueprints this is the brief description rendered on diagrams, so it is written to be read at a glance rather than as an introduction. |
| slug | [Slug](vocabulary.md#slug) | The identifier every other document uses to reference this element.<br><br>Unique across the corpus, compared without regard to letter case, and conventionally prefixed by the document type — `adr-013`, `container-spec-cli`, `module-document-core`. |


## DocumentFrontmatter

The metadata envelope every specification document carries.

Membership in the corpus is determined by this envelope rather than by
location: a markdown file whose frontmatter parses and declares a known
`type` is a specification, wherever it sits in the tree. Each document
type extends this struct with its own relationship and classification
properties, so the envelope is validated identically everywhere and only
the extension differs.

**Inherits from:** [ElementRecord](#elementrecord)

| Property | Type | Documentation |
|----------|------|---------------|
| type | [DocumentType](vocabulary.md#documenttype) | The document's classification, which selects the schema, slug pattern and required sections that govern it. |
| status | [Status](vocabulary.md#status) | The document's lifecycle state. |
| updated_at | DateTime | RFC 3339 timestamp of the last substantive modification.<br><br>Validated against the text as written in the file rather than against a parsed value, so a bare `2026-07-29` is reported as a defect instead of being silently widened to midnight UTC by the YAML parser. |
| created_at | DateTime? | RFC 3339 timestamp of the document's creation. Optional. |
| tags | Array<String>? | Free-form tags for filtering and search. |
| superseded_by | [Slug](vocabulary.md#slug)? | The document that replaces this one.<br><br>Required in practice on any document at `status: retired` — a retired document with no resolvable successor is a dead end for anyone who follows a reference to it. |
| related | Array<[Slug](vocabulary.md#slug)>? | Peer documents relevant to this one without a stronger relationship: the ADRs behind it, the standards it observes.<br><br>Every relationship property in this vocabulary names documents by slug, and every named slug must resolve to a corpus member. |


## Adr

An architecture decision record.

Carries no relationship properties of its own beyond the envelope's
`related`: an ADR's connections to other decisions are argued in its
prose, and the decisions it supersedes are named through
`superseded_by`.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

*No properties defined.*


## Frd

A feature requirements document: the user need, its scope boundaries and
the acceptance criteria that decide whether it has been met.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| sources | Array<[Slug](vocabulary.md#slug)>? | The documents this FRD is derived from — a brief, a research document, or an FRD it narrows. |


## Contract

An interface specification governing an exchange across a boundary of
independent change.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| provider | [Slug](vocabulary.md#slug) | The container or system that provides this interface.<br><br>Exactly one, because a contract with two providers is two contracts. |
| visibility | [Visibility](vocabulary.md#visibility) | Whether this interface is published beyond the repository that declares it, and therefore whether its versioning policy binds adopters outside this team. |
| consumers | Array<[Slug](vocabulary.md#slug)>? | Known consumers, by slug. Derived knowledge rather than a restriction — an unlisted consumer is not forbidden, it is merely unknown to the provider. |
| specs | Array<String>? | Paths or URLs of the machine-readable artifacts that make this contract executable: JSON Schemas, OpenAPI documents, protobuf definitions. |
| version | String | The contract's semantic version.<br><br>Required rather than optional: a contract without a version cannot state a compatibility promise, which is the only reason to write one down separately from the code that implements it. |


## Standards

Cross-cutting standards, policies, guidelines or constraints that apply
across documents or components rather than to one of them.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| version | String | The document's semantic version, quoted so YAML reads it as a string.<br><br>Unquoted `0.1` is a floating-point number to a YAML parser and fails validation here — which is the intended behaviour, because `0.10` and `0.1` are the same float and different versions. |


## SystemBlueprint

A C4 system blueprint: the containers a system bundles, the public
contracts it provides, and the systems it depends on.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| containers | Array<[Slug](vocabulary.md#slug)>? | Container blueprint slugs composed into this system. |
| contracts | Array<[Slug](vocabulary.md#slug)>? | Public contract slugs this system provides. |
| depends_on | Array<[Slug](vocabulary.md#slug)>? | Systems or contracts this system consumes. |


## ContainerBlueprint

A C4 container blueprint: one independently deployable or runnable
artifact, and the modules and components composed into it.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| system | [Slug](vocabulary.md#slug)? | The system this container belongs to.<br><br>Omitted when the container is standalone — that is, when no system blueprint documents it. Absence records a documented fact rather than an oversight. |
| components | Array<[Slug](vocabulary.md#slug)>? | Component blueprint slugs composed into this container. |
| modules | Array<[Slug](vocabulary.md#slug)>? | Module blueprint slugs composed into this container. |
| provides | Array<[Slug](vocabulary.md#slug)>? | Contract slugs this container provides. |
| consumes | Array<[Slug](vocabulary.md#slug)>? | Contract slugs this container consumes. |


## ModuleBlueprint

A C4 module blueprint: a coherent unit of code with an interface of its
own, composed into a container.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| kind | [BlueprintKind](vocabulary.md#blueprintkind) | Whether this blueprint documents one concrete module or a pattern several modules follow. |
| visibility | [Visibility](vocabulary.md#visibility) | Whether the module is published to a registry, and therefore whether its package API is bound by a compatibility policy. |
| provides | Array<[Slug](vocabulary.md#slug)>? | Contract slugs this module provides, where its interface warrants a standalone contract document. |
| components | Array<[Slug](vocabulary.md#slug)>? | Component or module blueprint slugs this module composes, where those children have been promoted to documents of their own.<br><br>A child is in exactly one state: promoted and linked here, or un-promoted and declared in a record block in the body. Never both. |
| source_paths | Array<String>? | Repository-relative source roots, so a reader can find the code. |
| test_paths | Array<String>? | Repository-relative test roots. |
| version | String? | The module's semantic version.<br><br>Required in practice when `visibility` is `public`, since a published package API owes its consumers a compatibility promise. |


## ComponentBlueprint

A C4 component blueprint: a single component promoted out of its
parent's roster into a document of its own, because it warrants more
than a record block.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| kind | [BlueprintKind](vocabulary.md#blueprintkind) | Whether this blueprint documents one concrete component or a pattern several components follow. |
| source_paths | Array<String>? | Repository-relative source roots. |
| test_paths | Array<String>? | Repository-relative test roots. |


## FeatureBlueprint

A transient change document describing how the system changes to deliver
one feature.

Unlike the blueprints it modifies, a feature blueprint describes a
change rather than a state, so it reaches `in-progress` and then
`retired` rather than `active`. Its `updates` and `creates` properties
are what let tooling derive when the change is complete: the feature is
not done until every document it names has been updated or activated.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| delivery | [Delivery](vocabulary.md#delivery) | The delivery unit grooming plans against. When `staged`, each stage boundary is enumerated in the blueprint's Sequencing Constraints. |
| sources | Array<[Slug](vocabulary.md#slug)> | The FRD this blueprint realizes, one to one.<br><br>A feature blueprint answers how the system changes; the FRD it names answers what the change must achieve. |
| updates | Array<[Slug](vocabulary.md#slug)>? | Slugs of existing documents this feature will change — blueprints at any C4 level, and contracts. |
| creates | Array<[Slug](vocabulary.md#slug)>? | Slugs of new documents authored alongside this blueprint, such as draft contracts for interfaces the feature introduces. |


## Rfc

An improvement proposal raised against a skill by a run that exercised
it.

The provenance properties below are the load-bearing part. A skill
behaves differently under different models, efforts and harnesses, so a
defect real under one may not reproduce under another, and triage cannot
scope a fix without knowing which run saw it. They are recorded
faithfully by the retro that raises the RFC and judged by the triage
that receives it — the two roles are deliberately separate.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| origin | String | Where the run happened: `self` when this repository critiqued its own skill, otherwise the name of the project the run took place in. |
| skill | String | The skill this RFC critiques. |
| skill_version | String | The version recorded in the critiqued skill's own frontmatter at the time of the run, so triage can tell whether a defect has already been fixed since. |
| model | String | The exact model identifier the run used — not a family name. A finding that reproduces under one model and not another is scoped by this property alone. |
| effort | String | The reasoning effort the run used.<br><br>Left unconstrained because the effort vocabulary belongs to the harness rather than to this framework, and `unknown` is a legitimate value when a transcript does not record it. |
| runs_observed | UInt ≥ 1 | How many runs stand behind the evidence.<br><br>This is the denominator for any claim about recurrence: an RFC may not assert more recurrence than it observed.<br><br>__Constraint:__ Values must be ≥ 1. |
| severity | [RfcSeverity](vocabulary.md#rfcseverity) | The highest severity among this RFC's proposals. |
| plugin_version | String | The plugin version the critiqued skill shipped in. |


## Guide

A how-to document explaining a technique that other document types
assume knowledge of — EARS syntax, C4 mapping, cross-document reference
conventions.

A guide carries no relationships of its own beyond the envelope's
`related`: it explains a technique rather than participating in the
architecture graph.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| version | String? | The guide's semantic version, quoted so YAML reads it as a string. |



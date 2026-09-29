# Frontmatter 

## Table of Contents

- [ElementRecord](#elementrecord)
- [DocumentFrontmatter](#documentfrontmatter)
- [ContractEdges](#contractedges)
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

The fields that declare an architectural element: what it is called, what
it is, how it is referenced, and what it is built with.

Deliberately shared by two artifacts that look unrelated. A child element
a parent document declares inline — a component of a container — carries
exactly these fields in a fenced record block. A child promoted to its
own blueprint carries them in its frontmatter, because every architecture
blueprint type inherits this struct.

One schema therefore validates a child in either state, which is what
makes promotion a lift-and-link operation rather than a re-authoring one:
the record block moves verbatim into the new blueprint's frontmatter, the
blueprint adds its document `title` and envelope, and every parent roster
replaces the block with a link. `description` and `slug` are declared on
both this struct and `DocumentFrontmatter`, and merge into one property
on a blueprint that inherits both.

| Property | Type | Documentation |
|----------|------|---------------|
| entity_name | String | The element's name: the canonical display name drawn on every diagram and matched against diagram labels. Free text.<br><br>Distinct from a blueprint's `title`, which names the document — "Spec CLI" is the entity name of the container whose blueprint is titled "Container Blueprint: Spec CLI". |
| description | String | A one- to two-sentence summary of what this is and what it is responsible for.<br><br>On C4 elements this is the brief description rendered on diagrams, so it is written to be read at a glance rather than as an introduction. |
| slug | [Slug](vocabulary.md#slug) | The identifier every other document uses to reference this.<br><br>Unique across the corpus, compared without regard to letter case, and conventionally prefixed by the document type — `adr-013`, `container-spec-cli`, `module-document-core`. |
| technology | String? | The language, framework, store or package type the element is built with — `Go, cobra`, `PostgreSQL 16`, `npm package`. Optional.<br><br>The technology line drawn on diagrams. |


## DocumentFrontmatter

The metadata envelope every specification document carries.

Membership in the corpus is determined by this envelope rather than by
location: a markdown file whose frontmatter parses and declares a known
`type` is a specification, wherever it sits in the tree. Each document
type extends this struct with its own relationship and classification
properties, so the envelope is validated identically everywhere and only
the extension differs.

A relationship property here or on any extension is declared at one end
only; the index derives the reverse from every other document's inbound
edges. Before adding one, check it against the Document Relationship
Standard — a property that merely mirrors the document's own prose
citations does not earn its maintenance cost.

| Property | Type | Documentation |
|----------|------|---------------|
| title | String | Human-readable document title. Free text.<br><br>Also the document's H1, and the two are kept in sync by hand. |
| description | String | A one- to two-sentence summary of what this is and what it is responsible for.<br><br>On C4 elements this is the brief description rendered on diagrams, so it is written to be read at a glance rather than as an introduction. |
| slug | [Slug](vocabulary.md#slug) | The identifier every other document uses to reference this.<br><br>Unique across the corpus, compared without regard to letter case, and conventionally prefixed by the document type — `adr-013`, `container-spec-cli`, `module-document-core`. |
| type | [DocumentType](vocabulary.md#documenttype) | The document's classification, which selects the schema, slug pattern and required sections that govern it. |
| status | [Status](vocabulary.md#status) | The document's lifecycle state. |
| updated_at | DateTime | RFC 3339 timestamp of the last substantive modification.<br><br>Validated against the text as written in the file rather than against a parsed value, so a bare `2026-07-29` is reported as a defect instead of being silently widened to midnight UTC by the YAML parser. |
| created_at | DateTime? | RFC 3339 timestamp of the document's creation. Optional. |
| tags | Array<String>? | Free-form tags for filtering and search. |
| superseded_by | [Citation](vocabulary.md#citation)? | The document that replaces this one.<br><br>Required in practice on any document at `status: retired` — a retired document with no resolvable successor is a dead end for anyone who follows a reference to it. |


## ContractEdges

The contract edges a C4 element owns: the interfaces it publishes, and
the interfaces it depends on.

Declared once and inherited by all four blueprint types rather than
repeated on each, because the same fact recorded at four altitudes
under three different spellings is how a vocabulary drifts. An element
provides a contract at whichever altitude it happens to sit; the
property is named for the concept, not for the level.

These are not the neighbour edges ADR-009 reserves to a parent. A
neighbour relationship varies by host — a shared module has different
collaborators in every container that composes it, which is why no
single outward list is correct and the parent owns it. A contract is a
named document with one identity, and the fact that an element
publishes or depends on it holds wherever that element is composed. So
this edge belongs to the element, and ADR-009's rule is untouched.

| Property | Type | Documentation |
|----------|------|---------------|
| contracts_provided | Array<[Citation](vocabulary.md#citation)>? | Contract slugs this element provides.<br><br>The interface is this element's own, so it is recorded here whatever the element's altitude — a system's public API, a container's service surface, a module's package API, a component's interface where it warrants a contract document of its own. |
| contracts_consumed | Array<[Citation](vocabulary.md#citation)>? | Contract slugs this element depends on.<br><br>Named contracts only. An element's untyped collaborators are not listed here — those are neighbour edges, and per ADR-009 they belong to the parent that composes the elements, or to the code. |


## Adr

An architecture decision record.

Carries no relationship properties of its own: an ADR's connections to
other decisions are argued in its prose, where the index reads them, and
the decision it supersedes is named through the envelope's
`superseded_by`.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

*No properties defined.*


## Frd

A feature requirements document: the user need, its scope boundaries and
the acceptance criteria that decide whether it has been met.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| sources | Array<[Citation](vocabulary.md#citation)>? | The documents this FRD is derived from — a brief, a research document, or an FRD it narrows. |


## Contract

An interface specification governing an exchange across a boundary of
independent change.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| provider | [Citation](vocabulary.md#citation) | The container or system that provides this interface.<br><br>Exactly one, because a contract with two providers is two contracts. |
| visibility | [Visibility](vocabulary.md#visibility) | Whether this interface is published beyond the repository that declares it, and therefore whether its versioning policy binds adopters outside this team. |
| consumers | Array<[Citation](vocabulary.md#citation)>? | Known consumers, by slug. Derived knowledge rather than a restriction — an unlisted consumer is not forbidden, it is merely unknown to the provider. |
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

A C4 system blueprint: the public contracts a system provides.

The containers a system bundles are not listed here. Each container
declares the system that composes it in its own `system` property, and
the roster is derived from those inbound edges — one authoritative end,
per the Document Relationship Standard.

A system's dependency on another system is not recorded in frontmatter
at all. Where the dependency is a named interface it is a contract, and
`contracts_consumed` holds it. Where it is looser, the blueprint's own
"Consumed Systems and Contracts" section says what is depended on and
what for — which a slug list cannot, and which is what a reader actually
needs. A repository spanning more than one system is rare enough that a
frontmatter property for the untyped case would be maintained by nobody.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter), [ElementRecord](#elementrecord), [ContractEdges](#contractedges)

*No properties defined.*


## ContainerBlueprint

A C4 container blueprint: one independently deployable or runnable
artifact.

The modules and components composed into it are not listed here, and
the two arrive by different routes. A module names its hosts itself, in
`contained_by`, because a module may be composed into several containers
and that set is a fact about the module. A component has exactly one
parent, as C4 requires, and that edge is recorded where a reader meets
it: the Components section of this document's body, where each child is
a fenced record block XOR a `[[slug]]` blueprint link under its own
heading, per ADR-008 and ADR-012. Either way the index derives this
container's roster rather than reading it from frontmatter.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter), [ElementRecord](#elementrecord), [ContractEdges](#contractedges)

| Property | Type | Documentation |
|----------|------|---------------|
| system | [Citation](vocabulary.md#citation)? | The system this container belongs to.<br><br>Omitted when the container is standalone — that is, when no system blueprint documents it. Absence records a documented fact rather than an oversight. |


## ModuleBlueprint

A C4 module blueprint: a coherent unit of code with an interface of its
own, composed into a container.

The components it composes are not listed in frontmatter. Each child is
declared in this document's Components section — a fenced record block
when un-promoted, a `[[slug]]` blueprint link once promoted, exactly one
of the two per ADR-008 — and the index derives the roster from there.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter), [ElementRecord](#elementrecord), [ContractEdges](#contractedges)

| Property | Type | Documentation |
|----------|------|---------------|
| kind | [BlueprintKind](vocabulary.md#blueprintkind) | Whether this blueprint documents one concrete module or a pattern several modules follow. |
| visibility | [Visibility](vocabulary.md#visibility) | Whether the module is published to a registry, and therefore whether its package API is bound by a compatibility policy. |
| contained_by | Array<[Citation](vocabulary.md#citation)>? | The containers, or the modules, that compose this module.<br><br>**The only multi-parent relationship in the vocabulary, and the reason the module level exists.** C4 gives every element exactly one parent — a container belongs to one system, a component to one container — and that edge is recorded by the parent, in the roster section of its body. Module is this framework's addition to C4, and being composed into more than one container is its whole purpose: a unit shared by two containers is a module precisely because it is shared.<br><br>So this is a list, and it is declared here rather than derived from parent rosters, because a shared module's hosts are a fact about the module. It is the direction ADR-009 tells modules to point when it denies them a context view — no single outward-edge list is correct for a module, but the set of hosts is, and this is where it goes.<br><br>Omitted when no host has a blueprint of its own — absence records a documented fact rather than an oversight. |
| source_paths | Array<String>? | Repository-relative source roots, so a reader can find the code. |
| test_paths | Array<String>? | Repository-relative test roots. |
| version | String? | The module's semantic version.<br><br>Required in practice when `visibility` is `public`, since a published package API owes its consumers a compatibility promise. |


## ComponentBlueprint

A C4 component blueprint: a single component promoted out of its
parent's roster into a document of its own, because it warrants more
than a record block.

Carries no parent property. A component has exactly one parent, as C4
requires, and its parent's roster section already links it by
`[[slug]]` under its own heading — so the edge is declared once, in the
one place a reader meets the child, and the index derives the component's
parent from that inbound link. `contained_by` is the module level's
alone, because being composed into more than one container is what
distinguishes a module from a component.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter), [ElementRecord](#elementrecord), [ContractEdges](#contractedges)

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
`retired` rather than `active`.

The documents the change creates and updates are described in the
blueprint's own "Documentation Changes" prose and tracked to completion
in its `plan`, which records done-ness natively. They are not mirrored
into frontmatter: a change set recorded in three places is three places
that must agree.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| delivery | [Delivery](vocabulary.md#delivery) | The delivery unit grooming plans against. When `staged`, each stage boundary is enumerated in the blueprint's Sequencing Constraints. |
| plan | String? | Where this feature's implementation plan lives, stamped in by the groom-feature-blueprint skill once grooming serializes the plan: an epic id (`agent-plugins-bge`) when the plan is beads, or a repository-relative directory (`plans/<feature-slug>/`) when it is task files. Absent before grooming, since an ungroomed blueprint has no plan yet; present through `retired`, since the pointer retires with the document rather than being cleared. |
| frd | [Citation](vocabulary.md#citation) | The FRD this blueprint realizes, one to one.<br><br>A feature blueprint answers how the system changes; the FRD it names answers what the change must achieve. Singular and required, so the one-to-one relationship is expressible in the schema rather than only in this comment, and named `frd` rather than `sources` so that one property name does not span two relationships — `Frd.sources` is optional, many-valued provenance, which is a different edge. |


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

A guide carries no relationships of its own: it explains a technique
rather than participating in the architecture graph.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

| Property | Type | Documentation |
|----------|------|---------------|
| version | String? | The guide's semantic version, quoted so YAML reads it as a string. |



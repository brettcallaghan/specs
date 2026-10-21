# Frontmatter 

## Table of Contents

- [Identified](#identified)
- [C4ElementRecord](#c4elementrecord)
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
- [Guide](#guide)
- [TypedFrontmatter](#typedfrontmatter)

## Identified

The two fields every described thing carries, document and element alike:
its identifier and its one-line summary.

| Property | Type | Documentation |
|----------|------|---------------|
| description | String | A one- to two-sentence summary of what this is and what it is responsible for.<br><br>On C4 elements this is the brief description rendered on diagrams, so it is written to be read at a glance rather than as an introduction. |
| slug | [Slug](vocabulary.md#slug) | The identifier every other document uses to reference this.<br><br>Unique across the corpus, compared without regard to letter case, and conventionally prefixed by the document type — `adr-013`, `container-spec-cli`, `module-document-core`. |


## C4ElementRecord

The fields that declare an architectural element: what it is called, what
it is, how it is referenced, and what it is built with.

Deliberately shared by two artifacts that look unrelated. An element
declared inline — a component in its container's roster — carries
exactly these fields in a fenced record block. An element with its own
blueprint carries them in its frontmatter, because every architecture
blueprint type inherits this struct.

One schema therefore validates every declaration of an element, and an
element may have several: a blueprint and the record block in its
parent's roster, or blocks in more than one document. They are one
element because they share a `slug`. Promotion copies a block's fields
into a new blueprint's frontmatter, which adds the document's `title`
and envelope; the block may stay where it is.

**Inherits from:** [Identified](#identified)

| Property | Type | Documentation |
|----------|------|---------------|
| entity_name | String | The element's name: the canonical display name drawn on every diagram and matched against diagram labels. Free text.<br><br>Distinct from a blueprint's `title`, which names the document. The two often read the same, but only this one is matched against diagram labels, so a title may be reworded without touching a diagram. |
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

There is deliberately no creation or last-modified timestamp: git
records when a file was added and changed, an index can derive both from
it, and a hand-maintained copy goes stale.

**Inherits from:** [Identified](#identified)

| Property | Type | Documentation |
|----------|------|---------------|
| title | String | Human-readable document title. Free text.<br><br>Also the document's H1, and the two are kept in sync by hand. It carries no type label — `type` says what the document is — except that an ADR's title leads with its number, `ADR-011: …`, because the number is how an ADR is cited. |
| status | [Status](vocabulary.md#status) | The document's lifecycle state. |
| tags | Array<String>? | Free-form tags for filtering and search. |
| superseded_by | [Citation](vocabulary.md#citation)? | The document that replaces this one.<br><br>Required in practice on any document at `status: retired` — a retired document with no resolvable successor is a dead end for anyone who follows a reference to it. |


## ContractEdges

The contract edge a C4 element declares: the interfaces it depends on.

Inherited by all four blueprint types, because an element consumes a
contract at whichever altitude it sits. The other end of a contract
edge — which element provides the contract — is declared once, on the
contract, as `Contract.provider`; an element's provided contracts are
derived from it. Each edge is declared on the document whose change
creates it: an element that starts depending on a contract is the
element edited, and a new contract names its provider when it is
written.

These are not the neighbor edges ADR-009 reserves to a parent. A
neighbor relationship varies by host; a contract is a named document
with one identity, and the fact that an element depends on it holds
wherever that element is composed.

| Property | Type | Documentation |
|----------|------|---------------|
| contracts_consumed | Array<[Citation](vocabulary.md#citation)>? | Contract slugs this element depends on.<br><br>Named contracts only. An element's untyped collaborators are not listed here — those are neighbor edges, and per ADR-009 they belong to the parent that composes the elements, or to the code. The contract's consumers are derived from this property. |


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
| provider | [Citation](vocabulary.md#citation) | The element that provides this interface — a system, container, module or component.<br><br>Exactly one, because a contract with two providers is two contracts. Declared here and nowhere else: the provider's list of provided contracts is derived from this property, so writing a new contract never requires editing the element that provides it. |
| visibility | [Visibility](vocabulary.md#visibility) | Whether this interface is published beyond the repository that declares it, and therefore whether its versioning policy binds adopters outside this team. |
| specs | Array<String>? | Paths or URLs of the machine-readable artifacts that make this contract executable: JSON Schemas, OpenAPI documents, protobuf definitions. |
| version | String | The contract's semantic version.<br><br>Required rather than optional: a contract without a version cannot state a compatibility promise, which is the only reason to write one down separately from the code that implements it. |


## Standards

Cross-cutting standards, policies, guidelines or constraints that apply
across documents or components rather than to one of them.

Carries no version of its own. A standard that ships to other
repositories is vendored into a skill, and the release version of the
plugin that ships the skill identifies the text a reader followed.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

*No properties defined.*


## SystemBlueprint

A C4 system blueprint: the public contracts a system provides.

The containers a system bundles are not listed here. Each container
declares the system that composes it in its own `system` property, and
the roster is derived from those inbound edges — one authoritative end,
per the Document Relationship Standard.

A system's dependency on another system is not recorded in frontmatter
at all. Where the dependency is a named interface it is a contract, and
`contracts_consumed` holds it. Where it is looser, the blueprint's own
"Consumed systems and Contracts" section says what is depended on and
what for — which a slug list cannot, and which is what a reader actually
needs. A repository spanning more than one system is rare enough that a
frontmatter property for the untyped case would be maintained by nobody.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter), [C4ElementRecord](#c4elementrecord), [ContractEdges](#contractedges)

*No properties defined.*


## ContainerBlueprint

A C4 container blueprint: one independently deployable or runnable
artifact.

The modules composed into it are listed in `modules`. Its components are
declared where a reader meets them: the Components section of this
document's body, where each child sits under its own heading as a
fenced record block, a `[[slug]]` blueprint link, or both, per ADR-008
and ADR-012.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter), [C4ElementRecord](#c4elementrecord), [ContractEdges](#contractedges)

| Property | Type | Documentation |
|----------|------|---------------|
| system | [Citation](vocabulary.md#citation)? | The system this container belongs to.<br><br>Omitted when the container is standalone — that is, when no system blueprint documents it. Absence records a documented fact rather than an oversight. |
| modules | Array<[Citation](vocabulary.md#citation)>? | The modules composed into this container.<br><br>Declared on the host rather than on the module, because adopting a module is a change to the container: the container is the document being edited, and the module — which may already serve other hosts — is not touched. A module's hosts are derived from this property. |


## ModuleBlueprint

A C4 module blueprint: a coherent unit of code with an interface of its
own, composed into a container.

The components it composes are declared in this document's Components
section — a fenced record block, a `[[slug]]` blueprint link, or both,
per ADR-008. The modules it composes are listed in `modules`, as on a
container. Its own hosts are not declared here; they are derived from
the `modules` of every container or module that composes it.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter), [C4ElementRecord](#c4elementrecord), [ContractEdges](#contractedges)

| Property | Type | Documentation |
|----------|------|---------------|
| kind | [BlueprintKind](vocabulary.md#blueprintkind) | Whether this blueprint documents one concrete module or a pattern several modules follow. |
| visibility | [Visibility](vocabulary.md#visibility) | Whether the module is published to a registry, and therefore whether its package API is bound by a compatibility policy. |
| modules | Array<[Citation](vocabulary.md#citation)>? | The modules composed into this one, declared on this host for the same reason a container declares its modules. |
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
parent from that inbound link.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter), [C4ElementRecord](#c4elementrecord), [ContractEdges](#contractedges)

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
| delivery | [Delivery](vocabulary.md#delivery) | The delivery unit grooming plans against. When `staged`, each stage boundary is enumerated in the blueprint's Sequencing constraints. |
| plan | String? | Where this feature's implementation plan lives, stamped in by the groom-feature-blueprint skill once grooming serializes the plan: an epic id (`bd-a3f8`) when the plan is beads, or a repository-relative directory (`plans/<feature-slug>/`) when it is task files. Absent before grooming, since an ungroomed blueprint has no plan yet; present through `retired`, since the pointer retires with the document rather than being cleared. |
| frd | [Citation](vocabulary.md#citation) | The FRD this blueprint realizes, one to one.<br><br>A feature blueprint answers how the system changes; the FRD it names answers what the change must achieve. Singular and required, so the one-to-one relationship is expressible in the schema rather than only in this comment, and named `frd` rather than `sources` so that one property name does not span two relationships — `Frd.sources` is optional, many-valued provenance, which is a different edge. |


## Guide

A how-to document explaining a technique that other document types
assume knowledge of — EARS syntax, C4 mapping, cross-document reference
conventions.

A guide carries no relationships of its own: it explains a technique
rather than participating in the architecture graph. Like a standard,
it carries no version of its own: a guide that ships to other
repositories is vendored into a skill, and the release version of the
plugin that ships the skill identifies the text.

**Inherits from:** [DocumentFrontmatter](#documentfrontmatter)

*No properties defined.*


## TypedFrontmatter

A document's frontmatter, as its `type` defines it: the shared envelope
plus every property that type adds.

`type` is this union's discriminator, so it is declared here rather than
on `DocumentFrontmatter` — Strata does not let a variant declare its
union's discriminator. Every variant's JSON Schema still requires `type`
as a constant, and every encoded value still carries it, so documents
and the index keep `type` exactly where they always had it. A document
is validated against its variant rather than the bare struct, because
the bare struct is closed and does not define `type`.

The tag values are the `DocumentType` vocabulary, one variant per value.

**Discriminator:** type (String)

| Tag | Type | Documentation |
|-----|------|---------------|
| "adr" | [Adr](#adr) |  |
| "frd" | [Frd](#frd) |  |
| "contract" | [Contract](#contract) |  |
| "standards" | [Standards](#standards) |  |
| "system-blueprint" | [SystemBlueprint](#systemblueprint) |  |
| "container-blueprint" | [ContainerBlueprint](#containerblueprint) |  |
| "module-blueprint" | [ModuleBlueprint](#moduleblueprint) |  |
| "component-blueprint" | [ComponentBlueprint](#componentblueprint) |  |
| "feature-blueprint" | [FeatureBlueprint](#featureblueprint) |  |
| "guide" | [Guide](#guide) |  |



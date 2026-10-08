# Vocabulary 

## Table of Contents

- [Slug](#slug)
- [Citation](#citation)
- [DocumentType](#documenttype)
- [RfcSeverity](#rfcseverity)
- [Status](#status)
- [Visibility](#visibility)
- [BlueprintKind](#blueprintkind)
- [Delivery](#delivery)

## Type Alias: Slug

A cross-reference identifier for a specification document.

Every relationship between documents is expressed as a slug rather than a
path, so moving a file never breaks a reference. Slugs are kebab-case:
lowercase alphanumeric words joined by single hyphens.

**Definition:**

```strata
alias Slug: String ~= `/^[a-z0-9]+(?:-[a-z0-9]+)*$/`
```


## Type Alias: Citation

A reference to another specification document, optionally qualified by
an anchor into one of its sections.

Held apart from `Slug` because the two type different things: `Slug` is
the identity a document declares about itself, and admitting an anchor
there would let a document declare its identity as `adr-012#Decision`,
breaking uniqueness and slug resolution. A citation names someone
else's identity and may legitimately point at a place inside it — the
slug half matches the same pattern `Slug` does, and a trailing `#` with
nothing after it is rejected rather than admitted as an anchor-less
citation.

**Definition:**

```strata
alias Citation: String ~= `/^[a-z0-9]+(?:-[a-z0-9]+)*(?:#.+)?$/`
```


## DocumentType

The classification of a specification document.

A document's type selects the schema, slug pattern and required sections
that govern it. This is the controlled vocabulary the specs
profile registers; a repository whose documents declare a type outside
this set is recorded as excluded with reason `unknown-type` rather than
being checked against the wrong rules.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Adr | "adr" | An architecture decision record: one decision, its alternatives and its consequences. |
| Frd | "frd" | A feature requirements document: the user-facing need and its acceptance criteria. |
| Contract | "contract" | An interface specification shared across a boundary of independent change. |
| Standards | "standards" | Cross-cutting standards, policies, guidelines or constraints. |
| SystemBlueprint | "system-blueprint" | A C4 system blueprint: the containers a system bundles and the contracts it provides. |
| ContainerBlueprint | "container-blueprint" | A C4 container blueprint: one deployable artifact and the modules composed into it. |
| ModuleBlueprint | "module-blueprint" | A C4 module blueprint: a coherent unit of code with its own interface. |
| ComponentBlueprint | "component-blueprint" | A C4 component blueprint: a single component promoted to its own document. |
| FeatureBlueprint | "feature-blueprint" | A transient change document describing how the system changes to deliver one feature. |
| Rfc | "rfc" | An improvement proposal raised against a skill by a run that exercised it. |
| Guide | "guide" | A how-to document explaining a technique the other document types assume. |


## RfcSeverity

How badly a run was impeded by the skill an RFC critiques.

An RFC carries the highest severity among its proposals, so that triage
can rank a queue without opening every document.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Blocker | "blocker" | Forced an unscripted improvisation, or the skill as written produced a wrong result. |
| Snag | "snag" | The skill was silent on a decision that materially changed the outcome. |
| Polish | "polish" | It worked, but a change would make the next run clearer, faster, or safer. |


## Status

The lifecycle state of a document.

State documents — blueprints, standards and contracts, which describe
what IS — use `active` and never `in-progress`. Transient change
documents, which describe a change, use `in-progress` in its place: a
change specification is implemented, not put into effect.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Draft | "draft" | Under authorship; content may be incomplete. |
| ReadyForReview | "ready-for-review" | Complete and awaiting review. |
| Approved | "approved" | Reviewed and accepted; not yet in effect. |
| InProgress | "in-progress" | Approved, and the change it specifies is being implemented. |
| Active | "active" | In effect; the current source of truth. |
| Retired | "retired" | No longer in effect. A retired document names its replacement in `superseded_by`. |


## Visibility

Whether an interface or module is published beyond the repository that
declares it.

Visibility governs the compatibility obligation, not the access
modifier: a public surface owes adopters a versioning policy, while an
internal one may change in lockstep with its only consumers.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Public | "public" | Published for consumption outside this repository; bound by its versioning policy. |
| Internal | "internal" | Unpublished and in-repo; evolves in lockstep with its consumers. |


## BlueprintKind

Whether a blueprint documents one concrete element or a pattern that
several elements follow.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Standard | "standard" | One concrete element, documented in its own right. |
| Pattern | "pattern" | A repeated shape that several elements instantiate. |


## Delivery

The delivery unit a feature blueprint is groomed against.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| SinglePr | "single-pr" | The whole feature lands in one pull request. |
| Staged | "staged" | The feature lands across several stages, each boundary enumerated in Sequencing Constraints. |



---
title: "Specs Profile"
entity_name: Specs Profile
technology: Go package
description: This framework's document vocabulary, registered into Document
  Core as one profile — the union schema, each type's slug pattern and
  numbering, reference-bearing properties, lifecycle, Gherkin tag
  vocabulary and architecture vocabulary. Change it to add a document type,
  alter a slug pattern, or register a framework rule the generated schema
  can't state.
slug: module-spec-profile
type: module-blueprint
status: ready-for-review
kind: standard
visibility: internal
contracts_consumed:
  - contract-specs-configuration
source_paths:
  - "cli/profile"
test_paths:
  - "cli/profile"
---

# Specs Profile

## Responsibility

This module supplies this framework's document vocabulary, as data, through
`corpus.Profile`: the extension point of an engine that indexes and checks
any corpus of Markdown documents carrying frontmatter and citing each other.
The engine knows nothing of ADRs, slug prefixes or Feature Blueprints until
a profile registers them, so a different taxonomy reuses the whole engine
by registering a different profile.

### Owns

The registration of this framework's vocabulary through `corpus.Profile`:
the embedded union schema, each document type's slug pattern and
numbering, the reference-bearing properties and typed references, the
status lifecycle, the convention defaults, the Gherkin tag vocabulary, the
record-block tags, and the architecture vocabulary.

### Hides

Which document types this framework defines and how each is cited,
numbered, referenced and retired, and how those facts are read out of the
generated schema's JSON. A new type or a regenerated schema changes nothing
behind `corpus.Profile`.

### Does not own

Parsing, indexing, reference resolution, and every check that reads the
registration, including the lifecycle and record-block checks —
[[module-document-core]] owns them.

The command surface, configuration loading and profile selection —
[[container-spec-cli]] owns them.

The slug prefix registry and the line-width rule —
[[spec-authoring-standards]] owns them.

The status vocabulary and the element-record format —
[[frontmatter-standards]] owns them.

The Gherkin tag vocabulary — [[frd-gherkin-adoption]] owns it.

The architecture vocabulary — [[c4-standards]] owns it.

The frontmatter schema and the document-type vocabulary — `schemas/*.strata`
is their single source ([[adr-016]]), and `scripts/generate-schemas.sh`
generates the schema this module embeds.

### Composed into

[[container-spec-cli]]; see that blueprint for this module's relationships
to sibling components.

### Vended as

The Go package `cli/profile` in the CLI's Go module (`cli/go.mod`), compiled
into the `specs` binary. `cli/cmd/registry.go` registers it under
`profile.Name`, and `[profile].name` in `specs.toml` selects it.

## Tech stack

**Language:** Go, standard library only, composed into its host as a
package. [[adr-015]]

**Vocabulary:** Strata. The union schema and the generated document-type
constants come from `schemas/*.strata`, and this module registers them
rather than declaring types of its own. [[adr-016]]

## Interface

### Surface

`profile.New` in `cli/profile/profile.go`, which returns a
`corpus.Profile` (`cli/core/corpus/profile.go`).
[[contract-specs-configuration]] specifies the registration surface as its
internal half.

### Provides

The module SHALL register a slug pattern for every document type this
framework defines, so that a conforming document is never excluded as
`unknown-type`.

The module SHALL register as reference-bearing exactly the frontmatter
properties that the generated schema types as `Citation` or `[Citation]`,
other than `slug`.

WHEN a Feature Blueprint declares `frd`, the registration SHALL require the
value to resolve to a document of type `frd`.

The module SHALL NOT register a pairing.

**Rationale:** every relationship this framework declares has its reverse
derived from inbound edges, so [[document-relationship-standards]]'s
pairing check has nothing to pair.

The module SHALL register `retired` as the only retired status.

The module SHALL register `superseded_by` as the successor property.

The module SHALL register every status other than `retired` as live.

**Rationale:** the retired-target warning tells an author what to cite
instead, and the author who most needs it is editing a draft.

The module SHALL register the `line_width` default that
[[spec-authoring-standards]] fixes.

The module SHALL register the Gherkin tag vocabulary that
[[frd-gherkin-adoption]] defines (`featureTags` in `cli/profile/profile.go`).

The module SHALL register the architecture vocabulary that [[c4-standards]]
defines, for the architecture blueprint types only ([[adr-030]]).

The module SHALL register the `component` record-block tag against the
generated `C4ElementRecord` definition ([[adr-008]]).

WHERE a type's documents in this corpus don't follow the prefix registry in
[[spec-authoring-standards]], the module SHALL register the type's least
restrictive slug pattern.

The module SHALL record each such divergence at the table in
`cli/profile/profile.go`.

WHERE a rule is expressible as a JSON Schema or a pattern, the module SHALL
register it as data rather than implement it as code.

**Rationale:** a registered pattern can be reviewed against the template
it mirrors. The same rule inside a predicate can't.

The module SHALL NOT parse document text.

### Requires

WHEN the package initializes, the embedded union schema (`schema.SpecModel`)
SHALL declare a definition for every document type in the authored table.

WHEN the package initializes, the embedded union schema SHALL declare the
`Status` enum.

The generated schema and constants this module reads SHALL be the current
output of `scripts/generate-schemas.sh` for `schemas/*.strata`.

WHEN a repository selects this profile, the host SHALL read the vocabulary
only through `corpus.Profile`, as [[contract-specs-configuration]]
specifies.

## Design

### Patterns and principles

Plugin (Buschmann et al., *POSA Vol. 1*, 1996): `corpus.Profile` is the
named extension point, and `profile.New` returns the plugin. The need is
the extension point itself: [[contract-specs-configuration]] selects a
profile by name, and the engine carries no vocabulary without one. It
rules out a plugin that drives the core, stated under Invariants.

Single Source of Truth (Codd, 1970; Hunt and Thomas, 1999): the generated
schema is the one authority for which types, citation properties and
statuses exist. Derived Registration reads them from it, and tests hold the
authored half to it. The need is that `schemas/*.strata` already states the
vocabulary ([[adr-016]]), so a transcribed second list could only drift.
It rules out a list of types, properties or statuses the schema doesn't
hold, stated under Invariants as one drift guard per registered set.

Fail Fast (Shore, 2004): a schema the module can't read stops the binary at
package initialization. The need is that a registration which quietly
registered nothing would make every check pass over nothing. It rules
out starting with an empty or partial registration, stated under
Invariants.

### Invariants

Outside its tests, the module SHALL NOT call Document Core except to
construct the values `corpus.Profile` returns.
Enforced by: review only.

The authored per-type table SHALL hold one entry, with a compiling slug
pattern, for every value of the generated `DocumentType` vocabulary.
Enforced by: `TestDocumentTypesCoversEveryVocabularyValue`, and the
compiler, because the table is keyed by the generated constants.

The registered reference-bearing properties of each type SHALL equal the
properties the generated schema types as `Citation` or `[Citation]` on that
type, other than `slug`.
Enforced by: `TestEveryCitationTypedPropertyInTheVocabularyIsRegistered`.

The generated schema SHALL declare each registered architecture property on
every type that property applies to.
Enforced by: `TestEveryRegisteredArchitecturePropertyIsDeclaredWhereItApplies`,
and `corpus.ValidateRegistration` at activation.

The registered feature and criterion identifier patterns, with their
capture-group names removed, SHALL equal the generated `FeatureId` and
`CriterionId` patterns.
Enforced by: `TestFeatureTagPatternsMatchTheGeneratedIdentifiers`.

The registration SHALL hold no defect that `corpus.ValidateRegistration`
reports.
Enforced by: `TestTheRegistrationHasNoDefects`.

IF the embedded schema can't be decoded, or declares no branch for a type
in the authored table, THEN the module SHALL panic at package
initialization.
Enforced by: review only.

### Extension points

A new document type — declare it in `schemas/*.strata`, regenerate, and
add its row to `documentTypes` in `cli/profile/profile.go`. Not by keying a
row with a string literal, and not by adding a per-type schema.

A framework rule the generated schema can't state — a new member on
`corpus.Profile`, as a change to the host. Not by a predicate special-cased
for this profile, and not by parsing documents here.

A different document taxonomy — a separate profile that implements
`corpus.Profile`. Not by parameterizing this one.

## Components

```mermaid
C4Component
    title Components — Specs Profile
    Container_Boundary(profile, "Specs Profile") {
        Component(reg, "Profile Registration", "Go", "The plugin entry point and the authored vocabulary")
        Component(derived, "Derived Registration", "Go", "The reference and lifecycle registration, derived from the generated schema")
    }
    Rel(reg, derived, "Gets the per-type registration and the lifecycle from")
```

Profile Registration gets the per-type registration and the lifecycle from
Derived Registration, so the host sees one registration with no seam between
what is authored and what is derived. Derived Registration widens Profile
Registration's authored per-type table rather than keeping a list of types
of its own, so a type is added in one place.

### Profile Registration

```component
entity_name: Profile Registration
slug: component-profile-registration
description: The plugin entry point and the authored vocabulary
technology: Go
```

(Go) — The plugin entry point and the authored vocabulary
(`cli/profile/profile.go`). It returns the `corpus.Profile` value and holds
the facts the generated schema can't state: each type's slug pattern and
numbering, the convention default, the Gherkin tag vocabulary, the
record-block tags, and the architecture vocabulary. It hides which of the
framework's rules are authored, so a reviewer checks them in one file.

### Derived Registration

```component
entity_name: Derived Registration
slug: component-derived-registration
description: The reference and lifecycle registration, derived from the generated schema
technology: Go
```

(Go) — The reference and lifecycle registration, derived from the generated
schema (`cli/profile/registration.go`). It reads each type's
reference-bearing properties and the status vocabulary from the embedded
schema, so a property or status added to `schemas/*.strata` is registered
the moment it's declared. It authors only what the schema can't state: the
typed reference, the empty pairing set, the retired status and the successor
property. It hides the generated schema's JSON shape, such as the `$ref`,
`oneOf` and `items` forms a citation takes.

## Cross-cutting Standards

This module SHALL comply with [[testing-standards]]: each component above
carries unit tests covering its exported surface, beside its subject.

## Change guidance

WHEN this framework's document vocabulary changes, the registration SHALL
be updated in the same change as the template or Standards document that
defines the change.

WHEN a rule appears to need parsing, indexing or reference resolution, the
change SHALL add that capability to [[module-document-core]] and register
against it here.

IF a rule can't be expressed through `corpus.Profile`, THEN the change SHALL
widen that interface as a change to the host, reviewed on its own merits.

The module SHALL NOT require a change to `corpus.Profile` for a reason
specific to this framework.

WHILE visibility is internal, the module and its in-repo consumers evolve
together: a breaking change MAY be made in the same change as the consumer
updates, with no compatibility window.

### Verification gates

#### Verify always

`go test ./profile/...`, run from `cli/`. Beside the isolation tests, the
package holds gates that build this repository's own corpus with the
profile: `cli/profile/self_application_test.go` and
`cli/profile/architecture_test.go`. They fail when a registered rule
drifts from the documents it governs, so a change to the corpus can fail
them too.

#### Verify conditionally

WHEN `schemas/*.strata` changes, run `scripts/generate-schemas.sh --check`.

WHEN the registration changes what Document Core checks, run
`go test ./...` from `cli/` and a whole-corpus `specs check`.

## Key decisions

[[adr-008]] — un-promoted elements are declared in record blocks validated
against a shared element-record schema.

[[adr-012]] — the reference conventions the registered types rely on.

[[adr-015]] — Go, and the profile-as-package composition model.

[[adr-016]] — Strata is the single source for the frontmatter schema and the
document-type vocabulary this module registers.

[[adr-021]] — Gherkin is where features and acceptance criteria are
authored, and this module registers their tag vocabulary.

[[adr-030]] — the C4 architecture vocabulary reaches Document Core through
this profile.

## Open questions

- [ ] Where record-block validation and the declaration-agreement rule from
      [[adr-008]] live: a component of this module, or a Document Core check
      that reads registered data. This module registers the `component`
      tag and its schema; neither the validation nor the rule is built.
- [ ] `kind` is `standard` because exactly one profile exists. WHEN a
      second profile is written, such as an enterprise taxonomy or a
      stripped Obsidian-vault profile, this blueprint SHOULD be
      reconsidered as `kind: pattern`, governing how any profile is built.
      Deciding before then would invent a pattern from one example.
- [ ] How much of the vocabulary ships as data files rather than
      compiled-in defaults. This forces the answer to the promotion
      question in [[contract-specs-configuration]], because a data file is
      most of the way to an override mechanism. Everything is compiled in
      today to keep the binary self-contained, which doesn't answer the
      question.
- [ ] Whether the deferred domain rules (AC traceability, C4 diagram
      conventions, blueprint AC-coverage keys, and Work Order readiness)
      become components here or a second profile layered on this one.

## References

[[container-spec-cli]] — the container this module is composed into.

[[contract-specs-configuration]] — the registration surface this module
populates.

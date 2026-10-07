---
title: "Module Blueprint: Spec-Driven Development Profile"
entity_name: Spec-Driven Development Profile
description: The plugin that supplies this framework's document
  vocabulary — one union schema over every document type, plus each
  type's slug pattern and numbered-ness — to Document Core
slug: module-spec-profile
type: module-blueprint
status: approved
updated_at: 2026-10-07T00:00:00Z
kind: standard
visibility: internal
source_paths:
  - "cli/profile"
test_paths:
  - "cli/profile"
version: "0.3"
---

# Spec-Driven Development Profile

## Overview

This module is the domain half of the CLI. It knows that an ADR is a
document type, that its slug is `adr-<nnn>`, and that ADRs are numbered
while FRDs are not. It also knows which of a document's frontmatter
properties hold references to other documents, that a feature
blueprint's `frd` must resolve to an FRD, and that `retired` is this
framework's one status at which a document is no longer in effect.
[[module-document-core]] knows none of that on its own; this module is
how it arrives, as data for Document Core's checks to read rather than
as a check this module runs itself. A fuller vocabulary — that a
container blueprint declares un-promoted components as fenced record
blocks — is designed below and not yet registered.

It is a **plugin**: it holds no engine of its own and implements no
parsing. Every capability it needs — frontmatter extraction, schema
validation, reference resolution, span queries — comes from Document
Core, and this module's job is to register vocabulary and rules into
that host through the extension point Document Core defines. The
dependency runs one way, and the value of that is concrete rather than
aesthetic: an enterprise with a different taxonomy writes a different
profile and reuses the whole engine.

Most of this module is configuration rather than code, and as built it is
configuration almost exclusively: the vocabulary is one generated union
schema, the reference and lifecycle registrations derived from it, and a
table of slug patterns and numbered flags — all supplied by a handful of
Go files with no engine among them. The one rule left that cannot be
expressed as data is the element-record invariant, not yet implemented.
Lifecycle integrity looked like a second such rule; it turned out to be
relational rather than an invariant on this module's own data, so its
checks belong in Document Core instead (see Components, below) and this
module's part of it is the registration alone.

- **Role:** supply this framework's document vocabulary and its
  non-schema-expressible rules to Document Core.
- **Scope:** in — the one union schema over every document type, each
  type's slug pattern and numbered-ness, the reference-bearing
  properties, typed references and pairings for each type, the status
  lifecycle, the convention defaults, and (as later capabilities)
  required sections, named section sets, and element record blocks and
  their invariants. Out — all parsing, indexing, reference resolution
  and lifecycle checking, which belong to Document Core; and all
  process-level concerns, which belong to the command surface of
  [[container-spec-cli]].
- **Composed into:** [[container-spec-cli]] — a Go module in the same
  repository, imported as a package and activated by configuration. See
  that blueprint for this module's relationships to its sibling
  components.
- **Source & tests:** `cli/profile`, with tests beside their subjects
  per Go convention.

## Interface

This module's surface is deliberately narrow: it exports one
constructor returning a value satisfying Document Core's `Profile`
interface. Everything else is internal.

- **Surface:** the exported `New` constructor in `cli/profile`,
  returning a `corpus.Profile`. The registered content — types,
  schemas, patterns — is described by
  [[contract-specs-configuration]].

- **Guarantees:**
  - The module SHALL implement Document Core's `Profile` interface
    without requiring changes to that interface for framework-specific
    reasons.
  - The module SHALL register every document type this framework
    defines, so a corpus authored to the framework produces no
    `unknown-type` exclusion for a conforming document.
  - The module SHALL NOT parse document text directly; it registers
    schemas, patterns and predicates that Document Core applies.
  - WHERE a rule is expressible as a JSON Schema or a pattern, the
    module SHALL register it as data rather than implement it as code.
    [[contract-specs-configuration]] does not expose that data publicly
    in v1, so the reason is not adopter override but reviewability: a
    registered pattern is checkable against the template it mirrors,
    while the same rule buried in a predicate is not.
  - The module SHALL register this framework's Gherkin tag vocabulary —
    the feature and criterion identifier patterns, the owning-document
    tag prefix, the derivation rule linking a criterion's code and
    number to its feature's, and the qualifier tags it assigns
    meaning to. The identifier patterns SHALL be checkable against the
    generated `FeatureId` and `CriterionId` patterns
    (`cli/core/specmodel`) rather than transcribed independently.

**Built — single-schema profile.** `cli/profile.New` returns a value
satisfying `corpus.Profile`, registered under the name `spec-driven-dev`.
It supplies **one** embedded union schema
(`cli/core/schema/spec-model.schema.json`, via Go's `//go:embed`, so the
binary reads no schema file at runtime), a thin per-type table of slug
pattern and numbered-ness for every `DocumentType`, the `line_width`
convention default of 80 codepoints, and — widened by the reference
integrity feature's PREG delta — the reference, typed-reference and
lifecycle registrations below. Widened again by the Gherkin adoption
feature's PREG delta, it also supplies this framework's tag vocabulary:
the feature and criterion identifier patterns, the owning-document tag
prefix, the code/number derivation rule, and the qualifier tag table.
There is no per-type schema, no required-sections tree, no named section
sets and no element-record registration.

**This profile fills the whole of the host's extension point.**
`corpus.DocumentTypeRules` carries a type's reference-bearing properties,
its typed references and its pairings, and `corpus.Profile` carries a
`Lifecycle()` member supplying the status vocabulary, the live and
retired statuses and the successor property. This profile registers, for
every frontmatter property `schemas/frontmatter.strata` types as a
citation, that property as reference-bearing; exactly one typed
reference (a feature blueprint's `frd` must resolve to `type: frd`); no
pairing, deliberately — every relationship this framework declares has
its reverse derived from inbound edges, so
[[document-relationship-standards]]'s pairing check has nothing to pair
here; and a lifecycle naming `retired` as the only retired status, every
other status as live, and `superseded_by` as the successor property. Both
the reference set and the status vocabulary are derived from the
generated schema rather than transcribed, so a property or a status added
to the vocabulary is registered the moment it is declared. See
`cli/profile/registration.go`.

`corpus.Profile`'s fifth member, `FeatureTags()`, is populated the same
way: the feature identifier pattern `^[A-Z]+-[A-Z0-9]+-[0-9]{3}$`, the
criterion identifier pattern `^AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$`, the
owning-document tag prefix `frd-`, the `code`/`number` capture groups the
two identifier patterns share (the derivation rule that a criterion's
code and number are its feature's), and the qualifier tags `negative`,
`manual`, and the `level-` prefixed vocabulary `unit`, `integration` and
`e2e`. The two identifier patterns are authored with named capture
groups added over the generated `FeatureId` and `CriterionId` patterns,
rather than transcribed independently — stripping the names recovers
the generated pattern exactly, which is what
`TestFeatureTagPatternsMatchTheGeneratedIdentifiers` checks. See
`cli/profile/profile.go`.

The components below have been reconciled to that shape: the separate
Document Types "type registry" is gone, folded into Profile Registration;
Lifecycle Rules is gone the same way, its registration folded into
Profile Registration and its checks placed in Document Core; and Element
Records is retained as designed but marked not yet built.

**The registered patterns are grounded in this repository's corpus, not
transcribed from the prefix registry**, and in two places the two
disagree. Guides: the registry has no row at all and this repository's four
guides split two-and-two on a `guide-` prefix, so the pattern is the
type's least restrictive form. RFCs: this repository mixes a numbered
form (`rfc-0007`) with an unnumbered self-critique form, so `Numbered` is
false for the type — a "highest number in use" query would silently
ignore every `self-` document, which is worse than not offering `next-id`
for RFCs at all. Each divergence is recorded at the table in
`cli/profile/profile.go`; the Change Guidance below is what obliges a
template change to come back here.

## Components

```mermaid
C4Component
    title Components — Spec-Driven Development Profile
    Container_Boundary(profile, "Spec-Driven Development Profile") {
        Component(reg, "Profile Registration", "Go package", "The plugin entry point: the union schema, the per-type slug and numbered table, the reference/typed-reference/lifecycle registrations, the Gherkin tag vocabulary, and the convention defaults")
        Component(records, "Element Records", "Go package (planned)", "Fenced element record blocks and the agreement of an element's declarations")
    }
    Rel(reg, records, "Registers the block rules of")
```

**One component is built.** Profile Registration is the whole of the
module today — `cli/profile/profile.go` and `cli/profile/registration.go`.
Element Records is planned and carries no package, so its edge describes
the intended wiring rather than an existing one.

**Lifecycle rules are data, not a component.** The statuses, the
successor property and the pairings are values Profile Registration
supplies like everything else it registers, and the checks that read
them — that a retired document names a resolvable successor, and that a
live document citing a retired one is surfaced — live in
[[module-document-core]] alongside the checks it already runs over the
corpus. This module carries no package for them.
`reference/retired-target`, `reference/pairing-incomplete` and
`reference/type-mismatch` are those checks, and every judgement each one
makes comes from the registration below. There is no relational logic
left for a package in this module to hold, so the component is folded
into Profile Registration rather than built where it was originally
drawn.

**The separate "Document Types" component has been removed.** The
earlier design gave it a type registry holding a schema per type, and the
host it would have registered into was never built that way: Document
Core's extension point accepts exactly one
union schema instead, precisely so the host never has to choose which
schema to reach for. What remains of a per-type registry is a two-column
table — slug pattern and numbered flag — small enough that a component of
its own would be a directory containing a map literal. It lives in
Profile Registration alongside the schema it complements.

### Profile Registration

```component
entity_name: Profile Registration
slug: component-profile-registration
description: The plugin entry point supplying the union schema, the per-type slug and numbered table, the reference/typed-reference/lifecycle registrations, the Gherkin tag vocabulary, and the convention defaults
```

(Go package) — The plugin entry point supplying this framework's
vocabulary to Document Core. It is the only component the host sees, and
it holds no engine and no parsing of its own. Its value is that the
conformance surface is one file: whether this module is a well-behaved
plugin is answerable by reading it.

**Built** — `cli/profile`. `New` returns a `corpus.Profile` whose seven
members are the whole of what this framework tells the host:

| Member | Registers |
|---|---|
| `Schema()` | the embedded union schema over every document type |
| `DocumentTypes()` | per type, its slug pattern and whether it is numbered (authored); its reference-bearing properties and its one typed reference (derived from the schema and authored, respectively); no pairings |
| `Conventions()` | the `line_width` default of 80 codepoints |
| `Lifecycle()` | the whole status vocabulary, `retired` as the only retired status, every other status as live, and `superseded_by` as the successor property |
| `FeatureTags()` | the feature and criterion identifier patterns, the owning-document tag prefix `frd-`, the `code`/`number` capture groups the two identifier patterns share, and the qualifier tags `negative`, `manual`, and the `level-` prefixed vocabulary `unit`, `integration`, `e2e` |
| `RecordBlocks()` | the `component` record-block tag, validated against the generated `C4ElementRecord` definition (`#/definitions/C4ElementRecord`) |
| `Architecture()` | the architecture vocabulary: `system-blueprint`, `container-blueprint`, `module-blueprint` and `component-blueprint` at the `system`, `container`, `module` and `component` levels; `entity_name`, `system`, `modules` and `contracts_consumed` as the entity-name, system, composed-modules and consumed-contracts properties; `contract` as the contract type with `provider` as its provider property; and the `Components` roster heading with the `component` record-block tag for container and module blueprints only |

**The union schema is generated, not authored here.** It comes from
`schemas/*.strata` through `scripts/generate-schemas.sh`, which emits
`cli/core/schema/spec-model.schema.json` and the matching Go vocabulary
in `cli/core/specmodel`. This module embeds the result. Three mechanisms
keep the registered vocabulary and the generated schema from drifting:
the per-type table is keyed by the generated `DocumentType` constants
rather than by string literals, so a renamed or removed type fails to
compile here; `TestDocumentTypesCoversEveryVocabularyValue` asserts one
table entry, with a compiling slug pattern, for every value in the
generated vocabulary, so an *added* type fails the suite; and
`TestEveryCitationTypedPropertyInTheVocabularyIsRegistered` asserts, for
every document type, that every frontmatter property the schema types as
`Citation` or `[Citation]` is registered as reference-bearing by that
type and that nothing wider is registered — the same drift guard applied
to the reference set. `cli/profile/registration.go` reads the reference
properties and the status vocabulary from the decoded schema at package
initialization; only the one typed reference (`frd` → `type: frd`) and
the lifecycle's live/retired split are authored, because the schema has
no way to state either.

**The architecture vocabulary is authored, and held to the schema by a
test.** Nothing in the schema says which property is an element's name
or which heading holds a roster, so `Architecture()` transcribes them
from the vocabulary and [[c4-standards]].
`TestEveryRegisteredArchitecturePropertyIsDeclaredWhereItApplies` asserts
that the generated schema declares each registered property on every
type it applies to, so a renamed property fails the suite rather than
silently resolving nothing; `corpus.ValidateRegistration` checks the same
at activation.

**The Gherkin identifier patterns carry the same guard, structurally
rather than by derivation.** `FeaturePattern` and `CriterionPattern` are
authored as the generated `FeatureId` and `CriterionId` patterns
(`cli/core/schema/spec-model.schema.json`) plus named capture groups —
adding a name changes nothing a pattern matches — so
`TestFeatureTagPatternsMatchTheGeneratedIdentifiers` strips the names
back off and asserts equality against the generated pattern directly.
Unlike the reference set and the status vocabulary, the patterns
themselves are not derived at package initialization, because the named
capture groups the derivation rule needs have no generated counterpart
to derive from; the drift guard is the test rather than the
construction.

Required sections, named section sets and the heading-tree requirement
described in [[contract-specs-configuration]] are later capabilities.
None is registered, because Document Core's extension point has nowhere
to receive them yet — widening it is a change to the host, per the
Change Guidance below.

### Element Records

```component
entity_name: Element Records
slug: component-element-records
description: Fenced element record blocks and the rule that every declaration of one element agrees
```

(Go package) — Fenced element record blocks and the rule that every
declaration of one element agrees. Registers the
`component` fence tag so Document Core extracts and parses those blocks,
supplies the base record schema they validate against, and implements
the declaration-agreement rule from [[adr-008]] — the one rule here that
no schema can express, because it is a constraint between declarations
in different places rather than a constraint on a value.

**Partly built.** The `component` fence tag is registered by Profile
Registration through `Profile.RecordBlocks`, with its schema pointing at
the generated `C4ElementRecord` definition in the embedded union schema.
Validating a block against it and the declaration-agreement rule are not
yet built.

**Profile Registration → Element Records.** Registration is a pure
wiring layer: it exposes what it registers, and hands it to the host. No
component here reaches the host directly, so there is exactly one place
to look when the plugin contract changes. Today it is the only built
place, because it is the only built component.

## Cross-Cutting Standards

This module SHALL comply with:

- [[testing-standards]] — every component declared below carries unit
  tests covering its exported surface.

Its registered rules derive from this framework's own reference
documents —
`plugin/skills/spec/references/frontmatter-standards.md`,
`plugin/skills/spec/references/template-authoring-standards.md` and
`plugin/skills/spec/references/cross-document-references.md` — which
are shipped with the `spec` skill rather than authored as project
standards documents. The derivation is not literal transcription: the
union schema reaches this module through `schemas/*.strata`, and each
slug pattern was checked against the documents this repository actually
holds before being registered, which is where the three divergences noted
under Interface come from.

## Change Guidance

- WHEN this framework's document vocabulary changes — a type added, a
  slug pattern altered, a required section introduced — the registered
  schema or pattern here SHALL be updated in the same change as the
  template or standard that defines it. This module is a copy of rules
  owned elsewhere, and a copy that lags is worse than no check at all.
- The module SHALL NOT implement parsing, indexing or reference
  resolution. WHEN a rule appears to need one of those, the capability
  belongs in [[module-document-core]] and this module registers against
  it.
- The module SHALL NOT require a change to Document Core's `Profile`
  interface for a framework-specific reason. WHEN a rule cannot be
  expressed through the existing extension point, that is a signal the
  extension point is too narrow, and widening it is a change to the host
  reviewed on its own merits — not a special case for this profile.
- WHILE visibility is internal, this module and its in-repo consumers
  evolve together, with no compatibility window.
- **Verification gates:** `go test ./profile/...` for any change here,
  which includes the conformance test asserting that this module
  satisfies Document Core's `Profile` interface, the vocabulary-coverage
  test above, and the self-application gate in
  `cli/profile/self_application_test.go` that runs discovery and checking
  over this repository's own corpus. That last gate is the one that
  catches a registered schema drifting from the templates it mirrors —
  and it is rooted at `specs/`, not `docs/`, which is the repository's
  actual corpus root per `specs.toml`. It resolves that root from its own
  file's location and fails outright if `specs.toml` and `specs/` are not
  there, because a gate that passes over zero documents is worse than no
  gate. CI runs the full suite.

## Key Decisions

- [[adr-008]] — element record blocks and the two-schema model this
  module's Element Records component will implement when it is built.
- [[adr-012]] — the reference conventions the registered types rely on.
- [[adr-021]] — Gherkin is where features and acceptance criteria are
  authored; Profile Registration registers this framework's feature and
  criterion identifier patterns and qualifier tag vocabulary for
  [[module-document-core]]'s checks to apply.
- [[adr-015]] — Go, and the profile-as-package composition model.

## Open Questions

- [ ] `kind` is `standard` because exactly one profile exists. WHEN a
      second profile is written — an enterprise taxonomy, or a stripped
      Obsidian-vault profile — this blueprint SHOULD be reconsidered as
      `kind: pattern`, governing how any profile is built, with the
      profiles as instances. Deciding early would be inventing a pattern
      from one example.
- [ ] How much of the vocabulary should be shipped as data files versus
      compiled-in defaults. The registered form is settled and v1
      exposes none of it publicly, so this is purely a question of where
      the bytes live — but it forces the answer to the promotion
      question in [[contract-specs-configuration]], because a data file
      is most of the way to an override mechanism already. As built,
      everything is compiled in: the schema through `//go:embed` and the
      per-type table as a Go map, so the binary reads no vocabulary file
      at runtime. That is an implementation default taken to keep the
      binary self-contained, not an answer to this question.
- [ ] Whether the deferred domain rules — AC traceability, C4 diagram
      conventions, blueprint AC-coverage keys, work-order readiness —
      become components here or a second profile layered on this one.
      They are out of scope for the current feature and the answer is
      not yet forced.

## References

- `c4-guide.md` — C4 mapping rules and diagram conventions
- [[frd-spec-index]] — the requirements this module's
  registrations satisfy
- [[container-spec-cli]] — the container this module is composed into
- [[module-document-core]] — the host it registers into
- [[contract-specs-configuration]] — the configuration surface it
  populates and extends
- [[adr-008]] — the declaration-agreement rule this profile registers
- [[adr-012]] — the reference conventions its slug patterns resolve
  against

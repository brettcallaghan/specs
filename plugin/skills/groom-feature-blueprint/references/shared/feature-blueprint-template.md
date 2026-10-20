<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: plugin/skills/spec/references/feature-blueprint-template.md
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: "<Feature Name>"
description: <One-sentence summary of how the system changes to deliver
  this feature>
slug: feature-<feature-name>
type: feature-blueprint
status: draft
delivery: single-pr                  # single-pr | staged. The delivery
                                     # unit grooming plans against. WHEN
                                     # staged, enumerate each stage
                                     # boundary in Sequencing Constraints.
frd: frd-<feature-name>              # the FRD this blueprint realizes (1:1)
---

<!--
Template: Feature Blueprint
Source: This specification framework. Reference conventions follow
  [[template-authoring-standards]].

Sections (complete each, then delete its instruction block)
  - Overview
  - Element Deltas
  - Contracts
  - Tracker Reconciliation
  - Sequencing Constraints
  - AC Coverage
  - Key Decisions
  - Definition of Done
  - Open Questions
  - References

Purpose
  Describe HOW a single feature is built. A Feature Blueprint maps 1:1 to
  a Feature Requirements Document — the FRD describes WHAT users need and
  how success is verified; this document describes HOW the system changes
  to satisfy it.

  Unlike architecture blueprints (which describe what the system IS), a
  feature blueprint describes a CHANGE: which elements are added,
  modified, or removed, which are relied upon unchanged, and which
  documents must be updated when the change lands. Architecture
  blueprints remain the single source of truth for the steady state at
  all times; this document never competes with them.

Reader profile (apply with the tech-writer skill)
- Reader: the human reviewer approving the design, then the planning agent
  that derives the implementation plan.
- Knows: the FRD and the architecture blueprints this feature changes.
- Wants: what changes, element by element, and how each acceptance criterion
  is covered.
- Budget: one focused read.
- Tone: precise and diff-shaped. Reference decisions, contracts, and steady-
  state architecture; don't restate them.

Audience and altitude
  Written for two readers: the human reviewer approving the design, and
  the planning agent that will derive an implementation plan (impl-plan
  or a work-item graph) from it. It carries design intent, element
  deltas, contracts, and hard constraints — NEVER the task list itself.
  Task decomposition, ordering beyond the stated constraints, and new
  code structure (function signatures, file layouts) are the planning
  and implementing agents' choices.

Lifecycle — this document is TRANSIENT
  draft            — under authorship; Open Questions section allowed
  ready-for-review — design complete, awaiting human review
  approved         — design accepted. Implementation planning MAY begin.
                     WHILE the Open Questions section is non-empty,
                     planning SHALL NOT begin.
  in-progress      — implementation planned or in progress (this type
                     uses in-progress in place of active — it specifies
                     a change, so it is never "in effect" the way a
                     state document is). At serialization the groom-feature-blueprint
                     skill stamps the plan location into frontmatter
                     (`plan: <epic-id>` for beads, `plan:
                     plans/<feature-slug>/` for task files) so readers
                     mid-flight can find the graph without searching
                     the tracker; the pointer retires with the document.
  retired          — the feature is merged and the Definition of Done is
                     complete: every delta has been folded into the
                     owning architecture blueprint, draft contracts
                     flipped to active, and this document retired.

  Because this document retires, nothing durable may live ONLY here:
  decisions go to standalone ADRs, interface shapes go to contract
  documents, and schema/component changes are written in the owning
  blueprint's vocabulary so folding them in at retirement is mechanical.

  Timing of the documentation updates: draft contracts and ADRs are
  authored at design time, alongside this blueprint. Folding a delta
  into its owning architecture blueprint happens on the feature branch,
  within the same work order that implements the delta — documentation
  travels with the code it describes, and the main branch only ever
  sees architecture blueprints that match merged reality. Doc updates
  are neither an upfront phase (implementation may still bounce design
  questions back) nor a final pass (the step most easily dropped under
  pressure); the Definition of Done below is the completeness CHECK,
  not a batch of work left for the end.

When to use
  - Designing the implementation of a feature specified in an FRD
  - Giving planning agents everything needed to plan and execute the
    change autonomously, with an explicit escalation path (Open
    Questions) when the spec is insufficient

When NOT to use
  - Recording a decision with trade-offs (author an ADR and reference it)
  - Specifying an interface's shape and semantics (author a contract
    document in draft and reference it)
  - Describing steady-state architecture (that is the architecture
    blueprints' job — update them, or reference them)
  - Enumerating implementation tasks (that is the impl-plan / work-item
    graph derived FROM this document)

Reference syntax
  [[slug]]         — documents (blueprints, contracts, ADRs, FRDs), and
                     promoted elements by their own slug
                     ([[component-<name>]] / [[module-<name>]])
  [[container-<name>#Component Name]]
                   — a cross-reference to an un-promoted component, by a
                     wikilink heading anchor into its parent's Components
                     roster (see [[c4-standards]])
  `ElementName`    — a code-level element that is not a modeled
                     component: a schema, domain type, event, flag, table

  Delta-block headings NAME the component they change (it may be new, so
  it has no anchor yet); cross-references to a component elsewhere use
  the anchor form above, whose parent qualifier removes the ambiguity a
  bare name would carry.

Code anchors — anchors, not signatures
  Delta blocks SHOULD name real, existing code anchors (packages, key
  symbols, migration directories) so planners start in the right place.
  They SHALL NOT prescribe new function signatures, new file layouts, or
  other structure that does not exist yet — that is the implementing
  agent's choice.

Slug format
  feature-<feature-name>, lowercase kebab-case, matching the FRD:
  frd-checkout → feature-checkout

This is a living document while in flight — update it when the design
changes, and update `status` when you do. Remove all
HTML comment blocks after completing each section.
-->

# <Feature Name>

## Overview

<!--
2-4 sentences: the design approach at a glance — how the system delivers
this feature. The reader has already read the FRD; do NOT restate the
requirements. Name the major elements touched so the Element Deltas
section is navigable from here. End the overview with a single sentence
listing the delta blocks in dependency order — the one-line map
reviewers and groomers orient by.
-->

**FRD:** [[frd-<feature-name>]]

<How the system delivers this feature, at a glance: the approach chosen
and the elements it touches.>

## Element Deltas

<!--
The normative core: one delta block per element added, modified, or
removed, ordered so a reader can follow the feature's flow through the
system. Change kinds: added | modified | removed.

Block grain IS plan grain: each delta block becomes exactly one work
item during grooming, and the dependency graph can only be as parallel
as the blocks are granular. Split a block when its parts are
independently implementable and testable, even if they touch the same
element — an element MAY appear in several blocks, disambiguated by a
parenthetical focus:
  ### `DocumentStore` in [[container-api]] (modified — package compilation)
  ### `DocumentStore` in [[container-api]] (modified — membership lifecycle)
As a heuristic, split any block covering more than ~8 ACs or mixing
concerns with different upstream dependencies. Group by unit of
delivery, not by narrative ("the LSP changes" is a story, not a block).

Every delta block SHALL have a short key (PD, CLI, DOC, ...) declared
in the AC Coverage legend. The key is the block's stable identity:
grooming stamps it onto the derived work order, and re-grooming diffs
by it. Never reuse a retired key for a different delta.

Write each delta in the OWNING document's vocabulary, so that folding it
in at retirement is a mechanical copy:
  - a new component → the declaration shape of the container's (or
    module's) Components section, including its record block (name,
    slug, brief description, detail)
  - a new or changed module → the module blueprint's vocabulary; a
    change to a PUBLIC module also honors its backwards-compatibility
    change rules (see the module's Change Guidance)
  - a schema change → the container blueprint's Data section shape
    (store, what it holds, EARS invariants)
  - a new behavior or guarantee → EARS statements

Each block contains, as applicable:
  - Intent: what changes and WHY, in prose (design intent, not tasks)
  - Behavior: new or changed guarantees, in EARS
  - Data: schema deltas for stores this element exclusively owns —
    fields, constraints, and the EARS invariants maintained over them.
    Shared surfaces do NOT go here; they get a contract document.
  - Anchors: existing packages / symbols / migration dirs where the
    change lands (anchors, not signatures — see document instructions)

Deltas SHALL be independently implementable and testable wherever
possible; where two deltas cannot land independently, say so in
Sequencing Constraints rather than merging the blocks.

Consult the owning blueprint's Change Guidance section when writing
each delta: its constraints are binding on the design (lockstep-update
rules, layering boundaries), and its verification gates are what the
AC Coverage "Verified by" column maps onto. WHEN artifacts coupled by
an owning blueprint's lockstep-update rule are assigned to different
delta blocks, the blueprint SHALL state how the rule is satisfied —
in the delta or in Sequencing Constraints (e.g., "single-PR delivery;
the DOC delta lands in the same change as the CLI delta"). Absence is
a content-check failure at groom time, not a call the groomer makes.

Before setting `ready-for-review`, run the decision sweep from
[[blueprint-readiness-checklist]] over every delta: pin each silent
decision an implementer would otherwise have to make (paths and names,
defaults, scope cuts, edge-case rules) as an EARS behavior line here,
or raise it as an Open Question. "Anchors, not signatures" warns
against pinning too much; the sweep guards against pinning too little.
-->

### [[container-<name>]] (modified)

**Intent:** <what changes in this container and why.>

**Behavior:**
- WHEN <event>, `<ComponentName>` SHALL <new or changed action>.

**New components** (declared in container/module vocabulary, with a
record block):
- **<NewComponentName>** (<technology / kind>) — <brief description, as
  it will appear on diagrams>. <Detailed description: responsibilities
  and boundaries.>

  ```component
  title: <NewComponentName>
  slug: component-<new-component-name>
  description: <brief description>
  ```

**Data:**
- **<Store name>** (<technology>) — <schema delta: table/field additions
  and constraints>. Migration source of truth: <link>.
  - The container SHALL <invariant maintained over the new data>.

**Anchors:** <existing packages, symbols, or directories where this
change lands.>

### `<ComponentName>` in [[container-<name>]] (modified)

**Intent:** <what changes in this component and why.>

**Behavior:**
- The component SHALL <new or changed guarantee>.

**Anchors:** <existing code anchors.>

### Relied-upon elements (unchanged)

<!--
Elements the feature depends on but does not change. One line each: what
is relied upon and for what. This is the blast-radius map — it tells
reviewers what could break and planners what to regression-test.
-->

- [[contract-<name>]] — <what this feature depends on it for>.
- [[module-<name>]] — <what this feature relies on it for>.
- [[container-<name>#ComponentName]] — <what is relied upon>.

## Contracts

<!--
Interface surfaces this feature introduces or changes. The specification
itself is NEVER here: new contracts are authored as contract documents
in draft status alongside this blueprint, and changes to existing
contracts are drafted as revisions in the contract document itself. Name
both below and in Documentation Changes; the `plan` tracks them to
completion. One line each.

Private schema changes (a container and its own store) belong in the
owning container's delta block above, not here.
-->

- [[contract-<name>]] (new, draft) — <what it vends, to whom>.
- [[contract-<name>]] (revised, drafted in place) — <what changes for
  consumers>.

## Tracker Reconciliation

<!--
Survey the project's issue tracker while authoring: this feature's
deltas may overlap work the tracker already carries, and an
unreconciled overlap invites an agent to implement the same behavior
twice, divergently. WHERE the project has no tracker, or the survey
finds no overlapping issues, state "None." — the absence is
information.

One line per overlapping issue, classified as exactly one of:

  subsumed      — this feature's deltas supersede the issue (including
                  tracked defects the feature rolls in and fixes); name
                  the superseding delta by key. Grooming blocks the
                  issue on the superseding work order with a
                  supersession note, and the Definition of Done closes
                  or re-scopes it.
  prerequisite  — the issue must land first (e.g., an open defect in
                  code a delta extends that the feature must not build
                  on top of); grooming makes it a dependency edge.
  out-of-scope  — adjacent work this feature must not silently absorb,
                  regress, or duplicate; state the boundary. Grooming
                  carries it as a coordination note on affected items.

WHEN a delta's anchors point at code containing a tracked open defect,
the relationship SHALL be recorded here under one of the three
classifications. IF an overlapping issue cannot be classified, THEN it
goes to Open Questions — the same escalation path as any other design
gap.
-->

- <tracker-id> — subsumed by the <KEY> delta: <what supersedes it>.
- <tracker-id> — prerequisite: <why it must land before <KEY>>.
- <tracker-id> — out-of-scope: <the boundary this feature respects>.

## Sequencing Constraints

<!--
Hard ordering constraints that are DESIGN decisions — the planner cannot
be allowed to derive them wrong. Migration ordering, backward
compatibility windows, feature flags, contract-first requirements. In
EARS, one per line.

Ordering between this feature's own deltas belongs here too, not in
Intent prose. IF delta B extends state or interfaces introduced by
delta A, THEN state it as a constraint ("The <A> delta SHALL land
before the <B> delta"). Intent prose MAY repeat it for readability,
but only constraints listed here are guaranteed to become dependency
edges in the derived plan.

This is not a phasing plan and not a task list: state only constraints;
the planning agent derives the actual graph from these plus the deltas.
IF there are no such constraints, state "None." rather than deleting the
section — the absence is information.

The frontmatter `delivery` declaration is part of this section's
contract: `single-pr` (the default grooming assumes when absent) or
`staged`. WHEN delivery is staged, each stage boundary SHALL appear
here as a constraint — the reviewer approves the delivery unit as part
of the design; grooming reads it rather than assuming it.
-->

- The <schema migration> SHALL deploy before <the element that reads it>.
- WHILE <rollout is incomplete>, the feature SHALL be gated behind
  `<flag-name>`.
- The system SHALL maintain backward compatibility with <surface> until
  <condition>.

## AC Coverage

<!--
The traceability map: every acceptance criterion in the FRD SHALL appear
in this table at least once. This is how a planning agent verifies the
design is complete before planning, and how a verification agent checks
the PR afterward. WHEN an AC has no satisfying delta, the design is
incomplete — resolve it before status may become approved.

Satisfied by: the delta block(s) that realize the AC. Every row SHALL
name at least one delta block from Element Deltas — this is what makes
the table mechanically translatable into work orders:
  - WHEN an AC is delivered by relied-upon unchanged machinery, name
    the delta that feeds it and append the mechanism: "CLI delta, via
    relied-upon `Compiler`".
  - WHEN an AC is guaranteed by a sequencing constraint, name the
    constrained delta(s) and note the constraint parenthetically.
  - A row naming only a relied-upon element, a sequencing constraint,
    or a Definition of Done entry is invalid — grooming bounces it.
A short-name legend above the table SHALL assign each delta block a
unique key (e.g. "PD = `PackageDiscovery` delta") and the Satisfied-by
column SHALL use the keys. Every legend entry SHALL expand to a delta
block and every delta block SHALL appear in the legend — the legend is
the delta key table grooming stamps onto work orders (see Element
Deltas).

Proving delta: WHEN a row names more than one satisfying delta,
exactly one SHALL be marked as the proving delta — the one whose tests
carry the AC citation at the declared level — with the others listed
as participating:

  | AC-PKG-006.1 | SR, **PD (proves)** | unit |

For `manual` ACs, the proving delta is the one after which the
behavior is end-to-end verifiable. Grooming places the citation
obligation on the proving item only, and generates verification-timing
notes for participating items that become ready before the proving
item.

Verified by:  the level at which the AC is proven —
              unit | integration | e2e | manual.

This table is design-time intent; the durable record is the tests
themselves, which cite their ACs in citation comments per
[[ac-traceability-standards]]. The two are reconciled in the Definition
of Done below.
-->

Legend: <KEY> = `<ComponentName>` delta; <KEY2> = [[container-<name>]]
delta (— <focus>); ...

| AC | Satisfied by | Verified by |
|----|--------------|-------------|
| AC-<PREFIX>-001.1 | <KEY> | unit |
| AC-<PREFIX>-001.2 | <KEY>, **<KEY2> (proves)** — via [[contract-<name>]] | integration |
| AC-<PREFIX>-002.1 | <...> | e2e |

## Key Decisions

<!--
Standalone ADRs only — this document retires, so a decision recorded
inline dies with it. WHEN a design choice with real alternatives is made
while authoring this blueprint, author the ADR and reference it here
(and from the owning architecture blueprint at retirement).
-->

- [[adr-<nnn>]] — <one-line statement of the decision>.

## Definition of Done

<!--
The retirement contract, populated per-feature from the deltas above.
Planning agents SHALL include these as work orders — the feature is not
done when the code merges; it is done when the documentation set is
truthful again. Keep one checkbox per affected document.
-->

- [ ] Every delta above is folded into its owning blueprint:
  - [ ] [[container-<name>]] — Components / Data / diagrams updated
  - [ ] [[module-<name>]] — Components / Interface / Versioning updated
  - [ ] [[component-<name>]] — <sections affected>
- [ ] [[contract-<name>]] status: draft → active
- [ ] New ADRs are referenced from the owning blueprints' Key Decisions
- [ ] Every AC in the coverage table is verified at its declared level
  and cited from its tests per [[ac-traceability-standards]]
- [ ] Every behaviour this blueprint decided that no FRD criterion asks
  for (a refusal, an error path, a registration) is lifted into the FRD
  as a criterion and cited from its test; it must not retire with this
  document
- [ ] Every issue listed as subsumed in Tracker Reconciliation is
  closed or re-scoped (omit WHERE that section states "None.")
- [ ] No live document references this blueprint (`specs refs
  <this-slug>` names no live document), per
  [[document-relationship-standards]]
- [ ] This blueprint's status → retired
- [ ] With this blueprint already retired, `specs check --strict` over
  the corpus is green — a gate run before retirement tests a state that
  will never exist again

## Open Questions

<!--
Draft-mode escalation path. WHEN a planning or authoring agent finds the
spec insufficient, the gap is recorded here and bounced back to the
human loop — never guessed at. WHILE this section contains any open
question, implementation planning SHALL NOT begin. WHEN a question is
resolved, record it as an ADR or fold it into the relevant section, then
delete it here. WHEN status becomes approved, remove the section or
state "None." explicitly — the absence of questions is information.
-->

- [ ] <open question> — <what is blocked on it, who decides>

## References

<!--
Optional. What a reader needs next, one line each: [[slug]] for
documents in this corpus, a URL for anything outside it.

The system, container, module and component definitions that delta
blocks are written against are in this skill's `shared/c4-standards.md`.
Read it while authoring, but do not cite it here: it is not a document
in your corpus.
-->

- <Additional references>

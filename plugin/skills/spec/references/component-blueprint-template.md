---
title: "<Component Name>"
entity_name: <Component Name>        # the element, as diagrams draw it
technology: <language, framework or package type>
                                     # optional; the technology line drawn on
                                     # diagrams — omit WHEN not meaningful
description: <What it owns>. Change it to <kinds of change>.
                                     # rule 7; the brief description, shown
                                     # on diagrams
slug: component-<component-name>
type: component-blueprint
status: draft
kind: <standard | pattern>           # see instructions below
contracts_consumed: []               # contract slugs this component depends on
source_paths: []                     # repo-relative source roots; required
                                     # WHEN status is active; never overlaps
                                     # a sibling's (rule 9)
test_paths: []                       # repo-relative test roots; required
                                     # WHEN status is active (rule 9)
---

<!--
Template: Component Blueprint
Source: Architecture documentation framework. C4 mapping rules are
  defined in `shared/c4-standards.md`. Load it,
  `shared/blueprint-content-standards.md`, and
  `shared/component-design-guide.md` before authoring. "Rule N" below
  cites a rule in the content standards.

Sections (complete each, then delete its instruction block)
  - Responsibility
  - Interface
  - Design
  - Pattern guidance
  - Cross-cutting Standards
  - Change guidance
  - Key decisions
  - Open questions
  - References

Purpose
  Document one COMPONENT: an internal building block of a SINGLE
  container, with a defined interface and — as a leaf — no internal
  components of its own. A component changes in lockstep with its
  container: it is NOT vended or versioned independently. It is the unit
  of change within a container.

  WHEN a unit is shared or vended across containers on their own
  schedules (so versioning and compatibility become concerns), OR it has
  internal components of its own, it is a MODULE — use a module blueprint.
  The dividing line is the same one the framework draws for contract
  visibility: does it cross a boundary of independent change? See
  `shared/c4-standards.md`.

Reader profile (apply with the tech-writer skill)
- Reader: an engineer or agent changing this component inside its container.
- Knows: the host container, not this component's internals.
- Wants: its responsibilities, its interface, and the change guidance feature
  work must respect.
- Budget: one focused read.
- Tone: matter-of-fact and exact. Name the file or type a rule depends on
  and link it; don't describe what the code shows (rules 2 and 3). Decisions
  and their alternatives belong in ADRs.

When to use — the trigger is the promotion predicate
  A component earns its own blueprint WHEN it has an interface other
  components depend on AND enough internal behavior that a change to it
  is a self-contained unit of work — a seam an agent can reason about
  locally — AND it can be documented so as to stand alone. A component
  that cannot be described context-independently should stay a bullet
  (with a record block) in its container's Components section, not
  become a blueprint.

  kind:
  standard — the default: an internal component of one container (a
             pricing calculator, a token verifier, a form validator).
  pattern  — a component type built repeatedly to the same design, where
             the blueprint governs how each new instance is built and is
             authoritative for every instance's conformance (one
             blueprint for "data-source adapter", many adapters).

When NOT to use
  - A unit shared or vended across containers (versioning becomes a
    concern), OR one with internal components of its own → that is a
    module; use a module blueprint.
  - A component that cannot be described context-independently, or a
    trivial one (glue, a single helper, no defined interface) — a bullet
    with a record block in the container's Components section is enough.
  - A capability with runtime elements of its own (a service, a worker,
    a queue) — that is a container or system (container blueprint /
    system blueprint).
  - A package published to a registry. It is a public module; use a
    module blueprint.
  - A unit with parts that each need their own interface or change
    guidance. It is a module; use a module blueprint.

Edges
  The host [[container-<name>]] blueprint owns this component's edges to
  its siblings; point to it (rule 8).

Keep the blueprint near 250 lines (rule 13).

Lifecycle migration
  WHEN a component becomes shared across containers (an internal module)
  or published to a registry (a public module), THEN it has become a
  module: author a module blueprint and fold this content into it,
  marking this one retired (superseded_by the module).
  WHEN a component gains internal components of its own, THEN it has
  likewise become a module — the same migration applies.
  WHEN a component gains runtime elements of its own (a service, a
  worker, a queue), THEN it has become a container or system: author the
  appropriate blueprint, mark this one retired, and point superseded_by
  at the successor.

Slug format
  component-<component-name>, lowercase kebab-case.
  Example: component-token-verifier

Status vocabulary
  draft | ready-for-review | approved | active | retired

This is a living document — update it when the component changes. Remove
all HTML comment blocks after completing each section. WHILE status is
draft, the Open questions section MAY contain open questions or the
literal text "None."; remove the section when the document becomes
active.
-->

# <Component Name>

## Responsibility

<!--
One paragraph on why this component exists, then the four bullets below.

- Owns: the responsibilities and data this component is the sole
  authority for.
- Hides: the design decision likely to change that no other element may
  depend on. Required, and "none" is not an answer: an element that hides
  no decision is a candidate to merge into a neighbor.
- Does not own: each excluded responsibility, with its owner's [[slug]]
  (rule 8).
- Composed into: the host container.

Source and test roots go in the source_paths and test_paths frontmatter,
not in the body.
-->

<Why this component exists.>

- **Owns:** <the responsibilities and data it is the sole authority for>.
- **Hides:** <the design decision likely to change that no other element
  may depend on>.
- **Does not own:** <excluded responsibility> — [[<owner-slug>]] owns it.
- **Composed into:** [[container-<name>]] — <how it is composed into the
  container>; see it for this component's relationships.

## Interface

<!--
What the host container's other components may depend on. An interface
consumed ACROSS containers is a contract on a shared unit: that unit is a
module, not a component (module blueprint), and its interface may warrant
a contract document.
-->

### Surface

<!--
The path or symbol that defines the interface (rule 3).
-->

`<path or symbol>`

### Provides

<!--
The postconditions consumers rely on, in EARS.
-->

- The component SHALL <postcondition consumers rely on>.
- The component SHALL NOT <thing consumers must not have to defend
  against>.

### Requires

<!--
The abstractions and contracts this component depends on, with their
preconditions, in EARS. Name an abstraction or a contract, never a sibling
component: edges belong to the host (rule 8).
-->

- WHEN <operation> is called, <abstraction or [[contract-<name>]]> SHALL
  <precondition this component depends on>.

## Design

<!--
How the component is built to stay sound as it changes.
-->

### Patterns and principles

<!--
For each pattern the code uses: its roles mapped to this component's
types, why it fits here, and what would break it.
-->

- <Pattern> — <roles mapped to this component's types>; <why it fits
  here>; <what would break it>.

### Invariants

<!--
EARS statements that hold across all operations, each ending "Enforced
by: <test, check, lint, or review only>" (rule 5).
-->

- The component SHALL <invariant that holds across all operations>.
  Enforced by: <test, check, lint, or review only>.

### Extension points

<!--
Where the component is designed to grow, the intended way in, and the
wrong way to extend it. If it has none, write "None: <why it is closed>"
(rule 12).
-->

- <Where it grows> — <the intended way in>. Not by <the wrong way to
  extend it>.

## Pattern guidance

<!--
kind: pattern only — delete this section otherwise.

Authoritative instructions for building a new instance of this component
type. Write conformance rules in EARS so an instance can be reviewed
against them, and keep a registry of instances.

WHEN an authoritative pattern document already exists outside the
framework (a pre-existing architecture or tenets doc), this blueprint
MAY wrap it rather than rewrite it: declare the pattern, its instances,
and conformance gates here, and reference the existing document (by
repo-relative path, per [[template-authoring-standards]]) as the detailed
specification.

-->

### Building a new instance

<!--
Building a new instance follows the extension point in Design. Refer to it
rather than repeating it.
-->

See the extension point in Design.

### Conformance

<!--
The rules every instance meets, in EARS.
-->

Every instance:

- SHALL <structural or behavioral rule all instances meet>.
- SHALL NOT <divergence that would break the pattern>.

### Instances

<!--
One line per instance: where it lives and how it differs (rule 11).
-->

- **<InstanceName>** — `<path>`; <how it differs>.

## Cross-cutting Standards

<!--
Standards this component complies with, by reference. Platform choices
that apply everywhere are mentioned in prose or standards docs, never
modeled as elements on this component's diagrams — see `shared/c4-standards.md`.
-->

This component SHALL comply with:

- [[<standard-slug>]] — <one-line scope>.

## Change guidance

<!--
Rules for feature work touching this component: the constraints a
feature blueprint — and the planning agent downstream of it — must
respect when changing this component. Write normative rules in EARS.

This section carries ARCHITECTURAL constraints: lockstep-update rules,
layering boundaries, regeneration obligations. Invariants go in Design.
It is not a task list or a commands reference — reference the repository's build and
contributor docs for procedure rather than duplicating them (duplication
drifts). Constraints that span multiple elements belong in a cross-
cutting requirements document, not here.
-->

- WHEN <a kind of change is made>, <the artifacts that must change
  together> SHALL be updated in the same change.
- <A layering or boundary rule any change must respect, in EARS.>

### Anti-patterns

<!--
Each a wrong change that caused a real defect, and the defect it caused.
-->

- <The wrong change> — <the defect it caused>.

### Verification gates

<!--
List the gates in two tiers (rule 10):
  - Tier 1, always: the isolation tests under test_paths, as runnable
    targets (test paths, ids, or task names).
  - Tier 2, WHEN <condition>: each wider suite, with the change that
    requires it, such as "WHEN the public interface changes, run the
    integration suite".
CI and downstream agents still run the full suite. Feature blueprints
map their AC coverage "Verified by" column onto these gates.
-->

- Tier 1 (always): <isolation tests under test_paths, as runnable targets>.
- Tier 2 (WHEN <condition>): <the wider suite that condition requires>.

## Key decisions

<!--
One line per significant decision: restate the outcome and reference
the ADR. The trade-off analysis lives in the ADR, not here.
-->

- [[adr-<nnn>]] — <one-line statement of the decision>.

## Open questions

<!--
Optional, draft only. Track unresolved questions while the document is
a draft. WHEN a question is resolved, fold the answer into the relevant
section (or, for an architecture decision, record it as an ADR) and
delete it here. WHILE status is draft, the section MAY instead state the
literal text "None." (no checkbox) — the absence of questions is
information. Remove the whole section when the document leaves draft.
-->

- [ ] <open question> — <what is blocked on it, who decides>

## References

<!--
Optional. What a reader needs next, one line each: [[slug]] for
documents in this corpus, a URL for anything outside it. Limit it to the
host, Contracts, ADRs, and guides (rule 8). Don't list sibling elements
or FRDs. Remove the section when there is nothing to cite.

C4 mapping rules and diagram conventions are in this skill's
`shared/c4-standards.md`. Read it while authoring, but do not cite it
here: it is not a document in your corpus.
-->

- [[container-<name>]] — the container this component is composed into
- <Additional references>

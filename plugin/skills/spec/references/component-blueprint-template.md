---
title: "<Component Name>"
entity_name: <Component Name>        # the element, as diagrams draw it
technology: <language, framework or package type>
                                     # optional; the technology line drawn on
                                     # diagrams — omit WHEN not meaningful
description: <One-sentence summary of the component and why it warrants
  standalone documentation — its brief description, shown on diagrams>
slug: component-<component-name>
type: component-blueprint
status: draft
kind: <standard | pattern>           # see instructions below
contracts_consumed: []               # contract slugs this component depends on
source_paths: []                     # repo-relative source roots
test_paths: []                       # repo-relative test roots
---

<!--
Template: Component Blueprint
Version: 1.10
Source: Architecture documentation framework. C4 mapping rules are
  defined in `shared/c4-standards.md` — load it before authoring.

Sections (complete each, then delete its instruction block)
  - Overview
  - Interface
  - Pattern Guidance
  - Cross-Cutting Standards
  - Change Guidance
  - Key Decisions
  - Open Questions
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
  reader: an engineer or agent changing this component inside its container.
  knows: the host container, not this component's internals.
  wants: its responsibilities, its interface, and the change guidance feature
      work must respect.
  budget: one focused read.
  tone: matter-of-fact and exact: name functions, types, and files. Decisions
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

Edges — light on outward edges, and on code
  A component does NOT document its outward relationships (to the sibling
  components of its host container): those edges are owned by the host
  [[container-<name>]] blueprint. Point to the host instead. It stays light
  on its own internals too — the code is the source of truth; this
  blueprint mentions structure only where a rule depends on it. A
  component sits at its altitude: light on the level above (edges) and
  the level below (code).

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
draft, the Open Questions section MAY contain open questions or the
literal text "None."; remove the section when the document becomes
active.
-->

# <Component Name>

## Overview

<!--
What this component is, why it exists as a documented unit, and where it
sits. Cover, in 1-2 paragraphs plus the facts below:
  - its role — what it does within the container that hosts it;
  - its scope — what is in scope and, explicitly, what is out of scope.

Source and test roots go in the source_paths and test_paths frontmatter,
not in the body.

Relationships: point to the host [[container-<name>]]; do NOT restate the
edges to sibling components here (see the instruction block above). Keep
internals light too — the code is the source of truth.
-->

- **Role:** <what this component does within its host container>
- **Scope:** in — <what this component owns>; out — <what it deliberately
  does not do>.
- **Composed into:** [[container-<name>]] — <how it is composed into the
  container>; see it for this component's relationships.

## Interface

<!--
The component's public surface — what its container's other components
may depend on. Point at the source of truth for the shape (type
definitions, exported API) rather than duplicating it. State the
semantic guarantees consumers can rely on, in EARS.

An interface consumed ACROSS containers is a contract on a shared unit —
that unit is a module, not a component (module blueprint), and its
interface may warrant a contract document.
-->

- **Surface:** <link to exported API / types>
- **Guarantees:**
  - The component SHALL <semantic guarantee consumers can rely on>.
  - The component SHALL NOT <thing consumers must not have to defend
    against>.

## Pattern Guidance

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

**Building a new instance:**

1. <step>
2. <step>

**Conformance — every instance:**

- SHALL <structural or behavioral rule all instances meet>.
- SHALL NOT <divergence that would break the pattern>.

**Instances:**

- **<InstanceName>** (<container or repo>) — <target/variant>.

## Cross-Cutting Standards

<!--
Standards this component complies with, by reference. Platform choices
that apply everywhere are mentioned in prose or standards docs, never
modeled as elements on this component's diagrams — see `shared/c4-standards.md`.
-->

This component SHALL comply with:

- [[<standard-slug>]] — <one-line scope>.

## Change Guidance

<!--
Rules for feature work touching this component: the constraints a
feature blueprint — and the planning agent downstream of it — must
respect when changing this component. Write normative rules in EARS.

This section carries ARCHITECTURAL constraints: invariants, lockstep-
update rules, layering boundaries, regeneration obligations. It is not a
task list or a commands reference — reference the repository's build and
contributor docs for procedure rather than duplicating them (duplication
drifts). Constraints that span multiple elements belong in a cross-
cutting requirements document, not here.

WHEN prose anywhere in this blueprint mentions a known defect or other
open tracker issue, it SHOULD cite the tracker ID in greppable form
(e.g. `proj-ab12`) — feature-blueprint authors harvest these IDs when
reconciling their deltas against the tracker.

End with a Verification gates entry. Name the SPECIFIC tests that MUST
run when this component changes — real runnable targets (test paths,
ids, or task names), not "the suite." This is the minimal, targeted set
an agent changing the component runs to prove the change before handoff:
running the whole suite for every change does not scale (a parallel sprint run
slows to a crawl), so the blueprint is the authority on what THIS
component's change must run. CI and downstream agents still run the full
suite. Feature blueprints map their AC Coverage "Verified by" column onto
these gates.
-->

- WHEN <a kind of change is made>, <the artifacts that must change
  together> SHALL be updated in the same change.
- <A layering or boundary rule any change must respect, in EARS.>
- **Verification gates:** <the specific tests that MUST run when this
  component changes — runnable targets (paths / ids / task names); the
  minimal targeted set an agent runs before handoff, not the full suite
  (which CI runs)>.

## Key Decisions

<!--
One line per significant decision: restate the outcome and reference
the ADR. The trade-off analysis lives in the ADR, not here.
-->

- [[adr-<nnn>]] — <one-line statement of the decision>.

## Open Questions

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
documents in this corpus, a URL for anything outside it. Remove the
section when there is nothing to cite.

C4 mapping rules and diagram conventions are in this skill's
`shared/c4-standards.md`. Read it while authoring, but do not cite it
here: it is not a document in your corpus.
-->

- [[container-<name>]] — the container this component is composed into
- <Additional references>

---
title: Document Type Boundary Cases
description: How to resolve the document-type choices the ordered questions
  in the spec skill do not settle — component vs. module, private schema vs.
  Contract, what may live only in a feature blueprint
slug: guide-document-type-boundaries
type: guide
status: active
---

# Document type boundary cases

The `spec` skill's ordered questions pick the document type on the common
path. The cases below are the ones where two types both look plausible, or
where the obvious answer is wrong.

- A single unit of executable work is a **Work Order**, but you don't author
  it through this skill. Grooming generates it to the Definition of Ready,
  from an approved Feature Blueprint (`groom-feature-blueprint`) or from an
  existing backlog item (`groom-beads`). Author the FRD and the blueprint that
  precede it. The Work Order is downstream of both. The plan graph that groups
  a feature's Work Orders stays a derived artifact, not a spec type.
- Component vs. module: the test is the boundary of independent change, the
  same test that separates a private schema from a contract. An internal
  building block of one container that changes in lockstep with it is a
  component. A unit that's shared or published across containers, or that's
  composite with internal components, is a module. When you're unsure,
  default to a component. Promote it to a module when it's first shared or
  published, or when it grows internal structure.
- A question that looks "architectural" is often a requirements question in
  disguise. Get the product and requirements context first, because it often
  resolves or constrains the technical choice.
- ADRs record *why an approach was chosen*. Requirements documents define
  *what must be satisfied*. Policy belongs in a requirements document, not an
  ADR. Trade-off analysis belongs in an ADR, not a requirements document.
- Don't nest sub-features inside an FRD. Author each sub-feature as its own
  FRD. Split features when each one delivers value independently, and keep
  them together when their requirements break without each other.
- Obvious choices with no real alternative don't need an ADR. Implementation
  details within a decided technology don't need a spec document at all.
- A private schema (a container and the database only it uses) is a
  contract, but not a Contract document. It's documented in the Container
  Blueprint's Data section and promoted to a Contract only when a second
  independently built consumer appears.
- Blueprints restate a decision in one line and reference the ADR. The
  trade-off analysis lives only in the ADR. Blueprints in draft may carry an
  Open questions section. Resolved questions become ADRs or content, and never
  stay as answered questions.
- Architecture blueprints and Contracts describe the system **as it is**, and
  they're updated in the same change that alters the architecture. FRDs and
  ADRs capture intent and decisions at a point in time. The Feature Blueprint
  is the deliberate exception: a **transient change document** that
  describes a delta and gates implementation planning (approved status, empty
  Open questions). It retires once its deltas are folded into the owning
  architecture blueprints at merge.
- Nothing durable may live only in a Feature Blueprint, because the Feature
  Blueprint retires. Decisions go to standalone ADRs. New or changed
  interfaces are authored as draft Contract documents alongside the Feature
  Blueprint. Schema and component changes are written in the owning
  blueprint's vocabulary, so folding them in at retirement is mechanical.

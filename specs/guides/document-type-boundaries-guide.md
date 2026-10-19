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
path. These are the cases where two types both look plausible, or where
the obvious answer is the wrong one.

- A single unit of executable work is a work order, but you don't author
  it through this skill. Grooming generates it to the Definition of Ready,
  from an approved feature blueprint (`groom-feature-blueprint`) or an
  existing backlog item (`groom-beads`). Author the FRD and blueprint that
  precede it. The work order is downstream of both. The plan graph that
  groups a feature's work orders stays a derived artifact, not a spec
  type.
- Component vs. module: the test is the boundary of independent change,
  the same one that separates a private schema from a contract. An
  internal building block of one container that changes in lockstep with
  it is a component. A unit shared or published across containers, or
  composite with internal components, is a module. When you're unsure,
  default to component, and promote it to a module when it's first
  shared or published, or when it grows internal structure.
- A question that looks "architectural" is often a requirements question.
  Get the product and requirements context first, because it often
  resolves or constrains the technical choice.
- ADRs record why an approach was chosen. Requirements documents define
  what must be satisfied. Policy belongs in a requirements document, not
  an ADR. Trade-off analysis belongs in an ADR, not a requirements
  document.
- Don't nest sub-features inside an FRD. Author each sub-feature as its
  own FRD. Split them when each feature delivers value independently, and
  keep them together when the requirements break without each other.
- Obvious choices with no real alternative don't need an ADR, and
  implementation details within a decided technology don't need a spec
  document at all.
- A private schema (a container and the database only it uses) is a
  contract, but not a Contract document. Document it in the container
  blueprint's Data section, and promote it to a Contract only when a
  second independently built consumer appears.
- Blueprints restate a decision in one line and reference the ADR. The
  trade-off analysis lives only in the ADR. A draft blueprint may carry
  an Open questions section. Resolved questions become ADRs or content
  and never stay as answered questions.
- Architecture blueprints and contracts describe the system as it is, and
  they're updated in the same change that alters the architecture. FRDs
  and ADRs capture intent and decisions at a point in time. The feature
  blueprint is the deliberate exception. It's a transient change
  document: it describes a delta, gates implementation planning (approved
  status, empty Open questions), and retires once its deltas are folded
  into the owning architecture blueprints at merge.
- Nothing durable may live only in a feature blueprint, because the
  blueprint retires. Decisions go to standalone ADRs. New or changed
  interfaces are authored as draft contract documents alongside the
  blueprint. Schema and component changes are written in the owning
  blueprint's vocabulary, so folding them in at retirement is mechanical.

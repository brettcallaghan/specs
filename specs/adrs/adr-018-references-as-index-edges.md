---
title: "ADR-018: References Are an Edge Collection in the Index"
description: The index records every cross-document reference as a row in
  one top-level collection carrying its own span, origin and resolution,
  and derives both directions from it, rather than embedding reference
  arrays in each document's entry
slug: adr-018
type: adr
status: approved
---

# ADR-018: References are an edge collection in the index

**Status:** approved  
**Date:** 2026-08-07

## Context

[[frd-spec-reference-integrity]] requires the emitted index to record three
things for every corpus member:

- the `[[slug]]` references parsed from its body, with the span of each
  occurrence (AC-SPECREF-001.1)
- the references held by its reference-bearing frontmatter properties
  (AC-SPECREF-001.7)
- its inbound references, derived from every other member's outbound ones
  (AC-SPECREF-001.3)

It also requires that a reference resolving to nothing is recorded rather
than dropped, because indexing never fails (AC-SPECREF-001.6). That FRD
deliberately places the shape of this data out of scope and defers it
here.

The index is `SpecIndex`, declared in `schemas/index.strata` and generated
per [[adr-016]]. Today it carries `entries` and `exclusions` and nothing
else.

Two constraints bear on the shape. First, AC-SPEC-002.3 gives every entry a
workspace identifier precisely so that indexes emitted by separate
repositories merge into one graph. It was raised during review of
[[frd-spec-index]], explicitly as architectural input for later, that
references occurring inside document bodies behave like edges rather than
like members of the objects they connect. Second, [[container-spec-cli]]
pins the index as additive under `schema_version` 1.0: a property added
alongside the existing shape doesn't bump the version, while changing what's
already there does.

## Decision

We will record references as one top-level `references` collection on
`SpecIndex`, with one row per occurrence. Each row carries the path it
occurs in, the slug and anchor halves of the reference, its origin (a
document body, or a named frontmatter property), its `Span`, and its
resolution outcome. Inbound and outbound directions are both derived by
indexing that one collection when the index is read. Neither direction is
stored a second time.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Outbound array embedded on each `IndexEntry`, inbound derived at load | One document's outbound references are a property access<br>Smallest change to the existing model | Inbound edges are absent from the artifact, so every consumer re-derives them<br>Merging two workspaces' indexes means rewriting entries rather than appending rows<br>A reference's own attributes (span, origin, resolution) sit on the source endpoint, which is only one of the two documents that care | The artifact is read by consumers that aren't this CLI (AC-SPEC-002.3's merged graph, the Architecture Catalogue in [[container-spec-cli]]); making each of them re-derive the reverse is exactly what the index exists to prevent |
| Outbound and inbound arrays embedded on each `IndexEntry` | Both directions readable with no indexing at all | Stores the reverse of every edge twice, so the two copies can disagree<br>Every edge is written into two entries, so an emitter bug is legible in one place and wrong in another<br>Contradicts the declare-once/derive-the-reverse rule [[document-relationship-standards]] rests on, in the very artifact that rule cites as its authority | A framework that tells authors not to mirror an edge can't mirror it in its own index |
| One top-level `references` edge collection; both directions derived at load | One row per occurrence, so a repeated reference is two rows with two spans (AC-SPECREF-001.5) rather than a special case<br>Cross-workspace merge is a row append<br>An unresolved reference is a row with a resolution state, so AC-SPECREF-001.6 needs no second collection<br>A reference's attributes belong to the edge, which is what they describe | "What does this document reference" costs an index build at load rather than a property access<br>The collection is the largest part of the index by row count: roughly 256 rows over 29 documents in this repository today | N/A (selected) |

## Consequences

**Positive:**

- `specs refs` answers two questions: what points at this document, and
  what this document points at (AC-SPECREF-003.1, AC-SPECREF-003.2). One
  list serves both, so the two answers can't disagree.
- Merging indexes across workspaces is appending rows. An edge from
  workspace A naming a slug declared in workspace B needs no rewrite of
  either repository's entries. That's the case AC-SPEC-002.3 exists for,
  and the one the FRD names as a known follow-on.
- Ordering is a property of one list. Sorting rows by path and then by
  position within the document satisfies two requirements in the same
  rule: AC-SPECREF-001.5's ordering and AC-SPEC-009.7's byte-identical
  output. There's no per-entry ordering to keep consistent.
- `schema_version` stays `1.0`. This is an addition alongside the existing
  shape, which [[container-spec-cli]]'s additive rule permits at the same
  version. `schemas/index.strata` already records that no property on
  `IndexEntry` forecloses it.
- A row is the natural home for the facts that belong to neither
  endpoint: where the reference occurs, whether it came from prose or
  from a named frontmatter property, and what it resolved to.

**Negative:**

- A consumer wanting one document's outbound references scans or indexes
  the collection rather than reading a property. For the CLI, this is a map
  built once when the index is opened. For a third-party consumer reading
  the raw JSON, it's work this shape imposes and the embedded shape
  wouldn't.
- The collection dominates the index by row count, so the emitted file
  grows by roughly an order of magnitude more rows than entries. It
  remains one file and one write.
- Nothing in the artifact prevents an edge whose source path names no
  entry. The invariant that every row's source is a member is the
  emitter's to hold rather than the schema's, because a JSON Schema
  can't express a cross-collection reference.

## References

- [[frd-spec-reference-integrity]]: the requirements this shape serves,
  and the document that defers the shape here
- [[frd-spec-index]]: AC-SPEC-002.2 and AC-SPEC-002.3, the
  versioned schema and the workspace identifier this shape is chosen
  against
- [[container-spec-cli]]: the additive-under-1.0 rule, and the open
  question on promoting the index to a contract document
- [[adr-016]]: Strata is the single source, so this shape is declared
  in `schemas/index.strata` rather than assembled in the emitter
- [[document-relationship-standards]]: declare once and derive the
  reverse, the rule the mirrored-arrays option contradicts
- [[adr-009]]: nodes-versus-edges ownership, the earlier decision this
  one extends from documents to the index that describes them

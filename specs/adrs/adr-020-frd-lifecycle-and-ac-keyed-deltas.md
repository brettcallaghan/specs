---
title: "ADR-020: FRDs Evolve in Place; Deltas Are Keyed on AC Identifiers"
description: An FRD is a living document edited in place under a same-feature test keyed on its slug, never amended by a successor FRD; changes between versions are diffed by AC identifier, not by line
slug: adr-020
type: adr
status: ready-for-review
---

# ADR-020: FRDs Evolve in Place; Deltas Are Keyed on AC Identifiers

**Status:** ready-for-review
**Date:** 2026-08-13

## Context

The framework has never decided how an FRD changes after it's first
written. Two questions were open. First, does a change to a feature's
requirements produce a new document or edit the existing one? Second,
whichever document carries the change, how does a reader or a tool find
what changed between versions? The CLI report/guard work is the first
change that needs an answer, because it revises acceptance criteria (ACs)
on an FRD that's already active. Without a settled process, every later
revision decides the same two questions again from scratch.

The repository already names its documents differently: ADRs are numbered,
and FRDs are named by feature slug. Nothing stated what that difference
implies for change, so nothing stopped a later revision from treating an
FRD as a record to append to instead of a statement to keep current.

**Scope.** This decision covers only the FRD lifecycle. How an ADR changes
is a separate question: whether it's amended in place or superseded, and
which kinds of change call for which. This ADR deliberately leaves that
question open and records it under Non-goals.

## Decision

An FRD is a **living document**. It's the current answer to "what must this
feature do?", and it can't do that job if a reader must replay a chain of
amendments to learn the answer. This ADR therefore decides the following:

1. **Amendment FRDs are rejected.** The test for how a change lands isn't
   "new document or edit?" but "is this the same feature?" The FRD's slug
   already carries feature identity. For the same feature, edit the FRD in
   place. For a new feature, write a new FRD.
2. **Deltas between two versions of the same FRD are keyed on AC
   identifiers, not on lines.** To find what changed, compare the set of AC
   identifiers that the old version declared with the set that the new
   version declares. Each identifier falls into one of three cases:

   - **Only in the new version:** a criterion was added. This is new work
     to implement and test.
   - **In both, but its text changed:** the requirement changed while its
     identifier stayed the same. Every blueprint section and work order
     that cites that identifier was written against the old wording, so
     each one needs revisiting.
   - **Only in the old version:** a criterion was removed. Find whatever
     implemented it, because that code may now be unspecified.

This decision depends on AC identifiers being append-only: an FRD never
renumbers an existing acceptance criterion's identifier
(`plugin/skills/spec/references/ac-traceability-standards.md`). An
identifier-keyed delta is sound only because an identifier keeps meaning
the same criterion across every revision of the FRD that carries it.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **Amendment FRDs:** a new FRD that modifies an earlier one | Preserves each version as an unchanged record. A superseding chain is a familiar pattern for decision records. | A reader must replay the chain of amendments to learn the feature's current requirements. The FRD's one job is to be *the* current answer, and a chain prevents that. A growing chain gets harder to replay with every revision. | Defeats the FRD's one job: being the single current answer |
| **Line-based diffs** to find what changed between FRD versions | Needs no new convention, because `git diff` already does it. Works for any document, not just FRDs. | Fails in both directions. Reordering or rewording a requirement produces line noise that isn't a requirements change. Meanwhile, `SHALL` → `SHOULD` or `100ms` → `500ms` is a one-word diff with large downstream impact, and a line diff doesn't distinguish it from the noise. | A line diff can't tell a cosmetic edit from a requirements change, so it neither suppresses false positives nor surfaces the changes that matter |
| **Edit-in-place FRDs with AC-identifier-keyed deltas** (chosen) | One current document per feature. The same-feature test is mechanical (slug equality). A delta reduces to a query over identifiers that tests and work orders already cite, not a new document type or a new citation format. | Depends on AC identifiers staying append-only, a rule this ADR doesn't enforce. The CLI doesn't compute the identifier-keyed delta yet. | N/A — selected |

## Consequences

**Positive:**
- Tests already cite AC identifiers, and work orders already name their
  sources (`plugin/skills/spec/references/ac-traceability-standards.md`).
  A changed AC therefore traces forward to the tests that verify it and the
  work orders that produced it. Finding everything that a requirements
  change affects is a query over existing machinery, not a new document
  type or a new citation format to maintain.
- A feature has exactly one current FRD to read. No reader has to
  reconstruct intent by replaying a chain of amendments.
- For most changes, the same-feature test is mechanical rather than a
  judgment call: same slug, edit in place; different slug, new FRD.

**Negative:**
- The scheme is sound only while AC identifiers stay append-only. A single
  renumbered identifier, accidental or not, silently retargets every test
  citation and work-order source that pointed at it. It also makes the
  identifier-keyed delta for that FRD meaningless. This ADR relies on the
  append-only rule but doesn't enforce it.
- Editing an FRD in place leaves no separate document trail of *why* a
  requirement changed. That history lives only in git blame and commit
  messages on the FRD file, which is weaker provenance than a document per
  revision would give.
- No tooling computes the identifier-keyed delta yet. Until it exists,
  applying the delta means reading two FRD versions by hand. A CLI
  capability to automate it (`specs diff` or equivalent) is out of scope
  for this decision and would need its own FRD.
- This ADR doesn't resolve one boundary case: a change large enough that
  "is this the same feature?" no longer has an obvious answer. The slug
  test answers the common case, but a genuinely ambiguous case still needs
  human judgment.

## Non-goals

- **The ADR lifecycle.** This ADR deliberately doesn't decide how an ADR
  changes, and it doesn't imply that an ADR is immutable. At least two
  kinds of change amend an ADR without reversing its decision, so they
  don't call for a successor:
  - Recording an additional positive or negative consequence discovered
    after the decision, when the decision itself still holds.
  - Recording an additional option that was considered, when none of the
    new options changes which option was selected.

  Whether those changes are amendments in place, and where the line falls
  between an amendment and a supersession, needs its own ADR. Deferring
  that question keeps this decision focused on the question at hand: how a
  feature's requirements and acceptance criteria change.
- **A `specs diff` CLI capability.** Computing the identifier-keyed delta
  automatically is out of scope and would need its own FRD. For details,
  see Consequences.

## References

- `plugin/skills/spec/references/ac-traceability-standards.md` — states
  that AC identifiers are append-only and never renumbered, and names
  the same downstream failure (a meaningless identifier-keyed delta)
  that motivates this decision
- `plugin/skills/spec/references/frd-template.md` — FRD structure and AC
  identifier conventions

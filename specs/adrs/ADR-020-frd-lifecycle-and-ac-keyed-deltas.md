---
title: "ADR-020: FRDs Evolve in Place; Deltas Are Keyed on AC Identifiers"
description: An FRD is a living document edited in place under a same-feature test keyed on its slug, never amended by a successor FRD; changes between versions are diffed by AC identifier, not by line
slug: adr-020
type: adr
status: ready-for-review
updated_at: 2026-08-13T18:25:00Z
---

# ADR-020: FRDs Evolve in Place; Deltas Are Keyed on AC Identifiers

**Status:** ready-for-review
**Date:** 2026-08-13

## Context

The framework has never decided how an FRD changes after it is first
written. Two questions were open: whether a change to a feature's
requirements produces a new document or edits the existing one, and, for
whichever document ends up carrying the change, how a reader or a tool
locates what actually changed between versions. The CLI report/guard
work is the first change that needs an answer — it revises acceptance
criteria on an already-active FRD — and without a settled process every
subsequent revision re-decides the same two questions from scratch.

The repository already names its documents differently: ADRs are
numbered, FRDs are slug-named per feature. Nothing stated what that
implies for change, so nothing stopped a later revision from treating an
FRD as a record to append to rather than a statement to keep current.

**Scope.** This decision covers the FRD lifecycle only. How an ADR
changes — whether it is amended in place or superseded, and which kinds
of change call for which — is a separate question, deliberately left
open here and recorded under Non-Goals below.

## Decision

An FRD is a **living document**: the current answer to "what must this
feature do," and that job is destroyed if a reader must replay a chain
of amendments to learn the answer. Consequently:

1. **Amendment FRDs are rejected.** The test for how a change lands is
   not "new document or edit" but "is this the same feature," and
   feature identity is already carried by the FRD's slug. Same feature
   → edit the FRD in place. New feature → new FRD.
2. **Deltas between two versions of the same FRD are keyed on AC
   identifiers, not on lines.** To find what changed, compare the set
   of AC identifiers the old version declared against the set the new
   version declares. Every identifier falls into one of three cases:

   - **Only in the new version** — a criterion was added. This is new
     work to implement and to test.
   - **In both, but its text changed** — the requirement moved while
     its identifier stayed put. Every blueprint section and work order
     citing that identifier was written against the old wording, so
     each needs revisiting.
   - **Only in the old version** — a criterion was removed. Find
     whatever implemented it, because that code may now be
     unspecified.

This depends on AC identifiers being append-only — an FRD never
renumbers an existing acceptance criterion's identifier
(`plugin/skills/spec/references/ac-traceability-standards.md`). An
identifier-keyed delta is sound only because an identifier keeps
meaning the same criterion across every revision of the FRD that
carries it.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **Amendment FRDs** — a new FRD that modifies an earlier one | Preserves each version as an unchanged record; a superseding chain is a familiar pattern for decision records | A reader has to replay the chain of amendments to learn the feature's current requirements; the FRD's one job is being *the* current answer, and a chain destroys that; a growing chain gets harder to replay with every revision | Directly defeats the FRD's one job: being the single current answer |
| **Line-based diffs** for targeting what changed between FRD versions | Needs no new convention — `git diff` already does it; works for any document, not just FRDs | Fails in both directions: reordering or rewording a requirement produces line noise that is not a requirements change, while `SHALL` → `SHOULD` or `100ms` → `500ms` is a one-word diff with large downstream impact that a line diff does not distinguish from the noise | A line diff cannot tell a cosmetic edit from a requirements change, so it neither suppresses false positives nor surfaces the changes that matter |
| **Edit-in-place FRDs with AC-identifier-keyed deltas** (chosen) | One current document per feature; the same-feature test is mechanical (slug equality); a delta reduces to a query over identifiers already threaded through tests and work orders, not a new document type or a new citation format | Depends on AC identifiers staying append-only, a rule this ADR does not itself enforce; the identifier-keyed delta is not yet built as a CLI capability | N/A — selected |

## Consequences

**Positive:**
- Tests already cite AC identifiers and work orders already name their
  sources (`plugin/skills/spec/references/ac-traceability-standards.md`),
  so a changed AC traces forward to the tests verifying it and the work
  orders that produced it. Locating the blast radius of a requirements
  change is a query over machinery that already exists, not a new
  document type or a new citation format to maintain.
- A feature has exactly one current FRD to read. Nothing asks a reader
  to reconstruct intent by replaying a chain of amendments.
- The same-feature test is mechanical rather than a judgment call for
  most changes: same slug, edit in place; different slug, new FRD.

**Negative:**
- The whole scheme is only as sound as AC identifiers staying
  append-only. A single renumbered identifier — accidental or not —
  silently retargets every test citation and work-order source that
  pointed at it, and makes the identifier-keyed delta for that FRD
  meaningless. This ADR relies on that rule; it does not enforce it.
- Editing an FRD in place means there is no separate document trail of
  *why* a requirement changed. That history lives only in git blame and
  commit messages on the FRD file, which is weaker provenance than a
  document per revision would give.
- No tooling computes the identifier-keyed delta yet. Until it exists,
  applying it is a manual read of two FRD versions. A CLI capability to
  automate it (`spec diff` or equivalent) is out of scope for this
  decision and would need its own FRD.
- The boundary case this ADR does not resolve is a change large enough
  that "is this the same feature" stops being obvious. The slug test
  gives an answer for the common case; a genuinely ambiguous case still
  needs human judgment.

## Non-Goals

- **The ADR lifecycle.** How an ADR changes is deliberately not decided
  here, and this decision does not imply that an ADR is immutable. At
  least two kinds of change amend an ADR without reversing its decision,
  and so without calling for a successor:
  - recording an additional positive or negative consequence discovered
    after the decision was made, where the decision itself still holds;
  - recording an additional option that was considered, where none of
    the new options changes which option was selected.

  Whether those are amendments in place, and where the line sits between
  an amendment and a supersession, needs its own ADR. Deferring it is
  what keeps this decision to the question actually at hand — how a
  feature's requirements and acceptance criteria change.
- **A `spec diff` CLI capability.** Computing the identifier-keyed delta
  automatically is out of scope and would need its own FRD (see
  Consequences).

## References

- `plugin/skills/spec/references/ac-traceability-standards.md` — states
  that AC identifiers are append-only and never renumbered, and names
  the same downstream failure (a meaningless identifier-keyed delta)
  that motivates this decision
- `plugin/skills/spec/references/frd-template.md` — FRD structure and AC
  identifier conventions

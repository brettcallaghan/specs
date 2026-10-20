---
title: "ADR-020: FRDs Evolve in Place; Deltas Are Keyed on AC Identifiers"
description: An FRD is a living document edited in place under a same-feature test keyed on its slug, never amended by a successor FRD; changes between versions are diffed by AC identifier, not by line
slug: adr-020
type: adr
status: ready-for-review
---

# ADR-020: FRDs evolve in place; deltas are keyed on AC identifiers

**Status:** ready-for-review
**Date:** 2026-08-13

## Context

A change to an FRD after it is first written raises two questions. First, does
a change to a feature's requirements produce a new document or edit the
existing one? Second, for whichever document carries the change, how does a
reader or a tool locate what changed between versions? Without a settled
process, every revision re-decides the same two questions from scratch.

The repository names its documents differently: ADRs are numbered, and FRDs are
slug-named per feature. Without a stated rule for what that implies for change,
nothing stops a revision from treating an FRD as a record to append to rather
than a statement to keep current.

**Scope.** This decision covers the FRD lifecycle only. How an ADR changes
(whether it's amended in place or superseded, and which kinds of change call
for which) is a separate question. It's left open here on purpose and recorded
under Non-goals.

## Decision

An FRD is a **living document**: the current answer to "what must this feature
do?" If a reader must replay a chain of amendments to learn that answer, the
FRD no longer does its job. Consequently:

1. **Amendment FRDs are rejected.** The test for how a change lands is "is
   this the same feature?", not "new document or edit?" The FRD's slug already
   carries feature identity. Same feature → edit the FRD in place. New feature
   → new FRD.
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

This depends on AC identifiers being append-only: an FRD never renumbers an
existing acceptance criterion's identifier
(`plugin/skills/spec/references/ac-traceability-standards.md`). An
identifier-keyed delta is sound only because an identifier keeps meaning the
same criterion across every revision of the FRD that carries it.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **Amendment FRDs** — a new FRD that modifies an earlier one | Preserves each version as an unchanged record. A superseding chain is a familiar pattern for decision records | A reader has to replay the chain of amendments to learn the feature's current requirements. The FRD's one job is to be *the* current answer, and a chain destroys that. A growing chain gets harder to replay with every revision | Directly defeats the FRD's one job: being the single current answer |
| **Line-based diffs** for targeting what changed between FRD versions | Needs no new convention, because `git diff` already does it. Works for any document, not just FRDs | Fails in both directions. Reordering or rewording a requirement produces line noise that isn't a requirements change. Meanwhile, `SHALL` → `SHOULD` or `100ms` → `500ms` is a one-word diff with large downstream impact, and a line diff doesn't distinguish it from the noise | A line diff cannot tell a cosmetic edit from a requirements change, so it neither suppresses false positives nor surfaces the changes that matter |
| **Edit-in-place FRDs with AC-identifier-keyed deltas** (chosen) | One current document per feature. The same-feature test is mechanical (slug equality). A delta reduces to a query over identifiers already threaded through tests and work orders, not a new document type or a new citation format | Depends on AC identifiers staying append-only, a rule this ADR doesn't itself enforce. The identifier-keyed delta isn't yet built as a CLI capability | N/A — selected |

## Consequences

**Positive:**
- Tests already cite AC identifiers, and work orders already name their sources
  (`plugin/skills/spec/references/ac-traceability-standards.md`). So a changed
  AC traces forward to the tests that verify it and the work orders that
  produced it. Locating the blast radius of a requirements change is a query
  over machinery that already exists, not a new document type or a new
  citation format to maintain.
- A feature has exactly one current FRD to read. Nothing asks a reader
  to reconstruct intent by replaying a chain of amendments.
- The same-feature test is mechanical rather than a judgment call for
  most changes: same slug, edit in place; different slug, new FRD.

**Negative:**
- The whole scheme is only as sound as AC identifiers staying append-only. A
  single renumbered identifier, accidental or not, silently retargets every
  test citation and work-order source that pointed at it. It also makes the
  identifier-keyed delta for that FRD meaningless. This ADR relies on that
  rule but doesn't enforce it.
- Editing an FRD in place leaves no separate document trail of *why* a
  requirement changed. That history lives only in git blame and commit
  messages on the FRD file, which is weaker provenance than a document per
  revision would give.
- No tooling computes the identifier-keyed delta yet. Until it exists,
  applying it is a manual read of two FRD versions. A CLI capability to
  automate it (`specs diff` or equivalent) is out of scope for this
  decision and would need its own FRD.
- This ADR doesn't resolve one boundary case: a change large enough that "is
  this the same feature?" stops being obvious. The slug test answers the common
  case, but a genuinely ambiguous case still needs human judgment.

## Non-goals

- **The ADR lifecycle.** How an ADR changes is left undecided here on purpose,
  and this decision doesn't imply that an ADR is immutable. At least two kinds
  of change amend an ADR without reversing its decision, and so without calling
  for a successor:
  - Recording an additional positive or negative consequence discovered
    after the decision was made, where the decision itself still holds.
  - Recording an additional option that was considered, where none of
    the new options changes which option was selected.

  Whether those are amendments in place, and where the line sits between an
  amendment and a supersession, needs its own ADR. Deferring it keeps this
  decision to the question at hand: how a feature's requirements and acceptance
  criteria change.
- **A `specs diff` CLI capability.** Computing the identifier-keyed delta
  automatically is out of scope and would need its own FRD (see
  Consequences).

## References

- `plugin/skills/spec/references/ac-traceability-standards.md` — states
  that AC identifiers are append-only and never renumbered, and names
  the same downstream failure (a meaningless identifier-keyed delta)
  that motivates this decision
- `plugin/skills/spec/references/frd-template.md` — FRD structure and AC
  identifier conventions

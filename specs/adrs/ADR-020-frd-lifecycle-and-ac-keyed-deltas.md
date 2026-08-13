---
title: "ADR-020: FRDs Evolve in Place; Deltas Are Keyed on AC Identifiers"
description: An FRD is a living document edited in place under a same-feature test keyed on its slug, never amended by a successor FRD; changes between versions are diffed by AC identifier, not by line
slug: adr-020
type: adr
status: ready-for-review
updated_at: 2026-08-13T06:59:14Z
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

The repository already treats ADRs and FRDs differently in one respect:
ADRs are numbered records that get superseded, FRDs are slug-named per
feature. That split was never stated as a principle, so nothing stopped
a later change from treating an FRD the way ADRs are treated — as a
record to append to rather than a statement to keep current.

## Decision

An FRD is a **living document**; an ADR is an **immutable record**. An
ADR is a decision at a moment, so it is superseded rather than edited.
An FRD is the current answer to "what must this feature do," and that
job is destroyed if a reader must replay a chain of amendments to learn
the answer. Consequently:

1. **Amendment FRDs are rejected.** The test for how a change lands is
   not "new document or edit" but "is this the same feature," and
   feature identity is already carried by the FRD's slug. Same feature
   → edit the FRD in place. New feature → new FRD.
2. **Deltas between two versions of the same FRD are keyed on AC
   identifiers, not on lines.** `in new, not in old` is new work; `in
   both, text changed` means the blueprint sections and work orders
   citing that identifier need revisiting; `in old, not in new` means
   finding what implemented it, because that code may now be
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
| **Amendment FRDs** — a new FRD that modifies an earlier one, mirroring how ADRs supersede | Consistent with the ADR pattern already in the repository; preserves each version as an immutable record | A reader has to replay the chain of amendments to learn the feature's current requirements; the FRD's one job is being *the* current answer, and a chain destroys that; a growing chain gets harder to replay with every revision | Directly defeats the reason an FRD exists as distinct from an ADR |
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
  *why* a requirement changed the way an ADR's Context and Alternatives
  Considered would record for a decision. That history lives only in
  git blame and commit messages on the FRD file, which is weaker
  provenance than a superseded ADR gives its own decisions.
- No tooling computes the identifier-keyed delta yet. Until it exists,
  applying it is a manual read of two FRD versions. A CLI capability to
  automate it (`spec diff` or equivalent) is out of scope for this
  decision and would need its own FRD.
- The boundary case this ADR does not resolve is a change large enough
  that "is this the same feature" stops being obvious. The slug test
  gives an answer for the common case; a genuinely ambiguous case still
  needs human judgment.

## References

- `plugin/skills/spec/references/ac-traceability-standards.md` — states
  that AC identifiers are append-only and never renumbered, and names
  the same downstream failure (a meaningless identifier-keyed delta)
  that motivates this decision
- `plugin/skills/spec/references/frd-template.md` — FRD structure and AC
  identifier conventions

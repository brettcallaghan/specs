<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/guides/anti-patterns-guide.md (version 0.1)
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: Spec-First Anti-Patterns
description: The failure modes of spec-first authoring and what each one
  costs — wrong document type, untestable prose, duplicated content,
  unverified checks, and deltas that never get folded back in
slug: guide-spec-anti-patterns
type: guide
status: active
version: "0.1"
---

# Spec-First Anti-Patterns

Read this before setting any document to `ready-for-review`, and when
reviewing someone else's. Each row names the pattern and the problem it
causes, so a reviewer can cite the row rather than argue from taste.

| Pattern | Problem |
|---------|---------|
| Writing code before the spec exists | Requirements get reverse-engineered from the implementation and disputes surface as rework |
| Writing an ADR before requirements context is known | Answers the wrong question |
| Putting policy or invariants in an ADR | ADRs are for decisions with trade-offs — use a standards document |
| Putting trade-off analysis in a standards document | Standards say WHAT must hold, not WHY an option won — use an ADR |
| Nesting sub-features inside one FRD | Each sub-feature is its own FRD |
| Free-form prose requirements ("should handle errors appropriately") | Untestable; every normative statement uses an EARS pattern |
| Duplicating content across documents | Copies drift; cross-reference with `[[document-slug]]` instead |
| Leaving template comment blocks in a finished document | Instruction bloat; delete them after completing each section |
| Reusing a retired REQ/AC identifier | Breaks traceability; never reuse IDs |
| Reporting a check as passed without running the command that decides it | Unverifiable: an agent that ran the check and one that assumed it produce identical output. Run `specs check` and cite it |
| Deriving the next ADR number or a slug's uniqueness by reading filenames | The slug is independent of the filename, so a scan misses the collision it exists to find. `specs next-id` and `specs slug-exists` read the index |
| Retiring or renaming a document without asking who references it | Every citation silently stops resolving, and nobody finds out until a reader follows one. `specs refs <slug>` names them at their own lines, before the change |
| Inventing a new document type | The taxonomy is deliberately small; flag the gap instead |
| Filing a private schema as a standalone Contract | Freezes a container's internal implementation into a de facto public interface |
| Deep component internals in a container blueprint | Bloat; extract to a component or module blueprint when it warrants standalone docs |
| Filing a shared/published or composite unit as a component | It crosses a boundary of independent change — author a module blueprint (versioning + change rules) |
| Per-container diagrams redrawing shared deployment topology | N copies drift; the system blueprint's deployment diagrams are authoritative |
| Modeling ambient platforms (logging, mesh) as diagram elements | Dilutes every view; use prose plus a cross-cutting standards document |
| Authoring a Contract for a third-party (SaaS) interface | The vendor's published contract governs; reference it by URL instead |
| Recording durable content only in a feature blueprint | The document retires on merge; decisions belong in ADRs, interface shapes in contract documents |
| Task lists or new code structure in a feature blueprint | Wrong altitude: planners derive tasks and implementers choose signatures; the blueprint carries deltas, constraints, and anchors |
| Merging a feature without folding its deltas into the owning blueprints | Architecture blueprints rot and the reference chain collapses; the Definition of Done includes the documentation updates |
| Beginning implementation planning against a draft feature blueprint | Planning SHALL NOT begin until status is approved and Open Questions is empty — gaps bounce back to the human loop, never get guessed at |
| Hand-authoring a work order instead of grooming one | Work orders are generated to the Definition of Ready by grooming; writing one by hand skips the readiness gate that makes it dispatchable |

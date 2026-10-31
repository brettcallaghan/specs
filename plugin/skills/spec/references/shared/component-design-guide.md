<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/guides/component-design-guide.md
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: Component Design Guide
description: The design principles a blueprint author or auditor applies to
  an architecture element, each as the question to ask of it
slug: guide-component-design
type: guide
status: active
---

# Component design guide

Ask these questions when you design an element or audit its blueprint. Each
row names a principle and the question that applies it. The answers feed a
blueprint's Responsibility, Interface, and Design sections.
[[blueprint-content-standards]] sets what those sections may say.

| Principle | Question |
|-----------|----------|
| Information hiding (Parnas, 1972) | What design decision likely to change does this element hide? |
| Comments state the non-obvious (Ousterhout, *A Philosophy of Software Design*) | Does each sentence say something the code can't, at a higher level than the code? |
| Deep modules (Ousterhout) | Is the interface much simpler than what it hides? |
| Common Closure (Martin, *Clean Architecture*) | Do the files that change together for one reason sit in one element? |
| Acyclic and Stable Dependencies (Martin) | Are there cycles, and does each element depend on something more stable than itself? |
| Single Responsibility (SOLID) | Does it have exactly one reason to change? |
| Open-Closed (SOLID) | Can a new case be added without editing it? |
| Interface Segregation (SOLID) | Does each consumer depend only on the members it uses? |
| Dependency Inversion (SOLID) | Does it require abstractions it owns, passed in through its interface, rather than concrete siblings or globals such as the clock? |
| Functional core, imperative shell (Bernhardt) | Are decisions made in pure code, with I/O confined to a thin outer layer? |
| Side-effect-free queries (from Meyer's Command-Query Separation) | Does every operation that only reads leave state unchanged? |
| Design by Contract (Meyer) | Are the preconditions in Requires, the postconditions in Provides, and the rest in Invariants? |
| Ports and Adapters (Cockburn) | For a container: which components are the core, and which adapt it to I/O? |
| Architecturally evident code (Fairbanks; Brown) | Can someone find this element from a directory listing? |

## References

- [[blueprint-content-standards]] — what a blueprint states, and the audit
- [[c4-standards]] — the C4 levels an element maps to

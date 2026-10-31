<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/guides/design-principles-components-guide.md
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: "Design Principles: Components"
description: The principles and patterns for composing code inside a component
  or module, each as the question to ask of its blueprint
slug: guide-design-principles-components
type: guide
status: active
---

# Design principles: components

Ask these questions when you design how the code of a component, or the code a
module holds directly, is composed. Ask them again when you audit that element's
blueprint. C4's code level has no blueprint of its own, so its patterns live
here. A Container, Module, or Component Blueprint may name entries from this
list that all of the element's code follows.

Use these entries with the universal guide, [[guide-design-principles]], which
also explains how to choose patterns and notes that these lists are incomplete.
Record each pattern the design adopts under
[[blueprint-content-standards#Name each pattern with its progenitor, constraint, and need]].

| Entry | Progenitors | Questions it answers |
|-------|-------------|----------------------|
| Liskov Substitution | Liskov (1987); Liskov and Wing (1994) | Can every implementation stand in for the abstraction with no surprises? |
| Program to an Interface | Gamma et al., *Design Patterns* (1994) | Does code depend on what a collaborator does, not on its concrete type? |
| Composition over Inheritance | Gamma et al. (1994) | Is behavior built from small parts rather than a class hierarchy? |
| Command-Query Separation | Meyer (1988) | Does every operation that only reads leave state unchanged? |
| Make Illegal States Unrepresentable | Minsky, "Effective ML" (2010) | Can the types construct an invalid combination? |
| Immutability by Default | Bloch, *Effective Java* (2001); Hickey, "Are We There Yet?" (2009) | Which values can change after creation, and why? |
| Strategy | Gamma et al. (1994) | How is an algorithm or policy swapped without editing its caller? |
| Decorator | Gamma et al. (1994) | How is behavior added around an implementation without changing it? |
| Adapter | Gamma et al. (1994) | How does code use an interface that doesn't match the one it expects? |
| Factory | Gamma et al. (1994) | Which code decides the concrete type to create, so callers don't? |
| Observer | Gamma et al. (1994) | How are dependents told of a change without the subject knowing them? |
| Command | Gamma et al. (1994) | Should a request be data that can be queued, logged, retried, or undone? |
| State and Typestate | Gamma et al. (1994); Strom and Yemini (1986) | How does behavior change with state, and can illegal transitions be ruled out by type? |
| Entity, Value Object, Aggregate | Evans, *Domain-Driven Design* (2003) | Which types have identity, which are values, and which object guards each invariant? |

## References

- [[guide-design-principles]] — the universal guide, and how to choose patterns
- [[blueprint-content-standards]] — how a blueprint names patterns, and the
  blueprint audit

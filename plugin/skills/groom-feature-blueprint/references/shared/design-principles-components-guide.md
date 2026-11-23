<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/guides/design-principles-components-guide.md
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: "Design Principles: Components"
description: The principles and patterns for composing code inside a
  component, each as the question to ask of its design
slug: guide-design-principles-components
type: guide
status: active
---

# Design principles: components

Ask these questions when you design how the code of a component is composed,
or review that design. The patterns of C4's code level live here. An element
at any level may adopt entries from this list that all of its code follows.

Use these entries with the universal guide, [[guide-design-principles]],
which explains
how to choose and name patterns and notes that these lists are incomplete.

| Entry | Progenitor | Questions it answers |
|-------|-------------|----------------------|
| Liskov Substitution | Liskov and Wing, "A Behavioral Notion of Subtyping" (1994) | Can every implementation stand in for the abstraction with no surprises? |
| Program to an Interface | Gamma et al., *Design Patterns* (1994), "Program to an interface, not an implementation" | Does code depend on what a collaborator does, not on its concrete type? |
| Composition over Inheritance | Gamma et al., *Design Patterns* (1994), "Favor object composition over class inheritance" | Is behavior built from small parts rather than a class hierarchy? |
| Command-Query Separation | Meyer, *Object-Oriented Software Construction* (1988), Command-Query Separation | Does every operation that only reads leave state unchanged? |
| Make Illegal States Unrepresentable | Minsky, "Effective ML" (2010) | Can the types construct an invalid combination? |
| Immutability by Default | Bloch, *Effective Java* (2001), "Favor immutability" | Which values can change after creation, and why? |
| Strategy | Gamma et al., *Design Patterns* (1994), Strategy | How is an algorithm or policy swapped without editing its caller? |
| Decorator | Gamma et al., *Design Patterns* (1994), Decorator | How is behavior added around an implementation without changing it? |
| Adapter | Gamma et al., *Design Patterns* (1994), Adapter | How does code use an interface that doesn't match the one it expects? |
| Factory | Gamma et al., *Design Patterns* (1994), Factory Method and Abstract Factory | Which code decides the concrete type to create, so callers don't? |
| Observer | Gamma et al., *Design Patterns* (1994), Observer | How are dependents told of a change without the subject knowing them? |
| Command | Gamma et al., *Design Patterns* (1994), Command | Should a request be data that can be queued, logged, retried, or undone? |
| State | Gamma et al., *Design Patterns* (1994), State | How does behavior change with state? |
| Typestate | Strom and Yemini, "Typestate" (1986) | Can illegal state transitions be ruled out by type? |
| Entity, Value Object, Aggregate | Evans, *Domain-Driven Design* (2003), Entities, Value Objects, and Aggregates | Which types have identity, which are values, and which object guards each invariant? |

## References

- [[guide-design-principles]] — the universal principles, and how to choose
  and name patterns

---
title: "Design Principles: Containers"
description: The principles and patterns for composing components into a
  container, each as the question to ask of its design
slug: guide-design-principles-containers
type: guide
status: active
---

# Design principles: containers

Ask these questions when you design how the components of a container
compose, or of any unit built from components. Ask them again when you review
that design. Use them with the universal guide, [[guide-design-principles]],
which explains how to choose and name patterns and notes that these lists
are incomplete.

| Entry | Progenitor | Questions it answers |
|-------|-------------|----------------------|
| Single Responsibility | Martin, *Clean Architecture* (2017), "SRP: The Single Responsibility Principle" | Does each component have one reason to change? |
| Common Closure and Common Reuse | Martin, *Agile Software Development* (2002), the Common Closure and Common Reuse principles | Do the files that change together for one reason sit in one component? Does each consumer of a component use all of it? |
| Interface Segregation | Martin, "The Interface Segregation Principle", *C++ Report* (1996) | Does each consumer depend only on the members it uses? |
| Functional Core, Imperative Shell | Bernhardt, "Boundaries" (2012) | Are decisions made in pure code, with I/O confined to a thin outer layer? |
| Ports and Adapters | Cockburn, "Hexagonal Architecture" (2005) | Which components are the core, and which adapt it to I/O? |
| Plugin, Microkernel | Buschmann et al., *POSA Vol. 1* (1996), Microkernel | How do new features join a small core without changing it? Does the core call plugins at named extension points, with no plugin driving the core? |
| Pipes and Filters | Buschmann et al., *POSA Vol. 1* (1996), Pipes and Filters | Can processing be a sequence of independent stages that can be reordered or replaced? |
| Chain of Responsibility, Middleware | Gamma et al., *Design Patterns* (1994), Chain of Responsibility | How are cross-cutting request concerns added without touching each handler? |
| Domain Events | Vernon, *Implementing Domain-Driven Design* (2013), "Domain Events" | How do components react to a change without calling each other? |
| Layers | Buschmann et al., *POSA Vol. 1* (1996), Layers | Which way do dependencies point between layers? |
| Clean Architecture | Martin, *Clean Architecture* (2017), the Dependency Rule | Do source dependencies point only inward, toward policy? |
| Service Layer | Fowler, *Patterns of Enterprise Application Architecture* (2002), Service Layer | Where does an application operation start and end? |
| Repository | Evans, *Domain-Driven Design* (2003), Repositories | How does domain logic reach storage without knowing how it's stored? |
| Dependency Injection, Composition Root | Seemann, *Dependency Injection in .NET* (2011), Composition Root | Where are components wired together, and can each be replaced in a test? |

## References

- [[guide-design-principles]] — the universal principles, and how to choose
  and name patterns
- [[guide-design-principles-components]] — patterns for the code inside a
  component

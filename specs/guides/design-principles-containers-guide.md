---
title: "Design Principles: Containers"
description: The principles and patterns for composing components into a
  container or module, each as the question to ask of its blueprint
slug: guide-design-principles-containers
type: guide
status: active
---

# Design principles: containers

Ask these questions when you design how the components of a container, or the
components and submodules of a module, compose. Ask them again when you audit
that element's blueprint. Use them with the universal guide,
[[guide-design-principles]], which also explains how to choose patterns and
notes that these lists are incomplete. Record each pattern the design adopts
under
[[blueprint-content-standards#Name each pattern with its progenitor, constraint, and need]].

| Entry | Progenitors | Questions it answers |
|-------|-------------|----------------------|
| Single Responsibility | Martin (2002, refined 2017) | Does each component have one reason to change? |
| Common Closure and Common Reuse | Martin, *C++ Report* (1996); *Agile Software Development* (2002) | Do the files that change together for one reason sit in one component? Does each consumer of a component use all of it? |
| Interface Segregation | Martin, *C++ Report* (1996) | Does each consumer depend only on the members it uses? |
| Functional Core, Imperative Shell | Bernhardt, "Boundaries" (2012) | Are decisions made in pure code, with I/O confined to a thin outer layer? |
| Ports and Adapters | Cockburn (2005) | Which components are the core, and which adapt it to I/O? |
| Plugin, Microkernel | Buschmann et al., *POSA Vol. 1* (1996) | How do new features join a small core without changing it? Does the core call plugins at named extension points, with no plugin driving the core? |
| Pipes and Filters | McIlroy, Unix pipes (1973); Buschmann et al. (1996) | Can processing be a sequence of independent stages that can be reordered or replaced? |
| Chain of Responsibility, Middleware | Gamma et al., *Design Patterns* (1994) | How are cross-cutting request concerns added without touching each handler? |
| Domain Events | Fowler (2005); Vernon, *Implementing Domain-Driven Design* (2013) | How do components react to a change without calling each other? |
| Layers, Onion, Clean Architecture | Buschmann et al. (1996); Palermo (2008); Martin (2012) | Which way do dependencies point between layers? |
| Service Layer, Use Case | Fowler, *PoEAA* (2002); Jacobson (1992) | Where does an application operation start and end? |
| Repository | Fowler (2002); Evans (2003) | How does domain logic reach storage without knowing how it's stored? |
| Dependency Injection, Composition Root | Fowler (2004); Seemann (2011) | Where are components wired together, and can each be replaced in a test? |

## References

- [[guide-design-principles]] — the universal guide, and how to choose patterns
- [[guide-design-principles-components]] — patterns for the code inside a
  component
- [[blueprint-content-standards]] — how a blueprint names patterns, and the
  blueprint audit

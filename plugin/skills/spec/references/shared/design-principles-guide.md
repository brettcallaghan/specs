<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/guides/design-principles-guide.md
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: Design Principles Guide
description: The principles that apply at every C4 level, each as the question
  to ask of an element, plus how to find the patterns for an element's level,
  application type, and language
slug: guide-design-principles
type: guide
status: active
---

# Design principles guide

Ask these questions when you design an element or audit its blueprint. This
universal guide holds the principles for every C4 level. Use it with the level
guide for the element's level. The answers feed a blueprint's Design section
and, in a Component or Module Blueprint, its Responsibility and Interface
sections. [[blueprint-content-standards]] sets what those sections may say.

The lists are a starting point, not a catalog. They hold the principles and
patterns that apply most widely in current architectures, with priority given to
those that keep a design modular and extensible. When another pattern fits the
problem better, use it, and name it with its own progenitor
([[blueprint-content-standards#Name each pattern with its progenitor, constraint, and need]]).

## Choose the level guide

| Element | Guides |
|---------|-------|
| System | Universal + [[guide-design-principles-systems]] |
| Container | Universal + [[guide-design-principles-containers]] |
| Module | Universal, [[guide-design-principles-containers]] for how its components and submodules compose, and [[guide-design-principles-components]] for the code it holds directly |
| Component | Universal + [[guide-design-principles-components]] |
| Feature Blueprint delta | Universal + the guide for the level of the element the delta changes |

A Container, Module, or Component Blueprint may also name entries from
[[guide-design-principles-components]] that all of the element's code
follows, such as Make Illegal States Unrepresentable.

Adopt a pattern only when its question has a real answer for the element today.
A question with no answer means the pattern doesn't belong. Record each pattern
you adopt as
[[blueprint-content-standards#Name each pattern with its progenitor, constraint, and need]]
requires: its name, its progenitor, the constraint it implies, and the need it
serves.

## Universal principles

| Entry | Progenitors | Questions it answers |
|-------|-------------|----------------------|
| Information Hiding | Parnas, "On the Criteria To Be Used in Decomposing Systems into Modules" (1972) | What design decision likely to change does this element hide? |
| Open-Closed | Meyer, *Object-Oriented Software Construction* (1988); Martin (1996) | Can a new case be added without editing what exists? Where are the extension points? |
| Dependency Inversion | Martin, *C++ Report* (1996) | Does each element depend on abstractions that the policy side owns, rather than on concrete details? |
| Separation of Concerns | Dijkstra, "On the role of scientific thought" (1974) | Does each part deal with one concern that can be reasoned about alone? |
| High Cohesion, Low Coupling | Stevens, Myers, and Constantine (1974); Constantine and Yourdon, *Structured Design* (1979) | Do things that change together sit together, and do the rest know little about each other? |
| Deep Modules | Ousterhout, *A Philosophy of Software Design* (2018) | Is the interface much simpler than what it hides? |
| Acyclic and Stable Dependencies | Martin, *Agile Software Development* (2002) | Are there dependency cycles? Does each element depend on something more stable than itself? |
| Design by Contract | Meyer (1986) | What are the preconditions, postconditions, and invariants, and which section of the blueprint states each? |
| Single Source of Truth | Codd, normalization (1970); Hunt and Thomas, DRY (1999) | Which element is the one authority for each piece of data or knowledge? |
| Parse, Don't Validate | King (2019) | Where is untrusted input parsed into trusted types? Is each rule a type can't express checked once, at that boundary or by the domain object that owns it, rather than checked again downstream? |
| Idempotence | HTTP/1.1, RFC 2068 (1997), now RFC 9110 | Can this operation be repeated or retried safely? |
| Least Privilege | Saltzer and Schroeder (1975) | Does the element get only the access it needs? |
| Fail Fast | Shore, *IEEE Software* (2004) | Is bad state detected at the earliest point, and loudly? |
| Gall's Law and YAGNI | Gall, *Systemantics* (1975); Beck, *Extreme Programming Explained* (1999); Jeffries | Is this built for a need that exists now? Did it grow from a simpler version that worked? |
| Architecturally Evident Code | Fairbanks, *Just Enough Software Architecture* (2010); Brown, *Software Architecture for Developers* (2012) | Can someone find this element from a directory listing? |
| Strangler Fig | Fowler (2004) | How is this element replaced in steps, with old and new running side by side? |

## Same idea, different level

Some patterns take a different name at each level. Name the one for the
element's level, so a system-level decision isn't recorded as its code-level
version.

| Idea | Components | Containers | Systems |
|------|------------|------------|---------|
| Translate an interface | Adapter | Ports and Adapters | Anti-Corruption Layer |
| Simplify the entry point | Facade | Service Layer | API Gateway, Backend for Frontend |
| Announce a change | Observer | Domain Events | Event-Driven Architecture |
| Intercept requests | Decorator | Middleware | Sidecar |
| Request as data | Command | Command handler | Message on a queue |
| Split reads from writes | Command-Query Separation | Separate command and query handlers | CQRS |
| Contain failure | Result, Fail Fast | Supervisor | Circuit Breaker, Bulkhead |

## Patterns by application type and language

The lists stay language-neutral. The examples below show the kind of pattern
that matters for a particular kind of application or language. They're
prompts, not a list to adopt.

By application type:

- **Frontend:** unidirectional data flow (Flux, Redux), the
  Model-View-Controller family (MVC, MVP, MVVM), signals and reactive streams,
  component composition and hooks.
- **Backend service:** Ports and Adapters, Repository, Transactional Outbox,
  CQRS.
- **Data pipeline:** Pipes and Filters, Idempotent Consumer, checkpointing.
- **CLI and developer tool:** Command, Plugin, configuration layering.

By language:

- **Rust:** typestate, newtype, RAII and ownership, `Result` with the `?`
  operator.
- **Go:** small consumer-defined interfaces, functional options, errors as
  values.
- **TypeScript:** discriminated unions, branded types.
- **Python:** protocols, context managers.
- **Java and C#:** dependency injection containers, records.
- **F# and other ML-family languages:** Result and Railway-Oriented Programming.

Identify the element's application type and language. Find the patterns that fit
them, whether this guide lists them or not, and record each one the design
adopts in its Design section, as
[[blueprint-content-standards#Name each pattern with its progenitor, constraint, and need]]
requires.

The guides leave out three kinds of pattern. Some are tied to one language or
runtime, such as Pimpl, Visitor where the language has pattern matching, and
typeclass-based patterns. Some are provided by a modern language or framework,
such as Iterator. Others are performance patterns, such as Flyweight and Object
Pool. An author can still adopt any of them under
[[blueprint-content-standards#Name each pattern with its progenitor, constraint, and need]].

## References

- [[guide-design-principles-systems]] — patterns for composing containers
- [[guide-design-principles-containers]] — patterns for composing components
- [[guide-design-principles-components]] — patterns for composing code
- [[blueprint-content-standards]] — what a blueprint states, how it names
  patterns, and the audit
- [[adr-034]] — design principles come from curated guides by C4 level
- [[c4-standards]] — the C4 levels an element maps to

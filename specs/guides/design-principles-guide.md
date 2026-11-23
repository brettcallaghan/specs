---
title: Design Principles Guide
description: The design principles that apply at every C4 level, each as the
  question to ask of an element, plus how to find the patterns for an
  element's level, application type, and language
slug: guide-design-principles
type: guide
status: active
---

# Design principles guide

Ask these questions when you design a software element or review its design.
This universal guide holds the principles for every level of the C4 model. Use
it with the guide for the element's level.

When you record a pattern a design adopts, name it with its progenitor: the one
source, and the topic within it, that fixes its meaning. "Decorator" is also
Python syntax, and "Repository" is often confused with a data access object;
"Decorator (Gamma et al., *Design Patterns*, Decorator)" is not. Each entry
below names one progenitor. Where an idea has two origins with different
meanings, it has one entry for each context, so pick the entry whose context
matches the design.

The lists are a starting point, not a catalog. They hold the principles and
patterns that apply most widely in current architectures, with priority given to
those that keep a design modular and extensible. When another pattern fits the
problem better, use it.

## Choose the level guide

Use this guide with one level guide:

- **A system**, deciding how its containers integrate:
  [[guide-design-principles-systems]].
- **A container**, deciding how its components compose:
  [[guide-design-principles-containers]].
- **A component**, deciding how its code is composed:
  [[guide-design-principles-components]].

An element at any level may also adopt components-guide entries that all of
its code follows, such as Make Illegal States Unrepresentable.

Adopt a pattern only when its question has a real answer for the element today.
A question with no answer means the pattern doesn't belong.

## Universal principles

| Entry | Progenitor | Questions it answers |
|-------|-------------|----------------------|
| Information Hiding | Parnas, "On the Criteria To Be Used in Decomposing Systems into Modules" (1972) | What design decision likely to change does this element hide? |
| Open-Closed | Meyer, *Object-Oriented Software Construction* (1988), the Open-Closed principle | Can a new case be added without editing what exists? Where are the extension points? |
| Dependency Inversion | Martin, "The Dependency Inversion Principle", *C++ Report* (1996) | Does each element depend on abstractions that the policy side owns, rather than on concrete details? |
| Separation of Concerns | Dijkstra, "On the role of scientific thought" (1974) | Does each part deal with one concern that can be reasoned about alone? |
| High Cohesion, Low Coupling | Stevens, Myers, and Constantine, "Structured Design", *IBM Systems Journal* (1974) | Do things that change together sit together, and do the rest know little about each other? |
| Deep Modules | Ousterhout, *A Philosophy of Software Design* (2018), "Modules Should Be Deep" | Is the interface much simpler than what it hides? |
| Acyclic and Stable Dependencies | Martin, *Agile Software Development* (2002), the Acyclic Dependencies and Stable Dependencies principles | Are there dependency cycles? Does each element depend on something more stable than itself? |
| Design by Contract | Meyer, "Applying 'Design by Contract'", *IEEE Computer* (1992) | What are the preconditions, postconditions, and invariants, and where is each stated? |
| Single Source of Truth, for knowledge in code | Hunt and Thomas, *The Pragmatic Programmer* (1999), "The Evils of Duplication" | Which element is the one authority for each piece of knowledge, such as a rule, a format, or a list of cases? |
| Single Source of Truth, for stored data | Codd, "A Relational Model of Data for Large Shared Data Banks" (1970), normalization | Is each stored fact recorded once, with every other copy derived from it? |
| Parse, Don't Validate | King, "Parse, Don't Validate" (2019) | Where is untrusted input parsed into trusted types? Is each rule a type can't express checked once, at that boundary or by the domain object that owns it, rather than checked again downstream? |
| Idempotence | RFC 9110 (2022), "Idempotent Methods" | Can this operation be repeated or retried safely? |
| Least Privilege | Saltzer and Schroeder, "The Protection of Information in Computer Systems" (1975) | Does the element get only the access it needs? |
| Fail Fast | Shore, "Fail Fast", *IEEE Software* (2004) | Is bad state detected at the earliest point, and loudly? |
| Gall's Law | Gall, *Systemantics* (1975) | Did this grow from a simpler version that worked? |
| YAGNI | Beck, *Extreme Programming Explained* (1999) | Is this built for a need that exists now? |
| Architecturally Evident Code | Fairbanks, *Just Enough Software Architecture* (2010), the architecturally-evident coding style | Can someone find this element from a directory listing? |
| Strangler Fig | Fowler, "StranglerFigApplication" (2004) | How is this element replaced in steps, with old and new running side by side? |

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

Identify the element's application type and language, and find the patterns
that fit them, whether this guide lists them or not.

The guides leave out three kinds of pattern. Some are tied to one language or
runtime, such as Pimpl, Visitor where the language has pattern matching, and
typeclass-based patterns. Some are provided by a modern language or framework,
such as Iterator. Others are performance patterns, such as Flyweight and Object
Pool. An author can still adopt any of them.

## References

- [[guide-design-principles-systems]] — patterns for composing containers
- [[guide-design-principles-containers]] — patterns for composing components
- [[guide-design-principles-components]] — patterns for composing code
- [[guide-c4]] — the C4 levels
- [The C4 model](https://c4model.com/) — the levels these guides follow

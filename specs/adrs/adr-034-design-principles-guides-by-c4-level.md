---
title: "ADR-034: Design Principles Come from Curated Guides by C4 Level"
description: Architecture blueprints draw principles and patterns from a
  universal guide plus one guide per C4 level, and name each pattern they
  adopt with its progenitor
slug: adr-034
type: adr
status: approved
---

# ADR-034: Design principles come from curated guides by C4 level

**Status:** approved  
**Date:** 2026-10-31

## Context

An architecture blueprint's Design section records the patterns and
principles an element follows. Agents read those names and apply what they
know about each pattern: its roles, its consequences, and the ways it fails.
A name that points at the wrong concept, or at no concept in particular,
leads the agent to apply the wrong pattern's roles and constraints.

The blueprint at each C4 level makes different decisions. A System Blueprint
decides how containers integrate and stay consistent. A Container Blueprint
decides how components are arranged and which way dependencies point. A
Component Blueprint decides how code is structured. One flat list of principles
leaves the author to work out which level each entry belongs to. Agents then
name a pattern at the wrong level, such as an in-process Observer where the
design meant a message broker.

Pattern catalogs are long, and most of their entries are language-specific,
provided by a modern language or framework, or no longer current practice. An
agent can't load a whole catalog into its context and still have room for the
task.

A pattern name also drifts without an anchor. "Decorator" is Python syntax as
well as a design pattern, and "Repository" is often confused with a data
access object. The progenitor, the source that defined the pattern, fixes
which meaning the blueprint intends.

## Decision

We will have architecture blueprints draw principles and patterns from curated
guides organized by C4 level. A universal guide covers the principles that hold
at every level, and one guide each covers Systems, Containers, and Components.
A blueprint names each pattern it adopts at its own level. It gives three
things: the pattern's progenitor (which fixes its meaning), the constraint the
pattern implies, and the present need it serves. [[blueprint-content-standards]] rule 14 applies this
decision, and [[guide-design-principles]] is the universal guide.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| One flat list of principles | Short<br>One file to load | Doesn't say which level a principle belongs to | Agents name patterns at the wrong level |
| An exhaustive pattern catalog | Complete | Too long to load into an agent's context<br>Most entries are language-specific or dated | Spends the agent's context on patterns that don't apply |
| One curated guide holding every level's list | One file to load | Every template loads all four lists, two of which don't apply to most elements | Costs context on every architecture blueprint for lists it can't use |
| A universal guide plus one curated guide per level | Each template loads only the universal guide and its own level's guide<br>Lists stay short enough to apply rather than skim | Four files to keep current<br>A Module Blueprint loads two level guides | N/A — selected |

## Consequences

**Positive:**
- A blueprint names patterns in vocabulary that agents and reviewers share,
  anchored to a progenitor.
- An agent authoring a blueprint loads only the universal guide and the guides
  for the element's level.
- Agents that implement a change get the pattern names through the Work Order
  and never load the guides.

**Negative:**
- The lists are curated, so someone has to keep them current and decide what
  to add or cut.
- The lists are incomplete by design. An author still has to consider patterns
  the guides don't list.
- Naming a pattern doesn't make an agent follow it. Rule 14 pairs each name
  with a constraint written as an enforced invariant.

## References

- [[blueprint-content-standards]] — rule 14 applies this decision
- [[guide-design-principles]] — the universal guide
- [[guide-design-principles-systems]] — patterns for composing containers
- [[guide-design-principles-containers]] — patterns for composing components
- [[guide-design-principles-components]] — patterns for composing code
- [[adr-032]] — blueprints state rules, not code facts
- [[c4-standards]] — the C4 levels these guides follow

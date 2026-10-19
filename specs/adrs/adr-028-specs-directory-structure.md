---
title: "ADR-028: Default Directory Structure for the Spec Corpus"
description: The spec root mirrors the C4 containment hierarchy only where the containment edge has no child-side declaration; everything else is a flat top-level directory
slug: adr-028
type: adr
status: active
---

# ADR-028: Default Directory Structure for the Spec Corpus

**Status:** active  
**Date:** 2026-08-29

## Context

The framework didn't prescribe a default layout for the spec root, and the
gap caused problems. The `spec` skill's authoring workflow named one set of
directories—`adrs/`, `frds/`, `standards/`, `contracts/`, and a shared
`blueprints/`—while this repository's own corpus used a different set:
`features/`, `architecture/`, and `guides/`, with no `frds/` or
`blueprints/`. The skill that vends the convention didn't follow it.

Nothing detected the divergence, because the corpus engine doesn't read
paths. `scanSources` walks the whole spec root and admits any Markdown file
that carries frontmatter. A document's `type` property sets its type, and
its `slug` sets its identity. No document type registration carries a path,
a directory, or a filename pattern. Layout is therefore only a question of
human ergonomics and agent predictability. It constrains nothing the tooling
does, which is why it drifted without consequence until an author had to
guess where a new document belonged.

One constraint makes the question answerable rather than a matter of taste:
this framework already declares containment in frontmatter, and
[[document-relationship-standards]] requires that a relationship be
declared at exactly one end. A directory tree that encodes containment is a
second declaration of the same edge, and no check reconciles the two. Where
a child already declares its parent, the path can only agree redundantly or
disagree silently.

## Decision

We will adopt a default structure in which **the directory tree mirrors
containment only at levels where the containment edge has no child-side
declaration and the child has exactly one parent**. Every other document
type lives in a flat top-level directory.

The following table applies that rule to the three composed levels:

| Level | Parent edge declared in | Directory |
|-------|-------------------------|-----------|
| Container | The child's `system` property | Nested under its system, but only when there are two or more systems |
| Module | Nowhere on the child; each host lists the module in its own `modules` property, and a module can have several hosts | Top level, because a tree can't express multiple parents |
| Component | Nowhere on the child; the parent's roster owns the edge | Nested under its one container or module |

The component is the case that earns nesting. Its blueprint template carries
no parent property, so the path is the only child-side signal of ownership,
and components are the most numerous promoted type under [[adr-007]].

The resulting default structure is as follows:

```
specs/
  adrs/                        one ADR per file
  standards/
  guides/
  contracts/
  systems/
    <system-name>/
      <system-name>-system.md
      containers/              WHEN two or more systems are documented
  containers/
    <container-name>/
      <container-name>-container.md
      components/
        <component-name>-component.md
  modules/
    <module-name>/
      <module-name>-module.md
      components/
        <component-name>-component.md
  features/
    <feature-name>/
      <feature-name>-frd.md
      <feature-name>-blueprint.md
      <subfeature-name>/
```

The system level graduates by the rule that [[c4-standards]] already states
for when a system blueprint exists:

| Systems documented | Where container directories live |
|--------------------|----------------------------------|
| None | `specs/containers/<container-name>/` |
| One | `specs/containers/<container-name>/`; the lone system blueprint is in `specs/systems/<system-name>/` |
| Two or more | `specs/systems/<system-name>/containers/<container-name>/` |

`modules/` stays at the top level in every case, because a module's hosts
can be containers in more than one system.

Standards, guides, ADRs, and contracts each get exactly one top-level home
rather than being filed against the level they govern. If two directories
could each correctly receive a new standard, the document's claim to be the
single place where a rule is recorded becomes false. Additionally, contracts
can't be filed under a provider, because [[adr-010]] establishes that a
contract outlives and can precede the implementations that realize it. A
contract with no provider yet, or with two providers, has no owning
directory.

This structure is a default, not a conformance rule. A project with an
established location for a type keeps that location.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| No default; each project decides | Nothing to migrate;<br>no rule to keep true | An author with no established location guesses;<br>the skill and its own repository already diverged this way | The absence of a default caused the drift that this ADR corrects |
| Full containment tree: systems, containers, modules, and components all nested by parent | The path states the whole hierarchy;<br>a container reads with its parts | Can't express a module's multiple parents;<br>duplicates `system` and `contained_by` with nothing to reconcile them;<br>every promotion under [[adr-007]] becomes a directory move | Encodes edges that the child already declares, in a form that can't represent them |
| Flat `architecture/` directory that holds all four blueprint levels | One directory;<br>promotion is a frontmatter edit and a sibling file | Components are the most numerous promoted type and carry no parent property, so nothing states their owner;<br>hard to scan past a few dozen documents | Loses the one ownership signal that the component level has |
| Mirror containment only where the child doesn't declare its parent | The path carries exactly the edges that frontmatter doesn't;<br>the common case stays shallow;<br>each nesting decision derives from an existing rule | Promotion from component to module is still a directory move;<br>a second system migrates container directories | Not applicable (selected) |

## Consequences

**Positive:**

- Where a document belongs is derivable from rules that the framework
  already states, rather than being a new judgment for each document.
- Containment is declared once. The path and the frontmatter can't disagree
  at the container and module levels, because the path says nothing about
  them.
- A container or module directory collects the unit and its promoted
  components, which is how they're read.
- The layout is checkable: a check can verify a component blueprint's
  directory against its parent's roster, and a container's directory against
  its `system` property. No such check exists yet.

**Negative:**

- Promoting a component to a module ([[adr-007]], and the lifecycle ladder
  in [[c4-standards]]) is a directory move as well as a document change. A
  retired component blueprint is in a different directory from the module
  blueprint that supersedes it.
- Documenting a second system migrates every container directory into
  `systems/<system-name>/containers/`.
- The structure is stated in prose and enforced by nothing.
  [[skill-authoring-standards]] identifies that as the condition under
  which an invariant silently breaks.

## References

- [[adr-007]]: promotion at a seam, which makes component blueprints
  numerous and their creation routine.
- [[adr-010]]: contracts are independent of any provider, so they aren't
  filed under one.
- [[c4-standards]]: the abstraction levels, the rule for when the system
  level is introduced, and the promotion lifecycle.
- [[document-relationship-standards]]: a relationship is declared at
  exactly one end.

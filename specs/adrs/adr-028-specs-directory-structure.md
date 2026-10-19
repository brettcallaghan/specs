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

The framework prescribed no default layout for the spec root, and that gap
had a cost. The `spec` skill's authoring workflow named one set of
directories: `adrs/`, `frds/`, `standards/`, `contracts/`, and a shared
`blueprints/`. This repository's own corpus used a different set:
`features/`, `architecture/`, and `guides/`, with no `frds/` or
`blueprints/` directory. The skill that vends the convention didn't follow
it.

Nothing detected the divergence, because the corpus engine doesn't read
paths. `scanSources` walks the whole spec root and admits any Markdown file
that carries frontmatter. A document's type comes from its `type` property,
and its identity comes from its `slug`. No document type registration
carries a path, a directory, or a filename pattern. Layout therefore
matters only for human ergonomics and agent predictability. It constrains
nothing the tooling does, which is why it drifted without consequence until
an author had to guess where a new document belonged.

One constraint makes the question answerable instead of a matter of taste.
This framework already declares containment in frontmatter, and
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
| Container | the child's `system` property | nested under its system, but only once there are two or more systems |
| Module | nowhere on the child; each host lists it in its own `modules`, and a module may have several hosts | top level; a tree can't express multiple parents |
| Component | nowhere on the child; the parent's roster owns the edge | nested under its one container or module |

The component level is the one that warrants nesting. Its blueprint
template carries no parent property, so the path is the only child-side
signal of ownership. Components are also the most numerous promoted type
under [[adr-007]].

The rule produces the following default structure:

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

Whether container directories nest under a system follows the rule that
[[c4-standards]] already states for when a system blueprint exists at all.
The following table shows where container directories live:

| Systems documented | Where container directories live |
|--------------------|----------------------------------|
| None | `specs/containers/<container-name>/` |
| One | `specs/containers/<container-name>/`. The lone system blueprint is in `specs/systems/<system-name>/`. |
| Two or more | `specs/systems/<system-name>/containers/<container-name>/` |

`modules/` stays at the top level in every case, because a module's hosts
may be containers in more than one system.

Standards, guides, ADRs, and contracts each get exactly one top-level
directory instead of being filed against the level they govern. If two
directories could each correctly receive a new standard, the standard's own
claim to be the single place where a rule is recorded becomes false.

Contracts also must not be filed under a provider. [[adr-010]] establishes
that a contract outlives the implementations that realize it and may
precede them. A contract with no provider yet, or with two providers, has
no owning directory.

This structure is a default, not a conformance rule. A project that already
has an established location for a document type keeps that location.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| No default; each project decides | Nothing to migrate;<br>no rule to keep true | An author with no established location guesses;<br>the skill and its own repository already diverged this way | The absence of a default produced the drift that this ADR corrects |
| Full containment tree: systems, containers, modules, and components all nested by parent | Path states the whole hierarchy;<br>a container reads with its parts | Can't express a module's multiple parents;<br>duplicates `system` and `contained_by` with nothing reconciling them;<br>every promotion under [[adr-007]] becomes a directory move | Encodes edges that the child already declares, in a form that can't represent them |
| Flat `architecture/` holding all four blueprint levels | One directory;<br>promotion is a frontmatter edit and a sibling file | Components are the most numerous promoted type and carry no parent property, so nothing states their owner;<br>hard to scan past a few dozen documents | Loses the one ownership signal that the component level has |
| Mirror containment only where the child doesn't declare its parent | Path carries exactly the edges that frontmatter doesn't;<br>common case stays shallow;<br>each nesting decision derives from an existing rule | Promotion from component to module is still a directory move;<br>a second system migrates container directories | N/A — selected |

## Consequences

**Positive:**

- Authors can derive where a document belongs from rules the framework
  already states, instead of making a fresh judgment for each document.
- Containment is declared once. The path and the frontmatter can't disagree
  at the container and module levels, because the path says nothing about
  them.
- A container or module directory holds the unit together with its
  promoted components, which matches how readers use them.
- The layout is checkable. A check could verify a component blueprint's
  directory against its parent's roster, and a container's directory
  against its `system` property. No such check exists yet.

**Negative:**

- Promoting a component to a module ([[adr-007]], and the lifecycle ladder
  in [[c4-standards]]) requires a directory move as well as a document
  change. A retired component blueprint stays in a different directory from
  the module blueprint that supersedes it.
- Documenting a second system requires moving every container directory
  into `systems/<system-name>/containers/`.
- The structure is stated in prose and enforced by nothing.
  [[skill-authoring-standards]] identifies that as the condition under
  which an invariant silently breaks.

## References

- [[adr-007]] — promotion at a seam, which makes component blueprints
  numerous and their creation routine
- [[adr-010]] — contracts are independent of any provider, so they aren't
  filed under one
- [[c4-standards]] — the abstraction levels, the rule for when the system
  level is introduced, and the promotion lifecycle
- [[document-relationship-standards]] — a relationship is declared at
  exactly one end

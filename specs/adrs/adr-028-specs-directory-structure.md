---
title: "ADR-028: Default Directory Structure for the Spec Corpus"
description: The spec root mirrors the C4 containment hierarchy only where the containment edge has no child-side declaration; everything else is a flat top-level directory
slug: adr-028
type: adr
status: active
updated_at: 2026-09-03T00:00:00Z
---

# ADR-028: Default Directory Structure for the Spec Corpus

**Status:** active  
**Date:** 2026-08-29

## Context

The framework prescribed no default layout for the spec root, and the
gap was not benign. The `spec` skill's authoring workflow named one set
of directories — `adrs/`, `frds/`, `standards/`, `contracts/`, and a
shared `blueprints/` — while this repository's own corpus used a
different set: `features/`, `architecture/`, `guides/`, and no `frds/`
or `blueprints/` at all. The skill that vends the convention did not
follow it.

Nothing detected the divergence, because the corpus engine does not read
paths. `scanSources` walks the whole spec root and admits any markdown
file carrying frontmatter; a document's type comes from its `type`
property and its identity from its `slug`. No document type registration
carries a path, a directory, or a filename pattern. Layout is therefore
a question of human ergonomics and agent predictability only — it
constrains nothing the tooling does, which is exactly why it drifted
without consequence until an author had to guess where a new document
belonged.

The constraint that makes the question answerable rather than a matter
of taste is that this framework already declares containment, in
frontmatter, and [[document-relationship-standards]] requires a
relationship be declared at exactly one end. A directory tree that
encodes containment is a second declaration of the same edge, and no
check reconciles the two. Where a child already declares its parent, the
path can only agree redundantly or disagree silently.

## Decision

We will adopt a default structure in which **the directory tree mirrors
containment only at levels where the containment edge has no child-side
declaration**, and every other document type lives in a flat top-level
directory.

Applying that rule to the three composed levels:

| Level | Parent edge declared in | Directory |
|-------|-------------------------|-----------|
| Container | the child's `system` property | nested under its system, but only once there are two or more systems |
| Module | the child's `contained_by` list — a LIST, because a module may be composed into several containers | top level; a tree cannot express multiple parents |
| Component | nowhere on the child; the parent's roster owns the edge | nested under its one container or module |

The component is the case that earns nesting: its blueprint template
carries no parent property at all, so the path is the only child-side
signal of ownership, and components are the most numerous promoted type
under [[adr-007]].

The resulting default:

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

The system level graduates by the rule [[c4-standards]] already states
for when a system blueprint exists at all:

| Systems documented | Where container directories live |
|--------------------|----------------------------------|
| None | `specs/containers/<container-name>/` |
| One | `specs/containers/<container-name>/` — the lone system blueprint sits in `specs/systems/<system-name>/` |
| Two or more | `specs/systems/<system-name>/containers/<container-name>/` |

`modules/` stays at the top level in every case: a module's
`contained_by` list may name containers in more than one system.

Standards, guides, ADRs and contracts each get exactly one top-level
home rather than being filed against the level they govern. Two
directories that could each correctly receive a new standard make the
document's own claim to be the single place a rule is recorded false.
Contracts additionally may not be filed under a provider, because
[[adr-010]] establishes that a contract outlives and may precede the
implementations that realize it — a contract with no provider yet, or
with two, has no owning directory.

This structure is a default, not a conformance rule. A project with an
established location for a type keeps it.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| No default; each project decides | Nothing to migrate;<br>no rule to keep true | An author with no established location guesses;<br>the skill and its own repository already diverged this way | The absence of a default is what produced the drift this ADR corrects |
| Full containment tree — systems, containers, modules and components all nested by parent | Path states the whole hierarchy;<br>a container reads with its parts | Cannot express a module's multiple parents;<br>duplicates `system` and `contained_by` with nothing reconciling them;<br>every promotion under [[adr-007]] becomes a directory move | Encodes edges the child already declares, in a form that cannot represent them |
| Flat `architecture/` holding all four blueprint levels | One directory; promotion is a frontmatter edit and a sibling file | Components are the most numerous promoted type and carry no parent property, so nothing states their owner;<br>stops scanning past a few dozen documents | Loses the one ownership signal the component level has |
| Mirror containment only where the child does not declare its parent | Path carries exactly the edges frontmatter does not;<br>common case stays shallow;<br>each nesting decision derives from an existing rule | Promotion from component to module is still a directory move;<br>a second system migrates container directories | N/A — selected |

## Consequences

**Positive:**

- Where a document belongs is derivable from rules the framework already
  states, rather than being a fresh judgment per document.
- Containment is declared once. The path and the frontmatter cannot
  disagree at the container and module levels, because the path says
  nothing about them.
- A container or module directory collects the unit and its promoted
  components, which is how they are read.
- The layout is checkable: a component blueprint's directory can be
  verified against its parent's roster, and a container's against its
  `system` property. No such check exists yet.

**Negative:**

- Promotion of a component to a module ([[adr-007]], and the lifecycle
  ladder in [[c4-standards]]) is a directory move as well as a document
  change, and a retired component blueprint sits in a different
  directory from the module blueprint that supersedes it.
- Documenting a second system migrates every container directory into
  `systems/<system-name>/containers/`.
- The structure is stated in prose and enforced by nothing, which
  [[skill-authoring-standards]] identifies as the condition under which
  an invariant silently breaks.

## References

- [[adr-007]] — promotion at a seam, which makes component blueprints
  numerous and their creation routine
- [[adr-010]] — contracts are independent of any provider, so they are
  not filed under one
- [[c4-standards]] — the abstraction levels, the rule for when the
  system level is introduced, and the promotion lifecycle
- [[document-relationship-standards]] — a relationship is declared at
  exactly one end

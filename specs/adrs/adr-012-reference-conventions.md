---
title: "ADR-012: Wikilink and Anchor Reference Conventions"
description: "References use [[slug]] wikilinks for documents, heading anchors for un-promoted components, and backticks for code elements — replacing @ and the # element convention"
slug: adr-012
type: adr
status: ready-for-review
---

# ADR-012: Wikilink and Anchor Reference Conventions

**Status:** ready-for-review  
**Date:** 2026-07-29

## Context

Framework documents reference each other and reference C4 model elements.
Two separate systems grew up for this: `@document-slug` for documents
(defined in `plugin/skills/spec/references/template-authoring-standards.md`)
and a homegrown `#ComponentName` reference ID for elements (defined in the C4
guide). Authors write and read both systems mostly inside coding agents and
edit them in Obsidian-style Markdown tooling, and both interact badly with
those environments:

- In a coding-agent environment, `@name` denotes a file mention that the
  agent resolves against the session root. A document slug isn't a path, so
  `@security-standards` collides with that meaning: agents try to resolve a
  file, and slug-based autocomplete doesn't fire.
- The `#ComponentName` convention has these problems:
  - It's homegrown and isn't part of C4.
  - It collides with Markdown heading syntax, worst inside heading lines
    such as `### #DocumentStore`.
  - It requires a custom PascalCase reference-ID declaration for multi-word
    names.
  - Its qualified form, `[[x]]#ComponentName`, renders in Obsidian as a link
    followed by a tag rather than as a reference.

The framework needs one coherent reference convention for documents and
elements that works in its authoring tools and that a script can validate.

## Decision

We use Obsidian-style wikilinks as the single reference system, and reserve
backticks for code:

- **Documents:** reference a document as `[[slug]]`, never with `@`. The slug
  resolves against the target's frontmatter `slug`, which is a flat,
  location-independent namespace.
- **Promoted elements:** reference an element that has its own blueprint by
  its slug: `[[component-<name>]]` or `[[module-<name>]]`.
- **Un-promoted components:** declare the component under a
  `### <Display Name>` heading in the Components roster of its parent
  container or module. Reference it with an Obsidian heading anchor into that
  parent: `[[container-<name>#Display Name]]`.
- **A section of the referring document:** reference it with an anchor that
  omits the slug: `[[#Terminology]]`.
- **Code-level elements that aren't modeled components:** reference schemas,
  domain types, events, flags, and tables with backtick identifiers
  (`` `ElementName` ``).
- **Framework guides and repository documents outside the set:** reference
  them by path in backticks (`` `references/c4-guide.md` ``).
- **ADRs:** reference an ADR by number only, as `[[adr-<nnn>]]`. The
  descriptive text lives in the ADR's title, so the slug stays stable when
  the ADR is retitled.

This decision removes the `@document-slug` notation and the `#ComponentName`
element reference-ID convention.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Keep `@document-slug` | One sigil; already in use across the corpus | `@` means a file mention in the coding-agent environment that the framework runs in; no closing delimiter, so trailing punctuation is ambiguous to a parser | Collides with the host environment; hard to validate |
| Adopt `#slug` for documents | Short | `#` is a Markdown heading at the start of a line, GitHub's issue-reference sigil, and already claimed for runtime components | Multiple hard collisions |
| Keep the custom `#ComponentName` element convention | A single searchable token per element | Not part of C4; collides with Markdown headings; qualified form renders incorrectly in Obsidian; needs an explicit reference ID on multi-word names | Homegrown complexity plus tool collisions |
| `[[slug]]` wikilinks, heading anchors, and backticks | Established Obsidian convention; unambiguous closing delimiter for clean extraction and validation; element anchors reuse the same system; anchors survive promotion because the roster heading persists | GitHub's native Markdown doesn't render them as links; fragments resolve through project tooling or Obsidian, not a generic renderer | Not applicable (selected) |

## Consequences

**Positive:**

- One reference system (`[[ ]]`) covers documents and un-promoted elements,
  and backticks are reserved for code that isn't a modeled component.
- The closing `]]` delimiter makes every reference unambiguous to extract and
  validate against a frontmatter index, with no trailing-punctuation
  ambiguity.
- A component heading anchor keeps resolving when the component is promoted
  to its own blueprint. The parent's roster heading persists and swaps its
  record block for the blueprint link, so references don't need an immediate
  rewrite.
- References don't collide with the coding-agent `@` file mention or with
  Markdown heading syntax.

**Negative:**

- References resolve through Obsidian or the project's own tooling. GitHub's
  native Markdown renderer doesn't render `[[ ]]`.
- Component references must be qualified by their parent
  (`[[container-<name>#Name]]`). There's no bare element token.
- A multi-word component reference loses the single searchable token that
  the old reference ID gave it. Searches now rely on the heading anchor.
- Reference validation depends on the planned frontmatter-index script. Until
  that script exists, nothing checks `[[ ]]` links.

## References

- `plugin/skills/spec/references/template-authoring-standards.md`: normative
  slug and reference rules.
- `references/c4-guide.md`: element declarations and anchor resolution.
- [[adr-007]]: promotion-based blueprint creation, which is why anchors must
  survive promotion.
- [[adr-008]]: machine-readable element record blocks.

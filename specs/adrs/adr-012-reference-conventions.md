---
title: "ADR-012: Wikilink and Anchor Reference Conventions"
description: "References use [[slug]] wikilinks for documents, heading anchors for un-promoted components, and backticks for code elements — replacing @ and the # element convention"
slug: adr-012
type: adr
status: ready-for-review
updated_at: 2026-09-03T00:00:00Z
---

# ADR-012: Wikilink and Anchor Reference Conventions

**Status:** ready-for-review  
**Date:** 2026-07-29

## Context

Framework documents reference each other and reference C4 model
elements, and two separate systems had grown up for it: `@document-slug`
for documents (defined in
`plugin/skills/spec/references/template-authoring-standards.md`) and a
homegrown `#ComponentName` reference-id for elements (defined in the C4
guide). Both are authored and consumed primarily inside coding agents
and edited in Obsidian-style Markdown tooling, and both interact badly
with those environments:

- In a coding-agent environment `@name` denotes a filesystem file-mention
  resolved against the session root. A document slug is not a path, so
  `@security-standards` collides with that meaning — agents try to
  resolve a file, and slug-based autocomplete does not fire.
- The `#ComponentName` convention was homegrown (not part of C4),
  collided with Markdown heading syntax (worst inside heading lines like
  `### #DocumentStore`), required a bespoke PascalCase reference-id
  declaration for multi-word names, and its qualified form
  `[[x]]#ComponentName` mis-renders in Obsidian as a link followed by a
  tag rather than a reference.

The framework needs one coherent reference convention — across documents
and elements — that survives its authoring tools and is mechanically
validatable.

## Decision

We will use Obsidian-style wikilinks as the single reference system, with
backticks reserved for code:

- **Documents** are referenced `[[slug]]`, resolved against the target's
  frontmatter `slug` (a flat, location-independent namespace), never `@`.
- **Promoted elements** (those with their own blueprint) are referenced
  by their slug: `[[component-<name>]]` / `[[module-<name>]]`.
- **Un-promoted components** are declared under a `### <Display Name>`
  heading in their parent container's or module's Components roster, and
  referenced by an Obsidian heading anchor into that parent:
  `[[container-<name>#Display Name]]`.
- **A section of the referring document itself** is referenced by an
  anchor with the slug omitted: `[[#Terminology]]`.
- **Code-level elements that are not modeled components** — schemas,
  domain types, events, flags, tables — use backtick identifiers
  (`` `ElementName` ``).
- **Framework guides and out-of-set repository docs** are referenced by
  path in backticks (`` `references/c4-guide.md` ``).
- **ADRs** are referenced number-only, `[[adr-<nnn>]]`; the descriptive
  text lives in the ADR's title, so its slug is stable across retitling.

The `@document-slug` notation and the `#ComponentName` element
reference-id convention are both removed.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Keep `@document-slug` | One sigil; already in use across the corpus | `@` means a file-mention in the coding-agent environment the framework runs in; no closing delimiter, so trailing punctuation is ambiguous to a parser | Collides with the host environment; weak to validate |
| Adopt `#slug` for documents | Terse | `#` is a Markdown heading at line start, GitHub's issue-reference sigil, and was already claimed for runtime components | Multiple hard collisions |
| Keep the bespoke `#ComponentName` element convention | A single greppable token per element | Not part of C4; collides with Markdown headings; qualified form mis-renders in Obsidian; needs an explicit reference-id on multi-word names | Homegrown complexity plus tool collisions |
| `[[slug]]` wikilinks + heading anchors + backticks | Established Obsidian convention; unambiguous closing delimiter for clean extraction/validation; element anchors reuse the same system; anchors survive promotion because the roster heading persists | Not rendered as links by GitHub's native Markdown; fragments resolve through project tooling/Obsidian, not a generic renderer | N/A — selected |

## Consequences

**Positive:**
- One reference system (`[[ ]]`) spans documents and un-promoted
  elements, with backticks cleanly reserved for code that is not a
  modeled component.
- The closing `]]` delimiter makes every reference unambiguous to
  extract and validate against a frontmatter index, with no
  trailing-punctuation ambiguity.
- A component heading anchor keeps resolving when the component is
  promoted to its own blueprint: the parent's roster heading persists and
  simply swaps its record block for the blueprint link, so references do
  not need an immediate rewrite.
- No collision with the coding-agent `@` file-mention or with Markdown
  heading syntax.

**Negative:**
- References resolve through Obsidian or the project's own tooling, not
  GitHub's native Markdown renderer, which does not render `[[ ]]`.
- Component references must be parent-qualified
  (`[[container-<name>#Name]]`); there is no bare element token.
- A multi-word component reference loses the single greppable token the
  old reference-id gave it; greppability now rides on the heading anchor.
- Reference validation depends on the planned frontmatter-index script
  existing; until it does, `[[ ]]` links are unchecked.

## References

- `plugin/skills/spec/references/template-authoring-standards.md` — normative
  slug and reference rules
- `references/c4-guide.md` — element declarations and anchor resolution
- [[adr-007]] — promotion-based blueprint creation (why anchors must
  survive promotion)
- [[adr-008]] — machine-readable element record blocks

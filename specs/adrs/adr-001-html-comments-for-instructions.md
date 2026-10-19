---
title: "ADR-001: Use HTML Comment Blocks for Template Instructions"
description: Use HTML comment blocks to embed guidance in specification templates
slug: adr-001
type: adr
status: active
---

# ADR-001: Use HTML Comment Blocks for Template Instructions

**Status:** active  
**Date:** 2026-07-16

## Context

Specification templates need to guide the users who fill them out. The
guidance covers each section's purpose, formatting requirements, examples, and
when to use or omit a section. Users need to see the instructions while editing
so they understand what to write, but the instructions shouldn't clutter the
final rendered document. The format must work across different Markdown
editors, version control systems, and rendering engines. It must not break the
document structure or require specialized tooling.

## Decision

We will use HTML comment blocks (`<!-- ... -->`) for all instructional content,
guidance, examples, and meta-commentary within specification templates.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Markdown blockquotes with "Note:" prefix | Highly visible in editors<br>Native Markdown syntax | Renders in final output unless manually removed<br>Clutters rendered documents<br>Harder to distinguish from actual quoted content | Instructions would appear in rendered specs, creating visual noise |
| Separate instruction documents | Clean separation of concerns<br>No risk of instructions in final docs | Users must context-switch between files<br>Instructions get out of sync with templates<br>Increases file count | Breaks the "fill in the template" workflow;<br>high friction for users |
| Inline italic text with markers | Simple to implement<br>Visible in all editors | Renders in output<br>Hard to programmatically detect and remove<br>No clear visual boundary | Cannot be cleanly removed;<br>would mix with legitimate emphasized text |
| HTML comment blocks | Hidden in rendered output<br>Standard HTML syntax<br>Easy to remove manually or programmatically<br>Works in all Markdown processors | Less visually prominent than blockquotes in some editors<br>Requires users to remember to remove them | N/A — selected |

## Consequences

**Positive:**
- Instructions are visible when users edit raw Markdown files.
- Rendered documents stay clean even if someone leaves comments in by mistake,
  because comments don't render in HTML or PDF output.
- The syntax is standard HTML, with universal support across Markdown
  processors.
- Authors can remove comments programmatically with simple regex patterns, or
  manually with search and replace.
- The `<!--` and `-->` delimiters give a clear visual boundary between
  instructions and content.

**Negative:**
- Comments are less prominent than blockquotes in editors that don't
  highlight HTML comment syntax.
- Users must remember to remove comments before they consider a document
  complete. The templates call this out explicitly.
- Comments increase the raw file size until they're removed.
- Content that users accidentally write inside a comment block could cause
  confusion.

## References

- Guidelines for removing template instructions, in the template headers
- Examples of this usage in all templates under
  `/plugin/skills/spec/references/`

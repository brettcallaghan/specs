---
title: "ADR-001: Use HTML Comment Blocks for Template Instructions"
description: Use HTML comment blocks to embed guidance in specification templates
slug: adr-001
type: adr
status: active
---

# ADR-001: Use HTML comment blocks for template instructions

**Status:** active  
**Date:** 2026-07-16

## Context

Specification templates need to guide the users who fill them out. The
guidance explains each section's purpose, the formatting requirements, and
when to use or omit a section, and it gives examples. Users need to see the
instructions while they edit so they know what to write, but the instructions
should not clutter the final rendered document. The format must work across
Markdown editors, version control systems, and rendering engines. It must not
break the document structure or require specialized tooling.

## Decision

We will use HTML comment blocks (`<!-- ... -->`) for all instructional content,
guidance, examples, and meta-commentary within specification templates.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Markdown blockquotes with a "Note:" prefix | Highly visible in editors<br>Native Markdown syntax | Renders in the final output unless removed by hand<br>Clutters rendered documents<br>Harder to tell apart from real quoted content | Instructions would appear in rendered specs as visual noise |
| Separate instruction documents | Clean separation of concerns<br>No risk of instructions in final docs | Users must switch between files<br>Instructions drift out of sync with templates<br>More files | Breaks the "fill in the template" workflow<br>High friction for users |
| Inline italic text with markers | Simple to implement<br>Visible in all editors | Renders in the output<br>Hard to detect and remove programmatically<br>No clear visual boundary | Can't be removed cleanly<br>Would mix with legitimate emphasized text |
| HTML comment blocks | Hidden in rendered output<br>Standard HTML syntax<br>Easy to remove by hand or programmatically<br>Works in all Markdown processors | Less prominent than blockquotes in some editors<br>Users must remember to remove them | N/A — selected |

## Consequences

**Positive:**
- Instructions are visible when you edit the raw Markdown file.
- Rendered documents stay clean even if comments are accidentally left in,
  because comments don't render in HTML or PDF output.
- HTML comments are standard HTML syntax that every Markdown processor
  supports.
- Comments can be removed programmatically with a simple regular expression,
  or by hand with search and replace.
- The `<!--` and `-->` delimiters draw a clear visual boundary between
  instructions and content.

**Negative:**
- Comments are less prominent than blockquotes in editors that don't
  highlight HTML comment syntax.
- Users must remember to remove comments before they consider a document
  complete. Templates say so explicitly.
- Comments increase the raw file size until they're removed.
- Content that users accidentally write inside a comment block could cause
  confusion.

## References

- Each template's header comment explains when to remove its instructions.
- Every template under `plugin/skills/spec/references/` uses comment blocks.

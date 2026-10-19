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

Specification templates need to guide the users who fill them out. This
guidance includes the purpose of each section, formatting requirements,
examples, and when to use or omit a section. Users need to see the
instructions while they edit, so that they understand what to write, but the
instructions shouldn't clutter the final rendered document. The format must
work across different Markdown editors, version control systems, and rendering
engines without breaking the document structure or requiring specialized
tooling.

## Decision

We use HTML comment blocks (`<!-- ... -->`) for all instructional content,
guidance, examples, and meta-commentary in specification templates.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Markdown blockquotes with a "Note:" prefix | Highly visible in editors.<br>Native Markdown syntax. | Renders in final output unless removed manually.<br>Clutters rendered documents.<br>Harder to distinguish from actual quoted content. | Instructions would appear in rendered specs and create visual noise. |
| Separate instruction documents | Clean separation of concerns.<br>No risk of instructions in final documents. | Users must switch context between files.<br>Instructions get out of sync with templates.<br>Increases the file count. | Breaks the "fill in the template" workflow;<br>high friction for users. |
| Inline italic text with markers | Simple to implement.<br>Visible in all editors. | Renders in output.<br>Hard to detect and remove programmatically.<br>No clear visual boundary. | Can't be removed cleanly;<br>would mix with legitimate emphasized text. |
| HTML comment blocks | Hidden in rendered output.<br>Standard HTML syntax.<br>Easy to remove manually or programmatically.<br>Works in all Markdown processors. | Less visually prominent than blockquotes in some editors.<br>Users must remember to remove them. | Not applicable; selected. |

## Consequences

**Positive:**

- Instructions are visible when you edit raw Markdown files.
- Rendered documents stay clean even if comments are accidentally left in,
  because comments don't render in HTML or PDF output.
- HTML comment syntax is standard and has universal support across Markdown
  processors.
- You can remove comments programmatically with simple regular expressions, or
  manually with search and replace.
- The `<!--` and `-->` delimiters create a clear visual boundary between
  instructions and content.

**Negative:**

- Comments are less prominent than blockquotes in editors that don't highlight
  HTML comment syntax.
- Users must remember to remove comments before they consider a document
  complete. The templates call out this step explicitly.
- Comments increase the raw file size until they're removed.
- Users who accidentally write content inside comment blocks might be
  confused.

## References

- Guidelines for removing template instructions, documented in template
  headers
- Example usage in all templates under `/plugin/skills/spec/references/`

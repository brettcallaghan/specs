---
title: "ADR-001: Use HTML Comment Blocks for Template Instructions"
description: Use HTML comment blocks to embed guidance in specification templates
slug: adr-001
type: adr
status: active
updated_at: 2026-08-05T00:00:00Z
---

# ADR-001: Use HTML Comment Blocks for Template Instructions

**Status:** active  
**Date:** 2026-07-16

## Context

Specification templates need to provide guidance to users filling them out.
This includes explanations of each section's purpose, formatting requirements,
examples, and when to use or omit sections. The instructions need to be visible
during editing so users understand what to write, but should not clutter the
final rendered document. The format must work across different markdown editors,
version control systems, and rendering engines without breaking the document
structure or requiring specialized tooling.

## Decision

We will use HTML comment blocks (`<!-- ... -->`) for all instructional content,
guidance, examples, and meta-commentary within specification templates.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Markdown blockquotes with "Note:" prefix | Highly visible in editors<br>Native markdown syntax | Renders in final output unless manually removed<br>Clutters rendered documents<br>Harder to distinguish from actual quoted content | Instructions would appear in rendered specs, creating visual noise |
| Separate instruction documents | Clean separation of concerns<br>No risk of instructions in final docs | Users must context-switch between files<br>Instructions get out of sync with templates<br>Increases file count | Breaks the "fill in the template" workflow;<br>high friction for users |
| Inline italic text with markers | Simple to implement<br>Visible in all editors | Renders in output<br>Hard to programmatically detect and remove<br>No clear visual boundary | Cannot be cleanly removed;<br>would mix with legitimate emphasized text |
| HTML comment blocks | Hidden in rendered output<br>Standard HTML syntax<br>Easy to remove manually or programmatically<br>Works in all markdown processors | Less visually prominent than blockquotes in some editors<br>Requires users to remember to remove them | N/A — selected |

## Consequences

**Positive:**
- Instructions are visible when editing raw markdown files
- Rendered documents are clean even if comments are accidentally left in
  (comments don't render in HTML/PDF output)
- Standard HTML syntax with universal support across markdown processors
- Can be removed programmatically with simple regex patterns or manually with
  search-and-replace
- Clear visual boundary between instructions and content (`<!--` and `-->`
  delimiters)

**Negative:**
- Less prominent than blockquotes in editors without syntax highlighting for
  HTML comments
- Users must remember to remove comments before considering a document
  "complete" (though this is explicitly called out in templates)
- Increases raw file size until comments are removed
- Could create confusion if users accidentally write content inside comment
  blocks

## References

- Template instruction removal guidelines documented in template headers
- Example usage in all templates under `/plugin/skills/spec/references/`

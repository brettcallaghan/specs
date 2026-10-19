---
title: "ADR-003: Use YAML Frontmatter for Document Metadata"
description: Store structured metadata in YAML frontmatter blocks at the start of specification documents
slug: adr-003
type: adr
status: active
---

# ADR-003: Use YAML frontmatter for document metadata

**Status:** active  
**Date:** 2026-07-16

## Context

Specification documents need structured metadata to enable automation and
navigation. This metadata includes the title, description, document type,
status, timestamps, and slugs for cross-references. Tools that index,
validate, and resolve cross-references must be able to read the metadata, and
people must be able to read and edit it by hand. The metadata must be clearly
separated from the document body so that it doesn't cause confusion and can be
parsed reliably. The solution must work with standard Markdown tools, and it
must not require specialized editors or proprietary formats.

## Decision

We will store structured metadata in a YAML frontmatter block, delimited by
`---`, at the beginning of every specification document. The frontmatter
block precedes all Markdown content and contains key-value pairs in YAML
format.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| JSON metadata blocks | Strict schema validation<br>Native JavaScript compatibility<br>Widely supported parsing libraries | Harder to read than YAML<br>Requires quotation marks around all strings<br>No support for comments<br>More verbose syntax | Poor editing experience;<br>strict syntax makes manual editing harder |
| Markdown tables at the start of the document | Native Markdown syntax<br>No special parser needed for basic use | Hard to parse reliably<br>Poor support for nested data<br>Not machine-readable without custom parsing<br>Hard to read for complex metadata | Not designed for structured data;<br>parsing is brittle and error-prone |
| Metadata in filenames only | No special syntax needed<br>Simple to implement<br>Works with any file | Holds very little metadata<br>Hard to keep consistent<br>Renaming a file breaks references<br>Can't store complex or nested data | Can't express the metadata that documents need |
| YAML frontmatter | Standard in Markdown tools (Jekyll, Hugo, Obsidian)<br>Easy to read and edit<br>Supports nested data structures<br>Clean separation from content<br>Many libraries can parse it | Tools need a YAML-aware parser<br>Syntax errors break parsing<br>Basic Markdown viewers don't render it (it's visible only in the raw file) | N/A—selected |

## Consequences

**Positive:**

- YAML frontmatter is standard practice in static site generators (Jekyll,
  Hugo, Gatsby) and Markdown tools (Obsidian, Foam).
- The syntax is easy to read; most strings don't need quotation marks.
- People can edit it by hand without specialized tools.
- Well-maintained libraries in most programming languages can parse it, for
  example PyYAML, js-yaml, and gopkg.in/yaml.
- Metadata is cleanly separated from the document body.
- It supports complex data structures, such as nested objects and arrays,
  when they're needed.
- It's hidden in rendered output, so it doesn't clutter the rendered document.
- It enables automation: indexing, validation, cross-reference resolution, and
  status tracking.

**Negative:**

- Any tool that processes the metadata needs a YAML-aware parser.
- YAML syntax errors, such as wrong indentation or unescaped special
  characters, break parsing and can be hard to spot.
- Basic Markdown viewers don't render it. The metadata is visible only in the
  raw file or in a renderer that supports frontmatter.
- The title is duplicated: the frontmatter `title` and the H1 in the body must
  be kept in sync.
- The indentation-sensitive syntax can frustrate people who are unfamiliar
  with YAML.
- Some processors fail silently on empty or invalid frontmatter.

## References

- All templates under `/plugin/skills/spec/references/` show how to use
  frontmatter.
- `plugin/skills/spec/references/template-authoring-standards.md`: the
  standard fields used across document types.
- [YAML 1.2 specification](https://yaml.org/spec/1.2/spec.html).
- The conventions of static site generators such as Jekyll and Hugo set a
  precedent for this pattern.

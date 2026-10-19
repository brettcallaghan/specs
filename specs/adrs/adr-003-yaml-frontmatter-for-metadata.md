---
title: "ADR-003: Use YAML Frontmatter for Document Metadata"
description: Store structured metadata in YAML frontmatter blocks at the start of specification documents
slug: adr-003
type: adr
status: active
---

# ADR-003: Use YAML Frontmatter for Document Metadata

**Status:** active  
**Date:** 2026-07-16

## Context

Specification documents need structured metadata: title, description,
document type, status, timestamps, slugs for cross-references, and other
fields that support automation and navigation. Tools use this metadata for
indexing, validation, and cross-reference resolution, so it must be
machine-readable. People edit it by hand, so it must also stay
human-readable. The metadata should be clearly separated from the document
body, which prevents confusion and makes parsing reliable. The solution must
work with standard Markdown tooling and must not require specialized editors
or proprietary formats.

## Decision

We will use YAML frontmatter blocks, delimited by `---`, at the beginning of
all specification documents to store structured metadata. The frontmatter
block precedes all Markdown content and contains key-value pairs in YAML
format.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| JSON metadata blocks | Strict schema validation<br>Native JavaScript compatibility<br>Widely supported parsing libraries | Less human-readable than YAML<br>Requires quotes around all strings<br>No comment support<br>More verbose syntax | Poor editing experience;<br>strict syntax creates friction for manual editing |
| Markdown tables at document start | Native Markdown syntax<br>No special parser needed for basic use | Hard to parse reliably<br>Poor support for nested data<br>Not machine-readable without custom parsing<br>Hard to read for complex metadata | Not designed for structured data;<br>parsing is brittle and error-prone |
| Filename-based metadata only | No special syntax needed<br>Simple to implement<br>Works with any file | Very limited metadata capacity<br>Hard to keep consistent<br>Filename changes break references<br>Can't store complex or nested data | Insufficient for real-world needs;<br>can't express the metadata the corpus requires |
| YAML frontmatter | Standard in Markdown tooling (Jekyll, Hugo, Obsidian)<br>Human-readable and easy to edit<br>Supports nested data structures<br>Clean separation from content<br>Parseable by many libraries | Requires YAML-aware parsers for tooling<br>Syntax errors break parsing<br>Not rendered by basic Markdown viewers (invisible unless you view the raw file) | N/A — selected |

## Consequences

**Positive:**
- Standard practice in static site generators (Jekyll, Hugo, Gatsby) and
  Markdown tools (Obsidian, Foam)
- Human-readable, with clean syntax (most strings need no quotes)
- Easy to edit by hand without specialized tooling
- Parseable by well-maintained libraries in most programming languages, such
  as PyYAML, js-yaml, and gopkg.in/yaml
- Clean separation between metadata and document body
- Supports complex data structures, such as nested objects and arrays, when
  needed
- Hidden in rendered output, so it doesn't clutter the rendered document
- Enables automation: indexing, validation, cross-reference resolution, and
  status tracking

**Negative:**
- Any tooling that processes the metadata needs a YAML-aware parser
- YAML syntax errors in indentation or special characters break parsing
  and can be hard to spot
- Basic Markdown viewers don't render it, so the metadata is invisible unless
  you view the raw file or use a frontmatter-aware renderer
- Duplicates the title: the frontmatter title and the H1 in the body must be
  kept in sync
- Indentation-sensitive syntax can frustrate users unfamiliar with YAML
- Some processors fail silently on empty or invalid frontmatter

## References

- All templates under `/plugin/skills/spec/references/` demonstrate
  frontmatter usage
- `plugin/skills/spec/references/template-authoring-standards.md` — the
  standardized fields used across document types
- [YAML 1.2 specification](https://yaml.org/spec/1.2/spec.html)
- Static site generator conventions (Jekyll, Hugo) set a precedent for this
  pattern

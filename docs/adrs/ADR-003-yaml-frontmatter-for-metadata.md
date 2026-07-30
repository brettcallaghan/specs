---
title: "ADR-003: Use YAML Frontmatter for Document Metadata"
description: Store structured metadata in YAML frontmatter blocks at the start of specification documents
slug: adr-003-yaml-frontmatter-for-metadata
type: adr
status: active
updated_at: 2026-07-29T00:00:00Z
---

# ADR-003: Use YAML Frontmatter for Document Metadata

**Status:** active  
**Date:** 2026-07-16

## Context

Specification documents need structured metadata including title, description,
document type, status, timestamps, slugs for cross-references, and other fields
that enable automation and navigation. This metadata must be machine-readable
for tooling (indexing, validation, cross-reference resolution) while remaining
human-readable for manual editing. The metadata should be clearly separated from
the document body to prevent confusion and enable reliable parsing. The solution
must work with standard markdown tooling and not require specialized editors or
proprietary formats.

## Decision

We will use YAML frontmatter blocks (delimited by `---`) at the beginning of all
specification documents to store structured metadata. The frontmatter block
precedes all markdown content and contains key-value pairs in YAML format.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| JSON metadata blocks | Strict schema validation<br>Native JavaScript compatibility<br>Widely supported parsing libraries | Less human-readable than YAML<br>Requires quotes around all strings<br>No comment support<br>More verbose syntax | Poor editing experience;<br>syntax strictness creates friction for manual editing |
| Markdown tables at document start | Native markdown syntax<br>No special parser needed for basic use | Hard to parse reliably<br>Poor support for nested data<br>Not machine-readable without custom parsing<br>Ugly for complex metadata | Not designed for structured data;<br>parsing is brittle and error-prone |
| Filename-based metadata only | No special syntax needed<br>Simple to implement<br>Works with any file | Very limited metadata capacity<br>Hard to maintain consistency<br>Filename changes break references<br>Cannot store complex or nested data | Insufficient for real-world needs;<br>cannot express required metadata richness |
| YAML frontmatter | Standard in markdown tooling (Jekyll, Hugo, Obsidian)<br>Human-readable and easy to edit<br>Supports nested data structures<br>Clean separation from content<br>Parseable by numerous libraries | Requires YAML-aware parsers for tooling<br>Syntax errors break parsing<br>Not rendered by basic markdown viewers (invisible unless viewing raw file) | N/A — selected |

## Consequences

**Positive:**
- Standard practice in static site generators (Jekyll, Hugo, Gatsby) and
  markdown tools (Obsidian, Foam)
- Human-readable with clean syntax (no quotes required for most strings)
- Easy to edit manually without specialized tooling
- Parseable by well-maintained libraries in most programming languages (PyYAML,
  js-yaml, gopkg.in/yaml, etc.)
- Clean separation between metadata and document body
- Supports complex data structures (nested objects, arrays) when needed
- Hidden in rendered output (doesn't clutter the rendered document)
- Enables automation: indexing, validation, cross-reference resolution, status
  tracking

**Negative:**
- Requires YAML-aware parsers for any tooling that processes the metadata
- YAML syntax errors (indentation, special characters) break parsing and can be
  subtle
- Not rendered by basic markdown viewers (metadata invisible unless viewing raw
  file or using frontmatter-aware renderer)
- Duplicates title information (frontmatter title and H1 in body must be kept in
  sync)
- Indentation-sensitive syntax can frustrate users unfamiliar with YAML
- Empty or invalid frontmatter silently fails in some processors

## References

- All templates under `/plugin/skills/spec/references/` demonstrate frontmatter
  usage
- [[template-authoring-standards]] — standardized fields used across document types
- YAML 1.2 specification: https://yaml.org/spec/1.2/spec.html
- Static site generator conventions (Jekyll, Hugo) establish precedent for this
  pattern

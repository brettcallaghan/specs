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

Specification documents need structured metadata: title, description,
document type, status, timestamps, slugs for cross-references, and other
fields that enable automation and navigation. Tooling must be able to read this
metadata for indexing, validation, and cross-reference resolution, and people
must be able to read and edit it by hand. The metadata should be clearly
separate from the document body, so that it doesn't confuse readers and parses
reliably. The format must work with standard Markdown tooling and must not
require specialized editors or proprietary formats.

## Decision

We will use YAML frontmatter blocks (delimited by `---`) at the beginning of all
specification documents to store structured metadata. The frontmatter block
precedes all markdown content and contains key-value pairs in YAML format.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| JSON metadata blocks | Strict schema validation<br>Native JavaScript compatibility<br>Widely supported parsing libraries | Harder for people to read than YAML<br>Every string needs quotes<br>No comments<br>More verbose syntax | Poor editing experience<br>Strict syntax makes editing by hand harder |
| Markdown tables at document start | Native Markdown syntax<br>No special parser needed for basic use | Hard to parse reliably<br>Poor support for nested data<br>Tools can't read it without a custom parser<br>Ugly for complex metadata | Not designed for structured data<br>Parsing is brittle and error-prone |
| Filename-based metadata only | No special syntax needed<br>Simple to implement<br>Works with any file | Holds very little metadata<br>Hard to keep consistent<br>Renaming a file breaks references<br>Can't store complex or nested data | Not enough for real-world needs<br>Can't express the metadata the documents require |
| YAML frontmatter | Standard in Markdown tooling (Jekyll, Hugo, Obsidian)<br>Easy for people to read and edit<br>Supports nested data structures<br>Cleanly separate from content<br>Many libraries parse it | Tooling needs a YAML-aware parser<br>Syntax errors break parsing<br>Basic Markdown viewers don't render it (invisible unless you view the raw file) | N/A — selected |

## Consequences

**Positive:**
- Frontmatter is standard practice in static site generators (Jekyll, Hugo,
  Gatsby) and Markdown tools (Obsidian, Foam).
- The syntax is clean and easy to read. Most strings need no quotes.
- People can edit it by hand without specialized tooling.
- Well-maintained libraries parse it in most programming languages, such as
  PyYAML, js-yaml, and gopkg.in/yaml.
- Metadata is cleanly separate from the document body.
- It supports complex data structures, such as nested objects and arrays,
  when needed.
- It's hidden in rendered output, so it doesn't clutter the rendered document.
- It enables automation: indexing, validation, cross-reference resolution, and
  status tracking.

**Negative:**
- Any tooling that processes the metadata needs a YAML-aware parser.
- YAML syntax errors, such as bad indentation or special characters, break
  parsing and can be subtle.
- Basic Markdown viewers don't render it. The metadata is invisible unless you
  view the raw file or use a renderer that reads frontmatter.
- The title appears twice: the frontmatter title and the H1 in the body must
  be kept in sync.
- Indentation-sensitive syntax can frustrate users who don't know YAML.
- Some processors fail silently on empty or invalid frontmatter.

## References

- All templates under `/plugin/skills/spec/references/` demonstrate frontmatter
  usage
- `plugin/skills/spec/references/template-authoring-standards.md` —
  standardized fields used across document types
- YAML 1.2 specification: https://yaml.org/spec/1.2/spec.html
- Static site generator conventions (Jekyll, Hugo) set the precedent for this
  pattern

---
title: Frontmatter Standards
description: Required, relationship, and optional frontmatter properties for
  every spec document type, plus the embedded element-record blocks a parent
  document uses to declare un-promoted children
slug: frontmatter-standards
type: standards
status: active
updated_at: 2026-08-06T00:00:00Z
version: "0.2"
---

# Frontmatter Standards

## Contents

- Format — YAML delimiters and placement
- Required Properties — title, description, slug, type, status, updated_at,
  including the full `status` vocabulary and the state-vs-transient rule
- Relationship Properties — provenance, composition back-pointers, and the
  typed edges a consumer keys on
- Optional Properties — version, created_at, tags, superseded_by
- Embedded Element Records — the record-block XOR blueprint-link rule for
  declaring child components

Companion to `template-authoring-standards.md`, which owns document
structure, slugs, instruction blocks, versioning, and content guidelines.
Cross-document reference syntax is in `cross-document-references.md`.

This document is the normative prose. The **per-type property tables**
derived from it — which properties each document type carries, and
which are required — are generated into `generated/frontmatter.md` from
the schemas in `schemas/`, alongside the JSON Schema that validates
them. WHEN this document and a generated table disagree, the schema is
the one tooling enforces, and the disagreement is a defect in one of
them to be resolved rather than a precedence rule to apply.

## Metadata (Frontmatter)

### Format

YAML frontmatter SHALL be used to provide structured metadata for documents 
created from templates.

The frontmatter block SHALL begin at the first line of the document.

The frontmatter block SHALL be delimited by triple-dash lines (`---`).

**Example:**
```yaml
---
title: Example Document
description: A brief summary of what this document contains
slug: example-document
type: example
status: draft
updated_at: 2026-07-16T14:30:00Z
---
```

**Rationale:** YAML frontmatter is a widely-adopted convention (Jekyll, Hugo, 
Obsidian) that provides machine-parseable metadata separate from content.

### Required Properties

Templates SHALL include the following metadata properties in their frontmatter:

```yaml
---
title: <title of the document>
description: <a brief summary of the document>
slug: <used to reference the document from other documents>
type: <document-type>
status: <status-value>
updated_at: <RFC 3339 date time in which the document was last updated>
---
```

#### Property Definitions

| Property | Type | Purpose | Format |
|----------|------|---------|--------|
| `title` | string | Human-readable document title | Free text |
| `description` | string | Brief summary of the document | 1-2 sentences |
| `slug` | string | Unique identifier for cross-references | Kebab-case (lowercase with hyphens) |
| `type` | string | Document classification | Controlled vocabulary defined by document system |
| `status` | string | Lifecycle state | One of: `draft`, `ready-for-review`, `approved`, `in-progress`, `active`, `retired` |
| `updated_at` | datetime | Last modification timestamp | RFC 3339 format: `YYYY-MM-DDTHH:MM:SSZ` |

The `type` property SHALL use a controlled vocabulary defined by the document 
system using the templates.

The `status` property SHALL use the following controlled vocabulary, common 
to all document types:

| Status | Meaning |
|--------|---------|
| `draft` | Under authorship; content may be incomplete |
| `ready-for-review` | Complete and awaiting review |
| `approved` | Reviewed and accepted; not yet in effect |
| `in-progress` | Approved, and the change it specifies is being implemented |
| `active` | In effect; the current source of truth |
| `retired` | No longer in effect |

Status values SHALL be lowercase with hyphen separators (kebab-case).

State documents (blueprints, standards, contracts — documents that
describe what IS) use `active` and never `in-progress`. Transient
change documents (e.g., feature blueprints — documents that describe a
change) use `in-progress` in place of `active`: a change specification
is implemented, not "in effect."

### Relationship Properties

A relationship property names other documents by `slug`. Which properties
exist is owned by [[document-relationship-standards]] and by the schemas in
`schemas/`; this section states how a template declares one.

WHEN a document has a relationship a body citation cannot carry, the template
SHALL include a relationship property for it in the frontmatter:

```yaml
---
# Provenance (what this is derived from)
sources: []                                 # Source document slugs

# Composition — declared on the CHILD, never on the parent
system: "<system slug>"                     # container -> its system
contained_by: []                            # module -> its host container(s)
container: "<container or module slug>"     # component -> its parent

# Typed edges a machine consumer keys on
depends_on: []                              # system -> system dependency
contracts_provided: []                      # contract slugs this provides
contracts_consumed: []                      # contract slugs this consumes
---
```

WHEN a document does not have relationships, these properties SHALL be omitted 
from the template.

All relationship properties SHALL reference documents by their `slug` value.

A template SHALL NOT introduce a relationship property that only mirrors the
document's own prose citations. Per [[document-relationship-standards]], a
relationship is declared at exactly one end and the index derives the reverse,
so a property earns its place only if it carries a typed relationship a
consumer keys on, names a target not reliably present in the prose, or is the
only home for the edge.

WHEN a document describes a change to the document set itself (e.g., a
feature blueprint that will modify architecture blueprints and author new
contracts), the documents it creates and updates SHALL be named in its prose
and tracked in its `plan`, not mirrored into frontmatter.

**Rationale:** Relationship properties enable traceability, dependency
management, and impact analysis across documents — but only where the graph
derived from prose citations cannot already answer the question. A parent's
membership roster is derivable from its children's back-pointers, and a
change set is tracked to completion by the plan, so neither is carried here.

### Optional Properties

Templates MAY include additional properties based on the needs of the document 
system:

```yaml
---
version: "0.1"                     # Semantic version
created_at: <RFC 3339 datetime>    # Creation timestamp
tags: []                           # Free-form tags for filtering/search
superseded_by: "<slug>"            # For superseded documents
---
```

WHEN a document system uses semantic versioning, templates SHALL include a 
`version` property.

WHEN a document lifecycle includes supersession, templates SHALL include a 
`superseded_by` property for use when a document is retired in favor of a 
replacement.

| Property | Type | Purpose | Format |
|----------|------|----------|--------|
| `version` | string | Document version | Semantic versioning: `MAJOR.MINOR.PATCH` |
| `created_at` | datetime | Creation timestamp | RFC 3339 format |
| `tags` | array | Free-form tags | Array of strings |
| `superseded_by` | string | Superseding document | Slug of replacement document |

Architecture blueprint templates (system, container, module, component)
SHALL include `superseded_by`: their lifecycle includes retirement and
migration between levels, and a retired blueprint points to its
successor.

---

## Embedded Element Records

WHEN a document declares child elements it composes — the components of a
container or a module — each child is recorded in one of two states, and
a document system MAY make that record machine-readable:

- **Un-promoted** (no blueprint of its own): the child's declaration
  SHALL carry a fenced record block holding the base record fields
  (`title`, `slug`, `description`):

  ```component
  title: <Child Name>
  slug: component-<child-name>
  description: <one-line brief description>
  ```

- **Promoted** (its own blueprint): the declaration links the blueprint
  (`[[component-<name>]]` / `[[module-<name>]]`) and carries NO record
  block.

A child is in exactly ONE state — record block XOR blueprint link, never
both. Promotion lifts the block verbatim into the new document's
frontmatter and replaces it in every parent roster with the link, in the
same change.

The base record fields and the document-envelope fields (status,
updated_at, and the type-specific frontmatter) SHOULD be defined as two
schemas and merged with JSON Schema `allOf` for a standalone document's
frontmatter: the embedded record block satisfies the base-record schema
alone, and a promoted document's frontmatter satisfies both. One schema
therefore validates a child whether it is embedded or promoted.

---

---
title: Frontmatter Standards
description: Required, relationship, and optional frontmatter properties for
  every spec document type, plus the embedded element-record blocks a parent
  document uses to declare un-promoted children
slug: frontmatter-standards
type: standards
status: active
updated_at: 2026-09-29T00:00:00Z
version: "0.3"
---

# Frontmatter Standards

## Contents

- Format — YAML delimiters and placement
- Required Properties — title, description, slug, type, status, updated_at,
  including the full `status` vocabulary and the state-vs-transient rule
- Relationship Properties — provenance, composition, and the typed edges a
  consumer keys on
- Optional Properties — version, created_at, tags, superseded_by
- Embedded Element Records — declaring child components by record
  block, blueprint, or both

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
| `description` | string | Brief summary of the document, written to be triaged from | 1-2 sentences |
| `slug` | string | Unique identifier for cross-references | Kebab-case (lowercase with hyphens) |
| `type` | string | Document classification | Controlled vocabulary defined by document system |
| `status` | string | Lifecycle state | One of: `draft`, `ready-for-review`, `approved`, `in-progress`, `active`, `retired` |
| `updated_at` | datetime | Last modification timestamp | RFC 3339 format: `YYYY-MM-DDTHH:MM:SSZ` |

The `description` property has two audiences, and the second one is why it
is required. A reader who has already opened the document reads it as an
introduction. An agent deciding which documents to open reads it *instead
of* the document: a corpus listing command reports every member's
`description`, so it is the basis on which other agents decide whether to
open this one at all. See [[spec-cli-reference]] for the command that
reports it.

Authors SHALL write `description` to be triaged from rather than as an
introduction. Name what the document decides, governs or specifies, and
what distinguishes it from its neighbours of the same type. "Notes on the
caching layer" tells a triaging agent nothing the title did not; "Why
reads are served from the replica and writes are not, and the staleness
budget that follows" tells it whether to open the document.

The one-to-two-sentence guidance above is a shape, not an enforced limit.
Nothing truncates a description and no check measures one: whether a
description is good enough to choose from is a judgement no check can
make, and a length rule would pass a useless description of the right
size.

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
exist is owned by `[[document-relationship-standards]]` and by the schemas in
`schemas/`; this section states how a template declares one.

WHEN a document has a relationship a body citation cannot carry, the template
SHALL include a relationship property for it in the frontmatter:

```yaml
---
# Provenance (what this is derived from)
sources: []                                 # Source document slugs

# Composition
system: "<system slug>"                     # container -> its one system
modules: []                                 # container or module -> the
                                            # modules composed into it

# Typed edges a machine consumer keys on
contracts_consumed: []                      # contract slugs this consumes
provider: "<element slug>"                  # contract -> its one provider
---
```

WHEN a document does not have relationships, these properties SHALL be omitted 
from the template.

All relationship properties SHALL reference documents by their `slug` value.

A template SHALL NOT introduce a relationship property that only mirrors the
document's own prose citations. Per `[[document-relationship-standards]]`, a
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
membership roster is derivable — from the child's own declaration where a
child may have several parents, and otherwise from the parent's roster
section, which links each promoted child by `[[slug]]` anyway — and a change
set is tracked to completion by the plan, so neither is carried here.

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
container or a module — each child is declared by a record block, by its
own blueprint, or both, and a document system MAY make that declaration
machine-readable:

- **Un-promoted** (no blueprint of its own): the child's declaration
  SHALL carry a fenced record block holding the element record fields
  (`entity_name`, `slug`, `description`, and optionally `technology`):

  ```component
  entity_name: <Child Name>
  slug: component-<child-name>
  description: <one-line brief description>
  technology: <language, framework or package type>
  ```

- **Promoted** (its own blueprint): the declaration links the blueprint
  (`[[component-<name>]]` / `[[module-<name>]]`), and MAY keep its record
  block.

An element is identified by its `slug` and MAY be declared more than
once — by its blueprint and by the record block in its parent's roster,
or by blocks in several documents. Promotion copies the block's fields
into the new blueprint's frontmatter, adds the document's `title`, and
links the blueprint from every parent roster in the same change; the
block MAY stay. Where an element has a blueprint, its values are
authoritative, and a declaration that disagrees with them SHALL be
reported.

Every architecture blueprint (system, container, module, component)
SHALL declare `entity_name`: the element's name as diagrams draw it. It
is distinct from `title`, which names the document — the container named
"Spec CLI" is documented by "Container Blueprint: Spec CLI".

The element record fields and the document-envelope fields (`title`,
status, updated_at, and the type-specific frontmatter) SHOULD be defined
as two schemas and merged with JSON Schema `allOf` for a standalone
document's frontmatter: the embedded record block satisfies the
element-record schema alone, and a promoted document's frontmatter
satisfies both. One schema therefore validates a child whether it is
embedded or promoted.

---

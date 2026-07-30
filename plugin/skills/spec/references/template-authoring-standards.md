---
title: Template Authoring Standards
description: System invariants for authoring document templates —
  structure, metadata, slugs, instructions, EARS, placeholders, and
  cross-references
slug: template-authoring-standards
type: standards
status: active
updated_at: 2026-07-29T00:00:00Z
version: "0.11"
---

# Template Authoring Standards

This document defines invariants that SHALL be applied when authoring templates 
for any type of structured document.

---

## Document Structure

### Line Length

The width of any line SHALL NOT exceed 80 characters.

WHEN a line contains a URL, code block content, or table cell data, the line 
MAY exceed 80 characters IF line breaks would break functionality.

**Rationale:** Ensures readability in terminals, code review tools, and 
side-by-side diff views without horizontal scrolling.

---

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

WHEN a document has relationships to other documents, the template SHALL 
include relationship properties in the frontmatter:

```yaml
---
# Parent/child hierarchies
parent_slug: "<parent document slug>"       # For nested documents
children: []                                # Child document slugs

# Source relationships (what this is based on)
sources: []                                 # Source document slugs

# Composition relationships (what this uses)
composes: []                                # Composed document slugs

# Dependency relationships (ordering)
depends_on: []                              # Prerequisite document slugs
blocks: []                                  # Blocked document slugs

# Change relationships (transient change documents)
updates: []                                 # Existing document slugs this
                                            # change will modify
creates: []                                 # New document slugs authored
                                            # alongside this document

# Peer relationships
related: []                                 # Related document slugs
---
```

WHEN a document does not have relationships, these properties SHALL be omitted 
from the template.

All relationship properties SHALL reference documents by their `slug` value.

WHEN a document describes a change to the document set itself (e.g., a 
feature blueprint that will modify architecture blueprints and author new 
contracts), the template SHALL include the `updates` and `creates` 
properties.

The `updates` property SHALL list existing documents the change will 
modify; the `creates` property SHALL list new documents authored 
alongside the change document (e.g., draft contracts).

**Rationale:** Relationship properties enable traceability, dependency 
management, and impact analysis across documents. Change relationships 
additionally let tooling derive completion criteria for in-flight 
changes: the change is not done until every listed document has been 
updated or activated.

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

## Slug Conventions

### Format

Document slugs SHALL use lowercase kebab-case format (ASCII letters,
digits, and hyphens only).

The slug SHALL be unique across every document in the document system.
The namespace is flat: a slug is a location-independent primary key, not
a path. Uniqueness is enforced mechanically against the frontmatter index
(see below), so a collision is a commit-time error, not a silent
overwrite.

**Format:** `<type-prefix>-<identifier>`

Every document type SHALL have an assigned slug prefix. The prefix is a
property of the type, drawn from the registry the document system
defines — it is not mechanically the `type` value. The prefix names the
subject or the type as that type is habitually cited; the identifier
names the specific document.

### Identity Rule

A slug encodes **identity only**. It SHALL NOT encode any mutable
attribute — status, version, date, or containment/hierarchy. Those live
in frontmatter (`status`, `version`, `parent_slug`, `composes`, …) where
they change without breaking references.

WHEN a document's title or subject is rephrased, its slug SHOULD NOT
change. WHEN a document is genuinely replaced, it is retired with
`superseded_by` rather than renamed. The only sanctioned slug rewrite is
**promotion** (a component promoted to a module), which is a deliberate,
tooling-assisted change that rewrites the slug and every reference to it
in one commit.

### Choosing the Identifier

The identifier SHALL be however that document type is habitually cited,
in the shortest form that stays unambiguous:

- **Named documents** (features, systems, containers, modules,
  components, contracts) use the subject's own distinctive name —
  `component-ledger`, not `component-the-ledger-subsystem`.
- **Numbered documents** (ADRs) use a monotonic number — `adr-011`. The
  number is the stable identity and encodes decision order; the human
  description lives in `title`, not the slug. The next number is derived
  from the frontmatter index (`max + 1`), never tracked by hand.

WHEN two named documents would collide (two containers each with a
`parser` component), disambiguate by making the name more distinctive —
prefer an intrinsic role/domain qualifier (`component-billing-ledger`),
and only where that does not suffice, a parent qualifier
(`component-<container>-<component>`, `container-<system>-<container>`).
A parent qualifier bakes a mutable relationship into the key, so a later
reparent becomes a slug rewrite; the frontmatter index flags every
dangling reference when that happens.

### Prefix Registry

The prefix set is small and closed. Terse prefixes are reserved for
**transient, agent-generated** artifacts (feature blueprints, work
orders), which are read far more than they are hand-authored; durable,
human-cited specs use full-word prefixes. Acronyms are used only for
genuine industry-standard terms (ADR, FRD), never for coined
abbreviations.

| Type | Prefix | Example |
|------|--------|---------|
| ADR | `adr-<nnn>` | `[[adr-011]]` |
| FRD | `frd-` | `[[frd-checkout]]` |
| Standards | *(none — descriptive slug)* | `[[security-standards]]` |
| System Blueprint | `system-` | `[[system-payments]]` |
| Container Blueprint | `container-` | `[[container-ledger-api]]` |
| Module Blueprint | `module-` | `[[module-auth-sdk]]` |
| Component Blueprint | `component-` | `[[component-ledger]]` |
| Contract | `contract-` | `[[contract-payments-api]]` |
| Feature Blueprint | `fb-` | `[[fb-guest-checkout]]` |
| Work Order | `wo-` | `[[wo-1234]]` |

The **Standards** type is the deliberate exception: it carries no fixed
prefix. Its documents are cross-cutting and are cited by a descriptive
name ending in an approved role noun — `-standards`, `-policy`,
`-guidelines`, `-constraints` (e.g. `security-standards`,
`data-retention-policy`). A forced `standards-` prefix would be both
redundant and less descriptive.

`feature-` is reserved for Gherkin `.feature` files and is NOT a spec
slug prefix; feature blueprints use `fb-`.

### Frontmatter Index

A document system SHOULD maintain a machine-readable index of all
documents' frontmatter (a JSON object written to disk). The index is the
authority for slug-uniqueness checks, next-number derivation for
numbered types, `[[slug]]` reference validation, and dangling-reference
detection after a promotion or reparent.

### Construction

WHEN creating a slug from a document title:

1. Convert to lowercase
2. Replace spaces with hyphens
3. Remove characters outside `[a-z0-9-]`
4. Prepend the type's registered prefix

**Example transformations:**
- "Guest Checkout" (feature blueprint) → `fb-guest-checkout`
- "Payments API" (contract) → `contract-payments-api`
- "Data Retention" (standards) → `data-retention-policy`

---

## Template Instructions

### Placement

Templates SHALL include instructions that describe:
- When the template should be used
- When the template should NOT be used
- Guidance for completing each section

Instructions SHALL be provided in two locations:

1. **Document-level instructions:** Immediately after the frontmatter
2. **Section-level instructions:** Within each section

### Format

Instructions SHALL be wrapped in HTML comment blocks (`<!-- -->`).

**Document-level example:**
```markdown
---
title: <Document Title>
description: <Brief summary>
slug: <document-slug>
type: <document-type>
status: draft
updated_at: <RFC 3339 timestamp>
---

<!--
Template: <Template Name>
Source: <Origin or authority, if applicable>

Purpose
  <One paragraph explaining what this template is for>

When to use
  - <Scenario 1>
  - <Scenario 2>
  
When NOT to use
  - <Wrong use case 1>
  - <Wrong use case 2>
-->
```

**Section-level example:**
```markdown
## <Section Title>

<!--
<Instructions for completing this section>

Guidance:
  - <Specific guidance point 1>
  - <Specific guidance point 2>

Example:
  <Brief example of what goes in this section>
-->
```

**Rationale:** HTML comments are passed through markdown processors but not 
rendered by browsers, making them invisible to end users while visible to 
document authors.

### Lifecycle

WHEN a template is used to author a new document, the author SHALL remove 
HTML comment blocks after reading and completing each section.

Instructions SHALL be deleted from the final document to prevent instruction 
bloat.

**Rationale:** Keeps final documents clean while providing guidance during 
authoring.

---

## Template Versioning

Every template SHALL carry a `Version:` line in its document-level
instruction block, immediately after the `Template:` line:

```markdown
<!--
Template: Feature Blueprint
Version: 1.3
...
-->
```

WHEN a template's normative content changes (a rule, a required
section, a check, reference syntax), the MINOR version SHALL be
incremented. WHEN only editorial content changes (typos, phrasing,
examples), the PATCH version MAY be incremented instead. Version
numbers use `MAJOR.MINOR` with an optional `.PATCH`.

The same rule applies to the skills (`version` in SKILL.md
frontmatter) and to versioned standards documents (`version` in
frontmatter): any normative change bumps the version in the same
commit.

**Rationale:** Improvement RFCs (see the `rfc` template) cite the
versions of the templates, skills, and standards a session followed.
Triage uses those citations to answer "is this proposal already
fixed?" mechanically; without version stamps that judgment requires
diff archaeology. The version lives in the instruction block — not
the template's frontmatter — because template frontmatter is the
placeholder frontmatter of generated documents, and instances do not
carry their template's version.

---

## Requirements Expression

### EARS Syntax

WHEN a template is used to express requirements or invariants, EARS (Easy 
Approach to Requirements Syntax) SHALL be adopted.

EARS patterns SHALL use capitalized keywords: WHEN, WHILE, IF, THEN, SHALL, 
SHALL NOT.

**EARS Patterns:**

| Pattern | Syntax | Use Case |
|---------|--------|----------|
| Ubiquitous | `The system SHALL [action]` | Always-active requirements |
| Event-driven | `WHEN [event], the system SHALL [action]` | Triggered by events |
| State-driven | `WHILE [state], the system SHALL [action]` | Active during a state |
| Conditional | `IF [condition], THEN the system SHALL [action]` | Conditional requirements |
| Negative | `The system SHALL NOT [action]` | Prohibitions |

**Example:**
```markdown
The system SHALL NOT transmit data to external services.

WHEN a user submits a form, the system SHALL validate all required fields.

WHILE processing is in progress, the system SHALL display a loading indicator.

IF validation fails, THEN the system SHALL display an error message.
```

**Rationale:** EARS provides unambiguous, testable constraint definitions 
with explicit triggers and conditions. Developed by Rolls-Royce for 
safety-critical systems.

### Modal Verbs (RFC 2119)

Requirements SHALL use RFC 2119 keywords to indicate requirement levels:

| Keyword | Meaning | Usage |
|---------|---------|-------|
| **SHALL** / **MUST** | Absolute requirement | Mandatory behavior |
| **SHALL NOT** / **MUST NOT** | Absolute prohibition | Forbidden behavior |
| **SHOULD** / **RECOMMENDED** | Strong recommendation | Recommended unless good reason to deviate |
| **SHOULD NOT** / **NOT RECOMMENDED** | Strong discouragement | Avoid unless specific justification |
| **MAY** / **OPTIONAL** | Truly optional | Implementer's choice |

WHEN expressing requirements, modal verbs SHALL be capitalized for clarity.

Modal verbs SHALL be used with care and sparingly, only where behavior must 
be enforced for interoperability or to prevent harm.

**Rationale:** RFC 2119 provides standardized language for requirement 
levels, widely recognized in technical specifications since 1997.

---

## Placeholders and Parameters

### Format

Placeholders in templates SHALL be wrapped in angle brackets: `<placeholder>`

Placeholders SHALL contain a human-readable description of the expected 
content.

**Examples:**
- `<Feature Name>`
- `<one-line summary>`
- `<PREFIX>` (for ID generation)
- `YYYY-MM-DD` (date format indicator)

**Rationale:** Angle bracket notation originates from Backus-Naur Form (1950s) 
and is universally recognized in technical documentation to distinguish 
literal text from placeholders.

### Variable Naming

Placeholders SHALL use one of these naming conventions:

1. **Human-readable description:** `<Feature Name>`, `<one-line summary>`
2. **Format indicator:** `YYYY-MM-DD`, `NNN` (zero-padded number)
3. **Variable name:** `<PREFIX>`, `<slug>`, `<TYPE>`

Placeholder names SHALL be self-explanatory without requiring external 
documentation.

### Environment Variables vs Placeholders

Template placeholders are distinct from environment variables injected at 
runtime.

**Template placeholder (manual replacement):**
```yaml
id: "FRD-<PREFIX>"
title: "<Feature Name>"
```

**Environment variable (automatic injection):**
```yaml
last_updated: ${CURRENT_DATE}
created_by: ${USER_EMAIL}
```

Environment variables SHALL use `${VARIABLE_NAME}` syntax when supported by 
the document generation tool.

**Note:** Environment variable injection is tool-specific and should be 
documented in the relevant tooling guide, not in individual templates.

---

## Cross-Document References

### Document Reference Syntax

WHEN one document references another document in the specification set,
double-bracket wikilink notation SHALL be used.

**Format:** `[[document-slug]]`

**Examples:**
- `[[security-standards]]`
- `[[contract-payments-api]]`
- `[[template-authoring-standards]]`

A reference resolves against the target document's frontmatter `slug` —
never a file path. The namespace is flat and location-independent, so a
reference is valid wherever the referencing document lives and wherever
the target moves. At-sign (`@`) notation SHALL NOT be used: in a coding
agent's environment `@` denotes a filesystem mention resolved against the
session root, which a slug is not.

### Reference Usage Rules

WHEN referencing a document, the full slug SHALL be used (not the title),
and only the slug SHALL appear between the brackets — display aliases
(`[[slug|text]]`) SHALL NOT be used, as they reintroduce the title drift
the slug exists to prevent. For readable prose, place words outside the
link: `the checkout FRD ([[frd-checkout]])`.

**Correct:** `See [[security-standards]] for requirements.`  
**Incorrect:** `See Security Standards for requirements.`

**Rationale:** An unambiguous closing delimiter lets tooling extract and
validate every reference against the frontmatter index and build
dependency graphs, with no trailing-punctuation ambiguity.

### Specialized Reference Syntax

Document systems MAY define additional reference syntaxes for
domain-specific elements. `[[ ]]` is reserved for spec-set *documents*;
the syntaxes below name things that are NOT documents.

WHEN specialized reference syntax is defined, it SHALL be documented in the 
template's instruction block.

**Example specialized syntaxes:**
- `` `Identifier` `` for a code-level element that is not a modeled
  component — a schema, domain type, event, flag, or table
- `$VARIABLE_NAME` for environment variables
- `{placeholder}` for template variables

An un-promoted component — declared under a `### <Name>` heading in its
parent container's or module's Components section, with no blueprint of
its own — is referenced with a standard wikilink heading anchor into
that parent: `[[container-<name>#Component Name]]`. Once promoted to its
own blueprint it is referenced as a document, `[[component-<name>]]`;
the parent's roster heading persists through promotion, so existing
anchor references keep resolving. Backticks stay reserved for code-level
elements that are not modeled components.

WHEN using specialized syntax, templates SHALL include examples in the 
instruction block.

### References Outside the Document Set

Documents that are not part of the specification set — contributor and
build docs (`AGENTS.md`, `CONTRIBUTING.md`), pre-existing design or
pattern documents, and the framework's own reference guides shipped with
a skill (`references/c4-guide.md`, `references/ears-guide.md`) — MAY be
referenced by path in backticks.

**Examples:** `` `AGENTS.md` ``, `` `references/c4-guide.md` ``

The `[[document-slug]]` notation SHALL be reserved for documents in the 
specification set; it SHALL NOT be used for documents outside it.

**Rationale:** Real repositories have authoritative documents that 
predate or sit outside the framework, and framework guides ship with the
skill rather than being copied into a user's spec set. Referencing them
by path avoids both duplication (which drifts) and forced migration; the
distinct syntax keeps `[[slug]]` reference validation sound.

---

## Template Content Guidelines

### Avoid Duplication

Templates SHALL instruct authors to cross-reference other documents rather 
than duplicate content.

WHEN content exists in another document, templates SHALL guide authors to 
reference that document using `[[document-slug]]` notation.

**Rationale:** Duplication leads to drift when only one copy is updated. 
Cross-referencing maintains a single source of truth.

### Avoid Volatile Facts

Documents SHOULD NOT state volatile facts — element counts, line 
numbers, dependency version numbers, file sizes — unless the fact is 
itself normative. Name the thing and point at its source of truth 
instead ("the `ValidationError` enum" and where it lives, not "the 
~50-variant enum").

Exception: versions that are contractual (a minimum supported runtime, 
a pinned toolchain) are requirements, not descriptions, and SHALL be 
expressed in EARS.

**Rationale:** Volatile precision reads as rigor but rots immediately: 
the count changes and the document silently becomes wrong. The 
underlying source of truth stays correct.

### Living Documents

WHEN templates are used for documentation that evolves over time, the template 
SHALL include instructions that documents are living artifacts.

**Example instruction:**
```markdown
<!--
This is a living document. Update when:
- Requirements change
- New information becomes available
- Errors are discovered
- Context shifts
-->
```

---

## Validation

### Frontmatter Validation

WHEN a document system implements validation tooling, JSON Schema definitions 
SHALL be created for each document type to validate:
- Required properties are present
- Property values match allowed formats
- Slug conventions are followed
- Status values are valid
- Embedded record blocks validate against the base-record schema, and
  every child element is in exactly one state (record block XOR
  blueprint link)

Schema validation SHOULD be run as a pre-commit hook to catch errors before 
commit.

### Reference Validation

WHEN a document system implements reference validation, tools SHOULD validate 
that all `[[document-slug]]` references point to existing documents.

### Dependency Graph

WHEN documents include relationship properties (`depends_on`, `blocks`, 
`sources`, `related`), tools SHOULD generate a dependency graph to:
- Visualize document relationships
- Detect circular dependencies
- Enable impact analysis (what's affected if this changes?)
- Find orphaned documents (no incoming references)

---

## References

- [RFC 2119: Key words for RFCs](https://www.ietf.org/rfc/rfc2119.txt) — Modal verb definitions
- [RFC 3339: Date and Time on the Internet](https://www.ietf.org/rfc/rfc3339.txt) — Timestamp format
- [EARS Wikipedia](https://en.wikipedia.org/wiki/Easy_Approach_to_Requirements_Syntax) — EARS background
- [Jekyll Frontmatter](https://jekyllrb.com/docs/front-matter/) — YAML frontmatter origin
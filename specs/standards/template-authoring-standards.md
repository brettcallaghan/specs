---
title: Template Authoring Standards
description: System invariants for authoring document templates —
  structure, slugs, instruction blocks, versioning, placeholders, content
  guidelines, and validation. Frontmatter, EARS, and cross-document
  reference syntax are defined in the companion documents it points to
slug: template-authoring-standards
type: standards
status: active
---

# Template Authoring Standards

This document defines invariants that SHALL be applied when authoring templates 
for any type of structured document.

## Contents

- Document Structure — line length, title and H1, section outline
- Metadata (Frontmatter) → `frontmatter-standards.md`
- Slug Conventions — format, identity rule, choosing the identifier, the
  prefix registry, the frontmatter index, construction
- Template Instructions — placement, HTML comment format, lifecycle
- Versioning — what identifies a template, skill, or standard
- Requirements Expression → `ears-guide.md`
- Placeholders and Parameters — format, naming, env vars vs placeholders
- Cross-Document References → `cross-document-references.md`
- Template Content Guidelines — avoid duplication, avoid volatile facts,
  living documents
- Validation — frontmatter schemas, reference validation, dependency graph
- References

---

## Document Structure

### Line Length

The width of any line SHALL NOT exceed 80 characters.

WHEN a line contains a URL, code block content, or table cell data, the line 
MAY exceed 80 characters IF line breaks would break functionality.

**Rationale:** Ensures readability in terminals, code review tools, and 
side-by-side diff views without horizontal scrolling.

### Title and H1

A document's H1 SHALL repeat its `title`. Neither SHALL carry a type
label (`FRD:`, `Container Blueprint:`): `type` already says what the
document is. The ADR is the one exception — its title and H1 lead with
its number, `ADR-011: <Decision Title>`, because the number is how an
ADR is cited.

### Section Outline

Templates that share a section SHALL give it the same name, the same
position and the same instruction block, with only the type noun
substituted ("this container", "this module"). The shared outline:

| Section | Present in | Position |
|---------|-----------|----------|
| `## Overview` | every type except ADR (which opens with `## Context`) and standards (an introductory paragraph under the H1) | first |
| *type-specific sections* | per template | after Overview |
| `## Cross-Cutting Standards` | system, container, module, component blueprints | after the type-specific sections |
| `## Change Guidance` | the same four | after Cross-Cutting Standards |
| `## Key Decisions` | the four blueprints, contract, feature blueprint | after Change Guidance, or after the type-specific sections where there is none |
| `## Open Questions` | the four blueprints, contract, feature blueprint, FRD | second to last |
| `## References` | every type | last |

Every `##` section SHALL carry a section-level instruction block. An
`###` heading that is a worked example of its parent's list (a roster
entry, an example requirement, an example delta) is covered by the
parent's block.

`## Open Questions` is **optional**: it exists while a document is a
draft and is removed when the document leaves draft. While present, it
MAY state the literal text "None." instead of a list. A tool that checks
required sections SHALL NOT require it. Every other section in the
outline above is required for the types it lists, except where a
template marks a section conditional on a frontmatter value (Pattern
Guidance for `kind: pattern`, Versioning and Compatibility for
`visibility: public`).

Body text SHALL NOT restate a frontmatter property's value unless the
restatement helps a human reader — a link to the provider's blueprint
does; repeating `visibility` or `source_paths` does not. Where both
exist, frontmatter is authoritative.

**Rationale:** A required-sections check reads one list per type from
configuration; this outline is where that list comes from, so it is
written down once rather than inferred from each template.

---

## Metadata (Frontmatter)

Defined in `frontmatter-standards.md` — required, relationship, and optional
properties, the `status` vocabulary, and the embedded element-record blocks.
Load it when authoring a template's frontmatter or validating a document's.

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
in frontmatter (`status`, `version`, `system`, `modules`, …) where
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
| Feature Blueprint | `feature-` | `[[feature-guest-checkout]]` |
| Work Order | `wo-` | `[[wo-1234]]` |
| Guide | *(none — descriptive slug)* | `[[spec-cli-reference]]` |

The **Standards** type is the deliberate exception: it carries no fixed
prefix. Its documents are cross-cutting and are cited by a descriptive
name ending in an approved role noun — `-standards`, `-policy`,
`-guidelines`, `-constraints` (e.g. `security-standards`,
`data-retention-policy`). A forced `standards-` prefix would be both
redundant and less descriptive.

**Guides** carry no template and no prefix: a guide is free-form and
follows only the rules every document follows (frontmatter, title and
H1, line length, References last).

### Filenames

A document's filename SHALL end in its document type, as a suffix, with
the ADR the single exception: its number leads, because the number is how
an ADR is cited and sorted.

| Type | Filename |
|------|----------|
| ADR | `adr-<nnn>-<name>.md` |
| FRD | `<name>-frd.md` |
| Feature Blueprint | `<name>-blueprint.md` |
| System Blueprint | `<name>-system.md` |
| Container Blueprint | `<name>-container.md` |
| Module Blueprint | `<name>-module.md` |
| Component Blueprint | `<name>-component.md` |
| Standards | `<name>-standards.md` |
| Contract | `<name>-contract.md` |
| Guide | `<name>-guide.md` |

The `<name>` is the document's own name — its slug with the registered
type prefix removed, since the suffix now carries what the prefix did.

A suffix SHALL NOT be doubled, and SHALL NOT be stacked on a name that
already ends in a role noun serving the same purpose (`-checklist`,
`-protocol`, `-reference`, `-guide`, `-definition-of-ready`). A standards
document named `blueprint-readiness-checklist` is already announced as
what it is; `blueprint-readiness-checklist-standards.md` says it twice
and reads worse for it.

**Rationale:** the suffix sorts a directory by subject rather than by
type, which is how a reader looks for a document — every FRD grouped
together is useless when the directory holds only FRDs, while `checkout`
next to `checkout-frd.md` and `checkout-blueprint.md` is not. The ADR
inverts this because its identity is its number: `[[adr-011]]` cites a
number, so a directory sorted by number is the one a reader can navigate.

Renaming a file is cheap: no `[[slug]]` reference addresses a path, and
the index is built by walking the spec root rather than by matching
names. Renaming a *slug* is not — it breaks every inbound reference — so
WHEN a filename and its slug disagree, prefer correcting the filename.

The one cost to weigh first: a standard or guide vendored into a skill
(`references/shared/`) takes its copy's basename from its source, so
renaming the source renames every copy and breaks each backticked
citation of it.

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
- "Guest Checkout" (feature blueprint) → `feature-guest-checkout`
- "Payments API" (contract) → `contract-payments-api`
- "Data Retention" (standards) → `data-retention-policy`

---

## Template Instructions

### Placement

Templates SHALL include instructions that describe:
- When the template should be used
- When the template should NOT be used
- Who reads the document, as a **Reader profile** in the document-level
  instructions: a list of Reader, Knows, Wants, Budget, and Tone, from
  which the `tech-writer` skill sets the document's tone
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
---

<!--
Template: <Template Name>
Source: <Origin or authority, if applicable>

Purpose
  <One paragraph explaining what this template is for>

Reader profile (apply with the tech-writer skill)
- Reader: <who reads it, and when>
- Knows: <what they already know, and what they don't>
- Wants: <the one thing they came to do or decide>
- Budget: <how long they'll give the document>
- Tone: <how the voice adjusts for this reader>

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

## Versioning

Templates SHALL NOT carry a version of their own. They inherit the
release version of the plugin that ships them (`version` in the plugin's
`plugin.json`).

**Rationale:** Issues raised by a retro (see the issue template) cite
the release a session ran, and triage compares it with the current one
to answer "is this issue already fixed?". A hand-maintained number on
every template drifts the moment someone forgets to bump it.

---

## Requirements Expression

WHEN a template is used to express requirements or invariants, EARS (Easy
Approach to Requirements Syntax) SHALL be adopted, with capitalized RFC 2119
modal verbs.

The patterns, the modal-verb rubric, style rules, and worked examples are
defined in `ears-guide.md` — the single, more complete definition. Load it
when writing any requirement; this document does not restate it.

**Rationale:** EARS provides unambiguous, testable constraint definitions
with explicit triggers and conditions. Developed by Rolls-Royce for
safety-critical systems.

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

Defined in `cross-document-references.md` — the `[[slug]]` wikilink syntax,
usage rules, component heading anchors, and the rules for referencing
documents outside the specification set. Load it when authoring or
validating references; do not restate it here.

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
- Embedded record blocks validate against the element-record schema, and
  every declaration of one element agrees with the others

Schema validation SHOULD be run as a pre-commit hook to catch errors before 
commit.

### Reference Validation

WHEN a document system implements reference validation, tools SHOULD validate 
that all `[[document-slug]]` references point to existing documents.

### Dependency Graph

WHEN documents include relationship properties (`sources`, `system`,
`modules`), tools SHOULD generate a dependency graph alongside the
graph derived from body citations to:
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
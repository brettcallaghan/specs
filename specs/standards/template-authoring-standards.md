---
title: Template Authoring Standards
description: The rules for writing the spec skill's templates — the shared
  section outline, instruction blocks, placeholders, and versioning
slug: template-authoring-standards
type: standards
status: active
---

# Template authoring standards

These rules govern writing the `spec` skill's templates. Every document
authored from a template also follows [[spec-authoring-standards]], and a
template's instructions SHALL stay consistent with it.

---

## Document structure

### Section outline

Templates that share a section SHALL give it the same name, the same
position and the same instruction block, with only the type noun
substituted ("this container", "this module"). The shared outline:

| Section | Present in | Position |
|---------|-----------|----------|
| `## Overview` | every type except ADR (which opens with `## Context`), Standards (an introductory paragraph under the H1), and Component and Module Blueprints (which open with `## Responsibility`) | first |
| *type-specific sections* | per template | after the first section |
| `## Cross-cutting Standards` | System, Container, Module, Component Blueprints | after the type-specific sections |
| `## Change guidance` | the same four | after Cross-cutting Standards |
| `## Key decisions` | the four blueprints, Contract, Feature Blueprint | after Change guidance, or after the type-specific sections where there is none |
| `## Open questions` | the four blueprints, Contract, Feature Blueprint, FRD | second to last |
| `## References` | every type | last |

Every `##` section SHALL carry a section-level instruction block. An
`###` heading that is a worked example of its parent's list (a roster
entry, an example requirement, an example delta) is covered by the
parent's block.

`## Open questions` is optional, and a template marks any section that's
conditional on a frontmatter value. [[spec-authoring-standards]] states
both rules for authors.

**Rationale:** A required-sections check reads one list per type from
configuration. This outline is where that list comes from, so it's written
down once rather than inferred from each template.

---

## Template instructions

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

**Rationale:** Markdown processors pass HTML comments through, and
browsers don't render them. So they're invisible to end users and visible
to document authors.

### Lifecycle

WHEN a template is used to author a new document, the author SHALL remove 
HTML comment blocks after reading and completing each section.

Instructions SHALL be deleted from the final document to prevent instruction 
bloat.

**Rationale:** Final documents stay clean, and authors still get guidance
while they write.

---

## Versioning

Templates SHALL NOT carry a version of their own. They inherit the
release version of the plugin that ships them (`version` in the plugin's
`plugin.json`).

**Rationale:** Issues raised by a retro (see the issue template) cite
the release a session ran. Triage compares it with the current release to
answer "is this issue already fixed?" A hand-maintained number on every
template drifts the moment someone forgets to bump it.

---

## Placeholders and parameters

### Format

Placeholders in templates SHALL be wrapped in angle brackets: `<placeholder>`

Placeholders SHALL contain a human-readable description of the expected 
content.

**Examples:**
- `<Feature Name>`
- `<one-line summary>`
- `<PREFIX>` (for ID generation)
- `YYYY-MM-DD` (date format indicator)

**Rationale:** Angle-bracket notation comes from Backus-Naur Form
(1950s). Technical documentation universally recognizes it as a way to tell
placeholders apart from literal text.

### Variable naming

Placeholders SHALL use one of these naming conventions:

1. **Human-readable description:** `<Feature Name>`, `<one-line summary>`
2. **Format indicator:** `YYYY-MM-DD`, `NNN` (zero-padded number)
3. **Variable name:** `<PREFIX>`, `<slug>`, `<TYPE>`

Placeholder names SHALL be self-explanatory without requiring external 
documentation.

### Environment variables vs placeholders

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

**Note:** Environment variable injection is tool-specific. It should be
documented in the relevant tooling guide, not in individual templates.

---

## Template content guidelines

### Avoid duplication

Templates SHALL instruct authors to cross-reference other documents rather 
than duplicate content.

WHEN content exists in another document, templates SHALL guide authors to 
reference that document using `[[document-slug]]` notation.

**Rationale:** When only one copy of duplicated content is updated, the
copies drift. A cross-reference keeps a single source of truth.

### Living documents

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

## References

- [RFC 2119: Key words for RFCs](https://www.ietf.org/rfc/rfc2119.txt) — Modal verb definitions
- [[spec-authoring-standards]] — the rules every document follows
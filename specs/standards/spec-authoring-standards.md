---
title: Spec Authoring Standards
description: The rules every specification document follows, whatever its
  type — line length, title and H1, slugs and the prefix registry, and
  filenames
slug: spec-authoring-standards
type: standards
status: active
---

# Spec authoring standards

Every specification document follows these rules, whatever its type. Each
template adds the rules for its own sections.

## Document structure

### Line length

The width of any line SHALL NOT exceed 80 characters.

WHEN a line contains a URL, code block content, or table cell data, the line 
MAY exceed 80 characters IF line breaks would break functionality.

**Rationale:** Lines stay readable in terminals, code review tools, and
side-by-side diff views without horizontal scrolling.

### Title and H1

A document's H1 SHALL repeat the words of its `title` in sentence case,
so the two may differ in case. Neither SHALL carry a type
label (`FRD:`, `Container Blueprint:`): `type` already says what the
document is. The ADR is the one exception: its title and H1 lead with its
number (`ADR-011: <Decision Title>`), because the number is how an ADR is
cited.

### Open questions and frontmatter

`## Open questions` is **optional**: it exists while a document is a
draft and is removed when the document leaves draft. While present, it
MAY state the literal text "None." instead of a list. A tool that checks
required sections SHALL NOT require it. Every other section a template
defines is required, except where the template marks it conditional on a
frontmatter value (Pattern guidance for `kind: pattern`, Versioning and
compatibility for `visibility: public`).

Body text SHALL NOT restate a frontmatter property's value unless the
restatement helps a human reader — a link to the provider's blueprint
does; repeating `visibility` or `source_paths` does not. Where both
exist, frontmatter is authoritative.

### Template instructions

WHEN you author a document from a template, remove each HTML comment block
once you've completed its section. A finished document carries no
instruction blocks.

**Rationale:** Final documents stay clean, and authors still get guidance
while they write.

---

## Slug conventions

### Format

Document slugs SHALL use lowercase kebab-case format (ASCII letters,
digits, and hyphens only).

The slug SHALL be unique across every document in the document system.
The namespace is flat: a slug is a location-independent primary key, not
a path. `specs slug-exists` answers whether a slug is taken, and
`specs check` reports a collision, so it's an error, not a silent
overwrite.

**Format:** `<type-prefix>-<identifier>`

Every document type SHALL have an assigned slug prefix. The prefix is a
property of the type, drawn from the registry the document system
defines. It isn't mechanically the `type` value. The prefix names the
subject or the type as that type is habitually cited. The identifier names
the specific document.

### Identity rule

A slug encodes **identity only**. It SHALL NOT encode any mutable
attribute — status, version, date, or containment/hierarchy. Those live
in frontmatter (`status`, `version`, `system`, `modules`, …), where they
change without breaking references.

WHEN a document's title or subject is rephrased, its slug SHOULD NOT
change. When a document is genuinely replaced, it's retired with
`superseded_by` rather than renamed. The only sanctioned slug rewrite is
**promotion** (a component promoted to a module). Promotion is a
deliberate, tooling-assisted change that rewrites the slug and every
reference to it in one commit.

### Choosing the identifier

The identifier SHALL be however that document type is habitually cited,
in the shortest form that stays unambiguous:

- **Named documents** (features, systems, containers, modules,
  components, contracts) use the subject's own distinctive name:
  `component-ledger`, not `component-the-ledger-subsystem`.
- **Numbered documents** (ADRs) use a monotonic number, such as `adr-011`.
  The number is the stable identity and encodes decision order. The human
  description lives in `title`, not the slug. The next number comes from
  `specs next-id`, never from reading filenames.

When two named documents would collide (two containers each with a
`parser` component), disambiguate by making the name more distinctive.
Prefer an intrinsic role or domain qualifier (`component-billing-ledger`).
Only where that doesn't suffice, use a parent qualifier
(`component-<container>-<component>`, `container-<system>-<container>`).
A parent qualifier bakes a mutable relationship into the key, so a later
reparent becomes a slug rewrite. `specs check` reports every dangling
reference when that happens.

### Prefix registry

The prefix set is small and closed. Terse prefixes are reserved for
**transient, agent-generated** artifacts (Feature Blueprints, Work
Orders), which are read far more than they're hand-authored. Durable,
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
| Guide | *(none — descriptive slug)* | `[[onboarding-guide]]` |

The **Standards** type is the deliberate exception: it carries no fixed
prefix. Its documents are cross-cutting and are cited by a descriptive
name ending in an approved role noun: `-standards`, `-policy`,
`-guidelines`, or `-constraints` (for example, `security-standards` or
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

The `<name>` is the document's own name: its slug with the registered
type prefix removed, because the suffix carries what the prefix did.

A suffix SHALL NOT be doubled, and SHALL NOT be stacked on a name that
already ends in a role noun serving the same purpose (`-checklist`,
`-protocol`, `-reference`, `-guide`, `-definition-of-ready`). A Standards
document named `work-order-definition-of-ready` already announces what it
is. `work-order-definition-of-ready-standards.md` says it twice and reads
worse for it.

**Rationale:** The suffix sorts a directory by subject rather than by
type, which is how a reader looks for a document. Every FRD grouped
together is useless when the directory holds only FRDs, while `checkout`
next to `checkout-frd.md` and `checkout-blueprint.md` isn't. The ADR
inverts this because its identity is its number. `[[adr-011]]` cites a
number, so a directory sorted by number is the one a reader can navigate.

Renaming a file is cheap: no `[[slug]]` reference addresses a path, and
the index is built by walking the spec root rather than by matching
names. Renaming a *slug* isn't cheap, because it breaks every inbound
reference. So when a filename and its slug disagree, prefer correcting the
filename.

There's one cost to weigh first: a standard or guide vendored into a skill
(`references/shared/`) takes its copy's basename from its source, so
renaming the source renames every copy and breaks each backticked citation
of it.

### Construction

When creating a slug from a document title:

1. Convert to lowercase
2. Replace spaces with hyphens
3. Remove characters outside `[a-z0-9-]`
4. Prepend the type's registered prefix

**Example transformations:**
- "Guest Checkout" (Feature Blueprint) → `feature-guest-checkout`
- "Payments API" (Contract) → `contract-payments-api`
- "Data Retention" (Standards) → `data-retention-policy`

---

## References

- [RFC 2119: Key words for RFCs](https://www.ietf.org/rfc/rfc2119.txt) — modal verb definitions

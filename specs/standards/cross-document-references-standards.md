---
title: Cross-Document Reference Standards
description: How documents reference each other and the elements they
  declare — wikilinks for spec documents, heading anchors for un-promoted
  components, backticked paths for everything outside the set
slug: cross-document-references
type: standards
status: active
updated_at: 2026-08-09T00:00:00Z
version: "0.3"
---

# Cross-Document Reference Standards

## Contents

- Document Reference Syntax — `[[slug]]`, and why `@` is not used
- Reference Usage Rules — full slugs, no display aliases
- Specialized Reference Syntax — backticked code elements, component
  heading anchors into a parent's roster
- References Outside the Document Set — backticked paths, and the rule
  that a path is written relative to the file containing it

The decision behind these conventions is recorded in [[adr-012]].
Companion to `template-authoring-standards.md`.

These conventions are mechanically checked. `spec check` resolves every
`[[slug]]` in a document's body and every slug-typed frontmatter
property against the corpus, and reports the ones that resolve to
nothing; `spec refs <slug>` answers who cites a document before you
change it. `spec-cli.md` states what each rule reports and the three
reference cases that are still yours. The rules below are what those
checks enforce — which is also why the exceptions are worth knowing:
a `[[slug]]` inside backticks or a fenced block is an illustration and
is deliberately not treated as a reference, so this document can name
the notation without citing documents that do not exist.

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
a skill (`c4-guide.md`, `ears-guide.md`) — MAY be referenced by path in
backticks.

#### Paths are relative to the file that contains them

A backticked path SHALL be written relative to the **file it appears
in**, not to the skill root. The same guide therefore has two correct
spellings depending on where it is cited from:

| Cited from | Path to the C4 guide |
|------------|----------------------|
| a skill's `SKILL.md` | `` `references/c4-guide.md` `` |
| another file inside `references/` | `` `c4-guide.md` `` |

**Rationale:** an agent resolves a path from the file it is currently
reading. A `references/`-prefixed path written *inside* `references/`
resolves to `references/references/…` — that is, to nothing — so the
reference silently fails to load or costs a search to recover. Writing
the path from the reader's position is the only spelling that works from
every location the file is read.

**Examples:** `` `AGENTS.md` ``, `` `references/c4-guide.md` `` (from a
SKILL.md), `` `c4-guide.md` `` (from a sibling reference file)

The `[[document-slug]]` notation SHALL be reserved for documents in the 
specification set; it SHALL NOT be used for documents outside it.

**Rationale:** Real repositories have authoritative documents that 
predate or sit outside the framework, and framework guides ship with the
skill rather than being copied into a user's spec set. Referencing them
by path avoids both duplication (which drifts) and forced migration; the
distinct syntax keeps `[[slug]]` reference validation sound.

---

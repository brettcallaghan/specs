---
title: Document Relationship Standards
description: How formal relationships between specification documents are
  declared — declare once and let the index derive the reverse, when a
  frontmatter relationship property earns its maintenance cost, which end
  of a relationship retains the edge, and when a paired declaration is
  justified
slug: document-relationship-standards
type: standards
status: active
updated_at: 2026-10-07T00:00:00Z
version: "0.3"
---

# Document Relationship Standards

These requirements govern how one specification document records a
relationship to another — a peer it observes, a source it derives from, a
child it composes, a document it supersedes. They apply to anyone
authoring or reviewing a specification document, and to anyone adding,
removing, or configuring a reference-bearing frontmatter property.

Their authority is a capability the framework now has and did not when the
vocabulary was designed: the reference-integrity feature parses every
`[[slug]]` link from a document's body and records the graph in the index
**in both directions**, deriving each document's inbound edges from every
other document's outbound ones ([[frd-spec-reference-integrity]],
AC-SPECREF-001.1 and AC-SPECREF-001.3), and exposes both directions to a
query (AC-SPECREF-003). Once the reverse of every relationship is computed
for free, a second copy of it — mirrored into a counterpart document's
frontmatter — is not redundancy the tool needs. It is drift surface an
author maintains by hand.

---

## Declaring a Relationship

### Declare Once, Derive the Reverse

A cross-document relationship SHALL be declared in exactly one place — one
document's prose or one document's frontmatter property — and SHALL NOT be
mirrored in the counterpart document.

**Rationale:** The index derives the reverse edge of anything declared
(AC-SPECREF-001.3), so a mirrored declaration adds no information the tool
cannot compute. What it adds is a second site that must be edited in
lockstep and a way for the two to disagree. A relationship recorded in one
authoritative place cannot drift against itself.

---

WHEN the reverse direction of a relationship is needed, a consumer SHALL
obtain it from the index's inbound edges rather than from a mirrored
frontmatter property.

**Rationale:** `specs refs <slug>` answers "what points at this" from the
same graph that answers "what does this point at"
([[frd-spec-reference-integrity]], AC-SPECREF-003.1 and AC-SPECREF-003.2).
A document does not have to carry its own back-references for the reverse
to be answerable.

---

## When a Frontmatter Relationship Property Is Justified

A frontmatter property that names other documents SHALL be introduced or
retained only if it satisfies at least one of the following, none of which
body-derivation can supply:

- it carries a **typed** relationship a prose mention cannot convey, and a
  machine consumer keys on that type;
- it names a target **not reliably present** in the document's prose; or
- it is the **only home** for an edge that has nowhere else to live.

**Rationale:** A prose `[[slug]]` is untyped — the index records that
document A mentions document B, not that B is A's *successor* or *source*.
A property clears the first bar when a check reads its meaning
(`superseded_by` drives the retired-target checks). It clears the second
when the relationship is not something an author would write in the body
at all. It clears the third when no other document and no derivation can
hold the edge. No property in this framework's own vocabulary currently
clears only the third test — the bar exists for an adopting profile whose
corpus has an edge with nowhere else to live, and a property offered
against it is expected to name that edge rather than assume one. A
property that clears none of these is doing work the index already does.

---

A frontmatter property that only mirrors a document's `## References`
prose links SHALL NOT be introduced, and an existing one SHOULD be removed.

**Rationale:** If every value of the property already appears in the body
as a `[[slug]]` — which the index now records outbound and inbound — the
property is a hand-maintained copy of a derived fact. It cannot be more
correct than the graph and can be less.

---

## Choosing Which End Retains the Edge

### Modify What Changed

WHEN a relationship is declared once and either end could hold it, the
author SHALL declare it on the end whose change creates the relationship
— the document being written or edited at the moment the relationship
comes into being — so that recording it never requires editing the other
end. The other end, which changes less often, SHALL NOT be required to
restate it; the index derives it.

**Rationale:** The end that changes when the relationship is created is
the end that will be open in the editor. Putting the edge anywhere else
turns one authoring act into two, and the second is the one that gets
forgotten. It also keeps a stable document stable: a shared module is
not edited every time another container adopts it.

The relationships this framework declares, and why each sits where it
does:

| Relationship | Declared on | Because creating it changes |
|---|---|---|
| a container belongs to a system | the container, `system` | the container, which joins the system |
| a container or module composes a module | the host, `modules` | the host, which adopts the module; the module may already serve other hosts |
| a container or module composes a component | the host, in its roster section | the host, whose roster gains the component |
| an element depends on a contract | the element, `contracts_consumed` | the element, which starts depending on it |
| an element provides a contract | the contract, `provider` | the contract, which is written for its provider |
| a feature blueprint realizes an FRD | the blueprint, `frd` | the blueprint, written after the FRD |
| a document is derived from sources | the document, `sources` | the document, written from them |
| a document is replaced by a successor | the retired document, `superseded_by` | the retired document, whose status changes to `retired` in the same edit |

---

### Nothing Durable Points at a Feature Blueprint

A document the profile registers at a live status SHALL NOT declare a
relationship to a feature blueprint, in prose or in frontmatter. The
feature blueprint declares its own edges (to its FRD, to the blueprints
it folds into, to the decisions it applies), and the reverse direction is
answered by `specs refs <slug>`.

**Rationale:** a feature blueprint retires when its Definition of Done is
complete, not because something replaced it. An edge declared *into* one
becomes a `reference/retired-target` warning on the day the feature
ships, and fails every `--strict` run from then on. The durable end of
the relationship is the ADR, the FRD or the architecture blueprint; those
stay live, and the graph already carries the direction that matters.

---

### Singularity

WHEN both ends change in the same act, the author SHALL declare the
relationship on the single-valued end and derive the list.

**Rationale:** A scalar drifts less than a list, and the list is exactly
what the index rebuilds cheaply from inbound edges.

---

### Intrinsic Over Host-Varying

WHEN neither rule above settles the choice and one end of a relationship
holds regardless of context while the other varies by the document that
observes it, the author SHALL declare it on the context-independent end.

**Rationale:** This is [[adr-009]]'s own logic. A shared module has
different neighbors in every container that composes it, so no single
outward list on the module is correct.

---

## Paired Declarations

A paired relationship — one the tool requires from both ends and reports as
broken when only one is present — SHALL be reserved for a relationship a
reader of one document must see **without consulting the index**.

**Rationale:** A pairing enforces a mirror, which the previous sections
say to avoid. Its only remaining justification is a document that must be
self-describing on its own, read without the graph the tool builds. That
is a narrow case, and it must be argued explicitly, not assumed.

---

WHILE the index is built on every run, a paired declaration SHOULD NOT be
used to enforce a mirror the index already derives.

**Rationale:** Requiring both ends when the tool computes either from the
other asks an author to maintain the redundant half by hand and then flags
them when they do not. That is maintenance burden with negative value: the
check can only ever fail on work the tool did not need done.

---

## References

- [[frd-spec-reference-integrity]] — records the reference graph in both
  directions (AC-SPECREF-001.1, AC-SPECREF-001.3) and exposes the reverse
  as a query (AC-SPECREF-003); the capability these standards rest on
- [[frd-spec-index]] — the index every relationship is derived
  into, and the artifact a consumer reads instead of re-deriving the graph
- [[adr-009]] — parents own the edges among their children; the source of
  the intrinsic-over-host-varying tie-breaker
- [[adr-012]] — the `[[slug]]` wikilink convention prose relationships are
  written in and resolved against
- `plugin/skills/spec/references/cross-document-references.md` — what a
  reference resolves against

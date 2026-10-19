---
title: "ADR-030: C4 Vocabulary Reaches Document Core Through the Profile"
description: The C4 diagram parser, element resolution and model assembly
  live in Document Core, and the framework's architecture vocabulary
  reaches them as a profile registration, as the Gherkin tag vocabulary
  already does
slug: adr-030
type: adr
status: approved
---

# ADR-030: C4 Vocabulary Reaches Document Core Through the Profile

**Status:** approved  
**Date:** 2026-10-06

## Context

[[frd-spec-c4-diagrams]] adds `c4/*` checks to `specs check`, and
[[frd-spec-c4]] adds a `specs c4` command that emits the corpus's C4 model.
Both need two different kinds of knowledge:

- **Notation:** how to write a Mermaid C4 statement. [[adr-029]] gives this
  knowledge to a parser that this CLI owns.
- **Vocabulary:** this framework's terms. For example, a
  `container-blueprint` declares a container, its `entity_name` is the name
  that a diagram draws, `system` and `modules` are its containment edges, its
  children sit under the `Components` heading in `component` record blocks,
  and a contract's `provider` names the element that publishes the contract.

[[container-spec-cli]]'s Change Guidance puts every capability in one of its
two modules, never in the command surface. [[module-document-core]] SHALL NOT
depend on [[module-spec-profile]] and holds no document type, status, or
property name of its own; a source scan of `cli/core/corpus` enforces that.
[[module-spec-profile]] SHALL NOT parse document text. So, as things stand,
neither module can hold both kinds of knowledge.

The Gherkin capability met the same split and settled it this way: the
parser and the feature and criterion records belong to Document Core, and
the identifier patterns, the owning-document prefix, and the qualifier tags
arrive through `Profile.FeatureTags()`.

## Decision

We put the C4 diagram parser, element resolution, and model assembly in
Document Core. We widen Document Core's `Profile` extension point with one
registration that carries the architecture vocabulary that these components
read:

- The document types that declare elements, and the C4 level of each.
- The frontmatter properties that name an element's entity name, its system,
  its composed modules, its consumed contracts, and a contract's provider.
- The contract document type.
- The roster heading and record-block tag that declare a parent's children.

[[module-spec-profile]] fills the registration. Document Core names none of
those values.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Document Core components, vocabulary registered by the profile | Follows the Gherkin precedent exactly<br>Keeps both modules' dependency rules intact<br>A profile with another taxonomy reuses the parser and the model | The extension point widens again<br>Document Core gains code that knows the C4 notation | Not applicable (selected) |
| Hard-code the vocabulary in Document Core | No registration to design | Breaks the module's no-vocabulary rule and fails its source scan | Violates [[module-document-core]]'s Change Guidance |
| Put the capability in the profile module | The vocabulary is already there | The profile would parse document text, which its interface forbids<br>Another profile couldn't reuse the parser | Violates [[module-spec-profile]]'s interface guarantees |
| A third module for C4, between core and profile | Document Core stays free of C4 notation | Checks run inside Document Core's `Check`, which can't take rules from another module<br>Requires a new module blueprint and a third dependency edge to police | Requires a rule-plugin mechanism that the CLI doesn't have, for a separation that the Gherkin precedent already decided against |
| Implement in the command surface | Fastest to build | A future API or MCP wrapper wouldn't have it | Violates [[container-spec-cli]]'s Change Guidance |

## Consequences

**Positive:**

- The parser, the resolution, and the model are available to any wrapper of
  Document Core, not only to `specs`.
- Every vocabulary value that the C4 code reads is registered in one place,
  and can be checked against the generated schema in the same way as the
  reference properties.
- A profile that registers no architecture vocabulary gets a corpus with no
  elements: the check raises no `c4/*` finding and the model is empty. This
  case isn't an error.

**Negative:**

- `Profile` gains a seventh member, and `ValidateRegistration` gains the
  defects that the member can carry.
- Document Core now knows the Mermaid C4 notation. A corpus that never draws
  a diagram pays nothing for it at run time, but the code is still present.
- The registration duplicates facts that the schema states, such as that
  `entity_name` exists on every architecture blueprint. It therefore needs
  the same drift test that the reference registration has.

## References

- [[adr-029]]: relationships are read from blueprint diagrams by a parser
  that this CLI owns.
- [[adr-021]]: the Gherkin decision whose split this decision follows.
- [[adr-008]]: record blocks, which declare un-promoted elements.
- [[module-document-core]]: the host, and its no-vocabulary rule.
- [[module-spec-profile]]: the plugin that fills the registration.
- [[frd-spec-c4-diagrams]] and [[frd-spec-c4]]: the requirements that need
  this decision.

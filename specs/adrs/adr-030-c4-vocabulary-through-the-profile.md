---
title: "ADR-030: C4 Vocabulary Reaches Document Core Through the Profile"
description: The C4 diagram parser, element resolution and model assembly
  live in Document Core, and the framework's architecture vocabulary
  reaches them as a profile registration, as the Gherkin tag vocabulary
  already does
slug: adr-030
type: adr
status: approved
updated_at: 2026-10-06T17:45:00Z
---

# ADR-030: C4 Vocabulary Reaches Document Core Through the Profile

**Status:** approved  
**Date:** 2026-10-06

## Context

[[frd-spec-c4-diagrams]] adds `c4/*` checks to `specs check`, and
[[frd-spec-c4]] adds a `specs c4` command emitting the corpus's C4 model.
Both need two different kinds of knowledge. One is notation: how a
Mermaid C4 statement is written, which [[adr-029]] gives to a parser this
CLI owns. The other is this framework's vocabulary: that a
`container-blueprint` declares a container, that its `entity_name` is the
name a diagram draws, that `system` and `modules` are its containment
edges, that its children sit under the `Components` heading in
`component` record blocks, and that a contract's `provider` names the
element that publishes it.

[[container-spec-cli]]'s Change Guidance puts every capability in one of
its two modules, never in the command surface. [[module-document-core]]
SHALL NOT depend on [[module-spec-profile]] and holds no document type,
status or property name of its own; a source scan of `cli/core/corpus`
enforces that. [[module-spec-profile]] SHALL NOT parse document text.
So neither module can hold both kinds of knowledge as things stand.

The Gherkin capability met the same split and settled it one way: the
parser and the feature and criterion records are Document Core's, and the
identifier patterns, the owning-document prefix and the qualifier tags
arrive through `Profile.FeatureTags()`.

## Decision

We will put the C4 diagram parser, element resolution and model
assembly in Document Core, and widen its `Profile` extension point with
one registration carrying the architecture vocabulary they read: the
document types that declare elements and each one's C4 level, the
frontmatter properties naming an element's entity name, its system, its
composed modules, its consumed contracts and a contract's provider, the
contract document type, and the roster heading and record-block tag that
declare a parent's children. [[module-spec-profile]] fills it. Document
Core names none of those values.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Document Core components, vocabulary registered by the profile | Follows the Gherkin precedent exactly<br>Keeps both modules' dependency rules intact<br>A profile with another taxonomy reuses the parser and the model | The extension point widens again<br>Document Core gains code that knows the C4 notation | N/A — selected |
| Hard-code the vocabulary in Document Core | No registration to design | Breaks the module's no-vocabulary rule and fails its source scan | Violates [[module-document-core]]'s Change Guidance |
| Put the capability in the profile module | The vocabulary is already there | The profile would parse document text, which its interface forbids<br>Another profile could not reuse the parser | Violates [[module-spec-profile]]'s interface guarantees |
| A third module for C4, between core and profile | Document Core stays free of C4 notation | Checks run inside Document Core's `Check`, which has no way to take rules from another module<br>A new module blueprint, and a third dependency edge to police | Costs a rule-plugin mechanism the CLI does not have, for a separation the Gherkin precedent already decided against |
| Implement in the command surface | Fastest to build | A future API or MCP wrapper would not have it | Violates [[container-spec-cli]]'s Change Guidance |

## Consequences

**Positive:**
- The parser, the resolution and the model are available to any wrapper
  of Document Core, not only to `specs`.
- Every vocabulary value the C4 code reads is registered in one place
  and can be checked against the generated schema, the way the reference
  properties already are.
- A profile registering no architecture vocabulary gets a corpus with no
  elements: no `c4/*` finding is raised and the model is empty, rather
  than an error.

**Negative:**
- `Profile` gains a seventh member, and `ValidateRegistration` gains the
  defects that member can carry.
- Document Core now knows the Mermaid C4 notation. A corpus that never
  draws a diagram pays nothing for it at run time, but the code is
  there.
- The registration duplicates facts the schema states — that
  `entity_name` exists on every architecture blueprint — so it needs the
  same drift test the reference registration has.

## References

- [[adr-029]] — relationships are read from blueprint diagrams by a
  parser this CLI owns
- [[adr-021]] — the Gherkin decision whose split this one follows
- [[adr-008]] — record blocks, which declare un-promoted elements
- [[module-document-core]] — the host, and its no-vocabulary rule
- [[module-spec-profile]] — the plugin that fills the registration
- [[frd-spec-c4-diagrams]] and [[frd-spec-c4]] — the requirements that
  need it

---
title: "Schema Registry"
entity_name: Schema Registry
technology: Go package
description: Validation of frontmatter against the branch of the profile's
  union schema a document's type selects, and decoding and encoding of the
  typed envelope the index carries. Change it to change how a branch is
  selected, how a violation is reported, or how frontmatter crosses into
  and out of the generated union.
slug: component-schema-registry
type: component-blueprint
status: ready-for-review
kind: standard
contracts_consumed: []
source_paths:
  - "cli/core/schemaregistry"
  - "cli/core/envelope"
test_paths:
  - "cli/core/schemaregistry"
  - "cli/core/envelope"
---

# Schema Registry

## Responsibility

A profile registers one schema: a discriminated union over every document
type it defines. Schema Registry compiles that union once and validates each
document's frontmatter against the branch its declared `type` selects,
reporting every violation in one pass. It also carries frontmatter across
the boundary into the generated union variant the index records, and back
out, without naming a document type in either direction.

### Owns

Branch selection, validation, and the slug-pattern check.

Questions about the union: whether a type declares a property, and which
values an enum property allows.

Decoding parsed frontmatter into its typed union variant, encoding it for
the index, and reading the shared envelope out of a variant.

The rule for which member holds a slug two members declare.

### Hides

The JSON Schema library, how the union is discovered in a schema, and how a
violation reached through `$ref` is traced to its leaf constraint.

### Does not own

The schema's content — [[module-spec-profile]] owns it.

Parsing frontmatter — [[component-frontmatter]] owns it.

The finding model — [[module-document-core#Findings]] owns it.

### Composed into

[[module-document-core]] — Go packages the Corpus checks and the model
components call; see that blueprint for this component's relationships.

## Interface

### Surface

The exported identifiers of `cli/core/schemaregistry`, entered through
`New`, and of `cli/core/envelope`.

### Provides

`Validate` SHALL report every violation in one pass, each with the JSON
Pointer at which it occurred.

`Validate` SHALL report a violation reached through `$ref` by its leaf
constraint.

WHERE the selected branch constrains a property with a format keyword,
`Validate` SHALL validate the property's raw source text rather than its
parsed value.

`Decode` SHALL decode frontmatter that lacks a required property.

IF `Decode` refuses a property's value, THEN it SHALL return a
`PropertyError` naming that property and the value as written.

IF a variant leaves a required enum undeclared, THEN `Encode` SHALL omit it
rather than refuse to encode.

WHERE frontmatter declares every enum, `Encode` SHALL produce the bytes the
generated encoder produces.

`HoldsSlugOver` SHALL prefer a member whose frontmatter parsed over one
whose slug was recovered, and then the first in path order.

### Requires

WHEN `New` is called, the caller SHALL pass a JSON Schema in which each
document type's branch can be selected from the value of `type`.

## Design

### Patterns and principles

Adapter (Gamma et al., 1994) — `Registry` adapts the JSON Schema library to
the module's own types: it takes a parsed `frontmatter.Document` and returns
`specmodel.Finding` values. It serves the need [[adr-015]] records: the
validator is a replaceable choice.

### Invariants

No type from the JSON Schema library SHALL appear in the package's exported
surface. Enforced by: review only.

The component SHALL discover the union's variants from the generated union
itself, with no table naming a type. Enforced by:
`TestEveryUnionVariantIsDiscoveredUnderItsGeneratedTag` and
`TestNoDocumentTypeInPackageSource`.

Every site that chooses between two members declaring one slug SHALL decide
through `HoldsSlugOver`. Enforced by:
`TestHoldsSlugOverPrefersAParsedMemberThenPathOrder`; review only for the
call sites.

### Extension points

A new document type — a branch in the profile's union schema. Not by a
per-type schema pointer or a type name in this component.

## Cross-cutting Standards

This component SHALL comply with [[testing-standards]]: unit tests covering
its exported surface, cited to the acceptance criteria they verify.

## Change guidance

WHEN code elsewhere in the module needs a fact about the union schema (a
branch, a declared property, an enum's values), it SHALL ask this component
rather than read the schema itself.

### Verification gates

#### Verify always

`cd cli && go test ./core/schemaregistry/ ./core/envelope/`

#### Verify conditionally

WHEN `Encode` or `Decode` changes, run `cd cli && go test ./profile/`, which
encodes this repository's corpus with the generated encoder.

## Key decisions

[[adr-015]] — `santhosh-tekuri/jsonschema` is the validator.

[[adr-016]] — the union schema and its typed variants are generated from
Strata.

## References

[[module-document-core]] — the module this component is composed into.

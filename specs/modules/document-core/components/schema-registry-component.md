---
title: "Schema Registry"
entity_name: Schema Registry
technology: Go package
description: Validation of frontmatter against the branch of the profile's
  union schema a document's type selects, and decoding and encoding of the
  typed frontmatter variant the index carries. Change it to change how a
  branch is selected, how a violation is reported, or how frontmatter
  crosses into and out of the generated union.
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

A profile is the set of document types a corpus uses, with their rules. It
registers one JSON Schema: a discriminated union with one branch per
document type, selected by the value of the document's `type` property.
Schema Registry compiles that union once and validates each document's
frontmatter against the branch its `type` selects. It reports every
violation in one pass. It also decodes frontmatter into a variant of the Go
union type generated from the same Strata source as the schema (one struct
per document type), and encodes a variant to the JSON the corpus index
stores. Neither direction hard-codes a document type.

### Owns

Branch selection, validation, and the slug-pattern check
(`CheckSlugPattern`).

Answers about the union: whether a type's branch declares a property
(`Registry.Declares`), and which values an enum property allows
(`Registry.Enum`).

Decoding parsed frontmatter into its union variant, encoding a variant for
the index, and reading a variant's envelope: its type and the properties
every document type shares, such as `slug` and `status`.

The rule for which corpus document answers lookups for a slug that two
corpus documents declare.

### Hides

The JSON Schema library, how the union is discovered in a schema, and how a
violation reached through `$ref` is traced to its leaf constraint.

### Does not own

The schema's content — [[module-spec-profile]] owns it.

Parsing frontmatter — [[component-frontmatter]] owns it.

The finding model — [[module-document-core#Findings]] owns it.

### Composed into

[[module-document-core]] — Go packages that the Corpus component and other
Document Core components call; see that blueprint for this component's
relationships.

## Interface

### Surface

The exported identifiers of the Go packages `cli/core/schemaregistry` and
`cli/core/envelope`. In `schemaregistry`, the `New` function returns a
`Registry`, and `Validate` is a method of that type. `Decode`, `Encode`,
and `HoldsSlugOver` are functions in `envelope`, and `PropertyError` is a
type there.

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
generated encoder (`json.Marshal` over the generated union type)
produces.

`HoldsSlugOver` SHALL prefer a corpus document whose frontmatter parsed over
one with a recovered slug, and then the first in path order.

### Requires

WHEN `New` is called, the caller SHALL pass a JSON Schema in which each
document type's branch can be selected from the value of `type`.

## Design

### Patterns and principles

Adapter (Gamma et al., 1994) — the `Registry` type adapts the JSON Schema
library to the module's own types: it takes a parsed `frontmatter.Document`
and returns `specmodel.Finding` values. It keeps the validator that
[[adr-015]] chose replaceable.

### Invariants

No type from the JSON Schema library SHALL appear in the exported surface of
`schemaregistry` or `envelope`. Enforced by: review only.

The component SHALL discover the union's variants from the generated union
itself, with no table naming a type. Enforced by:
`TestEveryUnionVariantIsDiscoveredUnderItsGeneratedTag`, which fails when
`envelope` misses a variant, and `TestNoDocumentTypeInPackageSource`, which
scans only `cli/core/schemaregistry`. A type named in `cli/core/envelope` is
caught by review only.

Every call site that chooses between two corpus documents declaring one slug
SHALL decide through `HoldsSlugOver`. Enforced by:
`TestHoldsSlugOverPrefersAParsedMemberThenPathOrder`; review only for the
call sites.

### Extension points

A new document type — a branch in the profile's union schema, not a type
name in this component nor a per-type JSON Pointer into that schema.

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
checks that the index of this repository's corpus encodes to the same bytes
as the generated encoder writes.

## Key decisions

[[adr-015]] — `santhosh-tekuri/jsonschema` is the validator.

[[adr-016]] — the union schema and its typed variants are generated from
Strata.

## References

[[module-document-core]] — the module this component is composed into.

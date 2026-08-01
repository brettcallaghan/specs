# Profile 

## Table of Contents

- [ProfileRegistration](#profileregistration)
- [DocumentTypeRegistration](#documenttyperegistration)
- [Section](#section)
- [PropertyPairing](#propertypairing)
- [BlockTagRegistration](#blocktagregistration)
- [BlockFormat](#blockformat)

## ProfileRegistration

What a profile registers into Document Core.

**Internal surface.** This is the vocabulary a profile hands to the
engine in-process, not something an adopting repository authors. It is
not reachable from `specs.toml` in v1 and is not an extension point
outside this repository; `[profile].name` selects one of these
registrations by name, and that naming is the whole of the relationship
between the two halves of the configuration contract.

It is modelled here rather than left to hand-written Go for two
reasons. Registration is data — type names, patterns, section trees,
block tags — and only the rules that cannot be expressed as data stay in
code. And the open question of which internal keys get promoted to the
public surface is answered by editing this struct, which is a smaller
and more reviewable act than editing an interface and its
implementations.

Property names are snake_case throughout, matching the public surface,
so that promoting a key to `specs.toml` is a move rather than a rename.

| Property | Type | Documentation |
|----------|------|---------------|
| name | String | The name an adopting repository writes in `[profile].name`. |
| document_types | Map<String, [DocumentTypeRegistration](#documenttyperegistration)> | The document types this profile defines, keyed by the value documents declare in their `type` property.<br><br>A markdown file declaring a type absent from this map is recorded as excluded with reason `unknown-type` and its declared value reported, so a reviewing agent can judge whether the profile or the document is wrong. |
| block_tags | Array<[BlockTagRegistration](#blocktagregistration)>? | Fenced block info strings whose contents are parsed as structured data and attached to the containing document and heading. |
| finished_statuses | Array<[Status](vocabulary.md#status)>? | The statuses at which a document is considered finished, and therefore at which leftover template placeholders are a defect rather than work in progress. |
| placeholder_patterns | Array<String>? | Patterns whose presence in a finished document indicates unreplaced template scaffolding. |
| conventions | [ConventionsConfig](config.md#conventionsconfig)? | The convention values this profile supplies when `specs.toml` omits them. |


## DocumentTypeRegistration

Everything the engine needs to check one document type.

| Property | Type | Documentation |
|----------|------|---------------|
| schema | String | The JSON Schema this type's frontmatter is validated against, expressed as a JSON Pointer into the bundled schema or as a path to a standalone one. |
| slug_pattern | String | The pattern a document of this type must match in its `slug`.<br><br>A mismatch is reported with both the actual slug and this pattern quoted, so the finding is self-contained and the fixing agent needs no further lookup. |
| numbered | Bool | Whether documents of this type carry a sequence number the CLI can derive the next value of.<br><br>Numbering is global per type: the next number is one greater than the highest in use anywhere in the corpus, regardless of directory. |
| sections | Array<[Section](#section)>? | The heading hierarchy a document of this type must contain. |
| section_sets | Map<String, Array<String>>? | Named sets of headings a span query can select, so a reviewing agent can ask for the normative sections of a document and read only those. |
| reference_properties | Array<String>? | Frontmatter properties of this type that hold document references.<br><br>A value in one of these properties that matches no corpus member's slug is reported the same way an unresolved `[[slug]]` in the body is. |
| pairings | Array<[PropertyPairing](#propertypairing)>? | Relationship properties that must be declared from both ends. |


## Section

A required heading, and the headings required beneath it.

Required sections are a hierarchy rather than a flat list because the
defects worth catching are about nesting: a missing `## Consequences`
and a `### Consequences` misfiled under `## Decision` are different
problems, and a flat list cannot tell them apart.

| Property | Type | Documentation |
|----------|------|---------------|
| title | String | The heading text, matched exactly. |
| required | Bool | Whether a document of this type must contain the heading. |
| repeated | Bool | Whether the heading occurs once per instance of something in the document rather than exactly once — one subsection per alternative considered, for example.<br><br>Without this distinction a section tree can only describe documents of fixed shape, and not all of the templates this framework ships are fixed. |
| children | Array<[Section](#section)>? | Headings required one level beneath this one. |


## PropertyPairing

One half of a relationship that must be declared from both ends.

A blueprint naming a child in `children` where that child declares no
`parent_slug` is a half-built edge, and it is reported against both
documents rather than only the one that happens to be checked first.

| Property | Type | Documentation |
|----------|------|---------------|
| property | String | The property declared on this document. |
| counterpart | String | The property the referenced document must declare in return. |


## BlockTagRegistration

A fenced block whose contents carry structured data about an element
declared inside a document.

| Property | Type | Documentation |
|----------|------|---------------|
| tag | String | The fenced block's info string — `component` for a component record block. |
| format | [BlockFormat](#blockformat) | The format the block's contents are parsed as. A block that fails to parse is reported with the parse error and its position within the containing document. |
| schema | String? | The JSON Schema the parsed contents are validated against, expressed as a JSON Pointer into the bundled schema or as a path to a standalone one. |


## BlockFormat

The formats a registered fenced block's contents may be parsed as.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Yaml | "yaml" | YAML, the format this framework's element record blocks use. |
| Json | "json" | JSON. |
| Toml | "toml" | TOML. |



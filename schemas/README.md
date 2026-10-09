# Schemas

The data models for the Spec CLI's configuration and for the frontmatter
every specification document carries, written once in
[Strata](https://github.com/strata-lang/strata) and transpiled to the
three forms this project consumes them in.

```
schemas/*.strata  ──►  cli/core/schema/*.json      JSON Schema, for validation
                  ──►  cli/core/specmodel/*.go     Go structs, for the CLI
                  ──►  plugin/skills/spec/references/generated/*.md
                                                    Markdown, shipped with the plugin
```

**The `.strata` files are the only source of truth.** Every generated
artifact carries a do-not-edit banner. Regenerate with:

```bash
scripts/generate-schemas.sh           # rewrite the generated tree
scripts/generate-schemas.sh --check   # fail if anything is stale (for CI)
specs check --strict                   # check this repo's own documents
```

## The sources

| File | Contents |
|---|---|
| `vocabulary.strata` | The controlled vocabularies: `Slug`, `DocumentType`, `Status`, `Visibility`, `BlueprintKind`, `Delivery`, `RfcSeverity` |
| `frontmatter.strata` | `ElementRecord`, the `DocumentFrontmatter` envelope, the `ContractEdges` mixin, and one struct per document type |
| `config.strata` | `SpecsConfig` — the public `specs.toml` surface |
| `profile.strata` | The internal profile registration surface |
| `index.strata` | `SpecIndex` — the emitted `specs.json`, its entries and exclusions, plus the `Span` and `ParseError` positions every report carries |
| `feature.strata` | `Feature` and `Criterion` — the feature and acceptance-criterion records `specs.json` carries, their steps, example rows, qualifiers and content hashes |

## Design decisions

**`ElementRecord` and `DocumentFrontmatter` share a root, and
architecture blueprints inherit both.** Both extend `Identified`
(`description` and `slug`); `ElementRecord` adds `entity_name` and
`technology`, and `DocumentFrontmatter` the document `title` and
envelope. Inheriting both is what `frontmatter-standards.md` asks for: the element record and the document
envelope defined as two schemas and merged with JSON Schema `allOf`, so
that one schema validates a child element whether it is embedded in a
parent's record block or promoted to a document of its own. An element's
name and its document's title are separate properties because they are
separate facts, even where they read the same — only the name is
matched against diagram labels. Strata's `struct Child: Parent` emits precisely that `allOf` +
`$ref` shape, so the requirement is met by the language rather than by
hand-assembly.

**Every `Span`-typed property is named `span`, or `<role>_span` where
a record carries more than one.** A span is an extent — a start and an
end — so neither `position` nor `location`, which name a point, is used.
A property that locates its own record is `span`; one that locates
something else names it: `section_span` for the content beneath a
heading, `block_span` for a feature's whole fence, `other_span` for the
other occurrence of a duplicate.

**Property names are snake_case, against Strata's camelCase
convention.** Strata emits the property name verbatim as the wire key,
and the frontmatter keys (`source_paths`, `superseded_by`, `entity_name`)
are already in use across the corpus. Matching the existing files wins
over matching the language's house style. The Go transpiler re-cases the
Go identifier anyway, so the generated structs read as idiomatic Go with
correct `json:"source_paths"` tags.

**`ContractEdges` is a mixin.** `contracts_consumed` means the same thing
on all four blueprint types, so it is declared once and inherited, which
makes divergence impossible rather than merely discouraged. The other end
of a contract edge is `Contract.provider`; neither end is mirrored. For
an architecture blueprint Strata emits a four-branch `allOf` — the
type's own properties, `DocumentFrontmatter`, `C4ElementRecord` and
`ContractEdges` — and the Go transpiler flattens all four into one
struct.

**`type` is the discriminator of the `TypedFrontmatter` union, not a
property of the base.** `TypedFrontmatter` is a `@discriminator("type")`
union with one variant per document type, and it is what the index
carries as an entry's `frontmatter` — so the index holds every property a
document's type defines, not only the envelope every type shares. Strata
forbids a variant from declaring its discriminator, so
`DocumentFrontmatter` and the per-type structs carry no `type` property;
each union branch carries it instead, as the constant for its variant
(`"adr"` on `Adr`). Documents and encoded values still carry `type`, and
the CLI validates a document against the branch its `type` selects
rather than against the bare struct, which has no `type` to accept. The
`DocumentType` enum remains as the vocabulary of declared types. Variant
names are the kebab-case of their tags (`SystemBlueprint` is
`system-blueprint`); the CLI derives one from the other, so a new variant
must keep the convention.

**Requiredness follows the templates, not the prose.** `version` is
required on `Contract` because its template declares it, and optional
on `ModuleBlueprint` because it is required only when `visibility` is
`public` — a conditional Strata cannot express and the
profile must enforce. Relationship arrays are all optional, because the
standards say to omit them when empty while the templates ship them as
`[]`; optional accepts both.

**Work orders are deliberately absent.** The work-order template carries
no `status`, so it cannot inherit `DocumentFrontmatter`, and work orders
are tracked in beads rather than as corpus documents. They already have
a schema at `plugin/skills/groom-beads/references/bead.schema.json`;
duplicating it here would create two authorities for one vocabulary.

**`rfc` and `guide` are included** because 13 real documents declare
them, even though neither has a template. Without them those documents
are excluded from every check as `unknown-type`.

**Undeclared keys are rejected.** The manifest sets
`rejectUnknownProperties: true`, which puts `additionalProperties: false`
on every emitted object — what `specs-configuration-contract.md` asks for,
on the grounds that a misspelled key is far more common than a key from
the future. A document carrying `releated:` fails validation naming the
key. The generated Go decoders enforce the same policy: unmarshalling an
undeclared key returns an `unknown field` error.

Closing a shape changes the emitted *shape* as well as its strictness:
a closed type cannot be composed with `allOf`, because the `$ref`'d
supertype would independently reject every property the subtype adds.
Strata therefore flattens ancestors into each type instead, so
`ContainerBlueprint` lists all twelve of its properties inline rather
than `$ref`ing `DocumentFrontmatter` and `ContractEdges`. For the same
reason a closed union inlines each variant, discriminator included, as an
`anyOf` branch rather than `$ref`ing the variant's own closed definition.
Semantics are identical; only `$ref` sharing is lost, and the generated
file is larger for it.

**The CLI still needs its own unknown-key detection.**
`specs-configuration-contract.md` requires each finding to carry a
**span** into `specs.toml`, and JSON Schema has no notion of source
position. The schema is the second line of defence, not the first.

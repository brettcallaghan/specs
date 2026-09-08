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

**`ElementRecord` is the root, and `DocumentFrontmatter` inherits it.**
`frontmatter-standards.md` asks for exactly this: the base record fields
and the document envelope defined as two schemas and merged with JSON
Schema `allOf`, so that one schema validates a child element whether it
is embedded in a parent's record block or promoted to a document of its
own. Strata's `struct Child: Parent` emits precisely that `allOf` +
`$ref` shape, so the requirement is met by the language rather than by
hand-assembly.

**Property names are snake_case, against Strata's camelCase
convention.** Strata emits the property name verbatim as the wire key,
and the frontmatter keys (`updated_at`, `source_paths`, `superseded_by`)
are already in use across the corpus. Matching the existing files wins
over matching the language's house style. The Go transpiler re-cases the
Go identifier anyway, so the generated structs read as idiomatic Go with
correct `json:"updated_at"` tags.

**`ContractEdges` is a mixin, and it is the reason multiple inheritance
is used at all.** `contracts_provided` and `contracts_consumed` mean the
same thing on all four blueprint types, and before this they were spelled
three different ways at three altitudes (`contracts` on a system,
`provides`/`consumes` on a container, `provides` alone on a module) with
components carrying none. Declaring them once and inheriting them makes
divergence impossible rather than merely discouraged. Strata emits a
three-branch `allOf` — the type's own properties, `DocumentFrontmatter`,
and `ContractEdges` — and the Go transpiler flattens all three into one
struct.

**`type` is an enum on the base, not a constant per subtype.** Strata
forbids redeclaring an inherited property, so `Adr` cannot narrow
`type` to the constant `"adr"`. This costs nothing in practice: the CLI
reads `type` in order to choose which schema to validate against, so by
the time the schema runs, the value has already been established. A
`@discriminator("type")` union was the alternative and is worse — it
forbids the base from declaring `type` at all, which would remove it
from the envelope every document actually carries.

**Requiredness follows the templates, not the prose.** `version` is
required on `Contract` and `Standards` because their templates declare
it, and optional on `ModuleBlueprint` because it is required only when
`visibility` is `public` — a conditional Strata cannot express and the
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
`jsonSchema.allowAdditionalProperties: false`, which puts
`additionalProperties: false` on every emitted object — what
`specs-configuration-contract.md` asks for, on the grounds that a
misspelled key is far more common than a key from the future. A document
carrying `releated:` fails validation naming the key.

Closing a shape changes the emitted *shape* as well as its strictness:
a closed type cannot be composed with `allOf`, because the `$ref`'d
supertype would independently reject every property the subtype adds.
Strata therefore flattens ancestors into each type instead, so
`ContainerBlueprint` lists all fifteen of its properties inline rather
than `$ref`ing `DocumentFrontmatter` and `ContractEdges`. Inheritance
semantics are identical; only `$ref` sharing is lost, and the generated
file is larger for it. The Markdown and Go targets are unaffected —
both flatten anyway.

**The CLI still needs its own unknown-key detection.**
`specs-configuration-contract.md` requires each finding to carry a
**span** into `specs.toml`, and JSON Schema has no notion of source
position. The schema is the second line of defence, not the first.

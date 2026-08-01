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
scripts/validate-corpus.py            # check this repo's own documents
```

## The sources

| File | Contents |
|---|---|
| `vocabulary.strata` | The controlled vocabularies: `Slug`, `DocumentType`, `Status`, `Visibility`, `BlueprintKind`, `Delivery`, `RfcSeverity` |
| `frontmatter.strata` | `ElementRecord`, the `DocumentFrontmatter` envelope, the `ContractEdges` mixin, and one struct per document type |
| `config.strata` | `SpecsConfig` — the public `specs.toml` surface |
| `profile.strata` | The internal profile registration surface |

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

## Known gap: undeclared keys are not rejected

`contract-specs-configuration.md` requires that a key `specs.toml` does
not define be reported, on the grounds that a misspelled key is far more
common than a key from the future. The natural way to express that is
`jsonSchema.allowAdditionalProperties: false`, which adds
`additionalProperties: false` to every emitted object.

**That option cannot be used today.** Setting it makes the Strata
compiler drop struct inheritance from the JSON Schema output entirely —
the `allOf`/`$ref` to the parent disappears, so `Contract` emits with
only its own five properties and none of the envelope's six. The result
is a schema that validates but is wrong in both directions: a document
missing `title` passes, and a document carrying `title` fails. Reduced
to a minimal case, and independent of `bundle`:

```strata
struct Base  { title: String }
struct Child: Base { extra: String }
```

| `allowAdditionalProperties` | `Child` emitted as |
|---|---|
| `true` (default) | `{"allOf": [{...}, {"$ref": "#/definitions/Base"}]}` |
| `false` | `{"properties": {"extra": ...}, "additionalProperties": false}` — inheritance silently lost |

The manifest therefore leaves the option at its default and the
generated schemas accept undeclared keys. Two consequences worth
holding onto:

- Unknown-key detection has to be implemented in the CLI regardless of
  this bug, because the contract requires each finding to carry a
  **span** into `specs.toml`, and JSON Schema has no notion of source
  position. The schema was always going to be the second line of defence
  rather than the first.
- When the compiler is fixed, turning the option back on is a
  one-line manifest change plus a regeneration. `scripts/validate-corpus.py`
  will start rejecting misspelled keys the moment it is.

Filed upstream against `strata-lang/strata`; see that repository's
`rfcs/` for the reproduction and the rest of the findings from
authoring this package.

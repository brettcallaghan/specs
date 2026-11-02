# The `specs` CLI

The `specs` CLI indexes a spec corpus, checks it for conformance, and answers
questions about it. For every flag, rule, output shape, and exit status, see
the full [Specs CLI reference](../../specs/guides/specs-cli-reference.md).

## Commands

| Command | Description |
|:---|:---|
| `specs index` | Walks the spec root, builds a machine-readable index of every document's metadata and location, and writes it to `specs.json` in the spec root, to the path in `--out`, or to standard output. |
| `specs check [path]` | Checks the corpus, or one document, for conformance: frontmatter against the active profile's schema, slug patterns, duplicate slugs, and unresolved references. Accepts `--json`, `--strict`, `--only <rule>`, and `--verbose`. |
| `specs next-id <type>` | Returns the next identifier for a numbered document type. |
| `specs slug-exists <slug>` | Reports whether any corpus member declares the slug. |
| `specs list` | Reports every document's path, slug, type, status, title, and description. `--type` filters the list, and `--all` includes superseded documents. |
| `specs refs <slug>` | Reports which documents refer to a slug or one of its headings, or what the slug refers to. |
| `specs headings` | Reports the headings that match a selection, with the span of each heading's section. |
| `specs criterion [id]` | Queries an acceptance criterion, or lists the criteria that changed locally. |
| `specs feature [id]` | Queries a feature, or writes one as a standalone Gherkin file. |
| `specs links <path>` | Rewrites a document's wikilinks as Markdown links. |
| `specs c4` | Writes the corpus's C4 model: the systems, containers, components, and modules that it documents, their relationships, and any relationship statements that it rejects. `--root <slug>` limits the model to one element. |

Every command also accepts `--config <file>` to name a `specs.toml` file
instead of searching for one. `specs --version` prints the build's version,
revision, time, and modified state as YAML.

## Checks in CI

`specs check` is the command to run in CI. It exits 0 when the corpus passes
and 1 when it finds a problem, and prints its findings on standard output.

| Command | Fails the build on |
|:---|:---|
| `specs check` | Error findings only. |
| `specs check --strict` | Any finding, including warnings and unverified findings such as a reference to a retired document. |

Use `--strict` when warnings should block a merge. This repository's own CI
runs `specs check --strict`.

## Configuration

The CLI reads `specs.toml`. See [Configuration](configuration.md).

## Profiles

The CLI's engine doesn't depend on any one framework. A **profile** supplies
the document vocabulary: a schema for every document type, plus each type's
slug pattern and whether it's numbered. For details, see
[Module: Specs Profile](../../specs/modules/spec-profile/spec-profile-module.md).

## Install

See [Install the plugin and the CLI](../how-to/install.md).

## Test coverage

For the CLI's test coverage by test layer and package, see
[Code coverage](../code-coverage.md).

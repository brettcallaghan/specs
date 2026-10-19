---
title: "ADR-016: Strata as the Single Source for Schemas and Generated Types"
description: Adopt Strata to define document frontmatter and CLI
  configuration once and generate JSON Schema, Go structs and Markdown,
  accepting a pre-1.0 dependency on a tool this project's owner also
  maintains
slug: adr-016
type: adr
status: ready-for-review
---

# ADR-016: Strata as the single source for schemas and generated types

**Status:** ready-for-review  
**Date:** 2026-08-01

## Context

The Spec CLI described by [[frd-spec-index]] needs the same data shapes in
three forms at once:

- JSON Schema to validate document frontmatter and `specs.toml`
  ([[contract-specs-configuration]])
- Go structs for the CLI and its modules ([[adr-015]])
- Human-readable Markdown, shipped inside the plugin, so an author knows
  what frontmatter to write

Today those three exist only as prose in
`plugin/skills/spec/references/frontmatter-standards.md` and the templates,
and they already disagree. A prototype found 26 broken references and one
unparseable document that no reviewer had caught.

Three properties of this problem constrain the choice. The first one
eliminates most candidates:

- There's no service. This is document and configuration modeling: YAML
  frontmatter in Markdown files, and a TOML config. There are no
  operations, no protocol, no client or server, and no wire negotiation.
  Most mature tools in this space are API description languages.
- Humans and agents author documents, which are then validated. The schema
  is the contract and the Go code is one consumer of it, so the direction
  must be schema-first rather than code-first.
- The toolchain must stay thin and reachable. The repository runs on bash,
  `python3`, and (per [[adr-015]]) Go. Adopting a tool that drags in a
  second language runtime is a real cost, and [[frd-spec-index]]
  AC-SPEC-009.5 rules out fetching executable content from the network at
  install time.

A conflict of interest applies to this decision, and this ADR records it
deliberately. Strata is maintained by this repository's own owner. That
makes it easier to get fixes landed and much harder to evaluate neutrally.
It's also why this ADR includes options that a same-owner tool would
normally not be compared against.

## Decision

We will adopt Strata as the single source of truth for the Spec CLI's
document-frontmatter and configuration data models, generating JSON Schema,
Go structs, and Markdown from `.strata` sources under `schemas/`.

The adoption is scoped to data-shape definition only and is explicitly
reversible. Generated artifacts are committed, so the repository remains
buildable and correct if Strata is removed.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| TypeSpec (Microsoft) | JSON Schema emitter is GA and battle-tested at vendor scale, and emits 2020-12 with `$defs` bundling<br>Can emit standalone models without declaring a service<br>Large vendor behind it, with an active 1.0 line | Requires a Node and npm runtime the repo doesn't otherwise need<br>Go generation is `@azure-tools/typespec-go`, an Azure SDK client emitter producing `models.go`/`service.go` with client and tracing options, not a plain-struct emitter<br>No Markdown documentation emitter in the standard library | The two artifacts this project most needs, plain Go structs and human-readable Markdown, are the two TypeSpec doesn't produce. Its JSON Schema emitter is more mature than Strata's: it's GA, against a pre-1.0 tool in which this evaluation found, and got fixed, a correctness bug. It isn't more capable. Strata was tested at `schemaVersion: Draft202012` and matches TypeSpec on 2020-12 output and `$defs` bundling, and adds root filtering and an OpenAI Structured Outputs dialect. On the schema half of the comparison, the two differ in reliability, not in features |
| Smithy (AWS) | The most mature option by a wide margin; its Go codegen powers `aws-sdk-go-v2`<br>Protocol-agnostic, with a JSON AST, strong validation, and a traits model<br>Documentation generation is a first-class stated capability | Java toolchain: a Gradle plugin, or the JVM Smithy CLI plus `smithy-build.json`<br>Designed around service and operation shapes; `smithy-jsonschema` contextualizes output by service shape ID<br>Go codegen is SDK-client-shaped, not plain data models | Wrong shape and heaviest toolchain. Modeling Markdown frontmatter as a Smithy service to get a JSON Schema out inverts the tool's purpose, and a JVM and Gradle dependency is a larger tax than the Rust one it would replace |
| Go structs as the source, JSON Schema by reflection (for example, `invopop/jsonschema`) | No new language runtime at all, because the CLI is already Go<br>Zero generated-artifact drift for the Go half, since Go is the source<br>Smallest possible toolchain | Inverts the direction: the schema becomes a projection of one consumer's implementation<br>Yields no Markdown artifact, so the plugin's shipped frontmatter reference stays hand-maintained and drifts<br>Struct tags and reflection express constraints far less directly than a schema language | The closest call on this list, and the one that would win if Go were the only output. It loses on the third deliverable. The plugin ships documentation of the frontmatter shape, and this option leaves that documentation hand-written, which is the drift the decision exists to end |
| Hand-author JSON Schema, Go structs, and Markdown separately | No tooling, no build step, and no new dependency<br>Full control over every artifact | Three artifacts to keep in agreement by discipline alone<br>This is the status quo that already produced the drift cited in Context | Rejected by the evidence. The prose-only version of these shapes is already inconsistent, and adding two more hand-maintained artifacts makes that worse |
| Strata | Generates all three needed artifacts (JSON Schema, plain Go structs, and Markdown) from one source<br>Struct inheritance emits the `allOf` + `$ref` shape that `frontmatter-standards.md` independently specified<br>Doc blocks become JSON Schema `description`s, Go doc comments, and Markdown tables from one text<br>Single static binary at runtime, and diagnostics name the fix | Pre-1.0 (`0.7.0`), with a thin ecosystem and effectively a bus factor of one<br>A silent correctness bug was found during evaluation: `allowAdditionalProperties: false` dropped struct inheritance from emitted schemas. It was fixed upstream and merged the same day, but only on `main`<br>No compatibility-diff tooling and no consumer-side output verification<br>Rust build-time dependency, and release binaries are unreachable from locked-down networks<br>Generated Go carries `json:` tags only, with no YAML or TOML tags | N/A (selected) |

## Consequences

**Positive:**

- One definition produces the three artifacts that previously disagreed,
  and the disagreement in Context becomes mechanically impossible rather
  than a review responsibility.
- `specs check` checks every document's frontmatter against the generated
  schema on each push, and found two documents writing an unquoted
  `version` that YAML reads as a float.
- Doc comments written once reach three audiences: the JSON Schema a tool
  reads, the Go doc comment a maintainer reads, and the Markdown an author
  reads.
- Strata's inheritance emits exactly the two-schema `allOf` composition
  that `frontmatter-standards.md` already required. A rule previously
  enforced by hand-assembly guidance is enforced by the language.
- The decision is cheap to reverse. Generated artifacts are committed, so
  removing Strata means freezing the current output and hand-editing it
  thereafter, with no migration and no rewrite.

**Negative:**

- The project depends on a pre-1.0 tool with one maintainer, who is also
  this project's owner. Fixes are easy to land and independent scrutiny is
  scarce. A defect found here will usually be found here first.
- Generated output isn't verifiable by its consumer. Strata offers no way
  to assert that an emitted schema still accepts the documents it should.
  The `allowAdditionalProperties` bug found during this evaluation exited
  zero and emitted well-formed, wrong JSON Schema. Inspection caught it,
  not any gate. That specific defect is fixed, but the general exposure
  isn't. Until Strata ships such a check, `specs check` is the only thing
  standing in for it, and it only covers what a stock JSON Schema
  validator can see.
- The schemas track Strata's `main`, not a release. The fix landed after
  `0.7.0`, so `scripts/generate-schemas.sh` must run against a compiler
  built from source until the next release. The binary reports
  `strata-cli 0.7.0` either way, because the crate version didn't change,
  so the version string can't distinguish a fixed compiler from the buggy
  release. `~/.cache/strata-cli/built-commit`, written by the
  session-start hook, is what identifies the build.
- Building the CLI requires a Rust toolchain, in a repository that is
  otherwise bash, Python, and Go. Release binaries exist for four
  platforms, but their bytes are unreachable through this environment's
  egress proxy, so a cold container pays for a `cargo build --release`
  (~2 minutes). Committing the generated artifacts mitigates this but
  doesn't solve it: only someone changing a `.strata` file needs the
  compiler. This is why `scripts/generate-schemas.sh --check` isn't yet a
  CI gate, and CI therefore can't catch a stale generated artifact.
- Generated Go structs carry `json:` tags only. Frontmatter is YAML and
  configuration is TOML, so neither deserializes directly into a generated
  struct. Both must be parsed to a generic model and converted to JSON
  first. That two-step path is the one [[module-document-core]]'s
  Frontmatter component already describes, and the component needs it
  anyway to retain each property's raw text for AC-SPEC-003.2.
- No compatibility tooling exists. Strata documents its own breaking-change
  rules precisely but enforces none of them, so "did this schema change
  break a consumer?" stays a human judgment.
- Strata's headline strength is largely unused here. It generates for
  eight languages, and this project needs one. The decision rests on the
  Markdown and JSON Schema targets plus the schema-first direction, not on
  the cross-language story. A future reader re-evaluating it should weigh
  it on those grounds, not on target count.

## References

- [[adr-015]]: Go as the CLI's implementation language, which fixes the
  Go-struct requirement this decision serves
- [[frd-spec-index]]: the feature these models exist for.
  AC-SPEC-009.5 constrains install-time network access
- [[contract-specs-configuration]]: the `specs.toml` surface modeled in
  `schemas/config.strata`
- [[module-document-core]]: the Frontmatter component that owns the
  YAML-to-JSON conversion step
- `schemas/README.md`: the design decisions inside the models, the
  regeneration commands, and what closing the shapes costs in emitted
  schema size
- `strata-lang/strata#289`: the upstream fix for the
  `allowAdditionalProperties` inheritance bug found during this
  evaluation
- TypeSpec JSON Schema emitter:
  https://www.npmjs.com/package/@typespec/json-schema
- TypeSpec Go emitter (Azure SDK client generator):
  https://www.npmjs.com/package/@azure-tools/typespec-go
- Smithy code generation guide:
  https://smithy.io/2.0/guides/using-code-generation/index.html
- `software.amazon.smithy:smithy-jsonschema`:
  https://mvnrepository.com/artifact/software.amazon.smithy/smithy-jsonschema

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

# ADR-016: Strata as the Single Source for Schemas and Generated Types

**Status:** ready-for-review  
**Date:** 2026-08-01

## Context

The Spec CLI that [[frd-spec-index]] describes needs the same data shapes
in three forms: **JSON Schema** to validate document frontmatter and
`specs.toml` ([[contract-specs-configuration]]), **Go structs** for the CLI
and its modules ([[adr-015]]), and **human-readable Markdown** that ships
inside the plugin so authors know what frontmatter to write. As of this
ADR's date, those three forms exist only as prose in
`plugin/skills/spec/references/frontmatter-standards.md` and the
templates, and the prose already disagrees with itself. A prototype found
26 broken references and one unparseable document that no reviewer had
caught.

Three properties of the problem constrain the choice. The first one
eliminates most candidates:

- **There is no service.** The task is to model documents and
  configuration: YAML frontmatter in Markdown files, and a TOML
  configuration file. There are no operations, no protocol, no client or
  server, and no wire negotiation. Most mature tools in this space are API
  description languages.
- **Humans and agents author documents, and a validator checks them.** The
  schema is the contract, and the Go code is one consumer of it, so the
  design must be schema-first rather than code-first.
- **The toolchain must stay small and reachable.** The repository runs on
  bash, `python3`, and (per [[adr-015]]) Go. A tool that requires a second
  language runtime has a real cost, and [[frd-spec-index]] AC-SPEC-009.5
  rules out fetching executable content from the network at install time.

**This decision has a conflict of interest, and this ADR records it on
purpose.** This repository's owner also maintains Strata. That makes fixes
easier to land and a neutral evaluation much harder. For that reason, this
ADR compares Strata against options that a same-owner tool would normally
not be compared against.

## Decision

We will adopt **Strata** as the single source of truth for the Spec CLI's
document-frontmatter and configuration data models. Strata generates JSON
Schema, Go structs, and Markdown from `.strata` sources under `schemas/`.

The adoption covers **data-shape definition only**, and you can reverse
it. The repository commits the generated artifacts, so it stays buildable
and correct if Strata is removed.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **TypeSpec** (Microsoft) | The JSON Schema emitter is GA and proven at vendor scale; it emits 2020-12 with `$defs` bundling<br>Can emit standalone models without declaring a service<br>A large vendor supports it; active 1.0 line | Requires a Node and npm runtime that the repository doesn't otherwise need<br>Go generation is `@azure-tools/typespec-go`, an Azure SDK **client** emitter that produces `models.go` and `service.go` with client and tracing options. It isn't a plain-struct emitter<br>No Markdown documentation emitter in the standard library | TypeSpec doesn't produce the two artifacts this project needs most: plain Go structs and human-readable Markdown. Its JSON Schema emitter is **more mature** than Strata's, which is pre-1.0 and had a correctness bug that this evaluation found and got fixed. It isn't more **capable**. Tested at `schemaVersion: Draft202012`, Strata matches it on 2020-12 output and `$defs` bundling, and adds root filtering and an OpenAI Structured Outputs dialect. On the schema side, the two differ in reliability, not in features |
| **Smithy** (AWS) | By far the most mature option; its Go code generation powers `aws-sdk-go-v2`<br>Protocol-agnostic, with a JSON AST and a strong validation and traits model<br>Documentation generation is a stated first-class capability | Java toolchain: a Gradle plugin, or the JVM Smithy CLI plus `smithy-build.json`<br>Designed around service and operation shapes; `smithy-jsonschema` places its output in the context of a service shape ID<br>Go code generation produces SDK clients, not plain data models | It has the wrong shape and the heaviest toolchain. Modeling Markdown frontmatter as a Smithy service to get a JSON Schema reverses the tool's purpose. A JVM and Gradle dependency also costs more than the Rust dependency it would replace |
| **Go structs as the source, JSON Schema by reflection** (for example, `invopop/jsonschema`) | No new language runtime, because the CLI is already Go<br>No generated-artifact drift for the Go half, because Go *is* the source<br>Smallest possible toolchain | Reverses the direction: the schema becomes a projection of one consumer's implementation<br>Produces no Markdown artifact, so the plugin's frontmatter reference stays hand-maintained and drifts<br>Struct tags and reflection express constraints much less directly than a schema language | This is the closest call on the list, and it would win if Go were the only output. It loses on the third deliverable. The plugin ships documentation of the frontmatter shape, and this option leaves that documentation hand-written, which is the drift this decision exists to end |
| **Hand-author JSON Schema, Go structs, and Markdown separately** | No tooling, no build step, and no new dependency<br>Full control over every artifact | Only discipline keeps three artifacts in agreement<br>This is the status quo that already produced the drift described in Context | The evidence rules it out. The prose-only version of these shapes is already inconsistent, and two more hand-maintained artifacts would make that worse |
| **Strata** | Generates all three required artifacts (JSON Schema, plain Go structs, and Markdown) from one source<br>Struct inheritance emits the `allOf` + `$ref` shape that `frontmatter-standards.md` specified independently<br>One doc block becomes a JSON Schema `description`, a Go doc comment, and a Markdown table entry<br>A single static binary at run time; diagnostics name the fix | Pre-1.0 (`0.7.0`), with a small ecosystem and effectively one maintainer<br>The evaluation found a silent correctness bug: `allowAdditionalProperties: false` dropped struct inheritance from emitted schemas. Upstream fixed and merged it the same day, but only on `main`<br>No compatibility-diff tooling and no way for a consumer to verify output<br>Rust build-time dependency; release binaries are unreachable from restricted networks<br>Generated Go carries `json:` tags only, with no YAML or TOML tags | N/A (selected) |

## Consequences

**Positive:**

- One definition produces the three artifacts that previously disagreed.
  The disagreement described in Context becomes mechanically impossible
  instead of a reviewer's responsibility.
- `specs check` checks every document's frontmatter against the generated
  schema on each push. It found two documents that write an unquoted
  `version`, which YAML reads as a float.
- You write a doc comment once, and it reaches three audiences: the JSON
  Schema that a tool reads, the Go doc comment that a maintainer reads,
  and the Markdown that an author reads.
- Strata's inheritance emits exactly the two-schema `allOf` composition
  that `frontmatter-standards.md` already required. The language now
  enforces a rule that hand-assembly guidance enforced before.
- The decision is cheap to reverse. The generated artifacts are committed,
  so removing Strata means freezing the last output and editing it by hand
  from then on. No migration or rewrite is needed.

**Negative:**

- **The project depends on a pre-1.0 tool with one maintainer, who also
  owns this project.** Fixes are easy to land, and independent review is
  scarce. This project will usually be the first to find a defect in
  Strata.
- **The consumer can't verify generated output.** Strata offers no way to
  assert that an emitted schema still accepts the documents it should. The
  `allowAdditionalProperties` bug that this evaluation found exited with
  status zero and emitted well-formed but wrong JSON Schema. Inspection
  caught it, not a gate. That defect is fixed, but the general exposure
  remains. Until Strata ships such a check, `specs check` is the only
  substitute, and it covers only what a standard JSON Schema validator can
  detect.
- **The schemas track Strata's `main` branch, not a release.** The fix
  landed after `0.7.0`, so `scripts/generate-schemas.sh` must run against
  a compiler built from source until the next release. The binary reports
  `strata-cli 0.7.0` either way, because the crate version didn't change.
  The version string therefore can't distinguish a fixed compiler from the
  buggy release. The build is identified by
  `~/.cache/strata-cli/built-commit`, which the session-start hook writes.
- **Building the Strata CLI requires a Rust toolchain**, in a repository
  that otherwise uses bash, Python, and Go. Release binaries exist for
  four platforms, but this environment's egress proxy blocks their
  download, so a cold container runs `cargo build --release` (about 2
  minutes). Committing the generated artifacts reduces this cost but
  doesn't remove it: only someone who *changes* a `.strata` file needs the
  compiler. For this reason, `scripts/generate-schemas.sh --check` isn't a
  CI gate as of this ADR's date, and CI can't catch a stale generated
  artifact.
- **Generated Go structs carry `json:` tags only.** Frontmatter is YAML
  and configuration is TOML, so neither deserializes directly into a
  generated struct. The CLI must parse both into a generic model and
  convert that model to JSON first. [[module-document-core]]'s Frontmatter
  component already describes this two-step path, and needs it anyway to
  keep each property's raw text for AC-SPEC-003.2.
- **No compatibility tooling exists.** Strata documents its own
  breaking-change rules precisely but enforces none of them, so a person
  must judge whether a schema change breaks a consumer.
- **This project uses little of Strata's main strength.** Strata generates
  code for eight languages, and this project needs one. The decision rests
  on the Markdown and JSON Schema targets and the schema-first direction,
  not on cross-language support. Anyone who re-evaluates the decision
  should weigh it on those grounds, not on the number of targets.

## References

- [[adr-015]] — Go as the CLI's implementation language, which fixes the
  Go-struct requirement this decision serves
- [[frd-spec-index]] — the feature these models exist for;
  AC-SPEC-009.5 constrains install-time network access
- [[contract-specs-configuration]] — the `specs.toml` surface modelled in
  `schemas/config.strata`
- [[module-document-core]] — the Frontmatter component that owns the
  YAML-to-JSON conversion step
- `schemas/README.md` — the design decisions inside the models, the
  regeneration commands, and what closing the shapes costs in emitted
  schema size
- `strata-lang/strata#289` — the upstream fix for the
  `allowAdditionalProperties` inheritance bug found during this
  evaluation
- TypeSpec JSON Schema emitter —
  https://www.npmjs.com/package/@typespec/json-schema
- TypeSpec Go emitter (Azure SDK client generator) —
  https://www.npmjs.com/package/@azure-tools/typespec-go
- Smithy code generation guide —
  https://smithy.io/2.0/guides/using-code-generation/index.html
- `software.amazon.smithy:smithy-jsonschema` —
  https://mvnrepository.com/artifact/software.amazon.smithy/smithy-jsonschema

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

- JSON Schema, to validate document frontmatter and `specs.toml`
  ([[contract-specs-configuration]]).
- Go structs, for the CLI and its modules ([[adr-015]]).
- Human-readable Markdown, shipped inside the plugin so that an author knows
  what frontmatter to write.

Today, those three forms exist only as prose in
`plugin/skills/spec/references/frontmatter-standards.md` and the templates,
and they already disagree. A prototype found 26 broken references and one
unparseable document that no reviewer had caught.

Three properties of this problem constrain the choice. The first one
eliminates most candidates:

- **There is no service.** This is document and configuration modeling: YAML
  frontmatter in Markdown files, and a TOML configuration file. There are no
  operations, no protocol, no client or server, and no wire negotiation. Most
  mature tools in this space are API description languages.
- **Humans and agents author the documents, and a tool then validates
  them.** The schema is the contract and the Go code is one consumer of it,
  so the direction must be schema-first rather than code-first.
- **The toolchain must stay thin and reachable.** The repository runs on
  bash, `python3`, and (per [[adr-015]]) Go. A tool that brings in a second
  language runtime has a real cost, and [[frd-spec-index]] AC-SPEC-009.5
  rules out fetching executable content from the network at install time.

This decision has a conflict of interest, and this ADR records it
deliberately. Strata is maintained by this repository's own owner. That
makes it easier to get fixes landed and much harder to evaluate neutrally.
It's the reason that this ADR compares options that a same-owner tool would
normally not be compared against.

## Decision

We adopt Strata as the single source of truth for the Spec CLI's
document-frontmatter and configuration data models. Strata generates JSON
Schema, Go structs, and Markdown from `.strata` sources under `schemas/`.

The adoption is scoped to data-shape definition only, and it's reversible.
The generated artifacts are committed, so the repository stays buildable and
correct if Strata is removed.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **TypeSpec** (Microsoft) | Its JSON Schema emitter is GA and proven at vendor scale; it emits 2020-12 with `$defs` bundling.<br>It can emit standalone models without declaring a service.<br>A large vendor maintains it, with an active 1.0 line. | It requires a Node and npm runtime that the repository doesn't otherwise need.<br>Go generation is `@azure-tools/typespec-go`, an Azure SDK **client** emitter that produces `models.go` and `service.go` with client and tracing options. It isn't a plain-struct emitter.<br>The standard library has no Markdown documentation emitter. | TypeSpec doesn't produce the two artifacts that this project most needs: plain Go structs and human-readable Markdown. Its JSON Schema emitter is **more mature** than Strata's. TypeSpec is GA, and Strata is a pre-1.0 tool in which this evaluation found a correctness bug and got it fixed. However, TypeSpec isn't more **capable**. Strata was tested at `schemaVersion: Draft202012`, matches TypeSpec on 2020-12 output and `$defs` bundling, and adds root filtering and an OpenAI Structured Outputs dialect. On the schema half of the comparison, the two differ in reliability, not in features. |
| **Smithy** (AWS) | It's the most mature option by a wide margin; its Go code generation powers `aws-sdk-go-v2`.<br>It's protocol-agnostic, with a JSON AST and a strong validation and traits model.<br>Documentation generation is a stated first-class capability. | It needs a Java toolchain: a Gradle plugin, or the JVM Smithy CLI plus `smithy-build.json`.<br>It's designed around service and operation shapes; `smithy-jsonschema` places its output in the context of a service shape ID.<br>Its Go code generation is shaped for SDK clients, not plain data models. | It has the wrong shape and the heaviest toolchain. Modeling Markdown frontmatter as a Smithy service to get a JSON Schema inverts the tool's purpose, and a JVM and Gradle dependency costs more than the Rust dependency that it would replace. |
| **Go structs as the source, with JSON Schema by reflection** (for example, `invopop/jsonschema`) | It adds no new language runtime, because the CLI is already Go.<br>The Go half has no generated-artifact drift, because Go *is* the source.<br>It has the smallest possible toolchain. | It inverts the direction: the schema becomes a projection of one consumer's implementation.<br>It yields no Markdown artifact, so the plugin's shipped frontmatter reference stays hand-maintained and drifts.<br>Struct tags and reflection express constraints much less directly than a schema language. | This is the closest call, and it would win if Go were the only output. It loses on the third deliverable. The plugin ships documentation of the frontmatter shape, and this option leaves that documentation hand-written, which is the drift that this decision exists to end. |
| **Hand-author JSON Schema, Go structs, and Markdown separately** | It needs no tooling, no build step, and no new dependency.<br>It gives full control over every artifact. | Three artifacts stay in agreement only through discipline.<br>This is the status quo that already produced the drift described in the Context section. | The evidence rules it out. The prose-only version of these shapes is already inconsistent, and two more hand-maintained artifacts make that worse. |
| **Strata** | It generates all three needed artifacts from one source: JSON Schema, plain Go structs, and Markdown.<br>Struct inheritance emits the `allOf` + `$ref` shape that `frontmatter-standards.md` independently specified.<br>Doc blocks become JSON Schema `description` values, Go doc comments, and Markdown tables from one text.<br>It's a single static binary at runtime, and its diagnostics name the fix. | It's pre-1.0 (`0.7.0`), with a thin ecosystem and effectively a bus factor of one.<br>The evaluation found a silent correctness bug: `allowAdditionalProperties: false` dropped struct inheritance from emitted schemas. The fix was merged upstream the same day, but only on `main`.<br>It has no compatibility-diff tooling and no consumer-side output verification.<br>It has a Rust build-time dependency, and its release binaries are unreachable from locked-down networks.<br>Generated Go carries `json:` tags only, with no YAML or TOML tags. | N/A: selected. |

## Consequences

**Positive:**

- One definition produces the three artifacts that previously disagreed. The
  disagreement described in the Context section becomes mechanically
  impossible rather than a review responsibility.
- On each push, `specs check` now checks every document's frontmatter
  against the generated schema. It found two documents that wrote an
  unquoted `version` that YAML reads as a float.
- A doc comment written once reaches three audiences: the JSON Schema that a
  tool reads, the Go doc comment that a maintainer reads, and the Markdown
  that an author reads.
- Strata's inheritance emits exactly the two-schema `allOf` composition that
  `frontmatter-standards.md` already required. The language now enforces a
  rule that hand-assembly guidance previously enforced.
- The decision is cheap to reverse. The generated artifacts are committed,
  so removing Strata means freezing the current output and hand-editing it
  afterward, with no migration and no rewrite.

**Negative:**

- **The project depends on a pre-1.0 tool with one maintainer, who is also
  this project's owner.** Fixes are easy to land, and independent scrutiny
  is scarce. A defect found here will usually be found here first.
- **The consumer can't verify the generated output.** Strata offers no way
  to assert that an emitted schema still accepts the documents that it
  should. The `allowAdditionalProperties` bug found during this evaluation
  exited zero and emitted well-formed, wrong JSON Schema. Inspection caught
  it, not a gate. That specific defect is fixed, but the general exposure
  remains. Until Strata ships such a check, `specs check` is the only
  substitute for it, and it covers only what a stock JSON Schema validator
  can see.
- **The schemas track Strata's `main` branch, not a release.** The fix
  landed after `0.7.0`, so `scripts/generate-schemas.sh` must run against a
  compiler built from source until the next release. The binary reports
  `strata-cli 0.7.0` either way, because the crate version didn't change, so
  the version string can't distinguish a fixed compiler from the buggy
  release. The build is identified by `~/.cache/strata-cli/built-commit`,
  which the session-start hook writes.
- **Building the Strata CLI requires a Rust toolchain**, in a repository
  that is otherwise bash, Python, and Go. Release binaries exist for four
  platforms, but their bytes are unreachable through this environment's
  egress proxy, so a cold container pays for a `cargo build --release`
  (about 2 minutes). Committing the generated artifacts mitigates this cost
  but doesn't remove it: only someone who *changes* a `.strata` file needs
  the compiler. For this reason, `scripts/generate-schemas.sh --check` isn't
  yet a CI gate, and CI can't currently catch a stale generated artifact.
- **Generated Go structs carry `json:` tags only.** Frontmatter is YAML and
  configuration is TOML, so neither deserializes directly into a generated
  struct. Both must be parsed to a generic model and converted to JSON
  first. This is the two-step path that the Frontmatter component of
  [[module-document-core]] already describes, and that component needs the
  path anyway to retain each property's raw text for AC-SPEC-003.2.
- **No compatibility tooling exists.** Strata documents its own
  breaking-change rules precisely but enforces none of them, so the question
  "did this schema change break a consumer?" stays a human judgment.
- **This project uses little of Strata's main strength.** Strata generates
  code for eight languages, and this project needs one. The decision rests
  on the Markdown and JSON Schema targets and on the schema-first direction,
  not on cross-language support. If you re-evaluate this decision, weigh it
  on those grounds, not on the number of targets.

## References

- [[adr-015]]: Go as the CLI's implementation language, which sets the
  Go-struct requirement that this decision serves.
- [[frd-spec-index]]: the feature that these models exist for.
  AC-SPEC-009.5 constrains network access at install time.
- [[contract-specs-configuration]]: the `specs.toml` surface that
  `schemas/config.strata` models.
- [[module-document-core]]: the Frontmatter component that owns the
  YAML-to-JSON conversion step.
- `schemas/README.md`: the design decisions inside the models, the
  regeneration commands, and what closing the shapes costs in emitted schema
  size.
- `strata-lang/strata#289`: the upstream fix for the
  `allowAdditionalProperties` inheritance bug found during this evaluation.
- TypeSpec JSON Schema emitter:
  https://www.npmjs.com/package/@typespec/json-schema
- TypeSpec Go emitter (Azure SDK client generator):
  https://www.npmjs.com/package/@azure-tools/typespec-go
- Smithy code generation guide:
  https://smithy.io/2.0/guides/using-code-generation/index.html
- `software.amazon.smithy:smithy-jsonschema`:
  https://mvnrepository.com/artifact/software.amazon.smithy/smithy-jsonschema

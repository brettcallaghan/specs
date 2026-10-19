---
title: "ADR-015: Go as the Implementation Language for the Spec Corpus CLI"
description: Go meets every stated requirement, ties Rust on the measures
  that discriminate, and is the only candidate with an official
  in-process Gherkin parser — which decides it
slug: adr-015
type: adr
status: approved
---

# ADR-015: Go as the implementation language for the spec corpus CLI

**Status:** approved  
**Date:** 2026-07-31

## Context

[[frd-spec-index]] states requirements rather than technology and defers this
choice. FEAT-SPEC-009 binds the following:

- A corpus query returns within 50 ms at p95 against a warm index.
- A single-document check returns within 100 ms.
- A cold 1,000-document index builds within 1 second.
- The tool runs "on a machine with no language runtime, interpreter, or
  package manager installed beyond the CLI itself" and never retrieves
  executable content at install time.

The tool's work is narrow and repetitive: split frontmatter, parse YAML,
validate against a JSON Schema that the active profile supplies at runtime,
walk a Markdown AST for headings and their spans, extract fenced and comment
blocks, resolve a reference graph, and emit versioned JSON. There's no hot
numerical loop; the performance requirement is about process startup and I/O.

We evaluated Rust, Go, Python, and TypeScript. We installed all four
toolchains and measured every claim below in this environment rather than
taking it from documentation.

### Measured: startup and distributable artifact

The figure that matters is the startup time of the *shipped artifact*. Each
figure is the median of 15 to 20 runs:

| Candidate | Startup | Artifact |
|-----------|---------|----------|
| Go, bare binary | **1.72 ms** | 2.2 MB |
| Rust, bare binary | **1.77 ms** | 1.0 MB |
| Bun, `bun -e ''` (runtime resident) | 2.47 ms | n/a |
| Node, `node -e ''` (before any module) | 29.68 ms | n/a |
| Python + `yaml,json,re` | 33.30 ms | n/a |
| Bun, `bun build --compile` | **32.29 ms** | **95 MB** |

Two results decided which candidates survive, and one corrected an earlier
draft of this ADR.

**Bun doesn't survive compilation.** Its resident startup is near-native,
which initially made TypeScript look viable. Compiling to the standalone
executable that FEAT-SPEC-009.4 requires gives back that entire advantage: the
compiled binary starts in 32.29 ms, 13 times the resident figure, because the
embedded runtime still bootstraps. A 50 ms p95 budget with 32 ms of fixed
overhead leaves 18 ms for the work, in a 95 MB artifact. Python and Node fail
FEAT-SPEC-009.4 outright without bundling, and bundling reproduces this
result.

**Go and Rust are tied on startup.** The difference between 1.72 ms and
1.77 ms is noise. A first draft of this ADR reported Go at 4.55 ms and treated
that as a Rust advantage. That measurement included two Gherkin parses, and
the draft compared it against a Rust binary that only rendered `--help`. After
the correction, the difference vanishes. Both candidates clear the requirement
by more than an order of magnitude.

### Measured: Gherkin

The acceptance-criteria notation was undecided when this decision was made,
and this decision had to keep Gherkin viable. Adopting Gherkin would mean that
it *replaces* EARS for acceptance criteria. [[adr-021]] has since adopted it,
which exercises the capability measured below. We parsed the same input with
the official Python and Go implementations and with `cucumber-rs/gherkin`, the
third-party Rust crate:

| | Tags on a valid feature | Error for `Wen` instead of `When` (defect on line 5) |
|---|---|---|
| Go (official) | `[@frd-checkout @AC-CHK-003.1 @req-CHK-003]`: Feature tag inherited | `(5:5): expected: … #TagLine, #ExamplesLine … got '    Wen b'` |
| Python (official) | `['@frd-checkout', '@AC-CHK-003.1', '@req-CHK-003']`: inherited | `(5:5): expected: #EOF, #TableRow, … #TagLine, #Examples` |
| Rust (third-party) | `["AC-CHK-003.1", "req-CHK-003"]`: **Feature tag missing**, `@` stripped | `Error at 6:1: {"unknown keyword", "_"}` |

All three divergences count against Rust:

- Tag inheritance doesn't happen, so a document-level tag wouldn't reach the
  criteria beneath it. Inheritance is the exact mechanism that makes tags
  attractive as carriers of AC identifiers.
- The error reports line 6 for a line-5 defect.
- The error carries no list of expected tokens. Against a stated need for
  errors that an agent can act on, `{"unknown keyword", "_"}` is materially
  worse than the official output.

In practice, a Rust implementation that adopted Gherkin would need a second
tool: either upstream contributions to a crate outside the shared test suite,
or a subprocess that calls the official CLI. Every other candidate imports an
official parser in-process.

### Measured: JSON Schema errors

The requirement is a JSON Pointer path, a human-readable description, and
**every** error rather than only the first, so that an agent can fix all of
them in one pass. We validated a frontmatter record with five simultaneous
defects against each candidate. The defects were a bad pattern, a wrong
`const`, an out-of-vocabulary `enum`, a malformed `date-time`, and a missing
required property:

```
Go — santhosh-tekuri/jsonschema v6
  - at '': missing property 'description'
  - at '/slug': 'adr-011-targeted-verification-gates' does not match pattern '^adr-[0-9]{3}$'
  - at '/type': value must be 'adr'
  - at '/status': value must be one of 'draft', 'ready-for-review', …
  - at '/updated_at': '2026-07-29' is not valid date-time: less than 20 characters long

Rust — boon 0.6
  (identical wording; same author)

Rust — jsonschema 0.26
  [1] at '/slug': "adr-011-…" does not match "^adr-[0-9]{3}$"
  (reports 4 of 5 — format included, required-property omitted)
```

All candidates report every error with a JSON Pointer. Go's `BasicOutput()`
and boon's `detailed_output()` both emit structured JSON that carries
`instanceLocation` and `keywordLocation`. This dimension is a tie, and
literally so: `boon` and `santhosh-tekuri/jsonschema` share an author and
produce the same message text.

### Measured: is YAML support adequate?

YAML is the frontmatter format, so it can't be a weak spot. `serde_yaml` was
archived in 2024, which raised the question. We tested `serde_yaml_ng` against
this repository's real corpus:

- **52 of 53 documents with frontmatter parsed.** The single failure is the
  known defect in `work-order-template.md`, which the parser correctly locates
  at line 28, column 23.
- Errors carry line, column, and context: `found unexpected end of stream at
  line 3 column 1, while scanning a quoted scalar at line 1 column 8`.
- The crate follows the YAML 1.2 core schema, so `updated_at: 2026-07-29`
  stays `String("2026-07-29")` rather than becoming a date object. That
  directly serves AC-SPEC-003.2, which requires validating the raw text.
  PyYAML's type coercion produced 42 false positives in an earlier prototype.

So the answer is yes: Rust's YAML support is adequate, and its YAML 1.2
semantics actively help. The remaining concern is maintenance, not capability.
The ecosystem churned through the archival of `serde_yaml` and through a fork,
`serde_yml`, that was itself deprecated over unsoundness. Go's maintenance
story is currently steadier. The YAML organization took over maintenance after
the original library was marked unmaintained in 2025, with `go.yaml.in/yaml/v4`
as the forward path, and `goccy/go-yaml` passes more of the YAML test suite
than `yaml.v3` did.

On capability, this dimension is close to a tie; on maintenance stability, Go
is ahead.

### On the map-then-deserialize step

An earlier draft argued that Rust's `serde` was decisive because the pipeline
ends by converting a generic map into a typed record, and Go has no
equivalent. That argument doesn't survive scrutiny.

Validation and deserialization are separable. The profile supplies JSON
Schemas at runtime (FEAT-SPEC-008), so schema validation must operate on a
generic document model in *any* language. But nothing requires the typed
record to be derived *from* that model. A second, direct unmarshal from the
same source text is equally valid, and it's what Go would do. The convenience
that `serde` provides is real but small, and it doesn't discriminate between
two languages that both meet every stated requirement.

### Library survey by dimension

This survey is assessed from documentation. Entries marked ✔ were exercised
directly.

| Dimension | Go | Rust | Python | TypeScript |
|-----------|-----|------|--------|------------|
| CLI harness / `--help` | **cobra**: powers `kubectl`, `gh`, and `docker` | **clap** ✔: doc comments become help text | **typer** | **commander** / **oclif** |
| YAML | **goccy/go-yaml**; `go.yaml.in/yaml/v4` maintained by the YAML organization | `serde_yaml_ng` ✔: works; ecosystem churned | PyYAML / ruamel: coerces types | `yaml` (eemeli): exposes a CST |
| JSON / TOML | stdlib; **BurntSushi/toml** | **serde_json** / **toml** | stdlib; **tomllib** stdlib | stdlib; `smol-toml` |
| JSON Schema | **santhosh-tekuri/jsonschema** ✔ | **boon** ✔ (same author) | `jsonschema` | **ajv** |
| Markdown AST + positions | **goldmark**: powers Hugo | **markdown-rs**: by remark's author | markdown-it-py | **remark / mdast**: strongest ecosystem |
| Gherkin | **official, in-process** ✔ | ⚠ third-party, diverges ✔ | **official** ✔ | **official** |
| Filesystem watching | **fsnotify** | **notify** | watchdog | chokidar |
| Git | **go-git** (pure Go) | **gix** / **git2** | pygit2 / GitPython | isomorphic-git |
| Distribution | Trivial cross-compile; the simplest of the four | **cargo-dist** → npm, PyPI, Homebrew, shell | Needs a runtime | Needs a runtime or a 95 MB bundle |

Shelling out to `git` is a reasonable alternative to any Git library and costs
the same in every language, so Git support doesn't discriminate.

### Ecosystem signal

- **Toward Go:** Beads, this framework's issue tracker, is written in Go, so
  contributors and agents who work in this repository already use a Go
  toolchain. Go's cross-compilation is the simplest distribution story of the
  four.
- **Toward Rust:** most of the prominent recent developer tooling is written
  in Rust, such as Astral's `uv`, `ruff`, and `ty`. `uv` reports over 126
  million downloads a month, and it's also the closest analogue to this tool's
  distribution problem.
- **Toward Python and TypeScript:** Anthropic's Agent SDK ships only in those
  two languages. That would dominate if this tool were a library that other
  agents import. It's a CLI that callers invoke as a subprocess, so the calling
  language doesn't need to match.

One correction to the framing that prompted this evaluation: Bun is written in
Zig, not Rust. The Rust developer-tooling examples are Astral's.

## Decision

We will implement the spec corpus CLI in Go, with `cobra` for the command
surface, `goccy/go-yaml` for YAML, `santhosh-tekuri/jsonschema` for JSON
Schema, `goldmark` for the Markdown AST, and the official `cucumber/gherkin` Go
implementation WHERE Gherkin support is later added.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **Rust** | 1.0 MB artifact, the smallest of the four<br>`serde` makes typed deserialization a single line<br>`clap` derives `--help` from doc comments, so help can't drift<br>`markdown-rs` is by remark's author<br>`cargo-dist` covers distribution through mirrorable registries<br>The strongest current signal in developer tooling | Gherkin support is third-party and measurably diverges: no tag inheritance, wrong line number, and opaque errors. Adoption would require a second tool<br>YAML support is a maintained fork of an archived crate; this corner has churned twice<br>An agent can't patch a compiled binary mid-session<br>Slowest of the four to a first working version | The closest call, and the first draft of this ADR selected it. Measurement reversed it: startup is a tie (1.77 ms versus 1.72 ms), JSON Schema is a tie (same author, identical messages), and the `serde` advantage doesn't survive separating validation from deserialization. What remains is a Gherkin liability with nothing left to outweigh it |
| **TypeScript** (Bun or Node) | `remark`/`mdast` is the best Markdown AST ecosystem anywhere<br>`ajv` is the best JSON Schema validator anywhere<br>Official Gherkin<br>Anthropic Agent SDK language<br>An agent can patch it mid-session | The standalone executable starts in 32.29 ms and weighs 95 MB, measured<br>Node without bundling takes 29.68 ms before it loads a module | Fails FEAT-SPEC-009.1 in practice: 32 ms of fixed overhead against a 50 ms p95 budget. A 95 MB artifact is a poor fit for a tool that's meant to sit in every adopter's CI |
| **Python** | Official Gherkin<br>Strongest parsing libraries overall; `tomllib` in the stdlib<br>Anthropic Agent SDK language<br>An agent can patch it mid-session<br>Fastest to a working prototype | 33.30 ms of fixed startup before any work; roughly 1 second to parse 1,000 documents<br>Requires an interpreter and dependencies on the adopter's machine<br>PyYAML coerces scalars, which AC-SPEC-003.2 must work around | Fails FEAT-SPEC-009.4 without bundling, and fails FEAT-SPEC-009.1 with or without it |
| **Go** | The only candidate that ties on every discriminating measure and wins the one that separates them: official in-process Gherkin, with better errors than the Rust crate<br>1.72 ms startup, 2.2 MB artifact<br>JSON Schema output identical to Rust's<br>Steadiest YAML maintenance story<br>Simplest cross-compilation of the four<br>Adjacent to Beads, already a project dependency<br>`cobra` is the most battle-tested CLI harness in any language | No sum types, so a Result-shaped pipeline carries `if err != nil` at each step<br>Reflection-based JSON is a looser guarantee than a derive macro for an index that is a published API<br>Larger artifact than Rust, and an agent can't patch a compiled binary | N/A: selected |

## Consequences

**Positive:**

- Gherkin is an import rather than a second tool. That kept the option
  genuinely open, and [[adr-021]] has since exercised it.
- The tool meets the latency and artifact requirements with more than an order
  of magnitude of margin, so a write-time validation hook stays imperceptible
  as corpora grow.
- Cross-compilation is a build flag, so producing the platform matrix that
  FEAT-SPEC-009.6 needs requires no extra machinery.
- Contributors, human or agent, use one toolchain for both this tool and
  Beads.

**Negative:**

- Every step of the composable pipeline carries explicit error handling, so
  the code is more verbose than the equivalent Rust and the happy path is
  harder to read at a glance. Libraries mitigate this less than they appear
  to; see the next section.
- The index is marshalled by reflection rather than from derived types, so a
  schema mistake surfaces at runtime in a consumer rather than at compile
  time. To mitigate this, treat the index schema as a tested artifact, not
  only a struct.
- The artifact is roughly twice the size of Rust's. This has no practical
  consequence: the tool runs in CI, in developer checkouts, and in agent
  sessions, and 2.2 MB is immaterial in all three.
- An agent that debugs the CLI mid-session can't edit and re-run it as it
  could a script. Keeping the domain layer in configuration, which stays
  editable, partly offsets this.

### On mitigating Go's error-handling verbosity

We considered three libraries:

- `samber/mo`: Option, Result, and Either monads.
- `samber/lo`: its `*Err` helper variants short-circuit on the first error.
- `samber/oops`: structured errors that carry contextual attributes and stack
  traces.

We attempted a `Result`-shaped rewrite of the pipeline with `samber/mo`, and
it **doesn't compile**:

```
cannot use ….FlatMap(…) (value of struct type mo.Result[string])
  as mo.Result[Doc] value in return statement
```

The cause is a Go language rule, not a library gap: methods can't introduce
type parameters, so `Result[A].FlatMap` is constrained to return `Result[A]`.
A pipeline that changes shape (text → map → typed record) can't be chained
fluently. A free function works:

```go
func bind[A, B any](r mo.Result[A], f func(A) mo.Result[B]) mo.Result[B]
```

However, calls to it nest rather than chain, and the result reads worse than
the `if err != nil` version it replaces. It also discards Go's `%w` error
wrapping. The idiomatic version yields `schema: at '/type': value must be
'adr'`, whereas the monadic version yields the bare `at '/type': value must be
'adr'` unless you re-add the context by hand.

`samber/oops` is the one worth adopting, for a different reason. It doesn't
reduce the number of error checks; it makes each error carry structured
context, which a finding needs anyway under AC-SPEC-006.1. Whether to take
that dependency belongs in the container blueprint, not here.

**Revisit when** Gherkin is definitively rejected for acceptance criteria,
which removes the advantage that decided this, *and* a measured need appears
that Go can't meet.

## References

- [[frd-spec-index]]: FEAT-SPEC-009, the requirements that this decision is
  measured against.
- [[adr-021]]: adopts Gherkin for acceptance criteria, which exercises the
  in-process parser that this decision preserved.
- `docs/plans/spec-validation-cli-brief.md`: earlier measurements and the
  distribution research.
- [cucumber/gherkin](https://github.com/cucumber/gherkin): the shared test
  suite that defines an official implementation.
- [cucumber-rs/gherkin](https://github.com/cucumber-rs/gherkin): the
  third-party Rust crate measured in this ADR.
- [santhosh-tekuri/jsonschema](https://github.com/santhosh-tekuri/jsonschema)
  and [boon](https://github.com/santhosh-tekuri/boon): the Go and Rust
  validators, by the same author.
- [yaml/go-yaml](https://github.com/yaml/go-yaml): the YAML organization's
  takeover of Go YAML maintenance.
- [goccy/go-yaml](https://github.com/goccy/go-yaml): the selected YAML
  implementation.
- [serde-yaml deprecation](https://users.rust-lang.org/t/serde-yaml-deprecation-alternatives/108868)

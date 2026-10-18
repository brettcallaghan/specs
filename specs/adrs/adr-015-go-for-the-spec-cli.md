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

[[frd-spec-index]] deliberately states requirements, not technology, and
defers this choice. FEAT-SPEC-009 binds these requirements:

- A corpus query returns within 50 ms at p95 against a warm index.
- A single-document check returns within 100 ms.
- A cold 1,000-document index finishes within 1 second.
- The tool runs "on a machine with no language runtime, interpreter, or
  package manager installed beyond the CLI itself."
- The tool never retrieves executable content at install time.

The tool's work is narrow and repetitive. It does these steps:

- Split frontmatter.
- Parse YAML.
- Validate against a JSON Schema that the active profile supplies at
  runtime.
- Walk a Markdown AST for headings and their spans.
- Extract fenced and comment blocks.
- Resolve a reference graph.
- Emit versioned JSON.

The tool has no hot numerical loop. The performance requirement is about
process startup and I/O.

We evaluated Rust, Go, Python, and TypeScript. We installed all four
toolchains and measured every claim below in this environment instead of
taking it from documentation.

### Measured: startup and distributable artifact

The figure that matters is the startup time of the *shipped artifact*. Each
figure is the median of 15 to 20 runs:

| Candidate | Startup | Artifact |
|-----------|---------|----------|
| Go, bare binary | 1.72 ms | 2.2 MB |
| Rust, bare binary | 1.77 ms | 1.0 MB |
| Bun, `bun -e ''` (runtime resident) | 2.47 ms | n/a |
| Node, `node -e ''` (before any module) | 29.68 ms | n/a |
| Python + `yaml,json,re` | 33.30 ms | n/a |
| Bun, `bun build --compile` | 32.29 ms | 95 MB |

Two results decided which candidates survive. One of them corrected an
earlier draft of this ADR.

Bun doesn't survive compilation. Its resident startup is near native, which
at first made TypeScript look viable. FEAT-SPEC-009.4 requires a standalone
executable, and compiling to one costs back that whole advantage. The
compiled binary starts in 32.29 ms, thirteen times the resident figure,
because the embedded runtime still bootstraps. A 50 ms p95 budget with 32 ms
of fixed overhead leaves 18 ms for the work, in a 95 MB artifact. Python and
Node fail FEAT-SPEC-009.4 outright without bundling, and bundling reproduces
this result.

Go and Rust tie on startup. The gap between 1.72 ms and 1.77 ms is noise. A
first draft of this ADR reported Go at 4.55 ms and treated that as a Rust
advantage. That measurement included two Gherkin parses, and it was compared
against a Rust binary that only rendered `--help`. With the comparison
corrected, the difference disappears. Both clear the requirement by more
than an order of magnitude.

### Measured: Gherkin

The acceptance-criteria notation was undecided when this decision was made,
and this decision had to keep Gherkin viable. Adopting Gherkin would mean
that it *replaces* EARS for acceptance criteria. [[adr-021]] has since
adopted it, which exercises the capability measured below. We parsed the
same input with the official Python and Go implementations and with
`cucumber-rs/gherkin`, the third-party Rust crate:

| | Tags on a valid feature | Error for `Wen` instead of `When` (defect on line 5) |
|---|---|---|
| Go (official) | `[@frd-checkout @AC-CHK-003.1 @req-CHK-003]` (Feature tag inherited) | `(5:5): expected: … #TagLine, #ExamplesLine … got '    Wen b'` |
| Python (official) | `['@frd-checkout', '@AC-CHK-003.1', '@req-CHK-003']` (inherited) | `(5:5): expected: #EOF, #TableRow, … #TagLine, #Examples` |
| Rust (third-party) | `["AC-CHK-003.1", "req-CHK-003"]` (Feature tag missing, `@` stripped) | `Error at 6:1: {"unknown keyword", "_"}` |

The Rust crate diverges from the official parsers in three ways. It doesn't
apply tag inheritance, so a document-level tag wouldn't reach the criteria
beneath it. Inheritance is the exact mechanism that makes tags attractive
for carrying AC identifiers. Its error reports line 6 for a line-5 defect
and carries no list of expected tokens. Against a stated need for errors an
agent can act on, `{"unknown keyword", "_"}` is materially worse than the
official output.

As a result, a Rust implementation that adopts Gherkin would need a second
tool: either upstream contributions to a crate outside the shared test
suite, or a subprocess that calls the official CLI. Every other candidate
imports an official parser in-process.

### Measured: JSON Schema errors

The requirement is a JSON Pointer path, a human-readable description, and
every error, not only the first, so that an agent fixes them in one pass. We
validated a frontmatter record with five simultaneous defects against each
candidate. The defects were a bad pattern, a wrong `const`, an
out-of-vocabulary `enum`, a malformed `date-time`, and a missing required
property. Each candidate produced this output:

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

All of them report every error with a JSON Pointer. Go's `BasicOutput()` and
boon's `detailed_output()` both emit structured JSON that carries
`instanceLocation` and `keywordLocation`. This dimension is a literal tie:
`boon` and `santhosh-tekuri/jsonschema` share an author and produce the same
message text.

### Measured: is YAML support adequate?

YAML is the frontmatter format, so YAML support can't be a weak spot. The
archiving of `serde_yaml` in 2024 raised the question. We tested
`serde_yaml_ng` against this repository's real corpus, with these results:

- 52 of 53 documents with frontmatter parsed. The one failure is the known
  defect in `work-order-template.md`, which the parser correctly located at
  line 28, column 23.
- Errors carry line and column, including context: `found unexpected end
  of stream at line 3 column 1, while scanning a quoted scalar at line 1
  column 8`.
- It follows the YAML 1.2 core schema, so `updated_at: 2026-07-29` stays
  `String("2026-07-29")` instead of becoming a date object. That directly
  serves AC-SPEC-003.2, which requires validating the raw text. In an
  earlier prototype, PyYAML's coercion produced 42 false positives.

The answer is yes: Rust's YAML support is adequate, and its 1.2 semantics
actively help. The remaining concern is maintenance, not capability. The
ecosystem churned through the archiving of `serde_yaml` and through a fork,
`serde_yml`, that was itself deprecated over unsoundness. Go's YAML
maintenance is currently steadier. The YAML org took over maintenance after
the original library was marked unmaintained in 2025, and
`go.yaml.in/yaml/v4` is the path forward. Also, `goccy/go-yaml` passes more
of the YAML test suite than `yaml.v3` did.

This dimension is close to a tie on capability, with Go ahead on
maintenance stability.

### On the map-then-deserialize step

An earlier draft argued that Rust's `serde` was decisive. Its reasoning was
that the pipeline ends by converting a generic map into a typed record, and
Go has no equivalent. That argument doesn't hold.

Validation and deserialization are separable. The profile supplies JSON
Schemas at runtime (FEAT-SPEC-008), so schema validation must operate on a
generic document model in *any* language. Nothing requires the typed record
to be derived *from* that model, though. A second, direct unmarshal from the
same source text is equally valid, and it's what Go would do. The
convenience that `serde` provides is real but small. It doesn't discriminate
between two languages that both meet every stated requirement.

### Library survey by dimension

We assessed these libraries from documentation, and we exercised the ones
marked ✔ directly.

| Dimension | Go | Rust | Python | TypeScript |
|-----------|-----|------|--------|------------|
| CLI harness / `--help` | cobra (powers `kubectl`, `gh`, `docker`) | clap ✔ (doc comments become help text) | typer | commander / oclif |
| YAML | goccy/go-yaml; `go.yaml.in/yaml/v4` org-maintained | `serde_yaml_ng` ✔ (works; ecosystem churned) | PyYAML / ruamel (coerce types) | `yaml` (eemeli), which exposes a CST |
| JSON / TOML | stdlib; BurntSushi/toml | serde_json / toml | stdlib; tomllib stdlib | stdlib; `smol-toml` |
| JSON Schema | santhosh-tekuri/jsonschema ✔ | boon ✔ (same author) | `jsonschema` | ajv |
| Markdown AST + positions | goldmark (powers Hugo) | markdown-rs (by remark's author) | markdown-it-py | remark / mdast (strongest ecosystem) |
| Gherkin | official, in-process ✔ | ⚠ third-party, diverges ✔ | official ✔ | official |
| Filesystem watching | fsnotify | notify | watchdog | chokidar |
| Git | go-git (pure Go) | gix / git2 | pygit2 / GitPython | isomorphic-git |
| Distribution | Trivial cross-compile, the simplest of the four | cargo-dist, for npm, PyPI, Homebrew, and shell | needs a runtime | needs a runtime or a 95 MB bundle |

Shelling out to `git` is a reasonable alternative to any git library, and it
costs the same in every language, so git doesn't discriminate.

### Ecosystem signal

- Toward Go: Beads, this framework's issue tracker, is written in Go, so
  contributors and agents working in this repository already meet a Go
  toolchain. Go's cross-compilation is the simplest distribution story of
  the four.
- Toward Rust: the dominant recent developer tooling is written in Rust:
  Astral's `uv`, `ruff`, and `ty`. `uv` reports over 126 million downloads a
  month, and it's also the closest analog to this tool's distribution
  problem.
- Toward Python and TypeScript: Anthropic's Agent SDK ships only in those
  two languages. That would dominate if this tool were a library that other
  agents import. It's a CLI invoked as a subprocess, so the calling language
  doesn't need to match.

One correction to the framing that prompted this evaluation: Bun is written
in Zig, not Rust. The Rust developer-tooling examples are Astral's.

## Decision

We will implement the spec corpus CLI in Go, with these libraries:

- `cobra` for the command surface
- `goccy/go-yaml` for YAML
- `santhosh-tekuri/jsonschema` for JSON Schema
- `goldmark` for the Markdown AST
- The official `cucumber/gherkin` Go implementation, where Gherkin support
  is added later

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Rust | 1.0 MB artifact, the smallest of the four<br>`serde` makes typed deserialization a single line<br>`clap` derives `--help` from doc comments, so help can't drift<br>`markdown-rs` is by remark's author<br>`cargo-dist` covers mirrorable-registry distribution<br>The strongest current signal in developer tooling | Gherkin is third-party and measurably diverges: no tag inheritance, wrong line number, and opaque errors. Adoption would require a second tool.<br>YAML support is a maintained fork of an archived crate, and this corner of the ecosystem has churned twice.<br>An agent can't patch a compiled binary mid-session.<br>Slowest of the four to a first working version | The closest call, and the first draft of this ADR selected it. Reversed on measurement: startup is a tie (1.77 ms versus 1.72 ms), JSON Schema is a tie (same author, identical messages), and the `serde` advantage doesn't survive separating validation from deserialization. What remains is a Gherkin liability with nothing left to outweigh it. |
| TypeScript (Bun or Node) | `remark`/`mdast` is the best Markdown AST ecosystem anywhere<br>`ajv` is the best JSON Schema validator anywhere<br>Official Gherkin<br>Anthropic Agent SDK language<br>An agent can patch it mid-session | The standalone executable starts in 32.29 ms and weighs 95 MB, measured.<br>Node without bundling takes 29.68 ms before it loads a module. | Fails FEAT-SPEC-009.1 in practice: 32 ms of fixed overhead against a 50 ms p95 budget. A 95 MB artifact is a poor fit for a tool meant to sit in every adopter's CI. |
| Python | Official Gherkin<br>Strongest parsing libraries overall, with `tomllib` in the stdlib<br>Anthropic Agent SDK language<br>An agent can patch it mid-session<br>Fastest to a working prototype | 33.30 ms fixed startup before any work, and roughly 1 s to parse 1,000 documents<br>Requires an interpreter and dependencies on the adopter's machine<br>PyYAML coerces scalars, which AC-SPEC-003.2 must work around | Fails FEAT-SPEC-009.4 without bundling and FEAT-SPEC-009.1 with or without it |
| Go | The only candidate that ties on every discriminating measure and wins the one that separates them: official in-process Gherkin, with better errors than the Rust crate<br>1.72 ms startup, 2.2 MB artifact<br>JSON Schema output identical to Rust's<br>Steadiest YAML maintenance story<br>Simplest cross-compilation of the four<br>Adjacent to Beads, already a project dependency<br>`cobra` is the most battle-tested CLI harness in any language | No sum types, so a Result-shaped pipeline carries `if err != nil` at each step<br>Reflection-based JSON is a looser guarantee than a derive macro for an index that is a published API<br>Larger artifact than Rust, and an agent can't patch a compiled binary | N/A (selected) |

## Consequences

**Positive:**
- Gherkin is an import, not a second tool. That kept the option open, and
  [[adr-021]] has since exercised it.
- The tool meets the latency and artifact requirements with more than an
  order of magnitude of margin, so a write-time validation hook stays
  imperceptible as corpora grow.
- Cross-compilation is a build flag, so producing the platform matrix that
  FEAT-SPEC-009.6 needs takes no extra machinery.
- Contributors, human or agent, meet one toolchain for both this tool and
  Beads.

**Negative:**
- Every step of the composable pipeline carries explicit error handling.
  The code is more verbose than the equivalent Rust, and the happy path is
  harder to read at a glance. Libraries mitigate this less than they appear
  to, as the next section explains.
- The index is marshaled by reflection instead of from derived types, so a
  schema mistake surfaces at runtime in a consumer instead of at compile
  time. Treating the index schema as a tested artifact, not only a struct,
  mitigates this.
- The artifact is roughly twice Rust's size. No consequence follows: the
  tool runs in CI, in developer checkouts, and in agent sessions, and
  2.2 MB is immaterial in all three.
- An agent debugging the CLI mid-session can't edit and rerun it as it
  could a script. Keeping the domain layer in configuration, which stays
  editable, partly offsets this.

### On mitigating Go's error-handling verbosity

We considered three libraries:

- `samber/mo`: Option, Result, and Either monads.
- `samber/lo`: its `*Err` helper variants short-circuit on the first error.
- `samber/oops`: structured errors that carry contextual attributes and
  stack traces.

We attempted a `Result`-shaped rewrite of the pipeline with `samber/mo`, and
it doesn't compile:

```
cannot use ….FlatMap(…) (value of struct type mo.Result[string])
  as mo.Result[Doc] value in return statement
```

The cause is a Go language rule, not a library gap. Methods can't introduce
type parameters, so `Result[A].FlatMap` is constrained to return
`Result[A]`. A pipeline that changes shape, from text to map to typed
record, can't be chained fluently. A free function works:

```go
func bind[A, B any](r mo.Result[A], f func(A) mo.Result[B]) mo.Result[B]
```

However, it nests instead of chaining, and the result reads worse than the
`if err != nil` version it replaces. It also discards Go's `%w` error
wrapping. The idiomatic version yields `schema: at '/type': value must
be 'adr'`. The monadic version yields the bare
`at '/type': value must be 'adr'` unless you add the context back by hand.

`samber/oops` is the one worth adopting, for a different reason. It doesn't
reduce the number of error checks. Instead, it makes each error carry
structured context, which a finding needs anyway under AC-SPEC-006.1. The
container blueprint, not this ADR, decides whether to take that dependency.

Revisit this decision when both of these happen:

- Gherkin is definitively rejected for acceptance criteria, which removes
  the advantage that decided this ADR.
- A measured need appears that Go can't meet.

## References

- [[frd-spec-index]]: FEAT-SPEC-009, the requirements this decision is
  measured against
- [[adr-021]]: adopts Gherkin for acceptance criteria, which exercises the
  in-process parser this decision preserved
- `docs/plans/spec-validation-cli-brief.md`: earlier measurements and the
  distribution research
- [cucumber/gherkin](https://github.com/cucumber/gherkin): the shared test
  suite that defines an official implementation
- [cucumber-rs/gherkin](https://github.com/cucumber-rs/gherkin): the
  third-party Rust crate measured above
- [santhosh-tekuri/jsonschema](https://github.com/santhosh-tekuri/jsonschema)
  and [boon](https://github.com/santhosh-tekuri/boon): the Go and Rust
  validators, which share an author
- [yaml/go-yaml](https://github.com/yaml/go-yaml): the YAML org's takeover
  of Go YAML maintenance
- [goccy/go-yaml](https://github.com/goccy/go-yaml): the YAML
  implementation selected
- [serde-yaml deprecation](https://users.rust-lang.org/t/serde-yaml-deprecation-alternatives/108868)

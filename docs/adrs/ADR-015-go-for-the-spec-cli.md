---
title: "ADR-015: Go as the Implementation Language for the Spec Corpus CLI"
description: Go meets every stated requirement, ties Rust on the measures
  that discriminate, and is the only candidate with an official
  in-process Gherkin parser — which decides it
slug: adr-015
type: adr
status: draft
updated_at: 2026-07-31T00:00:00Z
---

# ADR-015: Go as the Implementation Language for the Spec Corpus CLI

**Status:** draft  
**Date:** 2026-07-31

## Context

[[frd-spec-validation-cli]] deliberately states requirements rather than
technology and defers this choice. REQ-SPEC-009 binds: a corpus query
returns within 50 ms at p95 against a warm index, a single-document
check within 100 ms, a cold 1,000-document index within 1 second, and
the tool runs "on a machine with no language runtime, interpreter, or
package manager installed beyond the CLI itself" while never retrieving
executable content at install time.

The tool's work is narrow and repetitive: split frontmatter, parse YAML,
validate against a JSON Schema supplied at runtime by the active
profile, walk a markdown AST for headings and their spans, extract
fenced and comment blocks, resolve a reference graph, emit versioned
JSON. There is no hot numerical loop; the performance requirement is
about process startup and I/O.

Rust, Go, Python and TypeScript were evaluated. All four toolchains were
installed and every claim below was measured in this environment rather
than taken from documentation.

### Measured: startup and distributable artifact

The figure that matters is the *shipped artifact's* startup. Median of
15–20 runs each:

| Candidate | Startup | Artifact |
|-----------|---------|----------|
| Go, bare binary | **1.72 ms** | 2.2 MB |
| Rust, bare binary | **1.77 ms** | 1.0 MB |
| Bun, `bun -e ''` (runtime resident) | 2.47 ms | n/a |
| Node, `node -e ''` (before any module) | 29.68 ms | n/a |
| Python + `yaml,json,re` | 33.30 ms | n/a |
| Bun, `bun build --compile` | **32.29 ms** | **95 MB** |

Two results decided which candidates survive, and one corrected an
earlier draft of this ADR.

**Bun does not survive compilation.** Its resident startup is
near-native, which initially made TypeScript look viable. Compiling to
the standalone executable REQ-SPEC-009.4 actually requires costs that
back entirely — 32.29 ms, thirteen times the resident figure, because
the embedded runtime still bootstraps. A 50 ms p95 budget with 32 ms of
fixed overhead leaves 18 ms for the work, in a 95 MB artifact. Python
and Node fail REQ-SPEC-009.4 outright without bundling, and bundling
reproduces this result.

**Go and Rust are tied on startup.** 1.72 ms against 1.77 ms is noise. A
first draft of this ADR reported Go at 4.55 ms and treated that as a
Rust advantage; that measurement included two Gherkin parses and was
compared against a Rust binary merely rendering `--help`. Corrected, the
difference vanishes. Both clear the requirement by more than an order of
magnitude.

### Measured: Gherkin

[[adr-014]] defers Gherkin but requires this decision to preserve it,
and records that adopting it would mean Gherkin *replaces* EARS for
acceptance criteria. The same input was parsed with the official Python
and Go implementations and with `cucumber-rs/gherkin`, the third-party
Rust crate:

| | Tags on a valid feature | Error for `Wen` instead of `When` (defect on line 5) |
|---|---|---|
| Go (official) | `[@frd-checkout @AC-CHK-003.1 @req-CHK-003]` — Feature tag inherited | `(5:5): expected: … #TagLine, #ExamplesLine … got '    Wen b'` |
| Python (official) | `['@frd-checkout', '@AC-CHK-003.1', '@req-CHK-003']` — inherited | `(5:5): expected: #EOF, #TableRow, … #TagLine, #Examples` |
| Rust (third-party) | `["AC-CHK-003.1", "req-CHK-003"]` — **Feature tag missing**, `@` stripped | `Error at 6:1: {"unknown keyword", "_"}` |

Three divergences, all against Rust. Tag inheritance does not happen, so
a document-level tag would not reach the criteria beneath it — the exact
mechanism that makes tags attractive as AC identifier carriers. The
error reports line 6 for a line-5 defect and carries no expected-token
list; `{"unknown keyword", "_"}` is materially worse than the official
output against a stated need for errors an agent can act on.

The practical consequence is that a Rust implementation adopting Gherkin
would need a second tool — either upstream contributions to a crate
outside the shared test suite, or a subprocess to the official CLI.
Every other candidate imports an official parser in-process.

### Measured: JSON Schema errors

The requirement is a JSON Pointer path, a human-readable description,
and **every** error rather than the first, so an agent fixes them in one
pass. A frontmatter record with five simultaneous defects — bad pattern,
wrong `const`, out-of-vocabulary `enum`, malformed `date-time`, and a
missing required property — was validated with each candidate:

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

All report every error with a JSON Pointer. Go's `BasicOutput()` and
boon's `detailed_output()` both emit structured JSON carrying
`instanceLocation` and `keywordLocation`. This dimension is a tie, and
literally so: `boon` and `santhosh-tekuri/jsonschema` share an author
and produce the same message text.

### Measured: is YAML support adequate?

YAML is the frontmatter format, so this cannot be a weak spot.
`serde_yaml` was archived in 2024, which raised the question. Testing
`serde_yaml_ng` against this repository's real corpus:

- **52 of 53 documents with frontmatter parsed.** The single failure is
  the known defect in `work-order-template.md`, correctly located at
  line 28 column 23 (tracked as `agent-plugins-6n4`).
- Errors carry line and column, including context: `found unexpected end
  of stream at line 3 column 1, while scanning a quoted scalar at line 1
  column 8`.
- It follows the YAML 1.2 core schema, so `updated_at: 2026-07-29`
  stays `String("2026-07-29")` rather than becoming a date object. That
  directly serves AC-SPEC-003.2, which requires validating the raw text
  — PyYAML's coercion produced 42 false positives in an earlier
  prototype.

So the answer is yes: Rust's YAML support is adequate, and its 1.2
semantics actively help. The remaining concern is maintenance, not
capability — the ecosystem churned through `serde_yaml`'s archival and a
fork, `serde_yml`, that was itself deprecated over unsoundness. Go's
story is currently steadier: the YAML org took over maintenance after
the original was marked unmaintained in 2025, with `go.yaml.in/yaml/v4`
as the forward path, and `goccy/go-yaml` passes more of the YAML test
suite than `yaml.v3` did.

This dimension is close to a tie on capability, with Go ahead on
maintenance stability.

### On the map-then-deserialize step

An earlier draft argued that Rust's `serde` was decisive because the
pipeline ends by converting a generic map into a typed record, and Go
has no equivalent. That argument does not survive scrutiny.

Validation and deserialization are separable. The profile supplies JSON
Schemas at runtime (REQ-SPEC-008), so schema validation must operate on
a generic document model in *any* language. But nothing requires the
typed record to be derived *from* that model — a second, direct
unmarshal from the same source text is equally valid and is what Go
would do. The convenience `serde` provides is real but small, and it
does not discriminate between two languages that both meet every stated
requirement.

### Library survey by dimension

Assessed from documentation and, where marked ✔, exercised directly.

| Dimension | Go | Rust | Python | TypeScript |
|-----------|-----|------|--------|------------|
| CLI harness / `--help` | **cobra** — powers `kubectl`, `gh`, `docker` | **clap** ✔ — doc comments become help text | **typer** | **commander** / **oclif** |
| YAML | **goccy/go-yaml**; `go.yaml.in/yaml/v4` org-maintained | `serde_yaml_ng` ✔ — works; ecosystem churned | PyYAML / ruamel — coerces types | `yaml` (eemeli) — exposes a CST |
| JSON / TOML | stdlib; **BurntSushi/toml** | **serde_json** / **toml** | stdlib; **tomllib** stdlib | stdlib; `smol-toml` |
| JSON Schema | **santhosh-tekuri/jsonschema** ✔ | **boon** ✔ (same author) | `jsonschema` | **ajv** |
| Markdown AST + positions | **goldmark** — powers Hugo | **markdown-rs** — by remark's author | markdown-it-py | **remark / mdast** — strongest ecosystem |
| Gherkin | **official, in-process** ✔ | ⚠ third-party, diverges ✔ | **official** ✔ | **official** |
| Filesystem watching | **fsnotify** | **notify** | watchdog | chokidar |
| Git | **go-git** (pure Go) | **gix** / **git2** | pygit2 / GitPython | isomorphic-git |
| Distribution | Trivial cross-compile — the simplest of the four | **cargo-dist** → npm, PyPI, Homebrew, shell | needs a runtime | needs a runtime or a 95 MB bundle |

Shelling out to `git` is a reasonable alternative to any git library and
costs the same everywhere, so git does not discriminate.

### Ecosystem signal

- **Toward Go:** Beads, this framework's issue tracker, is written in
  Go, so contributors and agents working in this repository already meet
  a Go toolchain. Go's cross-compilation is the simplest distribution
  story of the four.
- **Toward Rust:** the dominant recent developer tooling is Rust —
  Astral's `uv`, `ruff` and `ty`, with `uv` reporting over 126 million
  downloads a month. `uv` is also the closest analogue to this tool's
  distribution problem.
- **Toward Python and TypeScript:** Anthropic's Agent SDK ships in those
  two only. That would dominate if this were a library other agents
  import; it is a CLI invoked as a subprocess, so the calling language
  need not match.

One correction to the framing that prompted this evaluation: Bun is
written in Zig, not Rust. The Rust developer-tooling examples are
Astral's.

## Decision

We will implement the spec corpus CLI in Go, with `cobra` for the
command surface, `goccy/go-yaml` for YAML, `santhosh-tekuri/jsonschema`
for JSON Schema, `goldmark` for the markdown AST, and the official
`cucumber/gherkin` Go implementation WHERE Gherkin support is later
added.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **Rust** | 1.0 MB artifact, the smallest of the four<br>`serde` makes typed deserialization a single line<br>`clap` derives `--help` from doc comments, so help cannot drift<br>`markdown-rs` is by remark's author<br>`cargo-dist` covers mirrorable-registry distribution<br>The strongest current signal in developer tooling | Gherkin is third-party and measurably diverges: no tag inheritance, wrong line number, opaque errors — adoption would require a second tool<br>YAML support is a maintained fork of an archived crate; this corner has churned twice<br>An agent cannot patch a compiled binary mid-session<br>Slowest of the four to a first working version | The closest call, and the first draft of this ADR selected it. Reversed on measurement: startup is a tie (1.77 ms vs 1.72 ms), JSON Schema is a tie (same author, identical messages), and the `serde` advantage does not survive separating validation from deserialization. What remains is a Gherkin liability with nothing left to outweigh it |
| **TypeScript** (Bun or Node) | `remark`/`mdast` is the best markdown AST ecosystem anywhere<br>`ajv` is the best JSON Schema validator anywhere<br>Official Gherkin<br>Anthropic Agent SDK language<br>An agent can patch it mid-session | The standalone executable starts in 32.29 ms and weighs 95 MB, measured<br>Node without bundling is 29.68 ms before loading a module | Fails REQ-SPEC-009.1 in practice: 32 ms of fixed overhead against a 50 ms p95 budget. A 95 MB artifact is a poor fit for a tool meant to sit in every adopter's CI |
| **Python** | Official Gherkin<br>Strongest parsing libraries overall; `tomllib` in the stdlib<br>Anthropic Agent SDK language<br>An agent can patch it mid-session<br>Fastest to a working prototype | 33.30 ms fixed startup before any work; roughly 1 s to parse 1,000 documents<br>Requires an interpreter and dependencies on the adopter's machine<br>PyYAML coerces scalars, which AC-SPEC-003.2 must work around | Fails REQ-SPEC-009.4 without bundling and REQ-SPEC-009.1 with or without it |
| **Go** | The only candidate that ties on every discriminating measure and wins the one that separates them: official in-process Gherkin, with better errors than the Rust crate<br>1.72 ms startup, 2.2 MB artifact<br>JSON Schema output identical to Rust's<br>Steadiest YAML maintenance story<br>Simplest cross-compilation of the four<br>Adjacent to Beads, already a project dependency<br>`cobra` is the most battle-tested CLI harness in any language | No sum types, so a Result-shaped pipeline carries `if err != nil` at each step<br>Reflection-based JSON is a looser guarantee than a derive macro for an index that is a published API<br>Larger artifact than Rust, and an agent cannot patch a compiled binary | N/A — selected |

## Consequences

**Positive:**
- Gherkin, if [[adr-014]] is later revisited, is an import rather than a
  second tool. That keeps the option genuinely open, which is what
  [[adr-014]] asked this decision to preserve.
- Latency and artifact requirements are met with more than an order of
  magnitude of margin, so a write-time validation hook stays
  imperceptible as corpora grow.
- Cross-compilation is a build flag, so producing the platform matrix
  REQ-SPEC-009.6 needs carries no extra machinery.
- Contributors — human or agent — meet one toolchain for both this tool
  and Beads.

**Negative:**
- Every step of the composable pipeline carries explicit error handling,
  so the code is more verbose than the equivalent Rust and the happy
  path is harder to read at a glance.
- The index is marshalled by reflection rather than from derived types,
  so a schema mistake surfaces at runtime in a consumer rather than at
  compile time. Mitigated by treating the index schema as a tested
  artifact, not merely a struct.
- The artifact is roughly twice Rust's size. Immaterial for CI, worth
  noting for embedded contexts.
- An agent debugging the CLI mid-session cannot edit and re-run it as it
  could a script. Partly offset by keeping the domain layer in
  configuration, which stays editable.

**Revisit when** Gherkin is definitively rejected for acceptance
criteria — removing the advantage that decided this — *and* a measured
need appears that Go cannot meet.

## References

- [[frd-spec-validation-cli]] — REQ-SPEC-009, the requirements this
  decision is measured against
- [[adr-014]] — defers Gherkin, and requires this decision to preserve it
- `docs/plans/spec-validation-cli-brief.md` — earlier measurements and
  the distribution research
- [cucumber/gherkin](https://github.com/cucumber/gherkin) — the shared
  test suite that defines an official implementation
- [cucumber-rs/gherkin](https://github.com/cucumber-rs/gherkin) — the
  third-party Rust crate measured above
- [santhosh-tekuri/jsonschema](https://github.com/santhosh-tekuri/jsonschema)
  and [boon](https://github.com/santhosh-tekuri/boon) — the Go and Rust
  validators, same author
- [yaml/go-yaml](https://github.com/yaml/go-yaml) — the YAML org's
  takeover of Go YAML maintenance
- [goccy/go-yaml](https://github.com/goccy/go-yaml) — the YAML
  implementation selected
- [serde-yaml deprecation](https://users.rust-lang.org/t/serde-yaml-deprecation-alternatives/108868)

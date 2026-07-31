---
title: "ADR-015: Rust as the Implementation Language for the Spec Corpus CLI"
description: Rust satisfies the latency and zero-runtime requirements the
  FRD states, and its serde pipeline matches the tool's dominant shape;
  the cost is a third-party Gherkin parser and a churning YAML ecosystem
slug: adr-015
type: adr
status: draft
updated_at: 2026-07-31T00:00:00Z
---

# ADR-015: Rust as the Implementation Language for the Spec Corpus CLI

**Status:** draft  
**Date:** 2026-07-31

## Context

[[frd-spec-validation-cli]] deliberately states requirements rather than
technology, and defers the choice here. REQ-SPEC-009 is the binding
constraint: a corpus query returns within 50 ms at p95 against a warm
index, a single-document check within 100 ms, a cold 1,000-document
index within 1 second, and the tool runs "on a machine with no language
runtime, interpreter, or package manager installed beyond the CLI
itself" while never retrieving executable content at install time.

Those numbers are not arbitrary. They come from the two invocation
patterns that dominate: an agent asking one question before writing a
document, and a write-time hook checking one file on every edit. A hook
that adds a third of a second to every save is a hook people disable.

The tool's work is narrow and repetitive: split frontmatter, parse
YAML, validate against JSON Schema, deserialize into typed records,
walk a markdown AST for headings and their spans, extract fenced and
comment blocks, resolve a reference graph, emit versioned JSON. There is
no hot numerical loop; the performance requirement is about process
startup and I/O, not throughput.

Four candidate languages were evaluated: Rust, Go, Python, and
TypeScript. All four toolchains were installed and the claims below were
measured in this repository's environment rather than taken from
documentation.

### Measured: startup and distributable artifact

The figure that matters is the *shipped artifact's* startup, not the
runtime's. Median of 15 runs each:

| Candidate | Startup | Artifact | Notes |
|-----------|---------|----------|-------|
| Rust (release, clap) | **2.06 ms** | **1.0 MB** | `--help` rendered |
| Go | **< 4.55 ms** | 3.6 MB | 4.55 ms includes two Gherkin parses |
| Bun, `bun -e ''` | 2.47 ms | n/a | Runtime already resident |
| Bun, `bun build --compile` | **32.29 ms** | **95 MB** | The shipped artifact |
| Node, `node -e ''` | 29.68 ms | n/a | Before loading any module |
| Python + yaml/json/re | 33.30 ms | n/a | Plus a runtime dependency |

The Bun row is the one that changed the analysis. Bun's resident startup
is near-native, which initially made TypeScript look viable against
REQ-SPEC-009. Compiling to the standalone executable the requirement
actually needs costs that back entirely: 32.29 ms, thirteen times the
resident figure, because the embedded runtime still bootstraps. A 50 ms
p95 budget with 32 ms of fixed overhead leaves 18 ms for the work.

Python and Node fail REQ-SPEC-009.4 outright without bundling, and
bundling reproduces Bun's result.

### Measured: Gherkin, the dimension where Rust is weakest

[[adr-014]] defers Gherkin adoption but requires the language choice to
preserve it. The same input was parsed with the official Python and Go
implementations and with `cucumber-rs/gherkin`, the third-party Rust
crate:

| | Tag extraction on a valid feature | Error on `Wen` for `When` |
|---|---|---|
| Python (official) | `['@frd-checkout', '@AC-CHK-003.1', '@req-CHK-003']` — Feature tag inherited | `(5:5): expected: #EOF, #TableRow, … #TagLine, #Examples` |
| Go (official) | `[@frd-checkout @AC-CHK-003.1 @req-CHK-003]` — Feature tag inherited | `(5:5): expected: … got '    Wen b'` |
| Rust (third-party) | `["AC-CHK-003.1", "req-CHK-003"]` — **Feature tag missing**; `@` stripped | `Error at 6:1: {"unknown keyword", "_"}` |

Three divergences, all against Rust. Tag inheritance does not happen, so
a document-level tag would not reach the criteria beneath it — the
mechanism [[adr-014]] identified as the reason tags are attractive. The
error reports line 6 for a defect on line 5, and carries no
expected-token list. Against the stated need for errors "an LLM can
easily understand", `{"unknown keyword", "_"}` is materially worse than
the official implementations' output.

This is what "official implementation" buys concretely: membership in
the `cucumber/gherkin` monorepo's shared test suite, which is what keeps
every implementation producing the same AST for the same input.

### Measured: the composable-function pipeline

The five-function pipeline — find the delimited block, parse YAML to a
generic map, compile a JSON Schema, validate the map, deserialize into a
typed record — was implemented in Rust to test whether the style fits.
It does, and each step is a separate function with no shared state. Run
against two real defects from this repository:

```
--- the real ADR-011 defect ---
  SCHEMA:
    at '/slug': 'adr-011-targeted-verification-gates'
      does not match pattern '^adr-[0-9]{3}$'
--- bad YAML (the work-order-template bug) ---
  YAML: mapping values are not allowed in this context at line 1 column 23
```

Both diagnostics carry the location and the violated rule, which is what
REQ-SPEC-006.1 and REQ-SPEC-006.2 require.

The step that discriminates between languages is the last one. The
pipeline needs the document in *both* forms: a generic map for schema
validation, then a typed record for everything downstream. In Rust this
is `serde_json::from_value::<T>(v)` — one line, and the type is the
schema. Go has no equivalent; `encoding/json` round-tripping through
`map[string]any` into a struct is weaker at expressing the framework's
optional-and-conditional frontmatter, and every step carries an
`if err != nil`. TypeScript's closest equivalent is `zod`, which is
genuinely good, but then the JSON Schema and the zod schema are two
sources of truth for one shape.

### Library survey by dimension

Assessed from documentation and, where marked ✔, exercised directly.

| Dimension | Rust | Go | Python | TypeScript |
|-----------|------|-----|--------|------------|
| CLI harness / `--help` | **clap** ✔ — doc comments *become* the help text, so help cannot drift from the code; errors suggest a fix | **cobra** — excellent, powers `kubectl`, `gh`, `docker` | **typer** — very good, type hints drive it | **commander** / **oclif** — good; oclif is heavier |
| YAML | ⚠ **serde_yaml archived** 2024; `serde_yaml_ng` ✔ works, `serde-saphyr` and `yaml-rust2` also live. A fork, `serde_yml`, was itself deprecated over unsoundness | **goccy/go-yaml** — passes the YAML test suite more completely than `yaml.v3`; the YAML org now maintains `go.yaml.in/yaml/v4` after the original was marked unmaintained in 2025 | **PyYAML** / **ruamel** — mature, ubiquitous | **yaml** (eemeli) — excellent, exposes a CST with positions |
| JSON | **serde_json** — the reference implementation of the pattern | `encoding/json` — stdlib, reflection-based | stdlib | stdlib |
| TOML | **toml** — serde-integrated | **BurntSushi/toml** | **tomllib** — stdlib since 3.11 | `smol-toml` |
| JSON Schema | **boon** ✔ — draft 2020-12, JSON-Pointer-located errors | **santhosh-tekuri/jsonschema** — same author as boon | **jsonschema** | **ajv** — the fastest and most complete anywhere |
| Markdown AST + positions | **markdown-rs** — by the author of remark; also `comrak`, `pulldown-cmark` | **goldmark** — solid AST | **markdown-it-py** | **remark / unified / mdast** — the strongest ecosystem in any language |
| Gherkin | ⚠ third-party, diverges ✔ | **official** ✔ | **official** ✔ | **official** |
| Filesystem watching | **notify** — the de-facto cross-platform watcher | **fsnotify** | **watchdog** | **chokidar** |
| Git | **gix** (pure Rust) / **git2** | **go-git** (pure Go) | **pygit2** / **GitPython** | **isomorphic-git** |
| Distribution | **cargo-dist** → platform packages for npm, PyPI, Homebrew, shell | Trivial cross-compile; the simplest story of the four | Needs a runtime, or fragile bundling | Needs a runtime, or a 95 MB bundle |

Shelling out to `git` is a reasonable alternative to any of the git
libraries and is the same cost in every language, so git integration
does not discriminate.

### Ecosystem and community signal

Signals point in two directions, and neither is decisive:

- **Toward Python and TypeScript:** Anthropic's Agent SDK ships in those
  two languages only, and the Claude Code plugin ecosystem is
  JavaScript-adjacent. If the goal were a library other agents import,
  this would dominate. It is a CLI invoked as a subprocess, so the
  calling language does not have to match.
- **Toward Rust:** the dominant recent developer-tooling tools are Rust
  — Astral's `uv`, `ruff` and `ty`, with `uv` reporting over 126 million
  downloads a month and Astral joining OpenAI in March 2026. `uv` is
  also the closest analogue to this tool's distribution problem, and it
  solved it as a single static binary shipped through several package
  registries.
- **Toward Go:** Beads, this framework's issue tracker, is written in
  Go, so a Go CLI would be adjacent to a tool the project already
  depends on.

One correction to the framing that prompted this: Bun is written in
Zig, not Rust. The Rust developer-tooling examples are Astral's.

## Decision

We will implement the spec corpus CLI in Rust, with `clap` for the
command surface, `serde` with `serde_json`, `serde_yaml_ng` and `toml`
for wire formats, `boon` for JSON Schema, and `markdown-rs` for the
markdown AST.

WHERE the CLI later adds Gherkin support, it SHALL obtain the parse from
an official Gherkin implementation rather than from the Rust crate —
either by contributing the missing tag inheritance and error positions
upstream, or by invoking the official CLI, which emits newline-delimited
JSON ASTs across a clean subprocess boundary.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| **Go** | Official Gherkin in-process, measurably better errors than the Rust crate<br>Healthiest YAML story of the four<br>Simplest cross-compilation and distribution<br>2–4 ms startup, 3.6 MB artifact<br>Adjacent to Beads, already a project dependency<br>`cobra` help is excellent | No `serde` equivalent: the map→typed-record step that ends every pipeline is weaker and more verbose<br>No sum types, so the Result-shaped composition the tool is made of carries `if err != nil` at every step<br>Reflection-based JSON is a looser guarantee for an index that is a published API | The closest call. Rejected on the composable-pipeline fit, which the tool exercises dozens of times, against a Gherkin advantage that is deferred and mitigable. IF Gherkin becomes load-bearing sooner than expected, THEN this is the option to revisit |
| **TypeScript** (Bun or Node) | `remark`/`mdast` is the best markdown AST ecosystem anywhere<br>`ajv` is the best JSON Schema validator anywhere<br>Official Gherkin<br>Anthropic Agent SDK language<br>An agent can patch it mid-session | The standalone executable starts in 32.29 ms and weighs 95 MB, measured — Bun's 2.47 ms does not survive compilation<br>Two sources of truth if `zod` is used alongside JSON Schema<br>Node without bundling is 29.68 ms before loading a module | Fails REQ-SPEC-009.1 in practice: 32 ms of fixed overhead against a 50 ms p95 budget. The 95 MB artifact is a poor fit for a tool meant to sit in every adopter's CI |
| **Python** | Official Gherkin<br>Strongest parsing libraries overall; `tomllib` in the stdlib<br>Anthropic Agent SDK language<br>An agent can patch it mid-session<br>Fastest to a working prototype | 33.30 ms fixed startup before any work; roughly 1 s to parse 1,000 documents<br>Requires an interpreter and dependencies on the adopter's machine, or fragile freezing | Fails REQ-SPEC-009.4 without bundling and REQ-SPEC-009.1 with or without it |
| **Rust** | 2.06 ms startup, 1.0 MB artifact — the widest margin against REQ-SPEC-009<br>`serde` makes the map→typed-record step one line, and the type *is* the schema<br>`clap` derives help from doc comments, so `--help` cannot drift<br>`boon` produces JSON-Pointer-located schema errors<br>`markdown-rs` is by remark's author<br>`cargo-dist` covers the mirrorable-registry requirement | Gherkin is third-party and measurably diverges — no tag inheritance, wrong line number, opaque errors<br>`serde_yaml` was archived; the replacement is a maintained fork, and this corner of the ecosystem has churned<br>An agent cannot patch a compiled binary mid-session<br>Slowest of the four to a first working version | N/A — selected |

## Consequences

**Positive:**
- The latency requirements are met with a wide margin rather than
  narrowly, so a write-time validation hook stays imperceptible and the
  budget survives the corpus growing.
- The distributable is ~1 MB with no runtime, which makes the
  locked-down-CI case in REQ-SPEC-009.5 and .6 straightforward.
- `--help` is generated from the same doc comments that document the
  code, so the agent-facing surface cannot drift from behaviour — a
  property the framework asks of its own documents.
- The index is emitted from typed records rather than assembled from
  maps, so a malformed index is a compile error rather than a runtime
  surprise in a consumer.

**Negative:**
- Gherkin support, if adopted, costs either upstream contribution or a
  subprocess boundary. It cannot simply be imported, as it could in the
  other three languages.
- The YAML dependency is a maintained fork of an archived crate. This
  corner needs watching, and a migration to `serde-saphyr` or a
  successor is a plausible future cost.
- An agent debugging the CLI mid-session cannot edit and re-run it the
  way it could a script. Partly offset by keeping the domain layer in
  configuration, which stays editable.
- Contributors need a Rust toolchain, which is a higher bar than Go or
  Python for a project whose contributors are often agents.
- Prototyping is slower, so the feedback loop from the first real
  corpora is longer than a Python version would have given.

**Revisit when** Gherkin moves from deferred to load-bearing and the
upstream Rust crate has not gained tag inheritance and positioned
errors; or if the YAML fork is abandoned without a maintained
successor.

## References

- [[frd-spec-validation-cli]] — REQ-SPEC-009, the requirements this
  decision is measured against
- [[adr-014]] — defers Gherkin, and requires this decision to preserve it
- `docs/plans/spec-validation-cli-brief.md` — the earlier measurements
  and the distribution research behind cargo-dist
- [cucumber/gherkin](https://github.com/cucumber/gherkin) — the shared
  test suite that defines an official implementation
- [cucumber-rs/gherkin](https://github.com/cucumber-rs/gherkin) — the
  third-party Rust crate measured above
- [serde-yaml deprecation](https://users.rust-lang.org/t/serde-yaml-deprecation-alternatives/108868)
- [yaml/go-yaml](https://github.com/yaml/go-yaml) — the YAML org's
  takeover of Go YAML maintenance
- [boon](https://github.com/santhosh-tekuri/boon) — JSON Schema for Rust
- [cargo-dist](https://axodotdev.github.io/cargo-dist/) — the
  distribution path REQ-SPEC-009.6 needs
- [Astral](https://astral.sh/) — uv, ruff, ty; the Rust
  developer-tooling signal

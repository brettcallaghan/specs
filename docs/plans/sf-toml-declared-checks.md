# Plan — declared checks in `sf.toml`

Replace the single executable `checks` script with an array of declared checks
in `sf.toml`, each carrying an `id` and a `command`. A plugin-shipped runner
executes them; the project supplies commands, not a protocol.

Transient. Delete when the change lands.

---

## 1. What changes

**Today.** A project writes an executable, declares its path in `sf.toml`
under `[checkpoints] checks`, and the contract is that script's exit status.
Everything the run would like to know about individual checks — their names,
which ones a bead should run on its inner loop, which one is known-red — is
unavailable, because the contract deliberately asks the project for nothing
beyond a status.

**After.** The project declares its checks:

```toml
[[checks]]
id = "vet"
command = 'cd cli && go vet ./...'

[[checks]]
id = "test-go"
command = 'cd cli && go test ./...'
```

`plugin/skills/bead-bonanza/scripts/bb-checks.sh` reads that array and runs it.
The floor to adopt `bead-bonanza` drops from "write an executable" to four
lines of TOML; a project that needs a script still points a `command` at one.

**What does not change.** Each command's own exit status is its verdict.
Nothing parses tool output. Attribution is still a bead comparing its own
checks output against the baseline's raw text. This moves *where the list of
commands lives*; it does not reverse [[adr-022]]'s load-bearing choice that a
check reports its tools' own output and no findings format.

## 2. The contract

### Schema

The array holds the **per-commit** quality bar, not the release bar: every bead
runs these before its own close, so a project puts the checks that are cheap
enough to pay every time here and leaves the multi-minute harnesses to CI. That
premise is what makes per-bead execution affordable, and it is unchanged by the
move from a script to a declaration.

`sf.toml`, repository root, array of tables:

| key | required | notes |
|---|---|---|
| `id` | yes | `[a-z0-9][a-z0-9-]*`, unique across the array |
| `command` | yes | non-empty string, run via `sh -c` |

**No other keys are accepted in v1**, and an unrecognised key is a hard error
rather than an ignored one. A typo (`comand = …`) that silently dropped a
check would read downstream as a check that passed.

Prefer TOML *literal* strings (single quotes) for `command`: they do no escape
processing, so a shell one-liner survives verbatim.

### Execution semantics

- **`sh -c "$command"`**, working directory the **repository root**,
  environment inherited. A check that needs a subdirectory writes
  `cd cli && …`.
- **Serial, in declaration order, always.** Never reordered, never
  parallelised — declaration order is the only place a project can express
  "this must run before that". This holds even when ids are passed as
  arguments: `bb-checks.sh build schemas` runs them in declared order, not
  argument order.
- **Every requested check runs, even after one fails.** This is the change's
  strongest single justification: it is currently a *convention* the contract
  cannot enforce, documented in a comment asking projects not to write the
  `&&` chain that is the natural thing to write.

### Runner interface

`bb-checks.sh` — shipped by the plugin, not by the project.

```
bb-checks.py                    # run every declared check
bb-checks.py <id> [<id>…]       # run only these, in declaration order
bb-checks.py --ids              # print declared ids, one per line
bb-checks.py --report <path>    # also write `<id><TAB><exit>` per check there
bb-checks.py --self-test
```

`--ids` is not the `list` subcommand [[adr-022]] rejected. That one asked the
*project's* script to enumerate what running the suite already prints; this
reads our own schema out of a config file and asks the project for nothing.

**Exit status** — matching `bb-status.sh`'s convention:

| | |
|---|---|
| 0 | every requested check passed |
| 1 | at least one failed |
| 3 | invocation or configuration error — malformed TOML, no `[[checks]]`, duplicate id, unknown key, missing `command`, unknown id argument |

A 3 runs **nothing**. A bad declaration must not half-execute.

### Output

Between delimiters, the tools' own bytes, untouched:

```
=== vet
# …whatever go vet printed…
=== vet FAILED (exit 2)
```

Delimiters go to stderr; the command's own streams pass through unmodified.

**They are a reading aid, not a machine interface.** They share a stream with
bytes the runner does not control, so a tool can print a line identical to one
— measured: a check echoing `=== other-check FAILED (exit 9)` made an in-band
reader report a passing check as failed. A caller needing verdicts passes
`--report <path>` and reads them out of band. The aggregate verdict remains the
runner's exit status, and bead-level attribution remains raw-text comparison
within a check's block.

### Resolution

**One mechanism.** `[[checks]]` in `sf.toml`, `$BB_SF_TOML` overriding the
file's location (fixtures and self-test need this). The `[checkpoints] checks`
key and the conventional `scripts/checks` fallback are **removed**, not kept
alongside — two ways to declare the same thing is a fork in the contract, and
nothing outside this branch has ever used either. A project wanting one script
declares one entry pointing at it.

`$BB_CHECKS` (a path to an executable) is retired with the key it overrode.

## 3. Steps

Ordered. Each step leaves the tree working.

### S1 — the runner

New: `plugin/skills/bead-bonanza/scripts/bb-checks.sh`.

- Resolve repo root via `git rev-parse --show-toplevel`.
- Parse `sf.toml` with **python3 `tomllib`** (stdlib since 3.11; measured here
  as 3.11.15). Do **not** hand-roll `[[checks]]` in awk — restricted parsing
  of a real format is the fragility class this whole review has been removing,
  and `sf_get`'s existing awk reader cannot do array-of-tables anyway.
  `jq` is already a hard pre-flight dependency; python3 joins it, asserted at
  the top with a message naming what is missing.
- Validate the whole array **before** running anything, then execute.
- `--self-test` against fixture TOML files, in the style `bb-status.sh` already
  uses.

`sf_get` stays where it is — `[models]` and `[run]` are still flat key reads
and do not need a parser.

### S2 — pre-flight

`plugin/skills/bead-bonanza/scripts/pre-flight-check.sh`, section 2 and
section 8.

- Replace the `sf_get checkpoints checks` block: locate `bb-checks.sh`,
  run `--ids`, and block the run when the declaration is absent or invalid,
  quoting the runner's own error rather than restating it.
- Emit `BB_CHECKS_IDS` (space-separated) alongside the existing facts.
  `BB_CHECKS` (a path) goes away.
- Baseline capture: run `bb-checks.sh`, write `$RUN_DIR/checks-baseline.md` as
  today. The captured text now carries `=== <id>` delimiters, so add a
  per-check verdict summary above the raw block, derived from them.

### S3 — the prompt and skill surfaces

- `references/subagent-prompts.md` — replace `Gates: format=` + `{FORMAT}`
  `{LINT}` `{BUILD}` `{TEST}` `{REGEN}` `{CUSTOM_GATES}` with
  `Checks: {CHECKS_IDS}`. **These placeholders are already dead** — pre-flight
  stopped parsing `AGENTS.md` in wave 2 and emits none of them, so lines 70,
  89, 157, 162-163, 180, 184, 189, 192, 197 and 200 currently interpolate
  nothing. Declared ids are what they were reaching for. File this as a defect
  regardless of whether this plan proceeds.
- `run-bead/SKILL.md` — Inputs table: the `checks` row becomes `bb-checks.sh`;
  add a `{CHECKS_IDS}` row. Step 4 gains one sentence: run the ids covering
  your surface while iterating, the full set before close. Watch the body byte
  budget.
- `references/project-setup.md` — rewrite step 1 for the declared form.
- `bead-bonanza/SKILL.md` — the ~8 lines naming "the project's `checks`
  script". Line 903's hedge ("targeted tests, WHERE the project's `checks`
  script offers a way to name them") becomes unconditional. Line 996 cites
  `BB_LINT` and `BB_MTIME` as example facts; pre-flight emits neither — same
  residue, fix it here.

### S4 — this repository's own migration

- `sf.toml`: nine `[[checks]]` entries replacing `[checkpoints]`, in the order
  `scripts/checks` currently uses.
- Delete `scripts/checks`.
- New `scripts/format-check`: `gofmt -l` exits 0 whether or not it found
  files, so the file list is the verdict. Expressing that as a `command`
  one-liner is possible and ugly; a two-line script is the honest form — and
  it exercises the escape hatch in the repository that defines the contract,
  which is worth having.
- `AGENTS.md`: the gates section points at `sf.toml`; the targeted-tests table
  stops being "a reading aid, not an interface" and becomes ids an agent can
  actually pass.

**Resolve during migration:** `AGENTS.md` claims `schemas` must precede
`build` ("the generated Go must still compile"), while `scripts/checks` runs
`build` first. Both are fine, because `--check` verifies without
regenerating — but that means this repository has **no** actual
ordering dependency, and [[adr-022]]'s "a regen before a test" is an
illustration, not something measured here. Say so rather than carrying an
example that does not exist.

### S5 — ADR-022

**Rewrite in place**, keeping slug `adr-022` so `[[adr-022]]` references hold.
Change the title and description; rewrite the Decision; add an Alternatives
row recording "one executable script, exit-status contract" as considered and
replaced, with its reasoning intact; re-read Negatives against the template's
activation rule.

Rewrite rather than supersede because neither design has run outside this
branch, so a supersession chain would document a decision that never took
effect. **If PR #27 merges before this lands, supersede instead** — once the
script design has shipped, replacing it silently is the wrong record.

## 4. Verification

`bb-checks.sh --self-test`, over fixtures:

| fixture | expected |
|---|---|
| two passing checks | exit 0, both blocks present |
| first check fails, second passes | exit 1, **both** ran |
| duplicate `id` | exit 3, nothing ran |
| unknown key | exit 3, nothing ran |
| missing `command` | exit 3 |
| malformed TOML | exit 3, parser error surfaced |
| no `[[checks]]` array | exit 3, distinct message from "empty array" |
| `bb-checks.sh b a` where declaration order is `a, b` | runs `a` then `b` |
| unknown id argument | exit 3, nothing ran |
| check that `cd`s | the next check still starts at repo root |

Plus, once: run the migrated declaration against this repository and diff its
output against what `scripts/checks` printed on the same tree. Modulo the
delimiter lines they must be identical — that is the migration's proof.

## 5. Not in v1

Named so they are not added by reflex, each because it re-imports something
the exit-status contract removed:

- `dir` — `cd X && …` covers it; two keys is the appeal.
- `depends_on` — declaration order already expresses sequence.
- `known_red` — a verdict format returning. Baseline comparison handles it,
  and the `id` already makes a known-red check's output separable.
- parallelism, per-check timeouts, an `env` key.

## 6. Decisions pinned

Settled here so they are not re-litigated mid-implementation:

1. `id` + `command`, no other keys; unknown key is a hard error.
2. `sh -c`, cwd = repository root, environment inherited.
3. Serial, declaration order, always — including when ids are arguments.
4. Every requested check runs even after one fails.
5. `--ids` exists; no `list`, `run`, or `targeted` subcommand.
6. Exit 0 / 1 / 3, matching `bb-status.sh`; a 3 runs nothing.
7. `=== <id>` delimiters and a `FAILED (exit N)` line, stderr, tool bytes
   untouched between them — for a human reading the log. Machine-readable
   verdicts come from `--report <path>`, out of band, because an in-band
   delimiter is forgeable by any tool the runner invokes.
8. One declaration mechanism. `[checkpoints]`, `$BB_CHECKS` and the
   `scripts/checks` fallback are removed.
9. python3 `tomllib` for parsing; python3 becomes a hard dependency.
10. ADR-022 rewritten in place, slug unchanged — unless PR #27 merges first.

## 7. Suggested bead breakdown

| | scope | depends on |
|---|---|---|
| B1 | `bb-checks.sh`, fixtures, self-test | — |
| B2 | pre-flight resolution, `BB_CHECKS_IDS`, baseline capture | B1 |
| B3 | prompt and skill surfaces (S3) | B2 |
| B4 | this repo's migration (S4) + ADR-022 (S5) | B2 |

B3 and B4 touch disjoint files and are wave-parallel.

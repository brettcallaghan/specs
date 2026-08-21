---
title: "ADR-022: Projects Declare Their Checks in sf.toml"
description: A project declares its checks as an array of id-and-command
  entries in a plugin-scoped sf.toml, holding the quality bar for a single
  commit rather than the release bar, and a plugin-shipped runner executes
  them, because a prose gate profile cannot express a multi-command gate, a
  bare suite-wide exit code cannot express a known-red baseline, and asking
  every project to write a runner puts the same code in every repository
slug: adr-022
type: adr
status: active
updated_at: 2026-08-21T00:00:00Z
---

# ADR-022: Projects Declare Their Checks in sf.toml

**Status:** active
**Date:** 2026-08-17

## Context

`bead-bonanza` recovers a project's gate commands by regex-parsing prose.
`scripts/pre-flight-check.sh` extracts "the first backticked span" from each
bullet under `## bead-bonanza gates` in `AGENTS.md`. Three consequences,
all measured on this repository:

**The parser cannot express a multi-command gate.** This repository's test
gate is a fenced `bash` block, because three checks must run independently
rather than chained with `&&` — one of them is known-red and would
short-circuit the other two. The bullet line therefore carries no backticked
span, and the parser emits `unparsed (no command in backticks)`, sets
`$key=unparsed`, and continues. Reproduced directly against the current
`AGENTS.md`. The gate it fails to measure is the only known-red one — the
exact gate the "no findings beyond baseline" bar depends on.

**A pass/fail verdict cannot carry a baseline, and the baseline never
reaches the agent that needs it.** `--gate-baseline` runs the profile on the
untouched tree, and `{GATE_BASELINE}` is passed to the grooming/planning
prompt and consumed by [[adr-019]]'s planner. It stops there. The per-bead
prompt's `Gates:` line has no baseline field, and its gate step says only
"On gate failure, fix the root cause and rerun". So on any known-red gate,
every bead is instructed to fix a failure that is not its own: a pre-existing
lint finding, a suite that fails for environmental reasons, or a finding owned
by a different bead in the same run.

`SKILL.md` already specifies the remedy and does not implement it — it
requires an "environmental-failure ledger" whose verdicts are "recorded in the
gate baseline pre-flight captured, then passed into every subsequent subagent
prompt as fact." No file holds it and no prompt field carries it.

**`AGENTS.md` accumulates content that applies only while one skill runs.**
The gates block, the model lineup, the targeted-tests-by-surface list and the
environmental ledger are read by every agent in the repository and used by
one workflow.

## Decision

A project declares its checks as an array of `[[checks]]` tables in `sf.toml`,
a plugin-scoped config file at the repository root. Two keys each, and no
others are accepted:

```toml
[[checks]]
id = "test-go"
command = 'cd cli && go test ./...'
```

`bead-bonanza` ships the runner. It reads the declaration and executes it; the
project supplies commands, not a protocol, and no prose is parsed anywhere.

**Each command's own exit status is its verdict**, and a check's output is
whatever its tool printed, byte for byte. This is the load-bearing choice and
it runs opposite to this ADR's first draft, which required findings carrying
file paths so a finding could be attributed to the bead owning that file. A
project's checks wrap `go test`, `eslint`, `gofmt`, `pytest` — formats the
project does not control. Requiring paths would not remove the parsing that
motivated this ADR; it would move it out of `bead-bonanza` and onto every
project that wants to be gated, once per tool.

**What belongs in the array is the quality bar for one commit, not the release
bar.** Every bead runs these before its own close, so the project selects
checks that give real confidence in a single change and are cheap enough to pay
every time; a multi-minute harness, or a check that only means anything across
the whole repository, belongs in CI. The selection is the project's, made once
when it writes the declaration — `bead-bonanza` runs what it finds, in full,
every time.

**Two rules move out of convention and into the runner.** Execution is serial
in declaration order, so a step that must precede another is sequenced by the
project; and every requested check runs even after one fails. Under a
project-written script both were things the project had to remember, and `&&`
— the natural thing to write — breaks the second, which is precisely how a
known-red check masks everything behind it.

**Why the ids live in the config and not in the script.** Successive drafts
asked the project's *script* for names — a `list` subcommand, a
`RESULT <id> <verdict>` line per check — and each failed the same test: *what
does a project have to build before it can be gated at all?* Declaring an id
beside its command asks for nothing. The two are the same line, so they cannot
drift; there is no protocol to implement, and no check that can be run but
forgotten in the report. That is the whole difference between a name a project
*declares* and a name a project must *emit*.

**Which ids matter for a change is a judgement, not a declaration.** An agent
touching `cli/**` runs `format vet build test-go` while iterating and the full
set before close, reading relevance off the slug. No project declares a
path-to-check mapping; the earlier `targeted` subcommand asked for exactly that
and is not reinstated.

**The floor is four lines of TOML**, where a project-written script was
previously mandatory. A project that needs a script still points a `command` at
one — strictly more expressive than requiring one of everybody.

**Attribution is by comparison, done by the bead.** The bar "no findings beyond
baseline" is a bead reading its own output against the baseline's: a line in
both was already there, a line only in its own is its own. That needs no
structure in either text and no precomputation, which retires this ADR's
original assignment of per-bead attribution to `plan-sprint`.

The run's known-red state is written once by pre-flight to
`$RUN_DIR/checks-baseline.md` — a per-check verdict table over the raw output,
verbatim — and is **inlined into every subagent prompt**, always, including
when it is empty, since "nothing is known-red" and "no baseline was taken" must
not look alike.

**What this does not change.** A checks finding is one of four kinds of finding
a run produces, and they are not interchangeable — conflating the first two is
what left the baseline undelivered:

| Finding | Channel | Read by | When |
|---|---|---|---|
| Pre-existing, owned by no queued bead | `$RUN_DIR/checks-baseline.md`, inlined in every prompt | every subagent | at dispatch |
| Pre-existing, owned by a queued bead | the same baseline block; the bead recognises its own by comparing its output against it | that bead | when it runs its checks |
| Discovered mid-run, someone must act | a bead (`bd create --discovered-from`) | this run's orchestrator | immediately |
| Friction or observation | a `retro:` comment on the bead | a later retrospective | Phase 4 |

Only the first two are this ADR's subject, and they share one channel: the
distinction between them is drawn by the bead at check time, not by the run
ahead of time. The third is the tracker, which is live and queryable mid-run;
the fourth is harvested after the run ends, which is why it is the wrong
channel for anything anyone must act on.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Keep parsing `AGENTS.md`, teach the parser fenced blocks | Smallest change; no project migration | Fixes one shape of a general defect; still cannot express per-surface tests, conditional gates, or a findings count; the parser silently takes a wrong span when prose precedes the command | Treats the instance, not the class. The parser would still return a string where the baseline needs a count |
| Discovery by `.specs/hooks/<name>` convention | Zero configuration; nothing to declare | Introduces a dotted directory one character from the existing `specs/` corpus root and meaning something entirely different; a magic directory is undiscoverable until it is documented, which is the problem being solved | The `specs/` / `.specs/` confusability is a standing trap for every reader of this repository |
| A strict machine-readable contract (JSON schema for findings) | Unambiguous; validatable | A schema the project must satisfy makes writing a checkpoint script a specification exercise; projects with a one-line lint gate would pay the full cost | Disproportionate — and, on implementation, worse than that: see the row below |
| Findings carrying file paths, without a full schema (this ADR's first draft) | Lets a pre-existing finding be attributed to the bead owning that file, precomputed by the planner | A project's script wraps tools whose output it does not control — `go test`, `eslint`, `gofmt`, `pytest` each print differently — so emitting paths means the PROJECT parses its own tools, once per tool. That is the parsing this ADR exists to remove, relocated rather than removed | Rejected on implementation. Attribution by comparison (a bead reads its gate output against the baseline's) delivers the same outcome with no format at all |
| Declare scripts in `specs.toml` | No new file; `specs.toml` already exists and the CLI already reads it | `specs.toml` is a **per-corpus-root manifest**, not a per-repo one — this repository has two real instances (`./specs.toml` declaring `specs/`, `./rfcs/specs.toml` declaring `rfcs/`) plus ~50 CLI fixtures. "Which `specs.toml` holds the checkpoint scripts?" has no answer, and every fixture would have to tolerate a field unrelated to a corpus | Structurally wrong file. `bead-bonanza` requires groomed beads, not a spec corpus; coupling its config to a corpus manifest couples two things this framework deliberately keeps separate |
| Environment variables in a `.env` file | No new format; the run already reads some env (`CLAUDE_CODE_SUBAGENT_MODEL`) | `.env` is conventionally per-environment and commonly gitignored, so the declaration would not travel with the repository — and a checkpoint script path is project structure shared by the team, not an environment knob or a secret. Flat keys express no structure, so multiple checkpoints become `SF_HOOK_PREFLIGHT`-style stringly-typed conventions | The declaration must be committed and shared; `.env` is the wrong lifetime and the wrong shape |
| Declare scripts in a plugin-scoped `sf.toml`, findings-with-paths, guideline contract | Explicit and greppable; per-repo, matching the plugin's own scope; the project points at scripts it already has rather than moving them | One more config file at the repository root | Right file, wrong contents — the findings format was rejected on implementation (row above) and the declaration reduced to a script path |
| One executable `checks` script declared by path, contract = its exit status | Nothing to parse; the project runs whatever it likes in whatever order; the smallest contract that can express a multi-command gate | Writing an executable is mandatory even for `cargo test`; "run every check even after one fails" is a convention the project must remember, and `&&` is the natural thing to write instead; the script's individual checks have no names, so nothing can be targeted and the baseline cannot be partitioned | Superseded by declared checks. Every project would have written the same runner, and the one convention the contract could not enforce is exactly the one that hides failures |
| Declare checks as `[[checks]]` in a plugin-scoped `sf.toml`, runner shipped by the plugin | Floor drops to four lines of TOML; the run-every-check and declaration-order rules become the runner's to enforce; ids cost nothing because they sit beside the command they name, so targeting works with no path mapping; a project that needs a script still points a `command` at one | One more config file at the repository root; the runner needs a real TOML parser, so python3 3.11+ joins `jq` as a hard dependency; commands now live in config rather than in a file with a shebang and a mode bit | N/A — selected |

## Consequences

**Positive:**
- A multi-command check is expressible, because it is a declared command
  rather than a span of prose inside a bullet.
- The floor for adopting `bead-bonanza` is four lines of TOML naming commands
  the project already runs. It never has to write a runner, and never has to
  understand what its tools printed.
- "Run every check even after one fails" is enforced rather than requested. It
  is the one rule that, when forgotten, hides failures rather than surfacing
  them.
- The known-red baseline reaches the agent that runs the checks, partitioned
  per check, and a bead is told which findings are correct to leave.
- The environmental-failure ledger gets a home, implementing a rule `SKILL.md`
  already states.
- `AGENTS.md` sheds the gate profile and the environmental ledger.
- Targeting needs no declaration: an agent picks ids off their slugs, so a
  project never maintains a path-to-check mapping.
- Attribution costs no new data and no new step: a bead compares its own
  output against the baseline text.

**Negative:**
- One more config file at the repository root, alongside `specs.toml`,
  `.beads/config.yaml` and `.claude/settings.json`.
- python3 3.11+ becomes a hard dependency of every run, because an array of
  tables needs a real TOML parser and a restricted hand-rolled one would fail
  on valid input.
- Commands live in config rather than in a file with a shebang, a mode bit and
  a diff history. Editing `sf.toml` is editing what runs, which is no larger a
  grant than editing a checks script but is a less obvious one.
- A check whose verdict is its output rather than its status — `gofmt -l` is
  the standing example — needs either an awkward shell one-liner or a small
  script, so the simplest declaration is not always the honest one.
- Every subagent prompt carries a checks-baseline block even when empty,
  costing a line of context per dispatch to remove the ambiguity between
  "nothing known-red" and "not measured".
- Attribution is best-effort rather than exact. Comparing raw output reads a
  line that moved or was reworded as "new", so a bead can be pointed at a
  pre-existing finding it did not cause. The failure mode is a bead
  investigating something harmless, not a real regression slipping through.

## References

- [[adr-019]] — skill-composed subagent orchestration; the planner this
  decision originally assigned attribution to, and no longer does
- [[adr-017]] — the Spec CLI as a plugin prerequisite; the precedent for a
  skill depending on an executable rather than branching on its absence
- `rfcs/inbox/bead-bonanza/rfc-self-acceptance-signals-and-gate-baseline.md`
  P2 — the fenced-gate parse failure this ADR supersedes
- `rfcs/inbox/bead-bonanza/rfc-self-worktree-composition-and-shared-gates.md`
  P3 — baseline ownership across concurrent beads, resized by this ADR

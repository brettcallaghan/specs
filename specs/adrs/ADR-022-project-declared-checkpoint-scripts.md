---
title: "ADR-022: Projects Declare Checkpoint Scripts"
description: A project declares its gate commands as one executable checks
  script referenced from a plugin-scoped sf.toml, whose contract is its exit
  status and whose output is whatever its tools printed, because a
  prose gate profile cannot express a multi-command gate and a bare suite-wide
  exit code cannot express a known-red baseline
slug: adr-022
type: adr
status: active
updated_at: 2026-08-17T00:00:00Z
---

# ADR-022: Projects Declare Checkpoint Scripts

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

A project declares **one executable `checks` script** — an executable it owns
— through `sf.toml`, a plugin-scoped config file at the repository root,
referenced by repository-root-relative path. `bead-bonanza` runs it and reads
its output; it parses no prose.

**The contract is the exit status.** Run the script: 0 means every check
passed, anything else means something failed, and whatever it printed says
what. No subcommands, no output format, no identifiers.

Successive drafts of this ADR asked for more and each addition failed the same
test — *what does a project have to build before it can be gated at all?*
Findings carrying file paths meant parsing tools it does not control. A
`targeted` subcommand meant encoding a path-to-gate mapping. A `list`
subcommand meant enumerating what running the suite already prints. A
`RESULT <gate> <verdict>` line per check meant inventing an identifier for
every check and echoing a line about it.

None of it was load-bearing, because the baseline is a **comparison of raw
text**, and a comparison needs no names. The floor is now a shell script that
runs the commands a project already runs.

The one thing the contract asks beyond the exit status is a convention rather
than a format: **run every check even after one fails.** Chaining with `&&`
stops at the first failure and hides the rest, which is precisely how a
known-red check masks everything behind it.

**A checkpoint reports its tools' own output, not a structured findings
format.** This is the load-bearing choice and it runs opposite to this ADR's
first draft, which required findings carrying file paths so a finding could be
attributed to the bead owning that file. A project's checks script wraps
`go test`, `eslint`, `gofmt`, `pytest` — formats the project does not control.
Requiring paths would not remove the parsing that motivated this ADR; it would
move it out of `bead-bonanza` and onto every project that wants to be gated,
once per tool. The verdict is the one thing the script must synthesize, because
it is the one thing the tools do not already say.

Ordering lives inside the script, so a step that must precede another — a
regen before a test — is sequenced by the project rather than by the caller. A check that is
known-red gets its own gate id, so the baseline can carry it without marking a
healthy check red.

The run's known-red state is written once by pre-flight to
`$RUN_DIR/gate-baseline.md` — the verdict table plus, verbatim, the output each
failing gate produced — and is **inlined into every subagent prompt**, always,
including when it is empty, since "nothing is known-red" and "no baseline was
taken" must not look alike.

**Attribution is by comparison, done by the bead.** The bar "no findings beyond
baseline" is a bead reading its own gate output against the baseline's: a line
in both was already there, a line only in its own is its own. That needs no
structure in either text, and it needs no precomputation — which retires this
ADR's assignment of per-bead attribution to `plan-sprint`.

There is **one** project-supplied checkpoint, not two. `bead-bonanza`'s own
`pre-flight-check.sh` is the preflight step, and it is the skill's, not the
project's; what the project supplies is the thing run per bead and at the
baseline, which is `checks`.

**What this does not change.** A checkpoint finding is one of four kinds of
finding a run produces, and they are not interchangeable — conflating the
first two is what left the baseline undelivered:

| Finding | Channel | Read by | When |
|---|---|---|---|
| Pre-existing, owned by no queued bead | `$RUN_DIR/gate-baseline.md`, inlined in every prompt | every subagent | at dispatch |
| Pre-existing, owned by a queued bead | the same baseline block; the bead recognises its own by comparing its gate output against it | that bead | when it runs its gates |
| Discovered mid-run, someone must act | a bead (`bd create --discovered-from`) | this run's orchestrator | immediately |
| Friction or observation | a `retro:` comment on the bead | a later retrospective | Phase 4 |

Only the first two are this ADR's subject, and they share one channel: the
distinction between them is drawn by the bead at gate time, not by the run
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
| Declare scripts in a plugin-scoped `sf.toml`, findings-with-paths, guideline contract | Explicit and greppable; per-repo, matching the plugin's own scope; the project points at scripts it already has rather than moving them; reuses the `BB_*` key=value emission convention `pre-flight-check.sh` already uses | One more config file at the repository root | N/A — selected |

## Consequences

**Positive:**
- A multi-command gate is expressible, because it is a script rather than a
  span of prose inside a bullet.
- The floor for adopting `bead-bonanza` is a shell script that runs the
  commands a project already runs, and prints one line per gate. It never has
  to understand what its tools printed.
- The known-red baseline reaches the agent that runs the gate. A bead is told
  the pre-existing findings count, which subset is its to clear, and that the
  remainder is correct to leave.
- The environmental-failure ledger gets a home, implementing a rule `SKILL.md`
  already states.
- `AGENTS.md` sheds the gate profile, the targeted-tests list and the
  environmental ledger.
- Attribution costs no new data and no new step: a bead compares its own gate
  output against the baseline text.

**Negative:**
- A project must write and maintain an executable to use `bead-bonanza`, where
  before it wrote a Markdown bullet. The floor for a trivial project rises,
  though only to a script that runs commands it already runs.
- One more config file at the repository root, alongside `specs.toml`,
  `.beads/config.yaml` and `.claude/settings.json`.
- A checkpoint script is opaque to a reader who will not open it, where the
  `AGENTS.md` bullet was readable in place. Accepted: it is the project's own
  script, and prose that must also be machine-parsed is what this ADR retires.
- Every subagent prompt carries a `Known-red gates:` block even when empty,
  costing a line of context per dispatch to remove the ambiguity between
  "nothing known-red" and "not measured".
- Attribution is best-effort rather than exact. Comparing raw output finds a
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

---
title: "ADR-022: Projects Declare Checkpoint Scripts"
description: A project declares its gate commands as executable checkpoint
  scripts referenced from specs.toml, and those scripts report findings with
  paths rather than a bare exit code, because a prose gate profile cannot
  express a multi-command gate and a bare exit code cannot express a
  known-red baseline
slug: adr-022
type: adr
status: draft
updated_at: 2026-08-17T00:00:00Z
---

# ADR-022: Projects Declare Checkpoint Scripts

**Status:** draft
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

A project declares **checkpoint scripts** — executables it owns — through
optional fields in `specs.toml`, referenced by repository-root-relative path.
`bead-bonanza` runs them and reads their output; it parses no prose.

Two checkpoints, the ones `AGENTS.md` gives instructions for today:

- **`preflight`** — run as the last step of Phase 0.
- **`check`** — run after every bead's implementation, in that bead's
  worktree.

A checkpoint script reports **findings with paths**, not a bare verdict. Its
exit status is the verdict; each finding it emits carries a file path, so a
finding can be attributed to the bead that owns that file.

The run's known-red state is written once by pre-flight to
`$RUN_DIR/gate-baseline.md` — the gate baseline and the environmental-failure
ledger in one file, which is what `SKILL.md` already says the ledger is — and
a rendered summary is **inlined into every subagent prompt**, always, including
when it is empty.

Per-bead attribution of baseline findings belongs to `plan-sprint`, not to the
script and not to the orchestrator: at Phase 2.5 the planner already holds the
enumerated queue and each bead's component and file surface, and already
computes file-level overlap for wave packing.

**What this does not change.** A checkpoint finding is one of four kinds of
finding a run produces, and they are not interchangeable — conflating the
first two is what left the baseline undelivered:

| Finding | Channel | Read by | When |
|---|---|---|---|
| Pre-existing, owned by no queued bead | `$RUN_DIR/gate-baseline.md`, inlined in every prompt | every subagent | at dispatch |
| Pre-existing, owned by a queued bead | that bead's slice of the prompt, computed by `plan-sprint` | that bead only | at dispatch |
| Discovered mid-run, someone must act | a bead (`bd create --discovered-from`) | this run's orchestrator | immediately |
| Friction or observation | a `retro:` comment on the bead | a later retrospective | Phase 4 |

Only the first two are this ADR's subject. The third is the tracker, which is
live and queryable mid-run; the fourth is harvested after the run ends, which
is why it is the wrong channel for anything anyone must act on.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Keep parsing `AGENTS.md`, teach the parser fenced blocks | Smallest change; no project migration | Fixes one shape of a general defect; still cannot express per-surface tests, conditional gates, or a findings count; the parser silently takes a wrong span when prose precedes the command | Treats the instance, not the class. The parser would still return a string where the baseline needs a count |
| Discovery by `.specs/hooks/<name>` convention | Zero configuration; nothing to declare | Introduces a dotted directory one character from the existing `specs/` corpus root and meaning something entirely different; a magic directory is undiscoverable until it is documented, which is the problem being solved | The `specs/` / `.specs/` confusability is a standing trap for every reader of this repository |
| A strict machine-readable contract (JSON schema for findings) | Unambiguous; validatable | A schema the project must satisfy makes writing a checkpoint script a specification exercise; projects with a one-line lint gate would pay the full cost | Disproportionate. The contract is guidance plus a thin emission convention, below |
| Declare scripts in `specs.toml`, findings-with-paths, guideline contract | Explicit and greppable; no new directory; the project chooses where its scripts live; reuses the `BB_*` key=value emission convention `pre-flight-check.sh` already uses | `specs.toml` is nominally the spec-corpus config and now also carries build-adjacent declarations | N/A — selected |

## Consequences

**Positive:**
- A multi-command gate is expressible, because it is a script rather than a
  span of prose inside a bullet.
- The known-red baseline reaches the agent that runs the gate. A bead is told
  the pre-existing findings count, which subset is its to clear, and that the
  remainder is correct to leave.
- The environmental-failure ledger gets a home, implementing a rule `SKILL.md`
  already states.
- `AGENTS.md` sheds the gate profile, the targeted-tests list and the
  environmental ledger.
- Attribution costs no new data: `plan-sprint` already computes bead-to-file
  overlap for wave packing.

**Negative:**
- A project must write and maintain executables to use `bead-bonanza`, where
  today it writes a Markdown bullet. The floor for a trivial project rises.
- `specs.toml` carries declarations that are not about the spec corpus.
- A checkpoint script is opaque to a reader who will not open it, where the
  `AGENTS.md` bullet was readable in place. Accepted: it is the project's own
  script, and prose that must also be machine-parsed is what this ADR retires.
- Every subagent prompt carries a `Known-red gates:` block even when empty,
  costing a line of context per dispatch to remove the ambiguity between
  "nothing known-red" and "not measured".
- Projects using the `AGENTS.md` block need migrating; both paths must be
  supported until they are.

## References

- [[adr-019]] — skill-composed subagent orchestration; the planner this
  decision assigns attribution to
- [[adr-017]] — the Spec CLI as a plugin prerequisite; the precedent for a
  skill depending on an executable rather than branching on its absence
- `rfcs/inbox/bead-bonanza/rfc-self-acceptance-signals-and-gate-baseline.md`
  P2 — the fenced-gate parse failure this ADR supersedes
- `rfcs/inbox/bead-bonanza/rfc-self-worktree-composition-and-shared-gates.md`
  P3 — baseline ownership across concurrent beads, resized by this ADR

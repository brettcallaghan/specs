# Preserved bead-bonanza run artifacts

`bead-bonanza` writes each run's plan, decisions, per-bead progress and
per-bead retros to `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/`. That
location dies with the container, and the run id is generated at runtime and
recorded only as a line of stdout in the orchestrator's context — so after a
context loss the artifacts are both unreachable and, eventually, gone.

This directory is a committed copy, so the run survives both.

## What each file is

- `dispatch.md` — the sprint plan: wave table (bead, class, bucket, tier,
  model, push policy, reason), the gate profile, and the environmental
  ledger passed into every subagent prompt as fact.
- `decisions.md` — decisions surfaced before dispatch.
- `<bead>.progress.jsonl` — one line per phase, closed vocabulary:
  `claimed` → `implementing` → `gates-running` → `gates-passed` →
  `retro-written` → `committed` (with sha). The tail says exactly where a
  bead died.
- `<bead>.retro.md` — the recovery anchor, written after `gates-passed` and
  before the close. For beads that closed, this text is also folded into the
  bead description; for a bead that died first, this file is the only copy.

## Run 1786121603-ri (2026-08-07, reference-integrity)

Killed by a user interrupt at 18:33Z mid-batch on `9kb.4`+`9kb.5`, then dead
for 2h16m. `agent-plugins-9kb.5.progress.jsonl` stops at `implementing` and
has no retro — that bead is the one this copy exists for. Post-mortem and
proposals: `rfcs/inbox/bead-bonanza/rfc-self-run-dir-discoverability-and-orchestrator-recovery.md`.

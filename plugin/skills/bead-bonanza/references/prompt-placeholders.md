# Dispatch placeholders — what the orchestrator substitutes

**Orchestrator-facing.** None of this is sent to a subagent; it is what the
orchestrator fills the prompt templates in with. Read it once per run, then
work from the prompt file for the worker you are spawning:

- [prompt-per-bead.md](prompt-per-bead.md) — one bead, its own worktree
- [prompt-fast-lane.md](prompt-fast-lane.md) — a cluster of trivial beads
- [prompt-shared-surface.md](prompt-shared-surface.md) — beads sharing a surface

**Substitute every `{PLACEHOLDER}` before sending.** An unsubstituted
placeholder is the single most common dispatch defect, because the subagent
has no session context to infer it from.

## Resolved once per run

| placeholder | where it comes from |
|---|---|
| `{EPIC}` | the run epic `sprint-scope` resolved (emitted as `BB_EPIC`). Also the run id |
| `{BRANCH}` | the sprint branch, from the manifest |
| `{SCRIPTS}` | absolute path to this skill's `scripts/` directory |
| `{CHECKS_IDS}` | the ids the project declared (emitted as `BB_CHECKS_IDS`) |
| `{CHECKS_BASELINE_TEXT}` | the manifest's baseline — `bd show {EPIC} --json \| jq -r '.[0].metadata.checks_baseline'` |

A subagent can infer none of them.

**There is no run directory.** The run's record is the epic: the manifest in
its metadata, the amendments in its comments.

**`{EPIC}` is also where incidents land.** A subagent that has to decide
something changing what "correct" means writes
`retro:<kind>:<target> incident — …` on the bead and
`amend: incident on <id> — …` on the epic. `run-bead` carries the rule; the
prompt only has to supply the id, which no subagent can infer.

## Resolved per dispatch

| placeholder | scope |
|---|---|
| `{DISPATCH_ID}` | **per subagent** — one id for a batch, shared by every bead it holds, which is what makes "what is that agent holding" a query |
| `{WORKTREE}` | **per dispatch** — the absolute path to the worktree the orchestrator created for it |

A subagent works in `{WORKTREE}` and nowhere else. It cannot infer the path,
and a prompt that omits it sends the agent into the shared checkout, which
defeats every guarantee this design rests on. How those worktrees are created,
merged and removed is [worktree-mechanics.md](worktree-mechanics.md).

## Resolved per bead

`{ID}`, `{PRIORITY}`, `{TITLE}`, `{DECISION}`, `{IMPLICIT_NOTES}` — from the
bead itself and from the dispatch plan. The two batch templates take numbered
forms (`{ID_1}`, `{TITLE_1}`, …), one line per bead, in processing order.

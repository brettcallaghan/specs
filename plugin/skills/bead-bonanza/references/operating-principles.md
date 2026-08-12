# bead-bonanza Operating Principles

Principles referenced from SKILL.md's Operating principles list, in full.
Read the summary line in SKILL.md first; come here for the rule's complete
statement.

## Contents

- One commit per bead, always — now without exceptions, and how the bead's own
  worktree holds interrupted work
- Findings harvest — the `bd export | jq` pipeline and its non-zero assertion
- Guard against foreign dirt — narrowed to the orchestrator's own commits by
  worktree adoption, and why the shared-database race it used to be paired
  with is gone entirely

## One commit per bead, always

Every closed bead = exactly one commit on the branch. The handoff file is
written BEFORE the close ceremony (crash recovery anchor); the close happens
via a plain `bd update --status closed` BEFORE staging, so the close is
folded into the implementation commit. Closing never rewrites the
description — grooming owns it.

### No exceptions: the worktree holds interrupted work

**Retired by `agent-plugins-92q.30`** (one worktree per bead). The
`WIP <bead-id>` commit was the one sanctioned exception to one-commit-per-bead.
It existed because a half-finished bead's changes sat in a working tree that
the next bead needed clean — the only way to get them out of the way without
losing them was to commit them, and the only way to restore
one-commit-per-bead afterwards was a squash.

Under one worktree per bead there is nothing to get out of the way. An
interrupted bead's changes stay uncommitted in **its own** worktree, on its
own branch, where no other agent will ever look. The worktree's continued
existence is precisely the signal that the bead is unfinished.

**To resume an interrupted bead:** re-enter its worktree, finish the work, and
run the close ceremony once. There is no WIP commit to find, no squash, no
history rewrite, and therefore no unpushed-only safety boundary to observe.
One bead still means exactly one commit — now without an exception.

**WHEN the work must be surrendered rather than resumed** — the bead is
genuinely blocked, not merely interrupted — leave the worktree in place, mark
the bead `blocked` with a reason, and report it. Do not commit partial work to
"save" it: the worktree already saves it, and a partial commit on the bead's
branch would merge into the sprint branch as an incomplete bead.

## Findings harvest

Wiring findings capture into a run and harvesting what it wrote are two
different steps. The harvest is the one that's easy to get wrong: `bd list`
defaults to open issues, and a drained closure set is almost entirely
closed, so a literal reading returns findings for the handful of beads that
never got closed and silently drops nearly everything else.

### Wiring: unchanged, and it stays

If `rfcs/findings-capture.md` exists (retro-spec protocol) in the project,
every batch prompt includes: record `retro:`-prefixed `bd comment`s on the
bead being worked, 1-3 sentences, friction only, zero analysis. This
channel is the shared protocol across retro-spec, `groom-feature-blueprint`
and the work-order template — it stays wired regardless of how the
harvest below is done; do not remove it anywhere.

### Harvesting: one `bd export | jq` pipeline, not `bd list` + `bd show --json`

`bd show --json` omits comments entirely (measured on bd 1.1.2); `bd show`'s
text output prints them, but parsing that text is fragile — a bead's
Description or Design can itself quote the word `retro:` in prose, which a
naive `grep '^\s*retro:'` over the whole `bd show` output miscounts as a
comment. `bd export` sidesteps both problems: it carries comments as a
`comments` array on the issue record, covers every status in one pass (no
`--status closed` flag to remember — and note that repeating `--status`
overwrites rather than accumulates, so it cannot be used to combine
open+closed in one invocation), and needs no text parsing:

```
bd export | jq -r 'select(.status=="closed")
  | select((.labels//[]) | index("<slug>"))
  | .comments[]? | select(.text|startswith("retro:")) | .text'
```

Drop the `select(.status=="closed")` clause (or run the pipeline a second
time without it) to also catch findings on beads still open at aggregation
time — a drained run is mostly closed but not necessarily all closed.

**Assert the count is non-zero before aggregating — this is a required
step with a stated failure action, not advice.** A harvest that returns
zero findings across a completed feature is a broken query, not a clean
run: an aggregated retro built on an empty set is worse than none, because
it reads as a confident, clean-looking artifact instead of an obvious gap.
WHEN the count is zero against a feature known to carry `retro:` comments,
**stop and re-run the query** — check the label spelling and confirm with
`bd show` that the beads you expect really are closed — before aggregating
anything. Do not go looking at `.beads/issues.jsonl`: `bd export` writes
stdout from the database at call time, so there is no stale file in this path
(`agent-plugins-92q.31` untracked the exports entirely).

## Guard against foreign dirt

Two hazards used to live here, discussed together because they shared a
symptom rather than a cause. One worktree per bead (`agent-plugins-92q.30`)
removed the first for subagents and narrowed it to the orchestrator; retiring
the committed tracker (`agent-plugins-92q.31`) removed the second outright.

### Working-tree dirt: narrowed to the orchestrator, not deleted

**Subagents no longer face this hazard at all.** Each implements in its own
worktree, which contains that bead's changes and nothing else. There is no
foreign dirt to guard against, nothing to enumerate as untouchable, and no
need to name the paths a bead touched so the close can stage them — inside a
worktree, "everything here" is exactly this bead's work.

**The orchestrator still faces it, and the guard survives for it unchanged.**
The orchestrator works in the shared checkout, where a diagnosis agent's
uncommitted fixes, a regenerated fixture, or a half-applied edit can be sitting
alongside its own work. It therefore commits by explicit pathspec
(`git commit -- <files>`), exactly as before. The same applies on the
main-loop fallback path (SKILL.md), where the orchestrator implements every
bead in the shared checkout because there are no subagents to isolate.

### The shared-database race: RETIRED — the tracker is no longer a file

**Retired by `agent-plugins-92q.31`.** Both the race and
`agent-plugins-92q.17`'s deferral rule are gone, and it is worth recording why,
because the fix is not the one worktree adoption looked like it would provide.

The race was: `.beads/issues.jsonl` was a projection of a database every agent
shared, and the close ceremony regenerated it in full at every close. So any
tracker write the orchestrator made while a subagent sat between `claimed` and
`committed` landed inside *that subagent's* commit. Pathspec could not prevent
it, because the capture — the export — happened before the git index was ever
consulted. The rule that compensated: while any subagent held a claim, the
orchestrator confined itself to read-only `bd` queries and deferred its own
tracker writes to the wave boundary.

**Worktrees did not fix this**, and briefly looked as though they had. A
worktree isolates *files*; this race was never about files. Every worktree
resolves the same database through git's common directory — precisely why `bd`
works inside one — so a subagent's close still exported shared state and still
captured the orchestrator's concurrent write. It merely landed on a different
branch.

What fixed it was removing the file. The tracker is `origin refs/dolt/data`;
the exports are untracked; the close ceremony stages code only and pushes the
database itself. There is no shared file for one agent's write to ride into
another agent's commit, so there is nothing to defer. **The orchestrator may
groom a discovered bead, bake a decision, or wire a dependency edge at any
time, including mid-wave.**

### The claim-reset rule (unchanged)

The same "don't touch what another agent owns" principle applies to
tracker *claims*: a subagent may only reset a claim it made itself — a
bead found `in_progress` that this agent didn't claim belongs to a live
sibling. The orchestrator, never a subagent, owns reconciling claims
across waves.

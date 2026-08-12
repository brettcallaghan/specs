# bead-bonanza Operating Principles

Principles referenced from SKILL.md's Operating principles list, in full.
Read the summary line in SKILL.md first; come here for the rule's complete
statement.

## Contents

- One commit per bead, always — where an interrupted bead's work lives, and how
  to resume it or surrender it
- Findings harvest — how to wire `retro:` capture, and the one pipeline that
  harvests it without silently returning nothing
- Guard against foreign dirt — how the orchestrator and a subagent each stage a
  commit, when a tracker write may happen, and whose claims an agent may reset

## One commit per bead, always

Every closed bead = exactly one commit on the branch. The handoff file is
written BEFORE the close ceremony (crash recovery anchor); the close happens
via a plain `bd update --status closed` BEFORE staging, so the close is
folded into the implementation commit. Closing never rewrites the
description — grooming owns it.

### No exceptions: the worktree holds interrupted work

An interrupted bead's changes stay **uncommitted in its own worktree**, on its
own branch, where no other agent will ever look. The worktree's continued
existence is precisely the signal that the bead is unfinished. There is no
`WIP` commit to make and nothing to squash.

**To resume an interrupted bead:** re-enter its worktree, finish the work, and
run the close ceremony once.

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

### Wiring

If `rfcs/findings-capture.md` exists (retro-spec protocol) in the project,
every batch prompt includes: record `retro:`-prefixed `bd comment`s on the
bead being worked, 1-3 sentences, friction only, zero analysis. This
channel is the shared protocol across retro-spec, `groom-feature-blueprint`
and the work-order template — keep it wired regardless of how the harvest
below is done; do not remove it anywhere.

### Harvesting: one `bd export | jq` pipeline, not `bd list` + `bd show --json`

`bd show --json` omits comments entirely (bd 1.1.2); `bd show`'s text output
prints them, but parsing that text is fragile — a bead's Description or Design
can itself quote the word `retro:` in prose, which a naive `grep '^\s*retro:'`
over the whole `bd show` output miscounts as a comment. `bd export` sidesteps
both problems: it carries comments as a `comments` array on the issue record,
covers every status in one pass (no `--status closed` flag to remember — and
note that repeating `--status` overwrites rather than accumulates, so it cannot
be used to combine open+closed in one invocation), and needs no text parsing:

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
stdout from the database at call time, so there is no stale file in this path.

## Guard against foreign dirt

### Working-tree dirt: the orchestrator's hazard, not a subagent's

**A subagent does not face this hazard.** It implements in its own worktree,
which contains that bead's changes and nothing else. There is no foreign dirt
to guard against, nothing to enumerate as untouchable, and no need to name the
paths a bead touched so the close can stage them — inside a worktree,
"everything here" is exactly this bead's work.

**The orchestrator does face it.** It works in the shared checkout, where a
diagnosis agent's uncommitted fixes, a regenerated fixture, or a half-applied
edit can be sitting alongside its own work. It therefore commits by explicit
pathspec (`git commit -- <files>`). The same applies on the main-loop fallback
path (SKILL.md), where the orchestrator implements every bead in the shared
checkout because there are no subagents to isolate.

### Tracker writes are never deferred

The tracker is a database synced to `origin refs/dolt/data`, not a file in the
working tree, and the close ceremony stages code only. No tracker write can
ride into another agent's commit, so **the orchestrator may groom a discovered
bead, bake a decision, or wire a dependency edge at any time, including
mid-wave.**

Worktrees are not what makes this safe, and must not be relied on for it: every
worktree resolves the same database through git's common directory — precisely
why `bd` works inside one.

### The claim-reset rule

The same "don't touch what another agent owns" principle applies to
tracker *claims*: a subagent may only reset a claim it made itself — a
bead found `in_progress` that this agent didn't claim belongs to a live
sibling. The orchestrator, never a subagent, owns reconciling claims
across waves.

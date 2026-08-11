# bead-bonanza Operating Principles

Principles referenced from SKILL.md's Operating principles list, in full.
Read the summary line in SKILL.md first; come here for the rule's complete
statement.

## Contents

- One commit per bead, always — now without exceptions, and how the bead's own
  worktree holds interrupted work
- Findings harvest — the `bd export | jq` pipeline and its non-zero assertion
- Guard against foreign dirt, and the shared-database race — which of the two
  worktrees fix, and why isolation cannot fix a tracker write landing in
  another agent's commit

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
**stop and re-run the query** — check the label spelling and confirm
`.beads/issues.jsonl` reflects the latest closes — before aggregating
anything.

## Guard against foreign dirt, and the shared-database race

Two different hazards live here. **One worktree per bead
(`agent-plugins-92q.30`) removes the first for subagents and narrows it to the
orchestrator; it does NOT touch the second.** Read that distinction carefully
before deleting anything here — they used to be discussed together because
they had the same symptom, not the same cause.

### Working-tree dirt: narrowed to the orchestrator, not deleted

**Subagents no longer face this hazard at all.** Each implements in its own
worktree, which contains that bead's changes and nothing else. There is no
foreign dirt to guard against, nothing to enumerate as untouchable, and no
need to name the paths a bead touched so the close can stage them — inside a
worktree, "everything here" is exactly this bead's work.

**The orchestrator still faces it, and the guard survives for it unchanged.**
The orchestrator works in the shared checkout, where a diagnosis agent's
uncommitted fixes, a stale export, or a half-applied edit can be sitting
alongside its own work. It therefore commits by explicit pathspec
(`git commit -- <files>`), exactly as before. The same applies on the
main-loop fallback path (SKILL.md), where the orchestrator implements every
bead in the shared checkout because there are no subagents to isolate.

### The shared-database race: worktrees do NOT fix it either

**Do not delete this rule on the strength of worktree adoption.** It is the
one hazard in this document that isolation cannot touch, and the reason is
worth stating plainly: a worktree isolates *files*, and this race is not
about files. Every worktree resolves the same tracker database through git's
common directory — that is exactly why `bd` works from inside one. So a
subagent's close, running in its own worktree on its own branch, still
exports the **shared** database, and still captures whatever the orchestrator
wrote to it moments earlier. The write lands in the subagent's commit as
surely as it did in a shared tree; it just lands on a different branch now.

This is `agent-plugins-92q.31`'s to fix, by making the tracker stop being a
committed file at all. Until that lands, the deferral rule below stands
verbatim.

The tracker ledger (`.beads/issues.jsonl`) is a projection of a database
every agent shares, not a file one agent owns. `../scripts/bb-close-bead.sh`
regenerates it in full — a plain `bd export` — at every close, so any
tracker write the orchestrator makes while a subagent is between `claimed`
and `committed` lands inside *that subagent's* commit. Pathspec cannot
prevent this, because the capture (the export) happens before the git
index is ever consulted — the orchestrator can commit `-- <files>` with
maximal discipline and still find "nothing to commit," because the
subagent's close already carried the write.

**WHILE any subagent holds a claim, the orchestrator SHALL confine itself
to read-only `bd` queries and defer its own tracker writes** — grooming a
newly discovered bead, baking a decision, wiring a dependency edge — until
the wave returns. **WHERE a write cannot wait,** make it and note in the
report which bead's commit will carry it, so a reader of that commit is
not surprised to find another actor's edit inside it.

**The trade.** Deferring can make a wave marginally slower — a discovered
bead's grooming waits for the wave boundary instead of happening the
moment it's found. That is the intended trade, not an oversight; the
WHERE clause above is the escape hatch for a write that genuinely cannot
wait, not a license to defer nothing.

**Considered and rejected: a `git status` check before committing.** This
does not fix the race — the window between the check and the commit is
exactly where the close ceremony lands its export. A run that checked
`git status` clean immediately before committing still landed another
actor's write, because `../scripts/bb-close-bead.sh`'s export ran, and its commit
landed, inside that same window. Ordering discipline on the
orchestrator's side cannot close a race that happens on the subagent's
side of the same shared file.

### The claim-reset rule (unchanged)

The same "don't touch what another agent owns" principle applies to
tracker *claims*: a subagent may only reset a claim it made itself — a
bead found `in_progress` that this agent didn't claim belongs to a live
sibling. The orchestrator, never a subagent, owns reconciling claims
across waves.

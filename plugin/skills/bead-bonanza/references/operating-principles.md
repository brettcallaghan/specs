# bead-bonanza Operating Principles

Principles referenced from SKILL.md's Operating principles list, in full.
Read the summary line in SKILL.md first; come here for the rule's complete
statement and its one sanctioned exception.

## Contents

- One commit per bead, always — the WIP-commit exception and how to resume one
- Findings harvest — the `bd export | jq` pipeline and its non-zero assertion
- Guard against foreign dirt, and the shared-database race — why pathspec
  can't fix a tracker write landing in another agent's commit

## One commit per bead, always

Every closed bead = exactly one commit on the branch. The handoff file is
written BEFORE the close ceremony (crash recovery anchor); the close happens
via a plain `bd update --status closed` BEFORE staging, so the close is
folded into the implementation commit. Closing never rewrites the
description — grooming owns it.

### The one sanctioned exception: an interrupted bead's WIP commit

**WHEN a bead is interrupted mid-implementation** and its work must be
persisted before the agent stops, commit it with a subject prefixed
`WIP <bead-id>` and leave the bead `in_progress`. A WIP commit is the one
sanctioned exception, and it is **not** a close: no `bd update --status
closed`, no retro, no committed progress event. Emit
`bb-progress.sh "$RUN_DIR" <id> blocked reason="interrupted; WIP <sha>"` so
the ledger records where it stopped.

The agent that resumes the bead finishes the work, then squashes its own
commit together with every `WIP <bead-id>` commit on the branch into the
single landing commit before running the close ceremony. Record the
interruption and the squash in the bead's retro.

**How to find prior WIP commits.** `git log --oneline --grep="^WIP
<bead-id>"` on the branch. State it, because the resuming agent may be a
different agent with no knowledge that a WIP exists — which is the whole
situation being handled.

**The safety boundary.** Squashing rewrites history, so it is permitted only
while the WIP commits are unpushed. WHERE a WIP commit has already been
pushed, do not rewrite it: land the completion as a normal second commit,
note the two-commit landing in the retro, and report it in Phase 4. The
existing prohibition on force-pushing a published branch
(`subagent-prompts.md`, Constraints) is not weakened by this
exception.

**`../scripts/bb-close-bead.sh` does not reject a prior WIP commit.** It closes,
exports, stages, and commits based on the currently-staged diff — it never
inspects `git log`, so a `WIP <bead-id>` commit already on the branch does
not block it. The resuming agent still must squash by hand first: the script
gives you a normal close, not a squash.

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

Multiple agents can share the tree: a diagnosis agent's uncommitted fixes,
a live subagent's staged `.beads/issues.jsonl`. Two different hazards
live here, and only one of them is fixed by pathspec discipline.

### Working-tree dirt: pathspec discipline still works

Batch prompts must name any known foreign files as untouchable. The
orchestrator commits its own work by explicit pathspec (`git commit --
<files>`) so foreign working-tree changes are not accidentally swept into
its commit. This defence is correct and unchanged for this hazard.

### The shared-database race: pathspec cannot fix it

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

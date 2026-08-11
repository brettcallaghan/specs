# bead-bonanza Operating Principles

Principles referenced from SKILL.md's Operating principles list, in full.
Read the summary line in SKILL.md first; come here for the rule's complete
statement and its one sanctioned exception.

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

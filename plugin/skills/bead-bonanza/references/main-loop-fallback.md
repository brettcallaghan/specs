# bead-bonanza Main-Loop Fallback

WHEN no `Agent` tool is available to this session, the whole skill's shape
changes: nothing can be delegated — not the per-bead work, not the Phase 1.5
grooming, not the Phase 2.5 planning. This file is that degraded mode. It is a
low-probability, read-once branch, so read it only WHEN the check below comes
back negative.

**Check that before Phase 3, not during it:** some harnesses expose no `Agent`
tool at all, and some sessions forbid it. Whether you have one is a harness
property, invisible from a shell, so no pre-flight script can measure it — you
must. WHEN you do have one, ignore this file entirely.

WHEN you do not, the run does NOT abort — it degrades to a defined mode:

- **Implement every bead in the orchestrator's own loop**, in queue
  order, following the full per-bead ceremony unchanged (claim, gates,
  retro, atomic close, single commit). The ceremony is what makes a bead
  landing correct; delegation is only how the work is distributed.
- **Drop batching, and work in the shared checkout.** Batching exists to
  amortize full test runs across a subagent's beads; with one agent there
  is nothing to amortize across, so it adds risk and no benefit. Worktrees
  likewise buy nothing here — they isolate agents from each other, and
  there is only one agent — so the orchestrator implements each bead on
  the sprint branch directly. Run sequential. **This is the one path on
  which the shared-tree cautions still apply**: commit by explicit
  pathspec, because a diagnosis agent's edits or a stale export can be
  sitting in that tree alongside your own work.
- **Plan the sprint yourself.** The `plan-sprint` skill is invoked in a
  subagent in the normal path; here there is no subagent, so run its workflow
  in your own loop. `dispatch.md` is still written on every run, still
  enumerates beads rather than naming a filter, and still carries the manifest
  you measured — none of that is what delegation bought.
- **Groom in your own loop too**, for the same reason, and hold Phase 1.5's
  gate at the same bar: a bead that cannot be brought to Ready is bounced, not
  guessed at.
- **Context is now the scarce resource**, because nothing is being
  isolated for you. Keep per-bead notes to one line, and re-read the run
  checklist rather than holding it in your head.
- **No tier routing occurs.** You cannot dispatch at a resolved model
  when you are the one doing the work — every bead runs at the session's
  own model regardless of its `tier`.
- **Say so once, at the top of the run**, so the user knows the tier
  labels were not honored.

This is a real degradation, not an equivalent path: an under-tiered bead
that a `frontier` subagent would have handled may now fail, and a
`light` bead costs full session model. Both are acceptable; concealing
them is not.

Phase 4 reports this differently from a clean run: write "Tier escalations:
N/A — no routing occurred" rather than `none`, exactly as SKILL.md's Phase 4
section specifies. `none` claims routing happened and found nothing to fix;
this run did no routing at all.

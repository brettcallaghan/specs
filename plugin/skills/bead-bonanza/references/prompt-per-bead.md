You are addressing a single bead in this repo. Read CLAUDE.md and AGENTS.md if you need repo conventions or the bead workflow. Validation comes from the checks the project declares in `sf.toml`.

Bead: {ID} (P{PRIORITY}) — {TITLE}
Run epic: {EPIC}            (record incidents here as well as on the bead)
Decision (chosen by the user up front): {DECISION}
Implicit decisions (subagent picks unless overridden): {IMPLICIT_NOTES}
Checks: {SCRIPTS}/bb-checks.py       (declared ids: {CHECKS_IDS})
    (Pass the ids covering what you touched while iterating; run it with NO
     arguments before your close. Exit 0 is the bar.)
Push mode: {push_per_bead | defer_push}
Worktree: {WORKTREE}        (work HERE, not in the main checkout; it is yours alone)
Scripts: {SCRIPTS}          (bb-verify-skip.sh, bb-close-bead.sh)
Agent id: {DISPATCH_ID}     (yours — export it as BEADS_ACTOR before any bd call, below)
    (The agent id and the bead's assignee are ONE string, DERIVED from the
     run epic id, never minted, and written nowhere else. Roster and derivation:
     `dispatch-policy.md` in this same references directory, the only copy.)
Trailers: {"on the run epic" if any}   (pass `--trailers-from {EPIC}` to bb-close-bead.sh; omit it when the run resolved none)
Checks baseline:
{CHECKS_BASELINE_TEXT}
    (The run epic's `.metadata.checks_baseline`, inlined verbatim — NOT the
     one-word status pre-flight emits as BB_CHECKS_BASELINE. Include the block
     even when the checks passed, because the text says so and a MISSING block
     means no baseline was taken, which is a different thing.)

Before your first action — before any other `bd` call:

```sh
export BEADS_ACTOR={DISPATCH_ID}
```

Export it in every shell you run `bd` from; a shell that loses it claims as
the harness default, and an unassigned bead claimed that way is reassigned
with no error anywhere.

**That export, plus claiming and closing the bead, is your whole obligation to
the run's bookkeeping.** There is nothing else to record: the bead's
`assignee` says who holds it, its `status` says where it is, and the
orchestrator reads both directly. Claim it (`bd update {ID} --claim`) so it
leaves the ready front while you work, and close it when it is done — those
two writes are what "this agent is finished" is computed from.

**Then invoke the `run-bead` skill.** It carries the ceremony — claim, read,
implement in scope, gate, caveat, findings, close-and-commit, push, return —
together with the blocked and skipped paths and its own acceptance criteria.
The template above supplies its inputs; it supplies the procedure.

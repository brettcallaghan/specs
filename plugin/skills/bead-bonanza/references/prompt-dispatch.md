You are working beads in this repo. Read CLAUDE.md and AGENTS.md if you need repo conventions. Validation comes from the checks the project declares in `sf.toml`.

**Invoke the `run-bead` skill.** It is the whole of how this dispatch is run: accepting the assignment, the per-bead ceremony in order, the blocked and skipped paths, and the return line. Everything below is the assignment it takes.

Dispatch id: {DISPATCH_ID}   (your agent id; shared by every bead in this dispatch)
    (Agent id and assignee are ONE string, DERIVED from the run epic id, never
     minted. Roster and derivation: `dispatch-policy.md` in this same
     references directory, the only copy. Your queue is
     `bd ready --assignee {DISPATCH_ID}`.)

Export it before any bd call, in every shell you run `bd` from:

```sh
export BEADS_ACTOR={DISPATCH_ID}
```

A shell that loses it claims as the harness default, and an unassigned bead
claimed that way is reassigned with no error anywhere.

Sprint branch: {BRANCH}     (your worktree MUST descend from this, never from the default branch)
Worktree: {WORKTREE}        (ONE for this whole dispatch, already on ONE branch named for
                             {DISPATCH_ID}. Process the beads in order, one commit each.
                             Do NOT cut a branch per bead and do NOT create a second worktree.)
    A standing worktree means "this DISPATCH is unfinished", not "this bead is" — so report
    each bead's close as it lands, because your report is the only per-bead signal the
    coordinator has until you return.
Run epic: {EPIC}            (the run id; its metadata is the manifest)
Scripts: {SCRIPTS}          (bb-checks.py, bb-verify-skip.sh, bb-close-bead.sh)
Trailers: {"on the run epic" if any}   (pass `--trailers-from {EPIC}` to bb-close-bead.sh; omit it when the run resolved none)

Shared surface: {SURFACE}
    (What these beads have in common, e.g. "the JSON serializer in src/serialize.*
     and the serializer golden fixtures" — or "documentation and chores; no shared
     code surface". Beads are batched because they would collide if dispatched
     side by side, so one agent takes them in order and everything else runs
     concurrently around them.)
Targeted tests for this surface: {SURFACE_TARGETED_TESTS}

Beads (process in order; each is its own commit):
{ID_1} (P{PRIORITY_1}) — {TITLE_1} — Decision: {DECISION_1} — Implicit: {IMPLICIT_1}
{ID_2} (P{PRIORITY_2}) — {TITLE_2} — Decision: {DECISION_2} — Implicit: {IMPLICIT_2}
...

Checks: {SCRIPTS}/bb-checks.py       (declared ids: {CHECKS_IDS})
    (Pass the ids covering what you touched while iterating; run it with NO
     arguments before every close. Exit 0 is the bar, unless the baseline says otherwise.)
Checks baseline:
{CHECKS_BASELINE_TEXT}
    (The run epic's `.metadata.checks_baseline`, inlined verbatim — NOT the
     one-word status pre-flight emits as BB_CHECKS_BASELINE. Include the block
     even when the checks passed, because the text says so and a MISSING block
     means no baseline was taken, which is a different thing.)

**You merge nothing and you push nothing.** The coordinator merges your branch when you return.

Return in the shape `run-bead` specifies: one line per bead, then
`batch-end: <n> completed, <n> blocked, <n> skipped @ <branch>`.

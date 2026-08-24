## Shared batch preamble

Every batch prompt opens with this block. **The orchestrator concatenates it
plus the batch's own template into the one text it sends** — a batch prompt
assembled without it is missing the `sprint-batch` invocation and the return
contract, with nothing to catch the loss downstream.

It is duplicated rather than cross-referenced so each batch prompt is readable
and pasteable on its own. `prompt-batch-preamble.md` is the authored copy and
the only one to edit; `prompt-fast-lane.md` and `prompt-shared-surface.md`
carry generated copies between their BEGIN/END block markers, and
`scripts/sync-shared-references.sh --check` fails when one drifts.

**The ceremony is not in these prompts.** `sprint-batch` carries accepting the
assignment (clean tree, base ref, manifest, dispatch record), running
`run-bead` per bead in order, and the return line. The templates supply only
what a subagent cannot infer — which beads, which surface, which push mode —
plus the handful of rules specific to that batch's kind.

```
Read CLAUDE.md and AGENTS.md if you need conventions; validation comes from the checks the project declares in `sf.toml`.

FIRST, invoke the `sprint-batch` skill. It is the whole of how a batch is run: accepting the assignment, run-bead per bead in order, the dispatch record, and the return line. Everything below is the assignment it takes.

Dispatch id: {DISPATCH_ID}   (yours; shared by every bead in this batch)
Sprint branch: {BRANCH}      (your worktree MUST descend from this, never from the default branch)
Worktree: {WORKTREE}        (ONE for this whole batch — process the beads in order,
`git checkout -b <suffix>` per bead so each branches off the last, one commit each,
and do NOT create a second worktree.)
A standing worktree means "this BATCH is unfinished", not "this bead is" — so report
each bead's close as it lands, because your report is the only per-bead signal the
orchestrator has until you return.
Run epic: {EPIC}            (the run id; its metadata is the manifest — and where incidents land)
Scripts: {SCRIPTS}
Trailers: {"on the run epic" if any}   (pass `--trailers-from {EPIC}` to bb-close-bead.sh; omit it when the run resolved none)
Checks baseline:
{CHECKS_BASELINE_TEXT}
    (The run epic's `.metadata.checks_baseline`, inlined verbatim — NOT the
     one-word status pre-flight emits as BB_CHECKS_BASELINE. Include the block
     even when the checks passed, because the text says so and a MISSING block
     means no baseline was taken, which is a different thing.)

Pass each bead's `--trailers-from {EPIC}` through to the close script (omit that flag when the run resolved no trailers). Read each committed sha with `git rev-parse --short HEAD` rather than from the script's printout. Run the template-specific steps below alongside `sprint-batch`'s per-bead loop.

Return in the shape `sprint-batch` specifies: one line per bead, then `batch-end: <n> completed, <n> blocked, <n> skipped @ <branch>`.
```

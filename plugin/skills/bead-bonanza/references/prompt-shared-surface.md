# Shared-surface batch prompt

**Beads that touch real code and share a surface, handed to ONE subagent.**
They are batched because they would collide if dispatched side by side — later
beads rebuild the same surface — so one agent takes them in order and the rest
of the wave runs concurrently around them. Trivial beads go to
[prompt-fast-lane.md](prompt-fast-lane.md) instead.

Fill every `{PLACEHOLDER}` from
[prompt-placeholders.md](prompt-placeholders.md); create and later remove the
worktree per [worktree-mechanics.md](worktree-mechanics.md). Send the preamble
below **followed by** the template below it, as one text.

<!-- BEGIN GENERATED BLOCK: plugin/skills/bead-bonanza/references/prompt-batch-preamble.md — DO NOT EDIT to the END marker; regenerate with scripts/sync-shared-references.sh -->
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
<!-- END GENERATED BLOCK -->

## The shared-surface template

```
You are processing {N} related beads that share a code/fixture surface ({SURFACE_DESCRIPTION}, e.g. "the JSON serializer in src/serialize.* and the serializer golden fixtures").

Beads (process in order; each is its own commit):
{ID_1} (P{PRIORITY_1}) — {TITLE_1} — Decision: {DECISION_1} — Implicit: {IMPLICIT_1}
{ID_2} (P{PRIORITY_2}) — {TITLE_2} — Decision: {DECISION_2} — Implicit: {IMPLICIT_2}
...

Checks: {SCRIPTS}/bb-checks.py       (declared ids: {CHECKS_IDS})
Targeted tests for this surface: {SURFACE_TARGETED_TESTS}
Push mode: {push_per_bead | defer_push}

These beads are NOT trivial — they touch real code. They are batched because they would collide if dispatched side by side — later beads rebuild the same surface — so one agent takes them in order and the rest of the wave runs concurrently around them. **Batching changes who does the work, not what gets validated: every bead still runs the full declared set before its own close.**

Beyond `sprint-batch`'s per-bead loop:

1. Before each bead's close, in this order — **none of it is optional**:
   - If generated output was touched, regenerate it with the project's own tooling — the repo's conventions doc names that command; this prompt does not carry it — then run the relevant fixture/snapshot test.
   - **Run the surface-relevant targeted tests** named above for the file(s) you touched. If none is listed for a surface you touched, find the test that covers it; if there's still nothing, that's a signal the change is broader than expected — pause and reconsider scope, don't take it as permission to skip.
   - **Run `{SCRIPTS}/bb-checks.py` with no arguments.** `run-bead` step 4 requires the full declared set per bead, and this batch does not override it.

   **Why this matters:** without per-bead targeted tests, a regression introduced in bead #2 only surfaces at batch end, after beads #3 and #4 may have built on the broken assumption. You then bisect, rework #2, and possibly redo #3 and #4. Targeted tests prevent that cascade.
2. Compare every run against the checks baseline above: a line in both was already failing, a line only in yours belongs to the bead you are on. Because every bead runs the checks, there is nothing to bisect — the bead whose run first shows a finding is where it entered. Fix it there before starting the next bead. WHERE investigating shows an earlier bead is actually at fault, follow the Blocked path for THAT bead rather than silently repairing it across several commits.
   **The push is the gate**: the last bead's passing run is what clears this batch, and whether CI also runs somewhere is the project's business, not a reason to skip it here.

## Constraints

- Same per-bead constraints as `run-bead`'s Constraints section. `sprint-batch` covers one-commit-per-bead and the in-batch conflict path.
```

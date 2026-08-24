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

Shared surface: {SURFACE}
    (What these beads have in common, e.g. "the JSON serializer in src/serialize.*
     and the serializer golden fixtures" — or "documentation and chores; no shared
     code surface". The beads are batched because they would collide if dispatched
     side by side, so one agent takes them in order and the rest of the wave runs
     concurrently around them.)
Targeted tests for this surface: {SURFACE_TARGETED_TESTS}

Beads (process in order; each is its own commit):
{ID_1} (P{PRIORITY_1}) — {TITLE_1} — Decision: {DECISION_1} — Implicit: {IMPLICIT_1}
{ID_2} (P{PRIORITY_2}) — {TITLE_2} — Decision: {DECISION_2} — Implicit: {IMPLICIT_2}
...

Checks: {SCRIPTS}/bb-checks.py       (declared ids: {CHECKS_IDS})
Push mode: {push_per_bead | defer_push}
Checks baseline:
{CHECKS_BASELINE_TEXT}
    (The run epic's `.metadata.checks_baseline`, inlined verbatim — NOT the
     one-word status pre-flight emits as BB_CHECKS_BASELINE. Include the block
     even when the checks passed, because the text says so and a MISSING block
     means no baseline was taken, which is a different thing.)

**Batching changes who does the work, not what gets validated: every bead still runs the full declared set before its own close.**

Beyond `sprint-batch`'s per-bead loop, before each bead's close, in this order — **none of it is optional**:

1. If generated output was touched, regenerate it with the project's own tooling — the repo's conventions doc names that command; this prompt does not carry it — then run the relevant fixture/snapshot test.
2. **Run the targeted tests** named above for the file(s) you touched. If none is listed for a surface you touched, find the test that covers it. If there is still nothing, that is a signal the change is broader than expected — pause and reconsider scope, don't take it as permission to skip. WHERE the change is genuinely untestable (a pure docs edit, a comment with zero output diff), say so explicitly in a `retro:` comment on the bead and continue; "I couldn't find a targeted test" without explanation is not acceptable.
3. **Run `{SCRIPTS}/bb-checks.py` with no arguments.** `run-bead` step 4 requires the full declared set per bead, and being in a batch does not override it — an unvalidated commit is not cheaper, it is just unvalidated.

   **Why the per-bead targeted tests matter:** without them, a regression introduced in bead #2 only surfaces at batch end, after beads #3 and #4 may have built on the broken assumption. You then bisect, rework #2, and possibly redo #3 and #4.

4. Compare every run against the checks baseline above: a line in both was already failing, a line only in yours belongs to the bead you are on. Because every bead runs the checks, there is nothing to bisect — the bead whose run first shows a finding is where it entered. Fix it there before starting the next bead. WHERE investigating shows an earlier bead is actually at fault, follow the Blocked path for THAT bead rather than silently repairing it across several commits.

   **The push is the gate**: the last bead's passing run is what clears this batch, and whether CI also runs somewhere is the project's business, not a reason to skip it here.

Read each committed sha with `git rev-parse --short HEAD` rather than from the close script's printout.

## Constraints

Same per-bead constraints as `run-bead`'s Constraints section. `sprint-batch` covers one-commit-per-bead and the in-batch conflict path. In `defer_push` mode, do not push at all — the orchestrator pushes after the batch.

Return in the shape `sprint-batch` specifies: one line per bead, then `batch-end: <n> completed, <n> blocked, <n> skipped @ <branch>`.

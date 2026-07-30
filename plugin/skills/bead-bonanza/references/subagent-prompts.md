# bead-bonanza Subagent Prompt Templates

The prompts the orchestrator fills in and dispatches. Read this at Phase 3
(dispatch), and at Phase 1.5 when a bead needs grooming first. Substitute
every `{PLACEHOLDER}` before sending — an unsubstituted placeholder is the
single most common dispatch defect, because the subagent has no session
context to infer it from.

`${TMPDIR:-/tmp}` below is the run root resolved in Phase 0 step 4. Use the
resolved `$RUN_DIR` value, not the literal.

## Contents

- Grooming subagent (Phase 1.5) — bring under-specified beads to Ready
- Per-bead subagent (sequential and worktree modes) — the canonical ceremony
- Doc/chore fast-lane subagent — a cluster of trivial beads, one subagent
- Shared-surface batch subagent — beads sharing a code/fixture surface
- Worktree parallel notes (orchestrator) — spawning and merging a wave

The per-bead template is the canonical one; the fast-lane and shared-surface
templates reference its workflow rather than restating it, so read it first.

---

## Grooming subagent prompt template

```
You are grooming {N} beads to the Definition of Ready so an orchestrator can dispatch each to an isolated, zero-context sub-agent. Use the `groom-beads` skill and follow its full workflow (Definition of Ready, dependency wiring, tier assignment, C4 blast radius, validate). It loads the spec-owned Definition of Ready itself — do not go looking for that file by path.

Beads to groom:
{ID_1} — {TITLE_1}
{ID_2} — {TITLE_2}
...

For each bead: bring it to Ready by editing it in place with `bd update` — enrich thin descriptions rather than discarding author intent; wire only dependencies you can justify; assign a `tier` with a `tier_reason`.

Do NOT implement any bead. Do NOT invent IDs, component names, or dependencies. Bounce, don't guess: any bead you cannot bring to Ready, leave as-is and report NOT-READY with the specific gap a human must close.

Return a report, one line per bead:
- `{ID}: ready — <notable change: description made context-independent / split into epic <id> + children <ids> / model assigned>`
- `{ID}: NOT-READY - <the gap that needs a human>`
```


---

## Per-bead subagent prompt template (sequential and worktree modes)

```
You are addressing a single bead in this repo. Read CLAUDE.md and AGENTS.md if you need repo conventions or the bead workflow. The gate commands are in AGENTS.md under "## bead-bonanza gates".

Bead: {ID} (P{PRIORITY}) — {TITLE}
Decision (chosen by the user up front): {DECISION}
Implicit decisions (subagent picks unless overridden): {IMPLICIT_NOTES}
Gates: format=`{FORMAT}` lint=`{LINT}` build=`{BUILD}` test=`{TEST}` regen=`{REGEN or "none"}` custom=`{CUSTOM_GATES or "none"}`
Push mode: {push_per_bead | defer_push}
Progress file: ${TMPDIR:-/tmp}/claude/bead-bonanza/{RUN_ID}/{ID}.progress.jsonl
Retro file: ${TMPDIR:-/tmp}/claude/bead-bonanza/{RUN_ID}/{ID}.retro.md

## Workflow — follow in order, do not deviate

Append a JSONL event to the progress file at each labeled step (use `printf '%s\n' '<json>' >> <file>`).

1. Emit `{"phase":"claimed"}`. Claim: `bd update {ID} --status in_progress`
2. Read full details: `bd show {ID}`
3. Emit `{"phase":"implementing"}`. Implement the change. **If this bead fixes a defect or regression, first write (or locate) a test that reproduces it and observe it FAIL for the stated reason BEFORE you fix it** — a test written after the fix encodes the fix's own assumptions and can stay green while the defect is live. Record the observed failure; it goes in the retro. Stay strictly within this bead's scope. If you find work beyond it, file a new bead via `bd create` with a `discovered-from` link to {ID}, noting in one line whether it *blocks* this bead (a prerequisite, an acceptance-criteria decomposition, or an expected test-coverage gap) or is merely adjacent — the orchestrator decides whether it enters the run. Do NOT implement it or pull it into your own scope. If this bead genuinely cannot be finished without it, take the Blocked path naming the new bead as the blocker; otherwise continue with the original — do NOT expand scope inline. For any implicit decision you must make, record what you chose (and why) — you'll put it in the retro.
4. Emit `{"phase":"gates-running"}`. **The push is the gate** — this is a single-bead push, so full gates run here before pushing. If the change affects generated output the tests compare against (codegen, snapshots, golden files, schemas) AND the gate profile defines a regen step, run `{REGEN}` BEFORE the test step.
   - `{FORMAT}`
   - `{LINT}`
   - `{BUILD}`
   - `{TEST}`
   - Any custom gates from the profile: `{CUSTOM_GATES}` (skip if none)
   On gate failure, fix the root cause and rerun. After two full failed cycles, follow the Blocked path — do not keep retrying. Emit `{"phase":"gates-passed"}` after success.
5. **Write the Retro to a file BEFORE closing the bead** (crash recovery anchor — non-negotiable).
   - Write `${TMPDIR:-/tmp}/claude/bead-bonanza/{RUN_ID}/{ID}.retro.md` containing the FULL bead description (original + a `## Retro` section appended). The Retro section covers:
     - Approach taken (which option, if a decision was made)
     - **For a defect/bug bead: the observed pre-fix failure** — the test and how it failed *before* the change (e.g. "test X failed with Y"), not merely "tests pass now"
     - **Implicit decisions you made** (one bullet each: what you chose + why)
     - Anything that differed from the original instructions
     - Anything surprising; ideas for improving prompts/skills/patterns
   - Emit `{"phase":"retro-written"}` after the file is on disk.
6. Atomic close — write description + status in ONE call, sourcing the retro file:
   ```
   bd update {ID} --status closed --body-file ${TMPDIR:-/tmp}/claude/bead-bonanza/{RUN_ID}/{ID}.retro.md
   ```
   Do NOT run `bd close` — `bd update --status closed` does both. This avoids a second commit.
7. **Refresh the durable tracker export, then stage.** `.beads/issues.jsonl` is a passive export — `bd` does not rewrite it on every write (Dolt auto-commit defaults off), so a closure that is not exported is lost if the container is reclaimed. Run `bd export` and make sure the result lands in `.beads/issues.jsonl` (redirect if your build of `bd` writes to stdout: `bd export > .beads/issues.jsonl`), and confirm the file actually changed. Then stage everything together: `git add <code paths>` AND `git add .beads/issues.jsonl` (also `.beads/interactions.jsonl` if modified).
8. Single commit (HEREDOC, no `--no-verify`, no `--amend`) with message exactly: `{ID} P{PRIORITY} {TITLE}`. Emit `{"phase":"committed","sha":"<short>"}`.
9. If push mode is `push_per_bead`: `git push` on the current branch (never main). Emit `{"phase":"pushed"}`. If `defer_push`: skip the push; orchestrator handles it.
10. Return one line as your final message:
    - Success → `{ID}: closed <short-sha>`
    - Blocked → `{ID}: blocked - <short reason>`
    - Skipped → `{ID}: skipped - <short reason>`

## Blocked path

If gates fail twice, the change is intractable, or the bead cannot be finished without a prerequisite you filed:
- Emit `{"phase":"blocked","reason":"<short>"}`.
- `bd update {ID} --status blocked --append-notes "<short reason>"` — if a filed prerequisite is the blocker, name its id in the reason.
- Stash any uncommitted changes: `git stash push -m "bead-bonanza blocked: {ID}"`
- Return the blocked summary line. Do not push.

## Skipped path

If the bead is already addressed, malformed, or duplicates a closed bead:

1. **Confirm your working tree is clean — this is a hard precondition, not a courtesy.** Before doing anything else, run the diff-stat self-check:

   ```bash
   git diff --stat            # unstaged changes
   git diff --staged --stat   # staged changes
   ```

   Both must produce empty output. If either shows changes:
   - You did work the bead should record. **This is not a skip.** Go back to step 4 (`gates-running`) of the per-bead workflow and finish the close ceremony as `completed`.
   - OR the changes are accidental / pre-existing / out-of-scope. Revert them: `git restore --staged <files>` then `git checkout -- <files>` (or `git stash push -m "out-of-scope before skip: {ID}"` if you want to preserve them for later inspection). Re-run the diff-stat check; it must be empty before you proceed.

   Do not skip this check. The orchestrator runs the same diff-stat verification on return (Phase 3 post-return step 3 above); a non-empty diff with a `skipped` claim will block the close and force a re-dispatch. The Phase 3 check is the safety net — your self-check here is the primary line of defense.

2. `bd update {ID} --status closed --append-notes "duplicate of <id>"` (or similar one-line rationale).

3. Return the skipped summary line **with a citation**: `{ID}: skipped - already addressed by <bead-id> @ <short-sha>` or `{ID}: skipped - duplicates <bead-id> (closed YYYY-MM-DD)`. "Already done" without a citation is not acceptable; the orchestrator can't verify it. To find the citation, use `git log --oneline --all | grep <keyword>` or `bd show <suspected-duplicate-id>`.

## Constraints

- Do not ask the user questions. Decisions are baked into this prompt.
- Do not modify code outside this bead's scope.
- Never reset or reclaim a bead's status that you did not set yourself. A bead already `in_progress` when you arrive belongs to another agent — treat it as untouchable, exactly like foreign working-tree changes; leave claim reconciliation to the orchestrator.
- Do not skip git hooks (`--no-verify`). If you realize a required step was missed (e.g. the tracker export) after committing, add a NEW follow-up commit — never amend a commit that may already be pushed, and never force-push a published branch. (Amending a local commit you are certain is unpushed, to fold in a missed step, is fine.)
- Only push the current branch — never push to main.
- Do not use `git reset --hard`, `git checkout .`, or other destructive recovery commands without first stashing.
```

## Doc/chore fast-lane subagent prompt template

```
You are processing {N} trivial beads in this repo. Read CLAUDE.md and AGENTS.md if you need conventions; gate commands are in AGENTS.md under "## bead-bonanza gates".

Beads (process in order; one commit per bead):
{ID_1} (P{PRIORITY_1}) — {TITLE_1} — Decision: {DECISION_1}
{ID_2} (P{PRIORITY_2}) — {TITLE_2} — Decision: {DECISION_2}
...

Gates: format=`{FORMAT}` lint=`{LINT}` build=`{BUILD}` test=`{TEST}` custom=`{CUSTOM_GATES or "none"}`
Push mode: {push_per_bead | defer_push}
Run id: {RUN_ID}

These are documentation, comments, single-line swaps, or other low-risk changes. For each bead in order:

1. Run the per-bead workflow (claim → read → implement → write retro file → atomic close → stage → single commit → optional push) exactly as in the sequential template, INCLUDING per-bead progress JSONL writes AND the per-bead retro file write to `${TMPDIR:-/tmp}/claude/bead-bonanza/{RUN_ID}/<bead-id>.retro.md` BEFORE close.
2. Between beads: skip the full `{TEST}` run. Run `{FORMAT}` + `{BUILD}` per bead. **Targeted tests are non-negotiable** — identify which test file covers the touched surface (see the profile's targeted-tests-by-surface list) and run it. If the change is so trivial there's no plausible test (e.g. pure docs edit, a comment with zero output diff), say so explicitly in the bead's retro and continue. "I couldn't find a targeted test" without explanation is not acceptable.
3. After the LAST bead in this batch, run the full gate suite once (`{FORMAT}`, `{LINT}`, `{BUILD}`, `{TEST}`, plus any custom gates). **The push is the gate** — full gates must pass before the orchestrator pushes. If full gates fail, the most-recent bead is the suspect — investigate and fix it (a new commit), then continue.

Return ONE message containing N+1 lines: one per-bead summary, then a final `batch-end: full gates {passed|failed}`.

## Constraints

Same as the per-bead constraints. Each bead is its own commit. In `defer_push` mode, do not push at all — the orchestrator pushes after the batch.
```

## Shared-surface batch subagent prompt template

```
You are processing {N} related beads that share a code/fixture surface ({SURFACE_DESCRIPTION}, e.g. "the JSON serializer in src/serialize.* and the serializer golden fixtures").

Read CLAUDE.md and AGENTS.md if you need conventions; gate commands are in AGENTS.md under "## bead-bonanza gates".

Beads (process in order; each is its own commit):
{ID_1} (P{PRIORITY_1}) — {TITLE_1} — Decision: {DECISION_1} — Implicit: {IMPLICIT_1}
{ID_2} (P{PRIORITY_2}) — {TITLE_2} — Decision: {DECISION_2} — Implicit: {IMPLICIT_2}
...

Gates: format=`{FORMAT}` lint=`{LINT}` build=`{BUILD}` test=`{TEST}` regen=`{REGEN or "none"}` custom=`{CUSTOM_GATES or "none"}`
Targeted tests for this surface: {SURFACE_TARGETED_TESTS}
Push mode: {push_per_bead | defer_push}
Run id: {RUN_ID}

These beads are NOT trivial — they touch real code. The reason they're batched is to avoid running the full `{TEST}` suite between them when later beads will rebuild the same surface. **The push at the end is the gate; per-bead validation in this batch is a regression-detector to prevent rework, not the final gate.**

For each bead in order:

1. Run the per-bead workflow (claim → read → implement → write retro file → atomic close → stage → single commit → optional push), INCLUDING progress JSONL writes per bead AND the per-bead retro file at `${TMPDIR:-/tmp}/claude/bead-bonanza/{RUN_ID}/<bead-id>.retro.md` BEFORE close.
2. Between beads, run gates AT THIS REDUCED LEVEL — but **targeted tests are non-negotiable**:
   - `{FORMAT}`
   - `{LINT}`
   - `{BUILD}`
   - If generated output was touched AND a regen step exists: run `{REGEN}` then the relevant fixture/snapshot test.
   - **Run the surface-relevant targeted tests** named above for the file(s) you touched. If none is listed for a surface you touched, consult the profile's targeted-tests-by-surface list; if there's still nothing, that's a signal the change is broader than expected — pause and reconsider scope, don't take it as permission to skip.

   **Why this matters:** without per-bead targeted tests, a regression introduced in bead #2 only surfaces at batch end, after beads #3 and #4 may have built on the broken assumption. You then bisect, rework #2, and possibly redo #3 and #4. Targeted tests prevent that cascade.
3. **After the LAST bead in this batch** (push gate). If the repo has a pre-push hook running a CI-mirror gauntlet AND a PR is open so pushes trigger full CI (ask the user once at run start — "slim gates"), SKIP this step entirely: the hook and CI *are* the full gate, and running the suite in-batch just duplicates them minutes apart. Otherwise run the full gate suite — must pass before orchestrator pushes:
   - `{FORMAT}`
   - `{LINT}`
   - `{BUILD}`
   - `{TEST}`
   - Any custom gates from the profile
   If it fails, identify which bead introduced the regression (bisect via the per-bead commits — `git log --oneline | head` then `git checkout <sha>` and rerun the failing test) and follow the Blocked path for THAT bead — do not silently fix it across multiple commits.

Return ONE message containing N+1 lines: one per-bead summary, then a final `batch-end: full gates {passed|failed}`.

## Constraints

- Each bead is its own commit. Never combine bead changes into one commit.
- If a bead's implementation conflicts with an earlier bead in this batch (e.g. one removed a function the next was about to modify), follow the Blocked path for the conflicting bead and continue with the rest.
- Same per-bead constraints as the sequential template.
```

## Worktree parallel notes (orchestrator)

When dispatching a wave of `isolated` beads in parallel:

- Spawn all subagents in a single message with multiple `Agent` tool calls, each with `isolation: "worktree"`.
- Each subagent uses the per-bead template above, runs in its own worktree on its own branch.
- After all subagents return, merge their branches into the parent branch one at a time:
  - `git merge <branch>` — if clean, continue.
  - If a fixture/snapshot conflict and the profile defines a regen step: take both sides, run `<regen>` on the merged tree, stage and amend the merge commit. Without a regen step, resolve the conflict manually.
  - If non-fixture conflict on shared code: triage failure — the beads weren't actually isolated. Mark the unmerged bead as `blocked` with a note about the conflict; let the user resolve.
- If the bucket used `defer-push`, push once after all merges complete. Otherwise individual subagents pushed their branches; the orchestrator's merge commits still need a `git push` per merge or one `git push` at the end (your call — defer is fine here).


# bead-bonanza Subagent Prompt Templates

The prompts the orchestrator fills in and dispatches. Read this at Phase 3
(dispatch), and at Phase 1.5 when a bead needs grooming first. Substitute
every `{PLACEHOLDER}` before sending — an unsubstituted placeholder is the
single most common dispatch defect, because the subagent has no session
context to infer it from.

Two placeholders are resolved once per run and substituted into every
prompt: `{RUN_DIR}` is the run root Phase 0's pre-flight script resolved, and `{SCRIPTS}` is
the absolute path to this skill's `scripts/` directory. A subagent runs from
the repo root and cannot infer either.

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
Run dir: {RUN_DIR}          (progress + retro files live here)
Scripts: {SCRIPTS}          (bb-progress.sh, bb-verify-skip.sh, bb-close-bead.sh)

## Workflow — follow in order, do not deviate

Emit a progress event at each labeled step by running the bundled script:
`{SCRIPTS}/bb-progress.sh "{RUN_DIR}" {ID} <phase> [key=value ...]`. It builds
and appends the JSON itself — never hand-write an event.

1. `bb-progress.sh "$RUN_DIR" {ID} claimed`. Claim: `bd update {ID} --status in_progress`
2. Read full details: `bd show {ID}`
3. `bb-progress.sh "$RUN_DIR" {ID} implementing`. Implement the change. **If this bead fixes a defect or regression, first write (or locate) a test that reproduces it and observe it FAIL for the stated reason BEFORE you fix it** — a test written after the fix encodes the fix's own assumptions and can stay green while the defect is live. Record the observed failure; it goes in the retro. Stay strictly within this bead's scope. If you find work beyond it, file a new bead via `bd create` with a `discovered-from` link to {ID}, noting in one line whether it *blocks* this bead (a prerequisite, an acceptance-criteria decomposition, or an expected test-coverage gap) or is merely adjacent — the orchestrator decides whether it enters the run. Do NOT implement it or pull it into your own scope. If this bead genuinely cannot be finished without it, take the Blocked path naming the new bead as the blocker; otherwise continue with the original — do NOT expand scope inline. For any implicit decision you must make, record what you chose (and why) — you'll put it in the retro.
4. `bb-progress.sh "$RUN_DIR" {ID} gates-running`. **The push is the gate** — this is a single-bead push, so full gates run here before pushing. If the change affects generated output the tests compare against (codegen, snapshots, golden files, schemas) AND the gate profile defines a regen step, run `{REGEN}` BEFORE the test step.
   - `{FORMAT}`
   - `{LINT}`
   - `{BUILD}`
   - `{TEST}`
   - Any custom gates from the profile: `{CUSTOM_GATES}` (skip if none)
   On gate failure, fix the root cause and rerun. After two full failed cycles, follow the Blocked path — do not keep retrying. Run `bb-progress.sh "$RUN_DIR" {ID} gates-passed` after success.
5. **Write the Retro to a file BEFORE closing the bead** (crash recovery anchor — non-negotiable).
   - Start the file mechanically, never by retyping: `bd show {ID} --json | jq -r '.[0].description' > "{RUN_DIR}/{ID}.retro.md"`. Retyping the description costs generated tokens and silently drops content — a groomed bead's own `## Retro` instructions are what usually go missing, and `{SCRIPTS}/bb-close-bead.sh` then writes the lossy copy back over the bead with `--body-file`.
   - Then append a `## Retro — implementation` section (that exact heading, so it cannot collide with or displace any `## Retro` the bead already carries). Four items, in order, and keep it SHORT — write for a reader who already has the diff, the commit, `bd show {ID}` and the gate output. Do NOT restate the approach, the diffstat, the files changed, the gate table or the acceptance criteria; `git show --stat <sha>` already holds all of it.
     1. **Findings** — one bullet per point of friction, and record EACH ALSO as its own `bd comment` on {ID} prefixed `retro:`. That comment channel is what downstream retros and RFCs actually read; the file alone does not reach them. Trigger list — record one WHEN: you needed information the bead and its references did not carry; you had to settle something the bead or spec should have pinned; the bead contradicted repo reality; you could not satisfy an acceptance criterion as declared; a sequencing surprise. 1–3 sentences each, friction only, no root-causing. **A clean bead records none — say so rather than inventing filler.**
     2. **What you tried that did not work** — the reading you took first and abandoned, and why. The diff shows only the reading that survived, so this is unrecoverable later.
     3. **What you decided silently** — only choices a diff-reader would not recognise AS choices. Narrower than "every implicit decision": an internal helper's name or the order of two independent guard clauses qualifies; the approach the bead already specified does not.
     4. **For a defect bead: the observed pre-fix failure**, quoted from actual output. If you could not reproduce it, say so and say what you substituted.
   - Run `bb-progress.sh "$RUN_DIR" {ID} retro-written` after the file is on disk.
6. Close and commit in one step — run the bundled script:
   `{SCRIPTS}/bb-close-bead.sh {ID} {PRIORITY} "{TITLE}" "{RUN_DIR}/{ID}.retro.md" <code paths>`
   It performs the whole ceremony in the required order: atomic close (`bd update --status closed`, never `bd close` — that would be a second commit), refresh the `.beads/issues.jsonl` export and verify it changed, stage code + tracker together, then one commit messaged `{ID} P{PRIORITY} {TITLE}`. It prints the short sha and refuses to run on `main`, without a retro file, or with nothing staged. Do not perform these steps by hand.
7. `{SCRIPTS}/bb-progress.sh "{RUN_DIR}" {ID} committed sha=<the sha it printed>`.
8. If push mode is `push_per_bead`: `git push origin <branch>` on the current branch (never main), then `{SCRIPTS}/bb-progress.sh "{RUN_DIR}" {ID} pushed`. If `defer_push`: skip the push; the orchestrator handles it.
9. Return one line as your final message:
    - Success → `{ID}: closed <short-sha>`
    - Blocked → `{ID}: blocked - <short reason>`
    - Skipped → `{ID}: skipped - <short reason>`

## Blocked path

If gates fail twice, the change is intractable, or the bead cannot be finished without a prerequisite you filed:
- `bb-progress.sh "$RUN_DIR" {ID} blocked reason="<short>"`.
- `bd update {ID} --status blocked --append-notes "<short reason>"` — if a filed prerequisite is the blocker, name its id in the reason.
- Stash any uncommitted changes: `git stash push -m "bead-bonanza blocked: {ID}"`
- Return the blocked summary line. Do not push.

## Skipped path

If the bead is already addressed, malformed, or duplicates a closed bead:

1. **Confirm your working tree is clean — this is a hard precondition, not a courtesy.** Run `{SCRIPTS}/bb-verify-skip.sh` before anything else. It must exit 0.

   If it exits 1 it prints what is dirty and the two ways out: finish the close ceremony as `completed` (the work belongs to this bead), or revert/stash the changes and re-run until CLEAN. Do not proceed while it exits 1.

   The orchestrator runs the same script on return (Phase 3 post-return step 3), so a contradictory claim will be caught and re-dispatched either way — but its check is the safety net, and yours is the primary line of defense.

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
Run dir: {RUN_DIR}
Scripts: {SCRIPTS}

These are documentation, comments, single-line swaps, or other low-risk changes. For each bead in order:

1. Run the per-bead workflow (claim → read → implement → write retro file → atomic close → stage → single commit → optional push) exactly as in the sequential template, INCLUDING per-bead progress JSONL writes AND the per-bead retro file write to `{RUN_DIR}/<bead-id>.retro.md` BEFORE close.
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
Run dir: {RUN_DIR}
Scripts: {SCRIPTS}

These beads are NOT trivial — they touch real code. The reason they're batched is to avoid running the full `{TEST}` suite between them when later beads will rebuild the same surface. **The push at the end is the gate; per-bead validation in this batch is a regression-detector to prevent rework, not the final gate.**

For each bead in order:

1. Run the per-bead workflow (claim → read → implement → write retro file → atomic close → stage → single commit → optional push), INCLUDING progress JSONL writes per bead AND the per-bead retro file at `{RUN_DIR}/<bead-id>.retro.md` BEFORE close.
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


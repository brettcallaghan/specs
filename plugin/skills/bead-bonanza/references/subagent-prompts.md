# bead-bonanza Subagent Prompt Templates

The prompts the orchestrator fills in and dispatches. Read this at Phase 3
(dispatch), and at Phase 1.5 when a bead needs grooming first. Substitute
every `{PLACEHOLDER}` before sending — an unsubstituted placeholder is the
single most common dispatch defect, because the subagent has no session
context to infer it from.

Two placeholders are resolved once per run and substituted into every
prompt: `{RUN_DIR}` is the run root Phase 0's pre-flight script resolved, and `{SCRIPTS}` is
the absolute path to this skill's `scripts/` directory. A subagent cannot
infer either.

`{WORKTREE}` is resolved **per bead** — the absolute path to the worktree the
orchestrator created for it. A subagent works there and nowhere else; it
cannot infer the path, and a prompt that omits it sends the agent into the
shared checkout, which defeats every guarantee this design rests on.

## Contents

- Grooming subagent (Phase 1.5) — bring under-specified beads to Ready
- Per-bead subagent — the canonical ceremony, run in the bead's own worktree
- Doc/chore fast-lane subagent — a cluster of trivial beads, one subagent
- Shared-surface batch subagent — beads sharing a code/fixture surface
- Worktree mechanics (orchestrator) — the base-ref trap, how `bd` behaves
  inside a worktree, the create/merge/remove lifecycle, and concurrent
  dispatch within a wave

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

For each bead: bring it to Ready by editing it in place with `bd update` — enrich thin descriptions rather than discarding author intent; wire only dependencies you can justify; assign a `tier` with a `tier_reason`; and where the repo has C4 docs, tag `c4_changes` / `c4_impacts` as arrays in metadata (the orchestrator buckets waves off those two fields).

Do NOT implement any bead. Do NOT invent IDs, component names, or dependencies. Bounce, don't guess: any bead you cannot bring to Ready, leave as-is and report NOT-READY with the specific gap a human must close.

Return a report, one line per bead:
- `{ID}: ready — <notable change: description made context-independent / split into epic <id> + children <ids> / model assigned>`
- `{ID}: NOT-READY - <the gap that needs a human>`

Then return the **wave preview** `groom-beads` produces, as its own block after the per-bead lines — the orchestrator's Phase 2.5 consumes it as a starting partition, so a report without this block loses it:
- `wave N: {ID}, {ID}, … — <why they are safe together>`, one line per candidate wave, grouped under both constraints (dependencies, and no two beads sharing a `c4_changes` component in a wave)
- `sequencing: {ID} after {ID} — consumes <component>` for each pair whose `c4_impacts` intersects another bead's `c4_changes`
- `wave-preview: none (no C4 docs)` instead of the above WHEN the repo documents no architecture and you skipped C4 tagging — the orchestrator must know the signal is absent, not empty
- `c4-untagged: {ID} — <why>` for any bead you could not tag (component with no doc, scope unclear)
```


---

## Per-bead subagent prompt template

```
You are addressing a single bead in this repo. Read CLAUDE.md and AGENTS.md if you need repo conventions or the bead workflow. The gate commands are in AGENTS.md under "## bead-bonanza gates".

Bead: {ID} (P{PRIORITY}) — {TITLE}
Decision (chosen by the user up front): {DECISION}
Implicit decisions (subagent picks unless overridden): {IMPLICIT_NOTES}
Gates: format=`{FORMAT}` lint=`{LINT}` build=`{BUILD}` test=`{TEST}` regen=`{REGEN or "none"}` custom=`{CUSTOM_GATES or "none"}`
Push mode: {push_per_bead | defer_push}
Worktree: {WORKTREE}        (work HERE, not in the main checkout; it is yours alone)
Run dir: {RUN_DIR}          (progress + handoff files live here)
Scripts: {SCRIPTS}          (bb-progress.sh, bb-verify-skip.sh, bb-close-bead.sh)
Trailers file: {TRAILERS_FILE or "none"}   (Phase 0 resolved; pass to bb-close-bead.sh --trailers-file)

## Workflow — follow in order, do not deviate

Emit a progress event at each labeled step by running the bundled script:
`{SCRIPTS}/bb-progress.sh "{RUN_DIR}" {ID} <phase> [key=value ...]`. It builds
and appends the JSON itself — never hand-write an event.

1. `bb-progress.sh "$RUN_DIR" {ID} claimed`. Claim: `bd update {ID} --status in_progress`
2. Read full details: `bd show {ID}`
3. `bb-progress.sh "$RUN_DIR" {ID} implementing`. Implement the change. **If this bead fixes a defect or regression, first write (or locate) a test that reproduces it and observe it FAIL for the stated reason BEFORE you fix it** — a test written after the fix encodes the fix's own assumptions and can stay green while the defect is live. Record the observed failure; it goes in a `retro:` comment (step 6). Stay strictly within this bead's scope. If you find work beyond it, file a new bead via `bd create` with a `discovered-from` link to {ID}, noting in one line whether it *blocks* this bead (a prerequisite, an acceptance-criteria decomposition, or an expected test-coverage gap) or is merely adjacent — the orchestrator decides whether it enters the run. Do NOT implement it or pull it into your own scope. If this bead genuinely cannot be finished without it, take the Blocked path naming the new bead as the blocker; otherwise continue with the original — do NOT expand scope inline. For any implicit decision you must make, record what you chose (and why) — it goes in a `retro:` comment (step 6).
4. `bb-progress.sh "$RUN_DIR" {ID} gates-running`. **The push is the gate** — this is a single-bead push, so full gates run here before pushing. If the change affects generated output the tests compare against (codegen, snapshots, golden files, schemas) AND the gate profile defines a regen step, run `{REGEN}` BEFORE the test step.
   - `{FORMAT}`
   - `{LINT}`
   - `{BUILD}`
   - `{TEST}`
   - Any custom gates from the profile: `{CUSTOM_GATES}` (skip if none)
   On gate failure, fix the root cause and rerun. After two full failed cycles, follow the Blocked path — do not keep retrying. Run `bb-progress.sh "$RUN_DIR" {ID} gates-passed` after success.
5. **Write the handoff file BEFORE closing the bead** (crash recovery anchor — non-negotiable). Write `{RUN_DIR}/{ID}.handoff.md` as a `# {ID} handoff` heading plus exactly one labeled line, and nothing else:
   - `Caveat:` — anything the close must do differently: a regen step that must run before staging, a fixture this bead regenerated. Write `none` when there is none.

   **The `Paths:` line is retired** (`agent-plugins-92q.30`). It existed because, in a tree another agent might be sharing, nobody could re-derive from `git status` which dirt belonged to this bead. You are in this bead's own worktree, so everything dirty here is yours by construction and the close stages all of it. Do not reintroduce a path list.

   Three lines total is typical. Do NOT restate the phase (the progress JSONL has it), the diff, the gates, or the bead, and do NOT put findings in it — those are step 6. Run `bb-progress.sh "$RUN_DIR" {ID} retro-written` after the file is on disk (the phase vocabulary keeps that name).
6. **Record findings as `retro:` comments on {ID}** — one `bd comment` per finding, prefixed `retro:`, 1–3 sentences each, friction only, no root-causing. That comment channel is what downstream retros and RFCs actually read. Trigger list — record one WHEN: you needed information the bead and its references did not carry; you had to settle something the bead or spec should have pinned; the bead contradicted repo reality; you could not satisfy an acceptance criterion as declared; a sequencing surprise. **A clean bead records none — say so rather than inventing filler.** Record each of these as its own `retro:` comment too:
     - **What you tried that did not work** — the reading you took first and abandoned, and why. The diff shows only the reading that survived, so this is unrecoverable later.
     - **What you decided silently** — only choices a diff-reader would not recognise AS choices. Narrower than "every implicit decision": an internal helper's name or the order of two independent guard clauses qualifies; the approach the bead already specified does not.
     - **For a defect bead: the observed pre-fix failure**, quoted from actual output. If you could not reproduce it, say so and say what you substituted.
7. Close and commit in one step — run the bundled script. WHEN `{TRAILERS_FILE}` is not "none", pass it first so the trailers land in the same commit and no amend is ever needed:
   `{SCRIPTS}/bb-close-bead.sh --trailers-file {TRAILERS_FILE} {ID} {PRIORITY} "{TITLE}"`
   Otherwise (`{TRAILERS_FILE}` is "none"): `{SCRIPTS}/bb-close-bead.sh {ID} {PRIORITY} "{TITLE}"`
   **Pass no paths.** You are in this bead's own worktree, so everything modified here is this bead's work and the script stages all of it. A path list can only be too narrow (`agent-plugins-92q.30`).
   It performs the whole ceremony in the required order: atomic close (`bd update --status closed`, never `bd close` — that would be a second commit), refresh the `.beads/issues.jsonl` export and verify it changed, stage code + tracker together, then one commit messaged `{ID} P{PRIORITY} {TITLE}` (plus the trailers file's contents in the body, when given). It does not touch the description — grooming owns that. It refuses to run on `main` or with nothing staged. Do not perform these steps by hand, and do not `git commit --amend` afterward — if a required trailer is missing, that means `{TRAILERS_FILE}` was resolved wrong at Phase 0, not something to patch here.
8. `{SCRIPTS}/bb-progress.sh "{RUN_DIR}" {ID} committed sha=$(git rev-parse --short HEAD)` — read the sha from the branch, not from the script's printout, so any rewrite between steps 7 and 8 cannot desync the ledger from what actually landed.
9. If push mode is `push_per_bead`: `git push origin <branch>` on the current branch (never main), then `{SCRIPTS}/bb-progress.sh "{RUN_DIR}" {ID} pushed`. If `defer_push`: skip the push; the orchestrator handles it.
10. Return one line as your final message:
    - Success → `{ID}: closed <short-sha>`
    - Blocked → `{ID}: blocked - <short reason>`
    - Skipped → `{ID}: skipped - <short reason>`

## Blocked path

If gates fail twice, the change is intractable, or the bead cannot be finished without a prerequisite you filed:
- `bb-progress.sh "$RUN_DIR" {ID} blocked reason="<short>"`.
- `bd update {ID} --status blocked --append-notes "<short reason>"` — if a filed prerequisite is the blocker, name its id in the reason.
- **Leave your uncommitted changes exactly where they are, in your worktree.** Do NOT stash and do NOT commit them. The worktree holds the partial work and its continued existence is what tells the orchestrator this bead is unfinished; stashing would move the work out of it and destroy that signal (`agent-plugins-92q.30`).
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
Worktrees: one per bead, created by YOU, held ONE AT A TIME. Before each bead:
  git -C <repo> worktree add -b <suffix> <repo>/../.bb-worktrees/<suffix> HEAD
Work that bead entirely inside its worktree. After its close commit lands, tell
the orchestrator and move on — never hold two worktrees open at once, because a
worktree that exists means "this bead is unfinished" to every other reader.
Run dir: {RUN_DIR}
Scripts: {SCRIPTS}
Trailers file: {TRAILERS_FILE or "none"}   (Phase 0 resolved; pass to bb-close-bead.sh --trailers-file)

These are documentation, comments, single-line swaps, or other low-risk changes. For each bead in order:

1. Run the per-bead workflow (claim → read → implement → write handoff file → `retro:` comments → atomic close → stage → single commit → optional push) exactly as in the sequential template — including its close step's `--trailers-file {TRAILERS_FILE}` and reading the committed sha with `git rev-parse --short HEAD` rather than the script's printout — INCLUDING per-bead progress JSONL writes AND the per-bead handoff file write to `{RUN_DIR}/<bead-id>.handoff.md` BEFORE close.
2. Between beads: skip the full `{TEST}` run. Run `{FORMAT}` + `{BUILD}` per bead. **Targeted tests are non-negotiable** — identify which test file covers the touched surface (see the profile's targeted-tests-by-surface list) and run it. If the change is so trivial there's no plausible test (e.g. pure docs edit, a comment with zero output diff), say so explicitly in a `retro:` comment on the bead and continue. "I couldn't find a targeted test" without explanation is not acceptable.
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
Worktrees: one per bead, created by YOU, held ONE AT A TIME. Before each bead:
  git -C <repo> worktree add -b <suffix> <repo>/../.bb-worktrees/<suffix> HEAD
Work that bead entirely inside its worktree. After its close commit lands, tell
the orchestrator and move on — never hold two worktrees open at once, because a
worktree that exists means "this bead is unfinished" to every other reader.
Run dir: {RUN_DIR}
Scripts: {SCRIPTS}
Trailers file: {TRAILERS_FILE or "none"}   (Phase 0 resolved; pass to bb-close-bead.sh --trailers-file)

These beads are NOT trivial — they touch real code. The reason they're batched is to avoid running the full `{TEST}` suite between them when later beads will rebuild the same surface. **The push at the end is the gate; per-bead validation in this batch is a regression-detector to prevent rework, not the final gate.**

For each bead in order:

1. Run the per-bead workflow (claim → read → implement → write handoff file → `retro:` comments → atomic close → stage → single commit → optional push) — including its close step's `--trailers-file {TRAILERS_FILE}` and reading the committed sha with `git rev-parse --short HEAD` rather than the script's printout — INCLUDING progress JSONL writes per bead AND the per-bead handoff file at `{RUN_DIR}/<bead-id>.handoff.md` BEFORE close.
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

## Worktree mechanics (orchestrator)

### The base-ref trap — read before the first wave

**Every worktree MUST branch from the sprint's own branch, never the
default branch.** Two conditions produce that, and both are checked: the
orchestrator's own HEAD is not the default branch (pre-flight blocks the
run otherwise), and `worktree.baseRef` must be `"head"` so the worktree
takes session HEAD rather than the repository's default branch. Unset,
the worktree branches from the repository's **default branch**, not
session HEAD — measured in this repo, `agent-plugins-jh6`. This is not
conditional on whether the run's beads build on each other: an unset
`baseRef` silently drops every commit the sprint branch carries from
every worktree, isolated beads included — it is a hard gate for
`isolation: "worktree"` full stop, not a special case for beads with a
dependency edge. Fix (user scope only; project-root untested):

```json
{ "worktree": { "baseRef": "head" } }
```

Takes effect immediately, no restart.

**Verify before every wave** — this is global state the repo doesn't own and
can change without warning. Note session HEAD (`git rev-parse --short HEAD`)
and give the **first** worktree subagent this instruction:

```
Before doing anything else, run in your worktree:
  git rev-parse --short HEAD
  git merge-base --is-ancestor <SESSION_HEAD_SHA> HEAD; echo "ancestor-exit=$?"
Report both lines verbatim as the first line of your reply.
```

WHEN `ancestor-exit` is non-zero, **abort the wave**: wrong base, do not
merge anything back. Fix `worktree.baseRef`, or fall back to sequential mode.

### Beads inside a worktree

`bd` is not an obstacle: it resolves the shared tracker database through
git's common directory from any worktree — `bd where`/`bd worktree info`
point at the main checkout, `bd list` returns the full set, no `embeddeddolt`
or redirect file created (measured, `BEADS_DIR` unset, `agent-plugins-jh6`,
`agent-plugins-t09`). All worktree agents share one tracker, which is
exactly what makes these four constraints real:

- **A worktree's own `.beads/issues.jsonl` goes stale while `git status`
  reads clean** — auto-export always writes the MAIN checkout's copy
  (measured 105 lines vs. 128, `agent-plugins-jh6`). **A worktree agent
  SHALL run `bd export > .beads/issues.jsonl` in its own directory before
  committing.**
- **Auto-export is best-effort, not a guarantee** — three creates sat
  unexported 20+ seconds, flushing only on a later main-context write
  (`agent-plugins-t09`). Never rely on it inside a worktree.
- **`export.git-add: true` (`.beads/config.yaml`) stages the refreshed
  ledger into the MAIN index from writes issued anywhere, including a
  worktree** — the case the orchestrator's `git commit -- <pathspec>` guard
  exists for (SKILL.md, "Guard against foreign dirt"); keep commits
  pathspec-scoped.
- **The exported ledger is gone.** Until `agent-plugins-92q.31` the mandated
  per-close export made concurrency unsafe: N beads each committed a
  near-identical full rewrite of one shared database onto N branches with no
  merge driver, so an N-way `.beads/issues.jsonl` conflict at merge time was
  guaranteed. The tracker is now `origin refs/dolt/data`, `.beads/*.jsonl` is
  untracked, and the close ceremony stages code only and pushes the database
  itself. There is no per-worktree export to conflict, and no export step for a
  worktree agent to remember.

### The worktree lifecycle — create, dispatch, merge, remove

The orchestrator owns creation and removal; the subagent owns only what
happens inside. Per bead, in order:

```sh
REPO="$(git rev-parse --show-toplevel)"
WT="$REPO/../.bb-worktrees/<suffix>"        # sibling of the repo, never inside it
git -C "$REPO" worktree add -b <suffix> "$WT" HEAD
```

Branching from `HEAD` — the sprint branch as it stands *now*, after the
previous bead merged — is what makes each bead build on the last. Where the
harness creates the worktree for you (`Agent` with `isolation: "worktree"`),
it uses `worktree.baseRef` instead, which is why that setting is a hard gate.

1. **Dispatch** the subagent with the per-bead template, telling it to work in
   `$WT`. It implements, gates, writes its handoff, records findings, and
   closes with `../scripts/bb-close-bead.sh` — passing **no paths**, since everything in
   that tree is this bead's work.
2. **Accept**: verify the bead reads `closed` and its commit exists on the
   bead's branch. This is the acceptance check; the worktree still existing at
   this point is what tells you the bead is not yet done.
3. **Merge** into the sprint branch, one bead at a time:
   ```sh
   git -C "$REPO" merge --ff-only <suffix>
   ```
   **`--ff-only` is deliberate — in a SEQUENTIAL wave.** Each bead branches
   from an already-updated head, so every merge is a fast-forward and a refusal
   means an invariant broke: a worktree branched from the wrong base. **Stop
   the wave and report it; do not resolve it inline** and do not retry without
   `--ff-only`.

   **In a CONCURRENT wave, only the first merge can fast-forward**, and that is
   not a fault — every branch in the wave shares one base by construction, so
   the second onward necessarily diverge. Measured 2026-08-12 on the first real
   concurrent wave (`agent-plugins-bhl`): `--ff-only` fast-forwarded bead one
   and then refused bead two with "Not possible to fast-forward" on a wave in
   which nothing was wrong. Applying the sequential rule literally would stop
   every concurrent wave of 2+ beads and report it as broken.

   So for each branch after the first: **rebase it onto the updated sprint head,
   then fast-forward.**

   ```sh
   git -C <worktree> rebase <sprint-head>   # replay this bead onto what landed
   git merge --ff-only <suffix>             # now it fast-forwards
   ```

   Rebase keeps one-commit-per-bead and a linear history, so it is the
   concurrent analogue of the sequential guarantee. **The collision signal moves
   to the rebase:** a rebase CONFLICT means two beads genuinely changed the same
   lines, which is the failure `--ff-only` was standing in for. Stop the wave on
   a rebase conflict, exactly as you would on a sequential `--ff-only` refusal.
4. **Remove the worktree — the final act of accepting the bead:**
   ```sh
   git -C "$REPO" worktree remove "$WT" && git -C "$REPO" branch -d <suffix>
   ```
   Removal is the orchestrator's, not the subagent's: an agent cannot remove
   the tree it is standing in. Because removal happens *after* the merge, a
   surviving worktree means "not yet accepted", which is the stronger reading
   of the in-flight signal than "not yet committed".
5. **Push** per the bucket's push mode.

WHEN a bead is blocked rather than finished, **leave its worktree in place**
and say so in the report. It holds the partial work, and its existence is the
durable record that the bead is unfinished — there is nothing to stash and no
WIP commit to make (`operating-principles.md`).

**Prune before trusting the list.** A reclaimed container deletes worktree
directories without deregistering them, and a stale entry reads as in-flight
forever: `git worktree prune` before `git worktree list` in pre-flight and in
recovery.

### Concurrent dispatch within a wave (enabled)

Spawn several beads at once — one `Agent` call per bead with
`isolation: "worktree"`, all in a single message. The committed ledger that
blocked this is gone (see the ledger bullet above). Waves themselves stay
sequential; concurrency is across beads *inside* a wave.

The lifecycle above is unchanged per bead. Two additions when running
concurrently: merge each branch as it returns rather than after all return, and
where a merge hits a fixture/snapshot conflict and the profile defines a regen
step, take both sides, run `<regen>` on the merged tree, stage and amend. A
non-fixture conflict on shared code remains a triage failure — the beads were
not actually isolated; mark the unmerged bead `blocked` and let the user
resolve.

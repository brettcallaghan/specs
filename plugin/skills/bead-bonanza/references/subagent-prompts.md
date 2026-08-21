# bead-bonanza Subagent Prompt Templates

The prompts the orchestrator fills in and dispatches. Read this at Phase 3
(dispatch), and at Phase 1.5 when a bead needs grooming first. Substitute
every `{PLACEHOLDER}` before sending — an unsubstituted placeholder is the
single most common dispatch defect, because the subagent has no session
context to infer it from.

Four placeholders are resolved once per run and substituted into every
prompt: `{RUN_DIR}` is the run root Phase 0's pre-flight script resolved,
`{SCRIPTS}` is the absolute path to this skill's `scripts/` directory,
`{CHECKS}` is the checks script Phase 0 resolved (emitted as `BB_CHECKS`), and
`{CHECKS_BASELINE_TEXT}` is the *contents* of `{RUN_DIR}/checks-baseline.md`.
A subagent can infer none of them.

`{WORKTREE}` is resolved **per bead** — the absolute path to the worktree the
orchestrator created for it. A subagent works there and nowhere else; it
cannot infer the path, and a prompt that omits it sends the agent into the
shared checkout, which defeats every guarantee this design rests on.

## Contents

- Grooming subagent (Phase 1.5) — bring under-specified beads to Ready, then
  invoke `plan-sprint` to write the run's wave table
- Per-bead subagent — the canonical ceremony, run in the bead's own worktree
- Shared batch preamble — boilerplate common to both batch templates below
- Doc/chore fast-lane subagent — a cluster of trivial beads, one subagent
- Shared-surface batch subagent — beads sharing a code/fixture surface
- Worktree mechanics (orchestrator) — the base-ref trap, how `bd` behaves
  inside a worktree, the create/merge/remove lifecycle, and concurrent
  dispatch within a wave

The per-bead template is the canonical one; the fast-lane and shared-surface
templates reference its workflow rather than restating it, so read it first.
Those two batch templates also open with a reference to the **Shared batch
preamble** below instead of repeating it. **The orchestrator concatenates
that preamble plus the template into the single text it actually sends** —
the subagent reads only the assembled prompt, never this file, so a batch
prompt sent without the preamble is incomplete: the subagent silently loses
the worktree-creation mechanics, the `run-bead` invocation, and the return
contract along with it.

---

## Grooming subagent prompt template

```
You are grooming {N} beads to the Definition of Ready so an orchestrator can dispatch each to an isolated, zero-context sub-agent. Use the `groom-beads` skill and follow its full workflow (Definition of Ready, dependency wiring, tier assignment, C4 blast radius, validate). It loads the spec-owned Definition of Ready itself — do not go looking for that file by path.

Beads to groom:
{ID_1} — {TITLE_1}
{ID_2} — {TITLE_2}
...

For each bead: bring it to Ready by editing it in place with `bd update` — enrich thin descriptions rather than discarding author intent; wire only dependencies you can justify; assign a `tier` with a `tier_reason`; and where the repo has C4 docs, tag `c4_changes` / `c4_impacts` as arrays in metadata (the wave table you write below is bucketed off those two fields).

Do NOT implement any bead. Do NOT invent IDs, component names, or dependencies. Bounce, don't guess: any bead you cannot bring to Ready, leave as-is and report NOT-READY with the specific gap a human must close.

Return a report, one line per bead:
- `{ID}: ready — <notable change: description made context-independent / split into epic <id> + children <ids> / model assigned>`
- `{ID}: NOT-READY - <the gap that needs a human>`

Then produce the **wave preview** `groom-beads` produces, as its own block after the per-bead lines — `plan-sprint` consumes it below as a starting partition, so losing this block costs the whole preview:
- `wave N: {ID}, {ID}, … — <why they are safe together>`, one line per candidate wave, grouped under both constraints (dependencies, and no two beads sharing a `c4_changes` component in a wave)
- `sequencing: {ID} after {ID} — consumes <component>` for each pair whose `c4_impacts` intersects another bead's `c4_changes`
- `wave-preview: none (no C4 docs)` instead of the above WHEN the repo documents no architecture and you skipped C4 tagging — the reader must know the signal is absent, not empty
- `c4-untagged: {ID} — <why>` for any bead you could not tag (component with no doc, scope unclear)

FINALLY, as your last act, invoke the `plan-sprint` skill and follow it to write the run's wave table into `{RUN_DIR}/dispatch.md`. You have just read every bead, which is why planning is yours and not a second agent's. Pass it:
- Run id `{RUN_ID}` and RUN_DIR `{RUN_DIR}` — `dispatch.md`'s manifest header is already written by the orchestrator; append under it and do NOT edit it
- The bead ids to plan, ENUMERATED: {PLAN_IDS} — this is the run's committed queue; never re-run a filter to rebuild it, and never add a bead you groomed that the list omits
- Checks: the project's `{CHECKS}` script, run in full; its baseline is at `{RUN_DIR}/checks-baseline.md`
- Model lineup: {MODEL_LINEUP or "none declared"}
- Standing authorization: {STANDING_AUTH or "none"}
- Your own wave preview, above

Return the wave table it produces as the last block of your report, plus one line per row where you overrode your own preview and why.
```


---

## Per-bead subagent prompt template

```
You are addressing a single bead in this repo. Read CLAUDE.md and AGENTS.md if you need repo conventions or the bead workflow. Validation comes from the project's `checks` script, which Phase 0 resolved.

Bead: {ID} (P{PRIORITY}) — {TITLE}
Decision (chosen by the user up front): {DECISION}
Implicit decisions (subagent picks unless overridden): {IMPLICIT_NOTES}
Checks: {CHECKS}            (run it; exit 0 is the bar, and there is nothing to select)
Push mode: {push_per_bead | defer_push}
Worktree: {WORKTREE}        (work HERE, not in the main checkout; it is yours alone)
Scripts: {SCRIPTS}          (bb-verify-skip.sh, bb-close-bead.sh)
Trailers file: {TRAILERS_FILE or "none"}   (Phase 0 resolved; pass to bb-close-bead.sh --trailers-file)
Checks baseline:
{CHECKS_BASELINE_TEXT}
    (The CONTENTS of {RUN_DIR}/checks-baseline.md, inlined verbatim — NOT the
     one-word status pre-flight emits as BB_CHECKS_BASELINE. Include the block
     even when the checks passed, because the file says so and a MISSING block
     means no baseline was taken, which is a different thing.)

**Then invoke the `run-bead` skill.** It carries the ceremony — claim, read,
implement in scope, gate, caveat, findings, close-and-commit, push, return —
together with the blocked and skipped paths and its own acceptance criteria.
The template above supplies its inputs; it supplies the procedure.
```

## Shared batch preamble

Both batch templates below (doc/chore fast-lane, shared-surface) open with
this block rather than repeating it. **The orchestrator concatenates this
preamble plus the chosen template into the one text it sends** — a batch
prompt assembled without it is missing the worktree-creation mechanics, the
`run-bead` invocation, and the return contract, with nothing to catch the
loss downstream.

```
Read CLAUDE.md and AGENTS.md if you need conventions; validation comes from the project's `checks` script.

Worktrees: **where the orchestrator already provisioned one for you (the harness's
`isolation: "worktree"`), you have exactly ONE worktree for EVERY bead in this batch —
process them in order, one commit each, and do NOT create more.** Where it did not,
create one per bead yourself, held ONE AT A TIME. Before each bead:
  First bead: git -C <repo> worktree add -b <id-1> <repo>/../.bb-worktrees/<id-1> HEAD
  Each later bead: git -C <repo> worktree add -b <id-N> <repo>/../.bb-worktrees/<id-N> <id-N-1>
  -- branch from the PREVIOUS bead's branch, not from HEAD. Inside a batch HEAD
  does not move: the orchestrator merges only after the batch returns, so
  branching every bead from HEAD gives beads that cannot see each other's work
  and conflict on any shared file.
Work that bead entirely inside its worktree. After its close commit lands, tell
the orchestrator and move on. WHERE you create your own, never hold two open at
once: a worktree that exists means "this bead is unfinished" to every other reader.
WHERE you were given one for the whole batch, that signal is coarser by
construction — it means "this BATCH is unfinished" — so report each bead's close
as it lands, because your report is the only per-bead signal the orchestrator has
until the batch returns.
Run dir: {RUN_DIR}
Scripts: {SCRIPTS}
Trailers file: {TRAILERS_FILE or "none"}   (Phase 0 resolved; pass to bb-close-bead.sh --trailers-file)
Checks baseline:
{CHECKS_BASELINE_TEXT}
    (The CONTENTS of {RUN_DIR}/checks-baseline.md, inlined verbatim — NOT the
     one-word status pre-flight emits as BB_CHECKS_BASELINE. Include the block
     even when the checks passed, because the file says so and a MISSING block
     means no baseline was taken, which is a different thing.)

For each bead, first invoke the `run-bead` skill — it carries the ceremony, and the blocked and skipped paths — passing this bead's id, priority, title and `--trailers-file {TRAILERS_FILE}`. Read the committed sha with `git rev-parse --short HEAD` rather than from the script's printout. Then run the template-specific steps below.

When the batch is done, return ONE message containing N+1 lines: one per-bead summary, then a final `batch-end: full gates {passed|failed}`.
```

## Doc/chore fast-lane subagent prompt template

```
You are processing {N} trivial beads in this repo.

Beads (process in order; one commit per bead):
{ID_1} (P{PRIORITY_1}) — {TITLE_1} — Decision: {DECISION_1}
{ID_2} (P{PRIORITY_2}) — {TITLE_2} — Decision: {DECISION_2}
...

Checks: {CHECKS}
Push mode: {push_per_bead | defer_push}

These are documentation, comments, single-line swaps, or other low-risk changes. Beyond the shared per-bead workflow above:

1. **Every bead runs `{CHECKS}` before its own close.** That is `run-bead` step 4, and being in a batch does not override it — an unvalidated commit is not cheaper, it is just unvalidated. **Targeted tests are non-negotiable on top of it** — identify the test that covers the touched surface and run it directly. If the change is so trivial there's no plausible test (e.g. pure docs edit, a comment with zero output diff), say so explicitly in a `retro:` comment on the bead and continue. "I couldn't find a targeted test" without explanation is not acceptable.
2. Compare each run against the checks baseline above: a line in both was already failing, a line only in yours belongs to the bead you are on. Because every bead runs the checks, the bead whose run first shows a finding is the one that introduced it — fix it there, before starting the next bead, rather than carrying it forward. **The push is the gate**: the last bead's passing run is what clears this batch for the orchestrator's push.

## Constraints

Same as the per-bead constraints. Each bead is its own commit. In `defer_push` mode, do not push at all — the orchestrator pushes after the batch.
```

## Shared-surface batch subagent prompt template

```
You are processing {N} related beads that share a code/fixture surface ({SURFACE_DESCRIPTION}, e.g. "the JSON serializer in src/serialize.* and the serializer golden fixtures").

Beads (process in order; each is its own commit):
{ID_1} (P{PRIORITY_1}) — {TITLE_1} — Decision: {DECISION_1} — Implicit: {IMPLICIT_1}
{ID_2} (P{PRIORITY_2}) — {TITLE_2} — Decision: {DECISION_2} — Implicit: {IMPLICIT_2}
...

Checks: {CHECKS}
Targeted tests for this surface: {SURFACE_TARGETED_TESTS}
Push mode: {push_per_bead | defer_push}

These beads are NOT trivial — they touch real code. They are batched because they would collide if dispatched side by side — later beads rebuild the same surface — so one agent takes them in order and the rest of the wave runs concurrently around them. **Batching changes who does the work, not what gets validated: every bead still runs `{CHECKS}` before its own close.**

Beyond the shared per-bead workflow above:

1. Before each bead's close, in this order — **none of it is optional**:
   - If generated output was touched, regenerate it with the project's own tooling — the repo's conventions doc names that command; this prompt does not carry it — then run the relevant fixture/snapshot test.
   - **Run the surface-relevant targeted tests** named above for the file(s) you touched. If none is listed for a surface you touched, find the test that covers it; if there's still nothing, that's a signal the change is broader than expected — pause and reconsider scope, don't take it as permission to skip.
   - **Run `{CHECKS}`.** `run-bead` step 4 requires it per bead and this batch does not override it.

   **Why this matters:** without per-bead targeted tests, a regression introduced in bead #2 only surfaces at batch end, after beads #3 and #4 may have built on the broken assumption. You then bisect, rework #2, and possibly redo #3 and #4. Targeted tests prevent that cascade.
2. Compare every run against the checks baseline above: a line in both was already failing, a line only in yours belongs to the bead you are on. Because every bead runs the checks, there is nothing to bisect — the bead whose run first shows a finding is where it entered. Fix it there before starting the next bead. WHERE investigating shows an earlier bead is actually at fault, follow the Blocked path for THAT bead rather than silently repairing it across several commits.
   **The push is the gate**: the last bead's passing run is what clears this batch, and whether CI also runs somewhere is the project's business, not a reason to skip it here.

## Constraints

- Each bead is its own commit. Never combine bead changes into one commit.
- If a bead's implementation conflicts with an earlier bead in this batch (e.g. one removed a function the next was about to modify), follow the Blocked path for the conflicting bead and continue with the rest.
- Same per-bead constraints as `run-bead`'s Constraints section.
```

## Worktree mechanics (orchestrator)

### The base-ref trap — read before the first wave

**Every worktree MUST branch from the sprint's own branch, never the
default branch.** Two conditions produce that, and both are checked: the
orchestrator's own HEAD is not the default branch (pre-flight blocks the
run otherwise), and `worktree.baseRef` must be `"head"` so the worktree
takes session HEAD rather than the repository's default branch. Unset,
the worktree branches from the repository's **default branch**, not
session HEAD. This is not conditional on whether the run's beads build on
each other: an unset `baseRef` silently drops every commit the sprint
branch carries from every worktree, isolated beads included — it is a hard
gate for `isolation: "worktree"` full stop, not a special case for beads
with a dependency edge. Fix (set it in user scope; project scope is untested):

```json
{ "worktree": { "baseRef": "head" } }
```

Takes effect immediately, no restart — a failed check is stop-and-fix, not
stop-and-abandon.

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
point at the main checkout, `bd list` returns the full set, and no
`embeddeddolt` or redirect file is created (with `BEADS_DIR` unset). All
worktree agents share one tracker, which is what makes this constraint real:

- **Never read `.beads/*.jsonl` as tracker state.** Ask the database: `bd show`,
  or `bd export` to stdout, which regenerates from it. There is no export step
  to run before committing — the close stages code only.

### The worktree lifecycle — create, dispatch, merge, remove

The orchestrator owns creation and removal; the subagent owns only what
happens inside. Per bead, in order:

```sh
REPO="$(git rev-parse --show-toplevel)"
WT="$REPO/../.bb-worktrees/<suffix>"        # sibling of the repo, never inside it
git -C "$REPO" worktree add -b <suffix> "$WT" HEAD
```

A worktree inside the repository would read as dirt in every `git status` the
run takes, which is why it is a sibling.

Branching from `HEAD` — the sprint branch as it stands *now*, after the
previous bead merged — is what makes each bead build on the last. Where the
harness creates the worktree for you (`Agent` with `isolation: "worktree"`),
it uses `worktree.baseRef` instead, which is why that setting is a hard gate —
**and it may place the worktree INSIDE the repository (observed:
`.claude/worktrees/`), which the sibling rule above exists to prevent. The path
is not yours to choose, so neutralise it: ensure the harness's worktree
directory is gitignored before the first dispatch.** Untracked worktrees inside
the repo are reported as uncommitted work by anything reading `git status`, and
committing them lands in-flight bead state on the sprint branch.

1. **Dispatch** the subagent with the per-bead template, telling it to work in
   `$WT`. It implements, gates, records findings, and
   closes with `../scripts/bb-close-bead.sh` — passing **no paths**, since everything in
   that tree is this bead's work.
2. **Accept**: verify the bead reads `closed` and its commit exists on the
   bead's branch. This is the acceptance check; the worktree still existing at
   this point is what tells you the bead is not yet done.
3. **Merge** into the sprint branch, one bead at a time. Take the merge slot
   first and release it after:

   ```sh
   bd merge-slot acquire            # --wait to queue behind a holder
   git -C "$REPO" merge --ff-only <suffix>
   bd merge-slot release
   ```

   `bd merge-slot create` once per run makes the slot bead (`<prefix>-merge-slot`)
   if the project has none. It works outside a Gas Town rig.

   **What the slot is for here, and what it is not.** Only the orchestrator
   merges, so within one run the serialization already holds by construction —
   the slot is not what makes merges one-at-a-time. It earns its place by making
   that state *visible*: `bd merge-slot check` names the holder, so a wave that
   looks stalled has an answer, and a second session that resumed the same run
   finds the slot held rather than merging concurrently with the first.

   Three measured behaviours to hold it correctly:

   - **Acquire is not idempotent.** Re-acquiring a slot you already hold fails
     with `slot held by: <you>`. So a failed release leaves your own next
     acquire failing, indistinguishable from a sibling holding it — release on
     every path out of a merge, including the failure paths.
   - **Release does not check the holder.** Any agent can release anyone's
     slot, and releasing an unheld slot succeeds silently. That is the failure
     path for a slot held by a dead agent — just release it — and equally the
     reason the slot is advisory: it records intent, it does not enforce it.
   - **The holder is the actor** (`$BEADS_ACTOR`, else git `user.name`), so set
     `BEADS_ACTOR` if you want the holder to name something more specific than
     the harness default.
   **`--ff-only` is deliberate — in a SEQUENTIAL wave.** Each bead branches
   from an already-updated head, so every merge is a fast-forward and a refusal
   means an invariant broke: a worktree branched from the wrong base. **Stop
   the wave and report it; do not resolve it inline** and do not retry without
   `--ff-only`.

   **In a CONCURRENT wave, only the first merge can fast-forward**, and that is
   not a fault — every branch in the wave shares one base by construction, so
   the second onward necessarily diverge and `--ff-only` refuses them with "Not
   possible to fast-forward". Applying the sequential rule literally would stop
   every concurrent wave of 2+ beads and report a healthy wave as broken.

   So for each branch after the first: **rebase it onto the updated sprint head,
   then fast-forward.**

   ```sh
   # Run from your own checkout, standing on the sprint branch.
   git -C <worktree> rebase "$(git rev-parse HEAD)"  # read the target NOW
   git merge --ff-only <suffix>                      # now it fast-forwards
   ```

   **Read the target at merge time, never earlier in the wave.** Each bead is
   merged into the sprint branch as it is accepted, so accepting the wave's
   first bead moves the head before its second is merged. A sha noted at wave
   start is already stale, and `--ff-only` is right to refuse it.

   **The merge slot does not cover this rule, so the rule stays.** The slot
   serializes *who merges*; staleness is a single merger reading a sha too
   early and still holding the slot when it uses it. Holding a lock across a
   stale read produces exactly the same wrong rebase, on time.

   Rebase keeps one-commit-per-bead and a linear history, so it is the
   concurrent analogue of the sequential guarantee. **The collision signal moves
   to the rebase:** a rebase CONFLICT means two beads genuinely changed the same
   lines, which is the failure `--ff-only` was standing in for. Stop the wave on
   a rebase conflict, exactly as you would on a sequential `--ff-only` refusal.
4. **Remove the worktree — the final act of accepting the bead:**
   ```sh
   git -C "$REPO" worktree remove "$WT" && git -C "$REPO" branch -d <suffix>
   ```
   **Raw `git`, not `bd worktree remove`, and the reason is not inertia.**
   `bd worktree remove` adds one safety check git lacks — it refuses a clean
   worktree whose branch carries unpushed commits — but the check treats a
   branch with **no upstream** as entirely unpushed. Every bead branch here is
   exactly that: created from the sprint head, never pushed on its own, merged
   locally. Measured against this repository: a bead-shaped worktree with *zero*
   new commits is refused with "worktree has unpushed commits". So the safety
   property is unreachable in this topology — the only way through is
   `--force`, which skips *all* the checks, including the dirty-tree one that
   plain `git worktree remove` still enforces. Adopting it would trade a real
   guard for a suppressed one.
   **Do not delete the branch until the fast-forward has succeeded.** A rebase
   onto a stale target leaves the rebased commit unreferenced, and deleting the
   branch is what makes it unreachable — recoverable only from the object
   store, and only if someone notices.

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
`isolation: "worktree"`, all in a single message. Waves themselves stay
sequential; concurrency is across beads *inside* a wave.

The lifecycle above is unchanged per bead. Two additions when running
concurrently: merge each branch as it returns rather than after all return, and
where a merge hits a fixture/snapshot conflict and the project has a regen
step, take both sides, regenerate on the merged tree, stage and amend. A
non-fixture conflict on shared code remains a triage failure — the beads were
not actually isolated; mark the unmerged bead `blocked` and let the user
resolve.

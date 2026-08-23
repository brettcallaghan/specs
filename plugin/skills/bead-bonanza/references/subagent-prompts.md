# bead-bonanza Subagent Prompt Templates

The prompts the orchestrator fills in and dispatches. Read this at Phase 3
(dispatch), and at Phase 1.5 when a bead needs grooming first. Substitute
every `{PLACEHOLDER}` before sending — an unsubstituted placeholder is the
single most common dispatch defect, because the subagent has no session
context to infer it from.

Five run-level placeholders are resolved once per run and substituted into every
prompt: `{EPIC}` is the run epic `sprint-scope` resolved (emitted as `BB_EPIC`)
and is also the run id, `{BRANCH}` is the sprint branch from the manifest,
`{SCRIPTS}` is the absolute path to this skill's `scripts/` directory,
`{CHECKS_IDS}` is the ids the project declared (emitted as `BB_CHECKS_IDS`),
and `{CHECKS_BASELINE_TEXT}` is the manifest's baseline —
`bd show {EPIC} --json | jq -r '.[0].metadata.checks_baseline'`.
A subagent can infer none of them.

**There is no run directory.** The run's record is the epic: the manifest in
its metadata, the amendments in its comments.

**`{EPIC}` is also where incidents land.** A subagent that has to decide
something changing what "correct" means writes
`retro:<kind>:<target> incident — …` on the bead and
`amend: incident on <id> — …` on the epic. `run-bead` carries the
rule; the prompt only has to supply the id, which no subagent can infer.

`{DISPATCH_ID}` is resolved **per subagent** — one id for a batch, shared by
every bead it holds, which is what makes "what is that agent holding" a query.

`{WORKTREE}` is resolved **per dispatch** — the absolute path to the worktree the
orchestrator created for it. A subagent works there and nowhere else; it
cannot infer the path, and a prompt that omits it sends the agent into the
shared checkout, which defeats every guarantee this design rests on.

## Contents

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

## Per-bead subagent prompt template

```
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
Trailers: {"on the run epic" if any}   (pass `--trailers-from {EPIC}` to bb-close-bead.sh; omit it when the run resolved none)
Checks baseline:
{CHECKS_BASELINE_TEXT}
    (The run epic's `.metadata.checks_baseline`, inlined verbatim — NOT the
     one-word status pre-flight emits as BB_CHECKS_BASELINE. Include the block
     even when the checks passed, because the text says so and a MISSING block
     means no baseline was taken, which is a different thing.)

Before your first action, record that you hold this bead:

```sh
bd update {ID} --set-metadata dispatch_id={DISPATCH_ID} \
  --set-metadata dispatch_state=running \
  --set-metadata worktree={WORKTREE} \
  --set-metadata dispatched_at="$(date -u +%FT%TZ)"
```

and as your last action, whatever the outcome:

```sh
bd update {ID} --set-metadata dispatch_state=returned \
  --set-metadata terminal_state=completed|blocked|skipped
```

That pair is what lets an orchestrator tell a live agent from a dead one
without guessing at your worktree. Write the second even when you are blocked —
especially then.

**Then invoke the `run-bead` skill.** It carries the ceremony — claim, read,
implement in scope, gate, caveat, findings, close-and-commit, push, return —
together with the blocked and skipped paths and its own acceptance criteria.
The template above supplies its inputs; it supplies the procedure.
```

## Shared batch preamble

Both batch templates below (doc/chore fast-lane, shared-surface) open with
this block rather than repeating it. **The orchestrator concatenates this
preamble plus the chosen template into the one text it sends** — a batch
prompt assembled without it is missing the `sprint-batch` invocation and the
return contract, with nothing to catch the loss downstream.

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

## Doc/chore fast-lane subagent prompt template

```
You are processing {N} trivial beads in this repo.

Beads (process in order; one commit per bead):
{ID_1} (P{PRIORITY_1}) — {TITLE_1} — Decision: {DECISION_1}
{ID_2} (P{PRIORITY_2}) — {TITLE_2} — Decision: {DECISION_2}
...

Checks: {SCRIPTS}/bb-checks.py       (declared ids: {CHECKS_IDS})
Push mode: {push_per_bead | defer_push}

These are documentation, comments, single-line swaps, or other low-risk changes. Beyond `sprint-batch`'s per-bead loop:

1. **Every bead runs `{SCRIPTS}/bb-checks.py` with no arguments before its own close.** That is `run-bead` step 4, and being in a batch does not override it — an unvalidated commit is not cheaper, it is just unvalidated. **Targeted tests are non-negotiable on top of it** — identify the test that covers the touched surface and run it directly. If the change is so trivial there's no plausible test (e.g. pure docs edit, a comment with zero output diff), say so explicitly in a `retro:` comment on the bead and continue. "I couldn't find a targeted test" without explanation is not acceptable.
2. Compare each run against the checks baseline above: a line in both was already failing, a line only in yours belongs to the bead you are on. Because every bead runs the checks, the bead whose run first shows a finding is the one that introduced it — fix it there, before starting the next bead, rather than carrying it forward. **The push is the gate**: the last bead's passing run is what clears this batch for the orchestrator's push.

## Constraints

Same as `run-bead`'s constraints. In `defer_push` mode, do not push at all — the orchestrator pushes after the batch.
```

## Shared-surface batch subagent prompt template

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
happens inside.

**One worktree per DISPATCH, not per bead.** A batch is one agent working its
beads serially, so a second worktree isolates it from nobody — the branches are
what keep the beads separate, and the agent makes those inside the one tree
with `git checkout -b`. A single bead dispatched alone is the same rule with
one bead in it.

```sh
REPO="$(git rev-parse --show-toplevel)"
WT="$REPO/../.bb-worktrees/<dispatch-id>"   # sibling of the repo, never inside it
git -C "$REPO" worktree add -b <first-suffix> "$WT" HEAD
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

1. **Dispatch** one subagent into `$WT`. It implements, gates, records findings,
   and closes each bead with `../scripts/bb-close-bead.sh` — passing **no
   paths**, since everything dirty in that tree is the current bead's work, the
   previous bead having committed before this one started. A batch subagent
   cuts its own branch per bead there with `git checkout -b <suffix>`.
2. **Accept, per bead**: verify each bead reads `closed` and its commit exists
   on that bead's branch.
3. **Merge, per dispatch.** A chained batch fast-forwards the LAST bead's
   branch, which already carries its predecessors — one merge, not one per
   bead. Take the merge slot first and release it after:

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
     slot, and releasing an unheld slot succeeds silently — so the remedy for a
     slot held by a dead agent and the way to corrupt a live merge are the same
     command. **Do not decide that by eye: run
     `{SCRIPTS}/bb-status.sh slot {EPIC}`.** It prints the holder beside the
     evidence for and against them still being alive — standing worktrees, the
     age of the last commit, and which dispatches report `running` — and
     recommends a release only when every signal agrees. It never releases
     anything itself, so a wrong reading costs a second look rather than
     someone else's merge.
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
4. **Remove the worktree — the final act of accepting the dispatch**, once
   every bead in it has merged:
   ```sh
   git -C "$REPO" worktree remove "$WT"
   git -C "$REPO" branch -d <suffix>          # each of the batch's branches
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
   of the in-flight signal than "not yet committed". At batch granularity it
   reads "this DISPATCH is unfinished"; which of its beads is unfinished comes
   from the dispatch record, not from counting trees.
5. **Push** per the batch's push mode.

WHEN a bead is blocked rather than finished, **leave its worktree in place**
and say so in the report. It holds the partial work, and its existence is the
durable record that the bead is unfinished — there is nothing to stash and no
WIP commit to make (`operating-principles.md`).

**In a batch that means the dispatch ends there IF the block left work in the
tree.** One worktree cannot both preserve a blocked bead's uncommitted changes
and give the next bead a clean tree to start from, and preserving them wins —
the alternative is the next bead committing someone else's half-finished work.
A bead blocked with a **clean** tree (nothing started, or a prerequisite found
before any edit) costs nothing and the batch continues. The skip-verification
script, run in the worktree, is the test, and its verdict is what the return
line reports.

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

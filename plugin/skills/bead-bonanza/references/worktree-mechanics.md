# Worktree mechanics (orchestrator)

**Not a prompt.** This is the orchestrator's own procedure: how a dispatch's
worktree is created, verified, merged and removed. No subagent reads it — a
subagent is told only its `{WORKTREE}` path, by
[prompt-placeholders.md](prompt-placeholders.md).

## Contents

- The base-ref trap — why every worktree must branch from the sprint branch,
  and the per-wave check that proves it did
- Beads inside a worktree — `bd` resolves the shared tracker from any of them
- The worktree lifecycle — create, dispatch, accept, merge (slot, `--ff-only`,
  rebase), remove
- Concurrent dispatch within a wave


## The base-ref trap — read before the first wave

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

## Beads inside a worktree

`bd` is not an obstacle: it resolves the shared tracker database through
git's common directory from any worktree — `bd where`/`bd worktree info`
point at the main checkout, `bd list` returns the full set, and no
`embeddeddolt` or redirect file is created (with `BEADS_DIR` unset). All
worktree agents share one tracker, which is what makes this constraint real:

- **Never read `.beads/*.jsonl` as tracker state.** Ask the database: `bd show`,
  or `bd export` to stdout, which regenerates from it. There is no export step
  to run before committing — the close stages code only.

## The worktree lifecycle — create, dispatch, merge, remove

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

## Concurrent dispatch within a wave (enabled)

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

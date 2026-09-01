# Worktree mechanics (coordinator)

**Not a prompt.** This is the coordinator's own procedure: how a dispatch's
worktree is created, verified, merged and removed. No subagent reads it, and
none is told the path either — an agent resolves its own worktree from the tree
it is standing in (`dispatching.md`).

## Contents

- The sprint branch — created once, reused on a resume, and where `base_sha`
  comes from
- The base-ref trap — why every worktree must branch from the sprint branch,
  and the per-dispatch check that proves it did
- Beads inside a worktree — `bd` resolves the shared tracker from any of them
- The worktree lifecycle — create, dispatch, accept, merge, remove
- Merging — one branch per dispatch, the two meanings of a `--ff-only` refusal,
  and the ordering guard removal sits behind
- A rebase conflict is a re-planning problem, and who resolves it
- Concurrent dispatch out of one ready set

## The sprint branch — created once, before the first worktree

The coordinator's, and only its. Run it immediately before the first dispatch:

```bash
out="$(swarm manifest get --run "$BB_EPIC" --key branch --json)"; st=$?
br="$(printf '%s' "$out" | jq -r '.branch // empty')"
if [ "$st" = 0 ]; then                            # RESUME — create nothing
  git checkout "$br"
else
  br="sprint/<epic-suffix>"
  git rev-parse --verify "$br" >/dev/null 2>&1 && { echo "branch $br taken"; exit 1; }
  git checkout -b "$br"
  swarm manifest set --run "$BB_EPIC" \
    --key branch   --value "\"$br\"" \
    --key base_sha --value "\"$(git rev-parse HEAD)\""
fi
```

**The manifest read comes first.** Reuse is what makes a resumed run land on the
same branch as the original; creating a second one forks the run's history from
the work it is meant to continue. A name already taken by an unrelated branch
stops the run here, before any worktree exists.

**Both keys go in ONE `swarm manifest set`.** The command applies a whole
invocation as a single merge, so a run that wrote them one at a time could
leave a branch recorded with no base against it. It also refuses to change
either once written — which is what makes re-running this block on a resume
safe rather than a second fork. The read above captures the command's output
first and takes `$?` from that, then reads the key out of it: an `attention`
status (exit 1) is how "not written" reaches you, and its answer carries
`branch.ok: false` rather than a value, so `jq -r '.branch // empty'` leaves the
variable empty. Do not put a pipe between the command and `$?` — that would
hand you the last stage's status instead.

**`base_sha` is recorded at creation** because that is the only moment it is
unambiguous — deriving it later from a merge-base assumes the base branch has
not moved, and on a long run it will. Three things consume it and none may
create it: `run-record.md`.

## The base-ref trap — read before the first dispatch

**Every worktree MUST branch from the sprint's own branch, never the
default branch.** Two conditions produce that, and both are checked: the
coordinator's own HEAD is not the default branch (pre-flight blocks the
run otherwise), and `worktree.baseRef` must be `"head"` so the worktree
takes session HEAD rather than the repository's default branch. Unset,
the worktree branches from the repository's **default branch**, not
session HEAD. This is not conditional on whether the run's beads build on
each other: an unset `baseRef` silently drops every commit the sprint
branch carries from every worktree, isolated beads included — it is a hard
gate for `isolation: "worktree"` full stop, not a special case for beads
with a dependency edge. Fix — commit it to the repository, so every clone and
every fresh container inherits it:

```json
// .claude/settings.json
{ "worktree": { "baseRef": "head" } }
```

The same key in `~/.claude/settings.json` works for one machine. Claude Code
resolves it highest-precedence-first — `.claude/settings.local.json`, then
`.claude/settings.json`, then `~/.claude/settings.json` — so a higher level
overrides a lower one, and a stray `"fresh"` in the local file defeats a
committed `"head"`.

Takes effect immediately, no restart — a failed check is stop-and-fix, not
stop-and-abandon.

**Verify before every dispatch** — this is global state the repo doesn't own and
can change without warning. Note session HEAD (`git rev-parse --short HEAD`)
and give the **first** worktree subagent this instruction:

```
Before doing anything else, run in your worktree:
  git rev-parse --short HEAD
  git merge-base --is-ancestor <SESSION_HEAD_SHA> HEAD; echo "ancestor-exit=$?"
Report both lines verbatim as the first line of your reply.
```

WHEN `ancestor-exit` is non-zero, **do not dispatch**: wrong base, and do not
merge anything back. Fix `worktree.baseRef`, or fall back to one bead at a
time.

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

The coordinator owns creation, merging and removal; the subagent owns only what
happens inside.

**One worktree and ONE BRANCH per DISPATCH.** A batch is one agent working its
beads serially, so a second worktree isolates it from nobody — and a branch per
bead buys nothing either, because **one commit per bead is what separates the
beads**, and the ceremony already guarantees it.

**Attribution reads the commits, never the branch name.** Coordinator-created
trees below name the branch for the dispatch id, but nothing downstream —
`swarm status`, `../scripts/sprint-cleanup.sh` — relies on that convention: each worktree
is matched to its dispatch through the bead its HEAD commit closed
(`<id> P<n> <title>`, the same handle recovery reads) and that bead's own
`.assignee`. This is what lets the SAME mechanism attribute a worktree the
harness itself created — see below — whose branch carries no dispatch id at
all.

```sh
REPO="$(git rev-parse --show-toplevel)"
WT="$REPO/../.bb-worktrees/<dispatch-id>"   # sibling of the repo, never inside it
git -C "$REPO" worktree add -b "bb/<dispatch-id>" "$WT" HEAD
```

A worktree inside the repository would read as dirt in every `git status` the
run takes, which is why it is a sibling — when this creation path runs at all.
It is the manual fallback; the path a dispatch actually takes today is the one
below, and its worktree is not a sibling.

Branching from `HEAD` — the sprint branch as it stands *now*, after the
previous dispatch merged — is what makes each dispatch build on the last. Where
the harness creates the worktree for you (`Agent` with `isolation: "worktree"`),
it uses `worktree.baseRef` instead, which is why that setting is a hard gate —
**and it may place the worktree INSIDE the repository (observed:
`.claude/worktrees/`), which the sibling rule above exists to prevent. The path
is not yours to choose, so neutralise it: ensure the harness's worktree
directory is gitignored before the first dispatch.** Untracked worktrees inside
the repo are reported as uncommitted work by anything reading `git status`, and
committing them lands in-flight bead state on the sprint branch.

**The branch it checks out is also not yours to choose, and that is fine.**
Measured against this harness: it names the branch `worktree-agent-<hash>` —
the harness's own agent id, carrying no dispatch id at all, nothing like
`bb/<dispatch-id>`. Nothing downstream cares: attribution reads the HEAD
commit, not the branch, as stated above. Do not "fix" this by trying to rename
or re-point the branch after the fact.

1. **Dispatch** one subagent into `$WT`. It implements, gates, records findings,
   and closes each bead with `../scripts/bb-close-bead.sh` — passing **no
   paths**, since everything dirty in that tree is the current bead's work, the
   previous bead having committed before this one started.
2. **Accept, per bead**: verify each bead reads `closed` and its commit exists
   on the dispatch's branch.
3. **Merge, per dispatch** — below.
4. **Remove the worktree** — the final act of accepting, and **only after the
   merge succeeded**:
   ```sh
   git merge-base --is-ancestor "bb/<dispatch-id>" HEAD || { echo "NOT merged"; exit 1; }
   git -C "$REPO" worktree remove "$WT"
   git -C "$REPO" branch -d "bb/<dispatch-id>"
   ```
   **The ancestry check is the guard, and it is not decoration.** A chained
   command whose merge exited 128 removed the tree anyway, and for ninety
   seconds a removed worktree coexisted with an unmerged branch — the one state
   the in-flight signal asserts cannot occur. `&&` between two commands is not
   a guard; this is.

   **Raw `git`, not `bd worktree remove`, and the reason is not inertia.**
   `bd worktree remove` adds one safety check git lacks — it refuses a clean
   worktree whose branch carries unpushed commits — but the check treats a
   branch with **no upstream** as entirely unpushed. Every dispatch branch here
   is exactly that: created from the sprint head, never pushed on its own,
   merged locally. Measured against this repository: a bead-shaped worktree with
   *zero* new commits is refused with "worktree has unpushed commits". So the
   safety property is unreachable in this topology — the only way through is
   `--force`, which skips *all* the checks, including the dirty-tree one that
   plain `git worktree remove` still enforces. Adopting it would trade a real
   guard for a suppressed one.

   Removal is the coordinator's, not the subagent's: an agent cannot remove
   the tree it is standing in. Because removal happens *after* the merge, a
   surviving worktree means "not yet accepted", which is the stronger reading
   of the in-flight signal than "not yet committed". It reads "this DISPATCH is
   unfinished"; which of its beads is unfinished comes from the beads'
   statuses, not from counting trees.
5. **Push** the sprint branch, then verify it landed (`dispatching.md`).

WHEN a bead is blocked rather than finished, **leave its dispatch's worktree in
place** and say so in the report. It holds the partial work, and its existence
is the durable record that the bead is unfinished — there is nothing to stash
and no WIP commit to make (`operating-principles.md`).

**In a batch that means the dispatch ends there IF the block left work in the
tree.** One worktree cannot both preserve a blocked bead's uncommitted changes
and give the next bead a clean tree to start from, and preserving them wins —
the alternative is the next bead committing someone else's half-finished work.
A bead blocked with a **clean** tree (nothing started, or a prerequisite found
before any edit) costs nothing and the batch continues. The skip-verification
script, run in the worktree, is the test.

**Prune before trusting the list.** A reclaimed container deletes worktree
directories without deregistering them, and a stale entry reads as in-flight
forever: `git worktree prune` before `git worktree list` in pre-flight and in
recovery.

## Merging — the coordinator does it, one dispatch at a time

**No queue, no merge bead, no lock — do not add one.** You are one agent, so
you merge one branch at a time by construction and there is nothing left for a
mechanism to serialize. Merge each dispatch's branch **as that dispatch
returns**, rather than after all of them return: a branch left unmerged is a
base every later dispatch is missing.

```sh
git merge --ff-only "bb/<dispatch-id>"
```

**A `--ff-only` refusal means one thing where dispatches were SERIALIZED and
another where they went out CONCURRENTLY, and reading it wrong stops a healthy
run.** Name which case you are in *before* you read the refusal, mechanically:

```sh
git merge-base --is-ancestor "$(git rev-parse HEAD@{1})" "bb/<dispatch-id>"
```

- **Serialized** — this branch was cut from the current head, so a fast-forward
  is guaranteed and a refusal means an invariant broke: the worktree branched
  from the wrong base. **Stop and report it; do not resolve it inline** and do
  not retry without `--ff-only`.
- **Concurrent** — every branch dispatched off one ready set shares its base by
  construction, so only the first can fast-forward and the second onward
  necessarily diverge. That is **not a fault**; applying the serialized rule
  literally would stop every concurrent dispatch of two or more beads and
  report a healthy run as broken.

For each branch after the first: **rebase it onto the updated sprint head,
then fast-forward.**

```sh
# Run from your own checkout, standing on the sprint branch.
git -C <worktree> rebase "$(git rev-parse HEAD)"  # read the target NOW
git merge --ff-only "bb/<dispatch-id>"            # now it fast-forwards
```

**Read the target at merge time, never earlier.** Each dispatch is merged as it
is accepted, so accepting the first moves the head before the second is merged.
A sha noted at dispatch time is already stale, and `--ff-only` is right to
refuse it.

Rebase keeps one-commit-per-bead and a linear history, so it is the concurrent
analogue of the sequential guarantee. **Do not delete the branch until the
fast-forward has succeeded** — a rebase onto a stale target leaves the rebased
commit unreferenced, and deleting the branch is what makes it unreachable.

### A rebase conflict is a re-planning problem

**The collision signal moves to the rebase:** a conflict means two beads
genuinely changed the same lines, which is the failure `--ff-only` was standing
in for. The packing is built on the claim that these beads do not touch each
other, so a conflict is that claim failing.

**Record it first, before doing anything else**, as kind `sequencing` — a
blocker the dependency graph did not encode — targeted at the bead it collided
with:

```sh
bd comment <bead> "retro:sequencing:<other bead> — rebase conflict on <files>; no collision edge separated them"
```

**Then dispatch ONE agent to resolve it, in that dispatch's worktree.**
Resolving is not yours: you hold no bead's context, you made neither change,
and a resolution you invent is an unreviewed edit landing under someone else's
commit. Hand the agent the conflicted worktree, both beads' ids, and the
instruction to resolve, run the declared checks, and hand the branch back —
nothing else. **One resolver at a time**, which holds for the same reason
merging does: you dispatch it, and you are one agent.

**Stop DISPATCHING NEW WORK while a resolution is in flight, not the run.**
Every bead already merged stays merged and every branch already returned is
still mergeable; what is unsafe is sending further work onto a base whose
isolation assumption just failed. WHEN the resolver cannot separate the two
changes, mark the unmerged bead `blocked`, leave its worktree standing, and
report — two beads that genuinely collided belong in ONE dispatch, sequential,
which is what a collision edge between them expresses.

## Concurrent dispatch out of one ready set (enabled)

Spawn several dispatches at once — one `Agent` call each with
`isolation: "worktree"`, all in a single message. What may go out together is
whatever the ready set holds: collision edges keep anything that would conflict out
of it.

The lifecycle above is unchanged per dispatch. Two additions when running
concurrently: merge each branch as it returns rather than after all return, and
where a merge hits a fixture/snapshot conflict and the project has a regen
step, take both sides, regenerate on the merged tree, stage and amend. A
non-fixture conflict on shared code is the re-planning problem above.

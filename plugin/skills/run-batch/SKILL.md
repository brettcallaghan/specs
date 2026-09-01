---
name: run-batch
description: Take a batch of beads — one or many — from claimed to committed.
  The per-bead ceremony with its gate step, findings capture and
  close-and-commit script, plus the batch rules that keep several beads on one
  shared surface honest. Use when dispatched as a sprint's developer subagent
  over the beads assigned to a dispatch id. Covers resolving the assignment
  from the tracker,
  claiming atomically, staying in scope while filing what you find, running the
  project's gates, recording findings as retro-prefixed comments, and closing so
  each bead's closure and its code land in one commit. Also carries the blocked
  and skipped paths, which are where a bead most often goes wrong.
version: "0.4"
user-invocable: false
---

# run-batch

Take **one bead at a time** from claimed to committed. The order is the
substance: the close is folded into the implementation commit, so a step taken
out of turn produces a bead that is closed but uncommitted, or committed but
unclosed.

**One bead or several, it is the same ceremony.** Dispatched with a batch you
work the beads **in order, one commit each, and account for every one of
them** — a batch returning fewer lines than your queue held is the failure
this tier exists to prevent. One bead is a batch of one.

## Inputs — three, and you resolve the rest

**Given, because nothing you can run would tell you:**

| | |
|---|---|
| `{DISPATCH_ID}` | your agent id. It is the `assignee` on every bead in this batch, and your `$BEADS_ACTOR` |
| `{EPIC}` | the run epic; `swarm manifest --run {EPIC}` reads its manifest |
| `{SCRIPTS}` | absolute path to the coordinator's `scripts/` directory |

**Everything else you resolve, in this order, before touching a bead.** A
dispatch prompt that carried these would be a second copy of facts the tracker
already holds, and a placeholder somebody forgot to substitute:

```sh
export BEADS_ACTOR={DISPATCH_ID}          # FIRST, in every shell you run bd from
SCRIPTS={SCRIPTS}

WORKTREE=$(git rev-parse --show-toplevel)                    # you are standing in it
swarm manifest require --run {EPIC} --key branch --key checks_baseline   # route on the status
RECORD=$(swarm manifest get --run {EPIC} --key branch --key checks_baseline --json)
BRANCH=$(printf '%s' "$RECORD" | jq -r .branch)              # the values themselves
BASELINE=$(printf '%s' "$RECORD" | jq -r .checks_baseline)
CHECKS_IDS=$($SCRIPTS/bb-checks.py --ids)
bd ready --parent {EPIC} --assignee {DISPATCH_ID} --exclude-type=epic   # your queue
bd list  --assignee {DISPATCH_ID} --status blocked                     # already blocked
```

**The manifest is `swarm manifest`'s, never the tracker's** — the command owns
the key set and the types, and your own `bd show | jq` is a second reader that
will disagree with it.

**Read each bead in full with `bd show <id>` when you reach it** — its Design,
its acceptance criteria and any decision it pins are there, and a summary
handed to you would be a lossy copy of what you are about to read anyway.

**An empty queue is a dispatch error, not an empty batch.** Report it and stop;
do not go looking for work to do.

**A missing manifest key is reported, never guessed** — `require` answers that
as a status, so route on it rather than testing for an empty string. An absent
`checks_baseline` in particular is not the same as one that passed; see the
checks step.

**Never work in the main checkout**, and never on the default branch.
**You never push and you never merge.** The coordinator merges your branch when
you return and pushes after it; a subagent that pushes lands work the
coordinator has not accepted.

## Two checks on the tree, before you touch a bead

Cheap, and each has caught a run. The queue itself you already resolved above.

1. **The worktree is clean** — `git status --porcelain` in it, empty.
2. **The base ref is the sprint branch, not the default branch:**

   ```sh
   git -C "$WORKTREE" merge-base --is-ancestor "$(git rev-parse "$BRANCH")" HEAD
   ```

   **This is the one that silently destroys work.** A worktree branched from
   the default branch cannot see anything the sprint branch already carries, so
   every bead here would be written against a tree missing its predecessors'
   commits — and the merge afterwards looks like a conflict rather than like a
   mistake. Report and **do not start**.

**A bead in your queue that is not assigned to you is a dispatch error**:
report it and do not start, rather than claiming your way into someone else's
queue. A bead you expected that is missing is blocked behind something — say
which, from `bd show <id>`.

**One worktree, one branch, both already yours.** The branch is named for the
dispatch and it is where every bead commits. **Do not cut a branch per bead and
do not create a second worktree**: the beads are separated by being committed
one at a time, which the ceremony already guarantees, and a second tree
isolates a serial agent from nobody.

## Per bead — follow in order

Progress is reported by the tracker, not by you: `bd` emits an event on every
status change, so claiming and closing the bead is what an orchestrator sees.
Write no progress file.

1. **Claim** — `bd update <id> --claim`. Atomic, and idempotent when the claim
   is already yours. It sets status **and** assignee in one write, where
   `--status in_progress` sets only the status and leaves the bead claimed by
   nobody. The assignee comes from the actor (`$BEADS_ACTOR`, else git
   `user.name`), so exporting your own `BEADS_ACTOR` makes you identifiable in
   the tracker as the agent holding this bead.

2. **Read it in full** — `bd show <id>`.

3. **Implement, strictly in scope.**

   **If this bead fixes a defect or regression, first write (or locate) a test
   that reproduces it and watch it FAIL for the stated reason BEFORE you fix
   it.** A test written after the fix encodes the fix's own assumptions and can
   stay green while the defect is live. Record the observed failure — it goes
   in a `retro:` comment at step 6.

   **A claim about what a command does is measured HERE, before you act on
   it.** The bead, your prompt, a reference file and a prior run's report all
   state such claims as plain prose, and prose records nothing about whether
   anybody ran them — an inherited claim and a measured one are
   indistinguishable on the page. So run the command, read the output, and
   carry the invocation and its output wherever you restate the claim. Where
   you cannot run it, write the claim as **unverified** and say what would
   settle it. This binds hardest when you are *correcting* someone else's
   claim: the replacement is a new claim and carries the same burden.

   **Finding a claim false is a sweep, not an edit.** Grep its distinctive
   phrasing across the tree *before* you fix the source — afterwards the string
   is gone, and the one place you were sure of it is the place you just
   destroyed. Then sweep the copies nobody re-reads: **durable trigger prompts
   first**, since they outlive the run and hand the retracted claim to a
   resuming agent as a constraint it cannot know was withdrawn; then sibling
   dispatch prompts, bead descriptions, epic comments, plan documents, and
   generated copies. Record any copy you could not reach as a `retro:`
   finding.

   **Work you find that is not this bead's** is filed, not done and not
   dropped: `bd create` with a `discovered-from` link to `<id>`, and one line
   saying whether it *blocks* this bead (a prerequisite, an
   acceptance-criteria decomposition, an expected test-coverage gap) or is
   merely adjacent. Do not implement it. This includes a gap nobody owns — a
   document no bead claims, a step every adjacent bead assumed another would
   take. **"Not this bead's to close" is a reason to FILE it, not a reason to
   leave it unfiled.** If the bead genuinely cannot be finished without it,
   take the Blocked path naming the new bead as the blocker; otherwise carry
   on with the original and do not widen scope inline.

   **Cite a durable anchor, never a position.** A path plus a quoted string, or
   a class, function or component name, survives an edit; a line number rots
   between filing and dispatch. Mark one as a hint beside the string, never as
   the handle.

4. **Checks — the close is the gate.** Run `$SCRIPTS/bb-checks.py` here. It
   runs everything the project declared, in declaration order, and keeps going
   after a failure so you see all of them at once.

   **While iterating, narrow it:** `bb-checks.py <id>...` runs only those, and
   `$CHECKS_IDS` says which exist. Pick by what you touched — the ids are
   slugs, so `test-go` covers Go tests. **Before your close, run it with NO
   arguments**: a narrowed run is a fast loop, never the gate.

   **Run the targeted tests too.** If the bead names tests for the surface it
   touches, run them; if none is named for a surface you touched, find the test
   that covers it. **Nothing there is a signal that the change is broader than
   expected** — reconsider scope rather than reading it as permission to skip.
   Where the change is genuinely untestable (a pure docs edit), say so in a
   `retro:` comment and continue.

   On failure, fix the root cause and rerun. **After two full failed cycles,
   take the Blocked path** — do not keep retrying.

   **A check that was already failing before you started is not yours to fix.**
   `$BASELINE` is what the checks printed on the untouched tree. The bar is
   **"no findings beyond that", not "exit 0"**.

   Compare, do not parse: run the checks, and read your output against the
   baseline's. A line in both was already there; a line only in yours is yours.

   A baseline that exited 0 means every check is expected to pass, so any
   failure is yours. An **unwritten** `checks_baseline` is different — none was
   taken, so you cannot tell a pre-existing failure from one you caused. Say so
   rather than guessing.

   **Being in a batch is not a reason to run less.** Because every bead runs
   the checks there is nothing to bisect: the bead whose run first shows a line
   absent from the baseline is where it entered. **Fix it in that bead, before
   starting the next.** Where an *earlier* bead turns out to be at fault, take
   the Blocked path for that bead rather than repairing it across commits.

5. **WHEN the close must do something differently, record it on the bead
   first** — `bd comment <id> "caveat: <what and why>"`. A regen step that must
   run before staging; a fixture this bead regenerated. **Only when there is
   one.** It goes on the bead because a recovering agent reads the tracker, and
   the tracker outlives the container.

6. **Record findings as `retro:` comments on <id>** — one `bd comment` per
   finding. That channel is what later retrospectives read.

   ```
   retro:<kind>[:<target>] <one line: what happened>
   evidence: <quote | path + quoted string | command + output | expected vs observed>
   ```

   **`kind` is one of nine, and it is what routes the finding** — so it costs
   a word now and saves a full re-read later:

   | | |
   |---|---|
   | `gap` | you needed information the bead and its references did not carry |
   | `decision` | you settled something the bead or spec should have pinned |
   | `stale` | the bead contradicted repo reality |
   | `unverifiable` | you could not satisfy an acceptance criterion as declared |
   | `sequencing` | a broken predecessor, or a blocker the graph did not encode |
   | `scope` | the bead was too large, trivial, or overlapped another |
   | `tooling` | a skill, script or template misled, failed, or had no rule |
   | `proof` | how you established a claim — a pre-fix failure observed, a negative control |
   | `environment` | a durable fact about a dependency or platform the next agent will hit |

   **Nothing here is a fifth thing you escalate for.** Four kinds of decision
   change what "correct" means rather than how the work is done — an acceptance
   criterion that needs amending, a standard and an AC in tension with no
   conservative branch, a contract or interface change, a spec silent where the
   bead requires it to speak. **Decide anyway, do not stop and do not ask**,
   and record the decision as a finding of its own kind: `decision` for the
   first two, `scope` for the third, `gap` for the fourth. Name the alternative
   you rejected and what it would have cost. A recorded decision gets reviewed
   at the pull request and revised in a later bead, which is cheaper than a run
   blocked on a question nobody is awake to answer.

   **Evidence in full; analysis not at all.** No length limit — the quote, the
   command and its output do not survive this session, and whoever reads your
   finding reads it cold. Root cause, proposal and priority are derivable later
   by an agent with the whole run in view. Include what you **tried that did
   not work** and what you **decided silently** — choices a diff-reader would
   not recognise AS choices.

   **`target` decides who gets it, and the test is: where would the fix
   land?** In this project's own repository → it is the project's. In the
   repository that vends the skill → it is the skill's. The words will not
   tell you apart — "the template has no rule for X" and "our document has no
   rule for X" read identically — so ask the question rather than guessing.

   Full format: `references/shared/findings-capture-protocol.md`.

   **Work, not friction.** If what you found is a unit of work someone must do,
   it is a BEAD (step 3), not a retro comment, even when it also reads as
   friction. A retro comment is read after the run; a bead is read during it.

   **A clean bead records none — say so rather than inventing filler.**

7. **Close and commit in one step** — run the script, do not perform it:

   ```sh
   $SCRIPTS/bb-close-bead.sh --trailers-from {EPIC} <id> <priority> "<title>"
   ```

   **The trailers live in the run's manifest and the script fetches them**
   before it closes or commits anything: pass `--trailers-from`; where the run
   resolved none it exits 2 with nothing done, so re-run without the flag.
   **Pass no paths** —
   everything modified in the worktree is this bead's work, and a path list can
   only be too narrow. That holds inside a batch's shared tree too: the previous
   bead committed before this one started, so nothing uncommitted is anyone
   else's.

   It performs the ceremony in the required order: atomic close
   (`bd update --status closed`, never `bd close`, which would be a second
   commit), verify the close against the DATABASE before staging, stage code
   only (`git add -A -- ':!.beads/*.jsonl'`), one commit subject
   `<id> P<priority> <title>` with the trailers in its body, then `bd dolt
   push` — the push, not a file, is what makes the closure durable. It refuses
   to run on the default branch or with nothing staged. Do not `git commit
   --amend` afterwards: a missing trailer means the trailers file was resolved
   wrong, which is not something to patch here.

   Read the committed sha with `git rev-parse --short HEAD`, not from the
   script's printout.

8. **Return one line** for this bead, then start the next:

   ```
   <id>: completed - <one clause on what changed> @ <short-sha>
   ```

## Blocked path

Gates failed twice, the change is intractable, or a prerequisite you filed
stands in the way:

- `bd update <id> --status blocked --append-notes "<short reason>"` — name the
  blocking bead's id when there is one.
- **Leave your uncommitted changes exactly where they are.** Do NOT stash and
  do NOT commit them. The worktree holds the partial work and its continued
  existence is what tells an orchestrator this bead is unfinished; stashing
  moves the work out of it and destroys that signal.
- Return the blocked line.

**Whether the batch continues depends on the tree, not the reason.** The
partial work is sitting where the next bead would start, so run
`$SCRIPTS/bb-verify-skip.sh` in your worktree:

- **CLEAN (exit 0)** — nothing was started. Continue with the next bead.
- **DIRTY (exit 1)** — the batch ends here. Return the blocked bead with its
  reason and every remaining bead as `blocked — not reached (batch ended at
  <id>)`. Leave the tree exactly as it is: committing or stashing it to free
  the tree destroys the thing the block exists to preserve.

**A bead that conflicts with an earlier bead in this batch** — one removed what
the next was about to change — takes this path.

## Skipped path

The bead is already addressed, malformed, or duplicates a closed bead:

1. **Confirm your tree is clean — a hard precondition, not a courtesy.** Run
   `$SCRIPTS/bb-verify-skip.sh` before anything else; it must exit 0. If it
   exits 1 it prints what is dirty and the two ways out: finish the close as
   `completed` (the work is this bead's), or revert the changes and re-run
   until CLEAN. Do not proceed while it exits 1.
2. `bd update <id> --status closed --append-notes "duplicate of <id>"`, or a
   similar one-line rationale.
3. Return the skipped line **with a citation**:
   `<id>: skipped - already addressed by <bead-id> @ <short-sha>`. "Already
   done" without one is not acceptable — nobody can verify it. Find it with
   `git log --oneline --all | grep <keyword>` or `bd show <suspected-id>`.

## Hand off

**Every bead ends in a tracker status, and that status is the hand-off.** A
completed bead is `closed`, a blocked one `blocked`, a skipped one closed with
its rationale in the notes. Leave none of them `in_progress` — an agent that
walks away from a claimed bead is indistinguishable from one still working,
which is the single reading the orchestrator cannot recover from.

Then return **one message**: one line per bead, then

```
batch-end: <n> completed, <n> blocked, <n> skipped @ <branch>
```

**Every bead your queue returned appears in exactly one of those counts.** A bead
you never reached is `blocked` with the reason, not omitted. **You hand the
branch back; you do not merge it and you do not push it.**

### How the orchestrator reads you

Not from your return line — from your queue. **You have DRAINED when no open,
ready bead in `{EPIC}` is assigned to you and you hold nothing `in_progress`.**
Both halves are required: `bd ready` excludes `in_progress`, so a bead you are
working right now empties the ready front, and the ready query alone would
report you finished mid-bead. **Scope the query to `{EPIC}`** — `--assignee`
alone spans the whole tracker, so work outside this run would read as yours.

| your queue | your worktree | reading |
|---|---|---|
| not drained | present | live worker |
| drained | present | finished; orphaned tree, the orchestrator removes it |
| not drained | absent | the worker died |
| drained | absent | done and cleaned up |

Nothing in that is yours to write, which is why nothing in it can go stale.
**Your obligation is to leave every bead in a truthful status** — claimed while
you work it, closed or `blocked` when you stop. A worktree is matched to you
through its checked-out branch, which is your dispatch id.

## Constraints

- Do not ask the user questions. Decisions are settled before you start —
  including the four that change what "correct" means, which are recorded as
  findings, never escalated.
- Do not modify code outside the current bead's scope.
- **Every bead is its own commit.** Never combine two.
- A bead claimed by a **different** assignee belongs to a live sibling — leave
  it alone, exactly like foreign working-tree changes. Re-claiming your own is
  safe and needs no rule.
- Missed a required step after committing? Add a NEW follow-up commit. Never
  amend a commit that may already be pushed. (Amending a local commit you are
  certain is unpushed, to fold in a missed step, is fine.)

## Acceptance criteria

Done when all of these hold for every bead your queue returned — the same signals an
orchestrator verifies with `bb-status.sh bead <id>`:

- [ ] The bead reads `closed` in the tracker, or `blocked` with a reason.
- [ ] Exactly ONE commit exists for it, subject `<id> P<priority> <title>`,
      carrying both the code and the closure.
- [ ] The working tree is clean and no `.beads/*.jsonl` export was staged.
- [ ] At least one `retro:` comment exists, or the return line states the bead
      was clean.
- [ ] Gates passed.
- [ ] Anything found beyond scope was filed as a bead, not implemented and not
      dropped.
- [ ] Every bead your queue returned appears in exactly one count of the batch-end
      line.

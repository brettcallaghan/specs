---
name: run-bead
description: Take one bead from claimed to committed — the ordered per-bead
  ceremony, with its gate step, findings capture and close-and-commit script.
  Use when working a single bead end to end, whether dispatched as a
  bead-bonanza subagent or working solo with no orchestrator at all. Covers
  claiming atomically, staying in scope while filing what you find, running
  the project's gates, recording findings as retro-prefixed comments, and
  closing so the bead's closure and its code land in one commit. Also carries
  the blocked and skipped paths, which are where a bead most often goes wrong.
version: "0.1"
user-invocable: true
---

# run-bead

Take **one** bead from claimed to committed. The order is the substance: the
close is folded into the implementation commit, so a step taken out of turn
produces a bead that is closed but uncommitted, or committed but unclosed.

Nothing here needs an orchestrator. Dispatched as a subagent you are handed the
inputs below; working solo you resolve them yourself, and the ceremony is
identical.

## Inputs

| | Dispatched | Solo |
|---|---|---|
| `{ID}`, `{PRIORITY}`, `{TITLE}` | given | `bd ready`, then `bd show <id>` |
| `{WORKTREE}` | given — work THERE, it is yours alone | your checkout, on a branch that is not the default |
| checks | run `{SCRIPTS}/bb-checks.py` | same |
| `{CHECKS_IDS}` | given — the ids the project declared | `scripts/bb-checks.py --ids` |
| `{SCRIPTS}` | given | `bead-bonanza/scripts/` |
| `{EPIC}` | given — the run epic, for incidents | `none` |
| `{TRAILERS_FILE}` | given, or `none` | `none` |
| `{CHECKS_BASELINE_TEXT}` | given — what the checks printed before the run | run `scripts/bb-checks.py` on a clean tree, before you start |
| push mode | `push_per_bead` or `defer_push` | push your branch when the work is done |

**Never work in the main checkout when a worktree was given**, and never work
on the default branch.

## Workflow — follow in order

Progress is reported by the tracker, not by you: `bd` emits an event on every
status change, so claiming and closing the bead is what an orchestrator sees.
Write no progress file.

1. **Claim** — `bd update {ID} --claim`. Atomic, and idempotent when the claim
   is already yours. It sets status **and** assignee in one write, where
   `--status in_progress` sets only the status and leaves the bead claimed by
   nobody. The assignee comes from the actor (`$BEADS_ACTOR`, else git
   `user.name`), so exporting your own `BEADS_ACTOR` makes you identifiable in
   the tracker as the agent holding this bead.

2. **Read it in full** — `bd show {ID}`.

3. **Implement, strictly in scope.**

   **If this bead fixes a defect or regression, first write (or locate) a test
   that reproduces it and watch it FAIL for the stated reason BEFORE you fix
   it.** A test written after the fix encodes the fix's own assumptions and can
   stay green while the defect is live. Record the observed failure — it goes
   in a `retro:` comment at step 6.

   **Work you find that is not this bead's** is filed, not done and not
   dropped: `bd create` with a `discovered-from` link to `{ID}`, and one line
   saying whether it *blocks* this bead (a prerequisite, an
   acceptance-criteria decomposition, an expected test-coverage gap) or is
   merely adjacent. Do not implement it. This includes a gap nobody owns — a
   document no bead claims, a step every adjacent bead assumed another would
   take. **"Not this bead's to close" is a reason to FILE it, not a reason to
   leave it unfiled.** If the bead genuinely cannot be finished without it,
   take the Blocked path naming the new bead as the blocker; otherwise carry
   on with the original and do not widen scope inline.

   For any implicit decision you have to make, record what you chose and why —
   that also goes in a `retro:` comment. **When the decision is one of the four
   below, it is an incident** and gets recorded differently — see "Incidents".

4. **Checks — the push is the gate.** Run `{SCRIPTS}/bb-checks.py` here,
   before any push. It runs everything the project declared, in declaration
   order, and keeps going after a failure so you see all of them at once.

   **While iterating, narrow it:** `bb-checks.py <id>...` runs only those, and
   `{CHECKS_IDS}` says which exist. Pick by what you touched — the ids are
   slugs, so `test-go` covers Go tests. **Before your close, run it with NO
   arguments**: a narrowed run is a fast loop, never the gate.

   On failure, fix the root cause and rerun. **After two full failed cycles,
   take the Blocked path** — do not keep retrying.

   **A check that was already failing before you started is not yours to fix.**
   Your prompt carries a `Checks baseline:` block — what the checks printed on
   the untouched tree, and the status they exited with. The bar is **"no
   findings beyond that", not "exit 0"**.

   Compare, do not parse: run the checks, and read your output against the
   baseline's. A line in both was already there; a line only in yours is yours.

   A baseline that exited 0 means every check is expected to pass, so any
   failure is yours. A **missing** block is different — no baseline was taken,
   so you cannot tell a pre-existing failure from one you caused. Say so rather
   than guessing.

5. **WHEN the close must do something differently, record it on the bead
   first** — `bd comment {ID} "caveat: <what and why>"`. A regen step that must
   run before staging; a fixture this bead regenerated. **Only when there is
   one.** It goes on the bead rather than in a file because the tracker is what
   a recovering agent already reads, and it outlives the container.

6. **Record findings as `retro:` comments on {ID}** — one `bd comment` per
   finding. That channel is what later retrospectives read.

   ```
   retro:<kind>[:<target>] <one line: what happened>
   evidence: <quote | path:line | command + output | expected vs observed>
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

   **Evidence in full; analysis not at all.** No length limit — the quote, the
   `path:line`, the command and its output do not survive this session, and
   whoever reads your finding reads it cold. Root cause, proposal and priority
   are derivable later by an agent with the whole run in view, so leave them
   out. Include what you **tried that did not work** (the diff shows only the
   reading that survived) and what you **decided silently** — choices a
   diff-reader would not recognise AS choices.

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
   {SCRIPTS}/bb-close-bead.sh --trailers-file {TRAILERS_FILE} {ID} {PRIORITY} "{TITLE}"
   ```

   Omit `--trailers-file` when it resolved to `none`. **Pass no paths** —
   everything modified in the worktree is this bead's work, and a path list can
   only be too narrow. That holds inside a batch's shared tree too: the previous
   bead committed before this one started, so nothing uncommitted is anyone
   else's.

   It performs the ceremony in the required order: atomic close
   (`bd update --status closed`, never `bd close`, which would be a second
   commit), verify the close against the DATABASE before staging, stage code
   only (`git add -A -- ':!.beads/*.jsonl'`), one commit subject
   `{ID} P{PRIORITY} {TITLE}` with the trailers in its body, then `bd dolt
   push` — the push, not a file, is what makes the closure durable. It refuses
   to run on the default branch or with nothing staged. Do not `git commit
   --amend` afterwards: a missing trailer means the trailers file was resolved
   wrong, which is not something to patch here.

8. **Push** — `push_per_bead`: `git push origin <branch>` on the current branch,
   never the default one. `defer_push`: skip it; the orchestrator pushes.

9. **Return one line**, in this shape:

   ```
   {ID}: completed - <one clause on what changed> @ <short-sha>
   ```

## Blocked path

Gates failed twice, the change is intractable, or a prerequisite you filed
stands in the way:

- `bd update {ID} --status blocked --append-notes "<short reason>"` — name the
  blocking bead's id when there is one.
- **Leave your uncommitted changes exactly where they are.** Do NOT stash and
  do NOT commit them. The worktree holds the partial work and its continued
  existence is what tells an orchestrator this bead is unfinished; stashing
  moves the work out of it and destroys that signal.
- Return the blocked line. Do not push.

## Skipped path

The bead is already addressed, malformed, or duplicates a closed bead:

1. **Confirm your tree is clean — a hard precondition, not a courtesy.** Run
   `{SCRIPTS}/bb-verify-skip.sh` before anything else; it must exit 0. If it
   exits 1 it prints what is dirty and the two ways out: finish the close as
   `completed` (the work is this bead's), or revert the changes and re-run
   until CLEAN. Do not proceed while it exits 1.
2. `bd update {ID} --status closed --append-notes "duplicate of <id>"`, or a
   similar one-line rationale.
3. Return the skipped line **with a citation**:
   `{ID}: skipped - already addressed by <bead-id> @ <short-sha>`. "Already
   done" without one is not acceptable — nobody can verify it. Find it with
   `git log --oneline --all | grep <keyword>` or `bd show <suspected-id>`.

## Incidents — decisions that change what "correct" means

Most decisions are about *how* to do the work, and a `retro:` comment is the
whole of their record. Four kinds change *what correct is*, and guessing at
those produces confidently wrong work that passes every check:

| | | kind |
|---|---|---|
| an **acceptance criterion needs amending** | the bead's definition of done is wrong, so any implementation is a guess at intent | `decision` |
| a standard, requirement and AC are in **tension** | there is no conservative branch — each choice violates something | `decision`, one per target |
| a **contract or interface** change | the blast radius is outside this bead's scope by construction | `scope` |
| a spec is **silent** where the bead requires it to speak | your choice writes an undocumented decision into the code | `gap` |

**Decide anyway. Do not stop and do not ask** — see Constraints. A recorded
decision gets reviewed at the pull request and revised in a later bead, and
that is cheaper than a run blocked on a question nobody is awake to answer.

**Record it two ways, both required:**

```bash
# <kind> from the table above; <target> is the document or bead it is about.
bd comment {ID} "retro:<kind>:<target> incident — <what you decided, and why>.
Alternative considered: <what you rejected, and what it would have cost>."

# and, WHEN {EPIC} is not `none`, so the run's record carries it
# independently of this bead
bd comment {EPIC} "amend: incident on {ID} — <one line>"
```

**An incident is a severity, not a kind** — it carries the ordinary
`retro:<kind>:<target>` prefix so routing still picks it up, and the word
` incident ` after the prefix is what makes the class greppable. A comment
spelled `retro: incident` with no kind routes nowhere and is invisible to the
report's sweep. The run's report names every one of them **before anything
else** — that section is what earns an autonomous run its autonomy.

**Working solo (`{EPIC}` is `none`)?** The bead comment alone. Say in your
return line that an incident was recorded, so whoever reads it knows to look.

An incident is not a failure and does not block the bead. **The incident count
is a quality metric for the specs, not for you** — each one marks a place the
corpus failed to answer a question, which is a finding against a document
rather than against the work.

## Constraints

- Do not ask the user questions. Decisions are settled before you start —
  including the four in "Incidents", which are recorded, never escalated.
- Do not modify code outside this bead's scope.
- A bead claimed by a **different** assignee belongs to a live sibling — leave
  it alone, exactly like foreign working-tree changes. Re-claiming your own is
  safe and needs no rule.
- Missed a required step after committing? Add a NEW follow-up commit. Never
  amend a commit that may already be pushed. (Amending a local commit you are
  certain is unpushed, to fold in a missed step, is fine.)
- Only push the current branch, never the default one.

## Acceptance criteria

This ceremony is done when all of these hold — they are the same signals an
orchestrator verifies with `bb-status.sh bead <id>`:

- [ ] The bead reads `closed` in the tracker.
- [ ] Exactly ONE commit exists for it, subject `{ID} P{PRIORITY} {TITLE}`,
      carrying both the code and the closure.
- [ ] The working tree is clean and no `.beads/*.jsonl` export was staged.
- [ ] At least one `retro:` comment exists, or the return line states the bead
      was clean.
- [ ] Gates passed, or the bead is `blocked` with a reason.
- [ ] Anything found beyond scope was filed as a bead, not implemented and not
      dropped.
- [ ] Any decision that changed what "correct" means was recorded as an
      incident on the bead, and on the run epic when there is one.

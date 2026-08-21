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
| checks | run the project's `checks` script | same |
| `{SCRIPTS}` | given | `bead-bonanza/scripts/` |
| `{TRAILERS_FILE}` | given, or `none` | `none` |
| `{CHECKS_BASELINE}` | given — what the checks printed before the run | run `checks` on a clean tree, before you start |
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
   that also goes in a `retro:` comment.

4. **Checks — the push is the gate.** Run the project's `checks` script here,
   before any push. It runs everything the project wants run, in its own order,
   so there is nothing to select or sequence.

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
   finding, prefixed `retro:`, 1–3 sentences, friction only, no root-causing.
   That channel is what later retrospectives read. Record one WHEN: you needed
   information the bead and its references did not carry; you had to settle
   something the bead or spec should have pinned; the bead contradicted repo
   reality; you could not satisfy an acceptance criterion as declared; a
   sequencing surprise. Include what you **tried that did not work** — the diff
   shows only the reading that survived — and what you **decided silently**,
   meaning choices a diff-reader would not recognise AS choices.

   **Work, not friction.** If what you found is a unit of work someone must do,
   it is a BEAD (step 3), not a retro comment, even when it also reads as
   friction. A retro comment is read after the run; a bead is read during it.

   **A clean bead records none — say so rather than inventing filler.**

7. **Close and commit in one step** — run the script, do not perform it:

   ```sh
   {SCRIPTS}/bb-close-bead.sh --trailers-file {TRAILERS_FILE} {ID} {PRIORITY} "{TITLE}"
   ```

   Omit `--trailers-file` when it resolved to `none`. **Pass no paths** — in
   this bead's own worktree everything modified is this bead's work, and a path
   list can only be too narrow.

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

## Constraints

- Do not ask the user questions. Decisions are settled before you start.
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

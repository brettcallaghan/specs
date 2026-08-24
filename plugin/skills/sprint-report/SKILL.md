---
name: sprint-report
description: Write a sprint's final report — the incidents the run decided, the
  per-bead results, routing evidence, follow-up beads, and what the run could
  not verify. Use when a sprint has drained its queue, tripped its circuit
  breaker, or otherwise reached its end and needs the human-facing account of
  what happened. Takes only the run epic id; everything it needs is on the epic
  or reachable from it, so a session that did not run the sprint can still
  write its report.
version: "0.1"
user-invocable: false
---

# sprint-report

The run's own output. **Its only input is the run epic id** — everything the
report needs is on that epic or reachable from it, which is what lets a session
that did not run the sprint still write one.

**Before your first `bd` call: `export BEADS_ACTOR=report-<run-id>`**, where
`<run-id>` is that epic id with the project prefix dropped. You write the
report's path back onto the manifest, so you write; unset, that write lands
under the harness default and `bd ready --assignee report-<run-id>` stops
answering for your phase. Take the id the orchestrator handed you rather than
minting one — the roster and the derivation rule are the coordinator's dispatch
policy's, their only copy.

## What you produce

One human-facing report, written to a file, with its path recorded in the
manifest so the report outlives the session that produced it:

```sh
bd update <epic> --set-metadata report_path=<path>
```

**A report is always produced**, including for a run that stopped early, was
interrupted, or tripped its circuit breaker. **A run that ends without one has
failed**, whatever else it achieved.

## Incidents come first — before the results table

Every decision this run made that changed what "correct" means: an acceptance
criterion amended, a standard/requirement/AC tension resolved, a contract or
interface changed, a spec silent where a bead needed it to speak.

An incident is a severity, not a kind: it is an ordinary
`retro:<kind>:<target>` finding whose text carries the word ` incident ` after
the prefix. Match on that, not on a `retro: incident` prefix — nothing writes
one.

```bash
bd export | jq -r '.comments[]?.text' | grep -E '^(retro:[a-z]+:[^ ]+ incident|amend: incident)'
```

One row each — the bead, what was decided, what was rejected and at what cost.

**State `none` explicitly** when there were none. The null is information; a
missing section reads like a clean run rather than an unreported one.

This section is first because it is the one a human skims, and **it is what
earns an autonomous run its autonomy**. The count is a quality metric for the
specs, not for the run: each incident marks a place the corpus failed to answer
a question, which makes it the retro's material rather than a complaint.

### Then, still before the results: what the run is waiting on a human for

An incident is decided. A **deferred review** is not — it is a decision the run
took and asked a human to look at before the run is handed over, filed as a
`human`-labelled bead under a late ceremony epic. Those are open questions
addressed to the reader of this very report, so they go above the results:

```bash
bd list --label human --status open --json
```

**Not `bd human list`, and this is the trap.** That command keeps listing a
bead after it has been closed — measured on bd 1.2.2: with `bd show` reporting
`status=closed`, `bd human list` still printed `Human-needed beads (1 found)`.
A report built on it says a human owes an answer that was given days ago, and
it never stops saying so. Filter by status, always.

One row each — the bead id, the question, and the phase epic it is holding
open. **State `none` explicitly** when there are none, for the same reason the
incidents section does. The three human channels, and why this one blocks only
at the end, are the coordinator's run-record reference's.

## Then, in order

1. A markdown table: `| ID | Status | Batch | Commit | Notes |`
2. Totals: `N closed, N blocked, N skipped`.
3. **Implicit decisions made** — the rest of the beads' `retro:` comments:
   which beads had implicit decisions, what the subagent chose. Surface to the
   user so they can audit non-obvious choices. Incidents are already in 0;
   do not repeat them here.
3a. **Tier escalations** — every bead whose tier you had to raise to get it
    done, one row each:

   ```
   | Bead | Assigned | Completed at | What failed at the assigned tier |
   |------|----------|--------------|----------------------------------|
   | demo-9h0 | light | standard | fixture regen wrong twice; gates failed both cycles |
   ```

   State `none` explicitly when there were no escalations — the null is
   the result, not the absence of one.

   **`CLAUDE_CODE_SUBAGENT_MODEL` override and no routing are different
   results.** WHEN `CLAUDE_CODE_SUBAGENT_MODEL` env var is set, every
   subagent runs at that forced model regardless of tier, so no routing
   occurs and tier escalations mean nothing — omit the table and write:

   ```
   Tier escalations: N/A — CLAUDE_CODE_SUBAGENT_MODEL is set, so every
   bead ran at that model and no tier routing occurred.
   ```


   **Say this alongside the table**, because the reader will otherwise
   draw the wrong conclusion from a short one: escalations reveal
   *under*-routing only. A bead groomed `frontier` that a `light` model
   would have finished never appears here — it succeeds, and the
   overspend is silent. So zero escalations across a large queue is not
   evidence of good calibration; it is equally consistent with routing
   everything too high (see the over-routing table below). WHEN a run of
   ≥10 beads produces zero escalations, say so and suggest the next run
   try one tier lower on the beads whose `tier_reason` was thinnest.
3b. **Deliberate over-routes** — every bead you knowingly ran a tier above
    its assigned one (e.g. batched into a higher-tier batch per the
    tier-batching rule), one row each:

   ```
   | Bead | Assigned | Ran at | Why raised |
   |------|----------|--------|------------|
   | demo-9h0 | light | standard | batched with a standard bead on the same surface |
   ```

   State `none` explicitly when there were no over-routes — the null is
   the result, not the absence of one. This table and the escalation table
   above are together the run's complete routing evidence, and
   `groom-beads`' retro loop needs both to recalibrate the rubric.
4. **Follow-up beads** — beads created mid-run via `bd create`, split into those
   admitted into the run (in-scope prerequisites / AC-decompositions / coverage
   gaps, with their outcome) and those left for the user (out-of-scope finds,
   filed with a `discovered-from` link, awaiting triage).
5. **Stashes** — if any were created during recovery, list with `git stash
   list`.
5a. **The retro's outcome** — read from the manifest, where `sprint-retro`
    records it. Not yours to run and not yours to sequence; name what is there:

    ```
    Retro: <retro_filed> filed, <retro_declined> declined,
           <retro_outbox> routed to .sf/outbox/
    Every finding, verbatim: <findings_path>
    ```

    **Cite `findings_path`, never restate the findings.** The complete set is
    what recurrence accumulates over, and a second copy that drifts is worse
    than a link. **WHEN the keys are absent**, say `Retro: not run` rather than
    omitting the line — a retro that was skipped and one that found nothing are
    different results, and only the first is a defect. Absent is the honest
    reading whatever the reason, so this needs no knowledge of what ran when.
6. **Pushes** — confirm all closed beads' commits are on the remote.
6a. **The durable out-of-session trigger's disposition** —
    `Durable trigger: cancelled (<id>)` or `Durable trigger: none armed`. Both
    branches are stateable, and an omitted line is indistinguishable from a
    forgotten cancellation. Cancelling it is post-flight's; saying which
    happened is yours.
7. The run epic id, so the user can read the manifest, the amendments and
   every bead the run touched.
8. **What this run could not verify** — named, not omitted. A check that never
   ran, a push whose landing was not confirmed, a bead accepted on four of five
   signals. An unstated gap is indistinguishable from one that does not exist.

Do NOT create a PR or merge to the default branch — that is the human's.

## Scope: exclusions and what was left

Two things a report must never let a reader misread as completeness:

- **Beads excluded at admission** — the manifest's `excluded` list, with each
  reason. A run that finished its admitted set while three beads sat out is
  done over a *narrowed* scope, and must say so rather than reading as complete
  over the scope the user first described.
- **Findings left at loop exhaustion** — filed as beads, named here. A review
  budget spent is not a failure; a budget spent silently is.

And **which reviewer ran**, from `sprint-review`'s `reviewer_source`:

| value | what the report says |
|---|---|
| `declared` | reviewed by the reviewer `sf.toml` declares — name it |
| `discovered` | **no reviewer declared**; reviewed by a skill found in the session — name it |
| `none` | **no reviewer declared and none available**; reviewed by a plain agent with no skill |

The last two are weaker evidence than the first and the report SHALL say so.
Collapsing them lets a bare-agent review read exactly like the project's own,
which is the same failure a silently skipped review used to produce.

## What you do not do

You do not re-derive the run. Read the tracker, the manifest, the amendments
and `git log` on the sprint branch — do not re-read diffs, re-run checks, or
re-open the beads' work. A report that costs a second pass over the run is a
report that gets skipped.

You do not add a conclusion the evidence does not carry. Where a question has
no evidence, the answer is `null result`, in writing.

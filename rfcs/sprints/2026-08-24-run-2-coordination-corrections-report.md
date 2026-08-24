# Sprint report — Run 2: make the coordination plan trustworthy, and fix what run 1 shipped

| Field | Value |
|-------|-------|
| Date | 2026-08-24 |
| Run epic | `agent-plugins-so39` |
| Branch | `claude/bead-bonanza-design-review-ne1plu` (PR #27) |
| Base sha | `5cc1ce9` |
| Head at wave close | `99ee4a4` (ten commits) |
| Branch head now | `c48e995` — the retro's findings document, committed after the wave |
| Admitted | 11 beads |
| Excluded at admission | 22 beads — **the run's scope was narrowed by design** (§9) |
| Outcome | **10 closed, 0 blocked, 1 bounced (never dispatched)** |
| Waves | 1 |
| Review | 1 pass of a 2-loop budget; 10 findings filed, **0 admitted** |
| Retro | 28 findings harvested, 14 candidates, **0 filed / 0 declined / 0 outbox — awaiting the human** |
| Findings, verbatim | `rfcs/sprints/2026-08-24-run-2-coordination-corrections-findings.md` |

**The run's thesis, and its own verdict on itself.** This run existed to stop
*documents asserting unmeasured tool behaviour* from propagating errors. It
corrected three such documents, and the corrections are right. It also
committed the same error three times itself, at three different tiers: inside
the corrected paragraph (`--set-metadata`), in a coordinator dispatch prompt
(`sync_dir`), and in our own script (`agent-plugins-b9ft`, the exact `pipefail`
trap `AGENTS.md` documents). A fourth — "`bd search` is non-functional" — was
carried into two epic comments, the retro's dispatch brief and the recovery
trigger's own prompt before the retro measured it false.

The corrections landed. The habit that produced the originals was not fixed by
correcting instances of it. Read §1 before §2.

---

## 1 · Incidents

Six. **Four of them are not reachable by this skill's own incident query**
(`^(retro:[a-z]+:[^ ]+ incident|amend: incident)`) — one is a bare `incident:`
comment, and three sit behind `retro:proof:` prefixes whose ` incident ` marker
falls mid-text. Every comment on the epic was read directly rather than
grepped. The four unreachable ones are the most consequential records the run
produced, and they lead.

### 1.1 — Review pass 1 filed 10 findings and admitted zero; three are defects in the corrections this run shipped

*Bare `incident:` comment on the epic — not matched by the query.*

Ten findings were filed as `agent-plugins-so39.2` … `.11`, all **deferred**,
all unadmitted. `standing_auth.admit_review_findings` is
`{value: false, source: "default"}` — **nobody chose it**, and the skill's
in-scope test requires gate authorization for scope growth, so severity cannot
override it. The review loop therefore closed at pass 1 inside a 2-loop budget.
Recorded as an incident rather than escalated, per the operating principle.

Three of the ten — `so39.4`, `so39.5`, `so39.6` — are defects **in the text this
run corrected**, on the exact thesis the run was opened to serve.

**What was decided:** merge with the findings deferred, and record the fact
loudly. **What was rejected, and at what cost:** admitting them mid-run. The
cost is that `plugin/skills/groom-beads/references/capability-tier.md` ships
today carrying a new false claim in the same paragraph that fixed an old one.

This is the *second consecutive run* where an unadmitted review finding was a
defect in the run's own output; run 1's was the disproved "P0". The mechanism
is unchanged, and it is a default nobody set.

### 1.2 — The coordinator made three false tool claims, each verified false

*`retro:proof:coordinator incident` — marker mid-text, not matched by the
query.*

| # | Claim | Measured reality |
|---|---|---|
| 1 | `--set-metadata` stores every value as a string, arrays included | **False.** On bd 1.2.2 JSON scalars keep their type: `num=5` → number, `b=true` → boolean, `f=1.5` → number, `n=null` → null. Only arrays, objects and non-numeric text stay strings. |
| 2 | `sync_dir` does `rm -rf "$dest"` then `cp -r`, in `sync-shared-references.sh` | **False attribution.** `sync_dir` lives in `scripts/generate-schemas.sh` (lines 183–201) and belongs to strata artifacts. `sync-shared-references.sh`'s writer is `render()`, which ends `cat "$1"`. |
| 3 | `bd search` is non-functional on this build — zero hits even for common terms | **False.** It excludes closed issues by default, as its own `--help` line 1 says. `bd search metadata` → 2 hits; `bd search metadata --status all` → 6, four of them closed. |

**Claim 1** was carried into at least six dispatch prompts, one bead
description and two epic comments, and it **landed inside
`agent-plugins-4tjk`'s corrected text** — filed back as `so39.5`, and confirmed
by the coordinator's own independent measurement rather than accepted on the
reviewer's word.

**Claim 2** was caught by the bead it was dispatched into. The correction is
not cosmetic: whole-file-by-`cat` is the **mechanical basis** of the
vendor-or-prose rule `agent-plugins-uva5` shipped, so the rule now rests on
measured behaviour instead of a misattribution. The detail came from
`sprint-plan`'s ledger notes about `generate-schemas.sh` and was transposed onto
the wrong script when the prompt was written.

**Claim 3** is the sharpest, because it disabled its own fix. The coordinator
filed `agent-plugins-vcka` and `agent-plugins-xpn8` as duplicates 27 minutes
apart; the recorded remedy was "run `bd search` before `bd create`" — and that
remedy was then declared impossible. `bd search strata --status all` returns
`agent-plugins-vcka` as its **second hit**. The duplicate was preventable by
the exact step that was declared unavailable.

**What was rejected:** nothing was rejected here — each claim was inherited or
relayed rather than measured at the point of use, which is the precise failure
`agent-plugins-4tjk` was opened to fix. The pattern is not in the documents. It
is in the coordinator's handling of them, and correcting instances has not
touched it.

### 1.3 — Dispatch bypassed `sprint-wave`, deliberately, with the reason recorded

*`retro:process:coordinator-dispatched-directly incident` — reachable by the
query, but `process` is not one of the nine kinds, so the retro could not route
it.*

The coordinator dispatched wave 1 itself rather than invoking `sprint-wave`.
Run 1 measured that `sprint-wave`, dispatched as a subagent, is given no
Task/Agent tool and **silently absorbs the batch it was meant to delegate** —
performing as self-verification the very check it exists to perform against an
independent agent. Invoking it again would have reproduced a known failure.

**What was decided:** direct dispatch. **What it costs:** worktree and branch
names come from the harness rather than the dispatch id, so the harness-id to
dispatch-id mapping is the coordinator's to record; and the coordinator holds
the return-verification reasoning it is designed *not* to hold. Both are
already filed against the skills
(`.sf/outbox/bead-bonanza/2026-08-24-no-dispatch-tool-is-a-hard-blocker.md`),
and plan item 8 makes direct dispatch the only path by deleting `sprint-wave`
outright.

### 1.4 — `agent-plugins-b9ft`: this run's own validation depends on a flaky check

*`retro:tooling:scripts incident` — reachable by the query.*

`scripts/sync-shared-references.sh`'s orphan guard tests membership with
`printf … | grep -qxF` under `set -o pipefail`. `grep -q` closes the pipe on
first match, `printf` dies on SIGPIPE, and `pipefail` turns a **successful
match into a failed pipeline** — so a correctly-mapped file is reported ORPHAN.
This is verbatim the trap `AGENTS.md` documents under "Pipes hide exit
statuses".

Measured: **2 failures in 30 runs** on an unchanged tree, naming a *different*
correctly-mapped file each time. Negative control with exactly one line changed
to a herestring: **100 runs, 0 failures.**

**What was decided:** filed P1, **not admitted mid-run** — it blocks no
remaining bead, batch 2 was already in flight, and the admitted set is fixed.
**What that costs:** the epic's `checks_baseline` records `shared-refs` as
*passed*, so the flake charges itself to whichever bead happens to hit it and
reads as that bead's regression. **Any red `shared-refs` seen in or after this
run must be re-run before it is believed.**

### 1.5 — The coordinator stalled at a ceremony boundary; the watchdog would have caught it

*`retro:proof:progress-streaming` and `retro:tooling:bead-bonanza incident` —
the first is not matched by the query.*

The stall was real: the coordinator ended a turn saying "moving to the retro"
with nothing dispatched and nothing in flight, and the run sat still until the
human asked. A run whose progress depends on a human noticing that nothing is
happening has no liveness property at its ceremony boundaries.

**The correction, in the coordinator's favour and recorded accurately:** the
claim that *no mechanical signal could have caught it* is **false**. The
durable recovery trigger armed at first dispatch fired at 17:57:32 carrying
exactly the right instruction — "if that run already finished, cancel this and
do nothing; otherwise invoke `sprint-recovery`" — and the run had not finished,
so it would have resumed the ceremony. The human simply asked a few minutes
earlier.

**What this means:** the watchdog design works and needs no new mechanism. What
it cannot do is close the gap between a stall and the next firing, which was set
at 50 minutes for this run. The finding worth keeping is about the **interval**
and about the coordinator, not about a missing signal.

### 1.6 — `agent-plugins-bbtl`: satisfying "Files clauses only" required adding sections

*`amend: incident` from `dev-so39-2` — reachable by the query.*

Rollout items 12 and 13 had **no `**Files.**` clause at all**, so satisfying
"every specified value names its writer" meant *adding* sections against a
criterion reading "no item's substance is changed — Files clauses only".

**Decided:** adding a missing Files clause is not a substance change; added
them. **Rejected:** recording items 12 and 13 as "cannot be audited, no Files
clause" and leaving them untouched — that satisfies criterion 1 while failing
criterion 2 on the two items where the omission is most complete, leaving the
exact defect class the bead exists to remove sitting in two of nine.

Related, and routed rather than fixed: item 13's sketch reads `$AGENT_ID` while
item 4 writes `BEADS_ACTOR`, and nothing writes `AGENT_ID`. Item 12's sketch
reads `$RUN_EPIC`; `grep -rn 'RUN_EPIC' plugin/` returns zero hits. Both were
stated in the new Files clauses and routed as comments to the deferred
`agent-plugins-qozp.14` and `agent-plugins-qozp.13`; neither bead's status was
touched.

---

## 2 · Results

| ID | Status | Batch | Commit | Notes |
|---|---|---|---|---|
| `agent-plugins-4tjk` | closed | #3 (`dev-so39-3`) | `8a68d6a` | Corpus said `--metadata` replaces; on bd 1.2.2 it merges at top level. Found and fixed a **4th** site the AC grep missed. |
| `agent-plugins-qozp.23` | closed | #6 (`dev-so39-6`) | `6d92098` | `sprint-batch` now exports `BEADS_ACTOR`. No findings recorded — see §3. |
| `agent-plugins-qozp.28` | closed | #6 (`dev-so39-6`) | `893cc43` | `references/scripts.md` side-effect list gains the `concurrency_cap` write. No findings recorded. |
| `agent-plugins-l1p2` | closed | #6 (`dev-so39-6`) | `582d458` | Stray code fence removed. No findings recorded. |
| `agent-plugins-qozp.19` | closed | #1 (`dev-so39-1`) | `f6d675e` | `bd epic close-eligible` takes no epic argument; fixed at both documented sites. |
| `agent-plugins-d6zz` | closed | #1 (`dev-so39-1`) | `f970cb5` | Ready-front vs `manifest.admitted` diff, demonstrated in a throwaway database. |
| `agent-plugins-qozp.20` | closed | #5 (`dev-so39-5`) | `726ab33` | Epic-DoR assignee gate gains a producer and moves after the assign step. Overran the byte cap by 236 bytes and paid for it by cutting two restatements — see §3. |
| `agent-plugins-uva5` | closed | #4 (`dev-so39-4`) | `9c28e4b` | Vendor-or-prose rule shipped as two **ordered** tests; endorses both of run 1's opposite outcomes. |
| `agent-plugins-gu3i` | closed | #2 (`dev-so39-2`) | `d6ad144` | **Not `cfcdcd6`** — see the note below. `BEADS_ACTOR` claim premise qualified at both sites. |
| `agent-plugins-bbtl` | closed | #2 (`dev-so39-2`) | `99ee4a4` | Rollout items 6–14 audited: 8 changed, 1 audited-clean. Incident §1.6. |
| `agent-plugins-xpn8` | **bounced** | — | — | Never dispatched. Deferred with a `human` label and three resolutions. §2.1. |

**Totals: 10 closed, 0 blocked, 0 skipped, 1 bounced.**

**Commit-sha correction.** The coordinator's hand-off recorded `cfcdcd6` for
`agent-plugins-gu3i`. That commit exists but is **not an ancestor of the branch
head** — it is the pre-rebase commit from the worktree.
`git merge-base --is-ancestor cfcdcd6 HEAD` exits 1. The commit as landed on
`claude/bead-bonanza-design-review-ne1plu` is `d6ad144`. Every other sha in the
hand-off verifies against `git log 5cc1ce9..99ee4a4`.

### 2.1 — The bounce: `agent-plugins-xpn8`

Bounced by `sprint-plan` at the readiness gate, **on a measurably false
premise in the bead's own Problem statement**. The bead said "this adds one
source file and one validator to machinery that already exists". Measured:
`schemas/` is *one* strata package scoped to the Spec CLI
(`packageName: "SpecModel"`, `GO_OUT=cli/core/specmodel`,
`MD_OUT=plugin/skills/spec/references/generated`). A run-manifest type there
emits a Go struct into a package that models specification documents, folds
into the bundled spec schema, and ships markdown inside the **spec** skill's
references. All three are wrong destinations. The work is L-shaped, not S.

Two further gaps: the "generated from" acceptance criterion offers two options
and selects neither, and the one it appears to prefer collides with
`agent-plugins-uva5` — the very rule this run was chartered to write.

**Three resolutions are recorded on the bead** for a human: split into a second
strata package plus the definition; accept the pollution and say so in
`schemas/README.md`; or drop strata and hand-write the JSON Schema. The bead
was set `deferred` deliberately, because it was *blocked by* `agent-plugins-4tjk`
— which this run closed — so it would otherwise have entered the ready front
with nothing flagging it, and it **is** in `manifest.admitted`, so a
ready-front-minus-admitted diff would not have caught it either.

`agent-plugins-vcka` is its superseded duplicate — closed by
`bd supersede --with=agent-plugins-xpn8`.

---

## 3 · Implicit decisions made

Incidents are in §1 and are not repeated.

| Bead | What the subagent chose, without being told to |
|---|---|
| `agent-plugins-4tjk` | Re-derived all three measurements in a throwaway `probe` database before writing the correction, rather than trusting the bead. Then widened the AC's own grep (`metadata.{0,80}replac`) as a sanity check and found a **4th** site the narrow AC pattern could not match — `capability-tier.md:11`, a TOC line. Corrected in the same commit. |
| `agent-plugins-uva5` | **Rejected the bead's own proposed wording** and shipped something else: the bead's two branches *both fire* on the disputed case, reproducing the ambiguity it exists to remove. Shipped instead an ordered pair of tests — (1) do you need the whole file, (2) is the file applied or extended — with test 1 winning ties. Verified the new rule endorses both of run 1's opposite outcomes, so nothing needs migrating. |
| `agent-plugins-gu3i` | Rollout item 4 carried the claim **nowhere in its prose**, so "keep the conclusion at each amended site" had nothing to attach to. Added two sentences to the `**Change.**` paragraph rather than burying the measurement in `Verify`, on the reasoning that a Verify clause is a test procedure, not a takeaway. Also decided silently: left the design doc's row-4 table cell shorter than its risk bullet, since a table cell that long stops being readable. |
| `agent-plugins-bbtl` | **Filed no bead** for rollout item 2 — the instance that motivated the whole bead — which still names no writer. Reasoning: item 2's work is closed, the rollout plan is transient per `CLAUDE.md`, and a wrong Files clause on a completed item has no future reader. Also echoed item 14's `(new)` convention for two paths that do not exist, rather than inventing settled ones for an explicitly unsettled decision. |
| `agent-plugins-qozp.20` | The assign step did not fit in 250 bytes of headroom. Chose to pay by deleting **two restatements, not two rules** — a merge-collision sentence already stated 30 lines earlier in the same file, and a ceremony-slot paragraph whose full rule and implementing `jq` live in `epic-definition-of-ready.md`, now cited by name. Rejected taking the Blocked path, on the grounds that the instruction forbids cutting a *rule* and neither cut removes one. |
| `agent-plugins-qozp.19` | Found the defective command documented at **two** sites with identical wording; fixed both to keep them in agreement. |
| `agent-plugins-d6zz` | Demonstrated the new creep check in a throwaway database across three states — under-admitted, matching, and empty ready front — and confirmed via `bd search` that nothing leaked into this project's tracker. |
| `agent-plugins-qozp.23`, `.28`, `l1p2` | **Recorded nothing.** All three returned with zero `retro:` comments; `bb-status.sh bead` flagged signal 5 "findings: none recorded", exit 1, on each. See the note below. |

**The light-tier silence is itself a finding.** The dispatch prompt asked for
findings explicitly. Two readings, and they are not equally likely: three
one-line edits genuinely producing no friction, or a light-tier agent dropping
a step that is last in the ceremony and whose output nobody waits for. The
other three batches all recorded findings. This is the same shape as run 1's
"near-zero `discovered-from` is suspicious" — an absent signal read as a clean
one — and it is worse at the light tier, because that is where the cheapest
agents run the most beads.

### 3a · Tier escalations

**None.** No bead's tier was raised to get it done; every bead completed at the
tier `sprint-plan` assigned. `CLAUDE_CODE_SUBAGENT_MODEL` is **unset**, so real
tier routing occurred and this null is a result rather than an artefact.

**Read this correctly.** Escalations reveal *under*-routing only. A bead
groomed `advanced` that a `light` model would have finished never appears here
— it succeeds, and the overspend is silent. **Zero escalations across 10 beads
is not evidence of good calibration**; it is equally consistent with routing
everything too high. The next run should try one tier lower on the beads whose
`tier_reason` was thinnest — on this queue that is `agent-plugins-4tjk`
(standard: "the measurement is already recorded on the bead and the three sites
are enumerated") and `agent-plugins-gu3i` (standard: "no new measurement and no
design judgement required"), both of which read like light-tier work.

Note that `agent-plugins-4tjk` found a fourth site its own AC could not match,
which is judgement a light model might well have missed — so the thin
`tier_reason` was not necessarily a wrong call, only an untested one.

### 3b · Deliberate over-routes

One.

| Bead | Assigned | Ran at | Why raised |
|---|---|---|---|
| `agent-plugins-gu3i` | standard | advanced (`dev-so39-2`, opus) | Batched with `agent-plugins-bbtl` (advanced) on disjoint hunks of the same file, `docs/plans/beads-coordination-rollout.md`, per the tier-batching rule. |

Every other bead ran at its assigned tier: batch #1 (`qozp.19`, `d6zz`) both
standard on sonnet; #3 (`4tjk`) standard on sonnet; #4 (`uva5`) and #5
(`qozp.20`) both advanced on opus; #6 (`qozp.23`, `qozp.28`, `l1p2`) all three
light on haiku.

Together §3a and §3b are the run's complete routing evidence.

---

## 4 · Follow-up beads

### Created mid-run and admitted into it

**None.** The admitted set was fixed at 11 at admission and never grew — which
is the same fact that §1.1 records as an incident, since it is also why ten
review findings and one P1 flake stayed out.

### Created mid-run and left for the user

| ID | P | Status | What it is |
|---|---|---|---|
| `agent-plugins-b9ft` | P1 | open | `sync-shared-references.sh --check` flakes: `pipefail` + `grep -q` misreports a random mapped file as ORPHAN. **The first follow-up** — §1.4. `discovered-from` `agent-plugins-uva5`. |
| `agent-plugins-b1gl` | P2 | open | `run-record.md`'s wave-table example still numbers batches per wave, contradicting the run-wide `<batch-id>` that `qozp.20` shipped. `discovered-from` `agent-plugins-qozp.20`. Adjacent, not blocking. |

### Not created, but requiring a disposition

`agent-plugins-qozp.22` (P1, **deferred**) is **stale, not merely deferred.**
It reports the same defect `agent-plugins-qozp.20` fixed, but proposes the
opposite direction — `dev-<run>-<wave>-<batch>` where `qozp.20` shipped
run-wide numbering. The two derivations are mutually exclusive. **Reviving it as
written would re-break what this run shipped.** It must be closed as superseded
or rewritten before any future run admits it; it is not safe to pick up from
the backlog as-is. It was correctly left untouched under the do-not-undefer
constraint, and it sits under epic `agent-plugins-qozp`, outside this run.

### Filed by the review, all deferred

`agent-plugins-so39.2` … `.11` — ten findings, **zero admitted** (§1.1).

| ID | P | One line |
|---|---|---|
| `so39.2` | 2 | `bb-status.sh` creep diff has no no-manifest guard — an epic without a manifest reports its whole ready front as creep |
| `so39.3` | 2 | The wave-boundary paragraph sends the coordinator to `bb-status.sh run <epic>`, which exits 2 whenever an admitted bead is still open |
| `so39.4` | 2 | **Defect in this run's own correction.** The corrected merge semantics never mention `--unset-metadata`, and `run-record.md`'s write-back recipe can no longer delete a key |
| `so39.5` | 2 | **Defect in this run's own correction.** `capability-tier.md` still claims `--set-metadata` "stores every value as a string"; on bd 1.2.2 it JSON-parses scalars |
| `so39.6` | 2 | **Defect in this run's own correction.** The epic-DoR's new "runs AFTER the assign step" note contradicts the file's own opening line and `groom-beads`' unchanged grooming-time invocation |
| `so39.7` | 2 | `sprint-plan`'s new assign step is a durable tracker write absent from the Interface's "Out — two things, and nothing else" |
| `so39.8` | 2 | `bb-status.sh`'s creep query ignores `bd ready`'s return code, so a failed query prints "no creep" |
| `so39.9` | 3 | `capability-tier.md`'s nested-replace transcript does not reproduce as shown |
| `so39.10` | 3 | `groom-beads` phase 5's example still restates the phase-4 tier keys the corrected prose says survive a merge |
| `so39.11` | 3 | `supervision.md` says "close only those explicitly" but names no command; the only command shown is the forbidden one |

**The review's own quality is worth stating**, because a bare count reads as
noise. It **rejected** one `code-review` finding as a false positive rather
than filing it — the claim that `726ab33` deleted the last explanation of the
merge-collision hazard, when the rationale survives in `run-record.md`, cited
two lines above the deletion — and **dropped two more as already filed**. The
ten are what survived verification, not what was reported.

### Retro candidates, not filed

Fourteen. Not beads. **Awaiting a human decision** — see §5a.

---

## 5 · Stashes

**None.** `git stash list` is empty; no recovery occurred.

## 5a · The retro's outcome

```
Retro: 0 filed, 0 declined, 0 routed to .sf/outbox/
       retro_status = awaiting-decision
       28 findings harvested across 8 beads → 14 candidates, 2 unrouted
Every finding, verbatim:
  rfcs/sprints/2026-08-24-run-2-coordination-corrections-findings.md
```

Nothing is filed because `sprint-retro` files only what a human accepts, and
that decision is outstanding. The candidates are ranked in §4 of the findings
document; six are P1. **The two unrouted findings are both the coordinator's**,
and both used `process`, which is not one of the nine kinds — the same drift
run 1 produced from the same role. Neither is lost; neither routed.

The findings document is 78 KB and is the complete record. It is not restated
here: a second copy that drifts is worse than a link.

## 6 · Pushes

**All confirmed on the remote.** `origin/claude/bead-bonanza-design-review-ne1plu`
is at `c48e995`, identical to local `HEAD`, and all ten bead commits
(`5cc1ce9..99ee4a4`) are ancestors of it. Working tree clean.

`push_mode` is `defer_push` (source: human) and `pull_request` is `none`
(source: human). No PR was created and nothing was merged to the default
branch — PR #27 already exists for this branch and predates the run.

## 6a · Durable trigger

**Durable trigger: armed, not cancelled — `trig_01TQcbztLi1xirmBeGHxjJar`,
next fire 2026-08-24T18:24Z.**

The trigger that fired at 17:57 (§1.5) re-armed itself at a short interval
after the stall. Cancelling it belongs to post-flight, not to this report; it is
named here so a forgotten cancellation is distinguishable from a deliberate one.

**One thing to know before it fires:** its prompt asserts "`bd search` is
NON-FUNCTIONAL on this build — zero hits even for common terms" as a hard
constraint. That is the false claim §1.2 retracts. It is a **fifth**
propagation site for it, and a session resuming from that trigger will inherit
the error unless the prompt is corrected or the trigger is cancelled.

## 7 · Run epic

`agent-plugins-so39` — "Run 2: make the coordination plan trustworthy, and fix
what run 1 shipped". The manifest, the amendments, every incident comment and
every bead the run touched are reachable from it:

```sh
bd show agent-plugins-so39 --json | jq '.[0].metadata'
bd comments agent-plugins-so39
```

## 8 · Review

**Reviewer: declared.** `sf.toml`'s `[review]` names `skill = "code-review"`,
and that is what ran. No fallback was used, so this is the project's own
reviewer and the strongest of the three evidence classes the skill defines.

One pass of a 2-loop budget. The loop closed at pass 1 because zero findings
were admitted, not because the budget was spent.

**A gap in the record:** the manifest carries `review_passes` and
`review_budget` but **no `reviewer_source` key**. The `declared` verdict above
is established from `sf.toml` and the hand-off, not from the manifest —
`sprint-review`'s own return contract names `reviewer_source` and this run did
not persist it. A later reader of the epic alone cannot tell which of the three
reviewers ran.

## 9 · Scope — what this run did not cover

**Twenty-two beads were excluded at admission.** This run finished its admitted
set over a **deliberately narrowed scope**, and does not represent completion
of the coordination rollout.

| Group | Count | Reason, as recorded in `manifest.excluded` |
|---|---|---|
| `qozp.6`–`.11`, `.13`, `.14`, `.15` | 9 | Rollout plan items 6–14 — **deferred to run 3.** The admitted `4tjk` / `gu3i` / `bbtl` correct the *documents that specify* items 6–14, so run 3 works from a corrected plan. |
| `qozp.21`, `.22`, `.24`, `.25`, `.26`, `.27` | 6 | Concern `sprint-wave` or wave-indexed ids, which **plan item 8 deletes outright**. To be retired *by* item 8 rather than fixed on code slated for deletion. (`qozp.22` needs a disposition first — §4.) |
| `agent-plugins-o76b` | 1 | Human-decision lifecycle design; deferred by the human pending a design document. |
| `agent-plugins-we50`, `-2nb8` | 2 | Upstream `bd` bug reports, not this project's code. |
| `agent-plugins-vqmc`, `-r9yj`, `qozp.17` | 3 | Real, but not admitted this run. |
| `agent-plugins-vcka` | 1 | Duplicate of admitted `xpn8`; flagged by `sprint-scope`, **not** closed or moved unilaterally. Since resolved by supersede (§2.1). |

**Findings left at loop exhaustion:** none — the review loop closed at pass 1
inside its budget. The ten unadmitted findings (§4) were left by the *scope
gate*, not by an exhausted budget, which is a different and more serious thing
(§1.1).

## 10 · What this run could not verify

- **CI. Nothing green ran.** GitHub Actions has been down at the account level
  since ~02:26 UTC. Verified again while writing this report: on the current
  head `c48e995`, the `checks` and `cli` workflow runs both created and
  completed at 18:07:54–18:07:59 — **4 and 5 seconds**, `conclusion: failure`,
  no log content. Every commit on this branch back to `c25818e` shows the same
  no-runner signature. **The red checks on PR #27 are not a verdict on this
  run's code.**
- **What stands in CI's place**, and it is weaker: `bb-checks.py` and
  `scripts/ci.sh` both exit 0 locally on `99ee4a4`, run by the coordinator. This
  report did not re-run them (a report must not cost a second pass over the
  run), so that is a relayed result, not one measured here. The head has moved
  to `c48e995` since — one markdown file added — and no gate has been run on
  that sha at all.
- **The first red `skill invariants and vendored references` job after Actions
  recovers must be re-run before it is believed.** That CI job runs
  `sync-shared-references.sh --check`, which `agent-plugins-b9ft` proves flaky
  at ~2-in-30 (§1.4). A single red there is as likely to be the flake as a
  regression.
- **Three `SKILL.md` files are effectively full**, measured on the current head
  with the checker's own method (`text.split('---', 2)[2]`, UTF-8 length)
  against the 20,000-byte cap: `sprint-plan` 19,946 (**54 free**),
  `groom-beads` 19,890 (**110 free**), `bead-bonanza` 19,240 (760 free). Three
  of this run's beads paid for their additions by cutting text. What cannot be
  verified is whether that remains safe: at 54 bytes free, the next bead
  touching `sprint-plan/SKILL.md` cannot add a sentence, and the pressure falls
  on whoever is least able to tell a restatement from a rule.
- **The light-tier silence has no explanation, only two candidates** (§3).
  Whether `qozp.23`, `qozp.28` and `l1p2` genuinely produced no friction or
  whether the findings step was dropped is **not determinable from what was
  recorded** — which is precisely the defect.
- **`dispatch_state` / `terminal_state` are written on one dispatch path of
  two.** Measured, same run, same wave: `4tjk` (dispatched to `run-bead`
  directly) returned both keys `null`; `qozp.19` and `d6zz` (dispatched through
  `sprint-batch`) returned `returned` / `completed`. Nothing failed here,
  because `bb-status.sh bead` reads git and status rather than the dispatch
  record. But `sprint-cleanup.sh` decides whether a standing worktree holds a
  live worker by `dispatch_state=running`, and supervision's three-deaths
  diagnosis has **no signal at all** when the key is absent. Both are degraded
  and neither was exercised this run.
- **`reviewer_source` was not persisted** to the manifest (§8). The `declared`
  verdict rests on `sf.toml` plus the hand-off, not on the run's own record.
- **The `agent-plugins-gu3i` commit sha in the hand-off was wrong** (`cfcdcd6`,
  a pre-rebase commit not on the branch; the real one is `d6ad144` — §2). Every
  other sha verified. That one error was caught by checking; **what cannot be
  ruled out is that other relayed figures in this report carry the same class of
  error**, since a report that re-derives the run is a report that gets skipped.
  Every figure here that was independently measured is marked as such.

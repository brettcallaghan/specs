# Sprint report — Roll coordination onto the beads graph (items 1–14)

| Field | Value |
|-------|-------|
| Date | 2026-08-24 |
| Run epic | `agent-plugins-qozp` |
| Branch | `claude/bead-bonanza-design-review-ne1plu` (PR #27) |
| Base sha | `aba7299` |
| Head | `9a88c6f` |
| Admitted | 6 beads (5 at admission, `agent-plugins-h3j8` mid-run) |
| Excluded at admission | 10 beads — **the run's scope was deliberately narrowed** (§9) |
| Outcome | **6 closed, 0 blocked, 0 skipped** |
| Waves | 2 |
| Review | 1 pass of a 2-loop budget; 11 findings filed, **0 admitted** |
| Retro | 33 findings harvested, 16 candidates drafted, **0 filed — awaiting one human decision** |
| Findings, verbatim | `rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md` |

This run was a trial of the `bead-bonanza` skill family on a queue whose
subject matter is that same coordination tooling. Read §1 before §2.

---

## 1 · Incidents

Five. Two of them are **not reachable by this skill's own incident query**
(`^(retro:[a-z]+:[^ ]+ incident|amend: incident)`) and were read directly from
the tracker on the retro's instruction — they are the two most consequential
records the run produced, and they lead.

### 1.1 — The run admitted none of its review (and the "P0" was later disproved)

`agent-plugins-qozp` comment #19, verbatim in part:

> incident: review pass 1 filed 11 findings and the run admitted ZERO of them,
> including a P0. `standing_auth.admit_review_findings` is
> `{value:false, source:'default'}` — **it took its default, nobody chose it** —
> and the skill's in-scope test requires BOTH a defect in what the run changed
> AND gate authorization for scope growth, so the second condition fails
> regardless of severity.

What was decided: the review loop **ended at pass 1** with 0 admitted, well
inside the budget of 2 loops, and all 11 findings were deferred with
`grooming_status=out-of-scope-this-run`.

What was rejected, and at what cost: admitting `agent-plugins-qozp.18` — then
believed to be a P0 silent-data-loss path **in code this run merged**
(`872f390`) — was rejected because the scope gate is unauthorized, not because
the defect looked small. The mechanism is what matters here and it is
unchanged by §1.2's correction: the gate would have refused a genuine P0 just
as readily, because severity is not an input to it.

A second cost, recorded in the same comment: the 11 findings **entered the
ready front on creation**, so without the manual deferral the next wave
boundary would have silently admitted exactly what the gate had just refused.
The deferral was the coordinator's act, not the skill's.

Escalation was not attempted mid-run, per the operating principle that a run
records rather than escalates.

### 1.2 — The "P0", and why the verification of it did not hold

`agent-plugins-qozp.18` comment #1 — the coordinator's own confirmation against
source rather than acceptance of the reviewer's testimony:

> `pre-flight-check.sh`'s `concurrency_cap` block at lines 712–736 never reads
> `bd show`'s return code — `r = subprocess.run(...)` then
> `json.loads(r.stdout)` with `except Exception: cur = {}` — and then calls
> `bd update --metadata json.dumps(cur)`, which REPLACES the object. On any
> `bd show` failure the manifest is reduced to `{"concurrency_cap": N}`,
> destroying `admitted`, `excluded`, `moved`, `branch`, `base_sha`,
> `standing_auth`, `checks_ids`, `checks_baseline`, `review_passes` and
> `review_budget`, and the script prints `written: concurrency_cap` — **a
> success line over a wipe**.
>
> IMPORTANT ADDITION not in the original finding: the IDENTICAL pattern is
> **pre-existing at lines 653–672** (the `checks_baseline` write), so
> `agent-plugins-h3j8` duplicated an existing defect rather than introducing a
> new class of one. A fix must repair BOTH sites, and the two should share one
> read-merge-write helper rather than a third copy.

**CORRECTION, added after the run — the P0 premise is false.** Both the
review finding and the coordinator's verification above rest on
`bd update --metadata` replacing the object. Measured afterwards on bd 1.2.2,
it does not: seeding `{"alpha":1,"beta":"two","gamma":["x","y"]}` and then
writing `{"delta":9}` leaves all four keys. A second control drove the exact
old code with a `PATH` shim failing only `bd show` — the transient-read case —
and a seeded manifest carrying `admitted`, `branch`, `base_sha`,
`standing_auth` and `review_passes` came through intact. **There was no
data-loss path, and nothing was shipped broken.**

The verification checked the code shape and not the consequence, which is the
lesson worth keeping: a code-shape match plus a *documented* behaviour is not a
measurement. The false premise came from
`groom-beads/references/capability-tier.md:57-60`, which states the replace
semantics as fact — corrected under `agent-plugins-4tjk`, and the more
valuable finding of the two, because it propagated into a P0 review finding
and a coordinator verification before anyone measured it.

What survived the correction, and was fixed under `agent-plugins-qozp.18`
(P2, commit `a3e7d38`): a failed read was reported as a write success, and
the script staked the admission record on undocumented merge behaviour that
is only **shallow** — a nested value IS replaced wholesale (`{"a":1,"b":2}`
overwritten with `{"c":3}` yields `{"c":3}`), and `standing_auth` is nested.
Both sites now share one `manifest_merge` helper that fails loudly on an
unreadable manifest instead of guessing an empty one.

### 1.3 — `agent-plugins-qozp.2` — the roster cannot be cited by path

Decided: `sprint-wave` cites `bead-bonanza`'s `references/dispatch-policy.md`
**by role in prose**, and the roster keeps its single home there.
Rejected: citing it by path, which `scripts/check-skills.sh` rule 3 forbids
(rule 2 additionally requires a backticked path to resolve locally).
Rejected explicitly: vendoring a second copy — the bead's AC forbade it.
Cost: the up-front decision the bead was dispatched with was overturned
mid-bead, and the reference is now discoverable only by prose description.

### 1.4 — `agent-plugins-qozp.3` — plan item 5 declined, no commit

Decided: decline `bd setup claude`'s update. `CLAUDE.md` and
`.claude/settings.json` are byte-identical to HEAD, so the bead lands with **no
commit and no diff** — a branch its own acceptance criteria explicitly
permitted.
Rejected: satisfying the item's Verify line (`bd setup claude --check reports
current`), which is unsatisfiable without installing 53 lines already present
verbatim in `AGENTS.md`.
Cost: plan item 5 is not done and cannot be done as written; two beads were
filed instead — `agent-plugins-2nb8` (upstream `bd` defects) and
`agent-plugins-vqmc` (this repo's split BEGIN/END markers, needs an owner
decision).

### 1.5 — `agent-plugins-qozp.5` — the bead's scope could not satisfy its own AC

Decided: expand the file set by two beyond the bead's two-file Design — one MAP
line in `scripts/sync-shared-references.sh` and the generated
`plugin/skills/sprint-plan/references/shared/epic-definition-of-ready.md` it
produces.
Rejected: leaving the citation AC unsatisfied, or citing across skills by path
(the same check-skills rules 2 and 3 as §1.3).
Cost: scope growth on a bead, and — noted by the retro as a corpus-level
problem — **`qozp.2` and `qozp.5` hit the same rule in one run and resolved it
opposite ways**: prose in one, vendoring in the other. Both are defensible
against their own ACs; together they leave the corpus demonstrating two
incompatible answers to one question. That is candidate `P1` in the findings
document.

---

## 2 · Results

| ID | Status | Batch | Commit | Notes |
|----|--------|-------|--------|-------|
| `agent-plugins-qozp.1` | closed | wave 1, `dev-qozp-1` | `c25818e` | Item 1 — `bd epic close-eligible` in the coordinator loop. Clean, no findings beyond scope. |
| `agent-plugins-qozp.4` | closed | wave 1, `dev-qozp-4` | `05af5b6` | Item 2 — cap the dev pool; record `concurrency_cap`. Discovered `h3j8`. |
| `agent-plugins-qozp.2` | closed | wave 1, `dev-qozp-2` | `8a8b86f` | Item 4 — `BEADS_ACTOR` and the agent roster. Incident §1.3. |
| `agent-plugins-qozp.3` | closed | wave 1, `dev-qozp-3` | **none — by design** | Item 5 — declined; `metadata.outcome=declined-no-diff`. Incident §1.4. |
| `agent-plugins-h3j8` | closed | wave 2, `dev-qozp-5` | `872f390` | Admitted mid-run. Carries the P0 of §1.2. |
| `agent-plugins-qozp.5` | closed | wave 2, `dev-qozp-5` | `9a88c6f` | Item 3 — two Definition-of-Ready checks. Incident §1.5. |

**Totals: 6 closed, 0 blocked, 0 skipped.**

`agent-plugins-qozp.3`'s missing commit is a correct outcome, not a gap: its
acceptance permitted a decline branch and it took it. `bb-status.sh bead`
cannot express that and exits 1 on it — recorded as a finding, candidate `S9`.

---

## 3 · Implicit decisions made

Choices the beads delegated to the subagent, or that the subagent had to make
without a stated rule. Incidents are in §1 and are not repeated.

| Bead | What was chosen |
|------|-----------------|
| `agent-plugins-h3j8` | The design offered two placements for the manifest merge — hoist section 8's merge out of its "checks were declared" branch, or add a second merge guarded on `EPIC`. Took the second: a new section 8b with its own python heredoc. |
| `agent-plugins-qozp.4` | Which expository prose to relocate out of `sprint-plan/SKILL.md` — the bead delegated the pick outright. |
| `agent-plugins-qozp.2` | That the agent id and `{DISPATCH_ID}` are one string, so no new placeholder was introduced. |
| `agent-plugins-qozp.1` | None — recorded a clean note. |

### 3a · Tier escalations

**None.** No bead was raised above its assigned tier; every bead completed at
the tier it was groomed to (`qozp.1` standard, `qozp.2`/`.3`/`.4` advanced,
`h3j8`/`qozp.5` standard), and no `retro:` comment on any bead records an
escalation.

`CLAUDE_CODE_SUBAGENT_MODEL` was **not** set — measured `<unset>` in this
agent's environment, and the wave-1 dispatch record names a different model per
bead (`sonnet` for `qozp.1`, `opus` for `qozp.2`/`.3`/`.4`), which a forced
model would have collapsed. So tier routing did occur and the null above is a
real result.

**Read it narrowly.** Escalations reveal *under*-routing only. A bead groomed
`advanced` that `standard` would have finished never appears here — it
succeeds, and the overspend is silent. Zero escalations across six beads is
equally consistent with routing everything one tier too high. This queue is
below the ≥10-bead threshold at which that inference would be worth acting on,
so the honest reading is: **no routing evidence either way**, and run 2 —
which inherits eight of these beads — is the first run large enough to say
something.

### 3b · Deliberate over-routes

**None.** Wave 2's chained batch paired `h3j8` and `qozp.5`, both already
`standard`, so the tier-batching rule never had to raise anything. Wave 1's
four beads were dispatched individually at their own tiers.

---

## 4 · Follow-up beads

### Created mid-run and admitted into it

| Bead | Outcome |
|------|---------|
| `agent-plugins-h3j8` | **Closed** at `872f390`. Discovered by `qozp.4`: `run-record.md` now specifies `manifest.concurrency_cap` but `pre-flight-check.sh` never wrote it, so item 3's within-cap DoR check (`qozp.5`) would have read `null`. Admitted as a producer for `qozp.5` in the same wave; file sets verified disjoint. |

### Created mid-run and left for the user

| Bead | P | Origin | Why it was not admitted |
|------|---|--------|-------------------------|
| `agent-plugins-2nb8` | P2 | `qozp.3` | Upstream `bd setup` defects — belongs to `gastownhall/beads`, not this repo. |
| `agent-plugins-vqmc` | P2 | `qozp.3` | Which file owns the managed BEADS INTEGRATION block; needs an owner decision. |
| `agent-plugins-l1p2` | P3 | `qozp.4` | Stray code fence in `run-record.md`. Cosmetic; no declared check catches it. |
| `agent-plugins-qozp.17` | P2 | `qozp.2` (`discovered-from`) | Ceremony agents should export their roster `BEADS_ACTOR` at their dispatch sites. Now **blocks** `agent-plugins-qozp.7`. |

### Filed by the review, all deferred

Eleven, `agent-plugins-qozp.18` through `.28` — see §8.

---

## 5 · Stashes

**None.** `git stash list` is empty; no recovery ran.

## 5a · The retro's outcome

```
Retro: 0 filed, 0 declined, 0 routed to .sf/outbox/
       33 findings harvested across 7 beads; 16 candidates drafted
       retro_status = candidates-pending-human-decision
Every finding, verbatim: rfcs/sprints/2026-08-24-beads-coordination-rollout-findings.md
```

The sixteen candidates — twelve against the skills, four against this project —
are drafted in full in §5 of that document and **await one human decision**.
Nothing is filed without it, and the run did not block on it; the coordinator
is relaying the decision. Two of the 33 findings are unrouted (a kindless clean
note on `qozp.1`, and an invented kind `process` on the epic). Five were deduped
against beads already filed. Recurrence is 0 for all sixteen — `.sf/outbox/` is
empty and `rfcs/inbox/` carries no prior issue against any of these targets,
which is a first-observation, not a clean bill.

`rfcs/retro-log.md` was deliberately **not** appended: its header scopes it to
items ingested by triage, and nothing here has been triaged.

## 6 · Pushes

All five commits are on the remote. Verified directly rather than from the
run's account: `git log aba7299..9a88c6f` lists exactly `c25818e`, `05af5b6`,
`8a8b86f`, `872f390`, `9a88c6f`, and `git branch -r --contains 9a88c6f` returns
`origin/claude/bead-bonanza-design-review-ne1plu`. `agent-plugins-qozp.3` has
no commit by design and therefore nothing to push.

No PR was created and nothing was merged to the default branch — PR #27 already
tracks this branch, and `standing_auth.merge_to_default` is `false`.

## 6a · Durable trigger

`Durable trigger: fired and self-disabled` — the run's recovery trigger was
armed as a one-shot before the first dispatch and re-armed twice as the run
outlasted each firing. The last of the chain
(`trig_01QQbVUwpoFyD7CfwyzpAmsC`) fired at 08:58 UTC into a live session,
found the run mid-ceremony, and self-disabled with
`ended_reason: run_once_fired`. None was cancelled, because a one-shot that
has fired needs no cancelling — so post-flight's "cancel the durable trigger"
step has nothing to act on. **The mechanism is worth noting for the retro: the
skill says to arm one durable trigger and cancel it at post-flight, but a
one-shot cannot cover a run of unknown length, so the coordinator re-armed by
hand each time it fired.**

One trigger unrelated to this run's machinery remains armed:
`trig_016ED5Y67Xci3NKj2rc4WmND`, an 11:02 UTC re-check of PR #27's CI, which
exists because of the Actions outage in §10.

## 7 · Run epic

`agent-plugins-qozp` — the manifest, the amendments, the incident record and
every bead the run touched hang off it.

---

## 8 · Review

**One pass, of a 2-loop budget.** The reviewer was `code-review`, the skill
`sf.toml`'s `[review]` **declares** — `reviewer_source: declared`, no fallback.
This is the strongest of the three evidence grades: it is the project's own
reviewer, not a skill found in the session and not a bare agent.

Eleven findings, all filed against their origin, **all deferred**:

| Bead | P | Finding |
|------|---|---------|
| `agent-plugins-qozp.18` | ~~P0~~ **P2** | Filed as "can silently replace the entire run manifest"; **disproved after the run** — `--metadata` merges on bd 1.2.2 (§1.2). What was real — a failed read reported as a write success — is fixed in `a3e7d38` |
| `agent-plugins-qozp.19` | P1 | `bd epic close-eligible` takes no epic argument — the coordinator loop closes epics repo-wide |
| `agent-plugins-qozp.20` | P1 | The new epic-DoR assignee gate has no producer and runs before assignees can exist |
| `agent-plugins-qozp.21` | P1 | `sprint-wave` cannot resolve `{DISPATCH_ID}` — the derivation rule exists only in `bead-bonanza`'s references |
| `agent-plugins-qozp.22` | P1 | batch-id is "position in its wave", so agent ids collide across waves of the same run |
| `agent-plugins-qozp.23` | P1 | `sprint-batch` issues `bd update` calls but never exports `BEADS_ACTOR` |
| `agent-plugins-qozp.24` | P2 | `run-record.md`'s cap check counts every assignee despite the line above saying "dev assignees only" |
| `agent-plugins-qozp.25` | P2 | The "cap − 1" budget is justified two incompatible ways across four files |
| `agent-plugins-qozp.26` | P2 | The epic-DoR cap snippet reports "budget -1" and exits 0 when `concurrency_cap` is absent |
| `agent-plugins-qozp.27` | P2 | The dev-pool cap check uses `--status all`, inflating N with agents from already-closed waves |
| `agent-plugins-qozp.28` | P3 | `references/scripts.md`'s side-effect list for `pre-flight-check.sh` omits the `concurrency_cap` write |

**The loop closed because zero were admitted, not because the budget ran out.**
One loop of the two is unspent. The reason is §1.1, and it is worth restating
plainly: the gate that refused them was never chosen — it took its default.

One thing reviewed and dismissed: `review-qozp` flagged the manifest's
`review_budget=2` against the coordinator's "budget is 3 passes" as a
discrepancy. It is not one — `sprint-scope/SKILL.md` defines `review_budget` as
"2 loops (three passes total)", so both statements are the same statement.
Recorded so a resumed session does not re-litigate it.

---

## 9 · Scope — what this run did not cover

**Ten beads were excluded at admission.** This run finished its admitted set
over a **deliberately narrowed scope**; it is not complete over the epic the
user first described.

| Beads | Reason recorded in `manifest.excluded` |
|-------|----------------------------------------|
| `qozp.6` `qozp.7` `qozp.8` `qozp.9` `qozp.10` `qozp.11` `qozp.14` `qozp.15` | trial run admits plan items 1–5 only; undefer for run 2 |
| `qozp.12` | DECISION D1 needs a human answer |
| `qozp.13` | owner: plan item 12 Stop hook not part of the bonanza |

Eight of those are plan items 6–14 held for run 2. Two are not: `qozp.12` is
blocked on a human, and `qozp.13` was ruled out of the bonanza's scope by its
owner.

Together with the eleven deferred review findings, **twenty-one beads under
this epic are currently deferred.** Nothing in this report undefers any of
them, and it files no beads of its own.

Two of the review findings bear directly on run 2's readiness before it starts:
`qozp.20` (the epic-DoR assignee gate has no producer) and `qozp.26` (the cap
snippet exits 0 with a nonsense budget when `concurrency_cap` is absent) are
both about gates that run 2 will hit at its first wave boundary.

---

## 10 · What this run could not verify

- **CI never ran.** GitHub Actions is down at the account level. Every job
  across ubuntu, macos and windows starts in the second it is created and dies
  in 2–4 seconds with no log content; log download returns 404. Verified
  independently on each pushed commit. **The red checks on PR #27 are a
  platform state, not a bead failure** — but the corollary is that no bead in
  this run has been verified on any platform other than this container. The
  three-platform matrix and the busybox smoke test are unrun.
- **Neither dispatch path was verified end-to-end as designed.** Wave 1's four
  dispatches were made by the coordinator directly rather than through
  `sprint-wave`, so worktree and branch names came from the harness and the
  coordinator mapped harness ids to bead ids by hand — and `sprint-wave`'s
  verification step was performed by the party it is supposed to be independent
  of. Wave 2 did invoke `sprint-wave`, which then found **no Task/Agent tool
  exposed to it**, and rather than stopping it absorbed the batch and ran
  `sprint-batch` → `run-bead` itself. So wave 2's verification was
  self-verification too. `pre-flight-check.sh` models this exact condition as
  `BB_ASSUME_NO_AGENT=1` and blocks on it, but pre-flight ran in the
  coordinator's session where the Agent tool **is** available: **the capability
  is per-agent, not per-session**, and a run-level probe cannot answer it for a
  subagent. Candidates `S1` and `S2`.
- **The gate checks were run by the coordinator, and this report did not re-run
  them.** `bb-checks.py` (all 13 declared checks), `scripts/ci.sh`, and
  `sync-shared-references.sh --check` are all reported exit 0 on `9a88c6f`.
  The first two are corroborated by the coordinator's own amendment comment on
  the epic; `sync-shared-references.sh --check` rests on the handoff alone. Per
  this skill, a report does not cost a second pass over the run — so treat
  these as the run's testimony, checked against its record, not as independent
  confirmation.
- **~~The P0 of §1.2 is confirmed but unfixed.~~ Superseded:** it was
  disproved after the run (§1.2) and the residual defect is fixed in
  `a3e7d38`. What this entry got right, and what remains unverified, is
  narrower: nothing logged which branch the old `except` path took, so
  "it never fired during this run" is still inference from an intact
  manifest rather than a measurement. The new `manifest_merge` makes the
  question moot going forward — that path now exits non-zero and names
  the cause instead of continuing silently.
- **The three environmental entries the retro drafted have no owner yet** and
  are not in `AGENTS.md`: `bd setup claude`'s conditional "stale" verdict, the
  Bash worktree-isolation guard's non-deterministic refusals, and the Actions
  outage above. Candidate `P4`. The Actions entry in particular is flagged in
  the findings document as a coordinator observation rather than a captured
  `retro:` finding.
- **`effort` is `null` in the manifest** — this harness exposes no effort
  setting, so every candidate issue carries that verbatim rather than a guess.
- **Routing calibration: null result.** See §3a — zero escalations over six
  beads distinguishes nothing.

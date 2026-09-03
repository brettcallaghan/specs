# Retro Log

One line per item ingested by `triage-rfc`. This is the complete record of
what the RFC pipeline has processed: every RFC and retro note, its origin,
the phase or skill retroed, and its outcome. Newest entries at the bottom.

Outcome counts are per proposal, not per RFC — the proposal is the unit of
decision.

| Date | Origin | Retroed | Item | Outcome |
|------|--------|---------|------|---------|
| 2026-07-31 | self | `spec` | RFC 0001 | 3 adopted (already implemented), 1 deferred |
| 2026-07-31 | self | `triage-rfc` | RFC 0002 | 4 adopted |
| 2026-07-31 | self | `groom-beads` | RFC 0003 | 4 adopted (1 with refinement) |
| 2026-07-31 | self | `bead-bonanza` | RFC 0004 | 4 adopted, 1 deferred |
| 2026-07-31 | self | `retro-skill` | RFC 0005 | 2 adopted |
| 2026-07-31 | self | `spec` | RFC 0006 | 4 adopted (1 with refinement), 1 deferred |
| 2026-07-31 | self | FRD authoring | retro note | null retro |
| 2026-08-01 | self | `spec` | RFC 0007 | 3 adopted (2 with refinement) |
| 2026-08-01 | self | `triage-rfc` | RFC 0008 | 6 adopted (1 with refinement) |
| 2026-08-01 | self | `triage-rfc` | RFC 0009 | 4 adopted (1 with refinement), 1 deferred, 1 triage-originated adopted |
| 2026-08-06 | self | spec (frontmatter relationship vocabulary) | RFC 0012 | 8 adopted (4 with refinement) |
| 2026-08-10 | self | `bead-bonanza` | RFC 0016 | 3 adopted (1 with refinement, 1 since retired), 1 rejected, 3 triage-originated adopted |
| 2026-08-10 | self | `bead-bonanza` | RFC 0017 | 5 adopted (3 with refinement), 1 triage-originated adopted |
| 2026-08-10 | self | `bead-bonanza` | RFC 0018 | 6 adopted (3 with refinement, 1 of those incomplete, 1 since retired), 2 triage-originated adopted |
| 2026-08-12 | self | `bead-bonanza` sprint (run `1786562615-17`) | [sprint report](sprints/2026-08-12-bead-bonanza-hardening.md) | 73 findings harvested, 16/17 beads closed; 11 cited by the two inbox RFCs, the remaining 62 unfiled and retained for recurrence |
| 2026-08-13 | self | `bead-bonanza` sprint (run `1786600009-7`) | [sprint report](sprints/2026-08-13-rfc-triage-adoption.md) | 14 findings harvested, 7/7 beads closed; implements the 2026-08-13 triage of RFCs `rfc-self-report-citations-and-batch-base` and `rfc-self-line-anchors-rot-across-a-sprint` |
| 2026-08-13 | self | `bead-bonanza` sprint (run `1786603162-7`) | [sprint report](sprints/2026-08-13-template-drift-and-check-report.md) | 16 findings harvested, 7/7 beads closed; 1 follow-up bead filed (agent-plugins-nu2); RFCs pending owner triage |
| 2026-08-14 | self | `bead-bonanza` sprint (run `1786686364-15`) | [sprint report](sprints/2026-08-14-markdown-links-and-shipped-specs.md) | 64 findings harvested, 16/16 beads closed (15 queued + 1 admitted mid-run); 4 RFCs filed — `bead-bonanza`, `plan-sprint`, `recover-sprint`, `standards` (rfc-0020); tier calibration not routed (zero escalations, no over-route signal); report written after the RFCs because Phase 4 was skipped and reconstructed |
| 2026-08-15 | self | `bead-bonanza` sprint (run `1786778523-19`) | [sprint report](sprints/2026-08-15-gherkin-adoption.md) | 73 findings harvested, 21/21 beads closed (20 queued + 1 admitted mid-run); 3 RFCs filed across the feature's phases (rfc-0021 FRD, rfc-0022 grooming, rfc-0023 implementation); `retro-skill` and `groom-beads` routes offered, not filed; 3 follow-up defects named (AC-GHK-006.6 proved 3 of 5 rows, VTAG's AC-SPEC-009.2 budget breach, ADR-021/guide disagreement); 1 bead filed (agent-plugins-h45m); zero tier escalations, 2 deliberate over-routes |
| 2026-08-24 | self | `bead-bonanza` sprint (epic `agent-plugins-qozp`) | tracker only — see `agent-plugins-qozp` | 33 findings harvested, 6/6 beads closed (5 queued + 1 admitted mid-run); 16 candidates drafted, all 16 accepted by the owner — 12 skill-audience issues written to `.sf/outbox/` (`bead-bonanza` 8, `sprint-review` 2, `sprint-plan` 1, `sprint-wave` 1) and 4 project-audience beads filed (`agent-plugins-uva5`, `agent-plugins-gu3i`, `agent-plugins-bbtl`, `agent-plugins-r9yj`); 0 declined, 2 findings unrouted, 5 deduped against beads already filed; 21 beads left deferred on purpose (10 excluded at admission, 11 review findings the scope gate refused) |
| 2026-08-24 | self | `bead-bonanza` sprint (epic `agent-plugins-so39`) | tracker only — see `agent-plugins-so39` | 28 findings harvested in 27 comments, 10/11 admitted beads closed (1 bounced); 14 candidates drafted — **10 filed, 3 declined, 1 undecided**. 8 skill-audience issues written to `.sf/outbox/`: 5 new (`bead-bonanza` 3, `sprint-batch` 1, `groom-beads` 1), 2 run-1 files extended to `recurrence: 1` (`bead-bonanza` 1, `sprint-review` 1), and 1 run-1 `sprint-plan` file corrected in place — it asserted the `--set-metadata` universal this run measured false. 2 project beads filed (`agent-plugins-9owj` P1, `agent-plugins-6pib` P3). Declined: C7 folded into `agent-plugins-x5tk`, C9 superseded by `agent-plugins-0913` (drop both keys rather than repair the write path), C10 deferred to the end of the refactor; C14 left undecided. C11 was split — the turn-ending invariant filed, the trigger-interval half dropped as subsumed by `agent-plugins-qozp.15`. 2 findings unrouted (both `retro:process:`, both the coordinator's), 3 deduped against beads already filed. 33 beads left deferred on purpose: 22 excluded at admission, 10 review findings the scope gate refused, and `agent-plugins-xpn8` bounced and `human`-labelled |
| 2026-08-25 | self | `bead-bonanza` sprint (epic `agent-plugins-mpnf`) | [findings](../docs/research/sprint-findings/2026-08-25-agent-plugins-mpnf.md) | 52 findings harvested (49 in flight + 3 by the retro), 11/23 admitted beads closed as `mpnf.N` children, 27 closed across the run's reachable set; 14 candidates drafted — **12 filed, 2 declined, 7 clusters deduped into existing beads**. NO outbox files and NO sprint document: commit `9cc8844` retired both channels at the owner's instruction, so 8 skill-audience issues were filed as beads under `agent-plugins-y63n` (`.18`–`.25`) and the findings document went to `docs/research/` per the where-documents-live rule. 4 project beads filed (`agent-plugins-sdu1` P2, `agent-plugins-eejf` P2, `agent-plugins-bfng` P3, `agent-plugins-lmrl` P3). Deduped onto `offx.11`, `offx.8`, `offx.4`, `y63n.17`, `y63n.10`, `r9yj`, `p567`. **2 findings unrouted — third consecutive run at 2**, and the new invention `retro:incident:` sent the run's most consequential finding (a removed worktree coexisting with an unmerged branch) to no audience at all. 15 beads left deferred: review findings refused at admission or de-admitted by owner decision as the branch wound up |
| 2026-08-29 | self | `bead-bonanza` sprint (epic `agent-plugins-bd97`) | [retro](2026-08-29-bd97-swarm-core.md) | **117 findings harvested** across 26 beads (the run's densest yet), 24/24 admitted beads closed (19 planned + 4 admitted mid-run), 1 deliberately deferred out (`ih18.22`); diff `3a57dc7..24f97e3` — 24 commits, 99 files, +17914/-839, nothing stranded. **8 candidates drafted — 0 filed, 0 declined, 8 AWAITING THE OWNER'S DECISION.** 16 beads were already filed by the coordinator in flight and are not re-proposed. Review ran 2 passes of a budget of 2: **17 findings, 0 admitted** (`admit_review_findings=false`, held to deliberately when the reviewer asked for an override). Two findings outrank the rest and are at the top of the record: `bd97.4.8`+`.4.9` — **the vended skill family is currently unusable by any project but this one** (pre-flight hard-blocks without `swarm`; the only remedy printed is a build path no consumer has; the capability probe passes a stale binary, verified live in this container) — and the run's own dominant cost, a collision mechanism that serialized 20 beads into **17 waves at max width 2, 14 of them width 1** (`221f`, `tvom`). Routing: 67 project, 18 skill, 26 `proof` and 3 `environment` terminating, **3 unrouted — a fourth consecutive run with unrouted findings and the first above 2**; all 3 are coordinator-authored `retro:process:`/`retro:work-clean`, the same invention `so39` and `mpnf` recorded (candidate C8). The coordinator retracted three of its own claims mid-run, **all three the same defect** — reading a verdict through a pipe, a `tail` truncation, or a status discarded to `/dev/null` — the rule `AGENTS.md` names and does not check (candidate C7). Could not verify: the **Windows CI leg has never executed** (`total_count` 0), the 13 checks were **never run over the merged head** (post-flight owns it), and `.4.9`'s stale binary was measured in this container only |
| 2026-09-02 | self | `bead-bonanza` sprint (epic `agent-plugins-e6xt`) | [retro](2026-09-02-e6xt-swarm-dispatch.md) | **113 findings harvested** across 42 beads, 19/19 admitted work orders closed plus 13 review-loop beads worked and 1 deferred; diff `2490026..37a9d83` — 32 commits, 87 files, +12376/-720, nothing stranded, 13/13 declared checks green at head (re-measured, not relayed). **13 candidates drafted — 0 filed, 0 declined, 13 AWAITING THE OWNER'S DECISION.** 13 beads were already filed by the run in flight and are not re-proposed. Review ran 2 passes of a budget of 2 with the DECLARED reviewer (`code-review`): **14 findings, 13 admitted and worked** — all row 1, defects the run's own diff introduced. **Every one of pass 2's eight findings traces to one of loop 1's six fix commits**, four of the eight to a single commit, and the budget guard then refused pass 3, so loop 2's seven commits are merged and unreviewed (`agent-plugins-3vxe`; candidate C13, the only time-sensitive item). Routing: 66 project, 7 skill, 30 `proof` and 10 `environment` terminating, **0 unrouted — the first clean run in five**. Two candidates recur from `bd97`, whose eight were never filed (`bd-health.sh`'s repo-global gate, the skill-file byte budget) and one is refined by it (the `swarm` capability probe, which this run shows cannot detect a binary the run itself makes stale — it put a false coordinator measurement into the tracker). Could not verify: **`bd dolt push` failed on every attempt all run and once more at retro time**, so all 113 findings, 32 closures and 13 filed beads exist only in this container — the previous run's epics (`bd97`, `ih18`) are already absent from this database, so the divergence is bidirectional and this document is the only durable copy; **CI has never executed on this branch** (five check runs, all `skipped`, because PR #33 is a draft); `AC-SWARM-DISP-004.1` proved at unit against a declared e2e level; FRD ceremony step 2 implemented nowhere; `1vya`/`gqwl`/`asv8` all green in e2e with no failing test; manifest `effort` reads `unrecorded` |
| 2026-09-02 | self | `bead-bonanza` sprint (epic `agent-plugins-bguq`) | [retro](2026-09-02-bguq-swarm-dispatch-defects.md) | **29 findings harvested** — 26 in flight across 7 beads, 3 by the retro; 7/7 admitted beads closed, 0 acceptance failures, 0 bounces, 0 tier escalations, breaker never advanced past 0; diff `e9e36fa..36598cd` — 7 commits, 14 files, +1117/-35, nothing stranded, 13/13 declared checks green at the merged head (re-measured, not relayed), and `bd dolt push` succeeds again after failing on every attempt in run `e6xt`. **8 candidates drafted — 0 filed, 0 declined, 8 AWAITING THE OWNER'S DECISION.** 11 beads were already filed in flight or by review and are not re-proposed; 7 `e6xt` siblings stayed deferred by decision. Review ran 1 pass of a budget of 2 with the DECLARED reviewer (`code-review`): **9 findings, 7 in scope, 0 worked** (`admit_review_findings=false`, and the coordinator held to it — admission, not budget, was the binding constraint). Two in-scope P2s are surfaced at the top of the record for the human BEFORE merge: `agent-plugins-bi7q` (the new `invocationsRequired` guard can discard `merged:true` after an irreversible `--ff-only` merge — the exact loss `4ca8`/`j5ah` were filed to prevent, bounded unreachable today but justified by a FALSE comment) and `agent-plugins-j2jf` (`u9uh` made `--threshold < 1` fatal while three documents still call 3 merely a default, in the same `sf.toml` block that documents `review_budget` as 0-means-off). **The run's dominant signal: four of seven work orders carried a claim false against the tree** — `74xc`'s AC 1 contradicted by a committed test, `f9lq`'s site-4 analysis refuted two lines below its own quote, `f9lq`'s five-site AC unsatisfiable for three of them, `4ca8`'s Steps to Reproduce unreachable through the real VC — every one inherited from a prior finding and never measured at grooming (candidate C4). Routing: 17 project, 1 skill, 11 `proof` terminating, **0 unrouted — second consecutive clean run**, though not one `tooling` finding was recorded by any agent in flight while two real tooling defects went into an `amend:` comment and a dispatch prompt instead (candidate C2). A coordinator error: `manifest.base_sha` was never set, `swarm manifest require` exited 1, and the review only happened because the reviewer reconstructed and corroborated the range rather than guessing (`agent-plugins-8gzr`, filed, repaired after the fact). Scope grew 5 → 7 mid-run by human instruction; one follow-up filed as `bguq.4.1` was run work by id alone and clearing its parent did not release it — refiled as `agent-plugins-dbvf` (candidate C3). C4 collision detection over-detected: `74xc` and `n3tn` are disjoint by file, function and symbol but share `component-swarm-command-surface`, which spans the whole of `cli/swarm/cli`, so parallelism 1 was accepted deliberately (candidate C5). Could not verify: **which model ran each dispatch** — `manifest.models` and `effort` are both absent, second run running (candidate C6); the harness's worktree DIRECTORY name; `74xc`'s AC 1 and `f9lq`'s sites 1-3 AC as declared; the six-head history of the Windows failure (restated from `agent-plugins-baih`, not re-measured). `go test (windows-latest)` is red at head with a package timeout (`FAIL cli/swarm/e2e 600.094s`) and `jed1` added a ninth e2e fixture to that suite, making `baih`'s token-blocked `-timeout 20m` patch load-bearing (candidate C8) |

## Run notes

**2026-07-31 — first triage run.** 5 RFCs, 19 proposals: 17 adopted (3 of
them already implemented before triage), 2 deferred, 0 rejected. All five
RFCs originated from this repo dogfooding its own skills, so every
proposal carries `origin: self` — no independent project has corroborated
any of them yet, and the accumulation model has therefore not been
exercised.

Zero rejections in a first run is worth noting rather than celebrating: it
is consistent with a well-calibrated bar and equally consistent with an
author and a triager who share every assumption. The two deferrals are the
only places the gate actually bit.

The 14 proposals implemented this run were groomed into beads and
sequenced by `groom-beads` before implementation, because seven of them
collided on shared files — four alone on `triage-rfc/SKILL.md`. The
resulting graph ran in four waves. Grooming also surfaced four defects in
the beads themselves (uniform boilerplate `tier_reason`, a deferral left
dispatchable, no owner for the single plugin version bump, and version-bump
owners not last in their chains), none of which the filing pass had caught.

**2026-08-01 — second triage run.** 1 retro note (null) and 2 RFCs, 9
proposals: 9 adopted (3 with refinement), 0 deferred, 0 rejected. One
further change — adding a `version` property to `c4-guide.md` and
`ears-guide.md` — was triage's own addition rather than either RFC's
proposal, and was approved separately.

Two of the three refinements came from not taking a proposal at its
word. RFC 0007 P1 asked triage to sweep rather than trust its list of
four `@` occurrences; the sweep found two more, in templates the
authoring run never opened. RFC 0007 P2's replacement text named two
JavaScript globals where its own Evidence section listed six — adopting
it verbatim would have shipped guidance that still fails on the third.
Both were caught by re-reading the RFC's evidence against the tree, not
by judging the proposal's plausibility.

Still zero rejections across two runs and 28 proposals. Every RFC so far
carries `origin: self`, so the accumulation model — deferral, recurrence
across independent projects, promotion — remains unexercised: the two
deferred beads from run 1 (`agent-plugins-6tp`, `agent-plugins-ual`) were
neither corroborated nor advanced this run. The bar has still never been
tested against a reporter who does not share the triager's assumptions.

RFC 0008 P1 landed a rule requiring collision-prone implementations to
route through `groom-beads`, and this run then did not follow it — 10
proposals with 6 colliding on one file were sequenced by hand instead.
The reasoning is recorded in that RFC's Resolution. A rule skipped on its
first run by the agent that adopted it is the most interesting thing in
this log for the next retro to look at.

**2026-08-01 — third triage run (RFC 0009).** 1 RFC, 5 proposals plus 1
triage-originated change: 5 adopted (1 with refinement), 1 deferred.
**The first deferral to be exercised rather than merely filed**, and the
first non-zero-rejection-adjacent outcome in the log — the accumulation
model finally has something moving through it.

Two firsts worth recording. **The sequencing rule was followed.** Asked
to choose, the human directed route (a); the resulting `groom-beads`
pass immediately found that all six beads carried `tier`,
`tier_reason` and `grooming_status` as prose in `notes` rather than in
`metadata`, leaving `tier: None` where the Definition of Ready requires
it non-null. That is the second consecutive grooming pass to find
defects the filing pass did not (RFC 0008's Resolution records the
first), which is now a pattern rather than an anecdote.

**The first triage-originated change to go through the front door.** T1
came from the repository owner reading `SKILL.md:122` and saying the
distinction it drew was not clear. The run before it had produced a
similar change and improvised an id and a Resolution heading for it;
this run adopted P3 first, which created the `T` numbering, and then
immediately used it. A rule adopted and exercised inside one run is the
cheapest possible test of it.

`bd-health.sh` exited 0 while reporting one hard check unavailable in
this `bd` build. Exit 0 was not read as a clean bill of health;
acyclicity was checked by hand. Worth remembering the next time that
script returns success.

**2026-08-10 — fourth triage run (RFCs 0016–0018).** 3 RFCs, 15
proposals plus 6 triage-originated changes: 14 adopted (7 with
refinement), 1 rejected, 0 deferred. The implementation ran as an epic
(`agent-plugins-92q`) across two `bead-bonanza` sprints and closed on
2026-08-12; the Resolutions were written last, against the tree as it
stood at the end, rather than at classification time.

**The first rejection in the log, and the first one produced by
measurement.** RFC 0016's P3 asserted that a git worktree cannot see a
`bd` database. Nobody had run the command. Two beads did, in two
checkouts, and the claim is false — `bd` resolves the main checkout's
database through git's common directory from any worktree. The
consequence is larger than one row: the mode P3 asked to *exclude* became
the run model for every bead in the second sprint. Zero rejections across
three prior runs and 28 proposals was, as run 1's note allowed,
"consistent with an author and a triager who share every assumption";
what broke the streak was not a sharper triager but somebody typing the
command the RFC reasoned about.

**Writing Resolutions last was correct, and would have been wrong for a
smaller batch.** Between classification and closeout, `agent-plugins-92q.30`
(one worktree per bead) and `agent-plugins-92q.31` (Dolt remote as the
tracker) deleted premises that adopted proposals rested on. Three
proposals landed and were then retired inside the same epic — RFC 0016's
P1 tracker-write deferral, RFC 0018's P5 WIP-commit exception, and RFC
0017's P2 mtime liveness table. Resolutions written at approval time
would now claim as current three rules the tree no longer contains. The
cost of the ordering is the opposite risk: an approval sitting in the
inbox with no Resolution for two days is exactly the silent-rejection
failure the skill names, and it is only safe because one bead owned the
closeout and ran last.

**A self-adoption recorded as unfinished.** RFC 0018's P3 — fit the
operative procedure under the 5,000-token re-injection cap — is adopted,
decomposed into five beads, and still not met: `SKILL.md`'s body is
61,153 bytes against a ~20,000-byte target, and the measurement bead that
would have verified the fit was closed by the repository owner for a
dedicated later round. The reordering that did land changes *what* is
truncated, from the close ceremony and the script interfaces to
explanatory prose, which is the whole of the improvement. It is recorded
as incomplete rather than as landed because a Resolution claiming
otherwise is what the next retro would read.

**Still `origin: self` everywhere, and now demonstrably not independent.**
All three RFCs come from this repository, two of them from the *same run*
(`1786121603-ri`) — RFC 0018 says so itself and holds `runs_observed` at
1. Their `model` fields are withheld under the harness policy tracked by
`rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md`, so
same-agent authorship can be neither confirmed nor excluded, and each
Resolution says that rather than implying independence. Four proposals
across the three ratify decisions their own author made mid-run. The
accumulation model has now processed 56 proposals across five runs
without a single reporter who does not share the triager's assumptions.

**2026-08-17 — partial triage of the orchestration inbox.** 8 RFCs across
`bead-bonanza` (6), `plan-sprint` (1) and `recover-sprint` (1); 37
proposals classified, **8 adopted, 0 deferred, 0 rejected, 3 closed as
already-implemented, 26 left undecided.** No table rows are added yet:
every one of the 8 RFCs still holds undecided proposals, so a row carrying
per-proposal counts would read as a completed triage. **Rows are owed for
all 8 when the remaining 26 are decided** — the interim outcomes are in
each RFC's Resolution, and all 8 stay in `rfcs/inbox/`, unnumbered and
unretired.

*What the provenance check bought.* Three proposals were already fixed
between the version they cite and the tree — the batch-base chaining fix
(landed mid-run by `agent-plugins-61f.2`), per-bead finding numbering, and
the Monitor re-arm rule. Two RFCs cite v1.3 and two cite v1.4 against a
v1.5 tree, which is where all three hits came from. Recorded as *adopted —
already implemented* rather than *rejected — already addressed*, since
each would have been adopted on its merits.

*What the extent and sufficiency checks changed.* Four proposals were
resized rather than accepted as scoped. `rfc-self-wave-chaining` P4 cites
one worktree assertion; there are two. `rfc-self-worktree-composition` P3
names two `Gates:` blocks; there are three. `rfc-self-acceptance-signals`
P2 states that its part 2 can land alone; it cannot — this project's own
test gate is fenced, so part 2 without part 1 would block every pre-flight
here. And `rfc-self-report-citations` P3 has no defect site at all: the
pattern it fixes appears in an orchestrator-authored prompt, not in any
skill text.

*The dedupe result was the opposite of the expectation.* One genuine
duplicate in 37 proposals. The RFCs read each other and routed around
overlap deliberately — `rfc-self-wave-chaining` P3 and
`rfc-self-worktree-composition` P6 close opposite directions of one gap
and say so; `rfc-self-wave-chaining` declines to re-propose the
drained-run trigger because a sibling already owns it. The one real
convergence is a *cross-skill* one that no single RFC could have seen:
`rfc-self-push-scope` P2 and `rfc-self-worktree-absence` P1 are the same
check written into two skills, and were adopted together with each naming
the other so they cannot drift.

*The tracker was empty at triage.* `bd list` across every status returned
0 issues and `bd label list-all` found no `theme:` keys, so dedupe against
prior deferrals was impossible and ran against the 13 retired RFCs
instead. No deferral bead could be filed for the 4 recommended deferrals
either. Whatever is wrong with the Dolt bootstrap, the accumulation
signal this pipeline depends on is currently unavailable — that is a
prerequisite for the next triage, not a footnote.

*Self-authorship, unchanged and now unambiguous for six of eight.* All 8
carry `origin: self`. Six name `claude-opus-5`, which is the model that
triaged them — **authored and triaged by the same model**, stated in each
Resolution rather than implied. Two withhold the model under the harness
policy. Six of the eight adopted proposals ratify a workaround their own
author improvised during the run being described. The accumulation model
has now processed 93 proposals without one reporter outside this
repository.

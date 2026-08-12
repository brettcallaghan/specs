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

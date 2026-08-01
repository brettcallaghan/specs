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

# Run `agent-plugins-bguq` — Seven swarm-dispatch defects PR #33 ships

Seven defects in code PR #33 itself creates, selected by the human from the
previous run's retro candidates and fixed on the same branch. Written by
`sprint-retro` as `retro-bguq`.

**29 findings harvested — 26 in flight across 7 beads, 3 by this retro. 7/7
admitted beads closed, 7 commits, nothing stranded, 13/13 declared checks green
at the merged head (re-measured, not relayed). 0 findings unrouted. 8 candidates
drafted — 0 filed, 0 declined, 8 AWAITING THE OWNER'S DECISION.**

## Read this first

### 1. Two P2 review findings want your eye BEFORE this PR merges

Review pass 1 filed 9 findings and worked none of them —
`standing_auth.admit_review_findings` is `false`, and the coordinator held to
it. Seven are in scope (defects this run's own diff introduced). Two of those
seven are the ones that matter at merge time:

- **`agent-plugins-bi7q` (P2)** — `agent-plugins-f9lq`'s new
  `invocationsRequired` guard returns a plain `*Refusal`, which `acceptDispatch`
  answers with `return Result{}, err` **after `git merge --ff-only` has already
  landed**, discarding `merged:true` and `merged_sha`. That is the exact loss
  `4ca8` and `j5ah` were filed to prevent, reintroduced by their own fix's
  sibling. It is bounded — this run's own `citation_test.go` header records
  three measured guards making the path unreachable today — but **the comment
  justifying the guard is false**, and a comment is what a future reader acts
  on.
- **`agent-plugins-j2jf` (P2)** — `agent-plugins-u9uh` made `--threshold < 1`
  fatal, while `sf.toml`, `sprint-scope/SKILL.md` and
  `bead-bonanza/references/dispatching.md` all still describe
  `dispatch_failure_threshold` as merely defaulting to 3, and the **same
  `sf.toml` `[run]` block** documents `review_budget` with a `0`-means-off
  idiom. A project that follows that idiom and sets `0` now fails at every
  dispatch return, mid-run, past the single human gate.

Both are filed and open. Neither is re-proposed as a candidate below.

### 2. Four of the seven work orders carried a claim that was false against the tree

Not a small thing, and the run's single largest signal. Each was caught by the
agent that received it, measured, and recorded rather than obeyed:

| bead | what the order said | what the tree says |
|---|---|---|
| `74xc` | AC 1: "for every combination … dry-run and execute agree on the stop verdict" | contradicted by `TestADryRunBelowTheThresholdStaysClear`, already committed, which pins the disagreement deliberately |
| `f9lq` | site 4's never-ran path yields `[][]string{nil}` | `tracker.ClosableFrom` refuses a nil `Value` two lines earlier, so that path was never reachable |
| `f9lq` | AC: "a unit test covers each of the five sites on the never-ran-record path" | unsatisfiable for sites 1–3, blocked by three independent guards |
| `4ca8` | Steps to Reproduce: "arrange the merged-head read to be skipped or absent" | unreachable through the real `gitvcs.VersionControl`; the test needed a substituted VC |

Every one of those bead texts was written by the coordinator from a prior
review or retro finding, and none was measured against the tree before it was
dispatched. `AGENTS.md`'s "tool claims are measured at the point of use" binds
a bead's own Steps to Reproduce and acceptance criteria, and nothing checks it.
That is candidate **C4**.

### 3. The retro's own documented harvest query returns nothing, and so does its documented replacement

Third run in a row this has cost a retro time. Measured here, bd 1.2.2
(`6c124203e`):

```
$ bd export | head -1 | jq -r 'keys[]' | tr '\n' ' '
_type assignee close_reason closed_at comment_count created_at created_by
dependencies dependency_count dependent_count description id issue_type labels
owner priority status title updated_at

$ bd show agent-plugins-jed1 --json | jq -r '.[0] | keys[]' | tr '\n' ' '
acceptance_criteria assignee closed_at comment_count created_at created_by
dependencies dependency_count dependent_count description design id issue_type
labels metadata owner parent priority started_at status title updated_at

$ bd show agent-plugins-jed1 --json | jq '.[0].comments | length'
0
```

`sprint-retro` §2 and `findings-capture-protocol.md` both prescribe
`bd export | jq '… .comments[]? …'`. **There is no `comments` key on either
form** — only `comment_count`. The dispatch prompt that opened this phase
carried the fix as "rebuild the harvest per-bead with `bd show <id> --json`",
and that is **also wrong** on this version, for the same reason.

The form that works:

```
$ bd comments agent-plugins-jed1 --json | jq 'length'
5
```

The harvest below was rebuilt on `bd comments <id> --json` and reconciled
against the count the beads themselves record — `sum(comment_count)` over the
seven admitted beads is 26, and 26 findings were harvested. That reconciliation
is what `FEAT-SWARM-RUN-007` specifies (`swarm finding harvest`), and no
`finding` command exists in `cli/swarm/cli` yet. Candidate **C1**.

## The run

| | |
|---|---|
| run epic | `agent-plugins-bguq` — "Run: fix the five swarm-dispatch defects PR #33 ships" |
| branch | `claude/swarm-dispatch-blueprint-5gob2e`, PR #33 (open, not draft, head `36598cd`) |
| range | `e9e36fa..36598cd` — 7 commits, 14 files, +1117/−35 |
| queue | 7 admitted: 5 at scope, 2 admitted mid-run by explicit human instruction |
| dispatches | 3 (`dev-bguq-1`, `dev-bguq-2`, `dev-bguq-3`), max parallelism 1 |
| circuit breaker | `consecutive_dispatch_failures: 0` — never advanced |
| review | 1 pass of a budget of 2, by the DECLARED reviewer (`sf.toml` `[review] skill = "code-review"`) |
| declared checks | 13: `format vet build test-go skills shared-refs strata-docs schemas corpus ref-links template-drift script-self-tests skills-strict` |
| checks baseline | exit 0, everything passed |
| collision edges | `.beads/plans/bguq.edges.jsonl` — 8 edges, one serial chain plus two tails |
| concurrency cap | 20 (irrelevant: the collision graph held parallelism to 1) |
| model lineup | **NOT RECORDED.** `manifest.models` and `manifest.effort` are both absent — see "What the run could NOT verify" and candidate C6 |

`base_sha` is the **run's** start, not the PR's. `git merge-base origin/main HEAD`
is `67be16f`, where PR #33 began; everything between `67be16f` and `e9e36fa`
belongs to the earlier completed run `agent-plugins-e6xt`. A review or a retro
scoped to the merge-base would have read ~15,000 lines of another run's work as
this one's.

All seven `standing_auth` rows resolved to the **skill default** and were
confirmed at the single human gate:
`pull_request=draft` (moot — PR already open),
`review_budget=2`, `tier_escalation=once`, `merge_to_default=false`,
`move_foreign_beads=false` (n/a — all five original beads were parentless and
adopted, not moved), `admit_review_findings=false`,
`dispatch_failure_threshold=3`.

## Results

7 of 7 admitted beads closed. Zero acceptance failures, zero bounces, zero tier
escalations. Two dispatches recorded their gate run as a finding — `f9lq#6` and
`u9uh#4`, each `bb-checks.py` at exit 0 with all thirteen checks present. The
remaining per-bead gate runs are the coordinator's account and were not
re-measured here; the merged head was, below.

| commit | bead | P | tier | agent | what it did |
|---|---|---|---|---|---|
| `9c20218` | `4ca8` | P1 | standard | `dev-bguq-1` | `foldIn` cites the post-fold resolve's own step declaration, so `merged_sha` names an invocation and `merged:true` survives |
| `ad8ddb5` | `j5ah` | P2 | standard | `dev-bguq-1` | a failed post-fold resolve takes `OutcomeAttention` with a resolution, and the two halves' resolutions now compose instead of one replacing the other |
| `658704b` | `74xc` | P1 | standard | `dev-bguq-1` | the dry-run breaker verdict follows the reset a positive `--completed` settles, without inventing a stop from a projection |
| `bec6637` | `n3tn` | P3 | light | `dev-bguq-1` | "the comparison did not run" gets its own resolution, proposing a read rather than a write |
| `d71200e` | `jed1` | P1 | advanced | `dev-bguq-2` | new `dispatch-concurrent-worktree` e2e fixture + a Change Guidance rule for fixture realism |
| `82d575f` | `f9lq` | P2 | standard | `dev-bguq-3` | five more fact sites cite a re-runnable invocation or refuse, behind one `invocationsRequired` seam |
| `36598cd` | `u9uh` | P3 | light | `dev-bguq-3` | `readyCmd` refuses `--threshold < 1` and `--completed < 0` before reading or writing anything |

The chain order departs from strict priority — `4ca8` (P1) → `j5ah` (P2) →
`74xc` (P1) → `n3tn` (P3) — deliberately, and the epic records why: within a
single-assignee chain the order is a work order for one agent, not a
parallelism decision, and `4ca8` and `j5ah` sit five lines apart in `foldIn`.

## Scope

The manifest records no `excluded` and no `moved` entries: this run's queue was
handed to it as a list, not resolved from a label or a priority, so nothing was
turned away at admission. Three scope movements did happen, and all three are
worth reading.

**Scope grew 5 → 7 mid-run, by human instruction, not by the run widening
itself.** `f9lq` and `u9uh` were both filed by `dev-bguq-1` as
`discovered-from` while working `4ca8` and `74xc`, and both were deliberately
left OUT at the time — the agent was right not to widen a closed scope
unilaterally. The human admitted them explicitly afterwards. This is **not**
the `admit_review_findings` row: that row governs findings from `sprint-review`,
which had not yet run, and it stayed `false` through the review phase.

**Seven sibling beads from the previous run stayed deferred by decision and
were never admitted**: `whlh`, `prnv`, `bdwi`, `p5w9`, `ypw2`, `pok6`, `l0ac`.

**One follow-up escaped into the run by construction, and clearing its parent
did not get it out.** `dev-bguq-3`, working `f9lq`, filed its overflow as
`agent-plugins-bguq.4.1` — a child of the run's TARGET epic, and therefore run
work by id alone. Clearing the parent pointer was not enough: run membership
resolves through the **id hierarchy**, so `swarm ready` kept offering it. It
was refiled outside the run as `agent-plugins-dbvf` and `bguq.4.1` closed as
superseded. This is a reusable fact about the tracker's model, not a one-off —
candidate **C3**.

## What the run decided

Sixteen of the 26 findings are the six work-order kinds; ten are `proof`. The
decisions below are the ones that changed what *correct* meant for a bead.

**`j5ah` widened its own change by one edit the order did not name.** `foldIn`
returns a `mergeOutcome` that carried no verdict field, and `acceptDispatch`
read only its facts and summary — so making the outcome *reach* the result took
three parts, not one. The third is a decision rather than plumbing: the settle
block's `result.Resolution` assignment became an **append**, because a fold
reporting attention followed by a removal half doing the same would otherwise
have had the fold's resolution silently replaced — "the same defect this bead
fixes, moved one line down".

**`4ca8` left an unguarded index in place, deliberately, and said what it would
cost.** `landedSteps[len(landedSteps)-1]` panics on an empty fragment *after*
an irreversible fast-forward. The agent declined to guard it on three measured
grounds (`vcs.Steps` is a contract that forbids the empty case; the only
implementation returns exactly one step; the same unguarded index is the
established pattern at five other sites) and named the rejected alternative's
cost: a fallback to `m.caseInv` would cite the **pre-fold** case read as the
invocation that established a **post-fold** sha — a wrong citation, which is
worse than a panic on a state the contract forbids.

**`74xc` chose a reading of a false AC rather than stopping for an
amendment.** AC 1 as written would have required deleting a committed test that
pins deliberate behaviour. The agent implemented the bead's Design instead, and
recorded the reading it took: *the verdict follows whichever value execute would
compare, whenever that value is one the report settles rather than one the
write would have invented.* It also named what would settle the question and
said plainly that it did not stop for the answer.

**`jed1` declared two worktrees because the AC's conditions cannot both hold
with one.** Condition 3 wants the target ahead of the dispatch branch;
condition 4 wants a worktree at the sprint head with no commit of its own. One
tree cannot be both. Two trees also model the real world better — one dispatch
that has committed and one that has not are live at the same time.

**`jed1` also diverged from its two sibling fixtures on purpose**, setting the
new fixture's `admitted` to all five beads where `dispatch-worktree` and
`dispatch-stale-base` leave it at the core's two. Following precedent would
have put a second, unchosen anomaly (`compareScope` reporting scope drift) into
a world built to assert about three specific ones — "a fixture modelling drift
it did not mean to model is the exact defect the Change Guidance rule added by
this bead condemns".

**`f9lq` consolidated three refusal sites behind one helper so the guard would
have a seam a test could reach.** Written inline, as the Design's remedy
implied, the guard would have had **no test at all**, because the never-ran path
at all three sites is unreachable end to end. `invocationsRequired` is the
result — and it is also the subject of `agent-plugins-bi7q`, above.

**`u9uh` placed its guards after the mode-resolving switch, not inside it**,
because the Design's "after the mutual-exclusion check and before the mode is
resolved" names a point that does not exist: `ready.go` does both in one
`switch`. Placing them before the switch would have made a threshold refusal
preempt the `--explain`/`--dry-run` one, changing which of two correct refusals
a caller sees for no gain.

### The gaps and what they say about the orders

All three `gap` findings are the same shape: **the bead described a state that
cannot be reached from outside the function under test.**

- `4ca8#2` — neither "skipped" nor "absent" is reachable through
  `gitvcs.VersionControl`; the test substitutes a VC. "Nothing in the bead said
  which altitude was required."
- `n3tn#2` — the "comparison did not run" branch is unreachable in *either*
  mode, not just dry run, so the test calls `landingVerdict` directly. The
  finding names the property that made that possible (it takes plain values and
  returns a verdict, with no runner and no I/O) and flags that a refactor giving
  it a runner would remove the only way to test it.
- `4ca8#5` — the coordinator's own, recorded **after** the bead closed and
  merged: `invocationOf` returns `nil` for a step carrying a reference argument
  (`if a.IsRef() { return nil }`), so the remedy can still yield
  `[][]string{nil}`. The bead's ACs are met and it was deliberately **not**
  reopened; `f9lq`'s design carries the ref-argument case explicitly instead.
  Recorded so the bead's record does not read as a complete solution to a
  problem it only narrows.

### The one `stale` finding is the run's sharpest evidence about grooming

`f9lq#1`: the bead's site-4 analysis quoted `phase.go` correctly and then drew
a conclusion the next two lines refute. `tracker.ClosableFrom` rejects a nil
`Value` before `closableInv` is ever built, so the `[][]string{nil}` the bead
predicted "was therefore not reachable even in principle". The repair still
landed as designed, because the *reachable* half of site 4 is real — refused
`close-N` records in a dry run — and that half now has a negative-controlled
test.

## Review

One pass, by the declared reviewer (`code-review` at effort `high`), over
`e9e36fa..36598cd`. 9 findings, 7 in scope, **0 worked**.

**The loop decision was the coordinator's and it was no.** The budget was
unspent (1 of 2), but budget was not the binding constraint — admission was.
Re-entering the dispatch loop would have worked beads the human's own
authorization says to file, and widened a scope the human had fixed at seven.

| bead | P | row | finding |
|---|---|---|---|
| `bi7q` | P2 | 1 | the new guard's refusal discards the fold's facts after an irreversible merge; its justifying comment is false |
| `j2jf` | P2 | 1 | `--threshold 0` is now fatal; the three surfaces that resolve it never state the minimum |
| `pzlb` | P2 | 1 | the two new argument refusals and the new dry-run verdict have no acceptance criterion |
| `xv5a` | P3 | 1 | `invocationsOf`'s doc comment was captured by `invocationsRequired` |
| `lp8n` | P3 | 1 | `acceptDispatch` appends an empty settle resolution, emitting a dangling `"; "` |
| `eobl` | P3 | 1 | `phase.go`'s two `invocationOf` substitutions rest on a hazard those sites do not have |
| `xrx6` | P3 | 1 | the new fixture-realism SHALL is enforced by nothing |
| `8gzr` | P2 | 3 | a run's manifest can record `branch` without `base_sha` |
| `25hp` | P3 | 2 | `container-swarm-cli.md`'s dry-run breaker paragraph contradicts `applyBreaker` — measured false at `e9e36fa` too, so pre-existing |

**The review only happened because the reviewer refused to guess.**
`manifest.base_sha` was never set by the coordinator, so
`swarm manifest require --run agent-plugins-bguq --key base_sha` exited 1 with
"base_sha is not set", and `sprint-review` step 2 says to review **nothing** in
that case. The reviewer reconstructed the range from its dispatch prompt,
corroborated it against the commits (each of the seven subjects carries exactly
one admitted bead id), labelled its own corroboration *a reviewer's
reconstruction rather than the run's record*, and filed the gap as
`agent-plugins-8gzr`. The key has since been repaired on the manifest; `8gzr`
stays open, because setting it on one run does not stop the next coordinator
omitting it.

The reviewer also recorded a deliberate deviation: `sprint-review` says to file
findings with `--parent {EPIC}`, and it filed them **parentless** with
`--discovered-from` only, because parenting nine open findings to the run epic
would leave it permanently ineligible for `swarm phase close`. That is the same
mechanism candidate C3 is about, arrived at independently from the other side.

## Routing

| branch | in flight | by the retro | total |
|---|---|---|---|
| project audience (`decision` 6, `gap` 4, `scope` 3, `unverifiable` 3, `stale` 1) | 15 | 2 | **17** |
| skill audience (`tooling`) | 0 | 1 | **1** |
| terminating at the record (`proof` 11, `environment` 0) | 10 | 1 | **11** |
| **unrouted** | 0 | 0 | **0** |
| **total** | 26 | 3 | **29** |

**Zero unrouted, and that is the second consecutive clean run** after four runs
at 2–3. It is also not the whole story: **not one `tooling` finding was
recorded by any agent in flight**, on a run that lost its review scoping to a
coordinator error and whose retro's own harvest query does not work. Both of
those *are* tooling defects. Neither was captured as one at the time — the
first landed in an `amend:` comment on the epic, the second in a dispatch
prompt, and neither surface is read by any routing step. The single `tooling`
finding in the table is this retro's, recorded after the fact. That is candidate **C2**, and it is the same hole the invented
`retro:process:` kind was reaching for on three earlier runs.

## Validation

**Baseline** (`manifest.checks_baseline`): exit 0, all 13 checks passed on the
untouched tree.

**At the merged head**, re-measured for this record rather than relayed —
`plugin/skills/bead-bonanza/scripts/bb-checks.py` on a clean tree at `36598cd`:

```
=== format  === vet  === build  === test-go  === skills  === shared-refs
=== strata-docs  === schemas  === corpus  === ref-links  === template-drift
=== script-self-tests  === skills-strict
EXIT=0
```

All 13 declared checks present, zero `FAIL` lines, exit 0. This closes the gap
run `bd97` recorded ("the 13 checks were never run over the merged head").

**CI at head `36598cd`**, measured this session via the checks API — 6 green,
1 red:

| check | conclusion |
|---|---|
| `go test (ubuntu-latest)` | success |
| `go test (macos-latest)` | success |
| `go test (windows-latest)` | **failure** |
| `cross-compile the release matrix` | success |
| `runs with no language runtime installed` | success |
| `skill invariants and vendored references` | success |
| `this repository's own specs conform` | success |

**The Windows failure is not this run's and is not re-filed.**
`agent-plugins-baih` carries the diagnosis and its comments carry the
measurements; the epic records the coordinator's own measurement that `e9e36fa`
— a commit touching exactly one markdown file — failed identically. Two things
there are settled and are restated here only to stop them being restated
wrongly: **the orphan-count "trend" is dead** (2+1 → 0 → 0 → 1 → 2 is noise),
and **the `-timeout 20m` patch is needed and blocked on a workflow-scoped
token**.

What *is* new is that the patch is now load-bearing. The failure at head is a
package timeout, not a scatter of cleanup errors:

```
FAIL	github.com/brettcallaghan/agent-plugins/cli/swarm/e2e	600.094s
```

and `jed1` added a **ninth** e2e fixture to that suite. Locally the same suite
takes 167s; on Windows it is at its 600s ceiling. Candidate **C8** is to
escalate `baih` rather than to file anything new.

## Pushes

Nothing stranded. `git log origin/claude/swarm-dispatch-blueprint-5gob2e..HEAD`
is empty; the branch head and the remote head are both `36598cd`.

**The tracker is durable this run**, which is a change from the last one. Run
`e6xt`'s record opens by saying it is the only durable copy of itself, because
`bd dolt push` failed on every attempt. Measured here, at retro time:

```
$ bd dolt push
Pushing to Dolt remote...
Push complete.
EXIT=0
```

So this run's 26 findings, 7 closures and the beads filed for it are on
`refs/dolt/data`, not only in this container.

## What the run could NOT verify

Named one per line, including the three the agents recorded themselves.

- **Which model actually ran each dispatch.** `manifest.models` and
  `manifest.effort` are both absent — `bd show agent-plugins-bguq --json | jq '.[0].metadata|keys'`
  returns no such key. The beads' declared tiers (`standard`×4, `light`×2,
  `advanced`×1) resolve through `sf.toml` `[models]` to sonnet/haiku/opus, but
  a declared tier is not a record of what served the turn. `sprint-retro`
  requires both values on every filed issue; this run cannot supply them.
  Second consecutive run — `e6xt` recorded "manifest `effort` reads
  `unrecorded`". Candidate C6.
- **The harness's worktree DIRECTORY name** (`jed1#4`). Only the BRANCH form
  (`worktree-agent-<hash>`) is measured, in `status.go`'s own note. No
  `worktree-agent-` tree was standing in the dispatch session to measure the
  path. The fixture marks its two `path` values as chosen rather than measured
  and names the command that would settle it:
  `git worktree list --porcelain` in a session where the harness has created a
  tree. A capture quoted in `agent-plugins-gqwl`'s description suggests
  `.claude/worktrees/agent-<hash>` instead — one quoted capture, not a
  measurement, so it stays unverified.
- **`74xc`'s acceptance criterion 1 as written.** It is false against a
  committed test; the run implemented a narrower reading and recorded it. What
  would settle it: an amendment scoping AC 1 to `--completed N>0`, or a
  statement from its author that the projection rule is being retired.
- **`f9lq`'s acceptance criterion for sites 1–3.** "A unit test covers each of
  the five sites on the never-ran-record path" cannot be satisfied as declared;
  three arrangements were tried and each was blocked by a guard. Sites 1–3 are
  covered at their shared guard and end to end instead, and the unreachability
  is recorded in `citation_test.go`'s own header.
- **The six-consecutive-heads history of the Windows failure.** Restated from
  `agent-plugins-baih`, not re-measured here. What was measured this session is
  the single data point at head `36598cd`, above.
- **Whether the 9 review findings are correct.** They were filed, not worked
  and not independently checked by this retro. `bi7q` and `j2jf` are surfaced
  at the top of this document on the reviewer's evidence, not on mine.
- **The `bd-health.sh` repo-wide failure has no bearing on this run — but it is
  still red.** Measured at retro time:
  `plugin/skills/groom-beads/scripts/bd-health.sh` exits 1, with 80
  `is not an epic or molecule` errors from feeding non-epic ids to
  `bd swarm validate`. Exactly one line in that output names anything of this
  run's, and it is `agent-plugins-dbvf` in a listing rather than a finding.
  Run directly, `bd swarm validate agent-plugins-bguq` exits 0. This is
  deferred bead `agent-plugins-prnv`; treating the script's exit status as
  blocking would stall every run in this repository.

## Candidates — your decision

Eight, ranked. **Nothing below is filed.** Each is one actionable change and is
decidable on its own; accept some, decline others. A declined candidate stays
in this document, and a later run hitting the same friction raises it again
with a higher recurrence count.

`skill_issues_dir` is deliberately unset in `sf.toml` — this repository vends
these skills — so an accepted `audience: skill` candidate is filed as an
ordinary bead here, not written to an outbox.

**Already filed during the run and NOT re-proposed:** `bi7q`, `j2jf`, `8gzr`,
`pzlb`, `xv5a`, `lp8n`, `eobl`, `xrx6`, `25hp` (review pass 1); `eu0f` and
`dbvf` (filed by dispatches in flight). **Deferred by human decision and never
admitted:** `whlh`, `prnv`, `bdwi`, `p5w9`, `ypw2`, `pok6`, `l0ac`.

---

### C1 — `sprint-retro`'s harvest query returns nothing on bd 1.2.2, and so does the workaround people reach for

```yaml
title: "Harvest a run's findings with `bd comments <id> --json`, reconciled against comment_count"
audience: skill
target: sprint-retro (SKILL.md §2) and findings-capture-protocol.md
priority: 1
findings: ["agent-plugins-bguq.6#1"]
recurrence: 1          # run e6xt's candidate C1, never filed
run: agent-plugins-bguq
session: session_01AFy5zV9jVHXpCnmxP6JDY5
branch: claude/swarm-dispatch-blueprint-5gob2e
skill_version: sprint-retro as committed at 36598cd
model: claude-opus-5
```

**What is wrong, and what it costs.** Both documents prescribe
`bd export | jq '… .comments[]? …'`. `bd export` on bd 1.2.2 emits no
`comments` key at all — only `comment_count` — so the prescribed harvest
returns **zero findings on a fully drained run**. The skill's own guard ("a
zero count across a DRAINED run is a broken query") tells the reader to check
scope and prefix spelling, which are not the cause, so the guard sends them the
wrong way. The obvious second guess is `bd show <id> --json`, which the skill
explicitly rejects for the right reason but which is **also** empty of comments;
the dispatch prompt for this very phase carried it as the fix. Cost, measured
across three runs: every retro either reports a clean run that was not clean, or
spends its opening minutes re-deriving the query. This run spent them.

**Evidence.**

- `agent-plugins-bguq.6#1` — "The documented harvest query returns nothing on
  bd 1.2.2, and the replacement carried in this phase's own dispatch prompt is
  wrong the same way."
- Recurrence: run `e6xt`'s retro, candidate C1, "`sprint-retro`'s documented
  harvest query returns zero findings on bd 1.2.2" — drafted, never filed.

**Proposed change.** In `sprint-retro/SKILL.md` §2 and the "Harvesting" section
of `findings-capture-protocol.md`, replace the `bd export` pipeline with the
per-bead form and a reconciliation:

```bash
for id in <the run's closure set>; do
  bd comments "$id" --json | jq -r --arg id "$id" \
    'to_entries[] | select(.value.text|startswith("retro:"))
     | "\($id)#\(.key+1)\t\(.value.author)\t\(.value.text)"'
done
```

and require the harvest to be reconciled against
`bd show <id> --json | jq '.[0].comment_count'` summed over the same set,
failing loudly on a mismatch rather than only on zero. Add one sentence naming
`bd export`'s and `bd show --json`'s missing `comments` key as the reason, so
neither is re-adopted. Done when a retro run against a drained run harvests the
same count the beads report.

---

### C2 — Coordinator-side friction has no finding form, so it lands in `amend:` comments that no routing step reads

```yaml
title: "Give the coordinator a capture form, so a run's own tooling friction routes"
audience: skill
target: findings-capture-protocol.md, bead-bonanza
priority: 2
findings: ["agent-plugins-bguq.6#1", "agent-plugins-bguq.6#2"]
recurrence: 3          # so39, mpnf and bd97 each recorded the invented
                       # `retro:process:` kind as unrouted
run: agent-plugins-bguq
session: session_01AFy5zV9jVHXpCnmxP6JDY5
branch: claude/swarm-dispatch-blueprint-5gob2e
skill_version: findings-capture-protocol.md as committed at 36598cd
model: claude-opus-5
```

**What is wrong, and what it costs.** The nine kinds are written for "an
implementing agent" noticing something about "the work it was given". A
coordinator has no such work order, so when it hits friction it has nowhere to
put it. This run's two real tooling defects were both coordinator-side — the
omitted `manifest.base_sha` that disabled the review's scoping step, and the
harvest query — and **neither was captured as a `tooling` finding**. The first
went into an `amend:` comment on the epic; the second into a dispatch prompt.
Routing reads neither. The result is a run that reports **0 unrouted and 0
tooling findings** while having lost two, which makes the health metric read
clean at exactly the moment it should not. Three earlier runs invented
`retro:process:` reaching for the same hole and had it counted as unrouted each
time.

**Evidence.**

- `agent-plugins-bguq.6#1` — the harvest defect, recorded as a `tooling`
  finding only after the fact, by the retro.
- `agent-plugins-bguq.6#2` — the manifest's missing `models`/`effort`, likewise.
- The epic's own `amend: tooling` comment: "the coordinator did not set
  `manifest.base_sha`, and it broke the review's scoping step. MY ERROR, not
  the reviewer's."
- Recurrence: `so39` (2 unrouted, both `retro:process:`), `mpnf` (2), `bd97`
  (3, "all 3 are coordinator-authored `retro:process:`/`retro:work-clean`").

**Proposed change.** State in `findings-capture-protocol.md` that the protocol
binds every agent in a run, coordinator included, and that a coordinator
records against the **run epic** (or the ceremony phase bead it is executing)
rather than against a work order. Add one line to `bead-bonanza`'s ceremony
saying that an `amend:` comment recording friction MUST be accompanied by a
`retro:<kind>` finding, because `amend:` is a decision record and is not
harvested. Done when a coordinator error of the `base_sha` shape appears in a
retro's routing table rather than only in its prose.

---

### C3 — A follow-up filed under the run's target is run work by id, and clearing its parent does not release it

```yaml
title: "A dispatch files its overflow parentless: run membership resolves through the id hierarchy"
audience: skill
target: run-batch, sprint-review
priority: 2
findings: ["agent-plugins-f9lq#5"]
recurrence: 0
run: agent-plugins-bguq
session: session_01AFy5zV9jVHXpCnmxP6JDY5
branch: claude/swarm-dispatch-blueprint-5gob2e
skill_version: run-batch as committed at 36598cd
model: unrecorded — manifest carries no `models`; `f9lq` declared tier `standard`
```

**What is wrong, and what it costs.** `run-batch` tells a dispatch to file what
it finds and stay in scope, and says nothing about **where** to file it.
`dev-bguq-3` filed `f9lq`'s overflow as `agent-plugins-bguq.4.1` — a child of
the run's TARGET epic — which makes it run work by construction, admitted by
nobody. The obvious remedy failed: clearing the parent pointer did not remove
it, because run membership resolves through the **id hierarchy**, and
`swarm ready` kept offering it. It had to be refiled outside the run as
`agent-plugins-dbvf` and the original closed as superseded. The cost is a bead
that widens a scope the human fixed, a wasted refile, and — if nobody notices —
a dispatch picking up work outside the authorization gate. The reviewer hit the
same mechanism from the other side the same afternoon and deviated from
`sprint-review`'s written instruction to avoid it.

**Evidence.**

- `agent-plugins-f9lq#5` — "The bead's five sites do not exhaust the shape it
  describes. Filed as `agent-plugins-bguq.4.1` rather than widened into this
  bead."
- The review phase's own deviation note: "`sprint-review` says
  `bd create ... --parent {EPIC}`. Findings were filed PARENTLESS with
  `--discovered-from` only … Parenting nine open findings to the run epic would
  leave it permanently ineligible for `swarm phase close`."

**Proposed change.** In `run-batch`'s findings section, state the rule
explicitly: *a bead filed from inside a dispatch is filed **parentless**, with
`--discovered-from <the bead in hand>` and no `--parent`. An id of the form
`<run-epic>.<n>.<m>` is inside the run whatever its parent pointer says, and
clearing the pointer does not remove it — it must be refiled under a fresh id.*
Correct `sprint-review`'s `--parent {EPIC}` instruction to match, and give it
the second reason (open children block `swarm phase close`). Done when both
skills say parentless and neither says `--parent`.

---

### C4 — Four of seven work orders carried a claim that was false against the tree, and nothing checks a bead's own evidence

```yaml
title: "A bead's Steps to Reproduce and acceptance criteria are measured against the tree before it is Ready"
audience: skill
target: groom-beads (Definition of Ready), sprint-plan
priority: 1
findings: ["agent-plugins-74xc#2", "agent-plugins-f9lq#1", "agent-plugins-f9lq#2", "agent-plugins-4ca8#2"]
recurrence: 0
run: agent-plugins-bguq
session: session_01AFy5zV9jVHXpCnmxP6JDY5
branch: claude/swarm-dispatch-blueprint-5gob2e
skill_version: groom-beads as committed at 36598cd
model: unrecorded — manifest carries no `models`; declared tiers `standard` ×3
```

**What is wrong, and what it costs.** The Definition of Ready checks that a
bead is self-sufficient — that it names files, a remedy and acceptance criteria
— and never checks that what it asserts is **true**. On a run of seven, four
work orders shipped with a false or unsatisfiable premise: an AC contradicted
by a committed test (`74xc`), a code analysis refuted two lines below the quote
it took (`f9lq` site 4), an AC unsatisfiable for three of its five sites
(`f9lq`), and a reproduction path unreachable through the real implementation
(`4ca8`). Every one was authored from a prior review or retro finding and
inherited rather than measured — the precise defect `AGENTS.md` names under
"Tool claims are measured at the point of use", which it also admits is
"enforced by authorial memory". The cost this run was absorbed by good agents:
each measured, chose a reading, and recorded it. That is luck, not a mechanism.
A cheaper agent obeys the order, deletes the test AC 1 contradicts, and the
regression guard is gone.

**Evidence.**

- `agent-plugins-74xc#2` — "acceptance criterion 1 is false as written, and
  satisfying it literally would delete a committed test that pins deliberate
  behaviour".
- `agent-plugins-f9lq#1` — "The bead's site-4 analysis is incomplete: an
  earlier guard already blocked the never-ran path it describes."
- `agent-plugins-f9lq#2` — "cannot be satisfied as declared for sites 1, 2 and
  3: that path is unreachable by construction, by three independent guards."
- `agent-plugins-4ca8#2` — "the bead's Steps to Reproduce cannot be carried out
  through the real Git Version Control".

**Proposed change.** Add one row to `groom-beads`' Definition of Ready: *every
claim a bead makes about the tree — a quoted symbol, a reproduction path, an
acceptance criterion asserting current behaviour — is run or read at grooming
time, and the bead records the command and its output beside the claim. A claim
that could not be checked is marked `unverified` in the bead and names what
would settle it.* Where the bead is derived from a prior finding, the grooming
pass re-measures rather than citing the finding. Done when a groomed bead
carries its own evidence line and a dispatch can cite it without re-deriving it.

---

### C5 — One C4 component spans a whole directory, so any two beads in it collide and the run serializes

```yaml
title: "A component-level collision yields to a measured file-and-symbol disjointness"
audience: skill
target: sprint-plan (partitioning), specs/architecture/container-swarm-cli.md
priority: 2
findings: []           # the epic's `amend: sequencing` comment; no in-flight finding
recurrence: 1          # bd97's collision mechanism, 17 waves at max width 2
run: agent-plugins-bguq
session: session_01AFy5zV9jVHXpCnmxP6JDY5
branch: claude/swarm-dispatch-blueprint-5gob2e
skill_version: sprint-plan as committed at 36598cd
model: claude-opus-5
```

**What is wrong, and what it costs.** `component-swarm-command-surface` is
declared as the whole command surface, and its implementation is the whole of
`cli/swarm/cli` — `ls cli/swarm/cli/*.go | wc -l` returns 30 today (27 when
this run was planned). Any two beads touching that directory therefore carry
the same `c4_changes` entry and collide, whatever they actually edit. Here
`74xc` (`ready.go`, `applyBreaker`) and `n3tn` (`branch.go`, `landingVerdict`)
are disjoint by file, by function and by modified symbol, and merged
`--ff-only` cleanly — and were still serialized. The union rule was followed
and parallelism 1 was accepted deliberately, which was right on four S/XS
beads. It will not be right on twenty: run `bd97` serialized 20 beads into 17
waves at max width 2, fourteen of them width 1, for the same reason.

**Evidence.**

- The epic's `amend: sequencing` comment: "File scan says DISJOINT … C4 says
  COLLIDE … The cause is granularity — that one component spans all 27 files of
  `cli/swarm/cli` — so it OVER-detects here."
- Recurrence: run `bd97`'s retro — "a collision mechanism that serialized 20
  beads into 17 waves at max width 2, 14 of them width 1 (`221f`, `tvom`)".

**Proposed change.** In `sprint-plan`'s partitioning rule, keep the union as
the default and add one escape: *where two beads share a component but a file
scan finds no shared file and no shared modified symbol, the component edge is
recorded as **advisory** and does not serialize them; the plan states which
edges were downgraded and on what scan, so the decision is auditable.* Propose
one change only — do not also re-cut the component blueprint here; if a finer
component boundary is wanted that is a separate issue against
`container-swarm-cli.md`. Done when a plan over two file-disjoint beads in one
component reports parallelism 2 with the downgrade recorded.

---

### C6 — The run manifest carries no model lineup and no effort, and every filed issue needs both

```yaml
title: "The manifest records `models` and `effort`, or the skills stop requiring them"
audience: skill
target: sprint-scope (manifest write), sprint-retro (close the loop)
priority: 3
findings: ["agent-plugins-bguq.6#2"]
recurrence: 1          # e6xt recorded "manifest `effort` reads `unrecorded`"
run: agent-plugins-bguq
session: session_01AFy5zV9jVHXpCnmxP6JDY5
branch: claude/swarm-dispatch-blueprint-5gob2e
skill_version: sprint-scope as committed at 36598cd
model: claude-opus-5
```

**What is wrong, and what it costs.** `sprint-retro`'s close-the-loop step says
every issue carries the run's model lineup and effort "from the manifest",
"which records them precisely so this step can". This run's manifest records
neither: its keys are `admitted`, `base_sha`, `branch`, `checks_baseline`,
`checks_ids`, `collision_edges`, `concurrency_cap`,
`consecutive_dispatch_failures`, `review_passes` and `standing_auth`. So all
eight candidates above ship with `model:` unfillable — and the issue template
calls that field load-bearing precisely because findings differ by capability
tier. The declared per-bead `tier` is recoverable and is *not* a substitute: a
tier is a dispatch instruction, and `tier_escalation=once` permits what ran to
differ from what was asked for.

**Evidence.**

- `agent-plugins-bguq.6#2` — "The run manifest records no model lineup and no
  effort, and `sprint-retro`'s close-the-loop step requires both on every issue
  it files."
- Recurrence: run `e6xt`'s retro, under what it could not verify — "manifest
  `effort` reads `unrecorded`".

**Proposed change.** Make `sprint-scope` write `models` and `effort` onto the
manifest at the authorization gate, alongside `standing_auth`, recording the
resolved value **and its source** the way the seven gate rows already do; and
have the coordinator update `models` when a `tier_escalation` actually fires,
so the key records what ran rather than what was planned. Done when
`swarm manifest get --run <epic> --key models --key effort` answers on a fresh
run.

---

### C7 — PR #33's description asserts a CI result that is false at its own head

```yaml
title: "Add the pull request description to the surfaces that carry a copy of a belief"
audience: project
target: AGENTS.md ("Surfaces that carry a copy of a belief"), PR #33's body
priority: 3
findings: ["agent-plugins-bguq.6#3"]
recurrence: 0
run: agent-plugins-bguq
session: session_01AFy5zV9jVHXpCnmxP6JDY5
branch: claude/swarm-dispatch-blueprint-5gob2e
skill_version: n/a
model: claude-opus-5
```

**What is wrong, and what it costs.** PR #33's body says "All 13 declared
checks pass, and CI passes on ubuntu, macOS and Windows plus the cross-compile
matrix." At its head `36598cd`, `go test (windows-latest)` is `failure`. The
claim was true when written by run `e6xt` and is now stale, and nothing sweeps
it: `AGENTS.md`'s table of "surfaces that carry a copy of a belief" lists
durable trigger prompts, dispatch prompts, outbox issues, bead descriptions,
epic comments, plan documents, committed documents, generated references and
this repository's own scripts — and **not the pull request description**, which
is the surface a human reads immediately before merging. The cost is the one
the table exists to prevent: a reviewer takes a green summary from the PR body
instead of the checks tab.

**Evidence.**

- `agent-plugins-bguq.6#3` — CI at head `36598cd`: ubuntu, macOS,
  cross-compile, no-runtime, skills and specs `success`;
  `go test (windows-latest)` `failure`.
- PR #33 body, "Verification, stated honestly": "CI passes on ubuntu, macOS and
  Windows plus the cross-compile matrix."

**Proposed change.** Add one row to the `AGENTS.md` table — *a pull request
description: lives until the PR merges; re-read by every reviewer and by
whoever merges; audited by nobody, and a run that appends to a long-lived PR
does not re-read what earlier runs claimed in it* — and one line to the
retraction sweep's order of reach, placing it with beads and epic comments.
Correcting PR #33's body itself is the immediate consequence, not the issue.

---

### C8 — Escalate `agent-plugins-baih`: this run made its blocked patch load-bearing

```yaml
title: "Raise agent-plugins-baih: the e2e suite now exceeds the Windows 600s package timeout"
audience: project
target: agent-plugins-baih (existing bead — escalate, do not re-file)
priority: 1
findings: ["agent-plugins-bguq.6#3"]
recurrence: 0
run: agent-plugins-bguq
session: session_01AFy5zV9jVHXpCnmxP6JDY5
branch: claude/swarm-dispatch-blueprint-5gob2e
skill_version: n/a
model: claude-opus-5
```

**What is wrong, and what it costs.** The Windows CI leg is not this run's
defect and is fully diagnosed on `agent-plugins-baih`, whose remedy — a
`-timeout 20m` patch to the workflow — is blocked on a workflow-scoped token.
That block was tolerable while the failure was a scatter of `TempDir`
`RemoveAll` errors. At head it is a **package timeout** (`FAIL … 600.094s`),
and `jed1` added a **ninth** e2e fixture to that suite this run. The same suite
takes 167s on Linux here. So the run has moved a bounded, understood failure
closer to a permanent red leg, and the unblocking step is administrative, not
technical. The cost of leaving it: `go test (windows-latest)` stays red on
every future head, and a red leg that is always red stops being read.

**Evidence.**

- `agent-plugins-bguq.6#3` — job `100339119103` log ends
  `FAIL github.com/brettcallaghan/agent-plugins/cli/swarm/e2e 600.094s`, five
  orphan processes terminated at cleanup.
- `agent-plugins-jed1` (`d71200e`) — adds the ninth e2e fixture variant,
  `dispatch-concurrent-worktree`.

**Proposed change.** Raise `agent-plugins-baih` to P1, add a comment recording
the timeout measurement at `36598cd` and the ninth-fixture change, and state
the one blocking action a human must take (grant a workflow-scoped token, or
apply the `-timeout 20m` change by hand). **Do not file a new bead.** Done when
`baih` names the administrative unblock as its next step and carries the head
measurement.

---

## What was filed

**Nothing.** All eight candidates await the owner's decision. Counts as they
stand: 29 findings harvested, 17 routed to the project audience, 1 to the skill
audience, 11 terminating at this record, 0 unrouted; 8 candidates drafted,
0 filed, 0 declined.

Filing an accepted candidate also means recording what it rests on, on the bead
that carried the finding:

```bash
bd comment <finding's bead> "retro:filed:<new-id> — <finding id>"
```

Recurrence counts unfiled findings only, so that comment is what stops a filed
finding being counted again next run.

## Every finding, verbatim

Grouped by kind, ordered as the taxonomy orders them. Findings `#1`–`#N` are
by position within their own bead. `agent-plugins-bguq.6#*` are this retro's
own, recorded on the retro phase bead.

### `decision` — 6

**`agent-plugins-4ca8#3`** · `dev-bguq-1`

the fix indexes landedSteps[len(landedSteps)-1] unguarded, as the bead's Design prescribes; an empty fragment would panic after an irreversible fast-forward

```
evidence: the installed expression is

  rec, _ := report.Record(landedSteps.Answer())
  landedInv := [][]string{invocationOf(landedSteps[len(landedSteps)-1], rec)}

An empty landedSteps panics there, and it would panic AFTER the fast-forward has landed — strictly
worse than the bug being fixed. I considered guarding it and did not, for three measured reasons:

1. vcs.Steps is a contract, not a convention: "an ordered fragment of a plan whose LAST step's value
   is the fragment's answer" (cli/swarm/plan/plan.go, `// Steps is an ordered fragment`). An empty
   fragment already breaks Steps.Answer() (returns "") and revision(), which foldIn calls two lines
   later.
2. The only implementation returns exactly one step:
     $ grep -n "func (v VersionControl) Resolve" -A 6 cli/swarm/vcs/git/git.go
     132:func (v VersionControl) Resolve(prefix, revision string) vcs.Steps {
     ...
     136:	return vcs.Steps{s}
3. The same unguarded index is the established in-repo pattern at five other sites:
     $ grep -n "len(.*Steps)-1\]" cli/swarm/cli/*.go
     cli/swarm/cli/accept_settle.go:245, cli/swarm/cli/branch.go:250, :385, :402,
     cli/swarm/cli/dispatch.go:406, :479
   Guarding one of six would be inconsistent noise and would add a branch nothing exercises.

Rejected alternative: `if len(landedSteps) > 0 { ... }` with a fallback to m.caseInv. Cost: it cites
the PRE-fold case read as the invocation that established a POST-fold sha, which is the exact
attribution error the comment above this block exists to prevent — a wrong citation is worse than a
panic on a state the contract forbids.
```

**`agent-plugins-j5ah#2`** · `dev-bguq-1`

the bead scoped the change to foldIn's lerr arm; making the outcome actually REACH the result required a second edit the bead did not mention, in acceptDispatch's composition of the two halves

```
evidence: foldIn returns a mergeOutcome, and acceptDispatch reads only its facts and summary:

    result.Facts = append(result.Facts, merged.facts...)
    result.Summary = merged.summary

There was no field to carry a verdict, and no line to apply one. So the change is three parts, not
one:

1. mergeOutcome gains `outcome Outcome` and `resolution string`, mirroring the `settlement` struct
   in cli/swarm/cli/accept_settle.go, which already carries exactly those two for the other half.
2. acceptDispatch applies them with the same `if merged.outcome != OutcomeClear` shape the settle
   block below it already uses.
3. The settle block's own resolution assignment becomes an append.

Part 3 is the one that is a decision rather than plumbing. Left as it was —

    if settled.outcome != OutcomeClear {
        result.Outcome = settled.outcome
        result.Resolution = settled.resolution
    }

— a fold that reported attention followed by a removal half that also did would have had the fold's
resolution silently replaced, which is the same defect this bead fixes moved one line down: a
caller told about the worktree and not about the irreversible write whose result is unknown. So it
now appends, matching breakerUnread in cli/swarm/cli/ready.go, which documents the same choice:

    if result.Resolution == "" {
        result.Resolution = resolution
    } else {
        result.Resolution += "; " + resolution
    }

Rejected alternative: keep the substitution and accept the loss on the ground that both halves
reporting attention is rare. Cost: it introduces, in the same commit, a narrower instance of the
bug being fixed — and rarity is exactly what makes it survive review.

The outcome itself still substitutes rather than composing, deliberately: the removal half's
verdict is the later and the more severe (attention over attention, stop over anything), and
foldIn never returns OutcomeStop through mergeOutcome — its stop paths return an error — so
`result.Outcome = settled.outcome` can only hold or escalate, never downgrade.
```

**`agent-plugins-jed1#2`** · `dev-bguq-2`

The AC's two world conditions cannot both hold with ONE worktree, so the fixture declares TWO. Decided rather than escalated.

```
evidence: AC bullet 3 requires 'the target ahead of the dispatch branch'; AC bullet 4 requires a worktree that 'sits at the sprint head with no commit of its own'. The bead's condition 1 makes the worktree's branch BE the dispatch branch ('git refuses to rebase a branch another worktree has checked out'), so one tree cannot be both behind the target and at its head. Two declared worktrees satisfy both literally: worktree-agent-3f8a2c1d at 016160f (behind the target, carries dev-fixture-2's own commit -> condition 1) and worktree-agent-b7e40952 at d54d47f, the sprint head, with 0 commits of its own (-> conditions 2 and 3). REJECTED: a single worktree behind the target whose HEAD commit was another dispatch's, which satisfies the substance of all three conditions but reads AC bullet 4's 'sprint head' as 'where the sprint head was when the tree was cut'. Cost of the rejection: one extra branch and one extra [[worktree]] entry in the bundle. Two trees also model the real run better -- one dispatch that has committed and one that has not are both live at once, which is the world agent-plugins-asv8 was measured in.
```

**`agent-plugins-jed1#3`** · `dev-bguq-2`

The new fixture's manifest lists all five beads in `admitted`, diverging from the two existing variants, which leave it at the core's two.

```
evidence: core/tracker.jsonl records `"admitted": ["fixture-run.1", "fixture-run.2"]`. dispatch-worktree and dispatch-stale-base each append beads .3/.4 to the tracker and leave `admitted` untouched, so both worlds carry a scope drift nobody chose: cli/swarm/status/status.go's `compareScope` reads `admitted` and reports every live bead absent from it as an addition. dispatch-concurrent-worktree sets `admitted` to .1 through .5 so its only anomalies are the three the fixture exists for. REJECTED: following the two variants' precedent, which costs nothing today (no case reads scope against them) but would put a second unrelated anomaly into the world this fixture's future cases assert against -- and a fixture modelling drift it did not mean to model is the exact defect the Change Guidance rule added by this bead condemns. Checked before writing: `validateAdmission` (cli/swarm/manifest/validate.go:238) only forbids a bead appearing in more than one of admitted/excluded/moved, so a five-id list breaks no rule. The two existing variants' `admitted` lists were NOT corrected here -- out of this bead's scope, and unfiled because no case reads scope against either fixture today.
```

**`agent-plugins-f9lq#4`** · `dev-bguq-3`

The three refusal sites were consolidated behind one helper rather than repeated inline, so the guard has a seam a test can reach.

```
The bead's Design pinned the remedy ("Guard the empty case at the call site and return RefuseUnrepairable naming what could not be cited") but not its shape. Written inline it is three near-identical guards with no callable seam — and since the never-ran path at all three is unreachable end to end (see the unverifiable finding), an inline guard would have had NO test at all.

evidence: cli/swarm/cli/ready.go now carries

    func invocationsRequired(what string, records []plan.Record) ([][]string, error) {
        inv := invocationsOf(records)
        if len(inv) == 0 {
            return nil, RefuseUnrepairable(
                "%s names no invocation a reader could re-run, so its answer is not reported", what)
        }
        return inv, nil
    }

called from acceptDispatch (accept.go), recheckDispatchBeads (accept_settle.go) and reportAgentQueue (ready.go), each passing its own subject string. The guard is still at the call site — each site decides whether to refuse and what to name — and the refusal is now reachable by TestInvocationsRequiredRefusesAReadNoRecordCites.

Rejected: three inline `if len(inv) == 0 { return ..., RefuseUnrepairable(...) }` blocks. It would have cost an untestable guard at each of the three sites and a third copy of the same sentence to keep in agreement.
```

**`agent-plugins-u9uh#2`** · `dev-bguq-3`

Placement: both guards sit AFTER the mode-resolving switch, not inside it.

```
The Design says "immediately after the existing --explain / --dry-run mutual-exclusion check and before the mode is resolved". Those two are not separable in the code: cli/swarm/cli/ready.go resolves the mode in the same switch statement that carries the mutual-exclusion refusal —

    switch {
    case explain && dryRun:
        return finish(streams, req, refusal(req, RefuseArguments(
            "--explain and --dry-run select different interpreters; pass one")))
    case explain:
        req.Mode = plan.ModeExplain
    case dryRun:
        req.Mode = plan.ModeDryRun
    default:
        req.Mode = plan.ModeExecute
    }

so there is no point that is after the one and before the other. Placed immediately after the switch instead, which preserves both consequences the Design named as its reasons: the guards are still ahead of `if req.Mode == plan.ModeExplain { return finish(streams, req, explainReady(...)) }`, so a malformed invocation is malformed in explain mode too, and still ahead of applyBreaker on both the --agent and whole-run paths, so nothing is read or written. Both are asserted by the test across all three modes.

Rejected: placing them before the switch. It would make a threshold refusal preempt the --explain/--dry-run one when both are malformed, changing which of two correct refusals a caller sees, for no gain.

evidence: with the guards where they are, `go test ./swarm/cli/` exits 0, and TestReadyRefusesANonPositiveThreshold covers nil, --dry-run and --explain for each of 0, -1 and -3.
```

### `gap` — 4

**`agent-plugins-4ca8#2`** · `dev-bguq-1`

the bead's Steps to Reproduce cannot be carried out through the real Git Version Control; the state it names is unreachable from outside foldIn, so the test needed a substituted VC

```
evidence: the bead says "Arrange the merged-head read to be skipped or absent". Measured against the
code as it stands, neither is reachable through gitvcs.VersionControl:

  ABSENT — plan.Report.Record answers !ok only when the report carries no record for the step.
  Reports carry a record for every step in dry-run and execute (cli/swarm/plan/run.go: interpret
  appends one per step, and skipRest appends one for each it did not reach). The one mode that
  produces no records at all is ModeExplain — Run returns `Report{Plan: p.Name}` with no Records
  (cli/swarm/plan/run.go, `case ModeExplain:`). But then the fast-forward's record is absent too,
  and foldIn returns early:
    "the fold of %s into %s produced no record"  (cli/swarm/cli/accept.go)

  SKIPPED — a step is recorded plan.StepSkipped in exactly two places (grep -n StepSkipped
  cli/swarm/plan/run.go -> run.go:180 and run.go:254). run.go:180 is an argument-resolution failure,
  which needs a plan.From reference argument; every argument gitvcs Resolve builds is a plan.Lit
  (cli/swarm/vcs/git/git.go, `func (v VersionControl) step` wraps each arg in plan.Lit). run.go:254
  is skipRest, which skips only the steps AFTER the one that stopped the plan — and in foldIn's plan
  the merged-head resolve is last, so anything that could stop ahead of it is the fast-forward,
  whose failure foldIn already returns on via refusedFastForward.

So the defect is latent rather than live: it is a broken defence for a state foldIn's own comment
anticipates ("this is the one site where the record may legitimately be absent"), not a path a user
can drive today. The test therefore substitutes vcs.VersionControl (skippedMergedHeadVC in
cli/swarm/cli/accept_test.go) to put a stopping guard step ahead of the real resolve, which produces
a genuine plan.StepSkipped record with literal args — the shape the real fragment would have.

What this means for the AC: "covers the skipped-record path after a completed fast-forward" is
satisfied at the foldIn level with a substituted VC, and cannot be satisfied end-to-end through
runAccept and the exec fake. Nothing in the bead said which altitude was required.
```

**`agent-plugins-4ca8#5`** · `coordinator-bguq`

— this fix has a RESIDUAL HOLE, found after the bead closed and merged (@ 9c20218).

```
Found by run bguq's second grooming pass while designing agent-plugins-f9lq, which generalises this same defect to five more sites.

THE HOLE: invocationOf itself returns nil when a step carries a reference argument --
    if a.IsRef() { return nil }
So the remedy applied here, `[][]string{invocationOf(...)}` at accept.go ~643, can still yield
    [][]string{nil}
which trips Validate's SECOND rule rather than its empty-slice rule. The bead's acceptance criteria are met — a fact built from a skipped or absent record now carries an invocation, and merged:true survives — but the guarantee is narrower than 'merged_sha always names an invocation' reads.

NOT REOPENED, deliberately. The bead's stated criteria are satisfied and its tests pass; reopening a merged bead to widen its guarantee is scope creep on a closed record. agent-plugins-f9lq's design instead tells its implementer to CHECK the property holds for the three phase.go steps it touches and to fall back to the refusal remedy where it does not — so the same class is handled there, with the ref-argument case explicit.

Recorded so this bead's record does not read as a complete solution to a problem it only narrows.
```

**`agent-plugins-n3tn#2`** · `dev-bguq-1`

the "comparison did not run" branch is unreachable from OUTSIDE the function in either mode, not just in a dry run — so the test had to call landingVerdict directly, which the bead did not anticipate

```
evidence: the bead establishes that a DRY RUN cannot reach the branch. Tracing the execute path as
well, nothing else can either:

  - The divergence step is a plain read with literal arguments
    (cli/swarm/vcs/git/git.go, `func (v VersionControl) Divergence` -> `v.step(prefix, plan.Read,
    ...)` with "rev-list", "--count", "--left-right" and one built string), so it cannot be
    refused by dry run and cannot be skipped by an argument that will not resolve — the only two
    non-skipRest routes to a non-StepRan record (cli/swarm/plan/run.go).
  - That leaves skipRest, which needs an earlier step to stop the plan. The steps ahead of it in
    verifyLanded's plan are push, fetch and counterpart. Each has a branch in landingVerdict (or,
    for push, in verifyLanded) that RETURNS before the comparison is reached:
    the push through `rejected(result, remote, rec)`, the fetch through the guard above, and the
    counterpart through `if counterpartRec.Failed()`.

So the branch is a defence against a plan shape the current step set cannot produce. That is not
an argument for deleting it — the fragments are built by an interface, and a second
implementation of vcs.VersionControl could order things differently — but it does mean the branch
has no end-to-end test and cannot be given one through runBranch and the exec fake.

The test therefore calls landingVerdict directly with a hand-built plan.Report. That is possible
only because landingVerdict takes a report and three step fragments as plain values and returns a
verdict — no context, no runner, no I/O. Worth noting as a property to preserve: the same shape is
what let this branch be pinned at all, and a refactor that gave landingVerdict a runner would
remove the only way to test it.

Two helpers were added to cli/swarm/cli/branch_test.go for it — landingFragments (the three real
fragments) and ranRecord (a step that ran and answered) — and they are reusable for any other
landingVerdict branch that later needs the same treatment.
```

**`agent-plugins-bguq.6#2`** · `retro-bguq` · target `sprint-scope`

The run manifest records no model lineup and no effort, and `sprint-retro`'s close-the-loop step requires both on every issue it files.

```
evidence: `sprint-retro` says "Every issue carries the run's model lineup and effort from the manifest into its provenance - `swarm manifest get --run {EPIC} --key models --key effort`. The manifest records them precisely so this step can, and an issue without them ships with an unfillable field." Measured:

  $ bd show agent-plugins-bguq --json | jq -r '.[0].metadata | keys[]' | tr '\n' ' '
  admitted base_sha branch checks_baseline checks_ids collision_edges concurrency_cap consecutive_dispatch_failures review_passes standing_auth

Neither `models` nor `effort` is present, so the eight candidates below carry an unfillable provenance field. What IS recoverable is the DECLARED tier per bead (`bd show <id> --json | jq '.[0].metadata.tier'` -> standard x4, light x2, advanced x1), which resolves through sf.toml's [models] to sonnet/haiku/opus - but a declared tier is a dispatch instruction, not a record of what served the turn, and tier_escalation=once means the two are permitted to differ. Recurrence: run e6xt's retro recorded "manifest `effort` reads `unrecorded`" under what it could not verify.
```

### `scope` — 3

**`agent-plugins-4ca8#4`** · `dev-bguq-1`

the same defect shape survives at five more sites in cli/swarm/cli; filed as agent-plugins-f9lq (discovered-from), adjacent and NOT blocking

```
evidence: the bead scoped the fix to one expression in foldIn, and it is one of six instances of
"a fact built from a variadic invocation slice nothing guarantees is non-empty". The other five,
found by tracing the two producers that can answer with nothing:

  $ grep -n "invocationsOf(\|argvOf(rec); argv != nil" cli/swarm/cli/*.go
  cli/swarm/cli/accept.go:234:	queueInv := invocationsOf(qRecords)
  cli/swarm/cli/accept_settle.go:434:	beadsInv := invocationsOf(recs)
  cli/swarm/cli/phase.go:307:		if argv := argvOf(rec); argv != nil {
  cli/swarm/cli/phase.go:368:		if argv := argvOf(rec); argv != nil {
  cli/swarm/cli/ready.go:570:	queueInv := invocationsOf(qRecords)
  cli/swarm/cli/ready.go:571:	establishedInv := invocationsOf(dRecords)

The sharpest of them is cli/swarm/cli/phase.go, which fails the SECOND Validate rule rather than
the first:

    closableInv := argvOf(closableRec)
    ...
    List("closed_epics", closedIDs, append([][]string{closableInv}, closeInvs...)...)

guarded only by `if !ok || closableRec.Failed()`. Record.Failed() is false for a refused or skipped
record (cli/swarm/plan/plan.go, `func (r Record) Failed`), so such a record passes the guard and
produces [][]string{nil} — length 1, argv empty, which Validate refuses as
"names an empty invocation".

Not implemented here: the bead's acceptance criteria are met without touching any of them, and a
six-site sweep is a different change with a design question of its own (per-site invocationOf, or
rejecting an empty invocation set in the Fact constructors so it fails where written rather than
after an irreversible write). That question is stated in f9lq.
```

**`agent-plugins-jed1#5`** · `dev-bguq-2`

Filed agent-plugins-eu0f (discovered-from this bead): dispatch-worktree/fixture.toml's claim that a worktree is 'attributable to its dispatch by branch name alone' is false against status.go, and the Change Guidance rule added here is what makes correcting it obligatory rather than cosmetic. Not fixed here because this bead's Design forbids it ('Do not edit dispatch-worktree's misleading comment as part of this order').

```
evidence: Sweep run BEFORE writing anything, per the retraction protocol: `grep -rn 'attributable to its dispatch' . --exclude-dir=.git` -> exactly three hits: cli/testdata/swarm-fixtures/dispatch-worktree/fixture.toml:6, docs/retros/2026-09-02-e6xt-swarm-dispatch.md:107 and :838. The two retro hits QUOTE the claim as the finding, so they are correct as written and must not be swept. `grep -rn 'branch name alone' . --exclude-dir=.git` -> no hits outside that same fixture.toml sentence. No durable trigger prompt, dispatch prompt or .sf/outbox copy carries it -- `.sf/outbox` does not exist in this tree. `bd search dispatch-worktree --status all` and `bd search fixture.toml --status all` both returned 'No issues found', so no existing bead owned it; agent-plugins-nqe8 is the nearest neighbour and is scoped to worktree-mechanics.md's merge/removal sections, not to the fixture.
```

**`agent-plugins-f9lq#5`** · `dev-bguq-3`

The bead's five sites do not exhaust the shape it describes. Filed as agent-plugins-bguq.4.1 rather than widened into this bead.

```
evidence: `func mustRecord(report plan.Report, name string) plan.Record` in cli/swarm/cli/dispatch.go discards the "record absent" signal —

    rec, _ := report.Record(name)
    return rec

— so a missing record yields the zero plan.Record, whose Invocation.Name is empty, which is exactly the state argvOf answers nil for. Three facts in acceptDispatch (cli/swarm/cli/accept.go) are cited from it with no guard between:

    targetInv:  argvOf(mustRecord(report, targetStep.Answer())),
    branchInv:  argvOf(mustRecord(report, branchStep.Answer())),
    caseInv:    argvOf(caseRec),

reaching Text("target", ...), Text("dispatch_branch", ...) and Text("case", ...) directly. Same standing as the five: latent, not live.

Not implemented here — this bead's AC names five sites and its Design pins a remedy per site, so repairing more would be a wider change than the order authorises, and the constructor-level alternative the Design rejected is the live question for the follow-up rather than for this one.
```

### `stale` — 1

**`agent-plugins-f9lq#1`** · `dev-bguq-3`

The bead's site-4 analysis is incomplete: an earlier guard already blocked the never-ran path it describes.

```
The bead says of closePhases: "The guard above it is `if !ok || closableRec.Failed()`, and `func (r Record) Failed` ... opens with `if r.Status != StepRan { return false }` — so a record that never ran passes the guard and yields `[][]string{nil}`."

evidence: it passes THAT guard and is then stopped by the next one. Read from cli/swarm/cli/phase.go, in order:

    closableRec, ok := closableReport.Record("closable")
    if !ok || closableRec.Failed() { ... RefuseUnrepairable ... }
    closableAnswer, err := tracker.ClosableFrom(closableRec.Value)
    if err != nil { return Result{}, RefuseUnrepairable("%s", err) }
    closableInv := argvOf(closableRec)      // <- the bead's site

and cli/swarm/tracker/tracker.go:

    func ClosableFrom(value any) (Closable, error) {
        if value == nil {
            return Closable{}, fmt.Errorf("tracker: the step produced no value")
        }

A record that never ran carries a nil Value (cli/swarm/plan/run.go sets Value only on the StepRan path), so ClosableFrom refuses before closableInv is ever built. The site-4 `[][]string{nil}` was therefore not reachable even in principle.

The repair still landed as designed — invocationOf in place of argvOf — because the reachable half of site 4 is real: closeEligible's `close-N` records ARE refused in a dry run, which is where the citation was being dropped. That half is now covered by TestPhaseCloseCitesEveryFactWhenNoCloseRan, and it fails against the pre-fix code for the stated reason (negative control run, output below).
```

### `unverifiable` — 3

**`agent-plugins-74xc#2`** · `dev-bguq-1`

acceptance criterion 1 is false as written, and satisfying it literally would delete a committed test that pins deliberate behaviour; I implemented the bead's Design instead and am recording which reading I took

```
evidence: AC 1 reads "For every combination of recorded count, threshold and --completed value,
dry-run and execute agree on the stop verdict."

That is contradicted by a test already in the tree, and by the bead's own Design paragraph two
lines further down ("Preserve the other half of the stated rule — a dry run must still surface a
stop the manifest already carries for --completed 0").

The counterexample is recorded 2, threshold 3, --completed 0:

  - execute:  next = recorded + 1 = 3, and `if next >= threshold` -> stopped = true
  - dry run:  the write is refused, the verdict rests on recorded = 2 -> stopped = false

They disagree, deliberately. cli/swarm/cli/ready_test.go pins it:

  // The regression guard for the test above: BELOW the threshold a refused
  // write still leaves the run clear, so what that test asserts is a stop
  // resting on the RECORDED count rather than the stop test having been
  // moved onto the projection.
  func TestADryRunBelowTheThresholdStaysClear(t *testing.T) {
      ...
      if stopped {
          t.Errorf("a dry run stopped the run from a recorded 2 against a threshold of 3 "+
              "(summary %q); the 3 it would have written is a number nothing wrote", summary)
      }
  }

and the rule it enforces is stated in applyBreaker's own comment: "A dry run may not INVENT a stop
from a projection; it must not HIDE one the manifest already carries."

So AC 1 cannot mean universal parity. The reading I took, and implemented:

  The verdict follows whichever value execute would compare, WHENEVER that value is one the
  REPORT settles rather than one the write would have invented.

  --completed N>0 fixes the next count at 0 by itself — the caller supplied it, and execute
  applies the reset unconditionally. Not a projection, so dry run uses it. (the fix)
  --completed 0 only proposes recorded+1, which nothing wrote. A projection, so dry run does
  not use it. (unchanged)

Under that reading AC 2, 3, 4 and 5 all hold as written, and AC 1 holds for every case where the
two modes are comparing the same settled value. Rejected alternative: make dry run use `next`
unconditionally, which satisfies AC 1 literally. Cost: it deletes
TestADryRunBelowTheThresholdStaysClear and the stated rule it guards, turning the dry run into a
mode that invents stops — the failure the comment was written to prevent, and a change well
outside "Scope is this branch only".

What would settle it: an amendment to AC 1 scoping it to `--completed N>0`, or a statement from
whoever wrote it that the projection rule is being retired. I did not stop for that answer.
```

**`agent-plugins-jed1#4`** · `dev-bguq-2`

The harness's worktree DIRECTORY name is unmeasured; only the BRANCH form is. The fixture says so rather than asserting it.

```
evidence: cli/swarm/status/status.go's note above `KindEpic` measures the branch form only -- 'Measured against this harness's own worktrees, whose branches read `worktree-agent-<hash>`' -- and says nothing about the path. I could not settle it here: `git worktree list --porcelain` in this session returns two entries, the main checkout on `refs/heads/claude/swarm-dispatch-blueprint-5gob2e` and `/home/user/.bb-worktrees/dev-bguq-2` on `refs/heads/bb/dev-bguq-2` -- both coordinator-created, no `worktree-agent-` tree standing to measure. The fixture's two `path` values are named for their branches and the fixture.toml marks that as chosen, not measured, and names the command that would settle it. An earlier draft of that comment asserted 'the name the harness gives the branch is the name it gives the directory' as fact; it was corrected before commit. Nothing in the fixture depends on the answer -- attribution reads the HEAD commit, and the paths only need to carry no dispatch id, which they do. Settle by running `git worktree list --porcelain` in a session where the harness has created a tree; agent-plugins-gqwl's own description happens to record one such capture (`/home/user/agent-plugins/.claude/worktrees/agent-a19815a2ef1fbb2c3701e2cd`), which suggests the path is `agent-<hash>` under `.claude/worktrees/` rather than `worktree-agent-<hash>` -- but that is one quoted capture in a bead description, not a measurement I took, so it stays unverified.
```

**`agent-plugins-f9lq#2`** · `dev-bguq-3`

The acceptance criterion "a unit test covers each of the five sites on the never-ran-record path" cannot be satisfied as declared for sites 1, 2 and 3: that path is unreachable by construction, by three independent guards.

```
evidence: three arrangements were tried and each was blocked.

1. Drive the real helpers into the never-ran state. Both tracker.Descendants and tracker.AgentQueue (cli/swarm/tracker/descend.go, queue.go) refuse a record that did not run, because such a record carries a nil Value:

       func BeadsFrom(value any) ([]Bead, error) {
           if value == nil { return nil, fmt.Errorf("tracker: the step produced no value") }

   plus AgentQueue's own `answered` map check ("a queue is not reported from one half"). So a skipped or refused record makes the helper return an ERROR, and all three sites return on that error before the citation is built.

2. Use a dry run to refuse the reads. Both helpers hardcode plan.ModeExecute in their own runner.Run calls, so req.Mode never reaches them.

3. Arrange a record that RAN but carries an empty Invocation.Name — the state the bead's own Steps to Reproduce prescribes ("Arrange its records so that every record's Invocation.Name is empty"). cli/swarm/plan/run.go sets `inv := swarmexec.Invocation{Name: s.Executable, ...}`, so this needs a step declaring no executable. Measured: a fake tracker returning such steps produced

       reason = "tracker: walking below run-uncited: plan: step \"children-0-0\" has no executable"

   from cli/swarm/plan/plan.go:400 — plan.Plan.Validate refuses it before anything runs.

Decision taken rather than escalated: sites 1-3 are covered at their guard (TestInvocationsRequiredRefusesAReadNoRecordCites) and on the reachable path end to end (TestAcceptCitesEveryFactItReports, TestAcceptCitesEveryFactWhenTheDispatchHasNotDrained, TestReadyAgentCitesEveryFactItReports, which assert every fact in the emitted JSON envelope names a non-empty invocation). Rejected alternative: widening tracker.AgentQueue and tracker.Descendants to return their steps so the state could be constructed — the bead's Design forbids it, and its AC pins their signatures. The unreachability is recorded in cli/swarm/cli/citation_test.go's own header so nobody re-litigates it.

Sites 4 and 5 ARE reachable and are covered end to end: a dry run refuses each `close-N` write, which is the never-ran state closeEligible used to skip without citing.
```

### `tooling` — 1

**`agent-plugins-bguq.6#1`** · `retro-bguq` · target `sprint-retro`

The documented harvest query returns nothing on bd 1.2.2, and the replacement carried in this phase's own dispatch prompt is wrong the same way.

```
evidence: `sprint-retro` SKILL.md section 2 and references/shared/findings-capture-protocol.md both prescribe `bd export | jq -r 'select(...) | .comments[]? | ...'`. Measured, bd 1.2.2 (6c124203e):

  $ bd export | head -1 | jq -r 'keys[]' | tr '\n' ' '
  _type assignee close_reason closed_at comment_count created_at created_by dependencies dependency_count dependent_count description id issue_type labels owner priority status title updated_at

No `comments` key exists, so `.comments[]?` yields nothing and the harvest reads as a clean run. The dispatch prompt that opened this phase carried the fix as "rebuild the harvest per-bead with `bd show <id> --json`". That is also wrong:

  $ bd show agent-plugins-jed1 --json | jq -r '.[0] | keys[]' | tr '\n' ' '
  acceptance_criteria assignee closed_at comment_count created_at created_by dependencies dependency_count dependent_count description design id issue_type labels metadata owner parent priority started_at status title updated_at
  $ bd show agent-plugins-jed1 --json | jq '.[0].comments | length'
  0

The form that works is a third one:

  $ bd comments agent-plugins-jed1 --json | jq 'length'
  5

`bd comments list` is explicitly rejected by bd ("Invalid - use bd comments <issue-id> to list comments"). This retro rebuilt the harvest on `bd comments <id> --json` and reconciled it against the count the beads carry: sum(comment_count) over the seven admitted beads is 26 and 26 findings were harvested. FEAT-SWARM-RUN-007 in specs/features/swarm-cli/swarm-cli-frd.md specifies exactly this (`swarm finding harvest`, reconciled against the beads' own count) and names both failing queries in its Feature narrative; no `finding` command exists in cli/swarm/cli today (`grep -rn 'finding harvest\|findingCmd' cli/swarm/cli/*.go cli/cmd/*.go` -> no hits). Recurrence: run e6xt's retro recorded the same defect as its candidate C1, which was never filed.
```

### `proof` — 11

**`agent-plugins-4ca8#1`** · `dev-bguq-1`

the defect was observed failing before the fix, and the fix is what turned it green

```
evidence: new test TestASkippedMergedHeadResolveStillCitesAnInvocation in cli/swarm/cli/accept_test.go,
run against the UNCHANGED foldIn:

  $ cd cli && go test ./swarm/cli/ -run TestASkippedMergedHeadResolveStillCitesAnInvocation -v
  === RUN   TestASkippedMergedHeadResolveStillCitesAnInvocation
      accept_test.go:810: merged_sha names no invocation that established it
      accept_test.go:824: the result does not validate, so it degrades to a bare refusal: cli: fact "merged_sha" names no invocation that established it
  --- FAIL: TestASkippedMergedHeadResolveStillCitesAnInvocation (0.00s)
  FAIL	github.com/brettcallaghan/agent-plugins/cli/swarm/cli	0.010s

The two t.Fatalf preconditions ahead of those lines (merged-head recorded plan.StepSkipped, and
carrying no invocation of its own) did NOT fire, so the scenario the bead describes did arise and
the failure is the reported one and not a mis-built world.

After replacing argvOf with invocationOf, the same test plus both incumbent merged_sha tests pass:

  $ go test ./swarm/cli/ -run 'TestASkippedMergedHeadResolveStillCitesAnInvocation|TestAFailedPostFoldResolveStillReportsTheMergeAndStillSettles|TestAPostFoldResolveThatAnswersStillEstablishesTheMergedSHA' -v
  --- PASS: TestAFailedPostFoldResolveStillReportsTheMergeAndStillSettles (0.00s)
  --- PASS: TestAPostFoldResolveThatAnswersStillEstablishesTheMergedSHA (0.00s)
  --- PASS: TestASkippedMergedHeadResolveStillCitesAnInvocation (0.00s)
  ok  	github.com/brettcallaghan/agent-plugins/cli/swarm/cli	0.024s

The last two are the negative controls the bug's own path needs: a merged-head record that RAN and
failed (exit 128) still cites its own argv, and one that answered still establishes the sha, so the
change did not simply stop reporting the fact.
```

**`agent-plugins-j5ah#1`** · `dev-bguq-1`

both the defect and the composition branch added to fix it were observed failing before their fix

```
evidence: (1) the defect. TestAFailedPostFoldResolveTakesTheAttentionOutcome, run against the
UNCHANGED foldIn:

  $ cd cli && go test ./swarm/cli/ -run TestAFailedPostFoldResolveTakesTheAttentionOutcome -v
  === RUN   TestAFailedPostFoldResolveTakesTheAttentionOutcome
      accept_test.go:856: exit status = 0, want 1: an unestablished merged_sha is a condition to act on
        {
          ...
          "outcome": "clear",
          "status": 0,
          "summary": "bb/dev-1 was serialized against sprint/run-1 and fast-forwarded into it; ...",
          "resolution": "",
  --- FAIL

Exactly the three the bead names: outcome clear, status 0, resolution empty.

(2) the composition branch. Applying the outcome to the result meant deciding what happens when the
REMOVAL half also reports a verdict, since the settle block below it assigns result.Resolution
outright. I made it append, and tested that with
TestBothHalvesNeedingAttentionKeepBothResolutions. Negative control — the append temporarily
reverted to the original substitution:

  $ go test ./swarm/cli/ -run TestBothHalvesNeedingAttentionKeepBothResolutions
  --- FAIL: TestBothHalvesNeedingAttentionKeepBothResolutions (0.00s)
      accept_test.go:922: the fold's resolution was dropped when the removal half added its own:
        "put each named signal right — close the bead, or record the finding it is missing — then
         accept the dispatch again"
  FAIL

and green with the append restored. So the branch is not untested code added on the way past.

That control also caught a bad assertion of my own: the test first checked the removal half's
resolution by looking for the branch name `bb/dev-1`, which the FOLD's resolution also carries — it
passed for the wrong reason. Under the control it failed on both lines, which is what exposed it.
It now looks for "accept the dispatch again", which only the removal half says.
```

**`agent-plugins-74xc#1`** · `dev-bguq-1`

the inversion was observed before the fix, in the exact shape the bead describes, and the preserved half was already green and stayed green

```
evidence: TestAPositiveCompletedNeverStopsTheRunInEitherMode (new, cli/swarm/cli/ready_test.go),
run against the UNCHANGED applyBreaker:

  $ cd cli && go test ./swarm/cli/ -run 'TestAPositiveCompletedNeverStopsTheRunInEitherMode|TestCompletedZeroStillReportsTheRecordedStopInADryRun' -v
  === RUN   .../recorded-3-dry-run
      ready_test.go:934: stopped = true after a dispatch that completed 2 bead(s): the count resets to 0, and 0 is below the threshold of 3 (summary "the run is stopped: 3 consecutive dispatch(es) returned with no bead completed (threshold 3)")
      ready_test.go:946: the result says the count resets to 0 AND that the run is stopped: "the write was refused: dry run: the step writes. The count stands at 3 and would have become 0 (threshold 3)" / "the run is stopped: 3 consecutive dispatch(es) returned with no bead completed (threshold 3)"
  === RUN   .../recorded-3-execute
  === RUN   .../recorded-5-dry-run
      ready_test.go:934: stopped = true after a dispatch that completed 2 bead(s) ... (summary "the run is stopped: 5 consecutive ...")
  === RUN   .../recorded-5-execute
      --- FAIL: .../recorded-3-dry-run
      --- PASS: .../recorded-3-execute
      --- FAIL: .../recorded-5-dry-run
      --- PASS: .../recorded-5-execute
  --- PASS: TestCompletedZeroStillReportsTheRecordedStopInADryRun

Three things that matter in that output:

1. Only the DRY-RUN rows failed. Both execute rows passed untouched, which is the
   divergence rather than a test that is simply wrong about the breaker.
2. Both AT the threshold (recorded 3) and ABOVE it (recorded 5) failed, as the AC asks.
3. The self-contradictory text is in the same result: "would have become 0" beside
   "the run is stopped".

After the fix all four rows pass, and so do every incumbent breaker test — including the two
that pin the half of the rule that must NOT change:

  --- PASS: TestADryRunReportsAStopTheRecordedCountAlreadyEstablishes
  --- PASS: TestADryRunBelowTheThresholdStaysClear
  --- PASS: TestReadyDryRunWithCompletedExitsStopAtARecordedThreshold
  --- PASS: TestBreakerStopsTheRunAtItsThreshold
  --- PASS: TestBreakerResetsToZeroWhenSomethingCompleted
  --- PASS: TestBreakerCountsOneDispatchReturnNotItsBeads
  --- PASS: TestBreakerBareQueryReadsWithoutWriting
  --- PASS: TestBreakerAbsentCountReadsAsZero
  --- PASS: TestBreakerSkipsAWriteThatWouldNotChangeTheCount
  --- PASS: TestADryRunOfReadyRecordsTheBreakerWriteAsRefused

Every one of those passes --completed 0 or omits it, so none of them exercised the reset path —
which is why the defect survived them.
```

**`agent-plugins-n3tn#1`** · `dev-bguq-1`

the wrong resolution was observed on both verbs before the fix, and the fetch guard's wording was green before and after

```
evidence: TestTheComparisonDidNotRunResolutionProposesNoWrite (new, cli/swarm/cli/branch_test.go),
run against the UNCHANGED landingVerdict:

  $ cd cli && go test ./swarm/cli/ -run 'TestTheComparisonDidNotRunResolutionProposesNoWrite|TestTheFetchGuardKeepsTheDryRunWording' -v
  === RUN   .../push
      branch_test.go:439: the resolution sends the caller out of a mode a dry run cannot reach this branch in: "run it again without --dry-run: swarm branch push --branch sprint/run-1 --remote origin"
      branch_test.go:443: the resolution proposes a WRITE as the remedy for a read that did not execute: "run it again without --dry-run: swarm branch push --branch sprint/run-1 --remote origin"
      branch_test.go:449: the resolution does not name the tracking ref the comparison was against: ...
      branch_test.go:453: the resolution names nothing the caller can run to obtain the comparison: ...
  === RUN   .../verify-landed
      branch_test.go:439, :449, :453  (same, minus the write assertion)
  --- FAIL: TestTheComparisonDidNotRunResolutionProposesNoWrite
  --- PASS: TestTheFetchGuardKeepsTheDryRunWording/push
  --- PASS: TestTheFetchGuardKeepsTheDryRunWording/verify-landed

The push row failing on line 443 where verify-landed does not is the bead's specific claim
reproduced: `verb` is interpolated, so the same sentence proposes a write for one verb and not the
other.

TestTheFetchGuardKeepsTheDryRunWording passed BEFORE the change and passes after — it is the
negative control for AC 4, since the risk of this fix is removing the dry-run advice from the one
branch that legitimately carries it. TestADryRunDoesNotReportLandedFromAnUnfetchedComparison, the
incumbent test over that same guard, also still passes.

Unreachability, measured by reading rather than by running: `Fetch` is declared a write —

  $ grep -n "func (v VersionControl) Fetch" -A 6 cli/swarm/vcs/git/git.go
  316:func (v VersionControl) Fetch(prefix, remote string) vcs.Steps {
  317:	s := v.step(prefix, plan.Write, true, discard,
  318:		"fetch", remote)

— and cli/swarm/plan/run.go refuses a write in dry run before it runs
(`if mode == ModeDryRun && s.Effect == Write { record.Status = StepRefused ... }`), so the fetch
record is StepRefused and the guard `fetchRec.Status != plan.StepRan` fires. A dry run returns
there and cannot arrive at the comparison branch. That is what the second test encodes as an
executable statement rather than a comment.
```

**`agent-plugins-jed1#1`** · `dev-bguq-2`

The world exercises defect 1 (rebase refused while the worktree stands), established with a negative control that isolates the WORKTREE as the cause rather than the file contents.

```
evidence: With both declared worktrees standing, in a clone of the new bundle: `git rebase claude/bead-bonanza-fixture worktree-agent-3f8a2c1d` -> exit 128, `fatal: 'worktree-agent-3f8a2c1d' is already used by worktree at <path>`. NEGATIVE CONTROL, same bundle, fresh clone, local branch created with `git branch worktree-agent-3f8a2c1d origin/worktree-agent-3f8a2c1d`, NO worktree standing: same rebase -> exit 0, `Successfully rebased and updated refs/heads/worktree-agent-3f8a2c1d.`, `git diff --name-only --diff-filter=U` empty. The two sides touch different files (work.txt vs notes.txt), so the refusal is the worktree and nothing else. Concurrency measured separately the way readCase measures it: `git merge-base --is-ancestor claude/bead-bonanza-fixture worktree-agent-3f8a2c1d` -> exit 1 (concurrent); merge-base ac4a37c vs target head d54d47f. Defect-3 condition measured: `git rev-list --count claude/bead-bonanza-fixture..worktree-agent-b7e40952` -> 0, and both refs resolve to d54d47f.
```

**`agent-plugins-f9lq#3`** · `dev-bguq-3`

Negative control: the new end-to-end test fails against the pre-fix code, for the stated reason.

```
evidence: with cli/swarm/cli/phase.go's closeEligible reverted to its pre-fix form — `if !ok || rec.Status != plan.StepRan { continue }` followed by `argvOf(rec)` — and nothing else changed:

    $ go test ./swarm/cli/ -run 'TestPhaseCloseCitesEveryFactWhenNoCloseRan'
    --- FAIL: TestPhaseCloseCitesEveryFactWhenNoCloseRan (0.00s)
        citation_test.go:200: closed_epics cites [[bd epic status --eligible-only --json]], want the close it would have run among them
    FAIL
    FAIL	github.com/brettcallaghan/agent-plugins/cli/swarm/cli	0.010s
    EXIT=1

With the fix restored the same run exits 0. The pre-fix citation list holds only the eligibility read; the refused write contributed nothing, which is the dropped citation this bead is about.

The test also pins the state it is asserting about before asserting anything, so it cannot pass vacuously if the scenario stops arising:

    if closed.Status != plan.StepRefused {
        t.Fatalf("close-0 is %q, want %q — the scenario did not arise", closed.Status, plan.StepRefused)
    }
```

**`agent-plugins-f9lq#6`** · `dev-bguq-3`

The gates pass with no finding beyond the run's baseline.

```
evidence: `plugin/skills/bead-bonanza/scripts/bb-checks.py` with no arguments, run in the dispatch worktree, exited 0 with all thirteen declared checks present and passing — format, vet, build, test-go, skills, shared-refs, strata-docs, schemas, corpus, ref-links, template-drift, script-self-tests, skills-strict. `grep -c FAIL` over the captured output returned 0. The run's checks_baseline records exit 0 with every check passing, so the bar was "no findings beyond that" and there are none.

Targeted tests for the surface touched: `go test ./swarm/...` exits 0, including the e2e tier (180.795s).
```

**`agent-plugins-u9uh#1`** · `dev-bguq-3`

The defect was reproduced and watched fail before the fix, in all three modes.

```
evidence: cli/swarm/cli/threshold_test.go written first, run against the unmodified tree:

    $ go test ./swarm/cli/ -run 'TestReadyRefusesANonPositiveThreshold|TestReadyRefusesANegativeCompleted|TestReadyAcceptsTheValidEdge'
    --- FAIL: TestReadyRefusesANonPositiveThreshold (0.00s)
        threshold_test.go:50: [ready --run run-1 --threshold 0] exited 2, want 3 (malformed)
        threshold_test.go:53: [ready --run run-1 --threshold 0] read or wrote something before refusing: [[bd ready --parent run-1 --exclude-type epic --exclude-label human --limit 0 --json]]
        threshold_test.go:57: [ready --run run-1 --threshold 0] does not name the flag it refused
        threshold_test.go:60: [ready --run run-1 --threshold 0] does not say what a valid value is
    EXIT=1

The second line is the sharper one: exit 2 was reached by RUNNING the ready query. So the pre-fix command did not merely mis-handle the value, it read the tracker before doing so — which is what the AC "before the breaker reads or writes anything" is about. The test asserts from the fake's own call record rather than from the exit status, so a refusal that had already read something would not pass a status-only check.

After the fix the same run exits 0, and `go test ./swarm/cli/` exits 0 whole — the existing breaker tests in ready_test.go pass unchanged, which is the AC that a positive --threshold is unaffected.
```

**`agent-plugins-u9uh#3`** · `dev-bguq-3`

No consumer in the tree passes a value either guard now refuses, so the refusals cannot break a caller.

```
evidence: `grep -rn -- "--threshold" .` over the working tree returns five non-test sites, and none of them supplies a literal value:

- specs/architecture/container-swarm-cli.md:299 — "`--threshold` defaults to 3 and is a flag"
- plugin/skills/bead-bonanza/references/dispatching.md:192 — the form `swarm ready --run <epic> --completed <n> [--threshold <n>]`, with the prose "the threshold is a run-level authorization row (`sf.toml`'s `[run] dispatch_failure_threshold`, else 3)"
- sf.toml:163-165 — the row itself, commented out at its default: `# dispatch_failure_threshold = 3`, described as "how many consecutive dispatch returns that completed no bead stop the RUN"
- cli/swarm/cli/ready.go:307 — `const breakerDefaultThreshold = 3`
- docs/retros/2026-09-02-e6xt-swarm-dispatch.md — a prior run's record, not a caller

So the only configured path resolves to 3 or to an operator-supplied row, and a row of 0 has no coherent reading under the documented semantics: it would stop every run before its first dispatch returned. `--completed 0` is untouched — the guard is `completed < 0`, and TestReadyAcceptsTheValidEdgeOfEachFlag pins `--completed 0`, `--threshold 1` and an omitted `--completed` as accepted.
```

**`agent-plugins-u9uh#4`** · `dev-bguq-3`

The gates pass with no finding beyond the run's baseline.

```
evidence: `plugin/skills/bead-bonanza/scripts/bb-checks.py` with no arguments, run in the dispatch worktree, exited 0 with all thirteen declared checks present — format, vet, build, test-go, skills, shared-refs, strata-docs, schemas, corpus, ref-links, template-drift, script-self-tests, skills-strict — and `grep -c FAIL` over the captured output returned 0. The run's checks_baseline records exit 0 with every check passing, so the bar was "no findings beyond that" and there are none.

The change is 29 inserted lines in one function, `git diff --stat` → `cli/swarm/cli/ready.go | 29 +++`, plus the new cli/swarm/cli/threshold_test.go. applyBreaker, stoppedSummary and the dry-run refusal text are untouched, which the AC requires and the diff shows.
```

**`agent-plugins-bguq.6#3`** · `retro-bguq`

The run's validation was re-measured at the merged head rather than relayed, and the tracker is durable this run.

```
evidence: (1) declared checks at 36598cd on a clean tree - `plugin/skills/bead-bonanza/scripts/bb-checks.py` exit 0, all thirteen ids present (format, vet, build, test-go, skills, shared-refs, strata-docs, schemas, corpus, ref-links, template-drift, script-self-tests, skills-strict), zero FAIL lines. This closes the gap run bd97 recorded, that the checks had never been run over a merged head. (2) CI at the same head, via the check-runs API: ubuntu, macOS, cross-compile, no-runtime, skills and specs all `success`; `go test (windows-latest)` `failure`, job 100339119103, whose log ends `FAIL github.com/brettcallaghan/agent-plugins/cli/swarm/e2e 600.094s` - a package timeout, with five orphan processes terminated at cleanup (bd, dolt, git x2, env). That is agent-plugins-baih and is not this run's; what is new is that jed1 added a ninth e2e fixture to a suite already at its 600s ceiling, which makes baih's blocked `-timeout 20m` patch load-bearing. (3) nothing stranded: `git log origin/claude/swarm-dispatch-blueprint-5gob2e..HEAD` is empty. (4) `bd dolt push` -> "Push complete.", exit 0 - a change from run e6xt, whose record opens by saying it is the only durable copy of itself because every push failed.
```


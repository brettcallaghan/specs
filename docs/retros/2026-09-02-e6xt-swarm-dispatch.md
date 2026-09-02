# Run `agent-plugins-e6xt` — Swarm Dispatch Loop

The dispatch loop from `specs/features/swarm-dispatch/swarm-dispatch-blueprint.md`
(now `status: retired`), built as 19 work orders and 13 review fixes across
12 dispatch batches. Written by `sprint-retro` as `retro-e6xt`.

**113 findings harvested across 42 beads. 19/19 work orders closed, 13/13 loop
beads worked, 32 commits, nothing stranded, all 13 declared checks green.
13 candidates drafted — 0 filed, 0 declined, 13 AWAITING THE OWNER'S DECISION.**

## Read this first

### 1. This document is the only durable copy of the run

`bd dolt push` failed on **every** attempt for the whole run, including once
more while writing this record. Measured 2026-09-02 in this container, bd 1.2.2
(`6c124203e`):

```
$ bd dolt push
exit 1
Error: push to origin/main: Error 1105: To git+https://github.com/brettcallaghan/agent-plugins.git
 ! [rejected]            main -> main (non-fast-forward)
hint: Updates were rejected because the tip of your current branch is behind
hint: its remote counterpart.
```

The documented remedy does not clear it — `bd dolt pull` aborts with
`merge conflicts in issues require operator resolution`, and `bd dolt --help`
exposes no conflict-resolution subcommand (`agent-plugins-wgxd`, P1). So the
113 findings below, the 32 bead closures, and the 13 beads this run filed
exist **only in this container's local Dolt database**.

**The divergence predates this run, and it is bidirectional.** The previous
run's epics are absent from this database entirely:

```
$ bd export | jq -s '[.[] | select(.id|startswith("agent-plugins-bd97"))] | length'
0
$ bd export | jq -s '[.[] | select(.id|startswith("agent-plugins-ih18"))] | length'
0
```

— while `qozp` (30), `so39` (13), `mpnf` (28) and `y63n` (26), all older, are
present. The 2026-08-29 run's record is committed at
`docs/retros/2026-08-29-bd97-swarm-core.md`, so that run happened; its beads
reached the remote and never reached here. The local database is behind the
remote by at least one whole run **and** ahead of it by this one, and the two
cannot be merged with the tooling in the container.

What survives the container regardless: the 32 commits on
`claude/swarm-dispatch-blueprint-5gob2e`, each carrying one bead's code and its
close-commit subject, all pushed. What does **not**: every `retro:` comment,
every coordinator amendment, and the 13 beads filed in flight. Committing this
document is what makes them durable.

### 2. Every one of review pass 2's eight findings was introduced by loop 1's six fixes — and loop 2's seven fixes shipped unreviewed

Traced from each pass-2 bead's own `## Origin` section, by grepping its
description for the loop-1 commit shas:

| pass-2 finding | introduced by | which fixed |
|---|---|---|
| `e6xt.13` | `5c2495f` | `e6xt.7` — dry-run performed a real manifest write |
| `e6xt.14` | `4be01c8` | `e6xt.11` — `--completed` dropped on the `--agent` arm |
| `e6xt.15` | `4be01c8` | `e6xt.11` |
| `e6xt.17` | `4be01c8` | `e6xt.11` |
| `e6xt.18` | `4be01c8` | `e6xt.11` |
| `e6xt.16` | `0b8c2dd` | `e6xt.10` — `merged_sha` named a pre-rebase commit |
| `e6xt.20` | `0b8c2dd` | `e6xt.10` |
| `e6xt.19` | `a46658d` | `e6xt.9` — `verify-landed --dry-run` read a stale ref |

Eight findings from six fix commits, and **four of the eight from one of
them** (`4be01c8`, which added the circuit breaker to the `--agent` arm).
Nothing in loop 2 traces to loop 0's work: pass 2 was scoped to the loop diff
`86f9ec6..2622f6c` and found only what loop 1's fixes had put there.

The containment is visible in the diffs. Loop 1 touched 18 files; loop 2
touched **three** non-test files — `cli/swarm/cli/accept.go`,
`cli/swarm/cli/branch.go`, `cli/swarm/cli/ready.go` — and loop 1 had touched
all three.

Then the manifest guard refused `review_passes=3` against `review_budget=2`,
so **loop 2's seven commits (`e163c8d`..`37a9d83`) are merged and have never
been reviewed.** They are the same kind of change, in the same three files, by
the same mechanism that has so far produced 1.33 new findings per fix commit.
The budget arithmetic is filed as `agent-plugins-3vxe`; the seven unreviewed
commits are candidate **C13** below and are the only time-sensitive item in
this document — PR #33 is still a draft.

### 3. The e2e suite is green on three defects because its fixtures model the world after adoption

`agent-plugins-1vya` (P1), `agent-plugins-gqwl` (P1) and `agent-plugins-asv8`
(P2) are all live against the real harness and all invisible to
`go test ./swarm/e2e`. The mechanism is one thing:

```
$ grep -rn -A2 '\[\[worktree\]\]' cli/testdata/swarm-fixtures/*/fixture.toml
cli/testdata/swarm-fixtures/dispatch-worktree/fixture.toml:35:[[worktree]]
cli/testdata/swarm-fixtures/dispatch-worktree/fixture.toml-36-branch = "bb/dev-fixture-2"
```

Exactly one fixture in the suite declares a worktree, and it stands on a
`bb/<dispatch-id>` branch — the form `swarm dispatch open` will create *once
the coordinator adopts it*. The fixture says so itself: "The branch takes the
`bb/<dispatch-id>` form the coordinator creates, so a worktree checked out on
it is attributable to its dispatch by branch."

Today every real worktree is created by the harness and named
`worktree-agent-<hash>`, which carries no dispatch id, so the primary match
never fires and only the fallback — attribution through the bead the
worktree's HEAD commit closed — runs. That fallback is wrong for any dispatch
that has not committed yet, which is exactly when a coordinator needs the
answer. `gqwl` states the consequence plainly: `swarm ready` reports a bead
available while a live agent holds it, which is two agents on one bead.

`1vya` is the same shape by omission rather than by naming: the two fixtures
its concurrent arm runs against declare no worktree at all, so git never gets
the chance to refuse the rebase it refuses in the real world.

**The tested path and the executed path are disjoint**, and no check says so.
Candidate **C3**.

## The run

| | |
|---|---|
| Run epic | `agent-plugins-e6xt` — Run: Swarm Dispatch Loop |
| Target | `agent-plugins-o63s` — adopted whole, flat, 19 work orders. Now CLOSED |
| Source | `specs/features/swarm-dispatch/swarm-dispatch-blueprint.md`, groomed this session. Blueprint now `status: retired` |
| Branch | `claude/swarm-dispatch-blueprint-5gob2e` |
| Pull request | [#33](https://github.com/brettcallaghan/agent-plugins/pull/33), **still draft**, base `main` |
| Diff | `2490026`..`37a9d83` — 32 commits, one per bead. `git diff --shortstat` → 87 files, +12376/-720 |
| Queue | 19 admitted work orders; 13 review-loop beads worked; 1 review finding deferred out |
| Declared checks | 13, from `sf.toml`: `format`, `vet`, `build`, `test-go`, `skills`, `shared-refs`, `strata-docs`, `schemas`, `corpus`, `ref-links`, `template-drift`, `script-self-tests`, `skills-strict` |
| Checks baseline | **0 — all 13 passed on the untouched tree** (`checks_baseline_status: 0`) |
| Model lineup | `light`=haiku, `standard`=sonnet, `advanced`=opus, `frontier`=fable |
| Effort | **`unrecorded`** — the manifest key carries the literal string, so no issue this run files can state it |
| Concurrency cap | 20 |
| Circuit breaker | `consecutive_dispatch_failures: 0` — never tripped |
| Partition | 10 batches, 12 waves, max parallelism 3, over 7 collision edges (loop 0); one serial batch per review loop |

**Standing authorization**, all seven rows resolved from skill defaults — `sf.toml`'s `[run]` block sets none of them:

| row | value | source |
|---|---|---|
| `pull_request` | `draft` | default (note: PR #33 already open as draft; mark ready at end) |
| `move_foreign_beads` | `false` | default |
| `admit_review_findings` | `false` | default |
| `tier_escalation` | `once` | default |
| `review_budget` | `2` | default |
| `merge_to_default` | `false` | default |
| `dispatch_failure_threshold` | `3` | default |

**Collision edges:** `.beads/plans/agent-plugins-e6xt.edges.jsonl` (loop 0, 7 edges) and `.beads/plans/agent-plugins-e6xt.loop1.edges.jsonl` (loop 1, 5 edges). The loop-0 edges serialized nine command beads that share `cli/swarm/cli/root.go`'s `register` function; batch b4 alone carried five of them in a chain.

**Pre-flight ledger**, six entries, recorded on the manifest:

- `pre-flight-check.sh agent-plugins-e6xt` exited 0; verdict OK; all 13 checks passed on the untouched tree.
- `worktree.baseRef` resolves to `head` from `.claude/settings.json` (project scope). Measured: a subagent with `isolation:worktree` came up on session HEAD `5d82354`.
- `swarm` was absent from PATH and was built this session: `(cd cli && go build -o /usr/local/bin/swarm ./swarm)` exit 0. **See C4 — this build went stale during the run and nothing rebuilt it.**
- `bd swarm validate agent-plugins-o63s` exits 0, Swarmable YES, 9 waves, max parallelism 5, 19 work orders, 84 `blocks` edges.
- The 19 target beads carry `tier`, `tier_reason`, `est_size` and `c4_*` as real JSON arrays; none has a null tier.
- `c4_changes` is deliberately empty on the nine command beads: the blueprint lists Swarm Command Surface under relied-upon-unchanged. Real co-edit clusters are the `coedit_command` metadata key.

## Results

**19 of 19 admitted work orders closed**, each in one commit carrying its own closure.

| bead | agent | commit | what it landed |
|---|---|---|---|
| `o63s.1` TRK | `dev-e6xt-b1` | `d63a5bd` | Tracker + Beads Tracker operations for the dispatch loop |
| `o63s.2` VCS | `dev-e6xt-b2` | `750007a` | Version Control + Git Version Control writes; typed `*vcs.Refusal` |
| `o63s.3` MFK | `dev-e6xt-b3` | `d2e96a2` | manifest gains `consecutive_dispatch_failures` |
| `o63s.4` FIX | `dev-e6xt-b3` | `11377cd` | five new e2e fixture variants; a restore test that had never existed |
| `o63s.14` FOLD-1 | `dev-e6xt-b8` | `51fbfa1` | stage 1 folded into `container-swarm-cli.md` |
| `o63s.5` PC | `dev-e6xt-b4` | `6354aac` | `swarm phase close` |
| `o63s.6` QF | `dev-e6xt-b4` | `4304bd3` | `swarm ready` — the dispatch set |
| `o63s.9` BR | `dev-e6xt-b4` | `26bae8e` | `swarm dispatch open` — the sprint branch |
| `o63s.10` WT | `dev-e6xt-b6` | `b21e070` | `swarm dispatch open` — the worktree |
| `o63s.7` AQ | `dev-e6xt-b5` | `63fc7f6` | `swarm ready --agent` — one dispatch's queue |
| `o63s.8` CB | `dev-e6xt-b5` | `de82237` | `swarm ready` — the circuit breaker |
| `o63s.11` MG | `dev-e6xt-b4` | `8666780` | `swarm dispatch accept` — drain, case and merge |
| `o63s.12` LD | `dev-e6xt-b4` | `e2c4fb2` | `swarm branch verify-landed` and `swarm branch push` |
| `o63s.13` RM | `dev-e6xt-b7` | `cfd8a90` | `swarm dispatch accept` — ancestry, removal, landing |
| `o63s.15` FOLD-2 | `dev-e6xt-b8` | `8037887` | stage 2 folded into `container-swarm-cli.md` |
| `o63s.16` ADP-C | `dev-e6xt-b9` | `1b237ac` | coordinator adopts the dispatch commands |
| `o63s.17` ADP-D | `dev-e6xt-b9` | `701e2cd` | dispatch-side adoption — Developer and Recovery |
| `o63s.18` FOLD-3 | `dev-e6xt-b10` | `faedf95` | stage 3 folded into `module-bead-bonanza.md` |
| `o63s.19` DOD | `dev-e6xt-b10` | `86f9ec6` | Definition of Done verified, blueprint retired, `retro-spec` run |

**13 of 13 review-loop beads worked**, in two serial batches.

| bead | pass | agent | commit | what it fixed |
|---|---|---|---|---|
| `e6xt.8` P1 | 1 | `dev-e6xt-b11` | `b97acb5` | all five new commands stamped `command`/`run`/`mode` only on explain |
| `e6xt.7` P1 | 1 | `dev-e6xt-b11` | `5c2495f` | `swarm ready --dry-run` performed a real manifest write with `--completed` |
| `e6xt.9` P1 | 1 | `dev-e6xt-b11` | `a46658d` | `verify-landed --dry-run` reported landed from a stale tracking ref |
| `e6xt.11` P2 | 1 | `dev-e6xt-b11` | `4be01c8` | `swarm ready` silently dropped `--completed` with `--agent` |
| `e6xt.10` P2 | 1 | `dev-e6xt-b11` | `0b8c2dd` | `merged_sha` named a pre-rebase commit not on the target |
| `e6xt.12` P3 | 1 | `dev-e6xt-b11` | `2622f6c` | the drain check's two halves applied different filters |
| `e6xt.13` P2 | 2 | `dev-e6xt-b12` | `e163c8d` | `--dry-run --completed` hid a stop verdict the count established |
| `e6xt.16` P2 | 2 | `dev-e6xt-b12` | `c475255` | accept abandoned a merge that had already landed |
| `e6xt.17` P3 | 2 | `dev-e6xt-b12` | `7db8681` | a breaker failure discarded the queue answer already measured |
| `e6xt.14` P3 | 2 | `dev-e6xt-b12` | `2899ced` | the stop override kept the superseded resolution, dropped its summary |
| `e6xt.18` P3 | 2 | `dev-e6xt-b12` | `4e28d33` | `reportReady`'s doc comment asserted the opposite of the code |
| `e6xt.20` P3 | 2 | `dev-e6xt-b12` | `e935489` | `mergeCase.branchSHA` dead after `merged_sha` moved |
| `e6xt.19` P3 | 2 | `dev-e6xt-b12` | `37a9d83` | the unfetched-remote resolution named the wrong command |

**Ceremony epics:** `e6xt.1` pre-flight, `.2` scope, `.3` groom & plan — all
closed. `.4` review, `.5` retro, `.6` post-flight — open; `.5` is this
document and `.4`/`.6` are the coordinator's to close.

### The Definition of Done, and what it did not tick

`o63s.19` verified the blueprint's DoD against the tree rather than against
issue statuses, and **shipped two entries recorded NOT MET rather than
ticked** — which is the right outcome and worth naming:

- *"Every AC in the coverage table is verified at its declared level"* — all
  58 carry a test citation, **57 at their declared level**.
  `AC-SWARM-DISP-004.1` is declared e2e and is proved at unit, because a
  successful `git merge --ff-only` implies ancestry by construction and a
  failed one stops at `AC-SWARM-DISP-004.6` before the guard. Carried by
  `agent-plugins-oyvf`.
- *"Every prose section named in ADP-C and ADP-D is deleted"* — every named
  section is gone except step 3 of `dispatching.md`'s "Verify against state",
  the per-bead pre-merge acceptance check, which ADP-C deliberately kept
  because **no swarm command performs it**. Carried by `agent-plugins-bhaj`.

`retro-spec` also ran inside `o63s.19`, over 81 of the findings below, and
produced `rfcs/inbox/feature-blueprint/rfc-0024-swarm-dispatch-delta-decomposition-and-verification-tiers.md`
(severity `blocker`, `ready-for-review`). **That RFC is about the specs this
feature was built from; this document is about the run.** Where they touch —
the `decision` findings, below — this record cites it rather than restating it.

## Scope

`excluded: []` and `moved: []`. Nothing was excluded at admission and nothing
was moved into the run. The run finished its whole admitted set.

Two scope events, both recorded as amendments on the epic:

- **`agent-plugins-e6xt.2.1`** — the authorization-gate task, created by the
  scope step after `manifest.admitted` was written. Expected, not creep.
- **`agent-plugins-o63s.20`** ("Decode a refused push into the run's
  vocabulary") — filed `discovered-from` by `dev-e6xt-b2` while implementing
  VCS, and **deferred out rather than admitted**, consistent with
  `admit_review_findings=false` and with the blueprint scoping refusal
  decoding to merge and rebase only. It was then **reparented out of the
  target epic**: created inside `o63s`, it kept the target permanently
  un-closable — `swarm phase close --dry-run` correctly refused with
  `open_child: agent-plugins-o63s.20`. It keeps its status, its notes and its
  `discovered-from` edge.

**The run grew by 13 beads it did not start with**, all review findings, all
row 1 (defects the run's own diff introduced). The coordinator's stated rule:
`admit_review_findings=false` governs *scope growth from review*, and a defect
in the run's own output is not growth. One review finding — `e6xt.15` — was
held out under row 1's own carve-out, because its only remedy is prose in
shipped skill documents that have not yet misled a worker, and it was then
deferred because it sat unassigned and unblocked in
`bd ready --parent agent-plugins-e6xt --exclude-type=epic` beside batch b12's
head, reachable by a ready-set-driven dispatch.

## What the run decided

**42 `decision` findings, 13 `gap`, 3 `sequencing`, 4 `stale`, 1 `scope` — 63
of 113.** Each marks a place a document did not answer a question a bead
needed, and every one was answered by the implementing agent rather than
bounced.

The set splits cleanly by *which document was silent*, and the split is the
run's shape in miniature:

| where the silence was | findings | owner |
|---|---|---|
| the FRD, the blueprint, or an AC | 34 `decision` on `o63s.*` beads | `rfc-0024` |
| the reviewer-authored bead `Design` | 8 `decision` on `e6xt.*` beads | **nothing — candidate C7** |

`rfc-0024` carries the first half and this document does not restate it. The
second half is new and is not covered anywhere: four of loop 2's own agents
recorded, in almost the same words, that their bead's `Design` left a choice
open and that a diff-reader would not see it as a choice —
`e6xt.13#2` ("extracted into one helper rather than written twice, which a
diff-reader would not recognise as a choice"), `e6xt.17#2` ("two things the
Design left open were settled silently"), `e6xt.14#2`, `e6xt.9#2`.

### The gaps worth naming individually

- **`o63s.13#1` / `o63s.1#4` — "the acceptance signals" is used nine times in
  the FRD and enumerated nowhere.** `grep -n "signal"
  specs/features/swarm-dispatch/swarm-dispatch-frd.md` returns 9 hits, every
  one using the phrase without a set. The implementer derived it from
  `plugin/skills/bead-bonanza/scripts/bb-status.sh` — a shell script that is
  not part of the spec corpus, and one the container is not allowed to depend
  on. **The phrase "re-check every bead with no signal deferred" is
  unimplementable as written.** Two independent beads hit this and derived the
  same five signals; the derivation is sound and it is still a derivation.
- **`o63s.16#3` + `o63s.18#1` — an acceptance line ADP-C was structurally
  forbidden to satisfy.** ADP-C's AC required the Coordinator's roster line
  and Change Guidance to name the new commands; both live in
  `module-bead-bonanza.md`, the file ADP-C's own Design says not to touch.
  FOLD-3's acceptance criteria do not mention either. **The obligation reached
  FOLD-3 only through the dispatch prompt.** Had that prompt not carried it,
  both beads would have silently dropped it — and the dispatch prompt is the
  one surface nothing audits (candidate C8).
- **`o63s.11#1` / `o63s.12#2` / `o63s.12#5` — three Version Control
  operations the deltas required and the landed interface did not have**:
  `Checkout`, a push decoder that reads the two stderr lines carrying the
  rejection cause, and `CurrentBranch`. Each was found by measuring git rather
  than by reading a spec.
- **`o63s.5#4` / `o63s.6#2`** — two places the FRD is silent on a scope
  boundary (whether the run epic itself is in `phase close`'s scope; what
  happens when the Status Collector's own measurement fails mid-walk). Both
  decided conservatively and recorded.

### One `scope` finding, and it is the only one in the run

`o63s.16#3`, above. A single `scope` finding across 19 work orders is a
grooming result worth stating positively: the decomposition held.

## Review

**Two passes of a budget of two. `sf.toml`'s declared reviewer ran both times**
— `[review] skill = "code-review"`, fallback path 1, `--effort high`. Both
loops were admitted and worked; 14 findings were filed in total, 13 worked and
1 deferred.

### Pass 1 — scoped to the manifest's `base_sha`, 6 findings, all row 1

Scoped to `2490026`..`86f9ec6`, the run's whole diff. Six filed; a seventh
deduped to the already-open `agent-plugins-2ekn` and was not re-filed.

All six were row 1 — defects the run's own diff introduced — so all six were
in scope regardless of `admit_review_findings=false`, and three were AC-level
failures in what the run had just shipped.

**The dominant class was execution-mode fidelity, and it was not the theme the
run carried into the review.** Five of the seven (the six filed plus `2ekn`)
are a command behaving differently under `--dry-run` than under execute:

| finding | the command | what dry run did |
|---|---|---|
| `e6xt.7` | `swarm ready` | performed a real manifest write |
| `e6xt.9` | `swarm branch verify-landed` | reported `landed: true` off a stale ref; execute mode was correct |
| `2ekn` | `swarm dispatch open` | errored instead of previewing |
| `e6xt.13` (pass 2) | `swarm ready` | hid a stop verdict the recorded count established |
| `e6xt.8` | all five new commands | reported `mode: explain` after a real merge |

Each had a correct sibling in its own package — `refused(report, steps)` in
`manifest.go`, `closeEligible`'s `rec.Status != plan.StepRan` guard in
`phase.go`, `status.go:174`. The pattern existed and was applied
inconsistently, because the nine command beads were built in parallel by
different agents behind a collision chain. **That is a partitioning
consequence**, and it is the single strongest argument in this document for a
mechanical check rather than more prose (candidates C2 and, already filed,
`agent-plugins-8vco`).

### Pass 2 — scoped to the loop diff, 8 findings, all of them fix-induced

Scoped to `86f9ec6`..`2622f6c` rather than to `base_sha`. That was the right
call: pass 1's findings were closed, and re-reviewing them would have spent
the last pass on settled work. Cross-fix interference: none found — 13/13
checks, `go test ./...` green, the new e2e genuinely running (15.6s, both
arms) rather than skipping.

The origin trace is headline 2, above: **8 of 8 traced to loop 1's own six
commits.** Seven were admitted; `e6xt.15` was held out and later deferred.

The sharpest was `e6xt.13`, and its evidence is worth quoting because it is
the class the container exists to eliminate — same run, same mode, opposite
verdict:

```
swarm ready --run run-1 --dry-run                 -> exit 2, "the run is stopped"
swarm ready --run run-1 --dry-run --completed 0   -> exit 0, outcome "clear"
```

…and the second reports, in its own count fact: *"the write was refused: dry
run: the step writes. The count stands at 3 and would have become 4 (threshold
3)."* The command establishes the threshold is reached and then answers clear.
**The circuit breaker fails open in the mode meant to be safe to rehearse.**

### There was no pass 3, and seven commits are unreviewed

`review_passes` reached 2 of a declared budget of 2, and the manifest guard
refused a third. The prose the coordinator was following describes the budget
as loops rather than passes — `review_budget = 2` is commented in `sf.toml` as
"2 = three passes total". The guard counts passes. Filed as
`agent-plugins-3vxe` (P2).

The consequence is not arithmetic. `e163c8d`..`37a9d83` — seven commits —
merged with no review pass over them, in the three files that produced eight
findings the last time they were touched. Candidate **C13**.

### The reviewer files a problem, not a work order

Measured, with a control. Every one of the 14 review-filed beads carries a
rich prose description — What / Evidence / Why it matters / Origin /
Suggested shape / Acceptance. What none of them carried on arrival is the
tracker's **structured** fields, which is what a dispatch reads:

| bead | `design` | `acceptance_criteria` | `metadata.tier` | `metadata.c4_changes` |
|---|---|---|---|---|
| `e6xt.7`–`.12` (loop 1) | filled | filled | filled | filled |
| `e6xt.13`, `.14`, `.16`–`.20` (loop 2) | filled | **EMPTY** | filled | filled |
| **`e6xt.15`** (held out, never groomed) | **EMPTY** | **EMPTY** | **NULL** | **NULL** |

`e6xt.15` is the control: the one review-filed bead no planner ever touched
carries nothing in any of the four. Both loop planners had to supply Design,
tier and C4 before anything could be dispatched, and the loop-2 planner left
acceptance in the description prose rather than in the field. **That is a real
seam between `sprint-review` and `groom-beads`** — candidate C7.

## Inherited claims, acted on without measurement — three times, twice by the coordinator

`AGENTS.md` already carries the rule: *"Before you act on such a claim, or
restate it anywhere, run the command and read the output."* It was broken
three times this run, and the coordinator recorded all three itself.

**1 — the stale `swarm` binary (`o63s.16#1`).** ADP-C's job was to adopt prose
instructing a coordinator to run five new commands. Measured:

```
/usr/local/bin/swarm (mtime Sep 1 07:22)
$ swarm --help   ->  Available Commands: manifest, status
```

The binary on PATH lacked every command the prose it was adopting instructs a
coordinator to run. The agent built from source to measure them.

**2 — a false measurement published from that binary, by the coordinator.**
Correcting another agent's flagged concern about `swarm ready --agent`'s exit
status, the coordinator probed `/usr/local/bin/swarm`, got exit 0, and wrote
onto `e6xt.11`: *"the `--agent` arm HONOURS `--completed` but does NOT take the
stop status… run-batch is unaffected."* Retracted the same night. The probe had
run the pre-fix build — the very build in which `--completed` was dropped on
that arm, which is the bug `e6xt.11` fixed. Re-measured after rebuilding:
exit 2 both with and without `--completed`, and `cli/swarm/cli/ready.go:542`
sets `OutcomeStop` in `reportAgentQueue` unconditionally. **The original
concern stands**, and is carried by `agent-plugins-e6xt.15`.

**3 — a planner's unmeasured claim relayed into a dispatch prompt, by the
coordinator.** The loop-2 planner's pinned Design for `e6xt.16` asserted that
`mustRecord` panics on the failure path. The coordinator restated it verbatim
in batch b12's dispatch prompt — *"the planner flagged a hazard the bead itself
missed: the established path cites via `mustRecord`, which PANICS on the
failure path"* — and ran nothing. The implementing agent measured it
(`e6xt.16#2`): `mustRecord` discards the `ok` and returns the **zero** Record,
and `argvOf` guards `rec.Invocation.Name == ""` and returns nil. No index, no
dereference, no map access. The real failure mode is a *silent* one — a nil
argv passed as one variadic invocation, attributing the fact to an empty
invocation rather than to none — and the fix is correct on its own merits
either way.

**Two of the three travelled through a stale binary; two of the three
travelled through a surface nothing audits.** `AGENTS.md`'s own table of
belief-carrying surfaces names the durable trigger prompt as the worst and the
dispatch prompt as "the same defect one tier less durable… audited by nobody;
it is not in the repo and no check sees it." Both of the coordinator's failures
went onto one. The prompts died with their agents; this document and the beads'
own findings are the durable record.

Two unreached copies were recorded rather than fixed, and remain unreached:
`e6xt.16`'s **pinned Design in the tracker** still asserts the panic, and
**batch b12's dispatch prompt** cannot be edited. `grep -rn "mustRecord"
--include=*.go --include=*.md .` returns 11 hits, all Go call sites plus the
definition and its doc comment — **nothing in the repository carries the false
claim.**

Candidates C4 (the probe cannot detect a binary the run itself made stale) and
C8 (nothing audits a dispatch prompt).

## Routing

**113 findings, 42 beads, and none of them unrouted.**

| kind | count | destination |
|---|---|---|
| `decision` | 42 | project |
| `gap` | 13 | project |
| `stale` | 4 | project |
| `sequencing` | 3 | project |
| `unverifiable` | 3 | project |
| `scope` | 1 | project |
| **project subtotal** | **66** | |
| `tooling` | 7 | skill |
| `proof` | 30 | terminates — the record, beside what it supports. No issue |
| `environment` | 10 | terminates — the environmental-failure ledger. No issue |
| **unrouted** | **0** | |

**Zero unrouted is the first in five runs.** `qozp` reported 2, `so39` 2,
`mpnf` 2, `bd97` 3 and rising. Every finding this run carries a kind from the
nine-kind taxonomy and no agent invented one — no `retro:process:`, no
`retro:incident:`, no `retro:work-clean`. Two findings put a word in the
`target` slot that is not a document or a skill (`retro:stale:project` ×2,
`retro:tooling:skill` ×1), which costs a cluster and not a route.

**`sf.toml` sets no `[retros] skill_issues_dir`**, deliberately: this
repository vends the skills this run exercised, so a `tooling` finding is
ordinary work in this tracker. Both audiences file as beads here. Nothing goes
to an outbox.

### The `environment` findings that belong in `AGENTS.md`'s ledger

These terminate as findings and are worth promoting into the ledger, because
the next agent will hit them:

- **A sibling dispatch's `bd dolt pull` blocks every `bd` call in the
  container**, with two faces depending on timing: the first invocation hangs
  with no timeout, a later one fails with `the database is locked by another
  dolt process`. Measured twice independently. `pre-flight-check.sh:608` runs
  `bd where` unconditionally, so **a project check — not just bead ceremony —
  stalls on it**; `script-self-tests` sat ~22 minutes with no indication why.
  Uncontended, the failing pull takes 4 seconds; the 22 minutes was all lock.
- **`bd init` commits its own scaffolding onto the checked-out branch.** Every
  restored e2e world had its sprint branch pointing at a `bd init` commit and
  its index carrying the same files, which made `git rebase` refuse. Fixed in
  the harness with a mixed reset back to the pre-`bd init` head.
- **The e2e harness pinned `HOME` but no git identity**, so the first case to
  write a commit could not: `fatal: unable to auto-detect email address`.
  Fixed by adding `GIT_AUTHOR_*`/`GIT_COMMITTER_*` to `world.env()`.
- **`bd` refuses a task→epic `blocks` edge in both directions.** Only
  epic-blocks-epic is accepted, and `bd ready` walks it transitively — that is
  the only mechanism by which closing one phase epic releases another phase's
  beads.
- **`bd show --json` carries `comment_count` and no `comments` key**, so
  findings are reachable only through `bd comments`. See C1: this is also why
  `sprint-retro`'s own documented harvest query returns zero.

## Validation

**Baseline — all 13 declared checks passed on the untouched tree**, exit 0,
recorded verbatim on the manifest as `checks_baseline` with
`checks_baseline_status: 0`.

**At head — re-measured for this record rather than inherited.** Run in the
foreground at `37a9d83`, exit status read directly and not through a pipe:

```
$ python3 plugin/skills/bead-bonanza/scripts/bb-checks.py
exit 0
=== format / vet / build / test-go / skills / shared-refs / strata-docs /
    schemas / corpus / ref-links / template-drift / script-self-tests /
    skills-strict
check-skills: 0 error(s), 0 warning(s)
shared-refs:  all 33 vendored references are current
strata-docs:  all 3 vendored strata docs match /root/.cache/strata-src
```

**13 of 13 green at head.** This matches what the coordinator declared after
every merge, and is now measured rather than relayed.

**CI has never run on this branch.** Every check run on PR #33 reports
`conclusion: "skipped"`:

| check run | conclusion |
|---|---|
| `go test (${{ matrix.os }})` | skipped |
| `cross-compile the release matrix` | skipped |
| `runs with no language runtime installed` | skipped |
| `this repository's own specs conform` | skipped |
| `skill invariants and vendored references` | skipped |

That is **by design, not a defect**: both `.github/workflows/checks.yml` and
`cli.yml` carry an `if` on each job that skips a draft pull request, and PR #33
is still a draft — *"a push to a working branch runs nothing, and neither does
a draft pull request."* Taking the PR out of draft fires `ready_for_review` and
starts the run.

**The consequence is still real and belongs on the record:** the whole run's
validation rests on 13 local checks on one Linux container. The three-platform
test matrix, the cross-compile matrix and the busybox smoke test — the checks
that only mean something across the repository — have **not** executed against
any of these 32 commits. Post-flight owns marking the PR ready; until it does,
nothing has compiled this code on macOS or Windows.

## Pushes

**Nothing is stranded.**

```
$ git log --oneline origin/claude/swarm-dispatch-blueprint-5gob2e..HEAD
(empty)
$ git status --porcelain
(empty)
```

All 32 commits are on the remote. PR #33's head sha is `37a9d838`, matching
local HEAD, and `mergeable_state` is `clean`.

**The tracker is the opposite.** `bd dolt push` → exit 1, non-fast-forward, on
every attempt of the run and once more while writing this. See headline 1.

## What the run could NOT verify

Named individually, because an unstated gap is indistinguishable from one that
does not exist.

1. **Loop 2's seven commits have never been reviewed.** `e163c8d`..`37a9d83`.
   The budget guard refused pass 3. Candidate C13.
2. **CI has never executed on this branch** — five check runs, all skipped,
   because the PR is a draft. No macOS leg, no Windows leg, no cross-compile,
   no busybox smoke test.
3. **`AC-SWARM-DISP-004.1` is not verified at its declared level.** Declared
   e2e, proved at unit. The state it describes — a dispatch branch that is not
   an ancestor of the target *after* a merge step that did not stop — cannot be
   constructed in a single-actor world. What would settle it: a fixture world
   carrying a real repository-local `post-merge` hook that moves the target,
   the technique `push-refused` already uses for a rejection. Carried by
   `agent-plugins-oyvf`.
4. **The pre-merge acceptance check (FRD ceremony step 2) is implemented
   nowhere and owned by nobody.** No delta's behavior lines cover it, and
   `dispatching.md`'s step 3 — the only place it was described — was kept with
   a line saying no command performs it. Carried by `agent-plugins-bhaj`.
5. **`e6xt.19`'s second corrected resolution is unasserted.** The
   divergence-step verdict cannot be reached against the `exectest` fake: the
   divergence step shares a plan with the two reads before it, so the state
   needs a plan that stopped between them, and the fake replays declared argvs
   rather than truncating a plan. The correction rests on two resolutions
   sharing one `verb` local four lines apart, not on a test. Deliberately not
   filed — test infrastructure for a path the bead did not introduce.
6. **`agent-plugins-1vya`, `gqwl` and `asv8` have no failing test.** All three
   are measured against the live harness and all three are green in e2e. No
   fixture reproduces any of them. Candidate C3.
7. **The manifest's `effort` reads the literal string `unrecorded`**, so no
   issue this run files can state the effort level it was produced at — a
   required provenance field on the issue template. `models` is complete;
   `effort` is not.
8. **The previous run's beads cannot be read from this container**, so
   dedupe against `bd97`'s eight candidates was done against the committed
   retro document rather than against the tracker. Two of them recur below
   (C5, C6) and one is refined (C4); if any of the eight was filed as a bead
   after that run, this container cannot see it.
9. **No consumer project was run against these skills.** As in the previous
   run, the claim that the vended skills work for a project other than this one
   remains untested — there is no second project in this session.

## Candidates — your decision

**13 candidates. Nothing has been filed.** `sprint-retro` files only what a
human accepts, and this run's retro had no human in the loop. Each carries the
issue template's frontmatter plus its three sections; a declined candidate
stays in this document and raises its recurrence count next time.

`model` on every candidate is `claude-opus-5`. `effort` cannot be stated — see
could-not-verify item 7. `run` is `agent-plugins-e6xt`, `branch` is
`claude/swarm-dispatch-blueprint-5gob2e`, `report` is this file.

**Read C13 first if you read only one.** It is the only one that expires: it
must be decided before PR #33 leaves draft.

---

### C13 — Seven merged commits have never been reviewed, and the file they touch has a measured defect rate

- **audience** project · **target** run `agent-plugins-e6xt` / PR #33 · **priority** 1
- **findings** — headline 2 above · **recurrence** 0 · **complements** `agent-plugins-3vxe`

**What is wrong, and what it costs.** `e163c8d`..`37a9d83` — loop 2's seven
fix commits — merged without a review pass, because the manifest guard refused
`review_passes=3` against `review_budget=2`. They change
`cli/swarm/cli/ready.go`, `accept.go` and `branch.go`, the same three files
loop 1 changed. Loop 1's six commits in those files produced **eight** review
findings, two of them P2 correctness defects in the circuit breaker. There is
no reason to expect loop 2's seven to be different in kind, and the run shipped
them with the reviewer's budget exhausted rather than with a clean pass. Cost:
whatever is in there merges to `main` unexamined.

**Evidence.** `agent-plugins-3vxe` (P2, filed) — *"review_budget counts passes
in the guard and loops in the prose, so the last loop ships unreviewed."*
Origin trace, headline 2: 8 of 8 pass-2 findings name a pass-1 commit.
`git diff --stat 2622f6c 37a9d83 -- '*.go'` → three non-test files, all three
touched by loop 1.

**Proposed change.** Before PR #33 leaves draft, run one more `code-review`
pass scoped to `2622f6c..37a9d83`, file what it returns, and decide each
finding explicitly — work it, or ship it as a filed bead with its measurement.
This is a one-off action on this run, not a change to the skill;
`agent-plugins-3vxe` is the durable half.

---

### C1 — `sprint-retro`'s documented harvest query returns zero findings on bd 1.2.2

- **audience** skill · **target** `sprint-retro`, and `specs/standards/findings-capture-protocol.md` · **priority** 0
- **findings** `agent-plugins-o63s.1#4` · **recurrence** 0

**What is wrong, and what it costs.** Both `sprint-retro/SKILL.md` §2 and the
Harvesting section of `findings-capture-protocol.md` prescribe
`bd export | jq -r 'select(...) | .comments[]? | select(.text | startswith("retro:")) | .text'`
as *the* harvest, explicitly preferring it over `bd list` + `bd show --json`.
On bd 1.2.2 `bd export` emits **no `comments` key at all** — only
`comment_count`. The query is not merely incomplete; it returns nothing. The
only reason this run's 113 findings were harvested is that the skill's own
"a zero count across a drained run is a broken query" rule fired and the
harvest was rebuilt by hand. A retro that trusted the documented query would
have written a complete-looking document reporting **zero findings on a run
that produced 113**, and — given headline 1 — that document would then have
been the only surviving record.

**Evidence.**

```
$ bd export | jq -r 'keys | join(", ")' | head -1
_type, assignee, close_reason, closed_at, comment_count, created_at, created_by,
dependencies, dependency_count, dependent_count, description, id, issue_type,
labels, owner, priority, status, title, updated_at
$ bd export | jq -s '[.[] | select(.comment_count > 0)] | length'
292
$ bd export | jq -s '[.[] | .comments[]?] | length'
0
```

- `agent-plugins-o63s.1#4` — measured the same thing from the other side while
  building the TRK delta: *"`bd show agent-plugins-qozp.1 --json | jq -c '.[0] |
  {comment_count, has_comments_key: (has("comments"))}'` → `{"comment_count":1,
  "has_comments_key":false}`. `bd show` carries a count and no comments, so the
  notes are only reachable through `bd comments`."* TRK ships
  `Findings(name, id)` over `bd comments <id> --json` for exactly this reason —
  the container already knows what the skill does not.

**Proposed change.** Replace the harvest in both documents with the per-bead
form the container itself uses:

```sh
for id in $(bd list --parent <run epic> --status all --json | jq -r '.[].id'); do
  bd comments "$id" --json | jq -r --arg id "$id" \
    '.[] | select(.text|startswith("retro:")) | "\($id)\t\(.text)"'
done
```

and keep the "zero is a broken query" rule, which is what saved this run.
`bd list` **must** carry `--status all` — see C11.

---

### C4 — The pre-flight capability probe cannot detect a binary that the run itself makes stale

- **audience** skill · **target** `bead-bonanza` / `scripts/pre-flight-check.sh` · **priority** 1
- **findings** `agent-plugins-o63s.16#1`, plus the coordinator's retraction on `e6xt.11` · **recurrence 1** (`bd97` candidate 2, `agent-plugins-bd97.4.9` — never filed)

**What is wrong, and what it costs.** `pre-flight-check.sh:635` classifies the
`swarm` binary by running `swarm status --help` and reading the exit status. A
binary that answers that one subcommand is `ok`, however far behind the tree it
is. This run built `swarm` at session start, passed pre-flight (`Verdict: OK`),
and then **spent the next nine hours adding four new subcommands to it**.
Nothing rebuilt it, and nothing could have noticed. Cost, both measured this
run: an adoption bead had to build from source to measure the commands it was
documenting, and the **coordinator published a false measurement into the
tracker** from the stale build while correcting another agent — the highest-
confidence moment `AGENTS.md` names as the least gated.

**Evidence.**

- `agent-plugins-o63s.16#1` — *"the `swarm` binary on PATH predates every
  command this item adopts, so a coordinator following the new prose would get
  'unknown command'. `/usr/local/bin/swarm` (mtime Sep 1 07:22) `swarm --help`
  lists only: manifest, status."*
- Coordinator retraction on `e6xt.11`, 2026-09-01T23:46Z — *"I measured a STALE
  BINARY… my earlier run returned exit 0 only because the old binary never
  applied the breaker on that arm at all."*
- The probe, unchanged since the previous run recorded it:
  `pre-flight-check.sh:635` — `SWARM_HELP_ERR="$(swarm status --help 2>&1 >/dev/null)"; SWARM_HELP_RC=$?`
- Prior run, `docs/retros/2026-08-29-bd97-swarm-core.md` headline 1 —
  *"`agent-plugins-bd97.4.9` — the capability probe tests only `swarm status
  --help`, so a binary behind on manifest keys passes it, and `session-start`'s
  rebuild guard then preserves that stale binary."* Verified live in that
  container too. It was one of eight candidates, **none of which was filed.**

**Proposed change.** The previous run proposed a better probe. This run shows
that is not enough, because the binary was *correct at probe time*. Add a
rebuild to the merge step: after each dispatch is accepted, when the run's diff
touches `cli/swarm/**`, rebuild `swarm` from the working tree before the next
command is run. A rebuild is seconds and is idempotent; a stale binary put a
false measurement in the tracker.

---

### C2 — Nothing tests a command's dry-run mode against its execute mode

- **audience** project · **target** `cli/swarm/cli` · **priority** 1
- **findings** review beads `e6xt.7`, `e6xt.9`, `e6xt.13`, `agent-plugins-2ekn` · **recurrence** 0

**What is wrong, and what it costs.** Four of the run's fifteen review findings
are a command doing the wrong thing **only** under `--dry-run`, in four
different commands, and one of them is the circuit breaker failing open. Every
one of them passed every declared check and every e2e case, because no test
compares a command's two modes. `agent-plugins-8vco` (filed) proposes
validating a rendered result against the generated JSON Schema; that catches
the `e6xt.8` stamping class and **none of these four** — a dry run that
performs a real write, or answers `clear` where execute answers `stop`, emits a
perfectly schema-valid document.

**Evidence.**

- `e6xt.7` P1 — *"`swarm ready --dry-run` performs a real manifest write when
  `--completed` is given."*
- `e6xt.9` P1 — *"`verify-landed --dry-run` reports landed from a stale tracking
  ref"*; execute mode was correct.
- `e6xt.13` P2 — measured: `swarm ready --run run-1 --dry-run` → exit 2 "the run
  is stopped"; `… --dry-run --completed 0` → exit 0 "clear", while its own count
  fact says the threshold is reached.
- `agent-plugins-2ekn` P3 — *"`dispatch open --dry-run` errors instead of a clean
  preview when the sprint branch does not yet exist."*

**Proposed change.** Add one table-driven test in `cli/swarm/cli` that, for
every registered command, runs a representative invocation in dry-run and in
execute against the same fake, and asserts two invariants: **(a)** the dry run
records no step whose `Status` is `StepRan` on a writing step, and **(b)** the
two runs agree on `outcome` and exit status wherever the execute run's writes
did not change the answer. Drive it off the live command tree the way
`e6xt.8`'s guard test already enumerates it, so a new command is covered by
existing.

---

### C3 — No e2e fixture models the world the code actually runs in

- **audience** project · **target** `cli/testdata/swarm-fixtures/` and `container-swarm-cli.md`'s Change Guidance · **priority** 1
- **findings** `agent-plugins-o63s.11#7`, `o63s.13#6`; beads `1vya`, `gqwl`, `asv8` · **recurrence** 0

**What is wrong, and what it costs.** Exactly one fixture declares a worktree,
and it stands on `bb/dev-fixture-2` — the branch form `swarm dispatch open`
creates once the coordinator adopts it. Every worktree that exists **today** is
harness-created and named `worktree-agent-<hash>`. So the primary attribution
path is the only one tested and the fallback path is the only one executed.
Three defects live in that gap right now, two of them P1, and all three are
green in `go test ./swarm/e2e`. `gqwl`'s consequence is a coordinator
dispatching a bead a live agent already holds.

**Evidence.**

- `cli/testdata/swarm-fixtures/dispatch-worktree/fixture.toml:36` —
  `branch = "bb/dev-fixture-2"`, with the assumption stated in its own comment:
  *"The branch takes the `bb/<dispatch-id>` form the coordinator creates, so a
  worktree checked out on it is attributable to its dispatch by branch."*
- `agent-plugins-gqwl` — *"its fixtures stand up worktrees on `bb/<dispatch-id>`
  branches… so the PRIMARY match succeeds and the fallback never runs. In the
  real harness today EVERY worktree is created by the harness with a
  `worktree-agent-<hash>` branch, so only the broken fallback runs. **The tested
  path and the executed path are disjoint.**"*
- `agent-plugins-o63s.11#7` — *"`grep -l worktree
  cli/testdata/swarm-fixtures/dispatch-stale-base/fixture.toml
  dispatch-rebase-conflict/fixture.toml` matches neither"* — so the concurrent
  arm's own fixtures declare no worktree at all.
- `agent-plugins-o63s.13#6` — *"A fixture that is both concurrent AND has a
  standing worktree would reproduce `1vya` end to end today. None exists."*

**Proposed change.** Add one fixture — a concurrent world (target ahead of the
dispatch branch) with a standing worktree whose branch is
`worktree-agent-<hash>`, carrying no dispatch id. It reproduces `1vya`, `gqwl`
and `asv8` in one world. Then add a line to `container-swarm-cli.md`'s Change
Guidance: **a fixture SHALL model the world the code runs in today; where it
models a world after an adoption, its `fixture.toml` SHALL say so and name the
adoption it presumes.** The existing `dispatch-worktree` fixture already says
so in a comment — make that obligatory rather than incidental.

---

### C5 — `bd-health.sh`'s repo-global exit status cannot gate one epic

- **audience** skill · **target** `groom-beads` / `scripts/bd-health.sh` · **priority** 2
- **findings** the coordinator's readiness pass, confirmed by four agents · **recurrence 1** (`bd97` candidate C1 — never filed; `rfcs/retired/rfc-0003` earlier)

**What is wrong, and what it costs.** `groom-beads/SKILL.md` says *"Do not
declare the set ready while it exits non-zero."* On this repository it exits
non-zero unconditionally, so the instruction is unsatisfiable and the gate
becomes something every run learns to ignore. Re-measured for this record:

```
$ bash plugin/skills/groom-beads/scripts/bd-health.sh   ; echo $?
1
  ✗ swarm validate     agent-plugins-qozp.6 (exit 1)
      Error: 'agent-plugins-qozp.6' is not an epic or molecule (type: task)
  … 16 more, every id foreign to this run …
RESULT: hard problems found in: swarm validate

$ bd swarm validate agent-plugins-o63s   ; echo $?
0
🐝 Swarm Analysis: Swarm Dispatch Loop … Total issues: 19 (19 closed)
```

`run_swarm_validate()` enumerates every active issue repository-wide and feeds
each to an epic-only validator. A non-zero exit means "some unrelated bead in
this repository is a task", not "this queue's graph is unsound".

**Evidence.** The commands and outputs above, run at `37a9d83`. Recurrence:
`docs/retros/2026-08-29-bd97-swarm-core.md` §C1 recorded the identical
measurement — exit 1, `RESULT: hard problems found in: swarm validate`, every
named id foreign, control `bd swarm validate <epic>` → exit 0.

**Proposed change.** Scope the check to the queue: have `bd-health.sh` take an
optional epic argument and validate only that epic and its descendants, and
have `groom-beads` pass the run's epic. Absent an argument, keep the repo-wide
sweep but report it as advisory rather than as a hard problem.

---

### C6 — A fixed byte budget on a skill file makes a correctness edit compete with existing prose

- **audience** skill · **target** `scripts/check-skills.sh`, `run-batch` · **priority** 2
- **findings** `agent-plugins-o63s.17#1` · **recurrence 1** (`bd97` candidate C6 — never filed)

**What is wrong, and what it costs.** `check-skills.sh` fails a skill whose
body exceeds `BODY_BYTE_MAX = 20000`. Measured now:

```
$ python3 -c "…strip frontmatter…"   plugin/skills/run-batch/SKILL.md
body 19993 bytes        # 7 bytes of headroom
```

ADP-D's job was to add one true sentence to that file — that
`swarm ready --agent` can now exit 2. It went over, and the agent trimmed three
separate paragraphs, twice, to get back under. **The budget makes a
correctness edit bid against unrelated prose, and the cheapest way to satisfy
it is to delete something else that is also true.** Cost this run: an agent's
time spent editing to a byte count, and a file now seven bytes from refusing
the next correction — including the one `agent-plugins-e6xt.15` is waiting to
make.

**Evidence.** `agent-plugins-o63s.17#1` — *"After the `swarm ready --agent`
adoption as first written: `scripts/check-skills.sh --strict` → exit 1, `WARN
plugin/skills/run-batch/SKILL.md: body is 20228 bytes (~5057 tokens)`, RESULT:
FAIL. After trimming: 19994 bytes, exit 0. Six bytes of headroom."* (This
record measures 19993/7; the one-byte difference is a trailing-newline
convention and does not change the finding.) Recurrence:
`docs/retros/2026-08-29-bd97-swarm-core.md` §C6, same defect, different file.

**Proposed change.** Keep the cap and make exceeding it a *reviewable* event
rather than a build failure: allow a skill to declare a budget override in its
frontmatter (e.g. `body_budget: 21000`) which `check-skills.sh` honours and
reports, so growing a skill is a visible decision instead of a silent trim.
`run-batch` is over its natural size and the next correctness edit should not
have to pay for that.

---

### C7 — `sprint-review` files a problem; the tracker fields a dispatch reads stay empty

- **audience** skill · **target** `sprint-review` · **priority** 2
- **findings** `e6xt.13#2`, `e6xt.14#2`, `e6xt.17#2`, `e6xt.9#2`; control `e6xt.15` · **recurrence** 0

**What is wrong, and what it costs.** Review findings arrive as excellent prose
and as empty structured fields. `e6xt.15` — the one review-filed bead no
planner ever groomed — carries `design` EMPTY, `acceptance_criteria` EMPTY,
`metadata.tier` NULL and `metadata.c4_changes` NULL. Both loop planners had to
supply Design, tier and C4 for every admitted finding before anything could be
dispatched, and the loop-2 planner still left `acceptance_criteria` empty on
all seven of its beads. Cost: a full grooming pass inside each review loop that
the run's shape does not budget for, and — because the planner writes the
Design under time pressure rather than the reviewer who did the analysis —
**eight `decision` findings in which an implementing agent records that its
bead's Design left a choice open and it settled it silently.**

**Evidence.**

- Field census, all 14 review-filed beads: loop 1 (`.7`–`.12`) design+AC
  filled; loop 2 (`.13`–`.20`) design filled, AC empty; `.15` all four empty.
- `e6xt.17#2` — *"two things the Design left open were settled silently, and a
  diff-reader would not see either as a choice."*
- `e6xt.13#2` — *"the stop sentence was extracted into one helper rather than
  written twice, which a diff-reader would not recognise as a choice."*
- `e6xt.14#2` — *"the invariant was shipped as a TEST, not as a
  `Result.Validate` rule… The bead left this open."*
- `e6xt.9#2` — *"the fetch guard runs BEFORE the counterpart check, and it also
  fires on a fetch that ran and FAILED — two points the bead Design left open."*

**Proposed change.** `sprint-review` writes the reviewer's own "Suggested
shape" into the bead's `design` field and its "Acceptance" into
`acceptance_criteria`, rather than leaving both inside the description prose —
the reviewer already produces both, and it is the only actor that has read the
code. Say explicitly in the skill that a review-filed bead is **not** Ready and
still passes through grooming for tier and C4; today the skill leaves that
implicit and the planner discovers it.

---

### C8 — Nothing audits a dispatch prompt, and it carried two of this run's three inherited-claim failures

- **audience** skill · **target** `bead-bonanza` dispatch policy · **priority** 2
- **findings** `e6xt.16#2`, the coordinator's acknowledgement on `e6xt.16`, `o63s.18#1` · **recurrence** 0

**What is wrong, and what it costs.** `AGENTS.md`'s own table of
belief-carrying surfaces says of the dispatch prompt: *"dies with its agent —
but goes to several agents at once… audited by nobody; it is not in the repo
and no check sees it."* This run supplies the cost. A planner's unmeasured
claim ("`mustRecord` PANICS on the failure path") was restated verbatim into
b12's dispatch prompt and shipped to an agent as a hard constraint; measured, it
is false. Separately, ADP-C's unsatisfiable acceptance line reached FOLD-3
**only** through a dispatch prompt — had that prompt not carried it, both beads
would have dropped it silently and the module blueprint would have shipped
contradicting itself. So the same unaudited surface both **carried a false
claim** and **carried the only copy of a true obligation**.

**Evidence.**

- `e6xt.16#2` — the full measurement retracting the panic claim, with
  `mustRecord`'s and `argvOf`'s source quoted, and the sweep
  `grep -rn "mustRecord" --include=*.go --include=*.md .` → 11 hits, none
  asserting a panic.
- Coordinator acknowledgement on `e6xt.16` — *"I RESTATED IT VERBATIM in this
  batch's dispatch prompt… and I did not run anything to check it… That is the
  THIRD inherited-claim failure in this run and the second of mine."*
- `o63s.18#1` — *"Nothing in FOLD-3's own acceptance criteria mentions the
  roster line or Change Guidance; the obligation reached this bead only through
  the dispatch prompt."*

**Proposed change.** Two lines in the dispatch policy. **(a)** A dispatch
prompt SHALL NOT assert a tool or code behaviour it does not carry the command
and output for; where the coordinator is relaying a claim it has not run, it
marks it `UNMEASURED —` and says what would settle it. **(b)** An obligation
that a dispatch must satisfy SHALL be on the bead, not only in the prompt: if
the coordinator finds itself writing a requirement into a prompt, that is a
bead edit it has not made.

---

### C9 — A `bd comment` written through a double-quoted shell string executes its own evidence

- **audience** project · **target** `AGENTS.md` · **priority** 2
- **findings** the coordinator's correction comment on `agent-plugins-e6xt` · **recurrence** 0

**What is wrong, and what it costs.** The run's most important review-loop
decision — the LOOP DECISION comment for pass 2 — was written through a
double-quoted bash string containing backticks. Bash executed the backticked
text as command substitution and pasted the **output of a real invocation**
into the comment in place of the evidence. The corruption is silent, it lands
in the tracker, and the damaged region reads exactly like evidence: a
plausible-looking YAML fragment with `contract_version`, `command`, `run`,
`mode`. It took a second comment to say "ignore that; here is what I meant".
`AGENTS.md` names the neighbouring trap ("Pipes hide exit statuses") and not
this one.

**Evidence.** The correction comment on `agent-plugins-e6xt`,
2026-09-01T23:47:44Z — *"I wrote that comment through a double-quoted bash
string containing backticks. Bash executed the backticked text as command
substitution and pasted the OUTPUT into the comment… A shell that executes your
evidence is the same class of error as a pipe that eats your exit status —
`AGENTS.md` names the second and not yet the first."*

**Proposed change.** Add to `AGENTS.md`, beside "Pipes hide exit statuses":

> **A double-quoted shell string executes your evidence.** Backticks and `$(…)`
> inside `"…"` are expanded before the command sees them, so evidence pasted
> into a `bd comment`, a commit message or a bead description is silently
> replaced by the output of running it. Pass the text as an argument list, a
> quoted heredoc (`<<'EOF'`), or `--file`. Never through `"…"`.

The same sentence belongs in `findings-capture-protocol.md`'s "Where to
record", which is the document that tells an agent to run `bd comment`.

---

### C10 — A `pgrep -f` wait predicate matches itself, and two shells spun for 46 and 42 minutes

- **audience** project · **target** `AGENTS.md` · **priority** 3
- **findings** the coordinator's live reproduction on `agent-plugins-wgxd` · **recurrence** 0

**What is wrong, and what it costs.** Two background shells launched as waits
by wave-1 dispatch agents were still spinning 46 and 42 minutes after those
agents had exited:

```
PID 31804, 46:15  until [ -s .../pull1.txt ] && ! pgrep -f "bd dolt pull" >/dev/null; do sleep 5; done
PID  9071, 42:13  until ! pgrep -f "bd dolt pull" >/dev/null 2>&1; do sleep 20; done
```

Each shell's own command line contains the literal `bd dolt pull`, so
`pgrep -f` matches the waiter itself — and both match each other. **The
condition is unsatisfiable by construction: they are not slow, they are
non-terminating.** The diagnosing `pgrep -a -f "dolt pull"` returned three
hits — the two waiters and the diagnostic itself — and zero actual pulls, no
pull having run for ~50 minutes. They held no lock and blocked nothing; the
cost is a task list full of phantom long-running work and processes that
outlive every agent that could reap them.

**Evidence.** The reproduction, with PIDs and elapsed times, recorded on
`agent-plugins-wgxd`. Verified fix in the same comment: with
`P="bd dolt"; P="${P} pu""ll"` the waiter's own command line never contains the
literal and real matches read 0, while the naive form reported 1 (itself).

**Proposed change.** Add to `AGENTS.md` beside C9's entry:

> **A `pgrep -f` predicate matches the shell that is waiting on it.** `until !
> pgrep -f "some command"; do sleep; done` never exits: the loop's own command
> line contains the pattern. Wait on a pid (`ps -p <pid>`) or on an output
> file. Where a name really is the only handle, build the pattern so the
> waiter's own command line cannot contain it.

C9 and C10 land in the same section and can be accepted as one edit; they are
listed separately because each is independently decidable.

---

### C11 — A completion query written into a work order must carry `--status all`

- **audience** skill · **target** `groom-beads` · **priority** 2
- **findings** `agent-plugins-o63s.19#2` · **recurrence** 0

**What is wrong, and what it costs.** The DoD work order pinned its own
verification query — *"the run's work is exactly `bd list --parent
agent-plugins-o63s --exclude-type=epic` — 19 beads. Verify against that query,
not against memory"* — and `bd` excludes closed issues by default. Measured on
bd 1.2.2, that exact command prints `Total: 3 issues`. **A verification query
that reads 3 on a finished run and 19 on an unfinished one is worse than no
query**, because the whole point of pinning it was to stop an agent verifying
from memory. The same defect is in `sprint-retro`'s harvest (C1) and would
silently empty a retro.

**Evidence.** `agent-plugins-o63s.19#2` — *"Measured on bd 1.2.2: that exact
command prints 'Total: 3 issues (1 open, 1 in progress)'… The same command with
`--status all` prints 'Total: 21 issues'."* Confirmed for this record: `bd list
--parent agent-plugins-o63s --exclude-type=epic --status all --json` returns 20
records (19 work orders plus the reparented `.20`).

**Proposed change.** One line in `groom-beads`' work-order guidance: **any
`bd list` written into a work order, an acceptance criterion or a completion
query SHALL carry `--status all`**, because the query will be run at the moment
the work it counts is closed. Add it to the Definition of Ready's checks if
there is a mechanical place for it.

---

### C12 — `format` is the check least likely to be selected and the most likely to fire

- **audience** skill · **target** `run-batch` · **priority** 3
- **findings** `agent-plugins-e6xt.19#3` · **recurrence** 0

**What is wrong, and what it costs.** `run-batch`'s fast loop tells a developer
to run `bb-checks.py <id>…` picking checks "by what you touched". A Go-only
bead naturally picks `test-go`, `vet` and `build` — and `format` is invisible to
all three by construction. It fired at the end of a bead in which every targeted
test had already passed and the binary had already built.

**Evidence.** `agent-plugins-e6xt.19#3`, measured in sequence:

```
cd cli && go test ./swarm/cli/ -run 'TestADryRun|TestAFailedFetch|Branch|Land'   -> 0
cd cli && go build -o /usr/local/bin/swarm ./swarm                               -> 0
python3 plugin/skills/bead-bonanza/scripts/bb-checks.py                          -> 1
  === format
  unformatted:
  swarm/cli/branch.go
```

Cause: a resolution string reflowed across two lines with an indentation gofmt
does not accept. `AGENTS.md` already documents the neighbouring trap (`gofmt -l`
exits 0 whether or not it found anything, which is why the check points at
`scripts/format-check`); this is the adjacent one.

**Proposed change.** In `run-batch`'s fast-loop guidance, name `format` as
always included: *"`bb-checks.py format <what you touched>` — `format` is not a
check about the files you edited, it is a check about the whole tree, and
nothing in a compile or a test can see it."*

## Already filed during the run — not re-proposed

Thirteen beads were filed in flight, by dispatch agents and by the coordinator.
They are the run's own follow-up queue and this document does not re-propose
them. **All thirteen exist only in this container's tracker** (headline 1),
which is why each is listed here with enough of its own evidence to be re-filed
from this document alone.

| bead | P | what it carries |
|---|---|---|
| `agent-plugins-wgxd` | 1 | **The tracker's Dolt remote diverged and cannot be reconciled from inside the container.** `bd dolt push` non-fast-forward, `bd dolt pull` aborts on `merge conflicts in issues require operator resolution`, `bd dolt --help` exposes no conflict subcommand, the dolt CLI is absent. Also carries the container-wide `bd` lock and the self-matching `pgrep` predicate (C10) |
| `agent-plugins-1vya` | 1 | `swarm dispatch accept` cannot rebase while the dispatch worktree stands — git 2.43.0 refuses (`exit 128`), and the FRD orders the merge before the worktree removal. Green in e2e because its fixtures declare no worktree |
| `agent-plugins-gqwl` | 1 | `swarm ready` reports a bead available while its dispatch's worktree stands — `AC-SWARM-DISP-002.9` unmet against the real harness. Two agents on one bead |
| `agent-plugins-asv8` | 2 | `swarm status` undercounts and mis-attributes dispatch worktrees before a dispatch's first commit; two live dispatches collapsed onto one finished one |
| `agent-plugins-8vco` | 2 | Validate every rendered `swarm` result against the generated JSON Schema in a test — the check that would have caught `e6xt.8`'s five-command stamping regression |
| `agent-plugins-3vxe` | 2 | `review_budget` counts passes in the guard and loops in the prose, so the last loop ships unreviewed |
| `agent-plugins-bhaj` | 2 | The pre-merge acceptance check (FRD ceremony step 2) is implemented nowhere and owned by nobody |
| `agent-plugins-oyvf` | 2 | `AC-SWARM-DISP-004.1` is declared e2e but cannot be built in a single-actor world |
| `agent-plugins-5w5l` | 2 | `pre-flight.md`'s harness-worktree bullet goes stale once `swarm dispatch open` owns the worktree |
| `agent-plugins-jou4` | 2 | Retract the stale `comment_count` claim in `bb-status.sh` |
| `agent-plugins-e6xt.15` | 3 | `swarm ready --agent` can now exit 2; `run-batch` and `dispatching.md` do not say so. **Deferred**, held out under row 1's carve-out |
| `agent-plugins-2ekn` | 3 | `dispatch open --dry-run` errors instead of a clean preview when the sprint branch does not exist. Pre-existing; pass 1 deduped onto it |
| `agent-plugins-o63s.20` | 2 | Decode a refused push into the run's vocabulary. **Deferred** as scope creep, then reparented out of the target epic so it could close |

## Every finding, verbatim

**113 findings across 42 beads, grouped by kind, in the order they were
recorded within each bead.** Finding ids are `<bead-id>#<ordinal>`, the ordinal
being the finding's position among that bead's own `retro:` records.

Nothing is filtered. A finding filtered out once can never be observed
appearing twice, and recurrence is the only signal separating a real defect
from one bad afternoon — this document is also, per headline 1, the only place
these survive.

### `decision` — 42

**`agent-plugins-o63s.3#2`** · dev-e6xt-b3 · 2026-09-01T17:08:52Z

```
retro:decision the delta's second behavior line fixes three of the four boundaries and leaves one open; I settled it as 'equal is not a decrease'.
evidence: the line reads 'SHALL refuse a write that decreases consecutive_dispatch_failures by more than a reset to zero'. That settles rising (allowed), a reset to 0 (allowed) and any lower non-zero value (refused). It does not say what happens when the write equals the recorded count. I implemented next >= recorded as accepted, so re-writing 3 over 3 goes through — the same shape the write-once rule already takes for the branch, where re-writing the recorded value is what lets a resumed run reuse it rather than fail on finding it. Rejected alternative: refuse an equal write as a no-op that signals a confused caller; it would make an idempotent retry after a dropped connection fail, which is the exact situation a durable count exists for. cli/swarm/manifest/validate.go, 'next >= recorded'.
```

**`agent-plugins-o63s.3#3`** · dev-e6xt-b3 · 2026-09-01T17:09:01Z

```
retro:decision the third behavior line ('report an absent count as absent rather than as zero') needed no code — I pinned it with a test instead of adding a mechanism.
evidence: Document.Get and Document.Decode already return present/absent as a separate answer from the value (cli/swarm/manifest/manifest.go: 'Present-and-absent are separate answers on purpose'), and Validate already reports a dispatch-stage key absent at the scope stage as SeverityPending rather than a failure. Writing a bespoke accessor for this one key would have been a second opinion about a behaviour the Document already has. Rejected alternative: a typed reader returning (int, bool) on the manifest surface — it would add an API nothing calls, since the caller that needs it (CB) is a later bead. The absence semantics are now pinned by TestTheDispatchFailureCountRisesOrResetsToZeroAndIsAbsentBeforeItStarts, which asserts Get reports absent, Decode returns (false, nil) leaving the target untouched, and Validate at the scope stage raises no failure for the key.
```

**`agent-plugins-o63s.1#2`** · dev-e6xt-b1 · 2026-09-01T17:40:24Z

```
retro:decision Two of the four operations CANNOT be scoped by their invocation, so the scope was moved into the decoded value instead of into the argv.

The delta requires each new operation to take the run scope as an argument and to return a plan step. bd cannot honour that for two of them, measured on 1.2.2:

    bd epic status --json | jq length                        -> 12
    bd epic status agent-plugins-e6xt --json | jq length     -> 12   (exit 0, stderr empty)

The positional argument is ignored — an invocation that LOOKS scoped is not. And for the assignee query:

    bd list --parent agent-plugins-e6xt --all --limit 0 --json | jq -r '.[].id'
    -> the run's seven DIRECT children only; no grandchild

    bd ready --parent agent-plugins-e6xt --assignee dev-e6xt-b1 --exclude-type=epic
    -> agent-plugins-o63s.1, a grandchild

So `--parent` is transitive on `bd ready` and direct-only on `bd list`, and there is no transitive assignee query at all.

DECIDED: `ClosableEpics` and `Holding` declare the widest correct invocation, and their decoders return `tracker.Closable` / `tracker.Held` — types that carry the scope asked about and whose ONLY accessor is `Within(beneath []Bead)`, taking the run's descendant set. A caller cannot read the wide answer as the run's without narrowing it, and a unit test asserts neither argv carries the scope (`TestTheTwoWideQueriesDoNotPretendToBeScoped`).

REJECTED: passing the scope to bd anyway. It costs nothing at the call site and is wrong in both directions — `bd list --parent` would silently drop every bead below the first level, and `bd epic status <run>` would read as filtered while returning the whole database. Both are exit-0 wrong answers, which is the failure this container exists to remove.

REJECTED: returning a plain `[]Bead`/`[]EpicClosure` and documenting that the caller must narrow. Prose does not stop a caller, and the caller here would close another run's epics.
```

**`agent-plugins-o63s.2#2`** · dev-e6xt-b2 · 2026-09-01T17:40:45Z

```
retro:decision A merge or rebase refusal is a plan-STOPPING failure carrying a typed *vcs.Refusal, not an exit status declared through Step.Answers.
evidence: The bead carries two behavior lines that pull in opposite directions — "WHEN a merge or a rebase is refused, Git Version Control SHALL decode the refusal into the run's vocabulary" and "A step whose answer is carried by its exit status SHALL declare the statuses it answers with, so a legitimate 'no' is not read as a failure". A refused fast-forward reads like a legitimate "no", so declaring Answers=[128] on it was the obvious reading.

Rejected, for two reasons that are both mechanical rather than stylistic:
  - plan.Record.Failed() is `r.Err != nil || !r.Answered`. Declaring 128 an answer does not make the step succeed while its decoder returns an error, so the declaration would buy nothing and would misdescribe the step.
  - The existing Merge already documents the ordering guard it depends on: "a worktree removal chained after a merge that exited non-zero is the incident this design exists to prevent". A refused fast-forward that let the plan continue would reintroduce exactly that.

So: the refusal stops the plan, and the CALLER branches on it by reading the report's record and doing errors.As(record.Err, &refusal). Proved end to end by TestARefusalReachesTheCallerThroughTheReport in cli/swarm/vcs/vcs_test.go — the *Refusal survives the trip through the Plan Runner rather than being flattened to prose.

Cost of the alternative had I taken it: commands built on this in stage 2 (MG, RM) would have had steps that continue after a fold that did not happen.

Second decision, settled silently because the blueprint states the no-force rule only for worktree removal: the same discipline is applied to branch deletion and push. DeleteBranch uses `git branch -d`, never -D — a dispatch branch that is not merged is one whose work has not landed. Push carries no --force — measured, a refused push exits 1 with " ! [rejected]        main -> main (non-fast-forward)", which is a condition to report rather than to overwrite. Neither is required by an acceptance criterion; both are enforced by the package-wide literal scan in TestTheWorktreeRemovalHasNoForcingForm, which forbids "--force", "-f" and "-D" outright.
```

**`agent-plugins-o63s.1#3`** · dev-e6xt-b1 · 2026-09-01T17:40:48Z

```
retro:decision The drain derivation (`tracker.AgentQueue`) was built HERE, in stage 1, rather than left to the AQ delta.

The blueprint's TRK behavior line asks the interface for 'an operation returning the beads one assignee holds beneath that same scope, PARTITIONED by whether each is ready or in progress'. No single bd invocation answers both halves — `bd ready` excludes in_progress by construction (its own help: 'Excludes in_progress, blocked, deferred, and hooked issues'), so the partition is necessarily two queries. A type alone (`tracker.Queue`) would satisfy the noun but not the word 'operation'.

DECIDED: `AgentQueue(ctx, t, runner, scope, agent, filter)` lives in `cli/swarm/tracker/queue.go`, composing `Ready` + `Holding` + the existing `Descendants` walk, and returning a `Queue` whose `Drained()` is true only when BOTH halves are empty. `tracker.Descendants` is the precedent: a derivation over steps, in this package, taking a `plan.Runner` and reaching no Process Runner.

REJECTED: leaving it to AQ. The AQ delta says 'The drain derivation SHALL be one implementation, reused by `swarm dispatch accept`'s drain check rather than restated in it' — two stage-2 deltas (AQ and MG) each need it, and a stage-1 home is the only one that is not one of them. The cost of the rejected option is the drain verdict being derived twice, which is precisely the duplication that line forbids.

Cost accepted: TRK now ships slightly more than the four operations its behavior lines enumerate. A reviewer who thinks the derivation belongs in stage 2 can move the file without touching the interface.

Note for whoever builds AQ: the two halves are scoped DIFFERENTLY and that is deliberate. `Ready` is scoped by bd's own transitive `--parent`; `Holding` cannot be, so it is narrowed against the walk. See `AgentQueue`'s doc comment.
```

**`agent-plugins-o63s.4#1`** · dev-e6xt-b3 · 2026-09-01T17:47:03Z

```
retro:decision a git bundle cannot carry a worktree registration or a remote, so two of the five required variants had no fixture form until I extended the harness.
evidence: the delta requires the variants to cover 'a standing dispatch worktree with its branch' and 'a remote that refuses a push', and Change Guidance fixes a fixture as one directory holding repository.bundle, tracker.jsonl and fixture.toml. A bundle carries refs and objects; a worktree is a registration under .git and a remote with a pre-receive hook is neither. The existing precedent (cli/swarm/e2e/worktree_attribution_test.go:29, 'git worktree add' inside the test) puts that half of the world in the TEST, which would mean two cases reading the same world had to agree by hand about what it contains. I added declarations to fixture.toml instead — [[worktree]] with branch and path, and [refusing_remote] with name, refuses and message — and cli/swarm/e2e/harness_test.go's restore now stands both up. Rejected alternative: leave them to each case, which is cheaper now and makes the world un-reviewable as data, contradicting 'adding either is data a reviewer reads rather than Go a reviewer runs'.
```

**`agent-plugins-o63s.4#4`** · dev-e6xt-b3 · 2026-09-01T17:47:31Z

```
retro:decision two of the five variants copy the core's bundle byte for byte rather than branching from it, because they need no different history.
evidence: Change Guidance says 'WHEN a new world shape is needed for a case, it SHALL be derived as a branch from the core fixture rather than authored separately'. phase-epics and push-refused need only a different tracker export and a different remote respectively, so branching would add a commit that means nothing. I followed the precedent already in the tree: cli/testdata/swarm-fixtures/status-attention/fixture.toml says 'repository.bundle copied from the core fixture', and 'cmp status-attention/repository.bundle core/repository.bundle' reports the files identical. Each new fixture.toml states which of the two it did. Rejected alternative: give every variant a distinguishing commit for uniformity, which would have made three bundles differ from the core for no reason a reader could act on.
```

**`agent-plugins-o63s.5#2`** · dev-e6xt-b4 · 2026-09-01T18:17:33Z

```
retro:decision the closed_epics fact reports the SAME id list in dry-run and execute mode — derived from the eligibility READ (bd epic status --eligible-only), never from the write's own record. Rejected alternative: mirror manifest set's dry-run convention (report the pre-write stored value, would-be state only in the summary prose) — that fits a manifest KEY that has an independently-readable stored value, but closed_epics has no such value to fall back to, and the FRD states 'a criterion below saying a command reports something is saying that thing is a property of the answer, not prose the caller has to parse,' so AC-SWARM-DISP-001.3's 'reports the epics it would close' has to be a real fact even under refusal. Cost of the rejected alternative: AC-SWARM-DISP-001.3 would have no fact to point at, only a summary line.
```

**`agent-plugins-o63s.5#3`** · dev-e6xt-b4 · 2026-09-01T18:17:39Z

```
retro:decision AC-SWARM-DISP-001.6 asks for 'the open child's id' as the reason an ineligible epic stays open, but bd's own eligibility answer (bd epic status) carries only total_children/closed_children counts, not which child is open. Decision: query each ineligible epic's own children (tracker.Children) and name the lowest-id one not in tracker.StateDone as the reason — deterministic, but arbitrary when more than one child is open (the FRD names none as authoritative). Rejected alternative: name every open child rather than one — rejected because the AC's wording ('the open child's id', singular) and the fixture's own scenario (exactly one open child) both point at one id, and a list would still need a rule for which to lead with.
```

**`agent-plugins-o63s.6#1`** · dev-e6xt-b4 · 2026-09-01T18:27:47Z

```
retro:decision reused status.Collect wholesale for worktree-to-dispatch attribution (held vs available) rather than extracting a smaller shared helper. Alternative rejected: duplicate measureWorktreeOwners's branch-then-HEAD-commit matching rule inside cli/swarm/cli — rejected because it is unexported in package status and this delta's own c4_impacts names Status Collector, and duplicating a matching rule this precise (branch name, falling back to the assignee of the bead a worktree's HEAD commit closed) is exactly the drift risk AGENTS.md's 'one implementation, reused' principle warns about. Cost: swarm ready now pays a full completion walk (tracker.Descendants) it does not otherwise need, just to learn which worktree each dispatch holds.
```

**`agent-plugins-o63s.9#1`** · dev-e6xt-b4 · 2026-09-01T18:39:00Z

```
retro:decision the sprint branch's own default base (when --base is absent, on the dispatch that CREATES it) is not stated anywhere in the FRD or blueprint — every stated default (AC-SWARM-DISP-003.10, 'the sprint branch's head as it stands now') is about the WORKTREE's base, which is WT's delta (agent-plugins-o63s.10), not BR's. Decision: default to HEAD of the coordinator's current checkout, the same default 'git branch <name>' itself uses with no start-point. Rejected alternative: require --base on the creating call (make it mandatory rather than defaulted) — rejected because the full command signature declares --base optional with no asterisk singling out the first dispatch, and a coordinator's first dispatch call has nothing else to name.
```

**`agent-plugins-o63s.9#2`** · dev-e6xt-b4 · 2026-09-01T18:39:08Z

```
retro:decision --dispatch is declared as a flag on 'swarm dispatch open' (parses without error) but this delta's own logic never reads it — the sprint-branch half BR owns needs only --run, --base and --branch. Kept in the flag set now rather than added when WT lands, because the FRD states ONE command signature (swarm dispatch open --run <id> --dispatch <dispatch-id> [--base <ref>] [--branch <name>]) for the whole of FEAT-SWARM-DISP-003, and a caller scripting against that signature today should not have an unknown-flag error where WT will need one. Not enforced as required here — that gate is WT's, once it has a use for the value.
```

**`agent-plugins-o63s.7#1`** · dev-e6xt-b5 · 2026-09-01T19:04:08Z

```
retro:decision AQ's establishment check (AC-SWARM-DISP-002.15) needed the run's WHOLE assignment set, but tracker.AgentQueue deliberately does not expose the 'beneath' descendant set it computes internally — its own doc comment says telling drained apart from never-established 'is the caller's' and 'needs the run's whole assignment set'. Rather than restate AgentQueue's walk-plus-join derivation to get at beneath, I call tracker.Descendants(scope) a second time in ready.go before calling AgentQueue, which walks the run's tree twice per --agent invocation. Reused the shared Descendants primitive rather than restating the drain derivation, per the bead's own instruction; recording the extra query cost as a decision rather than silently.
```

**`agent-plugins-o63s.7#2`** · dev-e6xt-b5 · 2026-09-01T19:04:09Z

```
retro:decision The flag name --agent (already implied by the delta's own command line, swarm ready --run <id> [--agent <dispatch-id>], in specs/features/swarm-dispatch/swarm-dispatch-frd.md) and the fact names established/drained/queue/queue_count on the JSON answer are not pinned anywhere else in the corpus. Chose established/drained as bare Flag facts and queue as a Structured list of {id,state}, state being "ready" or "in_progress" — no existing convention to match against.
```

**`agent-plugins-o63s.10#2`** · dev-e6xt-b6 · 2026-09-01T19:14:19Z

```
retro:decision No VCS method existed to read the repository's working-tree root, and cli/swarm/cli must never name "git" itself (cli/swarm/vcs/git/git_test.go's TestOnlyThisPackageNamesTheTool scans every non-test .go file under cli/swarm for the literal string "git" and fails any outside vcs/git that carries it). Added vcs.VersionControl.Root(prefix) (cli/swarm/vcs/vcs.go), implemented in git.go via `git rev-parse --show-toplevel` (decodeRoot in decode.go), so the coordinator's worktree path (a sibling of the repo root, per the pinned <parent>/.bb-worktrees/<dispatch-id> form) is computed from a real read rather than assuming the process's cwd is the root. Verified this bead's anchors cover it: the bead's own Design section names cli/swarm/vcs and cli/swarm/vcs/git as anchors alongside cli/swarm/cli, not cli/swarm/cli alone.
```

**`agent-plugins-o63s.10#4`** · dev-e6xt-b6 · 2026-09-01T19:14:39Z

```
retro:decision Two decisions the ACs left to implementation: (1) an existing worktree "for the dispatch" (AC-SWARM-DISP-003.11) is matched by branch name (bb/<dispatch-id>) against vcs.Worktree.Branch, not by path — the FRD's own Intent paragraph for this delta says naming the branch for the dispatch id is what lets `git worktree list` say which dispatch a tree belongs to, so branch is the identifying field by the delta's own reasoning. (2) The single --base flag serves double duty across the two halves this command now performs: for the sprint branch (BR's half, unchanged) it is the ref the branch is cut from when the branch doesn't exist yet, defaulting to HEAD; for the worktree (this bead's half) it is the ref the worktree is cut from, defaulting to the sprint branch's own current head (resolved fresh via vc.Resolve, never the recorded base_sha, so AC-SWARM-DISP-003.10 "read at that moment" holds even on a resumed run whose sprint branch has since moved). No second --base-like flag was added; the FRD's command signature ("swarm dispatch open --run <id> --dispatch <dispatch-id> [--base <ref>] [--branch <name>]") names only one.
```

**`agent-plugins-o63s.8#1`** · dev-e6xt-b5 · 2026-09-01T19:16:01Z

```
retro:decision The FRD's Command: line for FEAT-SWARM-DISP-002 (swarm ready --run <id> [--agent <dispatch-id>]) names no flag for reporting a dispatch's return, yet the behavior lines require the command to both read AND write consecutive_dispatch_failures from what a dispatch return completed. Concluded a caller-supplied signal is the only way (the CLI cannot observe 'a dispatch just returned' on its own) and that the Command: line is non-exhaustive — precedent: the already-merged QF delta's --explain/--dry-run/--json/--full-result on this same command are absent from that line too. Added --completed <n> (Cobra Changed()-detected, so --completed 0 differs from omitting it) and --threshold <n> (default 3, per the bead's own decision). Names are unpinned per the blueprint's own note.
```

**`agent-plugins-o63s.8#2`** · dev-e6xt-b5 · 2026-09-01T19:16:02Z

```
retro:decision Applied the breaker only in bare (--run, no --agent) mode. The CB delta's own heading is 'swarm ready ... (added — the circuit breaker)', not '--agent', and every AC's Given/When is phrased against 'the ready set is queried' rather than an agent's own queue — a dev agent looping on --agent for its own queue has no reason to carry breaker semantics. --threshold/--completed are silently inert when --agent is also passed rather than refused, since the FRD does not ask for a refusal and refusing would be scope beyond the delta.
```

**`agent-plugins-o63s.11#2`** · dev-e6xt-b4 · 2026-09-01T19:48:26Z

```
retro:decision the AC asks for the case to be decided by "comparing the dispatch branch merge-base with the target against the target current head". Implemented as the existing `VersionControl.Descends(target, dispatchBranch)` rather than a new merge-base-returning operation: `git merge-base --is-ancestor A B` is true exactly when merge-base(A,B) == A, so asking whether the target is an ancestor of the dispatch branch IS the equality the AC states, and it runs the same plumbing command. Rejected adding a `MergeBase` operation that returns a sha for the caller to compare: it would add a second way to ask one question and a second sha for a caller to mis-compare, for no extra information.
evidence: cli/swarm/vcs/git/git.go, `func (v VersionControl) Descends` -> `"merge-base", "--is-ancestor", ancestor, descendant`, with `s.Answers = []int{1}` so a legitimate "no" is an answer rather than a failure.
```

**`agent-plugins-o63s.11#5`** · dev-e6xt-b4 · 2026-09-01T19:48:56Z

```
retro:decision the AC requires the sequencing finding to name "both beads" without saying how the second is derived. Decided: the unmerged bead is the bead the dispatch branch head commit closed, and the bead it conflicts with is the bead the TARGET head commit closed — both read through `status.BeadIDFromCloseSubject`, which was unexported and is now the run one reading of a close commit subject. Rejected deriving the second bead from which commit introduced each conflicting hunk (git blame over the conflict): it is several more invocations, it has no single answer when a conflict spans commits, and the fixture the AC is proved against pins exactly the pairing chosen here. Where the branch head is not a close commit, the finding is not written on a guessed bead: `unmerged_bead` is reported and `finding_recorded` comes back unestablished.
evidence: `cli/testdata/swarm-fixtures/dispatch-rebase-conflict/fixture.toml` — "two closed beads the finding names — fixture-run.3, the unmerged dispatch own, and fixture-run.4, whose commit is the one on the target it conflicts with"; `git log --oneline --all` on that bundle shows 078f2b6 "fixture-run.4 P1 ..." at the target head and 99ce1dd "fixture-run.3 P1 ..." on bb/dev-fixture-2.
```

**`agent-plugins-o63s.11#6`** · dev-e6xt-b4 · 2026-09-01T19:49:04Z

```
retro:decision the delta says nothing about which branch the repository must be on, but `git merge --ff-only <branch>` advances whatever is CHECKED OUT. Without a guard the command could report the target merged while having advanced something else. Added a read of HEAD beside the case comparison and an attention refusal naming `git checkout <target>` when it is not at the target sha. Rejected leaving it out as unspecified: the whole point of naming the case before the refusal is that a fold lands where it is supposed to, and this is the cheapest thing that makes that true.
evidence: cli/swarm/cli/accept.go, `readCase`, the branch guarded by `if headSHA != targetSHA`.
```

**`agent-plugins-o63s.12#1`** · dev-e6xt-b4 · 2026-09-01T20:02:20Z

```
retro:decision the delta requires `--remote` to be honoured, but the landed `VersionControl.Divergence(prefix, branch)` compared against `<branch>@{upstream}`, which ignores the named remote and fails outright on a branch with no upstream configured even where the counterpart exists. Changed the signature to `Divergence(prefix, branch, remote)`, comparing `<branch>...<remote>/<branch>`. Rejected adding a second operation beside it: two operations answering one question is what the container is built to avoid, and Divergence had no production caller to break — only its own argv tests.
evidence: measured, git 2.43.0, scratch repo: `git rev-list --count --left-right main...origin/main` -> exit 0, "0\t0"; the same after one local commit -> "1\t0". `git rev-list --count --left-right claude/sprint...origin/claude/sprint` with no such remote-tracking ref -> exit 128, "fatal: ambiguous argument". `grep -rn "\.Divergence(" cli/ | grep -v _test.go` returned only the interface declaration and the git implementation.
```

**`agent-plugins-o63s.12#3`** · dev-e6xt-b4 · 2026-09-01T20:02:39Z

```
retro:decision "invocable with no run scope" left open whether `--run` should exist on these commands at all. Kept the flag declared and made only the refusal conditional, because the surface has a test asserting that EVERY invocable command in the tree carries the five contract flags, `--run` among them — dropping it would have failed that test, and the delta asks for an exception to the refusal rather than to the flag set. The scope is carried into the result when given; the contract already anticipates an empty one.
evidence: cli/swarm/cli/surface_test.go, `TestEverySubcommandInTheTreeCarriesTheContractFlags`, which walks the tree and checks `cmd.Flags().Lookup(name)` for "explain", "dry-run", "json", "run", "full-result". schemas/swarm/output.strata on `run`: "The run epic this invocation was scoped to. Empty for a command that is not run-scoped."
```

**`agent-plugins-o63s.12#4`** · dev-e6xt-b4 · 2026-09-01T20:02:48Z

```
retro:decision the delta says a branch with no counterpart takes the attention status, but the only measurement available conflates that with a broken comparison: `git rev-list --count --left-right A...origin/A` exits 128 for a missing counterpart AND for a malformed revision. Split it into two steps — a `Resolve` probe of `<remote>/<branch>` whose failure IS the no-counterpart answer, followed by the divergence count — so "no counterpart" (attention, exit 1) is never reported as "the comparison failed" (stop, exit 2). A caller routing on the status would otherwise halt a run over a branch that simply has not been pushed yet.
evidence: measured, git 2.43.0: `git rev-parse --verify origin/claude/sprint^{commit}` with no such ref -> exit 128, "fatal: Needed a single revision"; `git rev-list --count --left-right claude/sprint...origin/claude/sprint` -> exit 128, "fatal: ambiguous argument". Same status, two different conditions.
```

**`agent-plugins-o63s.13#2`** · dev-e6xt-b7 · 2026-09-01T20:41:30Z

```
retro:decision the pre-merge acceptance check (ceremony step 2) was filed, not implemented — filed as agent-plugins-bhaj
evidence: The FRD ceremony (swarm-dispatch-frd.md:679) has 8 steps; step 2 is "pre-merge acceptance the acceptance signals per bead, deferring the two a merge decides". Neither delta covers it. MG's behavior lines (blueprint "drain, case and merge") name the drain check, the case and the merge only; RM's name "After the merge and the removal, the command SHALL re-check every bead in the dispatch with no signal deferred". Reading the code: `acceptDispatch` in cli/swarm/cli/accept.go goes drain -> target -> case -> fold with nothing between, so MG shipped without it.
I implemented step 7 (recheckDispatchBeads) and left step 7's deferring twin unbuilt, because RM's acceptance criteria are satisfied without it and building it would add a fifth stop condition to MG's arm that no AC of this bead covers.
Rejected: wiring the same checker in ahead of readCase with the two signals deferred. It is ~10 lines given what this bead built, and the FRD calls for it ("Four things stop an acceptance ... a bead failing an acceptance signal at exit 2"). The cost of NOT doing it: an acceptance whose bead is unclosed or carries no finding now merges first and discovers it at step 7 — the work still lands and the run still stops before the push, so nothing is lost, but the FRD's second stop condition does not exist and the merge happens on a dispatch that should not have reached it.
```

**`agent-plugins-o63s.13#3`** · dev-e6xt-b7 · 2026-09-01T20:41:42Z

```
retro:decision --no-push exits CLEAR with landed=false, rather than attention
evidence: The behavior line is "WHERE --no-push is passed, the command SHALL complete every step before the push and report the branch as unlanded rather than pushing." It fixes the fact and not the status.
`swarm branch verify-landed` takes the attention status for an unlanded branch (AC-SWARM-DISP-005.1, .2 — see landingVerdict in cli/swarm/cli/branch.go, both arms return OutcomeAttention), so the consistent-looking choice was attention here too. Rejected: a coordinator routes on the exit status, and --no-push is a deliberate policy rather than a condition; exiting 1 would make every accepted dispatch in a no-push run read as needing attention, which is the signal the status exists to carry.
A second constraint settled the shape of the fact. `answer` in cli/swarm/cli/answer.go renders ONLY satisfied facts at a clear exit (`satisfied(f) = f.Established && f.Reason == ""`), so a `landed=false` carrying a Reason is silently dropped from the answer document — measured: the first version of this carried Reason "--no-push was passed, so nothing was pushed" and the test read `landed = <nil>` off the JSON. It is now a plain Flag with a Resolution and no Reason, which renders as `"landed": false`.
The fact's invocation is `swarm branch verify-landed --branch <target>` and not `swarm branch push`: an invocation on a fact is what REPRODUCES it, and the push would change it.
```

**`agent-plugins-o63s.16#4`** · dev-e6xt-b9 · 2026-09-01T21:05:01Z

```
retro:decision the FRD names 'the --exclude-label human paragraph in references/run-record.md' as replaced, but only its LAST SENTENCE is what swarm ready enforces. Deleted that sentence and named the command; kept the rest of the paragraph, which is bd behaviour no command replaces (bd create --add-label human; bd human respond/dismiss both fail with 'storage is nil'; bd human list keeps listing a closed bead, so the open-question query must be bd list --label human --status open). Rejected alternative: delete the whole paragraph, which would have dropped the only record of those three bd facts.
evidence: `swarm ready --help` — 'with the run's own filter applied — no epic, no bead closed or deferred, none awaiting a human answer'; cli/swarm/cli/ready.go:35 const humanLabel = "human", 'Excluded here and nowhere else'.
```

**`agent-plugins-o63s.16#6`** · dev-e6xt-b9 · 2026-09-01T21:07:08Z

```
retro:decision swept one file the AC's deletion list does not name, because removing the pre-flight block left it instructing a coordinator to run a check that no longer exists: dispatch-policy.md's concurrency entry conditions read '...and the base-ref check re-verified. Both are yours at dispatch'. Rewrote to one condition (no shared c4_changes) and said the base is an argument swarm dispatch open is given; also moved its re-query from 'bd ready --parent <epic> --exclude-type=epic' to 'swarm ready --run <epic>'. Rejected alternative: leave it, per the item's scope discipline — that would ship a dangling instruction.
evidence: plugin/skills/bead-bonanza/references/dispatch-policy.md, 'Entry conditions, both required and both about **safety**' — the second was the block this item deletes from pre-flight-check.sh. After the sweep, 'grep -rn "base-ref|baseRef" plugin/' returns nothing.
```

**`agent-plugins-o63s.15#2`** · dev-e6xt-b8 · 2026-09-01T21:07:44Z

```
retro:decision Change Guidance needed a judgment call on which stage-2 implementation facts count as 'lockstep rules' vs narrative belonging in a component's own prose. I added four SHALL/SHALL-NOT bullets: (1) a command whose plan is not knowable before an earlier read answers bypasses Spec/newSubcommand — generalizing swarm status's own reasoning to phase close, dispatch open and dispatch accept; (2) a command needing a dispatch's drain verdict SHALL reuse tracker.AgentQueue rather than re-deriving it (accept.go's own comment states this as the reason it does not re-derive the two queries); (3) a commit subject read for a bead id SHALL go through status.BeadIDFromCloseSubject, now shared by the Status Collector and dispatch accept; (4) a command SHALL refuse an absent run scope unless it is itself the run-recovery path (branch verify-landed/push). I left out narrower per-command facts (e.g. 'push and verify-landed share one implementation') as component prose rather than Change Guidance, since they describe one command pair rather than a rule future authors extending the surface need to follow.
```

**`agent-plugins-o63s.17#2`** · dev-e6xt-b9 · 2026-09-01T21:12:31Z

```
retro:decision sprint-recovery's step 1 loses one signal in the swap and I put it back by hand rather than dropping it. 'bb-status.sh branch' reported three things — tree cleanliness, landing, worktree count — and 'swarm branch verify-landed' reports only landing. The worktree count is already step 4's ('swarm status --run <epic>' prunes and counts), so only tree cleanliness was orphaned; step 1 now names 'git status --porcelain' beside the command. Rejected alternative: delete it, which would have left recovery with no read of the working tree until step 5's per-worktree table.
evidence: plugin/skills/bead-bonanza/scripts/bb-status.sh:208-238 — 'git status --porcelain' -> 'tree: clean', then the landing comparison, then 'git worktree prune && git worktree list'. cli/swarm/cli/branch.go emits only Flag("landed") and Count("unpushed") (lines 345, 385-386, 402-403, 419); no porcelain read anywhere in the file.
```

**`agent-plugins-o63s.18#2`** · dev-e6xt-b10 · 2026-09-01T21:24:34Z

```
retro:decision 'The Coordinator's roster line' is ambiguous in specs/architecture/module-bead-bonanza.md — the document holds two rosters. I put the command names on the 'Boundaries at a glance' table row.
evidence: Two candidates. (a) the C4 mermaid roster: 'Component(coord, "Coordinator", "skill", "Routes every phase; owns the manifest, the merge and the loops")'; (b) the 'Boundaries at a glance' table row: '| **Coordinator** | what happens next, and when is there nothing left? | anything requiring a source file to be read |'. Chose (b): the dispatch prompt pinned 'around line 198' and the table row is at line 198 (the mermaid coord line is 160), and the blueprint's own prose calls that table 'the point of this blueprint'. Rejected (a) because the mermaid Component() technology/description fields are one-clause role labels and four command names there would not render legibly; rejected doing both as duplicate maintenance of the same fact.
```

**`agent-plugins-o63s.19#4`** · dev-e6xt-b10 · 2026-09-01T21:38:10Z

```
retro:decision Filed the retro's RFC under rfcs/inbox/feature-blueprint/ although its findings span three document types, because retro-spec forbids splitting one retro's proposals across RFCs.
evidence: retro-spec's workflow step 5 says to file under 'rfcs/inbox/<spec-type>/' where spec-type is 'the document type the findings most concern', and the rfc-template's Scope says 'An RFC bundles ALL qualifying proposals from one retro for one variant... do NOT split proposals into separate RFC files.' The nine findings root-cause to the feature blueprint (P2, P3, P4), the FRD (P1, P5, P7) and the container blueprint (P6) roughly evenly. Chose feature-blueprint because the structurally largest proposals — binding a delta's Intent, Behavior lines, ACs and the ceremony to each other, and declaring a verification tier only where the world is constructible — are about how the blueprint decomposes deltas. Rejected filing under frd/: it would put the delta-decomposition proposals in a directory whose purpose is FRD recurrence. The directory convention and the one-RFC-per-retro rule are in tension whenever a retro spans document types, and nothing resolves it. Filed as rfc-0024 (next free number; 0001-0023 taken).
```

**`agent-plugins-e6xt.8#2`** · dev-e6xt-b11 · 2026-09-01T22:37:49Z

```
retro:decision the stamping helper was applied to EVERY result construction in the package, not only the five real ones the Design names, and the guard test enumerates the live command tree rather than a hand-written list of five
evidence: The Design asks for "ONE helper over five copies ... applied at each real construction site, so a sixth hand-declared command cannot silently repeat the omission". Applied only at the five real sites, the helper does not deliver that last clause: a sixth hand-declared command still omits the call. Two things were done about it.
(1) `Request.stamp` (cli/swarm/cli/command.go) is now the single place the three fields are set, and it is called from the refusal path, the explain paths of all six hand-declared commands, `reportStatus`, and `newSubcommand` (which previously assigned `result.Command`/`Run`/`Mode` inline). One definition of "stamped" rather than three spellings of it. No behaviour change on those paths — asserted by the pre-existing explain and refusal tests, all of which still pass.
(2) `TestEveryCommandInTheTreeHasARealResultCase` in cli/swarm/cli/stamp_test.go builds the real tree with `register(root, streams, fake)` — the one place root.go says a delta adding a command touches — walks its leaves, and fails any leaf with no case in `stampCases()`. All 11 leaves are covered today (the 5 new ones, `swarm status`, and the 4 `swarm manifest` subcommands), so a twelfth command fails the check until its own real result is asserted.
REJECTED: stamping inside `finish`, which is the single funnel every result already passes through and would make the omission structurally impossible. It was rejected because the Design explicitly scopes the change to the construction sites and forbids re-routing these commands through the surface; moving the assignment into `finish` is that same re-routing by another name, and it would also make the guard test above vacuous — nothing could fail it, including a command whose result was built wrong for some other reason. Cost accepted: a thirteenth command could still be written unstamped, and is caught by a test rather than by the type system.
```

**`agent-plugins-e6xt.7#2`** · dev-e6xt-b11 · 2026-09-01T22:45:32Z

```
retro:decision a dry run no longer reports the run STOPPED, and applyBreaker now returns its plan records; neither is spelled out in the bead
evidence: (1) STOP VERDICT. Before, `stopped` was derived from `next >= threshold` where `next` is the post-write value. Under a refused write nothing wrote that value, so reporting the run stopped from it is the same untruth as the write: `swarm ready --dry-run --completed 0` on a run one short of the threshold would have exited 2 (stop) and told a coordinator to stop a run the tracker still records as running. A refused write therefore returns Stopped=false, and the projection is carried in the fact reason instead — measured above: "the write was refused: dry run: the step writes. The count stands at 0 and would have become 1 (threshold 3)". Nothing is lost: a caller previewing the breaker reads the same two numbers, in a field that cannot be mistaken for a measurement.
REJECTED: keeping the stop verdict and only marking the fact unestablished. It would leave the result self-contradictory — an unestablished count beside an outcome derived from it — and the outcome is what skills route on without reading the output.
(2) RECORDS. The bead AC asks that "the advance-failure-count step is recorded as refused with a dry-run reason". applyBreaker built records and returned none of them, so the refusal was invisible to a --full-result consumer even once the mode was threaded — reportReady appended only the ready query and status.Collect. applyBreaker now returns a `breakerOutcome` carrying Facts, Records, Stopped and Summary, and reportReady appends the records. Chose a struct over a fifth positional return because a fifth positional return is exactly how a caller drops the records without noticing; the six existing breaker tests unpack it through one test helper (breakerPass) rather than each restating the shape.
(3) The existing breaker tests all passed `Request{Run: "run-1"}`, whose zero-value Mode is ModeExplain. They only passed because the mode was hard-coded; once threaded, explain runs nothing and the read has no answer. Each now passes `Mode: plan.ModeExecute` explicitly. That the whole existing suite depended on the constant it was testing around is the clearest evidence the defect had no test.
```

**`agent-plugins-e6xt.9#2`** · dev-e6xt-b11 · 2026-09-01T22:49:12Z

```
retro:decision the fetch guard runs BEFORE the counterpart check, and it also fires on a fetch that ran and FAILED — two points the bead Design left open
evidence: (1) ORDER. The Design says only to swap which record the existing branch consults. That branch sat AFTER the counterpart-missing branch, which returns an established `landed: false` ("has no counterpart on <remote>"). Left there, an unfetched dry run in a repository with no local tracking ref would still answer with an established false — the same defect in the other direction, since a missing LOCAL ref proves nothing about the remote when the remote was never consulted. The guard is therefore the first thing landingVerdict tests, before either read answer is read at all. REJECTED: leaving the order alone and accepting the established false; it is a wrong answer of exactly the kind the bead is about ("Do not report landed: true or landed: false from an unfetched comparison" is the Design own wording, and it applies to both branches).
(2) A FAILED FETCH. The guard is `!ok || fetchRec.Status != plan.StepRan || fetchRec.Failed()`. plan.Record.Failed() returns false unless Status == StepRan (cli/swarm/plan/plan.go:324-329), so the third clause is not redundant: it covers a fetch that really ran and failed — no network, a dead remote — after which the two reads still compare against the same stale ref and would report landed. The command help says the fetch is what makes a stale record unable to report a branch as landed; a fetch that failed did not do that. This is behaviour change in EXECUTE mode, which the Design says to leave alone, and it is deliberate: "leave execute mode alone" is about the ordinary verdict path, which is untouched — a successful fetch reaches exactly the code it reached before.
NOT COVERED BY A TEST: the failed-fetch case has no test of its own; the tests cover the refused (dry-run) case in both verbs plus the execute negative control. A test would need a fake whose `git fetch origin` recording exits non-zero while the two reads still answer.
```

**`agent-plugins-e6xt.9#3`** · dev-e6xt-b11 · 2026-09-01T22:49:39Z

```
retro:decision (addendum) the failed-fetch case named as untested in the previous finding is now covered by TestAFailedFetchDoesNotReportLandedEither in cli/swarm/cli/branch_test.go — a fake whose `git fetch origin` exits 128 while both reads still answer as an in-step branch; it asserts the attention status and an unestablished landed. Retracting the "NOT COVERED BY A TEST" line above.
```

**`agent-plugins-e6xt.11#2`** · dev-e6xt-b11 · 2026-09-01T22:54:57Z

```
retro:decision the --agent arm now also takes the STOP outcome when the breaker is at threshold, and it reads the manifest on every invocation — two consequences of "apply the breaker in both arms" the bead did not spell out
evidence: (1) STOP OVERRIDE. reportReady replaces the outcome and the summary when the breaker stops the run ("The breaker overrides whatever the dispatch set itself concluded"). reportAgentQueue now does the same. This is a behaviour change for an invocation that passes NO --completed: applyBreaker reports Stopped from the RECORDED count whether or not a return was reported, so `swarm ready --run E --agent A` on a run already at threshold now exits 2 (stop) where it exited 0. Deliberate: the caller reading one dispatch queue is precisely the one about to send more work into a stopped run, and the arm own facts (queue, queue_count, drained, established) are all still reported beside the count, so nothing that reads facts loses anything. REJECTED: applying the breaker but suppressing the stop on this arm. It would make "the same query with one more filter" false in the one way that matters — two arms disagreeing about whether the run is stopped — and it would leave the coordinator that reported a return through this arm with exit 0 on the invocation that stopped the run.
KNOWN CONSUMER, unverified against a live stopped run: the run-batch skill tells a developer subagent to route on this command exit status and reads its drain verdict from it. A worker whose run is stopped now sees exit 2 rather than exit 0 with `drained: false`. I could not test that end to end — no run in this container is at threshold (consecutive_dispatch_failures is not set on agent-plugins-e6xt at all, measured: `swarm manifest get --run agent-plugins-e6xt --key consecutive_dispatch_failures --json` reports "is not set"). What would settle it: a run whose manifest carries consecutive_dispatch_failures at or above --threshold, then `swarm ready --run <it> --agent <id>` with no --completed, checking that the skill drain reading still behaves.
(2) AN EXTRA READ. Every `--agent` query now makes one more `bd show <run> --json`. Accepted as the cost of the blueprint own "the breaker unconditionally on the command"; it is one read against a tracker the same invocation already queries three or more times.
```

**`agent-plugins-e6xt.10#2`** · dev-e6xt-b11 · 2026-09-01T23:02:53Z

```
retro:decision the landed sha is read by a STEP of the fold plan rather than by a second plan, and the e2e assertion is ancestry rather than equality with the target head
evidence: (1) ONE PLAN. `vc.Resolve("merged-head", target)` is appended to the same plan as the rebase, the checkout and the fast-forward, so the read is ordered behind the fold by the plan runner rather than by call order in Go, and it appears in the result own invocation list like every other step. A second plan run after foldIn would have worked too and is what the "read the target after the fast-forward" wording suggests; it was rejected because the ordering would then be a property of this function control flow and invisible to --explain, and because the value would cite an invocation that is not part of the plan the result reports.
Under a dry run the resolve still runs — it is a Read — but the refused-fold arm reports no merged_sha at all, so nothing consumes it.
(2) ANCESTRY, NOT EQUALITY. The e2e test asserts `git merge-base --is-ancestor <merged_sha> <target>` exits 0 rather than `merged_sha == git rev-parse <target>`. Equality is what the code produces today and would be the tighter assertion, but ancestry is the property a caller uses — "the commit is on the target" — and it survives anything landing on the target between the fold and the check. It is also the bead own acceptance criterion, worded that way.
(3) The pre-rebase branch sha is now unreported. `mergeCase.branchSHA` is still read (readCase needs it for nothing else, but m.branchInv still cites the branch read for the `dispatch_branch` fact). Nothing pinned the old meaning: `grep -rn "merged_sha" cli/ plugin/ specs/` returns one hit, the line that produces it — re-run after the change and it is still one hit, now the corrected line.
```

**`agent-plugins-e6xt.12#2`** · dev-e6xt-b11 · 2026-09-01T23:08:49Z

```
retro:decision Holding now takes the whole ReadyFilter (assignee included) rather than a filter beside an agent argument, and the type keeps the name ReadyFilter
evidence: (1) SIGNATURE. The Design offered "thread the tracker.ReadyFilter through to Holding". The narrowest form of that is Holding(name, scope, agent string, filter ReadyFilter), which leaves TWO places the assignee can come from and lets a caller pass a filter whose Assignee names a different dispatch. The form taken is Holding(name, scope string, filter ReadyFilter), with filter.Assignee as the agent, so AgentQueue sets filter.Assignee = agent once and hands the SAME value to both halves — the parity is then structural rather than a convention two call sites have to keep. Behaviour preserved for an empty assignee: --assignee is emitted unconditionally, exactly as before.
REJECTED, the Design own fallback: applying the exclusions in Held.Within. It would produce the same verdict, but it would leave the Tracker interface saying Holding is unfiltered while something downstream quietly filtered it, and the explain transcript — which is what an agent reads to know what a command will run — would still show an unfiltered `bd list`. The bead asks for the interface to be "honest about what it filters", and only the argv is visible to a reader.
(2) THE TYPE NAME. ReadyFilter is now taken by a query that is not the ready query. Renaming it (to BeadFilter, say) touches 20 references across 10 files and no AC asks for it, so the name stands and its doc comment now says so explicitly: "It is named for the ready query it was introduced for, and it is not that query alone". A rename is a defensible follow-up and is deliberately not taken here.
(3) AC 3 — "Queue.Drained() is true for a dispatch whose only remaining in_progress bead is one the post-merge re-check excludes" — is established by the argv assertion plus the measured bd behaviour above, not by a test that stands such a dispatch up. The queue-level fake (treeTracker in cli/swarm/tracker/descend_test.go) decodes bare ids and carries neither kind nor label, so it cannot honour a filter; making it do so would be asserting against a fake written for this assertion. The bead own Design says reachability is unverified and "Do not spend the session hunting a live reproduction", so no e2e case was built for it. What would settle it end to end: a fixture assigning a human-labelled bead to dev-fixture-2, left in_progress, then swarm dispatch accept against that dispatch.
```

**`agent-plugins-e6xt.13#2`** · dev-e6xt-b12 · 2026-09-02T00:10:37Z

```
retro:decision the stop sentence was extracted into one helper rather than
written twice, which a diff-reader would not recognise as a choice
evidence: the fix adds a second construction site for the breaker's stop
summary — the refused-write path now needs the same sentence the established
path already built inline at the bottom of applyBreaker. Two literals would
have been the smaller diff and would have let the two paths drift into two
different sentences, which is exactly the failure this bead is repairing one
level up (two execution modes answering differently about the same fact).

So cli/swarm/cli/ready.go now carries:

  func stoppedSummary(count, threshold int) string {
      return fmt.Sprintf(
          "the run is stopped: %d consecutive dispatch(es) returned with no bead completed (threshold %d)",
          count, threshold)
  }

and both sites call it — the established path with next, the refused path with
recorded. The count each passes is the whole difference between them, which is
the distinction the bead is about, and it is now the only difference.

REJECTED: a second fmt.Sprintf on the refused path. Cheaper diff, but it makes
the wording of a stop verdict a per-path property.

The test asserts the phrasing comes from the recorded count and not the
projection (it fails if the summary contains "4 consecutive"), so the two
callers cannot be swapped silently.
```

**`agent-plugins-e6xt.17#2`** · dev-e6xt-b12 · 2026-09-02T00:18:24Z

```
retro:decision two things the Design left open were settled silently, and a
diff-reader would not see either as a choice

1. THE FACT IS OMITTED WHEN THERE IS NO INVOCATION TO CITE.
The Design says to append Unestablished("consecutive_dispatch_failures", ...)
on every breaker refusal. One of applyBreaker's three refusals runs nothing at
all: deps.Store == nil returns breakerOutcome{} with no records, so there is no
invocation. Result.Validate refuses that outright --

  cli/swarm/cli/result.go, Validate:
    if len(f.Invocations) == 0 {
        return fmt.Errorf("cli: fact %q names no invocation that established it", f.Name)
    }

-- so appending it unconditionally would have turned a recoverable breaker
fault into an unrenderable result. breakerUnread therefore appends the fact
only when invocationsOf(breaking.Records) is non-empty, and the RESOLUTION
carries the reason on every path. That satisfies the bead's acceptance as
written: "still reports the facts the invocation established before it, or
states explicitly why it cannot".

REJECTED: citing a synthesised invocation so the fact could always be added. It
would name a command nobody ran, which is the fact-fabrication class this
command surface exists to prevent.

2. AN EXISTING RESOLUTION IS APPENDED TO, NOT REPLACED.
reportAgentQueue already sets a resolution on the !established branch ("assign
a bead in <run> to <agent>, or dispatch a different agent id"). Overwriting it
with the breaker's would answer one unsatisfied property by erasing another --
a mistyped dispatch id AND an unreadable breaker are both true at once and a
reader needs both. breakerUnread joins them with "; " when one is already
present.

REJECTED: keeping the first resolution and dropping the breaker's. It leaves
the attention status pointing only at the queue, with nothing saying the stop
verdict is unknown.

Both settled inside the bead because deciding needed a read of Result.Validate
and of reportAgentQueue's own attention branch, which is work inside it.
```

**`agent-plugins-e6xt.14#2`** · dev-e6xt-b12 · 2026-09-02T00:22:34Z

```
retro:decision the invariant was shipped as a TEST, not as a Result.Validate
rule, because a Validate rule would newly refuse results this package is
structurally written to produce

The bead left this bounded choice open and pinned the rule: prefer the test;
use Validate only if the rule would not newly refuse legitimate results. The
reading it asked for -- every OutcomeStop construction -- was done first.

command: grep -rn "OutcomeStop" --include=*.go . | grep -v _test.go
output (constructions only, from cli/):
  ./swarm/cli/branch.go:453   result.Outcome = OutcomeStop     (rejected())
  ./swarm/cli/branch.go:458   result.Outcome = OutcomeStop     (rejected())
  ./swarm/cli/ready.go:269    result.Outcome = OutcomeStop     (reportReady)
  ./swarm/cli/ready.go:618    result.Outcome = OutcomeStop     (reportAgentQueue)
  ./swarm/cli/manifest.go:218 result.Outcome = OutcomeStop
  ./swarm/cli/exit.go:102     &Refusal{Outcome: OutcomeStop}   (RefuseUnrepairable)

None of those pairs a stop with a resolution TODAY. But three call sites copy
an outcome and a resolution together with no stop check, so the shape is one
step away on paths that already exist:

  cli/swarm/cli/accept.go, acceptStopped (line 302 as a hint):
      result.Outcome    = stopped.refusal.Outcome
      result.Resolution = stopped.refusal.Resolution

  cli/swarm/cli/accept.go, settleDispatch fold (line 293 as a hint):
      if settled.outcome != OutcomeClear {
          result.Outcome    = settled.outcome
          result.Resolution = settled.resolution
      }

  cli/swarm/cli/accept_settle.go, the landing settlement (line 670 as a hint):
      outcome:    landed.Outcome,      // can be OutcomeStop, via rejected()
      resolution: landed.Resolution,

A resolution added to rejected() -- "pull --rebase and push again", which is
exactly what a refused push wants to say -- would travel through
that landing settlement into a stop-outcome result and the Validate rule would
then refuse a legitimate answer at render time.

The type's own documentation says the same. cli/swarm/cli/exit.go, Refusal:
  // Resolution names what would resolve the condition. Required when
  // the outcome is [OutcomeAttention]; the surface refuses to render an
  // attention result without one.
Required for attention; not forbidden elsewhere. Result.Resolution's doc is
worded identically. A Validate rule would contradict both.

So: the test ships and Validate is untouched, per the pinned rule. What the
bead observed is real but narrower than "no stop carries a resolution" -- it is
"no stop carries the resolution of the condition IT REPLACED", which is a
property of one override and not of the result type. It is enforced where it
holds, in reportAgentQueue, and asserted by
TestAStoppedRunDoesNotKeepTheSupersededCasesResolution.
```

### `gap` — 13

**`agent-plugins-o63s.1#4`** · dev-e6xt-b1 · 2026-09-01T17:41:04Z

```
retro:gap The delta's intent names FOUR questions the loop asks the tracker; only two of them have a SHALL line, so the fourth had to be decided rather than read.

The intent (blueprint, '### `Tracker` and `Beads Tracker` … the dispatch loop's operations'): 'which epics under a run are eligible to close, what is ready under a scope with the run's filters applied, what one assignee holds, and WHAT A BEAD'S ACCEPTANCE SIGNALS SAY'. The four Behavior bullets under it cover the first three and the shape/plan-step rules. Nothing states what an acceptance-signal operation returns, and the AC table lists 004.11 as satisfied by TRK and proved by RM — so TRK owes something for it with no line saying what.

DECIDED from `plugin/skills/bead-bonanza/scripts/bb-status.sh`, which is the only place the five signals are enumerated: 1 closed in the tracker, 2 a commit carrying the id, 3 merged, 4 worktree gone, 5 findings recorded. Signals 2-4 are Version Control's (the VCS delta). Signals 1 and 5 are the tracker's: 1 is already `Bead`, 5 needed an operation, so TRK ships `Findings(name, id)` over `bd comments <id> --json`.

Evidence that 5 needed a NEW operation rather than a field of an existing one — measured on bd 1.2.2:

    bd show agent-plugins-qozp.1 --json \
      | jq -c '.[0] | {comment_count, has_comments_key: (has("comments"))}'
    -> {"comment_count":1,"has_comments_key":false}

`bd show` carries a count and no comments, so the notes are only reachable through `bd comments`.

What would have removed the guesswork: one Behavior bullet naming the acceptance-signal operation's answer, the way the other three are named. As it stands the operation's shape rests on a shell script that is not part of the spec corpus.
```

**`agent-plugins-o63s.4#3`** · dev-e6xt-b3 · 2026-09-01T17:47:23Z

```
retro:gap nothing restored a fixture, so a corrupt bundle or a bad export would have surfaced as a later bead's case failing rather than as this bead's defect. I added the missing test.
evidence: before this change the only test walking cli/testdata/swarm-fixtures/ was TestEveryFixtureStampAgreesWithItsKind (cli/swarm/e2e/cases_test.go:233), which decodes fixture.toml and never opens repository.bundle or tracker.jsonl; TestSwarmCases only restores fixtures some case names. Five fixtures added by this bead are named by no case yet — the stage-2 beads write those — so all five would have been committed untested. cli/swarm/e2e/fixtures_test.go now restores every fixture through the real git and the real tracker and asserts the sprint branch is checked out, the run epic imported, declared worktrees stand and a declared refusing remote refuses. Measured: 8 subtests, all pass, e2e package 34s to 47s.
```

**`agent-plugins-o63s.5#4`** · dev-e6xt-b4 · 2026-09-01T18:17:45Z

```
retro:gap the FRD and blueprint never say explicitly whether the RUN epic itself (as opposed to its descendants) is in this command's closing scope. Decision: never — the command's own scope is exactly tracker.Descendants(run), which by construction excludes the scope argument itself, so the run epic is left to whatever closes it at retro/post-flight. This reads as the intended split ('the target epic closes on the same rule as the ceremony phases' talks about descendants, never about the run epic), but it is inferred rather than stated.
```

**`agent-plugins-o63s.6#2`** · dev-e6xt-b4 · 2026-09-01T18:27:55Z

```
retro:gap if the Status Collector's own worktree/completion measurement fails (tracker error mid-walk), swarm ready still exits clear and silently reports every bead 'held: false' rather than surfacing the measurement failure — there is no AC in FEAT-SWARM-DISP-002 covering this case (AC-SWARM-DISP-002.16 covers only the ready query's own payload shape, not the reused status.Collect's). Not fixed here: doing so would mean threading status.Collect's own Unestablished reasons into swarm ready's outcome, which is a design decision the FRD does not make and is out of this bead's scope to invent silently.
```

**`agent-plugins-o63s.10#1`** · dev-e6xt-b6 · 2026-09-01T19:14:11Z

```
retro:gap The ACs are silent on whether --dispatch is required, while the FRD's command signature ("swarm dispatch open --run <id> --dispatch <dispatch-id> [--base <ref>] [--branch <name>]") gives it no brackets, which in this doc's own convention reads as required. But the two pre-existing e2e tests from BR (TestDispatchOpenRefusesACollidingBranchTheManifestDoesNotRecord, TestDispatchOpenCreatesAndRecordsTheSprintBranch, cli/swarm/e2e/dispatch_open_test.go) call `dispatch open` with no --dispatch at all and expect success. Decided: --dispatch stays optional at the flag level; absent, only the sprint-branch half runs (unchanged from before this bead), and the worktree half runs only when a non-empty --dispatch is given. Rejected alternative: making --dispatch mandatory, which would have broken both pre-existing tests (a baseline regression) for a requirement no AC actually states.
```

**`agent-plugins-o63s.11#1`** · dev-e6xt-b4 · 2026-09-01T19:48:17Z

```
retro:gap the MG delta names no version-control operation for returning to the target after a rebase, and none existed. `git rebase <onto> <branch>` checks <branch> out, so the fast-forward that follows would advance the dispatch branch onto itself and leave the target unmoved. Added `Checkout(prefix, branch)` to `vcs.VersionControl` and its git implementation.
evidence: measured, git 2.43.0, scratch repo on `main`: `git rebase main bb/dev-1` -> exit 0, ""Successfully rebased and updated refs/heads/bb/dev-1."", then `git rev-parse --abbrev-ref HEAD` -> `bb/dev-1`. After `git checkout main; git merge --ff-only bb/dev-1` -> exit 0, "Updating eef6bea..43e1b46 / Fast-forward", and `git rev-parse main bb/dev-1` reports one sha twice.
```

**`agent-plugins-o63s.12#2`** · dev-e6xt-b4 · 2026-09-01T20:02:30Z

```
retro:gap AC-SWARM-DISP-005.3 and .7 require the rejection cause and the refused path to reach the caller, but `Push` decoded with `discard`, whose error is the LAST line of stderr — which for a refused push is `error: failed to push some refs to <url>`, naming neither. Added a `vcs.Rejection` type and a push decoder that reads the two lines that carry the answer. This is the "proving delta (VCS participates)" the item names, so the participation was expected; what the item did not say is that the landed decoder threw the evidence away.
evidence: measured, git 2.43.0 against a bare repo whose pre-receive hook declines refs/heads/claude/*: exit 1, stdout EMPTY, stderr = "remote: refusing refs/heads/claude/*: this remote does not accept sprint branches" / "To ../remote.git" / " ! [remote rejected] claude/sprint -> claude/sprint (pre-receive hook declined)" / "error: failed to push some refs to ../remote.git". `lastLine` over that returns the fourth line.
```

**`agent-plugins-o63s.12#5`** · dev-e6xt-b4 · 2026-09-01T20:02:56Z

```
retro:gap the delta names the branch default as "the one currently checked out" but the Version Control interface had no operation answering it — `Branches` lists every branch without saying which is checked out, and `Resolve` answers a sha. Added `CurrentBranch(prefix)`. It also forced a decision the delta does not cover: a detached HEAD. Refused with the attention status naming `--branch`, rather than checking a remote ref called `HEAD`, which would produce a true report about the wrong question.
evidence: measured, git 2.43.0: `git rev-parse --abbrev-ref HEAD` prints "main" on a branch and the literal "HEAD" after `git checkout --detach`, exit 0 both times.
```

**`agent-plugins-o63s.13#1`** · dev-e6xt-b7 · 2026-09-01T20:41:17Z

```
retro:gap the signal set behind "re-check every bead with no signal deferred" is named nowhere in the FRD, the blueprint or the bead
evidence: `grep -n "signal" specs/features/swarm-dispatch/swarm-dispatch-frd.md` returns 9 hits, and every one uses the word without enumerating the set: line 679 "pre-merge acceptance the acceptance signals per bead, deferring the two a merge decides", 684 "post-merge acceptance the same signals with none deferred", 690 "a bead failing an acceptance signal at exit 2", 799 "each bead in the dispatch is checked again with no signal deferred", 806 "the failing signal is named". `grep -rn "acceptance" cli/swarm --include=*.go | grep -v _test` returns 5 hits, all doc comments, none a signal set. So the phrase "with no signal deferred" is unimplementable as written: nothing says which signals, nor which two a merge decides.
The set I implemented is plugin/skills/bead-bonanza/scripts/bb-status.sh mode_bead: 1 closed in the tracker, 2 a commit carrying the id, 3 merged into HEAD, 4 the worktree is gone, 5 findings recorded — and its `--pre-merge` flag defers exactly 3 and 4 ("--pre-merge DEFERS SIGNALS 3 AND 4 ... both signals would fail by construction"), which is what identifies "the two a merge decides". `cli/swarm/tracker/tracker.go:280` is the corroboration inside the container: Findings is documented as "the tracker's half of a bead's acceptance signals: whether the work was recorded, beside the bead's own state."
The derivation is sound but it is a derivation. A reader of the FRD alone cannot make it, and the script it comes from is a skill this container is not allowed to depend on.
```

**`agent-plugins-o63s.16#2`** · dev-e6xt-b9 · 2026-09-01T21:04:48Z

```
retro:gap dispatching.md's 'Verify against state' step 3 was the ONLY place the PRE-merge acceptance check is described, and no swarm command performs it — deleting the step whole, as the AC's list says, would have removed the only statement of a still-required step. Deleted the post-merge half only and left the pre-merge invocation with a line saying no command performs it yet.
evidence: FEAT-SWARM-DISP-004's ceremony table names step 2 'pre-merge acceptance'; `swarm dispatch accept --help` says only 'Every bead is then re-checked with no signal deferred' — the post-merge half. agent-plugins-bhaj (open, P2) records the same gap: 'Step 2 is implemented nowhere, and NO delta's behavior lines cover it.'
```

**`agent-plugins-o63s.18#3`** · dev-e6xt-b10 · 2026-09-01T21:24:46Z

```
retro:gap FOLD-3's acceptance criteria named four component records but not the Overview 'Consumes' bullet, which is the module's own dependency statement and would have been left stale by this fold. I updated it anyway.
evidence: Before: 'The run manifest is reached through `swarm manifest`, and a run's progress, completion and per-dispatch drain through `swarm status`.' That enumeration named two of the eight swarm commands this module now calls, and stage 3 is exactly the change that added the other six. Leaving it would have shipped a module blueprint whose Overview contradicts its own component records. Decided to extend it in the same fold rather than file a bead, because a fold that leaves the folded fact contradicted at the top of the document has not folded it.
```

**`agent-plugins-e6xt.8#3`** · dev-e6xt-b11 · 2026-09-01T22:38:02Z

```
retro:gap the two specs the bead names were already consistent with the FIXED behaviour, so nothing in them could have caught the defect; one bullet was added to record the invariant on the component that now enforces it
evidence: `schemas/swarm/output.strata` lines 175-183 already declare command as "The invocation path, for example \"swarm manifest get\"", run as "The run epic this invocation was scoped to. Empty for a command that is not run-scoped", and mode as "Which interpreter ran." `specs/contracts/contract-swarm-json.md` states obligations about facts, invocations and versioning and says nothing about these three. So the schema described what the code did NOT do, and no check compares the two: the generated model makes every key optional on the wire, so an empty command and a zero-value mode validate.
Added to specs/architecture/container-swarm-cli.md, in the Swarm Command Surface bullet list: "Every result names the invocation that produced it ... and the surface stamps all three rather than each command setting them. A result built without them does not merely omit them: explain is the mode zero value ..." — so the invariant is recorded against the component that holds it rather than only in the schema that describes its output.
NOT FIXED, and out of this bead: nothing validates a rendered result against `cli/core/schema/swarm-model.schema.json`, which is why a five-command regression against a correct schema shipped. A test that rendered each command result and validated it would have caught this class without naming the three fields at all.
```

**`agent-plugins-e6xt.8#4`** · dev-e6xt-b11 · 2026-09-01T22:38:27Z

```
retro:gap (addendum) the schema-validation gap above is filed as agent-plugins-8vco, discovered-from this bead; adjacent, not blocking.
```

### `scope` — 1

**`agent-plugins-o63s.16#3`** · dev-e6xt-b9 · 2026-09-01T21:04:49Z

```
retro:scope one AC line of this item lands in a file this item is forbidden to touch, so it is NOT done here: 'The Coordinator's roster line and Change Guidance SHALL name the commands it now depends on.' Both live in specs/architecture/module-bead-bonanza.md — the roster table row and the '## Change Guidance' section — which the item's own Design reserves for FOLD-3 ('Do NOT update module-bead-bonanza.md's component records here; FOLD-3 does that for both adoption deltas').
evidence: `grep -n 'Coordinator|## Change Guidance' specs/architecture/module-bead-bonanza.md` -> 198 (roster row '| **Coordinator** | what happens next...'), 616 ('## Change Guidance'). No other file in the tree carries a Coordinator roster line.
```

### `stale` — 4

**`agent-plugins-o63s.3#4`** · dev-e6xt-b3 · 2026-09-01T17:09:11Z

```
retro:stale:project the core fixture's own description understates what its manifest carries, and re-recording it made that visible.
evidence: cli/testdata/swarm-fixtures/core/fixture.toml says 'its manifest is complete through the scope stage, its branch and base are the bundle's own refs'. The stored manifest also carries collision_edges and partition_preview (plan stage) and review_passes (dispatch), none of which the description accounts for — it names branch and base by hand instead. Nothing breaks: no test reads the description. But an agent choosing which fixture to use reads it, and 'complete through the scope stage' is the wrong mental model for a record that is complete through dispatch. Left unchanged here because editing it is not this delta's lockstep obligation.
```

**`agent-plugins-o63s.4#5`** · dev-e6xt-b3 · 2026-09-01T17:47:41Z

```
retro:stale:project container-swarm-cli.md's Data section names the fixture set by enumeration, and this bead makes that sentence wrong. FOLD-1 owns the edit; this names the exact string so it is not missed.
evidence: specs/architecture/container-swarm-cli.md, in the '## Data' entry for the end-to-end fixtures: 'Three are committed today — core, status-attention and status-narrowed-scope.' After this bead there are eight. The blueprint's Definition of Done already routes it — '[[container-swarm-cli]] — ... Data and Change Guidance updated for the new manifest key and the fixture variants' — and agent-plugins-o63s.14's Design forbids the delta beads from editing that file, so no bead is filed here. The same section will also need the two new fixture.toml declarations ([[worktree]] and [refusing_remote]) stated as part of what a fixture directory may hold, since Change Guidance currently says a fixture is 'repository.bundle, tracker.jsonl and fixture.toml' without saying what fixture.toml may declare.
```

**`agent-plugins-o63s.19#2`** · dev-e6xt-b10 · 2026-09-01T21:37:50Z

```
retro:stale The item's own completion query omits --status all, so run literally it reports 3 issues instead of the 19 it declares.
evidence: The AC says 'the run's work is exactly bd list --parent agent-plugins-o63s --exclude-type=epic — 19 beads. Verify against that query, not against memory of what was created.' Measured on bd 1.2.2: that exact command prints 'Total: 3 issues (1 open, 1 in progress)' — bd excludes closed issues by default, so the 18 closed children of the run are absent, and the query reports the OPPOSITE of a complete run. The same command with --status all prints 'Total: 21 issues', which is the epic itself plus .1 through .19 plus the deferred .20; the declared 19 is .1 through .19. A verification query that reads 3 on a finished run and 19 on an unfinished one is worse than no query.
```

**`agent-plugins-e6xt.16#2`** · dev-e6xt-b12 · 2026-09-02T00:14:44Z

```
retro:stale this bead's own Design asserts a hazard that does not exist as
stated: mustRecord does not panic on the failure path
evidence: the Design reads "the established path cites the invocation via
argvOf(mustRecord(report, landedSteps.Answer())), and mustRecord is not safe on
the failure path -- the record may be absent, which is one of the ways revision
fails ... rather than panicking on the path that exists to handle a failure."
The same claim reached me in the dispatch prompt, worded "the established path
cites via mustRecord, which PANICS on the failure path".

Measured. cli/swarm/cli/dispatch.go:318 --

  func mustRecord(report plan.Report, name string) plan.Record {
      rec, _ := report.Record(name)
      return rec
  }

It discards the ok and returns whatever Record came back. cli/swarm/plan/plan.go:367 --

  func (r Report) Record(step string) (Record, bool) {
      for _, rec := range r.Records {
          if rec.Step == step {
              return rec, true
          }
      }
      return Record{}, false
  }

An absent record yields the ZERO Record, not a panic. And
cli/swarm/cli/phase.go:442 --

  func argvOf(rec plan.Record) []string {
      if rec.Invocation.Name == "" {
          return nil
      }
      return append([]string{rec.Invocation.Name}, rec.Invocation.Args...)
  }

-- guards the zero value explicitly and returns nil. There is no index, no
dereference and no map access on that path. mustRecord's own doc comment is
"reads a named record out of a report the caller has already established
carries it", which is a PRECONDITION, not a panic.

The underlying concern is real, only misattributed: the failure mode is a
SILENT one. argvOf(mustRecord(...)) on an absent record produces a nil argv
that is still passed as one variadic invocation, so the fact would be attributed
to an empty invocation rather than to none. That is what the fix guards, using
the report.Record(...) + argvOf form the Design also offers:

  var landedInv [][]string
  if rec, ok := report.Record(landedSteps.Answer()); ok {
      if argv := argvOf(rec); argv != nil {
          landedInv = append(landedInv, argv)
      }
  }

so both the established and the unestablished merged_sha cite the resolve when
there is one to cite and nothing when there is not.

Sweep for other copies, run BEFORE the source was touched:
  command: grep -rn "mustRecord" --include=*.go --include=*.md .
  output: 11 hits, all Go call sites in cli/swarm/cli plus the definition and
  its doc comment. No document, skill or reference states the panic claim.
  command: grep -rn "PANICS|panics on the failure path" --include=*.md --include=*.go .
  output: no matches.

UNREACHED COPIES, recorded rather than fixed: the claim also stands in this
bead's pinned Design in the tracker (I have not rewritten a pinned Design) and
in batch b12's dispatch prompt, which is not in the repository and which I
cannot edit. A future agent reading either inherits it. Nothing in the repo
carries it.
```

### `sequencing` — 3

**`agent-plugins-o63s.11#7`** · dev-e6xt-b4 · 2026-09-01T19:49:12Z

```
retro:sequencing the concurrent arm this bead delivers cannot run against a real dispatch worktree: git refuses to rebase a branch another worktree has checked out, and the FRD orders the merge BEFORE the worktree removal. Filed as agent-plugins-1vya, discovered-from this bead. Not fixed here — the resolution touches RM ordering or the FRD, and this bead ACs are proved against the fixtures FIX recorded, which declare no worktree.
evidence: measured, git 2.43.0: after `git worktree add -q ../wt2 bb/dev-1`, `git rebase main bb/dev-1` -> exit 128, "fatal: bb/dev-1 is already used by worktree at .../wt2". `grep -l worktree cli/testdata/swarm-fixtures/dispatch-stale-base/fixture.toml cli/testdata/swarm-fixtures/dispatch-rebase-conflict/fixture.toml` matches neither.
```

**`agent-plugins-o63s.13#6`** · dev-e6xt-b7 · 2026-09-01T20:42:20Z

```
retro:sequencing agent-plugins-1vya (MG cannot rebase while the dispatch worktree stands) was left untouched, and this bead did not have to decide anything that entrenches it
evidence: 1vya is about WHERE the rebase runs. This delta owns steps 5-8 of the ceremony — the ancestry guard, the removal, the post-merge re-check and the landing — all of which are strictly after the fold. The merge-before-removal order is fixed by the FRD ceremony (swarm-dispatch-frd.md:679-686) and was implemented by MG; nothing here moved it, and no acceptance criterion of this bead required choosing a `Dir` for the rebase. So the rebase still runs in the coordinator's own checkout, exactly as MG left it.
The two e2e cases this bead added or changed do not hide it either, and one of them is now the first case in the suite to accept a dispatch whose worktree really stands: TestDispatchAcceptRechecksEveryBeadAfterTheMergeAndTheRemoval uses the dispatch-worktree fixture, which declares [[worktree]] branch = "bb/dev-fixture-2". It is the SERIALIZED case (the bundle's target is at ac4a37c and the branch is ac4a37c + 016160f, so the merge-base IS the target's head), so no rebase runs and 1vya is not reached. The concurrent case still runs against dispatch-stale-base, which declares no worktree.
A fixture that is both concurrent AND has a standing worktree would reproduce 1vya end to end today. None exists.
```

**`agent-plugins-o63s.18#1`** · dev-e6xt-b10 · 2026-09-01T21:24:25Z

```
retro:sequencing ADP-C carried an acceptance line it was structurally forbidden to satisfy, and handed it to FOLD-3 out-of-band rather than through the graph.
evidence: ADP-C (agent-plugins-o63s.16) AC: 'The Coordinator's roster line and Change Guidance SHALL name the commands it now depends on.' Its own Design says: 'Do NOT update module-bead-bonanza.md's component records here; FOLD-3 does that for both adoption deltas.' Both the roster line and Change Guidance live in specs/architecture/module-bead-bonanza.md, the file ADP-C was told not to touch. Nothing in FOLD-3's own acceptance criteria mentions the roster line or Change Guidance; the obligation reached this bead only through the dispatch prompt. Had that prompt not carried it, the line would have been silently dropped by both beads.
```

### `unverifiable` — 3

**`agent-plugins-o63s.13#5`** · dev-e6xt-b7 · 2026-09-01T20:42:06Z

```
retro:unverifiable AC-SWARM-DISP-004.1 cannot be verified at its declared e2e level — proved at unit instead, filed as agent-plugins-oyvf
evidence: The blueprint's AC Coverage table declares "AC-SWARM-DISP-004.1 | MG, RM (proves) | e2e". The criterion needs a dispatch branch that is NOT an ancestor of the target AFTER a merge step that did not stop the command. Against a real git those are mutually exclusive: a successful `git merge --ff-only <branch>` has by construction put the branch on the target, so `git merge-base --is-ancestor <branch> <target>` answers yes; and a fold that fails stops the acceptance at AC-SWARM-DISP-004.6, before the guard. The state exists only when a second actor moves the target between the two reads.
Proved at the unit tier instead — TestAWorktreeIsNotRemovedUnlessItsBranchIsAnAncestorOfTheTarget, where the fake answers the fold 0 and the ancestry read 1 independently: no `git worktree remove`, no `git branch -d`, no `git push`, exit 2, report says "not an ancestor". `go test ./swarm/cli/...` ok.
What would settle it at e2e: a fixture world carrying a real repository-local `post-merge` hook that moves the target — the same technique push-refused already uses for a rejection (a real pre-receive hook, standRefusingRemote in cli/swarm/e2e/harness_test.go). That needs a fixture-stamp declaration, which is the fixture harness's and not a command delta's.
```

**`agent-plugins-o63s.19#1`** · dev-e6xt-b10 · 2026-09-01T21:37:36Z

```
retro:unverifiable Two Definition-of-Done entries cannot be closed as written, and the blueprint now ships saying so rather than ticking them.
evidence: (1) 'Every AC in the coverage table is verified at its declared level' — all 58 carry a test citation, 57 at their declared level; AC-SWARM-DISP-004.1 is declared e2e and proved at unit (TestAWorktreeIsNotRemovedUnlessItsBranchIsAnAncestorOfTheTarget, cli/swarm/cli/accept_settle_test.go), because a successful merge --ff-only implies ancestry and a failed one stops at AC-SWARM-DISP-004.6 before the guard. Carried by agent-plugins-oyvf. (2) 'Every prose section named in the ADP-C and ADP-D deltas is deleted' — the gate passes and every named section is gone except step 3 of dispatching.md's 'Verify against state', the per-bead pre-merge acceptance check, which ADP-C deliberately kept because no swarm command performs that half. Carried by agent-plugins-bhaj. Both are recorded as NOT MET in the blueprint's own Definition of Done with the carrying issue named.
```

**`agent-plugins-e6xt.19#2`** · dev-e6xt-b12 · 2026-09-02T00:29:11Z

```
retro:unverifiable the second resolution the Design required fixing is not
reachable by any test in this package, so it is corrected but unasserted

The Design named two resolutions and called fixing both a stated scope decision
rather than creep. Both are fixed, and they now derive the verb from one local:

  verb := "verify-landed"
  if pushFirst {
      verb = "push"
  }

1. The unfetched-remote verdict. Reachable and asserted -- it is what
   TestADryRunDoesNotReportLandedFromAnUnfetchedComparison drives, and the
   strengthened assertion covers both verbs.

2. The divergence-step verdict, guarded by
   `if !ok || divergeRec.Status != plan.StepRan` AFTER the fetch has run and
   the counterpart probe has succeeded. I could not construct that state
   against the exectest fake: the divergence step is in the same plan as the
   two reads before it, so a plan in which the fetch RAN and the comparison did
   not is a plan that stopped between them, and the fake replays declared argvs
   rather than truncating a plan. Its own comment says as much -- "The fetch
   ran, so this is a plan that stopped before the comparison".

So the second correction rests on the two resolutions sharing one `verb`
variable four lines apart, not on a test. What would settle it: a seam that
lets a test hand landingVerdict a plan.Report directly, with the fetch and
counterpart records present and the divergence record absent or non-StepRan.
landingVerdict already takes a plan.Report and vcs.Steps, so such a test is a
unit test on the function rather than through the command tree, and it would
cover both resolutions and the no-counterpart one in the same table.

Not filed as a bead: it is test infrastructure for a path this bead did not
introduce and no acceptance criterion asks for, and the bead's own acceptance
is satisfied by the strengthened both-verbs test. Recorded here so the gap is
visible rather than implied by a green check set.
```

### `tooling` — 7

**`agent-plugins-o63s.9#3`** · dev-e6xt-b4 · 2026-09-01T18:39:17Z

```
retro:tooling caught in my own session, not shipped: a test asserting on captured stdout from dispatch() must pass the SAME buffer as both ioStreams.out and dispatch()'s stdout parameter — surface_test.go's runSurface does this (streams built from &out, then dispatch(root, streams, args, &out, &errOut) with the identical &out), but a test that builds streams from one buffer and passes a SEPARATE buffer to dispatch() silently asserts against an always-empty buffer and passes vacuously. Two of my own tests (ready_test.go, dispatch_test.go) had this exact bug during development; caught because dispatch_test.go's order-of-calls assertion still failed even though its (bugged) content assertion could not have. No AC covers it and no existing test in the package would have caught a REGRESSION here — a future test built by copy-paste from a --run-only refusal test (which never inspects stdout) risks the same silent pass.
```

**`agent-plugins-o63s.8#4`** · dev-e6xt-b5 · 2026-09-01T19:16:03Z

```
retro:tooling The committed e2e golden case cli/testdata/swarm-cases/ready-dispatch-set/expected/stdout hard-codes the bare swarm ready answer's exact field set. Adding CB's consecutive_dispatch_failures fact to that answer (required by the delta: the command SHALL read the count on every query) broke go test ./swarm/e2e's TestSwarmCases/ready-dispatch-set on an otherwise-correct change — expected, and fixed by appending the new field (confirmed 0 from the dispatch-worktree fixture's own tracker.jsonl) to the golden file, but worth flagging: any later delta touching this command's answer shape will hit the same golden-file update and should expect it rather than treat it as a regression.
```

**`agent-plugins-o63s.12#6`** · dev-e6xt-b4 · 2026-09-01T20:03:06Z

```
retro:tooling the answer document has TWO shapes and which one a case gets depends on the exit status — a clear exit prints each property value directly, every other status prints `{ok, value, issues}` per property. An e2e case asserting across both statuses (unpushed -> attention, then pushed -> clear) cannot use one struct for both, and the first attempt failed with "cannot unmarshal object into Go struct field landing.branch of type string". Nothing in the item or the output standards warns of it; the reader learns it from `cli/swarm/cli/answer.go`.
evidence: cli/swarm/cli/answer.go, `func answer(r Result)`: "if r.Status() == StatusClear { ... established } ... verdicts". Observed in `swarm branch verify-landed --json` at exit 1: {"issues":[...],"branch":{"ok":true,"value":"claude/bead-bonanza-fixture","issues":[]}, ...}.
```

**`agent-plugins-o63s.15#1`** · dev-e6xt-b8 · 2026-09-01T21:07:43Z

```
retro:tooling run-batch's ceremony has no guidance on how to wait for a long-running foreground check (bb-checks.py takes 1-2 min on this repo). I backgrounded it with run_in_background and then tried to idle-wait for the notification across several empty turns, which does not work in this harness the way I assumed — the coordinator had to interrupt and correct me to run it synchronously and read $? directly. AGENTS.md's pipe-hides-exit-status trap is well documented; the backgrounding trap for a check step that gates a close is not. Consider a line in run-batch's checks step: run bb-checks.py in the foreground and read its exit status directly; do not background it and then wait on the result.
```

**`agent-plugins-o63s.17#1`** · dev-e6xt-b9 · 2026-09-01T21:12:30Z

```
retro:tooling:skill run-batch/SKILL.md sits 32 bytes under check-skills.sh's 20000-byte body cap, so an adoption edit that nets even one extra sentence fails skills-strict. This item's replacement was net-positive on first write and I had to trim three separate paragraphs — twice — to get back under, which is editing to a byte budget rather than to the rule.
evidence: 'git show HEAD:plugin/skills/run-batch/SKILL.md' body = 19968 bytes. After the swarm ready --agent adoption as first written: 'scripts/check-skills.sh --strict' -> exit 1, 'WARN plugin/skills/run-batch/SKILL.md: body is 20228 bytes (~5057 tokens)', RESULT: FAIL. After trimming: 19994 bytes, exit 0. Six bytes of headroom.
```

**`agent-plugins-o63s.19#3`** · dev-e6xt-b10 · 2026-09-01T21:37:51Z

```
retro:tooling An AC-citation audit by plain grep reports five false gaps, because the tree cites continuation ids in a shorthand the pattern does not match.
evidence: grep -rno 'AC-SWARM-DISP-[0-9]*\.[0-9]*' cli/ --include=*.go finds 53 of the 58 declared ids and reports AC-SWARM-DISP-002.7, 002.9, 002.10, 003.10 and 003.12 as uncited. All five ARE cited: the source writes a run of ids as 'AC-SWARM-DISP-003.9, .10, .12' (cli/swarm/cli/dispatch_test.go) and 'AC-SWARM-DISP-003.1, .10, .12' (cli/swarm/cli/dispatch.go), where the trailing members carry no feature prefix; 002.7/.9/.10 are declared in a TOML fixture list (cli/testdata/swarm-cases/ready-dispatch-set/case.toml) rather than in Go. An audit must expand the ', .N' continuation against the preceding feature number and must read the case fixtures as e2e citations. With that expansion all 58 resolve, and the only real defect is AC-SWARM-DISP-004.1's level. Nothing in the tree performs this audit — it was written by hand for this item and thrown away; that is why the level defect survived to the DoD.
```

**`agent-plugins-e6xt.19#3`** · dev-e6xt-b12 · 2026-09-02T00:29:17Z

```
retro:tooling the `format` check is the one gate a narrowed run does not catch,
and it fired here after every targeted test had passed

Sequence, measured in this bead:
  cd cli && go test ./swarm/cli/ -run 'TestADryRun|TestAFailedFetch|Branch|Land'
  exit 0
  cd cli && go build -o /usr/local/bin/swarm ./swarm
  exit 0
  python3 plugin/skills/bead-bonanza/scripts/bb-checks.py
  exit 1
    === format
    unformatted:
    swarm/cli/branch.go
    === format FAILED (exit 1)
  (every other check passed in the same run)

The cause was a resolution string reflowed across two lines with an indentation
gofmt does not accept. Nothing in the compile, the vet or the tests can see it,
so the fast loop the skill recommends -- `bb-checks.py <id>...` while iterating
-- is silent on it by construction unless `format` is one of the ids picked.

Fixed by `gofmt -w ./swarm/cli/branch.go`; the full set then exited 0, matching
the run's baseline.

Worth noting because run-batch's guidance for the fast loop is "pick by what
you touched", and `format` reads as a check about formatting rather than about
a surface -- so a Go-only bead naturally picks `test-go` and `vet` and not
`format`. AGENTS.md already documents the neighbouring trap (gofmt -l exits 0
whether or not it found anything, which is why the check points at
scripts/format-check); this is the adjacent one, that `format` is the check
least likely to be selected and the one most likely to fire.
```

### `proof` — 30

**`agent-plugins-o63s.3#1`** · dev-e6xt-b3 · 2026-09-01T17:08:43Z

```
retro:proof the backwards-refusal rule was proved by a negative control, not by a green test.
evidence: with checkFailureCount commented out of CheckWrite, `go test ./swarm/manifest/ -run TestTheDispatchFailureCount -v` → exit 1, 'validate_test.go:179: one lower: the write was accepted' and 'almost a reset: the write was accepted'. With the call restored → exit 0, PASS. The test also carries three positive controls (rise to 4, reset to 0, re-write the same 3) and a first-write-against-an-empty-record control, so a rule that refused every write to the key could not pass it.
```

**`agent-plugins-o63s.1#1`** · dev-e6xt-b1 · 2026-09-01T17:40:08Z

```
retro:proof `bd ready` silently truncated the ready set at 100, and this delta's operation was one of the callers — reproduced pre-fix, then fixed.

Measured on bd 1.2.2 (`bd version` -> `bd version 1.2.2 (6c124203e: HEAD@6c124203e771)`) against this repository's database:

    bd ready --json          | jq length  -> 100
    bd ready --limit 0 --json | jq length -> 106

`Tracker.Ready` built its argv without `--limit`, so every ready query the dispatch loop is built on would have stopped at 100 beads at exit 0. The unit reproduction, run BEFORE the fix with the assertion updated and the source still carrying the old argv:

    $ go test ./swarm/tracker/... -run TestTheInvocationsItDeclares
    --- FAIL: TestTheInvocationsItDeclares (0.00s)
        beads_test.go:113: the whole ready set:
              got  bd ready --parent run-1 --json
              want bd ready --parent run-1 --limit 0 --json
        beads_test.go:113: a filtered ready set:
              got  bd ready --parent run-1 --assignee dev-1 --exclude-type epic --exclude-label human --json
              want bd ready --parent run-1 --assignee dev-1 --exclude-type epic --exclude-label human --limit 0 --json
    FAIL
    exit=1

Same command after adding `--limit 0` to the argv: exit 0.

Note the asymmetry that makes this non-obvious, and that `Children`'s existing comment gets backwards for `ready`: the default DOES apply to `bd ready`'s JSON output but did NOT to `bd list`'s — `bd list --all --json | jq length` and `bd list --all --limit 0 --json | jq length` both returned 585. Neither flag can be reasoned from the other.
```

**`agent-plugins-o63s.2#1`** · dev-e6xt-b2 · 2026-09-01T17:40:40Z

```
retro:proof Both load-bearing guards were confirmed by negative control before the bead closed — each was made to fail, then restored.
evidence:
(1) The no-force obligation. Injected the forcing form:
    sed -i 's|"worktree", "remove", path)|"worktree", "remove", "--force", path)|' swarm/vcs/git/git.go
    go test ./swarm/vcs/git/ -run 'TestTheWorktreeRemovalHasNoForcingForm|TestTheInvocationsItDeclares'
  -> exit 1
     --- FAIL: TestTheInvocationsItDeclares
         git_test.go:130: remove a worktree:
           got  [git worktree remove --force /w/dev-1]
           want [git worktree remove /w/dev-1]
     --- FAIL: TestTheWorktreeRemovalHasNoForcingForm
         git_test.go:273: the removal invocation carries the option "--force"; it is meant to carry none
         git_test.go:287: git.go names "--force" — the forcing form of a removal or a push
  Reverted; the same command then exits 0.

(2) The write/dry-run declaration. Flipped Fetch from plan.Write to plan.Read:
    go test ./swarm/vcs/...
  -> exit 1
     --- FAIL: TestADryRunPerformsNoneOfTheDispatchLoopsWrites
         vcs_test.go:189: a dry run ran [[git fetch origin]]
         vcs_test.go:198: "fetch" was "ran" in a dry run, want "refused"; it is not declared a write
     --- FAIL: TestEveryWriteDeclaresWhetherReRunningItIsSafe
         git_test.go:232: fetch is declared read, want write; a dry run would perform it
  Reverted; exit 0.

Every Repeatable declaration on the eight new writes is a measurement taken in a scratch repository on git 2.43.0, not an inference. The three that matter, because they are the ones a recovery depends on:
  git merge --ff-only <b> twice   -> exit 0 both times, second says "Already up to date."; and a REFUSED --ff-only leaves `git status --porcelain` empty, HEAD unmoved, no MERGE_HEAD or rebase directory  => Repeatable true
  git push origin main twice      -> exit 0 both, second says "Everything up-to-date"  => Repeatable true
  git worktree remove <p> twice   -> exit 0 then exit 128, "fatal: '<p>' is not a working tree"  => Repeatable false
```

**`agent-plugins-o63s.2#4`** · dev-e6xt-b2 · 2026-09-01T17:40:55Z

```
retro:proof git prints merge and rebase CONFLICT lines on STDOUT, not stderr — a decoder reading stderr for the reason finds an empty stream.
evidence: Measured on git 2.43.0 in a scratch repository, conflicting `git merge --no-ff -m "merge dev" dev`:
  exit 1
  stdout:
    Auto-merging f.txt
    CONFLICT (content): Merge conflict in f.txt
    Auto-merging g.txt
    CONFLICT (content): Merge conflict in g.txt
    Automatic merge failed; fix conflicts and then commit the result.
  stderr: (empty)

Conflicting `git rebase main dev`:
  exit 1
  stdout: the same CONFLICT lines
  stderr: Rebasing (1/1)error: could not apply 8a198f8... dev  + hints

Refused `git merge --ff-only dev` is the opposite — nothing on stdout, and the reason is on stderr behind several hint lines:
  exit 128
  stderr: ... hint: Diverging branches can't be fast-forwarded ...
          fatal: Not possible to fast-forward, aborting.

So the refusal decoder scans BOTH streams. Reading only stderr would have reported a conflicting merge as a refusal with no paths; reading only stdout would have missed every non-fast-forward.

Four conflict-line forms were measured, and the path decoder is held to them:
  CONFLICT (content): Merge conflict in f.txt
  CONFLICT (add/add): Merge conflict in sp ace.txt
  CONFLICT (modify/delete): del.txt deleted in dev and modified in HEAD.  Version HEAD of del.txt left in tree.
  CONFLICT (rename/rename): a.txt renamed to c.txt in HEAD and to b.txt in dev.

A path is printed UNQUOTED, spaces included — `sp ace.txt` above is one path, and any whitespace split would have made it two. That is why the decoder takes everything after "Merge conflict in " to end of line rather than tokenising.

Residual limit, stated rather than hidden: a path containing the verb phrase " deleted in " or " renamed to " would be cut short by the second and third rules, and nothing in git's output distinguishes that case. Every CONFLICT line is therefore kept verbatim in Refusal.Detail regardless of whether a path was extracted, so a conflict this cannot parse is visible rather than silently absent (TestAConflictLineIsKeptEvenWhenItsPathCannotBeParsed).
```

**`agent-plugins-o63s.1#6`** · dev-e6xt-b1 · 2026-09-01T17:41:37Z

```
retro:proof Negative control on the one assertion that carries this delta's scoping claim, so it is not a test that passes for the wrong reason.

`TestAnAgentsQueueDoesNotReachOutsideTheRun` asserts that a bead the same dispatch id holds in ANOTHER run does not reach this run's queue. Control: widen the descendant set the narrowing is taken against, changing `held.Within(beneath)` to `held.Within(append(beneath, Bead{ID: "other-9.1"}))`:

    $ go test ./swarm/tracker/ -run TestAnAgentsQueueDoesNotReachOutsideTheRun
    --- FAIL: TestAnAgentsQueueDoesNotReachOutsideTheRun (0.00s)
        queue_test.go:50: InProgress = [other-9.1 run-1.2], want only run-1.2 — other-9.1 is another run's
    FAIL
    exit=1

Restored, exit 0. So the fake really does deliver both ids and the assertion is what removes the foreign one.

Also decided silently, and worth knowing: `AgentQueue` tracks which halves ANSWERED rather than only which failed. A step the plan runner SKIPS records no failure, so a failure check alone would let a skipped half arrive as an empty one — and an empty half is exactly what makes a queue read as drained. The guard is the `answered` map at the end of `AgentQueue`; `TestAHalfThatCouldNotBeReadIsNotAnEmptyHalf` covers the failing half, not the skipped one, which stays untested because the fake cannot produce a skip.
```

**`agent-plugins-o63s.4#2`** · dev-e6xt-b3 · 2026-09-01T17:47:13Z

```
retro:proof each world was proved to exhibit its condition by running the condition, not by inspecting the bundle.
evidence: rebase probes on the built bundles — dispatch-stale-base: 'git rebase origin/claude/bead-bonanza-fixture' on bb/dev-fixture-2 exits 0, 'Successfully rebased and updated refs/heads/bb/dev-fixture-2', no unmerged paths; dispatch-rebase-conflict: exits 1, 'CONFLICT (content): Merge conflict in work.txt', 'git diff --name-only --diff-filter=U' names work.txt. Merge-bases measured: dispatch-worktree's bb/dev-fixture-2 has merge-base ac4a37c == sprint head ac4a37c (serialized), while stale-base and rebase-conflict have merge-base ac4a37c against sprint heads d54d47f and 078f2b6 (concurrent). Negative controls on the new harness declarations, both run: changing push-refused's refuses to 'refs/heads/nothing-matches-this/*' made the push SUCCEED and TestEveryFixtureRestores/push-refused fail with 'the push to the refusing remote succeeded'; changing dispatch-worktree's branch to bb/dev-fixture-99 failed the restore with 'fatal: invalid reference: bb/dev-fixture-99'. Both were reverted and the suite re-run green.
```

**`agent-plugins-o63s.14#1`** · dev-e6xt-b8 · 2026-09-01T18:00:40Z

```
retro:proof no findings beyond the ledger's own. Read the four stage-1 diffs in full (d63a5bd TRK, 750007a VCS, d2e96a2 MFK, 11377cd FIX) and cross-checked every new operation, decoder, manifest key and fixture against container-swarm-cli.md before editing. All 13 declared checks (bb-checks.py, no args) exited 0 after the edit, byte-identical to the untouched-tree baseline on the three checks whose output is otherwise easy to drift (corpus, ref-links, template-drift). Fixture count corrected to eight, matching 'ls cli/testdata/swarm-fixtures/': core, status-attention, status-narrowed-scope, dispatch-rebase-conflict, dispatch-stale-base, dispatch-worktree, phase-epics, push-refused.
```

**`agent-plugins-o63s.6#3`** · dev-e6xt-b4 · 2026-09-01T18:28:02Z

```
retro:proof AC-SWARM-DISP-002.9 (held vs available) is proved against the dispatch-worktree fixture, purpose-built for exactly this scenario by the FIX bead (agent-plugins-o63s.4) — cli/testdata/swarm-fixtures/dispatch-worktree/fixture.toml's own description names the two beads and the one standing worktree. No new fixture was needed; the case-driven test cli/testdata/swarm-cases/ready-dispatch-set consumes it as-is and also proves .7, .10 and .17 in the same invocation, since the fixture's beads sit one level below the run epic (target), span two dispatch ids, and the run and target epics are both present in the tracker but absent from the answer.
```

**`agent-plugins-o63s.7#3`** · dev-e6xt-b5 · 2026-09-01T19:04:10Z

```
retro:proof AC-SWARM-DISP-002.15 (not established vs drained) verified via a real e2e case against the core fixture: TestReadyAgentNeverGivenAnythingIsNotReportedAsDrained asks for --agent dev-never-given (a dispatch id no fixture bead ever names) and asserts exit 1 with established=false. Command + output: go test ./swarm/e2e/... -run TestReadyAgentNeverGivenAnythingIsNotReportedAsDrained -v -> PASS.
```

**`agent-plugins-o63s.10#3`** · dev-e6xt-b6 · 2026-09-01T19:14:29Z

```
retro:proof Measured a pre-existing dry-run defect in the sprint-branch half (BR, agent-plugins-o63s.9), not introduced by this bead: `swarm dispatch open --run <fresh-run-with-no-recorded-branch> --dry-run --json` against the core e2e fixture exits status 2 with summary `plan: step "manifest-write-write" produced no value (refused)`, instead of the dry-run guarantee the command's own --dry-run help text states. Root cause: openSprintBranch's document(writeSteps, writeReport) call reads a StepRefused write's Value, which plan.Steps.ValueFrom (cli/swarm/plan/plan.go) always errors on since a dry-run-refused write's record.Value is never set (cli/swarm/plan/run.go's interpret(), ModeDryRun branch). This bead's own worktree-half facts avoid the same trap via a dry-run-safe fallback (invocationOf in dispatch.go, falling back to the step's own literal Args when a record's Invocation is empty) rather than depending on ValueFrom for anything the step's own literals already carry. Filed as agent-plugins-2ekn (discovered-from this bead) rather than fixed here — it is BR's delta's code, and a proper fix touches how document()/ValueFrom handle a refused write generally, which is broader than this bead's scope.
```

**`agent-plugins-o63s.8#3`** · dev-e6xt-b5 · 2026-09-01T19:16:03Z

```
retro:proof AC-SWARM-DISP-002.4, .5 and .6 (declared unit) are tested directly against applyBreaker with a minimal fake (only the manifest show/update calls) rather than through the full command, since reportReady's bare path also runs status.Collect's full worktree/completion machinery — unrelated to the breaker and expensive to fake. AC-SWARM-DISP-002.11 (declared e2e, MFK participates) is tested against the real core fixture: swarm manifest set writes consecutive_dispatch_failures=2, then a bare swarm ready --json (no --completed) reads 2 back rather than 0. Command + output: go test ./swarm/cli/... -run TestBreaker -v -> 6/6 PASS; go test ./swarm/e2e/... -run TestReadyBreaker -v -> 2/2 PASS.
```

**`agent-plugins-o63s.13#4`** · dev-e6xt-b7 · 2026-09-01T20:41:54Z

```
retro:proof `git branch -d` after a rebase fails when the branch has an upstream — caught by the e2e, fixed in the fixture helper, not in the command
evidence: `go test ./swarm/e2e/ -run TestDispatchAcceptRebases` failed with exit 1 and the report line:
  bb/dev-fixture-2 was concurrent against claude/bead-bonanza-fixture, rebased onto d54d47f... and fast-forwarded into it; ... bb/dev-fixture-2 could not be deleted: plan: step "delete-branch": git: git exited 1: If you are sure you want to delete it, run "git branch -D bb/dev-fixture-2"
Cause: `git branch -d` checks the branch against its UPSTREAM when one is set, and against HEAD only when none is. The e2e helper standDispatchBranch created the local branch with `git branch bb/dev-fixture-2 origin/bb/dev-fixture-2`, which sets an upstream by default; after the rebase the new head is not in origin/bb/dev-fixture-2, so -d refused. A real run cuts the branch with `git worktree add -b <branch> <path> <sha>` (openDispatchWorktree, cli/swarm/cli/dispatch.go), which sets NO upstream — so the world was wrong, not the command.
Fix: `--no-track` in standDispatchBranch, with the measurement recorded in its doc comment. Re-run: ok, 9.266s. The forcing form (-D) was never an option — vcs.VersionControl.DeleteBranch documents "-d rather than -D: -D would take unmerged commits with it".
The command still reports a refused deletion honestly rather than retrying it: attention, branch_deleted=false, and a resolution naming `git log <branch> --not HEAD`.
```

**`agent-plugins-o63s.16#5`** · dev-e6xt-b9 · 2026-09-01T21:05:02Z

```
retro:proof the seventh authorization-gate row needs no schema change — measured before shipping it into sprint-scope's standing_auth example.
evidence: `swarm manifest set --run agent-plugins-e6xt --dry-run --key standing_auth --value '{...six rows...,"dispatch_failure_threshold":{"value":3,"source":"default"}}'` -> exit 0, 'would write standing_auth as one merge'. `swarm manifest validate --run agent-plugins-e6xt` -> exit 0 (this run's live manifest already carries the seventh row). schemas/swarm/manifest.strata's StandingAuth declares six required fields and the generated UnmarshalJSON (cli/swarm/swarmmodel/swarmmodel.go:223) rejects a MISSING required field, not an extra one.
```

**`agent-plugins-o63s.15#3`** · dev-e6xt-b8 · 2026-09-01T21:07:44Z

```
retro:proof Verified the command tree, the no-run-scope exception, and every VCS interface addition by reading the actual landed code rather than the deltas' descriptions, per the bead's Design note. Confirmed: root.go's register() wires phase/ready/dispatch/branch; branch.go's branchLandingCmd builds both verify-landed and push off one implementation with an explicit 'AC-SWARM-DISP-005.6: no run-scope refusal here' comment; vcs.go's VersionControl interface carries Root, CurrentBranch, Checkout, Divergence(prefix, branch, remote) and vcs.Rejection; git/decode.go's pushDecoder/rejection functions implement the real push-rejection reading (the ! [ and remote: lines, not the tail of stderr); status.go exports BeadIDFromCloseSubject and accept_settle.go/accept.go both call it. No stage-2 tracker.go changes existed to fold (git diff --stat across the nine commits for cli/swarm/tracker/ is empty), so the Tracker and Beads Tracker sections were left as FOLD-1 left them.
```

**`agent-plugins-o63s.17#3`** · dev-e6xt-b9 · 2026-09-01T21:12:40Z

```
retro:proof measured swarm ready --agent against this very dispatch before writing the prose that instructs a developer to use it, rather than restating the help text.
evidence: 'swarm ready --run agent-plugins-e6xt --agent dev-e6xt-b9' -> exit 0, 'established: true / queue: - id: agent-plugins-o63s.17, state: in_progress / queue_count: 1 / drained: false'. The 'a blocked bead counts as drained' claim now in supervision.md is from the code, not from the run: cli/swarm/tracker/queue.go:57-59 builds the queue from t.Ready(...) plus t.Holding(...), and cli/swarm/tracker/beads/beads.go:144-146 shows Holding is 'bd list --assignee <agent> --status in_progress' — bd ready excludes blocked and --status in_progress does not match it, so neither half sees a blocked bead. 'swarm branch verify-landed --branch <sprint> --explain' -> exit 0, listing 'git fetch origin' before the two comparisons, which is the fetch-first property the recovery prose now claims.
```

**`agent-plugins-o63s.18#4`** · dev-e6xt-b10 · 2026-09-01T21:24:47Z

```
retro:proof Documentation-only change; no targeted test exists beyond the corpus and reference-link checks, which is what validated it.
evidence: git diff --stat -> 'specs/architecture/module-bead-bonanza.md | 62 ++++-- 1 file changed, 52 insertions(+), 10 deletions(-)'. Nothing outside specs/ was touched, so no Go or script test covers this diff. Full declared suite run in the foreground: 'python3 plugin/skills/bead-bonanza/scripts/bb-checks.py' -> checks-exit=0, all 13 checks, with 'corpus: checked 68 documents / no findings' and 'ref-links: checked 73 reference documents, 34 links, 0 errors' covering the new [[container-swarm-cli]] citation in Change Guidance. Identical to the run's checks_baseline (exit 0, all 13 passed).
```

**`agent-plugins-o63s.19#5`** · dev-e6xt-b10 · 2026-09-01T21:38:11Z

```
retro:proof The DoD was verified against the tree by measurement, not against this epic's issue statuses, as the item requires. Recording the commands so the next auditor need not re-derive them.
evidence: Command surface — 'swarm branch/dispatch/manifest/phase --help' enumerate push|verify-landed, accept|open, get|require|set|validate, close; every one of the 13 commands named in container-swarm-cli.md and module-bead-bonanza.md exists, and 'swarm ready --help' carries both --agent and --threshold (default 3), which the module blueprint's Change Guidance now names. Manifest key — schemas/swarm/manifest.strata:197 'consecutive_dispatch_failures: Int?', present in cli/core/schema/swarm-model.schema.json and cli/swarm/swarmmodel/swarmmodel.go, and in cli/testdata/swarm-fixtures/core/tracker.jsonl. Gate row — sf.toml:165 documents dispatch_failure_threshold and plugin/skills/sprint-scope/SKILL.md:220 is the seventh row of a table running 214-220. Contract — 'git log 2490026..HEAD -- specs/contracts/contract-swarm-json.md' is empty, and the schema change (d2e96a2) added one optional field with nothing removed, renamed or retyped, so no version raise was owed. Retired mechanisms — grep over plugin/ and scripts/ for BB_WORKTREE_BASEREF, 'exclude-label human' and 'epic close-eligible' returns nothing. Checks — 'python3 plugin/skills/bead-bonanza/scripts/bb-checks.py' exit 0, all 13, identical to the run's checks_baseline.
```

**`agent-plugins-e6xt.8#1`** · dev-e6xt-b11 · 2026-09-01T22:37:35Z

```
retro:proof the defect was observed failing before the fix, in the new test, for the stated reason
evidence: `go test ./swarm/cli/ -run TestEveryRealResultNamesTheInvocationThatProducedIt` on the unmodified tree failed 9 of 14 subtests, each with the same three lines, e.g. for `swarm dispatch accept` after it really merged:
    stamp_test.go:263: command = "", want "swarm dispatch accept"
    stamp_test.go:266: run = "", want "run-1"
    stamp_test.go:269: mode = "explain", want "execute" — the interpreter that actually ran
The five failing paths were exactly `swarm phase close`, `swarm ready` (both the whole-run and the --agent arm), `swarm dispatch open`, `swarm dispatch accept` (both the merged and the not-drained results), `swarm branch verify-landed` (both with and without --run) and `swarm branch push`. `swarm status` and the four `swarm manifest` subcommands passed unchanged, which is the contrast the bead cites.
After the fix, measured on a binary built from the worktree (`go build -o .../swarm ./swarm`):
    $ swarm ready --run agent-plugins-e6xt --dry-run --full-result --json
    {
      "contract_version": "2.0",
      "command": "swarm ready",
      "run": "agent-plugins-e6xt",
      "mode": "dry-run",
      "outcome": "clear",
      "status": 0,
where the bead measured "command": "", "run": "", "mode": "explain".
```

**`agent-plugins-e6xt.7#1`** · dev-e6xt-b11 · 2026-09-01T22:45:16Z

```
retro:proof the write was observed happening under ModeDryRun before the fix, and refused after it — measured twice, once at unit level and once against this live run
evidence: BEFORE, `go test ./swarm/cli/ -run TestTheBreakerWritesNothingInADryRun` on the unmodified tree:
    ready_test.go:275: a dry run wrote: [bd update run-1 --metadata {"consecutive_dispatch_failures":2} --json]
    ready_test.go:283: the count is reported as established at &{count <nil> 0xc00059a500 <nil> <nil> <nil>}; a refused write leaves it unknown, not at the value the write would have set
NEGATIVE CONTROL, same commit: `TestTheBreakerStillWritesWhenItReallyRuns` (ModeExecute) PASSED before and after, so the assertion above is about the mode reaching the runner and not about the breaker having been switched off.
AFTER, against the live run epic with a binary built from the worktree:
    $ swarm manifest get --run agent-plugins-e6xt --key consecutive_dispatch_failures --json
    { "consecutive_dispatch_failures": { "ok": false, "value": null, "issues": ["consecutive_dispatch_failures is not set. ..."] } }   (exit 1)
    $ swarm ready --run agent-plugins-e6xt --dry-run --completed 0 --full-result --json     (exit 0)
      "mode": "dry-run",
      "name": "consecutive_dispatch_failures",
      "reason": "the write was refused: dry run: the step writes. The count stands at 0 and would have become 1 (threshold 3)",
      ... { "step": "advance-failure-count-write", "command": "", "status": "refused", "reason": "dry run: the step writes" }
    $ swarm manifest get --run agent-plugins-e6xt --key consecutive_dispatch_failures --json
    ... still not set (exit 1)
So the count is unchanged after a dry run that reported a return, the refused step reaches the full result with the runner own reason, and the fact is unestablished rather than carrying the post-write value.
```

**`agent-plugins-e6xt.9#1`** · dev-e6xt-b11 · 2026-09-01T22:48:57Z

```
retro:proof the false landed:true was observed before the fix and is gone after it, at unit level and against a real repository with a genuinely stale tracking ref
evidence: BEFORE, `go test ./swarm/cli/ -run TestADryRunDoesNotReportLandedFromAnUnfetchedComparison` on the unmodified tree:
    branch_test.go:289: verify-landed --dry-run exited 0, want 1 (attention)
      "mode": "dry-run",
      "outcome": "clear",
      "status": 0,
      "summary": "sprint/run-1 has landed: every commit on it is on origin/sprint/run-1",
NEGATIVE CONTROL, same commit: TestAFetchThatRanStillProducesTheOrdinaryVerdict (execute mode, same three recordings) PASSED before and after — so the guard fires on the fetch not having run, not on the command having stopped reporting.
AFTER, the bead own repro built with real git (bare remote, two clones, the second force-pushing origin/feat back one commit, the first never fetching):
    real origin/feat  = e76581a
    local feat        = b86dae5
    local origin/feat = b86dae5      (stale)
    $ swarm branch verify-landed --branch feat --remote origin --dry-run --full-result --json
    DRYRUN_EXIT=1
      "outcome": "attention",
      "summary": "whether feat has landed on origin was not established: origin was not fetched",
      "resolution": "run it again without --dry-run, or fetch first: swarm branch verify-landed --branch feat --remote origin",
      { "name": "landed", "established": false,
        "reason": "origin was not fetched, so the comparison against origin/feat would rest on a possibly stale local record of where origin points" }
      { "step": "fetch", "status": "refused", "reason": "dry run: the step writes" }
    $ swarm branch verify-landed --branch feat --remote origin --json
    EXECUTE_EXIT=1
      "issues": ["feat is 1 commit(s) ahead of origin/feat; that much has not landed. push it: ..."]
      "landed": { "ok": true, "value": false }, "unpushed": { "ok": true, "value": 1 }
Execute mode is byte-for-byte what the bead recorded for it, so the change is dry-run-specific as the Design required.
```

**`agent-plugins-e6xt.11#1`** · dev-e6xt-b11 · 2026-09-01T22:54:40Z

```
retro:proof the dropped --completed was observed and is now applied, measured at unit level and with the bead own --explain repro
evidence: BEFORE, `go test ./swarm/cli/ -run "TestTheAgentArm|TestExplainOfTheAgentArmNamesTheBreaker"` on the unmodified tree:
    ready_test.go:301: --agent --completed 0 never advanced the count: [[bd list --parent run-1 --all --limit 0 --json] [bd list --parent run-1.1 ...] [bd ready --parent run-1 --assignee dev-1 ...] [bd list --assignee dev-1 --status in_progress ...]]
    ready_test.go:307: the --agent arm reports no breaker count at all
    ready_test.go:338: --agent --completed 3 did not reset the count: (same four reads)
    ready_test.go:377: the --agent transcript does not name the breaker read
No `bd show` and no `bd update` anywhere in the recorded calls — the flag was accepted and discarded, as the bead measured.
AFTER, the bead own step 1 against the live run:
    $ swarm ready --run agent-plugins-e6xt --agent dev-e6xt-b1 --completed 0 --explain
    EXIT=0
    transcript:
      - "swarm ready --agent dev-e6xt-b1 reads the run whole descendant set, then joins two more queries against it and reads the circuit breaker count, ..."
      - "  one bd list --parent <bead> --all --limit 0 --json per level below agent-plugins-e6xt ..."
      - "  bd ready --parent agent-plugins-e6xt --assignee dev-e6xt-b1 --exclude-type epic --exclude-label human --limit 0 --json  (this dispatch ready half)"
      - "  bd list --assignee dev-e6xt-b1 --status in_progress --limit 0 --json  (this dispatch held half, ...)"
      - "  bd show agent-plugins-e6xt --json                              (the failure count; --completed also writes it back)"
The breaker line the bead reported as absent is now present, and it is the same line the no---agent transcript carries.
AND the dry-run composition e6xt.7 required:
    $ swarm ready --run agent-plugins-e6xt --agent dev-e6xt-b11 --completed 0 --dry-run --full-result --json     (exit 0)
      "mode": "dry-run",
      { "name": "consecutive_dispatch_failures", "established": false,
        "reason": "the write was refused: dry run: the step writes. The count stands at 0 and would have become 1 (threshold 3)" }
      beside the arm own facts: queue_count, drained
    $ swarm manifest get --run agent-plugins-e6xt --key consecutive_dispatch_failures --json
    ... still not set
So the --agent arm applies the breaker and its write obeys the mode.
```

**`agent-plugins-e6xt.10#1`** · dev-e6xt-b11 · 2026-09-01T23:02:39Z

```
retro:proof merged_sha was observed naming a commit not on the target, and now names one that is — on both arms, end to end against the fixtures
evidence: BEFORE, the new e2e test on the unmodified tree (`go test ./swarm/e2e/ -run TestDispatchAcceptReportsAMergedShaThatIsOnTheTarget -v`):
    === RUN   TestDispatchAcceptReportsAMergedShaThatIsOnTheTarget/concurrent
        dispatch_accept_test.go:375: merged_sha 016160f623edf1d515e2a292874cbbdd42ea72df is not on claude/bead-bonanza-fixture (which is at caaf268b25a00a4ec07a4ada83f252727d6dbb5b); the fact names a commit that did not land
    --- FAIL: .../concurrent (8.03s)
    --- PASS: .../serialized (7.82s)
016160f is the same pre-rebase branch sha the bead measured, so this is the bead defect and not a lookalike. The serialized arm passed before and after — the two arms disagreed, and the test proves the fix did not simply stop reporting.
AFTER, the whole accept e2e set:
    --- PASS: TestDispatchAcceptRefusesADispatchThatHasNotDrained (4.57s)
    --- PASS: TestDispatchAcceptRebasesAConcurrentBranchAndFastForwards (8.40s)
    --- PASS: TestDispatchAcceptBlocksTheBeadAndRecordsAFindingOnAConflict (7.27s)
    --- PASS: TestDispatchAcceptRechecksEveryBeadAfterTheMergeAndTheRemoval (8.65s)
    --- PASS: TestDispatchAcceptReportsAMergedShaThatIsOnTheTarget (17.38s)   [concurrent + serialized]
    ok  github.com/brettcallaghan/agent-plugins/cli/swarm/e2e 46.267s
The property asserted is `git merge-base --is-ancestor <merged_sha> claude/bead-bonanza-fixture` exit 0, which is the bead own acceptance criterion and the check that failed with exit 1 before.
```

**`agent-plugins-e6xt.12#1`** · dev-e6xt-b11 · 2026-09-01T23:08:32Z

```
retro:proof the asymmetry was observed before the fix and both halves now carry the same exclusions — at argv level and in the shipped binary explain output
evidence: BEFORE, the new argv test compiled against the old signature and the old rendering; the failing form was the bead own measurement, which I reproduced as a table case:
    {"what an agent holds, filtered", tr.Holding("s","run-1", ReadyFilter{Assignee:"dev-1", ExcludeKinds:["epic"], ExcludeLabels:["human"]}), "bd list --assignee dev-1 --status in_progress --exclude-type epic --exclude-label human --limit 0 --json"}
would not even compile: `cannot use tracker.ReadyFilter{...} as string value in argument to tr.Holding` — Holding took (name, scope, agent string) and could not be given a filter at all, which is the defect in its most literal form.
AFTER, the shipped binary own explain, the bead step 1:
    $ swarm ready --run agent-plugins-e6xt --agent dev-e6xt-b1 --explain
      "  bd ready --parent agent-plugins-e6xt --assignee dev-e6xt-b1 --exclude-type epic --exclude-label human --limit 0 --json  (this dispatch ready half)"
      "  bd list --assignee dev-e6xt-b1 --status in_progress --exclude-type epic --exclude-label human --limit 0 --json  (this dispatch held half, under the same exclusions as the ready half, narrowed to agent-plugins-e6xt)"
Both halves now carry --exclude-type epic and --exclude-label human.
THE FLAGS THEMSELVES, measured rather than inherited from the bead ("bd list accepts --exclude-type/--exclude-label"), on bd 1.2.2 (6c124203e) against this repository database:
    $ bd list --limit 0 --all --json | jq length                          -> 602
    $ bd list --limit 0 --all --json | jq "[.[]|select(.issue_type==\"epic\")]|length"  -> 40
    $ bd list --limit 0 --all --exclude-type epic --json | jq length      -> 562
    $ ... same, counting epics                                            -> 0
    $ bd list --limit 0 --all --label human --json | jq length            -> 1
    $ bd list --limit 0 --all --exclude-label human --json | jq length    -> 601   (0 carrying the label)
So bd really applies both exclusions on `bd list`, not only on `bd ready`.
NEGATIVE CONTROL: the whole swarm test tree passes (`go test ./swarm/...` exit 0), including the six accept_settle cases that failed loudly the moment the argv changed and the fixture had not — proof the new argv is what the code actually issues rather than only what a doc comment claims.
```

**`agent-plugins-e6xt.13#1`** · dev-e6xt-b12 · 2026-09-02T00:10:31Z

```
retro:proof the defect was reproduced and watched fail before the fix
evidence: three new tests written against the unfixed tree at 2622f6c, plus a
negative control that had to stay green.

command: cd cli && go test ./swarm/cli/ -run 'TestADryRunReportsAStopTheRecordedCountAlreadyEstablishes|TestADryRunBelowTheThresholdStaysClear|TestReadyDryRunWithCompletedExitsStopAtARecordedThreshold|TestTheAgentArmReportsAStopTheRecordedCountEstablishesInADryRun' -v

output (abridged to the verdicts and the failure reasons):
  === RUN   TestADryRunReportsAStopTheRecordedCountAlreadyEstablishes
      ready_test.go:512: stopped = false at a RECORDED count of 3 against a
        threshold of 3; the read established it and a refused write does not
        unestablish it
      ready_test.go:516: summary = "", want it to say the run is stopped
  --- FAIL: TestADryRunReportsAStopTheRecordedCountAlreadyEstablishes
  === RUN   TestADryRunBelowTheThresholdStaysClear
  --- PASS: TestADryRunBelowTheThresholdStaysClear
  === RUN   TestReadyDryRunWithCompletedExitsStopAtARecordedThreshold
      ready_test.go:586: exit status = 0, want 2 (the run is stopped)
  --- FAIL: TestReadyDryRunWithCompletedExitsStopAtARecordedThreshold
  === RUN   TestTheAgentArmReportsAStopTheRecordedCountEstablishesInADryRun
      ready_test.go:602: exit status = 0, want 2 (the run is stopped)
  --- FAIL: TestTheAgentArmReportsAStopTheRecordedCountEstablishesInADryRun
  FAIL

The bead's own count-fact evidence reproduced verbatim in the failing
--full-result envelope, which is what proves the failure is the reported one
and not a lookalike:
  "name": "consecutive_dispatch_failures",
  "established": false,
  "reason": "the write was refused: dry run: the step writes. The count stands
             at 3 and would have become 4 (threshold 3)"
  ... alongside "outcome": "clear", "status": 0.

The negative control is TestADryRunBelowTheThresholdStaysClear (recorded 2,
threshold 3): it PASSED before the fix and passes after, so the three failures
are the missing stop verdict rather than the stop test having been moved onto
the projection.

After the fix, same command: exit 0, all four pass.
```

**`agent-plugins-e6xt.16#1`** · dev-e6xt-b12 · 2026-09-02T00:14:38Z

```
retro:proof the defect was reproduced and watched fail before the fix, with a
negative control that stayed green
evidence: a new test declares a fake whose git rev-parse --verify
sprint/run-1^{commit} answers on its FIRST invocation (the case reading) and
fails on its SECOND (the post-fold merged-head resolve), so the fold is reached
with everything before it succeeding.

command: cd cli && go test ./swarm/cli/ -run 'TestAFailedPostFoldResolveStillReportsTheMergeAndStillSettles|TestAPostFoldResolveThatAnswersStillEstablishesTheMergedSHA' -v

output BEFORE the fix:
  === RUN   TestAFailedPostFoldResolveStillReportsTheMergeAndStillSettles
      accept_test.go:628: exit status = 2: the accept was abandoned after the
        fast-forward had already landed
          {
            "command": "swarm dispatch accept",
            "outcome": "stop",
            "status": 2,
            "summary": "\"sprint/run-1 after the fold\" could not be resolved:
                        plan: step \"merged-head\": git: git rev-parse exited
                        128: fatal: unable to read the repository",
            "facts": [],
            "invocations": [],
            "transcript": []
          }
  --- FAIL: TestAFailedPostFoldResolveStillReportsTheMergeAndStillSettles
  === RUN   TestAPostFoldResolveThatAnswersStillEstablishesTheMergedSHA
  --- PASS: TestAPostFoldResolveThatAnswersStillEstablishesTheMergedSHA
  FAIL

That envelope is the bead's claim in full: "facts": [] and "invocations": []
mean the merge that HAD landed left no trace a caller could read, and the stop
status means settleDispatch was never reached.

output AFTER the fix: exit 0, both PASS.

The negative control is TestAPostFoldResolveThatAnswersStillEstablishesTheMergedSHA
(same world, a resolve that answers): it passed before and after, so the fix is
the degraded report on a failed read rather than merged_sha having stopped
being established at all.

The settle half is asserted from the fake's own record — the post-merge
ancestry guard (git merge-base --is-ancestor bb/dev-1 sprint/run-1) and the
branch deletion (git branch -d bb/dev-1) both ran on the failing path.
```

**`agent-plugins-e6xt.17#1`** · dev-e6xt-b12 · 2026-09-02T00:18:18Z

```
retro:proof the defect was reproduced on BOTH arms and watched fail, with a
negative control that stayed green
evidence: four new tests. Three drive a breaker refusal after the arm's own
queries have already answered; the fourth is the control.

command: cd cli && go test ./swarm/cli/ -run 'TestABreakerFailureKeepsTheAgentArmsQueueAnswer|TestABreakerFailureKeepsTheWholeRunArmsReadySet|TestABreakerValueThatDoesNotDecodeKeepsTheQueueAnswer|TestABreakerThatReadsCleanlyStillAnswersClear'

output BEFORE the fix (three FAIL, the control silent because it passed):
  --- FAIL: TestABreakerFailureKeepsTheAgentArmsQueueAnswer
      ready_test.go:629: exit status = 2, want 1 (an answer with something unmeasured in it)
          { "command": "swarm ready", "outcome": "stop", "status": 2,
            "summary": "plan: step \"failure-count\": plan: step \"failure-count\":
                        beads: bd show exited 1: bd: the tracker is unavailable",
            "resolution": "", "facts": [], "invocations": [], "transcript": [] }
  --- FAIL: TestABreakerFailureKeepsTheWholeRunArmsReadySet
      ready_test.go:706: exit status = 2, want 1
          { ... "facts": [], "invocations": [] }
  --- FAIL: TestABreakerValueThatDoesNotDecodeKeepsTheQueueAnswer
      ready_test.go:742: exit status = 2, want 1
          { "summary": "consecutive_dispatch_failures does not decode: manifest:
                        reading key \"consecutive_dispatch_failures\": json: cannot
                        unmarshal string into Go value of type int",
            "facts": [], "invocations": [] }
  FAIL

"facts": [] and "invocations": [] on all three is the bead's claim exactly: the
established, queue, queue_count and drained facts (ready and ready_count on the
whole-run arm) were all taken from queries that SUCCEEDED, and every one was
discarded along with every record.

output AFTER the fix: exit 0, all four PASS.

The control is TestABreakerThatReadsCleanlyStillAnswersClear (a manifest that
reads and decodes): it passed before and after, so the three failures are the
discard, not the breaker having been removed from the path.

Both arms are covered because the bead's Note on symmetry says reportReady
discards the same way and always has; fixing one alone would manufacture the
asymmetry the bead is about.
```

**`agent-plugins-e6xt.14#1`** · dev-e6xt-b12 · 2026-09-02T00:22:28Z

```
retro:proof the defect was reproduced and watched fail, with two controls that
stayed green
evidence: three new tests. One drives the combination the acceptance names and
nothing covered (a dispatch id no bead establishes, on a run the breaker
already stops); two are controls.

command: cd cli && go test ./swarm/cli/ -run 'TestAStoppedRunDoesNotKeepTheSupersededCasesResolution|TestAStoppedRunOnAnEstablishedDispatchSaysOnlyThatItIsStopped|TestAnUnestablishedDispatchOnAClearRunKeepsItsOwnResolution'

output BEFORE the fix (the two controls passed and are therefore silent):
  --- FAIL: TestAStoppedRunDoesNotKeepTheSupersededCasesResolution
      ready_test.go:826: resolution = "assign a bead in run-1 to typo-1, or
        dispatch a different agent id" belongs to the !established case the
        stop replaced; no resolution un-stops a run
      ready_test.go:830: resolution = "assign a bead in run-1 to typo-1, or
        dispatch a different agent id"; a stopped run carries none
      ready_test.go:836: summary = "the run is stopped: 3 consecutive
        dispatch(es) returned with no bead completed (threshold 3)" drops the
        mistyped-dispatch signal entirely
      ready_test.go:839: summary = "the run is stopped: ..." does not name the
        dispatch id
  FAIL

That is the bead's measured evidence reproduced line for line: outcome stop,
the breaker's summary, and the !established case's resolution still attached to
it.

output AFTER the fix: exit 0, all three PASS. The summary now reads
"the run is stopped: 3 consecutive dispatch(es) returned with no bead completed
(threshold 3); typo-1 is not an established dispatch in run-1" with an empty
resolution.

The two controls are what keep it honest:
- TestAStoppedRunOnAnEstablishedDispatchSaysOnlyThatItIsStopped -- an
  ESTABLISHED dispatch on the same stopped run gets the breaker summary with no
  clause appended, so the clause is carried forward rather than appended
  unconditionally.
- TestAnUnestablishedDispatchOnAClearRunKeepsItsOwnResolution -- an
  unestablished dispatch on a run the breaker does NOT stop keeps its own
  attention outcome and its own resolution, so the change is scoped to the
  override.
Both passed before the fix and after it.
```

**`agent-plugins-e6xt.18#1`** · dev-e6xt-b12 · 2026-09-02T00:24:08Z

```
retro:proof a documentation-only bead carries no test, so the evidence is the
sweep and the verification the Design asked for

1. THE SPEC WAS VERIFIED, NOT RE-EDITED, as the Design directed.
  command: grep -n "both arms" specs/architecture/container-swarm-cli.md
  output:
    289:`swarm ready` also carries the run's circuit breaker, **on both arms**:
    352:**What landed is read from the target, after the fold, on both arms.**
  Already correct. The spec was left alone.

2. THE FALSE CLAIM IS GONE FROM THE TREE, swept after the edit for any other
copy of it:
  command: grep -rn "carries no circuit-breaker|apply only here|no circuit-breaker semantics" . --include=*.go --include=*.md
  output: no matches.

3. THE COMMENT WAS WRITTEN AGAINST THE CODE AS IT NOW STANDS, which is what
made this bead fifth in the batch rather than first. e6xt.13, e6xt.17 and
e6xt.14 all landed before it and all three changed what this comment describes,
so the sentence names the shared function and points at the one place the arms
genuinely differ instead of restating either arm's behaviour:

  // breaker and threshold are the CB delta's: whether this invocation is
  // reporting a dispatch return, and the count that stops the run. Both
  // arms carry them, through the one [applyBreaker] — --agent narrows
  // which beads are reported and changes nothing about the dispatch return
  // the same invocation reported. Where the two arms differ is in what a
  // stop supersedes, and that is stated on [reportAgentQueue], which is
  // the arm it is a property of.

The last sentence is new relative to the Design, which was written before
e6xt.14 established that the arms DO differ in one respect (reportAgentQueue
clears a superseded resolution; reportReady has none to clear). Saying "both
arms carry the breaker" and stopping there would have been true of the breaker
and misleading about the override three functions later. The Design's
instruction not to restate reportAgentQueue's own sentence is honoured: this
one points at it rather than copying it.

4. No behaviour change, so no test. The gate is the check set staying green:
  command: python3 plugin/skills/bead-bonanza/scripts/bb-checks.py
  exit: 0 — no findings beyond the run's baseline, which was also 0.
```

**`agent-plugins-e6xt.20#1`** · dev-e6xt-b12 · 2026-09-02T00:25:47Z

```
retro:proof the deadness verdict was re-measured against the code that ships,
not against the code the bead was filed on

The bead ordered this sixth so the grep would run after e6xt.16 rewrote the
post-fold merged_sha handling in the same file. It did.

BEFORE, on this batch's tree with e6xt.16 already committed (c475255):
  command: grep -rn "branchSHA" cli/swarm/
  output:
    cli/swarm/cli/accept.go:368:	branchSHA  string
    cli/swarm/cli/accept.go:426:	branchSHA, err := revision(report, branchStep, dispatchBranch)
    cli/swarm/cli/accept.go:452:		branchSHA:  branchSHA,
  Three lines, and no read — the same verdict the bead recorded at 2622f6c, so
  e6xt.16 neither revived the field nor added a consumer. Worth stating rather
  than assuming: e6xt.16's fix reaches for an invocation from the post-fold
  resolve, which is exactly the kind of change that could have reintroduced a
  pre-fold sha.

AFTER:
  command: grep -rn "branchSHA" cli/swarm/
  exit: 1, no output.

The retained check, per the acceptance:
  // Resolved for the check and not for the value: this refuses early
  // when the dispatch branch does not resolve at all, rather than
  // letting that failure surface later inside the fold.
  if _, err := revision(report, branchStep, dispatchBranch); err != nil {
      return mergeCase{}, report.Records, err
  }
The refusal behaviour is unchanged — same call, same steps, same error return;
only the value is discarded. The Design offered two shapes for that and left
the pick to whatever keeps the tooling clean; the "_" form was taken because
keeping an unused local would not compile in Go at all.

  command: cd cli && go vet ./swarm/...
  exit: 0

  command: python3 plugin/skills/bead-bonanza/scripts/bb-checks.py
  exit: 0 — no findings beyond the run's baseline, which was also 0.

m.branchInv was left alone: it is still read, citing the dispatch_branch fact.
```

**`agent-plugins-e6xt.19#1`** · dev-e6xt-b12 · 2026-09-02T00:29:05Z

```
retro:proof the test was strengthened FIRST and watched fail on the exact
assertion the bead says was missing

The bead's point is that the case was already covered and the assertion could
not distinguish the two verbs, so the order matters: the assertion was
tightened against the unfixed code before the code was touched.

The old assertion, in TestADryRunDoesNotReportLandedFromAnUnfetchedComparison
(cli/swarm/cli/branch_test.go), inside a loop over both verbs:
    if !strings.Contains(env.Resolution, "--dry-run") {

The new one adds both halves the Design requires:
    other := "push"
    if verb == "push" {
        other = "verify-landed"
    }
    if !strings.Contains(env.Resolution, "swarm branch "+verb) { ... }
    if strings.Contains(env.Resolution, "swarm branch "+other) { ... }

command: cd cli && go test ./swarm/cli/ -run 'TestADryRunDoesNotReportLandedFromAnUnfetchedComparison' -v

output BEFORE the fix (test strengthened, code untouched):
  === RUN   TestADryRunDoesNotReportLandedFromAnUnfetchedComparison
      branch_test.go:327: push --dry-run's resolution does not name the command
        the caller ran: "run it again without --dry-run, or fetch first: swarm
        branch verify-landed --branch sprint/run-1 --remote origin"
      branch_test.go:331: push --dry-run's resolution sends the caller to
        `swarm branch verify-landed`, which is a different command: "run it
        again without --dry-run, or fetch first: swarm branch verify-landed
        --branch sprint/run-1 --remote origin"
  --- FAIL: TestADryRunDoesNotReportLandedFromAnUnfetchedComparison
  FAIL

The verify-landed iteration passed throughout, which is the control built into
the loop: the shared string happened to name that verb correctly, so only the
push iteration failed.

output AFTER the fix: exit 0.

Sweep for other hardcoded instances of the same string, after the change:
  command: grep -rn "swarm branch verify-landed" --include=*.go cli/ | grep -v _test.go
  output:
    cli/swarm/cli/branch.go:1:// The `swarm branch verify-landed` and `swarm branch push` commands: one
    cli/swarm/cli/branch.go:3:// of them. Per the "swarm branch verify-landed and swarm branch push"
  Both are the file header naming the two commands, which is correct prose. No
  resolution hardcodes the verb any more.
```

### `environment` — 10

**`agent-plugins-o63s.3#5`** · dev-e6xt-b3 · 2026-09-01T17:09:21Z

```
retro:environment the installed strata is current, so the AGENTS.md stale-binary trap did not fire — measured, not assumed.
evidence: `cat ~/.cache/strata-cli/built-commit` → ce18d66f188adfb0a6dce54ff584f25db6227860. `scripts/generate-schemas.sh --check` on the UNTOUCHED tree, before any edit → exit 0, 'generated artifacts are up to date'. So the committed artifacts and this binary agree, and the regenerated output is trustworthy rather than the empty-properties corruption AGENTS.md warns about. After the edit, generate-schemas.sh rewrote exactly three artifacts (cli/core/schema/swarm-model.schema.json, cli/swarm/swarmmodel/swarmmodel.go, plugin/skills/bead-bonanza/references/generated/manifest.md), all additive: 75 insertions, 0 deletions across the four changed files including the .strata source.
```

**`agent-plugins-o63s.2#3`** · dev-e6xt-b2 · 2026-09-01T17:40:50Z

```
retro:environment A sibling dispatch's `bd dolt pull` serializes every bd call in the shared container, and the project's own script-self-tests check calls bd — so this run's gate blocked for ~22 minutes behind another agent's tracker sync.
evidence: Dispatches in this run share one container. At 17:16 both `bd create` (mine) and `bb-checks.py` were queued behind a sibling's pull:

  ps -eo pid=,args=  ->
    27830 bd dolt pull          (started 17:10, dev-e6xt-b3)
    10282 bd where
    24764 bd where
    32522 bd create Decode ...

  ps -o pid=,lstart=,etime=,time=,args= -p 27830  ->
    27830 Tue Sep  1 17:10:13 2026   18:06  00:02:00  bd dolt pull

The pull was working, not hung (CPU climbing 1:49 -> 2:00 over two minutes of wall clock). It exited at 17:38:13, and both my `bd create` and the checks completed within seconds of that.

The part that matters for the run's own tooling: `bb-checks.py`'s script-self-tests check stalled mid-run at `pre-flight-check --self-test`, because
  plugin/skills/bead-bonanza/scripts/pre-flight-check.sh:608
    BD_DB="$(bd where 2>/dev/null | sed -n 's/^ *database: *//p' | head -1)"
runs bd unconditionally. So a project check — not just bead ceremony — can block indefinitely on another dispatch's tracker operation. There is no timeout on it and nothing in the output says why it is stopped; from outside it is indistinguishable from a hung check.

Two traps for whoever reads this next:
  - `pgrep -f "bd dolt pull"` is NOT a usable wait predicate here. Every sibling shell waiting on the pull carries that string in its own command line, so the pattern matches the waiters and the until-loop never terminates. Two sibling shells in this container were spinning on exactly that (pids 4645 and 9071). Wait on the pid (`ps -p <pid>`) or on the output file instead.
  - The gate's wall-clock time is not its cost. Re-running bb-checks.py on the final tree after the lock cleared took under a minute; the first run's ~22 minutes was all lock.
```

**`agent-plugins-o63s.1#5`** · dev-e6xt-b1 · 2026-09-01T17:41:19Z

```
retro:environment A sibling dispatch running `bd dolt pull` holds an exclusive lock on the Dolt database, and every other `bd` invocation in the run blocks on it until it finishes — no error, no timeout of its own.

Observed while measuring subcommand shapes for the decoders. `bd list --status in_progress --limit 0 --json` ran past 120s and was still running; `ps aux` showed:

    root 27830 11.2 0.9 2797936 156412 ? Sl 17:10 0:36 bd dolt pull        (BEADS_ACTOR=dev-e6xt-b3)
    root 29202  0.1 0.5 1919548  83972 ? Sl 17:10 0:00 bd list --status in_progress --limit 0 --json

and a second invocation issued while the pull held the lock returned, on stderr:

    Error: failed to open database: embeddeddolt: init schema: embeddeddolt: open db: failed to load database "beads": the database is locked by another dolt process

So the failure has two faces depending on timing: a hang (the first invocation waits) and an explicit lock error (a later one). Neither is recoverable by retrying immediately.

For the next agent: this is not a broken tracker and not worth diagnosing again. `bd dolt pull` took roughly five minutes here. Wrap tracker reads in `timeout`, and where a read must succeed, wait for `pgrep -f 'bd dolt pull'` to go quiet rather than retrying in a loop. Concurrent dispatches on one shared Dolt database make this reachable in any run with more than one agent.
```

**`agent-plugins-o63s.2#5`** · dev-e6xt-b2 · 2026-09-01T17:43:00Z

```
retro:environment The tracker's Dolt remote cannot be pushed from this container: the pull that would make the push a fast forward aborts on a merge conflict that needs operator resolution. This bead's closure is in the LOCAL database only.
evidence: The close ceremony committed, then reported its own push failure:

  bb-close-bead.sh agent-plugins-o63s.2 2 "VCS: ..."
  -> exit 0
     ✓ Updated issue: agent-plugins-o63s.2
     bb-close-bead: WARNING — commit landed but `bd dolt push` failed. The closure is
     in the local database only; push it before the container is reclaimed.
     6f0f1f5

Retried by hand, twice, with no other bd process running (`ps -eo pid=,args= | grep "[b]d dolt"` empty before the first):

  bd dolt push
  -> exit 1
     Error: push to origin/main: Error 1105: To git+https://github.com/brettcallaghan/agent-plugins.git
      ! [rejected]            main -> main (non-fast-forward)
     hint: Updates were rejected because the tip of your current branch is behind
     hint: its remote counterpart. Integrate the remote changes (e.g. 'dolt pull ...')
     hint: before pushing again.

The documented remedy does not clear it:

  bd dolt pull
  -> exit 1
     Error: merge origin/main: merge conflicts in issues require operator resolution;
            merge aborted and working set restored

`bd dolt --help` exposes no conflict-resolution subcommand — only start/stop/status, show/set/test, commit/push/pull and remote add/list/remove. So this is not recoverable from inside a dispatch, and I did not attempt it: resolving tracker merge conflicts blind would rewrite bead states belonging to the sibling dispatches sharing this container.

State left behind, all verified rather than assumed:
  - `bd show agent-plugins-o63s.2 --json` -> "status": "closed", "assignee": "dev-e6xt-b2",
    "closed_at": "2026-09-01T17:41:10Z". The aborted pull restored the working set without
    losing the closure.
  - `git log --oneline -1` -> 6f0f1f5 agent-plugins-o63s.2 P2 VCS: Version Control and Git
    Version Control — the dispatch loop's writes
  - `git status --porcelain` -> empty.

So the code and the closure are both in the commit on the dispatch branch; only the tracker's remote copy is behind. Same container also showed a sibling's `bd dolt pull` running 28 minutes earlier (17:10:13 to 17:38:13), which is the likely other half of the race.
```

**`agent-plugins-o63s.1#7`** · dev-e6xt-b1 · 2026-09-01T17:43:16Z

```
retro:environment `bd dolt push` is BLOCKED for this run by a Dolt merge conflict that needs operator resolution. The closure of this bead is in the LOCAL database and in commit d63a5bd only.

Sequence, all after the commit landed:

    $ bd dolt push
    Error: push to origin/main: Error 1105: To git+https://github.com/brettcallaghan/agent-plugins.git
     ! [rejected]            main -> main (non-fast-forward)
    hint: Updates were rejected because the tip of your current branch is behind
    hint: its remote counterpart. Integrate the remote changes (e.g.
    hint: 'dolt pull ...') before pushing again.
    exit 1

    $ bd dolt pull
    Error: merge origin/main: merge conflicts in issues require operator resolution; merge aborted and working set restored
    exit 1

    $ bd dolt push          # retried after the aborted pull
    (same non-fast-forward rejection)
    exit 1

The aborted pull restored the working set cleanly — verified afterwards:

    bd show agent-plugins-o63s.1 --json | jq -r '.[0].status'                    -> closed
    bd comments agent-plugins-o63s.1 --json | jq '[.[]|select(.text|startswith("retro:"))] | length' -> 6

So nothing is lost locally, but nothing is durable either: the remote does not carry this closure, its six findings, or `agent-plugins-jou4` (the bead filed from this one).

NOT resolved here, deliberately. Resolving it means merging another dispatch's concurrent issue writes, which is outside this bead and outside what a dispatch subagent may do. Concurrent dispatches writing the same Dolt database is the cause; the run's coordinator is the one positioned to resolve the conflict and push.

The container being reclaimed before that happens is what loses the work.
```

**`agent-plugins-o63s.4#6`** · dev-e6xt-b3 · 2026-09-01T17:47:54Z

```
retro:environment the tracker cannot be pushed from this session: bd dolt pull aborts on merge conflicts in issues, so both beads in this batch are closed locally and unsynced.
evidence: after MFK's close, 'bd dolt push' exited 1 with 'push to origin/main: Error 1105: ! [rejected] main -> main (non-fast-forward)' — sibling dispatches in this run had pushed first. 'bd dolt pull' then ran for over 17 minutes holding an exclusive lock ('failed to load database beads: the database is locked by another dolt process' on any concurrent bd call, measured with a 60s timeout) and ended with 'Error: merge origin/main: merge conflicts in issues require operator resolution; merge aborted and working set restored'. The local database survived intact — 'bd show agent-plugins-o63s.3' reads CLOSED afterwards. So the closures are durable in git (each bead's close commit) and in the local Dolt database, and are NOT on the remote. The conflict needs an operator; nothing this dispatch can run resolves it. While the pull held the lock no bd call of any kind was possible, which is worth knowing: a batch that hits this stalls completely rather than degrading.
```

**`agent-plugins-o63s.5#1`** · dev-e6xt-b4 · 2026-09-01T18:17:25Z

```
retro:environment bd's dependency rule refuses a task 'blocks' an epic directly. Measured on this bd: `bd dep add <task> <epic> --type blocks --json` -> {"error":"tasks can only block other tasks, not epics"}; the reverse (`bd dep add <epic> <task> --type blocks`) -> {"error":"epics can only block other epics, not tasks"}. Only epic-blocks-epic is accepted, and bd's ready query walks that edge transitively: with fixture-run.phase-b blocked-by fixture-run.phase-a (both epics), fixture-run.phase-b.1 (phase-b's own open child task) does not read as ready until phase-a closes — confirmed via a real bd in cli/swarm/e2e/phase_close_test.go's TestPhaseCloseClosesOnlyItsOwnRunAndNamesWhatItReleased. This is the only mechanism by which closing one phase epic releases another phase's beads as ready; a future bead wiring cross-phase dependencies should use epic-to-epic blocks, not a task blocking on an epic.
```

**`agent-plugins-o63s.11#3`** · dev-e6xt-b4 · 2026-09-01T19:48:36Z

```
retro:environment `bd init` COMMITS its own scaffolding onto the checked-out branch, so every restored end-to-end world had its sprint branch pointing at a `bd init` commit rather than at the sha its fixture stamp documents, and its index carrying the same files. Nothing caught it before, because no case had folded a branch. Fixed in the harness with a mixed reset back to the head read before `bd init`.
evidence: measured in a clone of `dispatch-rebase-conflict/repository.bundle` with HOME pinned to a fresh dir: `git log --format=%s -n 1 HEAD` -> "fixture-run.4 P1 the fourth bead of the run" after clone+checkout; -> "bd init: initialize beads issue tracking" after `bd init --prefix fixture`; unchanged by `bd import`. `git rebase <target-sha> bb/dev-fixture-2` there -> exit 1, "error: cannot rebase: Your index contains uncommitted changes. error: Please commit or stash them." After `git reset --quiet 078f2b6`: head reads "fixture-run.4 P1 ..." again, `git status --porcelain --untracked-files=no` is empty, and `bd show fixture-run.3 --json` still answers.
```

**`agent-plugins-o63s.11#4`** · dev-e6xt-b4 · 2026-09-01T19:48:44Z

```
retro:environment the end-to-end harness pinned HOME into the world but no git identity, so a case that writes a commit could not. Added GIT_AUTHOR_/GIT_COMMITTER_ NAME and EMAIL to `world.env()`.
evidence: measured in a probe world with HOME set to an empty dir: `git rebase <sha> bb/dev-fixture-2` -> "fatal: unable to auto-detect email address (got root@vm.(none))". No prior case wrote a commit, so `git worktree add` and `git branch` — which need no identity — never surfaced it.
```

**`agent-plugins-o63s.16#1`** · dev-e6xt-b9 · 2026-09-01T21:04:35Z

```
retro:environment the `swarm` binary on PATH predates every command this item adopts, so a coordinator following the new prose would get 'unknown command'. Had to build from source to measure them.
evidence: /usr/local/bin/swarm (mtime Sep  1 07:22) `swarm --help` lists only: manifest, status. `cd cli && go build -o $SCRATCH/swarm ./swarm` then `$SCRATCH/swarm --help` lists: branch, dispatch, manifest, phase, ready, status. cli/swarm/cli/root.go register() adds phaseCmd, readyCmd, dispatchCmd, branchCmd.
```


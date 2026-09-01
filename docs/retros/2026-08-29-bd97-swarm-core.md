# Run `agent-plugins-bd97` — Swarm Core

The one human-facing document for this run: what it did, every finding it
captured, what it could not verify, and the eight decisions it asks you for.

---

## Read this first

Two things outrank the list of seventeen review findings below, and both are
easy to lose in it.

### 1. The skill family this repository vends is currently broken for every project except this one

This run made `swarm` a hard requirement of the vended skills and never gave a
consumer a way to get it. Two review findings together:

- **`agent-plugins-bd97.4.8`** — the vended `bead-bonanza` skill hard-requires
  the `swarm` binary. Pre-flight **blocks** without it, and the only remedy it
  prints is `(cd cli && go build -o /usr/local/bin/swarm ./swarm)` — a path that
  exists in this repository and in no consumer of it.

  Re-measured while writing this record, rather than inherited from the review:
  `pre-flight-check.sh:652` reads
  `block "swarm is not on PATH — swarm manifest and swarm status are required
  steps: (cd cli && go build -o /usr/local/bin/swarm ./swarm)"`, and `block` is
  the hard-stop path — with two further `block` calls at `:643` and `:647` for
  the stale-binary cases. `grep -ci swarm
  plugin/skills/bead-bonanza/references/project-setup.md` returns **0**: the
  document that tells a consumer how to set the project up does not mention the
  binary its skills now require.
- **`agent-plugins-bd97.4.9`** — the capability probe tests only
  `swarm status --help`, so a binary behind on manifest keys passes it, and
  `session-start`'s rebuild guard then preserves that stale binary.

`.4.9` was verified independently by the coordinator and was **live in this
very container**: `/usr/local/bin/swarm` (built 04:46, before `ih18.21`
merged) answers `swarm manifest get --key trailers` with exit 3
`not a declared manifest key` — the exact bug `ih18.21` fixed — while a build
from `cli/swarm` at HEAD answers exit 1 `unestablished: the key is not
written`.

Neither is admitted. `admit_review_findings=false` was resolved with a human
at the authorization gate, and the coordinator held to it rather than
overriding its own authorization because the findings turned out severe. That
was the right call and it is why these two are at the top of this document
instead of in a dispatch queue.

**Scope of the claim, stated honestly:** "broken for every project except this
one" is the reviewer's inference from the two findings above plus
`project-setup.md`'s silence. No consumer project was actually run against
these skills — there is no second project in this session to run them in. What
is measured is the pre-flight block, the remedy text, the probe's shape, and
the stale binary in this container. What is inferred is the consequence for a
consumer.

### 2. The run's dominant cost was its own collision mechanism

The plan serialized a 20-bead queue into **17 waves at max width 2, and 14 of
those 17 are width 1.** The run provisioned four dispatch agents for a graph
that admits roughly one at a time; `dev-bd97-4` and `dev-bd97-5` had no ready
bead by wave 11.

The cause is measured and it is not the dependency graph. Component-granular
collision avoidance — the two signals `partitioning.md` actually declares —
found 8 C4 pairs, **all of which were already dependency-ordered**, and
therefore prevented nothing. The real contention was a signal the reference
never names: `c4_docs`. 15 of 19 beads carry
`c4_docs=[specs/architecture/container-swarm-cli.md]` and every one of them
appends to that single blueprint. Serializing that cost five waves.

Both are filed: `agent-plugins-221f` (collision granularity) and
`agent-plugins-tvom` (emit the wave decomposition as an ASCII graph at plan
time, while the edges are still cheap to change).

The trade was taken deliberately and recorded, not stumbled into: dropping the
three edges would have restored waves 5–7 (17 waves at width 2 becoming ~12 at
width 3) at the cost of hand-merging the shared blueprint. The coordinator kept
the edges because a conflicting resolution on a shared roster or diagram is
silent damage. **The decision was sound; the cost is still the run's largest,
and the mechanism that forced the choice is the thing worth fixing.**

---

## The run

| | |
|---|---|
| Run epic | `agent-plugins-bd97` — Sprint: Swarm Core |
| Target | `agent-plugins-ih18` (Swarm Core) — a durable **feature** epic adopted as the run target |
| Branch | `claude/bead-bonanza-agent-plugins-ih18-2yjc8j` |
| Base → head | `3a57dc7` → `24f97e3` |
| Queue | 19 admitted at plan time → **24 admitted**, 1 excluded |
| Declared checks | 13 (`format`, `vet`, `build`, `test-go`, `skills`, `shared-refs`, `strata-docs`, `schemas`, `corpus`, `ref-links`, `template-drift`, `script-self-tests`, `skills-strict`) |
| Model lineup | light `haiku`, standard `sonnet`, advanced `opus`, frontier `fable` |
| Effort | `unrecorded` in the manifest |
| Review | 2 passes of a budget of 2; 17 findings, **0 admitted** |
| Dispatches | `dev-bd97-1`, `-2`, `-4`, `-5` (`-3` retired before dispatch during the tier-coherence re-partition) |

---

## Results

24 of 24 admitted beads closed, one commit each, all merged and pushed.

**Diff `3a57dc7..24f97e3`: 24 commits, 99 files, +17914 / −839.** Every bead
passed the five-signal acceptance check post-merge, and every worktree was
removed behind an ancestry guard.

| bead | P | commit | what landed |
|---|---|---|---|
| `ih18.1` | P1 | `d4ba587` | Stand up the `swarm` binary in the existing cli module |
| `ih18.2` | P1 | `195f186` | Process Runner: one argv invocation, separated streams |
| `ih18.3` | P1 | `4f2c820` | Plan Runner: steps as values, three interpreters over them |
| `ih18.4` | P1 | `96a9d63` | Command Surface: flags, four exit statuses, two output forms |
| `ih18.5` | P1 | `1ff852a` | Tracker interface and the Beads implementation |
| `ih18.6` | P1 | `4598e62` | Version Control interface and the git implementation |
| `ih18.7` | P1 | `6679e1d` | Strata schemas for the manifest key set and output envelope |
| `ih18.8` | P1 | `b41caaf` | Manifest read/write path with the store interface |
| `ih18.9` | P1 | `46bf165` | Manifest validation: the rules no single write can check |
| `ih18.10` | P1 | `d3ca962` | `swarm manifest`: get, set, require, validate |
| `ih18.11` | P1 | `a47079a` | End-to-end fixture: bundle, tracker export, case harness |
| `ih18.12` | P2 | `cbd9f4a` | Adopt the command contract; retire per-script exit-code prose |
| `ih18.13` | P1 | `ff34771` | Adopt `swarm manifest` into the skills, delete the hand method |
| `ih18.14` | P1 | `90cd22a` | Status Collector: take measurements, format nothing |
| `ih18.15` | P1 | `06b814b` | Status Reporter: report measurements, never a diagnosis |
| `ih18.16` | P1 | `cfdcb5d` | Adopt `swarm status`, retire the derived-health prose |
| `ih18.17` | P2 | `2f94f4a` | Flip `contract-swarm-json` from draft to active |
| `ih18.18` | P2 | `472635b` | Windows leg against the tracker install ADR-027 targets |
| `ih18.19` | P2 | `24f97e3` | Verify the DoD, retire the blueprint |
| `ih18.20` | P2 | `9527d46` | Reconcile ADR-025's Windows `.cmd`-shim open question |
| `ih18.21` | P1 | `9b67309` | `swarm manifest` serves the keys that have no stage |
| `ih18.23` | P2 | `9d446a3` | Worktree branches do not carry `bb/<dispatch-id>` |
| `ih18.24` | P2 | `b2b9214` | Re-home the readiness re-checks the retired mode carried |
| `ih18.25` | P2 | `b8368c0` | Nothing provisions or version-checks the `swarm` binary |

### The DoD verification, and what it did not tick

`ih18.19` passed **9 of 11** DoD items and recorded the other two as named
exceptions in the retired blueprint rather than ticking them — the Windows CI
leg (never run) and two re-scoped items. It also found and fixed **three false
claims in the two owning architecture blueprints, every one left behind by a
change made elsewhere in this same run.**

Its stated pattern is the most transferable thing the run produced:

> A DoD that enumerates the documents to fold a delta **into** does not
> enumerate the documents that **point at** them, and the pointer is where the
> stale copy survives.

---

## Scope

**Started at 19 admitted. Five beads were filed mid-run; four were admitted and
one deferred out.** The run finished its admitted set, so this is a run that
*grew* rather than narrowed — but one bead sat out and that must be said.

**Excluded — 1:**

- **`agent-plugins-ih18.22`** — deferred out. It would change
  `scripts/pre-flight-check.sh` to write the manifest through
  `swarm manifest set`. Correct as adoption, but it edits the coordinator's own
  pre-flight script while the run still depends on it, **and a resume re-runs
  that script.** It also raises a bootstrap question the delta does not settle:
  pre-flight runs before anything is built, so making it call the `swarm`
  binary makes the run's entry point depend on an artifact of the run. Deferred
  with `bd defer` rather than moved, because `ih18` is the durable feature epic
  and `.22` belongs to the feature.

**Consequence to carry forward:** `module-bead-bonanza`'s adoption is
**incomplete by decision** — the Environment Probe is not adopted — and the
retired blueprint says so rather than claiming full adoption.

**Moved: none.**

### One object, two definitions of done

The run adopted a durable feature epic (`ih18`) as its run target, and at the
review trigger the two semantics disagreed openly: the ready front was dry and
completion read 25 of 26, while `bd97.4` reported `blocked by agent-plugins-ih18
[open]` — because `ih18` cannot close while `.22` is deferred, and `.22` is
deferred *because this run excluded it*.

The coordinator advanced on the documented trigger ("when the ready front is
dry, before the retro") and **said so rather than letting it look automatic.**
Nothing was forced: no epic closed early, no bead moved out of the feature,
`ih18` stays honestly open. Filed as `agent-plugins-p39p`.

---

## What the run decided

34 `decision` findings — the largest single kind, and that is the signal worth
reading. **A `decision` finding marks a place a specification failed to answer
a question a bead needed answered**, and the agent had to answer it to proceed.
Thirty-four of them across 24 beads means the FRD and blueprints under-specified
this feature substantially, even though every bead passed the readiness gate.

The pattern is consistent: agents took the **conservative** reading, implemented
it, tested it, and named the rejected alternative. A representative sample —
all 34 are in the appendix:

- **`ih18.3#3`** — three resolution rules the delta does not state, each chosen
  conservatively, each with a test: a step publishes its decoded value only when
  it exited 0; a dry run that refuses a write *continues* rather than stopping;
  a step with no decoder yields trimmed stdout.
- **`ih18.2#3`** — a non-zero exit is **data** on the Result with a nil error;
  the error return is reserved for an invocation that produced no status at all.
  "Neither the delta nor ADR-025 says which of the two shapes to use, and the
  choice changes every call site above."
- **`ih18.15#1`** — the stall condition reads "no worktree stands" as no
  *dispatch* worktree, not the raw count. Measured: a real clone always
  registers at least the main checkout, so a literal `Standing==0` check could
  never fire outside a unit-test mock.
- **`ih18.6#3`** — the Version Control operations return a fragment while the
  Tracker's return a bare step. "The asymmetry is deliberate and is the one
  thing a reviewer should push back on if they disagree." *(No reviewer did —
  see the pass-1 coverage gap below.)*
- **`ih18.23#2`** — the coordinator's own worktree is excluded from attribution
  by structural comparison against the manifest's recorded branch, not by
  heuristic, because main's HEAD is very often itself a bead-close commit.

### Gaps, sequencing and staleness

- **`ih18.4#2`** — SURF must emit an envelope fixed by a `.strata` file the
  SCHEMA item **has not created yet, and no sequencing constraint puts SCHEMA
  before SURF.**
- **`ih18.3#6`** / **`ih18.4#1`** — the SURF delta's anchors name
  `cli/swarm/main.go` and `cli/swarm/cli/`; the SKEL delta had already put the
  surface at `cli/swarm/cmd/`. Resolved in the blueprint's favour by renaming.
- **`ih18.16#1`** — "The REP delta this one depends on is CLOSED in the tracker
  but its commit is not in my base tree." A tracker-closed dependency is not a
  merged one.
- **`ih18.24#1`** — the same failure again, and this one blocked a bead:
  `ih18.24`'s design assumes `bb-status.sh run` is already retired, but
  `ih18.16`'s retirement commit had not reached that dispatch's branch. Filed as
  `agent-plugins-xeli`.
- **`ih18.19#1`** — "The epic's own acceptance criterion counts work orders, so
  it went stale the first time the run admitted one." An AC that counts its own
  children cannot survive a run that admits any.
- **`ih18.2#5`** — two of the PROC delta's seven behaviours are verified by
  tests that **can cite no acceptance criterion, because none exists for them.**

---

## The coordinator's own mistakes

A retro that omits these is not a retro. The coordinator recorded every
divergence as an `amend:` comment on the run epic and retracted three of its own
claims as `correction:` comments. **Three of the errors are the same defect.**

### The same defect, three times: reading a verdict through a pipe or a truncating filter

`AGENTS.md` names this explicitly, under "Pipes hide exit statuses". The
coordinator had read it, and had already filed it against itself twice in the
same run, and did it again.

1. **`bd lint` read through `tail -4` and `tail -2`.** The coordinator claimed
   `.20`, `.21` and `.22` "were checked too and were already clean." Measured on
   untruncated output: **three of the four mid-run beads failed the readiness
   gate, not one.** `.20` was missing Acceptance Criteria; `.21` was missing both
   Steps to Reproduce and Acceptance Criteria; `.22` was not clean either. Only
   the last bead in each truncated listing had been read.
2. **`bd dep add` run with output redirected to `/dev/null`, status never read.**
   The coordinator claimed to have "wired `.24<-.16`". **That edge was never
   created.** Measured: `bd dep add` exits 1 with `dependency ... already exists
   with type "discovered-from" (requested "blocks"); remove it first`. The
   dispatch agent's blocking note said "No formal bd dep edge names this
   prerequisite" and was correct; the coordinator's amendment contradicting it
   was wrong.
3. **A false tool claim copied verbatim into a dispatch prompt.**
   `run-record.md:81` documents the bounce ceremony as
   `bd create --add-label human (--labels exits 1 with unknown flag)`. Measured
   on bd 1.2.2, **exactly inverted for that subcommand**: `bd create --labels`
   → exit 0; `bd create --add-label` → exit 1 unknown flag. The parenthetical is
   true of `bd update`, where the flags are the other way round. The coordinator
   copied the inversion into the planner's prompt.

The self-correction is worth as much as the errors. The coordinator's own words
on the second one: *"the same defect I filed against myself twice already
today."* It then **audited all seven edges it had claimed that run** and found
exactly one absent — `.24<-.16` — confirming `.19` was correctly gated behind
nine dependencies so the run could not close its DoD early.

**What this says about the rule.** `AGENTS.md` already states, at length, that
this rule "has no automated check" and "is enforced by authorial memory, which
is the condition it exists to correct." This run is the fourth documented
instance and the first where the same agent hit it three times in one session
*while actively correcting others for it*. Stating it at the point of use
demonstrably did not close the gap. **Candidate C7 below proposes closing it.**

### Two other coordinator errors, both caught and corrected

- **A wrong bead id in a retro comment.** A `retro:` comment cited
  `agent-plugins-cf5j` for the tier-split finding; that bead does not exist. The
  real id is `agent-plugins-2296`. Verified both ways and corrected in a
  follow-up comment. *(The false id survives in the original append-only comment,
  which is correct per the protocol — corrections supersede, they do not edit.)*
- **A bead admitted without passing the readiness gate.** `ih18.23` was assigned,
  tiered, collision-marked and edge-wired — four of the five things admission
  requires — and never gated. Caught by an hourly status check while the bead was
  already in a dispatch queue. Filed as `agent-plugins-rd4d`, then
  *strengthened* by the truncation correction above: the gate is missing on
  admission **and** on deferral, and a deferred-out bead can carry a readiness
  defect no gate in the run will ever report.

### A dispatch agent silently reverted two coordinator decisions

`dev-bd97-4` filed `ih18.21` and `.22`; the coordinator triaged them (admitted
`.21` with an assignment, deferred `.22` out); the agent then read those field
values as a `bd create` anomaly and **cleared both** — silently, both at exit 0.

Measured on the agent's return: `.21` assignee empty, `.22` status open with
`deferred_until` empty, `updated_at` eleven minutes after creation and after the
coordinator's verified writes. Both re-applied and re-verified.

**The agent's "bd create anomaly" report is not a bd defect and must not reach
triage as one.** "Assigned to `$BEADS_ACTOR`" was the coordinator assigning `.21`
to `dev-bd97-4` — that agent's own actor string — and the `2026-09-28` defer date
is +30d from the coordinator's own `bd defer`. The real defect is that **a bead
records no provenance for `assignee` or `defer`, so a dispatch agent cannot tell
a coordinator's deliberate triage write from a tool anomaly.** Filed as
`agent-plugins-hjyq`.

This supersedes finding **`ih18.13#7`**, which reports the same three
`bd create` calls as unexplained bd behaviour with "Cause not established". It is
retained verbatim in the appendix as raw provenance, and it is explained.

---

## Review

Two passes of a budget of two. **17 findings standing, 0 admitted.**

`agent-plugins-ih18` still has exactly 25 children, so no review finding grew
the run's target.

### Pass 1 — the declared reviewer, and a coverage gap it declared itself

Six findings, all against defects this run's own diff introduced, none present
at `base_sha`. Filed as `bd97.4.1`–`.4.6`.

The reviewer **declared its own coverage gap rather than hiding it**: it
returned findings only against Go under `cli/swarm` and one shell script, and
nothing against the 8 modified skill markdown files, the 4 modified spec
documents, the 2 workflow files or `session-start.sh`. It stated plainly that a
clean result and absent coverage are indistinguishable from its report.

**Since this repository's actual product is the skills and specs, that is the
half that most needed looking at.**

### Pass 2 — spent on coverage, not on re-review

The coordinator dispatched pass 2 scoped to exactly the files pass 1 missed, and
did **not** re-run the declared reviewer. The reason is worth keeping:

> "Does this prose still describe the binary that now exists" is not a question
> a diff-reading code reviewer answers.

Pass 2 verified each documented claim **by running the command the document
names.** Eleven findings stand (`bd97.4.8`–`.4.18`); two were withdrawn as
duplicates of pass 1 (`.4.7`, `.4.19`), each closed with a comment naming the
survivor.

This is the run's best process result: **the second review pass bought coverage
of the repository's actual product, and it found the two most consequential
findings in the run.**

### The decision the reviewer put to the coordinator

Six of pass 1's findings are "row 1" — defects the run's merged diff introduced,
which `sprint-review` classifies as in scope whatever the gate says. The reviewer
asked whether `admit_review_findings=false` should hold.

**It held.** The coordinator's reasoning, recorded rather than escalated:
overriding a standing authorization mid-run because the findings turned out
severe "would substitute the coordinator's judgement for the one decision the run
actually put to a person."

### All 17 review findings

| id | P | title |
|---|---|---|
| `bd97.4.1` | P1 | `sprint-cleanup.sh` force-removes a live dispatch worktree when its HEAD is the previous dispatch's close commit |
| `bd97.4.2` | P2 | `measureWorktreeOwners` misattributes fresh worktrees and collapses two into one, so `swarm status` can report a false stall |
| `bd97.4.3` | P1 | `swarm status --dry-run` runs `git worktree prune`, so a dry run writes |
| `bd97.4.4` | P2 | A JSON `null` passes the manifest type check for every key and then reads as written |
| `bd97.4.5` | P2 | `swarm manifest get` exits clear for a requested key it could not establish |
| `bd97.4.6` | P3 | plan interpreter: a stop inside an alternative's arm loses skip records, and a stopped plan can report clean |
| `bd97.4.8` | P1 | **The vended `bead-bonanza` skill hard-requires `swarm` and tells consumers to build it from a path only this repo has** |
| `bd97.4.9` | P1 | **The `swarm` capability probe passes a binary whose manifest key set is stale, and `session-start` never rebuilds it** |
| `bd97.4.10` | P2 | `sprint-scope`'s documented `standing_auth` write is refused: the example omits four required rows |
| `bd97.4.11` | P2 | `sprint-recovery` still attributes worktrees by branch name, a convention this run retired |
| `bd97.4.12` | P2 | The `bead-bonanza` skills now cite `[[frd-swarm-core]]` for exit statuses they deleted — it resolves to nothing for a consumer |
| `bd97.4.13` | P2 | `run-batch`'s "exit 2 means no trailers" conflates that with an unreadable manifest, silently dropping commit trailers |
| `bd97.4.14` | P3 | The retired blueprint's DoD claims pre-flight is the last hand-rolled manifest reader; `bb-close-bead.sh` is a second one |
| `bd97.4.15` | P3 | `cli.yml`'s busybox step says `swarm` has "no subcommand yet"; it has two |
| `bd97.4.16` | P3 | Two prose defects the swarm adoption left: a retired "Run-level" qualifier and a reflow artifact in the epic DoR |
| `bd97.4.17` | P2 | `sprint-review`'s `bd create` example names flags bd does not have (`--discovered-from`, `--title`) |
| `bd97.4.18` | P2 | `swarm status`'s text output repeats a 40-command invocation list on every fact line |

Two withdrawn as duplicates: `bd97.4.7` (dup of `.4.1`), `bd97.4.19` (dup of
`.4.3`).

Three of these — `.4.10`, `.4.11`, `.4.12`, plus `.4.17` — are defects **in the
vended skills**, found only because pass 2 was spent on coverage. They would
have shipped invisibly.

---

## Routing

117 findings harvested across 26 beads. `bd export` over the run's closure set;
`bd97` (6) and `bd97.3` (3) carry the coordinator's and the plan phase's own.

| kind | n | routed to |
|---|---|---|
| `decision` | 34 | project |
| `gap` | 10 | project |
| `scope` | 9 | project |
| `stale` | 6 | project |
| `sequencing` | 6 | project |
| `unverifiable` | 2 | project |
| **project subtotal** | **67** | |
| `tooling` | 18 | skill — beads here, since this repository vends them |
| `proof` | 26 | the record, beside what they support. **No issue.** |
| `environment` | 3 | the environmental-failure ledger. **No issue.** |
| **unrouted** | **3** | — |

`[retros].skill_issues_dir` is deliberately unset in `sf.toml`, so
skill-audience issues are ordinary beads in this tracker rather than outbox
files. That is correct here: this repository *is* the one that vends them.

### The 3 unrouted findings, and why they matter more than the count

All three use kinds absent from the nine-kind taxonomy. All three are on
light-tier beads worked by `dev-bd97-2`, and the authorship — checked rather
than assumed — splits two ways:

| finding | kind used | author |
|---|---|---|
| `ih18.12#1` | `retro:process:run-batch` | `coordinator-bd97` |
| `ih18.17#1` | `retro:process:run-batch` | `coordinator-bd97` |
| `ih18.20#1` | `retro:work-clean` | **`dev-bd97-2`** — the implementing agent |

**The third is the most damning of the three, and it is the one an earlier draft
of this document got wrong.** `ih18.20#1` was not a coordinator repair. The
light-tier agent *did* record the absence, exactly as `run-batch/SKILL.md:260`
asks — "Work was clean: documentation changes reconciled between two documents
with no test failures. All checks passed on the modified tree." — and, having no
kind in the taxonomy for that, **invented one**. Its reward was to route
nowhere.

So the agent that complied and the coordinator that repaired non-compliance both
ended up outside the taxonomy, by two different routes. That is a gap in the
kinds, not a discipline problem in either agent.

**This is a recurrence, not a one-off.** From `docs/retros/log.md`: the `so39`
run recorded "2 findings unrouted (both `retro:process:`, both the
coordinator's)", and the `mpnf` run recorded "2 findings unrouted — third
consecutive run at 2" plus a coordinator invention of `retro:incident:` that
"sent the run's most consequential finding to no audience at all."

So `retro:process:` has now been independently invented by a coordinator in
**at least three runs**, and this run is the fourth consecutive with unrouted
findings — the first to rise above 2. Candidate **C8** proposes the fix.

---

## Validation

**Checks baseline at `3a57dc7`: all 13 passed, exit 0.** Recorded in the
manifest in full, including the 34 `template-drift` advisories that are emitted
at exit 0 and are not failures.

**Per-bead:** every bead ran the 13 declared checks before its own close — that
is the close gate — and every bead passed the five-signal acceptance check
post-merge.

**At the merged head `24f97e3`: not recorded.** See the next section.

**Pushes: nothing stranded.**
`git log origin/claude/bead-bonanza-agent-plugins-ih18-2yjc8j..24f97e3` returns
empty. All 24 commits are on the remote.

One environment fact the run had to discover to get there, now in the manifest
ledger: **the session PAT lacks the `workflow` scope, and a commit touching
`.github/workflows/**` rejects the entire push** — all four commits stranded,
remote ref absent. The proxy's brokered token *does* carry it. This is the
**inverse** of the `refs/dolt` rule, where the proxy 403s and the PAT is
required.

---

## What the run could NOT verify

Stated as prominently as what it could, because an unstated gap is
indistinguishable from one that does not exist.

- **The Windows CI leg has never executed.** Not once. `ih18.18` built it and
  could not run it: this sandbox has no Windows runner and no push/merge access
  to trigger a real Actions run. The workflow's triggers are push-to-main,
  pull_request and workflow_dispatch; measured via the GitHub API with `cli.yml`
  filtered to this branch, `total_count` is **0**. The individual pieces were
  each verified independently — the dolt download URL (200, redirecting to a real
  v2.3.1 asset), the `bd` install command (byte-for-byte against both install
  scripts), and the server auto-start mechanism — **but never all three together
  on `windows-latest`.** The DoD item is recorded as a named exception, not
  ticked. Filed as `agent-plugins-6jnx`; the first non-draft PR settles it.
- **`bd97.4.9`'s stale binary was measured in this container only.** The finding
  that the capability probe passes a stale binary rests on one container's
  `/usr/local/bin/swarm`. That the *probe* is weak is structural and reviewable
  from its source; that a *consumer* would hit it is inferred.
- **The 13 declared checks were never run over the merged head `24f97e3`.** Each
  bead ran them in its own worktree before closing, and no epic comment records a
  whole-tree run at head. Post-flight (`agent-plugins-bd97.6`) is still open and
  owns it. Settle with `scripts/bb-checks.py` at `24f97e3`. **This is a gap in
  the ceremony, not just in this run:** per-bead green does not compose to
  head-green, because beads merge in an order no bead tested.
- **`ih18.1#5` — the busybox/no-runtime CI step could not be run.** This
  container has the docker client but no daemon.
- **Whether `ih18.12` and `ih18.17` were genuinely frictionless is unknown.**
  Both closed recording nothing; the single comment on each is the coordinator's
  repair, added to satisfy the gate, "not by anyone who did the work". The
  findings say so themselves.
- **Whether tier is the cause of that is unestablished, and the sample is
  thinner than it looks.** `agent-plugins-tgex` states n=2 with both from the
  same dispatch id — but this run actually ran **three** light-tier beads, not
  two. `ih18.20` was admitted mid-run, tiered light, and worked by the same
  `dev-bd97-2`, and **it did record a finding** (`ih18.20#1`, self-authored,
  saying the work was clean). It falls outside `tgex`'s table because that table
  covers the 19 planned beads and `.20` arrived later. So the light-tier
  no-findings rate this run is **2 of 3, not 2 of 2**, from one dispatch id at
  one model — a weaker signal than the filed bead states, and worth carrying to
  whoever works `tgex`.
- **The `bd create` field anomaly's mechanism is explained but was never
  reproduced.** The coordinator's account (its own writes, misread) fits every
  observed value. No one re-ran it.
- **`ih18.24#4` — `bd lint`'s false-negative rate is unquantified.** One
  positive and one negative control were run. How often it fires wrongly across
  a real queue is unmeasured.
- **The reviewer's pass-1 coverage gap was declared, not measured.** "A clean
  result and absent coverage are indistinguishable from its report" is the
  reviewer's own statement about its own output.
- **No consumer project was run against the vended skills.** See "Read this
  first", item 1.

---

## Candidates — your decision

Eight candidates. Each is one decidable change. **Nothing below is filed yet;
say yes or no per candidate and I file exactly what you accept.** A declined
candidate stays in this document, and a later run hitting the same friction
raises it again with a higher recurrence count.

Everything the coordinator already filed during the run (16 beads) and the 17
review findings are **not** re-proposed here — they are listed after this
section.

---

### C1 — `bd-health.sh`'s repo-global exit status cannot gate one epic

**audience** skill · **target** `groom-beads` · **priority** P1 ·
**findings** `bd97.3#1` · **recurrence** 0

**What is wrong, and what it costs.** `groom-beads`' SKILL.md says *"Do not
declare the set ready while it exits non-zero."* But the script's exit status
reflects the **whole repository**, not the queue being gated. Measured this run:
`bash .claude/skills/groom-beads/scripts/bd-health.sh` → exit 1, `RESULT: hard
problems found in: swarm validate`, where **every named id is foreign to this
run** (`agent-plugins-qozp.6`, `agent-plugins-so39.2`, …) and every error reads
`is not an epic or molecule (type: task)`. Meanwhile
`bd swarm validate agent-plugins-ih18` → **exit 0, `Swarmable: YES`**, 19
issues, max parallelism 3, 12 waves. The instruction as written is unfollowable
in any repository with pre-existing beads: the gate can never go green, so it
gets ignored, and a real graph problem in the queue would be invisible in the
noise.

**Proposed change.** Scope the gate to the queue. Either have `bd-health.sh`
take the epic and filter its findings to that epic's descendants, or change
SKILL.md's instruction to read the per-epic `bd swarm validate <epic>` status
instead of the script's global exit.

---

### C2 — `partitioning.md`'s sequencing jq reads `c4_impacts` inverted, and ordered 12 of 33 pairs backwards

**audience** skill · **target** `sprint-plan` · **priority** P1 ·
**findings** `bd97.3#2` · **recurrence** 0

**What is wrong, and what it costs.** The reference and the grooming skill
define `c4_impacts` in **opposite directions**. `partitioning.md` says *"Bead B
whose `c4_impacts` intersects bead A's `c4_changes` consumes a contract A is
changing"*; `groom-beads`' SKILL.md defines impacts as *"downstream components
affected but not edited"*. Run over this run's 19 beads, the reference's jq
emitted `sequence agent-plugins-ih18.2 after agent-plugins-ih18.6` and
`... after agent-plugins-ih18.5` — but `.2` (PROC) is the **producer both are
built on**, and the blueprint's Stage 2 says *"The PROC delta SHALL land before
the TRK and VCS deltas."* **12 of the 33 sequence lines it emitted are inverted
this way.** A planner that trusted this output would build the dependency graph
backwards and serialize a run in the wrong order — worse than no signal, because
it looks like one.

**Proposed change.** Fix the direction in `partitioning.md`'s sequencing jq so
it matches `groom-beads`' definition (a bead whose `c4_changes` intersects
another's `c4_impacts` is the producer and sequences **first**), and state the
direction in one sentence beside the query so the two documents cannot drift
apart silently again.

---

### C3 — The protocol tells an agent to record nothing on a clean bead; the gate fails it for doing so

**audience** skill · **target** `findings-capture-protocol` · **priority** P1 ·
**findings** `ih18.12#1`, `ih18.17#1`, `bd97#6` · **recurrence** 0

**What is wrong, and what it costs.** Three vended artifacts give three
incompatible instructions for the same situation:

- `specs/standards/findings-capture-protocol.md:91` — *"**A clean run is fine —
  record nothing. Do not manufacture findings.**"*
- `plugin/skills/run-batch/SKILL.md:260` — *"**A clean bead records none — say
  so** rather than inventing filler."*
- `plugin/skills/bead-bonanza/scripts/bb-status.sh:167` — zero `retro:` comments
  → **`FAIL`**, `note $RC_REPAIR`.

The protocol is the declared single source of truth and it says record
**nothing**. An agent that obeys it fails signal 5. `run-batch` says "record
none — say so", which cannot both be done, since saying so *is* recording one.
The script's own comment concedes the tension (*"Zero is not automatically wrong
— a clean bead may say so explicitly"*) and fails anyway. Both light-tier beads
in this run did exactly what the protocol says and were marked failures; the
coordinator then wrote a comment on each to satisfy the gate, which is why the
only record of those two beads' friction was authored by someone who did not do
the work.

`agent-plugins-tgex` is filed and covers the *symptom*, framing it as agents
missing the "say so". **It does not notice that the source-of-truth protocol
licenses the behaviour the gate penalizes.** This candidate is the root cause and
is independently decidable.

**Proposed change.** Make the three agree, in the protocol first since it is the
source. Replace *"A clean run is fine — record nothing"* with an explicit
positive obligation — a clean item records exactly one finding stating the
absence — and give it a kind so it routes (see **C8**). Then bring
`run-batch/SKILL.md:260` and `bb-status.sh`'s signal 5 into line with it.

---

### C4 — The re-homed readiness gate inherits a `bd lint` false negative that `supervision.md` does not state

**audience** skill · **target** `bead-bonanza` (`supervision.md`) ·
**priority** P2 · **findings** `ih18.24#4` · **recurrence** 0

**What is wrong, and what it costs.** `ih18.24` re-homed three gates the retired
`bb-status.sh run` mode carried, and documented the `bd lint` gate in
`supervision.md` as the check that catches an ungated bead. Measured this run,
`bd lint`'s `Missing: ## Acceptance Criteria` check is a **case-insensitive
substring match on the description text**, not a check that the section exists:
a scratch bead whose description merely read *"...with no acceptance
criteria..."* was **not** flagged (`No template warnings found`), while an
otherwise-identical bead with the phrase removed **was**. So any bead that
*discusses* acceptance criteria as a topic reads as satisfied — including, as the
finding notes, `ih18.24`'s own description. This is precisely the class of bead
`agent-plugins-rd4d` exists to catch, so the documented remedy for `rd4d` has a
known hole in it and does not say so. `bd` is an external dependency and is not
this repository's to fix; the *documentation of the gate* is.

**Proposed change.** Add one sentence to `supervision.md`'s
"Re-checking readiness at every re-query" section stating the false negative and
its shape — `bd lint`'s section checks are substring matches on the description,
so a bead that names a section in prose passes without having it — and say what
the gate therefore does not guarantee.

---

### C5 — `strata` emits a JSON Schema that rejects the documents its own Go target produces

**audience** project (upstream report) · **target** `strata-lang/strata` ·
**priority** P2 · **findings** `ih18.7#2` · **recurrence** 0

**What is wrong, and what it costs.** With `allowAdditionalProperties: false`, a
`@discriminator` union emits the discriminated variant as
`{additionalProperties: false, properties: {text}}` — **without the
discriminator key itself** — while the Go marshaller emits
`{"kind":"text","text":"x"}`. Validating the second against the first fails for
**all four variants**: `at '': additional properties 'kind' not allowed`. **The
two targets of one source disagree, which is the exact failure ADR-016 adopts
strata to prevent.**

It was worked around, not fixed: `FactValue` is now a plain struct with a `kind`
field and four optional typed fields, producing an identical wire form. The
rejected escape hatches are recorded — `allowAdditionalProperties: true`
surrenders the closed key set that gives two ACs by construction; the OpenAI
draft emits `anyOf` correctly but makes every property required, destroying the
optional manifest. **The cost carried forward:** "exactly one value field, and it
is the one the kind names" is no longer expressible in the schema and is enforced
in `Result.Validate` instead — a constraint that moved from the shared model into
one language's runtime.

**This repository cannot fix it.** The candidate is to file a tracking bead so
the workaround has an owner and the upstream report is not lost.

**Proposed change.** File a bead recording the measurement, the workaround, and
the two rejected escape hatches, and report it upstream to `strata-lang/strata`.
Decline if you would rather carry the workaround silently.

---

### C6 — A fixed byte budget on a skill file makes a correctness edit compete with existing prose

**audience** skill · **target** `check-skills.sh` · **priority** P2 ·
**findings** `ih18.21#3`, `ih18.13#4` · **recurrence** 0

**What is wrong, and what it costs.** `check-skills.sh` sets
`BODY_BYTE_MAX = 20000` and `--strict` turns the warning into a failure.
`run-batch/SKILL.md` had **ten bytes of headroom**. `ih18.21`'s Done-when asked
that the file be updated to the command form; **every wording that both dropped
the false claim and named the command measured 20004–20092 bytes of body — over
the limit in every case.** The correction landed, but the command itself,
`swarm manifest get --run <epic> --key trailers`, went into two *other*
documents instead. `ih18.13` hit the same wall and needed three rounds of
trimming.

The finding states the general shape better than a summary can:

> A fixed byte budget on a skill file makes a **correctness** edit compete with
> existing prose, and the loser is whichever the author judges least costly to
> drop. That is a bad trade to leave implicit in a check.

`agent-plugins-tnyu` is filed and widens **that one file's** headroom. It does
not change the trade, which recurs on the next skill file to approach 20 KB.

**Proposed change.** Make the check distinguish "over budget" from "over budget
**and this change added prose**" — e.g. fail on growth past the limit while
permitting a net-neutral or shrinking edit that stays over, so a correctness fix
is never blocked by prose it did not write. Decline if you would rather keep the
hard cap and pay the trimming pass each time.

---

### C7 — The pipe/truncation rule has no automated check, and this run is the evidence that prose is not enough

**audience** project · **target** `AGENTS.md` / `scripts/ci.sh` ·
**priority** P1 · **findings** the coordinator's three `correction:` comments on
`agent-plugins-bd97` · **recurrence** ≥1 (`AGENTS.md` records four prior
instances)

**What is wrong, and what it costs.** `AGENTS.md` devotes a full section to
"Pipes hide exit statuses", lists four prior instances, and states outright:
*"the rule has no automated check … it is enforced by authorial memory, which is
the condition it exists to correct."* This run is the strongest available test of
that hypothesis and it failed: **the coordinator hit the defect three times in
one session while actively correcting other documents for it**, and twice after
having already filed it against itself the same day. The costs were real — a
false "all clean" verdict on three beads' readiness, and a dependency edge
believed wired that did not exist, which produced a wrong amendment contradicting
a correct agent report.

Two of the three were not `$?` through a pipe but a **verdict read through a
truncating filter** (`tail -4`, `tail -2`) and a **status discarded to
`/dev/null`**. The existing section names only the first form. The coordinator's
own generalization is the one worth adopting: *"The `AGENTS.md` rule about not
reading a verdict through a pipe applies to reading the OUTPUT through a
truncating filter too, not only to reading an exit status."*

**Proposed change.** Add a check to `scripts/ci.sh` that greps this repository's
own scripts and skill files for the three shapes — a pipeline into
`grep -q`/`head`/`tail` whose status is then tested, a command whose output is
redirected to `/dev/null` and whose status is not read, and `bd`-style verdict
commands piped into `tail`/`head` — and extend `AGENTS.md`'s section to name the
truncating-filter and discarded-status forms alongside the pipe form. A check
that catches even the literal shapes closes more of the gap than a fifth
paragraph.

---

### C8 — The taxonomy has no kind for a run's own conduct, so agents and coordinators alike keep inventing one

**audience** skill · **target** `findings-capture-protocol` · **priority** P2 ·
**findings** `ih18.20#1`, `ih18.17#1`, `ih18.12#1` · **recurrence** ≥2
(`so39`, `mpnf`, per `docs/retros/log.md`)

**What is wrong, and what it costs.** The protocol's nine kinds describe what an
**implementing agent** notices about work it was given. Two situations fall
outside all nine, and this run produced both. A **coordinator** writing on a bead
it did not implement — repairing a gate, recording that a bead's findings were
supplied by someone other than its author — has no kind that fits
(`ih18.12#1`, `ih18.17#1`). And an **implementing agent stating that its bead was
clean**, which `run-batch` explicitly asks for, has none either: `dev-bd97-2`
wrote `ih18.20#1` in good faith and had to invent `retro:work-clean` to say it.
All three routed nowhere, and they are not new:

- `so39` — *"2 findings unrouted (both `retro:process:`, both the
  coordinator's)"*
- `mpnf` — *"third consecutive run at 2"*, plus an invented `retro:incident:`
  that *"sent the run's most consequential finding to no audience at all"*
- `bd97` (this run) — 3 unrouted: two `retro:process:`, one
  `retro:work-clean`

`retro:process:` has now been independently invented by a coordinator in at
least three runs. **A finding routed nowhere reaches its audience only if a human
happens to read it, which is not a mechanism** — and the `mpnf` precedent shows a
run's most consequential finding can be the one that lands there.

**Proposed change.** Add a tenth kind to
`specs/standards/findings-capture-protocol.md` covering **how the run was
conducted**, as distinct from what the work order contained — spanning both an
observation recorded by an agent other than the item's implementer and an
implementer's own statement that its bead was clean. Adopt `retro:process` as
the spelling: three runs' agents reached for it independently, which is stronger
evidence of the right word than a fresh coinage would be. Define its routing
(project audience — it is about the run, not the work order) and add it to
`sprint-retro`'s routing table so it stops falling through.

**Relationship to C3, stated so you can decide them separately.** C3 fixes the
*contradiction* — three documents giving three instructions for a clean bead. C8
fixes the *destination* — that the resulting record has no kind and routes
nowhere. Accepting C3 alone leaves clean-bead records landing in the unrouted
pile; accepting C8 alone leaves an agent still unsure whether to write one. They
compose, and neither requires the other.

---

## Already filed during the run — not re-proposed

The coordinator filed 16 beads as it went. They are part of this run's output and
need no decision from you here.

| bead | P | title |
|---|---|---|
| `agent-plugins-221f` | P1 | Collision avoidance is component-granular; the real contention is a single shared document |
| `agent-plugins-2296` | P1 | Nothing enforces the tier-split rule at dispatch, so batches silently run beads above tier |
| `agent-plugins-tvom` | P1 | Plan phase should emit an ASCII wave graph so a serialized run is visible before dispatch |
| `agent-plugins-hjyq` | P1 | A dev agent cannot tell a coordinator's deliberate write from a tool anomaly, and silently reverts run decisions |
| `agent-plugins-rd4d` | P1 | A bead admitted mid-run never passes the readiness gate that every planned bead must pass |
| `agent-plugins-tgex` | P1 | Light-tier dispatches record no retro findings, and the acceptance gate is the only thing that notices |
| `agent-plugins-xeli` | P1 | A blocks edge orders readiness, not the worktree base |
| `agent-plugins-p39p` | P1 | Adopting a durable feature epic as the run target makes one object serve two incompatible completion semantics |
| `agent-plugins-5hzc` | P2 | Derived health's last-commit age reads the coordinator's HEAD, so a committing batch reads as idle |
| `agent-plugins-mie3` | P2 | `bd` refuses a blocks edge between beads already linked by discovered-from, silently leaving them unordered |
| `agent-plugins-atuh` | P2 | `run-record.md`'s bounce example uses `bd update`'s `--add-label` on `bd create` |
| `agent-plugins-6jnx` | P2 | Record the Windows leg's first real result against the swarm end-to-end tier |
| `agent-plugins-bacf` | P2 | The AC coverage table declares a verification level `testing-standards` does not define |
| `agent-plugins-wpfn` | P2 | Install `bd` on a CI leg and set `SWARM_E2E_REQUIRE_TRACKER` there |
| `agent-plugins-tnyu` | P2 | `run-batch/SKILL.md` sits within ~50 bytes of the 20000-byte body warning |
| `agent-plugins-nqe8` | P3 | `worktree-mechanics.md`'s merge/removal steps still hardcode `bb/<dispatch-id>` |

**`agent-plugins-cf5j` does not exist** — a `retro:` comment cited it for the
tier-split finding and was corrected; the real bead is `agent-plugins-2296`.

---

## Every finding, verbatim

All 117, grouped by kind, ids derived as `<bead-id>#<ordinal>` by position
within that bead's own findings. Nothing is filtered: the complete set is what
recurrence runs on, and a finding removed once can never be observed appearing
twice.

### `decision`

**agent-plugins-ih18.21#2**

```
retro:decision the fix names the stageless-ness on the Key itself, and validate.go now branches on that rather than keeping a second derivation of the same set
evidence:
Two ways to make a stageless key addressable were open.

Rejected: leave `buildKeys()` alone and special-case `trailers` (or
`contract.AlwaysOptional()`) at each of the three call sites that refuse an
unknown key — `declared()` in cli/swarm/cli/manifest.go, `manifest.Check`, and
`manifest.Require`. Cost: three places to keep in step, and a key added to
`RunManifest` tomorrow would still be refused by whichever one was forgotten.

Taken: `buildKeys()` appends the keys `contract.AlwaysOptional()` reports, with
an EMPTY `Key.Stage`, and the field's doc says empty means "required at no
stage", never "unknown". Every reader of the declared set then gets the key for
free — `Lookup`, `Keys`, `Check`, `Require`, `Validate`, `declared()`,
`declaredNames()` — and a new stageless key needs no code change at all.

Consequence checked, not assumed: `reached()` in cli/swarm/cli/manifest.go
skips any key whose `Stage != stage`, so an empty stage matches no real stage
and the derived-stage calculation is unaffected.

Second decision inside that one: `validate.go` held its own `alwaysOptional`
map, built from `contract.AlwaysOptional()`, used solely to keep such a key out
of the pending findings. With the stageless-ness now on the Key, that map was a
second derivation of the same fact, so it is deleted and the switch branches on
`key.Stage == ""`. Rejected the alternative of keeping both: the package's own
header says the declared key set has one source and "this package cannot hold a
fourth opinion about any of them", and two derivations is how the opinions
start. The behaviour is unchanged and is pinned by the pre-existing assertion in
TestValidationSeparatesAKeyNotYetWrittenFromOneThatIsWrong that `trailers` is
reported at no severity at all.

```

**agent-plugins-ih18.16#3**

```
retro:decision Retiring the whole `run` mode drops three checks `swarm status` does not carry, and I retired it anyway rather than keeping a stump.
evidence: the mode ran, besides the counts and health the CLI replaces, (1) `bd lint` findings intersected with this run's ready queue, (2) `bd list --label human --status open`, (3) every ready bead carries `metadata.tier`. `swarm status --run agent-plugins-bd97` reports none of them (full output on the proof comment above). The alternative I rejected was keeping `run` as a stump holding only those three: it would leave two commands answering "where is this run", which is the ambiguity the delta exists to remove, and the blueprint's anchor note says "The prose the anchors name is deleted, not merely superseded." Cost accepted: the supervision-time re-check is gone; all three still have plan-time owners (sprint-plan/SKILL.md line 124 for lint, grooming for tier, run-record.md for the human queue). Filed as agent-plugins-ih18.24 for the coordinator to settle.
```

**agent-plugins-ih18.15#1**

```
retro:decision The stall condition (ready front non-empty, no worktree stands, commit older than threshold) reads 'no worktree stands' as no DISPATCH worktree (bb/<agent>) standing, not the Collector's raw Worktrees.Standing count. Measured: a real e2e clone always registers at least the main checkout (git worktree list --porcelain never returns zero entries), so a literal Worktrees.Standing==0 check could never fire outside a unit-test mock. Rejected alternative: use the raw count as given — makes the attention branch unreachable in any real repository. Implemented as dispatchWorktreeCount() in cli/swarm/cli/status.go, matching the old bb-status.sh derived-health check's own convention of counting only worktrees matching the dispatch naming pattern, not the main clone.
```

**agent-plugins-ih18.15#2**

```
retro:decision Any unestablished measurement (worktrees, last commit, ready front, completion, or scope) maps to OutcomeAttention uniformly, whatever the reason — including a legitimately-absent 'admitted' key (AC-SWARM-CORE-003.10), which is not a failure but is still reported unestablished and exits non-zero per the FRD's literal text. Rejected alternative: OutcomeStop for a query failure vs OutcomeAttention for a legitimate omission — the FRD's Design paragraph draws no such distinction ('An unestablished measurement reports as unknown and exits non-zero'), and a query failure is usually transient/retriable, which is what OutcomeAttention means.
```

**agent-plugins-ih18.15#4**

```
retro:decision --stall-after defaults to 30m (1800s), matching plugin/skills/bead-bonanza/scripts/bb-status.sh's BB_STALL_SECONDS default (measured: bb-status.sh:538 `_stall="${BB_STALL_SECONDS:-1800}"`) — the derived-health check this delta's report is meant to replace (retirement itself is agent-plugins-ih18.16's job, not this bead's; supervision.md and SKILL.md were left untouched).
```

**agent-plugins-ih18.14#4**

```
retro:decision Each measurement runs as its own plan rather than all of them as one. Rejected alternative: one plan with every step set to OnFailure=Continue, which is fewer runner calls but makes the collector rewrite steps the Tracker and Version Control interfaces handed it — and a step whose failure policy the collector overrode is one the interface can no longer reason about.
evidence: plan.Failure documents Stop as the zero value and the ordering guard ("a step declaring it is honoured by the interpreter, so no command can chain a removal after a failed merge"). Under one shared plan the first failed read erases every later measurement; AC-SWARM-CORE-003.5 requires the opposite. Asserted by TestATrackerQueryThatCannotAnswerCostsOnlyItsOwnLine: with the ready query failing, worktrees, last commit and completion all still read established and the commit age is still 90s.
```

**agent-plugins-ih18.13#2**

```
retro:decision Kept run-batch on a workaround for `trailers` rather than blocking the bead or fixing the CLI inline: it now passes --trailers-from unconditionally and reads bb-close-bead.shs exit 2 as "the run resolved none". Rejected alternative: take the Blocked path behind the filed CLI bug, which would have stalled three further beads of this batch over one of twenty-four keys.
evidence: the workaround is safe because the script resolves trailers before it mutates anything. Measured: `bb-close-bead.sh --trailers-from agent-plugins-bd97 zzz-probe 1 "probe"` -> exit 2, "bb-close-bead: agent-plugins-bd97 carries no `trailers` in its manifest — omit --trailers-from when the run resolved none", and `git status --porcelain` immediately after was empty. Reading plugin/skills/bead-bonanza/scripts/bb-close-bead.sh confirms the trailers read at line ~150 precedes `id="$1"` and every close/stage/commit step.
```

**agent-plugins-ih18.11#3**

```
retro:decision the fixture's tracker half is HAND-WRITTEN JSONL, not an export captured from a live tracker.
evidence: cli/testdata/swarm-fixtures/core/tracker.jsonl is four records authored by hand — fixture-run (the epic, carrying a full manifest through the scope stage), fixture-run.target, fixture-run.1 (closed) and fixture-run.2 (in_progress) — with pinned timestamps (2026-01-01T00:00:00Z) and parenthood expressed as parent-child dependency edges on the child.
Rejected alternative: run a real bd session, close a bead, and `bd export` the result. It costs a fixture that carries every field the installed bd version happens to write, including ones no case asserts, so a bd upgrade rewrites the fixture's diff for reasons unrelated to the contract; and the timestamps would be whenever the recording happened, which is the thing the version stamp is supposed to make legible.
Held together by the import measurement above: bd import accepts hand-written JSONL and preserves it verbatim, so hand-authoring costs nothing in fidelity. Confirmed working end to end — the whole tier passes against it: 'ok github.com/brettcallaghan/agent-plugins/cli/swarm/e2e 14.117s'.
Also decided silently: the git half IS a real repository, committed as a git bundle built with pinned author, committer and date, and verified reproducible by building it twice and comparing. Fixed shas are what let the cases assert output containing a sha with no normalization.
```

**agent-plugins-ih18.11#5**

```
retro:decision SWARM_E2E_REQUIRE_TRACKER is the mechanism for the blueprint's 'WHERE a case needs the tracker and the platform is one the project declares the tracker present on, an absent tracker SHALL fail the case rather than skip it'.
evidence: requireTools in cli/swarm/e2e/harness_test.go skips with the reason when git or bd is absent, unless SWARM_E2E_REQUIRE_TRACKER is set in the environment, in which case it fails. An environment variable rather than a build tag or a platform list in Go: the declaration belongs to the platform that installs the tools, which is the workflow file, and a list compiled into the test would have to be edited from the other side of the fence every time a leg changes.
The gap it leaves is real and is now a bead: no leg of .github/workflows/cli.yml sets it, because no leg installs bd, so the tier currently SKIPS on all three platforms and contributes nothing on CI. Filed as agent-plugins-wpfn (discovered-from this bead) rather than fixed here — installing bd on a runner is a workflow change with its own cost (libicu-dev, a source build) and is not this bead's scope. Named in the workflow's own Test-step comment and in the container blueprint so the silence is visible where somebody would read it.
```

**agent-plugins-ih18.10#2**

```
retro:decision Three shapes the FRD's command sketch leaves open, each decided against what a caller can do about the answer.
evidence: (1) the FRD writes 'swarm manifest set --key <key> --value <value>' with no plural form, but AC-SWARM-CORE-002.13 requires several keys in ONE invocation. Chose repeatable --key and --value paired IN ORDER, with a mismatched count refused as a malformed invocation naming both counts. Rejected: --key k=v pairs, which would need an escape for a value containing '='; and a JSON object argument, which puts the caller back to hand-writing the merge this command exists to perform.
(2) 'validate' takes no stage in the FRD, and validation cannot happen without one. Chose: --stage optional, and when absent validate against the stage the record has REACHED (the furthest stage all of whose keys are present), reporting both the stage and the fact that it was derived as facts of their own. Rejected: requiring --stage, which makes the commonest invocation need knowledge the record already carries; and defaulting to the last stage, which fails every run partway through.
(3) An object-valued key has no fact kind — the contract declares text, count, flag and list. Chose: report it as text carrying its compact JSON. Rejected: adding an object kind to the schema, which is a contract change made for a rendering convenience.
```

**agent-plugins-ih18.9#1**

```
retro:decision 'A key whose absence is always legitimate' had no way to be said, and the schema was changed so it could be said structurally rather than in a second list.
evidence: every manifest key is optional on the wire, so optionality cannot distinguish 'required at stage S' from 'never required'. The delta's own rules turn on that difference — AC-SWARM-CORE-002.17 asks validation to report a key absent at a stage that does not require it as not-yet-written, and trailers is documented in run-record.md as 'the commit trailers the run resolved, OR ABSENT when the run resolved none', i.e. legitimately absent forever. Reporting it as pending at every stage would tell a reader to wait for something never coming; reporting it as a failure would fail every honest run.
Rejected alternative: a list of never-required keys in Go, which is a second declaration of a key property and the exact drift AC-SWARM-CORE-002.19 fights. Chosen: move trailers OUT of the stage chain and declare it directly on RunManifest. Having no stage is then expressed by having no stage struct, and contract.AlwaysOptional() DERIVES the set as RunManifest's keys minus the chain's. The contract test asserts the derivation is non-empty and finds trailers, so it cannot rot into a vacuous pass.
```

**agent-plugins-ih18.9#2**

```
retro:decision Findings carry a rule IDENTIFIER and a severity, not just a message, and validation is split in two: Validate over a record as it stands, CheckWrite over a proposed change against it.
evidence: the delta asks for 'every failing rule rather than the first' (AC-SWARM-CORE-002.15) and for separating 'a key absent at a stage that does not yet require it' from 'one that is wrong' (002.17). A list of strings satisfies neither well — a caller cannot act on prose, and 'not yet written' and 'wrong' would be distinguishable only by wording. So a Finding is {Rule, Severity, Keys, Message}, Report.Failed() ignores pending findings, and every test asserts the RULE rather than the text.
The split is forced by which rules need what: write-once, counter-backwards and ledger-append-only can only be seen by comparing a change to the record, while admission overlap and the stage rules are properties of one record. Putting both behind one function would have meant passing a change nobody is making. The admission rule appears in both, deliberately — CheckWrite runs it over the record as it WOULD stand, so an overlap is refused before it happens rather than only reported afterwards, which is tested.
```

**agent-plugins-ih18.8#2**

```
retro:decision The declared key set is read by REFLECTION off the generated model rather than written down in this package. That is what makes AC-SWARM-CORE-002.19 true of the code and not only of the schema.
evidence: manifest.Keys() walks contract.StageDeclarations() — the five generated stage structs in run order — and takes each field's json tag as the key name, the first stage carrying a key as its stage, and the field's type with the optional pointer stripped as the type a value must unmarshal into. So Check() refuses an undeclared key and a wrongly-typed value using the same declaration a consumer validating against the JSON Schema would use, and a key added to schemas/swarm/manifest.strata is known to reading, requiring and validating at once with no edit here.
Rejected alternative: a table of keys and types in Go, which is a second declaration and the exact drift the AC names. The cost accepted: reflection means a mis-declared field is a runtime discovery rather than a compile error, which is why TestTheDeclaredKeySetIsReadFromTheGeneratedDeclaration asserts the derivation reached a plausible set (more than twenty keys, five spot-checked stages, no duplicates, every key typed) instead of trusting it.
```

**agent-plugins-ih18.8#4**

```
retro:decision Two structural calls the delta leaves open: (1) the metadata write is a method on the Beads Tracker's CONCRETE type, not on the Tracker interface; (2) the plan fragment type moved into the plan package and vcs.Steps became an alias of it.
evidence for (1): container-swarm-cli.md says both 'the Tracker interface does not carry a metadata-blob operation' and 'the store reaches beads through the Beads Tracker rather than invoking bd itself'. The only shape satisfying both is a method the interface does not declare, which the store reaches by naming the concrete type. So beads.Tracker.WriteMetadata exists and tracker.Tracker is unchanged, and the rule that only two components in this container name a tool still holds.
evidence for (2): the Version Control item introduced vcs.Steps for operations needing more than one invocation; this item needs the same thing for the conditional write. A third copy would be three types for one idea, so the type is now plan.Steps and 'type Steps = plan.Steps' in vcs keeps every existing call site and test unchanged — verified, the vcs suite passed without an edit. Scope note: that is a third change to the plan package from a bead whose c4_changes does not name it, after Record.Value and Step.Answers. All three were forced by a consumer needing something the model did not offer; the pattern is worth a look at the retro rather than three separate excuses.
```

**agent-plugins-ih18.7#1**

```
retro:decision The manifest's stage is expressed by INHERITANCE, not by a table: five key structs chain in run order — PreFlightKeys, ScopeKeys, PlanKeys, DispatchKeys, RetroKeys — and RunManifest composes them. Each key is written in exactly one struct, so its name, its type and its stage come from one declaration, which is what AC-SWARM-CORE-002.19 asks for literally.
evidence: Strata has no annotations, so 'the stage at which a key becomes required' cannot be attached to a field; the only structural way to say it is which declaration the field is in. Rejected alternative A: one all-optional RunManifest plus a separate stage table in Go — two lists of key names, and a key added to one and not the other is exactly the drift the AC exists to prevent. Rejected alternative B: per-stage structs with REQUIRED fields, so each stage's JSON Schema 'required' list is the stage's requirement — Strata cannot relax requiredness back down, so the Go model would have every field non-pointer and an absent branch would read as "", losing the absent-vs-empty distinction AC-SWARM-CORE-002.3 explicitly requires. Every key is therefore optional on the wire and the chain carries the stage; requiredness at a stage is computed by the validator from the chain, which is the MANV item's work.
```

**agent-plugins-ih18.7#3**

```
retro:decision standing_auth and models are STRUCTS with named fields, not maps. That is what let the schema be closed without changing a wire format the skills write today.
evidence: the live manifest on this run's own epic carries standing_auth rows whose values are mixed — {"pull_request":{"value":"draft"}}, {"review_budget":{"value":2}}, {"merge_to_default":{"value":false}}. A map [String: Row] would need Row.value to be of mixed type, which a closed schema has no way to express; every union or split-field encoding of it changes the bytes. But the rows are a FIXED set — sf.toml's [run] block lists exactly the six gate rows — and each row's value has a fixed type, so naming the six fields and giving each its own row type (TextAuthRow, FlagAuthRow, CountAuthRow) describes what is written today, byte for byte. Same reasoning for models, whose four keys are the four capability tiers.
Proved rather than argued: TestAManifestOfTheShapeTheSkillsWriteTodayValidates validates a manifest transcribed from the live one against the generated RunManifest schema, and also asserts that a partial manifest validates, that an undeclared key is refused, that a wrongly-typed key is refused, and that a wrongly-typed row inside standing_auth is refused. The delta's own instruction — 'carry over the keys run-record.md specifies today; do not redesign them' — is the constraint this satisfies.
```

**agent-plugins-ih18.6#3**

```
retro:decision The Version Control operations return a FRAGMENT (vcs.Steps, an ordered slice whose last step carries the answer) while the Tracker's return a bare plan.Step. The asymmetry is deliberate and is the one thing a reviewer should push back on if they disagree.
evidence: the delta requires 'The Git Version Control SHALL clear stale worktree registrations before returning a registration list', and clearing them is a second invocation — git worktree prune, then git worktree list. Rejected alternative A: make Worktrees the only method returning a slice, which leaves a caller having to remember which shape each method has. Rejected alternative B: make Worktrees a single list step and let the decoder merely FILTER prunable entries, satisfying the intent without 'clearing' anything — the stale registration then stays on disk forever and the delta's verb is not honoured. Rejected alternative C: retrofit fragments onto the Tracker for symmetry, which is ceremony on eight single-invocation operations with no case for it.
Both halves are in fact implemented: the prune runs AND the decoder drops prunable entries. That is not belt-and-braces — a dry run refuses the prune, since it is a write, so the filter is the only thing keeping the standing-worktree count right in that mode. TestADryRunRefusesThePruneAndStillDropsTheStaleRegistration is the case.
```

**agent-plugins-ih18.5#4**

```
retro:decision Four shape decisions the delta leaves open, each taken toward the conservative reading: (1) every Tracker operation returns a plan.Step rather than performing a call; (2) the transitive walk is Descendants in the INTERFACE package, not a method on the beads implementation; (3) edge writes are two named operations, BlockedBy and DiscoveredFrom, not one generic Link; (4) a status the run's vocabulary does not cover becomes StateUnknown with the tracker's own string kept beside it.
evidence: (1) container-swarm-cli.md Change Guidance — 'A command SHALL NOT invoke an external tool outside a plan step, since a call made directly is invisible to explain mode and unguarded by the ordering policy'; the C4 edge is Rel(beads, proc, 'Declares invocations and decoders for'), which reads as declaring rather than calling. (2) the delta says 'the Beads Tracker SHALL return it transitively', which would put the walk in the implementation; putting it over the interface makes the property hold for ANY implementation and keeps the beads package to declarations. The cost: the delta's sentence is now satisfied by a helper the implementation does not own. (3) measured, bd's own order reads backwards — 'bd dep add A B' means A depends on B — and the audit records getting it backwards as inverting every dependency at exit 0; naming the direction in the operation makes that impossible at a call site, at the price of a method per edge kind. (4) rejected: map an unknown status to StateOpen, which is what a default case would do and would make a bd release that adds a status read as 'nobody is working it' for every bead carrying it — a difference dispatch turns on.
```

**agent-plugins-ih18.4#1**

```
retro:decision Resolved the anchor conflict the Plan Runner item recorded, in the blueprint's favour: cli/swarm/cmd was renamed to cli/swarm/cli.
evidence: the SURF delta's anchors read '**Anchors:** `cli/swarm/main.go` and `cli/swarm/cli/`; `github.com/spf13/cobra`, already a dependency of the module.' The SKEL item had put the surface at cli/swarm/cmd/, mirroring cli/cmd/ — which its own anchors named as a CONVENTION to follow, with no destination. Chose to follow the blueprint (git mv cli/swarm/cmd cli/swarm/cli; package cmd -> package cli; one importer, cli/swarm/main.go). Rejected alternative: keep cli/swarm/cmd and edit the blueprint's Anchors line — editing a spec to match code is the drift direction this project explicitly opposes, and the rename cost one commit's worth of churn while a stale anchor costs every future reader. The rename lands in this item's commit rather than a separate one because SURF owns the Swarm Command Surface component, which is the component the directory holds.
```

**agent-plugins-ih18.4#3**

```
retro:decision The plan runner's output is CAPTURED, not streamed, in every mode. The delta says 'WHEN a plan is executed, the Plan Runner SHALL print each step's invocation AS IT RUNS IT', which reads as streaming.
evidence: two obligations collide. Streaming to stdout puts transcript lines ahead of a --json envelope on the same stream and corrupts it; and AC-SWARM-CORE-001.9 requires the machine-readable form to carry every fact the text form carries, which includes the transcript. Capturing into a buffer and rendering it in both forms satisfies both. Rejected alternatives: (a) stream in text mode and buffer in JSON mode, which makes the two forms carry different things and leaves 001.9 true only by inspection; (b) stream to stderr in JSON mode, which splits one command's output across two streams for a caller that reads one. Cost accepted: an operator watching a long execute sees nothing until it finishes. The audience is an agent capturing output, and a swarm command runs a handful of subprocesses for seconds — but this is the assumption to revisit first if a command ever gets slow.
```

**agent-plugins-ih18.4#4**

```
retro:decision Three refusals the delta implies but does not state, each enforced by the surface so it is true of every command rather than of the ones that remembered: (1) rendering a Result whose outcome is 'attention' with no resolution is refused; (2) rendering a fact that names no invocation is refused; (3) --explain together with --dry-run is a malformed invocation.
evidence: (1) the delta says 'WHEN a command exits with the attention status, its output SHALL name what would resolve the condition' — a command that forgot would otherwise ship a status with nothing to act on, so the surface turns it into a stop rather than printing a bare attention. (2) 'Each reported fact SHALL name the invocation that established it', same reasoning. Both are refused as OutcomeStop, not OutcomeMalformed: a defect in the command that built the result is not something the caller can repair by re-invoking. (3) is stated nowhere; the two flags select different interpreters, and silently preferring one would make a caller believe a dry run happened when it did not. Rejected for (1) and (2): render anyway and let a reviewer catch it, which is the class of silent failure this container exists to remove.
```

**agent-plugins-ih18.3#2**

```
retro:decision What happens AFTER a selected alternative arm finishes is specified nowhere. Chose: the arm's steps run immediately after the step carrying them, and the outer plan then continues with its remaining steps.
evidence: the delta says only 'WHERE a step carries an alternative arm, explain output SHALL render both arms and name the condition that selects between them' — a rendering rule, with no execution semantics. ADR-024 says 'Branching is a property of a step (an alternative plan attached to a named outcome), not of the code that built it' and names the motivating case as 'one conditional retry'. Run-and-continue is the semantics a conditional retry needs. Rejected alternative: the arm REPLACES the remainder of the plan (a true branch), which makes a retry impossible to express without duplicating every following step into the arm. Cost if wrong: a command needing replace-the-remainder must put the remainder in the arm, or split into two plans — which ADR-024 already contemplates ('a command whose logic genuinely needs a result before it can decide what to do next must either express that as a branch on the step or split into two plans').
```

**agent-plugins-ih18.3#3**

```
retro:decision Three resolution rules the delta does not state, each chosen for the conservative reading, each with a test: (1) a step publishes its decoded value only when it exited 0, so a later step referencing a FAILED step is skipped with its reason recorded rather than handed whatever the failing tool printed (TestAReferenceToAStepThatFailedIsNotResolved); (2) a dry run that refuses a write CONTINUES rather than stopping, because AC-SWARM-CORE-001.6 says 'its report reflects the state the reads returned' and a dry run that stopped at the first write would report almost nothing; (3) a step with no decoder yields its stdout with surrounding whitespace trimmed, so a single-value read needs no decoder at all.
evidence: the delta's behaviour list covers what a step DECLARES and what each interpreter RUNS, and says nothing about what a reference resolves to when its producer failed, nor whether a refusal ends a dry run. Rejected for (1): publish regardless, which is how a shell pipeline behaves and is the silent-failure shape this container exists to remove. Rejected for (3): require an explicit decoder on every step, which is ceremony on the commonest case and would have been the first thing a command author worked around.
```

**agent-plugins-ih18.3#4**

```
retro:decision Added Plan.Validate, which no delta behaviour asks for: it refuses a step with no name or no executable, a duplicate step name, an argument taken from a step that is not strictly earlier, and an alternative naming no condition — before anything is printed or run.
evidence: the delta says 'WHERE an argument is produced by an EARLIER step, explain output SHALL render it as that step'. Without a check, 'earlier' is a convention: a plan referencing a later step renders a placeholder in explain and fails at resolution time during execute, which is the drift-between-modes this design exists to prevent. Validate makes it a property of the plan. Rejected alternative: resolve lazily and error at execute time, which leaves explain confidently printing a plan that cannot run. Cost accepted: one more refusal a command author can hit, and it is refused at the same point in all three modes.
```

**agent-plugins-ih18.2#3**

```
retro:decision A non-zero exit is DATA on the Result, with a nil error; the error return is reserved for an invocation that produced no status at all. Neither the delta nor ADR-025 says which of the two shapes to use, and the choice changes every call site above.
evidence: the delta says 'The Process Runner SHALL read an invocation's exit status from the finished process' and 'IF the named executable cannot be found, THEN the Process Runner SHALL report that distinctly from a non-zero exit' — a distinction, with no shape. Chose (Result, nil) for a non-zero exit because a tool answering no is a successful invocation of it, and because the alternative — returning an error for every non-zero exit, the way os/exec itself does — puts the tracker's ordinary 'nothing found' answers on the error path, where a caller that checks err first stops reading the result it needs. Cost accepted: a caller that only checks err will read a failed step as a success, which is why Result.Succeeded() exists and why the exit status is the first field a decoder is expected to look at. A missing executable returns (Result{}, *NotFoundError) with a ZERO result rather than a partially-filled one, so a caller checking the status first cannot read it as exit 0.
```

**agent-plugins-ih18.2#4**

```
retro:decision Invocation carries Name, Args and Dir — and deliberately no Env and no String()/rendering method.
evidence: Dir is in because container-swarm-cli.md reads git 'for history, branches and worktrees' and the alternative (the calling process chdir-ing) is shared mutable state across concurrent invocations. Env is OUT because the run's own actor identity reaches a tool through the inherited environment, which os/exec gives for free, and a per-invocation Env field would invite call sites to construct one; adding it later is one field with no call-site churn. String() is OUT because AC-SWARM-CORE-001.13 requires explain output to be a RUNNABLE transcript, and how an argv is quoted so it can be pasted into a shell is the Plan Runner delta's decision, not this one's — a naive strings.Join here would have been adopted as the answer by whoever writes explain mode. The fake formats argv with %q locally for its own error messages, which is a diagnostic, not a transcript.
```

**agent-plugins-ih18.1#1**

```
retro:decision A bare invocation prints the command tree to STDERR, not stdout. The delta says 'print its command tree and exit with the malformed-invocation status' and names no stream; cobra's default usage rendering goes to the out writer.
evidence: specs/features/swarm-core/swarm-core-blueprint.md, SKEL Behavior: 'WHEN the binary is invoked with no arguments, it SHALL print its command tree and exit with the malformed-invocation status' — no stream named. Chose stderr so that stdout stays reserved for a command's answer, which the FRD's '--json | machine-readable output carrying the same facts as the text form' convention requires of every command. Rejected alternative: cobra's default (usage to stdout), which would put usage text where a --json consumer expects a result. Implemented by SetOut(streams.err) inside the root RunE in cli/swarm/cmd/root.go, so --help still goes to stdout. Measured on the built binary: bare invocation -> exit 3, stdout empty, stderr 'Usage:\n  swarm [flags]\n\nFlags:\n  -h, --help   help for swarm'.
```

**agent-plugins-ih18.1#2**

```
retro:decision All four exit statuses are named as constants in cli/swarm/cmd/root.go, though only exitClear and exitMalformed are reachable from this delta. The four-status mapping is the Swarm Command Surface delta's acceptance criteria (AC-SWARM-CORE-001.1/001.2), not this one's.
evidence: specs/features/swarm-core/swarm-core-frd.md FEAT-SWARM-CORE-001 table: 'exit 0 | the command answered, and its verdict is clear', 'exit 1 | ... needs attention', 'exit 2 | stop — not the caller's to repair in place', 'exit 3 | the invocation itself is malformed'. Rejected alternative: define only exitMalformed, which leaves the outcome-to-status mapping to be reinvented at each command that first needs 1 or 2 — and container-swarm-cli.md puts that mapping on this surface ('the mapping from an outcome to one of four statuses'). Go does not flag an unused constant, so nothing breaks while 1 and 2 are unreachable. Cost if wrong: the SURF delta renames or reshapes two constants in one file.
```

**agent-plugins-ih18.25#1**

```
retro:decision Provisioning decided as: .claude/hooks/session-start.sh builds swarm from THIS repo's own cli/swarm on every cloud session (new step, section 4, before strata), mirroring how bd and strata are provisioned there. Unlike bd/strata this needs no clone and no network -- cli/swarm is already in the checkout -- so it is cheap and placed early rather than last. Idempotency uses the SAME behavioural probe as pre-flight-check.sh (swarm status --help) rather than a version string, since swarm carries none. Escape hatch SWARM_CLI_SKIP=1, matching STRATA_CLI_SKIP's naming.
```

**agent-plugins-ih18.25#3**

```
retro:decision The swarm build in session-start.sh sits inside the hook's existing CLAUDE_CODE_REMOTE-only gate (the same early-exit bd and strata already live behind), rather than running unconditionally. Rejected alternative: build swarm regardless of remote/local, since it is cheap and repo-local -- rejected because a local dev session may already have its own build in /usr/local/bin/swarm and silently overwriting it on every session start would be surprising, matching the hook's own stated reasoning for gating bd/strata the same way.
```

**agent-plugins-ih18.24#2**

```
retro:decision Chose option (b) from the bead's own Design: the coordinator runs the three retired gates by hand at every re-query, documented in plugin/skills/bead-bonanza/references/supervision.md (new section 'Re-checking readiness at every re-query') and named as an owned Coordinator responsibility in specs/architecture/module-bead-bonanza.md's Coordinator bullet list, with a one-line pointer added to SKILL.md's existing 're-query before every dispatch' paragraph. Rejected (a) swarm status grows them: FEAT-SWARM-CORE-003's acceptance criteria (specs/features/swarm-core/swarm-core-frd.md) never scoped bd-lint/human-queue/tier as measurements, and building them into cli/swarm/status would mean new Measurements fields, new Deps seams, new Go tests, and FRD amendments touching specs/architecture/container-swarm-cli.md — which this dispatch is instructed to stay out of (a concurrent dispatch, dev-bd97-2, is working ih18.20 on exactly that file). Rejected (c) declaring the plan-time gate sufficient: the bead's own AC requires catching a mid-run addition, which by definition plan-time grooming cannot see.
```

**agent-plugins-ih18.23#1**

```
retro:decision Adopted the coordinator's directed rule: a worktree is attributed to a dispatch via its HEAD commit's subject (bb-close-bead.sh convention <id> P<priority> <title>) parsed for a bead id, then that bead's .assignee — never via the branch name. Implemented as measureWorktreeOwners() in cli/swarm/status/status.go (Go) and owner_for_worktree() in plugin/skills/bead-bonanza/scripts/sprint-cleanup.sh (bash, via git -C <path> log -1 --format=%s + bd show <id> --json). DispatchBranchPrefix and worktreeFor() removed from the Go collector; owner_for_branch() removed from sprint-cleanup.sh.
```

**agent-plugins-ih18.23#2**

```
retro:decision The sprint branch's own worktree (the coordinator's main checkout) is excluded from attribution by structural comparison (tree.Branch == m.Branch, the manifest's recorded sprint branch), not by heuristic (detachment, path, or trying to detect main). Rejected alternative: attempt commit-subject attribution on every tree including main — main's HEAD is very often itself a bead-close commit (the sprint branch's tip usually IS the last closed bead), so main would frequently misattribute to whichever dispatch closed last. When m.Branch is itself unestablished (manifest carries no branch key), this exclusion cannot apply and attribution is attempted on all trees including main as a degraded fallback — accepted as a rare, low-stakes edge case rather than solved further.
```

**agent-plugins-ih18.19#3**

```
retro:decision Ticked the AC-coverage DoD item rather than blocking on one row declared at a verification level no standard defines, and recorded the exception in the blueprint instead.
evidence: The item requires every AC 'verified at its declared level and cited from its tests per ac-traceability-standards'. All 45 are cited and no level is contradicted — but AC-SWARM-CORE-002.13 is declared 'integration', and 'grep -n integration specs/standards/testing-standards.md' returns nothing: that document's '## Test Layers' names two, '### Unit Tests' and '### Snapshot Cases'. The criterion IS verified twice (cli/swarm/cli/manifest_test.go 'TestSetAppliesSeveralKeysAsOneMerge'; cli/swarm/manifest/beads/store_test.go 'TestSeveralKeysInOneInvocationAreOneReadAndOneWrite'). Rejected alternative: re-declare the row 'unit' to make the item literally true — that would have silently deleted the evidence that the corpus's layer vocabulary is short one entry, and the gap governs every blueprint, not this one. Filed as agent-plugins-bacf. Cost of the choice: the blueprint retires carrying a named exception a reader must read past.
```

### `gap`

**agent-plugins-ih18.14#2**

```
retro:gap The COLL delta says what to do about a dispatch still holding work and nothing about naming one that has drained, so the first implementation named only agents with claimed, blocked or ready beads. The AC is what caught it; the delta text alone would have shipped a collector that cannot say "drained".
evidence: swarm-core-blueprint.md, the Status Collector delta, lists "SHALL determine a dispatch drain state from two queries" and "SHALL report a dispatch holding a bead in progress as still holding work" — neither requires a drained dispatch to appear at all. AC-SWARM-CORE-003.4 does: "Given two dispatches, one of which has drained ... And the drained one is reported as drained." Observed failure before the fix: "dispatches = [{... Agent:dev-1 ...}], want one per agent" — dev-2, whose only bead had closed, was absent from the set entirely. "Drained" and "never existed" printed identically.
```

**agent-plugins-ih18.10#3**

```
retro:gap The Manifest Store's write condition had to be generalised from a value to a predicate, one item after it shipped, because 'set' needs two conditions at once.
evidence: the store's Write took 'expect manifest.Document' — the values the covered keys held when the caller read — which expresses exactly one condition. This command needs three combinations: the logical rules that can only be seen against the record (write-once, counter, ledger, admission), --if-absent, and both together. So Write now takes a manifest.Guard (func(Document) error), with manifest.Unchanged, manifest.Absent, manifest.Rules and manifest.Guards composing them. The store's own tests changed by one call each and the property they assert is unchanged.
Why this is a gap rather than a mistake: the Manifest item's delta lists 'The store interface offers a write conditional on the record not having changed since the fetch' and says nothing about any other condition, so the narrower signature was the delta as written. What it could not anticipate is that the SAME step — the read whose decoder guards the write — is the only place any record-dependent rule can be checked, which the validation item then established. A predicate was the shape all along; nothing in either delta said so.
```

**agent-plugins-ih18.6#2**

```
retro:gap The Plan Runner had no way for a step to say that a non-zero exit is an ANSWER rather than a failure, and git's ancestry query needs exactly that. Added plan.Step.Answers and plan.Record.Answered in this commit.
evidence: 'git merge-base --is-ancestor A B' answers through its status — measured, exit 0 yes, exit 1 no, exit 128 unresolvable. Before this change a legitimate 'no' set Record.Failed() true, published no value, and (with the default OnFailure) stopped the plan — a correct answer read as a failure, which is the mirror image of the failure-read-as-a-pass the container exists to remove. Nothing in ADR-024, ADR-025 or the PLAN delta contemplates the case; every example in them is a tool whose answer is its output.
The alternative I rejected: express ancestry as two exit-0 invocations (rev-parse the ancestor, merge-base the pair, compare in Go). It avoids touching the plan package but makes the comparison a caller's job, so every consumer of 'is this branch cut from the sprint branch' reimplements it — which is the thing the interface exists to prevent.
Backwards compatible by construction: Answers empty means only 0, so Failed() is unchanged for every step that does not declare one, and the whole existing suite stayed green. Scope note: this bead's c4_impacts is ["Status Collector"], so the Plan Runner is again outside the declared impact set — same call as on the Tracker item, recorded rather than quiet.
```

**agent-plugins-ih18.5#3**

```
retro:gap The Plan Runner had no way to hand a decoded value to anything but a later step's reference argument, so a decoder's output was unreachable by the command that needs it. Added plan.Record.Value in this commit.
evidence: cli/swarm/plan/run.go kept decoded values in the interpreter's unexported 'state' and published them only through resolve(); plan.Record carried the invocation, status, duration and stream digests and nothing else. A command turning a plan's outcome into facts reads plan.Report, so every decoder this delta ships would have had no reader. The addition is one field, populated on the same condition a reference is (the step ran and exited 0), with its own test — TestTheRecordCarriesWhatTheDecoderMadeOfTheResult asserts both that a successful step's value is on the record and that a FAILED step publishes nothing.
Scope note: this bead's metadata declares c4_impacts ["Manifest","Status Collector"] and not the Plan Runner, so strictly this is a change outside the declared impact set. Filing it and blocking would have stranded the batch over one additive field; the alternative — decoders nothing can read — is not a shippable delta. Recorded here rather than quietly.
```

**agent-plugins-ih18.3#5**

```
retro:gap Two things the delta requires without saying in what form, decided here: the transcript's quoting and the stream digest's algorithm.
evidence: 'WHEN a plan is explained, the Plan Runner SHALL print each step's invocation in order' plus AC-SWARM-CORE-001.13 'when its printed invocations are run by hand in order, then they perform what the command would have performed' together require the transcript to be pasteable, but no document names a shell. Chose POSIX sh single-quoting, with an embedded quote closed-escaped-reopened; measured output for the argument "retro: it's a; rm -rf / $(echo hi)" is 'retro: it'\\''s a; rm -rf / $(echo hi)' as one word. A Windows reader adapts the transcript, which is recorded in the blueprint rather than left implicit. Separately, 'a digest of each stream' names no algorithm; chose sha256 rendered as 'sha256:<hex>' so the algorithm travels with the value and a later change is visible in the record rather than silent. Test digests were computed independently with sha256sum, not by calling the same function under test: printf 'value-1\\n' | sha256sum -> bb849f0f...93f2b, which is what TestEachStepRecordsItsInvocationStatusDurationAndStreamDigests asserts.
```

**agent-plugins-ih18.2#5**

```
retro:gap Two of the PROC delta's seven behaviours are verified by tests that can cite no acceptance criterion, because none exists for them.
evidence: the bead names three ACs (AC-SWARM-CORE-001.10, .11, .12) and the blueprint's AC Coverage table maps 'PROC (proves)' to exactly those three. But the delta also states 'The Process Runner SHALL invoke an external tool as an argv vector' / 'SHALL NOT construct a command string for a shell interpreter' and 'SHALL record each invocation's elapsed time', and no AC in specs/features/swarm-core/swarm-core-frd.md covers either — grep for 'elapsed' in the FRD returns nothing in a Gherkin block. TestAnArgumentThatLooksLikeShellSyntaxIsPassedThroughLiterally and TestEachInvocationRecordsItsElapsedTime therefore cite the delta in prose rather than an 'AC:' marker, which means ac-traceability-standards' greppable citation does not reach them. Not a defect in this bead; a note for whoever reconciles the coverage table at merge, since 'every AC is cited' will pass while two specified behaviours sit outside the citation graph.
```

**agent-plugins-ih18.1#6**

```
retro:gap 'Print its command tree' has no content at this delta: the tree holds no subcommands, so cobra renders a usage block with a Usage line and a Flags section and no 'Available Commands' section at all.
evidence: measured output of the built binary with no arguments, in full: 'Usage:\n  swarm [flags]\n\nFlags:\n  -h, --help   help for swarm'. A test asserting a literal expected string here would pin the empty tree as the answer and start failing on the first delta that adds a subcommand. cli/swarm/cmd/root_test.go therefore derives its expectation from newRootCmd(...).UsageString() instead, so it keeps asserting 'the tree is printed' as the tree grows. One wrinkle worth carrying: cobra registers --help during its own Execute rather than at construction, so an un-executed command's UsageString() has no Flags section and the two renderings differ until root.InitDefaultHelpFlag() is called.
```

**agent-plugins-ih18.24#6**

```
retro:gap Filed agent-plugins-atuh (discovered-from this bead, not blocking) while measuring the human-queue gate: references/run-record.md line 81 documents the bounce ceremony as 'bd create --add-label human' with '(--labels exits 1 with unknown flag)'; measured on bd 1.2.2 (bd version: '1.2.2 (6c124203e: HEAD@6c124203e771)'), the opposite holds for 'bd create' — '--add-label' is an unknown flag on create ('Error: unknown flag: --add-label') and '--labels human' is the one that succeeds. '--add-label' is real but only on 'bd update' ('bd update --help' lists it; 'bd create --help' does not). The same wrong example also appears in docs/research/swarm-cli/2026-08-25-capability-audit.md line 146. Not fixed inline — it is run-record.md's bounce-channel documentation, not this bead's three gates, and touching a load-bearing shared reference outside this bead's scope seemed likelier to collide with concurrent dispatches than to help.
```

**agent-plugins-ih18.24#7**

```
retro:gap This bead's metadata.c4_changes reads ["Environment Probe"] (set by grooming/planning, not by this dispatch — left as-is per this run's standing rule that a field a dispatch did not set is not its to clear, just to report). The Environment Probe (plugin/skills/bead-bonanza/scripts/pre-flight-check.sh, per specs/architecture/module-bead-bonanza.md) runs twice per run, not at every re-query, which is exactly the property this bead's own AC says the fix must NOT have ('gates run at every re-query, not once at plan time'). The actual change landed on the Coordinator component instead (module-bead-bonanza.md's Coordinator bullet list), which is where 'runs at every re-query' already lives in that document. Flagging the mismatch rather than editing the metadata.
```

**agent-plugins-ih18.19#2**

```
retro:gap A DoD item that only an unreachable CI run can discharge has no path to being ticked by the bead that owns it, and nothing in the item says who ticks it later.
evidence: The blueprint's item read '[ ] The Windows leg has run at least once against the tracker install adr-027 targets, and its result is recorded.' The leg exists — .github/workflows/cli.yml, matrix 'os: [ubuntu-latest, macos-latest, windows-latest]' with dolt and bd install steps under 'if: matrix.os == \'windows-latest\'' and SWARM_E2E_REQUIRE_TRACKER=1 on that leg. It has never run: the workflow's triggers are 'on.push.branches: [main]', 'on.pull_request' and 'workflow_dispatch', and this branch has neither merged nor opened a PR. Measured through the GitHub API, workflow cli.yml filtered to claude/bead-bonanza-agent-plugins-ih18-2yjc8j: total_count 0. agent-plugins-ih18.18's own retro comment reached the same conclusion from its side and named 'ih18.19's job, or the coordinator at merge' — but a DoD item is verified by the bead that owns the DoD, which runs before the merge, so 'or the coordinator at merge' is the half that has no bead. Filed as agent-plugins-6jnx. An item whose evidence arrives after the run should say so when it is written, and name the ceremony that collects it.
```

### `scope`

**agent-plugins-ih18.13#5**

```
retro:scope The delta names four skill documents as anchors, but "every place a skill reads or writes the run manifest by hand" reaches ten skill files plus one architecture document. Everything found was adopted; the one writer left out is a script, filed as its own bead.
evidence: grep for run-epic metadata access found hand reads/writes in plugin/skills/run-batch/SKILL.md, sprint-review/SKILL.md, sprint-plan/SKILL.md, sprint-plan/references/partitioning.md, groom-beads/references/epic-definition-of-ready.md (and its generated copy under sprint-plan/references/shared/), bead-bonanza/SKILL.md, bead-bonanza/references/run-record.md, bead-bonanza/references/worktree-mechanics.md, sprint-recovery/SKILL.md, sprint-retro/SKILL.md and sprint-scope/SKILL.md. Anchors named only four of them. Bead metadata reads (`metadata.c4_changes`, `metadata.tier`, `--set-metadata collision=`) were left alone: they are not the run manifest.
```

**agent-plugins-ih18.9#4**

```
retro:scope manifest.Require was implemented here although AC-SWARM-CORE-002.18 is the swarm-manifest command item's to prove.
evidence: this delta's own Design ends 'A precondition names every absent key', and the rule is a property of the record's model rather than of a command surface — Require needs the declared key set, which is here. The command item will wire it to 'swarm manifest require' and owns the AC citation at that level; this item's test cites 002.18 as well, which ac-traceability-standards permits ('an acceptance criterion MAY be cited by multiple tests') and which means the mechanism is covered even if the command's own e2e case is deferred. Also tested here: requiring an UNDECLARED key is itself a failure, because a caller requiring a misspelt key would otherwise wait forever for something that can never be written.
```

**agent-plugins-ih18.7#4**

```
retro:scope The bead's anchors name 'schemas/ beside the existing .strata sources' and 'cli/core/schema/ for generated output'. Taken literally the first would have been wrong, so it was not taken literally.
evidence: scripts/generate-schemas.sh builds the whole schemas/ directory as ONE strata package. Two .strata files added beside the existing ones would have joined the SpecModel bundle, landed swarm types in cli/core/specmodel (the document corpus's model package), and added swarm pages to plugin/skills/spec/references/generated (the spec skill's authoring reference) — three places none of them belongs. So schemas/swarm/ carries its own strata.config.json, which per strata's package rule ('the nearest ancestor containing strata.config.json is the package root ... except subtrees carved out by their own manifest') makes it a package of its own. Verified rather than assumed: after adding it, 'strata build schemas --target json-schema' produced a SpecModel bundle with zero occurrences of RunManifest or CommandResult.
The second anchor WAS taken literally, because contract-swarm-json.md's own artifact table names cli/core/schema as the validating artifact's home — so the JSON Schema goes there (swarm-model.schema.json, embedded beside the other two) while the Go model goes to cli/swarm/swarmmodel, which no document had named because none had needed to.
```

**agent-plugins-ih18.7#5**

```
retro:scope Rewrote cli/swarm/cli/result.go to emit the GENERATED envelope, replacing the hand-written envelope/factJSON/invocationJSON structs the Command Surface item shipped two commits earlier in this batch. That is a change to another delta's component and it is deliberate.
evidence: ADR-016 makes Strata the one source for schemas and generated models. A schema declared here beside a hand-written Go twin over there is two declarations of one shape — the drift this delta exists to end — and it would have shipped a contract nothing conforms to. The note I left on this bead while working the Command Surface item named exactly this reconciliation as this item's to do.
What it cost: cli.Fact's 'Value any' became a typed *swarmmodel.FactValue, so facts are now built with Text/Count/Flag/List/Unestablished rather than a single Established(name, value any). Four call sites in the surface's own tests changed. What it bought: TestWhatTheSurfaceEmitsConformsToTheEnvelopeSchema compiles the generated JSON Schema with santhosh-tekuri/jsonschema (already a module dependency) and validates what the surface actually prints, in all three modes and on a refusal — so the schema and the emitter cannot drift without a test failing. Its negative control, TestTheEnvelopeSchemaRejectsWhatItShould, checks the validator rejects an unknown key, an unknown outcome, a status outside 0..3 and a missing field, so a validator that accepted anything could not make the first test pass.
```

**agent-plugins-ih18.6#4**

```
retro:scope The interface carries seven operations, of which the delta's five behaviour bullets require two. The other five come from the container blueprint's roster line for this component — 'history and commit subjects, branches and their ancestry, worktree registrations, merges, and the comparison of a local branch against its remote' — and each maps to exactly one measured invocation.
evidence: Worktrees and Commits are what the Status Collector (agent-plugins-ih18.14) needs for the standing-worktree count and the last-commit age. Branches, Descends, Resolve, Divergence and Merge are the roster's remaining nouns. Merge in particular is there because the incident behind this whole feature is a worktree removed after a merge exited 128, and the merge step declaring OnFailure=Stop is what makes the guard real rather than described — TestEachStepDeclaresItsEffectAndWhetherReRunningIsSafe asserts it.
The risk I accepted: five operations with no caller yet, which a later delta may find shaped wrong. The alternative was an interface that satisfies its own delta and not the roster it belongs to, leaving the next delta to add the missing operations one at a time against a component blueprint that already listed them.
```

**agent-plugins-ih18.4#6**

```
retro:scope Exported plan.Argv from the Plan Runner package (a one-line wrapper over its existing renderResolved) so the surface renders an invocation exactly as the transcript does.
evidence: this bead's metadata already declares c4_impacts: ["Plan Runner"], so touching it is anticipated. The alternative was a second POSIX-quoting implementation in cli/swarm/cli, which is two conventions to keep in step for one thing — and AC-SWARM-CORE-001.8 ('every reported line names the command a reader can re-run') would then be satisfiable by a rendering the transcript does not produce. No behaviour of the plan package changed; the export is additive.
```

**agent-plugins-ih18.1#3**

```
retro:scope The delta's Behavior list names three obligations (build as a second binary in the existing module; cross-compile for every release target; bare invocation prints the tree and exits malformed). I also extended .github/workflows/cli.yml's no-runtime job to swarm — static-linkage assertion plus a busybox run — which is not on that list.
evidence: specs/architecture/container-swarm-cli.md, Infrastructure and Deployment: 'Its only runtime requirements are bd and git on PATH; it needs no runtime, interpreter or package manager of its own.' Nothing proved that for swarm; the job that proves it for spec is three lines away in the same file and was already building only ./spec. ADR-023's stated negative consequence is the same shape: 'The cross-compile leg and the release workflow must build a second target; a build that silently kept compiling only ./spec would leave the Windows guarantee unproven for this binary while CI stayed green.' Rejected alternative: leave the job spec-only and file a follow-up bead — which would leave a container-blueprint claim unproven across the whole feature for the sake of a smaller diff in one workflow file the delta's Anchors already name.
```

**agent-plugins-ih18.23#4**

```
retro:scope Filed agent-plugins-nqe8 (discovered-from this bead): worktree-mechanics.md's Merging and worktree-removal sections still show literal bb/<dispatch-id> in the merge/ancestry/removal commands, unchanged. Left unresolved here because the coordinator's decision comment scoped correction to the creation-step description and the attribution mechanism only, and whether the real merge step resolves the branch dynamically or hardcodes it was not established.
```

**agent-plugins-ih18.19#5**

```
retro:scope The bead's title and Design tell it to run the retro; the dispatch told it not to. Followed the dispatch and ran none.
evidence: Title: 'Verify the Definition of Done, retire the blueprint, run the retro'. Design: 'Then run the retro-spec skill over the implementation phase, aggregating the findings recorded on this feature's items — the approved plan is the standing authorization for that retro.' The dispatch prompt: 'Do not run the retro. ... this run has a separate retro ceremony (sprint-retro) that the coordinator dispatches afterwards, over the whole run.' Both are right about their own scope and they do not know about each other: groom-feature-blueprint plans a per-feature retro-spec pass into the feature's last work order, and bead-bonanza runs one sprint-retro over the run. When a feature is delivered AS a sprint the two collide, and the collision surfaces in the one bead that carries both. Nothing here decides which should win — but the plan-time item should say 'unless the run's own ceremony covers it', or the sprint should retire the item when it admits the feature.
```

### `stale`

**agent-plugins-ih18.13#3**

```
retro:stale run-batch/SKILL.md told the agent its checks baseline arrives in the dispatch prompt; the prompt has carried only three values since dispatching.md was rewritten, and the baseline comes off the manifest. Corrected in the same edit.
evidence: the removed sentence was "Your prompt carries a `Checks baseline:` block — what the checks printed on the untouched tree, and the status they exited with." plugin/skills/bead-bonanza/references/dispatching.md, "The prompt is three values": "Everything a prompt used to carry — the branch, the checks baseline, the check ids, the trailers, the worktree path ... is one query away for the agent and was a hand-copied second copy here." My own dispatch prompt carried exactly the three values and no baseline.
```

**agent-plugins-ih18.11#2**

```
retro:stale:project ADR-027 asserted that a restored fixture's tracker timestamps are fresh on every run and must be normalized. Measured false; corrected in this bead.
evidence: the claim, specs/adrs/ADR-027-two-test-tiers-and-a-committed-fixture.md, Consequences: 'The tracker half restores through the tracker's own import rather than being copied byte for byte, so its internal timestamps are fresh on every run and must be normalized even though the git shas need not be.'
Measurement (bd 1.2.2, scratch git repo, no remote, pinned HOME): git init; bd init --prefix fixture; bd import cli/testdata/swarm-fixtures/core/tracker.jsonl; then bd show fixture-run --json and bd show fixture-run.2 --json.
  export holds:  fixture-run 2026-01-01T00:00:00Z 2026-01-01T00:00:00Z
  after import:  fixture-run created_at= 2026-01-01T00:00:00Z updated_at= 2026-01-01T00:00:00Z status= open
                 fixture-run.2 created_at= 2026-01-01T00:00:00Z updated_at= 2026-01-01T00:00:00Z status= in_progress
Import preserves created_at, updated_at, ids and status verbatim. So the tracker half is as fixed as the git shas and neither needs normalizing.
What DOES need normalizing is what the surface measures rather than what the fixture holds: step duration and stream digests. Both cases declare exactly those two substitutions ([0-9]+ms -> <duration>, sha256:[0-9a-f]{64} -> sha256:<digest>) and no timestamp substitution, and both pass.
Sweep: grep -rn 'internal timestamps|timestamps are fresh|normalized even though' over specs/, docs/ and plugin/ returned that one bullet and nothing else; no bead, epic comment, trigger prompt or generated copy carried it.
```

**agent-plugins-ih18.11#4**

```
retro:stale:project the container blueprint's verification gate named a test that never existed, and an anchor pointing at the wrong harness. Both corrected in this bead.
evidence: specs/architecture/container-swarm-cli.md read 'cd cli && go test ./cmd/... -run TestSwarmSnapshotCases for the end-to-end cases against the committed fixture'. There is no TestSwarmSnapshotCases anywhere in the module and cli/cmd holds no swarm case: grep -rn TestSwarmSnapshotCases cli/ returns nothing. The tier as built is TestSwarmCases in package e2e at cli/swarm/e2e, so the gate is 'cd cli && go test ./swarm/e2e/'.
specs/features/swarm-core/swarm-core-blueprint.md anchored the harness at 'cli/cmd/ for the existing snapshot harness the case runner extends'. It extends nothing there: cli/cmd's harness drives a corpus in place through the spec binary's own entry point, where this one restores a git bundle and a tracker export into a temporary world per case and runs a COMPILED BINARY as a subprocess. They share the idea and no code, so the harness is its own package.
Both were plausible-looking predictions written before the code existed; neither would fail any check, because nothing verifies that a named gate command resolves.
```

**agent-plugins-ih18.2#6**

```
retro:stale ADR-025 states that container-swarm-cli carries the Windows .cmd-shim hazard as an open question. It does not — that blueprint's Open Questions section reads 'None.' Filed as agent-plugins-ih18.20 with a discovered-from edge to this bead rather than fixed here.
evidence: specs/adrs/ADR-025-argv-not-shell.md:101 'will land. `container-swarm-cli` carries it as an open question.'; specs/architecture/container-swarm-cli.md:523-525 '## Open Questions' / (blank) / 'None.'; 'grep -rn "\.cmd\b|shim" specs/' returns exactly one hit, ADR-025:98, so the hazard is recorded nowhere else in the corpus. Not this bead's to close: the runner resolves executables through os/exec's own rules and its callers are bd and git, neither an npm shim, so nothing in this delta forces the decision.
```

**agent-plugins-ih18.19#1**

```
retro:stale The epic's own acceptance criterion counts work orders, so it went stale the first time the run admitted one.
evidence: agent-plugins-ih18's ACCEPTANCE CRITERIA reads '[ ] Every one of the 19 work orders under this epic is closed.' Measured: 'bd list --parent agent-plugins-ih18 --status all --json' returns 25 children — 23 closed, 1 deferred (.22), 1 in_progress (.19). Five were filed mid-run (.20 .21 .23 .24 .25) and one excluded by decision (.22). A criterion phrased as a literal count cannot survive emergent work, which the run's own dispatch policy expects; phrased as 'every admitted work order, per the manifest's admission record', it would have stayed true and stayed checkable — and 'swarm status --run <epic>' already reports exactly that (completion_done 24, scope_excluded agent-plugins-ih18.22).
```

**agent-plugins-ih18.19#4**

```
retro:stale Three claims in the two owning architecture documents were false at the moment the DoD checked them, and each was left behind by a change made elsewhere in this same run.
evidence: (1) specs/architecture/container-swarm-cli.md's '## Data' section named the run manifest and the repository and NOT the fixture, while the DoD item explicitly reads 'and Data names the fixture'; 'grep -n fixture specs/architecture/container-swarm-cli.md' put every hit in Infrastructure and Change Guidance, none in Data — though cli/testdata/swarm-fixtures/ holds three (core, status-attention, status-narrowed-scope), each with repository.bundle, tracker.jsonl and fixture.toml. (2) The same file's References read '[[module-bead-bonanza]] — ... which names it as a planned dependency', after the ADOPT-MAN and ADOPT-STAT deltas removed 'planned' from the module's own Consumes note — the DoD listed the module's half and not the container's back-reference, so the sweep stopped one document short. (3) module-bead-bonanza.md's Open Questions comment said 'the swarm CLI is a CONTAINER that will carry its own blueprint once its FRD scopes it', which container-swarm-cli.md has falsified since the SKEL delta. All three fixed here. The pattern: a DoD that enumerates the documents to fold a delta INTO does not enumerate the documents that POINT AT them, and a pointer is where the stale copy survives.
```

### `sequencing`

**agent-plugins-ih18.16#1**

```
retro:sequencing The REP delta this one depends on is CLOSED in the tracker but its commit is not in my base tree, so `swarm status` did not exist in the worktree I adopted it from.
evidence: `bd show agent-plugins-ih18.15` -> `[P1 . CLOSED]`, assignee dev-bd97-5. `git log --oneline --all | grep ih18.15` -> `9d9a77e agent-plugins-ih18.15 P1 Status Reporter and swarm status...`; `git branch -a --contains 9d9a77e` -> `worktree-agent-ab2362ea36191e1b1` only (a sibling dispatch's tree), not the sprint branch. My base: `git rev-parse --short HEAD` -> `9b67309`, and `git merge-base --is-ancestor $(git rev-parse claude/bead-bonanza-agent-plugins-ih18-2yjc8j) HEAD` -> exit 0, so my base IS the sprint branch and the sprint branch does not carry REP. The blueprint's Stage 7 says "The ADOPT-STAT delta SHALL land after the REP delta" (specs/features/swarm-core/swarm-core-blueprint.md); "landed" was true of the bead and false of the tree I had to write against. run-batch's per-bead ceremony has no step that checks a dependency's commit is in the base ref, only that the base ref is the sprint branch.
```

**agent-plugins-ih18.13#1**

```
retro:sequencing `swarm manifest` cannot serve the `trailers` key, so ADOPT-MAN could not be completed for it and run-batch keeps a workaround; filed as a discovered-from bug against the epic.
evidence: `swarm manifest get --run agent-plugins-bd97 --key trailers` -> "summary: trailers: not a declared manifest key / outcome: malformed (exit 3)". Same for `require`. `swarm manifest set --run agent-plugins-bd97 --key trailers --value ${q}"x"${q} --dry-run` -> "manifest: refusing the write, and storing none of it — \"trailers\" is not a declared manifest key", exit 3. Cause read from cli/swarm/manifest/manifest.go: buildKeys() enumerates contract.StageDeclarations() only, and swarmmodel.RunManifest.Trailers is declared on RunManifest rather than in a stage struct. specs/architecture/container-swarm-cli.md ("A key whose absence is legitimate at every stage has no stage") intends it to be absent from VALIDATE findings, not to be unreadable.
```

**agent-plugins-ih18.10#1**

```
retro:sequencing Two of this item's four ACs are declared e2e and are NOT covered here, because the fixture they need is the next item.
evidence: the AC Coverage table gives this delta AC-SWARM-CORE-001.6 (e2e) and AC-SWARM-CORE-002.16 (e2e), and the blueprint's Sequencing Constraints say 'Stage 4 — the fixture. The FIX delta SHALL land before any end-to-end case in any other delta. WHILE the FIX delta is outstanding, a command delta's acceptance rests on its unit tests, and its end-to-end case is added when the fixture exists.' So both are covered here at the unit level against a faked process runner — TestADryRunOfSetReadsAndWritesNothing for 001.6 and TestValidateIsCleanForARecordThatSatisfiesEveryRule for 002.16 — and the e2e cases against the committed fixture belong to agent-plugins-ih18.11. The package doc of cli/swarm/cli/manifest_test.go says so where a reader will find it, rather than leaving the gap to be discovered from the coverage table.
```

**agent-plugins-ih18.4#2**

```
retro:sequencing SURF must emit an envelope that specs/contracts/contract-swarm-json.md says is fixed by schemas/swarm-output.strata — a file the SCHEMA item has not created yet, and no sequencing constraint puts SCHEMA before SURF.
evidence: the blueprint's Sequencing Constraints name 'The SCHEMA delta SHALL land before the MAN and MANV deltas' and nothing about SURF; the AC Coverage row for AC-SWARM-CORE-001.9 is 'SCHEMA, **SURF (proves)**', so SURF proves the same-facts property while SCHEMA supplies the shape. ADR-016 makes Strata the one source for generated models, so the hand-written Go envelope in cli/swarm/cli/result.go is a second description of a shape that is meant to have exactly one. Not resolvable here — writing the .strata file would be the SCHEMA item's whole delta. Recorded on agent-plugins-ih18.7 as a note naming the file, the structs and every field the envelope carries today, so that item reconciles against something real rather than re-deriving it.
```

**agent-plugins-ih18.3#6**

```
retro:sequencing The SURF delta's anchors name 'cli/swarm/main.go and cli/swarm/cli/' for the command surface, but the SKEL delta already put it at cli/swarm/cmd/. Noticed while reading the blueprint for this bead; it belongs to the Command Surface item, which is next in this batch.
evidence: specs/features/swarm-core/swarm-core-blueprint.md, the 'Swarm Command Surface in [[container-swarm-cli]] (added)' block: '**Anchors:** `cli/swarm/main.go` and `cli/swarm/cli/`; `github.com/spf13/cobra`, already a dependency of the module.' The SKEL bead's anchors said only 'cli/swarm/ for the new package tree; cli/cmd/ for the existing command-surface conventions to follow', which names cli/cmd as a CONVENTION to mirror and no destination — so mirroring it as cli/swarm/cmd/ was the reading available at the time, and the conflicting anchor is two deltas away in a block that bead did not have to read. Resolution belongs to the Command Surface item: either rename the package to cli/swarm/cli or correct the blueprint's anchor, and say which.
```

**agent-plugins-ih18.24#1**

```
retro:sequencing This bead's Anchors and Design assume bb-status.sh's run mode is ALREADY retired (per ih18.16, discovered-from parent), but ih18.16's retirement commit (cfdcb5d, 409 lines removed from bb-status.sh) has not reached this dispatch's branch (dev-bd97-5, cut before ih18.16 landed on the coordinator's sprint branch) -- no bd dep edge encodes this real prerequisite, only the discovered-from provenance link, which run-batch's queue resolution does not treat as a blocker. Took the Blocked path rather than editing code that is about to be deleted upstream (a guaranteed rebase conflict) or hand-splicing the coordinator's not-yet-merged commit into this branch myself.
```

### `unverifiable`

**agent-plugins-ih18.1#5**

```
retro:unverifiable The new 'Run swarm in an image containing only a shell and the binary' step in .github/workflows/cli.yml could not be run here: this container has the docker client but no daemon.
evidence: 'command -v docker' -> /usr/bin/docker; running the step's own command, 'docker run --rm -v $PWD/swarm-static:/swarm:ro busybox:latest /swarm', printed 'failed to connect to the docker API at unix:///var/run/docker.sock; check if the path is correct and if the daemon is running: dial unix /var/run/docker.sock: connect: no such file or directory' and exited 1. So what IS measured is the static linkage (file -b, above) and the exit status of the same binary run natively (3, stdout empty). What is NOT measured is that a busybox image can execute it. Settled by the first run of the no-runtime job on a GitHub runner, which has a daemon; the step asserts status 3 explicitly and fails loudly on anything else, so a false assumption there shows up as a red job rather than silently.
```

**agent-plugins-ih18.18#2**

```
retro:unverifiable This sandbox has no real Windows runner and no push/merge access to trigger a real GitHub Actions run, so the .github/workflows/cli.yml Windows leg (dolt+bd install steps) could not be observed green end to end -- only its individual pieces were verified independently: the dolt download URL (curl -sI -L returned 200, redirecting to a real dolthub/dolt v2.3.1 release asset), the bd install command (matches both beads' own install.sh CGO_ENABLED=0 fallback and install.ps1's only path, byte for byte on the go install invocation), and the server auto-start mechanism (measured directly, see the other retro comment) -- but never all three together on windows-latest. Settles blueprint DoD item 'The Windows leg has run at least once ... and its result is recorded' only partially: the leg now EXISTS and is well-evidenced piece by piece; whether it is GREEN is for the first real CI run of this branch to show, and that observation still needs recording by whoever watches it (ih18.19's job, or the coordinator at merge).
```

### `tooling`

**agent-plugins-ih18.21#3**

```
retro:tooling:skill run-batch/SKILL.md had no room for the correction the bead asked for, so the command form landed in the other two documents and not in the skill
evidence:
The bead's Done-when asks that plugin/skills/run-batch/SKILL.md be "updated to
the command form". The correction is in — the false claim "`trailers` is the one
key `swarm manifest` does not serve, so do not ask it" is gone, replaced by
"The trailers live in the run's manifest and the script fetches them" — but the
command itself, `swarm manifest get --run <epic> --key trailers`, is NOT named
there. It is named in plugin/skills/bead-bonanza/references/run-record.md and in
specs/architecture/container-swarm-cli.md instead.

Why: `scripts/check-skills.sh` sets BODY_BYTE_MAX = 20000 (line 43) and warns
at line 131 — `--strict` turns that warning into a failure. Measured:

  at 90cd22a   run-batch/SKILL.md = 20758 bytes; body 19990 (frontmatter 768)
  first draft  file 20860 bytes; check reported "body is 20092 bytes"
               => skills-strict RESULT: FAIL, exit 1, against a baseline of PASS
  as landed    file 20707 bytes; skills-strict RESULT: PASS, 0 warnings

So the file had TEN bytes of headroom before this bead touched it. Every wording
that both dropped the false claim and named the command measured between 20004
and 20092 bytes of body — over the limit in each case. The choice was accuracy
in the skill versus a green gate, and the gate won for this bead.

Filed rather than fixed here: agent-plugins-tnyu (discovered-from
agent-plugins-ih18.21) — give run-batch/SKILL.md real headroom under
BODY_BYTE_MAX. Not this bead's to close; widening it into a prose-trimming pass
over a 20 KB skill file would have put unrelated edits in a bug fix's commit.

The general shape, for whoever reads this in the retro: a fixed byte budget on a
skill file makes a CORRECTNESS edit compete with existing prose, and the loser is
whichever the author judges least costly to drop. That is a bad trade to leave
implicit in a check.

```

**agent-plugins-bd97.3#1**

```
retro:tooling:groom-beads — bd-health.sh's exit status cannot gate one epic. Ran 'bash .claude/skills/groom-beads/scripts/bd-health.sh' -> exit 1, 'RESULT: hard problems found in: swarm validate'; every named id is foreign to this run (agent-plugins-qozp.6, agent-plugins-so39.2 ... ) and every error reads "is not an epic or molecule (type: task)". Ran 'bd swarm validate agent-plugins-ih18' -> exit 0, 'Swarmable: YES', 19 issues, max parallelism 3, 12 waves. Expected: a non-zero exit means this queue's graph is unsound. Observed: it means some unrelated bead in the repo is not an epic. SKILL.md says 'Do not declare the set ready while it exits non-zero.'
```

**agent-plugins-bd97.3#2**

```
retro:tooling:sprint-plan — partitioning.md's sequencing jq reads c4_impacts in the opposite direction from how this repo's grooming writes it. The reference says 'Bead B whose c4_impacts intersects bead A's c4_changes consumes a contract A is changing'; groom-beads' SKILL.md defines impacts as 'downstream components affected but not edited'. Ran the reference's jq over the 19 queued beads: it emitted 'sequence agent-plugins-ih18.2 after agent-plugins-ih18.6: consumes Git Version Control' and 'sequence agent-plugins-ih18.2 after agent-plugins-ih18.5: consumes Beads Tracker'. Expected: PROC (.2) is the producer both are built on and the blueprint's Stage 2 says 'The PROC delta SHALL land before the TRK and VCS deltas'. Observed: the jq orders PROC last. 12 of the 33 sequence lines it emitted are inverted this way.
```

**agent-plugins-bd97.3#3**

```
retro:tooling:sprint-plan — partitioning.md names c4_changes (layer 1) and description file paths (layer 2) as the collision signals, and never names c4_docs. In this queue the dominant collision is c4_docs: 15 of 19 beads carry c4_docs=[specs/architecture/container-swarm-cli.md] and every one carries the acceptance line 'This delta is folded into its owning blueprint (roster, diagram or the section the delta is written in the vocabulary of)'. Ran the reference's co-edit jq: it returned 8 pairs, on Swarm Command Surface, Manifest and Coordinator, and none on that document. Expected: the signal that 15 beads all append to one markdown roster. Observed: silence from both declared layers — c4_changes because the fold is not a component edit, file paths because the path is in acceptance_criteria, not in the description the layer-2 scan reads.
```

**agent-plugins-bd97#1**

```
retro:tooling:bead-bonanza — the tier-split rule in dispatch-policy.md was violated by all three of the plan's batches and by the coordinator's first dispatch; caught by a human asking, not by any check. Filed as agent-plugins-cf5j.
```

**agent-plugins-bd97#2**

```
retro:tooling:sprint-plan — component-granular collision avoidance caught none of this run's real contention (8 C4 pairs, all already dependency-ordered) while an unmodelled signal, c4_docs, caught all of it: 15 of 19 beads append to one blueprint. Serializing it cost 5 waves. Filed as agent-plugins-221f.
```

**agent-plugins-bd97#3**

```
retro:tooling:sprint-plan — the plan reported '17 waves at max parallelism 2' accurately, and the run still provisioned four dispatch agents for a graph that admits one at a time (14 of 17 waves are width 1; dev-bd97-4 and -5 had no ready bead by wave 11). The number was right and the shape was invisible. Filed as agent-plugins-tvom: emit the wave decomposition as an ASCII graph at plan time, when the edges are still cheap to change.
```

**agent-plugins-bd97#4**

```
retro:tooling:bead-bonanza — bb-status.sh run's health evidence line reported 'last commit 73 minutes ago' while the live batch had committed 13 minutes ago and six beads in between; the age is read from the coordinator's HEAD, which only moves on merge. Verdict was still correct ('working'), the evidence was not. Filed as a bug.
```

**agent-plugins-bd97#5**

```
retro:tooling:run-batch — a dispatch agent has no way to distinguish a coordinator's deliberate triage write from a tool anomaly, because a bead records no provenance for assignee or defer. dev-bd97-4 reverted an admission and an exclusion this way, both silently and both at exit 0. Filed as agent-plugins-hjyq. NOTE for the retro: dev-bd97-4's reported 'bd create anomaly' is explained by this and is not evidence against bd.
```

**agent-plugins-bd97#6**

```
retro:tooling:run-batch — both light-tier beads in this run (.12, .17, both dev-bd97-2) closed with zero retro: comments and were caught by the acceptance gate's signal 5; all 17 non-light beads recorded 4-7. Both returns reasoned that straightforward work needs no findings, which is exactly what the Retro section pre-empts — the instruction to record the ABSENCE is what was missed. Filed as agent-plugins-tgex, with the n=2 / same-dispatch-id caveat stated on the bead.
```

**agent-plugins-ih18.16#4**

```
retro:tooling This run's durable check-in trigger hands a resuming agent the command this bead retires, and nothing in the ceremony sweeps a trigger prompt when a command it names is retired.
evidence: `list_triggers` returns trig_018Nudewp4Y9yBZ9BmZVxdu3 "bd97 sprint check-in", cron `18 * * * *`, enabled, next_run_at 2026-08-29T05:18:00Z, bound to persistent_session_id session_01VWAd7X9pHDZeEsucbqAWQA. Its prompt reads: "Re-derive the state from the tracker and from git:\n\n  /home/user/agent-plugins/plugin/skills/bead-bonanza/scripts/bb-status.sh run agent-plugins-bd97". After this commit merges that path exits 3. I did NOT edit the trigger: it is the coordinator's, and the replacement `swarm status --run` does not work in this container yet (`/usr/local/bin/swarm status --help` -> `Error: unknown command "status" for "swarm"`), so repointing it now would break it earlier than leaving it. Mitigation actually shipped: the retired `run` case dies with "the `run` mode is retired; use: swarm status --run <epic-id>", so a woken agent gets a self-correcting exit 3 rather than a wrong answer — asserted by the script's own self-test ("and it names its replacement"). Unreached copy, reported per AGENTS.md's retraction sweep: the coordinator owns repointing that prompt after the merge.
```

**agent-plugins-ih18.15#3**

```
retro:tooling status.Collect always runs its sub-plans in plan.ModeExecute internally, and its completion walk is unbounded-depth (one bd list query per level, depth unknown ahead of time) — so swarm status cannot fit the Spec.Plan()+Interpret() shape every other subcommand (cli/swarm/cli/manifest.go) uses: that shape assumes ONE static plan.Runner.Run() call whose report both explain-mode and execute-mode interpret. cli/swarm/cli/status.go is wired as a raw cobra.Command instead, declaring the four contract flags (explain/dry-run/json/run) by hand and reusing finish()/refusal()/Refuse*() so the exit-status mapping and both renderings stay the surface's. Documented as a deliberate exception in specs/architecture/container-swarm-cli.md's Swarm Command Surface section and the Components diagram (Rel(surface, report, ...)). No FRD or blueprint text flagged this incompatibility ahead of time; a future command whose measurement depends on an unbounded walk will hit the same wall.
```

**agent-plugins-ih18.13#4**

```
retro:tooling The skills body-size warning is a real budget the anchors did not mention: adopting the command in run-batch/SKILL.md pushed it past the ~20000-byte body limit, and three rounds of trimming were needed to get back under. A doc-editing bead can be made to fail its gate by prose it was told to add.
evidence: after the first pass, `bb-checks.py skills` printed "WARN plugin/skills/run-batch/SKILL.md: body is 20481 bytes (~5120 tokens); above ~20000 the tail is truncated from context after a compaction" and "--- 0 error(s), 1 warning(s) ---", against a baseline of "0 error(s), 0 warning(s)". The pre-change body was under the limit only by ~400 bytes (file 20373 bytes on HEAD). Final state after trimming: "--- 0 error(s), 0 warning(s) ---".
```

**agent-plugins-ih18.13#7**

```
retro:tooling Three `bd create` calls with the same flag shape produced three different results for fields nobody passed: one bead came back assigned to $BEADS_ACTOR, one came back DEFERRED for a month, one came back clean. A bead filed as out-of-scope work landing in the filing agent`s own ready queue, or hidden from `bd ready` entirely, defeats the purpose of filing it.
evidence: all three created with `bd create --title ... --type=... --priority=... --parent agent-plugins-ih18 --deps discovered-from:<id> --body-file <path>` and BEADS_ACTOR=dev-bd97-4; no --assignee and no --defer on any of them. Observed via `bd show <id> --json | jq .[0]`: ih18.21 -> assignee "dev-bd97-4", status open; ih18.22 -> assignee null, status "deferred", `bd show` header "Deferred: 2026-09-28"; ih18.23 -> assignee null, status open. ih18.21 then appeared in `bd ready --parent agent-plugins-bd97 --assignee dev-bd97-4 --exclude-type=epic` as if it had been dispatched to me. Repaired by hand: `bd update ih18.22 --defer ""` -> status open, defer_until null; `bd update ih18.21 --assignee ""` -> assignee null. Cause not established; bd 1.2.x, this container.
```

**agent-plugins-ih18.7#2**

```
retro:tooling:project strata emits a JSON Schema for a discriminated union that REJECTS the very documents its own Go target produces, when the package sets allowAdditionalProperties:false. Measured, then designed around.
evidence: with '@discriminator("kind") union FactValue { @tag("text") TextValue ... }', the generated bundle emits FactValue as allOf[ if kind==text then $ref TextValue ] while TextValue itself is {additionalProperties:false, properties:{text}} — no 'kind'. The Go marshaller emits {"kind":"text","text":"x"}. Validating that against the generated FactValue: 'at \'\': additional properties \'kind\' not allowed', for all four variants. So the two targets of one source disagree, which is the failure ADR-016 adopts strata to prevent.
Worked around rather than fixed: FactValue is now a plain struct with 'kind: FactKind' and four optional typed fields, which produces the IDENTICAL wire form ({"kind":"text","text":"x"}) and validates. The escape hatches I did not take: allowAdditionalProperties:true would surrender the closed key set that gives AC-SWARM-CORE-002.2 and 002.9 by construction; the OpenAi draft emits anyOf unions correctly but also makes every property required, which destroys the optional manifest. The cost of the workaround is that 'exactly one value field, and it is the one the kind names' is no longer expressible in the schema — it is enforced in Result.Validate and tested by TestAValueCarryingTheWrongFieldForItsKindIsRefused. Target: strata, not this repository.
```

**agent-plugins-ih18.25#4**

```
retro:tooling Avoided AGENTS.md's documented pipe-hides-exit-status trap while writing the swarm capability classifier: the natural expression (printf '%s' "$err" | grep -qi 'unknown command') pipes a small writer into grep -q under set -o pipefail, which the ledger already measured as intermittent (2/30) for an identical shape in this same script's own orphan guard. Used a captured variable matched with case/glob instead (swarm_capability() in pre-flight-check.sh), with no pipe at all.
```

**agent-plugins-ih18.24#4**

```
retro:tooling bd lint's 'Missing: ## Acceptance Criteria' check is a case-insensitive substring match on the description text, not a check that a real Acceptance Criteria section exists. Measured: a scratch bead (agent-plugins-bd97.8, since deleted) whose description read '...with no acceptance criteria...' was NOT flagged by 'bd lint <id>' (printed 'No template warnings found'), while an otherwise-identical bead with the phrase removed (agent-plugins-bd97.9) WAS flagged. So a bead whose description merely discusses 'acceptance criteria' as a topic — including, ironically, ih18.24's own description and this very comment — reads as satisfied to bd lint whether or not it supplies any. The re-homed gate documented in supervision.md inherits this false-negative; it only narrows bd lint's false-positive problem (cross-run noise), not this one. Not this repo's to fix (bd is an external dependency); flagging for anyone relying on the intersection as a complete guarantee.
```

**agent-plugins-ih18.18#3**

```
retro:tooling A WebFetch of beads' own installation docs (beads.gascity.com) returned partially fabricated details: it named environment variables BEADS_DOLT_SERVER_MODE, BEADS_DOLT_SERVER_HOST/PORT/USER as how bd is configured for server mode. Measured against the actually-installed bd (1.2.2) via 'bd init --help': those are CLI FLAGS (--server, --server-host, --server-port, --server-user), not environment variables -- only BEADS_DOLT_PASSWORD is an env var. The port default (3307) the fetch gave was correct. Did not act on the unverified env var names; used the measured flags instead. A second WebFetch of the same page, asked more narrowly, correctly reported it did not know the Windows dolt install method rather than inventing one.
```

### `proof`

**agent-plugins-ih18.21#1**

```
retro:proof the defect was reproduced from the command and from four unit tests before the fix, and each reproduction was re-run after it
evidence:
Command, before the fix (binary built from cli/ at 90cd22a, run against the live run epic):

  $ swarm manifest get --run agent-plugins-bd97 --key trailers
  summary: trailers: not a declared manifest key
  outcome: malformed (exit 3)

  $ swarm manifest require --run agent-plugins-bd97 --key trailers
  summary: trailers: not a declared manifest key
  outcome: malformed (exit 3)

  $ swarm manifest set --run agent-plugins-bd97 --key trailers --value '"x"' --dry-run
  summary: manifest: refusing the write, and storing none of it — "trailers" is not a declared manifest key
  outcome: malformed (exit 3)

Unit tests written first and observed failing for the stated reason
(go test ./swarm/manifest/... ./swarm/cli/... -run 'Trailer|StagelessKey|NoStage|ValidationSeparates', exit 1):

  --- FAIL: TestAKeyWithNoStageIsStillDeclared
      manifest_test.go:352: "trailers" is not in the declared set, though RunManifest declares it
  --- FAIL: TestRequiringAStagelessKeyReportsAbsenceRatherThanAnUnknownKey
      manifest_test.go:390: requiring an unwritten trailers gave [manifest/undeclared], want manifest/absent
      manifest_test.go:398: requiring a written trailers failed: "trailers" is not a declared manifest key
  --- FAIL: TestValidationSeparatesAKeyNotYetWrittenFromOneThatIsWrong
      validate_test.go:328: a record carrying trailers reports it as a key nothing declares
      validate_test.go:333: a numeric trailers value was accepted, though the key takes a string
  --- FAIL: TestTheVerbsServeAKeyThatHasNoStage
      manifest_test.go:353: get --key trailers exited 3, want 0

Two of those failures are a second defect the bead did not name: a manifest that
ACTUALLY CARRIES trailers was reported by `validate` as "a key nothing declares"
(RuleUndeclared, a failure), and a trailers value of the wrong type was accepted
because the key was never type-checked. Both fall out of the same cause and both
are fixed by the same change.

Same commands after the fix, same epic:

  $ swarm manifest get --run agent-plugins-bd97 --key trailers
  trailers = unestablished: the key is not written
  summary: trailers: not written
  what would resolve this: write it with: swarm manifest set --run agent-plugins-bd97 --key trailers --value <value>
  outcome: attention (exit 1)

  $ swarm manifest require --run agent-plugins-bd97 --key branch --key checks_baseline --key trailers
  branch = true / checks_baseline = true / trailers = false
  outcome: attention (exit 1)

  $ swarm manifest set --run agent-plugins-bd97 --key trailers --value '"x"' --dry-run
  summary: would write trailers as one merge; nothing was written
  outcome: clear (exit 0)

  $ go test ./swarm/...   -> exit 0, every package ok

```

**agent-plugins-ih18.16#2**

```
retro:proof I could not run `swarm status` from my own tree, so I extracted the REP commit and built the binary outside the worktree rather than restating the command's surface from its source.
evidence: `git archive 9d9a77e | tar -x -C <scratch>/repv` then `go build -o <scratch>/swarm-rep ./swarm` from `<scratch>/repv/cli` -> exit 0. Measured against the live run epic: `swarm-rep status --run agent-plugins-bd97` -> exit 0, `summary: 17/24 bead(s) done; 2 ready; last commit 16m35s ago`, and it reported `dispatches = coordinator-bd97: drained, dev-bd97-1: drained, dev-bd97-2: drained, dev-bd97-4: in progress: agent-plugins-ih18.16, dev-bd97-5: ready: agent-plugins-ih18.18,agent-plugins-ih18.23`, `scope_additions = agent-plugins-bd97.2.1, agent-plugins-ih18.22`, `scope_excluded = agent-plugins-ih18.22`. `swarm-rep status --run agent-plugins-bd97 --explain` -> exit 0, prints the six invocations and `summary: explained; nothing was measured`. `swarm-rep status --help` -> the four contract flags plus `--run` and `--stall-after` (default 30m0s). Every claim I wrote into supervision.md, SKILL.md and the module blueprint rests on one of those three runs; nothing was copied from cli/swarm/cli/status.go without running it.
```

**agent-plugins-ih18.14#1**

```
retro:proof The two-query drain rule was verified with a negative control, not just by a passing test: weakening drain to the ready query alone makes the AC-SWARM-CORE-003.4 test fail with exactly the message the AC is about.
evidence: with `d.Drained = len(d.Ready) == 0 && len(d.InProgress) == 0` replaced by `d.Drained = len(d.Ready) == 0`, `go test ./swarm/status/ -run TestEachAgent` -> "FAIL ... status_test.go:254: dev-1 reads drained though it holds a bead in progress; the ready query alone would say this". Restored, `go test ./swarm/status/` -> ok.
```

**agent-plugins-ih18.13#6**

```
retro:proof Every command form written into a skill was run against this runs real epic before it was documented, and one draft claim was retracted after measurement.
evidence: measured — `swarm manifest get --run agent-plugins-bd97 --key branch --key base_sha` exit 0; `... --key base_sha --json | jq -r .facts[0].value.text` -> the sha; `--key concurrency_cap --json | jq -c .facts[0].value` -> {"kind":"count","count":20}; `--key findings_path` -> "outcome: attention (exit 1)" with established:false; `require --key branch --key checks_baseline --key checks_ids` exit 0; `validate --run agent-plugins-bd97` -> "every rule holds at the dispatch stage", exit 0; `set --key branch --value ${q}"sprint/x"${q} --key base_sha --value ${q}"deadbeef"${q} --dry-run` -> write-once refusal naming both keys, exit 1; `set --key ledger --value ${q}["changed entry"]${q} --dry-run` -> "the ledger is append-only: the write drops 6 of its 7 entries", exit 1; `set --key retro_filed --value 3 --key retro_declined --value 0 --key retro_routed --value 1 --dry-run` -> "would write ... as one merge; nothing was written", exit 0. RETRACTED before it shipped: a draft sentence in sprint-scope/SKILL.md claimed `swarm manifest set` "refuses a partial write" of a nested key. It does not — Document.Merged in cli/swarm/manifest/manifest.go is `merged[k] = v`, a top-level replace — so the shipped sentence says a second write of standing_auth replaces the whole object instead.
```

**agent-plugins-ih18.11#1**

```
retro:proof the end-to-end tier caught a live defect on its first run: `manifest set --dry-run` exited 2 instead of 0.
evidence: the first recording of the case cli/testdata/swarm-cases/manifest-set-dry-run gave exit 2 where the case declares 0. Cause: Interpret read the fragment's answer through ValueFrom, and in dry-run the answer belongs to the write step, which the runner refuses — so the command reported a failure that had not happened. Fix: cli/swarm/cli/manifest.go now has a refused(report, steps) helper, and in dry-run reads the GUARD step's document and reports "would write <keys> as one merge; nothing was written" at OutcomeClear. Committed expected/stdout now reads: 'set-write: refused (dry run: the step writes)' / 'summary: would write review_passes as one merge; nothing was written' / 'outcome: clear (exit 0)'. Nothing in the unit tier would have found this: the unit tier asserts the PLAN, and the plan was correct — the defect was in how the surface read the report the runner returned.
```

**agent-plugins-ih18.10#4**

```
retro:proof Fifteen tests drive the REAL command tree with a faked process runner, so every 'nothing was read' claim is read off the fake's own record.
evidence: go test -v -count=1 ./swarm/cli/ — all pass. The ones carrying the weight: TestEveryManifestCommandRefusesAnEmptyScope drives all four verbs with no --run and asserts exit 3 AND zero invocations for each; TestGetRefusesAnUndeclaredKeyWithoutReadingAnything and the unknown --stage half of TestValidateDerivesTheStageAndSaysThatItDid assert the same for an argument error, so a refusal happens before the tracker is touched rather than after; TestSetRefusesTheWholeWriteWhenOneKeyIsRefused asserts a two-key write with one bad value touches the tracker not at all and names the offending key; TestSetRefusesAWriteBreakingARuleOnlyVisibleAgainstTheRecord and TestSetIfAbsentRefusesAKeyAlreadyWritten each DECLARE the write recording to the fake, so a command ignoring its guard would find something to run rather than failing for an unrelated reason, and then assert no --metadata invocation happened.
Also fixed by this item: the surface's registration guard asserted every top-level command carries --json, which a NOUN like 'manifest' does not — it has no RunE and cobra prints its verbs instead. It now walks the tree to the leaves, counts them, and fails if it reached none.
```

**agent-plugins-ih18.9#3**

```
retro:proof Every rule has a negative control, because a validator that refuses everything satisfies each rule's positive case and nothing else.
evidence, from go test -v -count=1 ./swarm/manifest/ — nine tests, all PASS:
  write-once: refuses a different branch AND a different base_sha; ACCEPTS re-writing the same value (which is what lets a resumed run reuse its branch rather than fail on finding one) and ACCEPTS the first write against a record with no branch, so the rule is write-once rather than write-never
  counter: refuses 1 and 3 against passes=2 budget=2; ACCEPTS 2 against passes=1 budget=2, so it is not 'refuse every write to the counter'
  ledger: refuses a removal, a reorder, a rewritten entry and an emptying; ACCEPTS an append AND an unchanged list, so the rule is about the prefix rather than about the length
  admission: names run-1.2 and run-1.3 for a three-list overlap; three non-overlapping lists produce no overlap finding
  effort: an absent key cites the unrecorded rule and the message says what to write; effort:"unrecorded" is clean; effort:"frantic" is a TYPE failure rather than an absence
  stages: a complete pre-flight record passes at pre-flight with branch, base_sha, admitted and findings_path reported as PENDING, and the SAME record fails at dispatch with branch, base_sha and admitted as failures while findings_path is still only pending — one record, two stages, opposite verdicts
  trailers is reported at neither severity at either stage
```

**agent-plugins-ih18.8#1**

```
retro:proof The conditional write is expressed INSIDE the plan model rather than around it: a two-step fragment whose first step is a read and whose DECODER is the guard, followed by the write. A key that moved makes the guard's decoder return an error, the step fails, and the plan's own ordering policy stops the write — there is no conjunction to write and none to forget.
evidence: TestAWriteThatWouldClobberAConcurrentChangeIsRefusedNamingTheKey declares BOTH invocations to the fake, so a store ignoring the guard would find the write recorded and run it; it asserts no argv containing --metadata was run, that Report.Stopped is the guard's step name, and that the guard's error names review_passes. Its negative control, TestAWriteAgainstAnUnchangedRecordGoesThrough, writes successfully against a record nobody touched, so a store that refused everything could not pass. TestADryRunPerformsTheGuardAndRefusesTheWrite shows a dry run performs the guard (a read) and refuses the write — a dry run that wrote the run record would be the worst bug this component could have.
The payload does not depend on the read, which is what keeps the whole fragment renderable before anything runs: TestSeveralKeysInOneInvocationAreOneReadAndOneWrite explains a four-key write and asserts the transcript carries both invocations.
```

**agent-plugins-ih18.8#3**

```
retro:proof bd's metadata write was re-measured here rather than taken from the FRD's transcript, and one thing the transcript does not say turned out to matter.
evidence, on bd 1.2.2 against a real bead:
  bd update <id> --metadata '{"probe_shape":{"a":1}}' --json -> exit 0; the bead's other ten metadata keys all survived (TOP-LEVEL MERGE), and --metadata takes the object INLINE as an argument — the help says 'JSON string or @file.json', and the string form works, so a plan step needs no temporary file
  bd update <id> --metadata '{"probe_shape":{"b":2}}' --json -> exit 0; probe_shape is now {"b":2}, so a NESTED value is replaced wholesale
  bd update <id> --metadata '{"probe_shape":null}' --json  -> exit 0; the key is now null, NOT removed
  bd update <id> --unset-metadata probe_shape --json         -> exit 0; the key is gone
The probe bead's metadata was restored to its original ten keys and verified. The null result is the one the FRD's transcript does not cover and it matters: a caller clearing a key must use --unset-metadata, because writing null leaves a key present with a null value — which a reader would report as present-and-null rather than absent, and AC-SWARM-CORE-002.3 turns on exactly that distinction. Nothing in this delta clears a key, so no operation for it was added; it is recorded here for whoever needs one.
```

**agent-plugins-ih18.7#6**

```
retro:proof The generator's own shape assertion was extended to the new package, for the same reason the spec model has one: the manifest is a chain of five inheriting structs, and the strata regression that assertion guards against empties an inheriting struct's properties while exiting 0. An empty RunManifest would make every manifest key read as unknown.
evidence: scripts/generate-schemas.sh now asserts, after building the swarm package, that every key declared by each of PreFlightKeys, ScopeKeys, PlanKeys, DispatchKeys and RetroKeys survives into RunManifest, and fails naming the missing keys and the upstream commit if not. Measured: 'scripts/generate-schemas.sh' exits 0 and reports four regenerated outputs; 'scripts/generate-schemas.sh --check' exits 0 with 'generated artifacts are up to date'; the full declared check set (13 checks) exits 0 with output matching the run's baseline line for line — 66 documents, no findings, 70 reference documents / 34 links / 0 errors, 13 templates / 0 errors / 34 advisories.
```

**agent-plugins-ih18.6#1**

```
retro:proof Every git shape the decoders pin was measured on git 2.43.0 here — in this repository, and in a scratch repository for the case that needs a deliberately broken worktree registration.
evidence, each probe with stdout and stderr to separate files and the status read directly:
  git worktree list --porcelain -> exit 0; records separated by blank lines, keys 'worktree <path>', 'HEAD <sha>', 'branch refs/heads/<name>', 'detached', 'locked [reason]'
  In a scratch repo with two worktrees, after 'rm -rf' on the second's directory: the deleted one is STILL LISTED, at exit 0, with an added line 'prunable gitdir file points to non-existent location'. 'git worktree prune' then exits 0 and it is gone. That is the hazard the delta names, reproduced rather than believed.
  git log -1 --format=%H%x00%ct%x00%s -> exit 0, three NUL-separated fields
  git merge-base --is-ancestor 3a57dc7 HEAD -> exit 0; the same pair reversed -> exit 1; with a bogus revision -> exit 128, 'fatal: Not a valid object name deadbeefdeadbeef'
  git rev-parse no-such-branch -> exit 128, 'fatal: ambiguous argument' on stderr AND the string 'no-such-branch' ECHOED ON STDOUT
  git rev-list --count --left-right HEAD...@{upstream} -> exit 128, 'fatal: no upstream configured for branch' (this worktree's branch has none)
  git for-each-ref --format='%(refname:short)%00%(objectname)%00%(upstream:short)' refs/heads/ -> exit 0; a branch with no upstream yields an empty third field
The rev-parse echo is the sharpest: a decoder that read stdout before the status would return 'no-such-branch' as a commit id. TestRevParseEchoingItsArgumentIsNotMistakenForACommit pins it.
```

**agent-plugins-ih18.6#5**

```
retro:proof The 'only the Git Version Control names git' rule is enforced by a test, verified by making it fail — the same mechanism as the tracker's half.
evidence: with a temporary cli/swarm/vcs/zz_probe_delete_me.go containing 'const probeToolName = "git"', 'go test -count=1 -run TestOnlyThisPackageNamesTheTool ./swarm/vcs/git/' printed '--- FAIL … git_test.go:379: ../../vcs/zz_probe_delete_me.go names "git"; only the Git Version Control may invoke it'. The probe was deleted and the full suite came back green across all nine swarm packages. It carries the same two anti-vacuity assertions as the tracker's guard (it scanned source outside the package; the walk reached the package itself) and the same stated limit (it catches a literal, not a name reaching an invocation through a variable).
```

**agent-plugins-ih18.5#1**

```
retro:proof Every shape the decoders pin was measured on THIS container against THIS repository's database (bd version 1.2.2 (6c124203e)), not inherited from the audit. The audit's own table is [measured here] but from a different session and a different database state, and two of its entries needed re-taking.
evidence, one probe per subcommand, each with stdout and stderr captured to separate files and the status read directly:
  bd show agent-plugins-ih18.5 --json        -> exit 0, type array, length 1
  bd show nonexistent-xyz --json             -> exit 1, type OBJECT on STDOUT: {"error":"no issues found matching the provided IDs","schema_version":1}
  bd list --parent nonexistent-xyz --json    -> exit 0, type array, []
  bd ready --parent nonexistent-xyz --json   -> exit 0, type array, []
  bd gate list --json                        -> exit 0, type NULL, body 'null'
  bd stats --json                            -> exit 0, type OBJECT, jq length 2 (schema_version, summary) — not a count of anything
  bd update agent-plugins-ih18.5 --status in_progress --json -> exit 0, type array
  bd update nonexistent-xyz --status open --json -> exit 1, STDOUT EMPTY, 'Error resolving nonexistent-xyz: no issue found matching' on stderr
  bd comment agent-plugins-ih18.5 <text> --json -> exit 0, type OBJECT {author,created_at,id,issue_id,schema_version,text}
  bd dep add agent-plugins-ih18.5 agent-plugins-ih18.2 --type blocks --json -> exit 0, type OBJECT, status:'added' EVEN THOUGH THE EDGE ALREADY EXISTED (which is what makes the write repeatable)
  bd dep add nonexistent-xyz agent-plugins-ih18.2 --json -> exit 1, type OBJECT carrying 'error' on stdout
The show/update pair is the one that settles 'a decoder per subcommand': their failure messages are on opposite streams, so no shared failure guard is right for both.
```

**agent-plugins-ih18.5#2**

```
retro:proof The parent-scope divergence was re-measured on the live run tree rather than inherited from agent-plugins-offx.6's scratch database, and the measurement is sharper than the record: the closed-bead omission is a SECOND divergence the bead does not mention.
evidence:
  bd list  --parent agent-plugins-bd97 --json              -> bd97.4 bd97.5 bd97.6 ih18            (direct children, OPEN only)
  bd list  --parent agent-plugins-bd97 --json --all --limit 0 -> bd97.1 bd97.2 bd97.3 bd97.4 bd97.5 bd97.6 ih18  (three CLOSED children appear)
  bd ready --parent agent-plugins-bd97 --json              -> ih18 ih18.12                          (TRANSITIVE: ih18.12 is a grandchild)
So list omits grandchildren AND closed beads, while ready reaches grandchildren but only the ready ones. Neither is 'every bead in this run', which is why Descendants walks and why Children passes --all.
Separately: bd list's help documents '--limit int  Limit results (default 50, use 0 for unlimited)', but measured, 'bd list --all --json' returned all 608 issues and '--limit 0' returned the same 608 — so the 50 default appears not to apply to JSON output. Children passes --limit 0 anyway: a query whose completeness rests on a default nobody stated is one silent truncation from reporting a run as smaller than it is.
```

**agent-plugins-ih18.5#5**

```
retro:proof The 'only the Beads Tracker names bd' rule is enforced by a test, and the test was verified by making it fail.
evidence: cli/swarm/tracker/beads/beads_test.go's TestOnlyThisPackageNamesTheTool parses every non-test .go file under cli/swarm and reports any string literal equal to the executable name outside that package. With a temporary cli/swarm/plan/zz_probe_delete_me.go containing 'const probeToolName = "bd"', 'go test -count=1 -run TestOnlyThisPackageNamesTheTool ./swarm/tracker/beads/' printed '--- FAIL … beads_test.go:238: ../../plan/zz_probe_delete_me.go names "bd"; only the Beads Tracker may invoke it'. The probe was deleted and the same command printed 'ok'. Two anti-vacuity assertions: it fails if it scanned no source outside the package, and if the walk never reached the package itself.
Stated limit, in the test's own doc: it catches a call site naming the tool directly and cannot catch a name reaching an invocation through a variable. That is what a textual guard can do, and it is more than the sentence it replaces.
```

**agent-plugins-ih18.4#5**

```
retro:proof No subcommand exists yet — SURF delivers the surface and the commands arrive with later deltas — so the surface is driven through a SYNTHETIC subcommand built from the same Spec a real one supplies, with the exectest fake as its process runner. That makes 'nothing was read or written' an assertion about the runner's record rather than an inference.
evidence: go test -v -count=1 ./swarm/cli/ — 21 tests, all PASS. TestACommandThatMutatesRunStateRefusesAnEmptyScope drives three shapes of empty scope (no --run, --run '', --run '   ') and asserts BOTH exit 3 and len(fake.Argvs())==0, so a surface that checked the scope after building and running a plan would fail even though its status was right. TestTheEnvelopeIsEmittedOnEveryExitStatus drives one case per status and fails if any of 0,1,2,3 went unexercised, so the coverage claim cannot rot into a subset. TestTheExitStatusSaysWhatTheCallerShouldDoNext asserts its table covers len(Outcomes()) entries, so adding a fifth outcome without extending the table fails rather than passing silently.
```

**agent-plugins-ih18.3#1**

```
retro:proof Not one test in cli/swarm/plan starts a process. The Process Runner is substituted by the fake from the PROC delta, which records every invocation it was asked for — so 'no tool is executed' is asserted rather than assumed.
evidence: TestExplainPrintsEveryInvocationInOrderAndExecutesNone builds the runner with a fake holding NO recordings, so any executed step would both error and leave a call; after explain it asserts len(fake.Argvs())==0 and that the three transcript lines are 'bd show run-1 --json', 'bd ready --parent run-1', 'bd update run-1.4 --claim' in that order. TestAStepWhoseFailurePolicyIsStopPreventsEveryLaterStep declares a recording for the step that must NOT run ('git worktree remove dev-1'), so a runner ignoring the guard would find something to run and fail on the assertion about what ran rather than on an unrelated 'no recording' error. go test -v -count=1 ./swarm/plan/ — 19 tests, all PASS.
```

**agent-plugins-ih18.2#1**

```
retro:proof The four properties ADR-025 selected this design for are tested against a REAL child process, not a fake, because a fake cannot demonstrate that a real process's status was read from the process. The child is this test binary re-invoked with a mode argument after a '--' separator (the standard library's own os/exec idiom), so the tests need no shell and no toolchain at run time and survive the windows-latest leg.
evidence: go test -v -count=1 ./swarm/exec/... — 9 tests in the runner package and 8 in the fake's, all PASS. Named: TestAFailingCommandIsReportedAsFailedEvenThoughItProducedOutput (child writes three lines AND exits 7 -> ExitStatus 7, output intact, err nil); TestASucceedingCommandWithMoreOutputThanAPipeHoldsIsNotReportedAsFailed (2 MiB of stdout, exit 0 -> ExitStatus 0 and len(Stdout)==2097152, i.e. not truncated at a 64 KiB pipe buffer); TestTheOutputAndErrorStreamsAreNeverMerged (child writes {"ok":true} to stdout and 'warning: beads.role not configured' to stderr -> json.Unmarshal(Stdout) succeeds and Stdout does not contain 'warning'); TestAnArgumentThatLooksLikeShellSyntaxIsPassedThroughLiterally (the single argument 'a; rm -rf /; $(echo hi) | tee x' comes back as exactly one argument, verbatim).
```

**agent-plugins-ih18.2#2**

```
retro:proof The 'test-only package' guard was verified by making it fail, not by reading it. cli/swarm/exec/exec_test.go's TestTheFakeIsNotReachableFromProductionCode parses every non-test .go file in the module and fails on an import of .../swarm/exec/exectest.
evidence: with a temporary cli/swarm/cmd/zz_probe_delete_me.go containing 'import _ "github.com/brettcallaghan/agent-plugins/cli/swarm/exec/exectest"', 'go test -count=1 -run TestTheFakeIsNotReachableFromProductionCode ./swarm/exec/' printed '--- FAIL: TestTheFakeIsNotReachableFromProductionCode (0.00s)  exec_test.go:323: ../../swarm/cmd/zz_probe_delete_me.go imports the test-only fake package; the fake must not be reachable from production code' and exited FAIL. The probe file was deleted and the same command then printed 'ok  github.com/brettcallaghan/agent-plugins/cli/swarm/exec  0.007s'. The test also carries two anti-vacuity assertions — it fails if it scanned zero non-test files (the walk is not reaching the module) and if no _test.go file imports the path (the literal string is stale after a rename).
```

**agent-plugins-ih18.1#4**

```
retro:proof The cross-compile matrix and the release archive loop were executed here rather than asserted, because both are shell edited inside YAML that nothing in the declared checks runs.
evidence: (1) bash -n over every 'run:' block in both workflows after the edit: 15 blocks, all exit 0. (2) The cross-compile loop, run verbatim from cli/ with CGO_ENABLED=0 over 'linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64' x './spec ./swarm': 12 builds, output ended 'CROSS_COMPILE_OK'. (3) The release loop, run verbatim with version=v0.0.0-test, produced exactly 12 archives and a 12-line SHA256SUMS: spec_v0.0.0-test_{darwin,linux}_{amd64,arm64}.tar.gz, spec_v0.0.0-test_windows_{amd64,arm64}.zip, and the same six for swarm — the spec archive names are byte-identical to what the loop produced before the edit, so no existing asset name changed. (4) file -b on the CGO_ENABLED=0 linux/amd64 swarm build: 'ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, BuildID[sha1]=..., stripped'.
```

**agent-plugins-ih18.25#2**

```
retro:proof Verified both directions live in this container. Before rebuilding: pre-flight-check.sh reported 'swarm: /usr/local/bin/swarm (STALE -- no status subcommand)' against the binary this run's own earlier bead had built before swarm status existed. Confirmed the exact measured failure mode the AC names: /tmp/.../scratchpad/swarm status --help (a preserved pre-status build) exits 3 with stderr 'Error: unknown command "status" for "swarm"'. After running (cd cli && go build -o /usr/local/bin/swarm ./swarm), the same probe reported plain 'swarm: /usr/local/bin/swarm' (ok). Also verified the session-start.sh addition's idempotency directly: with the now-capable binary in place it computed NEED_SWARM_BUILD=no and skipped the build; with a fake binary that exits 3 with 'unknown command' it computed NEED_SWARM_BUILD=yes and rebuilt successfully.
```

**agent-plugins-ih18.24#3**

```
retro:proof Verified the re-homed gates catch a real mid-run addition, using scratch beads under agent-plugins-bd97 created then deleted (bd delete) with no working-tree impact: (1) tier gate — created agent-plugins-bd97.8 (task, parent agent-plugins-bd97, assignee dev-bd97-6, no metadata.tier), and 'bd ready --parent agent-plugins-bd97 --exclude-type=epic --json | jq -r ...select((.metadata.tier // "")=="")...' printed its id. (2) human-queue gate — created agent-plugins-bd97.10 with --labels human; 'bd list --label human --status open --json' listed it. (3) bd-lint-intersection gate — created agent-plugins-bd97.9 with a description containing no mention of acceptance criteria; 'bd lint agent-plugins-bd97.9' printed 'Missing: ## Acceptance Criteria', and the documented awk+comm recipe against the live bd97 ready queue correctly returned empty (no false positives against the run's actual, properly-described queue). All three scratch beads were deleted immediately after (bd delete, 'Removed 1 dependency link(s)' each time); none appear in bd list --parent agent-plugins-bd97 now.
```

**agent-plugins-ih18.24#5**

```
retro:proof Portability trap caught before it shipped: the first draft of the bd-lint-intersection recipe used awk's Perl-style '\S' shorthand ('/^\S+ \[/'). Under mawk 1.3.4 (this container's default /usr/bin/awk — 'awk --version' -> 'mawk 1.3.4 20240123'), '\S' is not a supported escape and the pattern silently matches nothing: piping real 'bd lint' output through it produced 0 ids against an expected 93-94. Replaced with the POSIX class '[^[:space:]]+', re-ran, got the expected 94 (matching bd lint's own 'Template warnings (94 issues...)' summary line). The committed recipe in supervision.md uses the POSIX form.
```

**agent-plugins-ih18.23#3**

```
retro:proof Added cli/swarm/e2e/worktree_attribution_test.go's TestAWorktreeIsAttributedByItsCommitNotItsBranchName, which restores the real 'core' fixture, adds a real worktree on a branch named like the harness's own worktree-agent-HASH convention (no bb/ prefix) pointed at dev-fixture-1's own commit, then runs the compiled swarm binary's real status --json and asserts dev-fixture-1 is named with that worktree's real path -- a REAL 'git worktree list --porcelain' capture per the bead's Done-when, not a hand-written one. A second half of the same test is the negative/comparison control: the identical commit checked out under a bb/dev-fixture-1-shaped branch attributes identically, proving the branch's name plays no part either way.
```

**agent-plugins-ih18.19#6**

```
retro:proof Verified the two retired-command claims the dispatch handed down rather than inheriting them, and both held.
evidence: (1) 'bb-status.sh run <epic> is retired and exits 3' — ran 'plugin/skills/bead-bonanza/scripts/bb-status.sh run agent-plugins-bd97', captured to a file and read the status directly (not through a pipe): exit=3, stderr 'bb-status: the `run` mode is retired; use: swarm status --run <epic-id>'. Source: bb-status.sh:277 'run) die ...', and die() exits $RC_USAGE. (2) 'swarm status --run agent-plugins-bd97 now works' — exit 0, 24/26 done, and it named its own invocations per fact. Also negative-controlled the DoD's manifest-adoption claim rather than trusting the deltas: 'grep -rn' for a hand read of the manifest across plugin/ returns exactly one hit, pre-flight-check.sh:789, and that one is inside a 'say' string telling the reader how to look — which is the deferred .22 and nothing else.
```

**agent-plugins-ih18.18#1**

```
retro:proof Settled ADR-027's explicitly unmeasured question with a real local reproduction (Linux container, not Windows, but the bd/dolt server-mode mechanism is platform-independent): downloaded dolt 2.3.1's linux-amd64 release, put it on PATH, ran 'bd init --server --prefix X' with no --server-port given. bd auto-started its own managed dolt sql-server (confirmed via ps aux) with NO separate 'dolt sql-server' invocation, then bd create/bd list worked against it. Ran a second 'bd init --server' in a different directory immediately after: it came up on a DIFFERENT auto-selected port (36331 then 37165) with no collision -- which is what lets this suite's parallel (t.Parallel()) fixture restores coexist. Wired into cli/swarm/e2e/harness_test.go's restore(): passes --server to bd init only when runtime.GOOS==windows.
```

### `environment`

**agent-plugins-ih18.21#4**

```
retro:environment run agent-plugins-bd97 resolved no commit trailers, so every bead in it closes without --trailers-from
evidence:
  $ swarm manifest require --run agent-plugins-bd97 --key branch --key checks_baseline --key trailers
  branch = true
  checks_baseline = true
  trailers = false
  outcome: attention (exit 1)

  $ bb-close-bead.sh --trailers-from agent-plugins-bd97 agent-plugins-ih18.21 1 "<title>"
  bb-close-bead: agent-plugins-bd97 carries no `trailers` in its manifest — omit --trailers-from when the run resolved none
  exit 2

The script refused with nothing done, exactly as run-batch/SKILL.md says it
would, and the bead closed on the re-run without the flag. Nothing is wrong
here — it is recorded so the next dispatch in this run does not read the exit 2
as a defect in the fix it is standing on. Note that this measurement is only
readable at all BECAUSE of this bead: before it, `swarm manifest require --key
trailers` exited 3 malformed rather than reporting the key as not written.

```

**agent-plugins-ih18.16#5**

```
retro:environment The `swarm` binary on PATH in this container predates the status command, so every document this bead just wrote points at a subcommand the installed binary does not have.
evidence: `ls -la /usr/local/bin/swarm` -> `-rwxr-xr-x 1 root root 4859556 Aug 29 03:55`; `swarm status --help` -> `Error: unknown command "status" for "swarm"` while `swarm manifest require --run agent-plugins-bd97 --key branch --key checks_baseline` succeeds (exit 0). `grep -rn swarm .claude/hooks/session-start.sh` -> no output; the only `swarm` mentions in `plugin/skills/bead-bonanza/scripts/pre-flight-check.sh` are the two prose lines this bead edited. So the binary is hand-installed, nothing re-installs it, and the Environment Probe does not probe for it. Filed as agent-plugins-ih18.25.
```

**agent-plugins-ih18.14#3**

```
retro:environment The worktree-to-dispatch matching rule the module documents does not describe the worktrees this harness creates, so the standing-worktree half of the collector matches nothing in a real run. Filed as a discovered-from bug rather than fixed here.
evidence: `git worktree list --porcelain` in this dispatch, verbatim: main checkout on `refs/heads/claude/bead-bonanza-agent-plugins-ih18-2yjc8j`; dev worktree on `refs/heads/worktree-agent-add9532b34aa266aa`, locked "claude agent agent-add9532b34aa266aa". The documented rule is `git worktree add -b "bb/<dispatch-id>"` (worktree-mechanics.md:122) and `owner_for_branch` stripping `bb/` (sprint-cleanup.sh:156). The dispatch id here is dev-bd97-4; neither branch name carries it or the prefix.
```

### `unrouted (kind absent from the taxonomy)`

**agent-plugins-ih18.20#1**

```
retro:work-clean Work was clean: documentation changes reconciled between two documents with no test failures. All checks passed on the modified tree.
```

**agent-plugins-ih18.17#1**

```
retro:process:run-batch — dev-bd97-2 closed this bead recording NO retro comment. Its return said 'No finding comments were needed - the work was straightforward', but the bead's own Retro section requires the absence to be RECORDED ('A clean item records none; say so rather than inventing filler'), and the acceptance gate reads a missing comment as a failed signal 5, not as a clean item. Caught by the gate, not the agent. This line is the coordinator's repair; whether the work was genuinely frictionless is UNVERIFIED by anyone who did it.
```

**agent-plugins-ih18.12#1**

```
retro:process:run-batch — dev-bd97-2 closed this bead recording NO retro comment, which its own Retro section requires either way ('A clean item records none; say so rather than inventing filler'). Caught by the acceptance gate at signal 5, not by the agent. Whether the work was actually frictionless is UNVERIFIED: the agent stated neither findings nor their absence, and this line was added by the coordinator to repair the gate, not by anyone who did the work. First batch dispatched at the light tier; whether tier is the cause is unestablished on one instance.
```


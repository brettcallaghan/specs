# Sprint report — run 3, `agent-plugins-mpnf`

**Run epic:** `agent-plugins-mpnf` — "Run 3: adopt the beads graph (plan items 6-14), and the corrections that ride with them"
**Branch:** `claude/bead-bonanza-design-review-ne1plu` (PR #27) · base `653e885` · head `663abe0`
**Findings, verbatim:** `docs/research/sprint-findings/2026-08-25-agent-plugins-mpnf.md`

---

## The outcome: this run was wound down, not finished

The queue drained and every admitted bead closed. That is not what happened
here, and a reader who stops at the results table will take away the wrong
thing.

**Run 3 discovered that the architecture it was refining is the wrong one, and
its owner ended the branch mid-flight rather than driving it to convergence.**
The three measurements that reframed it were taken by the coordinator at
`7120474` and confirmed by the owner. All three are re-measured here at head
`663abe0`, because a claim binds where it is used:

**1. The skill family is prose, not code.**

```
$ git ls-tree -r --name-only 663abe0 -- plugin/skills | grep '\.md$'  | xargs -I{} git show 663abe0:{} | wc -l
20036
$ git ls-tree -r --name-only 663abe0 -- plugin/skills | grep -E '\.(sh|py)$' | xargs -I{} git show 663abe0:{} | wc -l
3442
```

At `7120474`, the commit the owner read, the same two commands returned
**20030** and **3165** — the coordinator's figures reproduce exactly. Of the
fourteen skills under `plugin/skills/`, **twelve contain no executable code at
all**; only `bead-bonanza` (2,865 lines) and `groom-beads` (577) do. Narrowed
to the ten skills of the sprint family, **eight of ten contain no code**.

**2. The prose has hit a hard ceiling.**

`scripts/check-skills.sh` caps a skill body at 20,000 bytes (`BODY_BYTE_MAX`,
line 43). Measured with the checker's own method — `text.split('---', 2)[2]`,
UTF-8 bytes:

| skill | body bytes | headroom |
|---|---:|---:|
| `sprint-plan` | 19,995 | **5** |
| `bead-bonanza` | 19,907 | 93 |
| `groom-beads` | 19,850 | 150 |

Rules in `sprint-plan/SKILL.md` now trade against each other for space. Bead
`agent-plugins-mpnf.4`'s own record reads *"byte budget spent to 5 bytes free,
not blocked"*; `qozp.6` met a 54-byte budget by *"losslessly tightening three
existing paragraphs"*. Neither cut a rule, and neither could do it twice.

**3. Review is filing faster than the run is closing.**

Of this run epic's 44 children, **13 were the queue it opened with**. The other
31 are what the run generated about itself: 24 review findings across two
passes (`mpnf.3`–`.14` from pass 1, `mpnf.15`–`.26` from pass 2), one mid-run
amendment, one bead a developer filed, three merge beads, one health event, and
`p567`. **Two review passes over a thirteen-bead queue produced twenty-four
findings, four of them P1, and every one of them a defect in what this run's
own beads had just merged.**

Across all three runs' epics:

```
$ for e in agent-plugins-qozp agent-plugins-so39 agent-plugins-mpnf; do \
    bd list --parent $e --status all --json | jq "length, ([.[]|select(.status==\"closed\")]|length)"; done
qozp:  16 children, 11 closed
so39:  23 children, 12 closed
mpnf:  44 children, 27 closed
```

**83 beads, 50 closed, 33 still open or deferred.** Each review pass found real
defects in the previous pass's remedies. That curve does not converge by
running another pass.

**The owner's conclusion, which is this run's actual outcome:** the swarm model
works and the substrate is failing. The skill family is a distributed system's
control plane expressed as documentation — invariants that only prose enforces,
state that only prose records, and a byte cap that now decides which rules get
to exist. A capability-CLI epic, **`agent-plugins-offx`** ("Swarm capability
CLI: replace prescribed command sequences with a tested action surface"), was
opened as a direct result with 15 children, and this branch was wound down.

The wind-down is visible in the run's own record: six admitted findings were
**de-admitted** by owner decision because they are prose-only corrections to
documents the CLI will rewrite or delete, and one mid-run bead was declined as
an amendment under the same decision. That is section *Scope* below.

---

## Incidents

Every decision this run made that changed what "correct" means. **Eleven
distinct**, by the protocol's own grep over the epic and its children:

```
$ bd export | jq -r '.comments[]?.text' | grep -E '^(retro:[a-z]+:[^ ]+ incident|amend: incident)'
```

scoped to `agent-plugins-mpnf` and its 44 children, that returns 17 records —
7 of which are a bead's own record mirrored onto the epic. **One further
incident the grep does not match at all**, and three decisions of the same
weight recorded in forms it also misses, are named at the end of this section.

### The three that matter most

**1. `retro:incident:coordinator-removed-worktree-before-the-merge-succeeded`
— a state the corpus says cannot occur.**

Review loop 2, batch 10. `git merge --ff-only bb/dev-mpnf-10`, `git worktree
remove` and the acceptance check ran as one chained command. **The merge exited
128** ("Not possible to fast-forward" — batch 9 had already moved the head from
`22c86ab` to `e2b69ad`) **and the removal ran anyway, and succeeded.**

`references/worktree-mechanics.md` makes worktree presence the in-flight
signal: *"you remove it as the final act of accepting, after the last of its
branches merges."* For roughly 90 seconds a removed worktree coexisted with an
unmerged branch. `bb-status.sh bead` reported it faithfully and exited 2:

```
OK   3 -> STOP  3 NOT merged into HEAD
OK   4         worktree: gone
```

Signal 4 read "accepted" while signal 3 read "not merged". **The check is not
wrong — the ordering it depends on is enforced by prose and by nothing else.**
Nothing was lost; the branch still carried the commit, and a rebase onto
`e2b69ad` produced `78b34e1`, after which all five signals read OK. Evidence:
`/tmp/m10.out` (exit 128), `/tmp/wr10.out` (exit 0), `/tmp/acc10.out` (exit 2),
`/tmp/acc10b.out` (exit 0).

Rejected: treating it as operator error. The cost of that would have been
leaving the ordering unguarded. Filed instead as `agent-plugins-y63n.20`
(removal behind an ancestry guard) and `y63n.21` (a `--ff-only` refusal in a
concurrent wave reads as the serialized STOP case).

**2. `retro:environment:retracted-bd-ready-json-returns-null-when-empty` — a
false tool claim reached three dispatch prompts, hours after this same run
merged the rule forbidding exactly that.**

The claim: *"`bd ready --json` returns `null` when empty, but `[]` for an
`--assignee`-filtered empty query."* **It does not reproduce.** The coordinator
measured three configurations on bd 1.2.2, all `[]`. `dev-mpnf-8` had already
measured four and flagged the contradiction rather than working around it —
which is the only reason it was caught. `mpnf.18`'s agent later measured
**eleven configurations with a positive control**, and found `null` in none of
them.

Its route in: the loop-2 grooming agent recorded it as a ledger entry citing
*"the run's existing ledger records `null`"* — and **no such ledger entry
exists**. The retro's autopsy found the citation was fabricated. The
coordinator then relayed the whole entry into three dispatch prompts without
measuring it. `agent-plugins-x5tk`, which merged in wave 1 of this same run,
had already added "Tool claims are measured at the point of use" to `AGENTS.md`
and enumerated dispatch prompts as an unaudited belief-carrying surface.

Nothing depended on it in any of the three beads, so no work is affected. The
sweep found the claim absent from the tracked corpus, from all three epic
ledgers, and from `.sf/outbox/`; the three dispatch prompts were unreachable
and are named rather than fixed. **The cheapest interception was one grep at
the moment the ledger entry was admitted — not vigilance at every point of
use.** Deduped onto `agent-plugins-offx.4`.

One thing the retraction is *not*: `bd human list --json` really does return
`null`. Measured by `mpnf.18`'s agent in the same scratch database. **bd's
empty-JSON shape is per-subcommand and cannot be generalised from one to
another.**

**3. `retro:decision:review-pass-2-admission` — the authorization gate was
overridden, for the second time in one run.**

`standing_auth.admit_review_findings` is `{value: false, source: "default"}`.
The human did not answer the row; `sf.toml` leaves the key commented out. In
review loop 1 the coordinator admitted 4 findings of 12 against it; in loop 2
it admitted 11 of 12 against it again.

The manifest's own note on that key:

> NOBODY CHOSE THIS. … In runs 1 and 2 this same silent default declined EVERY
> review finding, including defects in those runs' own diffs. The report must
> name this row as a default, not as a decision.

The decisive finding in loop 1 was `mpnf.4`, verified independently by the
coordinator in a scratch database: `bd create --graph` edges resolve **only**
against keys within the plan, so a collision edge between two beads that
already exist cannot be expressed at all. `qozp.8` had just deleted the wave
table and replaced it with collision edges. **The replacement for the mechanism
this run removed was non-functional, and nothing serialized two beads sharing a
file.** That is not scope growth; it is the run's own deliverable not working.

Rejected: honouring the gate, as runs 1 and 2 did. The cost of that would have
been shipping a branch whose central serialization mechanism does not exist.

**Across three runs this row has never once been chosen by anybody.** It took
its silent default every time, and in runs 1 and 2 that default declined every
finding, including P0s and P1s in those runs' own merged diffs. Filed as
**`agent-plugins-offx.8`**. A defaulting mechanism that produces the wrong
answer in three runs out of three is not a policy; it is a bug with a config
key in front of it.

### The rest

| # | Record | Decided | Rejected, and the cost |
|---|---|---|---|
| 4 | `retro:tooling:coordinator` — `bd dep add … >/dev/null 2>&1` treated as done | Re-ran reading every status; `dep remove` then `dep add` repaired the edge | Trusting the redirect. `0913` edits the same files as items 8 and 9 and would have dispatched concurrently with them. Caught only because the ready front was re-queried |
| 5 | `retro:proof:bd-graph` — **`bd create --graph` edge orientation is the reverse of the natural reading**; `from_key` WAITS FOR `to_key`. Measured twice, independently, with a negative control both times | Corrected semantics written into `sprint-plan/references/run-record.md`; dispatch prompts for items 8, 9, 11 carry the fact explicitly | Relying on agents to find it. Writing an edge the intuitive way inverts every dependency in a plan, at exit 0, with no warning. Two further silent traps measured the same pass: unknown node/edge fields are dropped with a warning at exit 0; `parent_id` holding a plan key yields an orphan |
| 6 | `retro:proof:sprint-scope` — the coordinator's scope note predicted `bd swarm validate` would raise no outside-dependency warning. **It raised six.** All six point at closed beads | `plan-mpnf` followed `run-record.md`'s unconditional rule and derived packing from the re-queried ready front | Believing the prediction, which would have licensed using a partition the rule forbids. Scope's underlying observation was right; its claim about tool behaviour was reasoned, not measured |
| 7 | `retro:process:coordinator` — **second stall of the session.** After admitting four findings the coordinator ended the turn having *described* the next dispatch rather than performing it | Recorded as recurrence 1 of an invariant filed and approved during this same run | Nothing. The human had to say "go ahead." The record names it as hedging in the shape of a pause: *"announcing a decision and then falling silent is neither deciding nor asking"* |
| 8 | `amend: incident on qozp.6` — the dispatch prompt's Scope list and its byte-budget paragraph named **different file sets** | Agent edited the fourth file (`sprint-plan/references/run-record.md`) and recorded an incident | Silently widening, or blocking. No sibling owned the file, so it cost nothing. The defect is the coordinator's: a dispatch prompt is an authoring surface with no reviewer. Filed as `y63n.23` |
| 9 | `amend: incident on qozp.13` — the bead's sketch read an env var a hook process cannot see | Arming became an on-disk file written by `bb-stop-ready-loop.sh --arm <epic>`; the hook is inert without it | The env-var design. Measured: `bd ready --parent ""` returns 60, not 0. Nothing in the dispatch path calls `--arm` yet — filed as `agent-plugins-89xd` |
| 10 | `amend: incident on qozp.17` — the agent roster has **no row for the scope phase and cannot be given one** (`sprint-scope` mints the run epic, so no run-id exists at dispatch) | Pre-flight, scope and post-flight act as `coordinator-<run-id>` | Leaving scope unassigned. Scope takes the authorization gate, and a hold under an id the coordinator cannot recompute is a hold nobody can release |
| 11 | `amend: incident on qozp.10` — acceptance criterion 5 **and** the bead's Design were both factually wrong on bd 1.2.2 (the `human` label does not hold a bead out of `bd ready`; `bd human respond`/`dismiss` both error with "storage is nil") | Criterion amended to the measured mechanism — the late parent phase defers, the dispatch front excludes the label, closing releases via `bd epic close-eligible` | Implementing the criterion as written, which was unsatisfiable |
| 12 | `amend: incident on mpnf.4` ×2 — the Design named two artifacts but not how the second path is recorded; and AC7 contradicted the bead's own SCOPE note | Manifest gains `edges_sidecar` beside `graph_plan`; AC7 honoured at batch end by `mpnf.3` rather than at `mpnf.4`'s commit | Re-deriving the sidecar path by extension substitution — rejected because recovery is meant to read the epic, not infer a filename convention |
| 13 | `amend: incident on mpnf.17` — AC3 ("a flat epic reports exactly what it reported before") met **in substance, not in letter** | State, verdict and exit code byte-identical on a flat epic; the citation column changed from `bd list --parent <epic> --status in_progress` to `bb-status.sh slot\|run <epic>` | Keeping the old citation, which re-derives to the wrong answer on a nested epic |

### Two counting problems, and they are the finding

**The run's best incident record routes nowhere.** `retro:incident:coordinator-removed-worktree-before-the-merge-succeeded`
was written with `incident` in the **kind** slot. `incident` is not one of the
protocol's nine kinds — it is a *severity*. The protocol says so and shows the
correct form (`retro:gap:ADR-014 incident — …`) but **nowhere states the wrong
form**, so an author who reads "incident" as a first-class concept has nothing
telling them it may not sit there. The routing table sends it to UNROUTED: a
count and a list, no audience, no candidate. It reached one only because a
human read it. Filed as **`agent-plugins-sdu1`**.

**The counts disagree, and none of them is wrong.** The retro records 8
incidents; the protocol's grep over this epic and its children returns 17
records collapsing to 11 distinct; and three further decisions of the same
weight — review pass 1's admission (written as a bare `incident:` prefix,
which `y63n.2` records is harvested by nothing), review pass 2's admission, and
the `bd ready` retraction — match no incident pattern at all. **A severity that
depends on an author remembering where to put a word is not a severity.**

---

## Waiting on a human

```
$ bd list --label human --status open --json
[]
```

**None.** No bead is holding a phase epic open on a human answer.

---

## Scope: the run narrowed itself after admission

The admitted set moved four times. The arithmetic:

| | beads | running total |
|---|---:|---:|
| admitted at scope | 13 | 13 |
| `+` mid-run amendment (`mpnf.2`) | +1 | 14 |
| `+` review pass 1, admitted against the gate | +4 | 18 |
| `+` review pass 2, admitted against the gate | +11 | 29 |
| `−` **de-admitted by owner decision** | −6 | **23** |

**Six beads were admitted and then de-admitted mid-run**, 2026-08-25, by owner
decision. The manifest records the reason on each:

> All six are prose-only corrections to bead-bonanza reference documents that
> the capability-CLI epic (`agent-plugins-offx`) is expected to rewrite or
> delete, so fixing them here would spend the run's last review loop on text
> being replaced. The finding and its measurement are preserved on each bead.
> **This run therefore NARROWED ITS OWN SCOPE after admission; the report must
> say so.**

| de-admitted | P | what it records |
|---|---|---|
| `agent-plugins-mpnf.15` | P1 | Graph-plan examples parent merge beads with `parent_key: "target"`, a key no plan contains: measured total abort |
| `agent-plugins-mpnf.19` | P2 | Merge beads not excluded from the dispatch front: a `sprint-batch` subagent is handed a merge bead only the coordinator may run |
| `agent-plugins-mpnf.21` | P2 | `dispatching.md`'s worktree table claims `sprint-cleanup` "applies exactly this rule": measured, row 3 is force-removed and the blocked row is absent |
| `agent-plugins-mpnf.22` | P2 | `recovery.md`'s "`bd ready --assignee` is exactly that remainder" is false under the new collision chain: measured 1 of 3 |
| `agent-plugins-mpnf.23` | P2 | `run-record.md` contradicts itself on graph-plan atomicity |
| `agent-plugins-p567` | P1 | Orchestrator loads only the graph plan, so the edge sidecar never reaches the tracker |

`agent-plugins-mxl3` (P2, `sprint-cleanup` run from inside a run worktree
treats that worktree as main and never tears it down) was filed mid-run by
`mpnf.20`'s agent and **declined as an amendment under the same wind-down
decision.** Real and measured, adjacent and non-destructive. Not an oversight.

**`p567` is the one that should worry a reader.** It is P1, it was admitted,
and it is the orchestrator half of the sidecar mechanism `mpnf.4` shipped the
planner half of. As it stands, an orchestrator following the merged text loads
the graph plan, wires no edges, and reports success.

### Excluded at admission

The manifest's `excluded` list holds 31 entries. Seven are the de-admissions
and the declined amendment above; the other **24 were never admitted**:

- **10 run-2 review findings** (`so39.2`–`.11`) — *"deferred because
  `admit_review_findings` was false."* Still deferred. This is the same silent
  default described above, one run earlier.
- **6 wave-related** (`qozp.21`, `.24`, `.25`, `.26`, `.27`, `b1gl`) — retired
  **by** item 8 rather than fixed, since item 8 deletes the code they report
  against. The batch grep-proved them void and correctly neither revived nor
  closed them.
- **6 "real work, not admitted this run"** — `9owj`, `r9yj`, `vqmc`, `2nb8`,
  `we50`, `o76b`.
- `4o4v` (capability-audit research, deferred by the owner) and `xpn8`
  (bounced in run 2 on a measurably false premise; belongs to `so39`).
- `qozp.22` is **closed as superseded**, not a live exclusion, and must not be
  revived.

### Findings left at loop exhaustion

The review budget was 2 and both passes were spent. **17 beads are left
deferred** under this epic — measured, `bd list --parent agent-plugins-mpnf
--status all --json | jq '[.[]|select(.status=="deferred")]|length'` → 17:
ten review findings never admitted (`mpnf.6`–`.11`, `.13`, `.14`, `.25`,
`.26`), the six de-admitted above, and `mxl3`. `mpnf.26` — three skill bodies
within 150 bytes of the cap — was deferred deliberately: the owner parked the
byte-cap category ("we'll address them towards the end of the refactor"), and
it remained a live constraint *on* loop 2.

There was no pass 3. Nothing checked the eleven loop-2 remedies for
regressions, and the coordinator recorded that as a reason to prefer
conservative fixes.

### Which reviewer ran

`reviewer_source: declared`. `sf.toml`'s `[review]` names `skill =
"code-review"`; it resolved in this session and ran at **high effort** over
`653e885...0b6cfa1`, with `base_sha` read off the epic's manifest rather than a
recomputed merge-base. The review agent independently reproduced the
load-bearing measurements and added two findings of its own. **This is the
strongest of the three reviewer sources** — not a skill discovered in the
session, and not a bare agent.

---

## Results

23 admitted beads, one commit each, in merge order.

| ID | Status | Agent | Commit | Notes |
|---|---|---|---|---|
| `agent-plugins-x5tk` | closed | dev-mpnf-2 | `0dfe94b` | Measure tool behaviour at the point of use; audit the surfaces carrying a copy of a belief |
| `agent-plugins-b9ft` | closed | dev-mpnf-1 | `88eb313` | `pipefail` + `grep -q` orphan-guard flake. Root cause reproduced deterministically |
| `agent-plugins-qozp.6` | closed | dev-mpnf-3 | `b8bcfe2` | Item 6: planning emits a `bd create --graph` plan. Keystone; blocked four beads |
| `agent-plugins-qozp.13` | closed | dev-mpnf-4 | `b6fe2ab` | Item 12: Stop hook exits 2 while `bd ready` is non-empty |
| `agent-plugins-qozp.15` | closed | dev-mpnf-4 | `1d450ed` | Item 14: background monitor on the bead event log |
| `agent-plugins-mpnf.2` | closed | dev-mpnf-7 | `0250403` | **Mid-run amendment.** Rollout plan's graph-plan example had the edge orientation backwards |
| `agent-plugins-qozp.8` | closed | dev-mpnf-5 | `1d3ec35` | Item 8: collisions as edges; **`sprint-wave` deleted outright** |
| `agent-plugins-qozp.9` | closed | dev-mpnf-5 | `d2b97cc` | Item 9: assignment replaces batch bookkeeping |
| `agent-plugins-qozp.7` | closed | dev-mpnf-5 | `f2bd70a` | Item 7: the run-epic shape, flat target, one swarm |
| `agent-plugins-qozp.11` | closed | dev-mpnf-5 | `07105db` | Item 11: merge beads and the merge slot replace the merge-queue prose |
| `agent-plugins-qozp.17` | closed | dev-mpnf-6 | `6892b26` | Ceremony agents export `BEADS_ACTOR` at their dispatch sites |
| `agent-plugins-0913` | closed | dev-mpnf-6 | `cb68439` | Drop `dispatch_state`/`terminal_state` |
| `agent-plugins-qozp.10` | closed | dev-mpnf-6 | `4781f3c` | Item 10: gates for authorization; deferred-review channel. **AC amended — see incident 11** |
| `agent-plugins-qozp.14` | closed | dev-mpnf-6 | `0b6cfa1` | Item 13: SubagentStop hook verifies the worker's queue |
| `agent-plugins-mpnf.5` | closed | dev-mpnf-9 | `e2b69ad` | *Review pass 1.* `sprint-cleanup` force-removed a BLOCKED worker's worktree |
| `agent-plugins-mpnf.12` | closed | dev-mpnf-10 | `78b34e1` | *Review pass 1.* Both hooks' block bound vanished when the counter file could not be written — measured 10 and 8 consecutive blocks against MAX=2 |
| `agent-plugins-mpnf.4` | closed | dev-mpnf-8 | `6ac33c7` | *Review pass 1.* Collision edges cannot be emitted in a graph plan. **The run's own deliverable, non-functional** |
| `agent-plugins-mpnf.3` | closed | dev-mpnf-8 | `7120474` | *Review pass 1.* Merge bead's fan-in edge named an existing bead id: whole plan aborts |
| `agent-plugins-mpnf.24` | closed | dev-mpnf-15 | `bc32a87` | *Review pass 2.* Both hooks keyed `state_dir` off `CLAUDE_PROJECT_DIR`/`PWD` |
| `agent-plugins-mpnf.17` | closed | dev-mpnf-14 | `5c8be0a` | *Review pass 2.* `bb-status` used direct-children-only `bd list --parent` against a nested run epic |
| `agent-plugins-mpnf.16` | closed | dev-mpnf-11 | `a420ab7` | *Review pass 2.* `agent_drained` read a FAILED `bd` query as "drained" |
| `agent-plugins-mpnf.20` | closed | dev-mpnf-11 | `c438c05` | *Review pass 2.* `prune_worktrees` ignored its epic argument and tore down repo-wide |
| `agent-plugins-mpnf.18` | closed | dev-mpnf-17 | `5cc2a58` | *Review pass 2.* The merged diff shipped five new copies of the claim this run retracted; one labelled "this is measured" |

**Totals: 23 closed, 0 blocked, 0 skipped.** Every admitted bead landed.

Three merge beads also closed under `merge-mpnf` and carry no commit of their
own: `agent-plugins-dyfv`, `agent-plugins-zuhy`, `agent-plugins-qq8o`. Three
commits on the branch belong to no bead: `22c86ab` (gitignore `.beads/plans/`),
`9cc8844` (drain the outbox and the run-3 sprint documents into the tracker,
at the owner's instruction), `663abe0` (the retro).

### The five-hour stall

The run stopped dead between review loop 1 and review loop 2 and nothing
noticed.

```
last bead closed before the gap:  agent-plugins-dyfv    2026-08-25T00:15:51Z
first bead closed after it:       agent-plugins-mpnf.16 2026-08-25T06:45:12Z
last commit before the gap:       7120474               2026-08-25T00:14:52Z
first commit after it:            9cc8844               2026-08-25T06:47:30Z
```

**Six hours and twenty-nine minutes**, with eleven groomed beads Ready, zero
worktrees, and nothing dispatched. The coordinator's own account puts the
dispatchable stall at roughly 01:00–06:09.

**The run's health label read `healthy` throughout.** Measured:

```
$ bd show agent-plugins-mpnf   ->  LABELS: health:healthy
```

and the only health event this run ever emitted is `agent-plugins-mpnf.1`,
"State change: health → healthy", at 2026-08-24T21:16:37Z. There is no
`stalled` event. Health was asserted once and never re-derived. Filed as
**`agent-plugins-offx.11`** — health must be derived from progress, not
asserted.

**Nothing was going to wake it.** The run's durable recovery trigger,
`trig_013tVibHWMLTaZkJ6A6woifQ` ("Recover sprint run agent-plugins-mpnf if
interrupted"), was a one-shot: `run_once_at 2026-08-24T21:52Z`,
`last_fired_at 2026-08-24T21:52:40Z`, `ended_reason: run_once_fired` — it fired
**before** the stall and disabled itself. Its `next_run_at` field still reads
`2026-08-25T21:52:39Z`, which is 15h43m after the stall ended and is not a
scheduled firing on a trigger carrying that `ended_reason`. On either reading
the liveness mechanism was useless: it had either already retired or would next
speak fifteen hours late. **What actually resumed the session was the owner's
PR review comments.** Filed as **`agent-plugins-offx.15`**.

### The cleanup's collateral: 25 dangling citations

Commit `9cc8844` deleted the four 2026-08-24 `rfcs/sprints/` documents at the
owner's instruction. **25 beads cited them; 21 of those are open.** Found by
the retro, verified independently by the coordinator:

```
$ bd export > e.jsonl
$ jq -r 'select((.description//"")+(.notes//"")+(.design//"") | test("rfcs/sprints/2026-08-24")) | .id' e.jsonl | wc -l
25
$ ... | select(.status!="closed") | .id' e.jsonl | wc -l
21
$ ls rfcs/sprints/ | grep 2026-08-24   ->  no output
```

Four distinct dead paths. **Worst instance: `agent-plugins-r9yj` (open, P3) —
its cited text *was* the whole of the bead's content**, leaving an open bead
whose work could not be performed from the bead. Repaired by inlining the
entries. Also repaired: `9owj`, `offx.2`, plus annotations on the `so39` and
`qozp` epics whose `report_path` manifest keys point at deleted files. **18 `y63n.*`
harvested beads** carry the citation inside verbatim quoted text — measured,
`jq -r 'select(.id|startswith("agent-plugins-y63n")) | select(…test("rfcs/sprints/2026-08-24"))|.id' | wc -l` → 18 — and their epic
carries an explicit hazard note instead. Filed as
**`agent-plugins-y63n.18`** — a bead whose only evidence is a document path
does not survive that document.

---

## Where agents did the hard correct thing

This run's failures are structural. Its agents were not the failure, and
several of them did the expensive thing when a cheap one was available.

**`mpnf.20` mutation-tested its own new assertions.** Not "the tests pass" —
three mutations of the shipped code, each reverted immediately, each confirmed
to break the assertion meant to catch it:

```
1. drop --all from the walk        -> FAIL scope, a nested epic's GRANDchildren are all found
2. stop pushing epic children      -> FAIL scope, expected [c1 d1 g1 tgt], got [d1 tgt]
3. scope_beads filters nothing     -> FAIL scope, got [g1, c1, acme-foreign]
```

Mutation 3 reproduced the exact data loss the bead exists to prevent. Before
the fix, the negative control passed on the *unfixed* script — `ok UNSCOPED
(the bug), the same foreign worktree IS removed` — so the assertions are known
to fail when they should.

**`mpnf.17` ran a negative control with descent disabled** and confirmed
precisely the two nested assertions fail and the three flat ones pass. It then
**volunteered a defect it had itself introduced**: its edit moved the line
`mpnf.18`'s finding cited, from `bb-status.sh:459` to `:560`, and it said so
with both greps rather than leaving the next agent to hit it. That observation
became `y63n.22` — *file:line anchors rot before dispatch; the string is the
durable handle.*

**`mpnf.24` removed a dependency it could not verify** rather than guessing.
It could not spawn a real harness-fired hook, so instead of asserting what
`CLAUDE_PROJECT_DIR` carries there, it made `state_dir` not depend on the
variable at all when a git repo is present — and said plainly which half stayed
unverified and what would settle it.

**`mpnf.18` ran eleven configurations with a positive control**, so that "`[]`
everywhere" could not be equally explained by a broken query. It also found
that `bd human list --json` genuinely *does* return `null`, establishing that
bd's empty-JSON shape is per-subcommand — the most likely origin of the false
claim, and a thing the retraction explicitly refused to sweep untested.

**`dev-mpnf-8` challenged a fact the coordinator handed it** rather than
working around it. Four configurations, plus a control proving the command
works at all. That is the only reason the false claim was caught before it
could be copied further.

Two more worth naming: **`b9ft`'s agent** isolated an intermittent bug into a
deterministic three-line reproduction and then audited the whole file for the
same shape, finding two more instances; **`qozp.9`'s new self-test caught a
real bug in its own subject within one run** — a shell function clobbering a
caller's global.

---

## Implicit decisions

Choices a diff reader would not recognise as choices. Incidents are not
repeated here.

| Bead | Decision |
|---|---|
| `x5tk` | The four instances are written in **ledger form** (claim / measurement / where stated), not narrative, because `AGENTS.md` carries two rules that pull opposite ways. The same rule is stated three times in three registers, deliberately |
| `qozp.6` | The 54-byte budget was met by **losslessly tightening three existing paragraphs, not by cutting a rule** — a diff reader sees edits to prose the bead has no other reason to touch. Also: the bead's own Design called `parent_key` and `parent_id` interchangeable; they are not, and the wrong one fails silently |
| `qozp.8` | How a collision edge is marked, and how a batch survives the removal of waves. The acceptance box named five files; the real edited set is **30 tracked files**, re-derived by grep at the point of use |
| `qozp.9` | A worktree is matched to its owner **through its checked-out branch**, not a stored path — which is what let `metadata.worktree` go without losing the mapping |
| `qozp.10` | `--exclude-label human` goes on the dispatch front and **deliberately not on every ready query in the tree** |
| `qozp.11` | The custom-type question was closed in grooming (`task` + label `merge`; do not run `bd config set types.custom`). Recorded, not re-opened |
| `qozp.13` | Two files outside the bead's named scope were edited, both forced by the file the bead creates |
| `qozp.14` | Four choices a diff reader would not see as choices; documentation placed in `references/supervision.md` beside the Stop hook, since the bead named a home for the file and not for the prose |
| `qozp.15` | Three such choices; and the bead's inherited claim that the monitor subsumes the durable recovery trigger's interval question **is false** — it does not |
| `qozp.17` | The identity rule is stated in **five different lengths** across the ceremony skills, and the variance is a byte budget, not an editorial choice |
| `mpnf.3` | Cited the planner **by role rather than by path** in `worktree-mechanics.md`, deliberately not restating the sidecar mechanism there |
| `mpnf.4` | Byte budget spent to **5 bytes free**, not blocked, measured with the checker's own method before and after |
| `mpnf.16` | `handle_worktree`'s three-way dispatch is a `case` on a captured return, not a chain of `if agent_drained` — because bd 1.2.2 distinguishes "nothing to report" from "cannot answer" by exit status |
| `mpnf.18` | The replacement sentences **do not merely swap `null` for `[]`**. Swapping alone would leave defensive code justified by a measurement that no longer supports it |
| `mpnf.24` | Two ACs were already true of pre-existing code; one AC is not literally true after the fix, because each hook retains an unrelated `CLAUDE_PROJECT_DIR` line in `mode_hook` |

---

## Routing evidence

**Tier escalations: none.** No bead's tier was raised to get it done.
`CLAUDE_CODE_SUBAGENT_MODEL` is unset, so tier routing did occur and the null
is a real result.

**Deliberate over-routes: none recorded.** No bead was knowingly run a tier
above its assigned one.

Assigned tiers, from each bead's `metadata.tier`: **11 `advanced`, 12
`standard`, 0 `light`, 0 `frontier`** across the 23.

**Read the zero carefully.** Escalations reveal *under*-routing only. A bead
groomed `advanced` that a `standard` model would have finished never appears
here — it succeeds, and the overspend is silent. **Zero escalations across a
23-bead queue is not evidence of good calibration; it is equally consistent
with routing everything too high.** No bead in this run was assigned `light`,
which on a queue containing single-file prose corrections is itself a signal
worth testing. The next run should try one tier lower on the beads whose
`tier_reason` was thinnest.

---

## Follow-up beads

### Admitted into the run

`mpnf.2` (amendment), `mpnf.3`, `.4`, `.5`, `.12` (review pass 1), `mpnf.16`,
`.17`, `.18`, `.20`, `.24` (review pass 2). **All ten closed** — rows above.

### Left for the user

Filed mid-run with `discovered-from` links, awaiting triage:

| Bead | P | What |
|---|---|---|
| `agent-plugins-89xd` | P2 | **Nothing in the dispatch path arms the Stop ready-loop hook.** The hook is inert until something calls `--arm` |
| `agent-plugins-w7x8` | P2 | `bd merge-slot release` is not owner-protected — any actor can release another's slot |
| `agent-plugins-6dbb` | P2 | Audit remaining scripts for `pipe \| grep -q/head/tail-n` status reads under `pipefail` (found in `session-start.sh` and three bead-bonanza scripts) |
| `agent-plugins-igf5` | P2 | `triage-issue` verifies that quoted evidence still exists, never that a tool claim in it is true |
| `agent-plugins-ddks` | P2 | `bead-bonanza/SKILL.md` still orders a Monitor armed per wave, after waves were deleted |
| `agent-plugins-0liz` | P2 | Two `.sf/outbox` issues propose fixes to `sprint-wave`, which no longer exists |
| `agent-plugins-jca7` | P3 | ADR-011 frames the problem as "orchestrated bead waves"; the mechanism is now continuous dispatch |
| `agent-plugins-fvnp` | P3 | `beads-coordination-design.md` §4a's graph-plan example lacks edge-orientation prose |
| `agent-plugins-mxl3` | P2 | `sprint-cleanup` from inside a run worktree treats it as main and never tears it down — **declined as an amendment** |
| `agent-plugins-p567` | P1 | Orchestrator loads only the graph plan — **admitted, then de-admitted** |

### The epic this run produced

**`agent-plugins-offx`** — "Swarm capability CLI: replace prescribed command
sequences with a tested action surface", 15 children. This is the run's real
deliverable. The four P1s:

- `offx.1` capability audit: every action the skill family performs, and the commands behind it
- `offx.2` FRD for the CLI, derived from the audit
- `offx.11` health must be **derived** from progress, not asserted
- `offx.14` scope the CLI as a **thin capability layer over `bd` and `git`, inside existing harnesses** — not a new orchestrator
- `offx.15` the recovery trigger slept through the stall it is named for; two orphans plus 14 CI triggers still armed

`offx.3` is the one that states the thesis outright: *retire `AGENTS.md`'s
"Tool claims are measured at the point of use" section once claims are tests.*
A rule that exists because nothing can check it is a rule waiting for a
checker.

---

## Stashes

```
$ git stash list
```

**None.** No recovery stash was created.

---

## Retro

```
Retro: 12 filed, 2 declined, 0 routed to .sf/outbox/
       52 harvested, 7 clusters deduped into beads that already exist,
       2 unrouted
Every finding, verbatim: docs/research/sprint-findings/2026-08-25-agent-plugins-mpnf.md
```

The retro appended its own row to `rfcs/retro-log.md` (line 33). No row is
added by this report.

**`retro_outbox` is 0 because the channel no longer exists.** Commit `9cc8844`
removed the 15 `.sf/outbox` issue documents at the owner's instruction, so the
eight skill-audience issues were filed as beads under `agent-plugins-y63n`
(`.18`–`.25`). The findings document went to `docs/research/` rather than
`rfcs/sprints/`, per `CLAUDE.md`'s where-documents-live rule — it decides
nothing, which is that rule's test. **`sprint-retro` hard-codes both retired
paths as literal strings and offers no way to redirect them**; the dispatch had
to override the skill in prose twice. Filed as `y63n.19`.

**Two findings unrouted — the third consecutive run at exactly 2.** Runs 1 and
2 each ended at 2, `y63n.17` already records that the count is not falling, and
this run makes three. This time the unrouted pair is `mpnf#5` (`retro:process:`)
and `mpnf#6` (`retro:incident:`) — the second being the worktree-removal
incident described above.

Four project beads filed: `agent-plugins-sdu1` (P2), `agent-plugins-eejf` (P2,
shellcheck is named in acceptance criteria, is not installed, and is in no
declared check), `agent-plugins-bfng` (P3), `agent-plugins-lmrl` (P3).

---

## Validation, and why there is no CI evidence

**`checks_baseline_status: 0`** — all 13 declared checks passed on the untouched
tree before any bead ran.

**Re-measured on head `663abe0` for this report:**

```
$ python3 plugin/skills/bead-bonanza/scripts/bb-checks.py --report checks.tsv
exit 0
format 0 · vet 0 · build 0 · test-go 0 · skills 0 · shared-refs 0 · strata-docs 0
schemas 0 · corpus 0 · ref-links 0 · template-drift 0 · script-self-tests 0 · skills-strict 0
```

**13 of 13, exit 0. This is the only validation signal the run has.**

### GitHub Actions was down at the account level for the entire run

**No CI result was obtainable for any commit in this run.** Measured here, not
inherited:

```
$ actions_list list_workflow_runs --branch claude/bead-bonanza-design-review-ne1plu
```

Every workflow run on this branch across the run window — 30 inspected, both
the `checks` and `cli` workflows — reports `conclusion: failure`, and each one
completes **3 to 7 seconds after it is created**. Drilling into the newest:

```
run 32821773148 (checks, head 663abe0)
  created_at 2026-08-25T07:28:33Z   updated_at 2026-08-25T07:28:37Z
  job "this repository's own specs conform"      created 07:28:34 started 07:28:34 completed 07:28:35
  job "skill invariants and vendored references" created 07:28:34 started 07:28:34 completed 07:28:35
```

**One second per job.** A job that starts in the second it is created and dies
in the next has executed no body; `get_check_run` returns empty `title`,
`summary` and `text` for the same reason. This is an account-level outage,
running since approximately 02:26 UTC on 2026-08-24 — roughly 29 hours — and it
covers every commit this run pushed.

**What this costs the reader:** the red X on PR #27 is the outage, not the
branch. But "the local checks pass" and "CI is green" are not the same
evidence, and no cross-platform signal exists for this branch at all — the
release bar (`.github/workflows/`, the cross-compile matrix, the three-platform
test matrix, the busybox smoke test) never ran. **A human must re-run CI before
treating this branch as validated.**

### Pushes

```
$ git rev-parse HEAD                                              -> 663abe013814724aafb28101cd9f46e284d1fabb
$ git rev-parse origin/claude/bead-bonanza-design-review-ne1plu   -> 663abe013814724aafb28101cd9f46e284d1fabb
$ git log --oneline origin/claude/...ne1plu..HEAD                 -> (empty)
$ git status --porcelain                                          -> (empty)
```

Every closed bead's commit is on the remote. `push_mode` is `defer_push`
(human-set, and `sf.toml` agrees); the coordinator pushes. `pull_request` is
`none` — PR #27 already tracks the branch, so this run opened none and this
report modified none. `merge_to_default` is `false`.

### Durable trigger

**Both this run's durable triggers self-disabled; neither needed cancelling.**

| trigger | armed for | disposition |
|---|---|---|
| `trig_013tVibHWMLTaZkJ6A6woifQ` | "Recover sprint run agent-plugins-mpnf if interrupted" | fired `2026-08-24T21:52:40Z`, `ended_reason: run_once_fired` |
| `trig_011pnBEDQaoVWXiAQtYtHt6Y` | "Re-check PR #27 CI after the Actions outage" | fired `2026-08-25T00:51:52Z`, `ended_reason: run_once_fired` |

Neither appears in a default `list_triggers` listing, which confirms both are
retired one-shots. **Both carry a snapshot prompt written at arming time and
never re-read** — the second one says so about itself, in its own first
paragraph, citing `agent-plugins-x5tk`. `offx.15` records that two orphaned
triggers and 14 CI triggers remain armed elsewhere on this account; **that
sweep is not this report's and has not been done.**

---

## What this run could not verify

Named, because an unstated gap is indistinguishable from one that does not
exist.

- **CI, for every commit.** Above. The local `bb-checks.py` result is the whole
  of the run's validation evidence.
- **`shellcheck`, on every shell change.** `command -v shellcheck` returns 1 in
  this container, and no declared check invokes one. Three beads (`mpnf.16`,
  `.20`, `.24`) carry an acceptance criterion naming it; each recorded the half
  it could not exercise, and `sh -n`/`bash -n` ran clean in place of it. Filed
  as `agent-plugins-eejf`. **Every shell change this run merged is unlinted.**
- **What `CLAUDE_PROJECT_DIR` carries inside a real harness-fired hook.**
  `mpnf.24` could not spawn one and made no claim about it; it removed the
  dependency instead. Settled by arming from a repo under a multi-root parent
  and invoking the hook through the harness.
- **Whether the SubagentStop payload actually carries an agent id.** `qozp.14`
  could not verify it; a prior comment on the bead asserts it does, and that
  assertion was not measured. `mpnf.6` (deferred) records that the hook queries
  the harness's id while beads carry the derived dispatch id, so it **can never
  block** — a defect that stands unfixed.
- **`qozp.15`'s acceptance criterion 2** — "closing a bead delivers a line to
  the session with nothing having armed a watch" — cannot be satisfied end to
  end from this environment. The agent said so rather than claiming it.
- **The eleven review-pass-2 remedies.** The budget was 2 passes and both were
  spent. Nothing reviewed them. A regression introduced by any of the five that
  merged would not have been caught.
- **`p567`'s absence.** The sidecar's orchestrator half is not shipped. An
  orchestrator following the merged text creates beads, wires no edges, and
  reports success. This is known, P1, and deliberately left.
- **The 18 `y63n.*` beads' dangling citations.** Repaired by an epic-level
  hazard note rather than by editing the verbatim quoted text each carries.
  Whether a triaging agent reads that note before acting on a bead is untested.
- **`bd`'s empty-JSON shape beyond `bd ready` and `bd human list`.** Both were
  measured; nothing generalises to a third subcommand, and this run has now
  been wrong about that shape twice.
- **Two of the retro's own findings routed to no audience**, and the run's
  most consequential incident record reached one only because a human read it.

---

## Reading the record

**Run epic: `agent-plugins-mpnf`.** It carries the manifest (`admitted`,
`excluded` with a reason on every entry, `ledger`, `standing_auth`,
`checks_baseline`, `scope_notes`), 30 comments holding every incident and
amendment quoted above, and `findings_path`.

```
bd show agent-plugins-mpnf
bd list --parent agent-plugins-mpnf --status all
bd show agent-plugins-offx        # what the run decided to do instead
```

**Successor: `agent-plugins-offx`, 15 children.** Start at `offx.1` (the
capability audit) and `offx.14` (scope it as a thin layer, not a new
orchestrator).


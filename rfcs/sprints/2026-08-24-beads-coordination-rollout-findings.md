# Sprint findings — Roll coordination onto the beads graph (items 1–14)

Run epic `agent-plugins-qozp` · branch `claude/bead-bonanza-design-review-ne1plu`
· base `aba7299` · head `9a88c6f` · PR #27.

Six beads admitted, six closed: `qozp.1` `c25818e`, `qozp.4` `05af5b6`,
`qozp.2` `8a8b86f`, `h3j8` `872f390`, `qozp.5` `9a88c6f`, and `qozp.3` closed
`declined-no-diff` with no commit by design. Two waves. Review loop closed at
pass 1 of a 2-loop budget with 0 of 11 findings admitted.

Model lineup from the manifest — light `haiku`, standard `sonnet`, advanced
`opus`, frontier `fable`. Effort: the manifest records `null`; this harness
exposes no effort setting, so every issue below carries that verbatim rather
than a guess.

**What this run was for.** It was a trial of the `bead-bonanza` skill family on
a queue whose subject matter is also coordination tooling. Findings about the
skills are the primary output, not a byproduct — 8 of 33 findings carry
`kind: tooling`, and several of the 13 project-audience findings are about the
authoring constraints those skills are written under.

---

## 1 · Harvest

One query over the run's closure set — the epic and its descendants, prefix
`agent-plugins-(qozp|h3j8)`:

```sh
bd export | jq -r 'select(.id | test("^agent-plugins-(qozp|h3j8)"))
  | .comments[]? | select(.text | startswith("retro:")) | .text'
```

**33 findings** across 7 beads. Every bead that ran recorded at least one;
`qozp.1` recorded a single "clean — no findings beyond scope" note, which is
the correct behaviour for a bead that hit no friction.

| bead | findings | tier |
|---|---|---|
| `agent-plugins-qozp` (epic, coordinator) | 10 | — |
| `agent-plugins-qozp.2` | 7 | advanced (`opus`) |
| `agent-plugins-qozp.3` | 5 | advanced (`opus`) |
| `agent-plugins-qozp.5` | 4 | standard (`sonnet`) |
| `agent-plugins-qozp.4` | 3 | advanced (`opus`) |
| `agent-plugins-h3j8` | 3 | standard (`sonnet`) |
| `agent-plugins-qozp.1` | 1 | standard (`sonnet`) |

### Two records the harvest does not reach — measured, not inferred

The harvest is `startswith("retro:")`; `sprint-report`'s incident harvest is
`grep -E '^(retro:[a-z]+:[^ ]+ incident|amend: incident)'`. Both were run over
this run's comment set. Two comments match **neither**:

| comment | first 60 chars |
|---|---|
| `agent-plugins-qozp` #19 | `incident: review pass 1 filed 11 findings and the run admitt…` |
| `agent-plugins-qozp.18` #1 | `coordinator verification (review-qozp's finding independentl…` |

The first is the run's single most consequential record — a confirmed P0
silent-data-loss path in code this run merged, and the gate decision that
deferred it out. The second is the coordinator's independent confirmation of
that defect, carrying an addition the original finding did not have (the same
clobber pattern is pre-existing at `pre-flight-check.sh` lines 653–672). Both
were recovered here only because the coordinator's handoff named them; a retro
running from the epic id alone, as this skill is designed to, would have lost
both. This is finding-shaped in its own right and becomes candidate **S12**.

## 2 · Routing

A lookup on `kind`, with no finding's content read:

| kind | count | destination |
|---|---|---|
| `tooling` | 8 | **skill** audience → `.sf/outbox/<skill>/` |
| `proof` | 9 | this document, beside what it supports. No issue. |
| `environment` | 1 | the project's environmental-failure ledger. No issue. |
| `gap` | 4 | **project** audience |
| `decision` | 6 | **project** audience |
| `stale` | 2 | **project** audience |
| `scope` | 1 | **project** audience |
| unrouted | 2 | counted and listed, never dropped |
| **total** | **33** | |

`proof` at 9 of 33 (27%) and `environment` at 1 — together 10, roughly a third
of the harvest terminating here with nothing for a human to decide. That is
consistent with the protocol's own note that this is not a rare path.

### Unrouted — 2 of 33 (6%)

The taxonomy's health metric. Both are the coordinator's:

- **`agent-plugins-qozp.1#1`** — prefix `retro:` with no kind. It is a
  no-findings declaration ("clean — no findings beyond scope"), not a finding
  the taxonomy failed to hold. Arguably the protocol's "record nothing" would
  have been better, but a positive statement that a bead was clean has value
  the silence does not.
- **`agent-plugins-qozp#8`** — `retro:process:wave-1-bypassed-sprint-wave`.
  `process` is not one of the nine kinds. The correct kind was `tooling` with
  target `bead-bonanza`; it is routed to the skill audience below on the merge
  test, and is the evidence behind candidates **S1** and **S3**.

One invented kind and one kindless note. Nothing here suggests the nine kinds
are drifting from what agents hit; re-check the ratio next run.

### Merge test applied where the kind did not settle the audience

This repository **vends the skills it runs**, so "project" and "skill" are not
"us" versus "them" — they are *where the file lives*. A fix landing in
`plugin/skills/<skill>/` goes to the outbox for that skill; a fix landing in
`scripts/`, `docs/`, `specs/`, `cli/` or `AGENTS.md` becomes a bead here.

Three findings needed the test rather than the kind:

- **`qozp#6`** (`gap:run-record`) — the fix lands in
  `plugin/skills/sprint-plan/references/run-record.md` and in
  `plugin/skills/bead-bonanza/scripts/bb-status.sh`. **Skill**, both, and two
  separate changes: candidates **S6** and **S7**.
- **`qozp#10`** (`gap:check-skills-rule-3`) — the fix lands in
  `scripts/check-skills.sh` and this repository's authoring guidance.
  **Project**: candidate **P1**.
- **`qozp.2#6`** (`tooling:agent-plugins`) — the fix lands in the Claude Code
  harness, which is neither. The actionable residue is a ledger entry:
  candidate **P4**.

## 3 · Dedupe — five findings already filed

Not candidates. Each already has a bead, and the finding is annotated here so
the audit closes in both directions.

| finding | already filed as | state |
|---|---|---|
| `qozp.2#5` — 7 of 8 roster agent kinds have no dispatch site | `agent-plugins-qozp.17` | open, blocked by `qozp.7` |
| `qozp.3#5` — the repo's BEADS INTEGRATION marker pair is split across two files | `agent-plugins-vqmc` | filed, needs an owner decision |
| `qozp.3#2` — `bd setup claude` reports success while writing nothing | `agent-plugins-2nb8` | filed as an upstream report |
| `qozp.4#2` — rollout item 2 specifies `concurrency_cap` but names no writer | `agent-plugins-h3j8` | closed `872f390` (the instance); the class becomes **P3** |
| the manifest-clobber path | `agent-plugins-qozp.18` | deferred, `out-of-scope-this-run` |

`agent-plugins-qozp.18` is worth a note: it is a **skill-audience** defect (in
`plugin/skills/bead-bonanza/scripts/pre-flight-check.sh`) filed as a bead in
this tracker by `sprint-review`. That is right *for this repository*, which
vends the skill, and would be wrong in a project that does not. Nothing to fix;
worth knowing before someone "corrects" it.

The other ten review findings — `agent-plugins-qozp.19` through `.28` — are
likewise filed and deferred. They are not re-derived here.

## 4 · Environment → the ledger

`environment` findings produce no issue. They belong in `AGENTS.md`'s
environmental-failure ledger, where the next agent will look. Three entries are
ready to paste; landing them needs an owner, which is candidate **P4**.

**1. `bd setup claude`'s "stale" verdict is conditional on `sync.remote`, not
on template drift.** From `qozp.3#4`, measured on bd 1.2.2 in a throwaway
project: with no `.beads/config.yaml`, the emitted block omits `bd dolt push`
and the marker hashes `6cd5cc61`; with `sync.remote` set it includes
`bd dolt push` and hashes `970c3bf2`. This repository's marker carries
`6cd5cc61`, i.e. it was installed before the remote was configured. A
consequence worth its own line: `bd -C <dir> setup claude --check` gives a
verdict about the **caller's cwd** file computed against `-C`'s config — the
same binary called the same file stale and then current, one command apart.

**2. The Bash worktree-isolation guard refuses plainly in-worktree commands,
non-deterministically.** From `qozp.2#6`. `bd update <id> --set-metadata
dispatched_at="$(date -u +%FT%TZ)"` — no git call in it at all — was refused
with "this command is too complex to verify that it stays inside the
worktree." A heredoc-plus-`chmod`-plus-`echo` was accepted once and refused
twice on different script contents. The guard's own advice ("split it into
plain, separate commands") was followed and still refused. **The pattern that
works: write files with the Write tool, run them with a single bare command.**
Cost on that bead alone: five extra round trips.

**3. GitHub Actions is down at the account level.** Jobs die in 2–4 seconds
with no log content. This is a platform state, not a defect in any workflow in
this repository, and CI results were unobtainable for this run's commits.
*Evidence is the coordinator's observation reported at retro handoff, not a
captured `retro:` finding* — flagged rather than dressed up, because the ledger
admits entries on evidence.

## 5 · Candidate issues

Sixteen, clustered from the 33 findings and deduped against what is already
filed. Each clears the bar: it changes how the next run goes. Every one carries
`recurrence: 0` — the outbox is empty and `rfcs/inbox/` carries no prior issue
against any of these targets, so each is the first run in the window to report
its kind/target pair. That is not a defect; it is the difference between a real
recurring problem and one bad afternoon, and these become informative from the
second run that reports them.

Provenance shared by all sixteen: `run: agent-plugins-qozp`,
`branch: claude/bead-bonanza-design-review-ne1plu`, `report:` this document,
`effort: <not recorded — this harness exposes none>`.

### Skill audience — 12, destined for `.sf/outbox/<target>/`

---

#### S1 · `sprint-wave` absorbs the batch when it has no dispatch tool

`audience: skill` · `target: sprint-wave` · `priority: 1` ·
`findings: ["qozp#9", "qozp#8"]` · `model: claude-opus-5 (coordinator)`

**What is wrong, and what it costs.** `sprint-wave`'s whole premise is
"dispatches a subagent per batch". In wave 2 no Task/Agent tool was exposed to
it, and instead of stopping it created the worktree itself and ran
`sprint-batch` → `run-bead` inline. It degraded silently. The cost is the
skill's entire reason for existing: the verification it performs against an
independent agent's account became self-verification, so wave 2's returns were
checked by the agent that produced them. A run cannot tell from the outside
that this happened.

**Proposed change.** Make a dispatch-tool probe `sprint-wave`'s first act,
before it creates a worktree or reads the batch. Where no dispatch tool is
available, stop and return a named blocker — the wave is undispatchable — and
do not perform the batch's work. Done: a `sprint-wave` invoked with no dispatch
tool returns a blocker and leaves the tracker and the working tree untouched.

---

#### S2 · `pre-flight-check.sh`'s no-agent probe answers only for the coordinator

`audience: skill` · `target: bead-bonanza` · `priority: 2` ·
`findings: ["qozp#9"]` · `model: claude-opus-5 (coordinator)`

**What is wrong, and what it costs.** `pre-flight-check.sh` already models a
missing dispatch capability as `BB_ASSUME_NO_AGENT=1` and blocks a run on it.
It ran clean here, because it ran in the **coordinator's** session, where the
Agent tool is available — and `sprint-wave`, dispatched as a subagent, had no
such tool. The capability is per-agent, not per-session. The cost is a
run-level gate that reads green while the condition it exists to catch is
present two levels down, which is worse than no gate: it licenses the run to
proceed.

**Proposed change.** State in the pre-flight's own output and in
`references/scripts.md` that `BB_ASSUME_NO_AGENT` is scoped to the session that
ran the probe and does not clear any subagent, and require each dispatching
agent to re-probe in its own session before it dispatches. Done: the pre-flight
emits the scope caveat alongside the result, and the reference no longer reads
as a run-level clearance.

---

#### S3 · Direct dispatch instead of `sprint-wave` is undocumented and costly

`audience: skill` · `target: bead-bonanza` · `priority: 3` ·
`findings: ["qozp#8"]` · `model: claude-opus-5 (coordinator)`

**What is wrong, and what it costs.** Wave 1's four dispatches were made by the
coordinator directly, with harness Agent calls, rather than by invoking
`sprint-wave`. Two costs fell out and nothing in `bead-bonanza` warns of
either: worktrees and branches were named `worktree-agent-<harness-id>` rather
than by dispatch id, so "presence is the in-flight signal" still worked but
"named for its dispatch id" did not, and the coordinator had to map harness ids
to bead ids by hand; and `sprint-wave`'s verification was performed by the
coordinator, which is precisely the reasoning the coordinator is designed not
to hold. The run carries both paths, so the comparison is measured rather than
argued.

**Proposed change.** In `bead-bonanza`'s wave step, state that a wave is
dispatched by invoking `sprint-wave` and that direct dispatch is a deviation,
naming its two costs; and give the deviation a fallback — where worktree names
come from the harness, record the harness-id → dispatch-id mapping in the
dispatch record at dispatch time, not by hand afterwards. Done: the wave step
names the deviation and the mapping requirement.

---

#### S4 · The checks-baseline inlining rule does not scale past one agent

`audience: skill` · `target: bead-bonanza` · `priority: 2` ·
`findings: ["qozp#1"]` · `model: claude-opus-5 (coordinator)`

**What is wrong, and what it costs.** `references/prompt-per-bead.md` requires
the run epic's `checks_baseline` be "inlined verbatim" into every subagent
prompt, with an explicit warning against sending the one-word status instead.
On this run the baseline text is ~6 KB and wave 1 dispatched 4 agents, so
obeying the rule literally costs ~24 KB of duplicated prompt — a cost that
grows linearly with the wave and is paid before the agent has done anything.
The coordinator deviated rather than pay it, which means the rule as written is
one an orchestrator will break under load.

**Proposed change.** Replace the verbatim-inlining requirement with: inline the
**verdict** (e.g. "all 13 declared checks passed, exit status 0" plus "any
failure you see is yours") and the exact command that reads the full text from
the manifest. Keep the distinction the current rule protects by requiring an
explicit "no baseline was taken" line where none exists, so a missing block is
never ambiguous with an unstated one. Done: `prompt-per-bead.md` states the
verdict-plus-command form and the explicit no-baseline line.

---

#### S5 · `git worktree remove` cannot remove a locked worktree, so the guard is lost

`audience: skill` · `target: bead-bonanza` · `priority: 2` ·
`findings: ["qozp#2"]` · `model: claude-opus-5 (coordinator)`

**What is wrong, and what it costs.** `references/worktree-mechanics.md`
mandates raw `git worktree remove` over `bd worktree remove` specifically to
preserve the dirty-tree check, arguing that `bd`'s version can only proceed via
`--force`, which skips every check. But a harness-created worktree is
**locked** — `git worktree list` showed `locked` on all four of wave 1's — and
plain `git worktree remove` refuses a locked worktree. The orchestrator is
forced to `--force` anyway, losing exactly the guard the rule was written to
keep. The rule silently converts into its own opposite. A clean counter-case is
on the record: wave 2's `sprint-wave`-created worktree
`/home/user/.bb-worktrees/dev-qozp-5` removed **without** `--force`, so the
lock comes from the harness, not from the skill's mechanics.

**Proposed change.** In the removal step, state that a locked worktree is
unlocked first — `git worktree unlock <path> && git worktree remove <path>` —
and that `--force` is not the escape hatch for a lock, because it discards the
dirty-tree check the step exists to run. Done: `worktree-mechanics.md` carries
the unlock-then-remove form and names `--force` as the wrong reach.

---

#### S6 · `run-record.md` documents the manifest's shape but not how to write it

`audience: skill` · `target: sprint-plan` · `priority: 1` ·
`findings: ["qozp#4", "qozp#6"]` · `model: claude-opus-5 (coordinator)`

**What is wrong, and what it costs.** `references/run-record.md` shows the
manifest's `admitted`, `excluded` and `models` keys as JSON arrays and objects,
and says nothing about writing them. `bd`'s `--set-metadata` stores a value as
a **string**; arrays need `--metadata`, which **replaces** the whole object and
therefore needs a read-modify-write. `groom-beads/references/capability-tier.md`
documents that trap for capability tiers, in a different skill, where an
orchestrator writing a manifest will not read it. Separately, the `excluded`
entry shape is `[{"id": …, "reason": …}]` — one id per entry — and the natural
thing to write when several beads share a reason is
`[{"ids": […], "reason": …}]`, which is valid JSON and wrong. Both cost this
run: the manifest was written with `--set-metadata`, stored as strings, and the
run's own status report became unreadable (see S7).

**Proposed change.** Add to `run-record.md`'s manifest section: array and
object keys are written with `--metadata` and a read-modify-write merge, never
`--set-metadata`, which stores a string; and show the `excluded` entry as one
id per entry with an explicit "not `ids`" note against the grouped form. Done:
an orchestrator can write every documented manifest key correctly from
`run-record.md` alone.

---

#### S7 · `bb-status.sh` renders a mistyped manifest key as plausible wrong output

`audience: skill` · `target: bead-bonanza` · `priority: 1` ·
`findings: ["qozp#4", "qozp#6"]` · `model: claude-opus-5 (coordinator)`

**What is wrong, and what it costs.** Given `admitted` stored as a string
rather than a list, `bb-status.sh` iterated it **character by character** and
reported `FAIL 116 admitted bead(s) not closed: [ " a g e n t - p l u g i n s …`
and `OK 347 bead(s) EXCLUDED at admission` — 116 and 347 being character
counts. Given `excluded` entries carrying `ids` rather than `id`, it read
`.id`, found nothing, and printed `None (reason); None (reason); None
(reason)`. In both cases the script produced confident, plausible-looking
output over a shape error. The cost is that the real cause is unguessable from
the report, and a reader who does not already know the trap will debug the
wrong thing.

**Proposed change.** Type-check each manifest key before use: a string where a
list is expected, or a list entry missing `id`, is a `FAIL` naming the key and
the observed type, not an iteration. Done: `bb-status.sh` run against a
manifest with `admitted` stored as a string reports a shape error rather than a
character count.

---

#### S8 · `bb-status.sh`'s human-queue check is repo-wide, not run-scoped

`audience: skill` · `target: bead-bonanza` · `priority: 2` ·
`findings: ["qozp#5"]` · `model: claude-opus-5 (coordinator)`

**What is wrong, and what it costs.** The check runs bare `bd human list` and
failed the run's health with "human queue: 1 item(s) awaiting a decision" for
`agent-plugins-qozp.12`, a bead that is **deferred and explicitly excluded at
admission**. Every other check in the same report scopes correctly — "bd lint:
nothing flagged in this run's queue", "every ready bead carries a tier". The
cost is that a run can never read clean while any deferred decision exists
anywhere in the tracker, which trains the operator to ignore the signal.

**Proposed change.** Scope the human-queue check to the run's admitted set, the
way the neighbouring checks already are. Done: a human-needed item outside
`manifest.admitted` does not fail the run's health.

---

#### S9 · `bb-status.sh bead` fails a bead that legitimately produced no commit

`audience: skill` · `target: bead-bonanza` · `priority: 2` ·
`findings: ["qozp#7"]` · `model: claude-opus-5 (coordinator)`

**What is wrong, and what it costs.** `agent-plugins-qozp.3`'s own acceptance
permitted a decline branch — "OR the bead closes as declined with an upstream
report filed as a bead and linked to this one" — and it took it: `bd setup
claude` wrote nothing, so there was no diff to commit. Signal 2 ("commit: none
found") has no way to express that, so a correct outcome reads as repairable
and the coordinator has to override the script by hand. The cost is an operator
learning to override a failing signal, which is how a real failure gets waved
through next time.

**Proposed change.** Have signals 2 and 3 read `metadata.outcome`: where it is
`declined-no-diff` (set on `qozp.3` this run), report OK-by-declaration naming
the outcome rather than `FAIL`. Done: `bb-status.sh bead` on a bead carrying
`outcome=declined-no-diff` exits 0 and says why.

---

#### S10 · `sprint-review` files findings into the ready front the gate just refused

`audience: skill` · `target: sprint-review` · `priority: 0` ·
`findings: ["qozp#9 (context)"]` · `model: claude-opus-5 (coordinator)`

**What is wrong, and what it costs.** Section 3 files every finding with
`bd create … --parent {EPIC}`, which lands it **open**. Section 4 then
classifies it, and the conservative default is out of scope. But an open bead
parented to the run epic is in `bd ready --parent <epic>`, so all 11 of this
run's review findings entered the run's ready front on creation — including a
P0 — and would have been silently admitted at the next wave boundary, which is
exactly what the gate had refused. The coordinator caught it and deferred all
11 by hand with `grooming_status=out-of-scope-this-run`. The cost of missing it
is a run that grows past its authorized scope through a side channel, with the
gate reading as honoured.

*Evidence note: this rests on the coordinator's `incident:` comment on the
epic, which neither downstream harvest reaches (see §1). It is quoted there in
full and corroborated by the tracker state — `qozp.18` through `.28` all carry
`deferred` and `grooming_status=out-of-scope-this-run`.*

**Proposed change.** In section 3, file each finding **deferred**; in section
4, undefer only what classification puts in scope. Filing stays free and losing
a finding stays impossible, but a finding no longer enters the run by default.
Done: findings created by `sprint-review` do not appear in
`bd ready --parent {EPIC}` until section 4 admits them.

---

#### S11 · Severity cannot override the scope gate, so a run ships its own regression

`audience: skill` · `target: sprint-review` · `priority: 1` ·
`findings: ["qozp#9 (context)"]` · `model: claude-opus-5 (coordinator)`

**What is wrong, and what it costs.** Section 4's in-scope test requires
**both** a defect in what the run changed **and** gate authorization for scope
growth. `admit_review_findings` resolved to `false` from the skill default, so
the second condition failed for every finding regardless of severity — and the
run therefore merged and shipped `872f390`, which review then confirmed carries
a silent data-loss path that wipes `admitted`, `excluded`, `branch`,
`base_sha`, `standing_auth`, `checks_ids`, `checks_baseline`, `review_passes`
and `review_budget` from a run manifest while printing a success line. That
is a P0 regression **the run itself introduced**, declined by a gate whose
purpose is to stop the run growing into work nobody asked for — a different
thing entirely. The gate is right to refuse new work; it should not be the
mechanism by which a run ships a defect it created.

**Proposed change.** Add a third row to section 4's classification table: a
defect **the run's own merged diff introduced**, at P0, is in scope regardless
of the gate — the run either fixes it or does not merge the commit that carries
it. Leave every other severity and every pre-existing defect on the
conservative default. Done: a P0 regression attributable to a bead in
`manifest.admitted` classifies in-scope with `admit_review_findings: false`.

---

#### S12 · A bare `incident:` comment is harvested by nothing

`audience: skill` · `target: bead-bonanza` · `priority: 1` ·
`findings: ["qozp#8 (kind drift, same class)"]` · `model: claude-opus-5 (retro)`

**What is wrong, and what it costs.** The protocol admits two forms for an
incident: `retro:<kind>:<target> … incident` on the bead, and (per
`run-record.md`) `amend: incident …` on the epic. `bead-bonanza` prescribes
`amend:` but never names the incident forms at the point a coordinator records
one, and this run's coordinator wrote a bare `incident: …` comment. Measured
against both harvests over this run's comment set: `sprint-retro`'s
`startswith("retro:")` misses it, and `sprint-report`'s
`^(retro:[a-z]+:[^ ]+ incident|amend: incident)` misses it. So does an
unprefixed coordinator-verification comment on `qozp.18`. The cost is total
loss: the run's P0 incident and its independent confirmation were invisible to
both downstream ceremonies, and were recovered here only because a human
handoff named them.

**Proposed change.** Where `bead-bonanza` tells the coordinator to record an
incident, name the two admissible prefixes verbatim and state that a bare
`incident:` is harvested by nothing; and require coordinator verification of a
review finding to be recorded as a `retro:proof:<bead>` comment rather than an
unprefixed one. Done: every comment a coordinator is instructed to write
carries a prefix one of the two harvests matches.

---

### Project audience — 4, destined for this tracker

---

#### P1 · No rule says when a cross-skill citation vendors and when it degrades to prose

`audience: project` · `target: scripts/check-skills.sh` · `priority: 1` ·
`findings: ["qozp#10", "qozp.2#3", "qozp.2#7", "qozp.5#4"]` ·
`model: claude-opus-5 (advanced tier) + claude-sonnet (standard tier)`

**What is wrong, and what it costs.** `check-skills.sh` rule 3 errors on any
`../<skill>/references/` path and rule 2 requires a backticked `*.md` path to
resolve from the citing file's own directory. Together they make it impossible
for one skill to cite another's reference, and the checker's error text names
exactly one remedy — vendor it through `scripts/sync-shared-references.sh`.
Nothing states when that remedy is right. Two beads in **this one run** hit the
rule and resolved it **opposite** ways, neither knowing about the other:
`qozp.2` cited the roster **by role in prose** and explicitly rejected
vendoring (its acceptance criterion forbade a second copy), while `qozp.5`
**vendored**, adding a MAP line and a generated
`sprint-plan/references/shared/epic-definition-of-ready.md`. Both are defensible
against their own acceptance criteria and the corpus now demonstrates two
incompatible answers to one question. Cost: the next bead picks whichever it
reads first, and the vended skills accumulate both patterns — one of which
leaves a reader with an unlinked hop and the other of which adds a file
`shared-refs` must keep current forever (34 → 35 this run).

**Proposed change.** State the rule and put it where the agent hits the wall:
in `check-skills.sh`'s rule-3 error text and in this repository's authoring
guidance. Concretely — vendor when the cited text is *normative for the citing
skill* and must travel with it; cite by role in prose when the cited file has a
single sanctioned home that a second copy would contradict. Done: the error
text names both branches and the test that chooses between them, so the two
resolutions this run produced are no longer both defensible.

---

#### P2 · The design doc's `BEADS_ACTOR` premise is stated unqualified and is false as written

`audience: project` · `target: docs/plans/beads-coordination-design.md` ·
`priority: 2` · `findings: ["qozp.2#2", "qozp.2#1"]` ·
`model: claude-opus-5 (advanced tier)`

**What is wrong, and what it costs.** Line 31 reads "A claim without it
silently rewrites assignee and decays the coordinator's ownership view", and
lines 1163–1164 repeat it; rollout item 4 and the bead's own description repeat
it again. Measured on bd 1.2.2 in a throwaway database, four arms: on an
**unassigned** bead, `env -u BEADS_ACTOR bd update <id> --claim` exits 0 and
silently sets `assignee` to the git-config fallback; on an **already-assigned**
bead the same command exits 1 with "issue already claimed by dev-x" and changes
nothing, and `bd ready --claim` on an already-assigned ready front returns `[]`
and claims nothing. So the silent-decay failure mode has exactly one entry
point. The cost is that items 6–14 — the whole of run 2 — are specified against
a premise that overstates the hazard, and an agent that discovers the
overstatement mid-run has to decide alone whether the rest of the design is
trustworthy. `dispatch-policy.md` has already been corrected; the design doc
and the rollout plan have not.

**Proposed change.** Amend those three sites to the measured statement — a
claim without `BEADS_ACTOR` reassigns an **unassigned** bead silently, and is
refused loudly on an assigned one — and keep the conclusion, which the
measurement strengthens rather than weakens: "export `BEADS_ACTOR` everywhere"
and "every bead carries an assignee before dispatch" are belt and braces, not
alternatives. Done: no site in `docs/plans/` states the claim unqualified.

---

#### P3 · Rollout items specify a manifest key without naming its writer

`audience: project` · `target: docs/plans/beads-coordination-rollout.md` ·
`priority: 2` · `findings: ["qozp.4#2"]` · `model: claude-opus-5 (advanced tier)`

**What is wrong, and what it costs.** Item 2's Files clause lists exactly
`sprint-plan/references/run-record.md` (the key spec) and
`sprint-plan/SKILL.md` (the rule). Both were done — and
`manifest.concurrency_cap` still read `null` at run time, because the manifest
is populated by `bead-bonanza/scripts/pre-flight-check.sh`, which the item
never names. Item 3's within-cap Definition-of-Ready check therefore had
nothing to compare against. It cost this run a mid-run admission
(`agent-plugins-h3j8`) and a re-entry to grooming at the wave 2 boundary. Items
6–14 are unstarted and carry the same structure, so the same shape can recur up
to eight more times.

**Proposed change.** Before run 2 opens, audit items 6–14: for every item whose
Files clause specifies a key, a metadata field or an emitted value that a later
item **reads**, add the producing component to that item's Files clause. Done:
no remaining rollout item specifies a value with no named writer.

---

#### P4 · Three measured environment facts have no home in the ledger

`audience: project` · `target: AGENTS.md` · `priority: 3` ·
`findings: ["qozp.3#4", "qozp.2#6"]` · `model: claude-opus-5 (advanced tier)`

**What is wrong, and what it costs.** `environment` findings produce no issue
by the routing table, but landing them in `AGENTS.md`'s environmental-failure
ledger is an edit that needs an owner, and without one they die in this
document. Three are ready and drafted verbatim in §4 above: the `bd setup
claude` "stale" verdict being conditional on `sync.remote` rather than on
template drift (with the `bd -C` consequence), the Bash worktree-isolation
guard's non-reproducible refusals and the pattern that works around them, and
GitHub Actions being down at the account level. The second cost a bead five
extra round trips this run; the first produced a full bead's worth of
investigation that ended in a decline. The ledger's purpose is that the next
agent does not repeat either.

**Proposed change.** Append the three entries from §4 to `AGENTS.md`'s
environmental-failure ledger, each stating the conditions it was measured under
and, for the GitHub Actions entry, that its evidence is a coordinator
observation rather than a captured finding. Done: all three are in the ledger
in the form the ledger's own admission rules require.

---

## 6 · Every finding, verbatim

Grouped by kind, in harvest order within each group. Nothing is filtered:
the complete set is the substrate recurrence runs on, and a finding filtered
out once can never be observed appearing twice.

### `gap` — 4

#### `agent-plugins-qozp.4#2`

```
retro:gap:docs/plans/beads-coordination-rollout.md item 2 specifies a manifest key but names no component that writes it.
evidence: rollout item 2's Files list is exactly 'plugin/skills/sprint-plan/references/run-record.md (the manifest key spec ...), plugin/skills/sprint-plan/SKILL.md (the rule)'. Both are now done. But the manifest is populated by plugin/skills/bead-bonanza/scripts/pre-flight-check.sh, whose header comment reads '#   Manifest + emitted:  checks_ids, checks_baseline, the model lineup' (line 21) and which sets only checks_baseline_status / checks_baseline (lines 667-669). grep -rn 'concurrency_cap' plugin/ returns hits in sprint-plan only. So manifest.concurrency_cap reads null at run time and item 3's DoR check has nothing to compare against.
Filed as a bead rather than fixed here: pre-flight-check.sh is a bead-bonanza script outside this bead's two owned files and another wave-1 bead may hold it. The spec side is unaffected — run-record.md now says the orchestrator resolves it with the rest of the manifest, before grooming is dispatched, as \${CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS:-20}.
```

#### `agent-plugins-qozp.2#5`

```
retro:gap the roster names eight agent kinds but only the dev path has a dispatch site today, so seven of them are documented and unwired.
evidence: dispatch-policy.md's roster carries coordinator-, groom-, plan-, dev-, merge-, review-, retro-, report-. The only files this bead could wire are prompt-per-bead.md, prompt-batch.md and sprint-wave/SKILL.md — all dev-side. The rollout plan's item 4 Files clause defers the rest: 'and — once item 7 lands — wherever each ceremony agent is dispatched from.' agent-plugins-qozp.7 creates those sites and none of its six acceptance criteria mentions BEADS_ACTOR or assignment, so nothing would have wired them. Filed as agent-plugins-qozp.17 (discovered-from qozp.2, blocked by qozp.7).
```

#### `agent-plugins-qozp#6`

```
retro:gap:run-record — the manifest's 'excluded' shape is easy to get wrong and nothing validates it. run-record.md documents [{"id":"<bead-id>","reason":...}] — ONE id per entry. I wrote [{"ids":[...],"reason":...}] to group beads sharing a reason, which is the natural thing to write and reads fine as JSON. bb-status.sh rendered it as 'None (reason); None (reason); None (reason)' — it read .id, found nothing, and printed None rather than reporting a shape error. Same class as the array-vs-string trap above: the manifest has a schema, nothing enforces it, and the failure surfaces as plausible-looking wrong output. A validator over the manifest keys would catch both.
```

#### `agent-plugins-qozp#10`

```
retro:gap:check-skills-rule-3 — two beads in this run hit the same rule and resolved it OPPOSITE ways, and neither knew about the other. check-skills.sh rule 3 forbids cross-skill reference paths and rule 2 requires a backticked path to resolve locally, so no skill can cite another skill's reference. agent-plugins-qozp.2 resolved it by citing the roster BY ROLE in prose and explicitly REJECTED vendoring (its AC forbade a second copy); agent-plugins-qozp.5 resolved it by VENDORING through scripts/sync-shared-references.sh, adding a MAP line and a generated shared/epic-definition-of-ready.md. Both are defensible against their own ACs, but the corpus now demonstrates two incompatible answers to one question. This needs a stated rule about when a cross-skill citation vendors and when it degrades to prose, or the next bead will pick whichever it happens to read first.
```

### `decision` — 6

#### `agent-plugins-h3j8#2`

```
retro:decision:plugin/skills/bead-bonanza/scripts/pre-flight-check.sh the design offered two placements — hoist the section-8 manifest merge out of its 'checks were declared' branch, or add a second merge guarded on EPIC. Took the second: a new section 8b with its own python heredoc.
evidence: hoisting would have had to move the \`_raw\`/\`_st\` baseline variables (which only exist inside that branch) or split the block anyway; the second merge is additive and leaves the existing block byte-identical.
Cost, stated because a diff-reader would not see it as a choice: an epic-scoped pass that DOES declare checks now makes two \`bd update <epic> --metadata\` calls instead of one. Both read-then-merge, so they compose; the extra call is the price of the key being written on the no-checks path too.
```

#### `agent-plugins-qozp.4#1`

```
retro:decision:plugin/skills/sprint-plan/SKILL.md chose which expository prose to relocate — the bead delegated this pick to the subagent.
evidence: three blocks moved into references/run-record.md, each leaving the rule plus a citation in the body; three further passages trimmed of pure rationale. Selected on one test — the passage is topically the run-record's subject (wave table, C4 overlap, packing), so the reference stays coherent rather than becoming a spillover bin.
  1. 'Worktrees do not make batching redundant' full explanation -> '### Why a co-edit collision batches rather than parallelizes'; body keeps 'the collision is at the merge, not between the agents. Isolation is never grounds to relax the rules below — why, in references/run-record.md.'
  2. 'bd swarm validate ... two conditions on trusting it' bullets -> '## Trusting bd swarm validate's preview'; body keeps 'two measured conditions void its partition — an outside-dependency warning, and a nested run epic.'
  3. The 'Reason column' worked restatement -> deleted from the body; run-record.md's '### The Reason column names the collision class' already carried it verbatim and the body cites it twice.
Trimmed rationale, no rule lost: 'those are the quality bar for one commit, sized to be affordable every time' (Batching does not buy fewer check runs); 'a blocks edge is honoured by every re-query ... honoured once, by whoever reads the table' -> 'a blocks edge survives every re-query of the ready front, a table row does not'; 'It is derivable from the graph, so a stored copy only drifts' -> 'see Out, above', which states it four sections earlier.
REJECTED, and worth naming: the bd human respond / bd gate --await measured evidence in the readiness-gate section is ~570 bytes of pure justification and the largest single candidate, but its subject is not the run record. Moving it there would have made the reference a grab bag; the bead also said not to invent a new structure, and run-record.md is the only reference this skill owns besides the grooming prompt template. It stays in the body.
Body: 19854 -> 19457 bytes, measured with python3 -c "print(len(open('plugin/skills/sprint-plan/SKILL.md').read().split('---',2)[2].encode()))". Net -397 while ADDING a 640-byte section, so ~1040 bytes were reclaimed. Leaves 543 bytes of headroom under the 20000-byte cap for the next bead on this file.
```

#### `agent-plugins-qozp.2#3`

```
retro:decision the chosen citation form — sprint-wave citing bead-bonanza's references/dispatch-policy.md by path — is forbidden by this repository's own skill check, so sprint-wave cites it by role instead.
evidence: scripts/check-skills.sh rule 2 requires every backticked <path>.md to resolve from the citing file's directory, and rule 3 errors on any '../<skill>/references/' with 'cites a sibling skill's references directly; vendor it into references/shared/ via scripts/sync-shared-references.sh'. First attempt wrote `bead-bonanza`'s `references/dispatch-policy.md` in sprint-wave/SKILL.md and in the two prompt files; bb-checks.py skills-strict returned exit 1 with six errors: 'backticked path does not resolve from this file: references/dispatch-policy.md' — twice for sprint-wave/SKILL.md, once each for prompt-batch.md and prompt-per-bead.md, and once each again through the .claude/skills/ symlinks.
Tried and rejected: vendoring the roster into references/shared/ (the checker's own sanctioned mechanism). It would move the roster's home off the file the user chose up front and would put a second physical copy under sprint-wave, which is what the single-home acceptance criterion exists to prevent.
Chosen: the roster stays in bead-bonanza/references/dispatch-policy.md. The two prompt files, being co-located, cite it as `dispatch-policy.md` and resolve. sprint-wave/SKILL.md cites it by ROLE in prose — 'the coordinator's dispatch-policy reference, which is their only copy' — matching the vocabulary already in that file ('The coordinator's prompt files'). The cost is one unlinked hop for a reader in sprint-wave. skills and skills-strict then returned exit 0.
```

#### `agent-plugins-qozp.2#4`

```
retro:decision the agent id and {DISPATCH_ID} are ONE string, so no new placeholder was introduced.
evidence: the bead's acceptance criterion reads 'sprint-wave/SKILL.md's placeholder table carries the agent id', which could have been met by adding an {AGENT_ID} row. sprint-wave/SKILL.md:75 already described {DISPATCH_ID} as 'one id per subagent, shared by every bead it holds' — the same granularity the roster gives dev-<run-id>-<batch-id>. This run's own dispatch confirms they are already the same value in practice: dispatch_id=dev-qozp-2 and BEADS_ACTOR=dev-qozp-2 on agent-plugins-qozp.2. Two placeholders holding one string is a drift source, so the existing {DISPATCH_ID} row was amended to say it IS the agent id, the assignee and $BEADS_ACTOR.
Side effect, deliberately not chased: plugin/skills/sprint-batch/SKILL.md:28 still reads '{DISPATCH_ID} | yours, shared by every bead in this batch'. Still true, no longer complete — it does not say the id is derived or that it is the actor. That file is outside this bead's declared scope, and the same wave held other beads, so it was left alone rather than edited.
Also decided silently: the N_dev <= cap - 1 formula was NOT written into dispatch-policy.md even though the rollout plan states it under item 4. agent-plugins-qozp.4's acceptance criteria place it in sprint-plan/SKILL.md; writing it here too would have created the second copy the single-home criterion forbids. dispatch-policy.md instead states only the roster fact the formula rests on — that the dev pool is the only concurrent kind, so a concurrency check counts concurrent assignees rather than distinct ones.
```

#### `agent-plugins-qozp.2#7` **· incident**

```
retro:decision:plugin/skills/sprint-wave/SKILL.md incident — the up-front decision said the roster is cited from sprint-wave, but scripts/check-skills.sh rule 3 forbids a skill citing a sibling skill's references by path ('cites a sibling skill's references directly; vendor it into references/shared/'), and rule 2 requires backticked paths to resolve locally. Decided: keep the single home in bead-bonanza/references/dispatch-policy.md and cite it from sprint-wave by ROLE in prose ('the coordinator's dispatch-policy reference, which is their only copy') rather than by path.
Alternative considered: vendor the roster into references/shared/ per the checker's sanctioned mechanism. Rejected — it puts a second physical copy under sprint-wave, which is exactly what the single-home acceptance criterion forbids, and moves the home off the file the user chose.
```

#### `agent-plugins-qozp.3#1` **· incident**

```
retro:decision:docs/plans/beads-coordination-rollout.md incident — declined bd setup claude's update; CLAUDE.md and .claude/settings.json are byte-identical to HEAD and the bead lands with no diff. The plan item's Verify line ('bd setup claude --check reports current') is unsatisfiable without installing 53 lines already present verbatim in AGENTS.md.
evidence: after `bd setup claude` (exit 0), `git status --porcelain` empty; `git rev-parse HEAD:CLAUDE.md` = 895b0e17f8f365e23ba3c78de725b6e41f328c1e = `git hash-object CLAUDE.md`; same for .claude/settings.json (1e367f8145d2a2007851995a0c9ebe97500ce177).
evidence: the only path to 'current' is a well-formed BEGIN..END pair in CLAUDE.md, which bd then fills with a 53-line body. Extracted that body from a sandbox install and diffed it against AGENTS.md lines '## Beads Issue Tracker' .. orphan END: both 53 lines, `diff -u` exit 0 — byte-identical. CLAUDE.md imports AGENTS.md via @AGENTS.md, so every session would load them twice.
Alternative considered: restore the markers and take the duplication, or delete CLAUDE.md's orphan BEGIN. Rejected — the first re-creates the duplicate-Beads-section problem commit 3293ed5 removed, and the second makes the duplication one accidental `bd setup claude` away (measured in a throwaway project: with no markers the tool takes its append path, '✓ Added beads section to existing CLAUDE.md'). Filed as agent-plugins-vqmc for an owner decision.
```

### `stale` — 2

#### `agent-plugins-qozp.2#2`

```
retro:stale:docs/plans/beads-coordination-design.md the 'a claim without BEADS_ACTOR silently rewrites assignee' premise is only true for an UNASSIGNED bead on bd 1.2.2; on an already-assigned bead the claim is REFUSED, loudly.
evidence: design doc line 31 — 'A claim without it silently rewrites assignee and decays the coordinator's ownership view'; lines 1163-1164 — '--claim rewrites assignee. Any worker that claims without BEADS_ACTOR set corrupts the coordinator's ownership view silently.' Rollout plan item 4 and this bead's own description repeat it. Measured (arm H above): bead assigned dev-x, `env -u BEADS_ACTOR bd update <id> --claim` -> exit 1, 'Error claiming probe-40s: issue already claimed by dev-x', assignee and status unchanged. `bd ready --claim` on a ready front where every bead already carried an assignee returned [] and claimed nothing.
So the silent-decay failure mode has exactly one entry point: a bead with no assignee. That does not weaken the rule — it locates it, and it is why 'export BEADS_ACTOR everywhere' and 'every bead carries an assignee before dispatch' are belt and braces rather than alternatives. dispatch-policy.md now states it that way, with the measured table, rather than repeating the unqualified 'silently rewrites' claim.
```

#### `agent-plugins-qozp.3#5`

```
retro:stale:CLAUDE.md the repo's own marker pair is split across two files, each half orphaned — CLAUDE.md:58 has BEGIN with no END, AGENTS.md:413 has END with no BEGIN. This, not bd, is why bd's update had nothing to delimit.
evidence: `grep -n 'BEADS INTEGRATION' CLAUDE.md` -> '58:<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:6cd5cc61 -->' (line 58 of 58, last line of the file). `grep -n 'BEADS INTEGRATION' AGENTS.md` -> '413:<!-- END BEADS INTEGRATION -->'. Nothing else matches in either file.
evidence (provenance): `git show 3293ed5 -- CLAUDE.md` hunk '@@ -45,216 +56,3 @@' deletes the block body and the END line from CLAUDE.md, keeping the BEGIN line as context — that commit moved the block into AGENTS.md and left the marker behind. `git show 9f418f4 -- AGENTS.md` deletes '-<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:970c3bf2 -->' in the same hunk that removed the model-lineup list ('Move the model lineup and run defaults into sf.toml'); the deletion is not mentioned in that commit message.
Filed as agent-plugins-vqmc — it needs an owner decision on which file owns the block, not a marker patch.
```

### `scope` — 1

#### `agent-plugins-qozp.5#4` **· incident**

```
retro:scope:agent-plugins-qozp.5 incident — the bead's Design scopes it to two files (groom-beads/references/epic-definition-of-ready.md and sprint-plan/SKILL.md), but the AC 'sprint-plan/SKILL.md's readiness gate runs them, citing the epic-DoR reference' cannot be satisfied inside that set. Expanded the set by two: one line in scripts/sync-shared-references.sh and the generated plugin/skills/sprint-plan/references/shared/epic-definition-of-ready.md it produces.
evidence: scripts/check-skills.sh rule 3, 'no skill reaches outside its own directory', errors on any `../<skill>/references/` in a SKILL.md — 'cites a sibling skill's references directly; vendor it into references/shared/ via scripts/sync-shared-references.sh'. Rule 2 independently errors on a backticked `*.md` path that does not resolve from the citing file's own directory. So a citation of groom-beads' copy is an error under either rule, and the repo's sanctioned way to cite it is the vendored copy.
Alternative considered and rejected: cite the filename WITHOUT backticks, which both rules' regexes miss (rule 2 matches only backticked paths; rule 3 matches only the literal ../<skill>/references/ form). It would have kept the file set at two, and it would have shipped sprint-plan with a dangling pointer to a file that is not in its tree — the precise failure the vendoring mechanism exists to prevent, evaded rather than met.
Alternative considered and rejected: restate the two checks in sprint-plan/SKILL.md. The AC forbids restating, and the byte cap could not have held them anyway (250 bytes of headroom remain after a 293-byte citation).
Cost of what was chosen: one more vendored file for shared-refs to keep current (34 -> 35), and epic-definition-of-ready.md now has two consumers, so an edit to it must be followed by scripts/sync-shared-references.sh.
```

### `tooling` — 8

#### `agent-plugins-qozp.2#6`

```
retro:tooling:agent-plugins the Bash worktree-isolation guard refused several plainly in-worktree commands, and the refusals were not reproducible on the same text.
evidence: 'export BEADS_ACTOR=dev-qozp-2 && bd update ... --set-metadata worktree="$(git rev-parse --show-toplevel)"' -> 'This agent is isolated in the worktree ..., but this command is too complex to verify that it stays inside the worktree.' Same for a bare 'bd update <id> --set-metadata dispatched_at="$(date -u +%FT%TZ)"' with no git call in it at all. Worked around by running 'date -u +%FT%TZ' alone and pasting the literal value.
The same shape then behaved inconsistently: a heredoc writing a script followed by 'chmod ...; echo written' was ACCEPTED; the same heredoc-plus-echo with different script contents was REFUSED twice, once with a 'git config user.name' line inside the heredoc and once after that line was removed. Fell back to the Write tool for the script and a separate Bash call to run it, which was accepted every time.
Net cost this bead: five extra round trips. The rule an agent can act on is 'write files with Write, run them with a single bare command' — the guard's own advice ('split it into plain, separate commands') was followed and still refused.
```

#### `agent-plugins-qozp#1`

```
retro:tooling:bead-bonanza — the checks-baseline inlining rule does not scale past one agent. prompt-per-bead.md requires the baseline block be 'inlined verbatim' into every subagent prompt, with an explicit warning against sending the one-word status instead. Measured on this run: the baseline text is ~6 KB, and wave 1 dispatches 4 agents, so obeying the rule literally costs ~24 KB of duplicated prompt. I deviated: each prompt carries the verdict ('all 13 declared checks passed, exit status 0', plus 'any failure you see is yours') and the exact command to read the full text from the manifest. The rule's stated purpose — that the agent running the checks actually sees the baseline rather than it stopping at the planner — is preserved, because the agent can fetch it. Candidate fix: state the verdict inline and the command for the detail, rather than requiring verbatim inlining.
```

#### `agent-plugins-qozp#2`

```
retro:tooling:bead-bonanza — the worktree-removal rule collides with harness-created worktrees, and the collision silently drops the guard the rule exists to keep. references/worktree-mechanics.md says to use raw 'git worktree remove', NOT 'bd worktree remove', specifically so the dirty-tree check is preserved (bd's version refuses a branch with no upstream, which every bead branch is, so its only path is --force which skips ALL checks). But a harness-created worktree is LOCKED — 'git worktree list' shows 'locked' on all four this wave — and plain 'git worktree remove' refuses a locked worktree. So the orchestrator is forced to --force anyway, losing exactly the dirty-tree guard the rule was protecting. Observed on agent-plugins-qozp.1. Candidate fix: 'git worktree unlock <path> && git worktree remove <path>' preserves the guard; the rule should say so, because --force is what an agent reaches for first.
```

#### `agent-plugins-qozp#4`

```
retro:tooling:bead-bonanza — the --set-metadata/--metadata array trap bit the RUN MANIFEST, and bb-status.sh reports the damage in a way that hides the cause. groom-beads/references/capability-tier.md already documents that --set-metadata stores a value as a STRING and arrays need --metadata (which REPLACES the whole object). The manifest's admitted/excluded/models keys are arrays and objects, and run-record.md shows them as JSON — but nothing warns that writing them needs --metadata. I wrote them with --set-metadata; they stored as strings; bb-status.sh then iterated the string CHARACTER BY CHARACTER and reported 'FAIL 116 admitted bead(s) not closed: [ " a g e n t - p l u g i n s ...' and 'OK 347 bead(s) EXCLUDED at admission'. 116 and 347 are character counts. Two fixes: run-record.md should state that array/object manifest keys require --metadata with a read-modify-write merge; and bb-status.sh should reject a string where it expects an array rather than iterating it, because the character-by-character output is what makes the real cause unguessable.
```

#### `agent-plugins-qozp#5`

```
retro:tooling:bead-bonanza — bb-status.sh's human-queue check is not scoped to the run's admitted set. It reported 'FAIL human queue: 1 item(s) awaiting a decision' for agent-plugins-qozp.12 (DECISION D1), which is DEFERRED and explicitly excluded at admission. Every other check in the same report scopes correctly ('bd lint: nothing flagged in this run's queue', 'every ready bead carries a tier'). A human item outside the admitted set should not fail the run's health, or the run can never read clean while any deferred decision exists anywhere in the epic.
```

#### `agent-plugins-qozp#7`

```
retro:tooling:bb-status-declined-bead — bb-status.sh bead exits 1 on a bead that legitimately produced no commit. agent-plugins-qozp.3's own acceptance permitted a decline branch ('OR the bead closes as declined with an upstream report filed as a bead and linked to this one'), which it took: bd setup claude wrote nothing, so there was no diff to commit. Signal 2 ('commit: none found') has no way to express that, so a correct outcome reads as repairable and the coordinator has to override the script by hand. Candidate fix: let a bead carry metadata.outcome=declined-no-diff (set here) and have signals 2 and 3 report OK-by-declaration against it rather than FAIL.
```

#### `agent-plugins-qozp#9`

```
retro:tooling:sprint-wave — sprint-wave could not dispatch a subagent: no Task/Agent tool was exposed to it, so it created the worktree as orchestrator and ran sprint-batch -> run-bead itself inside it. The skill's whole premise is 'dispatches a subagent per batch', and it degraded silently into doing the work rather than stopping. Two consequences: (a) the verification sprint-wave exists to perform against an independent agent's account was self-verification; (b) pre-flight-check.sh already models this environment as BB_ASSUME_NO_AGENT=1 and BLOCKS a run on it, yet the condition went undetected here because pre-flight ran in the coordinator's session, where the Agent tool IS available. So the capability is per-agent, not per-session, and a run-level pre-flight probe cannot answer it for a subagent. Candidate fix: sprint-wave probes for a dispatch tool as its first act and stops with a named blocker rather than absorbing the work.
```

#### `agent-plugins-qozp.3#2`

```
retro:tooling bd 1.2.2 `bd setup claude` reports '✓ Updated existing beads section in CLAUDE.md' and exits 0 while writing nothing, whenever the target file's BEGIN marker has no matching END. `--check` then reports stale forever and the command it names to fix it is the one that does nothing.
evidence: reproduced from scratch in a throwaway project (.beads/metadata.json + .beads/config.yaml with sync.remote set). CLAUDE.md = '# Proj / Hand-written intro. / <!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:6cd5cc61 -->' and nothing else.
  $ bd setup claude --check ; echo $?   -> '⚠ ... installed but stale' ; 1
  $ bd setup claude ; echo $?           -> '✓ Updated existing beads section in CLAUDE.md' ; 0
  $ md5sum before CLAUDE.md             -> 079602c9f186c6e9661b6c37509a35be twice
  $ diff -u before CLAUDE.md ; echo $?  -> (no output) ; 0
  $ bd setup claude --check ; echo $?   -> '⚠ ... installed but stale' ; 1
evidence (negative control): identical fixture except the block is well-formed (BEGIN / 'OLD MANAGED BODY' / END, hand-written prose above and below) — the update IS performed correctly: marker hash advances 6cd5cc61 -> 970c3bf2, the 53-line body is written, both hand-written paragraphs survive byte-for-byte, `--check` exits 0 '(current)'. So the update path is sound; only the unterminated-marker case is broken.
Filed upstream-report bead: agent-plugins-2nb8.
```

### `proof` — 9

#### `agent-plugins-h3j8#1`

```
retro:proof concurrency_cap reaches the manifest as a JSON number, demonstrated against a scratch beads database (own dir, own `bd init --prefix scratch`), never this tracker.
evidence: scratch repo /tmp/.../scratchpad/scratch-h3j8, epic scratch-tkn.
  $ pre-flight-check.sh scratch-tkn
    Concurrency
      concurrency_cap: 20 (a plan may size N_dev <= 19)
      written: concurrency_cap on scratch-tkn
  $ bd show scratch-tkn --json | jq '.[0].metadata.concurrency_cap'  -> 20
  $ bd show scratch-tkn --json | jq -r '.[0].metadata.concurrency_cap | type'  -> number
  stdout emit block carried: BB_CONCURRENCY_CAP=20
The scratch repo declares no sf.toml checks, so that pass took the 'Checks baseline / not captured' branch — which is the negative control the design asked for: the key is written on a pass that captures no baseline.
Non-numeric fallback: CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS=banana -> "...is not a number — using 20", manifest value 20, type number.
Numeric honoured: CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS=7 -> manifest value 7.
Merge, not replace: metadata before {"branch":"some/branch","base_sha":"deadbeef","concurrency_cap":20}; after the banana pass, identical — both pre-existing keys survived.
No bead was created in agent-plugins' tracker by any of this.
```

#### `agent-plugins-h3j8#3`

```
retro:proof full declared check set run in the batch worktree before close: bb-checks.py with no arguments, exit 0.
evidence: all 13 declared ids ran; skills/shared-refs/strata-docs/script-self-tests/skills-strict all RESULT: PASS, corpus 'no findings', ref-links 0 errors, template-drift 0 errors. Identical to the run manifest's checks_baseline (exit 0, everything passed) — no line present in mine and absent from the baseline.
pre-flight-check.sh --self-test alone: exit 0, 25/25 ok including 'no bash-4 case expansion', the portability lint, and '--json carries the same keys as the eval block' (the assertion the design named — BB_CONCURRENCY_CAP was added through the shared \`emit\` helper, so both output modes carry it).
```

#### `agent-plugins-qozp.5#1`

```
retro:proof both checks demonstrated against a scratch beads database (own dir /tmp/.../scratchpad/scratch-h3j8, own `bd init --prefix scratch`). No bead was created in agent-plugins' tracker: `bd list --status all --json | jq '[.[] | select(.created_by == "dev-qozp-5")] | length'` -> 0 (413 issues total).
Fixture epic scratch-oe6: scratch-oe6.1/.2/.3 assigned dev-demo-1/2/3, .4 groom-demo, .5 merge-demo, .6 unassigned. metadata.concurrency_cap=3.

CHECK 1 names the bead:
  $ bd list --parent scratch-oe6 --status all --json | jq -r '.[] | select((.assignee // "") == "") | .id'
  scratch-oe6.6

CHECK 2 names the count:
  $ cap=$(bd show scratch-oe6 --json | jq '.[0].metadata.concurrency_cap')   -> 3
  $ n=$(bd list --parent scratch-oe6 --status all --json | jq '[.[].assignee | select(. != null) | select(startswith("dev-"))] | unique | length')   -> 3
  dev pool 3, cap 3, budget 2
  FAIL: dev pool 3 exceeds budget 2 for cap 3

NEGATIVE CONTROL for dev-pool-vs-distinct, at cap=5 on the same fixture:
  cap=5 budget=4  dev-pool=3 -> PASS   all-distinct=5 -> FAIL
  Counting all distinct assignees would have rejected a run that fits, by exactly the 2 ceremony roles present (groom-demo, merge-demo). That is the under-subscription the check's wording warns about, measured rather than asserted.

--status all is load-bearing, measured on the same fixture with scratch-oe6.6 closed:
  default (open only): (empty)
  --status all:        scratch-oe6.6
An unassigned bead that is closed or deferred is invisible to a check written without it.
```

#### `agent-plugins-qozp.5#2`

```
retro:proof:plugin/skills/groom-beads/references/epic-definition-of-ready.md the `Owner:` trap re-measured in the scratch database rather than taken on trust, because it is the reason check 1 is written against --json:
  $ bd show scratch-oe6.6 | sed -n '2p'
  Owner: dev-qozp-5 · Type: task
  $ bd show scratch-oe6.6 --json | jq -c '.[0] | {assignee, created_by}'
  {"assignee":null,"created_by":"dev-qozp-5"}
bd 1.2.2. The Owner line prints created_by, so an 'every bead carries an assignee' check read off it passes on a bead with no assignee at all.
```

#### `agent-plugins-qozp.5#3`

```
retro:proof byte budget and gates.
evidence: plugin/skills/sprint-plan/SKILL.md body, measured with the command the bead's Design names —
  before: 19457   after: 19750   delta: +293   headroom to BODY_BYTE_MAX(20000): 250
qozp.4 left the room it promised, so the shortfall/Blocked path in the Design was not taken.
  $ scripts/check-skills.sh --strict  -> exit 0, '0 error(s), 0 warning(s)', RESULT: PASS
  $ scripts/ci.sh                     -> exit 0, '=== RESULT: PASS'
  $ bb-checks.py (no args, all 13 declared ids) -> exit 0; shared-refs now reports 35 vendored references (was 34) and ref-links 73 reference documents (was 72), both accounted for by the one new vendored copy. Everything else identical to the manifest's checks_baseline.
```

#### `agent-plugins-qozp.4#3`

```
retro:proof:plugin/skills/sprint-plan/SKILL.md byte-cap acceptance criterion verified, and the strict check confirmed clean rather than merely non-worse.
evidence: python3 -c "print(len(open('plugin/skills/sprint-plan/SKILL.md').read().split('---',2)[2].encode()))" -> 19457 (<= 19500, the bead's target). scripts/check-skills.sh's cap is BODY_BYTE_MAX = 20000 (line 43) applied to text.split('---', 2)[2] (line 128), so the one-liner measures exactly what the check measures.
plugin/skills/bead-bonanza/scripts/bb-checks.py with NO arguments: exit=0, all 13 declared checks, matching the run's baseline. The skills and skills-strict checks both print '--- 0 error(s), 0 warning(s) --- RESULT: PASS' — i.e. the file was not merely under the cap, it produced no warning at all, so no headroom was consumed silently.
```

#### `agent-plugins-qozp.2#1`

```
retro:proof negative control for BEADS_ACTOR, run in a THROWAWAY beads database (bd 1.2.2 (dev)); the live tracker was not touched.
evidence: scratch db at /tmp/.../scratchpad/probe4 — its own git repo, own `bd init --prefix probe`, no Dolt remote, every call made with `--sandbox` (disables Dolt auto-push). `bd where` printed .../probe4/.beads/embeddeddolt. Fallback identity in that dir: git user.name = harness-default, $USER unset.
Four arms of `bd update <id> --claim` (the form run-bead step 1 uses):
  E  assignee <null>, env -u BEADS_ACTOR  -> exit 0, '✓ Updated issue'; after: status=in_progress assignee=harness-default   ASSIGNEE CHANGED (null -> harness-default), silently
  F  assignee <null>, BEADS_ACTOR=dev-x   -> exit 0; after: status=in_progress assignee=dev-x
  G  assignee dev-x,  BEADS_ACTOR=dev-x   -> exit 0; after: status=in_progress assignee=dev-x            ASSIGNEE UNCHANGED
  H  assignee dev-x,  env -u BEADS_ACTOR  -> exit 1, 'Error claiming probe-40s: issue already claimed by dev-x'; after: status=open assignee=dev-x  (unchanged)
And `bd ready --claim`:
  I  assignee <null>, env -u BEADS_ACTOR  -> claimed probe-dlw, assignee=harness-default   ASSIGNEE CHANGED
  earlier run, every ready bead already assigned dev-x, with and without the var -> returned [] and claimed nothing
  `bd ready --claim --assignee <x>` -> {"error": "--claim cannot be combined with --assignee"}
Live-tracker non-mutation: `bd list --status all --json` captured before and after the experiment, reduced to '<id>\t<assignee>' and sorted; `diff` showed FOUR ADDED LINES ONLY (agent-plugins-2nb8, -h3j8, -l1p2, -vqmc — beads created by sibling agents in this wave, all unassigned) and ZERO changed lines. No existing bead's assignee moved. Assignee tally after: 110 Claude, dev-qozp-1, dev-qozp-2, dev-qozp-3, dev-qozp-4 — no 'harness-default' and no git user.name anywhere.
```

#### `agent-plugins-qozp#3`

```
retro:proof:agent-plugins-qozp.1 — verified against state rather than testimony before merging: bd show reported closed, terminal_state=completed, c25818e present on the worktree branch and absent from HEAD, diff limited to the two files the bead owns. Merged --ff-only under the merge slot held as coordinator-agent-plugins-qozp. bb-status post-merge exit 2 (worktree surviving), exit 0 after removal — read from a variable, not through a pipe, after first making the documented mistake of reading tail's status instead.
```

#### `agent-plugins-qozp.3#3`

```
retro:proof every file `bd setup claude` touched was enumerated by mtime, not by git status — git status would have hidden the two it rewrote with identical bytes, and would have missed anything written outside the repo.
evidence: touched a MARKER file, ran `bd setup claude`, then `find <worktree> -newer MARKER -not -path '*/.git/*'` and `find /root /home/user -maxdepth 6 -newer MARKER ...`. Inside the repo exactly two files: CLAUDE.md and .claude/settings.json — both byte-identical to HEAD afterwards (git hash-object == git rev-parse HEAD:<path>), so the tool rewrote or re-stat'd them without changing content. Outside the repo: /root/.beads/eventsData/*.evtq, /home/user/agent-plugins/.beads/interactions.jsonl, .beads/last-touched, .beads/embeddeddolt/beads/.dolt/{temptf,noms} — bd's own event/audit plumbing, touched by every bd invocation including `bd show`, all gitignored or outside the repo. No AGENTS.md write, no ~/.claude write, no .agents/ write.
evidence: setup output named both files itself — '✓ Hook already registered: SessionStart', '✓ Updated existing beads section in CLAUDE.md', 'Settings: <worktree>/.claude/settings.json'.
```

### `environment` — 1

#### `agent-plugins-qozp.3#4`

```
retro:environment the 'stale' verdict is NOT drift in bd's template — it is conditional on sync.remote, and this repo sets one. Without a remote the emitted block omits `bd dolt push` and the marker hashes 6cd5cc61; with a remote it includes `bd dolt push` and hashes 970c3bf2. Our marker carries 6cd5cc61, i.e. it was installed before the remote was configured.
evidence: throwaway project, no .beads/config.yaml, CLAUDE.md with no markers -> `bd setup claude` writes '<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:6cd5cc61 -->' and the block's Session Completion snippet contains 'git pull --rebase / git push / git status' with no `bd dolt push`. Same fixture plus .beads/config.yaml carrying sync.remote: 'git+https://github.com/brettcallaghan/agent-plugins.git' -> '<!-- BEGIN ... hash:970c3bf2 -->' and `grep -c 'bd dolt push' CLAUDE.md` = 1.
evidence: this also means `bd -C <dir> setup claude --check` gives a verdict about the CALLER'S cwd file computed against -C's config. Measured from this worktree: `bd -C /tmp/.../case-both setup claude --check` printed '✓ Claude Code integration installed: /home/user/agent-plugins/.claude/worktrees/agent-adcd876dffcae791f/CLAUDE.md (current)' — the worktree file, called stale by the same binary one command earlier, reported current because -C's project has no sync.remote. Reported upstream on agent-plugins-2nb8 as defect 2.
```

### `<unrouted>` — 2

#### `agent-plugins-qozp.1#1`

```
retro: clean — no findings beyond scope. Two prose edits (SKILL.md's wave-boundary loop line, and a matching new paragraph in supervision.md's 'Where is this run' section) both now state the two-command loop (bd epic close-eligible, then bd ready --parent <epic> --exclude-type=epic) and why omitting close-eligible is a silent deadlock. SKILL.md body measured 18,835 bytes (cap 20,000). All 13 declared checks pass, matching the baseline exactly.
```

#### `agent-plugins-qozp#8`

```
retro:process:wave-1-bypassed-sprint-wave — wave 1's four dispatches were made by the coordinator directly (harness Agent calls with isolation:worktree, branches worktree-agent-<harness-id>) rather than by invoking the sprint-wave skill as SKILL.md prescribes. Two consequences the retro should weigh: (a) the worktree names and branch names came from the harness, not from the dispatch id, so 'presence is the in-flight signal' worked but 'named for its dispatch id' did not, and the coordinator had to map harness ids to bead ids by hand; (b) sprint-wave's verification step was performed by the coordinator instead, which is exactly the reasoning the coordinator is supposed to not hold. Wave 2 invokes sprint-wave as designed, so the run carries one instance of each path and the retro can compare them.
```

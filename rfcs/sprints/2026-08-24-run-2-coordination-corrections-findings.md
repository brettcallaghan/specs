# Sprint findings — Run 2: make the coordination plan trustworthy, and fix what run 1 shipped

Run epic `agent-plugins-so39` · branch `claude/bead-bonanza-design-review-ne1plu`
· base `5cc1ce9` · head `99ee4a4`.

Eleven beads admitted, ten closed with commits, one bounced. One wave, six
concurrent batches, all merged in return order: `4tjk` `8a68d6a`; batch 6
(`qozp.23`, `qozp.28`, `l1p2`) `6d92098..582d458`; batch 1 (`qozp.19`, `d6zz`)
`f6d675e..f970cb5`; `qozp.20` `726ab33`; `uva5` `9c28e4b`; batch 2 (`gu3i`,
`bbtl`) `..99ee4a4`. `agent-plugins-xpn8` was bounced at the `sprint-plan`
readiness gate, deferred, and labelled `human` with three named resolutions.
The review loop closed at pass 1 of a 2-loop budget with 10 findings filed and
0 admitted.

**Model lineup — not from the manifest, because the manifest does not carry
it.** `metadata.models` and `metadata.effort` are both **absent** from
`agent-plugins-so39`, though `sprint-plan/references/run-record.md:58-59`
specifies both. Recovered instead from the coordinator's `amend:` dispatch
comment: `dev-so39-1` sonnet, `dev-so39-2` opus, `dev-so39-3` sonnet,
`dev-so39-4` opus, `dev-so39-5` opus, `dev-so39-6` haiku. Coordinator and
ceremony agents: opus (`claude-opus-5`). Effort: unrecorded — this harness
exposes no effort setting. Every issue below carries that, and the absence is
itself candidate **C3**.

**What this run was for, and what it did to itself.** The run existed to stop
*documents asserting unmeasured tool behaviour* from propagating errors. It
then committed the same error at three separate tiers, and each one was caught
by measurement rather than by review:

1. **In shipped text.** `--set-metadata` "stores every value as a string" is
   false on bd 1.2.2 — JSON scalars keep their type. The coordinator carried
   that false universal into six dispatch prompts, one bead description and two
   epic comments, and it landed **inside `agent-plugins-4tjk`'s corrected
   paragraph** — the very text fixing the previous unmeasured claim.
2. **In a dispatch prompt.** The coordinator attributed `sync_dir` / `rm -rf` to
   `sync-shared-references.sh`; it lives in `generate-schemas.sh`. `dev-so39-4`
   measured it instead of believing it, and the rule that shipped rests on the
   measured behaviour (`render()` ends `cat "$1"`) rather than on the
   misattribution.
3. **In this retro's own dispatch brief.** "`bd search` does not work" was
   handed to this agent as a hard constraint. Measured here: `bd search`
   **excludes closed issues by default** and `--status all` returns the full
   set. One negative observation was generalised to a universal, and the
   universal was then used to invalidate the recorded fix for the run's
   duplicate-filing incident. See `so39#11`.

The pattern is not carelessness at any one site. It is that a *plausible
mechanical claim* travels down the dispatch chain as a premise, and only an
agent whose own work forces it to touch the mechanism ever tests it. Three of
the four candidates ranked highest below attack that path directly.

---

## 1 · Harvest

One query over the run's closure set — the epic and its 22 descendants:

```sh
bd export | jq -r --slurpfile s scope.json 'select(.id as $i | $s[0]|index($i))
  | .comments[]? | select(.text | startswith("retro:")) | .text'
```

**27 comments carrying 28 findings**, across 8 beads. Two of the 28 were
recorded by this retro (`so39#11`, `so39#12`) — see "Two findings this retro
added" below.

| bead | findings | tier / agent |
|---|---|---|
| `agent-plugins-so39` (epic — coordinator, and this retro) | 12 | advanced (`opus`) |
| `agent-plugins-uva5` | 3 | advanced (`dev-so39-4`, opus) |
| `agent-plugins-qozp.20` | 3 (in 3 comments, one carrying two) | advanced (`dev-so39-5`, opus) |
| `agent-plugins-bbtl` | 3 | advanced (`dev-so39-2`, opus) |
| `agent-plugins-4tjk` | 2 | standard (`dev-so39-3`, sonnet) |
| `agent-plugins-gu3i` | 2 | advanced (`dev-so39-2`, opus) |
| `agent-plugins-d6zz` | 1 | standard (`dev-so39-1`, sonnet) |
| `agent-plugins-qozp.19` | 1 | standard (`dev-so39-1`, sonnet) |
| `agent-plugins-qozp.23` | **0** | light (`dev-so39-6`, haiku) |
| `agent-plugins-qozp.28` | **0** | light (`dev-so39-6`, haiku) |
| `agent-plugins-l1p2` | **0** | light (`dev-so39-6`, haiku) |

**The zero row is the harvest's most informative line.** Batch 6 — the light
tier, three beads, one haiku agent — recorded nothing at all, and `bb-status.sh
bead` caught it as signal 5 ("findings: none recorded", exit 1 on all three).
The dispatch prompt asked for findings explicitly. Every other batch this wave
recorded some. This is `so39#4` and candidate **C8**.

### Records this harvest does not reach

Run 1 measured that `startswith("retro:")` misses a bare `incident:` comment
and an unprefixed coordinator-verification comment, and filed it
(`.sf/outbox/bead-bonanza/2026-08-24-bare-incident-comment-is-harvested-by-nothing.md`).
**Run 2 reproduced all three shapes.** Measured over this run's comment set:

| comment | shape | reached by the harvest? |
|---|---|---|
| `agent-plugins-so39` #14 | `incident: review pass 1 filed 10 findings, admitted ZERO…` | **no** |
| `agent-plugins-so39.5` #1 | `coordinator verification — CONFIRMED by independent measurement…` | **no** |
| `agent-plugins-so39` #1, #4 | `retro:process:…` | reached, but **unroutable** — `process` is not one of the nine kinds |
| `agent-plugins-bbtl` #1 | `AUDIT RESULT — rollout items 6-14…` | **no** (a work record, correctly) |
| `agent-plugins-so39` #12 (`amend: incident on agent-plugins-bbtl`) | `amend: incident …` | no — but `sprint-report`'s incident harvest does reach it |

The `so39` #14 comment is again the run's single most consequential record: it
carries the review loop's whole disposition and the `--set-metadata` incident.
It reached this document only because the dispatch brief named it. That is a
recurrence of a filed issue, and is candidate **C4**.

### Two findings this retro added

Both are recorded on the epic as ordinary `retro:` comments so that every claim
in the candidates below traces to a finding id, per the issue template's
evidence rule.

- **`so39#11`** — `retro:tooling:bd incident`, carrying `corrects: so39#9`. The
  `bd search` correction. Measured in this session.
- **`so39#12`** — `retro:tooling:bead-bonanza incident`. The coordinator stall,
  **relayed** by the coordinator in this retro's dispatch brief and not
  observed by `retro-so39`. Marked as relayed in the finding itself.

**Ordering note.** `so39#10` (`retro:proof:progress-streaming`) was recorded by
the coordinator between this retro's first and second harvests, and already
corrects the stall's framing: a durable recovery trigger
(`trig_01Tr5qJXx7ZnRFVvZohSZpwc`) armed at first dispatch fired at 17:57:32
carrying the right instruction, so the watchdog design works and the gap is the
**interval** (50 minutes), not a missing mechanism. `so39#12` therefore records
a stall whose "no signal could have caught it" reading is already withdrawn;
candidate **C11** is scoped to the interval and to the coordinator's own
liveness, not to a new signal.

## 2 · Routing

A lookup on `kind`, with the merge test applied only where kind and audience
diverge — which in this repository they do constantly, because **we vend the
skills we run**, so a `tooling` finding may land in `scripts/` (ours) or in
`plugin/skills/` (the skill's).

| kind | count | destination |
|---|---|---|
| `proof` | 10 | this document, beside what it supports. **No issue.** |
| `tooling` | 5 | 4 skill, 1 project (`so39#7` — `scripts/sync-shared-references.sh` is repo-owned) |
| `decision` | 4 | 2 project (`docs/plans/`), 2 skill (`sprint-plan/SKILL.md`, grooming quality) |
| `gap` | 2 | skill — both targets (`run-bead`, `dispatch-policy.md`) are skill files |
| `stale` | 2 | skill — a dispatch prompt and an AC grep pattern |
| `scope` | 2 | project (`docs/plans/`, `sprint-plan/references/run-record.md`) |
| `sequencing` | 1 | project |
| `process` | 2 | **UNROUTED** |
| `environment` | 0 | — |

**Audience after the merge test: skill 10, project 6, terminated 10, unrouted 2.**

### Unrouted: 2 of 28, and both are the coordinator's

`so39#1` (`retro:process:duplicate-bead`) and `so39#2`
(`retro:process:coordinator-dispatched-directly`) use `process`, which is not
one of the nine kinds. **This is the same drift run 1 filed**, from the same
role: run 1's `qozp#8` was `retro:process:wave-1-bypassed-sprint-wave`. Two
runs, two coordinators, the same invented kind. The unrouted count is the
taxonomy's health metric and it is not falling; what it is measuring, on this
evidence, is that the coordinator role has no `retro:` examples in front of it
at the moment it records one. Both findings' content is squarely
`bead-bonanza`'s: neither is lost, but neither routed.

### Terminated here: `proof`, 10 of 28

Nothing for a human to decide, and the highest-quality writing in the run.
Worth naming three, because each is a negative control that changed a verdict:

- **`uva5#2`** — the `shared-refs` flake. 30 runs on an unchanged tree → 2
  failures naming a *different* correctly-mapped file each time; a scratchpad
  copy with exactly one line changed to a herestring → 100 runs, 0 failures.
  That is what turns "my change broke the gate" into `agent-plugins-b9ft`.
- **`so39#5`** — three data points across one wave establishing that
  `dispatch_state` / `terminal_state` are written on the `sprint-batch` path and
  not the bare `run-bead` path, converting `so39#3`'s hypothesis into a located
  defect. Supports **C9**.
- **`so39#8`** — the coordinator's own false claim in a dispatch prompt, caught
  by the bead it was sent to. Supports **C7**.

Two `proof` findings are load-bearing for candidates and are cited there:
`so39#6` (byte budget) under **C10**, and `so39#10` (recovery trigger) under
**C11**.

### Deduped against issues already filed — 3

| finding | already filed as |
|---|---|
| `so39#7`, `uva5#2` | `agent-plugins-b9ft` — the `printf \| grep -qxF` flake in `sync-shared-references.sh` |
| `qozp.20#3b` | `agent-plugins-b1gl` — `run-record.md`'s wave-table Batch column still numbers per wave |

`agent-plugins-b9ft` is the run's first follow-up by the coordinator's own
disposition. Neither is re-proposed here; **C14** is a *different* change built
on the same evidence (baseline semantics, not the pipeline bug).

## 3 · What the findings say, by theme

### Theme 1 — a mechanical claim travels as a premise and nobody measures it

Five findings, three tiers, one shape. `so39#8` (coordinator → dispatch prompt →
`uva5`), `uva5#3` (the same event from the bead's side), the review's `so39.5`
plus the coordinator's verification comment (`--set-metadata`), `so39#9`/`so39#11`
(`bd search`), and `4tjk#2` (an AC's own grep pattern undercounting its own
target set).

**What is common is not the claim; it is the position.** In every case the
claim was inherited from a document or a prompt written by something upstream,
and the only agent that measured it was one whose task forced it to touch the
mechanism. Where no task touched it — the `bd search` universal — it survived
into the *next* ceremony's brief. Candidates **C7** (dispatch prompts),
**C2** (the dedupe mechanism), **C1** (the ledger entry) and **C6** (the false
claim sitting in run 1's outbox) all attack this path.

### Theme 2 — an absent signal reads as a clean one

`so39#4` (batch 6's zero findings), `so39#3` (`dispatch_state` null, so
`sprint-cleanup.sh`'s live-worker test and supervision's three-deaths diagnosis
have no input at all), and `so39#7` (a flaky check against a single-run baseline
that records it green, so the flake charges itself to whichever bead hits it).

Run 1 recorded the same shape as "near-zero `discovered-from` is suspicious".
It is worse at the light tier, because that is where the cheapest agents run the
most beads and where a step that is last in the ceremony and produces no output
anyone waits for is cheapest to drop. Candidates **C8**, **C9**, **C14**.

### Theme 3 — the byte cap is now the binding constraint on skill authoring

`so39#6` measures it on the merged head: `sprint-plan/SKILL.md` 19,946 (54 free),
`groom-beads/SKILL.md` 19,890 (110 free), `bead-bonanza/SKILL.md` 19,240 (760
free). Three of this wave's beads paid for their additions by cutting text.
`qozp.20#2` is the model case done right — a 236-byte overrun funded by deleting
two *restatements*, each named, each shown to survive elsewhere, with the
Blocked path explicitly considered and rejected on the grounds that neither cut
removes a rule.

**That is the right call once and it is not a strategy.** At 54 bytes free the
next bead touching `sprint-plan/SKILL.md` cannot add a sentence, and the
pressure falls on whoever is least able to tell a restatement from a rule.
Candidate **C10**.

### Theme 4 — grooming specified two beads against cases it had not tested

`4tjk#2`: the bead's own AC grep (`metadata.{0,40}(replace|entire metadata)`)
found 3 sites; a fourth carried the same false claim in different words at
`capability-tier.md:11` and fell outside the 40-character window. Found only
because the agent widened the pattern as a sanity check before closing.

`uva5#1`: the bead's DESIGN proposed two branch tests, and applied to the very
case the bead was written to resolve (`qozp.2`'s roster citation) **both fire** —
reproducing the ambiguity the bead exists to remove. The agent shipped an
ordered pair of tests instead, and verified the new rule endorses both of run 1's
outcomes.

Both are the same defect at the grooming tier: an acceptance test was written
and never run against the case that motivated the bead. Candidate **C12**.

### Theme 5 — a bounce, and a deferred bead that is now actively unsafe

`agent-plugins-xpn8` is the run's one bounce and it is a good one: `plan-so39`
measured that "machinery that already exists" does not exist for this shape
(`schemas/` is one strata package scoped to the Spec CLI, so the bead is L not
S), that the "generated from" AC is unachievable as written, and that the bead
collides with `uva5`, which was unstarted at the time. Three named resolutions,
`human` label, `blocks` edge from `4tjk` intact, and `deferred` used
deliberately as the mechanism keeping it out of the ready front on `4tjk`'s
close — because it *is* in `manifest.admitted`, so a ready-front-minus-admitted
diff would not have caught it either.

`qozp.20#3` records the opposite case. `agent-plugins-qozp.22` reports the same
defect `qozp.20` fixed but proposes the **opposite** derivation
(`dev-<run-id>-<wave>-<batch>` versus the run-wide numbering that shipped).
It is now stale rather than merely deferred: reviving it as written re-breaks
what this run shipped. Left untouched, correctly, under the do-not-undefer
constraint. Candidate **C13**.

## 4 · Candidates

Fourteen. Ranked. Audience per the merge test; nothing here is filed.

| # | audience | target | priority | one line |
|---|---|---|---|---|
| C1 | project | `AGENTS.md` ledger | 1 | `bd search` excludes closed issues by default — record it, and retract "non-functional" |
| C2 | skill | `bead-bonanza` | 1 | A pre-create dedupe step, with the flag that makes it work |
| C6 | skill | `sprint-plan` (existing outbox file) | 1 | Run 1's outbox issue asserts the `--set-metadata` universal this run measured false |
| C7 | skill | `bead-bonanza` | 1 | A dispatch prompt is an unaudited authoring surface |
| C4 | skill | `bead-bonanza` (extend) | 1 | Kind drift and the bare `incident:` recurred — recurrence 0 → 1 |
| C5 | skill | `sprint-review` (extend) | 1 | Severity-vs-scope-gate recurred, on the run's own regressions — recurrence 0 → 1 |
| C9 | skill | `run-bead` | 2 | `dispatch_state`/`terminal_state` are written on one path of two |
| C8 | skill | `sprint-batch` | 2 | A batch's return line does not carry a findings count, so silence is invisible |
| C3 | skill | `bead-bonanza` | 2 | The manifest omits `models` and `effort`, so every issue's provenance is unfillable |
| C10 | project | `plugin/skills/**/SKILL.md` | 2 | Three SKILL.md files have under 800 bytes of headroom; two under 112 |
| C14 | skill | `bead-bonanza` | 2 | A one-run checks baseline cannot tell a flaky check from a regression |
| C12 | skill | `groom-beads` | 2 | Validate an acceptance test against the case that motivated the bead |
| C11 | skill | `bead-bonanza` | 3 | The recovery trigger's interval is the whole liveness guarantee at a ceremony boundary |
| C13 | project | `agent-plugins-qozp.22` | 3 | A deferred bead is now stale in a way that would re-break shipped work |

Full text for each candidate is in section 5. Findings not promoted stay in
section 6, where recurrence can still find them.

## 5 · Candidates in full

### C1 — `bd search` excludes closed issues by default (project · P1)

**What is wrong, and what it costs.** The run recorded, and this retro's own
dispatch brief restated as a hard constraint, that `bd search` is non-functional
on this build. It is not. It excludes closed issues by default — the same trap
the findings-capture protocol already documents for `bd list`, and for the same
reason the protocol mandates `bd export` over `bd list` when harvesting a
drained run. The cost already landed twice: the recorded fix for the run's
duplicate-filing incident ("run `bd search` before `bd create`") was written off
as unworkable, and the next ceremony was dispatched with a false premise in its
brief. `AGENTS.md`'s environmental-failure ledger is where the next agent will
look, and it says nothing about this.

**Evidence.** `so39#11` (measured), correcting `so39#9`; `so39#1`.

**Proposed change.** One ledger entry in `AGENTS.md`, in the ledger's own form —
the measurement, not the date. `bd search` excludes closed issues by default
(`bd search --help` line 1); `--status all` includes them. Measured on bd 1.2.2
(dev) against this tracker: `bd search metadata` → 2 hits, both deferred;
`bd search metadata --status all` → 6 hits, matching
`jq -r '.title' <export> | grep -ci metadata` → 6. A zero result from the
default is not evidence of a broken index.

### C2 — A pre-create dedupe step, with a mechanism that works (skill · `bead-bonanza` · P1)

**What is wrong, and what it costs.** The coordinator filed the same bead twice
27 minutes apart — `agent-plugins-vcka` (16:13:59Z) and `agent-plugins-xpn8`
(16:41:04Z), same two call sites, same pipeline, near-identical acceptance
criteria. Nothing caught it until `sprint-scope` did at admission, and scoping
could not resolve it either, because deduping is a structural change it may not
make unilaterally. No skill instructs an agent to check before `bd create`;
`groom-beads` treats duplication as an anti-pattern to *reject during grooming*,
which is a different step in a different role. Remembering what you filed is not
available: the first filing happened in a turn whose context the coordinator no
longer held, which is the normal condition for a long run. The cost is a
duplicate that consumed a bounce, a manifest `excluded` entry, and a
`bd supersede` — and, next time, one that is not caught.

**Evidence.** `so39#1`; `so39#11` (the mechanism); `so39#9` (the mechanism as
previously — wrongly — written off).

**Proposed change.** Where `bead-bonanza` instructs the coordinator to file a
bead, state the pre-create check and give it the flag that makes it work:
`bd search "<two or three distinctive title words>" --status all`, and inspect
the hits before creating. Naming `--status all` is the whole of the fix — the
bare form silently hides every closed bead, which on a long run is most of them.

### C3 — The manifest omits `models` and `effort` (skill · `bead-bonanza` · P2)

**What is wrong, and what it costs.** `sprint-plan/references/run-record.md`
lines 58-59 specify `"models": {"light": …, "standard": …, "advanced": …,
"frontier": …}` and `"effort": "low|medium|high|xhigh|unrecorded"`, and
`bead-bonanza/references/run-record.md` names both as the coordinator's to
write, with an explicit instruction to write `unrecorded` where the harness
exposes no effort setting. `agent-plugins-so39`'s manifest carries **neither
key**. `sprint-retro`'s closing step requires exactly these two values ("an
issue without them ships with an unfillable field"), so every skill-audience
issue this retro writes would carry a guess. Run 1's manifest carried both,
which makes this a regression rather than an unimplemented rule. The cost is
that findings cannot be weighed by capability tier, which is the one thing the
`model` field exists for.

**Evidence.** Measured on the epic: `jq '.metadata | keys'` returns 13 keys, and
`models` and `effort` are not among them; recovered instead from the
coordinator's `amend:` dispatch comment. No finding records the omission —
this candidate rests on that measurement, recorded here.

**Proposed change.** Add `models` and `effort` to the coordinator's manifest
write, and to whatever pre-dispatch check reads the manifest, as required keys —
failing loudly on absence rather than letting the run proceed. `unrecorded` is
already the specified value for a harness with no effort setting, so there is no
case where the key can legitimately be missing.

### C4 — Kind drift and the bare `incident:` recurred (skill · `bead-bonanza` · extend · P1)

**What is wrong, and what it costs.** This is not a new issue. It is run 2's
recurrence of `.sf/outbox/bead-bonanza/2026-08-24-bare-incident-comment-is-harvested-by-nothing.md`,
whose `recurrence` currently reads `0`. All three shapes that file names came
back: a bare `incident:` comment on the epic (`so39` #14) carrying the review
loop's whole disposition and the `--set-metadata` incident; an unprefixed
coordinator-verification comment (`so39.5` #1) carrying the independent
measurement; and `retro:process:` twice (`so39#1`, `so39#2`), which is the same
prefix drift as run 1's `qozp#8` and is what makes 2 of this run's 28 findings
unroutable. Both consequential records reached this document only because the
dispatch brief named them.

**Evidence.** `so39#1`, `so39#2` (kind drift); the harvest table in section 1
(bare `incident:`, unprefixed verification). Run 1: `qozp#8`.

**Proposed change.** Update the existing outbox file rather than filing beside
it: `recurrence: 0` → `1`, add run 2's findings to `findings:`, and add the
measured run-2 evidence — specifically that the drift is *coordinator-role*
drift in both runs, which argues for the fix landing where the coordinator is
told to record, not in the shared protocol reference it never opens.

### C5 — Severity cannot override the scope gate, again (skill · `sprint-review` · extend · P1)

**What is wrong, and what it costs.** Recurrence of
`.sf/outbox/sprint-review/2026-08-24-severity-cannot-override-the-scope-gate.md`,
`recurrence` currently `0`. Run 2 reproduced it with a sharper example than run
1's: review pass 1 filed 10 findings and admitted **zero**, and three of the ten
(`so39.4`, `so39.5`, `so39.6`) are defects **in the corrections this run
shipped** — on the exact thesis the run was opened to serve.
`standing_auth.admit_review_findings` was again `{value:false, source:"default"}`:
nobody chose it. `so39.5` is a false claim the run shipped in the same paragraph
in which it fixed the previous false claim, and the coordinator confirmed it by
independent measurement. It is deferred, unread on the merits, behind a *scope*
decision that never reached the severity question.

**Evidence.** `so39` #14 (the incident comment, quoted in section 6); `so39.5`'s
coordinator-verification comment; `so39#11` for the same class.

**Proposed change.** Update the existing outbox file: `recurrence: 0` → `1`, and
add run 2's evidence. The run-2 case strengthens the proposed classification row
in a specific way worth writing down — the defect was introduced by the run's
own merged diff *and* falsifies the run's own stated purpose, which is a
narrower and more defensible admission trigger than raw severity.

### C6 — Run 1's outbox issue asserts the claim this run measured false (skill · `sprint-plan` · correct · P1)

**What is wrong, and what it costs.** `.sf/outbox/sprint-plan/2026-08-24-run-record-omits-how-to-write-the-manifest.md`
states: "`bd`'s `--set-metadata` stores a value as a **string**; arrays need
`--metadata`". Measured on bd 1.2.2, the first clause is false as a universal:
`--set-metadata` JSON-parses scalars, so `5` → number, `true` → boolean, `1.5` →
number, `null` → null; only arrays, objects and non-numeric text are stored as
strings. The issue's *proposal* is unaffected — arrays do need `--metadata` —
but its stated rationale is wrong, and this file is queued for `triage-issue` to
sweep into this tracker and thence into `run-record.md`. If it lands as
written, the run that existed to stop unmeasured claims propagating will have
propagated one into the manifest key spec, through its own retro's outbox.
That is the same failure as `so39#8`, one further hop downstream.

**Evidence.** `so39.5`'s coordinator-verification comment (seven `--set-metadata`
pairs measured, quoted in full in section 6); review finding `agent-plugins-so39.5`.

**Proposed change.** Narrow the sentence in that file to what was measured:
`--set-metadata` JSON-parses its value, so arrays and objects — not scalars —
are what it stringifies, and those are what need `--metadata` plus a
read-modify-write. Everything else in the file stands.

### C7 — A dispatch prompt is an unaudited authoring surface (skill · `bead-bonanza` · P1)

**What is wrong, and what it costs.** Dispatch prompts are the highest-fanout
text in a run — six agents this wave — and nothing checks them. Two false
mechanical claims went out in them: `sync_dir` / `rm -rf` attributed to
`sync-shared-references.sh` when it lives in `generate-schemas.sh`, and
`--set-metadata` "stores every value as a string". Both were plausible, both
were inherited from real documents about adjacent subjects, and both were
transposed onto the wrong target while writing a prompt. The first was caught
because `dev-so39-4`'s task forced it to read the script; the second was not
caught by any bead and **shipped inside `4tjk`'s corrected text**. The cost is
asymmetric: a false claim in a prompt either wastes an agent's measurement pass
or becomes repository text, and which one happens is luck.

**Evidence.** `so39#8` (the `sync_dir` misattribution, from the coordinator's
side); `uva5#3` (the same event from the bead's side, with the corrected
mechanic and why it is stronger); the `so39.5` chain (the `--set-metadata`
universal into six prompts, a bead description and two epic comments).

**Proposed change.** A rule where `bead-bonanza` describes prompt authoring: a
dispatch prompt states the bead's problem and its constraints, and **any claim
about how a script or tool behaves carries the command that measured it or is
omitted**. A prompt that needs an agent to know a mechanism should send it to
the file, not restate the file. The counter-evidence is part of the case rather
than against it: in both instances the subagent measured rather than believing,
so the rule is cheap — it removes text rather than adding a check.

### C8 — A batch's return line does not carry a findings count (skill · `sprint-batch` · P2)

**What is wrong, and what it costs.** The light-tier batch — `dev-so39-6`,
haiku, three beads — recorded **zero** `retro:` comments, though the dispatch
prompt asked for them explicitly. Two readings, not equally likely: three
one-line edits genuinely produced no friction, or a light-tier agent dropped a
step that is last in the ceremony and has no output anyone waits for. It
surfaced only at acceptance, via `bb-status.sh bead` signal 5, exit 1 on all
three — after the batch had returned and its agent was gone. The cost is that
the run's cheapest agents, running the most beads, are exactly where the
evidence channel silently closes; and `bb-status.sh` failing three beads for it
trains the operator to override the signal.

**Evidence.** `so39#4`; the zero row in section 1's harvest table.

**Proposed change.** Add a findings count to the batch's return line, per bead,
so a zero is a value the batch *asserts* rather than an absence the orchestrator
discovers. A batch that reports `findings: 0` for a bead has made a claim that
can be questioned at return, while the agent is still alive; a batch that
reports nothing has not. **Sequencing note:** rollout plan item 9 deletes
`prompt-batch.md`'s return-line contract, so this change must be made against
whatever replaces it, or made part of item 9.

### C9 — `dispatch_state` and `terminal_state` are written on one path of two (skill · `run-bead` · P2)

**What is wrong, and what it costs.** A bead dispatched to `run-bead` directly
returns with `metadata.dispatch_state` and `metadata.terminal_state` both null;
a bead dispatched through `sprint-batch` returns with `returned` and `completed`.
Measured across one wave, three data points, same session. Nothing failed this
run, because `bb-status.sh bead` reads git and status rather than the dispatch
record — but two consumers that *do* read it are silently degraded:
`sprint-cleanup.sh` decides whether a standing worktree is a live worker by
`dispatch_state=running`, and supervision's three-deaths diagnosis has no signal
at all when the key is absent. So the cost lands not on the run that skipped the
write, but on the next run that has to decide whether an abandoned worktree
holds a living agent.

**Evidence.** `so39#3` (the observation and the hypothesis); `so39#5` (the
hypothesis confirmed and located: written on the `sprint-batch` path, not the
bare `run-bead` path).

**Proposed change.** `run-bead` writes both keys, since it is the skill both
paths share. Whatever `sprint-batch` writes today becomes redundant rather than
load-bearing, so the two paths cannot diverge again.

### C10 — Three SKILL.md files are at the byte cap (project · P2)

**What is wrong, and what it costs.** Measured on the merged head:
`sprint-plan/SKILL.md` 19,946 bytes against a 20,000 cap (**54 free**),
`groom-beads/SKILL.md` 19,890 (**110 free**), `bead-bonanza/SKILL.md` 19,240
(760 free). Three of this wave's beads had to pay for their additions by cutting
text. The cap is now the binding constraint on every change to these files, and
the next bead touching `sprint-plan/SKILL.md` cannot add a sentence at all — it
must first find something to delete, under time pressure, in a file it did not
write. `qozp.20` did this correctly and expensively: it measured the overrun,
identified two *restatements* rather than rules, quoted both, showed where each
rule survives, and recorded why the Blocked path was rejected. That is a
paragraph of reasoning per 125 bytes recovered, and it is not repeatable at
scale by agents less able to tell a restatement from a rule.

**Evidence.** `so39#6` (the measurement); `qozp.20#2` (the cut, its method and
its cost).

**Proposed change.** Relieve the pressure structurally before the next run
touches these files — move content that is genuinely reference material out of
`sprint-plan/SKILL.md` and `groom-beads/SKILL.md` into their `references/`,
restoring meaningful headroom. This is a project bead, not a skill issue: the
cap lives in this repository's `scripts/check-skills.sh` and the files are ours.
Scope it as "restore ≥2,000 bytes of headroom on each of the two files without
deleting a rule", so that the criterion is measurable and the anti-goal is named.

### C11 — The recovery trigger's interval is the liveness guarantee (skill · `bead-bonanza` · P3)

**What is wrong, and what it costs.** The coordinator ended a turn announcing
the next ceremony without dispatching it, with nothing in flight, and the run
sat still until the human asked. The mechanism that would have recovered it
exists and works — the durable trigger armed at first dispatch fired at 17:57:32
with the right instruction — so this is not a missing signal. What it is: at a
ceremony boundary, with the queue drained and no work in flight, the run's
entire liveness property is the trigger's interval, which was 50 minutes for
this run. A human noticing sooner is what actually resumed it. The cost is
bounded and real: up to one interval of dead time per stall, and an operator who
learns that runs need watching.

**Evidence.** `so39#12` (the stall, relayed); `so39#10` (the correction — the
trigger fired, the design works, the gap is the interval).

**Proposed change.** State the invariant where `bead-bonanza` describes ending a
turn: a coordinator turn ends with work in flight, a question to the human, or
the run's terminal state — **naming the next ceremony is not dispatching it**.
And say plainly what the recovery trigger is for, so the interval is chosen as a
liveness budget rather than inherited: it bounds dead time at a boundary, it does
not prevent one.

### C12 — Validate an acceptance test against the case that motivated the bead (skill · `groom-beads` · P2)

**What is wrong, and what it costs.** Two beads this run shipped with an
acceptance test that had never been run against the case the bead was written
for. `4tjk`'s AC specified `metadata.{0,40}(replace|entire metadata)` as its
completeness test; it found 3 sites and missed a fourth carrying the same false
claim in different words, at `capability-tier.md:11`, outside the 40-character
window — found only because the agent widened the pattern as a sanity check
before closing. `uva5`'s DESIGN proposed two branch tests, and on the disputed
case both fire, reproducing the ambiguity the bead existed to remove. Neither
cost this run anything, because both agents were advanced-tier and both checked;
a light-tier agent satisfying the AC as written would have closed `4tjk` leaving
the false claim in the corpus, which is precisely the failure that bead exists
to prevent.

**Evidence.** `4tjk#2`; `uva5#1`.

**Proposed change.** At grooming, where an AC states a mechanical completeness
test (a grep pattern, a count, a branch rule), run it against the motivating
case before marking the bead Ready, and record the result. Where it is a search
pattern, state in the AC that the pattern is a floor rather than a ceiling, so
an agent that widens it and finds more is obeying the criterion rather than
exceeding it.

### C13 — `agent-plugins-qozp.22` is stale, not merely deferred (project · P3)

**What is wrong, and what it costs.** `qozp.22` reports the same defect
`qozp.20` fixed and proposes the opposite derivation:
`dev-<run-id>-<wave>-<batch>` against the run-wide numbering that shipped
(`<batch-id>` is the batch's position in the run, numbered once across every
wave). The two are mutually exclusive. It sits deferred with
`grooming_status=out-of-scope-this-run`, which reads as "not yet" rather than
"no longer correct", so a future run picking it off the backlog would re-break
what this run shipped. Nobody touched it, correctly, under the do-not-undefer
constraint — which is exactly why it needs a record somewhere that is not this
document.

**Evidence.** `qozp.20#3`; the coordinator's `amend:` comment on the epic
(section 6).

**Proposed change.** File a bead to reconcile `qozp.22` — closed as superseded
by `qozp.20`, or rewritten to whatever residual defect survives the run-wide
numbering — with an explicit statement that it must not be admitted as written.
Filing a bead is not undeferring `qozp.22`; the decision stays with whoever
grooms the next run.

### C14 — A one-run checks baseline cannot tell a flake from a regression (skill · `bead-bonanza` · P2)

**What is wrong, and what it costs.** `agent-plugins-so39`'s `checks_baseline`
records `shared-refs` as **passed**, and `shared-refs` fails roughly 2 times in
30 on an unchanged tree. Every bead in the run is told that any failure it sees
is its own. So the flake charges itself to whichever bead happens to hit it and
reads as that bead's regression — which is what happened to `dev-so39-4`, and it
cost a full diagnostic pass (30 runs plus a 100-run negative control) to
disprove. The coordinator's disposition was to warn: "any red `shared-refs`
seen in this run should be re-run before it is believed" — advice that exists
nowhere the next run will read it.

**Evidence.** `uva5#2` (2/30 failures, distinct file each time; 100/100 clean
negative control); `so39#7` (the baseline-attribution consequence).

**Proposed change.** Where `bead-bonanza` states that a baseline failure is the
bead's own, add the escape: a check that the baseline records green and that
fails once is re-run before it is believed, and a check that fails
intermittently is filed against the check rather than against the bead. This is
distinct from `agent-plugins-b9ft`, which fixes this particular pipeline bug;
the next flaky check will not be that one. It is also distinct from the
already-filed `.sf/outbox/bead-bonanza/2026-08-24-checks-baseline-inlining-does-not-scale.md`,
which is about the baseline's size in a prompt, not its semantics.

## 6 · Every finding, verbatim, grouped by kind

Twenty-eight findings in twenty-seven comments, reproduced complete and
unedited. Nothing is filtered: a finding filtered out once can never be
observed appearing twice, and recurrence is the only signal separating a real
defect from one bad afternoon.

Finding ids are `<bead-id>#<ordinal>`, the ordinal being position among that
bead's own `retro:` records. **One caveat on `qozp.20#3`:** that single comment
carries two findings — a `retro:sequencing:` and a `retro:scope:` — so the
second is cited here as `qozp.20#3b`. The protocol assumes one finding per
comment, and where that does not hold the ordinal cannot address the second
finding at all.

### kind: `tooling` — 5 finding(s) · **skill** audience, except `so39#7` (repo-owned script → project)

#### `agent-plugins-so39#7` — Claude

```
retro:tooling:scripts incident — agent-plugins-b9ft (P1, filed by dev-so39-4) is a FLAKY CHECK this run's own validation depends on, and it is the exact trap AGENTS.md documents. scripts/sync-shared-references.sh's orphan guard tests membership with 'printf ... | grep -qxF' under 'set -o pipefail': grep -q closes the pipe as soon as it matches, printf dies on SIGPIPE, and pipefail turns a successful match into a failure - so a correctly-mapped file is reported ORPHAN. Measured: 2 failures in 30 runs on an unchanged tree, naming a DIFFERENT correctly-mapped file each time; negative control with only that line changed to a herestring, 100 runs, 0 failures. Coordinator disposition: NOT admitted mid-run - it blocks no remaining bead, batch 2 was already in flight, and the run's admitted set is fixed. But it is the first follow-up, because the epic's checks_baseline records shared-refs as PASSED, so the flake charges itself to whichever bead happens to hit it and reads as that bead's regression. Any red shared-refs seen in this run should be re-run before it is believed.
```

#### `agent-plugins-so39#9` — Claude

```
retro:tooling:bd incident — 'bd search' is NON-FUNCTIONAL on this build, and the run has now been bitten by it twice in opposite directions. review-so39 reports it returned zero hits even for the term 'metadata' across 440 beads, so it fell back to listing every bead and matching titles by hand to dedupe. Earlier the same session the coordinator filed agent-plugins-vcka and agent-plugins-xpn8 as duplicates 27 minutes apart, and the fix recorded for that was 'run bd search before bd create' — advice that cannot work on this build. The pre-create dedupe step is still right; it needs a mechanism that exists. Candidate: 'bd list --json' filtered on title, or fixing search upstream alongside agent-plugins-we50 and agent-plugins-2nb8.
```

#### `agent-plugins-so39#11` — retro-so39

```
retro:tooling:bd incident — 'bd search' is NOT non-functional; it EXCLUDES CLOSED ISSUES BY DEFAULT, which is the same trap the findings-capture protocol already documents for 'bd list'. The recorded fix for the duplicate filing therefore works after all, with one flag.
corrects: so39#9
evidence: bd 1.2.2 (dev), this tracker, 450 exported beads.
  `bd search --help` line 1: 'Search issues across title and ID (excludes closed issues by default).' line 5: 'Use --status all to include closed issues.'
  `bd search metadata` -> 'Found 2 issues' (agent-plugins-so39.5, agent-plugins-so39.4 — BOTH deferred).
  `bd search metadata --status all` -> 'Found 6 issues' — adds agent-plugins-vcka, -4tjk, -wyun, -n5ka, ALL FOUR closed.
  `jq -r '.title' export.jsonl | grep -ci metadata` -> 6. So --status all returns exactly the true set; the default returns the open/deferred subset.
  `bd search probe --status all` -> 'Found 54 issues'. Not a broken index.
context: review-so39 measured 'zero hits even for metadata across 440 beads' and the coordinator generalised it to NON-FUNCTIONAL on the epic. Both are consistent with the default filter: at the time review ran, agent-plugins-so39.4 and .5 did not yet exist (review filed them), and every other 'metadata' bead was already closed — so the true result under the default filter was zero. A single negative observation was carried into a universal without the flag being checked. Third instance this run of the failure the run was opened to stop, and this one propagated into the retro's own dispatch brief, which states 'bd search does not work' as a hard constraint.
```

#### `agent-plugins-so39#12` — retro-so39

```
retro:tooling:bead-bonanza incident — the coordinator STALLED at a ceremony boundary: it ended a turn announcing the next ceremony without dispatching it, with nothing in flight, and the run sat still until the human asked.
evidence: relayed by the coordinator in this retro's own dispatch brief — 'The coordinator stalled: it ended a turn saying "moving to the retro" without dispatching anything, with nothing in flight. The human had to ask. That is a coordinator defect, not a subagent one — capture it.' Not observed by retro-so39; recorded here so the observation carries a finding id and can be cited, per the append-only rule.
context: the run's queue was drained, the review loop had closed at pass 1, and the terminal ceremonies (report, retro) were the only work left. A run whose progress depends on a human noticing that nothing is happening has no liveness property at its ceremony boundaries — the same class as a silently-absorbed dispatch, one level up: nothing is wrong with any single step, and the run simply stops between them.
```

#### `agent-plugins-qozp.19#1` — dev-so39-1

```
retro:tooling The bare bd epic close-eligible command was documented at two sites (SKILL.md and references/supervision.md) with identical wording, so both needed the same fix to stay in agreement. No incident: acceptance criteria matched the design exactly, dry-run + intersect procedure as specified.
```

### kind: `gap` — 2 finding(s) · **skill** audience by the merge test — both targets are skill files

#### `agent-plugins-so39#3` — Claude

```
retro:gap:run-bead — a returning bead left metadata.dispatch_state and metadata.terminal_state BOTH null. Measured on agent-plugins-4tjk after dev-so39-3 returned: bd show reports status=closed, assignee=dev-so39-3, dispatch_state=null, terminal_state=null. In run 1 those keys carried 'returned' and 'completed' and the coordinator's acceptance check leaned on them. Nothing failed here because bb-status.sh bead reads git and status rather than the dispatch record, but two things that DO read it are degraded: sprint-cleanup.sh decides whether a standing worktree is a live worker by dispatch_state=running, and supervision's three-deaths diagnosis has no signal at all when the key is absent. The likely cause is that this dispatch invoked run-bead directly rather than through sprint-batch, so whichever of the two writes those keys was skipped - worth locating rather than assuming.
```

#### `agent-plugins-so39#4` — Claude

```
retro:gap:dispatch-policy — the light-tier trivial batch (dev-so39-6, haiku, three beads) recorded ZERO retro: comments, and bb-status.sh bead caught it: signal 5 'findings: none recorded', exit 1 on all three. The dispatch prompt asked for them explicitly. Two readings and they are not equally likely: three one-line edits genuinely producing no friction, or a light-tier agent dropping a step that is last in the ceremony and has no output anyone waits for. The other three batches this wave all recorded findings. This is the same shape as run 1's 'near-zero discovered-from is suspicious' observation - an absent signal read as a clean one - and it is worse at the light tier because that is where the cheapest agents run the most beads. Candidate fix: make the findings step produce something a return line must carry, so silence is visible at dispatch rather than at acceptance.
```

### kind: `decision` — 4 finding(s) · **project** for `docs/plans/`; **skill** for `sprint-plan/SKILL.md` and for grooming quality

#### `agent-plugins-uva5#1` — dev-so39-4

```
retro:decision:agent-plugins-uva5 sharpened the bead's proposed branch wording into two ORDERED tests, because the wording as proposed does not resolve the case it was written to resolve.
evidence: the bead's DESIGN proposed 'VENDOR when the cited text is normative for the citing skill and must travel with it' / 'CITE BY ROLE when the cited file has a single sanctioned home that a second copy would contradict'. Applied to agent-plugins-qozp.2's roster citation, BOTH fire: the id-derivation rule IS normative for sprint-wave and must travel with it (branch 1 says vendor), and the roster does have a single sanctioned home (branch 2 says prose). Two branches that both fire on the disputed case reproduce the ambiguity the bead exists to remove.
what I shipped instead, in AGENTS.md and in the rule-3 error text, is the same two branches behind an ordered pair of tests: (1) do you need the WHOLE file — the generator's render() is 'cat "$1"', whole files only, no section extraction, so 'one fact out of a larger document' has no vendoring form at all; (2) is the file APPLIED (procedure/checklist/standard/template, identical wherever read) or EXTENDED (registry/roster/id table, whose content asserts it is the one place an entry may be added). Test 1 wins where they disagree.
applied: qozp.2's roster fails test 1 (one fact from a coordinator-scoped policy document) and fails test 2 (a roster is extended) -> cite by role, CORRECT as shipped. qozp.5's epic-DoR passes test 1 (the file IS the two checks sprint-plan runs) and passes test 2 (a checklist, applied) -> vendor, CORRECT as shipped. The rule endorses both run-1 outcomes; nothing to migrate, and the bead's sequencing note about not deleting the vendored twin is moot.
```

#### `agent-plugins-qozp.20#2` — dev-so39-5

```
retro:decision:plugin/skills/sprint-plan/SKILL.md the assign step did not fit in 250 bytes; I paid for it by deleting two RESTATEMENTS, not two rules.

Measured with the checker's own method (text.split('---', 2)[2], len of utf-8):
  before  19750  (250 headroom)
  first draft of the new section  20236  (-236)
  after tightening the section    20033  (-33)
  after the two cuts below        19892
  after adding the ceremony-beads clause  19946  (54 headroom)

Cut 1 — SKILL.md:236-237, inside the C4 'shared-surface' bullet:
  '**The hazard is a *merge* collision**, not one between the agents, so
   worktree isolation does not relax this rule.'
Deleted because 30 lines earlier, under '## Bias toward batching', the same file
already says: '**Worktrees do not make batching redundant, because the collision
is at the merge, not between the agents.** Isolation is never grounds to relax
the rules below — why, in `references/run-record.md`.' The bullet is one of
'the rules below', so the rule survives with its rationale and its reference.

Cut 2 — '## The dev pool is capped':
  'Ceremony agents (groom, plan, review, retro, report, merge) are serialized by
   the graph and never run alongside one another, so they share ONE slot between
   them — counting them individually under-subscribes the dev pool by the number
   of ceremony roles.'
replaced by
  'Ceremony agents are serialized by the graph and share ONE slot between them,
   so count dev assignees only — the epic gate's reference carries that rule and
   the `jq` that implements it.'
The full rule, the roster of ceremony kinds and the jq that implements the count
are all in epic-definition-of-ready.md (its dev-pool check) and in
dispatch-policy.md, and the new section now sends the reader to the first of
those by name. Nothing lost, ~125 bytes recovered.

Alternative considered: taking the Blocked path per the byte-budget instruction.
Rejected because the instruction forbids cutting a RULE, and neither cut removes
one — both delete a second statement of a rule the same file (cut 1) or the
reference the new step points at (cut 2) still carries.
```

#### `agent-plugins-bbtl#3` — dev-so39-2

```
retro:decision:docs/plans/beads-coordination-rollout.md incident — the bead's criterion says 'no item's substance is changed — Files clauses only', but items 12 and 13 had NO Files clause, so satisfying criterion 2 required ADDING sections rather than extending them. I decided that adding a missing Files clause is not a substance change and did it.
Alternative considered: recording items 12 and 13 as 'cannot be audited, no Files clause' and leaving them untouched. Rejected — that satisfies criterion 1 (a result is recorded) while failing criterion 2 (no remaining item specifies a value with no named writer) on the two items where the omission is most complete, and it would leave the exact defect class this bead exists to remove sitting in two of the nine.
Second decision inside the same edit: item 13's new Files clause STATES the \$AGENT_ID-versus-BEADS_ACTOR mismatch rather than silently renaming the sketch's variable. Renaming it would have been a substance change to item 13's Change/Sketch and would have pre-empted a deferred bead (agent-plugins-qozp.14). The finding was instead routed as a comment on qozp.14 and on qozp.13 for the \$RUN_EPIC equivalent; neither bead's status was touched and both remain deferred.
```

#### `agent-plugins-gu3i#2` — dev-so39-2

```
retro:decision:docs/plans/beads-coordination-rollout.md item 4 carried the claim NOWHERE in its prose, so 'keep the conclusion at each amended site' had nothing to attach to; I added two sentences to item 4's **Change.** paragraph rather than burying the measurement in the Verify clause.
evidence: pre-edit item 4 read in full '**Change.** Every dispatch prompt exports `BEADS_ACTOR=<agent id>` before any `bd` call.' — no premise, no conclusion. The bead's acceptance criterion 3 requires 'Each amended site keeps the conclusion — export BEADS_ACTOR everywhere AND every bead carries an assignee before dispatch, as belt and braces rather than alternatives.' A Verify clause is a test procedure, not a statement a reader of the item takes away, so a conclusion placed only there would satisfy the grep and not the criterion. Alternative considered: leaving item 4's prose untouched and amending only the Verify clause — rejected because the item would then specify a procedure whose rationale lives in another document, which is the shape the sibling bead agent-plugins-bbtl exists to remove.
Also decided silently: the design doc's row-4 table cell and its late risk bullet state the same two arms at different lengths (one table cell, one bullet). The bullet carries the `bd ready --claim` -> [] arm; the cell does not, because a table cell that long stops being readable. dispatch-policy.md's four-row measured table remains the single full statement and was not touched.
```

### kind: `stale` — 2 finding(s) · **skill** audience by the merge test

#### `agent-plugins-4tjk#2` — dev-so39-3

```
retro:stale:agent-plugins-4tjk The bead's own AC grep (metadata.{0,40}(replace|entire metadata)) found exactly 3 sites, but a 4th carried the same false claim in different words: plugin/skills/groom-beads/references/capability-tier.md:11, a Contents/TOC line reading '--set-metadata versus --metadata -- additive versus replacing'. It sits outside the 40-char window and uses 'replacing' with no 'entire metadata' nearby, so the given pattern does not match it. Found by re-running the AC grep after editing and then widening the window (metadata.{0,80}replac) as a sanity check before closing -- the wider grep is what surfaced it. Corrected in the same commit; the narrow AC pattern itself is what undercounted, not repo drift after grooming.
evidence: grep -rniE 'metadata.{0,80}replac' plugin/ scripts/ docs/ specs/ --include=*.md --include=*.sh | grep -v references/shared/ -> plugin/skills/groom-beads/references/capability-tier.md:11 (before fix).
```

#### `agent-plugins-uva5#3` — dev-so39-4

```
retro:stale:agent-plugins-so39 the dispatch prompt attributed the wrong mechanic to the vendoring generator; the real one is stronger and the rule now rests on it.
evidence: the prompt said 'sync_dir does rm -rf "\$dest" then cp -r, so the generator owns whole directories and can never generate a section of a hand-authored file'. `grep -rn sync_dir scripts/` returns only scripts/generate-schemas.sh:183,199,200,201 — a different script, for strata schema artifacts. scripts/sync-shared-references.sh has no sync_dir; its writer is render(), which ends 'cat "\$1"' (scripts/sync-shared-references.sh:161-167) and writes one file per MAP entry.
so the whole-directory framing is wrong, but the conclusion it was reaching for is right and sharper: render() emits a four-line header plus the ENTIRE source file, so vendoring is whole-file-or-nothing. That is what test 1 in the shipped rule turns on.
```

### kind: `scope` — 1 finding(s) · **project** audience

#### `agent-plugins-bbtl#2` — dev-so39-2

```
retro:scope:docs/plans/beads-coordination-rollout.md item 2 — the very instance that motivated this bead — STILL names no writer, and I left it that way because it is outside 6-14 and its work is complete.
evidence: rollout.md item 2's Files clause reads in full: '`plugin/skills/sprint-plan/references/run-record.md` (the manifest key spec — the *only* place manifest keys are specified; `bead-bonanza`'s run-record covers run health and `base_sha`, not keys), `plugin/skills/sprint-plan/SKILL.md` (the rule).' `plugin/skills/bead-bonanza/scripts/pre-flight-check.sh` is still absent from it. The instance was fixed in the SCRIPT, not in the plan: `git show --stat 872f390` (agent-plugins-h3j8 P1 'pre-flight-check.sh does not emit concurrency_cap into the run manifest') touches exactly one file, plugin/skills/bead-bonanza/scripts/pre-flight-check.sh, 55 insertions 1 deletion. Nothing has ever amended item 2's prose.
Decided silently: no bead filed for it. Item 2's work is closed (agent-plugins-qozp.4), the rollout plan is transient by CLAUDE.md's 'Where documents live' table ('delete when the change lands'), and a wrong Files clause on a completed item has no future reader. If that judgement is wrong the fix is one line and belongs to whoever deletes the plan.
Also decided silently: item 12's and item 13's new Files clauses name `plugin/hooks/hooks.json` (new) as the hook destination, mirroring item 14's pre-existing '`plugin/monitors/monitors.json` (new) or this repository's own config, per D1' convention. Neither path exists in the tree; both are destinations, not writers, and D1 is explicitly unsettled, so inventing a settled path would have been worse than echoing the convention.
```

### kind: `sequencing` — 1 finding(s) · **project** audience

#### `agent-plugins-qozp.20#3` — dev-so39-5

```
retro:sequencing:agent-plugins-qozp.22 this bead's folded-in batch-id fix SUPERSEDES a deferred bead that carries a DIFFERENT fix direction. Left deferred, untouched, per dispatch instruction.

agent-plugins-qozp.22 ('batch-id is position in its wave, so agent ids collide across waves of the same run', P1, DEFERRED, metadata.grooming_status=out-of-scope-this-run) reports the same defect in the same lines of dispatch-policy.md, and proposes 'include the wave ordinal in the derivation (dev-<run-id>-<wave>-<batch>)'.

This bead's grooming selected run-wide numbering instead — <batch-id> is the batch's position in the run, numbered once across every wave — and that is what shipped. The two directions are mutually exclusive: only one of dev-<run-id>-<batch> and dev-<run-id>-<wave>-<batch> can be the derivation. qozp.22 is now stale rather than merely deferred, and reviving it as written would re-break what this bead fixed.

I did not undefer it, close it, or comment on it — the dispatch instruction says 'do not undefer anything'. Someone with the run in view should reconcile the two.

retro:scope:plugin/skills/sprint-plan/references/run-record.md filed as agent-plugins-b1gl (discovered-from qozp.20): the '## Wave table' example's Batch column still numbers per wave ('batch #1' in wave 1, 'shared-batch #1' in wave 2), which is the value the new run-wide rule rejects. Adjacent, not blocking — sprint-plan/SKILL.md's assign step names dispatch-policy.md as the derivation source and does not send the reader to the table for the number. run-record.md was not in this bead's file scope.
```

### kind: `process` — 2 finding(s) · **UNROUTED** — `process` is not one of the nine kinds

#### `agent-plugins-so39#1` — Claude

```
retro:process:duplicate-bead — the coordinator filed the SAME bead twice, 27 minutes apart: agent-plugins-vcka (16:13:59Z) and agent-plugins-xpn8 (16:41:04Z), both created_by Claude, both 'define bead manifests in .strata and validate before writing', same two call sites, same pipeline, near-identical acceptance criteria. sprint-scope caught it during admission; nothing earlier did. Two causes worth separating: (a) the coordinator did not run 'bd search' before 'bd create', and nothing in the skills tells it to — groom-beads warns about duplicates but only as an anti-pattern to reject during grooming, not as a pre-create step; (b) the first filing happened in a turn whose context the coordinator no longer held, which is the normal condition for a long run, so 'remember what you filed' is not an available strategy. Resolved by 'bd supersede agent-plugins-vcka --with=agent-plugins-xpn8'. Candidate fix: a pre-create search is cheap and mechanical - state it where a skill instructs an agent to file a bead.
```

#### `agent-plugins-so39#2` — Claude

```
retro:process:coordinator-dispatched-directly incident — the coordinator dispatched wave 1 itself rather than invoking sprint-wave, deliberately and with the reason recorded rather than by omission. Run 1 measured that sprint-wave, dispatched as a subagent, is given no Task/Agent tool and silently absorbs the batch it was meant to delegate — performing as self-verification the very check it exists to perform against an independent agent. Invoking it again would have reproduced a known failure. Two consequences the retro should weigh: worktree and branch names come from the harness rather than the dispatch id, so the harness-id to dispatch-id mapping is the coordinator's to record; and the coordinator holds the return-verification reasoning it is designed not to hold. Both are already filed against the skills (.sf/outbox/bead-bonanza/2026-08-24-no-dispatch-tool-is-a-hard-blocker.md), and plan item 8 makes direct dispatch the only path by deleting sprint-wave outright.
```

### kind: `proof` — 10 finding(s) · terminates here. No issue.

#### `agent-plugins-so39#5` — Claude

```
retro:proof:agent-plugins-4tjk — the dispatch-record gap recorded earlier is NARROWER than first stated, and the correction is measured. agent-plugins-4tjk (dispatched to run-bead directly) returned with dispatch_state=null and terminal_state=null; agent-plugins-qozp.19 and agent-plugins-d6zz (dispatched to sprint-batch, which invokes run-bead per bead) both returned dispatch_state=returned, terminal_state=completed. Same run, same wave, same session. So the keys are written on the sprint-batch path and not on the bare run-bead path - the hypothesis in the earlier comment, now confirmed rather than assumed. That locates it: run-bead is the skill both paths share, so the write belongs there rather than in sprint-batch.
```

#### `agent-plugins-so39#6` — Claude

```
retro:proof:byte-budget — the 20,000-byte skills-strict cap is now the run's binding constraint, measured on the merged head rather than predicted: sprint-plan/SKILL.md 19,946 (54 bytes free), groom-beads/SKILL.md 19,890 (110 free), bead-bonanza/SKILL.md 19,240 (760 free). Three of this wave's beads had to pay for their additions by cutting text, and agent-plugins-qozp.20 overran by 236 bytes and funded it by deleting two RESTATEMENTS rather than rules - a judgement it recorded as a retro:decision and named both cuts. That is the right call once. It is not a strategy: at 54 bytes free, the next bead touching sprint-plan/SKILL.md cannot add a sentence, and the pressure will fall on whoever is least able to tell a restatement from a rule.
```

#### `agent-plugins-so39#8` — Claude

```
retro:proof:coordinator incident — the coordinator put a FALSE MECHANICAL CLAIM in a dispatch prompt, and the bead's own work caught it. dev-so39-4's brief stated that sync-shared-references.sh's sync_dir does 'rm -rf $dest' then 'cp -r', and used that to argue the vendoring generator owns whole directories. That function lives in scripts/generate-schemas.sh and belongs to strata artifacts; sync-shared-references.sh's render() ends 'cat $1'. The correction matters rather than being cosmetic: whole-file-by-cat is the MECHANICAL BASIS of the vendor-or-prose rule that shipped, so the rule rests on the measured behaviour and not on the coordinator's misattribution. Provenance: the rm -rf detail came from sprint-plan's ledger notes about generate-schemas.sh and the coordinator transposed it onto the other script when writing the prompt. Same failure class as agent-plugins-4tjk, one level up - a plausible claim carried into work without being measured at the point of use.
```

#### `agent-plugins-so39#10` — Claude

```
retro:proof:progress-streaming — CORRECTION to the coordinator's own stall finding, in the coordinator's favour but worth recording accurately. The stall was real: the coordinator ended a turn saying 'moving to the retro' with nothing dispatched and nothing in flight, and the human asked before anything resumed. But the claim that no mechanical signal could have caught it is FALSE. The durable recovery trigger armed at first dispatch (trig_01Tr5qJXx7ZnRFVvZohSZpwc) fired at 17:57:32 carrying exactly the right instruction - 'if that run already finished, cancel this and do nothing; otherwise invoke sprint-recovery' - and the run had not finished, so it would have resumed the ceremony. The human simply asked a few minutes earlier. So the watchdog design works and needs no new mechanism; what it cannot do is close the gap between a stall and the next firing, which was set at 50 minutes for this run. The finding worth keeping is about the INTERVAL and about the coordinator, not about a missing signal.
```

#### `agent-plugins-d6zz#1` — dev-so39-1

```
retro:proof Demonstrated in a throwaway scratch database (bd init --prefix bbdemo, separate git repo, deleted after): with admitted=[bbdemo-2l5.1] and both beads open, bb-status.sh run reported 'ready front vs manifest.admitted: 1 bead(s) NOT admitted: bbdemo-2l5.2 — resolve before dispatch'; with admitted extended to both ids it reported 'ready front matches manifest.admitted — no creep'; with both beads closed (empty ready front) it still reported 'no creep' without error. Confirmed via bd search that no bead landed in this project's tracker from the demo.
```

#### `agent-plugins-4tjk#1` — dev-so39-3

```
retro:proof:groom-beads Reproduced all three measurements myself before writing the correction, in a throwaway scratch db (bd init --prefix probe, /tmp scratchpad, deleted after).
Top-level merge: bd update probe-oz0 --metadata '{"alpha":1,"beta":"two","gamma":["x","y"]}' then bd update probe-oz0 --metadata '{"delta":9}' then bd show probe-oz0 --json | jq -c '.[0].metadata' -> {"beta":"two","alpha":1,"delta":9,"gamma":["x","y"]}. All four keys survive; confirms merge, not replace.
Nested replacement: bd update probe-pj2 --metadata '{"nest":{"a":1,"b":2}}' then bd update probe-pj2 --metadata '{"nest":{"c":3}}' then bd show probe-pj2 --json | jq -c '.[0].metadata' -> {"nest":{"c":3}}. a and b are gone -- nested object value replaced wholesale, confirming the still-true half of the old claim.
Still-true trap: bd update probe-pj2 --set-metadata arr='["a","b"]' then bd show --json | jq -c '.[0].metadata' -> includes "arr":"[\"a\",\"b\"]" -- --set-metadata stores the literal string, confirming why --metadata is reached for at all.
bd version: 1.2.2 (dev).
```

#### `agent-plugins-uva5#2` — dev-so39-4

```
retro:proof:scripts/sync-shared-references.sh a red gate on my first full checks run was a pre-existing flake, not my change; filed as agent-plugins-b9ft with a discovered-from edge.
evidence: first `bb-checks.py` run in worktree agent-a2c5301b15d893e08 exited 1 with '=== shared-refs FAILED (exit 1)' and 'ORPHAN plugin/skills/sprint-plan/references/shared/epic-definition-of-ready.md — header names scripts/sync-shared-references.sh, but no MAP entry generates it'. That MAP entry exists: scripts/sync-shared-references.sh:67 reads 'sprint-plan<TAB>groom-beads/references/epic-definition-of-ready.md'. Immediately rerunning the same check exited 0.
reproduced: 30 consecutive runs of `bash scripts/sync-shared-references.sh --check` on an unchanged tree -> 2 failures, naming a DIFFERENT mapped file each time (sprint-retro/.../findings-capture-protocol.md; groom-beads/.../work-order-definition-of-ready.md).
negative control: a scratchpad copy of the script with exactly one line changed — 'if printf %s\\n "\$managed_dests" | grep -qxF -- "\$rel"' replaced by 'if grep -qxF -- "\$rel" <<< "\$managed_dests"' — ran 100 consecutive times with 0 failures. Nothing else changed, same tree, same worktree.
cause: grep -q closes the pipe on first match, printf dies on SIGPIPE, and the script's 'set -euo pipefail' turns that into a failed pipeline, so the 'if' reads a MATCH as 'not managed'. Exactly the inversion AGENTS.md 'Pipes hide exit statuses' documents.
baseline note: agent-plugins-so39's checks_baseline records shared-refs as 'passed', so a naive baseline comparison would have charged this to me.
```

#### `agent-plugins-qozp.20#1` — dev-so39-5

```
retro:proof pre-fix reproduction of 'no producer', observed before the fix and again after.

Command (pre-fix, against HEAD): git grep -n -- '--assignee' HEAD -- plugin/skills/
Output — both hits READ-side:
  dispatch-policy.md:31: `bd ready --assignee <id>` that agent's queue.
  dispatch-policy.md:57: `bd ready --assignee groom-<run-id>` the groom agent's queue

Same command on the worktree after the fix adds exactly one WRITE-side hit:
  sprint-plan/SKILL.md:335: here and not at the readiness gate. Per row: `bd update <id> --assignee

Ordering, pre-fix: sprint-plan/SKILL.md had '### 1b. Gate the EPIC as well as its
beads' at line 111, i.e. before '### 2. Groom what fails' (118) and before the
classification/dispatch-mode sections that build the wave table (215-327). The
gate therefore demanded dev-<run-id>-<batch-id> before the wave table <batch-id>
derives from existed. Post-fix the gate lives in '## Assign every planned bead,
then gate the epic' at line 331, after '## The dev pool is capped' and before
'## Return the plan'.
```

#### `agent-plugins-bbtl#1` — dev-so39-2

```
retro:proof:docs/plans/beads-coordination-rollout.md the grooming note's warning about item 7 is confirmed, and the audit found two items with NO Files clause at all — a failure mode the bead did not anticipate.
evidence: `grep -n '^### [0-9]' docs/plans/beads-coordination-rollout.md` returns headings for 1,2,3,4,5,6,8,9,10,11,12,13,14 — no 7. Item 7 is at line 12, `## 7 · The run-epic shape, and the constraint that governs it`. Before this bead's edit, `grep -n '^\*\*Files\.\*\*'` returned 12 hits for 14 items: items 12 and 13 carried **Change** / **Sketch** / **Verify** / **Risk** / **Size** and no **Files** at all. So the audit's own method — 'for every item whose Files clause specifies a key … add the producing component to that item's Files clause' — has no anchor on those two; an auditor iterating over existing Files clauses would have skipped them silently, which is the same shape of miss as the `^### [0-9]` skip of item 7.
The two writer-less values on those items are real, not hypothetical: `grep -rn 'RUN_EPIC' plugin/` -> zero hits (item 12's sketch reads `\$RUN_EPIC`); `grep -rn 'AGENT_ID' plugin/` -> no writer (item 13's sketch reads `\$AGENT_ID` while item 4 writes `BEADS_ACTOR`).
```

#### `agent-plugins-gu3i#1` — dev-so39-2

```
retro:proof:agent-plugins-gu3i the corrected acceptance grep finds the sites; the criterion's original plain-text grep is confirmed vacuous.
evidence: on the pre-edit file (git show HEAD:docs/plans/beads-coordination-design.md), `grep -n 'silently rewrites assignee'` exits 1 with no output — the live text backticks the identifier. The corrected pattern `grep -rniE 'silently rewrites .?assignee.?|corrupts the coordinator' docs/plans/` returned the two sites (design.md:31 and design.md:1164). After the edit the same pattern returns exactly one line, design.md:31, and it is qualified: 'silently rewrites `assignee` on an **unassigned** bead (exit 0, the git `user.name` fallback), and is refused loudly on an already-assigned one'.
```

## 7 · Records the harvest does not reach, quoted in full

These are not `retro:` findings and carry no finding id. They are reproduced
here because four of the fourteen candidates rest on them, and because a retro
running from the epic id alone — as `sprint-retro` is designed to — would not
have seen them. That is candidate **C4**.

### The review-loop incident — `agent-plugins-so39` comment 14, bare `incident:` prefix

```
incident: review pass 1 filed 10 findings, admitted ZERO, and three of them (so39.4, so39.5, so39.6) are defects IN THE CORRECTIONS THIS RUN SHIPPED — on the exact thesis the run was opened to serve. standing_auth.admit_review_findings is {value:false, source:'default'}: nobody chose it, and the skill's in-scope test needs gate authorization for scope growth, so severity cannot admit them. The review loop therefore ENDS at pass 1 with 0 admitted, inside the budget of 2 loops. Recorded as an incident rather than escalated, per the operating principle.

so39.5 is the one the report must lead with, and the coordinator confirmed it by independent measurement rather than accepting the finding: '--set-metadata stores every value as a string' is FALSE on bd 1.2.2 — JSON scalars keep their type (5 -> number, true -> boolean, 1.5 -> number, null -> null); only arrays, objects and non-numeric text stay strings. The coordinator carried that false universal into six dispatch prompts, one bead description and two epic comments this run, and it landed inside agent-plugins-4tjk's corrected text. A run that existed to stop unmeasured claims propagating shipped a new one in the same paragraph it fixed the old one.

Contrast worth keeping: the review also REJECTED a code-review finding as a false positive rather than filing it (the claim that 726ab33 deleted the last explanation of the merge-collision hazard — the rationale survives in run-record.md, cited two lines above the deletion), and dropped two more as already filed. So the ten are what survived verification, not what was reported.
```

### The coordinator's independent measurement — `agent-plugins-so39.5` comment 1, unprefixed

```
coordinator verification — CONFIRMED by independent measurement, and the claim is one the coordinator propagated all session. bd 1.2.2, throwaway database smlab, one bd update carrying seven --set-metadata pairs:

  num=5        -> 5        (number)
  b=true       -> true     (boolean)
  f=1.5        -> 1.5      (number)
  n=null       -> null     (null)
  s=hello      -> hello    (string)
  arr=["a","b"] -> '["a","b"]'  (STRING)
  obj={"k":1}   -> '{"k":1}'    (STRING)

So --set-metadata parses JSON scalars and keeps their type; only arrays, objects and non-numeric text are stored as strings. 'stores every value as a string, arrays included' is false as a universal, and the true statement is narrower and more useful: arrays and objects need --metadata because --set-metadata stringifies THEM, not everything.

Provenance, because it matters more than the fix: the coordinator carried this false universal into at least six dispatch prompts this run, into agent-plugins-o76b's description, and into comments on two epics - and it then landed inside agent-plugins-4tjk's corrected text. The run opened to stop unmeasured claims propagating shipped a new one in the same paragraph. Nobody measured it because it was inherited as a premise rather than treated as a question, which is precisely the failure agent-plugins-4tjk documents.
```

### The `qozp.22` disposition — `agent-plugins-so39` comment 8, `amend:` prefix

```
amend: agent-plugins-qozp.20 merged. It redefined <batch-id> as position in the RUN, numbered once across every wave, which makes agent-plugins-qozp.22 STALE rather than merely deferred - qozp.22 reports the same defect but proposes the opposite fix (dev-<run-id>-<wave>-<batch>), so reviving it as written would re-break what just shipped. It was correctly left untouched under the do-not-undefer constraint. Coordinator disposition: qozp.22 must be closed as superseded or rewritten before any future run admits it; it is NOT safe to pick up from the backlog as-is. Recorded here rather than acted on because it sits under epic agent-plugins-qozp, outside this run's admitted set.
```

### The `bbtl` amendment incident — `agent-plugins-so39` comment 12, `amend: incident` prefix

```
amend: incident on agent-plugins-bbtl — rollout items 12 and 13 had no **Files.** clause at all, so satisfying 'every specified value names its writer' meant adding sections, not extending them, against a criterion reading 'Files clauses only'. Added them; recorded on the bead. Related: item 13's sketch reads $AGENT_ID while item 4 writes BEADS_ACTOR and nothing writes AGENT_ID — flagged in the Files clause and routed to the deferred agent-plugins-qozp.14, not fixed here.
```

### The bounce mechanism — `agent-plugins-xpn8` comment 2, `plan-so39`

```
plan-so39: set status=deferred as the mechanism that keeps this bounce out of the run. It was BLOCKED by agent-plugins-4tjk, and 4tjk is dispatched in this run — so on 4tjk's close xpn8 would have entered 'bd ready --parent agent-plugins-so39 --exclude-type=epic' with nothing flagging it, and it IS in manifest.admitted, so a ready-front-minus-admitted diff would not have caught it either. Verified after deferring: the ready front returns 10 beads and does not include this one, and 'bd human list' still lists it (Status: deferred). The 'blocks' edge from 4tjk is intact. Undeferring is the coordinator's or the human's call, not a dev's.
```


## 8 · Counts

| | |
|---|---|
| harvested | **28 findings in 27 comments**, 8 beads (11 including the three light-tier beads that recorded none) |
| routed — skill | 10 |
| routed — project | 6 |
| terminated (`proof`) | 10 |
| **unrouted** | **2** — both `retro:process:`, both the coordinator's |
| deduped against beads already filed | 3 (`agent-plugins-b9ft` ×2, `agent-plugins-b1gl` ×1) |
| candidates | **14** — 9 new skill issues, 3 project beads, 2 extensions of run-1 outbox files, 1 correction to a run-1 outbox file (C4, C5 and C6 are the three that touch existing files) |
| filed | 0 — **awaiting the human's decision** |
| declined | 0 |
| outbox written | 0 |

**Nothing is filed.** `sprint-retro` files nothing without a human's yes, and
this retro was dispatched read-only besides. Every candidate in section 5 is
written to the point where filing is mechanical: a project-audience candidate is
one `bd create`, a skill-audience candidate is one file at
`.sf/outbox/<target>/2026-08-24-<slug>.md` in the issue template's shape, and
C4/C5/C6 are edits to named existing files.

Two closing steps were not taken, deliberately:

- **The `rfcs/retro-log.md` row.** It records the outcome counts, which do not
  exist until the decisions do, and the skill requires it committed. Both are
  outside this retro's authority. The row should read: `2026-08-24 | self |
  bead-bonanza sprint (epic agent-plugins-so39) | [findings](sprints/2026-08-24-run-2-coordination-corrections-findings.md)
  | 28 findings harvested, 10/11 admitted beads closed (1 bounced), 14
  candidates drafted, <n> accepted …`.
- **`retro:filed:` back-references.** They are written the moment a candidate is
  filed, so that recurrence counts unfiled findings only. None exist yet.

## 9 · What this document is not

It is not the run's report. `sprint-report` writes that from the manifest and
this document's path (`metadata.findings_path`); neither skill reads the
other's output. The 10 deferred review findings (`agent-plugins-so39.2`
through `.11`), the 22 beads excluded at admission, and `agent-plugins-xpn8`'s
`human` label with its three named resolutions are all still open decisions,
untouched by this retro and not candidates here — they are the report's
material and the next run's.

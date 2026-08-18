---
title: "RFC: Reconcile the batch template with harness-managed worktrees, and
  give a gate baseline owned by two beads in one wave a rule"
description: The batch prompt and concurrent dispatch prescribe incompatible
  worktree models, harness worktrees land inside the repo the skill says they
  must never be inside, and a baseline split across concurrent beads has no
  rule, so the orchestrator improvises all three.
slug: rfc-self-worktree-composition-and-shared-gates
type: rfc
status: ready-for-review
updated_at: 2026-08-13T00:00:00Z
origin: self
skill: bead-bonanza
skill_version: "1.4"
model: claude-opus-5
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "n/a (in-repo dogfooding)"
---

# RFC: Reconcile the batch template with harness-managed worktrees, and give a gate baseline owned by two beads in one wave a rule

## Summary

One `bead-bonanza` run of seven beads drained cleanly — 7 closed, 0 blocked,
0 skipped — but the orchestrator had to improvise three times and record two
`## Amendments` entries to get there. Each improvisation was forced by the
skill prescribing two things that cannot both hold, or by saying nothing where
a decision was unavoidable:

1. The batch prompt tells a subagent to create **one worktree per bead**;
   concurrent dispatch tells the orchestrator to use `Agent` with
   `isolation: "worktree"`, which yields **one worktree per agent**. A batch
   subagent processing two beads cannot satisfy both.
2. The worktree mechanics state worktrees must be a **sibling of the repo,
   never inside it**, and give the reason. The harness path the same file
   prescribes puts them **inside** the repo, producing exactly the
   `git status` dirt the rationale warns about — observed as a stop hook
   demanding the worktrees be committed, mid-run.
3. The gate baseline is defined against "a pre-existing failure … charged to
   the first bead". It has no rule for a baseline whose findings are **owned
   by two beads dispatched in the same wave**, where "gates must pass" is
   unsatisfiable for either alone and satisfying it literally requires each
   to edit the other's files.

A fourth, smaller defect: the findings-harvest pipeline the skill points at is
keyed on a shared label, which a `bead-bonanza` queue does not have — the skill
requires an enumerated id list instead. Run as written it returns zero, and the
protocol then instructs the agent to treat zero as a broken query and debug it.

## Provenance

| Field | Value |
|---|---|
| Skill critiqued | `bead-bonanza`, SKILL.md `version: "1.4"` |
| Task | Run `1786603162-7`: 7 groomed beads (3 bugs, 2 CLI features, 2 documents) across 2 planned waves |
| Model (orchestrator) | `claude-opus-5` |
| Model (subagents) | `sonnet` (tier `standard`, 5 beads), `haiku` (tier `light`, 1 bead), resolved from `AGENTS.md`'s `## model lineup`. `CLAUDE_CODE_SUBAGENT_MODEL` unset, so tier routing was honoured. |
| Effort | Not exposed by this harness; recorded as `unknown` rather than guessed |
| Harness | Claude Code on the web (remote container), `Agent` tool with `isolation: "worktree"` |
| Runs observed | 1 |
| Run date | 2026-08-13, branch `claude/validate-corpus-spec-cli-0t0cxo`, PR #22, base `b8e9436` |
| Evidence base | `rfcs/sprints/2026-08-13-template-drift-and-check-report.md` (16 verbatim `retro:` findings plus the orchestrator's own record), and `dispatch.md`'s `## Amendments` log from the run |
| References invoked | `subagent-prompts.md`, `progress-streaming.md`, `operating-principles.md`, `references/shared/findings-capture-protocol.md`, `references/shared/work-order-definition-of-ready.md` — none carries a `version:` field |
| Scripts invoked | `pre-flight-check.sh`, `bb-progress.sh`, `bb-close-bead.sh`, `bb-backup-run.sh` — all behaved as documented |

**Relationship to a prior RFC.** `rfcs/inbox/bead-bonanza/rfc-self-report-citations-and-batch-base.md`
(skill_version 1.3, run 2026-08-12) proposes fixing the batch template's
**base** — that `worktree add … HEAD` cannot advance within a batch, so a
batch's beads do not build on each other. Proposal 1 below is about the same
block of text but a different defect: its **composition** with harness-managed
isolation. They should be triaged together. Worth noting as evidence for that
RFC: this run used one harness worktree for the whole batch, which sidesteps
its bug entirely — both beads were sequential on one branch, so the second did
build on the first.

## Evidence

### E1 — the batch prompt and concurrent dispatch prescribe incompatible worktree models

`plugin/skills/bead-bonanza/references/subagent-prompts.md:177-181`, the shared
batch preamble:

> ```
> Worktrees: one per bead, created by YOU, held ONE AT A TIME. Before each bead:
>   git -C <repo> worktree add -b <suffix> <repo>/../.bb-worktrees/<suffix> HEAD
> Work that bead entirely inside its worktree. After its close commit lands, tell
> the orchestrator and move on — never hold two worktrees open at once, because a
> worktree that exists means "this bead is unfinished" to every other reader.
> ```

`subagent-prompts.md:381-383`, concurrent dispatch:

> Spawn several beads at once — one `Agent` call per bead with
> `isolation: "worktree"`, all in a single message.

`isolation: "worktree"` provisions one worktree per **agent**. The batch
preamble requires one per **bead**, created by the subagent itself. For the
`j5n`+`k9j` batch (2 beads, 1 agent) these are mutually exclusive, and the
skill offers no rule for which yields.

**What the orchestrator did:** gave the batch a single harness worktree and
instructed it not to create more, then recorded the deviation as
`dispatch.md` `## Amendments` entry 1. The alternative — letting a batch agent
run `git worktree add` against the shared checkout while three other subagents
were concurrently active in it — was judged the greater risk, but that judgment
is the orchestrator's invention, not the skill's.

### E2 — harness worktrees land inside the repository, contradicting the skill's stated rationale

`subagent-prompts.md:310-316`:

> ```sh
> REPO="$(git rev-parse --show-toplevel)"
> WT="$REPO/../.bb-worktrees/<suffix>"        # sibling of the repo, never inside it
> git -C "$REPO" worktree add -b <suffix> "$WT" HEAD
> ```
>
> A worktree inside the repository would read as dirt in every `git status` the
> run takes, which is why it is a sibling.

`subagent-prompts.md:319-321` then concedes the harness path:

> Where the harness creates the worktree for you (`Agent` with
> `isolation: "worktree"`), it uses `worktree.baseRef` instead, which is why
> that setting is a hard gate.

**Observed:** the harness placed every worktree at
`/home/user/agent-plugins/.claude/worktrees/agent-<id>` — inside the
repository. `git worktree list` during wave 1:

```
/home/user/agent-plugins/.claude/worktrees/agent-a0eda40486060fe31  f9afdc7 [worktree-agent-a0eda40486060fe31] locked
/home/user/agent-plugins/.claude/worktrees/agent-a9dde798784a29234  f9afdc7 [worktree-agent-a9dde798784a29234] locked
/home/user/agent-plugins/.claude/worktrees/agent-ac8ec897df796484a  0ecad82 [worktree-agent-ac8ec897df796484a] locked
```

A stop hook then fired mid-run: `There are untracked files in the repository.
Please commit and push these changes to the remote branch.` That is precisely
the "dirt in every `git status`" the rationale predicts. An orchestrator that
complied would have committed in-flight worktree contents onto the sprint
branch.

`grep -rn "gitignore\|untracked" SKILL.md references/*.md` returns no rule for
this: the only hits are `SKILL.md:900` (about `.beads/*.jsonl` staging) and
`recovery.md:35` (about a staging pathspec). **No rule at the worktree
lifecycle step for "the harness put the worktree inside the repo".**

**What the orchestrator did:** added `.claude/worktrees/` to `.gitignore`
mid-run — a file the skill never mentions.

### E3 — no rule for a gate baseline owned by two beads in the same wave

`SKILL.md:265-267`, the whole of what the skill says about baseline ownership:

> `--gate-baseline` also runs the gate profile on the untouched tree, so a
> pre-existing failure is recorded as the baseline instead of being charged to
> the first bead.

That resolves *pre-existing vs. this bead*. It does not resolve *this bead vs.
its sibling in the same wave*.
`grep -rn "baseline" SKILL.md references/*.md | grep -iE "split|share|own|two beads|per-bead"`
returns nothing.

**Observed:** `scripts/check-template-drift.py` reported 8 findings at
baseline — 7 in `plugin/skills/retro-spec/references/rfc-template.md` (owned by
`agent-plugins-3xx`) and 1 in `plugin/skills/spec/references/work-order-template.md`
(owned by `agent-plugins-j5n`). Both were dispatched concurrently in wave 1.
Under the skill's own gate instruction ("the gate as a whole fails if any of
the three failed", and the per-bead template's "On gate failure, fix the root
cause and rerun"), neither bead can pass: `3xx` cannot clear `j5n`'s finding
without editing `work-order-template.md`, and vice versa. Both would have
breached the scope constraint in their own prompts and collided at merge.

**What the orchestrator did:** replaced "gates must pass" with a per-bead
target in each prompt — `3xx` 8→1, `j5n` 8→7, run ends at 0 — plus an explicit
prohibition on touching the other's file. Recorded as `## Amendments` entry 3.
Both beads hit their targets exactly; the gate reached 0 once both merged.

### E4 — the prescribed findings harvest cannot run on a bead-bonanza queue

`SKILL.md:729-731` routes the harvest to the vendored protocol:

> Harvest with the `bd export | jq` pipeline in
> `references/operating-principles.md` ("Findings harvest"), once filtered to
> closed and once unfiltered.

`references/shared/findings-capture-protocol.md:93-94`:

> ```
> bd export | jq -r 'select((.labels//[]) | index("<feature-slug>"))
>   | .comments[]? | select(.text|startswith("retro:")) | .text'
> ```

The selector requires a shared label. A `bead-bonanza` queue has none — the
skill *requires* the opposite at `SKILL.md` Phase 2.5:

> **The enumerated bead list.** Pass `plan-sprint` explicit bead ids, never
> the filter that produced them.

Run as written against this run's seven beads, the selector matches nothing.
The protocol then compounds it at `findings-capture-protocol.md:98-101`:

> Assert the count is non-zero before aggregating — a harvest that returns
> zero findings across a completed feature is a broken query, not a clean run

so an agent following both is told to debug a query that cannot succeed by
construction.

**What the orchestrator did:** substituted an id-membership selector
(`select(.id as $i | $ids | index($i))`), which returned 16 findings across all
7 beads.

### E5 — "the first worktree subagent" is undefined under concurrent dispatch

`subagent-prompts.md:278-281`:

> **Verify before every wave** — this is global state the repo doesn't own and
> can change without warning. Note session HEAD (`git rev-parse --short HEAD`)
> and give the **first** worktree subagent this instruction:

Concurrent dispatch sends every agent of a wave "all in a single message"
(`:382`), so no agent is first. **What the orchestrator did:** gave the check
to all four wave-1 subagents. All four reported `ancestor-exit=0`.

### E6 — a missing `committed` event silently disables the Monitor for that bead

`progress-streaming.md:81`:

> **Filter to `committed|blocked` only** — those are the events the
> orchestrator acts on.

`agent-plugins-qz9.3`'s ledger, read from disk after it returned:

```
{"bead":"agent-plugins-qz9.3","phase":"gates-passed"}
{"bead":"agent-plugins-qz9.3","phase":"retro-written"}
{"bead":"agent-plugins-qz9.3","phase":"pushed"}
```

No `committed` event, so the Monitor never fired for the run's last bead and
the user saw silence for its ~6-minute duration — against the file's own
promise at `:75`, "Don't leave the user staring at silence."

Correctness was unaffected: the acceptance check deliberately does not consult
the ledger, and the bead was verified closed, committed and merged from state.
The gap is that nothing reconciles "a subagent returned" against "its terminal
event never arrived", so the degradation is invisible unless the orchestrator
reads the ledger by hand.

## Proposals

### P1 — State which worktree model wins when the harness provisions them (blocker, gap)

**Finding shape:** gap — no rule at the shared batch preamble
(`subagent-prompts.md:177`) for composing with `isolation: "worktree"`.

**Change.** Add to the shared batch preamble, replacing the opening
`Worktrees:` block's first sentence:

> Worktrees: **where the orchestrator provisioned one for you (the harness's
> `isolation: "worktree"`), you have exactly ONE worktree for every bead in
> this batch — process them in order, one commit each, and do NOT create
> more.** Where it did not, create one per bead yourself, held ONE AT A TIME:
> `git -C <repo> worktree add -b <suffix> <repo>/../.bb-worktrees/<suffix> HEAD`

And add to the orchestrator's worktree lifecycle section: a batch dispatched
with `isolation: "worktree"` gets one worktree for the batch, and the in-flight
signal is correspondingly coarser — the worktree existing means the BATCH is
unfinished.

**Why this and not the reverse.** Letting a batch subagent create worktrees is
strictly worse under concurrent dispatch: it runs `git worktree add` against a
checkout other agents are simultaneously using. It also inherits the
same-base bug that `rfc-self-report-citations-and-batch-base` documents, which
the one-worktree form avoids by construction.

### P2 — Give the harness worktree location a rule, and gitignore it (blocker, gap)

**Finding shape:** gap — no rule at the worktree lifecycle
(`subagent-prompts.md:304-321`) for a harness that ignores the sibling
requirement.

**Change.** Amend `subagent-prompts.md:319-321` to:

> Where the harness creates the worktree for you (`Agent` with
> `isolation: "worktree"`), it uses `worktree.baseRef` instead, which is why
> that setting is a hard gate — **and it may place the worktree INSIDE the
> repository (observed: `.claude/worktrees/`), which the sibling rule above
> exists to prevent. The path is not the orchestrator's to choose, so neutralise
> it instead: ensure the harness's worktree directory is gitignored before the
> first dispatch. Untracked worktrees inside the repo will otherwise be
> reported as uncommitted work by any tooling that reads `git status`, and
> committing them lands in-flight bead state on the sprint branch.**

Add to Phase 0's list of things no script settles: *"Confirm the harness's
worktree directory is gitignored — check `git status --porcelain` is clean
after the first worktree is created, not just before."*

### P3 — Define baseline ownership when two beads in one wave share a gate (blocker, gap)

**Finding shape:** gap — no rule at `SKILL.md:265-267` for a baseline split
across concurrently dispatched beads.

**Change.** Add to the gate-baseline paragraph:

> **Where a single gate's baseline findings are owned by more than one bead in
> the same wave, the wave-level instruction "gates must pass" is unsatisfiable
> for each of them alone, and satisfying it literally requires each to edit the
> other's files.** Split the baseline at dispatch: give each bead its own
> numeric target (`this gate holds N findings; yours are the M in <file>; your
> target is N → N-M, and the remainder belongs to <other bead> and is correct
> to leave`), plus an explicit prohibition on touching the other's file. State
> the run-end target (0, or whatever remains) so the orchestrator can verify
> the whole after merging. Do NOT tell either bead the gate must exit 0.

Add the same split to the per-bead and batch prompt templates' `Gates:` block,
as an optional `Your target for <gate>:` line.

### P4 — Give the findings harvest a selector that matches an enumerated queue (snag, script-defect)

**Finding shape:** script-defect — the prescribed command returns zero on a
valid, drained `bead-bonanza` run.

**Exact command (as prescribed), observed vs expected:**

```
$ bd export --output /tmp/h.jsonl
$ jq -r 'select((.labels//[]) | index("agent-plugins-…")) | .comments[]? | select(.text|startswith("retro:")) | .text' /tmp/h.jsonl
# observed: (no output)
# expected: 16 findings across the run's 7 beads
```

**Minimal repro:** close two beads with `retro:` comments and no shared label;
run the selector with any slug. Returns nothing.

**Change.** In `SKILL.md`'s sprint-retro "Harvest first, then assert"
paragraph, state the `bead-bonanza` form explicitly rather than deferring
wholly to the protocol:

> The protocol's selector filters on a shared label, which a feature's closure
> set has and **a bead-bonanza queue does not** — this skill's queue is an
> enumerated id list by construction (Phase 2.5). Filter on id membership
> instead, passing the run's committed queue:
>
> ```
> bd export --output <f> && jq -r --argjson ids '["<id>","<id>",…]' \
>   'select(.id as $i | $ids | index($i)) | .comments[]?
>    | select(.text|startswith("retro:")) | .text' <f>
> ```
>
> The protocol's zero-is-a-broken-query assertion still applies — but check the
> selector shape first, because the label form returns zero here by
> construction rather than by fault.

### P5 — Scope the base-ref ancestor check to every worktree subagent in a wave (polish, misleading-text)

**Finding shape:** misleading-text — `subagent-prompts.md:280` says
"the **first** worktree subagent", which has no referent when a wave is
dispatched in one message (`:382`).

**Change.** Replace "and give the **first** worktree subagent this
instruction" with:

> and give **every worktree subagent in the wave** this instruction — under
> concurrent dispatch they are sent in one message and none is "first", and
> the check is two commands, so per-agent coverage costs nothing and catches a
> per-worktree fault the wave-level check cannot see:

### P6 — Reconcile a returned subagent against its terminal progress event (polish, gap)

**Finding shape:** gap — no rule at Phase 3's post-return steps for "the
subagent returned but no `committed`/`blocked` event ever arrived".

**Change.** Add to Phase 3's post-return list, after the acceptance check:

> **Note a missing terminal event, do not act on it.** Where a subagent
> returned `completed` but its progress file carries no `committed` event, the
> Monitor never fired for that bead and the user saw silence. This does not
> affect acceptance — the check above reads state, not the ledger — but say so
> in the Phase 4 report, because an unremarked gap trains the reader to trust a
> stream that has silently stopped covering part of the run.

### P7 — Say what to do when a draft PR already exists (polish, gap)

**Finding shape:** gap — no rule at `SKILL.md:343`, which assumes the PR does
not exist.

**Observed:** `create_pull_request` returned
`422 Validation Failed … A pull request already exists for
brettcallaghan:claude/validate-corpus-spec-cli-0t0cxo`. PR #22 was already open
as a draft on the branch, opened by the repository owner.

**Change.** Extend the Draft PR bullet:

> Check whether one already exists for the branch before creating one; where it
> does, that satisfies this step — record its number in `dispatch.md` and leave
> its title and body alone. They are the author's, and a run that rewrites them
> is editing something it did not create.

## Considered and Not Proposed

- **`check-skills.sh` rejects cross-skill backticked paths.** `agent-plugins-j5n`
  cited `groom-beads`'s `references/bead.schema.json` from a `spec` skill
  document and rule 2 flagged it, because backticked paths must resolve
  relative to the citing file's directory. Real friction, but a property of
  `check-skills.sh` and the skill-authoring conventions — not of
  `bead-bonanza`. Belongs in a separate RFC against the skill-authoring
  standard.
- **`[[wikilink]]` fails for a document outside the corpus root.**
  `agent-plugins-lwu` hit `reference/unresolved` citing
  `[[ac-traceability-standards]]`, whose file carries `type: standards`
  frontmatter but lives outside `specs/`. A finding about the spec corpus
  model; belongs to `retro-spec`, not here.
- **An acceptance criterion asserted a document count that was wrong.**
  `agent-plugins-ziv`'s AC said the checked count "increases by exactly 1"; it
  increased by 2. A grooming defect, routed to `groom-beads`.
- **Pre-flight's `worktree.baseRef` block.** It exited 1, named the exact
  remedy and the file to write it in, and re-ran green. This is the skill
  working as designed — recorded as positive evidence, not a finding.
- **`bb-close-bead.sh`, `bb-progress.sh`, `bb-backup-run.sh`.** All behaved
  exactly as their published interfaces state across 7 beads. No finding.

## Non-Goals

- Changing the concurrency model. Concurrent dispatch worked: four units ran in
  parallel, all four ancestor checks passed, every merge fast-forwarded after
  its rebase, and no rebase conflicted.
- Changing the acceptance check's five signals. Its deliberate refusal to
  consult the progress ledger is what kept E6 from becoming a correctness
  failure; P6 only asks that the gap be reported.
- Re-litigating the batch template's base (`rfc-self-report-citations-and-batch-base`
  already proposes that fix). P1 touches the same text for a different reason
  and should be triaged alongside it.

## Adoption

P1, P2 and P3 are the blockers and are independent of each other; each is a
localized edit to one file. P3 additionally wants a line in the two prompt
templates.

P2 has a cheap partial: gitignoring the harness worktree directory is a
one-line repository change that can land ahead of any skill edit, and this run
already made it (`.gitignore`, commit `f696736`).

P4 is a self-contained paragraph in `SKILL.md`. P5, P6 and P7 are one-sentence
edits.

Nothing here blocks a run: this run drained fully with all three
improvisations. The cost of not adopting is that the next orchestrator invents
them again, differently, and that the improvisations are only as good as
whoever is holding the run.

## Resolution

**Partial triage 2026-08-17. This RFC is NOT retired** — six of its seven
proposals are recommended but not yet approved.

| Proposal | Outcome |
|---|---|
| P1 — state which worktree model wins under harness isolation | **adopted** 2026-08-17. Harness-provisioned wins: one worktree for every bead in the batch, do not create more. Refinement: the preamble also states the coarser in-flight signal (a surviving worktree means the BATCH is unfinished) and requires the agent to report each bead's close as it lands, since that is the orchestrator's only per-bead signal until the batch returns. The acceptance-check half landed as a refinement to `rfc-self-acceptance-signals-and-gate-baseline` P3. |
| P2 — give the harness worktree location a rule, and gitignore it | **adopted, trimmed to the rule.** The skill text now says the harness may place the worktree inside the repo and that the remedy is to gitignore its directory, and Phase 0 gained the check — create the first worktree, THEN confirm `git status --porcelain` is clean, because checking before it exists tests nothing. No repository change: `.claude/worktrees/` was already in this repo's `.gitignore:41`, and fixing our tree would have masked the defect for every install. |
| P3 — define baseline ownership when two beads share a gate | **reopened at review; not adopted as written.** Triage's own summary conflated the Phase 0 baseline MEASUREMENT with the per-bead gate RUNS, and review caught it. Measured: `{GATE_BASELINE}` reaches only the grooming/planning prompt (`references/subagent-prompts.md:70`); the per-bead `Gates:` line carries no baseline field, and step 4 says only "On gate failure, fix the root cause and rerun". The "no findings beyond the baseline" bar is stated in `plan-sprint` and in this repo's `AGENTS.md`, and nowhere in `bead-bonanza`'s implementer-facing text. So the defect is wider than two siblings sharing a baseline: **no implementing subagent is ever told the baseline at all.** Pending a resized proposal. |
| P4 — give the findings harvest a selector that matches an enumerated queue | **recommended adopt; not yet approved.** Confirmed live: `references/shared/findings-capture-protocol.md:93–94` still filters on a shared label, which a `bead-bonanza` queue does not have by construction. |
| P5 — scope the base-ref ancestor check to every worktree subagent | **recommended adopt; not yet approved.** Confirmed live at `references/subagent-prompts.md:285` (cited as `:280`; the line moved, the text did not). |
| P6 — reconcile a returned subagent against its terminal progress event | **adopted** as written, placed in Phase 3 step 4 after the no-return branch, with the Monitor's `committed\|blocked` filter cited as the reason the gap is invisible. |
| P7 — say what to do when a draft PR already exists | **recommended adopt; not yet approved.** Confirmed live: the Draft PR bullet at `SKILL.md:346` assumes the PR does not exist. |

**Dedupe — P6 and a sibling RFC.** P6 and
`rfc-self-wave-chaining-and-lost-returns` P3 close **opposite directions** of
one gap: P6 covers "returned, but no terminal event", P3 covers "never
returned". Neither subsumes the other, as both RFCs independently state. Both
were adopted in this pass and land adjacent to each other in the same block.

**Dedupe — the batch template.** This RFC asks (correctly) to be triaged
alongside `rfc-self-report-citations-and-batch-base`, whose P2 concerns the
same text. That P2 is now closed as already-implemented — the chaining fix
landed mid-run via bead `agent-plugins-61f.2` and is visible at
`references/subagent-prompts.md:178–179`. P1 here is unaffected: it is about
that block's *composition* with harness isolation, which is still live.

**Provenance check.** `bead-bonanza` v1.4 cited; **v1.5 current at triage**, so
every citation was re-verified. All seven anchors were still present. E2's
observation that the harness places worktrees inside the repository was not
re-measured — no run was in flight — but the gitignore entry corroborates that
someone hit it.

**Self-authorship, disclosed.** `origin: self`, `model: claude-opus-5`; this
session is configured for `claude-opus-5`. **The RFC was authored and triaged
by the same model.** P6's adoption ratifies an observation its own author made
about a run it was orchestrating.

**Files changed.** `plugin/skills/bead-bonanza/SKILL.md`.

**Version bumps.** `bead-bonanza` 1.5 → 1.6 (shared). Plugin 0.18.0 → 0.19.0.

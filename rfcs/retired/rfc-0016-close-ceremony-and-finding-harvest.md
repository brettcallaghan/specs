---
title: "RFC: The close ceremony commits shared tracker state, and the
  finding harvest it wires up returns zero by default"
description: bb-close-bead.sh exports a shared database into every bead
  commit so the orchestrator's pathspec guard cannot work; the prescribed
  retro harvest silently finds nothing; worktree-parallel is unusable in
  the backend the skill is built for
slug: rfc-0016
type: rfc
status: retired
updated_at: 2026-08-12T00:00:00Z
origin: self
skill: bead-bonanza
skill_version: "1.3"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "0.14.0"
---

# RFC: The close ceremony commits shared tracker state

## Summary

A ten-wave `bead-bonanza` run closed 23 beads with zero blocked and zero
skipped, so the skill's spine held. Three defects sit around that spine,
all in the seam between the orchestrator and its subagents. The close
ceremony serializes a **shared mutable database** into every bead commit,
which makes the skill's own anti-clobber guard structurally unable to
work. The finding-harvest the skill wires up returns **zero** under both
obvious query forms, so an aggregated retro can report confidently on an
empty set. And worktree-parallel — one of four dispatch modes — cannot be
used at all in the beads backend the skill exists to drive.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code on the web (remote execution environment) |
| Task the skill was applied to | Resuming an interrupted run over epic `agent-plugins-bge`: 11 remaining work orders plus 7 discovered mid-run, dispatched across 10 waves at haiku/sonnet/opus by bead tier, in a Go + docs repository |
| References / scripts / templates invoked | `bead-bonanza/SKILL.md` v1.3, `scripts/bb-close-bead.sh`, `scripts/bb-progress.sh`, `references/subagent-prompts.md`, `references/shared/work-order-definition-of-ready.md` v1.1, `rfcs/findings-capture.md` (vendored `findings-capture-protocol.md` v0.2), `groom-beads` v1.1 |
| Run dates | 2026-08-04 / 2026-08-05 |

The `model` property is withheld: this session runs under a harness
policy forbidding the model identifier in repository artifacts. The
placeholder follows the convention established by
`rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md`, the open
RFC covering that gate conflict. `effort` is `unknown` because the
harness does not expose it to the running agent.

## Evidence

1. **[script-defect · blocker] The close ceremony commits whatever is in
   the shared tracker database, so the orchestrator's pathspec guard
   cannot work.** `SKILL.md:787-796` states the guard: "the orchestrator
   commits its own work by explicit pathspec (`git commit -- <files>`) so
   a concurrently-staged tracker ledger rides with the next bead commit
   instead of being swept up mid-ceremony."

   That remedy addresses the git index. The race is upstream of it.
   `scripts/bb-close-bead.sh:64-72` runs a full `bd export`, regenerating
   the **entire** ledger from the Dolt database every agent in the tree
   shares, and line 85 then stages it unconditionally:

   ```sh
   before=""; [ -f .beads/issues.jsonl ] && before="$(git hash-object .beads/issues.jsonl)"
   if ! bd export > /tmp/bb-export.$$ 2>/dev/null; then
   ...
   [ -f .beads/issues.jsonl ]      && git add .beads/issues.jsonl
   ```

   So any tracker write the orchestrator makes while a subagent sits
   between `claimed` and `committed` is *inside the export that subagent
   commits* — regardless of what either party passes to `git add`.

   Observed: while the bge.12 subagent held the tree, the orchestrator
   baked a decision into `agent-plugins-tff` and wired two dependency
   edges (both tracker-only writes, no files touched). It then ran
   `git commit -- .beads/issues.jsonl .beads/interactions.jsonl` exactly
   as the guard prescribes, and got `nothing to commit, working tree
   clean` — because `bb-close-bead.sh` had already exported and committed
   those edits inside `39e0e6a agent-plugins-bge.12 P1 Build the
   four-subcommand command tree`. Confirmed by reading the committed
   ledger: `git show 39e0e6a:.beads/issues.jsonl` contains tff's
   `DECISION (2026-08-05)` section.

   Nothing was lost, but the skill's "One commit per bead, always"
   invariant (`SKILL.md` Operating principles) silently did not hold: that
   commit carries another actor's work. Repro: with a subagent between
   claim and close, run any `bd update` from the orchestrator, then let
   the subagent's close ceremony run.

2. **[script-defect · blocker] The prescribed finding harvest returns
   zero, by two independent causes.** `SKILL.md:797-802` wires findings
   capture — "record `retro:`-prefixed `bd comment`s on the bead being
   worked" — and hands aggregation to "the feature's final DoD bead". It
   names no harvest command. The vendored protocol
   (`rfcs/findings-capture.md`, Aggregation) supplies the only one:
   "`bd list --label <feature-slug>` then each issue's `retro:`
   comments."

   Both halves of that instruction fail silently on a completed run:

   ```
   $ bd list --label feature-spec-validation-cli --json | len   -> 16
   $ bd list --label feature-spec-validation-cli --status closed --json | len -> 20
   $ bd show agent-plugins-bge.8 --json   -> "comments": 0
   $ bd show agent-plugins-bge.8 | grep -cE '^\s+retro:'   -> 7
   ```

   `bd list` defaults to open issues, and a drained closure set is
   almost entirely closed. `bd show --json` omits comments entirely
   while the text output prints them. Following the protocol as written
   yields **0 findings** where **53 across 15 beads** exist. The orchestrator
   reported zero to the user before catching it.

   The failure is silent in the worst direction: an aggregated retro over
   an empty set produces a confident, clean-looking artifact. Note that
   `retro-skill/SKILL.md`'s own script-defect *example* already cites the
   `bd list` half of this ("returns ~nothing at aggregation time because
   it defaults to open issues") — it is known, illustrated in one skill,
   and still unfixed in the protocol that depends on it.

3. **[gap · blocker] Worktree-parallel has no exclusion for the beads
   backend.** `SKILL.md:433-441` defines the mode and gives exactly two
   skip conditions: "Skip this mode if there are <3 isolated beads
   (parallelism savings don't pay) or if the beads aren't actually
   isolated (you'll just hit merge conflicts)."

   At wave 5 the run had exactly three ready beads with disjoint
   `c4_changes` (`Numbering`, `Corpus`, `Profile Registration`) — the
   mode's stated trigger. It is nonetheless unusable here, for reasons
   neither condition covers: `.beads/embeddeddolt/` is gitignored and
   lives at the repo root, so a worktree cannot see the tracker database
   at all; and every agent commits `.beads/issues.jsonl`, which the
   project's own `CLAUDE.md` states has "no merge driver configured for
   it, so beads created on two branches will conflict textually on
   merge." Three worktrees would have produced a three-way conflict in a
   generated file, with no regen step able to settle it (the skill's
   fixture-conflict remedy at `SKILL.md:436-438` does not apply — the
   ledger is a projection of per-worktree databases that do not exist).

   The agent reasoned its way to skipping and said so, but the skill
   pointed the other way on its own criteria.

4. **[gap · snag] No rule for how much over-routing a mixed-tier bucket
   may cost.** `SKILL.md:475-478`: "prefer bucketing beads that share a
   tier. When a bucket mixes tiers, run the whole batch subagent at the
   **highest** tier among its members — never run a bead below its tier.
   If honoring that would drag several cheap beads up, split the bucket."

   "Several" is the whole rule and it is undefined. The run hit the
   decision twice with opposite outcomes and no principle to appeal to:
   bge.9 (`light`) was batched with a `standard` bead and ran one tier
   above its assignment; bge.8 (`advanced`) and bge.11 (`standard`) were
   split rather than batched. Both calls were defensible, neither was
   derived from the skill. The deciding factor turned out to be gate-suite
   cost — seconds in this repo, so batching bought nothing — which the
   rule never mentions even though saving full test runs is the stated
   reason batching exists at all.

## Proposals

### P1 — Replace the guard that cannot work [blocker]

**Problem.** `SKILL.md:787-796` prescribes explicit pathspec as the
defence against a bead commit absorbing the orchestrator's work. That
defence is aimed at the git index, but `bb-close-bead.sh:64-72` rebuilds
the ledger from a database both parties write, so the orchestrator's
edits are already inside the artifact before any `git add` runs
(finding 1). No pathspec discipline on either side can prevent it.

**Change.** Two parts, both needed.

Replace the second sentence of the guard at `SKILL.md:789-792` with:

> The tracker ledger is a projection of a database every agent shares,
> not a file one agent owns. `bb-close-bead.sh` regenerates it in full at
> each close, so any tracker write the orchestrator makes while a
> subagent is between `claimed` and `committed` lands in that subagent's
> commit — explicit pathspec cannot prevent this, because the capture
> happens before the index is consulted. WHILE any subagent holds a
> claim, the orchestrator SHALL confine itself to read-only `bd` queries
> and defer its own tracker writes (grooming a discovered bead, baking a
> decision, wiring an edge) until the wave returns. WHERE a write cannot
> wait, make it and note in the report which bead commit will carry it.

Add to `bb-close-bead.sh`'s commit message, or its stdout, a line naming
how many issue records the export changed, so a close that carried
foreign edits is visible in review rather than silent.

**Evidence.** Finding 1. The next run either defers its tracker writes or
records where they landed, and "one commit per bead" stops being quietly
false.

### P2 — Give the harvest a command that works [blocker]

**Problem.** The only harvest instruction is in the vendored protocol and
is wrong twice over: `bd list` hides closed beads and `bd show --json`
hides comments (finding 2). `SKILL.md:797-802` adds none of its own. An
aggregated retro over the resulting empty set looks identical to a clean
run.

**Change.** In `SKILL.md`'s findings-capture principle, state the harvest
explicitly rather than delegating it:

> Harvest with `bd list --label <slug> --status closed` **and** the
> default open query — a drained closure set is mostly closed, and `bd
> list` shows open issues only. Read comments from `bd show <id>` TEXT
> output: `--json` omits them. Assert the count is non-zero before
> aggregating; a harvest returning zero findings across a completed
> feature is a broken query, not a clean run, and an aggregated retro
> built on it is worse than none.

Propose the same correction to `findings-capture-protocol.md`'s
Aggregation section (owned by `retro-spec`), since that is where the
wrong command is published and it reaches every project that vendors it.

**Evidence.** Finding 2. The final DoD bead in this run received the
corrected mechanism in its dispatch prompt and harvested exactly 53
findings, matching the independently counted total — the same bead would
otherwise have aggregated nothing.

### P3 — Exclude worktree-parallel in the beads backend [blocker]

**Problem.** The mode's two skip conditions (`SKILL.md:438-441`) are
about bead count and surface overlap. Neither covers a tracker whose
database is gitignored and whose exported ledger has no merge driver —
which is the beads backend this skill is built for (finding 3).

**Change.** Add a third skip condition at `SKILL.md:441`:

> Skip this mode WHERE the tracker's state cannot be forked with the
> worktree. In the beads backend this is the normal case: the database is
> gitignored and lives at the repo root, so a worktree cannot see it, and
> every agent commits the exported ledger, which typically has no merge
> driver — N worktrees produce an N-way conflict in a generated file that
> no regen step can settle. Worktree-parallel therefore applies only
> WHERE the tracker is external to the repository, or the run is
> configured so subagents make no tracker writes.

**Evidence.** Finding 3. The next run reaches the same decision from the
skill rather than despite it, and a run that *is* correctly configured for
worktrees still gets the mode.

### P4 — Make the mixed-tier batching rule decidable [snag]

**Problem.** `SKILL.md:475-478` resolves mixed-tier buckets with "if
honoring that would drag several cheap beads up, split the bucket."
"Several" is undefined, and the rule omits the quantity that actually
decides it — what a full gate run costs (finding 4).

**Change.** Replace the final sentence at `SKILL.md:478` with:

> Split the bucket WHEN the number of beads dragged above their tier
> exceeds one, OR when any bead would be dragged more than one tier.
> Weigh it against what batching saves: batching exists to avoid repeated
> full `<test>` runs, so WHERE the gate suite completes in seconds the
> saving is negligible and the correct tier wins; WHERE it takes minutes,
> one bead one tier high is the cheaper error. Record every deliberate
> over-route in the Phase 4 report alongside escalations — it is the only
> evidence the run produces about over-routing, which otherwise succeeds
> silently and is paid for on every run.

**Evidence.** Finding 4. The two calls this run made by judgment become
derivable, and the Phase 4 report gains the over-routing signal its own
escalation section says is missing.

## Considered and Not Proposed

- **The two finding classes routed here by the implementation retro.**
  `rfc-0011` routes "correcting the slug prefix registry" and the
  generated-model-pinning finding to `retro-skill`. Neither is a
  `bead-bonanza` finding: the first concerns
  `template-authoring-standards.md`, a vendored document the `spec` skill
  owns (`rfc-0011` says so itself — "That file is a **vendored copy of a
  framework document**"), and the second concerns the feature blueprint
  template. Folding them here would put one RFC across three skills, the
  anti-pattern `retro-skill/SKILL.md` names explicitly. They need their
  own RFC against `spec`, from an agent with that phase's evidence.
- **The `bd show --json` omission as a `bd` bug.** It may well be one,
  but `bd` is a third-party tool outside this framework's control, and
  the fix that protects the pipeline is P2's corrected instruction plus
  the non-zero assertion. Filing upstream is not this RFC's business.
- **The orchestrator's own errors this run** — a wrong dependency
  direction (bge.16 blocking the issues it was meant to close, forcing
  `--force`), a bead Design asserting an issue did not exist when it did,
  and a shell-quoting mistake that command-substituted backticks in retro
  prose. The skill warned against none of these because none is a skill
  defect; they are recorded in the run report, not proposed here.
- **A `git status` check before the orchestrator commits.** Considered as
  a lighter fix for finding 1 and rejected: the run *did* check, and the
  window between the check and the commit is exactly where the close
  ceremony landed. Ordering discipline cannot close a race on shared
  state; P1's deferral rule can.

## Non-Goals

- No change to the per-bead ceremony's ordering (claim → implement →
  gates → retro → atomic close → single commit). It held for 23
  consecutive beads and is the reason zero work was lost to the one
  container restart this session.
- No change to the tier vocabulary or the escalation ladder. P4 makes an
  existing rule decidable; it does not re-tier anything.
- Not proposing that subagents stop writing to the tracker. Their
  claims and retro comments are the mechanism; P1 constrains the
  orchestrator, which is the party that can wait.

## Adoption

P1 and P2 are independent of each other and of the rest, and both are
text changes to `SKILL.md` plus one downstream proposal each (a stdout
line in `bb-close-bead.sh` for P1, a correction to
`findings-capture-protocol.md` for P2). P2's protocol half should travel
to `retro-spec` as its own proposal, since that skill owns the file and
every project vendoring it inherits the broken command.

P3 and P4 are single-paragraph edits to `SKILL.md` at lines 441 and 478
respectively, independent of everything.

Backward compatibility: all four are additive to guidance and change no
artifact format, no bead field, and no script interface. P1 constrains
orchestrator timing, which may make a wave marginally slower when a
discovered bead needs grooming mid-flight — that is the intended trade,
and P1's WHERE clause keeps the escape hatch. P3 removes a dispatch mode
in the beads backend, where finding 3 shows it never worked.

Overall effect: the two places this run produced a silently wrong result
— a bead commit carrying another actor's work, and a harvest that reports
nothing while 53 findings exist — become impossible to reach without
noticing.

## Resolution

**Triaged 2026-08-10; implemented across two `bead-bonanza` sprints and
resolved 2026-08-12.** 4 proposals: 3 adopted (1 with refinement, 1 since
retired), 1 rejected. 3 triage-originated changes are recorded below the
proposal table.

More than five proposals were approved across the three `bead-bonanza`
RFCs triaged together and most of them changed one file, so step 5's
"Sequence before you edit" branch applied: every approved proposal was
filed as a bead under epic `agent-plugins-92q` and run through
`groom-beads`, and the implementation is the epic rather than an inline
edit. The second sprint then changed premises several of these proposals
were written against; where that happened it is stated in the row rather
than glossed.

| Item | Outcome | Bead |
|------|---------|------|
| P1 — replace the guard that cannot work | **adopted, landed, and since retired.** The deferral rule and an export-delta counter landed as proposed (`d16abbc`). The counter over-reported by one, tripping its "another actor wrote" threshold on every close, and was corrected. Both halves are now gone: `agent-plugins-92q.31` moved the tracker onto `origin refs/dolt/data` and untracked `.beads/*.jsonl`, so the close ceremony stages code only and no export exists in the commit path to absorb anyone's tracker write. The finding was real; its remedy was scoped to a tracker model this repository has since abandoned. The current rule is the inverse — `references/operating-principles.md`, "Tracker writes are never deferred". | `agent-plugins-92q.17`, `agent-plugins-92q.29`; retired by `agent-plugins-92q.31` |
| P2 — give the harvest a command that works | **adopted with refinement**, in three places the proposal did not anticipate. (a) The corrected harvest is `bd export \| jq`, not P2's `bd list --status closed` — `bd list` was dropped from the harvest entirely rather than repaired, because a second query form is a second thing to get wrong. (b) The correction lives in `findings-capture-protocol.md`, the single statement of the channel, vendored into the `references/shared/` of `retro-spec`, `retro-skill` and `groom-feature-blueprint`; `SKILL.md` carries the one-line rule and points there. (c) The per-project vendored copy P2's evidence cites, `rfcs/findings-capture.md`, was deleted — the protocol is no longer copied into target projects at all. The non-zero assertion before aggregating landed as written. Separately, findings moved off the close artifact: `retro:` comments own findings and the close file became `<bead-id>.handoff.md`. P2's downstream half travelled as its own RFC to the skill that owns the protocol, `rfcs/outbox/retro-spec/rfc-agent-plugins-findings-capture-harvest.md`. | `agent-plugins-92q.18`, `agent-plugins-92q.22`, `agent-plugins-92q.28`, `agent-plugins-y1h` |
| P3 — exclude worktree-parallel in the beads backend | **rejected as written.** See *On the rejection* below. Replaced by `agent-plugins-92q.9`, which states the git-side constraints measurement supports. The mode P3 sought to exclude has since ceased to exist as a mode: `agent-plugins-92q.30` made one worktree per bead universal and not opt-in, so there are no skip conditions left to add a third to. | rejected; replaced by `agent-plugins-92q.9` |
| P4 — make the mixed-tier batching rule decidable | **adopted** as written — split WHEN more than one bead is dragged above its tier OR any bead is dragged more than one tier, weighed against gate-suite cost, with deliberate over-routes recorded in the Phase 4 report. Refinement in mechanism only: the gate-cost half is keyed to Phase 0's recorded gate baseline (`pre-flight-check.sh --gate-baseline`) rather than to an unmeasured impression of what the suite costs. The rule survives one-worktree-per-bead — batching remains a dispatch mode, with a batch subagent holding exactly one worktree at a time. | `agent-plugins-92q.10` |

**Triage-originated changes.** These came from the review session rather
than from this RFC's testimony, and are numbered `T` so they are not read
as proposals it made.

| Item | Outcome | Bead |
|------|---------|------|
| T1 — bucket beads on C4 blast radius, not prose-scanned file paths, and treat impacts-intersects-changes as a sequencing constraint | **adopted.** Under one worktree per bead the hazard it guards became merge conflict rather than working-tree collision; the rule holds, its stated reason was corrected. | `agent-plugins-92q.8` |
| T2 — add a sprint retro to Phase 4 | **adopted.** Closed 2026-08-12 on the sprint branch, after this worktree's base commit; its text is not in the tree this Resolution was verified against. | `agent-plugins-92q.19` |
| T3 — RFC-0004's deferral bead `agent-plugins-1a9` was never persisted; correct the record and audit for other lost deferrals | **adopted.** The corrected record is in RFC 0004's Resolution and in `CLAUDE.md`. | `agent-plugins-92q.20` |

**On the rejection.** P3's premise (`.beads/embeddeddolt/` is gitignored)
is true; the conclusion drawn from it — "a worktree cannot see the tracker
database at all" — does not hold and was never tested before this RFC
asserted it. `agent-plugins-t09` measured `bd` 1.1.2 from a worktree with
`BEADS_DIR` unset: `bd where` and `bd worktree info` resolve the MAIN
checkout's database through git's common directory, `bd list` returns the
identical issue set, and no `embeddeddolt/` or redirect file is created in
the worktree. `agent-plugins-jh6` independently reproduced this inside
this repository. Tracker access was never the obstacle. The replacement
bead states what measurement does support: the merge-conflict hazard in
the exported ledger, and the requirement that `worktree.baseRef` be `head`
so a worktree branches from session HEAD rather than the default branch.
A rejection is recorded so the idea is not relitigated without new
evidence.

**Self-authorship, disclosed.** This RFC carries `origin: self`: it was
authored inside this repository by an agent running the skill it
critiques, and triaged inside the same repository. Its `model` field is
withheld under the harness policy tracked by
`rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md`, so
whether the authoring agent and the triaging agent were the same model
**cannot be confirmed or excluded**. Read every adoption here as this
repository approving its own proposal, not as independent corroboration.
The one place that materially changes the reading is P1, whose deferral
rule ratifies a constraint its own author was operating under. The
rejection of P3 is the exception that cuts the other way: it was refuted
by measurements taken in `agent-plugins-t09` and `agent-plugins-jh6`,
beads run separately from the session that authored this RFC.

**Provenance check.** `skill: bead-bonanza` at `skill_version: "1.3"` and
`plugin_version: "0.14.0"` — both current when the RFC was filed and when
triage began. Every cited line was re-read against the tree before
classification; none had moved or been fixed since.

**Dedupe.** No match in RFCs 0001–0015. RFC 0004 (`bead-bonanza` dispatch
and script contracts) is the nearest prior item and overlaps at no
proposal; T3 corrects a bookkeeping failure in its Resolution rather than
revisiting its decisions.

**Files changed.** `plugin/skills/bead-bonanza/SKILL.md`;
`plugin/skills/bead-bonanza/references/operating-principles.md`,
`dispatch-plan.md`, `subagent-prompts.md`;
`plugin/skills/bead-bonanza/scripts/bb-close-bead.sh`;
`findings-capture-protocol.md` in `plugin/skills/retro-spec/references/`
and its vendored copies under `retro-skill` and `groom-feature-blueprint`;
`rfcs/findings-capture.md` (deleted);
`rfcs/outbox/retro-spec/rfc-agent-plugins-findings-capture-harvest.md`
(new); `CLAUDE.md`;
`rfcs/retired/rfc-0004-bead-bonanza-dispatch-and-script-contracts.md`.

**Version bumps.** `findings-capture-protocol` 0.2→0.4. Plugin
0.16.0→0.17.0, one MINOR bump covering this RFC and its two siblings.

The `version` frontmatter of the changed shipped **skills** has been
bumped in `agent-plugins-x8r`: `bead-bonanza` 1.3→1.4 and `spec` 1.8→1.9.
The other four skills (`groom-beads`, `groom-feature-blueprint`,
`retro-skill`, `retro-spec`) did not change in this sprint and received
no bumps.

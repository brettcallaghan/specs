---
title: "RFC: The findings-capture protocol's Aggregation harvest returns
  zero on a completed run"
description: bd list defaults to open issues and bd show --json omits
  comments, so the protocol's prescribed harvest silently returns
  nothing against a drained closure set; bd export | jq fixes both.
slug: rfc-self-findings-capture-harvest
type: rfc
status: ready-for-review
updated_at: 2026-08-11T00:00:00Z
origin: agent-plugins
skill: retro-spec
skill_version: "1.3.0"
model: "claude-sonnet-5"
effort: unknown
runs_observed: 2
severity: blocker
plugin_version: "n/a (in-repo dogfooding)"
---

# RFC: The findings-capture protocol's Aggregation harvest returns zero

## Summary

`findings-capture-protocol.md`'s Aggregation section prescribes a
two-command harvest — `bd list --label <slug>` then each issue's
`retro:` comments via `bd show --json` — for pulling every recorded
finding across a feature's closure set. Both commands fail silently on
exactly the run shape the section exists for: a drained set of mostly
closed beads. `bd list` defaults to open issues only, and `bd show
--json` omits comments entirely even though the text output prints
them. The fix is a single `bd export | jq` pipeline plus a required
non-zero assertion before aggregating.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code CLI, orchestrated via the `bead-bonanza` skill |
| Task the skill was applied to | Implementing `agent-plugins-92q.18`, a bead that itself exists to fix this defect in the consuming skill (`bead-bonanza`); verifying the protocol document `retro-spec` owns carries the same defect |
| References / scripts / templates invoked | `plugin/skills/retro-spec/references/findings-capture-protocol.md` v0.2 (as run), `bd` 1.1.2 |
| Run dates | 2026-08-04 / 2026-08-05 (prior run, cited in `agent-plugins-92q.18`'s Description, 53 findings across 15 beads), 2026-08-11 (this run) |

## Evidence

1. [script-defect · blocker] `bd list --label <slug>` defaults to open
   issues. Measured 2026-08-04/05 against a drained feature closure set:
   `bd list --label feature-spec-validation-cli --json` returned 16
   issues; `bd list --label feature-spec-validation-cli --status closed
   --json` returned a disjoint 20. The set was almost entirely closed,
   so the default query missed most of it.
2. [script-defect · blocker] `bd show <id> --json` omits comments.
   Measured the same run: `bd show agent-plugins-bge.8 --json | jq
   '.comments | length'` → `0`; `bd show agent-plugins-bge.8 | grep -cE
   '^\s+retro:'` → `7`. The text output carries what the JSON output
   drops.
3. [script-defect · blocker] Re-measured 2026-08-11 against this
   project's own drained set (`bead-bonanza-hardening` label, 18 closed
   beads): `bd show <id> --json | jq 'has("comments")'` → `false` for
   every id, confirming finding 2 still holds on `bd` 1.1.2.
   `bd export | jq 'select(.id=="agent-plugins-92q.9") | .comments |
   length'` → `1`, confirming `bd export` carries comments as a
   `comments` array on the issue record — the harvest is JSON-native
   after all, so the text-output-parsing fallback the two-command
   instruction implies is unnecessary and itself fragile (a bead's
   Description/Design prose can independently contain the substring
   `retro:`, which a `grep '^\s*retro:' <(bd show <id>)` over the whole
   text output would then over-count; verified: `agent-plugins-34g`
   scored 3 via `.comments[]` filtering vs 4 via whole-output grep,
   because its Description quotes another skill's `retro:` prose).
4. [script-defect · blocker] The corrected harvest was run end to end
   against the same 2026-08-11 closure set and its count checked against
   an independent count restricted to each bead's `COMMENTS` section
   (not its whole `bd show` output, per finding 3):
   `bd export | jq -r 'select(.status=="closed") | select((.labels//[])
   | index("bead-bonanza-hardening")) | .comments[]? |
   select(.text|startswith("retro:")) | .text' | wc -l` → `44`;
   independent count → `44`. Matching. The 2026-08-04/05 run reported
   the same corrected mechanism harvesting exactly 53 findings against
   an independently counted total of 53.

## Proposals

### P1 — Replace the Aggregation section's harvest command [blocker]

**Problem.** `findings-capture-protocol.md`'s Aggregation section (as
run, v0.2) instructs: "`bd list --label <feature-slug>` then each
issue's `retro:` comments" with no further mechanism named. Every
project that vendors this document inherits a harvest that returns zero
against its own intended use case (findings 1-3).

**Change.** Replace the Aggregation section's harvest instruction with:

> It harvests the findings across the feature's closure set with one
> `bd export | jq` pipeline, never `bd list` plus `bd show --json`: `bd
> list` defaults to open issues, and a drained closure set is almost
> entirely closed; `bd show --json` omits comments entirely even though
> the text output prints them. `bd export` carries comments as a
> `comments` array on the issue record and covers every status in one
> pass:
>
> ```
> bd export | jq -r 'select((.labels//[]) | index("<feature-slug>"))
>   | .comments[]? | select(.text|startswith("retro:")) | .text'
> ```
>
> Assert the count is non-zero before aggregating — a harvest that
> returns zero findings across a completed feature is a broken query,
> not a clean run, and an aggregated retro built on it is worse than
> none.

**Evidence.** Findings 1-4. The corrected command was measured working
on two independent closure sets across two projects (53/53 and 44/44),
so this is not a proposed-but-untested fix.

### P2 — Make the non-zero assertion an explicit step [snag]

**Problem.** Even with the corrected command, a harvest bug that
returns zero is indistinguishable from a genuinely clean feature unless
the aggregating agent is told to treat zero as suspicious and check.

**Change.** State the assertion as an explicit step with a stated
failure action (stop and re-run the query — check label spelling and
that the tracker's export reflects the latest closes — do not aggregate
on a zero count), not as background advice folded into the harvest
description. P1's replacement text above already carries this; this
proposal is to keep it a stated step if the section is otherwise
reworded during triage.

**Evidence.** Finding 4's corroborating case: the original run's
orchestrator "reported zero to the user before catching it" per
`agent-plugins-92q.18`'s Description — the assertion is what would have
caught it before the report went out, not after.

## Considered and Not Proposed

- **Fixing `bd show --json`'s comment omission as a `bd` bug report to
  the `bd` project.** May well be a real defect, but `bd` is a
  third-party tool outside this framework's control; the harvest fix
  that protects every consuming project is P1's corrected instruction,
  which does not depend on `bd show --json` changing.
- **Adding a `--status` accumulation warning to the protocol text.**
  Considered folding in "repeating `--status` overwrites rather than
  accumulates" as a caveat about the old `bd list` approach, but P1
  removes the `--status` flag from the prescribed command entirely
  (`bd export` covers every status in one pass), so the caveat would
  apply only to an approach the new text no longer recommends.

## Non-Goals

- No change to the finding format, the "when to record" trigger list,
  or the "where to record" section (`bd comment` prefixed `retro:`, or
  the task-file backend's `## Retro findings` section). Those are
  unaffected by the harvest defect.
- Not proposing a version of the protocol that drops the beads backend
  or the task-file backend — both remain, only the beads backend's
  harvest command changes.

## Adoption

P1 is the whole fix; P2 keeps its assertion language explicit through
triage. Both are text-only changes to
`findings-capture-protocol.md`'s Aggregation section — no field, no
artifact format, and no script interface changes. Every project
vendoring this document (via `scripts/sync-shared-references.sh` or an
equivalent copy step) picks up the fix on its next sync; no vendored
copy needs a structural change, only a content refresh.

Backward compatibility: fully additive. A project already using `bd
export | jq` (because it independently discovered the same defect, as
`agent-plugins-92q.18` did) sees no behavior change; a project still on
the old two-command instruction starts harvesting what it already
should have been harvesting.

Overall effect: the harvest returns what exists instead of reporting a
clean run over an empty set, for every project that vendors this
protocol — not just the one that found the defect.

## Resolution


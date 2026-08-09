---
title: "RFC: The work-order to bead mapping table names a field bd does
  not expose, so verifying acceptance criteria reports a false negative"
description: Correct the mapping table's acceptance row to name the
  field bd actually stores and returns, and distinguish the write flag
  from the stored field throughout
slug: rfc-self-acceptance-field-name
type: rfc
status: ready-for-review
updated_at: 2026-08-07T00:00:00Z
origin: self
skill: groom-beads
skill_version: "1.1"
model: "<withheld by session policy — see Provenance>"
effort: unknown
runs_observed: 1
severity: polish
plugin_version: "0.15.0"
---

# RFC: The mapping table's acceptance field name does not match bd

## Summary

`groom-beads` v1.1 was run as the finishing pass over ten work orders
decomposed from [[feature-spec-reference-integrity]]. The pass itself
went cleanly — `bd-health.sh` behaved exactly as documented, and the
`--metadata` versus `--set-metadata` distinction RFC-0003 P4 added was
load-bearing and correct. One thing misled: the work-order-to-bead
mapping table names the acceptance field `acceptance`, which is the
*write flag*, while `bd` stores and returns it as `acceptance_criteria`.
Checking the field by its documented name reports a false negative on a
bead whose acceptance criteria are present.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | agent-plugins (self) |
| Skill critiqued | `groom-beads` v1.1 |
| Task | Dispatch-readiness finishing pass over work orders `agent-plugins-9kb.1`–`.10`, invoked by `groom-feature-blueprint` v1.2 Phase 3 |
| `bd` version | 1.1.2 (20e493e56), embedded Dolt mode |
| References invoked | work-order-definition-of-ready v1.2; work-order-field-guide v1.2; work-order-template v1.3; `scripts/bd-health.sh` |
| Harness | Claude Code on the web, remote execution container |
| Model | Withheld: this session's harness supplies the model identifier under a policy forbidding it in any artifact pushed to a repository. Same handling as RFC-0010; tracked by `rfcs/inbox/retro-skill/rfc-self-retro-skill-withheld-model.md` |
| Runs observed | 1 |
| Run date | 2026-08-07 |
| Plugin version | 0.15.0 |

## Evidence

1. **[misleading-text, polish]** `plugin/skills/groom-beads/SKILL.md:70`
   maps the work order's Acceptance section onto a bead field:

   > \| `## Acceptance` \| `acceptance` \|

   and `SKILL.md:149` repeats it:

   > Acceptance become the bead's `description`, `design`, and
   > `acceptance`.

   `acceptance` is the name of the **write flag** (`bd create
   --acceptance`, confirmed in `bd create --help`), not of the stored
   field. After creating `agent-plugins-9kb.1` with
   `--acceptance "$(cat …)"`, the agent verified the result with:

   ```
   bd show agent-plugins-9kb.1 --json | python3 -c "… print('has
   acceptance:', bool(i.get('acceptance')))"
   ```

   Observed: `has acceptance: False`. Expected: true — the criteria were
   written and are present. The agent concluded the flag had silently
   failed and re-investigated, dumping the record's keys:

   ```
   ['acceptance_criteria', 'comment_count', 'created_at', 'created_by',
    'dependencies', 'dependency_count', 'dependent_count',
    'description', 'design', 'id', 'issue_type', 'labels', 'owner',
    'parent', 'priority', 'status', 'title', 'updated_at']
   ```

   The field is `acceptance_criteria`; the content was there all along.
   Minimal repro: `bd create "probe" --acceptance "- [ ] x" --silent`,
   then `bd show <id> --json | jq '.acceptance, .acceptance_criteria'` —
   the first is `null`, the second holds the text.

   `SKILL.md:0`'s standing instruction to probe the CLI does not catch
   this: probing `bd create --help` shows `--acceptance` and confirms
   the table, because the flag really is spelled that way. Only reading
   back a created bead exposes the difference.

## Proposals

### P1 — Name the stored field, not the write flag [polish]

**Shape.** Misleading text — `SKILL.md:70` and `SKILL.md:149` name a
field `bd` does not expose on read.

**Problem.** The mapping table is the skill's authority on where a work
order's sections live once materialized, and phase 6 asks the agent to
validate rather than eyeball the result. Validating acceptance coverage
by the documented name returns a false negative on every bead, which
either sends the agent re-investigating (this run) or, worse, prompts a
"fix" that rewrites criteria already present. The `description` and
`design` rows are correct, which makes the one wrong row harder to
suspect.

**Change.** At `SKILL.md:70`, replace the row with one that names both
spellings:

> \| `## Acceptance` \| `acceptance_criteria` (written with
> `--acceptance`) \|

and at `SKILL.md:149`, replace "and `acceptance`" with:

> and `acceptance_criteria` — note the asymmetry: the write flag is
> `--acceptance`, the stored and returned field is
> `acceptance_criteria`. Verify a materialized bead by the field name,
> not the flag name; `bd show <id> --json` returns the former.

Confirm both spellings against the installed `bd` at implementation
time, per this skill's own probe-the-CLI rule — this RFC observed
`bd` 1.1.2 in embedded mode only.

**Evidence.** Finding 1.

**Backward compatibility.** Documentation-only, and it corrects a name
rather than changing a behaviour. No command in the skill changes:
`--acceptance` remains the flag every `bd create` example uses. No
existing bead is affected. `groom-beads` would go 1.1 → 1.2.

## Considered and Not Proposed

- **`bd-health.sh` exiting 0 after running no checks.** Raised as a
  candidate script defect during this session and **retracted on
  verification**: the observation came from `echo "EXIT=$?"` after a
  pipeline, which reports `tail`'s status rather than the script's.
  Measured directly, the script exits **2** from a directory with no
  `.beads/`, matching `SKILL.md:279` ("2 when no check could run at
  all") and `bd-health.sh:26-27`. From the repository root it ran all
  five checks and reported `swarm validate` clean over 2 epics, no
  orphans, and a `lint` exit 1 naming only pre-existing beads. The
  script behaved correctly throughout; the error was the agent's
  measurement, and the fix for that is not a change to the script.
- **`bd` losing its workspace when the shell's cwd moves.** Two commands
  failed with "no beads database found" after a `cd` to a scratch
  directory outside the repository; `export BEADS_DIR=<repo>/.beads`
  resolved it and the error message named that remedy itself.
  `SKILL.md:0` already tells the agent to locate the beads directory
  before starting. Agent-side cwd management, not a skill gap.
- **`bd create --parent` not inheriting the parent's labels
  observably.** The skill notes a `--no-inherit-labels` flag exists, so
  inheritance is the default; the agent passed labels explicitly
  anyway and cannot say from this run whether inheritance fired. No
  evidence either way, so no finding.
- **The tier rubric and the C4 phase.** Both worked as written. The
  `--metadata` versus `--set-metadata` guidance added by RFC-0003 P4
  was directly load-bearing here — the `c4_changes` / `c4_impacts` /
  `c4_docs` arrays stored as real JSON arrays on all ten beads because
  the skill said to use `--metadata` and to restate the tier keys.
  Recorded as a positive, not a proposal.

## Non-Goals

- No change to `bd-health.sh`, which behaved correctly — see Considered
  and Not Proposed.
- No change to the mapping of `## Problem` or `## Design`; both name
  their stored fields correctly (`description`, `design`, verified in
  the same key dump).
- No change to the Definition of Ready or the tier rubric, both of which
  held up across ten work orders.

## Adoption

One proposal, one file
(`plugin/skills/groom-beads/SKILL.md`), two lines. No script, template,
or reference changes; nothing to migrate.

Worth landing opportunistically rather than urgently: the defect costs a
verification round-trip, and the run still reached a correct result. Its
value is that phase 6 tells the agent to validate rather than eyeball,
and the table is what the agent validates against — so a wrong field
name is a defect in the check itself, which is the class RFC-0011 and
RFC-0003 P1 both concern.

Expected effect: an agent verifying materialized work orders reads the
correct field the first time, and does not mistake present acceptance
criteria for missing ones.

## Resolution

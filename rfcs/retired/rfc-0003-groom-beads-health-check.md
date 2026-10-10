---
title: "RFC: bd-health.sh fails a clean backlog, and blocked beads stay
  in the queue"
description: The phase 6 health check reports hard graph problems on a
  backlog with zero cycles and zero orphans, and phase 2's blocked-bead
  flag does not remove the bead from the ready queue
slug: rfc-0003-groom-beads-health-check
type: rfc
status: retired
origin: self
skill: groom-beads
skill_version: "1.0"
model: claude-opus-5
effort: high
runs_observed: 1
severity: blocker
plugin_version: "0.7.0"
---

# RFC: bd-health.sh fails a clean backlog, and blocked beads stay in the queue

## Summary

`groom-beads` v1.0 was run to groom four `rfc-triage` beads to the
Definition of Ready. Phase 6 prescribes `scripts/bd-health.sh` as the
validation gate and instructs the agent not to declare the set ready
while it exits non-zero — but the script exited 1 on a backlog with
demonstrably zero cycles and zero orphans, for two reasons that are both
in the checker rather than the data. Separately, phase 2's instruction
for a bead that cannot reach Ready sets a metadata field that `bd ready`
does not read, so a correctly-flagged bead stays in the dispatch queue.
The thesis: the skill's two enforcement points — the validation gate and
the blocked-bead flag — are the two places it does not actually bind.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code, remote (web) session |
| Task the skill was applied to | Grooming four `rfc-triage` beads filed by a prior `triage-rfc` pass, to the Definition of Ready |
| References / scripts / templates invoked | `groom-beads/SKILL.md` v1.0; `references/shared/work-order-definition-of-ready.md` v1.1; `references/shared/work-order-field-guide.md` v1.1; `references/shared/work-order-template.md` v1.2; `references/bead.schema.json` (unversioned); `scripts/bd-health.sh` (unversioned); `scripts/validate_beads.py` (unversioned) — all under `plugin/skills/groom-beads/`. Tracker: `bd` 1.1.2 (dev), embedded Dolt mode |
| Run dates | 2026-07-30 |

Note: `plugin/skills/groom-beads/` and `.claude/skills/groom-beads/` were
byte-identical at run time (`diff -q` clean), so every line number below
holds in both copies. The canonical source is the `plugin/` path.

## Evidence

1. **[script-defect · blocker]** `bash
   plugin/skills/groom-beads/scripts/bd-health.sh` — observed exit 1 and
   the summary "RESULT: hard problems found — fix cycles/orphans before
   dispatching waves.", with `✗ orphans` printed directly above that
   check's own output, "✓ No orphaned issues found". Expected `✓ orphans
   clean` and exit 0. The cause is the success test at
   `plugin/skills/groom-beads/scripts/bd-health.sh:55`:

       if [ $rc -eq 0 ] && [ -z "${out//[[:space:]]/}" ]; then

   A check passes only when it produces **no output at all**. In `bd`
   1.1.2 the check succeeds *and* prints a success line, so the passing
   case is scored as a failure. Because `orphans` is registered `hard`
   at line 68, this sets `hard_fail=1` and the script exits 1.
   Minimal repro, in a beads repo with no orphaned issues:

       $ bd orphans; echo "rc=$?"
       ✓ No orphaned issues found
       rc=0
       $ bash plugin/skills/groom-beads/scripts/bd-health.sh; echo $?
       ...  ✗ orphans
             ✓ No orphaned issues found
       RESULT: hard problems found ...
       1

   The same inversion hit `lint` and `stale` (both registered `soft`, so
   display-only) and `doctor`. In this run **all five checks** printed
   `✗`, including three that had genuinely passed.

2. **[script-defect · blocker]** Line 67 of the same script,
   `plugin/skills/groom-beads/scripts/bd-health.sh`, invokes the epic
   validator with no argument:

       run_check "swarm validate" "swarm validate" hard   # cycles / orphans / disconnected per epic

   Observed: `bd swarm validate` → `Error: accepts 1 arg(s), received 0`,
   rc 1. Expected: either a per-epic invocation, or the check reported as
   unavailable. `bd swarm validate --help` states it validates "an epic's
   structure", so an epic ID is required. Repro:

       $ bd swarm validate; echo "rc=$?"
       Error: accepts 1 arg(s), received 0
       rc=1

   Two consequences. First, the availability probe at line 39 —
   `have() { "$BD" $1 --help >/dev/null 2>&1; }` — succeeds, because
   `--help` works. So the check never reaches the `skipped_hard` branch
   at lines 44-49 that exists precisely to report an unestablished
   guarantee ("A skipped hard check is a gap in coverage, not a pass"),
   and the summary at lines 84-88 that would have said so never fires.
   The check is scored as a graph *failure* instead of as *unverified*.
   Second, the repo under grooming contained zero epics (`bd list --type
   epic --status all` → "No issues found"), so no valid invocation of
   this check existed at all.

   Combined with finding 1, `plugin/skills/groom-beads/SKILL.md:240`
   ("**Do not declare the set ready while it exits non-zero.** Fix what
   it reports and re-run") has no reachable exit: there was nothing in
   the bead data to fix. Per the same line's escape hatch the agent ran
   two full cycles (identical exit 1 both times) and then surfaced the
   blocker, verifying graph integrity out of band instead — `bd dep
   cycles` → "✓ No dependency cycles detected", rc 0; `bd orphans` →
   "✓ No orphaned issues found", rc 0.

3. **[gap · snag]** No rule at `groom-beads` phase 2 for removing a
   blocked bead from the dispatch queue.
   `plugin/skills/groom-beads/SKILL.md:158-161` reads:

       WHEN a bead cannot be brought to Ready, do not dispatch it: set its
       `grooming_status` to `blocked-needs-info` (a specific fact is missing) or
       `blocked-ungroomable` (it needs redesign), and flag it in the report rather than
       shipping a half-groomed bead downstream silently.

   `grooming_status` is not a `bd` field — `bd update --help` on 1.1.2
   offers no such flag, so it lands in the metadata JSON. `bd ready`
   selects on the bead's `status`, which the skill never mentions.
   Observed: after one bead was flagged exactly as instructed
   (`metadata.grooming_status = blocked-ungroomable`), `bd ready` still
   listed it as available work. It left the queue only after the agent
   independently chose `bd update <bead-id> --status deferred`.
   That choice was an improvisation; the skill offered no rule, and a
   different agent would plausibly leave the bead at `status: open`,
   where an orchestrator picks it up. This defeats the gate the
   Definition of Ready states at
   `references/shared/work-order-definition-of-ready.md` lines 127-129
   (skill-relative): "It is flagged for the author on the tracker ...
   and excluded from dispatch until resolved."

   The schema cannot express the parked state either.
   `plugin/skills/groom-beads/references/bead.schema.json:23` is:

       "status": { "enum": ["open", "ready", "in_progress", "blocked", "closed"] },

   which has no `deferred`, though `bd statuses` lists it as a built-in
   and it is the only built-in status that is both frozen and hidden
   from `bd ready`.

4. **[misleading-text · snag]**
   `plugin/skills/groom-beads/SKILL.md:186` reads:

       bd update <id> --set-metadata '{"tier":"standard","tier_reason":"localized CRUD + tests, low ambiguity","est_size":"S"}'

   and `plugin/skills/groom-beads/SKILL.md:212` gives the same form for
   the C4 keys. Run verbatim, observed:

       $ bd update <bead-id> --set-metadata '{"tier":"standard","tier_reason":"probe"}'
       Error: metadata edit failed for <bead-id>: invalid
       --set-metadata: expected key=value, got
       "{\"tier\":\"standard\",\"tier_reason\":\"probe\"}"
       rc=1

   Expected: the metadata set. On `bd` 1.1.2 `--set-metadata` is a
   repeatable `key=value` flag and `--metadata` is the JSON-string flag,
   so the working form is `bd update <id> --metadata '{...}'`. This is
   scored a snag rather than a blocker because
   `plugin/skills/groom-beads/SKILL.md:88-93` pre-emptively instructs the
   agent to probe `--help` and "adapt to what `--help` actually shows" —
   which is what happened. It still misleads: it is the only concrete
   example for phases 4 and 5, the two phases that touch every bead.

Recurrence: one run (`runs_observed: 1`). No claim here asserts more.
Findings 1, 2 and 4 are `bd`-version-relative and were observed on
`bd` 1.1.2 (dev) in embedded Dolt mode; finding 3 is version-independent
because it concerns which field `bd ready` reads.

## Proposals

### P1 — Stop scoring a check's success message as a failure [blocker]

**Problem.** `plugin/skills/groom-beads/scripts/bd-health.sh:55` requires
empty output for a pass, so any `bd` build that prints "✓ No orphaned
issues found" on success fails a hard check. Finding 1: the script
reported "hard problems found — fix cycles/orphans" on a backlog with
zero of both, and all five checks rendered `✗`.

**Change.** Score on exit status, and treat output as detail rather than
as the verdict. Replace lines 55-62:

```sh
  if [ $rc -eq 0 ]; then
    if [ -z "${out//[[:space:]]/}" ]; then
      printf '  ✓ %-18s clean\n' "$label"
    else
      # A zero exit is the verdict; many `bd` builds print a success
      # line ("✓ No orphaned issues found"). Show it, but pass.
      printf '  ✓ %-18s clean\n' "$label"
      printf '%s\n' "$out" | sed 's/^/      /'
    fi
  else
    printf '  ✗ %-18s (exit %d)\n' "$label" "$rc"
    printf '%s\n' "$out" | sed 's/^/      /'
    if [ "$severity" = "hard" ]; then hard_fail=1; fi
  fi
```

**Evidence.** Finding 1. On the next run under the same model and `bd`
build, `orphans`, `lint` and `stale` report `✓ clean` and the script
stops asserting graph problems that do not exist — so phase 6's "do not
declare the set ready while it exits non-zero" becomes satisfiable by
grooming rather than unreachable. Hard-failure detection is not weakened:
a non-zero exit still fails, which is what actually signals cycles and
orphans.

### P2 — Scope `swarm validate` to an epic, or report it skipped [blocker]

**Problem.** `plugin/skills/groom-beads/scripts/bd-health.sh:67` calls
`bd swarm validate` with no argument; on `bd` 1.1.2 it takes exactly one
(an epic ID). Finding 2: it errors, is scored a hard graph failure, and —
because the `have()` probe at line 39 tests `--help` and therefore
succeeds — never reaches the `skipped_hard` path at lines 44-49 whose
whole purpose is to say the guarantee was not established. In a repo with
no epics there is no valid invocation at all.

**Change.** Replace the single call at line 67 with an epic-scoped loop
that degrades to the skipped path, and add a direct cycle check that
works without epics:

```sh
# `bd swarm validate` validates ONE epic; it takes an epic id and has no
# whole-graph form. Enumerate epics and validate each. No epics is not a
# failure — there is nothing for it to check.
if have "swarm validate"; then
  epics="$("$BD" list --type epic --status all --json 2>/dev/null \
           | grep -o '"id":"[^"]*"' | cut -d'"' -f4)"
  if [ -z "$epics" ]; then
    printf '  - %-18s (skipped: no epics in this backlog)\n' "swarm validate"
  else
    for e in $epics; do run_check "swarm validate $e" "swarm val $e" hard; done
  fi
else
  printf '  - %-18s (skipped: `bd swarm validate` not available)\n' "swarm validate"
  skipped_hard=$((skipped_hard + 1))
fi
run_check "dep cycles" "dep cycles" hard
```

Also tighten the `have()` contract so a probe that only proves the
subcommand *exists* cannot masquerade as proof it is *callable*: add a
comment at line 39 recording that `--help` succeeds for subcommands with
required positional arguments, so availability and applicability are
separate questions.

**Evidence.** Finding 2. Cycle detection becomes real rather than
nominal — `bd dep cycles` needs no epic and returned rc 0 on this
backlog — and a backlog with no epics reports honestly as skipped
instead of failing. Together with P1 this makes the phase 6 summary at
lines 84-88 mean what `plugin/skills/groom-beads/SKILL.md:241-243`
already tells the agent to read it as.

### P3 — Give phase 2 a rule for parking a blocked bead [snag]

**Problem.** `plugin/skills/groom-beads/SKILL.md:158-161` tells the agent
to set `grooming_status`, which is metadata; `bd ready` reads `status`.
Finding 3: the flagged bead stayed in the ready queue until the agent
improvised `--status deferred`, and `bead.schema.json:23` has no status
value that expresses "parked out of the queue".

**Change.** Two edits.

(a) Append to `plugin/skills/groom-beads/SKILL.md:161`:

```
`grooming_status` is metadata; `bd ready` selects on the bead's
`status`. So also move the bead out of the dispatch queue —
`bd update <id> --status deferred` (a frozen status, hidden from
`bd ready`) — or the orchestrator will pick up the very bead you
just flagged as un-ready. Record both, and name the blocker in the
bead's `design` so whoever unblocks it knows what is missing.
```

(b) Add `"deferred"` to the status enum at
`plugin/skills/groom-beads/references/bead.schema.json:23`:

```json
"status": { "enum": ["open", "ready", "in_progress", "blocked", "deferred", "closed"] },
```

**Evidence.** Finding 3. The next agent that blocks a bead removes it
from `bd ready` by rule rather than by invention, and the emitted-file
path can round-trip the same state the `bd` path produces.

### P4 — Correct the two `--set-metadata` examples [snag]

**Problem.** `plugin/skills/groom-beads/SKILL.md:186` and `:212` both
show a JSON object passed to `--set-metadata`, which `bd` 1.1.2 rejects
(`expected key=value`). Finding 4. These are the only concrete commands
for phases 4 and 5.

**Change.** Replace the command at line 186 with:

```
bd update <id> --metadata '{"tier":"standard","tier_reason":"localized CRUD + tests, low ambiguity","est_size":"S"}'
```

and at line 212 with:

```
bd update <id> --metadata '{"c4_changes":["auth-service"],"c4_impacts":["web-spa","billing-service"],"c4_docs":["docs/c4/auth-service/component.md"]}'
```

Then amend the flag list at `plugin/skills/groom-beads/SKILL.md:91` so
the probe step names both forms: `--metadata` for a whole JSON object and
`--set-metadata key=value` for single scalar keys.

**Evidence.** Finding 4. The example works as written on `bd` 1.1.2, and
the probe instruction at lines 88-93 keeps its role as the safety net for
other builds rather than being the only thing standing between the agent
and a failed command.

## Considered and Not Proposed

- **A bridge between phase 6's two validators.** `bd export` emits
  `issue_type`, `acceptance_criteria` (a string), and
  `dependencies[].depends_on_id` with `type: relates-to`, while
  `bead.schema.json` requires `type`, `acceptance` (an array), and
  `dependencies[].id` with `type: related` — so `validate_beads.py`
  cannot read `bd` output without a translation shim, which this run
  wrote (~50 lines) to obtain a supplementary check after P1/P2 left the
  prescribed gate unusable. Not proposed: line 244 of
  `plugin/skills/groom-beads/SKILL.md` scopes that validator to "the
  file-emit fallback", so running it on
  `bd` output is outside its stated contract and the shim was this
  agent's choice, not a forced improvisation. The one part of the
  mismatch that *is* in contract — the missing `deferred` status — is
  folded into P3.
- **A collision key for repos with no C4 docs.** Lines 220-224 of
  `plugin/skills/groom-beads/SKILL.md`
  make `c4_changes` the wave-collision key, and the phase 7 wave preview
  packs waves under it; in a repo with no C4 docs (this one — phase 5
  skipped cleanly per lines 227-228) there is no collision key, yet three
  beads here genuinely shared a file (`plugin/.claude-plugin/plugin.json`).
  Real gap, but no wrong action followed — the agent expressed the shared
  edit as an idempotent instruction inside each bead's `design` and one
  `blocks` edge, without improvising around the skill. One run, one repo;
  revisit if a run actually collides.
- **Phase 4's "never null" tier versus the schema's exemption for
  blocked beads.** `plugin/skills/groom-beads/SKILL.md:200-201` says to
  assign `standard` when unsure, while `bead.schema.json:27` exempts
  `blocked-*` beads from the design/acceptance/model requirements. The
  apparent tension resolved itself: a placeholder `standard` was assigned
  per the SKILL text, and `validate_beads.py` emitted the note
  "blocked-ungroomable: carries a placeholder tier — fine, but must not
  be dispatched until specified". The existing text handled it; no
  change needed.
- **`bd lint` silently dropping deferred beads.** Its count fell from 4
  to "3 issues checked" once one bead was deferred, so a
  parked bead's warnings stop being reported. This is `bd` behavior, not
  `groom-beads` guidance.
- **Phase 5's C4 skip.** Worked exactly as written; recorded as a clean
  pass, not a finding.

## Non-Goals

- Changing the Definition of Ready, the field guide, or the work-order
  template. `plugin/skills/groom-beads/SKILL.md:60-70` delegates those to
  the `spec` skill, and nothing in this run argued against their content.
- Changing what `tier` means or the tier-routing rubric. The rubric was
  applied without friction; whether the tiers assigned in this run were
  calibrated is an execution question no dispatch has yet answered.
- Replacing `bd swarm validate` as the epic-structure check. P2 keeps its
  intent and fixes only how it is invoked and scored.
- Reconciling `bd export`'s record shape with `bead.schema.json` — see
  Considered and Not Proposed.

## Adoption

P1 and P2 both patch `plugin/skills/groom-beads/scripts/bd-health.sh` and
SHOULD land together: P1 alone still leaves `swarm validate` forcing a
hard failure, and P2 alone still leaves `orphans` inverted. Together they
are what makes phase 6 a gate that can be passed by grooming.

P3 touches `plugin/skills/groom-beads/SKILL.md` and
`plugin/skills/groom-beads/references/bead.schema.json`; P4 touches
`SKILL.md` only. They are independent of each other and of P1/P2, though
P3 and P4 are naturally applied in one editing pass. Note that
`scripts/check-skills.sh:138-145` asserts the `tier` enum in
`bead.schema.json`; P3 edits the `status` enum only and does not disturb
it.

Both skill copies must move together: `plugin/skills/groom-beads/` is
canonical and `.claude/skills/groom-beads/` is its dogfood mirror
(`scripts/dogfood-install.sh`), and they were byte-identical at run time.

Backward compatibility: P1 changes failures into passes for `bd` builds
that print success lines and changes nothing for builds that print
nothing, so no currently-passing backlog starts failing. P2 changes a
hard failure into either a real per-epic check or an honest skip; a
project whose CI asserts `bd-health.sh` exits 1 on this repo would see
that flip — which is the defect being fixed. P3's schema change is
additive. P4 is example text only.

Expected effect: a grooming pass can complete phase 6 without verifying
the graph out of band, and a bead the skill declares un-ready stops
appearing in `bd ready`.

## Resolution

**Triaged 2026-07-31.** 4 proposals, all adopted and implemented.

| Proposal | Outcome |
|----------|---------|
| P1 — stop scoring a success message as failure | **adopted.** `bd-health.sh` now scores on exit code, not output emptiness, and prints a passing check's output without marking it failed. The summary line was also changed to name the checks that actually failed instead of asserting "cycles/orphans" — that assertion was a false blocker, since `groom-beads` phase 6 forbids proceeding while the script exits non-zero. |
| P2 — scope `swarm validate` to an epic or report it skipped | **adopted.** It now enumerates active epics and validates each; zero epics is recorded as a skipped hard check, never a failure. Closed epics are excluded. |
| P3 — rule for parking a blocked bead | **adopted.** Phase 2 now states that parking takes two writes and that `bd ready` reads `status`, not `grooming_status`. |
| P4 — correct the `--set-metadata` examples | **adopted-with-refinement.** Both examples corrected. The refinement: verification showed the two flags differ in more than syntax — `--set-metadata` is additive but stringifies values, `--metadata` takes real JSON but replaces the whole object. Since phase 5's `c4_*` fields are arrays, that distinction is load-bearing and is now documented alongside the fix. |

**Files changed:** `plugin/skills/groom-beads/scripts/bd-health.sh`,
`plugin/skills/groom-beads/SKILL.md`.

**Version bumps:** `groom-beads` 1.0 → 1.1.

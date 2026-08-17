---
title: "RFC: Probe push scope at pre-flight, and give discovered scope gaps
  one channel instead of two"
description: A push rejected for credential scope stalled a run for 20 minutes
  behind liveness signals that all read "finished", and a subagent routed a
  real scope gap to a retro comment because two steps of the dispatch prompt
  both claimed it.
slug: rfc-self-push-scope-and-discovered-work-channel
type: rfc
status: ready-for-review
updated_at: 2026-08-14T09:45:00Z
origin: self
skill: bead-bonanza
skill_version: "1.5"
model: claude-opus-5
effort: unrecorded
runs_observed: 1
severity: blocker
plugin_version: "n/a (in-repo dogfooding)"
---

# RFC: Probe push scope at pre-flight, and give discovered scope gaps one channel instead of two

## Summary

Run `1786686364-15` drained all 16 beads and produced correct work, but two
defects in the skill cost it a silent 20-minute stall and let a real scope gap
survive to the last wave.

Pre-flight measures push *mechanics* thoroughly — it even performs a real probe
push for the Dolt remote — but never measures push *scope* against a protected
path. A credential that cannot write `.github/workflows/**` is invisible until
a bead touches one, and the resulting rejection is indistinguishable from a
finished run by every liveness signal the skill prescribes.

Separately, the per-bead prompt gives a subagent two channels for work it
discovers outside its scope, and no rule for which wins. One took the cheaper
channel and the gap was caught two waves later by an epic's acceptance
criterion, not by the run.

## Provenance

| Field | Value |
|---|---|
| Harness | Claude Code (Claude Code on the web, remote execution environment) |
| Task the skill was applied to | 16 beads across 7 waves under two epics — `agent-plugins-ia5` (a CLI feature: new Document Core component + `links` subcommand) and `agent-plugins-eaa` (moving 14 shipped documents into `specs/` and inverting the vendoring), on branch `claude/validate-corpus-spec-cli-0t0cxo`, draft PR #22 |
| References / scripts / templates invoked | `bead-bonanza/SKILL.md` v1.5; `references/subagent-prompts.md`; `scripts/pre-flight-check.sh`; `scripts/bb-progress.sh`; `scripts/bb-close-bead.sh`; `scripts/bb-backup-run.sh`; `plan-sprint` v0.1; `recover-sprint` v0.1 |
| Run dates | 2026-08-14 |
| Model lineup (from `dispatch.md` manifest) | light `haiku`, standard `sonnet`, advanced `opus`, frontier `fable`; `CLAUDE_CODE_SUBAGENT_MODEL` unset, so tier assignments were honoured |
| Testimony of record | `rfcs/sprints/2026-08-14-markdown-links-and-shipped-specs.md` — the run's sprint report. Findings below are cited by its `<bead-id>#<ordinal>` ids; where this RFC quotes one, the report is authoritative |

## Evidence

### E1 — pre-flight probes push mechanics but never push scope [script-defect, blocker]

`scripts/pre-flight-check.sh` emits `BB_PUSH_NEEDS_REFSPEC` and
`BB_GIT_CONFIG_WRITABLE` (lines 515–516), and the skill leans on them:

> `SKILL.md:328` — "Those facts are ledger lines with a command behind them.
> **Do not re-assert from memory what the script just measured** — push
> mechanics in particular vary by container."

Neither fact covers whether the credential is *allowed* to write a given path.
This run's container had a PAT without the `workflow` scope. At wave 6b,
`agent-plugins-47b` committed `.github/workflows/checks.yml` and the push was
rejected:

```
$ git push -u origin claude/validate-corpus-spec-cli-0t0cxo
 ! [remote rejected] claude/validate-corpus-spec-cli-0t0cxo -> claude/validate-corpus-spec-cli-0t0cxo
   (refusing to allow a Personal Access Token to create or update workflow
    `.github/workflows/checks.yml` without `workflow` scope)
error: failed to push some refs
```

Observed: the run stopped. Expected: pre-flight blocks, or the wave table
routes workflow-touching beads to a path that works.

The rejection kills the **entire** push, not the offending file, so the bead's
own commit and the snapshot commit both stayed local. Two commits sat unpushed
for ~20 minutes while every prescribed signal read "finished": zero worktrees
(the orchestrator had merged and removed 47b's), a clean tree, the bead
`closed` in the tracker, and the progress ledger's last phase `pushed`.

The capability to catch this already exists in the same script. Lines 480–489
push a real scratch ref to measure Dolt write access:

```sh
_probe_ref="refs/dolt/preflight-probe-$$"
DOLT_WRITE_PROBE="git push ${REMOTE:-origin} HEAD:$_probe_ref"
if git push -q "${REMOTE:-origin}" "HEAD:$_probe_ref" >/dev/null 2>&1; then
```

The pattern is right; it is simply not applied to the path class that
actually blocks runs.

Minimal repro: on a container whose credential lacks the `workflow` scope, run
pre-flight (it reports no blocker), then commit any change under
`.github/workflows/` and push.

### E2 — two channels claim discovered scope gaps, and neither yields [gap, blocker]

`references/subagent-prompts.md:104` (step 3 of the per-bead prompt):

> "Stay strictly within this bead's scope. If you find work beyond it, file a
> new bead via `bd create` with a `discovered-from` link to {ID}, noting in one
> line whether it *blocks* this bead ... or is merely adjacent — the
> orchestrator decides whether it enters the run."

`references/subagent-prompts.md:118` (step 6):

> "Trigger list — record one WHEN: ... the bead contradicted repo reality; you
> could not satisfy an acceptance criterion as declared; a sequencing surprise."

A discovered scope gap satisfies both. `agent-plugins-eaa.3` hit exactly one
and took step 6 (sprint report finding `eaa.3#1`):

> `retro: Confirmed at start (per this bead's own AC item): agent-plugins-wu1
> already retyped c4-guide.md to standards, and agent-plugins-eaa.2's Design
> explicitly declined to claim it either ('it is not [handled in eaa.2];
> guides are separate'). So c4-guide.md is now a standards-typed document that
> ships nowhere and moves in neither bead — a real scope gap, not this bead's
> to close. This bead covers the 5 remaining guides.`

The subagent identified the gap precisely, named the two beads that created it,
and filed no bead. Nothing in the run then acted on it. It was caught only
because the parent epic's AC-1 states a **count** — "The **14** corpus-shaped
documents live under `specs/`" — which the orchestrator could check: `eaa.2`
moved 8 and `eaa.3` moved 5, so 13. `agent-plugins-eaa.4` was filed
mid-run to close it.

Had that acceptance criterion been unquantified ("the shipped documents live
under `specs/`"), the epic would have closed 13/14 and shipped a
standards-typed document vendored nowhere.

Step 6's own framing invites the wrong choice: it calls the channel "friction
only, no root-causing", which reads as *lower ceremony* than filing a bead, so
a subagent economizing on effort picks it.

## Proposals

### P1 — probe push scope against a protected path at pre-flight [blocker]

Add a probe mirroring the existing Dolt-write one, measuring what actually
blocks runs.

**Change** — in `scripts/pre-flight-check.sh`, beside the `refs/dolt` probe:

```sh
# Push SCOPE, not just push mechanics. A GitHub PAT without the `workflow`
# scope is refused on any push whose commits touch .github/workflows/**, and
# the refusal rejects the WHOLE push. Measure it now, on a scratch ref, rather
# than discovering it at whichever wave first touches a workflow file.
_wf_probe_ref="refs/heads/bb-preflight-wf-probe-$$"
_wf_tree=$(mktemp -d)
if [ -d .github/workflows ]; then
  git worktree add -q --detach "$_wf_tree" HEAD 2>/dev/null && \
  mkdir -p "$_wf_tree/.github/workflows" && \
  printf 'name: bb-probe\non: workflow_dispatch\njobs:\n  n:\n    runs-on: ubuntu-latest\n    steps: [{run: "true"}]\n' \
    > "$_wf_tree/.github/workflows/.bb-probe.yml" && \
  git -C "$_wf_tree" add -A .github/workflows/.bb-probe.yml && \
  git -C "$_wf_tree" -c user.email=bb@probe -c user.name=bb commit -q -m "bb probe" && \
  if git -C "$_wf_tree" push -q "${REMOTE:-origin}" "HEAD:$_wf_probe_ref" 2>/dev/null; then
    PUSH_WORKFLOW_SCOPE=yes
    git push -q "${REMOTE:-origin}" --delete "$_wf_probe_ref" 2>/dev/null || true
  else
    PUSH_WORKFLOW_SCOPE=no
  fi
  git worktree remove --force "$_wf_tree" 2>/dev/null || true
else
  PUSH_WORKFLOW_SCOPE=na
fi
emit BB_PUSH_WORKFLOW_SCOPE "$PUSH_WORKFLOW_SCOPE"
```

This does not `block` — a run touching no workflow file is unaffected. It
emits a fact the planner can route on.

**And** in `SKILL.md`, after the push-mechanics paragraph at line 328:

> WHEN `BB_PUSH_WORKFLOW_SCOPE=no`, the credential cannot push any commit that
> touches `.github/workflows/**`, and the rejection fails the WHOLE push rather
> than that file. Any bead whose scope includes a workflow edit SHALL be
> dispatched with that file excluded — it writes the intended content to
> `$RUN_DIR/<id>.workflow-fragment.md` and the orchestrator applies it through
> a credential that has the scope. Never let such a commit reach a push.

### P2 — make "compare HEAD to the remote ref" a landing check, not an inference [blocker]

The run's stall was invisible because nothing verified the push landed.
`SKILL.md:330` already says to "verify landings with `git rev-parse
<remote>/<branch>` after fetch, not `git status -sb`" — but only under the
`BB_PUSH_NEEDS_REFSPEC=yes` branch, so it reads as advice for one container
configuration rather than as the universal post-push step.

**Change** — lift it out of that conditional into its own rule at the end of
every wave:

> **A wave is not complete until its push is verified landed.** After pushing,
> `git fetch <remote> <branch>` and confirm
> `git log --oneline <remote>/<branch>..HEAD` is empty. A push can be rejected
> for reasons that leave the tree clean, the worktrees removed and the beads
> closed — every other signal reads "finished". This is the only one that
> distinguishes a landed wave from a rejected one, and it costs one fetch.

### P3 — name one channel for discovered work, and say why it is not the retro comment [blocker]

**Change** — in `references/subagent-prompts.md`, extend step 6's trigger list
with an explicit exclusion:

> **Work, not friction.** IF what you found is a unit of work someone must do
> — a gap in scope, a missing artifact, a prerequisite — it is a BEAD (step 3),
> not a retro comment, even when it also reads as friction. File it, then
> optionally record a `retro:` comment about *why the gap existed*. A retro
> comment is read by a later retrospective; a bead is read by this run's
> orchestrator. Routing work to the comment channel means nothing acts on it
> before the run ends.

**And** at step 3, name the case that produced this finding:

> This includes a gap you can see but that is not yours to close — a document
> no bead in the queue claims, a step every adjacent bead assumed another would
> take. "Not this bead's to close" is a reason to FILE it, not a reason to
> leave it unfiled.

## Considered and Not Proposed

- **"Have the orchestrator re-read every retro comment for missed work at each
  wave boundary."** Rejected: it makes the comment channel load-bearing for
  dispatch, which is the confusion P3 removes, and it scales with comment
  count (64 in this run) rather than with discovered work.
- **Blocking pre-flight when `BB_PUSH_WORKFLOW_SCOPE=no`.** Rejected as too
  strict: most runs touch no workflow file, and blocking them would be a false
  positive. The fact plus the routing rule in P1 is sufficient.
- **The orchestrator's own merge/worktree-removal sequence.** It behaved
  correctly throughout; the stall was the push, not the merge. No finding.
- **`bd` exposing acceptance criteria as `acceptance_criteria` rather than
  `acceptance` in `--json`.** Real friction, and it cost this session a wrong
  claim, but it is a `bd` interface property rather than a bead-bonanza defect.
  Below the bar for this RFC.

## Non-Goals

- Changing how waves are packed, or the C4 collision analysis. Two findings
  from this run bear on that and belong to `plan-sprint`, which owns it; they
  are filed separately.
- Changing the liveness signals `recover-sprint` prescribes. The same stall
  exposed a gap there; it is filed against that skill.
- Any change to the close ceremony, the progress ledger, or the snapshot
  mechanism — all three behaved correctly across 16 beads.

## Adoption

P1 and P2 are independent of P3 and can land in either order. P1's probe adds
one scratch push and one worktree to pre-flight; it is skipped entirely when
the repository has no `.github/workflows` directory. P2 is a text move plus a
sentence. P3 is text only.

All three are backward compatible: `BB_PUSH_WORKFLOW_SCOPE` is a new emitted
fact, and a run that ignores it behaves exactly as today.

Expected effect: the failure that cost this run 20 minutes becomes a pre-flight
fact and a routing rule, and the failure that cost it a two-wave-late scope gap
becomes a filed bead at the moment of discovery.

## Resolution

**Partial triage 2026-08-17. This RFC is NOT retired** — P1 and P3 are
recommended but not yet approved.

| Proposal | Outcome |
|---|---|
| P1 — probe push scope against a protected path at pre-flight | **recommended adopt-with-refinement; not yet approved.** Confirmed live: `scripts/pre-flight-check.sh` emits no `BB_PUSH_WORKFLOW_SCOPE`. Note for whoever implements it — the script already enumerates `.github/workflows/*.y{a,}ml` at line 415 for CI detection, so the directory probe has a precedent to follow rather than inventing one. **This is the root cause of the failure P2 makes visible**; adopting P2 without P1 means the stall still happens, just discovered in one fetch instead of twenty minutes. |
| P2 — make "compare HEAD to the remote ref" a landing check | **adopted with refinement.** The comparison was lifted out of the `BB_PUSH_NEEDS_REFSPEC=yes` conditional at `SKILL.md:328–332` — which now explicitly says verifying a push **landed** is not conditional on that fact — and stated as a rule at Phase 3 step 7, with the fetch-and-compare commands and the whole-push-rejection consequence. |
| P3 — name one channel for discovered work | **recommended adopt; not yet approved.** Confirmed live at `references/subagent-prompts.md:104` and `:118`, including the "friction only, no root-causing" framing this RFC identifies as inviting the wrong choice. |

**Refinement applied to P2, and why.** This proposal and
`rfcs/inbox/recover-sprint/rfc-self-worktree-absence-is-not-completion` P1 are
**the same check in two skills**, from the running and recovering sides of one
incident. Each now names the other in the same form so the two cannot silently
diverge. See that RFC's Resolution for the full note and for the shared-vendoring
option that was considered and deferred.

**Provenance check.** `bead-bonanza` v1.5 cited; v1.5 current at triage.
`SKILL.md:329–331` was present verbatim, with the landing check nested inside
the refspec conditional exactly as E1 describes.

**Dedupe.** No match in retired RFCs 0001–0023. P2 overlaps the
`recover-sprint` RFC at one point, handled above. P1 is the causal partner of
that same cluster and is recommended alongside it.

**Self-authorship, disclosed.** `origin: self`, `model: claude-opus-5`; this
session is configured for `claude-opus-5`. **The RFC was authored and triaged
by the same model.** P2's adoption ratifies the check its own author ran to
discover the stall.

**Files changed.** `plugin/skills/bead-bonanza/SKILL.md`.

**Version bumps.** `bead-bonanza` 1.5 → 1.6 (shared). Plugin 0.18.0 → 0.19.0.

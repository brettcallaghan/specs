# Sprint report — bead-bonanza hardening

| Field | Value |
|-------|-------|
| Date | 2026-08-12 |
| Run id | `1786562615-17` |
| Branch | `claude/bead-bonanza-refinement-261qvk` (PR #20) |
| Base sha | `b8b36d1` |
| Epic | `agent-plugins-41z` — cut bead-bonanza's redundancy — plus 14 sibling beads |
| Queue | 17 beads, enumerated at run start; no filter re-evaluated mid-run |
| Outcome | 16 closed, 1 blocked (`agent-plugins-9ko`), 0 skipped |
| Findings harvested | 73 `retro:` comments across all 17 beads |
| Orchestrator model | `claude-opus-5` |
| Subagent models | `sonnet` (waves 1a, 2A, 2B, 8ef.3), `opus` (d70, 8ef.1, 8ef.2, wave 4), `haiku` (wave 5) |

This is the sprint's own retrospective. It carries every harvested finding
verbatim, including the tail that warrants no RFC — `triage-rfc` defers on
recurrence, so a finding filtered out here can never be observed appearing
twice.

## 1. Grooming

**Every line number in every bead was stale, and this was the sprint's
dominant friction.** Fourteen of seventeen beads cited file-and-line
positions in their Design or description. All of them had moved by the time
the bead was dispatched, because earlier beads in the same run edited the
same files. Findings 1, 6, 15, 16, 38, 45, 53 all record it independently.
The mitigation that worked — stated in every dispatch prompt: *locate by
content, not by line number, and report the drift.* Nothing was mis-edited.
But this is a systemic property of any multi-bead sprint on one surface, and
grooming currently produces citations that are guaranteed to rot.

**Three beads' rule enumerations overclaimed.** `8ef.1` promises eighteen
rules and names fourteen; A12 and A14–A16 appear nowhere (finding 15).
`8ef.2`'s category B names B13 and B14 without describing them (finding 7)
and its B9 claims a rule appears twice when it appears exactly once in the
entire bundle (finding 8). `8ef.3`'s claim that two permission entries are
listed twice does not hold — one occurrence repo-wide (finding 46). In each
case the implementer found and actioned real duplicates by reading the
files, so the work landed; but the counts were asserted rather than counted,
and an acceptance criterion keyed to a count is not checkable.

**Two beads rested on premises the sprint itself invalidated.** `b13` asked
for disclosure of four side effects, two of which `8mv` had deleted earlier
in the same run (findings 54, 55) — its AC 1 was unsatisfiable as literally
worded. `9ko` assumed both size blockers would clear; only one did, so it
blocked (finding 73). Both are correct grooming at the time of writing,
invalidated by sequencing. The dependency graph expressed the ordering but
not the fact that the *content* of a downstream bead depended on an upstream
bead's outcome.

**One bead's AC was narrower than the property it meant to check.** `0z0`'s
AC 4 specifies a grep; a whole-line `grep -v jsonl` passes a line carrying
both the stale and the correct pathspec, and the SKILL.md site was prose the
literal string would never match (findings 29, 30). Only the broader AC 1
caught it. An AC written as a specific command can be strictly weaker than
the property.

**One AC was phrased as an absolute that had drifted.** `d70` cites a body
of 21,419 bytes; the measured baseline was 22,497 because the file grew
after filing (finding 67).

**A dependency edge made its own subtree undispatchable.** `8ef.1` carried a
blocking edge to `41z`, its own parent epic — a deadlock, since an epic
cannot complete until its children do. Five of the seventeen beads
(`8ef.1`→`8ef.2`→`8ef.3`, plus `x8r` and `9ko`) were unreachable. Repaired
before wave 1 by rewiring `8ef.1` onto `txq`, which is what `41z` itself
depends on.

**Two tier channels disagreed.** Four beads carry a `tier:` label; for
`txq`, `8ef.1` and `8ef.2` it contradicts `metadata.tier`, and the
`tier_reason` prose matches the metadata in every case. Resolved by taking
the higher of the two ("never run a bead below its tier"). Two channels that
can silently disagree is a routing hazard.

**Tier prediction accuracy: no escalations, two deliberate over-routes.**
See §"Routing evidence" below. Zero escalations across sixteen closed beads
is not evidence of good calibration — over-routing succeeds silently.

## 2. Specs

**`null result`.** No implementer's `retro:` comment named a spec document
as wrong, thin, or self-contradictory. Two findings cite
`specs/standards/skill-best-practices.md` (the extraction-beats-compression
rule and the 100-line ToC requirement) and both cite it as *load-bearing and
correct* — `8ef.1` used it to refuse a compression, and it is what makes the
new `findings-capture-protocol.md` ToC gap a real defect. The only
document-level defect found this sprint was in `AGENTS.md`'s environmental
ledger, which is repo configuration rather than a spec, and is answered in
§3.

## 3. Plan

Four amendments were recorded, all before or during wave 1: the `8ef.1`
deadlock repair, the `b13` readiness-gate exception, the tier-channel
resolution, and the batch worktree-chaining override. The wave partition
itself never needed resequencing — the C4 metadata predicted the collision
structure correctly, and epic `41z`'s own design decision 4 independently
reached the same verdict for its children.

**What would have produced a better initial bucketing: nothing about the
partition — but the plan should have carried a content-dependency layer.**
The wave order was right. What it did not encode is that `b13`'s *facts*
depended on `8mv`'s outcome and `9ko`'s *premise* depended on `41z`'s. Those
are not co-edit collisions (which the C4 signal catches) nor sequencing
constraints in the `c4_impacts` sense; they are "this bead's description
becomes false if that bead succeeds". Both were caught only because the
orchestrator read the beads and warned the subagents explicitly.

**The environmental ledger drifted mid-run, twice.** `d6x` observed 2
warnings where the orchestrator's ledger asserted 1 (finding 72) — an
orchestrator-supplied "established fact, do NOT re-diagnose" that had
already gone stale by the batch's second bead. Then `d70` fixed the second
warning, making `d6x`'s correction to "2" stale in turn; the orchestrator
re-measured on the merged tree and corrected it in a separate commit. A
ledger entry that a bead in the same run can invalidate needs a shorter
half-life than "the whole run".

## 4. Skill

**The batch template branches every bead in a batch from the same base.**
`references/subagent-prompts.md` tells a batch subagent to create each
worktree with `git worktree add -b <suffix> <path> HEAD`, but repo `HEAD`
only advances when the *orchestrator* merges, which happens after the batch
returns. So beads inside one batch do not build on each other — directly
contradicting the worktree lifecycle's "Branching from `HEAD` — the sprint
branch as it stands *now*, after the previous bead merged — is what makes
each bead build on the last." Harmless in wave 1 (two beads, `d6x` rebased
cleanly onto `txq`); wave 2A would have produced **four conflicting branches
on one file**. The orchestrator overrode the template with explicit chaining
(`worktree add -b td1 <path> q1h`) in every subsequent batch, and all five
batches then merged as single fast-forwards.

**`bd show --json` silently omits comment bodies.** It returns
`comments_omitted: true` and a `comment_count`, so a naive
`jq '.comments[]'` harvest returns zero findings and reads as a clean run.
The skill already mandates `bd export | jq` for the harvest; this is a
second, independent reason, and the orchestrator hit it on its first
acceptance check.

**Capturing a gate's exit status inside a worktree ships the output file.**
The skill's own guidance — capture `$?` directly, never through a pipe,
because `cmd | tail` reports tail's status — leads naturally to `cmd >out
2>&1`. Run inside the bead's worktree, `bb-close-bead.sh` stages that file
by design ("everything in that tree is this bead's work"). `x8r` shipped
`out1`–`out4` in its close commit and left four more untracked (finding 63).
The ceremony is not at fault; guidance that tells a subagent to redirect
output must name a path outside the worktree.

**The pipe-to-`tail` hazard is real and the orchestrator hit it too**, on
the merged tree: `scripts/check-skills.sh --strict | tail` reported `exit=0`
against a gate whose entire purpose is exiting 1. Now recorded in
`AGENTS.md` at the point of use.

**Monitor caps at 30 minutes regardless of `timeout_ms`.** Requested
3,600,000 ms; the tool reported 1,800,000 and expired three times across the
run, each requiring a manual re-arm. The skill tells the orchestrator to arm
a Monitor for any subagent past 5 minutes — most beads here ran 10–16
minutes and several batches ran over an hour — so re-arming is the norm, not
an edge case, and nothing in the skill mentions it.

**The `--gate-baseline` bar works and is load-bearing.** Every subagent
correctly compared against `0 error(s), 1 warning(s)` rather than expecting
exit 0, and none tried to "fix" the known-red warning. Recording a failing
baseline as the bar is a design that paid off.

**Verification-by-reconstruction is the right standard for template
refactors** and should be named in the skill. `8ef.3` assembled
preamble+template exactly as an orchestrator sends it and diffed normalized
word-sets both directions (finding 49), rather than inspecting the diff.

## Routing evidence

**Tier escalations: none.** No bead failed gates twice or required
re-dispatch at a higher tier. Sixteen of seventeen closed on the first
attempt; the one that did not (`9ko`) blocked on a false premise, not on
capability.

Escalations expose *under*-routing only. A bead groomed `frontier` that a
`light` model would have finished never appears here — it succeeds, and the
overspend is silent. **Zero escalations across a 17-bead queue is therefore
not a clean bill of health**; it is equally consistent with routing
everything too high. The next run should try one tier lower on the beads
whose `tier_reason` was thinnest — `td1` and `0z0` both read as mechanical
and both were over-routed to `standard` by batch tier without difficulty.

**Deliberate over-routes: two.**

| Bead | Assigned | Ran at | Why raised |
|------|----------|--------|------------|
| `td1` | light | standard | batched into wave 2A with three `standard` beads on `pre-flight-check.sh`; splitting would have cost a full gate run to save one tier on one S-sized bead |
| `0z0` | light | standard | batched into wave 2B with two `standard` beads on the same bundle; same reasoning |

Neither was dragged more than one tier and only one bead per bucket was
raised, so both are within the tier-batching rule. `txq` is a third,
ambiguous case: `metadata.tier` says `light` and its stale `tier:` label
says `standard`; it ran at `standard` under the take-the-higher rule.

## Findings, verbatim

All 73 harvested `retro:` comments, unfiltered, grouped by bead.

### `agent-plugins-txq` (4)

1. retro: Demonstrated the defect pre-fix by running the documented chain verbatim on the untouched tree: `scripts/check-skills.sh --strict && scripts/sync-shared-references.sh --check && scripts/sync-strata-docs.sh --check` printed only check-skills' own output (0 error(s), 2 warning(s), RESULT: FAIL) and stopped — no output at all from sync-shared-references.sh or sync-strata-docs.sh, chain exit=1. Confirms the && chain masks both later checks exactly as the bead states.

2. retro: Chose a fenced bash block running the three checks with ';' (not '&&') and per-check $? capture, printed as three labelled exit lines, over a scripts/ helper or three separate gate bullets. Reason: AGENTS.md already uses fenced bash blocks for the custom-gates section, so this matches existing doc convention and keeps the fix self-contained in AGENTS.md rather than adding a new script surface for a P1 doc-only bug.

3. retro: Ran the new documented form for real (not just read it) on the current tree: check-skills --strict exit=1 (0 error(s), 2 warning(s)), sync-shared-references --check exit=0 (all 11 vendored references current), sync-strata-docs --check exit=0 (all 3 vendored strata docs match /root/.cache/strata-src — this ran for real, not a SKIP, since a strata checkout is present). Overall gate exit=1, correctly non-zero because check-skills is red, while the other two now visibly report green instead of never running.

4. retro: Found beyond the bead's scope — check-skills.sh --strict currently reports 2 warnings (adds plugin/skills/spec/SKILL.md at ~22497 bytes) where the ledger's baseline states 1 warning. This predates my change (reproduced on the untouched tree too) and is a size-warning finding owned by agent-plugins-41z/agent-plugins-8ef.1 per this bead's own BLOCKS links, not this bead's fix. Left AGENTS.md's baseline text as-is since updating stale ledger numbers is agent-plugins-d6x's scope, not txq's; flagging here so d6x's citation audit picks it up.


### `agent-plugins-d6x` (5)

5. retro: Settled the admission rule's wording almost verbatim from the bead's own Design (it was already fully drafted there) but had to decide scope: applied it to agent-plugins/AGENTS.md's ledger, and to the bead-bonanza dispatch-plan.md and project-setup.md templates (both live in this repo). Could NOT apply it to /home/user/strata/AGENTS.md — that repo is not attached in this session and the top-level dispatch explicitly scoped this batch to 'AGENTS.md — the repository's bead-bonanza gates block and its environmental-failure ledger' (singular repo). Filed agent-plugins-oy4 (discovered-from, adjacent — does not block this bead) to carry the rule to strata's own AGENTS.md in a session that has it attached.

6. retro: Audited every existing ledger entry in AGENTS.md's 'Environmental-failure ledger' and 'strata CLI' sections (19 bullets total) against the new admission rule. Cheaply re-tested what I could: curl'd the actual github/proxy URLs the bd-install and strata-CLI entries cite (403/403/200/200, matching); read the current session-start.sh hook source line-by-line for every hook-behavior claim; read this session's own ~/.cache/strata-cli/build.log and measured target/ size (2m35s, 470M — the 470M matched exactly, the time ran higher than the documented ~1m50s). Two entries could not be cheaply tested (the generate-schemas.sh old-binary corruption claims) and were marked unverified with a settling command instead, per the rule's own escape hatch.

7. retro: Found the bead's own assumption did not match repo reality — the Design says to add the rule to CLAUDE.md 'wherever the equivalent guidance sits,' but agent-plugins/CLAUDE.md has no environmental-failure-ledger-equivalent section at all (grepped for 'ledger'/'proved'/'verified'/'claim'; git log --follow on CLAUDE.md going back to the file's introduction shows it never had one). Decided not to invent a new section there to force-satisfy the AC, since that would be scope creep beyond citation/audit work; left CLAUDE.md untouched and record the gap here rather than silently treating the AC as satisfied.

8. retro: Auditing turned up two ledger entries that were not merely uncited but factually stale: 'bd init + bd import' and 'bd init git-commits its own scaffolding, the hook unwinds it' — the current hook (read in full) uses bd bootstrap and contains zero git-commit/reset calls anywhere. This predates 2026-08-11's move to the Dolt remote (agent-plugins-92q.31) and nobody had updated these two bullets since. Corrected them in place with today's citation rather than just tagging 'unverified', since 'unverified' would understate it — they were actively wrong, not merely uncited.

9. retro: The check-skills.sh --strict baseline the top-level dispatch handed me as an established fact ('0 error(s), 1 warning(s))', 'do NOT re-diagnose') did not match what I actually observed running the exact same command today: 0 error(s), 2 warning(s) — a second SKILL.md (plugin/skills/spec) crossed the ~20000-byte guideline since that baseline was recorded. I did not re-diagnose or fix it (out of scope, owned by agent-plugins-41z/8ef.1 per agent-plugins-txq's BLOCKS links), but this bead's whole purpose is ledger accuracy, so I corrected AGENTS.md's own baseline text with today's citation. Sequencing surprise: an orchestrator-supplied 'established fact' had already drifted by the time the batch reached its second bead.


### `agent-plugins-d70` (4)

10. retro: What I tried first and abandoned — extracting only the 'Boundary cases' bullets. That lands the body at ~19,700 bytes: under the 20,000 cap but with ~300 bytes of headroom, so the next sentence added to this SKILL.md re-trips the gate. Extracting the Anti-Patterns table as well took it to 16,941 bytes.

11. retro: Implicit decision (which passages moved where) — two new files, both cited as new rows in the References table: references/document-type-boundaries.md holds the 10 'Boundary cases' bullets verbatim (conditional on the type choice being ambiguous), and references/anti-patterns.md holds the 25-row table verbatim (largely a restatement layer over rules already stated in the workflow and boundary bullets, and consulted at review time). Pointer wording at each cut site names the specific cases/failure modes so a cold reader knows whether the hop applies to them.

12. retro: Decided silently — (a) gave both new reference files 'type: guide' frontmatter so scripts/validate-corpus.py validates them instead of listing them as excluded/no-frontmatter; (b) bumped the SKILL.md frontmatter version 1.7 -> 1.8; (c) the new 'Reviewing a draft' References row also names references/document-readiness-checklist.md, which was previously reachable only from workflow step 7 and absent from that table.

13. retro: Bead states the body as 21,419 bytes; measured baseline in the worktree was 22,497 bytes (the file grew after the bead was filed). Same conclusion, but a criterion phrased as a delta rather than an absolute threshold would have been unsatisfiable as written.


### `agent-plugins-q1h` (3)

14. retro: observed pre-fix failure — the script printed 'Verdict / OK — safe to proceed to Phase 1.' and exited 0 for a run whose BB_DOLT_WRITE was computed afterward (awk over lines 470-527 showed the verdict block and 'exit $REPORT_FAIL' both preceding the DOLT_WRITE computation, with no block() call between them), matching the bead's cited repro exactly.

15. retro: decided silently — gave the moved probe its own numbered section ('9. tracker durability', renumbering the old '9. verdict' to '10. verdict') rather than folding it into the existing verdict section, and wrote the blocker text as one sentence combining the probe command, consequence, and fix+recheck instruction. The bead's Design named the shape ('measurement, consequence, fix, recheck') but not the exact wording or whether to add a section header.

16. retro: clean on the 'what didn't work' and 'bead lacked information' axes — the Design pinned the fix (move the probe above the verdict, call block() on 'no') and the fix applied cleanly on the first attempt; nothing to record there.


### `agent-plugins-td1` (3)

17. retro: observed pre-fix failure — 'grep -rn "worktree prune" plugin/skills/bead-bonanza/scripts/' returned nothing (exit 1) while SKILL.md:91 and subagent-prompts.md:352 both assert pre-flight runs it, matching the bead's cited repro.

18. retro: decided silently — the Design says the script should run prune 'before it enumerates worktrees', but pre-flight-check.sh does not enumerate worktrees anywhere (no 'git worktree list' call exists in it); the claim in SKILL.md/subagent-prompts.md is about the ORCHESTRATOR's later 'git worktree list' being accurate, not about anything inside pre-flight-check.sh itself. I placed the prune call at the very top of the script (right after cd to repo root, before any other check) rather than inventing an enumeration step to put it 'before' — this satisfies the docs' claim without adding scope the AC didn't ask for.

19. retro: clean on 'what didn't work' — the Design was fully pinned (add the prune call, do not weaken the docs) and applied without iteration.


### `agent-plugins-9p4` (3)

20. retro: observed pre-fix failure — 'grep -n block plugin/skills/bead-bonanza/scripts/pre-flight-check.sh' showed block() called for the HEAD-vs-default-branch condition (line 350) but not for worktree.baseRef, and the script printed 'worktree.baseRef: head (advisory)' — matching the bead's cited repro exactly (SKILL.md says 'both are checked', script comment says 'advisory, never blocking').

21. retro: decided silently — AC3 ('the per-wave re-run cadence is stated exactly once, in SKILL.md') was violated even before this bead: SKILL.md itself states the cadence twice (once in the 'Every bead runs in its own worktree' section, again near the Phase 0 BB_DOLT_WRITE paragraph). I kept the fuller statement in the first location and reduced the second to a one-line pointer, rather than touching references/subagent-prompts.md's separate 'Verify before every wave' operational procedure (ancestor-check instructions), which I read as the verification MECHANISM rather than a restatement of the cadence rule itself. The bead's Design didn't call out this internal SKILL.md duplication explicitly.

22. retro: what I tried that needed a workaround — to demonstrate the new block() firing without touching the real $HOME/.claude/settings.json (which the sprint depends on staying head-set per the run's ledger), I pointed HOME at a scratch directory with a fake settings.json for one throwaway invocation, then removed it immediately. Confirmed the blocker text appears in the BLOCKED list and the script exits 1.


### `agent-plugins-8mv` (3)

23. retro: decided silently — renamed DOLT_TELEMETRY's non-disabled value from 'skipped-sandbox' to 'not-suppressed' and DENY_ENTRY's non-present value from 'manual' to 'absent' (plus a new 'unknown' for no-jq/unreadable), since the old vocabulary named a write outcome ('skipped', 'manual' fix needed) that no longer applies once the script never attempts the write. Verified via grep that no other file in the plugin references these BB_* values or the old add_deny_entry function name, so the rename is self-contained.

24. retro: what I tried — confirmed the fix with two throwaway HOME overrides: one showing settings.json's md5sum is unchanged after a run (no write), another showing the real container's dolt config global list stays empty (no dolt CLI is even on PATH here, which is itself the 'na' branch, not the fixed 'not-suppressed' branch). Neither the sandbox-blocked branch nor the not-suppressed branch got exercised against a real dolt binary in this environment, so that specific message wording is unverified against real dolt output, only against the self-test's jq fixtures and code inspection.

25. retro: bead contradicted repo reality slightly — the real $HOME/.claude/settings.json already had the eventsapi.dolthub.com deny entry when I got to this bead, because the OLD (pre-fix) add_deny_entry ran during q1h/td1/9p4's earlier self-tests and functional exercises in this same session and wrote it. That's the bug this bead fixes, caught in the act; I left the existing entry in place (removing it is not in scope and worktree.baseRef, the setting the ledger says must survive, is untouched).


### `agent-plugins-b13` (5)

26. retro: the bead's cited line numbers (pre-flight-check.sh:269,:284,:455,:494-495; SKILL.md:1106-1113; bb-close-bead.sh:159-165) had all moved — four earlier-sprint beads (q1h, td1, 9p4, 8mv) touched pre-flight-check.sh. Located each effect by content (deny_entry_present, the DOLT_WRITE_PROBE push, the bash -c gate loop, msgfile) instead; matches the task instructions' warning.

27. retro: two of the bead's five disclosed facts had changed under it — agent-plugins-8mv (this sprint) removed the ~/.claude/settings.json and global-dolt-config WRITES the bead asked to disclose; pre-flight-check.sh now only detects and advises (deny_entry_present is read-only, see its own self-test at line ~122). Disclosed what the script does today (remote probe push, --gate-baseline bash -c, git worktree prune) instead of the two retired writes.

28. retro: AC 1 as literally read ("every write ... appears in the list") could not be satisfied against the bead's own enumerated facts, since two of those facts are no longer true. Satisfied its intent instead — the list now matches the script as it stands, verified live by running it and observing the actual probe push.

29. retro: decided silently to use 'git config extensions.worktreeConfig true' + a worktree-scoped core.hooksPath (rather than the shared .git/hooks) to synthesize a failing pre-commit hook in the scratch worktree, so the failure-path test of bb-close-bead.sh could not leak a hook into any other worktree sharing the same .git dir.

30. retro: pre-fix, running pre-flight-check.sh live showed the undisclosed push actually happening: 'refs/dolt/data write: yes (probed: git push origin HEAD:refs/dolt/preflight-probe-2925)' — a real write to the remote — while sed -n '1109,1113p' SKILL.md (and a grep for push/bash -c/prune over the whole disclosure bullet) showed no mention of it.


### `agent-plugins-0z0` (3)

31. retro: pre-fix, grep -rn ":!.beads'" plugin/skills/bead-bonanza/ | grep -v jsonl showed the stale pathspec at subagent-prompts.md:107 and recovery.md:159-160, while scripts/bb-close-bead.sh:154 actually runs 'git add -A -- \'':!.beads/*.jsonl'\''' — confirming the docs described a whole-directory exclusion the script does not perform.

32. retro: AC4 as literally worded ("a grep for ':!.beads' not followed by /*.jsonl") is line-based in the obvious implementation, so a whole-line grep -v jsonl silently passes a line that has BOTH the stale and the correct pathspec on it (recovery.md:160 does, in my first draft). Used a per-occurrence check instead of a whole-line one; worth pinning the AC to 'per-occurrence, not per-line' explicitly.

33. retro: SKILL.md's site (:1125) was prose — 'everything modified except `.beads/`' — not the literal ':!.beads' pathspec string, so AC4's grep alone would never have caught it; only AC1's broader 'no file describes the whole-directory exclusion' check does. Both ACs were needed together, not just the grep.


### `agent-plugins-viy` (3)

34. retro: what I tried that did not work first — cited the new pointer in operating-principles.md as 'references/shared/findings-capture-protocol.md' (matching how SKILL.md cites it). check-skills.sh --strict then reported 4 errors: 'backticked path does not resolve from this file'. The bug was mine — operating-principles.md already lives inside references/, so the correct relative cite from there is 'shared/findings-capture-protocol.md', not 'references/shared/...'. Fixed and re-verified back to the 0-error/1-warning baseline.

35. retro: decided silently to leave the '## Findings harvest' intro paragraph (bd list defaults to open issues...) in place rather than trimming it further, even though it echoes reasoning also present in the vendored protocol's Aggregation section. The Design scoped the deletion to the ~1,800-char '### Harvesting' pipeline subsection specifically (character count matched exactly), so I read the intro as bead-bonanza's own framing rather than in-scope restatement and left it untouched.

36. retro: agent-plugins-6uk (the linked BLOCKS bead deciding whether the harvest query should filter to closed beads) is still open, so the vendored copy was taken verbatim from retro-spec's current source with no query change, per Design's explicit out-of-scope note. No conflict encountered — recorded for visibility since the bead text called it out.


### `agent-plugins-8ef.1` (7)

37. retro: The bead promises eighteen rules but its description names only fourteen (A1-A11, A13, A17, A18); A12 and A14-A16 are never identified anywhere, so there was no way to action them as specified. I deduplicated the fourteen named rules plus three unnamed cross-file duplicates I found by reading the file pairs (recovery.md's interrupted-work restatement of operating-principles.md, SKILL.md's dispatch-manifest bullet, and the reference-header routing prose).

38. retro: Every line number in the description is stale, as warned, but two rules were also already partly done: 9p4 had reduced SKILL.md's second baseRef statement (the BB_WORKTREE_BASEREF paragraph) to a pointer, and viy had already moved A11's harvest pipeline out of operating-principles.md into the vendored shared/findings-capture-protocol.md, so A11's canonical location is now a two-hop pointer chain rather than the operating-principles.md range the bead names.

39. retro: A6 (bb-close-bead.sh, 5x) and A7 (bb-backup-run.sh, 4x) are specified with canonical locations inside SKILL.md and most of their duplicate sites also inside SKILL.md, but my dispatch prompt scoped this bead to CROSS-FILE duplication only and assigned within-SKILL.md duplication to agent-plugins-8ef.2. I actioned only their cross-file halves (recovery.md's restatements) and left SKILL.md's Phase-3-step-5 vs script-interface pairs for 8ef.2 — that pair is worth ~1,000 characters and 8ef.2 should be told it inherits it.

40. retro: What I tried that did not work — I first planned to delete recovery.md's 'Crash-resilient close' section wholesale, as the description instructs. That collides with acceptance criterion 4 ('no table of contents was modified'): the section has a Contents entry, and deleting it would either leave a dangling entry or force a ToC edit. I instead kept the heading and cut only the duplicated content (the write spec and the four re-listed recovery steps), leaving the rationale that lives nowhere else, so the ToC entry stays true and unmodified.

41. retro: Decided silently — where a deleted second statement carried a rationale the canonical statement lacked, I moved the rationale to the canonical site rather than dropping it (two sentences added to subagent-prompts.md: why a worktree is a sibling of the repo, and that a failed base-ref check is stop-and-fix not stop-and-abandon). This is the only place bytes moved rather than disappearing.

42. retro: Deviated on A8 ('the plan names beads, never a filter'). The description names dispatch-plan.md canonical and lists three sites; I cut SKILL.md's copy to a pointer but kept recovery.md's imperative in place, trimming only its rationale. A recovering orchestrator enters recovery.md cold mid-procedure and being sent to another file for 'do not re-run the filter' is exactly the pointer-vs-delete test's stranded reader.

43. retro: Left in both places deliberately, beyond the three named carry-overs: the environmental-failure-ledger rule is stated three times cross-file (SKILL.md's operating principle, project-setup.md's AGENTS.md block template, dispatch-plan.md's dispatch.md section). Each copy is the content of a different artifact a reader copies out, so cutting any one weakens a template rather than removing a restatement. Possibly one of the unnamed A14-A16; flagging rather than actioning.


### `agent-plugins-8ef.2` (9)

44. retro: every line number in the description had moved, but every named passage was still findable by content — nothing was pre-removed by the eleven earlier beads. The offsets are roughly uniform (-72 lines by the sprint-retro section), so locating by content cost little; the description's stale ranges are the only drift.

45. retro: the description's category-B enumeration is incomplete in the same way 8ef.1 reported for its sibling — B13 and B14 are named but never described. I found and cut three unenumerated within-file duplicates that plausibly are them: 'Two beads sharing a surface may batch; they SHALL NOT share a concurrent wave' (twice), 'The standing authorization finishes the launched scope; it never enlarges it' (verbatim twice), and the dispatch-plan-has-no-lean-path rationale.

46. retro: B9 (red-before-green, claimed 2x) is wrong — grep across SKILL.md, all references/ and scripts/ finds exactly one statement of it, in the operating principles. There is no copy in a dispatch template either, so nothing was cut and nothing should be: the single copy is the only one in the bundle.

47. retro: three cited passages were unactionable under the epic's binding constraints, so I left them and recorded rather than forcing them. recovery.md:15's 'the four states' count sits inside a '## Contents' block (constraint 1 forbids touching those); subagent-prompts.md:217's 'Why this matters' is inside a fenced dispatch-prompt template (constraint 2, and its neighbourhood is 8ef.3's scope); dispatch-plan.md:117-136 is a jq explanation plus a worked example, which constraint 4's no-compression rule protects — there is no historical narrative in it.

48. retro: decided silently, where the description names no canonical occurrence. B11 (Monitor + durable trigger): kept the 'Recovery and progress streaming' copy because it is in the pre-compaction opening, and trimmed the operating-principles bullet to a pointer. B2/B3 (CLAUDE_CODE_SUBAGENT_MODEL, main-loop no-routing): kept the Phase 4 3a copies, which carry the exact text to print, and pointed the Phase 2.5 copies at them. 'may batch / SHALL NOT share a concurrent wave': kept the C4-classification copy at the point of classification.

49. retro: settled AC 4 by restating rather than recounting — 'pre-flight's ten probes across two platforms' was deleted along with the whole clause that enumerated what each script encodes, which was itself a second statement of the per-script bullets below it. Also dropped four counts the description did not list: 'Two signals this check deliberately does NOT use', 'makes all three unreliable', 'the Definition of Ready's nine checks', and 'the two conditions above'.

50. retro: rationale judged deletable but KEPT under CLAUDE.md's exceptions. Exception 1 (looks removable, is not): 'There is no minimum bead count, and do not reintroduce one — a count in a list of safety conditions reads as one'; and 'triage-rfc defers on recurrence, so a finding recorded once and filtered out the next time can never be observed appearing twice'. Exception 2 (trap with a plausible wrong answer): 'This is a real degradation, not an equivalent path' in the main-loop fallback, and 'A zero count across a drained run is a broken query, not a clean run'.

51. retro: the description's ~10,900-character estimate is not reachable inside the epic's own constraints. Actual removal is 4,206 bytes (61,908 -> 57,702). The gap is not shortfall: ~2,000 of the estimate sits in passages the epic's constraints 1, 2 and 4 protect, and the rest was double-counted against the eleven earlier beads' cuts or is prose whose rule has no second home.

52. retro: 8ef.1's inherited pair is resolved — it is Phase 3 step 5 (bb-backup-run.sh) against the published script-interface bullet for the same script, and the interface bullet won per the keep-list. Step 5 now carries only the invocation, the never-from-a-worktree rule and the per-bead cadence carry-over; the ~900 characters of mechanics it restated (what the script copies, that it does not push, that Phase 4 deletes it) live in the interface bullet and the side-effects list.


### `agent-plugins-8ef.3` (5)

53. retro: the bead's Category-D line-number citations were mostly stale, as warned. SKILL.md:423-431 (bd-ready filter restatement) now sits at a different section; the actual filter block moved to ~359-364, and I verified via `bd prime` that it does NOT inject --label/--priority/-n syntax, so that block is not actually redundant with the environment and I left it untouched. SKILL.md:1116-1117 (dup of :325-326) no longer exists — the file is 1019 lines and grep finds the trailers-file text only once, already fixed upstream. SKILL.md:1131-1133 ('script usage line') is present in reworded form protected by the epic's binding constraint 3 (deliberate redundancy with script headers) — no action needed. recovery.md:160 (hand-replay of bb-close-bead.sh with the old pathspec) is already fixed: agent-plugins-0z0 is closed and the current text explicitly forbids by-hand replay with the correct :!.beads/*.jsonl pathspec.

54. retro: the description's claim that subagent-prompts.md 142-144 lists 'Bash(wait) and Bash(wait %1) twice' does not hold — a repo-wide grep finds Bash(wait)/Bash(wait %1) exactly once, in project-setup.md:91. No duplicate found or removed; noting it so the estimate isn't mistaken for a missed deletion.

55. retro: what I tried and abandoned — I first considered building the Shared batch preamble as text appended AFTER each template (so reading order in the file would match the old templates more closely), interleaving fragments back into their original positions. Abandoned it: the orchestrator's concatenation is a single preamble+template glue, and reinserting shared fragments at multiple points inside the template would require the orchestrator to splice, not concatenate — the Design's 'templates open by referencing it' phrasing points at a single front-loaded block, so I moved all four identical/near-identical fragments (conventions line, worktree block, per-bead-workflow step, return contract) to the front.

56. retro: silently decided — (1) renumbered each batch template's remaining per-template steps from 1/2/3 to 1/2 rather than keeping stale numbers or converting to prose, since the shared step moved into the preamble as unnumbered prose. (2) For the 'git hygiene' Constraints bullets (subagent-prompts.md, formerly cited as :142-144), I did not delete the full ~200-char span verbatim: the '--no-verify' clause and the 'never force-push a published branch' sub-clause are genuine restatements of the harness's own Git Safety Protocol (verified against this session's own system prompt) and were cut, but the surrounding follow-up-commit/no-amend guidance and 'only push current branch, never main' are bead-bonanza-specific rules not covered by that protocol, so those survive untouched — a literal 'delete the item' would have lost them.

57. retro: verification method for 'no rule lost' — for each of the two batch templates, I extracted the fenced prompt text only (stripped headers/prose) for: the pre-change template, the new preamble, and the new template; concatenated preamble+new-template to model exactly what the orchestrator sends; then took normalized (lowercased, punctuation-stripped) word-sets and diffed old-template-words minus assembled-words. Result for both fast-lane and shared-surface: the only word present in the old but absent from the assembled text was the numeral '3' (a list-numbering artifact from moving the shared step out of the numbered list, not lost content). The reverse diff (assembled minus old) showed only connective phrasing I added ('beyond', 'above', 'below', 'template-specific', 'when done') plus 'exactly' for the shared-surface template, which restores the 'exactly as in the sequential template' anchor that :209 had previously dropped (the drift the bead asked me to resolve).


### `agent-plugins-new` (5)

58. retro: Every line number in the Design was stale. recovery.md was 17,277 bytes / 251 lines, not 18,767 / 263 — the cited ranges :24-150, :151-171, :172-250, :251-263 all shifted, and :160 (the pathspec defect) had already been taken by agent-plugins-0z0 as the Design anticipated. Sections were located by heading text instead.

59. retro: The Design said the crash-resilient-close section 'deletes wholesale' as ~0% read, but agent-plugins-8ef.1 had already cut its duplicated half and kept the heading. What remained is two paragraphs stating WHY the handoff precedes the close — a rule not restated in SKILL.md — so it was kept with the subagent-crash procedures rather than deleted. No rule lost.

60. retro: Decided silently, not pinned by the bead: the new reference file is named references/progress-streaming.md; recovery.md keeps its filename but is retitled to subagent-crash-only; the split boundary is exactly the three former H2s, with 'Crash-resilient close' following the subagent procedure it serves. Also added .claude/skills/recover-sprint as a committed symlink (matching all six existing plugin skills) and listed the skill in .claude-plugin/marketplace.json, which the bead named as c4_impacts.

61. retro: What did not work: my first pass at SKILL.md's three-way routing grew the body by 842 bytes. Tightened to +720 (57,702 -> 58,422). SKILL.md necessarily grows here — one reference destination became three plus a skill name — and the win is on the read path: the ~90% Monitor/durable-trigger load is now 6,415 bytes instead of 17,277.

62. retro: user-invocable: false CONFIRMED against the installed build (Claude Code 2.1.228, /opt/claude-code/bin/claude). The binary's frontmatter key list includes 'user-invocable'; its schema text reads "If false, hides the slash command from users; only the model can invoke it via the Skill tool"; the parser defaults it to true when absent. scripts/check-skills.sh parses the frontmatter with yaml.safe_load and ignores unknown keys, so it accepts the flag (exit 0, 0 errors).


### `agent-plugins-b5z` (7)

63. retro: Every cited line range was stale. Phase 2.5 was SKILL.md:494-726, not :559-818; 'Classify each bead' was :505-572 not :570-642, 'Choose a dispatch mode' :574-614 not :643-686, 'Tier selection' :616-665, 'main-loop fallback' :667-701, 'Push policy' :703-712, 'The dispatch plan' :714-726. dispatch-plan.md was 8,163 bytes, not 7,663. Sections were located by heading text.

64. retro: The bead's MOVE list said all of 'Classify each bead' goes to plan-sprint, but that section textually ended with the 'Concurrent dispatch is ENABLED' block — entry conditions Phase 3 and 'Every bead runs in its own worktree' both dispatch against, and which the orchestrator re-verifies per wave. By the bead's own KEEP criterion (the orchestrator's own dispatch acts) I promoted it to its own '### Concurrent dispatch' section in SKILL.md rather than moving it. Recorded because the byte accounting in the Design assumed it moved.

65. retro: The bead did not pin what happens on a run with nothing to groom. Phase 1.5 dispatches a grooming subagent only when a bead fails the readiness check, so 'the grooming subagent invokes plan-sprint' would leave a fully-ready queue with no plan — contradicting AC 4. Settled it: one subagent, never two; where the queue needed no grooming, dispatch one for planning alone. SKILL.md Phase 1.5 step 1 and Phase 2.5 both say so.

66. retro: Decided silently — plan-sprint's internal structure is SKILL.md (Interface table in/out, Bias toward batching, Classify each bead, Choose a dispatch mode per bucket, Write the plan) plus references/dispatch-plan.md moved verbatim via git mv. dispatch-plan.md gained two edits only: its title lost the bead-bonanza prefix, and the run-manifest section now states in the file that the header is the orchestrator's and is not to be edited. The main-loop fallback became references/main-loop-fallback.md and gained two bullets, since with no Agent tool grooming and planning cannot be delegated either.

67. retro: What did not work: moving all of references/dispatch-plan.md takes the '## Amendments' rule out of the orchestrator's reach, yet amending the plan mid-run is the orchestrator's own act (admissions, tier escalations). Rather than split the file against the bead's binding list, SKILL.md's Phase 2.5 now carries the amendment rule as one of the three things that stay yours. The full format stays in plan-sprint.

68. retro: user-invocable: false CONFIRMED against the installed build, Claude Code 2.1.228 at /opt/claude-code/bin/claude. Checked: the bundled frontmatter key list contains 'user-invocable'; the schema string reads "If false, hides the slash command from users; only the model can invoke it via the Skill tool"; the parser reads e['user-invocable'] and defaults to true when absent. Nothing else in this repo used the flag before today. scripts/check-skills.sh yaml.safe_loads the frontmatter and ignores unknown keys, so it accepts it: 0 errors on both new skills.

69. retro: AC 5 verified by reading what remains, not by assertion: recover-sprint step 2 reads dispatch.md in full for the mandate, enumerated bead list, standing authorization, wave table, gate profile, environmental ledger and Amendments; recover-sprint cites no bead-bonanza reference file at all (check-skills path resolution is clean); and every manifest field recovery needs is in the header the orchestrator still writes itself.


### `agent-plugins-9ko` (1)

70. retro: Blocker agent-plugins-41z did not clear the size warning. scripts/check-skills.sh --strict exits 1 with bead-bonanza/SKILL.md at 51921 bytes, still over the ~20000-byte guideline. Enabling --strict in CI would break the default-branch workflow, so taking Blocked path per Design intent.


### `agent-plugins-x8r` (3)

71. retro: The bead's audit table was from the 92q epic, not this sprint. Actual changes this sprint: bead-bonanza (git log shows multiple beads touched it), spec (agent-plugins-d70). NOT changed: groom-beads, groom-feature-blueprint, retro-skill, retro-spec. Bumped only bead-bonanza 1.3→1.4 and spec 1.8→1.9, updated three RFC files to record this.

72. retro: the audit was scoped to the current sprint rather than to the 92q epic the bead names, so it bumped only bead-bonanza and spec and reported the other four skills as unchanged. 4443f24 in fact edited all four SKILL.md files while their version lines last moved 2026-07-30. The orchestrator completed the four bumps and corrected the three RFC closeout notes in a follow-up commit.

73. retro: the dispatch prompt instructed 'cmd >out 2>&1' for capturing gate exit statuses without a pipe. Run inside the bead's own worktree, those redirect files (out1..out4) were staged by bb-close-bead.sh, which stages everything in a worktree by design, and shipped in the close commit. Prompts that tell a subagent to redirect output must name a path outside the worktree.



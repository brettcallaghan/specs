import json

PLAN = "docs/plans/bead-bonanza-renovation.md"
RESEARCH = "docs/research/bead-bonanza-design/README.md"

nodes, edges = [], []

def bead(key, title, desc, typ="task", pri=2, parent=None, tier="standard",
         reason="", size="M", labels=None):
    n = {"key": key, "title": title, "type": typ, "priority": pri,
         "description": desc.strip(),
         "labels": (labels or []) + ["bb-renovation"],
         "metadata": {"tier": tier, "tier_reason": reason, "est_size": size,
                      "plan": PLAN}}
    if parent: n["parent_key"] = parent
    nodes.append(n)

def blocks(blocker, blocked):
    # bd dep add <blocked> <blocker>: the blocked issue depends on the blocker.
    edges.append({"from_key": blocked, "to_key": blocker, "type": "blocks"})

def body(context, ac, notes="", heading="Acceptance Criteria"):
    s = f"## Context\n\n{context.strip()}\n\n## {heading}\n\n" + \
        "\n".join(f"- [ ] {a}" for a in ac)
    if notes: s += f"\n\n## Notes\n\n{notes.strip()}"
    return s

CITE = f"Plan: `{PLAN}`. Review evidence: `{RESEARCH}`."

# ─────────────────────────────────────────────────────────── root + wave epics
bead("root", "Renovate bead-bonanza: eliminate, renovate, pivot to epics, mechanize, split",
     body(
       "A design review of `bead-bonanza` v1.7 found the procedure is 2.9x its own "
       "context budget (58,392 bytes against the 20,000-byte re-injection cap, so 34.3% "
       "survives a compaction), that it hand-builds in prose about eight coordination "
       "primitives `bd` already ships as commands, and that its retro loop terminates in "
       "a queue rather than an action.\n\n"
       "This epic carries out the plan. Ordering is eliminate -> renovate -> pivot -> "
       "mechanize -> split, with spikes first because seven claims are unverified and "
       "four of them gate an elimination.\n\n" + CITE,
       ["Every wave epic below is closed.",
        "`scripts/check-skills.sh --strict` exits 0 with no warnings - the first time it has.",
        "The plan's delta log reads `done` on every row, and research section 1 has been "
        "reissued as a second dated snapshot.",
        "`docs/plans/bead-bonanza-renovation.md` is deleted, per the docs/plans convention."],
       "Waves 0-2 and 5 can proceed independently of each other. Wave 3 gates wave 4. "
       "Wave 6 is last by design: by then the phases have been shrunk, rewritten, "
       "restructured and partly replaced by scripts, so the split is a smaller and "
       "better-informed cut."),
     typ="epic", pri=1, tier="advanced", size="XL")

waves = [
  ("w0", "Wave 0 - Spikes: settle the seven unverified claims", 1,
   "Seven timeboxed investigations, all independent. Four of them gate an elimination "
   "in wave 1, so this wave runs first. Per D12 this wave is run BY HAND using the "
   "`bd` commands the plan intends to adopt - not via a bead-bonanza run - so it "
   "doubles as a trial of the target state.",
   ["Every spike closes with a measurement in the form AGENTS.md's ledger admission "
    "rule requires: the command run, the conditions it ran under, the result observed.",
    "Each spike either files a follow-up bead or records an explicit 'no action'.",
    "Findings that change a wave-1 or wave-3 item are reflected in that bead before it starts."]),
  ("w1", "Wave 1 - Eliminate what the audit retired", 1,
   "Delete first, so wave 2 does not renovate something about to die. Everything here "
   "is removal: the progress ledger and its machinery, the snapshot script, the handoff "
   "file, two flags, and a fallback path that describes a different product.",
   ["Every listed mechanism is gone from `plugin/skills/`.",
    "`scripts/check-reference-links.py` passes, proving nothing surviving references a "
    "deleted mechanism.",
    "`SKILL.md` is measurably smaller and `scripts/` holds three scripts, not five."]),
  ("w2", "Wave 2 - Renovate: replace hand-rolled procedure with bd commands and scripts", 1,
   "Replace procedure the skill describes in prose with a `bd` command or a script. "
   "Nothing structural yet - that is wave 3. Ends by extracting the per-bead ceremony "
   "as its own skill, placed last deliberately because waves 1 and 2 both change the "
   "ceremony's content.",
   ["No executable or routing fact is parsed out of markdown any more.",
    "Claiming, merging and the human-decision channel each use a `bd` primitive rather "
    "than a prose rule.",
    "The per-bead ceremony exists as an independently invocable, independently testable skill."]),
  ("w3", "Wave 3 - The epic pivot: a run is an epic, and the epic is the queue", 1,
   "The structural change, and the one with the most downstream benefit. Stated as one "
   "decision followed by everything it unlocks, because the value is almost entirely in "
   "the second part: the epic id becomes the run id, `bd ready --parent` becomes the "
   "queue, the manifest attaches to the epic, and the swarm commands supply "
   "registration, waves, a live status view and a run-level completion signal.",
   ["A run can be started, recovered and reported on using only an epic id.",
    "`$RUN_DIR` holds nothing any procedure needs (D10).",
    "`plan-sprint`'s wave derivation is `bd swarm validate` plus a co-edit packing pass, "
    "not a hand-rolled partition."]),
  ("w4", "Wave 4 - Mechanize: three tiers, separated by precondition", 2,
   "Mechanization is not one wave but three, split by what each tier depends on. "
   "Separating them keeps the checks that are scriptable today from waiting on the "
   "checks that need the epic pivot or the checks script.",
   ["Every checklist item the review classed mechanical is verified by a script, not "
    "asserted by an agent.",
    "`recover-sprint` collapses to 'run the status script, then act'.",
    "A run has a health signal and a dispatch record - neither exists today."]),
  ("w5", "Wave 5 - Rebuild the retro end to end", 2,
   "Independent of waves 3 and 4. Starts with the end-to-end design, because a finding "
   "taxonomy decided before the flow it feeds is a guess. Per D11 the RFC path is "
   "partitioned by audience rather than retired: project-level findings become beads in "
   "the user's project; skill-level findings stay RFCs, because they are proposals about "
   "a product this repository vends to other projects.",
   ["One document describes capture, harvest, classify, report, decide and file, naming "
    "who performs each stage and where its output lives.",
    "A bead-bonanza run's project-level findings reach a human as candidate issues, not "
    "as a document nobody triages.",
    "The audience split is explicit in `bead-bonanza`, `retro-skill` and `retro-spec`."]),
  ("w6", "Wave 6 - Split SKILL.md into invoked skills", 2,
   "Last, per the sequencing rule. By now waves 1-4 have shrunk, rewritten and partly "
   "replaced the phases, so the cut is smaller and better informed than it would have "
   "been first. This is what finally retires the compaction cliff and the defences built "
   "around it.",
   ["`plugin/skills/bead-bonanza/SKILL.md` body is under 20,000 bytes.",
    "`scripts/check-skills.sh --strict` exits 0 with no warnings.",
    "The KNOWN RED exception for `--strict` is removed from AGENTS.md's ledger."]),
]
for key, title, pri, ctx, ac in waves:
    bead(key, title, body(ctx + "\n\n" + CITE, ac, heading="Success Criteria"),
         typ="epic", pri=pri, tier="advanced", size="L", parent="root")

# ─────────────────────────────────────────────────────────────── wave 0 spikes
SP = [
 ("s1","Spike: what does CLAUDE_CODE_SUBAGENT_MODEL actually override?","light","S",
  "`SKILL.md` asserts that when this variable is set, every subagent runs at that forced "
  "model regardless of tier, so no routing occurs - and Phase 4 replaces BOTH routing "
  "tables with an `N/A` on that basis. The premise is unverified and public accounts "
  "contradict each other: one puts the env var above the per-invocation `model` "
  "parameter, another says it only affects agents without an explicit model. "
  "claude-code issue 10993 exists because the behaviour is undocumented. bead-bonanza "
  "ALWAYS passes `model:` per-invocation, so if the parameter wins, routing happens "
  "normally and Phase 4 discards two correct tables.",
  ["Export the variable to a model outside the project lineup.",
   "Dispatch two subagents at two different resolved models via the `model` param.",
   "Ask each which model is serving it, and record both answers verbatim.",
   "State which of the two precedence rules holds, with the observed evidence.",
   "Record the finding so E6 can proceed or be cancelled."]),
 ("s2","Spike: what do event beads, set-state and metadata cost at run volume?","light","S",
  "Three storage decisions depend on this: the subagent dispatch record (D2, two event "
  "beads per dispatch), the health watchdog (`bd set-state <id> health=...`), and the "
  "manifest on the epic (D9, `--metadata` JSON). None has been measured, and `bd create "
  "--graph` types metadata as map[string]string, so array-valued metadata cannot go "
  "through the graph format at all.",
  ["Write 100 event beads against a scratch bead; time `bd dolt push` before and after.",
   "Run the eight phase transitions via `bd set-state` and record label churn.",
   "Establish whether `bd` bounds metadata size, and whether `--metadata` round-trips a "
   "nested object set via `bd update`.",
   "`bd delete` the scratch beads and confirm `bd doctor --check=pollution` is clean.",
   "Record all four results; state which of D2, D9 and M5 they support or refute."]),
 ("s3","Spike: does an epic-of-epics traverse for ready and swarm validate?","standard","S",
  "The multi-epic story rests on nesting: a cross-epic sprint becomes an outer epic "
  "parenting the feature epics. Nothing has confirmed that `bd ready --parent` and `bd "
  "swarm validate` reach grandchildren rather than direct children only. If they do not, "
  "P10 needs a different mechanism (`bd mol bond`, or explicit re-parenting).",
  ["Run `bd ready --parent <root>` against this renovation epic and confirm it returns "
   "leaf beads, not just the seven wave epics.",
   "Run `bd swarm validate <root>` and confirm it analyses the full two-level graph.",
   "Run `bd epic status` and record how a nested epic is counted.",
   "State the result and, if traversal fails, name the alternative mechanism for P10."],
  "This spike is answered as a side effect of filing this plan's own beads, since the "
  "graph is deliberately two levels deep."),
 ("s4","Spike: does a Monitor over a polled command match tail -F?","standard","S",
  "Deleting the progress ledger (E1) removes the run's only push signal: `tail -F` wakes "
  "the Monitor the moment an event lands, where `git log` must be polled. The replacement "
  "is a Monitor over `git log --oneline <branch>` - one line per bead close, which is "
  "exactly the `committed` event the Monitor already filters for, taken from state "
  "instead of testimony. Unmeasured.",
  ["Arm a Monitor of each kind over one wave.",
   "Record wake latency and whether the polled form misses events.",
   "State whether the polled form is an acceptable replacement, and at what interval."]),
 ("s5","Spike: does bd fire an issue-level script hook on create/update/close?","standard","S",
  "`docs/OBSERVABILITY.md` documents a `hook.exec` span carrying `hook.event` in "
  "{create, update, close} and a `hook.path` - which would give a push signal from the "
  "tracker write itself, with no parallel file, and could supersede S4 entirely. But "
  "`docs/CONFIG.md` documents no configuration key for it and `bd config list` shows none. "
  "Note this is distinct from `bd hooks`, which manages GIT hooks.",
  ["Grep the `bd` source for the hooks directory it resolves and the key that configures it.",
   "If a hook mechanism exists, demonstrate one firing on a `bd update`.",
   "State whether it can carry the live progress signal, and whether it supersedes S4."]),
 ("s6","Spike: is bd worktree better than raw git worktree here?","standard","S",
  "The bundled beads skill says to always use `bd worktree` rather than `git worktree`, "
  "because it auto-configures the beads redirect, gitignore entries and embedded mode. "
  "But bead-bonanza measured that `bd` already resolves the shared database through git's "
  "common directory from any worktree with no redirect file created - so the main stated "
  "benefit may be moot here. Its removal safety checks may still earn it.",
  ["Create a worktree each way and compare what `bd where`, `bd worktree info` and the "
   "filesystem show.",
   "Compare `bd worktree remove` against `git worktree remove` on a dirty worktree.",
   "State whether N6 should adopt it, and on which of the two grounds."]),
 ("s7","Spike: does embedded Dolt contend under concurrent subagents?","advanced","S",
  "Beads' documented default storage mode is embedded: single-writer, file-locked; "
  "multi-writer needs an external `dolt sql-server`. Concurrent worktree subagents each "
  "run `bd update`, `bd comment` and `bd dolt push` against one database. Five runs have "
  "reported no lock failure, so this may be fine at 2-4 agents - but nothing has measured "
  "it, and every proposal that adds tracker writes (D2, M5) makes it worse.",
  ["Dispatch a concurrent wave with `BD_OTEL_STDOUT=true`.",
   "Read `bd_db_lock_wait_ms` and `bd_db_retry_count_total` and record the distribution.",
   "State the concurrency ceiling this implies, and whether D2 and M5 change it."]),
]
for k,t,tier,size,ctx,ac,*extra in SP:
    bead(k, t, body(ctx + "\n\n" + CITE, ac, extra[0] if extra else ""),
         typ="spike", pri=1, parent="w0", tier=tier, size=size)

# ────────────────────────────────────────────────────────────── wave 1 eliminate
EL = [
 ("e1","Delete the progress ledger and its machinery","advanced","L",
  "A phase-by-phase audit found seven of the eight progress phases redundant and the "
  "eighth untrustworthy. `claimed` and `blocked` duplicate `bd` statuses; `committed`, "
  "`pushed` and `retro-written` duplicate acceptance signal 2, Phase 3 step 7 and a "
  "file-existence check; `implementing` and `gates-running` are read by nothing. Only "
  "`gates-passed` changes behaviour - it lets recovery's happy crash path skip a gate "
  "re-run - and it does so on TESTIMONY, which is the exact failure run 1786778523-19 "
  "recorded when a handoff reported a gate clean that exits 1 with 222 findings. The "
  "ledger's tail was measured right 29 of 31 times and 'accurate precisely when nothing "
  "is wrong and misleading precisely when something is'.",
  ["The eight-phase vocabulary and `scripts/bb-progress.sh` are deleted.",
   "`references/progress-streaming.md`'s schema and writing sections are deleted; the "
   "Monitor and durable-trigger guidance is retained and re-sourced per S4/S5.",
   "The 'ledger is a hint, never the authority' rule is removed from all four places it "
   "appears, and `recover-sprint` step 4's per-phase distrust table with it.",
   "Recovery re-runs gates rather than trusting a recorded pass; the cost is stated in "
   "the skill (1 crash in 31 beads measured, and the shortcut did not fire on it).",
   "Every dispatch template stops carrying the phase vocabulary."]),
 ("e2","Delete bb-backup-run.sh, the snapshot commits and the deletion commit","standard","M",
  "The script exists to make `$RUN_DIR` survive a container reclaim - so once E1 and E3 "
  "empty that directory, it has nothing to preserve. It also carries a verified defect: "
  "it and Phase 4 both claim that deleting `.bonanza/<run-id>/` loses nothing because "
  "branch history retains the adding commits, but this repository squash-merges (four "
  "recent PR merges each have a single parent) and no commit reachable from any ref "
  "touches `.bonanza/` at all.",
  ["`scripts/bb-backup-run.sh` is deleted.",
   "Phase 3 step 5 and Phase 4 step 6b are removed, along with the ~N per-bead snapshot "
   "commits and the deletion commit they produced.",
   "The 'the live directory is authoritative and the snapshot never is' rule is removed "
   "from all three documents that carry it.",
   "`recover-sprint`'s find-the-run step no longer offers the snapshot as a route."]),
 ("e3","Delete the per-bead handoff file; fold its caveat onto the bead","standard","M",
  "Measured across run 1786345424-bbh: the handoff file was written 23 times and needed "
  "ONCE - on the one bead that crashed, where it was ABSENT, because the agent died "
  "between emitting `retro-written` and writing the file. The close ceremony now writes a "
  "minimal one itself when none exists, so the file is produced by the ceremony anyway, "
  "and its entire remaining payload is one `Caveat:` line that is `none` in the "
  "overwhelming majority of cases.",
  ["`<id>.handoff.md` is no longer written by any template or script.",
   "The caveat is recorded on the bead itself; the mechanism is named in the skill.",
   "The 'write the handoff BEFORE the close' ordering rule is removed.",
   "`references/recovery.md`'s crash-resilient-close section is removed."]),
 ("e4","Delete the --gate-baseline flag; the baseline becomes unconditional","light","S",
  "`pre-flight-check.sh` defaults `DO_GATES=no`, but `SKILL.md`'s single canonical "
  "invocation passes `--gate-baseline` unconditionally - so it is already always-on and "
  "the flag only creates a way to get it wrong. Neither argument for opt-in survives: "
  "'it executes arbitrary commands' is an argument about where the commands come from, "
  "answered by the checks script; 'a full suite costs ten minutes' is backwards, since a "
  "slow suite is exactly where charging a pre-existing failure to the first bead costs "
  "most. There is no opt-out either.",
  ["The `--gate-baseline` flag is removed; the baseline is always captured.",
   "`--self-test` still works and still takes no run id.",
   "The published interface in `SKILL.md` and the script header agree.",
   "No caller passes the flag."]),
 ("e5","Delete the tier: label path and its unusable bulk-query example","light","S",
  "Tier is stored twice by two skills: `groom-beads` writes it to metadata, `bead-bonanza` "
  "reads a `tier:<name>` LABEL and documents `bd list --label tier:standard --json` as a "
  "bulk query. Measured on this database: 153 beads carry `metadata.tier`, 63 carry a "
  "`tier:` label, the sets do not coincide, and the documented query returns 0 because "
  "`bd list` defaults to open issues - the same trap the findings harvest hit and fixed.",
  ["Metadata is the single home for tier; the label path is removed from `bead-bonanza`.",
   "The `bd list --label tier:standard --json` example is deleted, not corrected.",
   "Any remaining bulk query over tier is demonstrated to return a non-zero count on this "
   "database.",
   "`groom-beads` and `bead-bonanza` agree on one storage location."]),
 ("e6","Delete the CLAUDE_CODE_SUBAGENT_MODEL handling","light","S",
  "Four lines of `pre-flight-check.sh`, a `BB_*` fact, four `SKILL.md` mentions and two of "
  "Phase 4's three special-case branches exist for an ambient variable this project never "
  "sets and that is unset in the container. If S1 finds the per-invocation `model` "
  "parameter wins, none of it can affect a dispatch and all of it deletes.",
  ["S1 has reported, and its finding is cited in this bead's resolution.",
   "If the parameter wins: the detection and both Phase 4 `N/A` branches are deleted.",
   "If the env var wins: nothing is deleted, the branch is confirmed correct, and the "
   "skill states the precedence as a measured fact rather than an assumption.",
   "Either way the skill no longer asserts an unverified precedence."]),
 ("e7","Delete main-loop-fallback.md; make no-Agent a blocking condition","standard","M",
  "Subtract what the fallback removes and nothing distinctive is left: no subagents so no "
  "context isolation (the skill's stated whole point), no worktrees, no batching, no tier "
  "routing, no concurrency, no plan-sprint, no grooming subagent. What remains is the "
  "per-bead ceremony - genuinely valuable, and not this skill. It is 54 lines describing a "
  "different product, reached by a condition no script can measure, that no sprint report "
  "shows was ever exercised.",
  ["`references/main-loop-fallback.md` is deleted.",
   "Phase 0 step 0.6 becomes a BLOCKING pre-flight condition naming the per-bead ceremony "
   "skill as the alternative.",
   "`BB_ASSUME_NO_AGENT=1` asserts the condition and is echoed as a `BB_*` fact, so it "
   "reaches the manifest - 0.6 is currently the only Phase 0 fact absent from it.",
   "Phase 4's 'no routing occurred' fallback branch is removed with it."]),
]
# e4 depends on the checks script (w2) and e7 on the ceremony extraction (w2), so
# leaving them in w1 makes a cross-wave cycle once wave epics block each other.
# m1 is the "no preconditions" mechanization tier and must not be stranded behind w3.
REHOMED = {"e4": "w2", "e7": "w2"}
for k,t,tier,size,ctx,ac in EL:
    bead(k, t, body(ctx + "\n\n" + CITE, ac), pri=1,
         parent=REHOMED.get(k, "w1"), tier=tier, size=size)

# ────────────────────────────────────────────────────────────── wave 2 renovate
RN = [
 ("n1","Replace gate-profile prose parsing with a project checks script","advanced","L",
  "Pre-flight parses executable commands out of `AGENTS.md` markdown bullets with awk, "
  "plus a documented rule about which bullets are commands and which are commentary, then "
  "runs `bash -c` on each. It is the most fragile step in the pre-flight and the reason "
  "`--gate-baseline` was ever opt-in. A project-supplied `checks` script gives one known "
  "entry point the project commits and owns. Supersedes existing bead agent-plugins-pq06 "
  "- reconcile with it rather than duplicating.",
  ["A project declares its gates by committing a `checks` script with a defined contract "
   "(modes, exit codes, machine-readable per-gate results).",
   "`pre-flight-check.sh` invokes it and stops parsing `AGENTS.md` for commands.",
   "The gate profile, the baseline and the targeted-tests-by-surface mapping all resolve "
   "through it.",
   "`references/project-setup.md`'s gate-profile template is replaced by the script contract.",
   "This repository ships its own `checks` script and the run uses it.",
   "agent-plugins-pq06 is closed or explicitly superseded."]),
 ("n2","Claim with bd update <id> --claim","light","S",
  "Ceremony step 1 runs `bd update <id> --status in_progress`, a non-atomic claim, and the "
  "skill then adds a 'never reset a claim you did not set' rule to cover the race. `bd "
  "update <id> --claim` is atomic - it sets assignee and status together and is idempotent "
  "if already claimed by you. Note this is NOT `bd ready --claim`, which claims the first "
  "ready issue matching a filter: a pull model for a worker choosing its own work, wrong "
  "for an orchestrator dispatching a named bead.",
  ["Every dispatch template claims with `bd update <id> --claim`.",
   "The claim-reset rule is re-scoped to what atomicity does not cover, or removed if it "
   "covers everything.",
   "The assignee the claim sets is stated, since D2's dispatch record may want to use it."]),
 ("n3","Serialize merges with bd merge-slot","standard","M",
  "The merge policy is prose the orchestrator maintains: merge one at a time, `--ff-only` "
  "in a sequential wave, rebase-then-fast-forward for the second onward in a concurrent "
  "wave, and read the rebase target at merge time rather than earlier. That invariant "
  "broke in run 1786778523-19 - a rebase onto a stale target, then a branch deleted before "
  "re-checking, leaving the commit unreferenced and recovered from the object store. `bd "
  "merge-slot` is an exclusive lock with a priority-ordered waiter queue, built for exactly "
  "this.",
  ["The orchestrator acquires the merge slot before merging and releases it after.",
   "The slot is visible to every reader, so a held slot explains a waiting wave.",
   "The read-the-target-at-merge-time rule is expressed by the lock rather than by prose, "
   "or is retained with a stated reason why the lock does not cover it.",
   "A failure path is defined for a slot held by a dead agent."]),
 ("n4","Add a project config file for the model lineup and run defaults","standard","M",
  "The `## model lineup` block in `AGENTS.md` is read as PROSE by the orchestrator to "
  "resolve tier to model. With N1 taking the gates, this is the last executable or routing "
  "fact parsed out of markdown. Note this is the project-level half only: per-bead tier "
  "metadata is N5's subject and a config file does not touch it.",
  ["A config file carries the tier-to-model lineup, and pre-flight emits it as a fact.",
   "Run defaults that are currently prose or asked-per-run move here where they are "
   "genuinely project-level (worktree root, default push mode, run root).",
   "The `## model lineup` block is retired from `AGENTS.md` and from "
   "`references/project-setup.md`.",
   "A project with no config file gets a stated, non-guessing fallback."]),
 ("n5","Align bead metadata to bd's documented execution_* convention","standard","M",
  "`bd`'s `docs/METADATA.md` defines a convention for exactly this data - "
  "`execution_agent_type`, `execution_suggested_model`, `execution_reasoning_effort`, "
  "`execution_parallel_group` - and states that orchestrator agents must consume these keys "
  "before spawning subagents. bead-bonanza and groom-beads invented a private vocabulary "
  "instead. Adopting the documented names makes a groomed bead portable to any orchestrator, "
  "and `execution_reasoning_effort` closes the `effort: unrecorded` hole in the manifest.",
  ["`groom-beads` writes the `execution_*` keys alongside `tier`/`tier_reason`.",
   "`bead-bonanza` and `plan-sprint` read them.",
   "The effort setting is recorded rather than written `unrecorded`.",
   "A migration path is stated for the 153 beads already carrying `metadata.tier`."]),
 ("n6","Adopt bd worktree if the spike says it earns its place","light","S",
  "Gated on S6. The bundled beads skill says always use `bd worktree`; this skill measured "
  "that `bd` already resolves the shared database through git's common directory, so the "
  "redirect configuration may be moot. Its removal safety checks may still earn it.",
  ["S6 has reported and its finding is cited.",
   "Either the worktree lifecycle uses `bd worktree create/remove`, or the bead records "
   "why raw git remains correct here - with the measurement behind it."]),
 ("n7","Route the decision-required bounce through bd human","standard","M",
  "A bead needing a human decision is bounced via a `decision-required` label plus a prose "
  "procedure. `bd human` is a first-class channel with list, respond, dismiss and stats, and "
  "`bd gate create --await human:<prompt> --timeout` blocks a step until a human answers. "
  "That turns the run's single decision point into a queryable state rather than a report "
  "line the user has to notice.",
  ["A bounced bead is raised through `bd human` (and/or a human gate), not only a label.",
   "The orchestrator surfaces the open human items at the point it currently prints the "
   "bounce list.",
   "A resolved human item re-enters the readiness gate without a full re-run.",
   "The interaction with the standing authorization is stated: a human item stops the "
   "chain, exactly as `decision-required` does today."]),
 ("n8","Observe CI with a gh:run gate, and stop asking the same question twice","standard","M",
  "`BB_PR_CI` is measured at pre-flight and gates one question: whether to open a draft PR. "
  "But the shared-surface batch template asks the user the SAME underlying question - is a "
  "PR open so pushes trigger full CI, the 'slim gates' decision - without referencing the "
  "measured fact. So a run can ask twice, and the uninformed ask is the one that decides "
  "whether a batch skips its full suite. Separately, 'CI is the backstop' is asserted and "
  "never observed; `bd gate create --await gh:run:<id>` observes it without polling.",
  ["The slim-gates decision consumes `BB_PR_CI` rather than asking independently.",
   "Where a PR is open, the run observes its CI result through a `gh:run` gate rather than "
   "assuming it.",
   "`BB_PR_CI`'s meaning is stated precisely: it answers whether a draft PR would cause "
   "anything to run, not whether CI is healthy or a PR already exists.",
   "The substring-match limitation is recorded (a `pull_request` mention in a comment counts)."]),
 ("x1","Extract the per-bead ceremony into its own skill","advanced","L",
  "The ceremony is a self-contained, ordered, ten-step procedure with three scripts behind "
  "it, currently living as a prose template inside `references/subagent-prompts.md` that the "
  "orchestrator is explicitly forbidden to paraphrase - a rule that exists only because the "
  "procedure has no other form. As a skill it becomes what a subagent invokes by name, what "
  "a solo agent with no Agent tool invokes directly per bead, and something independently "
  "testable, which nothing about the ceremony is today. Placed last in wave 2 deliberately: "
  "waves 1 and 2 both change the ceremony's content, so extracting earlier means editing it "
  "twice.",
  ["The ceremony exists as an invocable skill with its own acceptance criteria.",
   "Dispatch templates invoke it by name; the do-not-paraphrase rule is removed.",
   "The skill is usable standalone by an agent working one bead, with no orchestrator.",
   "`references/subagent-prompts.md` keeps only what is genuinely per-dispatch: the "
   "placeholders and the batch-specific steps.",
   "Its body is under the 20,000-byte re-injection cap."]),
]
for k,t,tier,size,ctx,ac in RN:
    bead(k, t, body(ctx + "\n\n" + CITE, ac), pri=1 if k in ("n1","x1") else 2,
         parent="w2", tier=tier, size=size)

# ───────────────────────────────────────────────────────────── wave 3 epic pivot
PV = [
 ("p1","A run is an epic: the epic id becomes the run id and the queue","advanced","L",
  "The structural decision. Today a run id is a generated `<timestamp>-<count>` that "
  "survives only in the orchestrator's transcript - unguessable, and grep for it across the "
  "tracker, git history and the worktree returns zero. And the queue is an enumerated id "
  "list, which is STRUCTURALLY incompatible with every `bd` command that computes over a "
  "graph: `bd list` has no positional ids at all, and swarm validate, swarm status, ready "
  "--parent and graph each take exactly one node. An epic is a graph node whose membership "
  "is an enumeration, so it gives both at once.",
  ["Grooming creates or names a sprint epic; the run's beads are its descendants.",
   "The epic id IS the run id, everywhere a run id appears.",
   "The queue is `bd ready --parent <epic>` - scoped and dependency-ordered from one "
   "durable fact.",
   "The 'enumerate, never filter' rule is removed from all four documents carrying it, and "
   "the reason it existed is addressed by P8 rather than by a snapshot.",
   "The run root and the `current` symlink are retired."]),
 ("p2","Attach the run manifest to the epic; retire $RUN_DIR","advanced","L",
  "Per D9 the plan splits into a derived part and a recorded part. The wave table is "
  "DERIVED - recomputable from the graph once P5 and P9 land, so storing it only creates "
  "drift. What is not derivable is the manifest (gate profile, baseline, lineup, effort, "
  "standing authorization, environmental ledger, base sha) and the amendments log. Per D10 "
  "the manifest is also exactly what gets injected into every child prompt, so moving it to "
  "the epic empties `$RUN_DIR` entirely.",
  ["The manifest is `--metadata` JSON on the epic; `bd show <epic> --json | jq .metadata` "
   "retrieves it whole.",
   "Amendments are `amend:`-prefixed `bd comment`s on the epic - append-only and timestamped "
   "by construction, no read-modify-write.",
   "The wave table is not stored; it is recomputed.",
   "Trailers move to the epic, and the session is recorded with `bd update <id> --session` "
   "(this also closes agent-plugins-oyf).",
   "`$RUN_DIR`, the run root and `dispatch.md` are retired.",
   "`recover-sprint` reads the epic, not a directory."]),
 ("p3","Register the run with bd swarm create, and use swarm status as the live view","standard","M",
  "A run currently has no status view at all and nothing shows that a run exists. `bd swarm "
  "create <epic>` registers it; `bd swarm list` shows every run in flight across the repo; "
  "`bd swarm status --json` is a computed completed/active/ready/blocked view.",
  ["A run registers a swarm at start.",
   "`bd swarm status --json` replaces hand-maintained progress narration as the source for "
   "the wave report.",
   "`bd swarm list` is the answer to 'is a run in flight', including for a fresh session.",
   "The run checklist reports from it rather than from memory."]),
 ("p4","Take dependency waves from bd swarm validate; leave co-edit packing to plan-sprint","advanced","L",
  "Verified: `bd swarm validate <epic>` reproduced this repo's own epic 9kb six-wave "
  "partition, max parallelism and worker-session estimate from the dependency graph alone. "
  "`plan-sprint` re-derives that by hand. It also reports what plan-sprint has no check for: "
  "wrong dependency direction, orphans, cycles and disconnected subgraphs - grooming defects "
  "the 2026-08-15 retro found by hand.",
  ["`plan-sprint` calls `bd swarm validate` for the dependency waves and stops re-deriving "
   "them.",
   "It keeps what `bd` cannot compute: C4 co-edit collisions and dispatch-mode packing.",
   "Its integrity findings (direction, orphans, cycles, disconnected subgraphs) are surfaced "
   "as grooming defects rather than discarded.",
   "The skill states the division: `bd` owns the graph, plan-sprint owns the packing."]),
 ("p5","Add a run-level Definition of Done via bd epic status","standard","M",
  "The acceptance check is strong per BEAD and absent per RUN: a run ends when the queue "
  "drains, not when the epic's stated outcome is demonstrated. The long-running-agent harness "
  "pattern's rule is 'mark complete only after end-to-end verification confirms it works, not "
  "when the code is written', and bead-bonanza holds it at one level only. The 2026-08-15 run "
  "shipped a spec-acknowledged budget breach that no bead owned - exactly what a run-level "
  "DoD would catch.",
  ["`bd epic status` supplies the mechanical half: all children closed.",
   "Phase 4 asserts the epic's success criteria are met, each with the command that shows it.",
   "A run whose beads all closed but whose epic outcome is not demonstrated is reported as "
   "such, not as complete.",
   "`bd epic close-eligible` is used or explicitly rejected."]),
 ("p6","Add an epic-level Definition of Ready to groom-beads","advanced","M",
  "The cheaper half of the whole enumeration argument. Enumeration exists to defend against a "
  "mid-sprint parenting changing the queue underneath a run - but that is a GROOMING "
  "DISCIPLINE failure, not a tracker limitation. An epic scoped and complete before the sprint "
  "starts does not gain children mid-run; one that does was not ready to implement. "
  "`groom-beads` owns a Definition of Ready per BEAD and has none per EPIC.",
  ["An epic-level DoR exists and is enforced before a sprint may take the epic as its queue.",
   "It covers at minimum: scope closed, children complete, no open decision at the epic level, "
   "success criteria stated and verifiable, dependency graph passes `bd swarm validate`.",
   "A bead parented in mid-run is an amendment, and the skill says so.",
   "The DoR is owned where the per-bead one is owned, so the two are read together."]),
 ("p7","Wire sequencing constraints as real blocks edges at grooming time","standard","M",
  "Grooming already tags `c4_changes` and `c4_impacts` and already EMITS `sequencing: {ID} "
  "after {ID} - consumes <component>` in its wave preview - then hands it forward as advice "
  "that plan-sprint re-derives. A sequencing constraint IS a dependency: as a `blocks` edge it "
  "is durable, visible to `bd ready`, honoured by `bd swarm validate`, and re-derived by "
  "nobody. As a table row it is none of those, and was misread once - run 1786778523-19 "
  "amendment 4 records the planner treating an advisory consumes-edge as a dependency and "
  "losing a wave to it.",
  ["Grooming wires sequencing constraints as `blocks` edges and reports which it wired.",
   "A co-edit collision is NOT wired as an edge - there is no edge type for 'must not share a "
   "wave' - and stays with plan-sprint's packing.",
   "The Reason column's two collision classes mean different things: one is an edge in the "
   "graph, the other a packing rule.",
   "plan-sprint stops re-deriving sequencing."]),
 ("p8","Support multi-epic runs via epic-of-epics","standard","M",
  "Gated on S3. Enumeration allowed one run to span several epics, and a sprint epic costs "
  "that unless nesting works. If `bd ready --parent` and `bd swarm validate` traverse two "
  "levels, a cross-epic sprint is an outer epic parenting the feature epics - and a bead "
  "belongs to its feature AND to the sprint.",
  ["S3 has reported and its finding is cited.",
   "If traversal works: the multi-epic pattern is documented with the commands that support it.",
   "If it does not: an alternative is chosen (`bd mol bond`, or re-parenting) and its cost "
   "stated.",
   "Either way the skill states whether a run may span epics and how."]),
]
for k,t,tier,size,ctx,ac in PV:
    bead(k, t, body(ctx + "\n\n" + CITE, ac), pri=1 if k in ("p1","p2") else 2,
         parent="w3", tier=tier, size=size)

# ───────────────────────────────────────────────────────────── wave 4 mechanize
MC = [
 ("m1","Write bb-status.sh: the checks that need no preconditions","advanced","L",
  "Tier M-A. About 60 of the run's ~95 steps are orchestrator judgment expressed as prose, "
  "and the clearest case is the five acceptance signals: all five are pure git and `bd` "
  "queries, they run per bead every time, they have a documented failure taxonomy, they must "
  "not be reordered, and they are currently a numbered list an agent performs by hand. Run "
  "1786778523-19 recorded two orchestrator conformance failures in one sprint - the kind of "
  "failure a script does not have.",
  ["`bb-status.sh` verifies the five acceptance signals, the skip check, the landing check, "
   "report completeness and the findings-harvest non-zero assertion.",
   "Each line prints the verified state AND the command that established it.",
   "Per D1 it is run-level: the orchestrator runs it after each bead's acceptance check and "
   "at every wave boundary; a recovering session runs it first; a woken trigger runs it. A "
   "SUBAGENT NEVER RUNS IT - inside its own worktree the signals read differently by "
   "construction.",
   "Exit codes distinguish all-clear, a repairable failure and a stop-the-wave failure.",
   "The prose procedure it replaces is deleted, not left alongside."]),
 ("m2","Extend bb-status.sh with the epic-keyed checks","standard","M",
  "Tier M-B: everything that needs a stable run handle, so it waits on wave 3. Once the epic "
  "IS the run, the whole checklist can be keyed by one readable id, which is what makes a "
  "restart cheap - a recovering session does not need to be told where it is, it runs one "
  "command and is told.",
  ["`bb-status.sh <epic-id>` reports the whole run checklist.",
   "Plan structural validation: every queued bead appears in exactly one wave with tier, "
   "bucket and push mode resolved.",
   "Readiness gate as a check: `bd lint` clean, no open human item, tier metadata non-null "
   "across `bd ready --parent <epic>`.",
   "Dispatched-versus-returned reconciliation from M4's records.",
   "It distinguishes mechanical items it verified from agentic items it can only list."]),
 ("m3","Collapse recover-sprint to: run bb-status.sh, then act","standard","M",
  "`recover-sprint` is six manual steps re-deriving run state from durable sources. Once "
  "wave 3 gives it an epic id and M1/M2 give it a script, most of that is one command. The "
  "per-phase distrust table goes with the ledger (E1).",
  ["The skill's derive-the-state steps are replaced by running `bb-status.sh <epic-id>`.",
   "What remains is judgment: what to do about each state it reports.",
   "The durable trigger's message carries the epic id and the skill name, and nothing else.",
   "The skill is materially shorter and still works from a session holding no context."]),
 ("m4","Give each subagent a first-class dispatch record","standard","M",
  "The middle tier of the orchestrator -> subagent -> task hierarchy has no artifacts. A bead "
  "has a ledger, a branch, a worktree, a commit and a comment thread; a subagent has no id, no "
  "record, no lifecycle event, no way to ask which agent is running and what it holds. That is "
  "why a batch holding one worktree for six beads breaks acceptance signal 4 and needs three "
  "paragraphs to disambiguate, and why lost completion notifications need their own "
  "sub-procedure. Per D2 the record is an event bead, two per dispatch.",
  ["Two event beads per dispatch (`dispatched`, `returned`) carrying dispatch id, model, tier, "
   "worktree path, assigned beads, dispatch time and terminal state.",
   "A wisp is NOT used - wisps are not synced, so they die with the container, which is the "
   "failure being designed against.",
   "'Which agent is running and what is it holding' is a query.",
   "A standing worktree is disambiguated from the record rather than from a three-case prose "
   "rule.",
   "A lost return is a comparison between dispatched and returned, not a five-signal "
   "re-derivation.",
   "S2's cost measurement is cited."]),
 ("m5","Add a health signal with bd set-state","standard","M",
  "The one thing neither git nor bead status can express. A stuck agent, a degraded run, a "
  "wave that has stopped making progress - none of those are derivable, and all of them are "
  "what a watchdog needs to read. `bd set-state <id> health=...` writes an event bead AND "
  "maintains a lookup label in one atomic operation, and is documented for exactly this shape "
  "of use.",
  ["A health dimension is defined with a closed value set and a stated meaning per value.",
   "The orchestrator sets it; the durable trigger and any recovering session read it.",
   "A run that has stopped progressing is distinguishable from one that is merely slow.",
   "`bb-status.sh` reports it.",
   "S2's cost measurement is cited."]),
]
for k,t,tier,size,ctx,ac in MC:
    bead(k, t, body(ctx + "\n\n" + CITE, ac), pri=2,
         parent="w2" if k == "m1" else "w4", tier=tier, size=size)

# ────────────────────────────────────────────────────────────── wave 5 retro
RT = [
 ("t0","Review the 27 queued RFCs and decide each one's disposition","standard","M",
  "`rfcs/inbox/` holds 27 RFCs - 25 ready-for-review, 2 draft, 8 of them against "
  "bead-bonanza and its two helper skills. The pipeline runs findings -> report -> RFC -> "
  "triage -> beads -> a future sprint, and each hop is a human decision point. The queue is "
  "where the latency lives, and several of these RFCs propose changes this renovation "
  "supersedes outright.",
  ["Every queued RFC is classified: superseded by this renovation, still wanted, or rejected.",
   "Superseded proposals are closed with a pointer to the bead that supersedes them.",
   "Still-wanted proposals are filed as beads or explicitly deferred with a reason.",
   "`rfcs/retro-log.md` records the dispositions.",
   "The inbox is empty or every remaining item has a stated reason for waiting."]),
 ("t1","Design the retro end to end, before deciding any of its details","advanced","L",
  "The retro is currently specified in pieces across four skills and a vendored protocol, and "
  "its loop terminates in a queue: 240 findings across five sprints, 27 RFCs waiting. Design "
  "the whole flow once - capture, harvest, classify, report, decide, file - naming what each "
  "stage produces, who performs it and where the output lives. Everything else in this wave is "
  "decided INSIDE this design rather than ahead of it.",
  ["One document covers all six stages end to end.",
   "It states the audience split (D11) as its organising principle.",
   "It names, for each stage, what is mechanical and what is agentic.",
   "It supersedes or explicitly retains each piece of the current specification.",
   "T2 and T3 are decided within it, not before it."]),
 ("t2","Turn project-level findings into candidate issues a human decides on","standard","M",
  "Per D5/D11 the objective is to identify high-quality issues, report each with a summary, and "
  "let a human decide bead-or-not per issue - not to author a proposal document. The sprint "
  "report keeps every finding verbatim, because that is the substrate recurrence detection "
  "runs on and a finding filtered out once can never be observed appearing twice.",
  ["The sprint report still carries every harvested finding verbatim.",
   "On top of it the run produces N candidate issues, each with a proposed title, priority, "
   "one-paragraph summary and the finding ids it rests on.",
   "The user is asked ONCE, at the point of maximum context, which to file.",
   "Accepted ones become beads with a `discovered-from` edge and a citation to the report.",
   "No RFC is authored for a project-level finding."]),
 ("t3","Decide the finding taxonomy, inside T1's flow","standard","S",
  "Held deliberately. `retro:<kind>[:<target>]` with seven kinds - six lifted from the Findings "
  "Capture Protocol's existing trigger list plus one for the skill-and-script class it does not "
  "name - is the candidate shape, and `target` is what makes routing mechanical. But a taxonomy "
  "decided before the flow it feeds is a guess. Measured: across 240 findings the protocol's own "
  "six categories are never recorded, so every retro re-derives each finding's kind by reading it.",
  ["The taxonomy is decided within T1's design, not ahead of it.",
   "It is validated against the 73 findings in the 2026-08-15 report; the residue is counted.",
   "Capture stays 1-3 sentences, friction only, no analysis - the capturing agent is a sensor.",
   "Routing by audience (D11) is mechanical given the kind and target."]),
 ("t4","Make the audience split explicit in bead-bonanza, retro-skill and retro-spec","standard","M",
  "Per D11 the RFC path is partitioned, not retired. A finding about the USER'S PROJECT becomes "
  "a bead in that project - it has a tracker and the audience is itself. A finding about the "
  "SKILLS THIS REPOSITORY VENDS stays an RFC, because it is a proposal about a product shipped "
  "to other projects, it needs the evidence-and-proposals structure, and `triage-rfc`'s "
  "dedupe-and-defer machinery is what accumulates recurrence across projects. Today the two are "
  "conflated and a downstream project can end up carrying RFCs only this repository cares about.",
  ["Each of the three skills states which audience its findings serve and where they go.",
   "`bead-bonanza` routes project-level findings to beads and skill-level findings to the "
   "retro skills.",
   "A downstream project using the plugin never accumulates RFCs about the plugin's own skills "
   "in its backlog.",
   "`triage-rfc`'s scope is restated as cross-project proposals."]),
]
for k,t,tier,size,ctx,ac in RT:
    bead(k, t, body(ctx + "\n\n" + CITE, ac), pri=2, parent="w5", tier=tier, size=size)

# ────────────────────────────────────────────────────────────────── wave 6 split
SPL = [
 ("x2","Extract start-sprint: pre-flight, queue and the readiness gate","advanced","L",
  "Phases 0, 1 and 1.5 are self-contained, run once each, and are precisely the part a "
  "compacted orchestrator cannot see - `head -c 20000` lands at line 349, before all three. "
  "They come out as one skill matching the existing plan-sprint / recover-sprint pattern.",
  ["`start-sprint` covers pre-flight, queue resolution and the readiness gate.",
   "It returns the facts the orchestrator needs and nothing else.",
   "Its body is under the 20,000-byte cap.",
   "`bead-bonanza` invokes it by name and no longer carries its text."]),
 ("x3","Extract close-sprint: report, retro and teardown","advanced","L",
  "Phase 4 is the other self-contained, run-once block, and it is the largest single section "
  "of the file. Its retro half is being rebuilt in wave 5, so this extraction should land "
  "after that design rather than carve up a section about to change.",
  ["`close-sprint` covers the final report, the sprint retro and teardown.",
   "It reflects wave 5's design rather than the current specification.",
   "Its body is under the 20,000-byte cap.",
   "`bead-bonanza` invokes it by name and no longer carries its text."]),
 ("x4","Retire the compaction defences the split makes unnecessary","standard","M",
  "Almost every defensive mechanism in the skill exists because 34.3% of it survives a "
  "compaction: the 're-read this file from Phase 0' instruction, the paragraph explaining why "
  "the document is ordered out of numeric sequence, the published script interfaces justified "
  "by 'an agent that must open a script to learn how to call it has already been sent to read "
  "it', and the KNOWN RED exception in AGENTS.md. Once the body fits, they are scaffolding "
  "around a removed constraint.",
  ["The re-read-from-Phase-0 instruction and the out-of-numeric-order explanation are removed.",
   "The document is reordered into its natural sequence.",
   "The KNOWN RED `--strict` exception is removed from AGENTS.md's environmental-failure "
   "ledger, and the gate-running instructions that work around it are simplified.",
   "`scripts/check-skills.sh --strict` exits 0 with no warnings.",
   "Each retained defence has a stated reason that survives the split."]),
]
for k,t,tier,size,ctx,ac in SPL:
    bead(k, t, body(ctx + "\n\n" + CITE, ac), pri=2, parent="w6", tier=tier, size=size)

# ──────────────────────────────────────────────────────────────────────── edges
for b, blocked in [
  ("s1","e6"), ("s3","p8"), ("s4","e1"), ("s5","e1"), ("s6","n6"),
  ("s2","m4"), ("s2","m5"), ("s2","p2"),
  ("e1","e2"), ("e1","e3"), ("n1","e4"),
  ("e1","x1"), ("e3","x1"), ("n1","x1"), ("n2","x1"), ("x1","e7"),
  ("p1","p2"), ("p1","p3"), ("p1","p4"), ("p1","p5"), ("p1","p8"),
  ("p4","p7"),
  ("m1","m2"), ("p2","m2"), ("m2","m3"), ("e1","m3"),
  ("t1","t2"), ("t1","t3"), ("t1","t4"), ("t0","t1"),
  ("x1","x2"), ("t1","x3"), ("x2","x4"), ("x3","x4"),
  ("w1","w2"), ("w2","w3"), ("w3","w4"), ("w4","w6"), ("w5","w6"),
]:
    blocks(b, blocked)

json.dump({"commit_message": "bb-renovation: file the plan as a bead graph",
           "nodes": nodes, "edges": edges},
          open("/tmp/claude-0/-home-user-agent-plugins/f8572822-d7c1-5b91-ac81-3f56485f4d15/scratchpad/graph_part1.json","w"), indent=1)
print("part 1:", len(nodes), "nodes")

# Capability audit — every action the bead-bonanza skill family performs

**Research, not a decision.** This document inventories the *actions* the
sprint skill family tells an agent to perform, the commands behind each one,
and the traps each action's prose exists to encode. It decides nothing: it is
the input to an FRD for a `swarm` CLI, and to the simplification pass running
beside it.

It lives under `docs/research/` rather than `specs/` because it would still be
true had the CLI decision gone the other way — this repository's test for
research.

**Companion bead:** `agent-plugins-offx.1`. The FRD it feeds is
`agent-plugins-offx.2`; the component blueprint has no bead yet.

## 0. Method, and how to read a claim here

**What was read.** Every file in the ten sprint-family skills, in full:

```
$ ls plugin/skills/{bead-bonanza,run-bead,sprint-*,groom-beads}/**
14 skills · 49 files · 3,892 lines of SKILL.md · 17 reference documents
6 bundled scripts (bb-status.sh 31.6KB, pre-flight-check.sh 38.4KB,
sprint-cleanup.sh 27.4KB, bb-checks.py, bb-close-bead.sh, bb-verify-skip.sh)
2 harness hooks · 1 monitor
```

**Command frequency across that corpus**, which is the crude shape of the
surface a CLI would cover:

```
$ grep -rhoE '\b(bd|git) [a-z][a-z-]*( [a-z][a-z-]*)?' plugin/skills/{bead-bonanza,run-bead,sprint-*,groom-beads}/ plugin/hooks plugin/monitors | sort | uniq -c | sort -rn
106 bd ready       66 bd show        62 bd list       59 bd update
23 git rev-parse   23 bd create      21 git worktree list
18 bd comment      16 bd dep add     14 git log       14 bd export
13 git worktree prune  13 bd human list   12 git status
11 bd merge-slot check  10 bd swarm status   9 bd swarm validate
8 git push  8 git commit  8 bd dolt push  8 bd lint
```

**Provenance.** Every factual claim below carries one of three marks. This
document is written under `AGENTS.md`'s rule that a tool claim binds only
where somebody ran it, and most of the traps recorded here were measured by
earlier runs, not by this one.

| mark | means |
|---|---|
| **[measured here]** | the command was run in this session; the invocation and output are quoted |
| **[corpus]** | the skill corpus records it as measured, with the run that measured it. Not re-run here |
| **[read]** | a structural fact about the corpus itself, established by reading it — no tool behaviour involved |

A **[corpus]** mark is not an endorsement. It says: this is what the family
currently instructs an agent to believe. Where the FRD turns one into a
guarantee, it must be re-measured at that point.

## 1. What the inventory shows

**79 distinct capabilities.** Their disposition is the argument:

| verdict | n | meaning |
|---|---:|---|
| **MECHANIZE** | 55 | a deterministic action over `bd` and `git`; the prose exists only because no command does it |
| **DELETE** | 10 | the simplification pass removes the mechanism; nothing to mechanize |
| **MIXED** | 5 | a mechanical half and a judgment half, separable |
| **HARNESS** | 5 | spawning agents, arming monitors and triggers. Outside a CLI's reach by construction |
| **PROSE** | 4 | genuine judgment — authoring, classifying, deciding. No CLI helps |

Counted from the tables in §2 rather than asserted:

```
$ python3 - <<'EOF'   # regex over the six-column rows, bucketed by the Verdict cell
MECHANIZE 55 · MIXED 5 · DELETE 10 · HARNESS 5 · PROSE 4   total 79
rows carrying something in the Trap column: 73
EOF
```

**Only eight of the 79 are genuine judgment or harness reach.** Everything
else is a command sequence, a query, or a mechanism the simplification pass
deletes. That ratio is the case for the CLI.

**73 of the 79 carry a recorded constraint in the Trap column** — a measured
way the obvious invocation does the wrong thing silently, or a rule the prose
exists to state because nothing else states it. A trap recorded in prose is
enforced by whoever remembers to read the paragraph; the same trap inside a
subcommand is enforced by the subcommand.

**The traps cluster into five shapes**, and each shape is a CLI requirement
rather than a subcommand:

1. **Scope silently wider than the argument suggests.** `bd epic
   close-eligible` ignores its positional epic and closes every eligible epic
   in the database **[corpus]**; `bd ready --parent ""` is repo-wide rather
   than empty **[corpus]**; `sprint-cleanup`'s prune ignored its epic argument
   and tore down repo-wide (`agent-plugins-mpnf.20`, fixed).
2. **Two queries that look interchangeable and are not.** `bd list --parent` is
   direct-children-only; `bd ready --parent` is transitive but omits closed
   beads; `bd epic status` is direct-children-only **[corpus]**. None of the
   three answers "every bead in this run", which is why `bb-status.sh` walks
   the tree itself.
3. **Fail-open writes.** `bd create --graph` reports an unknown node, edge or
   plan-object field as a *warning*, drops it, and exits 0 **[corpus]** — a
   mistyped `assignee` is then a bead nobody owns.
4. **Empty-JSON shape is per-subcommand.** Section 3.1 — measured fresh here,
   and it is worse than the corpus records.
5. **Ordering enforced by prose alone.** Worktree removal must follow a
   successful merge; a shell without `&&` between them is enough to break it,
   and did (`agent-plugins-y63n.20`).

## 2. The inventory

Columns: **actor** — who performs it. **Behind it** — the commands. **Trap** —
what the prose exists to prevent. **Verdict**.

### Phase 0 · Pre-flight and environment

| # | Capability | Actor | Behind it | Trap | Verdict |
|---|---|---|---|---|---|
| A1 | Probe the environment | coordinator | `pre-flight-check.sh` (38.4KB): git, `bd`, Dolt write probe, `worktree.baseRef`, event-log writability, sandbox | Probes tracker write with a real `git push … refs/dolt/preflight-probe-$$` then deletes the ref — a reachability test inside a session holding a grant measures the grant **[read]** | MECHANIZE (already a script) |
| A2 | Read the project's declarations | coordinator | `sf.toml` → `[[checks]]`, `[models]`, `[review]`, `[run]` | A missing declaration is drafted *with* the user, never invented **[read]** | MECHANIZE |
| A3 | Capture the checks baseline | coordinator | `bb-checks.py` on the untouched tree → manifest `checks_baseline` | No opt-out: the run where a baseline is most expensive to skip is the one somebody wants to skip it on **[read]** | MECHANIZE (already) |
| A4 | Resolve commit trailers | coordinator | read `CLAUDE.md`/`AGENTS.md` → manifest `trailers` | A session URL varies per session, so it is per-run and never committed **[read]** | MECHANIZE (partly — reading conventions is PROSE) |
| A5 | Assert an Agent tool exists | coordinator | none — a harness property | Invisible from a shell. Blocking, not degradable: with no delegation nothing of bead-bonanza survives except `run-bead` **[read]** | HARNESS |
| A6 | Confirm the harness worktree dir is gitignored | coordinator | create the first worktree, then `git status --porcelain` | Checking *before* that worktree exists tests nothing **[corpus]**. The harness may place it inside the repo (`.claude/worktrees/`). **Exists ONLY under path 2 — see §2.5** | MECHANIZE, or DELETE |

### Phase 1 · Scope and admission

| # | Capability | Actor | Behind it | Trap | Verdict |
|---|---|---|---|---|---|
| B1 | Resolve a selector to a bead set | `sprint-scope` | `bd list --status all`, label/priority filters | `bd list` defaults to open issues, so a filter written without `--status all` silently under-counts a groomed backlog **[corpus]** | MECHANIZE |
| B2 | Build the run's epic shape | `sprint-scope` | `bd create` ×9, `bd dep add` epic→epic ×7, `bd update --assignee` ×8, a gate task under scope | Three forced rules: phases must be epics (`bd dep add` refuses task→epic **and** epic→task); the target must be FLAT; the gate must be a task child of scope **[corpus]** | MECHANIZE — the single most trap-laden construction in the corpus |
| B3 | Open and resolve the authorization gate | `sprint-scope` | `bd gate create -t human --blocks <gate-task>`, `bd gate resolve` | Open the gate *before* asking, so a session that dies mid-question leaves the run held rather than authorized **[corpus]** | MECHANIZE (the *asking* is PROSE) |
| B4 | Write the run manifest | coordinator | `bd show --json \| jq .metadata` → edit → `bd update --metadata @file` | **Merge, never replace** — pre-flight, grooming and admission all write the same object. The key names are a contract; a renamed key reads back `null` and dispatch proceeds without it **[corpus]** | MECHANIZE |
| B5 | Register the swarm | coordinator | `bd swarm create <target>` | Pointed at the run epic it counts ceremony epics as work and never sees a bead; `--force` mints a duplicate; `bd swarm list` reports closed molecules as Active **[corpus]** | **DELETE** — §4.6 |
| B6 | Report the admission delta | coordinator | live front vs `manifest.admitted` | A run that silently grew is the failure; one that grew and said so is the feature **[read]** | MECHANIZE |

### Phase 2 · Groom and plan

| # | Capability | Actor | Behind it | Trap | Verdict |
|---|---|---|---|---|---|
| C1 | Gate every bead for readiness | `sprint-plan` | `validate_beads.py`, `bd lint`, the Definition of Ready | Cheap read-only check first; grooming only what fails **[read]** | MECHANIZE (the mechanical half) |
| C2 | Groom what fails | `groom-beads` | authoring a work order | Authoring | PROSE |
| C3 | Bounce what grooming cannot clear | `sprint-plan` | `bd create --add-label human`, parent to a late phase | `bd update --labels human` exits 1 (`unknown flag`); it is `--add-label` **[corpus]** | MECHANIZE |
| C4 | Classify beads and compute overlap | `sprint-plan` | read each bead's `c4_changes` metadata, intersect | Judgment on top of a set intersection | MECHANIZE (the intersection) + PROSE (the call) |
| C5 | Emit and load the graph plan | `sprint-plan` → coordinator | `bd create --graph <path>` | `from_key` **waits for** `to_key` — the reverse of the natural reading, and getting it backwards inverts every dependency at exit 0. `parent_key` vs `parent_id` are not interchangeable and the wrong one orphans silently. Unknown fields are dropped with a warning at exit 0 **[corpus]** | **DELETE** — §4.2 |
| C6 | Emit and load the edge sidecar | `sprint-plan` → coordinator | `bd dep add --file <path>` | An edge naming a bead that already exists aborts the *whole* graph load, which is why the sidecar exists at all **[corpus]** | MECHANIZE (survives C5's deletion as the only plan artifact) |
| C7 | Assign every planned bead | `sprint-plan` | `bd update <id> --assignee dev-<run-id>-<n>` | Assignment IS the record; ids are derived from the run epic id, never minted, so a resumed session recomputes them **[read]** | MECHANIZE |
| C8 | Cap the dev pool | `sprint-plan` | manifest `concurrency_cap` | — | MECHANIZE |

### Phase 3 · Branch and worktree lifecycle

| # | Capability | Actor | Behind it | Trap | Verdict |
|---|---|---|---|---|---|
| D1 | Create or reuse the sprint branch | coordinator | read `manifest.branch`; else `git checkout -b`, record `branch` + `base_sha` | **The manifest read comes first** — creating a second branch forks a resumed run from the work it continues. `base_sha` is recorded at creation because that is the only unambiguous moment **[corpus]** | MECHANIZE |
| D2 | Verify `worktree.baseRef` is `head` | coordinator | read global harness config | Global user state no repository owns, so it changes under you. Unset, every worktree branches from the *default* branch and silently drops everything the sprint branch carries **[corpus]**. **Exists ONLY under path 2 — see §2.5** | MECHANIZE, or DELETE |
| D3 | Create a worktree per dispatch | coordinator | `git worktree add -b <suffix> "$REPO/../.bb-worktrees/<id>" HEAD` | Sibling of the repo, never inside it — an inside worktree reads as dirt in every `git status` **[corpus]** | MECHANIZE |
| D4 | Verify a worktree's base | coordinator | `git merge-base --is-ancestor <sprint-head> HEAD` inside it | **Currently done by asking the subagent to print two lines and report them verbatim** — a prose workaround standing in for a missing command, and for not knowing what the harness branched from **[read]**. Under path 2 it is the only check available; under path 1 the base is an argument | MECHANIZE |
| D5 | Remove a worktree and its branches | coordinator | `git worktree remove`, `git branch -d` | Must follow a *successful* merge; a chained command whose merge exited 128 removed the tree anyway (`y63n.20`). `bd worktree remove` is unusable here: it treats a branch with no upstream as unpushed, so every bead branch is refused **[corpus]** | MECHANIZE — with the ancestry guard |

### 2.5 · Who creates the worktree — an unresolved fork, and three capabilities ride on it

**[read]** `worktree-mechanics.md` describes **two** creation paths and never
says which one runs:

| | **Path 1 — coordinator creates it** | **Path 2 — harness creates it** |
|---|---|---|
| command | `git worktree add -b bb/<id> "$REPO/../.bb-worktrees/<id>" HEAD` | `Agent` with `isolation: "worktree"` |
| who picks the path | the coordinator — sibling of the repo, by rule | the harness — observed at `.claude/worktrees/`, *inside* the repo |
| what it branches from | `HEAD`, as an argument, read at creation | the global `worktree.baseRef` setting |
| the agent's cwd at start | the main checkout — it must be told the path | the worktree itself |

**Three capabilities exist only under path 2.** A6 (gitignore the harness's
directory) is entirely path-2 — under path 1 the worktree is a sibling and
cannot be dirt. D2 (verify `worktree.baseRef`) likewise: it is a **blocking**
pre-flight condition today, gating the whole run on *global user state no
repository owns*, and under path 1 the base is an argument in the command with
nothing to verify. D4 shrinks: the "have the subagent print two lines and
report them verbatim" dance is standing in for not knowing what the harness
branched from.

**Path 1 also interacts with §3's dispatch-by-assignee treatment**, and this
was not stated when that change landed: `run-batch` derives its worktree from
`git rev-parse --show-toplevel`, which is correct **only under path 2**, where
the agent starts inside the tree. Under path 1 it starts in the main checkout,
so `{WORKTREE}` returns to the prompt as a fourth value.

**Recommendation: path 1.** One prompt value bought back in exchange for
deleting a blocking pre-flight gate that depends on a setting the repository
cannot see or fix, plus a deterministic worktree path the CLI can compute
rather than discover. It also makes create/verify/remove one subcommand family
with the base as an argument.

### Phase 4 · Dispatch

| # | Capability | Actor | Behind it | Trap | Verdict |
|---|---|---|---|---|---|
| E1 | Re-query the ready front | coordinator | `bd ready --parent <epic> --exclude-type=epic --exclude-label human` | `--exclude-label human` belongs on the DISPATCH query and nowhere else — on a completion count it reports a run drained while a human still owes it an answer **[corpus]** | MECHANIZE |
| E2 | Close eligible ceremony epics | coordinator | `bd epic close-eligible --dry-run`, intersect by hand, close explicitly | Takes **no** positional epic; a passed id is silently ignored, so the bare command closes every eligible epic in the database including other runs' **[corpus]**. Omitting it is a silent deadlock, not a slowdown | MECHANIZE — the corpus's most dangerous hand-scoping |
| E3 | Derive dispatch ids | coordinator | `dev-<run-id>-<n>`, `<role>-<run-id>` | Batch numbers never restart within a run; the `<run-id>` segment is not optional because batch ids restart every run while the tracker is shared **[corpus]** | MECHANIZE |
| E4 | Resolve every prompt placeholder | coordinator | manifest reads + per-bead reads → a filled prompt file | An unsubstituted placeholder is the most common dispatch defect **[corpus]**; a dispatch prompt is an authoring surface with no reviewer, and one named two different owned-file sets (`y63n.23`) | MECHANIZE — render the prompt, do not hand-fill it |
| E5 | Resolve a bead's tier to a model | coordinator | `metadata.execution_suggested_model` else `metadata.tier` → `sf.toml [models]` | Read the documented key first and fall back; a named `model` on a bead overrides the tier **[corpus]** | MECHANIZE |
| E6 | Spawn the subagent | coordinator | `Agent` tool, `isolation: worktree`, `model:` | — | HARNESS |
| E7 | Circuit breaker | coordinator | count consecutive blocked/failed returns | Three stops the RUN, not the batch — the next dispatch meets the same systemic fault **[read]** | MECHANIZE |
| E8 | Verify a skip claim | coordinator | `bb-verify-skip.sh` in the bead's worktree | Exit 1 with status `skipped` blocks the close **[read]** | MECHANIZE (already) |
| E9 | Acceptance-check a bead | coordinator | `bb-status.sh bead <id> --worktree <p> [--pre-merge]` | `--pre-merge` defers signals 3 and 4, which fail by construction before a merge; **re-running without it afterwards is the only thing that checks the merge and removal happened** **[read]** | MECHANIZE (already) |
| E10 | Decide an agent has drained | coordinator / hook | `bd ready --assignee <id>` **and** `bd list --assignee <id> --status in_progress`, both empty | `bd ready` excludes `in_progress`, so the ready query alone reports a live agent as finished **[corpus]** | MECHANIZE |

### Phase 5 · The per-bead ceremony

| # | Capability | Actor | Behind it | Trap | Verdict |
|---|---|---|---|---|---|
| F1 | Claim a bead atomically | dev agent | `export BEADS_ACTOR=<id>`, `bd update <id> --claim` | `--status in_progress` sets status but leaves the bead claimed by nobody. With `BEADS_ACTOR` unset an *unassigned* bead's assignee silently becomes git `user.name` at exit 0 **[corpus]** | MECHANIZE |
| F2 | Run the declared checks | dev agent | `bb-checks.py [ids…]` | Narrowed is a fast loop, never the gate; run with no arguments before the close **[read]** | MECHANIZE (already) |
| F3 | Compare against the baseline | dev agent | read this run's output against `manifest.checks_baseline` | The bar is "no findings beyond the baseline", not "exit 0". **"Compare, do not parse" is the instruction — i.e. the family has no mechanism for this** **[read]** | MECHANIZE |
| F4 | Record findings | dev agent | `bd comment <id> "retro:<kind>[:<target>] …"` | `kind` is one of nine and it is what routes. A comment spelled `retro:incident:` routes nowhere and is invisible to every sweep — it happened to the most consequential finding of run 3 **[corpus]** | MECHANIZE — validate the kind at write time |
| F5 | Close and commit as one act | dev agent | `bb-close-bead.sh <id> <pri> "<title>"` | Verifies the close against the **database** before staging; stages `git add -A -- ':!.beads/*.jsonl'`; a hand-rolled replay that excludes all of `.beads/` silently drops tracked tracker config **[corpus]** | MECHANIZE (already) |
| F6 | Push and prove it landed | dev agent / coordinator | `git push`, then `git fetch && git log <remote>/<branch>..HEAD` must be empty | A rejected push leaves every other signal reading "finished". A credential refused for `.github/workflows/**` rejects the WHOLE push **[corpus]** | MECHANIZE |
| F7 | Blocked and skipped paths | dev agent | `bd update --status blocked --append-notes`; leave the tree dirty | The worktree IS the durable record; stashing or committing destroys the signal **[read]** | MECHANIZE |
| F8 | File discovered work | dev agent | `bd create` + a `discovered-from` link | "Not this bead's to close" is a reason to FILE it, not to leave it unfiled **[read]** | MECHANIZE |

### Phase 6 · Merge and accept

| # | Capability | Actor | Behind it | Trap | Verdict |
|---|---|---|---|---|---|
| G1 | Hold the merge slot | coordinator-as-merge-agent | `bd merge-slot acquire/release/check` | `--wait` registers but does not block; release promotes nobody; acquire is not idempotent; **release does not check the holder, so the remedy for a dead agent's slot and the way to corrupt a live merge are the same command** **[corpus]** | **DELETE** — §4.1 |
| G2 | Schedule a batch's merge | planner | a merge **bead**: a task labelled `merge`, assigned to `merge-<run-id>`, fanned in by sidecar edges | Its fan-in cannot live in the graph plan, because a graph edge naming an existing bead aborts the whole load **[corpus]** | **DELETE** — §4.1 |
| G3 | Merge a dispatch's branch, **conflict handling included** | coordinator | `git merge --ff-only <suffix>`; for the 2nd+ concurrent branch, `git rebase <sprint-head>` then ff; on conflict, record a `sequencing` finding, mark the bead blocked, stop dispatching | **A `--ff-only` refusal means opposite things in the serialized and concurrent cases, and reading it wrong stops a healthy run** **[corpus]**. Read the merge target at merge time, never earlier. **G4's mechanical half is joined here**: naming which refusal case you are in, detecting the conflict and recording it are the same act as attempting the merge, and splitting them is what let a removal run after a merge exited 128 | MECHANIZE |
| G4 | **Resolve** a rebase conflict | a spawned agent | reconcile both changes in the conflicted worktree, run the declared checks, hand the branch back | Resolution is explicitly **not** the merging agent's — it holds no bead's context **[read]**. This is the judgment half only; the detection and the record are G3 | HARNESS + PROSE |

### Phase 7 · Supervision, health, liveness

| # | Capability | Actor | Behind it | Trap | Verdict |
|---|---|---|---|---|---|
| H1 | Assert run health | coordinator | `bd set-state <epic> health=<v> --reason` | **A signal whose only writer is the component it monitors reports "no failure" and "no monitor" identically.** It read `healthy` through a five-hour stall **[corpus]** | **DELETE** — §4.3 |
| H2 | Derive run health | — | worktree count + last-commit age + ready-front depth | The derived signal already existed and was already correct; the parenthetical in `bb-status.sh run` was doing the work and the word in front of it was the part that lied **[corpus]** | MECHANIZE — `offx.11` |
| H3 | Stream tracker events | coordinator | install `.beads/hooks/on_{create,update,close}` writing to `$BB_EVENT_LOG` | **`bd update --claim` emits no event**, and it is the claiming primitive every worker uses — so a bead being picked up is invisible and the first line for a bead is usually its close. `bd` runs hooks fire-and-forget and ignores exit status, so a failed append is silent **[corpus]** | **DELETE** — §4.5 |
| H4 | Watch in-session | coordinator | `Monitor`, `persistent`, explicit `timeout_ms`, `TaskStop` | Default timeout is 5 minutes, shorter than most runs **[corpus]** | HARNESS |
| H5 | Arm a durable out-of-session check-in | coordinator | a recurring trigger, else a re-armed `send_later` | **The one-shot armed for run 3 fired before the stall and disabled itself**; what resumed the session was the owner's PR comments **[corpus]** | HARNESS (the *content* of the wake is a CLI call) |
| H6 | Hold the session to the ready loop | harness | `Stop` hook `bb-stop-ready-loop.sh`, exit 2, ≤8 consecutive blocks | Arming writes a file outside the repo and **cannot** be an env var — a hook is spawned by the harness, so nothing you export reaches it. `bd ready --parent ""` is repo-wide, so a hook falling back to an unset variable blocks on unrelated work forever **[corpus]** | MECHANIZE (already a hook; it should call the CLI) |
| H7 | Hold a worker to its own queue | harness | `SubagentStop` hook, exit 2, ≤3 blocks per agent | A BLOCKED bead must be *accepted* or the worker is trapped forever with no way to satisfy the refusal **[corpus]** | MECHANIZE (already) |
| H8 | Decide the run is done | coordinator | `bb-status.sh run <epic>` | Walks descendants itself because all three obvious `bd` queries answer a different question **[corpus]** | MECHANIZE (already) |

### Phase 8 · Review loop

| # | Capability | Actor | Behind it | Trap | Verdict |
|---|---|---|---|---|---|
| I1 | Resolve the reviewer | `sprint-review` | `sf.toml [review]` → declared skill/command; else a session skill; else a bare agent | Undeclared is not a skip. The report must name which of the three ran, because they are not the same evidence **[read]** | MECHANIZE |
| I2 | Scope the diff | `sprint-review` | `manifest.base_sha`..sprint head | `base_sha` is read off the manifest, never a recomputed merge-base — the base branch moves **[corpus]** | MECHANIZE |
| I3 | File each finding as a bead | `sprint-review` | `bd create` against the finding's origin | **Line numbers rot before dispatch** — two of six had moved by the time `mpnf.18` was dispatched. The durable anchor is the quoted string **[corpus]** | MECHANIZE — refuse a bare `file:line` anchor |
| I4 | Classify in or out of authorized scope | `sprint-review` | `manifest.standing_auth.admit_review_findings` | The row has never been chosen by anybody across three runs and was overridden in the two most recent **[corpus]** | PROSE + criteria — §4.7 |
| I5 | Count passes and decide to loop | coordinator | `manifest.review_passes` | Durable in the manifest because a resumed run that restarted the counter never exhausts its budget **[read]** | MECHANIZE |

### Phase 9 · Retro and report

| # | Capability | Actor | Behind it | Trap | Verdict |
|---|---|---|---|---|---|
| J1 | Harvest every finding | `sprint-retro` | `bd export \| jq -r '.comments[]?.text'`, `bb-status.sh harvest` | **Not `bd list`** — it defaults to open issues and a drained run is almost entirely closed, so a literal reading drops nearly everything **[corpus]**. A zero total is a broken pipeline, not a flawless run | MECHANIZE |
| J2 | Route each finding by kind | `sprint-retro` | a nine-row lookup, applied without reading content | Two kinds never become issues (`proof`, `environment`). An unknown kind routes to UNROUTED — a count and a list, no audience **[read]** | MECHANIZE |
| J3 | Reason findings into candidates | `sprint-retro` | — | Judgment | PROSE |
| J4 | Put candidates to a human, once | `sprint-retro` | — | Judgment | PROSE |
| J5 | File what the human accepts | `sprint-retro` | `bd create`, or a file in a configured directory | **Destinations are hard-coded as literal strings** (`rfcs/sprints/`, `.sf/outbox/`) with no key to redirect them; run 3 had to override its own skill in prose twice **[corpus]** | MECHANIZE — §4.4 |
| J6 | Write the findings document | `sprint-retro` | a Markdown file + `bd update --set-metadata findings_path` | — | MECHANIZE (the path) + PROSE (the prose) |
| J7 | Write the run report | `sprint-report` | incidents grep, results table, routing evidence, pushes, stashes, validation, triggers | **Nearly every section is a query.** §4.4 | **DELETE the skill**, MECHANIZE the sections |

### Phase 10 · Post-flight

| # | Capability | Actor | Behind it | Trap | Verdict |
|---|---|---|---|---|---|
| K1 | Prune drained agents' worktrees | coordinator | `sprint-cleanup.sh <epic>` | Leaves anything ambiguous standing and exits 1 naming it. **Do not force it** **[read]** | MECHANIZE (already) |
| K2 | Release the merge slot | coordinator | `bd merge-slot release` | — | **DELETE** with G1 |
| K3 | Close the swarm molecule | coordinator | `bd swarm close` | — | **DELETE** with B5 |
| K4 | Set terminal health | coordinator | `bd set-state` | — | **DELETE** with H1 |
| K5 | Push the tracker | coordinator | `bd dolt push` | The push, not a file, is what makes a closure durable **[read]** | MECHANIZE (already) |
| K6 | Cancel the durable trigger | coordinator | a harness action | One left armed wakes an agent into a finished sprint, which trains the reader to ignore it **[read]** | HARNESS |

### Phase 11 · Recovery

| # | Capability | Actor | Behind it | Trap | Verdict |
|---|---|---|---|---|---|
| L1 | Derive the run from the branch | `sprint-recovery` | `git rev-parse --abbrev-ref HEAD`, `git log <default>..HEAD` — every close commit is `<id> P<n> <title>` | **There is no run directory to find, and that is the point** **[read]** | MECHANIZE |
| L2 | Read the manifest | `sprint-recovery` | `bd show <epic> --json` | Carry `review_passes` forward; do not restart it **[read]** | MECHANIZE |
| L3 | Classify every unfinished bead | `sprint-recovery` | `git worktree prune && git worktree list`, then bead status × worktree presence | **A pure function of two inputs, written as a five-row table for a human to apply** **[read]**. `git worktree prune` first is not optional: a reclaimed container leaves directories deleted but still registered, and a stale entry reads as in-flight forever | MECHANIZE — §4.8 |
| L4 | Check the branch landed | `sprint-recovery` | `bb-status.sh branch` | An orchestrator that died between committing and pushing leaves **every other signal reading "finished"** **[corpus]** | MECHANIZE (already) |
| L5 | Finish an outstanding close | `sprint-recovery` | `bb-close-bead.sh` from inside the worktree, honouring any `caveat:` comment | If the script cannot run, fix why — do not replay its steps by hand **[read]** | MECHANIZE (already) |
| L6 | Decide whether partial work is finishable | `sprint-recovery` | dispatch an agent to read the ACs against the diff | Green checks say the repository is not broken; they say nothing about whether the bead is done **[read]** | PROSE + HARNESS |

## 3. Cross-cutting requirements the traps imply

These are not subcommands. They are properties every subcommand needs, and
each one is derived from a trap in §2 rather than proposed on taste.

### 3.1 `bd`'s JSON is not uniform, and it is worse than the corpus records

**[measured here]**, this session, `bd version 1.2.2 (dev)`, against this
repository's own database:

```
$ bd ready   --parent nonexistent-xyz --json   -> []      exit 0   jq type: array
$ bd list    --parent nonexistent-xyz --json   -> []      exit 0   jq type: array
$ bd dep list agent-plugins-offx     --json    -> []      exit 0   jq type: array
$ bd human list                      --json    -> [ … ]   exit 0   jq type: array
$ bd blocked                         --json    -> [ … ]   exit 0   jq type: array
$ bd gate list                       --json    -> null    exit 0   jq type: null
$ bd swarm list                      --json    -> { … }   exit 0   jq type: object
$ bd stats                           --json    -> { … }   exit 0   jq type: object
$ bd merge-slot check                --json    -> { … }   exit 0   jq type: object
```

Four distinct empty/container shapes across nine subcommands, all at exit 0:

- **`bd gate list --json` returns bare `null` when there are no gates.** This
  is a *new* instance of the hazard, not the one the corpus records — the
  corpus's example is `bd human list`, which returned an **array** here
  because this database has human beads.
- **`bd swarm list` and `bd stats` return objects**, so `jq length` counts
  *keys*. Measured: `bd swarm list --json | jq length` → **2**
  (`schema_version`, `swarms`), which is not the number of swarms.
- **Null appears inside an object too**: `merge-slot check` returns
  `"waiters": null`, not `[]`, when nobody is queued.
- `jq 'length'` on `null` returns **0**, so the null case fails *quietly* on a
  count and loudly on an iteration — the worst combination.

The corpus's own defensive rule is *"count only a JSON array and treat any
other payload as empty"*. **That rule is wrong for `swarm list` and `stats`,
which are legitimately objects** — it would report every run as absent.

**Requirement.** Every `bd` call goes through a typed wrapper that declares the
expected shape per subcommand and normalizes `null`, absent and `[]` to the
same empty value — never a generic "is it an array" guard. The wrapper is the
only place that knows a subcommand's shape, so a new `bd` version that changes
one changes one file.

**Second requirement, also measured here:** `bd` writes unrelated warnings to
**stderr** while JSON goes to stdout —

```
$ bd ready --parent nonexistent-xyz --json >out 2>err
$ cat err
warning: beads.role not configured (GH#2950).
```

so the wrapper must never merge the streams. This is the same family as
`AGENTS.md`'s "pipes hide exit statuses": a wrapper that reads status through
a pipe, or JSON through a merged stream, inverts its own verdict.

### 3.2 Every subcommand explains itself

The owner's requirement, and the corpus already half-implements it:
`bb-status.sh` prints "the verified state AND the command that established it"
on every line, so a reader can re-run any single check. **Generalize it.**

- `--explain` prints the exact `bd`/`git` invocations the subcommand would run,
  in order, with arguments resolved, and exits without running them.
- Normal runs print the same list as they execute, so an agent reading a
  failure sees which step produced it.

This is the direct remedy for `agent-plugins-offx.3`: a claim that is a test
does not need `AGENTS.md`'s "measure at the point of use" section, and an
agent that can print the command does not have to be told to measure it.

### 3.3 Steps compose, and composition is where traceability lives

The owner asked how to string commands and transforms together in Go. The
traps say what the composition must carry:

| wrapper | why, from §2 |
|---|---|
| **ordering guard** | D5/G3: worktree removal must not run when the merge before it failed. A pipeline whose step N does not run when step N−1 failed is the structural fix for `y63n.20`; a chained shell command is not |
| **timing** | H2: health is derived from *how long since* something happened. Per-step duration is the raw material |
| **audit log** | every step's invocation, exit status, stdout/stderr digest and duration, appended to a run-scoped file. This is what a recovering session reads instead of a summary |
| **dry run** | §3.2 — `--explain` is the pipeline rendered rather than executed, so it costs nothing extra once steps are values |
| **idempotence declaration** | G1: `bd merge-slot acquire` is not idempotent, and a failed release leaves your own next acquire failing. A step that declares whether re-running is safe is what makes recovery mechanical |

### 3.4 Identity is derived, never minted

`<run-id>` is the run epic id with the project prefix dropped; `dev-<run-id>-<n>`
never restarts its counter within a run. **A resumed session must recompute
the same strings** — that is the whole reason they are derived — so the
derivation belongs in one function, not in a table three skills cite.

### 3.5 Scope is an argument, not a convention

E2 and B5 are the same defect: a `bd` subcommand whose scope is wider than its
argument suggests. **Every CLI subcommand that mutates takes an explicit run
scope and refuses to act outside it**, whatever the underlying `bd` command
does with its positional.

### 3.6 Anchors are strings, not line numbers

I3, and the owner's §4.5 comment. **A finding or work order cites a path plus a
quoted string**; a line number may accompany it, marked as a hint. The write
path refuses a bare `file:line`. Class, function, file and component names are
durable; positions are not.

### 3.7 Never name a tool the environment lacks

Four findings in run 3 hit one acceptance criterion naming `shellcheck`, which
is not installed and is in no declared check. **Any command that writes an
acceptance criterion resolves every executable it names against `PATH` or
against the declared check ids, and refuses otherwise.**

## 4. What the audit says to DELETE rather than mechanize

**Ten of the inventory rows**, plus four mechanisms too small to have a row of
their own (§4.9). Each is a mechanism whose cost is visible in §2 and whose
benefit is not.

### 4.1 The merge slot and the merge bead (G1, G2, K2)

**The corpus already says the slot is not what serializes merges:** *"Only the
merge agent merges, so within one run the serialization already holds by
construction — the slot is not what makes merges one-at-a-time."* Its stated
value is *visibility*. Against that it carries five measured hazards, one of
which is that **release does not check the holder**, so the remedy for a dead
agent's slot and the way to corrupt a live merge are the same command. It is
also repository-wide and branch-unaware, so two unrelated runs contend.

The merge **bead** exists only to schedule the merge in the graph. Its fan-in
cannot be expressed in the graph plan at all, which is the entire reason the
edge sidecar was invented.

**Replacement:** the coordinator merges each dispatch's branch as that dispatch
returns, one at a time, because it is a single agent and cannot do two things
at once. On a conflict it spawns one agent to resolve — which is also the
corpus's own rule (*"resolving the conflict is NOT the merging agent's"*),
mechanized instead of narrated. Deletes `agent-plugins-offx.12`,
`agent-plugins-w7x8`, `mpnf.15`, and the merge-bead half of `mpnf.19`.

### 4.2 The `bd create --graph` plan (C5)

**[read]** The work beads already exist before planning begins — the corpus's
own worked example says so: *"The work beads were admitted before planning
began, so their edges, their assignees and their collision marks are all
sidecar-side work."* So the plan's `nodes` array creates **only the ceremony
epics, the target, and the merge beads** — and `sprint-scope` already creates
the first two.

**Delete the merge beads (§4.1) and the plan has nothing left to create.**

What survives is the sidecar: collision edges between beads that already
exist, which is `bd dep add`. That removes the reverse-orientation trap, the
`parent_key`/`parent_id` trap, the unknown-field fail-open trap, three
validators, a worked example, and roughly 160 lines of
`sprint-plan/references/run-record.md`. It also retires `agent-plugins-p567`
(the orchestrator half of the sidecar mechanism, P1 and never shipped),
`mpnf.23` and `agent-plugins-fvnp` outright — they describe a document that
would no longer exist.

### 4.3 Asserted health (H1, K4)

Replaced by derived health (H2). The corpus's own account: the derived signal
already existed, was already correct, and was printed in parentheses *beside*
the asserted word that lied. Deletes the five-value vocabulary, the
`bd set-state` calls at every re-query, and `agent-plugins-mpnf.1`-style event
beads. Satisfies `agent-plugins-offx.11`.

### 4.4 The `sprint-report` skill (J7)

The owner's call, and the audit agrees on the evidence. Compare the run-3 pair:
the findings document (`…-mpnf.md`) already carries the counts, the routing
table, the per-kind appendix of every finding verbatim, the environmental
ledger entries and what still needs a decision. The report adds: a results
table (`bd list` + `git log`), routing evidence (manifest reads), pushes
(`git rev-parse` ×3), stashes (`git stash list`), validation (`bb-checks.py`),
triggers (a harness query), and a prose reframing of what the findings already
say.

**Every one of the additions is a query.** So: the retro's findings document is
the run's report, and the mechanical sections become one command that emits
them — `swarm report <epic>` — for the retro to include or the reader to run.

Its destination becomes project config rather than a literal string in a skill
(the owner's §2 comment, `agent-plugins-y63n.19`): `docs/reports/` for this
repository, keyed in `sf.toml`. **Nothing writes to `.sf/outbox/` for this
repository's own plugins.**

### 4.5 The bd event-hook stream (H3)

`bd update --claim` emits no event, and it is the claiming primitive every
worker uses — so the stream cannot show a bead being picked up, only closed.
The hooks are fire-and-forget with ignored exit status, so a failed append is
silent. `.beads/hooks/` is also `bd`'s git-hook directory and `bd init`
rewrites it.

**Three failure modes, all silent, for a stream whose one event type is
already visible in `git log`.** Replace with `swarm watch <epic>` — a polling
command over derived state that a Monitor can tail, which needs no hooks, no
`BB_EVENT_LOG`, and no install step. This is also the honest place to build the
heartbeat the owner's §4.2 comment asks for.

### 4.6 `bd swarm` registration (B5, K3)

Three subcommands, three recorded traps. **`swarm create` takes the TARGET
epic**, never the run epic — pointed at the run epic it counts the ceremony
epics as the work and never sees a bead, which is a silent wrong answer rather
than an error. `swarm list` reports closed molecules under "Active Swarms".
`swarm validate` warns about an outside dependency and then computes readiness
as if the edge were absent, and is one level deep so a nested epic hides its
leaves. The corpus's own instruction is
to **suppress** the preview in both cases where it would be consulted.

Its only remaining value is discoverability to a session that has nothing —
and L1 already solves that from the branch. Delete it.

### 4.7 "Incident" as a construct

The owner's §3 comment, and the evidence supports it strongly. An incident is
currently a *severity* that must be spelled inside a `retro:<kind>:<target>`
comment as the literal word ` incident `, mirrored onto the epic as a second
`amend:` comment, matched by a fragile grep, and reported first. Run 3's
counting problem is the proof: **three different counts (8, 11, 17) and none of
them wrong**, plus one record — the run's most consequential — written as
`retro:incident:` and routed nowhere, because the protocol shows the right form
and never states the wrong one.

**Replace with: an ordinary finding.** The nine kinds already carry it
(`decision`, `scope`, `gap`, `sequencing`), the routing already works, and a
finding that changed what "correct" means is legible from its own text. That
deletes the dual-record rule from `run-bead`, the incidents-first section from
the report, the grep, `agent-plugins-sdu1`, and one of the three human channels'
worth of prose from `run-record.md`.

### 4.8 The `sprint-recovery` skill (L1–L5)

The owner's question, and the answer is yes with one condition. **L3 — the
five-row classification — is a pure function of bead status and worktree
presence.** L1, L2, L4 and L5 are already commands or one query each. What is
left after mechanizing them is L6, a judgment call that needs an agent.

**Condition:** the durable trigger's prompt currently says *"invoke the
`sprint-recovery` skill"*, and it says a skill name rather than a path
precisely because a fresh session cannot resolve a path into a bundle it never
loaded. **A CLI binary on `PATH` is resolvable the same way a skill name is** —
so the trigger becomes "run `swarm status <epic>` and resume under
`bead-bonanza`". The skill can only be deleted *after* the CLI exists; until
then it stays and shrinks.

### 4.9 Four smaller ones

- **Push modes** (`push_per_bead` / `defer_push`). With the coordinator merging
  as dispatches return, subagents never push. One authorization row and two
  branches of prose in `run-bead`, `sprint-batch` and both prompt files go.
- **The deferred human channel** — `agent-plugins-offx.13` already proposes
  parking it. It rests on four measured `bd` quirks including two subcommands
  (`bd human respond`, `bd human dismiss`) that **both** error with "storage is
  nil", and one (`bd human list`) that keeps listing a closed bead forever.
- **Two prompt files.** `prompt-batch.md` and `prompt-per-bead.md` are the same
  prompt; a batch of one is a bead dispatched alone. One file, one placeholder
  set — and §3.2 says it should be rendered by a command, not filled by hand.
- **One branch per bead.** §5.

## 5. One branch per bead is redundant with one commit per bead

The owner's question, and the corpus answers it against itself.

**What the corpus claims the per-bead branches buy:** *"the branches keep the
beads separate"* (`worktree-mechanics.md`), *"it is the* branches *that keep
the beads separate"* (`sprint-batch`).

**[read]** That cannot be right. A batch is one agent working serially in one
worktree; the beads are separated by being *committed* one at a time, and the
ceremony already guarantees exactly one commit per bead. The branches are cut
in sequence off each other, so branch N contains every commit before it — which
is why **the merge is already one per dispatch**, fast-forwarding the last
branch, and why every intermediate branch is deleted immediately afterwards.

**What the per-bead branches actually buy** is one thing: the worktree→bead
mapping. `git worktree list` prints a checked-out branch, the branch is a
bead's hash suffix, so the branch names the bead and the bead's `assignee`
names the agent. Four tables in three files are built on that chain.

**One branch per dispatch, named for the dispatch id, makes that chain
shorter, not longer** — the branch names the agent directly, and the bead comes
from the commit subject (`<id> P<n> <title>`) which is already the canonical
format and already how recovery derives the completed set (L1).

**What it costs:** nothing measured. A bead cannot be merged independently of
its predecessors today either — the chained branches guarantee it.

**What it saves:** a `git checkout -b` per bead, a `git branch -d` per bead, the
"do not branch from the sprint branch for each bead" warning, the
"do not delete the branch until the fast-forward succeeded" warning, and the
suffix-derivation rule restated in five places.

## 6. What is working, what is friction, and what is not a factor

Measured this session over the whole tracker — three runs' worth of captured
findings — rather than read off any run's report:

```
$ bd export > e.jsonl
$ jq -r '.comments[]?.text' e.jsonl | grep -c '^retro:'          -> 822
$ jq -r '.comments[]?.text' e.jsonl | grep -oE '^retro:[a-z]+' | sort | uniq -c | sort -rn
   66 proof   52 filed   32 decision   30 tooling   19 gap   14 stale
   12 environment   9 scope   6 unverifiable   4 process   4 decided
   3 sequencing   2 incident   1 misleading
$ jq -r 'select(.status=="blocked") | .id' e.jsonl | wc -l       -> 1
```

**211 beads carried findings; 810 findings; mean 3.8 per bead.**

### The top reason it works: the per-bead ceremony, and nothing has gone wrong there

**Exactly ONE bead has ever taken the Blocked path**, across three runs and 211
beads that recorded findings. And the split of `tooling` findings by target is
the sharper measurement:

```
$ jq -r '.comments[]?.text' e.jsonl | grep -oE '^retro:tooling:[^ ]+' | sort | uniq -c | sort -rn
   8 bead-bonanza   4 bd   2 sprint-retro   2 scripts/check-skills.sh
   1 sprint-wave    1 sync-shared-references.sh   1 sprint-cleanup.sh
   1 scripts   1 findings-capture-protocol   1 coordinator   1 bb-status   1 agent-plugins

$ ... | grep -cE '^retro:tooling:(run-bead|sprint-batch)'     -> 0
$ ... | grep -cE '^retro:tooling:(bead-bonanza|coordinator|sprint-wave|sprint-cleanup|scripts)' -> 15
```

**Zero tooling findings against the developer tier. Fifteen against the
coordinator's machinery.** Three runs of agents claiming a bead, implementing
in scope, gating, recording findings and closing into one commit produced no
complaint about the procedure at all. That is the thing that is working, and it
is worth naming because it is the part nobody writes findings about.

Four properties do the work, and each is cheap:

1. **One worktree per dispatch** — a subagent cannot see or damage another's
   tree, so "did agent A break agent B" is not a question anybody asks.
2. **One commit per bead, with the close folded into it** — a bead is either
   closed-and-committed or neither, so there is no half-state to reconcile.
3. **Assignment IS the record.** One string is an agent's whole identity, and
   nothing is written at dispatch time to make "what is that agent holding" a
   query. There is no bookkeeping key to forget or to go stale.
4. **Verify against state, never against testimony.** Every acceptance verdict
   is read from git or the tracker, so a lost completion notification, a
   subagent that never returns, and a subagent that lies all reduce to the same
   five signals.

### The friction is entirely in the machinery around the work

Every recorded failure is coordination: routing, prompts, worktree lifecycle,
acceptance criteria, tool claims, liveness. The three that cost most, in order:

- **Liveness.** A five-hour stall with eleven beads Ready, a health label that
  read `healthy` throughout, and a one-shot recovery trigger that had fired
  before the stall and disabled itself. What resumed the session was the
  owner's PR comments. §4.3 and §4.5 are the two halves; the derived health
  signal already existed and was already correct.
- **The dispatch prompt.** It is an authoring surface with no reviewer, filled
  in by hand, and the highest-reach carrier of a false claim in the system — a
  retracted tool claim reached three of them in one run. E4 is the remedy:
  render it, do not fill it.
- **Prose-enforced ordering.** Merge-then-remove, `--ff-only`'s two meanings,
  `bd epic close-eligible`'s hand-scoping. Each is one shell character away
  from breaking, and one of them broke.

### What is not a factor in its success

- **`bd swarm` registration.** Three subcommands with three recorded traps,
  consulted for a preview the corpus itself instructs you to suppress. §4.6.
- **The merge slot.** Its own reference says the serialization it appears to
  provide already holds by construction. §4.1.
- **The bd event-hook stream.** `bd update --claim` emits no event and that is
  the primitive every worker uses, so the stream shows closes and nothing else.
  §4.5.
- **The graph plan.** It creates nothing that is not created elsewhere. §4.2.
- **Push modes.** Two branches of prose for a choice that stops existing once
  the coordinator merges as dispatches return.

### The taxonomy is leaking, and by more than the retros reported

Each run's retro reported **2** unrouted findings. Across the tracker the real
figure is higher, because a finding in an invented kind is only visible to a
sweep that goes looking for one:

```
$ jq -r '.comments[]?.text' e.jsonl | grep -E '^retro:(process|decided|incident|misleading)' | wc -l
11
```

Four inventions — `process` (4), `decided-silently` (4), `incident` (2),
`misleading-text` (1) — and **every one of the eleven was written by the
coordinator or is about a bead's own text.** None came from a developer agent
working a bead. The nine kinds are adequate for the work; they are not adequate
for describing the coordination around it, which is exactly where the friction
is. That is a finding about the taxonomy, not about the authors, and it is one
more argument for the write path validating the kind (F4).

## 7. What earns a CLI command

Five classes, and a test for what to leave alone. The first three are the
owner's; the last two are what the trap taxonomy in §1 turns out to contain
once the first three are subtracted.

### The five classes

**1. Composed workflows** — multiple commands, output parsing, and branching on
what came back. *Examples:* B2 (build the run's epic shape: nine creates, seven
edges, eight assignments, one gate, every step conditional on the last); F5 (the
close ceremony's fixed order); G3 (merge, name the refusal case, rebase, merge
again, record a conflict).

**2. Bespoke conventions** — the graph is the substrate, and a convention layered
on top of it has no meaning to `bd`. *Examples:* B4 (the run epic's `--metadata`
IS the run record, and it must merge rather than replace); J1 (findings live as
`retro:`-prefixed comments and must be harvested with `bd export | jq`, never
`bd list`); E3 (dispatch ids derived from the run epic id); D1 (`base_sha`
recorded at creation because that is the only unambiguous moment).

**3. Commands agents commonly get wrong** — undocumented, or behaving against
their documentation. *Examples:* `bd update --labels` exits 1; `bd human
respond` errors with `storage is nil`; `bd worktree remove` treats a branch with
no upstream as unpushed, making its one safety check unreachable here.

**4. Silent-failure actions — where a PERFECT agent still gets a wrong answer.**
This is class 3's harder half and deserves separating, because the remedy is
different: class 3 is fixed by knowing better, class 4 cannot be. The test:
*would an agent that did exactly the documented thing still lose?*

- `bd epic close-eligible` ignores its positional argument and closes every
  eligible epic in the database, exit 0.
- `bd create --graph` drops an unknown node, edge or plan-object field as a
  *warning*, exit 0 — a mistyped `assignee` is a bead nobody owns.
- `bd ready --parent ""` is repo-wide rather than empty.
- `bd update --claim` on an *unassigned* bead silently rewrites the assignee to
  git `user.name`, exit 0.
- `bd swarm create` pointed at the run epic counts ceremony epics as the work
  and reports `Swarmable: YES`.

**A wrapper here is not convenience — it is the only way to get a non-zero exit
at all.** Highest value per line of code in the whole inventory.

**5. Derivations — one answer that no single command gives.** Not a workflow
(no writes, no branching on side effects) but a walk or a join. *Examples:* H8
(`bb-status.sh run` walks the descendant tree itself, because `bd list
--parent`, `bd ready --parent` and `bd epic status` each answer a *different*
question and none answers "every bead in this run"); H2 (health = standing
worktrees × last-commit age × ready-front depth, a join across git and `bd`);
E10 (drained = two queries, both empty); F3 (this run's check output diffed
against the baseline).

**This is where token spend is worst.** An agent doing a derivation by hand
issues N queries and then reasons over the results in context; the same answer
as one command is a single line. It is the class that most directly serves the
mechanization hypothesis.

### Two prioritization axes, orthogonal to class

**Frequency.** A wrapper's cost amortizes over how often the action runs.
From §0's frequency count: `bd ready` 106 mentions, `bd show` 66, `bd list` 62,
`bd update` 59, `git worktree list` 21, `git worktree prune` 13. An action
performed once per run (B2, A1) is worth less than one performed per bead per
re-query (E1, E10, F3) at equal class.

**Blast radius.** Does failure *end* the run or *degrade* it? Health and the
dispatch front are liveness — the five-hour stall was a derivation nobody ran.
A missing `--explain` is an annoyance. Prefer the run-ending ones.

**Suggested order:** class 4 ∩ high frequency → class 5 ∩ liveness → class 1 →
class 2 → class 3.

### The test for what NOT to wrap

**A single command with no convention, no trap and no derivation stays a single
command.** Wrapping `bd comment` as `swarm comment` adds a layer, a version and
a thing to learn, and buys nothing. Two further exclusions, both already in the
inventory: anything whose value **is** the agent's judgment (I4 classifying
scope, J3 reasoning findings into candidates), and anything only the harness can
do (E6 spawning, H4 monitors, H5 triggers).

**A CLI that wraps everything is a second API to learn on top of the one it
hides.** The 5 HARNESS and 4 PROSE rows in §1 are the floor; the honest target
is well under the 55 MECHANIZE rows, because several collapse into one command
family (D3/D4/D5 into `worktree`; G3/G4's mechanical half into `merge`).

## 8. Open questions for the owner

1. **Who creates the worktree — path 1 or path 2?** §2.5. It decides whether
   A6 and D2 exist at all, whether a *blocking* pre-flight gate on global user
   state survives, and whether `{WORKTREE}` returns to the dispatch prompt. The
   recommendation is path 1; the audit has no evidence the harness path buys
   anything the coordinator path does not.
2. **Where does the CLI live?** This repository already has a Go module at
   `cli/` (`github.com/brettcallaghan/agent-plugins/cli`, Go 1.24, cobra) that
   builds the `spec` binary. A second binary in the same module shares the
   toolchain, the CI matrix and the check set; a separate module keeps the
   swarm CLI vendable on its own. The audit has no evidence either way.
3. **Does the CLI carry the retro's *routing*, or only its harvest?** J2 is a
   nine-row lookup — mechanically trivial — but routing is what decides who
   receives a finding, and mechanizing a decision is different from
   mechanizing a query.
4. **§4.7's replacement, concretely.** Dropping "incident" removes the report's
   lead section. What replaces it: findings of certain kinds surfaced first, or
   nothing at all?
5. **How far does `--explain` go?** Printing the `bd`/`git` invocations is
   cheap. Printing *why* a step exists — the trap it guards — is the thing that
   would let an agent troubleshoot without reading the corpus, and is a much
   larger surface to keep true.
6. **`admit_review_findings`** (`agent-plugins-offx.8`). The owner's §4.7
   comment leans toward clear criteria plus the agent's call. That is a
   decision the FRD needs, not the audit.

## 9. What this audit did not establish

- **No `bd` behaviour was re-measured except §3.1.** Every trap marked
  **[corpus]** is what the family currently instructs an agent to believe. Run
  3 retracted one such claim after it had reached three dispatch prompts, so
  the base rate of these being wrong is not zero. **Any trap the FRD turns into
  a guarantee must be re-measured at that point.**
- **Nothing here measures whether a capability is *used*.** The inventory is
  what the corpus instructs; a run may skip a step and nothing would show it
  here. §6 answers the adjacent question — where friction concentrated — from
  the captured findings, but a finding is evidence that a step went wrong, not
  evidence that it ran.
- **§6's per-bead evidence could not be joined to commits in this clone.** The
  three runs' branches were squash-merged, so no `<id> P<n> <title>` subject
  survives in this checkout's history (`git log --all --format=%s | grep -cE
  '^agent-plugins-.* P[0-4] '` → 0). The bead outcomes in §6 are read from the
  tracker alone; "one commit per bead landed" is inherited from the run
  records, not re-measured here.
- **No cost model.** "34 capabilities to mechanize" is a count, not an
  estimate; several (B2, E4, L3) are a day's work each and several are an hour.
- **The Windows target is unexercised.** Every measurement in §3.1 and every
  trap in §2 was recorded on Linux. `git worktree` path handling, `PATH`
  resolution for `bd`, and the sibling-directory convention for worktrees are
  the three places a Windows difference would land.

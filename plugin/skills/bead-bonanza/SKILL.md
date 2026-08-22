---
name: bead-bonanza
description: Process many beads autonomously by orchestrating subagents. Use
  when the user asks to work on multiple beads at once — e.g. "work on all open
  beads with label X", "address all P1 beads", "clear the bead backlog". Triages
  the queue, gates every bead for readiness (grooming under-specified beads via a
  subagent, or bouncing what can't be made ready), surfaces both explicit and
  implicit decisions up front, dispatches via the cheapest viable execution mode
  (doc fast lane / shared-surface batch / sequential), runs every bead in its
  own git worktree, streams per-bead progress, and lands each bead in a single
  commit.
version: "1.7"
---

# bead-bonanza

Batch-process beads via subagent orchestration. The orchestrator (you) stays
small — you only hold the queue, the user's decisions, the dispatch plan, and
one-line outcomes per bead. Each bead runs in a fresh subagent so context
doesn't compound.

Read `CLAUDE.md` and `AGENTS.md` once at the start so you have repo conventions
in scope before delegating.

**After any compaction, before dispatching anything:** re-read this file in
full, then invoke the **`sprint-recovery`** skill. A compacted session
re-injects only this file's first ~5,000 tokens, so the dispatch templates and
the published script interfaces are NOT in your context even though you have
"read the skill".
`sprint-recovery` re-derives run state from the durable sources (the branch,
the run epic's manifest, `git worktree list`, the tracker) rather than from a summary of
it, and a skill name survives truncation where a relative path does not. Then
re-run the queue query.

**This skill is language-agnostic.** It never hard-codes build, test, lint, or
format commands. Everything it needs to validate the work done for a bead comes
from the **checks the project declares** in `sf.toml`, run by `scripts/bb-checks.py`
in this skill's `scripts/`.
Running it is the whole interface: exit 0 means every check passed, anything
else means something failed, and the output says what. What those checks are,
and in what order, is the project's business.

## Run checklist

Copy this into your response at run start and check items off as the run
proceeds. It is the one artifact that survives a long autonomous run.
(It lives in your response, not in a task tool — nothing external to render,
nothing to fall out of sync with `bd`.)

```
bead-bonanza run:
- [ ] 0. pre-flight run, BB_* facts eval'd
- [ ] 1. `sprint-scope` invoked: run epic resolved, admission record
        (admitted / excluded / moved), authorization gate taken
- [ ] 1b. pre-flight re-run with the epic: checks baseline on the manifest
- [ ] 2. Dispatch plan: batch + tier + push mode per bead
- [ ] 2b. Sprint branch created (or reused), `branch` + `base_sha` in manifest
- [ ] 3. Execute waves via `sprint-wave`, merging each accepted branch
- [ ] 3b. Review via `sprint-review`; loop back to groom & plan when findings
        were admitted and the pass budget is not spent
- [ ] 4. Final report: INCIDENTS FIRST, table, totals, implicit decisions, TIER ESCALATIONS,
        follow-up beads, stashes, pushes, run epic id
```

**Most of that is checkable rather than assertable.** `bb-status.sh run <epic>`
verifies the mechanical half — scope and per-status counts, health set, lint
clean *within this run's queue*, no open human item, every ready bead carrying
a tier, dispatches still running, and every admitted bead closed — and then
**names the items it cannot verify** rather than leaving a clean report to
imply the whole checklist passed. Run it at every wave boundary and at the end.

Two things it works around, both measured: `bd lint` exits 0 even when it
reports warnings, so its findings are the verdict and its status is not; and it
has no scope flag, so its output is intersected with this run's queue rather
than blaming a run for the tracker's other beads.

**How this document is ordered.** The executable core comes first — worktrees
and merges, Phase 3's per-bead ceremony, the reference and script interfaces —
because only this file's opening survives a compaction, so the phases appear
out of numeric order.

## The sprint branch — you create it, once, before the first worktree

**Immediately before dispatching wave 1**, and not before. A run that bounced
every bead at the readiness gate has nothing to branch for, and a branch cut
before grooming records a `base_sha` from before the grooming commits land.

```bash
man="$(bd show "$BB_EPIC" --json | jq -r '.[0].metadata')"
br="$(printf '%s' "$man" | jq -r '.branch // empty')"
if [ -n "$br" ]; then
  git checkout "$br"                      # RESUME — reuse, create nothing
else
  br="sprint/<epic-suffix>"
  git rev-parse --verify "$br" >/dev/null 2>&1 && { echo "branch $br taken"; exit 1; }
  git checkout -b "$br"
  # merge into the manifest; never replace it
  bd update "$BB_EPIC" --set-metadata branch="$br" \
                       --set-metadata base_sha="$(git rev-parse HEAD)"
fi
```

**Reuse is not an optimization** — it is the whole of what makes a resumed run
land on the same branch. Read the manifest before you create anything.

**A name already taken by an unrelated branch stops the run**, here, before any
worktree exists. That is the cheapest place a run can stop.

**`base_sha` is recorded at creation because that is the only moment it is
unambiguous.** Deriving it later from the branch's merge-base assumes the base
branch has not moved, and on a long run it will have.

Three things consume it and none of them may create it: the review scopes its
diff `base_sha`..sprint head, every worktree branches from sprint HEAD (below),
and the pull request opens against it.

## Every bead runs in its own worktree

**This is universal, not opt-in.** Each dispatched bead gets a git worktree
named for its hash suffix (`jh6`, `92q.14` — not the full id), created from the
sprint branch and carrying a branch of the same name. Nothing is implemented in
the shared checkout except by the orchestrator itself.

**Presence is the in-flight signal.** A worktree that still exists means its
bead is not yet accepted — the orchestrator removes it as the final act of
accepting the bead, after the branch merges. `git worktree list` is the
complete answer to "what is still in flight": there is no mtime heuristic
anywhere, and no need to re-derive which paths a bead touched.

**Hard gate — every worktree must branch from the sprint's own branch, never
the default branch.** Verify `BB_WORKTREE_BASEREF` is `head` **before the first
worktree of each wave, and do NOT dispatch otherwise.** It is global user state
that no repository owns, so anything could change between waves: re-verify
before **each** wave, not once per run.

**Merge policy.** The **orchestrator** merges each bead's branch into the
sprint branch as it is accepted, one at a time; subagents never merge. A
chained batch is merged by fast-forwarding the LAST bead's branch, which
carries its predecessors — not by merging each bead's branch in turn. Either
way the result is linear history and one commit per bead. Take
`bd merge-slot acquire` before each merge and release it after — not to make
merges one-at-a-time, which the single-merger rule already does, but so a
stalled wave has a queryable answer and a resumed session cannot merge
alongside the original. What a `--ff-only` refusal means differs between a
sequential and a concurrent wave, and reading it wrong stops a healthy wave —
take that rule, and how to hold the slot, from the lifecycle named below before
you merge anything.

Where the worktree lives, how `bd` behaves inside one, the base-ref remedy, and
the create/merge/remove lifecycle are all in `references/subagent-prompts.md`,
"Worktree mechanics" — read it at dispatch, which is when the first worktree of
the run gets created.

## Execute the waves — invoke `sprint-wave`

Per wave, from the wave table: **invoke `sprint-wave`**. It dispatches a
subagent per batch, verifies every return against the tracker and git rather
than against the subagent's account, and returns you one line per bead —
`completed` / `blocked` / `skipped`, each with a sha — plus the wave's push
verdict.

**Then you merge**, one branch at a time, and remove each worktree as you
accept its bead. `sprint-wave` merges nothing: a single merger is what keeps
history linear, and a conflict is a verdict you act on rather than work you
delegate (see "Every bead runs in its own worktree" above).

**Re-query the ready front at every wave boundary** — `bd ready --parent
<epic> --exclude-type=epic`. That query is the order, not the stored table:
closed beads leave it, beads unblocked by the wave you just landed enter it,
and a bead added to the epic mid-run is in scope by construction.

**A wave that trips the circuit breaker stops the run**, not just the wave.
Three consecutive blocked or failed beads means something systemic — a broken
base, an environment fault, an under-groomed queue — and the next wave will
meet it too.


### When the run is done

**`bb-status.sh run <epic>`** — the run-level Definition of Done, computed from
state rather than from your recollection of what you dispatched. A run is done
when every bead the manifest's `admitted` list names is closed, or accounted
for as blocked-with-a-reason. Beads **excluded at admission** are reported
alongside, because a run that finished its admitted set while three beads sat
out is done over a *narrowed* scope, and must say so rather than reading as
complete over the scope the user first described.

It walks the epic's descendants rather than calling `bd epic status`, which
counts direct children only — on a nested run epic that answers about sub-epics
and read `0/2` with four leaves closed. `bd list --parent` is direct-children
only too; only `bd ready --parent` is transitive, and it omits closed beads. The
walk is the only correct count, so the script does it and says when the epic
was nested.

Exit 0 clear, 1 repairable (no admitted set recorded — the manifest was never
written), 2 stop (admitted beads still open).

### The run's health — the one thing nothing else can express

A stuck agent and a slow one look identical in git and in bead status. Neither
is derivable, and both are what a watchdog needs to tell apart, so the
orchestrator **asserts** health and the script reports the evidence for and
against the assertion.

```sh
bd set-state <epic> health=<value> --reason "<why>"
```

Closed value set — set it at run start and at every wave boundary:

| value | meaning |
|---|---|
| `healthy` | beads are closing and nothing is failing repeatedly |
| `slow` | nothing closed recently, but agents are alive — worktrees standing, commits landing |
| `stalled` | nothing closed recently **and** no evidence of a live agent |
| `degraded` | progressing, but with repeated failures or escalations |
| `blocked` | cannot proceed without a human; pair with `bd human` |

`slow` and `stalled` are the distinction the dimension exists for. The evidence
that separates them — standing worktrees and the age of the last commit — is
what `bb-status.sh run <epic>` prints beside the label, so a reader can
disagree with a stale claim rather than inherit it.

**Cheap, and measured so.** Eight transitions on one bead took 5.1s (0.64s
each), and the label **replaces** rather than accumulates: after eight changes
the bead carried exactly one. One current value, queryable, no history litter.

`bb-status.sh run` reports an unset health as repairable, because a watchdog
reading no label cannot tell a healthy run from an abandoned one.

### The live view

`bd swarm status <epic>` is the durable answer to "where is this run" —
computed from the beads, so it is correct after any crash and needs nothing
re-registered when scope changes. On a **nested** run epic it is one level
deep, reporting sub-epics rather than leaf beads: fall back to the re-queried
ready front plus a walked progress count, and say in the report which you used.

The event stream in `references/progress-streaming.md` is the other view — live,
ephemeral, and for a `Monitor` to tail. Losing it costs mid-flight visibility,
not the run's record.

### Recovery and progress streaming

Three situations, three loads — take the one you are in, not the whole corpus.
A subagent past 5 minutes or any question about the progress events:
`references/progress-streaming.md`. A subagent that crashed or timed out
mid-ceremony: `references/recovery.md`. Resuming a run whose state you do not
hold, because the *orchestrator* died or compacted: invoke the
**`sprint-recovery`** skill — a skill rather than a file, because the session
that needs it may be fresh, with no bundle loaded and no path to resolve.

Without being sent anywhere, the orchestrator arms a `Monitor` on the progress
file for any subagent expected to run past 5 minutes (any non-trivial bead,
given full test runs) so the user sees per-phase events live — armed
`persistent`, stopped with `TaskStop` when the wave drains — and, before the
first dispatch, arms the durable out-of-session trigger
`references/progress-streaming.md`'s "Back the Monitor with a durable trigger"
specifies, cancelling it in Phase 4.

### Subagent prompt templates

The dispatch prompts — per-bead (sequential/worktree), doc/chore
fast lane, shared-surface batch, plus the orchestrator's worktree
merge notes — are in `references/subagent-prompts.md`. **Read that file at
dispatch** and fill in every `{PLACEHOLDER}` before sending. The templates
now carry inputs and batch-specific steps only: the ceremony itself is the
**`run-bead`** skill, which the subagent invokes by name. A skill name
survives a paraphrase where a ten-step ordered procedure does not, which is
why the old rule against paraphrasing a template is gone rather than
restated.

## Reference files

Load the one that matches where you are in the run:

| Working on | Load |
|------------|------|
| Executing one wave: dispatching its batches and verifying the returns | invoke the **`sprint-wave`** skill |
| Working one batch of beads as the dispatched subagent | the **`sprint-batch`** skill |
| Reviewing the run's own diff and turning findings into beads | invoke the **`sprint-review`** skill |
| Dispatching an execution subagent | `references/subagent-prompts.md` |
| Grooming the queue and partitioning it into waves and batches | invoke the **`sprint-plan`** skill |
| The grooming subagent's prompt | the template that ships with `sprint-plan` |
| A subagent running past 5 minutes; arming the Monitor and the durable trigger; reading progress events | `references/progress-streaming.md` |
| A subagent that crashed or timed out mid-ceremony | `references/recovery.md` |
| Resuming a run whose state you do not hold (the orchestrator died or compacted) | invoke the `sprint-recovery` skill |
| A project with no declared `[[checks]]`, `[models]` lineup or `[review]` reviewer | `references/project-setup.md` |
| Resuming a bead interrupted mid-implementation (its worktree still exists) | `references/operating-principles.md` |
| Pre-flight blocked and you need the remedy | `references/pre-flight.md` |
| Resolving the run's scope and taking the authorization gate | invoke the **`sprint-scope`** skill |
| Verifying a bead's acceptance signals, a wave's landing, the run's Definition of Done, or whether a held merge slot is safe to release | run `scripts/bb-status.sh` |
| Checking a `skipped` claim against the tree | run `scripts/bb-verify-skip.sh` |
| Closing a bead: close + verify + stage + commit + push the tracker | run `scripts/bb-close-bead.sh` |
| The Definition of Ready the readiness gate enforces | `references/shared/work-order-definition-of-ready.md` |

The reference files are one hop from here — read them in full rather
than previewing.

The `scripts/` entries are **executed, not read**. Each encodes an ordering,
quoting or portability rule that is easy to state and easy to get wrong under
pressure. Run them rather than reproducing what they do.

Because they are not to be read, their interfaces are published here — an
agent that must open a script to learn how to call it has already been sent
to read it:

```
pre-flight-check.sh [<epic-id>] [--self-test]                   # sandbox ON
bb-verify-skip.sh [pathspec ...]        # no pathspec = whole working tree
bb-close-bead.sh [--trailers-file <path>|--trailers-from <epic>] <bead-id> <priority> <title> [path ...]
```

- **`scripts/pre-flight-check.sh`** — the environment probe. Human report on
  stderr, an evalable `SCRIPTS=`/`BB_*=` block on stdout. Exit 0
  proceed / 1 blocking / 2 bad invocation. **Run it with the sandbox ON** — the
  one write the sandbox refuses is the global Dolt telemetry config, which the
  script detects, skips, and reports as the single command to re-run with the
  sandbox off. It always runs the project's declared checks on the untouched
  tree, so a pre-existing failure is recorded as the baseline instead of being
  charged to the first bead — there is no opt-out, because the run where a
  baseline is most expensive to skip is exactly the slow one somebody would
  want to skip it on. `--self-test` exercises the portability wrappers and
  takes no run id.
- **`scripts/bb-status.sh`** — read-only; repairs nothing. `bead` verifies the
  five acceptance signals in order; `wave` reports tree cleanliness, the
  landing check (unpushed commits against the upstream), surviving worktrees
  and the merge slot's holder; `harvest` asserts a run's `retro:` total is
  non-zero, since a zero total is a broken pipeline rather than a flawless run.
  Exit 0 clear, 1 repairable, 2 stop-the-wave, 3 invocation error. **Run-level:
  a subagent never runs it** — inside its own worktree signal 4 reads
  differently by construction.
- **`scripts/bb-verify-skip.sh`** — the two-sided skip check. With no argument it
  examines the whole working tree; pass pathspecs to scope it.
- **`scripts/bb-close-bead.sh`** — the close ceremony, in its required order.
  `--trailers-file <path>` / `--trailers-from <epic>` are optional, mutually
  exclusive, and must precede `<bead-id>`. `--trailers-from` reads the
  manifest's `trailers` key off the run epic; the file form's
  contents are appended to the commit body after a blank line, so a project's
  required trailers (`Co-Authored-By`, a session URL) land in the one commit
  this script makes and no caller needs `git commit --amend` to add them.
  Omitting it appends no trailers.
  `<path ...>` are the code files this bead changed; WHEN omitted it stages
  everything currently modified, so pass explicit paths whenever another
  agent may be sharing the tree. It closes with a status-only `bd update` and
  never writes the description. It then **verifies the close against the
  database** — `bd show <id>` must read `closed` — and refuses to stage or
  commit anything when it does not, so a `bd update` that silently did not
  take cannot produce a commit claiming a closure that never happened. After
  the commit it runs `bd dolt push`; a push failure is a WARNING, not an
  error — the bead IS closed and committed, and the repair is another push,
  never a second `bd update`.

## Opening a run — pre-flight, then `sprint-scope`

Three things happen before the first bead, in this order:

```bash
eval "$(scripts/pre-flight-check.sh)"              # 1. environment
#   2. invoke `sprint-scope` — it returns BB_EPIC
eval "$(scripts/pre-flight-check.sh "$BB_EPIC")"   # 3. + checks baseline
```

**1 and 3 are a script you run yourself**, not a skill: run it, eval the
stdout, read the stderr report once. **Act on the verdict, not on the
evidence** — it names each blocker AND its remedy, so there is nothing to
re-derive and nothing to weigh. `references/pre-flight.md` only when it blocks.

**A blocking condition means the run does not start.** Do not work around it.

**2 is the `sprint-scope` skill.** Invoke it by name. It turns whatever the
user gave into exactly one run epic, takes the run's single authorization gate
in the same interaction, and dispatches nothing. What you need from it:

- `BB_EPIC` — its id is the run id and its descendants are the queue
  (`bd ready --parent <epic> --exclude-type=epic`).
- The admission record — admitted, **excluded** and why, moved and from where.
  A run that narrowed its own scope says so in the final report.
- The authorization set as decided, including the rows that took their default.

**The readiness gate is not here and not in `sprint-scope`** — it belongs to
grooming, because readiness is a property of a groomed bead and asking before
grooming answers a question about beads that are about to change.


## Groom & plan — invoke `sprint-plan`

**One dispatch, two returns.** `sprint-plan` grooms every bead that is not
Ready (invoking `groom-beads`), bounces what grooming cannot clear, and — while
it still holds every bead in context — partitions what came out. It returns:

1. **The readiness verdict** per bead — ready, or NOT-READY with the gap. You
   dispatch only ready beads; bounced beads are raised into `bd human` by the
   skill and named in your report.
2. **The wave table**, plus one line per override of its own wave preview.

**Grooming and planning are one step because they are one read.** Partitioning
needs exactly the knowledge grooming just acquired, so a second agent would
re-read the whole queue to learn what the first already knew. **One subagent,
never two** — WHERE the queue needed no grooming at all, it plans alone.

**Do not store the wave table.** It is derived from the ready front, and a
stored copy only drifts out of agreement with the graph it describes.

Three things stay yours:

1. **The run manifest — `--metadata` JSON on the run epic.** The declared
   check ids, the checks baseline, the model lineup, the run's effort setting
   (`metadata.execution_reasoning_effort` per bead, on the convention's
   `low`/`medium`/`high`/`xhigh` scale), standing authorization, the
   environmental ledger, the base sha, any commit trailers, and the
   **admission record** — which beads were admitted, which were excluded and
   why, and which were moved and from where. Written **before** you dispatch
   grooming. WHERE the harness exposes no effort setting, write `unrecorded` —
   never omit the field and never guess a value; an omitted field is
   indistinguishable from a forgotten one, and a guessed value is worse.

   **Merge, never replace.** Pre-flight writes the baseline into the same
   object, and the epic already carries grooming's own keys. Read it, update
   your keys, write it back:
   `bd show <epic> --json | jq .metadata` → edit → `bd update <epic> --metadata @file.json`.

   Recovery reads the epic rather than re-deriving the run, and the epic
   outlives a reclaimed container where a `/tmp` path does not.

   **Register the run: `bd swarm create <epic>`.** Once, after the manifest is
   written. It makes the run discoverable by `bd swarm list` to a session that
   has nothing else, and `bd swarm status` becomes the live view.

   - It is **not** silently idempotent — a second call prints
     `Swarm already exists: <id>`. On a restart that is the recovery path:
     read the existing id from that message. **Never pass `--force`**, which
     mints a duplicate swarm for one run.
   - **Never point it at a bare bead.** Given a non-epic it auto-wraps by
     creating an epic with that issue as its only child — a reparent, and
     exactly the structural damage admission exists to avoid.
   - The swarm molecule is **not ephemeral**. Close it when the run ends, or
     swarms accumulate.
   - **`bd swarm list` reports closed molecules under "Active Swarms"**, so it
     answers "which runs exist", not "which are in flight". Read each
     molecule's own status before believing a run is live.

   The swarm id is **not** the run id. A swarm molecule is `issue_type:
   molecule` with no children — `bd ready --parent <swarm-id>` returns nothing.
   It links to the epic; the epic owns the queue, and `bd swarm status` accepts
   either id, so the epic loses nothing by being the identity.
2. **The bead list.** Pass `sprint-plan` explicit bead ids, never
   the filter that produced them, and the epic they belong to. The queue is
   `bd ready --parent <epic> --exclude-type=epic` — `--exclude-type=epic` is
   mandatory, because `--parent` is transitive and would otherwise hand back
   the run's own sub-epics as if they were work.
3. **Amendments — `amend:`-prefixed comments on the run epic.** WHEN the run
   diverges mid-flight — a wave resequenced, a batch reclassified, a bead
   admitted or bounced, a tier escalated — `bd comment <epic> "amend: <what
   and why>"`. Append-only and timestamped by construction, so there is no
   read-modify-write to lose and no ordering to maintain. Never re-invoke
   `sprint-plan` to re-partition a run in flight.

   **The wave table is not stored anywhere.** It is derivable from the graph —
   `bd ready --parent <epic> --exclude-type=epic` re-queried, plus
   `bd swarm validate` for a preview — so a stored copy can only drift out of
   agreement with the beads it describes.

   **`swarm validate` is a preview, never the dispatch order.** It scopes
   dependency analysis to the epic it is given: it reports an edge to a bead
   outside that epic as a *warning*, then computes waves and readiness as if
   the edge were absent — measured placing a blocked bead in Wave 1. It is also
   one level deep, so on a nested run epic it lists sub-epics as work items and
   never sees their leaves. **Suppress the preview** when the run epic is
   nested or when `swarm validate` warns about an outside dependency; the
   re-queried front is correct in both cases.

**The manifest is written on every run** — there is no queue-size or
batch-count exemption. Chat is the one store that does not survive.

Print the returned wave table to chat. Don't ask for approval — just print and
proceed. The user can interrupt if they disagree.

WHEN no `Agent` tool is available, nothing can be delegated, planning included
— that is a blocking pre-flight condition, not a mode. Stop and hand the queue to
the `run-bead` skill, one bead at a time.

### Concurrent dispatch

**Concurrent dispatch is ENABLED.** Several beads may be dispatched at once
within a wave, each in its own worktree, merged as each returns.

Note what is and is not concurrent. **A wave is the unit of parallelism**, and
waves remain sequential — a wave boundary is where dependency edges resolve and
where the base ref advances. Concurrency is *within* a wave, across beads that
the dispatch plan established are independent.

**A wave is the current ready front, packed.** Re-query
`bd ready --parent <epic> --exclude-type=epic` at every boundary; what comes
back is what dependencies now allow, and the plan says only how to pack it.
Never advance by walking a stored wave list — a bead that closed early, a bead
blocked by something discovered mid-run, and a bead admitted since the plan was
made all make a stored position wrong, and the graph is right about all three.

Entry conditions, both required and both about **safety**: **no shared
`c4_changes` component** between any two beads in the wave, and the base-ref
check re-verified for this wave. Both are yours to check at dispatch, whatever
the plan says.

**There is no minimum bead count, and do not reintroduce one** — a count in a
list of safety conditions reads as one, and nothing about two concurrent beads
is less safe than three. Whether a wave is *worth* dispatching concurrently is
ordinary judgment; whether it is *safe* is the entry conditions above.

### Tier selection per bead

Beads carry a **capability tier** from grooming — the readiness gate's
gate guarantees one (grooming assigns a default, never null), stored as
`metadata.execution_suggested_model`, falling back to `metadata.tier`. Both
are written by grooming; the first is `bd`'s own documented convention
(`docs/METADATA.md`), which is what makes a groomed bead routable by an
orchestrator that never read this skill, and the second is what 205 of this
tracker's existing beads carry. **Read the documented key first and fall back**
— that is the whole migration, and no bulk rewrite is required. Despite its
name, `execution_suggested_model` holds a TIER: the convention defines it as
"a capability-tier suggestion", to be substituted for a same-tier model on a
different provider. The tiers are ordered, least to most capable:

```
light  →  standard  →  advanced  →  frontier
```

Beads name a **tier**, not a model, so a backlog groomed months ago still
routes correctly. Resolve tier → model from the project's `## model
lineup` block in `sf.toml`, echoed by pre-flight. Honor these:

- **Collect during grooming** — you're already reading each bead's
  readiness signals; record its tier then. For a bulk query, read the
  metadata and **pass `--status all`**: `bd list` defaults to open issues, so
  a filter written without it silently under-counts a groomed backlog.

  ```sh
  bd list --status all --json | jq '[.[] | select(.metadata.tier == "standard")] | length'
  ```
- **Dispatch at the resolved model** — pass `model: "<resolved>"` on the
  `Agent` tool call for that bead's subagent. Absent a lineup, omit the
  param (the subagent inherits the session model) and say so once.
- **A named `model` on a bead is an override** — grooming sets it only
  when a specific model is genuinely required. It wins over the tier.
- **Main-loop beads** — a bead at the orchestrator's own tier with a
  note saying "do not delegate" is implemented by the orchestrator
  directly, not dispatched. It still follows the full per-bead ceremony
  (claim, gates, retro, atomic close, single commit).
- **Batching across tiers** — prefer batching beads that share a tier.
  When a batch mixes tiers, run the whole batch subagent at the
  **highest** tier among its members — never run a bead below its tier.
  Split the batch WHEN more than one bead would be dragged above its
  tier, OR when any bead would be dragged more than one tier. Weigh that
  against what batching saves: one dispatch instead of N, one worktree,
  and — for a shared surface — one agent holding that surface's context
  across every bead touching it. It does **not** save check runs; every
  bead runs the project's checks before its own close either way.
- **Escalate, never downgrade** — if a bead fails gates/review twice or
  its subagent visibly flails, re-dispatch that bead one tier up. Never
  re-dispatch below the assigned tier, and never "save cost" by
  downgrading mid-run.
- **Record every escalation** — the bead, the assigned tier, the tier
  that completed it, and what failed. This is a required line in the
  Phase 4 report (see "Tier escalations"), and it is the run's only
  direct evidence about whether grooming's routing is calibrated. WHEN
  `CLAUDE_CODE_SUBAGENT_MODEL` is set, report it as Phase 4's "Tier
  escalations" section directs.
- **Record the routing** — the dispatch plan's wave table carries a
  `Tier` column for every bead.

### Push policy

- **`push-per-bead`** — subagent pushes after each commit. Best for sequential
  mode (recovery is easier on a per-bead failure).
- **`defer-push`** — subagent commits per bead, orchestrator pushes once at
  batch end. Default for fast-lane and shared-surface batch (one process;
  failures are observable).

For a single-bead "wave," `push-per-bead` and `defer-push` are equivalent —
don't bother labeling. Just push when the subagent returns.

## Review the run's own diff — invoke `sprint-review`

When the ready front is dry, **before the report**: invoke `sprint-review`. It
runs the reviewer the project declared in `sf.toml`'s `[review]`, scopes it to
`base_sha`..sprint head, files every finding as a bead, and classifies each as
in or out of the run's authorized scope.

**The loop is yours, and only the loop.** `sprint-review` owns finding things;
you own the counter, the budget and the decision to go round again:

```
review → findings admitted?  ──no──▶ done
              │yes
              ▼
    passes < budget?  ──no──▶ done, and say what is left
              │yes
              ▼
    re-enter at groom & plan  ──▶ execute the waves  ──▶ review
```

**Re-enter at groom & plan, never at dispatch.** Review *files*; it does not
groom. An admitted finding earns its readiness like any other bead, or bounces
like any other bead.

**The budget is 2 loops — three passes in total** unless the gate said
otherwise. **Findings from the human's pull-request review do not consume it**:
that counter bounds the machine reviewing itself, and a human asking for a
change is not a loop the run gets to decline.

**The counter is durable.** Keep it in the manifest as `review_passes`, not in
your head — a resumed run that restarted it would loop forever.

**Exhausting the budget is not a failure.** Remaining findings stay filed as
beads, are named in the report, and the run proceeds. A third pass that still
finds things is information for the human, not a reason to keep going.

**No reviewer declared** means `sprint-review` returns a skip, and the report
says so. Do not substitute your own review.


## Phase 4 — Final report

After the queue is drained (or the circuit breaker tripped), print:

0. **Incidents — first, before the table.** Every decision this run made that
   changed what "correct" means: an acceptance criterion amended, a
   standard/requirement/AC tension resolved, a contract or interface changed, a
   spec silent where a bead needed it to speak. Harvest them with

   ```bash
   bd export | jq -r '.comments[]?.text' | grep -E '^(retro: incident|amend: incident)'
   ```

   one row each — bead, what was decided, what was rejected. **State `none`
   explicitly** when there were none; the null is information, and a missing
   section reads like a clean run rather than an unreported one.

   This section is first because it is the one a human skims, and it is what
   earns an autonomous run its autonomy. **The count is a quality metric for
   the specs, not for the run**: each incident marks a place the corpus failed
   to answer a question, so each is the retro's material, not a complaint.
1. A markdown table: `| ID | Status | Batch | Commit | Notes |`
2. Totals: `N closed, N blocked, N skipped`.
3. **Implicit decisions made** — the rest of the beads' `retro:` comments:
   which beads had implicit decisions, what the subagent chose. Surface to the
   user so they can audit non-obvious choices. Incidents are already in 0;
   do not repeat them here.
3a. **Tier escalations** — every bead whose tier you had to raise to get it
    done, one row each:

   ```
   | Bead | Assigned | Completed at | What failed at the assigned tier |
   |------|----------|--------------|----------------------------------|
   | demo-9h0 | light | standard | fixture regen wrong twice; gates failed both cycles |
   ```

   State `none` explicitly when there were no escalations — the null is
   the result, not the absence of one.

   **`CLAUDE_CODE_SUBAGENT_MODEL` override and no routing are different
   results.** WHEN `CLAUDE_CODE_SUBAGENT_MODEL` env var is set, every
   subagent runs at that forced model regardless of tier, so no routing
   occurs and tier escalations mean nothing — omit the table and write:

   ```
   Tier escalations: N/A — CLAUDE_CODE_SUBAGENT_MODEL is set, so every
   bead ran at that model and no tier routing occurred.
   ```


   **Say this alongside the table**, because the reader will otherwise
   draw the wrong conclusion from a short one: escalations reveal
   *under*-routing only. A bead groomed `frontier` that a `light` model
   would have finished never appears here — it succeeds, and the
   overspend is silent. So zero escalations across a large queue is not
   evidence of good calibration; it is equally consistent with routing
   everything too high (see the over-routing table below). WHEN a run of
   ≥10 beads produces zero escalations, say so and suggest the next run
   try one tier lower on the beads whose `tier_reason` was thinnest.
3b. **Deliberate over-routes** — every bead you knowingly ran a tier above
    its assigned one (e.g. batched into a higher-tier batch per the
    tier-batching rule), one row each:

   ```
   | Bead | Assigned | Ran at | Why raised |
   |------|----------|--------|------------|
   | demo-9h0 | light | standard | batched with a standard bead on the same surface |
   ```

   State `none` explicitly when there were no over-routes — the null is
   the result, not the absence of one. This table and the escalation table
   above are together the run's complete routing evidence, and
   `groom-beads`' retro loop needs both to recalibrate the rubric.
4. **Follow-up beads** — beads created mid-run via `bd create`, split into those
   admitted into the run (in-scope prerequisites / AC-decompositions / coverage
   gaps, with their outcome) and those left for the user (out-of-scope finds,
   filed with a `discovered-from` link, awaiting triage).
5. **Stashes** — if any were created during recovery, list with `git stash
   list`.
5a. **Sprint retro** — synthesize the run's own evidence into how the next
    sprint runs better. Specified in "The sprint retro" below. It runs on
    every run that closed at least one bead, and it runs *here*: after the
    report, whose tables it reads.
6. **Pushes** — confirm all closed beads' commits are on the remote.
6a. **Cancel the durable out-of-session trigger** armed at first dispatch
    (`references/progress-streaming.md`, "Back the Monitor with a durable
    trigger") —
    a trigger left armed after the run drains wakes an agent into a finished
    sprint, which trains the reader to ignore it. **State its disposition in
    the report**: `Durable trigger: cancelled (<id>)` or `Durable trigger:
    none armed` — both branches are stateable, and an omitted line is
    indistinguishable from a forgotten cancellation.
7. The run epic id, so the user can
   review the dispatch plan.

Do NOT create a PR or merge to main — that is a separate, explicit user action.

### The sprint retro

A run is a sprint; this is its retrospective. It runs on **every run that
closed at least one bead**, not only on a run that struggled.

**Read the run's records; do not re-derive the run.** The inputs are the
harvested `retro:` comments, the tables you just printed (0 incidents, 3a
escalations, 3b over-routes, 4 follow-up and admitted/bounced beads), the run
epic's `amend:` comments, and `git log` on the sprint branch.

**Incidents are this retro's best material.** Each one marks a place the
corpus failed to answer a question a bead needed answered, so each is a
candidate finding against a *document* — an RFC for a spec, a gap in a
standard — rather than a complaint about the run. Route them accordingly:
they belong to `retro-spec`'s audience, not `retro-skill`'s, unless the thing
that failed to speak was a skill. Do NOT re-read diffs, re-run gates, or re-open the beads' work. A
retro that costs a second pass over the run is a retro that gets skipped.

**Harvest first, then assert.** Findings are `retro:` comments on the beads —
there is no per-bead retro file. Harvest with the `bd export | jq` pipeline in
`references/operating-principles.md` ("Findings harvest"), once filtered to
closed and once unfiltered. A zero count across a drained run is a broken
query, not a clean run: **stop, re-run it**, check the label spelling and
confirm the beads really are closed, before aggregating anything.

**Answer every question below. A question with no evidence gets `null result`
in writing** — omit it and an unasked question is indistinguishable from an
answered one.

1. **Grooming** — which beads were under-groomed, and how did it surface
   (bounced at the readiness gate, an AC unsatisfiable as written, a
   description that failed the fresh-agent test in practice, a wrong or
   missing dependency edge)? Which tier predictions were wrong, in which
   direction? *From:* the bounce list, 3a, 3b, `retro:` comments.
2. **Specs** — which spec documents did an implementer find wrong, thin or
   self-contradictory? *From:* `retro:` comments naming a document.
3. **Plan** — what did `## Amendments` record (resequencings,
   reclassifications, admissions, bounces), and what would have produced a
   better initial batching? *From:* the amendments log read against the wave
   table it amended.
4. **Skill** — where did `bead-bonanza` itself mislead, and what did a
   subagent improvise around? *From:* `retro:` comments citing this skill,
   its references or its scripts.

**Write one sprint report per run** at
`rfcs/sprints/<YYYY-MM-DD>-<epic-slug>.md`, dated and slugged rather than
named for the run id. It carries the answers and **every harvested finding
verbatim, including the tail that
warrants no RFC**, and it says so: `triage-rfc` defers on recurrence, so a
finding recorded once and filtered out the next time can never be observed
appearing twice. Filtering the report down to RFC-worthy findings breaks that
accumulation silently.

**Number findings per bead, not globally: `<bead-id>#<ordinal>`** (e.g.
`8ef.1#3`), the ordinal given by position within that finding's own
`### agent-plugins-<id>` heading. Cite findings by this form everywhere the
report's analysis references one. A per-bead id is stable the moment a
finding is harvested; a global sequence number does not exist until the
whole report is assembled, so analysis written before assembly — which is
how this report gets written — would be citing a number that isn't there
yet.

**The bar for anything past the report: it changes how the next run is
dispatched.** "Wave 2 needed three cycles" is a summary. "The `schema`-labelled
beads were batched shared-surface but each regenerated the same fixture, so
the batch serialized anyway — batch that label sequential" is a finding. Only
findings that clear that bar become RFCs.

**Route them, as offers.** Grooming and tier-calibration findings →
`groom-beads`; spec findings → `retro-spec`; skill findings → `retro-skill`.
Each route is one offer the user accepts or declines, naming the findings it
would carry. RFCs land in `rfcs/inbox/<skill>/` and **cite the sprint report
instead of re-inlining its evidence**, so report and RFC cannot disagree: the
report is the testimony, the RFC is the proposal built on it. Every routed RFC
SHALL carry the run's model lineup and effort setting from the run epic's
manifest into its Provenance — the manifest records them precisely so
this step can, and an RFC without them ships with an unfillable provenance
field.

The report itself is not an offer. Write it, append one row for the sprint to
`rfcs/retro-log.md` so one index covers all retro activity, and commit both
with any RFCs in a single commit.

## Operating principles

- **Groom before you dispatch.** bead-bonanza executes *groomed* beads; it never
  fixes an under-specified bead inline. The readiness gate is that gate, and it holds for
  beads pulled into scope mid-run too.
- **Bias toward batching, not sequential.** Find the largest safe batch —
  surface-overlap is the dominant signal, not bead complexity. Sequential is
  reserved for genuine surface divergence or single-bead queues, not "I'm being
  cautious."
- **The push is the gate.** Commits inside a branch are recoverable; only the
  push to origin is committed-to-the-world. That never licenses an unvalidated
  commit: every bead runs the project's declared checks before its own close,
  batched or not, so the last bead's passing run is what clears a batch.
- **Targeted tests per bead are non-negotiable**, on top of the declared checks
  rather than instead of it. Each bead runs the tests covering what it touched.
  If the subagent can't identify any, that's a signal the change is broader than
  expected — pause and reconsider scope, don't skip.
- **Red before green on defect beads.** A test written after its fix encodes the
  fix's own assumptions and can stay green while the bug is live. For a bead
  fixing a defect or regression, observe the reproducing test fail for the
  stated reason before fixing; the retro names that observed failure, not just
  "tests pass".
- **One commit per bead, always — no exceptions.** An interrupted bead's work
  stays uncommitted in its own worktree, so there is no `WIP` commit to make
  and nothing to squash. Full rule and resume path:
  `references/operating-principles.md`.
- **The dispatch plan has no lean path.** It is written every run, whatever
  its size.
- **Surface implicit decisions, don't bury them.** A bead with an open
  decision is bounced at the readiness gate. A choice a subagent still makes silently
  — one the bead's Design didn't pin — belongs in its `retro:` comment, not
  the diff.
- **Nothing escalates to a human mid-run.** Not an ambiguous acceptance
  criterion, not a contract change, not a silent spec. You decide, you record
  it as an **incident** — `retro: incident —` on the bead, `amend: incident on
  <id> —` on the run epic — and the report names it first. A recorded decision
  gets reviewed at the pull request and revised in a later run; that is cheaper
  than a night blocked on a question nobody is awake to answer. This binds you
  as much as your subagents: an incident you decide yourself is recorded the
  same two ways.
- **Visible progress beats opacity, and the fast path alone doesn't survive an
  interrupt.** Subagents write JSONL progress events; the orchestrator arms
  both wake mechanisms over them ("Recovery and progress streaming" above).
  Long opaque runs erode trust; a run whose only wake mechanism is in-session
  cannot survive an interrupt that lands on the orchestrator.
- **Subagents discard context on return.** This is why the skill prefers fresh
  subagents over `/compact`. Batch subagents (fast lane, shared-surface)
  compromise — they hold context for N beads but free it on return.
- **Don't read source files in the orchestrator.** Code reads happen inside
  subagents — grooming reads source to close decisions and wire
  dependencies before dispatch.
- **Capture decisions in the bead, not just in chat.** The bead description
  (selected options and alternatives, set by grooming) is the durable record;
  the dispatch plan is session-scratch.
- **Fail loudly, recover quietly.** If a subagent blocks, log it and move on;
  the circuit breaker above is what stops the run. If a subagent crashes
  mid-step, follow the recovery procedure (the worktree and the bead are the
  anchors).
- **Route by tier, and report every escalation.** A bead's tier —
  `metadata.execution_suggested_model`, else `metadata.tier` — is
  grooming's *prediction* of the capability the work needs; resolve it to a
  model via the project's `sf.toml` `[models]` lineup and dispatch there (`sprint-plan` → Tier
  selection). Assign nothing below the tier; escalate one tier on repeated
  failure; never downgrade to save cost mid-run. Every escalation is a required
  row in the Phase 4 report — it is the only routing evidence a run produces.
  Beads marked "do not delegate" run in the orchestrator's own loop.
- **Project rules win over skill defaults.** `sf.toml`'s `[[checks]]` is the
  source of truth for build/test/lint/format commands and anything else the
  project runs per bead. Progress visibility defaults to the run checklist (above); task
  tools are the exception for projects with no durable tracker of their own —
  read the project's `CLAUDE.md`/`AGENTS.md` once at session start. If the
  project has its own commit message conventions, follow them. The skill is
  opinionated but not authoritative.
- **Auto-chain waves under standing authorization.** When dependencies mean
  closing one wave unblocks the next, ask ONCE (at run start or first chain
  point) for standing approval: "re-resolve the queue after each push and
  dispatch the next wave until dry, implicit decisions defaulted, explicit
  `decision-required` beads still stop." With it granted, waves flow without
  per-wave prompts: emit the wave's report and **dispatch the next wave in
  the same turn**. The report is a progress notice, not a checkpoint — under
  standing authorization the turn ends only when the queue is dry, a
  `decision-required` bead is reached, or the circuit breaker trips (Phase 3:
  3 consecutive beads `blocked` or failed). Ending a turn with a non-empty
  queue and none of those causes is a defect, not a pause: nobody is waiting
  to restart it. Newly-filed beads that enter the queue mid-run (subagents
  filing adjacent issues) are admitted only by the mid-run scope test
  (`sprint-plan`'s emergent-work test), never merely because they match the filter.
- **Keep an environmental-failure ledger; never re-diagnose per wave.** Suites
  that fail — or silently PASS for environmental reasons — in the local sandbox
  get diagnosed ONCE, recorded in the checks baseline pre-flight captured,
  then the verdict is passed into every subsequent subagent prompt as fact. The
  silent-pass case is the dangerous one: a gate can go green without exercising
  your change (a harness running a stale prebuilt binary a plain test command
  didn't rebuild, a suite skipped by default, a cached result). When a gate
  *can* pass without touching your change, force a clean rebuild and confirm it
  actually ran before trusting it. Heavy suites (integration, codegen, anything
  needing sockets/toolchains the sandbox lacks) run locally only when a bead
  touches the surface they cover; when a PR is open, CI is the backstop.
- **Guard against foreign dirt; tracker writes need no deferring.** Multiple
  agents can share the tree, so name known foreign files as untouchable in
  batch prompts and commit your own work by explicit pathspec
  (`git commit -- <files>`). Tracker writes are exempt: groom a discovered
  bead, bake a decision, or wire a dependency edge at any time, including
  mid-wave. Why, and the claim-reset rule:
  `references/operating-principles.md`.
- **Wire findings capture into every dispatch prompt unconditionally — no
  project file gates it.** Full wiring and the harvest pipeline:
  `references/operating-principles.md`.

## Side effects — what each script leaves behind

These are the artifacts each script creates, writes, or repoints.

- **`scripts/pre-flight-check.sh`** — **writes the checks baseline into the
  run epic's manifest**, merging rather than replacing so grooming's and
  admission's keys survive; emits `BB_*` facts (`BB_CHECKS_IDS`,
  `BB_MODEL_*`, etc.) on stdout. It creates no run directory. It also: runs
  `git worktree prune` at startup; **pushes to the remote** — probes tracker
  write access with `git push <remote> HEAD:refs/dolt/preflight-probe-$$`,
  then deletes that ref, both against the same remote the bead's own work
  will later push to; and **executes the project's declared checks**, which
  run whatever commands that project declared. It does not write user-global state — where
  Dolt telemetry suppression or the `eventsapi.dolthub.com` deny entry is
  absent, it reports that and prints the command to run yourself.
- **`scripts/bb-verify-skip.sh`** — no side effects; exits 0 (skip
  verified), 1 (not skipped), or 2 (invocation error).
- **`scripts/bb-close-bead.sh`** — runs `bd update --status closed`, stages
  the specified paths (or, with none given, everything modified **except
  `.beads/*.jsonl`**), makes one commit, then runs `bd dolt push` when a Dolt remote
  is configured. It does **not** write or stage `.beads/issues.jsonl`: the
  tracker's transport is `origin refs/dolt/data`, the JSONL exports are
  untracked, and a close commit carries code only.

Each script's own header carries the same usage line. WHEN a usage line
here disagrees with the script, the script is authoritative — and that
disagreement is a finding worth a retro.

Files under `references/shared/` are **generated copies** of reference
documents owned by a sibling skill — vendored in by
the repo-level `scripts/sync-shared-references.sh` so this skill is
self-contained. Read them normally; never edit them — fix the source named
in each file's header and re-run that script.

## Retro this skill

Question 4 of Phase 4's sprint retro ("The sprint retro") owns this. It asks
where `bead-bonanza` misled and what a subagent improvised around, answers it
from the run's harvested `retro:` comments, and offers `retro-skill` when
those comments carry a finding. Do not gate a retro of this skill on whether
the run felt like it struggled — the comments are the evidence, and a run that
felt clean can still have produced several.

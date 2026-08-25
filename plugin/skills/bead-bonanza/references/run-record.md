# The run record — manifest, swarm, amendments

## Contents

- The run's epic shape — ceremony epics chained under the run epic, one flat
  swarmed target, and the gate as a task
- The three human channels — authorization, incident, deferred review — and
  the four measured `bd` facts the deferred one depends on
- The three things the orchestrator owns: the manifest, the swarm
  registration, the run's amendments
- `graph_plan` — validating and loading the `bd create --graph` document the
  plan returns, and why both of its failure modes are silent
- Run health — the asserted dimension and its five values
- `base_sha` — why it is recorded at creation and never derived

## The run's epic shape

`sprint-scope` builds it; everything below assumes it. The run epic holds one
epic per phase — pre-flight, scope, plan, TARGET, review, retro, report,
post-flight — chained `epic→epic` in that order, with the work beads as flat
task children of the TARGET.

**Ceremony phases are epics because nothing else can block the target.**
Measured on bd 1.2.2: `bd dep add <task> <epic>` fails with `tasks can only
block other tasks, not epics`, and `bd dep add <epic> <task>` with `epics can
only block other epics, not tasks`. So a task cannot gate a phase. One
`epic→epic` edge does the whole job: measured on a run epic shaped this way,
`bd ready --parent <target> --exclude-type=epic` returned nothing while the
target waited on the plan epic — **a blocked parent blocks its children** — and
a bead created under the target mid-run came back from `bd blocked` as
`Blocked by 1 open dependencies: [<target>]` with no edge wired for it.

**The TARGET is FLAT — tasks only — and flatness is an ADMISSION rule.** It
binds every bead added mid-run, not only the ones the plan created, because
`bd swarm validate` is one level deep and a sub-epic added later breaks the
count exactly as one planned would. Measured on the same database: `bd swarm
validate <target>` over two task children reported `Total issues: 2` and named
both beads; `bd swarm validate <run-epic>`, whose five children are all epics,
reported `Total issues: 5` and named the ceremony epics as the work — never
seeing the beads at all. Both printed `Swarmable: YES`. **Work that genuinely
decomposes into sub-epics becomes a SECOND target under the run epic**, swarmed
separately, rather than a sub-epic under the first.

**The queue expression is unchanged by the shape, and this is measured rather
than assumed.** `bd ready --parent <run-epic> --exclude-type=epic` still returns
exactly the target's open leaves: by the time the target opens, the phases
before it are CLOSED (so their task children are too — an epic is only
close-eligible once every child is closed), and the phases after it are still
blocked behind it. Observed on the demonstration run at that moment: three
rows, all three the target's, the mid-run addition among them. So the run epic
stays the one id a recovering session needs, and only `bd swarm` takes the
target.

**The authorization gate is a TASK child of the scope epic.** A gate cannot
attach to a ceremony epic — gating one would need the refused task→epic edge —
and it does not need to: the scope epic already blocks the plan epic, so
holding its task child holds the whole chain. Verified end to end: with
`bd gate create -t human --blocks <gate-task>` open, `bd ready --parent <scope>`
returned nothing; the gate task had entered the ready front the moment
pre-flight's own epic closed.

## The three human channels

**This table is their only home.** Every skill that opens one of these channels
states its own row inline and names this document by role rather than copying
the table; a second copy is what drifts.

Three ways a run reaches a human, separated by **when the run consumes the
answer** — not by urgency, and not by who wrote them:

| channel | filed as | blocks | resolved by |
|---|---|---|---|
| **authorization** | `bd gate create -t human --blocks <gate-task>` on the scope epic's task child | at the START — the gate holds the whole phase chain | `bd gate resolve <gate-id>` |
| **incident** | a `retro:<kind>:<target> incident — …` comment on the bead, and an `amend:` on the run epic | NEVER — recorded, and read at the retro and the report | nothing; it is a record, not a question |
| **deferred review** | a task labelled `human`, parented to a LATE ceremony epic (review or report) | only at the END — that epic cannot close while it is open | `bd update <id> --status closed` |

**A run is autonomous because of the middle row.** The first row is taken once,
before any work; the third does not interrupt anything; everything else in
between is decided by the agent and recorded. A fourth channel — stopping to
ask mid-run — is the thing this design exists to make unnecessary.

### Four measured facts the deferred channel depends on

All four on bd 1.2.2, in a scratch database with the phase chain above.

- **The `human` label does NOT hold a bead out of `bd ready`.** Only the
  parent phase does. Measured both ways: while the report epic was blocked
  behind the target, `bd ready --parent <run> --exclude-type=epic --json`
  returned `[]`; with the target closed, the same query returned the
  `human`-labelled bead as ready work. A control bead labelled `human` under
  the already-closed TARGET was returned immediately.

  **So the dispatch front must exclude it** — `bd ready --parent <epic>
  --exclude-type=epic --exclude-label human`. Without that, the run hands a
  question meant for a human to an agent. The exclusion belongs on the
  DISPATCH query only: a completion or queue count that hides `human` beads
  reports a run as drained while a human still owes it an answer.

- **`bd human list` shows it from the moment it is filed**, while it is still
  invisible to `bd ready` — which is the whole asymmetry the channel is built
  on. A human can answer whenever; the run never waits.

- **`bd human respond` and `bd human dismiss` BOTH fail.** Measured:
  `bd human respond <id> -r "…"` → exit 1,
  `Error: resolving issue ID <id>: cannot resolve issue ID "<id>": storage is
  nil`; `bd human dismiss <id>` → exit 1, identical error. Close the bead
  instead: `bd update <id> --status closed` exits 0, and `bd epic
  close-eligible` then closes the phase that was waiting on it (measured — the
  report epic went `open` → `closed` on the next call, and the run epic after
  it).

- **`bd human list` keeps listing a CLOSED bead.** Measured: after
  `bd update <id> --status closed` succeeded and `bd show` reported
  `status=closed`, `bd human list` still printed `Human-needed beads (1
  found)`. **Never read that command as a queue depth.** Anything asking "does
  a human still owe this run an answer" must filter by status —
  `bd list --label human --status open` — or it stays red forever after the
  first question is answered.

**Closing a ceremony epic is `bd epic close-eligible`, scoped BY HAND.** It
takes no positional epic. Measured: with one run's pre-flight child closed, the
bare command printed `✓ Closed 2 epic(s)` — pre-flight **and an unrelated epic
from another run in the same database**. Read what it names and confirm every
line belongs to this run.

## The three things you own

**Three things the orchestrator owns and never delegates.** Read this once,
after `sprint-scope` returns the epic and before you dispatch grooming.

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

   **The key names are the contract**, because dispatch reads them back rather
   than being handed them — `branch`, `checks_ids` and `checks_baseline` come
   off the manifest to build each batch prompt. A renamed key is a silent
   break: the reader gets `null` and dispatches without it. The
   full key set is the JSON block the **`sprint-plan`** skill writes, in its own
   run-record reference.

   **`graph_plan` is yours to write and the planner's to produce.** The plan
   returns the path of the `bd create --graph` JSON document that builds the
   run; you load it and record the path here, so a recovering session can diff
   what the run was supposed to be against what the graph now holds. Two rules
   the loader cannot skip, because both failures are silent:

   - **Re-run the plan's validators before loading.** `bd create --graph`
     reports an unknown node or edge field as a *warning*, drops it, and still
     exits 0 — a mistyped `assignee` is then a bead nobody owns and a mistyped
     edge key a dependency that was never wired. A malformed *value* is the
     opposite: it aborts before creating anything.
   - **Diff the plan before loading it.** A load that fails partway leaves the
     graph half-built, and being reviewable pre-load is most of what the
     document is for.

   **Merge, never replace.** Pre-flight writes the baseline into the same
   object, and the epic already carries grooming's own keys. Read it, update
   your keys, write it back:
   `bd show <epic> --json | jq '.[0].metadata'` → edit → `bd update <epic> --metadata @file.json`.

   Recovery reads the epic rather than re-deriving the run, and the epic
   outlives a reclaimed container where a `/tmp` path does not.

   **Register the run: `bd swarm create <target>` — the TARGET epic, never the
   run epic.** Once, after the manifest is written. It makes the run
   discoverable by `bd swarm list` to a session that has nothing else, and
   `bd swarm status` becomes the live view. Pointed at the run epic it counts
   the ceremony epics as the work and never sees a bead (measured, above).

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
   diverges mid-flight — an edge added, a batch reclassified, a bead
   admitted or bounced, a tier escalated — `bd comment <epic> "amend: <what
   and why>"`. Append-only and timestamped by construction, so there is no
   read-modify-write to lose and no ordering to maintain. Never re-invoke
   `sprint-plan` to re-partition a run in flight.

   **The dispatch order is not stored anywhere.** It is
   `bd ready --parent <epic> --exclude-type=epic` re-queried, so a stored copy
   can only drift out of agreement with the beads it describes.

   **`bd swarm validate` is a preview, never the dispatch order.** It scopes
   dependency analysis to the epic it is given: it reports an edge to a bead
   outside that epic as a *warning*, then computes readiness as if the edge
   were absent — measured, it grouped a blocked bead with the ready ones. It is
   also one level deep, so on a nested run epic it lists sub-epics as work
   items and never sees their leaves. **Suppress the preview** when the run
   epic is nested or when it warns about an outside dependency; the re-queried
   front is correct in both cases.

**The manifest is written on every run** — there is no queue-size or
batch-count exemption. Chat is the one store that does not survive.

Print the returned plan to chat. Don't ask for approval — just print and
proceed. The user can interrupt if they disagree.

WHEN no `Agent` tool is available, nothing can be delegated, planning included
— that is a blocking pre-flight condition, not a mode. Stop and hand the queue to
the `run-bead` skill, one bead at a time.

## Run health

Asserted by the orchestrator, never derived — a stuck agent and a slow one look
identical in git and in bead status. Set it at run start and at every re-query
of the front; `bb-status.sh run` prints the evidence for and against beside it.

```sh
bd set-state <epic> health=<value> --reason "<why>"
```

| value | meaning |
|---|---|
| `healthy` | beads closing, nothing failing repeatedly |
| `slow` | nothing closed recently, but agents are alive |
| `stalled` | nothing closed recently **and** no live agent |
| `degraded` | progressing, but with repeated failures |
| `blocked` | cannot proceed without a human; pair with `bd human` |

`slow` vs `stalled` is the distinction the dimension exists for. The label
**replaces** rather than accumulates.

## `base_sha` is recorded at creation, never derived

Because that is the only moment it is unambiguous. Deriving it later from a
merge-base assumes the base branch has not moved, and on a long run it will
have — leaving the review scoped to a diff that silently excludes whatever
landed on the base meanwhile.

Three things consume it and none may create it:

- the review scopes its diff `base_sha`..sprint head
- every worktree branches from sprint HEAD
- the pull request opens against it

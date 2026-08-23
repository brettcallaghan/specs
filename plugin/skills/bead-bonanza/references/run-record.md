# The run record — manifest, swarm, amendments

## Contents

- The three things the orchestrator owns: the manifest, the swarm
  registration, the run's amendments
- Run health — the asserted dimension and its five values

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

   **The key names are the contract**, because the wave tier reads them back
   rather than being handed them — `sprint-wave` takes `branch`, `checks_ids`
   and `checks_baseline` off the manifest to build each batch prompt. A renamed
   key is a silent break: the reader gets `null` and dispatches without it. The
   full key set is the JSON block the **`sprint-plan`** skill writes, in its own
   run-record reference.

   **Merge, never replace.** Pre-flight writes the baseline into the same
   object, and the epic already carries grooming's own keys. Read it, update
   your keys, write it back:
   `bd show <epic> --json | jq '.[0].metadata'` → edit → `bd update <epic> --metadata @file.json`.

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

## Run health

Asserted by the orchestrator, never derived — a stuck agent and a slow one look
identical in git and in bead status. Set it at run start and at every wave
boundary; `bb-status.sh run` prints the evidence for and against beside it.

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

---
title: "`bd swarm list` reports closed swarms under \"Active Swarms\""
description: A swarm molecule whose own status is `closed` is still listed,
  and counted, under the "Active Swarms" heading
target: beads (bd CLI)
version_observed: "1.2.2 (dev)"
status: ready-to-post
updated_at: 2026-08-22T00:00:00Z
---

## What is wrong, and what it costs

`bd swarm list` prints a heading of `🐝 Active Swarms (N)` and includes swarm
molecules whose own status is `closed`. The count in the heading includes them
too.

The cost is that **`swarm list` cannot answer "which runs are in flight"**,
which is the question its own heading claims to answer. Any consumer has to
read every listed molecule's status separately with `bd show`, so the command
is a list of swarms that have *ever existed* rather than of active ones.

For an orchestrator recovering a run with no context, this is the difference
between one command and N+1 — and the failure is silent: a closed swarm looks
exactly like a live one, so a recovering session can attach to a finished run
and wait for progress that will never come.

## Reproduction

Run on `bd version 1.2.2 (dev)`, against a Dolt-backed workspace, 2026-08-22:

```
$ bd swarm create <epic>            # note the molecule id
$ bd close <molecule-id>
$ bd show <molecule-id> --json | jq -r '.[0].status'
closed

$ bd swarm list

🐝 Active Swarms (2)

agent-plugins-tvcj Swarm: probe: swarm-list closed-molecule check
   Epic: agent-plugins-86gy (probe: swarm-list closed-molecule check)
   Progress: 0/1 (0%)
...
```

The molecule reports `closed` and is listed under "Active Swarms" in the same
breath.

## Proposed change

Either filter the listing to molecules whose status is not `closed`, or rename
the heading to say what it lists. **Filtering seems right** — the heading is
the more useful behaviour, and a `--all` flag could recover the current output
for anyone who wants history.

If neither, documenting that the heading means "all swarms" would at least let
consumers stop treating it as a liveness signal.

## How we work around it

Every consumer reads each listed molecule's own status before treating a run
as live, and our orchestrator's recovery path says so explicitly rather than
trusting the listing.

# Beads archive

`agent-plugins-beads-2026-10-06.jsonl` is the whole `bd` tracker as it stood
on 2026-10-06, when it moved off `refs/dolt/data` on this repository's git
remote into a new database, now on the `beads/data` branch. It is a one-time
`bd export`: one JSON object per line, 690 beads (561 closed, 107 open,
21 deferred, 1 in progress) with their comments and dependencies. It is never
updated.

Ids of the form `agent-plugins-…` anywhere in this repository resolve here,
not in the live tracker.

Only open, in-progress and deferred work moved to the new database, under the same
suffix with the `specs-` prefix: `agent-plugins-hx1x` became `specs-hx1x`.
Twelve non-closed beads stayed behind (six duplicates merged into or replaced
by another bead, and six that were not work for the new tracker), so an
`agent-plugins-` id with no `specs-` twin is either closed or one of those.
Comments and history on closed beads exist only here.

Read it with `jq`, no `bd` needed:

```bash
jq -r 'select(.id=="agent-plugins-hx1x") | .title, .status' docs/archive/agent-plugins-beads-2026-10-06.jsonl
jq -r 'select(.status!="closed") | [.id, .status, .title] | @tsv' docs/archive/agent-plugins-beads-2026-10-06.jsonl
```

`bd import` would load it into a scratch database. Never import it into the
live tracker: the ids and prefix differ, and it would resurrect closed work.

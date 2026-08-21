# bead-bonanza Progress Streaming and Mid-Stream Visibility

How the orchestrator sees a run while it is still in flight: the tracker-fed
event stream, the in-session `Monitor` armed over it, and the durable
out-of-session trigger that backs the Monitor up. Read this WHEN a subagent is
expected to run past 5 minutes — any non-trivial bead, given full test runs —
and before the first dispatch of a run, which is when the stream is armed and
the durable trigger set.

Two neighbouring situations are elsewhere, so this file can be read on its own:
WHEN a subagent dies mid-ceremony, read `recovery.md`; WHEN you are resuming a
run whose state you do not hold, invoke the **`recover-sprint`** skill.

**Two views, and they are not interchangeable.** `bd swarm status <epic>` is
the durable one — computed from the beads themselves, correct after any crash,
and the right answer to "where is this run". The event log below is a live
*stream* for a `Monitor` to tail; it is ephemeral by design and its absence
costs mid-flight visibility, not the run's record.

## Contents

- The event source — one line per tracker status change, written by `bd`
  rather than by the agent whose progress it reports
- Monitor any subagent past 5 minutes — the tail pattern, and what it filters to
- Back the Monitor with a durable trigger — the out-of-session check-in that
  survives an interrupt killing the orchestrator, and when it is cancelled

---

## The event source

Progress comes from the tracker write itself. `bd` runs
`.beads/hooks/on_update` on every status change, so a bead moving to
`in_progress` or `closed` emits a line without the subagent writing one — the
signal is a consequence of the state change rather than a report about it, and
an agent that dies between doing the work and describing it cannot desynchronize
the two.

Install the hook once per run, before the first dispatch:

```bash
mkdir -p .beads/hooks
cat > .beads/hooks/on_update <<'HOOK'
#!/bin/sh
# bd invokes this as: on_update <issue-id> <event-type>, issue JSON on stdin.
printf '%s\t%s\t%s\n' "$(date -u +%FT%TZ)" "$1" "$2" >> "$BB_EVENT_LOG"
HOOK
chmod +x .beads/hooks/on_update
```

`BB_EVENT_LOG` is emitted by the pre-flight script (as `BB_EVENT_LOG=`, in the
block you eval) and points at a per-run file under `${TMPDIR:-/tmp}`. **Export
it before installing the hook** — the hook appends to it, and with the variable
unset every append fails. `bd` runs hooks fire-and-forget and ignores their
exit status, so that failure is silent: the run reports nothing and nothing
says why.

The hook must also be **executable** — `bd` skips a non-executable hook
silently too, with the same presentation.

WHEN pre-flight reports the event log as unwritable it emits `BB_EVENT_LOG=''`;
skip the hook entirely and use `bd swarm status <epic>` for visibility rather
than installing a hook that cannot write.

Three properties of the mechanism that decide how it is used:

- **`on_close` does not fire on this project's close path.** `bd update <id>
  --status closed` fires `on_update`; only `bd close` fires `on_close`, and the
  close ceremony deliberately uses the former to avoid a second commit. Hang
  nothing on `on_close`.
- **Hooks are fire-and-forget.** `bd` runs them asynchronously, bounded by a
  timeout, and a hook that fails never blocks the write that triggered it. A
  missing event means a missing line, never a lost close.
- **`.beads/hooks/` is also `bd`'s git-hook directory**, and `bd init` rewrites
  its contents. Treat the file as something `bd` may overwrite: install it at
  run start rather than assuming it survived from last time.

## Monitor any subagent past 5 minutes

**Monitor any subagent past 5 minutes.** If a subagent is expected to (or
actually does) run >5 min — true for any non-trivial bead given full test runs
— the orchestrator must arm a `Monitor` on the event log so the user sees
events live. Don't leave the user staring at silence for 30+ minutes. Pattern:

```
tail -n +1 -F <run-dir>/events.tsv | grep -E --line-buffered 'close|update'
```

Arm it with `persistent: true` and stop it with `TaskStop` when the wave
drains. The default `timeout_ms` is five minutes — the exact threshold at
which this rule requires a Monitor — so a Monitor armed without it expires on
precisely the runs that need it. The `tail -F` pattern above never exits on
its own, so under a timeout it stays armed until it is killed rather than
ending when the work does. WHEN a Monitor does expire while its subagent is
still live, re-arm it; expiry is not information about the bead.

One log carries the whole run, including every bead a batch subagent processes,
so one Monitor covers a wave. The orchestrator acts on a bead reaching
`closed`; intermediate transitions are there for reading a stalled run, not for
waking on. Relaying every transition wakes the orchestrator per event and adds
a conversation turn each time for narration nobody needs.

## Back the Monitor with a durable trigger

A Monitor is in-session: it dies with the turn that armed it, so it cannot
report its own death. A user interrupt that kills the orchestrator kills the
Monitor at the same instant — and it kills the mechanism that would have
announced its children's deaths, so **nothing reports this to you**. A durable
run needs a second mechanism that survives that.

Before the first dispatch of a run, arm an out-of-session check-in **in
addition to** the Monitor:

- **Prefer a recurring trigger over a one-shot** — a cron/recurring trigger
  where the harness offers one, otherwise a re-armed `send_later`, at
  15–30 minute intervals. A recurring trigger re-arms itself instead of
  depending on the agent to remember, and the agent forgetting is the
  failure mode being defended against. State once at run start which kind
  the harness offers; WHERE it offers neither, say so once and proceed —
  the only wake path for this run is in-session.
- **The trigger's message must be self-contained, and it must name a skill
  rather than a file path.** It wakes a turn that may hold no run context and
  no bundle: carry the run epic id — the manifest is on it — and an instruction to
  **invoke the `recover-sprint` skill**. A fresh session cannot resolve a
  relative path into a bundle it never loaded, but it can resolve a skill
  name — and "check on the run" is not an instruction at all.
- **Do not key its liveness check on `bd list --status in_progress`.** A
  bead can be worked without ever being claimed, so that query reads
  "nothing in flight" while an agent is actively editing. Use
  `git worktree prune && git worktree list` instead — a worktree that still
  exists is a bead that is not finished.
- **Cancel it in Phase 4**, when the run drains. A trigger left armed after
  the run finishes wakes an agent into a finished sprint, which trains the
  reader to ignore it.

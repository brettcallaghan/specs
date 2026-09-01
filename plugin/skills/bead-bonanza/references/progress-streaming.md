# bead-bonanza Progress Streaming and Mid-Stream Visibility

How the orchestrator sees a run while it is still in flight: the tracker-fed
event log, the plugin's background monitor that reads it without being armed,
and the durable out-of-session trigger that survives the session dying. Read
this before the first dispatch of a run, which is when the log's writer is
installed and the durable trigger set.

Two neighbouring situations are elsewhere, so this file can be read on its own:
WHEN a subagent dies mid-ceremony, read `recovery.md`; WHEN you are resuming a
run whose state you do not hold, re-invoke **`bead-bonanza`** against the run
epic — it resumes rather than restarts.

**Two views, and they are not interchangeable.** `swarm status --run <epic>` is
the durable one — measured from the beads and from git on every call, correct
after any crash, and the right answer to "where is this run". The event log below is a
live *stream*; it is ephemeral by design and its absence costs mid-flight
visibility, not the run's record.

## Contents

- The event source — one line per tracker write, written by `bd` rather than
  by the agent whose progress it reports, and exactly which writes emit one
- The monitor reads it without being armed — what the plugin declares, and
  what a session sees
- WHEN there is no monitor — arming one by hand, and its expiry trap
- Back it with a durable trigger — the out-of-session check-in that survives
  an interrupt killing the session, and when it is cancelled

---

## The event source

Progress comes from the tracker write itself. `bd` runs a hook in
`.beads/hooks/` on every write, so a bead changing state emits a line without
the subagent writing one — the signal is a consequence of the state change
rather than a report about it, and an agent that dies between doing the work
and describing it cannot desynchronize the two.

Install the hooks once per run, before the first dispatch:

```bash
mkdir -p .beads/hooks
for h in on_create on_update on_close; do
cat > ".beads/hooks/$h" <<'HOOK'
#!/bin/sh
# bd invokes this as: <hook> <issue-id> <event-type>, issue JSON on stdin.
s=$(sed -n 's/.*"status":"\([^"]*\)".*/\1/p')
printf '%s\t%s\t%s\t%s\n' "$(date -u +%FT%TZ)" "$1" "$2" "${s:-?}" >> "$BB_EVENT_LOG"
HOOK
chmod +x ".beads/hooks/$h"
done
```

**Read the status off stdin, as above, rather than off `$2`.** The event-type
argument is the hook's own name (`create`, `update`, `close`) and never says
which status a bead reached, so a stream built from `$2` cannot tell a claim
from a close — the one distinction the orchestrator acts on. The issue JSON on
stdin carries `status`, which is why the line has four fields.

**Which writes emit a line, on bd 1.2.2:**

| write | hook | `$2` |
|---|---|---|
| `bd create` | `on_create` | `create` |
| `bd update <id> --status <s>` | `on_update` | `update` |
| `bd close <id>` | `on_close` | `close` |
| `bd update <id> --claim` | **none** | — |

**`--claim` emits nothing**, and it is the claiming primitive every worker
uses. A bead being picked up is therefore invisible in this stream; the first
line you see for a bead is usually its close. Do not read silence as a stalled
worker — for that, use the three signals in `supervision.md`.

`BB_EVENT_LOG` is emitted by the pre-flight script (as `BB_EVENT_LOG=`, in the
block you eval) and points at a per-run file under `${TMPDIR:-/tmp}`. **Export
it before installing the hooks** — they append to it, and with the variable
unset every append fails. `bd` runs hooks fire-and-forget and ignores their
exit status, so that failure is silent: the run reports nothing and nothing
says why.

The hooks must also be **executable** — `bd` skips a non-executable hook
silently too, with the same presentation.

WHEN pre-flight reports the event log as unwritable it emits `BB_EVENT_LOG=''`;
skip the hooks entirely and use `swarm status --run <epic>` for visibility rather
than installing one that cannot write.

Three properties of the mechanism that decide how it is used:

- **`on_close` does not fire on this project's close path.** The close ceremony
  uses `bd update <id> --status closed` to avoid a second commit, and that
  fires `on_update`. Installing `on_close` costs nothing and covers a
  hand-rolled `bd close`; hang nothing on it.
- **Hooks are fire-and-forget.** `bd` runs them asynchronously, bounded by a
  timeout, and a hook that fails never blocks the write that triggered it. A
  missing event means a missing line, never a lost close.
- **`.beads/hooks/` is also `bd`'s git-hook directory**, and `bd init` rewrites
  its contents. Treat these as files `bd` may overwrite: install them at run
  start rather than assuming they survived from last time.

## The monitor reads it without being armed

The plugin declares a background monitor — `monitors/monitors.json` at its
root, running `monitors/bb-events-monitor.sh` — and Claude Code starts it when
the **`bead-bonanza`** skill is invoked. Each line it prints arrives as a
session notification. **There is nothing to arm, and nothing to re-arm.**

It watches the *directory* `${TMPDIR:-/tmp}` for `bb-events-*.log` rather than
being handed a path, because it cannot be handed one: a monitor is spawned by
the harness, not by your shell, so the `BB_EVENT_LOG` you `eval` never reaches
it, and the run it will report may not be scoped yet when it starts. A log
appearing an hour later is picked up with no restart.

**A log already on disk when it starts is a backlog, not news**, so the monitor
begins at the end of one and at byte one of any log that appears afterwards.
That is what makes it safe to restart mid-run: it will not replay the run so
far as fresh notifications.

**Every line is a notification, so act on the ones that matter.** The
orchestrator acts on a bead reaching `closed` — that is the fourth field.
Intermediate transitions are there for reading a stalled run, not for waking
on.

## WHEN there is no monitor

**The monitor is a convenience over a path that still works without it**, and
several ordinary situations have none: the plugin is not installed, or the
harness offers no Monitor tool at all — on those hosts a declared monitor is
skipped silently.

WHERE the harness has a `Monitor` tool, arm one by hand, once, before the first
dispatch. One event log carries the whole run, so the run needs ONE Monitor
however many subagents it dispatches:

```
tail -n +1 -F "$BB_EVENT_LOG"
```

Arm it with `persistent: true` and stop it with `TaskStop` when the ready set
dries. **Set `timeout_ms` explicitly** — the default is five minutes, shorter
than most runs, so a Monitor armed without it expires part-way through the work
it exists to report. The `tail -F` pattern never exits on its own, so it stays
armed until killed rather than ending when the work does. WHEN one does expire
while the run is still live, re-arm it — expiry is not information about any
bead.

WHERE the harness has neither, say so once and use `swarm status --run <epic>`
between dispatches. No step in the ceremony depends on a monitor having
reported anything.

## Back it with a durable trigger

**A monitor is in-session, exactly as an armed `Monitor` is.** It dies with the
session, so it cannot report its own death — and a user interrupt that kills
the orchestrator kills the thing that would have announced its children's
deaths at the same instant, so **nothing reports this to you**. A durable run
needs a second mechanism that survives that.

Before the first dispatch of a run, arm an out-of-session check-in **in
addition to** the monitor:

- **Prefer a recurring trigger over a one-shot** — a cron/recurring trigger
  where the harness offers one, otherwise a re-armed `send_later`, at
  15–30 minute intervals. A recurring trigger re-arms itself instead of
  depending on the agent to remember, and the agent forgetting is the
  failure mode being defended against. State once at run start which kind
  the harness offers; WHERE it offers neither, say so once and proceed —
  the only wake path for this run is in-session.
- **The trigger's message must be self-contained, and it must name a skill
  rather than a file path.** It wakes a turn that may hold no run context and
  no bundle: carry the run epic id — the manifest is on it — and an instruction
  to **invoke `bead-bonanza` against that epic**, which recognises an existing
  run epic and resumes it. Name the user-facing skill, never an internal one: a
  fresh session cannot resolve a relative path into a bundle it never loaded,
  and "check on the run" is not an instruction at all.
- **Do not key its liveness check on `bd list --status in_progress`.** A
  bead can be worked without ever being claimed, so that query reads
  "nothing in flight" while an agent is actively editing. Use
  `git worktree prune && git worktree list` instead — a worktree that still
  exists is a bead that is not finished.
- **Cancel it in Phase 4**, when the run drains. A trigger left armed after
  the run finishes wakes an agent into a finished sprint, which trains the
  reader to ignore it.

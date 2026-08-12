# bead-bonanza Progress Streaming and Mid-Stream Visibility

How the orchestrator sees a run while it is still in flight: the progress
events subagents append, the rule bounding what those events prove, the
in-session `Monitor` armed over them, and the durable out-of-session trigger
that backs the Monitor up. Read this WHEN a subagent is expected to run past 5
minutes — any non-trivial bead, given full test runs — and before the first
dispatch of a run, which is when the durable trigger is armed.

Two neighbouring situations are elsewhere, so this file can be read on its own:
WHEN a subagent dies mid-ceremony, read `recovery.md`; WHEN you are resuming a
run whose state you do not hold, invoke the **`recover-sprint`** skill.

`${TMPDIR:-/tmp}` below is the run root Phase 0's pre-flight script resolved.

## Contents

- The progress-event schema — the closed phase vocabulary, one JSONL line per
  event, one file per bead
- The ledger is a hint, never the authority — what a trail does and does not
  prove, and the order a status question is resolved in
- Writing events — the one script that appends them, and why never by hand
- Monitor any subagent past 5 minutes — the tail pattern, and why it filters
  to two phases
- Back the Monitor with a durable trigger — the out-of-session check-in that
  survives an interrupt killing the orchestrator, and when it is cancelled

---

## The progress-event schema

Each subagent appends progress events to `${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>/<bead-id>.progress.jsonl`. Required event phases (one line each):

```jsonl
{"bead":"<id>","phase":"claimed"}
{"bead":"<id>","phase":"implementing"}
{"bead":"<id>","phase":"gates-running"}
{"bead":"<id>","phase":"gates-passed"}
{"bead":"<id>","phase":"retro-written"}
{"bead":"<id>","phase":"committed","sha":"<short>"}
{"bead":"<id>","phase":"pushed"}             # omitted in defer-push mode
{"bead":"<id>","phase":"blocked","reason":"..."}  # only on failure
```

**The ledger is a hint, never the authority.** It is written by the agent
doing the work, so it stops the instant that agent stops. Therefore:

- A trail ending before `committed` does **not** mean the bead is unfinished —
  only that nothing was appended after that point. A trail stopping at
  `implementing` routinely belongs to a bead that commits hours later, several
  beads on; read as a heartbeat it proves nothing but its own last write.
- An absent progress file does **not** mean no agent ran.
- A file's mtime bounds the last *write*, not the agent's liveness.

Resolve any status question in this order, stopping at the first answer:
`git worktree list` (present means unfinished — the one signal that is state
rather than testimony), then `bd show <id>`, then
`git log --oneline --grep=<id>`, then the ledger. **Name which one you used.** Never report a ledger fact you have not just read from
disk, and never from a backup copy while the live directory exists.

## Writing events

Events are written by `../scripts/bb-progress.sh`, never by hand:

```bash
../scripts/bb-progress.sh "$RUN_DIR" <bead-id> <phase> [key=value ...]
```

It validates the phase against the closed vocabulary above, escapes values, and appends one well-formed line. **Never build the event JSON in the caller** — a runtime-substituted `printf` operand trips the permission system's injection heuristic and prompts the user on *every* event write.

Batch subagents (fast lane, shared-surface batch) write events for every bead they process to the same per-bead progress files (one file per bead-id, even from the same batch subagent).

## Monitor any subagent past 5 minutes

**Monitor any subagent past 5 minutes.** If a subagent is expected to (or actually does) run >5 min — true for any non-trivial bead given full test runs — the orchestrator must arm a `Monitor` on its progress JSONL so the user sees per-phase events live. Don't leave the user staring at silence for 30+ minutes. Pattern:

```
tail -n +1 -F <run-dir>/<bead-id>.progress.jsonl | grep -E --line-buffered '"phase":"(committed|blocked)"'
```

For batch subagents, tail multiple files in one Monitor. **Filter to `committed|blocked` only** — those are the events the orchestrator acts on. Relaying every phase (implementing, gates-running, …) wakes the orchestrator per event and adds a conversation turn each time for narration nobody needs; the per-bead progress files still hold the full phase history for crash recovery.

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
  no bundle: carry the `RUN_DIR` path, the epic id, and an instruction to
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

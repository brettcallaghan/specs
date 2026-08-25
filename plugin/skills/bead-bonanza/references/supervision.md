# Supervising a run, and deciding it is done

Two questions the orchestrator asks repeatedly and must not answer from
recollection: *where is this run right now*, and *is it finished*. Both are
computed from state.

## Contents

- Where is this run — the durable view, and what it cannot see
- When something dies — which of three situations you are in
- When the run is done — the completion rule, and why the obvious query is wrong
- The `Stop` hook — the loop enforced by the harness, its bound, and how to
  stop anyway
- The `SubagentStop` hook — a worker held to its own queue, and why a BLOCKED
  bead is accepted where an unfinished one is not
- Running without the hooks — the fallback every step is written against
- The review loop's budget — what the counter bounds, and what it does not

## Where is this run

**`bb-status.sh run <epic>`** is the durable answer, computed from the beads
and from git, so it survives any crash and any lost session. It walks the run's
descendants itself rather than asking `bd` for them, for the reason in "When
the run is done" below, and prints the command behind every line so any of them
can be re-derived.

**Re-querying the ready front is two steps, in order.** First, close eligible
epics — scoped by hand, never with the bare command: `bd epic close-eligible`
takes NO epic argument (bd 1.2.2, `--help`) and a passed id is silently
ignored, so `bd epic close-eligible <epic>` reads as scoped and is not. Run
`bd epic close-eligible --dry-run`, intersect what it names with the epics
under this run, and close only those explicitly — the bare command (no
`--dry-run`) closes every eligible epic in the database, including epics
belonging to other runs and other humans. Then `bd ready --parent <epic>
--exclude-type=epic`. An epic does not close on its own when its children do
(measured), so skipping the first step is a silent deadlock, not a slowdown —
a finished ceremony phase stays open, the beads it gates never reach the
front, and an empty front reads as the run being done when it has only
stalled.

Live progress — the event stream, the plugin's background monitor that reads
it without being armed, and the durable out-of-session trigger — is
`progress-streaming.md`.

**Health is read off `bb-status.sh run`, never written.** It is derived from
surviving worktrees, the age of the last commit and the depth of the ready
front. A label the coordinator sets instead is a signal whose only writer is
the thing it monitors, so it reports "no failure" and "no monitor"
identically.

## When something dies

Take the one you are in; they have different remedies and reading the wrong one
wastes a dispatch:

| what died | where to go |
|---|---|
| nothing — you just want live progress | `progress-streaming.md` |
| a subagent, mid-ceremony | `recovery.md` |
| the orchestrator itself | invoke the **`sprint-recovery`** skill |

The last is a skill rather than a file because the session that needs it may be
fresh, with nothing loaded and no path it can resolve.

## When the run is done

**`bb-status.sh run <epic>`** decides — computed from state, not from your
recollection of what you dispatched. Exit 0 clear, 1 repairable (no admitted
set recorded), 2 stop.

A run is done when every bead the manifest's `admitted` list names is closed or
accounted for as blocked-with-a-reason.

**Beads excluded at admission are reported alongside.** A run that finished its
admitted set while three sat out is done over a *narrowed* scope, and a report
that does not say so reads as complete coverage.

**Why it walks descendants rather than asking `bd`.** Three queries look like
they answer "what is in this epic" and two of them lie:

| query | what it actually returns |
|---|---|
| `bd epic status` | direct children only — on a nested epic, an answer about sub-epics |
| `bd list --parent` | direct children only, same trap |
| `bd ready --parent` | transitive, but omits closed beads — so it shrinks as the run succeeds |

None of the three is "every bead in this run", which is why the script walks
the descendants itself.

## The `Stop` hook

The ready-loop is enforced by the harness rather than remembered. The plugin
declares a `Stop` hook in `hooks/hooks.json` at its root, running
`hooks/bb-stop-ready-loop.sh`, that exits **2** — which continues the
conversation — while the armed run epic still has ready beads.

**It is inert until armed, once per run:**

```sh
"${CLAUDE_PLUGIN_ROOT}"/hooks/bb-stop-ready-loop.sh --arm <run-epic>  # at the first dispatch
"${CLAUDE_PLUGIN_ROOT}"/hooks/bb-stop-ready-loop.sh --status          # armed, or inert?
"${CLAUDE_PLUGIN_ROOT}"/hooks/bb-stop-ready-loop.sh --disarm          # when the run drains
```

**Arming writes a file outside the repository; it is not an environment
variable, and it cannot be.** A hook is spawned by the harness, not by your
shell, so the `BB_*` facts pre-flight emits for you to `eval` never reach it.
An epic id it cannot read from disk is an epic id it does not have — and that
is not a cosmetic difference, because `bd ready --parent ""` is repo-wide
rather than empty. A hook falling back to an unset variable would block on
unrelated work, forever.

**The bound: at most 8 consecutive blocks per arming**
(`BB_STOP_HOOK_MAX_BLOCKS`), after which the hook stands down and says so. It
does **not** reset itself at the cap — only a drained queue, or a fresh
`--arm`, restores the budget. A queue that will not drain is a run that needs a
human, and holding the session hostage is not how it gets one.

**Stopping anyway.** Both forms are printed on every block, so neither is
something to remember:

| | |
|---|---|
| `bb-stop-ready-loop.sh --disarm` | removes this project's state; the hook goes inert |
| `BB_STOP_HOOK=0` | inert everywhere, touching no files |

**Everything that is not "armed, and the queue provably non-empty" exits 0** —
`bd` missing, the query failing, output that will not parse, an empty epic id.
A hook that blocked on its own malfunction would be a session an operator has
to kill, which is a worse failure than the one it prevents.

## The `SubagentStop` hook

The `Stop` hook holds YOU to the ready loop. This one holds a WORKER to its own
queue: the same `hooks/hooks.json` declares a `SubagentStop` hook running
`hooks/bb-subagent-stop-queue.sh`, which exits **2** when a subagent stops
while beads assigned to it are still ready or in progress.

**It moves acceptance from testimony to state.** A batch that returns
"completed" on a bead it never touched is the one failure the whole
verify-against-state apparatus exists to catch, and until now it was caught
*after* the fact, by the coordinator. The harness can now refuse the claim.

**It cannot detect a stall**, and nothing can: a stalled subagent never
finishes, so the event never fires. Absence of an event is not an event. This
catches a worker that stops EARLY, never one that stops never — the three
deaths above are still yours to tell apart.

```sh
"${CLAUDE_PLUGIN_ROOT}"/hooks/bb-subagent-stop-queue.sh --arm      # takes no epic
"${CLAUDE_PLUGIN_ROOT}"/hooks/bb-subagent-stop-queue.sh --status
"${CLAUDE_PLUGIN_ROOT}"/hooks/bb-subagent-stop-queue.sh --disarm
```

**A BLOCKED bead is accepted; an unfinished one is not.** That distinction is
the whole design, because getting it wrong traps a worker forever — refused on
every attempt to return, with no way to satisfy the refusal. It falls out of
two queries rather than a special case, measured on one bead in four states:

| the worker's bead | `bd ready --assignee` | `--status in_progress` | verdict |
|---|---|---|---|
| `open`, never started | 1 | 0 | **refused** |
| `in_progress` | 0 | 1 | **refused** |
| `blocked` (the `run-bead` Blocked path) | 0 | 0 | allowed |
| `closed` | 0 | 0 | allowed |

`bd ready` excludes blocked beads and `--status in_progress` does not match
them, so a worker that declared its blocker is drained by both. **So the way
out of a refusal is to declare the block, not to argue with the hook** — which
is what the ceremony asks for anyway.

**Scoped from the payload, or silent.** The agent id comes from the
`SubagentStop` payload on stdin and nowhere else: a hook is spawned by the
harness, so nothing you export reaches it, and a single id on disk could not
say which of several concurrent workers stopped. With no id in the payload the
hook exits 0 and says so. It must never guess — measured,
`bd ready --assignee ""` returned 3 rows and
`bd list --assignee "" --status in_progress` returned 2 on a tracker where the
intended agent held none.

**The bound is 3 consecutive blocks PER AGENT**
(`BB_SUBAGENT_HOOK_MAX_BLOCKS`), lower than the `Stop` hook's 8 because a
worker cannot dispatch its way out of a queue the way a coordinator can. Per
agent, so one wedged worker cannot spend another's budget. As with the other
hook it does **not** reset at the cap. Kill switches: `--disarm`, or
`BB_SUBAGENT_HOOK=0`.

## Running without the hooks

**Every hook this plugin ships is a convenience over a path that still works
without it.** An adopter may run with the plugin disabled, on a harness that is
not Claude Code, or on one whose hook surface differs. Nothing in this
ceremony may depend on a hook having fired.

WHEN no hook is installed, the loop is yours to run: after every batch returns, re-query
the ready front by the two steps above and stop only when it is empty. That is
the identical check, made by you instead of by the harness. `--status`, or the
absence of the script, says which of the two you are in.

**Same for the `SubagentStop` check.** With no hook, run its two queries
yourself as each batch returns, before you accept anything it reported:

```sh
bd ready --assignee <dispatch-id> --json                        # must be empty
bd list  --assignee <dispatch-id> --status in_progress --json   # must be empty
```

Both empty is the DRAINED test the dispatch procedure already states, so this
adds no new rule — the hook only makes it non-optional. **A bead the worker set
`blocked` satisfies both**, which is why "drained" and "every bead closed" are
different questions and only the first one gates acceptance.

## The review loop's budget

**2 loops — three passes** unless the authorization gate said otherwise.

**Findings from the human's pull-request review do not consume it.** The
counter bounds *the machine reviewing itself*; a human's findings are a
different channel and charging them to the same budget would let one human
comment exhaust the run's ability to re-review.

**It lives in the manifest as `review_passes`, never in the orchestrator's
head.** A resumed session that restarted the counter would re-enter the loop
from zero and never exhaust its budget.

**Exhausting it is not a failure.** What is left stays filed and is named in
the report. A third pass that still finds things is information for the human,
not a reason to keep going.

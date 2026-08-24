# Supervising a run, and deciding it is done

Two questions the orchestrator asks repeatedly and must not answer from
recollection: *where is this run right now*, and *is it finished*. Both are
computed from state.

## Contents

- Where is this run — the durable view, and what it cannot see
- When something dies — which of three situations you are in
- When the run is done — the completion rule, and why the obvious query is wrong
- The review loop's budget — what the counter bounds, and what it does not

## Where is this run

**`bd swarm status <epic>`** is the durable answer, computed from the beads, so
it survives any crash and any lost session.

**On a NESTED epic it is one level deep** and reports sub-epics rather than the
leaves under them. Where the run epic is nested, fall back to the re-queried
ready front plus a walked descendant count — and say which of the two you used,
because the numbers differ and a reader cannot tell them apart.

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

Live progress — the event stream, the per-wave `Monitor`, and the durable
out-of-session trigger — is `progress-streaming.md`.

## When something dies

Take the one you are in; they have different remedies and reading the wrong one
wastes a wave:

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

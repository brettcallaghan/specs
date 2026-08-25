# The run record — manifest, swarm, amendments

## Contents

- The run's epic shape — ceremony epics chained under the run epic, one flat
  target, and the gate as a task
- The two human channels — authorization at the start, findings at the end
- The two things the coordinator owns: the manifest and the run's amendments
- `collision_edges` — validating and loading the edge file the plan returns
- `base_sha` — why it is recorded at creation and never derived

## The run's epic shape

`sprint-scope` builds it; everything below assumes it. The run epic holds one
epic per phase — pre-flight, scope, plan, TARGET, review, retro, post-flight —
chained `epic→epic` in that order, with the work beads as flat task children of
the TARGET.

**Ceremony phases are epics because nothing else can block the target.**
Measured on bd 1.2.2: `bd dep add <task> <epic>` fails with `tasks can only
block other tasks, not epics`, and `bd dep add <epic> <task>` with `epics can
only block other epics, not tasks`. So a task cannot gate a phase. One
`epic→epic` edge does the whole job: measured on a run epic shaped this way,
`bd ready --parent <target> --exclude-type=epic` returned nothing while the
target waited on the plan epic — **a blocked parent blocks its children** — and
a bead created under the target mid-run came back from `bd blocked` as
`Blocked by 1 open dependencies: [<target>]` with no edge wired for it.

**The TARGET is FLAT — tasks only — and flatness is an ADMISSION rule**, so it
binds every bead added mid-run and not only the ones the plan created. Flatness
is what makes "the run's work" and "the target's children" the same set, so a
count, a results table and a completion verdict all read from one query rather
than from a tree walk that each reader has to get right. **Work that genuinely
decomposes into sub-epics becomes a SECOND target under the run epic**, run
alongside the first rather than nested inside it.

**The queue expression is unchanged by the shape, and this is measured rather
than assumed.** `bd ready --parent <run-epic> --exclude-type=epic` still returns
exactly the target's open leaves: by the time the target opens, the phases
before it are CLOSED (so their task children are too — an epic is only
close-eligible once every child is closed), and the phases after it are still
blocked behind it. Observed on the demonstration run at that moment: three
rows, all three the target's, the mid-run addition among them. So the run epic
stays the one id a recovering session needs.

**The authorization gate is a TASK child of the scope epic.** A gate cannot
attach to a ceremony epic — gating one would need the refused task→epic edge —
and it does not need to: the scope epic already blocks the plan epic, so
holding its task child holds the whole chain. Verified end to end: with
`bd gate create -t human --blocks <gate-task>` open, `bd ready --parent <scope>`
returned nothing; the gate task had entered the ready front the moment
pre-flight's own epic closed.

## The two human channels

**This table is their only home.** Every skill that opens one of these channels
states its own row inline and names this document by role rather than copying
the table; a second copy is what drifts.

Two ways a run reaches a human, separated by **when the run consumes the
answer**:

| channel | filed as | blocks | resolved by |
|---|---|---|---|
| **authorization** | `bd gate create -t human --blocks <gate-task>` on the scope epic's task child | at the START — the gate holds the whole phase chain | `bd gate resolve <gate-id>` |
| **findings** | a `retro:<kind>[:<target>]` comment on the bead | NEVER — recorded, and read at the retro | nothing; it is a record, not a question |

**A run is autonomous because of the second row.** The first is taken once,
before any work; everything after it is decided by the agent and recorded. A
third channel — stopping to ask mid-run — is the thing this design exists to
make unnecessary.

**A decision that changes what "correct" means is a finding, not a third
channel.** It carries its ordinary kind — `decision`, `scope`, `gap`,
`sequencing` — and routes by that kind like every other. Marking it with an
extra severity word instead routes it nowhere: the kind slot accepts only the
nine, and a word anywhere else is found only by whoever greps for it.

**A bead that genuinely needs a human's answer is BOUNCED, not deferred.**
Grooming raises it with `bd create --add-label human` (`--labels` exits 1 with
`unknown flag`) and the run goes narrower without it. Two measured reasons not
to build a wait-until-the-end channel on top of `bd human`: `bd human respond`
and `bd human dismiss` **both** fail with `storage is nil`, and `bd human list`
keeps listing a bead after it closes — so anything asking "does a human still
owe this run an answer" must read `bd list --label human --status open`
instead. The dispatch front carries `--exclude-label human` so a bounced bead
is never handed to an agent; no completion count carries it, or the run reports
itself drained while a human still owes it an answer.

**Closing a ceremony epic is `bd epic close-eligible`, scoped BY HAND.** It
takes no positional epic. Measured: with one run's pre-flight child closed, the
bare command printed `✓ Closed 2 epic(s)` — pre-flight **and an unrelated epic
from another run in the same database**. Read what it names and confirm every
line belongs to this run.

## The two things you own

**Two things the coordinator owns and never delegates.** Read this once,
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
   off the manifest to build each dispatch prompt. A renamed key is a silent
   break: the reader gets `null` and dispatches without it. The
   full key set is the JSON block the **`sprint-plan`** skill writes, in its own
   run-record reference.

   **`collision_edges` is yours to write and the planner's to produce.** The
   plan returns the path of one `bd dep add --file` document: `blocks` edges
   between beads that must not run at once. **It creates no beads** — the work
   beads already exist by the time planning starts, and the ceremony epics and
   the target were built by `sprint-scope` — so an edge file is the whole of
   what a plan contributes. Load it, then record the path here, so a recovering
   session can diff what the run was supposed to be against what the graph now
   holds.

   **Diff it and re-run the plan's validators before loading**, because a load
   that fails partway leaves the graph half-wired and being reviewable
   pre-load is most of what the document is for. Then **assert the ready front
   after loading** rather than trusting the file: an edge written the wrong way
   round inverts a dependency at exit 0.

   **Merge, never replace.** Pre-flight writes the baseline into the same
   object, and the epic already carries grooming's own keys. Read it, update
   your keys, write it back:
   `bd show <epic> --json | jq '.[0].metadata'` → edit → `bd update <epic> --metadata @file.json`.

   Recovery reads the epic rather than re-deriving the run, and the epic
   outlives a reclaimed container where a `/tmp` path does not.

2. **Amendments — `amend:`-prefixed comments on the run epic.** WHEN the run
   diverges mid-flight — an edge added, a batch reclassified, a bead
   admitted or bounced, a tier escalated — `bd comment <epic> "amend: <what
   and why>"`. Append-only and timestamped by construction, so there is no
   read-modify-write to lose and no ordering to maintain. Never re-invoke
   `sprint-plan` to re-partition a run in flight.

   **The dispatch order is not stored anywhere.** It is
   `bd ready --parent <epic> --exclude-type=epic` re-queried, so a stored copy
   can only drift out of agreement with the beads it describes.

**The manifest is written on every run** — there is no queue-size or
batch-count exemption. Chat is the one store that does not survive.

Print the returned plan to chat. Don't ask for approval — just print and
proceed. The user can interrupt if they disagree.

WHEN no `Agent` tool is available, nothing can be delegated, planning included
— that is a blocking pre-flight condition, not a mode. Stop and hand the queue to
the `run-bead` skill, one bead at a time.

## `base_sha` is recorded at creation, never derived

Because that is the only moment it is unambiguous. Deriving it later from a
merge-base assumes the base branch has not moved, and on a long run it will
have — leaving the review scoped to a diff that silently excludes whatever
landed on the base meanwhile.

Three things consume it and none may create it:

- the review scopes its diff `base_sha`..sprint head
- every worktree branches from sprint HEAD
- the pull request opens against it

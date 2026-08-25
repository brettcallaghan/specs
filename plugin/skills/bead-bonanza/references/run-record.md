# The run record — manifest, swarm, amendments

## Contents

- The run's epic shape — ceremony epics chained under the run epic, one flat
  target, and the gate as a task
- The two human channels — authorization at the start, findings at the end
- The two things the coordinator owns: the manifest and the run's amendments
- The run manifest — every key, and why the key names are a contract
- The environmental ledger — admitted on evidence, ordered, never reordered
- Amendments — `amend:` comments on the epic, append-only by construction
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
   break: the reader gets `null` and dispatches without it. Every key is
   specified under "Run manifest" below.

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
— that is a blocking pre-flight condition, not a mode. Stop and hand the queue
to the `run-batch` skill.

## Run manifest

`--metadata` JSON on the run epic, **written by the orchestrator itself** from
`scripts/pre-flight-check.sh`'s output plus what admission decided. Retrieve it
whole with `bd show <epic-id> --json | jq '.[0].metadata'`.

**`branch` and `base_sha` are the orchestrator's, written once.** It creates
the sprint branch immediately before its first dispatch — after grooming, so
`base_sha` includes the grooming commits — and writes both keys then. A run
whose manifest already carries them **reuses** the branch and creates nothing;
that is what makes a resumed run land where the original did. Three things read
them and none may write them: the review scopes its diff `base_sha`..sprint
head, every worktree branches from that branch's HEAD, and the pull request
opens against it.

```json
{
  "target": "<the flat target epic; its task children are the work>",
  "branch": "<branch>",
  "base_sha": "<sha at branch creation, after grooming>",
  "checks_ids": ["format", "vet", "build", "test-go", "…"],
  "checks_baseline": "<verbatim text — inline this into EVERY subagent prompt>",
  "checks_baseline_status": 1,
  "concurrency_cap": 20,
  "models": {"light": "…", "standard": "…", "advanced": "…", "frontier": "…"},
  "effort": "low|medium|high|xhigh|unrecorded",
  "standing_auth": {"<row>": {"value": "<resolved>", "source": "sf.toml|default"}},
  "trailers": "<commit trailers, or absent when the run resolved none>",
  "admitted": ["<bead-id>", "…"],
  "excluded": [{"id": "<bead-id>", "reason": "declined move out of <epic-id>"}],
  "moved": [{"id": "<bead-id>", "from": "<previous parent>"}],
  "partition_preview": "accepted | overridden — <why> | none (no C4 docs) | not returned",
  "review_passes": 0,
  "review_budget": 2,
  "collision_edges": "<path to the edge file this skill emitted>",
  "findings_path": "<sprint-retro's run record — the run's one document>",
  "retro_filed": 0,
  "retro_declined": 0,
  "retro_routed": 0,
  "ledger": ["<claim> — proved by <command>, observed <result>"]
}
```

**The four retro keys are `sprint-retro`'s, written at the end of the run.**
They are what lets a session that did not hold the run find its record from the
epic id alone. **An absent `findings_path` means the retro did not run** — a
different result from a retro that found nothing, and one a reader must not
have to guess at.

**Merge, never replace.** The epic already carries grooming's keys and
pre-flight writes the baseline into the same object: read it, update your keys,
write it back.

**`review_passes` is durable for one reason:** a resumed run that restarted the
counter would loop forever. `review_budget` records what the authorization gate
decided, so a run resumed by a session that never met the human still honours
it.

**`standing_auth` records each gate row's SOURCE as well as its value.** A row
resolved from `sf.toml`'s `[run]` and a row that took the skill default are
different facts, and the record exists to stop a decision being re-litigated
mid-run — a project-level default is one more place one can hide from whoever
reads the report.

`admitted` / `excluded` / `moved` are the admission record. A run that finished
its admitted set while beads sat out is done, and must say so rather than
reading as complete over a scope nobody remembers narrowing.

## Environmental ledger

Facts established this run that no dispatch should re-diagnose. Admitted on
evidence, not on reasoning: every entry — including a claim of sufficiency
("this check verifies X"), not only a claim of impossibility — names the
command that proved it and the result observed; run it first, record it
after. Note the conditions the command ran under when they could change the
result (attached repos, credentials, cache warmth, toolchain version). A
claim that cannot be tested cheaply is recorded as `unverified` with what
would settle it, not omitted. Keep the ORDER — dispatch prompts cite entries
by position ("ledger entry 1", "ledger entry 2"), so append rather than
reorder.

1. `<claim>` — proved by `<command>`, observed `<result>`
1. `<claim>` — unverified; would be settled by `<command>`

A gate recorded `fail` in the baseline is the bar for every bead in this run:
"no findings beyond the baseline", not "exit 0". A gate with no recorded
baseline is expected to exit 0.

## Amendments

Amendments are **`amend:`-prefixed comments on the run epic**. WHEN the run
diverges mid-flight — an edge added, a batch reclassified, a bead admitted or
bounced, a tier escalated — `bd comment <epic-id> "amend: …"`.

Append-only and timestamped by construction, so there is no read-modify-write
to lose and no ordering to maintain. There is no second copy to keep in step:
the graph is the plan, and the tracker is the only record.

```
bd comment <epic-id> "amend: collision edge added demo-9kb.9 <- demo-9kb.7 —
.7's change to the shared fixture invalidates .9's expected bytes, so they
cannot share a checks run."

bd comment <epic-id> "amend: demo-9kb.8 escalated standard -> advanced after
two failed checks cycles."
```

Read them back with `bd show <epic-id>`, or
`bd export | jq -r 'select(.id=="<epic-id>") | .comments[].text' | grep '^amend:'`.

## `base_sha` is recorded at creation, never derived

Because that is the only moment it is unambiguous. Deriving it later from a
merge-base assumes the base branch has not moved, and on a long run it will
have — leaving the review scoped to a diff that silently excludes whatever
landed on the base meanwhile.

Three things consume it and none may create it:

- the review scopes its diff `base_sha`..sprint head
- every worktree branches from sprint HEAD
- the pull request opens against it

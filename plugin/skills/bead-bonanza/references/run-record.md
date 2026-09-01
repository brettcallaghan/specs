# The run record — manifest, swarm, amendments

## Contents

- The run's epic shape — ceremony epics chained under the run epic, one flat
  target, and the gate as a task
- The two human channels — authorization at the start, findings at the end
- The two things the coordinator owns: the manifest and the run's amendments
- The run manifest — the four `swarm manifest` verbs, and where the key set
  and its meaning are generated to
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
`sequencing` — and routes by that kind like every other.

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

1. **The run manifest — the record `swarm manifest` reads and writes.**
   Written **before** you dispatch grooming, and it carries what pre-flight
   measured plus what admission decided.

   **The command owns the key set, the types and the merge semantics**, and
   this document does not restate them: run `swarm manifest get --run <epic>`
   and read what it prints, `--help` on any verb names the flags, and `swarm
   manifest validate --run <epic>` reports every rule the record breaks at the
   stage it has reached. A key nothing here mentions is still enforced there.
   `generated/manifest.md` is the same key set and meaning as a document,
   generated from the CLI's own model, for reading without a run in flight to
   point a command at.

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

   **One `swarm manifest set` per write, however many keys it carries.** The
   command merges rather than replaces, so pre-flight's baseline and grooming's
   own keys survive a write of yours — and keys it does not know are preserved
   rather than dropped. Splitting one logical write across several invocations
   is what leaves a run half-recorded.

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

**Written by the orchestrator itself** from `scripts/pre-flight-check.sh`'s
output plus what admission decided, and read and written through one command:

```sh
swarm manifest get      --run <epic>                       # every key the record carries
swarm manifest get      --run <epic> --key <key>           # that key and its value
swarm manifest require  --run <epic> --key <key> …         # are these written? routes on the status
swarm manifest set      --run <epic> --key <k> --value <json> …   # one merge, however many keys
swarm manifest validate --run <epic>                       # every rule, at the stage reached
```

**Run one and read it.** The output is YAML, or JSON under `--json`, and it
needs no key to decode. Three things it cannot tell you by being read:

- **Route on the exit status, which separates the three answers** — the values
  (0), a declared key nobody has written (1), a key the manifest does not
  declare (3). One invocation shows you one of them.
- **`--full-result` is the way back to the provenance** — the summary, the
  facts and the invocation behind each — when the values are not enough.
- **To capture one value in a shell**, ask for JSON and read the key:

  ```sh
  br="$(swarm manifest get --run <epic> --key branch --json | jq -r .branch)"
  ```

**Four consequences worth knowing before you write one**, because each is a
state you would otherwise have to guess at:

- **A key that is absent reads as absent**, never as an empty value. "Not yet
  written" and "written as nothing" are different states of a run, and
  `require` is what turns the difference into a status you route on.
- **A key the command does not know is preserved**, so a write of yours never
  deletes what another component recorded.
- **A refused write reports every reason and stores none of them** — so fix
  them together rather than re-running to find the next.
- **`branch` and `base_sha` are written once**, and re-writing the same value
  is not a change. That is what lets a resumed run reuse its branch: it writes
  the pair it read and the guard passes. Three things read them and none may
  change them — the review scopes its diff `base_sha`..sprint head, every
  worktree branches from that branch's HEAD, and the pull request opens against
  it.

**What the keys MEAN is `generated/manifest.md`'s, not this
document's.** It is generated from `schemas/swarm/manifest.strata` by
`scripts/generate-schemas.sh`, kept current by that script's `--check`
(declared check `schemas`), and lists every key, its type, and the stage its
struct declares it at, alongside the same docstring the CLI's model carries.
It is vendored rather than left as "run the command": the key set is read
while deciding whether a run is worth inspecting at all, before any epic
exists to point a command at.

Three facts about how a key is USED are not the schema's to carry, because
they describe this skill's own machinery rather than the manifest:

- **`checks_baseline` is inlined verbatim into every subagent prompt.** A gate
  recorded failing there is the bar for the run — "no findings beyond the
  baseline", not "exit 0".
- **`review_passes` is what a resumed run reads to avoid restarting the review
  counter**, honouring `review_budget` even in a session that never met the
  human at the authorization gate.
- **`bb-close-bead.sh --trailers-from <epic>` reads `trailers` from the
  manifest itself**, before it closes or commits anything, and exits 2 saying
  so when the run resolved none.

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
reorder; `swarm manifest set` refuses a ledger write that removes, reorders or
alters an entry, because a rewrite retargets every citation that addressed one
and nothing downstream can tell.

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

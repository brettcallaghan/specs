# Sprint findings — run epic `agent-plugins-mpnf`

Research, not a decision. This document records every finding the run captured,
verbatim, and what routing and clustering produced from them. Nothing here is
adopted until a bead says so; the beads this retro filed are named in
[§7](#7-candidates-and-what-was-filed) and each cites the findings it rests on.

It lives under `docs/research/` rather than `rfcs/sprints/` because it decides
nothing — per this repository's rule for where documents live, a document that
would still be true had the decision gone the other way is research. It carries
no corpus slug and is cited by path.

## 0. The run

| | |
|---|---|
| run epic | `agent-plugins-mpnf` — "Run 3: adopt the beads graph (plan items 6-14), and the corrections that ride with them" |
| branch | `claude/bead-bonanza-design-review-ne1plu` |
| base → head | `653e885` (2026-08-24 20:48 UTC) → `5cc2a58` (2026-08-25 07:05 UTC), 25 commits |
| queue | 23 admitted beads; 27 closed across the run's whole reachable set, 11 of them direct `mpnf.N` children |
| deferred | 15 `mpnf.N` children, all review findings — refused at admission, or admitted and then de-admitted by owner decision as the branch wound up |
| review passes | 2 of a budget of 2 — pass 2 filed twelve findings, eleven admitted, two deferred |
| checks | all 13 declared ids green at head (`format vet build test-go skills shared-refs strata-docs schemas corpus ref-links template-drift script-self-tests skills-strict`); `bb-status.sh run agent-plugins-mpnf` exits 0 |
| model | `claude-opus-5` for this retro. **The run manifest records no `models` and no `effort` key**, so the lineup for the beads themselves is unrecoverable — already filed as `agent-plugins-y63n.10` |

**This run rewrote its own machinery while using it.** The `bead-bonanza` skill
family was both the subject and the tooling. Five of the eleven closed children
edit scripts the coordinator was running at the time.

## 1. Counts

| | |
|---|---|
| findings harvested | **52** — 49 captured in flight, plus 3 recorded by this retro while running (`mpnf#9`, `#10`, `#11`) |
| routed to the **skill** audience (`tooling`) | 8 |
| routed to the **project** audience (the six work-order kinds) | 20 |
| terminated at `proof` | 17 |
| terminated at `environment` | 5 |
| **unrouted** | **2** |
| incidents (a severity, carried through routing) | 8 — `mpnf#1` `#3` `#4` `#5` `#6`, `mpnf.4#3` `.4#7` `.17#2` |
| candidates drafted | 14 |
| **filed** | **12** (8 skill-audience, 4 project-audience) |
| **declined / deduped into an existing bead** | **2 declined, 7 deduped** |

Harvest query, over the run's closure set:

```sh
bd export | jq -r 'select(.id|test("^agent-plugins-mpnf")) | . as $i
  | [(.comments//[])[] | select(.text|startswith("retro:"))]
  | to_entries[] | "\($i.id)#\(.key+1)\t\(.value.text)"'
```

Finding ids are `<bead-id>#<ordinal>`, the ordinal being the finding's position
among that bead's own `retro:` records. Derived, never authored.

### Where the findings were

`mpnf` (the epic, coordinator-recorded) 11 · `.4` 7 · `.18` 6 · `.24` 6 ·
`.20` 5 · `.16` 4 · `.17` 4 · `.3` 4 · `.2` 2 · `.12` 2 · `.5` 1.

Eleven beads carried findings; four of the fifteen deferred beads carried none,
which is expected — they were never worked.

## 2. Two channels this retro did not use, and why

**No `.sf/outbox/` files and no `rfcs/sprints/` document.** Commit `9cc8844`
("Drain the outbox and the run-3 sprint docs into the tracker") removed 15
outbox issue documents and the four 2026-08-24 sprint documents at the owner's
instruction, harvesting each into a bead under `agent-plugins-y63n` first.
`.sf/outbox/README.md` survives but the channel is not in use. Skill-audience
findings from this run were therefore filed as **beads under
`agent-plugins-y63n`**, one per candidate, each carrying the evidence an outbox
file would have carried.

That substitution is itself a finding — `mpnf#10`. The skill hard-codes both
destinations as literal strings and neither `sf.toml` nor the run manifest
carries a key that could redirect them, so this retro had to be told in prose
to override its own skill, twice.

**No candidate list was put to the owner and waited on.** The owner is winding
the branch up and had already taken the governing decisions this retro would
have asked about. Everything clearly evidenced was filed; §8 lists the two
things that genuinely still need a decision.

## 3. Routing

A lookup, applied without reading the findings' content:

| kind | destination | n |
|---|---|---|
| `tooling` | skill audience | 8 |
| `gap` `decision` `stale` `unverifiable` `sequencing` `scope` | project audience | 20 |
| `proof` | this document, beside what it supports. No issue. | 17 |
| `environment` | the environmental-failure ledger. No issue. | 5 |
| kind absent or unknown | **UNROUTED** | **2** |

Two findings carry a `tooling` kind but a target this repository owns rather
than vends — `mpnf.4#5` against `scripts/check-skills.sh`, and `mpnf#11`
against `specs/standards/findings-capture-protocol.md`. The merge test settles
both: the fix lands in this repository, so both were filed as project beads.
The lookup is the default; the merge test overrides it where the target is
visibly local.

### The unrouted count has been 2 for three consecutive runs

`agent-plugins-y63n.17` already records the first two. This is the third, and
the mechanism has changed:

| run | unrouted | invented kinds |
|---|---|---|
| `agent-plugins-qozp` | 2 | `retro:process:` |
| `agent-plugins-so39` | 2 | `retro:process:` (both the coordinator's) |
| `agent-plugins-mpnf` | 2 | `retro:process:` (`mpnf#5`) and **`retro:incident:`** (`mpnf#6`) |

`retro:incident:` is not a made-up word. The protocol names *incident* as a
first-class concept, says "Incidents are a severity, not a kind", and shows the
correct form — but never states the wrong one. An author who has just read that
section is the most likely person to write `retro:incident:`.

**The cost this run was concrete: the single most consequential finding of the
run routed nowhere.** `mpnf#6` is the only record of a state the corpus asserts
cannot occur, and it names two remedies against `bead-bonanza`. The routing
table sends it to "a count and a list". It is filed here as two candidates
(F3, F4) because a human read it, which is not a mechanism.

## 4. What the run actually shows

### 4.1 The swarm works; the substrate is what fails

Three runs, 45+ beads closed, each landing as one commit with the full declared
check set green. Not one finding in this run is about a subagent doing the
work badly. Every friction recorded is in the machinery *around* the work —
routing, prompts, worktree lifecycle, acceptance criteria, tool claims. That is
the case for `agent-plugins-offx`: a thin, tested capability layer over `bd`
and git inside the existing harnesses, rather than a better orchestrator.

### 4.2 The run's own liveness machinery reported green through its longest outage

Loop-3 grooming returned at 01:00 with eleven beads Ready and assigned. Nothing
was dispatched. The session was idle until 06:09 — **five hours** — and the
run's health label read `healthy` for all of it. `agent-plugins-mpnf.1` is a
closed `event` bead titled "State change: health → healthy"; when the
coordinator last set it, `bd set-state` replied *(no change: health already set
to healthy)*.

What woke the session was the owner's PR comments. No liveness mechanism fired.
The recovery trigger armed for exactly this slept through it.

The point is narrower than "health was wrong". **Health was asserted by the
coordinator, and the coordinator was the thing that had stopped.** A signal
whose only writer is the component it monitors reports the absence of a failure
and the absence of a monitor identically. The derived signal already existed
and was already correct — `bb-status.sh run <epic>` prints
`health: healthy (worktrees: 0, last commit: 2 minutes ago)`, and *worktrees: 0
plus a stale last commit plus a non-empty ready front* is a stall computable by
anyone with no cooperation from the coordinator. The parenthetical was doing the
work; the word in front of it was the part that lied.

Filed before this retro as `agent-plugins-offx.11` (derive health) and
`agent-plugins-offx.15` (the trigger fleet). Not re-filed.

### 4.3 A false tool claim survived four independent chances to die

The claim: "`bd ready --json` returns `null` when empty, but `[]` for an
`--assignee`-filtered empty query." It is false on bd 1.2.2.

| # | chance | outcome |
|---|---|---|
| 1 | a loop-2 **grooming agent** recorded it as a ledger entry, citing "the run's existing ledger records `null`" | admitted. The coordinator's later sweep of the epic ledgers on `mpnf`, `qozp` and `so39` found **no such ledger entry** — the citation named a surface that does not carry the claim, and nothing checked it |
| 2 | the **coordinator** relayed the whole entry into three dispatch prompts | unmeasured. This is hours after the same coordinator shipped `agent-plugins-x5tk`, the AGENTS.md section forbidding exactly this |
| 3 | **two of three receiving agents** (batches 9 and 10) | neither challenged it. Nothing in either bead depended on it, so silence cost nothing here — but silence is what it was |
| 4 | **`dev-mpnf-8`** (batch 8) | caught it. Ran four configurations, got `[]` in all four, and flagged the contradiction rather than working around it — `mpnf.3#2` |

`mpnf.18`'s agent then re-measured the *replacement* before writing it: eleven
configurations, a positive control, `null` in none of them (`mpnf.18#1`).

**What would have caught it earlier.** Not a rule — the rule already existed and
its own author broke it within hours of shipping it. AGENTS.md says so about
itself: *"the rule has no automated check; it is enforced by authorial memory,
which is the condition it exists to correct."* The three things that would
actually have bitten, in order of how cheaply they bite:

1. **A ledger entry that cites another document must name it precisely enough
   to be checked, and be checked.** Chance 1 is the cheapest kill: one grep
   against the named ledger, at admission, by the agent doing the admitting.
   This is the only one of the four that was a *fabrication* rather than a
   relay.
2. **An executable claim, not a prose one** — `agent-plugins-offx.4`, turning
   prose invariants into guards that exit non-zero. A claim about `bd`'s output
   shape is a two-line test.
3. **A dispatch prompt is an authoring surface with no reviewer.** Chance 2 is
   the highest-reach copy and the only one nothing audits. `mpnf#4` is the same
   defect in the same run on a different axis — a prompt whose Scope list and
   whose prose named different file sets, with nothing reconciling them.

Recorded as comments on `agent-plugins-offx.4`; no new bead, because the
mechanism is already scoped there.

### 4.4 Empty-JSON shape is per-subcommand, which is how the false universal formed

`mpnf.18#2`, measured in the same session and the same scratch database:

```
$ bd human list --json     ->  null   [exit 0]
$ bd ready --json          ->  []     [exit 0]
```

So somebody who measured `bd human list` and generalised to "bd returns `null`
when empty" would have produced exactly the retracted claim. The coordinator's
retraction comment had already fenced the `bd human list` claim off as
"DIFFERENT COMMAND — not retracted here, and not tested by me"; `mpnf.18`'s
agent then tested it and it is **true**, and must not be swept.

This is the generalisation failure worth naming: **a tool's behaviour measured
on one subcommand is evidence about that subcommand.** The value of `mpnf.18#3`
is the same lesson one level up — swapping `null` for `[]` in the corpus would
have left five sites of defensive code justified by a measurement that no longer
supports them, and a reader with grounds to delete the guards. Each replacement
was re-grounded on a hazard the agent measured instead (`jq -r length` on a
string payload returns 16, not 0 — *that* is what the `type=="array"` guards
earn their keep against). Filed as candidate P3.

### 4.5 Line-number rot, inside the run that documents it

`groom-beads` names unquoted `file:line` citations as an anti-pattern.
`agent-plugins-61f.4` closed that rule into the Definition of Ready and the
work-order template on 2026-08-13, citing a prior run where **every** cited
position had moved.

`mpnf.18` was filed by `sprint-review` in pass 2 with six line numbers. By
dispatch two had moved (`mpnf.18#4`):

```
bb-status.sh:459   ->  :560   (moved by sibling mpnf.17, in the same batch)
bb-stop-ready-loop.sh:213 -> :228
```

The description also said "five NEW copies plus a sixth"; the measured count was
**seven sites in five files**.

It survived for two reasons, and only one of them is a mechanism. The agent
matched by **string** rather than by line, which is the durable handle. And
`mpnf.17`'s agent volunteered `mpnf.17#4` — a finding about a *sibling's* bead,
recording that its own edit had moved a line the sibling cites. That is
excellent behaviour and it is not a control.

The rule exists; `sprint-review`'s finding→bead path does not apply it. Filed
as candidate F5.

### 4.6 An acceptance criterion named a tool that does not exist here

Two agents, independently, on two beads, hit the identical criterion — *"`sh -n`
and `shellcheck` (as `scripts/ci.sh` runs them) report nothing new"*:

```
$ grep -n 'shellcheck' scripts/ci.sh sf.toml   -> no output
$ grep -n 'sh -n\|bash -n' scripts/ci.sh sf.toml bb-checks.py  -> no output
$ command -v shellcheck  -> not found
```

Every clause of it is false: `ci.sh` runs neither, and `shellcheck` is not
installed. Four findings record it — `mpnf.16#4` and `mpnf.20#5` as
`unverifiable`, `mpnf.17#3` and `mpnf.24#6` as `environment`.

**Both agents recorded it UNVERIFIED rather than claiming a pass**, substituted
what was available (`sh -n`, `bash -n`, both exit 0), and said what would settle
it. That is the correct handling and it is worth naming as such. But grooming
wrote an acceptance criterion against a tool the environment does not have, and
nothing between grooming and dispatch checked.

Filed as two candidates: P2 (declare `shellcheck`, or stop naming it) and F8
(grooming must check a named verification tool against the environment).

### 4.7 The scope narrowed after admission

Pass 2 filed twelve findings, all classified in-scope by the reviewer, none
pre-admitted. The coordinator admitted **eleven** and deferred two
(`mpnf.25` cosmetic, `mpnf.26` the byte-cap category the owner had explicitly
parked). Of the eleven, **five closed and six were subsequently de-admitted by
owner decision** as the branch was wound up — `mpnf.15`, `.19`, `.21`, `.22`,
`.23` and `agent-plugins-p567` are deferred, not closed.

That is a real narrowing and the report must say so: the run's final tree fixes
five of the twelve defects its own second review pass found in its own diff.

The mechanism underneath it is `admit_review_findings`, and the manifest is
blunt about it:

> "NOBODY CHOSE THIS. `sf.toml` leaves the key commented out and the human did
> not answer the row, so it took the skill default. In runs 1 and 2 this same
> silent default declined EVERY review finding, including defects in those runs'
> own diffs."

Overridden twice in this run, on the owner-approved principle that a defect the
run's own merged diff introduced is in scope regardless of the default. **The
row has never been chosen by anybody across three runs and has been overridden
in the two most recent.** A defaulting mechanism whose default is overridden
every time it is reached is not a default. Already filed as
`agent-plugins-offx.8`; this run's data is recorded there as a comment.

### 4.8 A state the corpus says cannot occur

`mpnf#6`, review loop 2, batch 10. `git merge --ff-only bb/dev-mpnf-10`,
`git worktree remove` and the acceptance check were issued as one chained
command. The merge exited **128** ("Not possible to fast-forward" — batch 9 had
already moved the head). **The removal ran anyway, and succeeded.**

For ~90 seconds a removed worktree coexisted with an unmerged branch.
`worktree-mechanics.md` makes worktree presence *the* in-flight signal — "you
remove it as the final act of accepting, after the last of its branches merges".
`bb-status.sh bead` reported it faithfully and exited 2:

```
OK   3 -> STOP 3 NOT merged into HEAD
OK   4 worktree: gone
```

Nothing was lost; a rebase onto `e2b69ad` fixed it. But the check is only
meaningful because the coordinator is *supposed* to establish signal 3 before
signal 4, and **nothing enforces that ordering** — a shell without `&&` between
two commands is enough to break it. Two candidates, F3 and F4.

## 5. `environment` — for the ledger

Five findings terminate here. `agent-plugins-r9yj` is the open bead that owns
appending this class to `AGENTS.md`, and its own evidence pointer now dangles
(it cites a deleted sprint document). This run's entries were added to it
inline, verbatim, so the bead is self-contained:

- **`bd`'s empty-JSON shape is per-subcommand.** `bd human list --json` returns
  `null`; `bd ready --json` returns `[]`. Same database, same session, both exit
  0. Neither shape generalises to another subcommand. (`mpnf.18#2`, `mpnf#7`)
- **`bd` distinguishes "nothing to report" from "cannot answer" by exit status,
  not by payload.** An unanswerable query exits 1 and writes to stderr, so a
  command substitution captures the empty string — indistinguishable from an
  empty answer on content alone. This is what makes an ambiguity guard
  implementable at all. (`mpnf.16#2`)
- **`bd dep add` refuses to UPGRADE an edge type.** `related` does not silently
  become `blocks`; it errors with *"already exists with type `related`
  (requested `blocks`); remove it first"*. (`mpnf#1`)
- **`shellcheck` is not installed in this container and no declared check
  invokes one.** The 13 declared ids are `format vet build test-go skills
  shared-refs strata-docs schemas corpus ref-links template-drift
  script-self-tests skills-strict`. (`mpnf.17#3`, `mpnf.24#6`)

A fifth durable fact, from `mpnf.24#3`: **the harness's hook environment cannot
be observed from a session.** What `CLAUDE_PROJECT_DIR` carries inside a
real harness-fired hook invocation is unverifiable from inside a session, and
the bead's fix worked by removing the dependency on it rather than by settling
it. That is the right move when a claim cannot be measured: make it not matter.

## 6. What to reinforce

A retro that records only friction misses what worked. Four agents did the hard
correct thing without being asked:

- **`mpnf.20`'s agent mutation-tested its own new assertions, three ways.**
  Dropped `--all` from the walk; stopped pushing epic children onto the
  worklist; made `scope_beads` filter nothing. Each mutation broke exactly the
  assertions it should and no others, then was reverted. Without that, a new
  test suite that passes proves nothing about whether it can fail.
  (`mpnf.20#1`)
- **`mpnf.17`'s agent ran a negative control on its own self-test**, re-running
  the fixed script with the descent made unreachable. Only the two nested
  assertions failed, isolating the traversal depth rather than the helper.
  (`mpnf.17#1`)
- **`mpnf.18`'s agent ran a positive control** across eleven configurations,
  explicitly because "without it, `[]` everywhere is equally explained by a
  broken query" — and bounded its own claim to "measured on bd 1.2.2 in every
  configuration tried", never "always". (`mpnf.18#1`)
- **`mpnf.24`'s agent removed a dependency it could not verify** rather than
  guessing at it, and said in the same finding that it made no new claim about
  the thing it could not measure. (`mpnf.24#3`)
- **`mpnf.12`'s agent reverted only the fix hunk** while keeping the new
  assertion, confirmed the assertion failed, then restored and diffed
  byte-for-byte against the pre-revert file. (`mpnf.12#1`)

Two more worth naming because they are *scope* discipline rather than proof
discipline: `mpnf.2`'s and `mpnf.20`'s agents each found an adjacent defect,
**filed it** (`agent-plugins-fvnp`, `agent-plugins-mxl3`) and did not fix it,
because the bead's Design scoped it out. `mpnf.4`'s agent did the same for
`agent-plugins-p567`. That is the behaviour that keeps a bead reviewable as one
unit.

## 7. Candidates, and what was filed

Fourteen candidates were drafted from 52 findings. Twelve were filed as beads;
two were declined outright; seven further clusters deduped into beads that
already exist and were recorded there as comments rather than re-filed.

**No candidate list was put to the owner and waited on**, and no
`.sf/outbox/` file was written — see §2.

### Filed — skill audience, as beads under `agent-plugins-y63n`

| id | P | target | change | findings |
|---|---|---|---|---|
| `y63n.18` | 1 | `sprint-retro` | A filed bead must carry its Evidence bullets inline; 25 beads (21 open) cite a deleted findings document | `mpnf#9` |
| `y63n.19` | 2 | `sprint-retro` | `rfcs/sprints/` and `.sf/outbox/` become project config with the current values as defaults | `mpnf#10`, `mpnf#9` |
| `y63n.20` | 1 | `bead-bonanza` | Worktree removal goes behind `git merge-base --is-ancestor`, not behind a sentence | `mpnf#6` |
| `y63n.21` | 2 | `bead-bonanza` | Name which `--ff-only` refusal case you are in, mechanically, before reading the refusal | `mpnf#6` |
| `y63n.22` | 2 | `sprint-review` | Findings filed as beads cite path + quoted string; a line number is a hint, marked as such | `mpnf.18#4`, `mpnf.17#4` |
| `y63n.23` | 2 | `bead-bonanza` | A dispatch prompt declares its owned-file set once; pre-send greps every path it names against that list | `mpnf#4`, `mpnf.3#2` |
| `y63n.24` | 2 | `groom-beads` | An AC may not contradict the bead's own Design scope, nor pin verbatim output where the file declares a contract | `mpnf.17#2`, `mpnf.4#7` |
| `y63n.25` | 3 | `groom-beads` | Every executable named in an AC resolves on PATH or is a declared check id | `mpnf.16#4`, `.20#5`, `.17#3`, `.24#6` |

### Filed — project audience

| id | P | target | change | findings |
|---|---|---|---|---|
| `agent-plugins-sdu1` | 2 | `findings-capture-protocol.md` | State the wrong form: `retro:incident:` is not a valid finding | `mpnf#11`, `mpnf#6`, `mpnf#5` |
| `agent-plugins-eejf` | 2 | `sf.toml`, session-start hook | Install `shellcheck` and declare it as a check, or stop naming it in ACs | `mpnf.16#4`, `.20#5`, `.17#3`, `.24#6`, `.18#6` |
| `agent-plugins-bfng` | 3 | `AGENTS.md` | Retraction step 4 also re-grounds what the false claim was justifying | `mpnf.18#3`, `mpnf.18#4` |
| `agent-plugins-lmrl` | 3 | `scripts/check-skills.sh` | Rule 2 exempts fenced blocks and an explicit `example:` marker | `mpnf.4#5` |

### Deduped into an existing bead — recorded there, not re-filed

| existing bead | what this run added |
|---|---|
| `agent-plugins-offx.11` | the five-hour stall, and the framing that a signal whose only writer is the component it monitors cannot distinguish "no failure" from "no monitor" |
| `agent-plugins-offx.8` | `admit_review_findings` has never been chosen in three runs and was overridden twice in this one; five of twelve pass-2 findings closed after six were de-admitted |
| `agent-plugins-offx.4` | the four chances a false tool claim survived, and where the cheapest guard sits (chance 1, the fabricated ledger citation — not the point of use) |
| `agent-plugins-y63n.17` | third consecutive run at 2 unrouted; the new invention `retro:incident:` and what it cost |
| `agent-plugins-y63n.10` | the `model` field going in WRONG rather than missing — every issue filed here names the retro's model because the run's is unrecoverable |
| `agent-plugins-r9yj` | this run's five ledger entries, inline and verbatim, plus notice that the bead's own evidence pointer is dead |
| `agent-plugins-p567` | already open for `mpnf.4#4` (the orchestrator loads only the graph plan, so the sidecar never reaches the tracker); deferred by owner decision, not re-raised |

### Declined

- **`mpnf.18#5`, the unreachable copy.** The run epic's comment stream carries a
  pasted self-test transcript containing the old assertion name, and a run record
  is not edited after the fact. The agent flagged it rather than fixing it, which
  is what AGENTS.md's step 5 asks for. **The procedure worked**; there is no
  defect to file. It stays here so a future grep that hits the phrase in the epic
  has something to land on.
- **`mpnf.4#6`, the byte budget spent to 5 bytes free.** The owner explicitly
  parked the byte-cap category ("Ignore byte caps for now — we'll address them
  towards the end of the refactor"). `agent-plugins-offx.7` and
  `agent-plugins-mpnf.26` already hold it.

## 8. What still needs the owner's decision

Two things, neither blocking, both stated here rather than filed as a question:

1. **`agent-plugins-eejf` proposes installing `shellcheck` and declaring it as a
   check.** That is a new gate on every future run, over a body of shell nobody
   has linted, and the baseline is unknown. The alternative — strike `shellcheck`
   from AC text and rely on `agent-plugins-y63n.25` to stop it recurring — is
   cheaper and weaker. This is a bar decision, not an evidence question, and the
   evidence does not settle it.

2. **Whether the six de-admitted review findings stay deferred.**
   `agent-plugins-mpnf.15`, `.19`, `.21`, `.22`, `.23` and `agent-plugins-p567`
   are defects in this run's own merged diff, admitted by the coordinator and
   then de-admitted as the branch wound up. `p567` in particular means an
   orchestrator following the current text creates beads, wires no edges, and
   reports success — the mechanism `mpnf.4` shipped is half-landed. They are
   deferred, not closed, so nothing is lost; but the branch is being wound up and
   the next run will inherit them.

---

## Appendix — every finding, verbatim, grouped by kind

Nothing here is filtered or summarised. The complete set is the substrate
recurrence runs on: a finding filtered out once can never be observed appearing
twice, and recurrence is the only signal separating a real defect from one bad
afternoon.

Ordinals are as at `5cc2a58` plus this retro's three additions. The
`retro:filed:` back-links written after this document are bookkeeping, not
findings, and are excluded.

### `tooling` — the skill audience (8)

#### `agent-plugins-mpnf#1` · target `coordinator` · **incident**

```
retro:tooling:coordinator incident — the coordinator suppressed an exit status and lost an error message, in the run whose queue contains agent-plugins-x5tk about exactly this. Wiring the sequencing edges, 'bd dep add agent-plugins-0913 agent-plugins-qozp.8' was run as '>/dev/null 2>&1' and treated as done. It had exited 1 with: 'dependency agent-plugins-0913 -> agent-plugins-qozp.8 already exists with type "related" (requested "blocks"); remove it first with bd dep remove then re-add'. The related edges came from the coordinator's own earlier filing of 0913. Caught only because the ready front was re-queried afterwards and 0913 appeared in it when it should have been blocked.

Two rules broken, both already written down. AGENTS.md: 'never read $? through a pipe... capture, then test' - the same defect with redirection instead of a pipe, which the section does not name. And bd's own refusal to UPGRADE an edge type is undocumented in this corpus: 'related' does not silently become 'blocks', it errors.

Consequence had it gone unnoticed: agent-plugins-0913 edits the same files as items 8 and 9 and would have been dispatched concurrently with them. Repaired by dep remove then dep add, reading every status; front now correctly holds 0913 back. Candidate: extend AGENTS.md's pipes section to cover discarded status generally, and record the edge-upgrade refusal.
```

#### `agent-plugins-mpnf#9` · target `sprint-retro`

```
retro:tooling:sprint-retro a bead filed by this skill carries no Evidence section, so its only evidence pointer is a document path — and 25 beads now cite a findings document that no longer exists.
evidence: the skill's step 6 files a project-audience issue as
  bd create --title "<title>" --priority <n> --description "<the issue body>. Findings: <ids>. Report: <path>"
The issue template's "Evidence" section (one bullet per finding, quoting its headline) has no place in that form; `Findings: <ids>` is a list of ids that resolve only through the findings document, and `Report: <path>` is that document. Both collapse to one dependency: the file.
Measured on this tracker at 5cc2a58, after commit 9cc8844 deleted the four 2026-08-24 sprint documents:
  $ bd export > e.jsonl
  $ jq -r 'select((.description//"")+(.notes//"")+(.design//"") | test("rfcs/sprints/2026-08-24")) | .id' e.jsonl | wc -l
  25
  $ jq -r 'select(... same ...) | select(.status!="closed") | .id' e.jsonl | wc -l
  21
  $ ls rfcs/sprints/ | grep 2026-08-24   -> no output
Four distinct dead paths are cited: -beads-coordination-rollout-findings.md, -beads-coordination-rollout-report.md, -run-2-coordination-corrections-findings.md, and one truncated 'rfcs/sprints/2026-08-24-'.
Worst instance: agent-plugins-r9yj (open, P3) says "the three entries are drafted verbatim there" — the verbatim text was the whole of the bead's content and it is gone, leaving an open bead whose work cannot be performed from the bead.
context: filing this run's candidates and having to decide what a filed bead must carry to survive its findings document.
```

#### `agent-plugins-mpnf#10` · target `sprint-retro`

```
retro:tooling:sprint-retro the skill hard-codes two destinations this project has retired, and offers no way to say where they went.
evidence: SKILL.md step 4(a) writes
  path="rfcs/sprints/<YYYY-MM-DD>-<epic-slug>-findings.md"
and step 6's table sends the skill audience to
  .sf/outbox/<skill>/<YYYY-MM-DD>-<slug>.md
Commit 9cc8844 ("Drain the outbox and the run-3 sprint docs into the tracker") removed 15 .sf/outbox issue documents and the four 2026-08-24 sprint documents at the owner's instruction; `find .sf -type f` now returns only .sf/outbox/README.md and `ls rfcs/sprints/ | grep 2026-08-24` returns nothing. Both destinations are still literal strings in the skill, and neither sf.toml nor the run manifest carries a key that could redirect them: `jq -r 'select(.id=="agent-plugins-mpnf") | .metadata | keys' ` lists moved, branch, ledger, admitted, base_sha, excluded, checks_ids, moved_note, scope_notes, excluded_note, review_passes, standing_auth, checks_baseline, concurrency_cap, checks_baseline_status — no findings/outbox destination among them.
Consequence for this run: the dispatch had to override the skill in prose, twice. An agent running it without that override would have written into a channel the owner closed.
context: this retro was told, in its dispatch, not to use either path.
```

#### `agent-plugins-mpnf#11` · target `findings-capture-protocol`

```
retro:tooling:findings-capture-protocol `retro:incident:` was written as a KIND, so the run's most consequential finding routes nowhere — and the unrouted count has now been 2 for three consecutive runs.
evidence: agent-plugins-mpnf#6 is headed
  retro:incident:coordinator-removed-worktree-before-the-merge-succeeded
`incident` is not one of the protocol's nine kinds. The protocol says so — "Incidents are a severity, not a kind" — and shows the correct form, `retro:gap:ADR-014 incident — ...`, but nowhere states the wrong form, so an author who reads "incident" as a first-class concept has nothing telling them it may not occupy the kind slot.
The cost is not cosmetic: mpnf#6 is the run's only record of a state the corpus asserts cannot occur (a removed worktree coexisting with an unmerged branch), it names two remedies against bead-bonanza, and the routing table sends it to UNROUTED — "a count and a list", no audience, no candidate.
Unrouted per run, from rfcs/retro-log.md: run agent-plugins-qozp "2 findings unrouted"; run agent-plugins-so39 "2 findings unrouted (both retro:process:, both the coordinator's)"; this run 2 — mpnf#5 (retro:process) and mpnf#6 (retro:incident). agent-plugins-y63n.17 already records the first two runs and says the count is not falling; this is the third, and the new invention is drawn from the protocol's own escalation vocabulary rather than made up.
```

#### `agent-plugins-mpnf.4#5` · target `scripts/check-skills.sh`

```
retro:tooling:scripts/check-skills.sh rule 2 rejects a CONCRETE example path inside backticks, so a worked example cannot name a real artifact filename. Had to write the example's two files as placeholders instead.
evidence: with the example written as `.beads/plans/r7k.graph.json` —
  ERROR plugin/skills/sprint-plan/references/run-record.md: backticked path does not resolve from this file: .beads/plans/r7k.graph.json
  ERROR .claude/skills/sprint-plan/references/run-record.md: backticked path does not resolve from this file: .beads/plans/r7k.graph.json
  --- 2 error(s), 0 warning(s) ---  RESULT: FAIL
The rule's regex is `((?:\.\./)*[A-Za-z0-9_./-]+\.(?:md|json|sh|py))` (check-skills.sh:148), which does not match `<`, so the placeholder form `.beads/plans/<run-id>.graph.json` is invisible to it and passes. Renaming to the placeholder fixed it. Same file is reported twice because .claude/skills/ symlinks into plugin/skills/.
```

#### `agent-plugins-mpnf.12#2` · no target

```
retro:tooling the chmod-parent-directory self-test trap named in the bead's Design is real and was avoided by construction: the added self-test cases make the counter PATH ITSELF a directory (mkdir -p "$BB_STOP_HOOK_DIR/blocks" / "$BB_SUBAGENT_HOOK_DIR/blocks-<slug>"), never chmod a parent. No filler beyond that -- the bead's fix direction, self-test shape and message text were all followed as specified and needed no deviation.
```

#### `agent-plugins-mpnf.17#4` · no target

```
retro:tooling my change shifted the line number that agent-plugins-mpnf.18's finding cites, and I did NOT touch the claim itself.
evidence: the stale 'bd ready --json returns null (not []) when the front is empty' comment sat at bb-status.sh:459 at my base commit 7120474 and sits at bb-status.sh:560 after this bead. Measured:
  $ git show HEAD:plugin/skills/bead-bonanza/scripts/bb-status.sh | grep -n 'returns null (not'  -> 459
  $ grep -n 'returns null (not' plugin/skills/bead-bonanza/scripts/bb-status.sh                  -> 560
The comment is unchanged in text and is in the CREEP block of mode_run, which this bead did not rewire — it is neither more nor less wrong than it was. A bead dispatched with a line number rather than the string will miss it after any edit to the same file; the string is the durable handle.
```

#### `agent-plugins-mpnf.20#3` · target `plugin/skills/bead-bonanza/scripts/sprint-cleanup.sh`

```
retro:tooling:plugin/skills/bead-bonanza/scripts/sprint-cleanup.sh a self-test stub cannot report back through a global when the code under test calls it inside a command substitution — the assignment happens in a subshell and is discarded. Cost one failing assertion that looked like a bug in the implementation.
evidence: first attempt asserted --all with a flag the stub set:
  _sc_saw_all=no
  bd() { case " $* " in *" --all "*) _sc_saw_all=yes ;; esac; ... }
The call site is `_ed_json="$(bd list --parent "$_ed_cur" --all --json 2>/dev/null)"`, so the stub ran in a subshell. Result on a walk that DID pass --all:
  FAIL scope, the walk keeps --all so closed beads are still matched (the walk queried without --all)
Rewritten to assert through the stub's ANSWER instead — without --all the stub returns a bead id acme-MISSING-ALL that must not appear in the result — which crosses the subshell because it is stdout. Verified by mutation: dropping --all from the real query now produces `got [acme-MISSING-ALL ]`.
This is the same subshell hazard the file already documents twice (the note on prune_worktrees's while-loop, and the one telling callers not to capture handle_worktree's stdout); neither mentions the stub-observation form, which is a third face of it.
```

### `gap` — the project audience (3)

#### `agent-plugins-mpnf#4` · target `coordinator` · **incident**

```
retro:gap:coordinator — the dispatch prompt for agent-plugins-qozp.6 was SELF-INCONSISTENT: its Scope list named plugin/skills/sprint-plan/SKILL.md, bead-bonanza/SKILL.md and bead-bonanza/references/run-record.md as owned, while its byte-budget paragraph named sprint-plan/references/run-record.md as the place detail should go. The agent needed that fourth file - check-skills.sh rule 3 makes it the only file sprint-plan/SKILL.md can legally point at for its own schema - and edited it, recording an incident rather than either silently widening or blocking. No sibling in the wave owned it, so no collision resulted.

The defect is the coordinator's: two parts of one prompt described different file sets and nothing reconciled them before sending. Same class as the sync_dir misattribution in run 2 - a dispatch prompt is an authoring surface with no reviewer, which is exactly what agent-plugins-x5tk (merged earlier this wave) now documents. The agent's handling was correct and is the reason this cost nothing.
```

#### `agent-plugins-mpnf.4#3` · target `plugin/skills/sprint-plan/references/run-record.md` · **incident**

```
retro:gap:plugin/skills/sprint-plan/references/run-record.md incident — the bead's Design specifies TWO artifacts but says nothing about how the second path is RECORDED, and the manifest is a cross-skill interface. Decided: add `edges_sidecar` beside `graph_plan` in the documented manifest shape, same ownership rule (the plan returns the path; the orchestrator writes the key). Alternative considered: leave the sidecar path unrecorded and let the orchestrator re-derive it from graph_plan by extension substitution — rejected because a recovering session would have to infer a filename convention rather than read one, and run-record.md's whole premise is that recovery reads the epic instead of re-deriving the run.
evidence: run-record.md manifest block previously carried `graph_plan` alone; sprint-plan/SKILL.md's 'Return the plan' now returns two paths, so a single key cannot carry them.
```

#### `agent-plugins-mpnf.18#5` · target `retraction-sweep-reach`

```
retro:gap:retraction-sweep-reach the sweep of surfaces holding copies of the belief, in AGENTS.md's order of reach. One copy is unreachable and is named here rather than left silent.
evidence:
  DURABLE TRIGGERS — clean. list_triggers returned two: trig_01BEsKhRBN7TdijeQrazj3jY (PR #24 self check-in, enabled, next 2026-08-25T08:10Z) and trig_01VhJeqmHBjp5FinDTYkkSHi (PR #20 send_later, ended auto_disabled_session_gone). Read both prompt texts in full: neither mentions bd ready, JSON emptiness, or null.
  BEADS — clean. bd search 'null' --status all --json returned agent-plugins-tix (closed, strata null-clearing, unrelated) and this bead. bd search 'bd ready' --status all --json returned agent-plugins-offx.6, -mpnf.22, -so39.8, -qozp.13; bd show on each, grepped for the phrasings, returned nothing.
  .sf/outbox — clean, and this is a CHANGE from what the epic's retraction comment recorded. It named 2026-08-24-bb-status-human-queue-check-is-repo-wide.md as carrying a bd human list claim. That file is gone: find .sf -type f now returns only .sf/outbox/README.md, in both the main checkout and this worktree, and grep -rn null .sf/ returns nothing.
  UNREACHABLE — the run epic's own comment stream carries a pasted self-test transcript containing the OLD assertion name, 'ok   empty queue (null, not []) does not block'. It is a record of what a check printed at the time, not a live claim, and a run record is not edited after the fact. A future reader grepping the epic for the phrase will hit it. Flagged, not fixed.
  The bead description and the coordinator's 2026-08-25 00:16 epic comment also quote the false sentence, but as retraction records, which is what they are for.
```

### `decision` — the project audience (9)

#### `agent-plugins-mpnf#8` · target `review-pass-2-admission`

```
retro:decision:review-pass-2-admission

Pass 2 filed twelve findings, all classified in-scope by the reviewer, none pre-admitted. I admitted ELEVEN and deferred two.

Admitted: mpnf.15, .16, .17 (P1), .18, .19, .20, .21, .22, .23, .24 (P2), and agent-plugins-p567 (P1, filed by dev-mpnf-8 during batch 8 as discovered-from mpnf.4, reparented to this epic).

Deferred:
- mpnf.25 (P3) — `agent_id_from` lacks a jq fallback its sibling has. Cosmetic robustness, no observed failure.
- mpnf.26 (P2) — three skill bodies within 150 bytes of the 20,000 cap. The owner explicitly parked the byte-cap category: "Ignore byte caps for now — we'll address them towards the end of the refactor." Deferring it is honouring that, not overlooking it. It remains a live constraint ON this loop, carried into the planning prompt.

`admit_review_findings` is STILL `false` by silent default. This is the second consecutive loop in which I have overridden it, on the owner-approved principle that a defect the run's own merged diff introduced is in scope regardless of that default. The row has now never been chosen by anybody in three runs, and has been overridden twice in this one. That is a defaulting mechanism producing the wrong answer often enough that the report should say so plainly rather than record two one-off overrides.

Why p567 was admitted rather than left as a follow-up: it is the same defect class as mpnf.15 — the next run's plan is non-functional — and it was scoped OUT of mpnf.4 by mpnf.4's own Design, which is correct bead hygiene but leaves the mechanism half-shipped. Shipping the planner side of the sidecar without the orchestrator side means an orchestrator following the text creates beads, wires no edges, and reports success.

Budget: review_passes is 2 of 3. This is the LAST loop. Whatever is not fixed here is deferred and reported, and there is no pass 4 to catch a regression introduced by these eleven remedies — which is itself a reason to prefer conservative, well-tested fixes over ambitious ones.
```

#### `agent-plugins-mpnf.3#3` · target `plugin/skills/bead-bonanza/references/worktree-mechanics.md`

```
retro:decision:plugin/skills/bead-bonanza/references/worktree-mechanics.md cited the planner BY ROLE rather than by path, and deliberately did not restate the sidecar mechanism there. Wrote: 'The sidecar's format, its all-or-nothing loading and the order the two files load in belong to the planner; its run-record reference is where they live, and none of it is restated here.'
Alternative considered: vendor the sidecar section into bead-bonanza/references/shared/ so the citation resolves. Rejected on AGENTS.md's two tests — the citing skill needs one fact out of a larger document, not the whole file, and sync-shared-references.sh vendors whole files only (its render is `cat` over the source), so there is no vendoring form for this at all.
evidence: scripts/check-skills.sh rule 3 errors on any `../<skill>/references/` path and rule 2 requires a backticked *.md path to resolve from the citing file's own directory, so a path from bead-bonanza to sprint-plan's reference cannot be written. Checks pass with the by-role wording: skills / skills-strict / ref-links / shared-refs all PASS, exit 0.
```

#### `agent-plugins-mpnf.4#6` · target `plugin/skills/sprint-plan/SKILL.md`

```
retro:decision:plugin/skills/sprint-plan/SKILL.md byte budget spent to 5 bytes free, not blocked. Measured with the checker's own method before and after.
evidence:
before (22c86ab): python3 -c "t=open('plugin/skills/sprint-plan/SKILL.md').read(); b=t.split('---',2)[2].encode(); print(len(b), 20000-len(b))" -> 19948 52
after:  same command -> 19995 5
Three trims paid for the additions, all inside the mechanism this bead rewrites: (a) 'there is nothing to store and no schedule to keep in step' -> 'there is no schedule to keep in step'; (b) 'have different consequences and must not read alike; the marking that keeps them distinguishable' -> 'have different consequences; the marking that distinguishes them'; (c) the 'Out' item-2 sentence lost the enumeration of what the plan contains, which the reference carries. No unrelated rule was cut. The frontmatter description was also corrected (it claimed collision edges ride in the graph plan) — frontmatter is outside split('---',2)[2] so it costs no body bytes.
Decided silently: 5 bytes of headroom is accepted rather than cutting further. The next edit to this file will need its own trim, which is the pre-existing condition of the file, not something this bead introduced.
```

#### `agent-plugins-mpnf.4#7` · target `agent-plugins-mpnf.4` · **incident**

```
retro:decision:agent-plugins-mpnf.4 incident — AC7 ('no remaining instruction anywhere in sprint-plan/ tells a planner to put a pre-existing bead id in a graph-plan edge') is in tension with this bead's own Design SCOPE ('The merge bead's fan-in ... belongs to the sibling bead that waits on this one; do not fix it here'). Decided: honour the Design. The one remaining instance is the merge bead's fan-in, which agent-plugins-mpnf.3 owns and which lands next in the same batch behind a real blocks edge, so AC7 holds at batch end and not at this commit.
Alternative considered: fix the merge-bead block here to satisfy AC7 literally. Rejected — it would leave mpnf.3 with no work, and the Design forbids restating the mechanism in two places, which is exactly what a second author of the same block produces.
evidence: after this bead's edit, `grep -rn 'from_key|to_key' plugin/skills/sprint-plan/` leaves exactly one edge naming a real bead —
  run-record.md:232: {"from_key": "m1", "to_key": "<each bead in the batch>", "type": "blocks"}
run-record.md:330 also shows a graph-plan edge naming a tracker id, but that is the measured FAILURE transcript, quoted deliberately as what not to write. Every other from_key/to_key in the directory is a plan-local key.
```

#### `agent-plugins-mpnf.16#3` · target `plugin/skills/bead-bonanza/scripts/sprint-cleanup.sh`

```
retro:decision:plugin/skills/bead-bonanza/scripts/sprint-cleanup.sh handle_worktree's three-way dispatch is a `case` on a captured return, not a chain of `if agent_drained`.
evidence: `elif agent_drained "$_hw_who"; then` cannot see a return of 2 — `if` collapses every non-zero status to 'false', so an UNKNOWN verdict would have fallen into the existing 'still has a queue' branch. That branch happens to be non-destructive, so the bug would have been invisible, but the LEFT line would have named the wrong reason and a reader chasing a stale worktree would have looked for a live agent that does not exist. Captured as `agent_drained "$_hw_who"; _hw_ad=$?` then `case $_hw_ad in 0|2|*)`, giving unknown its own line: 'worktree, the tracker could not answer for <agent>'.
Alternative considered: keeping the if/elif chain and having agent_drained set a global on unknown. Rejected — a side-channel global is the shape that produced the `_n` collision this file's own comment already warns about.
```

#### `agent-plugins-mpnf.17#2` · target `agent-plugins-mpnf.17` · **incident**

```
retro:decision:agent-plugins-mpnf.17 incident — acceptance criterion 3 says a FLAT epic must report 'exactly what it reported before this change'. It cannot, literally: the third column of the two rewired lines is the command that established the reading, and the old text named 'bd list --parent $_epic --status in_progress' — a command that returns the WRONG answer on a nested epic and would hand a reader a re-derivation that disagrees with the line beside it. I changed those to 'bb-status.sh slot $_epic' / 'bb-status.sh run $_epic', which is the file's existing convention for a walked fact (the STATUS lines already cite 'bb-status.sh run $_epic'). AC 3 is therefore satisfied in substance and not in letter: on a flat epic the STATE, the VERDICT and the EXIT CODE are byte-identical; one column of citation text differs.
evidence: flat epic zt-1ux.1, before vs after, full diff of both modes —
  run:  9c9
        <   OK   beads in progress: zt-1ux.1.1(agent-x) bd list --parent zt-1ux.1 --status in_progress
        >   OK   beads in progress: zt-1ux.1.1(agent-x) bb-status.sh run zt-1ux.1
        exit 1 before, exit 1 after
  slot: 5,7c5,7 — only the two citation columns plus 'last commit: 78 seconds ago' -> '4 minutes ago' (wall clock, not the change)
        FAIL do NOT release - a bead is claimed and in progress: zt-1ux.1.1(agent-x)   (identical verdict text)
        exit 2 before, exit 2 after
The file's own header states the contract this defends: 'Every line prints the verified state AND the command that established it, so a reader can re-run any single check without reconstructing it.' Keeping a citation that re-derives to a different answer would have broken that instead.
Alternative considered: keep the old third column verbatim to satisfy AC 3 to the letter. Rejected — it costs nothing to type and leaves a documented lie in the output that the next reader re-runs and is misled by, which is the same class of failure this bead exists to remove.
```

#### `agent-plugins-mpnf.18#3` · target `replacement-justification`

```
retro:decision:replacement-justification the replacement sentences do not just swap null for []. Swapping alone would have left the defensive code justified by a measurement that no longer supports it — every site's guard is wider than 'accept [] or null', and a reader who saw only '[] is what bd emits' would have grounds to delete the guard. So each replacement is re-grounded on a hazard I measured instead.
evidence: jq is on PATH in this container; measured payload -> jq -r 'length':
    null                | 0   exit 0
    []                  | 0   exit 0
    "an error message"  | 16  exit 0
    {"a":1,"b":2}       | 2   exit 0
    [{"id":"x"}]        | 1   exit 0
  So null was never the hazard a bare jq length faced — it already yields 0. The real one is a STRING or OBJECT payload reporting a nonzero count, which is what the two hooks' 'if type=="array"' guards actually earn their keep against. Both hook comments now say that.
  Alternative considered and rejected: deleting the justifying sentences outright. The bead's Rejected section already names why — each one is justifying real defensive code, and removing the justification invites the code's removal.
  ALSO CORRECTED MID-WORK, silently, and worth naming because I got it wrong first: my initial sprint-cleanup.sh replacement justified json_empty's breadth by claiming the destructive reading was 'not empty'. That is inverted. agent_drained returns 0 (drained) when json_empty holds for both queries, and handle_worktree force-removes on drained — so EMPTY is the destructive verdict, not 'not empty'. What actually separates 'nothing there' from 'could not tell' is the query's exit status, read at sprint-cleanup.sh's agent_drained before json_empty is reached (_ad_rrc / _ad_prc; non-zero returns 2 = UNKNOWN). The committed comment says that instead, and agrees with the THREE-outcomes paragraph directly below it rather than contradicting it.
```

#### `agent-plugins-mpnf.24#4` · no target

```
retro:decision AC 'grep -n CLAUDE_PROJECT_DIR plugin/hooks/*.sh shows it only in a documented fallback path' is not literally true after the fix: each hook also has a pre-existing, unrelated line in mode_hook -- `cd "${CLAUDE_PROJECT_DIR:-.}" 2>/dev/null || true` -- used to point `bd` at the project root before querying the tracker, not part of state_dir. Left it untouched: the Design's Constraints scope this bead to 'where the state lives and what --arm/--status print' and explicitly forbid changing block/counter or unrelated hook_mode logic. Evidence: `grep -n CLAUDE_PROJECT_DIR plugin/hooks/*.sh` after the fix shows 2 hits per file inside state_dir's fallback (the documented one) plus 1 hit per file at that mode_hook cd line.
```

#### `agent-plugins-mpnf.24#5` · no target

```
retro:decision Design's Part B ('--arm and --status both print the resolved state_dir path' and 'status say plainly when no state directory exists') was already true of the pre-existing code in both hooks -- mode_arm already printed 'state: <dir>' and mode_status already printed either 'inert: no run armed. (state would live at <dir>)' or 'armed: <epic>' + 'state: <dir>'. No code change was needed for Part B; verified by reading the unmodified mode_arm/mode_status bodies and by running --status/--arm output above. Implemented only Part A (the state_dir derivation).
```

### `stale` — the project audience (2)

#### `agent-plugins-mpnf.3#2` · target `dispatch-prompt`

```
retro:stale:dispatch-prompt the dispatch prompt's handed-down fact '`bd ready --json` returns `null` when empty, but `[]` for an `--assignee`-filtered empty query. An emptiness test must accept both.' did NOT reproduce. On bd 1.2.2 (dev) every empty result I could construct returned `[]`, never `null`.
evidence — four configurations:
1. scratch DB with issues present, none ready:      $ bd ready --json            -> []
2. scratch DB with issues present, none ready:      $ bd ready --assignee merge-mb --json -> []
3. freshly initialised EMPTY scratch DB (prefix emp): $ bd ready --json           -> []
4. this repo's live tracker, nonexistent assignee:  $ bd ready --assignee no-such-agent-xyz --json -> []
Control that the command works at all: this repo's live `bd ready --json` returns a populated array beginning {"id":"agent-plugins-merge-slot"...}.
NOT worked around and NOT relied on: no document this batch touched asserts either shape. Flagging it because the claim is the kind that gets copied into an emptiness test, where accepting only `null` or only `[]` fails silently. What would settle the difference: the bd version and query the original observation was made on — it may predate 1.2.2 (dev) or use a flag combination I did not try.
```

#### `agent-plugins-mpnf.18#4` · target `bead-description`

```
retro:stale:bead-description the bead's site inventory was stale on both count and position by the time it was dispatched, exactly as groom-beads' unquoted-file:line anti-pattern predicts. Matching by string rather than by line number was what made it survivable.
evidence: the description says 'five NEW copies' plus a sixth pre-dating the diff, and cites 'plugin/skills/bead-bonanza/scripts/bb-status.sh:459' and 'bb-stop-ready-loop.sh:213'.
  Measured at c438c05, the commit this worktree is on:
    $ grep -rnE 'returns null|returns `null`|answers null|answers `null`|null, not \[\]|null rather than \[\]|`null` rather than|not `\[\]`' . --exclude-dir=.git
    ./plugin/skills/bead-bonanza/references/dispatching.md:258
    ./plugin/skills/bead-bonanza/scripts/sprint-cleanup.sh:173
    ./plugin/skills/bead-bonanza/scripts/sprint-cleanup.sh:315
    ./plugin/skills/bead-bonanza/scripts/bb-status.sh:560
    ./plugin/hooks/bb-stop-ready-loop.sh:84
    ./plugin/hooks/bb-stop-ready-loop.sh:228
    ./plugin/hooks/bb-subagent-stop-queue.sh:135
  SEVEN sites in five files, not six. bb-status.sh moved 459 -> 560 (sibling agent-plugins-mpnf.17's edit, as its agent reported) and the self-test name moved 213 -> 228.
  A second grep for other phrasings of the same belief — 'null when|null instead|instead of []|null payload|yields null|null on empty|emits null', case-insensitive, tree-wide — returned nothing (exit 1), so the seven were the whole set.
  After the sweep the same grep returns nothing (exit 1).
```

### `unverifiable` — the project audience (3)

#### `agent-plugins-mpnf.16#4` · no target

```
retro:unverifiable the acceptance criterion 'sh -n and shellcheck (as scripts/ci.sh runs them) report nothing new' cannot be satisfied as written: ci.sh runs NEITHER, and shellcheck is not installed in this container.
evidence:
  $ grep -n 'shellcheck' scripts/ci.sh sf.toml    -> no output
  $ grep -n 'sh -n\|bash -n' scripts/ci.sh sf.toml plugin/skills/bead-bonanza/scripts/bb-checks.py  -> no output
  $ command -v shellcheck  -> not found
Substituted what is available and ran it by hand: `sh -n plugin/skills/bead-bonanza/scripts/sprint-cleanup.sh` exits 0. shellcheck is UNVERIFIED for this change; installing shellcheck, or adding it to the declared check set, would settle it.
Same criterion appears verbatim on agent-plugins-mpnf.20.
```

#### `agent-plugins-mpnf.20#5` · no target

```
retro:unverifiable this bead's 'sh -n / shellcheck report nothing new beyond the recorded baseline' has the same defect as the identical criterion on agent-plugins-mpnf.16: scripts/ci.sh runs neither, and shellcheck is not installed.
evidence:
  $ grep -n 'shellcheck' scripts/ci.sh sf.toml    -> no output
  $ grep -n 'sh -n\|bash -n' scripts/ci.sh sf.toml plugin/skills/bead-bonanza/scripts/bb-checks.py  -> no output
  $ command -v shellcheck  -> not found
`sh -n plugin/skills/bead-bonanza/scripts/sprint-cleanup.sh` run by hand exits 0. shellcheck remains UNVERIFIED for this change.
```

#### `agent-plugins-mpnf.24#3` · no target

```
retro:unverifiable The bead's own remedy-sketch half concerning what CLAUDE_PROJECT_DIR carries inside a real harness-fired hook invocation remains UNVERIFIED -- I have no way to spawn a real harness hook from this session either. I did not attempt to settle it and made no new claim about it; the fix instead removes state_dir's dependency on that variable when a git repo is present, so its value no longer matters for the fixed behavior. What would settle the original claim is unchanged from the bead's own note: arm from a repo checked out under a multi-root parent and invoke the hook through the harness, comparing --status against what the hook reads.
```

### `sequencing` — the project audience (1)

#### `agent-plugins-mpnf.4#4` · target `plugin/skills/bead-bonanza`

```
retro:sequencing:plugin/skills/bead-bonanza the orchestrator's load step still knows about one file, so the mechanism this bead defines would never reach the tracker. Filed as agent-plugins-p567 (discovered-from mpnf.4), NOT fixed here — the bead's Design scopes this one to the mechanism in sprint-plan/.
evidence:
plugin/skills/bead-bonanza/SKILL.md:142 — '2. **The graph plan**, plus one line per override of its own preview. Diff it, re-run its validators, load it with `bd create --graph <path>`, then record the path as `graph_plan` on the manifest.'
plugin/skills/bead-bonanza/references/run-record.md:149 — '**`graph_plan` is yours to write and the planner'"'"'s to produce.** The plan returns the path of the `bd create --graph` JSON document that builds the run; you load it and record the path here'
Neither mentions a sidecar or `bd dep add --file`. An orchestrator following them loads the beads and wires no edges, reporting success.
```

### `scope` — the project audience (2)

#### `agent-plugins-mpnf.2#2` · no target

```
retro:scope The bead's own description flagged that design §4a (beads-coordination-design.md) shows the same --graph schema with no from_key/to_key orientation prose, but this bead's scope and acceptance criteria cover only docs/plans/beads-coordination-rollout.md. Filed agent-plugins-fvnp (discovered-from this bead, P3, non-blocking) rather than editing the design doc here. Checked the design doc's one concrete edge example (rev -> tgt) against the measured semantics: it is directionally consistent (review waits for the target to finish) but the doc states no rule, so a reader still cannot derive the correct direction from it alone.
```

#### `agent-plugins-mpnf.20#4` · target `agent-plugins-mxl3`

```
retro:scope:agent-plugins-mxl3 found while measuring, filed rather than fixed: sprint-cleanup run from inside a run worktree treats THAT worktree as main and skips it, because `git rev-parse --show-toplevel` returns the worktree's own path.
evidence: live dry run from /home/user/agent-plugins/.claude/worktrees/dev-mpnf-11 against agent-plugins-mpnf —
  LEFT     worktree, no bead owns its branch        /home/user/agent-plugins
  LEFT     worktree, no bead owns its branch        /home/user/agent-plugins/.claude/worktrees/dev-mpnf-14
The main checkout is reported LEFT and the worktree the command ran from does not appear at all. Adjacent, not blocking: the failure is a tree left standing, never a removal. Filed as agent-plugins-mxl3 with a discovered-from edge to this bead.
```

### `proof` — terminates here, beside what it supports (17)

#### `agent-plugins-mpnf#2` · target `sprint-scope`

```
retro:proof:sprint-scope — CORRECTION to the coordinator's own scope-note 4, found by plan-mpnf. Scope recorded that no admitted bead has an open blocker outside the epic and predicted 'bd swarm validate should raise no outside-dependency warning'. It raised SIX: qozp.13<-qozp.12, qozp.14<-qozp.2, qozp.14<-qozp.12, qozp.15<-qozp.12, qozp.7<-qozp.1, qozp.9<-qozp.2. All six point at CLOSED beads and are harmless in fact, so scope's underlying observation was right; the prediction about the tool's behaviour was wrong. bd swarm validate warns on an edge to an out-of-epic bead whether or not that bead is closed.

The consequence is not cosmetic: run-record.md's rule for trusting the preview is unconditional - 'if it warns about an outside dependency, do not use its partition' - so the prediction, had it been believed, would have licensed using a partition the rule forbids. plan-mpnf followed the rule instead and derived packing from the re-queried ready front, which is the right call. Recorded as another instance of a coordinator claim about tool behaviour that was reasoned rather than measured.
```

#### `agent-plugins-mpnf#3` · target `bd-graph` · **incident**

```
retro:proof:bd-graph incident — bd create --graph's EDGE ORIENTATION IS THE REVERSE OF THE NATURAL READING, and the corpus had it the natural way. Measured by dev-mpnf-3 and INDEPENDENTLY RE-MEASURED by the coordinator in a separate scratch database (prefix edg, bd 1.2.2), same negative control both times:

  edges: [{"from_key":"first","to_key":"second","type":"blocks"}]
  bd ready   -> edg-x3g  SECOND must run second
  bd blocked -> edg-13l  FIRST must run first

So from_key WAITS FOR to_key. Writing an edge the intuitive way inverts every dependency in the plan, at exit 0, with no warning.

Why this is an incident and not a note: items 8, 9 and 11 all consume the graph plan, and agent-plugins-mpnf.2 records that docs/plans/beads-coordination-rollout.md item 6's own example has the orientation the wrong way round. A wave-2 or wave-3 agent reading the plan document rather than the corrected reference would emit an inverted graph that validates, creates beads, and orders the entire run backwards. THE CORRECTED SEMANTICS ARE NOW IN plugin/skills/sprint-plan/references/run-record.md - dispatch prompts for items 8, 9 and 11 must carry this fact explicitly rather than relying on the agent to find it.

Two more silent traps measured in the same pass: unknown NODE fields are dropped with a warning at exit 0 (so blocked_by and id in a node are accepted and ignored, yielding no dependency at all), while a malformed VALUE aborts before creating anything; and parent_id holding a plan KEY is accepted with no warning and yields an orphan - parent_key is the in-plan reference.
```

#### `agent-plugins-mpnf.2#1` · no target

```
retro:proof Independently reproduced all three cited corrections in a fresh scratch database (bd init --prefix scr) before writing them into the plan doc, rather than transcribing the dispatch prompt's transcript verbatim.
evidence: edge orientation — nodes alpha/beta, edge {from_key:alpha,to_key:beta,type:blocks} -> bd create --graph -> Created 2 issues; bd ready listed beta (scr-h6j); bd blocked listed alpha (scr-d4c) as 'Blocked by 1 open dependencies: [scr-h6j]'. parent_id trap — node with parent_id set to a plan key (not a real id) -> Created 1 issues with no warning; bd show scr-odc printed no PARENT section; bd dep list scr-odc -> 'scr-odc has no dependencies'. Field acceptance — node with metadata+gate+acceptance+notes -> 'warning: graph plan node["n1"] has unknown field(s): [acceptance gate notes] (silently dropped ...)', exit 0, Created 1 issues; bd show scr-9j0 --json | jq .metadata -> {"tier": "advanced"}, confirming metadata landed while gate/acceptance/notes did not.
```

#### `agent-plugins-mpnf.3#1` · no target

```
retro:proof the corrected fan-in was measured end to end before either file was rewritten — merge bead created BY THE PLAN, fan-in wired by the sidecar. Fresh scratch DB, prefix mb, bd 1.2.2 (dev).
evidence:
$ bd create --graph plan.json
Created 1 issues
  m1 -> mb-3p1
$ cat merge.edges.jsonl
{"from":"mb-3p1","to":"mb-ff7","type":"blocks"}
{"from":"mb-3p1","to":"mb-1px","type":"blocks"}
$ bd dep add --file merge.edges.jsonl ; echo $?
✓ Added 2 dependencies
0
$ bd ready --assignee merge-mb --json     # both inputs open
[]
$ bd ready --assignee merge-mb --json     # after closing only mb-ff7
[]
$ bd ready --assignee merge-mb --json     # after closing both
[{"id":"mb-3p1","assignee":"merge-mb","labels":["merge"]}]
$ bd ready --label merge --json
["mb-3p1"]
So the CLAIM both files already made about the fan-in was correct and only the MECHANISM was wrong, exactly as the bead's Design predicted. The claim is preserved with this transcript behind it rather than the inherited one.
```

#### `agent-plugins-mpnf.3#4` · no target

```
retro:proof mpnf.4's AC7 ('no remaining instruction anywhere in sprint-plan/ tells a planner to put a pre-existing bead id in a graph-plan edge'), deferred at that bead's close, now holds at batch end.
evidence: $ grep -n 'from_key|to_key' plugin/skills/sprint-plan/references/run-record.md plugin/skills/bead-bonanza/references/worktree-mechanics.md
worktree-mechanics.md: no hits at all.
run-record.md: 10 hits, and none is an instruction to name a real id in a plan edge —
  241, 388  prose relating the sidecar's 'from' to the plan's 'from_key'
  357       the measured FAILURE transcript, quoted as what not to write
  470, 494-495, 507-508  the schema table and the orientation trap, both plan-local by construction
  570-571   the worked example's two epic-to-epic edges, plan keys only
The merge-bead block that carried the last one is gone; it is now sidecar lines.
```

#### `agent-plugins-mpnf.4#1` · no target

```
retro:proof re-measured every handed-down bd claim before restating it in the corpus, per AGENTS.md 'Tool claims are measured at the point of use'. Fresh scratch DB, prefix scr3, bd version 1.2.2 (dev).
evidence:
$ bd create --graph t1.json ; echo $?
Error: invalid graph plan: edge 0: to key "scr3-9c4" not found in plan
1
  -- bd list --status all afterwards returned only the 2 pre-existing beads; the m1 node was not created.
$ bd dep add --file bad.jsonl ; echo $?
Error: bulk dependency validation failed:
  line 2: resolving dependency ID scr3-nope: resolving issue ID scr3-nope: no issue found matching "scr3-nope"
1
$ bd dep list scr3-x88
scr3-x88 has no dependencies      # the VALID first line did not land either
$ bd dep add --file good.jsonl   -> '✓ Added 1 dependencies', exit 0; bd ready -> scr3-9c4; bd blocked -> 'scr3-x88: Blocked by 1 open dependencies: [scr3-9c4]'  (so 'from' is the WAITER)
$ bd dep add --file good.jsonl   (replay) -> '✓ Added 1 dependencies', exit 0, no duplicate error
$ bd update scr3-9c4 --set-metadata collision=c4_changes:auth-service -> {"collision":"c4_changes:auth-service"}
$ bd update scr3-x88 --set-metadata collision='"c4_changes:auth-service"' -> {"collision":"\"c4_changes:auth-service\""}   (quotes stored literally, exit 0)
Every fact in the dispatch prompt held. Nothing contradicted.
```

#### `agent-plugins-mpnf.4#2` · no target

```
retro:proof the worked example in run-record.md was BUILT, not composed. Fresh scratch DB, prefix r7k, bd 1.2.2 (dev): loaded the 4-epic graph plan, created 3 admitted beads under the target, then loaded the 2-line sidecar.
evidence:
$ bd create --graph r7k.graph.json
Created 4 issues
  preflight -> r7k-3gm
  report -> r7k-mdx
  run -> r7k-cco
  target -> r7k-m10
$ bd dep add --file r7k.edges.jsonl ; echo $?
✓ Added 2 dependencies
0
$ bd ready --json | jq -r '.[].id'   -> r7k-3gm, r7k-cco   (pre-flight and the run epic)
$ bd blocked                          -> r7k-m10.3, r7k-m10.2, r7k-m10.1, r7k-m10, r7k-mdx   (the other five)
$ bd show r7k-m10.2 --json | jq -c '.[0].metadata'  -> {"collision":"c4_changes:bead-bonanza"}
So the example's 'bd ready returned the run epic and pre-flight, and bd blocked returned the other five' is a measurement of THIS example, not carried over from the previous one. Note r7k-m10.1 is blocked with nothing in the sidecar naming it — its parent (the target) is what holds it.
```

#### `agent-plugins-mpnf.5#1` · no target

```
retro:proof Verified the pre-fix defect with a swap test: built a hybrid script combining the OLD handle_worktree (pre-fix, from HEAD) with the NEW self-test's blocked-bead case (assignee dev-done, status blocked, same assignee the drained stub reports as drained). Running --self-test against that hybrid: 'FAIL guard, a blocked bead's worktree is left despite a drained agent (expected left=1, got left=0 removal-attempted=yes)' - reproduces the defect for the stated reason (a blocked bead satisfies agent_drained, so the old code reached the remove-worktree arm). The same case passes against the fixed handle_worktree (status_for_branch consulted before agent_drained). Full checks (bb-checks.py, no args) exit 0, 26/26 self-test assertions pass in sprint-cleanup --self-test, template-drift shows the same 34 pre-existing info advisories as the baseline.
```

#### `agent-plugins-mpnf.12#1` · no target

```
retro:proof pre-fix failure reproduced exactly as the bead measured, then the fix and new self-test assertions verified with a negative control.
evidence: with the counter path made a directory (mkdir -p $d/state/blocks / $d/state/blocks-<agent>), before the fix: bb-stop-ready-loop.sh MAX=2 -> 10/10 invocations returned 2 (sequence '2 2 2 2 2 2 2 2 2 2'); bb-subagent-stop-queue.sh MAX=2 -> 8/8 returned 2 ('2 2 2 2 2 2 2 2'). After applying the fix (test the printf redirect's own exit status instead of discarding it with || true; warn and return 0 on failure): both sequences became all 0s.
Negative control on the new self-test assertions themselves: reverted only the fix hunk in each file (keeping the new 'unwritable counter never blocks, over repeated calls' assertion in mode_self_test), re-ran --self-test -> both failed with 'FAIL unwritable counter never blocks, over repeated calls (expected 000000, got 222222)', confirming the new assertion actually exercises the defect rather than passing vacuously. Restored the fix afterward; diff against the pre-revert file was empty (byte-identical restore).
```

#### `agent-plugins-mpnf.16#1` · no target

```
retro:proof the destructive behaviour was reproduced BEFORE the fix, driving the real agent_drained and the real handle_worktree with `bd` shadowed by a failing stub.
evidence: self-test on the unfixed script, with the new cases added and nothing else changed —
  FAIL drained verdict, both queries fail (expected 2, got 0)
  FAIL drained verdict, the in_progress query alone fails (expected 2, got 0)
  FAIL guard, an unreadable tracker leaves the worktree standing (expected left=1 removal-attempted=no, got left=0 removal-attempted=yes)
`removal-attempted=yes` is the destruction itself: handle_worktree reached its `run ... git worktree remove --force` branch for a live agent because the tracker could not answer. After the fix the same three cases pass and the whole self-test exits 0 with 31 ok / 0 FAIL.
```

#### `agent-plugins-mpnf.17#1` · no target

```
retro:proof pre-fix failure observed end to end in a scratch database, then again as a negative control against the new self-test.
evidence: scratch bd 1.2.2 database, shape zt-1ux (run epic) -> zt-1ux.1 (phase epic) -> zt-1ux.1.1 (claimed by agent-x, in_progress) + zt-1ux.1.2, plus direct child zt-1ux.2. Merge slot zt-merge-slot acquired by agent-x.

The divergence the bead rests on, re-measured rather than inherited:
  $ bd list  --parent zt-1ux --status in_progress --json   -> []
  $ bd list  --parent zt-1ux.1 --status in_progress --json -> [zt-1ux.1.1]
  $ bd list  --parent zt-1ux --json                        -> ['zt-1ux.2', 'zt-1ux.1']
  $ bd list  --parent zt-1ux --status open,in_progress,blocked,deferred,closed --json -> ['zt-1ux.2', 'zt-1ux.1']
  $ bd ready --parent zt-1ux --exclude-type=epic --json    -> ['zt-1ux.2', 'zt-1ux.1.2']
The grandchild appears in ready and is absent from list, under every status filter. Confirmed.

PRE-FIX, nested epic, slot mode (exit 0):
  $ bb-status.sh slot zt-1ux
    OK     beads in progress: none          bd list --parent zt-1ux --status in_progress
    OK   no worktree and no claimed bead backs agent-x
    OK   RELEASE IS DEFENSIBLE: bd merge-slot release
...while zt-1ux.1.1 was in_progress and assigned to the very agent holding the slot.

PRE-FIX, nested epic, run mode — the same run contradicts itself two lines apart:
    OK     in_progress (1): zt-1ux.1.1      bb-status.sh run zt-1ux   <- the walker, correct
    OK   beads in progress: none            bd list --parent zt-1ux   <- the ungwalked query

POST-FIX, same shape:
  slot zt-1ux -> exit 2, 'FAIL do NOT release - a bead is claimed and in progress: zt-1ux.1.1(agent-x)'
  run  zt-1ux -> 'OK beads in progress: zt-1ux.1.1(agent-x)'

NEGATIVE CONTROL for the new self-test: a copy of the FIXED script with the descent disabled (stack.append made unreachable, i.e. pre-fix direct-children-only semantics) run with --self-test:
  FAIL walk descends past a phase epic: expected e.1 e.1.1 e.1.2, got e.1
  FAIL nested epic: the claimed bead is seen: expected e.1.1(dev-a), got
  ok   flat epic: unchanged, and only in_progress
  ok   a claimed bead with no assignee reads unassigned
  ok   an epic with no children holds nobody
  RESULT: FAIL (exit 1)
Only the two nested assertions fail, so the test isolates the traversal depth rather than the helper as a whole. The fixed script: RESULT: PASS, exit 0.
```

#### `agent-plugins-mpnf.18#1` · target `bd-ready-json-empty-shape`

```
retro:proof:bd-ready-json-empty-shape re-measured the replacement claim before writing it, per the rule that a correction is itself a new claim.
evidence: bd version 1.2.2 (dev). Scratch database: mktemp -d; git init -q; bd init --prefix zt.
  EMPTY TRACKER (all exit 0):
    bd ready --json                                      -> []
    bd ready --assignee no-such-agent-xyz --json         -> []
    bd ready --parent zt-nope --json                     -> []
    bd ready --exclude-type=epic --json                  -> []
    bd ready --parent zt-nope --exclude-type=epic --json -> []
  TWO ISSUES PRESENT (zt-0zb task, zt-6m1 epic), filters matching none:
    bd ready --assignee no-such-agent-xyz --json         -> []
    bd ready --parent zt-6m1 --json                      -> []
    bd ready --parent zt-6m1 --exclude-type=epic --json  -> []
    POSITIVE CONTROL: bd ready --json                    -> a 2-element array
  BOTH ISSUES CLOSED (present, none ready):
    bd ready --json                                      -> []
    bd ready --exclude-type=epic --json                  -> []
    bd ready --assignee no-such-agent-xyz --json         -> []
  null was produced in none of the 11 configurations.
  The positive control matters: without it, [] everywhere is equally explained by a broken query.
  For comparison, bd list --parent zt-nope --json -> [] and bd list --status open --json -> [] (exit 0).
  NOT a proof over all inputs or all bd versions: every replacement sentence says 'measured on bd 1.2.2 in every configuration tried', never 'always'.
```

#### `agent-plugins-mpnf.18#6` · target `gates`

```
retro:proof:gates gate results on the final tree, both the declared set and the ci.sh superset one acceptance criterion names by hand.
evidence:
  $ plugin/skills/bead-bonanza/scripts/bb-checks.py    (no arguments — the gate, not a narrowed loop)
  exit 0; all 13 declared checks pass: format vet build test-go skills shared-refs strata-docs schemas corpus ref-links template-drift script-self-tests skills-strict.
  Compared against the dispatch baseline (all 13 passed, exit 0): no line in my output that is not in the baseline. template-drift's 34 'info' advisories are present and pre-existing.
  $ ./scripts/ci.sh
  exit 0; '=== RESULT: PASS'.
  The renamed assertion ran and passed under script-self-tests:
    ok   a non-list payload counts as an empty queue
  and its stub is unchanged (stub 'null'), so the non-list branch is still the thing exercised.
  Diff shape verified mechanically rather than by eye, since the ACs bound it: stripping comment lines from the *.sh hunks leaves exactly one changed pair —
    - is "empty queue (null, not []) does not block" "$(run)" 0
    + is "a non-list payload counts as an empty queue" "$(run)" 0
  No other non-comment line of shell changed. No added line matches agent-plugins-[0-9a-z], so the edited text cites no issue ids.
  NOT VERIFIED: shellcheck. It is not installed in this container and no declared check invokes it, so the shell edits carry no lint verdict — though every added line in the two hooks and the two scripts is a comment, apart from the one string above.
```

#### `agent-plugins-mpnf.20#1` · no target

```
retro:proof the destruction was reproduced before the fix and each new assertion was then mutation-tested, so none of them passes while asserting nothing.
evidence: on the UNFIXED script, with the new self-test cases added and nothing else changed, the negative control passed —
  ok   UNSCOPED (the bug), the same foreign worktree IS removed
That case feeds handle_worktree the repo-wide bead list prune_worktrees builds today; acme-foreign belongs to another run, its agent reads as drained, and removal-attempted=yes. The scoped cases failed alongside it (scope_beads: command not found).
After the fix, three mutations of the shipped code, each reverted immediately:
  1. drop --all from the walk ->
     FAIL scope, a nested epic's GRANDchildren are all found (expected [acme-c1 acme-d1 acme-g1 acme-tgt ], got [acme-MISSING-ALL ])
     FAIL scope, the walk keeps --all so closed beads are still matched (the walk queried without --all)
  2. stop pushing epic children onto the worklist ->
     FAIL scope, a nested epic's GRANDchildren are all found (expected [acme-c1 acme-d1 acme-g1 acme-tgt ], got [acme-d1 acme-tgt ])
  3. scope_beads filters nothing ([ .[]? | select(...) ] -> [ .[]? ]) ->
     FAIL scope, a bead outside the run is filtered out and a closed one is kept (expected acme-g1,acme-c1, got [acme-g1,acme-c1,acme-foreign])
     FAIL scope, an empty scope matches nothing rather than everything (expected 0, got [3])
     FAIL scoped, a CONCURRENT run's worktree is left, not force-removed (expected left=1 removal-attempted=no, got left=0 removal-attempted=yes)
Mutation 3's last line is the bug itself, caught by the new suite. Restored: exit 0, 40 ok, 0 FAIL.
```

#### `agent-plugins-mpnf.20#2` · no target

```
retro:proof the two scratch-database acceptance criteria were measured against a real bd 1.2.2 (dev) database, not against the stubs.
evidence: scratch DB built at $SCRATCH/scratchdb with prefix scr — run epic scr-2cy -> target epic scr-bzb -> tasks scr-2p1 (open) and scr-0p2 (CLOSED); a separate epic scr-qkr -> task scr-8um.
First, the premise the whole walk rests on:
  $ bd list --parent scr-2cy --all --json | jq -r '.[] | "\(.id) \(.issue_type)"'
  scr-bzb epic
  $ bd list --parent scr-bzb --all --json | jq -r '.[] | "\(.id) \(.issue_type)"'
  scr-0p2 task
  scr-2p1 task
DIRECT children only — one query on the run epic returns neither task, confirming the objection the old comment recorded.
Then the shipped functions, extracted verbatim from the script with awk and sourced:
  $ scope="$(epic_descendants scr-2cy)"; echo $?
  0
  scr-bzb
  scr-0p2
  scr-2p1
  # foreign bead scr-8um (child of scr-qkr):
  not in scope -- correct
  $ bd list --all --json | jq -r '[.[].id] | join(" ")'
  scr-0p2 scr-8um scr-qkr scr-2p1 scr-bzb scr-2cy
  $ bd list --all --json | scope_beads "$scope" | jq -r '[.[] | "\(.id):\(.status)"] | join(" ")'
  scr-0p2:closed scr-2p1:open scr-bzb:open
Both grandchildren are found through the nested epic, the CLOSED one included (--all), and the foreign bead is filtered out of the repo-wide list.
```

#### `agent-plugins-mpnf.24#1` · no target

```
retro:proof Re-measured the Design's git-common-dir claim myself before implementing (git 2.43.0, this repo): main checkout `git rev-parse --git-common-dir` -> `.git` (relative); inside a worktree -> `/home/user/agent-plugins/.git` (absolute). `cd "$(git rev-parse --git-common-dir)" && pwd` gives the SAME absolute answer from both. Confirms the Design's rejection of --show-toplevel and its choice of the cd+pwd form.
```

#### `agent-plugins-mpnf.24#2` · no target

```
retro:proof Post-fix, ran the patched state_dir (via --status) from the main checkout and from the worktree with BB_STOP_HOOK_DIR/CLAUDE_PROJECT_DIR both unset: both print the identical resolved directory `/tmp/bb-stop-ready-loop/.git-2115161750` (and the subagent hook's analogous `/tmp/bb-subagent-stop-queue/.git-2115161750`). Also confirmed CLAUDE_PROJECT_DIR set to a bogus multi-root-style parent (/home/user) no longer changes the resolved directory once inside the git repo -- state_dir stops depending on it. Fallback verified separately: non-repo cwd with git present resolves via CLAUDE_PROJECT_DIR/$PWD as before; non-repo cwd or git absent from PATH (tested with a PATH containing coreutils but no git binary) still returns a non-empty usable path via the unchanged fallback line.
```

### `environment` — terminates here, routed to the ledger (5)

#### `agent-plugins-mpnf#7` · target `retracted-bd-ready-json-returns-null-when-empty`

```
retro:environment:retracted-bd-ready-json-returns-null-when-empty

RETRACTED. The claim "`bd ready --json` returns `null` when empty, but `[]` for an `--assignee`-filtered empty query" does not reproduce.

Measured by the coordinator on bd 1.2.2 (dev), three configurations, all `[]`:

  bd ready --parent agent-plugins-mpnf --exclude-type=epic --json   -> []
  bd ready --assignee no-such-agent-xyz --json                       -> []
  bd ready --assignee merge-mpnf --json                              -> []

Independently measured first by dev-mpnf-8 across four configurations (empty DB; issues present but none ready; assignee-filtered; this repo's live tracker with a nonexistent assignee) — also `[]` in every case, never `null`. It flagged the contradiction rather than working around it, which is the behaviour the prompt asked for and the reason this was caught.

How it got here: the loop-2 grooming agent recorded it as a ledger entry, citing "the run's existing ledger records `null`" as the thing its own observation differed from. I relayed the whole entry into three dispatch prompts without measuring it. Nothing depended on it in any of the three beads, so no work is affected.

Sweep, per AGENTS.md's retraction procedure — grepped BEFORE any edit:

- tracked corpus (`*.md`, `*.sh`, `*.py`, `*.toml`): NOT PRESENT. The only hits were `sprint-cleanup.sh`'s `json_empty`, whose contract is "0 when null, [] or unreadable-as-a-list" — it accepts both and is unaffected either way.
- epic ledgers on agent-plugins-mpnf, -qozp, -so39: NOT PRESENT. The grooming agent's "the run's existing ledger records null" does not correspond to any ledger entry I can find; the nearest is a different and still-true claim, that `bd show --json` reports `parent_id: null` for a bead that has a parent.
- `.sf/outbox/`: NOT PRESENT for this command. There IS a claim that `bd human list --json` returns `null` not `[]` (2026-08-24-bb-status-human-queue-check-is-repo-wide.md, twice, where it is called load-bearing). DIFFERENT COMMAND — not retracted here, and not tested by me. Do not let this retraction be read as covering it.
- dispatch prompts for batches 8, 9 and 10: carried it, now dead with their agents. Unreachable, so named here rather than fixed.
- durable triggers: none armed carrying it.

The distinction the retracted claim drew — filtered vs unfiltered emptiness — appears to be nothing. An emptiness test should still accept both forms, because accepting both costs nothing and this is now the second belief about bd's empty-JSON shape that turned out to be wrong.
```

#### `agent-plugins-mpnf.16#2` · no target

```
retro:environment bd 1.2.2 (dev) distinguishes 'nothing to report' from 'cannot answer' by EXIT STATUS, which is what makes this guard implementable at all.
evidence: measured this session in the worktree —
  $ out="$(bd ready --assignee nobody-xyz --json 2>&1)"; echo "rc=$?"; printf '%s\n' "$out"
  rc=0
  []
  $ out="$(bd list --assignee nobody-xyz --status in_progress --json 2>&1)"; echo "rc=$?"; printf '%s\n' "$out"
  rc=0
  []
  $ cd <empty scratch dir> && out="$(bd ready --assignee nobody-xyz --json 2>&1)"; echo "rc=$?"
  rc=1
  Error: no beads database found
An unanswerable query exits 1 and writes its message to stderr, so the payload captured by the command substitution is empty — indistinguishable from an empty answer on content alone. The status is the only discriminator, and the pre-fix code discarded it.
```

#### `agent-plugins-mpnf.17#3` · no target

```
retro:environment shellcheck is not installed in this container, so half of acceptance criterion 6 is UNVERIFIED.
evidence: $ command -v shellcheck >/dev/null 2>&1; echo $?  -> 1. No shellcheck binary, and no declared check invokes one (sf.toml's script-self-tests runs each script's own --self-test; the 13 declared ids are format vet build test-go skills shared-refs strata-docs schemas corpus ref-links template-drift script-self-tests skills-strict — none is shellcheck).
What IS verified: $ bash -n plugin/skills/bead-bonanza/scripts/bb-status.sh -> exit 0, and the full declared set exits 0 with 34 template-drift info advisories, matching the run epic's recorded baseline exactly.
Settled by: installing shellcheck and running it against the file, or by the repo declaring a shellcheck check so the bar is measured rather than assumed.
```

#### `agent-plugins-mpnf.18#2` · target `bd-json-empty-shape-is-per-subcommand`

```
retro:environment:bd-json-empty-shape-is-per-subcommand bd's empty-JSON shape is NOT uniform across subcommands: bd human list --json really does return null while bd ready --json returns []. This is the most likely origin of the retracted claim, and it means neither shape can be generalised from one subcommand to another.
evidence: same scratch database, bd 1.2.2 (dev), no human-labelled beads present:
    $ bd human list --json
    null
    [exit 0]
    $ bd human list
    No human-needed beads found.
    [exit 0]
  Against the same database in the same session:
    $ bd ready --json
    []
    [exit 0]
  The run epic's own retraction comment (bb-coordinator-mpnf, 2026-08-25 00:16) fenced the bd human list claim off as 'DIFFERENT COMMAND — not retracted here, and not tested by me'. It is now tested: it is TRUE and must not be swept.
  Checked that nothing in the tree mis-states it: grep -rn 'human list' . --exclude-dir=.git returns 18 hits across specs/, plugin/ and docs/, none of which asserts a JSON shape for it. Nothing to fix, so no bead filed.
```

#### `agent-plugins-mpnf.24#6` · no target

```
retro:environment shellcheck is not installed in this container (command -v shellcheck found nothing), so its half of the AC's 'sh -n / shellcheck report nothing new' could not be exercised. sh -n and bash -n both ran clean on both hooks after the fix. shellcheck is not part of sf.toml's declared checks, so this doesn't affect the checks gate.
```

### UNROUTED — kind absent or not one of the nine (2)

#### `agent-plugins-mpnf#5` · written as `retro:process` · **incident**

```
retro:process:coordinator incident — SECOND STALL of this session, and a recurrence of an invariant filed and approved during it. After admitting four review findings and re-entering the loop, the coordinator ended the turn having described the next dispatch rather than performing it. The queue was not dry, no bead carried human, the circuit breaker had not tripped — none of the three conditions that permit ending a turn. The human had to say 'go ahead'.

This is the exact invariant from run 2's retro candidate C11, approved by the owner and shipped this run as .sf/outbox/bead-bonanza/2026-08-24-naming-the-next-ceremony-is-not-dispatching-it.md: 'a coordinator turn ends with work in flight, a question to the human, or the run's terminal state — naming the next ceremony is not dispatching it.' RECURRENCE 1, in the run that filed it.

The first stall (before sprint-retro) had no particular cause. This one does, and it is worth recording because a rule alone will not fix it: the pending act was OVERRIDING THE AUTHORIZATION GATE, reversing what runs 1 and 2 did. Stopping let the decision be vetoed without the coordinator owning it. That is not a scheduling slip; it is hedging wearing the shape of a pause, and the operating principle names it — 'a coordinator that escalates has converted an autonomous run into a supervised one without saying so'.

The distinction the outbox issue should carry: if a coordinator believes a decision genuinely needs the human, the compliant move is to ASK, explicitly, which the principle permits as one of the three turn-ending conditions. Announcing a decision and then falling silent is neither deciding nor asking, and it reads to the human as the run having finished.
```

#### `agent-plugins-mpnf#6` · written as `retro:incident` · **incident**

```
retro:incident:coordinator-removed-worktree-before-the-merge-succeeded

Review loop 2, batch 10. I ran `git merge --ff-only bb/dev-mpnf-10`, `git worktree remove`, and the acceptance check as one chained command. The merge exited 128 ("Not possible to fast-forward" — the concurrent-wave case, because batch 9 had already moved the head from 22c86ab to e2b69ad) but the removal ran anyway and succeeded.

Nothing was lost: the branch still carried the commit and the remedy is a rebase, not a recovery. But the state it produced is the one the corpus says cannot happen. `worktree-mechanics.md` makes worktree presence the in-flight signal — "you remove it as the final act of accepting, after the last of its branches merges" — and for ~90 seconds a removed worktree coexisted with an unmerged branch. `bb-status.sh bead` reported it faithfully and exited 2:

  OK   3 -> STOP 3 NOT merged into HEAD
  OK   4 worktree: gone

Signal 4 read "accepted" while signal 3 read "not merged". The check is not wrong — it reports what it observes — but the pair is only meaningful because the coordinator is supposed to establish 3 before 4, and nothing enforces that ordering.

Two candidates for the retro, both against bead-bonanza rather than against this run:

1. The reference tells the coordinator the correct order and relies on it being followed. A merge and a removal issued in one shell without `&&` between them breaks it silently. Either the guidance should say "never chain these", or the removal belongs behind a guard that refuses when the branch is not an ancestor of HEAD.

2. `worktree-mechanics.md` says a `--ff-only` refusal in a concurrent wave "is not a fault" and gives the rebase remedy — and that is exactly what happened and what fixed it. But the coordinator checklist has no step that anticipates it, so the refusal arrives looking like the serialized case (which says STOP AND REPORT, do not resolve inline). Reading it wrong stops a healthy run. Worth a line at the merge step naming which case you are in before you read the refusal.

Evidence: /tmp/m10.out (exit 128), /tmp/wr10.out (exit 0), /tmp/acc10.out (exit 2, signal 3 STOP), then rebase onto e2b69ad -> 78b34e1, /tmp/acc10b.out (exit 0, all five OK). Note the sha moved 4b0c304 -> 78b34e1 in the rebase; the acceptance check found it anyway because it greps by bead id rather than trusting a sha recorded earlier, which is the run-2 hand-off defect already fixed.
```


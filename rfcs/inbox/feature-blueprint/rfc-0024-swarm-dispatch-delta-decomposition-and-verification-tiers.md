---
title: "RFC: Delta decomposition, verification tiers and the vocabulary a
  feature blueprint owes its implementers"
description: Nine defects the swarm-dispatch implementation phase found in
  its own FRD and feature blueprint, and the seven changes that would have
  prevented them
slug: rfc-0024-swarm-dispatch-delta-decomposition-and-verification-tiers
type: rfc
status: ready-for-review
updated_at: 2026-09-01T00:00:00Z
origin: agent-plugins
skill: retro-spec
skill_version: "1.5.0"
model: claude-opus-5
effort: high
runs_observed: 1
severity: blocker
plugin_version: "0.23.0"
---

# RFC: Delta decomposition, verification tiers and the vocabulary a feature blueprint owes its implementers

## Summary

Nineteen work orders implemented [[feature-swarm-dispatch]] — five `swarm`
commands, two widened interfaces, five end-to-end fixture variants and the
adoption of all of it into the `bead-bonanza` skill family. The phase
produced 81 recorded findings, and the friction concentrated in one place:
**the blueprint's delta blocks name obligations that its Behavior lines,
its AC coverage table and its ceremony table do not agree about.** Four
separate implementers each had to invent a term, an operation or a
verification story the documents referenced but never stated. Two
Definition-of-Done entries could not be closed as a result. The seven
proposals bind a delta's Intent, its Behavior lines, its ACs and its
declared verification tier to one another, and close the specific
under-specifications that cost this run the most.

## Provenance

| Field | Value |
|-------|-------|
| Origin project | `agent-plugins` |
| Phase executed | Implemented the swarm-dispatch plan — 19 work orders under `agent-plugins-o63s`, 18 closed and committed, 1 (this DoD item) closing |
| Documents exercised | [[feature-swarm-dispatch]], [[frd-swarm-dispatch]], [[container-swarm-cli]], [[module-bead-bonanza]], [[ac-traceability-standards]], [[swarm-output-standards]], [[contract-swarm-json]] |
| Documents produced | `cli/swarm/**` (five commands, two widened interfaces), `cli/testdata/swarm-fixtures/**` (five variants), the adoption edits across `plugin/skills/{bead-bonanza,run-batch,sprint-recovery,sprint-scope}` |
| Skill(s) followed | `run-batch` (per work order), `retro-spec` v1.5.0 (this retro) |
| Templates / standards read | rfc-template v1.3; `ac-traceability-standards` v0.4.0; `swarm-output-standards` v1.0; `contract-swarm-json` v2.0; `findings-capture-protocol` v0.6 |
| Plugin version | 0.23.0 |
| Session date | 2026-09-01 |

## Evidence

1. **The FRD uses "acceptance signals" as a defined term and never
   defines it.** `swarm-dispatch-frd.md:679` — "`2  pre-merge acceptance
   the acceptance signals per bead, deferring the two a merge decides`" —
   and `:799` — "`Then each bead in the dispatch is checked again with no
   signal deferred`". Nothing in the FRD, the blueprint or any work order
   enumerates the signals or says which two a merge decides. The RM
   implementer recorded: "grep -n 'signal'
   specs/features/swarm-dispatch/swarm-dispatch-frd.md returns 9 hits, and
   every one uses the word without enumerating the set… So the phrase
   'with no signal deferred' is unimplementable as written." The set was
   recovered from `plugin/skills/bead-bonanza/scripts/bb-status.sh` — a
   *skill script outside the container being specified*. The TRK
   implementer, working a different work order days earlier, performed the
   same derivation from the same script independently.

2. **A delta's Intent names four obligations; its Behavior lines cover
   three.** `swarm-dispatch-blueprint.md:44-49`, the Tracker delta: "the
   loop asks the tracker four questions the core feature did not need:
   which epics under a run are eligible to close, what is ready under a
   scope with the run's filters applied, what one assignee holds, and what
   a bead's acceptance signals say." The TRK implementer recorded: "only
   two of them have a SHALL line, so the fourth had to be decided rather
   than read… the AC table lists 004.11 as satisfied by TRK and proved by
   RM — so TRK owes something for it with no line saying what."

3. **The AC coverage table declares a verification tier at which the
   criterion cannot be verified.** The table declares
   `AC-SWARM-DISP-004.1 | MG, RM (proves) | e2e`. The criterion needs a
   dispatch branch that is *not* an ancestor of the target *after* a merge
   step that did not stop the command. Measured by the RM implementer:
   against a real git those are mutually exclusive — a successful
   `git merge --ff-only` has by construction put the branch on the target,
   and a failed one stops the acceptance at `AC-SWARM-DISP-004.6`, before
   the guard. It is proved at `unit` instead
   (`TestAWorktreeIsNotRemovedUnlessItsBranchIsAnAncestorOfTheTarget`,
   `cli/swarm/cli/accept_settle_test.go`). Carried by `agent-plugins-oyvf`.
   This is the reason the DoD entry "Every AC in the coverage table is
   verified at its declared level" ships **not met**.

4. **An interface-widening delta omitted four operations its own consuming
   deltas required.** Three later implementers each had to add one and
   justify it as in scope:
   - `Checkout(prefix, branch)` — MG: "the MG delta names no
     version-control operation for returning to the target after a rebase,
     and none existed. `git rebase <onto> <branch>` checks out `<branch>`,
     so the fast-forward that follows would advance the dispatch branch
     onto itself and leave the target unmoved."
   - `CurrentBranch(prefix)` — LD: "the delta names the branch default as
     'the one currently checked out' but the Version Control interface had
     no operation answering it — `Branches` lists every branch without
     saying which is checked out, and `Resolve` answers a sha."
   - `Root(prefix)` — WT: "No VCS method existed to read the repository's
     working-tree root, and `cli/swarm/cli` must never name 'git' itself."
   - `vcs.Rejection` and a push decoder — LD: "`AC-SWARM-DISP-005.3` and
     `.7` require the rejection cause and the refused path to reach the
     caller, but `Push` decoded with `discard`, whose error is the LAST
     line of stderr — which for a refused push is `error: failed to push
     some refs to <url>`, naming neither."

5. **The ceremony's step order is unimplementable against git for the
   concurrent case.** `swarm-dispatch-frd.md:675-686` fixes the order:
   step 4 `merge` ("in the concurrent case rebase onto the target read
   now, then fast-forward"), step 6 `remove` ("the worktree, then the
   branch"). Measured by the MG implementer, git 2.43.0: "after
   `git worktree add -q ../wt2 bb/dev-1`, `git rebase main bb/dev-1` ->
   exit 128, `fatal: bb/dev-1 is already used by worktree at .../wt2`."
   Every real concurrent acceptance has a standing dispatch worktree at
   step 4, so the concurrent arm cannot run as ordered. Carried by
   `agent-plugins-1vya` (P1). The e2e suite is green only because no
   fixture is both concurrent and carries a standing worktree — the RM
   implementer named the missing fixture precisely: "A fixture that is
   both concurrent AND has a standing worktree would reproduce 1vya end to
   end today. None exists."

6. **A ceremony step is named in the FRD and assigned to no delta.** The
   ceremony table's step 2 is "`pre-merge acceptance`". `swarm dispatch
   accept --help` covers only the post-merge half. The ADP-C implementer
   recorded: "dispatching.md's 'Verify against state' step 3 was the ONLY
   place the PRE-merge acceptance check is described, and no swarm command
   performs it — deleting the step whole, as the AC's list says, would
   have removed the only statement of a still-required step." Carried by
   `agent-plugins-bhaj`. This is the reason the DoD entry "Every prose
   section named in the ADP-C and ADP-D deltas is deleted" ships **not
   met**.

7. **A document enumerates a set that any later delta invalidates.**
   `container-swarm-cli.md`, the `## Data` entry for the end-to-end
   fixtures, read "Three are committed today — `core`, `status-attention`
   and `status-narrowed-scope`." The FIX work order added five and the
   sentence became wrong immediately; its implementer filed the exact
   string so FOLD-1 would not miss it. FOLD-1 updated it to "Eight are
   committed today" plus the eight names — the same shape, now a fresh
   hostage to the next fixture. Nothing checks the roster against the
   directory.

8. **The FRD's command-signature convention contradicts the ACs and the
   landed tests.** The WT implementer: "The ACs are silent on whether
   `--dispatch` is required, while the FRD's command signature (`swarm
   dispatch open --run <id> --dispatch <id> [--base <ref>] [--branch
   <name>]`) gives it no brackets, which in this doc's own convention
   reads as required. But the two pre-existing e2e tests from BR… call
   `dispatch open` with no `--dispatch` at all and expect success."
   Resolved by making the flag optional — a decision the documents left to
   an implementer at the point where the tests already constrained it.

9. **Two smaller silences forced decisions at implementation time.** PC:
   "the FRD and blueprint never say explicitly whether the RUN epic itself
   (as opposed to its descendants) is in this command's closing scope…
   This reads as the intended split… but it is inferred rather than
   stated." QF: "if the Status Collector's own worktree/completion
   measurement fails (tracker error mid-walk), `swarm ready` still exits
   clear and silently reports every bead 'held: false' rather than
   surfacing the measurement failure — there is no AC in
   FEAT-SWARM-DISP-002 covering this case."

## Proposals

### P1 — Enumerate the acceptance-signal set in the FRD, as a named list

**Problem.** The FRD uses "the acceptance signals" and "the two a merge
decides" as though both were defined, and defines neither
(`swarm-dispatch-frd.md:679`, `:799`). The only enumeration in the
repository is `plugin/skills/bead-bonanza/scripts/bb-status.sh` — a script
belonging to the *module that consumes* this container, which the container
must not depend on. Two implementers on two work orders each rediscovered
the set from that script; a third reader would have to do it again.

**Change.** Add a named list to `swarm-dispatch-frd.md`, immediately above
the ceremony table, and make every later use of the term cite it:

> **The acceptance signals**, in order: (1) the bead reads `closed` in the
> tracker; (2) a commit exists whose subject carries the bead id; (3) that
> commit is an ancestor of the sprint branch's HEAD; (4) the dispatch's
> worktree has been removed; (5) at least one `retro:`-prefixed comment was
> recorded on the bead. **Signals 3 and 4 are the two a merge decides** —
> both are false by construction before it, which is what "deferring the
> two a merge decides" means.

The same five are already stated in [[module-bead-bonanza]]'s Interface
guarantees; this makes the FRD the statement and the module the citation
rather than the reverse. Backward compatible: it names what the landed code
already does.

**Evidence.** Findings 1, 2. The next implementer of any command touching
acceptance reads one list instead of deriving it from a sibling module's
shell script.

### P2 — Bind a delta's Intent, its Behavior lines, its ACs and the ceremony to each other

**Problem.** Three independent instances of the same shape. A delta's
Intent named four obligations and its Behavior lines covered three, while
the AC coverage table nonetheless assigned an AC to the missing one
(finding 2). The ceremony table named eight steps and the delta blocks
implemented seven, with step 2 assigned to nobody (finding 6). Nothing in
the blueprint or in [[blueprint-readiness-checklist]] requires these four
surfaces to agree, so each was internally consistent and the set was not.

**Change.** Add to the feature-blueprint readiness gate, and to
[[blueprint-readiness-checklist]], three checks that are mechanical:

> - Every obligation a delta's **Intent** names SHALL have at least one
>   **Behavior** line under that delta, or SHALL be struck from the Intent.
> - Every AC the **AC coverage table** attributes to a delta SHALL be
>   traceable to a Behavior line of that delta.
> - WHERE the FRD states an ordered ceremony, every step SHALL be named by
>   exactly one delta's Behavior lines, or SHALL be listed under the
>   blueprint's Non-Goals with the reason.

Backward compatible: existing blueprints stay valid documents; the checks
bind at the next readiness gate.

**Evidence.** Findings 2, 6. TRK would have been given a Behavior line for
the acceptance-signal operation instead of inventing one; step 2 would have
been either assigned or declared out of scope before ADP-C had to choose
between deleting a still-required step and leaving the delta unsatisfied.

### P3 — Declare a verification tier only where the world it needs is constructible

**Problem.** The AC coverage table's `Verified by` column is authored
before any test exists, and nothing checks that the declared tier *can*
build the world the criterion describes. `AC-SWARM-DISP-004.1` was declared
`e2e` for a state that a single-actor end-to-end world cannot reach at all
(finding 3). The cost lands at the end: the implementer proves it at the
tier that works, the DoD entry then fails, and the feature ships with a
recorded exception.

**Change.** Extend [[ac-traceability-standards]] with one rule and one
escape hatch:

> A criterion SHALL be declared at `e2e` only WHERE its Given state is
> reachable by the actors the end-to-end tier has. WHERE the state requires
> an actor the tier does not model — a concurrent writer, a clock, an
> external failure — the criterion SHALL be declared at the tier that can
> isolate it, and the coverage table row SHALL carry a one-clause reason in
> its `Verified by` cell (e.g. `unit — needs a second actor`).

Add the same question to the blueprint readiness gate: for each `e2e` row,
name the fixture or the world that provides its Given.

**Evidence.** Finding 3. `AC-SWARM-DISP-004.1` would have been declared
`unit — needs a second actor` at authoring time, `agent-plugins-oyvf` would
never have been filed, and the DoD entry would have closed honestly.

### P4 — An interface-widening delta shall enumerate the operations its consumers call

**Problem.** The VCS and TRK deltas widen `vcs.VersionControl` and
`tracker.Tracker` for the whole feature, and are ordered first precisely so
the command deltas can build on them. But the deltas enumerate the
operations their *own* Behavior lines need, not the ones the consuming
deltas call. Four operations were therefore missing at the moment they were
needed (finding 4), each added by a command implementer inside a work order
whose scope was a command — the exact widening the run-batch ceremony asks
implementers not to do inline.

**Change.** Add to the feature-blueprint template's guidance for
interface-widening deltas, and to [[blueprint-readiness-checklist]]:

> WHERE a delta widens an interface that later deltas consume, its Behavior
> lines SHALL enumerate every operation those deltas call, derived by
> walking each consuming delta's own Behavior lines — not only the
> operations this delta's own criteria exercise. A consuming delta that
> needs an operation the interface delta does not name is a blueprint
> defect, not an implementation decision.

Backward compatible; it constrains authoring, not existing documents.

**Evidence.** Finding 4. Four operations would have been in the interface
delta's Behavior lines, where they belong, instead of being added under
three command work orders that then had to argue they were in scope.

### P5 — Fix the ceremony ordering that git makes unimplementable

**Problem.** The FRD fixes merge (step 4) before worktree removal (step 6),
and the concurrent arm of step 4 is a rebase. Git refuses to rebase a
branch another worktree has checked out — measured, exit 128 — and in every
real concurrent acceptance the dispatch's worktree is standing at step 4
(finding 5). The order is not a preference the implementation may reorder:
the FRD says so explicitly ("`a lockstep the steps hold to each other
rather than a sequence a caller may reorder`"). So the concurrent arm as
specified cannot run outside a fixture that happens to declare no worktree.

**Change.** Amend the ceremony in `swarm-dispatch-frd.md:675-686` to state
where the rebase runs, adopting whichever of these the owner prefers — both
are implementable and the RFC does not pick for them:

> - **(a)** Step 4's concurrent arm SHALL run the rebase **in the
>   dispatch's own worktree**, which is the one checkout that holds the
>   branch; or
> - **(b)** Step 6's worktree removal SHALL be split, with the worktree
>   removed **before** step 4 in the concurrent case and the branch removed
>   after step 5, and the FRD SHALL state that the in-flight signal for a
>   concurrent dispatch is its unmerged branch rather than its worktree.

Whichever is chosen, add the fixture the suite lacks: **concurrent *and*
carrying a standing worktree**, so the e2e tier reaches the arm at all.

**Evidence.** Finding 5, and `agent-plugins-1vya` (P1). Not backward
compatible with the landed code — that is the point; the code implements
the order as written and is broken by it.

### P6 — Retire enumerations a later delta invalidates

**Problem.** `container-swarm-cli.md`'s Data section names the fixture set
by roster. The roster was wrong the moment the FIX work order landed, was
corrected by FOLD-1 to a longer roster, and is wrong again on the next
fixture (finding 7). Nothing checks it, so the only mechanism keeping it
true is an implementer noticing and filing the string for a fold to pick
up — which is what happened, and is not a mechanism.

**Change.** Two parts, both narrow:

> - Replace the roster with the rule: "Each fixture is a directory under
>   `cli/testdata/swarm-fixtures/` holding `repository.bundle`,
>   `tracker.jsonl` and `fixture.toml`; `ls` that directory for the set."
>   State what `fixture.toml` may declare — including `[[worktree]]` and
>   `[refusing_remote]`, which the current Change Guidance does not mention.
> - Add to [[c4-standards]]: a blueprint SHALL NOT enumerate a set that its
>   own Change Guidance permits later work to extend, unless a declared
>   check binds the enumeration to the set.

**Evidence.** Finding 7. The Data entry stops being a maintenance item, and
the `fixture.toml` declaration surface — which two work orders extended
without the document noticing — gets stated once.

### P7 — Close three under-specifications the run had to decide silently

**Problem.** Three places where an implementer made a defensible choice the
documents should have made (findings 8, 9). Each is small; together they
are the residue of the same authoring habit — stating a shape without
stating its edges.

**Change.**

> - **Signature convention.** State in `swarm-dispatch-frd.md`'s command
>   block, once: "In a command signature, `[…]` marks an optional flag and
>   an unbracketed flag is required." Then reconcile `swarm dispatch open`:
>   `--dispatch` is optional (the landed behaviour and the pre-existing
>   tests), so bracket it.
> - **`swarm phase close`'s scope.** Add a sentence: "The run epic itself
>   is never in scope — the command closes exactly
>   `Descendants(run)` — and is closed at retro or post-flight."
> - **A failed measurement is not a clear answer.** Add an AC to
>   FEAT-SWARM-DISP-002: WHEN the reused completion/worktree measurement
>   cannot be established, THEN `swarm ready` SHALL report the measurement
>   failure rather than reporting every bead as not held.

All three are backward compatible except the third, which changes an exit
status in a case that currently exits clear — that is the defect.

**Evidence.** Findings 8, 9.

## Considered and Not Proposed

- **The `core` fixture's own description understates its manifest** —
  `cli/testdata/swarm-fixtures/core/fixture.toml` says "its manifest is
  complete through the scope stage", while the stored manifest also carries
  `collision_edges`, `partition_preview` and `review_passes`. Real, cited,
  and a one-line content fix to one file — the template's "just fix the
  document" exclusion applies, and no machinery change would have prevented
  it. Worth someone's next edit, not an RFC proposal.
- **An acceptance line landing in a file the same work order was forbidden
  to touch** — ADP-C's criteria required naming the commands in
  `module-bead-bonanza.md`'s Coordinator roster line and Change Guidance,
  while ADP-C's own Design said "Do NOT update module-bead-bonanza.md's
  component records here; FOLD-3 does that for both adoption deltas."
  ADP-C's agent could not satisfy it and handed it forward out of band;
  nothing in FOLD-3's own criteria named it, so it would have been dropped
  had the dispatch prompt not carried it. Root-caused to how the grooming
  skill derives per-item criteria against per-item file ownership — the fix
  lands in the framework repository, so it belongs to `retro-skill`, not
  here.
- **`bd dolt push` fails non-fast-forward for the whole run** — an
  environment defect already carried by `agent-plugins-wgxd`; no document
  change would address it.
- **The answer document has two JSON shapes depending on exit status** —
  reported as tooling friction by the LD implementer, whose first e2e case
  failed to unmarshal across a status change. This is
  [[contract-swarm-json]] behaving as specified; the friction was in
  writing one test struct for two shapes, which is a test-authoring
  problem. No evidence that the contract itself misleads.
- **Raising `contract-swarm-json`'s version for the manifest key** — the
  schema change was purely additive (one optional field), so no version
  raise was owed. Recorded here because a future reader auditing the DoD
  will ask.

## Non-Goals

- No new document types, and no change to the delta-block *format* — the
  proposals bind existing surfaces to each other rather than adding one.
- No relaxation of any consumer check. P3 changes what a blueprint may
  *declare*, not what the traceability check *enforces*.
- No re-litigation of the two-test-tier decision ([[adr-027]]); P3 and P5
  both work within it.
- P5 does not choose between its two options — the FRD's owner does.

## Adoption

**Independently adoptable:** P1, P3, P4, P6, P7. Each touches one document
and needs nothing from the others.

**P2 and P3 land best together** — both add rows to the same blueprint
readiness gate, and P2's third check ("every ceremony step is named by
exactly one delta") is what would have surfaced P5's defect at authoring
time rather than at implementation. Adopting P2 alone still helps; adopting
P3 alone still helps.

**P5 is the blocker and is not backward compatible.** It contradicts the
landed implementation, which is correct: the code implements the ceremony
as written and the ceremony is wrong. It should be resolved before the next
run dispatches concurrently, and its fixture half (a world that is
concurrent *and* holds a standing worktree) is a prerequisite for proving
either option.

**Touch points.** `specs/features/swarm-dispatch/swarm-dispatch-frd.md`
(P1, P5, P7); `specs/features/swarm-dispatch/swarm-dispatch-blueprint.md`
(P2, P3); `specs/standards/blueprint-readiness-checklist.md` (P2, P3, P4);
`specs/standards/ac-traceability-standards.md` (P3);
`specs/standards/c4-standards.md` (P6);
`specs/architecture/container-swarm-cli.md` (P6);
`cli/testdata/swarm-fixtures/` (P5).

**Expected effect.** Of the nine findings, seven were an implementer
supplying a term, an operation, a tier or an edge the documents referenced
but never stated — each discovered at the one moment it was most expensive,
inside a work order scoped to something else. The proposals move all seven
discoveries to the blueprint's readiness gate, where they cost an author a
paragraph.

## Resolution

<!-- FILLED BY THE RESOLVING REVIEW -->

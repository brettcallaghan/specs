<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/standards/work-order-definition-of-ready.md (version unversioned)
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: Work Order Definition of Ready
description: The readiness gate a work order must pass before it can be
  dispatched to an isolated agent — the checklist grooming enforces
slug: work-order-definition-of-ready
type: standards
status: active
---

# Work order definition of ready

## Contents

A work order must pass nine checks before dispatch:

1. Self-contained Problem statement (the fresh-agent test)
2. Right size (single session)
3. `## Design` present (non-trivial, non-epic orders)
4. `## Acceptance` criteria (bugs and features)
5. Kind-appropriate completeness (bug, feature, task, chore, and epic)
6. Dependencies wired
7. Tier assigned, with a reason
8. C4 blast radius declared (if the repo uses C4)
9. No open decisions

After the checks come what to do when a work order can't reach ready, and
three worked examples that take an under-groomed order to ready.

A work order is ready when a fresh agent with zero session context can
complete it correctly from the work order alone, with no planning conversation
and no one to ask. Run every work order against the checks below. They're a
gate: an un-ready work order handed to an isolated agent is the failure the
whole concept exists to prevent.

This is the single definition of "ready to dispatch." The `spec` skill owns it,
and the grooming skills enforce it: `groom-beads` on any unit of work, and
`groom-feature-blueprint` on work decomposed from a blueprint.
`work-order-field-guide.md` gives field-by-field guidance, including the
distinction between the three text fields, the tier-routing rubric, and the C4
conventions. `work-order-template.md` is the format each check maps onto.

## The checks

### 1. Self-contained Problem statement (the fresh-agent test)
Read the `## Problem` section as if you just woke up in the repo with no chat
history. Can you act on it? It SHALL name concrete anchors — file paths,
function/class names, endpoints, tables, user-facing behavior — not point at a
conversation.

An anchor SHALL be durable under edits to the same file. File paths, symbol
names, headings, and short quoted strings are durable. Line numbers and line
ranges aren't, and neither are byte counts or item counts. WHERE a work order
names a position, it SHALL name the content at that position — quote the
heading or the first few words — and MAY give a line number only as a
convenience alongside it. In a queue where several orders touch one file, every
line number is stale before the second order is dispatched.

These red flags fail this check immediately: "as we discussed," "the file
above," "like the last one," "continue the refactor," "you know the pattern,"
and bare pronouns with no referent ("fix it," "wire this up").

To fix the order, inline the missing context instead of deleting the sentence.
If the author wrote "finish the refactor from earlier," find what that refactor
is and write it down.

### 2. Right size (single session)
The work order should be completable in one focused session, with a rough
ceiling of ~1 hour of agent work. The test is whether you can write crisp,
checkable acceptance criteria. If you can't, the scope is too fuzzy or too big.

Convert an oversized order to an epic (`kind: epic`) with child work orders,
each declaring `parent_slug`. Each child SHALL independently pass this whole
checklist. Don't keep an oversized order as a single order to avoid the work of
splitting it.

### 3. `## Design` present (non-trivial, non-epic orders)
Any order larger than a one-liner needs a Design: the chosen approach, and the
decisions that a time-pressured agent would otherwise re-litigate. The Design
is mutable. It records the alternatives that were weighed and the one that was
chosen. Keeping the rejected options is what stops a later agent from
re-litigating them.

It SHALL NOT leave the choice itself open. A conditional that a fresh agent can
resolve mechanically from the repository ("use the API's batch endpoint; if it
is absent on the pinned version, loop the single-item call") is a chosen
approach with a stated fallback, and it passes. A conditional whose resolution
needs a judgment the order doesn't supply ("leaning toward X because Y; fall
back to Z if that doesn't work out") is an open decision. It fails this check
and check 9. A real feature or task without a Design is a sign that the order
isn't groomed yet. Only a `kind: epic` order omits the Design.

### 4. `## Acceptance` criteria (`kind: bug` & `kind: feature`)
Acceptance criteria are outcome-focused, checkable, and
implementation-agnostic. They describe the observable end state, not the steps.

- Good: `- [ ] Expired tokens return 401` / `- [ ] Refresh rotates the token`
- Bad: `- [ ] Edit middleware.go` (that's a step, and it belongs in Design)

The criteria are the shared contract. An isolated agent uses them to know it's
done, and the orchestrator or verifier uses them to confirm it.

Every executable a criterion names SHALL resolve — on `PATH` in the
environment the work order will run in, or as one of the project's declared
check ids. An acceptance criterion that names a missing tool isn't strict. It's
unsatisfiable.

In one measured run, two agents independently hit the same criterion (`sh -n`
and `shellcheck`, "as the project's checks run them"), where the project ran
neither and neither was installed. Both did the right thing: they recorded the
criterion as UNVERIFIED and substituted what was available. But the criterion
should never have passed the gate. Check the names at grooming time, when there
is still somebody to fix them.

A criterion SHALL NOT cite a position. An agent can still find a path plus a
quoted string, or a class, function, file, or component name. A line number
rots between grooming and dispatch, and it does in practice. In one run's
measured findings, two of six anchors had moved before the beads were
dispatched, because a sibling in the same batch edited the file. Where a line
number helps, write it as a hint beside the string, never as the handle.

A criterion SHALL NOT contradict the work order's own `## Design` scope. Where
they disagree, one of them is wrong, and an implementing agent has no way to
tell which. Resolve the conflict here.

### 5. Kind-appropriate completeness
- **bug**: steps to reproduce, expected versus actual behavior, and where it
  manifests. A bug without a repro is really a research task. Either add a
  repro or reclassify its `kind`.
- **feature**: user-facing intent and acceptance criteria.
- **task**: the concrete change and its success signal.
- **chore**: what "done" looks like (for example, "CI green with the new
  linter").
- **epic**: a coherent goal, plus child work orders that cover it. The epic
  itself isn't directly executable.

### 6. Dependencies wired
Wire only real edges. Put each in the correct field (`depends_on` / `blocks`,
`parent_slug`, `related`, `discovered_from`), referencing another work order by
slug. Ordering SHALL live in the graph, not in prose or priority. A missing
real edge is a dispatch-time race, and a fabricated edge over-serializes.

### 7. Tier assigned (+ reason)
Set `tier` from the closed, ordered set (`light | standard | advanced |
frontier`, least to most capable). The tier is never null, and it always has a
short `tier_reason`. Assign the lowest tier that can complete the order
correctly. A named `model` is an optional override, not the routing field. See
`work-order-field-guide.md` for the rubric and the project lineup mapping.

### 8. C4 blast radius declared (if the repo uses C4)
Declare `c4_changes`, `c4_impacts`, and `c4_docs`, using only component names
that exist in the repo's C4 docs. Changed components SHALL get an acceptance
box to update their C4 doc. If the repo has no C4 docs, skip this check
cleanly.

### 9. No open decisions
The work order SHALL carry no unresolved choice. It fails this check when it
carries a "DECISION REQUIRED" marker, a "TBD" marker, or an "open question"
marker, or an options list with no option marked as selected.

A work order that fails this check SHALL be labelled `human` — the label
`bd human list`, `bd human respond`, `bd human dismiss` and `bd human stats`
all read. That keeps one label with tooling behind it, instead of adding a
second label that carries the same signal with no tooling.

Where an open choice would change the order's observable behavior, it's an
acceptance-criteria gap, and it fails check 4 as well. Pin the behavior in the
AC, not in prose.

A choice that a diff reader wouldn't recognize as a choice, such as an internal
helper's name or the order of two independent guard clauses, isn't an open
decision and doesn't fail this check. The implementing agent records what it
picked in its retro.

This check is why a groomed order can be dispatched without a conversation. An
orchestrator that must surface decisions at dispatch time is doing refinement
inside execution. That work belongs here, at grooming.

## When a work order can't reach ready

A work order that can't be brought to this bar isn't dispatched silently. It's
flagged for the author on the tracker, with `blocked-needs-info` when a
specific fact is missing or `blocked-ungroomable` when it needs redesign, and
it's excluded from dispatch until resolved. That flag lives on the tracker
issue, not in the work order frontmatter (see the status note in
`work-order-template.md`).

## Worked examples

**Example 1: under-groomed to ready (feature)**

Before:
```
title: Add auth
Problem: Wire up login like we talked about, similar to the other service.
kind: feature   priority: 2
```
It fails check 1 (not self-contained), check 3 (no Design), and check 4 (no
Acceptance), and it has no dependencies, tier, or C4 fields.

After:
```
title: Add JWT login to the accounts API
kind: feature   priority: 1
parent_slug: work-order-accounts-auth
tier: standard
tier_reason: well-specified, mirrors an existing pattern; low ambiguity
est_size: M
c4_changes: [accounts-service]
c4_impacts: [web-spa]
c4_docs: [docs/c4/accounts-service/component.md]

## Problem
Add username/password login to the accounts API (services/accounts). Issue a
signed JWT on success. Mirrors the scheme already used in services/billing
(see billing/auth/jwt.go): HS256, 15-minute access token, refresh via rotation.
Endpoint: POST /v1/login on the existing gin router in accounts/router.go.

## Design
Reuse billing's jwt helper by lifting it into pkg/auth (shared). Store the
signing key from env (ACCOUNTS_JWT_KEY). Add middleware AuthRequired for
protected routes. Uses golang-jwt/jwt v5 (already a billing dep) over
go-jose — same JWT support, and it is already vendored here.

## Acceptance
- [ ] POST /v1/login with valid creds returns 200 + access & refresh tokens
- [ ] Invalid creds return 401
- [ ] Access token expires after 15m; expired token → 401
- [ ] Refresh rotates the token (old refresh invalidated)
- [ ] C4 component doc for accounts-service updated to show the new auth flow
```

**Example 2: oversized order to epic and children**

Before: `title: Build the notifications system` (feature, no other fields).

After: an epic, "Notifications system," with child work orders such as "Define
notification event schema," "Add email channel adapter (SES)," "Add in-app
notification store + read/unread," and "Wire notification triggers on order
events." Each child independently passes this checklist and has `parent_slug`
set to the epic. Children have `depends_on` edges where one genuinely depends
on another (the schema blocks the adapters). Each child gets its own tier:
advanced for the schema design, and standard for the SES adapter.

**Example 3: bug missing repro**

Before: `title: Fix the flaky checkout test` (bug, empty Problem).

After: a Problem with the failing test path
(`checkout/checkout_test.go::TestConcurrentCheckout`), the observed flakiness
(fails ~1 in 5 under `-race` and passes in isolation), the suspected cause
(shared cart fixture mutated across goroutines), expected versus actual
behavior, and acceptance (`- [ ] Test passes 50/50 runs under -race`). The
order has tier `standard` and `c4_changes` `[checkout-service]`.

<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/standards/work-order-definition-of-ready.md
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

# Work Order definition of ready

## Contents

The nine checks a Work Order must pass before dispatch:

1. Self-contained Problem statement (the fresh-agent test)
2. Right size (single session)
3. `## Design` present (non-trivial, non-epic orders)
4. `## Acceptance` criteria (bugs & features)
5. Kind-appropriate completeness (bug / feature / task / chore / epic)
6. Dependencies wired
7. Tier assigned (+ reason)
8. C4 blast radius declared (if the repo uses C4)
9. No open decisions

Then: what to do when a Work Order can't reach ready, and three worked
under-groomed → ready examples.

A Work Order is **ready** when a fresh agent with zero session context can
complete it correctly from the Work Order alone, with no planning conversation
and no one to ask. Run every Work Order against the checks below. Treat them as
a gate, not a suggestion: the whole concept exists to prevent an un-ready Work
Order from reaching an isolated agent.

This standard is the single definition of "ready to dispatch." The `spec` skill
owns it, and the grooming skills enforce it: `groom-beads` on any unit of work,
and `groom-feature-blueprint` on work decomposed from a blueprint.
The work-order field guide holds the field-by-field guidance: the distinction
between the three text fields, the tier-routing rubric, and the C4 conventions.
Each check maps onto the format in the `spec` skill's Work Order template.

## The checks

### 1. Self-contained Problem statement (the fresh-agent test)
Read the `## Problem` section as if you just woke up in the repo with no chat
history. Can you act on it? It SHALL name concrete anchors — file paths,
function/class names, endpoints, tables, user-facing behavior — not point at a
conversation.

An anchor SHALL be durable under edits to the same file. File paths, symbol
names, headings, and short quoted strings are durable. Line numbers, line
ranges, byte counts, and item counts aren't. WHERE a work order
names a position, it SHALL name the content at that position — quote the
heading or the first few words — and MAY give a line number only as a
convenience alongside it. In a queue where several orders touch one file, every
line number is stale before the second order is dispatched.

Red flags that fail this check immediately: "as we discussed," "the file
above," "like the last one," "continue the refactor," "you know the pattern,"
and bare pronouns with no referent ("fix it," "wire this up").

The fix is to **inline the missing context**, not to delete the sentence. If
the author wrote "finish the refactor from earlier," find out what that
refactor is and write it down.

### 2. Right size (single session)
The Work Order should be completable in one focused session, with a rough
ceiling of about 1 hour of agent work. The real test is whether you can write
crisp, checkable acceptance criteria. If you can't, the scope is too fuzzy or
too big.

Convert an oversized order to an **epic** (`kind: epic`) with **child Work
Orders**, each declaring `parent_slug`. Each child SHALL independently pass
this whole checklist. Don't leave a boulder tagged as a single order just
because splitting it takes work.

### 3. `## Design` present (non-trivial, non-epic orders)
Anything past a one-liner needs a Design: the chosen approach, and the
decisions that a time-pressured agent would otherwise re-litigate. The Design
is mutable. It records the alternatives that were weighed and the one that was
chosen, and preserving the rejected options is what stops a later agent from
re-litigating them.
It SHALL NOT leave the choice itself open. A conditional that a fresh agent can
resolve mechanically from the repository ("use the API's batch endpoint; if it
is absent on the pinned version, loop the single-item call") is a chosen
approach with a stated fallback, and it passes. A conditional whose resolution
needs a judgment the order doesn't supply ("leaning toward X because Y; fall
back to Z if that doesn't work out") is an open decision, and it fails both
this check and check 9. An absent Design on a real feature or task is a sign
that the order isn't groomed yet. Only an order of `kind: epic` omits it.

### 4. `## Acceptance` criteria (`kind: bug` & `kind: feature`)
Acceptance criteria are outcome-focused, checkable, and
implementation-agnostic. Describe the observable end state, not the steps.

- Good: `- [ ] Expired tokens return 401` / `- [ ] Refresh rotates the token`
- Bad: `- [ ] Edit middleware.go` (that's a step, and it belongs in Design)

The criteria are the shared contract. They're how an isolated agent knows it's
done, and how the orchestrator or verifier confirms it.

**Every executable a criterion names SHALL resolve** — on `PATH` in the
environment the work order will run in, or as one of the project's declared
check ids. An acceptance criterion that names a tool that isn't there isn't
strict. It's unsatisfiable. Check the names at grooming time, while someone
can still fix them.

**A criterion SHALL NOT cite a position.** An agent can still find a path plus
a quoted string, or a class, function, file, or component name. A line number
rots between grooming and dispatch, for example when a sibling order in the
same batch edits the file. Where a line number helps, write it as a hint beside
the string, never as the handle.

**A criterion SHALL NOT contradict the work order's own `## Design` scope.**
Where they disagree, one of them is wrong, and an implementing agent has no way
to tell which. Resolve it here.

### 5. Kind-appropriate completeness
- **bug** — steps to reproduce, expected versus actual behavior, and where it
  manifests. A bug without a repro is a research task in disguise. Either add
  a repro or reclassify its `kind`.
- **feature** — user-facing intent and acceptance criteria.
- **task** — the concrete change and its success signal.
- **chore** — what "done" looks like (for example, "CI green with the new
  linter").
- **epic** — a coherent goal plus the child Work Orders that cover it. The
  epic itself isn't directly executable.

### 6. Dependencies wired
Wire real edges only, each in the correct field (`depends_on` / `blocks`,
`parent_slug`, `related`, `discovered_from`) and each referencing another Work
Order by slug.
Ordering SHALL live in the graph, not in prose or priority. A missing real edge
is a race at dispatch time. A fabricated edge over-serializes.

### 7. Tier assigned (+ reason)
Set `tier` from the closed, ordered set (`light | standard | advanced |
frontier`, least to most capable). It's never null, and it always has a short
`tier_reason`. Assign the lowest tier that can complete the order correctly. A
named `model` is an optional override, not the routing field. See the
work-order field guide for the rubric and the project lineup mapping.

### 8. C4 blast radius declared (if the repo uses C4)
Declare `c4_changes`, `c4_impacts`, and `c4_docs`, using only component names
that exist in the repo's C4 docs.
Changed components SHALL get an acceptance box to update
their C4 doc. Skip this check cleanly if the repo has no C4 docs.

### 9. No open decisions
The work order SHALL carry no unresolved choice. It fails this check when it
carries a "DECISION REQUIRED," "TBD," or "open question" marker, or an options
list with no option marked as selected.

A work order that fails this check SHALL be labelled **`human`** — the label
`bd human list`, `bd human respond`, `bd human dismiss` and `bd human stats`
all read. That's one label with tooling behind it, rather than a second label
that carries the same signal with none.

Where an open choice would change the order's observable behavior, it's an
acceptance-criteria gap and fails check 4 as well. Pin the behavior in the AC,
not in prose.

A choice that a diff reader wouldn't recognize *as* a choice, such as an
internal helper's name or the order of two independent guard clauses, isn't an
open decision and doesn't fail this check. The implementing agent records what
it picked in its retro.

**This check is why a groomed order is dispatchable without a conversation.**
An orchestrator that must surface decisions at dispatch time is refining
inside execution. That work belongs here, at grooming.

## When a Work Order can't reach ready

A Work Order that can't reach this bar is **not** silently dispatched. It's
flagged for the author on the tracker and excluded from dispatch until the
flag is resolved. The flag is `blocked-needs-info` when a specific fact is
missing, and `blocked-ungroomable` when the order needs redesign. The flag
lives on the tracker issue, not in the Work Order frontmatter (see the status
note in the `spec` skill's Work Order template).

## Worked examples

**Example 1 — under-groomed → ready (feature)**

Before:
```
title: Add auth
Problem: Wire up login like we talked about, similar to the other service.
kind: feature   priority: 2
```
Fails: not self-contained (#1), no Design (#3), no Acceptance (#4), and no
dependencies, tier, or C4 fields.

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

**Example 2 — boulder → epic + children**

Before: `title: Build the notifications system` (feature, no other fields).

After: an **epic**, "Notifications system," with child Work Orders such as
"Define notification event schema," "Add email channel adapter (SES)," "Add
in-app notification store + read/unread," and "Wire notification triggers on
order events." Each child independently passes this checklist and sets
`parent_slug` to the epic. A child has `depends_on` edges where it genuinely
depends on another (the schema blocks the adapters), and each child has its own
tier (schema design → advanced; SES adapter → standard).

**Example 3 — bug missing repro**

Before: `title: Fix the flaky checkout test` (bug, empty Problem).

After: a Problem with the failing test path
(`checkout/checkout_test.go::TestConcurrentCheckout`), the observed flakiness
(fails about 1 in 5 under `-race`, passes in isolation), the suspected cause
(shared cart fixture mutated across goroutines), expected versus actual
behavior, and acceptance (`- [ ] Test passes 50/50 runs under -race`). Tier:
`standard`. `c4_changes`: `[checkout-service]`.

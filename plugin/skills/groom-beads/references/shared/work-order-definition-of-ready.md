<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/standards/work-order-definition-of-ready.md (version 1.4)
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: Work Order Definition of Ready
description: The readiness gate a work order must pass before it can be
  dispatched to an isolated agent — the checklist grooming enforces
slug: work-order-definition-of-ready
type: standards
status: active
updated_at: 2026-08-13T00:00:00Z
version: "1.4"
---

# Work Order Definition of Ready

## Contents

The nine checks a work order must pass before dispatch:

1. Self-contained Problem statement (the fresh-agent test)
2. Right size (single session)
3. `## Design` present (non-trivial, non-epic orders)
4. `## Acceptance` criteria (bugs & features)
5. Kind-appropriate completeness (bug / feature / task / chore / epic)
6. Dependencies wired
7. Tier assigned (+ reason)
8. C4 blast radius declared (IF the repo uses C4)
9. No open decisions

Then: what to do when a work order can't reach ready, and three worked
under-groomed → ready examples.

A work order is **ready** when a fresh agent with zero session context can
complete it correctly from the work order alone — no planning conversation, no
one to ask. Run every work order against the checks below. Treat these as a
gate, not a suggestion: an un-ready work order handed to an isolated agent is
the failure mode the whole concept exists to prevent.

This is the single definition of "ready to dispatch." The `spec` skill owns it;
the grooming skills enforce it (`groom-beads` on any unit of work,
`groom-feature-blueprint` on blueprint-decomposed work). Field-by-field
guidance — the distinction between the three text fields, the tier-routing
rubric, the C4 conventions — is in `work-order-field-guide.md`. The format each
check maps onto is `work-order-template.md`.

## The checks

### 1. Self-contained Problem statement (the fresh-agent test)
Read the `## Problem` section as if you just woke up in the repo with no chat
history. Can you act on it? It SHALL name concrete anchors — file paths,
function/class names, endpoints, tables, user-facing behavior — not point at a
conversation.

An anchor SHALL be durable under edits to the same file. File paths, symbol
names, headings and short quoted strings are durable; line numbers and line
ranges are not, and neither are byte counts or item counts. WHERE a work order
names a position, it SHALL name the content at that position — quote the
heading or the first few words — and MAY give a line number only as a
convenience alongside it. In a queue where several orders touch one file, every
line number is stale before the second order is dispatched.

Red flags that fail this check immediately: "as we discussed", "the file
above", "like the last one", "continue the refactor", "you know the pattern",
bare pronouns with no referent ("fix it", "wire this up").

The fix is to **inline the missing context**, not delete the sentence. If the
author wrote "finish the refactor from earlier," find what that refactor is and
write it down.

### 2. Right size (single session)
The work order should be completable in one focused session — rough ceiling ~1
hour of agent work. The real test: can you write crisp, checkable acceptance
criteria? If you can't, the scope is too fuzzy or too big.

Oversized → convert to an **epic** (`kind: epic`) with **child work orders**
(each declaring `parent_slug`). Each child SHALL independently pass this whole
checklist. Don't leave a boulder tagged as a single order just because
splitting is work.

### 3. `## Design` present (non-trivial, non-epic orders)
Anything past a one-liner needs a Design: the chosen approach and the decisions
that would otherwise get re-litigated by a time-pressured agent. It's mutable,
and it records the alternatives that were weighed and the one that was chosen;
preserving the rejected options is what stops a later agent re-litigating them.
It SHALL NOT leave the choice itself open. A conditional a fresh agent can
resolve mechanically from the repository ("use the API's batch endpoint; if it
is absent on the pinned version, loop the single-item call") is a chosen
approach with a stated fallback and passes. A conditional whose resolution needs
a judgment the order does not supply ("leaning toward X because Y; fall back to
Z if that doesn't work out") is an open decision and fails this check — and
check 9. An absent Design on a real feature/task is a sign the order isn't
groomed yet. Omitted only for `kind: epic`.

### 4. `## Acceptance` criteria (`kind: bug` & `kind: feature`)
Outcome-focused, checkable, implementation-agnostic. Describe the observable end
state, not the steps.

- Good: `- [ ] Expired tokens return 401` / `- [ ] Refresh rotates the token`
- Bad: `- [ ] Edit middleware.go` (that's a step, and it belongs in Design)

This is the shared contract: it's how an isolated agent knows it's done and how
the orchestrator/verifier confirms it.

### 5. Kind-appropriate completeness
- **bug** — steps to reproduce, expected vs. actual behavior, and where it
  manifests. A bug without repro is a research task in disguise; either add
  repro or re-classify its `kind`.
- **feature** — user-facing intent and acceptance criteria.
- **task** — the concrete change and its success signal.
- **chore** — what "done" looks like (e.g. "CI green with the new linter").
- **epic** — a coherent goal plus child work orders that cover it; the epic
  itself isn't directly executable.

### 6. Dependencies wired
Real edges only, correct field (`depends_on` / `blocks`, `parent_slug`,
`related`, `discovered_from`), each referencing another work order by slug.
Ordering SHALL live in the graph, not in prose or priority. A missing real edge
is a wave-time race; a fabricated edge over-serializes.

### 7. Tier assigned (+ reason)
`tier` from the closed, ordered set (`light | standard | advanced |
frontier`, least to most capable), never null, always with a short
`tier_reason`. Assign the LOWEST tier that can complete the order
correctly. A named `model` is an optional override, not the routing field.
See `work-order-field-guide.md` for the rubric and the project lineup
mapping.

### 8. C4 blast radius declared (IF the repo uses C4)
`c4_changes` / `c4_impacts` / `c4_docs`, using only component names that exist
in the repo's C4 docs. Changed components SHALL get an acceptance box to update
their C4 doc. Skip cleanly if the repo has no C4 docs.

### 9. No open decisions
The work order SHALL carry no unresolved choice. It fails this check WHEN it
carries a decision-required label, a "DECISION REQUIRED" / "TBD" / "open
question" marker, or an options list with no option marked as selected.

WHERE an open choice would change the order's observable behavior, it is an
acceptance-criteria gap and fails check 4 as well — pin the behavior in the AC,
not in prose.

A choice a diff-reader would not recognize AS a choice — an internal helper's
name, the order of two independent guard clauses — is not an open decision and
does not fail this check. The implementing agent records what it picked in its
retro.

**This check is why a groomed order is dispatchable without a conversation.** An
orchestrator that must surface decisions at dispatch time is doing refinement
inside execution; that work belongs here, at grooming.

## When a work order can't reach ready

A work order that cannot be brought to this bar is **not** silently dispatched.
It is flagged for the author on the tracker — `blocked-needs-info` when a
specific fact is missing, `blocked-ungroomable` when it needs redesign — and
excluded from dispatch until resolved. That flag lives on the tracker issue, not
in the work order frontmatter (see the status note in `work-order-template.md`).

## Worked examples

**Example 1 — under-groomed → ready (feature)**

Before:
```
title: Add auth
Problem: Wire up login like we talked about, similar to the other service.
kind: feature   priority: 2
```
Fails: not self-contained (#1), no Design (#3), no Acceptance (#4), no
deps/tier/C4.

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

After: an **epic** "Notifications system" with child work orders such as
"Define notification event schema", "Add email channel adapter (SES)", "Add
in-app notification store + read/unread", "Wire notification triggers on order
events" — each independently passing this checklist, `parent_slug` set to the
epic, `depends_on` edges where a child genuinely depends on another (schema
blocks the adapters), and tiers assigned per child (schema design → advanced;
SES adapter → standard).

**Example 3 — bug missing repro**

Before: `title: Fix the flaky checkout test` (bug, empty Problem).

After: a Problem with the failing test path
(`checkout/checkout_test.go::TestConcurrentCheckout`), the observed flakiness
(fails ~1 in 5 under `-race`, passes in isolation), the suspected cause (shared
cart fixture mutated across goroutines), expected vs. actual, and acceptance
(`- [ ] Test passes 50/50 runs under -race`). Tier `standard`; `c4_changes`
`[checkout-service]`.

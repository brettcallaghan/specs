<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/standards/blueprint-content-standards.md
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: Blueprint Content Standards
description: What an architecture blueprint states and what it leaves to the
  code, the FRD, and the host blueprint, plus the audit an author runs before
  review
slug: blueprint-content-standards
type: standards
status: active
---

# Blueprint content standards

## Contents

- Every architecture blueprint — rules 1 to 8
- Component and Module Blueprints — rules 9 to 13
- Blueprint audit — the checklist to run before `ready-for-review`
- References

An architecture blueprint holds what the code can't say: an element's
responsibility, the decision it hides, its interface guarantees, its
invariants, and where it is designed to grow ([[adr-032]]). Everything else
stays with its owner. The code owns implementation detail. The FRD owns
behavior and acceptance criteria ([[adr-033]]). The host blueprint owns how
sibling elements interact ([[adr-009]]).

Rules 1 to 8 apply to every System, Container, Module, and Component
Blueprint. Rules 9 to 13 apply to Component and Module Blueprints. Templates
and skills cite these rules by number and don't restate them.

Each rule ends with what enforces it. "Planned" marks a `specs check` finding
that the CLI doesn't report yet. Until it does, the blueprint audit covers
that rule.

---

## Every architecture blueprint

### 1. State only what must stay true

A blueprint SHALL NOT state counts, versions, member lists, or line numbers
of code under its `source_paths`.

**Rationale:** Facts change with the code and nothing updates the blueprint.
Rules are enforced or reviewable.

**Enforced by:** the code-reference check (planned); the blueprint audit.

---

### 2. Link what the code reveals

WHEN reading `source_paths` would reveal a fact, the blueprint SHALL link the
code instead of stating the fact. A link names a repository-relative path or
symbol in backticks (`src/planner.rs`, `CodeBackend`), never a URL or a line
number.

**Rationale:** Derivable text costs the reader twice, once to read it and
once to doubt it, and it buries the rules. A backticked path or symbol is a
name that a tool can resolve against the source tree.

**Enforced by:** the blueprint audit.

---

### 3. Point the interface at its definition

The Interface section SHALL point to the defining type or export map. It
SHALL name a member only to state a guarantee that the signature can't
express.

**Rationale:** Generated API docs are the member reference.

**Enforced by:** the blueprint audit.

---

### 4. Leave behavior to the FRD

A blueprint SHALL NOT restate FRD behavior or list implemented features or
acceptance criteria. It MAY cite an AC id as the reason for a rule.

**Rationale:** Two owners of one behavior diverge, and the FRD is the copy
that tests cite. Element-to-AC traceability comes from test citations
([[ac-traceability-standards]]).

**Enforced by:** the blueprint audit.

---

### 5. Say what enforces each invariant

Each invariant SHALL end with "Enforced by:" naming a test, a check, a lint,
or "review only".

**Rationale:** Readers know which rules to trust and which to verify. "Review
only" marks a missing test.

**Enforced by:** the blueprint audit.

---

### 6. State a shared rule once

WHEN two elements share a rule, the blueprint of the element that enforces it
SHALL state the rule, and the other blueprint SHALL link to it.

**Rationale:** One rule stated in two places is information leakage in prose:
the copies diverge.

**Enforced by:** the blueprint audit.

---

### 7. Write `description` for triage

The `description` property SHALL take the form "<what it owns>. Change it to
<kinds of change>."

**Rationale:** Agents choose elements from `specs list`, which shows the
description.

**Enforced by:** review.

---

### 8. Name owners, not neighbors

A blueprint SHALL name the owner of each responsibility it excludes. It SHALL
NOT describe interactions with sibling elements. Its References section SHALL
list only the host, Contracts, ADRs, and guides.

**Rationale:** Edges belong to the host's blueprint, and they drift when
restated ([[c4-standards]]).

**Enforced by:** the blueprint audit.

---

## Component and Module Blueprints

### 9. Own a disjoint part of the source tree

WHILE status is `active`, a Component or Module Blueprint SHALL declare
`source_paths` and `test_paths`. Its `source_paths` SHALL NOT overlap a
sibling's.

**Rationale:** A file maps to one blueprint, so parallel work on siblings
never touches the same files.

**Enforced by:** the ownership check (planned).

---

### 10. List isolation tests first

The Verification gates SHALL list isolation tests from `test_paths` first.
They SHALL list wider suites second, each with the condition that requires
it.

**Rationale:** Isolation tests are fast and point at the broken element. The
condition says when the slow suite is owed.

**Enforced by:** the gate check (planned); review.

---

### 11. Give each pattern instance one line

In a pattern blueprint (`kind: pattern`), each instance SHALL get one line:
where it lives and how it differs.

**Rationale:** An instance's mechanics are in its code.

**Enforced by:** the blueprint audit.

---

### 12. State where the element grows

A blueprint SHALL state where its element is designed to grow and the wrong
way to extend it, or state that the element has no extension points and why.

**Rationale:** Code shows a seam's mechanism. It doesn't show that the seam
is the intended way in.

**Enforced by:** the blueprint audit.

---

### 13. Keep a Component Blueprint near 250 lines

A Component Blueprint SHOULD stay near 250 lines. Past that, cut it or split
the component.

**Rationale:** A budget prompts the audit. It's advice only, because length
can't measure quality.

**Enforced by:** none.

---

## Blueprint audit

Run this audit on an architecture blueprint before you set it to
`ready-for-review`:

1. Run `specs check <path>` and fix what it reports.
2. Classify each statement as a rule, a fact, a behavior, or an edge.
3. Apply rules 1 to 8.
4. Confirm that the Hides line, the extension points, and an "Enforced by:"
   note on each invariant exist.

---

## References

- [[adr-032]] — blueprints state rules, not code facts
- [[adr-033]] — behavior belongs to the FRD
- [[adr-009]] — parents own the edges among their children
- [[adr-011]] — blueprints declare targeted verification gates
- [[c4-standards]] — the C4 levels, and where edges and scope statements live
- [[ac-traceability-standards]] — how tests cite acceptance criteria
- [[guide-component-design]] — the design questions behind these rules

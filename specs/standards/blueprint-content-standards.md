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

- Rules for every architecture blueprint
- Rules for Component and Module Blueprints
- Blueprint audit — the checklist to run before `ready-for-review`
- References

An architecture blueprint holds what the code can't say: an element's
responsibility, the decision it hides, its interface guarantees, its
invariants, and where it is designed to grow ([[adr-032]]). Everything else
stays with its owner. The code owns implementation detail. The FRD owns
behavior and acceptance criteria ([[adr-033]]). The host blueprint owns how
sibling elements interact ([[adr-009]]).

Each section heading names the blueprints its rules apply to. Templates and
skills cite a rule by its heading and don't restate it.

---

## Rules for every architecture blueprint

### State only what must stay true

A blueprint SHALL NOT state counts, versions, member lists, or line numbers
of code under its `source_paths`.

**Rationale:** Facts change with the code and nothing updates the blueprint.
Rules are enforced or reviewable.

---

### Link what the code reveals

WHEN reading `source_paths` would reveal a fact, the blueprint SHALL link the
code instead of stating the fact. A link names a repository-relative path or
symbol in backticks (`src/planner.rs`, `CodeBackend`), never a URL or a line
number.

**Rationale:** Derivable text costs the reader twice, once to read it and
once to doubt it, and it buries the rules. A backticked path or symbol is a
name that a tool can resolve against the source tree.

---

### Point the interface at its definition

The Interface section SHALL point to the defining type or export map. It
SHALL name a member only to state a guarantee that the signature can't
express.

**Rationale:** Generated API docs are the member reference.

---

### Leave behavior to the FRD

A blueprint SHALL NOT restate FRD behavior or list implemented features or
acceptance criteria. It SHALL NOT cite an AC id except as the reason for
a rule.

**Rationale:** Two owners of one behavior diverge, and the FRD is the copy
that tests cite. Element-to-AC traceability comes from test citations
([[ac-traceability-standards]]).

---

### Say what enforces each invariant

Each invariant SHALL end with "Enforced by:" naming a test, a check, a lint,
or "review only".

**Rationale:** Readers know which rules to trust and which to verify. "Review
only" marks a missing test.

---

### State a shared rule once

WHEN two elements share a rule, the blueprint of the element that enforces it
SHALL state the rule, and the other blueprint SHALL link to it.

**Rationale:** One rule stated in two places is information leakage in prose:
the copies diverge.

---

### Write the description for triage

The `description` property SHALL take the form "<what it owns>. Change it to
<kinds of change>."

**Rationale:** Agents choose elements from `specs list`, which shows the
description.

---

### Name owners, not neighbors

A blueprint SHALL name the owner of each responsibility it excludes. It SHALL
NOT describe interactions with sibling elements. Its References section SHALL
list only the host, Contracts, ADRs, and guides.

**Rationale:** Edges belong to the host's blueprint, and they drift when
restated ([[c4-standards]]).

---

### Name each pattern with its progenitor, constraint, and need

A blueprint SHALL name each pattern it adopts, with its progenitor, as the
universal guide ([[guide-design-principles]]) or the level guide for the
blueprint's C4 level names it. A pattern the guides don't list SHALL be named
with its own progenitor. For each pattern, the blueprint SHALL state:

- The constraint the pattern implies and what it rules out, as an EARS
  invariant that ends with "Enforced by:"
  ([[#Say what enforces each invariant]]).
- The present need it serves: an existing variation, an FRD criterion, or a
  stated extension point. An anticipated need doesn't qualify.

**Rationale:** A name alone drifts: "Decorator" is also Python syntax, and
"Repository" is confused with a data access object. The progenitor fixes the
meaning, and the constraint is what a reviewer or a check can verify. The
need keeps a pattern from being adopted for its own sake.

---

### State where the element grows

A blueprint SHALL state where its element is designed to grow and the wrong
way to extend it, or state that the element has no extension points and why.

**Rationale:** Code shows a seam's mechanism. It doesn't show that the seam
is the intended way in.

---

### Verify isolation tests always

The Verification gates SHALL list isolation tests from `test_paths` under
Verify always. They SHALL list wider suites under Verify conditionally, each
with the condition that requires it.

A System Blueprint lists its system-level suites under Verify always.

**Rationale:** Isolation tests are fast and point at the broken element. The
condition says when the slow suite is owed.

---

## Rules for Component and Module Blueprints

### Own a disjoint branch of the source tree

WHILE status is `active`, a Component or Module Blueprint SHALL declare
`source_paths` and `test_paths`. Its `source_paths` SHALL NOT overlap a
sibling's.

**Rationale:** A file maps to one blueprint, so parallel work on siblings
never touches the same files.

---

### Give each pattern instance one line

In a pattern blueprint (`kind: pattern`), each instance SHALL get one line:
where it lives and how it differs.

**Rationale:** An instance's mechanics are in its code.

---

### Keep a Component Blueprint near 250 lines

A Component Blueprint SHOULD stay near 250 lines. Past that, cut it or split
the component.

**Rationale:** A budget prompts the audit. It's advice only, because length
can't measure quality.

---

## Blueprint audit

Run this audit on an architecture blueprint before you set it to
`ready-for-review`:

1. Run `specs check <path>` and fix what it reports.
2. Classify each statement as a rule, a fact, a behavior, or an edge. Keep
   the rules. Replace each fact with a link to the code, move each behavior
   to the FRD, and leave each edge to the host's blueprint.
3. Apply every rule in each section whose heading covers the blueprint.
4. Confirm that the extension points and an "Enforced by:" note on each
   invariant exist. In a Component or Module Blueprint, confirm the Hides
   section too.
5. For each named pattern, ask what breaks if you remove it. Remove a pattern
   whose removal breaks nothing the blueprint states.

---

## References

- [[adr-032]] — blueprints state rules, not code facts
- [[adr-033]] — behavior belongs to the FRD
- [[adr-009]] — parents own the edges among their children
- [[adr-011]] — blueprints declare targeted verification gates
- [[c4-standards]] — the C4 levels, and where edges and scope statements live
- [[ac-traceability-standards]] — how tests cite acceptance criteria
- [[adr-034]] — design principles come from curated guides by C4 level
- [[guide-design-principles]] — the design questions behind these rules, and
  the guide for each C4 level

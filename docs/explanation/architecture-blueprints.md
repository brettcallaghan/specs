# How architecture blueprints work

Agents read code quickly, so a blueprint that describes the code tells them
nothing new. It only gives them a second copy to doubt.
An architecture blueprint in Specs therefore holds what the code can't say:
why a boundary sits where it does, what an element must never take on, which
rules survive every change, and where the design expects to grow.

## The problem with describing the implementation

The natural way to document a component is to describe it: its types, its
files, the features it implements, the acceptance criteria it meets. Every one
of those facts is true on the day it's written. Then the code changes and
nothing updates the blueprint, so the copy drifts.

The drift costs more than the stale sentence. A reader can't tell a guarantee
from an accident of today's code, so a rule and a leftover fact look the same.
An agent that trusts the blueprint makes a change that is locally reasonable
and quietly breaks the design. An agent that distrusts it rereads the code and
gets nothing from the blueprint at all.

So a blueprint states rules, and links the code for implementation facts
([ADR-032](../../specs/adrs/adr-032-blueprints-state-rules-not-code-facts.md)).
A path or a symbol in backticks is a name that a reader, or a tool, can
resolve against the source tree. A list of members isn't.

## Each fact has one owner

The way to keep a blueprint short and true is to give every kind of fact a
single home, and to point at that home rather than copy it:

- **The code owns implementation detail.** Types, signatures, file layout,
  and anything else that reading the source would reveal.
- **The FRD owns behavior.** What users observe, written as acceptance
  criteria
  ([ADR-033](../../specs/adrs/adr-033-behavior-belongs-to-the-frd.md)). Tests
  cite those criteria, and the citations record which element implements
  which criterion, so a blueprint never needs a list of its own.
- **The host blueprint owns edges.** How sibling elements interact is the
  parent's business. A Component Blueprint names who owns what it excludes,
  but it doesn't describe its neighbors.
- **The blueprint owns intent.** Its responsibility, the decision it hides,
  its interface guarantees, its invariants, and its extension points.

When two elements share a rule, the blueprint of the element that enforces it
states the rule, and the other links to it.

## What each level decides

The [C4 model](../../specs/guides/c4-guide.md) describes software at
nested levels, and each level makes a
different kind of decision. Each architecture blueprint records the decisions
that belong to its level.

A **System Blueprint** decides how containers are composed and how they
interact across process boundaries: integration style, which container owns
each store, how consistency holds across them, and how one failure stays
contained.

A **Container Blueprint** decides how one deployable unit's components are
composed: its internal architecture style, which way dependencies point, and
where it's designed to be extended.

A **Component Blueprint** describes the unit of change within a container:
one building block with a clear role, its own branch of the source tree, and
tests that prove it in isolation. Sibling components' source paths don't
overlap, so agents can change siblings in parallel without touching the same
files. A
**Module Blueprint** plays the same role for a unit that is shared across
containers or holds components of its own.

A **Feature Blueprint** is different from the rest. It's the design for one
change: the deltas it makes to the blueprints above, written in their
vocabulary. The Work Order that implements each delta folds it into the
owning blueprint in the same change. Once the feature merges, the Feature
Blueprint retires. The architecture blueprints are
the lasting record, and the Feature Blueprint is the diff.

## What a Component Blueprint does

A Component Blueprint earns its place by doing what an agent needs before it
touches the code:

- **It helps a planner find where a change belongs.** Its description has the
  form "\<what it owns\>. Change it to \<kinds of change\>." A planning agent
  can scan every component's description with `specs list` and pick the ones
  a feature touches.
- **It draws the boundary.** It says what the component owns, the design
  decision it hides from everything else, and what it doesn't own, with the
  owner of each. A component that hides no decision is a candidate to merge
  into a neighbor.
- **It guides design.** It names the patterns the component follows, the
  invariants every change keeps, and the extension points where it's designed
  to grow. Design then evolves through the seams the component was built
  with, not around them.
- **It defines the interface.** It points at the code that defines the
  surface, and states only what a signature can't: the guarantees callers
  rely on and the preconditions the component depends on.
- **It keeps work local and verifiable.** Its source and test paths don't
  overlap a sibling's. Its verification gates name the isolation tests to run
  on every change. They name wider suites too, each with the condition that
  requires it. An agent proves its change quickly and knows when the slow
  suite is owed.

Each invariant also says what enforces it: a test, a check, a lint, or review
only. A reader then knows which rules to trust and which to verify, and
"review only" marks a missing test.

## Thinking in patterns

A design principle or pattern is a compact way to say a lot. "Ports and
Adapters" tells an experienced reader which code is the core, which code
adapts it to the outside world, and which way the dependencies point. Agents
know these patterns too, and naming one brings that knowledge into the
conversation.

A name alone isn't enough, though. "Decorator" is also Python syntax, and
"Repository" is often confused with a data access object. A blueprint
therefore names each pattern with its progenitor, the source that fixes its
meaning, such as Cockburn for Ports and Adapters. It then states the
constraint the pattern implies as an invariant, with what enforces it. A
change then has to conform to the pattern, not merely mention it.

Each pattern also names the present need it serves: a variation that exists
today, an acceptance criterion, or a stated extension point. A pattern
adopted for a need that might arrive one day adds indirection without
paying for it.

The `spec` skill ships curated guides for this
([ADR-034](../../specs/adrs/adr-034-design-principles-guides-by-c4-level.md)):
a universal guide for principles that hold at every level, and one guide each
for systems, containers, and components. They're organized by level because
one idea takes a different form at each level. Announcing a change is
Observer within a component, Domain Events within a container, and
Event-Driven Architecture across a system. A flat list leaves the author to
work out which level a pattern belongs to, and naming it at the wrong level
records the wrong design.

The guides aren't catalogs. Each holds the principles and patterns that apply
most widely in current architectures, with priority for those that keep a
design modular and extensible. When another pattern fits better, a blueprint
names it with its own progenitor.

## Keeping blueprints true

A blueprint stays true when the work that changes the code also changes the
blueprint. These practices make that routine:

- **Feature Blueprints fold back in.** Each delta is written in the owning
  blueprint's vocabulary, and the Work Order that implements it updates that
  blueprint in the same change.
- **An audit runs before review.** The author classifies each statement as a
  rule, a fact, a behavior, or an edge, keeps the rules, and sends the rest to
  their owners. For each named pattern, the audit asks what breaks if it's
  removed, and removes the pattern when its removal breaks nothing the
  blueprint states.
- **A budget prompts the audit.** A Component Blueprint aims to stay near 250
  lines. Past that, it's usually restating something, or the component is
  really two.

## Further reading

- [Blueprint Content Standards](../../specs/standards/blueprint-content-standards.md):
  the rules and the audit.
- [Design Principles Guide](../../specs/guides/design-principles-guide.md): the
  universal guide, with links to the guide for each level.
- [Document types](../reference/document-types.md): every type, with its
  template.
- [How the document model works](document-model.md): how product and
  architecture documents fit together.

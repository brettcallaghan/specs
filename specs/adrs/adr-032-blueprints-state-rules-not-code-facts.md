---
title: "ADR-032: Architecture Blueprints State Rules, Not Code Facts"
description: Architecture blueprints state the rules that must stay true
  and link the code for everything else, and each rule has one owning
  blueprint
slug: adr-032
type: adr
status: approved
---

# ADR-032: Architecture blueprints state rules, not code facts

**Status:** approved  
**Date:** 2026-10-31

## Context

Agents write most of the code in spec-first development, and they read code
quickly. Code shows what an element does today. It doesn't show why the
boundary sits where it does, what the element must never take on, which rules
must survive every change, or where the design expects to grow. Without that
intent, an agent makes a locally reasonable change that erodes the design.

The architecture templates let a blueprint describe the implementation.
Blueprints written from them restate member lists, counts, versions, and file
layouts that an agent can read from the code in seconds. Nothing updates
those facts when the code changes, so they drift. They also cost every reader
time and bury the rules that matter. A reader can't tell a guarantee from an
accident of today's code.

A rule that two elements share has the same problem when both blueprints
state it: the two copies diverge.

## Decision

We will have architecture blueprints state rules that must stay true, not
facts the code already states. A blueprint links the code that defines a fact
by its repository-relative path or symbol instead of restating it. Each rule
has one owning blueprint: the blueprint of the element that enforces it, which
the other elements' blueprints link to. [[blueprint-content-standards]] holds
the rules that apply this decision.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Blueprints describe the implementation, and reviews keep them current | Facts sit in one place without opening the code | Facts drift between reviews<br>Nothing marks which statements are guarantees | Restated facts drift because nothing updates them when the code changes |
| Generate the descriptive sections from code | Facts stay current | Repeats what an agent reads from the code in seconds<br>Still records no intent | Spends reader time on what the code already says, and doesn't supply the missing intent |
| Blueprints state rules only, each with one owning blueprint | Statements stay true as the code changes<br>Readers can tell guarantees from incidental facts | Authors classify each statement before writing it<br>A reader opens the code for counts, members, and layouts | N/A — selected |

## Consequences

**Positive:**
- A blueprint stays true as the code changes, unless a rule changes.
- A reader can trust every statement as a guarantee, and each guarantee says
  what enforces it.
- Code references are paths and symbols that a tool can resolve against the
  source tree.

**Negative:**
- A reader who wants a count, a member list, or a file layout reads the code.
- Authors run an audit before review, classifying each statement as a rule, a
  fact, a behavior, or an edge.
- A blueprint written to the earlier templates restates code until someone
  audits it.

## References

- [[blueprint-content-standards]] — the rules and the blueprint audit
- [[adr-009]] — parents own the edges among their children
- [[adr-033]] — behavior belongs to the FRD
- [[c4-standards]] — the C4 levels these blueprints document

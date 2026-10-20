---
title: "ADR-<NNN>: <Decision Title>"
description: <One-sentence summary of the decision>
slug: adr-<nnn>
type: adr
status: draft
---

<!--
Template: Architecture Decision Record (ADR)
Source: Adapted from Michael Nygard's ADR format

Sections (complete each, then delete its instruction block)
  - Context
  - Decision
  - Alternatives Considered
  - Consequences
  - References

Purpose
  Records significant architectural and technology decisions, their context,
  alternatives considered, and consequences. ADRs create a decision log that
  explains "why we built it this way" for future maintainers.

Reader profile (apply with the tech-writer skill)
- Reader: a future maintainer, human or agent, asking "why is it this way?",
  and today's reviewer.
- Knows: the system now, not the context at the time of the decision.
- Wants: the decision, its context, the alternatives, and the consequences,
  including the costs.
- Budget: a few minutes.
- Tone: neutral and direct. State the decision as "We will..." and give the
  rejected alternatives a fair hearing. Don't argue for the decision beyond
  its reasons.

When to use
  - Choosing between two or more real technology options
  - Making decisions with consequences that are hard to reverse
  - Recording choices that will be questioned later ("why did we pick X?")
  - Establishing architectural patterns or constraints

When NOT to use
  - Obvious choices with no real alternative
  - Implementation details within a decided technology
  - Decisions that don't survive past a single sprint
  - Tactical bug fixes or minor refactorings

Instructions
  1. Number ADRs sequentially: ADR-001, ADR-002, etc. Get the next number
     from `specs next-id adr` rather than from the highest filename you
     can see — the slug is independent of the filename, so a filename
     scan can miss the number actually in use.
  2. Update the frontmatter with the actual decision details
  3. Complete each section below, following the guidance in comments
  4. Remove all HTML comment blocks after completing the document
  5. Update the adrs/ README index with a one-line entry for this ADR,
     under whatever spec root specs.toml declares (not necessarily docs/)
  6. Set status using the controlled vocabulary:
     draft | ready-for-review | approved | active | retired
  7. If superseded later, set status to "retired" and reference the
     superseding ADR

Slug format
  - The slug is the bare type prefix and number: adr-<nnn>
  - Example: adr-003
  - The slug is independent of the filename. Name the FILE
    descriptively — adr-003-local-tesseract-for-ocr.md — so humans can
    find it; the slug stays adr-003 so [[adr-003]] resolves.
  - See template-authoring-standards.md, Prefix Registry.
-->

# ADR-<NNN>: <Decision Title>

**Status:** draft | ready-for-review | approved | active | retired
(superseded by [[adr-<nnn>]])  
**Date:** <YYYY-MM-DD>

## Context

<!--
Describe the situation that forced this decision:
- What problem exists that requires a choice?
- What constraints exist (technical, business, regulatory)?
- What would happen if no decision was made?

Guidance:
  - Write 2-4 sentences
  - Provide enough context that someone unfamiliar with the project
    understands why this decision matters
  - Name specific constraints (performance, cost, security, compliance)
  - Reference related documents using [[slug]] notation if applicable

Example:
  The pipeline needs to extract text from photos of physical documents
  taken on a laptop camera. Two paths exist: run OCR locally (Tesseract)
  or call a cloud vision API. The system is self-hosted on a home network
  and processes sensitive household documents (bills, medical EOBs, tax
  notices).
-->

<Describe the situation, constraints, and why this decision is necessary>

## Decision

<!--
State the choice made as one clear statement.

Guidance:
  - Be direct: "We will use X" not "We considered using X"
  - Do NOT list options here (that goes in Alternatives Considered)
  - Specify the technology, pattern, or approach chosen
  - Keep it brief (1-3 sentences maximum)

Example:
  We will use Tesseract 5 running inside the back-end container for OCR.
  No image content will be sent to cloud OCR services.
-->

<State the decision clearly and directly>

## Alternatives Considered

<!--
List all realistic options that were evaluated, including the chosen option.

Guidance:
  - Include at least 2-3 alternatives
  - Be honest about tradeoffs (helps future readers trust this ADR)
  - The "Why rejected" column should reference specific constraints from
    the Context section
  - Mark the chosen option with "N/A — selected" in the rejection column
  - If an option was rejected due to requirements, cite the requirement
    using [[slug]] notation

Example row:
  | Google Cloud Vision | Higher accuracy, handles poor lighting |
  Images leave the machine; cost per call; requires internet |
  Privacy constraint from [[privacy-requirements]] §2 |
-->

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| <Option A> | <Benefit 1><br><Benefit 2> | <Cost 1><br><Cost 2> | <Specific reason tied to constraints> |
| <Option B> | <Benefit 1><br><Benefit 2> | <Cost 1><br><Cost 2> | <Specific reason tied to constraints> |
| <Chosen option> | <Benefit 1><br><Benefit 2> | <Cost 1><br><Cost 2> | N/A — selected |

## Consequences

<!--
Describe the positive and negative outcomes of this decision.

Guidance:
  - Be explicit about costs, not just benefits
  - Name what this decision rules out (opportunity cost)
  - Include operational implications (maintenance, monitoring, training)
  - Mention technical debt or future work this creates
  - A Negative is a *standing* cost — one the decision keeps imposing for
    as long as it stands, that a reader adopting it today would still pay.
    The effort to get here is not a Negative: "we had to migrate 30 call
    sites" or "the initial build was substantial" is implementation scope,
    not a consequence — leave it out (it belongs in the plan). Before
    flipping status to `active`, re-read every Negative and drop any that
    merely describe work already done; keep only costs that persist.
    (Alternatives Considered keeps its full historical record; Consequences
    does not.)
  - Honest ADRs get trusted; sanitized ADRs get ignored

Example:
  Positive:
  - Document images never leave the local network
  - No per-call cost; no API key to manage
  
  Negative:
  - Accuracy lower than cloud alternatives, especially in poor lighting
  - Requires image pre-processing (contrast, deskew) to reach acceptable
    confidence
-->

**Positive:**
- <Specific benefit>
- <Specific benefit>
- <Specific benefit>

**Negative:**
- <Specific cost or constraint accepted>
- <What this decision rules out>
- <Operational or maintenance burden>

## References

<!--
Optional section. Include references to:
- Related ADRs using [[adr-<nnn>]] notation
- External documentation or standards
- Benchmarks or performance data that informed the decision
- RFCs, blog posts, or research papers

Remove this section if no references are needed.
-->

- <Reference 1>
- <Reference 2>

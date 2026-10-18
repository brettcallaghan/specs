---
title: Writing Requirements with EARS
description: How to use EARS (Easy Approach to Requirements Syntax) to author
  clear, testable requirements in any document — every normative statement
  except an FRD's features and acceptance criteria
slug: guide-ears
type: guide
status: draft
---

# Writing Requirements with EARS

## Contents

- Where EARS applies — and the one surface it does not
- Why EARS — what the syntax buys over free-form prose
- Sentence Patterns — ubiquitous, event-driven, state-driven, optional
  feature, conditional, negative (plus pattern selection and complex
  requirements)
- Modal Verbs — RFC 2119 keywords and when each applies
- Writing Style Rules
- Common Mistakes
- Examples Across Requirement Types
- Requirements in Documents
- Quick Checklist — run this before calling a requirement done
- References

EARS (Easy Approach to Requirements Syntax) is the convention for authoring
requirements in this project — functional requirements, non-functional
requirements, system invariants, policies and constraints, in any document
that expresses them ([[adr-002]]).

**Acceptance criteria, and the features that hold them, are the
exception.** Those are written in Gherkin ([[adr-021]]); [[gherkin-guide]]
is how. Nothing below applies to an FRD's `Feature` blocks. Everything
below applies to every other normative statement, including the ones in an
FRD's prose and in the standards, contracts and blueprints an FRD cites.

EARS is a writing convention, not a document type. Any document may contain
EARS requirements; this guide defines how to write them.

---

## Why EARS

Free-form prose requirements hide ambiguity: implicit assumptions, missing
conditions, and vague qualifiers that different readers interpret
differently. EARS eliminates this by forcing every requirement into one of a
small set of sentence patterns with explicit triggers, conditions, and
responses.

A well-formed EARS requirement has three properties:

1. **Unambiguous** — exactly one interpretation.
2. **Testable** — you could write a failing test for it.
3. **Atomic** — one requirement per sentence.

---

## Sentence Patterns

Every requirement SHALL use one of the following patterns. Keywords (WHEN,
WHILE, WHERE, IF, THEN, SHALL, SHALL NOT) SHALL be capitalized.

| Pattern | Syntax | Use when |
|---------|--------|----------|
| Ubiquitous | The system SHALL [action] | Always true, no trigger or condition |
| Event-driven | WHEN [event], the system SHALL [action] | Triggered by something that happens |
| State-driven | WHILE [state], the system SHALL [action] | True as long as a condition holds |
| Optional feature | WHERE [feature is present], the system SHALL [action] | Applies only when a feature or capability is included |
| Conditional / unwanted behavior | IF [condition], THEN the system SHALL [action] | Response to a precondition, often an error or undesired situation |
| Negative | The system SHALL NOT [action] | Hard prohibition |

Replace "the system" with the actual subject of the requirement — a named
component, service, template, process, or role. Prefer the most specific
subject that owns the behavior.

### Pattern Selection

Ask these questions in order:

1. Is this forbidding something? → **Negative** (SHALL NOT)
2. Does it only apply when an optional feature exists? → **Optional
   feature** (WHERE)
3. Does something *happen* to trigger it? → **Event-driven** (WHEN)
4. Is it true only *during* some state? → **State-driven** (WHILE)
5. Does it depend on a precondition or handle an unwanted situation? →
   **Conditional** (IF/THEN)
6. None of the above? → **Ubiquitous** (plain SHALL)

### Complex Requirements

Patterns MAY be combined when a requirement genuinely needs multiple
clauses. Keep clause order: WHERE, WHILE, WHEN, IF/THEN, then the SHALL
clause.

```markdown
WHILE ingestion is paused, WHEN a new file arrives, the system SHALL queue
the file without processing it.
```

Use combinations sparingly. If a sentence needs more than two clauses,
consider whether it is actually two requirements.

---

## Modal Verbs

EARS requirements SHALL use RFC 2119 modal verbs to express requirement
strength:

| Keyword | Meaning |
|---------|---------|
| SHALL / SHALL NOT | Absolute requirement or prohibition |
| SHOULD / SHOULD NOT | Strong recommendation; deviation needs justification |
| MAY | Truly optional; implementer's choice |

- Default to SHALL. A requirement that cannot be stated with SHALL is
  usually a preference, not a requirement.
- Use SHOULD only when a legitimate reason to deviate can exist. If you
  cannot imagine one, use SHALL.
- Modal verbs SHALL be capitalized, and SHALL be reserved for actual
  requirements — never use them casually in surrounding prose.

---

## Writing Style Rules

**One requirement per sentence.** Never join two SHALLs with "and." Split
them so each can be tested, traced, and disputed independently.

```markdown
❌ WHEN a user registers, the system SHALL send a confirmation email and
   SHALL create an audit log entry.

✅ WHEN a user registers, the system SHALL send a confirmation email.
✅ WHEN a user registers, the system SHALL create an audit log entry.
```

**Be specific enough to test.** Every requirement should imply a concrete
pass/fail check. Quantify where possible.

```markdown
❌ The system SHALL respond quickly.
✅ The system SHALL respond to search queries within 500 ms at the 95th
   percentile.
```

**Ban vague qualifiers.** Avoid "appropriate," "reasonable," "adequate,"
"user-friendly," "as needed," "etc." If you cannot define it, it is not
testable, and it is not a requirement yet.

**Use active voice with an explicit subject.** "The scheduler SHALL retry
failed jobs" — not "failed jobs SHALL be retried" (retried by what?).

**Name events and states precisely.** "WHEN the upload completes" is
testable; "WHEN the upload is done or whatever finishes first" is not.
Define terms in a glossary or inline if they could be misread.

**State the trigger, not the implementation.** Requirements say *what* and
*under which conditions*, never *how*. "The system SHALL persist audit
events durably" — not "the system SHALL write audit events to PostgreSQL
with fsync enabled" (unless the technology itself is the requirement).

---

## Common Mistakes

| Mistake | Problem | Fix |
|---------|---------|-----|
| Compound requirements ("...and...") | Cannot test or trace independently | Split into one sentence per SHALL |
| Missing trigger ("The system SHALL validate input") | Validate when? On every call? | Add WHEN/WHILE/IF clause if behavior is not truly ubiquitous |
| Vague qualifiers ("appropriate error message") | Untestable | Define the exact behavior or reference where it is defined |
| Escaping into prose mid-list | Ambiguity creeps back in | Every normative statement uses an EARS pattern |
| Negation buried in a positive pattern ("WHEN X, the system SHALL avoid Y") | Weakens the prohibition | Use SHALL NOT: "WHEN X, the system SHALL NOT Y" |
| Passive voice ("data SHALL be encrypted") | Actor is unspecified | Name the subject that performs the action |
| SHOULD used to dodge a decision | Requirement strength left ambiguous | Decide: SHALL if mandatory, MAY if optional, SHOULD only with a known deviation case |

---

## Examples Across Requirement Types

**Functional requirement:**

```markdown
WHEN a user submits the registration form, the system SHALL validate all
required fields before creating the account.

IF validation fails, THEN the system SHALL display a field-level error
message for each invalid field.
```

**Non-functional requirement:**

```markdown
The API SHALL respond to read requests within 200 ms at the 95th
percentile under a load of 100 concurrent users.

WHILE operating in degraded mode, the system SHALL continue serving
cached read requests.
```

**Policy / invariant:**

```markdown
The system SHALL NOT transmit raw image data to any external service.

WHEN an image is captured, the system SHALL assign it a UUID-based
filename with no user-identifiable strings in the path.

IF the operator has configured a retention period, THEN the system SHALL
purge images older than that period during the nightly maintenance window.
```

**Optional feature:**

```markdown
WHERE single sign-on is enabled, the system SHALL delegate authentication
to the configured identity provider.
```

**Process / authoring requirement:**

```markdown
WHEN a template is used to author a new document, the author SHALL remove
HTML comment blocks after completing each section.
```

---

## Requirements in Documents

EARS governs the sentences, not the document. Whatever the document type —
a requirements document, a standard, a specification, a template — the
normative statements within it follow this convention, with an FRD's
`Feature` blocks excepted as above:

- Group related requirements under descriptive section headings.
- Surrounding prose (context, rationale, examples) is welcome and SHOULD
  accompany requirements, but SHALL NOT itself carry normative weight.
  If a statement is binding, write it as an EARS sentence.
- A **Rationale** note after a requirement or group of requirements is the
  preferred place to explain *why* without diluting the requirement itself.
- Cross-reference rather than restate requirements that live in another
  document (see [[template-authoring-standards]] for reference syntax).

---

## Quick Checklist

Before committing a requirement, verify:

- [ ] It uses exactly one EARS pattern (or a justified combination).
- [ ] Keywords and modal verbs are capitalized.
- [ ] It contains exactly one SHALL / SHALL NOT / SHOULD / MAY.
- [ ] The subject performing the action is explicit.
- [ ] The trigger, state, or condition is precise and observable.
- [ ] A failing test could be written for it.
- [ ] It contains no vague qualifiers.
- [ ] It states *what*, not *how*.

---

## References

- [[adr-002]] — the decision this guide implements, and the scope it keeps
- [[gherkin-guide]] — the notation for features and acceptance criteria, the
  one surface this guide does not cover
- [[template-authoring-standards]] — Frontmatter, references, and modal verb
  standards
- Mavin, A., et al. (2009). "Easy Approach to Requirements Syntax (EARS)"
- [RFC 2119](https://www.ietf.org/rfc/rfc2119.txt) — Modal verb definitions
- [EARS on Wikipedia](https://en.wikipedia.org/wiki/Easy_Approach_to_Requirements_Syntax)

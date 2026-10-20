# Reader profiles

Default reader profiles by document type. Start from the one that matches,
then adjust it for the project and the document.

Contents: project docs (README, release notes, Issue); Diataxis (tutorial,
how-to guide, explanation, reference); specs (FRD, ADR, Standards,
architecture blueprints, Feature Blueprint, Contract).

## Project docs

**README**
- Reader: a developer seeing the repository for the first time.
- Knows: the problem space, not this project. Arrives from a link, a search, or a dependency list.
- Wants: what this is, whether it fits, and how to run it.
- Budget: about a minute before deciding to install it or leave.
- Tone: brisk and confident. Evaluating and mildly skeptical, so show, don't claim.

**Release notes**
- Reader: an existing user deciding whether and when to upgrade.
- Knows: the product, not the diff.
- Wants: what changed for them and what they must do.
- Budget: one scan.
- Tone: direct and factual. Breaking changes and required actions first, with no apology or spin.

**Issue** (a bug or a proposed change)
- Reader: the person deciding whether to file or accept it, then whoever fixes it.
- Knows: the code or the skill, not your setup or your use case.
- Wants: what happened or what's missing, the evidence, and the one change proposed.
- Budget: a minute or two to triage.
- Tone: neutral and specific. Describe what happened, not how frustrating it was. Persuade with evidence, not enthusiasm. No blame.

## Diataxis

**Tutorial**
- Reader: a newcomer learning the project.
- Knows: little or nothing about this project.
- Wants: to build something that works and understand it.
- Budget: 15 to 60 minutes.
- Tone: patient. Confirm what the reader built at each milestone. Never imply a step is trivial.

**How-to guide**
- Reader: a user who knows the basics and has one task, often mid-task.
- Knows: the basics.
- Wants: the steps to finish the job, including the fix when something broke.
- Budget: minutes.
- Tone: efficient, with minimal preamble. When the page fixes a problem, name the symptom, then fix it calmly.

**Explanation**
- Reader: someone who uses the feature and wants the reasoning.
- Knows: how to use it.
- Wants: the problem, the design, and the tradeoffs.
- Budget: one unhurried read.
- Tone: conversational and thoughtful. Longer, more varied sentences fit here.

**Reference**
- Reader: someone looking up one fact.
- Knows: what they're looking for.
- Wants: the exact value, signature, or option.
- Budget: seconds.
- Tone: neutral and precise, in parallel entries. Personality stays out.

## Specs

Specs are the written source of truth for everyone who works on the system,
coding agents included. Write for the human reviewer, and keep the text exact
enough that an agent can act on it without guessing.

**FRD**
- Reader: a reviewer approving what will be built, then the authors of the Feature Blueprint and tests that trace back to it.
- Knows: the product, not this feature.
- Wants: why the feature matters (the user story) and what it must do (the acceptance criteria).
- Budget: one focused read.
- Tone: precise and candid. State assumptions and open questions outright. Say nothing about how the system is built.

**ADR**
- Reader: a future maintainer, human or agent, asking "why is it this way?", and today's reviewer.
- Knows: the system now, not the context at the time of the decision.
- Wants: the decision, its context, the alternatives, and the consequences, including the costs.
- Budget: a few minutes.
- Tone: neutral and direct. State the decision as "We will…" and give the rejected alternatives a fair hearing. Don't argue for the decision beyond its reasons.

**Standards**
- Reader: an engineer or agent checking whether their work complies, often from a review comment or a link in a spec.
- Knows: their own work, not the rule.
- Wants: the rule, its scope, and what compliant looks like.
- Budget: a minute per rule.
- Tone: firm and unambiguous, not scolding. Give the reason for any rule that isn't obviously necessary, so readers can apply it to cases it doesn't name.

**System, Container, Component, and Module Blueprints**
- Reader: an engineer or agent about to change or build on the element, often while writing a Feature Blueprint, and a newcomer finding their way in. The System Blueprint is the entry point.
- Knows: the surrounding system, not this element. The deeper the level, the more code-level context the reader brings.
- Wants: the structure, responsibilities, interfaces, and the change guidance that feature work must respect.
- Budget: one focused read per level.
- Tone: matter-of-fact, describing the element as it is. At the system level, favor the big picture; at the component and module levels, favor exact names and contracts. Decisions and their alternatives belong in ADRs.

**Feature Blueprint**
- Reader: the human reviewer approving the design, then the planning agent that derives the implementation plan.
- Knows: the FRD and the architecture blueprints it changes.
- Wants: what changes, element by element, and how each acceptance criterion is covered.
- Budget: one focused read.
- Tone: precise and diff-shaped. Reference decisions, contracts, and steady-state architecture; don't restate them.

**Contract**
- Reader: a consumer or provider building against the interface on its own schedule.
- Knows: the machine-readable spec's shape, not its semantics.
- Wants: the guarantees: what each operation means, its errors, ordering, and limits.
- Budget: looked up repeatedly, one guarantee at a time.
- Tone: exact and neutral, like reference. State guarantees, not rationale.

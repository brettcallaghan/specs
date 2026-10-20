# Reader profiles

Default reader profiles for docs outside the spec skill. Start from the one
that matches, then adjust it for the project and the document. Each spec
document type carries its own profile in its template's header comment.

Contents: project docs (README, release notes); Diataxis (tutorial, how-to
guide, explanation, reference).

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

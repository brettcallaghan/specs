# Reader profiles

Default reader profiles for docs outside the spec skill. Start from the one
that matches, then adjust it for the project and the document. Each spec
document type carries its own profile in its template's header comment.

Contents: project docs (README, release notes); Diataxis (tutorial, how-to
guide, explanation, reference); agents (skills and other docs written only
for agents).

## Project docs

**README**
- Reader: a developer seeing the repository for the first time, evaluating it and mildly skeptical.
- Knows: the problem space, not this project. Arrives from a link, a search, or a dependency list.
- Wants: what this is, whether it fits, and how to run it.
- Budget: about a minute before deciding to install it or leave.
- Tone: brisk and confident. Show, don't claim.

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

## Agents

**Docs written only for agents**: skills and their references, `AGENTS.md` and
`CLAUDE.md`, subagent prompts, and the instruction comments in templates.
- Reader: an LLM agent that loads the text into its context, often mid-task, and follows it literally.
- Knows: public-domain knowledge, languages, and common tools at expert level. Knows nothing about this project except what the text and its links say.
- Wants: what to do, when to do it, and the constraints that change what it does.
- Budget: tokens. Every sentence takes context away from the task.
- Tone: terse and imperative. No persuasion, encouragement, or narrative.

Write for this reader as follows:
- Cut what a capable agent already knows: explanations of public concepts, definitions of common terms, and how standard tools work. Keep what is specific to the project or easy to get wrong.
- Cut history. State the rule and what breaks without it, not how the rule came about.
- Use one term for each concept, and never swap in a synonym. Define a project term once. Name files, commands, flags, and fields exactly.
- State obligations with an imperative or an RFC 2119 keyword, not with "you might want to" or "consider."
- Give a reason only where the agent would otherwise misapply the rule or remove it as unnecessary. Name the plausible wrong move when there is one.
- Show the shape with an example when a paragraph would describe it.
- Cut whole sentences rather than compressing them into fragments. A fragment saves few tokens and loses meaning.

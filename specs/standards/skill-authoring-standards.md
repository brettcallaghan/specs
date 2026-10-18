---
title: Skill Authoring Standards
description: Requirements for authoring agent skills in this repository —
  discovery metadata, progressive disclosure, bundled references and
  scripts, workflow design, content discipline, and validation
slug: skill-authoring-standards
type: standards
status: draft
---

# Skill Authoring Standards

These requirements govern every agent skill authored in this repository —
the skills shipped in the `specs` plugin and the repo-local
skills under `.claude/skills/`. They apply to skill authors and to any
agent modifying a skill on their behalf.

This document is the skill-side counterpart to
`plugin/skills/spec/references/template-authoring-standards.md`, which
governs document templates. Where that document asks "how is a template
written," this one asks "how is a skill written so an agent can find it,
load only what it needs, and follow it without improvising."

Authority is twofold. The upstream authority is Anthropic's published
skill-authoring best practices (see References). The rationales below draw
on that guidance **and** on defects observed in this repository's own
skills — each rationale marked *Observed* records a real failure, not a
hypothetical one.

---

## Discovery Metadata

### Frontmatter Validity

A SKILL.md file SHALL begin with a YAML frontmatter block that parses as a
mapping.

**Rationale:** *Observed.* A reflow wrapped a `description` value across
lines without indenting the continuations. The frontmatter stopped parsing,
the skill loaded with no description, and it would have effectively never
triggered. Nothing in review caught it — content diffing, line-length
checks, and fence-balance checks all passed — because none of them parsed
the YAML. The failure is invisible until the skill is loaded for real.

---

WHEN a frontmatter scalar wraps across lines, the author SHALL indent every
continuation line.

**Rationale:** An unindented continuation terminates the scalar and is
parsed as a new key, which fails. This is the specific mechanism behind the
defect above and is worth stating separately because reflow tooling and
manual editing both produce it.

---

A SKILL.md file SHALL declare a non-empty `name` and a non-empty
`description`.

**Rationale:** Both are contractual. Only these two fields are pre-loaded
into the system prompt at startup; everything else is read on demand. A
skill missing either cannot be selected.

---

### Naming

A skill's `name` SHALL match its containing directory.

**Rationale:** A mismatch makes the skill unaddressable by the name agents
and humans see, and silently breaks any cross-skill reference to it.

---

A skill's `name` SHALL use the verb-object form, EXCEPT where it belongs to
a named subsystem, in which case it SHALL carry that subsystem's prefix.

**Rationale:** Verb-object (`groom-beads`, `triage-issue`, `retro-spec` as
"retro the spec", `spec` as "spec out the thing") makes a skill's activity
legible from its name alone. A consistent pattern keeps the library coherent
as it grows, and is recorded so future skills do not each pick a different
convention.

**The subsystem exception, and its own rule.** A skill is prefixed WHEN it
only makes sense inside that subsystem — `sprint-scope`, `sprint-plan`,
`sprint-review`, `sprint-retro`, `sprint-recovery`. A skill that is useful
outside it keeps verb-object even when a sprint invokes it — `groom-beads` is
useful with no sprint in sight.

The trade is deliberate. **A skill is selected by its description, not its
name**, so the prefix costs nothing at dispatch and buys grouping in a listing
that now holds fourteen skills. Where a subsystem has one entry point, that
entry point does NOT take the prefix — it is the way in, not a member.

---

A skill's `name` SHALL NOT name a mood, a metaphor, or a project codename
in place of the activity it performs.

**Rationale:** The name is the first discovery signal. A name that does not
state the activity forces the description to carry the entire load, wasting
the strongest available cue.

---

### Descriptions

A skill's `description` SHALL state both what the skill does and when to
use it.

**Rationale:** Selection happens against the description alone, across
potentially a hundred skills. "What" without "when" leaves the agent
unable to decide whether this is the right skill for the task in front of
it.

---

A skill's `description` SHALL be written in the third person.

**Rationale:** The description is injected into the system prompt.
First-person or second-person phrasing ("I can help you…", "You can use
this to…") mixes points of view in that context and degrades selection.

---

A skill's `description` SHOULD include the concrete terms, file names, and
phrasings a user would actually say when the skill applies.

**Rationale:** Trigger terms are what selection matches on. `groom-beads`
naming `bd ready`, `.beads/`, and the literal phrase "groom these" is worth
more than an abstract summary of its capabilities.

---

A skill's `description` SHOULD NOT restate mechanism that the body already
covers.

**Rationale:** *Observed.* One description reached 94% of the 1,024-character
ceiling by appending an inventory of what the skill does internally. Trimming
the mechanism while keeping every trigger term cut it by 40% with no loss of
discoverability, and restored headroom for future triggers.

---

## Progressive Disclosure

### Body Budget

A SKILL.md body SHOULD NOT exceed roughly 40,000 characters.

**Rationale:** The published guidance states this budget as 500 lines,
which is a proxy for context cost at conventional line lengths. Stating it
in characters makes it immune to reflow: *Observed*, reflowing one skill's
prose to 80 columns raised its line count by 75% while changing zero tokens,
so the line metric reported a regression where the actual context cost had
fallen 28%. Measure the thing that costs, not its formatting.

---

WHEN a skill's body approaches the budget, the author SHALL extract content
into bundled reference files rather than compressing prose.

**Rationale:** Compression loses the specificity that makes guidance
followable. Extraction costs nothing until the file is read.

---

The author SHALL extract content whose need is conditional on a branch the
run may not take.

**Rationale:** *Observed.* One skill carried four dispatch prompt
templates, a crash-recovery procedure, and a first-run setup guide in its
body. All three are needed at one specific moment or not at all — at
dispatch, after a crash, and only in a project lacking configuration.
Extracting them cut the body 28% with no content removed.

---

### Reference Structure

A reference file longer than 100 lines SHALL carry a table of contents
immediately after its title.

**Rationale:** Files reached through a reference are often previewed rather
than read in full. Without a contents list, a partial read of a long
document returns whatever happens to be at the top and silently omits the
rest.

---

WHERE a reference file is a document template, its table of contents SHALL
be placed inside the template's instruction comment block.

**Rationale:** A template's body is copied into every generated document.
A contents section in the body would be inherited by every instance; the
instruction block is deleted at authoring time, so the guidance reaches the
author and not the artifact.

---

A table of contents SHALL name what each section answers, not merely repeat
the heading text.

**Rationale:** *Observed.* A generated contents list for a readiness
checklist rendered as the single entry "The checks", which told a reader
nothing. Naming all eight checks made the same list useful at a glance.

---

Every bundled file SHALL be reachable in one hop from its SKILL.md.

**Rationale:** A file reached at the second hop tends to be previewed
rather than read, so its content arrives incomplete. WHERE a template
depends on a guide, the SKILL.md SHALL name both as a bundle so they load
together.

---

## Bundled Files

### Path Resolution

A backticked path SHALL be written relative to the file that contains it.

**Rationale:** *Observed.* Thirty-nine citations inside a `references/`
directory carried a `references/` prefix — correct from the SKILL.md, but
resolving to `references/references/…` from where they actually appeared.
One checklist used both spellings two lines apart for two files in its own
directory. An agent resolves a path from the file it is reading; a path
correct from somewhere else costs a search at best and a skipped load at
worst.

---

A skill SHALL NOT reference a file outside its own directory.

**Rationale:** A skill's bundle is what travels with it. A `../` traversal
into a sibling skill silently depends on co-installation, dangles when the
skill is copied alone, and is undiscoverable from the referencing skill's
own tree.

---

WHERE two skills must share a reference document, the shared file SHALL be
authored once and vendored into each consumer by a generator.

**Rationale:** The two alternatives both fail. Hand-copying drifts —
*Observed*, two copies of one template diverged by 74 lines, including a
superseded term and a missing rule. Cross-skill paths break the bundle. A
generated copy gives single authorship and self-contained skills at once.

---

A vendored copy SHALL be mechanically verifiable against its source.

**Rationale:** Single authorship only holds if divergence is detectable.
Without a check, a vendored copy is just a hand copy with extra steps.

---

A generated file SHALL carry a header naming its source and how to
regenerate it.

**Rationale:** Without the header the copy looks editable, and an edit to a
generated file is lost on the next regeneration.

---

### Scripts

WHERE an operation is deterministic, fragile, or duplicated across
prompts, the skill SHALL provide a script rather than prose instructions.

**Rationale:** *Observed.* One skill taught agents to hand-type literal
single-quoted JSON to avoid a permission heuristic that fires on
runtime-substituted operands. That is a workaround for a missing script,
re-derived on every run and easy to get subtly wrong. Moving the operation
into one allowlisted script removed both the hazard and the instruction.

---

A bundled script SHALL handle its own error conditions.

**Rationale:** An agent recovering from a script's unhandled failure is
improvising at exactly the moment the script existed to make deterministic.

---

A bundled script SHALL NOT defer an error condition to the calling agent.

**Rationale:** Deferring is the same defect stated from the other side, and
worth prohibiting explicitly: a script that returns a raw error for the
agent to interpret has handed back the determinism it was written to
provide.

---

A bundled script SHALL exit non-zero when a check it performs does not
pass.

**Rationale:** *Observed.* A health-check script quoted a multi-word
subcommand, so the check never ran and reported itself "skipped" — while
the script still exited zero and reported no hard problems. A gate that
cannot fail is worse than no gate: it launders an unverified condition as a
verified one.

---

IF a script cannot run a check it was asked to perform, THEN it SHALL
report that check as unverified rather than as passed.

**Rationale:** Same defect, generalized. "No problems found among the
checks that ran" and "no problems" are different claims, and only one of
them is honest when a check was unavailable.

---

A bundled script's numeric or enumerated constants SHALL be justified in
place.

**Rationale:** An unexplained constant cannot be safely changed by anyone
who did not write it, including the agent that must adapt it.

---

The skill SHALL state whether a bundled file is to be executed or read.

**Rationale:** The two have different costs and different failure modes.
Executing a reference wastes a turn; reading a script into context wastes
tokens and invites reimplementation.

---

## Workflow Design

A skill whose workflow spans more than three phases SHALL provide a
checklist the agent can copy into its response.

**Rationale:** Long autonomous runs drop steps. A copied checklist is the
one artifact that survives the run and makes a skipped gate visible.

---

WHERE the project prohibits task-tracking tools, the skill SHALL state that
a copied checklist is not such a tool.

**Rationale:** *Observed.* This repository bans `TodoWrite` and
`TaskCreate` for tracking. Without an explicit carve-out, an agent
reconciling that rule against a checklist instruction is likely to drop the
checklist — losing the visibility for a rule that was never about markdown.

---

WHERE a skill prescribes a validator, it SHALL instruct the agent to fix
what the validator reports and re-run it.

**Rationale:** A validator whose output is advisory is decoration. The loop
is what converts it into a gate.

---

WHERE a skill prescribes a validator, it SHALL NOT permit proceeding while
that validator fails.

**Rationale:** *Observed.* One skill closed this loop for its file-emit
path ("iterate until it reports 0 errors") but not for its primary path two
paragraphs later, which merely ran the checks and summarized. Same phase,
two rigours, and only one of them gates.

---

A validation loop SHALL declare a bound and an action on exhaustion.

**Rationale:** An unbounded retry in an autonomous run burns a session
without surfacing the blocker. Two cycles then a defined stop is this
repository's established pattern.

---

A skill SHALL match its level of prescription to the cost of deviation.

**Rationale:** Where a wrong step is unrecoverable — an ordering that
produces a second commit, a close ceremony that loses a record — exact
steps are warranted. Where many routes succeed, prescription forecloses
better ones the author did not anticipate.

---

WHERE a skill prescribes an exact sequence, it SHOULD state the intent the
sequence serves.

**Rationale:** A capable agent meeting a situation the author did not
anticipate can satisfy a stated intent; it can only follow or abandon an
unexplained instruction. "The push is the gate" survives circumstances that
"run these four commands in order" does not.

---

WHERE a skill's workflow ends in a report to a human, that step SHALL
require every factual claim to cite the artifact it derives from — a
finding, a `file:line`, or a command and its output — or be marked as the
reporter's inference.

A skill SHALL NOT permit a report to state that a verification was
performed unless it was performed.

**Rationale:** *Observed.* Skills in this repository gate their artifacts
hard and their reports not at all: `retro-skill` required every finding to
carry a verbatim quote or a repro, and produced 17 that survived
spot-checking — while the prose summarising them cleared nothing. A
fabricated claim entered there, asserted with a verification that was never
run, and reached a pull-request description before being retracted. The
same asymmetry is present in every sibling's final step. Citation is the
cheapest possible gate for that layer, and unlike "be careful" it is
mechanically checkable: a sentence either names its source or it does not.

---

## Content Discipline

A skill SHALL use one term for one concept throughout.

**Rationale:** *Observed.* One routing vocabulary was declared in four
files and contradicted in a fifth — including the tier ordering, where two
skills disagreed about which model was most capable. A backlog groomed per
one skill's own worked example failed the other's validator.

---

A skill SHALL NOT state a count of items it also enumerates.

**Rationale:** *Observed.* A skill opened with "Ten specification document
types are currently defined" directly above the authoritative table. The
count rots the moment a type is added, and this repository's own template
standards already forbid volatile facts.

---

A skill SHALL NOT contain changelog entries, superseded values, or incident
provenance.

**Rationale:** Instructions are read by an agent deciding what to do now.
Text arguing with a previous version of itself costs tokens to read and
implies a history the reader does not have.

---

WHERE historical context is genuinely needed, it SHALL be confined to a
collapsed "old patterns" section.

**Rationale:** Keeps the deprecation available without diluting the current
instruction.

---

A skill SHALL NOT describe the output of a command it also instructs the
reader to run. It SHALL instruct the reader to run the command and read what
it prints.

**Rationale:** the description is a second copy of the command's behaviour,
kept in agreement with it by nobody. It competes for the body budget with
instructions that cannot be obtained any other way, and it goes stale on the
next change to the command — silently, because no check compares prose against
a program. *Observed.* One reference paragraph opened "the command owns what it
measures and how; this document does not restate it" and then restated it, in
the same paragraph, as a nine-item list of what the command reports. Another
asserted that a command "prints the invocation behind every fact"; a later
change moved the provenance behind a flag and the sentence became false with
nothing to catch it.

---

The test is whether a reader who ran the command and read its output would
still need the sentence. WHERE the answer is yes, the sentence is a RULE about
the output rather than a description of it, and it stays.

**Rationale:** the distinction is the whole of the judgement, and it is not
obvious from the sentence's grammar. Four kinds pass the test and are kept:
what to do with the output ("route on the exit status"); what one invocation
cannot show ("a status of 1 means the key is declared and unwritten, 3 that it
is not declared at all" — the reader saw one of the three); an affordance the
output does not advertise ("`--full-result` is the way back to the
provenance"); and a meaning the output states without explaining ("a dispatch
has drained when it holds nothing ready AND nothing in progress"). What fails
the test is the inventory: the list of fields, the enumeration of keys, the
table of every exit code beside a command whose own output names the one it
took.

---

A skill SHALL NOT assert what a command, tool or API does unless that
behaviour was measured, and the assertion SHALL carry the measurement — the
invocation and the observed output — with it.

**Rationale:** *Observed.* A reference file asserted that `bd`'s
`--set-metadata` "stores every value as a string." Measured on bd 1.2.2,
`num=5` stores a number, `b=true` a boolean, `f=1.5` a number and `n=null`
a null; only arrays, objects and non-numeric text are stored as strings. The
false universal was installed by the same change that removed a *different*
false claim about the neighbouring flag — the evidence standard was applied
to the sentence being deleted and not to the sentence being installed. A
reader can tell the two apart only if the measurement travels with the claim,
because prose does not record whether anybody ran the command.

---

WHERE a behaviour cannot be measured cheaply, a skill SHALL mark the claim
unverified and state what would settle it.

**Rationale:** An unverified claim a reader can recognise as unverified is
useful. One that reads as settled and is not is worse than absent: it is
acted on with the confidence of the sentence rather than the confidence of
the evidence.

---

WHEN an asserted tool behaviour is found false, the author SHALL sweep every
surface that carries a copy of it before treating the correction as done, and
SHALL record any copy that could not be reached.

**Rationale:** *Observed.* A tool claim retracted in conversation had already
been copied into an epic comment and into a durable trigger prompt, where a
resuming session would have received it as a hard constraint with no way to
learn of the retraction; it was removed only because a report happened to
catch it. Correcting the document in front of you leaves every other copy
authoritative, and the copies with the longest reach — prompts armed to fire
later — are the ones no check and no reader ever revisits. Search for the
claim's distinctive phrasing *before* editing the source; afterwards the
search term no longer exists.

---

A skill SHALL NOT hard-code an inventory of files that live outside it.

**Rationale:** *Observed.* One skill named three specific documents by
number and described their heading convention. Those documents are
repository content, not part of the skill; the reference was both stale and
outside the skill's ownership. Describe the shape to search for, not the
instances.

---

## Validation

A skill SHALL have at least three evaluations before it is relied upon.

**Rationale:** Evaluations are the only evidence that a skill solves a real
problem rather than a supposed one. Authoring them first also surfaces the
gaps worth writing about.

**Status in this repository:** not yet met. No skill here has evaluations.
This requirement is recorded as the bar, and the gap is open work.

---

A skill SHALL be tested against every model tier it dispatches work to.

**Rationale:** A skill is an addition to a model, so its sufficiency is
model-dependent. This repository's skills deliberately route work to the
cheapest capable tier, which means the least capable model in the lineup
receives the longest procedural prompts — precisely where under-specified
guidance fails first.

**Status in this repository:** not yet met.

---

Every invariant a skill asserts about itself SHALL be enforced by an
automated check.

**Rationale:** *Observed, repeatedly.* Eight invariants in this repository
were enforced only by authorial memory; at least three had silently broken
before anyone looked. A framework whose central claim is that quality comes
from gates rather than diligence cannot exempt itself.

---

WHEN a skill is changed in a way that alters its normative content, the
author SHALL increment its `version`.

**Rationale:** Retrospective issues cite the version they exercised,
and triage answers "is this already fixed?" from that citation. Without the
stamp the question requires diff archaeology.

---

## References

- `plugin/skills/spec/references/template-authoring-standards.md` — the
  counterpart standards for document templates: frontmatter, slugs,
  instruction blocks, versioning
- [[adr-013]] — terminology decision for the atomic unit of executable work
- [Anthropic — Skill authoring best practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices)
  — the upstream authority these requirements implement and extend

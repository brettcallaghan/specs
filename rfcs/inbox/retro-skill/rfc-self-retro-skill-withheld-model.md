---
title: "RFC: The readiness gate has one item a harness policy can forbid,
  and no defined way to satisfy it"
description: A run under a policy that forbids writing the model
  identifier into repository artifacts could not pass the readiness
  gate, and had to improvise a placeholder, an unscripted Provenance
  paragraph, and a held status.
slug: rfc-self-retro-skill-withheld-model
type: rfc
status: draft
updated_at: 2026-08-01T00:00:00Z
origin: self
skill: retro-skill
skill_version: "0.3"
model: "<withheld by session policy — see Provenance>"
effort: high
runs_observed: 1
severity: blocker
plugin_version: "0.12.0"
related: []
---

# The readiness gate has one item a harness policy can forbid

## Summary

`retro-skill` was run to critique the `strata` skill after a session
that exercised it. The retro itself went cleanly — the three finding
shapes, the bar, and the citation discipline in step 7 all did their
job, and the last of those caught three wrong counts in an artifact
already pushed. One item stopped it: the readiness gate requires an
exact model identifier in frontmatter, and this session runs under a
policy forbidding that identifier in any artifact pushed to a
repository. The gate has no satisfiable state under that constraint and
no rule for it, so the run improvised three separate things.

This RFC is its own evidence: the `model` field above carries the same
placeholder, for the same reason, and this RFC is held at `draft` for
the same reason.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code on the web, CLI v2.1.42 |
| Task the skill was applied to | Authoring a retro RFC against the `strata` skill (v1.0) following a run that modelled 30 types and drove three transpiler targets |
| References / scripts / templates invoked | `plugin/skills/retro-skill/SKILL.md` v0.3; `references/skill-rfc-template.md` v0.2 |
| Artifact produced | `rfcs/inbox/strata/rfc-self-strata-compile-gate.md` |
| Run dates | 2026-08-01 |

**On the `model` frontmatter field.** Withheld for the reason this RFC
proposes a rule about. Recording it as a marked placeholder rather than
omitting it is the improvisation described in finding 1, not a
recommended practice — P1 exists to make it one.

## Evidence

1. **[gap · blocker]** Three passages require an exact model identifier
   and none admits a case where it cannot be given.

   `plugin/skills/retro-skill/references/skill-rfc-template.md:70-71`,
   in the readiness gate — "Frontmatter sets model + effort (and
   Provenance names the harness) — a skill defect is model-relative".

   `plugin/skills/retro-skill/SKILL.md:107-108` — "**model** and
   **effort** — exact model id and reasoning effort (e.g.
   `claude-opus-4-8[1m]`, effort high). Not \"Claude\"."

   `SKILL.md:241`, in the anti-pattern table — "Omitting the model /
   effort from Provenance | A model-specific defect cannot be
   reproduced or scoped; the fix may be wrong for another model".

   This session operates under a harness policy that forbids writing
   its model identifier into commit messages, PR bodies, code comments,
   or any other artifact pushed to a repository. Every route the skill
   offers is therefore closed: the exact id cannot be written, and
   omission is named as an anti-pattern.

   Having no rule to follow, the run invented three things, none of
   them scripted anywhere in the skill or template:

   - a placeholder value, `model: "<withheld by session policy — see
     Provenance>"`, chosen so the field is present and visibly
     unfilled rather than absent;
   - an explanatory paragraph appended to the Provenance section,
     which the template describes as holding a fixed set of
     human-context fields and does not invite prose;
   - holding `status: draft` rather than `ready-for-review`, on the
     reasoning that a gate item was genuinely unmet.

   The escape hatch at `SKILL.md:193` — "Fix what it reports and
   re-run; do not proceed while it fails. After two full failed cycles,
   stop and surface the blocker rather than continuing to retry" —
   does not cover this. It is written for a gate reporting defects the
   author can fix; here the item is structurally unsatisfiable and
   re-running could never change the outcome. Following it literally
   means two pointless cycles before surfacing.

   The cost is not the improvisation itself but what it does to the
   pipeline: a retro that found a genuine blocker in another skill
   arrives at triage as `draft`, which is the status meaning "not yet
   ready to be read", for a reason that has nothing to do with the
   quality of its findings.

2. **[gap · polish]** `skill-rfc-template.md:19` — `plugin_version:
   "<x.y.z>"              # or "n/a (standalone install)"`.

   The critiqued skill in that run, `strata`, lives in
   `.claude/skills/` rather than in `plugin/skills/`, so it ships with
   no plugin version. The template offers one not-applicable spelling,
   `"n/a (standalone install)"`, which does not describe a repo-local
   skill. The repository's committed RFCs use a second spelling —
   `grep -h "plugin_version" rfcs/*.md | sort -u` returns both
   `"n/a (repo-local skill)"` and four `x.y.z` values.

   The run resolved it by grepping the existing RFCs and matching the
   in-use spelling, rather than from any instruction. A label whose
   values are decided by grep is a label that will fragment.

Recurrence: one run. No claim here is stronger than that.

## Proposals

### P1 — Define the withheld-model case and let the gate pass [blocker]

**Problem.** The gate item at `skill-rfc-template.md:70-71`, the
instruction at `SKILL.md:107-108`, and the anti-pattern at
`SKILL.md:241` together admit exactly two states — an exact model id,
or an anti-pattern. A harness whose policy forbids the identifier has
neither available, and finding 1 records the three improvisations that
followed. The `model` field's purpose is scoping, and scoping is served
by knowing the identifier is *deliberately withheld* rather than
forgotten — those are different facts, and the current text cannot
express the difference.

**Change.** Add a rule to the Provenance section of `SKILL.md`,
immediately after the `model` bullet at `SKILL.md:107-108`:

> **WHERE the harness forbids recording the model identifier** — some
> sessions run under a policy that keeps the model id out of artifacts
> pushed to a repository — set `model: withheld` exactly, and state the
> reason in the Provenance table as a `Model withheld` row. Do not
> invent a descriptive placeholder, and do not omit the field: `withheld`
> is a distinct, greppable value meaning "deliberately not recorded",
> which is the fact triage needs. `effort`, `runs_observed` and the
> harness are unaffected and remain required. An RFC whose `model` is
> `withheld` may reach `ready-for-review`.

And amend the gate item at `skill-rfc-template.md:70-71` to:

```
  [ ] Frontmatter sets model + effort (and Provenance names the
      harness) — a skill defect is model-relative. `model: withheld`
      satisfies this WHERE harness policy forbids the identifier, and
      Provenance says so; an empty or absent model does not.
```

And amend the anti-pattern row at `SKILL.md:241` to read "Omitting the
model / effort from Provenance, or writing a free-text placeholder
instead of `withheld`".

**Evidence.** Finding 1. The next run under such a policy has one
defined value to write, one place to explain it, and no reason to hold
the status — so a retro's findings reach triage on their own merits.
`withheld` being a fixed token also keeps the field greppable, which a
free-text placeholder does not.

### P2 — Give `plugin_version` a repo-local spelling [polish]

**Problem.** `skill-rfc-template.md:19` offers `"n/a (standalone
install)"` as the only not-applicable value, which does not describe a
skill living in `.claude/skills/` of the repository being critiqued.
The repository already uses a second spelling that the template does
not mention, and the run picked it by grep rather than by rule.

**Change.** Replace the comment on `skill-rfc-template.md:19`:

```
plugin_version: "<x.y.z>"   # the plugin the skill shipped in; or
                            # "n/a (repo-local skill)" for a skill in
                            # the critiqued repo's own .claude/skills/,
                            # or "n/a (standalone install)" for one
                            # installed outside a plugin
```

**Evidence.** Finding 2. Names the case that actually occurs in this
repository and fixes its spelling, so the label stays aggregatable.

## Considered and Not Proposed

- **The template's gate is inside a comment block the author is told to
  delete.** `skill-rfc-template.md:92` — "Complete each section per the
  comments, then delete all comment blocks" — while `SKILL.md:193` says
  "Complete every section, then run the template's readiness gate". An
  agent that copies the template to the target path and fills it in
  deletes the gate before step 5 asks it to be run. Not proposed: this
  run authored from the reference copy and kept it open, so no wrong
  action occurred. The trap is inferred from reading the two texts
  together, not observed, and the bar requires an observed wrong
  action, failed command, or forced guess. Worth watching for in the
  next run that copies first.
- **Per-skill inbox directories have no existing instance.** Step 6
  routes to `rfcs/inbox/<skill-name>/`; `rfcs/inbox/` was empty and all
  nine committed RFCs sit flat at `rfcs/rfc-000N-*.md`. Not proposed:
  the two are consistent once you see that flat numbered files are
  post-triage, `mkdir -p` resolved it, and no wrong action occurred.
- **Step 7's citation discipline.** "You SHALL NOT state that a check
  was performed unless it was performed" sent the run back to verify
  every number in its own report, which surfaced three wrong counts in
  a separate RFC pushed earlier in the session (30 types not 28; ten
  inherited properties lost not six; 40 documents not "roughly 50").
  Recorded as the skill working, not as a finding.
- **The checklist-in-response instruction** at `SKILL.md:148-149`,
  which explicitly reconciles itself with this project's ban on
  `TodoWrite`/`TaskCreate`. Worked with no friction; recorded as
  working.

## Non-Goals

- Weakening what provenance is collected. `model` should stay
  mandatory and `effort`, `runs_observed` and the harness should stay
  exactly as they are. P1 defines one additional legal value for one
  constrained case; it does not make the field optional.
- Changing `retro-spec`. Its `rfc-template.md` has a readiness gate but
  no `model` field at all, so this finding is a property of
  `retro-skill` alone and does not take the cross-skill route at
  `SKILL.md:181-190`.

## Adoption

P1 and P2 are independent. Both touch two files —
`plugin/skills/retro-skill/SKILL.md` and
`plugin/skills/retro-skill/references/skill-rfc-template.md` — and both
are text-only.

P1 changes whether an RFC can reach `ready-for-review` under a
policy-constrained harness, which is a widening: no RFC that passes the
gate today would stop passing it. Existing RFCs carrying real model ids
are unaffected, and `withheld` is distinguishable from them by exact
match, so triage can filter on it.

Note the template's Lineage block, which asks that it and
`retro-spec`'s `rfc-template.md` be kept "in sync where they overlap".
These proposals touch only fields the skill-focused fork added, so
there is nothing to mirror.

Expected effect: a retro run under a policy that forbids the model
identifier produces a `ready-for-review` RFC with one defined value and
a stated reason, instead of a `draft` carrying three improvisations —
including, once adopted, this RFC and its sibling against `strata`.

## Resolution

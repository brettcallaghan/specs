---
title: "RFC: The compile gate passed on a wrong schema, and the skill has
  no rule for describing a wire format it does not control"
description: A first real run of the strata skill produced a silently
  malformed JSON Schema through both of the skill's gates, and spent
  three probe cycles establishing rules the skill could have stated.
slug: rfc-self-strata-compile-gate
type: rfc
status: draft
updated_at: 2026-08-01T00:00:00Z
origin: self
skill: strata
skill_version: "1.0"
model: "<withheld by session policy — see Provenance>"
effort: high
runs_observed: 1
severity: blocker
plugin_version: "n/a (repo-local skill)"
related: []
---

# The compile gate passed on a wrong schema

## Summary

The `strata` skill was run to model this repository's CLI configuration
and document frontmatter — 30 types across four `.strata` files, driven
to JSON Schema, Go and Markdown. The skill's method is a two-step
compile loop whose signal is the exit code, and that loop passed on a
JSON Schema in which every inheriting struct had silently lost its
inherited properties. The skill never asks the agent to look at what was
emitted, so nothing in it would have caught the defect. Three further
findings are cases where the skill was silent on a rule it could have
stated, each costing a throwaway probe file to establish against the
compiler.

The thesis: the skill teaches that compiling is sufficient, and
compiling is not sufficient.

## Provenance

| Field | Value |
|-------|-------|
| Harness | Claude Code on the web, CLI v2.1.42 |
| Task the skill was applied to | Modelling `specs.toml` and the frontmatter of 11 document types, then generating JSON Schema, Go structs and Markdown |
| References / scripts / templates invoked | `.claude/skills/strata/SKILL.md` v1.0; `references/compiler-options.md`, `references/strata-llm-reference.md`, `references/language-reference.md` (all vendored from strata-lang/strata @ `44c195509444`) |
| Compiler under test | `strata-cli 0.7.0`, built from source in-session |
| Run dates | 2026-08-01 |

**On the `model` frontmatter field.** This session operates under a
policy that forbids writing its model identifier into any artifact
pushed to a repository. The field is therefore a placeholder rather than
omitted, because the template is right that it is load-bearing — a
defect may be model-relative and triage cannot scope it otherwise. The
status is held at `draft` for this reason alone: fill the field in and
the readiness gate passes. Every other gate item is satisfied.

## Evidence

1. **[gap · blocker]** `.claude/skills/strata/SKILL.md:53-54` — "The
   compiler is the only reliable validator, and it is fast (≈6 ms on a
   small file). Never hand back `.strata` you have not compiled." — and
   `SKILL.md:65-66` — "**`strata build` exits non-zero on any failure.**
   That exit code is the signal — check it, don't pattern-match the
   output."

   Following exactly this, with `jsonSchema.allowAdditionalProperties:
   false` set in the manifest, both steps of the loop passed:

   ```
   strata build schemas/ --target ast --output ... --error-format json  # exit 0
   strata build schemas/ --target json-schema --output ...              # exit 0
   ```

   The emitted schema was wrong. Every struct that inherits lost the
   `allOf` wrapper and the `$ref` to its parent, so `Contract` was
   emitted carrying only its own 5 properties, with none of the 10 it
   inherits reachable. The schema is wrong in
   both directions — a document missing `title` validates, and a
   document carrying `title` fails as an undeclared key.

   It was found only by inspecting a definition for an unrelated
   reason. Nothing in the skill's loop looks at the emitted artifact.
   The underlying compiler bug is filed separately against
   `strata-lang/strata`; this finding is about the skill's method,
   which reports success on that output.

2. **[gap · snag]** `SKILL.md:148-150` — "Identifiers are `PascalCase`
   for types, `camelCase` for properties, and **the JSON key is the
   Strata property name verbatim** — transpilers re-case the
   host-language identifier, never the wire key."

   These two clauses conflict whenever the wire format already exists
   and is not camelCase. The frontmatter keys here are `updated_at`,
   `parent_slug`, `source_paths`, `superseded_by`; honouring the
   convention would have renamed them across the 40 committed
   documents that carry this frontmatter, and one approved contract.
   The skill says nothing about
   which clause yields, so the run wrote a throwaway file to ask the
   compiler whether `snake_case` property identifiers parse at all.
   They do, and every target handles them correctly.

3. **[gap · snag]** `SKILL.md:30` — "## The six rules that break
   generated Strata".

   Two rules that break generated Strata are not among the six, and
   together they block the most common shape in any document corpus — a
   base envelope plus a per-type constant tag:

   ```
   Invalid redeclaration of ancestor property 'Base.type'
     Note: Inherited members cannot be redeclared in derived types.

   Variant 'Adr' may not override the discriminator property 'type'
     for union 'SpecDocument'.
     Note: Union variants cannot define a property with the same name
     as the discriminator.
   ```

   A base struct may declare `type` but no derived struct may constrain
   it; a `@discriminator("type")` union pins it per variant but then
   forbids the base from declaring it. Two probe cycles established
   this. The section exists precisely to pre-empt reflexes that produce
   errors, and these are two such reflexes.

4. **[gap · polish]** `SKILL.md:255-261` — "so `strata --version` should
   just work. If it does not:" followed by `cat
   ~/.cache/strata-cli/build.log      # the hook's build output, if it
   failed`.

   Observed: `strata: command not found`, and
   `~/.cache/strata-cli/` did not exist at all. An absent log is a
   different diagnosis from a log containing errors — it means the hook
   never ran, which is a registration problem with a different fix. The
   skill's diagnostic has no branch for it. The run fell through to the
   documented by-hand `cargo build --release -p strata-cli`, which
   worked.

Recurrence: one run. No claim here is stronger than that.

## Proposals

### P1 — Add an inspect-the-output step to the loop [blocker]

**Problem.** The loop at `SKILL.md:51-72` has two steps, validate and
transpile, and presents the exit code as the signal: "That exit code is
the signal — check it, don't pattern-match the output"
(`SKILL.md:65-66`). That instruction is correct about *diagnostics* and
is read as a claim about *correctness*. Finding 1 is a case where both
steps exit zero and the artifact is unusable. The skill offers no step
at which the agent would have looked.

**Change.** Add a third step to the loop, and soften the
sufficiency claim. Proposed replacement for the loop's closing bullets:

```markdown
# 3. Inspect. The exit code proves the SOURCE is valid; it does not
#    prove the OUTPUT is what you wanted. Check the emitted artifact
#    for the structural features you relied on.
```

- **The exit code is necessary, not sufficient.** It reports whether
  the compiler could process your source, not whether the emitted
  artifact carries what you meant. Before committing generated output,
  confirm the features you designed for actually survived: that an
  inheriting type still `$ref`s its parent, that constraints reached
  the schema, that required lists are populated.
- **Non-default emitter options deserve a look.** Anything set under
  `jsonSchema`, `go`, `typescript` and so on changes the emitter's
  path, and the combination you chose may be less exercised than the
  default. `jsonSchema.allowAdditionalProperties: false` is a known
  case: on `strata-cli 0.7.0` it silently drops struct inheritance from
  the output. Verify, then trust.

**Evidence.** Finding 1. The next run under the same model reads a step
that tells it to check the artifact, and a named example of the exact
combination that fails, so the malformed schema is caught before it is
committed rather than by luck.

### P2 — State that the property name follows the wire format [snag]

**Problem.** `SKILL.md:148-150` gives a casing convention and a wire-key
rule as though they never conflict. They conflict for every format the
schema author does not control, which is the common case when adopting
Strata for an existing project. The skill's silence made the run treat
"does snake_case even parse?" as an open question worth a probe.

**Change.** Replace the identifiers paragraph at `SKILL.md:148-150`
with:

> Identifiers are `PascalCase` for types and `camelCase` for
> properties. **The JSON key is the Strata property name verbatim**, so
> when you are describing a format that already exists, the property
> name follows the format and the casing convention yields —
> `updated_at` stays `updated_at`. Transpilers re-case the
> host-language identifier (`UpdatedAt` in Go, with
> `json:"updated_at"`), never the wire key. camelCase is the convention
> for formats Strata itself defines.

**Evidence.** Finding 2. Removes a probe cycle and, more importantly,
removes the risk that an agent reads the convention as binding and
proposes renaming an existing format's keys.

### P3 — Cover inheritance and discriminator collisions in the rules [snag]

**Problem.** The rules section at `SKILL.md:30` is the skill's
pre-emption of reflexes that produce parse and validation errors. The
two rules in finding 3 are exactly that category, are not listed, and
interact: between them, the base-envelope-plus-type-tag shape cannot be
expressed at all. An agent modelling any document corpus meets this
within the first few types.

**Change.** Add a seventh rule to the section, and retitle it "The
seven rules that break generated Strata":

> 7. **Inherited properties are frozen, and the discriminator is
>    exclusive.** A derived struct may not redeclare an inherited
>    property, even to narrow it (`Invalid redeclaration of ancestor
>    property`). A union variant may not declare a property with the
>    discriminator's name (`Discriminator Override`) — which includes
>    one it inherits. So a base envelope carrying `type` cannot be
>    pinned per subtype, and a `@discriminator("type")` union forbids
>    the base from declaring `type` at all. Pick one: an enum-typed
>    property on the base, or a union that owns the tag. Equality
>    constraints on enum-typed properties are also rejected
>    (`Equality constraints are not supported on enum types`).

**Evidence.** Finding 3. The next run picks a construct knowing the
constraint instead of discovering it by collision, and the "Choosing the
construct" section at `SKILL.md:229` gains the missing input.

### P4 — Give the install diagnostic a hook-never-ran branch [polish]

**Problem.** `SKILL.md:255-261` assumes the failure mode is a failed
build and points at its log. It has no branch for the log being absent,
which is what an unregistered hook looks like and which has a different
fix.

**Change.** Replace the comment on the `cat` line and add one bullet:

```bash
strata --version                       # already installed?
cat ~/.cache/strata-cli/build.log      # the hook's build output
```

> **No log file at all** means the hook did not run, rather than that
> it failed. Claude Code registers hooks from the project root only, so
> a session that attaches this repository alongside another never reads
> its `.claude/settings.json`. Run the hook by hand, or build the
> compiler directly with the commands below.

**Evidence.** Finding 4. Turns a dead end into a diagnosis. Note this
proposal is coupled to a repository-side change already made in
`docs/session-bootstrap.md` and `CLAUDE.md`; the skill text should match
whatever that settles on.

## Considered and Not Proposed

- **Output filename under `bundle: true`.** With bundling on, the
  emitted file is named from `packageName` (`SpecModel.schema.json`)
  rather than from the input, so the regeneration script needs a `mv`.
  Not proposed: `ls` on the output directory answered it immediately
  and no wrong action was taken. Below the bar.
- **A worked example of a multi-file package with a manifest.** The run
  assembled one from the three references rather than following an
  example. Not proposed: no wrong action resulted, and the references
  were accurate — this is a "would have been faster", which is taste
  without a defect behind it.
- **`references/compiler-options.md` accuracy.** Examined because P1's
  defect involves an option documented there. The document is correct:
  `allowAdditionalProperties`'s stated semantics match what the
  compiler does with the option; the bug is in the emitter's handling
  of inheritance, not in the option's description. No finding.
- **The "Choosing the construct" and "When to reach for Strata"
  sections.** Both led to correct decisions first time. Recorded as
  working, not as findings.

## Non-Goals

- Fixing the compiler. The `allowAdditionalProperties` inheritance bug
  is filed against `strata-lang/strata` and is not this skill's to fix.
  P1 stands regardless of whether that bug is fixed: the loop should
  verify output because emitters can be wrong, not because one is.
- Changing the vendored references under
  `.claude/skills/strata/references/`. They are generated copies owned
  upstream and regenerated by `scripts/sync-strata-docs.sh`; P2 and P3
  target `SKILL.md` only.

## Adoption

P1 through P4 are independent and touch one file,
`.claude/skills/strata/SKILL.md`. P1 is the one that changes outcomes
rather than speed and should land whether or not the others do.

P3 renames a section heading ("six rules" → "seven rules"), which is
visible to anything that cites it by name; a grep for "six rules" across
the repository found no other citation.

None of the four changes affects behaviour for a run under a different
model — they add rules and a step rather than removing any — so there is
no compatibility concern for existing project copies of the skill.

Expected effect: a run that sets a non-default emitter option is told to
look at what came out; a run describing an existing format stops
treating its key casing as an open question; and the two rules that
block the commonest modelling shape are stated before they are hit
rather than after.

## Resolution

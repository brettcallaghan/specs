# Skills Best-Practices Review — Stack-Ranked Adoption Plan

**Scope:** all 7 skills in this repo — `bead-bonanza`, `groom-beads`,
`groom-feature-blueprint`, `retro-skill`, `retro-spec`, `spec` (under
`plugin/skills/`) and `triage-rfc` (under `.claude/skills/`) — plus their
20 reference files and 2 scripts.

**Yardstick:** [Skill authoring best practices](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices)
(the "BP doc" below). Every recommendation cites the specific guidance it
rests on. Where this repo's *own* normative documents say the same thing,
that is cited too — a rule you already wrote is cheaper to adopt than a
new one.

**How this is ranked:** by expected cost of *not* fixing it. Tier 1 items
are live defects that will make a real run produce a wrong result. Tier 2
are structural — they degrade every run a little. Tier 3 is the missing
measurement layer. Tiers 4–5 are determinism and polish.

**Overall assessment.** These skills are unusually strong on the things
most skill collections get wrong: descriptions are specific, third-person,
and rich in trigger terms; single-source-of-truth is enforced by
convention with explicit "do not fork this" warnings; anti-pattern tables
are concrete; and the retro→triage loop is a genuinely sophisticated
feedback design. The weaknesses are concentrated in three places —
(1) a handful of hard contradictions between skills that no one is
checking, (2) progressive disclosure applied to documents but not to the
skills themselves, and (3) a complete absence of evaluations.

---

## Contents

- Tier 1 — Live defects (P1–P5)
- Tier 2 — Progressive disclosure and token budget (P6–P9)
- Tier 3 — Evaluation and multi-model testing (P10–P11)
- Tier 4 — Determinism, scripts, feedback loops (P12–P14)
- Tier 5 — Content-guideline polish (P15–P19)
- Appendix A — BP-doc checklist, scored
- Appendix B — What is already exemplary (do not "fix" these)

---

# Tier 1 — Live defects

These will produce a wrong result on a real run. Fix before anything else.

## P1. The model vocabulary contradicts itself across four files, and `bead-bonanza`'s own example fails validation

**Severity: blocker.** Two independent contradictions, one of membership
and one of ordering.

*Membership.* The closed set is declared as `opus | sonnet | fable` in
four places:

| File | Line | Text |
|---|---|---|
| `plugin/skills/spec/references/work-order-field-guide.md` | 44 | ``- `model` — closed set (`opus \| sonnet \| fable`). Never null.`` |
| `plugin/skills/spec/references/work-order-definition-of-ready.md` | 90 | ``` `model` from the closed set (e.g. `opus | sonnet | fable`) ``` |
| `plugin/skills/groom-beads/references/bead.schema.json` | 63 | `"enum": ["opus", "sonnet", "fable"]` |
| `plugin/skills/groom-beads/scripts/validate_beads.py` | 42 | `MODELS = {"opus", "sonnet", "fable"}` |

But `bead-bonanza/SKILL.md:256-258` declares a *different* set —
`` (`haiku` | `sonnet` | `opus` | `fable`) `` — and then uses it: the
worked dispatch table at `bead-bonanza/SKILL.md:302-303` assigns `haiku`
to two beads.

**Consequence:** a bead groomed exactly as `bead-bonanza`'s own example
shows fails `validate_beads.py` with a hard error —
`metadata.model 'haiku' not in ['fable', 'opus', 'sonnet']` (line 157) —
and fails `bead.schema.json` under `jsonschema`. The error message at
`validate_beads.py:178` even reprints the wrong-for-bead-bonanza set:
`"ready bead has no metadata.model (must be non-null; opus|sonnet|fable)"`.

*Ordering.* `work-order-field-guide.md:81` routes to `fable` as **"the
cheap/fast model"**. `bead-bonanza/SKILL.md:280-283` makes `fable` the
**top of the escalation ladder**: "re-dispatch that bead one tier up
(haiku→sonnet→opus→fable)". So the two skills disagree about whether
`fable` is the cheapest or the most capable tier. An orchestrator
following `bead-bonanza` escalates a struggling bead *down* the field
guide's capability ladder.

**Best practice violated.** [Use consistent terminology](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#use-consistent-terminology):
> "Choose one term and use it throughout the Skill... Consistency helps
> Claude parse and follow instructions."

And the [checklist](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#checklist-for-effective-skills):
"Consistent terminology throughout"; "Scripts solve problems rather than
defer to Claude". A validator that rejects the documented-correct input
is the script-defect shape `retro-skill/SKILL.md:71-76` itself defines.

**Recommended change.**

1. Decide the real lineup and the real tier order once. Record it in
   `work-order-field-guide.md` only — it is already declared the owner by
   `work-order-definition-of-ready.md:91` and `groom-beads/SKILL.md:177`
   (line 90 states the set; line 91 points at the field guide as owner).
2. **Express the ladder as tiers, not model names.** Replace the
   `haiku→sonnet→opus→fable` chain in `bead-bonanza:280-283` with
   `cheap-fast → mid → top`, and let the field guide hold the single
   name→tier mapping. This removes the class of bug rather than this
   instance of it (see plan2 §5 for the full argument).
3. Make `bead.schema.json:63` and `validate_beads.py:42` derive from —
   or at minimum be checked against — that one list. Cheapest version:
   drop the `enum` from the schema's `model` property, keep
   `model_reason` required, and let the CI consistency check (plan2 §2)
   assert schema-enum == field-guide list.
4. Fix `validate_beads.py:178`'s message to name the set it actually
   enforces, so the error is self-correcting.

---

## P2. `bd-health.sh` never runs its own hard graph-integrity check

**Severity: blocker.** `plugin/skills/groom-beads/scripts/bd-health.sh`
passes multi-word subcommands as a single quoted argv element:

```bash
# line 31
have() { "$BD" "$1" --help >/dev/null 2>&1; }
# line 41
out="$("$BD" "$sub" 2>&1)"; rc=$?
# line 54
run_check "swarm validate" "swarm validate" hard
```

`"$sub"` with `sub="swarm validate"` invokes `bd 'swarm validate' --help`
— one argument containing a space, not two arguments. No `bd` build has a
subcommand literally named `swarm validate`, so `have` always returns
non-zero and the check reports itself skipped.

**Reproduced:**

```console
$ sub="swarm validate"
$ fakebd "$sub" --help    # a bd that DOES support `swarm validate`
quoted   ("$sub"): reported NOT available  <-- bug
unquoted ($sub)  : AVAILABLE  <-- correct
```

**Consequence:** `swarm validate` is one of only two `hard` checks — it is
the one that finds "cycles / orphans / disconnected per epic"
(comment, line 54). It can never set `hard_fail`, so `bd-health.sh` can
never fail the run for a dependency cycle. Cycles are precisely what
`groom-beads/SKILL.md:234-235` calls out: *"cycles and orphans and
dangling references are hard problems — fix them, don't report around
them."* The script silently reports around them.

**Best practice violated.** [Solve, don't defer](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#solve-dont-defer)
and [Provide utility scripts](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#provide-utility-scripts)
("More reliable than generated code"); checklist item "Error handling is
explicit and helpful". A check that reports "skipped" when the capability
is present is worse than no check — it launders a missing gate as a
handled case.

**Recommended change.** Pass the subcommand as words. Minimal patch:

```bash
have() { "$BD" $1 --help >/dev/null 2>&1; }   # intentional word-split
run_check() {
  local sub="$1" label="$2" severity="${3:-soft}"
  if ! have "$sub"; then ... ; fi
  out="$("$BD" $sub 2>&1)"; rc=$?
  ...
}
```

Add `# shellcheck disable=SC2086` with a one-line reason so the
intentional split is not "fixed" back. Then add a self-test: the script
should print which checks actually ran, so a silent skip is visible.
Consider a `--require swarm-validate` flag that exits 2 when a check the
caller demanded is unavailable, so absence is loud.

---

## P3. `$TMPDIR` is unset on most Linux hosts, so every `bead-bonanza` run directory is an unwritable absolute path

**Severity: blocker on Linux.** `bead-bonanza/SKILL.md:59-63` prescribes:

```
$TMPDIR/claude/bead-bonanza/<run-id>/
```

and repeats that literal in nine more places (`:165`, `:297`, `:363`,
`:409-410`, `:436`, `:439`, `:502`, `:534`, `:582`). `TMPDIR` is set by
default on macOS but **not** on typical Linux distributions or in most
containers.

**Reproduced in this environment:**

```console
$ echo "TMPDIR=[${TMPDIR-<unset>}]"
TMPDIR=[<unset>]
$ echo "$TMPDIR/claude/bead-bonanza/run1"
/claude/bead-bonanza/run1
```

`mkdir -p /claude/...` fails with `Permission denied` for a non-root
user. Every downstream instruction then fails: the progress JSONL
(`:363`), the retro file that `:397` calls **"non-negotiable"** because it
is the crash-recovery anchor, the decisions doc, and the dispatch plan.
The skill's own recovery procedure (`:343-359`) reads from that path, so a
crash during a run started this way is unrecoverable — the exact failure
the retro-file design exists to prevent.

The skill is half-aware of the variance: `:53` allowlists
`Read($TMPDIR/**)`, `Read(/tmp/claude/**)` **and**
`Read(/private/tmp/claude/**)` — three roots — but the body only ever
writes to one.

**Best practice violated.** [Solve, don't defer](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#solve-dont-defer):
> "handle error conditions rather than deferring to Claude"

and its "voodoo constants" corollary — an unresolvable path is an
unjustified configuration value. Also
[Avoid assuming tools are installed](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#avoid-assuming-tools-are-installed),
whose principle is the same: do not assume the environment supplies
something.

**Recommended change.** Replace all ten occurrences with
`${TMPDIR:-/tmp}`, and add one line to Phase 0 step 4:

> Resolve the run root once: `RUN_DIR="${TMPDIR:-/tmp}/claude/bead-bonanza/<run-id>"`.
> Create it and confirm it is writable (`touch "$RUN_DIR/.ok"`) before
> Phase 1 — every later phase, including crash recovery, depends on it.

This also makes the three allowlist entries at `:53` collapse to one
predictable root, which is a permission-prompt win.

---

## P4. Sibling reference paths inside `references/` do not resolve from the file that cites them

**Severity: snag, high frequency.** Files inside
`plugin/skills/spec/references/` cite their own directory-mates with a
`references/` prefix — a path that is correct relative to the *skill
root* but wrong relative to the *citing file*:

| File | Line | Cites |
|---|---|---|
| `container-blueprint-template.md` | 23, 42, 44, 59, 94, 127, 159, 304, 368 | `` `references/c4-guide.md` `` (×8), `` `references/contract-template.md` `` |
| `module-blueprint-template.md` | 27, 48, 82, 183, 300, 375 | `` `references/c4-guide.md` ``, `` `references/contract-template.md` `` |
| `system-blueprint-template.md` | (5×) | `` `references/c4-guide.md` ``, `` `references/contract-template.md` `` |
| `component-blueprint-template.md` | 19 (+3) | `` `references/c4-guide.md` `` |
| `work-order-definition-of-ready.md` | 27, 28, 91, 104 | `` `references/work-order-field-guide.md` ``, `` `references/work-order-template.md` `` |
| `work-order-field-guide.md` | 18 | `` `references/work-order-template.md` `` |
| `blueprint-readiness-checklist.md` | 59, 104 | `` `references/c4-guide.md` `` |
| `feature-blueprint-template.md` | 111, 462 | `` `references/c4-guide.md` `` |

The clinching evidence that this is drift rather than design is
`blueprint-readiness-checklist.md:100-106`, where the References list
uses **two conventions two lines apart**:

```markdown
- `feature-blueprint-template.md` (same directory) — the document
  these checks apply to, including the AC Coverage row validity rules
- `references/c4-guide.md` — element declarations and reference rules
```

Both targets are in the same directory as the citing file. One path
resolves; one does not.

**Note on intent.** Backticked paths for framework guides are *deliberate*
— `docs/adrs/ADR-012-reference-conventions.md:56-57` and
`template-authoring-standards.md:632-643` both sanction
`` `references/c4-guide.md` ``. The convention is fine. The bug is that
the convention was written from the SKILL.md's vantage point and then
copied into files one level deeper, where the prefix is wrong.

**Best practice violated.** [Runtime environment](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#runtime-environment):
> "**File paths matter:** Claude navigates your skill directory like a
> filesystem."

Checklist: "File references are one level deep". A path that resolves
from nowhere the agent is standing costs a `Glob`/`Grep` round trip at
best, and a silently-skipped load at worst.

**Recommended change.** Add one normative rule to
`template-authoring-standards.md` §"References Outside the Document Set"
and apply it mechanically:

> A backticked path SHALL be written relative to the file that contains
> it. From `SKILL.md`, a bundled guide is `` `references/c4-guide.md` ``;
> from inside `references/`, the same guide is `` `c4-guide.md` ``.

Then fix the ~30 occurrences above. Add it to the CI consistency check
(plan2 §2): *no file under `*/references/` may contain the string
`` `references/ ``*. That is a one-line grep and it closes the class.

---

## P5. `templates/rfc.md` is a duplicate of the retro-spec RFC template, and it has already drifted

**Severity: snag.** `templates/rfc.md` (233 lines) and
`plugin/skills/retro-spec/references/rfc-template.md` (242 lines) are the
same template. They now differ by 74 diff lines:

| | `templates/rfc.md` | `retro-spec/references/rfc-template.md` |
|---|---|---|
| Instruction-block `Version:` | `1.0` | `1.1` |
| Terminology | "**dispositions** the RFC", "who **dispositions** the proposals" | "**resolves** the RFC", "who **resolves** the proposals" |
| Scope block | absent | "Scope — one RFC per retro per audience … do NOT split proposals into separate RFC files" |
| `title` placeholder | `<Proposal Title>` | `<Title Spanning the Proposal Set>` |

The stale copy still uses "disposition", which
`.claude/skills/triage-rfc/SKILL.md:56-57` describes as the *superseded*
heading. And it is missing the one-RFC-per-retro scope rule that
`retro-skill/SKILL.md:204` treats as an anti-pattern to enforce.

**Best practice violated.** This is the repo's own anti-pattern, from
`plugin/skills/spec/SKILL.md:272`:

> | Duplicating content across documents | Copies drift; cross-reference with `[[document-slug]]` instead |

and `template-authoring-standards.md:655-664` ("Avoid Duplication …
Duplication leads to drift when only one copy is updated"). The BP doc
frames the same point structurally: reference files are
[bundled with the skill that owns them](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#runtime-environment).

**Recommended change.** Delete `templates/rfc.md`. `README.md:8` already
carries this as a TODO ("Migrate @templates/rfc.md into an appropriate
skill; local or published") — it is now overdue and has cost a version of
drift. If a repo-root copy is wanted for humans, make it a one-line
pointer, not a copy. Then check whether `templates/` should exist at all;
it has one file.

---

# Tier 2 — Progressive disclosure and token budget

## P6. `bead-bonanza/SKILL.md` is 615 lines with no `references/` directory

**Severity: high, affects every run.** Measured line counts of every
SKILL.md body in the repo:

| Skill | Lines | Over the 500-line guidance? |
|---|---|---|
| `bead-bonanza` | **615** | **yes, +23%** |
| `groom-feature-blueprint` | 318 | no |
| `groom-beads` | 312 | no |
| `spec` | 294 | no |
| `retro-skill` | 212 | no |
| `retro-spec` | 184 | no |
| `triage-rfc` | 120 | no |

`bead-bonanza` is also the only skill with zero bundled files
(`references=0 scripts=0`), while `spec` bundles 17.

**Best practice violated.** [Token budgets](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#token-budgets):
> "Keep SKILL.md body under 500 lines for optimal performance. If your
> content exceeds this, split it into separate files using the
> progressive disclosure patterns described earlier."

And the [Conditional workflow pattern](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#conditional-workflow-pattern)
tip:
> "If workflows become large or complicated with many steps, consider
> pushing them into separate files and tell Claude to read the
> appropriate file based on the task at hand."

**Recommended change.** This skill is an unusually clean split candidate,
because large blocks of it are needed only at one specific moment:

| Extract | Current lines | ~Size | Loaded when |
|---|---|---|---|
| `references/subagent-prompts.md` — the 4 prompt templates (per-bead, fast-lane, shared-surface, worktree notes) | 399–571 | ~170 | at dispatch, Phase 3 |
| `references/gate-profile.md` — the profile template, custom gates, targeted-tests list, environmental-failure ledger | 19–48 | ~30 | Phase 0 step 1, and only on the slow path |
| `references/recovery.md` — crash/timeout recovery + progress-event schema + Monitor pattern | 343–397 | ~55 | only on a crash |

That is ~255 lines out of the body, landing SKILL.md near **360 lines**
with no content lost. It matches BP
[Pattern 3: Conditional details](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#pattern-3-conditional-details)
exactly — show the phase flow, link the machinery.

Keep in the body: the phase skeleton, the lean-path rules, the dispatch
classification, the push policy, and Operating principles.

Two cautions:
- Keep all three new files **one hop from SKILL.md** (BP:
  [Avoid deeply nested references](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#avoid-deeply-nested-references)),
  and make the execution intent explicit at each link — "read
  `references/subagent-prompts.md` and fill the template for this
  bucket", not a bare pointer (BP:
  [Make execution intent clear](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#runtime-environment)).
- The **Operating principles** section (`:586-606`, 21 bullets, ~1,900
  words) largely restates rules stated earlier in the phases. Its
  restatement value is real for a long autonomous run, but it should be
  audited for pure duplication and cut to the handful of principles that
  are *not* stated elsewhere. Cite BP
  [Concise is key](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#concise-is-key):
  "Does this paragraph justify its token cost?"

---

## P7. No reference file over 100 lines has a table of contents — 18 of 20 qualify

**Severity: high, affects every reference load.** Measured: zero files
match `^#+ (Contents|Table of Contents)`.

| Lines | File |
|---|---|
| 737 | `spec/references/template-authoring-standards.md` |
| 464 | `spec/references/feature-blueprint-template.md` |
| 455 | `spec/references/c4-guide.md` |
| 377 | `spec/references/module-blueprint-template.md` |
| 369 | `spec/references/container-blueprint-template.md` |
| 265 | `spec/references/system-blueprint-template.md` |
| 251 | `spec/references/ears-guide.md` |
| 243 | `spec/references/component-blueprint-template.md` |
| 242 | `retro-spec/references/rfc-template.md` |
| 213 | `spec/references/standards-template.md` |
| 208 | `spec/references/frd-template.md` |
| 196 | `retro-skill/references/skill-rfc-template.md` |
| 193 | `spec/references/contract-template.md` |
| 191 | `spec/references/work-order-template.md` |
| 179 | `spec/references/adr-template.md` |
| 171 | `spec/references/work-order-definition-of-ready.md` |
| 145 | `spec/references/ac-traceability-standards.md` |
| 106 | `spec/references/blueprint-readiness-checklist.md` |

**Best practice violated.** [Structure longer reference files with table of contents](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#structure-longer-reference-files-with-table-of-contents):
> "For reference files longer than 100 lines, include a table of contents
> at the top. **This ensures Claude can see the full scope of available
> information even when previewing with partial reads.**"

The partial-read risk is not hypothetical here. The BP doc warns that
when files are reached through a chain, "Claude might use commands like
`head -100` to preview content rather than reading entire files". A
737-line standards document reached at hop two (P8) with no TOC is the
worst case in the repo: `head -100` of
`template-authoring-standards.md` returns frontmatter, line-length, and
part of the frontmatter-format section — and none of the slug registry,
reference syntax, or versioning rules that consumers actually come for.

**Recommended change.** Add a `## Contents` section immediately after the
H1 of all 18 files. Two grades of effort:

- **Templates** (adr, frd, standards, contract, the four blueprints,
  work-order, both rfc): list the document sections in order. This is
  mechanical — generate from the `^## ` headings.
- **Guides and standards** (`template-authoring-standards`, `c4-guide`,
  `ears-guide`, `ac-traceability-standards`,
  `work-order-definition-of-ready`, `work-order-field-guide`,
  `blueprint-readiness-checklist`): hand-write, and make each entry say
  what question the section answers. The BP example is one line per
  topic, not a nested outline:

  ```markdown
  ## Contents
  - Line length, frontmatter, and required properties
  - Slug conventions and the prefix registry
  - Template instruction blocks and versioning
  - EARS syntax and RFC 2119 modal verbs
  - Cross-document references (`[[slug]]`, backticked paths)
  - Validation (frontmatter schemas, reference checks)
  ```

Add "TOC present if >100 lines" to the CI consistency check.

**Related:** consider whether `template-authoring-standards.md` at 737
lines should be split at all. It bundles six genuinely separable
concerns (structure, frontmatter, slugs, instructions, EARS,
references+validation). It is cited by name from 8 other files, so a
split is a coordinated change — but every one of those citations wants
exactly one of the six sections. A TOC is the cheap fix; a split is the
right one.

---

## P8. Two-hop reference chains: SKILL.md → template → guide

**Severity: medium.** The architecture-authoring path is a chain, not a
fan:

```
spec/SKILL.md
  └─→ references/container-blueprint-template.md   (hop 1)
        └─→ references/c4-guide.md                 (hop 2, cited 8×)
```

Same shape for `module-blueprint-template.md` (5×),
`system-blueprint-template.md` (5×), `component-blueprint-template.md`
(4×), `contract-template.md` (3×). The work-order trio has it too:
`work-order-definition-of-ready.md:27-28` → `work-order-field-guide.md`
and `work-order-template.md`.

**Best practice violated.** [Avoid deeply nested references](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#avoid-deeply-nested-references):
> "Claude may partially read files when they're referenced from other
> referenced files… **Keep references one level deep from SKILL.md.** All
> reference files should link directly from SKILL.md to ensure Claude
> reads complete files when needed."

**Mitigating credit — the repo already does this right once.**
`spec/SKILL.md:81-84` says: *"The architecture documents follow the C4
model with the framework's documented extensions — load
`references/c4-guide.md` before authoring any of them."* That is a
hop-1 load obligation stated at the SKILL.md level. The fix is to
generalize that pattern, not invent one.

**Recommended change.** Make the SKILL.md References table state
**bundles**, not single files, so the co-load is a hop-1 instruction and
the in-template mentions degrade to reminders:

```markdown
| Working on | Load together |
|------------|---------------|
| Documenting a deployable unit (app or owned data store) | `references/container-blueprint-template.md` + `references/c4-guide.md` |
| Documenting an atomic component of one container | `references/component-blueprint-template.md` + `references/c4-guide.md` |
| Understanding / filling / gating a work order | `references/work-order-template.md` + `references/work-order-field-guide.md` + `references/work-order-definition-of-ready.md` |
| Writing any requirement or acceptance criterion | `references/ears-guide.md` |
```

Do **not** duplicate the guide content into templates — that trades a
navigation problem for the drift problem P5 demonstrates. The change is
purely to where the load instruction lives.

---

## P9. Cross-skill `../` references escape the skill directory

**Severity: medium; design tension worth an explicit decision.** Four
skills read files owned by a sibling:

| Reader | Lines | Target |
|---|---|---|
| `groom-beads/SKILL.md` | 50-53, 87, 129-133, 177, 288-291 | `../spec/references/{work-order-definition-of-ready,work-order-field-guide,work-order-template}.md` |
| `groom-feature-blueprint/SKILL.md` | 36-45, 52, 87, 146, 182 | `../spec/references/{feature-blueprint-template,blueprint-readiness-checklist,work-order-template,ac-traceability-standards}.md`, `../retro-spec/references/findings-capture-protocol.md` |
| `retro-skill/SKILL.md` | 53 | `../retro-spec/references/findings-capture-protocol.md` |
| `retro-spec/SKILL.md` | 132-133 | `../spec/references/template-authoring-standards.md` |
| `bead-bonanza/SKILL.md` | 87, 133 | `../spec/references/work-order-definition-of-ready.md` |

The repo knows this is load-bearing and fragile —
`scripts/install-spec-skill.sh:6-9`:

> "`groom-feature-blueprint` references its sibling skills' reference
> files by relative path (`../spec/references/...`), so all skills under
> `plugin/skills/` must be installed together, side by side, for those
> cross-references to resolve."

**Best practice tension.** BP's
[Runtime environment](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#runtime-environment)
frames a skill's resources as *bundled*: "Bundle comprehensive resources:
include complete API docs, extensive examples, large datasets; no context
penalty until accessed." The whole progressive-disclosure model assumes
the tree under `SKILL.md` is self-contained. `../` breaks that: copy one
skill and you get dangling references; the target is undiscoverable from
the reader's own tree; and the install script has to document a
co-location invariant that nothing enforces.

**But the no-duplication choice is correct** and should not be reversed.
`groom-beads/SKILL.md:46` states the reason well: *"canonical copies live
in the `spec` skill, so authoring and grooming can never disagree."* P5
is what happens when you duplicate instead.

**Recommended change — generalize the vendoring pattern the repo already
invented.** `groom-feature-blueprint/SKILL.md:181-188` already solves
exactly this problem for one file, and states the reason crisply:

> "copy `../retro-spec/references/findings-capture-protocol.md` to
> `rfcs/findings-capture.md` in the target project… so any implementing
> agent — including one without this plugin installed — can follow the
> protocol by reading a single file inside the project, with no
> permission prompts and no skill loading."

Apply the same idea at *build* time instead of run time:

1. Author each shared file once, in its owning skill (unchanged).
2. Add `scripts/sync-shared-references.sh`: copy each spec-owned shared
   file into every consumer's `references/shared/`, prepending a stamped
   header — `<!-- GENERATED COPY. Source of truth:
   plugin/skills/spec/references/<f> v<X.Y>. Do not edit. -->`.
3. Rewrite the readers' citations from `../spec/references/<f>` to
   `references/shared/<f>` — now one hop, inside the bundle.
4. CI check: every generated copy is byte-identical to its source, or the
   build fails. One authored copy; N bundled copies; drift is impossible
   rather than merely discouraged.

**Cheaper alternative** if the build step is unwanted: keep `../` but add
a one-line "why this path leaves the skill" note at each site plus the
co-installation requirement, so an agent that fails the read knows the
diagnosis rather than improvising. This is strictly worse but costs
minutes.

---

# Tier 3 — Evaluation and multi-model testing

## P10. There are zero evaluations for 9,300 lines of skill content

**Severity: highest structural gap in the repo.** A filesystem search for
`*eval*` and `*test*` across the whole tree returns nothing.

**Best practice violated.** This is the most emphatic instruction in the
BP doc — [Build evaluations first](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#build-evaluations-first):

> "**Create evaluations BEFORE writing extensive documentation.** This
> ensures your Skill solves real problems rather than documenting
> imagined ones."

with the five-step loop:

> 1. **Identify gaps:** Run Claude on representative tasks without a
>    Skill. Document specific failures or missing context
> 2. **Create evaluations:** Build three scenarios that test these gaps
> 3. **Establish baseline:** Measure Claude's performance without the Skill
> 4. **Write minimal instructions:** Create just enough content to address
>    the gaps and pass evaluations
> 5. **Iterate:** Execute evaluations, compare against baseline, and refine

and the checklist item: "**At least three evaluations created**".

The BP doc's closing note is worth quoting for the sequencing argument:
"Evaluations are your source of truth for measuring Skill effectiveness."
Right now the source of truth for these skills is authorial conviction.
Every Tier 1 defect above is a bug an evaluation would have caught on
first run — P1 by grooming one bead and validating it, P2 by running the
script against a repo with a known cycle, P3 by running on Linux.

**Recommended change.** Create `evals/<skill>/NN-<scenario>.json` in the
BP schema, three per skill, 21 total. Order by risk:

1. **`groom-beads`** — the highest-value target, because its output is
   machine-checkable. Eval: an under-groomed bead in, `validate_beads.py`
   reports 0 errors out. This one eval catches P1 and P2 together.
2. **`spec`** — type selection is a classification task with a right
   answer. Three inputs from `SKILL.md:115-163`'s boundary cases (a
   private schema; a component-vs-module call; a policy that looks like an
   ADR).
3. **`groom-feature-blueprint`** — the Phase 1 gate is a
   bounce-or-proceed decision. One blueprint that must bounce (a content
   check fails), one that must stop without modification (a readiness
   check fails — the distinction at `SKILL.md:58-82` is subtle and
   high-consequence), one that must proceed.
4. **`bead-bonanza`** — dispatch-mode classification (`trivial` /
   `isolated` / `shared-surface`) from a fixture queue, plus the
   skip-claim reconciliation at `:330-336`.
5. **`retro-skill` / `retro-spec`** — the null result is the interesting
   case: feed a clean run and assert a retro *note*, not a padded RFC.
   `retro-skill:143` says a null "is a real, valuable result"; nothing
   currently verifies the skill produces one.
6. **`triage-rfc`** — assert the human gate holds: given an inbox, the
   run must stop at step 4 with a classification table and touch no
   files.

Example, in the BP doc's exact structure:

```json
{
  "skills": ["groom-beads"],
  "query": "Groom the beads in this backlog to the Definition of Ready, then validate.",
  "files": ["evals/fixtures/backlog-ungroomed.jsonl", "evals/fixtures/c4/components.md"],
  "expected_behavior": [
    "Loads the spec-owned Definition of Ready rather than paraphrasing it",
    "Keeps Problem / Design / Acceptance as three distinct fields",
    "Inlines the referent for every session-dependent phrase instead of deleting the sentence",
    "Assigns a non-null model with a model_reason to every ready bead",
    "Wires ordering as dependency edges, not priority numbers",
    "Runs validate_beads.py and iterates until it reports 0 errors",
    "Reports any bead it could not bring to Ready with the specific gap"
  ]
}
```

The BP doc notes there is no built-in runner, so also add
`evals/README.md` stating how a human runs these and records outcomes.
Keep the fixtures tiny — the point is regression detection, not coverage.

---

## P11. Multi-model testing is undocumented, and these skills are explicitly aimed at small models

**Severity: high.** No evidence of Haiku/Sonnet/Opus testing exists.

**Best practice violated.** [Test with all models you plan to use](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#test-with-all-models-you-plan-to-use):

> "Skills act as additions to models, so effectiveness depends on the
> underlying model… **Claude Haiku** (fast, economical): Does the Skill
> provide enough guidance? … What works perfectly for Opus might need
> more detail for Haiku."

Checklist: "Tested with Haiku, Sonnet, and Opus".

**Why this bites harder here than usual.** These skills do not merely
*run* on cheap models — they *route work to them by design*:

- `work-order-field-guide.md:81-84` routes mechanical work to the
  cheap/fast tier.
- `bead-bonanza:302-303` dispatches fast-lane beads on the cheapest
  model, and `:265-269` passes `model:` straight to the `Agent` call.
- `bead-bonanza:240` caps a fast-lane subagent at ~6 beads, each running
  the full per-bead ceremony from `:399-484` — a ~85-line prompt with
  10 ordered steps, three exit paths, and 6 constraints.

So the least-capable model in the lineup is handed the longest procedural
prompt in the repo. Meanwhile the prose style across these skills is
dense and allusive — "Bounce, don't guess", "The push is the gate", "Red
before green", "a bead that reads fine to someone who was in the planning
session but is unworkable cold". That compresses beautifully for Opus and
is exactly what BP warns may under-specify for Haiku.

**The repo already believes model identity matters** —
`retro-skill/SKILL.md:96-110` makes model and effort *load-bearing
provenance*:

> "A skill behaves differently under different models, efforts, and
> harnesses. A defect real under one model may not reproduce under
> another, and the triage side cannot scope the fix without knowing."

That is the argument for multi-model testing, already written down. It is
just never applied to the skills themselves.

**Recommended change.**

1. Declare a support matrix per skill in a repo-level doc: which models
   each skill is *authored for* and which its subagents may run on.
2. Run the P10 evals on the cheapest tier the skill dispatches to, and
   record results. The likely finding: the aphoristic principle bullets
   need a concrete restatement for small models.
3. Where a gap appears, prefer BP's
   [Examples pattern](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#examples-pattern)
   over more prose: "Examples convey the desired style and level of
   detail to Claude more clearly than descriptions alone." An
   under-groomed→ready input/output pair does more for a small model than
   a paragraph about context-independence.
   `work-order-definition-of-ready.md:106-172` already has three such
   worked examples — the best small-model affordance in the repo. Copy
   that move into `bead-bonanza` (one worked dispatch classification) and
   `retro-skill` (one worked finding at each of the three shapes;
   `:71-79` gestures at examples but they are fragments, not pairs).

---

# Tier 4 — Determinism, scripts, feedback loops

## P12. `bead-bonanza` hand-specifies fragile mechanics in prose where a script belongs

**Severity: high.** The skill has zero scripts and instead prescribes, in
English, several operations that are exactly the kind BP says to
pre-write:

| Prose mechanism | Lines | Why it is script-shaped |
|---|---|---|
| Progress-event append with a literal-quoting rule to dodge a permission heuristic | 376 | "**Events are FULLY LITERAL single-quoted strings** … Never embed `$(date ...)`, `"$var"`, or any runtime substitution: the permission system flags runtime-determined printf operands (zsh arith-eval injection heuristic) and prompts the user on EVERY event write." |
| Skip-claim reconciliation via diff-stat, orchestrator and subagent sides | 330–336, 459–470 | Two symmetrical procedures with a 3-way branch, described twice |
| Crash-recovery close ceremony | 343–359 | 7 ordered steps including a conditional `bd export` redirect |
| Atomic close + export + stage + commit | 434–441 | "redirect if your build of `bd` writes to stdout", "confirm the file actually changed" |

**Best practice violated.** [Provide utility scripts](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#provide-utility-scripts):

> "Even if Claude could write a script, pre-made scripts offer
> advantages: **More reliable** than generated code; **Save tokens** (no
> need to include code in context); **Save time**; **Ensure consistency**
> across uses."

and [Runtime environment](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#runtime-environment):
"**Prefer scripts for deterministic operations:** Write
`validate_form.py` rather than asking Claude to generate validation
code."

The progress-event case is the strongest: the skill spends a paragraph
teaching the agent to hand-satisfy a permission heuristic. That is a
workaround for a missing script, and it will be re-derived — or gotten
wrong — on every run.

**Recommended change.** Add `plugin/skills/bead-bonanza/scripts/`:

1. `bb-progress.sh <run-dir> <bead-id> <phase> [sha]` — appends one
   well-formed JSONL event. Deletes the entire quoting hazard at `:376`,
   and collapses N permission prompts into one allowlist entry
   (`Bash(bb-progress.sh:*)`), which is also a Phase-0 step-2
   simplification.
2. `bb-verify-skip.sh` — runs both diff-stats, exits `0` (clean skip
   consistent) / `1` (dirty tree, skip contradicted). Replaces two prose
   procedures with one shared check, so orchestrator and subagent cannot
   diverge. BP:
   [Create verifiable intermediate outputs](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#create-verifiable-intermediate-outputs)
   — "Machine-verifiable: Scripts provide objective verification."
3. `bb-close-bead.sh <id> <priority> <title> <retro-file>` — atomic close,
   `bd export` with the stdout-vs-file detection handled *in code*, stage,
   single commit with the canonical message. This is the highest-value
   one: it is the ceremony the skill calls non-negotiable, it is
   duplicated across three prompt templates (`:434-441`, `:502`,
   `:534`), and it is what crash recovery must replay by hand today.

Follow BP's script guidance when writing them: explicit error handling
("Solve, don't defer"), no unexplained constants, verbose failure
messages ("Make validation scripts verbose with specific error
messages"), and state clearly in SKILL.md that these are **executed**,
not read.

Also apply BP's
[Package dependencies](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#package-dependencies)
rule: keep them POSIX `sh`/`bash` + `git` + `bd` only, since
`bead-bonanza` claims to be language-agnostic.

---

## P13. Feedback loops are closed in one place and left open in five

**Severity: medium.** BP:
[Implement feedback loops](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#implement-feedback-loops):

> "**Common pattern:** Run validator → fix errors → repeat. This pattern
> greatly improves output quality." … "**Only proceed when validation
> passes**"

**The repo does this perfectly once.** `groom-beads/SKILL.md:230-232`:

> "**Iterate emit → validate → fix until it reports 0 errors** (warnings
> are advisory…)"

That is the model. Now the open ends:

| Site | Gap |
|---|---|
| `groom-beads/SKILL.md:215-219` (`bd` present path) | "run the health checks via `scripts/bd-health.sh`, which runs whatever of these exist and summarizes". No fix-and-re-run loop, and no "do not proceed while it exits 1" — even though the script's exit code 1 exists precisely for that gate. The file-emit sibling path two paragraphs later has the loop. Same phase, two rigours. |
| `spec/SKILL.md:211-222` (blueprint preflight) | "run the decision sweep and self-check against `references/blueprint-readiness-checklist.md`" — one pass, no re-check after fixing. Compare `groom-feature-blueprint/SKILL.md:49-82`, which *does* define failure responses precisely. |
| `retro-skill/SKILL.md:161-169` and `retro-spec/SKILL.md:129-138` | "run the template's readiness gate, set status `ready-for-review`" — no instruction to re-run after remediation. |
| `triage-rfc/SKILL.md:78-97` | Step 5 applies changes and bumps versions with no verification step; the version-bump obligation is exactly the kind of thing that gets half-done. |
| `bead-bonanza:425` | "On gate failure, fix the root cause and rerun. After two full failed cycles, follow the Blocked path" — this one *is* a proper bounded loop. Good; cite it as the in-repo pattern for the others. |

**Recommended change.** Standardize one sentence and use it at all five
sites: *"Fix what it reports and re-run; do not proceed while it exits
non-zero. After two full failed cycles, stop and surface."* The bound
matters — BP's loops are unbounded, but `bead-bonanza:425` and
`retro-skill`'s bar both show this repo prefers a bounded loop with an
explicit give-up path, which is the better pattern for autonomous runs.

For `groom-beads` phase 6 specifically, state the gate in terms of the
script's contract: *"`bd-health.sh` exits 1 on hard problems (cycles,
orphans). Do not declare the set ready while it exits 1."*

---

## P14. Multi-phase skills have no copyable progress checklist

**Severity: medium.** BP:
[Use workflows for complex tasks](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#use-workflows-for-complex-tasks):

> "Break complex operations into clear, sequential steps. For
> particularly complex workflows, **provide a checklist that Claude can
> copy into its response and check off as it progresses.**" … "Clear
> steps prevent Claude from skipping critical validation. The checklist
> helps both Claude and you track progress through multistep workflows."

The BP doc gives two worked examples, one for code-free skills and one
for script-bearing skills — both shapes are present in this repo.

Candidates, by step count:

| Skill | Structure | Checklist value |
|---|---|---|
| `bead-bonanza` | Phase 0 (7 steps) → 1 → 1.5 (4) → 2 (3) → 2.5 → 3 → 4 (7) | Highest — long autonomous run, easy to skip Phase 1.5's gate |
| `groom-beads` | 7 numbered phases | High — phase 6 validation is the skippable one |
| `triage-rfc` | 7 steps with a hard human gate at 4 | High — the gate is the whole design |
| `groom-feature-blueprint` | 3 phases + finishing pass | Medium |
| `retro-skill` / `retro-spec` | 7-step workflows each | Medium — step 6's null-result path is easy to skip when padding feels easier |

**Recommended change.** Add a fenced, copyable block at the top of each
workflow. For `groom-beads`:

```
Grooming progress:
- [ ] 0. Probed `bd` CLI + loaded bd prime / remembered insights
- [ ] 1. Scoped the grooming set; showed the user the plan
- [ ] 2. Every bead driven to the spec-owned Definition of Ready
- [ ] 3. Dependency graph wired (edges justified, none fabricated)
- [ ] 4. Model + model_reason on every bead (never null)
- [ ] 5. C4 blast radius tagged (or explicitly skipped: no C4 docs)
- [ ] 6. Validated — 0 errors, re-run after each fix
- [ ] 7. Reported: changes, still-not-ready + gaps, flags for the human
```

Note this interacts with `AGENTS.md:66` ("do NOT use TodoWrite,
TaskCreate"). A copied-into-the-response markdown checklist is not a task
tool, so it is compatible — and it is the *better* fit here, since it
gives the same visibility without the banned tooling. Worth saying so
explicitly at the checklist so an agent does not resolve the apparent
conflict by dropping the checklist.

---

# Tier 5 — Content-guideline polish

## P15. Changelog and provenance narration inside instructions

**Severity: low, but it accumulates.** BP:
[Avoid time-sensitive information](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#avoid-time-sensitive-information):

> "Don't include information that will become outdated"

with the sanctioned home being a `<details>`-wrapped "Old patterns"
section: "The old patterns section provides historical context without
cluttering the main content."

Instances:

| Site | Text | Problem |
|---|---|---|
| `bead-bonanza:244` | "**The previous "≥3" was too conservative.** Even at 2 beads the savings are real" | Argues with a version the reader never saw. State the rule; the rationale for the *current* threshold is what earns tokens. |
| `bead-bonanza:336` | "(Origin: a near-miss where a subagent claimed a skip while leaving a non-empty diff.)" | Incident provenance. Belongs in the RFC/retro log this repo already maintains. |
| `triage-rfc:56-58` | "rfc-0001 through rfc-0003 predate the term and use the older heading "Disposition"; check both" | Textbook time-sensitive: true today, wrong once those three are re-headed or archived. Move into an "Old patterns" `<details>`. |
| `groom-feature-blueprint:235-241` | HTML comment: "deliberately not yet defined; when it is, it supersedes this shape" | Roadmap in an instruction file. Keep the *current* rule ("model/C4 fields are absent on this path"); drop the futures. |
| `spec/SKILL.md:47` | "**Ten** specification document types are currently defined: three for … five for architecture … one … and one" | A volatile count directly below the authoritative table. |

`spec:47` is the sharpest one because it violates this repo's own
standard — `template-authoring-standards.md:666-680`, "Avoid Volatile
Facts":

> "Documents SHOULD NOT state volatile facts — **element counts**, line
> numbers, dependency version numbers, file sizes… **Rationale:**
> Volatile precision reads as rigor but rots immediately: the count
> changes and the document silently becomes wrong."

**Recommended change.** Delete the counts and the changelog asides; move
`triage-rfc`'s legacy-heading note into an `<details><summary>Old
patterns</summary>` block. Add "no counts of things the document
enumerates" to the CI check as a lint on the skills, since the standard
already demands it of the templates.

---

## P16. Naming is inconsistent across the collection, and one name is vague

**Severity: low mechanically, medium for discovery.** BP:
[Naming conventions](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#naming-conventions):

> "Consider using **gerund form** (verb + -ing)… **Acceptable
> alternatives:** Noun phrases… Action-oriented…
> **Avoid:** Vague names: `helper`, `utils`, `tools`; Overly generic:
> `documents`, `data`, `files`; **Inconsistent patterns within your skill
> collection.**"
>
> "Consistent naming makes it easier to… Maintain a professional,
> cohesive skill library."

Four patterns across seven skills:

| Pattern | Skills |
|---|---|
| Verb-object | `groom-beads`, `groom-feature-blueprint`, `triage-rfc` |
| Noun-noun | `retro-spec`, `retro-skill` |
| Bare generic noun | `spec` |
| Alliterative pun | `bead-bonanza` |

`bead-bonanza` is the clear miss: it names the mood rather than the
activity, and gives the discovery layer nothing to match on. Its
description carries the load alone — which works, but wastes the name.
`spec` is generic in the sense BP flags, though it is at least the
domain's real term of art.

All names pass the hard rules: ≤64 chars, lowercase/hyphens only, no
reserved words, no XML.

**Recommended change — deliberately conservative, because renaming is
expensive here.** A rename breaks installed copies, every `../<skill>/`
path (P9), and the `skill:` provenance field that
`retro-skill/references/skill-rfc-template.md:12` records in past RFCs.
So:

1. **Rename `bead-bonanza`** → `processing-bead-queues` (gerund) or
   `batch-process-beads` (action-oriented). Do it now, while it is the
   least-referenced skill.
2. **Keep `spec`.** It is established, its description is
   trigger-rich, and the cost of churning it exceeds the benefit.
3. **Adopt one pattern for anything new** and record it (see plan2 §6
   on a skill-authoring standard). Verb-object matches the plurality and
   reads well for this domain.
4. If any rename lands, add an alias note to `triage-rfc` so RFCs citing
   the old `skill:` value still dedupe correctly.

---

## P17. `groom-beads`'s description is at 94% of the character limit

**Severity: low.** Measured description lengths:

| Skill | Chars | Words |
|---|---|---|
| `groom-beads` | **960** | 151 |
| `spec` | 803 | 102 |
| `retro-spec` | 625 | 97 |
| `bead-bonanza` | 595 | 95 |
| `retro-skill` | 564 | 92 |
| `groom-feature-blueprint` | 407 | 63 |
| `triage-rfc` | 402 | 57 |

All are under the 1,024 limit and all are correctly third-person — no
"I can help you" / "You can use this" anywhere, which is the most common
failure BP warns about.

`groom-beads` is the outlier. Its *trigger* half is excellent and should
be untouched — it is exactly what BP asks for ("Be specific and include
key terms"): `bd ready`, `.beads/`, "groom these", "make these beads
ready", "running beads in waves", "assigning models to beads". But it
also carries a mechanism inventory that duplicates the body: "Brings each
bead to the Definition of Ready, wires the dependency graph, tags C4
container/component blast radius, assigns an execution model, and
(optionally) previews collision-free waves."

BP: "Your description must provide enough detail for Claude to know when
to select this Skill, **while the rest of SKILL.md provides the
implementation details.**"

**Recommended change.** Cut the mechanism list; keep every trigger term.
Target ~550 chars, in line with its siblings. This buys headroom for
future triggers without a rewrite. Minor consistency note: `groom-beads`
is the only skill using a YAML `>-` block scalar for `description` while
the others use plain multi-line — harmless, but worth unifying if
anything ever parses these mechanically.

---

## P18. `python` vs `python3`, and an unstated optional dependency

**Severity: low but easy.** `groom-beads/SKILL.md:222-224` prescribes:

```
python scripts/validate_beads.py <emitted.jsonl> --c4-docs <path/to/c4/docs...>
```

while the same file's Reference files section (`:298-300`) says the
script "Requires `python3`; uses `jsonschema` if installed and falls back
to built-in checks otherwise." On many modern systems `python` does not
exist on `PATH`.

**Best practice violated.** [Avoid assuming tools are installed](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#avoid-assuming-tools-are-installed)
and [Package dependencies](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#package-dependencies)
("List required packages in your SKILL.md"). Checklist: "Required
packages listed in instructions and verified as available".

**Recommended change.** Use `python3` at the call site, and move the
dependency note *to* the call site rather than 75 lines later:

```
python3 scripts/validate_beads.py <emitted.jsonl> --c4-docs <path/to/c4/docs...>
```
> Requires `python3`. Uses the `jsonschema` package for full structural
> validation when installed; otherwise falls back to equivalent built-in
> checks — the script never fails for a missing dependency.

Credit where due: `validate_beads.py:96-117` genuinely implements that
graceful degradation, which is the "Solve, don't defer" pattern done
right. The SKILL.md just does not say so where it matters.

---

## P19. Line length: `bead-bonanza` diverges from the repo's own 80-column rule

**Severity: cosmetic; listed for internal consistency.**
`template-authoring-standards.md:24` — "The width of any line SHALL NOT
exceed 80 characters" (with exemptions for URLs, code-block content and
table cells) — and `triage-rfc/SKILL.md:80` instructs keeping every
touched file "within its own conventions (80-column lines, EARS for
normative statements)".

Measured, raw counts (these include exempt tables and code blocks, so
treat as an upper bound):

| Skill | Lines >80 | Longest line |
|---|---|---|
| `groom-feature-blueprint` | **0** | 79 |
| `triage-rfc` | 7 | 167 |
| `groom-beads` | 34 | 154 |
| `spec` | 47 | 246 |
| `bead-bonanza` | **207** | **1,070** |

`groom-feature-blueprint` proves the standard is achievable in a SKILL.md
without loss. `bead-bonanza`'s 1,070-character line (`:13`) is a genuine
diff and review problem independent of any rule.

**Recommended change.** Reflow prose lines in `bead-bonanza` to 80
columns as part of the P6 split (they are the same edit), leaving tables
and fenced blocks alone. Then decide explicitly whether the 80-col rule
governs SKILL.md files — if yes, say so in the skill-authoring standard
(plan2 §6) and lint it; if no, narrow the standard's scope so
`triage-rfc:80` is not asking for something the repo does not want.

---

# Appendix A — BP-doc checklist, scored

From [Checklist for effective Skills](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#checklist-for-effective-skills).

### Core quality

| Item | Verdict |
|---|---|
| Description is specific and includes key terms | **Pass** — best-in-class; `groom-beads`'s trigger list is a model |
| Description includes what *and* when | **Pass** — all 7 |
| SKILL.md body under 500 lines | **Fail (1/7)** — `bead-bonanza` 615 → P6 |
| Additional details in separate files | **Partial** — `spec` exemplary (17 files); `bead-bonanza` 0 → P6 |
| No time-sensitive information | **Fail** — 5 sites → P15 |
| Consistent terminology throughout | **Fail** — model vocabulary contradicts across 5 files → P1; "work order" vs "work item" vs "bead" → plan2 §7 |
| Examples are concrete, not abstract | **Partial** — `work-order-definition-of-ready.md` excellent; `bead-bonanza`/`retro-skill` thin → P11 |
| File references are one level deep | **Fail** — two-hop chains → P8; `../` escapes → P9; wrong-root paths → P4 |
| Progressive disclosure used appropriately | **Partial** — strong in `spec`, absent in `bead-bonanza` → P6, P7 |
| Workflows have clear steps | **Pass** — phase structure is a genuine strength; no copyable checklist → P14 |

### Code and scripts

| Item | Verdict |
|---|---|
| Scripts solve problems rather than defer | **Partial** — `validate_beads.py` good; `bd-health.sh` silently skips its hard check → P2 |
| Error handling explicit and helpful | **Partial** — `validate_beads.py` yes; `bd-health.sh` no → P2 |
| No voodoo constants | **Pass** — thresholds are justified in prose (batch ≥2, 5-min Monitor, 3-strike breaker, ~1hr sizing) |
| Required packages listed and verified | **Partial** — → P18 |
| Scripts have clear documentation | **Pass** — both scripts have strong header blocks with usage and exit codes |
| No Windows-style paths | **Pass** — zero found |
| Validation steps for critical operations | **Partial** — → P13 |
| Feedback loops for quality-critical tasks | **Partial** — → P13 |

### Testing

| Item | Verdict |
|---|---|
| At least three evaluations created | **Fail** — zero for all 7 skills → P10 |
| Tested with Haiku, Sonnet, Opus | **Fail** — no evidence; skills dispatch to cheap tiers by design → P11 |
| Tested with real usage scenarios | **Partial** — `retro-skill:209-212` and `bead-bonanza:336` show real dogfooding happened; it is just not captured as repeatable tests |
| Team feedback incorporated | **Pass by design** — the retro→RFC→triage pipeline *is* this mechanism (but see plan2 §3: it has never run) |

### Also checked, clean

- `name` field: all ≤64 chars, lowercase/numbers/hyphens only, no XML, no
  reserved words (`anthropic`/`claude`).
- `description`: all non-empty, all ≤1,024, all third person.
- No Windows paths, no `@`-sigil file mentions in skill bodies.
- No orphaned bundled files — every `references/` and `scripts/` file is
  cited from its owning SKILL.md.
- No MCP tool references, so BP's fully-qualified-name rule does not
  apply. Harness tool names (`Agent`, `Monitor`, `TaskCreate`,
  `AskUserQuestion`) are used correctly and unqualified, as they should
  be.

---

# Appendix B — What is already exemplary

Do not regress these while fixing the above.

1. **Descriptions.** Third-person, specific, trigger-dense, and they
   state both what and when. `groom-beads`'s "Trigger even if the user
   just says 'groom these'… in a repo that uses beads" is precisely BP's
   "Be specific and include key terms".

2. **Single-source-of-truth discipline.** `groom-beads:45-57`,
   `groom-feature-blueprint:32-46`, `blueprint-readiness-checklist:15-20`
   all name their canonical owner and say "do not paraphrase or fork". P5
   shows what the alternative costs.

3. **Anti-pattern tables.** Every skill ends with a Pattern/Problem
   table. These are concrete, non-obvious, and each entry names a real
   failure. `retro-spec:161-173` and `spec:264-286` are the strongest.

4. **The null result as a first-class outcome.** `retro-skill:143` ("A
   run that surfaced nothing that clears the bar is a real, valuable
   result") and `retro-spec:139-143` build in a defense against
   padding-to-justify-the-artifact. Rare and correct.

5. **Bounded retry with an explicit give-up path.** `bead-bonanza:425`
   ("After two full failed cycles, follow the Blocked path — do not keep
   retrying") and the 3-strike circuit breaker at `:341`. BP's feedback
   loops are unbounded; this is the better autonomous-run pattern.

6. **The lean path.** `bead-bonanza:157-161` and `:213-226` let the skill
   skip its own ceremony when the queue does not earn it. This is BP's
   [degrees of freedom](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#set-appropriate-degrees-of-freedom)
   principle applied self-referentially, and it is the single best
   token-economy decision in the repo.

7. **Vendoring for zero-dependency consumers.**
   `groom-feature-blueprint:181-188` copies the findings protocol into the
   target project so an implementing agent needs one repo-local file and
   no plugin. That is a genuinely sophisticated progressive-disclosure
   move — and the template for fixing P9.

8. **Worked before/after examples.**
   `work-order-definition-of-ready.md:106-172` gives three
   under-groomed→ready transformations. This is BP's
   [Examples pattern](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices#examples-pattern)
   done properly, and it is the repo's best affordance for smaller
   models.

9. **Graceful degradation in scripts.** `validate_beads.py:96-117` uses
   `jsonschema` when present and mirrors the same rules in a built-in
   checker when not, so the script never fails for a missing dependency.
   `_closest()`'s Levenshtein suggestion for a mistyped C4 component
   ("did you mean…?") is exactly BP's "make validation scripts verbose
   with specific error messages".

10. **Explicit human gates.** `triage-rfc:74-77` ("Present the
    recommendation and STOP for approval. Implement nothing before the
    human approves") and `groom-feature-blueprint:80-82` ("Never resolve
    a gap by choosing an answer yourself, however obvious it seems"). Low
    freedom exactly where the cliffs are — BP's narrow-bridge case.

---

## Suggested sequencing

**Session 1 — unblock (half a day).** P1, P2, P3. Three defects, all
narrow, all with a known-good patch. Nothing else is worth doing while a
groomed bead can fail its own validator.

**Session 2 — measure (one day).** P10 for `groom-beads` and `spec`
only — six evals. This gives a regression net before restructuring, and
will surface Tier 1 bugs of its own.

**Session 3 — restructure (one day).** P6 (split `bead-bonanza`, which
subsumes P19), P7 (TOCs — mostly generated), P4 (path fix, mechanical).

**Session 4 — harden (one day).** P12 (three `bead-bonanza` scripts),
P13 (close the loops), P14 (checklists).

**Session 5 — polish and decide.** P5, P8, P9, P15–P18. P9 and P16 carry
real decisions; take them to the human rather than picking.

Then adopt plan2 §2 (CI consistency check), which is what stops Tier 1
and Tier 5 from silently recurring.

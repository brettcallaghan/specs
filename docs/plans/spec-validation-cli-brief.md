# Brief: deterministic validation tooling for the spec document set

Research and options for review. Not a spec — the input to one. Written
against the working tree at `dec1a33`.

---

## Recommendations at a glance

| Question | Recommendation | Confidence |
|----------|----------------|------------|
| Scripts or CLI? | **One CLI, many subcommands**, with narrow flags so a skill can invoke exactly one check | High |
| Language? | **Python**, single-file commands with PEP 723 inline deps, run via `uv run` with a plain-`python3` fallback | Medium-high |
| Bundling? | **Ship inside the plugin.** `${CLAUDE_SKILL_DIR}` + `allowed-tools` runs it with no permission prompt; no separate distribution needed | High |
| Cross-platform? | **Yes**, with named caveats (§8). The constraint is real and cheap to hold if designed in now | High |
| Primary artifact? | **The frontmatter index**, not the pass/fail report. Every check you listed is a query over it | High |
| Scope split? | **Two layers in one binary**: a generic document layer (config-driven) and a framework layer (built-in rules) | Medium |
| FRD? | The CLI needs one. The standards docs **stay standards** — the FRD complies by reference | High |

The one thing I'd change about the brief as you framed it: you listed
checks, but the checks are downstream. Build the index first and most of
the list falls out as queries (§3).

---

## 1. What I found when I actually ran it

I prototyped the core checks over this repo's own corpus — 50 markdown
files under `docs/`, `rfcs/`, and `plugin/skills/*/references/`, 42 of
them carrying a slug. Results, because "would this catch anything?" is a
better question than "is this a good idea?":

| Check | Fired | Verdict |
|-------|-------|---------|
| Dangling `[[slug]]` | 26 real | **Genuine live defect** — filed `agent-plugins-5nz` |
| Frontmatter parses | 1 failure | **Genuine** — filed `agent-plugins-6n4` |
| Slug uniqueness | 0 | Clean |
| Required properties | 0 | Clean |
| Status vocabulary | 0 | Clean |
| Dangling frontmatter refs | 2 | Minor — `related: work-order-template` points at a template, which has no slug |
| Line width > 80 | 14 | Mostly real, all trivial |
| Compound `SHALL … and … SHALL` | 18 raw | ~6 real, rest false positives (§5) |
| Leftover `<!--` | 14 | All templates — correct behaviour, wrong corpus (§4) |

Three things worth knowing before we design anything:

**The ADR defect is the interesting one.** Eleven of thirteen ADRs carry
a descriptive slug (`adr-011-targeted-verification-gates`) while every
citation in the repo writes the bare number (`[[adr-011]]`). The prefix
registry in `template-authoring-standards.md` says `adr-<nnn>`, so the
slugs are wrong and the links are right. This is exactly the class of
error the framework's philosophy says gates should catch, sitting
undetected in the repo that authored the philosophy. It also isn't
visible to a human reader — you only see it if you resolve every link.

**There is no CI.** `.github/workflows` does not exist. Two runnable
gates already ship (`scripts/check-skills.sh`,
`scripts/sync-shared-references.sh --check`) and nothing invokes them.
`skill-authoring-standards.md` states "Every invariant a skill asserts
about itself SHALL be enforced by an automated check — Status in this
repository: not yet met." Filed as `agent-plugins-bgy`; worth doing
before the CLI exists, since it's the harness the CLI slots into.

**A corpus model is a hard requirement, not a nicety.** Of 151 raw
unresolved wikilinks, 49 vanish once fenced and inline code spans are
excluded, and nearly all of the remainder are placeholders in templates
(`[[container-<name>]]`) or illustrations in guides
(`[[frd-checkout]]`). Templates are *not documents*: their frontmatter
is placeholder text, their instruction blocks are supposed to survive,
their slugs are `<angle-bracket>` forms. A validator that doesn't model
this reports ~100 findings on a clean repo and gets switched off in a
week. This is the single most important design decision in the whole
thing and it isn't in your list.

---

## 2. Does this make skills faster and more reliable?

Faster: yes, measurably. Reliable: yes, but only with skill-side wiring.

Count the mechanical share of the two checklists the skills already run
by reading:

- `document-readiness-checklist.md` — 8 checks. **6 are fully
  mechanical** (1 frontmatter, 2 slug rule, 3 status vocabulary, 6 line
  width, 7 instruction blocks, 8 link resolution). 4 and 5 are semantic,
  and 5 is partly mechanical.
- `blueprint-readiness-checklist.md` — 10 checks. **6 are mechanical**
  (1, 2, 3, 5, 8, 10), 4 is partly, and 6/7/9 are judgment.

Today an agent runs check 8 by globbing the corpus and reading every
candidate. On this repo that's ~42 file reads to answer one question the
index answers in one call. The savings compound because both skills run
their checklist, fix, and **re-run** it.

The reliability argument is stronger than the speed one, and it's
already written down here. `skill-authoring-standards.md` records that a
fabricated verification claim reached a pull-request description in this
repo — the artifacts were gated hard and the *report* not at all. A
validator's output is a citable artifact: "the checklist passed" becomes
a command and its output rather than an assertion.

The caveat, also already written down: a validator whose output is
advisory is decoration. The same standard requires the skill to fix and
re-run, refuse to proceed while failing, and bound the loop. **So the
deliverable is CLI + skill wiring, never the CLI alone.** If we ship the
binary and leave the skills to notice it, we've built a nicer version of
the thing nobody runs.

---

## 3. The reframe: the index is the product

`template-authoring-standards.md` §Frontmatter Index already specifies
this and nobody has built it:

> A document system SHOULD maintain a machine-readable index of all
> documents' frontmatter (a JSON object written to disk). The index is
> the authority for slug-uniqueness checks, next-number derivation for
> numbered types, `[[slug]]` reference validation, and dangling-reference
> detection after a promotion or reparent.

Your item "produce a JSON file that contains the frontmatter in all
files" *is* this, and it's not one item among nine — it's the substrate
the other eight run on. Slug uniqueness is a `GROUP BY` over the index.
Link validation is a join. Next-ADR-number is a `MAX`. Impact analysis
is a reverse edge walk.

It also unlocks things that aren't validation at all, which is where I
think the real leverage is:

- **`next-slug adr`** — today, authoring an ADR means reading every ADR
  to find the next free number. One call instead.
- **`slug-exists <slug>`** — answers your "validate that a slug isn't
  already in use" *before* the document is written, which is when the
  answer is useful.
- **`impact <slug>`** — what references this? The `retro-spec` and
  `triage-rfc` skills both need this and currently grep for it.
- **Injected context.** Skills support `` !`command` `` injection, which
  runs before the agent sees the skill body. The `spec` skill could open
  with the live slug registry already inlined. That's a meaningful
  reliability change: the agent stops *deriving* the namespace.

Design consequence: `index` must be fast and must not require the corpus
to be valid. It parses what it can, records what it couldn't, and never
fails. Validation is a separate consumer that may fail.

---

## 4. Catalogue of deterministic checks

Split by who they're useful to — that's your "universal vs. geared to
our skills" question. **U** = any markdown/Obsidian spec repo. **F** =
this framework specifically.

### Already in your list

| # | Check | Layer | Owner doc | Notes |
|---|-------|-------|-----------|-------|
| 1 | Frontmatter present + parses | U | `frontmatter-standards` | Must validate the **raw string**, not the parsed value — PyYAML silently coerces `updated_at` to a datetime, so a bare `2026-07-29` passes an RFC 3339 check against the parsed object. My probe hit this: 42 false positives. |
| 2 | Required properties per type | U | `frontmatter-standards` | JSON Schema per type, `allOf`-merged with the base record schema as the standard already prescribes |
| 3 | `[[slug]]` resolves | U | `cross-document-references` | Report file + line + column span |
| 4 | `[[slug#Anchor]]` heading exists | U | `cross-document-references` | Needs a heading index; slug-ify anchors consistently |
| 5 | Slug uniqueness | U | `template-authoring-standards` | Trivially an index query |
| 6 | Frontmatter index emission | U | `template-authoring-standards` | §3 — the substrate |
| 7 | Mermaid blocks parse | U | `c4-guide` | The awkward one; see §7 |
| 8 | `component` record blocks validate | F | `frontmatter-standards`, ADR-008 | Also enforce the **record-block XOR blueprint-link** rule — a child is in exactly one state, and nothing checks it today |

### Additional opportunities

Ordered roughly by value per unit of effort.

| # | Check | Layer | Why it's worth it |
|---|-------|-------|-------------------|
| 9 | **AC traceability coverage** | F | `ac-traceability-standards.md` *literally specifies this tool*: report uncovered ACs (in an active FRD, cited by no test) and orphaned citations (citing an ID no approved/active FRD defines), and fail the build on the former. The regex is given in the doc. Highest-value unbuilt item in the repo. |
| 10 | **Bidirectional relationship consistency** | U | If A declares `children: [B]`, B should declare `parent_slug: A`. Same for `blocks`/`depends_on`. Pure graph check, catches drift no human sees. |
| 11 | **Cycles and orphans** in the relationship graph | U | Specified in `template-authoring-standards` §Dependency Graph, unimplemented |
| 12 | **`updated_at` staleness** | U | `git log -1 --format=%cI -- <file>` newer than frontmatter `updated_at` ⇒ the author forgot to bump it. Fully deterministic, catches a rule everyone breaks. |
| 13 | **Version-bump enforcement** | U | `template-authoring-standards` mandates "any normative change bumps the version in the same commit". In CI, diff against the base ref: if a template/skill/standard body changed and `version` didn't, fail. Nothing enforces this today. |
| 14 | **AC Coverage legend ↔ delta keys** | F | Blueprint checklist #8: every delta has a unique short key, every legend key expands to a real delta block, no retired key reused. Table + heading parse. |
| 15 | **Change-document completion** | F | A feature blueprint's `updates`/`creates`: every listed slug exists, and everything in `creates` is status `draft`. Half of blueprint checklist #5. |
| 16 | **Retired/superseded integrity** | U | `status: retired` ⇒ `superseded_by` present and resolving; flag live documents still pointing at retired ones |
| 17 | **Placeholder leakage** | U | `<Feature Name>`, `${CURRENT_DATE}`, `TODO`, `TBD`, `<!--` surviving into a document at `ready-for-review` or beyond. Cheap; catches the embarrassing failures. |
| 18 | **Template section conformance** | U | A document declaring `type: adr` has the ADR template's required headings. Catches "author skipped a section" for near-zero cost. |
| 19 | **Work-order Definition of Ready** | F | 8 named checks; roughly 5 are mechanical (design section present, acceptance present, dependencies wired, tier assigned + reason, C4 blast radius). `validate_beads.py` already does much of this — fold it in rather than duplicating. |
| 20 | **Fold in the existing gates** | F | `check-skills.sh` and `sync-shared-references.sh --check` become subcommands. One gate, one severity model, one exit contract — the framework's own "one term for one concept" rule applied to tooling. |

### What should stay agent judgment

Worth stating explicitly so the FRD's non-goals are honest:

- **Is a requirement well-formed EARS?** Detecting a missing trigger is
  mechanical; deciding whether the behaviour is *truly ubiquitous* is
  not. Ship this as a warning-level lint, never a gate.
- **Blueprint check 7** (does a delta carry enough design intent?) —
  irreducibly semantic, and the checklist says so.
- **Tracker reconciliation classification** (subsumed / prerequisite /
  out-of-scope) — the *presence* of the section is checkable, the
  classification isn't.
- **"Is this the right document type?"** — the skill's job.

---

## 5. The EARS lint is the hard one — and it argues for severities

My compound-modal check fired 18 times. Genuine hits like the contract
template's "it SHALL remain available for `<window>`, and the provider
SHALL announce…" sit next to false positives like
`ac-traceability-standards`'s "It is SHOULD, not SHALL, because the
blueprint retires…" — prose *about* requirements, not a requirement.

The rule as written ("a sentence carrying a capitalized SHALL, SHOULD or
MAY is a requirement") is a good instruction for a human and a bad
predicate for a parser. Options:

1. **Scope by section** — only lint under Requirements/Acceptance
   headings. Under-covers standards docs, where every section is
   normative.
2. **Warning severity, never a gate.** Cheap, honest, and matches how
   the checklist already treats 4 and 5 differently from 1–3.
3. **Structural markup** — require normative statements to be list items
   or bolded. A real standards change; would make the lint exact.

I'd start with (2) and revisit (3) if the noise is annoying. Worth
deciding tomorrow, because it's the difference between a gate people
trust and one they `--no-verify` past.

**Implication:** the CLI needs three severities — `error` (gates),
`warn` (advises), `unverified` (couldn't run). The third is not
optional: `skill-authoring-standards` requires that "IF a script cannot
run a check it was asked to perform, THEN it SHALL report that check as
unverified rather than as passed." That's the mermaid case exactly (§7).

---

## 6. Tech: one CLI, and how it gets to the agent

### Scripts vs. CLI

**One CLI.** Ten scripts each re-globbing and re-parsing the corpus is
slow, and worse, they'd disagree — different exclusion rules, different
exit-code meanings, different output shapes. One binary gives one corpus
model, one config resolution, one severity vocabulary.

But expose it narrowly, because a skill needs *one* check, not all of
them:

```
sdd index [--out .spec-index.json]     # the substrate; never fails
sdd check [PATH...] [--only links,frontmatter] [--json] [--strict]
sdd next-slug <type>                   # max+1 for numbered types
sdd slug-exists <slug>
sdd impact <slug>                      # reverse references
sdd coverage --frds docs/ --tests src/ # AC traceability (§4 #9)
sdd graph [--format mermaid|dot|json]
```

`--json` matters more than it looks: it's the difference between an
agent parsing a report and an agent reading one.

Name is a placeholder — `sdd` after the plugin. Open question.

### Bundling — this is a solved problem, better than I expected

You do **not** need a separate distribution path. Three mechanisms:

1. **`${CLAUDE_SKILL_DIR}` + `allowed-tools`.** A skill can reference
   its own bundled script by absolute path and pre-approve it:

   ```yaml
   ---
   name: spec
   allowed-tools: Bash(${CLAUDE_SKILL_DIR}/scripts/sdd *)
   ---
   Run `${CLAUDE_SKILL_DIR}/scripts/sdd check --json <path>`.
   ```

   Both occurrences expand to the same path, so the rule matches the
   command the body prescribes and it runs **without a permission
   prompt**. (Requires Claude Code ≥ 2.1.129; older versions leave the
   literal string and it prompts — degraded, not broken.)

2. **Plugin hooks.** `hooks/hooks.json` in the plugin root can register
   a `PostToolUse` hook matching `Write|Edit`, invoked as
   `"${CLAUDE_PLUGIN_ROOT}/scripts/sdd"`. That makes validation
   *ambient* — a spec document gets checked when it's written, whether
   or not the skill remembered to. A `PostToolUse` hook can't block the
   write (it fires after), but it can return `additionalContext` or
   `decision: "block"` to stop the agent proceeding. This is the
   strongest reliability mechanism available and it needs no skill
   discipline at all.

3. **Skill-scoped hooks.** A skill's frontmatter can declare `hooks:`
   scoped to its own lifetime — so the write-time validation is active
   only while `spec` is running, and cleans itself up. Probably the
   right middle ground: ambient enough to be reliable, narrow enough
   not to fire on every unrelated markdown edit.

Plus `${CLAUDE_PLUGIN_DATA}` (`~/.claude/plugins/data/<id>/`), which
persists across plugin updates — the documented place to install
language dependencies once, with a `SessionStart` hook that diffs the
bundled manifest against the installed copy and reinstalls on change.
That's our escape hatch if we ever need a real dependency tree.

My read: **(1) for the deliberate checklist runs, (3) for a write-time
safety net.** Start with (1); (3) is a follow-on once the rules are
stable enough not to be noisy.

### Language

| Option | For | Against |
|--------|-----|---------|
| **Python + PEP 723 / `uv`** | Matches existing tooling (`validate_beads.py`, embedded Python in `check-skills.sh`); inline dep metadata means `uv run sdd` resolves PyYAML/jsonschema hermetically with no build step in a git-cloned plugin; agents can read and patch it | Needs `uv` or a `python3` with deps; `uv` isn't universally present yet |
| **Node/TypeScript** | Mermaid's real parser is JS — the only route to grammar-accurate diagram validation; `npx` gives zero-install | Needs `node_modules`; heavier; diverges from everything already here |
| **Go / Rust** | Single static binary, no runtime, fastest, trivially cross-platform | Must commit per-platform binaries into a git-distributed plugin, or require a toolchain; agents can't patch it mid-session |

**Recommendation: Python.** The work is text, YAML, and graphs — no
performance pressure — and the decisive factor is that a plugin is
distributed as a git clone with no build step, which makes shipping
compiled artifacts genuinely awkward. Python also keeps the tool
*legible to the agents running it*, which matters for a framework whose
whole thesis is agent-followable process.

The fallback ladder: try `uv run` → fall back to `python3` with a
dependency preflight → if deps are missing, report every affected check
as **unverified** and exit non-zero with the install command. Never
silently skip.

---

## 7. Mermaid: the one genuine carve-out

Mermaid has no dependency-free parser. The upstream library needs a DOM
even to *parse*, and jsdom/happy-dom are reportedly not quite enough for
all diagram types. Realistic options:

- **`mermaid-validate`** (JS, official parser + jsdom) — accurate,
  CLI-shaped, but pulls in Node
- **`mermaid-check`** (pure Go, published Feb 2026) — no browser, but a
  reimplemented grammar, so it will drift from upstream
- **Headless browser via `mmdc`** — most accurate, far too heavy for a
  gate

**Recommendation:** treat mermaid as an *optional* check. The CLI does
its own cheap structural pass always (fence is closed, first token is a
known diagram type, `-->`/`->` arrows are balanced, C4 naming rules from
`c4-guide.md` — no internal acronyms, arrow direction). If a real
mermaid validator is on `PATH`, shell out for grammar accuracy. If not,
report `unverified` — which the standard already demands and which is
honest, unlike the alternative of a structural pass masquerading as
validation.

This also keeps the hard dependency out of the default install, which is
what makes the Python recommendation hold.

---

## 8. Cross-platform: yes, with named caveats

Linux/macOS/Windows is achievable and cheap *if designed in now*. The
things that actually bite:

- **CRLF.** Windows checkouts will have them. Line-width and span
  arithmetic break silently. Normalize on read; never trust `len(line)`
  on raw bytes.
- **Unicode width ≠ character count.** The repo uses em dashes heavily.
  Decide now whether "80 characters" means codepoints or display
  columns, and write it down — currently ambiguous in
  `template-authoring-standards`.
- **Case-insensitive filesystems.** macOS and Windows will happily hold
  `Frd-Checkout.md` and `frd-checkout.md` as one file. Slug uniqueness
  must be case-folded.
- **Shell scripts don't run on Windows** without Git Bash. Our current
  gates are all `.sh`. Hook commands should invoke `python`/`uv`
  directly, not a `.sh` wrapper. (Skills do support `shell: powershell`
  for inline injection, but that's a per-skill setting, not a fix.)
- **Path separators.** `pathlib` internally, POSIX-style paths in JSON
  output, so reports are diffable across machines.

CI matrix should prove this rather than assert it: ubuntu × macos ×
windows, on the minimum supported Python and current.

---

## 9. Testing and CI

You asked for full coverage. Concretely:

**Fixture corpus.** `tests/fixtures/<rule-id>/` — one deliberately
broken document per rule, plus a clean control. Snapshot the `--json`
report. This is the only way to keep 20+ rules honest, and it doubles as
the rules' documentation.

**Meta-test.** Assert every registered rule ID has at least one
fixture and at least one passing control. Prevents a rule shipping
untested — the exact failure mode `skill-authoring-standards` complains
about.

**Property tests.** Slug parsing, link extraction, and code-span
stripping against generated input. The code-span stripper is the
highest-risk component (it decides what's a real link) and the easiest
to get subtly wrong.

**Self-application as an integration test.** Run the CLI against this
repo's own `docs/` in CI. Real data, and it dogfoods the framework —
today it would fail on the ADR slugs, which is the point.

**Golden-path regression.** A frozen sample project (one FRD, one
blueprint, one contract, one container) that must stay green.

**CI shape**, given none exists today:

```
lint      ruff + format check
test      pytest, ubuntu/macos/windows × 2 Python versions, coverage gate
self      sdd check docs/ rfcs/ plugin/ --strict
legacy    check-skills.sh --strict; sync-shared-references.sh --check
```

Wire `legacy` first — it's independently valuable and needs none of this
(`agent-plugins-bgy`).

---

## 10. Will standards docs turn into FRDs?

No — and the distinction is worth getting right, because it's a test of
the framework's own taxonomy.

The standards docs state invariants that hold across features. By the
`spec` skill's own routing question 4, that is exactly Standards, and
building a tool that enforces them doesn't change their type. The CLI is
a feature: it gets an **FRD** (the WHY and the EARS acceptance criteria)
and a **feature blueprint** (the HOW), and the FRD complies with the
standards *by reference* — `[[frontmatter-standards]]`,
`[[cross-document-references]]` — never by restating them. That's the
framework's own anti-duplication rule.

Two real consequences, though:

1. **Several standards will tighten.** They currently hedge in
   tooling-conditional form: "WHEN a document system implements
   validation tooling, JSON Schema definitions SHALL be created…",
   "tools SHOULD validate that all `[[document-slug]]` references point
   to existing documents." Once the tooling exists, those SHOULDs become
   SHALLs, and the conditional WHENs lose their condition. That's a
   handful of one-line edits, and each one bumps a version.

2. **This is the first end-to-end dogfood.** The repo has thirteen ADRs,
   several standards, and **zero FRDs, zero architecture blueprints, and
   zero feature blueprints**. Building the CLI is the first honest
   chance to run the full chain — FRD → feature blueprint →
   `groom-feature-blueprint` → beads → `bead-bonanza` → `retro-spec`.
   Whatever the CLI is worth, the evidence from running the framework on
   a real feature is worth as much. I'd treat that as an explicit goal
   of the project, not a side effect.

---

## 11. Open questions for tomorrow

1. **Corpus model.** How does the CLI know a template isn't a document?
   Config globs, a `type: template` convention, or a marker file? And
   how do guides mark illustrative links — mandatory backticks (a
   standards change), or an ignore-region comment?
2. **EARS lint severity.** Warning-only, section-scoped, or push for
   structural markup of normative statements (§5)?
3. **Scope of v1.** I'd cut it at: index, frontmatter schema, slug
   uniqueness, link + anchor resolution, record blocks, line width,
   instruction blocks, `next-slug`. That covers both readiness
   checklists' mechanical checks and nothing else. AC traceability (§4
   #9) is the highest-value *next* thing but it's a second corpus (the
   test suite) and deserves its own increment.
4. **Where does config live?** `.sddrc.toml` at the spec root, a
   `spec:` block in an existing file, or convention-only with flags?
   Affects how portable the tool is to a user's repo.
5. **Name.** `sdd`? `spec-check`? It ends up in every skill body and
   every CI log.
6. **Do we generalize deliberately?** The universal layer is genuinely
   reusable outside this framework (Obsidian vaults, docs-as-code
   repos). Designing for that costs a little now and is nearly
   impossible to retrofit. Worth it, or premature?
7. **Hook aggressiveness.** Skill-scoped write-time validation (§6.3)
   from day one, or wait until the rules are proven non-noisy?

---

## 12. Suggested sequence

1. Wire CI for the two existing scripts — `agent-plugins-bgy`. No
   dependency on any of this.
2. Fix the ADR slugs — `agent-plugins-5nz`. Also gives the link
   validator a green baseline to be built against.
3. Settle the open questions above.
4. Author the FRD, then the feature blueprint. Dogfood the chain.
5. Build v1 to the scope in Q3.
6. Wire the skills — checklist steps call the CLI, with the fix/re-run
   loop and a bound, per `skill-authoring-standards`.
7. Retro the whole run with `retro-spec` and `retro-skill`. This is the
   first project big enough to produce real evidence about the
   framework.

---

## Sources

- [Claude Code — Plugins reference](https://code.claude.com/docs/en/plugins-reference)
- [Claude Code — Extend Claude with skills](https://code.claude.com/docs/en/skills)
- [Claude Code — Hooks](https://code.claude.com/docs/en/hooks)
- [PEP 723 — Inline script metadata](https://peps.python.org/pep-0723/)
- [uv — PEP 723 inline script metadata](https://deepwiki.com/astral-sh/uv/12.1-pep-723-inline-script-metadata)
- [Zabaca/mermaid-validate](https://github.com/Zabaca/mermaid-validate)
- [tetrafolium/mermaid-check (Go)](https://pkg.go.dev/github.com/tetrafolium/mermaid-check)
- [rtuin/mcp-mermaid-validator](https://github.com/rtuin/mcp-mermaid-validator)

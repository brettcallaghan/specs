# Brief: deterministic validation tooling for the spec document set

Research and options. Input to an FRD, not a spec.

**Revision 2** — restructured around review comments. The major changes:
a product framing section (§1), the general/domain feature split you
asked for (§2), Gherkin research with empirical results (§4), a complete
agent-judgment catalogue (§5), and a reversed language recommendation
(§8) now that legacy weight is off the table.

---

## 1. What we're building and why

Three motivations, in the order they'd appear in an FRD overview.

**1. CI for other people's spec repos.** The primary consumer is not
this repo — it's a team that adopts the framework, points the tool at
their spec root, and gets a build failure when their documents drift.
That makes *portability* a first-class requirement, not a nice-to-have:
the tool cannot assume our directory layout, our document types, or our
prefix registry.

**2. Cheaper, more reliable agent validation.** Every mechanical check
an agent performs by reading is tokens spent and a chance to claim a
check it didn't run. Concretely, on this repo: readiness check 8
("every `[[slug]]` resolves") costs roughly 42 file reads today. The
skill runs its checklist, fixes, and **re-runs** — so that's paid twice.
The `spec` skill's reference bundle is ~39k tokens; the per-run checking
work is larger than the bundle.

**3. The index is an enterprise architecture artifact.** This is the
one I most underweighted in revision 1. The JSON isn't a build
intermediate — it's the deliverable that gets pushed to a centralized
service so an organization can see its whole system architecture in one
place. That's the C4 payoff: rigorous component modeling feels heavy for
a single app (though it still helps humans and agents), and pays off at
enterprise scale where no one can hold the system in their head.

Motivation 3 has design consequences that motivations 1 and 2 don't:

- **The index schema is a published API.** It needs a version field, a
  compatibility policy, and a much higher bar for breaking changes than
  an internal cache would.
- **Indexes must merge across repositories.** Two repos indexed
  independently have to compose into one graph. That means a
  workspace/system identifier, and slugs that are unique within a
  declared namespace rather than globally.
- **It should interoperate.** Structurizr already publishes a JSON
  schema for C4 workspaces and is explicitly designed as an open format
  to avoid lock-in. Emitting something convertible to it — even as a
  separate `export` subcommand — buys rendering and existing enterprise
  tooling for free. Worth an ADR; probably not v1.
- **Partial results must be usable.** A repo with three broken documents
  should still contribute its other 997 to the graph. Reinforces "index
  never fails; validation may."

Worth stating plainly in the FRD: **this is three products sharing one
engine** — a CI linter, an agent accelerator, and an architecture graph
exporter. If the FRD only describes the linter, the design will be wrong
for the other two.

---

## 2. The split: general engine vs. domain profile

You want a clean seam. Here's where I'd put it.

**The seam:** the general layer knows about *markdown documents with
frontmatter that reference each other*. It has no opinion about what a
document means. The domain layer supplies vocabulary — which types
exist, what each requires, what the prefixes are — and the handful of
checks that can't be expressed as configuration.

The test for "is this general?": **could an Obsidian vault or a
docs-as-code repo with nothing to do with C4 use it?** If yes, general.

### 2.1 General engine

| ID | Feature | Notes |
|----|---------|-------|
| G1 | **Corpus discovery & frontmatter parse** | Walk a root or a single file. Parse YAML frontmatter. Never fail on a bad document — record it and continue. |
| G2 | **Frontmatter schema validation** | JSON Schema per `type`, supplied by config. Validate against the **raw string**, not the parsed value — YAML silently coerces `updated_at: 2026-07-29` to a date, so a parsed-value RFC 3339 check passes junk. This produced 42 false positives in my prototype. |
| G3 | **Slug registry** | Uniqueness (case-insensitive — see §11), format per configured pattern, and the `number` field (§3). |
| G4 | **Wikilink index & resolution** | `[[slug]]` and `[[slug#Anchor]]`. Never scan fenced or inline code spans — that alone removed 49 of 151 false hits here. |
| G5 | **Heading structure & section spans** | Every heading with its line range, and the span of content beneath it. Feeds template conformance (G5a) *and* the targeted-agent-review pattern (§6). Probably the highest-leverage general feature after the index itself. |
| G6 | **Tagged structured-data block extraction** | Generalizes `component` record blocks: extract any fenced block with a given info-string tag, parse it as yaml/json/toml, validate against a schema, attach it to the containing document and heading. The domain layer just says "tag `component`, schema X". |
| G7 | **HTML comment block extraction** | Both as content (structured data in comments) and as hygiene (surviving template instructions). |
| G8 | **Reference graph** | Derived from body wikilinks and configured frontmatter properties. Emits forward *and reverse* edges. See §3 on killing the `related:` ceremony. |
| G9 | **Git-derived metadata** | Read `updated_at` and change-detection from git rather than trusting frontmatter — and **write it back** (§7). |
| G10 | **Text hygiene** | Line width in characters, placeholder leakage (`<Name>`, `TODO`, `TBD`), trailing whitespace. |
| G11 | **Index & report emission** | Versioned JSON index; separate JSON report with severities. |
| G12 | **Fenced-block validators (pluggable)** | Mermaid is the first; the general shape is "hand a fenced block to an external validator by info-string tag". |

### 2.2 Domain profile

| ID | Feature | Built on |
|----|---------|----------|
| D1 | **Type registry** — adr, frd, standards, system/container/component/feature blueprint, contract, work-order | G1, G2 |
| D2 | **Prefix & slug conventions per type** (`adr-<nnn>`, `frd-<feature>`, …) | G3 |
| D3 | **Element record blocks** + the **record-block XOR blueprint-link** invariant | G6 |
| D4 | **AC identifiers & traceability** to the test suite | G5, §4 |
| D5 | **Status lifecycle** — vocabulary, `superseded_by` integrity, retired-document references | G2, G8 |
| D6 | **EARS normative-span extraction** for agent review | G5, §6 |
| D7 | **C4 rules** — naming, arrow direction, element/relationship consistency | G6, G12 |
| D8 | **Blueprint AC-coverage legend ↔ delta keys** | G5 |
| D9 | **Work-order Definition of Ready** | G2, G5 |
| D10 | **Template section conformance per type** | G5 |

### 2.3 How much of the domain layer is just configuration?

Mostly, and that's the design goal. D1, D2, D5, D10 are pure config —
type names, JSON Schemas, regex patterns, required headings. D3, D6, D8
need real code. D4 and D7 need code plus external parsers.

So the shipped shape is: **a general binary + a `spec-driven-dev`
profile that is 80% a config bundle.** An enterprise with its own
document taxonomy writes its own profile and gets everything in §2.1
for free. That's also the honest path to open-sourcing the general half
later — the seam exists from day one rather than being retrofitted.

**Open question (§14 Q1):** does the domain layer ship as a plugin
interface (WASM? subprocess? dynamic registration?) or is it compiled
in, with config as the only extension point for v1? I'd start compiled
in, config-extensible — a plugin ABI is a lot of surface to commit to
before we know what a second profile needs.

---

## 3. The index

Confirming what you already agreed with, plus what your comments change.

**A `number` field.** Agreed, and it should be optional and general:
any type using an incrementing `<nnn>` convention declares
`number: 11`, and `next-slug adr` becomes `MAX(number)+1` over the type
rather than parsing it back out of slugs. It also makes gaps and
collisions visible, which slug-parsing hides.

**Derive the reference graph; drop the ceremony.** Strongly agree, and
it's the single biggest ergonomic win available. Today `related:`,
`sources:`, `composes:` are hand-maintained and drift silently — my
prototype found `related: work-order-template` pointing at a template,
which has no slug. If body `[[wikilinks]]` are the source of truth, the
index derives both directions and nobody maintains anything.

The nuance worth an FRD decision: **not all relationships are
references, and not all references are semantic.** `parent_slug`,
`supersedes`, and `composes` carry meaning a bare link doesn't — "this
document replaces that one" isn't "this document mentions that one." I'd
propose:

- **Derived, never authored:** `references` / `referenced_by` — computed
  from body links.
- **Authored, still validated:** typed edges where the *type* matters —
  `parent_slug`, `superseded_by`, `composes`.
- **Retired:** `related`, which is the ceremony with no semantics.

**Startup cost is real and I measured it.** Your point about firing up
the CLI for a one-shot check:

| Operation | Median |
|-----------|--------|
| `python3 -c pass` | 12.4 ms |
| `+ import yaml` | 34.7 ms |
| `+ yaml, json, re, pathlib` | 38.4 ms |
| Full parse, this repo's 50 docs | 56.6 ms |
| A Rust/Go binary, same work | ~1–3 ms startup |

So Python pays ~38 ms before doing anything. For `slug-exists` that's
essentially all overhead, and it scales badly: at ~1,000 documents the
parse alone is roughly a second. If a `PostToolUse` hook fires on every
write, that's the difference between imperceptible and irritating. This
is now a primary input to §8.

Two mitigations regardless of language: cache the index with an mtime
or content-hash check so repeat invocations skip reparsing, and offer a
long-lived `serve` mode for the interactive case.

---

## 4. Gherkin — I ran the experiments

You asked whether Gherkin can annotate every AC with an ID and whether
a tool extracts them deterministically. **Yes to both**, and it's a
better fit than I expected. Everything below is from actually running
the official `gherkin-official` Python parser, not from documentation.

**Tags carry our exact ID format.** `@AC-CHK-003.1` parses as a valid
tag — dots, hyphens and uppercase are all fine. The only hard rule is
no whitespace. Our existing `AC-<PREFIX>-<nnn>.<n>` convention needs no
change.

**Tags inherit and the parser resolves it.** A tag on `Feature:` or
`Rule:` propagates to every `Scenario:` beneath it, already merged in
the AST. So an FRD-level tag lands on every AC automatically.

**Every scenario carries a line number**, so diagnostics can point at
`file:line`.

**It parses from a string, not just a file** — which means fenced
` ```gherkin ` blocks embedded directly in an FRD work. I extracted two
blocks from a mock FRD, wrapped them in a synthetic `Feature:` header,
parsed, and got both ACs back with tags and steps intact.

**The parser genuinely rejects malformed ACs.** A typo'd step keyword
(`Wen` for `When`) produces:

```
Parser errors: (5:5): expected: #EOF, #TableRow, #DocStringSeparator,
#StepLine, #TagLine, #Examples
```

That is exactly the benefit you were after — a language parser, not a
regex, deciding whether an acceptance criterion is well formed.

**Two constraints found by testing:**

1. A fragment does **not** parse on its own. `Scenario:` alone is
   rejected; so is `Rule:` alone. Embedded blocks must be wrapped in a
   synthetic `Feature:` before parsing.
2. Because of the wrapping, scenario line numbers are relative to the
   synthetic document, not the markdown file. An offset map is required
   to report errors at the right line in the FRD. Small, but it's the
   kind of thing that gets discovered late and hurts.

**Official implementations:** .NET, Java, JavaScript, Ruby, Go, Python,
C, Perl, PHP, Dart, C++. **Rust is not on that list** — `cucumber-rs/gherkin`
is a third-party pure-Rust reimplementation. That interacts with §8 and
is called out there.

### Three ways to use it

| | Shape | Gets us |
|-|-------|---------|
| **A** | FRD keeps prose ACs; separate `.feature` files carry `@AC-*` tags | Traceability without touching the FRD template. Two artifacts to keep in sync — the drift problem we're trying to kill. |
| **B** | ACs **are** ` ```gherkin ` blocks inside the FRD | One artifact. Parser-validated structure. Deterministic ID extraction. Requires changing `frd-template.md` and `ac-traceability-standards.md`. |
| **C** | Prose EARS AC stays; the `*Example:*` beneath it becomes a gherkin block | Smallest change; validates the example, not the criterion. |

**I'd argue for B**, with one framing that makes it much less radical
than it looks: the FRD template *already* requires an EARS statement
plus a concrete example per AC. Option B just gives that pair a grammar
— the `Scenario:` name is the EARS sentence, the `Given/When/Then` steps
are the example that was already mandatory.

The reframe that matters: **adopting Gherkin is not adopting Cucumber.**
We'd use it as a validated grammar for acceptance criteria and a
deterministic ID carrier. Whether anyone *executes* the scenarios is a
separate, later, per-project choice. Teams not using Cucumber keep
citing IDs from their own tests exactly as `ac-traceability-standards`
already specifies.

The genuine tension — and I think this is the ADR: **EARS and Gherkin
overlap.** "WHEN payment succeeds, the system SHALL display X" and
"When payment succeeds / Then X is displayed" say the same thing twice.
Option B resolves it by assigning each a role (EARS = the normative
sentence in the scenario name; Gherkin steps = the concrete instance),
but a reasonable person would say pick one. That deserves its own ADR
after the FRD, and probably an experiment where we write one real FRD
both ways.

---

## 5. Catalogue of checks that stay with the agent

You asked for all of them. Grouped by *why* they can't be mechanized —
which matters, because "needs a model" and "needs a human" are different
problems.

### 5.1 Requires semantic judgment about content

| Check | Source | Why not mechanical |
|-------|--------|-------------------|
| Is a requirement genuinely testable? | `ears-guide` common mistakes | "Appropriate error message" is grammatically identical to a good AC |
| Is a vague qualifier vague? | `ears-guide` | Needs domain knowledge — "promptly" is vague, "within 200 ms" isn't, and only context distinguishes "reasonable default" |
| Is the EARS pattern the *right* one? | `ears-guide` pattern selection | Detecting a missing trigger is mechanical; deciding a behavior is truly ubiquitous is not |
| Is a compound requirement genuinely compound? | `ears-guide` | My prototype fired 18 times, ~6 real. Prose *about* requirements ("It is SHOULD, not SHALL, because…") is indistinguishable by regex |
| Does a delta carry enough design intent? | blueprint checklist 7 | Explicitly semantic; the checklist says so |
| Is the User Story a real user story? | `frd-template` anti-example | "As a user, I want payment to work" is structurally perfect and useless |
| Is the Problem statement self-contained? | work-order DoR 1 (fresh-agent test) | Requires simulating an agent with no context |
| Is this the right document type? | `spec` routing | The skill's core job |
| Is the C4 level right? | `c4-guide` | Component vs. module vs. container is a modeling judgment |
| Is an alternative in an ADR fairly represented? | `adr-template` | Straw-man detection is semantic |
| Is the blast radius correct? | work-order DoR 8 | Presence is checkable; correctness needs code knowledge |
| Is a test's AC citation *honest*? | `ac-traceability-standards` | We can prove a citation exists and resolves. We cannot prove the test actually exercises that criterion |

That last one is worth flagging loudly in the FRD's non-goals. AC
coverage tooling proves *citation*, never *verification*. A team that
reads a green coverage report as "all criteria are tested" has been
misled by us.

### 5.2 Requires judgment about scope or sizing

Right-sizing to a single session (DoR 2), kind-appropriate completeness
(DoR 5), tier assignment (DoR 7), tracker-reconciliation classification
(subsumed / prerequisite / out-of-scope), and whether an ADR is
warranted at all.

### 5.3 Requires a human, not an agent

Adopt/defer/reject on an improvement RFC; approving a status transition
to `approved`; accepting a deviation from a SHOULD; and deciding whether
a `blocked-ungroomable` bead gets reshaped or dropped.

### 5.4 Mechanical in principle, deliberately not in v1

Worth naming so they're recognizable as deferred rather than missed:
prose style and readability, terminology consistency across documents
(one term per concept — doable with a glossary, and a real project),
and whether a diagram matches the element records beside it.

---

## 6. Extract deterministically, then review narrowly

Your suggestion at line 216 is the best idea in the review, and it
generalizes well beyond EARS. The pattern:

> The CLI catalogs which sections are expected to contain normative
> statements, extracts those spans (G5), and hands **filenames + spans**
> to an agent. The agent reviews only the normative text.

Why this is better than either alternative:

- **Versus a regex lint:** it puts the judgment where judgment belongs.
  All twelve items in §5.1 become tractable — you're asking a model a
  semantic question, which it's good at, instead of asking a regex, which
  it isn't.
- **Versus "agent reads the document":** the agent reads the normative
  spans, not the frontmatter, not the context section, not the
  alternatives table. On `c4-guide.md` (3,600 words) the normative spans
  are a small fraction of the text.

It also makes the review *auditable*: the extraction is deterministic
and repeatable, so two runs review exactly the same spans. Today, two
agents reviewing "the requirements" may not be looking at the same text.

Design implication: `sdd spans --normative --json` is a first-class
subcommand, not a debug flag. Output is `{file, heading, start_line,
end_line, text}` — enough for an agent to review and cite precisely.

This is probably how §5.1 items 1–6 actually ship: not as CLI checks,
but as CLI-extracted work packages for a review subagent.

---

## 7. Git as the source of truth for metadata

Both your comments here point the same direction, and they change the
tool's role from reader to writer.

**`updated_at`:** derive it from `git log -1 --format=%cI -- <file>`.
Don't verify a hand-maintained field — own it. Removes an entire class
of drift and one line of ceremony per edit.

**Version bumps:** likewise derivable. If a document's normative content
changed between two refs and its `version` didn't, that's a fact git
knows.

This introduces a **write mode**, which is a genuine scope decision:

- `sdd fmt` — rewrite derivable frontmatter in place. Idempotent, safe
  to run in a pre-commit hook.
- `sdd fmt --check` — CI mode, fails if a rewrite would change anything.

Two caveats worth surfacing now:

1. **Working-tree edits have no commit yet.** An uncommitted document's
   git timestamp is stale by definition. Either `fmt` runs pre-commit
   (so it stamps what's about to be committed), or it uses the working
   tree's mtime when the file is dirty. Pre-commit is cleaner; it needs
   the FRD to say so.
2. **"Normative content changed" needs a definition.** A typo fix in a
   non-normative sentence shouldn't force a version bump. This is where
   G5 section spans pay off again: diff the *normative* spans, not the
   whole file.

Version bumping specifically may be better as a **warning that proposes
the bump** rather than an automatic rewrite — MAJOR/MINOR/PATCH is a
semantic call about whether a change breaks a consumer, and git can't
make it. My suggestion: `fmt` owns `updated_at` outright, and *proposes*
a version bump with a reason.

---

## 8. Language and distribution — I've changed my recommendation

Revision 1 recommended Python largely because it matched what's already
here. You've removed that consideration, and two things you raised push
hard the other way: the enterprise index artifact, and one-shot
invocation cost. So:

**Recommendation: Rust.** Not because you like it — because the
measurements in §3 point there and the distribution objection no longer
holds.

| | Rust | Go | Python |
|-|------|-----|--------|
| Startup | ~1–3 ms | ~1–3 ms | 38 ms fixed |
| 1,000-doc index | tens of ms | tens of ms | ~1 s |
| Adopter install | binary, no runtime | binary, no runtime | needs Python + deps |
| Official Gherkin parser | ✗ (third-party `cucumber-rs/gherkin`) | ✓ | ✓ |
| Agent can patch mid-session | ✗ | ✗ | ✓ |
| Prototyping speed | slowest | middle | fastest |

### Is binary distribution a problem for adopters?

Short answer: **no, not in 2026** — but only if we avoid one specific
mistake.

The mature path is `cargo-dist`, which builds per-platform binaries and
generates installers for npm, Homebrew, and shell/PowerShell one-liners.
Python wheels can also carry a binary via the `.data/scripts/`
directory, which makes `pip install` work as a cross-platform binary
installer — this is how `uv` itself ships.

**The mistake to avoid:** cargo-dist's npm installer traditionally uses
a `postinstall` script that downloads the binary from GitHub Releases at
install time. That fails in exactly the environments our enterprise
adopters live in — locked-down networks that mirror npm and PyPI but
block GitHub Releases — and it's a supply-chain surface. The fix is
well-trodden: ship **platform-specific packages with the binary inside**
(`optionalDependencies` on npm, platform wheels on PyPI), no postinstall
download. `cargo-npm` exists specifically to do this.

So the distribution matrix I'd propose: GitHub Releases (canonical),
npm platform packages, PyPI platform wheels, Homebrew, and
`cargo install` from source. An adopter uses whichever package manager
their CI already trusts.

### The two real costs of Rust

1. **No official Gherkin implementation.** `cucumber-rs/gherkin` is a
   pure-Rust reimplementation and reportedly covers the language well,
   but it isn't part of the upstream shared test suite that keeps the
   official implementations behaving identically. If Gherkin becomes
   load-bearing (§4 option B), this matters. Mitigations: use the Rust
   crate and accept the risk; shell out to the official CLI when
   present; or treat Gherkin the way we treat mermaid (§9) — an
   external validator, `unverified` when absent. I'd lean on the Rust
   crate for structure and keep the shell-out as a `--strict` escalation.
2. **Agents can't patch it mid-session.** Genuinely lost. Partly bought
   back by making the domain layer configuration (§2.3) — an agent can
   edit a JSON Schema or a type registry entry without a compiler.

**Honest counter-position:** if we want an FRD-to-working-tool loop in
days rather than weeks, Python or Go get there faster, and Go keeps the
official Gherkin parser and the fast startup. Rust's edge over Go here
is mostly ecosystem taste, not measurement. This is precisely an ADR
decision, so I'd rather the FRD state the *requirements* — sub-10 ms
one-shot latency, no runtime dependency for adopters, mirrorable
package distribution — and let the ADR pick the language against them.

---

## 9. Delivery into agent sessions

Unchanged from revision 1 and you agreed with all three, so briefly:
`${CLAUDE_SKILL_DIR}` + `allowed-tools` for prompt-free invocation of a
bundled binary; plugin `hooks/hooks.json` with a `PostToolUse` matcher
on `Write|Edit`; and skill-frontmatter `hooks:` scoped to the skill's
own lifetime.

**On "CI during session runs?" — yes, and that's the right way to say
it.** A `PostToolUse` hook is CI with a much shorter feedback loop: it
fires on the write rather than on the push, and it can return
`additionalContext` or `decision: "block"` so the agent must deal with
the failure before continuing. The hook can't prevent the write (it
fires after), but it can prevent the agent from moving on — which is the
part that matters.

This is also where §3's startup measurement bites hardest. A hook on
every `Write` at 38 ms of pure interpreter overhead is tolerable; at
~1 s on a large repo it isn't. Scoped hooks plus an index cache handle
it, but it's another vote for a fast binary.

**Mermaid** stays optional, as you agreed — cheap structural checks
always, shell out to a real validator when present, report `unverified`
otherwise (which `skill-authoring-standards` already requires). Agreed
that the recommended tool belongs in an ADR; the candidates to weigh are
`mermaid-validate` (official parser + jsdom, needs Node),
`mermaid-check` (pure Go, no browser, but a reimplemented grammar), and
`mmdc` with a headless browser (most accurate, far too heavy for a gate).

---

## 10. Does this let us lift standards docs out of the plugin?

Rereading your comment — the question is whether the *skill's* bundled
references shrink because validation moves to the CLI. Partly yes, and
the framing that makes it precise is:

> **Authoring knowledge must stay. Verification knowledge can go.**

An agent still needs to know the wikilink syntax and the required
frontmatter fields *in order to produce* a valid document. It does not
need the verification procedure, the checklists, or the failure
taxonomies — those become `sdd check`.

Measuring the `spec` skill's bundle (~39k tokens across 20 files):

| Reference | Tokens | Disposition |
|-----------|--------|-------------|
| `template-authoring-standards.md` | ~2,645 | Liftable — it's for template authors, not document authors |
| `blueprint-readiness-checklist.md` | ~1,041 | Mostly liftable (6/10 mechanical) |
| `ac-traceability-standards.md` | ~926 | Mostly liftable — it's a tooling spec |
| `cross-document-references.md` | ~982 | Partly — syntax stays, verification goes |
| `frontmatter-standards.md` | ~1,474 | Partly — field list stays, rules go |
| `document-readiness-checklist.md` | ~746 | Mostly liftable (6/8 mechanical) |

Realistic ceiling: **~6–7k tokens, roughly 17% of the bundle.** Real,
but not transformative on its own.

**The bigger saving is the work, not the bundle.** The references load
once. The *checking* happens every readiness run, twice per run
(check → fix → re-check), and readiness check 8 alone is ~42 file reads
on a repo this size. That cost scales with corpus size; the bundle
doesn't.

**The mechanism that makes lifting actually work: diagnostics carry the
rule.** If the error message says

```
adr-011-targeted-verification-gates: slug does not match the registered
form for type 'adr' (adr-<nnn>) — see cross-document-references.md
```

…the agent learns the rule at the moment it's violated. Context moves
from preload to on-demand. That's worth more than the 17%, and it's a
requirement on the CLI's error messages that belongs in the FRD:
**every diagnostic names the rule and cites the document that defines
it.**

---

## 11. Cross-platform — resolved

Per your comments: **80 characters means 80 characters** (Unicode
codepoints, not bytes and not display columns), and **slug uniqueness is
case-insensitive**. Both should be stated in the FRD so they're not
rediscovered later.

Still live: normalize CRLF on read (Windows checkouts break span
arithmetic otherwise); emit POSIX-style paths in the JSON index so
reports diff across machines; and prefer invoking the binary directly in
hook commands rather than a `.sh` wrapper, since our current gates are
all shell scripts and those don't run on Windows without Git Bash.

CI should prove this rather than assert it: ubuntu × macos × windows.

---

## 12. Testing, fixtures, and CI

**Where do fixtures live?** Your question at line 389. I'd propose the
CLI becomes **its own repository** — and this follows from §1 more than
from testing convenience:

- It's a general-purpose tool with an enterprise consumer; its release
  cadence and versioning shouldn't be coupled to a Claude Code plugin.
- It ships as compiled binaries via several package managers (§8), which
  wants its own release pipeline.
- The plugin then **depends on** a released version rather than vendoring
  source — the same way it would depend on any other tool.

Under that split: fixtures, unit tests, and the source live in the CLI
repo. This repo keeps a *thin* integration check — "the pinned binary
version still passes on our own docs" — which doubles as the dogfood
signal and would currently fail on the ADR slugs.

Counter-argument worth weighing: a single repo is simpler while the tool
and the framework are co-evolving daily, and they will be for a while.
A reasonable compromise is to start in-repo under a top-level directory
with its own release workflow, and split once the interface stabilizes.
Listed as Q6.

**Test strategy** (unchanged and still what I'd do): a fixture corpus of
one deliberately-broken document per rule plus a clean control, snapshot
tested against `--json`; a meta-test asserting every registered rule ID
has at least one fixture and one control; property tests on the
code-span stripper (highest-risk component — it decides what counts as a
real link); self-application against this repo in CI; and a frozen
golden-path sample project.

---

## 13. Decisions settled in review

Recording these so the FRD doesn't relitigate them.

| # | Decision |
|---|----------|
| 1 | The index is a first-class product artifact, not a cache — schema is an API |
| 2 | Clean general/domain split, with the general layer usable outside this framework |
| 3 | Corpus membership is determined by frontmatter presence + `type`; unknown/missing must be **reported** for agent judgment, not silently skipped |
| 4 | Optional `number` frontmatter field for `<nnn>`-numbered types |
| 5 | `references` relationships are derived from body links, not hand-maintained |
| 6 | `updated_at` is derived from git and written, not verified |
| 7 | Record-block extraction generalizes to any tagged structured-data block |
| 8 | HTML comment blocks are extracted too |
| 9 | Markdown headings and inter-heading spans are indexed |
| 10 | Mermaid validation is optional; the recommended tool gets an ADR |
| 11 | 80 characters means codepoints |
| 12 | Slug uniqueness is case-insensitive |
| 13 | No legacy-tooling weight on the language choice |

---

## 14. Open questions

**Q1 — Domain layer extensibility.** Config-only for v1 (my
recommendation), or a real plugin interface from the start? Affects
whether an enterprise can add a document type without a PR to us.

**Q2 — Gherkin adoption shape.** §4 options A, B, or C? B is the
strongest but rewrites `frd-template.md` and `ac-traceability-standards.md`.
Should we write one real FRD both ways before committing?

**Q3 — EARS vs. Gherkin overlap.** If B, does the EARS sentence live in
the `Scenario:` name (my suggestion), or do we drop one of the two
notations? This may be the most consequential spec decision on the table
— it changes the FRD template, the EARS guide, and the AC standards
together.

**Q4 — Language.** Do you want the FRD to state latency and
distribution *requirements* and defer the choice to an ADR (my
preference), or to name Rust up front?

**Q5 — Write mode scope.** Does v1 include `sdd fmt` (§7), or is v1
read-only with formatting as a follow-on? A tool that edits your
documents is a different trust proposition than one that reads them.

**Q6 — Repository split.** Separate CLI repo now (§12), or in-repo with
its own release workflow until the interface settles?

**Q7 — Naming.** `sddrc` was a bad placeholder — `rc` is the Unix
"run commands" convention (`.bashrc`, `.vimrc`), which is opaque and
says nothing. For the config file I'd suggest `spec.toml` or
`.spec-index.toml` at the spec root. For the binary: `spec` collides
with too much, `sdd` is terse and meaningless to a newcomer. Something
like `specd`, `speclint`, or a real product name — it will appear in
every skill body, every CI log, and every enterprise runbook. Worth
picking deliberately rather than inheriting my placeholder.

**Q8 — Index namespacing.** For cross-repo merge (§1), what identifies a
workspace? A field in the config file, the git remote, or an explicit
`--workspace` flag? Needed before the schema is published, hard to add
after.

**Q9 — Structurizr interop.** Worth an `export --format structurizr`
subcommand, or does that presume a rendering choice adopters should make
themselves?

---

## 15. Sequence

1. **FRD** for the CLI, stating the three products (§1), the
   general/domain split (§2), and requirements rather than technology.
2. **ADRs** for the decisions the FRD deliberately defers: language and
   runtime; Gherkin adoption and its interaction with EARS; index schema
   and versioning; mermaid validator; repository split.
3. **Container blueprint** for the CLI.
4. Groom → beads → build.

Two things worth doing in parallel that don't block any of it: wire CI
for the two existing scripts (`agent-plugins-bgy`), and fix the ADR
slugs (`agent-plugins-5nz`) so the link validator has a green baseline
to be built against.

---

## Sources

- [Claude Code — Plugins reference](https://code.claude.com/docs/en/plugins-reference)
- [Claude Code — Extend Claude with skills](https://code.claude.com/docs/en/skills)
- [Claude Code — Hooks](https://code.claude.com/docs/en/hooks)
- [cucumber/gherkin — parser and compiler](https://github.com/cucumber/gherkin)
- [Gherkin reference — Cucumber docs](https://cucumber.io/docs/gherkin/reference/)
- [cucumber-rs/gherkin — pure Rust implementation](https://github.com/cucumber-rs/gherkin)
- [structurizr/json — C4 workspace JSON schema](https://github.com/structurizr/json)
- [Structurizr — why "as code"](https://docs.structurizr.com/as-code)
- [cargo-dist](https://axodotdev.github.io/cargo-dist/)
- [cargo-npm — Rust CLIs on npm without postinstall](https://github.com/abemedia/cargo-npm)
- [Packaging a Rust tool — Rust CLI book](https://rust-cli.github.io/book/tutorial/packaging.html)
- [PEP 723 — Inline script metadata](https://peps.python.org/pep-0723/)
- [Zabaca/mermaid-validate](https://github.com/Zabaca/mermaid-validate)
- [tetrafolium/mermaid-check (Go)](https://pkg.go.dev/github.com/tetrafolium/mermaid-check)

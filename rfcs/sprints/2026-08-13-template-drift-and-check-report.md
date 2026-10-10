# Sprint report — template drift gate and the check-report work

Run `1786603162-7` · branch `claude/validate-corpus-spec-cli-0t0cxo` · PR #22
7 beads queued, 7 closed, 0 blocked, 0 skipped.

| Bead | Tier | Commit |
|------|------|--------|
| ziv | light | `c0e9997` |
| 3xx | standard | `e44f4a3` |
| j5n | standard | `5fe29ee` |
| k9j | light | `c1c4c88` |
| qz9.2 | standard | `08e69e4` |
| lwu | standard | `dee1f97` |
| qz9.3 | standard | `2256efb` |

Headline: `scripts/check-template-drift.py` went from 8 findings to **0**, which
unblocks 9fx (flip the CI step from `continue-on-error` to gating).

## 1. Grooming

**Under-groomed beads: one, and it was an acceptance criterion that was
confidently wrong.** ziv's AC said the checked-document count
"increases by exactly 1". It increased by 2. The substring rule
`"template" in name` was hiding TWO documents, not one:
`template-authoring-standards.md` and
`rfcs/retired/rfc-0007-spec-architecture-templates.md`, whose name contains
"templates" inside `spec-architecture-templates`. Measured on one tree: 42
checked before, 44 after, 0 findings both times.

The AC asserted a count without enumerating what the pattern actually matched.
A conforming implementation would have been judged as failing. **Lesson: an AC
that states a number must be derived from a query, not from the one example
that motivated the bead.**

**A second AC gap, self-reported.** qz9.3 records that the
document count and exclusions attach only to a whole-corpus run, not to
`spec check <path>`. That is defensible — a single-document check does not
answer a corpus-scale question — but the AC did not say so, so the implementer
had to settle it. An AC that says "check output carries the count" without
scoping it to whole-corpus runs leaves that open.

**Tier predictions: no escalations, and two beads were arguably over-tiered.**
See §3a/3b below.

**Dependency edges: one was missing.** lwu's ADR must cite the
append-only rule k9j adds. Grooming wired the pair as `related`,
not `blocks`, reasoning the ADR could be written first. The planner caught it
and sequenced lwu behind k9j anyway. The edge type was wrong: a document that
must cite text another bead introduces is blocked by it, not merely related.

## 2. Specs

**`[[wikilink]]` does not resolve for a document outside the corpus root, and
nothing warned in advance.** lwu cited
`[[ac-traceability-standards]]` and `spec check --strict` failed with
`reference/unresolved` — the file carries `type: standards` frontmatter but
lives under `plugin/skills/spec/references/`, outside the `specs/` root, so it
is never indexed. ADR-012's out-of-set-doc convention (backtick path) is the
right answer and the implementer found it, but only after the gate failed.
The trap is that the document *looks* citable: it has the frontmatter of a
corpus member.

**Version-bump policy is unstated on standards documents.**
k9j bumped `ac-traceability-standards.md` 0.2.0 -> 0.3.0 by
following `frontmatter-standards.md`'s semver convention, noting the document
itself states no bump policy. That inference was correct but was an inference.

## 3. Plan

`## Amendments` recorded two entries, both resequencings:

1. **lwu moved from wave 2 into wave 1** once k9j merged, since k9j landing was
   its only reason for being held. Safety re-checked before dispatch rather than
   assumed (no shared `c4_changes`, no file overlap, `worktree.baseRef` still
   `head`, ancestor check pinned to a HEAD containing k9j).
2. **Per-bead drift-gate targets replaced "exit 0"** for the two beads that
   jointly owned that gate's 8 findings.

**What would have produced a better initial bucketing:** the planner bucketed
on `c4_changes`, which is empty for five of seven beads because this repo's C4
model documents only the Go CLI. The real collision signal for those five was
file-level and the orchestrator had to supply it by hand (j5n and k9j both run
`sync-shared-references.sh` against `references/shared/`). Bucketing that
depends on a signal the repo does not populate is bucketing done by the
orchestrator, not the planner.

## 4. Skill

**A gate's own instruction can be unsatisfiable when two beads share it.**
`check-template-drift.py` held 8 findings split across two beads dispatched
concurrently. "Gates must pass" is unachievable for either alone, and both
would have had to edit the other's file to try — violating scope and colliding
on merge. The fix was per-bead targets (8->1 and 8->7, run ends at 0) plus an
explicit prohibition. The skill's baseline mechanism covers "a pre-existing
failure is not yours"; it does not cover "the baseline is owned by your
sibling in this same wave".

**The progress ledger has a gap and the acceptance check was right not to use
it.** qz9.3 emitted `gates-passed`, `retro-written`, `pushed` —
but never `committed`. The Monitor therefore never fired for the run's last
bead. The bead was fine (tracker closed, commit present, merged). This is a
live demonstration of "the ledger is a hint, never the authority".

**`check-skills.sh`'s path-resolution rule blocks cross-skill citation.**
j5n tried to cite `groom-beads`'s `references/bead.schema.json`
as a backticked path from a `spec` skill document; rule 2 requires backticked
paths to resolve relative to the citing file's own directory, so cross-skill
references are not expressible that way. Worked around by naming it in prose.

**Harness worktrees land inside the repository.** The skill's own mechanics
place worktrees as a *sibling* of the repo ("A worktree inside the repository
would read as dirt in every `git status` the run takes"). The harness's
`isolation: "worktree"` puts them under `.claude/worktrees/` instead, which is
exactly the dirt the skill warns about — a stop hook flagged them as untracked
files to commit mid-run. Added to `.gitignore` in this run.

## Full harvested findings, verbatim

Every `retro:` comment from the run's closure set, including those warranting
no RFC. `triage-rfc` defers on recurrence, so a finding filtered out here can
never be observed appearing twice.

### ziv

- retro: Pre-fix verification: `scripts/validate-corpus.py | grep template-authoring-standards` produced no output, confirming the file was silently absent (neither checked nor excluded). After fix, template-authoring-standards.md is included in the checked count.

### 3xx

- retro: Observed pre-fix drift (check-template-drift.py against rfc-template.md, verbatim): template/missing-required for effort, model, runs_observed, severity, skill, skill_version, plus template/unknown-property for audience — 7 findings total, matching the bead's Actual section exactly.
- retro: Silent decisions made — `skill:` is a fixed literal 'retro-spec' (not a placeholder), since this template only ever produces the project variant; skill_version, model, effort, runs_observed, severity kept as bracketed placeholders mirroring skill-rfc-template.md's shape. Prose reword: renamed 'Audience semantics' to 'RFC variants' and 'Scope … per audience' to 'Scope … per variant'; replaced every `audience: project` / `audience: skill-authors` label with plain 'project —' / 'skill-authors —' bullets, tying each to its slug shape (numbered vs unnumbered) and file location. SKILL.md lines ~149/159 reworded accordingly.
- retro: No dead ends — the fix was mechanical once the schema's required set and the sibling template's frontmatter shape were confirmed. Also bumped the template's internal 'Version: 1.2' -> '1.3' and retro-spec/SKILL.md 1.4.0 -> 1.5.0.

### j5n

- retro: observed pre-fix, check-template-drift.py: `template/unknown-type: work-order is not in the DocumentType vocabulary (…)` (checked 12 templates, 8 findings).
- retro: observed pre-fix, YAML parse of the frontmatter: `yaml.scanner.ScannerError: mapping values are not allowed here … line 28, column 23: tier_reason: <one line: what about th…`. Confirmed only line 28 has an actual colon-space fault (line 8's `<YYYY-MM-DDTHH:MM:SSZ>` has colons but none followed by a space, so it parses fine).
- retro: friction — my first edit cited `groom-beads`'s `references/bead.schema.json` as a backticked path; check-skills.sh's path-resolution check (rule 2) flagged it because backticked paths must resolve relative to the citing file's own directory, and cross-skill references aren't permitted that way. Fixed by referring to it in prose instead. Silent decision: left check-template-drift.py's docstring example unchanged since it's generic documentation, not an assertion about this file's state.

### k9j

- retro: straightforward — the Design's exact placement (beside the existing no-reuse sentence, before the first `---` separator) and EARS phrasing pointer matched the doc's existing rule-block style with no rework. Bumped version 0.2.0 -> 0.3.0 (minor, new normative rule) and updated_at, following frontmatter-standards.md's semver convention since the doc itself states no bump policy.

### lwu

- retro: Citing `[[ac-traceability-standards]]` failed `spec check --strict` (reference/unresolved) even though the file has `type: standards` frontmatter — it lives under plugin/skills/spec/references/, outside the specs/ corpus root, so it isn't indexed. Fixed by citing it as a backtick path per ADR-012's out-of-set-doc convention.
- retro: Silent decisions — filename slug `ADR-020-frd-lifecycle-and-ac-keyed-deltas.md`; section ordering follows the template/ADR-013 precedent unchanged; Consequences includes a Negative naming the unresolved same-feature boundary case (when a change is large enough that 'same feature' stops being obvious), since the Design didn't specify Consequences content beyond the append-only dependency.

### qz9.2

- retro: Reproduced the pre-fix bug before implementing: a config with root=docs pointing at an empty directory produced, both with and without --strict, 'no findings' followed by every rule listed 'passed', exit=0. Confirms the bead description exactly — a broken/empty configuration read as a clean pass.
- retro: Implicit decisions — rule id is exactly `corpus/empty` (Design's suggestion, adopted verbatim). Message: `spec root %q holds no corpus members: the run checked zero documents`, with the configured root both quoted in the message and used as the finding's Path (no document exists to point at, so the spec root fills that field; Position is the zero Span).
- retro: Friction — FindingDetail is a closed union (schemas/finding.strata) whose variants are generated, unexported-method Go types cli/core/finding cannot extend, and the gate baseline forbids regen (stale strata binary) and hand-editing generated code. No existing variant fits 'corpus-level, no document, freeform message'. Considered FindingUnverified{Reason} but rejected it: that tag is coupled 1:1 with SeverityUnverified everywhere else, and this finding is SeverityError by mandate. Settled on reusing FindingFrontmatterParse{Message}, the one variant already precedented at SeverityError with a bare message field — but a JSON consumer keying off `detail.kind` will see "frontmatter-parse" for a rule named corpus/empty, which is misleading. A future bead adding a real CorpusEmpty variant would remove this mismatch. **(Filed: nu2.)**

### qz9.3

- retro: document count and exclusions are attached only for a whole-corpus run (check.go's `path==""` branch), not for `check <path>`. RuleCorpusEmpty itself is scoped to `only==nil`, and CheckFor narrows to one file rather than answering a corpus-scale question, so a single-document check's Text() renders exactly as before — an implicit decision, not stated in the AC.
- retro: count-line wording is "checked N document(s)", matching scripts/validate-corpus.py's "checked N documents" phrasing it is meant to be an at-least-as-good replacement for. N = `len(Corpus.Index().Entries)`, the same corpus-member count RuleCorpusEmpty guards on, so the count and the corpus/empty finding stay self-consistent by construction.
- retro: exclusions render as `  <path>  <reason>` with an appended `(declared type %q)` only when reason is unknown-type and a type was recorded — mirrors corpus.go's notAMember wording. Findings are grouped with a bare `<path>:` header (Model() already sorts findings by path first, so grouping cost nothing but inserting the header between runs) and each finding line drops the now-redundant path prefix.

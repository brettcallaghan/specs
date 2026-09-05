# Run record — `agent-plugins-59ic`, spec corpus anchor references

*Written by `retro-59ic` on 2026-09-05. One document: the run's results, every
finding it captured, what it routed where, and what it could not establish.*

*A companion document exists and is not this one.
`rfcs/inbox/feature-blueprint/rfc-0025-implementing-anchor-references.md` was
written by `agent-plugins-rc14.8` as the run's last work order, through the
`retro-spec` lens: it argues from nine of these findings into six proposals
against **this project's specification documents**. It is cited throughout and
deliberately not restated. What it does not carry — and what this document
adds — is the run's results, its scope, its validation, the six tooling
defects recorded outside the finding channel, the routing of all 42 findings,
and what could not be verified.*

## The run

| | |
|---|---|
| run epic | `agent-plugins-59ic` — "Sprint run: spec corpus anchor references" |
| target | `agent-plugins-rc14` — "Spec corpus anchor references", adopted whole as the run's flat target |
| sprint branch | `sprint/agent-plugins-59ic` |
| `base_sha` → head | `225751e` → `a852925`, nine commits |
| queue | 9 admitted, 0 excluded (`manifest.excluded` is `[]`), 0 moved |
| dispatches | five — `dev-59ic-a` … `dev-59ic-e` |
| partition | `.beads/plans/agent-plugins-59ic.edges.jsonl`; `concurrency_cap` 20 |
| declared checks | 12 — `format`, `vet`, `build`, `test-go`, `skills`, `shared-refs`, `strata-docs`, `schemas`, `corpus`, `ref-links`, `template-drift`, `skills-strict` |
| capability tiers | 6 `standard` → `sonnet`, 3 `advanced` → `opus`, resolved through `swarm.toml`'s `[models]`. **Which model actually served each dispatch is not recorded anywhere durable** — see "What the run could not verify" |
| effort | `unrecorded` (`manifest.effort`) |
| review passes | **1 of 3. The budget is unspent, not exhausted** — pass 1 filed two beads, admitted none, and the loop ended |
| amendments | 7 `amend:` comments on the run epic — 2 scope, 1 mid-run decision, 4 tooling |
| circuit breaker | `consecutive_dispatch_failures` 0 throughout; never armed |

**Standing authorization** (`manifest.standing_auth`) — all seven rows at the
skill default, none set in `swarm.toml`:

| row | value | source |
|---|---|---|
| `pull_request` | `draft` | default |
| `move_foreign_beads` | `false` | default |
| `admit_review_findings` | `false` | default |
| `tier_escalation` | `once` | default |
| `review_budget` | `2` | default |
| `merge_to_default` | `false` | default |
| `dispatch_failure_threshold` | `3` | default |

## Results

**All nine admitted beads closed, one commit each, across five dispatches.
Zero blocked, zero skipped, zero tier escalations, zero dispatch failures.**

| bead | tier | agent | commit | what landed |
|---|---|---|---|---|
| `rc14.3` | standard | `dev-59ic-d` | `71a80b3` | Pinned and covered the anchor extraction semantics |
| `bem` | standard | `dev-59ic-c` | `d320990` | Corrected the false cost claim about the relational reference checks and added a guard test that actually exercises the disk-reading path |
| `rc14.1` | standard | `dev-59ic-a` | `c983e8a` | Declared the two anchor finding variants and their rule constants, left unwired |
| `rc14.2` | standard | `dev-59ic-a` | `fa8268a` | Retyped the reference-bearing frontmatter properties to `Citation`, so a frontmatter citation can carry an anchor |
| `rc14.4` | advanced | `dev-59ic-a` | `17ab870` | Declared `AnchorResolution` in the index model and resolved every anchor against the target's headings |
| `rc14.5` | standard | `dev-59ic-b` | `57018ba` | Derived the link fragment from the resolved heading through `markdown.Identifier` |
| `rc14.7` | standard | `dev-59ic-b` | `6bf8988` | Answered `refs` by heading, and migrated the whole command to YAML per `cli-output-standards` |
| `rc14.6` | advanced | `dev-59ic-b` | `52da393` | Raised `reference/unresolved-anchor` and `reference/ambiguous-anchor` as corpus checks |
| `rc14.8` | advanced | `dev-59ic-e` | `a852925` | Verified the Definition of Done against the tree, corrected two FRD claims, retired the blueprint, wrote `rfc-0025` |

- No commit belongs to no bead: `225751e..a852925` is exactly nine commits and
  nine beads.
- Measured on the finished tree: `git worktree list` returns one entry, the
  repository itself — **no dispatch worktree stands**. `git status --porcelain`
  is empty.
- **No commit carries a trailer.** Measured:
  `git log --format='%H%n%b%n---' 225751e..a852925 | grep -c 'Bead:'` → `0`, and
  `git log -1 --format='%B' a852925` is the subject line alone. The cause is
  established below.

## What the run decided

Twenty-six of the 42 findings are the six work-order kinds, and the run's
dominant shape is `decision` at eleven — **the beads were accurate about the
code and silent about the edges.** Four themes, each already argued at length
in `rfc-0025`:

1. **A criterion that cannot fail.** `rc14.5#1`, confirmed independently by
   `rc14.8#5`. A resolved anchor's identifier *is* the heading's identifier by
   construction, so AC-SPECMD-001.3's two fragment sources emit identical bytes
   for every reference the matcher can produce. The criterion was proved by
   fixture and is honest; what is missing is the sentence saying it is a guard
   against a future looser matcher rather than a description of observable
   behaviour. Anyone who measures it will conclude the branch is dead.
   (`rfc-0025` P5.)

2. **Two behaviours shipped that no criterion asks for.** `rc14.7#4` — `refs
   --outbound <slug>#<anchor>` names no set and was refused with an error;
   `rc14.6#2` and `rc14.8#4` — the `unverified` severity reports under a
   *borrowed* rule identifier. Both are defensible readings; neither is
   written down, so a later agent can undo either while satisfying every
   criterion as written. (`rfc-0025` P3, P4.)

3. **An FRD example naming a condition the implementation cannot reach.**
   `rc14.4#1` hit it empirically and recorded that the FRD contradicted the
   code; `rc14.8#1` measured it and corrected the document in place. A
   parse-error member *does* carry headings; the reachable `unverified` trigger
   is a path the index carries no entry for at all. (`rfc-0025` finding 2 —
   fixed, not proposed.)

4. **A derivation keyed on a name nobody said was load bearing.** `rc14.2#1` is
   the run's most expensive single finding: retyping the reference-bearing
   properties from `Slug` to `Citation` silently emptied
   `cli/profile/registration.go`'s auto-registration, and **the corpus checker
   was completely broken** until the implementing agent found it. The delta's
   Anchors line named three files and not that one. (`rfc-0025` P1.)

Two `gap` findings are not about this feature at all and are the run's most
consequential discoveries:

- **`rc14.8#6` — `adr-012` mandates a form the corpus cannot contain.**
  `[[#Terminology]]`, a section of the referring document, resolves its anchor
  and fails its slug half every time. A corpus gating on `spec check --strict`
  — this one does — cannot use the form the ADR requires. Filed as
  `agent-plugins-clau`; `rfc-0025` P6 asks review to decide which of the two
  documents gives way. **Review pass 1 added the consequence that makes it
  worse than a failing gate:** `spec refs 'adr-800-current#Local Section'`
  returns `references: []` at exit 0 for a document that *does* carry that
  link — a wrong answer rather than a failing one, which no gate sees.
- **`rc14.8#9` — two findings were routed to a "skill lens" that nobody ran.**
  See Routing; one half of it was routed out of this project by mistake.

**There were no `sequencing` and no `unverifiable` findings.** With five
dispatches over a partitioned queue and a collision chain three beads deep,
nothing inherited a broken build and no acceptance obligation went unmet.

## Scope

**The run finished its admitted set and its scope was not narrowed by
exclusion.** `manifest.excluded` is `[]` and `manifest.moved` is `[]`. Two
amendments widened it at the gate, both human-authorized:

- `agent-plugins-bem` (P2, previously unparented) was admitted into the target
  epic. It blocks `rc14.6`; without it `rc14.6` and `rc14.8` could never become
  ready. No bead left an epic — `bem` had no parent.
- `agent-plugins-rc14` (previously unparented) was adopted as the run's flat
  target. Its children did not move.

**Where the scope IS narrower than the run's own evidence, and it must be said
plainly:** `admit_review_findings` was `false`, so review pass 1's two findings
were **filed and not worked**. Both remain open:

| bead | priority | what it is |
|---|---|---|
| `agent-plugins-59ic.7` | P1 | `Report.Ran` sets `Passed` whenever `state.failed` is false, and `Report.Add`'s unverified branch never sets `failed` — so an unverified outcome is overwritten and its reason wiped. Measured end to end: one report printing both `reference/unresolved-anchor unverified … could not run` and `reference/unresolved-anchor passed`. **This run created the defect's first reachable emitter**, so the run owns it |
| `agent-plugins-59ic.8` | P2 | `specs/guides/spec-cli-reference.md` still says "an anchor is split off, not resolved" and names four reference rules where the tool now has six. Its vendored copy under `plugin/skills/spec/references/shared/` is generated, so it is one defect with one remedy — and that copy ships to every project that installs the `spec` skill |

`59ic.7` is the sharper of the two: **the run shipped a check whose "passed"
line can be vacuous**, and it shipped it green.

Three further beads were touched by findings and deliberately not worked:
`agent-plugins-bieu` (P2, `document-core` contradicts itself about Markdown
Structure — `rc14.5#3`, `rc14.8#8`), `agent-plugins-clau` (P2, above), and
`agent-plugins-z8r` (P4, deferred, untouched — verified by `rc14.8#3`).

## What the coordinator recorded

Seven `amend:` comments on the run epic. **Four are tooling defects, and none
of the four is a `retro:` finding** — so the harvest query in this skill cannot
see them. They survive into this document only because the coordinator handed
them over in the dispatch prompt. That is itself a finding, and it is
candidate **D** below.

1. **`swarm dispatch accept` folds into whatever `HEAD` is on, not into the
   manifest's branch — and reports success.** With the main checkout on
   `claude/swarm-agent-plugins-rc14-lclbkc` it fast-forwarded *that* branch,
   left `sprint/agent-plugins-59ic` at base, and still returned `merged: true`
   with `merged_sha` set to the pre-merge target head. Only the ancestry guard
   (`ancestor: false`, exit 2) caught it. Repaired by checking the repository
   out onto the sprint branch, resetting the wrongly-advanced branch to
   `225751e`, and re-running accept. **This is the run's most serious tooling
   defect: a silent write to the wrong branch, self-reported as a success.**
2. **`swarm dispatch accept`'s concurrent path rebases a branch still checked
   out in its own worktree**, which git refuses: `fatal: 'bb/dev-59ic-c' is
   already used by worktree`. Worked around by rebasing from inside
   `/home/user/.bb-worktrees/dev-59ic-c`, after which accept saw the serialized
   case and completed.
3. **`swarm dispatch accept` without `--no-push` was blocked by this session's
   permission classifier** at its git push step. Run with `--no-push`; the
   branch was pushed with plain git at post-flight.
4. **`README.md`'s generated coverage table conflicted between two dispatches.**
   `dev-59ic-a` and `dev-59ic-c` both regenerated it, so **neither side of the
   conflict was correct**. Resolved by taking one side to complete the rebase,
   re-running `scripts/coverage.sh` on the combined tree, and folding the
   result into `rc14.4`'s commit.

The remaining three record scope (the two admissions above) and one mid-run
decision: **`bem`'s Design conditional was resolved mechanically by the plan's
own measurement** — 0 of 693 references resolve to a retired target and
`pairings` is empty, so the relational checks do zero disk reads on this
corpus and the defect is latent. The cheap branch was taken and the two
structural alternatives recorded as rejected on the bead. Not escalated,
because a mid-run decision is a finding and this is where it surfaces.

Two more defects reached this run through the manifest's `ledger` and through
a finding rather than through an amendment:

5. **`groom-beads`' `scripts/bd-health.sh` exits 1 on a healthy graph.**
   Re-measured here at head: exit 1, `RESULT: hard problems found in: swarm
   validate`, with **23** `is not an epic or molecule (type: task)` errors —
   it runs `bd swarm validate` over every issue in the repository and that
   subcommand accepts only epics and molecules. Its exit code cannot gate
   anything. **Already filed three times** (`agent-plugins-7jxt`,
   `agent-plugins-84yu`, `agent-plugins-prnv`); nothing new to file.
6. **`scripts/regenerate-cli-snapshots.sh` already regenerates the fixtures
   three agents hand-patched** — `rc14.1#4`, `rc14.2#3`, `rc14.4#4`, measured
   as zero-diff on the finished tree by `rc14.8#7`. Pure wasted effort, three
   times over, and `rfc-0025` P2 proposes the documentation fix.

One hazard the run avoided rather than hit, recorded because the next run may
not: a dispatch agent left a stray untracked generated index
(`cli/testdata/cases/check-anchor-defects/input/docs/specs.json`) inside a
tracked fixture's input directory, and the review deleted it. That directory
now holds `adrs` alone, measured. It matters because `bb-close-bead.sh` stages
`git add -A -- ':!.beads/*.jsonl'` when the caller names no paths
(line 236) — a close ceremony taking that branch would have committed it.
Another agent edited the main checkout by mistake, caught it with
`git status`, and reverted; no trace remains.

## Every finding, verbatim

**42 findings across nine beads**, harvested with one query over the run's
closure set:

```bash
bd export | jq -r 'select(.id|test("^agent-plugins-(rc14(\\.[0-9]+)?|bem)$"))
  | .id as $i | (.comments//[])[] | select(.text|startswith("retro:"))'
```

Each is quoted whole, unedited. Ids are derived — `<bead>#<ordinal>`, the
ordinal being the finding's position among that bead's own `retro:` records.

### `decision` — 11

**`rc14.8#4`**

> retro:decision Confirmed rc14.6's choice to report AC-SPECANCH-002.4's
> unverified severity under `reference/unresolved-anchor` is what the FRD
> asks for, and proposed recording it rather than leaving it inferred.
> evidence: AC-SPECANCH-002.9 requires that the CLI "reports an anchor
> matching no heading and an ambiguous anchor under stable rule
> identifiers of their own" — two identifiers, both for error cases.
> AC-SPECANCH-002.4 requires "reports that reference with severity
> unverified naming what is missing" and names no identifier; the block's
> Examples paragraph says it reports "per AC-SPEC-006.3 and AC-SPEC-006.4",
> which is the generic Unverified mechanism rather than a rule. So the
> borrowed identifier conforms. What the FRD does NOT say is that it is
> borrowed, and a later agent adding `reference/unverified-anchor` would
> satisfy every criterion as written while breaking every caller selecting
> the anchor checks with one `--only`. Recorded as P4 of
> rfcs/inbox/feature-blueprint/rfc-0025-implementing-anchor-references.md
> rather than edited into the FRD, because it adds a statement to an
> approved FRD rather than correcting a false one.

**`rc14.8#5`**

> retro:decision Verified rc14.5's "the AC is unobservable on real input"
> reasoning holds, and that the criterion is still load bearing.
> evidence: cli/core/references/anchor.go matches by
> `markdown.Identifier(heading.Text) == wanted` where `wanted :=
> markdown.Identifier(*ref.Anchor)`, so a resolved anchor's identifier IS
> the heading's and the two fragment sources cannot differ for any
> reference the matcher can produce. The tests are still honest, because
> cli/core/mdlinks consumes a settled `AnchorResolution` and never
> re-derives the match: `anchorFragment` in cli/core/mdlinks/mdlinks.go
> branches on `ref.AnchorResolution` and takes the heading's text where
> one resolved. A fixture stating a divergent heading therefore exercises
> a real branch of shipped code, not a hypothetical one. The risk the
> reasoning creates is the opposite of an unverified claim: an agent who
> measures the same thing would conclude the branch is dead and collapse
> it back onto the authored anchor, which no corpus-input test would
> catch. Filed as P5 of rfc-0025.

**`rc14.7#1`**

> retro:decision The bead's 'output conforms to cli-output-standards' pulled the whole refs command off its formatted-line rendering, not just the anchored form.
> evidence: specs/standards/cli-output-standards.md's conformance table row for refs read 'prints today: path:line:col  slug per line; --json a compact object | to conform: render the sequence of references as YAML mappings', and the paragraph under it reads 'Migration is per command and belongs to whichever feature next touches it.' This bead is that feature. Decided: migrate refs entirely — YAML by default, JSON under --json, one mapping per reference, and the two prose outcome lines ('X: referenced by no corpus member', 'X: not declared by any corpus member') replaced by the declared/references properties that carry the same two facts. Rejected: rendering only the anchor-qualified form as YAML, which would have left one command printing two grammars depending on its argument. Four committed snapshot artifacts changed as a result (refs-inbound-default, refs-not-referenced, refs-outbound, refs-undeclared), plus help-refs and help-root for the new argument form; refs-json is byte-identical, which is the check that the JSON rendering did not regress.

**`rc14.7#5`**

> retro:decision The 'marked as matching more than one heading' mark is a new ambiguous_anchor property on the answer, not a new property of the generated Reference.
> evidence: AC-SPECANCH-003.4 requires the mark; specmodel.Reference already carries anchor_resolution with state 'ambiguous' and every match, so the fact is present but implicit. Decided: embed specmodel.Reference in a command-local refsMatch and add 'ambiguous_anchor' beside it, written only when true (omitempty). The mark is read off the reference's own resolution rather than decided a second time. Rejected: adding the property to schemas/index.strata, which would put a query's presentation concern in the index every consumer parses; and leaving the reader to infer it from anchor_resolution.state, which cli-output-standards forbids in spirit — 'the reader is an agent that knows which command it ran and nothing else'.

**`rc14.6#2`**

> retro:decision AC-SPECANCH-002.4's unverified severity is reported under the existing reference/unresolved-anchor identifier rather than a third rule of its own.
> evidence: rc14.1 declared exactly two rule constants (RuleUnresolvedAnchor, RuleAmbiguousAnchor in cli/core/finding/finding.go) and two finding variants; the CHK delta's third bullet — 'WHERE the target document's headings are unavailable, the CLI SHALL report the reference with severity unverified naming what is missing' — names no third identifier, and the FRD's example reads 'reports unverified per AC-SPEC-006.3 and AC-SPEC-006.4', which is the generic finding.Unverified mechanism rather than a rule. Decided: raise finding.Unverified(RuleUnresolvedAnchor, ...) carrying the resolution's own reason string. Consequence recorded in the code: RuleUnresolvedAnchor is now the one identifier in cli/core/finding that reports at more than one severity, and finding.go's standing comment 'No constant here names a check that reports SeverityUnverified ... this sprint registers no check that uses it' was replaced rather than left to rot. Rejected: a third rule id, which would have needed a schema change and a third finding variant for a case that is 'the same check, no result' rather than a different defect.

**`rc14.5#1`**

> retro:decision The resolved-heading fragment and the authored-anchor fragment cannot differ on any corpus input today, so the AC is only observable through a hand-built fixture.
> evidence: cli/core/references/anchor.go resolves an anchor by 'markdown.Identifier(heading.Text) == wanted' where 'wanted := markdown.Identifier(*ref.Anchor)'. Equality of identifiers is the matching rule, so a RESOLVED anchor's identifier is by construction the heading's identifier, and 'derive the fragment from the heading' and 'derive it from the anchor' emit the same bytes for every reference the matcher can produce. Decided: prove AC-SPECMD-001.3 with fixtures that state a heading whose text differs from the anchor (TestAResolvedAnchorTakesItsFragmentFromTheResolvedHeading in cli/core/mdlinks/mdlinks_test.go, e.g. anchor 'Reference Integrity Checks' resolved to heading 'FEAT-SPECREF-002: Reference Integrity Checks' -> fragment 'feat-specref-002-reference-integrity-checks'). Legitimate because mdlinks is a consumer of a settled resolution and never re-derives or re-checks the match — mdlinks_test.go's own header already states that fixtures assert the outcome rather than reproduce the corpus that produced it. Rejected: an end-to-end case through 'spec links', which would have proved nothing, since no corpus input can separate the two sources. The criterion still earns its keep — it is what stops a future looser matcher (a leading requirement code, a partially-named heading) from silently emitting fragments that reach nothing.

**`rc14.4#1`**

> retro:decision The FRD/blueprint text is ambiguous about what runtime case triggers 'the target document's headings are unavailable' (AC-SPECANCH-002.4) — its own illustrative example ('a target whose frontmatter failed to parse had no headings extracted from it') contradicts corpus.go, which calls headingsOf(src, doc) even when frontmatter has a ParseError, so a member with a parse error still carries headings. Worked out empirically instead: corpus.go records NO IndexEntry at all for a candidate excluded for no-frontmatter or unknown-type, and for a standalone .feature file (references.go's own package doc: 'a feature source is no candidate... though the file is not a corpus member'). So 'unavailable' means 'no IndexEntry exists for this path at all' (built as a map[path][]Heading from index.Entries; a missing key means unverified), never 'entry exists with zero headings' (that's a legitimate no-match, reported unresolved not unverified). Verified live: a self-anchor  written inside a standalone .feature file resolves to anchor_resolution.state=unverified with reason naming the path — the concrete, reachable trigger.

**`rc14.4#2`**

> retro:decision Designed the new schemas/index.strata shapes: AnchorResolution (discriminated union, tag 'state': resolved/unresolved/ambiguous/unverified) added as a new optional anchor_resolution field on Reference, reusing the AnchorMatch struct rc14.1 already declared in schemas/finding.strata (text+position) rather than declaring a second heading-match shape — strata resolves types across files in the same schema src dir (Span already crosses finding.strata/index.strata this way), so one AnchorMatch serves both the index model and the finding model. Reference.anchor_resolution is nil (omitted from JSON) both when the reference carries no anchor and when its slug half is present but did not resolve (AC-SPECANCH-002.8, one mistake one finding) — verified this by grepping cli/testdata fixtures for 'anchor_resolution' after regenerating: it appears only on references that actually resolved a target.

**`rc14.2#4`**

> retro:decision Added a new snapshot case (cli/testdata/cases/index-citation-anchor) to prove AC-SPECANCH-001.6 at the integration level, per the bead's own citation obligation and specs/standards/testing-standards.md's Snapshot Case Structure — no existing case exercised an anchor on a frontmatter-property citation. Modeled on index-unresolved-reference's corpus shape (docs/adrs/ + docs/features/) with args=["index"] so both checks.log and manifest.json are asserted.

**`rc14.1#1`**

> retro:decision Added the two anchor variants' Go constructors and RuleId constants (UnresolvedAnchor/AmbiguousAnchor, reference/unresolved-anchor and reference/ambiguous-anchor) in cli/core/finding in this bead, per the blueprint's FIND-delta Anchors line naming 'the constructors in cli/core/finding' as in scope. Left them unwired: not added to cli/cmd/check.go's checkRules registry nor to finding.corpusWideRules, since nothing raises these findings until the CHK bead (agent-plugins-rc14.6) implements the actual checks — registering an unused rule there would misreport it as 'passed' in every run's checks accounting before any check produces it.

**`rc14.1#2`**

> retro:decision Named FindingAmbiguousAnchor's target field target_path (a repo-relative path, matching FindingIncompletePairing/FindingDuplicateSlug convention) even though the FIND delta's Behavior bullet only explicitly names it for the unresolved-anchor variant ('holding the anchor as authored and the target document'). Included it on the ambiguous variant too since a consumer needs to know which document to open to see the matches, and the FRD's own AC-SPECANCH-002.9 example format ('no heading named X in Y') implies the target is always named.

### `gap` — 6

**`rc14.8#6`**

> retro:gap The slug-less anchor form adr-012 mandates is reported as an
> error by a check that predates this feature; anchor resolution is what
> made the contradiction visible. Filed as agent-plugins-clau, not fixed.
> evidence: specs/adrs/adr-012-reference-conventions.md:54-55 — "A section
> of the referring document itself is referenced by an anchor with the
> slug omitted: `[[#Terminology]]`." AC-SPECANCH-002.6 resolves that
> form's anchor against the containing document, and does. The slug half
> is a separate check and fails every time. Measured on a throwaway
> corpus whose adr-098 document contains `[[#Overview]]` above its own
> `## Overview` heading: `spec index --out -` gave
> `anchor_resolution={"state":"resolved","heading":{"text":"Overview",...}}`
> on that reference, and `spec check --strict` gave
> `docs/good.md: reference/unresolved  error  14:19  [[#Overview]] matches
> no document's slug`. So the tool reports the anchor resolved and the
> reference broken in one run, and a corpus gating on --strict — this one
> does — cannot contain the form the ADR requires. No corpus member uses
> it today (all eleven anchored references carry a slug), so no gate
> fails. The fix is a decision between exempting the form in
> `reference/unresolved` and withdrawing it from adr-012; recorded as P6
> of rfc-0025 and tracked as agent-plugins-clau.

**`rc14.8#9`**

> retro:gap Two findings from this run belong to the SKILL lens and are
> not in rfc-0025; nobody has run retro-skill for them.
> evidence: (1) agent-plugins-rc14.4 recorded that an `advanced`-tier bead
> landed on a dispatch also holding two `standard` ones because a shared
> collision mark (generated:cli/core/schema/spec-model.schema.json)
> serialized them, and asked whether tier is meant to gate assignment —
> that is a question about swarm-plan, whose fix lands in the repository
> that vends the skill. (2) The durable form of rfc-0025's finding 1 is a
> constraint on the feature-blueprint TEMPLATE — "a delta that renames a
> declaration lists every consumer that names it by name" — which also
> lands there. Both are recorded under rfc-0025's Considered and Not
> Proposed with the routing stated, so they are not lost, but no
> skill-lens RFC exists for them and this bead did not author one: the
> bead's Design names retro-spec only.

**`rc14.7#4`**

> retro:gap Neither the bead nor the CMD delta says what 'refs --outbound <slug>#<anchor>' means, and the combination is reachable from the command line.
> evidence: the CMD delta in specs/features/reference-integrity/anchors/anchor-references-blueprint.md states only 'WHEN refs is given an argument qualified by an anchor, the CLI SHALL list every reference whose anchor resolves to that heading' — every AC-SPECANCH-003.x scenario is inbound. The existing --outbound flag lists the references OUT of the named document, whose anchors name headings of other documents, so 'the references out of X into heading Y of X' names no set. Decided: refuse the combination with an error and exit 1 (errNoAnchorWithOutbound in cli/cmd/refs.go), covered by the refs-anchor-outbound-rejected snapshot case, following the errNoEmitWithJSON precedent already in cli/cmd/headings.go for 'the one flag combination that has no meaning'. Rejected: silently ignoring the anchor, which would answer a different question than the one typed; and inventing outbound-anchor semantics (references out of X whose anchor resolves to a heading named Y in THEIR targets), which is a feature nobody specified.

**`rc14.6#4`**

> retro:gap Nothing said where the finding gets the TARGET DOCUMENT's name from, and the obvious answer would have been a second derivation of what the anchor was matched against.
> evidence: finding.UnresolvedAnchor(path, at, anchor, targetPath) and finding.AmbiguousAnchor(...) both take a targetPath, but specmodel.Reference carries no such field — AnchorUnresolved is 'struct{}' and AnchorAmbiguous carries only its matches, so the check would have had to recompute 'ref.Slug == \"\" ? ref.Path : resolution.TargetPath'. That rule lived privately inside resolveAnchor in cli/core/references/anchor.go. Decided: export it as references.AnchorTarget(ref) (string, bool) and have resolveAnchor call it, so one derivation decides both what an anchor is matched against and what the finding names. Rejected: copying the four lines into cli/core/corpus/check.go, which would be free to disagree with the matcher — the same defect rc14.5 was raised to remove between mdlinks and markdown.

**`rc14.2#1`**

> retro:gap The bead's Design/AC list named schemas/vocabulary.strata, schemas/frontmatter.strata and scripts/generate-schemas.sh as the anchors, but retyping every reference-bearing property from Slug to Citation silently broke cli/profile/registration.go's derivation, which auto-registers reference-bearing properties by matching the JSON schema $ref "#/definitions/Slug". After the retype nothing matched, so every document type's ReferenceProperties came back empty, which cli/core/corpus.ValidateRegistration then reported as a fatal config/invalid-registration error at check time (measured: 'profile registration: "feature-blueprint.frd" is registered as a typed reference but not as a reference-bearing property') — the corpus checker was completely broken until this was found and fixed. Neither the bead nor the blueprint's SREF-delta Anchors line named cli/profile/registration.go, even though the AC 'module-spec-profile records the citation type in its registered frontmatter schema' pointed straight at it. Fixed by retargeting the derivation (and its drift-guard test in registration_test.go) at '#/definitions/Citation' instead.

**`bem#1`**

> retro:gap: TestWarmQueriesAndChecksMeetTheirBudgets builds its 1,000-doc corpus from a helper emitting 'status: approved' and no references, so the relational path is never taken. The budget is green because the code under test never runs, not because it is cheap. Any guard test this bead adds must exercise the relational path or it inherits the same blind spot.

### `stale` — 5

**`rc14.8#1`**

> retro:stale The FRD illustrated AC-SPECANCH-002.4 with a runtime
> condition the implementation cannot reach; corrected in place.
> evidence: specs/features/reference-integrity/anchors/anchor-references-frd.md
> read "A target whose frontmatter failed to parse had no headings
> extracted from it, so a reference into it reports `unverified`". Both
> halves are false. cli/core/corpus/corpus.go:453-454 sets
> `Headings: headingsOf(src, nil)` on the parse-error branch and :502-503
> sets `Headings: headingsOf(src, doc)` on the envelope-error branch.
> Measured on a throwaway corpus (a document whose frontmatter has an
> unterminated quoted title, plus one citing a heading of it):
> `spec index --out - | jq '.entries[]'` -> `docs/bad.md parse_error=true
> headings=2`, and the citing reference came back
> `resolution={"state":"unresolved"}` with `anchor_resolution=null` — the
> parse-error member decoded no slug, so the parent feature's slug rule
> fires and AC-SPECANCH-002.8 suppresses any anchor finding. The condition
> that DOES reach unverified is a path the index carries no entry for:
> same corpus, `[[#Probe]]` written in a standalone docs/probe.feature
> gave `reference/unresolved-anchor  unverified  4:48  check
> reference/unresolved-anchor could not run: no headings are recorded for
> docs/probe.feature`. agent-plugins-rc14.4 reached this empirically and
> recorded that the FRD contradicted the code. Replaced the example with
> the reachable case rather than filing it, per the RFC template's own
> scope note ("The fix is to one document's CONTENT ... just fix the
> document").

**`rc14.8#2`**

> retro:stale The count of anchored references was wrong in three places
> and each carried a different number; measured, it is eleven.
> evidence: the epic and this bead say nine; the FRD's Overview said "Nine
> references in the corpus name an anchor — four into
> `container-spec-cli#Command Surface`, five into component rosters of
> [[module-document-core]]"; agent-plugins-rc14.6 measured eleven. Did not
> inherit any of them. Measured from a freshly built index over this
> repository: `spec index --out - | jq '[.references[] | select(.anchor !=
> null and .anchor != "")] | length'` -> 11 (of 697 references), and
> grouping by target gives six into `container-spec-cli#Command Surface`
> (structure-, gherkin-adoption-, anchor-references-, reference-integrity-
> and markdown-links-blueprint, the last twice) and five into
> module-document-core rosters (Findings, Frontmatter, Numbering, Schema
> Registry, Spans — all in gherkin-adoption-blueprint). Every one reports
> `anchor_resolution.state == "resolved"`. The FRD's own figure went stale
> during its own implementation: this feature's blueprints added the two
> extra `#Command Surface` citations. Corrected the FRD in place to
> eleven/six/five.

**`rc14.6#1`**

> retro:stale The bead says nine anchored references in this corpus; there are eleven, and all eleven resolve, so the strict gate passed with nothing to repair.
> evidence: measured from a freshly built index — 'spec index --out -' over this repository, filtered to references carrying an anchor: 11 anchored references out of 697, every one anchor_resolution.state == resolved (six in gherkin-adoption-blueprint.md, and five '[[container-spec-cli#Command Surface]]' in markdown-links-, anchor-references-, reference-integrity- and structure-blueprint.md). 'spec check --strict' over this repository exits 0 with 'no findings' and both new rules 'passed'. The count moved because beads landing during this run added anchored references to the blueprints; the bead's figure was taken at grooming.

**`rc14.5#3`**

> retro:stale The module document's Markdown Links relationship paragraph sits beside three claims about Markdown Structure that the heading work has already falsified; filed rather than fixed.
> evidence: 'The Corpus -> Markdown Structure edge is still the intended shape rather than an existing one: nothing in Corpus calls `cli/core/markdown` yet' in specs/modules/document-core/document-core-module.md, contradicted by 'grep -n "markdown\." cli/core/corpus/corpus.go cli/core/corpus/features.go' -> 'corpus.go:556: found := markdown.HeadingsFrom(src, from)' and 'features.go:232: for _, block := range markdown.Fenced(...)'. Also 'Go package (fences built)' in the component's diagram entry, and 'Markdown Structure -> Spans is now real for the fenced-block half'. Filed as agent-plugins-bieu (discovered-from rc14.5), adjacent rather than blocking.

**`bem#4`**

> retro:stale The module doc specs/modules/document-core/document-core-module.md carried the same false claim the bead's Design targeted in cli/core/corpus/check.go's comment: 'Three rules walk the same already-resolved reference collection reference/unresolved reads, so each costs a lookup rather than a second corpus walk.' AC #8 on this bead (module-document-core's Corpus component described truthfully) required finding and fixing this copy too; a grep for the phrase 'lookup over data already in memory' / 'so each costs a lookup' across the tree turned up only this one additional instance beyond check.go, now corrected in the same commit.

### `scope` — 4

**`rc14.8#8`**

> retro:scope agent-plugins-bieu overlaps this fold and was left alone, as
> instructed. Its subject is one of the two documents this bead verified.
> evidence: specs/modules/document-core/document-core-module.md:1627 still
> reads "The Corpus -> Markdown Structure edge is still the intended shape
> rather than an existing one: nothing in Corpus calls `cli/core/markdown`
> yet", while the same document's Components roster (line ~276) reads
> "Both edges from Corpus into Markdown Structure and Gherkin are
> exercised" — the document contradicts itself. Verified this feature's
> own fold is complete regardless: the Markdown Structure section records
> `Identifier` and why anchor resolution matches through that one
> derivation, the component diagram gained `Rel(mdlinks, md, ...)`, and
> the relationships paragraph for Markdown Links names Markdown Structure.
> The stale claims were falsified by the structure feature (commit
> 225751e), not by this one, so the DoD item "Markdown Structure updated"
> holds and bieu remains the owner of the contradiction. Not implemented.

**`rc14.4#5`**

> retro:scope Tier was 'advanced' for this bead versus 'standard' for rc14.1 and rc14.2, all three assigned to the same dispatch (dev-59ic-a) via the shared collision mark generated:cli/core/schema/spec-model.schema.json. Proceeded rather than stopping, since the tracker had already assigned rc14.4 to this dispatch id and the dispatch prompt itself named it as part of the same serialized chain ('shared with rc14.2 and rc14.4'). Flagging in case tier is meant to gate which dispatch a bead lands on rather than only inform effort estimation — if so, the collision-chain assignment and the tier gate are in tension and a future swarm-plan run should reconcile them explicitly.

**`rc14.2#5`**

> retro:scope Renamed cli/profile's drift-guard test TestEverySlugTypedPropertyInTheVocabularyIsRegistered to TestEveryCitationTypedPropertyInTheVocabularyIsRegistered (and its helpers slugTypedPropertiesOf/mentionsSlug to citationTypedPropertiesOf/mentionsCitation) since it now asserts against the Citation definition, not Slug — the old names would have read as accurate while checking a different thing than what they said. Updated the one doc citation of the old test name in specs/modules/spec-profile/spec-profile-module.md. Did not rename TestTheProfileRegistersEverySlugHoldingProperty (registration_test.go) or its registeredReferenceProperties table, since those still correctly describe 'every property that (transitively) holds a document slug' — a citation still resolves to a slug — and renaming it was not needed to make the test pass or accurate.

**`rc14.3#1`**

> retro:scope: c4_changes declares component-references while this bead's acceptance criteria are pure coverage over shipped behaviour. Over-declared; left as-is because it only makes collision detection more cautious and a real edge already serializes it with rc14.4.

### `tooling` — 5

**`rc14.8#7`**

> retro:tooling Three agents in this run hand-patched the same five
> snapshot fixtures; the regeneration script already covers them.
> evidence: agent-plugins-rc14.1, .2 and .4 each recorded diffing a fresh
> run against the committed fixture and replacing the `input_hash` line by
> hand in cli/testdata/cases/{index-default,index-features,
> index-out-custom,index-out-stdout,index-unresolved-reference}/expected/
> manifest.json, because a schema edit changes opts.Profile.Schema()'s
> bytes and cli/core/corpus/hash.go folds them into every index's
> input_hash. rc14.4 recorded "the pattern is now three-for-three".
> Measured here: scripts/regenerate-cli-snapshots.sh's own header says it
> "writes expected/checks.log and expected/manifest.json for each case",
> and running it on the finished tree exited 0 with `git status
> --porcelain -- cli/testdata/cases` empty — so the script regenerates
> exactly the field they patched, and the committed fixtures are current.
> The hand-patching was avoidable and nothing an implementing agent reads
> says so; specs/standards/testing-standards.md's Snapshot Case Structure
> requires the script to exist and does not say what invalidates an
> artifact. Filed as P2 of rfc-0025.

**`rc14.7#2`**

> retro:tooling:project The YAML renderer did not honour json.Marshaler, so a union rendered its Go shape and the two renderings disagreed — a latent defect this bead is the first command answer to hit.
> evidence: measured before the fix, renderYAML over a specmodel.Reference printed 'origin:\n  Value: {}' and 'resolution:\n  Value:\n    target_path: docs/b.md' where json.Marshal of the same value produced '"origin":{"kind":"body"}' and '"resolution":{"state":"resolved","target_path":"docs/b.md"}' — the variant tag absent and a Go field name in its place. cli/cmd/yaml.go's own header claimed the opposite: 'The renderer walks the same struct encoding/json marshals ... deriving one rendering from the other's declaration is what makes it true by construction'. headings, the only prior YAML command, carries no union and never exposed it. Fixed in the renderer (marshaledValue + an order-preserving JSON decode) rather than by projecting Reference into hand-written command-local fields, because cli-output-standards' own References cite [[adr-016]] — 'a rendering is a view over a generated shape rather than a second declaration of it' — and a hand projection is that second declaration. Rejected alternative cost: ~40 lines of switch in cmd/refs.go duplicating the generated model's JSON shape, silently drifting whenever schemas/index.strata changes.

**`rc14.4#4`**

> retro:tooling Same class of fixture staleness as rc14.1 and rc14.2, third time in this collision chain: the schema change (adding AnchorResolution etc.) changed opts.Profile.Schema()'s bytes, invalidating the same input_hash fixtures again, plus index-citation-anchor (from rc14.2) needed its whole manifest regenerated since it now genuinely gains an anchor_resolution field (verified the diff was exactly the new field plus the hash, nothing else). Patched all 6 the same way. No further beads in this collision chain remain assigned to this dispatch (rc14.5 CHK/MDL/CMD deltas are on other dispatches), but the pattern is now three-for-three and worth a swarm-plan note: any bead regenerating the bundled schema should expect to touch these fixtures.

**`rc14.2#3`**

> retro:tooling Same class of fixture staleness as rc14.1: retyping to Citation changed opts.Profile.Schema()'s bytes again, invalidating the same 5 input_hash fixtures a second time in this run (index-default, index-features, index-out-custom, index-out-stdout, index-unresolved-reference). Patched the same way — diffed to confirm only input_hash changed, then replaced just that line. A bead later in this collision chain (rc14.4, also touching the bundled schema) should expect this too.

**`rc14.1#4`**

> retro:tooling The schema edit changed opts.Profile.Schema()'s bytes, which cli/core/corpus/hash.go folds into every index's input_hash. That broke TestSnapshotCases, TestCRLFCheckoutProducesIdenticalOutput and TestBinaryWritesTheIndexToStandardOutput — 5 fixtures under cli/testdata/cases/{index-default,index-features,index-out-custom,index-out-stdout,index-unresolved-reference}/expected/manifest.json carry a stale input_hash. Confirmed via diff that only the input_hash line differed (rest of each fixture byte-identical to a fresh run), then patched just that line in each. Any future bead that regenerates the bundled schema should expect this and check those 5 fixtures — nothing in AGENTS.md or the blueprint calls this out.

### `proof` — 10

**`rc14.8#3`**

> retro:proof Every Definition-of-Done item verified against the tree
> rather than against the beads' reports.
> evidence: (1) AC enumeration from the tool, not the FRD: `spec feature
> FEAT-SPECANCH-001` / `-002` / `-003` -> 4 + 8 + 4 = 16 criteria. (2)
> Citations: `grep -rn "AC:[[:space:]]*AC-SPECANCH" --include='*.go' cli/`
> -> 11 hits, all in test files, covering 001.1/.2/.5 and 002.1/.2/.4/.5/
> .6/.7/.8/.9 — every criterion the coverage table declares `unit`.
> `grep -rn "AC-SPECANCH" cli/testdata/` -> 8 case.toml `cites` lines
> covering 001.6, 002.1/.5/.9, 003.1 (twice), 003.2, 003.3, 003.4 — every
> criterion the table declares `integration`. No criterion is uncited and
> no citation names a retired identifier. (3) AC-SPECMD-001.3 and .7:
> `grep -rn "AC:[[:space:]]*AC-SPECMD-001\.(3|7)"` -> mdlinks_test.go:324,
> :426 (.3) and :278, :377 plus anchor_edges_test.go:36 (.7). (4)
> `spec check --strict` over this repository: exit 0, "no findings", with
> `reference/ambiguous-anchor  passed` and `reference/unresolved-anchor
> passed` both present in the per-rule list — 25 rules where the baseline
> had 23. (5) Tracker Reconciliation: agent-plugins-bem CLOSED,
> agent-plugins-z8r DEFERRED, agent-plugins-im99 OPEN; none subsumed, none
> touched. (6) Retirement is safe: `spec refs
> feature-spec-anchor-references` -> `references: []`, so no live document
> cites the blueprint and `reference/retired-target` cannot fire on it.

**`rc14.7#3`**

> retro:proof The renderer fix is pinned by two tests with a measured negative control.
> evidence: TestAValueThatMarshalsItselfRendersFromItsOwnJSON and TestTheTwoRenderingsCarryTheSameKeysInTheSameOrder in cli/cmd/yaml_test.go. Negative control: with 'if !ok || true { return v }' substituted into marshaledValue, 'go test ./cmd/ -run "TestAValueThatMarshalsItself|TestTheTwoRenderings"' exits 1 and prints the pre-fix rendering ('origin:\n  Value:\n    property: superseded_by'); with the line restored it exits 0. A third test covers the fallback path — a union holding no variant, whose MarshalJSON errors, still renders from its Go fields rather than vanishing.

**`rc14.6#3`**

> retro:proof Both new rules and the corpus-wide exception are pinned by tests with a measured negative control.
> evidence: cli/core/corpus/anchor_check_test.go, seven tests. Negative control: replacing 'report.Add(anchorFindings(ref)...)' in cli/core/corpus/check.go with '_ = anchorFindings(ref)' makes 'go test ./core/corpus/ -run Anchor' exit 1 with five of the seven failing (the two that do not are the corpus-wide-rule assertion and the one-mistake-one-finding negative, which is correct: neither depends on the wiring). Restored, it exits 0. The unverified case is exercised through a real corpus rather than a hand-built Reference: a '.feature' file's references are extracted against its own path, which is a feature source and not an index entry, so '[[#Decision]]' written there has no recorded headings to match against — the exact case cli/core/references/anchor.go's headingsByPath comment names.

**`rc14.5#2`**

> retro:proof Consuming markdown.Identifier changed two observable fragments on the authored-anchor fallback; both were caught by pre-existing tests failing before the fixtures were updated.
> evidence: after replacing slugifyAnchor with markdown.Identifier and before touching any test, 'go test ./core/mdlinks/...' printed exactly two failures — 'leading_and_trailing_spaces_become_hyphens' and 'an_anchor_of_only_spaces_survives_as_hyphens' (cli/core/mdlinks/anchor_edges_test.go:59). Measured with a probe over markdown.Identifier: ' Leading And Trailing ' -> 'leading-and-trailing' (was '-leading-and-trailing-'), '   ' -> '' (was '---'). Cause: Identifier derives from RENDERED inline content, and goldmark trims whitespace around a paragraph before inline rendering, so surrounding spaces are never characters the rule sees. Every other case in that table is byte-identical under both derivations ('Section 12.4 Detail' -> 'section-124-detail', 'Hello 🎉 World' -> 'hello--world', 'Tab\tSeparated' -> 'tabseparated', 'Раздел Один' -> 'раздел-один'). The new behaviour is the correct one: it is exactly what the matcher looks an anchor up under, so a fragment and its lookup key can no longer disagree.

**`rc14.4#3`**

> retro:proof Verified end-to-end with throwaway corpora (not committed, /tmp): (1) a frontmatter citation and three body-reference forms — an exact-case match, an ambiguous heading (two 'References' headings), a nonexistent anchor, a self-anchor '[[#Heading]]', and an embed '![[...]]' of a resolved anchor — all produced the expected anchor_resolution state (resolved/ambiguous/unresolved/resolved-against-self/resolved-same-as-link respectively); (2) a standalone .feature file's self-anchor produced state=unverified. All six of the bead's ACCEPTANCE CRITERIA bullets were exercised this way before being pinned as cli/core/references/anchor_test.go unit tests citing AC-SPECANCH-002.2, 002.6 and 002.7 per the item's own citation obligation.

**`rc14.2#2`**

> retro:proof Verified end-to-end with a throwaway corpus (adr-012-decision.md with a Decision heading, an frd declaring sources: ["adr-012-decision#Decision"]): 'spec check' exits 0 and 'spec index' records the reference with slug and anchor as separate fields, resolved to the target path — matching AC-SPECANCH-001.6 exactly. Also verified the two negative cases explicitly: a trailing '#' with no anchor (sources: ["adr-012#"]) fails frontmatter/schema ('does not satisfy $ref'), and a document's own slug carrying an anchor (slug: adr-013#Decision) still fails both frontmatter/schema and slug/pattern-mismatch — confirming the citation type admits exactly the two intended shapes and the identity property is unaffected.

**`rc14.3#2`**

> retro:proof Ran scripts/coverage.sh --check from the repo root after adding the new tests (cli/**-required per AGENTS.md): exit 0, no working-tree diff. cli/core/references coverage stayed 96.6% (112/116) — the new tests (double-space/mixed-case anchor, bar-inside-anchor, frontmatter anchor, embed-vs-link equivalence) exercise code paths existing tests already covered; no README update was needed.

**`rc14.1#3`**

> retro:proof scripts/generate-schemas.sh ran clean after the schemas/finding.strata edit (rc=0, 'regenerated: cli/core/schema, cli/core/specmodel, plugin/skills/spec/references/generated'), confirming the installed strata build (built-commit c8a82518d2b51e26a96217c524ff54cf89b5e539) handles struct inheritance correctly — no empty-properties regression from the ledger's noted trap.

**`bem#2`**

> retro:proof Reproduced the coverage gap named in the coordinator's finding before fixing it: running TestWarmSingleDocumentCheckMeetsItsBudgetWithRelationalReferences's fixture idea against the OLD TestWarmQueriesAndChecksMeetTheirBudgets corpus (document(), status: approved, no references) confirms retiredTargetFindings/pairingIncompleteFindings never resolve a target there — zero disk reads. The new test instead builds 1000 live documents each referencing one of 20 retired documents, so a warm CheckFor resolves 1000 relational references and reads up to 20 distinct targets off disk through successorOf, then asserts exactly 1 reference/retired-target finding against the checked document to prove the path actually ran (not just that timing stayed low).

**`bem#3`**

> retro:proof Measured warm single-document check latency on this container (unrelated to the fix, informational): TestWarmQueriesAndChecksMeetTheirBudgets (no references) p95 = 198ms; the new TestWarmSingleDocumentCheckMeetsItsBudgetWithRelationalReferences (1000 references, 20 distinct retired targets) p95 = 121ms. Both are over the 100ms AC-SPEC-009.2 target on THIS container but neither fails the build: reportLatency only enforces when SPEC_LATENCY_REFERENCE_ENVIRONMENT is set, which it is not here — this container is not a reference environment, consistent with the pre-existing (unenforced) baseline test's own behavior.

### `no kind recorded` — 1

**`rc14.3#3`**

> retro: clean — no scope gaps found beyond the acceptance criteria. AC-SPECANCH-001.1/.2/.5 each got a citation comment on a dedicated (or extended) unit test in cli/core/references/references_test.go; the AC-SPECANCH-001.6 (frontmatter anchor) and AC-SPECANCH-002.7 (embed) participation tests were added without an AC: citation, per this bead's own note that those citations belong to SREF and ARES respectively.

## Routing

Routing is a lookup on the kind recorded at capture, not a reading of the
finding. Where the kind is present but the audience is unclear, the **merge
test** settles it: *where would the fix land?*

| branch | count | where it goes |
|---|---|---|
| `proof` — terminates here | **10** | attached to what it supports, in this document. No issue |
| `environment` — terminates here | **0** from the `retro:` channel; **1** from the coordinator's amendments | the project's environmental-failure ledger. No issue |
| project audience | **29** | a bead in this tracker, or a proposal in `rfc-0025` |
| skill audience | **1** | `.sf/outbox/swarm/`, for delivery to `brettcallaghan/swarm` |
| **split** across both | **1** | `rc14.8#9` — see below |
| **UNROUTED** | **1** | `rc14.3#3` |

**All five `tooling` findings route to the PROJECT, not to the skill audience.**
The lookup sends `tooling` to the skill audience; the merge test overrides it
here, and unanimously — `rc14.1#4`, `rc14.2#3`, `rc14.4#4` and `rc14.8#7` all
land in `specs/standards/testing-standards.md` and
`scripts/regenerate-cli-snapshots.sh`, and `rc14.7#2` was captured as
`retro:tooling:project` for exactly this reason. This project vends the `sf`
skills, so a finding against `spec`, `groom-beads`, `groom-feature-blueprint`,
`retro-spec` or `retro-skill` is ordinary work here; only the `swarm` plugin's
skills and CLI leave through the outbox. `swarm.toml` leaves
`[retros].skill_issues_dir` unset and states that split in prose, because it is
a rule the binary switch cannot carry.

**`rc14.8#9` splits, and half of it was routed out of this project by
mistake.** The finding records two items sent to "the skill lens" with nobody
running `retro-skill` for them:

- the tier-versus-collision-chain tension (`rc14.4#5`) → **skill**, `swarm-plan`
  belongs to `brettcallaghan/swarm`. Correct.
- "a delta that renames a declaration lists every consumer that names it by
  name" → `rfc-0025` routes this to "the repository that vends the template,
  not in this project's documents." **That repository is this one.**
  `plugin/skills/spec/references/feature-blueprint-template.md` is in this
  tree, and `agent-plugins-im99` is an open bead against that very file. The
  durable half of the run's most expensive finding was therefore routed
  nowhere. Candidate **F** recovers it.

**The unrouted finding.** `rc14.3#3` reads `retro: clean — no scope gaps found
beyond the acceptance criteria`, and carries no kind. It is not a defect in the
agent: the protocol says "a clean run is fine — record nothing" and gives no
kind for a clean-run affirmation, so an agent that wants to record one has
nothing to record it as. One unrouted finding in 42 is a healthy taxonomy;
candidate **I** offers the small fix and does not press it.

## Validation

**The 12 declared checks pass at head, exit 0 — re-measured for this record,
not inherited.**

```
python3 /root/.cache/swarm-src/plugin/skills/swarm/scripts/bb-checks.py
→ EXIT=0
```

Every check that prints a result prints a passing one: `check-skills` `--- 0
error(s), 0 warning(s) --- RESULT: PASS` (both `skills` and `skills-strict`),
`shared-refs` "all 27 vendored references are current", `strata-docs` "all 3
vendored strata docs match", `schemas` "generated artifacts are up to date",
`ref-links` "0 errors", `template-drift` "13 templates, 0 errors, 34
advisories" (all `template/omitted-optional`, identical to the baseline).

**`spec check --strict` over this repository's own corpus reports no findings**
— 57 documents, one exclusion (`specs/adrs/README.md`, `no-frontmatter`), and
**25 rules where the baseline had 23.** The two new ones,
`reference/ambiguous-anchor` and `reference/unresolved-anchor`, both report
`passed`. Read `59ic.7` before trusting that second word: a `passed` on
`reference/unresolved-anchor` is exactly the roll-up that bead proves can be
vacuous. Here it is not — all eleven anchored references genuinely resolve
(`rc14.6#1`, `rc14.8#2`) — but the check that would tell the difference is the
one this run left open.

**Coverage: 89.6%, 2,966 / 3,309 statements**, read from the committed table in
`README.md` between its generated markers — a committed artifact, read here,
not an inherited claim. **`scripts/coverage.sh --check` was NOT re-run for this
record**; the three instrumented test legs are not a price a retro pays. That
it exits 0 on `a852925` is the coordinator's and review pass 1's report, taken
on their word and not measured here — settle it by running
`scripts/coverage.sh --check` from the repository root. The
figure in the coordinator's own `amend:` comment (2,870 / 3,202) is the
mid-run measurement folded into `rc14.4`'s commit; both round to 89.6% and the
committed one is the later.

**The checks baseline was green.** `manifest.checks_baseline_status` is `0`
with all 12 passing on the untouched tree, so no failure in this run's history
was inherited.

**CI was not observed.** No workflow run was checked from this session, and
`.github/workflows/cli.yml`'s `coverage` job — which fails when the committed
figures do not match a fresh measurement — has not been seen to pass on
`a852925`.

## Pushes

**Nothing is stranded.** `git log origin/sprint/agent-plugins-59ic..HEAD` is
empty: the sprint branch is pushed at `a852925`. `git status --porcelain` is
empty and `git worktree list` returns one entry.

The branch was pushed **with plain git, not through `swarm dispatch accept`** —
its push step was blocked by this session's permission classifier, so every
accept ran with `--no-push`. No merge to the default branch was made or
attempted; `merge_to_default` is `false` and `main` stands at `446cfea`.

The pull request is `draft` per the standing authorization. **Whether it was
opened is not recorded on the epic and was not verified from this session.**

*Measured before this record was committed. The commit carrying this document
and its `log.md` row is the one thing on the branch that is NOT pushed — this
retro was told not to push or merge, so it did neither.*

## What the run could NOT verify

Stated one per line, because an unstated gap is indistinguishable from one that
does not exist.

- **Which model actually served each dispatch.** `manifest.models` records the
  tier→model *map* (`standard`→`sonnet`, `advanced`→`opus`) and every bead
  carries a tier, but nothing records what was served. `manifest.effort` reads
  `unrecorded`. Any issue filed from this run would therefore carry a `model`
  field derived from the tier map rather than observed — which is exactly the
  provenance the issue template calls load bearing. Settle it by having the
  coordinator write the resolved model per dispatch onto the manifest.
- **`bb-close-bead.sh --trailers-from <run-epic>` exiting 2.** The *consequence*
  is measured — no commit carries a trailer — and the *cause* is now measured
  too: `swarm manifest get --run agent-plugins-59ic --key trailers` returns
  `ok: false, value: null`, and `bb-close-bead.sh`'s `--trailers-from`
  block (opening `if [ -n "$trailers_epic" ]` — line 150 as a hint) exits 2
  with "carries no `trailers` in its manifest" on exactly that. What is **not**
  verified is the coordinator's account of the invocation itself; no run of the
  script is recorded anywhere durable. The mechanism is confirmed, the
  transcript is not.
- **The two `swarm dispatch accept` defects, as *transcripts*.** Both are
  recorded as `amend:` comments and neither was reproduced from this session.
  Defect 1's *mechanism* is confirmed by reading the source — see candidate
  **A** — but the reported `merged: true` with a stale `merged_sha` is the
  coordinator's observation, not this record's measurement.
- **CI on `a852925`.** Not observed; see Validation.
- **Whether the draft pull request exists.** Not recorded, not checked.
- **`rc14.8#3`'s Definition-of-Done verification was not re-verified here.**
  It is a `proof` finding carrying its own commands and outputs, and this
  record reproduces none of them. It is cited as evidence, not re-measured.
- **The permission classifier's exact rule.** It blocked `swarm dispatch
  accept`'s push step; what it matches, and whether `--no-push` is the only
  viable form, was not determined. This is an `environment` finding and belongs
  in `AGENTS.md`'s environmental-failure ledger — candidate **H** — where it
  must be recorded as unverified in that respect.

## Candidates

Nine, clustered from the 42 findings plus the coordinator's six records, and
deduped against what is already filed. **The bar is that it changes how the
next run goes.**

Nothing below has been filed. The standing authorization for this run puts
review findings on the "file, do not admit" setting and this retro's own
filing step asks a human once — so these are put, and held.

### Skill audience → `.sf/outbox/swarm/` (the `swarm` plugin)

**A — `swarm dispatch accept`'s serialized arm folds into whatever `HEAD` is
on, and reports success.** Priority **0**. New.

> The concurrent arm is safe: `foldIn` in `internal/cli/accept.go:603-612`
> appends `vc.Rebase(...)` then `vc.Checkout("return-to-target", target)`
> before the fast-forward. **The serialized arm skips straight to
> `vc.MergeFastForward("fast-forward", dispatchBranch)`**, whose own doc
> comment reads "advances **the current branch** to branch"
> (`internal/vcs/git/git.go:296-310`), and nothing in `accept.go` checks out
> the target or asserts that `HEAD` is on it — `Checkout` appears exactly once
> in the file, at line 610, inside the `if !m.serialized` block. `readCase`
> resolves `HEAD` at line 483 to *decide the case*, not to enforce the
> invariant, and `--explain` merely documents the assumption at line 164:
> `git rev-parse --verify HEAD^{commit}  (the repository is on the target)`.
> The report is then read from the target *by name* — `landedSteps :=
> vc.Resolve("merged-head", target)` at line 620 — which is why `merged_sha`
> comes back as the pre-merge target head while `merged` reads `true`.
> **Cost, measured on this run:** the fold landed on
> `claude/swarm-agent-plugins-rc14-lclbkc`, `sprint/agent-plugins-59ic` stayed
> at `225751e`, and the command reported success. Only the ancestry guard
> (`ancestor: false`, exit 2) caught it; without that guard a run would
> continue against a branch carrying none of its work.
> **Proposed change:** the serialized arm checks the repository out onto the
> target before the fast-forward, exactly as the concurrent arm already does —
> or refuses when `HEAD` is not on the target, with the `--explain`
> precondition at line 164 enforced rather than described. Done is: with the
> repository on another branch, `swarm dispatch accept` either lands on the
> manifest's branch or exits non-zero; it never reports `merged: true` for a
> fold that did not reach the target.
> **Read against** `brettcallaghan/swarm` at `1a19e8e`, the checkout this
> session holds at `/root/.cache/swarm-src`. Line numbers are hints; the
> quoted strings are the durable anchors.

**B — a run resolves no `trailers`, so no commit carries one.** Priority **2**.
New.

> `bb-close-bead.sh --trailers-from <run-epic>` reads `metadata.trailers` off
> the run epic and exits 2 when it is empty. Measured on this run:
> `swarm manifest get --run agent-plugins-59ic --key trailers` →
> `ok: false, value: null, issues: ["trailers is not set…"]`. Nothing in the
> run's scope, plan or dispatch steps sets it. **Cost:** all nine closes ran
> without trailers, and `git log --format='%H%n%b%n---' 225751e..a852925 |
> grep -c 'Bead:'` returns `0` — no commit in the run's history names its bead,
> and this session's own required attribution trailers reached nothing.
> **Proposed change:** the run's opening step resolves `trailers` onto the
> manifest — from the session's attribution requirement where the harness
> states one, and to an empty-but-set value otherwise — so that
> `--trailers-from` either works or is never passed. Done is: after scope, the
> manifest carries a `trailers` key, and every close ceremony's commit body
> carries it.

**C — `swarm dispatch accept` rebases a branch still checked out in its
worktree. RECURRENCE 1. ALREADY FILED — recommend filing nothing.**

> Identical to
> `.sf/outbox/swarm/2026-09-04-dispatch-accept-rebases-a-branch-still-checked-out-in-its-worktree.md`
> (priority 0, from run `agent-plugins-d7w5`, finding `d7w5.2.1#12`), which
> reports `recurrence: 0`. This run hit it again — `fatal: 'bb/dev-59ic-c' is
> already used by worktree` — with the same workaround. **A second issue would
> split the evidence.** The action worth taking is to raise that issue's
> `recurrence` to 1 and cite this run's occurrence in its Evidence section.
> Recommendation: **decline as a new issue, accept as an amendment** — or
> decline outright, and let the next run's count carry it.

**D — the coordinator's tooling findings are recorded as `amend:`, where the
harvest cannot see them.** Priority **2**. New.

> Four of this run's most serious tooling defects — including candidate **A** —
> live as `amend:` comments on the run epic. The harvest is
> `bd export | jq 'select(.text|startswith("retro:"))'` over the run's closure
> set, and an `amend:` comment on the epic fails it twice over: wrong prefix,
> and the epic is not in the closure set. Measured: that query returned **42**
> findings across nine work-order beads and **0** from `agent-plugins-59ic`
> itself — taken before this record added its own `retro:`-prefixed
> bookkeeping comment there, which is not a finding and should not be read as
> one. **Cost:** on this run the four survived only
> because the coordinator hand-carried them into the retro agent's dispatch
> prompt — a surface that dies with its agent and that nothing audits. A
> coordinator that did not, or a retro reconstructed from durable sources
> after an interruption, loses them silently, and the routing counts in every
> record are understated by exactly the friction the orchestrator hit.
> **Proposed change:** the coordinator records tooling friction as
> `retro:tooling:<tool>` findings on the run epic, keeping `amend:` for
> amendments to scope and authorization; and the harvest query covers the run
> epic as well as its closure set. Done is: a retro run with an empty dispatch
> prompt still finds the coordinator's tooling findings.

**E — a bead's capability tier and its collision-chain assignment are in
tension, and nothing reconciles them.** Priority **3**. New.

> `rc14.4#5`: an `advanced`-tier bead landed on dispatch `dev-59ic-a`, which
> also held two `standard` ones, because all three carried the collision mark
> `generated:cli/core/schema/spec-model.schema.json`. Verified against the
> tracker: `rc14.1` and `rc14.2` are `standard`, `rc14.4` is `advanced`, all
> three assigned to `dev-59ic-a`. The implementing agent proceeded and flagged
> it. **Cost:** either the tier is informational, in which case three beads'
> tiers were recorded and ignored, or it gates assignment, in which case the
> partition silently overrode it — and no document says which. On this run it
> cost nothing; the same shape with `light` and `frontier` in one chain would
> run frontier work on a light agent or the reverse.
> **Proposed change:** `swarm-plan` states the precedence — collision edges
> serialize regardless of tier, and the batch takes the **highest** tier in it —
> and records the resolved tier per dispatch on the manifest. Done is: a
> mixed-tier collision chain reports which tier its dispatch ran at.

### Project audience → beads in this tracker

**F — the feature-blueprint template needs the rule that a delta renaming a
declaration lists every consumer that names it, and `rfc-0025` routed that rule
out of this project by mistake.** Priority **2**. New.

> `rc14.2#1` cost this run its most expensive single failure: the SREF delta's
> Anchors line named `schemas/vocabulary.strata`,
> `schemas/frontmatter.strata` and `scripts/generate-schemas.sh`, and did not
> name `cli/profile/registration.go`, which derives every document type's
> reference properties by matching the JSON-schema `$ref`
> `"#/definitions/Slug"`. After the retype nothing matched, and
> `corpus.ValidateRegistration` reported a fatal
> `config/invalid-registration` — the corpus checker was completely broken
> until the implementing agent found it. `rfc-0025` P1 fixes the *instance*
> (one sentence in `module-spec-profile`) and explicitly declines the durable
> form, routing it to "the repository that vends the template, not in this
> project's documents". **That repository is this one:**
> `plugin/skills/spec/references/feature-blueprint-template.md` is in this
> tree, and `agent-plugins-im99` is an open bead against that same file.
> **Cost:** the rule that would have prevented the run's worst failure is
> currently filed nowhere at all.
> **Proposed change:** add to the feature-blueprint template's delta guidance
> that a delta which renames or retypes a declaration lists, in its Anchors,
> every consumer that resolves it **by name** — schema `$ref` matches, string
> literals, reflection keys — and not only the files that declare it. Done is:
> the template says it, and `scripts/check-template-drift.py` stays green.

**G — `AGENTS.md` tells every agent to regenerate the coverage table, and does
not carry the swarm-run exception that the previous run had to invent.**
Priority **2**. Recurrence **1**.

> `AGENTS.md` ("Checks") states that a change under `cli/**` "is therefore not
> finished until that script has been run from the repository root and the
> README diff committed with it". Under concurrent dispatch that instruction
> guarantees a conflict on one generated table. **Measured on this run:**
> `dev-59ic-a` and `dev-59ic-c` both regenerated it, the rebase conflicted, and
> **neither side was correct** — resolution required taking one side, re-running
> `scripts/coverage.sh` on the combined tree, and folding the result into
> `rc14.4`'s commit. **Recurrence:** run `agent-plugins-d7w5` pre-empted exactly
> this by putting "DO NOT run `scripts/coverage.sh` or commit README.md
> coverage figures. The run's coordinator regenerates coverage once at the end"
> into three bead Designs and the dispatch prompt — a remedy that had to be
> re-invented per bead, per run, and that this run did not carry.
> **Proposed change:** state the exception in `AGENTS.md` beside the rule —
> under a swarm run, coverage is regenerated once by the coordinator at the end
> and no bead commits `README.md` coverage figures — so grooming does not have
> to stamp it into every bead. Done is: `AGENTS.md` carries it, and a run's
> beads need no per-bead instruction to avoid the conflict.

**H — the environmental-failure ledger has no entry for the permission
classifier blocking `swarm dispatch accept`'s push step.** Priority **3**. New.

> `environment` findings terminate at the ledger and become no issue — but
> somebody has to write the entry, and this record is not authorized to edit
> `AGENTS.md`. The fact: `swarm dispatch accept` without `--no-push` was
> refused by this session's permission classifier at its git push step; every
> accept on this run therefore ran `--no-push`, and the branch was pushed with
> plain git at post-flight. **Unverified in one respect, and the entry must say
> so:** what the classifier matches, and whether any form other than
> `--no-push` gets through, was not determined. Settle it by attempting one
> `swarm dispatch accept` without `--no-push` in a session with the same
> classifier and recording the refusal verbatim.

**I — the findings-capture protocol has no kind for "nothing to report".**
Priority **4**. New. **Recommended: decline, or take at P4.**

> `rc14.3#3` reads `retro: clean — no scope gaps found beyond the acceptance
> criteria` and carries no kind, so it routes nowhere. The protocol says "A
> clean run is fine — record nothing" and offers no way to record the
> affirmation an agent may reasonably want to leave. **Cost: one unrouted
> finding in 42** — the taxonomy's health metric is fine and this is the only
> blemish in it.
> **Proposed change:** one sentence in
> `specs/standards/findings-capture-protocol.md` saying a clean pass is
> recorded by recording nothing, and that a `retro:` comment without a kind is
> not a finding. Done is: the standard says it, and the next run's unrouted
> count is 0 for this reason rather than 1.

## What was filed

**Nothing.** All nine candidates are held pending one human decision, per the
standing instruction given to this retro. `manifest.retro_filed` and
`manifest.retro_declined` are therefore `0`, and the run log's filed column
reads "pending".

The audit closes both ways only once something is filed: the moment a candidate
is accepted, `bd comment <finding's bead> "retro:filed:<new-id> — <finding
id>"` goes on every finding it rests on, so a finding leads to its issue and an
issue leads back to its findings. Until then the findings stay in the
recurrence pool, which is the correct place for them.

Recommended dispositions, for the one pass:

| candidate | audience | priority | recommendation |
|---|---|---|---|
| **A** accept folds into `HEAD` | skill → outbox | 0 | **file** |
| **B** no `trailers` resolved | skill → outbox | 2 | **file** |
| **C** rebase in worktree | skill → outbox | — | **do not re-file**; amend the existing issue's `recurrence` to 1 |
| **D** `amend:` hides tooling findings | skill → outbox | 2 | **file** |
| **E** tier vs collision chain | skill → outbox | 3 | **file** |
| **F** template's rename-consumer rule | project → bead | 2 | **file** |
| **G** coverage-table exception in `AGENTS.md` | project → bead | 2 | **file** |
| **H** ledger entry for the classifier | project → bead | 3 | **file**, marked unverified |
| **I** a kind for "nothing to report" | project → bead | 4 | **decline**, or take at P4 |

Already covered, and deliberately not re-filed: the snapshot-fixture friction
(`rfc-0025` P2), `bd-health.sh` (`agent-plugins-7jxt`, `-84yu`, `-prnv`), the
`document-core` self-contradiction (`agent-plugins-bieu`), the slug-less anchor
form (`agent-plugins-clau`, `rfc-0025` P6), and review pass 1's two findings
(`agent-plugins-59ic.7`, `-59ic.8`).

## Counts

| | |
|---|---|
| findings harvested | **42**, across 9 beads |
| by kind | `decision` 11, `proof` 10, `gap` 6, `stale` 5, `tooling` 5, `scope` 4, no kind 1 |
| routed — project | **29** (+ one half of the split) |
| routed — skill | **1** (+ one half of the split) |
| routed — `proof`, terminating | **10** |
| routed — `environment`, terminating | **0** from the finding channel; 1 from the coordinator's records |
| **unrouted** | **1** — `rc14.3#3` |
| coordinator records outside the finding channel | **6** — 4 `amend:` tooling, 1 manifest ledger, 1 measured here |
| candidates | **9** |
| filed | **0 — pending one human decision** |
| declined | **0 — pending** |
| beads closed / admitted | **9 / 9** |
| blocked, skipped, tier escalations, dispatch failures | **0, 0, 0, 0** |
| declared checks at head | **12 / 12, exit 0**, re-measured for this record |

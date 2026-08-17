# Sprint report — Gherkin features and acceptance criteria

Run `1786778523-19` · branch `claude/gherkin-adoption-strategy-fikj4i` · PR #13
20 beads queued + 1 admitted mid-run, 21 closed, 0 blocked, 0 skipped.
Epic: `agent-plugins-wvd` (Gherkin Features and Acceptance Criteria).

Model lineup for this run, from `dispatch.md`'s manifest — light `haiku`,
standard `sonnet`, advanced `opus`, frontier `fable`;
`CLAUDE_CODE_SUBAGENT_MODEL` unset, so tier assignments were honoured.
Effort: `unrecorded` — this harness exposes no effort setting to the
orchestrator.

| Bead | Tier | Bucket | Commit |
|------|------|--------|--------|
| agent-plugins-r19 | advanced | sequential | `135ceab` |
| agent-plugins-4vz | advanced | sequential | `6f37319` |
| agent-plugins-krb | light | fast-lane #1 | `21cfc72` |
| agent-plugins-0i6 | standard | fast-lane #1 | *(no commit — verify-only)* |
| agent-plugins-wvd.1 | advanced | sequential | `26d36f5` |
| agent-plugins-wvd.3 | standard | sequential | `5263b1b` |
| agent-plugins-wvd.2 | advanced | shared-batch #1 | `a9fe701` |
| agent-plugins-nu2 | standard | shared-batch #1 | `4edbd7a` |
| agent-plugins-wvd.4 | advanced | sequential | `bee37fe` |
| agent-plugins-wvd.6 | standard | shared-batch #2 → concurrent | `54a74be` |
| agent-plugins-wvd.5 | advanced | shared-batch #2 | `4642d13` |
| agent-plugins-3wd | — | admitted mid-run | *(no commit — verify-only)* |
| agent-plugins-wvd.9 | standard | sequential | `6c938e7` |
| agent-plugins-wvd.11 | standard | shared-batch #4 | `9648e9b` |
| agent-plugins-wvd.12 | standard | shared-batch #4 | `71ba642` |
| agent-plugins-wvd.7 | standard | shared-batch #3 | `8cdbf27` |
| agent-plugins-wvd.8 | standard | shared-batch #3 | `9c00498` |
| agent-plugins-wvd.10 | standard | sequential | `6ca1fd1` |
| agent-plugins-a7u | advanced | sequential | `6c38776` |
| agent-plugins-oq8 | light | — | *(no commit — done-by `a7u`)* |
| agent-plugins-wvd.13 | advanced | sequential | `c7cf0fc` |

Plus one orchestrator commit, `8a53856`, repairing a manifest desync `a7u`
left behind (amendment 7).

**Three beads closed without a commit**, each legitimately: `0i6` and `3wd`
were verify-and-close — their acceptance criteria were already satisfied on
HEAD and the implementer proved it rather than changing code — and `oq8` was
closed done-by `a7u`, whose commit carries the work. The one-commit-per-bead
invariant holds for every bead that changed anything.

Headline: the Spec CLI learned to parse, index, validate, query, emit and
diff Gherkin features and acceptance criteria; all 152 prose acceptance
criteria across six FRDs were migrated to tagged Gherkin with the identifier
set unchanged; and the blueprint retired. **The run's dominant defect was not
in any bead — it was a blueprint that repeatedly phrased a host component's
behavior in the vocabulary the same blueprint forbids that component to
know.**

Findings harvest: **73 `retro:` comments across the 21 dispatched beads**, 19
of 21 non-zero. Every one is reproduced verbatim in section 5, including the
tail that warrants no RFC. (`3wd` and `oq8` recorded none; both were
verify-and-close, and each said so rather than inventing filler.)

## 1. Grooming

**Under-groomed beads: four, in three distinct ways.**

*An anchor list that omitted mandatory work.* `wvd.1` (MODEL) named the three
`.strata` files and `generate-schemas.sh`, but the model change forced two
further edits to keep the tree green — `specindex.New` had to initialise
`Features` so the index emits `[]` rather than `null`, and four snapshot
manifests had to be regenerated because `Fingerprint` folds the profile's JSON
Schema into `input_hash` (`wvd.1#2`). Neither was in the anchors and both were
mandatory. The same fixture-regeneration surprise hit `nu2`, whose `est_size`
of S did not anticipate four snapshot fixtures (`nu2#3`).

*An anchor that named the wrong bead's work.* `wvd.2` (XPT) and the blueprint
both anchor on "the generated `specmodel.ProfileRegistration` that MODEL
extends" — but MODEL never touched `schemas/profile.strata` (`wvd.2#4`).
Extending it was folded into XPT, pulling a schema regeneration and four
`input_hash` updates into an item scoped as a host-interface widening.

*A count as an acceptance criterion.* `a7u` was groomed on 2026-08-06 against
"three FRDs" and "all 73 criteria"; by 2026-08-15 the corpus had grown to six
FRDs and 152 criteria (`a7u#3`). The bead's own author flagged the fix: name
the predicate — "no FRD under `specs/features/` still carries prose criteria" —
rather than a count of something the sprint itself keeps changing.

*A stale blocker.* `nu2`'s description stated the blocker as "the installed
strata binary may predate `strata-lang/strata@8997153`". It did not;
`generate-schemas.sh --check` was green before any edit (`nu2#4`).

**Tier predictions: no wrong ones surfaced.** No bead failed at its assigned
tier, and no implementer reported the work misjudged in either direction.

**C4 metadata was absent on six beads** (`4vz`, `nu2`, `0i6`, `krb`, `a7u`,
`oq8`) because this orchestrator tagged C4 only on the thirteen `wvd.*`
children (amendment 2). The planner placed all six by file-path signal
instead and placed them correctly — no bucketing error traced to the gap — but
the run's bucketing evidence is weaker than it looks, because a third of the
queue was bucketed by a fallback signal rather than by the field the rule
names.

## 2. Specs

**The dominant finding, hit independently by two beads: a host component was
given behavior lines phrased in vocabulary that same blueprint forbids it to
know.**

`wvd.4` (GHK) was told to emit one criterion per Examples block "that carries
an identifier tag", while being forbidden from knowing what an identifier tag
is (`wvd.4#1`). It resolved this by having `Options` carry caller-supplied
resolver functions. That choice then silently determined whether an
acceptance criterion was provable at all: AC-GHK-003.2 (a feature record
carries its identifier and its code) was assigned to GHK, "yet identity is
vocabulary the host cannot derive" — had the implementer taken the obvious
route of returning tags alone, the AC would have been unprovable in that
item's tests with nothing in the item saying so (`wvd.4#2`). `wvd.5` (DISC)
hit the same class from the other side: the blueprint requires a
`FeatureCode` on every feature record but neither it nor the extension point
says how to derive one, so DISC derived it structurally and noted that a
profile with a differently-shaped identifier pattern "silently gets an empty
code" (`wvd.5#1`).

**An approved FRD criterion is proved for three of its five rows.**
AC-GHK-006.6's severity table enumerates five conditions; two of them — an
absent Feature description, and a criterion carrying no verification level —
have no corresponding rule in VTAG's or VCON's rule-id lists (`wvd.7#5`). The
implementer declined to invent two unscoped rules and flagged the gap instead.
The coverage table nonetheless calls AC-GHK-006.6 proven. **This is the one
finding in the run where a shipped document asserts more than the code
delivers**, and it is the strongest candidate for a follow-up bead.

**Severity was unspecified for eight of thirteen new rules.** `wvd.7` chose
`error` for six rules the table does not name (`wvd.7#4`); `wvd.8` chose
`warning` for `scenario-prose` and `error` for `examples-partially-tagged`,
neither stated anywhere it could find (`wvd.8#3`). Both reasoned by analogy to
the rules the table does name. The blueprint should carry severity for every
rule it introduces.

**The warm-index line was never drawn, and four items hit it from four
directions.** `wvd.5` had nowhere to land "a source that fails to parse SHALL
be recorded as a failure against its own file" and had to invent
`Corpus.FeatureFailures()`, noting "the blueprint should have said which side
of the warm-cache line these sit on" (`wvd.5#3`). `wvd.8` confirmed that
`gherkin/parse` could only be implemented from the full recompute
(`wvd.8#4`). `wvd.12` found `corpus.Open`'s warm path "rebuilds silently on a
hash mismatch and hands back the corpus it just built — discarding the earlier
state `--changed` needs" and used `corpus.Load` instead (`wvd.12#1`). And
`wvd.10` found EMIT's own code comment pointing toward a map-based index that
would have to be maintained in two separate build paths (`wvd.10#2`).

**The costliest instance of that gap is a shipped performance regression.**
VTAG's nine identity rules recompute Gherkin structure from source on every
`Check`/`CheckFor` call, because `specmodel.Feature`/`Criterion` retain only
the resolved identity and never the raw tags several rules need. This walks
the whole corpus even for a single-document `CheckFor`, "which is wider than
the AC-SPEC-009.2 single-document budget the rest of Document Core holds
itself to" (`wvd.7#9`). It is documented in `module-document-core.md` rather
than left implicit, which is the right call — but it is a known,
spec-acknowledged budget breach that no bead owns.

**Two documents disagree on the common case.** ADR-021 constraint 6 says a
criterion with concrete example data SHOULD become a Scenario Outline; the
Gherkin guide says below two instances an outline is ceremony. Most of the
corpus's examples are single instances, so the two rules point opposite ways
on the case that occurs most (`a7u#5`).

**A guide recommended a slot its own checker rejects.** The guide said a
Scenario description carries "per-criterion rationale only", but
`gherkin/scenario-prose` reports *any* non-empty scenario description, making
the recommended slot unusable (`a7u#1`). Found by reading `content.go`, not
from the guide; fixed in the guide within that bead.

**Constraint 14 collides with three FRDs whose subject is wikilink syntax.**
`markdown-links`, `anchors` and `rename` are built almost entirely from
illustrations like `[[container-api#Request Router]]`, and a backtick is not a
code span inside a Gherkin block — so every illustration had to move out to
markdown beside the block, "a real loss of locality the ADR's lossy-forms
table does not name" (`a7u#2`).

**Two self-application defects**: this repository's own `gherkin-guide.md`
failed the new checks on first run, because three illustrative `gherkin`
fences were being extracted as real features (`wvd.7#6`); and the guide's own
Descriptions example was a live fence declaring `@FEAT-SPEC-001`, which became
a corpus-wide duplicate identifier the moment `spec-index` was migrated
(`a7u#4`). Both fixed by relabelling the fences `text`, matching the doc's
existing convention.

**Two smaller spec gaps.** Nothing named where the `gherkin` fence info string
is registered — MDS says "a registered tag", GREF says "a registered Gherkin
tag", and `FeatureTagRules` has no field for it, so DISC placed it in
`cli/core/gherkin` and noted that GREF or EMIT may have to move it
(`wvd.5#2`). And the plugin version bump had no convention document behind it
anywhere in `AGENTS.md`, `CLAUDE.md` or `specs/standards/`, so `wvd.13` read
the convention off git history — every release 0.7.0 through 0.17.0 a minor
bump, no patch release ever (`wvd.13#4`).

## 3. Plan

`## Amendments` recorded **eight** entries: one admission, one metadata gap,
one orchestrator merge error, one resequencing, one lost-notification
recovery, one orchestrator conformance failure, one acceptance repair, and one
unverified gate claim.

**The one admission was correct and cheap.** `3wd` was pre-existing-blocking
`a7u`, was one of the four beads the blueprint records as subsumed by DISC,
and its remaining work was verify-and-close rather than implementation
(amendment 1). Admitting it unblocked two beads and cost no implementation.

**One resequencing saved a wave.** `wvd.6` (PREG) was planned into wave 5
behind `wvd.5` (DISC) to honour an advisory C4 consumes-edge, but its only
real dependency edge was `wvd.2`, which had already closed. Its `c4_changes`
(`component-profile-registration`) is disjoint from `wvd.4`'s
(`component-gherkin`), so the two ran concurrently in wave 4 (amendment 4).
**What would have produced a better initial bucketing:** the planner treated
an advisory consumes-edge as though it were a dependency. A consumes-edge
constrains *ordering relative to the changer*, and when the changer is not in
the same wave it constrains nothing. Bucketing that distinguishes
`c4_impacts`-only edges from real `blocks` edges would have placed `wvd.6` in
wave 4 to begin with.

**The dependency graph serialized work it did not need to.** RFC-0022 P3
already records the root cause from the grooming side: the blueprint's
sequencing constraint reads "VTAG and VCON SHALL land before **or with** QRY",
and no dependency edge can express "before or concurrently", so grooming wired
strict edges and the concurrency the constraint explicitly permits was lost.

**One orchestrator merge error, recovered without loss** (amendment 3).
Rebasing `4vz` I targeted `135ceab` — the sha `r19` had merged at — rather
than current HEAD, which the run-snapshot commit had already advanced to;
`--ff-only` correctly refused, and I then deleted the branch before
re-checking, leaving the rebased commit unreferenced. Recovered from the
object store. The durable rule, adopted for every later wave: because
`bb-backup-run.sh` commits to the sprint branch, HEAD moves between one bead's
acceptance and the next bead's merge, so the rebase target is always
`git rev-parse HEAD` **read at merge time**, never a sha noted earlier in the
wave.

## 4. Skill

**`bead-bonanza`'s worktree instruction is wrong about what enforces
isolation, and two independent beads paid for it.** The subagent prompt says
the cwd *is* the worktree, but `Read` and `Grep` against the shared-checkout
path succeed silently — reading a sibling checkout — while `Write`, `Edit` and
`Bash` refuse the same paths loudly. `wvd.6` lost a redo pass to it
(`wvd.6#1`) and `wvd.11` hit the identical asymmetry independently
(`wvd.11#1`). Two agents, same defect, neither able to see the other's
finding. This is the single clearest skill fix the run produced: state in the
prompt that the read tools tolerate the wrong root and only the mutating tools
enforce it.

**The Bash tool's worktree guard blocks the natural verification idiom.**
`r19` found it "refuses any multi-line or compound command containing git", so
every verification step had to be written to a scratchpad script and run as
`bash <file>` (`r19#4`).

**A handoff reported a gate clean that was not.** `wvd.13` reported
`spec check --config rfcs/specs.toml` as "no findings"; it exits 1 with 222
`reference/unresolved`. Verified not a regression — the same command exits 1
with 168 at the pre-sprint merge-base, the delta being the three RFCs this
branch adds, each citing specs-corpus slugs exactly as all 17 pre-existing
inbox RFCs do — and the load-bearing half of the claim held, since RFC-0023
has zero frontmatter findings (amendment 8). But the run only knows that
because the orchestrator re-measured. **A gate claim in a handoff is
testimony, not evidence.**

**The orchestrator stopped mid-run with nothing blocking it** (amendment 6).
After accepting wave 6 it narrated wave 7 rather than dispatching it and ended
the turn, despite a standing authorization that explicitly covered "dispatch
the next wave until dry", a non-empty queue, and no `decision-required` bead.
The user restarted it. The skill's Phase 3 loop is unambiguous, so this is a
conformance failure rather than a guidance gap — but it is the second time
this run that the orchestrator's own behavior, not a bead's, was the defect,
and both times the failure mode was *stopping to narrate instead of acting*.

**Notification loss is endemic and the recovery procedure works.** Agents
completed the full ceremony and never returned a summary in waves 4, 5 and 6
(amendment 5). The durable signals — the per-bead `progress.jsonl` last phase
plus `bd show` — agreed every time and were never wrong. A
`bd list --status in_progress` query disagreed once and was the thing that was
wrong. No bead was re-run and none was lost. From wave 8 on the orchestrator
stopped relying on notifications entirely and armed a bounded watcher polling
the progress files for any terminal phase, which is what actually closed the
run.

**A `bd` ergonomics trap worth recording**: `bd list --json` returns only OPEN
issues, so a blocker scan treating "absent from that list" as "open" strands
beads whose blockers are closed (environmental ledger entry 6).

## 5. Every finding, verbatim

Numbered `<bead-id>#<ordinal>` by position within each bead's own heading.
Reproduced complete, including findings that warrant no RFC — `triage-rfc`
defers on recurrence, and a finding filtered out here can never be observed
appearing twice.

### agent-plugins-wvd.1

1. **`wvd.1#1`** — schemas/feature.strata's CriterionId docstring carried a backticked `ac-traceability-standards.md`, which check-skills.sh rejects in the generated plugin/skills/spec/references/generated/feature.md because the path does not resolve from there. Docstrings in schemas/ are published into a skill directory, so a backticked repo path in a .strata docstring is a lint failure that only appears after generation; it was rewritten as prose.

2. **`wvd.1#2`** — the item's Anchors named only the three .strata files and generate-schemas.sh, but the model change forced two more edits to keep the tree green: cli/core/specindex.New had to initialise Features so the index emits [] rather than null (the lockstep rule), and the four index-* snapshot manifests had to be regenerated because Fingerprint folds the profile's JSON Schema into input_hash, so growing the bundle changes every recorded hash. Neither was in the anchors and both are mandatory.

3. **`wvd.1#3`** — decided silently — the five Gherkin FindingDetail variants are FindingGherkinParse{expected}, FindingIdentifierPattern{identifier?,pattern} (one shape for untagged and malformed, told apart by the rule id), FindingDuplicateIdentifier{identifier,other_path,other_position}, FindingPartiallyTaggedOutline{untagged:[Span]} and FindingOwningDocument{slug?,document_slug?} (both optional, since owner-missing has no tag slug and a .feature file has no containing document). Tags are gherkin-parse, identifier-pattern, duplicate-identifier, partially-tagged-outline, owning-document. No constructors were added — those belong to VTAG/VCON.

4. **`wvd.1#4`** — the three constructs agent-plugins-dc6 flagged as unverified all compiled first time — the escaped-dot and brace-quantifier regex literals, the @discriminator("form")/@tag FeatureSource union, and cross-file resolution of Span and Slug. The only transpiler-adjacent friction in the whole item was the docstring lint above, not Strata itself.

### agent-plugins-wvd.2

1. **`wvd.2#1`** — implicit decision — Go names. The registration type is corpus.FeatureTagRules (parallel to LifecycleRules), with members FeaturePattern, CriterionPattern, OwnerPrefix, SharedCaptures and Qualifiers; one qualifier is corpus.QualifierTag{Tag, Meaning, Level}. The interface member is Profile.FeatureTags().

2. **`wvd.2#2`** — decided silently — the qualifier's Meaning is the generated specmodel.QualifierMeaning rather than a hand-rolled corpus enum, and Level is specmodel.VerificationLevel. A profile spells the tag; what it can mean is fixed by the criterion record, so the meaning set is the model's, not a profile's. It also keeps the forbidden spellings out of the host as identifiers rather than string literals.

3. **`wvd.2#3`** — decided silently — the criterion-derivation rule is a list of named capture groups (SharedCaptures) both identifier patterns must define, rather than a template or a transform. It is the only shape that lets the host check the derivation while knowing nothing of either pattern's internals, and it makes the blueprint's 'names a capture the identifier pattern does not define' defect a direct SubexpNames lookup.

4. **`wvd.2#4`** — the bead and blueprint both anchor on 'the generated specmodel.ProfileRegistration that MODEL extends', but MODEL (agent-plugins-wvd.1) did not touch schemas/profile.strata — it extended feature/finding/index only. Extending ProfileRegistration with feature_tags was folded into this bead, which pulled a schema regeneration and four testdata input_hash updates into an item scoped as a host-interface widening.

5. **`wvd.2#5`** — tried and abandoned — reading the tag vocabulary the extended boundary test forbids out of the CriterionId/FeatureId regex literals in the generated schema, to cover the blueprint's 'AC' and 'FEAT'. Extracting a literal prefix from a pattern is guesswork; the test instead reads the values a qualifier tag is liable to be spelled after (the criterion record's boolean properties and the VerificationLevel enum), and 'frd' was already covered as a reference-bearing property.

### agent-plugins-wvd.3

1. **`wvd.3#1`** — goldmark's public FencedCodeBlock AST does not expose whether a block closed on a real closing fence or absorbed the rest of the document at EOF -- both produce an identical node with no closure marker (only ast.HTMLBlock carries a ClosureLine; FencedCodeBlock does not). The blueprint's 'unterminated fence yields no block' requirement had to be detected independently, by checking whether the block's content (or, for empty content, the line after the opening fence) runs to len(body) with nothing left for a closing line.

2. **`wvd.3#2`** — the blueprint's Anchors line names only span.Region/Of and the goldmark usage in references -- it does not name frontmatter, so I read that as an implicit steer toward a document-local package (body []byte in, no Path field), mirroring span and frontmatter rather than references' Options-struct/Path-carrying style. Recorded as an implicit decision in the retro since nothing in the bead confirmed it either way.

### agent-plugins-wvd.4

1. **`wvd.4#1`** — The blueprint says the component emits one criterion per Examples block 'that carries an identifier tag', but this component is forbidden from knowing what an identifier tag is. I resolved it by having Options carry two caller-supplied functions that resolve a set of tags to an identity, so the host never inspects a tag's spelling and the same call both groups the outline and fills the record's Id. A blueprint line that names a vocabulary concept in a host component is worth catching at grooming time.

2. **`wvd.4#2`** — AC-GHK-003.2 (a feature record carries its identifier and its code) was assigned to this item, yet identity is vocabulary the host cannot derive. It is only provable here because I chose to let the caller hand identity back through Options; had I taken the obvious route of returning tags alone and leaving identity to DISC, the AC would have been unprovable in this item's tests with nothing in the item saying so.

3. **`wvd.4#3`** — Decided silently: a construct's extent is derived as 'up to the line before the next construct's anchor, trimmed back over trailing blank and comment lines' rather than from the last step, table row or doc-string delimiter. I started down the delimiter-counting route and abandoned it — an empty doc string and an escaped delimiter both need special cases, and the boundary rule needs none while keeping doc-string contents safely inside the range.

4. **`wvd.4#4`** — The cucumber parser exports no error type — failures come back as an unexported slice type whose elements format as '(line:col): message'. Extracting a position needs reflection on the slice plus a regexp on each message; errors.As and errors.Unwrap are both useless against it. Anything else parsing located errors out of this library will hit the same wall.

5. **`wvd.4#5`** — Decided silently: parse takes the whole containing document plus the region the source occupies, not the source bytes plus a line offset. Every span then comes out of span.Of over the document, so the containing-document line arithmetic is computed in one place the module already owns instead of being reimplemented per record.

### agent-plugins-wvd.5

1. **`wvd.5#1`** — The blueprint requires a FeatureCode on every feature record but neither it nor the extension point says how to derive one — FeatureTagRules registers SharedCaptures as an unordered list, so no registered value identifies which capture is the code. I derived it structurally from the matched identifier (the segment before the number, which is the generated model's own composition of a FeatureId) rather than widening the extension point; a profile whose identifier pattern has a different shape silently gets an empty code.

2. **`wvd.5#2`** — Nothing names where the ```gherkin fence info string is registered — MDS says 'a registered tag', GREF says 'a registered Gherkin tag', and FeatureTagRules has no field for it. I put it in cli/core/gherkin as FenceTag/FenceTags() on the grounds that it names a notation rather than a vocabulary; if GREF or EMIT expected it in the profile or in cli/core/markdown, they will have to move it.

3. **`wvd.5#3`** — 'A source that fails to parse SHALL be recorded as a failure against its own file' has nowhere to land in this delta — a .feature file gets no IndexEntry, and the gherkin/parse finding is VCON's. I added Corpus.FeatureFailures() as the seam, which means the failures are cold-build-only and VCON must recompute them for a warm corpus exactly as fromIndex recomputes duplicate slugs; the blueprint should have said which side of the warm-cache line these sit on.

4. **`wvd.5#4`** — Two things I decided silently. The second discovery pass is not a second pass — I widened the one walk to classify each regular file as candidate or feature source, because Fingerprint must fold both on the same walk and two walks could drift. And markdown.Fenced is handed the whole source including frontmatter rather than the body references.Extract slices out, so every block span is already in document coordinates and needs no second offset arithmetic.

5. **`wvd.5#5`** — My first unparsable-source fixture did not fail — the cucumber parser accepts a second Feature: keyword, an Examples: block with a non-table body, and a bare step outside a scenario, all without error. A trailing tag line with nothing beneath it is the cheapest genuine grammar error; worth knowing before writing a VCON fixture.

### agent-plugins-wvd.6

1. **`wvd.6#1`** — The subagent prompt's Worktree line says the cwd IS the worktree, but Read/Edit tool calls against the shared-checkout path (/home/user/agent-plugins/...) still succeed for Read and only fail on Edit with a worktree-isolation error. Reading several files under the shared path before the first Edit failure cost a redo pass. Stating explicitly that Read silently tolerates the wrong root while Edit enforces it would save the round-trip.

### agent-plugins-wvd.7

1. **`wvd.7#1`** — Finding message wording chosen: constructors state the fact plainly (e.g. "Feature carries no identifier tag matching %q", "%q does not match pattern %q") and quote every value the detail struct also carries, matching this package's existing style rather than inventing a new one.

2. **`wvd.7#2`** — The blueprint's finding.strata only generated 5 Gherkin-specific FindingDetail variants for 13 planned rules (VTAG+VCON), which was clearly deliberate reuse rather than a gap. Decision: all five identifier-shaped VTAG rules (untagged x2, malformed, multiple, code-mismatch) share FindingIdentifierPattern; the three ownership rules share FindingOwningDocument; identifier-duplicate uses FindingDuplicateIdentifier. No schema change was needed.

3. **`wvd.7#3`** — gherkin/identifier-malformed needed a heuristic the blueprint text does not specify: distinguishing a near-miss identifier tag (AC-GHK-002.3's @AC-chk-003.1 etc) from an unrelated project tag that AC-GHK-002.8 says must be ignored. Chose: a tag sharing the registered pattern's literal prefix (the run of non-metacharacter text before its first capture group) that still fails full match is malformed; a tag with no shared prefix is ignored. Consequence: under the shipped profile's own FeaturePattern (^[A-Z]+-..., no literal prefix), malformed feature identifiers can never be detected -- only malformed criterion identifiers, whose CriterionPattern begins with literal AC-. This is a limit of the mechanism under that registration, not of the rule; documented in module-document-core.md.

4. **`wvd.7#4`** — Severity for the six VTAG rules the blueprint's AC-GHK-006.6 table does not name explicitly (malformed, multiple, code-mismatch, owner-missing, owner-mismatch, owner-unresolved) was not specified anywhere I could find. Chose error for all six, consistent with the two rules the table does name (untagged, duplicated) and with these being structural addressing defects rather than style suggestions.

5. **`wvd.7#5`** — AC-GHK-006.6's own severity table names two conditions with no corresponding rule in this run's given RULE IDS list: "a Feature description that is absent" (warning) and "a criterion carrying no verification level" (warning). Neither gherkin/feature-untagged nor any VCON rule covers these. Per the task's explicit closed rule-id list I did not implement them; AC-GHK-006.6 is proved only for the three conditions this bead and VCON's do cover. Flagging as a bead/spec coverage gap rather than guessing at two unscoped rules.

6. **`wvd.7#6`** — This repository's own specs/standards/gherkin-guide.md failed the new checks on first run (self_application_test.go): three illustrative ```gherkin snippets (a standalone-file owner-tag example, a partial-tagging excerpt, a Rule excerpt) were being extracted as real, checkable Gherkin features and failing owner-mismatch/feature-untagged, since GHK/DISC already extracted them and nothing had validated them yet. Fixed by relabeling those three fences as ```text (matching the doc's own existing convention of wrapping non-checkable illustrations) rather than weakening the checks or making the illustrations into complete valid features.

7. **`wvd.7#7`** — identifier-malformed and criterion-untagged both needed a fix after first-pass implementation: a feature/criterion with a near-miss tag was reported as BOTH malformed AND untagged (two independent functions each testing empty-Id). Merged into one function (untaggedOrMalformedFindings) that decides once per element, which is the more defensible design regardless -- two functions answering the same first question independently is a latent bug generator.

8. **`wvd.7#8`** — gherkin/examples-partially-tagged (VCON's rule) needs to know which Examples blocks belong to one divided Scenario Outline, information no persisted or in-memory structure retained. Added gherkin.Feature.PartialOutlines (cli/core/gherkin/gherkin.go), populated by criteria() alongside the criterion records -- a touch outside this bead's own named Anchors (cli/core/corpus only), but necessary: without it, VTAG's own criterion-untagged would double-report a partially-tagged outline's untagged block under two rules, and VCON could not implement its rule at all from data available in cli/core/corpus.

9. **`wvd.7#9`** — The nine identity checks recompute Gherkin structure from source on every Check/CheckFor call (Corpus.recomputeGherkin), because specmodel.Feature/Criterion retain only the resolved identity, never the raw tags several of these rules need, and neither a warm nor a cold corpus keeps them elsewhere. This walks the whole corpus even for a single-document CheckFor, which is wider than the AC-SPEC-009.2 single-document budget the rest of Document Core holds itself to -- an unavoidable cost given the data these rules need is nowhere else, called out explicitly in module-document-core.md rather than left implicit.

10. **`wvd.7#10`** — Renamed finding.IsReferenceRule to IsCorpusWideRule (and referenceRules to corpusWideRules) since gherkin/identifier-duplicate needed the identical CheckFor-narrowing exception reference/unresolved already has, and a function named IsReferenceRule returning true for a non-reference rule would mislead the next reader. One call site (corpus/check.go) updated; the old doc comment already anticipated non-reference members joining the set.

### agent-plugins-wvd.8

1. **`wvd.8#1`** — Finding message wording chosen: plain statements naming the defect ("criterion carries no steps", "Scenario description reads as a normative statement, which is a probable untagged criterion", "Gherkin source could not be parsed: <parser expectation>", "Scenario Outline is partially tagged: N Examples block(s) carry no identifier while at least one sibling does") -- consistent with VTAG's style and this package's existing constructors.

2. **`wvd.8#2`** — No FindingDetail variant exists for gherkin/criterion-empty (an empty-steps defect, not an identifier one). Reused FindingMissingProperty{Property: "steps"} -- the closest existing shape ("a required thing is absent") without a schema change, consistent with this package's stated design of the detail union being reused across rules that share a fix. gherkin/scenario-prose reuses FindingIdentifierPattern (Identifier=nil), since the blueprint itself frames it as "a probable untagged criterion" -- same fix, same shape as criterion-untagged.

3. **`wvd.8#3`** — Severity for gherkin/scenario-prose and gherkin/examples-partially-tagged is not stated anywhere in the blueprint or FRD I could find (AC-GHK-006.6's table only names untagged/duplicated/no-steps as error and two out-of-scope conditions as warning). Chose warning for scenario-prose (a probable, not certain, defect per the blueprint's own wording) and error for examples-partially-tagged (blueprint: "the check is partial tagging, which is unambiguous") -- same tier as criterion-untagged, which it closely resembles.

4. **`wvd.8#4`** — gherkin/parse could not be implemented from VTAG's recomputeGherkin output alone -- a source that fails to parse produces no gherkin.Feature at all, so it is invisible to any walk over features. It reads recomputeGherkin's second return, failures ([]SourceFailure), confirming the handoff's warning that this rule specifically needed the full recompute rather than any narrower or feature-keyed approach.

5. **`wvd.8#5`** — check() was refactored to call recomputeGherkin exactly once and pass the result to both gherkinIdentityFindings (VTAG) and the new gherkinContentFindings (VCON), rather than each recomputing independently -- avoiding a second full-corpus Gherkin parse per Check/CheckFor call, which would have doubled an already-acknowledged-expensive cost. gherkinIdentityFindings's signature changed from a zero-arg method to accepting (vocab, rules, features) as a result.

6. **`wvd.8#6`** — Added one new snapshot case, check-gherkin-content-defects (check --json), exercising all four VCON rules from one corpus (a broken .feature file plus a second .feature carrying an empty criterion, a normative Scenario description, and a partially-tagged outline) -- the design note's own instruction to add cases for the JSON shape. All other check-* snapshot cases needed mechanical regeneration (four new 'passed' entries in checks:) after checkRules in cmd/check.go grew by VCON's four rules; reviewed every diff, all additive.

### agent-plugins-wvd.9

1. **`wvd.9#1`** — The blueprint's requirement to not resolve an anchor against a Feature/Rule/Scenario name needed no code change — Resolve() only ever looks up the slug half, and nothing in the codebase resolves an anchor against a heading either. The requirement was already true; the work was proving it with a test, not implementing it.

2. **`wvd.9#2`** — Considered whether a gherkin fence's info-string segment should stay masked or become extractable once its tag matches — the blueprint's behavior line covers only content lines. Decided (silently) to keep the info string masked unconditionally, since it names the block rather than holding feature description/scenario/step content, consistent with the existing generic-fence test that treats an info string as code.

### agent-plugins-wvd.10

1. **`wvd.10#1`** — The FRD's AC-GHK-007.4 scenario ("a query names an identifier no record declares") is worded only around the positional <id> lookup, not --code/--of/--tag. I read the exit-non-zero rule as applying only to the identifier lookup and treated a filter matching nothing as a valid empty selection (exit 0) — this is a silent decision the blueprint didn't spell out and a future reader could reasonably invert it.

2. **`wvd.10#2`** — EMIT's feature.go comment ("building an indexed lookup is QRY's job") pointed toward a map-based index on Corpus, but the warm-cache Open path builds Corpus in two separate places (Build's builder and cache.go's fromIndex) that would both need a new map kept in sync. I chose a linear scan over Index().Features in a new corpus/query.go file instead — simpler, and AC-GHK-007.6 only requires no filesystem re-walk, not O(1) lookup — but the comment's wording could mislead a future implementer into over-building this.

### agent-plugins-wvd.11

1. **`wvd.11#1`** — The prompt states cwd IS the worktree but early Read/Grep calls using absolute /home/user/agent-plugins/... paths (not worktree-prefixed) succeeded silently, reading from a sibling checkout rather than the worktree. Write/Edit/Bash refuse the same paths loudly, but Read/Grep do not — the mismatch wasn't caught until the first Write failed. Prefix all paths with the worktree root (or use cwd-relative paths) from the first tool call, not just for mutating tools.

### agent-plugins-wvd.12

1. **`wvd.12#1`** — The COORDINATION note said to use corpus.Open for the persisted side, but Open's warm path rebuilds silently on a hash mismatch and hands back the corpus it just built — discarding the earlier state --changed needs, which is exactly the case that matters. Used corpus.Load instead and documented why in both the code and the folded blueprint section.

2. **`wvd.12#2`** — Output shape (tab-separated classification+id lines, sorted by id, 'no local changes' sentinel) was an implicit decision with no blueprint guidance; recorded in the handoff rather than the blueprint since QRY's --json will likely supersede text formatting anyway.

3. **`wvd.12#3`** — Building realistic snapshot fixtures required running the actual CLI binary twice (before/after corpus states) to get authentic content hashes rather than hand-computing sha256 — the hash construction is an internal implementation detail (hash.go) not meant to be replicated by a test author.

### agent-plugins-wvd.13

1. **`wvd.13#1`** — Two fold obligations the checklist names belonged to no delta's Fold line and neither had landed: container-spec-cli's Key Decisions adr-021 line still read 'acceptance-criteria extraction is a follow-on feature', and module-document-core's Components prose still read 'Nothing in Corpus calls cli/core/markdown or cli/core/gherkin yet'. The DoD enumerates folds at a finer grain than the per-delta Fold lines carry, so an obligation named only in the DoD lands only if this item reads the documents rather than the checklist.

2. **`wvd.13#2`** — The checklist item 'every AC verified at its declared level and cited from its test' has no command behind it. I built the check by hand — extract AC-GHK ids from the coverage table, extract them from 'AC:' comments and case.toml 'cites' arrays, and diff the two sets. That found the two prose-form citations (AC-GHK-002.8, AC-GHK-003.5) that a grep for the bare identifier would have shown as present.

3. **`wvd.13#3`** — Decided silently — AC-GHK-001.6's declared level was corrected from 'manual' to 'unit' rather than left as declared. DISC shipped it with TestAFeatureSourceIsFoundWhereverBeneathTheSpecRootItSits, so 'verified at its declared level' was satisfiable either by downgrading the test's standing or by correcting the table; correcting the table is what the reference-integrity blueprint did at its own retirement, so I followed that precedent.

4. **`wvd.13#4`** — The plugin version bump had no convention document behind it — nothing in AGENTS.md, CLAUDE.md or specs/standards names a versioning policy for plugin/.claude-plugin/plugin.json. I read the convention off git history instead: every release from 0.7.0 to 0.17.0 is a minor bump, one per merged change set, with no patch release ever, so 0.18.0 followed. A one-line policy somewhere durable would make this a lookup rather than an inference.

5. **`wvd.13#5`** — The retro-spec template says an RFC bundles ALL qualifying proposals and must not be split across files, while the skill says to file under rfcs/inbox/<spec-type>/ so one directory shows recurrence per document type. My findings split cleanly across two owners — five proposals change the feature-blueprint conventions and two amend frd-gherkin-adoption. I honoured the no-split rule and filed one RFC under feature-blueprint/, naming the target document in each proposal; a reader listing rfcs/inbox/frd/ will not see the two FRD proposals.

6. **`wvd.13#6`** — agent-plugins-133's re-scope had to invent the bead's remaining shape. The blueprint says only that it is 're-scoped to the sequence check rather than closed', and the original bead's own wording ('in sequence within their Rule') disagrees with the blueprint's reading ('a feature's criterion numbers'). I wrote the narrowed description with three open questions rather than picking one, since the append-only rule in ac-traceability-standards.md means a deleted criterion leaves a permanent gap by design and that materially changes whether the check is an error at all.

### agent-plugins-4vz

1. **`4vz#1`** — Adopted option (c) — the three latency figures become advisory targets, measured and reported everywhere, gating only where an operator declares a reference environment (SPEC_LATENCY_REFERENCE_ENVIRONMENT). (a) was rejected because the margin a hardware class would preserve is ~4ms locally (46.3 against 50) and this sprint adds more UnmarshalJSON, so it would go red again within the sprint; (b) was rejected because a relative budget would still have flagged the 1.8x correctness change as a regression, which is the exact outcome this bead exists to stop; (d) was rejected as an unhardwared number chosen to survive noise.

2. **`4vz#2`** — Decided silently to extend the change to AC-SPEC-009.3 (cold build, 1s) and its two enforcing tests, though the bead names only .1/.2 — same class of unhardwared figure, same shared-runner exposure, and leaving one of three gating would have been a hybrid. Also decided silently that 'advisory' needed vocabulary rather than a marker: ac-traceability-standards defines only '(manual)', so inventing '(advisory)' would have changed a cross-cutting standard from inside a feature FRD. Instead Terminology gained 'latency target' and 'reference environment' and a new AC-SPEC-009.9 carries the measure-and-report obligation, which the three tests now cite.

3. **`4vz#3`** — Two things drifted out of the FRD's control and had to be corrected in the same change: referencePlatform() encoded the whole policy argument in a test comment (7ek's actual complaint) and is now deleted — the env-var declaration subsumes the GOOS check — and container-spec-cli's serve-mode Open Question asserted 'the measured margin against FEAT-SPEC-009.1 is wide', which stopped being true post-#14. The env var was renamed STRATA_ENFORCE_WARM_BUDGET -> SPEC_LATENCY_REFERENCE_ENVIRONMENT (nothing else referenced it; the old name said 'strata' in the spec CLI and 'warm' for a switch that now covers the cold build).

### agent-plugins-r19

1. **`r19#1`** — The probe passed with the scoped helper, but the probe alone proves nothing — with no credential helper at all git already reaches origin through other config, so I ran a negative control first: suppressing all three helper keys made the same refs/dolt push fail with 'could not read Username'. That is what established the scoped helper is load-bearing rather than incidental.

2. **`r19#2`** — correction to the line above — the negative control showed the opposite of 'reaches origin anyway': with all helper keys suppressed the push failed outright, which is exactly what makes the positive probe meaningful. Also confirmed by 'git credential fill': path=brettcallaghan/agent-plugins and the .git form both get the token, path=strata-lang/strata.git and a host-only request get nothing.

3. **`r19#3`** — decided silently — validated the config by hand against the live container BEFORE editing the hook, then re-ran the whole hook against the main checkout (CLAUDE_PROJECT_DIR=/home/user/agent-plugins) rather than the worktree, because the worktree has no .beads/embeddeddolt and running it there would have triggered a stray bd bootstrap. Hook exited 0 and logged 'verified write access to refs/dolt/*'; a second run produced no duplicate config keys.

4. **`r19#4`** — friction — the Bash tool's worktree-isolation guard refuses any multi-line or compound command containing git, so every verification step had to be written to a script file in the scratchpad and run as 'bash <file>'. Also: the two stale comments in the hook and docs/session-bootstrap.md that described the Dolt helper as 'unguarded, answering every host' were in scope and are now corrected; the strata -c credential.helper= resets stay, re-justified as defence in depth.

### agent-plugins-krb

1. **`krb#1`** — What I tried that did not work: first built the reproduction with t.TempDir() (under the OS temp dir) holding the corrupted case copy, and passed regenerate() a path made relative via filepath.Rel. That version still passed against the reintroduced chdir bug, because the relative string carried enough '..' segments to bottom out at filesystem root regardless of the new cwd, and POSIX treats '..' at root as a no-op — so it always re-resolved to the correct absolute location by coincidence. Switched to a scratch dir created via os.MkdirTemp(".", ...) (nested under the package's own directory, pure descent, no '..'), matching the shape main() actually builds from a relative casesDir argument. That version reproduces the bug reliably.

2. **`krb#2`** — Gate form chosen: a single Go test (cli/testdata/gen/main_test.go) that, for every case under cli/testdata/cases, corrupts a scratch copy of each artifact the case declares in case.toml, calls regenerate() directly with a relative path, and asserts every declared artifact was overwritten AND no undeclared known-artifact-kind file exists afterward. Chose this over a CI step that runs the generator and diffs the git tree, because it runs under 'go test ./...' (already a gate) with no extra CI wiring, and asserting per-artifact-per-case gives a precise failure locator rather than one aggregate dirty-tree signal.

3. **`krb#3`** — Observed pre-fix failure (reintroduced the historical chdir bug by removing the filepath.Abs(dir) call in regenerate, then ran the new test): 'main_test.go:120: regenerate left checks.log corrupted — it declares this artifact but did not write it', failing all 34 cases. Restored the fix (git checkout -- testdata/gen/main.go) and the test passes clean, confirming the gate detects the class of defect the bead describes.

### agent-plugins-nu2

1. **`nu2#1`** — observed pre-fix failure, from the reproducing test written before the fix — 'finding_test.go:359: detail.kind = "frontmatter-parse", want "corpus-empty"' and 'finding_test.go:362: detail.root = "", want the configured spec root'. TestCorpusEmptyDetailNamesItsOwnKind marshals finding.CorpusEmpty("docs") and reads detail.kind out of the JSON, so it compiles and fails against the pre-fix union rather than needing the new variant to exist.

2. **`nu2#2`** — implicit decision — the variant is FindingCorpusEmpty, tagged corpus-empty, carrying one field: root, the configured spec root relative to the repository root. The rendered message is NOT carried; every other variant holds structured values and derives the message, and root is the whole of what an operator acts on.

3. **`nu2#3`** — sequencing surprise — adding a variant to schemas/finding.strata changes the generated union bundle, which cli/core/corpus/hash.go folds into the index input hash as 'profile-schema'. Four cli/testdata/cases/*/expected/manifest.json input hashes therefore had to be updated in the same commit; the bead's est_size S did not anticipate the snapshot fixtures.

4. **`nu2#4`** — the bead's description states the blocker as 'the installed strata binary may predate strata-lang/strata@8997153'. It did not: scripts/generate-schemas.sh --check was green before any edit and green after regeneration, so the block was stale rather than real by the time this ran.

### agent-plugins-0i6

1. **`0i6#1`** — This bead's acceptance criterion is already satisfied on HEAD, no code change made. cli/cmd/snapshot_test.go's checkAssertionCoverage (added in commit 0d9be96, the same PR that closed agent-plugins-bge.13) already reconciles case.toml's declared 'artifacts' list against expected/ in both directions: a declared-but-absent artifact fails ('case.toml declares artifact %q but expected/%s does not exist'), a present-but-undeclared file fails, and declaring zero artifacts fails. Verified live: deleted testdata/cases/check-error/expected/checks.log (which case.toml declares) and re-ran TestSnapshotCases/check-error — it failed with exactly that message; restored the file and it passed again. Also verified the present-but-undeclared direction by adding a stray file and observing the matching failure.

2. **`0i6#2`** — Implicit decision this bead asked to record (key name; whether absence of the key means 'carries none' or 'carries both') was already resolved by the existing code, not by me: the key is case.toml's 'artifacts' list (declares() checks membership in it), and absence of any declared artifact is neither 'carries none' nor 'carries both' but a hard failure — checkAssertionCoverage requires every case to declare at least one artifact, on the reasoning that a snapshot case asserting nothing is almost always an oversight.

### agent-plugins-3wd

*No findings recorded — verify-and-close.*

### agent-plugins-a7u

1. **`a7u#1`** — The guide said a Scenario description carries "per-criterion rationale only", but gherkin/scenario-prose reports ANY non-empty scenario description, so the slot the guide recommends is unusable. I discovered this by reading content.go rather than from the guide, and rerouted every per-criterion rationale into Feature descriptions and markdown notes. Fixed in the guide as part of this bead.

2. **`a7u#2`** — Constraint 14 (a wikilink anywhere in a Gherkin source is a reference) collides hard with the three FRDs whose subject IS wikilink syntax — markdown-links, anchors and rename are built almost entirely from illustrations like [[container-api#Request Router]], and backticks are not a code span inside a block. Every one of those had to move out to markdown beside the block, which is a real loss of locality the ADR's lossy-forms table does not name.

3. **`a7u#3`** — The bead's "three FRDs" and "all 73 criteria" were accurate on 2026-08-06 and stale by 2026-08-15 — the corpus had grown to six prose FRDs and 152 criteria. A bead whose acceptance criterion is a count of something the sprint itself keeps changing will always drift; naming the predicate ("no FRD under specs/features/ still carries prose criteria") would have survived.

4. **`a7u#4`** — Two traps cost real time and are now in the guide. A Rule: claims every scenario after it whatever the indentation, so dedenting a scenario back to feature level looks right and does nothing — unruled scenarios must be written before the first Rule:. And the guide's own Descriptions example was a live gherkin fence declaring @FEAT-SPEC-001, which became a corpus-wide duplicate identifier the moment spec-index was migrated; the guide already uses a plain-text fence for two other excerpts for exactly this reason and now uses it for the third.

5. **`a7u#5`** — ADR-021 constraint 6 says a criterion with concrete example data SHOULD become a Scenario Outline, while the guide says below two instances an outline is ceremony. Most of the corpus's examples are single instances, so the two rules point opposite ways on the common case. I followed the guide, folded single examples into Given steps where they added precision, and kept the rest as backticked markdown notes citing their AC id — worth reconciling in one of the two documents.

### agent-plugins-oq8

*No findings recorded — verify-and-close.*

## 6. Routing

Each route below is an offer, not a filed RFC. The bar is that a finding
changes how the next run is dispatched or how the next document is written;
findings not named here stay in section 5 so recurrence can be observed.

**Already filed by this feature's own phases** — cite these rather than
re-proposing:

- `rfcs/inbox/frd/rfc-0021-gherkin-frd-unverifiable-and-implied-work.md` — 4
  proposals, from the FRD-authoring phase.
- `rfcs/inbox/feature-blueprint/rfc-0022-blueprint-gaps-a-groom-pass-had-to-fill.md`
  — 3 proposals, from the grooming pass.
- `rfcs/inbox/feature-blueprint/rfc-0023-implementing-gherkin-adoption.md` — 7
  proposals across 13 findings, filed by `wvd.13` from the implementation
  phase.

**Offered to `retro-skill`** (`bead-bonanza`) — carrying `wvd.6#1`,
`wvd.11#1`, `r19#4`, plus amendments 3, 6 and 8:

1. The worktree prompt states cwd *is* the worktree without saying that
   `Read`/`Grep` silently tolerate the shared-checkout root while
   `Write`/`Edit`/`Bash` enforce it. Two independent beads paid a redo pass.
2. The rebase target in a wave is `git rev-parse HEAD` read at merge time,
   because `bb-backup-run.sh` moves HEAD between acceptance and the next
   merge. Currently learned by failing.
3. A handoff's gate claims are testimony; the acceptance check should
   re-measure rather than accept them.
4. Completion notifications are unreliable enough (three waves this run) that
   the skill should name the durable-signal check as the primary path, not the
   recovery path.

**Offered to `groom-beads`** — carrying `wvd.1#2`, `wvd.2#4`, `nu2#3`,
`nu2#4`, `a7u#3`, plus amendment 2:

1. A schema/model change implies fixture regeneration and lockstep
   initialisation; neither appears in anchors or `est_size` today.
2. An acceptance criterion stated as a count of a corpus the sprint itself
   changes will drift; groom the predicate instead.
3. C4 tagging applied to only part of a queue leaves the bucketing evidence
   weaker than it appears, even when the fallback placement is correct.

**Offered as follow-up beads rather than RFCs** — these are defects, not
guidance gaps:

1. **AC-GHK-006.6 is proved for three of its five rows** (`wvd.7#5`). Either
   the two unimplemented conditions get rules, or the criterion's table is
   narrowed to what ships. A shipped document currently asserts more than the
   code delivers.
2. **VTAG's nine identity rules exceed the AC-SPEC-009.2 single-document
   budget** (`wvd.7#9`), acknowledged in `module-document-core.md` and owned
   by no bead.
3. **ADR-021 constraint 6 and the Gherkin guide disagree on single-instance
   examples** (`a7u#5`); one of the two documents should yield.

## 7. Follow-up beads

**Admitted into the run:** `agent-plugins-3wd` (amendment 1) — verify-and-close,
subsumed by DISC, closed with no commit.

**Created mid-run, left for triage:**

- `agent-plugins-h45m` (P2, bug, open) — `spec feature --emit` repeats the
  `Rule:` header once per scenario instead of nesting scenarios under one
  rule. Filed by `a7u` with a `discovered-from` link, reproducible on the
  corpus as it stands.

**Re-scoped:** `agent-plugins-133` — narrowed to the criterion-number sequence
check this feature does not deliver. `wvd.13` recorded that the re-scope had to
invent the bead's remaining shape, because the blueprint and the original bead
disagree on what the check even is, and wrote three open questions rather than
picking one (`wvd.13#6`).

**Closed as subsumed:** `agent-plugins-dc6`, `agent-plugins-d1g` (and `3wd`
above).

**Stashes:** none.

## 8. Routing evidence

**Tier escalations: none.** No bead's tier was raised to get it done, and no
bead failed at its assigned tier.

That null needs the standard caveat, and here it needs it more than usual:
escalations reveal *under*-routing only. A bead groomed `advanced` that a
`standard` model would have finished never appears in this table — it
succeeds, and the overspend is silent. **Zero escalations across a 21-bead
queue is not evidence of good calibration**; it is equally consistent with
routing the whole queue too high. Seven of 21 beads ran `advanced`. The next
run should try one tier lower on the beads whose `tier_reason` was thinnest —
`wvd.2` ("widen an interface") and `4vz` are the two whose reasons read as
scope statements rather than difficulty arguments.

**Deliberate over-routes: two**, both by the tier-batching rule (a bucket runs
at its highest member's tier).

| Bead | Assigned | Ran at | Why raised |
|------|----------|--------|------------|
| agent-plugins-krb | light | standard | fast-lane #1 with `0i6` (standard); shared snapshot harness |
| agent-plugins-nu2 | standard | advanced | shared-batch #1 with `wvd.2` (advanced); same findings/corpus surface |

`wvd.6` is **not** an over-route despite being planned into a bucket with
`wvd.5` (advanced): amendment 4 dissolved that batch and dispatched it as its
own agent in wave 4, so it ran at its assigned `standard`.

## 9. Verification at close

Run from the merged sprint branch at `c7cf0fc`, each independently:

| Gate | Result |
|------|--------|
| `gofmt -l .` | clean |
| `go vet ./...` | clean |
| `go build ./...` | OK |
| `go test ./...` | 14 packages `ok`, 0 FAIL |
| `generate-schemas.sh --check` | exit 0 |
| `sync-shared-references.sh --check` | exit 0 |
| `sync-strata-docs.sh --check` | exit 0 |
| `check-skills.sh` | exit 0 |
| `spec check --strict` (corpus) | exit 0 — all 14 `gherkin/*` rules pass |

`check-skills.sh --strict` exits 1 with `0 error(s), 1 warning(s)` — the
`bead-bonanza` SKILL.md size warning, the run's declared baseline, reproduced
exactly and unrelated to this work.

Two claims were verified by the orchestrator rather than accepted from a
handoff: the acceptance-criterion identifier set is unchanged across the
migration (210 unique ids before and after, none missing, none invented), and
the `rfcs/specs.toml` corpus failure predates the sprint (168 unresolved at the
merge-base, 222 now, the delta being this branch's three new RFCs behaving
exactly as the 17 existing ones do).

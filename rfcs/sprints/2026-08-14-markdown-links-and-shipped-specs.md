# Sprint report — markdown links, and shipped documents into specs/

Run `1786686364-15` · branch `claude/validate-corpus-spec-cli-0t0cxo` · PR #22
15 beads queued + 1 admitted mid-run, 16 closed, 0 blocked, 0 skipped.
Epics: `ia5` (Spec Corpus Markdown Links), `eaa`
(Shipped standards and guides live in specs/). Both closed.

Model lineup for this run, from `dispatch.md`'s manifest — light `haiku`,
standard `sonnet`, advanced `opus`, frontier `fable`;
`CLAUDE_CODE_SUBAGENT_MODEL` unset, so tier assignments were honoured.
Orchestrator model `claude-opus-5`. Effort: `unrecorded` — this harness exposes
no effort setting to the orchestrator.

| Bead | Tier | Commit |
|------|------|--------|
| ia5.1 | standard | `82d46f8` |
| ia5.2 | standard | `56ec2bf` |
| ia5.3 | standard | `daa7f3c` |
| ia5.4 | standard | `59d8c05` |
| ia5.5 | standard | `b008ff5` |
| ia5.6 | advanced | `f00900e` |
| wu1 | standard | `986fecc` |
| 8ko | standard | `33078e0` |
| cb9 | light | `df4786d` |
| eaa.1 | standard | `e27f6a7` |
| eaa.2 | standard | `ea10378` |
| eaa.3 | standard | `a894938` |
| eaa.4 | **none** | `28ae2a2` |
| 47b | standard | `6456fa3` |
| 92u | standard | `c9ac182` |

Headline: the Spec CLI gained wikilink→markdown-link rewriting, 14 shipped
documents moved into `specs/` with the vendoring inverted, and
`scripts/validate-corpus.py` was deleted — the three turned out to be one
dependency chain, not three tasks. **The run also stalled silently for ~20
minutes at wave 6b behind a rejected push, and two defects reached CI that no
local gate could see.**

Findings harvest: 64 `retro:` comments across all 15 dispatched beads, none
with zero. Every one is reproduced verbatim in section 5, including the tail
that warrants no RFC.

## 1. Grooming

**Under-groomed beads: two, in opposite directions.**

`92u` carried an acceptance criterion that excluded the only
correct implementation. It said the replacement CI step lands "in place of" the
deleted one, which pins it to the `checks` job — a job with no Go toolchain,
so it cannot build the binary the step needs (`92u#2`). The implementer moved
it to the `corpus` job and recorded that the criterion's wording "does not
admit" the move. The same bead found two consequences grooming never named: a
second spec root produces a second index artifact that `.gitignore` did not
cover, and the pinned `--only` list leaves `rfcs/` with no zero-document guard
(`92u#3`).

`eaa.1` was groomed against a measurement that went stale inside
its own queue. Its description and the epic's both say "~50" illustrative
wikilinks; the true pre-fix count was **3** (`eaa.1#1`). The figure was taken
before `wu1` landed, and `wu1` — in the same run, one wave
earlier — sharply trimmed the population being counted. A 17× error, correct
when written.

**A mid-run admission skipped the readiness gate entirely.**
`eaa.4` was filed and dispatched during wave 6 and carries
**no `tier`, no `tier_reason` and no `est_size`** — verified from `bd show`
metadata, all three null. The operating principles say Phase 1.5 "holds for
beads pulled into scope mid-run too"; it did not. Nothing went wrong, because
the orchestrator hand-placed it in a split wave, but the bead was dispatched
outside the gate that exists to make dispatch safe.

**Tier predictions: zero escalations across 15 dispatched beads, and that is
not a clean bill of health.** Thirteen of fifteen ran `standard`, one `light`
(`cb9`), one `advanced` (`ia5.6`), one untiered (`eaa.4`). The field guide is
explicit that escalations surface under-routing while over-routing is
invisible and paid for on every run, so a zero-escalation run "is worth
examining, not celebrating." One data point suggests the floor is too high:
`cb9` was the only `light` bead and it reported "No friction encountered. Bead
was well-specified and all decisions were already pinned by the Design"
(`cb9#1`). Several `standard` beads reported the same — `wu1#4` calls itself
"Clean bead otherwise". Whether those could have run `light` is untestable
from this run's evidence, which is exactly the asymmetry the rubric warns
about. **Null result on direction: no evidence of under-routing; no mechanism
existed to detect over-routing.**

## 2. Specs

**Three documents were found wrong, thin, or self-contradictory.**

`[[cross-document-references]]` v0.3 is now falsified by the model this run
introduced. Its rationale claims a path relative to the citing file "is the
only spelling that works from every location the file is read." Vendoring
creates files read at two depths, so no single relative string is correct.
`eaa.3#2` and `eaa.4#2` hit this independently, on the same two files, and
invented the same workaround (convert to a wikilink). The standard also still
offers `c4-guide.md` and `ears-guide.md` as its canonical examples of
documents cited by path *because they sit outside the set* — and this run moved
both into the corpus.

**The invariant that governs what ships is in no document at all.** "A shipped
document may cite only other shipped documents" exists only in the description
of `eaa`, now closed; searching `specs/` returns nothing, and no
document mentions the vendoring model or the `GENERATED COPY` marker. It is
also not closed under transitivity, which `eaa.2#2` discovered at consume time
and resolved by demoting a citation to inert prose — satisfying the invariant's
letter while leaving the adopter the dangling pointer it exists to prevent.

The feature blueprint's own gaps are covered by `rfc-0019` (authored by
`ia5.6`, resting on `ia5.1`–`ia5.5` plus DoD findings) and are not restated
here.

## 3. Plan

`## Amendments` recorded six entries: one mid-run bead admission with a wave
split, one interface hazard carried by hand into a later dispatch prompt, one
correction to the Phase 0 CI record, one push-mode clarification, one note that
the draft PR already existed, and one on trigger cadence.

**Two collisions the initial bucketing could not have caught, because the
collision key is file-level.** Wave 2 ran `ia5.2` and `ia5.4` concurrently on
different components — the `c4_changes` test passed — but one *produces* the
link map and the other *consumes* one, and neither could see the other's types.
`ia5.2` defined its own `mdlinks.LinkMap` and leaned on Go's nil-vs-non-nil
distinction; `ia5.4` independently built `cli/cmd/linkmap.go` and split one
declared fault into two rule ids. Nothing broke, and nothing guaranteed that.
The hazard reached `ia5.5` only because the orchestrator wrote it into that
bead's dispatch prompt by hand.

The second: `47b` *measures* `plugin/skills/*/references/**` with a file-count
floor while `eaa.4` *changes* that tree. Both carry `c4_changes: []` and they
co-edit no file, so every test said concurrent was safe. Running them together
would have baked a floor true for no commit. Caught by splitting wave 6 by
hand.

**What would have produced a better initial bucketing:** an interface-seam key
derived from the beads' own text (one Design says "produces", another says
"consumes"), and a measured-surface key for a bead whose Acceptance names a
count or floor over a path it does not edit.

## 4. Skill

**`bead-bonanza` misled in one place and went silent in another, and the first
cost the run 20 minutes.**

Pre-flight measures push *mechanics* thoroughly — `BB_PUSH_NEEDS_REFSPEC`,
`BB_GIT_CONFIG_WRITABLE`, and a real probe push for the Dolt remote — but
never push *scope*. The container's PAT lacks the `workflow` scope, so wave
6b's push carrying `.github/workflows/checks.yml` was rejected outright, and
the rejection failed the whole push rather than that file. Every liveness
signal the skill and `recover-sprint` prescribe read "finished": zero
worktrees (removed before the push), a clean tree, the bead `closed`, and the
ledger's last phase `pushed`. Only `git log origin/<branch>..HEAD` disagreed,
and it is in neither skill's signal list.

The silence: the per-bead prompt gives a subagent two channels for discovered
work — "file a new bead via `bd create`" at step 3, and a `retro:` comment at
step 6 — with no rule for which wins. `eaa.3#1` identified a real scope gap
precisely, named the two beads that created it, called it "a real scope gap,
not this bead's to close", and filed nothing. It survived to be caught two
waves later by the parent epic's AC-1 **count**; had that criterion been
unquantified, the epic would have closed 13/14.

**A subagent improvised around a prescribed script.**
`scripts/regenerate-cli-snapshots.sh` reported `regenerated <case>` for every
case while writing nothing for nine of them — the generator has no branch for
`checks.err`. This is a second instance of the class `krb`
already records, with a different cause. It surfaced only because two fixtures
embedded platform-specific error text and Windows CI failed.

**Null result on `retro:` comments citing the skill directly:** of 64
findings, only three name skill mechanics, and all three are incidental
(`ia5.4`'s config placement, `eaa.4`'s workflow-edit routing, `92u`'s
`.gitignore` interaction with the close ceremony). The skill-level evidence in
this report is the orchestrator's and the amendments log's, not the subagents'.

## 5. Every harvested finding, verbatim

Numbered `<bead-id>#<ordinal>` by position within each bead. Reproduced in
full, including findings that warrant no RFC — `triage-rfc` defers on
recurrence, and a finding filtered out here can never be observed appearing
twice.

### ia5.1

**ia5.1#1** — Regenerating schemas/finding.strata changes the embedded profile schema bytes (cli/core/schema/spec-model.schema.json), which feed corpus.inputHasher's 'profile-schema' field — so four committed CLI snapshot fixtures under cli/testdata/cases/ (index-default, index-out-custom, index-out-stdout, index-unresolved-reference) failed on a stale input_hash until scripts/regenerate-cli-snapshots.sh was run. Neither the bead nor the FND delta's 'regeneration is safe' note mentions this fixture-invalidation side effect of adding a variant.

**ia5.1#2** — Silently decided — RuleId constant names RuleUnmappedCitation/RuleUnresolvedCitation, tags unmapped-citation/unresolved-citation, and struct field names CitingPath/MapPath (mirroring OtherPath/SourcePath/TargetPath elsewhere in finding.strata). The bead's Design fixed the field content and the TGT bead's text fixed the rule-id string values (links/unmapped, links/unresolved), but named nothing else, and no other item's scope covers adding these RuleId constants.

**ia5.1#3** — Considered adding the two new rules to cli/core/finding's referenceRules map (IsReferenceRule), since citation findings are raised against the referring document like the existing reference/* rules. Abandoned it: that map is consumed only by cli/core/corpus/check.go's --only single-file narrowing for the check command, which the markdown-links links subcommand (CMD/TGT, out of scope here) does not go through, so leaving it untouched keeps this bead's diff to schemas/finding.strata and cli/core/finding.


### ia5.2

**ia5.2#1** — the Design's "slug names a corpus document" isn't tied explicitly to ReferenceResolution's three states (resolved / non-member / unresolved); I read it as meaning only `resolved` (a true corpus member with valid frontmatter), treating `non-member` (an excluded or parse-failed file that still has a real path) the same as a bare unresolved slug. Naming the states explicitly in the bead text would remove the ambiguity — a reader could plausibly argue a non-member's own path should still be derivable from.

**ia5.2#2** — MAP (ia5.4) had not landed, so no link-map type existed yet to consume; I defined my own `mdlinks.LinkMap` (`map[string]string`) and used Go's nil-vs-non-nil map distinction to separate "no map supplied" from "a map supplied but resolving zero entries" — AC-SPECMD-006.3 needs the latter to still raise links/unmapped. MAP's own implementation should pass mdlinks a non-nil map even when the YAML decodes to zero entries, or the two components will disagree about whether a map was "given".

**ia5.2#3** — silent decision — the no-map derived path never gets a "./" prefix for a same-directory target (e.g. `adr-012.md`, not `./adr-012.md`). The Design says "no other normalization" but doesn't say whether a leading `./` counts as one; I read it as the minimal relative form.

**ia5.2#4** — silent decision — internal helper names (resolveOne, anchorFragment, slugifyAnchor, relativeTo, dirParts) and the exported shape (Resolve(refs, Options) ([]*string, []Finding), targets aligned with refs by index rather than a per-reference wrapper struct) are mine; the bead prescribed only the package path.


### ia5.3

**ia5.3#1** — The Design's Behavior list mentions only the [[...]] construct and never says what to do with a frontmatter-origin reference, which shares the same Reference/Resolution shape but has a Position covering a raw YAML value rather than bracket-delimited text. I read the FRD's Terminology (wikilink is defined strictly as [[slug]]/[[slug#Anchor]] body syntax) and the blueprint's Overview ("replace ... each targetable wikilink") as scoping substitution to body-origin references only, and filter frontmatter-origin references out in Substitute rather than rewriting them — treating them as forbidden-to-mutate content the same as any other unmentioned byte range.

**ia5.3#2** — Silent implementation choice: the construct's extent is derived by widening Reference.Position by hardcoded 2-byte '[[' / ']]' constants local to the mdlinks package, duplicating (not importing) the private delimiter constants references.go already defines. This is deliberate — the two packages should not share unexported symbols — but a diff reader could mistake it for drift if the wikilink notation ever changed in only one place.

**ia5.3#3** — Clean bead otherwise. mdlinks's existing shape (Resolve taking refs+Options, returning targets aligned by index) composed cleanly with Substitute taking the same refs plus targets — no awkwardness for CMD to inherit; the two calls chain directly (targets, findings := Resolve(refs, opts); out, count := Substitute(src, refs, targets)).


### ia5.4

**ia5.4#1** — Split the map-reading fault into two rule ids -- linkmap/invalid (unreadable, unparseable, or a duplicate top-level key) and linkmap/missing-path (a record with no path) -- mirroring this package's existing config/schema vs config/missing-property split, rather than one rule covering all three faults the Design lists together.

**ia5.4#2** — A record's path can't be told apart, as decoded, between absent and present-but-empty-string without a pointer field; I treat path: "" the same as a record with no path key at all, since the Design's "record with no path" doesn't address an explicit empty value and an empty path is unusable either way.

**ia5.4#3** — Link-map findings carry a whole-file span rather than a position inside the YAML, unlike specs.toml's TOML scanner (locateConfig) which locates a precise key/value span. goccy/go-yaml's decode error has no publicly errors.As-able positioned type the way BurntSushi/toml's ParseError does, and the Design only requires naming the file and the fault, not a precise position, so no YAML position scanner was built.

**ia5.4#4** — Otherwise clean -- the Design was fully closed (duplicate-key handling pinned as stated), and the referenced FRD ACs plus the existing cmd/config.go, cmd/toml.go and cmd/finding.go precedent were sufficient to implement without needing information the bead didn't carry.


### ia5.5

**ia5.5#1** — the zero-entry-map wiring constraint PASSES. Snapshot case links-map-empty supplies a 0-byte --map file; LoadLinkMap decodes it to a non-nil, empty LinkMap (make(LinkMap, 0) is never nil even when the source map is nil); passed through unchanged to mdlinks.Options.Map, the corpus-resolved citation the map does not cover raises links/unmapped (error), exit 1, no rewritten document on stdout. Verified both via 'go test ./cmd/...' and by running the built binary directly.

**ia5.5#2** — seam friction, not at mdlinks/linkmap (that composed exactly as ia5.3/ia5.4 described) but at CMD/corpus. The Design says links 'opens the corpus warm, as check, refs, slug-exists and next-id do' — but the shared openCorpus helper those four use best-effort persists a freshly-built index to specs.json, which directly conflicts with AC-SPECMD-005.3 ('creates, modifies and deletes no file') on a cold cache. Neither the bead nor the FRD reconciles this. Resolved by adding a narrower openCorpusReadOnly (cli/cmd/links.go) that opens through corpus.Open exactly as openCorpus does but skips the persist step; a Go test (TestLinksCreatesModifiesAndDeletesNoFile) runs on a cold cache specifically so a regression back onto openCorpus would be caught.

**ia5.5#3** — what I tried and abandoned. My first implementation read the operand with a bare os.ReadFile and no CRLF normalization, on the assumption that corpus.Build's own normalizeLineEndings (applied at its one file-read point) covered every file this feature touches. It doesn't: the operand need not be a corpus member and is read through a second, independent path this package owns. cmd/portability_test.go's CRLF replay caught the mismatch immediately (three links cases failed only under the CRLF checkout, not the LF one). Fixed by duplicating corpus's normalizeLineEndings (it's unexported) into cmd/links.go and applying it right after the operand read.

**ia5.5#4** — spec gap. No AC-SPECMD id exists for 'the named document cannot be read' (blueprint Behavior + this bead's checklist both require the behavior, but unlike the analogous link-map-unreadable path — AC-SPECMD-005.8/.9 — the FRD assigns it none). The links-unreadable-operand and links-no-strict-flag snapshot cases cite = [] for this reason rather than a fabricated id.

**ia5.5#5** — silent decisions a diff-reader would not recognise as decisions. New rule id document/unreadable (cmd/finding.go) for the unreadable-operand case, since no existing rule fit and none was pinned. finding.FrontmatterParse is reused (not a second unreadable-style rule) when the operand's frontmatter block is present but malformed, and that path is treated as blocking — it returns before extraction rather than falling through with doc==nil, because falling through would scan the raw frontmatter block as body text for [[...]] constructs. Internal helper names (openCorpusReadOnly, finishLinks, normalizeOperand, referenceProperties), the --map flag help text, and the stderr count-line wording ('%d link(s) rewritten') were all unpinned and picked freely. Snapshot case names: links-no-map, links-map, links-map-empty, links-map-incomplete, links-unresolved-warning, links-unreadable-map, links-unreadable-operand, links-no-strict-flag, help-links.


### ia5.6

**ia5.6#1** — indexed finding 1 (ia5.5, warm-open vs AC-SPECMD-005.3) CONFIRMED and the resolution is right. openCorpus (cli/cmd/root.go:170) does `_ = corpus.Write(c.Index(), dest, nil)` when !warm; openCorpusReadOnly (cli/cmd/links.go:153) does not, and TestLinksCreatesModifiesAndDeletesNoFile digests the tree before and after on a cold cache. The durable document now says so — container-spec-cli.md's Container Context names the exception explicitly — so nothing was left living only in the retired blueprint; what remains is that Change Guidance still does not tell the NEXT subcommand author the two helpers differ (rfc-0019 P2).

**ia5.6#2** — indexed finding 2 (ia5.5, missing AC id for the unreadable operand) CONFIRMED as a real FRD gap. Read every AC in markdown-links-frd.md: REQ-SPECMD-005 covers the link map twice (AC-SPECMD-005.8/.9) and the operand zero times, while the blueprint's CMD Behavior states the operand behavior as the one bullet in that list carrying no AC. document/unreadable ships with a rule id, two snapshot cases and help-visible surface, cited by nothing. Proposed as rfc-0019 P1 (append AC-SPECMD-005.11/.12); not fixed here, the FRD is approved and out of this bead's AC.

**ia5.6#3** — indexed finding 3 (ia5.3, frontmatter-origin references filtered not rewritten) CONFIRMED — the FRD supports the reading. Terminology defines Wikilink strictly as `[[slug]]`/`[[slug#Heading Text]]`, every REQ-SPECMD-001 criterion is phrased over 'wikilink', and AC-SPECMD-001.5 forbids altering anything else; a frontmatter Position covers a raw YAML value and is not that. substitute.go:41 filters on `ref.Origin.AsBody()`. Already fixed at the durable layer too: module-document-core.md's Markdown Links section now pins 'Only a body-origin reference is rewritten.' No RFC proposal.

**ia5.6#4** — indexed finding 4 (ia5.3, duplicated `[[`/`]]` constants) CONFIRMED as fact, declined as an RFC finding. cli/core/mdlinks/substitute.go:15-18 and cli/core/references/references.go:41-42 each define the pair unexported; the mdlinks copy is used only by constructSpan to widen a Position, never to search. No document is wrong and none would have prevented it, so it went to rfc-0019's Considered and Not Proposed rather than a proposal.

**ia5.6#5** — indexed finding 5 (ia5.2/ia5.5, nil vs non-nil LinkMap) CONFIRMED passing, re-verified independently of ia5.5's word. LoadLinkMap builds `make(LinkMap, len(raw))` (cli/cmd/linkmap.go:73), never nil; links.go only assigns it when mapPath != "", so nil still means 'no map'; mdlinks.Resolve branches on `options.Map != nil` at two sites. Snapshot case links-map-empty (0-byte map) exits 1 with links/unmapped and no document on stdout, and mdlinks_test.go:175 covers the same at unit level under AC-SPECMD-002.1.

**ia5.6#6** — indexed finding 6 (ia5.2, 'slug names a corpus document' ambiguous across the three resolution states) PARTLY REFUTED. The choice was right, but the FRD already settles it and did not need naming the states: Terminology says 'Corpus document — a document beneath the spec root, indexed, and citable by slug', which excludes a non-member. The ambiguity was in the work order's paraphrase, not in the spec — so it is a work-order-authoring concern for the retro-skill lens, and naming ReferenceResolution in an FRD would bind a requirement to an implementation type. Recorded in rfc-0019's Considered and Not Proposed with both reasons.

**ia5.6#7** — indexed finding 7 (ia5.1, schema regeneration invalidates snapshot fixtures) CONFIRMED still undocumented. container-spec-cli.md's Data section does list 'the active profile's schema and per-type table' among the input_hash's inputs, so the first link of the chain is written down — but the consequence is not, in testing-standards.md's Snapshot Case Structure, in schemas/README.md, in scripts/generate-schemas.sh, or in AGENTS.md's targeted-test row for schemas/*.strata (which names generate-schemas --check and go build only). Proposed as rfc-0019 P5.

**ia5.6#8** — the 31-AC citation audit is CLEAN — no dangling id, no uncited criterion. markdown-links-frd.md defines exactly 31 AC-SPECMD ids; the blueprint's AC Coverage table carries exactly those 31, with no row naming an id the FRD lacks and no criterion missing a row. All 31 are cited: 22 by a `// AC:` comment on a Go test, 9 by a snapshot case manifest's `cites` (links-map, links-no-map, links-map-empty, links-map-incomplete, links-unresolved-warning, links-unreadable-map) — testing-standards.md's Acceptance-Criteria Traceability accepts either form. Zero orphaned citations: every AC-SPECMD id appearing in any test, manifest or source comment resolves to a defined criterion.

**ia5.6#9** — one AC citation is at the declared level but from the wrong harness. AC-SPECMD-005.3 is declared integration and is proven by cli/cmd/links_test.go's TestLinksCreatesModifiesAndDeletesNoFile, not by a snapshot case — a case is pure data and cannot digest a tree before and after. testing-standards.md's Snapshot Cases section says observable CLI behavior SHALL be verified by snapshot cases rather than by unit tests of the command surface, so it is technically non-conforming. This is the third instance of the shape rfcs/inbox/feature-blueprint/rfc-0015 P2 already proposes a rule for, so I recorded it as corroboration there rather than re-proposing it in rfc-0019.

**ia5.6#10** — two DoD-verification findings the five implementers could not see, both filed as beads. (1) 38 backticked path citations across specs/ name documents eaa.2/eaa.3 moved into the corpus — including two this phase consumed, the blueprint's own DoD pointing at plugin/skills/spec/references/ac-traceability-standards.md and the FRD pointing twice at cross-document-references.md — and spec check --strict exits 0 over all of them because a backticked path is by construction the citation it does not resolve (jn6, rfc-0019 P4). (2) All five delta commits edited module-document-core.md and container-spec-cli.md without moving either updated_at; frontmatter-standards.md defines the property but obliges nobody to bump it (thb, rfc-0019 P3).

**ia5.6#11** — what I tried that did not work. I first treated the 9 integration ACs cited only in case.toml as an uncited hole and was about to add them to TestSnapshotCases's own `// AC:` comment, on the precedent that AC-SPECREF-003.1/.2/.3/.5 appear both in a case manifest and on that function. testing-standards.md's Acceptance-Criteria Traceability refuted it — 'cited by at least one snapshot case manifest OR at least one unit test' — so the harness comment is redundancy, not the citation mechanism, and editing it would have been churn. Also built the wrong binary once: `go build ./` in cli/ reports 'no Go files' (the main package is ./spec), and the stale binary from a prior path silently answered the first spec check run.

**ia5.6#12** — what I decided silently — the two implicit decisions this bead reserved. (1) RFC grouping: ONE RFC bundling all five proposals, filed as rfcs/inbox/feature-blueprint/rfc-0019-implementing-markdown-links.md, matching rfc-0015's placement for the analogous reference-integrity implementation phase so recurrence is visible by listing one directory — even though P3/P4/P5 change standards documents, because the majority concern the blueprint's own machinery (its DoD, its Behavior bullets, its AC Coverage table). (2) RFC wording: all mine; each proposal states normative text ready to paste. Also: I ticked the blueprint's DoD checkboxes and annotated three of them with what was verified, rather than retiring it with six empty boxes.

**ia5.6#13** — for eaa.4 — the retired blueprint's References section still cites `plugin/skills/spec/references/c4-guide.md` by backticked path (markdown-links-blueprint.md, References). I did not touch it: eaa.4 moves that document into specs/standards/ concurrently and owns the citation sweep for it. Once it lands, that line and the FRD's sibling citations become wikilink candidates under jn6.


### wu1

**wu1#1** — Left rfcs/retired/rfc-0007-spec-architecture-templates.md and rfcs/inbox/spec/rfc-self-greenfield-anchors-and-split-numbering.md untouched even though both contain the literal string 'guide-c4'. Both are point-in-time historical records (a retired RFC's invocation table, and a retro noting what a specific prior template run cited) rather than live corpus citations — neither is a [[guide-c4]] wikilink or a resolvable path reference, and rewriting them would edit history rather than fix a live reference. Not obvious from the diff alone.

**wu1#2** — Silent decision — did not bump updated_at or reword the description/title. The AC list only requires type/slug/schema-branch satisfaction, wikilink removal, and backticked-path resolution; nothing calls for a description rewrite, and the current description ('How to map concrete architectural elements to C4 levels...') reads fine under the Standards branch too, so I left it untouched to keep the change to exactly the retype.

**wu1#3** — Moved the c4-guide.md row in cli/profile/profile_test.go's realSlugs table from the guide cluster to the standards cluster (not just edited the type/slug values in place) to keep the table's existing same-type grouping convention intact. A diff-reader would see the line move and could read it as more churn than intended without this note.

**wu1#4** — Clean bead otherwise — the Design's schema-branch delta (Standards adds a required version field vs Guide's optional one) was correctly anticipated: the document already carried version: "0.2", so no new frontmatter property was needed, and scripts/validate-corpus.py plus spec check --strict both confirm the new frontmatter validates with 0 findings.


### 8ko

**8ko#1** — Observed pre-fix failure (reproduced per the bead's recipe): `rfcs/outbox/retro-spec/rfc-agent-plugins-findings-capture-harvest.md: slug/pattern-mismatch error 7:7 "rfc-agent-plugins-findings-capture-harvest" does not match pattern "^rfc-([0-9]{4}(-[a-z0-9]+(-[a-z0-9]+)*)?|self-[a-z0-9]+(-[a-z0-9]+)*)$"`. After the rename+frontmatter fix, the identical command reported "no findings" and `slug/pattern-mismatch passed`, exit 0.

**8ko#2** — What I tried that didn't work: the bead's recipe writes the config to /tmp/r.toml with root="rfcs" and doesn't say what "build the CLI" means concretely. Two things bit me. (1) The config's workspace.root resolves relative to the CONFIG FILE's directory (cli/cmd/config.go: filepath.Join(filepath.Dir(path), root)), not cwd — so a config at /tmp/r.toml with root="rfcs" looks for /tmp/rfcs, which never exists; I had to place the config at the worktree root instead. (2) `go build ./cmd/...` doesn't produce a runnable binary — cmd/ is a library package (no func main), so go build silently wrote a non-executable .a archive there; the real entrypoint is cli/spec/main.go. Neither is a bug, just missing from the recipe as literally transcribed.

**8ko#3** — Silent decisions: (1) found a third citation of the old slug string beyond the two the bead pre-located — rfcs/specs.json, an untracked spec-check index/cache artifact (matches .gitignore's documented pattern for specs/specs.json, just at a different path) — deleted it rather than editing it, since it's a regenerable build artifact, not a tracked citation. (2) Follow-up bead title/description and priority (P3) were my own call — the bead only said to record its id.

**8ko#4** — Filed follow-up bead kau (skill-rfc-template.md prescribes an RFC slug form the profile rejects), linked to 8ko via discovered-from, per the bead's Design instruction to file rather than fix the template disagreement.


### cb9

**cb9#1** — No friction encountered. Bead was well-specified and all decisions were already pinned by the Design.

**cb9#2** — Decided silently: output prefix is 'info' for advisory findings; finding type is 'template/omitted-optional' to distinguish it from the error-level findings.


### eaa.1

**eaa.1#1** — Only c4-guide.md needed an edit (lines 136, 138: backticked `[[component-<name>]]`, `[[module-<name>]]`, `[[container-<name>#Element Name]]`). The other 13 of the 14 corpus-shaped documents were already following the backtick-illustrative-wikilink convention when read — cross-document-references.md, document-readiness-checklist.md, template-authoring-standards.md, spec-cli.md, ears-guide.md, anti-patterns.md, work-order-field-guide.md, frontmatter-standards.md all had every illustrative [[...]] already inside a code span, and ac-traceability-standards.md, blueprint-readiness-checklist.md, work-order-definition-of-ready.md, document-type-boundaries.md, findings-capture-protocol.md had none needing it. The bead's ~50-dangling estimate was measured before wu1 landed (which trimmed c4-guide.md's wikilink count sharply); the actual pre-fix dangling count in the 14-doc corpus was 3, not ~50.

**eaa.1#2** — [[security-standards]] occurs 3 times across the 14 documents (cross-document-references.md x2, template-authoring-standards.md x1) and all 3 were already backticked as illustrative examples. Per the Design, no such document exists in specs/standards/ today, so it reads like a document someone intended to write; not inventing it, per instruction.

**eaa.1#3** — Initial pass with a hand-rolled Python code-span detector produced a false BARE classification for c4-guide.md:136/138 even after the fix was applied, because a stale .pyc in /tmp/claude-0/__pycache__ shadowed the corrected script. Discovered by re-running the same logic inline vs. via the cached module and comparing. Deleted the pycache before trusting the script's output; the corpus-wide count of 17 resolving citations it then reported matches the bead's own RESOLVES estimate exactly, which is what gave confidence the classification was complete.

**eaa.1#4** — Measured before/after with a throwaway corpus (specs.toml root=corpus, profile spec-driven-dev) built from spec CLI (cli/spec, built locally since not on PATH). BEFORE (14 docs only): 6 reference/unresolved findings — 3 illustrative in c4-guide.md + 3 genuine citations (adr-012, document-relationship-standards x2) that dangle only because their targets aren't among these 14 files. AFTER: 3 findings, all the same genuine-citation dangles. Adding ADR-012 and document-relationship-standards.md into the throwaway corpus resolves those too. A full merge of specs/ + the 14 docs reports 0 findings. Resolving citations measured at 17 (template-authoring-standards x10, document-relationship-standards x2, ac-traceability-standards x1, adr-012 x1, frontmatter-standards x1, cross-document-references x1, work-order-definition-of-ready x1), matching the bead's ~17 estimate exactly even though the ~50 dangling estimate did not hold up.


### eaa.2

**eaa.2#1** — The move broke a reference check-skills.sh caught but the AC list didn't name: plugin/skills/spec/references/spec-cli.md (a guide, out of scope until eaa.3) cited document-readiness-checklist.md by bare filename, sibling-relative. Fixed by repointing it to shared/document-readiness-checklist.md — moving a document breaks bare-filename citations from files that stay behind, not just from documents that move with it.

**eaa.2#2** — Silent decision — frontmatter-standards.md's two [[document-relationship-standards]] wikilinks were backticked (inert prose) rather than shipping document-relationship-standards.md via the map, because that document itself cites adr-009/adr-012/two FRDs that aren't shipped — adding it would cascade the ships-only-what-it-cites invariant outward instead of resolving it.

**eaa.2#3** — retro-spec/SKILL.md's own prose ('skills that capture or harvest findings vendor it...') implicitly excluded retro-spec itself, since it used to be the document's author. Updated the wording to say retro-spec vendors its own copy too, matching the design's 'including the spec skill's own' note extended to retro-spec (the actual original owner of findings-capture-protocol.md).


### eaa.3

**eaa.3#1** — Confirmed at start (per this bead's own AC item): wu1 already retyped c4-guide.md to standards, and eaa.2's Design explicitly declined to claim it either ('it is not [handled in eaa.2]; guides are separate'). So c4-guide.md is now a standards-typed document that ships nowhere and moves in neither bead — a real scope gap, not this bead's to close. This bead covers the 5 remaining guides.

**eaa.3#2** — work-order-template.md (a template, not moved) bare-cited work-order-field-guide.md by filename. After the move that citation is only correct from a directory where both files are flat siblings — true in every OTHER skill's vendored references/shared/, but not from the authored spec/references/ location, where the guide's copy is now one level down in shared/. A single relative-path string can't be correct in both places; changed both citations to [[work-order-field-guide]] (a real wikilink now that the guide is a corpus member), matching the adjacent [[work-order-definition-of-ready]] citation on the same line and sidestepping the ambiguity entirely.


### eaa.4

**eaa.4#1** — shipped the generated copy only to `spec` (MAP entry `spec\tspecs/standards/c4-guide.md`). Determined by grepping for `c4-guide.md` across plugin/skills/: it appeared inside groom-beads and groom-feature-blueprint too, but only as citations *within* templates that spec itself vendors out to them (work-order-template.md, feature-blueprint-template.md) — those became [[c4-standards]] wikilinks instead (see next comment), so groom-beads/groom-feature-blueprint need no direct copy of c4-guide.md itself.

**eaa.4#2** — repointed citations in both known shapes. Shape 1 (bare-filename from a file that stays behind, checked live by check-skills.sh's path-resolution rule): SKILL.md's `references/c4-guide.md` -> `references/shared/c4-guide.md`, and the five single-skill templates (system/container/module/component-blueprint-template.md, contract-template.md) bare `c4-guide.md` -> `shared/c4-guide.md`. Shape 2 (a relative path that can't work from both the authored spot and a shared/ copy one level deeper): feature-blueprint-template.md and work-order-template.md are each vendored into a second skill (groom-feature-blueprint, groom-beads), so their c4-guide.md citations became [[c4-standards]] wikilinks, mirroring eaa.3's fix to the same two files for work-order-field-guide.md.

**eaa.4#3** — what did not work / dead end avoided — I initially considered converting every specs/ prose mention of c4-guide.md (ADR-004/005/007/008/009/010/012, contract-specs-configuration.md, specs/standards/blueprint-readiness-checklist.md) to [[c4-standards]] for internal consistency, since c4-guide.md is now a corpus member. Checked precedent first: eaa.2 left an identically stale backtick path to template-authoring-standards.md unconverted in ADR-012, and neither check-skills.sh (scans plugin/skills/ only) nor spec check --strict (validates [[slug]] forms, not backticked paths) enforces it. Left all of these untouched to match established scope and avoid an unbounded sweep; recorded as a silent decision below.

**eaa.4#4** — decided silently — (1) MAP entry placement: inserted `spec\tspecs/standards/c4-guide.md` alphabetically between blueprint-readiness-checklist.md and cross-document-references.md in the existing spec cluster, matching the list's established a-z ordering. (2) Left prose backtick mentions of c4-guide.md in specs/adrs/, specs/contracts/contract-specs-configuration.md, and specs/standards/blueprint-readiness-checklist.md unconverted to [[c4-standards]] — narrative only, not gate-checked, matches eaa.2's precedent of leaving ADR-012's own stale path alone. (3) Found two more stale c4-guide.md path citations in specs/features/ (anchors-frd.md, markdown-links-blueprint.md) but specs/features/ was off-limits for this bead (concurrent ia5.6); filed 4me discovered-from this bead rather than touching it.


### 47b

**47b#1** — Glob floor: measured plugin/skills/*/references/**/*.md at exactly 52 files on the post-restructure tree (confirmed by two independent methods: find and Python pathlib.glob). Set MIN_FILES=40, well below the measured count, matching check-template-drift.py's ratio. The bead text names no specific floor figure to compare against, so nothing was stale there — the stale thing was the file population itself, which the bead's own note already flagged.

**47b#2** — Proof the script detects a break: appended a dangling link to a scratch copy of frontmatter-standards.md ([dangling target](does-not-exist.md)) and ran the script — 'error  plugin/skills/spec/references/shared/frontmatter-standards.md: link/unresolved: `does-not-exist.md` does not resolve to a file co-located with this document', exit=1 (both normal and --quiet mode). Reverted, re-ran, exit=0. Separately proved the glob-floor guard fires: pointed SEARCH at a nonexistent extension in a scratch copy of the script, got 'search/underpopulated: matched 0 reference documents against a floor of 40', exit=1.

**47b#3** — Overlap with jn6: jn6 covers 38 dangling backticked path citations in specs/ prose, a different surface (prose citations, not markdown links, and not shipped payload). check-reference-links.py is scoped to plugin/skills/*/references/**, so it does not see specs/ at all and cannot catch jn6's cases — confirmed no overlap in practice, not just in theory.

**47b#4** — What did not work as the bead's AC literally describes: 'Run against the current tree it reports the known live failure' cannot be satisfied today. The frontmatter-standards.md citation of document-relationship-standards is still `[[document-relationship-standards]]` (wikilink form) in both specs/standards/frontmatter-standards.md and its generated plugin/skills/spec/references/shared/ copy — the markdown-link rewrite capability (ia5.5) exists but has not been run over shipped copies yet. Design explicitly forbids checking wikilinks, so the script correctly reports 0 errors on the current tree; manually verified all 133 non-external link occurrences across the 52-file scope already resolve. This AC bullet is only satisfiable after a separate rewrite step runs, which is out of this bead's scope.

**47b#5** — Silent decisions: (1) script name scripts/check-reference-links.py, mirroring check-template-drift.py's check-*.py convention over sibling shell scripts, since link parsing reads more naturally in Python regex. (2) Per design's literal instruction not to reimplement extraction (which explicitly names code-span skipping as the CLI's job), the regex runs over raw text with no fence/code-span exclusion — verified safe on the current tree by checking the only two fenced code blocks with link syntax nearby (generated/finding.md, generated/vocabulary.md) contain no markdown-link syntax inside the fence. (3) --quiet flag name/behavior copied verbatim from check-template-drift.py rather than inventing --check, for gate-command consistency across the two scripts. (4) CI step placed in checks.yml immediately after 'Vendored shared references are current', ungated (no continue-on-error), since it passes clean today. (5) .github/workflows/checks.yml edit made directly per the orchestrator's instruction to attempt it normally; not verified pushable from this container (defer_push).


### 92u

**92u#1** — Chose `rfcs/specs.toml` with `root = "."` over a root-level `specs-rfcs.toml` with `root = "rfcs"`. The CLI resolves [workspace].root relative to the config file's directory, so co-locating the config with the corpus it describes makes the pairing self-evident and keeps `spec check` working from inside rfcs/ with no --config flag.

**92u#2** — The bead should have pinned WHICH JOB the new step lands in. The replacement needs the `spec` binary and the `checks` job is bash+python3 only with no Go toolchain, so the step has to go in the `corpus` job that already builds it — a job move, not an in-place swap, which the acceptance criterion's "in place of" wording does not admit.

**92u#3** — Two consequences the bead did not name. (1) A second spec root means a second index artifact — `spec check` writes `rfcs/specs.json`, and .gitignore only listed `specs/specs.json`, so without a new entry the close ceremony would stage a 99KB build artifact. (2) The pinned --only list renders corpus/empty `not-requested`, so rfcs/ has no zero-document guard either; folded into the follow-up bead y09 rather than widening the list here.

**92u#4** — Read the 'no remaining reference anywhere in the repository' criterion as covering live surfaces only — CI, AGENTS.md, schemas/README.md, specs/. The ~30 remaining hits are historical: archived RFCs and sprint reports under rfcs/ that cite the script as evidence of a past state, plus the in-flight run dir under .bonanza/. Rewriting those would falsify the record.


## 6. Routing

Four RFCs were authored from this report's evidence. **They were filed before
this report existed** — the run's Phase 4 was skipped and reconstructed
afterwards — so each re-inlines evidence the skill says should be cited from
here instead. They have been amended to cite this report's finding ids; the
report is the testimony, the RFC is the proposal built on it.

| Route | RFC | Carries |
|-------|-----|---------|
| `retro-skill` → `bead-bonanza` | `rfc-self-push-scope-and-discovered-work-channel` | §4's push-scope stall; `eaa.3#1`'s unfiled scope gap |
| `retro-skill` → `plan-sprint` | `rfc-self-collisions-the-c4-check-cannot-see` | §3's two collisions |
| `retro-skill` → `recover-sprint` | `rfc-self-worktree-absence-is-not-completion` | §4's liveness-signal gap |
| `retro-spec` → `standards` | `rfc-0020` | §2's three document findings; `eaa.1#1`; `92u#2` |

Not routed, and why:

- **Tier calibration → `groom-beads`.** No RFC. Zero escalations means no
  direct routing evidence in either direction (§1), and a proposal built on
  the absence of a signal would be exactly the opinion the bar excludes.
  Recorded here so a future run can observe it twice.
- **`eaa.4`'s missing tier.** Folded into the `bead-bonanza` RFC's evidence
  rather than its own proposal: one instance, and the operating principle it
  violates already exists in the skill's text.

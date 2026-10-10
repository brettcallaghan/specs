# Run record — corpus listing (`spec list`)

*Written by `retro-w85b` on 2026-09-07. One document: the run's results, every
finding it captured, what it routed where, and what it could not establish.*

## Read this first — the one decision the run refused to make

**`w85b.8` is parked, `deferred`, labelled `human`, and it is the
only thing in this run that is waiting on a person.**

`spec list`'s YAML and JSON renderings disagree on a multi-line description, and
that disagreement is now visible in committed snapshot expectations — `w85b.9`
was deliberately scoped to make it visible without asserting which side is
right. The root cause is not a bug in the renderer. It is a **contradiction
inside the corpus**:

- `specs/standards/cli-output-standards.md`, YAML Style, is normative: *"A
  multi-line string SHALL be trimmed of leading and trailing whitespace and
  rendered as a block scalar."* `cli/cmd/yaml.go`'s `writeBlockScalar`
  implements exactly that, and `TestAMultiLineStringIsTrimmedAndRenderedAsABlockScalar`
  in `cli/cmd/yaml_test.go` has asserted it since before this run.
- `AC-SPECLIST-004.2` requires both renderings to carry **the same values**.
- `specs/features/spec-list/spec-list-frd.md`'s own preamble to
  FEAT-SPECLIST-004 says the requirement *"adds nothing"* and defers to the
  output standards — the standard that mandates the loss.

So the AC is, by its own requirement's stated terms, subordinate to the standard
that contradicts it. Two options, both specification changes, both costed on the
bead by `plan-w85b`:

| | change | cost |
|---|---|---|
| **A** | Amend `cli-output-standards`' YAML Style rule so a block scalar carries the value faithfully — pick the chomping indicator (`\|-` / `\|` / `\|+`) from the value's own trailing newlines, fall back to a quoted scalar where a leading indent would need an explicit indentation indicator | A cross-cutting standard with a written rationale changes; `renderYAML` changes for **all three** callers (`spec list`, `spec headings`, `spec refs`); one existing test and every snapshot artifact carrying a block scalar move with it |
| **B** | Narrow `AC-SPECLIST-004.2` in the FRD to "the same values *as the output standards render them*", making the trim specified behaviour | A shipped acceptance criterion is weakened, and a consumer diffing the two renderings — or round-tripping the YAML against source frontmatter — keeps seeing a mismatch, with the guide now saying that is intended |

A third option (give `spec list` its own rendering path) was **rejected** by
grooming rather than offered: it contradicts `[[container-spec-cli]]`'s Change
Guidance and `renderList`'s construction, which derives both renderings from one
struct declaration precisely so they cannot drift.

**To act on it:** say A or B on `w85b.8`, remove the `human`
label, and set its status back to `open`. It then re-enters the
gate and its Design and acceptance criteria are written against the chosen
option in one pass. Whichever way it goes, the demonstration is already in the
tree and the losing side's regeneration is a one-command diff.

## The run

| | |
|---|---|
| run epic | "Run: Corpus Listing (`spec list`)" |
| target | "Corpus Listing (`spec list`)", a pre-existing feature epic adopted whole; its six work orders kept their ids, edges and parent, and **no work bead was moved** |
| specification | `specs/features/spec-list/spec-list-frd.md` (approved FRD) and `specs/features/spec-list/spec-list-blueprint.md` (feature blueprint, retired by `kgir.6`) |
| sprint branch | the sprint branch |
| `base_sha` → head | `ad64088` → `861fa52`, **ten commits** |
| queue | 6 admitted at the gate (`manifest.admitted`), plus 4 review findings admitted mid-run across two loops. `manifest.excluded`, `manifest.moved` and `manifest.ledger` are all **unset**, not `[]` |
| dispatches | six work dispatches — `dev-w85b-1` … `dev-w85b-6` — plus `plan-w85b`, `review-w85b`, `coordinator-w85b`, `retro-w85b` |
| `concurrency_cap` | 20 |
| declared checks | 12 — `format`, `vet`, `build`, `test-go`, `skills`, `shared-refs`, `strata-docs`, `schemas`, `corpus`, `ref-links`, `template-drift`, `skills-strict` |
| capability tiers | 4 `standard`, 5 `light`, 1 `advanced`, resolved through `swarm.toml`'s `[models]`. **Which model actually served each dispatch is recorded nowhere** — `manifest.models` and `manifest.effort` are both unset |
| review passes | **3 of 3 — the budget is EXHAUSTED, not unspent.** The run stopped looping because it ran out of passes, not because review came back clean |
| amendments | **none.** The run epic carries zero `amend:` comments |
| circuit breaker | `manifest.consecutive_dispatch_failures` is unset; the breaker is not recorded as having armed |

**Standing authorization** (`manifest.standing_auth`) — **five rows, not
seven**, all at the skill default and none set in `swarm.toml`:

| row | value | source |
|---|---|---|
| `review_budget` | `2` (three passes) | default |
| `tier_escalation` | one tier, once, reported | default |
| `merge_to_default` | `false` | default |
| `move_foreign_beads` | `false` | default |
| `dispatch_failure_threshold` | `3` | default |

The two missing rows matter and one of them is load-bearing for this run. See
"What the coordinator's tooling did", defect 4.

## Results

**Ten beads closed, ten commits, one commit each. Zero blocked, zero skipped,
zero tier escalations.**

| bead | tier | agent | commit | what landed |
|---|---|---|---|---|
| `kgir.1` | standard | `dev-w85b-1` | `f59c0b4` | CORE — `Corpus.DocumentSummaries` in Document Core, with the parse-error envelope as pointers |
| `kgir.2` | standard | `dev-w85b-1` | `51ce4d1` | FILT — type and retired-status filters, the latter read from `LifecycleRules.RetiredStatuses` rather than from a spelling |
| `kgir.3` | standard | `dev-w85b-1` | `e76a727` | CMD — the `spec list` subcommand and seven snapshot cases |
| `kgir.4` | light | `dev-w85b-2` | `8bff463` | DOC — `description`'s second audience in the frontmatter standards, and the `spec list` section of the CLI reference |
| `kgir.5` | light | `dev-w85b-1` | `110a461` | COV — regenerated coverage figures, README diff only |
| `kgir.6` | advanced | `dev-w85b-3` | `2c8bbf3` | DONE — Definition of Done verified entry by entry, blueprint retired, `rfc-0027` filed |
| `w85b.10` | light | `dev-w85b-5` | `40760c6` | Corrected the inverted parse-error rationale in `spec-cli-reference` and its vendored copy |
| `w85b.7` | standard | `dev-w85b-4` | `d0c1647` | `spec list --type ""` no longer bypasses the unregistered-type rejection |
| `w85b.9` | light | `dev-w85b-4` | `383bd06` | Converted the AC-SPECLIST-001.3 fixture to a literal block scalar in all **seven** copies, making the YAML/JSON divergence visible |
| `w85b.15` | light | `dev-w85b-6` | `861fa52` | Fixed the panic `w85b.7`'s own fix introduced — `typeFilterValue.Type()` said `"stringSlice"` while `String()` returned `""`, so `GetStringSlice("type")` panicked |

- **No commit belongs to no bead**: `ad64088..861fa52` is exactly ten commits
  and ten beads.
- **No commit carries a trailer.** Measured:
  `git log --format='%H%n%(trailers)' ad64088..<sprint branch>`
  prints ten hashes and ten empty trailer blocks. The cause is measured below.
- **`manifest.admitted` names only the six `kgir` work orders.** The four
  review findings this run actually worked and closed (`w85b.7`, `.9`, `.10`,
  `.15`) were never added to it, so the manifest under-reports the run's own
  closure set by 40%.

### How the three review passes went

| pass | phase epic | filed | admitted into | outcome |
|---|---|---|---|---|
| 1 | `w85b.4` | `w85b.7` P1, `.8` P1, `.9` P2, `.10` P2, `.11` P3 | target `w85b.13` (`.7`, `.8`, `.9`, `.10`); `.11` held out | `.7`, `.9`, `.10` fixed and merged; **`.8` bounced to a human** |
| 2 | `w85b.14` | `w85b.15` P2, `.16` P3 | target `w85b.18` (`.15`); `.16` held out | `.15` fixed and merged. **`.15` is a panic that `w85b.7`'s own fix introduced** |
| 3 | `w85b.19` | `w85b.20` P2, `.21` P2 | neither — budget exhausted | **filed and unworked** |

**Pass 3 caught a defect introduced by pass 2's fix, and then the run ran out
of passes.** That sequence is the strongest argument in this record for the
budget being a real constraint rather than a formality: each pass found
something the previous pass's repair created or left, and the loop stopped on
the clock rather than on a clean bill.

**`w85b.20` and `w85b.21` are open, unworked, and nobody is assigned to them.**
That is a documented outcome of exhausting `review_budget`, not a failure:

- **`w85b.20`** (P2, `run-introduced`) — the doc comment `861fa52` added to
  `typeFilterValue.Type()` excuses the loss of cobra flag-name completion with a
  reachability claim that is **measurably false**: `__complete` ships and
  answers, and `DisableDefaultCmd` gates only the user-facing `completion`
  command. The code is correct; the comment argues for it with a false premise.
- **`w85b.21`** (P2, `run-introduced`) — `specs/containers/spec-cli/spec-cli-container.md`
  still specifies `--type` as a `StringSliceVar` built exactly like `check`'s
  `--only`. That was true when written and is false now: `w85b.7` replaced the
  `StringSliceVar` precisely *because* it drops an empty occurrence, and
  `w85b.15` then repaired that replacement. The container document specifies the
  mechanism the run removed.

Three further beads are `deferred` and unworked by design:

| bead | priority | why it sits |
|---|---|---|
| `w85b.8` | P1 | **the human decision above** |
| `w85b.11` | P3 | `withheldStatuses` reimplements the package's existing `stringSet`. The reviewer's own words: "Not a defect today" |
| `w85b.16` | P3 | `--type` silently lost CSV field unquoting when `w85b.7` replaced the `StringSliceVar`, and the doc comment does not say so |

## What the run decided

Twenty-one of the 39 findings are the six work-order kinds, and the run's
dominant shape is `decision` at ten — **the specification pinned the behaviour
and was silent about the shapes**. Four themes.

**1. The blueprint named fields and never named identifiers.** `kgir.1#1` is
the clearest instance: the CORE delta pins the summary's fields and their order
and never names the query or its Go signature, so the implementing agent chose
both and FILT and CMD inherited them — then `kgir.2#3` *changed* the signature
it had just published, which was the plan's intended sequencing rather than a
defect, but means `CORE`'s published API existed for exactly one commit.
`kgir.1#2` chose pointer-plus-`omitempty` over bare strings so an *absent*
description stays distinguishable from an *authored empty* one, a distinction
no document asked for and every consumer now depends on.

**2. Measurement beat inheritance, four times, and one of those was the
coordinator.** This is the run's most transferable result.

- `kgir.6#2` is a **near-miss worth more than most of the fixes**: the agent
  ran `jq keys` on the *first* record of a `bd export` JSONL, saw no `comments`
  key, concluded the documented harvest path was broken, and began sweeping the
  tree to "correct" `specs/standards/findings-capture-protocol.md` and its three
  vendored copies. `jq -r 'keys[]' | sort -u` across **all** records shows
  `comments` is present — JSONL records are heterogeneous. Had the sweep landed,
  it would have broken the documented harvest for every skill that vendors that
  standard. Caught only because AGENTS.md requires the *replacement* to be
  measured too.
- `w85b.15#2` and `w85b.15#3` corrected two claims in a bead's own Design
  before they shipped into a committed doc comment: pflag's zero-value set for a
  custom `Value` is `{"", "0", "false", "<nil>"}` (the Design listed `""` twice
  and omitted `"<nil>"`), and cobra's repeatability test has **four** disjuncts
  against `cobra`'s own reduced `SliceValue` interface, not two against
  `pflag`'s. Neither correction changed the decision; both changed what shipped
  in the comment.
- `w85b.20#1` is the **coordinator admitting it propagated a false claim into a
  dispatch prompt** — "root.go sets `DisableDefaultCmd`, so the CLI ships no way
  to generate the scripts that call `__complete`" — inherited from a grooming
  report, stated as settled fact, used as the reason a trade-off was acceptable.
  Review pass 3 measured it false. It reached two agents first. This is exactly
  the propagation AGENTS.md's own table calls the **worst surface**, and it
  happened in the run whose repository owns that table.

**3. Two verification surfaces exist and one standard knows about one of
them.** `kgir.3#2` and `kgir.6#1` are the same defect from opposite ends. A
snapshot-proved AC has two citation sites — the case's `cites` key and the shared
`// AC:` comment on `TestSnapshotCases` — and neither document says which is
authoritative, so the implementer wrote both and the shared comment is now **34
identifiers on one line**, a citation that binds to no specific test.
`kgir.6#1` then measured the consequence: the DONE item prescribes
`grep -rn "AC: AC-SPECLIST" cli/` to confirm coverage, and that command reports
**11 of 21** non-exempt criteria; the other ten live in seven `case.toml`
`cites` arrays. An agent following the item verbatim would have reported ten
uncited criteria and either blocked the feature or "fixed" it with redundant
comments. Both are already argued in
`rfcs/inbox/standards/rfc-0027-two-standards-one-citation-surface.md` (P1).

**4. A gate the run's own beads could not see.** `w85b.7#3`, and independently
`kgir.5#2` and `w85b.15#4`. **Three separate dev agents each discovered on
their own that `AGENTS.md` makes `scripts/coverage.sh` a completion condition
for any `cli/**` change** — while no bead's acceptance criteria and no declared
check mentions coverage, and the run's own `checks_baseline` does not either.
`w85b.7` measured its tree stale (`scripts/coverage.sh --check` → exit 1) with
the declared gate fully green. Every one of the three regenerated the README
and folded the diff into its own commit, which is the right answer arrived at
three times by three agents from a document none of them was pointed at. This
is the run's cheapest available improvement and it is **candidate A** below.

`kgir.5#2` adds the second, sharper half: `scripts/coverage.sh` resolves its
root from **its own checkout**, which under dispatch is a per-dispatch worktree
that does not carry a sibling's merged work. It did not bite this run only
because the one sibling commit outstanding touched no `cli/**` path — measured,
not assumed.

**No `sequencing` findings.** With six dispatches over a partitioned queue,
nothing inherited a broken build. One `unverifiable` (`kgir.4#1`): the DOC item
documented a `spec list` surface that did not exist in the branch it was written
on, by design, and nothing re-checked the guide against a binary afterwards —
until `kgir.6#4` did it at retirement, and found it accurate.

## Scope

**The run finished its admitted set, and its scope was narrowed twice — once
by the review budget and once by a human bounce.** Both must be said plainly:

- **`review_budget` exhausted.** `manifest.review_passes` = 3 of 3.
  `w85b.20` and `w85b.21` are the pass-3 findings, filed and unworked. There is
  no fourth pass, so nothing in this run has looked at the tree as it now stands.
- **`w85b.8` bounced.** The only bead the run refused to decide; see the top of
  this document.
- Three P3 beads deferred: `w85b.11`, `w85b.16`, and one more
  (a third copy of the inverted parse-error clause, in the now-retired
  blueprint, filed by `w85b.10` rather than fixed inline).

**Nothing was excluded at the gate and nothing was moved** — but that is read
off the epic's `admitted` list and the absence of any `amend:` comment, not off
`manifest.excluded` or `manifest.moved`, which are unset rather than `[]`.

Four beads were filed against **other** parts of the tree and deliberately not
worked here: one P2 (open — `spec check --only ""` carries the
same pflag defect `w85b.7` fixed, but `--only` has a different behavioural
contract and the fix is not a mechanical copy), one P3 (open —
vendored `shared/` references cite siblings with a `shared/` prefix that
resolves one directory too deep), the third inverted-clause bead and `w85b.16`.

## What the coordinator's tooling did

Six items. Two of them — the `bd dolt pull` deadlock and the exclusive tracker
lock — are also recorded in the finding channel as `environment` findings, and
this record does **not** re-measure those two: doing so means freezing the
container. The other four are **not** `retro:` findings on any bead, so the
harvest query cannot see them; they reached this record only because the
coordinator hand-carried them in a dispatch prompt, and each of those four was
re-measured here before being written down. That gap is itself the recurrence of
an already-filed issue; see Routing.

**1. `bd dolt pull` deadlocks the container, and `bd`'s own error message
recommends it.** The trap is a chain: `bb-close-bead.sh` runs `bd dolt push`
after each close and, on failure, prints *"push it before the container is
reclaimed"* (`bb-close-bead.sh:271`). A hand-run `bd dolt push` that lost a race
to a sibling fails non-fast-forward with `hint: Integrate the remote changes
(e.g. 'dolt pull ...') before pushing again`. Following that hint takes the
shared embedded-Dolt write lock, never returns, **ignores SIGTERM**, and starves
every other `bd` call in the container. Two dispatches hit it independently
before a binding correction was issued; each freeze cost roughly twenty minutes
and ended in `kill -9`.

**2. The Dolt lock is exclusive, so `concurrency_cap: 20` buys less than it
says.** What is *measured* is narrower than the headline and should be stated as
such: during one `bd dolt pull`, three sibling `bd` calls sat at **zero CPU**
for 337s, 211s and 209s (`bd comment`, `bd list --parent`, `bd show --json`),
and all three completed the instant the pull was killed. That establishes an
exclusive writer, not that every ordinary `bd` call serializes against every
other. See "What this record could NOT verify".

**3. `groom-beads`' `scripts/bd-health.sh` mis-enumerates epics, and its exit
status cannot gate anything.** Re-measured for this record on the current
database:

```
bd list --type epic --json                                   → 10 epic records
sed -n 's/^[[:space:]]*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'   → 36 ids
```

26 of the 36 are not epics — they are bead ids scooped out of nested
`metadata.moved[]` and `metadata.excluded[]` arrays by a line-oriented `sed`.
Feed those to `bd swarm validate`, which accepts only epics, and the script
exits 1 repo-wide. **Already filed twice**, both P1 and open — so nothing new goes on the pile; the figures
above are recorded because AGENTS.md says to record the command for a volatile
number, not the number.

**4. `bb-close-bead.sh --trailers-from` cannot succeed, and two of the seven
standing-authorization rows have no manifest key at all.** Measured:

```
swarm manifest require --run <run epic> --key trailers
  → trailers: ok: false — "trailers is not set"        (exit 1)

swarm manifest get --run <run epic> --key admit_review_findings --key pull_request
  → "admit_review_findings, pull_request: not a declared manifest key"   (exit 3)
```

The first is why no commit in this run carries a trailer — already filed
upstream as `.sf/outbox/swarm/2026-09-05-a-run-resolves-no-trailers-so-no-commit-carries-one.md`,
now at recurrence 1. **The second is new and it is not cosmetic**: this run
demonstrably *did* admit review findings and work them, and the key that
authorizes that has nowhere to be recorded. A resumed session reading this
manifest cannot tell whether admitting `w85b.7` was authorized or improvised.
Candidate **C**.

**5. `swarm-review` prescribes a flag `bd` does not have.** `swarm-review`'s
SKILL.md line 112 tells the reviewer to file findings with
`bd create ... --discovered-from <origin-bead-id>`, and line 115 explains why
the flag matters. Measured: `bd create --help` lists no such flag; the string
`discovered-from` appears in its help exactly once, inside the description of
`--deps` (`--deps 'discovered-from:bd-20'`). The reviewer evidently worked
around it — every filed bead carries a correct `discovered-from` edge — but the
instruction as written cannot run. The identical defect in **`swarm-retro`** was
filed from the previous run
(`.sf/outbox/swarm/2026-09-05-retro-filing-snippet-uses-a-flag-bd-does-not-have.md`)
and `swarm-retro`'s SKILL.md line 202 still carries it, so this is the same
defect on a second surface and the first one is unfixed. Candidate **D**.

**6. `bd dep add --file` on an empty file is an error, not a no-op.** Measured
here: `bd dep add --file <empty>` → exit 1, `Error: no dependency edges found`.
Recorded so a partition step that legitimately produces zero edges is not read
as a failure.

## Every finding, verbatim

**39 findings across 11 beads.** Harvested with one query over the run's
closure set:

```bash
bd export | jq -r 'select(.id | startswith("<run epic>")
                        or startswith("<target epic>"))
  | .id as $i | (.comments // [])[] | select(.text | startswith("retro:"))'
```

`bd show <id> --json` carries `comment_count` and **no** `comments` array, so it
cannot be used for this; `bd export`'s JSONL records are heterogeneous, so a
`jq keys` on one line describes that record and not the schema. Both traps were
hit inside this run and are recorded as findings below (`kgir.4#5`, `kgir.6#2`).

Each finding is quoted whole and unedited. Ids are derived — `<bead>#<ordinal>`,
the ordinal being the finding's position among that bead's own `retro:` records.

### `decision` — 10

**`kgir.1#1`** — recorded by `dev-w85b-1`

> retro:decision The blueprint and the bead pin the summary's FIELDS and their ORDER but never name the query or its Go shape, so this item chose both and the FILT and CMD items inherit them. Chosen: `func (c *Corpus) DocumentSummaries() []DocumentSummary`, with `DocumentSummary` declared in `cli/core/corpus/query.go`. Rejected: `Summaries()` (reads ambiguously at package scope beside `Features`/`Criterion`) and taking an options struct now (FILT owns filters; a no-field options struct would be speculative generality). FILT will therefore CHANGE this signature rather than add to it.
> evidence: specs/features/spec-list/spec-list-blueprint.md, CORE delta Behavior list, says only "The query SHALL return one summary per corpus member" and its Anchors line names `cli/core/corpus/query.go` — no identifier. The CMD delta fixes the ANSWER key ("a mapping carrying one key, `documents`") but not the query name. specs/containers/spec-cli/spec-cli-container.md:184-185 names every other query by identifier (`corpus.Corpus.Feature`/`Features`/`FeaturesByCode`, `corpus.Corpus.Criterion`/`CriteriaOf`/`CriteriaByTag`), so the container doc will need this one added when CMD lands.

**`kgir.1#2`** — recorded by `dev-w85b-1`

> retro:decision The Design pins that a parse-error member carries NO slug/type/status/title/description and that null is the rejected alternative, but not how absence is represented in Go. Chosen: the five envelope fields are POINTERS (`*specmodel.Slug`, `*DocumentType`, `*Status`, `*string`, `*string`) with `omitempty`. Rejected: bare strings with `omitempty`, which is one character cheaper and cannot tell an ABSENT description from an AUTHORED EMPTY one — both render as an omitted key, so an empty `description:` in a document would silently read as "the index could not parse this". Cost accepted: FILT and CMD must dereference, and the type filter reads `summary.Type == nil` for the parse-error exclusion its Behavior list already requires.
> evidence: cli/cmd/yaml.go `isEmptyValue` mirrors encoding/json's omitempty test — `case reflect.Array, reflect.Map, reflect.Slice, reflect.String: return v.Len() == 0` — so a bare "" string field is omitted from BOTH renderings, indistinguishably from an absent one. `case reflect.Pointer, reflect.Interface: return v.IsNil()` is what makes the pointer form carry the distinction.

**`kgir.2#3`** — recorded by `dev-w85b-1`

> retro:decision This item CHANGED the CORE item's query signature rather than adding beside it: `DocumentSummaries()` became `DocumentSummaries(SummaryOptions) ([]DocumentSummary, error)`. Rejected: keeping the no-arg form and adding a second filtered method, which would have left two entry points where the blueprint describes one query, and would have made the caller choose between them. The consequence is that the CORE item's published API existed for exactly one commit (918263e) and no caller outside the tests ever saw it — which is the plan's intended sequencing (CORE -> FILT -> CMD, single-pr delivery), not a defect, but it does mean CORE's signature should not be cited anywhere as shipped.
> evidence: the CORE item's OUT OF SCOPE said "Filters of any kind (FILT owns them)", and this item's Description says it "EXTENDS the query the CORE item adds". The blueprint (specs/features/spec-list/spec-list-blueprint.md, FILT delta) states the filters as behavior of "the query", singular, throughout: "the query SHALL return only summaries whose type is among them". The error return is required by the CMD delta, which says "WHERE the query reports an unregistered type, the subcommand SHALL write that report to standard error and SHALL exit non-zero" — so the report had to be a value the caller receives, not a finding.

**`kgir.3#1`** — recorded by `dev-w85b-1`

> retro:decision The item lists AC-SPECLIST-004.3 ("two runs over an unchanged corpus emit byte-identical output") as `e2e`, but the snapshot harness structurally CANNOT prove it — it replays each case once and compares that one run against a committed artifact, so a second run is not expressible as a case. Proved instead by a Go test in cli/cmd/list_test.go, which is where cli/cmd/query_test.go already puts the same shape of claim. Rejected: adding a duplicate snapshot case with the same expected artifact, which would prove two SEPARATE runs from a clean tree agree, not that a warm run agrees with the cold one that persisted the index.
> evidence: cli/cmd/snapshot_test.go's `runSnapshotCaseWith` materializes `input/` into a fresh `t.TempDir()` and calls `cmd.Execute` exactly once per case. cli/cmd/query_test.go states the same limit for AC-GHK-007.6: "the harness in cli/cmd/snapshot_test.go can only compare what one run produces... only a Go test can". The new test asserts the two runs are byte-identical AND that the first persisted `docs/specs.json`, so the comparison is cold-run against warm-run rather than two cold runs — the paths differ (`openCorpus` builds and writes, then `corpus.Open` serves warm), and an order taken from a map rather than from the index would pass every snapshot case and fail this.

**`kgir.4#2`** — recorded by `dev-w85b-2`

> retro:decision The item forbids ADDING a length limit but says nothing about the length guidance already in the table row it names as an anchor, which reads `1-2 sentences`. Leaving it silent would have shipped a document that tells an author to write for triage while a table cell still reads as a cap. I decided to keep the existing "1-2 sentences" cell untouched and add one sentence stating it is a shape rather than an enforced limit — nothing truncates a description and no check measures one. Rejected alternative: deleting the cell, which would have been a change to the required-properties table affecting all 59 documents and outside the approved scope.
> evidence: item Design says "DECISION ALREADY MADE — do NOT add a length limit to description" and the AC reads "No length limit has been added to description in any document"; the anchor names the row "Brief summary of the document | 1-2 sentences", so the pre-existing guidance is inside the edited region. FEAT-SPECLIST-001 in specs/features/spec-list/spec-list-frd.md supplies the wording basis: "Nothing bounds a description's length: the schema constrains it to a string and the frontmatter standards ask for one to two sentences in prose, which no check enforces." Shipped sentence: "The one-to-two-sentence guidance above is a shape, not an enforced limit."

**`kgir.6#3`** — recorded by `dev-w85b-3`

> retro:decision The DoD entry "Every AC in the coverage table is verified at
> its DECLARED LEVEL" is not literally true for AC-SPECLIST-004.3, and I
> ticked it with a recorded deviation rather than leaving it unticked or
> silently passing. The blueprint declares 004.3 e2e; the snapshot harness
> replays each case once (cli/cmd/snapshot_test.go materializes a fresh
> t.TempDir() at line 124 and calls cmd.Execute once at line 148), so it
> cannot express "two runs agree" and the criterion is proved by a Go test
> in cli/cmd/list_test.go instead.
> Chosen: tick the box and add a "Verification note (recorded at
> retirement)" to the blueprint naming the deviation, its reason and the
> RFC proposing the fix. Rejected: (a) leaving the box unticked, which on a
> RETIRED blueprint reads as an unfinished feature and loses the reason
> forever, and (b) treating it as a gap to report and stop on, which would
> have blocked a feature whose criterion IS verified -- the level obligation
> in specs/standards/ac-traceability-standards.md is a SHOULD ("A test
> citing an acceptance criterion SHOULD verify it at the level declared in
> the feature blueprint AC Coverage table"), not a SHALL.
> evidence: this is the SECOND independent occurrence of the same harness
> limit. cli/cmd/query_test.go already states it for AC-GHK-007.6 -- "the
> harness in cli/cmd/snapshot_test.go can only compare what one run
> produces... only a Go test can" -- and the kgir.3 implementer rediscovered
> it from scratch. Nothing in specs/standards/testing-standards.md
> Snapshots section says what a case CANNOT assert. Filed as P2 of
> rfcs/inbox/standards/rfc-0027-two-standards-one-citation-surface.md.
> Adjacent but distinct from rfcs/inbox/standards/rfc-0011 P3, which covers
> a level unreachable because the mechanism is UNBUILT; here the mechanism
> works and the harness cannot express the assertion.

**`w85b.15#4`** — recorded by `dev-w85b-6`

> retro:decision Regenerated README coverage figures and folded them into this bead's commit, though no acceptance criterion asked for it.
> evidence: The bead's last AC names only "format, vet, build, test-go, corpus,
> ref-links", and bb-checks.py carries no coverage leg by design (AGENTS.md:
> "three instrumented test legs is not a price worth paying per bead").
>
> But AGENTS.md also says: "A change under cli/** is therefore not finished until
> that script has been run from the repository root and the README diff committed
> with it", and cli.yml's coverage job runs scripts/coverage.sh --check and fails
> when the committed figures do not match a fresh measurement. Adding a test file
> to cli/cmd moves those figures, so leaving them would have handed the run a red
> CI job that no bead owns.
>
> Run practice in this run agrees - both prior cli/-touching beads committed the
> README diff:
>   $ git log --format=%h ad64088..[sprint branch] -- cli/   (then per commit, --stat -- README.md)
>   383bd06 TOUCHED README   (Snapshot cases 2,594 -> 2,608)
>   d0c1647 TOUCHED README
>
> Measured here (toolchain matches cli/go.mod: go 1.24 vs installed go1.24.7):
>   $ ./scripts/coverage.sh        rc=0
>   Unit tests            2,665 -> 2,666 / 3,405   (78.3%, unchanged to 1dp)
>   cli/cmd               1,027 -> 1,028 / 1,201   85.5% -> 85.6%
>   All layers combined   3,055 -> 3,056 / 3,405   89.7% -> 89.8%  (badge too)
>   $ ./scripts/coverage.sh --check   rc=0
>
> The one new covered statement is typeFilterValue.Type()'s body, which the new
> test reaches and nothing did before.
>
> Alternative rejected: leave README alone and file a follow-up bead. That defers
> a guaranteed CI failure onto a later bead for no benefit, and splits one
> mechanical regeneration away from the change that caused it.

**`w85b.7#2`** — recorded by `dev-w85b-4`

> retro:decision typeFilterValue.String() returns "" for an unset flag rather than "[]", so that `spec list --help` reads exactly as before. Rejected the "[]" form, which adds a spurious `(default [])`.
>
> evidence: measured both, by building ./spec from cli/ with each form and grepping the
> help line:
>
>   String() returning "[]" when empty:
>     $ spec list --help | grep -- --type
>           --type strings   report only members of the named document types; repeatable
>                            or comma-separated (default [])
>
>   String() returning "" when empty (as shipped):
>     $ spec list --help | grep -- --type
>           --type strings   report only members of the named document types; repeatable
>                            or comma-separated
>
>   and against the pre-fix binary, byte-for-byte:
>     $ diff /tmp/help-before.txt /tmp/help-after.txt   -> no output
>
> pflag decides whether to print `(default ...)` from Flag.defaultIsZeroValue(), which
> switches on the CONCRETE Value type; a custom pflag.Value falls to the default branch,
> where only "false", "<nil>", "" and "0" read as a zero value. "[]" is recognised only
> for pflag's own *stringSliceValue, which this deliberately is not.
>
> Type() returns "stringSlice" for the same reason in the other direction: pflag's
> UnquoteUsage maps that name to the `strings` placeholder, so the usage line still reads
> `--type strings`. Measured above — the help output is byte-identical to the pre-fix
> binary.
>
> The cost accepted: an unset custom flag renders as "" rather than as "[]", which is a
> cosmetic inconsistency with pflag's own slice flags visible nowhere but a debugger.

**`w85b.9#2`** — recorded by `dev-w85b-4`

> retro:decision The Design's literal replacement block RE-WRAPS the description, which its own Trap and the acceptance criteria forbid. Followed the AC (kept the existing line breaks, changed only the scalar style); rejected the Design's block verbatim.
>
> evidence: the Design proposes
>
>   description: |
>     The longest description in this corpus, written to prove that a summary
>     reports a description exactly as it was authored — neither truncated at
>     some column an implementation chose, nor elided with an ellipsis, nor
>     folded onto one line when the author wrapped it across several.
>
> but the file as committed breaks the lines differently:
>
>   $ sed -n 3,6p cli/testdata/cases/list-all/input/docs/frds/frd-checkout.md
>   description: The longest description in this corpus, written to prove that a
>     summary reports a description exactly as it was authored — neither truncated
>     at some column an implementation chose, nor elided with an ellipsis, nor
>     folded onto one line when the author wrapped it across several.
>
> ("...that a" / "summary reports..." versus "...that a summary" / "reports...").
>
> The Trap says "Do not re-wrap or re-word the description text while converting it ...
> only the scalar style changes", and the second acceptance criterion says "no
> re-wording, no re-wrapping". Under a PLAIN scalar the wrapping is invisible — YAML folds
> it — so the two texts were interchangeable when the Design was written. Under a LITERAL
> block scalar the line breaks become part of the value, which is the whole point of the
> bead, so the choice is now load-bearing and the Design and the AC disagree.
>
> Resolved in favour of the AC, which the dispatch names as authoritative. The shipped
> diff is a one-line insertion:
>
>   -description: The longest description in this corpus, written to prove that a
>   +description: |
>   +  The longest description in this corpus, written to prove that a
>      summary reports a description exactly as it was authored — neither truncated
>      ...
>
> Cost of the alternative: taking the Design's block verbatim would have re-wrapped three
> of the four lines, and since the breaks now survive into the rendered output, a reader
> diffing the regenerated expectations could not tell a deliberate re-wrap from a
> transcription error in the fixture's evidence text.

**`w85b.9#3`** — recorded by `dev-w85b-4`

> retro:decision Converted SEVEN copies of frd-checkout.md, not the five the bead lists, to keep the duplicated fixture corpus byte-identical as the Design itself requires.
>
> evidence: the Design says "Keep the five copies byte-identical to each other. They are
> duplicated on purpose and a per-case divergence is a defect in its own right." Measured,
> there are seven identical copies under list-*, not five:
>
>   $ find cli/testdata/cases -path "*list-*" -name frd-checkout.md | xargs md5sum
>   a58097242f96bb0ec39b5fa97bb81b84  .../list-all-statuses/input/docs/frds/frd-checkout.md
>   a58097242f96bb0ec39b5fa97bb81b84  .../list-all/input/docs/frds/frd-checkout.md
>   a58097242f96bb0ec39b5fa97bb81b84  .../list-json/input/docs/frds/frd-checkout.md
>   a58097242f96bb0ec39b5fa97bb81b84  .../list-type-comma/input/docs/frds/frd-checkout.md
>   a58097242f96bb0ec39b5fa97bb81b84  .../list-type-empty/input/docs/frds/frd-checkout.md
>   a58097242f96bb0ec39b5fa97bb81b84  .../list-type-repeated/input/docs/frds/frd-checkout.md
>   a58097242f96bb0ec39b5fa97bb81b84  .../list-type-unregistered/input/docs/frds/frd-checkout.md
>
> after the conversion, all seven again share one hash:
>
>   $ ... | xargs md5sum | awk "{print \$1}" | sort -u
>   46019469ab1771ff8cbed7a5e24c994b
>
> The two the bead does not name are the error cases: list-type-unregistered predates this
> run, and list-type-empty was added an hour earlier by w85b.7 in the same
> batch, which is why the bead's Affected-files list could not have named it. Neither
> emits a listing, so neither has an expected artifact that moved — `git status` shows
> their input changed and their expected/ untouched.
>
> Rejected: converting only the five named copies. It satisfies the first acceptance
> criterion literally (those five are byte-identical to each other) while creating exactly
> the per-case divergence the Design calls a defect, and it would leave two copies of the
> "prove a description is reported as authored" fixture whose description is a plain
> scalar and therefore cannot.

### `gap` — 5

**`kgir.2#2`** — recorded by `dev-w85b-1`

> retro:gap The acceptance criterion "the string \"retired\" appears nowhere in the filter's own logic" is ALREADY enforced by a pre-existing test that neither this item nor the blueprint mentions, and an implementer who did not go looking would have written a second, redundant source scan to satisfy it. Naming the existing guard in the item would have saved that.
> evidence: `TestTheExtensionPointNamesNoRegisteredVocabulary` in cli/core/corpus/profile_test.go walks every non-`_test.go` Go file under the package, parses it, and errors on any string literal matching a value the profile registers — statuses included, read from the generated schema so the check cannot go stale. Its failure line under the negative control above was `profile_test.go:197: query.go:301 names the status "retired"`. The item's ANCHORS section names cli/core/corpus/query.go and cli/core/corpus/profile.go and stops there; the blueprint's FILT delta Anchors name the same two files. Neither names profile_test.go.

**`kgir.4#3`** — recorded by `dev-w85b-2`

> retro:gap Neither the item nor the blueprint says which citation form to use when a corpus document that is ALSO vendored into a skill needs to cite a second such document, and the two homes disagree about what resolves. specs/guides/spec-cli-reference.md uses backticked `shared/<name>.md` paths; specs/standards/frontmatter-standards.md uses [[slug]]. I followed each file's own existing convention rather than unifying them, and filed the unresolvable-in-both-homes problem as a bead (discovered-from this item) rather than settling it inline.
> evidence: measured — `plugin/skills/spec/references/shared/shared/frontmatter-standards.md` MISSING while `plugin/skills/spec/references/shared/frontmatter-standards.md` EXISTS, i.e. the `shared/` prefix is one directory too deep from the vendored file's own location; the same is true of the two pre-existing `shared/document-readiness-checklist.md` citations at lines 253 and 279 of the vendored copy. From the SOURCE at specs/guides/, neither `shared/x.md` nor a bare `x.md` resolves, because the targets live in specs/standards/. No check covers it: scripts/check-skills.sh skips vendored copies with four separate `if "/references/shared/" in p: continue` guards. AGENTS.md's "Citing another skill's reference" section governs skill-to-skill citation and does not reach this case.

**`kgir.5#2`** — recorded by `dev-w85b-1`

> retro:gap The item's Design says to run the script "from the REPOSITORY ROOT, not from cli/" and pins the Go version, but says nothing about WHICH tree — and under swarm dispatch the repository root you are standing in is a per-dispatch worktree that does not carry any sibling's merged work. Nothing in the item, the blueprint or the script warns of that, and the figures it writes are wrong without a word of warning if a sibling has already landed a cli/** change on the sprint branch.
> evidence: scripts/coverage.sh resolves `REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"` — the root of whatever checkout the script itself lives in, which in a dispatch is /home/user/.bb-worktrees/<dispatch-id>. The item's Design names the Go-version mismatch as "the one non-obvious failure mode here"; the stale-tree mode is a second one with the same symptom (CI's `--check` fails for a reason unrelated to the feature) and no stated remedy. It did not bite this run only because the sibling's single merged commit touched no cli/** path.

**`kgir.6#1`** — recorded by `dev-w85b-3`

> retro:gap The item prescribes `grep -rn "AC: AC-SPECLIST" cli/` as the way to
> confirm every non-@manual AC is cited, and that command reports HALF the
> coverage that exists. Citations live on two surfaces in this repo and the
> item names only one, so an agent following it verbatim would have reported
> 10 uncited criteria and either blocked the feature or "fixed" it by adding
> redundant comment citations.
> evidence: measured from the repository root, worktree bb/dev-w85b-3.
>   (cd cli && go build -o ../build/spec ./spec) && for f in 001 002 003 004 005;
>     do build/spec criterion --of "FEAT-SPECLIST-$f"; done
> -> 23 criteria. build/spec criterion --of FEAT-SPECLIST-005 --json shows
> 005.1 and 005.2 both carry "manual":true, so 21 are non-exempt.
>   grep -rn "AC: AC-SPECLIST" cli/
> -> 10 lines covering 11 criteria: 001.1, 001.4, 001.5, 001.6, 002.1, 002.3,
> 002.5, 003.1, 003.3, 003.4, 004.3.
>   grep -rn "AC-SPECLIST" cli/testdata/
> -> the other 10 in seven case.toml `cites` arrays: 001.2, 001.3, 001.7,
> 002.2, 002.4, 003.2, 004.1, 004.2, 004.4, 004.5. Union = all 21.
> Root cause is a corpus-level one, not the item: specs/standards/ac-
> traceability-standards.md (v0.4.0, active) defines exactly one citation
> form ("a citation comment SHALL be placed on the line(s) immediately
> preceding the test declaration") and never mentions a manifest -- measured,
> grep -nic "manifest|case.toml|data-driven" on it -> 0 -- while specs/
> standards/testing-standards.md (v0.1, approved) sanctions the manifest
> form twice ("The manifest exists precisely so the citation lives outside
> the fixture"). The cross-reference is one-way: testing-standards names
> ac-traceability 3 times, ac-traceability names testing-standards 0 times.
> Filed as P1 of rfcs/inbox/standards/rfc-0027-two-standards-one-citation-
> surface.md.

**`w85b.9#4`** — recorded by `dev-w85b-4`

> retro:gap The YAML/JSON divergence this bead exists to expose is now VISIBLE in committed snapshots, as intended. Recorded, not fixed — the decision belongs to the parked sibling bead.
>
> evidence: the same description now renders two different values in the two artifacts of
> the same command.
>
>   cli/testdata/cases/list-all/expected/checks.log (YAML), four lines, no trailing blank:
>       description: |-
>         The longest description in this corpus, written to prove that a
>         summary reports a description exactly as it was authored — neither truncated
>         at some column an implementation chose, nor elided with an ellipsis, nor
>         folded onto one line when the author wrapped it across several.
>
>   cli/testdata/cases/list-json/expected/checks.json (JSON), one string WITH a trailing
>   newline:
>       "description":"The longest description in this corpus, written to prove that a\n
>       summary reports a description exactly as it was authored — neither truncated\nat
>       some column an implementation chose, nor elided with an ellipsis, nor\nfolded onto
>       one line when the author wrapped it across several.\n"
>       (transcribed with line breaks for readability; the artifact is one line and the
>       value ends "...several.\\n")
>
> Mechanism: cli/cmd/yaml.go:207-208, writeBlockScalar writes `|-` and then splits
> `strings.TrimSpace(s)`, so the trailing newline a literal block scalar carries is
> dropped; encoding/json marshals the value as received and keeps it.
>
> Left exactly as it is, per this bead's scope boundary and the dispatch: no case or
> assertion compares the two renderings. Whether the trim is a defect in the renderer or
> specified behaviour is an unresolved conflict between
> specs/standards/cli-output-standards.md's YAML Style rule ("A multi-line string SHALL be
> trimmed of leading and trailing whitespace and rendered as a block scalar") and
> AC-SPECLIST-004.2, parked on a sibling bead awaiting a human. cli/cmd/yaml.go is
> unmodified by this bead.
>
> The gap, for whoever picks that decision up: the two artifacts now DISAGREE in a
> committed snapshot, so whichever way the human decides, the demonstration is already in
> the tree and the losing side's regeneration is a one-command diff.

### `gap:project` — 1

**`w85b.7#3`** — recorded by `dev-w85b-4`

> retro:gap:project The bead's checks criterion and the run's checks_baseline both omit coverage, but AGENTS.md makes a README coverage regeneration a completion condition for any cli/** change. Six new statements silently staled the committed figures; the declared gate stayed green.
>
> evidence: the bead's last acceptance criterion reads "The project's declared checks
> pass: format, vet, build, test-go, corpus, ref-links". All of them passed on the fixed
> tree:
>
>   $ /root/.claude/skills/swarm/scripts/bb-checks.py ; echo $?
>   0            (output differs from the run's checks_baseline only in test timings)
>
> But AGENTS.md, under "## Checks", says: "Coverage belongs to none of those lists ... A
> change under `cli/**` is therefore not finished until that script has been run from the
> repository root and the README diff committed with it. It is in `ci.sh` and
> deliberately not in `sf.toml`: three instrumented test legs is not a price worth paying
> per bead."
>
> Run by hand, it was stale:
>
>   $ bash scripts/coverage.sh --check ; echo $?
>   error: .../README.md does not match the measured coverage.
>          regenerate it with scripts/coverage.sh and commit the result.
>   1
>
>   committed: | **Total** | **3,399** | **3,051** | **89.8%** |  ; `cli/cmd` 1,195
>   measured:  | **Total** | **3,405** | **3,055** | **89.7%** |  ; `cli/cmd` 1,201
>
> The +6 statements are exactly typeFilterValue's three methods, so the staleness is this
> bead's, not pre-existing. Regenerated with `bash scripts/coverage.sh` and the README
> diff is in this bead's commit; the re-run gate is green.
>
> The friction, and why it is the project's rather than the skill's: a dispatched agent
> reads the bead's AC and the run's checks_baseline, and BOTH say the gate is
> format/vet/build/test-go/corpus/ref-links. Nothing in either surface mentions coverage,
> so an agent that does not independently read AGENTS.md's "## Checks" section lands a
> cli/** change that fails cli.yml's coverage job. The trade-off AGENTS.md documents
> (too slow to pay per bead) is deliberate and worth keeping; what is missing is a
> pointer at the surfaces an agent actually reads.
>
> Cheapest fix candidates, in the project's own tree: (a) grooming adds "and
> scripts/coverage.sh --check passes" to the declared-checks AC of any bead whose
> c4_changes touch the CLI; or (b) the checks_baseline the run records notes that
> coverage is out of the declared set and is a completion condition for cli/** changes.
> Not filed as a bead because it is a change to how beads are groomed and how a run
> records its baseline, not a code defect.
>
> Timing, for anyone weighing (a): `scripts/coverage.sh` took under three minutes on this
> container, measured as part of this bead — the three instrumented legs reported
> cli/profile 0.220s, cli/cmd 1.078s, cli/spec 2.116s plus the unit leg.

### `stale` — 2

**`w85b.15#2`** — recorded by `dev-w85b-6`

> retro:stale The bead's Design mis-stated pflag's defaultIsZeroValue zero set; I measured it before restating it in a committed doc comment.
> evidence: The Design's Rejected-alternatives bullet reads: 'defaultIsZeroValue
> (pflag@v1.0.9 flag.go) treats only "", "0", "false" and "" as a zero default for
> a custom Value.' That set lists "" twice and omits "<nil>".
>
> Measured, pflag@v1.0.9 (go list -m -f '{{.Dir}}' github.com/spf13/pflag), the
> default: branch - the only branch a custom pflag.Value such as typeFilterValue
> reaches, since it matches none of the concrete type cases above it:
>
>   default:
>       switch f.DefValue {
>       case "false":  return true
>       case "<nil>":  return true
>       case "":       return true
>       case "0":      return true
>       }
>       return false
>
> So for a custom Value the zero set is {"", "0", "false", "<nil>"}. "[]" is zero
> only via the earlier case for *intSliceValue, *stringSliceValue, *stringArrayValue,
> which typeFilterValue is not. The Design's CONCLUSION is unaffected and its
> measurement stands - returning "[]" from String() would print "(default [])" -
> but the stated reason was wrong in its details.
>
> The corrected wording is what shipped in String()'s doc comment. This is a
> correction of an inherited claim, so it was measured rather than reasoned:
> the burden applies to the replacement, not only to the sentence removed.

**`w85b.15#3`** — recorded by `dev-w85b-6`

> retro:stale The bead's cobra completion claim was two disjuncts short and named the wrong package's interface.
> evidence: The Design states cobra 'decides a flag may be repeated with
> flag.Value.(SliceValue) OR strings.Contains(flag.Value.Type(), "Slice")
> (cobra@v1.10 completions.go)'.
>
> Measured, cobra@v1.10.1 completions.go:445-448:
>
>   _, acceptsMultiple := flag.Value.(SliceValue)
>   acceptsMultiple = acceptsMultiple ||
>       strings.Contains(flag.Value.Type(), "Slice") ||
>       strings.Contains(flag.Value.Type(), "Array") ||
>       strings.HasPrefix(flag.Value.Type(), "stringTo")
>
> Four disjuncts, not two. And SliceValue there is cobra's OWN reduced interface,
> declared at completions.go:311 as `type SliceValue interface { GetSlice() []string }`,
> not pflag.SliceValue - so the Design's rejected alternative
> 'Implementing pflag.SliceValue (Append/Replace/GetSlice)' overstates what cobra
> needs: cobra requires only GetSlice.
>
> Neither correction changes the decision. "strings" matches none of the four, so
> the rename costs completion repeatability exactly as the Design says, and the
> alternative stays rejected for the reason given (DisableDefaultCmd is set at
> cli/cmd/root.go:119, so nothing generates the scripts that call __complete).
> The doc comment shipped states the measured condition rather than the bead's.

### `scope` — 2

**`w85b.10#2`** — recorded by `dev-w85b-5`

> retro:scope a THIRD copy of the same inverted clause exists outside this bead's Affected files; filed, not fixed
> evidence: the bead named two files (specs/guides/spec-cli-reference.md and its generated copy). A multiline sweep before editing found five sites carrying "would assert the document declares none": the two named, the two correct references (specs/modules/document-core/document-core-module.md, cli/core/corpus/query.go), AND specs/features/spec-list/spec-list-blueprint.md, which carries the inversion — "SHALL NOT carry a slug, type, status, title or description for it — an absent value here would assert the document declares none". After the fix, `grep -rIl -Pzo "absent\s+value\s+(here\s+)?would\s+assert" . --exclude-dir=.git` returns exactly that one file. Filed as a bead (P3, labels run-introduced, discovered-from this bead) rather than fixed inline: outside this bead's acceptance criteria, and the blueprint is `status: retired`. It was introduced by this run — `git cat-file -e 210058ee:specs/features/spec-list/spec-list-blueprint.md` -> "exists on disk, but not in 210058ee".

**`w85b.7#4`** — recorded by `dev-w85b-4`

> retro:scope The same pflag defect is live on `spec check --only`, which this bead is scoped out of. Filed as a bead (discovered-from w85b.7) rather than fixed inline, because --only has a DIFFERENT behavioural contract and the fix is not a mechanical copy.
>
> evidence: built ./spec from cli/ on the fixed tree and compared:
>
>   $ spec check           > /tmp/c-none.txt  2>&1 ; echo $?   -> 0
>   $ spec check --only "" > /tmp/c-empty.txt 2>&1 ; echo $?   -> 0
>   $ cmp /tmp/c-none.txt /tmp/c-empty.txt                     -> no output (identical)
>
> so `--only "$RULE"` with RULE unset runs EVERY check, the same "empty value dropped by
> pflag's CSV reader" mechanism this bead fixed for --type. Anchor: `cli/cmd/check.go`,
> the line `cmd.Flags().StringSliceVar(&only, "only", nil,`.
>
> The reason it is not a copy of this fix:
>
>   $ spec check --only nosuchrule ; echo $?   -> 0, every rule reported `not-requested`
>   $ spec list  --type runbook    ; echo $?   -> 1, "runbook" is not a registered
>                                                 document type; registered: adr, ...
>
> An unregistered --type value is REJECTED (AC-SPECLIST-002.4, argued explicitly in
> specs/features/spec-list/spec-list-frd.md as the one place the feature departs from
> [[cli-output-standards]]); an unregistered --only rule is answered as an empty
> selection. Making the empty value survive parsing on --only therefore turns
> `--only ""` into "run nothing" rather than into an error, unless somebody first decides
> whether --only should reject unregistered rule identifiers at all. That is a spec
> question, and it belongs to its own bead.
>
> Pre-existing, not run-introduced: `git log --oneline -1 -- cli/cmd/check.go` is 210058e,
> which is also `git log --oneline -1 origin/main`. So it is absent from this run's diff
> and carries no `run-introduced` label.

### `unverifiable` — 1

**`kgir.4#1`** — recorded by `dev-w85b-2`

> retro:unverifiable The `spec list` surface this item documents does not exist in the branch it was written on, so every claim in the new spec-cli-reference.md section is sourced from the approved blueprint rather than measured against a binary. The item states DOC is independent of CORE/FILT/CMD and may be worked in parallel, so this is by design — but it means the guide ships unverified until CMD lands, and nothing re-checks it afterwards.
> evidence: `cd cli && go run ./spec list` -> `Error: unknown command "list" for "spec"` / `Run 'spec --help' for usage.` / `exit status 1` (rc=1). `command -v spec` -> not on PATH. `grep -rn '"list"' cli/cmd/ cli/spec/` -> no matches; `ls cli/cmd/` shows no list.go. Every documented detail is instead quoted from specs/features/spec-list/spec-list-blueprint.md, "`Command Surface` in [[container-spec-cli]] (modified — `list`)": the flag set (`--type` repeatable and comma-separated, `--all`, `--json`), the single `documents` key, the field order path/slug/type/status/title/description, the non-zero exit on an unregistered type, and "SHALL write no count of the summaries, no heading naming the command and no trailing summary".

### `tooling` — 2

**`kgir.3#2`** — recorded by `dev-w85b-1`

> retro:tooling The item requires "Every AC above is cited from its test with an // AC: AC-SPECLIST-00X.Y comment", but a snapshot case is TOML with no test of its own, so a case-proved AC has TWO citation sites and neither document says which is authoritative: the case's `cites` key, and the shared `// AC:` comment on `TestSnapshotCases`. I wrote both. The shared comment is now 34 identifiers on one line and grows with every feature, which is a citation nobody can read back to a specific test — the property ac-traceability-standards asks for is "adjacent placement binds the citation to one specific test", and a comment covering 83 cases binds to none of them.
> evidence: cli/cmd/snapshot_test.go:67 is now a single `// AC:` line carrying AC-SPEC-002.5 through AC-SPECLIST-004.5 — 34 ids. Its own doc comment says the harness "does not itself enforce citation coverage" and describes `cites` as "declared for a human or a future coverage tool". specs/standards/ac-traceability-standards.md gives the rationale for adjacency as "tooling can attribute coverage without language-aware parsing", which the per-case `cites` key satisfies and the shared comment does not. Nothing in the standard mentions data-driven cases at all.

**`w85b.10#3`** — recorded by `dev-w85b-5`

> retro:tooling the multiline-grep warning in the bead is real and worth keeping — a single-line grep finds nothing
> evidence: the phrase spans a line break in both target files. `grep -rn "an absent value would assert" specs/guides/spec-cli-reference.md` returns nothing; the text is wrapped as "...and an absent value would\nassert the document declares none." Only `grep -rIl -Pzo` (or a -A/-B context grep on a shorter fragment) surfaces it. Any future sweep for this class of prose defect in this repo needs -Pzo, because the corpus is hard-wrapped at ~72 columns and almost every quoted sentence crosses a newline.

### `proof` — 11

**`kgir.1#3`** — recorded by `dev-w85b-1`

> retro:proof The defensive copy of the parse failure was verified by negative control rather than asserted: the test that guards it was run against a deliberately broken implementation and observed to FAIL for the stated reason, then against the real one and observed to pass.
> evidence: replaced `failure := *entry.ParseError; summary.ParseError = &failure` with `summary.ParseError = entry.ParseError` in cli/core/corpus/query.go, then `go test ./core/corpus/ -run TestASummarysParseFailureIsACopy` from cli/ →
>     --- FAIL: TestASummarysParseFailureIsACopy (0.00s)
>         query_test.go:199: writing through a summary's parse failure reached the index entry
>     FAIL	github.com/brettcallaghan/agent-plugins/cli/core/corpus	0.011s
> Restored, same command → `ok  github.com/brettcallaghan/agent-plugins/cli/core/corpus 0.013s`. The other five tests in query_test.go have no negative control: they cover a capability that did not exist, so they cannot compile against its absence.

**`kgir.2#1`** — recorded by `dev-w85b-1`

> retro:proof The item's one non-negotiable constraint — the retired filter reads `LifecycleRules.RetiredStatuses` and never tests for a literal spelling — was verified by negative control rather than asserted, and the control fired on TWO independent guards.
> evidence: replaced the body of `withheldStatuses` in cli/core/corpus/query.go with `return map[string]bool{"retired": true}`, then `go test ./core/corpus/ -run "TestTheStatusFilterFollowsTheProfilesRegistrationRatherThanASpelling|TestTheExtensionPointNamesNoRegisteredVocabulary|TestAProfileRegisteringNoRetiredStatuses"` from cli/ →
>     --- FAIL: TestTheExtensionPointNamesNoRegisteredVocabulary (0.01s)
>         profile_test.go:197: query.go:301 names the status "retired"
>     --- FAIL: TestTheStatusFilterFollowsTheProfilesRegistrationRatherThanASpelling (0.00s)
>         query_test.go:431: summaries = [specs/adrs/adr-001.md specs/adrs/adr-002.md specs/frds/frd-checkout.md], want [specs/adrs/adr-001.md specs/contracts/contract-old.md specs/frds/frd-checkout.md]
>     --- FAIL: TestAProfileRegisteringNoRetiredStatusesWithholdsNothing (0.00s)
>         query_test.go:444: summaries = [...], want all four members
>     FAIL	github.com/brettcallaghan/agent-plugins/cli/core/corpus	0.027s
> Restored; `go test ./core/...` from cli/ → every package ok. The behavioural test is the load-bearing one: it registers `draft` as the retired status over the SAME fixture corpus, so an implementation holding a spelling of its own withholds the wrong document and cannot pass it by coincidence.

**`kgir.3#3`** — recorded by `dev-w85b-1`

> retro:proof `scripts/regenerate-cli-snapshots.sh` was run and its whole diff reviewed before commit, as the item's last acceptance criterion requires. It touched exactly one committed artifact outside the new cases, and that change is the intended one.
> evidence: `git status --porcelain -- cli/testdata/cases` after the run showed one modification, `cli/testdata/cases/help-root/expected/checks.log`, and seven new directories. The modification is one added line:
>     +  list        Report every corpus member's path, slug, type, status, title and description
> which is cobra's own rendering of the new subcommand's Use and Short — the container's Change Guidance rule that help text is never hand-written is what makes that diff the proof the registration landed. No other committed case changed. `find cli/testdata/cases/list-* -type f` confirms the new cases carry only case.toml, expected/ and input/ — no `specs.json` was left behind by the generator, unlike the two `criterion-changed*` cases which carry one deliberately.

**`kgir.4#4`** — recorded by `dev-w85b-2`

> retro:proof Pure documentation change — four files, all markdown, no code touched — so there is no targeted test to run and both ACs are declared @manual in the FRD. Verification was the full gate plus reading the two ACs against the shipped text.
> evidence: `git status --porcelain` after the change lists only `specs/standards/frontmatter-standards.md`, `specs/guides/spec-cli-reference.md` and their two generated copies under `plugin/skills/spec/references/shared/`; no `cli/**` path, so scripts/coverage.sh was correctly not required. `scripts/ci.sh` from the repository root -> rc=0, `=== RESULT: PASS`. bb-checks.py with no arguments -> rc=0, byte-comparable to the run's checks_baseline: `27 vendored references are current`, `checked 59 documents / no findings`, `checked 50 reference documents, 37 links, 0 errors`, `checked 13 templates, 0 errors, 34 advisories`, check-skills `0 error(s), 0 warning(s)` in both the `skills` and `skills-strict` legs. AC-SPECLIST-005.1 read back at specs/standards/frontmatter-standards.md:84 ("written to be triaged from") and :93 ("a corpus listing command reports every member's `description`") plus the SHALL sentence beneath. AC-SPECLIST-005.2 read back at specs/guides/spec-cli-reference.md:53 (table row), :107-:147 (the `spec list` block: field order, `--type`, `--all`, `--json`, parse-error and exclusion handling, ordering) and :232-:234 (exit status). The no-length-limit AC verified negatively: `grep -niE 'maxlength|max length|at most [0-9]+ (char|word|sentence)|no more than [0-9]+'` over both edited documents -> rc=1, no matches.

**`kgir.5#1`** — recorded by `dev-w85b-1`

> retro:proof The measurement was taken on a tree that carries every cli/** change this feature makes, which is the one way this item can silently produce wrong figures — it runs in a per-dispatch worktree, not on the merged sprint branch, so a sibling's cli/** commit would be missing from it and CI's --check would then fail on a number nobody could reproduce.
> evidence: `git log --oneline HEAD..<sprint branch>` showed exactly one commit absent from this worktree, `8bff463` (the kgir.4 DOC commit), and `git diff --name-only HEAD...<sprint branch>` showed it touches only plugin/skills/spec/references/shared/frontmatter-standards.md, plugin/skills/spec/references/shared/spec-cli-reference.md, specs/guides/spec-cli-reference.md and specs/standards/frontmatter-standards.md — no cli/** path, so no statement it adds or removes is missing from these figures. The Go version matched the pin: cli/go.mod says `go 1.24`, `go version` says `go1.24.7 linux/amd64`. `scripts/coverage.sh --check` from the repository root exits 0 on the committed README.

**`kgir.6#2`** — recorded by `dev-w85b-3`

> retro:proof NEAR-MISS worth recording, because I nearly shipped a false
> retraction of a correct repo standard while following the rule that exists
> to prevent exactly that. My dispatch prompt claimed `bd export` can read
> comments back. I tested it with `jq keys` on the FIRST record of the JSONL
> and saw no `comments` key, concluded the claim was false, and started
> sweeping the tree for copies -- which found specs/standards/findings-
> capture-protocol.md (v0.6, active) prescribing that exact harvest, plus
> three vendored copies. Had I "corrected" them I would have broken the
> documented harvest path for every skill that vendors it.
> evidence: the wrong measurement --
>   head -1 export.jsonl | jq -r "keys | join(\", \")"
> -> _type, assignee, close_reason, ..., comment_count, ... (no `comments`).
> The right one, because JSONL records are HETEROGENEOUS and line 1 simply
> had no comments --
>   jq -r "keys[]" export.jsonl | sort -u
> -> includes `comments`; and
>   jq -r "select(.id==\"[id]\") | {comment_count,
>     has_comments: has(\"comments\"), n: (.comments|length)}"
> -> {"comment_count":2,"has_comments":true,"n":2}.
> So `bd export` DOES carry comment bodies and the standard is correct; only
> `bd show <id> --json` lacks them (its keys carry comment_count and no
> comments array -- consistent with the finding already recorded on
> kgir.4). `bd comments <id>` is a third working route.
> The transferable trap: `jq keys` on one line of a JSONL export describes
> that RECORD, not the schema; use `jq -r "keys[]" | sort -u` across all
> records before concluding a field is absent. AGENTS.md already warns that
> "correction is the moment of highest confidence, and nothing gates the
> replacement" -- this is a live instance, caught only because the rule says
> to measure the replacement before installing it.

**`kgir.6#4`** — recorded by `dev-w85b-3`

> retro:proof The Definition of Done was verified by running each entry, not
> by reading the blueprint. Recording the commands so the closure is
> reproducible, since a retired blueprint is the one artifact nobody
> re-checks.
> evidence: all from the repository root of worktree bb/dev-w85b-3.
> - Deltas folded: grep -n "list" specs/containers/spec-cli/spec-cli-
>   container.md -> line 181 lists `list` in the built command tree and
>   names corpus.Corpus.DocumentSummaries among the calls; lines 401-445
>   carry the full CMD delta. grep -ni "summary|listing" specs/modules/
>   document-core/document-core-module.md -> lines 990-1060 carry the
>   DocumentSummaries query and both filters.
> - frontmatter description second audience: specs/standards/frontmatter-
>   standards.md:90 "The `description` property has two audiences, and the
>   second one is why it is required", with the SHALL beneath it. Vendored
>   copy verified by CONTENT, not by trusting the sync check: tail -n +6
>   plugin/skills/spec/references/shared/frontmatter-standards.md diffed
>   against the source -> no differences. scripts/sync-shared-references.sh
>   --check -> rc=0, "all 27 vendored references are current".
> - spec-cli-reference documents list: lines 53 (table row), 107-147 (field
>   order, --type, --all, --json, parse-error vs exclusion, ordering).
> - Command actually behaves as documented (the DOC item wrote its section
>   before the command existed, so this was the first time the guide was
>   checked against a binary): build/spec list -> 54 entries under
>   `documents:`; --type adr -> 20; --type bogus -> rc=1 with
>   `Error: corpus: "bogus" is not a registered document type; registered:
>   adr, component-blueprint, ...`.
> - coverage: commit 110a461 (the kgir.5 COV commit) shows README.md
>   +8/-8 and no other path.
> - Tracker Reconciliation: one bead reads closed; three others
>   all read open, as their out-of-scope classification
>   requires.
> - Gates: bb-checks.py with NO arguments -> rc=0, and diffed line-by-line
>   against the run checks_baseline the only differences are go test
>   timings (e.g. cli/cmd 1.397s -> 1.011s); every RESULT line, the
>   "checked 59 documents / no findings" corpus block and the 34
>   template-drift advisories are byte-identical. scripts/ci.sh -> rc=0,
>   "=== RESULT: PASS", which additionally covers rfc-queue and coverage
>   --check that bb-checks.py does not run.

**`w85b.10#1`** — recorded by `dev-w85b-5`

> retro:proof the bead's claim about the code was measured, not inherited, before the prose was changed
> evidence: the bead asserts the CLI emits the five envelope fields as ABSENT. Read at cli/core/corpus/query.go, "type DocumentSummary struct": Path string `json:"path"`; Slug *specmodel.Slug `json:"slug,omitempty"`; Type *specmodel.DocumentType `json:"type,omitempty"`; Status *specmodel.Status `json:"status,omitempty"`; Title *string `json:"title,omitempty"`; Description *string `json:"description,omitempty"`. Pointer + omitempty is absence, not null, so the guide's "an absent value would assert the document declares none" did argue against the shipped behaviour. This is a docs-only bead with no test to write; the struct tags are the measurement. UNVERIFIED at the level of a live invocation: no `spec list --json` was run against a corpus containing a parse-error member, which is what would settle it end to end.

**`w85b.15#1`** — recorded by `dev-w85b-6`

> retro:proof Pre-fix failure observed before the fix, both rows, exactly as the bead described.
> evidence: Test written first, run against unfixed code:
>   $ go test ./cmd/ -run 'TestTypeFlag' -v   (rc=1)
>   === RUN   TestTypeFlagTypedAccessorRefusesRatherThanPanics/flag_unset
>       list_internal_test.go:57: Flags().GetStringSlice("type") panicked: runtime error: slice bounds out of range [:-1]; want a returned error
>   === RUN   TestTypeFlagTypedAccessorRefusesRatherThanPanics/flag_set_to_""
>       list_internal_test.go:64: Flags().GetStringSlice("type") err = nil, want non-nil: pflag must refuse the typed read, not answer it - [] reads as "no filter"
>   --- PASS: TestTypeFlagUsageLineUnchanged
> Note the third assertion PASSED pre-fix: it is the invariant to preserve, not a
> failure to fix, so it functions as a negative control - it would have gone red
> under either rejected remedy.
>
> After Type() -> "strings": same command rc=0, all three PASS.
>
> Byte-identical --help verified against a binary built from the unmodified tree:
>   $ go build -o spec-before ./spec   # at 383bd06, before any edit
>   $ spec-before list --help > help-before.txt
>   ... apply fix ...
>   $ go build -o spec-after ./spec && spec-after list --help > help-after.txt
>   $ cmp help-before.txt help-after.txt   -> no output, rc=0 (BYTE-IDENTICAL)
> The flag line reads "      --type strings   report only members of the named
> document types; repeatable or comma-separated" in both, with --all/--help/--json
> alignment unmoved and no "(default" anywhere.

**`w85b.7#1`** — recorded by `dev-w85b-4`

> retro:proof Both new regression tests were watched FAILING against the unmodified cli/cmd/list.go, for the stated reason, before the fix was restored.
>
> evidence: with cli/cmd/list.go reverted (`git checkout -- cli/cmd/list.go`) and the new
> test and snapshot case in place:
>
>   $ cd cli && go test ./cmd/ -run TestAnEmptyTypeBesideARegisteredOneIsStillRejected -v ; echo $?
>   === RUN   TestAnEmptyTypeBesideARegisteredOneIsStillRejected
>       list_test.go:147: exit code = 0, want non-zero: an empty --type value is not a
>           registered type and must be reported rather than dropped
>           --- stdout ---
>           documents:
>             - path: docs/adrs/adr-001-first.md
>           ...
>       list_test.go:150: stderr = "", want it to report the empty value itself
>       list_test.go:153: stderr = "", want it to name every registered type ...
>       list_test.go:156: stdout reported corpus members, so the empty value was answered
>           rather than rejected
>   1
>
>   $ go test ./cmd/ -run TestSnapshotCases/list-type-empty -v ; echo $?
>   === RUN   TestSnapshotCases/list-type-empty
>       snapshot_test.go:80: exit code = 0, want 1
>       --- stdout ---
>       documents:
>         - path: docs/adrs/adr-001-first.md
>       ...
>   1
>
> With the fix restored both pass (rc=0). The failure is the bead's stated defect: the
> empty value was dropped by pflag before Document Core could see it, so the whole corpus
> came back.

**`w85b.9#1`** — recorded by `dev-w85b-4`

> retro:proof writeBlockScalar was measured as UNREACHED by the snapshot layer before this bead and REACHED after, two independent ways.
>
> evidence — 1. the emitted marker. `|-` is written by exactly one place,
> cli/cmd/yaml.go:207 inside writeBlockScalar (`out.WriteString("|-\n")`), so its presence
> in an expected artifact is the branch being reached.
>
>   before (negative control, against the commit this bead started from):
>     $ git grep -c -- "|-$" HEAD -- "cli/testdata/cases/*/expected/*" | wc -l
>     0
>
>   after:
>     $ grep -rn -- "|-$" cli/testdata/cases/*/expected/
>     cli/testdata/cases/list-all/expected/checks.log:19:    description: |-
>     cli/testdata/cases/list-all-statuses/expected/checks.log:25:    description: |-
>     cli/testdata/cases/list-type-comma/expected/checks.log:25:    description: |-
>     cli/testdata/cases/list-type-repeated/expected/checks.log:25:    description: |-
>
> evidence — 2. the snapshot leg of the coverage measurement moved, and ONLY it:
>
>   $ bash scripts/coverage.sh --check
>   | Unit tests                 | 2,665 / 3,405 | 78.3% |     (unchanged)
>   | Snapshot cases             | 2,608 / 3,405 | 76.6% |     was 2,594 / 76.2%
>   | End-to-end (built binary)  | 1,055 / 3,405 | 31.0% |     (unchanged)
>   | **All layers combined**    | **3,055 / 3,405** | **89.7%** | (unchanged)
>
> +14 statements newly reached by the snapshot layer, with the combined figure flat —
> which is the bead's claim exactly: those statements were already covered by unit tests,
> and the SNAPSHOT layer, the one the blueprint's traceability table names as proving
> AC-SPECLIST-001.3 end-to-end, was not reaching them.

### `environment` — 4

**`kgir.1#4`** — recorded by `dev-w85b-1`

> retro:environment A `bd dolt pull` run by the SIBLING dispatch held the tracker lock for ~5.5 minutes and every other `bd` call in the container blocked behind it, the coordinator's included. It is indistinguishable from a hang while it lasts, and it stalls the per-bead ceremony at exactly the point where the ceremony is all `bd` calls (comment, close, dolt push).
> evidence: `ps -eo pid,etimes,args` showed `26487  269  bd dolt pull` (dispatch dev-w85b-2) with my `27559 bd comment <id>` and the coordinator's `30504 bd list --parent <run epic> --all --limit 0 --json` both queued behind it. `cat /proc/26487/status` → `State: S (sleeping)`, `Threads: 37`; `/proc/26487/wchan` → `futex_do_wait`; `ls /proc/26487/fd | grep -c socket` → `0`, so it was doing no network I/O and looked hung. It was not: it exited on its own at ~340s and every queued call then completed, mine with exit 0. Two consequences for a parallel run: a `bd` call that appears hung should be checked against `ps` for a sibling's `bd dolt` before being treated as a failure, and a 120s tool timeout is too short for any `bd` call made while a sibling may be pulling.

**`kgir.1#5`** — recorded by `dev-w85b-1`

> retro:environment CORRECTS the environment finding above, one clause of which I inferred rather than measured. `bd dolt pull` on this repository does NOT return: it takes the shared embedded-Dolt write lock, starves every other `bd` call in the container, and ignores SIGTERM. My earlier finding said the sibling's pull "exited on its own at ~340s" — I watched its PID disappear and inferred that; I did not measure who ended it, and the coordinator reports killing pulls with `kill -9`. Treat the whole claim as retracted. The trigger is a `bd dolt push` that loses a race to a sibling and fails non-fast-forward; the error text then instructs you to run `bd dolt pull`, which is the trap. A dispatch agent must do NEITHER — tracker durability is the coordinator's, and the repo's Stop hook is the documented pusher.
> evidence: two of my own pulls were killed by the coordinator, both reported as `Killed ... pull rc=137` (PIDs 9148 and 21556); the second was launched with no tool timeout and still had to be killed. `bd dolt push` before and after both attempts printed `Error: push to origin/main: Error 1105 ... ! [rejected] main -> main (non-fast-forward) ... hint: Integrate the remote changes (e.g. dolt pull ...) before pushing again`, `push rc=1`. Not memory pressure: `free -m` during the stall showed 14656 MB free of 16075, and `dmesg | tail` carried no OOM kill, only `idle-reclaim ... drop_caches`. No committed data was lost either time — `bd show <kgir.1>` still reads CLOSED and the four retro comments survived. Coordinator reports dispatch dev-w85b-2 hit the identical trap and measured a sibling stalled at zero CPU for 337s/211s/209s.

**`kgir.4#5`** — recorded by `dev-w85b-2`

> retro:environment `bd dolt push` fails non-fast-forward when a sibling dispatch pushed first, and the documented remedy — `bd dolt pull` — DEADLOCKS the whole run: the pull takes the shared Dolt write lock, never releases it, and every concurrent `bd` call from the live sibling blocks behind it indefinitely. It also ignores SIGTERM. Two dispatches sharing one database means one agent running the documented recovery can freeze the other; the recovery is more dangerous than the failure it treats.
> evidence: bb-close-bead.sh -> "WARNING — commit landed but `bd dolt push` failed" (commit 8bff463 landed, rc=0). Retried by hand: `bd dolt push` -> rc=1, `Error: push to origin/main: Error 1105: To git+https://github.com/brettcallaghan/agent-plugins.git / ! [rejected] main -> main (non-fast-forward) / hint: Updates were rejected because the tip of your current branch is behind its remote counterpart. Integrate the remote changes (e.g. 'dolt pull ...') before pushing again.` Ran that hint: `bd dolt pull` printed only `Pulling from Dolt remote...` and never returned. After 345s, `ps -eo pid,etimes,time,stat,args`: `26487 345 00:00:27 Sl bd dolt pull` (27s CPU, still resident) with three dev-w85b-1 processes stacked behind it at ZERO CPU and 337s/211s/209s elapsed — `bd comment <kgir.1>`, `bd list --parent <run epic> --all --limit 0 --json`, `bd show <kgir.1> --json` — i.e. the live sibling was fully stalled, not slow. `kill -TERM 26487` -> rc=0 and the process was STILL ALIVE 8s later. `kill -9 26487` cleared it; immediately afterwards the only bd process was a fresh sibling call (`bd list --parent <kgir.5>`), confirming all three had unblocked. Post-kill integrity check: `bd show <kgir.4> --json` -> status `closed`, assignee `dev-w85b-2`, closed_at 2026-09-07T07:27:23Z, comment_count 4 — the SIGKILL cost no committed data.
> also: `bd show <id> --json` carries `comment_count` and NO `comments` array (top-level keys: acceptance_criteria, assignee, closed_at, comment_count, created_at, created_by, dependencies, dependency_count, dependent_count, description, design, id, issue_type, labels, metadata, owner, parent, priority, started_at, status, title, updated_at). Reading `.comments` from it returns empty and looks exactly like lost findings; the plain `bd show` COMMENTS section is what actually lists them.
> left undone: the closure of kgir.4 is in the local shared database only — it is NOT pushed to refs/dolt/data. This dispatch pushes nothing by instruction, and the repo's Stop hook (.claude/hooks/dolt-push.sh) is the documented pusher, so whoever reclaims this container must confirm the push landed.

**`w85b.15#5`** — recorded by `dev-w85b-6`

> retro:environment The testdata-index hazard is avoidable by running the built binary from the repository root, where the index path is gitignored.
> evidence: The bead warns: "Do NOT run a built spec binary inside
> cli/testdata/cases/*/input: it writes a persisted index there and dirties the
> tree." Grooming hit this.
>
> The behavioural criteria (--type "", --type "" --type adr, --type adr,frd vs
> repeated, bare list) all need a real corpus, so they cannot simply be skipped.
> Running them from the worktree ROOT satisfies them and leaves the tree clean,
> because .gitignore:27-28 covers the default index location:
>
>   specs/specs.json
>   rfcs/specs.json
>
> Measured - all four criteria from the repo root, binary built to a scratchpad
> path outside the tree:
>   spec list --type ""                 rc=1  Error: corpus: "" is not a registered document type; registered: adr, component-blueprint, ... system-blueprint
>   spec list --type "" --type adr      rc=1  same error
>   spec list --type adr,frd            rc=0  29 entries
>   spec list --type adr --type frd     rc=0  29 entries, cmp against the above: identical
>   spec list                           rc=0  53 entries
>   $ git status --porcelain -- cli/testdata/   -> empty
>
> Final tree carried exactly three entries (README.md, cli/cmd/list.go modified;
> cli/cmd/list_internal_test.go new) - no stray index or output files anywhere.
> Worth stating positively in the next bead that needs a live corpus: the hazard
> is the working DIRECTORY, not the binary.

### `evidence` — 1

**`w85b.20#1`** — recorded by `coordinator-w85b`

> retro:evidence coordinator: I propagated the false reachability claim myself. My pass-3 review dispatch prompt stated, as settled fact and as the reason the completion trade was acceptable: 'Cobra flag-name completion is a known accepted cost — root.go sets DisableDefaultCmd = true, so the CLI ships no way to generate the scripts that would call __complete.' I inherited that from the pass-3 grooming report and did not measure it. Review pass 3 measured it and it is false: __complete ships and answers (rc=0), and DisableDefaultCmd gates only the user-facing completion command. This is exactly the propagation AGENTS.md's 'Tool claims are measured at the point of use' names, on the surface its own table calls worst — a dispatch prompt, which reaches a fresh agent with no way to learn the claim was retracted. It reached two agents before being caught.

## Routing

Routing is a lookup on the kind recorded at capture, not a reading of the
finding. Where a kind is present but the audience is unclear, the **merge test**
settles it: *where would the fix land?*

| branch | count | where it goes |
|---|---|---|
| `proof` — terminates here | **11** | attached to what it supports, in this document. No issue |
| `environment` — terminates here | **4** | `AGENTS.md`'s environmental-failure ledger. No issue |
| project audience | **23** | a bead in this tracker, or a proposal in `rfc-0027` |
| skill audience | **0** *from the `retro:` channel* | — |
| **UNROUTED** | **1** | `w85b.20#1` |

**Both `tooling` findings route to the PROJECT, not to the skill audience, and
the merge test is unanimous.** The lookup sends `tooling` to the skill
audience; here `kgir.3#2` lands in `specs/standards/ac-traceability-standards.md`
and `cli/cmd/snapshot_test.go`, and `w85b.10#3` lands in `AGENTS.md`'s sweep
procedure. This repository vends the `sf` skills — `spec`, `groom-beads`,
`groom-feature-blueprint`, `retro-spec`, `retro-skill` — so a finding against one
of those is ordinary work here. Only the `swarm` plugin's skills and CLI leave
through `.sf/outbox/swarm/`. `swarm.toml` deliberately leaves
`[retros].skill_issues_dir` unset and states that split in prose, because it is a
rule the binary switch cannot carry.

**Every skill-audience item in this record came from outside the finding
channel.** The four coordinator tooling items that are not `environment`
findings were hand-carried in a dispatch prompt; the harvest query matches them
on neither the prefix nor the scope.
That is the exact defect already filed as
`.sf/outbox/swarm/2026-09-05-coordinator-tooling-findings-are-invisible-to-the-harvest.md`,
now at **recurrence 1** — the second consecutive run where the record's most
actionable skill-audience material survived only because a human-readable prompt
carried it.

**The unrouted finding.** `w85b.20#1` is recorded as `retro:evidence`, and
`evidence` is not one of the protocol's nine kinds. It is the *coordinator*
recording, in the finding channel, that it propagated a false reachability claim
into a dispatch prompt. Two things are worth separating:

- It is an **improvement** on the previous run, where the coordinator used no
  finding channel at all. It reached the harvest.
- It chose a kind that does not exist, so the routing lookup drops it. On the
  taxonomy the closest fits are `stale` (a claim that contradicted repo reality)
  and `tooling` (a dispatch prompt that misled). One unrouted finding in 39 is a
  healthy taxonomy; this one is worth naming because of *who* wrote it.

## Validation

**The 12 declared checks pass at head, exit 0 — re-measured for this record,
not inherited.**

```
/root/.claude/skills/swarm/scripts/bb-checks.py    → exit 0
```

Every leg matches the run's `checks_baseline` line for line apart from `go test`
timings: `check-skills` `0 error(s), 0 warning(s)` in both the `skills` and
`skills-strict` legs, `all 27 vendored references are current`, `all 3 vendored
strata docs match /root/.cache/strata-src`, `generated artifacts are up to
date`, `checked 59 documents / no findings`, `checked 50 reference documents, 37
links, 0 errors`, `checked 13 templates, 0 errors, 34 advisories`. The baseline
itself recorded exit 0 on the untouched tree, so no failure in this run's
history was inherited.

**Coverage is not in that set and was not re-run here.** `scripts/coverage.sh
--check` is in `scripts/ci.sh` and in `.github/workflows/cli.yml`, not in
`swarm.toml`'s `[[checks]]`. Three beads regenerated the README figures during
the run and each reported `--check` green afterwards on its own tree; this
record did not re-measure it on the merged head. That is the one declared-gate
gap in this validation, and it is candidate **A**.

**CI was not observed.** No workflow run was checked from this session, and
`cli.yml`'s `coverage` job has not been seen to pass on `861fa52`.

**The working tree is clean.** `git status --porcelain` is empty; `git worktree
list` shows the repository itself at `861fa52` plus one stray detached-HEAD
worktree at `2c8bbf3` under this session's scratchpad, which is outside the
repository and holds no work.

## Pushes

**Nothing was stranded when this record was started.**
`git log origin/<sprint branch>..<sprint branch>` is empty;
both refs stand at `861fa52`.

No merge to the default branch was made or attempted — `merge_to_default` is
`false`.

**Whether a pull request was opened is not recorded on the epic and was not
verified from this session.** `pull_request` has no manifest key at all
(candidate **C**), so there is nowhere it could have been recorded.

*The commit carrying this document, its `log.md` row and the ledger entry is the
one thing on the branch that is NOT pushed: this retro was told not to push and
never to run `bd dolt push`, so it did neither.*

## What this record could NOT verify

Stated one per line, because an unstated gap is indistinguishable from one that
does not exist.

- **That `concurrency_cap: 20` is meaningless.** What is measured is that
  `bd dolt pull` holds an exclusive writer lock and starves siblings — three
  sibling `bd` calls at zero CPU for 337s/211s/209s, all released the instant
  the pull was SIGKILLed. Whether two ordinary `bd` calls (a `comment` and a
  `show`, say) serialize against each other in the absence of a pull was
  **never measured**, and the stronger claim should not be repeated as settled.
  Settle it by timing two concurrent `bd comment` calls against one database.
- **The `bd dolt pull` deadlock was not reproduced for this record**, and
  deliberately: reproducing it means freezing the container for twenty minutes.
  The evidence is `kgir.1#4`, `kgir.1#5` and `kgir.4#5`, which quote PIDs,
  `/proc` state, `free -m`, the ignored `kill -TERM` and the effective
  `kill -9`. Two independent dispatches measured it. The ledger entry added by
  this retro says so explicitly rather than presenting it as re-measured.
- **Which model served each dispatch.** `manifest.models` and `manifest.effort`
  are both unset — measured, `swarm manifest get --run <run epic> --key
  models --key effort` returns `ok: false` for both. Every bead carries a tier
  and `swarm.toml` maps tiers to models, but nothing records what was served.
  Any issue filed from this run therefore carries a `model` field that is
  *derived*, which is exactly the provenance the issue template calls load
  bearing. Already filed upstream as
  `.sf/outbox/swarm/2026-09-04-pre-flight-check-never-writes-models-or-ledger.md`;
  **recurrence 1**.
- **The review reports for passes 2 and 3.** Only pass 1's report is durable —
  one comment on `w85b.4`. `w85b.14` and `w85b.19` carry no comment and no
  description, so what pass 2 and pass 3 examined, and what they cleared, is
  reconstructible only from the beads they filed.
- **Whether `w85b.20`'s and `w85b.21`'s findings are the only ones left.** Pass
  3 was the last pass; nothing has reviewed the tree as it stands at `861fa52`
  after `w85b.15` landed. The run stopped on the budget, not on a clean pass.
- **`bb-close-bead.sh --trailers-from` exiting 2 as a transcript.** The cause is
  measured here (`trailers` unset) and the consequence is measured here (no
  commit carries a trailer); no invocation of the script is recorded anywhere
  durable.
- **`kgir.6#4`'s Definition-of-Done verification was not re-run.** It is a
  `proof` finding carrying its own commands and outputs, cited as evidence
  rather than re-measured.
- **The claims of the vendored `shared/` citation bead and the `spec check --only` pflag bead were not
  re-measured** for this record. Both beads quote their evidence inline.

## Candidates

Seven, clustered from the 39 findings plus the coordinator's six tooling records
and deduped against what is already filed. **The bar is that it changes how the
next run goes.**

**Nothing below has been filed.** This retro asks a human once, and no answer
has been recorded, so the candidates are put and held. They stay decidable from
this document alone.

### Project audience → beads in this tracker

**A — Coverage is a completion condition that no surface a dispatched agent
reads mentions.** Priority **2**. New. Recurrence **3 within this single run**.

> Three dev agents independently discovered, from `AGENTS.md` rather than from
> anything they were given, that `scripts/coverage.sh` must be re-run and the
> README diff committed for any `cli/**` change. No bead's acceptance criteria
> mention coverage; `swarm.toml`'s `[[checks]]` deliberately excludes it; the
> run's `checks_baseline` does not mention it. `w85b.7` measured its own tree
> stale — `scripts/coverage.sh --check` → exit 1, committed `3,399/3,051/89.8%`
> against a measured `3,405/3,055/89.7%` — with all declared checks green.
>
> The trade-off `AGENTS.md` documents (three instrumented test legs is too slow
> to pay per bead) is deliberate and worth keeping. What is missing is a pointer
> where an agent will see it. Proposed change, one of:
> **(a)** `groom-beads` adds "and `scripts/coverage.sh --check` passes" to the
> declared-checks acceptance criterion of any bead whose `c4_changes` touch
> `cli/**`; or **(b)** the `checks_baseline` a run records carries one sentence
> saying coverage is outside the declared set and is a completion condition for
> `cli/**` changes. (a) is preferred: it reaches the surface the agent actually
> reads. Measured cost: `scripts/coverage.sh` ran in under three minutes on this
> container.
>
> Findings: `w85b.7#3`, `kgir.5#2`, `w85b.15#4`.

**B — `AGENTS.md`'s retraction procedure tells you to grep, and a single-line
grep cannot find a hard-wrapped claim.** Priority **3**. New.

> Step 2 of "Retracting a claim found false" is *"Grep the claim's distinctive
> phrasing across the tree BEFORE you fix the source."* This corpus is
> hard-wrapped at roughly 72 columns, so almost every quoted sentence crosses a
> newline and a plain `grep -rn` returns nothing — measured on the exact clause
> `w85b.10` was sweeping for. Only `grep -rIl -Pzo` (or a short-fragment grep
> with context) surfaces it, and that sweep found a **third** copy the bead's
> Affected-files list did not name.
>
> Proposed change: step 2 gains one clause — sweep with `grep -rIl -Pzo` or on a
> fragment short enough to fit one line, because a line-oriented grep silently
> reports zero for a wrapped claim, which is indistinguishable from the claim not
> being there.
>
> Adjacent and distinct: an existing P3 bead (open) proposes a different
> addition, to step 4. `w85b.20` is a live instance of *bfng's* shape — a false
> claim left in place as the justification for correct code — and is an unfiled
> recurrence for it.
>
> Finding: `w85b.10#3`.

**C — A blueprint delta's Anchors do not name the tests that already enforce
its criteria.** Priority **3**. New.

> `kgir.2`'s acceptance criterion "the string `retired` appears nowhere in the
> filter's own logic" was **already enforced** by
> `TestTheExtensionPointNamesNoRegisteredVocabulary` in
> `cli/core/corpus/profile_test.go`, which walks every non-test Go file in the
> package and errors on any literal matching a value the profile registers.
> Neither the bead's Anchors nor the blueprint's FILT delta names that file, so
> an implementer who did not go looking would have written a second, redundant
> source scan to satisfy a criterion that already had a guard.
>
> Proposed change: the feature-blueprint template's delta Anchors section
> requires naming any **existing** test that already enforces a criterion in the
> delta, not only the files the delta changes.
>
> Finding: `kgir.2#2`.

*Deduped away, deliberately:* the coverage-measured-on-the-wrong-tree half of
`kgir.5#2` is already `rfc-0027` **P3**; the two-citation-surfaces cluster
(`kgir.3#2`, `kgir.6#1`) is `rfc-0027` **P1**; the snapshot-harness-cannot-
express-two-runs cluster (`kgir.3#1`, `kgir.6#3`) is `rfc-0027` **P2**;
`bd-health.sh` has two existing beads; the vendored
`shared/` citation defect, `spec check --only ""` and the third
inverted-clause copy each have their own bead.
**Seven clusters already have homes.** Nothing new goes on those piles.

### Skill audience → `.sf/outbox/swarm/`, for delivery to `brettcallaghan/swarm`

**D — Two of the seven standing-authorization rows are not declared manifest
keys, and one of them governed this run.** Priority **1**. New.

> Measured:
> `swarm manifest get --run <run epic> --key admit_review_findings --key pull_request`
> → exit 3, `"admit_review_findings, pull_request: not a declared manifest key"`.
> `manifest.standing_auth` accordingly carries five rows, not seven.
>
> The cost is not cosmetic. This run **did** admit review findings mid-run and
> worked four of them (`w85b.7`, `.9`, `.10`, `.15`), and the key that authorizes
> that has nowhere to be recorded. A resumed session, a retro, or an auditor
> reading this manifest cannot tell whether admitting them was authorized at the
> gate or improvised mid-run. `pull_request` has the same problem one tier down:
> whether a draft PR was opened is unrecordable and therefore unverifiable.
>
> Proposed change: declare `pull_request` and `admit_review_findings` as manifest
> keys, and have `swarm-scope` write all seven resolved rows into `standing_auth`
> rather than the five that happen to have homes.
>
> Extends `.sf/outbox/swarm/2026-09-04-standing-auth-has-no-home-for-dispatch-failure-threshold.md`,
> which named a third row — now fixed; `dispatch_failure_threshold` is present.

**E — `manifest.admitted` is never updated when a run admits review findings, so
it under-reports the closure set.** Priority **2**. New.

> `manifest.admitted` names six beads. The run closed **ten**, in ten commits.
> The four admitted mid-run through the review loop's target epics (`w85b.13`,
> `w85b.18`) were never added, so the manifest's own answer to "what did this run
> take on" is wrong by 40% and can only be corrected by walking `git log` and the
> epic tree — which is what this retro had to do.
>
> Proposed change: whatever admits a review finding into a loop's target epic
> appends it to `manifest.admitted` in the same write.

**F — `swarm-review`'s filing snippet prescribes `bd create --discovered-from`,
which `bd` does not have.** Priority **2**. Recurrence **1** — same defect, second
surface, first one still unfixed.

> `swarm-review/SKILL.md:112` tells the reviewer to file each finding with
> `bd create ... --parent {EPIC} --discovered-from <origin-bead-id>`, and :115
> explains why the edge matters. Measured on this container: `bd create --help`
> lists no such flag, and the string `discovered-from` appears in its help
> exactly once, inside the description of `--deps`
> (`--deps 'discovered-from:bd-20,blocks:bd-15'`). The correct form is
> `--deps discovered-from:<id>`.
>
> `review-w85b` evidently worked around it — every bead it filed carries a
> correct `discovered-from` edge — so the cost is a step that cannot be followed
> literally and a reviewer that must improvise a tracker write.
>
> The identical defect in **`swarm-retro`** was filed by the previous run
> (`.sf/outbox/swarm/2026-09-05-retro-filing-snippet-uses-a-flag-bd-does-not-have.md`)
> and `swarm-retro/SKILL.md:202` still carries it, so this is one defect on two
> surfaces with the first unfixed. Proposed change: correct both snippets, and
> add the flag to whatever self-test covers the scripts.

**G — The close ceremony points an agent at `bd dolt push`, whose failure
message points at `bd dolt pull`, which deadlocks the container.** Priority
**1**. New.

> The chain, measured across two dispatches in this run:
> `bb-close-bead.sh:270` runs `bd dolt push` after each close; on failure
> `:271` prints *"The closure is in the local database only; push it before the
> container is reclaimed."* A hand-run `bd dolt push` that lost a race to a
> sibling fails non-fast-forward, and `bd`'s own error text ends
> `hint: Integrate the remote changes (e.g. 'dolt pull ...') before pushing
> again`. Following that hint takes the shared embedded-Dolt write lock, never
> returns, **ignores `SIGTERM`**, and starves every other `bd` call in the
> container until `kill -9`.
>
> Cost: two dispatches hit it independently, each freezing the run for roughly
> twenty minutes, before the coordinator issued a binding correction. Nothing in
> `swarm-dispatch` or in `bb-close-bead.sh`'s own warning says a dispatch agent
> must never run `bd dolt pull`.
>
> Proposed change: `bb-close-bead.sh`'s warning names the safe remedy and names
> the trap — the closure is durable in the local database, tracker durability is
> the coordinator's, and `bd dolt pull` must not be run from a dispatch — and
> `swarm-dispatch`'s ceremony carries the same one-line prohibition.
>
> Findings: `kgir.1#4`, `kgir.1#5` (its own correction), `kgir.4#5`.

*Recurrence bumps, not new issues:*
`.sf/outbox/swarm/2026-09-05-a-run-resolves-no-trailers-so-no-commit-carries-one.md`
→ **recurrence 1** (no commit in this run carries a trailer; `trailers` unset,
measured);
`.sf/outbox/swarm/2026-09-05-coordinator-tooling-findings-are-invisible-to-the-harvest.md`
→ **recurrence 1** (three of this run's four coordinator tooling defects reached
this record only by dispatch prompt);
`.sf/outbox/swarm/2026-09-04-pre-flight-check-never-writes-models-or-ledger.md`
→ **recurrence 1** (`models`, `effort` and `ledger` all unset on this manifest).

## What was filed

**Nothing was filed as an issue, and that is a state rather than an outcome:**
the seven candidates above were put and no human answer has been recorded. They
are complete enough to file from without re-reading the run.

**One routing destination needed no decision and was written.** The four
`environment` findings terminate at the project's environmental-failure ledger,
where the next agent looks, and the `bd dolt pull` deadlock is the kind of
hazard that costs twenty minutes the first time somebody meets it. An entry was
added to `AGENTS.md` under "Environmental-failure ledger", citing the findings
that measured it and stating plainly that this record did not re-run it.

**`manifest.findings_path` on the run epic names this document**, which
is what makes it findable from the epic id alone.

## Counts

| | |
|---|---|
| findings harvested | **39** across 11 beads |
| routed — `proof` (terminates) | 11 |
| routed — `environment` (terminates, → ledger) | 4 |
| routed — project audience | 23 |
| routed — skill audience *from the finding channel* | 0 |
| **unrouted** | **1** (`w85b.20#1`, kind `evidence`, which does not exist) |
| coordinator tooling defects, outside the finding channel | 6 |
| candidates | **7** — 3 project, 4 skill |
| candidates deduped into existing homes | 7 clusters |
| filed | **0** — put to a human, no answer recorded |
| declined | **0** |
| beads left open and unworked | 5 — `w85b.8` (human), `.11`, `.16`, `.20`, `.21` |

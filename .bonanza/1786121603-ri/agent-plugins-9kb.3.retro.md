# agent-plugins-9kb.3 — Extract body and frontmatter references with their spans

## Bead description (as given)

**Blueprint:** specs/features/reference-integrity/reference-integrity-blueprint.md
— delta **REX**, section "References in [[module-document-core]] (added —
extraction) — REX".
**FRD:** specs/features/reference-integrity/reference-integrity-frd.md.
**Delivery:** staged (stage 1 of 2); this feature's closure set is
`bd list --label feature-spec-reference-integrity`.

Build the extraction half of [[module-document-core]]'s declared References
component: find every reference a document makes, in the body and in the
frontmatter properties the active profile registers, and record each with the
position a reader must open to fix it.

The two rules that matter most here are both about what is **not** a reference —
a backticked illustration, and a frontmatter value describing the notation —
because a checker that reports either is a checker that gets ignored. Both were
observed in this repository's own corpus.

**Verification timing — frontmatter extraction against the real corpus.** This
item's frontmatter half cannot be exercised against this repository's own corpus
until PREG registers this framework's reference-bearing properties. Verify it
here with a test profile that registers properties of its own; code work does
not wait on PREG. The profile self-application gate and the final
Definition-of-Done item are the backstop.

**Verification timing — participating criteria.** This item proves
AC-SPECREF-001.1, -001.2, -001.4, -001.5 and -001.7 (all unit) and owes those
citations. It also participates in AC-SPECREF-002.2 (proved by RES) and
AC-SPECREF-002.3 (proved by CHK), both outside this item's dependency closure;
verification of those two waits for RES and CHK.

**Boundary — agent-plugins-gz3** (out of scope, and a hard constraint here):
index entries carry only the DocumentFrontmatter envelope, so per-type
properties are lost. **Nothing in this item may reach a reference-bearing value
through IndexEntry.frontmatter**, because that envelope carries neither the
per-type properties nor the spans AC-SPECREF-001.7 requires. Read from the
generic parsed frontmatter instead. This item neither depends on gz3 nor
resolves it.

### DESIGN (as given)

Anchors: new package cli/core/references (the package [[module-document-core]]'s
References component names); cli/core/frontmatter for the per-item span
widening; cli/core/span; goldmark added to cli/go.mod.

- Parse every [[…]] occurrence from a document's **body** — the content after
  the frontmatter block — and record each with its Span (AC-SPECREF-001.1).
- Do **not** record an occurrence inside a fenced code block or an inline code
  span (AC-SPECREF-001.2). Locate code regions with goldmark, the markdown
  parser [[adr-015]] already chose for this module, rather than a hand-rolled
  backtick scan.
- Treat everything before the first # as the reference's slug half and
  everything after it as the anchor, which is recorded and **not interpreted**
  (AC-SPECREF-002.2). No other separator is interpreted.
- Read reference-bearing frontmatter properties from the **generic parsed
  frontmatter** and never from the typed DocumentFrontmatter envelope
  (AC-SPECREF-001.7).
- WHERE a reference-bearing property holds a sequence, record one reference per
  item, each at that item's own span.
- Do **not** record a [[…]] occurring in a frontmatter value as a reference
  (AC-SPECREF-001.4).
- Read the contents of an **HTML comment block as body text**.
- Record each occurrence separately and preserve source order
  (AC-SPECREF-001.5).
- Express every position as a Span from the Spans component.

Package layout and function signatures are yours to choose.

### Acceptance criteria (as given)

- Every [[…]] in a document body is recorded with its Span (**proves
  AC-SPECREF-001.1**, unit).
- A reference inside a fenced code block or an inline code span is not recorded,
  with code regions located by goldmark rather than a backtick scan (**proves
  AC-SPECREF-001.2**, unit).
- A [[…]] in a frontmatter value is not recorded as a reference (**proves
  AC-SPECREF-001.4**, unit).
- Repeated references to the same target are recorded as separate occurrences in
  source order, each with its own span (**proves AC-SPECREF-001.5**, unit).
- Values of profile-registered reference-bearing frontmatter properties are
  recorded as outbound references, distinguishable from body references and
  naming the property they came from (**proves AC-SPECREF-001.7**, unit).
- A sequence-valued reference-bearing property yields one reference per item,
  each at that item's own span.
- A reference carrying # is split on the first #; the anchor is recorded and not
  interpreted; a [[slug|text]] form carries slug|text as its slug half.
- Text inside an HTML comment block is read as body text and contributes
  references.
- No reference-bearing value is reached through IndexEntry.frontmatter
  (agent-plugins-gz3 boundary).
- Every position is a Span from the Spans component.
- All citations use the AC: marker adjacent to the test declaration per
  [[testing-standards]].
- [[module-document-core]] is folded: the References component records its
  extraction half and the package that realizes it; the Frontmatter component
  records per-item spans for sequence values; the Change Guidance rule on the
  Spans scanner records the goldmark narrowing. Set version to 0.3 if not
  already set.

## Retro

### Approach

Three code changes plus one fold:

1. **`cli/core/frontmatter/frontmatter.go`** — `Property` gains `Items []Item`;
   `Item` is `{Position, Raw, Value}`. Populated by a new `sequenceItems`
   helper from the YAML sequence node, with `elementText` and `isScalar`
   beside it.
2. **`cli/core/references/references.go`** (new package) — `Options` and
   `Extract(Options) []specmodel.Reference`, with `frontmatterReferences`,
   `bodyReferences`, `bodyStart` and `reference` unexported.
3. **`cli/core/references/code.go`** — the goldmark code-region mask:
   `codeRegions(body) regions`, `regions.covers`, `linesOf`, `spanOf`.
4. Folds into `specs/architecture/module-document-core.md`: References
   half-built with seven pinned semantics, Frontmatter's sixth pinned
   semantic, the Change Guidance narrowing on the Spans-scanner rule, the
   components diagram and the "six of eight are built" prose. Version was
   already 0.3.

`goldmark v1.8.5` added to `cli/go.mod` (direct after `go mod tidy`).

### Implicit decisions, and why

- **`Extract` returns `[]specmodel.Reference` with `Resolution` left unset**,
  rather than an extraction-local type. The bead forbids redeclaring the MODEL
  types, and a `Reference` is complete except for what only the corpus can
  answer. The happy side effect is a fail-closed one: `ReferenceResolution`
  marshals to an error when its variant is nil, so an unresolved-in-the-code
  sense reference cannot reach the index. A test asserts exactly that.
- **`Path` is a parameter, not something the package derives.** The caller
  walking the corpus knows it; this package sees bytes.
- **`Options` takes `ReferenceProperties []string`, not a `corpus.Profile`.**
  `corpus` will import `references` in RES, so the reverse import would be a
  cycle; and what is reference-bearing is a per-type answer the caller has
  already looked up.
- **Frontmatter references are read from `frontmatter.Document`, passed in
  already parsed**, not re-parsed and not read from `IndexEntry.frontmatter`.
  `TestExtractionReadsNoIndexEntry` parses this package's own imports and fails
  if `core/specindex` appears, so the gz3 boundary is a gate rather than a
  review note.
- **Position semantics differ by origin, deliberately, and are documented on
  `Extract`:** a body reference's position covers the text *between* the
  brackets (the brackets are notation, and it makes `src[Position] == Text`);
  a frontmatter reference's position covers the value's own source text
  **including** YAML quoting, because that is the extent a rewrite replaces.
- **`Text` for a frontmatter reference is the parsed scalar, not the raw.**
  The model's comment says "the raw value of a reference-bearing frontmatter
  property", which is ambiguous; taking it literally would make
  `Text == "\"adr-012\""` and the slug half carry quotes, so every quoted
  frontmatter slug would report unresolved.
- **No trimming of a body reference's text.** `cross-document-references.md`:
  "only the slug SHALL appear between the brackets". `[[ adr-012 ]]` therefore
  carries a slug half with spaces and reports as unresolved, which is what the
  standard says it is. Trimming would be inventing a tolerance the standard
  does not grant.
- **A registered property whose value is not text contributes nothing** — a
  number, a mapping, a nested sequence. Whether such a value belongs in a
  reference-bearing property is a schema question, and the schema has the
  vocabulary to answer it.
- **A fence's info string is masked with its content.** ```` ```[[adr-004]] ````
  is code by the same argument as the lines inside the fence. Covered by its own
  test.
- **`regions.covers` tests overlap, not containment.** A reference that starts
  inside a code span and ends outside it is still something the author wrote
  inside backticks; reporting half of it is worse than reporting none.
- **`sequenceItems` locates a scalar element by its own token, not by the extent
  running to the next element.** Measured: for a block sequence the
  extent-to-next-token approach swallows the following element's `- `
  indicator (`"container-spec-cli"\n  - `), and trimming that back off by hand
  is where a span parser starts guessing. Composite elements have no single
  token, so they are the one case that does use the extent, with a trailing
  indicator removed — verified exact for `- key: value`.
- **`Items` is nil for a non-sequence and non-nil-empty for `[]`**, so a caller
  can tell "not a list" from "an empty list". Both cases have a test.
- **File split**: extraction in `references.go`, the code mask in `code.go`. The
  mask is the only part that knows markdown exists, and keeping it in one file
  makes the goldmark surface one file wide.
- **Package-level `var parser = goldmark.New().Parser()`** — CommonMark, no
  extensions, built once. The only question asked of it is where the code is.

### Gate results

All run as independent steps from `/home/user/agent-plugins`.

| Gate | Exit |
|---|---|
| `cd cli && gofmt -l .` | 0, printed nothing |
| `cd cli && go vet ./...` | 0 |
| `cd cli && go build ./...` | 0 |
| `cd cli && go test ./core/... ./profile/...` | 0 |
| `cd cli && go test ./...` (whole module) | 0 |

No `schemas/*.strata` changed, so `generate-schemas.sh` was not run and no
generated artifact moved. No regeneration diffstat for this bead.

### Anything surprising

- Two throwaway probe tests were written and deleted while implementing, to
  find out what goccy/go-yaml reports for a sequence element's token. Worth
  recording: `tok.Position.Offset` is a **one-based rune** offset (the same
  trap `absolute()` already documents for property keys), `Origin` for a block
  element is `" \"container-spec-cli\""` — leading space, no `- ` — and an
  empty flow sequence `[]` yields a `SequenceNode` with zero values rather than
  a non-sequence node.
- goldmark masked all five negative fixtures on the first run: an inline code
  span, a fenced block, an indented block, a double-backtick span, and a fence
  opened with a trailing space. That is the case for the blueprint's insistence
  on a parser — the trailing-space fence and the double-backtick span are both
  cases a naive scanner gets wrong.
- The References component record block still describes the whole designed
  component ("Wikilink discovery, heading-anchor resolution, and the inbound and
  outbound reference graph"); only the **Built** paragraph was changed to
  "Half built — extraction". The block describes what the component is, not what
  exists yet.

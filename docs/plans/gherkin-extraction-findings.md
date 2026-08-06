# Gherkin extraction: what the Go library gives us

Research record behind [[adr-018]]. Everything here was measured against
`github.com/cucumber/gherkin/go/v42` (v42.0.1) and
`github.com/cucumber/messages/go/v34` (v34.2.0), both MIT.

The question this answers: after the library parses a Gherkin block,
what structure do we actually hold, and can we extract every field the
spec index and a downstream structured store need?

**Answer: yes, with five integration details that are easy to get wrong.
All five are recorded below with their resolution.**

---

## 1. The shape: a Go struct hierarchy with JSON tags

`ParseGherkinDocument` returns `*messages.GherkinDocument` — an ordinary
Go pointer tree, not an opaque handle. Every node carries `json:` tags,
so the tree *is* the serialization format; there is no separate DTO to
write.

```
GherkinDocument   { Uri, Feature *Feature, Comments []*Comment }
└── Feature       { Location, Tags[], Language, Keyword, Name,
                    Description, Children []*FeatureChild }
    └── FeatureChild { Rule *Rule | Background | Scenario }  // union
        └── Rule  { Location, Tags[], Keyword, Name, Description,
                    Children []*RuleChild, Id }
            └── Scenario { Location, Tags[], Keyword, Name, Description,
                           Steps []*Step, Examples []*Examples, Id }
                ├── Step     { Location, Keyword, KeywordType, Text,
                               DocString, DataTable, Id }
                └── Examples { Location, Tags[], Keyword, Name,
                               Description, TableHeader *TableRow,
                               TableBody []*TableRow, Id }
                    └── TableRow { Location, Cells []*TableCell, Id }

Location { Line int64, Column int64 }
Tag      { Location, Name, Id }
```

`StepKeywordType` is an enum with values `Context`, `Action`, `Outcome`,
`Conjunction`, `Unknown` — this is what makes precondition / trigger /
expected-result typed fields rather than parsed English.

**JSON round-trips losslessly.** Marshal → Unmarshal → Marshal produced
byte-identical output. The wire shape is the published `cucumber/messages`
schema, which is what external test-management tooling reads.

---

## 2. What we extract, end to end

Verified against an FRD-shaped markdown document containing a
`Scenario Outline` with an `Examples` table, a plain `Scenario`, a
`Rule` with a description, and an em dash in the prose. Output:

```json
{
  "ac_id": "AC-SPEC-003.5",
  "frd": "frd-spec-validation-cli",
  "req": "Metadata Conformance",
  "name": "Slug pattern mismatch names actual and expected",
  "negative": false, "manual": false, "level": "unit",
  "steps": [
    { "type": "Context",     "text": "a document of type <type> declaring slug \"<slug>\"" },
    { "type": "Action",      "text": "the CLI checks the document" },
    { "type": "Outcome",     "text": "a finding names the actual slug \"<slug>\"" },
    { "type": "Conjunction", "text": "the finding names the expected pattern \"<pattern>\"" }
  ],
  "examples": [
    { "type": "adr", "slug": "adr-x",    "pattern": "^adr-[0-9]{3}$" },
    { "type": "frd", "slug": "checkout", "pattern": "^frd-[a-z-]+$" }
  ],
  "rationale": "Frontmatter errors are reported mechanically so review attention goes to whether a document says the right thing — not whether it is shaped correctly.",
  "span": { "offset": 432, "length": 432, "line": 18, "column": 5 }
}
```

Field provenance:

| Field | Source |
|---|---|
| `ac_id` | the `Scenario` tag matching `^@AC-[A-Z0-9]+-[0-9]{3}\.[0-9]+$` |
| `frd` | the `Feature`-level tag |
| `req` | the enclosing `Feature.Name` (see the note below) |
| `negative`, `manual`, `level` | `@negative`, `@manual`, `@level-*` tags |
| `steps[].type` | `Step.KeywordType` |
| `examples[]` | `Examples.TableHeader` zipped with each `TableBody` row |
| `rationale` | `Rule.Description` where a `Rule` groups the criterion, else `Scenario.Description` |
| `span` | computed — see gap 1 |
| `story` / requirement context | `Feature.Description` — one text field per requirement |

**Note on `req`.** An early version of this record took `req` from
`Rule.Name`, because the first trial mapped a requirement onto `Rule:`.
[[adr-018]] constraint 1 corrects that: a requirement is a `Feature`,
and `Rule:` groups criteria under a genuine business rule. So `req` is
`Feature.Name`, and a criterion may additionally sit under a `Rule`
whose name is the rule it obeys. An extractor must walk **both** direct
`Feature` children and `Rule` children — a criterion can be either.

**Span correctness was verified by slicing the source at the computed
offsets and confirming the text returned is the intended scenario**,
in a document containing multi-byte characters.

---

## 3. The five integration details

### Gap 1 — `Location` is a START position only

`Location` is `{Line, Column}`. There is no end position, no byte
offset, and no length. This repo's `Span` (`cli/core/span`) is byte
offset plus byte length, with line and column derived.

*Resolution.* Compute it. Byte offset comes from converting
`(line, column)` against a line-start table for the file. End is derived
from the last `Step` — or the last `Examples` `TableBody` row when the
scenario is an Outline — and is therefore **line-granular, not exact**.
Good enough for a finding that points at a criterion; not exact enough
to replace a criterion by byte range without re-scanning.

### Gap 2 — columns count codepoints, which matches us exactly

Decisive test: a table row `| em — dash | SECOND |`. Gherkin reported
column **23** for `SECOND`; the 1-based codepoint index is **23** and
the byte index is **25**.

Gherkin counts Unicode codepoints from the start of the line — the same
convention `span.go` declares ("a column counts Unicode codepoints from
the start of its line, so an em dash advances it by one rather than by
its three UTF-8 bytes"). **No conversion layer is needed**, only the
codepoint→byte step already required for the offset.

### Gap 3 — node `Id`s are positional counters and are NOT stable

Inserting a scenario above an existing one changed the existing
scenario's `Id` from `4` to `7`. The `Id` is produced by the
caller-supplied generator in AST-build order.

*Resolution.* **Node identity comes from the `@AC-*` tag, never from
`Id`.** This matters most for the knowledge graph: keying nodes on `Id`
would re-identify every criterion on any edit above it. `Id` is useful
only within a single parse, for referring between nodes.

### Gap 4 — the synthetic wrapper is conditional

**Under the corrected mapping this does not arise for FRDs**, because
every block declares its own `Feature` and so is a complete Gherkin
document. It still matters for any other fenced Gherkin this framework
parses, and for FRDs written before the mapping was fixed.

A bare fragment (`Scenario:` with no `Feature:`) does not parse, so it
needs a synthetic `Feature:` line. But a block that *declares* its own
`Feature:` fails **with** the wrapper:

```
(3:1): expected: #TagLine, #RuleLine, #Comment, #Empty,
       got 'Feature: Spec Corpus Index'
```

*Resolution.* Detect whether the block declares a `Feature:` and set the
line offset accordingly — `md_line = fence_line + (parser_line - wrap)`
where `wrap` is 1 when wrapped and 0 when not. Getting this wrong
silently shifts every reported position by one line.

### Gap 5 — stripping the fence indentation shifts columns

A block indented two spaces reports `Scenario` at column 3; stripping
the indent before parsing reports column 1.

*Resolution.* **Do not strip.** Gherkin tolerates a uniformly indented
block, so passing the body through unmodified keeps columns true and
removes the need to add the stripped width back.

---

## 4. Two things worth knowing

**Comments are flat, not attached.** `#` comments live in
`GherkinDocument.Comments` as a document-level list with line numbers —
they are *not* hung off the scenario beneath them. Associating a comment
with a criterion is proximity logic we would own. Prefer
`Rule.Description` / `Scenario.Description`, which are attached.

**A `Feature` description holds multiple paragraphs and markdown.** Blank
lines are preserved; emphasis, `code`, `[[slug]]`, bullets and blockquotes
pass through as literal text. Two things do not survive: a line beginning
`#` becomes a comment and vanishes from the description, and a line
beginning with a Gherkin keyword terminates the description and becomes
structure. See `docs/plans/gherkin-feature-description-guidance.md`.

**Descriptions are AST-only.** `Pickles()` rows carry no description
field, so a store that loads only pickles loses every `Feature`,
`Rule` and `Scenario` description. Load them from the AST walk.

**`DocString` carries a media type.** A step can hold a fenced block with
`mediaType`, e.g. `"""markdown`, attached to that step and round-tripping
through JSON. This is a typed carrier for richer per-criterion rationale
if `Description` proves too thin.

---

## 5. Beyond the AST: `Pickles()`

`gherkin.Pickles(doc, uri, newId)` compiles the AST into one fully
resolved row per `Examples` row, with tags already inherited from
`Feature` and `Rule`. Two scenarios (one an Outline with two example
rows) produced three pickles. That is the denormalised fact table a
structured store loads, without writing the expansion ourselves.

Use the AST for authoring-time checks and for the index (it preserves
the general/illustrative distinction); use pickles for publication to a
store or to external tooling (it resolves it).

---

## References

- [[adr-018]] — the decision this record supports
- `cli/core/span` — the `Span` type gap 1 and gap 2 are reconciled against
- `docs/plans/gherkin-feature-description-guidance.md` — what belongs in a
  `Feature` description and what silently breaks one
- [cucumber/messages](https://github.com/cucumber/messages) — the wire
  schema the extracted records conform to

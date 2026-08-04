# Index 

## Table of Contents

- [Span](#span)
- [ParseError](#parseerror)
- [ExclusionReason](#exclusionreason)
- [IndexEntry](#indexentry)
- [IndexExclusion](#indexexclusion)
- [SpecIndex](#specindex)

## Span

A contiguous region of a document.

Byte offset and byte length are the authoritative extent, because spans
address regions *within* a line as well as whole sections — a malformed
`slug` value occupies part of a line, not the line itself. Line and
column are carried alongside for readability: they are what a human or
an agent reads in a report, and they are derived from the offset rather
than being an independent source of truth.

Lines and columns are one-based. A column counts Unicode codepoints from
the start of its line, matching the way this framework's conventions
measure width, so an em dash advances the column by one rather than by
its three UTF-8 bytes.

| Property | Type | Documentation |
|----------|------|---------------|
| offset | UInt | Byte offset of the first byte of the region, from the start of the document. |
| length | UInt | Length of the region in bytes. Zero is a valid, empty region — the content of `---\n---` is an empty span rather than an absent one. |
| start_line | UInt | One-based line on which the region starts. |
| start_column | UInt | One-based codepoint column at which the region starts. |
| end_line | UInt | One-based line on which the region ends. |
| end_column | UInt | One-based codepoint column one past the last codepoint of the region. |


## ParseError

A located failure to parse part of a document.

Recorded against the document it occurred in rather than raised, because
one unparsable file must not cost a repository the index of every other
file. The position is a `Span`, like every other position this framework
reports.

| Property | Type | Documentation |
|----------|------|---------------|
| message | String | What could not be parsed, stated so that the message needs no further lookup to act on. |
| position | [Span](#span) | Where in the document the failure occurred. |


## ExclusionReason

Why a candidate file beneath the spec root is not a corpus member.

Exclusions are recorded rather than dropped, so that a document which
*should* be a specification but is missing frontmatter is visible in the
index rather than absent from it — a judgement the tool cannot make for
the reader.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| NoFrontmatter | "no-frontmatter" | The file carries no frontmatter block at all. |
| UnknownType | "unknown-type" | The file declares a `type` the active profile does not define. |


## IndexEntry

One corpus member in the index: where it lives and what its frontmatter
declares.

The entry is deliberately additive. This sprint populates the path and
the frontmatter envelope; later features attach headings, extracted
record blocks and references to the same entry under the same
`schema_version`, and no property here forecloses modelling references
as their own collection later.

| Property | Type | Documentation |
|----------|------|---------------|
| path | String | The member's path, relative to the repository root and written with forward slashes on every platform. |
| frontmatter | [DocumentFrontmatter](frontmatter.md#documentfrontmatter)? | The member's parsed frontmatter.<br><br>Absent exactly when `parse_error` is present: a document whose frontmatter could not be parsed is still indexed, but it has no parsed envelope to carry. |
| parse_error | [ParseError](#parseerror)? | The failure that prevented this member's frontmatter from being parsed, with its position.<br><br>Present exactly when `frontmatter` is absent. A member carrying this property does not make the index a failure; the index is still emitted and still contains every other document. |


## IndexExclusion

One candidate file that is not a corpus member, and why.

| Property | Type | Documentation |
|----------|------|---------------|
| path | String | The excluded file's path, relative to the repository root and written with forward slashes on every platform. |
| reason | [ExclusionReason](#exclusionreason) | Why the file is not a member. |
| type | String? | The `type` value the file declared, reported when `reason` is `unknown-type` so that a reviewing agent can judge whether the profile or the document is wrong. |


## SpecIndex

The machine-readable description of a repository's specification corpus,
emitted to the file **specs.json** in the configured spec root.

This is the artifact CI, agents and a central architecture catalogue all
read instead of each re-deriving the corpus. It crosses a boundary, so
its shape is defined once here and generated into JSON Schema, Go and
Markdown rather than being assembled by hand in the emitter.

The model is additive by design: later features add properties to it and
to its entries under the same `schema_version`. `entries` and
`exclusions` are lists rather than maps precisely so that the emitter
owns their order and can make output byte-identical across platforms
without a change to this model.

| Property | Type | Documentation |
|----------|------|---------------|
| schema_version | String ~= `/^1\.0$/` | The version of this index's schema.<br><br>Pinned to `1.0` for as long as the shape only grows. A change to the shape of what is already here — as opposed to an addition alongside it — updates this value and the emitted index in the same change. |
| workspace | String | The workspace identifier from configuration, so that indexes emitted by several repositories can be merged into one graph without two documents that both declare `adr-001` colliding. |
| entries | Array<[IndexEntry](#indexentry)> | Every corpus member, in the order the emitter fixes. |
| exclusions | Array<[IndexExclusion](#indexexclusion)> | Every excluded candidate file, in the order the emitter fixes. |



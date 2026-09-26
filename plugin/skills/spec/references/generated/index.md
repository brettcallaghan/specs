# Index 

## Table of Contents

- [Span](#span)
- [ParseError](#parseerror)
- [ExclusionReason](#exclusionreason)
- [RecordBlock](#recordblock)
- [Heading](#heading)
- [IndexEntry](#indexentry)
- [IndexExclusion](#indexexclusion)
- [ReferenceBodyOrigin](#referencebodyorigin)
- [ReferenceFrontmatterOrigin](#referencefrontmatterorigin)
- [ReferenceOrigin](#referenceorigin)
- [ReferenceResolved](#referenceresolved)
- [ReferenceUnresolved](#referenceunresolved)
- [ReferenceNonMember](#referencenonmember)
- [ReferenceResolution](#referenceresolution)
- [AnchorResolved](#anchorresolved)
- [AnchorUnresolved](#anchorunresolved)
- [AnchorAmbiguous](#anchorambiguous)
- [AnchorUnverified](#anchorunverified)
- [AnchorResolution](#anchorresolution)
- [Reference](#reference)
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


## RecordBlock

A fenced block whose content the active profile registers as structured
data about an element declared inside a document.

A block is a record block only where the active profile registers its
info string as a tag; a profile that registers none yields a corpus with
no record blocks. The registration may name a schema the record is
validated against, so this shape names no particular tag and carries no
vocabulary of its own.

The record's content is not carried. Strata has no open value type, so
an arbitrary YAML mapping has no typed home in a generated model; a
command that needs a record reads it back through `position` and
decodes it into its own Go type, and `specs check` validates it. Adding
the content here later is an additive change.

| Property | Type | Documentation |
|----------|------|---------------|
| tag | String | The block's info string, which is the tag the active profile registered — `component` for a component record block. |
| position | [Span](#span) | The extent of the block as written, opening and closing fences included.<br><br>The fences are inside the span because promotion replaces the whole block with a link: a rewrite addressing only the content would leave the fences behind. The span is also how a command reads the record: the content between the fences is decoded on demand, into the Go type that command needs, rather than carried here. |


## Heading

One heading of a document body, with the extent of the content beneath
it.

The section span is what makes a place inside a document addressable. A
reader that wants the text under a heading reads it back through the
span, which is why no section content is carried here or anywhere else
in this model.

| Property | Type | Documentation |
|----------|------|---------------|
| text | String | The heading's text as authored, without the `#` markers that open the line. |
| level | UInt in 1..6 | The heading's level: 1 for `#`, through 6 for `######`.<br><br>__Constraint:__ Values must be ≥ 1 and ≤ 6. |
| position | [Span](#span) | Where the heading line itself is written. |
| section | [Span](#span) | The extent of the content beneath the heading: from the end of the heading line to the start of the next heading of the same or a higher level, and to the end of the document where there is none.<br><br>A heading that stands over nothing carries an empty span rather than an absent one, the same way an empty frontmatter block does. |
| record_blocks | Array<[RecordBlock](#recordblock)> | The record blocks attached to this heading: those whose extent falls within its section span and within no nested heading's, so a block belongs to the innermost heading that contains it. |


## IndexEntry

One corpus member in the index: where it lives and what its frontmatter
declares.

The entry is deliberately additive. It carries the path, the frontmatter
envelope, and the document's structure — its headings and the record
blocks attached to the document rather than to any heading — and later
features attach further properties under the same `schema_version`.

| Property | Type | Documentation |
|----------|------|---------------|
| path | String | The member's path, relative to the repository root and written with forward slashes on every platform. |
| frontmatter | [DocumentFrontmatter](frontmatter.md#documentfrontmatter)? | The member's parsed frontmatter.<br><br>Absent exactly when `parse_error` is present: a document whose frontmatter could not be parsed is still indexed, but it has no parsed envelope to carry. |
| parse_error | [ParseError](#parseerror)? | The failure that prevented this member's frontmatter from being parsed, with its position.<br><br>Present exactly when `frontmatter` is absent. A member carrying this property does not make the index a failure; the index is still emitted and still contains every other document. |
| headings | Array<[Heading](#heading)> | The document's headings, in document order.<br><br>Recorded for every member, including one whose frontmatter could not be parsed: a heading is markdown and does not depend on an envelope. |
| record_blocks | Array<[RecordBlock](#recordblock)> | The record blocks attached to the document rather than to any heading: those written before the first heading, and those in a document that has none.<br><br>A block inside a heading's section is carried by that heading instead, so every parsed block in a document appears exactly once. |


## IndexExclusion

One candidate file that is not a corpus member, and why.

| Property | Type | Documentation |
|----------|------|---------------|
| path | String | The excluded file's path, relative to the repository root and written with forward slashes on every platform. |
| reason | [ExclusionReason](#exclusionreason) | Why the file is not a member. |
| type | String? | The `type` value the file declared, reported when `reason` is `unknown-type` so that a reviewing agent can judge whether the profile or the document is wrong. |


## ReferenceBodyOrigin

A reference written in a document's body.

Carries nothing of its own. What a body reference adds beyond its
position is nothing, and its position is recorded on the reference
itself; this shape exists so that the origin is a value a consumer
branches on rather than a property it has to infer from the absence of
another.

*No properties defined.*


## ReferenceFrontmatterOrigin

A reference held by a frontmatter property the active profile registers
as reference-bearing.

| Property | Type | Documentation |
|----------|------|---------------|
| property | String | The name of the frontmatter property the reference was written in, so that a finding names the property without the document being reopened. |


## ReferenceOrigin

Where a reference was written.

The distinction is not decoration. A reference in prose and a reference
in a registered frontmatter property are reported differently — one
quotes the link, the other names the property that holds it — and they
are rewritten differently, so the origin belongs to the reference rather
than being inferred from where its span happens to fall.

**Discriminator:** kind (String)

| Tag | Type | Documentation |
|-----|------|---------------|
| "body" | Body | A reference parsed from the content after the frontmatter block. |
| "frontmatter-property" | FrontmatterProperty | A value held by a registered reference-bearing frontmatter property. |


## ReferenceResolved

A reference that reached a corpus member.

| Property | Type | Documentation |
|----------|------|---------------|
| target_path | String | The target's path, relative to the repository root and written with forward slashes on every platform.<br><br>The path rather than the slug, because the slug is already on the reference and a consumer following the edge wants the file. |


## ReferenceUnresolved

A reference whose slug matches no corpus member and no file the index
knows about at all.

Carries nothing: the slug that matched nothing is already on the
reference this resolution belongs to.

*No properties defined.*


## ReferenceNonMember

A reference whose slug matches a file the index recorded as excluded or
as a parse failure rather than as a corpus member.

Held apart from an ordinary unresolved reference because the reader's
next step differs: the document is right there, and what is wrong is its
frontmatter rather than the reference. Exactly one of `reason` and
`parse_error` is present — the two ways a file beneath the spec root
fails to become a member.

| Property | Type | Documentation |
|----------|------|---------------|
| path | String | The matched file's path, relative to the repository root and written with forward slashes on every platform. |
| reason | [ExclusionReason](#exclusionreason)? | Why the file is not a member, present when the index recorded it as an exclusion. |
| parse_error | [ParseError](#parseerror)? | The failure that prevented the file's frontmatter from being parsed, with its position, present when the index recorded it as a member whose frontmatter could not be read. |


## ReferenceResolution

What a reference reached.

A reference that reaches nothing is recorded with this state rather than
dropped, because indexing never fails: the index describes the corpus it
found, including the parts of it that are broken, and reporting a broken
reference is a check's job rather than the indexer's.

**Discriminator:** state (String)

| Tag | Type | Documentation |
|-----|------|---------------|
| "resolved" | Resolved | The slug matched a corpus member. |
| "unresolved" | Unresolved | The slug matched nothing in the corpus. |
| "non-member" | NonMember | The slug matched a file beneath the spec root that is not a member. |


## AnchorResolved

An anchor resolved to exactly one heading of its target document.

| Property | Type | Documentation |
|----------|------|---------------|
| heading | [AnchorMatch](finding.md#anchormatch) | The heading the anchor's identifier matched. |


## AnchorUnresolved

An anchor whose identifier matched no heading of its target document,
whose headings were available to check it against.

Held apart from `AnchorUnverified`: this is a considered "no", not an
absence of an answer — the target's headings were read and none of
them matched, which is what a rename-the-heading fix responds to.

*No properties defined.*


## AnchorAmbiguous

An anchor whose identifier matched more than one heading of its target
document.

| Property | Type | Documentation |
|----------|------|---------------|
| matches | Array<[AnchorMatch](finding.md#anchormatch)> | Every heading of the target document whose identifier matched, in document order. |


## AnchorUnverified

An anchor whose target document's headings could not be examined.

Distinct from `AnchorUnresolved`: the tool did not get to look, so it
says so rather than reporting a heading rename would fix — the same
distinction `FindingUnverified` draws for a check that could not run at
all.

| Property | Type | Documentation |
|----------|------|---------------|
| reason | String | Why the target's headings could not be examined, stated so the finding built from this needs no further lookup to act on. |


## AnchorResolution

What the anchor half of a reference resolved to.

Present on a `Reference` exactly when its anchor is present and its
target is known: the slug half reached a corpus member, or the anchor
names no slug and resolves against the document containing the
reference (AC-SPECANCH-002.6). Absent when the reference carries no
anchor, and absent when the slug half is present and did not resolve —
the parent feature already reports that reference as broken, and an
anchor finding on top of it would be a second finding for one mistake
(AC-SPECANCH-002.8).

**Discriminator:** state (String)

| Tag | Type | Documentation |
|-----|------|---------------|
| "resolved" | Resolved | Exactly one heading of the target matched the anchor's identifier. |
| "unresolved" | Unresolved | No heading of the target matched, though its headings were available. |
| "ambiguous" | Ambiguous | More than one heading of the target matched the anchor's identifier. |
| "unverified" | Unverified | The target document's headings could not be examined. |


## Reference

One occurrence of one document referencing another.

A reference is an edge, and it is recorded as a row of its own rather
than as a property of either document it connects. The facts that make
it actionable — where it was written, whether it came from prose or from
a named frontmatter property, and what it reached — belong to neither
endpoint, and both directions of the graph are derived by indexing this
one collection rather than stored twice.

One row per occurrence. A document that references the same target twice
contributes two rows carrying two spans, so a finding about either cites
the line the reader must open.

| Property | Type | Documentation |
|----------|------|---------------|
| path | String | The path of the document the reference occurs in, relative to the repository root and written with forward slashes on every platform. |
| text | String | The reference exactly as written, without the delimiters that enclose it — `adr-012`, or `adr-012#Decision`, or the raw value of a reference-bearing frontmatter property.<br><br>Recorded alongside the halves it splits into, because the text is what a reader sees and what a rewrite replaces, and neither is recoverable from the halves once letter case or spacing has been normalized. |
| slug | String | The slug half: everything before the first `#`, or the whole reference where it names no anchor.<br><br>A plain `String` rather than a `Slug`, deliberately. A reference is written by hand and may name something that is not a well-formed slug at all; constraining this property would make such a reference unrepresentable in the index, which is the one place it must be visible in order to be reported. |
| anchor | String? | The anchor half: everything after the first `#`, absent where the reference names no anchor.<br><br>Recorded and not interpreted: splitting on the first `#` is the whole of what extraction knows about anchors, which is why an anchored reference resolves on its slug rather than reporting as broken. What the anchor itself matched is carried separately, in `anchor_resolution`. |
| origin | [ReferenceOrigin](#referenceorigin) | Where the reference was written. |
| position | [Span](#span) | Where in the containing document the reference occurs, as a span rather than a bare line number, so a report and a rewrite both address the reference itself. |
| resolution | [ReferenceResolution](#referenceresolution) | What the reference reached. |
| anchor_resolution | [AnchorResolution](#anchorresolution)? | What the anchor half of this reference resolved to.<br><br>Present exactly when `anchor` is present and there is a target to check it against: `resolution` reached a corpus member, or the reference names no slug at all and resolves against the document named in `path`. Absent when the reference carries no anchor, and absent when it names a slug that did not resolve — `resolution` already reports that reference as broken. |


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
| references | Array<[Reference](#reference)> | Every reference any corpus member makes, in the order the emitter fixes: by the path the reference occurs in, then by its position within that document.<br><br>One top-level collection rather than an array on each entry. A reference is an edge; its attributes describe the edge and not either endpoint, both directions are derived from this one list when the index is read, and merging the indexes of two repositories into one graph is appending rows rather than rewriting entries.<br><br>Ordering by path and then by position is what makes a document's outbound references ordered by position in the document and the emitted file byte-identical across platforms, in one rule. |
| features | Array<[Feature](feature.md#feature)> | Every feature any corpus member declares, with its acceptance criteria nested beneath it, in the order the emitter fixes: by the path the feature is written in, then by its position within that file.<br><br>A top-level collection because a feature is addressed by its slug and not by the document that happens to hold it — nesting features under entries would make a feature's address depend on its location, which is what the slug exists to avoid. Criteria, by contrast, are nested inside their feature: a criterion belongs to exactly one, and its identifier carries that feature's code, so lookup by identifier is an index built over this tree rather than a reason to flatten it. |
| input_hash | String? | A hash of everything this index was built from, written as `sha256:` followed by lowercase hexadecimal.<br><br>This is what makes the emitted file double as a warm cache: a later invocation recomputes the hash of the corpus's current inputs and serves this index unchanged when the two agree, rather than rebuilding. It is a hash of **content** — every candidate file's repository-relative path and bytes, plus the active profile's vocabulary — and never of modification time, so a fresh checkout whose files all carry new timestamps and identical content is still served warm.<br><br>Absent on an index that was not built by walking a corpus. An index carrying no hash is never served warm; it is rebuilt. |



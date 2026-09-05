# Finding 

## Table of Contents

- [RuleId](#ruleid)
- [Severity](#severity)
- [FindingSchemaViolation](#findingschemaviolation)
- [FindingMissingProperty](#findingmissingproperty)
- [FindingSlugPattern](#findingslugpattern)
- [FindingDuplicateSlug](#findingduplicateslug)
- [FindingFrontmatterParse](#findingfrontmatterparse)
- [FindingCorpusEmpty](#findingcorpusempty)
- [FindingUnverified](#findingunverified)
- [FindingUnresolvedReference](#findingunresolvedreference)
- [FindingRetiredTarget](#findingretiredtarget)
- [FindingIncompletePairing](#findingincompletepairing)
- [FindingReferenceType](#findingreferencetype)
- [FindingUnmappedCitation](#findingunmappedcitation)
- [FindingUnresolvedCitation](#findingunresolvedcitation)
- [FindingGherkinParse](#findinggherkinparse)
- [FindingIdentifierPattern](#findingidentifierpattern)
- [FindingDuplicateIdentifier](#findingduplicateidentifier)
- [FindingPartiallyTaggedOutline](#findingpartiallytaggedoutline)
- [FindingOwningDocument](#findingowningdocument)
- [FindingUnresolvedAnchor](#findingunresolvedanchor)
- [FindingAmbiguousAnchor](#findingambiguousanchor)
- [AnchorMatch](#anchormatch)
- [FindingDetail](#findingdetail)
- [Finding](#finding)
- [CheckOutcome](#checkoutcome)
- [CheckResult](#checkresult)
- [FindingReport](#findingreport)

## Type Alias: RuleId

The identifier of one check, stable across releases.

Every finding names the rule that raised it, and the run report accounts
for every rule it knows about, so the identifier is the join between the
two: a caller naming a subset of checks to run names them by these
values, and reads the outcome of each back under the same value.

Identifiers are namespaced kebab-case — `slug/pattern-mismatch`,
`frontmatter/missing-property` — so the family a check belongs to is
legible without a lookup table. The shape is stated rather than
constrained because a Strata regex literal cannot contain a forward
slash.

**Definition:**

```strata
alias RuleId: String
```


## Severity

How badly a finding bears on a run.

The three values are the whole vocabulary, and the distinction that
matters most is the third: `unverified` is not a milder failure, it is
the absence of a result. A check the tool was asked to perform and could
not run reports `unverified` rather than passing silently, because a
consumer that cannot tell the two apart will read a missing check as a
clean one.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Error | "error" | A conformance failure. A run producing one exits non-zero. |
| Warning | "warning" | A probable problem that does not by itself fail a build. |
| Unverified | "unverified" | A check that was requested and could not run, so nothing is known. |


## FindingSchemaViolation

A frontmatter value that violates the schema its document type is
validated against.

The pointer and the constraint together make the finding actionable
without re-running the validator: the pointer says which property, and
the constraint says what it had to satisfy.

| Property | Type | Documentation |
|----------|------|---------------|
| pointer | String | The JSON Pointer at which the violation occurred, relative to the document's frontmatter. |
| constraint | String | The constraint the value failed, quoted as the schema states it. |
| value | String | The offending value, as raw source text rather than as a YAML parser reinterpreted it, so a date written unquoted is reported as it was written. |


## FindingMissingProperty

A property the document's type requires and the document does not
declare.

Modelled apart from the general schema violation because a missing
property is the most common defect by a wide margin, and a consumer
offering to insert one needs the property name as a value rather than as
a substring of a message.

| Property | Type | Documentation |
|----------|------|---------------|
| pointer | String | The JSON Pointer of the object that should carry the property. |
| property | String | The name of the property that is absent. |


## FindingSlugPattern

A slug that does not match the pattern the active profile supplies for
the document's type.

Both halves are carried so the finding is self-contained: an agent
fixing it needs neither the profile nor the document open.

| Property | Type | Documentation |
|----------|------|---------------|
| slug | String | The slug the document declares. |
| pattern | String | The pattern the active profile supplies for this document's type. |


## FindingDuplicateSlug

A slug declared by more than one corpus member.

Reported against each of the two documents, each naming the other, so
that checking either one surfaces the collision.

| Property | Type | Documentation |
|----------|------|---------------|
| slug | String | The slug both documents declare. |
| other_path | String | The other document's path, relative to the repository root and written with forward slashes on every platform. |


## FindingFrontmatterParse

A frontmatter block that could not be read as a document envelope.

Distinct from a schema violation, and reported by a different route: a
schema violation is found by validating a parsed envelope, while this is
what happens when there is no parsed envelope to validate. The document
declared itself a specification and got the declaration wrong badly
enough that nothing downstream can look at it.

Modelled separately because it is the one finding that suppresses all
the others for its document. A consumer must be able to tell "this
document passed every check" from "no check could run against this
document", and the rule identifier alone does not carry that: the other
checks report `passed` for a document they never examined unless
something says otherwise.

| Property | Type | Documentation |
|----------|------|---------------|
| message | String | What could not be read, quoted from the parser so the message needs no further lookup to act on. |


## FindingCorpusEmpty

A configured spec root beneath which discovery found no corpus member.

The only finding whose subject is the corpus rather than a document, and
the reason it carries a variant of its own: a consumer keying off the
kind must be able to tell "the run examined nothing" from a defect in
some file, and there is no file here to name. The root is carried
because it is the whole of what an operator acts on — a typo'd path, a
moved directory, an ignore rule that swallowed the corpus.

| Property | Type | Documentation |
|----------|------|---------------|
| root | String | The configured spec root, relative to the repository root, that holds no corpus member. |


## FindingUnverified

A check that was requested and could not run.

This is the structured half of the `unverified` severity: the reason
states what was missing, so a consumer can decide whether to install a
dependency, adjust configuration, or accept the gap.

| Property | Type | Documentation |
|----------|------|---------------|
| reason | String | Why the check could not run, stated so that it needs no further lookup to act on. |


## FindingUnresolvedReference

A reference that reached no corpus member.

The origin is carried because it decides how the defect reads: a `[[…]]`
in prose is quoted back, while a value in a registered frontmatter
property is reported by naming the property that holds it. The
non-member half is carried because a slug that matched a file the index
excluded or could not parse is a different problem with a different fix,
and telling the reader which one they have saves them a search for a
document that is right there.

| Property | Type | Documentation |
|----------|------|---------------|
| slug | String | The slug half of the reference, as written — everything before the first `#`, whose anchor half is not part of this defect. |
| origin | [ReferenceOrigin](index.md#referenceorigin) | Where the reference was written: the document's body, or a named reference-bearing frontmatter property. |
| non_member | [ReferenceNonMember](index.md#referencenonmember)? | The non-member file the slug matched, with the reason it is not a member.<br><br>Absent when the slug matched nothing at all, which is the ordinary case; present when the target exists beneath the spec root and is excluded or unparsable. |


## FindingRetiredTarget

A reference from a live document to one the active profile treats as
retired.

The target resolves — the reader reaches a real document that says it has
been replaced — so this is a warning rather than an error. The successor
is what makes it worth reporting at all: it is the one finding that tells
an author what to cite instead.

| Property | Type | Documentation |
|----------|------|---------------|
| slug | String | The retired target's slug. |
| successor | String? | The replacement the retired target names, read from the property the active profile registers for that purpose.<br><br>Absent where the target declares none. Not every retirement has a successor, and a finding that implied otherwise would send the reader looking for one. |


## FindingIncompletePairing

One half of a paired relationship the counterpart does not declare.

Reported against both documents, each carrying this same detail, so that
checking either one surfaces the half-built edge and the two copies
describe the relationship in the same direction rather than each from its
own point of view.

| Property | Type | Documentation |
|----------|------|---------------|
| property | String | The reference-bearing property the referring document declares. |
| counterpart | String | The property the referenced document must declare in return, and does not. |
| source_path | String | The referring document's path, relative to the repository root and written with forward slashes on every platform. |
| target_path | String | The referenced document's path, relative to the repository root and written with forward slashes on every platform. |


## FindingReferenceType

A typed reference whose target declares a document type other than the
one the active profile requires for that property.

A frontmatter schema can require that a property holds exactly one slug;
only the index knows what that slug resolves to, so the type half of the
constraint is checked here and reported with both types quoted.

| Property | Type | Documentation |
|----------|------|---------------|
| property | String | The reference-bearing property whose target's type is constrained. |
| slug | String | The slug the property names. |
| expected_type | String | The document `type` the active profile requires the target to declare. |
| actual_type | String | The `type` the target actually declares. |


## FindingUnmappedCitation

A citation whose slug names a corpus document the caller's link map does
not cover.

Distinct from `FindingUnresolvedCitation`: the slug does resolve to a
real corpus document, so the defect is a gap in the map rather than a
bad citation, and the fix is to add the missing entry rather than to fix
the prose. The map's path is carried alongside the slug and the citing
document so the finding is actionable without the caller having the map
open.

| Property | Type | Documentation |
|----------|------|---------------|
| slug | String | The slug the citation names. |
| citing_path | String | The citing document's path, relative to the repository root and written with forward slashes on every platform. |
| map_path | String | The link map's path, relative to the repository root and written with forward slashes on every platform. |


## FindingUnresolvedCitation

A citation whose slug names no corpus document at all.

Distinct from `FindingUnmappedCitation`: nothing in the corpus resolves
the slug, so the citation is as likely to be illustrative prose — a
template teaching citation syntax — as a mistyped reference, and the two
cannot be told apart by the tool. Carrying only the slug and the citing
document (no map path) reflects that a link map, present or absent, is
not what decided this outcome.

| Property | Type | Documentation |
|----------|------|---------------|
| slug | String | The slug the citation names. |
| citing_path | String | The citing document's path, relative to the repository root and written with forward slashes on every platform. |


## FindingGherkinParse

A Gherkin source that could not be parsed.

Held apart from a frontmatter parse failure because the two are found by
different readers and fixed in different places: a Gherkin source sits
inside a document whose envelope parsed perfectly well, and only the
feature it declares is lost. The parser's own expectation is carried so
that the author is told what was admissible at that position rather than
only that something was not.

| Property | Type | Documentation |
|----------|------|---------------|
| expected | String | What the parser expected at the offending position, quoted from the parser so the message needs no further lookup to act on. |


## FindingIdentifierPattern

A record whose identifier tag is absent or does not match the pattern
registered for it.

One shape for both defects, because the fix is the same act — write a
conforming identifier — and the two are told apart by the rule that
raised the finding. The pattern is carried in either case: an author who
wrote no identifier needs to know the form of the one to write just as
much as an author who wrote a wrong one.

| Property | Type | Documentation |
|----------|------|---------------|
| identifier | String? | The identifier as written, absent where the record carries no identifier tag at all. |
| pattern | String | The pattern the active profile registers for identifiers of this kind. |


## FindingDuplicateIdentifier

An identifier declared by more than one record in the corpus.

Reported against each of the two records, each naming the other, so that
checking either one surfaces the collision. The other occurrence's span
is carried alongside its path because a file holds many records and an
identifier is not a heading: without the span the reader is told which
file to open and left to search it.

| Property | Type | Documentation |
|----------|------|---------------|
| identifier | String | The identifier both records declare. |
| other_path | String | The other occurrence's path, relative to the repository root and written with forward slashes on every platform. |
| other_position | [Span](index.md#span) | Where in that file the other occurrence is written. |


## FindingPartiallyTaggedOutline

A scenario outline whose examples blocks are inconsistently tagged: some
carry an identifier and others do not.

The defect is the inconsistency, not the several blocks — an outline
whose blocks are all tagged is several criteria, and one whose blocks are
all untagged is a single criterion written in named groups. Both are
admissible, and the mixture is what no reader can resolve.

The untagged blocks are carried as spans so that a report points at the
blocks to fix rather than at the outline containing them.

| Property | Type | Documentation |
|----------|------|---------------|
| untagged | Array<[Span](index.md#span)> | Where each untagged examples block is written, in the order the blocks occur. |


## FindingOwningDocument

A feature whose owning-document tag is absent, names a document other
than the one the feature is written in, or names no corpus member at all.

The three defects share a shape because they share a question — which
document owns this feature — and the rule that raised the finding says
which of them occurred. Both slugs are optional because each defect
leaves a different half unknown: a feature carrying no tag has no slug to
quote, and a standalone feature file is no document and so has none of
its own to disagree with.

| Property | Type | Documentation |
|----------|------|---------------|
| slug | String? | The slug the feature's owning-document tag names, as written, absent where the feature carries no such tag. |
| document_slug | String? | The slug of the document the feature is written in, absent where the feature is a standalone feature file rather than a fenced block. |


## FindingUnresolvedAnchor

An anchor that names no heading of its target document.

Recorded once the target's slug has resolved and its headings are
known — a reference whose slug does not resolve at all is the parent
feature's own unresolved-reference finding, and this variant adds
nothing to it (AC-SPECANCH-002.1, AC-SPECANCH-002.8). The anchor is
carried as authored, matching how the index records it, so the finding
needs no further lookup to quote the text a fix has to edit.

| Property | Type | Documentation |
|----------|------|---------------|
| anchor | String | The anchor as authored, without normalization. |
| target_path | String | The target document's path, relative to the repository root and written with forward slashes on every platform. |


## FindingAmbiguousAnchor

An anchor whose identifier is the identifier of more than one heading of
its target document.

Every matching heading is carried, not merely the count: the reader's
next step is to open one of them and rename it, and a finding naming
only "two matches" would send them back to the document to find what
this check already found (AC-SPECANCH-002.5).

| Property | Type | Documentation |
|----------|------|---------------|
| anchor | String | The anchor as authored, without normalization. |
| target_path | String | The target document's path, relative to the repository root and written with forward slashes on every platform. |
| matches | Array<[AnchorMatch](#anchormatch)> | Every heading of the target document whose identifier matches the anchor's, in document order. |


## AnchorMatch

One heading matched against an ambiguous anchor.

Held apart from the full `Heading` shape because a match names only what
the reader needs to find the right one — its text and where it is
written — and carries none of a heading's own section content or nested
record blocks.

| Property | Type | Documentation |
|----------|------|---------------|
| text | String | The heading's text as authored, without the `#` markers that open the line. |
| position | [Span](index.md#span) | Where the heading line itself is written. |


## FindingDetail

The structured detail of one finding, selected by the kind of defect.

Every finding carries one. The message on the finding is derivable from
this value, so a consumer branches on structure rather than parsing
prose, and a renderer that wants to offer a fix reads the fields it needs
as values.

The union is open to extension in the sense that matters: a later check
adds a variant without changing the envelope that carries it, and a
consumer that does not recognise a variant still has the rule
identifier, the severity, the path, the span and the message.

**Discriminator:** kind (String)

| Tag | Type | Documentation |
|-----|------|---------------|
| "schema-violation" | SchemaViolation | A frontmatter value violating the schema for its document type. |
| "missing-property" | MissingProperty | A required frontmatter property the document does not declare. |
| "slug-pattern" | SlugPattern | A slug not matching the pattern registered for its type. |
| "duplicate-slug" | DuplicateSlug | A slug declared by two corpus members. |
| "frontmatter-parse" | FrontmatterParse | A frontmatter block that could not be read as a document envelope. |
| "corpus-empty" | CorpusEmpty | A spec root beneath which discovery found no corpus member. |
| "unverified" | Unverified | A requested check that could not run. |
| "unresolved-reference" | UnresolvedReference | A reference whose slug matches no corpus member. |
| "retired-target" | RetiredTarget | A reference from a live document to a retired one. |
| "incomplete-pairing" | IncompletePairing | A paired relationship declared from one end only. |
| "reference-type" | ReferenceType | A typed reference whose target declares a different document type. |
| "unmapped-citation" | UnmappedCitation | A citation whose slug names a corpus document the link map does not cover. |
| "unresolved-citation" | UnresolvedCitation | A citation whose slug names no corpus document. |
| "gherkin-parse" | GherkinParse | A Gherkin source that could not be parsed. |
| "identifier-pattern" | IdentifierPattern | An identifier tag that is absent or malformed. |
| "duplicate-identifier" | DuplicateIdentifier | An identifier declared by two records in the corpus. |
| "partially-tagged-outline" | PartiallyTaggedOutline | A scenario outline whose examples blocks are inconsistently tagged. |
| "owning-document" | OwningDocument | A feature whose owning document is unstated, mismatched or unresolved. |
| "unresolved-anchor" | UnresolvedAnchor | An anchor matching no heading of its target document. |
| "ambiguous-anchor" | AmbiguousAnchor | An anchor matching more than one heading of its target document. |


## Finding

One reported conformance result.

Every check in this tool reports through this shape, so a consumer
renders one thing and a check added later is readable by a consumer
written before it. The five common properties are what make a diagnostic
self-contained — the rule that raised it, how badly it bears on the run,
which file, where in that file, and what is wrong — and `detail` carries
the same information again in structured form for a consumer that acts
on it rather than printing it.

| Property | Type | Documentation |
|----------|------|---------------|
| rule | [RuleId](#ruleid) | The check that raised this finding. |
| severity | [Severity](#severity) | How badly this finding bears on the run. |
| path | String | The file the finding is against, relative to the repository root and written with forward slashes on every platform. |
| position | [Span](index.md#span) | Where in the file the finding is, as a span rather than a bare line number, so a report can point at a value inside a line. |
| message | String | What is wrong, stating the violated constraint and the value that violated it, so the finding needs no further lookup to act on. |
| detail | [FindingDetail](#findingdetail) | The structured form of the same defect. |


## CheckOutcome

What became of one check in a run.

A run reports an outcome for every check it knows about, including the
ones it was not asked to perform, because the alternative — omitting
them — leaves a consumer unable to tell a check that passed from a check
that never ran.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| Passed | "passed" | Requested, ran, and raised no finding. |
| Failed | "failed" | Requested, ran, and raised at least one finding. |
| Unverified | "unverified" | Requested and could not run. Nothing is known about it. |
| NotRequested | "not-requested" | Not requested for this run. Nothing is known about it. |


## CheckResult

The outcome of one check in a run.

| Property | Type | Documentation |
|----------|------|---------------|
| rule | [RuleId](#ruleid) | The check this result is for. |
| outcome | [CheckOutcome](#checkoutcome) | What became of it. |
| reason | String? | Why the check could not run, present exactly when `outcome` is `unverified`. |


## FindingReport

Everything one check run produced.

The two collections answer different questions and neither substitutes
for the other. `findings` is what is wrong; `checks` is what was looked
at. A report carrying no findings is only good news if every check the
caller cares about is accounted for in `checks` as having passed.

Both are lists rather than maps so that the emitter owns their order and
can make output byte-identical across platforms without a change to this
model.

| Property | Type | Documentation |
|----------|------|---------------|
| checks | Array<[CheckResult](#checkresult)> | Every check the run knows about, in the order the emitter fixes, each with what became of it. |
| findings | Array<[Finding](#finding)> | Every finding the run produced, in the order the emitter fixes. |



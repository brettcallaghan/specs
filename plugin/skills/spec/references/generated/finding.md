# Finding 

## Table of Contents

- [RuleId](#ruleid)
- [Severity](#severity)
- [FindingSchemaViolation](#findingschemaviolation)
- [FindingMissingProperty](#findingmissingproperty)
- [FindingSlugPattern](#findingslugpattern)
- [FindingDuplicateSlug](#findingduplicateslug)
- [FindingUnverified](#findingunverified)
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


## FindingUnverified

A check that was requested and could not run.

This is the structured half of the `unverified` severity: the reason
states what was missing, so a consumer can decide whether to install a
dependency, adjust configuration, or accept the gap.

| Property | Type | Documentation |
|----------|------|---------------|
| reason | String | Why the check could not run, stated so that it needs no further lookup to act on. |


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
| "unverified" | Unverified | A requested check that could not run. |


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



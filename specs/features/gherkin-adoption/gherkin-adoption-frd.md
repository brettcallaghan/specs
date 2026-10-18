---
title: "Gherkin Features and Acceptance Criteria"
description: The CLI parses Gherkin from FRD blocks and standalone feature
  files, validates it with diagnostics an agent can act on, extracts features
  and acceptance criteria into the index, and answers queries by tag
slug: frd-gherkin-adoption
type: frd
status: approved
sources: []
---

# Gherkin Features and Acceptance Criteria

## Overview

[[adr-021]] adopts Gherkin as the notation for features and acceptance
criteria. Nothing implements it. Today an acceptance criterion is a
markdown bullet that no tool reads: the CLI indexes a document's
frontmatter and location, and the criteria inside it are prose to
everything downstream.

This feature makes features and criteria first-class records. The CLI
parses Gherkin wherever the framework admits it and validates what the
Gherkin parser cannot: that every criterion is identified, uniquely,
and says something. It extracts each feature and criterion into the
index as a typed record and answers direct questions about them by tag.

The payoff is addressability. An agent implementing a work order asks
for one criterion instead of reading an FRD. A reviewer asks which
criteria a branch changes instead of reading a diff. An
organization publishes the index to a central store and gets features
and criteria as rows rather than paragraphs — the reason [[adr-021]]
chose a parsed notation over prose.

**Out of scope for this feature.** Each is deferred deliberately:

- **Executing scenarios.** Nothing here runs a scenario against a
  system under test. These are specification records; no runner is
  adopted, and the `@level-*` tag records intent rather than driving
  execution.
- **Migrating existing criteria.** Rewriting the prose criteria in the
  three existing FRDs is separate work, tracked on its own. It is a
  translation, not a compatibility problem: extraction reads Gherkin
  only ([[adr-021]] constraint 9), so until a document is translated it
  contributes no feature records rather than being read a second way.
- **Reading prose criteria.** No extractor recognises a prose
  criterion. An earlier revision required one, and it was withdrawn:
  a second reader is transitional machinery for a transition of three
  documents, and it would outlive the transition in the model and in
  every consumer.
- **Judging criterion quality.** Whether a criterion is atomic,
  testable, or well-phrased stays with a reviewing agent. This feature
  reports structural defects only.
- **Metadata on a standalone `.feature` file.** A feature file carries
  none: it is a feature source, and its metadata is its owning
  document's. Whether it should ever be a document in its own right,
  with a status of its own, is a later question — nothing here
  depends on the answer.
- **Rendering or diagram generation** from feature records.
- **Test-to-criterion coverage reconciliation.** Enumerating criteria is
  the prerequisite and is delivered here; comparing them against test
  citations is a follow-on feature.

## Terminology

- **Gherkin source** — a region of text the CLI parses as Gherkin:
  either a fenced ` ```gherkin ` block inside a markdown document, or a
  standalone `.feature` file.
- **Feature record** — the extracted representation of one Gherkin
  `Feature`: its identifier, name, description, owning document, source
  span, and the criteria beneath it.
- **Criterion record** — the extracted representation of one acceptance
  criterion: a `Scenario`, a `Scenario Outline`, or one tagged
  `Examples` block of an outline.
- **Tag vocabulary** — the set of tags the framework assigns meaning to,
  as distinct from tags a project adds for its own purposes.
- **Content hash** — a digest over a criterion's name and text, used to
  detect that a criterion changed rather than moved.
- **Feature source** — a `.feature` file, which holds features but is not
  itself a document: it carries no frontmatter, and its owning document
  is the one its `@frd-<slug>` tag names.
- **Owning document** — the document a feature belongs to. For a fenced
  block it is the document the block is written in; a feature file, being
  no document, names its owner with a tag.
- **Feature code** — the short uppercase code that groups features by
  topic, domain or capability. Features share a code; a code and a number
  together identify one.
- **Feature identifier** — a feature's durable address, composed of a
  prefix, its group code and its number: `FEAT-GHK-001`.

## Requirements

### FEAT-GHK-001: Admissible Gherkin Sources

**Acceptance Criteria:**

```gherkin
@FEAT-GHK-001
Feature: Admissible Gherkin Sources

  As an authoring team
  I want to write a feature either inside an FRD or as its own file
  So that features needing prose context and features needing none both have a home

  Two forms are admissible and produce identical records. A fenced block
  is the default, because most features benefit from the overview,
  terminology and references around them. A standalone file suits a
  feature that needs none, and is the form external tooling ingests.

  A standalone file lives beside the documents of the feature it belongs
  to, in specs/features/<feature-slug>/, which is where FRDs live today.
  Discovery therefore needs no new location to search.

  A standalone file carries no frontmatter. It is a feature source rather
  than a document: its owning document is the one its tag names, and its
  metadata is that document's.

  Ownership is derived wherever it can be. A fenced block is written
  inside the document that owns it, so nothing needs to be declared; only
  a feature file, having no document around it, names its owner.

  Executing scenarios is out of scope; these are specification records.

  @AC-GHK-001.1
  Scenario: A fenced gherkin block in a markdown document is a Gherkin source
    Given a markdown document containing a fenced block tagged as gherkin
    When the CLI reads that document
    Then the CLI parses the block contents as a Gherkin document

  @AC-GHK-001.2
  Scenario: A .feature file beneath the spec root is a Gherkin source
    Given a file whose name ends in .feature beneath the spec root
    When the CLI discovers the corpus
    Then the CLI parses its whole contents as a Gherkin document
    And the CLI attributes its features to the document their owning-document tag names

  @AC-GHK-001.3
  Scenario: Both forms yield the same records
    Given one feature expressed as a fenced block and the same feature as a standalone file
    When the CLI extracts each
    Then the extracted feature and criterion records are equal except for their source location

  @AC-GHK-001.4
  Scenario: A block that declares no Feature is still parsed
    Given a fenced gherkin block whose contents begin with a Rule or a Scenario
    When the CLI parses that block
    Then the CLI supplies a synthetic Feature header before parsing
    And every reported position maps to the line in the containing document

  @AC-GHK-001.5 @negative
  Scenario: Positions never point at the synthetic header
    When the CLI reports any finding or span within a fenced block
    Then the CLI does not report a position that the synthetic header occupies

  @AC-GHK-001.6
  Scenario: A standalone feature file sits with the documents of its feature
    Given a feature owned by an FRD with the slug frd-checkout
    When that feature is written as a standalone file
    Then the file sits in the same directory as the FRD that owns it

  @AC-GHK-001.7
  Scenario: A fenced block is owned by the document it is written in
    Given a fenced gherkin block in a document whose frontmatter declares a slug
    When the CLI extracts the features in that block
    Then the CLI attributes them to that document
    And the CLI requires no tag on them to say so
```

*Note:* AC-GHK-001.2 requires a change to [[frd-spec-index]],
whose AC-SPEC-001.6 admits only files ending in `.md` as corpus
candidates. A `.feature` file is not a corpus member — it carries no
frontmatter and is not a document — so what that criterion needs is a
second discovery pass for feature sources, not an exception to corpus
membership.

### FEAT-GHK-002: Tag Vocabulary

**Acceptance Criteria:**

```gherkin
@FEAT-GHK-002
Feature: Tag Vocabulary

  As an agent addressing a feature or a criterion
  I want every one of them to carry an identifying tag
  So that I can fetch one by tag instead of searching for it

  Tags are the addressing scheme, so they are required rather than
  encouraged. The Gherkin parser accepts any non-whitespace tag and
  attaches no meaning to any of them; every rule below is the CLI's.

  A project may use tags outside this vocabulary. The CLI reports each
  as a warning rather than an error: a misspelt qualifier otherwise drops
  its meaning silently, while a project's own conventions and a runner's
  tags still coexist with the framework's without failing the run.

  @AC-GHK-002.1
  Scenario Outline: The framework tag vocabulary
    Given a Gherkin source
    When the CLI extracts its records
    Then the CLI reads "<tag>" on a "<element>" as "<meaning>"

    Examples: document ownership
      | tag         | element                               | meaning                            |
      | @frd-<slug> | Feature in a standalone .feature file | the document that owns the feature |

    Examples: identity
      | tag                  | element  | meaning                        |
      | @<PREFIX>-<CODE>-NNN | Feature  | the feature's identifier       |
      | @AC-<CODE>-NNN.N     | Scenario | the criterion's identifier     |

    Examples: qualifiers
      | tag        | element  | meaning                              |
      | @negative  | Scenario | the criterion states a prohibition   |
      | @manual    | Scenario | the criterion is verified by hand    |
      | @level-*   | Scenario | the intended verification level      |

  @AC-GHK-002.2
  Scenario: Every feature carries an identifier tag
    Given a Feature in any Gherkin source
    When the CLI validates it
    Then the CLI requires exactly one feature identifier tag

  @AC-GHK-002.10
  Scenario: A feature in a standalone file names its owning document
    Given a Feature in a standalone .feature file
    When the CLI validates it
    Then the CLI requires exactly one owning-document tag

  @AC-GHK-002.11 @negative
  Scenario: An owning-document tag in a fenced block may not contradict its document
    Given a Feature in a fenced block carrying an owning-document tag that names another document
    When the CLI validates it
    Then the CLI reports a finding naming both documents

  @AC-GHK-002.12 @negative
  Scenario: An owning-document tag naming no corpus member is reported
    Given a Feature in a standalone .feature file whose owning-document tag names a slug no corpus member declares
    When the CLI validates it
    Then the CLI reports a finding naming the slug

  @AC-GHK-002.3
  Scenario Outline: An acceptance criterion identifier is well formed
    Given a criterion tagged "<tag>"
    When the CLI validates it
    Then the CLI reports it as "<verdict>"

    Examples:
      | tag             | verdict     |
      | @AC-CHK-003.1   | well formed |
      | @AC-SPEC-001.12 | well formed |
      | @AC-chk-003.1   | malformed   |
      | @AC-CHK-3.1     | malformed   |
      | @AC-CHK-003     | malformed   |

  @AC-GHK-002.4
  Scenario: An identifier is unique across the corpus
    Given two criteria in the corpus carrying the same identifier
    When the CLI validates the corpus
    Then the CLI reports a finding against each, naming the other's location

  @AC-GHK-002.5
  Scenario: A criterion carrying more than one identifier is reported
    Given a Scenario carrying two tags that both match the identifier pattern
    When the CLI validates it
    Then the CLI reports a finding naming both

  @AC-GHK-002.6
  Scenario: Identifiers on an outline live on its Examples blocks
    Given a Scenario Outline whose Examples blocks each carry an identifier
    When the CLI extracts its records
    Then the CLI emits one criterion record per tagged Examples block

  @AC-GHK-002.7
  Scenario: Partially tagged Examples blocks are reported
    Given a Scenario Outline where some Examples blocks carry an identifier and others do not
    When the CLI validates it
    Then the CLI reports a finding naming the untagged blocks

  @AC-GHK-002.9
  Scenario: An outline with untagged Examples blocks is one criterion
    Given a Scenario Outline carrying one identifier and several untagged Examples blocks
    When the CLI extracts its records
    Then the CLI emits one criterion record whose example rows are those of every block

  @AC-GHK-002.8
  Scenario: A tag outside the vocabulary is reported as a warning
    Given a Feature or a Scenario carrying a tag the framework assigns no meaning
    When the CLI validates it
    Then the CLI reports a finding naming that tag with severity "warning"
```

*Rationale for AC-GHK-002.7 and AC-GHK-002.9:* the dangerous case is an
author who means several criteria and tags only the scenario, so all but
the first cease to exist in a document that parses cleanly. That intent
is not machine-detectable: several `Examples` blocks under one
identifier are equally a single criterion presented in named groups,
which this document itself uses. So the check is **partial** tagging,
which is unambiguous, and the authoring rule lives in the Gherkin guide
where a human or agent can apply judgment. An earlier draft made the
multi-block case itself an error, and this document would have failed
it.

### FEAT-GHK-003: Feature Identity and Grouping

**Acceptance Criteria:**

```gherkin
@FEAT-GHK-003
Feature: Feature Identity and Grouping

  As an architecture owner
  I want a feature to carry a durable identifier and a grouping code
  So that a feature is addressable independently of the document that holds it

  A feature is identified by its code together with its number —
  FEAT-GHK-001, FEAT-SPEC-009. The identifier is durable: moving a feature
  between documents, or promoting a fenced block into a standalone file,
  does not change it.

  Features share a code. SPEC covers eleven requirements of the
  spec-index feature, which is what makes a code a grouping rather than
  an identity. Grouping by code rather than by a path-shaped slug is
  forced: a framework slug admits only lowercase alphanumerics and
  hyphens, so a hyphenated slug cannot be split back into segments
  unambiguously.

  @AC-GHK-003.1
  Scenario: A feature identifier survives its feature moving
    Given a feature whose identifier is unchanged
    When the feature moves from a fenced block into a standalone file
    Then a query for that identifier returns the feature from its new location

  @AC-GHK-003.2
  Scenario: A feature record carries its identifier and its code
    When the CLI extracts a feature
    Then the record carries the identifier and the code that identifier groups it under

  @AC-GHK-003.3
  Scenario: A code selects every feature grouped under it
    Given several features sharing a code
    When a query names that code
    Then the CLI returns every one of them

  @AC-GHK-003.4
  Scenario: A feature identifier is unique across the corpus
    Given two features declaring the same identifier
    When the CLI validates the corpus
    Then the CLI reports a finding against each, naming the other's location

  @AC-GHK-003.5 @negative
  Scenario: Sharing a code is not a defect
    Given several features declaring the same code with different numbers
    When the CLI validates the corpus
    Then the CLI reports no finding for the shared code

  @AC-GHK-003.6
  Scenario: A criterion identifier derives from its feature's identifier
    Given a criterion beneath a feature
    When the CLI validates it
    Then the CLI requires the identifier's code and number to be that feature's
```

*Rationale:* an earlier draft of this requirement made a feature's code
unique and gave the feature a separate slug. Both were wrong, and the
corpus says so: `AC-SPEC-001.1` through `AC-SPEC-011.5` share the code
`SPEC` across eleven requirements. A code groups; a code and a number
together identify. The separate slug is dropped because the identifier
already is the durable address, and a second one would be a second thing
to keep in step.

### FEAT-GHK-004: References Inside Gherkin

**Acceptance Criteria:**

```gherkin
@FEAT-GHK-004
Feature: References Inside Gherkin

  As a reviewer relying on reference integrity
  I want a wikilink inside Gherkin to count as a reference
  So that a feature cannot cite a document the index does not know about

  Gherkin never sees a wikilink. Measured: `[[adr-012]]` parses as
  ordinary text in a feature description, a scenario name, a step, a rule
  name and an examples cell, and raises nothing anywhere. That makes the
  question purely this framework's, and the answer is that a reference is
  a reference wherever it is written.

  @AC-GHK-004.1
  Scenario Outline: A wikilink anywhere in a Gherkin source is a reference
    Given a Gherkin source carrying a wikilink in its "<position>"
    When the CLI extracts references
    Then the CLI records a reference with the span of the wikilink

    Examples:
      | position            |
      | feature description |
      | scenario name       |
      | step text           |
      | rule name           |
      | examples cell       |

  @AC-GHK-004.2
  Scenario: A reference inside a fenced block is positioned in its document
    Given a wikilink inside a fenced gherkin block
    When the CLI records the reference
    Then the span addresses the wikilink's position in the containing document

  @AC-GHK-004.3
  Scenario: References inside Gherkin resolve like any other
    Given a wikilink inside a Gherkin source naming no corpus member
    When the CLI checks references
    Then the CLI reports it unresolved as it would in prose

  @AC-GHK-004.4 @negative
  Scenario: An anchor is not resolved against Gherkin structure
    Given a wikilink inside a Gherkin source naming an anchor
    When the CLI records the reference
    Then the CLI splits the slug from the anchor as it does elsewhere
    And the CLI does not attempt to resolve the anchor against a Feature, Rule or Scenario name
```

*Rationale for AC-GHK-004.4:* a scenario name is not a heading, and
anchors resolve against headings. Recording the anchor without resolving
it keeps this feature from silently inventing a second anchor namespace
that [[frd-spec-reference-integrity]] does not define.

### FEAT-GHK-005: Feature and Criterion Records

**Acceptance Criteria:**

```gherkin
@FEAT-GHK-005
Feature: Feature and Criterion Records

  As a consumer of the index
  I want features and criteria as addressable records
  So that a structured store or a knowledge graph loads them without re-parsing prose

  The criterion is the atomic unit. A criterion's body is carried as text
  rather than decomposed into typed steps and example rows, because
  decomposition would model one notation: those parts are recoverable from
  Gherkin and from almost nothing else, so they would sit empty for every
  prose criterion and could not be filled from another notation by
  parsing. Nothing this framework does today operates below the level of a
  whole criterion.

  Criteria nest inside their feature. A criterion belongs to exactly one,
  and its identifier carries that feature's code, so lookup by identifier
  is an index over the tree rather than a reason to flatten it. A
  criterion record still names where it lives — its document's path and
  its feature's identifier — so a lookup of one criterion locates it
  without a second query.

  Identity never comes from a parser node id. Those are positional
  counters that shift when anything above them is edited.

  The shape these criteria describe is defined in schemas/feature.strata
  and transpiled; this requirement states what it must carry, not how it
  is written down.

  @AC-GHK-005.1
  Scenario: A feature record carries its identity, prose and provenance
    When the CLI extracts a feature
    Then the record carries its identifier, its code, its owning document, its name, its description, its source form and its span

  @AC-GHK-005.2
  Scenario: A criterion record carries its identity, its text and its provenance
    When the CLI extracts a criterion
    Then the record carries its identifier, its name, its text, its tags, its qualifiers, its description and its span
    And the record carries the path of its owning document and the identifier of the feature it belongs to

  @AC-GHK-005.3
  Scenario: A criterion's body is recorded as text, not decomposed
    Given a criterion written as a scenario with steps and an examples table
    When the CLI extracts it
    Then the record carries the steps and the table verbatim as the criterion's text
    And the record does not decompose them into separate step or row properties

  @AC-GHK-005.4
  Scenario: Criteria are nested within their feature
    When the CLI extracts a feature
    Then the feature record carries its criteria in the order they are written

  @AC-GHK-005.5
  Scenario: A criterion record carries a content hash
    When the CLI extracts a criterion
    Then the record carries a hash over the criterion's name and text

  @AC-GHK-005.6 @negative
  Scenario: The content hash excludes position and parser identity
    Given a criterion whose content is unchanged
    When another criterion is inserted above it and the corpus is re-indexed
    Then the criterion's content hash does not change

  @AC-GHK-005.7
  Scenario: Editing a criterion changes its hash
    Given a criterion whose text is edited
    When the corpus is re-indexed
    Then the criterion's content hash changes

  @AC-GHK-005.8
  Scenario: Records reach the index under its declared schema
    When the CLI builds the index
    Then feature records are present in the index under its schema version
```

*Rationale:* an earlier draft of this requirement decomposed a criterion
into typed steps and keyed example rows, on the argument that a parsed
notation yields fields where prose yields one string. It does — but no
requirement in this document consumes those fields, and they are fillable
from Gherkin alone. Carrying them would put a Gherkin-shaped model at the
centre of a framework that treats the notation as replaceable. The index
schema is additive, so decomposition can be introduced when a consumer
needs it; withdrawing it later could not be.

### FEAT-GHK-006: Validation and Diagnostics

**Acceptance Criteria:**

```gherkin
@FEAT-GHK-006
Feature: Validation and Diagnostics

  As an authoring agent
  I want a diagnostic that names what is wrong and where
  So that I can correct the source without loading the standard that defines the rule

  The Gherkin parser validates its own grammar and nothing else. It has
  no concept of an acceptance criterion, so every check here is the
  CLI's. Several of the mistakes below produce a document that parses
  cleanly, which is precisely why they need reporting.

  Judging whether a criterion is well written stays with a reviewing
  agent; this feature reports structural defects only.

  @AC-GHK-006.1
  Scenario: A grammar error is reported at its position in the containing document
    Given a Gherkin source containing a malformed step keyword
    When the CLI validates it
    Then the CLI reports a finding at the line and column of the offending text in the containing document
    And the message names what the parser expected

  @AC-GHK-006.2
  Scenario Outline: Structural defects the parser accepts are reported
    Given a Gherkin source containing "<defect>"
    When the CLI validates it
    Then the CLI reports a finding

    Examples: identity
      | defect                                                   |
      | a Scenario carrying no acceptance-criterion identifier   |
      | a malformed acceptance-criterion identifier              |
      | an identifier duplicated elsewhere in the corpus         |
      | a Feature carrying no identifier tag                     |

    Examples: content
      | defect                                                   |
      | a Scenario with no steps                                 |
      | normative prose in a Scenario description                |
      | partially tagged Examples blocks on one outline          |

  @AC-GHK-006.3
  Scenario: Every finding carries a stable rule identifier
    When the CLI reports a Gherkin finding
    Then the finding carries a rule identifier, a severity, a path, a span and a message

  @AC-GHK-006.4 @negative
  Scenario: A Feature description is not reported as normative prose
    Given a Feature description carrying a user story and a design note
    When the CLI validates it
    Then the CLI does not report that description as an untagged criterion

  @AC-GHK-006.5
  Scenario: A diagnostic states the offending value and the expected form
    Given a criterion whose identifier does not match the required pattern
    When the CLI reports it
    Then the message quotes the identifier as written and the pattern it must match

  @AC-GHK-006.6
  Scenario Outline: A structural defect fails the run
    Given a source containing "<condition>"
    When the CLI validates it
    Then the CLI reports it with severity "error"

    Examples:
      | condition                                     |
      | an untagged criterion                         |
      | a duplicated identifier                       |
      | a criterion with no steps                     |

  @AC-GHK-006.7
  Scenario: Validating one document reports that document's findings
    When the CLI validates a single document containing Gherkin
    Then the CLI reports the findings for that document
    And the CLI resolves identifier uniqueness against the whole corpus
```

Every Gherkin rule is an error but one: a tag outside the vocabulary is
a warning (AC-GHK-002.8), because a project may use tags of its own. A
Feature with no description and a criterion with no `@level-*` tag are
not findings at all. The description is the author's to write or omit,
and [[gherkin-guide]] makes the level tag optional. A warning on either
would fire on most of the corpus and teach readers to ignore the
severity.

### FEAT-GHK-007: Queries by Tag

**Acceptance Criteria:**

```gherkin
@FEAT-GHK-007
Feature: Queries by Tag

  As an agent working a single unit of work
  I want to fetch one feature or one criterion by tag
  So that I read what I need instead of the document that contains it

  These queries are what make the extracted records worth building. They
  answer from the index rather than by re-reading the corpus, so they
  inherit the latency budgets the index already carries.

  Every one of them is a query in the sense of [[cli-exit-status-standards]]:
  it exits zero whenever it answered, and an answer is a list of records
  in one shape per command, whether one record matched, several or none.
  Identifiers match regardless of case, as slugs do.

  @AC-GHK-007.1
  Scenario: A feature is fetched by its identifier
    When a query names a feature identifier
    Then the CLI returns a list holding that feature, in the view the query selects, its record including its criteria under the full option
    And the answer has the shape of every other feature query

  @AC-GHK-007.2
  Scenario: An acceptance criterion is fetched by its identifier
    When a query names an acceptance-criterion identifier
    Then the CLI returns a list holding that criterion's record
    And the answer has the shape of every other criterion query

  @AC-GHK-007.3
  Scenario Outline: Features and criteria are listed by selection
    When a query lists "<selection>"
    Then the CLI returns "<result>"

    Examples:
      | selection                     | result                                  |
      | all features                  | every feature in the corpus             |
      | features carrying a code      | every feature grouped under that code   |
      | criteria of a feature         | every criterion beneath that feature    |
      | criteria carrying a qualifier | every criterion tagged with it          |

  @AC-GHK-007.4
  Scenario Outline: A query for an undeclared subject answers that it is not declared
    When a query names a "<subject>" that the corpus does not declare
    Then the CLI returns an empty list
    And the answer carries a property stating that the subject is not declared
    And the CLI exits 0

    Examples:
      | subject                         |
      | acceptance-criterion identifier |
      | feature identifier              |
      | feature code                    |

  @AC-GHK-007.5
  Scenario: Query output is available as JSON
    When a query is invoked with the JSON option
    Then the CLI writes the matching records as JSON to standard output

  @AC-GHK-007.6
  Scenario: Queries answer from the index
    Given an index whose inputs are unchanged since it was built
    When a feature or criterion query other than a change listing runs
    Then the CLI answers from the persisted index without re-walking the corpus

  @AC-GHK-007.7
  Scenario Outline: Identifiers match regardless of case
    Given the corpus declares "<declared>"
    When a query names "<named>"
    Then the CLI returns the record declared as "<declared>"

    Examples:
      | declared     | named        |
      | FEAT-GHK-001 | feat-ghk-001 |
      | AC-GHK-001.1 | ac-ghk-001.1 |
      | GHK          | ghk          |

  @AC-GHK-007.8 @negative
  Scenario: A qualifier outside the vocabulary is an invocation error
    When a query selects criteria by a qualifier the tag vocabulary does not register
    Then the CLI writes the value and every registered qualifier to standard error
    And the CLI exits 2

  @AC-GHK-007.9
  Scenario: A qualifier may be written with its leading at sign
    When a query selects criteria by "@negative"
    Then the CLI returns the same answer as a query selecting by "negative"

  @AC-GHK-007.10
  Scenario: A feature query answers in the compact view by default
    When a feature query runs without the full option
    Then the CLI returns each feature as its "id", its "code", its "name", its "path" and its "criteria_count", the number of criteria it holds, and nothing else
    And the compact answer holds every feature the full answer holds, in the same order

  @AC-GHK-007.11
  Scenario: The full option returns every feature's whole record
    When a feature query runs with the full option
    Then the CLI returns each feature's record as AC-GHK-005.1 and AC-GHK-005.4 describe it, its criteria included

  @AC-GHK-007.12 @negative
  Scenario: The full option is refused with the emit option
    When a feature query is asked to emit a feature file with the full option
    Then the CLI writes nothing to standard output
    And the CLI exits 2
```

*Rationale:* AC-GHK-007.6 binds these queries to the warm-index
behaviour [[frd-spec-index]] already specifies, rather than
restating a latency budget this feature does not own.

*Rationale for AC-GHK-007.10 to AC-GHK-007.12:* a feature record
carries its description and every criterion's text, so listing every
feature in full answers "which features exist?" with most of the corpus's
normative text. The compact view is the default for the reason the
reference query's is, and is a named view as [[cli-output-standards]]
defines one: every feature stays and every kept value is whole.
`criteria_count` is derived — the full record carries the criteria
themselves, not their count — and is named apart from the full view's
`criteria` so one key never holds two types. An emitted feature file is a
document payload rather than an answer, so neither view applies to it,
and the combination is refused rather than one option ignored.

*Rationale for AC-GHK-007.4 and AC-GHK-007.8:* an identifier or a code
is a fact about the corpus, which an author may write tomorrow; a
qualifier is drawn from a registered set no document can extend. So an
undeclared identifier is an answer and an unregistered qualifier is a
mistake in the question, as [[cli-exit-status-standards]] draws the line.
The registered qualifiers are those AC-GHK-002.1 lists: `negative`,
`manual` and `level-*`.

### FEAT-GHK-008: Local Change Detection

**Acceptance Criteria:**

```gherkin
@FEAT-GHK-008
Feature: Local Change Detection

  As a reviewer or an agent preparing a change
  I want to know which acceptance criteria changed
  So that I can attend to the criteria themselves rather than to a document diff

  Content hashes make this a comparison of records rather than of text,
  so a criterion that moved is not reported and a criterion whose
  wording changed is.

  The comparison is against a git revision, never the persisted index,
  which any query may rebuild. The base is the merge-base of HEAD and a
  named revision — the repository's default branch unless --since names
  a branch, a tag or a commit — so the answer is what a pull request from
  this branch would show, plus whatever is edited but not yet committed.

  A change listing is a query in the sense of [[cli-exit-status-standards]]:
  it exits 0 whatever it lists, unless the caller opts into failing on a
  non-empty answer with --exit-code.

  This requirement is the lowest priority in this document and may be
  delivered after the others.

  @AC-GHK-008.1
  Scenario: Criteria whose content changed are listed
    Given a base revision of the repository
    When the CLI compares the current corpus against the corpus at that base
    Then the CLI lists every criterion whose content hash differs

  @AC-GHK-008.2 @negative
  Scenario: A criterion that only moved is not listed as changed
    Given a criterion whose content is unchanged since the base but whose position moved
    When the CLI compares the current corpus against the base
    Then the CLI does not list that criterion

  @AC-GHK-008.3
  Scenario Outline: Added and removed criteria are distinguished from changed ones
    Given a criterion that is "<state>" relative to the base
    When the CLI compares
    Then the CLI reports it as "<classification>"

    Examples:
      | state                            | classification |
      | present in both, content differs | changed        |
      | present now, absent before       | added          |
      | absent now, present before       | removed        |

  @AC-GHK-008.4
  Scenario: The default base is the repository's default branch
    Given a query for changed criteria that names no revision
    When the CLI compares
    Then the base is the merge-base of HEAD and the repository's default branch

  @AC-GHK-008.5
  Scenario Outline: A named revision replaces the default
    Given a query for changed criteria whose --since option names a "<revision>"
    When the CLI compares
    Then the base is the merge-base of HEAD and that revision

    Examples:
      | revision |
      | branch   |
      | tag      |
      | commit   |

  @AC-GHK-008.6
  Scenario: Uncommitted edits are compared
    Given a criterion edited in the working tree and not committed
    When the CLI compares the current corpus against the base
    Then the CLI lists that criterion as changed

  @AC-GHK-008.7 @negative
  Scenario: Rebuilding the index does not change the answer
    Given a criterion edited since the base
    And the index rebuilt after that edit
    When the CLI compares the current corpus against the base
    Then the CLI still lists that criterion as changed

  @AC-GHK-008.8
  Scenario Outline: A change listing fails only when asked to
    Given a comparison whose answer is "<answer>"
    When the CLI lists changed criteria "<option>"
    Then the CLI exits "<status>"

    Examples:
      | answer    | option              | status |
      | empty     | without --exit-code | 0      |
      | non-empty | without --exit-code | 0      |
      | empty     | with --exit-code    | 0      |
      | non-empty | with --exit-code    | 1      |

  @AC-GHK-008.9 @negative
  Scenario Outline: A comparison that cannot be made answers nothing
    Given a query for changed criteria run "<condition>"
    When the CLI attempts the comparison
    Then the CLI writes nothing to standard output
    And the CLI reports a "<class>" naming the condition and exits "<status>"

    Examples:
      | condition                                                  | class             | status |
      | outside a git repository                                   | environment error | 3      |
      | naming no revision where no default branch can be resolved | environment error | 3      |
      | where HEAD and the base share no merge-base                | environment error | 3      |
      | naming a revision git does not know                        | invocation error  | 2      |

  @AC-GHK-008.10
  Scenario Outline: The default branch is the first candidate that exists
    Given a repository holding "<references>"
    When a query for changed criteria names no revision
    Then the base is the merge-base of HEAD and "<default>"
    And the candidates are tried in the order: the target of origin/HEAD, origin/main, origin/master, main, master

    Examples:
      | references                            | default       |
      | origin/HEAD naming origin/trunk, main | origin/trunk  |
      | main, origin/main, master             | origin/main   |
      | origin/main, master                   | origin/main   |
      | master, origin/master                 | origin/master |
      | origin/master                         | origin/master |
      | main, master                          | main          |
      | master                                | master        |

  @AC-GHK-008.11
  Scenario: A change listing names its base and locates each listed criterion
    When the CLI lists changed criteria
    Then the answer holds the base, as the revision compared against and the merge-base commit, followed by the listed criteria
    And each listed criterion carries its identifier, its "name", its classification and the path of its document, the base's path for a removed criterion
    And the "name" is the current name for an added or changed criterion and the base's name for a removed one, and is omitted where the criterion has none
    And the listed criteria are ordered by identifier

  @AC-GHK-008.12 @negative
  Scenario Outline: A comparison option without --changed is an invocation error
    When a criterion query is given "<option>" without --changed
    Then the CLI writes nothing to standard output
    And the CLI exits 2

    Examples:
      | option       |
      | --since main |
      | --exit-code  |
      | --full       |

  @AC-GHK-008.13 @negative
  Scenario: A spec root outside the git work tree cannot be compared
    Given a configured spec root that lies outside the git work tree
    When a query for changed criteria runs
    Then the CLI writes nothing to standard output
    And the CLI reports an environment error naming the spec root and the work tree, and exits 3

  @AC-GHK-008.14
  Scenario: A local default branch behind its remote-tracking branch is reported
    Given a query for changed criteria that names no revision
    And the default branch chosen is a remote-tracking branch
    And a local branch of the same name exists whose commit is a strict ancestor of the remote-tracking branch's
    When the CLI compares
    Then the CLI writes a warning to standard error naming both branches and their short commits
    And the CLI writes to standard output exactly the answer it writes where no such local branch exists

  @AC-GHK-008.15 @negative
  Scenario Outline: No staleness warning is written where nothing is behind
    Given a query for changed criteria "<condition>"
    When the CLI compares
    Then the CLI writes no warning to standard error

    Examples:
      | condition                                                              |
      | whose local default branch is at the remote-tracking branch's commit   |
      | whose local default branch has commits the remote-tracking one lacks   |
      | whose --since option names a revision                                  |

  @AC-GHK-008.16
  Scenario Outline: The full option carries each listed criterion's records
    Given a change listing invoked with the full option
    When the CLI lists a criterion classified "<classification>"
    Then the listed criterion carries "<records>"
    And a "record" is the criterion's record exactly as a query naming its identifier returns it, and a "base_record" is the same record taken from the base

    Examples:
      | classification | records                |
      | added          | record                 |
      | changed        | record and base_record |
      | removed        | base_record            |
```

*Rationale for AC-GHK-008.10, AC-GHK-008.14 and AC-GHK-008.15:* a local
branch is only as current as the last time someone updated it, and a
cloud session or a CI checkout typically never does. Its remote-tracking
branch is refreshed by every fetch. A base taken from a
stale local `main` silently attributes every change merged since to the
branch under review, so the remote-tracking branch is preferred and the
local one is used only where no remote-tracking candidate exists. The
warning names the stale local branch because the caller who reads it may
be relying on that branch elsewhere; it goes to standard error so the
answer a caller parses is the same whether or not it fires. A local
branch that is ahead of or diverged from its remote-tracking branch is
not "behind", so it draws no warning.

*Rationale for AC-GHK-008.11 and AC-GHK-008.16:* a reviewer reading a
change listing needs to know what each listed criterion is about without
one further query per identifier, so every entry carries its name. The
full option carries the records themselves — both sides of a change, so
the reviewer can compare them — and is valid only with a change listing,
since a criterion query already returns the whole record.

### FEAT-GHK-009: Publication

**Acceptance Criteria:**

```gherkin
@FEAT-GHK-009
Feature: Publication

  As an organization publishing specifications outward
  I want conforming feature files emitted from whatever form the source took
  So that external product and quality tooling ingests them without a second authoring format

  Emission is what makes the fenced form the default rather than a
  compromise: a feature authored inside an FRD is still publishable.

  Extraction reads Gherkin only. The documents still carrying prose
  criteria are translated rather than read by a second extractor, so
  nothing here is a compatibility mode ([[adr-021]] constraint 9).

  @AC-GHK-009.1
  Scenario: A conforming feature file is emitted for a feature
    Given a feature authored as a fenced block
    When the CLI is asked to emit it
    Then the CLI writes a valid Gherkin feature file carrying the same tags, description and criteria

  @AC-GHK-009.2
  Scenario: Emitted output round-trips
    Given an emitted feature file
    When the CLI parses it
    Then the extracted records equal those extracted from the source, except for their source location

  @AC-GHK-009.4
  Scenario: The notation of a record is stated
    When the CLI extracts a criterion
    Then the record states which notation it was written in
```

## References

- [[adr-021]] — the decision this feature implements, and the fourteen
  constraints its requirements realize
- [[adr-002]] — EARS, which continues to govern every normative
  statement that is not a feature or an acceptance criterion
- [[frd-spec-index]] — the index, the finding record, the warm
  index these queries answer from, and AC-SPEC-001.6, whose `.md`-only
  corpus membership FEAT-GHK-001 leaves standing
- [[cli-exit-status-standards]] — the query and gate statuses these
  queries and the change listing follow
- [[cli-output-standards]] — the shape their answers take
- [[frd-spec-reference-integrity]] — the reference extraction that
  FEAT-GHK-004 extends over Gherkin sources
- `schemas/feature.strata` — the `Feature` and `Criterion` record shapes
  FEAT-GHK-005 specifies, and the `features` collection they occupy on
  `SpecIndex`. The Strata definition is the contract; the JSON Schema, Go
  types and generated documentation are transpiled from it
- [[gherkin-guide]] — how to write
  conforming Gherkin; the authoring counterpart to this document's rules
- `plugin/skills/spec/references/ac-traceability-standards.md` — the
  citation format that consumes these identifiers
- `docs/research/gherkin/extraction-findings.md` — the measurements
  behind the comment header, the absence of block comments, and the
  content hash

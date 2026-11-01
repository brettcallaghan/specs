# Rules 

## Table of Contents

- [RuleId](#ruleid)

## RuleId

The identifier of one check, stable across releases.

Every finding names the rule that raised it, and `specs check --only` selects
checks by these values. Each identifier is namespaced kebab-case, so its
family is legible from the name. Each value's description starts with the
severity it reports at.

**Value Type:** String

| Name | Value | Documentation |
|------|-------|---------------|
| FrontmatterParse | "frontmatter/parse" | Error. The frontmatter can't be read as a document envelope, so no other check examined the document. Fix it first. The usual cause is an enumerated value, such as `status`, outside its vocabulary. |
| FrontmatterSchema | "frontmatter/schema" | Error. A frontmatter value violates its document type's schema. The finding names the property, the value, and the constraint. |
| FrontmatterMissingProperty | "frontmatter/missing-property" | Error. The document doesn't declare a property its type requires. |
| FrontmatterInlineComment | "frontmatter/inline-comment" | Warning. A YAML comment follows a frontmatter value on the same line, so the value ends at the `#`. Quote the value or move the comment. |
| SlugPatternMismatch | "slug/pattern-mismatch" | Error. The slug doesn't match the pattern the profile registers for the document's type. |
| SlugDuplicate | "slug/duplicate" | Error. More than one document declares the slug. The finding names the other document. |
| CorpusEmpty | "corpus/empty" | Error. The spec root holds no corpus member, so a clean report would mean nothing was examined. |
| ReferenceUnresolved | "reference/unresolved" | Error. A `[[slug]]` in the body, or a slug in a reference-bearing frontmatter property, matches no corpus member. |
| ReferenceUnresolvedAnchor | "reference/unresolved-anchor" | Error when an anchor matches no heading of the cited document. Unverified when the cited document's headings aren't recorded, so the check couldn't look. |
| ReferenceAmbiguousAnchor | "reference/ambiguous-anchor" | Error. An anchor matches more than one heading of the cited document. Qualify it with the parent heading, as in `[[slug#Parent#Heading]]`. |
| ReferenceRetiredTarget | "reference/retired-target" | Warning. A live document cites a retired one. The finding quotes the target's `superseded_by`, which is what to cite instead. |
| ReferencePairingIncomplete | "reference/pairing-incomplete" | Error. A paired relationship is declared from one end, and the target doesn't declare the counterpart. |
| ReferenceTypeMismatch | "reference/type-mismatch" | Error. A typed reference resolves to a document of a different `type`, such as a Feature Blueprint's `frd` naming something that isn't an FRD. |
| LinksUnmapped | "links/unmapped" | Error. `specs links` was given a link map that doesn't cover a slug the corpus resolves. |
| LinksUnresolved | "links/unresolved" | Warning. `specs links` found a citation whose slug names no corpus document, with or without a link map. |
| GherkinParse | "gherkin/parse" | Error. A Gherkin source can't be parsed. |
| GherkinFeatureUntagged | "gherkin/feature-untagged" | Error. A `Feature` carries no identifier tag. |
| GherkinCriterionUntagged | "gherkin/criterion-untagged" | Error. A `Scenario` carries no identifier tag. |
| GherkinIdentifierMalformed | "gherkin/identifier-malformed" | Error. An identifier tag doesn't match the pattern registered for its kind. |
| GherkinIdentifierMultiple | "gherkin/identifier-multiple" | Error. A `Feature` or `Scenario` carries more than one identifier tag. |
| GherkinIdentifierDuplicate | "gherkin/identifier-duplicate" | Error. Two records in the corpus declare the same feature or criterion identifier. |
| GherkinCriterionCodeMismatch | "gherkin/criterion-code-mismatch" | Error. A criterion's identifier doesn't carry its feature's code and number. |
| GherkinOwnerMissing | "gherkin/owner-missing" | Error. A feature in a standalone `.feature` file carries no owning-document tag. |
| GherkinOwnerMismatch | "gherkin/owner-mismatch" | Error. A feature in a fenced block carries an owning-document tag that names a different document. |
| GherkinOwnerUnresolved | "gherkin/owner-unresolved" | Error. An owning-document tag names no corpus member. |
| GherkinTagUnregistered | "gherkin/tag-unregistered" | Warning. A tag has no meaning in the profile's tag vocabulary, such as a misspelled `@negatve`. The finding names the registered qualifiers. |
| GherkinCriterionEmpty | "gherkin/criterion-empty" | Error. A criterion has no steps. |
| GherkinScenarioProse | "gherkin/scenario-prose" | Warning. A `Scenario` has a description, which usually means an untagged criterion was written as prose. |
| GherkinExamplesPartiallyTagged | "gherkin/examples-partially-tagged" | Error. Some `Examples` blocks of a `Scenario Outline` carry an identifier and others don't. |
| C4Parse | "c4/parse" | Error when a C4 diagram statement is outside the subset the parser reads. Unverified when a `mermaid` block's diagram type isn't read at all, such as `C4Dynamic`. |
| C4UndeclaredNode | "c4/undeclared-node" | Error. A non-external node resolves to no element in its document's scope and isn't a passive store. |
| C4AmbiguousNode | "c4/ambiguous-node" | Error. A node's label matches the entity names of more than one element in scope. The finding names each. |
| C4Bidirectional | "c4/bidirectional" | Error. A diagram uses `BiRel`. An arrow states one dependency, in one direction. |
| C4DuplicateEdge | "c4/duplicate-edge" | Error. One diagram draws two relationships with the same source and target, after `Rel_Back` is normalized. |
| C4DrawingConflict | "c4/drawing-conflict" | Warning. One relationship is drawn in two places with a different technology, description, or tags. The finding names the other drawing. |
| C4InteractionConflict | "c4/interaction-conflict" | Error. A relationship is tagged both `sync` and `async`. |
| C4ContractUnresolved | "c4/contract-unresolved" | Error. A `contract:` tag names a slug that no Contract declares. |
| C4ContractMismatch | "c4/contract-mismatch" | Warning. A `contract:` tag names a Contract that the source element doesn't consume or the target element doesn't provide. |
| C4TagSeparator | "c4/tag-separator" | Warning. A relationship's tags are joined with a comma instead of `+`. |
| ConfigSchema | "config/schema" | Error. A `specs.toml` value violates the configuration schema. |
| ConfigMissingProperty | "config/missing-property" | Error. A `specs.toml` table is missing a property the configuration schema requires. |
| ConfigUndeclaredKey | "config/undeclared-key" | Error. `specs.toml` declares a key the configuration schema doesn't define, which is usually a typo. |
| ConfigUnknownProfile | "config/unknown-profile" | Error. `[profile].name` names no registered profile. |
| ConfigInvalidRegistration | "config/invalid-registration" | Error. The active profile's own registration is defective. The finding sits at `[profile].name` and names the offending key. |
| LinkmapInvalid | "linkmap/invalid" | Error. The link map can't be read or parsed, including a slug repeated as a top-level key. |
| LinkmapMissingPath | "linkmap/missing-path" | Error. A link-map record declares no `path`. |
| DocumentUnreadable | "document/unreadable" | Error. The document named on the command line can't be opened. |



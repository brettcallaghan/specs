---
title: Glossary
description: The internal terms this project's specs and docs use, each with
  its one meaning and the words to avoid for it
slug: guide-glossary
type: guide
status: active
---

# Glossary

Each term below has one meaning in this project's specs and docs. Use it only
in that meaning. When a document needs an internal term that isn't here, add
it here rather than defining it in the document, or use plain words. The
Avoid column lists words not to use for that term's meaning; it doesn't ban
them when they carry their ordinary meaning.

| Term | Meaning | Avoid |
|------|---------|-------|
| anchor | The `#heading` part of a wikilink, such as `Design` in `[[component-corpus#Design]]`. | |
| architecture vocabulary | What a profile registers about how its documents declare C4 elements: the element types and their C4 levels, the properties that carry an element's name, system, composed modules, and consumed contracts, the contract type, and each type's Components heading and record block tag. | architecture registration, when meaning the vocabulary rather than the act |
| base revision | The git revision whose slugs `specs check` compares a document against, to find references to a slug the document no longer declares. The Spec CLI reads `HEAD`. | comparison base; "base" alone |
| C4 drawing document | A corpus document that contains a C4 diagram and whose type the profile registers as declaring a C4 element. The C4 checks resolve the nodes and relationship statements of its diagrams. | drawing document, alone |
| C4 element | A system, container, component, or module, declared by its own blueprint or by a record block, or drawn on a C4 diagram. | element, alone |
| C4 model | The C4 elements a corpus declares and the relationships among them, as data. A diagram is a view drawn from it. | rendering, for building the model |
| C4 relationship statement | A relationship statement on a C4 diagram: `Rel`, `BiRel`, or one of their directional forms. A directional statement is resolved when each endpoint names exactly one C4 element, and unresolved otherwise; a `BiRel` states no direction, so it is neither. A relationship in the C4 model groups the resolved statements that share a source, target, and label; an unresolved statement never becomes a relationship. | drawing |
| Components section | The section of a Container or Module Blueprint that declares each component it composes. | roster |
| corpus | The spec documents that the index includes. A document the index includes is a corpus document. | member, corpus member |
| criterion qualifier | A criterion property or verification level that a Gherkin tag can set, as the generated schema declares it. | |
| external C4 element | A person or an external system drawn on a C4 diagram: `Person` and every `_Ext` form. The corpus doesn't declare it, so the C4 checks don't resolve it. | external node |
| Gherkin qualifier tag | A Gherkin tag that the profile gives a meaning on a criterion: it sets a criterion qualifier, such as prohibition or manual verification. | qualifier tag, alone |
| heading path | Headings from a document's top level down to one section, joined by `#`, such as `Design#Invariants`. It selects that section. | path, when it could be read as a file path |
| host blueprint | The blueprint of the C4 element that composes another C4 element. A component's host is its container or module. | parent, unless the relationship is the C4 one |
| index | The model of the corpus that the Spec CLI builds from every document. `specs/specs.json` caches it between runs. | |
| paired reference | Two references that two documents must make to each other, one from each end, such as a property each declares naming the other. The `reference/pairing-incomplete` rule reports a pair with only one end. | paired relationship, pairing, mirror, bidirectional |
| passive store | A database or queue drawn on a C4 diagram (a `Db` or `Queue` form) that names no C4 element in scope: a data store that the drawn C4 element uses and that has no blueprint. The C4 checks don't report it as undeclared. | |
| profile | The configuration that tells the Spec CLI which document types, front matter properties, and rules a corpus uses. This repository's is the Specs Profile. | |
| record block | A fenced YAML block, tagged with a type the profile registers, that declares an entity inside another document and holds its `entity_name`, `slug`, and `description`. The Specs Profile registers the `component` tag, for a component declared in its host's Components section. | |
| recovered slug | A slug read from front matter that failed to parse, so the document still counts as declaring it. | recovered, alone |
| reference | A citation of another document by slug: a wikilink in the body, such as `[[slug]]` or `[[slug#heading]]`, or the value of a front matter property the profile registers as a reference property. Resolving it finds the document that declares the slug. | link, for a citation the tool resolves |
| reference property | A front matter property that the profile registers as holding references: its value cites other documents by slug. | reference-bearing property |
| span | A contiguous region of a document: a byte offset and a byte length, with the line and column derived from the offset. | extent, byte range |

## References

- [[guide-c4]]: the C4 levels and C4 element kinds
- [[document-relationship-standards]]: relationships and paired references

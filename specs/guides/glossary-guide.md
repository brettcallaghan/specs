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
| comparison base | The git revision whose slugs `specs check` compares a document against, to find references to a slug the document no longer declares. The Spec CLI reads `HEAD`. | "base" alone |
| Components section | The section of a Container or Module Blueprint that declares each component it composes. | roster |
| corpus | The spec documents that the index includes. A document the index includes is a corpus document. | member, corpus member |
| C4 model | The elements a corpus declares and the relationships among them, as data. A diagram is a view drawn from it. | rendering, for building the model |
| element | A system, container, component, or module, declared by its own blueprint or by a record block. | |
| external node | A C4 diagram node for a person or an external system: `Person` and every `_Ext` form. It names no element of the corpus and isn't resolved. | |
| heading path | Headings from a document's top level down to one section, joined by `/`, such as `Design/Invariants`. It selects that section. | path, when it could be read as a file path |
| host blueprint | The blueprint of the element that composes another element. A component's host is its container or module. | parent, unless the relationship is the C4 one |
| index | The model of the corpus that the Spec CLI builds from every document. `specs/specs.json` caches it between runs. | |
| paired relationship | A relationship that both documents must declare, each naming the other. | pairing, mirror |
| passive store | A `Db` or `Queue` diagram node that names no element in scope: a data store the drawn element uses, which has no blueprint. | |
| profile | The configuration that tells the Spec CLI which document types, front matter properties, and rules a corpus uses. This repository's is the Specs Profile. | |
| record block | A fenced block in a host blueprint that declares an element with no blueprint of its own, holding its `entity_name`, `slug`, and `description`. | |
| recovered slug | A slug read from front matter that failed to parse, so the document still counts as declaring it. | recovered, alone |
| reference | A wikilink in a document, such as `[[slug]]` or `[[slug#heading]]`. Resolving it finds the document that declares the slug. | link, for a wikilink the tool resolves |
| span | A contiguous region of a document: a byte offset and a byte length, with the line and column derived from the offset. | extent, byte range |

## References

- [[guide-c4]]: the C4 levels and element kinds
- [[document-relationship-standards]]: relationships and paired declarations

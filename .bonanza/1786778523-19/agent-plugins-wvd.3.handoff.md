# agent-plugins-wvd.3 handoff

Built the fenced-block half of Markdown Structure: a new `cli/core/markdown`
package with one exported entry point, `Fenced(body []byte, tags []string)
[]Block`. It walks `body` once with the goldmark parser already a module
dependency (via References), matches each `*ast.FencedCodeBlock` whose info
string equals one of `tags`, and returns each match in document order as a
`Block{Tag, Content, Outer, Inner}` — `Outer` covers both fence lines in full,
`Inner` covers the content alone, both computed through `cli/core/span`'s
`Of`.

The one piece of real design work: goldmark's public AST does not expose
whether a fenced block closed on a real closing line or simply absorbed the
rest of the document at EOF — both produce an identical node shape. I detect
this myself: if the block's content (or, for empty content, the line right
after the opening fence) runs to `len(body)` with nothing left for a closing
line, the fence is treated as unterminated and yields no block. This is
covered by `TestAnUnterminatedFenceYieldsNoBlock` with three sub-cases
(content never closed, opening line only with/without a trailing newline).

Implicit decisions made and not overridden by the bead:
- Exported surface: `markdown.Fenced(body []byte, tags []string) []markdown.Block`,
  plain positional args rather than an Options struct (only two inputs, no
  Path field — the package stays document-local like `span` and
  `frontmatter`, matching the Anchors list's silence on a frontmatter
  dependency).
- Tag matching is an exact match against goldmark's own already-trimmed info
  string (no first-token/language-only matching).
- Outer covers whole lines (leading indentation and all), not just the
  delimiter runs — chosen because the blueprint's Behavior bullet says
  "including both fence lines" and because goldmark's own line-segment
  convention (used elsewhere in this module) already treats a "line" as
  newline-inclusive.
- `Content` is `[]byte`, matching `span.Slice`'s convention rather than a
  string conversion.

Folded the blueprint delta into `specs/architecture/module-document-core.md`:
the Markdown Structure section's "Not yet built" paragraph is now split into
a **Built — fenced blocks only** paragraph naming `cli/core/markdown` and its
`Fenced`/`Block` surface, and a trailing paragraph still naming the heading
tree and comment-block location as not yet built. Updated the component
diagram's label, the module overview paragraph, and the Cross-Cutting
Relationships note that used to say every Markdown Structure edge was
"intended, not existing" — now only the Corpus → Markdown Structure edge is.
Also corrected a stale "neither...nor" claim in
`specs/architecture/module-spec-profile.md`'s Element Records section, which
said Document Core had no fenced-block extraction at all; it now says
extraction exists but nothing yet lets a profile register a tag into it
(the extension point is still unbuilt, out of scope here).

Gates: `gofmt -l .` clean; `go build ./...` clean; `go test ./...` all green
including the 13 new `core/markdown` tests; `check-skills.sh` (non-strict)
passes; `check-skills.sh --strict` exits 1 with 0 errors / 1 warning (the
pre-existing bead-bonanza/SKILL.md body-size warning — the documented
baseline, untouched by this change); `sync-shared-references.sh --check` and
`sync-strata-docs.sh --check` both pass; `generate-schemas.sh --check` stays
green (no schema touched).

Caveat: none.

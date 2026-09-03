---
title: Spec CLI Reference
description: The spec binary's command surface — what each subcommand
  decides, what it does not yet check, exit statuses, and how to install
  it — for skills that delegate corpus queries and conformance checks
slug: spec-cli-reference
type: guide
status: active
updated_at: 2026-08-09T00:00:00Z
version: "0.2"
---

# Spec CLI Reference

`spec` reads a specification corpus and answers questions about it that
would otherwise mean opening every document. Prefer it over reading the
corpus yourself for anything it decides: it is faster, and its answer is
verifiable by whoever reviews the work, where yours is an assertion.

## Contents

- Prerequisites — what must exist before any command works
- Commands — the five subcommands and the question each answers
- Reference checks — the four rules `check` applies to `[[slug]]` links
  and reference-bearing frontmatter, and the severity each carries
- Exit statuses — how to read a run's result
- What it does not check — the gap you still cover by hand
- When it is not installed

## Prerequisites

Two things, both cheap to confirm:

- **The binary.** `command -v spec`. If it is missing, see the last
  section — do not silently fall back to reading documents.
- **`specs.toml` at the repository root**, declaring `[workspace].root`
  (the corpus directory), `[workspace].name`, and `[profile].name`. Every
  command resolves the corpus through it. Its absence is not a default to
  infer: the CLI reports the path it looked for and exits non-zero.

`[workspace].root` is also the answer to "where do documents go" — read
it rather than assuming `docs/`.

## Commands

| Command | Answers |
|---------|---------|
| `spec check [path]` | Does the corpus — or one document — conform? |
| `spec index [--out <path>\|-]` | The whole corpus's frontmatter and reference graph, as JSON |
| `spec next-id <type>` | What number does the next document of this type take? |
| `spec slug-exists <slug>` | Is this slug already taken, and by which file? |
| `spec refs <slug> [--outbound]` | Who references this document — or what does it reference? |

**`spec check`** validates every document's frontmatter against the
active profile's schema and slug patterns, reports duplicate slugs, and
applies the four reference rules in the next section. Given a path it
narrows to that one document — the form to use after authoring — except
for unresolved references, which it reports wherever they were written,
so renaming a slug names the documents left pointing at nothing. `--json`
emits the full finding set with spans; `--strict` makes warning and
unverified findings fail too. `--only <rule>` narrows to named rules.

**`spec index`** writes the corpus index. Read it when you need many
documents' metadata at once — every slug in use, every document of a
given type, everything at a given status — instead of opening files. It
also carries the reference graph: each entry's outbound references with
their spans, and the inbound edges derived from every other member's
outbound ones. `--out -` sends it to stdout.

Indexing never fails on a broken reference. An unresolved reference is
recorded in the index and reported by `check`; that split is deliberate,
so a repository whose links are broken still contributes its graph.

`check`, `refs`, `slug-exists` and `next-id` refresh that same file when
they had to rebuild, so the second of two runs is served from it. You
never need to run `spec index` to make the others fast — run it when you
want the artifact itself. It is derived and disposable: deleting it
changes no answer, and it belongs in `.gitignore`.

**`spec next-id <type>`** returns one greater than the highest number in
use for a numbered type, across the whole corpus. This is the only
correct way to pick an ADR number; scanning filenames misses a document
whose filename and slug disagree.

**`spec slug-exists <slug>`** reports whether any document declares that
slug, and where. Case-insensitive, because a case-insensitive filesystem
cannot hold two files that differ only in case. Run it before claiming a
slug.

**`spec refs <slug>`** answers the blast-radius question. By default it
lists every corpus member whose body or reference-bearing frontmatter
names that slug, each at the line and column of the reference — run it
*before* retiring or renaming a document, not after. `--outbound`
reverses the direction and lists what that document itself references,
which is what to read before editing one. `--json` emits each reference
with its path, span, origin (body or which frontmatter property) and
what it resolved to, so a rewrite opens only the lines that carry the
reference.

Both directions exit zero when the answer is empty. "No corpus member
references this slug" is an answer, not a failure, and it is reported
distinctly from "no corpus member declares this slug" — the first means
the document is unreferenced, the second means you asked about a slug
that does not exist. Read the output, not the status.

## Reference checks

`spec check` applies four rules to references. All four resolve a
reference against the target's frontmatter `slug`, never a file path,
and without regard to letter case.

| Rule | Severity | Reports |
|------|----------|---------|
| `reference/unresolved` | error | A `[[slug]]` or reference-bearing frontmatter value matching no corpus member |
| `reference/retired-target` | warning | A live document referencing one at `status: retired`, quoting its `superseded_by` when it has one |
| `reference/pairing-incomplete` | error | One half of a paired relationship whose target does not declare the counterpart |
| `reference/type-mismatch` | error | A typed reference whose target declares a different `type` — a feature blueprint's `frd` naming something that is not an FRD |

Three things about `reference/unresolved` decide whether you trust it:

- **A backticked `[[slug]]` is not a reference.** Text inside a fenced
  code block or an inline code span is an illustration of the notation,
  and contributes nothing. Neither does a `[[slug]]` written in a
  frontmatter *value* — frontmatter contributes references only through
  the properties the profile registers as reference-bearing.
- **An anchor is split off, not resolved.** `[[container-x#Some
  Heading]]` resolves on `container-x` alone; whether that document has
  such a heading is not asked, and no finding is raised about the anchor.
- **A target that exists but is not a corpus member reports as such** —
  "target exists but is not a corpus member: parse failure at line 28"
  rather than as a slug matching nothing, because the reader's next step
  differs.

`reference/retired-target` is the only rule in the tool that emits
`warning`. A reference to a retired document still resolves — the reader
reaches a real document that says it has been replaced — so it does not
fail a build unless you pass `--strict`. It is also the one finding that
tells an author what to cite instead.

**`spec check <path>` reports the reference findings your edit caused
elsewhere.** All four reference rules escape the narrowing to the file
you name, because the document that breaks a reference is never the
document that holds it: rename a slug, retire a document or change its
`type`, check that one file, and it names every document left citing it,
at that document's own line. Every other rule stays scoped to the file
you named.

What no form of the check can tell you is whether a broken reference is
*yours*. It has no memory of what those references used to resolve to,
so breakage already in the corpus is reported alongside breakage you
caused. In a corpus you keep green that distinction is empty. `spec refs
<slug>` before the change still answers the blast-radius question ahead
of time, and is the cheaper habit.

**A path that names no corpus member is an error, not a pass.**
`spec check docs/typo.md` exits 1 and says the path is not a member —
and where the file exists but was excluded, it names the reason
(`no-frontmatter`, `unknown-type`). A clean report is never what you get
for a path the corpus does not hold.

Which frontmatter properties are reference-bearing is **derived from the
profile's schema**, not a hand-kept list: every property typed as a
document slug is one, except `slug` itself, which names its own document.
`superseded_by`, `sources`, `frd`, `system`, `contained_by`,
`contracts_provided` and `contracts_consumed` are all reference-bearing
for that reason, and a slug-typed property added to the vocabulary
becomes one the moment it is declared. This framework's own profile
registers **no** pairing, so `reference/pairing-incomplete` is inert
against its own corpus by design — it exists for adopting profiles whose
relationships are declared from both ends.

## Exit statuses

- **0** — the run succeeded. For `check`, no error-severity findings
  (and, under `--strict`, none at all).
- **1** — `check` found something, or the invocation itself failed: no
  `specs.toml`, an unknown profile, a malformed configuration.

A run that reports only `reference/retired-target` exits **0** without
`--strict` and **1** with it. That is the only place the two differ
today, so `--strict` is worth passing in CI where citing a retired
document should block.

`slug-exists` and `refs` exit 0 whether or not they find anything — "not
found" is an answer, not a failure. Read their output, not their status.

## What it does not check

`spec check` covers frontmatter properties, slug patterns, slug
uniqueness and reference integrity. It does **not** yet check required
sections, element records, or line width. Those are specified and not
built.

Within references, three things are deliberately out of scope and a
clean run says nothing about them: the **anchor** half of a reference
(only the slug half resolves), backticked **relative paths** to files
outside the spec set, and resolution **across workspaces** — a slug that
exists only in another repository's index reports as unresolved.
Reference findings also never suggest or perform a repair; they state
the fact and leave the fix to you.

This matters because a clean `check` looks like a clean document. It is
not: it means the mechanized rules passed. The rules that are still
yours are named in `shared/document-readiness-checklist.md`, which marks
each check with who decides it.

**A document the tool could not read** — frontmatter that parses as YAML
but does not fit the document envelope, most often a `status` outside the
vocabulary or an `updated_at` that is not an RFC 3339 date-time — is
reported as `frontmatter/parse`, at severity error. Fix that first: it
means nothing else could examine the document, so the other checks have
not cleared it. The commonest cause by far is a template placeholder
(`updated_at: <YYYY-MM-DDTHH:MM:SSZ>`) left unfilled.

## When it is not installed

Say so, and name the install, rather than reading the corpus by hand and
reporting a result you cannot show your work for:

```
go install github.com/brettcallaghan/agent-plugins/cli/spec@latest
```

or unpack a release archive from the repository's releases page onto
`PATH` (a single static binary; nothing runs at install time).

Reading the corpus by hand is the correct fallback for a project that
has chosen not to install it — but it is a fallback, and the human
should know that is what happened. The checks in
`shared/document-readiness-checklist.md` are defined independently of
the tool, so all of them can be run by hand; the cost is that the result
is an assertion rather than a verifiable one.

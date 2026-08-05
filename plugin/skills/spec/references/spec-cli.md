---
title: Spec CLI Reference
description: The spec binary's command surface — what each subcommand
  decides, what it does not yet check, exit statuses, and how to install
  it — for skills that delegate corpus queries and conformance checks
slug: spec-cli-reference
type: guide
status: active
updated_at: 2026-08-05T00:00:00Z
version: "0.1"
---

# Spec CLI Reference

`spec` reads a specification corpus and answers questions about it that
would otherwise mean opening every document. Prefer it over reading the
corpus yourself for anything it decides: it is faster, and its answer is
verifiable by whoever reviews the work, where yours is an assertion.

## Contents

- Prerequisites — what must exist before any command works
- Commands — the four subcommands and the question each answers
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
| `spec index [--out <path>\|-]` | The whole corpus's frontmatter, as JSON |
| `spec next-id <type>` | What number does the next document of this type take? |
| `spec slug-exists <slug>` | Is this slug already taken, and by which file? |

**`spec check`** validates every document's frontmatter against the
active profile's schema and slug patterns, and reports duplicate slugs.
Given a path it narrows to that one document — the form to use after
authoring. `--json` emits the full finding set with spans; `--strict`
makes warning and unverified findings fail too. `--only <rule>` narrows
to named rules.

**`spec index`** writes the corpus index. Read it when you need many
documents' metadata at once — every slug in use, every document of a
given type, everything at a given status — instead of opening files.
`--out -` sends it to stdout.

**`spec next-id <type>`** returns one greater than the highest number in
use for a numbered type, across the whole corpus. This is the only
correct way to pick an ADR number; scanning filenames misses a document
whose filename and slug disagree.

**`spec slug-exists <slug>`** reports whether any document declares that
slug, and where. Case-insensitive, because a case-insensitive filesystem
cannot hold two files that differ only in case. Run it before claiming a
slug.

## Exit statuses

- **0** — the run succeeded. For `check`, no error-severity findings
  (and, under `--strict`, none at all).
- **1** — `check` found something, or the invocation itself failed: no
  `specs.toml`, an unknown profile, a malformed configuration.

`slug-exists` exits 0 whether or not the slug is found — "not found" is
an answer, not a failure. Read its output, not its status.

## What it does not check

`spec check` covers frontmatter properties, slug patterns and slug
uniqueness. It does **not** yet check reference integrity, required
sections, element records, or line width. Those are specified and not
built.

This matters because a clean `check` looks like a clean document. It is
not: it means the mechanized rules passed. The rules that are still
yours are named in `document-readiness-checklist.md`, which marks each
check with who decides it.

**One case where `check` is silent rather than clean.** A document whose
frontmatter parses as YAML but does not fit the document envelope — an
`status` outside the vocabulary, most often — is recorded in the index
with a `parse_error` and is skipped by every check, which then reports
`passed`. Until that is fixed, confirm a newly authored document appears
in `spec index --out -` with a `frontmatter` object and no `parse_error`
before trusting a clean `check` on it. This is the one place where
reading the tool's own output matters more than its exit status.

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
`document-readiness-checklist.md` are defined independently of the tool,
so all of them can be run by hand; the cost is that the result is an
assertion rather than a verifiable one.

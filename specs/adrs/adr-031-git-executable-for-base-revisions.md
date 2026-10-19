---
title: "ADR-031: Base Revisions Are Read Through the git Executable"
description: The Spec CLI reads a corpus at an earlier git revision by
  running the git executable and copying that revision's tracked files
  into a temporary directory, rather than linking a git library or
  asking git to export an archive
slug: adr-031
type: adr
status: approved
---

# ADR-031: Base revisions are read through the git executable

**Status:** approved  
**Date:** 2026-10-08

## Context

FEAT-GHK-008 in [[frd-gherkin-adoption]] compares the criteria that the corpus
holds now against the criteria that it held at a git revision. That revision
is the merge-base of HEAD and the repository's default branch, or the
merge-base of HEAD and a revision that `--since` names. The persisted index
can't stand in for that revision, because any query might rebuild it. The CLI
must therefore read documents as they were at a commit, and nothing in the CLI
does that today.

Two constraints bind the choice:

- [[container-spec-cli]] ships one statically linked binary built with
  `CGO_ENABLED=0`, and a CI step runs it with no language runtime beyond the
  binary itself.
- The Corpus component builds an index by walking a directory on disk
  ([[module-document-core#Corpus]]). A base read that doesn't produce a
  directory needs a second build path.

The revision that's read is a commit in a repository that the caller already
has. A comparison against git is only requested where git is in use, and a
working tree without the git executable has no ordinary way to have been
checked out.

## Decision

We will read a base revision by running the `git` executable. The CLI does the
following:

1. Resolves the revision and the merge-base with `git rev-parse` and
   `git merge-base`.
1. Lists the spec root's tracked files at that commit with `git ls-tree`.
1. Writes each blob that `git cat-file --batch` returns into a temporary
   directory, at the blob's repository-relative path.
1. Builds the base corpus from that directory, using the same Corpus build
   that the working tree uses.

The CLI removes the temporary directory before the command exits.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| `git` executable, tracked blobs copied into a temporary directory | Reads exactly the bytes that the commit holds<br>Reuses the existing directory-walking build unchanged<br>Supports every repository layout that git supports | A runtime dependency on `git` for this one command<br>Writes a copy of the spec root to the temporary directory | N/A—selected |
| A pure-Go git library (`go-git`) | No runtime dependency beyond the binary<br>Static linking is unaffected | A large dependency tree for one flag<br>Lags git on repository formats that a caller might use (worktrees, partial clones, newer object formats)<br>The build still needs a directory or a second walk over a tree object | Adds more surface than the one comparison it serves, and fails on layouts that git handles |
| `git archive` piped into a temporary directory | One git invocation | Honors `export-ignore`: a document marked for exclusion from releases silently disappears from the base | Measured: with `specs/b.md export-ignore` in `.gitattributes`, `git archive HEAD -- specs \| tar -t` lists only `specs/a.md`, while `git ls-tree -r --name-only HEAD -- specs` lists both (git 2.43.0) |
| `git worktree add` at the base | A real checkout, so line endings and filters match the working tree | Writes into `.git`, and an interrupted run leaves a registered worktree behind | The CLI never modifies the repository that it reads |

## Consequences

- `criterion --changed` needs `git` on `PATH`. If `git` is absent, the CLI
  reports an environment error. No other command gains the dependency.
- The base corpus is built from blob bytes, without git's checkout filters.
  The Corpus read path already normalizes CRLF to LF, so a criterion's content
  hash is the same whichever line endings a blob was committed with.
- The CLI doesn't copy a file that the commit records as a symbolic link or a
  submodule, so a criterion that's reached only through one of these reads as
  added.
- Corpus discovery stays one walk over one directory. Nothing in Document Core
  learns about git objects, and the same tests exercise the base build and the
  working-tree build.
- The cost of a comparison grows with the size of the spec root at the base.
  The CLI reads nothing outside the spec root.

## References

- [[frd-gherkin-adoption]]: FEAT-GHK-008, the comparison that this decision
  serves.
- [[container-spec-cli]]: the single-binary and no-runtime constraints.
- [[module-document-core]]: the Corpus build that the base reuses, and the
  Revisions component that realizes this decision.
- [[cli-exit-status-standards]]: the class that a missing `git` is reported
  under.

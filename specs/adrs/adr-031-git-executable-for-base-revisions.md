---
title: "ADR-031: Base Revisions Are Read Through the git Executable"
description: The Spec CLI reads a corpus at an earlier git revision by
  running the git executable and copying that revision's tracked files
  into a temporary directory, rather than linking a git library or
  asking git to export an archive
slug: adr-031
type: adr
status: draft
updated_at: 2026-10-08T00:00:00Z
---

# ADR-031: Base Revisions Are Read Through the git Executable

**Status:** draft  
**Date:** 2026-10-08

## Context

[[frd-gherkin-adoption]]'s FEAT-GHK-008 compares the criteria the corpus
holds now against those it held at a git revision: the merge-base of
HEAD and the repository's default branch, or of HEAD and a revision
`--since` names. The persisted index cannot stand in for that revision,
because any query may rebuild it. So the CLI must read documents as
they were at a commit, which nothing in it does today.

Two constraints bind the choice. [[container-spec-cli]] ships one
statically linked binary built with `CGO_ENABLED=0`, and a CI step runs
it with no language runtime beyond itself. And the Corpus component
builds an index by walking a directory on disk
([[module-document-core#Corpus]]), so a base read that does not produce
a directory needs a second build path.

The revision being read is a commit in a repository the caller already
has. A comparison against git is only asked where git is in use, and a
working tree without the git executable has no ordinary way to have
been checked out.

## Decision

We will read a base revision by running the `git` executable. The CLI
resolves the revision and the merge-base with `git rev-parse` and
`git merge-base`, lists the spec root's tracked files at that commit
with `git ls-tree`, writes each blob `git cat-file --batch` returns into
a temporary directory at its repository-relative path, and builds the
base corpus from that directory with the same Corpus build the working
tree uses. The temporary directory is removed before the command exits.

## Alternatives Considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| `git` executable, tracked blobs copied into a temporary directory | Reads exactly the bytes the commit holds<br>Reuses the existing directory-walking build unchanged<br>Supports every repository layout git itself supports | A runtime dependency on `git` for this one command<br>Writes a copy of the spec root to the temporary directory | N/A — selected |
| A pure-Go git library (`go-git`) | No runtime dependency beyond the binary<br>Static linking unaffected | A large dependency tree for one flag<br>Lags git on repository formats a caller may use (worktrees, partial clones, newer object formats)<br>The build still needs a directory or a second walk over a tree object | Adds more surface than the one comparison it serves, and fails on layouts git handles |
| `git archive` piped into a temporary directory | One git invocation | Honours `export-ignore`: a document marked for exclusion from releases silently vanishes from the base | Measured: with `specs/b.md export-ignore` in `.gitattributes`, `git archive HEAD -- specs \| tar -t` lists only `specs/a.md`, while `git ls-tree -r --name-only HEAD -- specs` lists both (git 2.43.0) |
| `git worktree add` at the base | A real checkout, so line endings and filters match the working tree | Writes into `.git`, and an interrupted run leaves a registered worktree behind | The CLI never modifies the repository it reads |

## Consequences

- `criterion --changed` needs `git` on `PATH`. Its absence is an
  environment error, reported as one; no other command gains the
  dependency.
- The base corpus is built from blob bytes, without git's checkout
  filters. The Corpus read path already normalizes CRLF to LF, so a
  criterion's content hash is the same whichever form a blob was
  committed in.
- A file the commit records as a symbolic link or a submodule is not
  copied, so a criterion reached only through one reads as added.
- Corpus discovery stays one walk over one directory. Nothing in
  Document Core learns about git objects, and the base build is
  exercised by the same tests as the working-tree build.
- The cost of a comparison grows with the size of the spec root at the
  base. Nothing outside the spec root is read.

## References

- [[frd-gherkin-adoption]] — FEAT-GHK-008, the comparison this decision
  serves
- [[feature-gherkin-change-detection]] — the blueprint that builds it
- [[container-spec-cli]] — the single-binary and no-runtime constraints
- [[module-document-core]] — the Corpus build the base reuses
- [[cli-exit-status-standards]] — the class a missing `git` is reported
  under

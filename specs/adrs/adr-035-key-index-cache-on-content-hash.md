---
title: "ADR-035: Key the Index Cache on a Content Hash"
description: The Spec CLI serves a persisted index only when a hash of its
  inputs' contents matches the hash recorded on it, and never decides by
  modification time or absolute path
slug: adr-035
type: adr
status: ready-for-review
---

# ADR-035: Key the index cache on a content hash

**Status:** ready-for-review  
**Date:** 2026-11-23

## Context

The Spec CLI persists its index of the corpus in `specs.json` and answers
a query from it instead of rebuilding it. [[frd-spec-index]] sets a
95th-percentile target of 50 milliseconds for a query against a warm index
of 1,000 documents (AC-SPEC-009.1), requires a warm index to answer without
a rebuild (AC-SPEC-010.5), and requires any changed input to rebuild it
(AC-SPEC-010.6). So the CLI needs a cache key that tells whether a persisted
index was built from the inputs as they stand now.

A file's modification time is not a reliable stand-in for its content. A
fresh clone, a CI checkout with a committed or restored `specs.json`, and a
git checkout or rebase give files new modification times over the same
bytes. An edit can also keep a file's modification time and size, under a
coarse timestamp or a tool that preserves times. And the same corpus sits
at a different absolute path on each machine. A stale index is a silent
wrong answer, so the key must never call a changed corpus unchanged.

## Decision

We will key the persisted index on a SHA-256 hash over every source file's
repository-relative path and bytes, folded in path order. The hash also
covers the hash format's version, the index schema version, the workspace,
and the parts of the profile that decide what the index records. A query
serves the persisted index only when this hash equals the one recorded on
it, and otherwise rebuilds the whole index.

Modification time and absolute path never enter the key. Parts of the
profile that only checks read, against the live profile on every run, stay
out of the key, so changing one of them doesn't rebuild the index.

## Alternatives considered

| Option | Pros | Cons | Why rejected |
|--------|------|------|--------------|
| Modification time, with or without file size | A `stat` per file and no reads<br>Common in build tools | A fresh clone or checkout rebuilds over identical bytes<br>An edit that keeps time and size serves a stale index | Can serve a stale index, which AC-SPEC-010.6 rules out |
| Modification time and size first, content hash only when they differ | Skips reading unchanged files on a developer machine, where most queries run | An edit that keeps time and size is still served stale<br>Two paths that must agree on every input | Keeps the stale-index case of the first option; the saving is in reading, not parsing |
| Git object ids of the spec root's files | Git has already hashed committed files | Needs a repository and the `git` executable on every query, though the CLI also runs outside one (AC-SPEC-008.9)<br>Uncommitted and untracked edits need hashing anyway | Fails outside a repository and still hashes uncommitted edits, so it adds a dependency without removing the reads |
| Per-file hashes, rebuilding only changed entries | A one-file edit rebuilds less | Cross-document findings depend on other files, so a partial rebuild must track every dependency<br>A new way to serve a subtly wrong index | A full rebuild already has a one-second target with nothing cached (AC-SPEC-009.3); the saving doesn't pay for the risk |
| Content hash of every input, rebuilding the whole index | Same key on every machine and checkout<br>Any byte change rebuilds | Every warm query reads every source file | N/A — selected |

## Consequences

**Positive:**

- A fresh clone or checkout of an unchanged corpus with a persisted
  `specs.json` is served warm.
- The key is the same on every machine, wherever the repository sits.
- Any change to an input's bytes, including an unreadable file becoming
  readable, rebuilds the index.
- Deleting `specs.json` is always safe: the next query rebuilds it.

**Negative:**

- Every warm query walks the spec root and reads every source file to
  compute the hash. It skips parsing, not reading, so warm latency grows
  with corpus size even when nothing changed.
- A one-file edit rebuilds the whole index.
- An input the build reads that the hash doesn't cover serves a stale
  index. The function that computes the key must read exactly what the
  build reads; [[component-corpus]] holds that as an invariant.

## References

- [[component-corpus]]: the component that computes the hash and decides
  whether to serve a persisted index
- [[frd-spec-index]]: AC-SPEC-009.1, AC-SPEC-009.3, AC-SPEC-010.5 and
  AC-SPEC-010.6

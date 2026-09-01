// Package e2e is the swarm CLI's end-to-end tier: the compiled binary,
// run as a subprocess against a fixture restored from committed
// artifacts — per ADR-027 and the FIX delta of
// specs/features/swarm-core/swarm-core-blueprint.md.
//
// It carries no code of its own. Everything is in _test.go files: the
// harness that restores a fixture and replays a case, and the cases that
// need more than a data file can say. This file exists so the package
// has a name and a doc comment; a directory of only test files has
// neither.
//
// # What a fixture is, and how a case uses one
//
// A fixture is a directory under cli/testdata/swarm-fixtures holding
// three things: a `git bundle` of the repository, a tracker export, and
// a `fixture.toml` stamp saying which kind of fixture it is and which
// version of the JSON contract it was recorded against.
//
// A case materializes its own copy into a temporary directory, runs
// there, and deletes it. Nothing is restored in place and nothing is
// rolled back, so a case cannot lose a world another case needed and the
// suite parallelizes without further work.
//
// # Why committed artifacts rather than a generator
//
// A generator is code that can be wrong, and generated commits mint
// fresh shas that every snapshot would then have to normalize away. The
// bundle's shas are fixed because every commit in it was made with a
// pinned author, committer and date; the tracker export's ids and
// timestamps are fixed because it is hand-written. Measured: restoring
// the export with `bd import` returns the same ids and the same
// timestamps the export carries, so — contrary to what ADR-027
// anticipated — the tracker half needs no time normalization either.
package e2e

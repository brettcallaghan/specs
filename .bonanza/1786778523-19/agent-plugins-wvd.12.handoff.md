# agent-plugins-wvd.12 handoff

Added `spec criterion --changed`: a new `cli/cmd/criterion.go` subcommand
that compares the persisted index against a freshly built corpus and lists
every acceptance criterion whose content hash differs, classified as
changed/added/removed by whether its identifier is present in both, only
now, or only before (AC-GHK-008.1, AC-GHK-008.3). A criterion whose hash is
unchanged is never listed regardless of position (AC-GHK-008.2), since the
comparison reads only identifier and content hash from each side, never a
position or markdown diff.

The persisted side is read with `corpus.Load(indexPath)` rather than
`corpus.Open`: `Open`'s warm-cache path rebuilds silently on a hash
mismatch and returns the corpus it just built, discarding the earlier
state this comparison needs — exactly the case `--changed` exists to
report. Any unusable persisted index (absent, unreadable, unparsable, or
of an older schema version) is folded into "no persisted index" and the
command says so (`no persisted index at <path>; run \`spec index\` first`)
rather than reporting every criterion as added. The current side always
builds fresh via `corpus.Build`; neither call persists — a query must
never overwrite the baseline it is comparing against.

Output shape (my implicit decision, not pinned by the blueprint): one
tab-separated `<classification>\t<id>` line per change, sorted by
identifier for determinism, or the fixed line `no local changes` when the
comparison finds none. No `--json` here — that flag spans QRY's whole
query surface, not this one flag's output.

Command registered in `cli/cmd/root.go`. Folded into
`specs/architecture/container-spec-cli.md`'s Command Surface section
(`corpus.Load` added to the call list; a paragraph on `criterion
--changed`'s behavior and the `Open`-vs-`Load` choice added after the
`feature --emit` paragraph).

New snapshot cases: `cli/testdata/cases/criterion-changed` (moved
criterion not listed, one changed, one removed, one added, all in a
single fixture built by running the real CLI against a before/after
corpus so content hashes are authentic), `criterion-changed-no-index` (no
persisted index), `criterion-changed-none` (persisted and current agree).
Regenerating snapshots also updated `help-root`'s expected output to list
the new `criterion` line — the only incidental diff, reviewed and correct.

Caveat: created `spec criterion` command carrying only `--changed`, per
EMIT's own precedent for `spec feature --emit` — cobra.NoArgs, and
`spec criterion` without `--changed` returns an explicit "not yet
supported" error. QRY (agent-plugins-wvd.10) extends this same command
with identifier lookup, `--code`, `--of`, `--tag` and `--json`; nothing
here should collide with that, since no positional-arg handling or query
flags exist yet on this command.

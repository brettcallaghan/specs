# Run records

One document per `bead-bonanza` run, written by `sprint-retro` at the end of
it: the results, every finding the run captured verbatim, the routing, the
validation evidence, and what the run could not verify.

**There is no second report.** A run's record and its findings are the same
evidence read for two purposes, and this repository ran three sprints with them
split across two documents before deciding the split cost more than it bought.

The path is configured in `sf.toml`'s `[reports] dir`, and each run's record is
recorded on its run epic as `metadata.findings_path`, so `bd show <run-epic>`
finds it from the epic id alone.

Records written before 2026-08-25 live under `docs/research/sprint-findings/`
and, for the retired report half, in git history on the branches that produced
them.

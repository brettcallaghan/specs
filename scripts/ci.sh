#!/usr/bin/env bash
#
# ci.sh — run the project's declared checks by hand.
#
# There is ONE list, and it is `sf.toml`'s `[[checks]]`. This script runs it
# through the same runner a bead's close uses, so "what I ran locally" and
# "what the bead ran" cannot drift.
#
# What this deliberately does NOT run is the release bar — the cross-compile
# matrix, the three-platform test matrix, and the static-binary smoke test in a
# busybox image. Those only mean something across the whole repository at
# release time and are the workflows' job, not a developer's inner loop.
#
# Exit 0 all passed, 1 something failed, 3 the declaration itself is broken.

set -uo pipefail
cd "$(git rev-parse --show-toplevel)"
exec plugin/skills/bead-bonanza/scripts/bb-checks.py "$@"

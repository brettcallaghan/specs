#!/usr/bin/env bash
#
# Regenerate the expected artifacts of every CLI snapshot case under
# cli/testdata/cases/, per specs/standards/testing-standards.md's "a
# script that regenerates every expected artifact... SHALL be provided."
#
# This only writes the expected artifacts each case declares (checks.log,
# checks.json indented, checks.err, manifest.json), from the current
# cli/cmd implementation. It does not judge
# whether the resulting diff is correct — that review is the author's,
# per the same standard's "WHEN a regenerated snapshot differs from the
# committed one, the author SHALL confirm the difference was intended
# before committing it."
#
# Usage:
#   scripts/regenerate-cli-snapshots.sh
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT/cli"

go run ./testdata/gen ./testdata/cases

echo
echo "review the diff before committing:"
echo "  git -C \"$REPO_ROOT\" diff -- cli/testdata/cases"

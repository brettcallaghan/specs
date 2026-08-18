#!/usr/bin/env bash
#
# Measure statement coverage of the Spec CLI across every layer that
# exercises it, and render the result into the root README.
#
# Three legs, run separately so each layer's contribution is visible
# rather than only their union:
#
#   unit      every package's own tests, minus the snapshot replays
#   snapshot  the data-driven cases under cli/testdata/cases/, replayed
#             in-process against the command tree
#   e2e       cli/spec's tests, which build the binary and run it as a
#             process
#
# The end-to-end leg is the one that needs machinery. Its subject is a
# separate process, so `go test -coverprofile` sees nothing it does: the
# binary is built with `go build -cover` and each run writes its counters
# to GOCOVERDIR, which `go tool covdata textfmt` then converts into a
# profile the other two legs' output can be merged with. Without this the
# layer that proves the exit statuses would report as zero coverage.
#
# Usage:
#   scripts/coverage.sh              measure, write README, write reports
#   scripts/coverage.sh --check      measure, fail if the README is stale
#   scripts/coverage.sh --min 70     also fail below a total percentage
#
# Reports are written to build/coverage/ (gitignored):
#
#   summary.md         the badge and the tables, as published in the README
#   coverage-func.txt  per-function coverage and the total
#   coverage.html      annotated source, one file at a time
#   coverage.out       the merged profile — input to other tools, not a report
#
# Regenerate with the Go version in cli/go.mod. Statement counts come
# from the toolchain, so a different one can move the numbers and make
# --check fail on an unrelated change.
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$REPO_ROOT/build/coverage"
MODULE="github.com/brettcallaghan/agent-plugins/cli"

# Instrument every package in the module, not merely the one under test:
# a snapshot case run from cli/cmd spends nearly all of its statements in
# cli/core, and the default (each package covering itself) would not
# count any of them.
COVERPKG="$MODULE/..."

# The tests that replay cli/testdata/cases/. They are split out of the
# unit leg by name because they live beside the unit tests of the same
# package. A snapshot replay added without being listed here is reported
# under "unit" — a misattribution between two columns, which leaves the
# combined total correct.
SNAPSHOT_TESTS='TestSnapshotCases|TestCRLFCheckoutProducesIdenticalOutput'

REPORT_ARGS=()
while [ $# -gt 0 ]; do
  case "$1" in
    --check)
      REPORT_ARGS+=(--check)
      ;;
    --min)
      [ $# -ge 2 ] || { echo "--min requires a percentage" >&2; exit 2; }
      REPORT_ARGS+=(--min "$2")
      shift
      ;;
    -h|--help)
      echo "usage: scripts/coverage.sh [--check] [--min PCT]"
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 2
      ;;
  esac
  shift
done

# A stale directory would let a deleted package keep reporting coverage,
# and would let a leg that failed to run be scored from its last result.
rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR/e2e-data"

cd "$REPO_ROOT/cli"

# -count=1 on every leg: a cached test result is replayed without running
# the test, so the end-to-end leg would write no counters at all and the
# in-process legs would report whatever the cache last held.

# The unit leg covers every package except cli/spec, whose only tests are
# the end-to-end ones measured by the third leg.
unit_packages=()
while IFS= read -r pkg; do
  [ "$pkg" = "$MODULE/spec" ] && continue
  unit_packages+=("$pkg")
done < <(go list ./...)

echo "==> unit tests"
go test -count=1 -coverpkg="$COVERPKG" -coverprofile="$OUT_DIR/unit.cov" \
  -skip "$SNAPSHOT_TESTS" "${unit_packages[@]}"

echo
echo "==> snapshot cases"
go test -count=1 -coverpkg="$COVERPKG" -coverprofile="$OUT_DIR/snapshot.cov" \
  -run "$SNAPSHOT_TESTS" ./cmd

echo
echo "==> end-to-end binary tests"
SPEC_E2E_COVERDIR="$OUT_DIR/e2e-data" go test -count=1 ./spec

# The counters are per-process and binary-format; textfmt turns the whole
# directory into one profile in the same shape as the legs above.
go tool covdata textfmt -i="$OUT_DIR/e2e-data" -o="$OUT_DIR/e2e.cov"

echo
# The report's exit status is held rather than propagated immediately, so
# that a failing --check or --min still leaves the annotated source
# behind — that report is what such a failure is diagnosed from.
report_status=0
python3 "$REPO_ROOT/scripts/coverage-report.py" \
  --module "$MODULE" \
  --source-root "$REPO_ROOT/cli" \
  --exclude-generated \
  --layer "Unit tests=$OUT_DIR/unit.cov" \
  --layer "Snapshot cases=$OUT_DIR/snapshot.cov" \
  --layer "End-to-end (built binary)=$OUT_DIR/e2e.cov" \
  --merged-out "$OUT_DIR/coverage.out" \
  --markdown-out "$OUT_DIR/summary.md" \
  --readme "$REPO_ROOT/README.md" \
  "${REPORT_ARGS[@]+"${REPORT_ARGS[@]}"}" || report_status=$?

if [ -f "$OUT_DIR/coverage.out" ]; then
  # Three shapes, because they answer different questions and only the
  # first is readable without a browser: the tables (summary.md, written
  # by the report above), the per-function list, and the annotated
  # source. coverage.out itself is a machine profile — it is kept for
  # tooling, not for reading.
  go tool cover -func="$OUT_DIR/coverage.out" -o "$OUT_DIR/coverage-func.txt"
  go tool cover -html="$OUT_DIR/coverage.out" -o "$OUT_DIR/coverage.html"
  echo "tables:           build/coverage/summary.md"
  echo "per function:     build/coverage/coverage-func.txt"
  echo "annotated source: build/coverage/coverage.html"
fi

exit "$report_status"

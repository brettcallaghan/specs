#!/usr/bin/env bash
#
# ci.sh — the full check picture for this repository, run by hand.
#
# THIS IS NOT `swarm.toml`'s `[[checks]]`, and it deliberately runs more.
#
#   swarm.toml   the quality bar for ONE COMMIT — what every bead runs before its
#             own close. Sized to be affordable every time.
#   ci.sh     everything a contributor should see green before pushing.
#
# They overlap today and will not stay that way: anything too slow to pay per
# bead, or that only means something across the whole repository, belongs here
# and not there. Wiring this script to the declaration would erase that
# distinction the first time the two need to differ.
#
# Not here either: the release bar — the six-target cross-compile, the
# three-platform test matrix, and the static-binary smoke test in a busybox
# image. Those need CI's runners, and `.github/workflows/` owns them.
#
# Exit 0 everything passed, 1 something failed.

set -uo pipefail
cd "$(git rev-parse --show-toplevel)"

rc=0
run() { # run <label> <command...>
  printf '\n=== %s\n' "$1"; shift
  # No `&&` chaining and no `set -e`: one failure must not hide the checks
  # behind it. Every status is recorded and the script fails at the end.
  if "$@"; then :; else rc=1; printf '!!! FAILED: %s\n' "$1"; fi
}
sh_run() { printf '\n=== %s\n' "$1"; shift; if sh -c "$*"; then :; else rc=1; printf '!!! FAILED\n'; fi; }

# --- Go module under cli/ -------------------------------------------------
sh_run "format"        'scripts/format-check'
sh_run "vet"           'cd cli && go vet ./...'
sh_run "build"         'cd cli && go build ./...'
sh_run "test-go"       'cd cli && go test ./...'
# Re-runs the suite three times instrumented, so it is here rather than in
# sf.toml. Fails when README.md's committed figures are not what it measured.
sh_run "coverage"      'scripts/coverage.sh --check'

# --- documentation corpus and skills --------------------------------------
sh_run "skills"        'scripts/check-skills.sh --strict'
sh_run "shared-refs"   'scripts/sync-shared-references.sh --check'
sh_run "ref-links"     'scripts/check-reference-links.py'
sh_run "template-drift" 'scripts/check-template-drift.py'

# --- the Spec CLI against this repository's own documents ------------------
sh_run "corpus"        'cd cli && go build -o ../build/spec ./spec && cd .. && build/spec check --strict'
sh_run "rfc-queue"     'build/spec check --config rfcs/specs.toml --only frontmatter/parse,frontmatter/schema,frontmatter/missing-property,slug/pattern-mismatch'

# --- generated artifacts (need a local strata build; each SKIPs without one)
sh_run "strata-docs"   'scripts/sync-strata-docs.sh --check'
sh_run "schemas"       'scripts/generate-schemas.sh --check'

printf '\n=== RESULT: %s\n' "$([ $rc -eq 0 ] && echo PASS || echo FAIL)"
exit $rc

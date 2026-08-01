#!/usr/bin/env bash
#
# Regenerate every artifact derived from schemas/*.strata.
#
# The .strata files are the single source of truth. Everything this
# script writes is generated and must not be hand-edited — each output
# carries a generated-by banner saying so.
#
# Usage:
#   scripts/generate-schemas.sh          # regenerate in place
#   scripts/generate-schemas.sh --check  # fail if any output is stale
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

SRC="schemas"
SCHEMA_OUT="cli/core/schema"
GO_OUT="cli/core/specmodel"
MD_OUT="plugin/skills/spec/references/generated"

if ! command -v strata >/dev/null 2>&1; then
  echo "error: strata not found on PATH." >&2
  echo "  The session-start hook installs it; see .claude/skills/strata/SKILL.md" >&2
  echo "  to build it by hand from a strata-lang/strata checkout." >&2
  exit 1
fi

CHECK=0
[[ "${1:-}" == "--check" ]] && CHECK=1

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# 1. Validate. --target ast runs the parser and every semantic check but
#    emits only JSON, so it is the cheapest correctness gate. Nothing is
#    written to the tree until this passes.
strata build "$SRC" --target ast --output "$WORK/ast" --error-format json

# 2. Bundled JSON Schema covering every type. The profile registers each
#    document type against a JSON Pointer into this bundle.
strata build "$SRC" --target json-schema --output "$WORK/schema"
mv "$WORK/schema/SpecModel.schema.json" "$WORK/schema/spec-model.schema.json"

# 3. The standalone specs.toml schema, filtered to SpecsConfig and its
#    transitive dependencies. This is the artifact contract-specs-
#    configuration.md names as normative for the shape of specs.toml.
strata build "$SRC" --target json-schema \
  --config "$SRC/overlays/specs-config.json" \
  --output "$WORK/cfg"
mv "$WORK/cfg/SpecModel.schema.json" "$WORK/schema/specs-config.schema.json"

# 4. Go structs for the CLI and its modules.
strata build "$SRC" --target go --output "$WORK/go"

# 5. Markdown reference shipped with the plugin, describing the shape of
#    frontmatter to anyone authoring a document by hand.
strata build "$SRC" --target markdown --output "$WORK/md"

cat >"$WORK/md/README.md" <<'BANNER'
# Generated schema reference

**Generated — do not edit.** Every file in this directory is produced by
`scripts/generate-schemas.sh` from the `.strata` sources in `schemas/`.
Edit those and re-run the script; a change made here is lost on the next
regeneration.

These pages describe the shape of the frontmatter and configuration a
document author writes. The normative artifacts derived from the same
sources are the JSON Schemas under `cli/core/schema/`.
BANNER

sync_dir() {
  local src="$1" dest="$2"
  if [[ $CHECK -eq 1 ]]; then
    if ! diff -r -q "$src" "$dest" >/dev/null 2>&1; then
      echo "error: $dest is stale — run scripts/generate-schemas.sh" >&2
      diff -r "$src" "$dest" >&2 || true
      return 1
    fi
  else
    rm -rf "$dest"
    mkdir -p "$(dirname "$dest")"
    cp -r "$src" "$dest"
  fi
}

rc=0
sync_dir "$WORK/schema" "$SCHEMA_OUT" || rc=1
sync_dir "$WORK/go" "$GO_OUT" || rc=1
sync_dir "$WORK/md" "$MD_OUT" || rc=1

if [[ $rc -ne 0 ]]; then
  exit 1
fi

if [[ $CHECK -eq 1 ]]; then
  echo "generated artifacts are up to date"
else
  echo "regenerated:"
  echo "  $SCHEMA_OUT  (JSON Schema)"
  echo "  $GO_OUT  (Go)"
  echo "  $MD_OUT  (Markdown)"
fi

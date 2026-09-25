#!/usr/bin/env bash
#
# Install the `strata` CLI from a strata-lang/strata GitHub release.
#
# Downloads the prebuilt binary for this platform, verifies it against the
# sha256 digest GitHub publishes for the asset, proves it runs, and installs
# it. A no-op when the requested release is already installed.
#
# Usage:
#   scripts/install-strata-cli.sh            # latest release
#   scripts/install-strata-cli.sh v0.8.0     # a specific tag
#
# Environment:
#   STRATA_ACCESS_TOKEN  GitHub token with read access to strata-lang/strata.
#                        Required while the repository is private; sent only
#                        to api.github.com.
#   STRATA_VERSION       Release tag to install when no argument is given.
#                        Default: latest.
#   STRATA_INSTALL_DIR   Where the binary goes. Default: /usr/local/bin.
#
# Exit status: 0 installed or already current, 1 on any failure.
set -euo pipefail

REPO="strata-lang/strata"
API="https://api.github.com/repos/$REPO"
WANT="${1:-${STRATA_VERSION:-latest}}"
INSTALL_DIR="${STRATA_INSTALL_DIR:-/usr/local/bin}"
BIN="$INSTALL_DIR/strata"
CACHE="${HOME:-/root}/.cache/strata-cli"
STAMP="$CACHE/installed-release"

log() { echo "[install-strata-cli] $*" >&2; }
die() { log "ERROR: $*"; exit 1; }

command -v curl >/dev/null 2>&1 || die "curl is required"
command -v python3 >/dev/null 2>&1 || die "python3 is required"
command -v tar >/dev/null 2>&1 || die "tar is required"

case "$(uname -s)/$(uname -m)" in
  Linux/x86_64)          TARGET=x86_64-unknown-linux-gnu ;;
  Linux/aarch64)         TARGET=aarch64-unknown-linux-gnu ;;
  Darwin/arm64)          TARGET=aarch64-apple-darwin ;;
  Darwin/x86_64)         TARGET=x86_64-apple-darwin ;;
  *) die "no strata release binary for $(uname -s)/$(uname -m)" ;;
esac
ASSET="strata-$TARGET.tar.gz"

# GitHub API request. Tried through the environment's proxy settings first,
# then directly: an egress proxy that brokers its own GitHub credential
# replaces the Authorization header, and answers 403 for a repository the
# session was not granted — which strata, private, usually is.
gh_curl() {
  local out="$1" url="$2" accept="$3"
  local args=(-fsSL --retry 2 -o "$out" -H "Accept: $accept"
              -H "X-GitHub-Api-Version: 2022-11-28")
  [ -n "${STRATA_ACCESS_TOKEN:-}" ] \
    && args+=(-H "Authorization: Bearer $STRATA_ACCESS_TOKEN")
  curl "${args[@]}" "$url" 2>/dev/null \
    || curl --noproxy '*' "${args[@]}" "$url"
}

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

if [ "$WANT" = latest ]; then
  release_url="$API/releases/latest"
else
  release_url="$API/releases/tags/$WANT"
fi
if ! gh_curl "$WORK/release.json" "$release_url" application/vnd.github+json; then
  [ -n "${STRATA_ACCESS_TOKEN:-}" ] \
    || die "could not read release '$WANT' from $REPO — STRATA_ACCESS_TOKEN is unset, and the repository is private"
  die "could not read release '$WANT' from $REPO — check STRATA_ACCESS_TOKEN has read access to it"
fi

# tag <TAB> asset API url <TAB> sha256 — one line, or nothing if the asset is
# absent from the release.
meta="$(python3 - "$WORK/release.json" "$ASSET" <<'PY'
import json, sys
release = json.load(open(sys.argv[1]))
for asset in release.get("assets", []):
    if asset["name"] == sys.argv[2]:
        digest = asset.get("digest") or ""
        sha = digest.split(":", 1)[1] if digest.startswith("sha256:") else ""
        print(f'{release["tag_name"]}\t{asset["url"]}\t{sha}')
PY
)"
[ -n "$meta" ] || die "release '$WANT' has no asset $ASSET"
IFS=$'\t' read -r TAG ASSET_URL SHA <<<"$meta"

if [ -x "$BIN" ] && [ "$(cat "$STAMP" 2>/dev/null)" = "$TAG" ] \
   && "$BIN" --version >/dev/null 2>&1; then
  log "strata $TAG already installed at $BIN"
  exit 0
fi

log "downloading $ASSET ($TAG)"
gh_curl "$WORK/$ASSET" "$ASSET_URL" application/octet-stream \
  || die "download of $ASSET failed"

if [ -n "$SHA" ]; then
  if command -v sha256sum >/dev/null 2>&1; then
    got="$(sha256sum "$WORK/$ASSET")"
  else
    got="$(shasum -a 256 "$WORK/$ASSET")"
  fi
  [ "${got%% *}" = "$SHA" ] || die "$ASSET sha256 is ${got%% *}, release says $SHA"
else
  log "warning: release publishes no digest for $ASSET — not verified"
fi

tar -xzf "$WORK/$ASSET" -C "$WORK" strata || die "$ASSET has no strata binary"
version="$("$WORK/strata" --version 2>&1)" || die "the downloaded binary does not run: $version"

mkdir -p "$INSTALL_DIR" "$CACHE"
install -m 0755 "$WORK/strata" "$BIN" || die "could not install to $BIN"
printf '%s' "$TAG" >"$STAMP"
log "installed $version ($TAG) at $BIN"

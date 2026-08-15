#!/usr/bin/env bash
# Installs the zharness CLI from a GitHub release into ~/.local/bin.
#
# Usage: install-zharness.sh [tag]
#   tag defaults to the latest zharness release. Releases are triggered by
#   pushing a "cli/vX.Y.Z" tag, but goreleaser requires its current-tag to
#   parse as semver, so the published release itself is always tagged with
#   the bare version (e.g. "v0.1.0"), not the "cli/v..." trigger tag.
#
# Requires: gh (authenticated against this repo), tar.
set -euo pipefail

REPO="therealtinhtute/mono-harness"
INSTALL_DIR="${HOME}/.local/bin"

if ! command -v gh >/dev/null 2>&1; then
  echo "error: gh (GitHub CLI) is required and not on PATH" >&2
  exit 1
fi

case "$(uname -s)" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *)
    echo "error: unsupported OS $(uname -s)" >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *)
    echo "error: unsupported architecture $(uname -m)" >&2
    exit 1
    ;;
esac

tag="${1:-}"
if [ -z "$tag" ]; then
  tag=$(gh release list --repo "$REPO" --limit 50 --json tagName,name,isDraft \
    --jq '[.[] | select(.isDraft==false) | select(.name | startswith("zharness "))][0].tagName')
  if [ -z "$tag" ] || [ "$tag" = "null" ]; then
    echo "error: no zharness release found on $REPO" >&2
    exit 1
  fi
fi

asset="zharness_${os}_${arch}.tar.gz"
work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' EXIT

echo "installing zharness ${tag} (${os}/${arch}) from ${REPO}..."
gh release download "$tag" --repo "$REPO" --pattern "$asset" --dir "$work_dir"

mkdir -p "$INSTALL_DIR"
tar -xzf "${work_dir}/${asset}" -C "$work_dir" zharness
install -m 0755 "${work_dir}/zharness" "${INSTALL_DIR}/zharness"

echo "installed to ${INSTALL_DIR}/zharness"
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *) echo "warning: ${INSTALL_DIR} is not on PATH" >&2 ;;
esac

"${INSTALL_DIR}/zharness" --version

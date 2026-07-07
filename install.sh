#!/usr/bin/env sh
# Installs the latest (or a pinned) release of tg into $INSTALL_DIR.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/jlimas/tg/main/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/jlimas/tg/main/install.sh | VERSION=v0.1.0 sh
#
# Env vars:
#   VERSION      release tag to install (default: latest)
#   INSTALL_DIR  install directory (default: $HOME/.local/bin)
set -eu

REPO="jlimas/tg"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-latest}"

err() {
  echo "error: $*" >&2
  exit 1
}

need() {
  command -v "$1" >/dev/null 2>&1 || err "$1 is required but not installed"
}

need curl
need tar

os=""
case "$(uname -s)" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *) err "unsupported OS: $(uname -s) (tg ships linux/darwin binaries only)" ;;
esac

arch=""
case "$(uname -m)" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) err "unsupported architecture: $(uname -m) (tg ships amd64/arm64 binaries only)" ;;
esac

if [ "$VERSION" = "latest" ]; then
  tag=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
  [ -n "$tag" ] || err "could not determine the latest release tag"
else
  tag="$VERSION"
fi

archive="tg_${os}_${arch}.tar.gz"
url="https://github.com/${REPO}/releases/download/${tag}/${archive}"

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

echo "downloading tg ${tag} for ${os}/${arch}..." >&2
curl -fsSL "$url" -o "$tmp_dir/$archive" || err "failed to download $url (does this release/platform combination exist?)"

tar -xzf "$tmp_dir/$archive" -C "$tmp_dir" tg

mkdir -p "$INSTALL_DIR"
install -m 755 "$tmp_dir/tg" "$INSTALL_DIR/tg"

echo "installed tg ${tag} to ${INSTALL_DIR}/tg" >&2

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "note: ${INSTALL_DIR} is not on your PATH — add it to your shell profile" >&2 ;;
esac

echo "run 'tg config set --bot-token \"<token>\"' to get started" >&2

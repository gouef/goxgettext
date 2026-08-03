#!/usr/bin/env sh
set -eu

REPO="gouef/goxgettext"
VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

os_name=$(uname -s)
arch_name=$(uname -m)

case "$os_name" in
  Linux)
    case "$arch_name" in
      x86_64|amd64) asset="goxgettext-linux-amd64" ;;
      aarch64|arm64) asset="goxgettext-linux-arm64" ;;
      *) echo "Unsupported architecture: $arch_name" >&2; exit 1 ;;
    esac
    ;;
  Darwin)
    case "$arch_name" in
      x86_64|amd64) asset="goxgettext-darwin-amd64" ;;
      arm64|aarch64) asset="goxgettext-darwin-arm64" ;;
      *) echo "Unsupported architecture: $arch_name" >&2; exit 1 ;;
    esac
    ;;
  *)
    echo "Unsupported platform: $os_name" >&2
    exit 1
    ;;
esac

if [ "$VERSION" = "latest" ]; then
  url="https://github.com/${REPO}/releases/latest/download/${asset}"
else
  url="https://github.com/${REPO}/releases/download/${VERSION}/${asset}"
fi

mkdir -p "$INSTALL_DIR"

tmp_file=$(mktemp "${TMPDIR:-/tmp}/goxgettext.XXXXXX")
trap 'rm -f "$tmp_file"' EXIT HUP INT TERM

curl -fsSL "$url" -o "$tmp_file"
chmod 0755 "$tmp_file"

install_path="$INSTALL_DIR/goxgettext"
mv "$tmp_file" "$install_path"

echo "Installed goxgettext to $install_path"
echo "Run it with: $install_path --help"

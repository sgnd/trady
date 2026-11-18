#!/bin/sh
# Install the latest trady release (placeholder until first GitHub Release).
#
#   curl -fsSL https://raw.githubusercontent.com/sgnd/trady/main/install.sh | sh
#
# Installs to /usr/local/bin when writable, otherwise ~/.local/bin.
# Override with TRADY_INSTALL_DIR. Pin a version with TRADY_VERSION=v0.1.0.
set -eu

REPO="sgnd/trady"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux | darwin) ;;
  mingw* | msys* | cygwin*)
    echo "This installer is for Linux and macOS. On Windows, download" >&2
    echo "trady_windows_amd64.zip from https://github.com/$REPO/releases." >&2
    exit 1 ;;
  *) echo "trady supports Linux, macOS and Windows; there is no release for $os." >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) echo "There is no trady release for this CPU ($(uname -m)): amd64 and arm64 only." >&2; exit 1 ;;
esac

for tool in curl tar; do
  command -v "$tool" >/dev/null 2>&1 || { echo "The installer needs $tool." >&2; exit 1; }
done

if [ -n "${TRADY_INSTALL_DIR:-}" ]; then
  dir=$TRADY_INSTALL_DIR
  mkdir -p "$dir"
elif [ -w /usr/local/bin ] 2>/dev/null || mkdir -p /usr/local/bin 2>/dev/null && [ -w /usr/local/bin ]; then
  dir=/usr/local/bin
else
  dir=${HOME}/.local/bin
  mkdir -p "$dir"
fi

ver=${TRADY_VERSION:-latest}
if [ "$ver" = latest ]; then
  echo "trady: no GitHub Release published yet (Phase 0). Build from source:" >&2
  echo "  git clone https://github.com/$REPO.git && cd trady && make build && ./bin/trady version" >&2
  exit 1
fi

asset="trady_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/${ver}/${asset}"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
curl -fsSL "$url" -o "$tmp/$asset"
tar -xzf "$tmp/$asset" -C "$tmp"
install -m 755 "$tmp/trady" "$dir/trady"
echo "Installed trady to $dir/trady"
"$dir/trady" version

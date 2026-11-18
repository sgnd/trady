#!/bin/sh
# Install a trady release. The archive is installed only when its sha256
# matches the checksums.txt published with that release.
#
#   curl -fsSL https://raw.githubusercontent.com/sgnd/trady/main/install.sh | sh
#
# Installs to /usr/local/bin when writable, otherwise ~/.local/bin.
# Override with TRADY_INSTALL_DIR. Pin a version with TRADY_VERSION=v0.0.1.
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
command -v sha256sum >/dev/null 2>&1 || command -v shasum >/dev/null 2>&1 ||
  { echo "The installer needs sha256sum or shasum to check the download." >&2; exit 1; }

# sha256_of FILE: the sha256 of FILE, in hex
sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  else
    shasum -a 256 "$1" | cut -d' ' -f1
  fi
}

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
  ver=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1)
  if [ -z "$ver" ]; then
    echo "trady: no GitHub Release found. Build from source:" >&2
    echo "  git clone https://github.com/$REPO.git && cd trady && make build && ./bin/trady version" >&2
    exit 1
  fi
fi

asset="trady_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/${ver}/${asset}"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
curl -fsSL "$url" -o "$tmp/$asset"
curl -fsSL "https://github.com/$REPO/releases/download/${ver}/checksums.txt" -o "$tmp/checksums.txt"
want=$(awk -v f="$asset" '$2 == f || $2 == "*" f { print $1 }' "$tmp/checksums.txt")
got=$(sha256_of "$tmp/$asset")
if [ -z "$want" ] || [ "$want" != "$got" ]; then
  echo "trady: $asset does not match checksums.txt of $ver; nothing was installed." >&2
  exit 1
fi
tar -xzf "$tmp/$asset" -C "$tmp"
install -m 755 "$tmp/trady" "$dir/trady"
echo "Installed trady to $dir/trady"
"$dir/trady" version

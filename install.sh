#!/bin/sh
# Install delta from a GitHub release.
#
#   curl -fsSL https://raw.githubusercontent.com/ronz204/deltas-ai/main/install.sh | sh
#
# Environment:
#   DELTA_VERSION      version to install, without the leading v (default: the latest release)
#   DELTA_INSTALL_DIR  where to put the binary (default: $HOME/.local/bin)
#   DELTA_BASE_URL     where the archives and checksums.txt are (default: the GitHub release).
#                      Meant for testing against a local folder.

set -eu

REPO="ronz204/deltas-ai"

say() { printf '%s\n' "$*"; }
fail() { printf 'install: %s\n' "$*" >&2; exit 1; }

have() { command -v "$1" >/dev/null 2>&1; }

# fetch URL DEST: download a file.
fetch() {
	if have curl; then
		curl -fsSL "$1" -o "$2"
	elif have wget; then
		wget -q "$1" -O "$2"
	else
		fail "curl or wget is required"
	fi
}

# sha256 FILE: print the SHA-256 of a file.
sha256() {
	if have sha256sum; then
		sha256sum "$1" | cut -d ' ' -f 1
	elif have shasum; then
		shasum -a 256 "$1" | cut -d ' ' -f 1
	else
		fail "sha256sum or shasum is required to verify the download"
	fi
}

have tar || fail "tar is required"

case "$(uname -s)" in
	Linux) os=linux ;;
	Darwin) os=darwin ;;
	*) fail "unsupported system $(uname -s); on Windows use install.ps1" ;;
esac

case "$(uname -m)" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) fail "unsupported architecture $(uname -m)" ;;
esac

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT INT TERM

version="${DELTA_VERSION:-}"
if [ -z "$version" ]; then
	none="no stable release of $REPO was found (pre-releases do not count). Set DELTA_VERSION to install one, for example DELTA_VERSION=0.1.0-rc1"
	fetch "https://api.github.com/repos/$REPO/releases/latest" "$tmp/latest.json" 2>/dev/null || fail "$none"
	version="$(sed -n 's/.*"tag_name": *"v\{0,1\}\([^"]*\)".*/\1/p' "$tmp/latest.json" | head -n 1)"
	[ -n "$version" ] || fail "$none"
fi
version="${version#v}"

archive="delta_${version}_${os}_${arch}.tar.gz"
base="${DELTA_BASE_URL:-https://github.com/$REPO/releases/download/v$version}"

say "Installing delta $version for $os/$arch"
fetch "$base/$archive" "$tmp/$archive" || fail "could not download $base/$archive"
fetch "$base/checksums.txt" "$tmp/checksums.txt" || fail "could not download $base/checksums.txt"

expected="$(awk -v f="$archive" '$2 == f { print $1 }' "$tmp/checksums.txt")"
[ -n "$expected" ] || fail "$archive is not listed in checksums.txt"
actual="$(sha256 "$tmp/$archive")"
[ "$expected" = "$actual" ] || fail "checksum mismatch for $archive (expected $expected, got $actual)"

tar -xzf "$tmp/$archive" -C "$tmp" delta || fail "could not extract delta from $archive"

dir="${DELTA_INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p "$dir"
install -m 755 "$tmp/delta" "$dir/delta"

"$dir/delta" version || fail "the installed binary does not run"
say "Installed to $dir/delta"

case ":$PATH:" in
	*":$dir:"*) ;;
	*)
		say ""
		say "$dir is not on your PATH. Add this line to your shell profile, then restart the terminal:"
		say "  export PATH=\"$dir:\$PATH\""
		;;
esac

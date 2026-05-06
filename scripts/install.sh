#!/usr/bin/env sh
# install.sh - install the zd-cli binary on macOS or Linux.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/hackath0r/zd-cli/main/scripts/install.sh | sh
#
# Optional environment variables:
#   ZD_VERSION   pin a specific release tag (default: latest)
#   ZD_PREFIX    install prefix (default: /usr/local or ~/.local if not writable)
#   ZD_NO_SUDO   set to any value to skip sudo escalation
#
# What this script does:
#   1. Detects OS + arch
#   2. Resolves the latest GitHub release (or the pinned ZD_VERSION)
#   3. Downloads the matching tarball + checksum
#   4. Verifies the sha256
#   5. Extracts zd into ${ZD_PREFIX}/bin
#   6. Creates a 'ximr' symlink next to it
#   7. On macOS, removes the com.apple.quarantine xattr so Gatekeeper does not
#      block the unsigned binary on first run.
#
# zd-cli is currently distributed unsigned on macOS; passing the binary
# through xattr -d com.apple.quarantine is the documented workaround. If
# you would prefer notarized binaries, please open an issue.

set -eu

REPO="hackath0r/zd-cli"
BIN="zd"
ALIAS="ximr"

err()  { printf 'install.sh: %s\n' "$*" >&2; exit 1; }
info() { printf 'install.sh: %s\n' "$*"; }

detect_os() {
    case "$(uname -s)" in
        Linux*)   echo linux ;;
        Darwin*)  echo darwin ;;
        *)        err "unsupported OS: $(uname -s); see Windows install.ps1" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo amd64 ;;
        arm64|aarch64)  echo arm64 ;;
        *)              err "unsupported arch: $(uname -m)" ;;
    esac
}

resolve_version() {
    if [ "${ZD_VERSION:-}" != "" ]; then
        echo "$ZD_VERSION"
        return
    fi
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
            | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" \
            | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1
    else
        err "neither curl nor wget is installed"
    fi
}

resolve_prefix() {
    if [ "${ZD_PREFIX:-}" != "" ]; then
        echo "$ZD_PREFIX"
        return
    fi
    if [ -w /usr/local/bin ] 2>/dev/null; then
        echo /usr/local
        return
    fi
    if [ -z "${ZD_NO_SUDO:-}" ] && command -v sudo >/dev/null 2>&1; then
        echo /usr/local
        return
    fi
    mkdir -p "$HOME/.local/bin"
    echo "$HOME/.local"
}

download() {
    url="$1"; out="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fL --progress-bar -o "$out" "$url"
    else
        wget -O "$out" "$url"
    fi
}

main() {
    os=$(detect_os)
    arch=$(detect_arch)
    version=$(resolve_version)
    [ -n "$version" ] || err "could not resolve version (try ZD_VERSION=v0.1.0)"
    prefix=$(resolve_prefix)
    info "installing zd-cli ${version} for ${os}/${arch} into ${prefix}/bin"

    archive="${BIN}_${version#v}_${os}_${arch}.tar.gz"
    url="https://github.com/${REPO}/releases/download/${version}/${archive}"
    sums_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"

    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT

    info "downloading ${archive}"
    download "$url" "$tmp/$archive"
    download "$sums_url" "$tmp/checksums.txt"

    if command -v sha256sum >/dev/null 2>&1; then
        (cd "$tmp" && grep " ${archive}\$" checksums.txt | sha256sum -c -) || err "checksum mismatch"
    elif command -v shasum >/dev/null 2>&1; then
        (cd "$tmp" && grep " ${archive}\$" checksums.txt | shasum -a 256 -c -) || err "checksum mismatch"
    else
        info "warning: no sha256sum/shasum available; skipping checksum verification"
    fi

    info "extracting"
    tar -xzf "$tmp/$archive" -C "$tmp"

    install_bin="${prefix}/bin/${BIN}"
    install_alias="${prefix}/bin/${ALIAS}"

    if [ -w "${prefix}/bin" ] 2>/dev/null || [ "$(id -u)" = "0" ]; then
        mv "$tmp/$BIN" "$install_bin"
        ln -sf "$install_bin" "$install_alias"
    else
        info "writing to ${prefix} requires sudo"
        sudo mv "$tmp/$BIN" "$install_bin"
        sudo ln -sf "$install_bin" "$install_alias"
    fi

    chmod +x "$install_bin" 2>/dev/null || sudo chmod +x "$install_bin"

    if [ "$os" = "darwin" ]; then
        info "clearing macOS quarantine xattr (zd-cli is unsigned)"
        xattr -d com.apple.quarantine "$install_bin" 2>/dev/null || true
        xattr -d com.apple.quarantine "$install_alias" 2>/dev/null || true
    fi

    info "installed:"
    "$install_bin" version || true
    info "next: run '${BIN} config init' to set up a profile"
}

main "$@"

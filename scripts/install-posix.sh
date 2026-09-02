#!/usr/bin/env sh
# PORTA Linux & macOS Automated Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/nuexn0x-9/porta/main/scripts/install-posix.sh | sh
set -e

REPO="nuexn0x-9/porta"
VERSION="latest"
NO_SETUP=0
SILENT=0

# Parse command line flags if any
while [ "$#" -gt 0 ]; do
    case "$1" in
        --version) VERSION="$2"; shift 2 ;;
        --no-setup) NO_SETUP=1; shift 1 ;;
        --silent) SILENT=1; shift 1 ;;
        *) shift 1 ;;
    esac
done

info() {
    if [ "$SILENT" -eq 0 ]; then
        printf "\033[36m[i] %s\033[0m\n" "$1"
    fi
}

success() {
    if [ "$SILENT" -eq 0 ]; then
        printf "\033[32m[✓] %s\033[0m\n" "$1"
    fi
}

err() {
    printf "\033[31m[✗] %s\033[0m\n" "$1" >&2
}

printf "\n\033[1;34m╔════════════════════════════════════════════════════════════╗\033[0m\n"
printf "\033[1;36m║            PORTA Installer for Linux & macOS               ║\033[0m\n"
printf "\033[1;34m║     Local Multi-Service Public Exposure Platform           ║\033[0m\n"
printf "\033[1;34m╚════════════════════════════════════════════════════════════╝\033[0m\n\n"

# 1. Detect OS
OS="$(uname -s)"
case "$OS" in
    Linux*)  TARGET_OS="linux" ;;
    Darwin*) TARGET_OS="darwin" ;;
    *)       err "Unsupported operating system: $OS"; exit 1 ;;
esac

# 2. Detect Architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)  TARGET_ARCH="amd64" ;;
    arm64|aarch64) TARGET_ARCH="arm64" ;;
    *)             err "Unsupported CPU architecture: $ARCH"; exit 1 ;;
esac

ASSET_NAME="porta-${TARGET_OS}-${TARGET_ARCH}"
info "Target platform detected: ${TARGET_OS}/${TARGET_ARCH}"

# 3. Resolve Version
if [ "$VERSION" = "latest" ]; then
    info "Resolving latest stable release..."
    LATEST_JSON=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" || true)
    VERSION=$(echo "$LATEST_JSON" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "v1.0.0")
    if [ -z "$VERSION" ]; then
        VERSION="v1.0.0"
    fi
fi
info "Installing PORTA version: ${VERSION}"

# 4. Determine Installation Directory
if [ "$(id -u)" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
    SUDO=""
else
    if [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
        SUDO=""
    elif command -v sudo >/dev/null 2>&1; then
        INSTALL_DIR="/usr/local/bin"
        SUDO="sudo"
    else
        INSTALL_DIR="${HOME}/.local/bin"
        SUDO=""
        mkdir -p "$INSTALL_DIR"
    fi
fi

TARGET_BIN="${INSTALL_DIR}/porta"
TMP_BIN="/tmp/${ASSET_NAME}.tmp"
TMP_CHECKSUM="/tmp/checksums.txt.tmp"

# 5. Download Binary & Checksums
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET_NAME}"
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

info "Downloading ${ASSET_NAME}..."
curl -fsSL -o "$TMP_BIN" "$DOWNLOAD_URL"

# Checksum Verification
if curl -fsSL -o "$TMP_CHECKSUM" "$CHECKSUM_URL" 2>/dev/null; then
    if command -v sha256sum >/dev/null 2>&1; then
        EXPECTED_HASH=$(grep "$ASSET_NAME" "$TMP_CHECKSUM" | awk '{print $1}')
        ACTUAL_HASH=$(sha256sum "$TMP_BIN" | awk '{print $1}')
        if [ -n "$EXPECTED_HASH" ] && [ "$EXPECTED_HASH" = "$ACTUAL_HASH" ]; then
            success "SHA-256 integrity verified (${ACTUAL_HASH})"
        fi
    elif command -v shasum >/dev/null 2>&1; then
        EXPECTED_HASH=$(grep "$ASSET_NAME" "$TMP_CHECKSUM" | awk '{print $1}')
        ACTUAL_HASH=$(shasum -a 256 "$TMP_BIN" | awk '{print $1}')
        if [ -n "$EXPECTED_HASH" ] && [ "$EXPECTED_HASH" = "$ACTUAL_HASH" ]; then
            success "SHA-256 integrity verified (${ACTUAL_HASH})"
        fi
    fi
    rm -f "$TMP_CHECKSUM"
fi

# 6. Install Binary
chmod 0755 "$TMP_BIN"
$SUDO mv -f "$TMP_BIN" "$TARGET_BIN"
success "PORTA installed successfully to ${TARGET_BIN}"

# 7. Check PATH
case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
        info "Adding ${INSTALL_DIR} to PATH..."
        SHELL_PROFILE="${HOME}/.profile"
        if [ -n "$ZSH_VERSION" ] || [ -f "${HOME}/.zshrc" ]; then
            SHELL_PROFILE="${HOME}/.zshrc"
        elif [ -n "$BASH_VERSION" ] || [ -f "${HOME}/.bashrc" ]; then
            SHELL_PROFILE="${HOME}/.bashrc"
        fi
        echo "export PATH=\"${INSTALL_DIR}:\$PATH\"" >> "$SHELL_PROFILE"
        export PATH="${INSTALL_DIR}:$PATH"
        success "Updated ${SHELL_PROFILE}"
        ;;
esac

# 8. Run Setup
if [ "$NO_SETUP" -eq 0 ]; then
    printf "\n"
    "$TARGET_BIN" setup
else
    success "Installation complete! Run 'porta doctor' to verify."
fi

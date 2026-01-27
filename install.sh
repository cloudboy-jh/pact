#!/bin/sh
set -e

REPO="cloudboy-jh/pact"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="pact"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

info() { printf "${GREEN}▸${NC} %s\n" "$1"; }
warn() { printf "${YELLOW}▸${NC} %s\n" "$1"; }
error() { printf "${RED}▸${NC} %s\n" "$1" >&2; exit 1; }

detect_os() {
    case "$(uname -s)" in
        Darwin*) echo "darwin" ;;
        Linux*)  echo "linux" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) error "Unsupported OS: $(uname -s)" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | \
        grep '"tag_name":' | \
        sed -E 's/.*"([^"]+)".*/\1/'
}

main() {
    info "Installing Pact CLI..."

    OS=$(detect_os)
    ARCH=$(detect_arch)
    VERSION=$(get_latest_version)

    [ -z "$VERSION" ] && error "Failed to get latest version"

    VERSION_NUM=$(echo "$VERSION" | sed 's/^v//')

    info "Detected: ${OS}/${ARCH}"
    info "Version: ${VERSION}"

    if [ "$OS" = "windows" ]; then
        FILENAME="pact_${VERSION_NUM}_${OS}_${ARCH}.zip"
    else
        FILENAME="pact_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
    fi

    URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    info "Downloading ${FILENAME}..."

    TMP_DIR=$(mktemp -d)
    trap "rm -rf ${TMP_DIR}" EXIT

    curl -fsSL "$URL" -o "${TMP_DIR}/${FILENAME}" || error "Download failed: ${URL}"

    cd "$TMP_DIR"

    if [ "$OS" = "windows" ]; then
        command -v unzip >/dev/null 2>&1 || error "unzip is required"
        unzip -q "${FILENAME}" || error "Failed to unzip ${FILENAME}"
    else
        tar -xzf "${FILENAME}" || error "Failed to extract ${FILENAME}"
    fi

    [ -f "${BINARY_NAME}" ] || error "Binary ${BINARY_NAME} not found in archive"

    if [ ! -d "$INSTALL_DIR" ]; then
        warn "${INSTALL_DIR} does not exist, creating..."
        mkdir -p "$INSTALL_DIR" || error "Failed to create ${INSTALL_DIR}"
    fi

    if [ ! -w "$INSTALL_DIR" ]; then
        info "Elevating permissions to install in ${INSTALL_DIR}"
        sudo install -m 755 "${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}" || \
            error "Failed to install ${BINARY_NAME} to ${INSTALL_DIR}"
    else
        install -m 755 "${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}" || \
            error "Failed to install ${BINARY_NAME} to ${INSTALL_DIR}"
    fi

    info "Installed ${BINARY_NAME} to ${INSTALL_DIR}"
    info "Run: ${BINARY_NAME} --help"
}

main "$@"

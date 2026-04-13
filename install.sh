#!/bin/sh
set -e

REPO="cesarbqz/platform-tool"
INSTALL_NAME="platform-tool"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Normalize arch
case "$ARCH" in
x86_64 | amd64) ARCH="amd64" ;;
arm64 | aarch64) ARCH="arm64" ;;
*)
  echo "Unsupported architecture: $ARCH" >&2
  exit 1
  ;;
esac

# Normalize OS
case "$OS" in
linux | darwin | windows) ;;
*)
  echo "Unsupported OS: $OS" >&2
  exit 1
  ;;
esac

# Append .exe if windows
EXT=""
if [ "$OS" = "windows" ]; then
  EXT=".exe"
fi

# Auth header (optional for public repo, helps avoid rate limits)
AUTH_HEADER=""
if [ -n "$GH_TOKEN" ]; then
  AUTH_HEADER="-H \"Authorization: token ${GH_TOKEN}\""
fi

# Get latest version tag from GitHub API
LATEST=$(curl -sf $AUTH_HEADER \
  "https://api.github.com/repos/$REPO/releases/latest" |
  grep '"tag_name":' | cut -d '"' -f4)

if [ -z "$LATEST" ]; then
  echo "Could not fetch the latest release. Try setting GH_TOKEN if you are being rate limited."
  exit 1
fi

echo "Installing $INSTALL_NAME $LATEST for $OS $ARCH"

# Build asset name
ASSET_NAME="${INSTALL_NAME}_${OS}_${ARCH}${EXT}"

# Get asset download URL from GitHub API
DOWNLOAD_URL=$(curl -sf $AUTH_HEADER \
  "https://api.github.com/repos/$REPO/releases/tags/$LATEST" |
  grep -A1 "\"name\": \"$ASSET_NAME\"" | grep '"browser_download_url"' | cut -d '"' -f4)

if [ -z "$DOWNLOAD_URL" ]; then
  echo "Asset '$ASSET_NAME' not found in release $LATEST" >&2
  exit 1
fi

echo "Downloading $ASSET_NAME from $REPO..."

mkdir -p "$INSTALL_DIR"

TMP_FILE=$(mktemp)
curl -sfL $AUTH_HEADER "$DOWNLOAD_URL" -o "$TMP_FILE"

FINAL_PATH="$INSTALL_DIR/${INSTALL_NAME}${EXT}"
mv "$TMP_FILE" "$FINAL_PATH"
chmod +x "$FINAL_PATH"

echo "Installed to $FINAL_PATH"

# PATH check
case ":$PATH:" in
*":$INSTALL_DIR:"*) ;;
*)
  echo "$INSTALL_DIR is not in your PATH."
  echo "Add this to your shell profile:"
  echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
  ;;
esac

echo "Running: ${INSTALL_NAME} version"
"$FINAL_PATH" version || echo "Could not run the CLI."

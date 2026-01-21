#!/bin/bash
set -e

# Install GitHub CLI (without using deb repository)
type -p curl >/dev/null || (echo "curl not found" && exit 1)

ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH=amd64
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH=arm64
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

GH_VERSION=$(curl -s https://api.github.com/repos/cli/cli/releases/latest | grep '"tag_name":' | sed -E 's/.*"v([^"]+)".*/\1/')
if [ -z "$GH_VERSION" ]; then
    echo "Could not determine latest GitHub CLI version"
    exit 1
fi

TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

curl -fsSL -o gh.tar.gz "https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_${ARCH}.tar.gz"
tar -xzf gh.tar.gz
sudo cp "gh_${GH_VERSION}_linux_${ARCH}/bin/gh" /usr/local/bin/gh
sudo chmod +x /usr/local/bin/gh

cd /
rm -rf "$TMP_DIR"

# Verify installation
if command -v gh &> /dev/null; then
    echo "GitHub CLI installed successfully"
    gh --version
else
    echo "GitHub CLI installation failed"
    exit 1
fi

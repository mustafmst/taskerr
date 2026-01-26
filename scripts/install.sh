#!/bin/bash
set -e

REPO_URL="https://github.com/mustafmst/taskerr.git"
INSTALL_DIR="$HOME/.local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}==>${NC} $1"
}

warn() {
    echo -e "${YELLOW}Warning:${NC} $1"
}

error() {
    echo -e "${RED}Error:${NC} $1"
    exit 1
}

# Check for required dependencies
check_dependency() {
    if ! command -v "$1" &> /dev/null; then
        error "$1 is required but not installed. Please install $1 and try again."
    fi
}

info "Checking dependencies..."
check_dependency git
check_dependency go
check_dependency make

# Check Go version
GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+' | head -1)
info "Found Go version: $GO_VERSION"

# Clone to temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

info "Cloning taskerr..."
git clone --depth 1 "$REPO_URL" "$TMP_DIR" --quiet

info "Building and installing..."
cd "$TMP_DIR"
make install

# Ensure ~/.local/bin exists
mkdir -p "$INSTALL_DIR"

# Add ~/.local/bin to PATH if not already present
add_to_path() {
    local shell_rc="$1"
    local shell_name="$2"
    
    if [ -f "$shell_rc" ]; then
        if ! grep -q 'export PATH="$HOME/.local/bin:$PATH"' "$shell_rc" 2>/dev/null; then
            echo '' >> "$shell_rc"
            echo '# Added by taskerr installer' >> "$shell_rc"
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$shell_rc"
            info "Added ~/.local/bin to PATH in $shell_rc"
            return 0
        fi
    fi
    return 1
}

PATH_UPDATED=false

# Detect shell and update appropriate config
if [ -n "$ZSH_VERSION" ] || [ "$SHELL" = "/bin/zsh" ] || [ "$SHELL" = "/usr/bin/zsh" ]; then
    if add_to_path "$HOME/.zshrc" "zsh"; then
        PATH_UPDATED=true
    fi
elif [ -n "$BASH_VERSION" ] || [ "$SHELL" = "/bin/bash" ] || [ "$SHELL" = "/usr/bin/bash" ]; then
    if add_to_path "$HOME/.bashrc" "bash"; then
        PATH_UPDATED=true
    fi
fi

# Also check for .profile as a fallback
if [ -f "$HOME/.profile" ] && ! grep -q 'export PATH="$HOME/.local/bin:$PATH"' "$HOME/.profile" 2>/dev/null; then
    if [ "$PATH_UPDATED" = false ]; then
        echo '' >> "$HOME/.profile"
        echo '# Added by taskerr installer' >> "$HOME/.profile"
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.profile"
        info "Added ~/.local/bin to PATH in ~/.profile"
        PATH_UPDATED=true
    fi
fi

echo ""
info "Installation complete!"
echo ""
echo "taskerr has been installed to ~/.local/bin/taskerr"
echo ""

if [ "$PATH_UPDATED" = true ]; then
    echo "Please restart your shell or run:"
    echo "  source ~/.bashrc  # or ~/.zshrc for zsh users"
    echo ""
fi

echo "Then run 'taskerr' to start the application."

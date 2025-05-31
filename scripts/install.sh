#!/usr/bin/env bash
#
# GoFi Installation Script
# This script downloads and installs the latest GoFi binary for your platform
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# GitHub repository
REPO="d-fi/GoFi"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="gofi"

# Detect OS and Architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$OS" in
        darwin)
            PLATFORM="darwin"
            ;;
        linux)
            PLATFORM="linux"
            ;;
        *)
            echo -e "${RED}Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}Detected platform: ${PLATFORM}-${ARCH}${NC}"
}

# Get the latest release download URL
get_download_url() {
    local api_url="https://api.github.com/repos/${REPO}/releases/latest"
    local download_file="gofi-${PLATFORM}-${ARCH}.tar.gz"
    
    echo -e "${YELLOW}Fetching latest release...${NC}"
    
    # Use curl to get the latest release data
    local release_data=$(curl -s "$api_url")
    
    # Extract download URL using grep and sed (more portable than jq)
    local download_url=$(echo "$release_data" | grep -o "\"browser_download_url\": \"[^\"]*${download_file}\"" | sed 's/.*: "\(.*\)"/\1/')
    
    if [ -z "$download_url" ]; then
        echo -e "${RED}Could not find download URL for ${download_file}${NC}"
        echo -e "${RED}Please check https://github.com/${REPO}/releases for available downloads${NC}"
        exit 1
    fi
    
    echo "$download_url"
}

# Download and install the binary
install_binary() {
    local download_url="$1"
    local temp_dir=$(mktemp -d)
    local archive_path="${temp_dir}/gofi.tar.gz"
    
    echo -e "${YELLOW}Downloading GoFi...${NC}"
    curl -L -o "$archive_path" "$download_url"
    
    echo -e "${YELLOW}Extracting archive...${NC}"
    tar -xzf "$archive_path" -C "$temp_dir"
    
    # Find the binary (it should be named gofi-platform-arch)
    local binary_path=$(find "$temp_dir" -name "gofi-*" -type f | head -n 1)
    
    if [ -z "$binary_path" ]; then
        echo -e "${RED}Could not find binary in archive${NC}"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        echo -e "${YELLOW}Installing to ${INSTALL_DIR}/${BINARY_NAME}...${NC}"
        mv "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        echo -e "${YELLOW}Installing to ${INSTALL_DIR}/${BINARY_NAME} (requires sudo)...${NC}"
        sudo mv "$binary_path" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    # Clean up
    rm -rf "$temp_dir"
    
    echo -e "${GREEN}✓ GoFi has been installed successfully!${NC}"
    echo -e "${GREEN}Run 'gofi --help' to get started${NC}"
}

# Verify installation
verify_installation() {
    if command -v gofi &> /dev/null; then
        local version=$(gofi --version 2>/dev/null || echo "unknown")
        echo -e "${GREEN}GoFi is installed at: $(which gofi)${NC}"
        echo -e "${GREEN}Version: $version${NC}"
    else
        echo -e "${RED}Installation verification failed${NC}"
        exit 1
    fi
}

# Main installation flow
main() {
    echo -e "${GREEN}GoFi Installation Script${NC}"
    echo "========================"
    
    # Check for curl
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}Error: curl is required but not installed${NC}"
        echo "Please install curl and try again"
        exit 1
    fi
    
    detect_platform
    
    # Allow custom install directory
    if [ -n "$1" ]; then
        INSTALL_DIR="$1"
        echo -e "${YELLOW}Using custom install directory: $INSTALL_DIR${NC}"
    fi
    
    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        echo -e "${YELLOW}Creating install directory: $INSTALL_DIR${NC}"
        if [ -w "$(dirname "$INSTALL_DIR")" ]; then
            mkdir -p "$INSTALL_DIR"
        else
            sudo mkdir -p "$INSTALL_DIR"
        fi
    fi
    
    local download_url=$(get_download_url)
    echo -e "${GREEN}Download URL: $download_url${NC}"
    
    install_binary "$download_url"
    verify_installation
}

# Run main function
main "$@"
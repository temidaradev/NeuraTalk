#!/bin/bash

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Application configuration
APP_NAME="neuratalk"
APP_DISPLAY_NAME="Neura Talk"
BIN_DIR="/usr/local/bin"
DESKTOP_DIR="/usr/share/applications"
ICON_DIR="/usr/share/icons/hicolor/512x512/apps"

# macOS specific paths
if [[ "$OSTYPE" == "darwin"* ]]; then
    BIN_DIR="/usr/local/bin"
    DESKTOP_DIR="$HOME/Applications"
    APP_BUNDLE="$DESKTOP_DIR/$APP_NAME.app"
fi

# System detection
detect_pkg_manager() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        if command -v brew &> /dev/null; then
            PKG_MANAGER="brew"
        else
            echo -e "${RED}Error: Homebrew not found.${NC}"
            exit 1
        fi
    elif command -v apt-get &> /dev/null; then
        PKG_MANAGER="apt-get"
    elif command -v dnf &> /dev/null; then
        PKG_MANAGER="dnf"
    elif command -v yum &> /dev/null; then
        PKG_MANAGER="yum"
    elif command -v pacman &> /dev/null; then
        PKG_MANAGER="pacman"
    else
        echo -e "${RED}Error: Unsupported package manager${NC}"
        exit 1
    fi
}

remove_application() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo -e "${YELLOW}Removing macOS application...${NC}"
        if [[ -d "$APP_BUNDLE" ]]; then
            rm -rf "$APP_BUNDLE"
            echo -e "${GREEN}Removed $APP_BUNDLE${NC}"
        fi
        if [[ -f "$BIN_DIR/$APP_NAME" ]]; then
            rm -f "$BIN_DIR/$APP_NAME"
            echo -e "${GREEN}Removed $BIN_DIR/$APP_NAME${NC}"
        fi
    else
        echo -e "${YELLOW}Removing Linux application...${NC}"
        if [[ -f "$BIN_DIR/$APP_NAME" ]]; then
            sudo rm -f "$BIN_DIR/$APP_NAME"
            echo -e "${GREEN}Removed $BIN_DIR/$APP_NAME${NC}"
        fi
        if [[ -f "$DESKTOP_DIR/$APP_NAME.desktop" ]]; then
            sudo rm -f "$DESKTOP_DIR/$APP_NAME.desktop"
            echo -e "${GREEN}Removed desktop entry${NC}"
        fi
        if [[ -f "$ICON_DIR/$APP_NAME.png" ]]; then
            sudo rm -f "$ICON_DIR/$APP_NAME.png"
            echo -e "${GREEN}Removed application icon${NC}"
        fi
        if command -v update-desktop-database &> /dev/null; then
            sudo update-desktop-database "$DESKTOP_DIR"
        fi
        if command -v gtk-update-icon-cache &> /dev/null; then
            sudo gtk-update-icon-cache -f /usr/share/icons/hicolor/
        fi
    fi
}

cleanup_tmp_files() {
    echo -e "${YELLOW}Cleaning up temporary files...${NC}"
    if [[ -d "./tmp" ]]; then
        rm -rf "./tmp"
        echo -e "${GREEN}Removed temporary directory${NC}"
    fi
    if [[ -d "./conversations" ]]; then
        rm -rf "./conversations"
        echo -e "${GREEN}Removed conversations directory${NC}"
    fi
}

main() {
    # Check root privileges for Linux
    if [[ "$OSTYPE" != "darwin"* ]] && [[ $EUID -ne 0 ]]; then
        echo -e "${RED}Error: This script requires root privileges on Linux.${NC}"
        echo "Please run with sudo or as root"
        exit 1
    fi

    detect_pkg_manager
    remove_application
    cleanup_tmp_files

    echo -e "\n${GREEN}Uninstallation completed successfully!${NC}"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "Application has been removed from your system."
    else
        echo "Application has been removed from your system."
        echo "You may need to log out and back in for all changes to take effect."
    fi
}

main

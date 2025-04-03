#!/bin/bash

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Application configuration
APP_NAME="neuratalk"
APP_DISPLAY_NAME="Neura Talk"  # Human-readable name
BIN_DIR="/usr/local/bin"
DESKTOP_DIR="/usr/share/applications"
ICON_DIR="/usr/share/icons/hicolor/512x512/apps"
ICON_SRC="icon.png"
CATEGORIES="Utility;Development;"  # Standard categories: https://standards.freedesktop.org/menu-spec/latest/apa.html

# System detection
detect_pkg_manager() {
    if command -v apt-get &> /dev/null; then
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

# Install system packages without system upgrades
install_packages() {
    local packages=("$@")
    case $PKG_MANAGER in
        "apt-get") sudo apt-get install -y --no-upgrade "${packages[@]}" ;;
        "dnf"|"yum") sudo $PKG_MANAGER install -y "${packages[@]}" ;;
        "pacman") sudo pacman -S --needed --noconfirm "${packages[@]}" ;;
    esac
}

install_dependencies() {
    echo -e "${YELLOW}Checking system dependencies...${NC}"

    # Common dependencies
    if ! command -v gcc &> /dev/null; then
        echo -e "${YELLOW}Installing build essentials...${NC}"
        case $PKG_MANAGER in
            "apt-get") install_packages build-essential ;;
            "dnf"|"yum") install_packages gcc gcc-c++ make ;;
            "pacman") install_packages base-devel ;;
        esac
    fi

    # Graphics dependencies
    declare -A pkg_map=(
        ["apt-get"]="libgl1-mesa-dev xorg-dev libgtk-3-dev"
        ["dnf"]="mesa-libGL-devel libX11-devel libXcursor-devel libXrandr-devel gtk3-devel"
        ["yum"]="mesa-libGL-devel libX11-devel libXcursor-devel libXrandr-devel gtk3-devel"
        ["pacman"]="mesa libx11 libxcursor libxrandr gtk3"
    )

    echo -e "${YELLOW}Installing graphics dependencies...${NC}"
    install_packages ${pkg_map[$PKG_MANAGER]}
}

create_desktop_entry() {
    echo -e "${YELLOW}Creating desktop entry...${NC}"

    # Create desktop file with proper localization support
    cat << EOF | sudo tee "$DESKTOP_DIR/$APP_NAME.desktop" > /dev/null
[Desktop Entry]
Version=1.0
Type=Application
Name=$APP_DISPLAY_NAME
GenericName=Application
Comment=A Fyne-based application
Exec=$BIN_DIR/$APP_NAME
Icon=$APP_NAME
Terminal=false
Categories=$CATEGORIES
StartupWMClass=$APP_NAME
Keywords=app;fyne;
EOF

    # Update desktop database
    if command -v update-desktop-database &> /dev/null; then
        echo -e "${YELLOW}Updating desktop database...${NC}"
        sudo update-desktop-database "$DESKTOP_DIR"
    fi
}

install_icon() {
    if [[ -f "$ICON_SRC" ]]; then
        echo -e "${YELLOW}Installing application icon...${NC}"
        sudo mkdir -p "$ICON_DIR"
        sudo cp "$ICON_SRC" "$ICON_DIR/$APP_NAME.png"

        # Update icon cache
        if command -v gtk-update-icon-cache &> /dev/null; then
            echo -e "${YELLOW}Updating icon cache...${NC}"
            sudo gtk-update-icon-cache -f /usr/share/icons/hicolor/
        fi
    else
        echo -e "${YELLOW}Warning: Icon file not found ($ICON_SRC)${NC}"
    fi
}

main() {
    # Check root privileges
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}Error: This script requires root privileges.${NC}"
        echo "Please run with sudo or as root"
        exit 1
    fi

    detect_pkg_manager
    install_dependencies

    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go not found!${NC}"
        exit 1
    fi

    echo -e "${YELLOW}Building application with CGO...${NC}"
    CGO_ENABLED=1 go build -o "$APP_NAME"

    echo -e "${YELLOW}Installing application...${NC}"
    sudo install -Dm755 "$APP_NAME" "$BIN_DIR/$APP_NAME"

    create_desktop_entry
    install_icon

    echo -e "\n${GREEN}Installation completed successfully!${NC}"
    echo "Application should now appear in your application menu."
    echo "You may need to log out and back in for changes to take full effect."
}

main

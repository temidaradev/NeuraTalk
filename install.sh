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

# macOS specific paths
if [[ "$OSTYPE" == "darwin"* ]]; then
    BIN_DIR="/usr/local/bin"
    DESKTOP_DIR="$HOME/Applications"
    ICON_DIR="$HOME/Applications/$APP_NAME.app/Contents/Resources"
fi

# System detection
detect_pkg_manager() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        if command -v brew &> /dev/null; then
            PKG_MANAGER="brew"
        else
            echo -e "${RED}Error: Homebrew not found. Please install Homebrew first.${NC}"
            echo "Visit https://brew.sh for installation instructions."
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

# Install system packages without system upgrades
install_packages() {
    local packages=("$@")
    case $PKG_MANAGER in
        "brew") brew install "${packages[@]}" ;;
        "apt-get") sudo apt-get install -y --no-upgrade "${packages[@]}" ;;
        "dnf"|"yum") sudo $PKG_MANAGER install -y "${packages[@]}" ;;
        "pacman") sudo pacman -S --needed --noconfirm "${packages[@]}" ;;
    esac
}

install_dependencies() {
    echo -e "${YELLOW}Checking system dependencies...${NC}"

    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS dependencies
        echo -e "${YELLOW}Installing macOS dependencies...${NC}"
        install_packages go gcc pkg-config
        
        # Install Xcode command line tools if not present
        if ! xcode-select -p &> /dev/null; then
            echo -e "${YELLOW}Installing Xcode command line tools...${NC}"
            xcode-select --install
        fi
    else
        # Common dependencies for Linux
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
    fi
}

create_icns() {
    local png_file="$1"
    local icns_name="$2"
    local iconset_name="${icns_name}.iconset"
    
    # Create iconset directory
    mkdir -p "$iconset_name"
    
    # Generate different icon sizes
    for size in 16 32 64 128 256 512 1024; do
        # Regular size
        sips -z $size $size "$png_file" --out "$iconset_name/icon_${size}x${size}.png" > /dev/null
        # Retina size
        sips -z $((size*2)) $((size*2)) "$png_file" --out "$iconset_name/icon_${size}x${size}@2x.png" > /dev/null
    done
    
    # Convert iconset to icns
    iconutil -c icns "$iconset_name" -o "${icns_name}.icns"
    
    # Clean up iconset directory
    rm -rf "$iconset_name"
}

create_desktop_entry() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo -e "${YELLOW}Creating macOS application bundle...${NC}"
        
        # Create .app bundle structure
        APP_BUNDLE="$DESKTOP_DIR/$APP_NAME.app"
        mkdir -p "$APP_BUNDLE/Contents/MacOS"
        mkdir -p "$APP_BUNDLE/Contents/Resources"
        
        # Create Info.plist
        cat << EOF > "$APP_BUNDLE/Contents/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>$APP_NAME</string>
    <key>CFBundleIconFile</key>
    <string>$APP_NAME</string>
    <key>CFBundleIdentifier</key>
    <string>com.$APP_NAME</string>
    <key>CFBundleName</key>
    <string>$APP_DISPLAY_NAME</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.10</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF
        
        # Copy binary to MacOS directory
        cp "$APP_NAME" "$APP_BUNDLE/Contents/MacOS/"
    else
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
    fi
}

install_icon() {
    if [[ -f "$ICON_SRC" ]]; then
        echo -e "${YELLOW}Installing application icon...${NC}"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # For macOS, create and install .icns file
            echo -e "${YELLOW}Creating macOS icon...${NC}"
            create_icns "$ICON_SRC" "$APP_NAME"
            cp "$APP_NAME.icns" "$APP_BUNDLE/Contents/Resources/"
            rm "$APP_NAME.icns"  # Clean up temporary file
        else
            # For Linux
            sudo mkdir -p "$ICON_DIR"
            sudo cp "$ICON_SRC" "$ICON_DIR/$APP_NAME.png"

            # Update icon cache
            if command -v gtk-update-icon-cache &> /dev/null; then
                echo -e "${YELLOW}Updating icon cache...${NC}"
                sudo gtk-update-icon-cache -f /usr/share/icons/hicolor/
            fi
        fi
    else
        echo -e "${YELLOW}Warning: Icon file not found ($ICON_SRC)${NC}"
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
    install_dependencies

    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go not found!${NC}"
        exit 1
    fi

    echo -e "${YELLOW}Building application with CGO...${NC}"
    CGO_ENABLED=1 go build -o "$APP_NAME"

    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo -e "${YELLOW}Installing application for macOS...${NC}"
        create_desktop_entry
        install_icon
    else
        echo -e "${YELLOW}Installing application for Linux...${NC}"
        sudo install -Dm755 "$APP_NAME" "$BIN_DIR/$APP_NAME"
        create_desktop_entry
        install_icon
    fi

    echo -e "\n${GREEN}Installation completed successfully!${NC}"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "Application has been installed to $DESKTOP_DIR/$APP_NAME.app"
        echo "You can find it in your Applications folder."
    else
        echo "Application should now appear in your application menu."
        echo "You may need to log out and back in for changes to take full effect."
    fi
}

main

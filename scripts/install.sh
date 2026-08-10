#!/bin/bash
set -e

echo "Installing rbxpmux..."

BIN_DIR="$HOME/.local/bin"
APP_DIR="$HOME/.local/share/applications"
ICON_DIR="$HOME/.local/share/pixmaps"

mkdir -p "$BIN_DIR"
mkdir -p "$APP_DIR"
mkdir -p "$ICON_DIR"

# Ensure we are in the directory containing the files
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"

if [ ! -f "$SCRIPT_DIR/rbxpmux" ]; then
    echo "Error: rbxpmux binary not found in $SCRIPT_DIR"
    exit 1
fi

cp "$SCRIPT_DIR/rbxpmux" "$BIN_DIR/"
cp "$SCRIPT_DIR/rbxpmux.png" "$ICON_DIR/"

# Copy desktop file and update paths to be absolute
sed "s|^Exec=.*|Exec=$BIN_DIR/rbxpmux|" "$SCRIPT_DIR/rbxpmux.desktop" > "$APP_DIR/rbxpmux.desktop"
sed -i "s|^Icon=.*|Icon=$ICON_DIR/rbxpmux.png|" "$APP_DIR/rbxpmux.desktop"
chmod +x "$APP_DIR/rbxpmux.desktop"

if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "$APP_DIR"
fi

echo "Installation complete! You can now launch rbxpmux from your application menu."
echo "Note: Make sure $BIN_DIR is in your PATH."

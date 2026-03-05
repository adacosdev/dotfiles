#!/bin/bash
set -eo pipefail

# =============================================================================
# Progress Bar Helper Functions
# =============================================================================
show_progress() {
  local msg="$1"
  local pct="${2:-0}"
  local bar_length=30
  local filled=$((pct * bar_length / 100))
  local empty=$((bar_length - filled))
  
  local bar="["
  for ((i = 0; i < filled; i++)); do bar+="="; done
  for ((i = 0; i < empty; i++)); do bar+="-"; done
  bar+="]"
  
  printf "\r\033[K%s %s %3d%%" "$msg" "$bar" "$pct"
}

show_progress_done() {
  echo ""
  echo "✅ $1"
}

# =============================================================================
# Main Script: Install Nerd Fonts
# =============================================================================
show_progress "📝 Installing fonts" 65

# Check dependencies
install_pkg() {
    if command -v apt &> /dev/null; then
        sudo apt update && sudo apt install -y "$@"
    elif command -v dnf &> /dev/null; then
        sudo dnf install -y "$@"
    elif command -v pacman &> /dev/null; then
        sudo pacman -S --noconfirm "$@"
    else
        echo "❌ No se pudo determinar el gestor de paquetes. Instala $@ manualmente."
        exit 1
    fi
}

deps=("curl" "unzip" "fontconfig")
missing_deps=()

for dep in "${deps[@]}"; do
    if ! command -v "$dep" &> /dev/null && ! dpkg -l 2>/dev/null | grep -q "^ii.*$dep"; then
        missing_deps+=("$dep")
    fi
done

if [ ${#missing_deps[@]} -ne 0 ]; then
    show_progress "📝 Installing fonts: dependencies ${missing_deps[*]}" 67
    install_pkg "${missing_deps[@]}"
fi

# Font installation
FONT_DIR="$HOME/.local/share/fonts"
mkdir -p "$FONT_DIR"

fonts=(
    "JetBrainsMono"
    "FiraCode"
    "Hack"
    "Iosevka"
)

show_progress "📝 Installing fonts: checking" 68
need_update=false

for font in "${fonts[@]}"; do
    if ls "$FONT_DIR" 2>/dev/null | grep -iq "$font"; then
        : # Already installed
    else
        pct=$((68 + $(echo "${fonts[@]}" | grep -o "$font" | head -1 | wc -c)))
        show_progress "📝 Installing fonts: $font" "$pct"
        
        TEMP_DIR=$(mktemp -d)
        URL="https://github.com/ryanoasis/nerd-fonts/releases/latest/download/${font}.zip"
        
        if curl -L "$URL" -o "$TEMP_DIR/font.zip" 2>/dev/null; then
            unzip -o "$TEMP_DIR/font.zip" -d "$TEMP_DIR" >/dev/null 2>&1
            find "$TEMP_DIR" -name "*.[ot]tf" -exec cp {} "$FONT_DIR/" \;
            need_update=true
        fi
        rm -rf "$TEMP_DIR"
    fi
done

if [ "$need_update" = true ]; then
    show_progress "📝 Installing fonts: updating cache" 79
    fc-cache -f
fi

show_progress_done "Fonts installed successfully"

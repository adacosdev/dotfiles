#!/bin/bash

# 1. DETECCIÓN DE DEPENDENCIAS
echo "🔍 Comprobando dependencias del sistema..."

# Función para instalar paquetes según el gestor disponible
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

# Lista de herramientas necesarias
deps=("curl" "unzip" "fontconfig")
missing_deps=()

for dep in "${deps[@]}"; do
    if ! command -v "$dep" &> /dev/null; then
        missing_deps+=("$dep")
    fi
done

if [ ${#missing_deps[@]} -ne 0 ]; then
    echo "📦 Instalando dependencias faltantes: ${missing_deps[*]}"
    install_pkg "${missing_deps[@]}"
fi

# 2. CONFIGURACIÓN DE FUENTES
FONT_DIR="$HOME/.local/share/fonts"
mkdir -p "$FONT_DIR"

fonts=(
    "JetBrainsMono"
    "FiraCode"
    "Hack"
    "Iosevka"
)

echo "📊 Comprobando fuentes..."
need_update=false

for font in "${fonts[@]}"; do
    if ls "$FONT_DIR" | grep -iq "$font"; then
        echo "✅ $font ya está instalada."
    else
        echo "📥 Instalando $font..."
        TEMP_DIR=$(mktemp -d)
        URL="https://github.com/ryanoasis/nerd-fonts/releases/latest/download/${font}.zip"
        
        if curl -L "$URL" -o "$TEMP_DIR/font.zip"; then
            unzip -o "$TEMP_DIR/font.zip" -d "$TEMP_DIR"
            find "$TEMP_DIR" -name "*.[ot]tf" -exec cp {} "$FONT_DIR/" \;
            need_update=true
            echo "✨ $font instalada con éxito."
        else
            echo "❌ Error al descargar $font."
        fi
        rm -rf "$TEMP_DIR"
    fi
done

if [ "$need_update" = true ]; then
    echo "🔄 Actualizando caché de fuentes..."
    fc-cache -f
    echo "🚀 ¡Proceso finalizado!"
else
    echo "😎 Todo al día."
fi

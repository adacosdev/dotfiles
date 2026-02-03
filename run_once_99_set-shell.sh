#!/bin/bash

# Este script cambia la shell por defecto a zsh se hiciese falta
if [[ "$SHELL" != */zsh ]]; then
    if command -v zsh &> /dev/null; then
        echo "🐚 Cambiando la shell por defecto a zsh..."
        chsh -s $(which zsh)
    else
        echo "⚠️ zsh no está instalado. No se puede cambiar la shell."
    fi
fi

#!/bin/bash

# --- Progress Banner ---
echo -e "\033[0;32m"
echo "████████████████████ [100%] Fase 6/6: Finalizando Configuración"
echo -e "\033[0m"

# Este script cambia la shell por defecto a zsh se hiciese falta
if [[ "$SHELL" != */zsh ]]; then
    if command -v zsh &> /dev/null; then
        echo "🐚 Cambiando la shell por defecto a zsh..."
        chsh -s $(which zsh)
    else
        echo "⚠️ zsh no está instalado. No se puede cambiar la shell."
    fi
fi

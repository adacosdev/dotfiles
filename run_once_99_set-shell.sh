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
# Main Script: Set Default Shell to Zsh
# =============================================================================
show_progress "🐚 Finalizing configuration" 100

if [[ "$SHELL" != */zsh ]]; then
    if command -v zsh &> /dev/null; then
        show_progress "🐚 Finalizing configuration: changing shell to zsh" 100
        chsh -s "$(which zsh)"
    else
        echo "⚠️ zsh is not installed. Unable to change default shell."
    fi
fi

show_progress_done "Configuration finalized successfully. Please restart your terminal session."

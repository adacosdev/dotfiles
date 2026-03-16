#!/usr/bin/env bash
# pre-commit hook for chezmoi - auto-imports and commits changes from target dirs
# This runs BEFORE each commit, checking if target files have changed

set -eo pipefail

CHEZMOI_DIR="${CHEZMOI_DIR:-$HOME/.local/share/chezmoi}"

# Directories to watch (relative to home, chezmoi manages these)
TARGET_DIRS=(
  ".config/nvim"
  ".config/opencode"
  ".config/Cursor"
  ".config/zsh"
  ".config/starship.toml"
)

# Check each target directory for changes
CHANGES_FOUND=0

for target in "${TARGET_DIRS[@]}"; do
  source_path="$CHEZMOI_DIR/dot_$(echo "$target" | tr '/' '_' | sed 's/\./_/g' | sed 's/_dot_/dot_/')"
  
  # Handle special cases for file vs directory
  if [[ "$target" == *"."* ]] && [[ "$target" != *"/"* ]]; then
    # It's a file like starship.toml
    basename=$(basename "$target")
    source_path="$CHEZMOI_DIR/dot_config/$basename"
  else
    # It's a directory
    dir_name=$(echo "$target" | tr '/' '_')
    source_path="$CHEZMOI_DIR/dot_config/$dir_name"
  fi
  
  # Check if this source path exists in chezmoi
  if [[ -e "$source_path" ]]; then
    # Get the actual target path
    target_full="$HOME/$target"
    
    # Use chezmoi data to check if it's a private_ file
    if [[ -d "$source_path" ]] && [[ "$(basename "$source_path")" == private_* ]]; then
      # For private directories, check recursively
      if [[ -d "$target_full" ]]; then
        # Compare with rsync dry-run
        if ! rsync -a --delete --dry-run "$target_full/" "$source_path/" >/dev/null 2>&1; then
          echo "📦 Changes detected in $target (private directory)"
          CHANGES_FOUND=1
        fi
      fi
    fi
  fi
done

# If changes found, we could auto-add them
# For now, just warn the user
if [[ $CHANGES_FOUND -eq 1 ]]; then
  echo "⚠️  Target directories have changes. Run 'cm' to auto-commit to chezmoi."
fi

exit 0

#!/usr/bin/env bash
# chezmoi-auto-commit.sh - Auto-commit changes in chezmoi source state
# Usage: Run manually, or add to crontab, or use with inotify

set -euo pipefail

CHEZMOI_DIR="${CHEZMOI_DIR:-$HOME/.local/share/chezmoi}"
AUTO_PUSH="${AUTO_PUSH:-false}"  # Set to "true" to auto-push
BRANCH="${BRANCH:-main}"

cd "$CHEZMOI_DIR"

# Check for changes
if [[ -z "$(git status --porcelain)" ]]; then
  echo "📝 No changes to commit"
  exit 0
fi

# Get list of changed files
CHANGED_FILES=$(git status --porcelain | awk '{print $2}' | head -10)
FILE_COUNT=$(git status --porcelain | wc -l)

# Auto-generate commit message based on changed files
if [[ "$FILE_COUNT" -eq 1 ]]; then
  MSG="chore: update $(echo "$CHANGED_FILES" | head -1 | xargs basename)"
else
  MSG="chore: update $FILE_COUNT files ($(echo "$CHANGED_FILES" | head -3 | xargs basename | tr '\n' ',' | sed 's/,$//'))"
fi

echo "📦 Changes detected:"
git status --short
echo ""
echo "📝 Committing as: $MSG"

# Add all changes and commit
git add -A
git commit -m "$MSG"

if [[ "$AUTO_PUSH" == "true" ]]; then
  echo "🚀 Pushing to remote..."
  git push origin "$BRANCH"
fi

echo "✅ Done!"

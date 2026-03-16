#!/usr/bin/env bash
# chezmoi-watch-daemon.sh - Watches target directories and auto-commits changes
# Usage: Run as systemd user service or manually

set -euo pipefail

CHEZMOI_DIR="${CHEZMOI_DIR:-$HOME/.local/share/chezmoi}"
LOG_FILE="${LOG_FILE:-$HOME/.local/share/chezmoi/auto-commit.log}"

# Directories/files to watch (chezmoi-managed targets)
WATCH_PATHS=(
  "$HOME/.config/nvim"
  "$HOME/.config/opencode"
  "$HOME/.config/Cursor"
  "$HOME/.config/zsh"
  "$HOME/.config/starship.toml"
  "$HOME/.zshrc"
  "$HOME/.gitconfig"
)

# Debounce time (seconds to wait after changes before committing)
DEBOUNCE=5

# Log function
log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Check if inotify-tools is installed
check_dependencies() {
  if ! command -v inotifywait &>/dev/null; then
    log "ERROR: inotify-tools not installed. Install with: sudo pacman -S inotify-tools"
    exit 1
  fi
}

# Get list of watch paths that actually exist
get_watch_paths() {
  local paths=()
  for p in "${WATCH_PATHS[@]}"; do
    if [[ -e "$p" ]]; then
      paths+=("$p")
    fi
  done
  printf '%s\n' "${paths[@]}"
}

# Main daemon loop
main() {
  check_dependencies
  
  log "🚀 Starting chezmoi watch daemon..."
  log "Watching paths: ${WATCH_PATHS[*]}"
  
  # Get existing watch paths
  WATCHABLE=$(get_watch_paths)
  
  if [[ -z "$WATCHABLE" ]]; then
    log "ERROR: No valid paths to watch"
    exit 1
  fi
  
  # Build inotifywait command
  # -m: monitor mode (continuous)
  # -r: recursive
  # -e: events to watch (modify, create, delete, move)
  # --exclude: patterns to ignore
  local inotify_cmd="inotifywait -m -r -e modify,create,delete,move,close_write"
  local exclude_pattern="--exclude '(\.git|lazy-lock\.json|plugin/packer\.lock|undo|swap)'"
  
  log "📡 Watching for changes..."
  
  # Use a file to track if we need to commit
  local trigger_file="/tmp/chezmoi-watch-trigger"
  
  # Trap to cleanup
  trap 'log "🛑 Daemon stopped"; rm -f "$trigger_file"; exit 0' SIGINT SIGTERM
  
  # Start watching in background, write to a named pipe
  local fifo="/tmp/chezmoi-watch-fifo-$$"
  mkfifo "$fifo"
  
  # Start inotify in background
  eval "$inotify_cmd $exclude_pattern $WATCHABLE" > "$fifo" 2>/dev/null &
  local inotify_pid=$!
  
  # Function to do the commit
  do_commit() {
    # Small debounce
    sleep "$DEBOUNCE"
    
    # Check for changes in chezmoi source dir
    cd "$CHEZMOI_DIR"
    
    if [[ -z "$(git status --porcelain 2>/dev/null)" ]]; then
      log "No changes to commit"
      return
    fi
    
    local changed_files
    changed_files=$(git status --porcelain | awk '{print $2}' | head -5 | xargs -I{} basename {} 2>/dev/null | tr '\n' ',' | sed 's/,$//')
    local file_count
    file_count=$(git status --porcelain | wc -l)
    
    local msg
    if [[ "$file_count" -eq 1 ]]; then
      msg="chore: update $changed_files"
    else
      msg="chore: update $file_count files ($changed_files)"
    fi
    
    log "📦 Committing: $msg"
    
    # Add all changes and commit
    git add -A
    git commit -q -m "$msg"
    
    # Auto-push (optional - comment out if not wanted)
    if git config --get remote.origin.url &>/dev/null; then
      git push -q origin main 2>/dev/null || log "⚠️  Push failed (no remote or no network)"
    fi
    
    log "✅ Committed and pushed: $msg"
  }
  
  # Read events from fifo with timeout
  while true; do
    # Check if inotify is still running
    if ! kill -0 "$inotify_pid" 2>/dev/null; then
      log "ERROR: inotifywait died, restarting..."
      eval "$inotify_cmd $exclude_pattern $WATCHABLE" > "$fifo" 2>/dev/null &
      inotify_pid=$!
    fi
    
    # Read with timeout using read -t
    if read -t 2 line < "$fifo" 2>/dev/null; then
      log "📝 Change detected: $line"
      # Trigger commit in background
      do_commit &
    fi
  done
  
  # Cleanup (unreachable unless loop breaks)
  rm -f "$fifo"
  kill "$inotify_pid" 2>/dev/null || true
}

# Run as daemon (background)
run_daemon() {
  main "$@" &
  log "Daemon PID: $!"
  disown
  echo "Daemon started in background. PID: $!"
}

# Run once (check and commit)
run_once() {
  log "Running single check..."
  cd "$CHEZMOI_DIR"
  
  if [[ -n "$(git status --porcelain)" ]]; then
    local changed_files
    changed_files=$(git status --porcelain | awk '{print $2}' | head -3 | xargs -I{} basename {} | tr '\n' ',' | sed 's/,$//')
    local msg="chore: auto-update $(date '+%Y-%m-%d')"
    
    git add -A
    git commit -m "$msg"
    log "✅ Committed: $msg"
  else
    log "No changes"
  fi
}

# Parse arguments
case "${1:-daemon}" in
  daemon)
    main
    ;;
  start)
    run_daemon
    ;;
  once)
    run_once
    ;;
  *)
    echo "Usage: $0 {daemon|start|once}"
    exit 1
    ;;
esac

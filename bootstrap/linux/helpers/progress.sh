#!/bin/bash

# Shared progress helpers for Linux bootstrap entrypoints.

show_progress() {
  local msg="$1"
  local pct="${2:-0}"
  local bar_length=30
  local filled=$((pct * bar_length / 100))
  local empty=$((bar_length - filled))
  local bar="["
  local i

  for ((i = 0; i < filled; i++)); do bar+="="; done
  for ((i = 0; i < empty; i++)); do bar+="-"; done
  bar+="]"

  printf "\r\033[K%s %s %3d%%" "$msg" "$bar" "$pct"
}

show_progress_done() {
  echo ""
  echo "✅ $1"
}

PROGRESS_FILE=""
ACTIVITY_LOG=""
PROGRESS_PID=""

setup_live_progress_files() {
  local suffix="${1:-$$}"
  PROGRESS_FILE="/tmp/install-progress-${suffix}"
  ACTIVITY_LOG="/tmp/install-activity-${suffix}"
}

update_progress() {
  local pct="$1"
  local msg="$2"

  echo "$pct" > "$PROGRESS_FILE"
  echo "$msg" > "$PROGRESS_FILE.msg"
}

log_activity() {
  local msg="$1"

  echo "  📝 $msg" >> "$ACTIVITY_LOG"
}

render_progress() {
  local pct msg bar_length filled empty bar i

  pct=$(cat "$PROGRESS_FILE" 2>/dev/null || echo "0")
  msg=$(cat "$PROGRESS_FILE.msg" 2>/dev/null || echo "Processing")
  bar_length=30
  filled=$((pct * bar_length / 100))
  empty=$((bar_length - filled))
  bar="["

  for ((i = 0; i < filled; i++)); do bar+="="; done
  for ((i = 0; i < empty; i++)); do bar+="-"; done
  bar+="]"

  printf "\033[2J\033[H"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  printf "📦 %s\n" "$msg"
  printf "%s %3d%%\n" "$bar" "$pct"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
  echo "Activity Log:"

  if [[ -f "$ACTIVITY_LOG" ]]; then
    tail -8 "$ACTIVITY_LOG"
  fi

  echo ""
}

start_progress_renderer() {
  while true; do
    render_progress
    sleep 1
  done
}

init_progress() {
  setup_live_progress_files "$1"
  echo "0" > "$PROGRESS_FILE"
  : > "$ACTIVITY_LOG"
  start_progress_renderer &
  PROGRESS_PID=$!
}

cleanup_progress() {
  if [[ -n "$PROGRESS_PID" ]]; then
    kill "$PROGRESS_PID" 2>/dev/null || true
  fi

  rm -f "$PROGRESS_FILE" "$PROGRESS_FILE.msg" "$ACTIVITY_LOG"
}

#!/usr/bin/env bash
# chezmoi-config-sync: Sync UI configuration changes back to chezmoi dotfiles
# Usage: chezmoi-config-sync [--auto-commit] [--push]
# 
# This script detects changes in managed application configs (VS Code, Cursor, Warp, etc.)
# and syncs them back to the chezmoi repository, optionally committing and pushing changes.
#
# The script works by:
# 1. Tracking file modification times since last sync (using ~/.cache/chezmoi-sync-timestamp)
# 2. For changed files, syncing to the chezmoi directory and preserving templates
# 3. Validating all templates with 'chezmoi dump' before committing
# 4. Creating git commits with 'sync:' prefix for easy filtering
# 5. Optionally pushing changes to the remote repository
#
# Managed applications:
# - VS Code & Cursor & Antigravity: Settings, keybindings, snippets
# - Warp Terminal: User preferences (managed keys only)
# - Git: ~/.gitconfig
# - Starship: ~/.config/starship.toml
# - Zsh: Shell configuration files
#
# Exit codes:
#  0: Success
#  1: Template validation failed or git error

set -eo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CHEZMOI_DIR="${CHEZMOI_DIR:-$HOME/.local/share/chezmoi}"
HOME_DIR="$HOME"
SYNC_TIMESTAMP_FILE="$HOME/.cache/chezmoi-sync-timestamp"
AUTO_COMMIT=false
PUSH_CHANGES=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --auto-commit) AUTO_COMMIT=true; shift ;;
    --push) PUSH_CHANGES=true; shift ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# Ensure chezmoi directory exists
if [[ ! -d "$CHEZMOI_DIR" ]]; then
  echo -e "${RED}Error: Chezmoi directory not found: $CHEZMOI_DIR${NC}"
  exit 1
fi

# Initialize sync timestamp file
mkdir -p "$(dirname "$SYNC_TIMESTAMP_FILE")"
if [[ ! -f "$SYNC_TIMESTAMP_FILE" ]]; then
  touch "$SYNC_TIMESTAMP_FILE"
fi

echo -e "${BLUE}Starting chezmoi config sync...${NC}\n"

# Track if any changes were made
CHANGES_MADE=false
SYNC_LOG="/tmp/chezmoi-sync-$(date +%s).log"

# Helper function to check and sync a config file
sync_config() {
  local source="$1"
  local dest="$2"
  local config_name="$3"
  
  if [[ ! -f "$source" ]]; then
    return 0
  fi
  
  # Create destination directory if needed
  mkdir -p "$(dirname "$dest")"
  
  # Check if source is newer than last sync
  if [[ "$source" -nt "$SYNC_TIMESTAMP_FILE" ]]; then
    echo -e "${YELLOW}[Sync]${NC} $config_name"
    
    # Back up old version if exists
    if [[ -f "$dest" ]]; then
      cp "$dest" "${dest}.backup-$(date +%s)"
    fi
    
    # Copy file to chezmoi (preserve .tmpl extension if applicable)
    cp "$source" "$dest"
    echo "  ✓ Synced: $source → $dest" | tee -a "$SYNC_LOG"
    CHANGES_MADE=true
  fi
}

# VS Code Settings
echo -e "${BLUE}Checking VS Code configuration...${NC}"
sync_config \
  "$HOME_DIR/.config/Code/User/settings.json" \
  "$CHEZMOI_DIR/dot_config/private_Code/User/settings.json.tmpl" \
  "VS Code settings"

sync_config \
  "$HOME_DIR/.config/Code/User/keybindings.json" \
  "$CHEZMOI_DIR/dot_config/private_Code/User/keybindings.json" \
  "VS Code keybindings"

sync_config \
  "$HOME_DIR/.config/Code/User/snippets" \
  "$CHEZMOI_DIR/dot_config/private_Code/User/snippets" \
  "VS Code snippets"

# Cursor IDE Settings
echo -e "${BLUE}Checking Cursor IDE configuration...${NC}"
sync_config \
  "$HOME_DIR/.config/Cursor/User/settings.json" \
  "$CHEZMOI_DIR/dot_config/private_Cursor/User/settings.json.tmpl" \
  "Cursor settings"

sync_config \
  "$HOME_DIR/.config/Cursor/User/keybindings.json" \
  "$CHEZMOI_DIR/dot_config/private_Cursor/User/keybindings.json" \
  "Cursor keybindings"

# Antigravity Settings
echo -e "${BLUE}Checking Antigravity configuration...${NC}"
sync_config \
  "$HOME_DIR/.config/Antigravity/User/settings.json" \
  "$CHEZMOI_DIR/dot_config/private_Antigravity/User/settings.json.tmpl" \
  "Antigravity settings"

sync_config \
  "$HOME_DIR/.config/Antigravity/User/keybindings.json" \
  "$CHEZMOI_DIR/dot_config/private_Antigravity/User/keybindings.json" \
  "Antigravity keybindings"

# Warp Terminal Settings
# Note: Warp stores both runtime state and user preferences in a single JSON file.
# We only sync the managed keys and preserve template variables (e.g., {{ .chezmoi.homeDir }})
echo -e "${BLUE}Checking Warp Terminal configuration...${NC}"
if [[ -f "$HOME_DIR/.config/warp-terminal/user_preferences.json" ]]; then
  # Extract only the user preferences we care about and preserve the template variables
  warp_dest="$CHEZMOI_DIR/dot_config/warp-terminal/user_preferences.json.tmpl"
  
  if [[ "$HOME_DIR/.config/warp-terminal/user_preferences.json" -nt "$SYNC_TIMESTAMP_FILE" ]]; then
    echo -e "${YELLOW}[Sync]${NC} Warp Terminal preferences"
    
    # Create backup
    if [[ -f "$warp_dest" ]]; then
      cp "$warp_dest" "${warp_dest}.backup-$(date +%s)"
    fi
    
    # Sync: This needs careful handling since Warp stores runtime data mixed with config
    # Strategy: Extract only managed keys and preserve any template variables (like {{ .chezmoi.homeDir }})
    python3 << 'EOF'
import json
import sys
import re
import os

try:
    # Get variables from environment or use defaults
    home_dir = os.environ.get('HOME', os.path.expanduser('~'))
    chezmoi_dir = os.environ.get('CHEZMOI_DIR', os.path.join(home_dir, '.local/share/chezmoi'))
    
    # Read the current user preferences from Warp's config directory
    with open(f'{home_dir}/.config/warp-terminal/user_preferences.json', 'r') as f:
        warp_prefs = json.load(f)
    
    # Read existing template to preserve any chezmoi template variables
    # This allows us to keep things like {{ .chezmoi.homeDir }} in the MCP path
    try:
        with open(f'{chezmoi_dir}/dot_config/warp-terminal/user_preferences.json.tmpl', 'r') as f:
            template_content = f.read()
    except:
        template_content = None
    
    # Only sync these keys to avoid syncing transient runtime state
    # (API quotas, UI state, experimental flags, etc.)
    managed_keys = {
        'Theme', 'OverrideOpacity', 'TelemetryEnabled', 'CrashReportingEnabled',
        'Notifications', 'NLDInTerminalEnabled', 'WorkflowsBoxOpen',
        'HasAutoOpenedConversationList', 'CloudConversationStorageEnabled',
        'IsSettingsSyncEnabled', 'InputBoxTypeSetting', 'CustomSecretRegexList'
    }
    
    # Extract MCP path from template if it exists
    mcp_path = None
    if template_content:
        mcp_match = re.search(r'"MCPExecutionPath":\s*"([^"]*)"', template_content)
        if mcp_match:
            mcp_path = mcp_match.group(1)
    
    # Build new preferences object with managed keys only
    new_prefs = {'prefs': {}}
    for key in managed_keys:
        if key in warp_prefs['prefs']:
            new_prefs['prefs'][key] = warp_prefs['prefs'][key]
    
    # Preserve MCP path if it had template variables
    if mcp_path and '{{' in mcp_path:
        new_prefs['prefs']['MCPExecutionPath'] = mcp_path
    elif 'MCPExecutionPath' in warp_prefs['prefs']:
        new_prefs['prefs']['MCPExecutionPath'] = warp_prefs['prefs']['MCPExecutionPath']
    
    # Write back to chezmoi
    with open(f'{chezmoi_dir}/dot_config/warp-terminal/user_preferences.json.tmpl', 'w') as f:
        json.dump(new_prefs, f, indent=2)
    
    print("  ✓ Synced Warp preferences")
except Exception as e:
    print(f"  ✗ Error syncing Warp: {e}", file=sys.stderr)
    sys.exit(1)
EOF
    
    CHANGES_MADE=true
  fi
fi

# Git configuration
echo -e "${BLUE}Checking Git configuration...${NC}"
if [[ -f "$HOME_DIR/.gitconfig" ]]; then
  if [[ "$HOME_DIR/.gitconfig" -nt "$SYNC_TIMESTAMP_FILE" ]]; then
    echo -e "${YELLOW}[Sync]${NC} Git configuration"
    cp "$HOME_DIR/.gitconfig" "$CHEZMOI_DIR/dot_gitconfig"
    echo "  ✓ Synced: $HOME_DIR/.gitconfig" | tee -a "$SYNC_LOG"
    CHANGES_MADE=true
  fi
fi

# Starship configuration
echo -e "${BLUE}Checking Starship configuration...${NC}"
if [[ -f "$HOME_DIR/.config/starship.toml" ]]; then
  if [[ "$HOME_DIR/.config/starship.toml" -nt "$SYNC_TIMESTAMP_FILE" ]]; then
    echo -e "${YELLOW}[Sync]${NC} Starship configuration"
    cp "$HOME_DIR/.config/starship.toml" "$CHEZMOI_DIR/dot_config/starship.toml.tmpl"
    echo "  ✓ Synced: $HOME_DIR/.config/starship.toml" | tee -a "$SYNC_LOG"
    CHANGES_MADE=true
  fi
fi

# Zsh configuration
echo -e "${BLUE}Checking Zsh configuration...${NC}"
for file in aliases functions plugins; do
  src="$HOME_DIR/.config/zsh/${file}.zsh"
  if [[ -f "$src" ]]; then
    if [[ "$src" -nt "$SYNC_TIMESTAMP_FILE" ]]; then
      echo -e "${YELLOW}[Sync]${NC} Zsh $file configuration"
      cp "$src" "$CHEZMOI_DIR/dot_config/zsh/${file}.zsh.tmpl"
      echo "  ✓ Synced: $src" | tee -a "$SYNC_LOG"
      CHANGES_MADE=true
    fi
  fi
done

# Validate all templates
echo -e "\n${BLUE}Validating chezmoi templates...${NC}"
if cd "$CHEZMOI_DIR" && chezmoi dump --format=json > /dev/null 2>&1; then
  echo -e "${GREEN}✓ All templates are valid${NC}"
else
  echo -e "${RED}✗ Template validation failed!${NC}"
  echo "  Run 'chezmoi dump --format=json' for details"
  exit 1
fi

# Git operations: Stage, commit, and optionally push changes
# The --auto-commit flag determines whether we automatically commit or just show instructions
if [[ "$CHANGES_MADE" == true ]]; then
  echo -e "\n${BLUE}Processing changes...${NC}"
  
  cd "$CHEZMOI_DIR"
  
  # Display the changes to the user
  echo -e "${YELLOW}Changes detected:${NC}"
  git status --short
  
  if [[ "$AUTO_COMMIT" == true ]]; then
    # Stage all changes and create a commit with meaningful message
    echo -e "\n${BLUE}Auto-committing changes...${NC}"
    
    git add -A
    
    # Create a commit message with 'sync:' prefix for easy filtering
    # Includes timestamp and list of first 5 changed files
    summary=$(git diff --cached --name-only | head -5 | tr '\n' ', ' | sed 's/,$//')
    commit_msg="sync: Update configs from UI changes ($(date +%Y-%m-%d\ %H:%M:%S))

Updated files:
$summary"
    
    if git commit -m "$commit_msg"; then
      echo -e "${GREEN}✓ Committed changes${NC}"
      
      if [[ "$PUSH_CHANGES" == true ]]; then
        echo -e "${BLUE}Pushing to remote...${NC}"
        if git push; then
          echo -e "${GREEN}✓ Pushed to remote${NC}"
        else
          echo -e "${YELLOW}⚠ Could not push (offline or no remote)${NC}"
        fi
      fi
    fi
  else
    echo -e "\n${YELLOW}Review changes with:${NC}"
    echo "  cd $CHEZMOI_DIR && git diff --cached"
    echo -e "\n${YELLOW}To stage and commit changes:${NC}"
    echo "  chezmoi-config-sync --auto-commit [--push]"
  fi
  
  # Update sync timestamp
  touch "$SYNC_TIMESTAMP_FILE"
else
  echo -e "${GREEN}✓ All configurations up to date${NC}"
fi

# Show log
if [[ -f "$SYNC_LOG" ]]; then
  echo -e "\n${BLUE}Sync log:${NC}"
  cat "$SYNC_LOG"
fi

echo -e "\n${GREEN}Sync complete!${NC}"

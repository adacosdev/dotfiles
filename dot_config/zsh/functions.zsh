# =============================================================================
# ~/.config/zsh/functions.zsh — Utility Functions
# =============================================================================

# Open files in VS Code with fuzzy find
cf() { # > Find file and open in VS Code
  local file
  file=$(fd --type f --strip-cwd-prefix --hidden --exclude .git | \
    fzf --preview 'bat --color=always --style=numbers --line-range :500 {}' \
        --header "Select file to open in VS Code")
  
  [[ -n "$file" ]] && code "$file"
}

# Fuzzy find and jump into a directory
zf() { # > Find folder and jump into it
  local dir
  dir=$(fd --type d --strip-cwd-prefix --hidden --exclude .git | \
    fzf --header "Select directory")
  
  [[ -n "$dir" ]] && z "$dir"
}

# Search projects directory and open in VS Code
p() { # > Search project and open in VS Code
  local dir
  dir=$(fd --type d --max-depth 2 . ~/Dev 2>/dev/null | \
    fzf --header "Select project")
  
  # Only change directory and open if a selection was made
  [[ -n "$dir" ]] && cd "$dir" && code .
}

# Show aliases and functions with descriptions
h() { # > Show aliases and functions help
  if ! command -v fzf &>/dev/null; then
    echo "⚠️ fzf is required for this command."
    return 1
  fi

  (
    # Extract aliases with descriptions
    grep -Eh '^alias ' $HOME/.config/zsh/*.zsh 2>/dev/null | perl -ne '
      if (/alias\s+([^=]+)=\x27?([^\x27#]+)\x27?\s*#\s*>\s*(.*)/) {
        printf "\033[32m%-12s\033[0m \033[34m→\033[0m \033[37m%-55s\033[0m \033[90m# %s\033[0m\n", $1, $2, $3;
      }'

    # Extract functions with descriptions
    grep -Eh '^[a-zA-Z0-9_-]+\s*\(\)\s*\{' $HOME/.config/zsh/*.zsh 2>/dev/null | perl -ne '
      if (/^([a-zA-Z0-9_-]+)\s*\(\)\s*\{\s*#\s*>\s*(.*)/) {
        printf "\033[36m%-12s\033[0m \033[34m(f)\033[0m \033[37m%-55s\033[0m \033[90m# %s\033[0m\n", $1, "function", $2;
      }'
  ) | sort | fzf \
    --ansi \
    --height 60% \
    --reverse \
    --border \
    --header " 💊 DOTFILES CHEATSHEET " \
    --color="header:italic:cyan,border:blue,fg+:yellow,pointer:magenta" \
    --prompt "Search > " \
    --preview 'echo {} | perl -pe "s/.*# //"' \
    --preview-window=up:wrap
}

# Interactive history search with fzf
hs() { # > Interactive history search with fzf
  if ! command -v fzf &>/dev/null; then
    history 0
    return
  fi
  
  local cmd
  # fc -l (list) -n (no numbers) -r (reverse order)
  cmd=$(fc -lnr 1 | \
    fzf --no-sort --exact --query "$1" \
        --header "Press ENTER to add to command line for editing")
  
  if [[ -n "$cmd" ]]; then
    # Use print -z to add to command line for review, not execute
    print -z "$cmd"
  fi
}

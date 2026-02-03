# Open files in VS Code with fuzzy find (CTRL+P in terminal)
# Just type 'cf' and start typing the filename
cf() { # > Find file and open in VS Code
  local file
  file=$(fd --type f --strip-cwd-prefix --hidden --exclude .git | fzf --preview 'bat --color=always --style=numbers --line-range :500 {}')
  [ -n "$file" ] && code "$file"
}

# Bonus: Quickly find and move to a folder
zf() { # > Find folder and jump into it
  local dir
  dir=$(fd --type d --strip-cwd-prefix --hidden --exclude .git | fzf)
  [ -n "$dir" ] && z "$dir"
}


p() { # > Search project and open in VS Code
  cd ~/Dev && cd $(fd --type d --max-depth 2 | fzf) && code .
}

# Unalias h if it exists to avoid conflicts
unalias h 2>/dev/null || true

h() { # > Show aliases and functions help
  if ! command -v fzf &> /dev/null; then
      echo "⚠️ fzf is required for this command."
      return 1
  fi

  (
    # Extract aliases with descriptions
    grep -Eh '^alias ' $HOME/.config/zsh/*.zsh | perl -ne '
      if (/alias\s+([^=]+)=\x27?([^\x27#]+)\x27?\s*#\s*>\s*(.*)/) {
        printf "\033[32m%-12s\033[0m \033[34m→\033[0m \033[37m%-55s\033[0m \033[90m# %s\033[0m\n", $1, $2, $3;
      }'

    # Extract functions with descriptions
    grep -Eh '^[a-zA-Z0-9_-]+\s*\(\)\s*\{' $HOME/.config/zsh/*.zsh | perl -ne '
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
    --preview 'echo {} | ack -o "(?<=# ).*"' \
    --preview-window=up:wrap
}

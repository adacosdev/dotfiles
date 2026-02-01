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

h() { # > Show aliases and functions help
  grep -E '^(alias|.*(\(\)\s*\{))' "$HOME/.config/zsh/aliases.zsh" "$HOME/.config/zsh/functions.zsh" | \
  sed -E 's/alias //g; s/\(\)\s*\{//g' | \
  column -t -s '#' | \
  fzf --height 40% --reverse --header "Personal User Manual"
}

p() { # > Search project and open in VS Code
  cd ~/Dev && cd $(fd --type d --max-depth 2 | fzf) && code .
}

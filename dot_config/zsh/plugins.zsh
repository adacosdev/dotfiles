# =============================================================================
# ~/.config/zsh/plugins.zsh — Plugin Initialization
# =============================================================================

# ---------------------------------------------------------------------------
# Starship Prompt
# ---------------------------------------------------------------------------
if command -v starship &>/dev/null; then
  eval "$(starship init zsh)"
fi

# ---------------------------------------------------------------------------
# Zoxide (smart directory navigation)
# ---------------------------------------------------------------------------
if command -v zoxide &>/dev/null; then
  eval "$(zoxide init zsh)"
fi

# ---------------------------------------------------------------------------
# Completion System (Cached compinit for performance)
# ---------------------------------------------------------------------------
autoload -Uz compinit
# Regenerate completions only once per day (~24 hours)
if [[ -n ${ZDOTDIR:-$HOME}/.zcompdump(#qN.mh+24) ]]; then
  compinit
else
  compinit -C  # Use cached completions
fi

# ---------------------------------------------------------------------------
# Autosuggestions Plugin (must load BEFORE syntax highlighting)
# ---------------------------------------------------------------------------
_plugin_paths=(
  "/usr/share/zsh/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh"
  "/usr/share/zsh-autosuggestions/zsh-autosuggestions.zsh"
  "$HOME/.zsh/zsh-autosuggestions/zsh-autosuggestions.zsh"
)
for _plugin_path in "${_plugin_paths[@]}"; do
  if [[ -f "$_plugin_path" ]]; then
    source "$_plugin_path"
    break
  fi
done

# ---------------------------------------------------------------------------
# fzf Integration (Keybindings + Completion)
# ---------------------------------------------------------------------------
if command -v fzf &>/dev/null; then
  # Try fzf 0.48+ built-in init (recommended)
  if eval "$(fzf --zsh 2>/dev/null)" 2>/dev/null; then
    # Built-in fzf init worked
    :
  else
    # Fallback for older fzf versions — manually source keybindings/completion
    # Arch/EndeavourOS
    if [[ -f /usr/share/fzf/key-bindings.zsh ]]; then
      source /usr/share/fzf/key-bindings.zsh
      [[ -f /usr/share/fzf/completion.zsh ]] && source /usr/share/fzf/completion.zsh
    # Ubuntu/Debian
    elif [[ -f /usr/share/doc/fzf/examples/key-bindings.zsh ]]; then
      source /usr/share/doc/fzf/examples/key-bindings.zsh
      [[ -f /usr/share/doc/fzf/examples/completion.zsh ]] && source /usr/share/doc/fzf/examples/completion.zsh
    fi
  fi

  # fzf Configuration
  export FZF_DEFAULT_OPTS="--height 40% --layout=reverse --border --info=inline"
  
  # Use fd for better default command if available
  if command -v fd &>/dev/null; then
    export FZF_DEFAULT_COMMAND="fd --type f --strip-cwd-prefix --hidden --exclude .git"
    export FZF_CTRL_T_COMMAND="$FZF_DEFAULT_COMMAND"
    export FZF_ALT_C_COMMAND="fd --type d --strip-cwd-prefix --hidden --exclude .git"
  fi
fi

# ---------------------------------------------------------------------------
# Syntax Highlighting Plugin (MUST be last plugin loaded)
# ---------------------------------------------------------------------------
_plugin_paths=(
  "/usr/share/zsh/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
  "/usr/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
  "$HOME/.zsh/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
)
for _plugin_path in "${_plugin_paths[@]}"; do
  if [[ -f "$_plugin_path" ]]; then
    source "$_plugin_path"
    break
  fi
done

# Clean up temporary variables
unset _plugin_path _plugin_paths

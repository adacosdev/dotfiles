# Init tools if they exist
if command -v starship &> /dev/null; then
  eval "$(starship init zsh)"
fi

if command -v zoxide &> /dev/null; then
  eval "$(zoxide init zsh)"
fi

# Plugins paths (Common locations)
PLUGINS=(
    "/usr/share/zsh/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
    "/usr/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
    "$HOME/.zsh/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
    "/usr/share/zsh/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh"
    "/usr/share/zsh-autosuggestions/zsh-autosuggestions.zsh"
    "$HOME/.zsh/zsh-autosuggestions/zsh-autosuggestions.zsh"
)

for plugin in $PLUGINS; do
    if [ -f "$plugin" ]; then
        source "$plugin"
    fi
done

autoload -Uz compinit && compinit

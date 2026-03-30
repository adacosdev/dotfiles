{{- if eq .chezmoi.os "linux" }}
#!/bin/bash

set -e

{{- $shell_path := "/usr/bin/zsh" }}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🐚 Setting default shell to zsh"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [[ -f "$HOME/.zshrc" ]] && [[ ! -f "$HOME/.zshrc.pre-chezmoi" ]]; then
  echo "  ⏩ .zshrc already exists — skipping"
  return 0
fi

if [[ "$SHELL" == "$shell_path" ]]; then
  echo "  ⏩ zsh is already the default shell"
  return 0
fi

if command -v chsh &>/dev/null; then
  echo "  🔧 Running: chsh -s $shell_path"
  if [[ $EUID -eq 0 ]]; then
    chsh -s "$shell_path"
  else
    sudo chsh -s "$shell_path" "$USER"
  fi
  echo "  ✅ Default shell set to zsh"
  echo "  ⚠️  Log out and back in for changes to take effect."
else
  echo "  ⚠️  chsh not available — add 'exec zsh' to your ~/.bashrc manually"
fi

{{- end }}

{{- if eq .chezmoi.os "darwin" }}
#!/bin/bash

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🐚 Setting default shell to zsh (macOS)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [[ -f "$HOME/.zshrc" ]] && [[ ! -f "$HOME/.zshrc.pre-chezmoi" ]]; then
  echo "  ⏩ .zshrc already exists — skipping"
  return 0
fi

# On macOS, zsh is usually already the default on modern versions
# But we can still set it explicitly via dscl or chsh
if command -v chsh &>/dev/null; then
  echo "  🔧 Running: chsh -s /bin/zsh"
  chsh -s /bin/zsh
  echo "  ✅ Default shell set to zsh"
  echo "  ⚠️  Open a new terminal for changes to take effect."
else
  echo "  ⚠️  chsh not available on macOS (use System Preferences → Users & Groups)"
fi

{{- end }}

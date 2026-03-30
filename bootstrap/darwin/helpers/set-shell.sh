{{- if eq .chezmoi.os "darwin" -}}
#!/bin/bash

{{-   if .headless }}
set -e
{{-   end }}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🐚 Setting default shell to zsh (macOS)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# On modern macOS (Catalina+) zsh is already the default.
# We still set it explicitly to ensure consistency.
if [[ "$SHELL" == "/bin/zsh" ]]; then
  echo "  ⏩ zsh is already the default shell"
  return 0
fi

if command -v chsh &>/dev/null; then
  echo "  🔧 Running: chsh -s /bin/zsh"
  chsh -s /bin/zsh
  echo "  ✅ Default shell set to zsh"
  echo "  ⚠️  Open a new terminal for changes to take effect."
else
  echo "  ⚠️  chsh not available — use System Preferences → Users & Groups"
fi

{{- end }}

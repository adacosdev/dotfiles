# macOS Bootstrap

Provisioning for macOS workstations — Homebrew, editors, shells, fonts.

## Structure

```
bootstrap/darwin/
├── README.md                          ← you are here
├── bootstrap-darwin.sh.tmpl           ← main orchestrator (run via chezmoi apply)
└── helpers/
    ├── homebrew.sh.tmpl               ← Homebrew + packages
    ├── fonts.sh.tmpl                  ← Nerd Fonts (JetBrains Mono, Iosevka Term NF)
    ├── extensions.sh.tmpl             ← VS Code, Cursor, Antigravity extensions
    └── set-shell.sh                   ← ensure zsh is default shell
```

## Prerequisites

1. **Xcode Command Line Tools** — run once:
   ```bash
   xcode-select --install
   ```
   Follow the GUI prompt to complete installation.

2. **Clone dotfiles** (if not already):
   ```bash
   git clone https://github.com/adacosdev/dotfiles.git ~/.local/share/chezmoi
   ```

3. **Apply chezmoi**:
   ```bash
   chezmoi apply
   ```

## What it does

| Helper           | Responsibility                                                        |
| ---------------- | -------------------------------------------------------------------- |
| `homebrew.sh`    | Installs Homebrew (Apple Silicon + Intel), then formula packages     |
| `fonts.sh`       | Downloads and installs JetBrains Mono + Iosevka Term Nerd Font       |
| `extensions.sh`  | Installs VS Code, Cursor, and Antigravity extensions from `.chezmoidata.yaml` |
| `set-shell.sh`   | Ensures zsh is the default shell (macOS Catalina+ already uses zsh) |

## Key differences from Linux

- **No `sudo`** — Homebrew manages everything in `~/` on Apple Silicon or `/usr/local` on Intel
- **No system package manager** — Homebrew is the single package source
- **Fonts go to `~/Library/Fonts`** — not `~/.local/share/fonts`
- **`fc-cache`** is available but macOS font system handles most font discovery automatically

## After bootstrap

```bash
# Add Homebrew to your PATH (Apple Silicon)
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
eval "$(/opt/homebrew/bin/brew shellenv)"

# Restart terminal or source
exec zsh

# Install nvim plugins (first launch)
nvim

# Install tmux plugins (inside tmux session)
tmux
# Press Ctrl+A I to install TPM plugins
```

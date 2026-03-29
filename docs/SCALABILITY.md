# 📈 Scalability & Maintenance

This repository is designed to scale by separating **portable config**,
**platform-specific config**, and **bootstrap data**. Current provisioning is
still strongest on Linux (**Arch/EndeavourOS** and **Ubuntu/Debian**), but the
data model now prepares for `darwin` and `windows` too.

## 📦 Package Management

Portable package data is centralized in `.chezmoidata.yaml` so installation
scripts can stay focused on execution policy.

### Structure of `.chezmoidata.yaml`

```yaml
packages:
  common:
    cli:
      - git
      - zsh
  linux:
    ubuntu:
      - build-essential
    arch:
      - yay
```

### How to add a package

1. Open `.chezmoidata.yaml`.
2. Decide whether the package is:
   - portable CLI tooling → `packages.common.cli`
   - Linux distro specific → `packages.linux.<distro>`
3. Add the exact package name.
   - On **Arch**, the script uses `yay` (supports official repos and AUR).
   - On **Ubuntu**, the script uses `apt`.

### How to add a new Distribution (e.g., Fedora)

1. Add a new list in `.chezmoidata.yaml`:
   ```yaml
   packages:
     linux:
       fedora:
         - git
         - zsh
   ```
2. Edit `run_onchange_00_install-packages.sh.tmpl`:
   - Add a conditional block:
     ```bash
     {{ else if eq .chezmoi.osRelease.id "fedora" -}}
     packages=(
     {{ range .packages.common.cli -}}
       {{ . }}
     {{ end -}}
     {{ range .packages.linux.fedora -}}
       {{ . }}
     {{ end -}}
     )
     sudo dnf install -y "${packages[@]}"
     ```

## 🔌 Extension Management

Editor extensions are also centralized in `.chezmoidata.yaml`:

```yaml
extensions:
  common:
    editors:
      - dbaeumer.vscode-eslint
      - esbenp.prettier-vscode
```

The sync script installs this shared list into any supported editor CLI found on
the machine (`code`, `cursor`, `antigravity`).

## 🖥️ Script Configuration

Scripts use `chezmoi` templates (`.tmpl`). You can use Go template logic to condition execution.

### Useful Variables

- `{{ .chezmoi.os }}`: `linux`, `darwin` (macOS), `windows`.
- `{{ .chezmoi.osRelease.id }}`: `arch`, `ubuntu`, `debian`, `fedora`.
- `{{ .entorno }}`: Custom variable defined during init (`personal` or `adaion`).

### Current conventions

- `packages.common.*` → portable install lists
- `packages.linux.*` → distro-specific Linux install lists
- `extensions.common.*` → shared editor extension sets
- `platforms.*` → capability placeholders for future darwin/windows layering

Example usage in a script:
```bash
{{ if eq .entorno "adaion" }}
# Work specific configuration
{{ end }}
```

## 📂 Recommended File Structure

- **One-time install scripts:** `run_once_*.sh.tmpl` (run only if they don't exist or content changes).
- **Change-driven scripts:** `run_onchange_*.sh.tmpl` (run every time the file content changes, useful for package lists).
- **Dotfiles:** Use `dot_config/folder/file` to map to `~/.config/folder/file`.

---

Keep this file updated if you change the main installation logic.

# 📈 Scalability & Maintenance

This repository is designed to be scalable and support multiple Linux distributions (currently **Arch/EndeavourOS** and **Ubuntu/Debian**). This guide explains how to add new configurations or packages.

## 📦 Package Management

Packages are not defined inside the shell scripts to keep them clean. Instead, they are centralized in `.chezmoidata.yaml` for easy reading and management.

### Structure of `.chezmoidata.yaml`

```yaml
packages:
  # List for Ubuntu/Debian
  ubuntu:
    - packagename
    - another-package
  # List for Arch/EndeavourOS
  arch:
    - packagename
    - another-package-aur
```

### How to add a package

1. Open `.chezmoidata.yaml`.
2. Find the section for your distro (`arch` or `ubuntu`).
3. Add the exact package name.
   - On **Arch**, the script uses `yay` (supports official repos and AUR).
   - On **Ubuntu**, the script uses `apt`.

### How to add a new Distribution (e.g., Fedora)

1. Add a new list in `.chezmoidata.yaml`:
   ```yaml
   packages:
     fedora:
       - git
       - zsh
   ```
2. Edit `run_onchange_00_install-packages.sh.tmpl`:
   - Add a conditional block:
     ```bash
     {{ else if eq .chezmoi.osRelease.id "fedora" -}}
     packages=(
     {{ range .packages.fedora -}}
       {{ . }}
     {{ end -}}
     )
     sudo dnf install -y "${packages[@]}"
     ```

## 🖥️ Script Configuration

Scripts use `chezmoi` templates (`.tmpl`). You can use Go template logic to condition execution.

### Useful Variables

- `{{ .chezmoi.os }}`: `linux`, `darwin` (macOS), `windows`.
- `{{ .chezmoi.osRelease.id }}`: `arch`, `ubuntu`, `debian`, `fedora`.
- `{{ .entorno }}`: Custom variable defined during init (`personal` or `adaion`).

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

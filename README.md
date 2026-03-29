# 🚀 adacosdev Dotfiles

![Chezmoi](https://img.shields.io/badge/Managed%20by-Chezmoi-black?logo=chezmoi)
![Linux](https://img.shields.io/badge/Bootstrap-Linux-blue?logo=linux)
![Cross Platform](https://img.shields.io/badge/Strategy-Common%20%2B%20Platform-informational)

My personal configuration platform powered by **Chezmoi**.

The repository now follows a clearer model:

- **Common dotfiles** are the portable core.
- **Platform layers** add OS-specific behavior.
- **Bootstrap scripts** provision machines, but are not the source of truth.
- The **chezmoi source repo** is canonical.

At the moment, the bootstrap workflow is still **Linux-first** (Ubuntu/Debian and Arch/EndeavourOS), while the repository structure is being prepared for cleaner `linux`, `darwin`, and `windows` layering.

## ✨ Key Features

- 🎨 **Adaptive VS Code:** Status bar changes color based on environment (Work/Personal).
- 🖋️ **Fonts:** Automated installation of Nerd Fonts (*JetBrainsMono, Iosevka, FiraCode, Hack*).
- 🛠️ **Runtimes:** Ready-to-use setup for Docker, Node.js (`fnm`), and Python (`pyenv`).
- 🐚 **Zsh & Warp:** Optimized aliases, dynamic prompts (`starship`), and plugin management.
- 🧭 **Source of truth defined:** reverse-sync is import-only; review changes before commit/push.

## 📥 Quick Install

On a fresh installation, you only need `curl` (and `git` usually). Just run this "Zero to Hero" command:

```bash
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply adacosdev
```

If you already have `chezmoi`:

```bash
chezmoi init --apply adacosdev
```

## 📂 Scripts Structure
Scripts execute in a deterministic order thanks to numbered prefixes:

| Script                        | Function |
|----------------------------------|---------|
| `run_onchange_00_...` | Installs essential system packages (apt/pacman). |
| `run_once_10_...` | Installs and configures Docker & permissions. |
| `run_once_20_...` | Sets up runtimes like `fnm` (Node) and `pyenv` (Python). |
| `run_once_30_...` | Downloads and caches Nerd Fonts in ~/.local/share/fonts. |
| `run_once_40_...` | Syncs VS Code extensions via CLI. |
| `run_once_99_...` | Finalizes setup and ensures Zsh is the default shell. |

## 📂 Documentation & Scaling

This repository includes guides to facilitate maintenance:

- [🏗️ Architecture Guide](docs/ARCHITECTURE.md): Explains the repository domains (`common`, `platform`, `bootstrap`, `sync`) and the source-of-truth model.
- [🖥️ Platform Guide](docs/PLATFORMS.md): Clarifies what belongs in shared config vs Linux/macOS/Windows-specific layers.
- [📈 Scalability Guide](docs/SCALABILITY.md): Explains how to add new packages, support new distros, and understand the `.chezmoidata.yaml` structure.
- [🛠️ Tools Guide](docs/TOOLS.md): Discover how to use the included productivity tools like `h` (aliases cheatsheet), `hs` (smart history), `zoxide`, `lazygit`, and more.
- [🔄 Sync Guide](SYNC_GUIDE.md): Documents reverse-sync, source of truth, backups, and the recommended review workflow.

## 🔧 Environment Configuration
This repo uses dynamic Chezmoi templates. The first time you run `chezmoi init`, you will be prompted for your email and environment type (`personal` or `adaion`).

To change these values later:

```bash
chezmoi init
```
Or edit the configuration file directly:
```bash
chezmoi edit-config
```

Made with ❤️ by [adacosdev](https://github.com/adacosdev)

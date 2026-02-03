# 🚀 adacosdev Dotfiles

![Linux](https://img.shields.io/badge/OS-Linux-blue?logo=linux)
![Chezmoi](https://img.shields.io/badge/Managed%20by-Chezmoi-black?logo=chezmoi)
![VSCode](https://img.shields.io/badge/Editor-VS%20Code-007ACC?logo=visual-studio-code)

My personal and professional fullstack configuration automated with **Chezmoi**. Optimized for a hybrid workflow on **Ubuntu** and **EndeavourOS**.

## ✨ Key Features

- 🎨 **Adaptive VS Code:** Status bar changes color based on environment (Work/Personal).
- 🖋️ **Fonts:** Automated installation of Nerd Fonts (*JetBrainsMono, Iosevka, FiraCode, Hack*).
- 🛠️ **Runtimes:** Ready-to-use setup for Docker, Node.js (`fnm`), and Python (`pyenv`).
- 🐚 **Zsh & Warp:** Optimized aliases, dynamic prompts (`starship`), and plugin management.

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

- [📈 Scalability Guide](docs/SCALABILITY.md): Explains how to add new packages, support new distros, and understand the `.chezmoidata.yaml` structure.
- [🛠️ Tools Guide](docs/TOOLS.md): Discover how to use the included productivity tools like `h` (supercharged history), `zoxide`, `lazygit`, and more.

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

# 🏗️ Repository Architecture

This repository is organized around **responsibility boundaries**, not only file types.

## The Four Domains

### 1. Common dotfiles

Portable configuration that should remain as consistent as possible across machines:

- `git`
- `zsh`
- `tmux`
- `nvim`
- CLI tooling configuration

These files represent the core developer experience and should stay simple,
reviewable, and mostly OS-agnostic.

### 2. Platform layers

Platform-specific differences belong in explicit OS layers:

- `linux`
- `darwin`
- `windows`
- optional `wsl`

Examples:

- package managers
- service managers
- filesystem paths
- desktop integrations
- terminal or editor behaviors that differ by OS

### 3. Bootstrap

Bootstrap provisions a machine. It is **not** the source of truth for the
configuration itself.

Examples:

- install packages
- install runtimes
- install fonts
- install editor extensions
- set default shell

Bootstrap is allowed to be imperative. Dotfiles should stay declarative.

### 4. Sync / import

Reverse-sync imports changes from application-managed config files back into the
repo. It exists to capture intentional changes made through UIs.

Rules:

- sync is import-only
- repo remains canonical
- commit/push stays intentional
- backups live outside the repo

## Current State

Today the repo already has a solid portable core, but bootstrap is still
Linux-first and some platform boundaries are implicit rather than structural.

This is acceptable during transition, but new work should follow the target
model below.

## Target Model

When adding or refactoring files, place them mentally in one of these domains:

- **common** → portable dotfiles
- **platform** → OS-specific overrides
- **bootstrap** → machine provisioning
- **sync** → import workflows and utilities

If a file does more than one of these jobs, it probably needs to be split.

## Source of Truth

The chezmoi source repository is the canonical source of truth.

- Edit templates in the repo when possible.
- Import UI changes through sync only when necessary.
- Review diffs before committing.
- Treat Git history as the long-term audit trail.

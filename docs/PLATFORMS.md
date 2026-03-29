# 🖥️ Platform Strategy

The goal is not byte-for-byte identical configuration on every operating
system. The goal is a **consistent mental model** with platform-aware
implementation.

## Shared vs Platform-Specific

### Shared by default

Keep these portable whenever possible:

- shell UX
- git identity and aliases
- tmux workflow
- neovim configuration
- prompt design
- CLI tool configuration

### Platform-specific by design

These should be isolated instead of forced into a single generic config:

- package installation
- service management (`systemd`, launch agents, scheduled tasks)
- GUI application paths
- clipboard integration
- OS terminal integrations
- global shortcuts

## Linux

Current bootstrap support is strongest on Linux.

Linux-specific concerns currently include:

- distro package management
- `systemd` user services
- desktop packages and fonts
- editor extension installation via CLI

## macOS

Future `darwin` work should focus on:

- Homebrew packages and casks
- `/opt/homebrew` vs `/usr/local` differences
- shell/environment bootstrap
- terminal compatibility

## Windows / WSL

Windows should be treated as a first-class platform, not a Linux clone.

Examples of Windows-specific handling:

- PowerShell
- `%APPDATA%` and Windows-specific paths
- `winget` / `scoop` / `choco`
- Windows Terminal or WSL-specific integration

WSL can share parts of the Linux CLI layer, but should still document its own
exceptions.

## Decision Rule

Before adding a new config or script, ask:

1. Is this portable?
2. Is this OS-specific?
3. Is this bootstrap or actual config?
4. Should this live in common, platform, or bootstrap?

If the answer is unclear, do not hide it behind ad-hoc conditionals. Model the
difference explicitly.

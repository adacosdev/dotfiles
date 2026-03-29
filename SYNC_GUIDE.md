# Chezmoi Config Sync Guide

## Overview

This repository treats the **chezmoi source repo as the source of truth**.
Live application configs can be imported back into the repo, but only as a
review step. The goal is to capture intentional UI changes without letting
automation silently rewrite or push your dotfiles.

## Principles

- The chezmoi source repo is canonical.
- Reverse-sync is for importing UI changes, not replacing design decisions.
- Commit and push should stay intentional.
- Operational backups do not belong inside the repo.

## How It Works

1. Apps write to live config files under `$HOME`.
2. `chezmoi-config-sync` imports supported changes into the chezmoi source repo.
3. Templates are validated with `chezmoi dump`.
4. You review the diff.
5. You commit and push manually unless you explicitly opt into automation.

## Managed Configurations

| Application | Config Path | Chezmoi Path |
|-------------|-------------|--------------|
| VS Code | `~/.config/Code/User/` | `dot_config/private_Code/User/` |
| Cursor IDE | `~/.config/Cursor/User/` | `dot_config/private_Cursor/User/` |
| Antigravity | `~/.config/Antigravity/User/` | `dot_config/private_Antigravity/User/` |
| Warp Terminal | `~/.config/warp-terminal/` | `dot_config/warp-terminal/` |
| Git | `~/.gitconfig` | `dot_gitconfig` |
| Starship | `~/.config/starship.toml` | `dot_config/starship.toml.tmpl` |
| Zsh | `~/.config/zsh/*.zsh` | `dot_config/zsh/*.zsh.tmpl` |

## Not Synced Intentionally

- caches and compiled artifacts
- sessions and temporary state
- secrets, tokens, passwords, API keys
- machine-specific overrides that should stay local

## Source of Truth

### Rule

The chezmoi source repository is the canonical source of truth.

### What that means

- Edit templates in the repo when possible.
- Use reverse-sync only to import changes made through application UIs.
- Review imported changes before committing.
- Do not rely on timers or auto-push as your audit trail.

## Recommended Workflow

1. Change settings in the app UI if needed.
2. Run:

   ```bash
   chezmoi-config-sync
   ```

3. Review changes:

   ```bash
   cd ~/.local/share/chezmoi
   git diff
   ```

4. Stage only intentional changes:

   ```bash
   git add <files>
   ```

5. Commit manually:

   ```bash
   git commit -m "sync: import reviewed UI changes"
   ```

6. Push manually when ready:

   ```bash
   git push
   ```

## Optional Timer

The systemd timer is optional. If you use it, keep it import-only:

```ini
ExecStart=%h/.local/bin/chezmoi-config-sync
```

Avoid unattended `--push` workflows.

## Backups

Backups created by the sync script live in:

```bash
~/.local/state/chezmoi-sync/backups
```

This keeps operational snapshots out of the chezmoi repo. Git history remains
the long-term audit trail.

## Validation

After each import, validate with:

```bash
chezmoi dump --format=json
```

This ensures templates still render correctly.

## Troubleshooting

### Changes not detected

- verify the app actually wrote to disk
- compare mtimes with `~/.cache/chezmoi-sync-timestamp`
- run `chezmoi-config-sync` manually

### Validation fails

- run `chezmoi dump --format=json`
- inspect the imported file
- fix the template before committing

### Timer issues

- `systemctl --user status chezmoi-config-sync.timer`
- `journalctl --user -u chezmoi-config-sync.service -n 50`

## Summary

Use sync as an **import tool**, not as a silent configuration authority.

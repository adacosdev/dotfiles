# Chezmoi Config Sync Guide

## Overview

This system automatically syncs configuration changes made in VS Code, Cursor, Warp Terminal, and other applications back to your chezmoi dotfiles repository. This allows you to:

- Make UI changes in your favorite applications
- Automatically capture those changes in your dotfiles
- Keep your dotfiles in sync with your actual configurations
- Share configurations across machines

## How It Works

### Architecture

```
┌─────────────────────────────────────────────────────┐
│ Application UIs (VS Code, Cursor, Warp, etc.)       │
│ User makes changes via preferences/settings         │
└────────────────────┬────────────────────────────────┘
                     │ writes to
                     ▼
┌─────────────────────────────────────────────────────┐
│ Config files (~/.config/Code, ~/.config/Cursor...)  │
└────────────────────┬────────────────────────────────┘
                     │ detected by
                     ▼
┌─────────────────────────────────────────────────────┐
│ chezmoi-config-sync script (runs periodically)      │
│ - Detects changed files                             │
│ - Validates templates                               │
│ - Auto-commits to git                               │
└────────────────────┬────────────────────────────────┘
                     │ writes to
                     ▼
┌─────────────────────────────────────────────────────┐
│ Chezmoi dotfiles repository                         │
│ ~/.local/share/chezmoi/dot_config/...               │
└─────────────────────────────────────────────────────┘
```

### Components

1. **chezmoi-config-sync** (main script)
   - Detects changes in managed config files
   - Copies files to chezmoi repository
   - Validates templates with `chezmoi dump`
   - Optionally commits and pushes changes

2. **systemd timer** (chezmoi-config-sync.timer)
   - Runs sync script every 5 minutes
   - Automatic and hands-off
   - Can be enabled/disabled per-user

3. **git post-commit hook**
   - Validates templates after every commit
   - Prevents invalid configs from being committed
   - Shows sync summary

## Installation & Setup

### 1. Enable the sync script

The sync script is installed when you run `chezmoi apply`:

```bash
chezmoi apply
```

This copies `run_once_chezmoi_config_sync_install.sh` to your system, which will:
- Install the `chezmoi-config-sync` command to `~/.local/bin`
- Make it executable

### 2. Enable automatic syncing (Optional)

To automatically sync configs every 5 minutes, enable the systemd timer:

```bash
systemctl --user enable --now chezmoi-config-sync.timer
```

Check status:
```bash
systemctl --user status chezmoi-config-sync.timer
systemctl --user list-timers chezmoi-config-sync.timer
```

View sync logs:
```bash
journalctl --user -u chezmoi-config-sync.service -f
```

### 3. Disable automatic syncing (Optional)

```bash
systemctl --user disable --stop chezmoi-config-sync.timer
```

## Usage

### Manual Sync

**Just detect changes (don't commit):**
```bash
chezmoi-config-sync
```

Shows what files have changed since last sync.

**Sync and commit automatically:**
```bash
chezmoi-config-sync --auto-commit
```

Automatically stages, commits, and validates changes.

**Sync, commit, and push:**
```bash
chezmoi-config-sync --auto-commit --push
```

Also pushes changes to your remote repository.

### Workflow Examples

#### Example 1: Manual sync workflow

1. Install VS Code extension, adjust settings
2. Change Warp theme in terminal
3. Modify Cursor keybindings
4. Run sync detection:
   ```bash
   chezmoi-config-sync
   ```
5. Review changes:
   ```bash
   cd ~/.local/share/chezmoi
   git diff --cached
   ```
6. Manually commit:
   ```bash
   git add dot_config/private_Code dot_config/warp-terminal
   git commit -m "sync: Update VS Code and Warp configs"
   ```

#### Example 2: Automatic sync (systemd timer)

1. Enable the timer:
   ```bash
   systemctl --user enable --now chezmoi-config-sync.timer
   ```
2. Make any UI changes (VS Code, Cursor, Warp, etc.)
3. Changes are automatically detected and committed every 5 minutes
4. Monitor sync activity:
   ```bash
   journalctl --user -u chezmoi-config-sync.service -f
   ```

#### Example 3: Full automation

1. Enable timer with push:
   ```bash
   systemctl --user enable --now chezmoi-config-sync.timer
   ```
2. Modify timer to auto-push (edit service file):
   ```bash
   systemctl --user edit chezmoi-config-sync.service
   ```
   Change:
   ```ini
   ExecStart=%h/.local/bin/chezmoi-config-sync --auto-commit
   ```
   To:
   ```ini
   ExecStart=%h/.local/bin/chezmoi-config-sync --auto-commit --push
   ```
3. Reload and restart:
   ```bash
   systemctl --user daemon-reload
   systemctl --user restart chezmoi-config-sync.timer
   ```

## Managed Configurations

### Currently Synced

| Application | Config Path | Chezmoi Path | Status |
|-------------|-------------|--------------|--------|
| VS Code | `~/.config/Code/User/` | `dot_config/private_Code/User/` | ✅ |
| Cursor IDE | `~/.config/Cursor/User/` | `dot_config/private_Cursor/User/` | ✅ |
| Antigravity | `~/.config/Antigravity/User/` | `dot_config/private_Antigravity/User/` | ✅ |
| Warp Terminal | `~/.config/warp-terminal/` | `dot_config/warp-terminal/` | ✅ |
| Git | `~/.gitconfig` | `dot_gitconfig` | ✅ |
| Starship | `~/.config/starship.toml` | `dot_config/starship.toml.tmpl` | ✅ |
| Zsh | `~/.config/zsh/*.zsh` | `dot_config/zsh/*.zsh.tmpl` | ✅ |

### NOT Synced (intentionally)

- Application caches (node_modules, compiled code, etc.)
- Session/temporary data (browser history, recent files, etc.)
- API keys, passwords, tokens (use `chezmoidata` template variables instead)
- Machine-specific settings (use conditional templates with `.chezmoi.os`, `.chezmoi.hostname`)

## How Sync Detects Changes

The script uses file modification timestamps:

1. **First run**: Records current time in `~/.cache/chezmoi-sync-timestamp`
2. **Subsequent runs**: Compares file timestamps against this marker
3. **Files newer than marker**: Considered "changed" and synced
4. **After successful sync**: Updates timestamp marker

This is lightweight and doesn't require watching file systems or hashing.

## Template Variables in Synced Files

Some files contain template variables (e.g., `{{ .chezmoi.homeDir }}`). The sync script:

1. **Preserves existing templates** when syncing
2. **Replaces hardcoded paths** with `{{ .chezmoi.homeDir }}`
3. **Validates all templates** before committing

For example, Warp's MCP path:
```json
"MCPExecutionPath": "{{ .chezmoi.homeDir }}/.opencode/bin:..."
```

## Validation

After every sync, the script validates all templates:

```bash
chezmoi dump --format=json
```

This ensures:
- No syntax errors in templates
- All template variables are valid
- All files render correctly for your environment

If validation fails, changes are NOT committed.

## Git Integration

### Commit Messages

Auto-generated sync commits follow this format:

```
sync: Update configs from UI changes (2026-03-05 14:30:42)

Updated files:
dot_config/private_Code/User/settings.json.tmpl, dot_config/warp-terminal/user_preferences.json.tmpl
```

### Post-Commit Hook

After each commit, a hook runs to:
1. Validate all templates
2. Prevent invalid configs from being pushed
3. Show sync summary

## Troubleshooting

### Changes not being detected

**Problem**: I made UI changes but sync doesn't detect them.

**Solutions**:
1. Ensure application actually writes config to disk (some apps cache in memory)
2. Check file timestamps: `ls -la ~/.config/Code/User/settings.json`
3. Run sync manually: `chezmoi-config-sync`
4. Check sync timestamp: `cat ~/.cache/chezmoi-sync-timestamp`

### Template validation fails

**Problem**: `chezmoi dump` fails after sync.

**Solutions**:
1. Check what failed: `chezmoi dump 2>&1 | head -20`
2. Verify your template variables are valid:
   - `{{ .chezmoi.homeDir }}` - Always available
   - `{{ .entorno }}` - Set in `.chezmoidata.yaml`
   - Custom variables - Check `.chezmoidata.yaml`
3. Backup and fix the file manually
4. Re-run sync

### Systemd timer not running

**Problem**: Timer is enabled but sync isn't running.

**Solutions**:
1. Check timer status: `systemctl --user status chezmoi-config-sync.timer`
2. View logs: `journalctl --user -u chezmoi-config-sync.service -n 20`
3. Manually trigger: `systemctl --user start chezmoi-config-sync.service`
4. Check user timers: `systemctl --user list-timers`

### Git push fails from timer

**Problem**: Auto-commit works but push fails.

**Solutions**:
1. Check remote setup: `cd ~/.local/share/chezmoi && git remote -v`
2. Ensure SSH key or credentials are available in user session
3. Use `--ssh-key-path` with git config if needed
4. Disable auto-push: `systemctl --user edit chezmoi-config-sync.service`

## Advanced Configuration

### Adjust sync frequency

Edit the timer:

```bash
systemctl --user edit chezmoi-config-sync.timer
```

Change `OnUnitActiveSec` (default: 5min):

```ini
OnUnitActiveSec=10min    # Sync every 10 minutes
OnUnitActiveSec=1h       # Sync every hour
OnUnitActiveSec=5s       # Sync every 5 seconds (not recommended!)
```

### Add custom configs to sync

Edit the sync script:

```bash
nano ~/.local/share/chezmoi/run_once_chezmoi_config_sync_install.sh
```

Add new sync_config call:

```bash
sync_config \
  "$HOME_DIR/.config/my-app/settings.json" \
  "$CHEZMOI_DIR/dot_config/private_my_app/settings.json" \
  "My App settings"
```

### Disable sync for specific configs

Comment out the relevant `sync_config` call in the script.

### View all synced files

```bash
cd ~/.local/share/chezmoi
git log --name-status | grep "^sync:"
```

## Best Practices

1. **Enable the systemd timer** for hands-off syncing
2. **Review commits** periodically to catch unexpected changes
3. **Don't manually edit files** in chezmoi when timer is running (use UI instead)
4. **Push to remote** to backup configs across machines
5. **Disable during major upgrades** to avoid syncing temporary migration changes
6. **Monitor journal** to catch validation failures: `journalctl --user -u chezmoi-config-sync -f`

## FAQ

**Q: Will this sync passwords/secrets?**
A: No. Sensitive data should use template variables like `{{ .secrets.api_key }}` from `.chezmoidata.yaml` (which is in `.gitignore`).

**Q: Can I have different configs per machine?**
A: Yes! Use template conditions:
```json
{{ if eq .chezmoi.hostname "work-laptop" }}
...work-specific config...
{{ else }}
...home config...
{{ end }}
```

**Q: What if I want to revert a change?**
A: Use git:
```bash
cd ~/.local/share/chezmoi
git log --oneline | grep sync
git revert <commit-hash>
chezmoi apply
```

**Q: Can I sync application data (not just settings)?**
A: Not recommended. Sync is for configuration only. Application data (caches, databases, etc.) should not be in dotfiles.

**Q: How do I exclude certain files?**
A: Add them to `.chezmoiignore` or comment out their `sync_config` line in the script.

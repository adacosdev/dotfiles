# adacosdev-dots

Interactive TUI for managing dotfiles and bootstrapping new machines.

## Features

- **Bootstrap Wizard** — Interactive TUI to install all system components
- **Diff Viewer** — Browse pending dotfile changes with syntax highlighting
- **Apply Flow** — Preview and apply dotfile changes with confirmation
- **Status Dashboard** — Health check for all installed components

## Installation

```bash
# From source
git clone https://github.com/adacosdev/dotfiles.git
cd dotfiles/bootstrap/tui
make install

# Or with go install
go install github.com/adacosdev/dotfiles/bootstrap/tui/cmd/adacosdev-dots@latest
```

### chezmoi Integration

The binary works seamlessly with chezmoi. You can:

1. **Symlink from chezmoi's bin directory:**
   ```bash
   ln -s $(which adacosdev-dots) ~/.local/bin/adacosdev-dots
   ```

2. **Install to ~/.local/bin (shown above):**
   ```bash
   make install
   ```

The binary detects chezmoi automatically and provides diff/apply commands that wrap `chezmoi diff` and `chezmoi apply`.

## Usage

```bash
adacosdev-dots [command] [flags]

Commands:
  bootstrap    Bootstrap system components (wizard or non-interactive)
  diff        Show pending dotfile changes
  apply       Apply dotfile changes
  status      Show system component status

Flags:
  --force     Skip all confirmations
  --dry-run   Show what would be done
  --json      Output machine-readable JSON
```

## Examples

```bash
# Interactive bootstrap
adacosdev-dots bootstrap

# Check status
adacosdev-dots status

# Show diff
adacosdev-dots diff

# Apply in dry-run mode
adacosdev-dots apply --dry-run

# Force apply
adacosdev-dots apply --force

# JSON output for scripting
adacosdev-dots status --json
```

## Development

```bash
make build      # Build binary
make build-dev  # Build with debug info
make test       # Run tests
make test-cover # Run tests with coverage
make lint       # Run linter
make fmt        # Format code
make tidy       # Tidy go modules
make clean      # Clean artifacts
```

## Requirements

- Go 1.23+
- chezmoi (for diff/apply commands)

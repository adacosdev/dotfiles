// Package main is the entry point for adacosdev-dots.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/cli"
	"github.com/adacosdev/dotfiles/bootstrap/tui/pkg/tty"
	"github.com/charmbracelet/bubbletea"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		// In JSON mode, print error as JSON
		if tty.ForceJSON() {
			fmt.Printf(`{"type":"error","message":%q}`, err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "\033[31mError: %s\033[0m\n", err)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	// If no args, show help
	if len(args) == 0 {
		return printHelp(nil)
	}

	// Get subcommand first
	subcommand, subArgs := cli.GetSubcommand(args)

	// Validate subcommand
	if !cli.ValidateSubcommand(string(subcommand)) {
		// Check if they passed --help or just invalid args
		if contains(args, "--help") || contains(args, "-h") {
			return printHelp(args)
		}
		return fmt.Errorf("unknown command: %s\nRun 'adacosdev-dots --help' for usage", args[0])
	}

	// Parse global flags (from args before subcommand)
	preForce, preDryRun, preJsonMode, preHelp, _ := cli.ParseFlags(args)

	// Parse subcommand-specific flags (after subcommand)
	subForce, subDryRun, subJsonMode, subHelp, _ := cli.ParseFlags(subArgs)

	// Combine flags from both positions
	force := preForce || subForce
	dryRun := preDryRun || subDryRun
	jsonMode := preJsonMode || subJsonMode
	help := preHelp || subHelp

	// If --json was passed (either before or after subcommand), set it
	if jsonMode {
		tty.SetForceJSON(true)
	}

	// If --help was passed for subcommand, show subcommand help
	if help {
		return printSubcommandHelp(subcommand)
	}

	// Create router
	router := cli.NewRouter(force, dryRun, jsonMode)

	// Dispatch based on subcommand and TTY mode
	switch subcommand {
	case cli.SubcommandBootstrap:
		return runBootstrap(router, subArgs)
	case cli.SubcommandDiff:
		return runDiff(router)
	case cli.SubcommandApply:
		return runApply(router, subArgs)
	case cli.SubcommandStatus:
		return runStatus(router)
	case cli.SubcommandSelect:
		return runSelect(router)
	default:
		return printHelp(args)
	}
}

// runBootstrap runs the bootstrap command in TUI or CLI mode.
func runBootstrap(router *cli.Router, args []string) error {
	if tty.IsTerminal() {
		return runWithErrorHandler(func() error {
			model := router.Bootstrap()
			program := tea.NewProgram(model,
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)
			_, err := program.Run()
			return err
		})
	}

	// CLI mode
	return runWithErrorHandler(func() error {
		return router.BootstrapCLI()
	})
}

// runDiff runs the diff command in TUI or CLI mode.
func runDiff(router *cli.Router) error {
	if tty.IsTerminal() {
		return runWithErrorHandler(func() error {
			model := router.Diff()
			program := tea.NewProgram(model,
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)
			_, err := program.Run()
			return err
		})
	}

	// CLI mode - just run chezmoi diff
	return runWithErrorHandler(func() error {
		return router.DiffCLI()
	})
}

// runApply runs the apply command in TUI or CLI mode.
func runApply(router *cli.Router, args []string) error {
	if tty.IsTerminal() {
		return runWithErrorHandler(func() error {
			model := router.Apply()
			program := tea.NewProgram(model,
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)
			_, err := program.Run()
			return err
		})
	}

	// CLI mode
	return runWithErrorHandler(func() error {
		return router.ApplyCLI()
	})
}

// runStatus runs the status command in TUI or CLI mode.
func runStatus(router *cli.Router) error {
	if tty.IsTerminal() {
		return runWithErrorHandler(func() error {
			model := router.Status()
			program := tea.NewProgram(model,
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)
			_, err := program.Run()
			return err
		})
	}

	// CLI mode
	return runWithErrorHandler(func() error {
		return router.StatusCLI()
	})
}

// runSelect runs the select command in TUI or CLI mode.
func runSelect(router *cli.Router) error {
	if tty.IsTerminal() {
		return runWithErrorHandler(func() error {
			model := router.Select()
			program := tea.NewProgram(model,
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)
			_, err := program.Run()
			return err
		})
	}

	// CLI mode
	return runWithErrorHandler(func() error {
		return router.SelectCLI()
	})
}

// runWithErrorHandler wraps a function with error handling.
// In CLI mode, errors are printed directly.
// In TTY mode, errors are shown in a Bubbletea error view and wait for keypress.
func runWithErrorHandler(fn func() error) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- fn()
	}()

	err := <-errCh
	if err != nil {
		// Check if we should use TTY error handling
		if tty.IsTerminal() && !tty.ForceJSON() {
			// In TTY mode, show error in a message and wait for keypress
			return fmt.Errorf("%s", err)
		}
		return err
	}
	return nil
}

// printHelp prints the main help text.
func printHelp(args []string) error {
	if len(args) > 0 && (args[0] == "--json" || args[0] == "-j") {
		// JSON help mode
		fmt.Println(`{"type":"help","usage":"adacosdev-dots <command> [flags]"}`)
		return nil
	}

	fmt.Print(`adacosdev-dots - Dotfiles management TUI

Usage: adacosdev-dots <command> [flags]

Commands:
  bootstrap    Bootstrap system components (wizard or non-interactive)
  diff         Show pending dotfile changes
  apply        Apply dotfile changes
  status       Show system component status
  select       Select which configs to apply (interactive)

Flags:
  --force, -f    Skip all confirmations
  --dry-run, -n  Show what would be done without doing it
  --json         Output machine-readable JSON
  -h, --help     Show this help

Examples:
  adacosdev-dots status
  adacosdev-dots bootstrap
  adacosdev-dots diff --json
  adacosdev-dots apply --force
  adacosdev-dots select

Run without arguments to start the interactive TUI.
`)
	return nil
}

// printSubcommandHelp prints help for a specific subcommand.
func printSubcommandHelp(subcommand cli.Subcommand) error {
	switch subcommand {
	case cli.SubcommandBootstrap:
		fmt.Print(`adacosdev-dots bootstrap - Bootstrap system components

Usage: adacosdev-dots bootstrap [flags]

Flags:
  --force, -f    Skip all confirmations
  --dry-run, -n  Show what would be done without doing it
  --json         Output machine-readable JSON
  -h, --help     Show this help

Description:
  Detects your operating system and shows available components to install.
  In TTY mode, runs an interactive wizard.
  In CLI mode, shows component status and exits.
`)
	case cli.SubcommandDiff:
		fmt.Print(`adacosdev-dots diff - Show pending dotfile changes

Usage: adacosdev-dots diff [flags]

Flags:
  --json         Output machine-readable JSON
  -h, --help     Show this help

Description:
  Shows all pending dotfile changes using chezmoi diff.
  In TTY mode, runs an interactive diff viewer.
  In CLI mode, shows plain text diff output.
`)
	case cli.SubcommandApply:
		fmt.Print(`adacosdev-dots apply - Apply dotfile changes

Usage: adacosdev-dots apply [flags]

Flags:
  --force, -f    Skip confirmation prompt
  --dry-run, -n  Show what would be applied without applying
  --json         Output machine-readable JSON
  -h, --help     Show this help

Description:
  Applies pending dotfile changes using chezmoi apply.
  In TTY mode, runs an interactive apply flow.
  In CLI mode, applies changes directly (use --force to skip prompt).
`)
	case cli.SubcommandStatus:
		fmt.Print(`adacosdev-dots status - Show system component status

Usage: adacosdev-dots status [flags]

Flags:
  --json         Output machine-readable JSON
  -h, --help     Show this help

Description:
  Checks and displays the status of all system components.
  In TTY mode, runs an interactive status dashboard.
  In CLI mode, shows a plain text status table.
`)
	case cli.SubcommandSelect:
		fmt.Print(`adacosdev-dots select - Select which configs to apply

Usage: adacosdev-dots select [flags]

Flags:
  --force, -f    Skip confirmation prompt
  --dry-run, -n  Show what would be applied without applying
  --json         Output machine-readable JSON
  -h, --help     Show this help

Description:
  Interactively select which dotfile configs to apply.
  In TTY mode, shows a wizard with checkboxes for each config.
  In CLI mode, lists available configs.

Categories:
  Shells: zsh, fish, nushell
  Terminals: kitty, alacritty, wezterm, ghostty
  Multiplexers: tmux, zellij
  Editors: nvim
  Tools: starship, git
  Claude: settings, statusline
`)
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
	return nil
}

// contains checks if a string slice contains a value.
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if strings.ToLower(s) == strings.ToLower(val) {
			return true
		}
	}
	return false
}

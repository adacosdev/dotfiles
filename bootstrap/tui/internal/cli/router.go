// Package cli provides CLI routing and global flag handling.
package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/apply"
	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/bootstrap"
	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/bootstrap/detector"
	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/diff"
	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/select"
	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/status"
	"github.com/adacosdev/dotfiles/bootstrap/tui/pkg/tty"
	"github.com/charmbracelet/bubbletea"
)

// Router handles CLI command routing with global flags.
type Router struct {
	force    bool
	dryRun   bool
	jsonMode bool
	isTTY    bool
}

// NewRouter creates a new CLI router with the given flags.
func NewRouter(force, dryRun, jsonMode bool) *Router {
	return &Router{
		force:    force,
		dryRun:   dryRun,
		jsonMode: jsonMode,
		isTTY:    tty.IsTerminal(),
	}
}

// BootstrapCLI runs bootstrap in CLI (non-interactive) mode.
func (r *Router) BootstrapCLI() error {
	osInfo, err := detector.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect OS: %w", err)
	}

	components := bootstrap.AllComponents(osInfo)

	if r.jsonMode {
		return r.bootstrapJSON(osInfo, components)
	}

	// Human-readable mode
	fmt.Printf("Detected OS: %s (%s)\n\n", osInfo.Name, osInfo.ID)

	byCategory := bootstrap.ComponentsByCategory(osInfo)
	for _, category := range []string{"system", "runtime", "font", "tool", "extension"} {
		comps, ok := byCategory[category]
		if !ok {
			continue
		}

		fmt.Printf("%s:\n", bootstrap.CategoryDisplayName(category))
		for _, c := range comps {
			installed, version := c.IsInstalled(osInfo)
			statusStr := "not installed"
			if installed {
				statusStr = fmt.Sprintf("installed (%s)", version)
			}
			fmt.Printf("  [%s] %s - %s\n", statusStr, c.Name, c.Description)
		}
		fmt.Println()
	}

	return nil
}

// bootstrapJSON outputs bootstrap status as JSON lines.
func (r *Router) bootstrapJSON(osInfo *detector.OSInfo, components []bootstrap.Component) error {
	// Output OS info
	osData := map[string]string{
		"type":    "bootstrap",
		"os_id":   osInfo.ID,
		"os_name": osInfo.Name,
		"family":  osInfo.Family,
		"arch":    osInfo.Arch,
	}
	data, _ := json.Marshal(osData)
	fmt.Println(string(data))

	// Output each component status
	for _, c := range components {
		installed, version := c.IsInstalled(osInfo)
		statusStr := "ok"
		if !installed {
			statusStr = "error"
		}

		compData := map[string]interface{}{
			"type":    "component",
			"name":    c.Name,
			"status":  statusStr,
			"version": version,
		}
		if !installed {
			compData["message"] = "not installed"
		}

		data, _ := json.Marshal(compData)
		fmt.Println(string(data))
	}

	return nil
}

// DiffCLI runs diff in CLI mode (plain text chezmoi diff).
func (r *Router) DiffCLI() error {
	if r.jsonMode {
		return r.diffJSON()
	}

	cmd := exec.Command("chezmoi", "diff", "--color=never")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// diffJSON outputs diff as JSON lines.
func (r *Router) diffJSON() error {
	cmd := exec.Command("chezmoi", "diff", "--color=never")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("chezmoi diff failed: %w", err)
	}

	// Parse diff output into JSON
	files := diff.ParseDiff(stdout.String())
	for _, f := range files {
		additions := 0
		deletions := 0
		for _, hunk := range f.Hunks {
			for _, line := range hunk.Lines {
				switch line.Type {
				case "add":
					additions++
				case "del":
					deletions++
				}
			}
		}

		fileData := map[string]interface{}{
			"type":      "diff",
			"file":      f.Path,
			"status":    f.Status,
			"additions": additions,
			"deletions": deletions,
		}
		data, _ := json.Marshal(fileData)
		fmt.Println(string(data))
	}

	return nil
}

// ApplyCLI runs apply in CLI mode.
func (r *Router) ApplyCLI() error {
	if r.jsonMode {
		return r.applyJSON()
	}

	if r.dryRun {
		cmd := exec.Command("chezmoi", "diff", "--color=never")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// In CLI mode, if not force, we use --force by default since there's no TTY for prompting
	script := "chezmoi apply"
	if !r.force {
		script = "chezmoi apply --force"
	}

	cmd := exec.Command("bash", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// applyJSON outputs apply results as JSON lines.
func (r *Router) applyJSON() error {
	// First, get the diff to know what would be applied
	cmd := exec.Command("chezmoi", "diff", "--color=never")
	var diffOut bytes.Buffer
	cmd.Stdout = &diffOut
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// No changes or error
		errData := map[string]string{
			"type":    "error",
			"message": "failed to get diff",
		}
		if diffOut.Len() > 0 {
			errData["details"] = diffOut.String()
		}
		data, _ := json.Marshal(errData)
		fmt.Println(string(data))
		return nil
	}

	files := diff.ParseDiff(diffOut.String())

	if r.dryRun {
		// Just output what would be applied
		for _, f := range files {
			fileData := map[string]interface{}{
				"type":   "dry-run",
				"file":   f.Path,
				"status": f.Status,
			}
			data, _ := json.Marshal(fileData)
			fmt.Println(string(data))
		}
		return nil
	}

	// Run apply
	applyCmd := exec.Command("chezmoi", "apply")
	var applyOut bytes.Buffer
	applyCmd.Stdout = &applyOut
	applyCmd.Stderr = &applyOut

	applyErr := applyCmd.Run()

	// Output applied files
	seen := make(map[string]bool)
	for _, f := range files {
		if seen[f.Path] {
			continue
		}
		seen[f.Path] = true

		fileData := map[string]interface{}{
			"type":   "applied",
			"file":   f.Path,
			"status": "applied",
		}
		data, _ := json.Marshal(fileData)
		fmt.Println(string(data))
	}

	if applyErr != nil {
		errData := map[string]string{
			"type":    "error",
			"message": applyErr.Error(),
		}
		data, _ := json.Marshal(errData)
		fmt.Println(string(data))
	}

	return nil
}

// StatusCLI runs status in CLI mode (text table).
func (r *Router) StatusCLI() error {
	model := status.NewDashboardModel()

	// Set reasonable width for CLI
	model.Width = 80
	model.Height = 24

	if r.jsonMode {
		return r.statusJSON(model)
	}

	// Text table output
	return r.statusText(model)
}

// statusJSON outputs status as JSON lines.
func (r *Router) statusJSON(model *status.DashboardModel) error {
	checks := model.GetChecks()

	for _, check := range checks {
		checkData := map[string]interface{}{
			"type":     "component",
			"name":     check.Name,
			"category": check.Category,
			"status":   check.Status.String(),
			"version":  check.Version,
		}
		if check.Message != "" && check.Message != check.Name {
			checkData["message"] = check.Message
		}

		data, _ := json.Marshal(checkData)
		fmt.Println(string(data))
	}

	return nil
}

// statusText outputs status as a text table.
func (r *Router) statusText(model *status.DashboardModel) error {
	checks := model.GetChecks()

	// Simple text table format
	fmt.Println("=== System Status ===")

	// Group by category
	categoryMap := make(map[string][]status.HealthCheck)
	for _, check := range checks {
		categoryMap[check.Category] = append(categoryMap[check.Category], check)
	}

	for _, category := range []string{"system", "runtime", "font", "tool", "extension"} {
		categoryChecks, ok := categoryMap[category]
		if !ok || len(categoryChecks) == 0 {
			continue
		}

		fmt.Printf("--- %s ---\n", bootstrap.CategoryDisplayName(category))
		for _, check := range categoryChecks {
			statusIcon := "❓"
			switch check.Status {
			case status.StatusOk:
				statusIcon = "✅"
			case status.StatusWarn:
				statusIcon = "⚠️"
			case status.StatusError:
				statusIcon = "❌"
			}

			version := check.Version
			if version == "" || version == "not installed" {
				version = "-"
			}

			fmt.Printf("  %s %-20s %-10s %s\n", statusIcon, check.Name, version, check.Message)
		}
		fmt.Println()
	}

	return nil
}

// Bootstrap returns a Bubbletea model for TUI bootstrap wizard.
func (r *Router) Bootstrap() tea.Model {
	osInfo, err := detector.Detect()
	if err != nil {
		osInfo = &detector.OSInfo{
			ID:   "unknown",
			Name: "Unknown OS",
		}
	}

	helpersPath := findHelpersPath()
	model := bootstrap.NewWizardModel(osInfo, helpersPath)
	model.Executor().DryRun = r.dryRun
	model.Executor().Force = r.force

	return model
}

// Diff returns a Bubbletea model for TUI diff viewer.
func (r *Router) Diff() tea.Model {
	return diff.NewDiffModel()
}

// Apply returns a Bubbletea model for TUI apply flow.
func (r *Router) Apply() tea.Model {
	return apply.NewApplyModel(r.dryRun, r.force)
}

// Status returns a Bubbletea model for TUI status dashboard.
func (r *Router) Status() tea.Model {
	return status.NewDashboardModel()
}

// Select returns a Bubbletea model for TUI dotfiles selector.
func (r *Router) Select() tea.Model {
	model := selectpkg.NewSelectorModel()
	model.DryRun = r.dryRun
	model.Force = r.force
	model.JsonMode = r.jsonMode
	return model
}

// SelectCLI runs select in CLI mode (list available configs).
func (r *Router) SelectCLI() error {
	model := selectpkg.NewSelectorModel()
	return model.RunCLI(r.jsonMode)
}

// findHelpersPath finds the path to the bootstrap helpers directory.
func findHelpersPath() string {
	// Check environment variable first
	if path := os.Getenv("ADACOSDEV_HELPERS"); path != "" {
		return path
	}

	// Common locations
	locations := []string{
		filepath.Join(os.Getenv("HOME"), ".local/bin/bootstrap-helpers"),
		filepath.Join(os.Getenv("HOME"), ".bootstrap-helpers"),
		"/usr/local/bin/bootstrap-helpers",
		"/usr/bin/bootstrap-helpers",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return ""
}

// ParseFlags parses global flags from args and returns remaining args.
// Recognized flags: --force, -f, --dry-run, -n, --json, --help, -h
func ParseFlags(args []string) (force, dryRun, jsonMode, help bool, remaining []string) {
	for i, arg := range args {
		switch strings.ToLower(arg) {
		case "--force", "-f":
			force = true
		case "--dry-run", "-n":
			dryRun = true
		case "--json":
			jsonMode = true
		case "--help", "-h":
			help = true
		default:
			if !strings.HasPrefix(arg, "-") {
				remaining = args[i:]
				return
			}
		}
	}
	remaining = []string{}
	return
}

// Subcommand represents a CLI subcommand.
type Subcommand string

const (
	SubcommandBootstrap Subcommand = "bootstrap"
	SubcommandDiff      Subcommand = "diff"
	SubcommandApply     Subcommand = "apply"
	SubcommandStatus    Subcommand = "status"
	SubcommandSelect    Subcommand = "select"
)

// ValidateSubcommand checks if a subcommand name is valid.
func ValidateSubcommand(name string) bool {
	switch name {
	case "bootstrap", "diff", "apply", "status", "select":
		return true
	default:
		return false
	}
}

// GetSubcommand parses and returns the subcommand from args.
// It skips any leading flags before finding the subcommand.
func GetSubcommand(args []string) (Subcommand, []string) {
	if len(args) == 0 {
		return "", nil
	}

	// Skip leading flags to find the subcommand
	subIdx := 0
	for subIdx < len(args) && strings.HasPrefix(args[subIdx], "-") {
		subIdx++
	}

	if subIdx >= len(args) {
		return "", nil
	}

	sub := Subcommand(args[subIdx])
	switch sub {
	case SubcommandBootstrap, SubcommandDiff, SubcommandApply, SubcommandStatus, SubcommandSelect:
		// Return subcommand and args before it (flags) + args after it
		flags := args[:subIdx]
		remaining := args[subIdx+1:]
		return sub, append(flags, remaining...)
	default:
		return "", args
	}
}

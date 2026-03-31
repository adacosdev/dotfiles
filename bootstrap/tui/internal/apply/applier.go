// Package apply provides the apply flow for chezmoi dotfiles management.
package apply

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/adacosdev/dotfiles/bootstrap/tui/pkg/tty"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ApplyState represents the current state of the apply flow.
type ApplyState int

const (
	stateDiff ApplyState = iota // showing diff summary
	stateConfirm                // waiting for confirmation
	stateApplying               // running chezmoi apply
	stateDone                   // showing results
)

// ApplyResult represents the result of applying a single file.
type ApplyResult struct {
	File   string
	Status string // "applied", "skipped", "error"
	Error  string
}

// ApplyModel is the Bubbletea model for the apply flow.
type ApplyModel struct {
	state      ApplyState
	dryRun     bool
	force      bool
	diffOutput string    // cached diff output
	results    []ApplyResult
	selected   int
	width      int
	height     int
	logLines   []string
	fileCount  int
	spinnerIdx int
	err        error
}

// NewApplyModel creates a new apply model.
func NewApplyModel(dryRun, force bool) *ApplyModel {
	return &ApplyModel{
		state:      stateDiff,
		dryRun:     dryRun,
		force:      force,
		results:    []ApplyResult{},
		logLines:   []string{},
		fileCount:  0,
		spinnerIdx: 0,
	}
}

// Init initializes the apply model.
func (m *ApplyModel) Init() tea.Cmd {
	return m.fetchDiff()
}

// fetchDiff runs chezmoi diff and caches the output.
func (m *ApplyModel) fetchDiff() tea.Cmd {
	return func() tea.Msg {
		var script string
		if m.dryRun {
			script = "chezmoi diff --color=never"
		} else {
			script = "chezmoi diff --color=never 2>&1"
		}

		cmd := exec.Command("bash", "-c", script)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			m.err = fmt.Errorf("failed to run chezmoi diff: %w", err)
			m.state = stateDone
			return nil
		}

		m.diffOutput = stdout.String()
		if m.diffOutput == "" {
			m.diffOutput = "No changes to apply."
			m.fileCount = 0
		} else {
			// Count files in diff output
			m.fileCount = countFilesInDiff(m.diffOutput)
		}

		// If force flag is set, skip confirmation and start applying
		if m.force {
			m.state = stateApplying
			return m.runApply()
		}

		m.state = stateConfirm
		return nil
	}
}

// runApply executes chezmoi apply and streams output.
func (m *ApplyModel) runApply() tea.Cmd {
	return func() tea.Msg {
		m.state = stateApplying
		m.logLines = []string{"Starting chezmoi apply..."}

		script := "chezmoi apply 2>&1"
		cmd := exec.Command("bash", "-c", script)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		output := stdout.String() + stderr.String()
		lines := strings.Split(output, "\n")

		for _, line := range lines {
			if line = strings.TrimSpace(line); line != "" {
				m.logLines = append(m.logLines, line)
			}
		}

		if err != nil {
			m.logLines = append(m.logLines, fmt.Sprintf("Error: %v", err))
			m.err = err
		}

		// Parse results from output
		m.parseApplyOutput(output)

		m.state = stateDone
		return nil
	}
}

// parseApplyOutput parses the output of chezmoi apply.
func (m *ApplyModel) parseApplyOutput(output string) {
	// chezmoi apply outputs file paths that were changed
	// Pattern: any line that looks like a file path or contains "applied" or status
	filePattern := regexp.MustCompile(`(?:create|update|delete|applied?)?\s*[:\.]?\s*(~?/\S+)`)
	appliedPattern := regexp.MustCompile(`(?i)(applied|created|updated|deleted|skipped)`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := filePattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			m.results = append(m.results, ApplyResult{
				File:   matches[1],
				Status: "applied",
				Error:  "",
			})
		} else if appliedPattern.MatchString(line) {
			// Try to extract file from line
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "/") || strings.HasPrefix(part, "~") {
					m.results = append(m.results, ApplyResult{
						File:   part,
						Status: parseStatusFromLine(line),
						Error:  "",
					})
					break
				}
			}
		}
	}

	// If no results parsed, check exit code
	if len(m.results) == 0 && m.err == nil {
		m.results = append(m.results, ApplyResult{
			File:   "all files",
			Status: "applied",
			Error:  "",
		})
	}
}

// parseStatusFromLine determines status from an output line.
func parseStatusFromLine(line string) string {
	lower := strings.ToLower(line)
	if strings.Contains(lower, "skipped") {
		return "skipped"
	}
	if strings.Contains(lower, "error") || strings.Contains(lower, "failed") {
		return "error"
	}
	if strings.Contains(lower, "applied") || strings.Contains(lower, "created") ||
		strings.Contains(lower, "updated") || strings.Contains(lower, "deleted") {
		return "applied"
	}
	return "applied"
}

// countFilesInDiff counts the number of files in a diff output.
func countFilesInDiff(diff string) int {
	// Look for lines starting with diff header or file paths
	pattern := regexp.MustCompile(`(?:^diff|^index|^---|\+\+\+|create mode)`)
	lines := strings.Split(diff, "\n")
	count := 0
	inFile := false
	for _, line := range lines {
		if pattern.MatchString(line) {
			count++
			inFile = true
		} else if inFile && strings.HasPrefix(line, "diff ") {
			// New diff started
			count++
		} else if !strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "-") &&
			!strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\\") {
			inFile = false
		}
	}
	// Fallback: count lines starting with +++ or ---
	pattern2 := regexp.MustCompile(`^[+-]{3}\s`)
	for _, line := range lines {
		if pattern2.MatchString(line) {
			count++
		}
	}
	if count == 0 && diff != "" {
		count = 1
	}
	return count
}

// Update handles messages and updates the model.
func (m *ApplyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	default:
		return m, nil
	}
}

// handleKey handles keyboard input based on current state.
func (m *ApplyModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateDiff:
		// Waiting for diff to load
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

	case stateConfirm:
		switch msg.String() {
		case "y", "Y":
			m.state = stateApplying
			return m, m.runApply()
		case "n", "N", "q", "Q":
			return m, tea.Quit
		case "d", "D":
			// Toggle dry-run
			m.dryRun = !m.dryRun
			return m, m.fetchDiff()
		}

	case stateApplying:
		if msg.Type == tea.KeyCtrlC {
			m.logLines = append(m.logLines, "Cancelled by user")
			m.state = stateDone
		}

	case stateDone:
		if msg.Type == tea.KeyEnter || msg.Type == tea.KeyEsc || msg.Type == tea.KeyCtrlC || msg.String() == "q" || msg.String() == "Q" {
			return m, tea.Quit
		}
		if (msg.String() == "r" || msg.String() == "R") && m.hasErrors() {
			m.results = nil
			m.err = nil
			m.state = stateApplying
			return m, m.runApply()
		}
	}

	return m, nil
}

// handleRetry handles retry logic.
func (m *ApplyModel) handleRetry() (tea.Model, tea.Cmd) {
	if m.hasErrors() {
		m.results = nil
		m.err = nil
		m.state = stateApplying
		return m, m.runApply()
	}
	return m, nil
}

// hasErrors returns true if any results have errors.
func (m *ApplyModel) hasErrors() bool {
	for _, r := range m.results {
		if r.Status == "error" {
			return true
		}
	}
	return m.err != nil
}

// View returns the string view of the model.
func (m *ApplyModel) View() string {
	switch m.state {
	case stateDiff:
		return m.viewDiff()
	case stateConfirm:
		return m.viewConfirm()
	case stateApplying:
		return m.viewApplying()
	case stateDone:
		return m.viewDone()
	default:
		return "Unknown state"
	}
}

// viewDiff renders the diff loading screen.
func (m *ApplyModel) viewDiff() string {
	var sb strings.Builder

	sb.WriteString("\n  Fetching changes...\n\n")
	sb.WriteString(spinner(m.spinnerIdx) + "  Running chezmoi diff\n")

	return sb.String()
}

// viewConfirm renders the confirmation screen.
func (m *ApplyModel) viewConfirm() string {
	var sb strings.Builder

	// Header
	sb.WriteString("\n  Apply Changes\n")
	sb.WriteString("  ─────────────────────────────────────────────────────────\n\n")

	// Summary
	action := "will be changed"
	if m.dryRun {
		action = "would be changed (dry-run)"
	}
	sb.WriteString(fmt.Sprintf("  %d file(s) %s\n\n", m.fileCount, action))

	// Preview first few files
	preview := m.getDiffPreview()
	if preview != "" {
		sb.WriteString("  Preview:\n")
		sb.WriteString("  " + strings.Repeat("─", 50) + "\n")
		sb.WriteString(preview)
		sb.WriteString("\n")
	}

	// Instructions
	sb.WriteString("  ─────────────────────────────────────────────────────────\n")
	sb.WriteString("  [y] Apply changes   [n] Cancel   [d] Toggle dry-run   [q] Quit\n")

	return sb.String()
}

// getDiffPreview returns a preview of the diff (first few files).
func (m *ApplyModel) getDiffPreview() string {
	lines := strings.Split(m.diffOutput, "\n")
	var sb strings.Builder
	count := 0
	maxLines := 20

	for _, line := range lines {
		if count >= maxLines {
			sb.WriteString(fmt.Sprintf("\n  ... and %d more lines (run 'chezmoi diff' for full output)", len(lines)-maxLines))
			break
		}
		// Highlight file headers in diff
		if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "---") ||
			strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "index ") {
			sb.WriteString(dimStyle.Render("  " + line + "\n"))
		} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			sb.WriteString(greenStyle.Render("  " + line + "\n"))
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			sb.WriteString(redStyle.Render("  " + line + "\n"))
		} else {
			sb.WriteString("  " + line + "\n")
		}
		count++
	}

	return sb.String()
}

// viewApplying renders the applying progress screen.
func (m *ApplyModel) viewApplying() string {
	var sb strings.Builder

	sb.WriteString("\n  Applying changes...\n")
	sb.WriteString("  ─────────────────────────────────────────────────────────\n\n")

	// Animated spinner
	sb.WriteString(fmt.Sprintf("  %s Processing files...\n\n", spinner(m.spinnerIdx)))

	// Log lines (last 15)
	logLines := m.logLines
	if len(logLines) > 15 {
		logLines = logLines[len(logLines)-15:]
	}
	for _, line := range logLines {
		sb.WriteString(fmt.Sprintf("  %s\n", line))
	}

	sb.WriteString("\n\n  Press Ctrl+C to cancel\n")

	return sb.String()
}

// viewDone renders the completion screen.
func (m *ApplyModel) viewDone() string {
	var sb strings.Builder

	sb.WriteString("\n  Apply Complete\n")
	sb.WriteString("  ─────────────────────────────────────────────────────────\n\n")

	// Results summary
	applied := 0
	skipped := 0
	errorCount := 0

	for _, r := range m.results {
		var statusIcon, statusText string
		switch r.Status {
		case "applied":
			statusIcon = checkStyle.Render("✓")
			statusText = appliedStyle.Render("applied")
			applied++
		case "skipped":
			statusIcon = skipStyle.Render("⏭")
			statusText = skipStyle.Render("skipped")
			skipped++
		case "error":
			statusIcon = errorStyle.Render("✗")
			statusText = errorStyle.Render("error")
			errorCount++
		default:
			statusIcon = pendingStyle.Render("?")
			statusText = pendingStyle.Render(r.Status)
		}

		fileName := r.File
		if len(fileName) > 50 {
			fileName = "..." + fileName[len(fileName)-47:]
		}

		if r.Error != "" {
			sb.WriteString(fmt.Sprintf("  %s %-50s %s: %s\n", statusIcon, fileName, statusText, r.Error))
		} else {
			sb.WriteString(fmt.Sprintf("  %s %-50s %s\n", statusIcon, fileName, statusText))
		}
	}

	// If no individual results, show summary
	if len(m.results) == 0 && m.err == nil {
		sb.WriteString("  All files applied successfully.\n")
	} else if m.err != nil {
		sb.WriteString(fmt.Sprintf("\n  Error: %v\n", m.err))
	}

	sb.WriteString("\n")

	// Summary counts
	if applied > 0 {
		sb.WriteString(fmt.Sprintf("  %s %d applied", appliedStyle.Render("✓"), applied))
	}
	if skipped > 0 {
		sb.WriteString(fmt.Sprintf("   %s %d skipped", skipStyle.Render("⏭"), skipped))
	}
	if errorCount > 0 {
		sb.WriteString(fmt.Sprintf("   %s %d failed", errorStyle.Render("✗"), errorCount))
	}
	sb.WriteString("\n")

	// Footer
	sb.WriteString("  ─────────────────────────────────────────────────────────\n")
	if m.hasErrors() {
		sb.WriteString("  [r] Retry   [Enter/q] Quit\n")
	} else {
		sb.WriteString("  [Enter/q] Quit\n")
	}

	return sb.String()
}

// spinner returns the current spinner character.
func spinner(idx int) string {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return spinners[idx%len(spinners)]
}

// Run starts the apply flow and returns any error.
func Run(dryRun, force bool) error {
	model := NewApplyModel(dryRun, force)

	program := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("failed to run apply: %w", err)
	}

	return nil
}

// RunCLI runs apply in CLI mode (non-interactive).
func RunCLI(dryRun, force bool) error {
	if dryRun {
		// Run chezmoi diff
		script := "chezmoi diff --color=never"
		cmd := exec.Command("bash", "-c", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	if force {
		// Run chezmoi apply directly
		script := "chezmoi apply"
		cmd := exec.Command("bash", "-c", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Interactive mode
	if !tty.IsTerminal() {
		// Not a TTY, use --force by default
		script := "chezmoi apply"
		cmd := exec.Command("bash", "-c", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Run TUI
	return Run(dryRun, force)
}

// SpinnerTick returns a command that ticks the spinner.
func SpinnerTick() tea.Msg {
	time.Sleep(80 * time.Millisecond)
	return nil
}

// Styles for the apply UI.
var (
	appliedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // green
	skipStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))  // yellow
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))   // red
	pendingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))   // dark gray
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))   // dark gray
	greenStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // green
	redStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))   // red
	checkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // green
)

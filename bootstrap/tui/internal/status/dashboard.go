// Package status provides the system status dashboard TUI.
package status

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/bootstrap"
	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/bootstrap/detector"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DashboardModel is the Bubbletea model for the status dashboard.
type DashboardModel struct {
	os           *detector.OSInfo
	checks       []HealthCheck
	selected     int
	Width        int
	Height       int
	lastRefresh  time.Time
	categoryIdx  int
	showDetails  bool
	detailIndex  int
	categories   []string
	autoRefresh  *time.Ticker
	quitting     bool
}

// HealthCheck represents a single health check result.
type HealthCheck struct {
	Name      string
	Category  string
	Status    Status
	Version   string
	Message   string
	Command   string
	CheckedAt time.Time
}

// Status represents the health check status.
type Status int

const (
	StatusOk Status = iota
	StatusWarn
	StatusError
	StatusUnknown
)

// String returns a string representation of the status.
func (s Status) String() string {
	switch s {
	case StatusOk:
		return "ok"
	case StatusWarn:
		return "warn"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// Icon returns the icon for the status.
func (s Status) Icon() string {
	switch s {
	case StatusOk:
		return "✅"
	case StatusWarn:
		return "⚠️"
	case StatusError:
		return "❌"
	default:
		return "❓"
	}
}

// NewDashboardModel creates a new dashboard model.
func NewDashboardModel() *DashboardModel {
	osInfo, err := detector.Detect()
	if err != nil {
		osInfo = &detector.OSInfo{
			ID:   "unknown",
			Name: "Unknown OS",
		}
	}

	m := &DashboardModel{
		os:          osInfo,
		checks:      []HealthCheck{},
		selected:    0,
		lastRefresh: time.Time{},
		categoryIdx: 0,
		showDetails: false,
		categories:  []string{"system", "runtime", "font", "tool", "extension"},
	}

	m.refreshChecks()
	return m
}

// Init initializes the dashboard model.
func (m *DashboardModel) Init() tea.Cmd {
	m.autoRefresh = time.NewTicker(30 * time.Second)
	return func() tea.Msg {
		for {
			select {
			case <-m.autoRefresh.C:
				return refreshMsg{}
			}
		}
	}
}

// Update handles messages and updates the model.
func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case refreshMsg:
		m.refreshChecks()
		return m, nil

	default:
		return m, nil
	}
}

type refreshMsg struct{}

// handleKey handles keyboard input.
func (m *DashboardModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showDetails {
		return m.handleDetailKey(msg)
	}

	switch msg.Type {
	case tea.KeyUp, tea.KeyShiftTab:
		if m.selected > 0 {
			m.selected--
		}
	case tea.KeyDown, tea.KeyTab:
		if m.selected < len(m.checks)-1 {
			m.selected++
		}
	case tea.KeyHome:
		m.selected = 0
	case tea.KeyEnd:
		m.selected = len(m.checks) - 1
	case tea.KeyEnter:
		if m.selected >= 0 && m.selected < len(m.checks) {
			m.showDetails = true
			m.detailIndex = m.selected
		}
	case tea.KeyEsc:
		m.quitting = true
		return m, tea.Quit
	case tea.KeyRunes:
		switch msg.String() {
		case "r":
			m.refreshChecks()
		case "q":
			m.quitting = true
			return m, tea.Quit
		case "j":
			if m.selected < len(m.checks)-1 {
				m.selected++
			}
		case "k":
			if m.selected > 0 {
				m.selected--
			}
		}
	}

	return m, nil
}

// handleDetailKey handles keys when detail view is open.
func (m *DashboardModel) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyEnter:
		m.showDetails = false
	}
	return m, nil
}

// refreshChecks runs all health checks.
func (m *DashboardModel) refreshChecks() {
	m.checks = []HealthCheck{}
	m.lastRefresh = time.Now()

	// Run OS check
	m.checks = append(m.checks, m.checkOS())

	// Get all components from bootstrap
	components := bootstrap.AllComponents(m.os)

	// Add chezmoi check
	m.checks = append(m.checks, m.checkChezmoi())

	// Add shell check
	m.checks = append(m.checks, m.checkShell())

	// Add Docker check (both installed and running)
	m.checks = append(m.checks, m.checkDocker())

	// Check each component
	for _, c := range components {
		m.checks = append(m.checks, m.checkComponent(c))
	}

	// Add editor extension checks
	m.checks = append(m.checks, m.checkVSCodeExtensions())
	m.checks = append(m.checks, m.checkCursorExtensions())
}

// checkOS performs the OS health check.
func (m *DashboardModel) checkOS() HealthCheck {
	status := StatusOk
	version := m.os.Name
	if m.os.Arch != "" {
		version = fmt.Sprintf("%s (%s)", m.os.Name, m.os.Arch)
	}
	return HealthCheck{
		Name:      "OS",
		Category:  "system",
		Status:    status,
		Version:   version,
		Message:   m.os.Name,
		Command:   "OS detection",
		CheckedAt: time.Now(),
	}
}

// checkComponent performs a health check for a component.
func (m *DashboardModel) checkComponent(c bootstrap.Component) HealthCheck {
	installed, version := c.IsInstalled(m.os)

	var status Status
	switch {
	case installed && strings.Contains(strings.ToLower(version), "error"):
		status = StatusError
	case installed:
		status = StatusOk
	case strings.Contains(version, "not installed"):
		status = StatusError
	default:
		status = StatusWarn
	}

	return HealthCheck{
		Name:      c.Name,
		Category:  c.Category,
		Status:    status,
		Version:   version,
		Message:   c.Description,
		Command:   c.InstallHint,
		CheckedAt: time.Now(),
	}
}

// checkChezmoi performs the chezmoi health check.
func (m *DashboardModel) checkChezmoi() HealthCheck {
	// Check if chezmoi command exists
	cmd := exec.Command("chezmoi", "--version")
	output, err := cmd.Output()
	if err != nil {
		return HealthCheck{
			Name:      "chezmoi",
			Category:  "tool",
			Status:    StatusError,
			Version:   "not installed",
			Message:   "Dotfiles manager not found",
			Command:   "brew install chezmoi",
			CheckedAt: time.Now(),
		}
	}

	version := strings.TrimSpace(string(output))

	// Try to get last apply date
	lastApplyCmd := exec.Command("chezmoi", "data", "--format", "json")
	lastApplyOutput, _ := lastApplyCmd.Output()

	message := "installed"
	if len(lastApplyOutput) > 0 {
		message = "config active"
	}

	return HealthCheck{
		Name:      "chezmoi",
		Category:  "tool",
		Status:    StatusOk,
		Version:   version,
		Message:   message,
		Command:   "chezmoi --version",
		CheckedAt: time.Now(),
	}
}

// checkShell performs the shell health check.
func (m *DashboardModel) checkShell() HealthCheck {
	// Check if running zsh
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/sh"
	}

	var shellName string
	var shellVersion string

	switch {
	case strings.Contains(shellPath, "zsh"):
		shellName = "zsh"
		cmd := exec.Command("zsh", "--version")
		output, err := cmd.Output()
		if err != nil {
			shellVersion = "unknown"
		} else {
			shellVersion = strings.TrimSpace(string(output))
		}
	case strings.Contains(shellPath, "bash"):
		shellName = "bash"
		cmd := exec.Command("bash", "--version")
		output, err := cmd.Output()
		if err != nil {
			shellVersion = "unknown"
		} else {
			lines := strings.Split(string(output), "\n")
			shellVersion = lines[0]
		}
	default:
		shellName = shellPath
		shellVersion = "unknown"
	}

	status := StatusOk
	if shellName != "zsh" {
		status = StatusWarn
	}

	return HealthCheck{
		Name:      shellName,
		Category:  "system",
		Status:    status,
		Version:   shellVersion,
		Message:   fmt.Sprintf("Current shell: %s", shellName),
		Command:   shellPath,
		CheckedAt: time.Now(),
	}
}

// checkDocker performs the Docker health check (installed AND running).
func (m *DashboardModel) checkDocker() HealthCheck {
	// First check if docker is installed
	version, ok := checkCommandVersion("docker", "--version")
	if !ok {
		return HealthCheck{
			Name:      "docker",
			Category:  "system",
			Status:    StatusError,
			Version:   "not installed",
			Message:   "Docker not found",
			Command:   "curl -fsSL https://get.docker.com | sh",
			CheckedAt: time.Now(),
		}
	}

	// Now check if docker daemon is running
	runningCmd := exec.Command("docker", "info")
	err := runningCmd.Run()

	status := StatusOk
	message := version

	if err != nil {
		status = StatusWarn
		message = version + " (daemon not running)"
	}

	return HealthCheck{
		Name:      "docker",
		Category:  "system",
		Status:    status,
		Version:   version,
		Message:   message,
		Command:   "docker info",
		CheckedAt: time.Now(),
	}
}

// checkVSCodeExtensions performs VS Code extensions check.
func (m *DashboardModel) checkVSCodeExtensions() HealthCheck {
	// Check if code command exists
	if !commandExists("code") {
		return HealthCheck{
			Name:      "VSCode",
			Category:  "extension",
			Status:    StatusError,
			Version:   "not installed",
			Message:   "VS Code not found",
			Command:   "code --version",
			CheckedAt: time.Now(),
		}
	}

	// Get version
	version, _ := checkCommandVersion("code", "--version")

	// Count extensions
	cmd := exec.Command("code", "--list-extensions")
	output, err := cmd.Output()
	extensionCount := 0
	if err == nil {
		lines := strings.Split(string(output), "\n")
		extensionCount = len(lines)
	}

	message := fmt.Sprintf("%d extensions", extensionCount)

	return HealthCheck{
		Name:      "VSCode",
		Category:  "extension",
		Status:    StatusOk,
		Version:   version,
		Message:   message,
		Command:   "code --list-extensions",
		CheckedAt: time.Now(),
	}
}

// checkCursorExtensions performs Cursor extensions check.
func (m *DashboardModel) checkCursorExtensions() HealthCheck {
	// Check if cursor command exists
	if !commandExists("cursor") {
		return HealthCheck{
			Name:      "Cursor",
			Category:  "extension",
			Status:    StatusError,
			Version:   "not installed",
			Message:   "Cursor not found",
			Command:   "cursor --version",
			CheckedAt: time.Now(),
		}
	}

	// Get version
	version, _ := checkCommandVersion("cursor", "--version")

	// Count extensions
	cmd := exec.Command("cursor", "--list-extensions")
	output, err := cmd.Output()
	extensionCount := 0
	if err == nil {
		lines := strings.Split(string(output), "\n")
		extensionCount = len(lines)
	}

	message := fmt.Sprintf("%d extensions", extensionCount)

	return HealthCheck{
		Name:      "Cursor",
		Category:  "extension",
		Status:    StatusOk,
		Version:   version,
		Message:   message,
		Command:   "cursor --list-extensions",
		CheckedAt: time.Now(),
	}
}

// checkCommandVersion checks if a command exists and gets its version.
func checkCommandVersion(cmd string, versionFlag string) (string, bool) {
	versionCmd := exec.Command(cmd, versionFlag)
	output, err := versionCmd.Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(output)), true
}

// commandExists checks if a command exists in PATH.
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// View returns the string view of the model.
func (m *DashboardModel) View() string {
	if m.showDetails {
		return m.viewDetails()
	}
	return m.viewDashboard()
}

// viewDashboard renders the main dashboard view.
func (m *DashboardModel) viewDashboard() string {
	var sb strings.Builder

	// Header
	sb.WriteString(m.renderHeader())

	// OS info bar
	sb.WriteString(m.renderOSBar())

	// Category columns
	sb.WriteString(m.renderCategories())

	// Footer
	sb.WriteString(m.renderFooter())

	return sb.String()
}

// renderHeader renders the dashboard header.
func (m *DashboardModel) renderHeader() string {
	elapsed := ""
	if !m.lastRefresh.IsZero() {
		elapsed = fmt.Sprintf("refresh: %s ago", time.Since(m.lastRefresh).Round(time.Second))
	}

	title := " adacosdev-dots status "
	headerWidth := m.Width - 1
	if headerWidth < 0 {
		headerWidth = 80
	}

	border := borderStyle.Render("")
	topLine := fmt.Sprintf("%s%s%s", border, strings.Repeat("─", headerWidth), border)

	titleLine := fmt.Sprintf("%s %s%s%s %s %s", 
		border, 
		titleStyle.Render(title), 
		strings.Repeat(" ", max(0, headerWidth-len(title)-len(elapsed)-6)), 
		refreshStyle.Render(elapsed), 
		keyStyle.Render("[r]"), 
		border)

	return topLine + "\n" + titleLine + "\n" + borderStyle.Render("") + strings.Repeat("─", headerWidth) + border + "\n"
}

// renderOSBar renders the OS info bar.
func (m *DashboardModel) renderOSBar() string {
	osInfo := fmt.Sprintf("OS: %s", m.os.Name)
	if m.os.Arch != "" {
		osInfo = fmt.Sprintf("OS: %s (%s)", m.os.Name, m.os.Arch)
	}
	border := borderStyle.Render("")
	width := m.Width
	if width < 1 {
		width = 80 // fallback
	}
	return fmt.Sprintf("%s %s\n%s%s%s\n", border, osInfo, border, strings.Repeat("─", width-1), border)
}

// renderCategories renders the category columns.
// renderCategories renders all health checks as a simple left-aligned list.
func (m *DashboardModel) renderCategories() string {
	border := borderStyle.Render("")
	lineWidth := max(m.Width-1, 1)
	sep := strings.Repeat("─", lineWidth)

	var sb strings.Builder

	for _, check := range m.checks {
		extra := ""
		if check.Version != "" {
			extra = " v" + check.Version
		}
		if check.Message != "" {
			extra += " - " + check.Message
		}
		line := check.Name + " " + check.Status.String() + extra
		sb.WriteString(fmt.Sprintf("%s%s%s\n", border, line, border))
	}

	sb.WriteString(fmt.Sprintf("%s%s%s\n", border, sep, border))

	return sb.String()
}

// renderCategoryContent renders the content for a category column.
func (m *DashboardModel) renderCategoryContent(checks []HealthCheck, width int) string {
	var sb strings.Builder
	for i, check := range checks {
		icon := check.Status.Icon()
		statusColor := m.statusStyle(check.Status)

		// Truncate name if needed
		name := check.Name
		if len(name) > width-12 {
			name = name[:width-15] + "..."
		}

		version := check.Version
		if len(version) > 12 {
			version = version[:12] + "..."
		}

		cursor := " "
		if m.getSelectedIndex(check) == m.selected {
			cursor = selectedStyle.Render(">")
		}

		line := fmt.Sprintf("%s %s %s %s %s",
			cursor,
			icon,
			statusColor.Render(name),
			statusColor.Render(version),
			"")

		sb.WriteString(line)
		if i < len(checks)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// getSelectedIndex returns the index in checks for a given check.
func (m *DashboardModel) getSelectedIndex(check HealthCheck) int {
	for i, c := range m.checks {
		if c.Name == check.Name && c.Category == check.Category {
			return i
		}
	}
	return -1
}

// statusStyle returns the lipgloss style for a status.
func (m *DashboardModel) statusStyle(s Status) lipgloss.Style {
	switch s {
	case StatusOk:
		return statusOkStyle
	case StatusWarn:
		return statusWarnStyle
	case StatusError:
		return statusErrorStyle
	default:
		return statusUnknownStyle
	}
}

// renderFooter renders the dashboard footer.
func (m *DashboardModel) renderFooter() string {
	// Count issues
	issueCount := 0
	for _, check := range m.checks {
		if check.Status == StatusWarn || check.Status == StatusError {
			issueCount++
		}
	}

	border := borderStyle.Render("")
	separator := border + strings.Repeat("─", max(m.Width-1, 1)) + border

	var issueText string
	var issueStyle lipgloss.Style
	if issueCount > 0 {
		issueText = fmt.Sprintf("⚠️ %d issue(s) detected", issueCount)
		issueStyle = statusWarnStyle
	} else {
		issueText = "✓ All systems operational"
		issueStyle = statusOkStyle
	}

	footerLine := fmt.Sprintf("%s %s   %s %s   %s %s %s",
		border,
		issueStyle.Render(issueText),
		keyStyle.Render("[Enter]"),
		textStyle.Render("details"),
		keyStyle.Render("[q]"),
		textStyle.Render("quit"),
		border)

	return separator + "\n" + footerLine + "\n"
}

// viewDetails renders the detail popup view.
func (m *DashboardModel) viewDetails() string {
	if m.detailIndex < 0 || m.detailIndex >= len(m.checks) {
		return "No item selected"
	}

	check := m.checks[m.detailIndex]
	statusColor := m.statusStyle(check.Status)

	var sb strings.Builder

	// Title bar
	title := fmt.Sprintf(" %s - Details ", check.Name)
	titleWidth := len(title) + 2
	padding := (m.Width - titleWidth) / 2

	border := borderStyle.Render("")
	separator := border + strings.Repeat("─", max(m.Width-1, 1)) + border

	sb.WriteString(separator + "\n")
	sb.WriteString(fmt.Sprintf("%s%s%s%s%s\n",
		border,
		strings.Repeat(" ", padding),
		titleStyle.Render(title),
		strings.Repeat(" ", m.Width-titleWidth-padding-1),
		border))
	sb.WriteString(separator + "\n")

	// Details
	details := []struct {
		label string
		value string
	}{
		{"Name", check.Name},
		{"Category", check.Category},
		{"Status", fmt.Sprintf("%s %s", check.Status.Icon(), statusColor.Render(check.Status.String()))},
		{"Version", check.Version},
		{"Message", check.Message},
		{"Check Command", check.Command},
		{"Checked At", check.CheckedAt.Format(time.RFC1123)},
	}

	for _, d := range details {
		label := fmt.Sprintf("  %-14s", d.label+":")
		sb.WriteString(fmt.Sprintf("%s%s %s%s\n", border, labelStyle.Render(label), valueStyle.Render(d.value), border))
	}

	// Bottom bar
	sb.WriteString(separator + "\n")
	sb.WriteString(fmt.Sprintf("%s %s %s %s\n", border, keyStyle.Render("[ESC]"), textStyle.Render("Back"), border))

	return sb.String()
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetChecks returns the current health checks.
func (m *DashboardModel) GetChecks() []HealthCheck {
	return m.checks
}

// GetSelected returns the selected check index.
func (m *DashboardModel) GetSelected() int {
	return m.selected
}

// Stop stops the auto-refresh ticker.
func (m *DashboardModel) Stop() {
	if m.autoRefresh != nil {
		m.autoRefresh.Stop()
	}
}

// Categories returns the list of categories.
func Categories() []string {
	return []string{"system", "runtime", "font", "tool", "extension"}
}

// CategoryDisplayName returns a human-readable name for a category.
func CategoryDisplayName(category string) string {
	names := map[string]string{
		"system":    "System",
		"runtime":   "Runtimes",
		"font":      "Fonts",
		"tool":      "Tools",
		"extension": "Extensions",
	}
	if name, ok := names[category]; ok {
		return name
	}
	return category
}

// Run starts the status dashboard TUI.
func Run() error {
	model := NewDashboardModel()

	program := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("failed to run status dashboard: %w", err)
	}

	return nil
}

// RunCLI runs the status dashboard in CLI mode.
func RunCLI() error {
	model := NewDashboardModel()

	// Set reasonable default width/height for CLI
	model.Width = 80
	model.Height = 24

	// Print the dashboard view
	fmt.Println(model.View())

	return nil
}

// Styles for consistent UI rendering.
var (
	// Border style - white/gray
	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	// Title style - cyan for emphasis
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Bold(true)

	// Key style - cyan for keyboard hints
	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))

	// Text style - white for general text
	textStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	// Label style - white for labels
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	// Value style - cyan for values
	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))

	// Refresh style - yellow for time info
	refreshStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))

	// Selected style - highlighted
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	// Status styles
	statusOkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")) // green

	statusWarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")) // yellow

	statusErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")) // red

	statusUnknownStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")) // gray

	// Category styles
	categoryGreen = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")) // green

	categoryYellow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")) // yellow
)

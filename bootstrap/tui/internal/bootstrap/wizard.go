// Package bootstrap provides the bootstrap wizard and component management.
package bootstrap

import (
	"fmt"
	"strings"

	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/bootstrap/detector"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WizardState represents the current state of the wizard.
type WizardState int

const (
	stateWelcome WizardState = iota
	stateSelect
	statePreview
	stateConfirm
	stateInstalling
	stateDone
)

// String returns a string representation of the state.
func (s WizardState) String() string {
	switch s {
	case stateWelcome:
		return "welcome"
	case stateSelect:
		return "select"
	case statePreview:
		return "preview"
	case stateConfirm:
		return "confirm"
	case stateInstalling:
		return "installing"
	case stateDone:
		return "done"
	default:
		return "unknown"
	}
}

// ComponentItem represents a component in the selection list.
type ComponentItem struct {
	Component Component
	Selected  bool
	Status    string // "pending", "installed", "not-installed", "error"
}

// ComponentResult represents the result of installing a component.
type ComponentResult struct {
	Component Component
	Status    string
	Message   string
}

// WizardModel is the Bubbletea model for the bootstrap wizard.
type WizardModel struct {
	state      WizardState
	osInfo     *detector.OSInfo
	items      []ComponentItem
	results    []ComponentResult
	preview    string
	width      int
	height     int
	executor   *Executor
	cursor     int
	scrollOffset int
	spinnerIdx int
	logLines   []string
}

// NewWizardModel creates a new wizard model.
func NewWizardModel(osInfo *detector.OSInfo, helpersPath string) *WizardModel {
	components := AllComponents(osInfo)
	items := make([]ComponentItem, len(components))
	for i, c := range components {
		installed, _ := c.IsInstalled(osInfo)
		status := "not-installed"
		if installed {
			status = "installed"
		}
		items[i] = ComponentItem{
			Component: c,
			Selected:  false,
			Status:    status,
		}
	}

	return &WizardModel{
		state:    stateWelcome,
		osInfo:   osInfo,
		items:    items,
		executor: &Executor{OS: osInfo, Helpers: helpersPath},
	}
}

// Init initializes the wizard model.
func (m *WizardModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model.
func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *WizardModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateWelcome:
		if msg.Type == tea.KeyEnter {
			m.state = stateSelect
		} else if msg.String() == "q" || msg.Type == tea.KeyEsc {
			return m, tea.Quit
		}

	case stateSelect:
		m.handleSelectKey(msg)

	case statePreview:
		if msg.Type == tea.KeyEnter {
			m.state = stateConfirm
		} else if msg.Type == tea.KeyEsc {
			m.state = stateSelect
		}

	case stateConfirm:
		if msg.Type == tea.KeyEnter || msg.String() == "y" || msg.String() == "Y" {
			m.state = stateInstalling
			return m, m.runInstallation()
		} else if msg.String() == "n" || msg.String() == "N" {
			m.state = stateDone
		}

	case stateInstalling:
		// No keyboard input during installation
		if msg.Type == tea.KeyCtrlC {
			// Cancel installation
			m.state = stateDone
		}

	case stateDone:
		if msg.Type == tea.KeyEnter || msg.Type == tea.KeyEsc || msg.String() == "q" {
			return m, tea.Quit
		}
	}

	return m, nil
}

// handleSelectKey handles keyboard input in the select state.
func (m *WizardModel) handleSelectKey(msg tea.KeyMsg) {
	switch msg.Type {
	case tea.KeyUp, tea.KeyShiftTab:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown, tea.KeyTab:
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
	case tea.KeyHome:
		m.cursor = 0
	case tea.KeyEnd:
		m.cursor = len(m.items) - 1
	case tea.KeySpace:
		m.items[m.cursor].Selected = !m.items[m.cursor].Selected
		m.updatePreview()
	case tea.KeyEnter:
		if m.hasSelectedItems() {
			m.state = statePreview
			m.updatePreview()
		}
	case tea.KeyEsc:
		if msg.Type == tea.KeyEsc {
			m.state = stateWelcome
		}
	}

	// Handle 'a' for select all and 'n' for deselect all
	switch msg.String() {
	case "a":
		for i := range m.items {
			if m.items[i].Status != "installed" {
				m.items[i].Selected = true
			}
		}
		m.updatePreview()
	case "n":
		for i := range m.items {
			m.items[i].Selected = false
		}
		m.updatePreview()
	case "q":
		// Can't quit from select state, go to welcome
		m.state = stateWelcome
	}
}

// runInstallation runs the component installation.
func (m *WizardModel) runInstallation() tea.Cmd {
	selectedComponents := m.getSelectedComponents()

	return func() tea.Msg {
		m.logLines = []string{fmt.Sprintf("Starting installation of %d components...", len(selectedComponents))}
		
		results := []ComponentResult{}
		// Build a map for quick lookup
		componentMap := make(map[string]Component)
		for _, c := range selectedComponents {
			componentMap[c.ID] = c
		}
		
		for result := range m.executor.Execute(selectedComponents, false) {
			comp := componentMap[result.Component]
			results = append(results, ComponentResult{
				Component: comp,
				Status:    result.Status,
				Message:   result.Message,
			})
			m.logLines = append(m.logLines, fmt.Sprintf("[%s] %s: %s", 
				strings.ToUpper(result.Status), result.Name, result.Message))
		}
		
		m.results = results
		m.logLines = append(m.logLines, "Installation complete!")
		m.state = stateDone
		
		return tea.KeyMsg{Type: tea.KeyEnter}
	}
}

// updatePreview updates the preview string with selected components.
func (m *WizardModel) updatePreview() {
	var sb strings.Builder
	selected := m.getSelectedComponents()
	sb.WriteString("Selected components:\n\n")
	
	for _, c := range selected {
		sb.WriteString(fmt.Sprintf("  • %s - %s\n", c.Name, c.Description))
	}
	
	m.preview = sb.String()
}

// hasSelectedItems returns true if any items are selected.
func (m *WizardModel) hasSelectedItems() bool {
	for _, item := range m.items {
		if item.Selected {
			return true
		}
	}
	return false
}

// getSelectedComponents returns the list of selected components.
func (m *WizardModel) getSelectedComponents() []Component {
	var selected []Component
	for _, item := range m.items {
		if item.Selected {
			selected = append(selected, item.Component)
		}
	}
	return selected
}

// View returns the string view of the model.
func (m *WizardModel) View() string {
	switch m.state {
	case stateWelcome:
		return m.viewWelcome()
	case stateSelect:
		return m.viewSelect()
	case statePreview:
		return m.viewPreview()
	case stateConfirm:
		return m.viewConfirm()
	case stateInstalling:
		return m.viewInstalling()
	case stateDone:
		return m.viewDone()
	default:
		return "Unknown state"
	}
}

// viewWelcome renders the welcome screen.
func (m *WizardModel) viewWelcome() string {
	logo := `
  ADACOSDEV BOOTSTRAP
  ─────────────────────────────────────────────────────

  Welcome! This wizard will help you set up your development
  environment on ` + m.osInfo.Name + ` (` + m.osInfo.Arch + `).

`

	info := ""
	if m.osInfo.IsWSL {
		info = "  [Running on WSL]\n"
	}

	footer := "\n  [ENTER] Continue    [Q] Quit\n"

	return lipgloss.NewStyle().Width(m.width).Render(logo + info + footer)
}

// viewSelect renders the component selection screen.
func (m *WizardModel) viewSelect() string {
	var sb strings.Builder

	// Group items by category
	categoryMap := make(map[string][]ComponentItem)
	for _, item := range m.items {
		categoryMap[item.Component.Category] = append(categoryMap[item.Component.Category], item)
	}

	// Title
	sb.WriteString("\n  Select components to install\n")
	sb.WriteString("  ─────────────────────────────────────────────────────────\n\n")

	// Navigation hints
	sb.WriteString("  [↑/↓] Navigate  [Space] Toggle  [a] Select all  [n] Deselect all  [Enter] Continue  [q] Quit\n\n")

	// Grouped list
	lineNum := 0
	for _, category := range []string{"system", "runtime", "font", "tool", "extension"} {
		items, ok := categoryMap[category]
		if !ok || len(items) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("  %s\n", CategoryDisplayName(category)))
		
		for _, item := range items {
			checkbox := "[ ]"
			if item.Selected {
				checkbox = "[✓]"
			}

			status := ""
			statusColor := ""
			switch item.Status {
			case "installed":
				status = "installed"
				statusColor = greenStyle.Render("✓")
			case "error":
				status = "error"
				statusColor = redStyle.Render("✗")
			default:
				status = ""
				statusColor = ""
			}

			cursor := " "
			if lineNum == m.cursor {
				cursor = ">"
			}

			name := item.Component.Name
			if item.Status == "installed" {
				name = dimStyle.Render(name + " (installed)")
			}

			line := fmt.Sprintf("  %s %s %s %s %s\n", cursor, checkbox, name, statusColor, status)
			if lineNum == m.cursor {
				line = highlightStyle.Render(line)
			}
			sb.WriteString(line)
			lineNum++
		}
		sb.WriteString("\n")
	}

	// Selected count
	selected := m.countSelected()
	sb.WriteString(fmt.Sprintf("\n  %d component(s) selected\n", selected))

	return sb.String()
}

// viewPreview renders the preview screen.
func (m *WizardModel) viewPreview() string {
	var sb strings.Builder

	sb.WriteString("\n  Preview - Components to install\n")
	sb.WriteString("  ─────────────────────────────────────────────────────────\n\n")

	sb.WriteString(m.preview)

	sb.WriteString("\n  Press ENTER to confirm or ESC to go back\n")

	return sb.String()
}

// viewConfirm renders the confirmation screen.
func (m *WizardModel) viewConfirm() string {
	selected := m.countSelected()

	var sb strings.Builder
	sb.WriteString("\n  Ready to install\n\n")
	sb.WriteString(fmt.Sprintf("  %d component(s) will be installed on %s\n\n", selected, m.osInfo.Name))
	sb.WriteString("  Press ENTER to continue or 'n' to cancel\n")

	return sb.String()
}

// viewInstalling renders the installation progress screen.
func (m *WizardModel) viewInstalling() string {
	var sb strings.Builder

	sb.WriteString("\n  Installing components...\n")
	sb.WriteString("  ─────────────────────────────────────────────────────────\n\n")

	// Show log lines
	for _, line := range m.logLines {
		sb.WriteString(fmt.Sprintf("  %s\n", line))
	}

	sb.WriteString("\n\n  Press Ctrl+C to cancel\n")

	return sb.String()
}

// viewDone renders the completion screen.
func (m *WizardModel) viewDone() string {
	var sb strings.Builder

	sb.WriteString("\n  Installation Complete\n")
	sb.WriteString("  ─────────────────────────────────────────────────────────\n\n")

	// Summary table
	successCount := 0
	errorCount := 0
	for _, r := range m.results {
		status := "❌"
		if r.Status == "installed" || r.Status == "skipped" {
			status = "✅"
			successCount++
		} else if r.Status == "error" {
			errorCount++
		}
		sb.WriteString(fmt.Sprintf("  %s %-20s %s\n", status, r.Component.Name, r.Message))
	}

	sb.WriteString(fmt.Sprintf("\n  ✅ %d succeeded  ❌ %d failed\n", successCount, errorCount))

	sb.WriteString("\n  Press ENTER or q to quit\n")

	return sb.String()
}

// countSelected returns the number of selected items.
func (m *WizardModel) countSelected() int {
	count := 0
	for _, item := range m.items {
		if item.Selected {
			count++
		}
	}
	return count
}

// centerText centers text within the given width.
func centerText(text string, width int) string {
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if len(line) < width {
			padding := (width - len(line)) / 2
			line = strings.Repeat(" ", padding) + line
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// Styles for the wizard UI.
var (
	greenStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // green
	redStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))   // red
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))   // dark gray
	highlightStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	boldStyle      = lipgloss.NewStyle().Bold(true)
)

// Run starts the wizard and returns the final model.
func Run(osInfo *detector.OSInfo, helpersPath string) error {
	model := NewWizardModel(osInfo, helpersPath)
	
	program := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("failed to run wizard: %w", err)
	}
	
	return nil
}

// RunWithOptions starts the wizard with custom options.
func RunWithOptions(osInfo *detector.OSInfo, helpersPath string, dryRun, force bool) error {
	model := NewWizardModel(osInfo, helpersPath)
	model.executor.DryRun = dryRun
	model.executor.Force = force
	
	program := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("failed to run wizard: %w", err)
	}
	
	return nil
}

// GetSelectedComponents returns selected components from a finished wizard run.
func (m *WizardModel) GetSelectedComponents() []Component {
	return m.getSelectedComponents()
}

// GetResults returns the installation results.
func (m *WizardModel) GetResults() []ComponentResult {
	return m.results
}

// Executor returns the executor for testing purposes.
func (m *WizardModel) Executor() *Executor {
	return m.executor
}

// Package select provides the dotfiles selector TUI.
package selectpkg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigCategory represents a category of dotfile configs.
type ConfigCategory struct {
	Name        string
	Description string
	Items       []ConfigItem
}

// ConfigItem represents a single dotfile config.
type ConfigItem struct {
	Name        string // Display name (e.g., "Zsh Shell")
	Destination string // Where it deploys (e.g., "~/.zshrc + ~/.config/zsh/*")
	Source      string // Source path in chezmoi (e.g., "dot_zshrc.tmpl")
	Selected    bool
}

// SelectorState represents the current state of the selector.
type SelectorState int

const (
	stateBrowse SelectorState = iota
	statePreview
	stateApplying
	stateDone
)

// SelectorModel is the Bubbletea model for the dotfiles selector.
type SelectorModel struct {
	state        SelectorState
	categories   []ConfigCategory
	cursor       int
	categoryIdx  int // which category cursor is in
	width        int
	height       int
	DryRun       bool
	Force        bool
	JsonMode     bool
	logLines     []string
	results      []ApplyResult
	scrollOffset int
}

// ApplyResult represents the result of applying a config.
type ApplyResult struct {
	Item   ConfigItem
	Status string // "applied", "skipped", "error"
	Error  error
}

// NewSelectorModel creates a new selector model.
func NewSelectorModel() *SelectorModel {
	model := &SelectorModel{
		state: stateBrowse,
	}
	model.categories = model.discoverConfigs()
	return model
}

// discoverConfigs scans dot_config/ and returns available configs.
func (m *SelectorModel) discoverConfigs() []ConfigCategory {
	categories := []ConfigCategory{}

	// Get chezmoi source directory
	chezmoiDir := os.Getenv("CHEZMOI_SOURCE_DIR")
	if chezmoiDir == "" {
		home, _ := os.UserHomeDir()
		chezmoiDir = filepath.Join(home, ".local/share/chezmoi")
	}

	// Shells category
	shells := ConfigCategory{
		Name:        "Shells",
		Description: "Shell configurations",
		Items:       []ConfigItem{},
	}
	if m.hasConfig(chezmoiDir, "dot_zshrc.tmpl") {
		shells.Items = append(shells.Items, ConfigItem{
			Name:        "Zsh",
			Destination: "~/.zshrc + ~/.config/zsh/*",
			Source:      "dot_zshrc.tmpl",
		})
	}
	if m.hasConfig(chezmoiDir, "dot_config/fish") {
		shells.Items = append(shells.Items, ConfigItem{
			Name:        "Fish",
			Destination: "~/.config/fish/*",
			Source:      "dot_config/fish",
		})
	}
	if m.hasConfig(chezmoiDir, "dot_config/nushell") {
		shells.Items = append(shells.Items, ConfigItem{
			Name:        "Nushell",
			Destination: "~/.config/nushell/*",
			Source:      "dot_config/nushell",
		})
	}
	if len(shells.Items) > 0 {
		categories = append(categories, shells)
	}

	// Terminals category
	terminals := ConfigCategory{
		Name:        "Terminals",
		Description: "Terminal emulator configurations",
		Items:       []ConfigItem{},
	}
	if m.hasConfig(chezmoiDir, "dot_config/kitty") {
		terminals.Items = append(terminals.Items, ConfigItem{
			Name:        "Kitty",
			Destination: "~/.config/kitty/kitty.conf",
			Source:      "dot_config/kitty",
		})
	}
	if m.hasConfig(chezmoiDir, "dot_config/alacritty") {
		terminals.Items = append(terminals.Items, ConfigItem{
			Name:        "Alacritty",
			Destination: "~/.config/alacritty/alacritty.toml",
			Source:      "dot_config/alacritty",
		})
	}
	if m.hasConfig(chezmoiDir, "dot_wezterm.lua") {
		terminals.Items = append(terminals.Items, ConfigItem{
			Name:        "WezTerm",
			Destination: "~/.wezterm.lua",
			Source:      "dot_wezterm.lua",
		})
	}
	if m.hasConfig(chezmoiDir, "dot_config/ghostty") {
		terminals.Items = append(terminals.Items, ConfigItem{
			Name:        "Ghostty",
			Destination: "~/.config/ghostty/*",
			Source:      "dot_config/ghostty",
		})
	}
	if len(terminals.Items) > 0 {
		categories = append(categories, terminals)
	}

	// Multiplexers category
	multiplexers := ConfigCategory{
		Name:        "Multiplexers",
		Description: "Terminal multiplexer configurations",
		Items:       []ConfigItem{},
	}
	if m.hasConfig(chezmoiDir, "dot_tmux.conf") {
		multiplexers.Items = append(multiplexers.Items, ConfigItem{
			Name:        "Tmux",
			Destination: "~/.tmux.conf",
			Source:      "dot_tmux.conf",
		})
	}
	if m.hasConfig(chezmoiDir, "dot_config/zellij") {
		multiplexers.Items = append(multiplexers.Items, ConfigItem{
			Name:        "Zellij",
			Destination: "~/.config/zellij/*",
			Source:      "dot_config/zellij",
		})
	}
	if len(multiplexers.Items) > 0 {
		categories = append(categories, multiplexers)
	}

	// Editors category
	editors := ConfigCategory{
		Name:        "Editors",
		Description: "Editor configurations",
		Items:       []ConfigItem{},
	}
	if m.hasConfig(chezmoiDir, "dot_config/nvim") {
		editors.Items = append(editors.Items, ConfigItem{
			Name:        "Neovim",
			Destination: "~/.config/nvim/*",
			Source:      "dot_config/nvim",
		})
	}
	if len(editors.Items) > 0 {
		categories = append(categories, editors)
	}

	// Tools category
	tools := ConfigCategory{
		Name:        "Tools",
		Description: "CLI tool configurations",
		Items:       []ConfigItem{},
	}
	if m.hasConfig(chezmoiDir, "dot_config/starship.toml.tmpl") || m.hasConfig(chezmoiDir, "dot_config/starship.toml") {
		tools.Items = append(tools.Items, ConfigItem{
			Name:        "Starship",
			Destination: "~/.config/starship.toml",
			Source:      "dot_config/starship.toml.tmpl",
		})
	}
	if m.hasConfig(chezmoiDir, "dot_gitconfig.tmpl") {
		tools.Items = append(tools.Items, ConfigItem{
			Name:        "Git",
			Destination: "~/.gitconfig",
			Source:      "dot_gitconfig.tmpl",
		})
	}
	if len(tools.Items) > 0 {
		categories = append(categories, tools)
	}

	// Claude category
	claude := ConfigCategory{
		Name:        "Claude Code",
		Description: "Claude Code assistant configuration",
		Items:       []ConfigItem{},
	}
	if m.hasConfig(chezmoiDir, "dot_claude") {
		claude.Items = append(claude.Items, ConfigItem{
			Name:        "Claude Settings",
			Destination: "~/.claude/*",
			Source:      "dot_claude",
		})
	}
	if len(claude.Items) > 0 {
		categories = append(categories, claude)
	}

	return categories
}

// hasConfig checks if a config exists in the chezmoi source directory.
func (m *SelectorModel) hasConfig(chezmoiDir, path string) bool {
	fullPath := filepath.Join(chezmoiDir, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// Init initializes the model.
func (m *SelectorModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m *SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case applyDoneMsg:
		m.results = msg.results
		m.state = stateDone
		return m, nil

	default:
		return m, nil
	}
}

type applyDoneMsg struct {
	results []ApplyResult
}

// handleKey handles keyboard input.
func (m *SelectorModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateBrowse:
		return m.handleBrowseKey(msg)
	case statePreview:
		return m.handlePreviewKey(msg)
	case stateDone:
		if msg.Type == tea.KeyEnter || msg.Type == tea.KeyEsc || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// handleBrowseKey handles keyboard input in browse state.
func (m *SelectorModel) handleBrowseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Calculate total items across all categories
	totalItems := 0
	catOffsets := make([]int, len(m.categories))
	for i, cat := range m.categories {
		catOffsets[i] = totalItems
		totalItems += len(cat.Items)
	}

	switch msg.Type {
	case tea.KeyUp, tea.KeyShiftTab:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown, tea.KeyTab:
		if m.cursor < totalItems-1 {
			m.cursor++
		}
	case tea.KeyHome:
		m.cursor = 0
	case tea.KeyEnd:
		m.cursor = totalItems - 1
	case tea.KeySpace:
		m.toggleCurrent()
	case tea.KeyEnter:
		if m.hasSelectedItems() {
			m.state = statePreview
		}
	case tea.KeyEsc:
		return m, tea.Quit
	}

	// Handle 'a' and 'n' keys
	switch msg.String() {
	case "a":
		m.selectAllCurrentCategory()
	case "n":
		m.deselectAllCurrentCategory()
	case "q":
		return m, tea.Quit
	}

	return m, nil
}

// handlePreviewKey handles keyboard input in preview state.
func (m *SelectorModel) handlePreviewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Start applying
		m.state = stateApplying
		return m, m.runApply()
	case tea.KeyEsc:
		m.state = stateBrowse
		m.cursor = 0
	}
	return m, nil
}

// toggleCurrent toggles the selection of the current item.
func (m *SelectorModel) toggleCurrent() {
	for i, cat := range m.categories {
		for j := range cat.Items {
			if m.cursor == 0 {
				m.categories[i].Items[j].Selected = !m.categories[i].Items[j].Selected
				return
			}
			m.cursor--
		}
	}
}

// selectAllCurrentCategory selects all items in the current category.
func (m *SelectorModel) selectAllCurrentCategory() {
	currCursor := m.cursor
	for i, cat := range m.categories {
		if currCursor < len(cat.Items) {
			for j := range cat.Items {
				m.categories[i].Items[j].Selected = true
			}
			return
		}
		currCursor -= len(cat.Items)
	}
}

// deselectAllCurrentCategory deselects all items in the current category.
func (m *SelectorModel) deselectAllCurrentCategory() {
	currCursor := m.cursor
	for i, cat := range m.categories {
		if currCursor < len(cat.Items) {
			for j := range cat.Items {
				m.categories[i].Items[j].Selected = false
			}
			return
		}
		currCursor -= len(cat.Items)
	}
}

// hasSelectedItems returns true if any item is selected.
func (m *SelectorModel) hasSelectedItems() bool {
	for _, cat := range m.categories {
		for _, item := range cat.Items {
			if item.Selected {
				return true
			}
		}
	}
	return false
}

// runApply applies selected configs.
func (m *SelectorModel) runApply() tea.Cmd {
	return func() tea.Msg {
		results := []ApplyResult{}

		for _, cat := range m.categories {
			for _, item := range cat.Items {
				if !item.Selected {
					continue
				}

				// Run chezmoi apply for this specific config
				args := []string{"apply"}
				if m.Force {
					args = append(args, "--force")
				}
				args = append(args, item.Source)

				cmd := exec.Command("chezmoi", args...)
				if m.DryRun {
					cmd.Args = append(cmd.Args, "--dry-run")
				}

				output, err := cmd.CombinedOutput()
				if err != nil {
					results = append(results, ApplyResult{
						Item:   item,
						Status: "error",
						Error:  fmt.Errorf("%s: %w", string(output), err),
					})
				} else {
					results = append(results, ApplyResult{
						Item:   item,
						Status: "applied",
					})
				}
			}
		}

		return applyDoneMsg{results: results}
	}
}

// View renders the TUI.
func (m *SelectorModel) View() string {
	switch m.state {
	case stateBrowse:
		return m.viewBrowse()
	case statePreview:
		return m.viewPreview()
	case stateApplying:
		return m.viewApplying()
	case stateDone:
		return m.viewDone()
	}
	return ""
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7FB4CA")).
			MarginBottom(1)

	categoryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#B7CC85")).
			MarginTop(1)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F3F6F9"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF8DD7"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DEBA87"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5C6170")).
			MarginTop(1)

	previewStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F3F6F9"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CB7C94"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B7CC85"))
)

// viewBrowse renders the browse state.
func (m *SelectorModel) viewBrowse() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  ADACOSDEV DOTFILES — Select configs to apply"))
	b.WriteString("\n")

	cursor := 0
	for _, cat := range m.categories {
		b.WriteString(categoryStyle.Render(fmt.Sprintf("  %s", cat.Name)))
		b.WriteString("\n")

		for _, item := range cat.Items {
			checkbox := "[ ]"
			if item.Selected {
				checkbox = selectedStyle.Render("[✓]")
			}

			line := fmt.Sprintf("    %s %s", checkbox, item.Name)
			if cursor == m.cursor {
				line = cursorStyle.Render("▸ ") + line[2:]
			} else {
				line = "  " + line[2:]
			}

			b.WriteString(itemStyle.Render(line))
			b.WriteString("\n")
			cursor++
		}
	}

	b.WriteString(helpStyle.Render("\n  ↑/↓ navigate  •  Space toggle  •  a select category  •  n deselect category  •  Enter confirm  •  q quit"))

	return b.String()
}

// viewPreview renders the preview state.
func (m *SelectorModel) viewPreview() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  ADACOSDEV DOTFILES — Preview"))
	b.WriteString("\n\n")

	for _, cat := range m.categories {
		for _, item := range cat.Items {
			if item.Selected {
				b.WriteString(fmt.Sprintf("  • %s → %s\n", item.Name, item.Destination))
			}
		}
	}

	b.WriteString(helpStyle.Render("\n  Enter apply  •  ESC back"))

	return b.String()
}

// viewApplying renders the applying state.
func (m *SelectorModel) viewApplying() string {
	return titleStyle.Render("  Applying selected configs...")
}

// viewDone renders the done state.
func (m *SelectorModel) viewDone() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  ADACOSDEV DOTFILES — Results"))
	b.WriteString("\n\n")

	applied := 0
	errors := 0

	for _, result := range m.results {
		var statusIcon string
		var line string

		if result.Status == "applied" {
			statusIcon = successStyle.Render("✅")
			line = fmt.Sprintf("  %s %s → %s\n", statusIcon, result.Item.Name, result.Item.Destination)
			applied++
		} else {
			statusIcon = errorStyle.Render("❌")
			line = fmt.Sprintf("  %s %s → %s: %v\n", statusIcon, result.Item.Name, result.Item.Destination, result.Error)
			errors++
		}

		b.WriteString(line)
	}

	b.WriteString(fmt.Sprintf("\n  Applied: %d  Errors: %d\n", applied, errors))
	b.WriteString(helpStyle.Render("\n  Enter/ESC/q to quit"))

	return b.String()
}

// CLI mode functions

// RunCLI runs the selector in CLI mode.
func (m *SelectorModel) RunCLI(jsonMode bool) error {
	m.JsonMode = jsonMode

	if jsonMode {
		return m.runJSON()
	}

	return m.runText()
}

// runJSON outputs available configs as JSON.
func (m *SelectorModel) runJSON() error {
	// Output categories and items as JSON
	fmt.Println(`{"type":"info","message":"select command not yet implemented in JSON mode"}`)
	return nil
}

// runText outputs available configs as text.
func (m *SelectorModel) runText() error {
	fmt.Println("ADACOSDEV DOTFILES — Available configs")
	fmt.Println()

	for _, cat := range m.categories {
		fmt.Printf("%s:\n", cat.Name)
		for _, item := range cat.Items {
			fmt.Printf("  - %s → %s\n", item.Name, item.Destination)
		}
		fmt.Println()
	}

	fmt.Println("Use --json for JSON output")
	return nil
}

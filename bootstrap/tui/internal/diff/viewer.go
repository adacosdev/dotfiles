// Package diff provides a TUI for viewing chezmoi diff output.
package diff

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode represents the diff display mode.
type ViewMode string

const (
	ViewModeUnified    ViewMode = "unified"
	ViewModeSideBySide ViewMode = "side-by-side"
)

// DiffFile represents a file with diff hunks.
type DiffFile struct {
	Path   string
	Hunks  []DiffHunk
	Status string // "added", "deleted", "modified", "unchanged"
}

// DiffHunk represents a hunk in a diff.
type DiffHunk struct {
	Lines    []DiffLine
	OldStart int
	OldCount int
	NewStart int
	NewCount int
}

// DiffLine represents a single line in a hunk.
type DiffLine struct {
	Content string
	Type    string // "context", "add", "del", "header"
	OldNum  int
	NewNum  int
}

// DiffModel is the Bubbletea model for the diff viewer.
type DiffModel struct {
	files        []DiffFile
	filtered     []DiffFile
	selected     int
	viewMode     ViewMode
	filter       string
	width        int
	height       int
	focus        string // "filelist", "diff"
	loading      bool
	errorMsg     string
	diffOutput   string
	scrollOffset int
	hunkOffset   int
}

// NewDiffModel creates a new diff viewer model.
func NewDiffModel() *DiffModel {
	return &DiffModel{
		selected:  0,
		viewMode: ViewModeSideBySide,
		focus:    "filelist",
		loading:  true,
	}
}

// Init initializes the diff viewer model.
func (m *DiffModel) Init() tea.Cmd {
	return m.loadDiff()
}

// loadDiff runs chezmoi diff and parses the output.
func (m *DiffModel) loadDiff() tea.Cmd {
	return func() tea.Msg {
		// Run chezmoi diff --no-color
		cmd := exec.Command("chezmoi", "diff", "--color=never")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		if err != nil {
			m.errorMsg = fmt.Sprintf("failed to run chezmoi diff: %v\n%s", err, stderr.String())
			m.loading = false
			return nil
		}

		m.diffOutput = stdout.String()
		m.files = ParseDiff(m.diffOutput)
		m.filtered = m.files
		m.loading = false

		return tea.WindowSizeMsg{Width: 80, Height: 24}
	}
}

// Update handles messages and updates the model.
func (m *DiffModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

// handleKey handles keyboard input.
func (m *DiffModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp, tea.KeyShiftTab:
		if m.focus == "filelist" {
			if m.selected > 0 {
				m.selected--
				m.scrollOffset = 0
				m.hunkOffset = 0
			}
		} else {
			if m.hunkOffset > 0 {
				m.hunkOffset--
			}
		}

	case tea.KeyDown, tea.KeyTab:
		if m.focus == "filelist" {
			if m.selected < len(m.filtered)-1 {
				m.selected++
				m.scrollOffset = 0
				m.hunkOffset = 0
			}
		} else {
			m.hunkOffset++
		}

	case tea.KeyHome:
		if m.focus == "filelist" {
			m.selected = 0
			m.scrollOffset = 0
		} else {
			m.hunkOffset = 0
		}

	case tea.KeyEnd:
		if m.focus == "filelist" {
			m.selected = len(m.filtered) - 1
			m.scrollOffset = 0
		}

	case tea.KeyLeft:
		m.focus = "filelist"

	case tea.KeyRight:
		m.focus = "diff"

	case tea.KeyRunes:
		switch msg.String() {
		case "j":
			if m.focus == "filelist" {
				if m.selected < len(m.filtered)-1 {
					m.selected++
					m.scrollOffset = 0
					m.hunkOffset = 0
				}
			} else {
				m.hunkOffset++
			}
		case "k":
			if m.focus == "filelist" {
				if m.selected > 0 {
					m.selected--
					m.scrollOffset = 0
					m.hunkOffset = 0
				}
			} else {
				if m.hunkOffset > 0 {
					m.hunkOffset--
				}
			}
		case "h":
			m.focus = "filelist"
		case "l":
			m.focus = "diff"
		case "\t":
			if m.viewMode == ViewModeUnified {
				m.viewMode = ViewModeSideBySide
			} else {
				m.viewMode = ViewModeUnified
			}
		case "q", "Q":
			return m, tea.Quit
		}
	}

	return m, nil
}

// View returns the string view of the model.
func (m *DiffModel) View() string {
	if m.loading {
		return "\n  Loading diff...\n"
	}

	if m.errorMsg != "" {
		return fmt.Sprintf("\n  Error: %s\n", m.errorMsg)
	}

	if len(m.filtered) == 0 {
		return "\n  No changes found.\n\n  Press q to quit.\n"
	}

	var sb strings.Builder

	// Calculate dimensions
	fileListWidth := 35
	if m.width < 80 {
		fileListWidth = m.width / 3
	}
	diffWidth := m.width - fileListWidth - 3

	// Header
	sb.WriteString("\n")
	sb.WriteString(headerStyle.Render(" chezmoi diff "))
	sb.WriteString(fmt.Sprintf("  %d file(s)", len(m.filtered)))
	if m.viewMode == ViewModeSideBySide {
		sb.WriteString("  [Tab] unified  [q] quit\n\n")
	} else {
		sb.WriteString("  [Tab] side-by-side  [q] quit\n\n")
	}

	// File list
	fileList := m.renderFileList(fileListWidth)

	// Diff view
	diffView := m.renderDiffView(diffWidth)

	// Combine side by side
	if m.viewMode == ViewModeSideBySide {
		sb.WriteString(lipgloss.JoinHorizontal(
			lipgloss.Top,
			fileList,
			"| ",
			diffView,
		))
	} else {
		sb.WriteString(fileList)
		sb.WriteString("\n")
		sb.WriteString(diffView)
	}

	return sb.String()
}

// renderFileList renders the file list panel.
func (m *DiffModel) renderFileList(width int) string {
	var sb strings.Builder

	listHeight := m.height - 6
	if listHeight < 1 {
		listHeight = 10
	}

	// Panel header
	sb.WriteString(fileListHeaderStyle.Width(width).Render(" Files "))
	sb.WriteString("\n")

	// Files
	visibleFiles := m.filtered
	if len(visibleFiles) > listHeight {
		start := m.scrollOffset
		end := start + listHeight
		if end > len(visibleFiles) {
			end = len(visibleFiles)
		}
		visibleFiles = visibleFiles[start:end]
	}

	for i, file := range visibleFiles {
		actualIndex := i + m.scrollOffset
		isSelected := actualIndex == m.selected
		isFocused := m.focus == "filelist"

		statusStr := getStatusIcon(file.Status)
		statusStyle := getStatusStyle(file.Status)

		name := file.Path
		if len(name) > width-6 {
			name = name[:width-9] + "..."
		}

		line := fmt.Sprintf(" %s %s %s", statusStr, statusStyle.Render(name), getStatusBadge(file.Status))

		if isSelected {
			if isFocused {
				line = selectedFocusedStyle.Width(width).Render(line)
			} else {
				line = selectedStyle.Width(width).Render(line)
			}
		}

		sb.WriteString(line)
		if i < len(visibleFiles)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// renderDiffView renders the diff content panel.
func (m *DiffModel) renderDiffView(width int) string {
	if m.selected >= len(m.filtered) {
		return emptyDiffStyle.Width(width).Render(" No file selected ")
	}

	file := m.filtered[m.selected]
	if len(file.Hunks) == 0 {
		return emptyDiffStyle.Width(width).Render(" No changes in this file ")
	}

	var sb strings.Builder

	// Panel header with file path
	headerText := " " + file.Path + " "
	sb.WriteString(diffHeaderStyle.Width(width).Render(headerText))
	sb.WriteString("\n")

	// Render hunks
	hunkHeight := m.height - 8
	if hunkHeight < 1 {
		hunkHeight = 10
	}

	visibleLines := 0
	for i, hunk := range file.Hunks {
		// Hunk header
		hunkHeader := fmt.Sprintf("@@ -%d,%d +%d,%d @@",
			hunk.OldStart, hunk.OldCount, hunk.NewStart, hunk.NewCount)

		if visibleLines < m.hunkOffset {
			// Skip lines before scroll offset
			visibleLines += len(hunk.Lines) + 1 // +1 for hunk header
			continue
		}

		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(hunkHeaderStyle.Width(width).Render(" " + hunkHeader))
		sb.WriteString("\n")

		// Hunk lines
		for _, line := range hunk.Lines {
			if visibleLines >= m.hunkOffset+hunkHeight {
				break
			}
			visibleLines++

			rendered := m.renderLine(line, width)
			sb.WriteString(rendered)
			sb.WriteString("\n")
		}

		if visibleLines >= m.hunkOffset+hunkHeight {
			break
		}
	}

	return sb.String()
}

// renderLine renders a single diff line.
func (m *DiffModel) renderLine(line DiffLine, width int) string {
	switch line.Type {
	case "add":
		return addLineStyle.Width(width).Render("+ " + line.Content)
	case "del":
		return delLineStyle.Width(width).Render("- " + line.Content)
	case "header":
		return hunkHeaderStyle.Width(width).Render(line.Content)
	default:
		return contextLineStyle.Width(width).Render(" " + line.Content)
	}
}

// ParseDiff parses the output of `chezmoi diff --color=never`.
func ParseDiff(output string) []DiffFile {
	var files []DiffFile
	var currentFile *DiffFile
	var currentHunk *DiffHunk

	oldNum := 0
	newNum := 0

	scanner := bufio.NewScanner(strings.NewReader(output))
	// Increase buffer for long lines
	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	// Regex for hunk header
	hunkRe := regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
	// Regex for diff --git line
	diffGitRe := regexp.MustCompile(`^diff --git a/(.+) b/.+$`)

	for scanner.Scan() {
		line := scanner.Text()

		// New file entry
		if strings.HasPrefix(line, "diff --git a/") {
			// Save previous file
			if currentFile != nil {
				if currentHunk != nil {
					currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
				}
				files = append(files, *currentFile)
			}

			// Extract path from "diff --git a/path b/path"
			matches := diffGitRe.FindStringSubmatch(line)
			path := ""
			if len(matches) > 1 {
				path = matches[1]
			}

			currentFile = &DiffFile{
				Path:   path,
				Hunks:  []DiffHunk{},
				Status: "modified",
			}
			currentHunk = nil
			oldNum = 0
			newNum = 0
			continue
		}

		// File paths and status from --- and +++ lines
		if strings.HasPrefix(line, "--- ") {
			path := strings.TrimPrefix(line, "--- a/")
			path = strings.TrimPrefix(path, "--- /dev/null")
			path = strings.TrimSpace(path)
			path = strings.Split(path, "\t")[0]
			if path == "/dev/null" {
				path = currentFile.Path
			}

			if currentFile != nil && currentFile.Path == "" {
				currentFile.Path = path
			}

			// If --- is /dev/null, it's a new file (status will be set to "added")
			if line == "--- /dev/null" {
				if currentFile != nil {
					currentFile.Status = "added"
				}
			}
			continue
		}

		if strings.HasPrefix(line, "+++ ") {
			path := strings.TrimPrefix(line, "+++ b/")
			path = strings.TrimPrefix(path, "+++ /dev/null")
			path = strings.TrimSpace(path)
			if path == "/dev/null" {
				path = currentFile.Path
			}

			if currentFile != nil {
				if line == "+++ /dev/null" {
					currentFile.Status = "deleted"
				} else if currentFile.Status != "added" {
					// If not added and not deleted, it's modified
					currentFile.Status = "modified"
				}
			}
			continue
		}

		// Hunk header
		if strings.HasPrefix(line, "@@") {
			// Save previous hunk
			if currentHunk != nil && currentFile != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}

			matches := hunkRe.FindStringSubmatch(line)
			if len(matches) >= 4 {
				oldNum, _ = atoiSafe(matches[1])
				newNum, _ = atoiSafe(matches[3])

				currentHunk = &DiffHunk{
					Lines:    []DiffLine{},
					OldStart: oldNum,
					NewStart: newNum,
				}
			}
			continue
		}

		// Diff content lines
		if currentHunk != nil {
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Content: strings.TrimPrefix(line, "+"),
					Type:    "add",
					OldNum:  0,
					NewNum:  newNum,
				})
				newNum++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Content: strings.TrimPrefix(line, "-"),
					Type:    "del",
					OldNum:  oldNum,
					NewNum:  0,
				})
				oldNum++
			} else if strings.HasPrefix(line, " ") || line == "" {
				content := line
				if strings.HasPrefix(line, " ") {
					content = strings.TrimPrefix(line, " ")
				}
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Content: content,
					Type:    "context",
					OldNum:  oldNum,
					NewNum:  newNum,
				})
				oldNum++
				newNum++
			}
		}
	}

	// Save last file and hunk
	if currentFile != nil {
		if currentHunk != nil {
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
		}
		files = append(files, *currentFile)
	}

	// Handle case where there's no "diff --git" header but there are changes
	if len(files) == 0 && output != "" {
		// Try to parse as simple diff format
		files = parseSimpleDiff(output)
	}

	return files
}

// parseSimpleDiff parses a simple diff format when no diff header is present.
func parseSimpleDiff(output string) []DiffFile {
	var files []DiffFile
	var currentFile *DiffFile
	var currentHunk *DiffHunk

	scanner := bufio.NewScanner(strings.NewReader(output))
	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for file path indicators
		if strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") {
			path := strings.TrimPrefix(line, "--- ")
			path = strings.TrimPrefix(path, "+++ ")
			path = strings.Split(path, "\t")[0]
			path = strings.TrimSpace(path)

			if currentFile == nil {
				currentFile = &DiffFile{
					Path:   path,
					Hunks:  []DiffHunk{},
					Status: "modified",
				}
			} else {
				currentFile.Path = path
			}
			continue
		}

		// Hunk header
		if strings.HasPrefix(line, "@@") {
			if currentHunk != nil && currentFile != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}

			hunkRe := regexp.MustCompile(`@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
			matches := hunkRe.FindStringSubmatch(line)
			if len(matches) >= 4 {
				oldNum, _ := atoiSafe(matches[1])
				newNum, _ := atoiSafe(matches[3])

				currentHunk = &DiffHunk{
					Lines:    []DiffLine{},
					OldStart: oldNum,
					NewStart: newNum,
				}
			}
			continue
		}

		// Content lines
		if currentHunk != nil {
			if strings.HasPrefix(line, "+") {
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Content: strings.TrimPrefix(line, "+"),
					Type:    "add",
				})
			} else if strings.HasPrefix(line, "-") {
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Content: strings.TrimPrefix(line, "-"),
					Type:    "del",
				})
			} else if strings.HasPrefix(line, " ") || line == "" {
				content := line
				if strings.HasPrefix(line, " ") {
					content = strings.TrimPrefix(line, " ")
				}
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{
					Content: content,
					Type:    "context",
				})
			}
		}
	}

	if currentFile != nil {
		if currentHunk != nil {
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
		}
		files = append(files, *currentFile)
	}

	return files
}

// atoiSafe safely converts a string to int.
func atoiSafe(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// getStatusIcon returns the icon for a file status.
func getStatusIcon(status string) string {
	switch status {
	case "added":
		return "+"
	case "deleted":
		return "-"
	case "modified":
		return "~"
	default:
		return "="
	}
}

// getStatusStyle returns the lipgloss style for a file status.
func getStatusStyle(status string) lipgloss.Style {
	switch status {
	case "added":
		return addedFileStyle
	case "deleted":
		return deletedFileStyle
	case "modified":
		return modifiedFileStyle
	default:
		return unchangedFileStyle
	}
}

// getStatusBadge returns a status badge string.
func getStatusBadge(status string) string {
	switch status {
	case "added":
		return "[added]"
	case "deleted":
		return "[deleted]"
	case "modified":
		return "[modified]"
	default:
		return "[unchanged]"
	}
}

// Run starts the diff viewer TUI.
func Run() error {
	model := NewDiffModel()

	program := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("failed to run diff viewer: %w", err)
	}

	return nil
}

// RunInCLI runs the diff viewer in CLI mode (plain text output).
func RunInCLI() error {
	cmd := exec.Command("chezmoi", "diff", "--color=never")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Styles for the diff viewer UI.
var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).  // white
			Background(lipgloss.Color("8")).  // dark gray
			Bold(true).
			Padding(0, 1)

	fileListHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7")).  // white
				Background(lipgloss.Color("8")).  // dark gray
				Bold(true)

	diffHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).  // white
			Background(lipgloss.Color("8")). // dark gray
			Bold(true)

	hunkHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")) // cyan

	// Line styles with colors matching requirements
	addLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("52")).  // dark green
			Foreground(lipgloss.Color("2"))   // green

	delLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("52")).  // dark red
			Foreground(lipgloss.Color("1"))   // red

	contextLineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7")) // white/gray

	// File status styles
	addedFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")) // green

	deletedFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")) // red

	modifiedFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")) // yellow

	unchangedFileStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")) // dark gray

	// Selection styles
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("8")). // dark gray
			Foreground(lipgloss.Color("7"))   // white

	selectedFocusedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("4")).  // blue background
				Foreground(lipgloss.Color("7")).   // white text
				Bold(true)

	emptyDiffStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")) // dark gray
)

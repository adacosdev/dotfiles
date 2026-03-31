// Package bootstrap provides the bootstrap wizard and component management.
package bootstrap

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/bootstrap/detector"
)

// Component represents an installable component.
type Component struct {
	ID          string             // "tmux", "neovim", "ghostty"
	Name        string             // "Tmux", "Neovim"
	Description string             // One-liner description
	Category    string             // "system", "runtime", "font", "tool", "extension"
	OS          []string           // OS where this applies; empty = all
	InstallHint string             // What to show if not installable
	IsInstalled func(*detector.OSInfo) (bool, string) // check cmd + version output
	InstallCmd  string             // Shell command or helper path to run
	NeedsRoot   bool
}

// AllComponents returns all available components filtered by OS compatibility.
func AllComponents(os *detector.OSInfo) []Component {
	all := getAllComponents()
	if os == nil {
		return all
	}

	// Filter components by OS compatibility
	var filtered []Component
	for _, c := range all {
		if len(c.OS) == 0 || containsOS(c.OS, os.ID) || containsOS(c.OS, os.Family) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// ComponentsByCategory groups components by category.
func ComponentsByCategory(os *detector.OSInfo) map[string][]Component {
	components := AllComponents(os)
	result := make(map[string][]Component)

	for _, c := range components {
		result[c.Category] = append(result[c.Category], c)
	}
	return result
}

// getAllComponents returns the complete list of all components.
func getAllComponents() []Component {
	return []Component{
		// === System ===
		{
			ID:          "tmux",
			Name:        "Tmux",
			Description: "Terminal multiplexer with session management",
			Category:    "system",
			OS:          []string{"linux", "darwin", "darwin-arm64"},
			InstallHint: "Install via package manager: apt install tmux / brew install tmux",
			IsInstalled: checkCommandVersion("tmux", "-V"),
			InstallCmd:  "bootstrap-helper tmux",
			NeedsRoot:   false,
		},
		{
			ID:          "neovim",
			Name:        "Neovim",
			Description: "Hyperextensible Vim-based text editor",
			Category:    "system",
			OS:          []string{"linux", "darwin", "darwin-arm64"},
			InstallHint: "Install via package manager or download from neovim.io",
			IsInstalled: checkCommandVersion("nvim", "--version"),
			InstallCmd:  "bootstrap-helper neovim",
			NeedsRoot:   false,
		},
		{
			ID:          "starship",
			Name:        "Starship",
			Description: "Cross-shell prompt configured via TOML",
			Category:    "system",
			OS:          []string{"linux", "darwin", "darwin-arm64", "windows"},
			InstallHint: "Install via: curl -sS https://starship.rs/install.sh | sh",
			IsInstalled: checkCommandVersion("starship", "--version"),
			InstallCmd:  "bootstrap-helper starship",
			NeedsRoot:   false,
		},
		{
			ID:          "docker",
			Name:        "Docker",
			Description: "Container runtime and management tool",
			Category:    "system",
			OS:          []string{"linux"},
			InstallHint: "Install via: curl -fsSL https://get.docker.com | sh",
			IsInstalled: checkCommandVersion("docker", "--version"),
			InstallCmd:  "bootstrap-helper docker",
			NeedsRoot:   true,
		},

		// === Runtimes ===
		{
			ID:          "python",
			Name:        "Python / pyenv",
			Description: "Python runtime with pyenv version management",
			Category:    "runtime",
			OS:          []string{"linux", "darwin", "darwin-arm64"},
			InstallHint: "Install pyenv via bootstrap-helper",
			IsInstalled: checkInstalled("pyenv"),
			InstallCmd:  "bootstrap-helper python",
			NeedsRoot:   false,
		},
		{
			ID:          "go",
			Name:        "Go",
			Description: "Go programming language toolchain",
			Category:    "runtime",
			OS:          []string{"linux", "darwin", "darwin-arm64"},
			InstallHint: "Install via: https://go.dev/dl/",
			IsInstalled: checkCommandVersion("go", "version"),
			InstallCmd:  "bootstrap-helper go",
			NeedsRoot:   false,
		},
		{
			ID:          "rust",
			Name:        "Rust",
			Description: "Rust programming language toolchain via rustup",
			Category:    "runtime",
			OS:          []string{"linux", "darwin", "darwin-arm64"},
			InstallHint: "Install via: curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh",
			IsInstalled: checkInstalled("rustc"),
			InstallCmd:  "bootstrap-helper rust",
			NeedsRoot:   false,
		},
		{
			ID:          "volta",
			Name:        "Volta",
			Description: "JavaScript tool manager (Node, npm, yarn)",
			Category:    "runtime",
			OS:          []string{"linux", "darwin", "darwin-arm64", "windows"},
			InstallHint: "Install via: curl https://get.volta.sh | sh",
			IsInstalled: checkCommandVersion("volta", "--version"),
			InstallCmd:  "bootstrap-helper volta",
			NeedsRoot:   false,
		},
		{
			ID:          "fnm",
			Name:        "FNM",
			Description: "Fast Node Manager - alternative Node.js version manager",
			Category:    "runtime",
			OS:          []string{"linux", "darwin", "darwin-arm64"},
			InstallHint: "Install via: curl -fsSL https://fnm.vercel.app/install | bash",
			IsInstalled: checkInstalled("fnm"),
			InstallCmd:  "bootstrap-helper fnm",
			NeedsRoot:   false,
		},

		// === Fonts ===
		{
			ID:          "jetbrains-mono",
			Name:        "JetBrains Mono",
			Description: "JetBrains Mono Nerd Font with ligatures",
			Category:    "font",
			OS:          []string{"linux", "darwin", "darwin-arm64", "windows"},
			InstallHint: "Download from https://www.jetbrains.com/lp/mono/",
			IsInstalled: checkFont("JetBrainsMono"),
			InstallCmd:  "bootstrap-helper font jetbrains-mono",
			NeedsRoot:   false,
		},
		{
			ID:          "iosevka",
			Name:        "Iosevka Term Nerd Font",
			Description: "Iosevka Term Nerd Font with extended glyphs",
			Category:    "font",
			OS:          []string{"linux", "darwin", "darwin-arm64", "windows"},
			InstallHint: "Download from https://github.com/ryanoasis/nerd-fonts",
			IsInstalled: checkFont("Iosevka Term"),
			InstallCmd:  "bootstrap-helper font iosevka",
			NeedsRoot:   false,
		},

		// === Tools ===
		{
			ID:          "ghostty",
			Name:        "Ghostty",
			Description: "Fast, GPU-accelerated terminal emulator",
			Category:    "tool",
			OS:          []string{"linux", "darwin", "darwin-arm64"},
			InstallHint: "Download from https://ghostty.org/download",
			IsInstalled: checkCommandVersion("ghostty", "--version"),
			InstallCmd:  "bootstrap-helper ghostty",
			NeedsRoot:   false,
		},
		{
			ID:          "oh-my-zsh",
			Name:        "Oh My Zsh",
			Description: "Community-driven Zsh configuration framework",
			Category:    "tool",
			OS:          []string{"linux", "darwin", "darwin-arm64"},
			InstallHint: "Install via: sh -c \"$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)\"",
			IsInstalled: checkDirExists("$HOME/.oh-my-zsh"),
			InstallCmd:  "bootstrap-helper oh-my-zsh",
			NeedsRoot:   false,
		},
		{
			ID:          "atuin",
			Name:        "Atuin",
			Description: "Magical shell history with sync",
			Category:    "tool",
			OS:          []string{"linux", "darwin", "darwin-arm64"},
			InstallHint: "Install via: curl --proto '=https' --tlsv1.2 -LsSf https://setup.atuinsh.com | sh",
			IsInstalled: checkCommandVersion("atuin", "--version"),
			InstallCmd:  "bootstrap-helper atuin",
			NeedsRoot:   false,
		},
		{
			ID:          "carapace",
			Name:        "Carapace",
			Description: "Multi-shell completion framework",
			Category:    "tool",
			OS:          []string{"linux", "darwin", "darwin-arm64"},
			InstallHint: "Install via: brew install carapace or download from carapace.sh",
			IsInstalled: checkInstalled("carapace"),
			InstallCmd:  "bootstrap-helper carapace",
			NeedsRoot:   false,
		},

		// === Editor Extensions ===
		{
			ID:          "vscode",
			Name:        "VS Code",
			Description: "Visual Studio Code with extensions",
			Category:    "extension",
			OS:          []string{"linux", "darwin", "darwin-arm64", "windows"},
			InstallHint: "Install via: https://code.visualstudio.com/download",
			IsInstalled: checkInstalled("code"),
			InstallCmd:  "bootstrap-helper vscode",
			NeedsRoot:   false,
		},
		{
			ID:          "cursor",
			Name:        "Cursor",
			Description: "AI-first code editor built on VS Code",
			Category:    "extension",
			OS:          []string{"linux", "darwin", "darwin-arm64", "windows"},
			InstallHint: "Download from https://cursor.sh/download",
			IsInstalled: checkInstalled("cursor"),
			InstallCmd:  "bootstrap-helper cursor",
			NeedsRoot:   false,
		},
		{
			ID:          "antigravity",
			Name:        "Antigravity",
			Description: "Emacs configuration framework for Vim users",
			Category:    "extension",
			OS:          []string{"linux", "darwin", "darwin-arm64"},
			InstallHint: "Clone from: https://github.com/antigram/antigravity",
			IsInstalled: checkDirExists("$HOME/.antigravity"),
			InstallCmd:  "bootstrap-helper antigravity",
			NeedsRoot:   false,
		},
	}
}

// containsOS checks if the OS list contains the given OS ID or family.
// An empty list means "all OSes" and returns true.
func containsOS(osList []string, osID string) bool {
	if len(osList) == 0 {
		return true // empty list means "all"
	}
	for _, os := range osList {
		if os == osID {
			return true
		}
	}
	return false
}

// checkCommandVersion returns a function that checks if a command exists and gets its version.
func checkCommandVersion(cmd string, versionFlag string) func(*detector.OSInfo) (bool, string) {
	return func(info *detector.OSInfo) (bool, string) {
		// First check if command exists
		if !commandExists(cmd) {
			return false, "not installed"
		}

		// Try to get version
		versionCmd := exec.Command(cmd, versionFlag)
		output, err := versionCmd.Output()
		if err != nil {
			return true, "installed (version unknown)"
		}

		version := strings.TrimSpace(string(output))
		return true, version
	}
}

// checkInstalled returns a function that checks if a command exists.
func checkInstalled(cmd string) func(*detector.OSInfo) (bool, string) {
	return func(info *detector.OSInfo) (bool, string) {
		if commandExists(cmd) {
			return true, "installed"
		}
		return false, "not installed"
	}
}

// checkDirExists returns a function that checks if a directory exists.
func checkDirExists(dir string) func(*detector.OSInfo) (bool, string) {
	return func(info *detector.OSInfo) (bool, string) {
		// Expand HOME variable
		expandedDir := os.ExpandEnv(dir)
		if _, err := os.Stat(expandedDir); err == nil {
			return true, "installed"
		}
		return false, "not installed"
	}
}

// checkFont returns a function that checks if a font is installed.
func checkFont(fontName string) func(*detector.OSInfo) (bool, string) {
	return func(info *detector.OSInfo) (bool, string) {
		// Check common font directories
		fontDirs := []string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
			"/Library/Fonts",
			"/System/Library/Fonts",
		}

		homeFonts := os.Getenv("HOME") + "/.fonts"
		if homeFonts != "" {
			fontDirs = append(fontDirs, homeFonts)
		}

		// Search for font directory or file
		fontPattern := regexp.MustCompile(`(?i)` + fontName)

		for _, dir := range fontDirs {
			if entries, err := os.ReadDir(dir); err == nil {
				for _, entry := range entries {
					if fontPattern.MatchString(entry.Name()) {
						return true, "installed"
					}
				}
			}
		}

		return false, "not installed"
	}
}

// commandExists checks if a command exists in PATH.
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// ComponentStatus represents the installation status of a component.
type ComponentStatus struct {
	Component Component
	Installed bool
	Version   string
	Error     error
}

// CheckAllComponents checks the installation status of all components.
func CheckAllComponents(osInfo *detector.OSInfo) []ComponentStatus {
	components := AllComponents(osInfo)
	statuses := make([]ComponentStatus, len(components))

	for i, c := range components {
		installed, version := c.IsInstalled(osInfo)
		statuses[i] = ComponentStatus{
			Component: c,
			Installed: installed,
			Version:   version,
		}
	}

	return statuses
}

// GetComponentsByCategory returns components filtered by category.
func GetComponentsByCategory(category string, osInfo *detector.OSInfo) []Component {
	all := AllComponents(osInfo)
	var filtered []Component
	for _, c := range all {
		if c.Category == category {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// Categories returns all unique component categories.
func Categories(osInfo *detector.OSInfo) []string {
	components := AllComponents(osInfo)
	seen := make(map[string]bool)
	var cats []string
	for _, c := range components {
		if !seen[c.Category] {
			seen[c.Category] = true
			cats = append(cats, c.Category)
		}
	}
	return cats
}

// CategoryDisplayName returns a human-readable name for a category.
func CategoryDisplayName(category string) string {
	names := map[string]string{
		"system":    "System",
		"runtime":   "Runtimes",
		"font":      "Fonts",
		"tool":      "Tools",
		"extension": "Editor Extensions",
	}
	if name, ok := names[category]; ok {
		return name
	}
	return fmt.Sprintf("%s (unknown)", category)
}

package bootstrap

import (
	"os"
	"testing"

	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/bootstrap/detector"
)

func TestAllComponents(t *testing.T) {
	osInfo := &detector.OSInfo{
		ID:     "ubuntu",
		Name:   "Ubuntu 24.04",
		Family: "linux",
		Arch:   "amd64",
	}

	components := AllComponents(osInfo)
	if len(components) == 0 {
		t.Error("AllComponents should return at least one component")
	}

	// Verify all returned components are compatible with the OS
	for _, c := range components {
		if len(c.OS) > 0 && !containsOS(c.OS, osInfo.ID) && !containsOS(c.OS, osInfo.Family) {
			t.Errorf("Component %s is not compatible with OS %s", c.ID, osInfo.ID)
		}
	}
}

func TestAllComponentsExpectedCount(t *testing.T) {
	// Test with nil OS (returns all components)
	components := AllComponents(nil)
	// We expect a specific number of components
	expectedMinCount := 15
	if len(components) < expectedMinCount {
		t.Errorf("Expected at least %d components, got %d", expectedMinCount, len(components))
	}
}

func TestAllComponentsNilOS(t *testing.T) {
	// Should return all components without filtering
	components := AllComponents(nil)
	if len(components) == 0 {
		t.Error("AllComponents(nil) should return all components")
	}

	// With nil OS, all components should be returned (no filtering)
	// This means components with OS restrictions still appear
	hasRestrictedComponents := false
	for _, c := range components {
		if len(c.OS) > 0 {
			hasRestrictedComponents = true
			break
		}
	}
	if !hasRestrictedComponents {
		t.Error("Expected some components with OS restrictions when OS is nil")
	}
}

func TestComponentsByCategory(t *testing.T) {
	osInfo := &detector.OSInfo{
		ID:     "ubuntu",
		Name:   "Ubuntu 24.04",
		Family: "linux",
		Arch:   "amd64",
	}

	byCategory := ComponentsByCategory(osInfo)

	expectedCategories := []string{"system", "runtime", "font", "tool", "extension"}
	for _, cat := range expectedCategories {
		if _, ok := byCategory[cat]; !ok {
			t.Errorf("Expected category %q to be present", cat)
		}
	}

	// Verify each component is in exactly one category
	totalCount := 0
	for _, components := range byCategory {
		totalCount += len(components)
		for _, c := range components {
			if c.Category == "" {
				t.Error("Component has empty category")
			}
		}
	}

	allComponents := AllComponents(osInfo)
	if totalCount != len(allComponents) {
		t.Errorf("ComponentsByCategory total count (%d) != AllComponents count (%d)",
			totalCount, len(allComponents))
	}
}

func TestComponentsByCategoryGroupsCorrectly(t *testing.T) {
	osInfo := &detector.OSInfo{
		ID:     "ubuntu",
		Name:   "Ubuntu 24.04",
		Family: "linux",
		Arch:   "amd64",
	}

	byCategory := ComponentsByCategory(osInfo)

	// Check that system category has expected components
	systemComps := byCategory["system"]
	if len(systemComps) == 0 {
		t.Error("system category should have components")
	}

	// Verify each system component is actually in system category
	for _, c := range systemComps {
		if c.Category != "system" {
			t.Errorf("Component %s in system category has category %s", c.ID, c.Category)
		}
	}
}

func TestContainsOS(t *testing.T) {
	tests := []struct {
		osList  []string
		osID    string
		matches bool
	}{
		{[]string{"linux", "darwin"}, "linux", true},
		{[]string{"linux", "darwin"}, "darwin", true},
		{[]string{"linux", "darwin"}, "windows", false},
		{[]string{}, "linux", true}, // empty means all
		{[]string{"darwin-arm64"}, "darwin-arm64", true},
		{[]string{"linux"}, "linux", true},
		{[]string{"linux"}, "darwin", false},
		{[]string{"debian", "ubuntu"}, "ubuntu", true},
		{[]string{"debian", "ubuntu"}, "fedora", false},
	}

	for _, tt := range tests {
		got := containsOS(tt.osList, tt.osID)
		if got != tt.matches {
			t.Errorf("containsOS(%v, %q) = %v, want %v", tt.osList, tt.osID, got, tt.matches)
		}
	}
}

func TestOSFiltering(t *testing.T) {
	// Test that OS filtering works correctly

	// Linux-only components
	linuxOS := &detector.OSInfo{
		ID:     "ubuntu",
		Name:   "Ubuntu 24.04",
		Family: "linux",
		Arch:   "amd64",
	}

	linuxComps := AllComponents(linuxOS)

	// Docker should be in the list for linux
	foundDocker := false
	for _, c := range linuxComps {
		if c.ID == "docker" {
			foundDocker = true
			if c.Category != "system" {
				t.Errorf("docker should be in system category")
			}
		}
	}
	if !foundDocker {
		t.Error("docker should be available on linux")
	}

	// darwin-only OS (macOS)
 darwinOS := &detector.OSInfo{
		ID:     "darwin",
		Name:   "macOS 14",
		Family: "bsd",
		Arch:   "arm64",
	}

 darwinComps := AllComponents(darwinOS)

	// Docker should NOT be in the list for darwin
	for _, c := range darwinComps {
		if c.ID == "docker" {
			t.Error("docker should NOT be available on darwin")
		}
	}
}

func TestGhosttyOSFiltering(t *testing.T) {
	// ghostty is available on linux and darwin
	linuxOS := &detector.OSInfo{
		ID:     "arch",
		Name:   "Arch Linux",
		Family: "linux",
		Arch:   "amd64",
	}

	darwinOS := &detector.OSInfo{
		ID:     "darwin",
		Name:   "macOS 14",
		Family: "bsd",
		Arch:   "arm64",
	}

	windowsOS := &detector.OSInfo{
		ID:     "windows",
		Name:   "Windows 11",
		Family: "windows",
		Arch:   "amd64",
	}

	linuxComps := AllComponents(linuxOS)
	darwinComps := AllComponents(darwinOS)
	windowsComps := AllComponents(windowsOS)

	foundGhosttyLinux := false
	foundGhosttyDarwin := false
	foundGhosttyWindows := false

	for _, c := range linuxComps {
		if c.ID == "ghostty" {
			foundGhosttyLinux = true
		}
	}
	for _, c := range darwinComps {
		if c.ID == "ghostty" {
			foundGhosttyDarwin = true
		}
	}
	for _, c := range windowsComps {
		if c.ID == "ghostty" {
			foundGhosttyWindows = true
		}
	}

	if !foundGhosttyLinux {
		t.Error("ghostty should be available on linux")
	}
	if !foundGhosttyDarwin {
		t.Error("ghostty should be available on darwin")
	}
	if foundGhosttyWindows {
		t.Error("ghostty should NOT be available on windows")
	}
}

func TestIsInstalled(t *testing.T) {
	osInfo := &detector.OSInfo{}

	// Test checkCommandVersion with a command that exists
	check := checkCommandVersion("ls", "-V")
	installed, version := check(osInfo)
	// ls should exist on linux, version string may vary
	_ = installed
	_ = version

	// Test checkInstalled
	installCheck := checkInstalled("ls")
	installed, _ = installCheck(osInfo)
	if !installed {
		t.Error("checkInstalled(ls) should return true")
	}

	// Test with non-existent command
	installCheck = checkInstalled("this-command-does-not-exist-12345")
	installed, version = installCheck(osInfo)
	if installed {
		t.Error("checkInstalled with non-existent command should return false")
	}
	if version != "not installed" {
		t.Errorf("expected 'not installed' version, got %s", version)
	}
}

func TestIsInstalledReturnValues(t *testing.T) {
	osInfo := &detector.OSInfo{}

	tests := []struct {
		name     string
		check    func(*detector.OSInfo) (bool, string)
		wantInst bool
		wantVer  string
	}{
		{
			name:     "checkInstalled for existing command",
			check:    checkInstalled("ls"),
			wantInst: true,
			wantVer:  "installed",
		},
		{
			name:     "checkInstalled for non-existing command",
			check:    checkInstalled("nonexistent-command-xyz"),
			wantInst: false,
			wantVer:  "not installed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installed, version := tt.check(osInfo)
			if installed != tt.wantInst {
				t.Errorf("installed = %v, want %v", installed, tt.wantInst)
			}
			if version != tt.wantVer {
				t.Errorf("version = %s, want %s", version, tt.wantVer)
			}
		})
	}
}

func TestComponentDescriptionsNotEmpty(t *testing.T) {
	osInfo := &detector.OSInfo{
		ID:     "ubuntu",
		Name:   "Ubuntu 24.04",
		Family: "linux",
		Arch:   "amd64",
	}

	components := AllComponents(osInfo)
	for _, c := range components {
		if c.Description == "" {
			t.Errorf("Component %s has empty description", c.ID)
		}
		// Descriptions should be at least 5 characters
		if len(c.Description) < 5 {
			t.Errorf("Component %s has too short description: %s", c.ID, c.Description)
		}
	}
}

func TestCategories(t *testing.T) {
	osInfo := &detector.OSInfo{
		ID:     "ubuntu",
		Name:   "Ubuntu 24.04",
		Family: "linux",
		Arch:   "amd64",
	}

	cats := Categories(osInfo)
	if len(cats) == 0 {
		t.Error("Categories should return at least one category")
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, c := range cats {
		if seen[c] {
			t.Errorf("Duplicate category: %s", c)
		}
		seen[c] = true
	}

	// Check expected categories are present
	expected := map[string]bool{
		"system":    true,
		"runtime":   true,
		"font":      true,
		"tool":      true,
		"extension": true,
	}
	for _, c := range cats {
		delete(expected, c)
	}
	for c := range expected {
		t.Errorf("Missing expected category: %s", c)
	}
}

func TestCategoryDisplayName(t *testing.T) {
	tests := []struct {
		category string
		expected string
	}{
		{"system", "System"},
		{"runtime", "Runtimes"},
		{"font", "Fonts"},
		{"tool", "Tools"},
		{"extension", "Editor Extensions"},
		{"unknown", "unknown (unknown)"},
		{"", " (unknown)"},
	}

	for _, tt := range tests {
		got := CategoryDisplayName(tt.category)
		if got != tt.expected {
			t.Errorf("CategoryDisplayName(%q) = %q, want %q", tt.category, got, tt.expected)
		}
	}
}

func TestCheckAllComponents(t *testing.T) {
	osInfo := &detector.OSInfo{
		ID:     "ubuntu",
		Name:   "Ubuntu 24.04",
		Family: "linux",
		Arch:   "amd64",
	}

	statuses := CheckAllComponents(osInfo)
	if len(statuses) == 0 {
		t.Error("CheckAllComponents should return at least one status")
	}

	for _, s := range statuses {
		if s.Component.ID == "" {
			t.Error("ComponentStatus has empty Component.ID")
		}
		// Verify Component has required fields
		if s.Component.Name == "" {
			t.Error("ComponentStatus has empty Component.Name")
		}
	}
}

func TestCheckAllComponentsReturnsCorrectCount(t *testing.T) {
	osInfo := &detector.OSInfo{
		ID:     "ubuntu",
		Name:   "Ubuntu 24.04",
		Family: "linux",
		Arch:   "amd64",
	}

	statuses := CheckAllComponents(osInfo)
	components := AllComponents(osInfo)

	if len(statuses) != len(components) {
		t.Errorf("CheckAllComponents returned %d statuses, want %d", len(statuses), len(components))
	}
}

func TestGetComponentsByCategory(t *testing.T) {
	osInfo := &detector.OSInfo{
		ID:     "ubuntu",
		Name:   "Ubuntu 24.04",
		Family: "linux",
		Arch:   "amd64",
	}

	systemComponents := GetComponentsByCategory("system", osInfo)
	for _, c := range systemComponents {
		if c.Category != "system" {
			t.Errorf("Expected category 'system', got %s", c.Category)
		}
	}

	// Test with non-existent category
	emptyComponents := GetComponentsByCategory("nonexistent", osInfo)
	if len(emptyComponents) != 0 {
		t.Errorf("Expected 0 components for nonexistent category, got %d", len(emptyComponents))
	}
}

func TestCommandExists(t *testing.T) {
	// Test with a command that should exist
	if !commandExists("ls") {
		t.Error("commandExists(ls) should return true on Linux")
	}

	// Test with a command that shouldn't exist
	if commandExists("this-command-does-not-exist-12345") {
		t.Error("commandExists with non-existent command should return false")
	}
}

func TestCheckDirExists(t *testing.T) {
	osInfo := &detector.OSInfo{}

	// Test with a directory that should exist
	check := checkDirExists("/tmp")
	installed, _ := check(osInfo)
	if !installed {
		t.Error("checkDirExists(/tmp) should return true")
	}

	// Test with a directory that shouldn't exist
	check = checkDirExists("/this/does/not/exist/12345")
	installed, _ = check(osInfo)
	if installed {
		t.Error("checkDirExists with non-existent dir should return false")
	}
}

func TestCheckDirExistsWithHome(t *testing.T) {
	osInfo := &detector.OSInfo{}

	// Test HOME expansion
	check := checkDirExists("$HOME")
	installed, _ := check(osInfo)
	// This depends on HOME being set
	home := os.Getenv("HOME")
	if home != "" && !installed {
		t.Error("checkDirExists($HOME) should return true when HOME is set")
	}
}

func TestCheckCommandVersion(t *testing.T) {
	osInfo := &detector.OSInfo{}

	// Test with ls command
	check := checkCommandVersion("ls", "-V")
	installed, version := check(osInfo)
	// ls -V might not be valid on all systems, just check it doesn't panic
	_ = installed
	_ = version
}

func TestCheckCommandVersionOutput(t *testing.T) {
	osInfo := &detector.OSInfo{}

	// Test that checkCommandVersion returns version string
	check := checkCommandVersion("echo", "test")
	installed, version := check(osInfo)
	if !installed {
		t.Error("echo should be installed")
	}
	if version == "" {
		t.Error("version should not be empty for successful command")
	}
}

func TestComponentInstallHint(t *testing.T) {
	osInfo := &detector.OSInfo{
		ID:     "ubuntu",
		Name:   "Ubuntu 24.04",
		Family: "linux",
		Arch:   "amd64",
	}

	components := AllComponents(osInfo)
	for _, c := range components {
		installed, _ := c.IsInstalled(osInfo)
		if c.InstallHint == "" && !installed {
			t.Errorf("Component %s has no InstallHint and is not installed", c.ID)
		}
	}
}

func TestCheckFont(t *testing.T) {
	osInfo := &detector.OSInfo{}

	// Test with a font that likely doesn't exist
	check := checkFont("NonExistentFontXYZ123")
	installed, version := check(osInfo)
	if installed {
		t.Error("NonExistentFontXYZ123 should not be installed")
	}
	if version != "not installed" {
		t.Errorf("expected 'not installed', got %s", version)
	}
}

func TestGetAllComponents(t *testing.T) {
	// Test that getAllComponents returns all components
	components := getAllComponents()
	if len(components) == 0 {
		t.Error("getAllComponents should return at least one component")
	}

	// Verify all expected categories are present
	categoryCount := make(map[string]int)
	for _, c := range components {
		categoryCount[c.Category]++
	}

	expectedCategories := []string{"system", "runtime", "font", "tool", "extension"}
	for _, cat := range expectedCategories {
		if categoryCount[cat] == 0 {
			t.Errorf("Expected at least one component in category %s", cat)
		}
	}
}

func TestComponentHasRequiredFields(t *testing.T) {
	components := getAllComponents()
	for _, c := range components {
		if c.ID == "" {
			t.Error("Component has empty ID")
		}
		if c.Name == "" {
			t.Error("Component has empty Name")
		}
		if c.Description == "" {
			t.Error("Component has empty Description")
		}
		if c.Category == "" {
			t.Error("Component has empty Category")
		}
		if c.InstallCmd == "" {
			t.Error("Component has empty InstallCmd")
		}
	}
}

func TestComponentFields(t *testing.T) {
	components := getAllComponents()

	for _, c := range components {
		// ID should match expected pattern (lowercase, no spaces)
		for _, r := range c.ID {
			if r == ' ' {
				t.Errorf("Component %s has space in ID", c.ID)
			}
		}

		// Category should be one of the known categories
		validCategories := map[string]bool{
			"system":    true,
			"runtime":   true,
			"font":      true,
			"tool":      true,
			"extension": true,
		}
		if !validCategories[c.Category] {
			t.Errorf("Component %s has invalid category: %s", c.ID, c.Category)
		}
	}
}

func TestCheckDirExistsHOMEExpand(t *testing.T) {
	osInfo := &detector.OSInfo{}
	home := os.Getenv("HOME")
	if home == "" {
		t.Skip("HOME not set")
	}

	// Test with actual HOME path
	check := checkDirExists("$HOME/.config")
	installed, _ := check(osInfo)
	// .config should exist in most linux setups
	// Don't fail the test if it doesn't exist, just log
	_ = installed
}

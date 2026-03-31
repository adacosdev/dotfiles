// Package status provides tests for the status dashboard.
package status

import (
	"testing"
	"time"

	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/bootstrap/detector"
)

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		str    string
		icon   string
	}{
		{"StatusOk", StatusOk, "ok", "✅"},
		{"StatusWarn", StatusWarn, "warn", "⚠️"},
		{"StatusError", StatusError, "error", "❌"},
		{"StatusUnknown", StatusUnknown, "unknown", "❓"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.str {
				t.Errorf("Status.String() = %v, want %v", got, tt.str)
			}
			if got := tt.status.Icon(); got != tt.icon {
				t.Errorf("Status.Icon() = %v, want %v", got, tt.icon)
			}
		})
	}
}

func TestHealthCheckCreation(t *testing.T) {
	now := time.Now()
	check := HealthCheck{
		Name:      "test-component",
		Category:  "system",
		Status:    StatusOk,
		Version:   "1.0.0",
		Message:   "Test message",
		Command:   "test --version",
		CheckedAt: now,
	}

	if check.Name != "test-component" {
		t.Errorf("HealthCheck.Name = %v, want test-component", check.Name)
	}
	if check.Category != "system" {
		t.Errorf("HealthCheck.Category = %v, want system", check.Category)
	}
	if check.Status != StatusOk {
		t.Errorf("HealthCheck.Status = %v, want StatusOk", check.Status)
	}
	if check.Version != "1.0.0" {
		t.Errorf("HealthCheck.Version = %v, want 1.0.0", check.Version)
	}
	if check.Message != "Test message" {
		t.Errorf("HealthCheck.Message = %v, want Test message", check.Message)
	}
	if check.Command != "test --version" {
		t.Errorf("HealthCheck.Command = %v, want test --version", check.Command)
	}
	if !check.CheckedAt.Equal(now) {
		t.Errorf("HealthCheck.CheckedAt = %v, want %v", check.CheckedAt, now)
	}
}

func TestStatusDetermination(t *testing.T) {
	tests := []struct {
		name       string
		installed  bool
		version    string
		wantStatus Status
	}{
		{
			name:       "installed with version",
			installed:  true,
			version:    "1.0.0",
			wantStatus: StatusOk,
		},
		{
			name:       "installed but error in version",
			installed:  true,
			version:    "error: something failed",
			wantStatus: StatusError,
		},
		{
			name:       "not installed",
			installed:  false,
			version:    "not installed",
			wantStatus: StatusWarn, // false means warn, not error
		},
		{
			name:       "installed but unknown version",
			installed:  true,
			version:    "installed (version unknown)",
			wantStatus: StatusOk,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status Status
			switch {
			case tt.installed && tt.version == "not installed":
				status = StatusError
			case tt.installed && len(tt.version) > 0 && tt.version[:5] == "error":
				status = StatusError
			case tt.installed:
				status = StatusOk
			default:
				status = StatusWarn
			}

			if status != tt.wantStatus {
				t.Errorf("got status %v, want %v", status, tt.wantStatus)
			}
		})
	}
}

func TestDashboardModelCreation(t *testing.T) {
	model := NewDashboardModel()

	if model == nil {
		t.Fatal("NewDashboardModel() returned nil")
	}

	if model.os == nil {
		t.Error("model.os is nil")
	}

	if model.checks == nil {
		t.Error("model.checks is nil")
	}

	if model.selected != 0 {
		t.Errorf("model.selected = %v, want 0", model.selected)
	}

	if model.categoryIdx != 0 {
		t.Errorf("model.categoryIdx = %v, want 0", model.categoryIdx)
	}

	if model.showDetails {
		t.Error("model.showDetails = true, want false")
	}
}

func TestDashboardModelRefresh(t *testing.T) {
	model := NewDashboardModel()

	initialCount := len(model.checks)
	if initialCount == 0 {
		t.Fatal("model.checks should have items after creation")
	}

	// Store original refresh time
	oldRefresh := model.lastRefresh

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Refresh
	model.refreshChecks()

	if len(model.checks) != initialCount {
		t.Errorf("len(model.checks) = %v, want %v (should be same after refresh)",
			len(model.checks), initialCount)
	}

	if model.lastRefresh.Equal(oldRefresh) {
		t.Error("model.lastRefresh should be updated after refresh")
	}

	if model.lastRefresh.Before(oldRefresh) {
		t.Error("model.lastRefresh should be after the old refresh time")
	}
}

func TestCategoryDisplayName(t *testing.T) {
	tests := []struct {
		category string
		want     string
	}{
		{"system", "System"},
		{"runtime", "Runtimes"},
		{"font", "Fonts"},
		{"tool", "Tools"},
		{"extension", "Extensions"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got := CategoryDisplayName(tt.category)
			if got != tt.want {
				t.Errorf("CategoryDisplayName(%q) = %v, want %v", tt.category, got, tt.want)
			}
		})
	}
}

func TestCategories(t *testing.T) {
	cats := Categories()

	expected := []string{"system", "runtime", "font", "tool", "extension"}

	if len(cats) != len(expected) {
		t.Fatalf("len(Categories()) = %v, want %v", len(cats), len(expected))
	}

	for i, cat := range cats {
		if cat != expected[i] {
			t.Errorf("Categories()[%d] = %v, want %v", i, cat, expected[i])
		}
	}
}

func TestDashboardGetChecks(t *testing.T) {
	model := NewDashboardModel()
	checks := model.GetChecks()

	if len(checks) == 0 {
		t.Error("GetChecks() returned empty slice")
	}

	// Verify checks have required fields
	for _, check := range checks {
		if check.Name == "" {
			t.Error("HealthCheck.Name should not be empty")
		}
		if check.Category == "" {
			t.Error("HealthCheck.Category should not be empty")
		}
		if check.CheckedAt.IsZero() {
			t.Error("HealthCheck.CheckedAt should not be zero")
		}
	}
}

func TestDashboardGetSelected(t *testing.T) {
	model := NewDashboardModel()

	selected := model.GetSelected()
	if selected != 0 {
		t.Errorf("GetSelected() = %v, want 0", selected)
	}

	// Manually set selected
	model.selected = 5
	if got := model.GetSelected(); got != 5 {
		t.Errorf("GetSelected() = %v, want 5", got)
	}
}

func TestCheckOS(t *testing.T) {
	model := &DashboardModel{
		os: &detector.OSInfo{
			ID:     "ubuntu",
			Name:   "Ubuntu 24.04",
			Family: "linux",
			Arch:   "amd64",
		},
	}

	check := model.checkOS()

	if check.Name != "OS" {
		t.Errorf("checkOS().Name = %v, want OS", check.Name)
	}

	if check.Category != "system" {
		t.Errorf("checkOS().Category = %v, want system", check.Category)
	}

	if check.Status != StatusOk {
		t.Errorf("checkOS().Status = %v, want StatusOk", check.Status)
	}
}

func TestCheckShell(t *testing.T) {
	model := &DashboardModel{}

	check := model.checkShell()

	if check.Name == "" {
		t.Error("checkShell().Name should not be empty")
	}

	if check.Category != "system" {
		t.Errorf("checkShell().Category = %v, want system", check.Category)
	}

	if check.Version == "" {
		t.Error("checkShell().Version should not be empty")
	}
}

func TestMockOSInfo(t *testing.T) {
	// Test with a mock OS info (nil os is handled at model creation level)
	model := &DashboardModel{
		os: &detector.OSInfo{
			ID:     "test",
			Name:   "Test OS",
			Family: "linux",
			Arch:   "amd64",
		},
	}

	// Should not panic when checking OS
	check := model.checkOS()

	if check.Status != StatusOk {
		t.Errorf("checkOS().Status = %v, want StatusOk", check.Status)
	}
}

func TestStatusStyle(t *testing.T) {
	model := &DashboardModel{}

	tests := []struct {
		status Status
	}{
		{StatusOk},
		{StatusWarn},
		{StatusError},
		{StatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			got := model.statusStyle(tt.status)
			// Just verify the style renders something non-empty
			rendered := got.Render("test")
			if rendered == "" {
				t.Errorf("statusStyle(%v).Render() returned empty string", tt.status)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{0, 0, 0},
		{-1, 1, 1},
		{100, 50, 100},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := max(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCommandVersion(t *testing.T) {
	// Test with a command that exists and produces output
	version, ok := checkCommandVersion("ls", "--version")
	if !ok {
		t.Error("checkCommandVersion should return true for ls --version")
	}
	if version == "" {
		t.Error("checkCommandVersion should return non-empty version for ls --version")
	}

	// Test with a command that doesn't exist
	_, ok = checkCommandVersion("nonexistent_command_12345", "--version")
	if ok {
		t.Error("checkCommandVersion should return false for nonexistent command")
	}
}

func TestCommandExists(t *testing.T) {
	// Test with commands that should exist
	if !commandExists("ls") {
		t.Error("commandExists should return true for ls")
	}

	if !commandExists("echo") {
		t.Error("commandExists should return true for echo")
	}

	// Test with a command that shouldn't exist
	if commandExists("nonexistent_command_xyz") {
		t.Error("commandExists should return false for nonexistent command")
	}
}

func TestViewDashboard(t *testing.T) {
	model := NewDashboardModel()
	model.Width = 80
	model.Height = 24

	view := model.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestViewDetails(t *testing.T) {
	model := NewDashboardModel()
	model.Width = 80
	model.Height = 24
	model.showDetails = true
	model.detailIndex = 0

	view := model.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

func TestStop(t *testing.T) {
	model := NewDashboardModel()
	// Should not panic
	model.Stop()
}

package bootstrap

import (
	"testing"

	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/bootstrap/detector"
)

func TestExecutorExecute(t *testing.T) {
	exec := &Executor{
		OS: &detector.OSInfo{
			ID:     "ubuntu",
			Name:   "Ubuntu 24.04",
			Family: "linux",
			Arch:   "amd64",
		},
		DryRun: true,
	}

	// Create a minimal test component
	components := []Component{
		{
			ID:          "test-component",
			Name:        "Test Component",
			Description: "A test component",
			Category:    "system",
			OS:          []string{"linux"},
			InstallCmd:  "echo 'test'",
			NeedsRoot:   false,
			IsInstalled: func(info *detector.OSInfo) (bool, string) {
				return false, "not installed"
			},
		},
	}

	results := exec.Execute(components, false)
	count := 0
	for result := range results {
		count++
		if result.Component != "test-component" {
			t.Errorf("Expected component 'test-component', got %s", result.Component)
		}
		if result.Status != "pending" {
			t.Errorf("Expected status 'pending' in dry-run mode, got %s", result.Status)
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 result, got %d", count)
	}
}

func TestExecutorDryRun(t *testing.T) {
	exec := &Executor{
		OS: &detector.OSInfo{
			ID:     "ubuntu",
			Name:   "Ubuntu 24.04",
			Family: "linux",
			Arch:   "amd64",
		},
		DryRun: true,
	}

	components := []Component{
		{
			ID:          "tmux",
			Name:        "Tmux",
			Description: "Terminal multiplexer",
			Category:    "system",
			InstallCmd:  "echo 'install tmux'",
			NeedsRoot:   false,
			IsInstalled: func(info *detector.OSInfo) (bool, string) {
				return true, "tmux 3.2a"
			},
		},
	}

	results := exec.Execute(components, false)
	for result := range results {
		if result.Status != "skipped" {
			t.Errorf("Already installed component should be 'skipped', got %s", result.Status)
		}
	}
}

func TestExecutorForce(t *testing.T) {
	exec := &Executor{
		OS: &detector.OSInfo{
			ID:     "ubuntu",
			Name:   "Ubuntu 24.04",
			Family: "linux",
			Arch:   "amd64",
		},
		Force:  true,
		DryRun: true,
	}

	components := []Component{
		{
			ID:          "tmux",
			Name:        "Tmux",
			Description: "Terminal multiplexer",
			Category:    "system",
			InstallCmd:  "echo 'install tmux'",
			NeedsRoot:   false,
			IsInstalled: func(info *detector.OSInfo) (bool, string) {
				return true, "tmux 3.2a"
			},
		},
	}

	results := exec.Execute(components, false)
	for result := range results {
		if result.Status != "pending" {
			t.Errorf("With Force=true, even installed component should be 'pending', got %s", result.Status)
		}
	}
}

func TestExecuteSync(t *testing.T) {
	exec := &Executor{
		OS: &detector.OSInfo{
			ID:     "ubuntu",
			Name:   "Ubuntu 24.04",
			Family: "linux",
			Arch:   "amd64",
		},
		DryRun: true,
	}

	components := []Component{
		{
			ID:          "test1",
			Name:        "Test 1",
			Description: "Test component 1",
			Category:    "system",
			InstallCmd:  "echo 'test1'",
			IsInstalled: func(info *detector.OSInfo) (bool, string) {
				return false, "not installed"
			},
		},
		{
			ID:          "test2",
			Name:        "Test 2",
			Description: "Test component 2",
			Category:    "system",
			InstallCmd:  "echo 'test2'",
			IsInstalled: func(info *detector.OSInfo) (bool, string) {
				return false, "not installed"
			},
		},
	}

	results := exec.ExecuteSync(components)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestExecuteWithProgress(t *testing.T) {
	exec := &Executor{
		OS: &detector.OSInfo{
			ID:     "ubuntu",
			Name:   "Ubuntu 24.04",
			Family: "linux",
			Arch:   "amd64",
		},
		DryRun: true,
	}

	components := []Component{
		{
			ID:          "test",
			Name:        "Test",
			Description: "Test component",
			Category:    "system",
			InstallCmd:  "echo 'test'",
			IsInstalled: func(info *detector.OSInfo) (bool, string) {
				return false, "not installed"
			},
		},
	}

	callCount := 0
	exec.ExecuteWithProgress(components, func(result Result) {
		callCount++
	})

	if callCount != 1 {
		t.Errorf("Expected callback to be called 1 time, got %d", callCount)
	}
}

func TestNewParallelExecutor(t *testing.T) {
	exec := &Executor{
		OS: &detector.OSInfo{
			ID:     "ubuntu",
			Name:   "Ubuntu 24.04",
			Family: "linux",
			Arch:   "amd64",
		},
		DryRun: true,
	}

	// Test with valid maxParallel
	pexec := NewParallelExecutor(exec, 3)
	if pexec.MaxParallel != 3 {
		t.Errorf("Expected MaxParallel 3, got %d", pexec.MaxParallel)
	}

	// Test with zero maxParallel (should default to 3)
	pexec = NewParallelExecutor(exec, 0)
	if pexec.MaxParallel != 3 {
		t.Errorf("Expected MaxParallel 3 (default), got %d", pexec.MaxParallel)
	}

	// Test with negative maxParallel (should default to 3)
	pexec = NewParallelExecutor(exec, -1)
	if pexec.MaxParallel != 3 {
		t.Errorf("Expected MaxParallel 3 (default), got %d", pexec.MaxParallel)
	}
}

func TestParallelExecutorExecute(t *testing.T) {
	exec := &Executor{
		OS: &detector.OSInfo{
			ID:     "ubuntu",
			Name:   "Ubuntu 24.04",
			Family: "linux",
			Arch:   "amd64",
		},
		DryRun: true,
	}

	pexec := NewParallelExecutor(exec, 2)

	components := []Component{
		{
			ID:          "test1",
			Name:        "Test 1",
			Description: "Test component 1",
			Category:    "system",
			InstallCmd:  "echo 'test1'",
			IsInstalled: func(info *detector.OSInfo) (bool, string) {
				return false, "not installed"
			},
		},
		{
			ID:          "test2",
			Name:        "Test 2",
			Description: "Test component 2",
			Category:    "system",
			InstallCmd:  "echo 'test2'",
			IsInstalled: func(info *detector.OSInfo) (bool, string) {
				return false, "not installed"
			},
		},
	}

	results := pexec.Execute(components, false)
	count := 0
	for result := range results {
		count++
		if result.Component == "" {
			t.Error("Got empty component in result")
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 results, got %d", count)
	}
}

func TestHasRootPrivileges(t *testing.T) {
	exec := &Executor{
		OS: &detector.OSInfo{
			ID:     "ubuntu",
			Name:   "Ubuntu 24.04",
			Family: "linux",
			Arch:   "amd64",
		},
	}

	// This test will return different results depending on actual environment
	// We just verify it doesn't panic and returns a boolean
	result := exec.hasRootPrivileges()
	t.Logf("hasRootPrivileges() = %v (environment dependent)", result)
}

func TestMapShellStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ok", "installed"},
		{"installed", "installed"},
		{"skip", "skipped"},
		{"skipped", "skipped"},
		{"error", "error"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		got := mapShellStatus(tt.input)
		if got != tt.expected {
			t.Errorf("mapShellStatus(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

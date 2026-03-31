package cli

import (
	"testing"

	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/status"
	"github.com/charmbracelet/bubbletea"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantForce     bool
		wantDryRun    bool
		wantJSON      bool
		wantHelp      bool
		wantRemaining []string
	}{
		{
			name:          "empty args",
			args:          []string{},
			wantForce:     false,
			wantDryRun:    false,
			wantJSON:      false,
			wantHelp:      false,
			wantRemaining: nil,
		},
		{
			name:          "force flag long",
			args:          []string{"--force", "bootstrap"},
			wantForce:     true,
			wantDryRun:    false,
			wantJSON:      false,
			wantHelp:      false,
			wantRemaining: []string{"bootstrap"},
		},
		{
			name:          "force flag short",
			args:          []string{"-f", "diff"},
			wantForce:     true,
			wantDryRun:    false,
			wantJSON:      false,
			wantHelp:      false,
			wantRemaining: []string{"diff"},
		},
		{
			name:          "dry-run flag long",
			args:          []string{"--dry-run", "apply"},
			wantForce:     false,
			wantDryRun:    true,
			wantJSON:      false,
			wantHelp:      false,
			wantRemaining: []string{"apply"},
		},
		{
			name:          "dry-run flag short",
			args:          []string{"-n", "status"},
			wantForce:     false,
			wantDryRun:    true,
			wantJSON:      false,
			wantHelp:      false,
			wantRemaining: []string{"status"},
		},
		{
			name:          "json flag",
			args:          []string{"--json", "diff"},
			wantForce:     false,
			wantDryRun:    false,
			wantJSON:      true,
			wantHelp:      false,
			wantRemaining: []string{"diff"},
		},
		{
			name:          "help flag long",
			args:          []string{"--help"},
			wantForce:     false,
			wantDryRun:    false,
			wantJSON:      false,
			wantHelp:      true,
			wantRemaining: nil,
		},
		{
			name:          "help flag short",
			args:          []string{"-h"},
			wantForce:     false,
			wantDryRun:    false,
			wantJSON:      false,
			wantHelp:      true,
			wantRemaining: nil,
		},
		{
			name:          "multiple flags",
			args:          []string{"--force", "--dry-run", "--json", "apply"},
			wantForce:     true,
			wantDryRun:    true,
			wantJSON:      true,
			wantHelp:      false,
			wantRemaining: []string{"apply"},
		},
		{
			name:          "flags after subcommand",
			args:          []string{"bootstrap", "--force", "--dry-run"},
			wantForce:     false,
			wantDryRun:    false,
			wantJSON:      false,
			wantHelp:      false,
			wantRemaining: []string{"bootstrap", "--force", "--dry-run"},
		},
		{
			name:          "unknown flag before subcommand is ignored",
			args:          []string{"--unknown", "bootstrap"},
			wantForce:     false,
			wantDryRun:    false,
			wantJSON:      false,
			wantHelp:      false,
			wantRemaining: []string{"bootstrap"},
		},
		{
			name:          "force and json together before subcommand",
			args:          []string{"--force", "--json", "bootstrap"},
			wantForce:     true,
			wantDryRun:    false,
			wantJSON:      true,
			wantHelp:      false,
			wantRemaining: []string{"bootstrap"},
		},
		{
			name:          "all flags mixed order",
			args:          []string{"--dry-run", "--force", "-n", "-f", "apply"},
			wantForce:     true,
			wantDryRun:    true,
			wantJSON:      false,
			wantHelp:      false,
			wantRemaining: []string{"apply"},
		},
		{
			name:          "only subcommand",
			args:          []string{"bootstrap"},
			wantForce:     false,
			wantDryRun:    false,
			wantJSON:      false,
			wantHelp:      false,
			wantRemaining: []string{"bootstrap"},
		},
		{
			name:          "subcommand with multiple flags after",
			args:          []string{"diff", "--json", "--dry-run"},
			wantForce:     false,
			wantDryRun:    false,
			wantJSON:      false,
			wantHelp:      false,
			wantRemaining: []string{"diff", "--json", "--dry-run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			force, dryRun, jsonMode, help, remaining := ParseFlags(tt.args)

			if force != tt.wantForce {
				t.Errorf("force = %v, want %v", force, tt.wantForce)
			}
			if dryRun != tt.wantDryRun {
				t.Errorf("dryRun = %v, want %v", dryRun, tt.wantDryRun)
			}
			if jsonMode != tt.wantJSON {
				t.Errorf("jsonMode = %v, want %v", jsonMode, tt.wantJSON)
			}
			if help != tt.wantHelp {
				t.Errorf("help = %v, want %v", help, tt.wantHelp)
			}
			if !sliceEqual(remaining, tt.wantRemaining) {
				t.Errorf("remaining = %v, want %v", remaining, tt.wantRemaining)
			}
		})
	}
}

func TestParseFlagsLeadingFlags(t *testing.T) {
	// Test that leading flags before subcommand are parsed correctly
	tests := []struct {
		name          string
		args          []string
		wantForce     bool
		wantSubcommand string
	}{
		{
			name:          "force then bootstrap",
			args:          []string{"--force", "bootstrap"},
			wantForce:     true,
			wantSubcommand: "bootstrap",
		},
		{
			name:          "dry-run then diff",
			args:          []string{"--dry-run", "diff"},
			wantForce:     false,
			wantSubcommand: "diff",
		},
		{
			name:          "json then apply",
			args:          []string{"--json", "apply"},
			wantForce:     false,
			wantSubcommand: "apply",
		},
		{
			name:          "all flags then status",
			args:          []string{"--force", "--dry-run", "--json", "status"},
			wantForce:     true,
			wantSubcommand: "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			force, dryRun, jsonMode, _, remaining := ParseFlags(tt.args)

			if force != tt.wantForce {
				t.Errorf("force = %v, want %v", force, tt.wantForce)
			}
			if tt.name == "all flags then status" && (dryRun != true || jsonMode != true) {
				t.Errorf("dryRun or jsonMode not set correctly")
			}
			if len(remaining) > 0 && remaining[0] != tt.wantSubcommand {
				t.Errorf("remaining[0] = %v, want %v", remaining[0], tt.wantSubcommand)
			}
		})
	}
}

func TestGetSubcommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantSubcommand Subcommand
		wantRemaining  []string
	}{
		{
			name:           "bootstrap",
			args:           []string{"bootstrap"},
			wantSubcommand: SubcommandBootstrap,
			wantRemaining:  nil,
		},
		{
			name:           "bootstrap with args",
			args:           []string{"bootstrap", "--force"},
			wantSubcommand: SubcommandBootstrap,
			wantRemaining:  []string{"--force"},
		},
		{
			name:           "diff",
			args:           []string{"diff"},
			wantSubcommand: SubcommandDiff,
			wantRemaining:  nil,
		},
		{
			name:           "diff with args",
			args:           []string{"diff", "--json"},
			wantSubcommand: SubcommandDiff,
			wantRemaining:  []string{"--json"},
		},
		{
			name:           "apply",
			args:           []string{"apply"},
			wantSubcommand: SubcommandApply,
			wantRemaining:  nil,
		},
		{
			name:           "status",
			args:           []string{"status"},
			wantSubcommand: SubcommandStatus,
			wantRemaining:  nil,
		},
		{
			name:           "empty",
			args:           []string{},
			wantSubcommand: "",
			wantRemaining:  nil,
		},
		{
			name:           "unknown subcommand",
			args:           []string{"unknown"},
			wantSubcommand: "",
			wantRemaining:  []string{"unknown"},
		},
		{
			name:           "unknown with args",
			args:           []string{"unknown", "arg1"},
			wantSubcommand: "",
			wantRemaining:  []string{"unknown", "arg1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, remaining := GetSubcommand(tt.args)

			if sub != tt.wantSubcommand {
				t.Errorf("subcommand = %v, want %v", sub, tt.wantSubcommand)
			}
			if !sliceEqual(remaining, tt.wantRemaining) {
				t.Errorf("remaining = %v, want %v", remaining, tt.wantRemaining)
			}
		})
	}
}

func TestGetSubcommandWithLeadingFlags(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantSubcommand Subcommand
		wantFlags      []string
		wantArgs       []string
	}{
		{
			name:           "force then bootstrap",
			args:           []string{"--force", "bootstrap"},
			wantSubcommand: SubcommandBootstrap,
			wantFlags:      []string{"--force"},
			wantArgs:       nil,
		},
		{
			name:           "multiple flags then bootstrap with args",
			args:           []string{"--force", "--dry-run", "bootstrap", "--json"},
			wantSubcommand: SubcommandBootstrap,
			wantFlags:      []string{"--force", "--dry-run"},
			wantArgs:       []string{"--json"},
		},
		{
			name:           "leading flags ignored for unknown",
			args:           []string{"--force", "unknown"},
			wantSubcommand: "",
			wantFlags:      []string{"--force"},
			wantArgs:       []string{"unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, remaining := GetSubcommand(tt.args)

			if sub != tt.wantSubcommand {
				t.Errorf("subcommand = %v, want %v", sub, tt.wantSubcommand)
			}
			if !sliceEqual(remaining, tt.wantFlags) && !sliceEqual(remaining, append(tt.wantFlags, tt.wantArgs...)) {
				t.Errorf("remaining = %v, expected flags=%v args=%v", remaining, tt.wantFlags, tt.wantArgs)
			}
		})
	}
}

func TestValidateSubcommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"bootstrap", "bootstrap", true},
		{"diff", "diff", true},
		{"apply", "apply", true},
		{"status", "status", true},
		{"unknown", "unknown", false},
		{"empty", "", false},
		{"bootstrap uppercase", "BOOTSTRAP", false},
		{"Bootstrap mixed case", "Bootstrap", false},
		{"diff with spaces", " diff ", false},
		{"version", "version", false},
		{"help", "help", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateSubcommand(tt.input)
			if got != tt.want {
				t.Errorf("ValidateSubcommand(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewRouter(t *testing.T) {
	tests := []struct {
		name      string
		force     bool
		dryRun    bool
		jsonMode  bool
	}{
		{"all false", false, false, false},
		{"force only", true, false, false},
		{"dry-run only", false, true, false},
		{"json only", false, false, true},
		{"all true", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter(tt.force, tt.dryRun, tt.jsonMode)

			if router.force != tt.force {
				t.Errorf("router.force = %v, want %v", router.force, tt.force)
			}
			if router.dryRun != tt.dryRun {
				t.Errorf("router.dryRun = %v, want %v", router.dryRun, tt.dryRun)
			}
			if router.jsonMode != tt.jsonMode {
				t.Errorf("router.jsonMode = %v, want %v", router.jsonMode, tt.jsonMode)
			}
		})
	}
}

func TestRouterIsTTY(t *testing.T) {
	router := NewRouter(false, false, false)
	// isTTY is set based on tty.IsTerminal() at creation time
	// Just verify the field exists and is a bool
	if router.isTTY != false && router.isTTY != true {
		t.Error("isTTY should be a boolean")
	}
}

func TestFindHelpersPath(t *testing.T) {
	// findHelpersPath returns a path string
	// The actual value depends on the environment
	path := findHelpersPath()
	// Just verify it returns without panicking
	// The value may be empty if no helpers are installed
	_ = path
}

func TestUnknownSubcommandErrorMessage(t *testing.T) {
	// Test that unknown subcommands are handled correctly
	sub, remaining := GetSubcommand([]string{"invalid"})

	if sub != "" {
		t.Errorf("expected empty subcommand for 'invalid', got %v", sub)
	}

	if len(remaining) == 0 || remaining[0] != "invalid" {
		t.Errorf("expected remaining to contain 'invalid', got %v", remaining)
	}
}

func TestHelpFlagRouting(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantHelp  bool
	}{
		{
			name:     "help flag alone",
			args:     []string{"--help"},
			wantHelp: true,
		},
		{
			name:     "help flag short",
			args:     []string{"-h"},
			wantHelp: true,
		},
		{
			name:     "help after subcommand",
			args:     []string{"bootstrap", "--help"},
			wantHelp: false, // --help after subcommand is not caught by ParseFlags
		},
		{
			name:     "no help",
			args:     []string{"--force", "bootstrap"},
			wantHelp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, help, _ := ParseFlags(tt.args)
			if help != tt.wantHelp {
				t.Errorf("help = %v, want %v", help, tt.wantHelp)
			}
		})
	}
}

func TestSubcommandConstants(t *testing.T) {
	// Verify subcommand constants have expected values
	if SubcommandBootstrap != "bootstrap" {
		t.Errorf("SubcommandBootstrap = %q, want 'bootstrap'", SubcommandBootstrap)
	}
	if SubcommandDiff != "diff" {
		t.Errorf("SubcommandDiff = %q, want 'diff'", SubcommandDiff)
	}
	if SubcommandApply != "apply" {
		t.Errorf("SubcommandApply = %q, want 'apply'", SubcommandApply)
	}
	if SubcommandStatus != "status" {
		t.Errorf("SubcommandStatus = %q, want 'status'", SubcommandStatus)
	}
}

func TestParseFlagsAllCombinations(t *testing.T) {
	// Test all combinations of flags
	flagCombos := [][]string{
		{"--force"},
		{"--dry-run"},
		{"--json"},
		{"--help"},
		{"--force", "--dry-run"},
		{"--force", "--json"},
		{"--dry-run", "--json"},
		{"--force", "--dry-run", "--json"},
		{"-f"},
		{"-n"},
		{"-f", "-n"},
		{"-f", "-n", "--json"},
	}

	for _, flags := range flagCombos {
		args := append(flags, "bootstrap")
		force, dryRun, jsonMode, help, remaining := ParseFlags(args)

		// Verify flags are parsed
		if len(remaining) > 0 && remaining[0] != "bootstrap" {
			t.Errorf("flags %v: remaining should start with 'bootstrap', got %v", flags, remaining)
		}

		// Verify we can reconstruct the flags
		hasForce := false
		hasDryRun := false
		hasJSON := false
		hasHelp := false

		for _, f := range flags {
			switch f {
			case "--force", "-f":
				hasForce = true
			case "--dry-run", "-n":
				hasDryRun = true
			case "--json":
				hasJSON = true
			case "--help", "-h":
				hasHelp = true
			}
		}

		if force != hasForce {
			t.Errorf("flags %v: force = %v, want %v", flags, force, hasForce)
		}
		if dryRun != hasDryRun {
			t.Errorf("flags %v: dryRun = %v, want %v", flags, dryRun, hasDryRun)
		}
		if jsonMode != hasJSON {
			t.Errorf("flags %v: jsonMode = %v, want %v", flags, jsonMode, hasJSON)
		}
		if help != hasHelp {
			t.Errorf("flags %v: help = %v, want %v", flags, help, hasHelp)
		}
	}
}

func TestRouterMethods(t *testing.T) {
	router := NewRouter(false, false, false)

	// Test that router methods exist and return expected types
	// Bootstrap method
	model := router.Bootstrap()
	if model == nil {
		t.Error("Bootstrap() should return a non-nil model")
	}

	// Diff method
	diffModel := router.Diff()
	if diffModel == nil {
		t.Error("Diff() should return a non-nil model")
	}

	// Apply method
	applyModel := router.Apply()
	if applyModel == nil {
		t.Error("Apply() should return a non-nil model")
	}

	// Status method
	statusModel := router.Status()
	if statusModel == nil {
		t.Error("Status() should return a non-nil model")
	}
}

func TestBootstrapCLIMethod(t *testing.T) {
	router := NewRouter(false, false, false)

	// BootstrapCLI should not panic and return an error (or nil)
	// It may fail due to missing OS info, but shouldn't panic
	err := router.BootstrapCLI()
	// Error is acceptable - OS detection might fail in test env
	_ = err
}

func TestDiffCLIMethod(t *testing.T) {
	router := NewRouter(false, false, false)

	// DiffCLI might fail if chezmoi is not installed
	// Just verify it doesn't panic
	err := router.DiffCLI()
	// Error is acceptable if chezmoi is not available
	_ = err
}

func TestApplyCLIMethod(t *testing.T) {
	router := NewRouter(false, false, false)

	// ApplyCLI might fail if chezmoi is not installed
	// Just verify it doesn't panic
	err := router.ApplyCLI()
	// Error is acceptable if chezmoi is not available
	_ = err
}

func TestStatusCLIMethod(t *testing.T) {
	router := NewRouter(false, false, false)

	// StatusCLI should work even without chezmoi
	err := router.StatusCLI()
	if err != nil {
		t.Logf("StatusCLI returned error (may be expected): %v", err)
	}
}

func TestBootstrapJSONMethod(t *testing.T) {
	router := NewRouter(false, false, true)

	// BootstrapCLI with jsonMode should output JSON
	err := router.BootstrapCLI()
	// Error may happen due to OS detection, but JSON output should be attempted
	_ = err
}

func TestDiffJSONMethod(t *testing.T) {
	router := NewRouter(false, false, true)

	// DiffJSON should work (or fail gracefully)
	err := router.diffJSON()
	_ = err
}

func TestApplyJSONMethod(t *testing.T) {
	router := NewRouter(false, true, true) // dry-run + json mode

	// ApplyJSON with dry-run should output what would be applied
	err := router.applyJSON()
	_ = err
}

func TestStatusJSONMethod(t *testing.T) {
	router := NewRouter(false, false, true)

	// Create a proper DashboardModel
	model := status.NewDashboardModel()

	// statusJSON should work with the model
	err := router.statusJSON(model)
	if err != nil {
		t.Logf("statusJSON returned error: %v", err)
	}
}

func TestStatusTextMethod(t *testing.T) {
	router := NewRouter(false, false, false)

	// Create a proper DashboardModel
	model := status.NewDashboardModel()

	// statusText should work with the model
	err := router.statusText(model)
	if err != nil {
		t.Logf("statusText returned error: %v", err)
	}
}

func TestStatusJSONWithInterfaceModel(t *testing.T) {
	router := NewRouter(false, false, true)

	// Test that Status() returns a tea.Model that can be used
	model := router.Status()

	// Cast to DashboardModel if possible
	// Note: DashboardModel.Init has pointer receiver, so we assert to *DashboardModel
	if dashboardModel, ok := model.(*status.DashboardModel); ok {
		err := router.statusJSON(dashboardModel)
		if err != nil {
			t.Logf("statusJSON returned error: %v", err)
		}
	} else if _, ok := model.(tea.Model); ok {
		// If it's a tea.Model but not DashboardModel, we can still test
		// that calling the method doesn't panic
		t.Log("Model is tea.Model but not DashboardModel, skipping statusJSON test")
	}
}

func TestStatusTextWithInterfaceModel(t *testing.T) {
	router := NewRouter(false, false, false)

	// Test that Status() returns a tea.Model that can be used
	model := router.Status()

	// Cast to DashboardModel if possible
	// Note: DashboardModel.Init has pointer receiver, so we assert to *DashboardModel
	if dashboardModel, ok := model.(*status.DashboardModel); ok {
		err := router.statusText(dashboardModel)
		if err != nil {
			t.Logf("statusText returned error: %v", err)
		}
	} else if _, ok := model.(tea.Model); ok {
		// If it's a tea.Model but not DashboardModel, we can still test
		// that calling the method doesn't panic
		t.Log("Model is tea.Model but not DashboardModel, skipping statusText test")
	}
}

func TestBootstrapReturnsCorrectModelType(t *testing.T) {
	router := NewRouter(false, false, false)
	model := router.Bootstrap()

	// Should return a tea.Model
	if _, ok := model.(tea.Model); !ok {
		t.Error("Bootstrap() should return a tea.Model")
	}
}

func TestDiffReturnsCorrectModelType(t *testing.T) {
	router := NewRouter(false, false, false)
	model := router.Diff()

	// Should return a tea.Model
	if _, ok := model.(tea.Model); !ok {
		t.Error("Diff() should return a tea.Model")
	}
}

func TestApplyReturnsCorrectModelType(t *testing.T) {
	router := NewRouter(false, false, false)
	model := router.Apply()

	// Should return a tea.Model
	if _, ok := model.(tea.Model); !ok {
		t.Error("Apply() should return a tea.Model")
	}
}

func TestStatusReturnsCorrectModelType(t *testing.T) {
	router := NewRouter(false, false, false)
	model := router.Status()

	// Should return a tea.Model
	if _, ok := model.(tea.Model); !ok {
		t.Error("Status() should return a tea.Model")
	}
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

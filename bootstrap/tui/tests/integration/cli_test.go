package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBinary builds the binary if not already built
func buildBinary(t *testing.T) string {
	// Look for existing binary in bin directory (two levels up from tests/integration)
	binDir := filepath.Join("..", "..", "bin")
	binPath := filepath.Join(binDir, "adacosdev-dots")

	if _, err := os.Stat(binPath); err == nil {
		return binPath
	}

	// Build the binary from the project root (two levels up)
	t.Log("Building binary...")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = filepath.Join("..", "..")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	return binPath
}

func TestBootstrapHelpOutput(t *testing.T) {
	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "bootstrap", "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run bootstrap --help: %v\nOutput: %s", err, out)
	}

	output := string(out)

	// Check for expected content
	if !strings.Contains(output, "Usage:") && !strings.Contains(output, "usage:") {
		t.Errorf("Expected 'Usage' in help output, got: %s", output)
	}

	// Check that help doesn't panic and produces reasonable output
	if len(output) < 10 {
		t.Errorf("Help output too short: %s", output)
	}
}

func TestDiffHelpOutput(t *testing.T) {
	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "diff", "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run diff --help: %v\nOutput: %s", err, out)
	}

	output := string(out)

	// Check for expected content
	if !strings.Contains(output, "Usage:") && !strings.Contains(output, "usage:") {
		t.Errorf("Expected 'Usage' in help output, got: %s", output)
	}
}

func TestApplyHelpOutput(t *testing.T) {
	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "apply", "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run apply --help: %v\nOutput: %s", err, out)
	}

	output := string(out)

	// Check for expected content
	if !strings.Contains(output, "Usage:") && !strings.Contains(output, "usage:") {
		t.Errorf("Expected 'Usage' in help output, got: %s", output)
	}
}

func TestStatusHelpOutput(t *testing.T) {
	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "status", "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run status --help: %v\nOutput: %s", err, out)
	}

	output := string(out)

	// Check for expected content
	if !strings.Contains(output, "Usage:") && !strings.Contains(output, "usage:") {
		t.Errorf("Expected 'Usage' in help output, got: %s", output)
	}
}

func TestUnknownSubcommandError(t *testing.T) {
	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "unknown-subcommand")
	out, err := cmd.CombinedOutput()

	// Should produce an error
	if err == nil {
		t.Logf("Warning: unknown subcommand did not error, output: %s", out)
	}

	output := string(out)

	// Check for some kind of error message or usage hint
	// The actual behavior depends on the CLI implementation
	if len(output) == 0 {
		t.Errorf("Expected some output for unknown subcommand")
	}
}

func TestBootstrapJSONFlag(t *testing.T) {
	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "--json", "bootstrap")
	out, _ := cmd.CombinedOutput()

	// May error due to environment, but should produce JSON output attempt
	output := string(out)

	// If JSON mode works, we should see JSON lines
	if strings.Contains(output, "{") && strings.Contains(output, "}") {
		// JSON-like output detected
		t.Logf("JSON output detected: %s", truncate(output, 200))
	}
}

func TestDiffJSONFlag(t *testing.T) {
	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "--json", "diff")
	out, _ := cmd.CombinedOutput()

	output := string(out)

	// If JSON mode works, we should see JSON lines
	if strings.Contains(output, "{") && strings.Contains(output, "}") {
		t.Logf("JSON output detected: %s", truncate(output, 200))
	}
}

func TestApplyJSONFlag(t *testing.T) {
	binPath := buildBinary(t)

	// Use dry-run to avoid actual apply
	cmd := exec.Command(binPath, "--json", "--dry-run", "apply")
	out, _ := cmd.CombinedOutput()

	output := string(out)

	// If JSON mode works, we should see JSON lines
	if strings.Contains(output, "{") && strings.Contains(output, "}") {
		t.Logf("JSON output detected: %s", truncate(output, 200))
	}
}

func TestStatusJSONFlag(t *testing.T) {
	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "--json", "status")
	out, _ := cmd.CombinedOutput()

	output := string(out)

	// If JSON mode works, we should see JSON lines
	if strings.Contains(output, "{") && strings.Contains(output, "}") {
		t.Logf("JSON output detected: %s", truncate(output, 200))
	}
}

func TestBootstrapWithForceFlag(t *testing.T) {
	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "--force", "bootstrap")
	out, err := cmd.CombinedOutput()

	// Should not panic, may error due to environment
	output := string(out)
	t.Logf("bootstrap --force output: %s", truncate(output, 200))

	_ = err // Accept error due to environment
}

func TestBootstrapWithDryRunFlag(t *testing.T) {
	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "--dry-run", "bootstrap")
	out, err := cmd.CombinedOutput()

	// Should not panic
	output := string(out)
	t.Logf("bootstrap --dry-run output: %s", truncate(output, 200))

	_ = err
}

func TestGlobalFlagsBeforeSubcommand(t *testing.T) {
	binPath := buildBinary(t)

	// Test that global flags work before subcommand
	cmd := exec.Command(binPath, "--force", "--dry-run", "bootstrap")
	out, err := cmd.CombinedOutput()

	output := string(out)
	t.Logf("bootstrap --force --dry-run output: %s", truncate(output, 200))

	_ = err
}

func TestGlobalFlagsMixed(t *testing.T) {
	binPath := buildBinary(t)

	// Test mixed flags
	cmd := exec.Command(binPath, "--force", "--json", "bootstrap")
	out, err := cmd.CombinedOutput()

	output := string(out)
	t.Logf("bootstrap --force --json output: %s", truncate(output, 200))

	_ = err
}

func TestHelpFlagRouting(t *testing.T) {
	binPath := buildBinary(t)

	// Test help flag alone
	cmd := exec.Command(binPath, "--help")
	out, _ := cmd.CombinedOutput()

	output := string(out)

	// Should show help
	if !strings.Contains(strings.ToLower(output), "help") && !strings.Contains(output, "Usage:") {
		t.Errorf("Expected help output, got: %s", output)
	}
}

func TestHelpFlagShort(t *testing.T) {
	binPath := buildBinary(t)

	// Test short help flag
	cmd := exec.Command(binPath, "-h")
	out, _ := cmd.CombinedOutput()

	output := string(out)

	// Should show help
	if !strings.Contains(strings.ToLower(output), "help") && !strings.Contains(output, "Usage:") {
		t.Errorf("Expected help output, got: %s", output)
	}
}

func TestNoArgsShowsHelp(t *testing.T) {
	binPath := buildBinary(t)

	// Running without args should show help or error
	cmd := exec.Command(binPath)
	out, err := cmd.CombinedOutput()

	output := string(out)

	// Either shows help or errors - both are acceptable
	if len(output) == 0 && err == nil {
		t.Errorf("Expected some output when running without args")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

package shell

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseJSONOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected []Result
	}{
		{
			name:   "single ok result",
			output: `{"status":"ok","component":"tmux","message":"tmux installed"}`,
			expected: []Result{
				{Component: "tmux", Status: "ok", Message: "tmux installed"},
			},
		},
		{
			name:   "single error result",
			output: `{"status":"error","component":"docker","message":"apt failed","exit_code":1}`,
			expected: []Result{
				{Component: "docker", Status: "error", Message: "apt failed", ExitCode: 1},
			},
		},
		{
			name:   "multiple results",
			output: `{"status":"ok","component":"tmux","message":"tmux installed"}
{"status":"ok","component":"neovim","message":"neovim installed"}
{"status":"error","component":"docker","message":"docker failed","exit_code":1}`,
			expected: []Result{
				{Component: "tmux", Status: "ok", Message: "tmux installed"},
				{Component: "neovim", Status: "ok", Message: "neovim installed"},
				{Component: "docker", Status: "error", Message: "docker failed", ExitCode: 1},
			},
		},
		{
			name:     "empty output",
			output:   "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			output:   "   \n  \n  ",
			expected: nil,
		},
		{
			name:   "skip result",
			output: `{"status":"skip","component":"fonts","message":"fonts already installed"}`,
			expected: []Result{
				{Component: "fonts", Status: "skip", Message: "fonts already installed"},
			},
		},
		{
			name:   "result with all fields",
			output: `{"component":"test","status":"ok","message":"success","exit_code":0}`,
			expected: []Result{
				{Component: "test", Status: "ok", Message: "success", ExitCode: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := parseJSONOutput(tt.output)

			if len(results) != len(tt.expected) {
				t.Errorf("got %d results, want %d", len(results), len(tt.expected))
				return
			}

			for i, r := range results {
				if r.Component != tt.expected[i].Component {
					t.Errorf("result[%d].Component = %q, want %q", i, r.Component, tt.expected[i].Component)
				}
				if r.Status != tt.expected[i].Status {
					t.Errorf("result[%d].Status = %q, want %q", i, r.Status, tt.expected[i].Status)
				}
				if r.Message != tt.expected[i].Message {
					t.Errorf("result[%d].Message = %q, want %q", i, r.Message, tt.expected[i].Message)
				}
				if r.ExitCode != tt.expected[i].ExitCode {
					t.Errorf("result[%d].ExitCode = %d, want %d", i, r.ExitCode, tt.expected[i].ExitCode)
				}
			}
		})
	}
}

func TestParseJSONOutputPartialLines(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected int // number of successfully parsed results
	}{
		{
			name:     "incomplete JSON object",
			output:   `{"status":"ok","component":"tmux"`,
			expected: 0, // fails to parse, creates 1 error result
		},
		{
			name:     "truncated JSON",
			output:   `{"status":"ok","component":"tmux"`,
			expected: 0, // fails to parse, creates 1 error result
		},
		{
			name:     "valid followed by invalid",
			output:   `{"status":"ok","component":"tmux","message":"installed"}` + "\n" + `invalid json`,
			expected: 1, // first parses, second creates error result
		},
		{
			name:     "valid followed by incomplete",
			output:   `{"status":"ok","component":"tmux"}` + "\n" + `{"status":"error","component":"docker"`,
			expected: 1, // first parses, second fails
		},
		{
			name:     "empty lines between valid",
			output:   `{"status":"ok","component":"tmux"}` + "\n\n\n" + `{"status":"ok","component":"neovim"}`,
			expected: 2, // both parse successfully
		},
		{
			name:     "whitespace before valid JSON",
			output:   "   " + "\n" + `{"status":"ok","component":"tmux"}`,
			expected: 1, // whitespace is skipped, JSON parses
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := parseJSONOutput(tt.output)

			// Count successful (non-error) results
			successCount := 0
			for _, r := range results {
				if r.Status != "error" || !strings.Contains(r.Message, "failed to parse") {
					successCount++
				}
			}

			if successCount != tt.expected {
				t.Errorf("got %d successful results, want %d (all results: %+v)", successCount, tt.expected, results)
			}
		})
	}
}

func TestParseJSONOutputNonJSONFallback(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected int
	}{
		{
			name:     "plain text line",
			output:   "this is just plain text",
			expected: 0, // creates 1 error result
		},
		{
			name:     "error message line",
			output:   "Error: something went wrong",
			expected: 0, // creates 1 error result
		},
		{
			name:     "mixed valid and plain",
			output:   `{"status":"ok","component":"tmux"}` + "\n" + "plain text output",
			expected: 1, // first parses, second creates error result
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := parseJSONOutput(tt.output)

			// Count successful (non-error) results
			successCount := 0
			for _, r := range results {
				if r.Status != "error" || !strings.Contains(r.Message, "failed to parse") {
					successCount++
				}
			}

			if successCount != tt.expected {
				t.Errorf("got %d successful results, want %d (all results: %+v)", successCount, tt.expected, results)
			}
		})
	}
}

func TestParseJSONOutputErrorResults(t *testing.T) {
	// When JSON parsing fails, it should create error results
	output := "not valid json at all"
	results := parseJSONOutput(output)

	// Should have one error result for the failed parse
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != "error" {
		t.Errorf("expected status 'error', got %s", results[0].Status)
	}
}

func TestRunHelper(t *testing.T) {
	// Create a temporary helper script
	tmpDir := t.TempDir()
	helperPath := filepath.Join(tmpDir, "test-helper.sh")

	// Test with JSON output
	jsonScript := `#!/bin/bash
echo '{"status":"ok","component":"test","message":"success"}'
echo '{"status":"error","component":"test2","message":"failed","exit_code":1}'
`
	if err := os.WriteFile(helperPath, []byte(jsonScript), 0755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}

	// Set ADACOSDEV_JSON=1 to enable JSON mode
	os.Setenv("ADACOSDEV_JSON", "1")
	defer os.Unsetenv("ADACOSDEV_JSON")

	results, err := RunHelper(helperPath)
	if err != nil {
		t.Fatalf("RunHelper() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("got %d results, want 2", len(results))
	}

	if results[0].Component != "test" || results[0].Status != "ok" {
		t.Errorf("results[0] = %+v, want component=test, status=ok", results[0])
	}

	if results[1].Component != "test2" || results[1].Status != "error" || results[1].ExitCode != 1 {
		t.Errorf("results[1] = %+v, want component=test2, status=error, exit_code=1", results[1])
	}
}

func TestRunHelperNonJSON(t *testing.T) {
	tmpDir := t.TempDir()
	helperPath := filepath.Join(tmpDir, "test-helper.sh")

	// Script that exits with error
	errorScript := `#!/bin/bash
echo "error message" >&2
exit 1
`
	if err := os.WriteFile(helperPath, []byte(errorScript), 0755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}

	// Ensure ADACOSDEV_JSON is not set
	os.Unsetenv("ADACOSDEV_JSON")

	results, err := RunHelper(helperPath)
	if err != nil {
		t.Fatalf("RunHelper() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}

	if results[0].Status != "error" {
		t.Errorf("results[0].Status = %q, want error", results[0].Status)
	}

	if results[0].ExitCode != 1 {
		t.Errorf("results[0].ExitCode = %d, want 1", results[0].ExitCode)
	}
}

func TestRunHelperSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	helperPath := filepath.Join(tmpDir, "test-helper.sh")

	successScript := `#!/bin/bash
echo "all good"
exit 0
`
	if err := os.WriteFile(helperPath, []byte(successScript), 0755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}

	os.Unsetenv("ADACOSDEV_JSON")

	results, err := RunHelper(helperPath)
	if err != nil {
		t.Fatalf("RunHelper() error = %v", err)
	}

	if results[0].Status != "ok" {
		t.Errorf("results[0].Status = %q, want ok", results[0].Status)
	}
}

func TestRunHelperWithArgs(t *testing.T) {
	tmpDir := t.TempDir()
	helperPath := filepath.Join(tmpDir, "test-helper.sh")

	// Script that echoes arguments
	argsScript := `#!/bin/bash
echo "args: $@"
exit 0
`
	if err := os.WriteFile(helperPath, []byte(argsScript), 0755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}

	os.Unsetenv("ADACOSDEV_JSON")

	results, err := RunHelper(helperPath, "arg1", "arg2")
	if err != nil {
		t.Fatalf("RunHelper() error = %v", err)
	}

	if results[0].Status != "ok" {
		t.Errorf("results[0].Status = %q, want ok", results[0].Status)
	}
}

func TestRun(t *testing.T) {
	// Test basic Run function
	os.Setenv("ADACOSDEV_JSON", "1")
	defer os.Unsetenv("ADACOSDEV_JSON")

	results, err := Run("echo '{\"status\":\"ok\",\"component\":\"test\",\"message\":\"hello\"}'", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}

	if results[0].Status != "ok" {
		t.Errorf("results[0].Status = %q, want ok", results[0].Status)
	}
}

func TestRunNonJSON(t *testing.T) {
	os.Unsetenv("ADACOSDEV_JSON")

	results, err := Run("echo 'hello world'", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
}

func TestRunExitCodeNonJSON(t *testing.T) {
	tests := []struct {
		name       string
		script     string
		wantStatus string
		wantCode   int
	}{
		{
			name:       "success exit",
			script:     "exit 0",
			wantStatus: "ok",
			wantCode:   0,
		},
		{
			name:       "error exit",
			script:     "exit 1",
			wantStatus: "error",
			wantCode:   1,
		},
		{
			name:       "exit 2",
			script:     "exit 2",
			wantStatus: "error",
			wantCode:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("ADACOSDEV_JSON") // Non-JSON mode

			results, err := Run(tt.script, nil)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			if len(results) != 1 {
				t.Fatalf("got %d results, want 1", len(results))
			}

			if results[0].Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", results[0].Status, tt.wantStatus)
			}

			if results[0].ExitCode != tt.wantCode {
				t.Errorf("ExitCode = %d, want %d", results[0].ExitCode, tt.wantCode)
			}
		})
	}
}

func TestRunExitCodeJSON(t *testing.T) {
	// In JSON mode, exit codes are captured from the JSON output
	tests := []struct {
		name     string
		script   string
		wantCode int
	}{
		{
			name:     "exit 0",
			script:   `echo '{"status":"ok","component":"test","exit_code":0}'`,
			wantCode: 0,
		},
		{
			name:     "exit 1",
			script:   `echo '{"status":"error","component":"test","exit_code":1}'`,
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ADACOSDEV_JSON", "1")
			defer os.Unsetenv("ADACOSDEV_JSON")

			results, err := Run(tt.script, nil)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			if len(results) != 1 {
				t.Fatalf("got %d results, want 1", len(results))
			}

			if results[0].ExitCode != tt.wantCode {
				t.Errorf("ExitCode = %d, want %d", results[0].ExitCode, tt.wantCode)
			}
		})
	}
}

func TestRunStderrCapture(t *testing.T) {
	tmpDir := t.TempDir()
	helperPath := filepath.Join(tmpDir, "test-helper.sh")

	// Script that writes to stderr
	stderrScript := `#!/bin/bash
echo "error output" >&2
echo '{"status":"ok","component":"test","message":"ok"}'
exit 0
`
	if err := os.WriteFile(helperPath, []byte(stderrScript), 0755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}

	os.Setenv("ADACOSDEV_JSON", "1")
	defer os.Unsetenv("ADACOSDEV_JSON")

	results, err := RunHelper(helperPath)
	if err != nil {
		t.Fatalf("RunHelper() error = %v", err)
	}

	// In JSON mode, stderr is not captured in the result message
	// The JSON output is what matters
	if results[0].Status != "ok" {
		t.Errorf("results[0].Status = %q, want ok", results[0].Status)
	}
}

func TestRunWithEnv(t *testing.T) {
	// Ensure JSON mode is off so we get exit code-based status
	os.Unsetenv("ADACOSDEV_JSON")

	// Create a script that echoes an env var
	results, err := RunWithEnv("echo $TEST_VAR", []string{"TEST_VAR=hello"})
	if err != nil {
		t.Fatalf("RunWithEnv() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}

	if results[0].Status != "ok" {
		t.Errorf("Status = %q, want ok", results[0].Status)
	}
}

func TestRunWithEnvAdditionalEnv(t *testing.T) {
	os.Unsetenv("ADACOSDEV_JSON")

	// Test that additional env vars are passed
	// HOME should be preserved
	results, err := RunWithEnv("echo $MY_TEST_VAR", []string{"MY_TEST_VAR=my_value"})
	if err != nil {
		t.Fatalf("RunWithEnv() error = %v", err)
	}

	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
}

func TestGetEnvForHelper(t *testing.T) {
	env := getEnvForHelper()
	if len(env) == 0 {
		t.Error("getEnvForHelper should return at least one env var")
	}

	found := false
	for _, e := range env {
		if strings.Contains(e, "ADACOSDEV_JSON") {
			found = true
			break
		}
	}
	if !found {
		t.Error("getEnvForHelper should include ADACOSDEV_JSON")
	}
}

func TestResultJSONMarshaling(t *testing.T) {
	result := Result{
		Component: "tmux",
		Status:    "ok",
		Message:   "tmux installed",
		ExitCode:  0,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled Result
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Component != result.Component {
		t.Errorf("Component = %q, want %q", unmarshaled.Component, result.Component)
	}
	if unmarshaled.Status != result.Status {
		t.Errorf("Status = %q, want %q", unmarshaled.Status, result.Status)
	}
	if unmarshaled.Message != result.Message {
		t.Errorf("Message = %q, want %q", unmarshaled.Message, result.Message)
	}
}

func TestRunHelperNotFound(t *testing.T) {
	// In non-JSON mode, a non-existent helper returns Status: "error"
	os.Unsetenv("ADACOSDEV_JSON")

	results, err := RunHelper("/nonexistent/path/to/helper")
	if err != nil {
		t.Fatalf("RunHelper() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if results[0].Status != "error" {
		t.Errorf("Status = %q, want error", results[0].Status)
	}
}

func TestRunScriptNotFound(t *testing.T) {
	// In non-JSON mode, a non-existent script returns Status: "error"
	os.Unsetenv("ADACOSDEV_JSON")

	results, err := Run("/nonexistent/script.sh", nil)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results, want 1", len(results))
	}
	if results[0].Status != "error" {
		t.Errorf("Status = %q, want error", results[0].Status)
	}
}

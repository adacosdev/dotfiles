// Package shell provides shell command execution utilities.
package shell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result represents a shell command result.
type Result struct {
	Component string `json:"component"`
	Status    string `json:"status"` // "ok", "error", "skip"
	Message   string `json:"message"`
	ExitCode  int    `json:"exit_code,omitempty"`
}

// Run executes a shell script and returns structured results.
// When ADACOSDEV_JSON=1 is set, helpers must emit JSON lines to stdout.
func Run(script string, env []string) ([]Result, error) {
	useJSON := os.Getenv("ADACOSDEV_JSON") == "1"

	cmd := exec.Command("bash", "-c", script)
	cmd.Env = append(os.Environ(), env...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute script: %w", err)
		}
	}

	if useJSON {
		return parseJSONOutput(stdout.String()), nil
	}

	status := "ok"
	if exitCode != 0 {
		status = "error"
	}
	return []Result{{
		Component: script,
		Status:    status,
		Message:   strings.TrimSpace(stderr.String()),
		ExitCode:  exitCode,
	}}, nil
}

// RunHelper runs a specific helper (e.g., "linux/bootstrap-linux.sh")
// and parses its JSON output if ADACOSDEV_JSON=1.
func RunHelper(helperPath string, args ...string) ([]Result, error) {
	useJSON := os.Getenv("ADACOSDEV_JSON") == "1"

	cmd := exec.Command("bash", append([]string{helperPath}, args...)...)
	cmd.Env = append(os.Environ(), getEnvForHelper()...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute %s: %w", helperPath, err)
		}
	}

	if useJSON {
		return parseJSONOutput(stdout.String()), nil
	}

	// Non-JSON mode: return a simple result
	status := "ok"
	if exitCode != 0 {
		status = "error"
	}
	return []Result{{
		Component: helperPath,
		Status:    status,
		Message:   strings.TrimSpace(stderr.String()),
		ExitCode:  exitCode,
	}}, nil
}

// parseJSONOutput parses JSON lines from helper output.
func parseJSONOutput(output string) []Result {
	var results []Result
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		var r Result
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			// If we can't parse, create an error result
			results = append(results, Result{
				Status:  "error",
				Message: fmt.Sprintf("failed to parse JSON: %s", err),
			})
			continue
		}
		results = append(results, r)
	}
	return results
}

// getEnvForHelper returns the environment variables to pass to helpers.
func getEnvForHelper() []string {
	return []string{
		"ADACOSDEV_JSON=1",
	}
}

// RunWithEnv executes a shell script with additional environment variables
// and returns structured results.
func RunWithEnv(script string, additionalEnv []string) ([]Result, error) {
	useJSON := os.Getenv("ADACOSDEV_JSON") == "1"

	cmd := exec.Command("bash", "-c", script)
	cmd.Env = append(os.Environ(), additionalEnv...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute script: %w", err)
		}
	}

	if useJSON {
		return parseJSONOutput(stdout.String()), nil
	}

	status := "ok"
	if exitCode != 0 {
		status = "error"
	}
	return []Result{{
		Component: script,
		Status:    status,
		Message:   strings.TrimSpace(stderr.String()),
		ExitCode:  exitCode,
	}}, nil
}

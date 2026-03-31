// Package bootstrap provides the bootstrap wizard and component management.
package bootstrap

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/adacosdev/dotfiles/bootstrap/tui/internal/bootstrap/detector"
	"github.com/adacosdev/dotfiles/bootstrap/tui/pkg/shell"
)

// Executor runs the bootstrap process.
type Executor struct {
	OS      *detector.OSInfo
	Helpers string // path to bootstrap helpers dir
	DryRun  bool
	Force   bool
}

// Result represents the result of a component installation.
type Result struct {
	Component string `json:"component"`
	Name      string `json:"name"`
	Status    string `json:"status"` // "pending", "installed", "skipped", "error"
	Message   string `json:"message"`
	ExitCode  int    `json:"exit_code,omitempty"`
}

// Execute runs the given components and streams results via channel.
func (e *Executor) Execute(components []Component, jsonMode bool) <-chan Result {
	results := make(chan Result)

	go func() {
		defer close(results)

		for _, c := range components {
			result := Result{
				Component: c.ID,
				Name:      c.Name,
				Status:    "pending",
			}

			// Check if already installed
			installed, version := c.IsInstalled(e.OS)
			if installed && !e.Force {
				result.Status = "skipped"
				result.Message = fmt.Sprintf("Already installed: %s", version)
				results <- result
				continue
			}

			// Dry run mode
			if e.DryRun {
				result.Status = "pending"
				result.Message = fmt.Sprintf("Would install: %s", c.InstallCmd)
				results <- result
				continue
			}

			// Execute installation
			execResult := e.executeComponent(c)
			result.Status = execResult.Status
			result.Message = execResult.Message
			result.ExitCode = execResult.ExitCode
			results <- result
		}
	}()

	return results
}

// executeComponent runs the installation for a single component.
func (e *Executor) executeComponent(c Component) Result {
	// Try using helper first
	if e.Helpers != "" {
		helperPath := fmt.Sprintf("%s/%s.sh", e.Helpers, c.ID)
		if _, err := os.Stat(helperPath); err == nil {
			return e.runHelper(helperPath, c)
		}
	}

	// Fallback to direct command
	return e.runCommand(c)
}

// runHelper runs a helper script and returns the result.
func (e *Executor) runHelper(helperPath string, c Component) Result {
	// Set JSON mode for structured output
	env := os.Environ()
	env = append(env, "ADACOSDEV_JSON=1")

	cmd := exec.Command("bash", helperPath)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return Result{
				Component: c.ID,
				Name:      c.Name,
				Status:    "error",
				Message:   fmt.Sprintf("failed to execute helper: %v", err),
			}
		}
	}

	// Parse JSON output from helper
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) > 0 {
		// Try to parse as JSON
		for _, line := range lines {
			if line == "" {
				continue
			}
			var shellResult shell.Result
			if err := json.Unmarshal([]byte(line), &shellResult); err == nil {
				return Result{
					Component: shellResult.Component,
					Name:      c.Name,
					Status:    mapShellStatus(shellResult.Status),
					Message:   shellResult.Message,
					ExitCode:  shellResult.ExitCode,
				}
			}
		}
	}

	// Fallback: interpret exit code
	status := "installed"
	if exitCode != 0 {
		status = "error"
	}

	return Result{
		Component: c.ID,
		Name:      c.Name,
		Status:    status,
		Message:   strings.TrimSpace(string(output)),
		ExitCode:  exitCode,
	}
}

// runCommand runs the component's install command directly.
func (e *Executor) runCommand(c Component) Result {
	// Check if we need root
	if c.NeedsRoot && !e.hasRootPrivileges() {
		return Result{
			Component: c.ID,
			Name:      c.Name,
			Status:    "error",
			Message:   "requires root privileges",
		}
	}

	// Handle bootstrap-helper commands
	cmdStr := c.InstallCmd
	if strings.HasPrefix(cmdStr, "bootstrap-helper ") {
		helper := strings.TrimPrefix(cmdStr, "bootstrap-helper ")
		if e.Helpers != "" {
			cmdStr = fmt.Sprintf("%s/%s.sh", e.Helpers, helper)
		} else {
			// Look in standard locations
			cmdStr = e.findHelper(helper)
		}
	}

	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return Result{
				Component: c.ID,
				Name:      c.Name,
				Status:    "error",
				Message:   fmt.Sprintf("failed to execute: %v", err),
			}
		}
	}

	status := "installed"
	if exitCode != 0 {
		status = "error"
	}

	return Result{
		Component: c.ID,
		Name:      c.Name,
		Status:    status,
		Message:   strings.TrimSpace(string(output)),
		ExitCode:  exitCode,
	}
}

// findHelper searches for a helper script in standard locations.
func (e *Executor) findHelper(name string) string {
	locations := []string{
		"/usr/local/bin/bootstrap-helpers",
		"/usr/bin/bootstrap-helpers",
		filepath.Join(os.Getenv("HOME"), ".local/bin/bootstrap-helpers"),
		filepath.Join(os.Getenv("HOME"), ".bootstrap-helpers"),
	}

	for _, loc := range locations {
		helperPath := filepath.Join(loc, name+".sh")
		if _, err := os.Stat(helperPath); err == nil {
			return helperPath
		}
	}

	// Fallback to assuming it's in PATH
	return name
}

// hasRootPrivileges checks if running with root privileges.
func (e *Executor) hasRootPrivileges() bool {
	return os.Geteuid() == 0
}

// mapShellStatus maps shell.Result status to Executor Result status.
func mapShellStatus(shellStatus string) string {
	switch shellStatus {
	case "ok", "installed":
		return "installed"
	case "skip", "skipped":
		return "skipped"
	case "error":
		return "error"
	default:
		return shellStatus
	}
}

// ExecuteSync runs components and returns all results as a slice.
func (e *Executor) ExecuteSync(components []Component) []Result {
	results := make([]Result, 0, len(components))
	for result := range e.Execute(components, false) {
		results = append(results, result)
	}
	return results
}

// ParallelExecutor is a parallel version of Executor.
type ParallelExecutor struct {
	*Executor
	MaxParallel int
}

// NewParallelExecutor creates a new parallel executor.
func NewParallelExecutor(exec *Executor, maxParallel int) *ParallelExecutor {
	if maxParallel <= 0 {
		maxParallel = 3
	}
	return &ParallelExecutor{
		Executor:   exec,
		MaxParallel: maxParallel,
	}
}

// Execute runs components in parallel and streams results via channel.
func (p *ParallelExecutor) Execute(components []Component, jsonMode bool) <-chan Result {
	results := make(chan Result)
	input := make(chan Component, len(components))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < p.MaxParallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range input {
				for result := range p.Executor.Execute([]Component{c}, jsonMode) {
					results <- result
				}
			}
		}()
	}

	// Send components to workers
	go func() {
		defer close(input)
		for _, c := range components {
			input <- c
		}
	}()

	// Close results when workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

// ExecuteWithProgress runs components and calls onResult for each result.
func (e *Executor) ExecuteWithProgress(components []Component, onResult func(Result)) {
	for result := range e.Execute(components, false) {
		onResult(result)
	}
}

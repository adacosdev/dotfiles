// Package detector provides OS detection utilities.
package detector

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// OSInfo contains information about the detected operating system.
type OSInfo struct {
	ID     string // "ubuntu", "debian", "arch", "fedora", "darwin", "windows"
	Name   string // "Ubuntu 24.04", "macOS 14", "Windows 11"
	Family string // "linux", "bsd", "windows"
	IsWSL  bool
	Arch   string // "amd64", "arm64"
}

// Detect returns OS information for the current system.
func Detect() (*OSInfo, error) {
	osInfo := &OSInfo{
		Arch: getGoArch(),
	}

	// Check for WSL first
	osInfo.IsWSL = isWSL()

	switch getGoOS() {
	case "linux":
		if err := detectLinux(osInfo); err != nil {
			return nil, fmt.Errorf("failed to detect Linux: %w", err)
		}
		osInfo.Family = "linux"
	case "darwin":
		if err := detectMacOS(osInfo); err != nil {
			return nil, fmt.Errorf("failed to detect macOS: %w", err)
		}
		osInfo.Family = "bsd"
	case "windows":
		osInfo.ID = "windows"
		osInfo.Name = "Windows"
		osInfo.Family = "windows"
		if osInfo.IsWSL {
			osInfo.Family = "linux"
		}
	default:
		return nil, fmt.Errorf("unsupported OS: %s", getGoOS())
	}

	return osInfo, nil
}

// DetectOS is the legacy single-value function that returns the OS ID.
func DetectOS() string {
	info, err := Detect()
	if err != nil {
		return "unknown"
	}
	return info.ID
}

// detectLinux populates OSInfo for Linux systems.
func detectLinux(osInfo *OSInfo) error {
	// Try to read /etc/os-release
	file, err := os.Open("/etc/os-release")
	if err == nil {
		defer file.Close()
		return parseOSRelease(file, osInfo)
	}

	// Fallback to uname
	return fallbackUname(osInfo)
}

// parseOSRelease parses /etc/os-release to extract OS info.
func parseOSRelease(file *os.File, osInfo *OSInfo) error {
	scanner := bufio.NewScanner(file)
	var (
		id, name, versionID string
	)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			id = strings.TrimPrefix(line, "ID=")
			id = strings.Trim(id, `"`)
		} else if strings.HasPrefix(line, "NAME=") {
			name = strings.TrimPrefix(line, "NAME=")
			name = strings.Trim(name, `"`)
		} else if strings.HasPrefix(line, "VERSION_ID=") {
			versionID = strings.TrimPrefix(line, "VERSION_ID=")
			versionID = strings.Trim(versionID, `"`)
		}
	}

	if id == "" {
		return fmt.Errorf("could not parse ID from /etc/os-release")
	}

	osInfo.ID = id
	if name != "" {
		if versionID != "" {
			osInfo.Name = fmt.Sprintf("%s %s", name, versionID)
		} else {
			osInfo.Name = name
		}
	} else {
		osInfo.Name = id
	}

	return nil
}

// fallbackUname uses uname as fallback for Linux detection.
func fallbackUname(osInfo *OSInfo) error {
	output, err := exec.Command("uname", "-s").Output()
	if err != nil {
		return fmt.Errorf("uname failed: %w", err)
	}
	osInfo.ID = strings.ToLower(strings.TrimSpace(string(output)))
	osInfo.Name = strings.TrimSpace(string(output))
	return nil
}

// detectMacOS populates OSInfo for macOS systems.
func detectMacOS(osInfo *OSInfo) error {
	// Get product name and version via sw_vers
	productName, err := runCommand("sw_vers", "-productName")
	if err != nil {
		osInfo.Name = "macOS"
	} else {
		osInfo.Name = strings.TrimSpace(productName)
	}

	productVersion, err := runCommand("sw_vers", "-productVersion")
	if err == nil {
		// Append version if we got product name
		if osInfo.Name != "macOS" {
			osInfo.Name = fmt.Sprintf("%s %s", osInfo.Name, strings.TrimSpace(productVersion))
		} else {
			osInfo.Name = strings.TrimSpace(productVersion)
		}
	}

	// Get architecture via uname
	arch, err := runCommand("uname", "-m")
	if err == nil {
		arch = strings.TrimSpace(arch)
		// Normalize architecture names
		switch arch {
		case "x86_64":
			osInfo.Arch = "amd64"
		case "arm64":
			osInfo.Arch = "arm64"
		default:
			osInfo.Arch = arch
		}
	}

	// Set ID based on Apple Silicon vs Intel
	if strings.Contains(arch, "arm64") {
		osInfo.ID = "darwin-arm64"
	} else {
		osInfo.ID = "darwin"
	}

	return nil
}

// isWSL checks if running on Windows Subsystem for Linux.
func isWSL() bool {
	// Check /proc/version for WSL signature
	file, err := os.Open("/proc/version")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		content := strings.ToLower(scanner.Text())
		if strings.Contains(content, "microsoft") && strings.Contains(content, "wsl") {
			return true
		}
	}

	return false
}

// getGoOS returns GOOS environment variable or runtime.GOOS.
func getGoOS() string {
	if os.Getenv("GOOS") != "" {
		return os.Getenv("GOOS")
	}
	out, err := exec.Command("go", "env", "GOOS").Output()
	if err != nil {
		return "linux" // sensible default
	}
	return strings.TrimSpace(string(out))
}

// getGoArch returns GOARCH environment variable or runtime.GOARCH.
func getGoArch() string {
	if os.Getenv("GOARCH") != "" {
		return os.Getenv("GOARCH")
	}
	out, err := exec.Command("go", "env", "GOARCH").Output()
	if err != nil {
		return "amd64" // sensible default
	}
	return strings.TrimSpace(string(out))
}

// runCommand runs a command and returns its trimmed output.
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// NormalizeOSFamily returns a normalized family name for an OS ID.
func NormalizeOSFamily(id string) string {
	families := map[string]string{
		"ubuntu":     "linux",
		"debian":     "linux",
		"arch":       "linux",
		"fedora":     "linux",
		"centos":     "linux",
		"rhel":       "linux",
		"amazon":     "linux",
		"flatcar":    "linux",
		"darwin":     "bsd",
		"macos":      "bsd",
		"freebsd":    "bsd",
		"openbsd":    "bsd",
		"netbsd":     "bsd",
		"windows":    "windows",
		"darwin-arm64": "bsd",
	}

	if family, ok := families[id]; ok {
		return family
	}
	if strings.HasPrefix(id, "darwin") {
		return "bsd"
	}
	if strings.HasPrefix(id, "linux") {
		return "linux"
	}
	return id
}

// IsSupportedDistro checks if a given OS ID is a supported distribution.
func IsSupportedDistro(id string) bool {
	supportedDistros := []string{
		"ubuntu", "debian", "arch", "fedora", "darwin", "windows",
		"darwin-arm64", "amazon", "centos", "rhel", "flatcar",
	}
	for _, d := range supportedDistros {
		if d == id {
			return true
		}
	}
	return false
}

// distroInfoRegex matches VERSION_ID in os-release files.
var distroInfoRegex = regexp.MustCompile(`VERSION_ID="?([^"\n]+)"?`)

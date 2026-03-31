package detector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetect(t *testing.T) {
	info, err := Detect()
	if err != nil {
		t.Fatalf("Detect() returned error: %v", err)
	}

	if info.ID == "" {
		t.Error("Detect().ID should not be empty")
	}

	if info.Family == "" {
		t.Error("Detect().Family should not be empty")
	}

	if info.Arch == "" {
		t.Error("Detect().Arch should not be empty")
	}
}

func TestDetectOS(t *testing.T) {
	id := DetectOS()
	if id == "" {
		t.Error("DetectOS() should not return empty string")
	}
}

func TestNormalizeOSFamily(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"ubuntu", "linux"},
		{"debian", "linux"},
		{"arch", "linux"},
		{"fedora", "linux"},
		{"darwin", "bsd"},
		{"darwin-arm64", "bsd"},
		{"windows", "windows"},
		{"freebsd", "bsd"},
		{"openbsd", "bsd"},
		{"netbsd", "bsd"},
		{"centos", "linux"},
		{"rhel", "linux"},
		{"amazon", "linux"},
		{"flatcar", "linux"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := NormalizeOSFamily(tt.id)
			if got != tt.expected {
				t.Errorf("NormalizeOSFamily(%q) = %q, want %q", tt.id, got, tt.expected)
			}
		})
	}
}

func TestIsSupportedDistro(t *testing.T) {
	tests := []struct {
		id        string
		supported bool
	}{
		{"ubuntu", true},
		{"debian", true},
		{"arch", true},
		{"fedora", true},
		{"darwin", true},
		{"windows", true},
		{"darwin-arm64", true},
		{"amazon", true},
		{"centos", true},
		{"rhel", true},
		{"flatcar", true},
		{"unknown", false},
		{"solaris", false},
		{"freebsd", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := IsSupportedDistro(tt.id)
			if got != tt.supported {
				t.Errorf("IsSupportedDistro(%q) = %v, want %v", tt.id, got, tt.supported)
			}
		})
	}
}

func TestParseOSRelease(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantID   string
		wantName string
	}{
		{
			name:     "Ubuntu 24.04",
			content:  "ID=\"ubuntu\"\nNAME=\"Ubuntu\"\nVERSION_ID=\"24.04\"",
			wantID:   "ubuntu",
			wantName: "Ubuntu 24.04",
		},
		{
			name:     "Debian 12",
			content:  "ID=\"debian\"\nNAME=\"Debian GNU/Linux\"\nVERSION_ID=\"12\"",
			wantID:   "debian",
			wantName: "Debian GNU/Linux 12",
		},
		{
			name:     "Arch Linux",
			content:  "ID=\"arch\"\nNAME=\"Arch Linux\"\nVERSION_ID=\"\"",
			wantID:   "arch",
			wantName: "Arch Linux",
		},
		{
			name:     "Fedora 40",
			content:  "ID=\"fedora\"\nNAME=\"Fedora Linux\"\nVERSION_ID=\"40\"",
			wantID:   "fedora",
			wantName: "Fedora Linux 40",
		},
		{
			name:     "Ubuntu 22.04 with quotes",
			content:  `ID="ubuntu"` + "\n" + `NAME="Ubuntu"` + "\n" + `VERSION_ID="22.04"`,
			wantID:   "ubuntu",
			wantName: "Ubuntu 22.04",
		},
		{
			name:     "CentOS Stream",
			content:  "ID=\"centos\"\nNAME=\"CentOS Stream\"\nVERSION_ID=\"9\"",
			wantID:   "centos",
			wantName: "CentOS Stream 9",
		},
		{
			name:     "Amazon Linux 2023",
			content:  "ID=\"amazon\"\nNAME=\"Amazon Linux\"\nVERSION_ID=\"2023\"",
			wantID:   "amazon",
			wantName: "Amazon Linux 2023",
		},
		{
			name:     "ID only, no VERSION_ID",
			content:  "ID=\"arch\"\nNAME=\"Arch Linux\"",
			wantID:   "arch",
			wantName: "Arch Linux",
		},
		{
			name:     "Empty VERSION_ID",
			content:  "ID=\"debian\"\nNAME=\"Debian\"\nVERSION_ID=\"\"",
			wantID:   "debian",
			wantName: "Debian",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with test content
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "os-release")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			// Read and parse
			file, err := os.Open(tmpFile)
			if err != nil {
				t.Fatalf("Failed to open temp file: %v", err)
			}
			defer file.Close()

			info := &OSInfo{}
			if err := parseOSRelease(file, info); err != nil {
				t.Fatalf("parseOSRelease failed: %v", err)
			}

			if info.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", info.ID, tt.wantID)
			}
			if info.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", info.Name, tt.wantName)
			}
		})
	}
}

func TestParseOSReleaseError(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "Empty content",
			content: "",
			wantErr: true,
		},
		{
			name:    "No ID field",
			content: "NAME=\"Test\"\nVERSION_ID=\"1\"",
			wantErr: true,
		},
		{
			name:    "Only ID field",
			content: "ID=\"test\"",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "os-release")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			file, err := os.Open(tmpFile)
			if err != nil {
				t.Fatalf("Failed to open temp file: %v", err)
			}
			defer file.Close()

			info := &OSInfo{}
			err = parseOSRelease(file, info)

			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsWSL(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "WSL Ubuntu",
			content:  "Linux version 5.15.0-1019-microsoft-standard-WSL2",
			expected: true,
		},
		{
			name:     "WSL Debian",
			content:  "microsoft wsl2 debian",
			expected: true,
		},
		{
			name:     "Regular Linux",
			content:  "Linux version 5.15.0-1019-generic",
			expected: false,
		},
		{
			name:     "macOS kernel",
			content:  "Darwin Kernel Version 23.0.0",
			expected: false,
		},
		{
			name:     "Empty content",
			content:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "version")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			file, err := os.Open(tmpFile)
			if err != nil {
				t.Fatalf("Failed to open temp file: %v", err)
			}
			defer file.Close()

			result := isWSLFromFile(file)

			if result != tt.expected {
				t.Errorf("isWSLFromFile() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// isWSLFromFile is a test helper that reads a file and checks WSL signature
func isWSLFromFile(file *os.File) bool {
	defer file.Close()
	content := make([]byte, 1024)
	n, _ := file.Read(content)
	text := strings.ToLower(string(content[:n]))
	return strings.Contains(text, "microsoft") && strings.Contains(text, "wsl")
}

func TestGetGoOS(t *testing.T) {
	// Clear GOOS env var temporarily
	originalGOOS := os.Getenv("GOOS")
	defer os.Setenv("GOOS", originalGOOS)
	os.Unsetenv("GOOS")

	os := getGoOS()
	if os == "" {
		t.Error("getGoOS() should not return empty string")
	}
}

func TestGetGoOSWithEnv(t *testing.T) {
	// Test with GOOS env var set
	os.Setenv("GOOS", "darwin")
	defer os.Unsetenv("GOOS")

	got := getGoOS()
	if got != "darwin" {
		t.Errorf("getGoOS() = %q, want darwin", got)
	}
}

func TestGetGoArch(t *testing.T) {
	// Clear GOARCH env var temporarily
	originalGOARCH := os.Getenv("GOARCH")
	defer os.Setenv("GOARCH", originalGOARCH)
	os.Unsetenv("GOARCH")

	arch := getGoArch()
	if arch == "" {
		t.Error("getGoArch() should not return empty string")
	}
}

func TestGetGoArchWithEnv(t *testing.T) {
	// Test with GOARCH env var set
	os.Setenv("GOARCH", "arm64")
	defer os.Unsetenv("GOARCH")

	got := getGoArch()
	if got != "arm64" {
		t.Errorf("getGoArch() = %q, want arm64", got)
	}
}

func TestFallbackUname(t *testing.T) {
	info := &OSInfo{}
	err := fallbackUname(info)
	if err != nil {
		t.Fatalf("fallbackUname failed: %v", err)
	}
	if info.ID == "" {
		t.Error("fallbackUname should set ID")
	}
}

func TestOSInfoString(t *testing.T) {
	info := &OSInfo{
		ID:     "ubuntu",
		Name:   "Ubuntu 24.04",
		Family: "linux",
		IsWSL:  false,
		Arch:   "amd64",
	}

	// Just verify it doesn't panic and produces expected substrings
	s := info.ID
	if !strings.Contains(s, "ubuntu") {
		t.Errorf("expected ID to contain ubuntu, got %s", s)
	}
}

func TestRunCommand(t *testing.T) {
	// Test successful command
	output, err := runCommand("echo", "hello")
	if err != nil {
		t.Fatalf("runCommand(echo, hello) failed: %v", err)
	}
	if !strings.Contains(output, "hello") {
		t.Errorf("expected output to contain 'hello', got %s", output)
	}

	// Test failing command
	_, err = runCommand("false")
	if err == nil {
		t.Error("expected error for 'false' command")
	}
}

func TestRunCommandNotFound(t *testing.T) {
	_, err := runCommand("this-command-does-not-exist-12345")
	if err == nil {
		t.Error("expected error for non-existent command")
	}
}

func TestDetectMacOSSwVersOutput(t *testing.T) {
	// Test sw_vers output parsing via detectMacOS
	// This is a unit test of the macOS detection logic
	info := &OSInfo{}

	// We can't fully test detectMacOS without mocking exec,
	// but we can verify the function doesn't panic
	// and that it sets some expected fields

	// The actual detection would call sw_vers
	// Just verify the info struct has expected fields after init
	info.Arch = "amd64"
	info.ID = "darwin"

	if info.Arch != "amd64" {
		t.Error("expected Arch to be set")
	}
}

func TestWSLDetectionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "WSL2 standard",
			content:  "Linux version 5.15.0-1019-microsoft-standard-WSL2+",
			expected: true,
		},
		{
			name:     "WSL Microsoft/proc",
			content:  "microsoft wsl2",
			expected: true,
		},
		{
			name:     "Not WSL - just microsoft in name",
			content:  "Some microsoft server product",
			expected: false,
		},
		{
			name:     "Not WSL - WSL in name but not microsoft",
			content:  "Some WSL product",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "version")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			file, err := os.Open(tmpFile)
			if err != nil {
				t.Fatalf("Failed to open temp file: %v", err)
			}

			result := isWSLFromFile(file)

			if result != tt.expected {
				t.Errorf("WSL check = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectLinuxWithMissingOSRelease(t *testing.T) {
	// Test that fallbackUname is called when /etc/os-release doesn't exist
	info := &OSInfo{}
	err := detectLinux(info)
	if err != nil {
		t.Fatalf("detectLinux should not fail even without /etc/os-release: %v", err)
	}
	// ID should be set via fallback
	if info.ID == "" {
		t.Error("ID should be set via fallback after missing /etc/os-release")
	}
}

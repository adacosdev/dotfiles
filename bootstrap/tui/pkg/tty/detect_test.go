package tty

import (
	"os"
	"testing"
)

func TestIsTerminal(t *testing.T) {
	// In a non-TTY test environment, isatty returns false.
	// These tests verify the logic flow, not actual TTY detection.
	tests := []struct {
		name        string
		interactive bool
		json        bool
		term        string
		want        bool
	}{
		{
			name: "dumb terminal is not interactive",
			term: "dumb",
			want: false,
		},
		{
			name:        "force interactive overrides",
			interactive: true,
			want:        true,
		},
		{
			name: "force json disables interactive",
			json: true,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Reset()

			if tt.term != "" {
				os.Setenv("TERM", tt.term)
				defer os.Unsetenv("TERM")
			}
			if tt.interactive {
				SetForceInteractive(true)
			}
			if tt.json {
				SetForceJSON(true)
			}

			got := IsTerminal()
			if got != tt.want {
				t.Errorf("IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestForceInteractive(t *testing.T) {
	Reset()

	if ForceInteractive() {
		t.Error("ForceInteractive() = true, want false initially")
	}

	SetForceInteractive(true)
	if !ForceInteractive() {
		t.Error("ForceInteractive() = false, want true after SetForceInteractive")
	}

	Reset()
}

func TestForceJSON(t *testing.T) {
	Reset()

	if ForceJSON() {
		t.Error("ForceJSON() = true, want false initially")
	}

	SetForceJSON(true)
	if !ForceJSON() {
		t.Error("ForceJSON() = false, want true after SetForceJSON")
	}

	Reset()
}

func TestReset(t *testing.T) {
	SetForceInteractive(true)
	SetForceJSON(true)

	Reset()

	if ForceInteractive() {
		t.Error("ForceInteractive() = true after Reset, want false")
	}
	if ForceJSON() {
		t.Error("ForceJSON() = true after Reset, want false")
	}
}

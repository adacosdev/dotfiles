// Package tty provides terminal detection utilities.
package tty

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	forceInteractive bool
	forceJSON        bool
)

// IsTerminal returns true if stdout is a terminal.
func IsTerminal() bool {
	if forceJSON {
		return false
	}
	if forceInteractive {
		return true
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return isatty(os.Stdout.Fd())
}

// ForceInteractive forces interactive mode even if not a tty.
func ForceInteractive() bool {
	return forceInteractive
}

// ForceJSON forces JSON output mode.
func ForceJSON() bool {
	return forceJSON
}

// SetForceInteractive sets the force interactive flag (for testing).
func SetForceInteractive(v bool) {
	forceInteractive = v
}

// SetForceJSON sets the force JSON flag (for testing).
func SetForceJSON(v bool) {
	forceJSON = v
}

// Reset resets the global flags (for testing).
func Reset() {
	forceInteractive = false
	forceJSON = false
}

// isatty checks if the given file descriptor is a terminal.
func isatty(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

// ioctlReadTermios is the ioctl number for reading termios.
const ioctlReadTermios = 0x5401

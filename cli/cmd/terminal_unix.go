//go:build !windows

package cmd

import (
	"fmt"
)

// setupTerminal disables mouse reporting on Unix systems
func setupTerminal() error {
	// Disable all mouse reporting modes
	fmt.Print("\033[?1000l\033[?1002l\033[?1006l\033[?1015l")
	return nil
}

// restoreTerminal restores terminal state (placeholder for Unix)
func restoreTerminal() error {
	return nil
}

//go:build windows

package cmd

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// disableMouseMode disables mouse input for Windows Console
func disableMouseMode() error {
	// Get console handle
	handle := windows.Handle(os.Stdin.Fd())

	// Get current console mode
	var mode uint32
	err := windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return err
	}

	// Disable mouse input (ENABLE_MOUSE_INPUT = 0x0010)
	// Keep other flags enabled
	mode &^= windows.ENABLE_MOUSE_INPUT

	// Set the new mode
	err = windows.SetConsoleMode(handle, mode)
	if err != nil {
		return err
	}

	return nil
}

// enableMouseMode re-enables mouse input (for cleanup if needed)
func enableMouseMode() error {
	handle := windows.Handle(os.Stdin.Fd())

	var mode uint32
	err := windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return err
	}

	// Re-enable mouse input
	mode |= windows.ENABLE_MOUSE_INPUT

	err = windows.SetConsoleMode(handle, mode)
	if err != nil {
		return err
	}

	return nil
}

// setupTerminal disables mouse and prepares terminal for raw mode on Windows
func setupTerminal() error {
	// Try to disable mouse via Windows API
	if err := disableMouseMode(); err != nil {
		// Fallback to escape sequences (might work in some terminals like WT)
		fmt.Print("\033[?1000l\033[?1002l\033[?1006l\033[?1015l")
	}
	return nil
}

// restoreTerminal restores terminal state (placeholder for Windows)
func restoreTerminal() error {
	// We don't re-enable mouse here to avoid interference after exit
	return nil
}

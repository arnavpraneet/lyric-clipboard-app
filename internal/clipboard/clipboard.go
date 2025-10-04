package clipboard

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/atotto/clipboard"
)

// Manager handles clipboard operations
type Manager struct {
	useXClip bool
}

// NewManager creates a new clipboard manager
func NewManager() *Manager {
	m := &Manager{}

	// On Linux, check if xclip is available (for WSL compatibility)
	if runtime.GOOS == "linux" {
		if _, err := exec.LookPath("xclip"); err == nil {
			m.useXClip = true
		}
	}

	return m
}

// Write writes text to the system clipboard
func (m *Manager) Write(text string) error {
	// Try xclip first on Linux if available (WSL compatible)
	if m.useXClip {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = bytes.NewBufferString(text)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("xclip failed: %w", err)
		}
		return nil
	}

	// Fallback to atotto/clipboard
	if err := clipboard.WriteAll(text); err != nil {
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}
	return nil
}

// Read reads text from the system clipboard
func (m *Manager) Read() (string, error) {
	// Try xclip first on Linux if available
	if m.useXClip {
		cmd := exec.Command("xclip", "-selection", "clipboard", "-o")
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("xclip read failed: %w", err)
		}
		return string(output), nil
	}

	// Fallback to atotto/clipboard
	text, err := clipboard.ReadAll()
	if err != nil {
		return "", fmt.Errorf("failed to read from clipboard: %w", err)
	}
	return text, nil
}

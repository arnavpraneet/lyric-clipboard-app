package clipboard

import (
	"fmt"

	"github.com/atotto/clipboard"
)

// Manager handles clipboard operations
type Manager struct{}

// NewManager creates a new clipboard manager
func NewManager() *Manager {
	return &Manager{}
}

// Write writes text to the system clipboard
func (m *Manager) Write(text string) error {
	if err := clipboard.WriteAll(text); err != nil {
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}
	return nil
}

// Read reads text from the system clipboard
func (m *Manager) Read() (string, error) {
	text, err := clipboard.ReadAll()
	if err != nil {
		return "", fmt.Errorf("failed to read from clipboard: %w", err)
	}
	return text, nil
}

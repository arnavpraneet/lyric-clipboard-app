package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the application configuration
type Config struct {
	// General settings
	PollInterval time.Duration `json:"poll_interval"` // How often to check for song updates (in milliseconds)

	// Lyrics settings
	LyricOffset  time.Duration `json:"lyric_offset"`  // Time offset to apply to lyrics (in milliseconds)
	EnableCache  bool          `json:"enable_cache"`  // Enable lyrics caching

	// Clipboard settings
	UpdateClipboard bool `json:"update_clipboard"` // Enable clipboard updates

	// Demo mode settings
	DemoMode   bool   `json:"demo_mode"`   // Run in demo mode
	DemoArtist string `json:"demo_artist"` // Artist for demo mode
	DemoTitle  string `json:"demo_title"`  // Title for demo mode

	// GUI settings
	StartMinimized bool `json:"start_minimized"` // Start app minimized to system tray
	ShowNotifications bool `json:"show_notifications"` // Show notifications for song changes
}

// configFile represents the JSON structure for the config file
type configFile struct {
	PollIntervalMs    int    `json:"poll_interval_ms"`
	LyricOffsetMs     int    `json:"lyric_offset_ms"`
	EnableCache       bool   `json:"enable_cache"`
	UpdateClipboard   bool   `json:"update_clipboard"`
	DemoMode          bool   `json:"demo_mode"`
	DemoArtist        string `json:"demo_artist"`
	DemoTitle         string `json:"demo_title"`
	StartMinimized    bool   `json:"start_minimized"`
	ShowNotifications bool   `json:"show_notifications"`
}

// Default returns a Config with sensible default values
func Default() *Config {
	return &Config{
		PollInterval:      300 * time.Millisecond,
		LyricOffset:       0,
		EnableCache:       true,
		UpdateClipboard:   true,
		DemoMode:          false,
		DemoArtist:        "Rick Astley",
		DemoTitle:         "Never Gonna Give You Up",
		StartMinimized:    false,
		ShowNotifications: true,
	}
}

// Load loads configuration from the specified file
// If the file doesn't exist, returns default configuration
func Load(path string) (*Config, error) {
	// If path is empty, use default location
	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return nil, fmt.Errorf("failed to get default config path: %w", err)
		}
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist, return defaults
		return Default(), nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var cf configFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Convert to Config
	config := &Config{
		PollInterval:      time.Duration(cf.PollIntervalMs) * time.Millisecond,
		LyricOffset:       time.Duration(cf.LyricOffsetMs) * time.Millisecond,
		EnableCache:       cf.EnableCache,
		UpdateClipboard:   cf.UpdateClipboard,
		DemoMode:          cf.DemoMode,
		DemoArtist:        cf.DemoArtist,
		DemoTitle:         cf.DemoTitle,
		StartMinimized:    cf.StartMinimized,
		ShowNotifications: cf.ShowNotifications,
	}

	// Apply defaults for zero values
	if config.PollInterval == 0 {
		config.PollInterval = 300 * time.Millisecond
	}
	if config.DemoArtist == "" {
		config.DemoArtist = "Rick Astley"
	}
	if config.DemoTitle == "" {
		config.DemoTitle = "Never Gonna Give You Up"
	}

	return config, nil
}

// Save saves the configuration to the specified file
func (c *Config) Save(path string) error {
	// If path is empty, use default location
	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get default config path: %w", err)
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Convert to configFile
	cf := configFile{
		PollIntervalMs:    int(c.PollInterval.Milliseconds()),
		LyricOffsetMs:     int(c.LyricOffset.Milliseconds()),
		EnableCache:       c.EnableCache,
		UpdateClipboard:   c.UpdateClipboard,
		DemoMode:          c.DemoMode,
		DemoArtist:        c.DemoArtist,
		DemoTitle:         c.DemoTitle,
		StartMinimized:    c.StartMinimized,
		ShowNotifications: c.ShowNotifications,
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfigPath returns the default configuration file path
func DefaultConfigPath() (string, error) {
	// Get user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// Create app-specific directory path
	appConfigDir := filepath.Join(configDir, "lyric-clipboard")
	return filepath.Join(appConfigDir, "config.json"), nil
}

// GenerateExample generates an example configuration file at the default location
func GenerateExample() error {
	path, err := DefaultConfigPath()
	if err != nil {
		return err
	}

	config := Default()
	return config.Save(path)
}

//go:build windows

package detector

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// WindowsDetector uses PowerShell to access Windows Media Transport Controls
type WindowsDetector struct {
	lastError error
}

// NewDetector creates a new Windows detector
func NewDetector() (Detector, error) {
	return &WindowsDetector{}, nil
}

// mediaResult represents the JSON output from PowerShell
type mediaResult struct {
	Artist    string  `json:"artist"`
	Title     string  `json:"title"`
	Album     string  `json:"album"`
	Position  float64 `json:"position"`  // Position in seconds
	Duration  float64 `json:"duration"`  // Duration in seconds
	IsPlaying bool    `json:"isPlaying"`
}

// GetCurrentSong retrieves the currently playing song from Windows Media Transport Controls
func (d *WindowsDetector) GetCurrentSong() (*SongInfo, error) {
	// PowerShell script to access GlobalSystemMediaTransportControlsSessionManager
	script := `
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$null = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager, Windows.Media.Control, ContentType = WindowsRuntime]
$null = [Windows.Media.Control.GlobalSystemMediaTransportControlsSession, Windows.Media.Control, ContentType = WindowsRuntime]

$sessionManager = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager]::RequestAsync()
$sessionManager.AsTask().GetAwaiter().GetResult()

$session = $sessionManager.GetCurrentSession()
if ($null -eq $session) {
    Write-Output "{}"
    exit
}

$mediaProps = $session.TryGetMediaPropertiesAsync()
$mediaProps.AsTask().GetAwaiter().GetResult()

$timelineProps = $session.GetTimelineProperties()
$playbackInfo = $session.GetPlaybackInfo()

$position = 0
$duration = 0
if ($null -ne $timelineProps) {
    $position = $timelineProps.Position.TotalSeconds
    $duration = $timelineProps.EndTime.TotalSeconds
}

$isPlaying = $false
if ($null -ne $playbackInfo) {
    $isPlaying = $playbackInfo.PlaybackStatus -eq 4  # 4 = Playing
}

$result = @{
    artist = $mediaProps.Artist
    title = $mediaProps.Title
    album = $mediaProps.AlbumTitle
    position = $position
    duration = $duration
    isPlaying = $isPlaying
}

ConvertTo-Json $result
`

	// Execute PowerShell script
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		d.lastError = fmt.Errorf("failed to execute PowerShell: %w", err)
		return nil, d.lastError
	}

	// Parse JSON output
	var result mediaResult
	if err := json.Unmarshal(output, &result); err != nil {
		d.lastError = fmt.Errorf("failed to parse media info: %w", err)
		return nil, d.lastError
	}

	// Check if we got valid data
	if result.Title == "" {
		return nil, nil // No song playing
	}

	// Convert to SongInfo
	songInfo := &SongInfo{
		Artist:    result.Artist,
		Title:     result.Title,
		Album:     result.Album,
		Position:  time.Duration(result.Position * float64(time.Second)),
		IsPlaying: result.IsPlaying,
	}

	return songInfo, nil
}

// Close cleans up resources (no-op for Windows detector)
func (d *WindowsDetector) Close() error {
	return nil
}

// String returns a string representation of the detector
func (d *WindowsDetector) String() string {
	return "WindowsDetector (PowerShell-based)"
}

// getErrorDetails returns additional error information if available
func (d *WindowsDetector) getErrorDetails() string {
	if d.lastError != nil {
		return d.lastError.Error()
	}
	return "no error"
}

// isAvailable checks if the detector can run on this system
func isAvailable() bool {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Write-Output 'test'")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "test"
}

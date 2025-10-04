//go:build linux

package detector

import (
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

// LinuxDetector implements song detection using D-Bus MPRIS on Linux
type LinuxDetector struct {
	conn *dbus.Conn
}

// NewDetector creates a new platform-specific detector
func NewDetector() (Detector, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}

	return &LinuxDetector{
		conn: conn,
	}, nil
}

// GetCurrentSong retrieves the currently playing song from MPRIS-compatible players
func (d *LinuxDetector) GetCurrentSong() (*SongInfo, error) {
	// List of common media players to check
	players := []string{
		"org.mpris.MediaPlayer2.spotify",
		"org.mpris.MediaPlayer2.vlc",
		"org.mpris.MediaPlayer2.rhythmbox",
		"org.mpris.MediaPlayer2.chromium",
	}

	for _, player := range players {
		info, err := d.getPlayerInfo(player)
		if err == nil && info != nil {
			return info, nil
		}
	}

	return nil, fmt.Errorf("no active media player found")
}

func (d *LinuxDetector) getPlayerInfo(serviceName string) (*SongInfo, error) {
	obj := d.conn.Object(serviceName, "/org/mpris/MediaPlayer2")

	// Get playback status
	statusVariant, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
	if err != nil {
		return nil, err
	}

	status, ok := statusVariant.Value().(string)
	if !ok || status != "Playing" {
		return nil, fmt.Errorf("player not playing")
	}

	// Get metadata
	metadataVariant, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
	if err != nil {
		return nil, err
	}

	metadata, ok := metadataVariant.Value().(map[string]dbus.Variant)
	if !ok {
		return nil, fmt.Errorf("invalid metadata format")
	}

	// Extract song information
	info := &SongInfo{
		IsPlaying: true,
	}

	if title, ok := metadata["xesam:title"].Value().(string); ok {
		info.Title = title
	}

	if artists, ok := metadata["xesam:artist"].Value().([]string); ok && len(artists) > 0 {
		info.Artist = artists[0]
	}

	if album, ok := metadata["xesam:album"].Value().(string); ok {
		info.Album = album
	}

	// Get playback position
	positionVariant, err := obj.GetProperty("org.mpris.MediaPlayer2.Player.Position")
	if err == nil {
		if pos, ok := positionVariant.Value().(int64); ok {
			// Position is in microseconds, convert to duration
			info.Position = time.Duration(pos) * time.Microsecond
		}
	}

	// Validate we have at least artist and title
	if info.Artist == "" || info.Title == "" {
		return nil, fmt.Errorf("incomplete song information")
	}

	return info, nil
}

// Close closes the D-Bus connection
func (d *LinuxDetector) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

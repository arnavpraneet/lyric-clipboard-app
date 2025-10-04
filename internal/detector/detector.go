package detector

import "time"

// SongInfo represents the currently playing song and its state
type SongInfo struct {
	Artist    string
	Title     string
	Album     string
	Position  time.Duration // Current playback position
	IsPlaying bool
}

// Detector is the interface for platform-specific song detection
type Detector interface {
	// GetCurrentSong returns the currently playing song information
	// Returns nil if no song is playing or if detection fails
	GetCurrentSong() (*SongInfo, error)

	// Close cleans up any resources used by the detector
	Close() error
}

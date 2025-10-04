package detector

import (
	"fmt"
	"time"
)

// DemoDetector simulates a playing song for testing purposes
type DemoDetector struct {
	startTime time.Time
	artist    string
	title     string
}

// NewDemoDetector creates a detector that simulates a playing song
func NewDemoDetector(artist, title string) Detector {
	return &DemoDetector{
		startTime: time.Now(),
		artist:    artist,
		title:     title,
	}
}

// GetCurrentSong returns simulated song information
func (d *DemoDetector) GetCurrentSong() (*SongInfo, error) {
	// Calculate elapsed time since start
	elapsed := time.Since(d.startTime)

	return &SongInfo{
		Artist:    d.artist,
		Title:     d.title,
		Album:     "Demo Album",
		Position:  elapsed,
		IsPlaying: true,
	}, nil
}

// Close is a no-op for demo detector
func (d *DemoDetector) Close() error {
	fmt.Println("Demo detector closed")
	return nil
}

package orchestrator

import (
	"fmt"
	"log"
	"time"

	"github.com/arnavpraneet/lyric-clipboard-app/internal/clipboard"
	"github.com/arnavpraneet/lyric-clipboard-app/internal/detector"
	"github.com/arnavpraneet/lyric-clipboard-app/internal/lyrics"
)

// Orchestrator is the core component that coordinates all modules
type Orchestrator struct {
	detector          detector.Detector
	lyricsFetcher     *lyrics.Fetcher
	clipboardMgr      *clipboard.Manager
	pollInterval      time.Duration
	lyricOffset       time.Duration
	updateClipboard   bool
	currentSongKey    string
	currentLyrics     *lyrics.SyncedLyrics
	lastLyricText     string
	stopChan          chan struct{}
	statusCallback    func(status string)
}

// Config holds configuration for the orchestrator
type Config struct {
	PollInterval    time.Duration // How often to check for song updates
	LyricOffset     time.Duration // Time offset to apply to lyrics
	UpdateClipboard bool          // Enable clipboard updates
	DemoMode        bool          // Run in demo mode
	DemoArtist      string        // Artist for demo mode
	DemoTitle       string        // Title for demo mode
}

// NewOrchestrator creates a new orchestrator with the given configuration
func NewOrchestrator(config Config) (*Orchestrator, error) {
	var det detector.Detector
	var err error

	if config.DemoMode {
		det = detector.NewDemoDetector(config.DemoArtist, config.DemoTitle)
	} else {
		det, err = detector.NewDetector()
		if err != nil {
			return nil, fmt.Errorf("failed to create detector: %w", err)
		}
	}

	return &Orchestrator{
		detector:        det,
		lyricsFetcher:   lyrics.NewFetcher(),
		clipboardMgr:    clipboard.NewManager(),
		pollInterval:    config.PollInterval,
		lyricOffset:     config.LyricOffset,
		updateClipboard: config.UpdateClipboard,
		stopChan:        make(chan struct{}),
	}, nil
}

// Start begins the orchestrator's main loop
func (o *Orchestrator) Start() {
	log.Println("Starting Lyric Clipboard App...")
	ticker := time.NewTicker(o.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.tick()
		case <-o.stopChan:
			log.Println("Stopping orchestrator...")
			return
		}
	}
}

// tick performs one iteration of the main loop
func (o *Orchestrator) tick() {
	// Get current song
	songInfo, err := o.detector.GetCurrentSong()
	if err != nil {
		// No song playing or detection failed - clear state
		if o.currentSongKey != "" {
			log.Println("No song detected, clearing state")
			o.currentSongKey = ""
			o.currentLyrics = nil
			o.lastLyricText = ""
		}
		return
	}

	// Create a unique key for this song
	songKey := fmt.Sprintf("%s - %s", songInfo.Artist, songInfo.Title)

	// Check if this is a new song
	if songKey != o.currentSongKey {
		log.Printf("New song detected: %s", songKey)
		o.currentSongKey = songKey
		o.lastLyricText = ""

		// Fetch lyrics for the new song
		lyrics, err := o.lyricsFetcher.FetchLyrics(songInfo.Artist, songInfo.Title)
		if err != nil {
			log.Printf("Failed to fetch lyrics for %s: %v", songKey, err)
			o.currentLyrics = nil
			return
		}

		o.currentLyrics = lyrics
		log.Printf("Lyrics fetched successfully (%d lines)", len(lyrics.Lines))
	}

	// If we don't have lyrics, nothing to do
	if o.currentLyrics == nil {
		return
	}

	// Apply lyric offset to playback position
	adjustedPosition := songInfo.Position + o.lyricOffset

	// Get the current lyric line based on adjusted playback position
	currentLine := o.currentLyrics.GetLineAtTime(adjustedPosition)
	if currentLine == nil {
		return
	}

	// Update clipboard if the lyric has changed
	if currentLine.Text != o.lastLyricText {
		log.Printf("[%s] %s", formatDuration(songInfo.Position), currentLine.Text)

		if o.updateClipboard {
			if err := o.clipboardMgr.Write(currentLine.Text); err != nil {
				log.Printf("Failed to update clipboard: %v", err)
				return
			}
		}

		o.lastLyricText = currentLine.Text

		// Notify status callback if set
		if o.statusCallback != nil {
			o.statusCallback(currentLine.Text)
		}
	}
}

// Stop stops the orchestrator
func (o *Orchestrator) Stop() {
	close(o.stopChan)
	if o.detector != nil {
		o.detector.Close()
	}
}

// SetStatusCallback sets a callback function for status updates
func (o *Orchestrator) SetStatusCallback(callback func(status string)) {
	o.statusCallback = callback
}

// GetCurrentStatus returns the current playback status
func (o *Orchestrator) GetCurrentStatus() string {
	if o.currentSongKey == "" {
		return "No song detected"
	}
	if o.lastLyricText == "" {
		return fmt.Sprintf("Playing: %s", o.currentSongKey)
	}
	return fmt.Sprintf("%s: %s", o.currentSongKey, o.lastLyricText)
}

// SetLyricOffset updates the lyric offset dynamically
func (o *Orchestrator) SetLyricOffset(offset time.Duration) {
	o.lyricOffset = offset
	log.Printf("Lyric offset updated to %v", offset)
}

// SetUpdateClipboard enables or disables clipboard updates
func (o *Orchestrator) SetUpdateClipboard(enabled bool) {
	o.updateClipboard = enabled
	if enabled {
		log.Println("Clipboard updates enabled")
	} else {
		log.Println("Clipboard updates disabled")
	}
}

// formatDuration formats a duration as mm:ss
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

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
	detector        detector.Detector
	lyricsFetcher   *lyrics.Fetcher
	clipboardMgr    *clipboard.Manager
	pollInterval    time.Duration
	currentSongKey  string
	currentLyrics   *lyrics.SyncedLyrics
	lastLyricText   string
	stopChan        chan struct{}
}

// Config holds configuration for the orchestrator
type Config struct {
	PollInterval time.Duration // How often to check for song updates
}

// NewOrchestrator creates a new orchestrator with the given configuration
func NewOrchestrator(config Config) (*Orchestrator, error) {
	det, err := detector.NewDetector()
	if err != nil {
		return nil, fmt.Errorf("failed to create detector: %w", err)
	}

	return &Orchestrator{
		detector:      det,
		lyricsFetcher: lyrics.NewFetcher(),
		clipboardMgr:  clipboard.NewManager(),
		pollInterval:  config.PollInterval,
		stopChan:      make(chan struct{}),
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

	// Get the current lyric line based on playback position
	currentLine := o.currentLyrics.GetLineAtTime(songInfo.Position)
	if currentLine == nil {
		return
	}

	// Update clipboard if the lyric has changed
	if currentLine.Text != o.lastLyricText {
		log.Printf("[%s] %s", formatDuration(songInfo.Position), currentLine.Text)

		if err := o.clipboardMgr.Write(currentLine.Text); err != nil {
			log.Printf("Failed to update clipboard: %v", err)
			return
		}

		o.lastLyricText = currentLine.Text
	}
}

// Stop stops the orchestrator
func (o *Orchestrator) Stop() {
	close(o.stopChan)
	if o.detector != nil {
		o.detector.Close()
	}
}

// formatDuration formats a duration as mm:ss
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

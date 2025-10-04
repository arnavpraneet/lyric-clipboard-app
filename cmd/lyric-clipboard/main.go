package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arnavpraneet/lyric-clipboard-app/internal/orchestrator"
)

func main() {
	// Command-line flags
	demoMode := flag.Bool("demo", false, "Run in demo mode with a sample song")
	demoArtist := flag.String("artist", "Rick Astley", "Artist name for demo mode")
	demoTitle := flag.String("title", "Never Gonna Give You Up", "Song title for demo mode")
	flag.Parse()

	// Create orchestrator with configuration
	config := orchestrator.Config{
		PollInterval: 300 * time.Millisecond, // Poll every 300ms
		DemoMode:     *demoMode,
		DemoArtist:   *demoArtist,
		DemoTitle:    *demoTitle,
	}

	orch, err := orchestrator.NewOrchestrator(config)
	if err != nil {
		log.Fatalf("Failed to create orchestrator: %v", err)
	}

	if *demoMode {
		log.Printf("Running in DEMO mode with: %s - %s", *demoArtist, *demoTitle)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start orchestrator in a goroutine
	go orch.Start()

	// Wait for shutdown signal
	<-sigChan
	log.Println("\nReceived shutdown signal...")

	// Stop orchestrator
	orch.Stop()

	log.Println("Goodbye!")
}

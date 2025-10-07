package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arnavpraneet/lyric-clipboard-app/internal/config"
	"github.com/arnavpraneet/lyric-clipboard-app/internal/orchestrator"
)

func main() {
	// Command-line flags
	configPath := flag.String("config", "", "Path to configuration file (default: ~/.config/lyric-clipboard/config.json)")
	demoMode := flag.Bool("demo", false, "Run in demo mode with a sample song")
	demoArtist := flag.String("artist", "", "Artist name for demo mode")
	demoTitle := flag.String("title", "", "Song title for demo mode")
	generateConfig := flag.Bool("generate-config", false, "Generate example configuration file and exit")
	flag.Parse()

	// Generate config if requested
	if *generateConfig {
		if err := config.GenerateExample(); err != nil {
			log.Fatalf("Failed to generate config: %v", err)
		}
		path, _ := config.DefaultConfigPath()
		log.Printf("Configuration file generated at: %s", path)
		return
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override config with command-line flags
	if *demoMode {
		cfg.DemoMode = true
	}
	if *demoArtist != "" {
		cfg.DemoArtist = *demoArtist
	}
	if *demoTitle != "" {
		cfg.DemoTitle = *demoTitle
	}

	// Create orchestrator with configuration
	orchConfig := orchestrator.Config{
		PollInterval:    cfg.PollInterval,
		LyricOffset:     cfg.LyricOffset,
		UpdateClipboard: cfg.UpdateClipboard,
		DemoMode:        cfg.DemoMode,
		DemoArtist:      cfg.DemoArtist,
		DemoTitle:       cfg.DemoTitle,
	}

	orch, err := orchestrator.NewOrchestrator(orchConfig)
	if err != nil {
		log.Fatalf("Failed to create orchestrator: %v", err)
	}

	if cfg.DemoMode {
		log.Printf("Running in DEMO mode with: %s - %s", cfg.DemoArtist, cfg.DemoTitle)
	}
	if cfg.LyricOffset != 0 {
		log.Printf("Lyric offset: %v", cfg.LyricOffset)
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

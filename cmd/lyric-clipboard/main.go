package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arnavpraneet/lyric-clipboard-app/internal/orchestrator"
)

func main() {
	// Create orchestrator with configuration
	config := orchestrator.Config{
		PollInterval: 300 * time.Millisecond, // Poll every 300ms
	}

	orch, err := orchestrator.NewOrchestrator(config)
	if err != nil {
		log.Fatalf("Failed to create orchestrator: %v", err)
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

package main

import (
	"context"
	"homecourt-api/games"
	"homecourt-api/receiver"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatalf("err loading: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go receiver.Receiver(ctx)

	server := &http.Server{Addr: ":8080"}
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	gamesManager, err := games.NewGamesManager(":6379")
	receiver.Manager = gamesManager
	// Listen for OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Println("Shutdown signal received")

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}

	// Cancel the context to stop the Receiver
	cancel()

	// Optionally, wait for the Receiver to finish
	// For example, using a sync.WaitGroup (not shown here)
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"homecourt-api/games"
	"homecourt-api/handlers"
	"homecourt-api/receiver"

	)

// enableCORS sets the necessary CORS headers and handles preflight requests.
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins (adjust as needed)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load environment variables from .env.local
	
	// Initialize GamesManager with Redis
	gamesManager, err := games.NewGamesManager("localhost:6379")
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	// Assign the GamesManager to receiver and handlers
	receiver.Manager = gamesManager
	handlers.Manager = gamesManager

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the receiver in a separate goroutine
	go receiver.Receiver(ctx)

	// Create a new ServeMux and register handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/get", handlers.GetHandler)

	// Wrap the mux with CORS middleware
	handlerWithCORS := enableCORS(mux)

	// Initialize the HTTP server with the wrapped handler
	server := &http.Server{
		Addr:    ":8080",
		Handler: handlerWithCORS,
	}

	// Start the server in a separate goroutine
	go func() {
		log.Println("Starting HTTP server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Listen for OS signals for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Println("Shutdown signal received")

	// Create a context with timeout for the shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Attempt to gracefully shutdown the server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	} else {
		log.Println("HTTP server gracefully stopped")
	}

	// Cancel the main context to stop the Receiver
	cancel()

	log.Println("Server shutdown complete")
}

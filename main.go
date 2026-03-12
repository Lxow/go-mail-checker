package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lxow/go-mail-checker/api"
)

func main() {
	// Because every good API needs a fancy ASCII art welcome message
	fmt.Println(`
	   _____ ____    __  __       _ _    _____ _               _             
	  / ____/ __ \  |  \/  |     (_) |  / ____| |             | |            
	 | |   | |  | | | \  / | __ _ _| | | |    | |__   ___  ___| | _____ _ __ 
	 | |   | |  | | | |\/| |/ _` + "`" + ` | | | | |    | '_ \ / _ \/ __| |/ / _ \ '__|
	 | |___| |__| | | |  | | (_| | | | | |____| | | |  __/ (__|   <  __/ |   
	  \_____\____/  |_|  |_|\__,_|_|_|  \_____|_| |_|\___|\___|_|\_\___|_|   
	
	Starting Go Mail Checker API on port 8080...
	`)

	// Set up HTTP routes - keeping it simple, no fancy frameworks needed
	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.HealthHandler)
	mux.HandleFunc("/check-domain", api.CheckDomainHandler)

	// Create server with reasonable timeouts (nobody likes hanging connections...)
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown setup cause crashing is not professional
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on http://localhost:8080")
		log.Printf("Health check: curl http://localhost:8080/health")
		log.Printf("Domain check: curl http://localhost:8080/check-domain?domain=gmail.com")
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-done
	log.Print("Server shutting down gracefully...")

	// Give outstanding requests 30 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Print("Server exited successfully. Bye!")
}
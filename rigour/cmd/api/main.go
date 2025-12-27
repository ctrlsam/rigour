package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ctrlsam/rigour/internal/api"
	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/internal/storage/mongodb"
	"github.com/spf13/cobra"
)

type cliConfig struct {
	mongoURI   string
	database   string
	collection string
	addr       string
}

var config cliConfig

var rootCmd = &cobra.Command{
	Use:   "rigour-api",
	Short: "REST API server for Rigour",
	Long:  "A REST API server for querying scanned hosts and services from MongoDB",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer(cmd.Context())
	},
}

func init() {
	rootCmd.Flags().StringVar(&config.mongoURI, "mongo-uri", "mongodb://localhost:27017", "MongoDB connection URI")
	rootCmd.Flags().StringVar(&config.database, "mongo-db", "rigour", "MongoDB database name")
	rootCmd.Flags().StringVar(&config.collection, "mongo-collection", "hosts", "MongoDB collection name")
	rootCmd.Flags().StringVar(&config.addr, "addr", ":8080", "Server address (host:port)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runServer(ctx context.Context) error {
	// Validate inputs
	if config.mongoURI == "" {
		return fmt.Errorf("mongo-uri is required")
	}

	// Create MongoDB client
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	mongoClient, err := mongodb.NewClient(connectCtx, config.mongoURI, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer mongoClient.Close(context.Background())

	// Create repository
	repository, err := mongoClient.NewRepository(connectCtx, storage.RepositoryConfig{
		Database:   config.database,
		Collection: config.collection,
		Timeout:    10,
	})
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	// Create router and handler
	router := api.NewRouter(repository)

	// Setup HTTP server
	server := &http.Server{
		Addr:         config.addr,
		Handler:      router.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("Starting API server on %s\n", config.addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error: server failed: %v\n", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	fmt.Println("\nShutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	fmt.Println("Server stopped")
	return nil
}

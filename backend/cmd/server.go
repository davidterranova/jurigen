package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	pkg "davidterranova/jurigen/backend/internal"
	"davidterranova/jurigen/backend/internal/adapter/http"
	"davidterranova/jurigen/backend/internal/port"
	"davidterranova/jurigen/backend/pkg/xhttp"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

// Server configuration flags
var (
	dagPath        string
	writeThrough   bool
	syncOnShutdown bool
	address        string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP API server with hybrid DAG repository (file + memory)",
	Long: `Starts the HTTP API server using the HybridDAGRepository for optimal performance.

The hybrid repository:
- Loads DAGs from files at startup for persistence
- Serves DAGs from memory for fast runtime access  
- Optionally writes changes back to files for durability
- Provides statistics and sync capabilities`,
	Example: `  # Start server with write-through enabled (changes immediately persisted)
  jurigen server --dag-path ./data --write-through

  # Start server with write-through disabled (manual sync required)
  jurigen server --dag-path ./data --write-through=false

  # Start server with custom address and sync-on-shutdown
  jurigen server --dag-path ./data --address :8081 --sync-on-shutdown`,
	RunE: runServer,
}

func runServer(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Set up structured logging
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("component", "server").
		Logger()

	logger.Info().
		Str("dag_path", dagPath).
		Bool("write_through", writeThrough).
		Bool("sync_on_shutdown", syncOnShutdown).
		Str("address", address).
		Msg("Starting server with hybrid DAG repository")

	// Create hybrid repository
	hybridRepo := port.NewHybridDAGRepository(port.HybridDAGRepositoryConfig{
		FilePath:     dagPath,
		WriteThrough: writeThrough,
		Logger:       &logger,
	})

	// Initialize repository (load DAGs from files into memory)
	logger.Info().Msg("Initializing hybrid repository...")
	err := hybridRepo.Initialize(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize hybrid repository")
		return fmt.Errorf("failed to initialize hybrid repository: %w", err)
	}

	// Display repository statistics
	stats, err := hybridRepo.GetStats(ctx)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to get repository stats")
	} else {
		logger.Info().
			Int("memory_dags", stats.MemoryDAGCount).
			Int("file_dags", stats.FileDAGCount).
			Bool("write_through", stats.WriteThrough).
			Msg("Repository initialized successfully")
	}

	// Create application layer
	appLayer := pkg.New(hybridRepo)

	// Parse address to extract host and port
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		logger.Error().Err(err).Str("address", address).Msg("Invalid server address format")
		return fmt.Errorf("invalid server address format: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Error().Err(err).Str("port", portStr).Msg("Invalid port number")
		return fmt.Errorf("invalid port number: %w", err)
	}

	// Create HTTP server
	router := http.New(appLayer, nil) // No authentication for now
	server := xhttp.NewServer(router, host, port)

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle shutdown signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		logger.Info().Str("address", server.Address()).Msg("HTTP server starting")
		if err := server.Serve(ctx); err != nil {
			serverErrChan <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-signalChan:
		logger.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
		cancel()

		// Perform graceful shutdown
		if syncOnShutdown && !writeThrough {
			logger.Info().Msg("Syncing in-memory DAGs to files before shutdown...")
			if err := hybridRepo.Sync(ctx); err != nil {
				logger.Error().Err(err).Msg("Failed to sync DAGs to files during shutdown")
			} else {
				logger.Info().Msg("Successfully synced DAGs to files")
			}
		}

		logger.Info().Msg("Server shutdown completed")

	case err := <-serverErrChan:
		logger.Error().Err(err).Msg("Server error")
		cancel()
		return err
	}

	return nil
}

func init() {
	// Add the command to the root command
	rootCmd.AddCommand(serverCmd)

	// Add configuration flags
	serverCmd.Flags().StringVar(&dagPath, "dag-path", "data", "Directory path for DAG files")
	serverCmd.Flags().BoolVar(&writeThrough, "write-through", true, "Enable write-through to files (immediate persistence)")
	serverCmd.Flags().BoolVar(&syncOnShutdown, "sync-on-shutdown", true, "Sync in-memory DAGs to files on graceful shutdown")
	serverCmd.Flags().StringVar(&address, "address", ":8080", "Server address (host:port)")
}

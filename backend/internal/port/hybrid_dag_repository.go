package port

import (
	"context"
	"davidterranova/jurigen/backend/internal/dag"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// HybridDAGRepository combines file-based persistence with in-memory performance
// It loads DAGs from files at startup and serves them from memory for fast access
// Changes are persisted back to files for durability
type HybridDAGRepository struct {
	fileRepo   *FileDAGRepository
	memoryRepo *InMemoryDAGRepository
	logger     zerolog.Logger
	// writeThrough determines if changes are immediately persisted to file
	writeThrough bool
}

// HybridDAGRepositoryConfig configures the hybrid repository behavior
type HybridDAGRepositoryConfig struct {
	FilePath     string
	WriteThrough bool            // If true, changes are immediately persisted to file
	Logger       *zerolog.Logger // Optional logger, if nil a default will be created
}

// NewHybridDAGRepository creates a new hybrid repository
func NewHybridDAGRepository(config HybridDAGRepositoryConfig) *HybridDAGRepository {
	fileRepo := NewFileDAGRepository(config.FilePath)
	memoryRepo := NewInMemoryDAGRepository()

	var logger zerolog.Logger
	if config.Logger == nil {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		logger = *config.Logger
	}

	return &HybridDAGRepository{
		fileRepo:     fileRepo,
		memoryRepo:   memoryRepo,
		logger:       logger,
		writeThrough: config.WriteThrough,
	}
}

// Initialize loads all DAGs from the file repository into memory
// This should be called once during application startup
func (r *HybridDAGRepository) Initialize(ctx context.Context) error {
	r.logger.Info().Msg("Initializing hybrid DAG repository: loading DAGs from files into memory")

	// List all DAGs from file system
	dagIds, err := r.fileRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list DAGs from file repository: %w", err)
	}

	r.logger.Info().Int("count", len(dagIds)).Msg("Found DAGs in file system")

	// Load each DAG from file into memory
	loadedCount := 0
	for _, dagId := range dagIds {
		dagObj, err := r.fileRepo.Get(ctx, dagId)
		if err != nil {
			r.logger.Warn().
				Str("dag_id", dagId.String()).
				Err(err).
				Msg("Failed to load DAG from file, skipping")
			continue
		}

		// Store in memory repository
		err = r.memoryRepo.Create(ctx, dagObj)
		if err != nil {
			r.logger.Warn().
				Str("dag_id", dagId.String()).
				Err(err).
				Msg("Failed to store DAG in memory, skipping")
			continue
		}

		loadedCount++
	}

	r.logger.Info().
		Int("total_found", len(dagIds)).
		Int("successfully_loaded", loadedCount).
		Msg("DAG repository initialization completed")

	return nil
}

// Sync persists all in-memory DAGs back to the file system
// Useful for batch persistence or shutdown procedures
func (r *HybridDAGRepository) Sync(ctx context.Context) error {
	r.logger.Info().Msg("Syncing in-memory DAGs to file system")

	// Get all DAGs from memory
	dagIds, err := r.memoryRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list DAGs from memory: %w", err)
	}

	syncedCount := 0
	for _, dagId := range dagIds {
		dagObj, err := r.memoryRepo.Get(ctx, dagId)
		if err != nil {
			r.logger.Warn().
				Str("dag_id", dagId.String()).
				Err(err).
				Msg("Failed to get DAG from memory during sync")
			continue
		}

		// Try to get from file first to determine if it's create or update
		_, err = r.fileRepo.Get(ctx, dagId)
		if err != nil {
			// DAG doesn't exist in file, create it
			err = r.fileRepo.Create(ctx, dagObj)
		} else {
			// DAG exists in file, update it
			err = r.fileRepo.Update(ctx, dagId, func(existing dag.DAG) (dag.DAG, error) {
				return *dagObj, nil
			})
		}

		if err != nil {
			r.logger.Warn().
				Str("dag_id", dagId.String()).
				Err(err).
				Msg("Failed to sync DAG to file")
			continue
		}

		syncedCount++
	}

	r.logger.Info().
		Int("total_dags", len(dagIds)).
		Int("successfully_synced", syncedCount).
		Msg("DAG sync completed")

	return nil
}

// List returns all DAG IDs from memory (fast operation)
func (r *HybridDAGRepository) List(ctx context.Context) ([]uuid.UUID, error) {
	return r.memoryRepo.List(ctx)
}

// Get retrieves a DAG from memory (fast operation)
func (r *HybridDAGRepository) Get(ctx context.Context, id uuid.UUID) (*dag.DAG, error) {
	return r.memoryRepo.Get(ctx, id)
}

// Create stores a DAG in memory and optionally persists to file
func (r *HybridDAGRepository) Create(ctx context.Context, dagObj *dag.DAG) error {
	// Store in memory first
	err := r.memoryRepo.Create(ctx, dagObj)
	if err != nil {
		return fmt.Errorf("failed to create DAG in memory: %w", err)
	}

	// Persist to file if write-through is enabled
	if r.writeThrough {
		err = r.fileRepo.Create(ctx, dagObj)
		if err != nil {
			// Rollback memory operation on file failure
			deleteErr := r.memoryRepo.Delete(ctx, dagObj.Id)
			if deleteErr != nil {
				r.logger.Error().
					Str("dag_id", dagObj.Id.String()).
					Err(deleteErr).
					Msg("Failed to rollback memory operation after file create failure")
			}
			return fmt.Errorf("failed to create DAG in file (memory rolled back): %w", err)
		}

		r.logger.Debug().
			Str("dag_id", dagObj.Id.String()).
			Msg("DAG created in both memory and file")
	} else {
		r.logger.Debug().
			Str("dag_id", dagObj.Id.String()).
			Msg("DAG created in memory (write-through disabled)")
	}

	return nil
}

// Update modifies a DAG in memory and optionally persists to file
func (r *HybridDAGRepository) Update(ctx context.Context, id uuid.UUID, fnUpdate func(dag dag.DAG) (dag.DAG, error)) error {
	// Update in memory first
	err := r.memoryRepo.Update(ctx, id, fnUpdate)
	if err != nil {
		return fmt.Errorf("failed to update DAG in memory: %w", err)
	}

	// Persist to file if write-through is enabled
	if r.writeThrough {
		err = r.fileRepo.Update(ctx, id, fnUpdate)
		if err != nil {
			r.logger.Error().
				Str("dag_id", id.String()).
				Err(err).
				Msg("Failed to update DAG in file, memory and file are now inconsistent")
			return fmt.Errorf("failed to update DAG in file: %w", err)
		}

		r.logger.Debug().
			Str("dag_id", id.String()).
			Msg("DAG updated in both memory and file")
	} else {
		r.logger.Debug().
			Str("dag_id", id.String()).
			Msg("DAG updated in memory (write-through disabled)")
	}

	return nil
}

// Delete removes a DAG from memory and optionally from file
func (r *HybridDAGRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Delete from memory first
	err := r.memoryRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete DAG from memory: %w", err)
	}

	// Delete from file if write-through is enabled
	if r.writeThrough {
		err = r.fileRepo.Delete(ctx, id)
		if err != nil {
			r.logger.Error().
				Str("dag_id", id.String()).
				Err(err).
				Msg("Failed to delete DAG from file, memory and file are now inconsistent")
			return fmt.Errorf("failed to delete DAG from file: %w", err)
		}

		r.logger.Debug().
			Str("dag_id", id.String()).
			Msg("DAG deleted from both memory and file")
	} else {
		r.logger.Debug().
			Str("dag_id", id.String()).
			Msg("DAG deleted from memory (write-through disabled)")
	}

	return nil
}

// GetStats returns statistics about the repository state
func (r *HybridDAGRepository) GetStats(ctx context.Context) (HybridRepositoryStats, error) {
	memoryIds, err := r.memoryRepo.List(ctx)
	if err != nil {
		return HybridRepositoryStats{}, fmt.Errorf("failed to get memory stats: %w", err)
	}

	fileIds, err := r.fileRepo.List(ctx)
	if err != nil {
		return HybridRepositoryStats{}, fmt.Errorf("failed to get file stats: %w", err)
	}

	return HybridRepositoryStats{
		MemoryDAGCount: len(memoryIds),
		FileDAGCount:   len(fileIds),
		WriteThrough:   r.writeThrough,
	}, nil
}

// HybridRepositoryStats provides insights into repository state
type HybridRepositoryStats struct {
	MemoryDAGCount int  `json:"memory_dag_count"`
	FileDAGCount   int  `json:"file_dag_count"`
	WriteThrough   bool `json:"write_through"`
}

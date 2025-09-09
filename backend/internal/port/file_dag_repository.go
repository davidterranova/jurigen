package port

import (
	"context"
	"davidterranova/jurigen/backend/internal/dag"
	"davidterranova/jurigen/backend/internal/usecase"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const dagFileExtension = ".json"

type FileDAGRepository struct {
	filePath string
}

func NewFileDAGRepository(filePath string) *FileDAGRepository {
	return &FileDAGRepository{
		filePath: filePath,
	}
}

func (r *FileDAGRepository) Get(ctx context.Context, id uuid.UUID) (*dag.DAG, error) {
	dagFile := filepath.Join(r.filePath, id.String()+dagFileExtension)
	data, err := os.ReadFile(dagFile)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %s",
			usecase.ErrNotFound,
			fmt.Errorf("error reading file '%s': %w", dagFile, err),
		)
	}

	var dag = dag.NewDAG("Untitled DAG")
	err = dag.UnmarshalJSON(data)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %s",
			usecase.ErrInternal,
			fmt.Errorf("error unmarshalling file '%s': %w", dagFile, err),
		)
	}

	return dag, nil
}

// List returns all DAG IDs found in the file directory
func (r *FileDAGRepository) List(ctx context.Context) ([]uuid.UUID, error) {
	entries, err := os.ReadDir(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory '%s': %w", r.filePath, err)
	}

	//nolint:prealloc // This is a valid use of range
	var ids []uuid.UUID
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, dagFileExtension) {
			continue
		}

		// Extract UUID from filename
		idStr := strings.TrimSuffix(filename, dagFileExtension)
		id, err := uuid.Parse(idStr)
		if err != nil {
			// Skip invalid UUID filenames
			continue
		}

		ids = append(ids, id)
	}

	return ids, nil
}

// Create stores a new DAG to a file
func (r *FileDAGRepository) Create(ctx context.Context, dagObj *dag.DAG) error {
	if dagObj == nil {
		return fmt.Errorf("%w: DAG cannot be nil", usecase.ErrInvalidCommand)
	}

	dagFile := filepath.Join(r.filePath, dagObj.Id.String()+dagFileExtension)

	// Check if file already exists
	if _, err := os.Stat(dagFile); err == nil {
		return fmt.Errorf("%w: DAG with id %s already exists", usecase.ErrInvalidCommand, dagObj.Id.String())
	}

	// Marshal DAG to JSON
	data, err := dagObj.MarshalJSON()
	if err != nil {
		return fmt.Errorf("%w: error marshalling DAG: %w", usecase.ErrInternal, err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(r.filePath, 0755); err != nil {
		return fmt.Errorf("%w: error creating directory '%s': %w", usecase.ErrInternal, r.filePath, err)
	}

	// Write to file
	err = os.WriteFile(dagFile, data, 0644)
	if err != nil {
		return fmt.Errorf("%w: error writing file '%s': %w", usecase.ErrInternal, dagFile, err)
	}

	return nil
}

// Update modifies an existing DAG file using the provided function
func (r *FileDAGRepository) Update(ctx context.Context, id uuid.UUID, fnUpdate func(dag dag.DAG) (dag.DAG, error)) error {
	// First, get the existing DAG
	existingDAG, err := r.Get(ctx, id)
	if err != nil {
		return err // Error already wrapped by Get method
	}

	// Apply the update function
	updatedDAG, err := fnUpdate(*existingDAG)
	if err != nil {
		return fmt.Errorf("update function failed: %w", err)
	}

	// Validate that the ID hasn't changed
	if updatedDAG.Id != existingDAG.Id {
		return fmt.Errorf(
			"%w: update function cannot change DAG ID from %s to %s",
			usecase.ErrInvalidCommand,
			existingDAG.Id,
			updatedDAG.Id,
		)
	}

	// Marshal updated DAG to JSON
	data, err := updatedDAG.MarshalJSON()
	if err != nil {
		return fmt.Errorf("%w: error marshalling updated DAG: %w", usecase.ErrInternal, err)
	}

	// Write back to file
	dagFile := filepath.Join(r.filePath, id.String()+dagFileExtension)
	err = os.WriteFile(dagFile, data, 0644)
	if err != nil {
		return fmt.Errorf("%w: error writing updated file '%s': %w", usecase.ErrInternal, dagFile, err)
	}

	return nil
}

// Delete removes a DAG file from the file system
func (r *FileDAGRepository) Delete(ctx context.Context, id uuid.UUID) error {
	dagFile := filepath.Join(r.filePath, id.String()+dagFileExtension)

	// Check if file exists
	if _, err := os.Stat(dagFile); os.IsNotExist(err) {
		return fmt.Errorf(
			"%w: DAG with id %s not found in file system",
			usecase.ErrNotFound,
			id.String(),
		)
	}

	// Remove the file
	err := os.Remove(dagFile)
	if err != nil {
		return fmt.Errorf("%w: error deleting file '%s': %w", usecase.ErrInternal, dagFile, err)
	}

	return nil
}

package port

import (
	"context"
	"davidterranova/jurigen/internal/dag"
	"davidterranova/jurigen/internal/usecase"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// InMemoryDAGRepository implements the DAGRepository interface using in-memory storage
type InMemoryDAGRepository struct {
	dags map[uuid.UUID]*dag.DAG
	mu   sync.RWMutex // Protects concurrent access to the dags map
}

// NewInMemoryDAGRepository creates a new instance of InMemoryDAGRepository
func NewInMemoryDAGRepository() *InMemoryDAGRepository {
	return &InMemoryDAGRepository{
		dags: make(map[uuid.UUID]*dag.DAG),
	}
}

// Get retrieves a DAG by its ID from memory
func (r *InMemoryDAGRepository) Get(ctx context.Context, id uuid.UUID) (*dag.DAG, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dagObj, exists := r.dags[id]
	if !exists {
		return nil, fmt.Errorf(
			"%w: DAG with id %s not found in memory",
			usecase.ErrNotFound,
			id.String(),
		)
	}

	return dagObj, nil
}

// Create stores a DAG in memory
func (r *InMemoryDAGRepository) Create(ctx context.Context, dagObj *dag.DAG) error {
	if dagObj == nil {
		return fmt.Errorf("%w: DAG cannot be nil", usecase.ErrInvalidCommand)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.dags[dagObj.Id] = dagObj
	return nil
}

// Delete removes a DAG from memory
func (r *InMemoryDAGRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.dags[id]; !exists {
		return fmt.Errorf(
			"%w: DAG with id %s not found in memory",
			usecase.ErrNotFound,
			id.String(),
		)
	}

	delete(r.dags, id)
	return nil
}

// List returns all DAG IDs stored in memory
func (r *InMemoryDAGRepository) List(ctx context.Context) ([]uuid.UUID, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]uuid.UUID, 0, len(r.dags))
	for id := range r.dags {
		ids = append(ids, id)
	}

	return ids, nil
}

// Update modifies an existing DAG in memory using the provided function
func (r *InMemoryDAGRepository) Update(ctx context.Context, id uuid.UUID, fnUpdate func(dag dag.DAG) (dag.DAG, error)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if DAG exists using direct map access (to avoid deadlock with Get method)
	existingDAG, exists := r.dags[id]
	if !exists {
		return fmt.Errorf(
			"%w: DAG with id %s not found in memory",
			usecase.ErrNotFound,
			id.String(),
		)
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

	// Store the updated DAG
	r.dags[id] = &updatedDAG
	return nil
}

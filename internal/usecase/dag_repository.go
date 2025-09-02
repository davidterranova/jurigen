package usecase

import (
	"context"
	"davidterranova/jurigen/internal/dag"

	"github.com/google/uuid"
)

type DAGRepository interface {
	List(ctx context.Context) ([]uuid.UUID, error)
	Get(ctx context.Context, id uuid.UUID) (*dag.DAG, error)
	Create(ctx context.Context, dag *dag.DAG) error
	Update(ctx context.Context, id uuid.UUID, fnUpdate func(dag dag.DAG) (dag.DAG, error)) error
	Delete(ctx context.Context, id uuid.UUID) error
}

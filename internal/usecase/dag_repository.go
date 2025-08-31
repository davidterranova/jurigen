package usecase

import (
	"context"
	"davidterranova/jurigen/internal/dag"

	"github.com/google/uuid"
)

type DAGRepository interface {
	GetDAG(ctx context.Context, id uuid.UUID) (*dag.DAG, error)
}

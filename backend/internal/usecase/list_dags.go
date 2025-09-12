package usecase

import (
	"context"
	"davidterranova/jurigen/backend/internal/model"

	"github.com/google/uuid"
)

type CmdListDAGs struct {
	// No parameters needed for listing all DAGs
}

type ListDAGsUseCase struct {
	dagRepository DAGRepository
}

func NewListDAGsUseCase(dagRepository DAGRepository) *ListDAGsUseCase {
	return &ListDAGsUseCase{
		dagRepository: dagRepository,
	}
}

func (u *ListDAGsUseCase) List(ctx context.Context, cmd CmdListDAGs) ([]uuid.UUID, error) {
	return u.dagRepository.List(ctx)
}

// ListDAGs returns full DAG objects instead of just IDs
func (u *ListDAGsUseCase) ListDAGs(ctx context.Context, cmd CmdListDAGs) ([]*model.DAG, error) {
	dagIds, err := u.dagRepository.List(ctx)
	if err != nil {
		return nil, err
	}

	dags := make([]*model.DAG, 0, len(dagIds))
	for _, id := range dagIds {
		dag, err := u.dagRepository.Get(ctx, id)
		if err != nil {
			// Skip DAGs that can't be loaded, but log the error
			continue
		}
		dags = append(dags, dag)
	}

	return dags, nil
}

package usecase

import (
	"context"

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

func (u *ListDAGsUseCase) Execute(ctx context.Context, cmd CmdListDAGs) ([]uuid.UUID, error) {
	return u.dagRepository.List(ctx)
}

package usecase

import (
	"context"
	"davidterranova/jurigen/backend/internal/dag"
	"fmt"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
)

type CmdGetDAG struct {
	DAGId string `validate:"required,uuid"`
}

type GetDAGUseCase struct {
	dagRepository DAGRepository
	validator     *validator.Validate
}

func NewGetDAGUseCase(dagRepository DAGRepository) *GetDAGUseCase {
	return &GetDAGUseCase{
		dagRepository: dagRepository,
		validator:     validator.New(),
	}
}

func (u *GetDAGUseCase) Get(ctx context.Context, cmdGetDag CmdGetDAG) (*dag.DAG, error) {
	err := u.validator.Struct(cmdGetDag)
	if err != nil {
		return nil, fmt.Errorf("%w: %s (%v)", ErrInvalidCommand, err, cmdGetDag)
	}

	id, err := uuid.Parse(cmdGetDag.DAGId)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCommand, err)
	}

	return u.dagRepository.Get(ctx, id)
}

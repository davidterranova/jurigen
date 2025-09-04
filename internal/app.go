package pkg

import (
	"context"
	"davidterranova/jurigen/internal/dag"
	"davidterranova/jurigen/internal/usecase"

	"github.com/google/uuid"
)

type App struct {
	dagUseCase *dagUseCase
}

type dagUseCase struct {
	GetDAGUseCase
	ListDAGsUseCase
	UpdateDAGUseCase
}

type GetDAGUseCase interface {
	Get(ctx context.Context, cmd usecase.CmdGetDAG) (*dag.DAG, error)
}

type ListDAGsUseCase interface {
	List(ctx context.Context, cmd usecase.CmdListDAGs) ([]uuid.UUID, error)
}

type UpdateDAGUseCase interface {
	Execute(ctx context.Context, cmd usecase.CmdUpdateDAG) (*dag.DAG, error)
}

func New(dagRepository usecase.DAGRepository) *App {
	return &App{
		dagUseCase: &dagUseCase{
			usecase.NewGetDAGUseCase(dagRepository),
			usecase.NewListDAGsUseCase(dagRepository),
			usecase.NewUpdateDAGUseCase(dagRepository),
		},
	}
}

func (a *App) Get(ctx context.Context, cmd usecase.CmdGetDAG) (*dag.DAG, error) {
	return a.dagUseCase.Get(ctx, cmd)
}

func (a *App) List(ctx context.Context, cmd usecase.CmdListDAGs) ([]uuid.UUID, error) {
	return a.dagUseCase.List(ctx, cmd)
}

func (a *App) Update(ctx context.Context, cmd usecase.CmdUpdateDAG) (*dag.DAG, error) {
	return a.dagUseCase.Execute(ctx, cmd)
}

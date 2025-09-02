package pkg

import (
	"context"
	"davidterranova/jurigen/internal/dag"
	"davidterranova/jurigen/internal/usecase"

	"github.com/google/uuid"
)

type App struct {
	getDAGUseCase   GetDAGUseCase
	listDAGsUseCase ListDAGsUseCase
}

type GetDAGUseCase interface {
	Execute(ctx context.Context, cmd usecase.CmdGetDAG) (*dag.DAG, error)
}

type ListDAGsUseCase interface {
	Execute(ctx context.Context, cmd usecase.CmdListDAGs) ([]uuid.UUID, error)
}

func New(dagRepository usecase.DAGRepository) *App {
	return &App{
		getDAGUseCase:   usecase.NewGetDAGUseCase(dagRepository),
		listDAGsUseCase: usecase.NewListDAGsUseCase(dagRepository),
	}
}

func (a *App) GetDAG(ctx context.Context, cmd usecase.CmdGetDAG) (*dag.DAG, error) {
	return a.getDAGUseCase.Execute(ctx, cmd)
}

func (a *App) ListDAGs(ctx context.Context, cmd usecase.CmdListDAGs) ([]uuid.UUID, error) {
	return a.listDAGsUseCase.Execute(ctx, cmd)
}

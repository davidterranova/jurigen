package pkg

import (
	"context"
	"davidterranova/jurigen/internal/dag"
	"davidterranova/jurigen/internal/usecase"
)

type App struct {
	getDAGUseCase GetDAGUseCase
}

type GetDAGUseCase interface {
	Execute(ctx context.Context, cmd usecase.CmdGetDAG) (*dag.DAG, error)
}

func New(dagRepository usecase.DAGRepository) *App {
	return &App{
		getDAGUseCase: usecase.NewGetDAGUseCase(dagRepository),
	}
}

func (a *App) GetDAG(ctx context.Context, cmd usecase.CmdGetDAG) (*dag.DAG, error) {
	return a.getDAGUseCase.Execute(ctx, cmd)
}

package pkg

import (
	"context"
	"davidterranova/jurigen/backend/internal/model"
	"davidterranova/jurigen/backend/internal/usecase"

	"github.com/google/uuid"
)

type App struct {
	dagUseCase *dagUseCase
}

type dagUseCase struct {
	GetDAGUseCase
	ListDAGsUseCase
	UpdateDAGUseCase
	ValidateStoredDAGUseCase
}

type GetDAGUseCase interface {
	Get(ctx context.Context, cmd usecase.CmdGetDAG) (*model.DAG, error)
}

type ListDAGsUseCase interface {
	List(ctx context.Context, cmd usecase.CmdListDAGs) ([]uuid.UUID, error)
	ListDAGs(ctx context.Context, cmd usecase.CmdListDAGs) ([]*model.DAG, error)
}

type UpdateDAGUseCase interface {
	Execute(ctx context.Context, cmd usecase.CmdUpdateDAG) (*model.DAG, error)
}

type ValidateStoredDAGUseCase interface {
	Execute(ctx context.Context, cmd usecase.CmdValidateStoredDAG) (*usecase.ValidationResult, error)
}

func New(dagRepository usecase.DAGRepository) *App {
	return &App{
		dagUseCase: &dagUseCase{
			usecase.NewGetDAGUseCase(dagRepository),
			usecase.NewListDAGsUseCase(dagRepository),
			usecase.NewUpdateDAGUseCase(dagRepository),
			usecase.NewValidateStoredDAGUseCase(dagRepository),
		},
	}
}

func (a *App) Get(ctx context.Context, cmd usecase.CmdGetDAG) (*model.DAG, error) {
	return a.dagUseCase.Get(ctx, cmd)
}

func (a *App) List(ctx context.Context, cmd usecase.CmdListDAGs) ([]uuid.UUID, error) {
	return a.dagUseCase.List(ctx, cmd)
}

func (a *App) Update(ctx context.Context, cmd usecase.CmdUpdateDAG) (*model.DAG, error) {
	return a.dagUseCase.UpdateDAGUseCase.Execute(ctx, cmd)
}

func (a *App) ListDAGs(ctx context.Context, cmd usecase.CmdListDAGs) ([]*model.DAG, error) {
	return a.dagUseCase.ListDAGs(ctx, cmd)
}

func (a *App) ValidateStoredDAG(ctx context.Context, cmd usecase.CmdValidateStoredDAG) (*usecase.ValidationResult, error) {
	return a.dagUseCase.ValidateStoredDAGUseCase.Execute(ctx, cmd)
}

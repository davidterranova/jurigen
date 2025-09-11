package usecase

import (
	"context"
	"davidterranova/jurigen/backend/internal/model"

	"github.com/google/uuid"
)

//go:generate go run github.com/golang/mock/mockgen -source=dag_repository.go -destination=testdata/mocks/dag_repository_mock.go -package=mocks

type DAGRepository interface {
	List(ctx context.Context) ([]uuid.UUID, error)
	Get(ctx context.Context, id uuid.UUID) (*model.DAG, error)
	Create(ctx context.Context, dag *model.DAG) error
	Update(ctx context.Context, id uuid.UUID, fnUpdate func(dag model.DAG) (model.DAG, error)) error
	Delete(ctx context.Context, id uuid.UUID) error
}

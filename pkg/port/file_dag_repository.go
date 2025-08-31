package port

import (
	"context"
	"davidterranova/jurigen/internal/dag"
	"davidterranova/jurigen/pkg/usecase"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const dagFileExtension = ".json"

type FileDAGRepository struct {
	filePath string
}

func NewFileDAGRepository(filePath string) *FileDAGRepository {
	return &FileDAGRepository{
		filePath: filePath,
	}
}

func (r *FileDAGRepository) GetDAG(ctx context.Context, id uuid.UUID) (*dag.DAG, error) {
	dagFile := filepath.Join(r.filePath, id.String()+dagFileExtension)
	data, err := os.ReadFile(dagFile)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %s",
			usecase.ErrNotFound,
			fmt.Errorf("error reading file '%s': %w", dagFile, err),
		)
	}

	var dag = dag.NewDAG()
	err = dag.UnmarshalJSON(data)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %s",
			usecase.ErrInternal,
			fmt.Errorf("error unmarshalling file '%s': %w", dagFile, err),
		)
	}

	return dag, nil
}

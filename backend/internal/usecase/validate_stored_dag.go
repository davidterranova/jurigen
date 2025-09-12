package usecase

import (
	"context"
	"davidterranova/jurigen/backend/internal/model"
	"fmt"
	"time"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
)

type CmdValidateStoredDAG struct {
	DAGId string `validate:"required,uuid"`
}

type ValidateStoredDAGUseCase struct {
	dagRepository DAGRepository
	validator     *validator.Validate
}

func NewValidateStoredDAGUseCase(dagRepository DAGRepository) *ValidateStoredDAGUseCase {
	return &ValidateStoredDAGUseCase{
		dagRepository: dagRepository,
		validator:     validator.New(),
	}
}

// Execute validates a stored DAG and persists the metadata
func (u *ValidateStoredDAGUseCase) Execute(ctx context.Context, cmd CmdValidateStoredDAG) (*ValidationResult, error) {
	// Validate the command
	err := u.validator.Struct(cmd)
	if err != nil {
		return nil, fmt.Errorf("%w: %s (%v)", ErrInvalidCommand, err, cmd)
	}

	// Parse and validate the UUID
	id, err := uuid.Parse(cmd.DAGId)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid UUID format: %s", ErrInvalidCommand, err)
	}

	// Retrieve the DAG
	dag, err := u.dagRepository.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve DAG for validation: %w", err)
	}

	// Validate the DAG
	dagValidator := NewDAGValidator()
	validationResult := dagValidator.ValidateDAG(dag)

	// Update DAG metadata with validation results and persist
	err = u.dagRepository.Update(ctx, id, func(existingDAG model.DAG) (model.DAG, error) {
		// Initialize metadata if it doesn't exist
		if existingDAG.Metadata == nil {
			existingDAG.Metadata = model.NewDAGMetadata()
		}

		// Update metadata with validation results
		existingDAG.Metadata.IsValid = validationResult.IsValid
		existingDAG.Metadata.Statistics = u.convertValidationStatsToModel(validationResult.Statistics)
		existingDAG.Metadata.LastValidatedAt = time.Now()

		return existingDAG, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to persist validation metadata: %w", err)
	}

	return &validationResult, nil
}

// convertValidationStatsToModel converts usecase ValidationStatistics to model ValidationStatistics
func (u *ValidateStoredDAGUseCase) convertValidationStatsToModel(stats ValidationStatistics) model.ValidationStatistics {
	return model.ValidationStatistics{
		TotalNodes:   stats.TotalNodes,
		RootNodes:    stats.RootNodes,
		LeafNodes:    stats.LeafNodes,
		TotalAnswers: stats.TotalAnswers,
		MaxDepth:     stats.MaxDepth,
		HasCycles:    stats.HasCycles,
		RootNodeIDs:  stats.RootNodeIDs,
		LeafNodeIDs:  stats.LeafNodeIDs,
		CyclePaths:   stats.CyclePaths,
	}
}

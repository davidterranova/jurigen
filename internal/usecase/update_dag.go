package usecase

import (
	"context"
	"davidterranova/jurigen/internal/dag"
	"fmt"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
)

type CmdUpdateDAG struct {
	DAGId string   `validate:"required,uuid"`
	DAG   *dag.DAG `validate:"required"`
}

type UpdateDAGUseCase struct {
	dagRepository DAGRepository
	validator     *validator.Validate
}

func NewUpdateDAGUseCase(dagRepository DAGRepository) *UpdateDAGUseCase {
	return &UpdateDAGUseCase{
		dagRepository: dagRepository,
		validator:     validator.New(),
	}
}

func (u *UpdateDAGUseCase) Execute(ctx context.Context, cmd CmdUpdateDAG) (*dag.DAG, error) {
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

	// Update the DAG using repository's Update method
	var updatedDAG *dag.DAG
	err = u.dagRepository.Update(ctx, id, func(existingDAG dag.DAG) (dag.DAG, error) {
		// Validate that the DAG ID in the command matches the DAG ID in the payload
		if cmd.DAG.Id != id {
			return existingDAG, fmt.Errorf("%w: DAG ID mismatch - URL ID: %s, payload ID: %s", ErrInvalidCommand, id, cmd.DAG.Id)
		}

		// Validate DAG structure
		if err := u.validateDAGStructure(cmd.DAG); err != nil {
			return existingDAG, fmt.Errorf("%w: %s", ErrInvalidCommand, err)
		}

		// Replace the entire DAG with the new one
		updatedDAG = cmd.DAG

		return *cmd.DAG, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update DAG: %w", err)
	}

	return updatedDAG, nil
}

// validateDAGStructure performs comprehensive structural validation on the DAG
func (u *UpdateDAGUseCase) validateDAGStructure(d *dag.DAG) error {
	validator := NewDAGValidator()
	result := validator.ValidateDAG(d)

	if !result.IsValid {
		// Combine all error messages into a single error
		var errorMessages []string
		for _, err := range result.Errors {
			errorMessages = append(errorMessages, err.Message)
		}
		return fmt.Errorf("DAG validation failed: %v", errorMessages)
	}

	return nil
}

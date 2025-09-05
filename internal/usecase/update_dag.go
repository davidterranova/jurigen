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

// validateDAGStructure performs structural validation on the DAG
func (u *UpdateDAGUseCase) validateDAGStructure(d *dag.DAG) error {
	if d == nil {
		return fmt.Errorf("DAG cannot be nil")
	}

	if d.Id == uuid.Nil {
		return fmt.Errorf("DAG ID cannot be empty")
	}

	if d.Title == "" {
		return fmt.Errorf("DAG title cannot be empty")
	}

	// Validate nodes
	if len(d.Nodes) == 0 {
		return fmt.Errorf("DAG must contain at least one node")
	}

	// Validate each node
	for nodeId, node := range d.Nodes {
		if nodeId != node.Id {
			return fmt.Errorf("node map key %s does not match node ID %s", nodeId, node.Id)
		}

		if node.Question == "" {
			return fmt.Errorf("node %s must have a non-empty question", node.Id)
		}

		// Validate answers
		for i, answer := range node.Answers {
			if answer.Id == uuid.Nil {
				return fmt.Errorf("answer %d in node %s must have a valid ID", i, node.Id)
			}

			if answer.Statement == "" {
				return fmt.Errorf("answer %s in node %s must have a non-empty statement", answer.Id, node.Id)
			}

			// If NextNode is specified, validate it exists in the DAG
			if answer.NextNode != nil {
				if _, exists := d.Nodes[*answer.NextNode]; !exists {
					return fmt.Errorf("answer %s references non-existent next node %s", answer.Id, *answer.NextNode)
				}
			}
		}
	}

	// Validate DAG has exactly one root node
	rootNode, err := d.GetRootNode()
	if err != nil {
		return fmt.Errorf("DAG structure validation failed: %s", err.Error())
	}

	// Ensure root node exists in the nodes map
	if _, exists := d.Nodes[rootNode.Id]; !exists {
		return fmt.Errorf("root node %s not found in nodes map", rootNode.Id)
	}

	return nil
}

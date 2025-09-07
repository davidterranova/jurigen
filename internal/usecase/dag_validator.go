package usecase

import (
	"davidterranova/jurigen/internal/dag"
	"fmt"

	"github.com/google/uuid"
)

// ValidationResult represents the result of DAG validation
type ValidationResult struct {
	IsValid    bool                 `json:"is_valid"`
	Errors     []ValidationError    `json:"errors,omitempty"`
	Warnings   []ValidationWarning  `json:"warnings,omitempty"`
	Statistics ValidationStatistics `json:"statistics"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	NodeID   string `json:"node_id,omitempty"`
	AnswerID string `json:"answer_id,omitempty"`
	Severity string `json:"severity"` // "error", "warning"
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	NodeID   string `json:"node_id,omitempty"`
	AnswerID string `json:"answer_id,omitempty"`
}

// ValidationStatistics provides DAG structure statistics
type ValidationStatistics struct {
	TotalNodes   int      `json:"total_nodes"`
	RootNodes    int      `json:"root_nodes"`
	LeafNodes    int      `json:"leaf_nodes"`
	TotalAnswers int      `json:"total_answers"`
	MaxDepth     int      `json:"max_depth"`
	HasCycles    bool     `json:"has_cycles"`
	RootNodeIDs  []string `json:"root_node_ids,omitempty"`
	LeafNodeIDs  []string `json:"leaf_node_ids,omitempty"`
	CyclePaths   []string `json:"cycle_paths,omitempty"`
}

// DAGValidator provides comprehensive DAG validation functionality
type DAGValidator struct{}

// NewDAGValidator creates a new DAG validator instance
func NewDAGValidator() *DAGValidator {
	return &DAGValidator{}
}

// ValidateDAG performs comprehensive validation of a DAG structure
func (v *DAGValidator) ValidateDAG(d *dag.DAG) ValidationResult {
	result := ValidationResult{
		IsValid:    true,
		Errors:     []ValidationError{},
		Warnings:   []ValidationWarning{},
		Statistics: ValidationStatistics{},
	}

	if d == nil {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:     "DAG_NULL",
			Message:  "DAG cannot be nil",
			Severity: "error",
		})
		return result
	}

	// Perform all validations
	v.validateBasicStructure(d, &result)
	v.validateNodes(d, &result)
	v.validateRootNode(d, &result)
	v.validateCycles(d, &result)
	v.calculateStatistics(d, &result)

	return result
}

// validateBasicStructure validates basic DAG properties
func (v *DAGValidator) validateBasicStructure(d *dag.DAG, result *ValidationResult) {
	if d.Id == uuid.Nil {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:     "DAG_INVALID_ID",
			Message:  "DAG ID cannot be empty",
			Severity: "error",
		})
	}

	if d.Title == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:     "DAG_EMPTY_TITLE",
			Message:  "DAG title cannot be empty",
			Severity: "error",
		})
	}

	if len(d.Nodes) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:     "DAG_NO_NODES",
			Message:  "DAG must contain at least one node",
			Severity: "error",
		})
	}
}

// validateNodes validates individual nodes and their answers
func (v *DAGValidator) validateNodes(d *dag.DAG, result *ValidationResult) {
	for nodeId, node := range d.Nodes {
		// Validate node ID consistency
		if nodeId != node.Id {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:     "NODE_ID_MISMATCH",
				Message:  fmt.Sprintf("node map key %s does not match node ID %s", nodeId, node.Id),
				NodeID:   node.Id.String(),
				Severity: "error",
			})
		}

		// Validate node question
		if node.Question == "" {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:     "NODE_EMPTY_QUESTION",
				Message:  fmt.Sprintf("node %s must have a non-empty question", node.Id),
				NodeID:   node.Id.String(),
				Severity: "error",
			})
		}

		// Validate answers
		v.validateAnswers(d, node, result)
	}
}

// validateAnswers validates answers for a specific node
func (v *DAGValidator) validateAnswers(d *dag.DAG, node dag.Node, result *ValidationResult) {
	for i, answer := range node.Answers {
		// Validate answer ID
		if answer.Id == uuid.Nil {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ANSWER_INVALID_ID",
				Message:  fmt.Sprintf("answer %d in node %s must have a valid ID", i, node.Id),
				NodeID:   node.Id.String(),
				Severity: "error",
			})
		}

		// Validate answer statement
		if answer.Statement == "" {
			result.IsValid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:     "ANSWER_EMPTY_STATEMENT",
				Message:  fmt.Sprintf("answer %s in node %s must have a non-empty statement", answer.Id, node.Id),
				NodeID:   node.Id.String(),
				AnswerID: answer.Id.String(),
				Severity: "error",
			})
		}

		// Validate next_node references
		if answer.NextNode != nil {
			if _, exists := d.Nodes[*answer.NextNode]; !exists {
				result.IsValid = false
				result.Errors = append(result.Errors, ValidationError{
					Code:     "ANSWER_INVALID_REFERENCE",
					Message:  fmt.Sprintf("answer %s references non-existent next node %s", answer.Id, *answer.NextNode),
					NodeID:   node.Id.String(),
					AnswerID: answer.Id.String(),
					Severity: "error",
				})
			}
		}
	}
}

// validateRootNode ensures the DAG has exactly one root node
func (v *DAGValidator) validateRootNode(d *dag.DAG, result *ValidationResult) {
	// Find all nodes that are not referenced as next_node
	referencedNodes := make(map[uuid.UUID]bool)
	allNodes := make(map[uuid.UUID]bool)

	for nodeId := range d.Nodes {
		allNodes[nodeId] = true
	}

	for _, node := range d.Nodes {
		for _, answer := range node.Answers {
			if answer.NextNode != nil {
				referencedNodes[*answer.NextNode] = true
			}
		}
	}

	// Find root nodes (not referenced by any answer)
	var rootNodes []uuid.UUID
	for nodeId := range allNodes {
		if !referencedNodes[nodeId] {
			rootNodes = append(rootNodes, nodeId)
		}
	}

	if len(rootNodes) == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:     "DAG_NO_ROOT",
			Message:  "DAG has no root node - this indicates a circular reference",
			Severity: "error",
		})
	} else if len(rootNodes) > 1 {
		result.IsValid = false
		rootNodeIDs := make([]string, len(rootNodes))
		for i, id := range rootNodes {
			rootNodeIDs[i] = id.String()
		}
		result.Errors = append(result.Errors, ValidationError{
			Code:     "DAG_MULTIPLE_ROOTS",
			Message:  fmt.Sprintf("DAG has %d root nodes, expected exactly 1. Root nodes: %v", len(rootNodes), rootNodeIDs),
			Severity: "error",
		})
	}

	// Store root node statistics
	result.Statistics.RootNodes = len(rootNodes)
	result.Statistics.RootNodeIDs = make([]string, len(rootNodes))
	for i, id := range rootNodes {
		result.Statistics.RootNodeIDs[i] = id.String()
	}
}

// validateCycles detects cycles in the DAG using DFS
func (v *DAGValidator) validateCycles(d *dag.DAG, result *ValidationResult) {
	visited := make(map[uuid.UUID]bool)
	inStack := make(map[uuid.UUID]bool)
	cycles := []string{}

	var dfs func(uuid.UUID, []uuid.UUID) bool
	dfs = func(nodeId uuid.UUID, path []uuid.UUID) bool {
		if inStack[nodeId] {
			// Found a cycle - construct cycle path
			cycleStart := -1
			for i, id := range path {
				if id == nodeId {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cyclePath := make([]string, len(path[cycleStart:])+1)
				for i, id := range path[cycleStart:] {
					cyclePath[i] = id.String()
				}
				cyclePath[len(cyclePath)-1] = nodeId.String() // Close the cycle
				cycles = append(cycles, fmt.Sprintf("%v", cyclePath))
			}
			return true
		}

		if visited[nodeId] {
			return false
		}

		visited[nodeId] = true
		inStack[nodeId] = true
		//nolint:gocritic // This is a valid use of append
		newPath := append(path, nodeId)

		node, exists := d.Nodes[nodeId]
		if exists {
			for _, answer := range node.Answers {
				if answer.NextNode != nil {
					if dfs(*answer.NextNode, newPath) {
						return true
					}
				}
			}
		}

		inStack[nodeId] = false
		return false
	}

	// Check for cycles from each unvisited node
	hasCycles := false
	for nodeId := range d.Nodes {
		if !visited[nodeId] {
			if dfs(nodeId, []uuid.UUID{}) {
				hasCycles = true
			}
		}
	}

	result.Statistics.HasCycles = hasCycles
	result.Statistics.CyclePaths = cycles

	if hasCycles {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:     "DAG_HAS_CYCLES",
			Message:  fmt.Sprintf("DAG contains %d cycle(s). A valid DAG must be acyclic", len(cycles)),
			Severity: "error",
		})
	}
}

// calculateStatistics computes various DAG statistics
func (v *DAGValidator) calculateStatistics(d *dag.DAG, result *ValidationResult) {
	result.Statistics.TotalNodes = len(d.Nodes)

	// Count leaf nodes and total answers
	leafNodes := []string{}
	totalAnswers := 0

	for nodeId, node := range d.Nodes {
		totalAnswers += len(node.Answers)

		// A leaf node is one where all answers have no next_node
		isLeaf := len(node.Answers) == 0 || func() bool {
			for _, answer := range node.Answers {
				if answer.NextNode != nil {
					return false
				}
			}
			return true
		}()

		if isLeaf {
			leafNodes = append(leafNodes, nodeId.String())
		}
	}

	result.Statistics.LeafNodes = len(leafNodes)
	result.Statistics.LeafNodeIDs = leafNodes
	result.Statistics.TotalAnswers = totalAnswers

	// Calculate maximum depth using BFS from root nodes
	if len(result.Statistics.RootNodeIDs) > 0 && !result.Statistics.HasCycles {
		maxDepth := v.calculateMaxDepth(d, result.Statistics.RootNodeIDs[0])
		result.Statistics.MaxDepth = maxDepth
	}
}

// calculateMaxDepth calculates the maximum depth of the DAG using BFS
func (v *DAGValidator) calculateMaxDepth(d *dag.DAG, rootNodeID string) int {
	if rootNodeID == "" {
		return 0
	}

	rootID, err := uuid.Parse(rootNodeID)
	if err != nil {
		return 0
	}

	visited := make(map[uuid.UUID]bool)
	queue := []struct {
		nodeId uuid.UUID
		depth  int
	}{{rootID, 0}}

	maxDepth := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.nodeId] {
			continue
		}

		visited[current.nodeId] = true
		if current.depth > maxDepth {
			maxDepth = current.depth
		}

		node, exists := d.Nodes[current.nodeId]
		if exists {
			for _, answer := range node.Answers {
				if answer.NextNode != nil && !visited[*answer.NextNode] {
					queue = append(queue, struct {
						nodeId uuid.UUID
						depth  int
					}{*answer.NextNode, current.depth + 1})
				}
			}
		}
	}

	return maxDepth
}

// IsValidDAG performs a quick validation check (returns boolean only)
func (v *DAGValidator) IsValidDAG(d *dag.DAG) bool {
	result := v.ValidateDAG(d)
	return result.IsValid
}

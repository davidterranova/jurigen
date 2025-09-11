package usecase

import (
	"davidterranova/jurigen/backend/internal/model"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewDAGValidator(t *testing.T) {
	t.Parallel()

	validator := NewDAGValidator()
	assert.NotNil(t, validator)
}

func TestDAGValidator_ValidateDAG(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		dag                *model.DAG
		expectValid        bool
		expectedErrorCodes []string
		expectedStats      ValidationStatistics
	}{
		{
			name:               "nil DAG",
			dag:                nil,
			expectValid:        false,
			expectedErrorCodes: []string{"DAG_NULL"},
		},
		{
			name:               "valid single root DAG",
			dag:                createValidSingleRootDAG(),
			expectValid:        true,
			expectedErrorCodes: []string{},
			expectedStats: ValidationStatistics{
				TotalNodes:   3,
				RootNodes:    1,
				LeafNodes:    1,
				TotalAnswers: 4,
				MaxDepth:     2,
				HasCycles:    false,
			},
		},
		{
			name: "DAG with empty ID",
			dag: &model.DAG{
				Id:    uuid.Nil,
				Title: "Test DAG",
				Nodes: map[uuid.UUID]model.Node{},
			},
			expectValid:        false,
			expectedErrorCodes: []string{"DAG_INVALID_ID", "DAG_NO_NODES"},
		},
		{
			name: "DAG with empty title",
			dag: &model.DAG{
				Id:    uuid.New(),
				Title: "",
				Nodes: map[uuid.UUID]model.Node{},
			},
			expectValid:        false,
			expectedErrorCodes: []string{"DAG_EMPTY_TITLE", "DAG_NO_NODES"},
		},
		{
			name: "DAG with no nodes",
			dag: &model.DAG{
				Id:    uuid.New(),
				Title: "Test DAG",
				Nodes: map[uuid.UUID]model.Node{},
			},
			expectValid:        false,
			expectedErrorCodes: []string{"DAG_NO_NODES"},
		},
		{
			name:               "DAG with multiple root nodes",
			dag:                createMultipleRootDAG(),
			expectValid:        false,
			expectedErrorCodes: []string{"DAG_MULTIPLE_ROOTS"},
		},
		{
			name:               "DAG with cycles",
			dag:                createCyclicDAG(),
			expectValid:        false,
			expectedErrorCodes: []string{"DAG_HAS_CYCLES"},
		},
		{
			name:               "DAG with no root (circular reference)",
			dag:                createNoRootDAG(),
			expectValid:        false,
			expectedErrorCodes: []string{"DAG_NO_ROOT"},
		},
		{
			name:               "DAG with invalid node structure",
			dag:                createInvalidNodeStructureDAG(),
			expectValid:        false,
			expectedErrorCodes: []string{"NODE_ID_MISMATCH", "NODE_EMPTY_QUESTION"},
		},
		{
			name:               "DAG with invalid answers",
			dag:                createInvalidAnswersDAG(),
			expectValid:        false,
			expectedErrorCodes: []string{"ANSWER_EMPTY_STATEMENT", "ANSWER_INVALID_REFERENCE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validator := NewDAGValidator()
			result := validator.ValidateDAG(tt.dag)

			assert.Equal(t, tt.expectValid, result.IsValid, "Expected validity mismatch")

			// Check for expected error codes
			actualErrorCodes := make([]string, len(result.Errors))
			for i, err := range result.Errors {
				actualErrorCodes[i] = err.Code
			}

			for _, expectedCode := range tt.expectedErrorCodes {
				assert.Contains(t, actualErrorCodes, expectedCode, "Expected error code %s not found", expectedCode)
			}

			// Validate statistics if provided
			if tt.expectValid && tt.expectedStats.TotalNodes > 0 {
				assert.Equal(t, tt.expectedStats.TotalNodes, result.Statistics.TotalNodes)
				assert.Equal(t, tt.expectedStats.RootNodes, result.Statistics.RootNodes)
				assert.Equal(t, tt.expectedStats.LeafNodes, result.Statistics.LeafNodes)
				assert.Equal(t, tt.expectedStats.TotalAnswers, result.Statistics.TotalAnswers)
				assert.Equal(t, tt.expectedStats.HasCycles, result.Statistics.HasCycles)
			}
		})
	}
}

func TestDAGValidator_IsValidDAG(t *testing.T) {
	t.Parallel()

	validator := NewDAGValidator()

	// Test valid DAG
	validDAG := createValidSingleRootDAG()
	assert.True(t, validator.IsValidDAG(validDAG))

	// Test invalid DAG
	invalidDAG := createCyclicDAG()
	assert.False(t, validator.IsValidDAG(invalidDAG))

	// Test nil DAG
	assert.False(t, validator.IsValidDAG(nil))
}

func TestDAGValidator_CycleDetection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		dag           *model.DAG
		expectCycles  bool
		expectedPaths int
	}{
		{
			name:          "no cycles",
			dag:           createValidSingleRootDAG(),
			expectCycles:  false,
			expectedPaths: 0,
		},
		{
			name:          "simple cycle",
			dag:           createSimpleCyclicDAG(),
			expectCycles:  true,
			expectedPaths: 1,
		},
		{
			name:          "multiple cycles",
			dag:           createMultipleCyclesDAG(),
			expectCycles:  true,
			expectedPaths: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validator := NewDAGValidator()
			result := validator.ValidateDAG(tt.dag)

			assert.Equal(t, tt.expectCycles, result.Statistics.HasCycles)
			assert.Equal(t, tt.expectedPaths, len(result.Statistics.CyclePaths))

			if tt.expectCycles {
				assert.False(t, result.IsValid)
				// Should have DAG_HAS_CYCLES error
				found := false
				for _, err := range result.Errors {
					if err.Code == "DAG_HAS_CYCLES" {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected DAG_HAS_CYCLES error")
			}
		})
	}
}

func TestDAGValidator_DepthCalculation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		dag           *model.DAG
		expectedDepth int
	}{
		{
			name:          "single node",
			dag:           createSingleNodeDAG(),
			expectedDepth: 0,
		},
		{
			name:          "linear chain",
			dag:           createLinearChainDAG(5),
			expectedDepth: 4,
		},
		{
			name:          "branched DAG",
			dag:           createValidSingleRootDAG(),
			expectedDepth: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validator := NewDAGValidator()
			result := validator.ValidateDAG(tt.dag)

			if result.IsValid {
				assert.Equal(t, tt.expectedDepth, result.Statistics.MaxDepth)
			}
		})
	}
}

// Helper functions to create test DAGs

func createValidSingleRootDAG() *model.DAG {
	rootID := uuid.New()
	middleID := uuid.New()
	leafID := uuid.New()

	answer1ID := uuid.New()
	answer2ID := uuid.New()
	answer3ID := uuid.New()
	answer4ID := uuid.New()

	return &model.DAG{
		Id:    uuid.New(),
		Title: "Valid Single Root DAG",
		Nodes: map[uuid.UUID]model.Node{
			rootID: {
				Id:       rootID,
				Question: "Root question?",
				Answers: []model.Answer{
					{
						Id:        answer1ID,
						Statement: "Go to middle",
						NextNode:  &middleID,
					},
					{
						Id:        answer2ID,
						Statement: "Go to leaf",
						NextNode:  &leafID,
					},
				},
			},
			middleID: {
				Id:       middleID,
				Question: "Middle question?",
				Answers: []model.Answer{
					{
						Id:        answer3ID,
						Statement: "Go to leaf",
						NextNode:  &leafID,
					},
					{
						Id:        answer4ID,
						Statement: "Stay here",
						NextNode:  nil, // Leaf answer
					},
				},
			},
			leafID: {
				Id:       leafID,
				Question: "Leaf question?",
				Answers:  []model.Answer{}, // Leaf node
			},
		},
	}
}

func createMultipleRootDAG() *model.DAG {
	root1ID := uuid.New()
	root2ID := uuid.New()
	leafID := uuid.New()

	return &model.DAG{
		Id:    uuid.New(),
		Title: "Multiple Root DAG",
		Nodes: map[uuid.UUID]model.Node{
			root1ID: {
				Id:       root1ID,
				Question: "Root 1 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go to leaf",
						NextNode:  &leafID,
					},
				},
			},
			root2ID: {
				Id:       root2ID,
				Question: "Root 2 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go to leaf",
						NextNode:  &leafID,
					},
				},
			},
			leafID: {
				Id:       leafID,
				Question: "Leaf question?",
				Answers:  []model.Answer{},
			},
		},
	}
}

func createCyclicDAG() *model.DAG {
	node1ID := uuid.New()
	node2ID := uuid.New()
	node3ID := uuid.New()

	return &model.DAG{
		Id:    uuid.New(),
		Title: "Cyclic DAG",
		Nodes: map[uuid.UUID]model.Node{
			node1ID: {
				Id:       node1ID,
				Question: "Node 1 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go to node 2",
						NextNode:  &node2ID,
					},
				},
			},
			node2ID: {
				Id:       node2ID,
				Question: "Node 2 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go to node 3",
						NextNode:  &node3ID,
					},
				},
			},
			node3ID: {
				Id:       node3ID,
				Question: "Node 3 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go back to node 1", // Creates cycle
						NextNode:  &node1ID,
					},
				},
			},
		},
	}
}

func createNoRootDAG() *model.DAG {
	node1ID := uuid.New()
	node2ID := uuid.New()

	return &model.DAG{
		Id:    uuid.New(),
		Title: "No Root DAG",
		Nodes: map[uuid.UUID]model.Node{
			node1ID: {
				Id:       node1ID,
				Question: "Node 1 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go to node 2",
						NextNode:  &node2ID,
					},
				},
			},
			node2ID: {
				Id:       node2ID,
				Question: "Node 2 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go back to node 1",
						NextNode:  &node1ID,
					},
				},
			},
		},
	}
}

func createInvalidNodeStructureDAG() *model.DAG {
	nodeID := uuid.New()
	wrongID := uuid.New()

	return &model.DAG{
		Id:    uuid.New(),
		Title: "Invalid Node Structure DAG",
		Nodes: map[uuid.UUID]model.Node{
			nodeID: {
				Id:       wrongID, // Mismatch between map key and node ID
				Question: "",      // Empty question
				Answers:  []model.Answer{},
			},
		},
	}
}

func createInvalidAnswersDAG() *model.DAG {
	nodeID := uuid.New()
	nonExistentID := uuid.New()

	return &model.DAG{
		Id:    uuid.New(),
		Title: "Invalid Answers DAG",
		Nodes: map[uuid.UUID]model.Node{
			nodeID: {
				Id:       nodeID,
				Question: "Test question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "", // Empty statement
						NextNode:  nil,
					},
					{
						Id:        uuid.New(),
						Statement: "Valid statement",
						NextNode:  &nonExistentID, // References non-existent node
					},
				},
			},
		},
	}
}

func createSimpleCyclicDAG() *model.DAG {
	node1ID := uuid.New()
	node2ID := uuid.New()

	return &model.DAG{
		Id:    uuid.New(),
		Title: "Simple Cyclic DAG",
		Nodes: map[uuid.UUID]model.Node{
			node1ID: {
				Id:       node1ID,
				Question: "Node 1 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go to node 2",
						NextNode:  &node2ID,
					},
				},
			},
			node2ID: {
				Id:       node2ID,
				Question: "Node 2 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go back to node 1",
						NextNode:  &node1ID,
					},
				},
			},
		},
	}
}

func createMultipleCyclesDAG() *model.DAG {
	node1ID := uuid.New()
	node2ID := uuid.New()
	node3ID := uuid.New()
	node4ID := uuid.New()

	return &model.DAG{
		Id:    uuid.New(),
		Title: "Multiple Cycles DAG",
		Nodes: map[uuid.UUID]model.Node{
			node1ID: {
				Id:       node1ID,
				Question: "Node 1 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go to node 2",
						NextNode:  &node2ID,
					},
					{
						Id:        uuid.New(),
						Statement: "Go to node 3",
						NextNode:  &node3ID,
					},
				},
			},
			node2ID: {
				Id:       node2ID,
				Question: "Node 2 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go back to node 1", // Cycle 1
						NextNode:  &node1ID,
					},
				},
			},
			node3ID: {
				Id:       node3ID,
				Question: "Node 3 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go to node 4",
						NextNode:  &node4ID,
					},
				},
			},
			node4ID: {
				Id:       node4ID,
				Question: "Node 4 question?",
				Answers: []model.Answer{
					{
						Id:        uuid.New(),
						Statement: "Go back to node 3", // Cycle 2
						NextNode:  &node3ID,
					},
				},
			},
		},
	}
}

func createSingleNodeDAG() *model.DAG {
	nodeID := uuid.New()

	return &model.DAG{
		Id:    uuid.New(),
		Title: "Single Node DAG",
		Nodes: map[uuid.UUID]model.Node{
			nodeID: {
				Id:       nodeID,
				Question: "Only question?",
				Answers:  []model.Answer{},
			},
		},
	}
}

func createLinearChainDAG(length int) *model.DAG {
	if length <= 0 {
		return createSingleNodeDAG()
	}

	nodes := make(map[uuid.UUID]model.Node)
	nodeIDs := make([]uuid.UUID, length)

	// Create node IDs
	for i := 0; i < length; i++ {
		nodeIDs[i] = uuid.New()
	}

	// Create nodes with linear chain structure
	for i, nodeID := range nodeIDs {
		answers := []model.Answer{}
		if i < length-1 {
			// Not the last node, create answer pointing to next node
			answers = append(answers, model.Answer{
				Id:        uuid.New(),
				Statement: fmt.Sprintf("Go to node %d", i+2),
				NextNode:  &nodeIDs[i+1],
			})
		}

		nodes[nodeID] = model.Node{
			Id:       nodeID,
			Question: fmt.Sprintf("Question %d?", i+1),
			Answers:  answers,
		}
	}

	return &model.DAG{
		Id:    uuid.New(),
		Title: fmt.Sprintf("Linear Chain DAG (length %d)", length),
		Nodes: nodes,
	}
}

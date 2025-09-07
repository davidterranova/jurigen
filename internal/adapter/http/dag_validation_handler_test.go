package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDAGHandler_ValidateDAG(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		requestBody        interface{}
		expectedStatus     int
		expectedValid      bool
		expectedErrorCodes []string
	}{
		{
			name:           "valid DAG structure",
			requestBody:    createValidDAGRequest(),
			expectedStatus: http.StatusOK,
			expectedValid:  true,
		},
		{
			name:               "DAG with multiple root nodes",
			requestBody:        createMultipleRootDAGRequest(),
			expectedStatus:     http.StatusOK,
			expectedValid:      false,
			expectedErrorCodes: []string{"DAG_MULTIPLE_ROOTS"},
		},
		{
			name:               "DAG with cycles",
			requestBody:        createCyclicDAGRequest(),
			expectedStatus:     http.StatusOK,
			expectedValid:      false,
			expectedErrorCodes: []string{"DAG_HAS_CYCLES"},
		},
		{
			name:               "DAG with empty title",
			requestBody:        createEmptyTitleDAGRequest(),
			expectedStatus:     http.StatusOK,
			expectedValid:      false,
			expectedErrorCodes: []string{"DAG_EMPTY_TITLE"},
		},
		{
			name:           "invalid JSON request",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty request body",
			requestBody: ValidateRequest{
				DAG: DAGPresenter{},
			},
			expectedStatus:     http.StatusOK,
			expectedValid:      false,
			expectedErrorCodes: []string{"DAG_INVALID_ID", "DAG_EMPTY_TITLE", "DAG_NO_NODES"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create request body
			var requestBody []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			// Create HTTP request
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/dags/validate", bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create handler
			handler := NewDAGHandler(nil) // No app needed for validation

			// Execute request
			handler.ValidateDAG(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				// Parse response
				var response ValidationResultPresenter
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				// Check validation result
				assert.Equal(t, tt.expectedValid, response.IsValid)

				// Check for expected error codes
				if !tt.expectedValid && len(tt.expectedErrorCodes) > 0 {
					actualErrorCodes := make([]string, len(response.Errors))
					for i, err := range response.Errors {
						actualErrorCodes[i] = err.Code
					}

					for _, expectedCode := range tt.expectedErrorCodes {
						assert.Contains(t, actualErrorCodes, expectedCode, "Expected error code %s not found", expectedCode)
					}
				}

				// Validate response structure
				assert.GreaterOrEqual(t, response.Statistics.TotalNodes, 0)
				assert.GreaterOrEqual(t, response.Statistics.RootNodes, 0)
				assert.GreaterOrEqual(t, response.Statistics.LeafNodes, 0)
				assert.GreaterOrEqual(t, response.Statistics.TotalAnswers, 0)
			}
		})
	}
}

func TestDAGHandler_ValidateDAG_Integration(t *testing.T) {
	t.Parallel()

	// Test with a complex valid DAG
	request := createComplexValidDAGRequest()
	requestBody, err := json.Marshal(request)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/dags/validate", bytes.NewBuffer(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := NewDAGHandler(nil)

	handler.ValidateDAG(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response ValidationResultPresenter
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.IsValid)
	assert.Equal(t, 1, response.Statistics.RootNodes)
	assert.False(t, response.Statistics.HasCycles)
	assert.GreaterOrEqual(t, response.Statistics.MaxDepth, 0)
}

// Helper functions for creating test DAGs

func createValidDAGRequest() ValidateRequest {
	rootID := uuid.New()
	leafID := uuid.New()

	return ValidateRequest{
		DAG: DAGPresenter{
			Id:    uuid.New(),
			Title: "Valid Test DAG",
			Nodes: []NodePresenter{
				{
					Id:       rootID,
					Question: "Root question?",
					Answers: []AnswerPresenter{
						{
							Id:        uuid.New(),
							Statement: "Go to leaf",
							NextNode:  &leafID,
						},
						{
							Id:        uuid.New(),
							Statement: "End here",
							NextNode:  nil,
						},
					},
				},
				{
					Id:       leafID,
					Question: "Leaf question?",
					Answers:  []AnswerPresenter{},
				},
			},
		},
	}
}

func createMultipleRootDAGRequest() ValidateRequest {
	root1ID := uuid.New()
	root2ID := uuid.New()
	leafID := uuid.New()

	return ValidateRequest{
		DAG: DAGPresenter{
			Id:    uuid.New(),
			Title: "Multiple Root DAG",
			Nodes: []NodePresenter{
				{
					Id:       root1ID,
					Question: "Root 1 question?",
					Answers: []AnswerPresenter{
						{
							Id:        uuid.New(),
							Statement: "Go to leaf",
							NextNode:  &leafID,
						},
					},
				},
				{
					Id:       root2ID,
					Question: "Root 2 question?",
					Answers: []AnswerPresenter{
						{
							Id:        uuid.New(),
							Statement: "Go to leaf",
							NextNode:  &leafID,
						},
					},
				},
				{
					Id:       leafID,
					Question: "Leaf question?",
					Answers:  []AnswerPresenter{},
				},
			},
		},
	}
}

func createCyclicDAGRequest() ValidateRequest {
	node1ID := uuid.New()
	node2ID := uuid.New()

	return ValidateRequest{
		DAG: DAGPresenter{
			Id:    uuid.New(),
			Title: "Cyclic DAG",
			Nodes: []NodePresenter{
				{
					Id:       node1ID,
					Question: "Node 1 question?",
					Answers: []AnswerPresenter{
						{
							Id:        uuid.New(),
							Statement: "Go to node 2",
							NextNode:  &node2ID,
						},
					},
				},
				{
					Id:       node2ID,
					Question: "Node 2 question?",
					Answers: []AnswerPresenter{
						{
							Id:        uuid.New(),
							Statement: "Go back to node 1",
							NextNode:  &node1ID,
						},
					},
				},
			},
		},
	}
}

func createEmptyTitleDAGRequest() ValidateRequest {
	nodeID := uuid.New()

	return ValidateRequest{
		DAG: DAGPresenter{
			Id:    uuid.New(),
			Title: "", // Empty title
			Nodes: []NodePresenter{
				{
					Id:       nodeID,
					Question: "Test question?",
					Answers:  []AnswerPresenter{},
				},
			},
		},
	}
}

func createComplexValidDAGRequest() ValidateRequest {
	rootID := uuid.New()
	middle1ID := uuid.New()
	middle2ID := uuid.New()
	leaf1ID := uuid.New()
	leaf2ID := uuid.New()

	return ValidateRequest{
		DAG: DAGPresenter{
			Id:    uuid.New(),
			Title: "Complex Valid DAG",
			Nodes: []NodePresenter{
				{
					Id:       rootID,
					Question: "What type of case is this?",
					Answers: []AnswerPresenter{
						{
							Id:        uuid.New(),
							Statement: "Employment Issue",
							NextNode:  &middle1ID,
						},
						{
							Id:        uuid.New(),
							Statement: "Contract Dispute",
							NextNode:  &middle2ID,
						},
					},
				},
				{
					Id:       middle1ID,
					Question: "Employment details?",
					Answers: []AnswerPresenter{
						{
							Id:        uuid.New(),
							Statement: "Discrimination",
							NextNode:  &leaf1ID,
						},
						{
							Id:        uuid.New(),
							Statement: "Wrongful Termination",
							NextNode:  &leaf2ID,
						},
					},
				},
				{
					Id:       middle2ID,
					Question: "Contract type?",
					Answers: []AnswerPresenter{
						{
							Id:        uuid.New(),
							Statement: "Employment Contract",
							NextNode:  &leaf1ID,
						},
						{
							Id:        uuid.New(),
							Statement: "Service Agreement",
							NextNode:  &leaf2ID,
						},
					},
				},
				{
					Id:       leaf1ID,
					Question: "Gather evidence for discrimination case",
					Answers:  []AnswerPresenter{},
				},
				{
					Id:       leaf2ID,
					Question: "Review contract terms",
					Answers:  []AnswerPresenter{},
				},
			},
		},
	}
}

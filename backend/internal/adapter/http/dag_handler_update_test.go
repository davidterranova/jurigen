package http

import (
	"bytes"
	"context"
	"davidterranova/jurigen/backend/internal/adapter/http/testdata/mocks"
	"davidterranova/jurigen/backend/internal/dag"
	"davidterranova/jurigen/backend/internal/usecase"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDAGHandler_Update(t *testing.T) {
	testDAG := createTestDAG()

	tests := []struct {
		name           string
		dagId          string
		requestBody    interface{}
		setupMock      func(*mocks.MockApp)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successfully updates DAG",
			dagId:       testDAG.Id.String(),
			requestBody: NewDAGPresenter(testDAG),
			setupMock: func(mockApp *mocks.MockApp) {
				mockApp.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, cmd usecase.CmdUpdateDAG) (*dag.DAG, error) {
						// Verify the command has the correct DAG ID
						assert.Equal(t, testDAG.Id.String(), cmd.DAGId)
						assert.NotNil(t, cmd.DAG)
						return testDAG, nil
					},
				)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response DAGPresenter
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, testDAG.Id, response.Id)
				assert.Len(t, response.Nodes, len(testDAG.Nodes))
			},
		},
		{
			name:        "returns 400 for invalid JSON",
			dagId:       testDAG.Id.String(),
			requestBody: "invalid json",
			setupMock: func(mockApp *mocks.MockApp) {
				// No app call expected due to JSON parsing failure
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "invalid request body")
			},
		},
		{
			name:        "returns 400 for invalid DAG data",
			dagId:       testDAG.Id.String(),
			requestBody: NewDAGPresenter(testDAG),
			setupMock: func(mockApp *mocks.MockApp) {
				mockApp.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, usecase.ErrInvalidCommand)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "invalid DAG data")
			},
		},
		{
			name:        "returns 404 when DAG not found",
			dagId:       testDAG.Id.String(),
			requestBody: NewDAGPresenter(testDAG),
			setupMock: func(mockApp *mocks.MockApp) {
				mockApp.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, usecase.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "DAG not found")
			},
		},
		{
			name:        "returns 500 for internal server error",
			dagId:       testDAG.Id.String(),
			requestBody: NewDAGPresenter(testDAG),
			setupMock: func(mockApp *mocks.MockApp) {
				mockApp.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, usecase.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "failed to update DAG")
			},
		},
		{
			name:        "returns 400 for malformed UUID",
			dagId:       "invalid-uuid",
			requestBody: NewDAGPresenter(testDAG),
			setupMock: func(mockApp *mocks.MockApp) {
				mockApp.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, usecase.ErrInvalidCommand)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "invalid DAG data")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockApp := mocks.NewMockApp(ctrl)
			tt.setupMock(mockApp)

			handler := NewDAGHandler(mockApp)

			// Create request body
			var requestBody []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				requestBody, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req, err := http.NewRequest("PUT", "/v1/dags/"+tt.dagId, bytes.NewBuffer(requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Set up mux vars
			req = mux.SetURLVars(req, map[string]string{"dagId": tt.dagId})

			rr := httptest.NewRecorder()

			handler.Update(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestDAGHandler_PresenterToDAG(t *testing.T) {
	testDAG := createTestDAG()
	presenter := NewDAGPresenter(testDAG)

	handler := NewDAGHandler(nil) // App not needed for this test
	convertedDAG := handler.presenterToDAG(presenter)

	// Verify the conversion
	assert.Equal(t, testDAG.Id, convertedDAG.Id)
	assert.Equal(t, len(testDAG.Nodes), len(convertedDAG.Nodes))

	// Verify each node
	for nodeId, originalNode := range testDAG.Nodes {
		convertedNode, exists := convertedDAG.Nodes[nodeId]
		assert.True(t, exists)
		assert.Equal(t, originalNode.Id, convertedNode.Id)
		assert.Equal(t, originalNode.Question, convertedNode.Question)
		assert.Equal(t, len(originalNode.Answers), len(convertedNode.Answers))

		// Verify answers
		for i, originalAnswer := range originalNode.Answers {
			convertedAnswer := convertedNode.Answers[i]
			assert.Equal(t, originalAnswer.Id, convertedAnswer.Id)
			assert.Equal(t, originalAnswer.Statement, convertedAnswer.Statement)
			assert.Equal(t, originalAnswer.NextNode, convertedAnswer.NextNode)
			assert.Equal(t, originalAnswer.UserContext, convertedAnswer.UserContext)
			assert.Equal(t, originalAnswer.Metadata, convertedAnswer.Metadata)

			// Verify parent pointer is set correctly
			assert.NotNil(t, convertedAnswer.ParentNode)
			assert.Equal(t, convertedNode.Id, convertedAnswer.ParentNode.Id)
		}
	}
}

func TestDAGHandler_Update_ComplexDAG(t *testing.T) {
	// Create a more complex DAG with multiple nodes and metadata
	complexDAG := createComplexTestDAG()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockApp(ctrl)
	mockApp.EXPECT().Update(gomock.Any(), gomock.Any()).Return(complexDAG, nil)

	handler := NewDAGHandler(mockApp)
	presenter := NewDAGPresenter(complexDAG)

	requestBody, err := json.Marshal(presenter)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", "/v1/dags/"+complexDAG.Id.String(), bytes.NewBuffer(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"dagId": complexDAG.Id.String()})

	rr := httptest.NewRecorder()
	handler.Update(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response DAGPresenter
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, complexDAG.Id, response.Id)

	// Verify metadata is preserved
	hasMetadata := false
	for _, node := range response.Nodes {
		for _, answer := range node.Answers {
			if len(answer.Metadata) > 0 {
				hasMetadata = true
				break
			}
		}
	}
	assert.True(t, hasMetadata, "Metadata should be preserved in complex DAG")
}

// Helper function to create a test DAG
func createTestDAG() *dag.DAG {
	dagId := uuid.New()
	nodeId := uuid.New()
	answerId := uuid.New()

	return &dag.DAG{
		Id: dagId,
		Nodes: map[uuid.UUID]dag.Node{
			nodeId: {
				Id:       nodeId,
				Question: "Is this a test question?",
				Answers: []dag.Answer{
					{
						Id:        answerId,
						Statement: "Yes, this is a test",
						NextNode:  nil, // Leaf node
					},
				},
			},
		},
	}
}

// Helper function to create a complex test DAG with metadata
func createComplexTestDAG() *dag.DAG {
	dagId := uuid.New()
	rootNodeId := uuid.New()
	leafNodeId := uuid.New()
	answerId1 := uuid.New()
	answerId2 := uuid.New()
	leafAnswerId := uuid.New()

	return &dag.DAG{
		Id: dagId,
		Nodes: map[uuid.UUID]dag.Node{
			rootNodeId: {
				Id:       rootNodeId,
				Question: "Did you experience workplace discrimination?",
				Answers: []dag.Answer{
					{
						Id:          answerId1,
						Statement:   "Yes, I experienced discrimination",
						NextNode:    &leafNodeId,
						UserContext: "It happened during my performance review",
						Metadata: map[string]interface{}{
							"confidence": 0.8,
							"tags":       []string{"discrimination", "workplace"},
							"severity":   "high",
						},
					},
					{
						Id:        answerId2,
						Statement: "No discrimination occurred",
						NextNode:  nil, // Leaf node
					},
				},
			},
			leafNodeId: {
				Id:       leafNodeId,
				Question: "What type of discrimination did you experience?",
				Answers: []dag.Answer{
					{
						Id:          leafAnswerId,
						Statement:   "Age-based discrimination",
						NextNode:    nil, // Leaf node
						UserContext: "Comments were made about my age affecting performance",
						Metadata: map[string]interface{}{
							"legal_strength": 0.7,
							"evidence_type":  "verbal",
							"witnesses":      2,
						},
					},
				},
			},
		},
	}
}

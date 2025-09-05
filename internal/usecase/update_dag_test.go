package usecase

import (
	"context"
	"davidterranova/jurigen/internal/dag"
	"davidterranova/jurigen/internal/usecase/testdata/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpdateDAGUseCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockDAGRepository(ctrl)
	useCase := NewUpdateDAGUseCase(mockRepo)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.dagRepository)
	assert.NotNil(t, useCase.validator)
}

func TestUpdateDAGUseCase_Execute(t *testing.T) {
	testDAG := createValidTestDAG()
	unknownUUID := uuid.New()

	tests := []struct {
		name        string
		cmd         CmdUpdateDAG
		setupMock   func(*mocks.MockDAGRepository)
		expectedDAG *dag.DAG
		expectError bool
		errorType   error
	}{
		{
			name: "successfully updates existing DAG",
			cmd: CmdUpdateDAG{
				DAGId: testDAG.Id.String(),
				DAG:   testDAG,
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				mockRepo.EXPECT().Update(gomock.Any(), testDAG.Id, gomock.Any()).DoAndReturn(
					func(ctx context.Context, id uuid.UUID, fnUpdate func(dag.DAG) (dag.DAG, error)) error {
						_, err := fnUpdate(*testDAG)
						return err
					},
				)
			},
			expectedDAG: testDAG,
			expectError: false,
		},
		{
			name: "returns validation error for empty DAG ID",
			cmd: CmdUpdateDAG{
				DAGId: "",
				DAG:   testDAG,
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				// No repository call expected due to validation failure
			},
			expectError: true,
			errorType:   ErrInvalidCommand,
		},
		{
			name: "returns validation error for invalid UUID format",
			cmd: CmdUpdateDAG{
				DAGId: "invalid-uuid-format",
				DAG:   testDAG,
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				// No repository call expected due to validation failure
			},
			expectError: true,
			errorType:   ErrInvalidCommand,
		},
		{
			name: "returns validation error for nil DAG",
			cmd: CmdUpdateDAG{
				DAGId: testDAG.Id.String(),
				DAG:   nil,
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				// No repository call expected due to validation failure
			},
			expectError: true,
			errorType:   ErrInvalidCommand,
		},
		{
			name: "returns validation error for DAG ID mismatch",
			cmd: CmdUpdateDAG{
				DAGId: unknownUUID.String(), // Different ID
				DAG:   testDAG,
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				mockRepo.EXPECT().Update(gomock.Any(), unknownUUID, gomock.Any()).Return(ErrInvalidCommand)
			},
			expectError: true,
			errorType:   ErrInvalidCommand,
		},
		{
			name: "returns not found error when DAG doesn't exist",
			cmd: CmdUpdateDAG{
				DAGId: testDAG.Id.String(),
				DAG:   testDAG,
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				mockRepo.EXPECT().Update(gomock.Any(), testDAG.Id, gomock.Any()).Return(ErrNotFound)
			},
			expectError: true,
			errorType:   ErrNotFound,
		},
		{
			name: "returns internal error when repository update fails",
			cmd: CmdUpdateDAG{
				DAGId: testDAG.Id.String(),
				DAG:   testDAG,
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				mockRepo.EXPECT().Update(gomock.Any(), testDAG.Id, gomock.Any()).Return(ErrInternal)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockDAGRepository(ctrl)
			tt.setupMock(mockRepo)

			useCase := NewUpdateDAGUseCase(mockRepo)
			ctx := context.Background()

			result, err := useCase.Execute(ctx, tt.cmd)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)

				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.expectedDAG != nil {
				assert.Equal(t, tt.expectedDAG.Id, result.Id)
				assert.Equal(t, len(tt.expectedDAG.Nodes), len(result.Nodes))
			}
		})
	}
}

func TestUpdateDAGUseCase_ValidateDAGStructure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockDAGRepository(ctrl)
	useCase := NewUpdateDAGUseCase(mockRepo)

	tests := []struct {
		name      string
		dag       *dag.DAG
		wantError bool
		errorMsg  string
	}{
		{
			name:      "validates nil DAG",
			dag:       nil,
			wantError: true,
			errorMsg:  "DAG cannot be nil",
		},
		{
			name: "validates empty DAG ID",
			dag: &dag.DAG{
				Id:    uuid.Nil,
				Title: "Valid Title",
				Nodes: map[uuid.UUID]dag.Node{},
			},
			wantError: true,
			errorMsg:  "DAG ID cannot be empty",
		},
		{
			name: "validates empty title",
			dag: &dag.DAG{
				Id:    uuid.New(),
				Title: "",
				Nodes: map[uuid.UUID]dag.Node{},
			},
			wantError: true,
			errorMsg:  "DAG title cannot be empty",
		},
		{
			name: "validates empty nodes",
			dag: &dag.DAG{
				Id:    uuid.New(),
				Title: "Valid Title",
				Nodes: map[uuid.UUID]dag.Node{},
			},
			wantError: true,
			errorMsg:  "DAG must contain at least one node",
		},
		{
			name: "validates node ID mismatch",
			dag: func() *dag.DAG {
				nodeId := uuid.New()
				return &dag.DAG{
					Id:    uuid.New(),
					Title: "Valid Title",
					Nodes: map[uuid.UUID]dag.Node{
						nodeId: {
							Id:       uuid.New(), // Different ID
							Question: "Test question?",
							Answers:  []dag.Answer{},
						},
					},
				}
			}(),
			wantError: true,
			errorMsg:  "node map key",
		},
		{
			name: "validates empty question",
			dag: func() *dag.DAG {
				nodeId := uuid.New()
				return &dag.DAG{
					Id:    uuid.New(),
					Title: "Valid Title",
					Nodes: map[uuid.UUID]dag.Node{
						nodeId: {
							Id:       nodeId,
							Question: "", // Empty question
							Answers:  []dag.Answer{},
						},
					},
				}
			}(),
			wantError: true,
			errorMsg:  "must have a non-empty question",
		},
		{
			name: "validates answer with empty statement",
			dag: func() *dag.DAG {
				nodeId := uuid.New()
				answerId := uuid.New()
				return &dag.DAG{
					Id:    uuid.New(),
					Title: "Valid Title",
					Nodes: map[uuid.UUID]dag.Node{
						nodeId: {
							Id:       nodeId,
							Question: "Test question?",
							Answers: []dag.Answer{
								{
									Id:        answerId,
									Statement: "", // Empty statement
								},
							},
						},
					},
				}
			}(),
			wantError: true,
			errorMsg:  "must have a non-empty statement",
		},
		{
			name: "validates reference to non-existent next node",
			dag: func() *dag.DAG {
				nodeId := uuid.New()
				answerId := uuid.New()
				nonExistentNodeId := uuid.New()
				return &dag.DAG{
					Id:    uuid.New(),
					Title: "Valid Title",
					Nodes: map[uuid.UUID]dag.Node{
						nodeId: {
							Id:       nodeId,
							Question: "Test question?",
							Answers: []dag.Answer{
								{
									Id:        answerId,
									Statement: "Yes",
									NextNode:  &nonExistentNodeId, // References non-existent node
								},
							},
						},
					},
				}
			}(),
			wantError: true,
			errorMsg:  "references non-existent next node",
		},
		{
			name:      "validates correct DAG structure",
			dag:       createValidTestDAG(),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := useCase.validateDAGStructure(tt.dag)

			if tt.wantError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateDAGUseCase_Execute_ContextPropagation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDAG := createValidTestDAG()
	mockRepo := mocks.NewMockDAGRepository(ctrl)

	// Verify that context is properly passed to repository
	type testContextKey string
	contextKey := testContextKey("test-key")
	expectedCtx := context.WithValue(context.Background(), contextKey, "test-value")

	mockRepo.EXPECT().Update(expectedCtx, testDAG.Id, gomock.Any()).Return(nil)

	useCase := NewUpdateDAGUseCase(mockRepo)

	_, err := useCase.Execute(expectedCtx, CmdUpdateDAG{
		DAGId: testDAG.Id.String(),
		DAG:   testDAG,
	})

	require.NoError(t, err)
}

// createValidTestDAG creates a valid DAG for testing purposes
func createValidTestDAG() *dag.DAG {
	dagId := uuid.New()
	rootNodeId := uuid.New()
	leafNodeId := uuid.New()
	answerId1 := uuid.New()
	answerId2 := uuid.New()

	return &dag.DAG{
		Id:    dagId,
		Title: "Test Legal Case",
		Nodes: map[uuid.UUID]dag.Node{
			rootNodeId: {
				Id:       rootNodeId,
				Question: "Are you experiencing workplace discrimination?",
				Answers: []dag.Answer{
					{
						Id:        answerId1,
						Statement: "Yes, I believe so",
						NextNode:  &leafNodeId,
					},
					{
						Id:        answerId2,
						Statement: "No, not really",
						NextNode:  nil, // This makes it a leaf
					},
				},
			},
			leafNodeId: {
				Id:       leafNodeId,
				Question: "What type of discrimination?",
				Answers: []dag.Answer{
					{
						Id:        uuid.New(),
						Statement: "Age discrimination",
						NextNode:  nil, // Leaf answer
					},
					{
						Id:        uuid.New(),
						Statement: "Gender discrimination",
						NextNode:  nil, // Leaf answer
					},
				},
			},
		},
	}
}

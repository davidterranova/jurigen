package usecase

import (
	"context"
	"davidterranova/jurigen/backend/internal/model"
	"davidterranova/jurigen/backend/internal/usecase/testdata/mocks"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateStoredDAGUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	dagId := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Create a test DAG without metadata initially
	testDAG := createValidTestDAGForValidation()
	testDAG.Id = dagId
	testDAG.Metadata = nil // Start without metadata

	tests := []struct {
		name           string
		cmd            CmdValidateStoredDAG
		setupMocks     func(*mocks.MockDAGRepository)
		expectedResult func(*testing.T, *ValidationResult, error)
	}{
		{
			name: "successfully validates stored DAG and persists metadata",
			cmd: CmdValidateStoredDAG{
				DAGId: dagId.String(),
			},
			setupMocks: func(mockRepo *mocks.MockDAGRepository) {
				// Mock Get to return the DAG
				mockRepo.EXPECT().Get(ctx, dagId).Return(testDAG, nil)

				// Mock Update to persist the metadata
				mockRepo.EXPECT().Update(ctx, dagId, gomock.Any()).DoAndReturn(
					func(ctx context.Context, id uuid.UUID, updateFn func(model.DAG) (model.DAG, error)) error {
						// Simulate the update function call
						updatedDAG, err := updateFn(*testDAG)
						if err != nil {
							return err
						}

						// Verify metadata was set correctly
						assert.NotNil(t, updatedDAG.Metadata)
						assert.True(t, updatedDAG.Metadata.IsValid)
						assert.Greater(t, updatedDAG.Metadata.Statistics.TotalNodes, 0)
						assert.False(t, updatedDAG.Metadata.LastValidatedAt.IsZero())

						return nil
					})
			},
			expectedResult: func(t *testing.T, result *ValidationResult, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, result.IsValid)
				assert.Greater(t, result.Statistics.TotalNodes, 0)
				assert.Empty(t, result.Errors)
			},
		},
		{
			name: "validates DAG with validation errors and persists metadata",
			cmd: CmdValidateStoredDAG{
				DAGId: dagId.String(),
			},
			setupMocks: func(mockRepo *mocks.MockDAGRepository) {
				// Create an invalid DAG (empty title)
				invalidDAG := &model.DAG{
					Id:    dagId,
					Title: "", // Invalid: empty title
					Nodes: make(map[uuid.UUID]model.Node),
				}

				mockRepo.EXPECT().Get(ctx, dagId).Return(invalidDAG, nil)

				// Mock Update to persist the metadata
				mockRepo.EXPECT().Update(ctx, dagId, gomock.Any()).DoAndReturn(
					func(ctx context.Context, id uuid.UUID, updateFn func(model.DAG) (model.DAG, error)) error {
						updatedDAG, err := updateFn(*invalidDAG)
						if err != nil {
							return err
						}

						// Verify metadata was set with validation failure
						assert.NotNil(t, updatedDAG.Metadata)
						assert.False(t, updatedDAG.Metadata.IsValid)
						assert.False(t, updatedDAG.Metadata.LastValidatedAt.IsZero())

						return nil
					})
			},
			expectedResult: func(t *testing.T, result *ValidationResult, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.False(t, result.IsValid)
				assert.NotEmpty(t, result.Errors)
			},
		},
		{
			name: "handles invalid DAG ID format",
			cmd: CmdValidateStoredDAG{
				DAGId: "invalid-uuid",
			},
			setupMocks: func(mockRepo *mocks.MockDAGRepository) {
				// No mocks needed for this test
			},
			expectedResult: func(t *testing.T, result *ValidationResult, err error) {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidCommand)
				assert.Nil(t, result)
			},
		},
		{
			name: "handles DAG not found",
			cmd: CmdValidateStoredDAG{
				DAGId: dagId.String(),
			},
			setupMocks: func(mockRepo *mocks.MockDAGRepository) {
				mockRepo.EXPECT().Get(ctx, dagId).Return(nil, ErrNotFound)
			},
			expectedResult: func(t *testing.T, result *ValidationResult, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to retrieve DAG for validation")
				assert.Nil(t, result)
			},
		},
		{
			name: "handles repository update failure",
			cmd: CmdValidateStoredDAG{
				DAGId: dagId.String(),
			},
			setupMocks: func(mockRepo *mocks.MockDAGRepository) {
				mockRepo.EXPECT().Get(ctx, dagId).Return(testDAG, nil)
				mockRepo.EXPECT().Update(ctx, dagId, gomock.Any()).Return(assert.AnError)
			},
			expectedResult: func(t *testing.T, result *ValidationResult, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to persist validation metadata")
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockDAGRepository(ctrl)
			tt.setupMocks(mockRepo)

			useCase := NewValidateStoredDAGUseCase(mockRepo)
			result, err := useCase.Execute(ctx, tt.cmd)

			tt.expectedResult(t, result, err)
		})
	}
}

func TestValidateStoredDAGUseCase_ConvertValidationStatsToModel(t *testing.T) {
	useCase := &ValidateStoredDAGUseCase{}

	usecaseStats := ValidationStatistics{
		TotalNodes:   5,
		RootNodes:    1,
		LeafNodes:    2,
		TotalAnswers: 8,
		MaxDepth:     3,
		HasCycles:    false,
		RootNodeIDs:  []string{"root-1"},
		LeafNodeIDs:  []string{"leaf-1", "leaf-2"},
		CyclePaths:   []string{},
	}

	modelStats := useCase.convertValidationStatsToModel(usecaseStats)

	assert.Equal(t, usecaseStats.TotalNodes, modelStats.TotalNodes)
	assert.Equal(t, usecaseStats.RootNodes, modelStats.RootNodes)
	assert.Equal(t, usecaseStats.LeafNodes, modelStats.LeafNodes)
	assert.Equal(t, usecaseStats.TotalAnswers, modelStats.TotalAnswers)
	assert.Equal(t, usecaseStats.MaxDepth, modelStats.MaxDepth)
	assert.Equal(t, usecaseStats.HasCycles, modelStats.HasCycles)
	assert.Equal(t, usecaseStats.RootNodeIDs, modelStats.RootNodeIDs)
	assert.Equal(t, usecaseStats.LeafNodeIDs, modelStats.LeafNodeIDs)
	assert.Equal(t, usecaseStats.CyclePaths, modelStats.CyclePaths)
}

func TestValidateStoredDAGUseCase_UpdateExistingMetadata(t *testing.T) {
	ctx := context.Background()
	dagId := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Create a test DAG with existing metadata
	testDAG := createValidTestDAGForValidation()
	testDAG.Id = dagId
	testDAG.Metadata = &model.DAGMetadata{
		IsValid:         false,                                     // Previously invalid
		Statistics:      model.ValidationStatistics{TotalNodes: 0}, // Old stats
		LastValidatedAt: time.Now().Add(-24 * time.Hour),           // Old validation time
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockDAGRepository(ctrl)
	mockRepo.EXPECT().Get(ctx, dagId).Return(testDAG, nil)

	mockRepo.EXPECT().Update(ctx, dagId, gomock.Any()).DoAndReturn(
		func(ctx context.Context, id uuid.UUID, updateFn func(model.DAG) (model.DAG, error)) error {
			updatedDAG, err := updateFn(*testDAG)
			require.NoError(t, err)

			// Verify metadata was updated correctly
			assert.NotNil(t, updatedDAG.Metadata)
			assert.True(t, updatedDAG.Metadata.IsValid)                     // Should now be valid
			assert.Greater(t, updatedDAG.Metadata.Statistics.TotalNodes, 0) // Updated stats

			// Validation time should be updated (not zero time)
			assert.False(t, updatedDAG.Metadata.LastValidatedAt.IsZero(), "Validation time should be updated")

			return nil
		})

	useCase := NewValidateStoredDAGUseCase(mockRepo)
	result, err := useCase.Execute(ctx, CmdValidateStoredDAG{
		DAGId: dagId.String(),
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsValid)
}

// Helper function to create a valid test DAG for validation testing
func createValidTestDAGForValidation() *model.DAG {
	rootNodeId := uuid.New()
	leafNodeId := uuid.New()

	dag := model.NewDAG("Test Valid DAG")

	rootNode := model.Node{
		Id:       rootNodeId,
		Question: "Is this a test?",
		Answers: []model.Answer{
			{
				Id:        uuid.New(),
				Statement: "Yes",
				NextNode:  &leafNodeId,
			},
			{
				Id:        uuid.New(),
				Statement: "No",
				NextNode:  nil,
			},
		},
	}

	leafNode := model.Node{
		Id:       leafNodeId,
		Question: "Final question?",
		Answers: []model.Answer{
			{
				Id:        uuid.New(),
				Statement: "Done",
				NextNode:  nil,
			},
		},
	}

	// Set parent pointers
	for i := range rootNode.Answers {
		rootNode.Answers[i].ParentNode = &rootNode
	}
	for i := range leafNode.Answers {
		leafNode.Answers[i].ParentNode = &leafNode
	}

	dag.Nodes[rootNodeId] = rootNode
	dag.Nodes[leafNodeId] = leafNode

	return dag
}

package usecase

import (
	"context"
	"davidterranova/jurigen/backend/internal/model"
	"davidterranova/jurigen/backend/internal/usecase/testdata/mocks"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGetDAGUseCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockDAGRepository(ctrl)
	useCase := NewGetDAGUseCase(mockRepo)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.dagRepository)
	assert.NotNil(t, useCase.validator)
}

func TestGetDAGUseCase_Execute(t *testing.T) {
	testDAG := model.NewDAG("Test DAG")
	validUUID := testDAG.Id.String()

	tests := []struct {
		name        string
		cmd         CmdGetDAG
		setupMock   func(*mocks.MockDAGRepository)
		expectedDAG *model.DAG
		expectError bool
		errorType   error
	}{
		{
			name: "successfully retrieves existing DAG",
			cmd: CmdGetDAG{
				DAGId: validUUID,
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				mockRepo.EXPECT().Get(gomock.Any(), testDAG.Id).Return(testDAG, nil)
			},
			expectedDAG: testDAG,
			expectError: false,
		},
		{
			name: "returns validation error for empty DAG ID",
			cmd: CmdGetDAG{
				DAGId: "",
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				// No repository call expected due to validation failure
			},
			expectError: true,
			errorType:   ErrInvalidCommand,
		},
		{
			name: "returns validation error for invalid UUID format",
			cmd: CmdGetDAG{
				DAGId: "invalid-uuid-format",
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				// No repository call expected due to validation failure
			},
			expectError: true,
			errorType:   ErrInvalidCommand,
		},
		{
			name: "returns validation error for non-UUID string",
			cmd: CmdGetDAG{
				DAGId: "not-a-uuid-at-all",
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				// No repository call expected due to validation failure
			},
			expectError: true,
			errorType:   ErrInvalidCommand,
		},
		{
			name: "returns not found error when DAG doesn't exist",
			cmd: CmdGetDAG{
				DAGId: validUUID,
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				mockRepo.EXPECT().Get(gomock.Any(), testDAG.Id).Return(nil, ErrNotFound)
			},
			expectError: true,
			errorType:   ErrNotFound,
		},
		{
			name: "returns internal error when repository fails",
			cmd: CmdGetDAG{
				DAGId: validUUID,
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				mockRepo.EXPECT().Get(gomock.Any(), testDAG.Id).Return(nil, ErrInternal)
			},
			expectError: true,
			errorType:   ErrInternal,
		},
		{
			name: "handles valid UUID with different format",
			cmd: CmdGetDAG{
				DAGId: "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				expectedUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				retrievedDAG := model.NewDAG("Test DAG")
				retrievedDAG.Id = expectedUUID
				mockRepo.EXPECT().Get(gomock.Any(), expectedUUID).Return(retrievedDAG, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockDAGRepository(ctrl)
			tt.setupMock(mockRepo)

			useCase := NewGetDAGUseCase(mockRepo)
			ctx := context.Background()

			result, err := useCase.Get(ctx, tt.cmd)

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
				assert.Equal(t, tt.expectedDAG.Nodes, result.Nodes)
			}
		})
	}
}

func TestGetDAGUseCase_Execute_ValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		cmd       CmdGetDAG
		setupMock func(*mocks.MockDAGRepository)
	}{
		{
			name: "validates UUID format with lowercase (validator requires lowercase)",
			cmd: CmdGetDAG{
				DAGId: "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				expectedUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				testDAG := model.NewDAG("Test DAG")
				testDAG.Id = expectedUUID
				mockRepo.EXPECT().Get(gomock.Any(), expectedUUID).Return(testDAG, nil)
			},
		},
		{
			name: "validates UUID format without hyphens fails",
			cmd: CmdGetDAG{
				DAGId: "550e8400e29b41d4a716446655440000",
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				// No repository call expected due to validation failure
			},
		},
		{
			name: "validates empty spaces in UUID fails",
			cmd: CmdGetDAG{
				DAGId: " 550e8400-e29b-41d4-a716-446655440000 ",
			},
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				// No repository call expected due to validation failure
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockDAGRepository(ctrl)
			tt.setupMock(mockRepo)

			useCase := NewGetDAGUseCase(mockRepo)
			ctx := context.Background()

			_, err := useCase.Get(ctx, tt.cmd)

			// Test cases with invalid UUIDs should fail validation
			if tt.cmd.DAGId == "550e8400e29b41d4a716446655440000" ||
				tt.cmd.DAGId == " 550e8400-e29b-41d4-a716-446655440000 " {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidCommand)
			} else {
				// Valid UUID format should succeed (assuming repository succeeds)
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDAGUseCase_Execute_ContextPropagation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDAG := model.NewDAG("Test DAG")
	mockRepo := mocks.NewMockDAGRepository(ctrl)

	// Verify that context is properly passed to repository
	//nolint:staticcheck // This is a valid use of context.WithValue (for test purpose only)
	expectedCtx := context.WithValue(context.Background(), "test-key", "test-value")
	mockRepo.EXPECT().Get(expectedCtx, testDAG.Id).Return(testDAG, nil)

	useCase := NewGetDAGUseCase(mockRepo)

	result, err := useCase.Get(expectedCtx, CmdGetDAG{
		DAGId: testDAG.Id.String(),
	})

	require.NoError(t, err)
	assert.Equal(t, testDAG.Id, result.Id)
}

func TestGetDAGUseCase_Execute_RepositoryErrorWrapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testDAG := model.NewDAG("Test DAG")
	mockRepo := mocks.NewMockDAGRepository(ctrl)

	// Test that repository errors are properly propagated
	repositoryError := fmt.Errorf("database connection failed")
	mockRepo.EXPECT().Get(gomock.Any(), testDAG.Id).Return(nil, repositoryError)

	useCase := NewGetDAGUseCase(mockRepo)
	ctx := context.Background()

	result, err := useCase.Get(ctx, CmdGetDAG{
		DAGId: testDAG.Id.String(),
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repositoryError, err)
}

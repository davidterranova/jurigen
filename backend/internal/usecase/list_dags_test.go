package usecase

import (
	"context"
	"davidterranova/jurigen/backend/internal/usecase/testdata/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewListDAGsUseCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockDAGRepository(ctrl)
	useCase := NewListDAGsUseCase(mockRepo)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockRepo, useCase.dagRepository)
}

func TestListDAGsUseCase_List(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockDAGRepository)
		expectedResult []uuid.UUID
		expectError    bool
	}{
		{
			name: "successfully returns list of DAG IDs",
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				dagIds := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
				mockRepo.EXPECT().List(gomock.Any()).Return(dagIds, nil)
			},
			expectError: false,
		},
		{
			name: "returns empty list when no DAGs exist",
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				mockRepo.EXPECT().List(gomock.Any()).Return([]uuid.UUID{}, nil)
			},
			expectedResult: []uuid.UUID{},
			expectError:    false,
		},
		{
			name: "returns error when repository fails",
			setupMock: func(mockRepo *mocks.MockDAGRepository) {
				mockRepo.EXPECT().List(gomock.Any()).Return([]uuid.UUID{}, ErrInternal)
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

			useCase := NewListDAGsUseCase(mockRepo)
			ctx := context.Background()

			result, err := useCase.List(ctx, CmdListDAGs{})

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.expectedResult != nil {
				assert.Equal(t, tt.expectedResult, result)
			} else {
				// For the case with random UUIDs, just check length
				assert.Len(t, result, 3)
			}
		})
	}
}

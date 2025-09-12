package http

import (
	"davidterranova/jurigen/backend/internal/adapter/http/testdata/mocks"
	"davidterranova/jurigen/backend/internal/model"
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

func TestNewDAGHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockApp(ctrl)
	handler := NewDAGHandler(mockApp)

	assert.NotNil(t, handler)
	assert.Equal(t, mockApp, handler.app)
}

func TestDAGHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockApp)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successfully returns list of DAGs with summary info",
			setupMock: func(mockApp *mocks.MockApp) {
				dagId1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
				dagId2 := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

				dag1 := &model.DAG{
					Id:    dagId1,
					Title: "Employment Law Case",
					Nodes: make(map[uuid.UUID]model.Node),
					Metadata: &model.DAGMetadata{
						IsValid:    true,
						Statistics: model.ValidationStatistics{TotalNodes: 5},
					},
				}
				dag2 := &model.DAG{
					Id:       dagId2,
					Title:    "Contract Dispute",
					Nodes:    make(map[uuid.UUID]model.Node),
					Metadata: nil, // No metadata should default to invalid
				}

				dags := []*model.DAG{dag1, dag2}
				mockApp.EXPECT().ListDAGs(gomock.Any(), usecase.CmdListDAGs{}).Return(dags, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response DAGSummaryListPresenter
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Len(t, response.DAGs, 2)
				assert.Equal(t, 2, response.Count)

				// Check first DAG
				assert.Equal(t, uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), response.DAGs[0].Id)
				assert.Equal(t, "Employment Law Case", response.DAGs[0].Title)
				assert.True(t, response.DAGs[0].IsValid)

				// Check second DAG
				assert.Equal(t, uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"), response.DAGs[1].Id)
				assert.Equal(t, "Contract Dispute", response.DAGs[1].Title)
				assert.False(t, response.DAGs[1].IsValid) // Should be false when metadata is nil
			},
		},
		{
			name: "returns empty list when no DAGs exist",
			setupMock: func(mockApp *mocks.MockApp) {
				mockApp.EXPECT().ListDAGs(gomock.Any(), usecase.CmdListDAGs{}).Return([]*model.DAG{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response DAGSummaryListPresenter
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Len(t, response.DAGs, 0)
				assert.Equal(t, 0, response.Count)
			},
		},
		{
			name: "returns 500 when app layer fails",
			setupMock: func(mockApp *mocks.MockApp) {
				mockApp.EXPECT().ListDAGs(gomock.Any(), usecase.CmdListDAGs{}).Return(nil, usecase.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "failed to list DAGs")
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

			req, err := http.NewRequest("GET", "/v1/dags", nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			handler.List(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestDAGHandler_GetDAG(t *testing.T) {
	testDAG := model.NewDAG("Test DAG")

	tests := []struct {
		name           string
		dagId          string
		setupMock      func(*mocks.MockApp)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:  "successfully returns DAG",
			dagId: testDAG.Id.String(),
			setupMock: func(mockApp *mocks.MockApp) {
				mockApp.EXPECT().Get(gomock.Any(), usecase.CmdGetDAG{DAGId: testDAG.Id.String()}).Return(testDAG, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response DAGPresenter
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, testDAG.Id, response.Id)
			},
		},
		{
			name:  "returns 400 for invalid UUID",
			dagId: "invalid-uuid",
			setupMock: func(mockApp *mocks.MockApp) {
				mockApp.EXPECT().Get(gomock.Any(), usecase.CmdGetDAG{DAGId: "invalid-uuid"}).Return(nil, usecase.ErrInvalidCommand)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "invalid DAG ID format")
			},
		},
		{
			name:  "returns 404 when DAG not found",
			dagId: testDAG.Id.String(),
			setupMock: func(mockApp *mocks.MockApp) {
				mockApp.EXPECT().Get(gomock.Any(), usecase.CmdGetDAG{DAGId: testDAG.Id.String()}).Return(nil, usecase.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Body.String(), "DAG not found")
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

			req, err := http.NewRequest("GET", "/v1/dags/"+tt.dagId, nil)
			require.NoError(t, err)

			// Set up mux vars
			req = mux.SetURLVars(req, map[string]string{"dagId": tt.dagId})

			rr := httptest.NewRecorder()

			handler.Get(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestDAGListPresenter(t *testing.T) {
	tests := []struct {
		name     string
		dagIds   []uuid.UUID
		expected DAGListPresenter
	}{
		{
			name:   "creates presenter with multiple DAGs",
			dagIds: []uuid.UUID{uuid.New(), uuid.New()},
			expected: DAGListPresenter{
				Count: 2,
			},
		},
		{
			name:   "creates presenter with empty list",
			dagIds: []uuid.UUID{},
			expected: DAGListPresenter{
				DAGs:  []uuid.UUID{},
				Count: 0,
			},
		},
		{
			name:   "creates presenter with single DAG",
			dagIds: []uuid.UUID{uuid.New()},
			expected: DAGListPresenter{
				Count: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenter := NewDAGListPresenter(tt.dagIds)

			assert.Equal(t, tt.expected.Count, presenter.Count)
			assert.Len(t, presenter.DAGs, tt.expected.Count)

			if tt.expected.DAGs != nil {
				assert.Equal(t, tt.expected.DAGs, presenter.DAGs)
			}
		})
	}
}

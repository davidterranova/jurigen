package http

import (
	"davidterranova/jurigen/internal/adapter/http/testdata/mocks"
	"davidterranova/jurigen/internal/dag"
	"davidterranova/jurigen/internal/usecase"
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

func TestDAGHandler_ListDAGs(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockApp)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successfully returns list of DAG IDs",
			setupMock: func(mockApp *mocks.MockApp) {
				dagIds := []uuid.UUID{
					uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
					uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"),
				}
				mockApp.EXPECT().ListDAGs(gomock.Any(), usecase.CmdListDAGs{}).Return(dagIds, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response DAGListPresenter
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Len(t, response.DAGs, 2)
				assert.Equal(t, 2, response.Count)
				assert.Contains(t, response.DAGs, uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"))
				assert.Contains(t, response.DAGs, uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
			},
		},
		{
			name: "returns empty list when no DAGs exist",
			setupMock: func(mockApp *mocks.MockApp) {
				mockApp.EXPECT().ListDAGs(gomock.Any(), usecase.CmdListDAGs{}).Return([]uuid.UUID{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response DAGListPresenter
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Len(t, response.DAGs, 0)
				assert.Equal(t, 0, response.Count)
			},
		},
		{
			name: "returns 500 when app layer fails",
			setupMock: func(mockApp *mocks.MockApp) {
				mockApp.EXPECT().ListDAGs(gomock.Any(), usecase.CmdListDAGs{}).Return([]uuid.UUID{}, usecase.ErrInternal)
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

			handler.ListDAGs(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestDAGHandler_GetDAG(t *testing.T) {
	testDAG := dag.NewDAG()

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
				mockApp.EXPECT().GetDAG(gomock.Any(), usecase.CmdGetDAG{DAGId: testDAG.Id.String()}).Return(testDAG, nil)
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
				mockApp.EXPECT().GetDAG(gomock.Any(), usecase.CmdGetDAG{DAGId: "invalid-uuid"}).Return(nil, usecase.ErrInvalidCommand)
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
				mockApp.EXPECT().GetDAG(gomock.Any(), usecase.CmdGetDAG{DAGId: testDAG.Id.String()}).Return(nil, usecase.ErrNotFound)
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

			handler.GetDAG(rr, req)

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

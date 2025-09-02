package http

import (
	"context"
	"davidterranova/jurigen/internal/dag"
	"davidterranova/jurigen/internal/usecase"
	"davidterranova/jurigen/pkg/xhttp"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

//go:generate go run github.com/golang/mock/mockgen -source=dag.go -destination=testdata/mocks/app_mock.go -package=mocks

type App interface {
	GetDAG(ctx context.Context, cmd usecase.CmdGetDAG) (*dag.DAG, error)
	ListDAGs(ctx context.Context, cmd usecase.CmdListDAGs) ([]uuid.UUID, error)
}

type dagHandler struct {
	app App
}

func NewDAGHandler(app App) *dagHandler {
	return &dagHandler{
		app: app,
	}
}

// GetDAG retrieves a specific Legal Case DAG by its unique identifier
//
// @Summary Get Legal Case DAG
// @Description Retrieve a complete Legal Case DAG structure including all questions, answers, and collected context
// @Tags DAGs
// @Accept json
// @Produce json
// @Param dagId path string true "DAG unique identifier (UUID)"
// @Success 200 {object} DAGPresenter "Successfully retrieved DAG"
// @Failure 400 {object} xhttp.ErrorResponse "Invalid DAG ID format"
// @Failure 404 {object} xhttp.ErrorResponse "DAG not found"
// @Failure 500 {object} xhttp.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /dags/{dagId} [get]
func (h *dagHandler) GetDAG(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := mux.Vars(r)[dagId]

	dag, err := h.app.GetDAG(ctx, usecase.CmdGetDAG{
		DAGId: id,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to get DAG")
		switch {
		case errors.Is(err, usecase.ErrInvalidCommand):
			xhttp.WriteError(ctx, w, http.StatusBadRequest, "invalid DAG ID format", err)
			return
		case errors.Is(err, usecase.ErrNotFound):
			xhttp.WriteError(ctx, w, http.StatusNotFound, "DAG not found", err)
			return
		default:
			xhttp.WriteError(ctx, w, http.StatusInternalServerError, "failed to get DAG", err)
			return
		}
	}

	xhttp.WriteObject(ctx, w, http.StatusOK, NewDAGPresenter(dag))
}

// ListDAGs retrieves all available Legal Case DAG identifiers
//
// @Summary List Legal Case DAGs
// @Description Retrieve a list of all available Legal Case DAG identifiers in the system
// @Tags DAGs
// @Accept json
// @Produce json
// @Success 200 {object} DAGListPresenter "Successfully retrieved DAG list"
// @Failure 500 {object} xhttp.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /dags [get]
func (h *dagHandler) ListDAGs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dagIds, err := h.app.ListDAGs(ctx, usecase.CmdListDAGs{})
	if err != nil {
		log.Error().Err(err).Msg("failed to list DAGs")
		xhttp.WriteError(ctx, w, http.StatusInternalServerError, "failed to list DAGs", err)
		return
	}

	xhttp.WriteObject(ctx, w, http.StatusOK, NewDAGListPresenter(dagIds))
}

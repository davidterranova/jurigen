package http

import (
	"context"
	"davidterranova/jurigen/internal/dag"
	"davidterranova/jurigen/internal/usecase"
	"davidterranova/jurigen/pkg/xhttp"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type App interface {
	GetDAG(ctx context.Context, cmd usecase.CmdGetDAG) (*dag.DAG, error)
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

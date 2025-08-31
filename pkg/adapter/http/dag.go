package http

import (
	"context"
	"davidterranova/jurigen/internal/dag"
	"davidterranova/jurigen/pkg/usecase"
	"davidterranova/jurigen/pkg/xhttp"
	"net/http"

	"github.com/gorilla/mux"
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

func (h *dagHandler) GetDAG(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := mux.Vars(r)[dagId]

	dag, err := h.app.GetDAG(ctx, usecase.CmdGetDAG{
		DAGId: id,
	})
	if err != nil {
		xhttp.WriteError(ctx, w, http.StatusInternalServerError, "failed to get DAG", err)
		return
	}

	xhttp.WriteObject(ctx, w, http.StatusOK, NewDAGPresenter(dag))
}

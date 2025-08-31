package http

import (
	"davidterranova/jurigen/pkg/xhttp"
	"net/http"

	"github.com/gorilla/mux"
)

const dagId = "dagId"

func New(app App, authFn xhttp.AuthFn) *mux.Router {
	root := mux.NewRouter()
	mountV1DAG(root, authFn, app)

	return root
}

func mountV1DAG(router *mux.Router, authFn xhttp.AuthFn, app App) {
	dagHandler := NewDAGHandler(app)
	v1 := router.PathPrefix("/v1/dags").Subrouter()

	if authFn != nil {
		v1.Use(xhttp.AuthMiddleware(authFn))
	}

	v1.HandleFunc("/{"+dagId+"}", dagHandler.GetDAG).Methods(http.MethodGet)
}

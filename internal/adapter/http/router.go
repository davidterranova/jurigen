package http

import (
	"davidterranova/jurigen/pkg/xhttp"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	// Import generated docs
	_ "davidterranova/jurigen/docs/swagger"
)

const dagId = "dagId"

func New(app App, authFn xhttp.AuthFn) *mux.Router {
	root := mux.NewRouter()
	mountV1DAG(root, authFn, app)
	mountSwaggerUI(root)

	return root
}

func mountV1DAG(router *mux.Router, authFn xhttp.AuthFn, app App) {
	dagHandler := NewDAGHandler(app)
	v1 := router.PathPrefix("/v1/dags").Subrouter()

	if authFn != nil {
		v1.Use(xhttp.AuthMiddleware(authFn))
	}

	v1.HandleFunc("", dagHandler.List).Methods(http.MethodGet)
	v1.HandleFunc("/{"+dagId+"}", dagHandler.Get).Methods(http.MethodGet)
}

// mountSwaggerUI mounts the Swagger UI documentation endpoint
func mountSwaggerUI(router *mux.Router) {
	// Serve Swagger UI at /swagger/
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler).Methods(http.MethodGet)
}

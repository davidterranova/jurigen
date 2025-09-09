package xhttp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// ErrorResponse represents an API error response
//
// @Description Standard error response format for API failures
// @Example {"message": "failed to get DAG", "error": "DAG not found"}
type ErrorResponse struct {
	Message string `json:"message" example:"failed to get DAG" description:"Human-readable error description"`
	Error   string `json:"error" example:"DAG not found" description:"Technical error details"`
}

func WriteObject(ctx context.Context, w http.ResponseWriter, status int, obj any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(obj)
	if err != nil {
		log.
			Err(err).
			Msg("failed to write json object")
	}
}

func WriteError(ctx context.Context, w http.ResponseWriter, status int, contextualMessage string, err error) {
	WriteObject(
		ctx,
		w,
		status,
		ErrorResponse{
			Message: contextualMessage,
			Error:   err.Error(),
		},
	)
}

func Heartbeat(w http.ResponseWriter, r *http.Request) {
	WriteObject(context.Background(), w, http.StatusOK, []byte(`{"status": "ok"}`))
}

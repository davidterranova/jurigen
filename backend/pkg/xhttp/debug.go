package xhttp

import (
	"context"
	"net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func ServeDebugHandlers(ctx context.Context, port int) {
	router := mux.NewRouter()

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	router.HandleFunc("/debug/pprof/{cmd}", pprof.Index)

	httpServer := NewServer(router, "", port)
	err := httpServer.Serve(ctx)
	if err != nil {
		log.
			Panic().
			Err(err).
			Msg("failed to start http server")
	}
	log.Info().Int("port", port).Msg("debug http server started")
}

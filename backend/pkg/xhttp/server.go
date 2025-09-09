package xhttp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// DefaultWriteTimeout for the http server
	DefaultWriteTimeout = 5 * time.Second

	// DefaultReadTimeout for the http server
	DefaultReadTimeout = 5 * time.Second
)

// Server is a filestorage http server
type Server struct {
	host    string
	port    int
	handler http.Handler
}

// NewServer creates a new http server given a handler and a configuration
func NewServer(handler http.Handler, host string, port int) *Server {
	return &Server{
		host:    host,
		port:    port,
		handler: handler,
	}
}

// Address returns the host and port expected from an http server
func (s Server) Address() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

// Serve starts the server
func (s Server) Serve(ctx context.Context) error {
	srv := http.Server{
		Addr:              s.Address(),
		Handler:           CORS()(s.handler),
		WriteTimeout:      DefaultWriteTimeout,
		ReadTimeout:       DefaultReadTimeout,
		ReadHeaderTimeout: DefaultReadTimeout,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.
				Fatal().
				Err(err).
				Msg("http server crashed")
		}
	}()

	log.
		Info().
		Str("address", s.Address()).
		Msg("http server started")

	<-ctx.Done()

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(ctxShutDown)
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to shutdown http server properly: %s", err)
	}

	return nil
}

package server

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1"
	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/server/v1/http"

	"github.com/google/wire"
)

type (
	// Server is our API server
	Server struct {
		logger     logging.Logger
		config     *config.ServerConfig
		httpServer *httpserver.Server
	}
)

var (
	// Providers is our wire superset of providers this package offers
	Providers = wire.NewSet(
		ProvideServer,
	)
)

// ProvideServer builds a new Server instance
func ProvideServer(
	logger logging.Logger,
	cfg *config.ServerConfig,
	httpServer *httpserver.Server,
) (*Server, error) {
	srv := &Server{
		config:     cfg,
		httpServer: httpServer,
		logger:     logger,
	}

	return srv, nil
}

// Serve serves HTTP traffic
func (s *Server) Serve() {
	s.httpServer.Serve()
}

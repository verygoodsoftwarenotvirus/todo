package server

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/server/v1/http"

	"github.com/google/wire"
)

type (
	// Server is the structure responsible for hosting all available protocols
	// In the events we adopted a gRPC implementation of the surface, this is
	// the structure that would contain it and be responsible for calling its
	// serve method
	Server struct {
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
func ProvideServer(cfg *config.ServerConfig, httpServer *httpserver.Server) (*Server, error) {
	srv := &Server{
		config:     cfg,
		httpServer: httpServer,
	}

	return srv, nil
}

// Serve serves HTTP traffic
func (s *Server) Serve() {
	s.httpServer.Serve()
}

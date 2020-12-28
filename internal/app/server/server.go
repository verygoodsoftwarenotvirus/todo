package server

import (
	"errors"

	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"

	"github.com/google/wire"
)

type (
	// Server is the structure responsible for hosting all available protocols.
	// In the events we adopted a gRPC implementation of the surface, this is
	// the structure that would contain it and be responsible for calling its
	// serve method.
	Server struct {
		config     *config.ServerConfig
		httpServer *httpserver.Server
	}
)

// Providers is our wire superset of providers this package offers.
var Providers = wire.NewSet(
	ProvideServer,
)

// ProvideServer builds a new Server instance.
func ProvideServer(cfg *config.ServerConfig, httpServer *httpserver.Server) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("provided config was nil")
	} else if httpServer == nil {
		return nil, errors.New("provided http server was nil")
	}

	srv := &Server{
		config:     cfg,
		httpServer: httpServer,
	}

	return srv, nil
}

// Serve serves HTTP traffic.
func (s *Server) Serve() {
	s.httpServer.Serve()
}

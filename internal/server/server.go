package server

import (
	"errors"

	httpserver "gitlab.com/verygoodsoftwarenotvirus/todo/internal/server/http"

	"github.com/google/wire"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
)

type (
	// Server is the structure responsible for hosting all available protocols. In the event
	// we adopted, say, a gRPC implementation of the surface, this is the structure that would
	// contain it and be responsible for calling its serve method.
	Server struct {
		config     *config.ServerConfig
		httpServer *httpserver.Server
	}
)

var (
	// Providers is our wire superset of providers this package offers.
	Providers = wire.NewSet(
		ProvideServer,
	)

	errNilServer       = errors.New("provided HTTP server was nil")
	errNilServerConfig = errors.New("provided server config was nil")
)

// ProvideServer builds a new Server instance.
func ProvideServer(cfg *config.ServerConfig, httpServer *httpserver.Server) (*Server, error) {
	if cfg == nil {
		return nil, errNilServerConfig
	}

	if httpServer == nil {
		return nil, errNilServer
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

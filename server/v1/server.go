package server

import (
	"context"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1/grpc"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1/http"

	"github.com/google/wire"
)

const (
	maxTimeout = 120 * time.Second
)

type (
	// Server is our API server
	Server struct {
		logger     logging.Logger
		config     *config.ServerConfig
		grpcServer *grpcserver.GRPCServer
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
	database database.Database,
	logger logging.Logger,
	config *config.ServerConfig,
	grpcServer *grpcserver.GRPCServer,
	httpServer *httpserver.Server,
) (*Server, error) {

	srv := &Server{
		config:     config,
		grpcServer: grpcServer,
		httpServer: httpServer,
		logger:     logger,
	}

	logger.Info("migrating database")
	if err := database.Migrate(context.Background()); err != nil {
		return nil, err
	}
	logger.Info("database migrated!")

	return srv, nil
}

// Serve serves HTTP traffic
func (s *Server) Serve() {
	go s.grpcServer.Serve()
	s.httpServer.Serve()
}

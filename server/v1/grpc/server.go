package grpcserver

import (
	"fmt"
	"net"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	// "gitlab.com/verygoodsoftwarenotvirus/todo/proto/v1"

	"google.golang.org/grpc"
)

// var _ todoproto.TodoServer = (*GRPCServer)(nil)

// GRPCServer is a gRPC server implementation
type GRPCServer struct {
	// services
	itemsService         *items.Service
	usersService         *users.Service
	oauth2ClientsService *oauth2clients.Service

	// infra
	port       uint16
	logger     logging.Logger
	grpcServer grpc.Server
}

// ProvideGRPCServer provides a gRPC compatible Todo server for dependency injection
func ProvideGRPCServer(
	// services
	logger logging.Logger,
	itemsService *items.Service,
	usersService *users.Service,
	oauth2Service *oauth2clients.Service,
) (*GRPCServer, error) {
	return &GRPCServer{
		logger:               logger,
		itemsService:         itemsService,
		usersService:         usersService,
		oauth2ClientsService: oauth2Service,
	}, nil
}

// Serve serves
func (s *GRPCServer) Serve() {
	if s.port == 0 {
		s.port = 8888
	}

	grpcAddr := fmt.Sprintf(":%d", s.port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		s.logger.Fatal(err)
	}
	defer lis.Close()

	s.logger.WithValue("grpcAddress", grpcAddr).Debug("listening for grpc requests")

	s.grpcServer.Serve(lis)
}

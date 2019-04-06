package grpcserver

import (
	"context"
	"fmt"
	"net"

	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/proto/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	"google.golang.org/grpc"
)

var _ todoproto.TodoServer = (*GRPCServer)(nil)

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
	itemsService *items.Service,
	usersService *users.Service,
	oauth2Service *oauth2clients.Service,

) (*GRPCServer, error) {
	return &GRPCServer{
		itemsService:         itemsService,
		usersService:         usersService,
		oauth2ClientsService: oauth2Service,
	}, nil
}

// Serve serves
func (s *GRPCServer) Serve() {
	grpcAddr := fmt.Sprintf(":%d", s.port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		s.logger.Fatal(err)
	}
	defer lis.Close()

	s.grpcServer.Serve(lis)
}

// GetItem implements our requisite gRPC server method
func (s *GRPCServer) GetItem(ctx context.Context, req *todoproto.GetItemRequest) (*todoproto.GetItemResponse, error) {
	return s.itemsService.GetItem(ctx, req)
}

// GetItemCount implements our requisite gRPC server method
func (s *GRPCServer) GetItemCount(ctx context.Context, req *todoproto.ItemListRequest) (*todoproto.ItemCountResponse, error) {
	return s.itemsService.GetItemCount(ctx, req)
}

// GetItems implements our requisite gRPC server method
func (s *GRPCServer) GetItems(ctx context.Context, req *todoproto.ItemListRequest) (*todoproto.ItemListResponse, error) {
	return s.itemsService.GetItems(ctx, req)
}

// CreateItem implements our requisite gRPC server method
func (s *GRPCServer) CreateItem(ctx context.Context, req *todoproto.CreateItemRequest) (*todoproto.CreateItemResponse, error) {
	return s.itemsService.CreateItem(ctx, req)
}

// UpdateItem implements our requisite gRPC server method
func (s *GRPCServer) UpdateItem(ctx context.Context, req *todoproto.UpdateItemRequest) (*todoproto.UpdateItemResponse, error) {
	return s.itemsService.UpdateItem(ctx, req)
}

// DeleteItem implements our requisite gRPC server method
func (s *GRPCServer) DeleteItem(ctx context.Context, req *todoproto.DeleteItemRequest) (*todoproto.ErrorResponse, error) {
	return s.itemsService.DeleteItem(ctx, req)
}

// GetUser implements our requisite gRPC server method
func (s *GRPCServer) GetUser(ctx context.Context, req *todoproto.GetUserRequest) (*todoproto.GetUserResponse, error) {
	return s.usersService.GetUser(ctx, req)
}

// GetUserByUsername implements our requisite gRPC server method
func (s *GRPCServer) GetUserByUsername(ctx context.Context, req *todoproto.GetUserByUsernameRequest) (*todoproto.GetUserResponse, error) {
	return s.usersService.GetUserByUsername(ctx, req)
}

// GetUserCount implements our requisite gRPC server method
func (s *GRPCServer) GetUserCount(ctx context.Context, req *todoproto.UserListRequest) (*todoproto.UserCountResponse, error) {
	return s.usersService.GetUserCount(ctx, req)
}

// GetUsers implements our requisite gRPC server method
func (s *GRPCServer) GetUsers(ctx context.Context, req *todoproto.UserListRequest) (*todoproto.UserListResponse, error) {
	return s.usersService.GetUsers(ctx, req)
}

// CreateUser implements our requisite gRPC server method
func (s *GRPCServer) CreateUser(ctx context.Context, req *todoproto.CreateUserRequest) (*todoproto.CreateUserResponse, error) {
	return s.usersService.CreateUser(ctx, req)
}

// func (s *GRPCServer) UpdateUserPassword(ctx context.Context, req *todoproto.UpdateUserPasswordRequest) (*todoproto.ErrorResponse, error) {
// 	return s.usersService.UpdateUserPassword(ctx, req)
// }

// DeleteUser implements our requisite gRPC server method
func (s *GRPCServer) DeleteUser(ctx context.Context, req *todoproto.DeleteUserRequest) (*todoproto.ErrorResponse, error) {
	return s.usersService.DeleteUser(ctx, req)
}

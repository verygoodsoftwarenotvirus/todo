package grpcserver

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/proto/v1"
)

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

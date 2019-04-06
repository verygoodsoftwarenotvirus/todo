package users

import (
	"context"

	// "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/proto/v1"
)

// GetUser implements our gRPC server's interface
func (s *Service) GetUser(ctx context.Context, in *todoproto.GetUserRequest) (*todoproto.GetUserResponse, error) {
	user, err := s.database.GetUser(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return &todoproto.GetUserResponse{
		// DELTEME? User: todoproto.ProtoUserFromModel(user),
		Id:                    user.ID,
		Username:              user.Username,
		IsAdmin:               user.IsAdmin,
		PasswordLastChangedOn: *user.PasswordLastChangedOn,
		CreatedOn:             user.CreatedOn,
		UpdatedOn:             *user.UpdatedOn,
		ArchivedOn:            *user.ArchivedOn,
	}, nil
}

// GetUserByUsername implements our gRPC server's interface
func (s *Service) GetUserByUsername(ctx context.Context, in *todoproto.GetUserByUsernameRequest) (*todoproto.GetUserResponse, error) {
	user, err := s.database.GetUserByUsername(ctx, in.Username)
	if err != nil {
		return nil, err
	}

	return &todoproto.GetUserResponse{
		Id:                    user.ID,
		Username:              user.Username,
		IsAdmin:               user.IsAdmin,
		PasswordLastChangedOn: *user.PasswordLastChangedOn,
		CreatedOn:             user.CreatedOn,
		UpdatedOn:             *user.UpdatedOn,
		ArchivedOn:            *user.ArchivedOn,
	}, nil
}

// GetUserCount implements our gRPC server's interface
func (s *Service) GetUserCount(ctx context.Context, in *todoproto.UserListRequest) (*todoproto.UserCountResponse, error) {
	count, err := s.database.GetUserCount(ctx, in.Filter.ToModelQueryFilter())
	if err != nil {
		return nil, err
	}

	return &todoproto.UserCountResponse{
		Count: count,
	}, nil
}

// GetUsers implements our gRPC server's interface
func (s *Service) GetUsers(ctx context.Context, in *todoproto.UserListRequest) (*todoproto.UserListResponse, error) {
	users, err := s.database.GetUsers(ctx, in.Filter.ToModelQueryFilter())
	if err != nil {
		return nil, err
	}

	res := &todoproto.UserListResponse{
		Users: todoproto.ProtoUsersFromModels(users.Users),
	}

	return res, nil
}

// CreateUser implements our gRPC server's interface
func (s *Service) CreateUser(ctx context.Context, in *todoproto.CreateUserRequest) (*todoproto.CreateUserResponse, error) {
	user, err := s.database.CreateUser(ctx, in.ToUserInput())
	if err != nil {
		return nil, err
	}

	res := &todoproto.CreateUserResponse{
		Id:                    user.ID,
		Username:              user.Username,
		TwoFactorSecret:       user.TwoFactorSecret,
		IsAdmin:               user.IsAdmin,
		PasswordLastChangedOn: *user.PasswordLastChangedOn,
		CreatedOn:             user.CreatedOn,
		UpdatedOn:             *user.UpdatedOn,
		ArchivedOn:            *user.ArchivedOn,
	}

	return res, nil
}

// DeleteUser implements our gRPC server's interface
func (s *Service) DeleteUser(ctx context.Context, in *todoproto.DeleteUserRequest) (*todoproto.ErrorResponse, error) {
	err := s.database.DeleteUser(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

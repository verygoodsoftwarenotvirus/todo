package todoproto

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// ProtoUserFromModel converts a models user into a gRPC User.
func ProtoUserFromModel(u *models.User) *User {
	return &User{
		Id:                    u.ID,
		Username:              u.Username,
		HashedPassword:        u.HashedPassword,
		TwoFactorSecret:       u.TwoFactorSecret,
		IsAdmin:               u.IsAdmin,
		PasswordLastChangedOn: *u.PasswordLastChangedOn,
		CreatedOn:             u.CreatedOn,
		UpdatedOn:             *u.UpdatedOn,
		ArchivedOn:            *u.ArchivedOn,
	}
}

// ProtoUsersFromModels converts a slice of models users into a gRPC User slice.
func ProtoUsersFromModels(in []models.User) (out []*User) {
	for _, i := range in {
		out = append(out, ProtoUserFromModel(&i))
	}
	return
}

// ToModelUser converts a gRPC User into a models user
func (u *User) ToModelUser() *models.User {
	return &models.User{
		ID:                    u.Id,
		Username:              u.Username,
		HashedPassword:        u.HashedPassword,
		TwoFactorSecret:       u.TwoFactorSecret,
		IsAdmin:               u.IsAdmin,
		PasswordLastChangedOn: &u.PasswordLastChangedOn,
		CreatedOn:             u.CreatedOn,
		UpdatedOn:             &u.UpdatedOn,
		ArchivedOn:            &u.ArchivedOn,
	}
}

// ToUserInput converts a gRPC CreateUserRequest into a models UserInput
func (r *CreateUserRequest) ToUserInput() *models.UserInput {
	return &models.UserInput{
		Username: r.Username,
		Password: r.Password,
		IsAdmin:  r.IsAdmin,
	}
}

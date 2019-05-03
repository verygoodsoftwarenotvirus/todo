package models

import (
	"context"
	"net/http"
)

// UserDataManager describes a structure which can manage users in permanent storage
type UserDataManager interface {
	GetUser(ctx context.Context, userID uint64) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserCount(ctx context.Context, filter *QueryFilter) (uint64, error)
	GetUsers(ctx context.Context, filter *QueryFilter) (*UserList, error)
	CreateUser(ctx context.Context, input *UserInput) (*User, error)
	UpdateUser(ctx context.Context, updated *User) error
	DeleteUser(ctx context.Context, userID uint64) error
}

// UserDataServer describes a structure capable of serving traffic related to users
type UserDataServer interface {
	UserLoginInputMiddleware(next http.Handler) http.Handler
	UserInputMiddleware(next http.Handler) http.Handler
	PasswordUpdateInputMiddleware(next http.Handler) http.Handler
	TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler

	List(res http.ResponseWriter, req *http.Request)
	Create(res http.ResponseWriter, req *http.Request)
	Read(res http.ResponseWriter, req *http.Request)
	NewTOTPSecret(res http.ResponseWriter, req *http.Request)
	UpdatePassword(res http.ResponseWriter, req *http.Request)
	Delete(res http.ResponseWriter, req *http.Request)
}

const (
	// UserKey is the non-string type we use for referencing a user in a context
	UserKey ContextKey = "user"
	// UserIDKey is the non-string type we use for referencing a user ID in a context
	UserIDKey ContextKey = "user_id"
)

// UserLoginInput represents the payload used to log in a user
type UserLoginInput struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	TOTPToken string `json:"totp_token"`
}

// UserInput represents the input required to modify/create users
type UserInput struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	IsAdmin         bool   `json:"is_admin"`
	TwoFactorSecret string `json:"-"`
}

// UserCreationResponse is a response structure for Users that doesn't contain password fields
type UserCreationResponse struct {
	ID                    uint64  `json:"id"`
	Username              string  `json:"username"`
	TwoFactorSecret       string  `json:"two_factor_secret"`
	IsAdmin               bool    `json:"is_admin"`
	PasswordLastChangedOn *uint64 `json:"password_last_changed_on"`
	CreatedOn             uint64  `json:"created_on"`
	UpdatedOn             *uint64 `json:"updated_on"`
	ArchivedOn            *uint64 `json:"archived_on"`
}

// User represents a user
type User struct {
	ID                    uint64  `json:"id"`
	Username              string  `json:"username"`
	HashedPassword        string  `json:"-"`
	Salt                  []byte  `json:"-"`
	TwoFactorSecret       string  `json:"-"`
	IsAdmin               bool    `json:"is_admin"`
	PasswordLastChangedOn *uint64 `json:"password_last_changed_on"`
	CreatedOn             uint64  `json:"created_on"`
	UpdatedOn             *uint64 `json:"updated_on"`
	ArchivedOn            *uint64 `json:"archived_on"`
}

// Update accepts a User as input and merges those values if they're set
func (u *User) Update(input User) {
	if input.Username != "" && input.Username != u.Username {
		u.Username = input.Username
	}

	if input.HashedPassword != "" && input.HashedPassword != u.HashedPassword {
		u.HashedPassword = input.HashedPassword
	}

	if input.TwoFactorSecret != "" && input.TwoFactorSecret != u.TwoFactorSecret {
		u.TwoFactorSecret = input.TwoFactorSecret
	}
}

// UserList represents a list of users
type UserList struct {
	Pagination
	Users []User `json:"users"`
}

// PasswordUpdateInput represents input a user would provide when updating their password
type PasswordUpdateInput struct {
	NewPassword string `json:"new_password"`
	TOTPSecretRefreshInput
}

// TOTPSecretRefreshInput represents input a user would provide when updating their 2FA secret
type TOTPSecretRefreshInput struct {
	CurrentPassword string `json:"current_password"`
	TOTPToken       string `json:"totp_token"`
}

// TOTPSecretRefreshResponse represents the response we provide to a user when updating their 2FA secret
type TOTPSecretRefreshResponse struct {
	TwoFactorSecret string `json:"two_factor_secret"`
}

package models

import (
	"context"
	"net/http"
)

const (
	// UserKey is the non-string type we use for referencing a user in a context
	UserKey ContextKey = "user"
	// UserIDKey is the non-string type we use for referencing a user ID in a context
	UserIDKey ContextKey = "user_id"
	// UserIsAdminKey is the non-string type we use for referencing a user's admin status in a context
	UserIsAdminKey ContextKey = "is_admin"
)

type (
	// User represents a user
	User struct {
		ID                    uint64  `json:"id"`
		Username              string  `json:"username"`
		HashedPassword        string  `json:"-"`
		Salt                  []byte  `json:"-"`
		TwoFactorSecret       string  `json:"-"`
		PasswordLastChangedOn *uint64 `json:"password_last_changed_on"`
		IsAdmin               bool    `json:"is_admin"`
		CreatedOn             uint64  `json:"created_on"`
		UpdatedOn             *uint64 `json:"updated_on"`
		ArchivedOn            *uint64 `json:"archived_on"`
	}

	// UserList represents a list of users
	UserList struct {
		Pagination
		Users []User `json:"users"`
	}

	// UserLoginInput represents the payload used to log in a user
	UserLoginInput struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		TOTPToken string `json:"totp_token"`
	}

	// UserInput represents the input required to modify/create users
	UserInput struct {
		Username        string `json:"username"`
		Password        string `json:"password"`
		TwoFactorSecret string `json:"-"`
	}

	// UserCreationResponse is a response structure for Users that doesn't contain password fields, but does contain the two factor secret
	UserCreationResponse struct {
		ID                    uint64  `json:"id"`
		Username              string  `json:"username"`
		TwoFactorSecret       string  `json:"two_factor_secret"`
		PasswordLastChangedOn *uint64 `json:"password_last_changed_on"`
		IsAdmin               bool    `json:"is_admin"`
		CreatedOn             uint64  `json:"created_on"`
		UpdatedOn             *uint64 `json:"updated_on"`
		ArchivedOn            *uint64 `json:"archived_on"`
		TwoFactorQRCode       string  `json:"qr_code"`
	}

	// PasswordUpdateInput represents input a user would provide when updating their password
	PasswordUpdateInput struct {
		NewPassword     string `json:"new_password"`
		CurrentPassword string `json:"current_password"`
		TOTPToken       string `json:"totp_token"`
	}

	// TOTPSecretRefreshInput represents input a user would provide when updating their 2FA secret
	TOTPSecretRefreshInput struct {
		CurrentPassword string `json:"current_password"`
		TOTPToken       string `json:"totp_token"`
	}

	// TOTPSecretRefreshResponse represents the response we provide to a user when updating their 2FA secret
	TOTPSecretRefreshResponse struct {
		TwoFactorSecret string `json:"two_factor_secret"`
	}

	// UserDataManager describes a structure which can manage users in permanent storage
	UserDataManager interface {
		GetUser(ctx context.Context, userID uint64) (*User, error)
		GetUserByUsername(ctx context.Context, username string) (*User, error)
		GetUserCount(ctx context.Context, filter *QueryFilter) (uint64, error)
		GetUsers(ctx context.Context, filter *QueryFilter) (*UserList, error)
		CreateUser(ctx context.Context, input *UserInput) (*User, error)
		UpdateUser(ctx context.Context, updated *User) error
		ArchiveUser(ctx context.Context, userID uint64) error
	}

	// UserDataServer describes a structure capable of serving traffic related to users
	UserDataServer interface {
		UserInputMiddleware(next http.Handler) http.Handler
		PasswordUpdateInputMiddleware(next http.Handler) http.Handler
		TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler

		ListHandler() http.HandlerFunc
		CreateHandler() http.HandlerFunc
		ReadHandler() http.HandlerFunc
		NewTOTPSecretHandler() http.HandlerFunc
		UpdatePasswordHandler() http.HandlerFunc
		ArchiveHandler() http.HandlerFunc
	}
)

// Update accepts a User as input and merges those values if they're set
func (u *User) Update(input *User) {
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

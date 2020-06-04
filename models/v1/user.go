package models

import (
	"context"
	"net/http"
)

type (
	// User represents a user.
	User struct {
		Salt                      []byte  `json:"-"`
		Username                  string  `json:"username"`
		HashedPassword            string  `json:"-"`
		TwoFactorSecret           string  `json:"-"`
		ID                        uint64  `json:"id"`
		PasswordLastChangedOn     *uint64 `json:"passwordLastChangedOn"`
		TwoFactorSecretVerifiedOn *uint64 `json:"-"`
		CreatedOn                 uint64  `json:"createdOn"`
		UpdatedOn                 *uint64 `json:"updatedOn"`
		ArchivedOn                *uint64 `json:"archivedOn"`
		IsAdmin                   bool    `json:"isAdmin"`
		RequiresPasswordChange    bool    `json:"requiresPasswordChange"`
	}

	// UserList represents a list of users.
	UserList struct {
		Pagination
		Users []User `json:"users"`
	}

	// UserLoginInput represents the payload used to log in a user.
	UserLoginInput struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		TOTPToken string `json:"totpToken"`
	}

	// UserCreationInput represents the input required from users to register an account.
	UserCreationInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// UserDatabaseCreationInput is used by the user creation route to communicate with the database.
	UserDatabaseCreationInput struct {
		Username        string `json:"-"`
		HashedPassword  string `json:"-"`
		TwoFactorSecret string `json:"-"`
	}

	// UserCreationResponse is a response structure for Users that doesn't contain password fields, but does contain the two factor secret.
	UserCreationResponse struct {
		ID                    uint64  `json:"id"`
		Username              string  `json:"username"`
		TwoFactorSecret       string  `json:"twoFactorSecret"`
		PasswordLastChangedOn *uint64 `json:"passwordLastChangedOn"`
		IsAdmin               bool    `json:"isAdmin"`
		CreatedOn             uint64  `json:"createdOn"`
		UpdatedOn             *uint64 `json:"updatedOn"`
		ArchivedOn            *uint64 `json:"archivedOn"`
		TwoFactorQRCode       string  `json:"qrCode"`
	}

	// PasswordUpdateInput represents input a user would provide when updating their password.
	PasswordUpdateInput struct {
		NewPassword     string `json:"newPassword"`
		CurrentPassword string `json:"currentPassword"`
		TOTPToken       string `json:"totpToken"`
	}

	// TOTPSecretRefreshInput represents input a user would provide when updating their 2FA secret.
	TOTPSecretRefreshInput struct {
		CurrentPassword string `json:"currentPassword"`
		TOTPToken       string `json:"totpToken"`
	}

	// TOTPSecretVerificationInput represents input a user would provide when validating their 2FA secret.
	TOTPSecretVerificationInput struct {
		UserID    uint64 `json:"userID"`
		TOTPToken string `json:"totpToken"`
	}

	// TOTPSecretRefreshResponse represents the response we provide to a user when updating their 2FA secret.
	TOTPSecretRefreshResponse struct {
		TwoFactorSecret string `json:"twoFactorSecret"`
	}

	// UserDataManager describes a structure which can manage users in permanent storage.
	UserDataManager interface {
		GetUser(ctx context.Context, userID uint64) (*User, error)
		GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*User, error)
		VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error
		GetUserByUsername(ctx context.Context, username string) (*User, error)
		GetAllUsersCount(ctx context.Context) (uint64, error)
		GetUsers(ctx context.Context, filter *QueryFilter) (*UserList, error)
		CreateUser(ctx context.Context, input UserDatabaseCreationInput) (*User, error)
		UpdateUser(ctx context.Context, updated *User) error
		UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error
		ArchiveUser(ctx context.Context, userID uint64) error
	}

	// UserDataServer describes a structure capable of serving traffic related to users.
	UserDataServer interface {
		UserInputMiddleware(next http.Handler) http.Handler
		PasswordUpdateInputMiddleware(next http.Handler) http.Handler
		TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler
		TOTPSecretVerificationInputMiddleware(next http.Handler) http.Handler

		ListHandler() http.HandlerFunc
		CreateHandler() http.HandlerFunc
		ReadHandler() http.HandlerFunc
		NewTOTPSecretHandler() http.HandlerFunc
		TOTPSecretVerificationHandler() http.HandlerFunc
		UpdatePasswordHandler() http.HandlerFunc
		ArchiveHandler() http.HandlerFunc
	}
)

// Update accepts a User as input and merges those values if they're set.
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

// ToSessionInfo accepts a User as input and merges those values if they're set.
func (u *User) ToSessionInfo() *SessionInfo {
	return &SessionInfo{
		UserID:      u.ID,
		UserIsAdmin: u.IsAdmin,
	}
}

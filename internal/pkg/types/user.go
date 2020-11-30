package types

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search"
)

const (
	// UsersSearchIndexName is the name of the index used to search through items.
	UsersSearchIndexName search.IndexName = "users"

	// GoodStandingAccountStatus indicates a User's account is in good standing.
	GoodStandingAccountStatus userAccountStatus = "good"
	// UnverifiedAccountStatus indicates a User's account requires two factor secret verification.
	UnverifiedAccountStatus userAccountStatus = "unverified"
	// BannedAccountStatus indicates a User's account is banned.
	BannedAccountStatus userAccountStatus = "banned"
	// TerminatedAccountStatus indicates a User's account is banned.
	TerminatedAccountStatus userAccountStatus = "terminated"
)

// IsValidAccountStatus returns whether or not the provided string is a valid userAccountStatus.
func IsValidAccountStatus(s string) bool {
	switch s {
	case string(GoodStandingAccountStatus),
		string(UnverifiedAccountStatus),
		string(BannedAccountStatus),
		string(TerminatedAccountStatus):
		return true
	default:
		return false
	}
}

type (
	userAccountStatus string

	// User represents a User.
	User struct {
		Salt                      []byte                          `json:"-"`
		Username                  string                          `json:"username"`
		HashedPassword            string                          `json:"-"`
		TwoFactorSecret           string                          `json:"-"`
		AccountStatus             userAccountStatus               `json:"accountStatus"`
		AccountStatusExplanation  string                          `json:"accountStatusExplanation"`
		ID                        uint64                          `json:"id"`
		PasswordLastChangedOn     *uint64                         `json:"passwordLastChangedOn"`
		TwoFactorSecretVerifiedOn *uint64                         `json:"-"`
		CreatedOn                 uint64                          `json:"createdOn"`
		LastUpdatedOn             *uint64                         `json:"lastUpdatedOn"`
		ArchivedOn                *uint64                         `json:"archivedOn"`
		AdminPermissions          bitmask.AdminPermissionsBitmask `json:"adminPermissions"`
		IsAdmin                   bool                            `json:"isAdmin"`
		RequiresPasswordChange    bool                            `json:"requiresPasswordChange"`
	}

	// UserList represents a list of users.
	UserList struct {
		Pagination
		Users []User `json:"users"`
	}

	// UserCreationInput represents the input required from users to register an account.
	UserCreationInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// UserDataStoreCreationInput is used by the User creation route to communicate with the data store.
	UserDataStoreCreationInput struct {
		Salt            []byte `json:"-"`
		Username        string `json:"-"`
		HashedPassword  string `json:"-"`
		TwoFactorSecret string `json:"-"`
	}

	// UserCreationResponse is a response structure for Users that doesn't contain password fields, but does contain the two factor secret.
	UserCreationResponse struct {
		ID              uint64            `json:"id"`
		Username        string            `json:"username"`
		IsAdmin         bool              `json:"isAdmin"`
		CreatedOn       uint64            `json:"createdOn"`
		AccountStatus   userAccountStatus `json:"accountStatus"`
		TwoFactorQRCode string            `json:"qrCode"`
	}

	// UserLoginInput represents the payload used to log in a User.
	UserLoginInput struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		TOTPToken string `json:"totpToken"`
	}

	// PasswordUpdateInput represents input a User would provide when updating their password.
	PasswordUpdateInput struct {
		NewPassword     string `json:"newPassword"`
		CurrentPassword string `json:"currentPassword"`
		TOTPToken       string `json:"totpToken"`
	}

	// TOTPSecretRefreshInput represents input a User would provide when updating their 2FA secret.
	TOTPSecretRefreshInput struct {
		CurrentPassword string `json:"currentPassword"`
		TOTPToken       string `json:"totpToken"`
	}

	// TOTPSecretVerificationInput represents input a User would provide when validating their 2FA secret.
	TOTPSecretVerificationInput struct {
		UserID    uint64 `json:"userID"`
		TOTPToken string `json:"totpToken"`
	}

	// TOTPSecretRefreshResponse represents the response we provide to a User when updating their 2FA secret.
	TOTPSecretRefreshResponse struct {
		TwoFactorQRCode string `json:"qrCode"`
		TwoFactorSecret string `json:"twoFactorSecret"`
	}

	// AdminUserDataManager contains administrative User functions that we don't necessarily want to expose
	// to, say, the collection of handlers.
	AdminUserDataManager interface {
		UpdateUserAccountStatus(ctx context.Context, userID uint64, input AccountStatusUpdateInput) error
	}

	// UserDataManager describes a structure which can manage users in permanent storage.
	UserDataManager interface {
		// GetUser retrieves a User from the data store via their identifier.
		GetUser(ctx context.Context, userID uint64) (*User, error)
		// GetUserWithUnverifiedTwoFactorSecret retrieves a User from the data store via their identifier, with the strict
		// caveat that the User associated with that row must also have an unverified two factor secret. This is used
		// for the two factor secret verification route, as all other User retrieval functions restrict to verified
		// two factor secrets.
		GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*User, error)
		// VerifyUserTwoFactorSecret marks a User with a given identifier as having a verified two factor secret.
		VerifyUserTwoFactorSecret(ctx context.Context, userID uint64) error
		// GetUserByUsername retrieves a User via their username.
		GetUserByUsername(ctx context.Context, username string) (*User, error)
		// SearchForUsersByUsername is intended to be a SUPPORT ONLY function, used within an interface to find a
		// User quickly while only typing the first few letters of their username. No search index is utilized.
		SearchForUsersByUsername(ctx context.Context, usernameQuery string) ([]User, error)
		// GetAllUsersCount fetches the current User count.
		GetAllUsersCount(ctx context.Context) (uint64, error)
		// GetUsers is intended to be a SUPPORT ONLY function, and fetches a page of users adhering to a given filter.
		GetUsers(ctx context.Context, filter *QueryFilter) (*UserList, error)
		// CreateUser creates a new User in the data store.
		CreateUser(ctx context.Context, input UserDataStoreCreationInput) (*User, error)
		// UpdateUser updates a User in the data store.
		UpdateUser(ctx context.Context, updated *User) error
		// UpdateUserPassword  updates a given User's password exclusively in the data store.
		UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error
		// ArchiveUser marks a User as archived in the data store.
		ArchiveUser(ctx context.Context, userID uint64) error
	}

	// UserAuditManager describes a structure capable of logging audit events related to users.
	UserAuditManager interface {
		GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]AuditLogEntry, error)
		LogUserCreationEvent(ctx context.Context, user *User)
		LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64)
		LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64)
		LogUserUpdatePasswordEvent(ctx context.Context, userID uint64)
		LogUserArchiveEvent(ctx context.Context, userID uint64)
	}

	// UserDataServer describes a structure capable of serving traffic related to users.
	UserDataServer interface {
		UserInputMiddleware(next http.Handler) http.Handler
		PasswordUpdateInputMiddleware(next http.Handler) http.Handler
		TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler
		TOTPSecretVerificationInputMiddleware(next http.Handler) http.Handler

		ListHandler(res http.ResponseWriter, req *http.Request)
		AuditEntryHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		SelfHandler(res http.ResponseWriter, req *http.Request)
		UsernameSearchHandler(res http.ResponseWriter, req *http.Request)
		NewTOTPSecretHandler(res http.ResponseWriter, req *http.Request)
		TOTPSecretVerificationHandler(res http.ResponseWriter, req *http.Request)
		UpdatePasswordHandler(res http.ResponseWriter, req *http.Request)
		ArchiveHandler(res http.ResponseWriter, req *http.Request)
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

// ToSessionInfo produces a SessionInfo object from a User's data.
func (u *User) ToSessionInfo() *SessionInfo {
	return &SessionInfo{
		User:              u,
		UserID:            u.ID,
		UserIsAdmin:       u.IsAdmin,
		UserAccountStatus: u.AccountStatus,
		AdminPermissions:  u.AdminPermissions,
	}
}

// ToStatusResponse produces a UserStatusResponse object from a User's data.
func (u *User) ToStatusResponse() *UserStatusResponse {
	return &UserStatusResponse{
		UserIsAdmin:              u.IsAdmin,
		UserAccountStatus:        u.AccountStatus,
		AccountStatusExplanation: u.AccountStatusExplanation,
		AdminPermissions:         u.AdminPermissions.Summary(),
	}
}

// IsBanned is a handy helper function.
func (u *User) IsBanned() bool {
	return u.AccountStatus == BannedAccountStatus
}

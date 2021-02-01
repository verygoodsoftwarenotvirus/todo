package types

import (
	"context"
	"math"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions/bitmask"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// GoodStandingAccountStatus indicates a User's account is in good standing.
	GoodStandingAccountStatus userReputation = "good"
	// UnverifiedAccountStatus indicates a User's account requires two factor secret verification.
	UnverifiedAccountStatus userReputation = "unverified"
	// BannedAccountStatus indicates a User's account is banned.
	BannedAccountStatus userReputation = "banned"
	// TerminatedAccountStatus indicates a User's account is banned.
	TerminatedAccountStatus userReputation = "terminated"

	validTOTPTokenLength = 6
)

var (
	totpTokenLengthRule = validation.Length(validTOTPTokenLength, validTOTPTokenLength)
)

// IsValidAccountStatus returns whether or not the provided string is a valid userReputation.
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
	userReputation string

	// User represents a User.
	User struct {
		Salt                      []byte                       `json:"-"`
		Username                  string                       `json:"username"`
		HashedPassword            string                       `json:"-"`
		TwoFactorSecret           string                       `json:"-"`
		AccountStatus             userReputation               `json:"accountStatus"`
		AccountStatusExplanation  string                       `json:"accountStatusExplanation"`
		ID                        uint64                       `json:"id"`
		ExternalID                string                       `json:"externalID"`
		PasswordLastChangedOn     *uint64                      `json:"passwordLastChangedOn"`
		TwoFactorSecretVerifiedOn *uint64                      `json:"-"`
		CreatedOn                 uint64                       `json:"createdOn"`
		LastUpdatedOn             *uint64                      `json:"lastUpdatedOn"`
		ArchivedOn                *uint64                      `json:"archivedOn"`
		SiteAdminPermissions      bitmask.SiteAdminPermissions `json:"siteAdminPermissions"`
		IsSiteAdmin               bool                         `json:"isSiteAdmin"`
		RequiresPasswordChange    bool                         `json:"requiresPasswordChange"`
		AvatarSrc                 *string                      `json:"avatar"`
	}

	// TestUserCreationConfig is a helper struct because of cyclical imports.
	TestUserCreationConfig struct {
		// Username defines our test user's username we create in the event we create them.
		Username string `json:"username" mapstructure:"username" toml:"username,omitempty"`
		// Password defines our test user's password we create in the event we create them.
		Password string `json:"password" mapstructure:"password" toml:"password,omitempty"`
		// HashedPassword is the hashed form of the above password.
		HashedPassword string `json:"hashed_password" mapstructure:"hashed_password" toml:"hashed_password,omitempty"`
		// IsSiteAdmin defines our test user's admin status we create in the event we create them.
		IsSiteAdmin bool `json:"is_site_admin" mapstructure:"is_site_admin" toml:"is_site_admin,omitempty"`
	}

	// UserList represents a list of users.
	UserList struct {
		Pagination
		Users []*User `json:"users"`
	}

	// NewUserCreationInput represents the input required from users to register an account.
	NewUserCreationInput struct {
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
		ID              uint64         `json:"id"`
		Username        string         `json:"username"`
		IsAdmin         bool           `json:"isAdmin"`
		CreatedOn       uint64         `json:"createdOn"`
		AccountStatus   userReputation `json:"accountStatus"`
		TwoFactorQRCode string         `json:"qrCode"`
	}

	// PASETOCreationInput represents the payload used to create a PASETO token.
	PASETOCreationInput struct {
		ClientID     string `json:"clientID"`
		ClientSecret string `json:"clientSecret"`
		TOTPToken    string `json:"totpToken"`
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

	// UserSQLQueryBuilder describes a structure capable of generating query/arg pairs for certain situations.
	UserSQLQueryBuilder interface {
		BuildGetUserQuery(userID uint64) (query string, args []interface{})
		BuildGetUsersQuery(filter *QueryFilter) (query string, args []interface{})
		BuildGetUserWithUnverifiedTwoFactorSecretQuery(userID uint64) (query string, args []interface{})
		BuildGetUserByUsernameQuery(username string) (query string, args []interface{})
		BuildSearchForUserByUsernameQuery(usernameQuery string) (query string, args []interface{})
		BuildGetAllUsersCountQuery() (query string)
		BuildCreateUserQuery(input UserDataStoreCreationInput) (query string, args []interface{})
		BuildUpdateUserQuery(input *User) (query string, args []interface{})
		BuildUpdateUserPasswordQuery(userID uint64, newHash string) (query string, args []interface{})
		BuildVerifyUserTwoFactorSecretQuery(userID uint64) (query string, args []interface{})
		BuildArchiveUserQuery(userID uint64) (query string, args []interface{})
		BuildGetAuditLogEntriesForUserQuery(userID uint64) (query string, args []interface{})
		BuildSetUserStatusQuery(userID uint64, input UserReputationUpdateInput) (query string, args []interface{})
	}

	// AdminUserDataManager contains administrative User functions that we don't necessarily want to expose
	// to, say, the collection of handlers.
	AdminUserDataManager interface {
		UpdateUserAccountStatus(ctx context.Context, userID uint64, input UserReputationUpdateInput) error
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
		SearchForUsersByUsername(ctx context.Context, usernameQuery string) ([]*User, error)
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
		GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]*AuditLogEntry, error)
		LogUserCreationEvent(ctx context.Context, user *User)
		LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64)
		LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64)
		LogUserUpdatePasswordEvent(ctx context.Context, userID uint64)
		LogUserArchiveEvent(ctx context.Context, userID uint64)
	}

	// UserDataService describes a structure capable of serving traffic related to users.
	UserDataService interface {
		UserCreationInputMiddleware(next http.Handler) http.Handler
		PasswordUpdateInputMiddleware(next http.Handler) http.Handler
		TOTPSecretRefreshInputMiddleware(next http.Handler) http.Handler
		TOTPSecretVerificationInputMiddleware(next http.Handler) http.Handler
		AvatarUploadMiddleware(next http.Handler) http.Handler

		ListHandler(res http.ResponseWriter, req *http.Request)
		AuditEntryHandler(res http.ResponseWriter, req *http.Request)
		CreateHandler(res http.ResponseWriter, req *http.Request)
		ReadHandler(res http.ResponseWriter, req *http.Request)
		SelfHandler(res http.ResponseWriter, req *http.Request)
		UsernameSearchHandler(res http.ResponseWriter, req *http.Request)
		NewTOTPSecretHandler(res http.ResponseWriter, req *http.Request)
		TOTPSecretVerificationHandler(res http.ResponseWriter, req *http.Request)
		UpdatePasswordHandler(res http.ResponseWriter, req *http.Request)
		AvatarUploadHandler(res http.ResponseWriter, req *http.Request)
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
		Username:          u.Username,
		UserID:            u.ID,
		UserIsSiteAdmin:   u.IsSiteAdmin,
		UserAccountStatus: u.AccountStatus,
		AdminPermissions:  u.SiteAdminPermissions,
	}
}

// ToStatusResponse produces a UserStatusResponse object from a User's data.
func (u *User) ToStatusResponse() *UserStatusResponse {
	return &UserStatusResponse{
		UserIsAdmin:              u.IsSiteAdmin,
		UserAccountStatus:        u.AccountStatus,
		AccountStatusExplanation: u.AccountStatusExplanation,
		AdminPermissions:         u.SiteAdminPermissions.SiteAdminPermissionsSummary(),
	}
}

// IsBanned is a handy helper function.
func (u *User) IsBanned() bool {
	return u.AccountStatus == BannedAccountStatus
}

// Validate ensures our provided NewUserCreationInput meets expectations.
func (i *NewUserCreationInput) Validate(ctx context.Context, minUsernameLength, minPasswordLength uint8) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.Username, validation.Required, validation.Length(int(minUsernameLength), math.MaxInt8)),
		validation.Field(&i.Password, validation.Required, validation.Length(int(minPasswordLength), math.MaxInt8)),
	)
}

// Validate ensures our  provided UserLoginInput meets expectations.
func (i *UserLoginInput) Validate(ctx context.Context, minUsernameLength, minPasswordLength uint8) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.Username, validation.Required, validation.Length(int(minUsernameLength), math.MaxInt8)),
		validation.Field(&i.Password, validation.Required, validation.Length(int(minPasswordLength), math.MaxInt8)),
		validation.Field(&i.TOTPToken, validation.Required, totpTokenLengthRule),
	)
}

// Validate ensures our  provided UserLoginInput meets expectations.
func (i *PASETOCreationInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.ClientID, validation.Required),
		validation.Field(&i.ClientSecret, validation.Required),
		validation.Field(&i.TOTPToken, validation.Required, totpTokenLengthRule),
	)
}

// Validate ensures our provided PasswordUpdateInput meets expectations.
func (i *PasswordUpdateInput) Validate(ctx context.Context, minPasswordLength uint8) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.CurrentPassword, validation.Required, validation.Length(int(minPasswordLength), math.MaxInt8)),
		validation.Field(&i.NewPassword, validation.Required, validation.Length(int(minPasswordLength), math.MaxInt8)),
		validation.Field(&i.TOTPToken, validation.Required, totpTokenLengthRule),
	)
}

// Validate ensures our provided TOTPSecretRefreshInput meets expectations.
func (i *TOTPSecretRefreshInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.CurrentPassword, validation.Required),
		validation.Field(&i.TOTPToken, validation.Required, totpTokenLengthRule),
	)
}

// Validate ensures our provided TOTPSecretVerificationInput meets expectations.
func (i *TOTPSecretVerificationInput) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.UserID, validation.Required),
		validation.Field(&i.TOTPToken, validation.Required, totpTokenLengthRule),
	)
}

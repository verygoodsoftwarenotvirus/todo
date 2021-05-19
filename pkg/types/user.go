package types

import (
	"context"
	"math"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// GoodStandingAccountStatus indicates a User's account is in good standing.
	GoodStandingAccountStatus accountStatus = "good"
	// UnverifiedAccountStatus indicates a User's account requires two factor secret verification.
	UnverifiedAccountStatus accountStatus = "unverified"
	// BannedUserAccountStatus indicates a User's account is banned.
	BannedUserAccountStatus accountStatus = "banned"
	// TerminatedUserReputation indicates a User's account is banned.
	TerminatedUserReputation accountStatus = "terminated"

	validTOTPTokenLength = 6
)

var (
	totpTokenLengthRule = validation.Length(validTOTPTokenLength, validTOTPTokenLength)
)

type (
	accountStatus string

	// User represents a User.
	User struct {
		PasswordLastChangedOn     *uint64       `json:"passwordLastChangedOn"`
		ArchivedOn                *uint64       `json:"archivedOn"`
		LastUpdatedOn             *uint64       `json:"lastUpdatedOn"`
		TwoFactorSecretVerifiedOn *uint64       `json:"-"`
		AvatarSrc                 *string       `json:"avatar"`
		ExternalID                string        `json:"externalID"`
		Username                  string        `json:"username"`
		ReputationExplanation     string        `json:"reputationExplanation"`
		ServiceAccountStatus      accountStatus `json:"reputation"`
		TwoFactorSecret           string        `json:"-"`
		HashedPassword            string        `json:"-"`
		ServiceRoles              []string      `json:"serviceRole"`
		ID                        uint64        `json:"id"`
		CreatedOn                 uint64        `json:"createdOn"`
		RequiresPasswordChange    bool          `json:"requiresPasswordChange"`
	}

	// TestUserCreationConfig is here because of cyclical imports.
	TestUserCreationConfig struct {
		Username       string `json:"username" mapstructure:"username" toml:"username,omitempty"`
		Password       string `json:"password" mapstructure:"password" toml:"password,omitempty"`
		HashedPassword string `json:"hashed_password" mapstructure:"hashed_password" toml:"hashed_password,omitempty"`
		IsServiceAdmin bool   `json:"is_site_admin" mapstructure:"is_site_admin" toml:"is_site_admin,omitempty"`
	}

	// UserList represents a list of users.
	UserList struct {
		Users []*User `json:"users"`
		Pagination
	}

	// UserRegistrationInput represents the input required from users to register an account.
	UserRegistrationInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// UserDataStoreCreationInput is used by the User creation route to communicate with the data store.
	UserDataStoreCreationInput struct {
		Username        string `json:"-"`
		HashedPassword  string `json:"-"`
		TwoFactorSecret string `json:"-"`
	}

	// UserCreationResponse is a response structure for Users that doesn't contain passwords fields, but does contain the two factor secret.
	UserCreationResponse struct {
		Username        string        `json:"username"`
		AccountStatus   accountStatus `json:"accountStatus"`
		TwoFactorSecret string        `json:"twoFactorSecret"`
		TwoFactorQRCode string        `json:"qrCode"`
		CreatedUserID   uint64        `json:"ID"`
		CreatedOn       uint64        `json:"createdOn"`
		IsAdmin         bool          `json:"isAdmin"`
	}

	// UserLoginInput represents the payload used to log in a User.
	UserLoginInput struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		TOTPToken string `json:"totpToken"`
	}

	// PasswordUpdateInput represents input a User would provide when updating their passwords.
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
		TOTPToken string `json:"totpToken"`
		UserID    uint64 `json:"userID"`
	}

	// TOTPSecretRefreshResponse represents the response we provide to a User when updating their 2FA secret.
	TOTPSecretRefreshResponse struct {
		TwoFactorQRCode string `json:"qrCode"`
		TwoFactorSecret string `json:"twoFactorSecret"`
	}

	// AdminUserDataManager contains administrative User functions that we don't necessarily want to expose
	// to, say, the collection of handlers.
	AdminUserDataManager interface {
		UpdateUserReputation(ctx context.Context, userID uint64, input *UserReputationUpdateInput) error
	}

	// UserDataManager describes a structure which can manage users in permanent storage.
	UserDataManager interface {
		UserHasStatus(ctx context.Context, userID uint64, statuses ...string) (bool, error)
		GetUser(ctx context.Context, userID uint64) (*User, error)
		GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID uint64) (*User, error)
		MarkUserTwoFactorSecretAsVerified(ctx context.Context, userID uint64) error
		GetUserByUsername(ctx context.Context, username string) (*User, error)
		SearchForUsersByUsername(ctx context.Context, usernameQuery string) ([]*User, error)
		GetAllUsersCount(ctx context.Context) (uint64, error)
		GetUsers(ctx context.Context, filter *QueryFilter) (*UserList, error)
		CreateUser(ctx context.Context, input *UserDataStoreCreationInput) (*User, error)
		UpdateUser(ctx context.Context, updated *User, changes []*FieldChangeSummary) error
		UpdateUserPassword(ctx context.Context, userID uint64, newHash string) error
		ArchiveUser(ctx context.Context, userID uint64) error
		GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]*AuditLogEntry, error)
	}

	// UserDataService describes a structure capable of serving traffic related to users.
	UserDataService interface {
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

		RegisterUser(ctx context.Context, registrationInput *UserRegistrationInput) (*UserCreationResponse, error)
		VerifyUserTwoFactorSecret(ctx context.Context, input *TOTPSecretVerificationInput) error
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

// IsValidAccountStatus returns whether the provided string is a valid accountStatus.
func IsValidAccountStatus(s string) bool {
	switch s {
	case string(GoodStandingAccountStatus),
		string(UnverifiedAccountStatus),
		string(BannedUserAccountStatus),
		string(TerminatedUserReputation):
		return true
	default:
		return false
	}
}

// IsBanned is a handy helper function.
func (u *User) IsBanned() bool {
	return u.ServiceAccountStatus == BannedUserAccountStatus
}

// ValidateWithContext ensures our provided UserRegistrationInput meets expectations.
func (i *UserRegistrationInput) ValidateWithContext(ctx context.Context, minUsernameLength, minPasswordLength uint8) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.Username, validation.Required, validation.Length(int(minUsernameLength), math.MaxInt8)),
		validation.Field(&i.Password, validation.Required, validation.Length(int(minPasswordLength), math.MaxInt8)),
	)
}

// ValidateWithContext ensures our provided UserLoginInput meets expectations.
func (i *UserLoginInput) ValidateWithContext(ctx context.Context, minUsernameLength, minPasswordLength uint8) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.Username, validation.Required, validation.Length(int(minUsernameLength), math.MaxInt8)),
		validation.Field(&i.Password, validation.Required, validation.Length(int(minPasswordLength), math.MaxInt8)),
		validation.Field(&i.TOTPToken, validation.Required, totpTokenLengthRule),
	)
}

// ValidateWithContext ensures our provided PasswordUpdateInput meets expectations.
func (i *PasswordUpdateInput) ValidateWithContext(ctx context.Context, minPasswordLength uint8) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.CurrentPassword, validation.Required, validation.Length(int(minPasswordLength), math.MaxInt8)),
		validation.Field(&i.NewPassword, validation.Required, validation.Length(int(minPasswordLength), math.MaxInt8)),
		validation.Field(&i.TOTPToken, validation.Required, totpTokenLengthRule),
	)
}

var _ validation.ValidatableWithContext = (*TOTPSecretRefreshInput)(nil)

// ValidateWithContext ensures our provided TOTPSecretRefreshInput meets expectations.
func (i *TOTPSecretRefreshInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.CurrentPassword, validation.Required),
		validation.Field(&i.TOTPToken, validation.Required, totpTokenLengthRule),
	)
}

var _ validation.ValidatableWithContext = (*TOTPSecretVerificationInput)(nil)

// ValidateWithContext ensures our provided TOTPSecretVerificationInput meets expectations.
func (i *TOTPSecretVerificationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, i,
		validation.Field(&i.UserID, validation.Required),
		validation.Field(&i.TOTPToken, validation.Required, totpTokenLengthRule),
	)
}

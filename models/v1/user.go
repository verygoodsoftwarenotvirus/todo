package models

type UserHandler interface {
	GetUser(username string) (*User, error)
	GetUserCount(filter *QueryFilter) (uint64, error)
	GetUsers(filter *QueryFilter) (*UserList, error)
	CreateUser(input *UserInput, totpSecret string) (*User, error)
	UpdateUser(updated *User) error
	DeleteUser(username string) error
}

const (
	UserKey   ContextKey = "user"
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
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

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
	TwoFactorSecret       string  `json:"-"`
	IsAdmin               bool    `json:"is_admin"`
	PasswordLastChangedOn *uint64 `json:"password_last_changed_on"`
	CreatedOn             uint64  `json:"created_on"`
	UpdatedOn             *uint64 `json:"updated_on"`
	ArchivedOn            *uint64 `json:"archived_on"`
}

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

type UserList struct {
	Pagination
	Users []User `json:"users"`
}

type PasswordUpdateInput struct {
	NewPassword string `json:"new_password"`
	TOTPSecretRefreshInput
}

type TOTPSecretRefreshInput struct {
	CurrentPassword string `json:"current_password"`
	TOTPToken       string `json:"totp_token"`
}

package models

type UserHandler interface {
	GetUser(identifier string) (*User, error)
	GetUsers(filter *QueryFilter) ([]User, error)
	CreateUser(input *UserCreationInput) (*User, error)
	UpdateUser(updated *User) error
	DeleteUser(id uint) error
}

// UserLoginInput represents the payload used to log in a user
type UserLoginInput struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	TOTPToken string `json:"totp_token"`
}

// UserCreationInput represents a user
type UserCreationInput struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	TOTPSecret string `json:"totp_secret"`
}

// User represents a user
type User struct {
	ID                    uint64  `json:"id"`
	Username              string  `json:"username"`
	HashedPassword        string  `json:"-"`
	TwoFactorSecret       string  `json:"-"`
	PasswordLastChangedOn *uint64 `json:"password_last_changed_on"`
	CreatedOn             uint64  `json:"created_on"`
	UpdatedOn             *uint64 `json:"updated_on"`
	ArchivedOn            *uint64 `json:"archived_on"`
}

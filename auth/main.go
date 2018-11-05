package auth

type Enticator interface {
	HashPassword(password string) (string, error)
	PasswordIsAcceptable(password string) bool
	PasswordMatches(hashedPassword, providedPassword string) bool
	ValidateLogin(hashedPassword, providedPassword, twoFactorSecret, twoFactorCode string) (bool, error)
}

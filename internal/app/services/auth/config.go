package auth

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// DefaultCookieName is the default Cookie.Name.
	DefaultCookieName = "todocookie"
	// DefaultCookieDomain is the default Cookie.Domain.
	DefaultCookieDomain = "localhost"
	// DefaultCookieLifetime is the how long a cookie is valid.
	DefaultCookieLifetime = 24 * time.Hour
)

// Config represents our authentication configuration.
type Config struct {
	// CookieName indicates what the cookies' name will be.
	CookieName string `json:"cookie_name" mapstructure:"cookie_name" toml:"cookie_name,omitempty"`
	// CookieDomain indicates what domain the cookies will have set for them.
	CookieDomain string `json:"cookie_domain" mapstructure:"cookie_domain" toml:"cookie_domain,omitempty"`
	// CookieSigningKey indicates the secret the cookie builder should use.
	CookieSigningKey string `json:"cookie_signing_key" mapstructure:"cookie_signing_key" toml:"cookie_signing_key,omitempty"`
	// CookieLifetime indicates how long the cookies built should last.
	CookieLifetime time.Duration `json:"cookie_lifetime" mapstructure:"cookie_lifetime" toml:"cookie_lifetime,omitempty"`
	// Debug determines if debug logging or other development conditions are active.
	Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	// SecureCookiesOnly indicates if the cookies built should be marked as HTTPS only.
	SecureCookiesOnly bool `json:"secure_cookies_only" mapstructure:"secure_cookies_only" toml:"secure_cookies_only,omitempty"`
	// EnableUserSignup enables user signups.
	EnableUserSignup bool `json:"enable_user_signup" mapstructure:"enable_user_signup" toml:"enable_user_signup,omitempty"`
	// MinimumUsernameLength indicates how short a username can be.
	MinimumUsernameLength uint8 `json:"minimum_username_length" mapstructure:"minimum_username_length" toml:"minimum_username_length,omitempty"`
	// MinimumPasswordLength indicates how short a password can be.
	MinimumPasswordLength uint8 `json:"minimum_password_length" mapstructure:"minimum_password_length" toml:"minimum_password_length,omitempty"`
}

// Validate validates an AuthSettings struct.
func (cfg *Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.CookieName, validation.Required),
		validation.Field(&cfg.CookieDomain, validation.Required),
		validation.Field(&cfg.CookieLifetime, validation.Required),
		validation.Field(&cfg.MinimumUsernameLength, validation.Required),
		validation.Field(&cfg.MinimumPasswordLength, validation.Required),
	)
}

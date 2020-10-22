package config

import (
	"time"
)

const (
	// DefaultCookieLifetime is the default amount of time we authorize cookies for
	DefaultCookieLifetime = 24 * time.Hour
)

// AuthSettings represents our authentication configuration.
type AuthSettings struct {
	// CookieDomain indicates what domain the cookies will have set for them.
	CookieDomain string `json:"cookie_domain" mapstructure:"cookie_domain" toml:"cookie_domain,omitempty"`
	// CookieSecret indicates the secret the cookie builder should use.
	CookieSecret string `json:"cookie_secret" mapstructure:"cookie_secret" toml:"cookie_secret,omitempty"`
	// CookieLifetime indicates how long the cookies built should last.
	CookieLifetime time.Duration `json:"cookie_lifetime" mapstructure:"cookie_lifetime" toml:"cookie_lifetime,omitempty"`
	// Debug determines if debug logging or other development conditions are active.
	Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
	// SecureCookiesOnly indicates if the cookies built should be marked as HTTPS only.
	SecureCookiesOnly bool `json:"secure_cookies_only" mapstructure:"secure_cookies_only" toml:"secure_cookies_only,omitempty"`
	// EnableUserSignup enables user signups.
	EnableUserSignup bool `json:"enable_user_signup" mapstructure:"enable_user_signup" toml:"enable_user_signup,omitempty"`
}

package auth

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// DefaultCookieName is the default Cookie.BucketName.
	DefaultCookieName = "todocookie"
	// DefaultCookieDomain is the default Cookie.Domain.
	DefaultCookieDomain = "localhost"
	// DefaultCookieLifetime is the how long a cookie is valid.
	DefaultCookieLifetime = 24 * time.Hour

	pasetoKeyRequiredLength = 32
)

type (
	// CookieConfig holds our cookie settings.
	CookieConfig struct {
		// Name indicates what the cookies' name will be.
		Name string `json:"name" mapstructure:"name" toml:"name,omitempty"`
		// Domain indicates what domain the cookies will have set for them.
		Domain string `json:"domain" mapstructure:"domain" toml:"domain,omitempty"`
		// SigningKey indicates the secret the cookie builder should use.
		SigningKey string `json:"signing_key" mapstructure:"signing_key" toml:"signing_key,omitempty"`
		// Lifetime indicates how long the cookies built should last.
		Lifetime time.Duration `json:"lifetime" mapstructure:"lifetime" toml:"lifetime,omitempty"`
		// SecureOnly indicates if the cookies built should be marked as HTTPS only.
		SecureOnly bool `json:"secure_only" mapstructure:"secure_only" toml:"secure_only,omitempty"`
	}

	// PASETOConfig holds our PASETO token settings.
	PASETOConfig struct {
		// Issuer is the Issuer value that goes into our PASETO tokens.
		Issuer string `json:"issuer" mapstructure:"issuer" toml:"issuer,omitempty"`
		// Lifetime indicates how long the cookies built should last.
		Lifetime time.Duration `json:"lifetime" mapstructure:"lifetime" toml:"lifetime,omitempty"`
		// LocalModeKey is the key used to sign local PASETO tokens. Needs to be 32 bytes.
		LocalModeKey []byte `json:"local_mode_key" mapstructure:"local_mode_key" toml:"local_mode_key,omitempty"`
	}

	// Config represents our authentication configuration.
	Config struct {
		// Cookies configures our cookie settings.
		Cookies CookieConfig `json:"cookies" mapstructure:"cookies" toml:"cookies,omitempty"`
		// PASETO configures our PASETO token settings.
		PASETO PASETOConfig `json:"paseto" mapstructure:"paseto" toml:"paseto,omitempty"`
		// Debug determines if debug logging or other development conditions are active.
		Debug bool `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
		// EnableUserSignup enables user signups.
		EnableUserSignup bool `json:"enable_user_signup" mapstructure:"enable_user_signup" toml:"enable_user_signup,omitempty"`
		// MinimumUsernameLength indicates how short a username can be.
		MinimumUsernameLength uint8 `json:"minimum_username_length" mapstructure:"minimum_username_length" toml:"minimum_username_length,omitempty"`
		// MinimumPasswordLength indicates how short a authentication can be.
		MinimumPasswordLength uint8 `json:"minimum_password_length" mapstructure:"minimum_password_length" toml:"minimum_password_length,omitempty"`
	}
)

// Validate validates a CookieConfig struct.
func (cfg *CookieConfig) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Name, validation.Required),
		validation.Field(&cfg.Domain, validation.Required),
		validation.Field(&cfg.Lifetime, validation.Required),
	)
}

// Validate validates a PASETOConfig struct.
func (cfg *PASETOConfig) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Issuer, validation.Required),
		validation.Field(&cfg.LocalModeKey, validation.Required, validation.Length(pasetoKeyRequiredLength, pasetoKeyRequiredLength)),
	)
}

// Validate validates a Config struct.
func (cfg *Config) Validate(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Cookies, validation.Required),
		validation.Field(&cfg.PASETO, validation.Required),
		validation.Field(&cfg.MinimumUsernameLength, validation.Required),
		validation.Field(&cfg.MinimumPasswordLength, validation.Required),
	)
}

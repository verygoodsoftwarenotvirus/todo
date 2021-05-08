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

	staticError             = "error encountered, please try again later"
	pasetoKeyRequiredLength = 32
	pasetoDataKey           = "paseto_data"
	maxPASETOLifetime       = 10 * time.Minute
)

type (
	// CookieConfig holds our cookie settings.
	CookieConfig struct {
		Name       string        `json:"name" mapstructure:"name" toml:"name,omitempty"`
		Domain     string        `json:"domain" mapstructure:"domain" toml:"domain,omitempty"`
		HashKey    string        `json:"hash_key" mapstructure:"hash_key" toml:"hash_key,omitempty"`
		SigningKey string        `json:"signing_key" mapstructure:"signing_key" toml:"signing_key,omitempty"`
		Lifetime   time.Duration `json:"lifetime" mapstructure:"lifetime" toml:"lifetime,omitempty"`
		SecureOnly bool          `json:"secure_only" mapstructure:"secure_only" toml:"secure_only,omitempty"`
	}

	// PASETOConfig holds our PASETO settings.
	PASETOConfig struct {
		Issuer       string        `json:"issuer" mapstructure:"issuer" toml:"issuer,omitempty"`
		LocalModeKey []byte        `json:"local_mode_key" mapstructure:"local_mode_key" toml:"local_mode_key,omitempty"`
		Lifetime     time.Duration `json:"lifetime" mapstructure:"lifetime" toml:"lifetime,omitempty"`
	}

	// Config represents our passwords configuration.
	Config struct {
		PASETO                PASETOConfig `json:"paseto" mapstructure:"paseto" toml:"paseto,omitempty"`
		Cookies               CookieConfig `json:"cookies" mapstructure:"cookies" toml:"cookies,omitempty"`
		Debug                 bool         `json:"debug" mapstructure:"debug" toml:"debug,omitempty"`
		EnableUserSignup      bool         `json:"enable_user_signup" mapstructure:"enable_user_signup" toml:"enable_user_signup,omitempty"`
		MinimumUsernameLength uint8        `json:"minimum_username_length" mapstructure:"minimum_username_length" toml:"minimum_username_length,omitempty"`
		MinimumPasswordLength uint8        `json:"minimum_password_length" mapstructure:"minimum_password_length" toml:"minimum_password_length,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*CookieConfig)(nil)

// ValidateWithContext validates a CookieConfig struct.
func (cfg *CookieConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Name, validation.Required),
		validation.Field(&cfg.Domain, validation.Required),
		validation.Field(&cfg.Lifetime, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*PASETOConfig)(nil)

// ValidateWithContext validates a PASETOConfig struct.
func (cfg *PASETOConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Issuer, validation.Required),
		validation.Field(&cfg.LocalModeKey, validation.Required, validation.Length(pasetoKeyRequiredLength, pasetoKeyRequiredLength)),
	)
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Cookies, validation.Required),
		validation.Field(&cfg.PASETO, validation.Required),
		validation.Field(&cfg.MinimumUsernameLength, validation.Required),
		validation.Field(&cfg.MinimumPasswordLength, validation.Required),
	)
}

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuth2Client_GetID(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expected := "uint64(123)"
		oac := &OAuth2Client{
			ClientID: expected,
		}
		assert.Equal(t, expected, oac.GetID())
	})
}

func TestOAuth2Client_GetSecret(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expected := "uint64(123)"
		oac := &OAuth2Client{
			ClientSecret: expected,
		}
		assert.Equal(t, expected, oac.GetSecret())
	})
}

func TestOAuth2Client_GetDomain(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expected := "uint64(123)"
		oac := &OAuth2Client{
			RedirectURI: expected,
		}
		assert.Equal(t, expected, oac.GetDomain())
	})
}

func TestOAuth2Client_GetUserID(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		expectation := uint64(123)
		expected := "123"
		oac := &OAuth2Client{
			BelongsTo: expectation,
		}
		assert.Equal(t, expected, oac.GetUserID())
	})
}

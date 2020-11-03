package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuth2Client_GetID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		expected := "123"
		oac := &OAuth2Client{
			ClientID: expected,
		}
		assert.Equal(t, expected, oac.GetID())
	})
}

func TestOAuth2Client_GetSecret(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		expected := "123"
		oac := &OAuth2Client{
			ClientSecret: expected,
		}
		assert.Equal(t, expected, oac.GetSecret())
	})
}

func TestOAuth2Client_GetDomain(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		expected := "123"
		oac := &OAuth2Client{
			RedirectURI: expected,
		}
		assert.Equal(t, expected, oac.GetDomain())
	})
}

func TestOAuth2Client_GetUserID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		expectation := uint64(123)
		expected := fmt.Sprintf("%d", expectation)
		oac := &OAuth2Client{
			BelongsToUser: expectation,
		}
		assert.Equal(t, expected, oac.GetUserID())
	})
}

func TestOAuth2Client_HasScope(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		oac := &OAuth2Client{
			Scopes: []string{"things", "and", "stuff"},
		}

		assert.True(t, oac.HasScope(oac.Scopes[0]))
		assert.False(t, oac.HasScope("blah"))
		assert.False(t, oac.HasScope(""))
	})
}

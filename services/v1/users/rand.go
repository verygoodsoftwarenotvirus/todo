package users

import (
	"crypto/rand"
	"encoding/base32"
)

const (
	saltSize         = 16
	randomSecretSize = 64
)

// this function tests that we have appropriate access to crypto/rand
func init() {
	b := make([]byte, randomSecretSize)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

var _ secretGenerator = (*standardSecretGenerator)(nil)

type standardSecretGenerator struct{}

func (g *standardSecretGenerator) GenerateTwoFactorSecret() (string, error) {
	b := make([]byte, randomSecretSize)

	// Note that err == nil only if we read len(b) bytes.
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(b), nil
}

func (g *standardSecretGenerator) GenerateSalt() ([]byte, error) {
	b := make([]byte, saltSize)

	// Note that err == nil only if we read len(b) bytes.
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}

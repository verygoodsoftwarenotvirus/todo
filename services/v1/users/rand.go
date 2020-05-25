package users

import (
	"crypto/rand"
	"encoding/base32"
)

const (
	randomReadSize = 64
)

// this function tests that we have appropriate access to crypto/rand
func init() {
	b := make([]byte, randomReadSize)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

type standardSecretGenerator struct{}

func (g *standardSecretGenerator) GenerateTwoFactorSecret() (string, error) {
	b := make([]byte, randomReadSize)

	// Note that err == nil only if we read len(b) bytes.
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(b), nil
}

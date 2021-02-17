package delegatedclients

import (
	"crypto/rand"
	"encoding/base64"
)

func init() {
	if _, err := rand.Read(make([]byte, clientSecretSize)); err != nil {
		panic(err)
	}
}

type secretGenerator interface {
	GenerateClientID() (string, error)
	GenerateClientSecret() ([]byte, error)
}

var _ secretGenerator = (*standardSecretGenerator)(nil)

type standardSecretGenerator struct{}

func (g *standardSecretGenerator) GenerateClientID() (string, error) {
	b := make([]byte, clientIDSize)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (g *standardSecretGenerator) GenerateClientSecret() ([]byte, error) {
	b := make([]byte, clientSecretSize)

	// Note that err == nil only if we read len(b) bytes.
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}

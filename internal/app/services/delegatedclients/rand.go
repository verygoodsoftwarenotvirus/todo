package delegatedclients

import (
	"crypto/rand"
	"encoding/base32"
)

const (
	clientSecretSize = 128
	clientIDSize     = 32
)

func init() {
	b := make([]byte, clientSecretSize)
	if _, err := rand.Read(b); err != nil {
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

	return base32.StdEncoding.EncodeToString(b), nil
}

func (g *standardSecretGenerator) GenerateClientSecret() ([]byte, error) {
	b := make([]byte, clientSecretSize)

	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return b, nil
}

package delegatedclients

import (
	mock "github.com/stretchr/testify/mock"
)

var _ secretGenerator = (*mockSecretGenerator)(nil)

type mockSecretGenerator struct {
	mock.Mock
}

func (m *mockSecretGenerator) GenerateClientID() (string, error) {
	args := m.Called()

	return args.String(0), args.Error(1)
}

func (m *mockSecretGenerator) GenerateClientSecret() ([]byte, error) {
	args := m.Called()

	return args.Get(0).([]byte), args.Error(1)
}

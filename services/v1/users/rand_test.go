package users

import "github.com/stretchr/testify/mock"

type mockSecretGenerator struct {
	mock.Mock
}

func (m *mockSecretGenerator) GenerateTwoFactorSecret() (string, error) {
	args := m.Called()

	return args.String(0), args.Error(1)
}

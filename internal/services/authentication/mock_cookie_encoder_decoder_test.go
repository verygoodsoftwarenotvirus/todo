package authentication

import (
	"github.com/stretchr/testify/mock"
)

var _ cookieEncoderDecoder = (*mockCookieEncoderDecoder)(nil)

type mockCookieEncoderDecoder struct {
	mock.Mock
}

func (m *mockCookieEncoderDecoder) Encode(name string, value interface{}) (string, error) {
	args := m.Called(name, value)
	return args.String(0), args.Error(1)
}

func (m *mockCookieEncoderDecoder) Decode(name, value string, dst interface{}) error {
	args := m.Called(name, value, dst)
	return args.Error(0)
}

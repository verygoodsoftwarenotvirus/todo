package auth

import (
	"context"
	"net/http"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ OAuth2ClientValidator = (*mockOAuth2ClientValidator)(nil)

type mockOAuth2ClientValidator struct {
	mock.Mock
}

func (m *mockOAuth2ClientValidator) ExtractOAuth2ClientFromRequest(ctx context.Context, req *http.Request) (*models.OAuth2Client, error) {
	args := m.Called(req)
	return args.Get(0).(*models.OAuth2Client), args.Error(1)
}

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

var _ http.Handler = (*MockHTTPHandler)(nil)

type MockHTTPHandler struct {
	mock.Mock
}

func (m *MockHTTPHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

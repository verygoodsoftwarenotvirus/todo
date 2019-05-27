package auth

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ oauth2ClientValidator = (*mockOAuth2ClientValidator)(nil)

type mockOAuth2ClientValidator struct {
	mock.Mock
}

func (m *mockOAuth2ClientValidator) RequestIsAuthenticated(req *http.Request) (*models.OAuth2Client, error) {
	args := m.Called(req)
	return args.Get(0).(*models.OAuth2Client), args.Error(1)
}

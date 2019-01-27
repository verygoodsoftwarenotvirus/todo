package mock

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/mock"
)

var _ database.Database = (*mockDatabase)(nil)

func NewMockDatabase() *mockDatabase {
	md := &mockDatabase{
		ItemHandler:         &mockmodels.MockItemHandler{},
		UserHandler:         &mockmodels.MockUserHandler{},
		Oauth2ClientHandler: &mockmodels.MockOauth2ClientHandler{},
	}
	return md
}

type mockDatabase struct {
	mock.Mock

	models.ItemHandler
	models.UserHandler
	models.Oauth2ClientHandler
}

func (m *mockDatabase) Migrate(schemaDir string) error {
	return m.Called(schemaDir).Error(0)
}

package mock

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/mock"
)

var _ database.Database = (*Database)(nil)

// NewMockDatabase builds a new MockDatabase,
func NewMockDatabase() *Database {
	md := &Database{
		ItemHandler:         &mockmodels.MockItemHandler{},
		UserHandler:         &mockmodels.MockUserHandler{},
		OAuth2ClientHandler: &mockmodels.MockOauth2ClientHandler{},
	}
	return md
}

// Database is a mock database
type Database struct {
	mock.Mock

	models.ItemHandler
	models.UserHandler
	models.OAuth2ClientHandler
}

// Migrate mocks the call to MockDatabase
func (m *Database) Migrate(schemaDir database.SchemaDirectory) error {
	return m.Called(schemaDir).Error(0)
}

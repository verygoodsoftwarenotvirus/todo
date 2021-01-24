package mock

import (
	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.OAuth2ClientSQLQueryBuilder = (*OAuth2ClientSQLQueryBuilder)(nil)

// OAuth2ClientSQLQueryBuilder is a mocked types.OAuth2ClientSQLQueryBuilder for testing.
type OAuth2ClientSQLQueryBuilder struct {
	mock.Mock
}

// BuildGetOAuth2ClientByClientIDQuery implements our interface.
func (m *OAuth2ClientSQLQueryBuilder) BuildGetOAuth2ClientByClientIDQuery(clientID string) (query string, args []interface{}) {
	returnArgs := m.Called(clientID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetBatchOfOAuth2ClientsQuery implements our interface.
func (m *OAuth2ClientSQLQueryBuilder) BuildGetBatchOfOAuth2ClientsQuery(beginID, endID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(beginID, endID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetOAuth2ClientQuery implements our interface.
func (m *OAuth2ClientSQLQueryBuilder) BuildGetOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(clientID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAllOAuth2ClientsCountQuery implements our interface.
func (m *OAuth2ClientSQLQueryBuilder) BuildGetAllOAuth2ClientsCountQuery() string {
	returnArgs := m.Called()

	return returnArgs.String(0)
}

// BuildGetOAuth2ClientsQuery implements our interface.
func (m *OAuth2ClientSQLQueryBuilder) BuildGetOAuth2ClientsQuery(userID uint64, filter *types.QueryFilter) (query string, args []interface{}) {
	returnArgs := m.Called(userID, filter)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildCreateOAuth2ClientQuery implements our interface.
func (m *OAuth2ClientSQLQueryBuilder) BuildCreateOAuth2ClientQuery(input *types.OAuth2ClientCreationInput) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildUpdateOAuth2ClientQuery implements our interface.
func (m *OAuth2ClientSQLQueryBuilder) BuildUpdateOAuth2ClientQuery(input *types.OAuth2Client) (query string, args []interface{}) {
	returnArgs := m.Called(input)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildArchiveOAuth2ClientQuery implements our interface.
func (m *OAuth2ClientSQLQueryBuilder) BuildArchiveOAuth2ClientQuery(clientID, userID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(clientID, userID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

// BuildGetAuditLogEntriesForOAuth2ClientQuery implements our interface.
func (m *OAuth2ClientSQLQueryBuilder) BuildGetAuditLogEntriesForOAuth2ClientQuery(clientID uint64) (query string, args []interface{}) {
	returnArgs := m.Called(clientID)

	return returnArgs.String(0), returnArgs.Get(1).([]interface{})
}

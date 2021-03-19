package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditLogEntries(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(auditLogEntriesTestSuite))
}

type auditLogEntriesTestSuite struct {
	suite.Suite

	ctx                      context.Context
	filter                   *types.QueryFilter
	exampleAuditLogEntry     *types.AuditLogEntry
	exampleInput             *types.AuditLogEntryCreationInput
	exampleAuditLogEntryList *types.AuditLogEntryList
	expectedPath             string
}

var _ suite.SetupTestSuite = (*auditLogEntriesTestSuite)(nil)

func (s *auditLogEntriesTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.expectedPath = "/api/v1/_admin_/audit_log"
	s.filter = (*types.QueryFilter)(nil)
	s.exampleAuditLogEntry = fakes.BuildFakeAuditLogEntry()
	s.exampleInput = fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(s.exampleAuditLogEntry)
	s.exampleAuditLogEntryList = fakes.BuildFakeAuditLogEntryList()
}

func (s *auditLogEntriesTestSuite) TestV1Client_GetAuditLogEntries() {
	const (
		expectedMethod = http.MethodGet
	)

	spec := newRequestSpec(true, expectedMethod, "includeArchived=false&limit=20&page=1&sortBy=asc", s.expectedPath)

	s.Run("standard", func() {
		t := s.T()

		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAuditLogEntryList)
		actual, err := c.GetAuditLogEntries(s.ctx, s.filter)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleAuditLogEntryList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogEntries(s.ctx, s.filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetAuditLogEntries(s.ctx, s.filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

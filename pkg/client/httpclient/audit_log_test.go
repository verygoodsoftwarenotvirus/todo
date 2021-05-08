package httpclient

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
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
	exampleAuditLogEntryList *types.AuditLogEntryList
}

var _ suite.SetupTestSuite = (*auditLogEntriesTestSuite)(nil)

func (s *auditLogEntriesTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.filter = (*types.QueryFilter)(nil)
	s.exampleAuditLogEntry = fakes.BuildFakeAuditLogEntry()
	s.exampleAuditLogEntryList = fakes.BuildFakeAuditLogEntryList()
}

func (s *auditLogEntriesTestSuite) TestClient_GetAuditLogEntry() {
	const expectedPath = "/api/v1/admin/audit_log/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, s.exampleAuditLogEntry.ID)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAuditLogEntry)

		actual, err := c.GetAuditLogEntry(s.ctx, s.exampleAuditLogEntry.ID)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleAuditLogEntry, actual)
	})

	s.Run("with invalid entry ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.GetAuditLogEntry(s.ctx, 0)
		assert.Error(t, err, " error should be returned")
		assert.Nil(t, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.GetAuditLogEntry(s.ctx, s.exampleAuditLogEntry.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, s.exampleAuditLogEntry.ID)
		c := buildTestClientWithInvalidResponse(t, spec)

		actual, err := c.GetAuditLogEntry(s.ctx, s.exampleAuditLogEntry.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *auditLogEntriesTestSuite) TestClient_GetAuditLogEntries() {
	const expectedPath = "/api/v1/admin/audit_log"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAuditLogEntryList)

		actual, err := c.GetAuditLogEntries(s.ctx, s.filter)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleAuditLogEntryList, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.GetAuditLogEntries(s.ctx, s.filter)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)
		c := buildTestClientWithInvalidResponse(t, spec)

		actual, err := c.GetAuditLogEntries(s.ctx, s.filter)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

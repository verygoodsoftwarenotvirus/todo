package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	exampleAuditLogEntry     *types.AuditLogEntry
	exampleInput             *types.AuditLogEntryCreationInput
	exampleAuditLogEntryList *types.AuditLogEntryList
}

var _ suite.SetupTestSuite = (*auditLogEntriesTestSuite)(nil)

func (s *auditLogEntriesTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleAuditLogEntry = fakes.BuildFakeAuditLogEntry()
	s.exampleInput = fakes.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(s.exampleAuditLogEntry)
	s.exampleAuditLogEntryList = fakes.BuildFakeAuditLogEntryList()
}

func (s *auditLogEntriesTestSuite) TestV1Client_BuildGetAuditLogEntriesRequest() {
	const expectedPath = "/api/v1/_admin_/audit_log"

	s.Run("happy path", func() {
		t := s.T()

		ctx := context.Background()
		filter := (*types.QueryFilter)(nil)
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetAuditLogEntriesRequest(ctx, filter)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)
		assertRequestQuality(t, actual, spec)
	})
}

func (s *auditLogEntriesTestSuite) TestV1Client_GetAuditLogEntries() {
	const (
		expectedPath   = "/api/v1/_admin_/audit_log"
		expectedMethod = http.MethodGet
	)

	spec := newRequestSpec(true, expectedMethod, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)
	filter := (*types.QueryFilter)(nil)

	s.Run("happy path", func() {
		t := s.T()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode(s.exampleAuditLogEntryList))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogEntries(s.ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleAuditLogEntryList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogEntries(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogEntries(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

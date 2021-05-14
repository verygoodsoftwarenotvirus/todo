package integration

import (
	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

func (s *TestSuite) TestAuditLogEntryListing() {
	s.runForEachClientExcept("should be able to be read in a list by an admin", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			actual, err := testClients.admin.GetAuditLogEntries(ctx, nil)
			requireNotNilAndNoProblems(t, actual, err)

			assert.NotEmpty(t, actual.Entries)
		}
	})
}

func (s *TestSuite) TestAuditLogEntryReading() {
	s.runForEachClientExcept("should be able to be read as an individual by an admin", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			actual, err := testClients.admin.GetAuditLogEntries(ctx, nil)
			requireNotNilAndNoProblems(t, actual, err)

			for _, x := range actual.Entries {
				y, entryFetchErr := testClients.admin.GetAuditLogEntry(ctx, x.ID)
				requireNotNilAndNoProblems(t, y, entryFetchErr)
			}

			assert.NotEmpty(t, actual.Entries)
		}
	})
}

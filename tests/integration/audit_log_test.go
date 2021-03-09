package integration

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func (s *TestSuite) TestAuditLogEntryListing() {
	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be able to be read in a list by an admin via %s", authType), func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			actual, err := testClients.admin.GetAuditLogEntries(ctx, nil)
			requireNotNilAndNoProblems(t, actual, err)

			assert.NotEmpty(t, actual.Entries)
		})
	}
}

func (s *TestSuite) TestAuditLogEntryReading() {
	for a, c := range s.eachClient() {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("should be able to be read as an individual by an admin via %s", authType), func() {
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
		})
	}
}

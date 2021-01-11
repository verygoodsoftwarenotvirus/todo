package integration

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func TestAuditLogEntries(test *testing.T) {
	test.Parallel()

	test.Run("Listing", func(t *testing.T) {
		t.Parallel()

		t.Run("should be able to be read in a list by an admin", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			actual, err := adminClient.GetAuditLogEntries(ctx, nil)
			checkValueAndError(t, actual, err)

			assert.NotEmpty(t, actual.Entries)
		})
	})

	test.Run("Reading", func(t *testing.T) {
		t.Parallel()

		t.Run("should be able to be read as an individual by an admin", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			adminClientLock.Lock()
			defer adminClientLock.Unlock()

			actual, err := adminClient.GetAuditLogEntries(ctx, nil)
			checkValueAndError(t, actual, err)

			for _, x := range actual.Entries {
				y, entryFetchErr := adminClient.GetAuditLogEntry(ctx, x.ID)
				checkValueAndError(t, y, entryFetchErr)
			}

			assert.NotEmpty(t, actual.Entries)
		})
	})
}

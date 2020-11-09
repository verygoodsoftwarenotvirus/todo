package integration

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"

	"github.com/stretchr/testify/assert"
)

func TestAuditLogEntries(test *testing.T) {
	test.Run("Listing", func(t *testing.T) {
		t.Run("should be able to be read in a list by an admin", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			actual, err := adminClient.GetAuditLogEntries(ctx, nil)
			checkValueAndError(t, actual, err)

			assert.NotEmpty(t, actual.Entries)
		})
	})
}

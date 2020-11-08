package integration

import (
	"context"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"

	"github.com/stretchr/testify/assert"
)

func checkAuditLogEntrySliceEquality(t *testing.T, expectedSlice, actualSlice []models.AuditLogEntry) {
	t.Helper()

	require.Equal(t, len(expectedSlice), len(actualSlice))

	for i := 0; i < len(expectedSlice); i++ {
		assert.Equal(t, expectedSlice[i].EventType, actualSlice[i].EventType, "expected EventType for ID %d to be %v, but it was %v ", expectedSlice[i].ID, expectedSlice[i].EventType, actualSlice[i].EventType)
		for k, v := range expectedSlice[i].Context {
			if k == "changes" || k == "created" {
				assert.NotNil(t, actualSlice[i].Context[k])
			} else {
				assert.EqualValues(t, v, actualSlice[i].Context[k], "key %q in event %d is %v, expected %v", k, i, actualSlice[i].Context[k], v)
			}
		}
	}
}

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

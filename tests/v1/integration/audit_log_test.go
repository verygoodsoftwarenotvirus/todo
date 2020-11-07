package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
)

func TestAuditLogEntries(test *testing.T) {
	test.Run("Listing", func(t *testing.T) {
		t.Run("should be able to be read in a list by an admin", func(t *testing.T) {
			ctx, span := tracing.StartSpan(context.Background(), t.Name())
			defer span.End()

			//// Create items.
			//for i := 0; i < 5; i++ {
			//	// Create item.
			//	exampleItem := fakemodels.BuildFakeItem()
			//	exampleItemInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)
			//	createdItem, itemCreationErr := todoClient.CreateItem(ctx, exampleItemInput)
			//	checkValueAndError(t, createdItem, itemCreationErr)
			//}

			// Assert item list equality.
			actual, err := adminClient.GetAuditLogEntries(ctx, nil)
			checkValueAndError(t, actual, err)

			assert.NotEmpty(t, actual.Entries)
		})
	})
}

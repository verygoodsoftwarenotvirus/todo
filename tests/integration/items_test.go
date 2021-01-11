package integration

import (
	"context"
	"fmt"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/converters"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkItemEquality(t *testing.T, expected, actual *types.Item) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected BucketName for item #%d to be %v, but it was %v ", expected.ID, expected.Name, actual.Name)
	assert.Equal(t, expected.Details, actual.Details, "expected Details for item #%d to be %v, but it was %v ", expected.ID, expected.Details, actual.Details)
	assert.NotZero(t, actual.CreatedOn)
}

func TestItems(test *testing.T) {
	test.Parallel()

	test.Run("Creating", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be createable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create item.
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := testClient.CreateItem(ctx, exampleItemInput)
			checkValueAndError(t, createdItem, err)

			// Assert item equality.
			checkItemEquality(t, exampleItem, createdItem)

			// Clean up.
			err = testClient.ArchiveItem(ctx, createdItem.ID)
			assert.NoError(t, err)

			adminClientLock.Lock()
			defer adminClientLock.Unlock()
			auditLogEntries, err := adminClient.GetAuditLogForItem(ctx, createdItem.ID)

			require.NoError(t, err)
			require.Len(t, auditLogEntries, 2)

			expectedEventTypes := []string{
				audit.ItemCreationEvent,
				audit.ItemArchiveEvent,
			}
			actualEventTypes := []string{}

			for _, e := range auditLogEntries {
				actualEventTypes = append(actualEventTypes, e.EventType)
				require.Contains(t, e.Context, audit.ItemAssignmentKey)
				assert.EqualValues(t, createdItem.ID, e.Context[audit.ItemAssignmentKey])
			}

			assert.Subset(t, expectedEventTypes, actualEventTypes)
		})
	})

	test.Run("Listing", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be able to be read in a list", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create items.
			var expected []*types.Item
			for i := 0; i < 5; i++ {
				// Create item.
				exampleItem := fakes.BuildFakeItem()
				exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
				createdItem, itemCreationErr := testClient.CreateItem(ctx, exampleItemInput)
				checkValueAndError(t, createdItem, itemCreationErr)

				expected = append(expected, createdItem)
			}

			// Assert item list equality.
			actual, err := testClient.GetItems(ctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual.Items),
				"expected %d to be <= %d",
				len(expected),
				len(actual.Items),
			)

			// Clean up.
			for _, createdItem := range actual.Items {
				err = testClient.ArchiveItem(ctx, createdItem.ID)
				assert.NoError(t, err)
			}
		})
	})

	test.Run("Searching", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("should be able to be search for items", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create items.
			exampleItem := fakes.BuildFakeItem()
			var expected []*types.Item
			for i := 0; i < 5; i++ {
				// Create item.
				exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
				exampleItemInput.Name = fmt.Sprintf("%s %d", exampleItemInput.Name, i)
				createdItem, itemCreationErr := testClient.CreateItem(ctx, exampleItemInput)
				checkValueAndError(t, createdItem, itemCreationErr)

				expected = append(expected, createdItem)
			}

			exampleLimit := uint8(20)

			// Assert item list equality.
			actual, err := testClient.SearchItems(ctx, exampleItem.Name, exampleLimit)
			checkValueAndError(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual),
				"expected results length %d to be <= %d",
				len(expected),
				len(actual),
			)

			// Clean up.
			for _, createdItem := range expected {
				err = testClient.ArchiveItem(ctx, createdItem.ID)
				assert.NoError(t, err)
			}
		})

		subtest.Run("should only receive your own items", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			exampleLimit := uint8(20)
			_, clientA := createUserAndClientForTest(ctx, t)
			_, clientB := createUserAndClientForTest(ctx, t)

			// Create items for user A.
			exampleItemA := fakes.BuildFakeItem()
			var createdForA []*types.Item
			for i := 0; i < 5; i++ {
				// Create item.
				exampleItemInputA := fakes.BuildFakeItemCreationInputFromItem(exampleItemA)
				exampleItemInputA.Name = fmt.Sprintf("%s %d", exampleItemInputA.Name, i)

				createdItem, itemCreationErr := clientA.CreateItem(ctx, exampleItemInputA)
				checkValueAndError(t, createdItem, itemCreationErr)

				createdForA = append(createdForA, createdItem)
			}
			query := exampleItemA.Name

			// Create items for user B.
			exampleItemB := fakes.BuildFakeItem()
			exampleItemB.Name = reverse(exampleItemA.Name)
			var createdForB []*types.Item
			for i := 0; i < 5; i++ {
				// Create item.
				exampleItemInputB := fakes.BuildFakeItemCreationInputFromItem(exampleItemB)
				exampleItemInputB.Name = fmt.Sprintf("%s %d", exampleItemInputB.Name, i)

				createdItem, itemCreationErr := clientB.CreateItem(ctx, exampleItemInputB)
				checkValueAndError(t, createdItem, itemCreationErr)

				createdForB = append(createdForB, createdItem)
			}

			expected := createdForA

			// Assert item list equality.
			actual, err := clientA.SearchItems(ctx, query, exampleLimit)
			checkValueAndError(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual),
				"expected results length %d to be <= %d",
				len(expected),
				len(actual),
			)

			// Clean up.
			for _, createdItem := range createdForA {
				err = clientA.ArchiveItem(ctx, createdItem.ID)
				assert.NoError(t, err)
			}

			for _, createdItem := range createdForB {
				err = clientB.ArchiveItem(ctx, createdItem.ID)
				assert.NoError(t, err)
			}
		})
	})

	test.Run("ExistenceChecking", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return false with no error when checking something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Attempt to fetch nonexistent item.
			actual, err := testClient.ItemExists(ctx, nonexistentID)
			assert.NoError(t, err)
			assert.False(t, actual)
		})

		subtest.Run("it should return true with no error when the relevant item exists", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create item.
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := testClient.CreateItem(ctx, exampleItemInput)
			checkValueAndError(t, createdItem, err)

			// Fetch item.
			actual, err := testClient.ItemExists(ctx, createdItem.ID)
			assert.NoError(t, err)
			assert.True(t, actual)

			// Clean up item.
			assert.NoError(t, testClient.ArchiveItem(ctx, createdItem.ID))
		})
	})

	test.Run("Reading", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to read something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Attempt to fetch nonexistent item.
			_, err := testClient.GetItem(ctx, nonexistentID)
			assert.Error(t, err)
		})

		subtest.Run("it should be readable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create item.
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := testClient.CreateItem(ctx, exampleItemInput)
			checkValueAndError(t, createdItem, err)

			// Fetch item.
			actual, err := testClient.GetItem(ctx, createdItem.ID)
			checkValueAndError(t, actual, err)

			// Assert item equality.
			checkItemEquality(t, exampleItem, actual)

			// Clean up item.
			assert.NoError(t, testClient.ArchiveItem(ctx, createdItem.ID))
		})
	})

	test.Run("Updating", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to update something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			exampleItem := fakes.BuildFakeItem()
			exampleItem.ID = nonexistentID

			assert.Error(t, testClient.UpdateItem(ctx, exampleItem))
		})

		subtest.Run("it should be updatable", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create item.
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := testClient.CreateItem(ctx, exampleItemInput)
			checkValueAndError(t, createdItem, err)

			// Change item.
			createdItem.Update(converters.ConvertItemToItemUpdateInput(exampleItem))
			assert.NoError(t, testClient.UpdateItem(ctx, createdItem))

			// Fetch item.
			actual, err := testClient.GetItem(ctx, createdItem.ID)
			checkValueAndError(t, actual, err)

			// Assert item equality.
			checkItemEquality(t, exampleItem, actual)
			assert.NotNil(t, actual.LastUpdatedOn)

			// Clean up item.
			assert.NoError(t, testClient.ArchiveItem(ctx, createdItem.ID))

			adminClientLock.Lock()
			defer adminClientLock.Unlock()
			auditLogEntries, err := adminClient.GetAuditLogForItem(ctx, createdItem.ID)

			require.Len(t, auditLogEntries, 3)
			require.NoError(t, err)

			expectedEventTypes := []string{
				audit.ItemCreationEvent,
				audit.ItemUpdateEvent,
				audit.ItemArchiveEvent,
			}
			actualEventTypes := []string{}

			for _, e := range auditLogEntries {
				actualEventTypes = append(actualEventTypes, e.EventType)
				require.Contains(t, e.Context, audit.ItemAssignmentKey)
				assert.EqualValues(t, createdItem.ID, e.Context[audit.ItemAssignmentKey])
			}

			assert.Subset(t, expectedEventTypes, actualEventTypes)
		})
	})

	test.Run("Deleting", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to delete something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			assert.Error(t, testClient.ArchiveItem(ctx, nonexistentID))
		})

		subtest.Run("should be able to be deleted", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create item.
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := testClient.CreateItem(ctx, exampleItemInput)
			checkValueAndError(t, createdItem, err)

			// Clean up item.
			assert.NoError(t, testClient.ArchiveItem(ctx, createdItem.ID))
		})
	})

	test.Run("Auditing", func(subtest *testing.T) {
		subtest.Parallel()

		subtest.Run("it should return an error when trying to audit something that does not exist", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			exampleItem := fakes.BuildFakeItem()
			exampleItem.ID = nonexistentID

			adminClientLock.Lock()
			defer adminClientLock.Unlock()
			x, err := adminClient.GetAuditLogForItem(ctx, exampleItem.ID)

			assert.NoError(t, err)
			assert.Empty(t, x)
		})

		subtest.Run("it should not be auditable by a non-admin", func(t *testing.T) {
			t.Parallel()

			ctx, span := tracing.StartSpan(context.Background())
			defer span.End()

			_, testClient := createUserAndClientForTest(ctx, t)

			// Create item.
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := testClient.CreateItem(ctx, exampleItemInput)
			checkValueAndError(t, createdItem, err)

			// fetch audit log entries
			actual, err := testClient.GetAuditLogForItem(ctx, createdItem.ID)
			assert.Error(t, err)
			assert.Nil(t, actual)

			// Clean up item.
			assert.NoError(t, testClient.ArchiveItem(ctx, createdItem.ID))
		})
	})
}

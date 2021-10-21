package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

func checkItemEquality(t *testing.T, expected, actual *types.Item) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected Name for item %s to be %v, but it was %v", expected.ID, expected.Name, actual.Name)
	assert.Equal(t, expected.Details, actual.Details, "expected Details for item %s to be %v, but it was %v", expected.ID, expected.Details, actual.Details)
	assert.NotZero(t, actual.CreatedOn)
}

// convertItemToItemUpdateInput creates an ItemUpdateRequestInput struct from an item.
func convertItemToItemUpdateInput(x *types.Item) *types.ItemUpdateRequestInput {
	return &types.ItemUpdateRequestInput{
		Name:    x.Name,
		Details: x.Details,
	}
}

func (s *TestSuite) TestItems_CompleteLifecycle() {
	s.runForCookieClient("should be creatable and readable and updatable and deletable", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			stopChan := make(chan bool, 1)
			notificationsChan, err := testClients.main.SubscribeToDataChangeNotifications(ctx, stopChan)
			require.NotNil(t, notificationsChan)
			require.NoError(t, err)

			var n *types.DataChangeMessage

			t.Log("creating item")
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationRequestInputFromItem(exampleItem)
			createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
			require.NoError(t, err)
			t.Logf("item %q created", createdItemID)

			n = <-notificationsChan
			assert.Equal(t, n.DataType, types.ItemDataType)
			require.NotNil(t, n.Item)
			checkItemEquality(t, exampleItem, n.Item)

			createdItem, err := testClients.main.GetItem(ctx, createdItemID)
			requireNotNilAndNoProblems(t, createdItem, err)

			checkItemEquality(t, exampleItem, createdItem)

			t.Log("changing item")
			newItem := fakes.BuildFakeItem()
			createdItem.Update(convertItemToItemUpdateInput(newItem))
			assert.NoError(t, testClients.main.UpdateItem(ctx, createdItem))

			n = <-notificationsChan
			assert.Equal(t, n.DataType, types.ItemDataType)

			t.Log("fetching changed item")
			actual, err := testClients.main.GetItem(ctx, createdItemID)
			requireNotNilAndNoProblems(t, actual, err)

			// assert item equality
			checkItemEquality(t, newItem, actual)
			assert.NotNil(t, actual.LastUpdatedOn)

			t.Log("cleaning up item")
			assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItemID))
		}
	})

	s.runForPASETOClient("should be creatable and readable and updatable and deletable", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			var checkFunc func() bool
			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			t.Log("creating item")
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationRequestInputFromItem(exampleItem)
			createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
			require.NoError(t, err)
			t.Logf("item %q created", createdItemID)

			var createdItem *types.Item
			checkFunc = func() bool {
				createdItem, err = testClients.main.GetItem(ctx, createdItemID)
				return assert.NotNil(t, createdItem) && assert.NoError(t, err)
			}
			assert.Eventually(t, checkFunc, creationTimeout, waitPeriod)
			checkItemEquality(t, exampleItem, createdItem)

			// assert item equality
			checkItemEquality(t, exampleItem, createdItem)

			// change item
			newItem := fakes.BuildFakeItem()
			createdItem.Update(convertItemToItemUpdateInput(newItem))
			assert.NoError(t, testClients.main.UpdateItem(ctx, createdItem))

			time.Sleep(time.Second)

			// retrieve changed item
			var actual *types.Item
			checkFunc = func() bool {
				actual, err = testClients.main.GetItem(ctx, createdItemID)
				return assert.NotNil(t, createdItem) && assert.NoError(t, err)
			}
			assert.Eventually(t, checkFunc, creationTimeout, waitPeriod)

			requireNotNilAndNoProblems(t, actual, err)

			// assert item equality
			checkItemEquality(t, newItem, actual)
			assert.NotNil(t, actual.LastUpdatedOn)

			t.Log("cleaning up item")
			assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItemID))
		}
	})
}

func (s *TestSuite) TestItems_Listing() {
	s.runForCookieClient("should be readable in paginated form", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			stopChan := make(chan bool, 1)
			notificationsChan, notificationsChanErr := testClients.main.SubscribeToDataChangeNotifications(ctx, stopChan)
			require.NotNil(t, notificationsChan)
			require.NoError(t, notificationsChanErr)

			var n *types.DataChangeMessage

			t.Log("creating items")
			var expected []*types.Item
			for i := 0; i < 5; i++ {
				exampleItem := fakes.BuildFakeItem()
				exampleItemInput := fakes.BuildFakeItemCreationRequestInputFromItem(exampleItem)
				createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
				require.NoError(t, err)
				t.Logf("item %q created", createdItemID)

				n = <-notificationsChan
				assert.Equal(t, n.DataType, types.ItemDataType)
				require.NotNil(t, n.Item)
				checkItemEquality(t, exampleItem, n.Item)

				createdItem, err := testClients.main.GetItem(ctx, createdItemID)
				requireNotNilAndNoProblems(t, createdItem, err)

				expected = append(expected, createdItem)
			}

			// assert item list equality
			actual, err := testClients.main.GetItems(ctx, nil)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual.Items),
				"expected %d to be <= %d",
				len(expected),
				len(actual.Items),
			)

			t.Log("cleaning up")
			for _, createdItem := range expected {
				assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
			}
		}
	})

	s.runForPASETOClient("should be readable in paginated form", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			var checkFunc func() bool
			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			t.Log("creating items")
			var expected []*types.Item
			for i := 0; i < 5; i++ {
				exampleItem := fakes.BuildFakeItem()
				exampleItemInput := fakes.BuildFakeItemCreationRequestInputFromItem(exampleItem)
				createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
				require.NoError(t, err)

				var createdItem *types.Item
				checkFunc = func() bool {
					createdItem, err = testClients.main.GetItem(ctx, createdItemID)
					return assert.NotNil(t, createdItem) && assert.NoError(t, err)
				}
				assert.Eventually(t, checkFunc, creationTimeout, waitPeriod)
				checkItemEquality(t, exampleItem, createdItem)

				expected = append(expected, createdItem)
			}

			// assert item list equality
			actual, err := testClients.main.GetItems(ctx, nil)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual.Items),
				"expected %d to be <= %d",
				len(expected),
				len(actual.Items),
			)

			t.Log("cleaning up")
			for _, createdItem := range expected {
				assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
			}
		}
	})
}

func (s *TestSuite) TestItems_Searching() {
	s.runForCookieClient("should be able to be search for items", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			stopChan := make(chan bool, 1)
			notificationsChan, err := testClients.main.SubscribeToDataChangeNotifications(ctx, stopChan)
			require.NotNil(t, notificationsChan)
			require.NoError(t, err)

			var n *types.DataChangeMessage

			t.Log("creating items")
			var expected []*types.Item
			exampleItem := fakes.BuildFakeItem()
			searchQuery := exampleItem.Name
			for i := 0; i < 5; i++ {
				exampleItem.Name = fmt.Sprintf("%s %d", searchQuery, i)
				exampleItemInput := fakes.BuildFakeItemCreationRequestInputFromItem(exampleItem)
				createdItemID, itemCreationErr := testClients.main.CreateItem(ctx, exampleItemInput)
				require.NoError(t, itemCreationErr)
				t.Logf("item %q created", createdItemID)

				n = <-notificationsChan
				assert.Equal(t, n.DataType, types.ItemDataType)
				require.NotNil(t, n.Item)
				checkItemEquality(t, exampleItem, n.Item)

				createdItem, itemCreationErr := testClients.main.GetItem(ctx, createdItemID)
				requireNotNilAndNoProblems(t, createdItem, itemCreationErr)

				expected = append(expected, createdItem)
			}

			exampleLimit := uint8(20)

			// give the index a moment
			time.Sleep(3 * time.Second)

			// assert item list equality
			actual, err := testClients.main.SearchItems(ctx, searchQuery, exampleLimit)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual),
				"expected %d to be <= %d",
				len(expected),
				len(actual),
			)

			t.Log("cleaning up")
			for _, createdItem := range expected {
				assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
			}
		}
	})

	s.runForPASETOClient("should be able to be search for items", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			var checkFunc func() bool
			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			t.Log("creating items")
			var expected []*types.Item
			exampleItem := fakes.BuildFakeItem()
			searchQuery := exampleItem.Name
			for i := 0; i < 5; i++ {
				exampleItem.Name = fmt.Sprintf("%s %d", searchQuery, i)
				exampleItemInput := fakes.BuildFakeItemCreationRequestInputFromItem(exampleItem)
				createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
				require.NoError(t, err)
				t.Logf("item %q created", createdItemID)

				var createdItem *types.Item
				checkFunc = func() bool {
					createdItem, err = testClients.main.GetItem(ctx, createdItemID)
					return assert.NotNil(t, createdItem) && assert.NoError(t, err)
				}
				assert.Eventually(t, checkFunc, creationTimeout, waitPeriod)
				requireNotNilAndNoProblems(t, createdItem, err)

				expected = append(expected, createdItem)
			}

			exampleLimit := uint8(20)
			time.Sleep(time.Second) // give the index a moment

			// assert item list equality
			actual, err := testClients.main.SearchItems(ctx, searchQuery, exampleLimit)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual),
				"expected %d to be <= %d",
				len(expected),
				len(actual),
			)

			t.Log("cleaning up")
			for _, createdItem := range expected {
				assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
			}
		}
	})
}

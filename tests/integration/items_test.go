package integration

import (
	"fmt"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkItemEquality(t *testing.T, expected, actual *types.Item) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected Name for item %s to be %v, but it was %v ", expected.ID, expected.Name, actual.Name)
	assert.Equal(t, expected.Details, actual.Details, "expected Details for item %s to be %v, but it was %v ", expected.ID, expected.Details, actual.Details)
	assert.NotZero(t, actual.CreatedOn)
}

func (s *TestSuite) TestItems_Creating() {
	s.runForEachClientExcept("should be creatable", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create item.
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
			require.NoError(t, err)

			waitForAsynchronousStuffBecauseProperWebhookNotificationsHaveNotBeenImplementedYet()

			createdItem, err := testClients.main.GetItem(ctx, createdItemID)
			requireNotNilAndNoProblems(t, createdItem, err)

			// assert item equality
			checkItemEquality(t, exampleItem, createdItem)

			auditLogEntries, err := testClients.admin.GetAuditLogForItem(ctx, createdItem.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.ItemCreationEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdItem.ID, audit.ItemAssignmentKey)

			// Clean up item.
			assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
		}
	})
}

func (s *TestSuite) TestItems_Listing() {
	s.runForEachClientExcept("should be readable in paginated form", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// create items
			var expected []*types.Item
			for i := 0; i < 5; i++ {
				exampleItem := fakes.BuildFakeItem()
				exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

				createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
				require.NoError(t, err)

				waitForAsynchronousStuffBecauseProperWebhookNotificationsHaveNotBeenImplementedYet()

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

			// clean up
			for _, createdItem := range actual.Items {
				assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
			}
		}
	})
}

func (s *TestSuite) TestItems_Searching() {
	s.T().SkipNow()

	s.runForEachClientExcept("should be able to be search for items", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// create items
			exampleItem := fakes.BuildFakeItem()
			var expected []*types.Item
			for i := 0; i < 5; i++ {
				exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
				exampleItemInput.Name = fmt.Sprintf("%s %d", exampleItemInput.Name, i)

				createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
				require.NoError(t, err)

				waitForAsynchronousStuffBecauseProperWebhookNotificationsHaveNotBeenImplementedYet()

				createdItem, err := testClients.main.GetItem(ctx, createdItemID)
				requireNotNilAndNoProblems(t, createdItem, err)

				expected = append(expected, createdItem)
			}

			exampleLimit := uint8(20)

			// assert item list equality
			actual, err := testClients.main.SearchItems(ctx, exampleItem.Name, exampleLimit)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual),
				"expected results length %d to be <= %d",
				len(expected),
				len(actual),
			)

			// clean up
			for _, createdItem := range expected {
				assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
			}
		}
	})
}

func (s *TestSuite) TestItems_Searching_ReturnsOnlyItemsThatBelongToYou() {
	s.T().SkipNow()

	s.runForEachClientExcept("should only receive your own items", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// create items
			exampleItem := fakes.BuildFakeItem()
			var expected []*types.Item
			for i := 0; i < 5; i++ {
				exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
				exampleItemInput.Name = fmt.Sprintf("%s %d", exampleItemInput.Name, i)

				createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
				require.NoError(t, err)

				waitForAsynchronousStuffBecauseProperWebhookNotificationsHaveNotBeenImplementedYet()

				createdItem, err := testClients.main.GetItem(ctx, createdItemID)
				requireNotNilAndNoProblems(t, createdItem, err)

				expected = append(expected, createdItem)
			}

			exampleLimit := uint8(20)

			// assert item list equality
			actual, err := testClients.main.SearchItems(ctx, exampleItem.Name, exampleLimit)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual),
				"expected results length %d to be <= %d",
				len(expected),
				len(actual),
			)

			// clean up
			for _, createdItem := range expected {
				assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
			}
		}
	})
}

func (s *TestSuite) TestItems_ExistenceChecking_ReturnsFalseForNonexistentItem() {
	s.runForEachClientExcept("should not return an error for nonexistent item", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			actual, err := testClients.main.ItemExists(ctx, nonexistentID)
			assert.NoError(t, err)
			assert.False(t, actual)
		}
	})
}

//
func (s *TestSuite) TestItems_ExistenceChecking_ReturnsTrueForValidItem() {
	s.runForEachClientExcept("should not return an error for existent item", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// create item
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
			require.NoError(t, err)

			waitForAsynchronousStuffBecauseProperWebhookNotificationsHaveNotBeenImplementedYet()

			createdItem, err := testClients.main.GetItem(ctx, createdItemID)
			requireNotNilAndNoProblems(t, createdItem, err)

			// retrieve item
			actual, err := testClients.main.ItemExists(ctx, createdItem.ID)
			assert.NoError(t, err)
			assert.True(t, actual)

			// clean up item
			assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
		}
	})
}

func (s *TestSuite) TestItems_Reading_Returns404ForNonexistentItem() {
	s.runForEachClientExcept("it should return an error when trying to read an item that does not exist", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			_, err := testClients.main.GetItem(ctx, nonexistentID)
			assert.Error(t, err)
		}
	})
}

func (s *TestSuite) TestItems_Reading() {
	s.runForEachClientExcept("it should be readable", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// create item
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
			require.NoError(t, err)

			waitForAsynchronousStuffBecauseProperWebhookNotificationsHaveNotBeenImplementedYet()

			createdItem, err := testClients.main.GetItem(ctx, createdItemID)
			requireNotNilAndNoProblems(t, createdItem, err)

			// retrieve item
			actual, err := testClients.main.GetItem(ctx, createdItem.ID)
			requireNotNilAndNoProblems(t, actual, err)

			// assert item equality
			checkItemEquality(t, exampleItem, actual)

			// clean up item
			assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
		}
	})
}

func (s *TestSuite) TestItems_Updating_Returns404ForNonexistentItem() {
	s.runForEachClientExcept("it should return an error when trying to update something that does not exist", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			exampleItem := fakes.BuildFakeItem()
			exampleItem.ID = nonexistentID

			assert.Error(t, testClients.main.UpdateItem(ctx, exampleItem))
		}
	})
}

// convertItemToItemUpdateInput creates an ItemUpdateInput struct from an item.
func convertItemToItemUpdateInput(x *types.Item) *types.ItemUpdateInput {
	return &types.ItemUpdateInput{
		Name:    x.Name,
		Details: x.Details,
	}
}

func (s *TestSuite) TestItems_Updating() {
	s.runForEachClientExcept("it should be possible to update an item", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// create item
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
			require.NoError(t, err)

			waitForAsynchronousStuffBecauseProperWebhookNotificationsHaveNotBeenImplementedYet()

			createdItem, err := testClients.main.GetItem(ctx, createdItemID)
			requireNotNilAndNoProblems(t, createdItem, err)

			// change item
			createdItem.Update(convertItemToItemUpdateInput(exampleItem))
			assert.NoError(t, testClients.main.UpdateItem(ctx, createdItem))

			// retrieve changed item
			actual, err := testClients.main.GetItem(ctx, createdItem.ID)
			requireNotNilAndNoProblems(t, actual, err)

			// assert item equality
			checkItemEquality(t, exampleItem, actual)
			assert.NotNil(t, actual.LastUpdatedOn)

			auditLogEntries, err := testClients.admin.GetAuditLogForItem(ctx, createdItem.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.ItemCreationEvent},
				{EventType: audit.ItemUpdateEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdItem.ID, audit.ItemAssignmentKey)

			// clean up item
			assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
		}
	})
}

func (s *TestSuite) TestItems_Archiving_Returns404ForNonexistentItem() {
	s.runForEachClientExcept("it should return an error when trying to delete something that does not exist", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			assert.Error(t, testClients.main.ArchiveItem(ctx, nonexistentID))
		}
	})
}

func (s *TestSuite) TestItems_Archiving() {
	s.runForEachClientExcept("it should be possible to delete an item", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// create item
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItemID, err := testClients.main.CreateItem(ctx, exampleItemInput)
			require.NoError(t, err)

			waitForAsynchronousStuffBecauseProperWebhookNotificationsHaveNotBeenImplementedYet()

			createdItem, err := testClients.main.GetItem(ctx, createdItemID)
			requireNotNilAndNoProblems(t, createdItem, err)

			// clean up item
			assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItemID))

			auditLogEntries, err := testClients.admin.GetAuditLogForItem(ctx, createdItemID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.ItemCreationEvent},
				{EventType: audit.ItemArchiveEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdItemID, audit.ItemAssignmentKey)
		}
	})
}

func (s *TestSuite) TestItems_Auditing_Returns404ForNonexistentItem() {
	s.runForEachClientExcept("it should return an error when trying to audit something that does not exist", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			x, err := testClients.admin.GetAuditLogForItem(ctx, nonexistentID)

			assert.NoError(t, err)
			assert.Empty(t, x)
		}
	})
}

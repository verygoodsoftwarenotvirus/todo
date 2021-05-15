package integration

import (
	"fmt"
	"testing"

	audit "gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/converters"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

func checkItemEquality(t *testing.T, expected, actual *types.Item) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected BucketName for item #%d to be %v, but it was %v ", expected.ID, expected.Name, actual.Name)
	assert.Equal(t, expected.Details, actual.Details, "expected Details for item #%d to be %v, but it was %v ", expected.ID, expected.Details, actual.Details)
	assert.NotZero(t, actual.CreatedOn)
}

func (s *TestSuite) TestItems_Creating() {
	s.runForEachClientExcept("should be creatable", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := testClients.main.CreateItem(ctx, exampleItemInput)
			requireNotNilAndNoProblems(t, createdItem, err)

			// Assert item equality.
			checkItemEquality(t, exampleItem, createdItem)

			auditLogEntries, err := testClients.admin.GetAuditLogForItem(ctx, createdItem.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.ItemCreationEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdItem.ID, audit.ItemAssignmentKey)

			// Clean up.
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

			// Create items.
			var expected []*types.Item
			for i := 0; i < 5; i++ {
				// Create item.
				exampleItem := fakes.BuildFakeItem()
				exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
				createdItem, itemCreationErr := testClients.main.CreateItem(ctx, exampleItemInput)
				requireNotNilAndNoProblems(t, createdItem, itemCreationErr)

				expected = append(expected, createdItem)
			}

			// Assert item list equality.
			actual, err := testClients.main.GetItems(ctx, nil)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual.Items),
				"expected %d to be <= %d",
				len(expected),
				len(actual.Items),
			)

			// Clean up.
			for _, createdItem := range actual.Items {
				assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
			}
		}
	})
}

func (s *TestSuite) TestItems_Searching() {
	s.runForEachClientExcept("should be able to be search for items", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create items.
			exampleItem := fakes.BuildFakeItem()
			var expected []*types.Item
			for i := 0; i < 5; i++ {
				// Create item.
				exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
				exampleItemInput.Name = fmt.Sprintf("%s %d", exampleItemInput.Name, i)
				createdItem, itemCreationErr := testClients.main.CreateItem(ctx, exampleItemInput)
				requireNotNilAndNoProblems(t, createdItem, itemCreationErr)

				expected = append(expected, createdItem)
			}

			exampleLimit := uint8(20)

			// Assert item list equality.
			actual, err := testClients.main.SearchItems(ctx, exampleItem.Name, exampleLimit)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual),
				"expected results length %d to be <= %d",
				len(expected),
				len(actual),
			)

			// Clean up.
			for _, createdItem := range expected {
				assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
			}
		}
	})
}

func (s *TestSuite) TestItems_Searching_ReturnsOnlyItemsThatBelongToYou() {
	s.runForEachClientExcept("should only receive your own items", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create items.
			exampleItem := fakes.BuildFakeItem()
			var expected []*types.Item
			for i := 0; i < 5; i++ {
				// Create item.
				exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
				exampleItemInput.Name = fmt.Sprintf("%s %d", exampleItemInput.Name, i)
				createdItem, itemCreationErr := testClients.main.CreateItem(ctx, exampleItemInput)
				requireNotNilAndNoProblems(t, createdItem, itemCreationErr)

				expected = append(expected, createdItem)
			}

			exampleLimit := uint8(20)

			// Assert item list equality.
			actual, err := testClients.main.SearchItems(ctx, exampleItem.Name, exampleLimit)
			requireNotNilAndNoProblems(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual),
				"expected results length %d to be <= %d",
				len(expected),
				len(actual),
			)

			// Clean up.
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

			// Attempt to fetch nonexistent item.
			actual, err := testClients.main.ItemExists(ctx, nonexistentID)
			assert.NoError(t, err)
			assert.False(t, actual)
		}
	})
}

func (s *TestSuite) TestItems_ExistenceChecking_ReturnsTrueForValidItem() {
	s.runForEachClientExcept("should not return an error for existent item", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create item.
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := testClients.main.CreateItem(ctx, exampleItemInput)
			requireNotNilAndNoProblems(t, createdItem, err)

			// Fetch item.
			actual, err := testClients.main.ItemExists(ctx, createdItem.ID)
			assert.NoError(t, err)
			assert.True(t, actual)

			// Clean up item.
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

			// Attempt to fetch nonexistent item.
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

			// Create item.
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := testClients.main.CreateItem(ctx, exampleItemInput)
			requireNotNilAndNoProblems(t, createdItem, err)

			// Fetch item.
			actual, err := testClients.main.GetItem(ctx, createdItem.ID)
			requireNotNilAndNoProblems(t, actual, err)

			// Assert item equality.
			checkItemEquality(t, exampleItem, actual)

			// Clean up item.
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

func (s *TestSuite) TestItems_Updating() {
	s.runForEachClientExcept("it should be possible to update an item", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create item.
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := testClients.main.CreateItem(ctx, exampleItemInput)
			requireNotNilAndNoProblems(t, createdItem, err)

			// Change item.
			createdItem.Update(converters.ConvertItemToItemUpdateInput(exampleItem))
			assert.NoError(t, testClients.main.UpdateItem(ctx, createdItem))

			// Fetch item.
			actual, err := testClients.main.GetItem(ctx, createdItem.ID)
			requireNotNilAndNoProblems(t, actual, err)

			// Assert item equality.
			checkItemEquality(t, exampleItem, actual)
			assert.NotNil(t, actual.LastUpdatedOn)

			auditLogEntries, err := testClients.admin.GetAuditLogForItem(ctx, createdItem.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.ItemCreationEvent},
				{EventType: audit.ItemUpdateEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdItem.ID, audit.ItemAssignmentKey)

			// Clean up item.
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

			// Create item.
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := testClients.main.CreateItem(ctx, exampleItemInput)
			requireNotNilAndNoProblems(t, createdItem, err)

			// Clean up item.
			assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))

			auditLogEntries, err := testClients.admin.GetAuditLogForItem(ctx, createdItem.ID)
			require.NoError(t, err)

			expectedAuditLogEntries := []*types.AuditLogEntry{
				{EventType: audit.ItemCreationEvent},
				{EventType: audit.ItemArchiveEvent},
			}
			validateAuditLogEntries(t, expectedAuditLogEntries, auditLogEntries, createdItem.ID, audit.ItemAssignmentKey)
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

func (s *TestSuite) TestItems_Auditing() {
	s.runForEachClientExcept("it should not be auditable by a non-admin", func(testClients *testClientWrapper) func() {
		return func() {
			t := s.T()

			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
			defer span.End()

			// Create item.
			exampleItem := fakes.BuildFakeItem()
			exampleItemInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)
			createdItem, err := testClients.main.CreateItem(ctx, exampleItemInput)
			requireNotNilAndNoProblems(t, createdItem, err)

			// fetch audit log entries
			actual, err := testClients.main.GetAuditLogForItem(ctx, createdItem.ID)
			assert.Error(t, err)
			assert.Nil(t, actual)

			// Clean up item.
			assert.NoError(t, testClients.main.ArchiveItem(ctx, createdItem.ID))
		}
	})
}
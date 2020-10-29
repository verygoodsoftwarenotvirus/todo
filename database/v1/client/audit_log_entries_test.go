package dbclient

import (
	"context"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ctx := context.Background()

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()

		c, mockDB := buildTestClient()
		mockDB.AuditLogEntryDataManager.On("GetAuditLogEntry", mock.Anything, exampleAuditLogEntry.ID).Return(exampleAuditLogEntry, nil)

		actual, err := c.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntry, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllAuditLogEntriesCount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ctx := context.Background()

		exampleCount := uint64(123)

		c, mockDB := buildTestClient()
		mockDB.AuditLogEntryDataManager.On("GetAllAuditLogEntriesCount", mock.Anything).Return(exampleCount, nil)

		actual, err := c.GetAllAuditLogEntriesCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ctx := context.Background()

		results := make(chan []models.AuditLogEntry)

		c, mockDB := buildTestClient()
		mockDB.AuditLogEntryDataManager.On("GetAllAuditLogEntries", mock.Anything, results).Return(nil)

		err := c.GetAllAuditLogEntries(ctx, results)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ctx := context.Background()

		filter := models.DefaultQueryFilter()
		exampleAuditLogEntryList := fakemodels.BuildFakeAuditLogEntryList()

		c, mockDB := buildTestClient()
		mockDB.AuditLogEntryDataManager.On("GetAuditLogEntries", mock.Anything, filter).Return(exampleAuditLogEntryList, nil)

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with nil filter", func(t *testing.T) {
		ctx := context.Background()

		filter := (*models.QueryFilter)(nil)
		exampleAuditLogEntryList := fakemodels.BuildFakeAuditLogEntryList()

		c, mockDB := buildTestClient()
		mockDB.AuditLogEntryDataManager.On("GetAuditLogEntries", mock.Anything, filter).Return(exampleAuditLogEntryList, nil)

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_CreateAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ctx := context.Background()

		exampleAuditLogEntry := fakemodels.BuildFakeAuditLogEntry()
		exampleInput := fakemodels.BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(exampleAuditLogEntry)

		c, mockDB := buildTestClient()
		mockDB.AuditLogEntryDataManager.On("CreateAuditLogEntry", mock.Anything, exampleInput).Return(nil)

		err := c.CreateAuditLogEntry(ctx, exampleInput)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

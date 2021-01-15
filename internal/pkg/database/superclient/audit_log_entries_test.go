package superclient

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAuditLogEntry := fakes.BuildFakeAuditLogEntry()

		c, _, mockDB := buildTestClient(t)
		mockDB.AuditLogEntryDataManager.On("GetAuditLogEntry", mock.Anything, exampleAuditLogEntry.ID).Return(exampleAuditLogEntry, nil)

		actual, err := c.GetAuditLogEntry(ctx, exampleAuditLogEntry.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntry, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []types.AuditLogEntry)
		exampleBucketSize := uint16(1000)

		c, _, mockDB := buildTestClient(t)
		mockDB.AuditLogEntryDataManager.On("GetAllAuditLogEntries", mock.Anything, results, exampleBucketSize).Return(nil)

		err := c.GetAllAuditLogEntries(ctx, results, exampleBucketSize)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAuditLogEntries(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()

		c, _, mockDB := buildTestClient(t)
		mockDB.AuditLogEntryDataManager.On("GetAuditLogEntries", mock.Anything, filter).Return(exampleAuditLogEntryList, nil)

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList()

		c, _, mockDB := buildTestClient(t)
		mockDB.AuditLogEntryDataManager.On("GetAuditLogEntries", mock.Anything, filter).Return(exampleAuditLogEntryList, nil)

		actual, err := c.GetAuditLogEntries(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

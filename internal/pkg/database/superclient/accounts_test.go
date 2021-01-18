package superclient

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetAccount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountDataManager.On("GetAccount", mock.Anything, exampleAccount.ID, exampleAccount.BelongsToUser).Return(exampleAccount, nil)

		actual, err := c.GetAccount(ctx, exampleAccount.ID, exampleAccount.BelongsToUser)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccount, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllAccountsCount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountDataManager.On("GetAllAccountsCount", mock.Anything).Return(exampleCount, nil)

		actual, err := c.GetAllAccountsCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllAccounts(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		results := make(chan []*types.Account)
		exampleBatchSize := uint16(1000)

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountDataManager.On("GetAllAccounts", mock.Anything, results, exampleBatchSize).Return(nil)

		err := c.GetAllAccounts(ctx, results, exampleBatchSize)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAccounts(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := types.DefaultQueryFilter()
		exampleAccountList := fakes.BuildFakeAccountList()

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountDataManager.On("GetAccounts", mock.Anything, exampleUser.ID, filter).Return(exampleAccountList, nil)

		actual, err := c.GetAccounts(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)
		exampleAccountList := fakes.BuildFakeAccountList()

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountDataManager.On("GetAccounts", mock.Anything, exampleUser.ID, filter).Return(exampleAccountList, nil)

		actual, err := c.GetAccounts(ctx, exampleUser.ID, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccountList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_CreateAccount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeAccountCreationInputFromAccount(exampleAccount)

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountDataManager.On("CreateAccount", mock.Anything, exampleInput).Return(exampleAccount, nil)

		actual, err := c.CreateAccount(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleAccount, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_UpdateAccount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		var expected error

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, _, mockDB := buildTestClient(t)

		mockDB.AccountDataManager.On("UpdateAccount", mock.Anything, exampleAccount).Return(expected)

		err := c.UpdateAccount(ctx, exampleAccount)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_ArchiveAccount(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		var expected error

		exampleUser := fakes.BuildFakeUser()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAccount.BelongsToUser = exampleUser.ID

		c, _, mockDB := buildTestClient(t)
		mockDB.AccountDataManager.On("ArchiveAccount", mock.Anything, exampleAccount.ID, exampleAccount.BelongsToUser).Return(expected)

		err := c.ArchiveAccount(ctx, exampleAccount.ID, exampleAccount.BelongsToUser)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

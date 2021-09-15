package postgres

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQuerier_UpdateUserReputation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := &types.UserReputationUpdateInput{
			TargetUserID:  exampleUser.ID,
			NewReputation: "new",
			Reason:        "because",
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnResult(newArbitraryDatabaseResult(exampleUser.ID))

		assert.NoError(t, c.UpdateUserReputation(ctx, exampleUser.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, db)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := &types.UserReputationUpdateInput{
			TargetUserID:  exampleUser.ID,
			NewReputation: "new",
			Reason:        "because",
		}

		ctx := context.Background()
		c, db := buildTestClient(t)

		db.ExpectExec(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		assert.Error(t, c.UpdateUserReputation(ctx, exampleUser.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, db)
	})
}

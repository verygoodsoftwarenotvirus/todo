package integration

import (
	"context"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	fake "github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/trace"
)

func checkItemEquality(t *testing.T, expected, actual *models.Item) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected Name for ID %d to be %v, but it was %v ", expected.ID, expected.Name, actual.Name)
	assert.Equal(t, expected.Details, actual.Details, "expected Details for ID %d to be %v, but it was %v ", expected.ID, expected.Details, actual.Details)
	assert.NotZero(t, actual.CreatedOn)
}

func buildDummyItem(t *testing.T) *models.Item {
	t.Helper()

	ctx := context.Background()
	x := &models.ItemCreationInput{
		Name:    fake.Word(),
		Details: fake.Word(),
	}
	y, err := todoClient.CreateItem(ctx, x)
	require.NoError(t, err)

	return y
}

func TestItems(test *testing.T) {
	test.Parallel()

	test.Run("Creating", func(T *testing.T) {
		T.Run("should be createable", func(t *testing.T) {
			tctx := context.Background()
			ctx, span := trace.StartSpan(tctx, t.Name())
			defer span.End()

			// Create item
			expected := &models.Item{
				Name:    fake.Word(),
				Details: fake.Word(),
			}
			premade, err := todoClient.CreateItem(ctx, &models.ItemCreationInput{
				Name:    expected.Name,
				Details: expected.Details,
			})
			checkValueAndError(t, premade, err)

			// Assert item equality
			checkItemEquality(t, expected, premade)

			// Clean up
			err = todoClient.ArchiveItem(ctx, premade.ID)
			assert.NoError(t, err)

			actual, err := todoClient.GetItem(ctx, premade.ID)
			checkValueAndError(t, actual, err)
			checkItemEquality(t, expected, actual)
			assert.NotZero(t, actual.ArchivedOn)
		})
	})

	test.Run("Listing", func(T *testing.T) {
		T.Run("should be able to be read in a list", func(t *testing.T) {
			tctx := context.Background()
			ctx, span := trace.StartSpan(tctx, t.Name())
			defer span.End()

			// Create items
			var expected []*models.Item
			for i := 0; i < 5; i++ {
				expected = append(expected, buildDummyItem(t))
			}

			// Assert item list equality
			actual, err := todoClient.GetItems(ctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(
				t,
				len(expected) <= len(actual.Items),
				"expected %d to be <= %d",
				len(expected),
				len(actual.Items),
			)

			// Clean up
			for _, x := range actual.Items {
				err = todoClient.ArchiveItem(ctx, x.ID)
				assert.NoError(t, err)
			}
		})
	})

	test.Run("Reading", func(T *testing.T) {
		T.Run("it should return an error when trying to read something that doesn't exist", func(t *testing.T) {
			tctx := context.Background()
			ctx, span := trace.StartSpan(tctx, t.Name())
			defer span.End()

			// Fetch item
			_, err := todoClient.GetItem(ctx, nonexistentID)
			assert.Error(t, err)
		})

		T.Run("it should be readable", func(t *testing.T) {
			tctx := context.Background()
			ctx, span := trace.StartSpan(tctx, t.Name())
			defer span.End()

			// Create item
			expected := &models.Item{
				Name:    fake.Word(),
				Details: fake.Word(),
			}
			premade, err := todoClient.CreateItem(ctx, &models.ItemCreationInput{
				Name:    expected.Name,
				Details: expected.Details,
			})
			checkValueAndError(t, premade, err)

			// Fetch item
			actual, err := todoClient.GetItem(ctx, premade.ID)
			checkValueAndError(t, actual, err)

			// Assert item equality
			checkItemEquality(t, expected, actual)

			// Clean up
			err = todoClient.ArchiveItem(ctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Updating", func(T *testing.T) {
		T.Run("it should return an error when trying to update something that doesn't exist", func(t *testing.T) {
			tctx := context.Background()
			ctx, span := trace.StartSpan(tctx, t.Name())
			defer span.End()

			err := todoClient.UpdateItem(ctx, &models.Item{ID: nonexistentID})
			assert.Error(t, err)
		})

		T.Run("it should be updatable", func(t *testing.T) {
			tctx := context.Background()
			ctx, span := trace.StartSpan(tctx, t.Name())
			defer span.End()

			// Create item
			expected := &models.Item{
				Name:    fake.Word(),
				Details: fake.Word(),
			}
			premade, err := todoClient.CreateItem(tctx, &models.ItemCreationInput{
				Name:    fake.Word(),
				Details: fake.Word(),
			})
			checkValueAndError(t, premade, err)

			// Change item
			premade.Update(expected.ToInput())
			err = todoClient.UpdateItem(ctx, premade)
			assert.NoError(t, err)

			// Fetch item
			actual, err := todoClient.GetItem(ctx, premade.ID)
			checkValueAndError(t, actual, err)

			// Assert item equality
			checkItemEquality(t, expected, actual)
			assert.NotNil(t, actual.UpdatedOn)

			// Clean up
			err = todoClient.ArchiveItem(ctx, actual.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Deleting", func(T *testing.T) {
		T.Run("should be able to be deleted", func(t *testing.T) {
			tctx := context.Background()
			ctx, span := trace.StartSpan(tctx, t.Name())
			defer span.End()

			// Create item
			expected := &models.Item{
				Name:    fake.Word(),
				Details: fake.Word(),
			}
			premade, err := todoClient.CreateItem(ctx, &models.ItemCreationInput{
				Name:    expected.Name,
				Details: expected.Details,
			})
			checkValueAndError(t, premade, err)

			// Clean up
			err = todoClient.ArchiveItem(ctx, premade.ID)
			assert.NoError(t, err)
		})
	})
}

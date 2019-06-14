package dbclient

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_ExportData(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := &models.User{ID: 123}
		expected := &models.DataExport{}

		c, mockDB := buildTestClient()
		mockDB.
			On("ExportData", mock.Anything, exampleUser).
			Return(expected, nil)

		actual, err := c.ExportData(ctx, exampleUser)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

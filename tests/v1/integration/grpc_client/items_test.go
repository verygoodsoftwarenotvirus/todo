package grpcclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/verygoodsoftwarenotvirus/todo/proto/v1"
)

func TestCreateItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()

		req := &todoproto.CreateItemRequest{
			Name:    "testing",
			Details: "roger dodger",
		}

		item, err := todoClient.CreateItem(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, item)
	})
}

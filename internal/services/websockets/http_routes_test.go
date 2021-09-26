package websockets

import (
	"testing"
)

func TestWebsocketsService_SubscribeHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		buildTestHelper(t)
	})
}

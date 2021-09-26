package requests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_BuildSubscribeToNotificationsURL(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/websockets/data_changes"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		expected := fmt.Sprintf("ws://%s%s", exampleDomain, expectedPathFormat)

		actual := helper.builder.BuildSubscribeToDataChangesWebsocketURL(helper.ctx)

		assert.Equal(t, expected, actual)
	})
}

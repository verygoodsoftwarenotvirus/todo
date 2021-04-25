package requests

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_BuildUserReputationUpdateInputRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/_admin_/users/status"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		exampleInput := fakes.BuildFakeUserReputationUpdateInputFromUser(h.exampleUser)
		spec := newRequestSpec(false, http.MethodPost, "", expectedPathFormat)

		actual, err := h.builder.BuildUserReputationUpdateInputRequest(h.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildUserReputationUpdateInputRequest(h.ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildUserReputationUpdateInputRequest(h.ctx, &types.UserReputationUpdateInput{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

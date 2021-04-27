package types

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIClientCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &APIClientCreationInput{
			UserLoginInput: UserLoginInput{
				Username:  t.Name(),
				Password:  t.Name(),
				TOTPToken: "123456",
			},
			Name: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx, 1, 1))
	})

	T.Run("with invalid UserLoginInput", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		x := &APIClientCreationInput{
			UserLoginInput: UserLoginInput{
				Username:  "",
				Password:  "",
				TOTPToken: "",
			},
			Name: t.Name(),
		}

		assert.Error(t, x.ValidateWithContext(ctx, 1, 1))
	})
}

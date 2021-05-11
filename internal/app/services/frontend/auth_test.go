package frontend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/assert"
)

func buildFormFromLoginRequest(input *types.UserLoginInput) url.Values {
	form := url.Values{}

	form.Set(usernameFormKey, input.Username)
	form.Set(passwordFormKey, input.Password)
	form.Set(totpTokenFormKey, input.TOTPToken)

	return form
}

func TestParseFormEncodedLoginRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		expected := &types.UserLoginInput{
			Username:  "username",
			Password:  "password",
			TOTPToken: "123456",
		}

		form := buildFormFromLoginRequest(expected)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))

		s := &Service{
			tracer: tracing.NewTracer("testing"),
			logger: zerolog.NewLogger(),
		}

		expectedRedirectTo := ""
		actual, actualRedirectTo := s.parseFormEncodedLoginRequest(ctx, req)

		assert.Equal(t, expected, actual)
		assert.Equal(t, expectedRedirectTo, actualRedirectTo)
	})
}

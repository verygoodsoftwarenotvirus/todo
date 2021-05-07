package frontend

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

		expected := &types.UserLoginInput{
			Username:  "username",
			Password:  "password",
			TOTPToken: "123456",
		}

		form := buildFormFromLoginRequest(expected)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))

		expectedRedirectTo := ""
		actual, actualRedirectTo := parseFormEncodedLoginRequest(req)

		assert.Equal(t, expected, actual)
		assert.Equal(t, expectedRedirectTo, actualRedirectTo)
	})

	T.Run("returns nil for invalid request", func(t *testing.T) {
		t.Parallel()

		badReader := &testutil.MockReadCloser{}
		badReader.On("Read", mock.IsType([]byte{})).Return(0, errors.New("blah"))

		actual, actualRedirectTo := parseFormEncodedLoginRequest(&http.Request{Body: badReader})

		assert.Nil(t, actual)
		assert.Empty(t, actualRedirectTo)
	})
}

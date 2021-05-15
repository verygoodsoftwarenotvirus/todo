package httpclient

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_newCookieRoundTripper(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, _ := buildSimpleTestClient(t)
		assert.NotNil(t, newCookieRoundTripper(c, &http.Cookie{}))
	})
}

type mockRoundTripper struct {
	mock.Mock
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	returnValues := m.Called(req)

	return returnValues.Get(0).(*http.Response), returnValues.Error(1)
}

func Test_cookieRoundtripper_RoundTrip(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		spec := newRequestSpec(true, http.MethodPost, "", "")
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)
		exampleCookie := &http.Cookie{Name: "testcookie", Value: t.Name()}
		rt := newCookieRoundTripper(c, exampleCookie)
		exampleResponse := &http.Response{
			Header: map[string][]string{
				"Set-Cookie": {exampleCookie.String()},
			},
			StatusCode: http.StatusTeapot,
		}

		mrt := &mockRoundTripper{}
		mrt.On("RoundTrip", mock.IsType(&http.Request{})).Return(exampleResponse, nil)
		rt.base = mrt

		req := httptest.NewRequest(http.MethodPost, c.URL().String(), nil)

		res, err := rt.RoundTrip(req)
		assert.NoError(t, err)
		assert.NotNil(t, res)

		assert.Equal(t, exampleResponse, res)

		mock.AssertExpectationsForObjects(t, mrt)
	})

	T.Run("with error executing RoundTrip", func(t *testing.T) {
		t.Parallel()

		spec := newRequestSpec(true, http.MethodPost, "", "")
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)
		exampleCookie := &http.Cookie{Name: "testcookie", Value: t.Name()}
		rt := newCookieRoundTripper(c, exampleCookie)

		mrt := &mockRoundTripper{}
		mrt.On("RoundTrip", mock.IsType(&http.Request{})).Return((*http.Response)(nil), errors.New("blah"))
		rt.base = mrt

		req := httptest.NewRequest(http.MethodPost, c.URL().String(), nil)

		res, err := rt.RoundTrip(req)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

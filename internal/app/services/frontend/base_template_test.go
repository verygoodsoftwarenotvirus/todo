package frontend

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
)

func TestService_homepage(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return fakes.BuildFakeSessionContextData(), nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		s.homepage(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		s.homepage(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func Test_wrapTemplateInContentDefinition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleInput := "<div>hi</div>"

		expected := `{{ define "content" }}
	<div>hi</div>
{{ end }}
`
		actual := wrapTemplateInContentDefinition(exampleInput)

		assert.Equal(t, expected, actual)
	})
}

package frontend

import (
	"context"
	"errors"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var exampleInvalidForm io.Reader = strings.NewReader("a=|%%%=%%%%%%")

func Test_buildRedirectURL(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		buildRedirectURL("/from", "/to")
	})
}

func Test_pluckRedirectURL(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/", nil)

		expected := ""
		actual := pluckRedirectURL(req)

		assert.Equal(t, expected, actual)
	})
}

func Test_htmxRedirectTo(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		res := httptest.NewRecorder()

		htmxRedirectTo(res, "/example")
	})
}

func Test_parseListOfTemplates(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleTemplateA := `<div> hi </div>`
		exampleTemplateB := `<div> bye </div>`

		actual := parseListOfTemplates(nil, exampleTemplateA, exampleTemplateB)
		assert.NotNil(t, actual)
	})
}

func TestService_renderStringToResponse(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		thing := t.Name()
		res := httptest.NewRecorder()
		s := buildTestService(t)

		s.renderStringToResponse(thing, res)
	})
}

func TestService_renderBytesToResponse(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		thing := []byte(t.Name())
		res := httptest.NewRecorder()
		s := buildTestService(t)

		s.renderBytesToResponse(thing, res)
	})

	T.Run("with error writing response", func(t *testing.T) {
		t.Parallel()

		thing := []byte(t.Name())
		res := &testutils.MockHTTPResponseWriter{}
		res.On("Write", mock.Anything).Return(-1, errors.New("blah"))

		s := buildTestService(t)

		s.renderBytesToResponse(thing, res)
	})
}

func Test_mergeFuncMaps(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		inputA := map[string]interface{}{"things": func() {}}
		inputB := map[string]interface{}{"stuff": func() {}}

		expected := template.FuncMap{
			"things": func() {},
			"stuff":  func() {},
		}

		actual := mergeFuncMaps(inputA, inputB)

		for key := range expected {
			assert.Contains(t, actual, key)
		}
	})
}

func TestService_extractFormFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		expected := url.Values{
			"things": []string{"stuff"},
		}

		exampleReq := httptest.NewRequest(http.MethodPost, "/things", strings.NewReader(expected.Encode()))

		actual, err := s.extractFormFromRequest(ctx, exampleReq)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with nil request body", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleBody := &testutils.MockReadCloser{}
		exampleBody.On("Read", mock.Anything).Return(0, errors.New("blah"))
		exampleReq := &http.Request{
			Body: exampleBody,
		}

		actual, err := s.extractFormFromRequest(ctx, exampleReq)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid body", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleReq := httptest.NewRequest(http.MethodPost, "/things", exampleInvalidForm)

		actual, err := s.extractFormFromRequest(ctx, exampleReq)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestService_renderTemplateToResponse(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleTemplate := `<div> hi </div>`
		tmpl := s.parseTemplate(ctx, "", exampleTemplate, nil)

		res := httptest.NewRecorder()

		s.renderTemplateToResponse(ctx, tmpl, nil, res)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleTemplate := `<div> {{ .Something }} </div>`
		tmpl := s.parseTemplate(ctx, "", exampleTemplate, nil)

		res := httptest.NewRecorder()

		s.renderTemplateToResponse(ctx, tmpl, struct{ Thing string }{}, res)
	})
}

func TestService_renderTemplateIntoBaseTemplate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		assert.NotNil(t, s.renderTemplateIntoBaseTemplate("<div>hi</div>", nil))
	})
}

func TestService_parseTemplate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		exampleTemplate := `<div> hi </div>`

		actual := s.parseTemplate(ctx, "", exampleTemplate, nil)
		assert.NotNil(t, actual)
	})
}

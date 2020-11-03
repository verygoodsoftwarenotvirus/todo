package encoding

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type example struct {
	Name string `json:"name" xml:"name"`
}

func TestServerEncoderDecoder_EncodeResponse(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		expectation := "name"
		ex := &example{Name: expectation}
		ed := ProvideResponseEncoder(noop.NewLogger())

		res := httptest.NewRecorder()

		ed.EncodeResponse(res, ex)
		assert.Equal(t, res.Body.String(), fmt.Sprintf("{%q:%q}\n", "name", ex.Name))
	})

	T.Run("as XML", func(t *testing.T) {
		t.Parallel()
		expectation := "name"
		ex := &example{Name: expectation}
		ed := ProvideResponseEncoder(noop.NewLogger())

		res := httptest.NewRecorder()
		res.Header().Set(ContentTypeHeader, "application/xml")

		ed.EncodeResponse(res, ex)
		assert.Equal(t, fmt.Sprintf("<example><name>%s</name></example>", expectation), res.Body.String())
	})
}

func TestServerEncoderDecoder_EncodeError(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		exampleMessage := "something went awry"
		exampleCode := http.StatusBadRequest

		ed := ProvideResponseEncoder(noop.NewLogger())

		res := httptest.NewRecorder()

		ed.EncodeErrorResponse(res, exampleMessage, exampleCode)
		assert.Equal(t, res.Body.String(), fmt.Sprintf("{\"message\":%q,\"code\":%d}\n", exampleMessage, exampleCode))
		assert.Equal(t, exampleCode, res.Code, "expected status code to match")
	})

	T.Run("as XML", func(t *testing.T) {
		t.Parallel()
		exampleMessage := "something went awry"
		exampleCode := http.StatusBadRequest

		ed := ProvideResponseEncoder(noop.NewLogger())

		res := httptest.NewRecorder()
		res.Header().Set(ContentTypeHeader, "application/xml")

		ed.EncodeErrorResponse(res, exampleMessage, exampleCode)
		assert.Equal(t, fmt.Sprintf("<ErrorResponse><Message>%s</Message><Code>%d</Code></ErrorResponse>", exampleMessage, exampleCode), res.Body.String())
		assert.Equal(t, exampleCode, res.Code, "expected status code to match")
	})
}

func TestServerEncoderDecoder_DecodeRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		expectation := "name"
		e := &example{Name: expectation}
		ed := ProvideResponseEncoder(noop.NewLogger())

		bs, err := json.Marshal(e)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(bs))
		require.NoError(t, err)

		var x example
		assert.NoError(t, ed.DecodeRequest(req, &x))
		assert.Equal(t, x.Name, e.Name)
	})

	T.Run("as XML", func(t *testing.T) {
		t.Parallel()
		expectation := "name"
		e := &example{Name: expectation}
		ed := ProvideResponseEncoder(noop.NewLogger())

		bs, err := xml.Marshal(e)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(bs))
		require.NoError(t, err)
		req.Header.Set(ContentTypeHeader, XMLContentType)

		var x example
		assert.NoError(t, ed.DecodeRequest(req, &x))
		assert.Equal(t, x.Name, e.Name)
	})
}

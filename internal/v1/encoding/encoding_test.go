package encoding

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type example struct {
	Name string `json:"name" xml:"name"`
}

func TestServerEncoderDecoder_EncodeResponse(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectation := "name"
		ex := &example{Name: expectation}
		ed := ProvideResponseEncoder()

		res := httptest.NewRecorder()
		err := ed.EncodeResponse(res, ex)

		assert.NoError(t, err)
		assert.Equal(t, res.Body.String(), fmt.Sprintf("{%q:%q}\n", "name", ex.Name))
	})

	T.Run("as XML", func(t *testing.T) {
		expectation := "name"
		ex := &example{Name: expectation}
		ed := ProvideResponseEncoder()

		res := httptest.NewRecorder()
		res.Header().Set(ContentTypeHeader, "application/xml")

		err := ed.EncodeResponse(res, ex)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("<example><name>%s</name></example>", expectation), res.Body.String())
	})
}

func TestServerEncoderDecoder_DecodeRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectation := "name"
		e := &example{Name: expectation}
		ed := ProvideResponseEncoder()

		bs, err := json.Marshal(e)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "http://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(bs))
		require.NoError(t, err)

		var x example
		assert.NoError(t, ed.DecodeRequest(req, &x))
		assert.Equal(t, x.Name, e.Name)
	})

	T.Run("as XML", func(t *testing.T) {
		expectation := "name"
		e := &example{Name: expectation}
		ed := ProvideResponseEncoder()

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

package encoding

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
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
		ed := ProvideHTTPResponseEncoder(logging.NewNonOperationalLogger())

		ctx := context.Background()
		res := httptest.NewRecorder()

		ed.EncodeResponse(ctx, res, ex)
		assert.Equal(t, res.Body.String(), fmt.Sprintf("{%q:%q}\n", "name", ex.Name))
	})

	T.Run("as XML", func(t *testing.T) {
		t.Parallel()
		expectation := "name"
		ex := &example{Name: expectation}
		ed := ProvideHTTPResponseEncoder(logging.NewNonOperationalLogger())

		ctx := context.Background()
		res := httptest.NewRecorder()
		res.Header().Set(ContentTypeHeaderKey, "application/xml")

		ed.EncodeResponse(ctx, res, ex)
		assert.Equal(t, fmt.Sprintf("<example><name>%s</name></example>", expectation), res.Body.String())
	})
}

func TestServerEncoderDecoder_EncodeResponseWithStatus(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		expectation := "name"
		ex := &example{Name: expectation}
		ed := ProvideHTTPResponseEncoder(logging.NewNonOperationalLogger())

		ctx := context.Background()
		res := httptest.NewRecorder()

		expected := 666
		ed.EncodeResponseWithStatus(ctx, res, ex, expected)

		assert.Equal(t, expected, res.Code, "expected code to be %d, but got %d", expected, res.Code)
		assert.Equal(t, res.Body.String(), fmt.Sprintf("{%q:%q}\n", "name", ex.Name))
	})
}

func TestServerEncoderDecoder_EncodeErrorResponse(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		exampleMessage := "something went awry"
		exampleCode := http.StatusBadRequest

		ed := ProvideHTTPResponseEncoder(logging.NewNonOperationalLogger())

		ctx := context.Background()
		res := httptest.NewRecorder()

		ed.EncodeErrorResponse(ctx, res, exampleMessage, exampleCode)
		assert.Equal(t, res.Body.String(), fmt.Sprintf("{\"message\":%q,\"code\":%d}\n", exampleMessage, exampleCode))
		assert.Equal(t, exampleCode, res.Code, "expected status code to match")
	})

	T.Run("as XML", func(t *testing.T) {
		t.Parallel()
		exampleMessage := "something went awry"
		exampleCode := http.StatusBadRequest

		ed := ProvideHTTPResponseEncoder(logging.NewNonOperationalLogger())

		ctx := context.Background()
		res := httptest.NewRecorder()
		res.Header().Set(ContentTypeHeaderKey, "application/xml")

		ed.EncodeErrorResponse(ctx, res, exampleMessage, exampleCode)
		assert.Equal(t, fmt.Sprintf("<ErrorResponse><Message>%s</Message><Code>%d</Code></ErrorResponse>", exampleMessage, exampleCode), res.Body.String())
		assert.Equal(t, exampleCode, res.Code, "expected status code to match")
	})
}

func TestEncodeInvalidInputResponse(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		res := httptest.NewRecorder()

		ed := ProvideHTTPResponseEncoder(logging.NewNonOperationalLogger())
		ed.EncodeInvalidInputResponse(ctx, res)

		expectedCode := http.StatusBadRequest
		assert.EqualValues(t, expectedCode, res.Code, "expected code to be %d, got %d instead", expectedCode, res.Code)
	})
}

func TestEncodeNotFoundResponse(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		res := httptest.NewRecorder()

		ed := ProvideHTTPResponseEncoder(logging.NewNonOperationalLogger())
		ed.EncodeNotFoundResponse(ctx, res)

		expectedCode := http.StatusNotFound
		assert.EqualValues(t, expectedCode, res.Code, "expected code to be %d, got %d instead", expectedCode, res.Code)
	})
}

func TestEncodeUnspecifiedInternalServerErrorResponse(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		res := httptest.NewRecorder()

		ed := ProvideHTTPResponseEncoder(logging.NewNonOperationalLogger())
		ed.EncodeUnspecifiedInternalServerErrorResponse(ctx, res)

		expectedCode := http.StatusInternalServerError
		assert.EqualValues(t, expectedCode, res.Code, "expected code to be %d, got %d instead", expectedCode, res.Code)
	})
}

func TestEncodeUnauthorizedResponse(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		res := httptest.NewRecorder()

		ed := ProvideHTTPResponseEncoder(logging.NewNonOperationalLogger())
		ed.EncodeUnauthorizedResponse(ctx, res)

		expectedCode := http.StatusUnauthorized
		assert.EqualValues(t, expectedCode, res.Code, "expected code to be %d, got %d instead", expectedCode, res.Code)
	})
}

func TestServerEncoderDecoder_DecodeRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectation := "name"
		e := &example{Name: expectation}
		ed := ProvideHTTPResponseEncoder(logging.NewNonOperationalLogger())

		bs, err := json.Marshal(e)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			bytes.NewReader(bs),
		)
		require.NoError(t, err)

		var x example
		assert.NoError(t, ed.DecodeRequest(ctx, req, &x))
		assert.Equal(t, x.Name, e.Name)
	})

	T.Run("as XML", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectation := "name"
		e := &example{Name: expectation}
		ed := ProvideHTTPResponseEncoder(logging.NewNonOperationalLogger())

		bs, err := xml.Marshal(e)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			bytes.NewReader(bs),
		)
		require.NoError(t, err)
		req.Header.Set(ContentTypeHeaderKey, ContentTypeXML)

		var x example
		assert.NoError(t, ed.DecodeRequest(ctx, req, &x))
		assert.Equal(t, x.Name, e.Name)
	})
}

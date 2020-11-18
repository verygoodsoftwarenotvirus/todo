package mock

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"

	"github.com/stretchr/testify/mock"
)

var _ encoding.EncoderDecoder = (*EncoderDecoder)(nil)

// EncoderDecoder is a mock EncoderDecoder.
type EncoderDecoder struct {
	mock.Mock
}

// EncodeResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeResponse(res http.ResponseWriter, val interface{}) {
	m.Called(res, val)
}

// EncodeResponseWithStatus satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeResponseWithStatus(res http.ResponseWriter, val interface{}, statusCode int) {
	m.Called(res, val, statusCode)
	res.WriteHeader(statusCode)
}

// EncodeErrorResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeErrorResponse(res http.ResponseWriter, msg string, statusCode int) {
	m.Called(res, msg, statusCode)
	res.WriteHeader(statusCode)
}

// EncodeNoInputResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeNoInputResponse(res http.ResponseWriter) {
	m.Called(res)
	res.WriteHeader(http.StatusBadRequest)
}

// EncodeNotFoundResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeNotFoundResponse(res http.ResponseWriter) {
	m.Called(res)
	res.WriteHeader(http.StatusNotFound)
}

// EncodeUnspecifiedInternalServerErrorResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeUnspecifiedInternalServerErrorResponse(res http.ResponseWriter) {
	m.Called(res)
	res.WriteHeader(http.StatusInternalServerError)
}

// EncodeUnauthorizedResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeUnauthorizedResponse(res http.ResponseWriter) {
	m.Called(res)
	res.WriteHeader(http.StatusUnauthorized)
}

// EncodeInvalidPermissionsResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeInvalidPermissionsResponse(res http.ResponseWriter) {
	m.Called(res)
	res.WriteHeader(http.StatusForbidden)
}

// DecodeRequest satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) DecodeRequest(req *http.Request, v interface{}) error {
	return m.Called(req, v).Error(0)
}

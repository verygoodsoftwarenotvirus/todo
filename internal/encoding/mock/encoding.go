package mock

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
)

var _ encoding.ServerEncoderDecoder = (*EncoderDecoder)(nil)

// NewMockEncoderDecoder produces a mock EncoderDecoder.
func NewMockEncoderDecoder() *EncoderDecoder {
	return &EncoderDecoder{}
}

// EncoderDecoder is a mock EncoderDecoder.
type EncoderDecoder struct {
	mock.Mock
}

// MustEncode satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) MustEncode(ctx context.Context, v interface{}) []byte {
	return m.Called(ctx, v).Get(0).([]byte)
}

// MustEncodeJSON satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) MustEncodeJSON(ctx context.Context, v interface{}) []byte {
	return m.Called(ctx, v).Get(0).([]byte)
}

// RespondWithData satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) RespondWithData(ctx context.Context, res http.ResponseWriter, val interface{}) {
	m.Called(ctx, res, val)
}

// EncodeResponseWithStatus satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeResponseWithStatus(ctx context.Context, res http.ResponseWriter, val interface{}, statusCode int) {
	m.Called(ctx, res, val, statusCode)
	res.WriteHeader(statusCode)
}

// EncodeErrorResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeErrorResponse(ctx context.Context, res http.ResponseWriter, msg string, statusCode int) {
	m.Called(ctx, res, msg, statusCode)
	res.WriteHeader(statusCode)
}

// EncodeInvalidInputResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeInvalidInputResponse(ctx context.Context, res http.ResponseWriter) {
	m.Called(ctx, res)
	res.WriteHeader(http.StatusBadRequest)
}

// EncodeNotFoundResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeNotFoundResponse(ctx context.Context, res http.ResponseWriter) {
	m.Called(ctx, res)
	res.WriteHeader(http.StatusNotFound)
}

// EncodeUnspecifiedInternalServerErrorResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeUnspecifiedInternalServerErrorResponse(ctx context.Context, res http.ResponseWriter) {
	m.Called(ctx, res)
	res.WriteHeader(http.StatusInternalServerError)
}

// EncodeUnauthorizedResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeUnauthorizedResponse(ctx context.Context, res http.ResponseWriter) {
	m.Called(ctx, res)
	res.WriteHeader(http.StatusUnauthorized)
}

// EncodeInvalidPermissionsResponse satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) EncodeInvalidPermissionsResponse(ctx context.Context, res http.ResponseWriter) {
	m.Called(ctx, res)
	res.WriteHeader(http.StatusForbidden)
}

// DecodeRequest satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) DecodeRequest(ctx context.Context, req *http.Request, v interface{}) error {
	return m.Called(ctx, req, v).Error(0)
}

// DecodeBytes satisfies our EncoderDecoder interface.
func (m *EncoderDecoder) DecodeBytes(ctx context.Context, data []byte, v interface{}) error {
	return m.Called(ctx, data, v).Error(0)
}

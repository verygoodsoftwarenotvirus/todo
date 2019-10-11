package mock

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"

	"github.com/stretchr/testify/mock"
)

var _ encoding.EncoderDecoder = (*EncoderDecoder)(nil)

// EncoderDecoder is a mock EncoderDecoder
type EncoderDecoder struct {
	mock.Mock
}

// EncodeResponse satisfies our EncoderDecoder interface
func (m *EncoderDecoder) EncodeResponse(res http.ResponseWriter, v interface{}) error {
	return m.Called(res, v).Error(0)
}

// DecodeRequest satisfies our EncoderDecoder interface
func (m *EncoderDecoder) DecodeRequest(req *http.Request, v interface{}) error {
	return m.Called(req, v).Error(0)
}

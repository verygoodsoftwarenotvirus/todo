package mock

import (
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"net/http"
)

var _ encoding.EncoderDecoder = (*EncoderDecoder)(nil)

type EncoderDecoder struct {
	mock.Mock
}

func (m *EncoderDecoder) EncodeResponse(res http.ResponseWriter, v interface{}) error {
	return m.Called(res, v).Error(0)
}

func (m *EncoderDecoder) DecodeRequest(req *http.Request, v interface{}) error {
	return m.Called(req, v).Error(0)
}

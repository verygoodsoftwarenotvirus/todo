package images

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
)

var _ ImageUploadProcessor = (*MockImageUploadProcessor)(nil)

// MockImageUploadProcessor is a mock ImageUploadProcessor.
type MockImageUploadProcessor struct {
	mock.Mock
}

// Process satisfies the ImageUploadProcessor interface.
func (m *MockImageUploadProcessor) Process(ctx context.Context, req *http.Request, filename string) (*Image, error) {
	args := m.Called(ctx, req, filename)

	return args.Get(0).(*Image), args.Error(1)
}

// BuildAvatarUploadMiddleware satisfies the ImageUploadProcessor interface.
func (m *MockImageUploadProcessor) BuildAvatarUploadMiddleware(next http.Handler, encoderDecoder encoding.ServerEncoderDecoder, filename string) http.Handler {
	args := m.Called(next, encoderDecoder, filename)

	return args.Get(0).(http.Handler)
}

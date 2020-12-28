package images

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

var _ ImageUploadProcessor = (*MockImageUploadProcessor)(nil)

// MockImageUploadProcessor is a mock ImageUploadProcessor.
type MockImageUploadProcessor struct {
	mock.Mock
}

// Process satisfies the ImageUploadProcessor interface.
func (m *MockImageUploadProcessor) Process(r *http.Request, filename string) (*Image, error) {
	args := m.Called(r, filename)

	return args.Get(0).(*Image), args.Error(1)
}

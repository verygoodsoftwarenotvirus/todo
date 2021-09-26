package mock

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads"
)

var _ uploads.UploadManager = (*UploadManager)(nil)

// UploadManager is a mock UploadManager.
type UploadManager struct {
	mock.Mock
}

// SaveFile satisfies the UploadManager interface.
func (m *UploadManager) SaveFile(ctx context.Context, path string, content []byte) error {
	return m.Called(ctx, path, content).Error(0)
}

// ReadFile satisfies the UploadManager interface.
func (m *UploadManager) ReadFile(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)

	return args.Get(0).([]byte), args.Error(1)
}

// ServeFiles satisfies the UploadManager interface.
func (m *UploadManager) ServeFiles(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

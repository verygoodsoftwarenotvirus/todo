package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads"

	"github.com/stretchr/testify/mock"
)

type mockUploadManager struct {
	mock.Mock
}

// NewMockUploadManager creates a new UploadManager.
func NewMockUploadManager() uploads.UploadManager {
	return &mockUploadManager{}
}

func (m *mockUploadManager) SaveFile(ctx context.Context, path string, content []byte) error {
	return m.Called(ctx, path, content).Error(0)
}

func (m *mockUploadManager) ReadFile(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)

	return args.Get(0).([]byte), args.Error(1)
}

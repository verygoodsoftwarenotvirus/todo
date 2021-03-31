package mock

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.FrontendService = (*FrontendService)(nil)

// FrontendService is a mock types.FrontendService.
type FrontendService struct {
	mock.Mock
}

// StaticDir implements our FrontendService interface.
func (m *FrontendService) StaticDir(ctx context.Context, staticFilesDirectory string) (http.HandlerFunc, error) {
	args := m.Called(staticFilesDirectory)

	return args.Get(0).(http.HandlerFunc), args.Error(1)
}

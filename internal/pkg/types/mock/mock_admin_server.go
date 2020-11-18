package mock

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AdminServer = (*AdminServer)(nil)

// AdminServer is a mocked types.AdminServer for testing.
type AdminServer struct {
	mock.Mock
}

// BanHandler implements our interface requirements.
func (m *AdminServer) BanHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

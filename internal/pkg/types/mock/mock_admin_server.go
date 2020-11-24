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

// UserAccountStatusChangeHandler implements our interface requirements.
func (m *AdminServer) UserAccountStatusChangeHandler(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// AccountStatusUpdateInputMiddleware implements our interface requirements.
func (m *AdminServer) AccountStatusUpdateInputMiddleware(next http.Handler) http.Handler {
	return m.Called(next).Get(0).(http.Handler)
}

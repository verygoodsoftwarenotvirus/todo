package mock

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.ItemDataServer = (*ItemDataServer)(nil)

// ItemDataServer describes a structure capable of serving traffic related to users
type ItemDataServer struct {
	mock.Mock
}

// CreationInputMiddleware implements our interface requirements
func (m *ItemDataServer) CreationInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// UpdateInputMiddleware implements our interface requirements
func (m *ItemDataServer) UpdateInputMiddleware(next http.Handler) http.Handler {
	args := m.Called(next)
	return args.Get(0).(http.Handler)
}

// ListHandler implements our interface requirements
func (m *ItemDataServer) ListHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// CreateHandler implements our interface requirements
func (m *ItemDataServer) CreateHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// ReadHandler implements our interface requirements
func (m *ItemDataServer) ReadHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// UpdateHandler implements our interface requirements
func (m *ItemDataServer) UpdateHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

// ArchiveHandler implements our interface requirements
func (m *ItemDataServer) ArchiveHandler() http.HandlerFunc {
	args := m.Called()
	return args.Get(0).(http.HandlerFunc)
}

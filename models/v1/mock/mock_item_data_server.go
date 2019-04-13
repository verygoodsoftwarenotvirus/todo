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

// List implements our interface requirements
func (m *ItemDataServer) List(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// Create implements our interface requirements
func (m *ItemDataServer) Create(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// Read implements our interface requirements
func (m *ItemDataServer) Read(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// Update implements our interface requirements
func (m *ItemDataServer) Update(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

// Delete implements our interface requirements
func (m *ItemDataServer) Delete(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

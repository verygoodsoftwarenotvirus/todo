package mock

import (
	"net/http"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
)

// NewMockRouter provides our mock Router to tests.
func NewMockRouter() *Router {
	return &Router{}
}

// Router is a mock routing.Router.
type Router struct {
	mock.Mock
}

// LogRoutes satisfies the interface contract.
func (m *Router) LogRoutes() {
	m.Called()
}

// Handler satisfies the interface contract.
func (m *Router) Handler() http.Handler {
	return m.Called().Get(0).(http.Handler)
}

// WithMiddleware satisfies the interface contract.
func (m *Router) WithMiddleware(middleware ...routing.Middleware) routing.Router {
	return m.Called(middleware).Get(0).(routing.Router)
}

// AddRoute satisfies the interface contract.
func (m *Router) AddRoute(method, path string, handler http.HandlerFunc, middleware ...routing.Middleware) error {
	return m.Called(method, path, handler, middleware).Error(0)
}

// Handle satisfies the interface contract.
func (m *Router) Handle(pattern string, handler http.Handler) {
	m.Called(pattern, handler)
}

// HandleFunc satisfies the interface contract.
func (m *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	m.Called(pattern, handler)
}

// Route satisfies the interface contract.
func (m *Router) Route(pattern string, fn func(r Router)) routing.Router {
	return m.Called(pattern, fn).Get(0).(routing.Router)
}

// Connect satisfies the interface contract.
func (m *Router) Connect(pattern string, handler http.HandlerFunc) {
	m.Called(pattern, handler)
}

// Delete satisfies the interface contract.
func (m *Router) Delete(pattern string, handler http.HandlerFunc) {
	m.Called(pattern, handler)
}

// Get satisfies the interface contract.
func (m *Router) Get(pattern string, handler http.HandlerFunc) {
	m.Called(pattern, handler)
}

// Head satisfies the interface contract.
func (m *Router) Head(pattern string, handler http.HandlerFunc) {
	m.Called(pattern, handler)
}

// Options satisfies the interface contract.
func (m *Router) Options(pattern string, handler http.HandlerFunc) {
	m.Called(pattern, handler)
}

// Patch satisfies the interface contract.
func (m *Router) Patch(pattern string, handler http.HandlerFunc) {
	m.Called(pattern, handler)
}

// Post satisfies the interface contract.
func (m *Router) Post(pattern string, handler http.HandlerFunc) {
	m.Called(pattern, handler)
}

// Put satisfies the interface contract.
func (m *Router) Put(pattern string, handler http.HandlerFunc) {
	m.Called(pattern, handler)
}

// Trace satisfies the interface contract.
func (m *Router) Trace(pattern string, handler http.HandlerFunc) {
	m.Called(pattern, handler)
}

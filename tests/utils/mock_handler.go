package testutils

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

var _ http.Handler = (*MockHTTPHandler)(nil)

// MockHTTPHandler is a mocked http.Handler.
type MockHTTPHandler struct {
	mock.Mock
}

// ServeHTTP satisfies our interface requirements.
func (m *MockHTTPHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	m.Called(res, req)
}

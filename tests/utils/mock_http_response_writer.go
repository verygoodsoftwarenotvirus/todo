package testutils

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

var _ http.ResponseWriter = (*MockHTTPResponseWriter)(nil)

// MockHTTPResponseWriter is a mock http.ResponseWriter.
type MockHTTPResponseWriter struct {
	mock.Mock
}

// Header satisfies our interface requirements.
func (m *MockHTTPResponseWriter) Header() http.Header {
	return m.Called().Get(0).(http.Header)
}

// Write satisfies our interface requirements.
func (m *MockHTTPResponseWriter) Write(in []byte) (int, error) {
	returnValues := m.Called(in)

	return returnValues.Int(0), returnValues.Error(1)
}

// WriteHeader satisfies our interface requirements.
func (m *MockHTTPResponseWriter) WriteHeader(statusCode int) {
	m.Called(statusCode)
}

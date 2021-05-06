package testutils

import "github.com/stretchr/testify/mock"

// MockReadCloser mocks a io.ReadCloser.
type MockReadCloser struct {
	mock.Mock
}

// Read implements the io.ReadCloser interface.
func (m *MockReadCloser) Read(p []byte) (n int, err error) {
	returnValues := m.Called(p)

	return returnValues.Int(0), returnValues.Error(1)
}

// Close implements the io.ReadCloser interface.
func (m *MockReadCloser) Close() (err error) {
	returnValues := m.Called()

	return returnValues.Error(0)
}

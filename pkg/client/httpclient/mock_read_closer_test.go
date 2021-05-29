package httpclient

import (
	"io"

	"github.com/stretchr/testify/mock"
)

var _ io.ReadCloser = (*mockReadCloser)(nil)

// NOTE: there does exist a testutil.MockReadCloser, but since testutil currently imports this package, we need to duplicate the code here.

// mockReadCloser is a mock io.ReadCloser for testing purposes.
type mockReadCloser struct {
	mock.Mock
}

// newMockReadCloser returns a new mock io.ReadCloser.
func newMockReadCloser() *mockReadCloser {
	return &mockReadCloser{}
}

// ReadHandler implements the ReadHandler part of our mockReadCloser.
func (m *mockReadCloser) Read(b []byte) (int, error) {
	retVals := m.Called(b)
	return retVals.Int(0), retVals.Error(1)
}

// Close implements the Closer part of our mockReadCloser.
func (m *mockReadCloser) Close() (err error) {
	return m.Called().Error(0)
}

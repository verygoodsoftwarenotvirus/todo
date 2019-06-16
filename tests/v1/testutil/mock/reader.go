package mock

import (
	"io"

	"github.com/stretchr/testify/mock"
)

var _ io.ReadCloser = (*ReadCloser)(nil)

// ReadCloser is a mock io.ReadCloser for testing purposes
type ReadCloser struct {
	mock.Mock
}

// NewMockReadCloser returns a new mock io.ReadCloser
func NewMockReadCloser() *ReadCloser {
	return &ReadCloser{}
}

// ReadHandler implements the ReadHandler part of our ReadCloser
func (m *ReadCloser) Read(b []byte) (i int, err error) {
	retVals := m.Called(b)
	return retVals.Int(0), retVals.Error(1)
}

// Close implements the Closer part of our ReadCloser
func (m *ReadCloser) Close() (err error) {
	return m.Called().Error(1)
}

package panicking

import (
	"github.com/stretchr/testify/mock"
)

// NewMockPanicker produces a production-ready panicker that will actually panic when called.
func NewMockPanicker() *MockPanicker {
	return &MockPanicker{}
}

// MockPanicker implements Panicker for tests.
type MockPanicker struct {
	mock.Mock
}

// Panic satisfies our interface.
func (p *MockPanicker) Panic(msg interface{}) {
	p.Called(msg)
}

// Panicf satisfies our interface.
func (p *MockPanicker) Panicf(format string, args ...interface{}) {
	p.Called(format, args)
}

package mockpanicking

import (
	"github.com/stretchr/testify/mock"
)

// Panicker implements Panicker for tests.
type Panicker struct {
	mock.Mock
}

// NewMockPanicker produces a production-ready panicker that will actually panic when called.
func NewMockPanicker() *Panicker {
	return &Panicker{}
}

// Panic satisfies our interface.
func (p *Panicker) Panic(msg interface{}) {
	p.Called(msg)
}

// Panicf satisfies our interface.
func (p *Panicker) Panicf(format string, args ...interface{}) {
	p.Called(format, args)
}

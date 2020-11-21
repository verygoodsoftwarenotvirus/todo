package queriers

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// TimeTeller can tell you the time.
type TimeTeller interface {
	Now() uint64
}

var _ TimeTeller = (*StandardTimeTeller)(nil)

// StandardTimeTeller satisfies our TimeTeller interface via the standard library.
type StandardTimeTeller struct{}

// Now returns the current unix time.
func (t *StandardTimeTeller) Now() uint64 {
	return uint64(time.Now().Unix())
}

var _ TimeTeller = (*MockTimeTeller)(nil)

// MockTimeTeller is a mock TimeTeller.
type MockTimeTeller struct {
	mock.Mock
}

// Now satisfies the TimeTeller interface.
func (m *MockTimeTeller) Now() uint64 {
	return m.Called().Get(0).(uint64)
}

package querybuilding

import (
	"github.com/stretchr/testify/mock"
)

// MockExternalIDGenerator generates external IDs.
type MockExternalIDGenerator struct {
	mock.Mock
}

// NewExternalID implements our interface.
func (m *MockExternalIDGenerator) NewExternalID() string {
	return m.Called().String(0)
}

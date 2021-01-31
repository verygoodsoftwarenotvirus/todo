package querybuilding

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// ExternalIDGenerator generates external IDs.
type ExternalIDGenerator interface {
	NewExternalID() string
}

// UUIDExternalIDGenerator generates external IDs.
type UUIDExternalIDGenerator struct{}

// NewExternalID implements our interface.
func (g UUIDExternalIDGenerator) NewExternalID() string {
	return uuid.New().String()
}

// MockExternalIDGenerator generates external IDs.
type MockExternalIDGenerator struct {
	mock.Mock
}

// NewExternalID implements our interface.
func (m *MockExternalIDGenerator) NewExternalID() string {
	return m.Called().String(0)
}

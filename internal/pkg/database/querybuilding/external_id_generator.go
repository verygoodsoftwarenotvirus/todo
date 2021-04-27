package querybuilding

import (
	"github.com/google/uuid"
)

// ExternalIDGenerator generates external IDs.
type ExternalIDGenerator interface {
	NewExternalID() string
}

// UUIDExternalIDGenerator generates external IDs.
type UUIDExternalIDGenerator struct{}

// NewExternalID implements our interface.
func (g UUIDExternalIDGenerator) NewExternalID() string {
	return uuid.NewString()
}

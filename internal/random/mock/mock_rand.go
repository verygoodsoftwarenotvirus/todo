package mockrandom

import (
	"context"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/random"
)

var _ random.Generator = (*Generator)(nil)

// Generator is a mock Generator.
type Generator struct {
	mock.Mock
}

// GenerateBase32EncodedString implements our interface.
func (m *Generator) GenerateBase32EncodedString(ctx context.Context, length int) (string, error) {
	args := m.Called(ctx, length)

	return args.String(0), args.Error(1)
}

// GenerateBase64EncodedString implements our interface.
func (m *Generator) GenerateBase64EncodedString(ctx context.Context, length int) (string, error) {
	args := m.Called(ctx, length)

	return args.String(0), args.Error(1)
}

// GenerateRawBytes implements our interface.
func (m *Generator) GenerateRawBytes(ctx context.Context, length int) ([]byte, error) {
	args := m.Called(ctx, length)

	return args.Get(0).([]byte), args.Error(1)
}

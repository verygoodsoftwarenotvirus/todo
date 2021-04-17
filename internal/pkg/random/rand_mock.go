package random

import (
	"context"

	mock "github.com/stretchr/testify/mock"
)

var _ Generator = (*MockGenerator)(nil)

// MockGenerator is a mock Generator.
type MockGenerator struct {
	mock.Mock
}

// GenerateBase32EncodedString implements our interface.
func (m *MockGenerator) GenerateBase32EncodedString(ctx context.Context, length int) (string, error) {
	args := m.Called(ctx, length)

	return args.String(0), args.Error(1)
}

// GenerateBase64EncodedString implements our interface.
func (m *MockGenerator) GenerateBase64EncodedString(ctx context.Context, length int) (string, error) {
	args := m.Called(ctx, length)

	return args.String(0), args.Error(1)
}

// GenerateRawBytes implements our interface.
func (m *MockGenerator) GenerateRawBytes(ctx context.Context, length int) ([]byte, error) {
	args := m.Called(ctx, length)

	return args.Get(0).([]byte), args.Error(1)
}

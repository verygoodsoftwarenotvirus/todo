package authentication

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type mockSessionManager struct {
	mock.Mock
}

func (m *mockSessionManager) Load(ctx context.Context, token string) (context.Context, error) {
	returnArgs := m.Called(ctx, token)

	return returnArgs.Get(0).(context.Context), returnArgs.Error(1)
}

func (m *mockSessionManager) RenewToken(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockSessionManager) Get(ctx context.Context, key string) interface{} {
	return m.Called(ctx, key).Get(0)
}

func (m *mockSessionManager) Put(ctx context.Context, key string, val interface{}) {
	m.Called(ctx, key, val)
}

func (m *mockSessionManager) Commit(ctx context.Context) (string, time.Time, error) {
	returnArgs := m.Called(ctx)

	return returnArgs.String(0), returnArgs.Get(1).(time.Time), returnArgs.Error(2)
}

func (m *mockSessionManager) Destroy(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

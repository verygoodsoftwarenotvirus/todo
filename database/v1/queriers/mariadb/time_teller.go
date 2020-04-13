package mariadb

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type timeTeller interface {
	Now() uint64
}

type stdLibTimeTeller struct{}

func (t *stdLibTimeTeller) Now() uint64 {
	return uint64(time.Now().Unix())
}

type mockTimeTeller struct {
	mock.Mock
}

func (m *mockTimeTeller) Now() uint64 {
	return m.Called().Get(0).(uint64)
}

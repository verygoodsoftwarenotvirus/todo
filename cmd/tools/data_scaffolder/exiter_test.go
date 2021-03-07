package main

import (
	"github.com/stretchr/testify/mock"
)

var _ Quitter = &mockQuitter{}

type mockQuitter struct {
	mock.Mock
}

func (m *mockQuitter) Quit(code int) {
	m.Called(code)
}

func (m *mockQuitter) ComplainAndQuit(v ...interface{}) {
	m.Called(v...)
}

func (m *mockQuitter) ComplainAndQuitf(s string, args ...interface{}) {
	m.Called(s, args)
}

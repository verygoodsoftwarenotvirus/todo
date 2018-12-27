package server

import (
	"net/http"
)

type LoginMonitor interface {
	LogSuccessfulAttempt(userID string)
	LogUnsuccessfulAttempt(userID string)
	LoginAttemptsExhausted(userID string) error
	NotifyExhaustedAttempts(res http.ResponseWriter)
}

type NoopLoginMonitor struct{}

func (n *NoopLoginMonitor) LogSuccessfulAttempt(userID string)              {}
func (n *NoopLoginMonitor) LogUnsuccessfulAttempt(userID string)            {}
func (n *NoopLoginMonitor) NotifyExhaustedAttempts(res http.ResponseWriter) {}
func (n *NoopLoginMonitor) LoginAttemptsExhausted(userID string) error      { return nil }

package server

type LoginMonitor interface {
	LogSuccessfulAttempt(userID string)
	LogUnsuccessfulAttempt(userID string)
	LoginAttemptsExhausted(userID string) error
}

type NoopLoginMonitor struct{}

func (n *NoopLoginMonitor) LogSuccessfulAttempt(userID string)         {}
func (n *NoopLoginMonitor) LogUnsuccessfulAttempt(userID string)       {}
func (n *NoopLoginMonitor) LoginAttemptsExhausted(userID string) error { return nil }

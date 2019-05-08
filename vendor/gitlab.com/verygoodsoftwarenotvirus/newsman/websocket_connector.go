package newsman

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// WebsocketConnector is an abstraction around a gorilla/websocket.Conn
type WebsocketConnector interface {
	SetWriteDeadline(time.Time) error
	WriteMessage(msg int, data []byte) error
	WriteJSON(v interface{}) error
	Close() error
}

// MockWebsocketConnector is a mock WebsocketConnector
type MockWebsocketConnector struct {
	mock.Mock
}

// SetWriteDeadline satisfies our interface
func (m *MockWebsocketConnector) SetWriteDeadline(time.Time) error {
	return m.Called().Error(0)
}

// WriteMessage satisfies our interface
func (m *MockWebsocketConnector) WriteMessage(msg int, data []byte) error {
	return m.Called().Error(0)
}

// WriteJSON satisfies our interface
func (m *MockWebsocketConnector) WriteJSON(v interface{}) error {
	return m.Called().Error(0)
}

// Close satisfies our interface
func (m *MockWebsocketConnector) Close() error {
	return m.Called().Error(0)
}

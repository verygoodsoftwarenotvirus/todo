package newsman

// Listener is our interface that can be implemented via webhooks/websockets/etc.
type Listener interface {
	Listen()

	Channel() chan Event
	Config() *ListenerConfig
}

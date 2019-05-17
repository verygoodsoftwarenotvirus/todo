// +build wasm

package html

// EventTarget represents the interface described here: https://developer.mozilla.org/en-US/docs/Web/API/EventTarget
type EventTarget interface {
	AddEventListener()
	RemoveEventListener()
	DispatchEvent()
}

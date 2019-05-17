// +build wasm

package html

import (
	"syscall/js"
)

var (
	globalWindow Window
)

// Window is a stand-in for a browser's `window` object
type Window interface {
	AddEventListener(eventName string, callback js.Func)
	Location() Location
}

type window struct {
	jsWindow js.Value
}

func (w *window) AddEventListener(eventName string, callback js.Func) {
	w.jsWindow.Call("addEventListener", eventName, callback)
}

func (w *window) Location() Location {
	l := &location{
		jsLocation: w.jsWindow.Get("location").JSValue(),
	}
	return l
}

// GetWindow returns the js.Value for the `window` object in a browser
func GetWindow() Window {
	if globalWindow != nil {
		return globalWindow
	}

	jsw := js.Global().Get("window")
	globalWindow = &window{
		jsWindow: jsw,
	}

	return globalWindow
}

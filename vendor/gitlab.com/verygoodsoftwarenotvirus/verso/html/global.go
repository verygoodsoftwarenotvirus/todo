// +build wasm

package html

import (
	"syscall/js"
)

// Valuer is an interface which is satisfied by any element which can represent itself as a js.Value
type Valuer interface {
	JSValue() js.Value
}

// Alert attempts to call `alert` for your message
func Alert(message string) {
	js.Global().Call("alert", message)
}
